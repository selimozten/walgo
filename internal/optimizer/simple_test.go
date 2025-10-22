package optimizer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test the types.go file functions
func TestNewDefaultOptimizerConfig(t *testing.T) {
	cfg := NewDefaultOptimizerConfig()

	if !cfg.Enabled {
		t.Error("Expected optimizer to be enabled by default")
	}

	if !cfg.HTML.Enabled {
		t.Error("Expected HTML optimizer to be enabled by default")
	}

	if !cfg.CSS.Enabled {
		t.Error("Expected CSS optimizer to be enabled by default")
	}

	if !cfg.JS.Enabled {
		t.Error("Expected JS optimizer to be enabled by default")
	}

	if len(cfg.SkipPatterns) == 0 {
		t.Error("Expected default skip patterns")
	}
}

// Simple tests for HTML optimizer
func TestHTMLOptimizerBasic(t *testing.T) {
	config := HTMLConfig{
		Enabled:          true,
		MinifyHTML:       true,
		RemoveComments:   true,
		RemoveWhitespace: true,
	}

	optimizer := NewHTMLOptimizer(config)
	if optimizer == nil {
		t.Fatal("NewHTMLOptimizer returned nil")
	}

	input := []byte(`<html>
		<!-- Comment -->
		<body>
			<h1>Test</h1>
		</body>
	</html>`)

	output, err := optimizer.Optimize(input)
	if err != nil {
		t.Fatalf("Optimize returned error: %v", err)
	}

	if len(output) >= len(input) {
		t.Error("Expected output to be smaller than input")
	}
}

// Simple tests for CSS optimizer
func TestCSSOptimizerBasic(t *testing.T) {
	config := CSSConfig{
		Enabled:        true,
		MinifyCSS:      true,
		RemoveComments: true,
	}

	optimizer := NewCSSOptimizer(config)
	if optimizer == nil {
		t.Fatal("NewCSSOptimizer returned nil")
	}

	input := []byte(`/* Comment */
	body {
		margin: 0;
		padding: 0;
	}`)

	output, err := optimizer.Optimize(input)
	if err != nil {
		t.Fatalf("Optimize returned error: %v", err)
	}

	if len(output) >= len(input) {
		t.Error("Expected output to be smaller than input")
	}
}

// Simple tests for JS optimizer
func TestJSOptimizerBasic(t *testing.T) {
	config := JSConfig{
		Enabled:        true,
		MinifyJS:       true,
		RemoveComments: true,
	}

	optimizer := NewJSOptimizer(config)
	if optimizer == nil {
		t.Fatal("NewJSOptimizer returned nil")
	}

	input := []byte(`// Comment
	function test() {
		console.log("Hello");
	}`)

	output, err := optimizer.Optimize(input)
	if err != nil {
		t.Fatalf("Optimize returned error: %v", err)
	}

	if len(output) >= len(input) {
		t.Error("Expected output to be smaller than input")
	}
}

// Test the main engine with a simple directory
func TestEngineOptimizeDirectorySimple(t *testing.T) {
	// Create test directory
	testDir := t.TempDir()

	// Create test files
	htmlFile := filepath.Join(testDir, "index.html")
	os.WriteFile(htmlFile, []byte(`<html><body>Test</body></html>`), 0644)

	cssFile := filepath.Join(testDir, "style.css")
	os.WriteFile(cssFile, []byte(`body { margin: 0; }`), 0644)

	jsFile := filepath.Join(testDir, "script.js")
	os.WriteFile(jsFile, []byte(`function test() { return true; }`), 0644)

	// Create engine with all optimizers enabled
	config := NewDefaultOptimizerConfig()
	engine := NewEngine(config)

	// Optimize directory
	stats, err := engine.OptimizeDirectory(testDir)
	if err != nil {
		t.Fatalf("OptimizeDirectory failed: %v", err)
	}

	if stats.FilesProcessed == 0 {
		t.Error("No files were processed")
	}
}

// Test disabled optimizer
func TestEngineDisabled(t *testing.T) {
	testDir := t.TempDir()

	// Create a test file
	htmlFile := filepath.Join(testDir, "index.html")
	os.WriteFile(htmlFile, []byte(`<html><body>Test</body></html>`), 0644)

	// Create engine with optimizer disabled
	config := OptimizerConfig{
		Enabled: false,
	}
	engine := NewEngine(config)

	// Optimize directory
	stats, err := engine.OptimizeDirectory(testDir)
	if err != nil {
		t.Fatalf("OptimizeDirectory failed: %v", err)
	}

	if stats.FilesProcessed != 0 {
		t.Error("Files should not be processed when optimizer is disabled")
	}
}

// Test skip patterns
func TestEngineSkipPatterns(t *testing.T) {
	testDir := t.TempDir()

	// Create test files
	minFile := filepath.Join(testDir, "script.min.js")
	os.WriteFile(minFile, []byte(`function test(){}`), 0644)

	regularFile := filepath.Join(testDir, "script.js")
	os.WriteFile(regularFile, []byte(`function test() { return true; }`), 0644)

	// Create engine with skip patterns
	config := NewDefaultOptimizerConfig()
	engine := NewEngine(config)

	// The default config includes *.min.js in skip patterns
	stats, err := engine.OptimizeDirectory(testDir)
	if err != nil {
		t.Fatalf("OptimizeDirectory failed: %v", err)
	}

	if stats.FilesSkipped == 0 {
		t.Error("Expected some files to be skipped")
	}
}

// Test PrintStats
func TestEnginePrintStats(t *testing.T) {
	engine := NewEngine(NewDefaultOptimizerConfig())
	stats := &OptimizationStats{
		FilesProcessed: 10,
		FilesOptimized: 8,
		FilesSkipped:   2,
		OriginalSize:   10000,
		OptimizedSize:  8000,
		SavingsBytes:   2000,
		SavingsPercent: 20.0,
		Duration:       1 * time.Second,
	}

	// Just make sure it doesn't panic
	engine.PrintStats(stats)
}

// Test formatBytes function
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
	}

	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, got, tt.want)
		}
	}
}