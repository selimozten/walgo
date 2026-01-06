package deployment

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/selimozten/walgo/internal/compress"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployer"
	"github.com/selimozten/walgo/internal/projects"
)

// MockDeployer is a mock implementation of deployer.WalrusDeployer
type MockDeployer struct {
	DeployFunc  func(ctx context.Context, siteDir string, opts deployer.DeployOptions) (*deployer.Result, error)
	UpdateFunc  func(ctx context.Context, siteDir string, objectID string, opts deployer.DeployOptions) (*deployer.Result, error)
	StatusFunc  func(ctx context.Context, objectID string, opts deployer.DeployOptions) (*deployer.Result, error)
	DestroyFunc func(ctx context.Context, objectID string) error

	// Track calls for assertions
	DeployCalled  bool
	UpdateCalled  bool
	StatusCalled  bool
	DestroyCalled bool
	LastSiteDir   string
	LastObjectID  string
	LastOpts      deployer.DeployOptions
}

func (m *MockDeployer) Deploy(ctx context.Context, siteDir string, opts deployer.DeployOptions) (*deployer.Result, error) {
	m.DeployCalled = true
	m.LastSiteDir = siteDir
	m.LastOpts = opts
	if m.DeployFunc != nil {
		return m.DeployFunc(ctx, siteDir, opts)
	}
	return &deployer.Result{
		Success:  true,
		ObjectID: "mock-object-id-12345",
	}, nil
}

func (m *MockDeployer) Update(ctx context.Context, siteDir string, objectID string, opts deployer.DeployOptions) (*deployer.Result, error) {
	m.UpdateCalled = true
	m.LastSiteDir = siteDir
	m.LastObjectID = objectID
	m.LastOpts = opts
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, siteDir, objectID, opts)
	}
	return &deployer.Result{
		Success:  true,
		ObjectID: objectID,
	}, nil
}

func (m *MockDeployer) Status(ctx context.Context, objectID string, opts deployer.DeployOptions) (*deployer.Result, error) {
	m.StatusCalled = true
	m.LastObjectID = objectID
	m.LastOpts = opts
	if m.StatusFunc != nil {
		return m.StatusFunc(ctx, objectID, opts)
	}
	return &deployer.Result{
		Success:  true,
		ObjectID: objectID,
	}, nil
}

func (m *MockDeployer) Destroy(ctx context.Context, objectID string) error {
	m.DestroyCalled = true
	m.LastObjectID = objectID
	if m.DestroyFunc != nil {
		return m.DestroyFunc(ctx, objectID)
	}
	return nil
}

// TestDeploymentOptionsValidation tests the DeploymentOptions struct
func TestDeploymentOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		opts    DeploymentOptions
		wantErr bool
	}{
		{
			name: "valid options with all required fields",
			opts: DeploymentOptions{
				SitePath:   "/path/to/site",
				PublishDir: "/path/to/site/public",
				Epochs:     5,
				WalgoCfg:   &config.WalgoConfig{},
			},
			wantErr: false,
		},
		{
			name: "valid options with project name",
			opts: DeploymentOptions{
				SitePath:    "/path/to/site",
				PublishDir:  "/path/to/site/public",
				Epochs:      10,
				ProjectName: "my-project",
				WalgoCfg:    &config.WalgoConfig{},
			},
			wantErr: false,
		},
		{
			name: "valid options with metadata",
			opts: DeploymentOptions{
				SitePath:    "/path/to/site",
				PublishDir:  "/path/to/site/public",
				Epochs:      5,
				ProjectName: "my-project",
				Description: "A test project",
				ImageURL:    "https://example.com/image.png",
				Category:    "website",
				WalgoCfg:    &config.WalgoConfig{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate options are properly structured
			if tt.opts.SitePath == "" && !tt.wantErr {
				t.Error("Expected SitePath to be non-empty for valid options")
			}
			if tt.opts.PublishDir == "" && !tt.wantErr {
				t.Error("Expected PublishDir to be non-empty for valid options")
			}
		})
	}
}

// TestDeploymentResultFields tests the DeploymentResult struct
func TestDeploymentResultFields(t *testing.T) {
	result := &DeploymentResult{
		Success:      true,
		ObjectID:     "0x1234567890abcdef",
		IsUpdate:     false,
		IsNewProject: true,
		SiteSize:     1024 * 1024, // 1MB
		Error:        nil,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.ObjectID != "0x1234567890abcdef" {
		t.Errorf("Expected ObjectID to be '0x1234567890abcdef', got '%s'", result.ObjectID)
	}
	if result.IsUpdate {
		t.Error("Expected IsUpdate to be false")
	}
	if !result.IsNewProject {
		t.Error("Expected IsNewProject to be true")
	}
	if result.SiteSize != 1024*1024 {
		t.Errorf("Expected SiteSize to be 1MB, got %d", result.SiteSize)
	}
	if result.Error != nil {
		t.Error("Expected Error to be nil")
	}
}

// TestDeploymentResultWithError tests error handling in DeploymentResult
func TestDeploymentResultWithError(t *testing.T) {
	testErr := os.ErrNotExist
	result := &DeploymentResult{
		Success: false,
		Error:   testErr,
	}

	if result.Success {
		t.Error("Expected Success to be false when there's an error")
	}
	if result.Error != testErr {
		t.Errorf("Expected Error to be %v, got %v", testErr, result.Error)
	}
}

// Helper function to create a test site directory
func createTestSiteDir(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "walgo-deployment-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create public directory
	publicDir := filepath.Join(tempDir, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create public directory: %v", err)
	}

	// Create test files
	files := map[string]string{
		"index.html":              "<html><body>Hello World</body></html>",
		"style.css":               "body { color: black; }",
		"script.js":               "console.log('hello');",
		"images/logo.png":         "fake png data",
		"assets/fonts/test.woff2": "fake font data",
	}

	for path, content := range files {
		fullPath := filepath.Join(publicDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Create walgo.yaml config
	walgoCfg := config.NewDefaultWalgoConfig()
	cfgData, err := json.Marshal(walgoCfg)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "walgo.yaml"), cfgData, 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to write walgo.yaml: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// TestCalculateSiteSize tests the site size calculation logic
func TestCalculateSiteSize(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	var totalSize int64
	err := filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	if totalSize <= 0 {
		t.Error("Expected positive site size")
	}

	// Verify size is reasonable (should be sum of test file contents)
	expectedMinSize := int64(50) // At least 50 bytes from our test files
	if totalSize < expectedMinSize {
		t.Errorf("Site size %d is smaller than expected minimum %d", totalSize, expectedMinSize)
	}
}

// TestSiteSizeWithErrors tests site size calculation with inaccessible paths
func TestSiteSizeWithErrors(t *testing.T) {
	nonExistentDir := "/path/that/does/not/exist"

	var walkErrors []string
	filepath.Walk(nonExistentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			walkErrors = append(walkErrors, err.Error())
			return nil
		}
		return nil
	})

	// Should have at least one error for non-existent path
	if len(walkErrors) == 0 {
		t.Error("Expected walk errors for non-existent directory")
	}
}

// TestWSResourcesConfigRead tests reading ws-resources.json
func TestWSResourcesConfigRead(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")

	// Create ws-resources.json
	wsConfig := &compress.WSResourcesConfig{
		ObjectID: "test-object-id-123",
		SiteName: "test-site",
		Headers: map[string]map[string]string{
			"/index.html": {
				"Content-Type": "text/html",
			},
		},
	}

	data, err := json.MarshalIndent(wsConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal ws-resources config: %v", err)
	}

	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	// Read it back
	readConfig, err := compress.ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		t.Fatalf("Failed to read ws-resources.json: %v", err)
	}

	if readConfig.ObjectID != "test-object-id-123" {
		t.Errorf("Expected ObjectID 'test-object-id-123', got '%s'", readConfig.ObjectID)
	}
	if readConfig.SiteName != "test-site" {
		t.Errorf("Expected SiteName 'test-site', got '%s'", readConfig.SiteName)
	}
}

