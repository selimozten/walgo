package walrus

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"walgo/internal/config"
)

// Test utilities for mocking external dependencies
// Note: execLookPath and execCommand are defined in walrus.go for dependency injection

func TestDeploySite(t *testing.T) {
	tests := []struct {
		name             string
		deployDir        string
		walrusCfg        config.WalrusConfig
		epochs           int
		siteBuilderFound bool
		configExists     bool
		expectedError    bool
		expectedInArgs   []string
	}{
		{
			name:      "Successful deployment setup",
			deployDir: "/path/to/public",
			walrusCfg: config.WalrusConfig{
				ProjectID: "test-project-id",
			},
			epochs:           5,
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    false,
			expectedInArgs:   []string{"publish", "/path/to/public", "--epochs", "5", "--context", "testnet"},
		},
		{
			name:      "Zero epochs",
			deployDir: "/path/to/public",
			walrusCfg: config.WalrusConfig{
				ProjectID: "test-project-id",
			},
			epochs:           0, // Passes 0 directly (defaulting happens at command level)
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    false,
			expectedInArgs:   []string{"publish", "/path/to/public", "--epochs", "0", "--context", "testnet"},
		},
		{
			name:      "site-builder not found",
			deployDir: "/path/to/public",
			walrusCfg: config.WalrusConfig{
				ProjectID: "test-project-id",
			},
			epochs:           5,
			siteBuilderFound: false,
			configExists:     true,
			expectedError:    true,
		},
		{
			name:      "config not found",
			deployDir: "/path/to/public",
			walrusCfg: config.WalrusConfig{
				ProjectID: "test-project-id",
			},
			epochs:           5,
			siteBuilderFound: true,
			configExists:     false,
			expectedError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original functions
			originalLookPath := execLookPath
			originalCommand := execCommand

			// Mock execLookPath
			execLookPath = func(file string) (string, error) {
				if file == siteBuilderCmd {
					if tt.siteBuilderFound {
						return "/usr/bin/site-builder", nil
					}
					return "", fmt.Errorf("executable file not found in $PATH")
				}
				return originalLookPath(file)
			}

			// Mock execCommand to capture arguments
			var capturedArgs []string
			execCommand = func(name string, args ...string) *exec.Cmd {
				capturedArgs = args
				// Return a command that will fail execution (since we're just testing setup)
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Restore original functions after test
			defer func() {
				execLookPath = originalLookPath
				execCommand = originalCommand
			}()

			output, err := DeploySite(tt.deployDir, tt.walrusCfg, tt.epochs)

			if tt.expectedError && err == nil {
				t.Errorf("DeploySite() expected error but got none")
			}
			if !tt.expectedError && err != nil && !strings.Contains(err.Error(), "failed to execute") {
				// Allow "failed to execute" errors since we can't mock the actual execution
				t.Errorf("DeploySite() unexpected error: %v", err)
			}

			// For successful cases, output should not be nil
			if !tt.expectedError && err != nil && strings.Contains(err.Error(), "failed to execute") {
				// This is expected for mocked execution
				if output != nil {
					t.Errorf("DeploySite() should return nil output on execution failure")
				}
			}

			// Check that the correct arguments were passed (for non-error cases where site-builder is found)
			if !tt.expectedError && tt.siteBuilderFound && tt.configExists && len(tt.expectedInArgs) > 0 {
				for _, expectedArg := range tt.expectedInArgs {
					found := false
					for _, capturedArg := range capturedArgs {
						if capturedArg == expectedArg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("DeploySite() missing expected argument: %s in %v", expectedArg, capturedArgs)
					}
				}
			}
		})
	}
}

