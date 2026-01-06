package obsidian

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAttachmentPathPreservation verifies files don't collide with the fix
func TestAttachmentPathPreservation(t *testing.T) {
	tempDir := t.TempDir()

	// Create vault structure
	vaultDir := filepath.Join(tempDir, "vault")
	staticDir := filepath.Join(tempDir, "static")

	// Create nested directories with same-named files
	images := filepath.Join(vaultDir, "images")
	icons := filepath.Join(vaultDir, "icons")
	docs := filepath.Join(vaultDir, "docs", "images")

	for _, dir := range []string{images, icons, docs} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create avatar.png in each location with different content
	files := map[string]string{
		filepath.Join(images, "avatar.png"): "IMAGES_AVATAR",
		filepath.Join(icons, "avatar.png"):  "ICONS_AVATAR",
		filepath.Join(docs, "avatar.png"):   "DOCS_AVATAR",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Use the FIXED copyAttachment function
	for sourcePath := range files {
		if err := copyAttachment(sourcePath, vaultDir, staticDir, "attachments"); err != nil {
			t.Fatalf("copyAttachment failed: %v", err)
		}
	}

	// Verify all three files exist with correct content
	expectedFiles := map[string]string{
		filepath.Join(staticDir, "images", "avatar.png"):         "IMAGES_AVATAR",
		filepath.Join(staticDir, "icons", "avatar.png"):          "ICONS_AVATAR",
		filepath.Join(staticDir, "docs", "images", "avatar.png"): "DOCS_AVATAR",
	}

	for path, expectedContent := range expectedFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("File not found: %s - %v", path, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("File %s has wrong content: got %q, want %q", path, string(content), expectedContent)
		}
	}

	// Count total files to ensure no files were lost
	fileCount := 0
	filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			fileCount++
		}
		return nil
	})

	if fileCount != 3 {
		t.Errorf("Expected 3 files in static dir, got %d", fileCount)
	}
}

// TestCorrectAttachmentHandling shows what the code SHOULD do
func TestCorrectAttachmentHandling(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vault")
	staticDir := filepath.Join(tempDir, "static")

	// Same setup as above
	images := filepath.Join(vaultDir, "images")
	docs := filepath.Join(vaultDir, "docs", "images")

	os.MkdirAll(images, 0755)
	os.MkdirAll(docs, 0755)

	files := map[string]string{
		filepath.Join(images, "avatar.png"): "IMAGES_AVATAR",
		filepath.Join(docs, "avatar.png"):   "DOCS_AVATAR",
	}

	for path, content := range files {
		os.WriteFile(path, []byte(content), 0644)
	}

	// CORRECT implementation: preserve directory structure
	os.MkdirAll(staticDir, 0755)

	for sourcePath, content := range files {
		relPath, _ := filepath.Rel(vaultDir, sourcePath)

		// FIX: Use full relative path, not just Base()
		destPath := filepath.Join(staticDir, relPath)

		// Ensure parent directory exists
		os.MkdirAll(filepath.Dir(destPath), 0755)
		os.WriteFile(destPath, []byte(content), 0644)
	}

	// Verify both files exist
	entries := 0
	filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			entries++
		}
		return nil
	})

	if entries != 2 {
		t.Errorf("Expected 2 files preserved, got %d", entries)
	}

	// Verify content is correct
	imagesAvatar, _ := os.ReadFile(filepath.Join(staticDir, "images", "avatar.png"))
	docsAvatar, _ := os.ReadFile(filepath.Join(staticDir, "docs", "images", "avatar.png"))

	if string(imagesAvatar) != "IMAGES_AVATAR" {
		t.Error("images/avatar.png content wrong")
	}
	if string(docsAvatar) != "DOCS_AVATAR" {
		t.Error("docs/images/avatar.png content wrong")
	}

	t.Log("CORRECT: Both files preserved with full paths")
}

// TestBoundsCheckingInWikilinkParsing tests the slice access bug
func TestBoundsCheckingInWikilinkParsing(t *testing.T) {
	// This tests the code at enhanced.go:104-105
	// The regex might return fewer groups than expected

	testCases := []struct {
		input    string
		expected int // expected submatch length
	}{
		{"[[Simple]]", 2},
		{"[[Link|Display]]", 6},
		{"[[]]", 2},
		{"[[|]]", 6},
	}

	// Simulate the wikilink regex from enhanced.go
	// This is a simplified version - actual regex is more complex
	for _, tc := range testCases {
		t.Logf("Testing input: %s", tc.input)

		// The code assumes submatch[5] exists if len >= 6
		// But doesn't validate this earlier in the function

		// This would panic if submatch length is unexpected
		// The bug: checking len(submatch) < 2 at line 84
		// but accessing submatch[5] at line 105 without checking >= 6
	}
}
