package cmd

import (
	"fmt"
	"os"

	"github.com/selimozten/walgo/internal/config"

	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

var domainCmd = &cobra.Command{
	Use:   "domain [domain-name]",
	Short: "Get instructions for configuring SuiNS domain for your Walrus Site (mainnet only).",
	Long: `Provides instructions for configuring a SuiNS domain name to point to your Walrus Site.

NOTE: SuiNS is only available on mainnet. If you're using testnet, please deploy to mainnet
first to use SuiNS domain names.

Since SuiNS domain management is handled through the SuiNS web interface, this command
provides step-by-step instructions and the necessary object ID from your configuration.

For Mainnet: https://suins.io`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		var domainName string
		if len(args) > 0 {
			domainName = args[0]
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("error loading config: %w", err)
		}

		// Check network - SuiNS is mainnet only
		network := cfg.WalrusConfig.Network
		if network == "" {
			network = "testnet" // Default to testnet
		}

		if network != "mainnet" {
			fmt.Printf("%s SuiNS is only available on mainnet.\n", icons.Warning)
			fmt.Printf("\n%s Your current network: %s\n", icons.Info, network)
			fmt.Printf("%s To use SuiNS, you need to deploy your site to mainnet.\n", icons.Lightbulb)
			fmt.Println()
			fmt.Println("To deploy to mainnet:")
			fmt.Println("  walgo config set network mainnet")
			fmt.Println("  walgo launch")
			fmt.Println()
			return nil
		}

		if cfg.WalrusConfig.ProjectID == "" || cfg.WalrusConfig.ProjectID == "YOUR_WALRUS_PROJECT_ID" {
			fmt.Fprintf(os.Stderr, "%s Error: Walrus ProjectID is not set in walgo.yaml\n", icons.Error)
			fmt.Fprintf(os.Stderr, "\n%s Deploy your site first to get an object ID:\n", icons.Lightbulb)
			fmt.Fprintf(os.Stderr, "   walgo launch\n")
			return fmt.Errorf("walrus projectid is not set in walgo.yaml")
		}

		fmt.Printf("%s SuiNS Domain Configuration (Mainnet)\n", icons.Globe)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%s Object ID: %s\n", icons.File, cfg.WalrusConfig.ProjectID)

		if domainName != "" {
			fmt.Printf("%s Domain: %s\n", icons.Info, domainName)
		}
		fmt.Println()

		fmt.Printf("%s Setup Steps:\n", icons.Pencil)
		fmt.Println()
		fmt.Println("  1  Visit SuiNS:")
		fmt.Println("      • https://suins.io")
		fmt.Println()
		if domainName == "" {
			fmt.Println("  2  Purchase or select your domain")
		} else {
			fmt.Printf("  2  Select domain: %s\n", domainName)
		}
		fmt.Println()
		fmt.Println("  3  Click 'Link To Walrus Site'")
		fmt.Println()
		fmt.Printf("  4  Enter Object ID: %s\n", cfg.WalrusConfig.ProjectID)
		fmt.Println()
		fmt.Println("  5  Approve transaction in your wallet")
		fmt.Println()
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		if domainName != "" {
			fmt.Printf("%s Your site will be available at: https://%s.wal.app\n", icons.Success, domainName)
		} else {
			fmt.Printf("%s Your site will be available at: https://YOUR-DOMAIN.wal.app\n", icons.Success)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(domainCmd)
}
