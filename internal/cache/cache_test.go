package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Check that cache directory was created
	cacheDir := filepath.Join(tmpDir, CacheDir)
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}

	// Check that database file exists
	dbPath := filepath.Join(cacheDir, CacheDBName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Cache database was not created")
	}
}

func TestSaveAndGetFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create test file record
	record := FileRecord{
		Path:         "test.html",
		Hash:         "abc123",
		Size:         1024,
		ModTime:      time.Now(),
		BlobID:       "blob-123",
		LastDeployed: time.Now(),
	}

	// Save file record
	if err := manager.SaveFile(record); err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Retrieve file record
	retrieved, err := manager.GetFile("test.html")
	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}

	if retrieved == nil {
		t.Fatal("File record not found")
	}

	// Verify fields
	if retrieved.Path != record.Path {
		t.Errorf("Path mismatch: got %s, want %s", retrieved.Path, record.Path)
	}

	if retrieved.Hash != record.Hash {
		t.Errorf("Hash mismatch: got %s, want %s", retrieved.Hash, record.Hash)
	}

	if retrieved.BlobID != record.BlobID {
		t.Errorf("BlobID mismatch: got %s, want %s", retrieved.BlobID, record.BlobID)
	}
}

func TestSaveManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create test manifest
	manifest := &BuildManifest{
		SiteRoot:     tmpDir,
		BuildTime:    time.Now(),
		ProjectID:    "project-123",
		LastDeployID: "deploy-456",
		Files: map[string]FileRecord{
			"index.html": {
				Path:    "index.html",
				Hash:    "hash1",
				Size:    2048,
				ModTime: time.Now(),
			},
			"style.css": {
				Path:    "style.css",
				Hash:    "hash2",
				Size:    512,
				ModTime: time.Now(),
			},
		},
	}

	// Save manifest
	if err := manager.SaveManifest(manifest); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Retrieve manifest
	retrieved, err := manager.GetLatestManifest()
	if err != nil {
		t.Fatalf("Failed to get manifest: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Manifest not found")
	}

	// Verify fields
	if retrieved.ProjectID != manifest.ProjectID {
		t.Errorf("ProjectID mismatch: got %s, want %s", retrieved.ProjectID, manifest.ProjectID)
	}

	if len(retrieved.Files) != len(manifest.Files) {
		t.Errorf("Files count mismatch: got %d, want %d", len(retrieved.Files), len(manifest.Files))
	}
}

func TestComputeChangeSet(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create test files
	testFile1 := filepath.Join(tmpDir, "file1.txt")
	testFile2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(testFile1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := os.WriteFile(testFile2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// First changeset (no cache) - all files should be new
	changes1, err := manager.ComputeChangeSet(tmpDir)
	if err != nil {
		t.Fatalf("Failed to compute changeset: %v", err)
	}

	if len(changes1.Added) != 2 {
		t.Errorf("Expected 2 added files, got %d", len(changes1.Added))
	}

	// Save manifest
	hashes, err := HashDirectory(tmpDir)
	if err != nil {
		t.Fatalf("Failed to hash directory: %v", err)
	}

	manifest := &BuildManifest{
		SiteRoot:  tmpDir,
		BuildTime: time.Now(),
		Files:     make(map[string]FileRecord),
	}

	for path, hash := range hashes {
		info, _ := os.Stat(filepath.Join(tmpDir, path))
		manifest.Files[path] = FileRecord{
			Path:    path,
			Hash:    hash,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
	}

	if err := manager.SaveManifest(manifest); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Second changeset (with cache) - no changes
	changes2, err := manager.ComputeChangeSet(tmpDir)
	if err != nil {
		t.Fatalf("Failed to compute changeset: %v", err)
	}

	if len(changes2.Unchanged) != 2 {
		t.Errorf("Expected 2 unchanged files, got %d", len(changes2.Unchanged))
	}

	// Modify a file
	if err := os.WriteFile(testFile1, []byte("modified content"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Third changeset - one modified file
	changes3, err := manager.ComputeChangeSet(tmpDir)
	if err != nil {
		t.Fatalf("Failed to compute changeset: %v", err)
	}

	if len(changes3.Modified) != 1 {
		t.Errorf("Expected 1 modified file, got %d", len(changes3.Modified))
	}

	if len(changes3.Unchanged) != 1 {
		t.Errorf("Expected 1 unchanged file, got %d", len(changes3.Unchanged))
	}
}

func TestUpdateBlobID(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Save initial file record without blob ID
	record := FileRecord{
		Path:    "test.html",
		Hash:    "abc123",
		Size:    1024,
		ModTime: time.Now(),
	}

	if err := manager.SaveFile(record); err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Update blob ID
	if err := manager.UpdateBlobID("test.html", "blob-xyz"); err != nil {
		t.Fatalf("Failed to update blob ID: %v", err)
	}

	// Retrieve and verify
	retrieved, err := manager.GetFile("test.html")
	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}

	if retrieved.BlobID != "blob-xyz" {
		t.Errorf("BlobID not updated: got %s, want blob-xyz", retrieved.BlobID)
	}

	if retrieved.LastDeployed.IsZero() {
		t.Error("LastDeployed should be set after update")
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Add some data
	record := FileRecord{
		Path:    "test.html",
		Hash:    "abc123",
		Size:    1024,
		ModTime: time.Now(),
	}

	if err := manager.SaveFile(record); err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Clear cache
	if err := manager.Clear(); err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Verify cache is empty
	files, err := manager.GetAllFiles()
	if err != nil {
		t.Fatalf("Failed to get all files: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Cache not empty after clear: got %d files", len(files))
	}
}
