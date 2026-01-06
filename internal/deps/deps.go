// Package deps provides centralized dependency checking and installation utilities
// for Walgo's required tools: sui, walrus, and site-builder (via suiup).
package deps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Tool represents a required CLI tool
type Tool struct {
	Name        string
	Description string
	Required    bool
}

// RequiredTools returns the list of tools required for Walgo
func RequiredTools() []Tool {
	return []Tool{
		{Name: "sui", Description: "Sui CLI for blockchain operations", Required: true},
		{Name: "walrus", Description: "Walrus CLI for storage operations", Required: true},
		{Name: "site-builder", Description: "Site-builder for Walrus Sites", Required: true},
		// Hugo removed - users should install via package manager (brew/apt/choco)
	}
}

// ToolStatus represents the installation status of a tool
type ToolStatus struct {
	Name      string
	Installed bool
	Path      string
	Version   string
	Error     error
}

// CheckTool checks if a single tool is available
func CheckTool(name string) *ToolStatus {
	status := &ToolStatus{Name: name}

	path, err := LookPath(name)
	if err != nil {
		status.Installed = false
		status.Error = err
		return status
	}

	status.Installed = true
	status.Path = path

	// Try to get version
	cmd := exec.Command(path, "--version")
	output, err := cmd.CombinedOutput()
	if err == nil {
		// Extract first line of version output
		lines := strings.Split(string(output), "\n")
		if len(lines) > 0 {
			status.Version = strings.TrimSpace(lines[0])
		}
	}

	return status
}

// CheckAllTools checks all required tools and returns their status
func CheckAllTools() []*ToolStatus {
	tools := RequiredTools()
	results := make([]*ToolStatus, len(tools))

	for i, tool := range tools {
		results[i] = CheckTool(tool.Name)
	}

	return results
}

// GetMissingTools returns a list of missing tool names
func GetMissingTools() []string {
	var missing []string
	for _, status := range CheckAllTools() {
		if !status.Installed {
			missing = append(missing, status.Name)
		}
	}
	return missing
}

// AllToolsInstalled returns true if all required tools are installed
func AllToolsInstalled() bool {
	return len(GetMissingTools()) == 0
}

// LookPath finds an executable in PATH or ~/.local/bin (where suiup installs)
func LookPath(name string) (string, error) {
	// First try standard PATH
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	// Fallback to ~/.local/bin (where suiup installs binaries)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH", name)
	}

	localBinPath := filepath.Join(home, ".local", "bin", name)
	if info, err := os.Stat(localBinPath); err == nil && !info.IsDir() {
		// Check if executable
		if info.Mode()&0111 != 0 {
			return localBinPath, nil
		}
	}

	return "", fmt.Errorf("%s not found in PATH or ~/.local/bin", name)
}

// InstallInstructions returns installation instructions for missing tools
func InstallInstructions(network string) string {
	if network == "" {
		network = "testnet"
	}

	return fmt.Sprintf(`To install missing dependencies:

1. Install suiup:
   curl -sSfL https://raw.githubusercontent.com/Mystenlabs/suiup/main/install.sh | sh

2. Install tools via suiup:
   suiup install sui@%s
   suiup install walrus@%s
   suiup install site-builder@mainnet
   suiup default set sui@%s walrus@%s site-builder@mainnet

Or run: walgo setup-deps
`, network, network, network, network)
}

// SuiupInstalled checks if suiup is installed
func SuiupInstalled() bool {
	_, err := LookPath("suiup")
	return err == nil
}

// GetSuiupPath returns the path to suiup binary
func GetSuiupPath() (string, error) {
	return LookPath("suiup")
}

