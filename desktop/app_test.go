package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// =============================================================================
// validateFilePath — Security-critical path traversal prevention
// =============================================================================

func TestValidateFilePath_InsideHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	// Path inside home should be accepted
	testPath := filepath.Join(home, "Documents", "test.txt")
	result, err := validateFilePath(testPath)
	if err != nil {
		t.Errorf("expected no error for path inside home, got: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result path")
	}
}

func TestValidateFilePath_HomeItself(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	// Home directory itself should be accepted
	result, err := validateFilePath(home)
	if err != nil {
		t.Errorf("expected no error for home dir itself, got: %v", err)
	}
	absHome, _ := filepath.Abs(home)
	if result != absHome {
		t.Errorf("expected %s, got %s", absHome, result)
	}
}

func TestValidateFilePath_OutsideHome(t *testing.T) {
	outsidePaths := []string{
		"/etc/passwd",
		"/tmp/evil",
		"/var/log/syslog",
	}
	if runtime.GOOS == "windows" {
		outsidePaths = []string{
			`C:\Windows\System32\config`,
			`C:\temp\evil`,
		}
	}

	for _, p := range outsidePaths {
		_, err := validateFilePath(p)
		if err == nil {
			t.Errorf("expected error for path outside home: %s", p)
		}
		if err != nil && !strings.Contains(err.Error(), "access denied") {
			t.Errorf("expected 'access denied' error for %s, got: %v", p, err)
		}
	}
}

func TestValidateFilePath_TraversalAttempt(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	// Path traversal: go into home then escape with ..
	traversalPath := filepath.Join(home, "Documents", "..", "..", "..", "etc", "passwd")
	_, err = validateFilePath(traversalPath)
	if err == nil {
		t.Error("expected error for path traversal attempt")
	}
}

func TestValidateFilePath_RelativePath(t *testing.T) {
	// Relative paths get resolved to CWD — only valid if CWD is inside home
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get cwd: %v", err)
	}

	result, err := validateFilePath("relative/path/file.txt")
	expected := filepath.Join(cwd, "relative", "path", "file.txt")

	// If CWD is inside home, should succeed; otherwise should fail
	absHome, _ := filepath.Abs(home)
	if strings.HasPrefix(cwd, absHome) {
		if err != nil {
			t.Errorf("expected success for relative path under home, got: %v", err)
		}
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	} else {
		if err == nil {
			t.Error("expected error for relative path outside home")
		}
	}
}

func TestValidateFilePath_CleansDots(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	// Path with . and .. that still resolves inside home
	dirtyPath := filepath.Join(home, "Documents", ".", "subdir", "..", "file.txt")
	result, err := validateFilePath(dirtyPath)
	if err != nil {
		t.Errorf("expected no error for cleaned path inside home, got: %v", err)
	}

	// Should be cleaned
	expected := filepath.Join(home, "Documents", "file.txt")
	absExpected, _ := filepath.Abs(expected)
	if result != absExpected {
		t.Errorf("expected cleaned path %s, got %s", absExpected, result)
	}
}

func TestValidateFilePath_EmptyPath(t *testing.T) {
	// Empty path resolves to CWD
	result, err := validateFilePath("")
	if err != nil {
		// May error if CWD is outside home (e.g., in CI)
		if !strings.Contains(err.Error(), "access denied") {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}
	// If it succeeds, result should be an absolute path
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got: %s", result)
	}
}

