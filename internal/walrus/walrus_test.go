package walrus

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"walgo/internal/config"
)

// Test utilities for mocking external dependencies
// Note: execLookPath and execCommand are defined in walrus.go for dependency injection

func TestDeploySite(t *testing.T) {
	// Save original runPreflight setting
	originalRunPreflight := runPreflight
	// Disable preflight checks for all tests
	runPreflight = false
	// Restore after all tests
	defer func() { runPreflight = originalRunPreflight }()

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
			expectedInArgs:   []string{"publish", "/path/to/public", "--epochs", "5"},
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
			expectedInArgs:   []string{"publish", "/path/to/public", "--epochs", "0"},
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
			originalOsStat := osStat

			// Mock osStat for config file checking
			osStat = func(name string) (os.FileInfo, error) {
				if strings.Contains(name, "sites-config.yaml") {
					if tt.configExists {
						// Return nil to indicate file exists
						return nil, nil
					}
					return nil, os.ErrNotExist
				}
				return originalOsStat(name)
			}

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
				osStat = originalOsStat
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
	// Disable preflight checks for all tests
	originalRunPreflight := runPreflight
	runPreflight = false
	defer func() { runPreflight = originalRunPreflight }()

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
			expectedInArgs:   []string{"update", "--epochs", "3", "/path/to/public", "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf"},
		},
		{
			name:             "Default epochs",
			deployDir:        "/path/to/public",
			objectID:         "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			epochs:           0, // Passes 0 directly (defaulting happens at command level)
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    false,
			expectedInArgs:   []string{"update", "--epochs", "0", "/path/to/public", "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf"},
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
			originalOsStat := osStat

			// Mock osStat for config file checking
			osStat = func(name string) (os.FileInfo, error) {
				if strings.Contains(name, "sites-config.yaml") {
					if tt.configExists {
						return nil, nil
					}
					return nil, os.ErrNotExist
				}
				return originalOsStat(name)
			}

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
				osStat = originalOsStat
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
			expectedInArgs:   []string{"sitemap", "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf"},
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
			originalOsStat := osStat

			// Mock osStat for config file checking
			osStat = func(name string) (os.FileInfo, error) {
				if strings.Contains(name, "sites-config.yaml") {
					if tt.configExists {
						return nil, nil
					}
					return nil, os.ErrNotExist
				}
				return originalOsStat(name)
			}

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
				osStat = originalOsStat
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
			expectedInArgs:   []string{"convert", "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf"},
		},
		{
			name:             "Short object ID",
			objectID:         "0x123abc",
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    false,
			expectedInArgs:   []string{"convert", "0x123abc"},
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
			originalOsStat := osStat

			// Mock osStat for config file checking
			osStat = func(name string) (os.FileInfo, error) {
				if strings.Contains(name, "sites-config.yaml") {
					if tt.configExists {
						return nil, nil
					}
					return nil, os.ErrNotExist
				}
				return originalOsStat(name)
			}

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
				osStat = originalOsStat
			}()

			output, err := ConvertObjectID(tt.objectID)

			if tt.expectedError && err == nil {
				t.Errorf("ConvertObjectID() expected error but got none")
			}
			if !tt.expectedError && err != nil && !strings.Contains(err.Error(), "failed to execute") {
				t.Errorf("ConvertObjectID() unexpected error: %v", err)
			}

			// For successful cases, output should not be empty
			if !tt.expectedError && err != nil && strings.Contains(err.Error(), "failed to execute") {
				// This is expected for mocked execution
				if output != "" {
					t.Errorf("ConvertObjectID() should return empty output on execution failure")
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
	// Disable preflight checks for all tests
	originalRunPreflight := runPreflight
	runPreflight = false
	defer func() { runPreflight = originalRunPreflight }()

	tests := []struct {
		name      string
		projectID string
		expectErr bool
	}{
		{
			name:      "Valid config",
			projectID: "test-project-id",
			expectErr: false,
		},
		{
			name:      "Empty project ID",
			projectID: "",
			expectErr: false, // Empty project ID is valid for new deployments
		},
		{
			name:      "Default project ID",
			projectID: "your-project-id",
			expectErr: false, // Default is also valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original functions
			originalLookPath := execLookPath
			originalCommand := execCommand
			originalOsStat := osStat

			// Always mock config as existing for this test
			osStat = func(name string) (os.FileInfo, error) {
				if strings.Contains(name, "sites-config.yaml") {
					return nil, nil // File exists
				}
				return originalOsStat(name)
			}

			// Mock site-builder to always exist
			execLookPath = func(file string) (string, error) {
				if file == siteBuilderCmd {
					return "/usr/bin/site-builder", nil
				}
				if file == "walrus" {
					return "/usr/bin/site-builder", nil
				}
				if file == "sui" {
					return "/usr/bin/site-builder", nil
				}
				return originalLookPath(file)
			}

			// Mock execCommand
			execCommand = func(name string, args ...string) *exec.Cmd {
				// Mock successful execution for info commands
				if strings.Contains(name, "walrus") && len(args) > 0 && args[0] == "info" {
					cmd := exec.Command("true") // Use true command for successful mock
					return cmd
				}
				if strings.Contains(name, "sui") {
					cmd := exec.Command("echo", `{"balance": 1000000}`)
					return cmd
				}
				// For site-builder commands, return a failing command
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Restore original functions after test
			defer func() {
				execLookPath = originalLookPath
				execCommand = originalCommand
				osStat = originalOsStat
			}()

			cfg := config.WalrusConfig{ProjectID: tt.projectID}
			_, err := DeploySite("/test", cfg, 1)

			// We expect the command to fail at execution (not validation)
			if err == nil {
				t.Errorf("Expected error due to mocked execution failure")
			} else if !strings.Contains(err.Error(), "failed to execute") {
				t.Errorf("Unexpected error for valid config: %v", err)
			}
		})
	}
}

func TestParseSiteBuilderOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected SiteBuilderOutput
	}{
		{
			name: "Parse deployment output",
			output: `
Publishing site to Walrus Sites...
Site published successfully
Object ID: 0x123abc
Browse at: https://walrus-sites.com/browse/123abc
`,
			expected: SiteBuilderOutput{
				ObjectID: "0x123abc",
				SiteURL:  "https://walrus-sites.com/browse/123abc",
				Success:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSiteBuilderOutput(tt.output)

			// Basic check for parsing
			if result == nil {
				t.Errorf("parseSiteBuilderOutput() returned nil")
			}
		})
	}
}