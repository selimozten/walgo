// manager_test.go - Tests for the projects database manager
//
// These tests use a real SQLite database (in temp dir) to verify actual behavior.
// No mocks, no fakes - just real database operations.

package projects

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestNewManagerCreatesDatabase verifies that NewManager actually creates a database
func TestNewManagerCreatesDatabase(t *testing.T) {
	// Override home directory to use temp dir
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	// Verify database file was created
	dbPath := filepath.Join(tempDir, ProjectsDir, ProjectsDBName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}
}

// TestCRUDOperations tests the full create/read/update/delete cycle
func TestCRUDOperations(t *testing.T) {
	manager := setupTestManager(t)
	defer manager.Close()

	// CREATE
	project := &Project{
		Name:       "test-project",
		Category:   "blog",
		Network:    "testnet",
		ObjectID:   "0x1234567890abcdef",
		WalletAddr: "0xwallet123",
		Epochs:     5,
		SitePath:   "/tmp/test-site",
		GasFee:     "1000000",
	}

	err := manager.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	if project.ID == 0 {
		t.Error("CreateProject should set ID, got 0")
	}

	// READ
	retrieved, err := manager.GetProject(project.ID)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}

	if retrieved.Name != project.Name {
		t.Errorf("Name mismatch: got %q, want %q", retrieved.Name, project.Name)
	}
	if retrieved.ObjectID != project.ObjectID {
		t.Errorf("ObjectID mismatch: got %q, want %q", retrieved.ObjectID, project.ObjectID)
	}
	if retrieved.Status != "active" {
		t.Errorf("Status should be 'active' by default, got %q", retrieved.Status)
	}

	// UPDATE
	project.Name = "updated-project"
	project.ObjectID = "0xnewobjectid"
	err = manager.UpdateProject(project)
	if err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}

	retrieved, err = manager.GetProject(project.ID)
	if err != nil {
		t.Fatalf("GetProject after update failed: %v", err)
	}
	if retrieved.Name != "updated-project" {
		t.Errorf("Update didn't persist: Name is %q", retrieved.Name)
	}

	// DELETE
	err = manager.DeleteProject(project.ID)
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	_, err = manager.GetProject(project.ID)
	if err == nil {
		t.Error("GetProject should fail after delete")
	}
}

// TestProjectNameExists verifies duplicate name detection
func TestProjectNameExists(t *testing.T) {
	manager := setupTestManager(t)
	defer manager.Close()

	project := &Project{
		Name:       "unique-name",
		Network:    "testnet",
		ObjectID:   "0x123",
		WalletAddr: "0xwallet",
		Epochs:     1,
		SitePath:   "/tmp/site",
	}

	// Should not exist before creation
	exists, err := manager.ProjectNameExists("unique-name")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("Project should not exist before creation")
	}

	// Create project
	if err := manager.CreateProject(project); err != nil {
		t.Fatal(err)
	}

	// Should exist after creation
	exists, err = manager.ProjectNameExists("unique-name")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("Project should exist after creation")
	}

	// Different name should not exist
	exists, err = manager.ProjectNameExists("different-name")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("Different name should not exist")
	}
}

