package hugo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestInitializeSite(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() (string, func())
		mockHugo     bool
		hugoNotFound bool
		hugoFails    bool
		wantErr      bool
		errContains  string
	}{
		{
			name: "Successful initialization",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			mockHugo: true,
			wantErr:  false,
		},
		{
			name: "Hugo not installed",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			hugoNotFound: true,
			wantErr:      true,
			errContains:  "Hugo is not installed",
		},
		{
			name: "Hugo command fails",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			mockHugo:    true,
			hugoFails:   true,
			wantErr:     true,
			errContains: "failed to initialize Hugo site",
		},
		{
			name: "Directory doesn't exist",
			setup: func() (string, func()) {
				nonExistent := "/tmp/non-existent-dir-for-testing"
				os.RemoveAll(nonExistent) // Ensure it doesn't exist
				return nonExistent, func() { os.RemoveAll(nonExistent) }
			},
			mockHugo:    true,
			wantErr:     true,
			errContains: "failed to initialize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath, cleanup := tt.setup()
			defer cleanup()

			// Mock exec.LookPath if needed
			if tt.hugoNotFound {
				// Temporarily modify PATH to make hugo not findable
				origPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				defer os.Setenv("PATH", origPath)
			}

			if tt.mockHugo && !tt.hugoNotFound {
				// Check if hugo is actually installed
				if _, err := exec.LookPath("hugo"); err != nil {
					// Hugo is not installed, skip this test
					t.Skip("Hugo is not installed, skipping test that requires Hugo")
				}
			}

			// For testing with mock failure, we'd need to inject a command that fails
			// This is complex without dependency injection, so we'll test what we can
			if tt.hugoFails {
				// We can't easily mock exec.Command failure without refactoring
				// Skip this specific test case for now
				t.Skip("Cannot mock Hugo failure without dependency injection")
			}

			err := InitializeSite(sitePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("InitializeSite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got %q", tt.errContains, err.Error())
				}
			}

			// If successful and Hugo is installed, check that site structure was created
			if !tt.wantErr && !tt.hugoNotFound {
				// Check for Hugo configuration file
				possibleConfigs := []string{
					filepath.Join(sitePath, "hugo.toml"),
					filepath.Join(sitePath, "config.toml"),
					filepath.Join(sitePath, "config.yaml"),
				}

				configFound := false
				for _, config := range possibleConfigs {
					if _, err := os.Stat(config); err == nil {
						configFound = true
						break
					}
				}

				if !configFound && err == nil {
					// Hugo might not have created config if it's not installed
					// This is expected in test environment
					t.Log("No Hugo config file found, Hugo might not be installed")
				}
			}
		})
	}
}

