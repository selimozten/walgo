package projects

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

const (
	// ProjectsDir is the directory where project database is stored
	ProjectsDir = ".walgo"
	// ProjectsDBName is the name of the SQLite database file
	ProjectsDBName = "projects.db"
)

// Manager manages all project database operations including CRUD and querying.
type Manager struct {
	db *sql.DB
}

// NewManager initializes a new project manager with a global database in the user's home directory.
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	projectsDir := filepath.Join(homeDir, ProjectsDir)

	// Create projects directory if it doesn't exist
	// #nosec G301 - projects directory needs standard permissions
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create projects directory: %w", err)
	}

	// Open database with connection parameters for concurrency support
	dbPath := filepath.Join(projectsDir, ProjectsDBName)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open projects database: %w", err)
	}

	// Set busy timeout FIRST — before any schema changes — so concurrent
	// connections wait instead of failing immediately with SQLITE_BUSY.
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Enable WAL mode for better concurrent access
	// WAL allows readers and writers to operate concurrently
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Limit connections to prevent excessive file handle usage
	db.SetMaxOpenConns(1)

	manager := &Manager{db: db}

	// Initialize database schema
	if err := manager.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return manager, nil
}

// Close terminates the database connection and releases resources.
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// schemaVersion defines the current database schema version.
// Increment this value when adding new migrations:
// Version 1: Initial schema with projects and deployments tables
// Version 2: Added description and image_url columns to projects table
const schemaVersion = 2

// initSchema creates database tables and applies pending migrations.
func (m *Manager) initSchema() error {
	// First, create the schema_version table if it doesn't exist
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	// Get current schema version from database
	dbVersion := 0
	row := m.db.QueryRow("SELECT MAX(version) FROM schema_version")
	_ = row.Scan(&dbVersion) // Ignore error - table might be empty

	// Apply migrations in order up to schemaVersion
	if dbVersion < 1 {
		if err := m.applyMigration1(); err != nil {
			return fmt.Errorf("failed to apply migration 1: %w", err)
		}
	}

	if dbVersion < 2 && schemaVersion >= 2 {
		if err := m.applyMigration2(); err != nil {
			return fmt.Errorf("failed to apply migration 2: %w", err)
		}
	}

	return nil
}

