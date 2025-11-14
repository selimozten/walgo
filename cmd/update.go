package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"walgo/internal/cache"
	"walgo/internal/config"
	"walgo/internal/deployer"
	sb "walgo/internal/deployer/sitebuilder"
	"walgo/internal/metrics"

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
		// Initialize telemetry if enabled
		telemetry, _ := cmd.Flags().GetBool("telemetry")
		var collector *metrics.Collector
		var startTime time.Time
		var deployMetrics metrics.DeployMetrics
		success := false
		if telemetry {
			collector = metrics.New(true)
			startTime = collector.Start()
			defer func() {
				if collector != nil {
					_ = collector.RecordDeploy(startTime, &deployMetrics, success)
				}
			}()
		}

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

		// Get verbose flag
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading verbose flag: %v\n", err)
			os.Exit(1)
		}

		// Initialize cache helper
		cacheHelper, err := cache.NewDeployHelper(sitePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: Cache initialization failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "  Continuing without incremental build optimization...\n")
		} else {
			defer cacheHelper.Close()
		}

		// Check for dry-run mode
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading dry-run flag: %v\n", err)
			os.Exit(1)
		}

		// Prepare deployment plan
		if cacheHelper != nil {
			fmt.Println("\n📊 Analyzing changes...")
			plan, err := cacheHelper.PrepareDeployment(deployDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to analyze changes: %v\n", err)
			} else {
				if telemetry {
					deployMetrics.TotalFiles = plan.TotalFiles
					if plan.ChangeSet != nil {
						deployMetrics.ChangedFiles = len(plan.ChangeSet.Added) + len(plan.ChangeSet.Modified)
					}
				}

				if verbose {
					plan.PrintVerboseSummary()
				} else {
					plan.PrintSummary()
				}

				// If dry-run, stop here
				if dryRun {
					fmt.Println("\n🔍 Dry-run mode: No files will be uploaded")
					fmt.Printf("📋 Would update site: %s\n", objectID)
					fmt.Println("✅ Update plan complete!")
					fmt.Printf("\n💡 To actually update, run without --dry-run flag\n")
					os.Exit(0)
				}
			}
		} else if dryRun {
			fmt.Println("\n⚠️  Note: Dry-run without cache - cannot show file-level changes")
			fmt.Printf("🔍 Would update site %s with all files in: %s\n", objectID, deployDir)
			fmt.Println("\n💡 To see detailed changes, ensure cache is enabled")
			os.Exit(0)
		}

		// Get epochs flag
		epochs, err := cmd.Flags().GetInt("epochs")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading epochs flag: %v\n", err)
			os.Exit(1)
		}
		if epochs <= 0 {
			epochs = 1 // Default to 1 epoch
		}

		fmt.Printf("\nStoring for %d epoch(s)\n", epochs)

		uploadStart := time.Now()
		d := sb.New()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		output, err := d.Update(ctx, deployDir, objectID, deployer.DeployOptions{Epochs: epochs, WalrusCfg: cfg.WalrusConfig})
		if telemetry {
			deployMetrics.UploadDuration = time.Since(uploadStart).Milliseconds()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating Walrus Site: %v\n", err)
			os.Exit(1)
		}

		if output.Success {
			// Mark update as successful
			success = true

			// Update cache with deployment info
			if cacheHelper != nil {
				fmt.Println("\n📝 Updating cache...")
				err := cacheHelper.FinalizeDeployment(deployDir, objectID, objectID, output.FileToBlobID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to update cache: %v\n", err)
				} else {
					fmt.Println("  ✓ Cache updated")
				}
			}

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
	updateCmd.Flags().BoolP("verbose", "v", false, "Show detailed change summary")
	updateCmd.Flags().Bool("dry-run", false, "Preview update plan without actually updating")
	updateCmd.Flags().Bool("telemetry", false, "Record update metrics to local JSON file (~/.walgo/metrics.json)")
}
