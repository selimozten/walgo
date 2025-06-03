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
- Copies attachments to the static directory  
- Adds Hugo frontmatter to files that don't have it
- Preserves directory structure from the vault`,
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
		outputDir, _ := cmd.Flags().GetString("output-dir")
		overwrite, _ := cmd.Flags().GetBool("overwrite")
		convertWikilinks, _ := cmd.Flags().GetBool("convert-wikilinks")
		attachmentDir, _ := cmd.Flags().GetString("attachment-dir")
		frontmatterFormat, _ := cmd.Flags().GetString("frontmatter-format")

		// Determine target content directory
		hugoContentDir := filepath.Join(sitePath, cfg.HugoConfig.ContentDir)
		if outputDir != "" {
			hugoContentDir = filepath.Join(hugoContentDir, outputDir)
		}

		fmt.Printf("Target Hugo content directory: %s\n", hugoContentDir)

		// Check if target directory exists and has files (unless overwrite is enabled)
		if !overwrite {
			if files, err := os.ReadDir(hugoContentDir); err == nil && len(files) > 0 {
				fmt.Fprintf(os.Stderr, "Target directory %s is not empty. Use --overwrite to proceed anyway.\n", hugoContentDir)
				os.Exit(1)
			}
		}

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
}