func TestValidateFilePath_NestedDeep(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	// Deeply nested path inside home
	deepPath := filepath.Join(home, "a", "b", "c", "d", "e", "f", "g", "file.txt")
	result, err := validateFilePath(deepPath)
	if err != nil {
		t.Errorf("expected no error for deep path inside home, got: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// =============================================================================
// countFilesRecursive — File/folder counting
// =============================================================================

func TestCountFilesRecursive_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	fileCount, folderCount, totalSize := countFilesRecursive(dir)
	if fileCount != 0 {
		t.Errorf("expected 0 files, got %d", fileCount)
	}
	if folderCount != 0 {
		t.Errorf("expected 0 folders, got %d", folderCount)
	}
	if totalSize != 0 {
		t.Errorf("expected 0 size, got %d", totalSize)
	}
}

func TestCountFilesRecursive_FilesOnly(t *testing.T) {
	dir := t.TempDir()

	// Create 3 files with known sizes
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		os.WriteFile(filepath.Join(dir, name), []byte("hello"), 0644)
	}

	fileCount, folderCount, totalSize := countFilesRecursive(dir)
	if fileCount != 3 {
		t.Errorf("expected 3 files, got %d", fileCount)
	}
	if folderCount != 0 {
		t.Errorf("expected 0 folders, got %d", folderCount)
	}
	if totalSize != 15 { // 3 files × 5 bytes
		t.Errorf("expected 15 bytes, got %d", totalSize)
	}
}

func TestCountFilesRecursive_NestedDirs(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure:
	// dir/
	//   file1.txt (5 bytes)
	//   sub1/
	//     file2.txt (5 bytes)
	//     sub2/
	//       file3.txt (5 bytes)
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("hello"), 0644)
	os.MkdirAll(filepath.Join(dir, "sub1", "sub2"), 0755)
	os.WriteFile(filepath.Join(dir, "sub1", "file2.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(dir, "sub1", "sub2", "file3.txt"), []byte("hello"), 0644)

	fileCount, folderCount, totalSize := countFilesRecursive(dir)
	if fileCount != 3 {
		t.Errorf("expected 3 files, got %d", fileCount)
	}
	if folderCount != 2 { // sub1 + sub2
		t.Errorf("expected 2 folders, got %d", folderCount)
	}
	if totalSize != 15 {
		t.Errorf("expected 15 bytes, got %d", totalSize)
	}
}

func TestCountFilesRecursive_SkipsHiddenFiles(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("secret"), 0644)
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("gitconfig"), 0644)

	fileCount, folderCount, totalSize := countFilesRecursive(dir)
	if fileCount != 1 {
		t.Errorf("expected 1 file (hidden skipped), got %d", fileCount)
	}
	if folderCount != 0 {
		t.Errorf("expected 0 folders (hidden skipped), got %d", folderCount)
	}
	if totalSize != 4 { // "data" = 4 bytes
		t.Errorf("expected 4 bytes, got %d", totalSize)
	}
}

func TestCountFilesRecursive_NonExistentDir(t *testing.T) {
	fileCount, folderCount, totalSize := countFilesRecursive("/nonexistent/path/does/not/exist")
	if fileCount != 0 || folderCount != 0 || totalSize != 0 {
		t.Error("expected all zeros for non-existent directory")
	}
}

func TestCountFilesRecursive_DifferentFileSizes(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "small.txt"), []byte("a"), 0644)            // 1 byte
	os.WriteFile(filepath.Join(dir, "medium.txt"), make([]byte, 1024), 0644)    // 1 KB
	os.WriteFile(filepath.Join(dir, "large.txt"), make([]byte, 1024*100), 0644) // 100 KB

	fileCount, folderCount, totalSize := countFilesRecursive(dir)
	if fileCount != 3 {
		t.Errorf("expected 3 files, got %d", fileCount)
	}
	if folderCount != 0 {
		t.Errorf("expected 0 folders, got %d", folderCount)
	}
	expectedSize := int64(1 + 1024 + 1024*100)
	if totalSize != expectedSize {
		t.Errorf("expected %d bytes, got %d", expectedSize, totalSize)
	}
}

// =============================================================================
// copyDirWithDepth — Recursive directory copy with safeguards
// =============================================================================

func TestCopyDir_BasicCopy(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	// Create source structure
	os.WriteFile(filepath.Join(src, "file1.txt"), []byte("content1"), 0644)
	os.MkdirAll(filepath.Join(src, "subdir"), 0755)
	os.WriteFile(filepath.Join(src, "subdir", "file2.txt"), []byte("content2"), 0644)

	err := copyDir(src, dst)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify file1.txt
	data, err := os.ReadFile(filepath.Join(dst, "file1.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file1: %v", err)
	}
	if string(data) != "content1" {
		t.Errorf("file1 content mismatch: got %q", string(data))
	}

	// Verify subdir/file2.txt
	data, err = os.ReadFile(filepath.Join(dst, "subdir", "file2.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file2: %v", err)
	}
	if string(data) != "content2" {
		t.Errorf("file2 content mismatch: got %q", string(data))
	}
}

