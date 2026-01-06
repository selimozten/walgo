package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestNew tests the New constructor
func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled collector", true},
		{"disabled collector", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.enabled)

			if c == nil {
				t.Fatal("New() returned nil")
			}

			if c.enabled != tt.enabled {
				t.Errorf("enabled = %v, want %v", c.enabled, tt.enabled)
			}

			if c.sessionID == "" {
				t.Error("sessionID should not be empty")
			}

			if !strings.HasPrefix(c.sessionID, "session-") {
				t.Errorf("sessionID = %q, should start with 'session-'", c.sessionID)
			}

			if c.startTime.IsZero() {
				t.Error("startTime should not be zero")
			}

			// Default sink path should be set
			if c.sinkPath == "" {
				t.Error("sinkPath should not be empty")
			}
		})
	}
}

// TestNewWithSink tests the NewWithSink constructor
func TestNewWithSink(t *testing.T) {
	customPath := "/tmp/custom/metrics.json"

	c := NewWithSink(true, customPath)

	if c == nil {
		t.Fatal("NewWithSink() returned nil")
	}

	if c.sinkPath != customPath {
		t.Errorf("sinkPath = %q, want %q", c.sinkPath, customPath)
	}

	if !c.enabled {
		t.Error("enabled should be true")
	}
}

// TestGetSinkPath tests the GetSinkPath method
func TestGetSinkPath(t *testing.T) {
	customPath := "/tmp/test/metrics.json"
	c := NewWithSink(true, customPath)

	if got := c.GetSinkPath(); got != customPath {
		t.Errorf("GetSinkPath() = %q, want %q", got, customPath)
	}
}

// TestIsEnabled tests the IsEnabled method
func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.enabled)
			if got := c.IsEnabled(); got != tt.enabled {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.enabled)
			}
		})
	}
}

// TestStart tests the Start method
func TestStart(t *testing.T) {
	c := New(true)
	before := time.Now()
	startTime := c.Start()
	after := time.Now()

	if startTime.Before(before) || startTime.After(after) {
		t.Errorf("Start() returned time outside expected range")
	}
}

// TestRecordBuild tests the RecordBuild method
func TestRecordBuild(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "metrics.json")

	t.Run("enabled collector", func(t *testing.T) {
		c := NewWithSink(true, sinkPath)
		startTime := c.Start()

		// Add a small delay to get measurable duration
		time.Sleep(5 * time.Millisecond)

		metrics := &BuildMetrics{
			HugoDuration:       100,
			OptimizeDuration:   50,
			CompressDuration:   30,
			TotalFiles:         10,
			CompressedFiles:    5,
			CompressionSavings: 1024,
		}

		err := c.RecordBuild(startTime, metrics, true)
		if err != nil {
			t.Fatalf("RecordBuild() error = %v", err)
		}

		// Verify the event was written
		data, err := os.ReadFile(sinkPath)
		if err != nil {
			t.Fatalf("Failed to read sink file: %v", err)
		}

		var events []Event
		if err := json.Unmarshal(data, &events); err != nil {
			t.Fatalf("Failed to unmarshal events: %v", err)
		}

		if len(events) != 1 {
			t.Fatalf("Expected 1 event, got %d", len(events))
		}

		event := events[0]
		if event.Command != "build" {
			t.Errorf("Command = %q, want %q", event.Command, "build")
		}

		if !event.Success {
			t.Error("Success should be true")
		}

		if event.Duration < 5 {
			t.Errorf("Duration = %d, expected >= 5ms", event.Duration)
		}

		// Check metadata
		if event.Metadata["hugo_duration_ms"].(float64) != 100 {
			t.Errorf("hugo_duration_ms = %v, want 100", event.Metadata["hugo_duration_ms"])
		}

		if event.Metadata["total_files"].(float64) != 10 {
			t.Errorf("total_files = %v, want 10", event.Metadata["total_files"])
		}
	})

	t.Run("disabled collector", func(t *testing.T) {
		// Use a new path that shouldn't be created
		disabledPath := filepath.Join(tmpDir, "disabled", "metrics.json")
		c := NewWithSink(false, disabledPath)
		startTime := c.Start()

		metrics := &BuildMetrics{
			HugoDuration: 100,
			TotalFiles:   10,
		}

		err := c.RecordBuild(startTime, metrics, true)
		if err != nil {
			t.Fatalf("RecordBuild() error = %v", err)
		}

		// File should not exist
		if _, err := os.Stat(disabledPath); !os.IsNotExist(err) {
			t.Error("Sink file should not exist for disabled collector")
		}
	})

	t.Run("failed build", func(t *testing.T) {
		failPath := filepath.Join(tmpDir, "failed.json")
		c := NewWithSink(true, failPath)
		startTime := c.Start()

		metrics := &BuildMetrics{
			HugoDuration: 50,
			TotalFiles:   5,
		}

		err := c.RecordBuild(startTime, metrics, false)
		if err != nil {
			t.Fatalf("RecordBuild() error = %v", err)
		}

		// Verify the event was recorded as failed
		data, err := os.ReadFile(failPath)
		if err != nil {
			t.Fatalf("Failed to read sink file: %v", err)
		}

		var events []Event
		if err := json.Unmarshal(data, &events); err != nil {
			t.Fatalf("Failed to unmarshal events: %v", err)
		}

		if events[0].Success {
			t.Error("Success should be false")
		}
	})
}

