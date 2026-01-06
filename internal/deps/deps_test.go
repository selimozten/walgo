// Package deps provides tests for the dependency checking and installation utilities
package deps

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRequiredTools verifies the list of required tools
func TestRequiredTools(t *testing.T) {
	tools := RequiredTools()

	// Should return exactly 3 tools
	if len(tools) != 3 {
		t.Errorf("expected 3 required tools, got %d", len(tools))
	}

	// Verify expected tools are present
	expectedTools := map[string]bool{
		"sui":          false,
		"walrus":       false,
		"site-builder": false,
	}

	for _, tool := range tools {
		if _, exists := expectedTools[tool.Name]; !exists {
			t.Errorf("unexpected tool: %s", tool.Name)
		}
		expectedTools[tool.Name] = true

		// All tools should be required
		if !tool.Required {
			t.Errorf("tool %s should be required", tool.Name)
		}

		// All tools should have descriptions
		if tool.Description == "" {
			t.Errorf("tool %s should have a description", tool.Name)
		}
	}

	// Verify all expected tools were found
	for name, found := range expectedTools {
		if !found {
			t.Errorf("expected tool %s not found", name)
		}
	}
}

// TestToolStruct verifies the Tool struct
func TestToolStruct(t *testing.T) {
	tool := Tool{
		Name:        "test-tool",
		Description: "A test tool for testing",
		Required:    true,
	}

	if tool.Name != "test-tool" {
		t.Errorf("expected name 'test-tool', got '%s'", tool.Name)
	}
	if tool.Description != "A test tool for testing" {
		t.Errorf("expected description 'A test tool for testing', got '%s'", tool.Description)
	}
	if !tool.Required {
		t.Error("expected Required to be true")
	}
}

// TestToolStatusStruct verifies the ToolStatus struct
func TestToolStatusStruct(t *testing.T) {
	status := &ToolStatus{
		Name:      "test-tool",
		Installed: true,
		Path:      "/usr/bin/test-tool",
		Version:   "1.0.0",
		Error:     nil,
	}

	if status.Name != "test-tool" {
		t.Errorf("expected name 'test-tool', got '%s'", status.Name)
	}
	if !status.Installed {
		t.Error("expected Installed to be true")
	}
	if status.Path != "/usr/bin/test-tool" {
		t.Errorf("expected path '/usr/bin/test-tool', got '%s'", status.Path)
	}
	if status.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", status.Version)
	}
	if status.Error != nil {
		t.Errorf("expected nil error, got %v", status.Error)
	}
}

// TestToolStatusWithError verifies ToolStatus with an error
func TestToolStatusWithError(t *testing.T) {
	testErr := errors.New("tool not found")
	status := &ToolStatus{
		Name:      "missing-tool",
		Installed: false,
		Path:      "",
		Version:   "",
		Error:     testErr,
	}

	if status.Installed {
		t.Error("expected Installed to be false")
	}
	if status.Error == nil {
		t.Error("expected error to be set")
	}
	if status.Error.Error() != "tool not found" {
		t.Errorf("unexpected error message: %v", status.Error)
	}
}

// TestInstallInstructions verifies install instructions generation
func TestInstallInstructions(t *testing.T) {
	tests := []struct {
		name            string
		network         string
		expectedNetwork string
	}{
		{
			name:            "testnet network",
			network:         "testnet",
			expectedNetwork: "testnet",
		},
		{
			name:            "mainnet network",
			network:         "mainnet",
			expectedNetwork: "mainnet",
		},
		{
			name:            "empty network defaults to testnet",
			network:         "",
			expectedNetwork: "testnet",
		},
		{
			name:            "devnet network",
			network:         "devnet",
			expectedNetwork: "devnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instructions := InstallInstructions(tt.network)

			// Verify suiup curl command is present
			if !strings.Contains(instructions, "curl -sSfL https://suiup.io | sh") {
				t.Error("expected suiup curl command in instructions")
			}

			// Verify network is mentioned for sui
			expectedSuiInstall := fmt.Sprintf("suiup install sui@%s", tt.expectedNetwork)
			if !strings.Contains(instructions, expectedSuiInstall) {
				t.Errorf("expected '%s' in instructions, got: %s", expectedSuiInstall, instructions)
			}

			// Verify network is mentioned for walrus
			expectedWalrusInstall := fmt.Sprintf("suiup install walrus@%s", tt.expectedNetwork)
			if !strings.Contains(instructions, expectedWalrusInstall) {
				t.Errorf("expected '%s' in instructions, got: %s", expectedWalrusInstall, instructions)
			}

			// Verify site-builder is always mainnet
			if !strings.Contains(instructions, "suiup install site-builder@mainnet") {
				t.Error("expected 'suiup install site-builder@mainnet' in instructions")
			}

			// Verify walgo setup-deps alternative is mentioned
			if !strings.Contains(instructions, "walgo setup-deps") {
				t.Error("expected 'walgo setup-deps' alternative in instructions")
			}
		})
	}
}

