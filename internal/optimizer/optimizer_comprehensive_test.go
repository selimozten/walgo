package optimizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCSSOptimizer_FindUnusedRules(t *testing.T) {
	tests := []struct {
		name     string
		css      string
		html     string
		wantLen  int
		contains []string
	}{
		{
			name: "Find unused CSS rules",
			css: `.used { color: red; }
.unused { color: blue; }
#used-id { color: green; }
#unused-id { color: yellow; }`,
			html: `<div class="used" id="used-id">Test</div>`,
			wantLen:  2, // Should find 2 unused rules
			contains: []string{".unused", "#unused-id"},
		},
		{
			name:     "No unused rules",
			css:      `.test { color: red; }`,
			html:     `<div class="test">Test</div>`,
			wantLen:  0,
			contains: []string{},
		},
		{
			name: "All rules unused",
			css: `.unused1 { color: red; }
.unused2 { color: blue; }`,
			html:     `<div>No classes</div>`,
			wantLen:  2,
			contains: []string{".unused1", ".unused2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := NewCSSOptimizer(CSSConfig{
				RemoveUnused: true,
			})

			// Test findUnusedRules
			result := opt.findUnusedRules([]byte(tt.css), []byte(tt.html))

			if len(result) != tt.wantLen {
				t.Errorf("findUnusedRules() returned %d rules, want %d", len(result), tt.wantLen)
			}

			for _, expected := range tt.contains {
				found := false
				for _, rule := range result {
					if strings.Contains(rule, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find %s in unused rules", expected)
				}
			}
		})
	}
}

func TestCSSOptimizer_IsSelectorUsed(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		html     string
		want     bool
	}{
		{
			name:     "Class selector used",
			selector: ".test-class",
			html:     `<div class="test-class">Test</div>`,
			want:     true,
		},
		{
			name:     "Class selector not used",
			selector: ".unused-class",
			html:     `<div class="other-class">Test</div>`,
			want:     false,
		},
		{
			name:     "ID selector used",
			selector: "#test-id",
			html:     `<div id="test-id">Test</div>`,
			want:     true,
		},
		{
			name:     "Element selector used",
			selector: "div",
			html:     `<div>Test</div>`,
			want:     true,
		},
		{
			name:     "Complex selector with pseudo-class",
			selector: ".btn:hover",
			html:     `<button class="btn">Click me</button>`,
			want:     true,
		},
		{
			name:     "Attribute selector",
			selector: "[data-test]",
			html:     `<div data-test="value">Test</div>`,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := NewCSSOptimizer(CSSConfig{})
			got := opt.isSelectorUsed(tt.selector, tt.html)

			if got != tt.want {
				t.Errorf("isSelectorUsed(%s) = %v, want %v", tt.selector, got, tt.want)
			}
		})
	}
}

func TestCSSOptimizer_RemoveUnusedRules(t *testing.T) {
	tests := []struct {
		name        string
		css         string
		html        string
		shouldKeep  []string
		shouldRemove []string
	}{
		{
			name: "Remove unused rules",
			css: `.used { color: red; }
.unused { color: blue; }
#header { background: white; }
.footer { padding: 10px; }`,
			html: `<!DOCTYPE html>
<html>
<body>
	<div id="header">Header</div>
	<div class="used">Content</div>
</body>
</html>`,
			shouldKeep:   []string{".used", "#header"},
			shouldRemove: []string{".unused", ".footer"},
		},
		{
			name: "Keep all when RemoveUnused is false",
			css: `.test1 { color: red; }
.test2 { color: blue; }`,
			html:        `<div>No classes</div>`,
			shouldKeep:  []string{".test1", ".test2"},
			shouldRemove: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := NewCSSOptimizer(CSSConfig{
				RemoveUnused: tt.name != "Keep all when RemoveUnused is false",
			})

			result := opt.RemoveUnusedRules([]byte(tt.css), []byte(tt.html))
			resultStr := string(result)

			// Check that expected rules are kept
			for _, rule := range tt.shouldKeep {
				if !strings.Contains(resultStr, rule) {
					t.Errorf("RemoveUnusedRules() removed %s which should be kept", rule)
				}
			}

			// Check that expected rules are removed
			for _, rule := range tt.shouldRemove {
				if strings.Contains(resultStr, rule) {
					t.Errorf("RemoveUnusedRules() didn't remove %s", rule)
				}
			}
		})
	}
}

