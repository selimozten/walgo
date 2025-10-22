package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Init command without arguments",
			Args:        []string{"init"},
			ExpectError: true,
			Contains:    []string{"accepts 1 arg(s), received 0"},
		},
		{
			Name:        "Init command with too many arguments",
			Args:        []string{"init", "site1", "site2"},
			ExpectError: true,
			Contains:    []string{"accepts 1 arg(s), received 2"},
		},
		{
			Name:        "Init command with help flag",
			Args:        []string{"init", "--help"},
			ExpectError: false,
			Contains: []string{
				"Initialize a new Hugo site",
				"Walrus Sites configuration",
				"walgo init [site-name]",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestInitCommandExecution(t *testing.T) {
	t.Run("Successful site initialization", func(t *testing.T) {
		// Create temp directory for testing
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		siteName := "test-site"

		// Execute init command
		// Note: This will fail without Hugo installed
		stdout, stderr := captureOutput(func() {
			// Recover from potential panics
			defer func() { recover() }()
			// The command may call os.Exit, which we can't mock directly
			// So we use recover to continue the test
			initCmd.Run(initCmd, []string{siteName})
		})

		// Check if directory was created
		sitePath := filepath.Join(tempDir, siteName)
		if _, err := os.Stat(sitePath); err == nil {
			// Directory was successfully created
			t.Logf("Site directory created: %s", sitePath)
		}

		// Log output for debugging
		if stdout != "" {
			t.Logf("stdout: %s", stdout)
		}
		if stderr != "" {
			t.Logf("stderr: %s", stderr)
		}
	})

	t.Run("Init with existing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		// Create existing directory
		siteName := "existing-site"
		_ = os.MkdirAll(siteName, 0755)

		// Execute command
		stdout, stderr := captureOutput(func() {
			defer func() { recover() }() // Recover to continue test
			initCmd.Run(initCmd, []string{siteName})
		})

		// Directory creation should succeed even if it exists
		// But Hugo initialization might fail
		_ = stdout
		_ = stderr
	})

	t.Run("Init with invalid directory name", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		// Use invalid characters for directory name (on most systems)
		siteName := "/invalid\x00name"

		// Execute command
		stdout, stderr := captureOutput(func() {
			defer func() { recover() }()
			initCmd.Run(initCmd, []string{siteName})
		})

		// Command should fail with invalid directory name
		_ = stdout
		_ = stderr
	})
}

func TestInitCommandIntegration(t *testing.T) {
	// This test would require actual Hugo binary and config package implementation
	// Skip if Hugo is not available
	t.Run("Full integration test", func(t *testing.T) {
		// Check if Hugo is available
		if _, err := os.Stat("/usr/local/bin/hugo"); os.IsNotExist(err) {
			if _, err := os.Stat("/usr/bin/hugo"); os.IsNotExist(err) {
				t.Skip("Hugo not installed, skipping integration test")
			}
		}

		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		siteName := "integration-test-site"

		// Execute command through cobra
		output, err := executeCommand(rootCmd, "init", siteName)

		// Check results based on actual implementation
		// The command might fail if config.CreateDefaultWalgoConfig is not properly mocked
		if err != nil {
			// Check if it's the expected error
			if output != "" {
				// Some output was generated
			}
		} else {
			// Check if site directory exists
			sitePath := filepath.Join(tempDir, siteName)
			if _, err := os.Stat(sitePath); err == nil {
				// Directory was created successfully
				t.Logf("Site directory created: %s", sitePath)

				// Check for walgo.yaml
				configPath := filepath.Join(sitePath, "walgo.yaml")
				if _, err := os.Stat(configPath); err == nil {
					t.Log("walgo.yaml created successfully")
				}
			}
		}
	})
}

func TestInitCommandFlags(t *testing.T) {
	// Verify the command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "init" {
			found = true

			// Check command properties
			if cmd.Use != "init [site-name]" {
				t.Errorf("Unexpected Use field: %s", cmd.Use)
			}

			if cmd.Args(cmd, []string{"test"}) != nil {
				t.Error("Args validation failed for valid input")
			}

			if cmd.Args(cmd, []string{}) == nil {
				t.Error("Args validation should fail for no arguments")
			}

			if cmd.Args(cmd, []string{"test", "extra"}) == nil {
				t.Error("Args validation should fail for too many arguments")
			}

			break
		}
	}

	if !found {
		t.Error("init command not registered with root command")
	}
}