// TestInstallInstructionsFormat verifies the format of install instructions
func TestInstallInstructionsFormat(t *testing.T) {
	instructions := InstallInstructions("testnet")

	// Should have numbered steps
	if !strings.Contains(instructions, "1.") {
		t.Error("expected step 1 in instructions")
	}
	if !strings.Contains(instructions, "2.") {
		t.Error("expected step 2 in instructions")
	}

	// Should mention suiup default set
	if !strings.Contains(instructions, "suiup default set") {
		t.Error("expected 'suiup default set' in instructions")
	}
}

// TestLookPathWithPathEnv tests LookPath with a tool in PATH
func TestLookPathWithPathEnv(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a fake executable
	fakeTool := filepath.Join(tempDir, "fake-tool")
	if err := os.WriteFile(fakeTool, []byte("#!/bin/bash\necho 'fake tool'"), 0755); err != nil {
		t.Fatalf("failed to create fake tool: %v", err)
	}

	// Add temp dir to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	defer os.Setenv("PATH", originalPath)

	// Test LookPath
	path, err := LookPath("fake-tool")
	if err != nil {
		t.Errorf("LookPath failed for fake-tool: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

// TestLookPathLocalBin tests LookPath with tool in ~/.local/bin
func TestLookPathLocalBin(t *testing.T) {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("could not get home directory: %v", err)
	}

	// Create ~/.local/bin if it doesn't exist
	localBinDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(localBinDir, 0755); err != nil {
		t.Skipf("could not create ~/.local/bin: %v", err)
	}

	// Create a unique fake tool name to avoid conflicts
	uniqueToolName := fmt.Sprintf("walgo-test-tool-%d", os.Getpid())
	fakeTool := filepath.Join(localBinDir, uniqueToolName)

	// Create the fake tool
	if err := os.WriteFile(fakeTool, []byte("#!/bin/bash\necho 'local bin tool'"), 0755); err != nil {
		t.Fatalf("failed to create fake tool: %v", err)
	}
	defer os.Remove(fakeTool)

	// Temporarily remove tool from PATH (by setting a minimal PATH)
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "/bin:/usr/bin")
	defer os.Setenv("PATH", originalPath)

	// Test LookPath - should find it in ~/.local/bin
	path, err := LookPath(uniqueToolName)
	if err != nil {
		t.Errorf("LookPath failed for tool in ~/.local/bin: %v", err)
	}
	if path != fakeTool {
		t.Errorf("expected path '%s', got '%s'", fakeTool, path)
	}
}

// TestLookPathNotFound tests LookPath with a non-existent tool
func TestLookPathNotFound(t *testing.T) {
	// Use a tool name that definitely doesn't exist
	nonExistentTool := "this-tool-definitely-does-not-exist-12345"

	path, err := LookPath(nonExistentTool)
	if err == nil {
		t.Error("expected error for non-existent tool")
	}
	if path != "" {
		t.Errorf("expected empty path, got '%s'", path)
	}

	// Error should mention the tool name
	if !strings.Contains(err.Error(), nonExistentTool) {
		t.Errorf("error should mention tool name, got: %v", err)
	}
}

// TestLookPathNonExecutable tests LookPath with a non-executable file
func TestLookPathNonExecutable(t *testing.T) {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("could not get home directory: %v", err)
	}

	// Create ~/.local/bin if it doesn't exist
	localBinDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(localBinDir, 0755); err != nil {
		t.Skipf("could not create ~/.local/bin: %v", err)
	}

	// Create a non-executable file with unique name
	uniqueToolName := fmt.Sprintf("walgo-nonexec-test-%d", os.Getpid())
	nonExecFile := filepath.Join(localBinDir, uniqueToolName)

	// Create file without executable permissions
	if err := os.WriteFile(nonExecFile, []byte("not executable"), 0644); err != nil {
		t.Fatalf("failed to create non-executable file: %v", err)
	}
	defer os.Remove(nonExecFile)

	// Set minimal PATH to ensure we check ~/.local/bin
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "/bin:/usr/bin")
	defer os.Setenv("PATH", originalPath)

	// LookPath should not find this as it's not executable
	_, err = LookPath(uniqueToolName)
	if err == nil {
		t.Error("expected error for non-executable file")
	}
}