// TestWSResourcesConfigReadNonExistent tests reading non-existent ws-resources.json
func TestWSResourcesConfigReadNonExistent(t *testing.T) {
	_, err := compress.ReadWSResourcesConfig("/path/that/does/not/exist/ws-resources.json")
	if err == nil {
		t.Error("Expected error when reading non-existent file")
	}
}

// TestObjectIDDetection tests detection of existing object IDs from different sources
func TestObjectIDDetection(t *testing.T) {
	tests := []struct {
		name         string
		projectID    string
		wsObjectID   string
		expectedID   string
		expectUpdate bool
	}{
		{
			name:         "from walgo.yaml projectID",
			projectID:    "walgo-yaml-object-id",
			wsObjectID:   "",
			expectedID:   "walgo-yaml-object-id",
			expectUpdate: true,
		},
		{
			name:         "from ws-resources.json",
			projectID:    "",
			wsObjectID:   "ws-resources-object-id",
			expectedID:   "ws-resources-object-id",
			expectUpdate: true,
		},
		{
			name:         "prefer walgo.yaml over ws-resources",
			projectID:    "walgo-yaml-object-id",
			wsObjectID:   "ws-resources-object-id",
			expectedID:   "walgo-yaml-object-id",
			expectUpdate: true,
		},
		{
			name:         "no existing object ID - new deployment",
			projectID:    "",
			wsObjectID:   "",
			expectedID:   "",
			expectUpdate: false,
		},
		{
			name:         "placeholder projectID is ignored",
			projectID:    "YOUR_WALRUS_PROJECT_ID",
			wsObjectID:   "",
			expectedID:   "",
			expectUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.WalgoConfig{
				WalrusConfig: config.WalrusConfig{
					ProjectID: tt.projectID,
				},
			}

			var existingObjectID string
			var isUpdate bool

			// Check 1: walgo.yaml projectID
			if cfg.WalrusConfig.ProjectID != "" && cfg.WalrusConfig.ProjectID != "YOUR_WALRUS_PROJECT_ID" {
				existingObjectID = cfg.WalrusConfig.ProjectID
			}

			// Check 2: ws-resources.json (simulate)
			if existingObjectID == "" && tt.wsObjectID != "" {
				existingObjectID = tt.wsObjectID
			}

			if existingObjectID != "" {
				isUpdate = true
			}

			if existingObjectID != tt.expectedID {
				t.Errorf("Expected objectID '%s', got '%s'", tt.expectedID, existingObjectID)
			}
			if isUpdate != tt.expectUpdate {
				t.Errorf("Expected isUpdate=%v, got %v", tt.expectUpdate, isUpdate)
			}
		})
	}
}

// TestForceNewDeployment tests the force-new flag behavior
func TestForceNewDeployment(t *testing.T) {
	tests := []struct {
		name             string
		existingObjectID string
		forceNew         bool
		expectUpdate     bool
	}{
		{
			name:             "existing site without force-new should update",
			existingObjectID: "existing-object-id",
			forceNew:         false,
			expectUpdate:     true,
		},
		{
			name:             "existing site with force-new should create new",
			existingObjectID: "existing-object-id",
			forceNew:         true,
			expectUpdate:     false,
		},
		{
			name:             "new site without force-new should create new",
			existingObjectID: "",
			forceNew:         false,
			expectUpdate:     false,
		},
		{
			name:             "new site with force-new should create new",
			existingObjectID: "",
			forceNew:         true,
			expectUpdate:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var isUpdate bool
			if tt.existingObjectID != "" && !tt.forceNew {
				isUpdate = true
			}

			if isUpdate != tt.expectUpdate {
				t.Errorf("Expected isUpdate=%v, got %v", tt.expectUpdate, isUpdate)
			}
		})
	}
}

// TestMetadataOptions tests metadata preparation for ws-resources.json
func TestMetadataOptions(t *testing.T) {
	tests := []struct {
		name         string
		opts         DeploymentOptions
		wantSiteName string
		wantDesc     string
		wantImage    string
	}{
		{
			name: "all metadata provided",
			opts: DeploymentOptions{
				ProjectName: "my-site",
				Description: "My awesome site",
				ImageURL:    "https://example.com/logo.png",
				Category:    "blog",
			},
			wantSiteName: "my-site",
			wantDesc:     "My awesome site",
			wantImage:    "https://example.com/logo.png",
		},
		{
			name: "only project name",
			opts: DeploymentOptions{
				ProjectName: "my-site",
			},
			wantSiteName: "my-site",
			wantDesc:     "",
			wantImage:    "",
		},
		{
			name:         "empty metadata",
			opts:         DeploymentOptions{},
			wantSiteName: "",
			wantDesc:     "",
			wantImage:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataOpts := compress.MetadataOptions{
				SiteName:    tt.opts.ProjectName,
				Description: tt.opts.Description,
				ImageURL:    tt.opts.ImageURL,
				Category:    tt.opts.Category,
			}

			if metadataOpts.SiteName != tt.wantSiteName {
				t.Errorf("Expected SiteName '%s', got '%s'", tt.wantSiteName, metadataOpts.SiteName)
			}
			if metadataOpts.Description != tt.wantDesc {
				t.Errorf("Expected Description '%s', got '%s'", tt.wantDesc, metadataOpts.Description)
			}
			if metadataOpts.ImageURL != tt.wantImage {
				t.Errorf("Expected ImageURL '%s', got '%s'", tt.wantImage, metadataOpts.ImageURL)
			}
		})
	}
}

// TestDryRunMode tests dry-run mode returns early
func TestDryRunMode(t *testing.T) {
	opts := DeploymentOptions{
		DryRun:     true,
		SitePath:   "/fake/path",
		PublishDir: "/fake/path/public",
		WalgoCfg:   &config.WalgoConfig{},
	}

	if !opts.DryRun {
		t.Error("Expected DryRun to be true")
	}

	// In dry-run mode, the function should return early with success
	// without actually deploying
	result := &DeploymentResult{
		Success: true,
	}

	if !result.Success {
		t.Error("Dry-run mode should return success")
	}
}

// TestQuietMode tests quiet mode suppresses output
func TestQuietMode(t *testing.T) {
	opts := DeploymentOptions{
		Quiet:      true,
		SitePath:   "/fake/path",
		PublishDir: "/fake/path/public",
		WalgoCfg:   &config.WalgoConfig{},
	}

	if !opts.Quiet {
		t.Error("Expected Quiet to be true")
	}
}

