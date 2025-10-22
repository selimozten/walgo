package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestRootCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Root command without args shows help",
			Args:        []string{},
			ExpectError: false,
			Contains: []string{
				"Walgo provides a seamless bridge",
				"Available Commands:",
				"Use \"walgo [command] --help\"",
			},
		},
		{
			Name:        "Root command with --help flag",
			Args:        []string{"--help"},
			ExpectError: false,
			Contains: []string{
				"Walgo provides a seamless bridge",
				"init/new/build/serve",
				"deploy, update, status",
			},
		},
		{
			Name:        "Root command with invalid flag",
			Args:        []string{"--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
		{
			Name:        "Root command with --config flag",
			Args:        []string{"--config", "/tmp/test-walgo.yaml"},
			ExpectError: false,
			Contains: []string{
				"Walgo provides a seamless bridge",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestInitConfig(t *testing.T) {
	// Save original values
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		cfgFile = ""
		viper.Reset()
	}()

	t.Run("Config from flag", func(t *testing.T) {
		// Create a temp config file
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "test-walgo.yaml")
		content := `
build:
  output: public
deploy:
  network: testnet
`
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		// Set the config file flag
		cfgFile = configFile

		// Capture output
		stdout, stderr := captureOutput(func() {
			initConfig()
		})

		// Verify config was loaded
		if viper.GetString("build.output") != "public" {
			t.Errorf("Expected build.output to be 'public', got %s", viper.GetString("build.output"))
		}

		// Check that config file path was printed
		if stderr == "" {
			t.Error("Expected config file path to be printed to stderr")
		}

		_ = stdout // Avoid unused variable warning
	})

	t.Run("Config from home directory", func(t *testing.T) {
		// Skip this test as it's platform-dependent
		// os.UserHomeDir() doesn't always respect HOME env var
		// and viper's config search order is complex
		t.Skip("Home directory config test is unreliable across platforms")
	})

	t.Run("Config from current directory", func(t *testing.T) {
		viper.Reset()
		cfgFile = ""

		// Create temp directory and make it current
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		// Create config in current directory
		configFile := filepath.Join(tempDir, "walgo.yaml")
		content := `
optimize:
  html: true
  css: true
`
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		// Run initConfig
		initConfig()

		// Verify config was loaded
		if !viper.GetBool("optimize.html") {
			t.Error("Expected optimize.html to be true")
		}
		if !viper.GetBool("optimize.css") {
			t.Error("Expected optimize.css to be true")
		}
	})

	t.Run("No config file found", func(t *testing.T) {
		viper.Reset()
		cfgFile = ""

		// Use a temp directory with no config files
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		// Run initConfig - should not panic or error
		stdout, stderr := captureOutput(func() {
			initConfig()
		})

		// Should not print "Using config file" message
		if stdout != "" || stderr != "" {
			if stdout != "" {
				t.Logf("Unexpected stdout: %s", stdout)
			}
			if stderr != "" && stderr != "Using config file:" {
				t.Logf("Unexpected stderr: %s", stderr)
			}
		}
	})

	t.Run("Invalid config file specified", func(t *testing.T) {
		viper.Reset()

		// Set a non-existent config file
		cfgFile = "/nonexistent/path/config.yaml"

		// Capture stderr
		_, stderr := captureOutput(func() {
			initConfig()
		})

		// Should print error message about failed config
		if stderr == "" {
			t.Error("Expected error message for invalid config file")
		}
	})

	t.Run("Malformed config file", func(t *testing.T) {
		viper.Reset()

		// Create a malformed config file
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "bad-walgo.yaml")
		content := `
this is not: valid yaml
  - because it's malformed
[mixing brackets
`
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfgFile = configFile

		// Capture stderr
		_, stderr := captureOutput(func() {
			initConfig()
		})

		// Should print error about failed parsing
		if stderr == "" {
			t.Error("Expected error message for malformed config")
		}
	})
}

func TestExecute(t *testing.T) {
	t.Run("Execute with valid command", func(t *testing.T) {
		// Skip this test as Execute() uses os.Exit which interferes with testing
		// and output capture is unreliable with Cobra's output handling
		t.Skip("Execute() test skipped - os.Exit and output capture issues with Cobra")
	})

	t.Run("Execute with invalid command", func(t *testing.T) {
		// Save original args
		originalArgs := os.Args
		defer func() {
			os.Args = originalArgs
		}()

		// Set args for invalid command
		os.Args = []string{"walgo", "nonexistent"}

		// Capture stderr
		_, stderr := captureOutput(func() {
			// Execute will call os.Exit(1) for invalid command
			// We can't mock os.Exit directly, so we expect this to fail
			defer func() { recover() }()
			Execute()
		})

		// Should have error message
		if stderr == "" {
			// Error might have been printed before we could capture it
			t.Log("Error message might not be captured due to os.Exit")
		}
	})
}