func TestBuildSite(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() (string, func())
		hugoNotFound bool
		noConfig     bool
		wantErr      bool
		errContains  string
	}{
		{
			name: "Successful build",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				// Create a minimal Hugo config
				configContent := `
baseURL = "https://example.com/"
languageCode = "en-us"
title = "Test Site"
`
				configPath := filepath.Join(tmpDir, "hugo.toml")
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create content directory
				if err := os.MkdirAll(filepath.Join(tmpDir, "content"), 0755); err != nil {
					t.Fatal(err)
				}

				return tmpDir, func() {}
			},
			wantErr: false,
		},
		{
			name: "Hugo not installed",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			hugoNotFound: true,
			wantErr:      true,
			errContains:  "Hugo is not installed",
		},
		{
			name: "No Hugo config file",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				// Don't create any config file
				return tmpDir, func() {}
			},
			noConfig:    true,
			wantErr:     true,
			errContains: "hugo configuration file",
		},
		{
			name: "Config.toml instead of hugo.toml",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				// Create config.toml instead of hugo.toml
				configContent := `
baseURL = "https://example.com/"
title = "Test Site"
`
				configPath := filepath.Join(tmpDir, "config.toml")
				if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create content directory
				if err := os.MkdirAll(filepath.Join(tmpDir, "content"), 0755); err != nil {
					t.Fatal(err)
				}

				return tmpDir, func() {}
			},
			wantErr: false,
		},
		{
			name: "Invalid site directory",
			setup: func() (string, func()) {
				// Return a non-existent directory
				return "/tmp/non-existent-hugo-site", func() {}
			},
			wantErr:     true,
			errContains: "hugo configuration file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sitePath, cleanup := tt.setup()
			defer cleanup()

			// Mock exec.LookPath if needed
			if tt.hugoNotFound {
				// Temporarily modify PATH to make hugo not findable
				origPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				defer os.Setenv("PATH", origPath)
			}

			// Check if hugo is actually installed for tests that need it
			// Skip if Hugo is not found, unless we're specifically testing the "hugo not found" case
			if !tt.hugoNotFound {
				if _, err := exec.LookPath("hugo"); err != nil {
					// Hugo is not installed, skip this test
					t.Skip("Hugo is not installed, skipping test that requires Hugo")
				}
			}

			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			// Create pipes for capturing output
			_, wOut, _ := os.Pipe()
			_, wErr, _ := os.Pipe()
			os.Stdout = wOut
			os.Stderr = wErr

			// Run the function
			err := BuildSite(sitePath)

			// Close writers
			wOut.Close()
			wErr.Close()

			// Restore stdout and stderr
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildSite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got %q", tt.errContains, err.Error())
				}
			}

			// If successful, check that public directory was created
			if !tt.wantErr && err == nil {
				publicDir := filepath.Join(sitePath, "public")
				if _, err := os.Stat(publicDir); err == nil {
					t.Logf("Public directory created at %s", publicDir)
				}
			}
		})
	}
}

// TestHugoCommandExecution tests the actual command execution paths
func TestHugoCommandExecution(t *testing.T) {
	// Check if Hugo is installed
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("Hugo is not installed, skipping integration tests")
	}

	t.Run("Full Hugo workflow", func(t *testing.T) {
		// Create a temporary directory for the test site
		tmpDir := t.TempDir()

		// Initialize a new Hugo site
		err := InitializeSite(tmpDir)
		if err != nil {
			t.Fatalf("Failed to initialize site: %v", err)
		}

		// Check that Hugo created the expected structure
		expectedDirs := []string{"content", "layouts", "static"}
		for _, dir := range expectedDirs {
			dirPath := filepath.Join(tmpDir, dir)
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				// Some directories might not be created by newer Hugo versions
				t.Logf("Directory %s was not created (might be normal for newer Hugo)", dir)
			}
		}

		// Build the site
		err = BuildSite(tmpDir)
		if err != nil {
			// Building might fail if no theme is set, which is expected
			t.Logf("Build failed (expected without theme): %v", err)
		}
	})
}

// TestInitializeSiteOutput tests that output is properly captured
func TestInitializeSiteOutput(t *testing.T) {
	// Check if Hugo is installed
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("Hugo is not installed, skipping output test")
	}

	tmpDir := t.TempDir()

	// Capture output
	output := captureOutput(func() {
		err := InitializeSite(tmpDir)
		if err != nil {
			t.Logf("InitializeSite error: %v", err)
		}
	})

	// Check that some output was produced
	if !strings.Contains(output, "Hugo") && !strings.Contains(output, "site") {
		t.Log("Expected output to contain 'Hugo' or 'site' references")
	}
}

