package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

const (
	// CacheDir is the directory where cache files are stored
	CacheDir = ".walgo"
	// CacheDBName is the name of the SQLite database file
	CacheDBName = "cache.db"
)

// Manager handles the cache database operations
type Manager struct {
	db       *sql.DB
	siteRoot string
}

// NewManager creates a new cache manager
func NewManager(siteRoot string) (*Manager, error) {
	cacheDir := filepath.Join(siteRoot, CacheDir)

	// Create cache directory if it doesn't exist
	// #nosec G301 - cache directory needs standard permissions
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Open database
	dbPath := filepath.Join(cacheDir, CacheDBName)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache database: %w", err)
	}

	manager := &Manager{
		db:       db,
		siteRoot: siteRoot,
	}

	// Initialize database schema
	if err := manager.initSchema(); err != nil {
		_ = db.Close() // Ignore close error when init already failed
		return nil, err
	}

	return manager, nil
}

// Close closes the cache database
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// initSchema creates the database tables if they don't exist
func (m *Manager) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		path TEXT PRIMARY KEY,
		hash TEXT NOT NULL,
		size INTEGER NOT NULL,
		mod_time DATETIME NOT NULL,
		blob_id TEXT,
		last_deployed DATETIME
	);

	CREATE TABLE IF NOT EXISTS manifests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		site_root TEXT NOT NULL,
		build_time DATETIME NOT NULL,
		project_id TEXT,
		deploy_id TEXT,
		manifest_data TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_files_hash ON files(hash);
	CREATE INDEX IF NOT EXISTS idx_manifests_build_time ON manifests(build_time);
	`

	_, err := m.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// SaveFile stores or updates a file record in the cache
func (m *Manager) SaveFile(record FileRecord) error {
	query := `
	INSERT INTO files (path, hash, size, mod_time, blob_id, last_deployed)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		hash = excluded.hash,
		size = excluded.size,
		mod_time = excluded.mod_time,
		blob_id = excluded.blob_id,
		last_deployed = excluded.last_deployed
	`

	_, err := m.db.Exec(query,
		record.Path,
		record.Hash,
		record.Size,
		record.ModTime,
		record.BlobID,
		record.LastDeployed,
	)

	if err != nil {
		return fmt.Errorf("failed to save file record: %w", err)
	}

	return nil
}

// GetFile retrieves a file record from the cache
func (m *Manager) GetFile(path string) (*FileRecord, error) {
	query := `
	SELECT path, hash, size, mod_time, blob_id, last_deployed
	FROM files
	WHERE path = ?
	`

	var record FileRecord
	var lastDeployed sql.NullTime
	var blobID sql.NullString

	err := m.db.QueryRow(query, path).Scan(
		&record.Path,
		&record.Hash,
		&record.Size,
		&record.ModTime,
		&blobID,
		&lastDeployed,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get file record: %w", err)
	}

	if blobID.Valid {
		record.BlobID = blobID.String
	}

	if lastDeployed.Valid {
		record.LastDeployed = lastDeployed.Time
	}

	return &record, nil
}

// GetAllFiles retrieves all file records from the cache
func (m *Manager) GetAllFiles() (map[string]FileRecord, error) {
	query := `SELECT path, hash, size, mod_time, blob_id, last_deployed FROM files`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	files := make(map[string]FileRecord)

	for rows.Next() {
		var record FileRecord
		var lastDeployed sql.NullTime
		var blobID sql.NullString

		err := rows.Scan(
			&record.Path,
			&record.Hash,
			&record.Size,
			&record.ModTime,
			&blobID,
			&lastDeployed,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan file record: %w", err)
		}

		if blobID.Valid {
			record.BlobID = blobID.String
		}

		if lastDeployed.Valid {
			record.LastDeployed = lastDeployed.Time
		}

		files[record.Path] = record
	}

	return files, nil
}

// SaveManifest stores a build manifest
func (m *Manager) SaveManifest(manifest *BuildManifest) error {
	// Serialize manifest to JSON
	manifestData, err := json.Marshal(manifest.Files)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	query := `
	INSERT INTO manifests (site_root, build_time, project_id, deploy_id, manifest_data)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err = m.db.Exec(query,
		manifest.SiteRoot,
		manifest.BuildTime,
		manifest.ProjectID,
		manifest.LastDeployID,
		string(manifestData),
	)

	if err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	// Update individual file records
	for _, file := range manifest.Files {
		if err := m.SaveFile(file); err != nil {
			return err
		}
	}

	return nil
}

