package compress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldCompress(t *testing.T) {
	c := New(DefaultConfig())

	tests := []struct {
		filename string
		expected bool
	}{
		// Should compress
		{"index.html", true},
		{"style.css", true},
		{"script.js", true},
		{"data.json", true},
		{"image.svg", true},
		{"readme.txt", true},

		// Should NOT compress
		{"photo.jpg", false},
		{"video.mp4", false},
		{"archive.zip", false},
		{"font.woff2", false},
		{"binary.exe", false},
	}

	for _, tt := range tests {
		got := c.ShouldCompress(tt.filename)
		if got != tt.expected {
			t.Errorf("ShouldCompress(%s) = %v, want %v", tt.filename, got, tt.expected)
		}
	}
}

func TestCompressFile(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(DefaultConfig())

	// Create a test file with compressible content
	testContent := []byte("This is a test file with some repetitive content. " +
		"This is a test file with some repetitive content. " +
		"This is a test file with some repetitive content.")

	inputPath := filepath.Join(tmpDir, "test.txt")
	outputPath := filepath.Join(tmpDir, "test.txt.br")

	if err := os.WriteFile(inputPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Compress the file
	result, err := c.CompressFile(inputPath, outputPath)
	if err != nil {
		t.Fatalf("CompressFile failed: %v", err)
	}

	// Verify result
	if result.OriginalSize != len(testContent) {
		t.Errorf("OriginalSize = %d, want %d", result.OriginalSize, len(testContent))
	}

	if result.CompressedSize >= result.OriginalSize {
		t.Error("Compressed size should be smaller than original")
	}

	if !result.Compressed {
		t.Error("File should have been compressed")
	}

	// Verify compressed file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Compressed file was not created")
	}
}

func TestCompressBuffer(t *testing.T) {
	data := []byte("This is a test. This is a test. This is a test.")

	compressed, err := CompressBuffer(data, 6)
	if err != nil {
		t.Fatalf("CompressBuffer failed: %v", err)
	}

	if len(compressed) >= len(data) {
		t.Error("Compressed data should be smaller than original")
	}

	// Test decompression
	decompressed, err := DecompressBuffer(compressed)
	if err != nil {
		t.Fatalf("DecompressBuffer failed: %v", err)
	}

	if string(decompressed) != string(data) {
		t.Error("Decompressed data doesn't match original")
	}
}

func TestCompressDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(DefaultConfig())

	// Create test files
	files := map[string][]byte{
		"index.html": []byte("<html><body>Hello World Hello World Hello World</body></html>"),
		"style.css":  []byte("body { margin: 0; } body { margin: 0; } body { margin: 0; }"),
		"image.jpg":  []byte("fake-jpg-data"),
		"script.js":  []byte("function test() { return 'hello'; }"),
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Compress directory
	stats, err := c.CompressDirectory(tmpDir)
	if err != nil {
		t.Fatalf("CompressDirectory failed: %v", err)
	}

	// Verify stats
	if stats.Compressed+stats.NotWorthCompressing < 3 {
		t.Errorf("Expected at least 3 compressed files, got %d", stats.Compressed+stats.NotWorthCompressing)
	}

	if stats.Skipped != 1 { // image.jpg should be skipped
		t.Errorf("Expected 1 skipped file, got %d", stats.Skipped)
	}

	if stats.TotalOriginalSize == 0 {
		t.Error("TotalOriginalSize should not be zero")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Enabled {
		t.Error("Default config should have compression enabled")
	}

	if cfg.BrotliLevel != 6 {
		t.Errorf("Default Brotli level = %d, want 6", cfg.BrotliLevel)
	}

	if len(cfg.SkipExtensions) == 0 {
		t.Error("Default config should have skip extensions")
	}
}
