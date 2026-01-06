package ai

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConsoleProgressHandler(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handler := ConsoleProgressHandler(true)

	// Test different event types
	events := []ProgressEvent{
		{
			Phase:     PhasePlanning,
			EventType: ProgressStart,
			Message:   "starting planning",
		},
		{
			Phase:     PhaseGenerating,
			EventType: ProgressStart,
			Message:   "starting generation",
		},
		{
			Phase:     PhaseGenerating,
			EventType: ProgressPageStart,
			PagePath:  "content/about.md",
		},
		{
			Phase:     PhaseGenerating,
			EventType: ProgressPageDone,
			PagePath:  "content/about.md",
			Progress:  0.5,
			Current:   1,
			Total:     2,
		},
		{
			Phase:     PhaseGenerating,
			EventType: ProgressRetry,
			PagePath:  "content/test.md",
			Message:   "retry attempt 2/3",
		},
		{
			Phase:     PhaseGenerating,
			EventType: ProgressSkip,
			PagePath:  "content/existing.md",
			Message:   "file exists",
		},
		{
			Phase:     PhaseGenerating,
			EventType: ProgressError,
			PagePath:  "content/error.md",
			Message:   "generation failed",
		},
		{
			Phase:     PhasePlanning,
			EventType: ProgressComplete,
			Message:   "plan created with 5 pages",
		},
		{
			Phase:     PhaseGenerating,
			EventType: ProgressComplete,
			Message:   "generation done",
		},
		{
			Phase:     PhaseCompleted,
			EventType: ProgressComplete,
			Message:   "completed: 5/5 pages",
		},
	}

	for _, event := range events {
		handler(event)
	}

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Errorf("Failed to copy output: %v", err)
	}
	output := buf.String()

	// Verify some output was generated
	if len(output) == 0 {
		t.Error("expected output from console handler")
	}
}

func TestConsoleProgressHandler_NonVerbose(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handler := ConsoleProgressHandler(false) // Non-verbose

	// Page start should not print in non-verbose mode
	handler(ProgressEvent{
		Phase:     PhaseGenerating,
		EventType: ProgressPageStart,
		PagePath:  "content/about.md",
	})

	// Skip should not print in non-verbose mode
	handler(ProgressEvent{
		Phase:     PhaseGenerating,
		EventType: ProgressSkip,
		PagePath:  "content/existing.md",
		Message:   "file exists",
	})

	// But page done should still print
	handler(ProgressEvent{
		Phase:     PhaseGenerating,
		EventType: ProgressPageDone,
		PagePath:  "content/about.md",
		Progress:  0.5,
		Current:   1,
		Total:     2,
	})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Errorf("Failed to copy output: %v", err)
	}
	output := buf.String()

	// Should only have page done output
	if !strings.Contains(output, "about.md") {
		t.Error("expected page done output")
	}
}

func TestJSONProgressHandler(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handler := JSONProgressHandler()

	event := ProgressEvent{
		Timestamp: time.Now(),
		Phase:     PhasePlanning,
		EventType: ProgressStart,
		Message:   "starting",
		Progress:  0.5,
		Current:   5,
		Total:     10,
	}

	handler(event)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Errorf("Failed to copy output: %v", err)
	}
	output := buf.String()

	// Verify it's valid JSON
	var parsed ProgressEvent
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed)
	if err != nil {
		t.Errorf("output should be valid JSON: %v", err)
	}

	if parsed.Phase != PhasePlanning {
		t.Errorf("expected phase %s, got %s", PhasePlanning, parsed.Phase)
	}
	if parsed.EventType != ProgressStart {
		t.Errorf("expected eventType %s, got %s", ProgressStart, parsed.EventType)
	}
	if parsed.Message != "starting" {
		t.Errorf("expected message 'starting', got %s", parsed.Message)
	}
}

func TestSilentProgressHandler(t *testing.T) {
	handler := SilentProgressHandler()

	// Should not panic or do anything
	handler(ProgressEvent{
		Phase:     PhasePlanning,
		EventType: ProgressStart,
		Message:   "test",
	})
}

func TestNewProgressAggregator(t *testing.T) {
	agg := NewProgressAggregator()

	if agg == nil {
		t.Fatal("expected aggregator to be created")
	}
	if agg.events == nil {
		t.Error("events should be initialized")
	}
	if agg.startTime.IsZero() {
		t.Error("startTime should be set")
	}
}

