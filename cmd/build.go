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
		fmt.Println("Executing build command...")

		// Determine site path (current directory by default)
		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		// Load Walgo configuration to potentially get Hugo settings
		// like custom publishDir if we need to verify it later.
		walgoCfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		// Check if clean flag is set
		if clean, _ := cmd.Flags().GetBool("clean"); clean {
			publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
			fmt.Printf("Cleaning publish directory: %s\n", publishDir)
			if err := os.RemoveAll(publishDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to clean publish directory: %v\n", err)
			}
		}

		// Execute Hugo build
		if err := hugo.BuildSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error building Hugo site: %v\n", err)
			os.Exit(1)
		}

		publishDirPath := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
		fmt.Printf("Hugo build complete. Output in: %s\n", publishDirPath)

		// Run optimization if enabled and --no-optimize flag is not set
		if walgoCfg.OptimizerConfig.Enabled && !cmd.Flags().Changed("no-optimize") {
			if noOptimize, _ := cmd.Flags().GetBool("no-optimize"); !noOptimize {
				fmt.Println("Running asset optimization...")

				optimizerEngine := optimizer.NewEngine(walgoCfg.OptimizerConfig)
				stats, err := optimizerEngine.OptimizeDirectory(publishDirPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Optimization failed: %v\n", err)
				} else {
					optimizerEngine.PrintStats(stats)
				}
			}
		}

		fmt.Println("Site build complete with optimizations.")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolP("clean", "c", false, "Clean the public directory before building")
	buildCmd.Flags().Bool("no-optimize", false, "Skip asset optimization after building")
}
