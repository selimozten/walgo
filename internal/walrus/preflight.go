package walrus

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/selimozten/walgo/internal/sui"
	"github.com/selimozten/walgo/internal/ui"
)

// CheckSiteBuilderSetup verifies site-builder installation and configuration.
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

// handleSiteBuilderError converts site-builder errors into actionable messages.
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

// PreflightCheck validates environment before deployment.
func PreflightCheck() error {
	icons := ui.GetIcons()
	fmt.Printf("%s Running pre-flight checks...\n", icons.Search)

	walrusPath, err := execLookPath("walrus")
	if err != nil {
		return fmt.Errorf("walrus CLI not found in PATH")
	}
	fmt.Printf("   %s Walrus CLI found at: %s\n", icons.Check, walrusPath)

	walrusContext := GetWalrusContext()
	infoCmd := execCommand("walrus", "info", "--json", "--context", walrusContext)
	infoCmd.Stdout = nil
	infoCmd.Stderr = nil
	if err := infoCmd.Run(); err != nil {
		return fmt.Errorf("cannot connect to Walrus network (%s)", walrusContext)
	}
	fmt.Printf("   %s Walrus network is reachable (%s)\n", icons.Check, walrusContext)

	suiPath, err := execLookPath("sui")
	if err != nil {
		return fmt.Errorf("sui CLI not found. Install via suiup:\n   curl -sSfL https://raw.githubusercontent.com/MystenLabs/suiup/main/install.sh | sh\n   suiup install sui@testnet\n\n   Or run: walgo setup-deps")
	}
	fmt.Printf("   %s Sui CLI found at: %s\n", icons.Check, suiPath)

	balance, err := sui.GetBalance()
	if err != nil {
		fmt.Printf("   %s Could not check wallet balance (continuing anyway)\n", icons.Warning)
	} else {
		if balance.SUI > 0 {
			fmt.Printf("   %s SUI balance: %.2f SUI\n", icons.Check, balance.SUI)
		} else {
			fmt.Printf("   %s Warning: No SUI balance detected. You may need testnet SUI from:\n", icons.Warning)
			fmt.Println("      https://discord.com/channels/916379725201563759/971488439931392130")
		}

		if balance.WAL > 0 {
			fmt.Printf("   %s WAL balance: %.2f WAL\n", icons.Check, balance.WAL)
		} else {
			fmt.Printf("   %s No WAL tokens found (may need for storage quota)\n", icons.Info)
		}
	}

	fmt.Printf("%s Pre-flight checks passed\n", icons.Success)
	return nil
}
