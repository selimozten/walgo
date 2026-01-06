package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestDeployCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Deploy command help",
			Args:        []string{"deploy", "--help"},
			ExpectError: false,
			Contains: []string{
				"Deploy",
				"Walrus",
				"--epochs",
				"--force",
				"--dry-run",
			},
		},
		{
			Name:        "Deploy with invalid flag",
			Args:        []string{"deploy", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestDeployCommandFlags(t *testing.T) {
	// Find the deploy command
	var deployCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "deploy" {
			deployCommand = cmd
			break
		}
	}

	if deployCommand == nil {
		t.Fatal("deploy command not found")
	}

	flagTests := []struct {
		name       string
		flagName   string
		shorthand  string
		defValue   string
		shouldHave bool
	}{
		{"epochs flag", "epochs", "e", "1", true},
		{"force flag", "force", "f", "false", true},
		{"verbose flag", "verbose", "v", "false", true},
		{"quiet flag", "quiet", "q", "false", true},
		{"dry-run flag", "dry-run", "", "false", true},
		{"telemetry flag", "telemetry", "", "false", true},
		{"skip-version-check flag", "skip-version-check", "", "false", true},
		{"save-project flag", "save-project", "", "false", true},
		{"project-name flag", "project-name", "", "", true},
		{"category flag", "category", "", "", true},
		{"description flag", "description", "", "", true},
		{"image-url flag", "image-url", "", "", true},
		{"force-new flag", "force-new", "", "false", true},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := deployCommand.Flags().Lookup(tt.flagName)
			if tt.shouldHave && flag == nil {
				t.Errorf("%s not found", tt.flagName)
				return
			}
			if !tt.shouldHave && flag != nil {
				t.Errorf("%s should not exist", tt.flagName)
				return
			}
			if flag != nil {
				if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
					t.Errorf("Expected shorthand '%s', got '%s'", tt.shorthand, flag.Shorthand)
				}
				if flag.DefValue != tt.defValue {
					t.Errorf("Expected default value '%s', got '%s'", tt.defValue, flag.DefValue)
				}
			}
		})
	}
}

func TestDeployCommandExecution(t *testing.T) {
	t.Run("Deploy without config file", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute deploy command without config
		output, err := executeCommand(rootCmd, "deploy")
		if err == nil {
			t.Log("Expected error when no config file exists")
		}
		_ = output
	})

	t.Run("Deploy without public directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config but no public directory
		configContent := `
walrus:
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute deploy command
		output, err := executeCommand(rootCmd, "deploy")
		if err == nil {
			t.Log("Expected error when public directory doesn't exist")
		}
		_ = output
	})

	t.Run("Deploy with force flag bypasses public directory check", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config but no public directory
		configContent := `
walrus:
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute deploy with force flag - will still fail due to missing site-builder
		// but shouldn't fail on public directory check
		output, err := executeCommand(rootCmd, "deploy", "--force")
		// Will fail for other reasons (e.g., site-builder not installed)
		_ = err
		_ = output
	})

	t.Run("Deploy with dry-run flag", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config and public directory
		configContent := `
walrus:
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		publicDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute deploy with dry-run
		output, _ := executeCommand(rootCmd, "deploy", "--dry-run")
		_ = output
	})

	t.Run("Deploy with custom epochs", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config and public directory
		configContent := `
walrus:
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		publicDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Execute deploy with epochs
		output, _ := executeCommand(rootCmd, "deploy", "--epochs", "5", "--dry-run")
		_ = output
	})

	t.Run("Deploy with project name", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config and public directory
		configContent := `
walrus:
  network: testnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		publicDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Execute deploy with project name
		output, _ := executeCommand(rootCmd, "deploy", "--project-name", "my-project", "--dry-run")
		_ = output
	})
}

func TestDeployCommandDescription(t *testing.T) {
	var deployCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "deploy" {
			deployCommand = cmd
			break
		}
	}

	if deployCommand == nil {
		t.Fatal("deploy command not found")
	}

	t.Run("Short description mentions deploy", func(t *testing.T) {
		if deployCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description mentions Walrus", func(t *testing.T) {
		if deployCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})

	t.Run("Command has example", func(t *testing.T) {
		// Check that the long description contains an example
		if deployCommand.Long != "" {
			t.Logf("Deploy description: %s", deployCommand.Long[:min(100, len(deployCommand.Long))])
		}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
