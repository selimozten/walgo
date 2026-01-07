// Package deps provides centralized dependency checking and installation utilities
// for Walgo's required tools: sui, walrus, and site-builder (via suiup).
package deps

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	suiupRepo     = "MystenLabs/suiup"
	releaseAPIURL = "https://api.github.com/repos/%s/releases/latest"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func execCandidates(name string) []string {
	candidates := []string{name}
	if runtime.GOOS == "windows" && filepath.Ext(name) == "" {
		candidates = append([]string{name + ".exe"}, candidates...)
	}
	return candidates
}

func samePath(a, b string) bool {
	cleanA := filepath.Clean(a)
	cleanB := filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(cleanA, cleanB)
	}
	return cleanA == cleanB
}

func pathListContains(list, dir string) bool {
	for _, part := range filepath.SplitList(list) {
		if samePath(part, dir) {
			return true
		}
	}
	return false
}

func ensureLocalBinOnPath(localBin string) {
	if localBin == "" {
		return
	}
	current := os.Getenv("PATH")
	if pathListContains(current, localBin) {
		return
	}
	if current == "" {
		_ = os.Setenv("PATH", localBin)
		return
	}
	updated := fmt.Sprintf("%s%c%s", localBin, os.PathListSeparator, current)
	_ = os.Setenv("PATH", updated)
}

func getLocalBinDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".local", "bin"), nil
}

func extraBinaryDirs() []string {
	var dirs []string
	switch runtime.GOOS {
	case "darwin":
		dirs = []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin"}
	case "linux":
		dirs = []string{"/usr/local/bin", "/usr/bin", "/bin"}
	case "windows":
		// Windows installers typically add to PATH already
	}
	return dirs
}

func fetchLatestTag(repo string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf(releaseAPIURL, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "walgo-installer")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to query %s: %s", repo, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("no releases found for %s", repo)
	}
	return release.TagName, nil
}

func downloadToFile(url, destination string) error {
	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "walgo-installer")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: %s", url, resp.Status)
	}
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}
	return nil
}

func installSuiupWindows() error {
	localBin, err := getLocalBinDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(localBin, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", localBin, err)
	}
	ensureLocalBinOnPath(localBin)

	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return fmt.Errorf("unsupported architecture for suiup: %s", runtime.GOARCH)
	}

	tag, err := fetchLatestTag(suiupRepo)
	if err != nil {
		return fmt.Errorf("failed to fetch suiup release: %w", err)
	}

	filename := fmt.Sprintf("suiup-windows-%s.zip", arch)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", suiupRepo, tag, filename)
	tmpFile, err := os.CreateTemp("", "suiup-*.zip")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpName)

	if err := downloadToFile(url, tmpName); err != nil {
		return err
	}

	r, err := zip.OpenReader(tmpName)
	if err != nil {
		return err
	}
	defer r.Close()

	var extracted bool
	for _, f := range r.File {
		if strings.EqualFold(filepath.Base(f.Name), "suiup.exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			dest := filepath.Join(localBin, "suiup.exe")
			out, err := os.Create(dest)
			if err != nil {
				rc.Close()
				return err
			}
			if _, err := io.Copy(out, rc); err != nil {
				out.Close()
				rc.Close()
				return err
			}
			out.Close()
			rc.Close()
			extracted = true
			break
		}
	}

	if !extracted {
		return fmt.Errorf("suiup.exe not found in downloaded archive")
	}

	return nil
}