func TestCopyDir_PreservesFileMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file mode tests not reliable on Windows")
	}

	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	// Create file with specific mode
	os.WriteFile(filepath.Join(src, "script.sh"), []byte("#!/bin/sh"), 0755)

	err := copyDir(src, dst)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(dst, "script.sh"))
	if err != nil {
		t.Fatalf("failed to stat copied file: %v", err)
	}

	if info.Mode().Perm()&0100 == 0 {
		t.Error("expected executable bit to be preserved")
	}
}

func TestCopyDir_SameSourceAndDest(t *testing.T) {
	dir := t.TempDir()

	err := copyDir(dir, dir)
	if err == nil {
		t.Error("expected error for same source and destination")
	}
	if err != nil && !strings.Contains(err.Error(), "same") {
		t.Errorf("expected 'same' error, got: %v", err)
	}
}

func TestCopyDir_DestInsideSource(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(src, "subdir", "dest")

	err := copyDir(src, dst)
	if err == nil {
		t.Error("expected error for destination inside source")
	}
	if err != nil && !strings.Contains(err.Error(), "into itself") {
		t.Errorf("expected 'into itself' error, got: %v", err)
	}
}

func TestCopyDir_NonExistentSource(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "dest")

	err := copyDir("/nonexistent/source/path", dst)
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}

func TestCopyDir_EmptyDir(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	err := copyDir(src, dst)
	if err != nil {
		t.Fatalf("copyDir failed for empty dir: %v", err)
	}

	// Verify destination was created
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("destination not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("destination should be a directory")
	}
}

func TestCopyDir_DeeplyNested(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	// Create 10 levels deep
	deepPath := src
	for i := 0; i < 10; i++ {
		deepPath = filepath.Join(deepPath, "level")
	}
	os.MkdirAll(deepPath, 0755)
	os.WriteFile(filepath.Join(deepPath, "deep.txt"), []byte("deep"), 0644)

	err := copyDir(src, dst)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify deepest file was copied
	dstDeep := dst
	for i := 0; i < 10; i++ {
		dstDeep = filepath.Join(dstDeep, "level")
	}
	data, err := os.ReadFile(filepath.Join(dstDeep, "deep.txt"))
	if err != nil {
		t.Fatalf("deep file not copied: %v", err)
	}
	if string(data) != "deep" {
		t.Errorf("deep file content mismatch: got %q", string(data))
	}
}

func TestCopyDirWithDepth_ExceedsMaxDepth(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	err := copyDirWithDepth(src, dst, 101) // Over max depth of 100
	if err == nil {
		t.Error("expected error for exceeding max depth")
	}
	if err != nil && !strings.Contains(err.Error(), "maximum directory depth") {
		t.Errorf("expected 'maximum directory depth' error, got: %v", err)
	}
}

func TestCopyDir_SkipsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests not reliable on Windows")
	}

	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	// Create a real file and a symlink
	os.WriteFile(filepath.Join(src, "real.txt"), []byte("real"), 0644)
	os.Symlink(filepath.Join(src, "real.txt"), filepath.Join(src, "link.txt"))

	err := copyDir(src, dst)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Real file should be copied
	if _, err := os.Stat(filepath.Join(dst, "real.txt")); err != nil {
		t.Error("real file should be copied")
	}

	// Symlink should be skipped
	if _, err := os.Stat(filepath.Join(dst, "link.txt")); err == nil {
		t.Error("symlink should be skipped")
	}
}

func TestCopyDir_MultipleFilesAndDirs(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	// Create complex structure
	dirs := []string{"css", "js", "images", "content/posts"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(src, d), 0755)
	}
	files := map[string]string{
		"index.html":           "<html>",
		"css/style.css":        "body{}",
		"js/app.js":            "console.log('hi')",
		"images/logo.png":      "fake-png-data",
		"content/posts/one.md": "---\ntitle: One\n---",
	}
	for name, content := range files {
		os.WriteFile(filepath.Join(src, name), []byte(content), 0644)
	}

	err := copyDir(src, dst)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify all files exist with correct content
	for name, expectedContent := range files {
		data, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Errorf("missing copied file %s: %v", name, err)
			continue
		}
		if string(data) != expectedContent {
			t.Errorf("content mismatch for %s: got %q, want %q", name, string(data), expectedContent)
		}
	}
}

// =============================================================================
// copyFile — Single file copy
// =============================================================================

func TestCopyFile_Basic(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	os.WriteFile(src, []byte("file content"), 0644)

	err := copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read dst: %v", err)
	}
	if string(data) != "file content" {
		t.Errorf("content mismatch: got %q", string(data))
	}
}

