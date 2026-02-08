package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDeployHelper(t *testing.T) {
	t.Run("creates helper successfully", func(t *testing.T) {
		tmpDir := t.TempDir()

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		if helper.manager == nil {
			t.Error("manager should not be nil")
		}
		if helper.siteRoot != tmpDir {
			t.Errorf("siteRoot = %q, want %q", helper.siteRoot, tmpDir)
		}
	})

	t.Run("creates cache directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		cacheDir := filepath.Join(tmpDir, CacheDir)
		if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
			t.Error("cache directory should be created")
		}
	})
}

func TestDeployHelperClose(t *testing.T) {
	tmpDir := t.TempDir()

	helper, err := NewDeployHelper(tmpDir)
	if err != nil {
		t.Fatalf("NewDeployHelper() error = %v", err)
	}

	// Close should not error
	if err := helper.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Double close should also not error (Manager.Close is safe to call multiple times)
	if err := helper.Close(); err != nil {
		t.Errorf("second Close() error = %v", err)
	}
}

func TestPrepareDeployment(t *testing.T) {
	t.Run("first deployment - all files new", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, "build")
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			t.Fatalf("failed to create build dir: %v", err)
		}

		// Create test files
		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html>hello</html>"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(buildDir, "style.css"), []byte("body { color: red; }"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		plan, err := helper.PrepareDeployment(buildDir)
		if err != nil {
			t.Fatalf("PrepareDeployment() error = %v", err)
		}

		if plan == nil {
			t.Fatal("plan should not be nil")
		}

		if plan.TotalFiles != 2 {
			t.Errorf("TotalFiles = %d, want 2", plan.TotalFiles)
		}

		if plan.TotalSize <= 0 {
			t.Errorf("TotalSize = %d, should be positive", plan.TotalSize)
		}

		if len(plan.ChangeSet.Added) != 2 {
			t.Errorf("Added = %d, want 2", len(plan.ChangeSet.Added))
		}

		if plan.IsIncremental {
			t.Error("first deployment should not be incremental")
		}
	})

	t.Run("incremental deployment - some files changed", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, "build")
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			t.Fatalf("failed to create build dir: %v", err)
		}

		// Create initial files
		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html>hello</html>"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(buildDir, "style.css"), []byte("body { color: red; }"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		// Finalize first deployment to populate cache
		if err := helper.FinalizeDeployment(buildDir, "proj-1", "deploy-1", nil); err != nil {
			t.Fatalf("FinalizeDeployment() error = %v", err)
		}

		// Modify one file
		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html>updated</html>"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		plan, err := helper.PrepareDeployment(buildDir)
		if err != nil {
			t.Fatalf("PrepareDeployment() error = %v", err)
		}

		if !plan.IsIncremental {
			t.Error("should be incremental deployment")
		}

		if len(plan.ChangeSet.Modified) != 1 {
			t.Errorf("Modified = %d, want 1", len(plan.ChangeSet.Modified))
		}

		if len(plan.ChangeSet.Unchanged) != 1 {
			t.Errorf("Unchanged = %d, want 1", len(plan.ChangeSet.Unchanged))
		}

		if plan.ChangedSize <= 0 {
			t.Errorf("ChangedSize = %d, should be positive for modified files", plan.ChangedSize)
		}

		if plan.ChangedSize >= plan.TotalSize {
			t.Errorf("ChangedSize (%d) should be less than TotalSize (%d) for partial changes", plan.ChangedSize, plan.TotalSize)
		}
	})

	t.Run("no changes since last deployment", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, "build")
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			t.Fatalf("failed to create build dir: %v", err)
		}

		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html>hello</html>"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		// Finalize first deployment
		if err := helper.FinalizeDeployment(buildDir, "proj-1", "deploy-1", nil); err != nil {
			t.Fatalf("FinalizeDeployment() error = %v", err)
		}

		plan, err := helper.PrepareDeployment(buildDir)
		if err != nil {
			t.Fatalf("PrepareDeployment() error = %v", err)
		}

		if len(plan.ChangeSet.Unchanged) != 1 {
			t.Errorf("Unchanged = %d, want 1", len(plan.ChangeSet.Unchanged))
		}

		if len(plan.ChangeSet.Added) != 0 {
			t.Errorf("Added = %d, want 0", len(plan.ChangeSet.Added))
		}

		if len(plan.ChangeSet.Modified) != 0 {
			t.Errorf("Modified = %d, want 0", len(plan.ChangeSet.Modified))
		}

		if plan.ChangedSize != 0 {
			t.Errorf("ChangedSize = %d, want 0", plan.ChangedSize)
		}
	})

	t.Run("new files added since last deployment", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, "build")
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			t.Fatalf("failed to create build dir: %v", err)
		}

		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html>hello</html>"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		// Finalize first deployment
		if err := helper.FinalizeDeployment(buildDir, "proj-1", "deploy-1", nil); err != nil {
			t.Fatalf("FinalizeDeployment() error = %v", err)
		}

		// Add a new file
		if err := os.WriteFile(filepath.Join(buildDir, "new.js"), []byte("console.log('hello');"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		plan, err := helper.PrepareDeployment(buildDir)
		if err != nil {
			t.Fatalf("PrepareDeployment() error = %v", err)
		}

		if len(plan.ChangeSet.Added) != 1 {
			t.Errorf("Added = %d, want 1", len(plan.ChangeSet.Added))
		}

		if len(plan.ChangeSet.Unchanged) != 1 {
			t.Errorf("Unchanged = %d, want 1", len(plan.ChangeSet.Unchanged))
		}

		if plan.TotalFiles != 2 {
			t.Errorf("TotalFiles = %d, want 2", plan.TotalFiles)
		}
	})

	t.Run("nonexistent build directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		_, err = helper.PrepareDeployment("/nonexistent/path")
		if err == nil {
			t.Error("PrepareDeployment() should error for nonexistent directory")
		}
	})
}

