package walrus

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"walgo/internal/config" // Used for WalrusConfig type
)

// Walrus Interaction Strategy
//
// Primary Method: site-builder CLI
// Walgo interacts with Walrus Sites by executing the official `site-builder` command-line tool.
// This approach leverages the existing official tool dedicated to Walrus Sites operations.
//
// Expected `site-builder` CLI Command Structures (based on official documentation):
// 1. Publishing a new site:
//    `site-builder publish <directory> --epochs <number>`
//
// 2. Updating an existing site:
//    `site-builder update --epochs <number> <directory> <object-id>`
//
// 3. Other commands available:
//    - `site-builder convert <object-id>` - Convert hex to Base36 format
//    - `site-builder sitemap <object-id>` - Show site resources
//    - `site-builder list-directory <directory>` - Generate index.html preview
//    - `site-builder destroy <object-id>` - Destroy site
//
// Authentication for `site-builder`:
// The `site-builder` CLI handles its own authentication through:
//   - Configuration file: `sites-config.yaml` (contains wallet paths, RPC URLs, package info)
//   - Default location: `~/.config/walrus/` or specified via --config flag
//   - Sui wallet authentication
//
// SuiNS domain management is handled separately through the SuiNS interface,
// not directly through site-builder commands.

const siteBuilderCmd = "site-builder"

// Function variables for dependency injection in tests
var (
	execLookPath = exec.LookPath
	execCommand  = exec.Command
)

// SiteBuilderOutput represents parsed output from site-builder commands
type SiteBuilderOutput struct {
	ObjectID   string
	SiteURL    string
	BrowseURLs []string
	Resources  []Resource
	Base36ID   string
	Success    bool
}

// Resource represents a site resource from sitemap output
type Resource struct {
	Path   string
	BlobID string
}

// CheckSiteBuilderSetup verifies that site-builder is properly configured
func CheckSiteBuilderSetup() error {
	// Check if site-builder is installed
	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH.\n\nInstallation instructions:\n1. Download from: https://storage.googleapis.com/mysten-walrus-binaries/\n2. Choose: site-builder-testnet-latest-<your-system>\n3. Make executable: chmod +x site-builder\n4. Move to PATH: sudo mv site-builder /usr/local/bin/", siteBuilderCmd)
	}

	fmt.Printf("✓ site-builder found at: %s\n", builderPath)

	// Check if sites-config.yaml exists
	configPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".config", "walrus", "sites-config.yaml"),
		filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "walrus", "sites-config.yaml"),
		"sites-config.yaml",
	}

	var configFound bool
	var configPath string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFound = true
			configPath = path
			break
		}
	}

	if !configFound {
		return fmt.Errorf("site-builder configuration not found. Please run 'walgo setup' to configure site-builder")
	}

	fmt.Printf("✓ site-builder config found at: %s\n", configPath)
	return nil
}

// SetupSiteBuilder helps users set up the site-builder configuration
func SetupSiteBuilder(network string, force bool) error {
	if network == "" {
		network = "testnet" // Default to testnet for development
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "walrus")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "sites-config.yaml")

	// If config exists and not forcing, stop. If forcing, proceed to overwrite
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("site-builder config already exists at %s. Use --force to overwrite", configPath)
	}

	// Create default configuration based on network
	var packageID string
	var rpcURL string
	switch network {
	case "mainnet":
		packageID = "0x26eb7ee8688da02c5f671679524e379f0b837a12f1d1d799f255b7eea260ad27"
		rpcURL = "https://fullnode.mainnet.sui.io:443"
	case "testnet":
		packageID = "0xf99aee9f21493e1590e7e5a9aea6f343a1f381031a04a732724871fc294be799"
		rpcURL = "https://fullnode.testnet.sui.io:443"
	case "devnet":
		// Devnet package may change frequently; using testnet package as a placeholder is not ideal,
		// but allows configuration to proceed for HTTP-only workflows.
		packageID = "0xf99aee9f21493e1590e7e5a9aea6f343a1f381031a04a732724871fc294be799"
		rpcURL = "https://fullnode.devnet.sui.io:443"
	default:
		return fmt.Errorf("unsupported network: %s. Use 'mainnet', 'testnet' or 'devnet'", network)
	}

	walletPath := filepath.Join(homeDir, ".sui", "sui_config", "client.yaml")
	walrusConfig := filepath.Join(homeDir, ".config", "walrus", "client_config.yaml")
	walrusBinary := "/usr/local/bin/walrus"

	configContent := fmt.Sprintf(`contexts:
  %s:
    package: %s
    general:
      rpc_url: %s
      wallet: %s
      walrus_binary: %s
      walrus_config: %s
      gas_budget: 500000000

default_context: %s
`, network, packageID, rpcURL, walletPath, walrusBinary, walrusConfig, network)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write site-builder config: %w", err)
	}

	fmt.Printf("✓ Created site-builder configuration at: %s\n", configPath)
	fmt.Printf("✓ Network: %s\n", network)
	fmt.Printf("✓ Package ID: %s\n", packageID)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Ensure you have a Sui wallet configured: sui client addresses")
	fmt.Println("2. Test the setup: site-builder --help")
	fmt.Println("3. You're ready to deploy: walgo deploy")

	return nil
}