func TestCopyFile_NonExistentSource(t *testing.T) {
	dir := t.TempDir()
	err := copyFile("/nonexistent/file.txt", filepath.Join(dir, "dst.txt"))
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}

func TestCopyFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "empty.txt")
	dst := filepath.Join(dir, "copy.txt")

	os.WriteFile(src, []byte{}, 0644)

	err := copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile failed for empty file: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("failed to stat copy: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("expected 0 bytes, got %d", info.Size())
	}
}

func TestCopyFile_LargeFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "large.bin")
	dst := filepath.Join(dir, "copy.bin")

	// 1 MB file
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	os.WriteFile(src, data, 0644)

	err := copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile failed for large file: %v", err)
	}

	copied, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read copy: %v", err)
	}
	if len(copied) != len(data) {
		t.Errorf("size mismatch: expected %d, got %d", len(data), len(copied))
	}
	// Spot-check a few bytes
	if copied[0] != 0 || copied[255] != 255 || copied[256] != 0 {
		t.Error("content verification failed")
	}
}

// =============================================================================
// findUniquePath — Unique filename generation
// =============================================================================

func TestFindUniquePath_NoConflict(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newfile.txt")

	result := findUniquePath(path)
	if result != path {
		t.Errorf("expected original path when no conflict, got %s", result)
	}
}

func TestFindUniquePath_WithExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")

	// Create the conflicting file
	os.WriteFile(path, []byte("existing"), 0644)

	result := findUniquePath(path)
	expected := filepath.Join(dir, "file (1).txt")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestFindUniquePath_WithoutExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folder")

	// Create conflicting directory
	os.MkdirAll(path, 0755)

	result := findUniquePath(path)
	expected := filepath.Join(dir, "folder (1)")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestFindUniquePath_MultipleConflicts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.pdf")

	// Create original + (1) + (2)
	os.WriteFile(path, []byte("v1"), 0644)
	os.WriteFile(filepath.Join(dir, "doc (1).pdf"), []byte("v2"), 0644)
	os.WriteFile(filepath.Join(dir, "doc (2).pdf"), []byte("v3"), 0644)

	result := findUniquePath(path)
	expected := filepath.Join(dir, "doc (3).pdf")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestFindUniquePath_DirectoryConflict(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "project")

	// Create conflicting directories
	os.MkdirAll(path, 0755)
	os.MkdirAll(filepath.Join(dir, "project (1)"), 0755)

	result := findUniquePath(path)
	expected := filepath.Join(dir, "project (2)")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestFindUniquePath_DotFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	os.WriteFile(path, []byte("SECRET=123"), 0644)

	result := findUniquePath(path)
	// .env has ext=".env" and nameWithoutExt="" so result is " (1).env"
	// This is valid behavior — the function works with what filepath gives it
	if result == path {
		t.Error("expected different path for existing .env file")
	}
}

// =============================================================================
// getDirectoryDepth — Directory tree depth calculation
// =============================================================================

func TestGetDirectoryDepth_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	depth, err := getDirectoryDepth(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if depth != 0 {
		t.Errorf("expected depth 0 for empty dir, got %d", depth)
	}
}

func TestGetDirectoryDepth_FlatFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)

	depth, err := getDirectoryDepth(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if depth != 0 {
		t.Errorf("expected depth 0 for flat dir (files only), got %d", depth)
	}
}

func TestGetDirectoryDepth_SingleLevel(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)

	depth, err := getDirectoryDepth(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if depth != 1 {
		t.Errorf("expected depth 1, got %d", depth)
	}
}

