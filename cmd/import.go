package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/obsidian"
	"github.com/selimozten/walgo/internal/projects"

	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [obsidian-vault-path]",
	Short: "Import content from an Obsidian vault into a new Hugo site.",
	Long: `Imports content from a specified Obsidian vault path into a new Hugo site.
This command will:
1. Create a new Hugo site
2. Add walgo.yaml configuration
3. Import and convert Obsidian markdown and attachments

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

		// Validate and sanitize vault path to prevent path traversal attacks
		absVaultPath, err := filepath.Abs(vaultPath)
		if err != nil {
			return fmt.Errorf("invalid vault path: %w", err)
		}
		vaultPath = filepath.Clean(absVaultPath)

		// Verify the path exists and is a directory
		info, err := os.Stat(vaultPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("vault path does not exist: %s", vaultPath)
			}
			return fmt.Errorf("cannot access vault path: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("vault path is not a directory: %s", vaultPath)
		}

		// Get site name from flag or use vault directory name
		siteName, err := cmd.Flags().GetString("site-name")
		if err != nil {
			return fmt.Errorf("error reading site-name flag: %w", err)
		}
		siteName = strings.TrimSpace(siteName)
		if siteName == "" {
			// Use vault directory name as default
			siteName = filepath.Base(vaultPath)
		}

		// Get parent directory for site creation
		parentDir, err := cmd.Flags().GetString("parent-dir")
		if err != nil {
			return fmt.Errorf("error reading parent-dir flag: %w", err)
		}
		if parentDir == "" {
			parentDir, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot determine current directory: %w", err)
			}
		}

		sitePath := filepath.Join(parentDir, siteName)

		// Check if site already exists
		if _, err := os.Stat(sitePath); err == nil {
			return fmt.Errorf("site directory already exists: %s", sitePath)
		}

		fmt.Printf("%s Creating new Hugo site and importing Obsidian vault\n", icons.Rocket)
		fmt.Printf("   Vault: %s\n", vaultPath)
		fmt.Printf("   Site:  %s\n\n", sitePath)

		// Step 1: Create site directory
		if err := os.MkdirAll(sitePath, 0755); err != nil {
			return fmt.Errorf("failed to create site directory: %w", err)
		}
		fmt.Printf("%s Created site directory\n", icons.Check)

		// Step 2: Initialize Hugo site
		if err := hugo.InitializeSite(sitePath); err != nil {
			return fmt.Errorf("failed to initialize Hugo site: %w", err)
		}
		fmt.Printf("%s Initialized Hugo site\n", icons.Check)

		// Step 3: Create walgo.yaml
		if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
			return fmt.Errorf("failed to create walgo.yaml: %w", err)
		}
		fmt.Printf("%s Created walgo.yaml configuration\n", icons.Check)

		// Step 4: Import Obsidian vault
		fmt.Printf("\n%s Importing Obsidian content...\n", icons.Book)

		// Load config from the new site
		cfg, err := config.LoadConfigFrom(sitePath)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		outputDir, err := cmd.Flags().GetString("output-dir")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading output-dir flag: %v\n", err)
			return fmt.Errorf("error reading output-dir flag: %w", err)
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
			// Clean up the created site directory in dry-run mode
			os.RemoveAll(sitePath)
			return nil
		}

		// No need for overwrite check - we just created the site
		fmt.Printf("%s Importing content...\n", icons.Package)
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

		err = hugo.BuildSite(sitePath)
		if err != nil {
			return fmt.Errorf("failed to build site: %w", err)
		}

		// Step 5: Create draft project
		manager, err := projects.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create project manager: %w", err)
		}
		defer manager.Close()
		if err := manager.CreateDraftProject(siteName, sitePath); err != nil {
			return fmt.Errorf("failed to create draft project: %w", err)
		}
		fmt.Printf("\n%s Created draft project: %s\n", icons.Check, siteName)

		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%s Site created and imported successfully!\n", icons.Success)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
		fmt.Printf("   - cd %s\n", siteName)
		fmt.Println("   - Build site: walgo build")
		fmt.Println("   - Preview: walgo serve")
		fmt.Println("   - Deploy: walgo launch")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("site-name", "n", "", "Name for the new site (defaults to vault directory name)")
	importCmd.Flags().StringP("parent-dir", "p", "", "Parent directory for site creation (defaults to current directory)")
	importCmd.Flags().StringP("output-dir", "o", "", "Specify a subdirectory in content for imported files")
	importCmd.Flags().Bool("convert-wikilinks", true, "Convert [[wikilinks]] to Hugo markdown links")
	importCmd.Flags().String("attachment-dir", "", "Directory name for attachments (relative to static/)")
	importCmd.Flags().String("frontmatter-format", "", "Frontmatter format for new files (yaml, toml, json)")
	importCmd.Flags().String("link-style", "markdown", "Link conversion style: 'markdown' (default, avoids REF_NOT_FOUND) or 'relref' (strict Hugo shortcodes)")
	importCmd.Flags().Bool("dry-run", false, "Preview import without actually copying files")
}