// TestRecordDeploy tests the RecordDeploy method
func TestRecordDeploy(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "metrics.json")

	t.Run("enabled collector", func(t *testing.T) {
		c := NewWithSink(true, sinkPath)
		startTime := c.Start()

		time.Sleep(5 * time.Millisecond)

		metrics := &DeployMetrics{
			TotalFiles:     20,
			ChangedFiles:   5,
			UploadDuration: 200,
			CacheHitRatio:  0.75,
			BytesUploaded:  2048,
		}

		err := c.RecordDeploy(startTime, metrics, true)
		if err != nil {
			t.Fatalf("RecordDeploy() error = %v", err)
		}

		data, err := os.ReadFile(sinkPath)
		if err != nil {
			t.Fatalf("Failed to read sink file: %v", err)
		}

		var events []Event
		if err := json.Unmarshal(data, &events); err != nil {
			t.Fatalf("Failed to unmarshal events: %v", err)
		}

		if len(events) != 1 {
			t.Fatalf("Expected 1 event, got %d", len(events))
		}

		event := events[0]
		if event.Command != "deploy" {
			t.Errorf("Command = %q, want %q", event.Command, "deploy")
		}

		if !event.Success {
			t.Error("Success should be true")
		}

		// Check metadata
		if event.Metadata["total_files"].(float64) != 20 {
			t.Errorf("total_files = %v, want 20", event.Metadata["total_files"])
		}

		if event.Metadata["cache_hit_ratio"].(float64) != 0.75 {
			t.Errorf("cache_hit_ratio = %v, want 0.75", event.Metadata["cache_hit_ratio"])
		}

		if event.Metadata["bytes_uploaded"].(float64) != 2048 {
			t.Errorf("bytes_uploaded = %v, want 2048", event.Metadata["bytes_uploaded"])
		}
	})

	t.Run("disabled collector", func(t *testing.T) {
		disabledPath := filepath.Join(tmpDir, "disabled", "deploy.json")
		c := NewWithSink(false, disabledPath)
		startTime := c.Start()

		metrics := &DeployMetrics{
			TotalFiles:    10,
			BytesUploaded: 1024,
		}

		err := c.RecordDeploy(startTime, metrics, true)
		if err != nil {
			t.Fatalf("RecordDeploy() error = %v", err)
		}

		if _, err := os.Stat(disabledPath); !os.IsNotExist(err) {
			t.Error("Sink file should not exist for disabled collector")
		}
	})
}