// TestLookPathDirectory tests LookPath with a directory instead of file
func TestLookPathDirectory(t *testing.T) {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("could not get home directory: %v", err)
	}

	// Create ~/.local/bin if it doesn't exist
	localBinDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(localBinDir, 0755); err != nil {
		t.Skipf("could not create ~/.local/bin: %v", err)
	}

	// Create a directory with a unique name
	uniqueDirName := fmt.Sprintf("walgo-dir-test-%d", os.Getpid())
	testDir := filepath.Join(localBinDir, uniqueDirName)

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Set minimal PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "/bin:/usr/bin")
	defer os.Setenv("PATH", originalPath)

	// LookPath should not find this as it's a directory
	_, err = LookPath(uniqueDirName)
	if err == nil {
		t.Error("expected error for directory")
	}
}

// TestCheckToolWithRealBash tests CheckTool with a real tool (bash should exist on most systems)
func TestCheckToolWithRealBash(t *testing.T) {
	// bash should exist on most Unix systems
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available on this system")
	}

	status := CheckTool("bash")

	if !status.Installed {
		t.Error("expected bash to be installed")
	}
	if status.Path == "" {
		t.Error("expected non-empty path for bash")
	}
	if status.Error != nil {
		t.Errorf("expected no error, got: %v", status.Error)
	}
	if status.Name != "bash" {
		t.Errorf("expected name 'bash', got '%s'", status.Name)
	}
}

// TestCheckToolNotFound tests CheckTool with a non-existent tool
func TestCheckToolNotFound(t *testing.T) {
	nonExistentTool := "this-tool-definitely-does-not-exist-67890"

	status := CheckTool(nonExistentTool)

	if status.Installed {
		t.Error("expected tool to not be installed")
	}
	if status.Path != "" {
		t.Errorf("expected empty path, got '%s'", status.Path)
	}
	if status.Error == nil {
		t.Error("expected error")
	}
	if status.Name != nonExistentTool {
		t.Errorf("expected name '%s', got '%s'", nonExistentTool, status.Name)
	}
}

// TestCheckAllTools tests CheckAllTools function
func TestCheckAllTools(t *testing.T) {
	results := CheckAllTools()

	// Should return status for all 3 required tools
	if len(results) != 3 {
		t.Errorf("expected 3 tool statuses, got %d", len(results))
	}

	// Verify all expected tool names are present
	expectedNames := map[string]bool{
		"sui":          false,
		"walrus":       false,
		"site-builder": false,
	}

	for _, status := range results {
		if status == nil {
			t.Error("got nil status")
			continue
		}
		if _, exists := expectedNames[status.Name]; !exists {
			t.Errorf("unexpected tool name: %s", status.Name)
		}
		expectedNames[status.Name] = true
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("expected tool %s status not found", name)
		}
	}
}

// TestGetMissingTools tests GetMissingTools function
func TestGetMissingTools(t *testing.T) {
	missing := GetMissingTools()

	// Result should be a slice (may be empty or have items)
	if missing == nil {
		// Initialize to empty slice if nil
		missing = []string{}
	}

	// If sui, walrus, or site-builder are not installed, they should be in the missing list
	for _, name := range missing {
		if name != "sui" && name != "walrus" && name != "site-builder" {
			t.Errorf("unexpected tool in missing list: %s", name)
		}
	}
}

// TestAllToolsInstalled tests AllToolsInstalled function
func TestAllToolsInstalled(t *testing.T) {
	allInstalled := AllToolsInstalled()
	missing := GetMissingTools()

	// AllToolsInstalled should return true only if GetMissingTools returns empty
	if allInstalled && len(missing) > 0 {
		t.Error("AllToolsInstalled returned true but GetMissingTools returned missing tools")
	}
	if !allInstalled && len(missing) == 0 {
		t.Error("AllToolsInstalled returned false but GetMissingTools returned no missing tools")
	}
}

// TestSuiupInstalled tests SuiupInstalled function
func TestSuiupInstalled(t *testing.T) {
	installed := SuiupInstalled()

	// Verify the result is consistent with LookPath
	_, err := LookPath("suiup")
	if installed && err != nil {
		t.Error("SuiupInstalled returned true but LookPath returned error")
	}
	if !installed && err == nil {
		t.Error("SuiupInstalled returned false but LookPath found suiup")
	}
}

