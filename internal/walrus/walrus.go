package walrus

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/sui"
	"github.com/selimozten/walgo/internal/ui"
)

// Package walrus provides integration with Walrus decentralized storage.
// It wraps the official site-builder CLI for publishing, updating, and managing sites.
// Authentication is handled via sites-config.yaml in ~/.config/walrus/

const siteBuilderCmd = "site-builder"

// TokenInfo represents information about a token type
type TokenInfo struct {
	Decimals    int    `json:"decimals"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	ID          string `json:"id"`
}

// CoinBalance represents a coin balance
type CoinBalance struct {
	CoinType            string `json:"coinType"`
	CoinObjectId        string `json:"coinObjectId"`
	Version             string `json:"version"`
	Digest              string `json:"digest"`
	Balance             string `json:"balance"`
	PreviousTransaction string `json:"previousTransaction"`
}

// BalanceEntry is an entry in the balance array containing token info and coins
type BalanceEntry struct {
	TokenInfo TokenInfo     `json:"-"`
	Coins     []CoinBalance `json:"-"`
}

// WalletBalance represents the full wallet balance response
type WalletBalance struct {
	Entries []BalanceEntry
	HasMore bool `json:"hasMore"`
}

// DefaultCommandTimeout is the maximum time allowed for site-builder operations
// Deployments can take a while for large sites, so we use a generous timeout
const DefaultCommandTimeout = 10 * time.Minute

// Test hooks for dependency injection
var (
	execLookPath = deps.LookPath
	execCommand  = exec.Command
	verboseMode  = false
	runPreflight = true
	osStat       = os.Stat
)

// execCommandContext is a test hook for creating context-aware commands
var execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// runCommandWithTimeout executes a command with a timeout context
// Returns stdout, stderr, and any error
func runCommandWithTimeout(ctx context.Context, name string, args []string, streamOutput bool) (string, string, error) {
	// Create timeout context if none provided
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), DefaultCommandTimeout)
		defer cancel()
	}

	cmd := execCommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer

	if streamOutput {
		// Stream output in real-time while also capturing it
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()

	// Check if the context deadline was exceeded
	if ctx.Err() == context.DeadlineExceeded {
		return stdout.String(), stderr.String(), fmt.Errorf("command timed out after %v - the operation took too long", DefaultCommandTimeout)
	}

	if ctx.Err() == context.Canceled {
		return stdout.String(), stderr.String(), fmt.Errorf("command was cancelled")
	}

	return stdout.String(), stderr.String(), err
}

func SetVerbose(verbose bool) {
	verboseMode = verbose
}

// SiteBuilderOutput contains the result of site-builder operations
type SiteBuilderOutput struct {
	ObjectID   string
	SiteURL    string
	BrowseURLs []string
	Resources  []Resource
	Base36ID   string
	Success    bool
}

type Resource struct {
	Path   string
	BlobID string
}

// validateObjectID ensures objectID is hexadecimal to prevent command injection
func validateObjectID(objectID string) error {
	if objectID == "" {
		return fmt.Errorf("object ID cannot be empty")
	}

	validObjectID := regexp.MustCompile(`^(0x)?[0-9a-fA-F]+$`)
	if !validObjectID.MatchString(objectID) {
		return fmt.Errorf("invalid object ID format: %s (must be hexadecimal, optionally prefixed with 0x)", objectID)
	}

	if strings.ContainsAny(objectID, "\x00\n\r;|&$`(){}[]<>") {
		return fmt.Errorf("object ID contains invalid characters")
	}

	return nil
}

// CheckSiteBuilderSetup verifies site-builder installation and configuration
func CheckSiteBuilderSetup() error {
	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return fmt.Errorf("'%s' CLI not found. Please install it using suiup:\n\n"+
			"  1. Install suiup (if not installed):\n"+
			"     curl -sSfL https://raw.githubusercontent.com/MystenLabs/suiup/main/install.sh | sh\n\n"+
			"  2. Install site-builder:\n"+
			"     suiup install site-builder@mainnet\n"+
			"     suiup default set site-builder@mainnet\n\n"+
			"  Or run: walgo setup-deps", siteBuilderCmd)
	}

	icons := ui.GetIcons()
	fmt.Printf("%s site-builder found at: %s\n", icons.Check, builderPath)

	configPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".config", "walrus", "sites-config.yaml"),
		filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "walrus", "sites-config.yaml"),
		"sites-config.yaml",
	}

	var configFound bool
	var configPath string
	for _, path := range configPaths {
		if _, err := osStat(path); err == nil {
			configFound = true
			configPath = path
			break
		}
	}

	if !configFound {
		return fmt.Errorf("site-builder configuration not found. Please run 'walgo setup' to configure site-builder")
	}

	fmt.Printf("%s site-builder config found at: %s\n", icons.Check, configPath)
	return nil
}