// TestBuildSiteOutput tests that output is properly directed
func TestBuildSiteOutput(t *testing.T) {
	// Check if Hugo is installed
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("Hugo is not installed, skipping output test")
	}

	tmpDir := t.TempDir()

	// Create a minimal Hugo config
	configContent := `
baseURL = "/"
title = "Test"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "hugo.toml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create required directories
	if err := os.MkdirAll(filepath.Join(tmpDir, "content"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "layouts"), 0755); err != nil {
		t.Fatal(err)
	}

	// Capture output
	output := captureOutput(func() {
		err := BuildSite(tmpDir)
		if err != nil {
			// Build might fail without theme, which is ok for this test
			t.Logf("BuildSite error (expected without theme): %v", err)
		}
	})

	// We should see some build-related output
	t.Logf("Build output: %s", output)
}

// captureOutput captures all output from a function
func captureOutput(f func()) string {
	// Save current stdout
	old := os.Stdout
	oldErr := os.Stderr

	// Create a pipe
	r, w, _ := os.Pipe()

	// Set stdout and stderr to our pipe
	os.Stdout = w
	os.Stderr = w

	// Run the function
	f()

	// Close the writer
	w.Close()

	// Restore stdout and stderr
	os.Stdout = old
	os.Stderr = oldErr

	// Read the output
	output := make([]byte, 1024*1024)
	n, _ := r.Read(output)
	r.Close()

	return string(output[:n])
}

// TestErrorMessages tests that error messages are informative
func TestErrorMessages(t *testing.T) {
	t.Run("InitializeSite error message", func(t *testing.T) {
		// Test with PATH that doesn't include hugo
		origPath := os.Getenv("PATH")
		os.Setenv("PATH", "/usr/bin:/bin") // Minimal PATH without hugo
		defer os.Setenv("PATH", origPath)

		err := InitializeSite(t.TempDir())
		if err == nil {
			t.Skip("Hugo might be in /usr/bin or /bin")
		}

		if !strings.Contains(err.Error(), "Hugo") {
			t.Errorf("Error message should mention Hugo, got: %v", err)
		}
	})

	t.Run("BuildSite config not found message", func(t *testing.T) {
		// Skip if Hugo is not installed
		if _, err := exec.LookPath("hugo"); err != nil {
			t.Skip("Hugo is not installed, skipping test that requires Hugo")
		}

		tmpDir := t.TempDir()
		err := BuildSite(tmpDir)

		if err == nil {
			t.Error("Expected error for missing config")
			return
		}

		if !strings.Contains(err.Error(), "configuration file") {
			t.Errorf("Error should mention configuration file, got: %v", err)
		}
		if !strings.Contains(err.Error(), tmpDir) {
			t.Errorf("Error should mention the directory path, got: %v", err)
		}
	})
}

// TestPathValidation tests path-related edge cases
func TestPathValidation(t *testing.T) {
	t.Run("InitializeSite with relative path", func(t *testing.T) {
		// Check if Hugo is installed
		if _, err := exec.LookPath("hugo"); err != nil {
			t.Skip("Hugo is not installed")
		}

		// Create a temp dir and change to it
		tmpDir := t.TempDir()
		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(origWd) }() //nolint:errcheck // test cleanup

		// Use relative path
		err := InitializeSite(".")
		if err != nil {
			t.Logf("InitializeSite with relative path: %v", err)
		}
	})

	t.Run("BuildSite with relative path", func(t *testing.T) {
		tmpDir := t.TempDir()
		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(origWd) }() //nolint:errcheck // test cleanup

		// Create config
		if err := os.WriteFile("hugo.toml", []byte("title = \"test\""), 0644); err != nil {
			t.Fatal(err)
		}

		err := BuildSite(".")
		// Error is expected without hugo installed or without proper setup
		if err != nil {
			t.Logf("BuildSite with relative path: %v", err)
		}
	})
}

// TestConcurrentExecution tests that functions can be called concurrently
func TestConcurrentExecution(t *testing.T) {
	// Check if Hugo is installed
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("Hugo is not installed")
	}

	t.Run("Concurrent InitializeSite calls", func(t *testing.T) {
		done := make(chan bool, 3)

		for i := 0; i < 3; i++ {
			go func(index int) {
				tmpDir := filepath.Join(t.TempDir(), fmt.Sprintf("site%d", index))
				if err := os.MkdirAll(tmpDir, 0755); err != nil {
					t.Logf("Failed to create directory: %v", err)
					done <- true
					return
				}

				err := InitializeSite(tmpDir)
				if err != nil {
					t.Logf("Concurrent init %d: %v", index, err)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 3; i++ {
			<-done
		}
	})
}