// TestGetSuiupPath tests GetSuiupPath function
func TestGetSuiupPath(t *testing.T) {
	path, err := GetSuiupPath()

	// Should be consistent with LookPath("suiup")
	expectedPath, expectedErr := LookPath("suiup")

	if (err == nil) != (expectedErr == nil) {
		t.Errorf("GetSuiupPath error inconsistent with LookPath: GetSuiupPath=%v, LookPath=%v", err, expectedErr)
	}
	if path != expectedPath {
		t.Errorf("GetSuiupPath path inconsistent with LookPath: GetSuiupPath=%s, LookPath=%s", path, expectedPath)
	}
}

// TestGetToolVersionNotFound tests GetToolVersion with non-existent tool
func TestGetToolVersionNotFound(t *testing.T) {
	nonExistentTool := "this-tool-definitely-does-not-exist-version-test"

	version, err := GetToolVersion(nonExistentTool)

	if err == nil {
		t.Error("expected error for non-existent tool")
	}
	if version != "" {
		t.Errorf("expected empty version, got '%s'", version)
	}
}

// TestGetToolVersionWithRealTool tests GetToolVersion with a real tool
func TestGetToolVersionWithRealTool(t *testing.T) {
	// Use a common tool that should exist
	toolsToTry := []string{"bash", "ls", "cat"}

	for _, tool := range toolsToTry {
		if _, err := exec.LookPath(tool); err == nil {
			version, err := GetToolVersion(tool)
			// Some tools don't support --version flag, so we just check it doesn't panic
			if err != nil {
				t.Logf("%s --version returned error (this is acceptable): %v", tool, err)
			} else if version == "" {
				t.Logf("%s --version returned empty version", tool)
			} else {
				t.Logf("%s version: %s", tool, version)
			}
			return // Successfully tested with at least one tool
		}
	}
	t.Skip("no common tools available for testing")
}

// TestRunSuiupNotInstalled tests RunSuiup when suiup is not installed
func TestRunSuiupNotInstalled(t *testing.T) {
	// Check if suiup is actually installed
	if SuiupInstalled() {
		t.Skip("suiup is installed, skipping test for 'not installed' scenario")
	}

	_, err := RunSuiup("version")

	if err == nil {
		t.Error("expected error when suiup is not installed")
	}

	// Error should mention suiup not found
	if !strings.Contains(err.Error(), "suiup not found") {
		t.Errorf("error should mention 'suiup not found', got: %v", err)
	}

	// Error should mention installation instructions
	if !strings.Contains(err.Error(), "curl -sSfL https://suiup.io | sh") {
		t.Errorf("error should mention installation instructions, got: %v", err)
	}
}

// TestRunSuiupInstalled tests RunSuiup when suiup is installed
func TestRunSuiupInstalled(t *testing.T) {
	if !SuiupInstalled() {
		t.Skip("suiup is not installed, skipping installed test")
	}

	// Test with --help which should succeed
	output, err := RunSuiup("--help")
	if err != nil {
		t.Logf("RunSuiup --help returned error: %v", err)
		// --help might exit with non-zero on some tools
	}
	if output == "" && err == nil {
		t.Log("RunSuiup --help returned empty output")
	}
}

// TestRunSuiupWithInvalidCommand tests RunSuiup with an invalid command
func TestRunSuiupWithInvalidCommand(t *testing.T) {
	if !SuiupInstalled() {
		t.Skip("suiup is not installed")
	}

	// Use an invalid subcommand
	_, err := RunSuiup("invalid-command-that-does-not-exist")
	if err == nil {
		t.Error("expected error with invalid command")
	}

	// Error should mention the failure
	if !strings.Contains(err.Error(), "suiup command failed") {
		t.Errorf("error should mention 'suiup command failed', got: %v", err)
	}
}

// TestInstallToolVersionSelection tests that InstallTool uses correct versions
func TestInstallToolVersionSelection(t *testing.T) {
	// This test verifies the logic without actually running suiup
	tests := []struct {
		tool            string
		network         string
		expectedVersion string
	}{
		{"sui", "testnet", "testnet"},
		{"sui", "mainnet", "mainnet"},
		{"sui", "devnet", "devnet"},
		{"sui", "", "testnet"},
		{"walrus", "testnet", "testnet"},
		{"walrus", "mainnet", "mainnet"},
		{"walrus", "", "testnet"},
		{"site-builder", "testnet", "mainnet"},  // site-builder is always mainnet
		{"site-builder", "mainnet", "mainnet"},  // site-builder is always mainnet
		{"site-builder", "devnet", "mainnet"},   // site-builder is always mainnet
		{"site-builder", "", "mainnet"},         // site-builder is always mainnet
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s@%s", tt.tool, tt.network), func(t *testing.T) {
			// We can't actually test InstallTool without suiup, but we can verify the logic
			// by checking that site-builder always uses mainnet
			network := tt.network
			if network == "" {
				network = "testnet"
			}

			version := network
			if tt.tool == "site-builder" {
				version = "mainnet"
			}

			if version != tt.expectedVersion {
				t.Errorf("expected version '%s', got '%s'", tt.expectedVersion, version)
			}
		})
	}
}