// handleSiteBuilderError converts site-builder errors into actionable messages
func handleSiteBuilderError(err error, errorOutput string) error {
	icons := ui.GetIcons()
	if strings.Contains(errorOutput, "could not retrieve enough confirmations") {
		return fmt.Errorf("\n%s Walrus testnet is experiencing network issues\n\n"+
			"The storage nodes couldn't provide enough confirmations.\n"+
			"This is typically temporary. You can:\n\n"+
			"  1. Wait 5-10 minutes and retry: walgo launch\n"+
			"  2. Check Walrus status: walrus info\n"+
			"  3. Try with fewer epochs (select 1 epoch in wizard)\n"+
			"  4. Join Discord for help: https://discord.gg/walrus\n\n"+
			"Technical error: %v", icons.Error, err)
	}

	if strings.Contains(errorOutput, "insufficient funds") || strings.Contains(errorOutput, "InsufficientGas") {
		return fmt.Errorf("\n%s Insufficient SUI balance\n\n"+
			"Your wallet doesn't have enough SUI for this transaction.\n\n"+
			"  Check balance: sui client balance\n"+
			"  Get testnet SUI: https://discord.com/channels/916379725201563759/971488439931392130\n\n"+
			"Technical error: %v", icons.Error, err)
	}

	if strings.Contains(errorOutput, "data did not match any variant") {
		return fmt.Errorf("\n%s Configuration format error\n\n"+
			"The Walrus client config file has incorrect formatting.\n"+
			"Please ensure object IDs are in hex format (starting with 0x).\n\n"+
			"Config location: ~/.config/walrus/client_config.yaml\n\n"+
			"Technical error: %v", icons.Error, err)
	}

	if strings.Contains(errorOutput, "wallet not found") || strings.Contains(errorOutput, "Cannot open wallet") {
		return fmt.Errorf("\n%s Wallet configuration error\n\n"+
			"Cannot find or open the Sui wallet.\n\n"+
			"  Setup wallet: sui client\n"+
			"  Check config: cat ~/.sui/sui_config/client.yaml\n\n"+
			"Technical error: %v", icons.Error, err)
	}

	if strings.Contains(errorOutput, "Request rejected `429`") || strings.Contains(errorOutput, "rate limit") {
		return fmt.Errorf("\n%s Rate limit error\n\n"+
			"The Sui RPC node is rate limiting requests.\n"+
			"This is common on public RPC endpoints.\n\n"+
			"Solutions:\n"+
			"  1. Wait 30-60 seconds and retry: walgo launch\n"+
			"  2. Use a private RPC endpoint (update sites-config.yaml)\n"+
			"  3. Try again during off-peak hours\n\n"+
			"Technical error: %v", icons.Error, err)
	}

	return fmt.Errorf("failed to execute site-builder: %v", err)
}