func TestCSSOptimizer_GetFileExtensions(t *testing.T) {
	opt := NewCSSOptimizer(CSSConfig{})
	exts := opt.GetFileExtensions()

	expected := []string{".css"}
	if len(exts) != len(expected) {
		t.Errorf("GetFileExtensions() returned %d extensions, want %d", len(exts), len(expected))
	}

	for i, ext := range exts {
		if ext != expected[i] {
			t.Errorf("GetFileExtensions()[%d] = %s, want %s", i, ext, expected[i])
		}
	}
}

func TestHTMLOptimizer_CompressInlineCSS(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "Compress style tags",
			html: `<html>
<head>
<style>
  body {
    margin: 0;
    padding: 0;
  }
  .test {
    color: red;
  }
</style>
</head>
<body></body>
</html>`,
			want: `<html>
<head>
<style>body{margin:0;padding:0}.test{color:red}</style>
</head>
<body></body>
</html>`,
		},
		{
			name: "Compress inline styles",
			html: `<div style="  color:  red;  margin:  10px  ">Test</div>`,
			want: `<div style="color:red;margin:10px">Test</div>`,
		},
		{
			name: "Handle multiple style tags",
			html: `<style>
  .a { color: red; }
</style>
<style>
  .b { color: blue; }
</style>`,
			want: `<style>.a{color:red}</style>
<style>.b{color:blue}</style>`,
		},
		{
			name: "Empty style tag",
			html: `<style></style>`,
			want: `<style></style>`,
		},
		{
			name: "No styles",
			html: `<div>No styles here</div>`,
			want: `<div>No styles here</div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := NewHTMLOptimizer(HTMLConfig{
				CompressInlineCSS: true,
			})
			got := opt.compressInlineCSS([]byte(tt.html))

			// Normalize whitespace for comparison
			gotNorm := strings.TrimSpace(string(got))
			wantNorm := strings.TrimSpace(tt.want)

			if gotNorm != wantNorm {
				t.Errorf("compressInlineCSS() = %v, want %v", gotNorm, wantNorm)
			}
		})
	}
}

func TestHTMLOptimizer_CompressInlineJS(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "Compress script tags",
			html: `<html>
<head>
<script>
  function test() {
    console.log("hello");
    return true;
  }
</script>
</head>
<body></body>
</html>`,
			want: `<html>
<head>
<script>function test(){console.log("hello");return true;}</script>
</head>
<body></body>
</html>`,
		},
		{
			name: "Compress onclick attributes",
			html: `<button onclick="  alert('test');  return false;  ">Click</button>`,
			want: `<button onclick="alert('test');return false;">Click</button>`,
		},
		{
			name: "Handle multiple event handlers",
			html: `<div onclick="test()" onmouseover="  hover()  " onmouseout="leave()">Test</div>`,
			want: `<div onclick="test()" onmouseover="hover()" onmouseout="leave()">Test</div>`,
		},
		{
			name: "Empty script tag",
			html: `<script></script>`,
			want: `<script></script>`,
		},
		{
			name: "Script with src attribute",
			html: `<script src="app.js"></script>`,
			want: `<script src="app.js"></script>`,
		},
		{
			name: "No scripts",
			html: `<div>No scripts here</div>`,
			want: `<div>No scripts here</div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := NewHTMLOptimizer(HTMLConfig{
				CompressInlineJS: true,
			})
			got := opt.compressInlineJS([]byte(tt.html))

			// Normalize whitespace for comparison
			gotNorm := strings.TrimSpace(string(got))
			wantNorm := strings.TrimSpace(tt.want)

			if gotNorm != wantNorm {
				t.Errorf("compressInlineJS() = %v, want %v", gotNorm, wantNorm)
			}
		})
	}
}

func TestHTMLOptimizer_GetFileExtensions(t *testing.T) {
	opt := NewHTMLOptimizer(HTMLConfig{})
	exts := opt.GetFileExtensions()

	expected := []string{".html", ".htm"}
	if len(exts) != len(expected) {
		t.Errorf("GetFileExtensions() returned %d extensions, want %d", len(exts), len(expected))
	}

	for i, ext := range exts {
		if ext != expected[i] {
			t.Errorf("GetFileExtensions()[%d] = %s, want %s", i, ext, expected[i])
		}
	}
}

