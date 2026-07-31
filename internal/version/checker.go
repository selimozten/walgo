package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ganbitlabs/walgo/internal/deps"
	"github.com/ganbitlabs/walgo/internal/sui"
	"github.com/ganbitlabs/walgo/internal/ui"
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
	Hugo        *ToolVersion // Hugo version info (installed via package manager)
	HasUpdates  bool
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

// GetLatestWalrusVersion fetches the latest mainnet Walrus version.
// Walrus releases live in the walrus repo; walrus-sites only carries site-builder.
func GetLatestWalrusVersion() (string, error) {
	version, err := getLatestGitHubReleaseForNetwork("MystenLabs", "walrus", "mainnet")
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

// GetLatestSiteBuilderVersion fetches the latest mainnet site-builder version.
// Site-builder is released via the walrus-sites repository.
func GetLatestSiteBuilderVersion() (string, error) {
	version, err := getLatestGitHubReleaseForNetwork("MystenLabs", "walrus-sites", "mainnet")
	if err == nil && version != "" {
		return version, nil
	}

	// Fallback: the repo's latest release, whichever network it targets
	return getLatestGitHubRelease("MystenLabs", "walrus-sites")
}

// GetLatestHugoVersion fetches the latest Hugo version from GitHub releases
func GetLatestHugoVersion() (string, error) {
	return getLatestGitHubRelease("gohugoio", "hugo")
}

// GetHugoVersion gets the current Hugo version and checks if it's Extended
func GetHugoVersion() (currentVersion string, isExtended bool, err error) {
	// Use deps package to check Hugo Extended
	isInstalled, extended, version, checkErr := deps.CheckHugoExtended()
	if checkErr != nil {
		return "", false, checkErr
	}
	if !isInstalled {
		return "", false, fmt.Errorf("hugo not installed")
	}

	// Parse version from output
	parsedVersion := parseVersion(version)
	if parsedVersion == "" {
		parsedVersion = "unknown"
	}

	return parsedVersion, extended, nil
}

// getLatestGitHubRelease fetches the latest release version from GitHub
func getLatestGitHubRelease(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
		neturl.PathEscape(owner), neturl.PathEscape(repo))

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

	return normalizeReleaseTag(release.TagName), nil
}

// getLatestGitHubReleaseForNetwork returns the newest release whose tag starts
// with "<network>-v". Walrus and site-builder publish per-network tags
// (e.g. mainnet-v1.52.1), and the repo's "latest" release may be for another
// network, so the plain latest endpoint is not enough.
func getLatestGitHubReleaseForNetwork(owner, repo, network string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=30",
		neturl.PathEscape(owner), neturl.PathEscape(repo))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
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

	var releases []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	prefix := strings.ToLower(network) + "-v"
	best := ""
	for _, release := range releases {
		if !strings.HasPrefix(strings.ToLower(release.TagName), prefix) {
			continue
		}
		version := normalizeReleaseTag(release.TagName)
		if best == "" || CompareVersions(version, best) > 0 {
			best = version
		}
	}

	if best == "" {
		return "", fmt.Errorf("no %s release found for %s/%s", network, owner, repo)
	}

	return best, nil
}

// normalizeReleaseTag turns a release tag into a bare semantic version,
// dropping any network prefix ("mainnet-v1.52.1" becomes "1.52.1").
func normalizeReleaseTag(tag string) string {
	version := strings.TrimSpace(tag)
	for _, prefix := range []string{"mainnet-", "testnet-", "devnet-"} {
		version = strings.TrimPrefix(version, prefix)
	}
	return strings.TrimPrefix(version, "v")
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

	// Check Hugo (managed via package manager, so we don't auto-update)
	if currentHugo, isExtended, err := GetHugoVersion(); err == nil {
		latestHugo, err := GetLatestHugoVersion()
		if err != nil {
			latestHugo = "unknown"
		}

		// Note: We check for updates but don't set HasUpdates for Hugo
		// since it's managed via package manager (brew/apt/choco), not suiup
		updateRequired := false
		if latestHugo != "unknown" && CompareVersions(latestHugo, currentHugo) > 0 {
			updateRequired = true
			// Don't set result.HasUpdates = true for Hugo (user manages via package manager)
		}

		// Add note to version if not Extended
		versionNote := currentHugo + " extended"
		if !isExtended {
			versionNote = currentHugo + " (standard - Extended recommended)"
		}

		result.Hugo = &ToolVersion{
			Tool:           "hugo",
			CurrentVersion: versionNote,
			LatestVersion:  latestHugo,
			UpdateRequired: updateRequired,
		}
	}

	return result, nil
}

// Minimum tool versions required since Sui Foundation disabled JSON-RPC on its
// public fullnodes (2026-07-31). Older builds still speak JSON-RPC and fail with
// "Method not found ... JSON-RPC on public fullnodes has been deprecated".
//
//	sui 1.75.0          - first release whose CLI reads balances over gRPC (fix #26999)
//	walrus 1.52.0       - first release that can run entirely on gRPC (fix #3526)
//	site-builder 2.12.0 - first release off JSON-RPC-only client methods (fix #725)
var (
	minSuiVersion         = [3]int{1, 75, 0}
	minWalrusVersion      = [3]int{1, 52, 0}
	minSiteBuilderVersion = [3]int{2, 12, 0}
)