// PreflightCheck validates environment before deployment
func PreflightCheck() error {
	icons := ui.GetIcons()
	fmt.Printf("%s Running pre-flight checks...\n", icons.Search)

	walrusPath, err := execLookPath("walrus")
	if err != nil {
		return fmt.Errorf("walrus CLI not found in PATH")
	}
	fmt.Printf("   %s Walrus CLI found at: %s\n", icons.Check, walrusPath)

	// Check Walrus network connectivity (using --json and --context for correct network)
	walrusContext := GetWalrusContext()
	infoCmd := execCommand("walrus", "info", "--json", "--context", walrusContext)
	infoCmd.Stdout = nil
	infoCmd.Stderr = nil
	if err := infoCmd.Run(); err != nil {
		return fmt.Errorf("cannot connect to Walrus network (%s)", walrusContext)
	}
	fmt.Printf("   %s Walrus network is reachable (%s)\n", icons.Check, walrusContext)

	// Check if sui CLI is available
	suiPath, err := execLookPath("sui")
	if err != nil {
		return fmt.Errorf("sui CLI not found. Install via suiup:\n   curl -sSfL https://raw.githubusercontent.com/MystenLabs/suiup/main/install.sh | sh\n   suiup install sui@testnet\n\n   Or run: walgo setup-deps")
	}
	fmt.Printf("   %s Sui CLI found at: %s\n", icons.Check, suiPath)

	// Check wallet balance using sui package
	balance, err := sui.GetBalance()
	if err != nil {
		fmt.Printf("   %s Could not check wallet balance (continuing anyway)\n", icons.Warning)
	} else {
		// Display SUI balance
		if balance.SUI > 0 {
			fmt.Printf("   %s SUI balance: %.2f SUI\n", icons.Check, balance.SUI)
		} else {
			fmt.Printf("   %s Warning: No SUI balance detected. You may need testnet SUI from:\n", icons.Warning)
			fmt.Println("      https://discord.com/channels/916379725201563759/971488439931392130")
		}

		// Display WAL balance (if available)
		if balance.WAL > 0 {
			fmt.Printf("   %s WAL balance: %.2f WAL\n", icons.Check, balance.WAL)
		} else {
			fmt.Printf("   %s No WAL tokens found (may need for storage quota)\n", icons.Info)
		}
	}

	fmt.Printf("%s Pre-flight checks passed\n", icons.Success)
	return nil
}

// SetupSiteBuilder helps users set up the site-builder configuration.
// Uses the same approach as install.sh - downloads official configs from upstream.
func SetupSiteBuilder(network string, force bool) error {
	if network == "" {
		network = "testnet" // Default to testnet for development
	}

	// Validate network
	if network != "mainnet" && network != "testnet" {
		return fmt.Errorf("unsupported network: %s. Use 'mainnet', 'testnet'", network)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "walrus")
	// #nosec G301 - config directory needs standard permissions
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	icons := ui.GetIcons()
	fmt.Printf("%s Setting up Walrus configuration...\n", icons.Wrench)
	fmt.Printf("   Network: %s\n", network)
	fmt.Println()

	// Check if required tools are installed
	fmt.Printf("%s Checking dependencies...\n", icons.Clipboard)
	missingTools := checkWalrusDependencies()

	if len(missingTools) > 0 {
		fmt.Printf("%s Missing tools: %s\n", icons.Warning, strings.Join(missingTools, ", "))
		fmt.Println()
		fmt.Printf("%s To install missing dependencies:\n", icons.Lightbulb)
		fmt.Println("   1. Install suiup:")
		fmt.Println("      curl -sSfL https://raw.githubusercontent.com/MystenLabs/suiup/main/install.sh | sh")
		fmt.Println()
		fmt.Println("   2. Install tools via suiup:")
		fmt.Printf("      suiup install sui@%s\n", network)
		fmt.Printf("      suiup install walrus@%s\n", network)
		fmt.Println("      suiup install site-builder@mainnet")
		fmt.Println("      suiup default set site-builder@mainnet")
		fmt.Println()
		fmt.Println("   Or run: walgo setup-deps")
		fmt.Println()
	} else {
		fmt.Printf("%s All required tools found\n", icons.Check)
	}

	// Download Walrus client config
	clientConfigPath := filepath.Join(configDir, "client_config.yaml")
	if _, err := os.Stat(clientConfigPath); os.IsNotExist(err) || force {
		fmt.Printf("%s Downloading Walrus client configuration...\n", icons.Download)
		if err := downloadConfig(
			"https://docs.wal.app/setup/client_config.yaml",
			clientConfigPath,
		); err != nil {
			fmt.Printf("   %s Warning: Failed to download client config: %v\n", icons.Warning, err)
			fmt.Println("   Continuing with site-builder setup...")
		} else {
			fmt.Printf("%s Walrus client config downloaded\n", icons.Check)
		}
	} else {
		fmt.Printf("%s Walrus client config exists\n", icons.Check)
	}

	// Download site-builder config
	sitesConfigPath := filepath.Join(configDir, "sites-config.yaml")
	if _, err := os.Stat(sitesConfigPath); err == nil && !force {
		return fmt.Errorf("site-builder config already exists at %s. Use --force to overwrite", sitesConfigPath)
	}

	fmt.Printf("%s Downloading site-builder configuration...\n", icons.Download)

	// Use the official config from GitHub
	configURL := "https://raw.githubusercontent.com/MystenLabs/walrus-sites/refs/heads/mainnet/sites-config.yaml"
	if err := downloadConfig(configURL, sitesConfigPath); err != nil {
		// Fallback to creating config manually if download fails
		fmt.Printf("   %s Warning: Failed to download config: %v\n", icons.Warning, err)
		fmt.Println("   Creating config from template...")

		if err := createSiteBuilderConfigFromTemplate(sitesConfigPath, network); err != nil {
			return fmt.Errorf("failed to create site-builder config: %w", err)
		}
	} else {
		fmt.Printf("%s site-builder config downloaded\n", icons.Check)

		// Update default context if needed
		if network != "testnet" {
			if err := updateDefaultContext(sitesConfigPath, network); err != nil {
				fmt.Printf("   %s Warning: Failed to update default context: %v\n", icons.Warning, err)
			}
		}
	}

	// Also update client_config.yaml default context if it was downloaded
	if network != "testnet" {
		if _, err := os.Stat(clientConfigPath); err == nil {
			if err := updateDefaultContext(clientConfigPath, network); err != nil {
				fmt.Printf("   %s Warning: Failed to update client config context: %v\n", icons.Warning, err)
			}
		}
	}

	fmt.Println()
	fmt.Printf("%s Setup complete!\n", icons.Success)
	fmt.Printf("   Config directory: %s\n", configDir)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Ensure Sui wallet is configured:")
	fmt.Println("     sui client addresses")
	fmt.Println()
	fmt.Println("  2. Fund your wallet:")
	if network == "testnet" {
		fmt.Println("     Visit: https://faucet.sui.io/")
		fmt.Println("     Get WAL tokens: walrus get-wal --context testnet")
	} else {
		fmt.Println("     Buy SUI and send to: $(sui client active-address)")
	}
	fmt.Println()
	fmt.Println("  3. Test the setup:")
	fmt.Println("     site-builder --help")
	fmt.Println()
	fmt.Println("  4. Deploy your site:")
	fmt.Println("     walgo launch")
	fmt.Println()

	return nil
}