// TestDeploymentRecording verifies deployment history is tracked correctly
func TestDeploymentRecording(t *testing.T) {
	manager := setupTestManager(t)
	defer manager.Close()

	// Create a project first
	project := &Project{
		Name:       "deploy-test",
		Network:    "testnet",
		ObjectID:   "0xinitial",
		WalletAddr: "0xwallet",
		Epochs:     1,
		SitePath:   "/tmp/site",
	}
	if err := manager.CreateProject(project); err != nil {
		t.Fatal(err)
	}

	initialDeployCount := project.DeployCount

	// Record a deployment
	deployment := &DeploymentRecord{
		ProjectID: project.ID,
		ObjectID:  "0xnewversion",
		Network:   "testnet",
		Epochs:    1,
		GasFee:    "500000",
		Success:   true,
	}

	err := manager.RecordDeployment(deployment)
	if err != nil {
		t.Fatalf("RecordDeployment failed: %v", err)
	}

	if deployment.ID == 0 {
		t.Error("Deployment ID should be set after recording")
	}

	// Verify project was updated
	updated, err := manager.GetProject(project.ID)
	if err != nil {
		t.Fatal(err)
	}

	if updated.DeployCount != initialDeployCount+1 {
		t.Errorf("DeployCount should be %d, got %d", initialDeployCount+1, updated.DeployCount)
	}

	if updated.ObjectID != "0xnewversion" {
		t.Errorf("ObjectID should be updated to new version, got %q", updated.ObjectID)
	}

	// Verify deployment history
	deployments, err := manager.GetProjectDeployments(project.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(deployments) != 1 {
		t.Errorf("Expected 1 deployment, got %d", len(deployments))
	}

	if deployments[0].ObjectID != "0xnewversion" {
		t.Errorf("Deployment ObjectID mismatch: got %q", deployments[0].ObjectID)
	}
}

// TestListProjectsFiltering verifies network and status filters work
func TestListProjectsFiltering(t *testing.T) {
	manager := setupTestManager(t)
	defer manager.Close()

	// Create projects with different networks/statuses
	projects := []*Project{
		{Name: "testnet-active", Network: "testnet", Status: "active", ObjectID: "0x1", WalletAddr: "0xw", Epochs: 1, SitePath: "/tmp/1"},
		{Name: "mainnet-active", Network: "mainnet", Status: "active", ObjectID: "0x2", WalletAddr: "0xw", Epochs: 1, SitePath: "/tmp/2"},
		{Name: "testnet-archived", Network: "testnet", Status: "active", ObjectID: "0x3", WalletAddr: "0xw", Epochs: 1, SitePath: "/tmp/3"},
	}

	for _, p := range projects {
		if err := manager.CreateProject(p); err != nil {
			t.Fatal(err)
		}
	}

	// Archive one
	manager.ArchiveProject(projects[2].ID)

	// Test network filter
	testnetProjects, err := manager.ListProjects("testnet", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(testnetProjects) != 2 {
		t.Errorf("Expected 2 testnet projects, got %d", len(testnetProjects))
	}

	// Test status filter
	activeProjects, err := manager.ListProjects("", "active")
	if err != nil {
		t.Fatal(err)
	}
	if len(activeProjects) != 2 {
		t.Errorf("Expected 2 active projects, got %d", len(activeProjects))
	}

	// Test combined filter
	testnetActive, err := manager.ListProjects("testnet", "active")
	if err != nil {
		t.Fatal(err)
	}
	if len(testnetActive) != 1 {
		t.Errorf("Expected 1 testnet+active project, got %d", len(testnetActive))
	}
}

// TestArchiveAndRestore verifies archive/restore status changes
func TestArchiveAndRestore(t *testing.T) {
	manager := setupTestManager(t)
	defer manager.Close()

	project := &Project{
		Name:       "archive-test",
		Network:    "testnet",
		ObjectID:   "0x123",
		WalletAddr: "0xwallet",
		Epochs:     1,
		SitePath:   "/tmp/site",
	}

	if err := manager.CreateProject(project); err != nil {
		t.Fatal(err)
	}

	// Should be active initially
	retrieved, _ := manager.GetProject(project.ID)
	if retrieved.Status != "active" {
		t.Errorf("Initial status should be 'active', got %q", retrieved.Status)
	}

	// Archive
	if err := manager.ArchiveProject(project.ID); err != nil {
		t.Fatal(err)
	}

	retrieved, _ = manager.GetProject(project.ID)
	if retrieved.Status != "archived" {
		t.Errorf("Status after archive should be 'archived', got %q", retrieved.Status)
	}

	// Restore
	if err := manager.RestoreProject(project.ID); err != nil {
		t.Fatal(err)
	}

	retrieved, _ = manager.GetProject(project.ID)
	if retrieved.Status != "active" {
		t.Errorf("Status after restore should be 'active', got %q", retrieved.Status)
	}
}

// TestConcurrentAccess verifies the database handles concurrent operations
// This is CRITICAL for a CLI tool that might have multiple instances running
func TestConcurrentAccess(t *testing.T) {
	manager := setupTestManager(t)
	defer manager.Close()

	// Create a project to work with
	project := &Project{
		Name:       "concurrent-test",
		Network:    "testnet",
		ObjectID:   "0xinitial",
		WalletAddr: "0xwallet",
		Epochs:     1,
		SitePath:   "/tmp/site",
	}
	if err := manager.CreateProject(project); err != nil {
		t.Fatal(err)
	}

	// Simulate concurrent deployments
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			deployment := &DeploymentRecord{
				ProjectID: project.ID,
				ObjectID:  "0xversion" + string(rune('0'+idx)),
				Network:   "testnet",
				Epochs:    1,
				Success:   true,
			}
			if err := manager.RecordDeployment(deployment); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent deployment error: %v", err)
	}

	// Verify all deployments were recorded
	deployments, err := manager.GetProjectDeployments(project.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(deployments) != 10 {
		t.Errorf("Expected 10 deployments, got %d (possible race condition)", len(deployments))
	}
}

// TestProjectStats verifies statistics calculation
func TestProjectStats(t *testing.T) {
	manager := setupTestManager(t)
	defer manager.Close()

	project := &Project{
		Name:       "stats-test",
		Network:    "testnet",
		ObjectID:   "0x123",
		WalletAddr: "0xwallet",
		Epochs:     1,
		SitePath:   "/tmp/site",
	}
	if err := manager.CreateProject(project); err != nil {
		t.Fatal(err)
	}

	// Record some deployments with mixed success
	for i := 0; i < 5; i++ {
		deployment := &DeploymentRecord{
			ProjectID: project.ID,
			ObjectID:  "0x" + string(rune('a'+i)),
			Network:   "testnet",
			Epochs:    1,
			Success:   i < 3, // First 3 succeed, last 2 fail
		}
		if err := manager.RecordDeployment(deployment); err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	stats, err := manager.GetProjectStats(project.ID)
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalDeployments != 5 {
		t.Errorf("TotalDeployments should be 5, got %d", stats.TotalDeployments)
	}

	if stats.SuccessfulDeploys != 3 {
		t.Errorf("SuccessfulDeploys should be 3, got %d", stats.SuccessfulDeploys)
	}

	if stats.FailedDeploys != 2 {
		t.Errorf("FailedDeploys should be 2, got %d", stats.FailedDeploys)
	}
}

// TestSchemaMigration verifies database migrations work correctly
func TestSchemaMigration(t *testing.T) {
	manager := setupTestManager(t)

	// Create a project with all v2 fields
	project := &Project{
		Name:        "migration-test",
		Network:     "testnet",
		ObjectID:    "0x123",
		WalletAddr:  "0xwallet",
		Epochs:      1,
		SitePath:    "/tmp/site",
		Description: "Test description",
		ImageURL:    "https://example.com/image.png",
	}

	if err := manager.CreateProject(project); err != nil {
		t.Fatal(err)
	}

	// Close and reopen to verify migration persists
	manager.Close()

	manager2, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to reopen manager: %v", err)
	}
	defer manager2.Close()

	// Verify v2 fields are preserved
	retrieved, err := manager2.GetProject(project.ID)
	if err != nil {
		t.Fatalf("GetProject after reopen failed: %v", err)
	}

	if retrieved.Description != "Test description" {
		t.Errorf("Description not preserved: got %q", retrieved.Description)
	}
	if retrieved.ImageURL != "https://example.com/image.png" {
		t.Errorf("ImageURL not preserved: got %q", retrieved.ImageURL)
	}
}

// TestDeleteCascadesDeployments verifies deleting a project removes its deployments
func TestDeleteCascadesDeployments(t *testing.T) {
	manager := setupTestManager(t)
	defer manager.Close()

	project := &Project{
		Name:       "cascade-test",
		Network:    "testnet",
		ObjectID:   "0x123",
		WalletAddr: "0xwallet",
		Epochs:     1,
		SitePath:   "/tmp/site",
	}
	if err := manager.CreateProject(project); err != nil {
		t.Fatal(err)
	}

	// Add some deployments
	for i := 0; i < 3; i++ {
		deployment := &DeploymentRecord{
			ProjectID: project.ID,
			ObjectID:  "0x" + string(rune('a'+i)),
			Network:   "testnet",
			Epochs:    1,
			Success:   true,
		}
		if err := manager.RecordDeployment(deployment); err != nil {
			t.Fatal(err)
		}
	}

	// Verify deployments exist
	deployments, _ := manager.GetProjectDeployments(project.ID)
	if len(deployments) != 3 {
		t.Fatalf("Expected 3 deployments, got %d", len(deployments))
	}

	// Delete project
	if err := manager.DeleteProject(project.ID); err != nil {
		t.Fatal(err)
	}

	// Verify deployments are also gone (need to query directly)
	// This is a real-world concern: orphaned deployment records
	deployments, err := manager.GetProjectDeployments(project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(deployments) != 0 {
		t.Errorf("Deployments should be deleted with project, got %d remaining", len(deployments))
	}
}

// setupTestManager creates a manager with a temp database for testing
func setupTestManager(t *testing.T) *Manager {
	t.Helper()

	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	return manager
}