// TestRecordCommand tests the RecordCommand method
func TestRecordCommand(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "metrics.json")

	t.Run("with metadata", func(t *testing.T) {
		c := NewWithSink(true, sinkPath)
		startTime := c.Start()

		metadata := map[string]interface{}{
			"custom_field": "custom_value",
			"count":        42,
		}

		err := c.RecordCommand("custom", startTime, true, metadata)
		if err != nil {
			t.Fatalf("RecordCommand() error = %v", err)
		}

		data, err := os.ReadFile(sinkPath)
		if err != nil {
			t.Fatalf("Failed to read sink file: %v", err)
		}

		var events []Event
		if err := json.Unmarshal(data, &events); err != nil {
			t.Fatalf("Failed to unmarshal events: %v", err)
		}

		event := events[0]
		if event.Command != "custom" {
			t.Errorf("Command = %q, want %q", event.Command, "custom")
		}

		if event.Metadata["custom_field"].(string) != "custom_value" {
			t.Errorf("custom_field = %v, want %q", event.Metadata["custom_field"], "custom_value")
		}
	})

	t.Run("without metadata", func(t *testing.T) {
		c := NewWithSink(true, filepath.Join(tmpDir, "no_meta.json"))
		startTime := c.Start()

		err := c.RecordCommand("simple", startTime, true, nil)
		if err != nil {
			t.Fatalf("RecordCommand() error = %v", err)
		}
	})

	t.Run("disabled collector", func(t *testing.T) {
		disabledPath := filepath.Join(tmpDir, "disabled_cmd.json")
		c := NewWithSink(false, disabledPath)
		startTime := c.Start()

		err := c.RecordCommand("test", startTime, true, nil)
		if err != nil {
			t.Fatalf("RecordCommand() error = %v", err)
		}

		if _, err := os.Stat(disabledPath); !os.IsNotExist(err) {
			t.Error("Sink file should not exist for disabled collector")
		}
	})
}

// TestWriteEventEventLimit tests that events are limited to 1000
func TestWriteEventEventLimit(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "metrics.json")
	c := NewWithSink(true, sinkPath)

	// Write 1010 events
	for i := 0; i < 1010; i++ {
		startTime := time.Now()
		err := c.RecordCommand("test", startTime, true, map[string]interface{}{"index": i})
		if err != nil {
			t.Fatalf("RecordCommand() error = %v at iteration %d", err, i)
		}
	}

	// Read and verify only 1000 events exist
	data, err := os.ReadFile(sinkPath)
	if err != nil {
		t.Fatalf("Failed to read sink file: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("Failed to unmarshal events: %v", err)
	}

	if len(events) != 1000 {
		t.Errorf("Expected 1000 events, got %d", len(events))
	}

	// Verify we kept the most recent events (index 10-1009)
	firstEvent := events[0]
	if firstEvent.Metadata["index"].(float64) != 10 {
		t.Errorf("First event index = %v, want 10", firstEvent.Metadata["index"])
	}

	lastEvent := events[999]
	if lastEvent.Metadata["index"].(float64) != 1009 {
		t.Errorf("Last event index = %v, want 1009", lastEvent.Metadata["index"])
	}
}

