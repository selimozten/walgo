package hugo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/selimozten/walgo/internal/ai"
)

// =============================================================================
// Helper functions
// =============================================================================

// mkdirAll creates nested directories under a base path.
func mkdirAll(t *testing.T, parts ...string) {
	t.Helper()
	dir := filepath.Join(parts...)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
}

// writeFile writes content to a file, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Failed to create parent dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// =============================================================================
// dirExists Tests
// =============================================================================

func TestDirExists(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) string
		want  bool
	}{
		{
			name: "Existing directory returns true",
			setup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "existing")
				mkdirAll(t, dir)
				return dir
			},
			want: true,
		},
		{
			name: "Non-existent path returns false",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			want: false,
		},
		{
			name: "File (not directory) returns false",
			setup: func(t *testing.T) string {
				f := filepath.Join(t.TempDir(), "afile.txt")
				writeFile(t, f, "hello")
				return f
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			if got := dirExists(path); got != tt.want {
				t.Errorf("dirExists(%q) = %v, want %v", path, got, tt.want)
			}
		})
	}
}

// =============================================================================
// extractThemeFromConfig Tests
// =============================================================================

func TestExtractThemeFromConfig(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   string
	}{
		{
			name:   "Standard theme line with double quotes",
			config: `theme = "ananke"`,
			want:   "ananke",
		},
		{
			name:   "Standard theme line with single quotes",
			config: `theme = 'hugo-book'`,
			want:   "hugo-book",
		},
		{
			name: "Theme line among other config",
			config: `baseURL = "https://example.com/"
title = "My Site"
theme = "PaperMod"
languageCode = "en-us"`,
			want: "PaperMod",
		},
		{
			name:   "Theme with spaces around equals",
			config: `theme  =  "blowfish"`,
			want:   "blowfish",
		},
		{
			name:   "No theme line",
			config: `baseURL = "https://example.com/"`,
			want:   "",
		},
		{
			name:   "Empty config",
			config: "",
			want:   "",
		},
		{
			name:   "Theme without quotes (bare value)",
			config: `theme = myTheme`,
			want:   "myTheme",
		},
		{
			name:   "Commented out theme line is skipped",
			config: `# theme = "oldtheme"`,
			want:   "",
		},
		{
			name: "themeDir is not matched — only exact theme key",
			config: `themeDir = "mythemes"
theme = "actual-theme"`,
			want: "actual-theme",
		},
		{
			name: "theme line before themeDir returns correct theme",
			config: `theme = "actual-theme"
themeDir = "mythemes"`,
			want: "actual-theme",
		},
		{
			name:   "themeColor is not matched",
			config: `themeColor = "#ff0000"`,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractThemeFromConfig(tt.config)
			if got != tt.want {
				t.Errorf("extractThemeFromConfig() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// parseConfigForTheme Tests
// =============================================================================

func TestParseConfigForTheme(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		setupSite func(t *testing.T) string // returns sitePath
		want      ai.SiteType
	}{
		{
			name:   "walgo-biolink theme detected",
			config: `theme = "walgo-biolink"`,
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeBiolink,
		},
		{
			name:   "walgo-whitepaper theme detected (case insensitive)",
			config: `theme = "Walgo-Whitepaper"`,
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeWhitepaper,
		},
		{
			name:   "hugo-book theme detected as docs",
			config: `theme = "hugo-book"`,
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeDocs,
		},
		{
			name:   "docsy theme detected as docs",
			config: `theme = "docsy"`,
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeDocs,
		},
		{
			name:   "learn theme detected as docs",
			config: `theme = "learn"`,
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeDocs,
		},
		{
			name:   "doks theme detected as docs",
			config: `theme = "doks"`,
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeDocs,
		},
		{
			name:   "geekdoc theme detected as docs",
			config: `theme = "geekdoc"`,
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeDocs,
		},
		{
			name:   "techdoc theme detected as docs",
			config: `theme = "techdoc"`,
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeDocs,
		},
		{
			name:   "Generic theme falls back to content detection - docs dir",
			config: `theme = "ananke"`,
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content", "docs")
				return dir
			},
			want: ai.SiteTypeDocs,
		},
		{
			name:   "Generic theme falls back to content detection - whitepaper dir",
			config: `theme = "ananke"`,
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content", "whitepaper")
				return dir
			},
			want: ai.SiteTypeWhitepaper,
		},
		{
			name:   "Generic theme falls back to blog when no special dirs",
			config: `theme = "ananke"`,
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content", "posts")
				return dir
			},
			want: ai.SiteTypeBlog,
		},
		{
			name:   "No theme in config falls back to blog",
			config: `baseURL = "/"`,
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeBlog,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath := tt.setupSite(t)
			got := parseConfigForTheme(tt.config, sitePath)
			if got != tt.want {
				t.Errorf("parseConfigForTheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// detectSiteTypeFromContent Tests
// =============================================================================

func TestDetectSiteTypeFromContent(t *testing.T) {
	tests := []struct {
		name      string
		setupSite func(t *testing.T) string
		want      ai.SiteType
	}{
		{
			name: "Whitepaper content directory",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content", "whitepaper")
				return dir
			},
			want: ai.SiteTypeWhitepaper,
		},
		{
			name: "Docs content directory",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content", "docs")
				return dir
			},
			want: ai.SiteTypeDocs,
		},
		{
			name: "Both whitepaper and docs prefers whitepaper",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content", "whitepaper")
				mkdirAll(t, dir, "content", "docs")
				return dir
			},
			want: ai.SiteTypeWhitepaper,
		},
		{
			name: "Posts only defaults to blog",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content", "posts")
				return dir
			},
			want: ai.SiteTypeBlog,
		},
		{
			name: "No content directory defaults to blog",
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeBlog,
		},
		{
			name: "Empty content directory defaults to blog",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content")
				return dir
			},
			want: ai.SiteTypeBlog,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath := tt.setupSite(t)
			got := detectSiteTypeFromContent(sitePath)
			if got != tt.want {
				t.Errorf("detectSiteTypeFromContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// detectFromConfig Tests
// =============================================================================

func TestDetectFromConfig(t *testing.T) {
	tests := []struct {
		name      string
		setupSite func(t *testing.T) string
		want      ai.SiteType
	}{
		{
			name: "hugo.toml with biolink theme",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "walgo-biolink"`)
				return dir
			},
			want: ai.SiteTypeBiolink,
		},
		{
			name: "config.toml fallback with whitepaper theme",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				// No hugo.toml, only config.toml
				writeFile(t, filepath.Join(dir, "config.toml"), `theme = "walgo-whitepaper"`)
				return dir
			},
			want: ai.SiteTypeWhitepaper,
		},
		{
			name: "hugo.toml takes precedence over config.toml",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "walgo-biolink"`)
				writeFile(t, filepath.Join(dir, "config.toml"), `theme = "walgo-whitepaper"`)
				return dir
			},
			want: ai.SiteTypeBiolink,
		},
		{
			name: "No config files defaults to blog",
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeBlog,
		},
		{
			name: "Config with generic theme defaults to blog",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "ananke"`)
				return dir
			},
			want: ai.SiteTypeBlog,
		},
		{
			name: "Config with docs theme",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "hugo-book"`)
				return dir
			},
			want: ai.SiteTypeDocs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath := tt.setupSite(t)
			got := detectFromConfig(sitePath)
			if got != tt.want {
				t.Errorf("detectFromConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// DetectSiteType Tests (exported function - integration of detectFromConfig + detectFromContentStructure)
// =============================================================================

func TestDetectSiteType(t *testing.T) {
	tests := []struct {
		name      string
		setupSite func(t *testing.T) string
		want      ai.SiteType
	}{
		{
			name: "Biolink detected from config theme",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "walgo-biolink"`)
				mkdirAll(t, dir, "content")
				return dir
			},
			want: ai.SiteTypeBiolink,
		},
		{
			name: "Whitepaper detected from config theme",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "walgo-whitepaper"`)
				mkdirAll(t, dir, "content")
				return dir
			},
			want: ai.SiteTypeWhitepaper,
		},
		{
			name: "Docs detected from docs theme",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "docsy"`)
				mkdirAll(t, dir, "content")
				return dir
			},
			want: ai.SiteTypeDocs,
		},
		{
			name: "Default blog with no indicators",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "ananke"`)
				mkdirAll(t, dir, "content", "posts")
				return dir
			},
			want: ai.SiteTypeBlog,
		},
		{
			name: "No config at all defaults to blog",
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: ai.SiteTypeBlog,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath := tt.setupSite(t)
			got := DetectSiteType(sitePath)
			if got != tt.want {
				t.Errorf("DetectSiteType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// GetThemeName Tests
// =============================================================================

func TestGetThemeName(t *testing.T) {
	tests := []struct {
		name      string
		setupSite func(t *testing.T) string
		want      string
	}{
		{
			name: "Theme from hugo.toml",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "ananke"`)
				return dir
			},
			want: "ananke",
		},
		{
			name: "Theme from config.toml when hugo.toml absent",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "config.toml"), `theme = "PaperMod"`)
				return dir
			},
			want: "PaperMod",
		},
		{
			name: "hugo.toml takes precedence over config.toml",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "ananke"`)
				writeFile(t, filepath.Join(dir, "config.toml"), `theme = "PaperMod"`)
				return dir
			},
			want: "ananke",
		},
		{
			name: "No config files returns empty string",
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: "",
		},
		{
			name: "Config without theme line returns empty string",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `baseURL = "/"`)
				return dir
			},
			want: "",
		},
		{
			name: "Theme with quotes in config",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "hugo.toml"), `theme = "hugo-book"`)
				return dir
			},
			want: "hugo-book",
		},
		{
			name: "Theme line in hugo.toml with extra config",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				config := `baseURL = "https://example.com/"
title = "My Docs"
theme = "docsy"
languageCode = "en-us"
`
				writeFile(t, filepath.Join(dir, "hugo.toml"), config)
				return dir
			},
			want: "docsy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath := tt.setupSite(t)
			got := GetThemeName(sitePath)
			if got != tt.want {
				t.Errorf("GetThemeName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// DetectContentTypes Tests (from hugo.go, but related to site detection)
// =============================================================================

func TestDetectContentTypes(t *testing.T) {
	tests := []struct {
		name      string
		setupSite func(t *testing.T) string
		wantCount int
		wantFirst string // Name of the first (most files) content type
		wantErr   bool
	}{
		{
			name: "Multiple content types sorted by file count",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				// Posts has 3 files, docs has 1
				writeFile(t, filepath.Join(dir, "content", "posts", "p1.md"), "---\ntitle: P1\n---\n")
				writeFile(t, filepath.Join(dir, "content", "posts", "p2.md"), "---\ntitle: P2\n---\n")
				writeFile(t, filepath.Join(dir, "content", "posts", "p3.md"), "---\ntitle: P3\n---\n")
				writeFile(t, filepath.Join(dir, "content", "docs", "d1.md"), "---\ntitle: D1\n---\n")
				return dir
			},
			wantCount: 2,
			wantFirst: "posts",
		},
		{
			name: "Single content type",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "content", "pages", "about.md"), "# About\n")
				return dir
			},
			wantCount: 1,
			wantFirst: "pages",
		},
		{
			name: "No content directory returns nil",
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			wantCount: 0,
		},
		{
			name: "Empty content directory",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content")
				return dir
			},
			wantCount: 0,
		},
		{
			name: "Hidden directories are skipped",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "content", ".hidden", "file.md"), "hidden\n")
				writeFile(t, filepath.Join(dir, "content", "visible", "file.md"), "visible\n")
				return dir
			},
			wantCount: 1,
			wantFirst: "visible",
		},
		{
			name: "Files at content root are not counted as types",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "content", "_index.md"), "# Home\n")
				writeFile(t, filepath.Join(dir, "content", "about.md"), "# About\n")
				writeFile(t, filepath.Join(dir, "content", "posts", "p1.md"), "# Post\n")
				return dir
			},
			wantCount: 1,
			wantFirst: "posts",
		},
		{
			name: "Nested markdown files are counted",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "content", "docs", "chapter1", "intro.md"), "intro\n")
				writeFile(t, filepath.Join(dir, "content", "docs", "chapter2", "advanced.md"), "advanced\n")
				writeFile(t, filepath.Join(dir, "content", "docs", "_index.md"), "docs index\n")
				return dir
			},
			wantCount: 1,
			wantFirst: "docs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath := tt.setupSite(t)
			types, err := DetectContentTypes(sitePath)

			if (err != nil) != tt.wantErr {
				t.Fatalf("DetectContentTypes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(types) != tt.wantCount {
				t.Errorf("DetectContentTypes() returned %d types, want %d", len(types), tt.wantCount)
				for _, ct := range types {
					t.Logf("  type: %s (%d files)", ct.Name, ct.FileCount)
				}
				return
			}

			if tt.wantFirst != "" && len(types) > 0 {
				if types[0].Name != tt.wantFirst {
					t.Errorf("First content type = %q, want %q", types[0].Name, tt.wantFirst)
				}
			}
		})
	}
}

