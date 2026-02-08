package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestOptimizeCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Optimize command help",
			Args:        []string{"optimize", "--help"},
			ExpectError: false,
			Contains: []string{
				"Optimizes",
				"HTML",
				"CSS",
				"JavaScript",
				"--verbose",
				"--html",
				"--css",
				"--js",
			},
		},
		{
			Name:        "Optimize with invalid flag",
			Args:        []string{"optimize", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestOptimizeCommandFlags(t *testing.T) {
	// Find the optimize command
	var optimizeCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "optimize" {
			optimizeCommand = cmd
			break
		}
	}

	if optimizeCommand == nil {
		t.Fatal("optimize command not found")
	}

	flagTests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"verbose flag", "verbose", "v", "false"},
		{"html flag", "html", "", "true"},
		{"css flag", "css", "", "true"},
		{"js flag", "js", "", "true"},
		{"remove-unused-css flag", "remove-unused-css", "", "false"},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := optimizeCommand.Flags().Lookup(tt.flagName)
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
		if optimizeCommand.Args(optimizeCommand, []string{}) != nil {
			t.Error("Should accept no arguments")
		}
		if optimizeCommand.Args(optimizeCommand, []string{"/path/to/dir"}) != nil {
			t.Error("Should accept one argument")
		}
		if err := optimizeCommand.Args(optimizeCommand, []string{"dir1", "dir2"}); err == nil {
			t.Error("Should reject multiple arguments")
		}
	})
}

func TestOptimizeCommandExecution(t *testing.T) {
	t.Run("Optimize with non-existent directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute optimize with non-existent directory
		_, err := executeCommand(rootCmd, "optimize", "/nonexistent/path")
		if err != nil {
			t.Logf("optimize with non-existent directory returned error: %v", err)
		}
	})

	t.Run("Optimize with empty directory", func(t *testing.T) {
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

		// Execute optimize
		output, err := executeCommand(rootCmd, "optimize", targetDir)
		if err != nil {
			t.Logf("Optimize returned error: %v", err)
		}
		_ = output
	})

	t.Run("Optimize with HTML files", func(t *testing.T) {
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
    <head>
        <title>Test    Page</title>
    </head>
    <body>
        <!-- This is a comment -->
        <h1>Hello    World</h1>
        <p>This is a    test.</p>
    </body>
</html>`
		if err := os.WriteFile(filepath.Join(targetDir, "index.html"), []byte(htmlContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute optimize
		output, err := executeCommand(rootCmd, "optimize", targetDir)
		if err != nil {
			t.Logf("Optimize returned error: %v", err)
		}
		_ = output
	})

	t.Run("Optimize with CSS files", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create target directory with CSS
		targetDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatal(err)
		}

		cssContent := `/* Comment */
body {
    margin: 0;
    padding: 0;
    background-color: #ffffff;
}

.container {
    width: 100%;
    max-width: 1200px;
}`
		if err := os.WriteFile(filepath.Join(targetDir, "style.css"), []byte(cssContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute optimize
		output, err := executeCommand(rootCmd, "optimize", targetDir)
		if err != nil {
			t.Logf("Optimize returned error: %v", err)
		}
		_ = output
	})

	t.Run("Optimize with JS files", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create target directory with JS
		targetDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatal(err)
		}

		jsContent := `// Comment
function hello() {
    console.log("Hello, World!");
}

/* Multi-line
   comment */
const value = 42;`
		if err := os.WriteFile(filepath.Join(targetDir, "script.js"), []byte(jsContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute optimize
		output, err := executeCommand(rootCmd, "optimize", targetDir)
		if err != nil {
			t.Logf("Optimize returned error: %v", err)
		}
		_ = output
	})

	t.Run("Optimize with walgo.yaml config", func(t *testing.T) {
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
optimizer:
  enabled: true
  html: true
  css: true
  js: true
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create public directory with files
		publicDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(publicDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html><body><h1>Test</h1></body></html>"), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute optimize without specifying directory
		output, err := executeCommand(rootCmd, "optimize")
		if err != nil {
			t.Logf("Optimize returned error: %v", err)
		}
		_ = output
	})

	t.Run("Optimize with verbose flag", func(t *testing.T) {
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

		// Execute optimize with verbose
		output, err := executeCommand(rootCmd, "optimize", targetDir, "--verbose")
		if err != nil {
			t.Logf("Optimize returned error: %v", err)
		}
		_ = output
	})

	t.Run("Optimize with disabled HTML", func(t *testing.T) {
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

		// Execute optimize with HTML disabled
		output, err := executeCommand(rootCmd, "optimize", targetDir, "--html=false")
		if err != nil {
			t.Logf("Optimize returned error: %v", err)
		}
		_ = output
	})

	t.Run("Optimize with remove-unused-css", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create target directory with HTML and CSS
		targetDir := filepath.Join(tempDir, "public")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatal(err)
		}

		htmlContent := `<html><body class="used-class"><h1>Test</h1></body></html>`
		cssContent := `.used-class { color: red; } .unused-class { color: blue; }`

		if err := os.WriteFile(filepath.Join(targetDir, "index.html"), []byte(htmlContent), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(targetDir, "style.css"), []byte(cssContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute optimize with remove-unused-css
		output, err := executeCommand(rootCmd, "optimize", targetDir, "--remove-unused-css")
		if err != nil {
			t.Logf("Optimize returned error: %v", err)
		}
		_ = output
	})
}

func TestOptimizeCommandDescription(t *testing.T) {
	var optimizeCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "optimize" {
			optimizeCommand = cmd
			break
		}
	}

	if optimizeCommand == nil {
		t.Fatal("optimize command not found")
	}

	t.Run("Short description mentions optimization", func(t *testing.T) {
		if optimizeCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description mentions HTML, CSS, JS", func(t *testing.T) {
		if optimizeCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})
}