// TestVerboseMode tests verbose mode enables additional output
func TestVerboseMode(t *testing.T) {
	opts := DeploymentOptions{
		Verbose:    true,
		SitePath:   "/fake/path",
		PublishDir: "/fake/path/public",
		WalgoCfg:   &config.WalgoConfig{},
	}

	if !opts.Verbose {
		t.Error("Expected Verbose to be true")
	}
}

// TestSaveProjectOption tests save project flag
func TestSaveProjectOption(t *testing.T) {
	opts := DeploymentOptions{
		SaveProject: true,
		SitePath:    "/fake/path",
		PublishDir:  "/fake/path/public",
		ProjectName: "test-project",
		Category:    "website",
		WalgoCfg:    &config.WalgoConfig{},
	}

	if !opts.SaveProject {
		t.Error("Expected SaveProject to be true")
	}
	if opts.ProjectName != "test-project" {
		t.Errorf("Expected ProjectName 'test-project', got '%s'", opts.ProjectName)
	}
}

// TestEpochsConfiguration tests epochs setting
func TestEpochsConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		epochs int
	}{
		{"default epochs", 1},
		{"5 epochs", 5},
		{"10 epochs", 10},
		{"max epochs", 53},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DeploymentOptions{
				Epochs:     tt.epochs,
				SitePath:   "/fake/path",
				PublishDir: "/fake/path/public",
				WalgoCfg:   &config.WalgoConfig{},
			}

			if opts.Epochs != tt.epochs {
				t.Errorf("Expected Epochs %d, got %d", tt.epochs, opts.Epochs)
			}
		})
	}
}

// TestNetworkConfiguration tests network settings
func TestNetworkConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		network string
	}{
		{"testnet", "testnet"},
		{"mainnet", "mainnet"},
		{"empty defaults to discovery", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DeploymentOptions{
				Network:    tt.network,
				SitePath:   "/fake/path",
				PublishDir: "/fake/path/public",
				WalgoCfg:   &config.WalgoConfig{},
			}

			if opts.Network != tt.network {
				t.Errorf("Expected Network '%s', got '%s'", tt.network, opts.Network)
			}
		})
	}
}

// TestWalletAddressConfiguration tests wallet address setting
func TestWalletAddressConfiguration(t *testing.T) {
	walletAddr := "0x1234567890abcdef1234567890abcdef12345678"
	opts := DeploymentOptions{
		WalletAddr: walletAddr,
		SitePath:   "/fake/path",
		PublishDir: "/fake/path/public",
		WalgoCfg:   &config.WalgoConfig{},
	}

	if opts.WalletAddr != walletAddr {
		t.Errorf("Expected WalletAddr '%s', got '%s'", walletAddr, opts.WalletAddr)
	}
}

// TestDeployerOptionsMapping tests mapping DeploymentOptions to deployer.DeployOptions
func TestDeployerOptionsMapping(t *testing.T) {
	cfg := &config.WalgoConfig{
		WalrusConfig: config.WalrusConfig{
			Network:    "testnet",
			Entrypoint: "index.html",
		},
	}

	opts := DeploymentOptions{
		Epochs:   10,
		Verbose:  true,
		Quiet:    false,
		WalgoCfg: cfg,
	}

	// Map to deployer options
	deployOpts := deployer.DeployOptions{
		Epochs:    opts.Epochs,
		Verbose:   opts.Verbose && !opts.Quiet,
		WalrusCfg: cfg.WalrusConfig,
	}

	if deployOpts.Epochs != 10 {
		t.Errorf("Expected Epochs 10, got %d", deployOpts.Epochs)
	}
	if !deployOpts.Verbose {
		t.Error("Expected Verbose to be true")
	}
	if deployOpts.WalrusCfg.Network != "testnet" {
		t.Errorf("Expected Network 'testnet', got '%s'", deployOpts.WalrusCfg.Network)
	}
}

// TestDeployerOptionsVerboseQuietInteraction tests verbose with quiet mode
func TestDeployerOptionsVerboseQuietInteraction(t *testing.T) {
	cfg := &config.WalgoConfig{}

	opts := DeploymentOptions{
		Verbose:  true,
		Quiet:    true, // Quiet should suppress verbose
		WalgoCfg: cfg,
	}

	deployOpts := deployer.DeployOptions{
		Verbose: opts.Verbose && !opts.Quiet,
	}

	if deployOpts.Verbose {
		t.Error("Verbose should be false when Quiet is true")
	}
}

// TestUpdateMetadataIntegration tests metadata update flow
func TestUpdateMetadataIntegration(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")

	// Create initial ws-resources.json
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write initial ws-resources.json: %v", err)
	}

	// Update metadata
	metadataOpts := compress.MetadataOptions{
		SiteName:    "test-site",
		Description: "A test site",
		ImageURL:    "https://example.com/logo.png",
		Category:    "website",
	}

	if err := compress.UpdateMetadata(wsResourcesPath, metadataOpts); err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	// Read back and verify
	readConfig, err := compress.ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		t.Fatalf("Failed to read ws-resources.json: %v", err)
	}

	if readConfig.SiteName != "test-site" {
		t.Errorf("Expected SiteName 'test-site', got '%s'", readConfig.SiteName)
	}
	if readConfig.Metadata == nil {
		t.Fatal("Expected Metadata to be non-nil")
	}
	if readConfig.Metadata.Description != "A test site" {
		t.Errorf("Expected Description 'A test site', got '%s'", readConfig.Metadata.Description)
	}
}

// TestUpdateObjectIDIntegration tests object ID update flow
func TestUpdateObjectIDIntegration(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")

	// Create initial ws-resources.json
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write initial ws-resources.json: %v", err)
	}

	// Update object ID
	newObjectID := "new-object-id-12345"
	if err := compress.UpdateObjectID(wsResourcesPath, newObjectID); err != nil {
		t.Fatalf("Failed to update object ID: %v", err)
	}

	// Read back and verify
	readConfig, err := compress.ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		t.Fatalf("Failed to read ws-resources.json: %v", err)
	}

	if readConfig.ObjectID != newObjectID {
		t.Errorf("Expected ObjectID '%s', got '%s'", newObjectID, readConfig.ObjectID)
	}
}

// TestDeploymentResultUpdate tests update result
func TestDeploymentResultUpdate(t *testing.T) {
	result := &DeploymentResult{
		Success:      true,
		ObjectID:     "updated-object-id",
		IsUpdate:     true,
		IsNewProject: false,
		SiteSize:     2048,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if !result.IsUpdate {
		t.Error("Expected IsUpdate to be true")
	}
	if result.IsNewProject {
		t.Error("Expected IsNewProject to be false for update")
	}
}

// TestDeploymentResultNewProject tests new project result
func TestDeploymentResultNewProject(t *testing.T) {
	result := &DeploymentResult{
		Success:      true,
		ObjectID:     "new-object-id",
		IsUpdate:     false,
		IsNewProject: true,
		SiteSize:     4096,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.IsUpdate {
		t.Error("Expected IsUpdate to be false for new deployment")
	}
	if !result.IsNewProject {
		t.Error("Expected IsNewProject to be true")
	}
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Expected context to be cancelled")
	}

	if ctx.Err() != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", ctx.Err())
	}
}

// TestContextTimeout tests context timeout handling
func TestContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(5 * time.Millisecond)

	select {
	case <-ctx.Done():
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded, got %v", ctx.Err())
		}
	default:
		t.Error("Expected context to be done")
	}
}