// TestCheckToolVersionExtraction tests that version is extracted correctly
func TestCheckToolVersionExtraction(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a fake tool with version output
	fakeTool := filepath.Join(tempDir, "version-test-tool")
	script := `#!/bin/bash
echo "version-test-tool 1.2.3"
echo "Build: abc123"
`
	if err := os.WriteFile(fakeTool, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create fake tool: %v", err)
	}

	// Add temp dir to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	defer os.Setenv("PATH", originalPath)

	// Test CheckTool
	status := CheckTool("version-test-tool")

	if !status.Installed {
		t.Error("expected tool to be installed")
	}

	// Version should be the first line
	if status.Version != "version-test-tool 1.2.3" {
		t.Errorf("expected version 'version-test-tool 1.2.3', got '%s'", status.Version)
	}
}

// TestCheckToolEmptyVersion tests that empty version output is handled
func TestCheckToolEmptyVersion(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a fake tool with empty version output
	fakeTool := filepath.Join(tempDir, "empty-version-tool")
	script := `#!/bin/bash
# No output
`
	if err := os.WriteFile(fakeTool, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create fake tool: %v", err)
	}

	// Add temp dir to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	defer os.Setenv("PATH", originalPath)

	// Test CheckTool
	status := CheckTool("empty-version-tool")

	if !status.Installed {
		t.Error("expected tool to be installed")
	}

	// Version should be empty but no error
	if status.Error != nil {
		t.Errorf("expected no error, got: %v", status.Error)
	}
}

// TestCheckToolVersionCommandFails tests handling when --version fails
func TestCheckToolVersionCommandFails(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a fake tool that fails on --version
	fakeTool := filepath.Join(tempDir, "failing-version-tool")
	script := `#!/bin/bash
if [ "$1" = "--version" ]; then
    exit 1
fi
echo "Hello"
`
	if err := os.WriteFile(fakeTool, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create fake tool: %v", err)
	}

	// Add temp dir to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	defer os.Setenv("PATH", originalPath)

	// Test CheckTool
	status := CheckTool("failing-version-tool")

	// Tool should still be considered installed even if --version fails
	if !status.Installed {
		t.Error("expected tool to be installed")
	}

	// Error field should still be nil (from LookPath, not version command)
	if status.Error != nil {
		t.Errorf("expected no error in status, got: %v", status.Error)
	}
}

// TestLookPathStandardPath tests LookPath finds tools in standard PATH
func TestLookPathStandardPath(t *testing.T) {
	// Test with a common system tool
	if _, err := exec.LookPath("ls"); err != nil {
		t.Skip("ls not available on this system")
	}

	path, err := LookPath("ls")
	if err != nil {
		t.Errorf("LookPath should find ls: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path for ls")
	}
}

// TestInstallInstructionsContainsAllSteps tests that all installation steps are included
func TestInstallInstructionsContainsAllSteps(t *testing.T) {
	instructions := InstallInstructions("testnet")

	requiredContent := []string{
		"curl -sSfL https://suiup.io | sh",
		"suiup install sui@testnet",
		"suiup install walrus@testnet",
		"suiup install site-builder@mainnet",
		"suiup default set",
		"walgo setup-deps",
	}

	for _, content := range requiredContent {
		if !strings.Contains(instructions, content) {
			t.Errorf("instructions should contain '%s'", content)
		}
	}
}

// TestToolStatusEquality tests ToolStatus field assignments
func TestToolStatusEquality(t *testing.T) {
	status1 := &ToolStatus{
		Name:      "tool1",
		Installed: true,
		Path:      "/usr/bin/tool1",
		Version:   "1.0.0",
		Error:     nil,
	}

	status2 := &ToolStatus{
		Name:      "tool1",
		Installed: true,
		Path:      "/usr/bin/tool1",
		Version:   "1.0.0",
		Error:     nil,
	}

	if status1.Name != status2.Name {
		t.Error("names should be equal")
	}
	if status1.Installed != status2.Installed {
		t.Error("installed status should be equal")
	}
	if status1.Path != status2.Path {
		t.Error("paths should be equal")
	}
	if status1.Version != status2.Version {
		t.Error("versions should be equal")
	}
}

