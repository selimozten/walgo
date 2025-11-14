package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"walgo/internal/compress"
	"walgo/internal/config"

	"github.com/spf13/cobra"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress [directory]",
	Short: "Compress files in a directory using Brotli compression.",
	Long: `Compresses HTML, CSS, JavaScript, JSON, and SVG files in the specified directory
using Brotli compression. This is useful for manually compressing files or
re-compressing after changes.

By default, compresses the Hugo publish directory (usually 'public').
You can specify a different directory as an argument.

Example:
  walgo compress              # Compress public/ directory
  walgo compress ./dist       # Compress ./dist directory
  walgo compress --level 11   # Maximum compression`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine directory to compress
		var targetDir string
		if len(args) > 0 {
			targetDir = args[0]
		} else {
			// Use publish directory from config
			sitePath, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Error: Cannot determine current directory: %v\n", err)
				os.Exit(1)
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
				fmt.Fprintf(os.Stderr, "\n💡 Tip: Specify a directory explicitly: walgo compress ./public\n")
				os.Exit(1)
			}

			targetDir = filepath.Join(sitePath, cfg.HugoConfig.PublishDir)
		}

		// Verify directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "❌ Error: Directory not found: %s\n", targetDir)
			os.Exit(1)
		}

		// Get flags
		level, err := cmd.Flags().GetInt("level")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading level flag: %v\n", err)
			os.Exit(1)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading verbose flag: %v\n", err)
			os.Exit(1)
		}

		inPlace, err := cmd.Flags().GetBool("in-place")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading in-place flag: %v\n", err)
			os.Exit(1)
		}

		generateWS, err := cmd.Flags().GetBool("generate-ws-resources")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading generate-ws-resources flag: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("📦 Compressing files in: %s\n", targetDir)
		fmt.Printf("  Compression level: %d\n", level)
		if inPlace {
			fmt.Println("  Mode: In-place (replaces originals)")
		} else {
			fmt.Println("  Mode: Creates .br files")
		}
		fmt.Println()

		// Configure compression
		compressConfig := compress.Config{
			Enabled:        true,
			BrotliLevel:    level,
			GzipEnabled:    false,
			SkipExtensions: compress.DefaultConfig().SkipExtensions,
		}

		compressor := compress.New(compressConfig)

		// Compress directory
		var stats *compress.DirectoryCompressionStats
		if inPlace {
			stats, err = compressor.CompressInPlace(targetDir)
		} else {
			stats, err = compressor.CompressDirectory(targetDir)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: Compression failed: %v\n", err)
			os.Exit(1)
		}

		// Print stats
		if verbose {
			stats.PrintVerboseSummary()
		} else {
			stats.PrintSummary()
		}

		// Generate ws-resources.json if requested
		if generateWS {
			fmt.Println("\n📝 Generating ws-resources.json...")

			// Load config for cache settings
			cfg, err := config.LoadConfig()
			cacheConfig := compress.DefaultCacheControlConfig()
			if err == nil {
				cacheConfig.Enabled = cfg.CacheConfig.Enabled
				cacheConfig.ImmutableMaxAge = cfg.CacheConfig.ImmutableMaxAge
				cacheConfig.MutableMaxAge = cfg.CacheConfig.MutableMaxAge

				if cacheConfig.ImmutableMaxAge == 0 {
					cacheConfig.ImmutableMaxAge = 31536000
				}
				if cacheConfig.MutableMaxAge == 0 {
					cacheConfig.MutableMaxAge = 300
				}
			}

			wsConfig, err := compress.GenerateWSResourcesConfig(targetDir, stats, cacheConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to generate ws-resources.json: %v\n", err)
			} else {
				outputPath := filepath.Join(targetDir, "ws-resources.json")
				if err := compress.WriteWSResourcesConfig(wsConfig, outputPath); err != nil {
					fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to write ws-resources.json: %v\n", err)
				} else {
					fmt.Printf("✅ Generated ws-resources.json (%d resources)\n", len(wsConfig.Headers))
				}
			}
		}

		fmt.Printf("\n✅ Compression complete!\n")
		if !inPlace {
			fmt.Printf("\n💡 Tip: Compressed files have .br extension\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(compressCmd)

	compressCmd.Flags().IntP("level", "l", 6, "Brotli compression level (0-11, default: 6)")
	compressCmd.Flags().BoolP("verbose", "v", false, "Show detailed file-by-file statistics")
	compressCmd.Flags().Bool("in-place", false, "Replace original files with compressed versions (use with caution)")
	compressCmd.Flags().Bool("generate-ws-resources", false, "Generate ws-resources.json for Walrus Sites")
}
