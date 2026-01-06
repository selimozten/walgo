package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/selimozten/walgo/internal/ui"
)

// Event represents a single telemetry event
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Command   string                 `json:"command"`
	Duration  int64                  `json:"duration_ms"` // Milliseconds
	Success   bool                   `json:"success"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BuildMetrics captures build-specific metrics
type BuildMetrics struct {
	HugoDuration       int64 `json:"hugo_duration_ms"`
	OptimizeDuration   int64 `json:"optimize_duration_ms,omitempty"`
	CompressDuration   int64 `json:"compress_duration_ms,omitempty"`
	TotalFiles         int   `json:"total_files"`
	CompressedFiles    int   `json:"compressed_files,omitempty"`
	CompressionSavings int64 `json:"compression_savings_bytes,omitempty"`
}

// DeployMetrics captures deployment-specific metrics
type DeployMetrics struct {
	TotalFiles     int     `json:"total_files"`
	ChangedFiles   int     `json:"changed_files,omitempty"`
	UploadDuration int64   `json:"upload_duration_ms"`
	CacheHitRatio  float64 `json:"cache_hit_ratio,omitempty"`
	BytesUploaded  int64   `json:"bytes_uploaded,omitempty"`
}

// Collector manages telemetry collection
type Collector struct {
	enabled   bool
	sinkPath  string
	sessionID string
	startTime time.Time
}

// defaultSinkPath returns the default path for telemetry data
func defaultSinkPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".walgo-metrics.json"
	}
	return filepath.Join(home, ".walgo", "metrics.json")
}

// New creates a new telemetry collector
func New(enabled bool) *Collector {
	return &Collector{
		enabled:   enabled,
		sinkPath:  defaultSinkPath(),
		sessionID: generateSessionID(),
		startTime: time.Now(),
	}
}

// NewWithSink creates a collector with a custom sink path
func NewWithSink(enabled bool, sinkPath string) *Collector {
	return &Collector{
		enabled:   enabled,
		sinkPath:  sinkPath,
		sessionID: generateSessionID(),
		startTime: time.Now(),
	}
}

// Start marks the beginning of a command execution
func (c *Collector) Start() time.Time {
	return time.Now()
}

// RecordBuild records build metrics
func (c *Collector) RecordBuild(startTime time.Time, metrics *BuildMetrics, success bool) error {
	if !c.enabled {
		return nil
	}

	duration := time.Since(startTime).Milliseconds()

	event := Event{
		Timestamp: startTime,
		Command:   "build",
		Duration:  duration,
		Success:   success,
		Metadata: map[string]interface{}{
			"hugo_duration_ms":          metrics.HugoDuration,
			"optimize_duration_ms":      metrics.OptimizeDuration,
			"compress_duration_ms":      metrics.CompressDuration,
			"total_files":               metrics.TotalFiles,
			"compressed_files":          metrics.CompressedFiles,
			"compression_savings_bytes": metrics.CompressionSavings,
		},
	}

	return c.writeEvent(event)
}

// RecordDeploy records deployment metrics
func (c *Collector) RecordDeploy(startTime time.Time, metrics *DeployMetrics, success bool) error {
	if !c.enabled {
		return nil
	}

	duration := time.Since(startTime).Milliseconds()

	event := Event{
		Timestamp: startTime,
		Command:   "deploy",
		Duration:  duration,
		Success:   success,
		Metadata: map[string]interface{}{
			"total_files":        metrics.TotalFiles,
			"changed_files":      metrics.ChangedFiles,
			"upload_duration_ms": metrics.UploadDuration,
			"cache_hit_ratio":    metrics.CacheHitRatio,
			"bytes_uploaded":     metrics.BytesUploaded,
		},
	}

	return c.writeEvent(event)
}

// RecordCommand records a generic command execution
func (c *Collector) RecordCommand(command string, startTime time.Time, success bool, metadata map[string]interface{}) error {
	if !c.enabled {
		return nil
	}

	duration := time.Since(startTime).Milliseconds()

	event := Event{
		Timestamp: startTime,
		Command:   command,
		Duration:  duration,
		Success:   success,
		Metadata:  metadata,
	}

	return c.writeEvent(event)
}

// writeEvent writes an event to the sink
func (c *Collector) writeEvent(event Event) error {
	// Ensure directory exists
	dir := filepath.Dir(c.sinkPath)
	// #nosec G301 - metrics directory needs standard permissions
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create metrics directory: %w", err)
	}

	// Read existing events
	var events []Event
	if data, err := os.ReadFile(c.sinkPath); err == nil {
		_ = json.Unmarshal(data, &events)
	}

	// Append new event
	events = append(events, event)

	// Keep only last 1000 events to prevent unbounded growth
	if len(events) > 1000 {
		events = events[len(events)-1000:]
	}

	// Write back
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	// #nosec G306 - metrics file can be readable
	if err := os.WriteFile(c.sinkPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write events: %w", err)
	}

	return nil
}

// generateSessionID generates a simple session identifier
func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().Unix())
}

// GetSinkPath returns the current sink path
func (c *Collector) GetSinkPath() string {
	return c.sinkPath
}

// IsEnabled returns whether telemetry is enabled
func (c *Collector) IsEnabled() bool {
	return c.enabled
}

// Stats provides statistics about collected metrics
type Stats struct {
	TotalEvents   int                `json:"total_events"`
	CommandCounts map[string]int     `json:"command_counts"`
	AvgDurations  map[string]int64   `json:"avg_durations_ms"`
	SuccessRates  map[string]float64 `json:"success_rates"`
	FirstEvent    time.Time          `json:"first_event,omitempty"`
	LastEvent     time.Time          `json:"last_event,omitempty"`
}

// GetStats reads and computes statistics from the telemetry data
func GetStats(sinkPath string) (*Stats, error) {
	data, err := os.ReadFile(sinkPath) // #nosec G304 - Reading user's local telemetry file is intended behavior
	if err != nil {
		return nil, err
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return &Stats{CommandCounts: make(map[string]int)}, nil
	}

	stats := &Stats{
		TotalEvents:   len(events),
		CommandCounts: make(map[string]int),
		AvgDurations:  make(map[string]int64),
		SuccessRates:  make(map[string]float64),
		FirstEvent:    events[0].Timestamp,
		LastEvent:     events[len(events)-1].Timestamp,
	}

	// Compute statistics
	commandDurations := make(map[string][]int64)
	commandSuccess := make(map[string][]bool)

	for _, event := range events {
		stats.CommandCounts[event.Command]++
		commandDurations[event.Command] = append(commandDurations[event.Command], event.Duration)
		commandSuccess[event.Command] = append(commandSuccess[event.Command], event.Success)
	}

	// Calculate averages
	for cmd, durations := range commandDurations {
		var sum int64
		for _, d := range durations {
			sum += d
		}
		stats.AvgDurations[cmd] = sum / int64(len(durations))
	}

	// Calculate success rates
	for cmd, successes := range commandSuccess {
		var successCount int
		for _, s := range successes {
			if s {
				successCount++
			}
		}
		stats.SuccessRates[cmd] = float64(successCount) / float64(len(successes)) * 100
	}

	return stats, nil
}

// PrintStats prints telemetry statistics in a human-readable format
func PrintStats(sinkPath string) error {
	icons := ui.GetIcons()
	stats, err := GetStats(sinkPath)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s Walgo Telemetry Statistics\n", icons.Chart)
	fmt.Println(ui.Separator())
	fmt.Printf("Total Events: %d\n", stats.TotalEvents)

	if !stats.FirstEvent.IsZero() {
		fmt.Printf("First Event: %s\n", stats.FirstEvent.Format("2006-01-02 15:04:05"))
		fmt.Printf("Last Event: %s\n", stats.LastEvent.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("\n%s Command Usage:\n", icons.Clipboard)
	for cmd, count := range stats.CommandCounts {
		avgDuration := stats.AvgDurations[cmd]
		successRate := stats.SuccessRates[cmd]
		fmt.Printf("  %s: %d times (avg: %dms, success: %.1f%%)\n",
			cmd, count, avgDuration, successRate)
	}

	return nil
}

// Clear removes all telemetry data
func Clear(sinkPath string) error {
	return os.Remove(sinkPath)
}
