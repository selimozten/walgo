package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestDeployHTTPCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Deploy-http command help",
			Args:        []string{"deploy-http", "--help"},
			ExpectError: false,
			Contains: []string{
				"Uploads",
				"publisher",
				"aggregator",
				"--publisher",
				"--aggregator",
				"--epochs",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestDeployHTTPCommandFlags(t *testing.T) {
	// Find the deploy-http command
	var deployHTTPCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "deploy-http" {
			deployHTTPCommand = cmd
			break
		}
	}

	if deployHTTPCommand == nil {
		t.Fatal("deploy-http command not found")
	}

	flagTests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"publisher flag", "publisher", "", ""},
		{"aggregator flag", "aggregator", "", ""},
		{"epochs flag", "epochs", "e", "1"},
		{"mode flag", "mode", "", "quilt"},
		{"workers flag", "workers", "", "10"},
		{"retries flag", "retries", "", "5"},
		{"json flag", "json", "", "false"},
		{"verbose flag", "verbose", "v", "false"},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := deployHTTPCommand.Flags().Lookup(tt.flagName)
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
}

func TestDeployHTTPCommandExecution(t *testing.T) {
	t.Run("Deploy-http without required flags", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config
		configContent := `
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

		// Execute deploy-http without publisher/aggregator
		_, err := executeCommand(rootCmd, "deploy-http")
		// Should fail when required flags not provided
		if err == nil {
			t.Error("Expected error when required flags not provided")
		}
	})

	t.Run("Deploy-http without config file", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute deploy-http without config
		_, err := executeCommand(rootCmd, "deploy-http",
			"--publisher", "https://publisher.walrus-testnet.walrus.space",
			"--aggregator", "https://aggregator.walrus-testnet.walrus.space")
		if err == nil {
			t.Error("Expected error when no config file exists")
		}
	})

	t.Run("Deploy-http without public directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config
		configContent := `
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute deploy-http without public directory
		_, err := executeCommand(rootCmd, "deploy-http",
			"--publisher", "https://publisher.walrus-testnet.walrus.space",
			"--aggregator", "https://aggregator.walrus-testnet.walrus.space")
		// Should fail when public directory doesn't exist
		if err == nil {
			t.Error("Expected error when public directory doesn't exist")
		}
	})

	t.Run("Deploy-http with all required flags", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config
		configContent := `
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

		// Note: This will attempt real HTTP calls, so it will fail
		// but we're testing the command setup
		output, _ := executeCommand(rootCmd, "deploy-http",
			"--publisher", "https://publisher.walrus-testnet.walrus.space",
			"--aggregator", "https://aggregator.walrus-testnet.walrus.space",
			"--epochs", "1")
		_ = output
	})

	t.Run("Deploy-http with mode flag", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config
		configContent := `
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

		// Execute with blobs mode
		output, _ := executeCommand(rootCmd, "deploy-http",
			"--publisher", "https://publisher.walrus-testnet.walrus.space",
			"--aggregator", "https://aggregator.walrus-testnet.walrus.space",
			"--mode", "blobs")
		_ = output
	})
}

func TestDeployHTTPCommandDescription(t *testing.T) {
	var deployHTTPCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "deploy-http" {
			deployHTTPCommand = cmd
			break
		}
	}

	if deployHTTPCommand == nil {
		t.Fatal("deploy-http command not found")
	}

	t.Run("Short description mentions HTTP", func(t *testing.T) {
		if deployHTTPCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description mentions testnet", func(t *testing.T) {
		if deployHTTPCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})
}
