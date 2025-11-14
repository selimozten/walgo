package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"walgo/internal/compress"
	"walgo/internal/config"
	"walgo/internal/hugo"
	"walgo/internal/metrics"
	"walgo/internal/optimizer"

	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Hugo site.",
	Long: `Builds the Hugo site using the configuration found in the current directory
(or the directory specified by global --config flag if walgo.yaml is there).
This command runs the 'hugo' command to generate static files typically into the 'public' directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		quiet, _ := cmd.Flags().GetBool("quiet")

		// Initialize telemetry if enabled
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

		// Determine site path (current directory by default)
		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: Cannot determine current directory: %v\n", err)
			os.Exit(1)
		}

		// Load Walgo configuration
		walgoCfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n💡 Did you run 'walgo init' to create a site?\n")
			os.Exit(1)
		}

		if !quiet {
			fmt.Println("🔨 Building site...")
		}

		// Check if clean flag is set
		clean, err := cmd.Flags().GetBool("clean")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading clean flag: %v\n", err)
			os.Exit(1)
		}
		if clean {
			publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
			if !quiet {
				fmt.Printf("  [1/3] Cleaning %s...\n", publishDir)
			}
			if err := os.RemoveAll(publishDir); err != nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "  ⚠ Warning: Failed to clean: %v\n", err)
				}
			} else if !quiet {
				fmt.Println("  ✓ Cleaned")
			}
		}

		// Execute Hugo build
		if !quiet {
			stepNum := 2
			if !clean {
				stepNum = 1
			}
			fmt.Printf("  [%d/3] Running Hugo build...\n", stepNum)
		}
		hugoStart := time.Now()
		if err := hugo.BuildSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: Hugo build failed: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "💡 Troubleshooting:\n")
			fmt.Fprintf(os.Stderr, "  - Check that Hugo is installed: hugo version\n")
			fmt.Fprintf(os.Stderr, "  - Check hugo.toml for syntax errors\n")
			fmt.Fprintf(os.Stderr, "  - Run: hugo --verbose (for detailed output)\n")
			os.Exit(1)
		}
		if telemetry {
			buildMetrics.HugoDuration = time.Since(hugoStart).Milliseconds()
		}
		if !quiet {
			fmt.Println("  ✓ Hugo build complete")
		}

		publishDirPath := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)

		// Get flags
		noOptimize, err := cmd.Flags().GetBool("no-optimize")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading no-optimize flag: %v\n", err)
			os.Exit(1)
		}

		noCompress, err := cmd.Flags().GetBool("no-compress")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading no-compress flag: %v\n", err)
			os.Exit(1)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading verbose flag: %v\n", err)
			os.Exit(1)
		}

		// Calculate step number
		var currentStep int
		if clean {
			currentStep = 3
		} else {
			currentStep = 2
		}

		// Run optimization if enabled and --no-optimize flag is not set
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
					fmt.Fprintf(os.Stderr, "  ⚠ Warning: Optimization failed: %v\n", err)
				}
			} else if !quiet {
				optimizerEngine.PrintStats(stats)
				fmt.Println("  ✓ Optimization complete")
			}
		}

		// Run compression if enabled and --no-compress flag is not set
		var compressionStats *compress.DirectoryCompressionStats
		if walgoCfg.CompressConfig.Enabled && !noCompress {
			if !quiet {
				fmt.Printf("  [%d] Compressing assets...\n", currentStep)
			}
			currentStep++

			// Configure compression
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
					fmt.Fprintf(os.Stderr, "  ⚠ Warning: Compression failed: %v\n", err)
				}
			} else {
				compressionStats = stats
				if !quiet {
					if verbose {
						stats.PrintVerboseSummary()
					} else {
						stats.PrintSummary()
					}
					fmt.Println("  ✓ Compression complete")
				}
			}
		}

		// Generate ws-resources.json if enabled
		if walgoCfg.CompressConfig.GenerateWSResources {
			if !quiet {
				fmt.Printf("  [%d] Generating ws-resources.json...\n", currentStep)
			}

			// Configure cache control
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

			// Add default immutable patterns if empty
			if len(cacheConfig.ImmutablePatterns) == 0 {
				cacheConfig.ImmutablePatterns = compress.DefaultCacheControlConfig().ImmutablePatterns
			}

			wsConfig, err := compress.GenerateWSResourcesConfig(publishDirPath, compressionStats, cacheConfig)
			if err != nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "  ⚠ Warning: Failed to generate ws-resources.json: %v\n", err)
				}
			} else {
				outputPath := filepath.Join(publishDirPath, "ws-resources.json")
				if err := compress.WriteWSResourcesConfig(wsConfig, outputPath); err != nil {
					if !quiet {
						fmt.Fprintf(os.Stderr, "  ⚠ Warning: Failed to write ws-resources.json: %v\n", err)
					}
				} else if !quiet {
					fmt.Printf("  ✓ Generated ws-resources.json (%d resources)\n", len(wsConfig.Headers))
				}
			}
		}

		// Mark build as successful
		success = true

		if !quiet {
			fmt.Printf("\n✅ Build complete! Output: %s\n", publishDirPath)
			fmt.Printf("\n💡 Next steps:\n")
			fmt.Printf("  - Preview: walgo serve\n")
			fmt.Printf("  - Deploy: walgo deploy --epochs 1\n")
		}
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
