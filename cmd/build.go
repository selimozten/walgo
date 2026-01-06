package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/selimozten/walgo/internal/compress"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/metrics"
	"github.com/selimozten/walgo/internal/optimizer"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Hugo site.",
	Long: `Builds the Hugo site using the configuration found in the current directory
(or the directory specified by global --config flag if walgo.yaml is there).
This command runs the 'hugo' command to generate static files typically into the 'public' directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		quiet, _ := cmd.Flags().GetBool("quiet")

		telemetry, _ := cmd.Flags().GetBool("telemetry")
		var collector *metrics.Collector
		var startTime time.Time
		var buildMetrics metrics.BuildMetrics
		success := false
		if telemetry {
			collector = metrics.New(true)
			startTime = collector.Start()
			defer func() {
				if collector != nil {
					_ = collector.RecordBuild(startTime, &buildMetrics, success)
				}
			}()
		}

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		walgoCfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s Did you run 'walgo init' to create a site?\n", icons.Lightbulb)
			return err
		}

		if !quiet {
			fmt.Printf("%s Building site...\n", icons.Package)
		}

		clean, err := cmd.Flags().GetBool("clean")
		if err != nil {
			return fmt.Errorf("error reading clean flag: %w", err)
		}
		if clean {
			publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
			if !quiet {
				fmt.Printf("  [1/3] Cleaning %s...\n", publishDir)
			}
			if err := os.RemoveAll(publishDir); err != nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "  %s Warning: Failed to clean: %v\n", icons.Warning, err)
				}
			} else if !quiet {
				fmt.Printf("  %s Cleaned\n", icons.Check)
			}
		}

		if !quiet {
			stepNum := 2
			if !clean {
				stepNum = 1
			}
			fmt.Printf("  [%d/3] Running Hugo build...\n", stepNum)
		}
		hugoStart := time.Now()
		if err := hugo.BuildSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "\n%s Troubleshooting:\n", icons.Lightbulb)
			fmt.Fprintf(os.Stderr, "  - Check that Hugo is installed: hugo version\n")
			fmt.Fprintf(os.Stderr, "  - Check hugo.toml for syntax errors\n")
			fmt.Fprintf(os.Stderr, "  - Run: hugo --verbose (for detailed output)\n")
			return fmt.Errorf("hugo build failed: %w", err)
		}
		if telemetry {
			buildMetrics.HugoDuration = time.Since(hugoStart).Milliseconds()
		}
		if !quiet {
			fmt.Printf("  %s Hugo build complete\n", icons.Check)
		}

		publishDirPath := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)

		noOptimize, err := cmd.Flags().GetBool("no-optimize")
		if err != nil {
			return fmt.Errorf("error reading no-optimize flag: %w", err)
		}

		noCompress, err := cmd.Flags().GetBool("no-compress")
		if err != nil {
			return fmt.Errorf("error reading no-compress flag: %w", err)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return fmt.Errorf("error reading verbose flag: %w", err)
		}

		var currentStep int
		if clean {
			currentStep = 3
		} else {
			currentStep = 2
		}

		if walgoCfg.OptimizerConfig.Enabled && !cmd.Flags().Changed("no-optimize") && !noOptimize {
			if !quiet {
				fmt.Printf("  [%d] Optimizing assets...\n", currentStep)
			}
			currentStep++

			optimizeStart := time.Now()
			optimizerEngine := optimizer.NewEngine(walgoCfg.OptimizerConfig)
			stats, err := optimizerEngine.OptimizeDirectory(publishDirPath)
			if telemetry {
				buildMetrics.OptimizeDuration = time.Since(optimizeStart).Milliseconds()
			}
			if err != nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "  %s Warning: Optimization failed: %v\n", icons.Warning, err)
				}
			} else if !quiet {
				optimizerEngine.PrintStats(stats)
				fmt.Printf("  %s Optimization complete\n", icons.Check)
			}
		}

		var compressionStats *compress.DirectoryCompressionStats
		if walgoCfg.CompressConfig.Enabled && !noCompress {
			if !quiet {
				fmt.Printf("  [%d] Compressing assets...\n", currentStep)
			}
			currentStep++

			compressConfig := compress.Config{
				Enabled:        true,
				BrotliLevel:    walgoCfg.CompressConfig.Level,
				GzipEnabled:    false,
				SkipExtensions: compress.DefaultConfig().SkipExtensions,
			}

			if compressConfig.BrotliLevel == 0 {
				compressConfig.BrotliLevel = 6 // Default level
			}

			compressStart := time.Now()
			compressor := compress.New(compressConfig)
			stats, err := compressor.CompressDirectory(publishDirPath)
			if telemetry {
				buildMetrics.CompressDuration = time.Since(compressStart).Milliseconds()
				if stats != nil {
					buildMetrics.TotalFiles = stats.Compressed + stats.NotWorthCompressing + stats.Skipped
					buildMetrics.CompressedFiles = stats.Compressed
					buildMetrics.CompressionSavings = int64(stats.TotalOriginalSize - stats.TotalCompressedSize)
				}
			}
			if err != nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "  %s Warning: Compression failed: %v\n", icons.Warning, err)
				}
			} else {
				compressionStats = stats
				if !quiet {
					if verbose {
						stats.PrintVerboseSummary()
					} else {
						stats.PrintSummary()
					}
					fmt.Printf("  %s Compression complete\n", icons.Check)
				}
			}
		}

		if walgoCfg.CompressConfig.GenerateWSResources {
			if !quiet {
				fmt.Printf("  [%d] Generating ws-resources.json...\n", currentStep)
			}

			cacheConfig := compress.CacheControlConfig{
				Enabled:         walgoCfg.CacheConfig.Enabled,
				ImmutableMaxAge: walgoCfg.CacheConfig.ImmutableMaxAge,
				MutableMaxAge:   walgoCfg.CacheConfig.MutableMaxAge,
			}

			if cacheConfig.ImmutableMaxAge == 0 {
				cacheConfig.ImmutableMaxAge = 31536000 // 1 year default
			}
			if cacheConfig.MutableMaxAge == 0 {
				cacheConfig.MutableMaxAge = 300 // 5 minutes default
			}

			if len(cacheConfig.ImmutablePatterns) == 0 {
				cacheConfig.ImmutablePatterns = compress.DefaultCacheControlConfig().ImmutablePatterns
			}
			wsOptions := compress.WSResourcesOptions{
				CompressionStats: compressionStats,
				CacheConfig:      cacheConfig,
				CustomRoutes:     walgoCfg.CompressConfig.CustomRoutes,
				CustomIgnore:     walgoCfg.CompressConfig.IgnorePatterns,
			}
			pm, err := projects.NewManager()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Failed to get project manager: %v\n", icons.Warning, err)
				return fmt.Errorf("failed to get project manager: %w", err)
			}
			defer pm.Close()
			project, err := pm.GetProjectBySitePath(sitePath)
			if err != nil || project == nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "%s Warning: Project not found for path: %s (using defaults)\n", icons.Warning, publishDirPath)
				}
				// Use default values if project not found
				wsOptions.SiteName = filepath.Base(sitePath)
				wsOptions.Description = ""
				wsOptions.ImageURL = ""
				wsOptions.Link = compress.DefaultLink
				wsOptions.ProjectURL = compress.DefaultProjectURL
				wsOptions.Creator = compress.DefaultCreator
				wsOptions.Category = ""
			} else {
				// Use project metadata if available
				wsOptions.SiteName = project.Name
				wsOptions.Description = project.Description
				wsOptions.ImageURL = project.ImageURL
				wsOptions.Link = compress.DefaultLink
				wsOptions.ProjectURL = compress.DefaultProjectURL
				wsOptions.Creator = compress.DefaultCreator
				wsOptions.Category = project.Category
			}

			wsConfig, err := compress.GenerateWSResourcesConfig(publishDirPath, wsOptions)
			if err != nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "  %s Warning: Failed to generate ws-resources.json: %v\n", icons.Warning, err)
				}
			} else {
				outputPath := filepath.Join(publishDirPath, "ws-resources.json")
				if err := compress.WriteWSResourcesConfig(wsConfig, outputPath); err != nil {
					if !quiet {
						fmt.Fprintf(os.Stderr, "  %s Warning: Failed to write ws-resources.json: %v\n", icons.Warning, err)
					}
				} else if !quiet {
					fmt.Printf("  %s Generated ws-resources.json (%d resources)\n", icons.Check, len(wsConfig.Headers))
				}
			}
		}

		routes, err := compress.GenerateRoutesFromPublic(publishDirPath)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "  %s Warning: Failed to generate routes: %v\n", icons.Warning, err)
			}
		} else {
			compress.MergeRoutesIntoWSResources(filepath.Join(publishDirPath, "ws-resources.json"), routes)
		}

		success = true

		if !quiet {
			fmt.Printf("\n%s Build complete! Output: %s\n", icons.Success, publishDirPath)
			fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
			fmt.Printf("  - Preview: walgo serve\n")
			fmt.Printf("  - Deploy:  walgo launch\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolP("clean", "c", false, "Clean the public directory before building")
	buildCmd.Flags().Bool("no-optimize", false, "Skip asset optimization after building")
	buildCmd.Flags().Bool("no-compress", false, "Skip Brotli compression after building")
	buildCmd.Flags().BoolP("verbose", "v", false, "Show detailed compression and optimization stats")
	buildCmd.Flags().BoolP("quiet", "q", false, "Suppress output (used internally by quickstart)")
	buildCmd.Flags().Bool("telemetry", false, "Record build metrics to local JSON file (~/.walgo/metrics.json)")
}
