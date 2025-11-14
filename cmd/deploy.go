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
		quiet, _ := cmd.Flags().GetBool("quiet")

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

		if !quiet {
			fmt.Println("🚀 Deploying to Walrus Sites...")
			fmt.Println("  [1/4] Checking environment...")
		}

		// Initialize cache helper
		cacheHelper, err := cache.NewDeployHelper(sitePath)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  Warning: Cache initialization failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "  Continuing without incremental build optimization...\n")
			}
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
		if cacheHelper != nil && !quiet {
			fmt.Println("  [2/4] Analyzing changes...")
			plan, err := cacheHelper.PrepareDeployment(publishDir)
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
					fmt.Println("✅ Deployment plan complete!")
					fmt.Printf("\n💡 To actually deploy, run without --dry-run flag\n")
					os.Exit(0)
				}
			}
		} else if dryRun && !quiet {
			// No cache helper but dry-run requested
			fmt.Println("\n⚠️  Note: Dry-run without cache - cannot show file-level changes")
			fmt.Printf("🔍 Would deploy all files in: %s\n", publishDir)
			fmt.Println("\n💡 To see detailed changes, ensure cache is enabled")
			os.Exit(0)
		}

		// Deploy the site via adapter interface
		if !quiet {
			if cacheHelper != nil {
				fmt.Println("  [3/4] Uploading site...")
			} else {
				fmt.Println("  [2/3] Uploading site...")
			}
		}
		uploadStart := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		output, err := d.Deploy(ctx, publishDir, deployer.DeployOptions{Epochs: epochs, Verbose: verbose && !quiet, WalrusCfg: walgoCfg.WalrusConfig})
		if telemetry {
			deployMetrics.UploadDuration = time.Since(uploadStart).Milliseconds()
		}
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
			// Mark deployment as successful
			success = true

			// Update cache with deployment info
			if cacheHelper != nil {
				if !quiet {
					if cacheHelper != nil {
						fmt.Println("  [4/4] Updating cache...")
					} else {
						fmt.Println("  [3/3] Confirming deployment...")
					}
				}

				err := cacheHelper.FinalizeDeployment(publishDir, output.ObjectID, output.ObjectID, output.FileToBlobID)
				if err != nil && !quiet {
					fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to update cache: %v\n", err)
				} else if !quiet {
					fmt.Println("  ✓ Cache updated")
				}
			} else if !quiet {
				fmt.Println("  [3/3] Confirming deployment...")
			}

			if !quiet {
				fmt.Println("  ✓ Deployment confirmed")
				fmt.Printf("\n🎉 Deployment successful!\n\n")
				fmt.Printf("📋 Site Object ID: %s\n", output.ObjectID)
				fmt.Printf("\n💡 Next steps:\n")
				fmt.Printf("1. Update walgo.yaml with this Object ID:\n")
				fmt.Printf("   walrus:\n")
				fmt.Printf("     projectID: \"%s\"\n", output.ObjectID)
				fmt.Printf("\n2. Configure a domain: walgo domain <your-domain>\n")
				fmt.Printf("3. Check status: walgo status\n")
			} else {
				// In quiet mode, just output the object ID for parsing by quickstart
				fmt.Printf("Site Object ID: %s\n", output.ObjectID)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().IntP("epochs", "e", 1, "Number of epochs to store the site")
	deployCmd.Flags().BoolP("force", "f", false, "Deploy even if public directory doesn't exist")
	deployCmd.Flags().BoolP("verbose", "v", false, "Show detailed output for debugging")
	deployCmd.Flags().BoolP("quiet", "q", false, "Suppress output (used internally by quickstart)")
	deployCmd.Flags().Bool("dry-run", false, "Preview deployment plan without actually deploying")
	deployCmd.Flags().Bool("telemetry", false, "Record deployment metrics to local JSON file (~/.walgo/metrics.json)")
}