// DeploySite handles the deployment of the site to Walrus.
// It executes the `site-builder publish` command for new sites.
func DeploySite(deployDir string, walrusCfg config.WalrusConfig, epochs int) (*SiteBuilderOutput, error) {
	// Check setup first
	if err := CheckSiteBuilderSetup(); err != nil {
		return nil, fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	// For new deployments, we don't need the ProjectID to be set yet
	// (it will be set after successful deployment)

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	// For new sites, use 'site-builder publish'
	// site-builder publish <directory> --epochs <number>
	args := []string{
		"publish",
		deployDir,
		"--epochs", fmt.Sprintf("%d", epochs),
	}

	fmt.Printf("Executing: %s %s\n", builderPath, strings.Join(args, " "))

	cmd := execCommand(builderPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errorMsg := fmt.Sprintf("failed to execute %s: %v", siteBuilderCmd, err)
		if stderr.Len() > 0 {
			errorMsg += fmt.Sprintf("\nstderr:\n%s", stderr.String())
		}
		if stdout.Len() > 0 {
			errorMsg += fmt.Sprintf("\nstdout:\n%s", stdout.String())
		}
		return nil, fmt.Errorf("%s", errorMsg)
	}

	fmt.Println("Site deployment command executed successfully.")

	// Parse the output
	output := parseSiteBuilderOutput(stdout.String())
	output.Success = true

	if stdout.Len() > 0 {
		fmt.Printf("Output from %s:\n%s\n", siteBuilderCmd, stdout.String())
	}

	if stderr.Len() > 0 {
		fmt.Printf("Stderr from %s (may contain warnings or informational messages):\n%s\n", siteBuilderCmd, stderr.String())
	}

	// Provide helpful guidance
	if output.ObjectID != "" {
		fmt.Println("\n🎉 Deployment successful!")
		fmt.Printf("📋 Site Object ID: %s\n", output.ObjectID)
		fmt.Println("\n📝 Next steps:")
		fmt.Printf("1. Save this Object ID in your walgo.yaml:\n")
		fmt.Printf("   walrus:\n")
		fmt.Printf("     projectID: \"%s\"\n", output.ObjectID)
		fmt.Printf("2. Configure a SuiNS domain: walgo domain <your-domain>\n")
		fmt.Printf("3. Update your site: walgo update\n")
		fmt.Printf("4. Check status: walgo status\n")

		if len(output.BrowseURLs) > 0 {
			fmt.Printf("\n🌐 Browse your site:\n")
			for _, url := range output.BrowseURLs {
				fmt.Printf("   %s\n", url)
			}
		}
	}

	return output, nil
}

// UpdateSite handles updating an existing site on Walrus.
// It executes the `site-builder update` command.
func UpdateSite(deployDir, objectID string, epochs int) (*SiteBuilderOutput, error) {
	if objectID == "" {
		return nil, fmt.Errorf("object ID is required for updating a site")
	}

	// Check setup first
	if err := CheckSiteBuilderSetup(); err != nil {
		return nil, fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	// site-builder update --epochs <number> <directory> <object-id>
	args := []string{
		"update",
		"--epochs", fmt.Sprintf("%d", epochs),
		deployDir,
		objectID,
	}

	fmt.Printf("Executing: %s %s\n", builderPath, strings.Join(args, " "))

	cmd := execCommand(builderPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errorMsg := fmt.Sprintf("failed to execute %s: %v", siteBuilderCmd, err)
		if stderr.Len() > 0 {
			errorMsg += fmt.Sprintf("\nstderr:\n%s", stderr.String())
		}
		if stdout.Len() > 0 {
			errorMsg += fmt.Sprintf("\nstdout:\n%s", stdout.String())
		}
		return nil, fmt.Errorf("%s", errorMsg)
	}

	fmt.Println("Site update command executed successfully.")

	// Parse the output
	output := parseSiteBuilderOutput(stdout.String())
	output.Success = true
	output.ObjectID = objectID // For updates, we know the object ID

	if stdout.Len() > 0 {
		fmt.Printf("Output from %s:\n%s\n", siteBuilderCmd, stdout.String())
	}

	if stderr.Len() > 0 {
		fmt.Printf("Stderr from %s:\n%s\n", siteBuilderCmd, stderr.String())
	}

	fmt.Println("\n✅ Site updated successfully!")
	fmt.Printf("📋 Object ID: %s\n", objectID)
	fmt.Println("🌐 Your site should be updated at the same URLs as before")

	return output, nil
}

// GetSiteStatus checks the status of a Walrus site.
// Note: The site-builder doesn't have a direct "status" command, but we can use sitemap.
func GetSiteStatus(objectID string) (*SiteBuilderOutput, error) {
	if objectID == "" {
		return nil, fmt.Errorf("object ID is required for checking site status")
	}

	// Check setup first
	if err := CheckSiteBuilderSetup(); err != nil {
		return nil, fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	// Use sitemap to show site resources as a form of status
	args := []string{
		"sitemap",
		objectID,
	}

	fmt.Printf("Executing: %s %s\n", builderPath, strings.Join(args, " "))

	cmd := execCommand(builderPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errorMsg := fmt.Sprintf("failed to execute %s: %v", siteBuilderCmd, err)
		if stderr.Len() > 0 {
			errorMsg += fmt.Sprintf("\nstderr:\n%s", stderr.String())
		}
		if stdout.Len() > 0 {
			errorMsg += fmt.Sprintf("\nstdout:\n%s", stdout.String())
		}
		return nil, fmt.Errorf("%s", errorMsg)
	}

	fmt.Println("Site status retrieved successfully.")

	// Parse the sitemap output
	output := parseSitemapOutput(stdout.String())
	output.Success = true
	output.ObjectID = objectID

	if stdout.Len() > 0 {
		fmt.Printf("Site resources:\n%s\n", stdout.String())
	}

	if stderr.Len() > 0 {
		fmt.Printf("Stderr from %s:\n%s\n", siteBuilderCmd, stderr.String())
	}

	return output, nil
}

// ConvertObjectID converts a hex object ID to Base36 format
func ConvertObjectID(objectID string) (string, error) {
	if objectID == "" {
		return "", fmt.Errorf("object ID is required for conversion")
	}

	// Check setup first
	if err := CheckSiteBuilderSetup(); err != nil {
		return "", fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return "", fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	args := []string{
		"convert",
		objectID,
	}

	fmt.Printf("Executing: %s %s\n", builderPath, strings.Join(args, " "))

	cmd := execCommand(builderPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errorMsg := fmt.Sprintf("failed to execute %s: %v", siteBuilderCmd, err)
		if stderr.Len() > 0 {
			errorMsg += fmt.Sprintf("\nstderr:\n%s", stderr.String())
		}
		if stdout.Len() > 0 {
			errorMsg += fmt.Sprintf("\nstdout:\n%s", stdout.String())
		}
		return "", fmt.Errorf("%s", errorMsg)
	}

	var base36ID string
	if stdout.Len() > 0 {
		// Extract the last non-empty token that looks like base36 (lowercase letters and digits)
		out := stdout.String()
		lines := strings.Split(out, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(lines[i])
			if candidate == "" {
				continue
			}
			// If line contains spaces, take the last whitespace-separated token
			if strings.Contains(candidate, " ") {
				parts := strings.Fields(candidate)
				candidate = parts[len(parts)-1]
			}
			lower := strings.ToLower(candidate)
			if lower == candidate && regexp.MustCompile(`^[0-9a-z]+$`).MatchString(candidate) {
				base36ID = candidate
				break
			}
		}

		if base36ID != "" {
			fmt.Printf("Base36 representation: %s\n", base36ID)
			fmt.Printf("\n🌐 Direct access URLs:\n")
			fmt.Printf("   https://%s.wal.app\n", base36ID)
			fmt.Printf("   http://%s.localhost:3000 (local portal)\n", base36ID)
		} else {
			// Fallback: show raw output for debugging
			fmt.Printf("Warning: could not parse Base36 ID from output. Raw output follows:\n%s\n", out)
		}
	}

	if stderr.Len() > 0 {
		fmt.Printf("Stderr from %s:\n%s\n", siteBuilderCmd, stderr.String())
	}

	return base36ID, nil
}

// parseSiteBuilderOutput extracts key information from site-builder command output
func parseSiteBuilderOutput(output string) *SiteBuilderOutput {
	result := &SiteBuilderOutput{
		BrowseURLs: make([]string, 0),
		Resources:  make([]Resource, 0),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract object ID
		if strings.Contains(line, "New site object ID:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				result.ObjectID = strings.TrimSpace(parts[1])
			}
		}

		// Extract site object ID (for updates)
		if strings.Contains(line, "Site object ID:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				result.ObjectID = strings.TrimSpace(parts[1])
			}
		}

		// Extract browse URLs
		if strings.Contains(line, "http://") || strings.Contains(line, "https://") {
			// Extract URLs from the line
			urlRegex := regexp.MustCompile(`https?://[^\s]+`)
			urls := urlRegex.FindAllString(line, -1)
			result.BrowseURLs = append(result.BrowseURLs, urls...)
		}
	}

	return result
}

// parseSitemapOutput extracts resources from sitemap command output
func parseSitemapOutput(output string) *SiteBuilderOutput {
	result := &SiteBuilderOutput{
		Resources: make([]Resource, 0),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse resource lines (format may vary, this is a basic parser)
		if strings.Contains(line, "blob ID") {
			// Example: "- created resource /index.html with blob ID ABC123..."
			parts := strings.Fields(line)
			var path, blobID string

			for i, part := range parts {
				if part == "resource" && i+1 < len(parts) {
					path = parts[i+1]
				}
				if part == "ID" && i+1 < len(parts) {
					blobID = parts[i+1]
				}
			}

			if path != "" && blobID != "" {
				result.Resources = append(result.Resources, Resource{
					Path:   path,
					BlobID: blobID,
				})
			}
		}
	}

	return result
}