func TestUpdateSite(t *testing.T) {
	tests := []struct {
		name             string
		deployDir        string
		objectID         string
		epochs           int
		siteBuilderFound bool
		configExists     bool
		expectedError    bool
		expectedInArgs   []string
	}{
		{
			name:             "Successful update setup",
			deployDir:        "/path/to/public",
			objectID:         "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			epochs:           3,
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    false,
			expectedInArgs:   []string{"update", "--epochs", "3", "/path/to/public", "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf", "--context", "testnet"},
		},
		{
			name:             "Default epochs",
			deployDir:        "/path/to/public",
			objectID:         "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			epochs:           0, // Passes 0 directly (defaulting happens at command level)
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    false,
			expectedInArgs:   []string{"update", "--epochs", "0", "/path/to/public", "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf", "--context", "testnet"},
		},
		{
			name:          "Empty object ID",
			deployDir:     "/path/to/public",
			objectID:      "",
			epochs:        3,
			expectedError: true,
		},
		{
			name:             "site-builder not found",
			deployDir:        "/path/to/public",
			objectID:         "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			epochs:           3,
			siteBuilderFound: false,
			configExists:     true,
			expectedError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original functions
			originalLookPath := execLookPath
			originalCommand := execCommand

			// Mock execLookPath
			execLookPath = func(file string) (string, error) {
				if file == siteBuilderCmd {
					if tt.siteBuilderFound {
						return "/usr/bin/site-builder", nil
					}
					return "", fmt.Errorf("executable file not found in $PATH")
				}
				return originalLookPath(file)
			}

			// Mock execCommand to capture arguments
			var capturedArgs []string
			execCommand = func(name string, args ...string) *exec.Cmd {
				capturedArgs = args
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Restore original functions after test
			defer func() {
				execLookPath = originalLookPath
				execCommand = originalCommand
			}()

			output, err := UpdateSite(tt.deployDir, tt.objectID, tt.epochs)

			if tt.expectedError && err == nil {
				t.Errorf("UpdateSite() expected error but got none")
			}
			if !tt.expectedError && err != nil && !strings.Contains(err.Error(), "failed to execute") {
				t.Errorf("UpdateSite() unexpected error: %v", err)
			}

			// For successful cases, output should not be nil
			if !tt.expectedError && err != nil && strings.Contains(err.Error(), "failed to execute") {
				// This is expected for mocked execution
				if output != nil {
					t.Errorf("UpdateSite() should return nil output on execution failure")
				}
			}

			// Check arguments for non-error cases where site-builder is found
			if !tt.expectedError && tt.siteBuilderFound && tt.configExists && len(tt.expectedInArgs) > 0 {
				for _, expectedArg := range tt.expectedInArgs {
					found := false
					for _, capturedArg := range capturedArgs {
						if capturedArg == expectedArg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("UpdateSite() missing expected argument: %s in %v", expectedArg, capturedArgs)
					}
				}
			}
		})
	}
}

func TestGetSiteStatus(t *testing.T) {
	tests := []struct {
		name             string
		objectID         string
		siteBuilderFound bool
		configExists     bool
		expectedError    bool
		expectedInArgs   []string
	}{
		{
			name:             "Valid object ID",
			objectID:         "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    false,
			expectedInArgs:   []string{"sitemap", "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf", "--context", "testnet"},
		},
		{
			name:          "Empty object ID",
			objectID:      "",
			expectedError: true,
		},
		{
			name:             "site-builder not found",
			objectID:         "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			siteBuilderFound: false,
			configExists:     true,
			expectedError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original functions
			originalLookPath := execLookPath
			originalCommand := execCommand

			// Mock execLookPath
			execLookPath = func(file string) (string, error) {
				if file == siteBuilderCmd {
					if tt.siteBuilderFound {
						return "/usr/bin/site-builder", nil
					}
					return "", fmt.Errorf("executable file not found in $PATH")
				}
				return originalLookPath(file)
			}

			// Mock execCommand to capture arguments
			var capturedArgs []string
			execCommand = func(name string, args ...string) *exec.Cmd {
				capturedArgs = args
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Restore original functions after test
			defer func() {
				execLookPath = originalLookPath
				execCommand = originalCommand
			}()

			output, err := GetSiteStatus(tt.objectID)

			if tt.expectedError && err == nil {
				t.Errorf("GetSiteStatus() expected error but got none")
			}
			if !tt.expectedError && err != nil && !strings.Contains(err.Error(), "failed to execute") {
				t.Errorf("GetSiteStatus() unexpected error: %v", err)
			}

			// For successful cases, output should not be nil
			if !tt.expectedError && err != nil && strings.Contains(err.Error(), "failed to execute") {
				// This is expected for mocked execution
				if output != nil {
					t.Errorf("GetSiteStatus() should return nil output on execution failure")
				}
			}

			// Check arguments for non-error cases where site-builder is found
			if !tt.expectedError && tt.siteBuilderFound && tt.configExists && len(tt.expectedInArgs) > 0 {
				for _, expectedArg := range tt.expectedInArgs {
					found := false
					for _, capturedArg := range capturedArgs {
						if capturedArg == expectedArg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("GetSiteStatus() missing expected argument: %s in %v", expectedArg, capturedArgs)
					}
				}
			}
		})
	}
}