func TestProgressAggregator_Handler(t *testing.T) {
	agg := NewProgressAggregator()
	handler := agg.Handler()

	// Send some events
	events := []ProgressEvent{
		{EventType: ProgressPageStart, PageID: "page1"},
		{EventType: ProgressPageDone, PageID: "page1"},
		{EventType: ProgressPageStart, PageID: "page2"},
		{EventType: ProgressError, PageID: "page2"},
		{Phase: PhaseCompleted, EventType: ProgressComplete},
	}

	for _, e := range events {
		handler(e)
	}

	// Verify events were collected
	agg.mu.RLock()
	defer agg.mu.RUnlock()
	if len(agg.events) != 5 {
		t.Errorf("expected 5 events, got %d", len(agg.events))
	}
	if agg.endTime.IsZero() {
		t.Error("endTime should be set after complete event")
	}
}

func TestProgressAggregator_Summary(t *testing.T) {
	agg := NewProgressAggregator()
	handler := agg.Handler()

	// Simulate a pipeline run
	handler(ProgressEvent{EventType: ProgressPageStart, PageID: "page1"})
	handler(ProgressEvent{EventType: ProgressPageDone, PageID: "page1"})
	handler(ProgressEvent{EventType: ProgressPageStart, PageID: "page2"})
	handler(ProgressEvent{EventType: ProgressPageDone, PageID: "page2"})
	handler(ProgressEvent{EventType: ProgressPageStart, PageID: "page3"})
	handler(ProgressEvent{EventType: ProgressError, PageID: "page3"})
	handler(ProgressEvent{EventType: ProgressSkip, PageID: "page4"})
	handler(ProgressEvent{Phase: PhaseCompleted, EventType: ProgressComplete})

	summary := agg.Summary()

	if !strings.Contains(summary, "Started: 3") {
		t.Errorf("expected 'Started: 3' in summary: %s", summary)
	}
	if !strings.Contains(summary, "Completed: 2") {
		t.Errorf("expected 'Completed: 2' in summary: %s", summary)
	}
	if !strings.Contains(summary, "Failed: 1") {
		t.Errorf("expected 'Failed: 1' in summary: %s", summary)
	}
	if !strings.Contains(summary, "Skipped: 1") {
		t.Errorf("expected 'Skipped: 1' in summary: %s", summary)
	}
	if !strings.Contains(summary, "Duration:") {
		t.Errorf("expected 'Duration:' in summary: %s", summary)
	}
}

func TestProgressAggregator_PageResults(t *testing.T) {
	agg := NewProgressAggregator()
	handler := agg.Handler()

	handler(ProgressEvent{EventType: ProgressPageStart, PageID: "page1"})
	handler(ProgressEvent{EventType: ProgressPageDone, PageID: "page1"})
	handler(ProgressEvent{EventType: ProgressPageStart, PageID: "page2"})
	handler(ProgressEvent{EventType: ProgressError, PageID: "page2"})
	handler(ProgressEvent{EventType: ProgressSkip, PageID: "page3"})

	results := agg.PageResults()

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	if results["page1"] != ProgressPageDone {
		t.Errorf("expected page1 to be done, got %s", results["page1"])
	}
	if results["page2"] != ProgressError {
		t.Errorf("expected page2 to be error, got %s", results["page2"])
	}
	if results["page3"] != ProgressSkip {
		t.Errorf("expected page3 to be skipped, got %s", results["page3"])
	}
}

func TestProgressAggregator_Duration(t *testing.T) {
	agg := NewProgressAggregator()
	handler := agg.Handler()

	time.Sleep(10 * time.Millisecond)

	// Before completion
	duration1 := agg.Duration()
	if duration1 < 10*time.Millisecond {
		t.Error("duration should be at least 10ms")
	}

	// After completion
	handler(ProgressEvent{Phase: PhaseCompleted, EventType: ProgressComplete})
	duration2 := agg.Duration()

	// Should be roughly the same after completion
	if duration2 < duration1 {
		t.Error("duration after completion should be >= duration before")
	}
}

func TestProgressAggregator_Concurrent(t *testing.T) {
	agg := NewProgressAggregator()
	handler := agg.Handler()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			handler(ProgressEvent{
				EventType: ProgressPageStart,
				PageID:    string(rune('A' + id%26)),
			})
		}(i)
	}

	wg.Wait()

	// Should have all events without data race
	agg.mu.RLock()
	defer agg.mu.RUnlock()
	if len(agg.events) != 100 {
		t.Errorf("expected 100 events, got %d", len(agg.events))
	}
}