func TestJSOptimizer_GetFileExtensions(t *testing.T) {
	opt := NewJSOptimizer(JSConfig{})
	exts := opt.GetFileExtensions()

	expected := []string{".js", ".mjs"}
	if len(exts) != len(expected) {
		t.Errorf("GetFileExtensions() returned %d extensions, want %d", len(exts), len(expected))
	}

	for i, ext := range exts {
		if ext != expected[i] {
			t.Errorf("GetFileExtensions()[%d] = %s, want %s", i, ext, expected[i])
		}
	}
}

func TestEngine_OptimizeWithAllOptimizersEnabled(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<style>
		body {
			margin: 0;
			padding: 0;
		}
	</style>
	<script>
		function init() {
			console.log("Initialized");
		}
	</script>
</head>
<body onload="init()">
	<h1 style="  color:  blue;  ">Hello World</h1>
</body>
</html>`

	cssContent := `/* Main styles */
body {
	font-family: Arial, sans-serif;
	background-color: #ffffff;
}

/* Unused class */
.unused {
	display: none;
}`

	jsContent := `// Initialize app
function initApp() {
	// Log message
	console.log("App started");

	// Setup event listeners
	document.addEventListener("click", function(e) {
		console.log("Clicked");
	});
}

// Call init
initApp();`

	// Write test files
	htmlFile := filepath.Join(tmpDir, "index.html")
	cssFile := filepath.Join(tmpDir, "style.css")
	jsFile := filepath.Join(tmpDir, "app.js")

	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cssFile, []byte(cssContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(jsFile, []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create engine with all optimizations enabled
	config := OptimizerConfig{
		Enabled: true,
		HTML: HTMLConfig{
			Enabled:           true,
			MinifyHTML:        true,
			RemoveComments:    true,
			RemoveWhitespace:  true,
			CompressInlineCSS: true,
			CompressInlineJS:  true,
		},
		CSS: CSSConfig{
			Enabled:        true,
			MinifyCSS:      true,
			RemoveComments: true,
			RemoveUnused:   true,
			CompressColors: true,
		},
		JS: JSConfig{
			Enabled:        true,
			MinifyJS:       true,
			RemoveComments: true,
		},
		Verbose: true,
	}

	engine := NewEngine(config)
	stats, err := engine.OptimizeDirectory(tmpDir)

	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	// Verify stats
	if stats.FilesProcessed == 0 {
		t.Error("No files were processed")
	}
	if stats.FilesOptimized == 0 {
		t.Error("No files were optimized")
	}
	if stats.OriginalSize == 0 {
		t.Error("Original size is 0")
	}
	if stats.OptimizedSize >= stats.OriginalSize {
		t.Error("Optimized size should be less than original")
	}

	// Check that files were actually modified
	htmlOptimized, _ := os.ReadFile(htmlFile)
	if len(htmlOptimized) >= len(htmlContent) {
		t.Error("HTML file was not optimized")
	}

	cssOptimized, _ := os.ReadFile(cssFile)
	if len(cssOptimized) >= len(cssContent) {
		t.Error("CSS file was not optimized")
	}

	jsOptimized, _ := os.ReadFile(jsFile)
	if len(jsOptimized) >= len(jsContent) {
		t.Error("JS file was not optimized")
	}
}

func TestEngine_OptimizeWithSkipPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "optimize.css")
	file2 := filepath.Join(tmpDir, "skip.min.css")

	cssContent := `body { margin: 0; padding: 0; }`

	if err := os.WriteFile(file1, []byte(cssContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(cssContent), 0644); err != nil {
		t.Fatal(err)
	}

	config := OptimizerConfig{
		Enabled: true,
		CSS: CSSConfig{
			Enabled:   true,
			MinifyCSS: true,
		},
		SkipPatterns: []string{"*.min.css"},
	}

	engine := NewEngine(config)
	stats, err := engine.OptimizeDirectory(tmpDir)

	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	// Check that skip.min.css was skipped
	if stats.FilesSkipped == 0 {
		t.Error("Expected files to be skipped")
	}

	// Check that optimize.css was processed
	if stats.FilesOptimized == 0 {
		t.Error("Expected files to be optimized")
	}
}