func TestConvertObjectID(t *testing.T) {
	tests := []struct {
		name             string
		objectID         string
		siteBuilderFound bool
		configExists     bool
		expectedError    bool
		expectedInArgs   []string
	}{
		{
			name:             "Valid object ID",
			objectID:         "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    false,
			expectedInArgs:   []string{"convert", "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf", "--context", "testnet"},
		},
		{
			name:             "Short object ID",
			objectID:         "0x123abc",
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    false,
			expectedInArgs:   []string{"convert", "0x123abc", "--context", "testnet"},
		},
		{
			name:          "Empty object ID",
			objectID:      "",
			expectedError: true,
		},
		{
			name:             "site-builder not found",
			objectID:         "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			siteBuilderFound: false,
			configExists:     true,
			expectedError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original functions
			originalLookPath := execLookPath
			originalCommand := execCommand

			// Mock execLookPath
			execLookPath = func(file string) (string, error) {
				if file == siteBuilderCmd {
					if tt.siteBuilderFound {
						return "/usr/bin/site-builder", nil
					}
					return "", fmt.Errorf("executable file not found in $PATH")
				}
				return originalLookPath(file)
			}

			// Mock execCommand to capture arguments
			var capturedArgs []string
			execCommand = func(name string, args ...string) *exec.Cmd {
				capturedArgs = args
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Restore original functions after test
			defer func() {
				execLookPath = originalLookPath
				execCommand = originalCommand
			}()

			base36ID, err := ConvertObjectID(tt.objectID)

			if tt.expectedError && err == nil {
				t.Errorf("ConvertObjectID() expected error but got none")
			}
			if !tt.expectedError && err != nil && !strings.Contains(err.Error(), "failed to execute") {
				t.Errorf("ConvertObjectID() unexpected error: %v", err)
			}

			// For successful cases, we might get an empty base36ID due to mocked execution
			if !tt.expectedError && err != nil && strings.Contains(err.Error(), "failed to execute") {
				// This is expected for mocked execution
				if base36ID != "" {
					t.Errorf("ConvertObjectID() should return empty string on execution failure")
				}
			}

			// Check arguments for non-error cases where site-builder is found
			if !tt.expectedError && tt.siteBuilderFound && tt.configExists && len(tt.expectedInArgs) > 0 {
				for _, expectedArg := range tt.expectedInArgs {
					found := false
					for _, capturedArg := range capturedArgs {
						if capturedArg == expectedArg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("ConvertObjectID() missing expected argument: %s in %v", expectedArg, capturedArgs)
					}
				}
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    config.WalrusConfig
		expectErr bool
	}{
		{
			name:      "Valid config",
			config:    config.WalrusConfig{ProjectID: "valid-project-id"},
			expectErr: false,
		},
		{
			name:      "Empty project ID",
			config:    config.WalrusConfig{ProjectID: ""},
			expectErr: false, // For new deployments, empty ProjectID is now allowed
		},
		{
			name:      "Default project ID",
			config:    config.WalrusConfig{ProjectID: "YOUR_WALRUS_PROJECT_ID"},
			expectErr: false, // For new deployments, this is also allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock successful site-builder lookup and config
			originalLookPath := execLookPath
			execLookPath = func(file string) (string, error) {
				return "/usr/bin/site-builder", nil
			}
			defer func() { execLookPath = originalLookPath }()

			_, err := DeploySite("/test", tt.config, 1)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for config %+v but got none", tt.config)
			}
			if !tt.expectErr && err != nil && !strings.Contains(err.Error(), "failed to execute") && !strings.Contains(err.Error(), "site-builder setup issue") {
				t.Errorf("Unexpected error for valid config: %v", err)
			}
		})
	}
}

// Test output parsing functions
func TestParseSiteBuilderOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected SiteBuilderOutput
	}{
		{
			name: "Parse deployment output",
			output: `Created new site: test site
New site object ID: 0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf
To browse the site, you have the following options:
        1. Run a local portal, and browse the site through it: e.g. http://5qs1ypn4wn90d6mv7d7dkwvvl49hdrlpqulr11ngpykoifycwf.localhost:3000
        2. Use a third-party portal (e.g. wal.app), which will require a SuiNS name.`,
			expected: SiteBuilderOutput{
				ObjectID: "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
				BrowseURLs: []string{
					"http://5qs1ypn4wn90d6mv7d7dkwvvl49hdrlpqulr11ngpykoifycwf.localhost:3000",
				},
				Resources: []Resource{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSiteBuilderOutput(tt.output)

			if result.ObjectID != tt.expected.ObjectID {
				t.Errorf("parseSiteBuilderOutput() ObjectID = %v, want %v", result.ObjectID, tt.expected.ObjectID)
			}

			if len(result.BrowseURLs) != len(tt.expected.BrowseURLs) {
				t.Errorf("parseSiteBuilderOutput() BrowseURLs length = %v, want %v", len(result.BrowseURLs), len(tt.expected.BrowseURLs))
			}
		})
	}
}

// Benchmark tests to ensure performance
func BenchmarkDeploySite(b *testing.B) {
	cfg := config.WalrusConfig{ProjectID: "test-id"}

	// Mock exec.LookPath for benchmarking
	originalLookPath := execLookPath
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/site-builder", nil
	}
	defer func() { execLookPath = originalLookPath }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail at execution but we're testing the setup performance
		DeploySite("/test", cfg, 1)
	}
}