func TestCombinedProgressHandler(t *testing.T) {
	count1 := 0
	count2 := 0
	count3 := 0

	handler1 := func(_ ProgressEvent) { count1++ }
	handler2 := func(_ ProgressEvent) { count2++ }
	handler3 := func(_ ProgressEvent) { count3++ }

	combined := CombinedProgressHandler(handler1, handler2, nil, handler3) // Include nil

	event := ProgressEvent{EventType: ProgressStart}
	combined(event)
	combined(event)

	if count1 != 2 {
		t.Errorf("expected count1=2, got %d", count1)
	}
	if count2 != 2 {
		t.Errorf("expected count2=2, got %d", count2)
	}
	if count3 != 2 {
		t.Errorf("expected count3=2, got %d", count3)
	}
}

func TestCombinedProgressHandler_Empty(t *testing.T) {
	combined := CombinedProgressHandler()

	// Should not panic
	combined(ProgressEvent{EventType: ProgressStart})
}

func TestRenderProgressBar(t *testing.T) {
	tests := []struct {
		progress float64
		width    int
		expected string
	}{
		{0.0, 10, "[░░░░░░░░░░]"},
		{0.5, 10, "[█████░░░░░]"},
		{1.0, 10, "[██████████]"},
		{-0.5, 10, "[░░░░░░░░░░]"}, // Negative clamps to 0
		{1.5, 10, "[██████████]"},  // Over 1 clamps to 1
		{0.25, 20, "[█████░░░░░░░░░░░░░░░]"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := renderProgressBar(tt.progress, tt.width)
			if result != tt.expected {
				t.Errorf("renderProgressBar(%f, %d) = %s, want %s",
					tt.progress, tt.width, result, tt.expected)
			}
		})
	}
}

func TestFormatProgress(t *testing.T) {
	tests := []struct {
		current  int
		total    int
		expected string
	}{
		{0, 10, "0/10"},
		{5, 10, "5/10"},
		{10, 10, "10/10"},
		{100, 100, "100/100"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := formatProgress(tt.current, tt.total)
			if result != tt.expected {
				t.Errorf("formatProgress(%d, %d) = %s, want %s",
					tt.current, tt.total, result, tt.expected)
			}
		})
	}
}

func TestProgressAggregator_Summary_BeforeCompletion(t *testing.T) {
	agg := NewProgressAggregator()
	handler := agg.Handler()

	handler(ProgressEvent{EventType: ProgressPageStart, PageID: "page1"})
	handler(ProgressEvent{EventType: ProgressPageDone, PageID: "page1"})

	// Don't send complete event - endTime should be zero

	summary := agg.Summary()

	// Should still produce a summary with current duration
	if !strings.Contains(summary, "Duration:") {
		t.Errorf("expected 'Duration:' in summary: %s", summary)
	}
}

func TestProgressAggregator_PageResults_UpdatesOnMultipleEvents(t *testing.T) {
	agg := NewProgressAggregator()
	handler := agg.Handler()

	// Multiple events for same page - last one wins
	handler(ProgressEvent{EventType: ProgressPageStart, PageID: "page1"})
	handler(ProgressEvent{EventType: ProgressRetry, PageID: "page1"}) // Not tracked in PageResults
	handler(ProgressEvent{EventType: ProgressPageDone, PageID: "page1"})

	results := agg.PageResults()

	if results["page1"] != ProgressPageDone {
		t.Errorf("expected final status to be done, got %s", results["page1"])
	}
}

func TestProgressEvent_Fields(t *testing.T) {
	now := time.Now()
	event := ProgressEvent{
		Timestamp: now,
		Phase:     PhaseGenerating,
		EventType: ProgressPageDone,
		PageID:    "page1",
		PagePath:  "content/about.md",
		Message:   "completed",
		Progress:  0.75,
		Current:   3,
		Total:     4,
	}

	if event.Timestamp != now {
		t.Error("Timestamp should match")
	}
	if event.Phase != PhaseGenerating {
		t.Error("Phase should match")
	}
	if event.EventType != ProgressPageDone {
		t.Error("EventType should match")
	}
	if event.PageID != "page1" {
		t.Error("PageID should match")
	}
	if event.PagePath != "content/about.md" {
		t.Error("PagePath should match")
	}
	if event.Message != "completed" {
		t.Error("Message should match")
	}
	if event.Progress != 0.75 {
		t.Error("Progress should match")
	}
	if event.Current != 3 {
		t.Error("Current should match")
	}
	if event.Total != 4 {
		t.Error("Total should match")
	}
}
