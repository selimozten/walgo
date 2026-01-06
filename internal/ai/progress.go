package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/selimozten/walgo/internal/ui"
)

// =============================================================================
// Console Progress Handler
// =============================================================================

// ConsoleProgressHandler creates and returns a ProgressHandler that outputs to console.
func ConsoleProgressHandler(verbose bool) ProgressHandler {
	icons := ui.GetIcons()

	return func(event ProgressEvent) {
		switch event.EventType {
		case ProgressStart:
			switch event.Phase {
			case PhasePlanning:
				fmt.Printf("\n%s %s\n", icons.Robot, event.Message)
			case PhaseGenerating:
				fmt.Printf("\n%s %s\n", icons.Rocket, event.Message)
			}

		case ProgressPageStart:
			if verbose {
				fmt.Printf("   %s Generating: %s\n", icons.Spinner, event.PagePath)
			}

		case ProgressPageDone:
			progressBar := renderProgressBar(event.Progress, 20)
			fmt.Printf("   %s %s %s [%s]\n",
				icons.Check, event.PagePath, progressBar,
				formatProgress(event.Current+1, event.Total))

		case ProgressRetry:
			fmt.Printf("   %s %s: %s\n", icons.Warning, event.PagePath, event.Message)

		case ProgressSkip:
			if verbose {
				fmt.Printf("   %s %s (skipped: %s)\n", icons.Info, event.PagePath, event.Message)
			}

		case ProgressError:
			fmt.Printf("   %s %s: %s\n", icons.Error, event.PagePath, event.Message)

		case ProgressComplete:
			switch event.Phase {
			case PhasePlanning:
				fmt.Printf("%s %s\n", icons.Success, event.Message)
			case PhaseGenerating:
				fmt.Printf("\n%s Generation complete\n", icons.Success)
			case PhaseCompleted:
				fmt.Printf("\n%s %s\n", icons.Celebrate, event.Message)
			}
		}
	}
}

// =============================================================================
// JSON Progress Handler
// =============================================================================

// JSONProgressHandler creates and returns a ProgressHandler that outputs JSON lines.
func JSONProgressHandler() ProgressHandler {
	return func(event ProgressEvent) {
		data, _ := json.Marshal(event)
		fmt.Println(string(data))
	}
}

// =============================================================================
// Silent Progress Handler
// =============================================================================

// SilentProgressHandler creates and returns a no-op progress handler.
func SilentProgressHandler() ProgressHandler {
	return func(_ ProgressEvent) {}
}

// =============================================================================
// Progress Aggregator
// =============================================================================

// ProgressAggregator collects and stores progress events for analysis.
// Thread-safe for concurrent access.
type ProgressAggregator struct {
	mu        sync.RWMutex
	events    []ProgressEvent
	startTime time.Time
	endTime   time.Time
}

// NewProgressAggregator initializes and returns a new ProgressAggregator instance.
func NewProgressAggregator() *ProgressAggregator {
	return &ProgressAggregator{
		events:    make([]ProgressEvent, 0),
		startTime: time.Now(),
	}
}

// Handler returns a ProgressHandler that collects and stores events.
// Thread-safe for concurrent calls.
func (a *ProgressAggregator) Handler() ProgressHandler {
	return func(event ProgressEvent) {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.events = append(a.events, event)
		if event.EventType == ProgressComplete && event.Phase == PhaseCompleted {
			a.endTime = time.Now()
		}
	}
}

// Summary generates and returns a summary of collected progress events.
// Thread-safe for concurrent access.
func (a *ProgressAggregator) Summary() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var started, completed, failed, skipped int

	for _, e := range a.events {
		switch e.EventType {
		case ProgressPageStart:
			started++
		case ProgressPageDone:
			completed++
		case ProgressError:
			failed++
		case ProgressSkip:
			skipped++
		}
	}

	duration := a.endTime.Sub(a.startTime)
	if a.endTime.IsZero() {
		duration = time.Since(a.startTime)
	}

	return fmt.Sprintf("Started: %d, Completed: %d, Failed: %d, Skipped: %d, Duration: %v",
		started, completed, failed, skipped, duration.Round(time.Second))
}

// PageResults returns a map of page ID to final status.
// Thread-safe for concurrent access.
func (a *ProgressAggregator) PageResults() map[string]ProgressType {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := make(map[string]ProgressType)

	for _, e := range a.events {
		if e.PageID != "" {
			// Keep updating - last event for a page is its final status
			switch e.EventType {
			case ProgressPageDone, ProgressError, ProgressSkip:
				results[e.PageID] = e.EventType
			}
		}
	}

	return results
}

// Duration returns the total duration of collected events.
// Thread-safe for concurrent access.
func (a *ProgressAggregator) Duration() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.endTime.IsZero() {
		return time.Since(a.startTime)
	}
	return a.endTime.Sub(a.startTime)
}

// =============================================================================
// Combined Progress Handler
// =============================================================================

// CombinedProgressHandler creates a handler that calls multiple progress handlers.
func CombinedProgressHandler(handlers ...ProgressHandler) ProgressHandler {
	return func(event ProgressEvent) {
		for _, h := range handlers {
			if h != nil {
				h(event)
			}
		}
	}
}

// =============================================================================
// Helpers
// =============================================================================

// renderProgressBar generates a simple ASCII progress bar.
func renderProgressBar(progress float64, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filled := int(progress * float64(width))
	empty := width - filled

	if filled < 0 {
		filled = 0
	}
	if empty < 0 {
		empty = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("[%s]", bar)
}

// formatProgress formats the progress as "current/total".
func formatProgress(current, total int) string {
	return fmt.Sprintf("%d/%d", current, total)
}