func TestDetectContentTypes_FileCountAccuracy(t *testing.T) {
	dir := t.TempDir()

	// Create exactly 5 markdown files in posts
	for i := 1; i <= 5; i++ {
		writeFile(t, filepath.Join(dir, "content", "posts", "post"+string(rune('0'+i))+".md"), "# Post\n")
	}
	// Create 2 markdown files and 1 non-md file in docs
	writeFile(t, filepath.Join(dir, "content", "docs", "a.md"), "doc a\n")
	writeFile(t, filepath.Join(dir, "content", "docs", "b.md"), "doc b\n")
	writeFile(t, filepath.Join(dir, "content", "docs", "readme.txt"), "not counted\n")

	types, err := DetectContentTypes(dir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(types) != 2 {
		t.Fatalf("Expected 2 content types, got %d", len(types))
	}

	// First should be posts (5 files), second should be docs (2 files)
	if types[0].Name != "posts" {
		t.Errorf("Expected first type to be 'posts', got %q", types[0].Name)
	}
	if types[0].FileCount != 5 {
		t.Errorf("Expected posts to have 5 files, got %d", types[0].FileCount)
	}
	if types[1].Name != "docs" {
		t.Errorf("Expected second type to be 'docs', got %q", types[1].Name)
	}
	if types[1].FileCount != 2 {
		t.Errorf("Expected docs to have 2 files, got %d", types[1].FileCount)
	}
}

// =============================================================================
// GetDefaultContentType Tests
// =============================================================================

func TestGetDefaultContentType(t *testing.T) {
	tests := []struct {
		name      string
		setupSite func(t *testing.T) string
		want      string
	}{
		{
			name: "Returns most used content type",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				// docs has more files than posts
				writeFile(t, filepath.Join(dir, "content", "docs", "a.md"), "a\n")
				writeFile(t, filepath.Join(dir, "content", "docs", "b.md"), "b\n")
				writeFile(t, filepath.Join(dir, "content", "docs", "c.md"), "c\n")
				writeFile(t, filepath.Join(dir, "content", "posts", "p1.md"), "p\n")
				return dir
			},
			want: "docs",
		},
		{
			name: "Returns posts when no content types found",
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			want: "posts",
		},
		{
			name: "Returns posts for empty content dir",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "content")
				return dir
			},
			want: "posts",
		},
		{
			name: "Returns the single content type available",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				writeFile(t, filepath.Join(dir, "content", "tutorials", "tut1.md"), "tutorial\n")
				return dir
			},
			want: "tutorials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath := tt.setupSite(t)
			got := GetDefaultContentType(sitePath)
			if got != tt.want {
				t.Errorf("GetDefaultContentType() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// countMarkdownFiles Tests
// =============================================================================

func TestCountMarkdownFiles(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) string
		want  int
	}{
		{
			name: "Counts only .md files",
			setup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "test")
				writeFile(t, filepath.Join(dir, "a.md"), "md\n")
				writeFile(t, filepath.Join(dir, "b.md"), "md\n")
				writeFile(t, filepath.Join(dir, "c.txt"), "txt\n")
				writeFile(t, filepath.Join(dir, "d.html"), "html\n")
				return dir
			},
			want: 2,
		},
		{
			name: "Counts nested .md files",
			setup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "test")
				writeFile(t, filepath.Join(dir, "root.md"), "r\n")
				writeFile(t, filepath.Join(dir, "sub", "nested.md"), "n\n")
				writeFile(t, filepath.Join(dir, "sub", "deep", "deep.md"), "d\n")
				return dir
			},
			want: 3,
		},
		{
			name: "Empty directory returns 0",
			setup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "empty")
				mkdirAll(t, dir)
				return dir
			},
			want: 0,
		},
		{
			name: "No .md files returns 0",
			setup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "nomd")
				writeFile(t, filepath.Join(dir, "a.txt"), "txt\n")
				writeFile(t, filepath.Join(dir, "b.html"), "html\n")
				return dir
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			got := countMarkdownFiles(dir)
			if got != tt.want {
				t.Errorf("countMarkdownFiles() = %d, want %d", got, tt.want)
			}
		})
	}
}