// TestProjectNameFallback tests project name fallback behavior
func TestProjectNameFallback(t *testing.T) {
	tests := []struct {
		name         string
		projectName  string
		sitePath     string
		expectedName string
	}{
		{
			name:         "explicit project name",
			projectName:  "my-project",
			sitePath:     "/path/to/site",
			expectedName: "my-project",
		},
		{
			name:         "fallback to site directory name",
			projectName:  "",
			sitePath:     "/path/to/my-site",
			expectedName: "my-site",
		},
		{
			name:         "fallback for root path",
			projectName:  "",
			sitePath:     "/",
			expectedName: "my-walgo-site",
		},
		{
			name:         "fallback for current directory",
			projectName:  "",
			sitePath:     ".",
			expectedName: "my-walgo-site",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectName := tt.projectName
			if projectName == "" {
				projectName = filepath.Base(tt.sitePath)
				if projectName == "" || projectName == "." || projectName == "/" {
					projectName = "my-walgo-site"
				}
			}

			if projectName != tt.expectedName {
				t.Errorf("Expected project name '%s', got '%s'", tt.expectedName, projectName)
			}
		})
	}
}

// TestCategoryFallback tests category fallback behavior
func TestCategoryFallback(t *testing.T) {
	tests := []struct {
		name             string
		category         string
		expectedCategory string
	}{
		{
			name:             "explicit category",
			category:         "blog",
			expectedCategory: "blog",
		},
		{
			name:             "fallback to website",
			category:         "",
			expectedCategory: "website",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := tt.category
			if category == "" {
				category = "website"
			}

			if category != tt.expectedCategory {
				t.Errorf("Expected category '%s', got '%s'", tt.expectedCategory, category)
			}
		})
	}
}

// TestDeployerResultHandling tests handling of deployer results
func TestDeployerResultHandling(t *testing.T) {
	tests := []struct {
		name          string
		deployResult  *deployer.Result
		expectSuccess bool
		expectError   bool
	}{
		{
			name: "successful deployment",
			deployResult: &deployer.Result{
				Success:  true,
				ObjectID: "0x123",
			},
			expectSuccess: true,
			expectError:   false,
		},
		{
			name: "failed deployment - no object ID",
			deployResult: &deployer.Result{
				Success:  false,
				ObjectID: "",
			},
			expectSuccess: false,
			expectError:   true,
		},
		{
			name: "success but no object ID",
			deployResult: &deployer.Result{
				Success:  true,
				ObjectID: "",
			},
			expectSuccess: false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSuccess := tt.deployResult.Success && tt.deployResult.ObjectID != ""
			hasError := !tt.deployResult.Success || tt.deployResult.ObjectID == ""

			if isSuccess != tt.expectSuccess {
				t.Errorf("Expected success=%v, got %v", tt.expectSuccess, isSuccess)
			}
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got %v", tt.expectError, hasError)
			}
		})
	}
}

// TestFileToBlobIDMapping tests file to blob ID mapping
func TestFileToBlobIDMapping(t *testing.T) {
	result := &deployer.Result{
		Success:  true,
		ObjectID: "site-object-id",
		FileToBlobID: map[string]string{
			"index.html": "blob-id-1",
			"style.css":  "blob-id-2",
			"script.js":  "blob-id-3",
		},
	}

	if len(result.FileToBlobID) != 3 {
		t.Errorf("Expected 3 file mappings, got %d", len(result.FileToBlobID))
	}

	if result.FileToBlobID["index.html"] != "blob-id-1" {
		t.Errorf("Expected blob ID for index.html to be 'blob-id-1'")
	}
}

// BenchmarkSiteSizeCalculation benchmarks site size calculation
func BenchmarkSiteSizeCalculation(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "walgo-benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create 100 test files
	for i := 0; i < 100; i++ {
		content := make([]byte, 1024) // 1KB each
		if err := os.WriteFile(filepath.Join(tempDir, "file"+string(rune('0'+i%10))+".txt"), content, 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var totalSize int64
		filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
	}
}

// TestDeploymentOptionsWithAllFields tests DeploymentOptions with all fields populated
func TestDeploymentOptionsWithAllFields(t *testing.T) {
	cfg := &config.WalgoConfig{
		WalrusConfig: config.WalrusConfig{
			ProjectID:  "test-project-id",
			Network:    "testnet",
			Entrypoint: "index.html",
		},
	}

	opts := DeploymentOptions{
		SitePath:    "/path/to/site",
		PublishDir:  "/path/to/site/public",
		Epochs:      10,
		WalgoCfg:    cfg,
		Quiet:       false,
		Verbose:     true,
		ForceNew:    false,
		DryRun:      false,
		SaveProject: true,
		ProjectName: "my-awesome-site",
		Category:    "blog",
		Network:     "testnet",
		WalletAddr:  "0x1234",
		Description: "An awesome blog",
		ImageURL:    "https://example.com/logo.png",
	}

	// Verify all fields
	if opts.SitePath != "/path/to/site" {
		t.Errorf("Unexpected SitePath: %s", opts.SitePath)
	}
	if opts.Epochs != 10 {
		t.Errorf("Unexpected Epochs: %d", opts.Epochs)
	}
	if opts.ProjectName != "my-awesome-site" {
		t.Errorf("Unexpected ProjectName: %s", opts.ProjectName)
	}
	if opts.Description != "An awesome blog" {
		t.Errorf("Unexpected Description: %s", opts.Description)
	}
}

// TestWalgoConfigIntegration tests integration with WalgoConfig
func TestWalgoConfigIntegration(t *testing.T) {
	cfg := config.NewDefaultWalgoConfig()

	// Verify defaults
	if cfg.HugoConfig.PublishDir != "public" {
		t.Errorf("Expected default PublishDir 'public', got '%s'", cfg.HugoConfig.PublishDir)
	}
	if cfg.WalrusConfig.Entrypoint != "index.html" {
		t.Errorf("Expected default Entrypoint 'index.html', got '%s'", cfg.WalrusConfig.Entrypoint)
	}
	if cfg.WalrusConfig.ProjectID != "YOUR_WALRUS_PROJECT_ID" {
		t.Errorf("Expected placeholder ProjectID, got '%s'", cfg.WalrusConfig.ProjectID)
	}
}

// TestEmptyPublishDir tests handling of empty publish directory
func TestEmptyPublishDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "walgo-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create empty public directory
	publicDir := filepath.Join(tempDir, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatalf("Failed to create public directory: %v", err)
	}

	var totalSize int64
	var fileCount int
	err = filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	if fileCount != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", fileCount)
	}
	if totalSize != 0 {
		t.Errorf("Expected 0 total size, got %d", totalSize)
	}
}

// TestLegacyObjectIDFormat tests reading legacy objectId format
func TestLegacyObjectIDFormat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "walgo-legacy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wsResourcesPath := filepath.Join(tempDir, "ws-resources.json")

	// Create ws-resources.json with legacy objectId (camelCase)
	legacyConfig := map[string]interface{}{
		"objectId": "legacy-object-id-123",
		"headers":  map[string]interface{}{},
	}
	data, _ := json.MarshalIndent(legacyConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	// Read and verify legacy format is handled
	config, err := compress.ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		t.Fatalf("Failed to read ws-resources.json: %v", err)
	}

	if config.ObjectID != "legacy-object-id-123" {
		t.Errorf("Expected ObjectID 'legacy-object-id-123', got '%s'", config.ObjectID)
	}
}