// applyMigration1 creates the initial database schema (version 1).
func (m *Manager) applyMigration1() error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		category TEXT,
		network TEXT NOT NULL,
		object_id TEXT NOT NULL,
		suins TEXT,
		wallet_addr TEXT NOT NULL,
		epochs INTEGER NOT NULL,
		gas_fee TEXT,
		site_path TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		last_deploy_at DATETIME NOT NULL,
		deploy_count INTEGER DEFAULT 1,
		status TEXT DEFAULT 'active'
	);

	CREATE TABLE IF NOT EXISTS deployments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		object_id TEXT NOT NULL,
		network TEXT NOT NULL,
		epochs INTEGER NOT NULL,
		gas_fee TEXT,
		version TEXT,
		notes TEXT,
		success INTEGER NOT NULL,
		error TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (project_id) REFERENCES projects(id)
	);

	CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
	CREATE INDEX IF NOT EXISTS idx_projects_network ON projects(network);
	CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
	CREATE INDEX IF NOT EXISTS idx_deployments_project_id ON deployments(project_id);
	CREATE INDEX IF NOT EXISTS idx_deployments_created_at ON deployments(created_at);
	`

	_, err = tx.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Record migration version (OR IGNORE for idempotency if concurrent connections race)
	_, err = tx.Exec("INSERT OR IGNORE INTO schema_version (version, applied_at) VALUES (?, ?)", 1, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration version: %w", err)
	}

	return tx.Commit()
}

// allowedTables is a whitelist of table names that can be used in PRAGMA queries.
var allowedTables = map[string]bool{
	"projects":    true,
	"deployments": true,
}

// columnExists checks if a column exists in a table
// If tx is provided, uses the transaction, otherwise uses the main db connection
func (m *Manager) columnExists(tx *sql.Tx, table, column string) bool {
	if !allowedTables[table] {
		return false
	}

	var rows *sql.Rows
	var err error

	query := fmt.Sprintf("PRAGMA table_info(%s)", table) // #nosec G201 - table name validated against whitelist
	if tx != nil {
		rows, err = tx.Query(query)
	} else {
		rows, err = m.db.Query(query)
	}

	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			continue
		}
		if name == column {
			return true
		}
	}
	return false
}

// applyMigration2 adds description and image_url columns to projects table (version 2).
func (m *Manager) applyMigration2() error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// Add description column if it doesn't exist
	if !m.columnExists(tx, "projects", "description") {
		if _, err := tx.Exec(`ALTER TABLE projects ADD COLUMN description TEXT DEFAULT ''`); err != nil {
			return fmt.Errorf("failed to add description column: %w", err)
		}
	}

	// Add image_url column if it doesn't exist
	if !m.columnExists(tx, "projects", "image_url") {
		if _, err := tx.Exec(`ALTER TABLE projects ADD COLUMN image_url TEXT DEFAULT ''`); err != nil {
			return fmt.Errorf("failed to add image_url column: %w", err)
		}
	}

	// Record migration version (OR IGNORE for idempotency if concurrent connections race)
	if _, err := tx.Exec("INSERT OR IGNORE INTO schema_version (version, applied_at) VALUES (?, ?)", 2, time.Now()); err != nil {
		return fmt.Errorf("failed to record migration version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	committed = true
	return nil
}

// CreateProject creates a new project record in the database.
func (m *Manager) CreateProject(project *Project) error {
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	project.LastDeployAt = now
	project.DeployCount = 1
	project.Status = "active"

	result, err := m.db.Exec(`
		INSERT INTO projects (name, category, network, object_id, suins, wallet_addr, epochs, gas_fee, site_path, created_at, updated_at, last_deploy_at, deploy_count, status, description, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, project.Name, project.Category, project.Network, project.ObjectID, project.SuiNS, project.WalletAddr, project.Epochs, project.GasFee, project.SitePath, project.CreatedAt, project.UpdatedAt, project.LastDeployAt, project.DeployCount, project.Status, project.Description, project.ImageURL)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get project ID: %w", err)
	}

	project.ID = id
	return nil
}