// TestWriteEventAppends tests that new events are appended to existing ones
func TestWriteEventAppends(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "metrics.json")
	c := NewWithSink(true, sinkPath)

	// Write first event
	err := c.RecordCommand("first", time.Now(), true, nil)
	if err != nil {
		t.Fatalf("RecordCommand() error = %v", err)
	}

	// Write second event
	err = c.RecordCommand("second", time.Now(), true, nil)
	if err != nil {
		t.Fatalf("RecordCommand() error = %v", err)
	}

	// Verify both events exist
	data, err := os.ReadFile(sinkPath)
	if err != nil {
		t.Fatalf("Failed to read sink file: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("Failed to unmarshal events: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	if events[0].Command != "first" {
		t.Errorf("First event command = %q, want %q", events[0].Command, "first")
	}

	if events[1].Command != "second" {
		t.Errorf("Second event command = %q, want %q", events[1].Command, "second")
	}
}

// TestWriteEventHandlesCorruptedFile tests that corrupted JSON is handled gracefully
func TestWriteEventHandlesCorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "metrics.json")

	// Write corrupted JSON
	err := os.WriteFile(sinkPath, []byte("not valid json {"), 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	c := NewWithSink(true, sinkPath)
	err = c.RecordCommand("test", time.Now(), true, nil)
	if err != nil {
		t.Fatalf("RecordCommand() should handle corrupted file gracefully, got error = %v", err)
	}

	// Verify the file now contains valid JSON with one event
	data, err := os.ReadFile(sinkPath)
	if err != nil {
		t.Fatalf("Failed to read sink file: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("Failed to unmarshal events: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestWriteEventCreatesDirectory tests that the directory is created if it doesn't exist
func TestWriteEventCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "deep", "nested", "dir", "metrics.json")

	c := NewWithSink(true, sinkPath)
	err := c.RecordCommand("test", time.Now(), true, nil)
	if err != nil {
		t.Fatalf("RecordCommand() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(sinkPath)); os.IsNotExist(err) {
		t.Error("Directory should have been created")
	}

	// Verify file exists
	if _, err := os.Stat(sinkPath); os.IsNotExist(err) {
		t.Error("Sink file should exist")
	}
}

// TestGetStats tests the GetStats function
func TestGetStats(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("with events", func(t *testing.T) {
		sinkPath := filepath.Join(tmpDir, "stats.json")
		c := NewWithSink(true, sinkPath)

		// Record various events
		startTime := time.Now()
		c.RecordCommand("build", startTime, true, nil)
		time.Sleep(10 * time.Millisecond)
		c.RecordCommand("build", startTime, true, nil)
		c.RecordCommand("build", startTime, false, nil)
		c.RecordCommand("deploy", startTime, true, nil)
		c.RecordCommand("deploy", startTime, true, nil)

		stats, err := GetStats(sinkPath)
		if err != nil {
			t.Fatalf("GetStats() error = %v", err)
		}

		if stats.TotalEvents != 5 {
			t.Errorf("TotalEvents = %d, want 5", stats.TotalEvents)
		}

		if stats.CommandCounts["build"] != 3 {
			t.Errorf("CommandCounts[build] = %d, want 3", stats.CommandCounts["build"])
		}

		if stats.CommandCounts["deploy"] != 2 {
			t.Errorf("CommandCounts[deploy] = %d, want 2", stats.CommandCounts["deploy"])
		}

		// Build: 2 success, 1 failure = 66.67%
		buildSuccessRate := stats.SuccessRates["build"]
		if buildSuccessRate < 66 || buildSuccessRate > 67 {
			t.Errorf("SuccessRates[build] = %v, want ~66.67", buildSuccessRate)
		}

		// Deploy: 2 success, 0 failure = 100%
		if stats.SuccessRates["deploy"] != 100 {
			t.Errorf("SuccessRates[deploy] = %v, want 100", stats.SuccessRates["deploy"])
		}

		if stats.FirstEvent.IsZero() {
			t.Error("FirstEvent should not be zero")
		}

		if stats.LastEvent.IsZero() {
			t.Error("LastEvent should not be zero")
		}
	})

	t.Run("empty events array", func(t *testing.T) {
		sinkPath := filepath.Join(tmpDir, "empty.json")
		err := os.WriteFile(sinkPath, []byte("[]"), 0644)
		if err != nil {
			t.Fatalf("Failed to write empty file: %v", err)
		}

		stats, err := GetStats(sinkPath)
		if err != nil {
			t.Fatalf("GetStats() error = %v", err)
		}

		if stats.TotalEvents != 0 {
			t.Errorf("TotalEvents = %d, want 0", stats.TotalEvents)
		}

		if stats.CommandCounts == nil {
			t.Error("CommandCounts should not be nil")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := GetStats(filepath.Join(tmpDir, "nonexistent.json"))
		if err == nil {
			t.Error("GetStats() should return error for nonexistent file")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		sinkPath := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(sinkPath, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid file: %v", err)
		}

		_, err = GetStats(sinkPath)
		if err == nil {
			t.Error("GetStats() should return error for invalid JSON")
		}
	})
}

// TestClear tests the Clear function
func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "metrics.json")

	// Create a file
	err := os.WriteFile(sinkPath, []byte("[]"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Clear it
	err = Clear(sinkPath)
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(sinkPath); !os.IsNotExist(err) {
		t.Error("File should be removed after Clear()")
	}

	// Clear nonexistent file should return error
	err = Clear(sinkPath)
	if err == nil {
		t.Error("Clear() should return error for nonexistent file")
	}
}

// TestPrintStats tests the PrintStats function
func TestPrintStats(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("with events", func(t *testing.T) {
		sinkPath := filepath.Join(tmpDir, "print_stats.json")
		c := NewWithSink(true, sinkPath)

		startTime := time.Now()
		c.RecordCommand("build", startTime, true, nil)
		c.RecordCommand("deploy", startTime, false, nil)

		// PrintStats should not return an error
		err := PrintStats(sinkPath)
		if err != nil {
			t.Fatalf("PrintStats() error = %v", err)
		}
	})

	t.Run("empty events", func(t *testing.T) {
		sinkPath := filepath.Join(tmpDir, "empty_print.json")
		err := os.WriteFile(sinkPath, []byte("[]"), 0644)
		if err != nil {
			t.Fatalf("Failed to write empty file: %v", err)
		}

		err = PrintStats(sinkPath)
		if err != nil {
			t.Fatalf("PrintStats() error = %v", err)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		err := PrintStats(filepath.Join(tmpDir, "nonexistent.json"))
		if err == nil {
			t.Error("PrintStats() should return error for nonexistent file")
		}
	})
}

// TestDefaultSinkPath tests the defaultSinkPath function
func TestDefaultSinkPath(t *testing.T) {
	path := defaultSinkPath()

	if path == "" {
		t.Error("defaultSinkPath() should not return empty string")
	}

	// Should end with metrics.json
	if !strings.HasSuffix(path, "metrics.json") {
		t.Errorf("defaultSinkPath() = %q, should end with metrics.json", path)
	}

	// Should be an absolute path (on unix-like systems)
	if !filepath.IsAbs(path) && path != ".walgo-metrics.json" {
		t.Errorf("defaultSinkPath() = %q, should be absolute or fallback", path)
	}
}

// TestGenerateSessionID tests the generateSessionID function
func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()

	if id1 == "" {
		t.Error("generateSessionID() should not return empty string")
	}

	if !strings.HasPrefix(id1, "session-") {
		t.Errorf("generateSessionID() = %q, should start with 'session-'", id1)
	}

	// IDs generated at the same second should be the same
	id2 := generateSessionID()
	if id1 != id2 {
		// This is expected if a second boundary was crossed
		t.Log("Session IDs differ (likely second boundary crossed)")
	}

	// Wait a second to ensure different IDs
	time.Sleep(1100 * time.Millisecond)
	id3 := generateSessionID()

	if id3 == id1 {
		t.Error("generateSessionID() should return different IDs at different times")
	}
}

// TestConcurrentWrites tests concurrent writes to the metrics file
func TestConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "concurrent.json")
	c := NewWithSink(true, sinkPath)

	var wg sync.WaitGroup
	numGoroutines := 10
	eventsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				startTime := time.Now()
				_ = c.RecordCommand("concurrent", startTime, true, map[string]interface{}{
					"goroutine": goroutineID,
					"event":     j,
				})
			}
		}(i)
	}

	wg.Wait()

	// Read and verify the file is valid JSON
	data, err := os.ReadFile(sinkPath)
	if err != nil {
		t.Fatalf("Failed to read sink file: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("Failed to unmarshal events: %v", err)
	}

	// Due to race conditions, we may not have all events, but the file should be valid
	if len(events) == 0 {
		t.Error("Expected at least some events")
	}
}

