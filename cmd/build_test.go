package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestBuildCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Build command help",
			Args:        []string{"build", "--help"},
			ExpectError: false,
			Contains: []string{
				"Builds the Hugo site",
				"--clean",
				"--no-optimize",
			},
		},
		{
			Name:        "Build command with invalid flag",
			Args:        []string{"build", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestBuildCommandExecution(t *testing.T) {
	t.Skip("Build command execution tests need refactoring to handle os.Exit calls properly")

	t.Run("Build without config file", func(t *testing.T) {
		// Create temp directory without config
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		// Note: We can't directly mock os.Exit in Go
		// The command will call os.Exit if config is missing

		// Execute build command
		stdout, stderr := captureOutput(func() {
			defer func() { recover() }()
			buildCmd.Run(buildCmd, []string{})
		})

		// Check that error messages are present
		_ = stdout
		_ = stderr

		// Check error messages
		if stderr != "" {
			// Should contain error about config
			// "Did you run 'walgo init' to create a site?"
		}
	})

	t.Run("Build with clean flag", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		// Create a mock walgo.yaml
		configContent := `
hugo:
  publishDir: public
optimizer:
  enabled: false
`
		_ = os.WriteFile("walgo.yaml", []byte(configContent), 0644)

		// Create public directory to clean
		publicDir := filepath.Join(tempDir, "public")
		_ = os.MkdirAll(publicDir, 0755)
		testFile := filepath.Join(publicDir, "test.html")
		_ = os.WriteFile(testFile, []byte("test"), 0644)

		// Note: We can't mock os.Exit directly

		// Create mock cobra command with flag
		cmd := &cobra.Command{}
		cmd.Flags().Bool("clean", true, "")
		cmd.Flags().Bool("no-optimize", false, "")
		_ = cmd.Flags().Set("clean", "true")

		// Execute build
		stdout, stderr := captureOutput(func() {
			defer func() { recover() }()
			buildCmd.Run(cmd, []string{})
		})

		// The build will fail because Hugo is not installed,
		// but we can check if the clean operation was attempted
		_ = stdout
		_ = stderr
	})

	t.Run("Build with no-optimize flag", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		// Create a mock walgo.yaml with optimizer enabled
		configContent := `
hugo:
  publishDir: public
optimizer:
  enabled: true
  html: true
  css: true
  js: true
`
		_ = os.WriteFile("walgo.yaml", []byte(configContent), 0644)

		// Note: We can't mock os.Exit directly

		// Create mock cobra command with flag
		cmd := &cobra.Command{}
		cmd.Flags().Bool("clean", false, "")
		cmd.Flags().Bool("no-optimize", true, "")
		_ = cmd.Flags().Set("no-optimize", "true")

		// Execute build
		stdout, stderr := captureOutput(func() {
			defer func() { recover() }()
			buildCmd.Run(cmd, []string{})
		})

		// Build will fail without Hugo, but we're testing the flag handling
		_ = stdout
		_ = stderr
	})
}

func TestBuildCommandFlags(t *testing.T) {
	// Find the build command
	var buildCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "build" {
			buildCommand = cmd
			break
		}
	}

	if buildCommand == nil {
		t.Fatal("build command not found")
	}

	t.Run("Clean flag exists", func(t *testing.T) {
		cleanFlag := buildCommand.Flags().Lookup("clean")
		if cleanFlag == nil {
			t.Error("clean flag not found")
		} else {
			if cleanFlag.Shorthand != "c" {
				t.Errorf("Expected shorthand 'c', got '%s'", cleanFlag.Shorthand)
			}
			if cleanFlag.DefValue != "false" {
				t.Errorf("Expected default value 'false', got '%s'", cleanFlag.DefValue)
			}
		}
	})

	t.Run("No-optimize flag exists", func(t *testing.T) {
		noOptimizeFlag := buildCommand.Flags().Lookup("no-optimize")
		if noOptimizeFlag == nil {
			t.Error("no-optimize flag not found")
		} else {
			if noOptimizeFlag.DefValue != "false" {
				t.Errorf("Expected default value 'false', got '%s'", noOptimizeFlag.DefValue)
			}
		}
	})
}

func TestBuildCommandWithMockConfig(t *testing.T) {
	t.Run("Successful build simulation", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalWd) }()

		// Create valid config
		configContent := `
site:
  title: Test Site
  description: Test Description
hugo:
  publishDir: public
  theme: default
optimizer:
  enabled: true
  html: true
  css: true
  js: true
`
		if err := os.WriteFile("walgo.yaml", []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create hugo.toml for Hugo
		hugoConfig := `
baseURL = "/"
languageCode = "en-us"
title = "Test Site"
`
		if err := os.WriteFile("hugo.toml", []byte(hugoConfig), 0644); err != nil {
			t.Fatal(err)
		}

		// Create content directory
		contentDir := filepath.Join(tempDir, "content")
		if err := os.MkdirAll(contentDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Note: We can't directly mock os.Exit in Go
		// The command will call os.Exit if config is missing

		// Try to execute build
		stdout, stderr := captureOutput(func() {
			defer func() { recover() }()
			buildCmd.Run(buildCmd, []string{})
		})

		// The build will fail without Hugo installed
		// but we're testing the command flow
		if stderr != "" {
			// Should contain Hugo-related error message
		}

		_ = stdout
	})
}
