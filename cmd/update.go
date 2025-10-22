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

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [object-id]",
	Short: "Update an existing Walrus Site with new content.",
	Long: `Updates an existing Walrus Site with the content from the Hugo public directory.
This is more efficient than deploying a new site when you want to update existing content.

You can provide the object ID as an argument, or the command will use the ProjectID from walgo.yaml.
Assumes the site has been built using 'walgo build'.`,
	Args: cobra.MaximumNArgs(1), // Optional object ID argument
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executing update command...")

		var objectID string

		// Get object ID from argument or config
		if len(args) > 0 {
			objectID = args[0]
			fmt.Printf("Updating site with object ID: %s\n", objectID)
		} else {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if cfg.WalrusConfig.ProjectID == "" || cfg.WalrusConfig.ProjectID == "YOUR_WALRUS_PROJECT_ID" {
				fmt.Fprintf(os.Stderr, "No object ID provided and no valid ProjectID in walgo.yaml.\n")
				fmt.Fprintf(os.Stderr, "Usage: walgo update <object-id>\n")
				fmt.Fprintf(os.Stderr, "Or configure the ProjectID in walgo.yaml with your site's object ID.\n")
				os.Exit(1)
			}

			objectID = cfg.WalrusConfig.ProjectID
			fmt.Printf("Using object ID from walgo.yaml: %s\n", objectID)
		}

		// Determine site path (current directory by default)
		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		// Load Walgo configuration for deploy directory
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		// Determine the directory to deploy (e.g., "public")
		deployDir := filepath.Join(sitePath, cfg.HugoConfig.PublishDir)
		if _, err := os.Stat(deployDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Publish directory '%s' not found. Please run 'walgo build' first.\n", deployDir)
			os.Exit(1)
		}

		fmt.Printf("Preparing to update site with content from: %s\n", deployDir)

		// Get epochs flag
		epochs, err := cmd.Flags().GetInt("epochs")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading epochs flag: %v\n", err)
			os.Exit(1)
		}
		if epochs <= 0 {
			epochs = 1 // Default to 1 epoch
		}

		fmt.Printf("Storing for %d epoch(s)\n", epochs)

		d := sb.New()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		output, err := d.Update(ctx, deployDir, objectID, deployer.DeployOptions{Epochs: epochs, WalrusCfg: cfg.WalrusConfig})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating Walrus Site: %v\n", err)
			os.Exit(1)
		}

		if output.Success {
			fmt.Println("\n🎉 Site update completed successfully!")
			fmt.Printf("📋 Object ID: %s\n", objectID)
			fmt.Println("🌐 Your updated site should be available at the same URLs as before")
			fmt.Println("Use 'walgo status' to check the updated resources.")
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().IntP("epochs", "e", 1, "Number of epochs to store the site data (default: 1)")
}
