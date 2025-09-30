package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"walgo/internal/config"
	"walgo/internal/walrus"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your Hugo site to Walrus Sites.",
	Long: `Deploys your Hugo site to Walrus Sites decentralized storage.
This command builds your site and uploads it to the Walrus network.

The site will be stored for the specified number of epochs (default: 1).
After deployment, you'll receive an object ID that you can use to access
your site and configure domain names.

Example: walgo deploy --epochs 5`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Deploying site to Walrus Sites...")

		// Get current working directory
		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		// Load Walgo configuration
		walgoCfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		// Get flags
		epochs, _ := cmd.Flags().GetInt("epochs")
		force, _ := cmd.Flags().GetBool("force")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Set verbose mode in walrus package
		walrus.SetVerbose(verbose)

		// Check if public directory exists
		publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
		if _, err := os.Stat(publishDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Public directory '%s' not found. Please run 'walgo build' first.\n", publishDir)
			if !force {
				os.Exit(1)
			}
		}

		// Deploy the site
		output, err := walrus.DeploySite(publishDir, walgoCfg.WalrusConfig, epochs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deploying site: %v\n", err)
			os.Exit(1)
		}

		if output.Success && output.ObjectID != "" {
			fmt.Printf("\n🎉 Deployment successful!\n")
			fmt.Printf("📋 Site Object ID: %s\n", output.ObjectID)
			fmt.Printf("\n💡 Next steps:\n")
			fmt.Printf("1. Update walgo.yaml with this Object ID:\n")
			fmt.Printf("   walrus:\n")
			fmt.Printf("     projectID: \"%s\"\n", output.ObjectID)
			fmt.Printf("2. Configure a domain: walgo domain <your-domain>\n")
			fmt.Printf("3. Check status: walgo status\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().IntP("epochs", "e", 1, "Number of epochs to store the site")
	deployCmd.Flags().BoolP("force", "f", false, "Deploy even if public directory doesn't exist")
	deployCmd.Flags().BoolP("verbose", "v", false, "Show detailed output for debugging")
}
