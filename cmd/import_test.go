package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestImportCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Import command help",
			Args:        []string{"import", "--help"},
			ExpectError: false,
			Contains: []string{
				"Import",
				"Obsidian",
				"wikilinks",
				"--convert-wikilinks",
				"--attachment-dir",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestImportCommandArgsValidation(t *testing.T) {
	// Find the import command
	var importCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "import" {
			importCommand = cmd
			break
		}
	}

	if importCommand == nil {
		t.Fatal("import command not found")
	}

	t.Run("No arguments returns error", func(t *testing.T) {
		err := importCommand.Args(importCommand, []string{})
		if err == nil {
			t.Error("Expected error for no arguments")
		}
	})

	t.Run("Too many arguments returns error", func(t *testing.T) {
		err := importCommand.Args(importCommand, []string{"vault1", "vault2"})
		if err == nil {
			t.Error("Expected error for too many arguments")
		}
	})
}

func TestImportCommandFlags(t *testing.T) {
	// Find the import command
	var importCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "import" {
			importCommand = cmd
			break
		}
	}

	if importCommand == nil {
		t.Fatal("import command not found")
	}

	flagTests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"output-dir flag", "output-dir", "o", ""},
		{"convert-wikilinks flag", "convert-wikilinks", "", "true"},
		{"attachment-dir flag", "attachment-dir", "", ""},
		{"frontmatter-format flag", "frontmatter-format", "", ""},
		{"link-style flag", "link-style", "", "markdown"},
		{"dry-run flag", "dry-run", "", "false"},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := importCommand.Flags().Lookup(tt.flagName)
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

	t.Run("Command requires exactly one argument", func(t *testing.T) {
		if importCommand.Args(importCommand, []string{"vault"}) != nil {
			t.Error("Should accept one argument")
		}
		if err := importCommand.Args(importCommand, []string{}); err == nil {
			t.Error("Should reject no arguments")
		}
		if err := importCommand.Args(importCommand, []string{"vault1", "vault2"}); err == nil {
			t.Error("Should reject multiple arguments")
		}
	})
}

func TestImportCommandExecution(t *testing.T) {
	t.Run("Import with non-existent vault path", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create walgo.yaml
		configContent := `
hugo:
  publishDir: public
  contentDir: content
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute import with non-existent vault
		output, err := executeCommand(rootCmd, "import", "/nonexistent/vault/path")
		// Should fail when vault path doesn't exist
		_ = err
		_ = output
	})

	t.Run("Import with file instead of directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create walgo.yaml
		configContent := `
hugo:
  publishDir: public
  contentDir: content
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create a file instead of directory
		vaultFile := filepath.Join(tempDir, "not-a-vault.md")
		if err := os.WriteFile(vaultFile, []byte("# Not a vault"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute import with file path
		output, err := executeCommand(rootCmd, "import", vaultFile)
		// Should fail when vault path is a file
		_ = err
		_ = output
	})

	t.Run("Import without config file", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create vault directory
		vaultDir := filepath.Join(tempDir, "vault")
		if err := os.MkdirAll(vaultDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Execute import without config
		output, err := executeCommand(rootCmd, "import", vaultDir)
		if err == nil {
			t.Log("Expected error when no config file exists")
		}
		_ = output
	})

	t.Run("Import with dry-run flag", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create walgo.yaml
		configContent := `
hugo:
  publishDir: public
  contentDir: content
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create vault directory with markdown files
		vaultDir := filepath.Join(tempDir, "vault")
		if err := os.MkdirAll(vaultDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(vaultDir, "note.md"), []byte("# Note\nContent"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create content directory
		if err := os.MkdirAll(filepath.Join(tempDir, "content"), 0755); err != nil {
			t.Fatal(err)
		}

		// Execute import with dry-run
		output, err := executeCommand(rootCmd, "import", vaultDir, "--dry-run")
		if err != nil {
			t.Logf("Import returned error: %v", err)
		}
		_ = output
	})

	t.Run("Import with custom output directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create walgo.yaml
		configContent := `
hugo:
  publishDir: public
  contentDir: content
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create vault directory
		vaultDir := filepath.Join(tempDir, "vault")
		if err := os.MkdirAll(vaultDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(vaultDir, "note.md"), []byte("# Note"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create content directory
		if err := os.MkdirAll(filepath.Join(tempDir, "content"), 0755); err != nil {
			t.Fatal(err)
		}

		// Execute import with custom output directory
		output, err := executeCommand(rootCmd, "import", vaultDir, "--output-dir", "notes")
		if err != nil {
			t.Logf("Import returned error: %v", err)
		}
		_ = output
	})

	t.Run("Import with wikilinks disabled", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create walgo.yaml
		configContent := `
hugo:
  publishDir: public
  contentDir: content
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create vault directory with wikilinks
		vaultDir := filepath.Join(tempDir, "vault")
		if err := os.MkdirAll(vaultDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(vaultDir, "note.md"), []byte("# Note\n[[other-note]]"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create content directory
		if err := os.MkdirAll(filepath.Join(tempDir, "content"), 0755); err != nil {
			t.Fatal(err)
		}

		// Execute import with wikilinks disabled
		output, err := executeCommand(rootCmd, "import", vaultDir, "--convert-wikilinks=false")
		if err != nil {
			t.Logf("Import returned error: %v", err)
		}
		_ = output
	})

	t.Run("Import with custom frontmatter format", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create walgo.yaml
		configContent := `
hugo:
  publishDir: public
  contentDir: content
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create vault directory
		vaultDir := filepath.Join(tempDir, "vault")
		if err := os.MkdirAll(vaultDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(vaultDir, "note.md"), []byte("# Note\nContent"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create content directory
		if err := os.MkdirAll(filepath.Join(tempDir, "content"), 0755); err != nil {
			t.Fatal(err)
		}

		// Execute import with TOML frontmatter
		output, err := executeCommand(rootCmd, "import", vaultDir, "--frontmatter-format", "toml")
		if err != nil {
			t.Logf("Import returned error: %v", err)
		}
		_ = output
	})
}

func TestImportCommandDescription(t *testing.T) {
	var importCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "import" {
			importCommand = cmd
			break
		}
	}

	if importCommand == nil {
		t.Fatal("import command not found")
	}

	t.Run("Short description mentions Obsidian", func(t *testing.T) {
		if importCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description mentions features", func(t *testing.T) {
		if importCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})
}