// checkWalrusDependencies checks if required tools are installed
func checkWalrusDependencies() []string {
	var missing []string

	tools := []string{"sui", "walrus", "site-builder"}
	for _, tool := range tools {
		if _, err := execLookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}

	return missing
}

// downloadConfig downloads a configuration file from a URL
func downloadConfig(url, destPath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download the config
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Read the response
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Write to file
	// #nosec G306 - config file needs to be readable
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// createSiteBuilderConfigFromTemplate creates a config file from hardcoded template
// This is a fallback when downloading the official config fails
func createSiteBuilderConfigFromTemplate(configPath, network string) error {
	var packageID string
	var rpcURL string

	switch network {
	case "mainnet":
		packageID = "0x26eb7ee8688da02c5f671679524e379f0b837a12f1d1d799f255b7eea260ad27"
		rpcURL = "https://fullnode.mainnet.sui.io:443"
	case "testnet":
		packageID = "0xf99aee9f21493e1590e7e5a9aea6f343a1f381031a04a732724871fc294be799"
		rpcURL = "https://fullnode.testnet.sui.io:443"
	default:
		return fmt.Errorf("unsupported network: %s", network)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
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

	// #nosec G306 - config file needs to be readable for site-builder
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// updateDefaultContext updates the default_context field in a YAML config file
func updateDefaultContext(configPath, network string) error {
	// Read the file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Replace the default context line
	// This is a simple regex replacement - for production use a proper YAML parser
	oldContext := "default_context: testnet"
	newContext := fmt.Sprintf("default_context: %s", network)

	updated := strings.ReplaceAll(string(content), oldContext, newContext)

	// Write back
	// #nosec G306 - config file needs to be readable
	if err := os.WriteFile(configPath, []byte(updated), 0644); err != nil {
		return err
	}

	return nil
}

// DeploySite handles the deployment of the site to Walrus.
// It executes the `site-builder publish` command for new sites.
// The context can be used to cancel or timeout the operation.
func DeploySite(ctx context.Context, deployDir string, walrusCfg config.WalrusConfig, epochs int) (*SiteBuilderOutput, error) {
	// Validate epochs
	if epochs <= 0 {
		return nil, fmt.Errorf("epochs must be greater than 0, got %d", epochs)
	}

	icons := ui.GetIcons()

	// Analyze deployment directory
	fmt.Printf("%s Analyzing deployment directory...\n", icons.Chart)
	fileCount := 0
	totalSize := int64(0)

	err := filepath.Walk(deployDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		fmt.Printf("   %s Warning: Could not analyze directory: %v\n", icons.Warning, err)
	} else {
		// Use robust cost estimation with detailed breakdown
		icons := ui.GetIcons()
		fmt.Printf("%s Calculating deployment costs...\n", icons.Chart)

		// Get network context from config
		network := "testnet" // Default
		if walrusCfg.Network != "" {
			network = walrusCfg.Network
		}

		// Use robust cost estimation
		options := CostOptions{
			SiteSize:  totalSize,
			Epochs:    epochs,
			Network:   network,
			FileCount: fileCount,
			RPCURL:    "", // Will use default from sites-config.yaml
		}

		breakdown, err := CalculateCost(options)
		if err != nil {
			fmt.Printf("   %s Warning: Cost estimation failed: %v\n", icons.Warning, err)
			// Fall back to simple estimate
			sizeMB := float64(totalSize) / (1024 * 1024)
			fallbackCost := sizeMB * 0.01 * float64(epochs)
			fmt.Printf("   %s Using fallback estimate: ~%.4f SUI\n", icons.Info, fallbackCost)
		} else {
			// Display detailed cost breakdown
			fmt.Printf("%s\n", FormatCostBreakdown(*breakdown))
			fmt.Printf("\n%s Summary: %s\n\n", icons.Info,
				FormatCostSummary(breakdown.GasCostSUI+breakdown.TotalWAL, breakdown.FileCount, epochs))

			// Show practical guidance
			if breakdown.GasCostSUI+breakdown.TotalWAL > 0.5 {
				fmt.Printf("%s %s Tip: Consider using `update-resources` for small changes\n", icons.Lightbulb, icons.Info)
				fmt.Printf("%s %s Tip: Use longer epochs for storage duration efficiency\n", icons.Lightbulb, icons.Info)
			}
		}
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found in PATH", siteBuilderCmd)
	}

	// For new sites, use 'site-builder --context <network> publish <directory> --epochs <number>'
	// Get context from Sui active environment to ensure correct network
	// Note: --context must come BEFORE the subcommand
	siteBuilderContext := GetWalrusContext()
	args := []string{
		"--context", siteBuilderContext,
		"publish",
		deployDir,
		"--epochs", fmt.Sprintf("%d", epochs),
	}

	if verboseMode {
		fmt.Printf("%s Verbose mode enabled\n", icons.Wrench)
		fmt.Printf("%s Builder path: %s\n", icons.Wrench, builderPath)
		fmt.Printf("%s Arguments: %v\n", icons.Wrench, args)
	}

	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, strings.Join(args, " "))
	fmt.Printf("%s Uploading site files to Walrus...\n", icons.Upload)
	fmt.Printf("%s This may take several minutes depending on file count and network conditions...\n", icons.Hourglass)
	fmt.Printf("   (timeout: %v)\n", DefaultCommandTimeout)
	fmt.Println()

	// Run with timeout to prevent hanging indefinitely
	stdoutStr, stderrStr, err := runCommandWithTimeout(ctx, builderPath, args, true)
	if err != nil {
		return nil, handleSiteBuilderError(err, stderrStr)
	}

	fmt.Printf("\n%s Site deployment command executed successfully.\n", icons.Success)

	// Parse the output - check both stdout and stderr as site-builder may output to either
	// Combine both streams for parsing since the "New site object ID:" line may be in stderr
	combinedOutput := stdoutStr + "\n" + stderrStr
	output := parseSiteBuilderOutput(combinedOutput)
	output.Success = true

	// Provide helpful guidance
	if output.ObjectID != "" {
		fmt.Printf("\n%s Deployment successful!\n", icons.Celebrate)
		fmt.Printf("%s Site Object ID: %s\n", icons.Clipboard, output.ObjectID)
		fmt.Printf("\n%s Next steps:\n", icons.Pencil)
		fmt.Printf("1. Save this Object ID in your walgo.yaml:\n")
		fmt.Printf("   walrus:\n")
		fmt.Printf("     projectID: \"%s\"\n", output.ObjectID)
		fmt.Printf("2. Configure a SuiNS domain: walgo domain <your-domain>\n")
		fmt.Printf("3. Update your site: walgo update\n")
		fmt.Printf("4. Check status: walgo status\n")

		if len(output.BrowseURLs) > 0 {
			fmt.Printf("\n%s Browse your site:\n", icons.Globe)
			for _, url := range output.BrowseURLs {
				fmt.Printf("   %s\n", url)
			}
		}
	}

	return output, nil
}

// UpdateSite handles updating an existing site on Walrus.
// It executes the `site-builder update` command.
// The context can be used to cancel or timeout the operation.
func UpdateSite(ctx context.Context, deployDir, objectID string, epochs int) (*SiteBuilderOutput, error) {
	// Validate objectID to prevent command injection
	if err := validateObjectID(objectID); err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}

	// Validate epochs
	if epochs <= 0 {
		return nil, fmt.Errorf("epochs must be greater than 0, got %d", epochs)
	}

	// Check setup first
	if err := CheckSiteBuilderSetup(); err != nil {
		return nil, fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	// site-builder --context <network> update --epochs <number> <directory> <object-id>
	// Get context from Sui active environment to ensure correct network
	// Note: --context must come BEFORE the subcommand
	siteBuilderContext := GetWalrusContext()
	args := []string{
		"--context", siteBuilderContext,
		"update",
		"--epochs", fmt.Sprintf("%d", epochs),
		deployDir,
		objectID,
	}

	icons := ui.GetIcons()
	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, strings.Join(args, " "))
	fmt.Printf("%s Updating site files on Walrus...\n", icons.Upload)

	// Calculate size for cost estimation
	var totalSize int64
	var fileCount int
	_ = filepath.Walk(deployDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})

	// Display basic cost estimate for updates
	sizeMB := float64(totalSize) / (1024 * 1024)
	simpleCost := sizeMB * 0.01 * float64(epochs)
	fmt.Printf("   %s Estimated cost: ~%.4f SUI (%d files, %.2f MB)\n", icons.Info, simpleCost, fileCount, sizeMB)

	fmt.Printf("%s This may take several minutes depending on file count and network conditions...\n", icons.Hourglass)
	fmt.Printf("   (timeout: %v)\n", DefaultCommandTimeout)
	fmt.Println()

	// Run with timeout to prevent hanging indefinitely
	stdoutStr, stderrStr, err := runCommandWithTimeout(ctx, builderPath, args, true)
	if err != nil {
		return nil, handleSiteBuilderError(err, stderrStr)
	}

	fmt.Printf("\n%s Site update command executed successfully.\n", icons.Success)

	// Parse the output - combine stdout and stderr for parsing
	combinedOutput := stdoutStr + "\n" + stderrStr
	output := parseSiteBuilderOutput(combinedOutput)
	output.Success = true
	output.ObjectID = objectID // For updates, we know the object ID

	fmt.Printf("\n%s Site updated successfully!\n", icons.Success)
	fmt.Printf("%s Object ID: %s\n", icons.Clipboard, objectID)
	fmt.Printf("%s Your site should be updated at the same URLs as before\n", icons.Globe)

	return output, nil
}

