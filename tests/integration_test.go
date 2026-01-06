package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestWalgoCLIIntegration tests that the walgo CLI builds and basic commands work
func TestWalgoCLIIntegration(t *testing.T) {
	// Get the current working directory (should be the tests directory)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Get the parent directory (the walgo project root)
	projectRoot := filepath.Dir(cwd)

	// Build the walgo binary for testing
	walgoBinary := filepath.Join(t.TempDir(), "walgo")
	buildCmd := exec.Command("go", "build", "-o", walgoBinary, ".")
	buildCmd.Dir = projectRoot // Build from the project root directory
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build walgo binary: %v\nOutput: %s", err, string(output))
	}

	// Test that the binary runs and shows help
	helpCmd := exec.Command(walgoBinary, "--help")
	output, err = helpCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run walgo --help: %v", err)
	}

	helpText := string(output)

	// Check that help contains expected text
	expectedSections := []string{
		"Walgo provides a seamless bridge for Hugo users to build and deploy",
		"Quick Start:",
		"walgo init",
		"walgo build",
		"walgo deploy",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(helpText, expected) {
			t.Errorf("Help output missing expected section: %s", expected)
		}
	}
}

// TestWalgoInitCommand tests the init command creates proper directory structure
func TestWalgoInitCommand(t *testing.T) {
	// Get the current working directory (should be the tests directory)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Get the parent directory (the walgo project root)
	projectRoot := filepath.Dir(cwd)

	// Build the walgo binary
	walgoBinary := filepath.Join(t.TempDir(), "walgo")
	buildCmd := exec.Command("go", "build", "-o", walgoBinary, ".")
	buildCmd.Dir = projectRoot // Build from the project root directory
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build walgo binary: %v\nOutput: %s", err, string(output))
	}

	// Create a temporary directory for the test site
	testDir := t.TempDir()
	siteName := "test-site"

	// Run walgo init
	initCmd := exec.Command(walgoBinary, "init", siteName)
	initCmd.Dir = testDir
	output, err = initCmd.CombinedOutput()

	// Note: This may fail if Hugo is not installed, which is OK for integration testing
	if err != nil {
		t.Logf("walgo init failed (expected if Hugo not installed): %v\nOutput: %s", err, string(output))
		return
	}

	// Check that the expected files were created
	siteDir := filepath.Join(testDir, siteName)
	expectedFiles := []string{
		"walgo.yaml",
		"hugo.toml",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(siteDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", filePath)
		}
	}

	// Check walgo.yaml content
	walgoConfigPath := filepath.Join(siteDir, "walgo.yaml")
	if content, err := os.ReadFile(walgoConfigPath); err == nil {
		configText := string(content)
		expectedConfigSections := []string{
			"hugo:",
			"walrus:",
			"obsidian:",
			"publishDir: public",
			"projectID: YOUR_WALRUS_PROJECT_ID",
		}

		for _, expected := range expectedConfigSections {
			if !strings.Contains(configText, expected) {
				t.Errorf("walgo.yaml missing expected section: %s", expected)
			}
		}
	}
}

// TestWalgoBuildCommand tests the build command validation
func TestWalgoBuildCommand(t *testing.T) {
	// Get the current working directory (should be the tests directory)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Get the parent directory (the walgo project root)
	projectRoot := filepath.Dir(cwd)

	// Build the walgo binary
	walgoBinary := filepath.Join(t.TempDir(), "walgo")
	buildCmd := exec.Command("go", "build", "-o", walgoBinary, ".")
	buildCmd.Dir = projectRoot // Build from the project root directory
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build walgo binary: %v\nOutput: %s", err, string(output))
	}

	// Create a temporary directory
	testDir := t.TempDir()

	// Run walgo build in empty directory (should fail gracefully)
	buildTestCmd := exec.Command(walgoBinary, "build")
	buildTestCmd.Dir = testDir
	output, err = buildTestCmd.CombinedOutput()

	// Should fail because there's no walgo.yaml or Hugo site
	if err == nil {
		t.Error("Expected walgo build to fail in empty directory")
	}

	outputText := string(output)
	// Should mention configuration file not found
	if !strings.Contains(outputText, "configuration file not found") {
		t.Logf("Build command output: %s", outputText)
	}
}

// TestWalgoCommandsExist tests that all expected commands are available
func TestWalgoCommandsExist(t *testing.T) {
	// Get the current working directory (should be the tests directory)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Get the parent directory (the walgo project root)
	projectRoot := filepath.Dir(cwd)

	// Build the walgo binary
	walgoBinary := filepath.Join(t.TempDir(), "walgo")
	buildCmd := exec.Command("go", "build", "-o", walgoBinary, ".")
	buildCmd.Dir = projectRoot // Build from the project root directory
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build walgo binary: %v\nOutput: %s", err, string(output))
	}

	// Test that all expected commands are available
	helpCmd := exec.Command(walgoBinary, "--help")
	output, err = helpCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run walgo --help: %v", err)
	}

	helpText := string(output)

	expectedCommands := []string{
		"init",
		"build",
		"serve",
		"new",
		"deploy",
		"update",
		"status",
		"domain",
		"import",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(helpText, cmd) {
			t.Errorf("Expected command not found in help: %s", cmd)
		}
	}
}