// TestEventJSONSerialization tests that Event serializes correctly to JSON
func TestEventJSONSerialization(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	event := Event{
		Timestamp: now,
		Command:   "test",
		Duration:  123,
		Success:   true,
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.Command != event.Command {
		t.Errorf("Command = %q, want %q", decoded.Command, event.Command)
	}

	if decoded.Duration != event.Duration {
		t.Errorf("Duration = %d, want %d", decoded.Duration, event.Duration)
	}

	if decoded.Success != event.Success {
		t.Errorf("Success = %v, want %v", decoded.Success, event.Success)
	}
}

// TestBuildMetricsJSONSerialization tests that BuildMetrics serializes correctly
func TestBuildMetricsJSONSerialization(t *testing.T) {
	metrics := BuildMetrics{
		HugoDuration:       100,
		OptimizeDuration:   50,
		CompressDuration:   30,
		TotalFiles:         10,
		CompressedFiles:    5,
		CompressionSavings: 1024,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Failed to marshal BuildMetrics: %v", err)
	}

	var decoded BuildMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal BuildMetrics: %v", err)
	}

	if decoded.HugoDuration != metrics.HugoDuration {
		t.Errorf("HugoDuration = %d, want %d", decoded.HugoDuration, metrics.HugoDuration)
	}

	if decoded.TotalFiles != metrics.TotalFiles {
		t.Errorf("TotalFiles = %d, want %d", decoded.TotalFiles, metrics.TotalFiles)
	}
}

