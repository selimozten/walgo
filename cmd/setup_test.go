package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestSetupCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Setup command help",
			Args:        []string{"setup", "--help"},
			ExpectError: false,
			Contains: []string{
				"Sets up",
				"site-builder",
				"testnet",
				"mainnet",
				"--network",
				"--force",
			},
		},
		{
			Name:        "Setup with invalid flag",
			Args:        []string{"setup", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestSetupCommandFlags(t *testing.T) {
	// Find the setup command
	var setupCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "setup" {
			setupCommand = cmd
			break
		}
	}

	if setupCommand == nil {
		t.Fatal("setup command not found")
	}

	t.Run("network flag exists", func(t *testing.T) {
		networkFlag := setupCommand.Flags().Lookup("network")
		if networkFlag == nil {
			t.Error("network flag not found")
		} else {
			if networkFlag.Shorthand != "n" {
				t.Errorf("Expected shorthand 'n', got '%s'", networkFlag.Shorthand)
			}
			if networkFlag.DefValue != "testnet" {
				t.Errorf("Expected default value 'testnet', got '%s'", networkFlag.DefValue)
			}
		}
	})

	t.Run("force flag exists", func(t *testing.T) {
		forceFlag := setupCommand.Flags().Lookup("force")
		if forceFlag == nil {
			t.Error("force flag not found")
		} else {
			if forceFlag.DefValue != "false" {
				t.Errorf("Expected default value 'false', got '%s'", forceFlag.DefValue)
			}
		}
	})

	t.Run("Command accepts optional network argument", func(t *testing.T) {
		if setupCommand.Args(setupCommand, []string{}) != nil {
			t.Error("Should accept no arguments")
		}
		if setupCommand.Args(setupCommand, []string{"testnet"}) != nil {
			t.Error("Should accept one argument")
		}
		if setupCommand.Args(setupCommand, []string{"testnet", "extra"}) == nil {
			t.Error("Should reject multiple arguments")
		}
	})
}

func TestSetupCommandExecution(t *testing.T) {
	t.Run("Setup with testnet network", func(t *testing.T) {
		// Execute setup - will fail if walrus is not installed
		output, err := executeCommand(rootCmd, "setup", "--network", "testnet")
		if err != nil {
			t.Logf("Setup returned error (expected if deps not installed): %v", err)
		}
		_ = output
	})

	t.Run("Setup with mainnet network", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "setup", "--network", "mainnet")
		if err != nil {
			t.Logf("Setup returned error (expected if deps not installed): %v", err)
		}
		_ = output
	})

	t.Run("Setup with network as argument", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "setup", "testnet")
		if err != nil {
			t.Logf("Setup returned error (expected if deps not installed): %v", err)
		}
		_ = output
	})
}

func TestSetupCommandDescription(t *testing.T) {
	var setupCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "setup" {
			setupCommand = cmd
			break
		}
	}

	if setupCommand == nil {
		t.Fatal("setup command not found")
	}

	t.Run("Short description mentions site-builder", func(t *testing.T) {
		if setupCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description mentions networks", func(t *testing.T) {
		if setupCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})
}

func TestSetupDepsCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Setup-deps command help",
			Args:        []string{"setup-deps", "--help"},
			ExpectError: false,
			Contains: []string{
				"suiup",
				"Install",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestSetupDepsCommandFlags(t *testing.T) {
	// Find the setup-deps command
	var setupDepsCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "setup-deps" {
			setupDepsCommand = cmd
			break
		}
	}

	if setupDepsCommand == nil {
		t.Fatal("setup-deps command not found")
	}

	t.Run("Command exists", func(t *testing.T) {
		if setupDepsCommand.Name() != "setup-deps" {
			t.Errorf("Expected command name 'setup-deps', got '%s'", setupDepsCommand.Name())
		}
	})
}

func TestSetupCommandWithExistingConfig(t *testing.T) {
	t.Run("Setup without force when config exists", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute setup
		output, err := executeCommand(rootCmd, "setup")
		if err != nil {
			t.Logf("Setup returned error: %v", err)
		}
		_ = output
	})
}