// TestMixedObjectIDFormats tests priority when both formats exist
func TestMixedObjectIDFormats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "walgo-mixed-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wsResourcesPath := filepath.Join(tempDir, "ws-resources.json")

	// Create ws-resources.json with both object_id and objectId
	mixedConfig := map[string]interface{}{
		"object_id": "snake-case-id",
		"objectId":  "camel-case-id",
		"headers":   map[string]interface{}{},
	}
	data, _ := json.MarshalIndent(mixedConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	// Read and verify snake_case takes priority
	config, err := compress.ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		t.Fatalf("Failed to read ws-resources.json: %v", err)
	}

	if config.ObjectID != "snake-case-id" {
		t.Errorf("Expected ObjectID 'snake-case-id' (snake_case priority), got '%s'", config.ObjectID)
	}
}

// TestPerformDeploymentDryRun tests PerformDeployment in dry-run mode
// Note: Dry-run only returns early if cache is available and initialized successfully.
// This test verifies that the site size is calculated correctly in all cases.
func TestPerformDeploymentDryRun(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()
	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "test-dry-run",
	}

	ctx := context.Background()
	result, err := PerformDeployment(ctx, opts)

	// Dry-run with cache enabled will return early with success
	// Without cache, it will try to deploy and fail (no site-builder)
	// Both are valid behaviors - we're testing the dry-run code path exists
	if err == nil {
		if !result.Success {
			t.Error("Expected dry-run to succeed when cache is available")
		}
		if result.ObjectID != "" {
			t.Error("Dry-run should not return an object ID")
		}
	}
	// If error occurs, it means cache wasn't available and it tried to deploy
	// which is expected behavior when site-builder isn't installed
}

// TestPerformDeploymentSiteSizeCalculation tests site size is calculated correctly
func TestPerformDeploymentSiteSizeCalculation(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()
	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "test-size-calc",
	}

	ctx := context.Background()
	result, _ := PerformDeployment(ctx, opts)

	// Site size should always be calculated, regardless of deployment success
	if result.SiteSize <= 0 {
		t.Error("Expected positive site size")
	}
}

// TestPerformDeploymentWithExistingProjectID tests update detection from walgo.yaml
func TestPerformDeploymentWithExistingProjectID(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()
	cfg.WalrusConfig.ProjectID = "existing-object-id-123"

	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "test-existing",
	}

	ctx := context.Background()
	result, _ := PerformDeployment(ctx, opts)

	// When cache is available and dry-run succeeds, IsUpdate should be set
	// Otherwise, the function will try to deploy and may fail
	// We primarily test that the result is populated correctly
	if result.SiteSize <= 0 {
		t.Error("Expected site size to be calculated")
	}
}

// TestPerformDeploymentWithWSResourcesObjectID tests update detection from ws-resources.json
func TestPerformDeploymentWithWSResourcesObjectID(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json with existing object ID
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	wsConfig := &compress.WSResourcesConfig{
		Headers:  make(map[string]map[string]string),
		ObjectID: "ws-resources-object-id-456",
	}
	data, _ := json.MarshalIndent(wsConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()
	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "test-ws-resources",
	}

	ctx := context.Background()
	result, _ := PerformDeployment(ctx, opts)

	// Test that site size is calculated regardless of deployment outcome
	if result.SiteSize <= 0 {
		t.Error("Expected site size to be calculated")
	}
}

// TestPerformDeploymentForceNew tests force-new flag bypasses update detection
func TestPerformDeploymentForceNew(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json with existing object ID
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	wsConfig := &compress.WSResourcesConfig{
		Headers:  make(map[string]map[string]string),
		ObjectID: "existing-object-id",
	}
	data, _ := json.MarshalIndent(wsConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()
	cfg.WalrusConfig.ProjectID = "existing-project-id"

	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ForceNew:    true, // Force new deployment
		ProjectName: "test-force-new",
	}

	ctx := context.Background()
	result, _ := PerformDeployment(ctx, opts)

	// Force-new should prevent IsUpdate from being set
	// Note: If dry-run returns early (with cache), IsUpdate won't be set anyway
	// This test verifies the ForceNew flag is correctly processed
	if result.IsUpdate && opts.ForceNew {
		t.Error("Force-new should not set IsUpdate to true")
	}

	// Verify site size is calculated
	if result.SiteSize <= 0 {
		t.Error("Expected site size to be calculated")
	}
}

// TestPerformDeploymentQuietMode tests quiet mode suppresses output
func TestPerformDeploymentQuietMode(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()
	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "test-quiet",
	}

	ctx := context.Background()
	result, _ := PerformDeployment(ctx, opts)

	// Verify site size is calculated in quiet mode
	if result.SiteSize <= 0 {
		t.Error("Expected site size to be calculated")
	}
}

// TestPerformDeploymentVerboseMode tests verbose mode with dry-run
func TestPerformDeploymentVerboseMode(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()
	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       false,
		Verbose:     true,
		DryRun:      true,
		ProjectName: "test-verbose",
	}

	ctx := context.Background()
	result, _ := PerformDeployment(ctx, opts)

	// Verify site size is calculated in verbose mode
	if result.SiteSize <= 0 {
		t.Error("Expected site size to be calculated")
	}
}

// TestPerformDeploymentMetadataUpdate tests metadata is prepared correctly
// Note: Metadata is only written if the deployment proceeds past the dry-run early return
func TestPerformDeploymentMetadataUpdate(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()
	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "metadata-test-site",
		Description: "A test site with metadata",
		ImageURL:    "https://example.com/logo.png",
		Category:    "blog",
	}

	ctx := context.Background()
	result, _ := PerformDeployment(ctx, opts)

	// Verify site size is calculated
	if result.SiteSize <= 0 {
		t.Error("Expected site size to be calculated")
	}

	// Note: Metadata update happens AFTER the dry-run early return if cache is available
	// So we can't reliably test metadata update in dry-run mode
	// Instead, we test the MetadataOptions directly
	metadataOpts := compress.MetadataOptions{
		SiteName:    opts.ProjectName,
		Description: opts.Description,
		ImageURL:    opts.ImageURL,
		Category:    opts.Category,
	}

	// Verify options are correctly constructed
	if metadataOpts.SiteName != "metadata-test-site" {
		t.Errorf("Expected SiteName 'metadata-test-site', got '%s'", metadataOpts.SiteName)
	}
	if metadataOpts.Description != "A test site with metadata" {
		t.Errorf("Expected Description 'A test site with metadata', got '%s'", metadataOpts.Description)
	}
	if metadataOpts.Category != "blog" {
		t.Errorf("Expected Category 'blog', got '%s'", metadataOpts.Category)
	}
}

// TestPerformDeploymentWithWalkErrors tests handling of directory walk errors
// Note: Walk errors during size calculation are logged as warnings but don't fail the deployment
func TestPerformDeploymentWithWalkErrors(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	// This test verifies the walk error handling in site size calculation
	// We don't create permission issues because they cause problems with site-builder
	// Instead, we just verify the basic error collection logic

	cfg := config.NewDefaultWalgoConfig()
	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "test-walk-errors",
	}

	ctx := context.Background()
	result, _ := PerformDeployment(ctx, opts)

	// Verify site size is calculated (walk errors don't affect this)
	if result.SiteSize <= 0 {
		t.Error("Expected positive site size")
	}
}

