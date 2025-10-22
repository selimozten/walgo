package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"walgo/internal/config"
	"walgo/internal/optimizer"

	"github.com/spf13/cobra"
)

// optimizeCmd represents the optimize command
var optimizeCmd = &cobra.Command{
	Use:   "optimize [directory]",
	Short: "Optimize HTML, CSS, and JavaScript files for better performance.",
	Long: `Optimizes HTML, CSS, and JavaScript files in the specified directory (or current directory).
This command provides asset optimization including:

HTML Optimization:
• Minify HTML by removing unnecessary whitespace
• Remove HTML comments (preserving conditional comments)  
• Compress inline CSS and JavaScript

CSS Optimization:
• Minify CSS by removing whitespace and comments
• Compress color values (e.g., #ffffff -> #fff)
• Remove unused CSS rules (when enabled)
• Remove unnecessary quotes and semicolons

JavaScript Optimization:
• Minify JavaScript by removing whitespace and comments
• Basic variable name obfuscation (when enabled)
• Preserve string contents and regular expressions

The optimization settings can be configured in walgo.yaml under the 'optimizer' section.`,
	Args: cobra.MaximumNArgs(1), // Optional directory argument
	Run: func(cmd *cobra.Command, args []string) {
		// Determine target directory
		var targetDir string
		if len(args) > 0 {
			targetDir = args[0]
		} else {
			// Use current directory or Hugo publish directory if available
			sitePath, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
				os.Exit(1)
			}

			// Try to load config to get publish directory
			if cfg, err := config.LoadConfig(); err == nil {
				publishDir := filepath.Join(sitePath, cfg.HugoConfig.PublishDir)
				if _, err := os.Stat(publishDir); err == nil {
					targetDir = publishDir
					fmt.Printf("Using Hugo publish directory: %s\n", publishDir)
				} else {
					targetDir = sitePath
				}
			} else {
				targetDir = sitePath
			}
		}

		// Verify target directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Directory %s does not exist\n", targetDir)
			os.Exit(1)
		}

		fmt.Printf("Optimizing files in: %s\n", targetDir)

		// Load configuration
		var optimizerConfig optimizer.OptimizerConfig
		if cfg, err := config.LoadConfig(); err == nil {
			optimizerConfig = cfg.OptimizerConfig
		} else {
			// Use default configuration if walgo.yaml is not available
			optimizerConfig = optimizer.NewDefaultOptimizerConfig()
			fmt.Println("Using default optimization settings (no walgo.yaml found)")
		}

		// Override config with command line flags
		if cmd.Flags().Changed("verbose") {
			verbose, _ := cmd.Flags().GetBool("verbose")
			optimizerConfig.Verbose = verbose
		}

		if cmd.Flags().Changed("html") {
			html, _ := cmd.Flags().GetBool("html")
			optimizerConfig.HTML.Enabled = html
		}

		if cmd.Flags().Changed("css") {
			css, _ := cmd.Flags().GetBool("css")
			optimizerConfig.CSS.Enabled = css
		}

		if cmd.Flags().Changed("js") {
			js, _ := cmd.Flags().GetBool("js")
			optimizerConfig.JS.Enabled = js
		}

		if cmd.Flags().Changed("remove-unused-css") {
			removeUnused, _ := cmd.Flags().GetBool("remove-unused-css")
			optimizerConfig.CSS.RemoveUnused = removeUnused
		}

		// Create and run optimizer
		engine := optimizer.NewEngine(optimizerConfig)
		stats, err := engine.OptimizeDirectory(targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during optimization: %v\n", err)
			os.Exit(1)
		}

		// Print results
		engine.PrintStats(stats)

		if stats.FilesOptimized > 0 {
			fmt.Printf("\n✅ Optimization complete! %d files optimized.\n", stats.FilesOptimized)
		} else {
			fmt.Println("\n✅ No files needed optimization.")
		}
	},
}

func init() {
	rootCmd.AddCommand(optimizeCmd)

	optimizeCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	optimizeCmd.Flags().Bool("html", true, "Enable HTML optimization")
	optimizeCmd.Flags().Bool("css", true, "Enable CSS optimization")
	optimizeCmd.Flags().Bool("js", true, "Enable JavaScript optimization")
	optimizeCmd.Flags().Bool("remove-unused-css", false, "Remove unused CSS rules (aggressive)")
}