func TestFinalizeDeployment(t *testing.T) {
	t.Run("saves manifest with file records", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, "build")
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			t.Fatalf("failed to create build dir: %v", err)
		}

		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html>hello</html>"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(buildDir, "style.css"), []byte("body { color: red; }"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		fileToBlobID := map[string]string{
			"index.html": "blob-index-123",
			"style.css":  "blob-style-456",
		}

		err = helper.FinalizeDeployment(buildDir, "proj-1", "deploy-1", fileToBlobID)
		if err != nil {
			t.Fatalf("FinalizeDeployment() error = %v", err)
		}

		// Verify manifest was saved
		manifest, err := helper.GetLastDeployment()
		if err != nil {
			t.Fatalf("GetLastDeployment() error = %v", err)
		}
		if manifest == nil {
			t.Fatal("manifest should not be nil after FinalizeDeployment")
		}

		if manifest.ProjectID != "proj-1" {
			t.Errorf("ProjectID = %q, want %q", manifest.ProjectID, "proj-1")
		}
		if manifest.LastDeployID != "deploy-1" {
			t.Errorf("LastDeployID = %q, want %q", manifest.LastDeployID, "deploy-1")
		}
		if len(manifest.Files) != 2 {
			t.Errorf("Files count = %d, want 2", len(manifest.Files))
		}

		// Verify blob IDs were set
		if record, ok := manifest.Files["index.html"]; ok {
			if record.BlobID != "blob-index-123" {
				t.Errorf("index.html BlobID = %q, want %q", record.BlobID, "blob-index-123")
			}
			if record.Hash == "" {
				t.Error("index.html Hash should not be empty")
			}
			if record.Size <= 0 {
				t.Errorf("index.html Size = %d, should be positive", record.Size)
			}
		} else {
			t.Error("index.html not found in manifest files")
		}
	})

	t.Run("saves manifest without blob IDs", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, "build")
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			t.Fatalf("failed to create build dir: %v", err)
		}

		if err := os.WriteFile(filepath.Join(buildDir, "page.html"), []byte("<html>page</html>"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		// Pass nil for fileToBlobID
		err = helper.FinalizeDeployment(buildDir, "proj-2", "deploy-2", nil)
		if err != nil {
			t.Fatalf("FinalizeDeployment() error = %v", err)
		}

		manifest, err := helper.GetLastDeployment()
		if err != nil {
			t.Fatalf("GetLastDeployment() error = %v", err)
		}
		if manifest == nil {
			t.Fatal("manifest should not be nil")
		}

		if record, ok := manifest.Files["page.html"]; ok {
			if record.BlobID != "" {
				t.Errorf("BlobID = %q, want empty when not provided", record.BlobID)
			}
		} else {
			t.Error("page.html not found in manifest files")
		}
	})

	t.Run("nonexistent build directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		err = helper.FinalizeDeployment("/nonexistent/path", "proj-1", "deploy-1", nil)
		if err == nil {
			t.Error("FinalizeDeployment() should error for nonexistent directory")
		}
	})
}

