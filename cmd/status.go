package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployer"
	sb "github.com/selimozten/walgo/internal/deployer/sitebuilder"
	"github.com/selimozten/walgo/internal/ui"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [object-id]",
	Short: "Check the status and resources of a Walrus Site.",
	Long: `Checks and displays the current status and resources of your site on Walrus Sites.
This command uses the site-builder's 'sitemap' command to show the resources that compose the site.

You can provide the object ID as an argument, or the command will look for it in walgo.yaml.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		var objectID string

		if len(args) > 0 {
			objectID = args[0]
			fmt.Printf("Checking status for object ID: %s\n", objectID)
		} else {

			sitePath, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
				return fmt.Errorf("error getting cwd: %w", err)
			}

			cfg, err := config.LoadConfigFrom(sitePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return fmt.Errorf("error loading config: %w", err)
			}

			if cfg.WalrusConfig.ProjectID == "" || cfg.WalrusConfig.ProjectID == "YOUR_WALRUS_PROJECT_ID" {
				fmt.Fprintf(os.Stderr, "No object ID provided and no valid ProjectID in walgo.yaml.\n")
				fmt.Fprintf(os.Stderr, "Usage: walgo status <object-id>\n")
				fmt.Fprintf(os.Stderr, "Or configure the ProjectID in walgo.yaml if it represents a site object ID.\n")
				return fmt.Errorf("no object ID provided and no valid ProjectID in walgo.yaml")
			}

			objectID = cfg.WalrusConfig.ProjectID
			fmt.Printf("Using object ID from walgo.yaml: %s\n", objectID)
		}

		d := sb.New()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		output, err := d.Status(ctx, objectID, deployer.DeployOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting site status: %v\n", err)
			return fmt.Errorf("error getting site status: %w", err)
		}

		if output.Success {
			fmt.Printf("\n%s Site Status Summary:\n", icons.Info)
			fmt.Printf("%s Object ID: %s\n", icons.File, objectID)
			if output.ResourceCount > 0 {
				fmt.Printf("%s Resources: %d files\n", icons.Folder, output.ResourceCount)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
