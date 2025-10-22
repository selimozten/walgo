package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"walgo/internal/config"
	"walgo/internal/deployer"
	sb "walgo/internal/deployer/sitebuilder"

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
		// Get current working directory
		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: Cannot determine current directory: %v\n", err)
			os.Exit(1)
		}

		// Load Walgo configuration
		walgoCfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n💡 Tip: Run 'walgo init <site-name>' to create a new site\n")
			os.Exit(1)
		}

		// Get flags
		epochs, err := cmd.Flags().GetInt("epochs")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading epochs flag: %v\n", err)
			os.Exit(1)
		}
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading force flag: %v\n", err)
			os.Exit(1)
		}
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading verbose flag: %v\n", err)
			os.Exit(1)
		}

		// Prepare deployer
		d := sb.New()

		// Check if public directory exists
		publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
		if _, err := os.Stat(publishDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "❌ Error: Build directory '%s' not found\n\n", publishDir)
			fmt.Fprintf(os.Stderr, "💡 Run this first:\n")
			fmt.Fprintf(os.Stderr, "   walgo build\n")
			if !force {
				os.Exit(1)
			}
		}

		fmt.Println("🚀 Deploying to Walrus Sites...")
		fmt.Println("  [1/3] Checking environment...")

		// Deploy the site via adapter interface
		fmt.Println("  [2/3] Uploading site...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		output, err := d.Deploy(ctx, publishDir, deployer.DeployOptions{Epochs: epochs, Verbose: verbose, WalrusCfg: walgoCfg.WalrusConfig})
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Deployment failed: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "💡 Troubleshooting:\n")
			fmt.Fprintf(os.Stderr, "  - Check setup: walgo doctor\n")
			fmt.Fprintf(os.Stderr, "  - Verify wallet: sui client active-address\n")
			fmt.Fprintf(os.Stderr, "  - Check gas: sui client gas\n")
			fmt.Fprintf(os.Stderr, "  - Try HTTP deploy: walgo deploy-http --help\n")
			os.Exit(1)
		}

		if output.Success && output.ObjectID != "" {
			fmt.Println("  [3/3] Confirming deployment...")
			fmt.Println("  ✓ Deployment confirmed")
			fmt.Printf("\n🎉 Deployment successful!\n\n")
			fmt.Printf("📋 Site Object ID: %s\n", output.ObjectID)
			fmt.Printf("\n💡 Next steps:\n")
			fmt.Printf("1. Update walgo.yaml with this Object ID:\n")
			fmt.Printf("   walrus:\n")
			fmt.Printf("     projectID: \"%s\"\n", output.ObjectID)
			fmt.Printf("\n2. Configure a domain: walgo domain <your-domain>\n")
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
