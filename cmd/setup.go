package cmd

import (
	"fmt"
	"os"

	"walgo/internal/walrus"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup [network]",
	Short: "Set up site-builder configuration for Walrus Sites.",
	Long: `Sets up the site-builder configuration file required for deploying to Walrus Sites.
This command creates the sites-config.yaml file with the correct package IDs and network settings.

Available networks:
  testnet  - Walrus Testnet (default, recommended for development)
  mainnet  - Walrus Mainnet (for production deployments)

The configuration will be created at ~/.config/walrus/sites-config.yaml`,
	Args: cobra.MaximumNArgs(1), // Optional network argument
	Run: func(cmd *cobra.Command, args []string) {
		var network string
		if len(args) > 0 {
			network = args[0]
		} else {
			// Get network from flag or default to testnet
			network, _ = cmd.Flags().GetString("network")
		}

		if network == "" {
			network = "testnet" // Default to testnet
		}

		fmt.Printf("Setting up site-builder configuration for %s...\n", network)

		// Check if already configured
		if err := walrus.CheckSiteBuilderSetup(); err == nil {
			fmt.Println("✓ site-builder is already configured!")
			fmt.Println("If you want to reconfigure, delete ~/.config/walrus/sites-config.yaml first")
			return
		}

		// Setup site-builder configuration
		if err := walrus.SetupSiteBuilder(network); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up site-builder: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n🎉 Setup complete!")
		fmt.Println("You can now deploy sites with: walgo deploy")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().StringP("network", "n", "testnet", "Network to configure (testnet or mainnet)")
}