func TestGetDirectoryDepth_DeepNesting(t *testing.T) {
	dir := t.TempDir()

	// Create 5 levels deep
	deepPath := dir
	for i := 0; i < 5; i++ {
		deepPath = filepath.Join(deepPath, "level")
	}
	os.MkdirAll(deepPath, 0755)

	depth, err := getDirectoryDepth(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if depth != 5 {
		t.Errorf("expected depth 5, got %d", depth)
	}
}

func TestGetDirectoryDepth_AsymmetricTree(t *testing.T) {
	dir := t.TempDir()

	// Left branch: 2 deep
	os.MkdirAll(filepath.Join(dir, "a", "b"), 0755)
	// Right branch: 4 deep
	os.MkdirAll(filepath.Join(dir, "x", "y", "z", "w"), 0755)

	depth, err := getDirectoryDepth(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if depth != 4 {
		t.Errorf("expected max depth 4 (right branch), got %d", depth)
	}
}

func TestGetDirectoryDepth_SkipsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests not reliable on Windows")
	}

	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	os.MkdirAll(filepath.Join(realDir, "deep", "deeper"), 0755)

	// Create symlink that would add depth if followed
	os.Symlink(realDir, filepath.Join(dir, "link"))

	depth, err := getDirectoryDepth(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only count real/ branch (depth 3: real, deep, deeper)
	// Not the symlink branch
	if depth != 3 {
		t.Errorf("expected depth 3 (symlinks skipped), got %d", depth)
	}
}

func TestGetDirectoryDepth_NonExistentDir(t *testing.T) {
	_, err := getDirectoryDepth("/nonexistent/path")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

// =============================================================================
// lookPath — Executable discovery
// =============================================================================

func TestLookPath_KnownBinary(t *testing.T) {
	// "ls" exists on all Unix systems, "cmd.exe" on Windows
	var target string
	if runtime.GOOS == "windows" {
		target = "cmd"
	} else {
		target = "ls"
	}

	path, err := lookPath(target)
	if err != nil {
		t.Errorf("expected to find %s: %v", target, err)
	}
	if path == "" {
		t.Errorf("expected non-empty path for %s", target)
	}

	// Verify the path actually exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("returned path %s does not exist: %v", path, err)
	}
}

func TestLookPath_NonExistentBinary(t *testing.T) {
	_, err := lookPath("walgo_nonexistent_binary_12345")
	if err == nil {
		t.Error("expected error for non-existent binary")
	}
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestLookPath_ReturnsAbsolutePath(t *testing.T) {
	var target string
	if runtime.GOOS == "windows" {
		target = "cmd"
	} else {
		target = "ls"
	}

	path, err := lookPath(target)
	if err != nil {
		t.Skipf("cannot find %s, skipping: %v", target, err)
	}

	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got: %s", path)
	}
}

func TestLookPath_Go(t *testing.T) {
	// Go binary should be available since we're running tests with it
	path, err := lookPath("go")
	if err != nil {
		t.Skipf("go not in path: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path for go")
	}
}

func TestLookPath_EmptyName(t *testing.T) {
	_, err := lookPath("")
	if err == nil {
		t.Error("expected error for empty binary name")
	}
}

// =============================================================================
// NewApp — Constructor
// =============================================================================

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.serveCmd != nil {
		t.Error("serveCmd should be nil on new app")
	}
	if app.serverPort != 0 {
		t.Error("serverPort should be 0 on new app")
	}
	if app.serveSitePath != "" {
		t.Error("serveSitePath should be empty on new app")
	}
}

// =============================================================================
// Integration: validateFilePath + file operations
// =============================================================================

func TestValidateFilePath_WithRealFile(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	// Create a real temp file inside home
	tmpDir := filepath.Join(home, ".walgo-test-tmp")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	// Validate the path
	safePath, err := validateFilePath(tmpFile)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Read through the safe path
	data, err := os.ReadFile(safePath)
	if err != nil {
		t.Fatalf("failed to read via safe path: %v", err)
	}
	if string(data) != "test" {
		t.Errorf("content mismatch via safe path: got %q", string(data))
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkCountFilesRecursive(b *testing.B) {
	dir := b.TempDir()

	// Create 100 files across 10 directories
	for i := 0; i < 10; i++ {
		subDir := filepath.Join(dir, "dir"+strings.Repeat("x", i))
		os.MkdirAll(subDir, 0755)
		for j := 0; j < 10; j++ {
			os.WriteFile(filepath.Join(subDir, "file"+strings.Repeat("x", j)+".txt"), []byte("data"), 0644)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countFilesRecursive(dir)
	}
}

func BenchmarkValidateFilePath(b *testing.B) {
	home, _ := os.UserHomeDir()
	testPath := filepath.Join(home, "Documents", "test.txt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateFilePath(testPath)
	}
}

func BenchmarkFindUniquePath_NoConflict(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findUniquePath(path)
	}
}

func BenchmarkCopyDir_Small(b *testing.B) {
	src := b.TempDir()
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(src, "b.txt"), []byte("world"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := filepath.Join(b.TempDir(), "dest")
		copyDir(src, dst)
	}
}