// DestroySite handles destroying an existing site on Walrus.
// It executes the `site-builder destroy` command.
// The context can be used to cancel or timeout the operation.
func DestroySite(ctx context.Context, objectID string) error {
	// Validate objectID to prevent command injection
	if err := validateObjectID(objectID); err != nil {
		return fmt.Errorf("invalid object ID: %w", err)
	}

	// Check setup first
	if err := CheckSiteBuilderSetup(); err != nil {
		return fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	// site-builder --context <network> destroy <object-id>
	// Get context from Sui active environment to ensure correct network
	// Note: --context must come BEFORE the subcommand
	siteBuilderContext := GetWalrusContext()
	args := []string{
		"--context", siteBuilderContext,
		"destroy",
		objectID,
	}

	icons := ui.GetIcons()
	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, strings.Join(args, " "))
	fmt.Printf("%s Destroying site on Walrus...\n", icons.Garbage)
	fmt.Println()

	cmd := execCommand(builderPath, args...)
	var stdout, stderr bytes.Buffer

	// Stream output in real-time while also capturing it for parsing
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	// Set up context cancellation
	if ctx != nil {
		// Start command
		if err := cmd.Start(); err != nil {
			return handleSiteBuilderError(err, stderr.String())
		}

		// Wait for either completion or context cancellation
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-ctx.Done():
			// Context cancelled, kill the process
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			return fmt.Errorf("destroy operation cancelled: %w", ctx.Err())
		case err := <-done:
			if err != nil {
				return handleSiteBuilderError(err, stderr.String())
			}
		}
	} else {
		// No context, just run normally
		if err := cmd.Run(); err != nil {
			return handleSiteBuilderError(err, stderr.String())
		}
	}

	fmt.Printf("\n%s Site destroyed successfully on Walrus!\n", icons.Success)

	return nil
}

