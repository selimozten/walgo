package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/sui"
	"github.com/selimozten/walgo/internal/ui"
)

// ToolVersion represents version information for a tool
type ToolVersion struct {
	Tool           string
	CurrentVersion string
	LatestVersion  string
	UpdateRequired bool
}

// CheckResult contains the results of version checking
type CheckResult struct {
	Sui         *ToolVersion
	Walrus      *ToolVersion
	SiteBuilder *ToolVersion
	// Hugo removed - users manage via package manager (brew/apt/choco)
	HasUpdates bool
}

// GetCurrentVersion retrieves the currently installed version of a tool
func GetCurrentVersion(tool string) (string, error) {
	// Validate tool name (suiup tools only)
	switch tool {
	case "sui", "walrus", "site-builder":
		// Valid tool
	default:
		return "", fmt.Errorf("unknown tool: %s", tool)
	}

	// Use deps package to get version
	output, err := deps.GetToolVersion(tool)
	if err != nil {
		return "", fmt.Errorf("failed to get %s version: %w", tool, err)
	}

	version := parseVersion(output)
	if version == "" {
		return "", fmt.Errorf("could not parse version from output: %s", output)
	}

	return version, nil
}

// parseVersion extracts version number from command output
func parseVersion(output string) string {
	// Try to match semantic version (e.g., 1.2.3, v1.2.3)
	re := regexp.MustCompile(`v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try to match simpler version (e.g., 1.2)
	re = regexp.MustCompile(`v?(\d+\.\d+)`)
	matches = re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// GetLatestSuiVersion fetches the latest Sui version from GitHub releases
func GetLatestSuiVersion() (string, error) {
	return getLatestGitHubRelease("MystenLabs", "sui")
}

// GetLatestWalrusVersion fetches the latest Walrus version
// Walrus binaries are distributed via the walrus-sites repo releases
func GetLatestWalrusVersion() (string, error) {
	// Try walrus-sites repo first (this is where site-builder releases are)
	version, err := getLatestGitHubRelease("MystenLabs", "walrus-sites")
	if err == nil && version != "" {
		return version, nil
	}

	// Fallback: try to get version from walrus CLI directly
	// This happens when GitHub API is unavailable or rate-limited
	currentVersion, err := GetCurrentVersion("walrus")
	if err == nil {
		// Return current version with a note that we couldn't verify latest
		return currentVersion, nil
	}

	return "", fmt.Errorf("unable to determine latest Walrus version: GitHub API unavailable and local walrus not found")
}

// GetLatestSiteBuilderVersion fetches the latest site-builder version
// Site-builder is released via the walrus-sites repository
func GetLatestSiteBuilderVersion() (string, error) {
	// Site-builder releases are in walrus-sites repo
	return getLatestGitHubRelease("MystenLabs", "walrus-sites")
}

// Hugo version checking removed - users manage Hugo via package manager

// getLatestGitHubRelease fetches the latest release version from GitHub
func getLatestGitHubRelease(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to avoid GitHub API rate limiting
	req.Header.Set("User-Agent", "walgo-version-checker")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Remove 'v' prefix if present
	version := strings.TrimPrefix(release.TagName, "v")
	return version, nil
}

// CompareVersions compares two semantic versions
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func CompareVersions(v1, v2 string) int {
	v1Parts := parseVersionParts(v1)
	v2Parts := parseVersionParts(v2)

	for i := 0; i < 3; i++ {
		if v1Parts[i] > v2Parts[i] {
			return 1
		}
		if v1Parts[i] < v2Parts[i] {
			return -1
		}
	}

	return 0
}

// parseVersionParts splits version into major, minor, patch
func parseVersionParts(version string) [3]int {
	parts := [3]int{0, 0, 0}

	// Remove any pre-release suffix (e.g., -alpha, -beta)
	version = strings.Split(version, "-")[0]

	// Split by dots
	segments := strings.Split(version, ".")

	for i, segment := range segments {
		if i >= 3 {
			break
		}
		if num, err := strconv.Atoi(strings.TrimSpace(segment)); err == nil {
			parts[i] = num
		}
	}

	return parts
}

// CheckAllVersions checks versions of all required tools
func CheckAllVersions() (*CheckResult, error) {
	result := &CheckResult{}

	// Check Sui
	if currentSui, err := GetCurrentVersion("sui"); err == nil {
		latestSui, err := GetLatestSuiVersion()
		if err != nil {
			// Don't fail if we can't fetch latest, just log
			latestSui = "unknown"
		}

		updateRequired := false
		if latestSui != "unknown" && CompareVersions(latestSui, currentSui) > 0 {
			updateRequired = true
			result.HasUpdates = true
		}

		result.Sui = &ToolVersion{
			Tool:           "sui",
			CurrentVersion: currentSui,
			LatestVersion:  latestSui,
			UpdateRequired: updateRequired,
		}
	}

	// Check Walrus
	if currentWalrus, err := GetCurrentVersion("walrus"); err == nil {
		latestWalrus, err := GetLatestWalrusVersion()
		if err != nil {
			latestWalrus = "unknown"
		}

		updateRequired := false
		if latestWalrus != "unknown" && CompareVersions(latestWalrus, currentWalrus) > 0 {
			updateRequired = true
			result.HasUpdates = true
		}

		result.Walrus = &ToolVersion{
			Tool:           "walrus",
			CurrentVersion: currentWalrus,
			LatestVersion:  latestWalrus,
			UpdateRequired: updateRequired,
		}
	}

	// Check site-builder
	if currentSB, err := GetCurrentVersion("site-builder"); err == nil {
		latestSB, err := GetLatestSiteBuilderVersion()
		if err != nil {
			latestSB = "unknown"
		}

		updateRequired := false
		if latestSB != "unknown" && CompareVersions(latestSB, currentSB) > 0 {
			updateRequired = true
			result.HasUpdates = true
		}

		result.SiteBuilder = &ToolVersion{
			Tool:           "site-builder",
			CurrentVersion: currentSB,
			LatestVersion:  latestSB,
			UpdateRequired: updateRequired,
		}
	}

	// Hugo version check removed - users manage Hugo via package manager

	return result, nil
}

// UpdateTool updates a tool to the latest version using suiup or direct download
func UpdateTool(tool string, network string) error {
	// Validate tool name
	switch tool {
	case "sui", "walrus", "site-builder":
		// Use suiup to install/update tool
		return deps.InstallTool(tool, network)
	default:
		return fmt.Errorf("unknown tool: %s", tool)
	}
}

// UpdateAllTools updates all tools to their latest versions
func UpdateAllTools(network string) error {
	icons := ui.GetIcons()
	tools := []string{"sui", "walrus", "site-builder"}

	for _, tool := range tools {
		fmt.Printf("Updating %s...\n", tool)
		if err := UpdateTool(tool, network); err != nil {
			return fmt.Errorf("failed to update %s: %w", tool, err)
		}
		fmt.Printf("%s %s updated successfully\n", icons.Check, tool)
	}

	return nil
}

// CheckAndUpdateVersions checks if tools need updates and prompts user to update for mainnet
func CheckAndUpdateVersions(quiet bool) error {
	icons := ui.GetIcons()
	network, err := sui.GetActiveEnv()
	if err != nil {
		return fmt.Errorf("failed to get active network: %w", err)
	}

	// Only enforce version checking for mainnet deployments
	if !strings.Contains(strings.ToLower(network), "mainnet") {
		return nil
	}

	if !quiet {
		fmt.Printf("  %s Checking tool versions for mainnet deployment...\n", icons.Info)
	}

	// Check versions
	result, err := CheckAllVersions()
	if err != nil {
		return fmt.Errorf("failed to check versions: %w", err)
	}

	// If no updates needed, continue
	if !result.HasUpdates {
		if !quiet {
			fmt.Printf("  %s All tools are up to date\n", icons.Check)
		}
		return nil
	}

	// Display update information
	if !quiet {
		fmt.Println()
		fmt.Printf("%s Updates available for mainnet deployment:\n", icons.Warning)
		fmt.Println()

		if result.Sui != nil && result.Sui.UpdateRequired {
			fmt.Printf("  • Sui: %s → %s\n", result.Sui.CurrentVersion, result.Sui.LatestVersion)
		}
		if result.Walrus != nil && result.Walrus.UpdateRequired {
			fmt.Printf("  • Walrus: %s → %s\n", result.Walrus.CurrentVersion, result.Walrus.LatestVersion)
		}
		if result.SiteBuilder != nil && result.SiteBuilder.UpdateRequired {
			fmt.Printf("  • Site-builder: %s → %s\n", result.SiteBuilder.CurrentVersion, result.SiteBuilder.LatestVersion)
		}

		fmt.Println()
		fmt.Printf("%s For mainnet deployments, it's recommended to use the latest versions.\n", icons.Lightbulb)
		fmt.Print("\nWould you like to update now? [Y/n]: ")

		// Read user response
		var response string
		fmt.Scanln(&response)
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "" || response == "y" || response == "yes" {
			fmt.Println()
			fmt.Println("Updating tools...")
			fmt.Println()

			if err := UpdateAllTools(network); err != nil {
				return fmt.Errorf("failed to update tools: %w", err)
			}

			fmt.Println()
			fmt.Printf("%s All tools updated successfully!\n", icons.Check)
			fmt.Println()
		} else {
			fmt.Println()
			fmt.Printf("%s Continuing with current versions...\n", icons.Warning)
			fmt.Println("   (Use --skip-version-check to skip this prompt)")
			fmt.Println()
		}
	}

	return nil
}