// GetLatestManifest retrieves the most recent build manifest
func (m *Manager) GetLatestManifest() (*BuildManifest, error) {
	query := `
	SELECT site_root, build_time, project_id, deploy_id, manifest_data
	FROM manifests
	WHERE site_root = ?
	ORDER BY build_time DESC
	LIMIT 1
	`

	var manifest BuildManifest
	var manifestData string
	var projectID, deployID sql.NullString

	err := m.db.QueryRow(query, m.siteRoot).Scan(
		&manifest.SiteRoot,
		&manifest.BuildTime,
		&projectID,
		&deployID,
		&manifestData,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get latest manifest: %w", err)
	}

	if projectID.Valid {
		manifest.ProjectID = projectID.String
	}

	if deployID.Valid {
		manifest.LastDeployID = deployID.String
	}

	// Deserialize manifest data
	manifest.Files = make(map[string]FileRecord)
	if err := json.Unmarshal([]byte(manifestData), &manifest.Files); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// ComputeChangeSet compares the current directory with the cached manifest
func (m *Manager) ComputeChangeSet(dir string) (*ChangeSet, error) {
	// Get latest manifest
	manifest, err := m.GetLatestManifest()
	if err != nil {
		return nil, err
	}

	// If no manifest exists, all files are new
	if manifest == nil {
		hashes, err := HashDirectory(dir)
		if err != nil {
			return nil, err
		}

		changes := &ChangeSet{
			Added:     make([]string, 0, len(hashes)),
			Modified:  make([]string, 0),
			Deleted:   make([]string, 0),
			Unchanged: make([]string, 0),
		}

		for path := range hashes {
			changes.Added = append(changes.Added, path)
		}

		return changes, nil
	}

	// Compute hashes for current directory
	currentHashes, err := HashDirectory(dir)
	if err != nil {
		return nil, err
	}

	// Extract old hashes from manifest
	oldHashes := make(map[string]string)
	for path, record := range manifest.Files {
		oldHashes[path] = record.Hash
	}

	// Compare hashes
	return CompareHashes(oldHashes, currentHashes), nil
}

// GetStats returns cache statistics
func (m *Manager) GetStats() (*CacheStats, error) {
	manifest, err := m.GetLatestManifest()
	if err != nil {
		return nil, err
	}

	stats := &CacheStats{}

	if manifest != nil {
		stats.TotalFiles = len(manifest.Files)
		stats.LastBuildTime = manifest.BuildTime

		// Count cached files (files with blob IDs)
		for _, file := range manifest.Files {
			if file.BlobID != "" {
				stats.CachedFiles++
			}
		}

		if stats.TotalFiles > 0 {
			stats.CacheHitRatio = float64(stats.CachedFiles) / float64(stats.TotalFiles)
		}
	}

	return stats, nil
}

// Clear removes all cache data
func (m *Manager) Clear() error {
	queries := []string{
		"DELETE FROM files",
		"DELETE FROM manifests",
	}

	for _, query := range queries {
		if _, err := m.db.Exec(query); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
	}

	return nil
}

// UpdateBlobID updates the blob ID for a file
func (m *Manager) UpdateBlobID(path, blobID string) error {
	query := `UPDATE files SET blob_id = ?, last_deployed = ? WHERE path = ?`

	_, err := m.db.Exec(query, blobID, time.Now(), path)
	if err != nil {
		return fmt.Errorf("failed to update blob ID: %w", err)
	}

	return nil
}
