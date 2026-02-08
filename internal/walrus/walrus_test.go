package walrus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/selimozten/walgo/internal/config"
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
			expectedInArgs:   []string{"--walrus-binary", "deploy", "/path/to/public", "--epochs", "5"},
		},
		{
			name:      "Zero epochs - should fail validation",
			deployDir: "/path/to/public",
			walrusCfg: config.WalrusConfig{
				ProjectID: "test-project-id",
			},
			epochs:           0, // Zero epochs should be rejected
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    true, // Epochs must be > 0
			expectedInArgs:   []string{},
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
			originalCommandContext := execCommandContext
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
				if file == "walrus" {
					// Mock walrus binary location
					return "/usr/bin/walrus", nil
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

			// Mock execCommandContext (used by runCommandWithTimeout)
			execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
				capturedArgs = args
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Restore original functions after test
			defer func() {
				execLookPath = originalLookPath
				execCommand = originalCommand
				execCommandContext = originalCommandContext
				osStat = originalOsStat
			}()

			output, err := DeploySite(context.Background(), tt.deployDir, tt.walrusCfg, tt.epochs)

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
			expectedInArgs:   []string{"--walrus-binary", "deploy", "--epochs", "3", "/path/to/public"},
		},
		{
			name:             "Zero epochs - should fail validation",
			deployDir:        "/path/to/public",
			objectID:         "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			epochs:           0, // Zero epochs should be rejected
			siteBuilderFound: true,
			configExists:     true,
			expectedError:    true, // Epochs must be > 0
			expectedInArgs:   []string{},
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
			originalCommandContext := execCommandContext
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
				if file == "walrus" {
					// Mock walrus binary location
					return "/usr/bin/walrus", nil
				}
				return originalLookPath(file)
			}

			// Mock execCommand to capture arguments
			var capturedArgs []string
			execCommand = func(name string, args ...string) *exec.Cmd {
				capturedArgs = args
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Mock execCommandContext (used by runCommandWithTimeout)
			execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
				capturedArgs = args
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Restore original functions after test
			defer func() {
				execLookPath = originalLookPath
				execCommand = originalCommand
				execCommandContext = originalCommandContext
				osStat = originalOsStat
			}()

			output, err := UpdateSite(context.Background(), tt.deployDir, tt.objectID, tt.epochs)

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
			expectedInArgs:   []string{"--walrus-binary", "sitemap", "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf"},
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
			originalCommandContext := execCommandContext
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
				if file == "walrus" {
					// Mock walrus binary location
					return "/usr/bin/walrus", nil
				}
				return originalLookPath(file)
			}

			// Mock execCommand to capture arguments
			var capturedArgs []string
			execCommand = func(name string, args ...string) *exec.Cmd {
				capturedArgs = args
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Mock execCommandContext (used by runCommandWithTimeout)
			execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
				capturedArgs = args
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Restore original functions after test
			defer func() {
				execLookPath = originalLookPath
				execCommand = originalCommand
				execCommandContext = originalCommandContext
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

func TestConfigValidation(t *testing.T) {

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
			originalCommandContext := execCommandContext
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

			// Mock execCommandContext (used by runCommandWithTimeout)
			execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
				return &exec.Cmd{Path: name, Args: append([]string{name}, args...)}
			}

			// Restore original functions after test
			defer func() {
				execLookPath = originalLookPath
				execCommand = originalCommand
				execCommandContext = originalCommandContext
				osStat = originalOsStat
			}()

			cfg := config.WalrusConfig{ProjectID: tt.projectID}
			_, err := DeploySite(context.Background(), "/test", cfg, 1)

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
	// Sample 64-char hex object ID for testing
	testObjectID := "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf"

	tests := []struct {
		name         string
		output       string
		expectedID   string
		expectedURLs int
		shouldFindID bool
	}{
		{
			name: "Current format - New site object ID",
			output: `
Publishing site to Walrus Sites...
Site published successfully
New site object ID: ` + testObjectID + `
Browse at: https://walrus-sites.com/browse/123abc
`,
			expectedID:   testObjectID,
			expectedURLs: 1,
			shouldFindID: true,
		},
		{
			name: "Alternative format - Site object ID",
			output: `
Updating site...
Site object ID: ` + testObjectID + `
`,
			expectedID:   testObjectID,
			expectedURLs: 0,
			shouldFindID: true,
		},
		{
			name: "Generic Object ID pattern - should NOT match (too broad)",
			output: `
Site created successfully
Object ID = ` + testObjectID + `
`,
			expectedID:   "",
			expectedURLs: 0,
			shouldFindID: false, // We only match strict "site object ID" patterns now
		},
		{
			name: "lowercase site object ID",
			output: `
Deployment complete
site object ID: ` + testObjectID + `
`,
			expectedID:   testObjectID,
			expectedURLs: 0,
			shouldFindID: true, // This is a supported pattern
		},
		{
			name: "Generic Created site pattern - should NOT match (too broad)",
			output: `
Created site with ID ` + testObjectID + `
`,
			expectedID:   "",
			expectedURLs: 0,
			shouldFindID: false, // We only match strict patterns now
		},
		{
			name: "Multiple URLs",
			output: `
Site object ID: ` + testObjectID + `
Browse at: https://example1.walrus.site
Also available at: https://example2.walrus.site
`,
			expectedID:   testObjectID,
			expectedURLs: 2,
			shouldFindID: true,
		},
		{
			name:         "No object ID in output",
			output:       "Some random output without any ID",
			expectedID:   "",
			expectedURLs: 0,
			shouldFindID: false,
		},
		{
			name: "Multiple IDs - should pick site object ID not blob ID",
			output: `Uploading blob to Walrus...
Created object 0xf99aee9f21493e1590e7e5a9aea6f343a1f381031a04a732724871fc294be799
Processing site resources...
Created new site!
New site object ID: ` + testObjectID + `
Browse your site at http://example.localhost:3000`,
			expectedID:   testObjectID,
			expectedURLs: 1,
			shouldFindID: true,
		},
		{
			name:         "Generic site keyword without explicit pattern - should NOT match",
			output:       `The site has been deployed to 0xaaaabbbbccccdddd1111222233334444aaaabbbbccccdddd1111222233334444`,
			expectedID:   "",
			expectedURLs: 0,
			shouldFindID: false, // Only strict patterns should match
		},
		{
			name: "URL with trailing punctuation",
			output: `
Site object ID: ` + testObjectID + `
Visit: https://example.walrus.site.
More info at: https://docs.walrus.site,
See also: https://help.walrus.site;
`,
			expectedID:   testObjectID,
			expectedURLs: 3,
			shouldFindID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSiteBuilderOutput(tt.output)

			if result == nil {
				t.Fatalf("parseSiteBuilderOutput() returned nil")
			}

			if tt.shouldFindID {
				if result.ObjectID != tt.expectedID {
					t.Errorf("ObjectID = %q, want %q", result.ObjectID, tt.expectedID)
				}
			} else {
				if result.ObjectID != "" {
					t.Errorf("ObjectID should be empty, got %q", result.ObjectID)
				}
			}

			if len(result.BrowseURLs) != tt.expectedURLs {
				t.Errorf("BrowseURLs count = %d, want %d. Got: %v", len(result.BrowseURLs), tt.expectedURLs, result.BrowseURLs)
			}

			// Verify URLs don't have trailing punctuation
			for _, url := range result.BrowseURLs {
				if strings.HasSuffix(url, ".") || strings.HasSuffix(url, ",") || strings.HasSuffix(url, ";") || strings.HasSuffix(url, ":") {
					t.Errorf("URL has trailing punctuation: %q", url)
				}
			}
		})
	}
}

func TestValidateObjectID(t *testing.T) {
	tests := []struct {
		name      string
		objectID  string
		expectErr bool
	}{
		{
			name:      "Valid 64-char hex ID",
			objectID:  "0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			expectErr: false,
		},
		{
			name:      "Valid without 0x prefix",
			objectID:  "e674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf",
			expectErr: false,
		},
		{
			name:      "Empty ID",
			objectID:  "",
			expectErr: true,
		},
		{
			name:      "Invalid chars - command injection attempt",
			objectID:  "0x123; rm -rf /",
			expectErr: true,
		},
		{
			name:      "Invalid chars - pipe",
			objectID:  "0x123|cat /etc/passwd",
			expectErr: true,
		},
		{
			name:      "Invalid chars - newline",
			objectID:  "0x123\nmalicious",
			expectErr: true,
		},
		{
			name:      "Invalid - non-hex chars",
			objectID:  "0xGHIJKL",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateObjectID(tt.objectID)
			if tt.expectErr && err == nil {
				t.Errorf("validateObjectID(%q) expected error but got none", tt.objectID)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("validateObjectID(%q) unexpected error: %v", tt.objectID, err)
			}
		})
	}
}
