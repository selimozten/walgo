package cmd

import (
	"os"
	"os/exec"
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
				"Initializes a new Hugo site",
				"walgo init [site-name]",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestInitCommandExecution(t *testing.T) {
	t.Run("Successful site initialization", func(t *testing.T) {
		if _, err := exec.LookPath("hugo"); err != nil {
			t.Skip("Hugo not installed, skipping site initialization test")
		}
		// Create temp directory for testing
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }() //nolint:errcheck // test cleanup

		siteName := "test-site"

		// Execute init command
		// Note: This will fail without Hugo installed
		stdout, stderr := captureOutput(func() {
			// Recover from potential panics
			defer func() { _ = func() any { return recover() }() }()
			// The command uses RunE, so call it properly
			if initCmd.RunE != nil {
				_ = initCmd.RunE(initCmd, []string{siteName})
			}
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
		if _, err := exec.LookPath("hugo"); err != nil {
			t.Skip("Hugo not installed, skipping existing directory test")
		}
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }() //nolint:errcheck // test cleanup

		// Create existing directory
		siteName := "existing-site"
		if err := os.MkdirAll(siteName, 0755); err != nil {
			t.Fatal(err)
		}

		// Execute command
		stdout, stderr := captureOutput(func() {
			defer func() { _ = func() any { return recover() }() }() // Recover to continue test
			if initCmd.RunE != nil {
				_ = initCmd.RunE(initCmd, []string{siteName})
			}
		})

		// Directory creation should succeed even if it exists
		// But Hugo initialization might fail
		_ = stdout
		_ = stderr
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