// =============================================================================
// GetThemeInfo Tests (from hugo.go, but closely related to site detection)
// =============================================================================

func TestGetThemeInfo(t *testing.T) {
	tests := []struct {
		name        string
		siteType    SiteType
		wantName    string
		wantDirName string
	}{
		{
			name:        "Blog type returns Ananke",
			siteType:    SiteTypeBlog,
			wantName:    "Ananke",
			wantDirName: "ananke",
		},
		{
			name:        "Docs type returns Book",
			siteType:    SiteTypeDocs,
			wantName:    "Book",
			wantDirName: "hugo-book",
		},
		{
			name:        "Biolink type returns Walgo Biolink",
			siteType:    SiteTypeBiolink,
			wantName:    "Walgo Biolink",
			wantDirName: "walgo-biolink",
		},
		{
			name:        "Whitepaper type returns Walgo Whitepaper",
			siteType:    SiteTypeWhitepaper,
			wantName:    "Walgo Whitepaper",
			wantDirName: "walgo-whitepaper",
		},
		{
			name:        "Unknown type defaults to Ananke",
			siteType:    "unknown",
			wantName:    "Ananke",
			wantDirName: "ananke",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := GetThemeInfo(tt.siteType)
			if info.Name != tt.wantName {
				t.Errorf("GetThemeInfo(%q).Name = %q, want %q", tt.siteType, info.Name, tt.wantName)
			}
			if info.DirName != tt.wantDirName {
				t.Errorf("GetThemeInfo(%q).DirName = %q, want %q", tt.siteType, info.DirName, tt.wantDirName)
			}
		})
	}
}