// CreateDraftProject creates a new project with draft status (not yet deployed)
// This is used for newly created sites that haven't been deployed to the blockchain yet.
func (m *Manager) CreateDraftProject(name, sitePath string) error {
	now := time.Now()
	project := &Project{
		Name:         name,
		SitePath:     sitePath,
		Status:       "draft",
		Category:     "website",
		CreatedAt:    now,
		UpdatedAt:    now,
		LastDeployAt: time.Time{}, // Zero value - never deployed
		DeployCount:  0,
	}

	result, err := m.db.Exec(`
		INSERT INTO projects (name, category, network, object_id, suins, wallet_addr, epochs, gas_fee, site_path, created_at, updated_at, last_deploy_at, deploy_count, status, description, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, project.Name, project.Category, project.Network, project.ObjectID, project.SuiNS, project.WalletAddr, project.Epochs, project.GasFee, project.SitePath, project.CreatedAt, project.UpdatedAt, project.LastDeployAt, project.DeployCount, project.Status, project.Description, project.ImageURL)

	if err != nil {
		return fmt.Errorf("failed to create draft project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get project ID: %w", err)
	}

	project.ID = id
	return nil
}

// GetProject retrieves a project record by its unique identifier.
func (m *Manager) GetProject(id int64) (*Project, error) {
	project := &Project{}

	err := m.db.QueryRow(`
		SELECT id, name, category, network, object_id, suins, wallet_addr, epochs, gas_fee, site_path, created_at, updated_at, last_deploy_at, deploy_count, status, description, image_url
		FROM projects WHERE id = ?
	`, id).Scan(&project.ID, &project.Name, &project.Category, &project.Network, &project.ObjectID, &project.SuiNS, &project.WalletAddr, &project.Epochs, &project.GasFee, &project.SitePath, &project.CreatedAt, &project.UpdatedAt, &project.LastDeployAt, &project.DeployCount, &project.Status, &project.Description, &project.ImageURL)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// GetProjectByName retrieves the most recent project record by name.
func (m *Manager) GetProjectByName(name string) (*Project, error) {
	project := &Project{}

	err := m.db.QueryRow(`
		SELECT id, name, category, network, object_id, suins, wallet_addr, epochs, gas_fee, site_path, created_at, updated_at, last_deploy_at, deploy_count, status, description, image_url
		FROM projects WHERE name = ? ORDER BY created_at DESC LIMIT 1
	`, name).Scan(&project.ID, &project.Name, &project.Category, &project.Network, &project.ObjectID, &project.SuiNS, &project.WalletAddr, &project.Epochs, &project.GasFee, &project.SitePath, &project.CreatedAt, &project.UpdatedAt, &project.LastDeployAt, &project.DeployCount, &project.Status, &project.Description, &project.ImageURL)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// ProjectNameExists checks if a project with the given name already exists
func (m *Manager) ProjectNameExists(name string) (bool, error) {
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM projects WHERE name = ?", name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check project name: %w", err)
	}
	return count > 0, nil
}

// GetProjectBySitePath retrieves a project record by its site path.
func (m *Manager) GetProjectBySitePath(sitePath string) (*Project, error) {
	project := &Project{}

	err := m.db.QueryRow(`
		SELECT id, name, category, network, object_id, suins, wallet_addr, epochs, gas_fee, site_path, created_at, updated_at, last_deploy_at, deploy_count, status, description, image_url
		FROM projects WHERE site_path = ? ORDER BY created_at DESC LIMIT 1
	`, sitePath).Scan(&project.ID, &project.Name, &project.Category, &project.Network, &project.ObjectID, &project.SuiNS, &project.WalletAddr, &project.Epochs, &project.GasFee, &project.SitePath, &project.CreatedAt, &project.UpdatedAt, &project.LastDeployAt, &project.DeployCount, &project.Status, &project.Description, &project.ImageURL)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil if not found (not an error)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project by site path: %w", err)
	}

	return project, nil
}

// ListProjects retrieves all projects with optional network and status filters.
func (m *Manager) ListProjects(network string, status string) ([]*Project, error) {
	query := `SELECT id, name, category, network, object_id, suins, wallet_addr, epochs, gas_fee, site_path, created_at, updated_at, last_deploy_at, deploy_count, status, description, image_url FROM projects WHERE 1=1`
	args := []interface{}{}

	if network != "" {
		query += " AND network = ?"
		args = append(args, network)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY last_deploy_at DESC"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(&project.ID, &project.Name, &project.Category, &project.Network, &project.ObjectID, &project.SuiNS, &project.WalletAddr, &project.Epochs, &project.GasFee, &project.SitePath, &project.CreatedAt, &project.UpdatedAt, &project.LastDeployAt, &project.DeployCount, &project.Status, &project.Description, &project.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// UpdateProject modifies an existing project record in the database.
func (m *Manager) UpdateProject(project *Project) error {
	project.UpdatedAt = time.Now()

	_, err := m.db.Exec(`
		UPDATE projects SET name = ?, category = ?, network = ?, object_id = ?, suins = ?, wallet_addr = ?, epochs = ?, gas_fee = ?, site_path = ?, updated_at = ?, last_deploy_at = ?, deploy_count = ?, status = ?, description = ?, image_url = ?
		WHERE id = ?
	`, project.Name, project.Category, project.Network, project.ObjectID, project.SuiNS, project.WalletAddr, project.Epochs, project.GasFee, project.SitePath, project.UpdatedAt, project.LastDeployAt, project.DeployCount, project.Status, project.Description, project.ImageURL, project.ID)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// ActivateProjectBySitePath activates a draft project after successful deployment
// This updates the status to "active" and records deployment information
func (m *Manager) ActivateProjectBySitePath(sitePath, objectID, network string, epochs int) error {
	// Find the project by site path
	project, err := m.GetProjectBySitePath(sitePath)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		// No project found - this is okay, not all sites are in project management
		return nil
	}

	// Update project with deployment information
	now := time.Now()
	project.Status = "active"
	project.ObjectID = objectID
	project.Network = network
	project.Epochs = epochs
	project.LastDeployAt = now
	project.DeployCount++
	project.UpdatedAt = now

	return m.UpdateProject(project)
}

// RecordDeployment creates a new deployment record for a project.
// Uses a transaction to ensure atomicity between deployment recording and project updates.
func (m *Manager) RecordDeployment(deployment *DeploymentRecord) error {
	deployment.CreatedAt = time.Now()

	// Start a transaction for atomicity
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.Exec(`
		INSERT INTO deployments (project_id, object_id, network, epochs, gas_fee, version, notes, success, error, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, deployment.ProjectID, deployment.ObjectID, deployment.Network, deployment.Epochs, deployment.GasFee, deployment.Version, deployment.Notes, deployment.Success, deployment.Error, deployment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to record deployment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get deployment ID: %w", err)
	}

	deployment.ID = id

	// Update project's last deploy time and deploy count
	_, err = tx.Exec(`
		UPDATE projects SET last_deploy_at = ?, deploy_count = deploy_count + 1, object_id = ? WHERE id = ?
	`, deployment.CreatedAt, deployment.ObjectID, deployment.ProjectID)

	if err != nil {
		return fmt.Errorf("failed to update project after deployment: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetProjectDeployments retrieves all deployment records for a specified project.
func (m *Manager) GetProjectDeployments(projectID int64) ([]*DeploymentRecord, error) {
	rows, err := m.db.Query(`
		SELECT id, project_id, object_id, network, epochs, gas_fee, version, notes, success, error, created_at
		FROM deployments WHERE project_id = ? ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*DeploymentRecord
	for rows.Next() {
		d := &DeploymentRecord{}
		var version, notes, gasErr sql.NullString
		err := rows.Scan(&d.ID, &d.ProjectID, &d.ObjectID, &d.Network, &d.Epochs, &d.GasFee, &version, &notes, &d.Success, &gasErr, &d.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}
		if version.Valid {
			d.Version = version.String
		}
		if notes.Valid {
			d.Notes = notes.String
		}
		if gasErr.Valid {
			d.Error = gasErr.String
		}
		deployments = append(deployments, d)
	}

	return deployments, nil
}

// GetProjectStats computes deployment statistics for a specified project.
func (m *Manager) GetProjectStats(projectID int64) (*ProjectStats, error) {
	stats := &ProjectStats{}

	// Get deployment counts
	err := m.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed
		FROM deployments WHERE project_id = ?
	`, projectID).Scan(&stats.TotalDeployments, &stats.SuccessfulDeploys, &stats.FailedDeploys)

	if err != nil {
		return nil, fmt.Errorf("failed to get deployment stats: %w", err)
	}

	// Get project info
	project, err := m.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	stats.CurrentNetwork = project.Network
	stats.CurrentObjectID = project.ObjectID

	// Get first and last deployment times
	if stats.TotalDeployments > 0 {
		var firstDeploy, lastDeploy time.Time
		err = m.db.QueryRow(`
			SELECT
				MIN(created_at) as first,
				MAX(created_at) as last
			FROM deployments WHERE project_id = ?
		`, projectID).Scan(&firstDeploy, &lastDeploy)

		if err == nil {
			stats.FirstDeployment = firstDeploy
			stats.LastDeployment = lastDeploy
		}
	}

	return stats, nil
}

// DeleteProject removes a project and all its deployment records.
// Uses a transaction to ensure atomic deletion of all related records.
// Also deletes the site folder from the filesystem if deleteSiteFolder is true.
func (m *Manager) DeleteProject(id int64) error {
	return m.DeleteProjectWithOptions(id, true)
}

// DeleteProjectWithOptions removes a project with options to control site folder deletion.
func (m *Manager) DeleteProjectWithOptions(id int64, deleteSiteFolder bool) error {
	// Get project details before deleting (need site path)
	project, err := m.GetProject(id)
	if err != nil {
		return fmt.Errorf("failed to get project details: %w", err)
	}

	// Start a transaction for atomicity
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Delete deployments first
	_, err = tx.Exec("DELETE FROM deployments WHERE project_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete deployments: %w", err)
	}

	// Delete project
	_, err = tx.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Delete site folder if requested and path exists
	if deleteSiteFolder && project.SitePath != "" {
		// Check if the folder exists
		if _, err := os.Stat(project.SitePath); err == nil {
			// Folder exists, delete it
			if err := os.RemoveAll(project.SitePath); err != nil {
				// Return a warning but don't fail the overall operation
				// since the database deletion already succeeded
				return fmt.Errorf("project deleted from database but failed to delete site folder at %s: %w", project.SitePath, err)
			}
		}
		// If folder doesn't exist, that's fine - it may have been manually deleted
	}

	return nil
}

// ArchiveProject marks a project record as archived in the database.
func (m *Manager) ArchiveProject(id int64) error {
	_, err := m.db.Exec("UPDATE projects SET status = 'archived', updated_at = ? WHERE id = ?", time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}
	return nil
}

// RestoreProject marks a previously archived project as active.
func (m *Manager) RestoreProject(id int64) error {
	_, err := m.db.Exec("UPDATE projects SET status = 'active', updated_at = ? WHERE id = ?", time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to restore project: %w", err)
	}
	return nil
}

// EpochInfo contains epoch and timing information for expiry calculation
type EpochInfo struct {
	TotalEpochs       int       // Sum of all successful deployment epochs
	FirstDeploymentAt time.Time // Date of first successful deployment
	LastDeploymentAt  time.Time // Date of last successful deployment
	DeploymentCount   int       // Number of successful deployments
}

// GetEpochInfo retrieves epoch information from all successful deployments for a project.
// This is used to calculate the correct storage expiry date by summing all epochs.
func (m *Manager) GetEpochInfo(projectID int64) (*EpochInfo, error) {
	info := &EpochInfo{}

	// Get total epochs from all successful deployments
	err := m.db.QueryRow(`
		SELECT
			COALESCE(SUM(epochs), 0) as total_epochs,
			COUNT(*) as deployment_count
		FROM deployments
		WHERE project_id = ? AND success = 1
	`, projectID).Scan(&info.TotalEpochs, &info.DeploymentCount)

	if err != nil {
		return nil, fmt.Errorf("failed to get epoch info: %w", err)
	}

	if info.DeploymentCount == 0 {
		return info, nil
	}

	// Get first and last deployment dates as strings (SQLite stores datetime as text)
	var firstDeployStr, lastDeployStr string
	err = m.db.QueryRow(`
		SELECT
			MIN(created_at) as first_deploy,
			MAX(created_at) as last_deploy
		FROM deployments
		WHERE project_id = ? AND success = 1
	`, projectID).Scan(&firstDeployStr, &lastDeployStr)

	if err != nil {
		return nil, fmt.Errorf("failed to get deployment dates: %w", err)
	}

	// Parse datetime strings - try multiple formats
	parseTime := func(s string) (time.Time, error) {
		if s == "" {
			return time.Time{}, fmt.Errorf("empty time string")
		}

		// Strip Go's monotonic clock reading (e.g., " m=+63.651456751")
		if idx := strings.Index(s, " m="); idx != -1 {
			s = s[:idx]
		}

		// Handle dual timezone format (e.g., "+0300 +03")
		// Keep only the first timezone offset
		parts := strings.Fields(s)
		if len(parts) >= 3 {
			// Check if last two parts look like timezone offsets
			last := parts[len(parts)-1]
			secondLast := parts[len(parts)-2]
			if len(last) <= 4 && (strings.HasPrefix(last, "+") || strings.HasPrefix(last, "-")) {
				if len(secondLast) >= 5 && (strings.HasPrefix(secondLast, "+") || strings.HasPrefix(secondLast, "-")) {
					// Remove the shorter timezone
					s = strings.Join(parts[:len(parts)-1], " ")
				}
			}
		}

		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05.999999999 -0700",
			"2006-01-02 15:04:05.999999 -0700",
			"2006-01-02 15:04:05 -0700",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, s); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to parse time %q with any known format", s)
	}

	// Parse deployment dates — non-fatal if parsing fails (use zero time)
	if t, err := parseTime(firstDeployStr); err == nil {
		info.FirstDeploymentAt = t
	}
	if t, err := parseTime(lastDeployStr); err == nil {
		info.LastDeploymentAt = t
	}

	return info, nil
}

// SetStatus sets the status of a project to the specified value.
func (m *Manager) SetStatus(id int64, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"draft":    true,
		"active":   true,
		"archived": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s (must be draft, active, or archived)", status)
	}

	_, err := m.db.Exec("UPDATE projects SET status = ?, updated_at = ? WHERE id = ?", status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to set project status: %w", err)
	}
	return nil
}
