package cmd

import (
	"fmt"
	"os"

	"walgo/internal/config"

	"github.com/spf13/cobra"
)

// domainCmd represents the domain command
var domainCmd = &cobra.Command{
	Use:   "domain [domain-name]",
	Short: "Get instructions for configuring SuiNS domain for your Walrus Site.",
	Long: `Provides instructions for configuring a SuiNS domain name to point to your Walrus Site.

Since SuiNS domain management is handled through the SuiNS web interface, this command 
provides step-by-step instructions and the necessary object ID from your configuration.

For Mainnet: https://suins.io
For Testnet: https://testnet.suins.io`,
	Args: cobra.MaximumNArgs(1), // Optional domain name argument
	Run: func(cmd *cobra.Command, args []string) {
		var domainName string
		if len(args) > 0 {
			domainName = args[0]
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if cfg.WalrusConfig.ProjectID == "" || cfg.WalrusConfig.ProjectID == "YOUR_WALRUS_PROJECT_ID" {
			fmt.Fprintf(os.Stderr, "Walrus ProjectID is not set in walgo.yaml. Please configure it first.\n")
			fmt.Fprintf(os.Stderr, "You need the object ID of your deployed Walrus Site to configure a domain.\n")
			os.Exit(1)
		}

		fmt.Println("🌐 SuiNS Domain Configuration for Walrus Sites")
		fmt.Println("==============================================")
		fmt.Printf("Your Walrus Site Object ID: %s\n\n", cfg.WalrusConfig.ProjectID)

		if domainName != "" {
			fmt.Printf("Setting up domain: %s\n\n", domainName)
		}

		fmt.Println("Steps to configure your SuiNS domain:")
		fmt.Println("1. Go to the SuiNS website:")
		fmt.Println("   • Mainnet: https://suins.io")
		fmt.Println("   • Testnet: https://testnet.suins.io")
		fmt.Println()

		if domainName == "" {
			fmt.Println("2. Purchase or select a domain name you own")
		} else {
			fmt.Printf("2. Purchase or select the domain: %s\n", domainName)
		}
		fmt.Println("   • Domain names can only contain letters (a-z) and numbers (0-9)")
		fmt.Println("   • No special characters like hyphens are allowed")
		fmt.Println()

		fmt.Println("3. In the 'Names you own' section:")
		fmt.Println("   • Click the 'three dots' menu icon above your domain")
		fmt.Println("   • Click 'Link To Walrus Site'")
		fmt.Println()

		fmt.Printf("4. Paste your Walrus Site Object ID: %s\n", cfg.WalrusConfig.ProjectID)
		fmt.Println("   • Double-check that the ID is correct")
		fmt.Println("   • Click 'Apply'")
		fmt.Println()

		fmt.Println("5. Approve the transaction in your wallet")
		fmt.Println()

		if domainName != "" {
			fmt.Printf("Once completed, your site will be available at: https://%s.wal.app\n\n", domainName)
		} else {
			fmt.Println("Once completed, your site will be available at: https://YOUR-DOMAIN.wal.app")
			fmt.Println()
		}

		fmt.Println("💡 Tips:")
		fmt.Println("• You can also update your walgo.yaml to include the domain:")
		fmt.Println("  walrus:")
		if domainName != "" {
			fmt.Printf("    suinsDomain: \"%s.sui\"\n", domainName)
		} else {
			fmt.Println("    suinsDomain: \"your-domain.sui\"")
		}
		fmt.Println("• Use 'walgo status' to check your site's resources")
		fmt.Println("• Your site is also accessible via Base36 encoding (use 'walgo status --convert')")
	},
}

func init() {
	rootCmd.AddCommand(domainCmd)
}