// =============================================================================
// GetInstalledThemes Tests
// =============================================================================

func TestGetInstalledThemes(t *testing.T) {
	tests := []struct {
		name      string
		setupSite func(t *testing.T) string
		wantCount int
		wantErr   bool
	}{
		{
			name: "Multiple installed themes",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "themes", "ananke")
				mkdirAll(t, dir, "themes", "hugo-book")
				return dir
			},
			wantCount: 2,
		},
		{
			name: "No themes directory returns nil",
			setupSite: func(t *testing.T) string {
				return t.TempDir()
			},
			wantCount: 0,
		},
		{
			name: "Empty themes directory",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "themes")
				return dir
			},
			wantCount: 0,
		},
		{
			name: "Files in themes dir are not counted",
			setupSite: func(t *testing.T) string {
				dir := t.TempDir()
				mkdirAll(t, dir, "themes", "real-theme")
				writeFile(t, filepath.Join(dir, "themes", "not-a-theme.txt"), "file\n")
				return dir
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath := tt.setupSite(t)
			themes, err := GetInstalledThemes(sitePath)

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetInstalledThemes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(themes) != tt.wantCount {
				t.Errorf("GetInstalledThemes() returned %d themes, want %d: %v", len(themes), tt.wantCount, themes)
			}
		})
	}
}