// GetSiteStatus checks the status of a Walrus site.
// Note: The site-builder doesn't have a direct "status" command, but we can use sitemap.
func GetSiteStatus(objectID string) (*SiteBuilderOutput, error) {
	// Validate objectID to prevent command injection
	if err := validateObjectID(objectID); err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
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
	// Get context from Sui active environment to ensure correct network
	// Note: --context must come BEFORE the subcommand
	siteBuilderContext := GetWalrusContext()
	args := []string{
		"--context", siteBuilderContext,
		"sitemap",
		objectID,
	}

	icons := ui.GetIcons()
	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, strings.Join(args, " "))

	// Run with timeout (shorter timeout for status checks)
	statusTimeout := 2 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), statusTimeout)
	defer cancel()

	stdoutStr, stderrStr, err := runCommandWithTimeout(ctx, builderPath, args, false)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to execute %s: %v", siteBuilderCmd, err)
		if stderrStr != "" {
			errorMsg += fmt.Sprintf("\nstderr:\n%s", stderrStr)
		}
		if stdoutStr != "" {
			errorMsg += fmt.Sprintf("\nstdout:\n%s", stdoutStr)
		}
		return nil, fmt.Errorf("%s", errorMsg)
	}

	fmt.Println("Site status retrieved successfully.")

	// Parse the sitemap output
	output := parseSitemapOutput(stdoutStr)
	output.Success = true
	output.ObjectID = objectID

	if stdoutStr != "" {
		fmt.Printf("Site resources:\n%s\n", stdoutStr)
	}

	if stderrStr != "" {
		fmt.Printf("Stderr from %s:\n%s\n", siteBuilderCmd, stderrStr)
	}

	return output, nil
}

