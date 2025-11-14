package cache

import "time"

// FileRecord represents a cached file entry with its metadata
type FileRecord struct {
	Path         string    // Relative path from the site root
	Hash         string    // SHA-256 hash of the file content
	Size         int64     // File size in bytes
	ModTime      time.Time // Last modification time
	BlobID       string    // Walrus blob ID (if uploaded)
	LastDeployed time.Time // When this file was last deployed
}

// BuildManifest represents a complete build cache
type BuildManifest struct {
	SiteRoot     string                 // Root directory of the site
	BuildTime    time.Time              // When this manifest was created
	Files        map[string]FileRecord  // Map of path -> FileRecord
	ProjectID    string                 // Walrus project ID
	LastDeployID string                 // Last deployment object ID
}

// ChangeSet represents the differences between two manifests
type ChangeSet struct {
	Added    []string // Files that were added
	Modified []string // Files that were modified
	Deleted  []string // Files that were deleted
	Unchanged []string // Files that haven't changed
}

// CacheStats provides statistics about cache usage
type CacheStats struct {
	TotalFiles     int
	CachedFiles    int
	ChangedFiles   int
	BytesSaved     int64
	LastBuildTime  time.Time
	CacheHitRatio  float64
}
