package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/selimozten/walgo/internal/ui"
)

// DeployHelper assists with incremental deployments
type DeployHelper struct {
	manager  *Manager
	siteRoot string
}

// NewDeployHelper creates a new deployment helper
func NewDeployHelper(siteRoot string) (*DeployHelper, error) {
	manager, err := NewManager(siteRoot)
	if err != nil {
		return nil, err
	}

	return &DeployHelper{
		manager:  manager,
		siteRoot: siteRoot,
	}, nil
}

// Close closes the deployment helper
func (h *DeployHelper) Close() error {
	return h.manager.Close()
}

// PrepareDeployment prepares a directory for deployment by computing changes
func (h *DeployHelper) PrepareDeployment(buildDir string) (*DeploymentPlan, error) {
	// Compute changeset
	changes, err := h.manager.ComputeChangeSet(buildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to compute changes: %w", err)
	}

	// Get all files in the build directory
	hashes, err := HashDirectory(buildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to hash directory: %w", err)
	}

	// Count total size
	var totalSize int64
	var changedSize int64

	for path := range hashes {
		fullPath := filepath.Join(buildDir, path)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		totalSize += info.Size()

		// Check if file changed
		isChanged := false
		for _, p := range changes.Added {
			if p == path {
				isChanged = true
				break
			}
		}
		if !isChanged {
			for _, p := range changes.Modified {
				if p == path {
					isChanged = true
					break
				}
			}
		}

		if isChanged {
			changedSize += info.Size()
		}
	}

	plan := &DeploymentPlan{
		ChangeSet:     changes,
		TotalFiles:    len(hashes),
		TotalSize:     totalSize,
		ChangedSize:   changedSize,
		IsIncremental: len(changes.Unchanged) > 0,
	}

	return plan, nil
}

// FinalizeDeployment updates the cache after a successful deployment
func (h *DeployHelper) FinalizeDeployment(buildDir, projectID, deployID string, fileToBlobID map[string]string) error {
	// Compute hashes for all files
	hashes, err := HashDirectory(buildDir)
	if err != nil {
		return fmt.Errorf("failed to hash directory: %w", err)
	}

	// Create manifest
	manifest := &BuildManifest{
		SiteRoot:     h.siteRoot,
		BuildTime:    time.Now(),
		ProjectID:    projectID,
		LastDeployID: deployID,
		Files:        make(map[string]FileRecord),
	}

	// Build file records
	for path, hash := range hashes {
		fullPath := filepath.Join(buildDir, path)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		record := FileRecord{
			Path:         path,
			Hash:         hash,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			LastDeployed: time.Now(),
		}

		// Set blob ID if available
		if blobID, ok := fileToBlobID[path]; ok {
			record.BlobID = blobID
		}

		manifest.Files[path] = record
	}

	// Save manifest
	if err := h.manager.SaveManifest(manifest); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	return nil
}

// GetLastDeployment retrieves information about the last deployment
func (h *DeployHelper) GetLastDeployment() (*BuildManifest, error) {
	return h.manager.GetLatestManifest()
}

// ShouldOptimizeFile checks if a file should be re-optimized
func (h *DeployHelper) ShouldOptimizeFile(path string) (bool, error) {
	record, err := h.manager.GetFile(path)
	if err != nil {
		return true, err // Re-optimize on error
	}

	if record == nil {
		return true, nil // New file, needs optimization
	}

	// Check if file was modified
	fullPath := filepath.Join(h.siteRoot, path)
	_, err = os.Stat(fullPath)
	if err != nil {
		return true, err
	}

	// Compute current hash
	currentHash, err := HashFile(fullPath)
	if err != nil {
		return true, err
	}

	// If hash matches, skip optimization
	return currentHash != record.Hash, nil
}

// DeploymentPlan contains information about what will be deployed
type DeploymentPlan struct {
	ChangeSet     *ChangeSet
	TotalFiles    int
	TotalSize     int64
	ChangedSize   int64
	IsIncremental bool
}

// PrintSummary prints a human-readable deployment plan
func (p *DeploymentPlan) PrintSummary() {
	icons := ui.GetIcons()
	fmt.Printf("\n%s Deployment Plan:\n", icons.Chart)
	fmt.Printf("  Total files: %d (%.2f MB)\n", p.TotalFiles, float64(p.TotalSize)/(1024*1024))

	if p.IsIncremental {
		fmt.Printf("\n  %s Incremental deployment:\n", icons.Sparkles)
		fmt.Printf("    %s Added: %d files\n", icons.Pencil, len(p.ChangeSet.Added))
		fmt.Printf("    %s Modified: %d files\n", icons.Spinner, len(p.ChangeSet.Modified))
		fmt.Printf("    %s Deleted: %d files\n", icons.Garbage, len(p.ChangeSet.Deleted))
		fmt.Printf("    %s Unchanged: %d files (%.2f MB)\n", icons.Success,
			len(p.ChangeSet.Unchanged),
			float64(p.TotalSize-p.ChangedSize)/(1024*1024))

		if p.TotalSize > 0 {
			percentSaved := float64(p.TotalSize-p.ChangedSize) / float64(p.TotalSize) * 100
			fmt.Printf("\n  %s Space saved: %.1f%%\n", icons.Database, percentSaved)
		}
	} else {
		fmt.Printf("  %s First deployment - all files are new\n", icons.Rocket)
	}
}

// PrintVerboseSummary prints detailed file-by-file changes
func (p *DeploymentPlan) PrintVerboseSummary() {
	p.PrintSummary()

	icons := ui.GetIcons()
	if len(p.ChangeSet.Added) > 0 {
		fmt.Printf("\n  %s Added files:\n", icons.Pencil)
		for i, file := range p.ChangeSet.Added {
			if i < 10 { // Limit to first 10
				fmt.Printf("    + %s\n", file)
			}
		}
		if len(p.ChangeSet.Added) > 10 {
			fmt.Printf("    ... and %d more\n", len(p.ChangeSet.Added)-10)
		}
	}

	if len(p.ChangeSet.Modified) > 0 {
		fmt.Printf("\n  %s Modified files:\n", icons.Spinner)
		for i, file := range p.ChangeSet.Modified {
			if i < 10 {
				fmt.Printf("    %s %s\n", icons.Arrow, file)
			}
		}
		if len(p.ChangeSet.Modified) > 10 {
			fmt.Printf("    ... and %d more\n", len(p.ChangeSet.Modified)-10)
		}
	}

	if len(p.ChangeSet.Deleted) > 0 {
		fmt.Printf("\n  %s Deleted files:\n", icons.Garbage)
		for i, file := range p.ChangeSet.Deleted {
			if i < 10 {
				fmt.Printf("    - %s\n", file)
			}
		}
		if len(p.ChangeSet.Deleted) > 10 {
			fmt.Printf("    ... and %d more\n", len(p.ChangeSet.Deleted)-10)
		}
	}
}