// =============================================================================
// updateHugoConfigTheme Tests
// =============================================================================

func TestUpdateHugoConfigTheme(t *testing.T) {
	tests := []struct {
		name            string
		initialConfig   string
		configFile      string // "hugo.toml" or "config.toml"
		themeName       string
		wantContains    string
		wantNotContains string
		wantErr         bool
	}{
		{
			name:          "Replace existing theme in hugo.toml",
			initialConfig: "baseURL = \"/\"\ntheme = \"old-theme\"\ntitle = \"Test\"\n",
			configFile:    "hugo.toml",
			themeName:     "new-theme",
			wantContains:  `theme = "new-theme"`,
		},
		{
			name:          "Add theme when not present",
			initialConfig: "baseURL = \"/\"\ntitle = \"Test\"\n",
			configFile:    "hugo.toml",
			themeName:     "ananke",
			wantContains:  `theme = "ananke"`,
		},
		{
			name:          "Works with config.toml",
			initialConfig: "baseURL = \"/\"\ntheme = \"old\"\n",
			configFile:    "config.toml",
			themeName:     "new",
			wantContains:  `theme = "new"`,
		},
		{
			name:            "Does not match themeDir as theme",
			initialConfig:   "themeDir = \"mythemes\"\nbaseURL = \"/\"\n",
			configFile:      "hugo.toml",
			themeName:       "ananke",
			wantContains:    `theme = "ananke"`,
			wantNotContains: `themeDir = "ananke"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, tt.configFile)
			if err := os.WriteFile(configPath, []byte(tt.initialConfig), 0644); err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			err := updateHugoConfigTheme(tmpDir, tt.themeName)

			if (err != nil) != tt.wantErr {
				t.Fatalf("updateHugoConfigTheme() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read result: %v", err)
			}
			result := string(content)

			if tt.wantContains != "" {
				if !strings.Contains(result, tt.wantContains) {
					t.Errorf("Result missing %q\nGot:\n%s", tt.wantContains, result)
				}
			}
			if tt.wantNotContains != "" {
				if strings.Contains(result, tt.wantNotContains) {
					t.Errorf("Result should not contain %q\nGot:\n%s", tt.wantNotContains, result)
				}
			}
		})
	}
}

func TestUpdateHugoConfigTheme_NoConfig(t *testing.T) {
	err := updateHugoConfigTheme(t.TempDir(), "ananke")
	if err == nil {
		t.Fatal("Expected error when no config file exists")
	}
}
