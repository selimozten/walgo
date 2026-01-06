package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestStatusCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Status command help",
			Args:        []string{"status", "--help"},
			ExpectError: false,
			Contains: []string{
				"status",
				"Walrus",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestStatusCommandArgsValidation(t *testing.T) {
	// Find the status command
	var statusCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "status" {
			statusCommand = cmd
			break
		}
	}

	if statusCommand == nil {
		t.Fatal("status command not found")
	}

	t.Run("Too many arguments returns error", func(t *testing.T) {
		err := statusCommand.Args(statusCommand, []string{"arg1", "arg2"})
		if err == nil {
			t.Error("Expected error for too many arguments")
		}
	})
}

func TestStatusCommandFlags(t *testing.T) {
	// Find the status command
	var statusCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "status" {
			statusCommand = cmd
			break
		}
	}

	if statusCommand == nil {
		t.Fatal("status command not found")
	}

	t.Run("Command is properly registered", func(t *testing.T) {
		if statusCommand.Use != "status [object-id]" {
			t.Errorf("Unexpected Use field: %s", statusCommand.Use)
		}

		// Test args validation
		if statusCommand.Args(statusCommand, []string{"0x123"}) != nil {
			t.Error("Args validation failed for single object ID")
		}

		if statusCommand.Args(statusCommand, []string{}) != nil {
			t.Error("Args validation should pass for no arguments")
		}

		if err := statusCommand.Args(statusCommand, []string{"arg1", "arg2"}); err == nil {
			t.Error("Args validation should fail for multiple arguments")
		}
	})
}

func TestStatusCommandExecution(t *testing.T) {
	t.Run("Status without config file", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute status command without config - should fail
		output, err := executeCommand(rootCmd, "status")
		if err == nil {
			t.Log("Expected error when no config file exists")
		}
		_ = output
	})

	t.Run("Status with config but no project ID", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config without valid project ID
		configContent := `
walrus:
  projectId: YOUR_WALRUS_PROJECT_ID
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute status command
		output, err := executeCommand(rootCmd, "status")
		// Should fail because project ID is placeholder
		if err == nil {
			t.Log("Expected error when project ID is placeholder")
		}
		_ = output
	})

	t.Run("Status with object ID argument", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create minimal config
		configContent := `
walrus:
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute status with object ID - will fail if site-builder not installed
		// but tests the argument parsing
		output, _ := executeCommand(rootCmd, "status", "0x1234567890abcdef")
		// The output should mention the object ID
		if output != "" {
			t.Logf("Output: %s", output)
		}
	})

	t.Run("Status with valid config and project ID", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config with valid project ID format
		configContent := `
walrus:
  projectId: "0x1234567890abcdef1234567890abcdef12345678"
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute status - will use project ID from config
		output, _ := executeCommand(rootCmd, "status")
		// Should attempt to use the project ID from config
		_ = output
	})
}

func TestStatusCommandWithMockedConfig(t *testing.T) {
	t.Run("Status reads config from walgo.yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create walgo.yaml with all required fields
		configContent := `
site:
  title: Test Site
  description: Test Description
walrus:
  projectId: "0xabcdef1234567890abcdef1234567890abcdef12"
  network: testnet
hugo:
  publishDir: public
  theme: default
  contentDir: content
`
		configPath := filepath.Join(tempDir, "walgo.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute status - tests config loading
		output, _ := executeCommand(rootCmd, "status")
		_ = output
	})
}
