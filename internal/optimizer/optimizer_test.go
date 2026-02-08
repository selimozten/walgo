package optimizer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

// Test CSS optimizer preserves decimal values correctly
func TestCSSOptimizerDecimalValues(t *testing.T) {
	config := CSSConfig{
		Enabled:   true,
		MinifyCSS: true,
	}

	optimizer := NewCSSOptimizer(config)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "rgba with small decimal",
			input:    "color: rgba(255, 255, 255, 0.08);",
			expected: "color:rgba(255,255,255,.08);",
		},
		{
			name:     "letter-spacing with small negative decimal",
			input:    "letter-spacing: -0.03em;",
			expected: "letter-spacing:-.03em;",
		},
		{
			name:     "letter-spacing with small positive decimal",
			input:    "letter-spacing: 0.08em;",
			expected: "letter-spacing:.08em;",
		},
		{
			name:     "opacity with regular decimal",
			input:    "opacity: 0.5;",
			expected: "opacity:.5;",
		},
		{
			name:     "opacity with very small decimal",
			input:    "opacity: 0.001;",
			expected: "opacity:.001;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := optimizer.Optimize([]byte(tt.input))
			if err != nil {
				t.Fatalf("Optimize returned error: %v", err)
			}

			result := string(output)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
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

// Test JS optimizer preserves UTF-8 characters
func TestJSOptimizerUTF8Preservation(t *testing.T) {
	config := JSConfig{
		Enabled:        true,
		MinifyJS:       true,
		RemoveComments: true,
	}

	optimizer := NewJSOptimizer(config)

	// Test with em dash (U+2014, UTF-8: E2 80 94)
	input := []byte("var x = '—';")
	output, err := optimizer.Optimize(input)
	if err != nil {
		t.Fatalf("Optimize returned error: %v", err)
	}

	// The em dash should be preserved as-is (3 bytes: E2 80 94)
	result := string(output)
	if !strings.Contains(result, "—") {
		t.Errorf("Em dash was not preserved. Got: %q (hex: %x)", result, output)
	}

	// Verify the actual bytes are correct
	emDashBytes := []byte{0xE2, 0x80, 0x94}
	if !bytes.Contains(output, emDashBytes) {
		t.Errorf("Em dash bytes not found. Expected E2 80 94, got: %x", output)
	}
}

// Test JS optimizer preserves template literal whitespace
func TestJSOptimizerTemplateLiteralPreservation(t *testing.T) {
	config := JSConfig{
		Enabled:        true,
		MinifyJS:       true,
		RemoveComments: true,
	}

	optimizer := NewJSOptimizer(config)

	// GraphQL query in template literal - whitespace must be preserved
	input := []byte("const query = `\n    query {\n        objects(filter: { type: \"test\" })\n    }\n`;")
	output, err := optimizer.Optimize(input)
	if err != nil {
		t.Fatalf("Optimize returned error: %v", err)
	}

	result := string(output)
	// The newlines and spaces inside the template literal should be preserved
	if !strings.Contains(result, "query {") {
		t.Errorf("Template literal whitespace not preserved. Got: %q", result)
	}
	if !strings.Contains(result, "filter: { type:") {
		t.Errorf("Template literal content corrupted. Got: %q", result)
	}
}

// Test the main engine with a simple directory
func TestEngineOptimizeDirectorySimple(t *testing.T) {
	// Create test directory
	testDir := t.TempDir()

	// Create test files
	htmlFile := filepath.Join(testDir, "index.html")
	if err := os.WriteFile(htmlFile, []byte(`<html><body>Test</body></html>`), 0644); err != nil {
		t.Fatal(err)
	}

	cssFile := filepath.Join(testDir, "style.css")
	if err := os.WriteFile(cssFile, []byte(`body { margin: 0; }`), 0644); err != nil {
		t.Fatal(err)
	}

	jsFile := filepath.Join(testDir, "script.js")
	if err := os.WriteFile(jsFile, []byte(`function test() { return true; }`), 0644); err != nil {
		t.Fatal(err)
	}

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
	if err := os.WriteFile(htmlFile, []byte(`<html><body>Test</body></html>`), 0644); err != nil {
		t.Fatal(err)
	}

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
	if err := os.WriteFile(minFile, []byte(`function test(){}`), 0644); err != nil {
		t.Fatal(err)
	}

	regularFile := filepath.Join(testDir, "script.js")
	if err := os.WriteFile(regularFile, []byte(`function test() { return true; }`), 0644); err != nil {
		t.Fatal(err)
	}

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

// Test CSS optimizer preserves color names inside strings
func TestCSSOptimizerPreservesStrings(t *testing.T) {
	config := CSSConfig{
		Enabled:        true,
		MinifyCSS:      true,
		CompressColors: true,
	}

	optimizer := NewCSSOptimizer(config)

	tests := []struct {
		name     string
		input    string
		contains string // String that should be preserved
	}{
		{
			name:     "color name in content property should be preserved",
			input:    `div { content: "white background"; }`,
			contains: `"white background"`,
		},
		{
			name:     "color name in single quotes should be preserved",
			input:    `div { content: 'black text'; }`,
			contains: `'black text'`,
		},
		{
			name:     "color name outside string is preserved (no named color replacement)",
			input:    `div { color: white; }`,
			contains: `white`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := optimizer.Optimize([]byte(tt.input))
			if err != nil {
				t.Fatalf("Optimize returned error: %v", err)
			}

			result := string(output)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// Test HTML optimizer handles comments with special characters
func TestHTMLOptimizerCommentsWithSpecialChars(t *testing.T) {
	config := HTMLConfig{
		Enabled:        true,
		RemoveComments: true,
	}

	optimizer := NewHTMLOptimizer(config)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "comment with > character",
			input:    `<div><!-- comment with > symbol --></div>`,
			expected: `<div></div>`,
		},
		{
			name:     "comment with < character",
			input:    `<div><!-- comment with < symbol --></div>`,
			expected: `<div></div>`,
		},
		{
			name:     "comment with arrows",
			input:    `<div><!-- -> and <- arrows --></div>`,
			expected: `<div></div>`,
		},
		{
			name:     "multiline comment",
			input:    "<div><!--\nmultiline\ncomment\n--></div>",
			expected: `<div></div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := optimizer.Optimize([]byte(tt.input))
			if err != nil {
				t.Fatalf("Optimize returned error: %v", err)
			}

			result := string(output)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test HTML minify preserves pre/code/script/style content
func TestHTMLMinifyPreservesSpecialTags(t *testing.T) {
	config := HTMLConfig{
		Enabled:          true,
		MinifyHTML:       true,
		RemoveWhitespace: true,
	}

	optimizer := NewHTMLOptimizer(config)

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "pre tag whitespace preserved",
			input:    `<div><pre>  code  with  spaces  </pre></div>`,
			contains: `  code  with  spaces  `,
		},
		{
			name:     "code tag whitespace preserved",
			input:    `<div><code>  code  </code></div>`,
			contains: `  code  `,
		},
		{
			name:     "script content preserved",
			input:    `<div><script>const x = 1;  const y = 2;</script></div>`,
			contains: `const x = 1;  const y = 2;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := optimizer.Optimize([]byte(tt.input))
			if err != nil {
				t.Fatalf("Optimize returned error: %v", err)
			}

			result := string(output)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// Test JS obfuscation is disabled (doesn't break code)
func TestJSObfuscationDisabled(t *testing.T) {
	config := JSConfig{
		Enabled:   true,
		MinifyJS:  true,
		Obfuscate: true, // Even when enabled, it should not break code
	}

	optimizer := NewJSOptimizer(config)

	input := []byte(`function test() { document.getElementById('test'); window.alert('hello'); }`)
	output, err := optimizer.Optimize(input)
	if err != nil {
		t.Fatalf("Optimize returned error: %v", err)
	}

	result := string(output)

	// These keywords should NOT be replaced
	if !strings.Contains(result, "function") {
		t.Error("function keyword was incorrectly replaced")
	}
	if !strings.Contains(result, "document") {
		t.Error("document was incorrectly replaced")
	}
	if !strings.Contains(result, "window") {
		t.Error("window was incorrectly replaced")
	}
}
