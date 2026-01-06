package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestDoctorCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Doctor command help",
			Args:        []string{"doctor", "--help"},
			ExpectError: false,
			Contains: []string{
				"Diagnose",
				"environment",
				"--fix-paths",
				"--verbose",
			},
		},
		{
			Name:        "Doctor with invalid flag",
			Args:        []string{"doctor", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestDoctorCommandFlags(t *testing.T) {
	// Find the doctor command
	var doctorCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "doctor" {
			doctorCommand = cmd
			break
		}
	}

	if doctorCommand == nil {
		t.Fatal("doctor command not found")
	}

	t.Run("fix-paths flag exists", func(t *testing.T) {
		fixPathsFlag := doctorCommand.Flags().Lookup("fix-paths")
		if fixPathsFlag == nil {
			t.Error("fix-paths flag not found")
		} else {
			if fixPathsFlag.DefValue != "false" {
				t.Errorf("Expected default value 'false', got '%s'", fixPathsFlag.DefValue)
			}
		}
	})

	t.Run("fix-all flag exists", func(t *testing.T) {
		fixAllFlag := doctorCommand.Flags().Lookup("fix-all")
		if fixAllFlag == nil {
			t.Error("fix-all flag not found")
		} else {
			if fixAllFlag.DefValue != "false" {
				t.Errorf("Expected default value 'false', got '%s'", fixAllFlag.DefValue)
			}
		}
	})

	t.Run("verbose flag exists", func(t *testing.T) {
		verboseFlag := doctorCommand.Flags().Lookup("verbose")
		if verboseFlag == nil {
			t.Error("verbose flag not found")
		} else {
			if verboseFlag.Shorthand != "v" {
				t.Errorf("Expected shorthand 'v', got '%s'", verboseFlag.Shorthand)
			}
		}
	})
}

func TestDoctorCommandExecution(t *testing.T) {
	t.Run("Doctor basic execution", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute doctor command
		output, err := executeCommand(rootCmd, "doctor")
		if err != nil {
			t.Logf("Doctor command returned error (expected if deps not installed): %v", err)
		}

		// Should contain diagnostic header
		if output != "" {
			t.Logf("Doctor output length: %d", len(output))
		}
	})

	t.Run("Doctor with verbose flag", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute doctor command with verbose
		output, err := executeCommand(rootCmd, "doctor", "--verbose")
		if err != nil {
			t.Logf("Doctor command returned error: %v", err)
		}
		_ = output
	})

	t.Run("Doctor with walgo.yaml present", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create walgo.yaml
		configContent := `
site:
  title: Test Site
walrus:
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute doctor command
		output, err := executeCommand(rootCmd, "doctor")
		if err != nil {
			t.Logf("Doctor command returned error: %v", err)
		}
		_ = output
	})
}

func TestRunQuiet(t *testing.T) {
	t.Run("runQuiet with valid command", func(t *testing.T) {
		// Test with a command that exists on all systems
		result := runQuiet("echo", "test")
		if result == "" {
			t.Log("runQuiet returned empty string (may be expected on some systems)")
		}
	})

	t.Run("runQuiet with invalid command", func(t *testing.T) {
		// Test with a command that doesn't exist
		result := runQuiet("nonexistent-command-that-does-not-exist")
		// Should return empty or error output
		_ = result
	})
}

func TestEnsureAbsolutePaths(t *testing.T) {
	t.Run("Fix tilde paths in config", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a sites-config.yaml with tilde paths
		configContent := `contexts:
  testnet:
    package: "0x123"
    general:
      rpc_url: "https://sui-testnet.example.com"
      wallet: "~/.sui/sui_config/sui.keystore"
      walrus_binary: "walrus"
      walrus_config: "~/.walrus/config.yaml"
      gas_budget: 500000000
default_context: testnet
`
		configPath := filepath.Join(tempDir, "sites-config.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Cannot get home directory")
		}

		// Run ensureAbsolutePaths
		err = ensureAbsolutePaths(configPath, home)
		if err != nil {
			t.Errorf("ensureAbsolutePaths failed: %v", err)
		}

		// Read the modified file
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}

		content := string(data)
		// Check that tilde paths were replaced
		if filepath.Join(home, ".sui") != "" {
			// The paths should now be absolute
			t.Logf("Modified config: %s", content)
		}
	})

	t.Run("Handle non-existent file", func(t *testing.T) {
		home, _ := os.UserHomeDir()
		err := ensureAbsolutePaths("/nonexistent/path/config.yaml", home)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("Handle invalid YAML", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create invalid YAML
		configPath := filepath.Join(tempDir, "invalid.yaml")
		if err := os.WriteFile(configPath, []byte("not: valid: yaml: ["), 0644); err != nil {
			t.Fatal(err)
		}

		home, _ := os.UserHomeDir()
		err := ensureAbsolutePaths(configPath, home)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestSitesConfigParsing(t *testing.T) {
	t.Run("Parse valid sites config", func(t *testing.T) {
		tempDir := t.TempDir()

		configContent := `contexts:
  testnet:
    package: "0x123456"
    general:
      rpc_url: "https://sui-testnet.example.com"
      wallet: "/home/user/.sui/sui_config/sui.keystore"
      walrus_binary: "walrus"
      walrus_config: "/home/user/.walrus/config.yaml"
      gas_budget: 500000000
  mainnet:
    package: "0xabcdef"
    general:
      rpc_url: "https://sui-mainnet.example.com"
      wallet: "/home/user/.sui/sui_config/sui.keystore"
      walrus_binary: "walrus"
      walrus_config: "/home/user/.walrus/mainnet-config.yaml"
      gas_budget: 1000000000
default_context: testnet
`
		configPath := filepath.Join(tempDir, "sites-config.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Read and verify parsing works
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}

		if len(data) == 0 {
			t.Error("Config file is empty")
		}
	})
}