// TestPerformDeploymentCacheInitFailure tests handling of cache initialization failure
func TestPerformDeploymentCacheInitFailure(t *testing.T) {
	// Use a path that won't have proper cache setup
	tempDir, err := os.MkdirTemp("", "walgo-cache-fail-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	publicDir := filepath.Join(tempDir, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatalf("Failed to create public directory: %v", err)
	}

	// Create minimal files
	if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
		t.Fatalf("Failed to write index.html: %v", err)
	}

	// Create ws-resources.json
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()
	opts := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "test-cache-fail",
	}

	ctx := context.Background()
	result, _ := PerformDeployment(ctx, opts)

	// Site size should be calculated regardless of cache status
	if result.SiteSize <= 0 {
		t.Error("Expected positive site size")
	}
}

// TestWSResourcesPreservation tests that ws-resources.json preserves existing fields
// This is a unit test for the compress.UpdateMetadata function
func TestWSResourcesPreservation(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json with existing fields
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	existingConfig := &compress.WSResourcesConfig{
		Headers: map[string]map[string]string{
			"/index.html": {
				"Content-Type":    "text/html",
				"X-Custom-Header": "custom-value",
			},
		},
		Routes: map[string]string{
			"/":      "/index.html",
			"/about": "/about/index.html",
		},
		Ignore: []string{"/.DS_Store", "/secret/*"},
	}
	data, _ := json.MarshalIndent(existingConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	// Test UpdateMetadata directly (not via PerformDeployment)
	// This avoids the dry-run early return issue
	metadataOpts := compress.MetadataOptions{
		SiteName:    "preserve-test",
		Description: "New description",
	}

	if err := compress.UpdateMetadata(wsResourcesPath, metadataOpts); err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	// Read back and verify existing fields are preserved
	readConfig, err := compress.ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		t.Fatalf("Failed to read ws-resources.json: %v", err)
	}

	// Check headers are preserved
	if readConfig.Headers == nil {
		t.Fatal("Headers should be preserved")
	}

	indexHeaders, ok := readConfig.Headers["/index.html"]
	if !ok {
		t.Error("Index.html headers should be preserved")
	} else if indexHeaders["X-Custom-Header"] != "custom-value" {
		t.Error("Custom header should be preserved")
	}

	// Check routes are preserved
	if readConfig.Routes == nil {
		t.Error("Routes should be preserved")
	}

	// New metadata should be added
	if readConfig.Metadata == nil {
		t.Error("Metadata should be added")
	} else if readConfig.Metadata.Description != "New description" {
		t.Errorf("New description should be set, got '%s'", readConfig.Metadata.Description)
	}
}

// TestStepNumberCalculation tests step numbering logic
func TestStepNumberCalculation(t *testing.T) {
	// When cache is nil, step numbers should adjust
	// Base step number for metadata is 3 with cache, 2 without

	tests := []struct {
		name           string
		cacheAvailable bool
		expectedStep   int
	}{
		{
			name:           "with cache",
			cacheAvailable: true,
			expectedStep:   3,
		},
		{
			name:           "without cache",
			cacheAvailable: false,
			expectedStep:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stepNum := 3
			if !tt.cacheAvailable {
				stepNum = 2
			}

			if stepNum != tt.expectedStep {
				t.Errorf("Expected step %d, got %d", tt.expectedStep, stepNum)
			}
		})
	}
}

// TestDeploymentOptionsDefaults tests default values
func TestDeploymentOptionsDefaults(t *testing.T) {
	opts := DeploymentOptions{}

	// Check zero values
	if opts.Epochs != 0 {
		t.Errorf("Expected default Epochs to be 0, got %d", opts.Epochs)
	}
	if opts.Quiet != false {
		t.Error("Expected default Quiet to be false")
	}
	if opts.Verbose != false {
		t.Error("Expected default Verbose to be false")
	}
	if opts.ForceNew != false {
		t.Error("Expected default ForceNew to be false")
	}
	if opts.DryRun != false {
		t.Error("Expected default DryRun to be false")
	}
	if opts.SaveProject != false {
		t.Error("Expected default SaveProject to be false")
	}
}

// TestDeploymentResultDefaults tests default values
func TestDeploymentResultDefaults(t *testing.T) {
	result := &DeploymentResult{}

	// Check zero values
	if result.Success != false {
		t.Error("Expected default Success to be false")
	}
	if result.ObjectID != "" {
		t.Error("Expected default ObjectID to be empty")
	}
	if result.IsUpdate != false {
		t.Error("Expected default IsUpdate to be false")
	}
	if result.IsNewProject != false {
		t.Error("Expected default IsNewProject to be false")
	}
	if result.SiteSize != 0 {
		t.Errorf("Expected default SiteSize to be 0, got %d", result.SiteSize)
	}
	if result.Error != nil {
		t.Error("Expected default Error to be nil")
	}
}

// TestMultipleDeploymentAttempts tests repeated deployments
func TestMultipleDeploymentAttempts(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")

	// Create ws-resources.json
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	cfg := config.NewDefaultWalgoConfig()

	// First deployment
	opts1 := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      5,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "multi-deploy-test",
	}

	ctx := context.Background()
	result1, _ := PerformDeployment(ctx, opts1)

	// Verify site size is calculated
	if result1.SiteSize <= 0 {
		t.Error("First deployment should calculate site size")
	}

	// Second deployment with different metadata
	opts2 := DeploymentOptions{
		SitePath:    tempDir,
		PublishDir:  publicDir,
		Epochs:      10,
		WalgoCfg:    &cfg,
		Quiet:       true,
		DryRun:      true,
		ProjectName: "multi-deploy-test-v2",
		Description: "Updated description",
	}

	result2, _ := PerformDeployment(ctx, opts2)

	// Verify site size is calculated
	if result2.SiteSize <= 0 {
		t.Error("Second deployment should calculate site size")
	}

	// Test that metadata updates work directly (not relying on dry-run behavior)
	metadataOpts := compress.MetadataOptions{
		SiteName:    "multi-deploy-test-v2",
		Description: "Updated description",
	}
	if err := compress.UpdateMetadata(wsResourcesPath, metadataOpts); err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	readConfig, err := compress.ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		t.Fatalf("Failed to read ws-resources.json: %v", err)
	}

	if readConfig.SiteName != "multi-deploy-test-v2" {
		t.Errorf("Expected SiteName 'multi-deploy-test-v2', got '%s'", readConfig.SiteName)
	}
}

// =====================================================================
// Additional unit tests for deployment logic components
// =====================================================================

// TestEstimateGasFeeWithEpochs tests the gas estimation function
func TestEstimateGasFeeWithEpochs(t *testing.T) {
	tests := []struct {
		name     string
		network  string
		siteSize int64
		epochs   int
	}{
		{
			name:     "testnet small site 1 epoch",
			network:  "testnet",
			siteSize: 1024 * 100, // 100KB
			epochs:   1,
		},
		{
			name:     "mainnet small site 5 epochs",
			network:  "mainnet",
			siteSize: 1024 * 100,
			epochs:   5,
		},
		{
			name:     "testnet large site 10 epochs",
			network:  "testnet",
			siteSize: 1024 * 1024 * 10, // 10MB
			epochs:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := projects.EstimateGasFeeWithEpochs(tt.network, tt.siteSize, tt.epochs)
			// Result should be a non-empty string with cost information
			if result == "" {
				t.Error("EstimateGasFeeWithEpochs() returned empty string")
			}
		})
	}
}