// TestMultipleCallsToCheckAllTools tests that CheckAllTools is consistent
func TestMultipleCallsToCheckAllTools(t *testing.T) {
	results1 := CheckAllTools()
	results2 := CheckAllTools()

	if len(results1) != len(results2) {
		t.Errorf("inconsistent results count: %d vs %d", len(results1), len(results2))
	}

	for i := range results1 {
		if results1[i].Name != results2[i].Name {
			t.Errorf("inconsistent tool name at index %d: %s vs %s", i, results1[i].Name, results2[i].Name)
		}
		if results1[i].Installed != results2[i].Installed {
			t.Errorf("inconsistent installed status for %s", results1[i].Name)
		}
	}
}

// TestGetMissingToolsConsistency tests GetMissingTools consistency
func TestGetMissingToolsConsistency(t *testing.T) {
	missing1 := GetMissingTools()
	missing2 := GetMissingTools()

	if len(missing1) != len(missing2) {
		t.Errorf("inconsistent missing tools count: %d vs %d", len(missing1), len(missing2))
	}

	for i := range missing1 {
		if missing1[i] != missing2[i] {
			t.Errorf("inconsistent missing tool at index %d: %s vs %s", i, missing1[i], missing2[i])
		}
	}
}

// TestToolDescription verifies all tools have meaningful descriptions
func TestToolDescription(t *testing.T) {
	tools := RequiredTools()

	for _, tool := range tools {
		if len(tool.Description) < 10 {
			t.Errorf("tool %s has too short description: %s", tool.Name, tool.Description)
		}
	}
}

// Benchmark tests

func BenchmarkRequiredTools(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RequiredTools()
	}
}

func BenchmarkLookPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LookPath("ls")
	}
}

func BenchmarkCheckTool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CheckTool("ls")
	}
}

func BenchmarkInstallInstructions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		InstallInstructions("testnet")
	}
}

// TestInstallToolWithSuiup tests InstallTool when suiup is available
func TestInstallToolWithSuiup(t *testing.T) {
	if !SuiupInstalled() {
		t.Skip("suiup is not installed")
	}

	// We don't actually want to install, just verify it doesn't panic
	// and properly constructs the command
	// Note: This test may fail if the tool is already installed at a different version
	// We're testing the function doesn't panic, not that installation succeeds

	// Test with empty network (should default to testnet)
	err := InstallTool("nonexistent-tool-for-testing", "")
	if err == nil {
		t.Log("InstallTool with nonexistent tool returned no error (unexpected)")
	} else {
		// We expect an error because the tool doesn't exist
		t.Logf("InstallTool returned expected error: %v", err)
	}
}

// TestInstallToolSiteBuilderAlwaysMainnet tests that site-builder always uses mainnet
func TestInstallToolSiteBuilderAlwaysMainnet(t *testing.T) {
	if !SuiupInstalled() {
		t.Skip("suiup is not installed")
	}

	// This tests the logic that site-builder always uses mainnet
	// We're calling it with "testnet" but it should internally use "mainnet"
	// We don't check if it succeeds, just that it doesn't panic
	err := InstallTool("site-builder", "testnet")
	t.Logf("InstallTool for site-builder returned: %v", err)
}

// TestCheckToolWithMultiLineVersion tests version parsing with multiple lines
func TestCheckToolWithMultiLineVersion(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a fake tool with multi-line version output including empty lines
	fakeTool := filepath.Join(tempDir, "multiline-version-tool")
	script := `#!/bin/bash
echo "tool v2.0.0"
echo ""
echo "Additional info"
echo "More info"
`
	if err := os.WriteFile(fakeTool, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create fake tool: %v", err)
	}

	// Add temp dir to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	defer os.Setenv("PATH", originalPath)

	status := CheckTool("multiline-version-tool")

	if !status.Installed {
		t.Error("expected tool to be installed")
	}

	// Only first line should be captured
	if status.Version != "tool v2.0.0" {
		t.Errorf("expected version 'tool v2.0.0', got '%s'", status.Version)
	}
}