func installSuiupPosix() error {
	localBin, err := getLocalBinDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(localBin, 0755); err != nil {
		return fmt.Errorf("failed to create ~/.local/bin: %w", err)
	}
	ensureLocalBinOnPath(localBin)

	installScript := "https://raw.githubusercontent.com/Mystenlabs/suiup/main/install.sh"

	var cmd *exec.Cmd
	if _, err := exec.LookPath("curl"); err == nil {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("curl -sSfL %s | sh", installScript))
	} else if _, err := exec.LookPath("wget"); err == nil {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("wget -qO- %s | sh", installScript))
	} else {
		return fmt.Errorf("neither curl nor wget found. Please install one of them")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("HOME=%s", home))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("suiup installation failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

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
	for _, candidate := range execCandidates(name) {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}

	localBin, err := getLocalBinDir()
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH", name)
	}

	for _, candidate := range execCandidates(name) {
		localBinPath := filepath.Join(localBin, candidate)
		if info, err := os.Stat(localBinPath); err == nil && !info.IsDir() {
			if runtime.GOOS == "windows" || info.Mode()&0111 != 0 {
				return localBinPath, nil
			}
		}
	}

	for _, dir := range extraBinaryDirs() {
		for _, candidate := range execCandidates(name) {
			candidatePath := filepath.Join(dir, candidate)
			if info, err := os.Stat(candidatePath); err == nil && !info.IsDir() {
				if runtime.GOOS == "windows" || info.Mode()&0111 != 0 {
					return candidatePath, nil
				}
			}
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

// CheckHugoExtended checks if Hugo Extended is installed
// Returns: (isInstalled bool, isExtended bool, version string, error)
func CheckHugoExtended() (bool, bool, string, error) {
	// Check if Hugo is installed
	path, err := LookPath("hugo")
	if err != nil {
		return false, false, "", err
	}

	// Get Hugo version
	cmd := exec.Command(path, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return true, false, "", fmt.Errorf("failed to get hugo version: %w", err)
	}

	version := strings.TrimSpace(string(output))

	// Check if Extended version is installed
	isExtended := strings.Contains(strings.ToLower(version), "extended")

	return true, isExtended, version, nil
}

// InstallSuiup installs suiup automatically
func InstallSuiup() error {
	// Check if already installed
	if SuiupInstalled() {
		return nil
	}

	var err error
	switch runtime.GOOS {
	case "windows":
		err = installSuiupWindows()
	default:
		err = installSuiupPosix()
	}
	if err != nil {
		return err
	}

	localBin, binErr := getLocalBinDir()
	if binErr == nil {
		ensureLocalBinOnPath(localBin)
	}

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

	localBin, err := getLocalBinDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(localBin, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", localBin, err)
	}
	ensureLocalBinOnPath(localBin)

	// Step 1: Install the tool
	installCmd := exec.Command(suiupPath, "install", fmt.Sprintf("%s@%s", tool, version))

	output, err := installCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", tool, err, string(output))
	}

	// Step 2: Set as default (creates symlink in ~/.local/bin/)
	// This is critical - without this, the binary won't be accessible
	defaultCmd := exec.Command(suiupPath, "default", "set", fmt.Sprintf("%s@%s", tool, version))

	defaultOutput, err := defaultCmd.CombinedOutput()
	if err != nil {
		// Don't fail if default set fails, just log it
		fmt.Fprintf(os.Stderr, "Warning: failed to set default for %s: %v\nOutput: %s\n", tool, err, string(defaultOutput))
	}

	// Step 3: Verify the binary exists
	toolPath := ""
	for _, candidate := range execCandidates(tool) {
		candidatePath := filepath.Join(localBin, candidate)
		if _, err := os.Stat(candidatePath); err == nil {
			toolPath = candidatePath
			break
		}
	}
	if toolPath == "" {
		// Retry default set
		retryCmd := exec.Command(suiupPath, "default", "set", fmt.Sprintf("%s@%s", tool, version))
		_, _ = retryCmd.CombinedOutput() // Ignore error

		for _, candidate := range execCandidates(tool) {
			candidatePath := filepath.Join(localBin, candidate)
			if _, err := os.Stat(candidatePath); err == nil {
				toolPath = candidatePath
				break
			}
		}
		if toolPath == "" {
			return fmt.Errorf("%s binary not found in %s after installation", tool, localBin)
		}
	}

	return nil
}