// GetToolVersion returns the version string for a tool
func GetToolVersion(name string) (string, error) {
	path, err := LookPath(name)
	if err != nil {
		return "", err
	}

	// Get version
	cmd := exec.Command(path, "--version")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get %s version: %w", name, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// InstallSuiup installs suiup automatically
func InstallSuiup() error {
	// Check if already installed
	if SuiupInstalled() {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	localBin := filepath.Join(home, ".local", "bin")

	// Create ~/.local/bin if it doesn't exist
	if err := os.MkdirAll(localBin, 0755); err != nil {
		return fmt.Errorf("failed to create ~/.local/bin: %w", err)
	}

	// Download and execute suiup installer
	installScript := "https://raw.githubusercontent.com/Mystenlabs/suiup/main/install.sh"

	// Use curl or wget to download and execute
	var cmd *exec.Cmd
	if _, err := exec.LookPath("curl"); err == nil {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("curl -sSfL %s | sh", installScript))
	} else if _, err := exec.LookPath("wget"); err == nil {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("wget -qO- %s | sh", installScript))
	} else {
		return fmt.Errorf("neither curl nor wget found. Please install one of them")
	}

	// Set HOME environment variable and add ~/.local/bin to PATH
	currentPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%s:%s", localBin, currentPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", home),
		fmt.Sprintf("PATH=%s", newPath),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("suiup installation failed: %w\nOutput: %s", err, string(output))
	}

	// Update PATH for current process
	os.Setenv("PATH", newPath)

	// Verify installation
	if !SuiupInstalled() {
		return fmt.Errorf("suiup installation completed but binary not found in PATH")
	}

	return nil
}

// RunSuiup executes a suiup command with the given arguments
// Automatically installs suiup if not found
func RunSuiup(args ...string) (string, error) {
	suiupPath, err := GetSuiupPath()
	if err != nil {
		// Try to install suiup automatically
		if installErr := InstallSuiup(); installErr != nil {
			return "", fmt.Errorf("suiup not found and auto-installation failed: %w", installErr)
		}

		// Try to get path again after installation
		suiupPath, err = GetSuiupPath()
		if err != nil {
			return "", fmt.Errorf("suiup installation succeeded but binary not found: %w", err)
		}
	}

	cmd := exec.Command(suiupPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("suiup command failed: %w\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// InstallTool installs a tool using suiup
func InstallTool(tool, network string) error {
	if network == "" {
		network = "testnet"
	}

	// site-builder is always mainnet (no testnet version exists)
	version := network
	if tool == "site-builder" {
		version = "mainnet"
	}

	// Get suiup path
	suiupPath, err := GetSuiupPath()
	if err != nil {
		// If suiup not found, try to install it
		if installErr := InstallSuiup(); installErr != nil {
			return fmt.Errorf("suiup not found and auto-installation failed: %w", installErr)
		}
		// Try to get path again
		suiupPath, err = GetSuiupPath()
		if err != nil {
			return fmt.Errorf("suiup installation succeeded but binary not found: %w", err)
		}
	}

	// Ensure PATH includes ~/.local/bin
	home, _ := os.UserHomeDir()
	localBin := filepath.Join(home, ".local", "bin")
	currentPath := os.Getenv("PATH")
	envVars := os.Environ()
	if !strings.Contains(currentPath, localBin) {
		envVars = append(envVars, fmt.Sprintf("PATH=%s:%s", localBin, currentPath))
	}

	// Step 1: Install the tool
	installCmd := exec.Command(suiupPath, "install", fmt.Sprintf("%s@%s", tool, version))
	installCmd.Env = envVars

	output, err := installCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", tool, err, string(output))
	}

	// Step 2: Set as default (creates symlink in ~/.local/bin/)
	// This is critical - without this, the binary won't be accessible
	defaultCmd := exec.Command(suiupPath, "default", "set", fmt.Sprintf("%s@%s", tool, version))
	defaultCmd.Env = envVars

	defaultOutput, err := defaultCmd.CombinedOutput()
	if err != nil {
		// Don't fail if default set fails, just log it
		fmt.Fprintf(os.Stderr, "Warning: failed to set default for %s: %v\nOutput: %s\n", tool, err, string(defaultOutput))
	}

	// Step 3: Verify the binary exists
	toolPath := filepath.Join(localBin, tool)
	if _, err := os.Stat(toolPath); os.IsNotExist(err) {
		// Retry default set
		retryCmd := exec.Command(suiupPath, "default", "set", fmt.Sprintf("%s@%s", tool, version))
		retryCmd.Env = envVars
		retryCmd.CombinedOutput() // Ignore error

		// Check again
		if _, err := os.Stat(toolPath); os.IsNotExist(err) {
			return fmt.Errorf("%s binary not found in %s after installation", tool, localBin)
		}
	}

	return nil
}