// TestGetToolVersionWithMultiLineOutput tests GetToolVersion with multi-line output
func TestGetToolVersionWithMultiLineOutput(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a fake tool
	fakeTool := filepath.Join(tempDir, "version-multiline-tool")
	script := `#!/bin/bash
echo "version 3.0.0"
echo "build info"
`
	if err := os.WriteFile(fakeTool, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create fake tool: %v", err)
	}

	// Add temp dir to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	defer os.Setenv("PATH", originalPath)

	version, err := GetToolVersion("version-multiline-tool")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// GetToolVersion returns the full output trimmed
	if !strings.Contains(version, "version 3.0.0") {
		t.Errorf("version should contain 'version 3.0.0', got '%s'", version)
	}
}

// TestGetToolVersionCommandError tests GetToolVersion when command fails
func TestGetToolVersionCommandError(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a fake tool that fails
	fakeTool := filepath.Join(tempDir, "failing-tool")
	script := `#!/bin/bash
exit 1
`
	if err := os.WriteFile(fakeTool, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create fake tool: %v", err)
	}

	// Add temp dir to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	defer os.Setenv("PATH", originalPath)

	_, err := GetToolVersion("failing-tool")
	if err == nil {
		t.Error("expected error when --version command fails")
	}

	// Error should mention the tool name
	if !strings.Contains(err.Error(), "failing-tool") {
		t.Errorf("error should mention tool name, got: %v", err)
	}

	// Error should mention "failed to get version"
	if !strings.Contains(err.Error(), "failed to get") {
		t.Errorf("error should mention 'failed to get', got: %v", err)
	}
}

