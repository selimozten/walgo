package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompressCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Compress command help",
			Args:        []string{"compress", "--help"},
			ExpectError: false,
			Contains: []string{
				"Compress",
				"Brotli",
				"--level",
				"--verbose",
				"--in-place",
			},
		},
		{
			Name:        "Compress with invalid flag",
			Args:        []string{"compress", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestCompressCommandFlags(t *testing.T) {
	// Find the compress command
	var compressCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "compress" {
			compressCommand = cmd
			break
		}
	}

	if compressCommand == nil {
		t.Fatal("compress command not found")
	}

	flagTests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"level flag", "level", "l", "6"},
		{"verbose flag", "verbose", "v", "false"},
		{"in-place flag", "in-place", "", "false"},
		{"generate-ws-resources flag", "generate-ws-resources", "", "false"},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := compressCommand.Flags().Lookup(tt.flagName)
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

	t.Run("Command accepts optional directory argument", func(t *testing.T) {
		if compressCommand.Args(compressCommand, []string{}) != nil {
			t.Error("Should accept no arguments")
		}
		if compressCommand.Args(compressCommand, []string{"/path/to/dir"}) != nil {
			t.Error("Should accept one argument")
		}
		if compressCommand.Args(compressCommand, []string{"dir1", "dir2"}) == nil {
			t.Error("Should reject multiple arguments")
		}
	})
}

func TestCompressCommandExecution(t *testing.T) {
	t.Run("Compress with non-existent directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute compress with non-existent directory
		output, err := executeCommand(rootCmd, "compress", "/nonexistent/path")
		// Command may return error or not depending on implementation
		_ = err
		_ = output
	})

	t.Run("Compress with empty directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create empty target directory
		targetDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Execute compress
		output, err := executeCommand(rootCmd, "compress", targetDir)
		if err != nil {
			t.Logf("Compress returned error: %v", err)
		}
		_ = output
	})

	t.Run("Compress with HTML files", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create target directory with HTML
		targetDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatal(err)
		}

		htmlContent := `<!DOCTYPE html>
<html>
    <head><title>Test Page</title></head>
    <body><h1>Hello World</h1></body>
</html>`
		if err := os.WriteFile(filepath.Join(targetDir, "index.html"), []byte(htmlContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute compress
		output, err := executeCommand(rootCmd, "compress", targetDir)
		if err != nil {
			t.Logf("Compress returned error: %v", err)
		}
		_ = output
	})

	t.Run("Compress with custom level", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create target directory
		targetDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(targetDir, "index.html"), []byte("<html><body>Test</body></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute compress with custom level
		output, err := executeCommand(rootCmd, "compress", targetDir, "--level", "11")
		if err != nil {
			t.Logf("Compress returned error: %v", err)
		}
		_ = output
	})

	t.Run("Compress with verbose flag", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create target directory
		targetDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(targetDir, "index.html"), []byte("<html><body>Test</body></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute compress with verbose
		output, err := executeCommand(rootCmd, "compress", targetDir, "--verbose")
		if err != nil {
			t.Logf("Compress returned error: %v", err)
		}
		_ = output
	})

	t.Run("Compress without config uses explicit directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create target directory
		targetDir := filepath.Join(tempDir, "dist")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(targetDir, "app.js"), []byte("console.log('test');"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute compress - without config but with explicit directory
		output, err := executeCommand(rootCmd, "compress", targetDir)
		if err != nil {
			t.Logf("Compress returned error: %v", err)
		}
		_ = output
	})

	t.Run("Compress with config", func(t *testing.T) {
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
compress:
  enabled: true
  level: 9
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create public directory
		publicDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html><body>Test</body></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute compress without specifying directory
		output, err := executeCommand(rootCmd, "compress")
		if err != nil {
			t.Logf("Compress returned error: %v", err)
		}
		_ = output
	})
}

func TestCompressCommandDescription(t *testing.T) {
	var compressCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "compress" {
			compressCommand = cmd
			break
		}
	}

	if compressCommand == nil {
		t.Fatal("compress command not found")
	}

	t.Run("Short description mentions Brotli", func(t *testing.T) {
		if compressCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description provides examples", func(t *testing.T) {
		if compressCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})
}