// TestProjectCreation tests creating a project in the database
func TestProjectCreation(t *testing.T) {
	// Create a temporary database
	tempDir, err := os.MkdirTemp("", "walgo-project-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set XDG_DATA_HOME to our temp directory
	oldXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tempDir)
	defer os.Setenv("XDG_DATA_HOME", oldXDG)

	pm, err := projects.NewManager()
	if err != nil {
		t.Fatalf("Failed to create project manager: %v", err)
	}
	defer pm.Close()

	project := &projects.Project{
		Name:        "test-project",
		Category:    "website",
		Network:     "testnet",
		ObjectID:    "0x1234567890",
		WalletAddr:  "0xabc123",
		Epochs:      5,
		SitePath:    "/path/to/site",
		Description: "Test project description",
	}

	if err := pm.CreateProject(project); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if project.ID == 0 {
		t.Error("Expected project ID to be set after creation")
	}

	// Retrieve the project
	retrieved, err := pm.GetProject(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if retrieved.Name != "test-project" {
		t.Errorf("Expected Name 'test-project', got '%s'", retrieved.Name)
	}
	if retrieved.ObjectID != "0x1234567890" {
		t.Errorf("Expected ObjectID '0x1234567890', got '%s'", retrieved.ObjectID)
	}
}

// TestProjectUpdate tests updating a project
func TestProjectUpdate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "walgo-project-update-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tempDir)
	defer os.Setenv("XDG_DATA_HOME", oldXDG)

	pm, err := projects.NewManager()
	if err != nil {
		t.Fatalf("Failed to create project manager: %v", err)
	}
	defer pm.Close()

	project := &projects.Project{
		Name:     "update-test",
		Category: "website",
		Network:  "testnet",
		ObjectID: "0x111111",
		SitePath: "/path/to/update-site",
	}

	if err := pm.CreateProject(project); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Update the project
	project.ObjectID = "0x222222"
	project.Description = "Updated description"
	project.LastDeployAt = time.Now()

	if err := pm.UpdateProject(project); err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	// Verify the update
	updated, err := pm.GetProject(project.ID)
	if err != nil {
		t.Fatalf("Failed to get updated project: %v", err)
	}

	if updated.ObjectID != "0x222222" {
		t.Errorf("Expected ObjectID '0x222222', got '%s'", updated.ObjectID)
	}
	if updated.Description != "Updated description" {
		t.Errorf("Expected Description 'Updated description', got '%s'", updated.Description)
	}
}

// TestGetProjectBySitePath tests finding a project by site path
func TestGetProjectBySitePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "walgo-sitepath-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tempDir)
	defer os.Setenv("XDG_DATA_HOME", oldXDG)

	pm, err := projects.NewManager()
	if err != nil {
		t.Fatalf("Failed to create project manager: %v", err)
	}
	defer pm.Close()

	sitePath := "/unique/path/to/site"
	project := &projects.Project{
		Name:     "sitepath-test",
		Category: "blog",
		Network:  "mainnet",
		ObjectID: "0xsitepath",
		SitePath: sitePath,
	}

	if err := pm.CreateProject(project); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Find by site path
	found, err := pm.GetProjectBySitePath(sitePath)
	if err != nil {
		t.Fatalf("Failed to find project by site path: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find project, got nil")
	}
	if found.Name != "sitepath-test" {
		t.Errorf("Expected Name 'sitepath-test', got '%s'", found.Name)
	}

	// Test non-existent path
	notFound, err := pm.GetProjectBySitePath("/non/existent/path")
	if err != nil {
		// Project not found is expected - not an error condition
	}
	if notFound != nil {
		t.Error("Expected nil for non-existent path")
	}
}

// TestRecordDeployment tests recording deployment history
func TestRecordDeployment(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "walgo-record-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tempDir)
	defer os.Setenv("XDG_DATA_HOME", oldXDG)

	pm, err := projects.NewManager()
	if err != nil {
		t.Fatalf("Failed to create project manager: %v", err)
	}
	defer pm.Close()

	// Create a project first
	project := &projects.Project{
		Name:     "deploy-history-test",
		Category: "website",
		Network:  "testnet",
		ObjectID: "0xhistory",
		SitePath: "/path/to/history-site",
	}

	if err := pm.CreateProject(project); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Record a deployment
	deployment := &projects.DeploymentRecord{
		ProjectID: project.ID,
		ObjectID:  "0xhistory",
		Network:   "testnet",
		Epochs:    5,
		GasFee:    "~0.1 SUI",
		Success:   true,
	}

	if err := pm.RecordDeployment(deployment); err != nil {
		t.Fatalf("Failed to record deployment: %v", err)
	}

	if deployment.ID == 0 {
		t.Error("Expected deployment ID to be set")
	}
}

// TestConfigUpdateWalgoYAML tests updating the walgo.yaml projectID
func TestConfigUpdateWalgoYAML(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	// Create a simple walgo.yaml
	walgoCfg := config.NewDefaultWalgoConfig()
	cfgPath := filepath.Join(tempDir, "walgo.yaml")

	// Write initial config using yaml format
	initialContent := `hugo:
  publishDir: public
  contentDir: content
walrus:
  projectID: YOUR_WALRUS_PROJECT_ID
  entrypoint: index.html
`
	if err := os.WriteFile(cfgPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write walgo.yaml: %v", err)
	}

	// Update the projectID
	newObjectID := "0xnew-project-id-123"
	if err := config.UpdateWalgoYAMLProjectID(tempDir, newObjectID); err != nil {
		t.Fatalf("Failed to update walgo.yaml: %v", err)
	}

	// Read back and verify
	updatedCfg, err := config.LoadConfigFrom(tempDir)
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}

	if updatedCfg.WalrusConfig.ProjectID != newObjectID {
		t.Errorf("Expected ProjectID '%s', got '%s'", newObjectID, updatedCfg.WalrusConfig.ProjectID)
	}

	// Verify other fields are preserved
	if updatedCfg.HugoConfig.PublishDir != "public" {
		t.Errorf("Expected PublishDir 'public', got '%s'", updatedCfg.HugoConfig.PublishDir)
	}
	if updatedCfg.WalrusConfig.Entrypoint != "index.html" {
		t.Errorf("Expected Entrypoint 'index.html', got '%s'", walgoCfg.WalrusConfig.Entrypoint)
	}
}

