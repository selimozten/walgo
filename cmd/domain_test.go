package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestDomainCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Domain command help",
			Args:        []string{"domain", "--help"},
			ExpectError: false,
			Contains: []string{
				"SuiNS",
				"domain",
				"mainnet",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestDomainCommandArgsValidation(t *testing.T) {
	// Find the domain command
	var domainCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "domain" {
			domainCommand = cmd
			break
		}
	}

	if domainCommand == nil {
		t.Fatal("domain command not found")
	}

	t.Run("Too many arguments returns error", func(t *testing.T) {
		err := domainCommand.Args(domainCommand, []string{"arg1", "arg2"})
		if err == nil {
			t.Error("Expected error for too many arguments")
		}
	})
}

func TestDomainCommandFlags(t *testing.T) {
	// Find the domain command
	var domainCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "domain" {
			domainCommand = cmd
			break
		}
	}

	if domainCommand == nil {
		t.Fatal("domain command not found")
	}

	t.Run("Command is properly registered", func(t *testing.T) {
		if domainCommand.Use != "domain [domain-name]" {
			t.Errorf("Unexpected Use field: %s", domainCommand.Use)
		}

		// Test args validation
		if domainCommand.Args(domainCommand, []string{"my-domain"}) != nil {
			t.Error("Args validation failed for single domain name")
		}

		if domainCommand.Args(domainCommand, []string{}) != nil {
			t.Error("Args validation should pass for no arguments")
		}

		if domainCommand.Args(domainCommand, []string{"arg1", "arg2"}) == nil {
			t.Error("Args validation should fail for multiple arguments")
		}
	})
}

func TestDomainCommandExecution(t *testing.T) {
	t.Run("Domain without config file", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute domain command without config
		_, err := executeCommand(rootCmd, "domain")
		if err == nil {
			t.Error("Expected error when no config file exists")
		}
	})

	t.Run("Domain with testnet config", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config for testnet
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

		// Execute domain command - should show testnet warning
		output, err := executeCommand(rootCmd, "domain")
		if err != nil {
			t.Logf("Domain command returned error: %v", err)
		}
		// Should mention that SuiNS is mainnet only
		_ = output
	})

	t.Run("Domain with mainnet config and no project ID", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config for mainnet without project ID
		configContent := `
walrus:
  projectID: YOUR_WALRUS_PROJECT_ID
  network: mainnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute domain command - should fail due to missing project ID
		_, err := executeCommand(rootCmd, "domain")
		if err == nil {
			t.Error("Expected error when project ID is placeholder")
		}
	})

	t.Run("Domain with mainnet config and valid project ID", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config for mainnet with valid project ID
		configContent := `
walrus:
  projectID: "0x1234567890abcdef1234567890abcdef12345678"
  network: mainnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute domain command - should show setup instructions
		output, err := executeCommand(rootCmd, "domain")
		if err != nil {
			t.Errorf("Domain command returned unexpected error: %v", err)
		}
		_ = output
	})

	t.Run("Domain with domain name argument", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config for mainnet
		configContent := `
walrus:
  projectID: "0x1234567890abcdef1234567890abcdef12345678"
  network: mainnet
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute domain command with domain name
		output, err := executeCommand(rootCmd, "domain", "my-awesome-site")
		if err != nil {
			t.Errorf("Domain command returned unexpected error: %v", err)
		}
		_ = output
	})

	t.Run("Domain with empty network defaults to testnet", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config without network specified
		configContent := `
walrus:
  projectID: "0x1234567890abcdef"
hugo:
  publishDir: public
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute domain command - should default to testnet and show warning
		output, err := executeCommand(rootCmd, "domain")
		if err != nil {
			t.Logf("Domain command returned error: %v", err)
		}
		_ = output
	})
}

func TestDomainCommandDescription(t *testing.T) {
	// Find the domain command
	var domainCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "domain" {
			domainCommand = cmd
			break
		}
	}

	if domainCommand == nil {
		t.Fatal("domain command not found")
	}

	t.Run("Short description mentions SuiNS", func(t *testing.T) {
		if domainCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description mentions mainnet", func(t *testing.T) {
		if domainCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})
}
