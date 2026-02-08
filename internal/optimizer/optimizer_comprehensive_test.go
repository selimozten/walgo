package optimizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTMLOptimizer_CompressInlineCSS(t *testing.T) {
	t.Skip("HTML inline CSS compression not yet implemented")
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
	t.Skip("HTML inline JavaScript compression not yet implemented")
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
	htmlOptimized, err := os.ReadFile(htmlFile)
	if err != nil {
		t.Fatalf("Failed to read optimized HTML file: %v", err)
	}
	if len(htmlOptimized) >= len(htmlContent) {
		t.Error("HTML file was not optimized")
	}

	cssOptimized, err := os.ReadFile(cssFile)
	if err != nil {
		t.Fatalf("Failed to read optimized CSS file: %v", err)
	}
	if len(cssOptimized) >= len(cssContent) {
		t.Error("CSS file was not optimized")
	}

	jsOptimized, err := os.ReadFile(jsFile)
	if err != nil {
		t.Fatalf("Failed to read optimized JS file: %v", err)
	}
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
