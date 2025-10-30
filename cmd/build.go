package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"walgo/internal/config"
	"walgo/internal/hugo"
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
		if err := hugo.BuildSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: Hugo build failed: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "💡 Troubleshooting:\n")
			fmt.Fprintf(os.Stderr, "  - Check that Hugo is installed: hugo version\n")
			fmt.Fprintf(os.Stderr, "  - Check hugo.toml for syntax errors\n")
			fmt.Fprintf(os.Stderr, "  - Run: hugo --verbose (for detailed output)\n")
			os.Exit(1)
		}
		if !quiet {
			fmt.Println("  ✓ Hugo build complete")
		}

		publishDirPath := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)

		// Run optimization if enabled and --no-optimize flag is not set
		noOptimize, err := cmd.Flags().GetBool("no-optimize")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading no-optimize flag: %v\n", err)
			os.Exit(1)
		}
		if walgoCfg.OptimizerConfig.Enabled && !cmd.Flags().Changed("no-optimize") && !noOptimize {
			if !quiet {
				fmt.Println("  [3/3] Optimizing assets...")
			}

			optimizerEngine := optimizer.NewEngine(walgoCfg.OptimizerConfig)
			stats, err := optimizerEngine.OptimizeDirectory(publishDirPath)
			if err != nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "  ⚠ Warning: Optimization failed: %v\n", err)
				}
			} else if !quiet {
				optimizerEngine.PrintStats(stats)
				fmt.Println("  ✓ Optimization complete")
			}
		}

		if !quiet {
			fmt.Printf("\n✅ Build complete! Output: %s\n", publishDirPath)
			fmt.Printf("\n💡 Next steps:\n")
			fmt.Printf("  - Preview: walgo serve\n")
			fmt.Printf("  - Deploy: walgo deploy-http --publisher ... --aggregator ... --epochs 1\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolP("clean", "c", false, "Clean the public directory before building")
	buildCmd.Flags().Bool("no-optimize", false, "Skip asset optimization after building")
	buildCmd.Flags().BoolP("quiet", "q", false, "Suppress output (used internally by quickstart)")
}
