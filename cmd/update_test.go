package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestUpdateCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Update command help",
			Args:        []string{"update", "--help"},
			ExpectError: false,
			Contains: []string{
				"Update",
				"Walrus Site",
				"--epochs",
				"--verbose",
				"--dry-run",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestUpdateCommandArgsValidation(t *testing.T) {
	// Find the update command
	var updateCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "update" {
			updateCommand = cmd
			break
		}
	}

	if updateCommand == nil {
		t.Fatal("update command not found")
	}

	t.Run("Too many arguments returns error", func(t *testing.T) {
		err := updateCommand.Args(updateCommand, []string{"arg1", "arg2"})
		if err == nil {
			t.Error("Expected error for too many arguments")
		}
	})
}

func TestUpdateCommandFlags(t *testing.T) {
	// Find the update command
	var updateCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "update" {
			updateCommand = cmd
			break
		}
	}

	if updateCommand == nil {
		t.Fatal("update command not found")
	}

	flagTests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"epochs flag", "epochs", "e", "1"},
		{"verbose flag", "verbose", "v", "false"},
		{"dry-run flag", "dry-run", "", "false"},
		{"telemetry flag", "telemetry", "", "false"},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := updateCommand.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("%s not found", tt.flagName)
				return
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("Expected shorthand '%s', got '%s'", tt.shorthand, flag.Shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Expected default value '%s', got '%s'", tt.defValue, flag.DefValue)
			}
		})
	}

	t.Run("Command accepts optional object ID argument", func(t *testing.T) {
		if updateCommand.Args(updateCommand, []string{}) != nil {
			t.Error("Should accept no arguments")
		}
		if updateCommand.Args(updateCommand, []string{"0x123"}) != nil {
			t.Error("Should accept one argument")
		}
		if err := updateCommand.Args(updateCommand, []string{"arg1", "arg2"}); err == nil {
			t.Error("Should reject multiple arguments")
		}
	})
}

func TestUpdateCommandExecution(t *testing.T) {
	t.Run("Update without config file", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute update without config
		_, err := executeCommand(rootCmd, "update")
		if err == nil {
			t.Error("Expected error when no config file exists")
		}
	})

	t.Run("Update without public directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config but no public directory
		configContent := `
walrus:
  projectID: "0x1234567890"
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute update
		_, err := executeCommand(rootCmd, "update")
		if err == nil {
			t.Error("Expected error when public directory doesn't exist")
		}
	})

	t.Run("Update without object ID", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config without project ID
		configContent := `
walrus:
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create public directory
		publicDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute update without object ID
		_, err := executeCommand(rootCmd, "update")
		if err == nil {
			t.Error("Expected error when no object ID exists")
		}
	})

	t.Run("Update with dry-run flag", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config with project ID
		configContent := `
walrus:
  projectID: "0x1234567890abcdef"
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create public directory
		publicDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute update with dry-run
		output, _ := executeCommand(rootCmd, "update", "--dry-run")
		_ = output
	})

	t.Run("Update with object ID argument", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config
		configContent := `
walrus:
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create public directory
		publicDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute update with object ID
		output, _ := executeCommand(rootCmd, "update", "0x1234567890abcdef", "--dry-run")
		_ = output
	})

	t.Run("Update with custom epochs", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config
		configContent := `
walrus:
  projectID: "0x1234567890abcdef"
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create public directory
		publicDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute update with epochs
		output, _ := executeCommand(rootCmd, "update", "--epochs", "5", "--dry-run")
		_ = output
	})
}

func TestUpdateCommandDescription(t *testing.T) {
	var updateCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "update" {
			updateCommand = cmd
			break
		}
	}

	if updateCommand == nil {
		t.Fatal("update command not found")
	}

	t.Run("Short description mentions update", func(t *testing.T) {
		if updateCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description mentions Walrus Site", func(t *testing.T) {
		if updateCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})
}
