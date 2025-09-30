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

		fmt.Println("🔨 Building site...")

		// Check if clean flag is set
		if clean, _ := cmd.Flags().GetBool("clean"); clean {
			publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
			fmt.Printf("  [1/3] Cleaning %s...\n", publishDir)
			if err := os.RemoveAll(publishDir); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ Warning: Failed to clean: %v\n", err)
			} else {
				fmt.Println("  ✓ Cleaned")
			}
		}

		// Execute Hugo build
		stepNum := 2
		if clean, _ := cmd.Flags().GetBool("clean"); !clean {
			stepNum = 1
		}
		fmt.Printf("  [%d/3] Running Hugo build...\n", stepNum)
		if err := hugo.BuildSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: Hugo build failed: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "💡 Troubleshooting:\n")
			fmt.Fprintf(os.Stderr, "  - Check that Hugo is installed: hugo version\n")
			fmt.Fprintf(os.Stderr, "  - Check hugo.toml for syntax errors\n")
			fmt.Fprintf(os.Stderr, "  - Run: hugo --verbose (for detailed output)\n")
			os.Exit(1)
		}
		fmt.Println("  ✓ Hugo build complete")

		publishDirPath := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)

		// Run optimization if enabled and --no-optimize flag is not set
		if walgoCfg.OptimizerConfig.Enabled && !cmd.Flags().Changed("no-optimize") {
			if noOptimize, _ := cmd.Flags().GetBool("no-optimize"); !noOptimize {
				fmt.Println("  [3/3] Optimizing assets...")

				optimizerEngine := optimizer.NewEngine(walgoCfg.OptimizerConfig)
				stats, err := optimizerEngine.OptimizeDirectory(publishDirPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  ⚠ Warning: Optimization failed: %v\n", err)
				} else {
					optimizerEngine.PrintStats(stats)
					fmt.Println("  ✓ Optimization complete")
				}
			}
		}

		fmt.Printf("\n✅ Build complete! Output: %s\n", publishDirPath)
		fmt.Printf("\n💡 Next steps:\n")
		fmt.Printf("  - Preview: walgo serve\n")
		fmt.Printf("  - Deploy: walgo deploy-http --publisher ... --aggregator ... --epochs 1\n")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolP("clean", "c", false, "Clean the public directory before building")
	buildCmd.Flags().Bool("no-optimize", false, "Skip asset optimization after building")
}