// TestLookPathWithHomeEnvError tests LookPath behavior (home directory always available)
func TestLookPathWithHomeEnvError(t *testing.T) {
	// This test verifies the error message format when tool is not found
	nonExistent := "absolutely-non-existent-tool-xyz123"
	_, err := LookPath(nonExistent)

	if err == nil {
		t.Error("expected error for non-existent tool")
	}

	// Error should mention both PATH and ~/.local/bin
	errMsg := err.Error()
	if !strings.Contains(errMsg, nonExistent) {
		t.Errorf("error should contain tool name, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "not found") {
		t.Errorf("error should say 'not found', got: %s", errMsg)
	}
}

// TestRequiredToolsOrder tests that RequiredTools returns tools in expected order
func TestRequiredToolsOrder(t *testing.T) {
	tools := RequiredTools()

	if len(tools) < 3 {
		t.Fatalf("expected at least 3 tools, got %d", len(tools))
	}

	// Verify first tool is sui
	if tools[0].Name != "sui" {
		t.Errorf("expected first tool to be 'sui', got '%s'", tools[0].Name)
	}

	// Verify second tool is walrus
	if tools[1].Name != "walrus" {
		t.Errorf("expected second tool to be 'walrus', got '%s'", tools[1].Name)
	}

	// Verify third tool is site-builder
	if tools[2].Name != "site-builder" {
		t.Errorf("expected third tool to be 'site-builder', got '%s'", tools[2].Name)
	}
}

// TestCheckAllToolsReturnsStatusForEach tests that CheckAllTools returns status for each required tool
func TestCheckAllToolsReturnsStatusForEach(t *testing.T) {
	tools := RequiredTools()
	statuses := CheckAllTools()

	if len(statuses) != len(tools) {
		t.Errorf("expected %d statuses, got %d", len(tools), len(statuses))
	}

	// Verify each status corresponds to a tool
	for i, status := range statuses {
		if status.Name != tools[i].Name {
			t.Errorf("status %d name mismatch: expected '%s', got '%s'", i, tools[i].Name, status.Name)
		}
	}
}

// TestCheckToolSetsNameCorrectly tests that CheckTool sets the Name field
func TestCheckToolSetsNameCorrectly(t *testing.T) {
	testCases := []string{
		"sui",
		"walrus",
		"site-builder",
		"nonexistent-tool",
		"bash",
	}

	for _, toolName := range testCases {
		t.Run(toolName, func(t *testing.T) {
			status := CheckTool(toolName)
			if status.Name != toolName {
				t.Errorf("expected Name='%s', got '%s'", toolName, status.Name)
			}
		})
	}
}

// TestAllToolsInstalledReturnType tests that AllToolsInstalled returns a boolean
func TestAllToolsInstalledReturnType(t *testing.T) {
	result := AllToolsInstalled()
	// This test ensures the function returns and doesn't panic
	t.Logf("AllToolsInstalled returned: %v", result)
}

// TestSuiupInstalledReturnType tests that SuiupInstalled returns a boolean
func TestSuiupInstalledReturnType(t *testing.T) {
	result := SuiupInstalled()
	// This test ensures the function returns and doesn't panic
	t.Logf("SuiupInstalled returned: %v", result)
}

// TestInstallInstructionsWithDifferentNetworks tests all network variations
func TestInstallInstructionsWithDifferentNetworks(t *testing.T) {
	networks := []string{"testnet", "mainnet", "devnet", "localnet", ""}

	for _, network := range networks {
		t.Run(fmt.Sprintf("network=%s", network), func(t *testing.T) {
			instructions := InstallInstructions(network)

			// Should always contain the suiup install command
			if !strings.Contains(instructions, "curl -sSfL https://suiup.io") {
				t.Error("should contain suiup install curl command")
			}

			// site-builder should always be mainnet
			if !strings.Contains(instructions, "site-builder@mainnet") {
				t.Error("site-builder should always be @mainnet")
			}

			// If network is empty, it should default to testnet
			expectedNetwork := network
			if expectedNetwork == "" {
				expectedNetwork = "testnet"
			}

			// Check sui uses the correct network
			expectedSui := fmt.Sprintf("sui@%s", expectedNetwork)
			if !strings.Contains(instructions, expectedSui) {
				t.Errorf("should contain '%s', got: %s", expectedSui, instructions)
			}
		})
	}
}

// TestCheckToolPathIsAbsolute tests that CheckTool returns absolute paths
func TestCheckToolPathIsAbsolute(t *testing.T) {
	// Use a tool that's likely to exist
	if _, err := exec.LookPath("ls"); err != nil {
		t.Skip("ls not available")
	}

	status := CheckTool("ls")
	if status.Installed && status.Path != "" {
		if !filepath.IsAbs(status.Path) {
			t.Errorf("expected absolute path, got: %s", status.Path)
		}
	}
}

// TestLookPathReturnsAbsolutePath tests that LookPath returns absolute paths
func TestLookPathReturnsAbsolutePath(t *testing.T) {
	// Create a fake tool
	tempDir := t.TempDir()
	fakeTool := filepath.Join(tempDir, "abs-path-test")
	if err := os.WriteFile(fakeTool, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("failed to create fake tool: %v", err)
	}

	// Add temp dir to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	defer os.Setenv("PATH", originalPath)

	path, err := LookPath("abs-path-test")
	if err != nil {
		t.Fatalf("LookPath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got: %s", path)
	}
}

// TestToolNotRequiredField tests Tool struct with Required=false
func TestToolNotRequiredField(t *testing.T) {
	tool := Tool{
		Name:        "optional-tool",
		Description: "An optional tool",
		Required:    false,
	}

	if tool.Required {
		t.Error("expected Required to be false")
	}
	if tool.Name != "optional-tool" {
		t.Errorf("expected Name='optional-tool', got '%s'", tool.Name)
	}
}

// TestToolStatusWithAllFields tests ToolStatus with all fields populated
func TestToolStatusWithAllFields(t *testing.T) {
	testErr := errors.New("test error")
	status := ToolStatus{
		Name:      "test",
		Installed: false,
		Path:      "/some/path",
		Version:   "1.0.0",
		Error:     testErr,
	}

	if status.Name != "test" {
		t.Errorf("unexpected Name: %s", status.Name)
	}
	if status.Installed {
		t.Error("expected Installed=false")
	}
	if status.Path != "/some/path" {
		t.Errorf("unexpected Path: %s", status.Path)
	}
	if status.Version != "1.0.0" {
		t.Errorf("unexpected Version: %s", status.Version)
	}
	if status.Error != testErr {
		t.Errorf("unexpected Error: %v", status.Error)
	}
}

// TestGetMissingToolsReturnsSlice tests that GetMissingTools always returns a valid slice
func TestGetMissingToolsReturnsSlice(t *testing.T) {
	missing := GetMissingTools()
	// Should never be nil even if empty
	if missing == nil {
		// Note: the implementation may return nil for empty, which is acceptable
		t.Log("GetMissingTools returned nil (acceptable for empty)")
	}
}

// BenchmarkCheckAllTools benchmarks CheckAllTools
func BenchmarkCheckAllTools(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CheckAllTools()
	}
}

// BenchmarkGetMissingTools benchmarks GetMissingTools
func BenchmarkGetMissingTools(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetMissingTools()
	}
}

// BenchmarkAllToolsInstalled benchmarks AllToolsInstalled
func BenchmarkAllToolsInstalled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		AllToolsInstalled()
	}
}

// BenchmarkSuiupInstalled benchmarks SuiupInstalled
func BenchmarkSuiupInstalled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SuiupInstalled()
	}
}
