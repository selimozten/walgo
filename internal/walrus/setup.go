package walrus

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/ui"
)

// SetupSiteBuilder helps users set up the site-builder configuration.
// Uses the same approach as install.sh - downloads official configs from upstream.
func SetupSiteBuilder(network string, force bool) error {
	if network == "" {
		network = "testnet"
	}

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

	sitesConfigPath := filepath.Join(configDir, "sites-config.yaml")
	if _, err := os.Stat(sitesConfigPath); err == nil && !force {
		return fmt.Errorf("site-builder config already exists at %s. Use --force to overwrite", sitesConfigPath)
	}

	fmt.Printf("%s Downloading site-builder configuration...\n", icons.Download)

	configURL := "https://raw.githubusercontent.com/MystenLabs/walrus-sites/refs/heads/mainnet/sites-config.yaml"
	if err := downloadConfig(configURL, sitesConfigPath); err != nil {
		fmt.Printf("   %s Warning: Failed to download config: %v\n", icons.Warning, err)
		fmt.Println("   Creating config from template...")

		if err := createSiteBuilderConfigFromTemplate(sitesConfigPath, network); err != nil {
			return fmt.Errorf("failed to create site-builder config: %w", err)
		}
	} else {
		fmt.Printf("%s site-builder config downloaded\n", icons.Check)

		if network != "testnet" {
			if err := updateDefaultContext(sitesConfigPath, network); err != nil {
				fmt.Printf("   %s Warning: Failed to update default context: %v\n", icons.Warning, err)
			}
		}
	}

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

// checkWalrusDependencies checks if required tools are installed.
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

// downloadConfig downloads a configuration file from a URL.
func downloadConfig(url, destPath string) error {
	// Validate URL scheme
	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("only HTTPS URLs are allowed, got: %s", url)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url) // #nosec G107 - URL is validated above and comes from hardcoded sources
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// #nosec G306 - config file needs to be readable
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// createSiteBuilderConfigFromTemplate creates a config file from hardcoded template.
// This is a fallback when downloading the official config fails.
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
	walrusBinary := "walrus"
	if path, err := execLookPath("walrus"); err == nil {
		walrusBinary = path
	}

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

// updateDefaultContext updates the default_context field in a YAML config file.
func updateDefaultContext(configPath, network string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	oldContext := "default_context: testnet"
	newContext := fmt.Sprintf("default_context: %s", network)

	updated := strings.ReplaceAll(string(content), oldContext, newContext)

	// #nosec G306 - config file needs to be readable
	if err := os.WriteFile(configPath, []byte(updated), 0644); err != nil {
		return err
	}

	return nil
}
