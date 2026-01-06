package cmd

import (
	"fmt"
	"os"

	"github.com/selimozten/walgo/internal/ui"
	"github.com/selimozten/walgo/internal/walrus"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup [network]",
	Short: "Set up site-builder configuration for Walrus Sites.",
	Long: `Sets up the site-builder configuration file required for deploying to Walrus Sites.
This command creates the sites-config.yaml file with the correct package IDs and network settings.

Available networks:
  testnet  - Walrus Testnet (default, recommended for development)
  mainnet  - Walrus Mainnet (for production deployments)

The configuration will be created at ~/.config/walrus/sites-config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var network string
		if len(args) > 0 {
			network = args[0]
		} else {
			network, _ = cmd.Flags().GetString("network")
		}

		if network == "" {
			network = "testnet"
		}

		fmt.Printf("Setting up site-builder configuration for %s...\n", network)

		icons := ui.GetIcons()
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading force flag: %v\n", err)
			return fmt.Errorf("error reading force flag: %w", err)
		}

		if !force {
			if err := walrus.CheckSiteBuilderSetup(); err == nil {
				fmt.Printf("%s site-builder is already configured!\n", icons.Check)
				fmt.Println("Use --force to overwrite the configuration, or delete ~/.config/walrus/sites-config.yaml")
				return nil
			}
		}

		if err := walrus.SetupSiteBuilder(network, force); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up site-builder: %v\n", err)
			return fmt.Errorf("error setting up site-builder: %w", err)
		}

		fmt.Printf("\n%s Setup complete!\n", icons.Celebrate)
		fmt.Println("You can now deploy sites with: walgo launch")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().StringP("network", "n", "testnet", "Network to configure (testnet or mainnet)")
	setupCmd.Flags().Bool("force", false, "Overwrite existing site-builder configuration if present")
}
