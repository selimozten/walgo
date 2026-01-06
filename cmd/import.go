package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/obsidian"

	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

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
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		vaultPath := args[0]
		fmt.Printf("%s Importing content from Obsidian vault: %s\n", icons.Book, vaultPath)

		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("error loading config: %w", err)
		}

		outputDir, err := cmd.Flags().GetString("output-dir")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading output-dir flag: %v\n", err)
			return fmt.Errorf("error reading output-dir flag: %w", err)
		}
		overwrite, err := cmd.Flags().GetBool("overwrite")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading overwrite flag: %v\n", err)
			return fmt.Errorf("error reading overwrite flag: %w", err)
		}
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading dry-run flag: %v\n", err)
			return fmt.Errorf("error reading dry-run flag: %w", err)
		}
		convertWikilinks, err := cmd.Flags().GetBool("convert-wikilinks")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading convert-wikilinks flag: %v\n", err)
			return fmt.Errorf("error reading convert-wikilinks flag: %w", err)
		}
		attachmentDir, err := cmd.Flags().GetString("attachment-dir")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading attachment-dir flag: %v\n", err)
			return fmt.Errorf("error reading attachment-dir flag: %w", err)
		}
		frontmatterFormat, err := cmd.Flags().GetString("frontmatter-format")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading frontmatter-format flag: %v\n", err)
			return fmt.Errorf("error reading frontmatter-format flag: %w", err)
		}
		linkStyle, err := cmd.Flags().GetString("link-style")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading link-style flag: %v\n", err)
			return fmt.Errorf("error reading link-style flag: %w", err)
		}

		hugoContentDir := filepath.Join(sitePath, cfg.HugoConfig.ContentDir)
		if outputDir != "" {
			hugoContentDir = filepath.Join(hugoContentDir, outputDir)
		}

		fmt.Printf("%s Target Hugo content directory: %s\n", icons.Folder, hugoContentDir)

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

		if dryRun {
			fmt.Printf("\n%s Dry-run mode: Analyzing vault without importing...\n", icons.Info)
			stats, err := obsidian.DryRunImport(vaultPath, hugoContentDir, obsidianCfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Error: Failed to analyze vault: %v\n", icons.Error, err)
				return fmt.Errorf("failed to analyze vault: %w", err)
			}

			stats.PrintSummary()
			fmt.Printf("\n%s To actually import, run without --dry-run flag\n", icons.Lightbulb)
			return nil
		}

		if !overwrite {
			if files, err := os.ReadDir(hugoContentDir); err == nil && len(files) > 0 {
				fmt.Fprintf(os.Stderr, "%s Error: Target directory %s is not empty\n", icons.Error, hugoContentDir)
				fmt.Fprintf(os.Stderr, "\n%s Use --overwrite to proceed anyway\n", icons.Lightbulb)
				return fmt.Errorf("target directory is not empty: %s", hugoContentDir)
			}
		}

		fmt.Printf("\n%s Importing content...\n", icons.Package)
		stats, err := obsidian.ImportVault(vaultPath, hugoContentDir, obsidianCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s Error: Import failed: %v\n", icons.Error, err)
			return fmt.Errorf("import failed: %w", err)
		}

		fmt.Printf("\n%s Import completed successfully!\n", icons.Success)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%s Files processed: %d\n", icons.File, stats.FilesProcessed)
		if stats.FilesError > 0 {
			fmt.Printf("%s Files with errors: %d\n", icons.Warning, stats.FilesError)
		}
		fmt.Printf("%s Attachments copied: %d\n", icons.Package, stats.AttachmentsCopied)

		if stats.FilesError > 0 {
			fmt.Printf("\n%s Some files had errors during import. Check the output above for details.\n", icons.Warning)
		}

		fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
		fmt.Println("   - Review the imported content in your Hugo site")
		fmt.Println("   - Run 'walgo build' to build your site")
		fmt.Println("   - Run 'walgo serve' to preview your site locally")

		return nil
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