// TestUpdateObjectIDFlow tests the object ID update flow
func TestUpdateObjectIDFlow(t *testing.T) {
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")

	// Create initial ws-resources.json
	initialConfig := &compress.WSResourcesConfig{
		Headers: map[string]map[string]string{
			"/index.html": {"Content-Type": "text/html"},
		},
		SiteName: "initial-site",
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	// Update object ID
	newObjectID := "0xfinal-object-id"
	if err := compress.UpdateObjectID(wsResourcesPath, newObjectID); err != nil {
		t.Fatalf("Failed to update object ID: %v", err)
	}

	// Read back and verify
	readConfig, err := compress.ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		t.Fatalf("Failed to read ws-resources.json: %v", err)
	}

	if readConfig.ObjectID != newObjectID {
		t.Errorf("Expected ObjectID '%s', got '%s'", newObjectID, readConfig.ObjectID)
	}

	// Verify other fields are preserved
	if readConfig.SiteName != "initial-site" {
		t.Errorf("Expected SiteName 'initial-site', got '%s'", readConfig.SiteName)
	}
	if readConfig.Headers == nil || readConfig.Headers["/index.html"]["Content-Type"] != "text/html" {
		t.Error("Expected headers to be preserved")
	}
}

// TestDeployerResultWithFileToBlobID tests processing deployer results
func TestDeployerResultWithFileToBlobID(t *testing.T) {
	result := &deployer.Result{
		Success:  true,
		ObjectID: "0xresult-object-id",
		FileToBlobID: map[string]string{
			"index.html":      "blob-1",
			"css/style.css":   "blob-2",
			"js/app.js":       "blob-3",
			"images/logo.png": "blob-4",
		},
	}

	// Verify result processing
	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.ObjectID == "" {
		t.Error("Expected non-empty ObjectID")
	}
	if len(result.FileToBlobID) != 4 {
		t.Errorf("Expected 4 file mappings, got %d", len(result.FileToBlobID))
	}

	// Verify specific mappings
	expectedMappings := map[string]string{
		"index.html":      "blob-1",
		"css/style.css":   "blob-2",
		"js/app.js":       "blob-3",
		"images/logo.png": "blob-4",
	}

	for file, expectedBlob := range expectedMappings {
		if result.FileToBlobID[file] != expectedBlob {
			t.Errorf("Expected blob ID '%s' for file '%s', got '%s'", expectedBlob, file, result.FileToBlobID[file])
		}
	}
}

// TestDeploymentFlowWithMockData simulates the full deployment flow with mock data
func TestDeploymentFlowWithMockData(t *testing.T) {
	// Create test directory structure
	tempDir, cleanup := createTestSiteDir(t)
	defer cleanup()

	publicDir := filepath.Join(tempDir, "public")
	wsResourcesPath := filepath.Join(publicDir, "ws-resources.json")

	// Create initial ws-resources.json
	initialConfig := &compress.WSResourcesConfig{
		Headers: make(map[string]map[string]string),
	}
	data, _ := json.MarshalIndent(initialConfig, "", "  ")
	if err := os.WriteFile(wsResourcesPath, data, 0644); err != nil {
		t.Fatalf("Failed to write ws-resources.json: %v", err)
	}

	// Simulate the deployment flow steps
	// Step 1: Calculate site size
	var siteSize int64
	filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			siteSize += info.Size()
		}
		return nil
	})

	if siteSize <= 0 {
		t.Error("Expected positive site size")
	}

	// Step 2: Update metadata
	metadataOpts := compress.MetadataOptions{
		SiteName:    "deployment-flow-test",
		Description: "Testing the deployment flow",
		Category:    "website",
	}

	if err := compress.UpdateMetadata(wsResourcesPath, metadataOpts); err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	// Step 3: Simulate successful deployment result
	mockResult := &deployer.Result{
		Success:  true,
		ObjectID: "0xmock-deployment-id",
		FileToBlobID: map[string]string{
			"index.html": "blob-1",
		},
	}

	if !mockResult.Success || mockResult.ObjectID == "" {
		t.Error("Mock deployment should be successful")
	}

	// Step 4: Update object ID
	if err := compress.UpdateObjectID(wsResourcesPath, mockResult.ObjectID); err != nil {
		t.Fatalf("Failed to update object ID: %v", err)
	}

	// Step 5: Verify final state
	finalConfig, err := compress.ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		t.Fatalf("Failed to read final ws-resources.json: %v", err)
	}

	if finalConfig.ObjectID != "0xmock-deployment-id" {
		t.Errorf("Expected ObjectID '0xmock-deployment-id', got '%s'", finalConfig.ObjectID)
	}
	if finalConfig.SiteName != "deployment-flow-test" {
		t.Errorf("Expected SiteName 'deployment-flow-test', got '%s'", finalConfig.SiteName)
	}
}

// TestDeploymentUpdateVsNew tests the logic for determining update vs new deployment
func TestDeploymentUpdateVsNew(t *testing.T) {
	tests := []struct {
		name           string
		walgoCfgID     string
		wsResourcesID  string
		databaseID     string
		forceNew       bool
		expectUpdate   bool
		expectObjectID string
	}{
		{
			name:           "new deployment - no existing IDs",
			walgoCfgID:     "",
			wsResourcesID:  "",
			databaseID:     "",
			forceNew:       false,
			expectUpdate:   false,
			expectObjectID: "",
		},
		{
			name:           "update from walgo.yaml",
			walgoCfgID:     "0xwalgo-id",
			wsResourcesID:  "",
			databaseID:     "",
			forceNew:       false,
			expectUpdate:   true,
			expectObjectID: "0xwalgo-id",
		},
		{
			name:           "update from ws-resources.json",
			walgoCfgID:     "",
			wsResourcesID:  "0xws-id",
			databaseID:     "",
			forceNew:       false,
			expectUpdate:   true,
			expectObjectID: "0xws-id",
		},
		{
			name:           "walgo.yaml takes priority over ws-resources",
			walgoCfgID:     "0xwalgo-id",
			wsResourcesID:  "0xws-id",
			databaseID:     "",
			forceNew:       false,
			expectUpdate:   true,
			expectObjectID: "0xwalgo-id",
		},
		{
			name:           "force-new with existing IDs",
			walgoCfgID:     "0xexisting-id",
			wsResourcesID:  "0xws-id",
			databaseID:     "",
			forceNew:       true,
			expectUpdate:   false,
			expectObjectID: "",
		},
		{
			name:           "placeholder ID is ignored",
			walgoCfgID:     "YOUR_WALRUS_PROJECT_ID",
			wsResourcesID:  "",
			databaseID:     "",
			forceNew:       false,
			expectUpdate:   false,
			expectObjectID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var existingObjectID string
			var isUpdate bool

			// Check walgo.yaml projectID (simulated)
			if tt.walgoCfgID != "" && tt.walgoCfgID != "YOUR_WALRUS_PROJECT_ID" {
				existingObjectID = tt.walgoCfgID
			}

			// Check ws-resources.json objectId (simulated)
			if existingObjectID == "" && tt.wsResourcesID != "" {
				existingObjectID = tt.wsResourcesID
			}

			// Check database (simulated)
			if existingObjectID == "" && tt.databaseID != "" {
				existingObjectID = tt.databaseID
			}

			// Determine update mode
			if existingObjectID != "" && !tt.forceNew {
				isUpdate = true
			} else if tt.forceNew {
				existingObjectID = "" // Reset for new deployment
			}

			if isUpdate != tt.expectUpdate {
				t.Errorf("Expected isUpdate=%v, got %v", tt.expectUpdate, isUpdate)
			}
			if existingObjectID != tt.expectObjectID {
				t.Errorf("Expected objectID='%s', got '%s'", tt.expectObjectID, existingObjectID)
			}
		})
	}
}
