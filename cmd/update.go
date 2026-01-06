package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/selimozten/walgo/internal/cache"
	"github.com/selimozten/walgo/internal/compress"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployer"
	sb "github.com/selimozten/walgo/internal/deployer/sitebuilder"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/metrics"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [object-id]",
	Short: "Update an existing Walrus Site with new content.",
	Long: `Updates an existing Walrus Site with the content from the Hugo public directory.
This is more efficient than deploying a new site when you want to update existing content.

You can provide the object ID as an argument, or the command will use the ProjectID from walgo.yaml.
Assumes the site has been built using 'walgo build'.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
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

		fmt.Printf("%s Executing update command...\n", icons.Rocket)

		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		cfg, err := config.LoadConfigFrom(sitePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("error loading config: %w", err)
		}

		deployDir := filepath.Join(sitePath, cfg.HugoConfig.PublishDir)
		if _, err := os.Stat(deployDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s Error: Publish directory '%s' not found\n", icons.Error, deployDir)
			fmt.Fprintf(os.Stderr, "%s Run 'walgo build' first\n", icons.Lightbulb)
			return fmt.Errorf("publish directory not found: %w", err)
		}

		// Get object ID with priority: CLI arg > ws-resources.json > walgo.yaml
		var objectID string
		if len(args) > 0 {
			objectID = args[0]
			fmt.Printf("%s Using object ID from command argument: %s\n", icons.Info, objectID)
		} else {
			wsResourcesPath := filepath.Join(deployDir, "ws-resources.json")
			wsConfig, err := compress.ReadWSResourcesConfig(wsResourcesPath)
			if err == nil && wsConfig.ObjectID != "" {
				objectID = wsConfig.ObjectID
				fmt.Printf("%s Using object ID from ws-resources.json: %s\n", icons.Info, objectID)
			} else if cfg.WalrusConfig.ProjectID != "" && cfg.WalrusConfig.ProjectID != "YOUR_WALRUS_PROJECT_ID" {
				objectID = cfg.WalrusConfig.ProjectID
				fmt.Printf("%s Using object ID from walgo.yaml: %s\n", icons.Info, objectID)
			} else {
				fmt.Fprintf(os.Stderr, "%s Error: No object ID found in:\n", icons.Error)
				fmt.Fprintf(os.Stderr, "  %s Command arguments\n", icons.Cross)
				fmt.Fprintf(os.Stderr, "  %s ws-resources.json (%s)\n", icons.Cross, wsResourcesPath)
				fmt.Fprintf(os.Stderr, "  %s walgo.yaml (WalrusConfig.ProjectID)\n", icons.Cross)
				fmt.Fprintf(os.Stderr, "\n%s Usage: walgo update <object-id>\n", icons.Lightbulb)
				fmt.Fprintf(os.Stderr, "  Or run 'walgo launch' first to create a site.\n")
				return fmt.Errorf("no object ID found")
			}
		}

		fmt.Printf("%s Preparing to update site with content from: %s\n", icons.Package, deployDir)

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading verbose flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading verbose flag: %w", err)
		}

		cacheHelper, err := cache.NewDeployHelper(sitePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Cache initialization failed: %v\n", icons.Warning, err)
			fmt.Fprintf(os.Stderr, "  Continuing without incremental build optimization...\n")
		} else {
			defer cacheHelper.Close()
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading dry-run flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading dry-run flag: %w", err)
		}

		if cacheHelper != nil {
			fmt.Printf("\n%s Analyzing changes...\n", icons.Info)
			plan, err := cacheHelper.PrepareDeployment(deployDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Failed to analyze changes: %v\n", icons.Warning, err)
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

				if dryRun {
					fmt.Printf("\n%s Dry-run mode: No files will be uploaded\n", icons.Info)
					fmt.Printf("%s Would update site: %s\n", icons.File, objectID)
					fmt.Printf("%s Update plan complete!\n", icons.Success)
					fmt.Printf("\n%s To actually update, run without --dry-run flag\n", icons.Lightbulb)
					return nil
				}
			}
		} else if dryRun {
			fmt.Printf("\n%s Note: Dry-run without cache - cannot show file-level changes\n", icons.Warning)
			fmt.Printf("%s Would update site %s with all files in: %s\n", icons.Info, objectID, deployDir)
			fmt.Printf("\n%s To see detailed changes, ensure cache is enabled\n", icons.Lightbulb)
			return nil
		}

		epochs, err := cmd.Flags().GetInt("epochs")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading epochs flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading epochs flag: %w", err)
		}
		if epochs <= 0 {
			epochs = 1 // Default to 1 epoch
		}

		fmt.Printf("\n%s Storing for %d epoch(s)\n", icons.Database, epochs)

		err = hugo.BuildSite(sitePath)
		if err != nil {
			return fmt.Errorf("failed to build site: %w", err)
		}

		uploadStart := time.Now()
		d := sb.New()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		output, err := d.Update(ctx, deployDir, objectID, deployer.DeployOptions{Epochs: epochs, WalrusCfg: cfg.WalrusConfig})
		if telemetry {
			deployMetrics.UploadDuration = time.Since(uploadStart).Milliseconds()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: updating Walrus Site: %v\n", icons.Error, err)
			return fmt.Errorf("error updating Walrus Site: %w", err)
		}

		if output.Success {
			success = true

			if cacheHelper != nil {
				fmt.Printf("\n%s Updating cache...\n", icons.Pencil)
				err := cacheHelper.FinalizeDeployment(deployDir, objectID, objectID, output.FileToBlobID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s Warning: Failed to update cache: %v\n", icons.Warning, err)
				} else {
					fmt.Printf("  %s Cache updated\n", icons.Check)
				}

				fmt.Printf("\n%s Updating metadata...\n", icons.Package)
				wsResourcesPath := filepath.Join(deployDir, "ws-resources.json")
				if err := compress.UpdateObjectID(wsResourcesPath, objectID); err != nil {
					fmt.Fprintf(os.Stderr, "%s Error: Failed to save objectID to ws-resources.json: %v\n", icons.Error, err)
					return fmt.Errorf("failed to save objectID to ws-resources.json: %w", err)
				}
				fmt.Printf("  %s ObjectID saved to ws-resources.json\n", icons.Check)

				// Update walgo.yaml with objectID
				if err := config.UpdateWalgoYAMLProjectID(sitePath, objectID); err != nil {
					fmt.Fprintf(os.Stderr, "%s Error: Failed to update walgo.yaml: %v\n", icons.Error, err)
					return fmt.Errorf("failed to update walgo.yaml with Object ID: %w", err)
				}
				fmt.Printf("  %s Updated walgo.yaml\n", icons.Check)

				pm, err := projects.NewManager()
				if err == nil {
					defer pm.Close()

					existingProj, err := pm.GetProjectBySitePath(sitePath)
					if err == nil && existingProj != nil {
						var siteSize int64
						if err := filepath.Walk(deployDir, func(path string, info os.FileInfo, err error) error {
							if err == nil && !info.IsDir() {
								siteSize += info.Size()
							}
							return nil
						}); err != nil {
							// Log warning but continue
							fmt.Fprintf(os.Stderr, "%s Warning: Failed to calculate site size: %v\n", icons.Warning, err)
						}

						existingProj.ObjectID = objectID
						existingProj.Epochs = epochs
						existingProj.LastDeployAt = time.Now()

						if err := pm.UpdateProject(existingProj); err != nil {
							fmt.Fprintf(os.Stderr, "%s Warning: Failed to update project in database: %v\n", icons.Warning, err)
						} else {
							// Use epoch-aware cost estimation
							estimatedGas := projects.EstimateGasFeeWithEpochs(existingProj.Network, siteSize, epochs)
							deployment := &projects.DeploymentRecord{
								ProjectID: existingProj.ID,
								ObjectID:  objectID,
								Network:   existingProj.Network,
								Epochs:    epochs,
								GasFee:    estimatedGas,
								Success:   true,
							}
							_ = pm.RecordDeployment(deployment)

							fmt.Printf("  %s Project updated in database\n", icons.Check)
						}
					}
				}
			}

			fmt.Printf("\n%s Site update completed successfully!\n", icons.Celebrate)
			fmt.Printf("%s Object ID: %s\n", icons.File, objectID)
			fmt.Printf("%s Your updated site should be available at the same URLs as before\n", icons.Globe)
			fmt.Println("Use 'walgo status' to check the updated resources.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().IntP("epochs", "e", 1, "Number of epochs to store the site data (default: 1)")
	updateCmd.Flags().BoolP("verbose", "v", false, "Show detailed change summary")
	updateCmd.Flags().Bool("dry-run", false, "Preview update plan without actually updating")
	updateCmd.Flags().Bool("telemetry", false, "Record update metrics to local JSON file (~/.walgo/metrics.json)")
}