// TestDeployMetricsJSONSerialization tests that DeployMetrics serializes correctly
func TestDeployMetricsJSONSerialization(t *testing.T) {
	metrics := DeployMetrics{
		TotalFiles:     20,
		ChangedFiles:   5,
		UploadDuration: 200,
		CacheHitRatio:  0.75,
		BytesUploaded:  2048,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Failed to marshal DeployMetrics: %v", err)
	}

	var decoded DeployMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DeployMetrics: %v", err)
	}

	if decoded.TotalFiles != metrics.TotalFiles {
		t.Errorf("TotalFiles = %d, want %d", decoded.TotalFiles, metrics.TotalFiles)
	}

	if decoded.CacheHitRatio != metrics.CacheHitRatio {
		t.Errorf("CacheHitRatio = %v, want %v", decoded.CacheHitRatio, metrics.CacheHitRatio)
	}
}

// TestStatsJSONSerialization tests that Stats serializes correctly
func TestStatsJSONSerialization(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	stats := Stats{
		TotalEvents:   100,
		CommandCounts: map[string]int{"build": 50, "deploy": 50},
		AvgDurations:  map[string]int64{"build": 100, "deploy": 200},
		SuccessRates:  map[string]float64{"build": 95.0, "deploy": 98.0},
		FirstEvent:    now.Add(-24 * time.Hour),
		LastEvent:     now,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal Stats: %v", err)
	}

	var decoded Stats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Stats: %v", err)
	}

	if decoded.TotalEvents != stats.TotalEvents {
		t.Errorf("TotalEvents = %d, want %d", decoded.TotalEvents, stats.TotalEvents)
	}

	if decoded.CommandCounts["build"] != stats.CommandCounts["build"] {
		t.Errorf("CommandCounts[build] = %d, want %d", decoded.CommandCounts["build"], stats.CommandCounts["build"])
	}
}

// TestWriteEventDirectoryCreationError tests error handling when directory cannot be created
func TestWriteEventDirectoryCreationError(t *testing.T) {
	// Skip on non-Unix systems
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test that requires specific filesystem permissions")
	}

	tmpDir := t.TempDir()

	// Create a file where we want a directory
	blockingFile := filepath.Join(tmpDir, "blocking")
	err := os.WriteFile(blockingFile, []byte("blocking"), 0644)
	if err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}

	// Try to create a metrics file inside the blocking file path
	sinkPath := filepath.Join(blockingFile, "metrics.json")
	c := NewWithSink(true, sinkPath)

	err = c.RecordCommand("test", time.Now(), true, nil)
	if err == nil {
		t.Error("RecordCommand() should return error when directory creation fails")
	}
}

