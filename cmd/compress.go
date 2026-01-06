package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/selimozten/walgo/internal/compress"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/projects"

	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

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
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		var targetDir string
		if len(args) > 0 {
			targetDir = args[0]
		} else {
			sitePath, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
				return fmt.Errorf("error getting current directory: %w", err)
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
				fmt.Fprintf(os.Stderr, "\n%s Tip: Specify a directory explicitly: walgo compress ./public\n", icons.Lightbulb)
				return fmt.Errorf("error loading config: %w", err)
			}

			targetDir = filepath.Join(sitePath, cfg.HugoConfig.PublishDir)
		}

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s Error: Directory not found: %s\n", icons.Error, targetDir)
			return fmt.Errorf("directory not found: %s", targetDir)
		}

		level, err := cmd.Flags().GetInt("level")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading level flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading level flag: %w", err)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading verbose flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading verbose flag: %w", err)
		}

		inPlace, err := cmd.Flags().GetBool("in-place")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading in-place flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading in-place flag: %w", err)
		}

		generateWS, err := cmd.Flags().GetBool("generate-ws-resources")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading generate-ws-resources flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading generate-ws-resources flag: %w", err)
		}

		fmt.Printf("%s Compressing files in: %s\n", icons.Package, targetDir)
		fmt.Printf("  Compression level: %d\n", level)
		if inPlace {
			fmt.Println("  Mode: In-place (replaces originals)")
		} else {
			fmt.Println("  Mode: Creates .br files")
		}
		fmt.Println()

		compressConfig := compress.Config{
			Enabled:        true,
			BrotliLevel:    level,
			GzipEnabled:    false,
			SkipExtensions: compress.DefaultConfig().SkipExtensions,
		}

		compressor := compress.New(compressConfig)

		var stats *compress.DirectoryCompressionStats
		if inPlace {
			stats, err = compressor.CompressInPlace(targetDir)
		} else {
			stats, err = compressor.CompressDirectory(targetDir)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Compression failed: %v\n", icons.Error, err)
			return fmt.Errorf("compression failed: %w", err)
		}

		if verbose {
			stats.PrintVerboseSummary()
		} else {
			stats.PrintSummary()
		}

		if generateWS {
			fmt.Printf("\n%s Generating ws-resources.json...\n", icons.Pencil)

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

			customRoutes := make(map[string]string)
			var customIgnore []string
			if cfg != nil {
				customRoutes = cfg.CompressConfig.CustomRoutes
				customIgnore = cfg.CompressConfig.IgnorePatterns
			}

			wsOptions := compress.WSResourcesOptions{
				CompressionStats: stats,
				CacheConfig:      cacheConfig,
				CustomRoutes:     customRoutes,
				CustomIgnore:     customIgnore,
			}

			pm, err := projects.NewManager()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Failed to get project manager: %v\n", icons.Warning, err)
				return fmt.Errorf("failed to get project manager: %w", err)
			}
			defer pm.Close()
			sitePath, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
				return fmt.Errorf("error getting current directory: %w", err)
			}
			project, err := pm.GetProjectBySitePath(sitePath)
			if err != nil || project == nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Project not found for path: %s (using defaults)\n", icons.Warning, targetDir)
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

			wsConfig, err := compress.GenerateWSResourcesConfig(targetDir, wsOptions)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Failed to generate ws-resources.json: %v\n", icons.Warning, err)
			} else {
				outputPath := filepath.Join(targetDir, "ws-resources.json")
				if err := compress.WriteWSResourcesConfig(wsConfig, outputPath); err != nil {
					fmt.Fprintf(os.Stderr, "%s Warning: Failed to write ws-resources.json: %v\n", icons.Warning, err)
				} else {
					fmt.Printf("%s Generated ws-resources.json (%d resources)\n", icons.Success, len(wsConfig.Headers))
				}
			}
		}
		routes, err := compress.GenerateRoutesFromPublic(targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Failed to generate routes: %v\n", icons.Warning, err)
		} else {
			compress.MergeRoutesIntoWSResources(filepath.Join(targetDir, "ws-resources.json"), routes)
		}

		fmt.Printf("\n%s Compression complete!\n", icons.Success)
		if !inPlace {
			fmt.Printf("\n%s Tip: Compressed files have .br extension\n", icons.Lightbulb)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(compressCmd)

	compressCmd.Flags().IntP("level", "l", 6, "Brotli compression level (0-11, default: 6)")
	compressCmd.Flags().BoolP("verbose", "v", false, "Show detailed file-by-file statistics")
	compressCmd.Flags().Bool("in-place", false, "Replace original files with compressed versions (use with caution)")
	compressCmd.Flags().Bool("generate-ws-resources", false, "Generate ws-resources.json for Walrus Sites")
}
