package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"walgo/internal/config"
	"walgo/internal/obsidian"

	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [obsidian-vault-path]",
	Short: "Import content from an Obsidian vault.",
	Long: `Imports content from a specified Obsidian vault path into the Hugo site structure.
This command will attempt to convert Obsidian markdown and attachments into a Hugo-compatible format.

Features:
- Converts [[wikilinks]] to Hugo markdown links
- Supports transclusions ![[note]] and ![[note#heading]]
- Copies attachments to the static directory
- Adds Hugo frontmatter to files that don't have it
- Preserves directory structure from the vault
- Enhanced alias and heading support`,
	Args: cobra.ExactArgs(1), // Expects exactly one argument: obsidian-vault-path
	Run: func(cmd *cobra.Command, args []string) {
		vaultPath := args[0]
		fmt.Printf("Importing content from Obsidian vault: %s\n", vaultPath)

		// Get current directory to load config and determine target content path
		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		// Get flag values
		outputDir, err := cmd.Flags().GetString("output-dir")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading output-dir flag: %v\n", err)
			os.Exit(1)
		}
		overwrite, err := cmd.Flags().GetBool("overwrite")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading overwrite flag: %v\n", err)
			os.Exit(1)
		}
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading dry-run flag: %v\n", err)
			os.Exit(1)
		}
		convertWikilinks, err := cmd.Flags().GetBool("convert-wikilinks")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading convert-wikilinks flag: %v\n", err)
			os.Exit(1)
		}
		attachmentDir, err := cmd.Flags().GetString("attachment-dir")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading attachment-dir flag: %v\n", err)
			os.Exit(1)
		}
		frontmatterFormat, err := cmd.Flags().GetString("frontmatter-format")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading frontmatter-format flag: %v\n", err)
			os.Exit(1)
		}
		linkStyle, err := cmd.Flags().GetString("link-style")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading link-style flag: %v\n", err)
			os.Exit(1)
		}

		// Determine target content directory
		hugoContentDir := filepath.Join(sitePath, cfg.HugoConfig.ContentDir)
		if outputDir != "" {
			hugoContentDir = filepath.Join(hugoContentDir, outputDir)
		}

		fmt.Printf("Target Hugo content directory: %s\n", hugoContentDir)

		// Configure Obsidian import settings
		obsidianCfg := cfg.ObsidianConfig
		if attachmentDir != "" {
			obsidianCfg.AttachmentDir = attachmentDir
		}
		if cmd.Flags().Changed("convert-wikilinks") {
			obsidianCfg.ConvertWikilinks = convertWikilinks
		}
		if frontmatterFormat != "" {
			obsidianCfg.FrontmatterFormat = frontmatterFormat
		}
		if linkStyle != "" {
			obsidianCfg.LinkStyle = linkStyle
		}

		// If dry-run, just analyze and exit
		if dryRun {
			fmt.Println("\n🔍 Dry-run mode: Analyzing vault without importing...")
			stats, err := obsidian.DryRunImport(vaultPath, hugoContentDir, obsidianCfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error analyzing vault: %v\n", err)
				os.Exit(1)
			}

			stats.PrintSummary()
			fmt.Println("\n💡 To actually import, run without --dry-run flag")
			os.Exit(0)
		}

		// Check if target directory exists and has files (unless overwrite is enabled)
		if !overwrite {
			if files, err := os.ReadDir(hugoContentDir); err == nil && len(files) > 0 {
				fmt.Fprintf(os.Stderr, "Target directory %s is not empty. Use --overwrite to proceed anyway.\n", hugoContentDir)
				os.Exit(1)
			}
		}

		// Perform the import
		stats, err := obsidian.ImportVault(vaultPath, hugoContentDir, obsidianCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error importing Obsidian vault: %v\n", err)
			os.Exit(1)
		}

		// Display results
		fmt.Println("\nImport completed successfully!")
		fmt.Printf("Files processed: %d\n", stats.FilesProcessed)
		fmt.Printf("Files with errors: %d\n", stats.FilesError)
		fmt.Printf("Attachments copied: %d\n", stats.AttachmentsCopied)

		if stats.FilesError > 0 {
			fmt.Println("\nSome files had errors during import. Check the output above for details.")
		}

		fmt.Println("\nNext steps:")
		fmt.Println("- Review the imported content in your Hugo site")
		fmt.Println("- Run 'walgo build' to build your site")
		fmt.Println("- Run 'walgo serve' to preview your site locally")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("output-dir", "o", "", "Specify a subdirectory in content for imported files")
	importCmd.Flags().BoolP("overwrite", "f", false, "Overwrite existing files during import")
	importCmd.Flags().Bool("convert-wikilinks", true, "Convert [[wikilinks]] to Hugo markdown links")
	importCmd.Flags().String("attachment-dir", "", "Directory name for attachments (relative to static/)")
	importCmd.Flags().String("frontmatter-format", "", "Frontmatter format for new files (yaml, toml, json)")
	importCmd.Flags().String("link-style", "markdown", "Link conversion style: 'markdown' (default, avoids REF_NOT_FOUND) or 'relref' (strict Hugo shortcodes)")
	importCmd.Flags().Bool("dry-run", false, "Preview import without actually copying files")
}