func TestGetLastDeployment(t *testing.T) {
	t.Run("no previous deployment", func(t *testing.T) {
		tmpDir := t.TempDir()
		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		manifest, err := helper.GetLastDeployment()
		if err != nil {
			t.Fatalf("GetLastDeployment() error = %v", err)
		}
		if manifest != nil {
			t.Error("manifest should be nil when no previous deployment exists")
		}
	})

	t.Run("returns latest deployment", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, "build")
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			t.Fatalf("failed to create build dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html>hello</html>"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		// Finalize a deployment
		if err := helper.FinalizeDeployment(buildDir, "proj-1", "deploy-1", nil); err != nil {
			t.Fatalf("FinalizeDeployment() error = %v", err)
		}

		manifest, err := helper.GetLastDeployment()
		if err != nil {
			t.Fatalf("GetLastDeployment() error = %v", err)
		}
		if manifest == nil {
			t.Fatal("manifest should not be nil after FinalizeDeployment")
		}

		if manifest.ProjectID != "proj-1" {
			t.Errorf("ProjectID = %q, want %q", manifest.ProjectID, "proj-1")
		}
	})
}

func TestShouldOptimizeFile(t *testing.T) {
	t.Run("new file needs optimization", func(t *testing.T) {
		tmpDir := t.TempDir()
		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		// File not in cache - should need optimization
		shouldOptimize, err := helper.ShouldOptimizeFile("nonexistent.html")
		if err != nil {
			t.Fatalf("ShouldOptimizeFile() error = %v", err)
		}
		if !shouldOptimize {
			t.Error("new file should need optimization")
		}
	})

	t.Run("unchanged file skips optimization", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a file
		filePath := filepath.Join(tmpDir, "style.css")
		fileContent := []byte("body { color: red; }")
		if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		// Hash the file and save a record with matching hash
		hash, err := HashFile(filePath)
		if err != nil {
			t.Fatalf("HashFile() error = %v", err)
		}

		record := FileRecord{
			Path:    "style.css",
			Hash:    hash,
			Size:    int64(len(fileContent)),
			ModTime: time.Now(),
		}
		if err := helper.manager.SaveFile(record); err != nil {
			t.Fatalf("SaveFile() error = %v", err)
		}

		shouldOptimize, err := helper.ShouldOptimizeFile("style.css")
		if err != nil {
			t.Fatalf("ShouldOptimizeFile() error = %v", err)
		}
		if shouldOptimize {
			t.Error("unchanged file should skip optimization")
		}
	})

	t.Run("modified file needs optimization", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a file
		filePath := filepath.Join(tmpDir, "style.css")
		if err := os.WriteFile(filePath, []byte("body { color: red; }"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		helper, err := NewDeployHelper(tmpDir)
		if err != nil {
			t.Fatalf("NewDeployHelper() error = %v", err)
		}
		defer helper.Close()

		// Save a record with old hash
		record := FileRecord{
			Path:    "style.css",
			Hash:    "oldhash123",
			Size:    100,
			ModTime: time.Now(),
		}
		if err := helper.manager.SaveFile(record); err != nil {
			t.Fatalf("SaveFile() error = %v", err)
		}

		shouldOptimize, err := helper.ShouldOptimizeFile("style.css")
		if err != nil {
			t.Fatalf("ShouldOptimizeFile() error = %v", err)
		}
		if !shouldOptimize {
			t.Error("modified file should need optimization")
		}
	})
}