// ToolStatus describes one installed tool: where it resolved from, which version
// answered, and whether that version is new enough.
type ToolStatus struct {
	Tool      string
	Path      string
	Version   string
	Minimum   string
	Installed bool
	Outdated  bool
}

// InspectToolchain resolves sui, walrus and site-builder the same way walgo does
// when it shells out, so a stale copy earlier in PATH shows up as itself rather
// than as whatever the installer thinks is current.
func InspectToolchain() []ToolStatus {
	tools := []struct {
		name string
		min  [3]int
	}{
		{"sui", minSuiVersion},
		{"walrus", minWalrusVersion},
		{"site-builder", minSiteBuilderVersion},
	}

	statuses := make([]ToolStatus, 0, len(tools))
	for _, tool := range tools {
		status := ToolStatus{Tool: tool.name, Minimum: formatVersionParts(tool.min)}

		path, err := deps.LookPath(tool.name)
		if err != nil {
			statuses = append(statuses, status)
			continue
		}
		status.Installed = true
		status.Path = path

		version, err := GetCurrentVersion(tool.name)
		if err != nil {
			statuses = append(statuses, status)
			continue
		}
		status.Version = version
		status.Outdated = versionLess(parseVersionParts(version), tool.min)

		statuses = append(statuses, status)
	}

	return statuses
}

// ToolchainProblems summarizes outdated tools from InspectToolchain output.
// Missing tools are not reported here; callers handle those separately since
// site-builder is only needed for on-chain deployments.
func ToolchainProblems(statuses []ToolStatus) error {
	var outdated []string
	for _, status := range statuses {
		if status.Outdated {
			outdated = append(outdated, fmt.Sprintf("%s v%s (need v%s or newer)",
				status.Tool, status.Version, status.Minimum))
		}
	}

	if len(outdated) == 0 {
		return nil
	}

	return fmt.Errorf(
		"outdated tools detected: %s\n\n"+
			"  Sui disabled JSON-RPC on public fullnodes on 2026-07-31. Older builds still\n"+
			"  use it and fail with \"Method not found\" on every deployment.\n\n"+
			"  To fix:\n"+
			"    suiup install sui@mainnet\n"+
			"    suiup install walrus@mainnet\n"+
			"    suiup install site-builder@mainnet\n\n"+
			"  If a tool still reports an old version afterwards, an older copy earlier in\n"+
			"  PATH is winning; check the paths printed above.",
		strings.Join(outdated, ", "),
	)
}

// versionLess reports whether version a is older than version b.
func versionLess(a, b [3]int) bool {
	for i := range 3 {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// formatVersionParts renders a parsed version back as "major.minor.patch".
func formatVersionParts(v [3]int) string {
	return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])
}

// CheckCompatibility verifies that installed walrus and site-builder versions can
// still talk to Sui. Builds older than the minimums below rely on the retired
// JSON-RPC API and fail on every deployment.
func CheckCompatibility(walrusVersion, siteBuilderVersion string) error {
	if walrusVersion == "" || siteBuilderVersion == "" {
		return nil // can't check if versions are unknown
	}

	walrusParts := parseVersionParts(walrusVersion)
	sbParts := parseVersionParts(siteBuilderVersion)

	walrusOutdated := versionLess(walrusParts, minWalrusVersion)
	sbOutdated := versionLess(sbParts, minSiteBuilderVersion)

	if !walrusOutdated && !sbOutdated {
		return nil
	}

	var outdated []string
	if walrusOutdated {
		outdated = append(outdated, fmt.Sprintf("walrus v%s (need v%s or newer)",
			walrusVersion, formatVersionParts(minWalrusVersion)))
	}
	if sbOutdated {
		outdated = append(outdated, fmt.Sprintf("site-builder v%s (need v%s or newer)",
			siteBuilderVersion, formatVersionParts(minSiteBuilderVersion)))
	}

	return fmt.Errorf(
		"outdated tools detected: %s\n\n"+
			"  Sui disabled JSON-RPC on public fullnodes on 2026-07-31. Older builds still\n"+
			"  use it and fail with \"Method not found\" on every deployment.\n\n"+
			"  To fix:\n"+
			"    suiup install walrus@mainnet\n"+
			"    suiup install site-builder@mainnet\n\n"+
			"  See: https://docs.sui.io/develop/accessing-data/json-rpc-migration",
		strings.Join(outdated, " and "),
	)
}

// CheckInstalledCompatibility is a convenience function that reads installed versions
// and checks their compatibility. Returns nil if tools are missing (not an error here).
func CheckInstalledCompatibility() error {
	walrusVer, err := GetCurrentVersion("walrus")
	if err != nil {
		return nil // walrus not installed, nothing to check
	}
	sbVer, err := GetCurrentVersion("site-builder")
	if err != nil {
		return nil // site-builder not installed, nothing to check
	}
	return CheckCompatibility(walrusVer, sbVer)
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
		if _, err := fmt.Scanln(&response); err != nil {
			// Default to "yes" if there's an error reading input
			response = ""
		}
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
