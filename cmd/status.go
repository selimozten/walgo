package cmd

import (
	"fmt"
	"os"

	"walgo/internal/config"
	"walgo/internal/walrus"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [object-id]",
	Short: "Check the status and resources of a Walrus Site.",
	Long: `Checks and displays the current status and resources of your site on Walrus Sites.
This command uses the site-builder's 'sitemap' command to show the resources that compose the site.

You can provide the object ID as an argument, or the command will look for it in walgo.yaml.`,
	Args: cobra.MaximumNArgs(1), // Optional object ID argument
	Run: func(cmd *cobra.Command, args []string) {
		var objectID string

		// Get object ID from argument or config
		if len(args) > 0 {
			objectID = args[0]
			fmt.Printf("Checking status for object ID: %s\n", objectID)
		} else {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if cfg.WalrusConfig.ProjectID == "" || cfg.WalrusConfig.ProjectID == "YOUR_WALRUS_PROJECT_ID" {
				fmt.Fprintf(os.Stderr, "No object ID provided and no valid ProjectID in walgo.yaml.\n")
				fmt.Fprintf(os.Stderr, "Usage: walgo status <object-id>\n")
				fmt.Fprintf(os.Stderr, "Or configure the ProjectID in walgo.yaml if it represents a site object ID.\n")
				os.Exit(1)
			}

			objectID = cfg.WalrusConfig.ProjectID
			fmt.Printf("Using object ID from walgo.yaml: %s\n", objectID)
		}

		// Get site status/resources using sitemap
		output, err := walrus.GetSiteStatus(objectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting site status: %v\n", err)
			os.Exit(1)
		}

		if output.Success {
			fmt.Printf("\n📊 Site Status Summary:\n")
			fmt.Printf("📋 Object ID: %s\n", objectID)
			if len(output.Resources) > 0 {
				fmt.Printf("📁 Resources: %d files\n", len(output.Resources))
			}
		}

		// If the --convert flag is set, also show the Base36 representation
		if convert, _ := cmd.Flags().GetBool("convert"); convert {
			fmt.Println("\nConverting to Base36 format:")
			if base36, err := walrus.ConvertObjectID(objectID); err != nil {
				fmt.Fprintf(os.Stderr, "Error converting object ID: %v\n", err)
			} else if base36 != "" {
				fmt.Printf("🔗 Base36 ID: %s\n", base36)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolP("convert", "c", false, "Also show the Base36 representation of the object ID")
}