func TestDeploymentPlanFields(t *testing.T) {
	tests := []struct {
		name          string
		plan          DeploymentPlan
		wantFiles     int
		wantSize      int64
		wantChanged   int64
		wantIncrement bool
	}{
		{
			name: "first deployment",
			plan: DeploymentPlan{
				ChangeSet: &ChangeSet{
					Added:     []string{"a.html", "b.css"},
					Modified:  []string{},
					Deleted:   []string{},
					Unchanged: []string{},
				},
				TotalFiles:    2,
				TotalSize:     2048,
				ChangedSize:   2048,
				IsIncremental: false,
			},
			wantFiles:     2,
			wantSize:      2048,
			wantChanged:   2048,
			wantIncrement: false,
		},
		{
			name: "incremental deployment",
			plan: DeploymentPlan{
				ChangeSet: &ChangeSet{
					Added:     []string{"c.js"},
					Modified:  []string{"a.html"},
					Deleted:   []string{},
					Unchanged: []string{"b.css"},
				},
				TotalFiles:    3,
				TotalSize:     4096,
				ChangedSize:   1024,
				IsIncremental: true,
			},
			wantFiles:     3,
			wantSize:      4096,
			wantChanged:   1024,
			wantIncrement: true,
		},
		{
			name: "empty deployment",
			plan: DeploymentPlan{
				ChangeSet: &ChangeSet{
					Added:     []string{},
					Modified:  []string{},
					Deleted:   []string{},
					Unchanged: []string{},
				},
				TotalFiles:    0,
				TotalSize:     0,
				ChangedSize:   0,
				IsIncremental: false,
			},
			wantFiles:     0,
			wantSize:      0,
			wantChanged:   0,
			wantIncrement: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.plan.TotalFiles != tt.wantFiles {
				t.Errorf("TotalFiles = %d, want %d", tt.plan.TotalFiles, tt.wantFiles)
			}
			if tt.plan.TotalSize != tt.wantSize {
				t.Errorf("TotalSize = %d, want %d", tt.plan.TotalSize, tt.wantSize)
			}
			if tt.plan.ChangedSize != tt.wantChanged {
				t.Errorf("ChangedSize = %d, want %d", tt.plan.ChangedSize, tt.wantChanged)
			}
			if tt.plan.IsIncremental != tt.wantIncrement {
				t.Errorf("IsIncremental = %v, want %v", tt.plan.IsIncremental, tt.wantIncrement)
			}
		})
	}
}

func TestPrepareDeploymentWithDeletedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}

	// Create initial files
	if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html>hello</html>"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(buildDir, "about.html"), []byte("<html>about</html>"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	helper, err := NewDeployHelper(tmpDir)
	if err != nil {
		t.Fatalf("NewDeployHelper() error = %v", err)
	}
	defer helper.Close()

	// Finalize first deployment
	if err := helper.FinalizeDeployment(buildDir, "proj-1", "deploy-1", nil); err != nil {
		t.Fatalf("FinalizeDeployment() error = %v", err)
	}

	// Delete a file
	if err := os.Remove(filepath.Join(buildDir, "about.html")); err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}

	plan, err := helper.PrepareDeployment(buildDir)
	if err != nil {
		t.Fatalf("PrepareDeployment() error = %v", err)
	}

	if len(plan.ChangeSet.Deleted) != 1 {
		t.Errorf("Deleted = %d, want 1", len(plan.ChangeSet.Deleted))
	}

	if len(plan.ChangeSet.Unchanged) != 1 {
		t.Errorf("Unchanged = %d, want 1", len(plan.ChangeSet.Unchanged))
	}

	if plan.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want 1", plan.TotalFiles)
	}
}

func TestFinalizeDeploymentPreservesHashes(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}

	content := []byte("<html>consistent content</html>")
	if err := os.WriteFile(filepath.Join(buildDir, "index.html"), content, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	helper, err := NewDeployHelper(tmpDir)
	if err != nil {
		t.Fatalf("NewDeployHelper() error = %v", err)
	}
	defer helper.Close()

	// Compute expected hash
	expectedHash, err := HashFile(filepath.Join(buildDir, "index.html"))
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}

	// Finalize deployment
	if err := helper.FinalizeDeployment(buildDir, "proj-1", "deploy-1", nil); err != nil {
		t.Fatalf("FinalizeDeployment() error = %v", err)
	}

	// Verify hash in manifest matches
	manifest, err := helper.GetLastDeployment()
	if err != nil {
		t.Fatalf("GetLastDeployment() error = %v", err)
	}

	record, ok := manifest.Files["index.html"]
	if !ok {
		t.Fatal("index.html not found in manifest")
	}

	if record.Hash != expectedHash {
		t.Errorf("Hash = %q, want %q", record.Hash, expectedHash)
	}
}