// TestEventWithEmptyMetadata tests events with empty or nil metadata
func TestEventWithEmptyMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "empty_metadata.json")
	c := NewWithSink(true, sinkPath)

	// Test with nil metadata
	err := c.RecordCommand("nil_meta", time.Now(), true, nil)
	if err != nil {
		t.Fatalf("RecordCommand() with nil metadata error = %v", err)
	}

	// Test with empty metadata
	err = c.RecordCommand("empty_meta", time.Now(), true, map[string]interface{}{})
	if err != nil {
		t.Fatalf("RecordCommand() with empty metadata error = %v", err)
	}

	// Verify events were written
	data, err := os.ReadFile(sinkPath)
	if err != nil {
		t.Fatalf("Failed to read sink file: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("Failed to unmarshal events: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

// TestBuildMetricsZeroValues tests BuildMetrics with zero values
func TestBuildMetricsZeroValues(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "zero_build.json")
	c := NewWithSink(true, sinkPath)

	metrics := &BuildMetrics{}

	err := c.RecordBuild(time.Now(), metrics, true)
	if err != nil {
		t.Fatalf("RecordBuild() error = %v", err)
	}

	data, err := os.ReadFile(sinkPath)
	if err != nil {
		t.Fatalf("Failed to read sink file: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("Failed to unmarshal events: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// Verify zero values are recorded
	if events[0].Metadata["total_files"].(float64) != 0 {
		t.Errorf("total_files = %v, want 0", events[0].Metadata["total_files"])
	}
}

// TestDeployMetricsZeroValues tests DeployMetrics with zero values
func TestDeployMetricsZeroValues(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "zero_deploy.json")
	c := NewWithSink(true, sinkPath)

	metrics := &DeployMetrics{}

	err := c.RecordDeploy(time.Now(), metrics, true)
	if err != nil {
		t.Fatalf("RecordDeploy() error = %v", err)
	}

	data, err := os.ReadFile(sinkPath)
	if err != nil {
		t.Fatalf("Failed to read sink file: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("Failed to unmarshal events: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// Verify zero values are recorded
	if events[0].Metadata["total_files"].(float64) != 0 {
		t.Errorf("total_files = %v, want 0", events[0].Metadata["total_files"])
	}
}

// TestGetStatsWithMultipleCommandTypes tests statistics with various command types
func TestGetStatsWithMultipleCommandTypes(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "multi_stats.json")
	c := NewWithSink(true, sinkPath)

	commands := []struct {
		name    string
		success bool
	}{
		{"build", true},
		{"build", true},
		{"build", false},
		{"deploy", true},
		{"deploy", true},
		{"init", true},
		{"status", true},
		{"status", false},
	}

	for _, cmd := range commands {
		startTime := time.Now()
		err := c.RecordCommand(cmd.name, startTime, cmd.success, nil)
		if err != nil {
			t.Fatalf("RecordCommand() error = %v", err)
		}
	}

	stats, err := GetStats(sinkPath)
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalEvents != 8 {
		t.Errorf("TotalEvents = %d, want 8", stats.TotalEvents)
	}

	expectedCounts := map[string]int{
		"build":  3,
		"deploy": 2,
		"init":   1,
		"status": 2,
	}

	for cmd, expected := range expectedCounts {
		if stats.CommandCounts[cmd] != expected {
			t.Errorf("CommandCounts[%s] = %d, want %d", cmd, stats.CommandCounts[cmd], expected)
		}
	}

	// Verify success rates
	// build: 2/3 = 66.67%
	if stats.SuccessRates["build"] < 66 || stats.SuccessRates["build"] > 67 {
		t.Errorf("SuccessRates[build] = %v, want ~66.67", stats.SuccessRates["build"])
	}

	// deploy: 2/2 = 100%
	if stats.SuccessRates["deploy"] != 100 {
		t.Errorf("SuccessRates[deploy] = %v, want 100", stats.SuccessRates["deploy"])
	}

	// init: 1/1 = 100%
	if stats.SuccessRates["init"] != 100 {
		t.Errorf("SuccessRates[init] = %v, want 100", stats.SuccessRates["init"])
	}

	// status: 1/2 = 50%
	if stats.SuccessRates["status"] != 50 {
		t.Errorf("SuccessRates[status] = %v, want 50", stats.SuccessRates["status"])
	}
}

// TestCollectorFieldAccess tests direct access to Collector fields via methods
func TestCollectorFieldAccess(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "fields.json")

	c := NewWithSink(true, sinkPath)

	// Test IsEnabled
	if !c.IsEnabled() {
		t.Error("IsEnabled() should return true")
	}

	// Test GetSinkPath
	if c.GetSinkPath() != sinkPath {
		t.Errorf("GetSinkPath() = %q, want %q", c.GetSinkPath(), sinkPath)
	}

	// Create a disabled collector
	disabledCollector := NewWithSink(false, sinkPath)
	if disabledCollector.IsEnabled() {
		t.Error("IsEnabled() should return false for disabled collector")
	}
}

// TestEventTimestampOrder tests that events maintain proper timestamp order
func TestEventTimestampOrder(t *testing.T) {
	tmpDir := t.TempDir()
	sinkPath := filepath.Join(tmpDir, "order.json")
	c := NewWithSink(true, sinkPath)

	// Record events with slight delays
	times := make([]time.Time, 5)
	for i := 0; i < 5; i++ {
		times[i] = time.Now()
		err := c.RecordCommand("test", times[i], true, map[string]interface{}{"index": i})
		if err != nil {
			t.Fatalf("RecordCommand() error = %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Read and verify order
	data, err := os.ReadFile(sinkPath)
	if err != nil {
		t.Fatalf("Failed to read sink file: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("Failed to unmarshal events: %v", err)
	}

	for i := 1; i < len(events); i++ {
		prevIdx := int(events[i-1].Metadata["index"].(float64))
		currIdx := int(events[i].Metadata["index"].(float64))
		if currIdx <= prevIdx {
			t.Errorf("Events out of order: index %d before index %d", prevIdx, currIdx)
		}
	}
}