// parseSiteBuilderOutput extracts key information from site-builder command output
// This parser is designed to be resilient to format changes in site-builder output
func parseSiteBuilderOutput(output string) *SiteBuilderOutput {
	result := &SiteBuilderOutput{
		BrowseURLs: make([]string, 0),
		Resources:  make([]Resource, 0),
	}

	lines := strings.Split(output, "\n")

	// Strict patterns for site object ID - these are the ONLY patterns we trust
	// Ordered by priority (most specific first)
	siteObjectPatterns := []string{
		"New site object ID:",
		"Site object ID:",
		"site object ID:",
	}

	// Regex to extract hex object IDs (0x followed by 64 hex chars)
	objectIDRegex := regexp.MustCompile(`0x[0-9a-fA-F]{64}`)

	// First pass: Look ONLY for the specific "site object ID" patterns
	// Parse in reverse order to get the LAST matching pattern (most relevant)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])

		if result.ObjectID == "" {
			for _, pattern := range siteObjectPatterns {
				if strings.Contains(line, pattern) {
					if match := objectIDRegex.FindString(line); match != "" {
						result.ObjectID = match
						break
					}
				}
			}
		}
	}

	// Second pass: Extract browse URLs (scan forward)
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract browse URLs
		if strings.Contains(line, "http://") || strings.Contains(line, "https://") {
			// Extract URLs from the line
			urlRegex := regexp.MustCompile(`https?://[^\s\]\)\"\']+`)
			urls := urlRegex.FindAllString(line, -1)
			for _, url := range urls {
				// Clean up trailing punctuation
				url = strings.TrimRight(url, ".,;:")
				result.BrowseURLs = append(result.BrowseURLs, url)
			}
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