func TestPrepareDeploymentChangedSizeAccuracy(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}

	smallContent := []byte("small")
	largeContent := []byte("this is a much larger piece of content for testing size calculations")

	if err := os.WriteFile(filepath.Join(buildDir, "small.txt"), smallContent, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(buildDir, "large.txt"), largeContent, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	helper, err := NewDeployHelper(tmpDir)
	if err != nil {
		t.Fatalf("NewDeployHelper() error = %v", err)
	}
	defer helper.Close()

	// Finalize first deployment
	if err := helper.FinalizeDeployment(buildDir, "proj-1", "deploy-1", nil); err != nil {
		t.Fatalf("FinalizeDeployment() error = %v", err)
	}

	// Modify only the small file
	newSmallContent := []byte("modified small")
	if err := os.WriteFile(filepath.Join(buildDir, "small.txt"), newSmallContent, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	plan, err := helper.PrepareDeployment(buildDir)
	if err != nil {
		t.Fatalf("PrepareDeployment() error = %v", err)
	}

	// ChangedSize should be approximately the size of the modified file
	expectedChangedSize := int64(len(newSmallContent))
	if plan.ChangedSize != expectedChangedSize {
		t.Errorf("ChangedSize = %d, want %d", plan.ChangedSize, expectedChangedSize)
	}

	expectedTotalSize := int64(len(newSmallContent)) + int64(len(largeContent))
	if plan.TotalSize != expectedTotalSize {
		t.Errorf("TotalSize = %d, want %d", plan.TotalSize, expectedTotalSize)
	}
}

func TestMultipleDeploymentCycles(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}

	helper, err := NewDeployHelper(tmpDir)
	if err != nil {
		t.Fatalf("NewDeployHelper() error = %v", err)
	}
	defer helper.Close()

	// Cycle 1: Deploy initial content
	if err := os.WriteFile(filepath.Join(buildDir, "v1.txt"), []byte("version 1"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	plan1, err := helper.PrepareDeployment(buildDir)
	if err != nil {
		t.Fatalf("PrepareDeployment() cycle 1 error = %v", err)
	}
	if len(plan1.ChangeSet.Added) != 1 {
		t.Errorf("cycle 1: Added = %d, want 1", len(plan1.ChangeSet.Added))
	}

	if err := helper.FinalizeDeployment(buildDir, "proj", "deploy-1", nil); err != nil {
		t.Fatalf("FinalizeDeployment() cycle 1 error = %v", err)
	}

	// Cycle 2: Add another file
	if err := os.WriteFile(filepath.Join(buildDir, "v2.txt"), []byte("version 2"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	plan2, err := helper.PrepareDeployment(buildDir)
	if err != nil {
		t.Fatalf("PrepareDeployment() cycle 2 error = %v", err)
	}
	if len(plan2.ChangeSet.Added) != 1 {
		t.Errorf("cycle 2: Added = %d, want 1", len(plan2.ChangeSet.Added))
	}
	if len(plan2.ChangeSet.Unchanged) != 1 {
		t.Errorf("cycle 2: Unchanged = %d, want 1", len(plan2.ChangeSet.Unchanged))
	}

	if err := helper.FinalizeDeployment(buildDir, "proj", "deploy-2", nil); err != nil {
		t.Fatalf("FinalizeDeployment() cycle 2 error = %v", err)
	}

	// Cycle 3: No changes
	plan3, err := helper.PrepareDeployment(buildDir)
	if err != nil {
		t.Fatalf("PrepareDeployment() cycle 3 error = %v", err)
	}
	if len(plan3.ChangeSet.Unchanged) != 2 {
		t.Errorf("cycle 3: Unchanged = %d, want 2", len(plan3.ChangeSet.Unchanged))
	}
	if len(plan3.ChangeSet.Added) != 0 {
		t.Errorf("cycle 3: Added = %d, want 0", len(plan3.ChangeSet.Added))
	}

	// Verify latest manifest reflects the last finalize
	manifest, err := helper.GetLastDeployment()
	if err != nil {
		t.Fatalf("GetLastDeployment() error = %v", err)
	}
	if manifest.LastDeployID != "deploy-2" {
		t.Errorf("LastDeployID = %q, want %q", manifest.LastDeployID, "deploy-2")
	}
}

func TestPrepareDeploymentWithSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	subDir := filepath.Join(buildDir, "css")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "style.css"), []byte("body{}"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	helper, err := NewDeployHelper(tmpDir)
	if err != nil {
		t.Fatalf("NewDeployHelper() error = %v", err)
	}
	defer helper.Close()

	plan, err := helper.PrepareDeployment(buildDir)
	if err != nil {
		t.Fatalf("PrepareDeployment() error = %v", err)
	}

	if plan.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2 (files in subdirectories should be counted)", plan.TotalFiles)
	}
}
