package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/selimozten/walgo/internal/utils"
	"github.com/spf13/cobra"
)

var quickstartCmd = &cobra.Command{
	Use:   "quickstart <site-name>",
	Short: "Quick start: init and build in one command",
	Long: `Creates a new Hugo site, adds sample content, and builds it.

This command will:
1. Initialize a new Hugo site
2. Add sample content
3. Build the site

Example:
  walgo quickstart my-blog
  walgo quickstart my-portfolio --skip-build`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		siteName := args[0]
		skipBuild, err := cmd.Flags().GetBool("skip-build")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading skip-build flag: %v\n", err)
			return fmt.Errorf("failed to read skip-build flag: %w", err)
		}

		if !utils.IsValidSiteName(siteName) {
			fmt.Fprintf(os.Stderr, "%s Error: Invalid site name\n", icons.Error)
			fmt.Fprintf(os.Stderr, "\n%s Use only letters, numbers, hyphens and underscores\n", icons.Lightbulb)
			return fmt.Errorf("invalid site name: %s", siteName)
		}

		fmt.Printf("%s Walgo Quick Start\n", icons.Rocket)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%s Creating site: %s\n\n", icons.Package, siteName)

		// [1/4] Check dependencies
		fmt.Println("  [1/4] Checking dependencies...")
		if _, err := deps.LookPath("hugo"); err != nil {
			fmt.Fprintf(os.Stderr, "\n%s Error: Hugo not found\n", icons.Error)
			fmt.Fprintf(os.Stderr, "\n%s Install from https://gohugo.io/installation/\n", icons.Lightbulb)
			return fmt.Errorf("hugo not found: %w", err)
		}
		fmt.Printf("        %s Hugo found\n", icons.Check)

		// [2/4] Create site directory and initialize Hugo
		fmt.Println("\n  [2/4] Creating Hugo site...")

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}

		sitePath = filepath.Join(sitePath, siteName)
		// Create site directory
		if err := os.MkdirAll(sitePath, 0755); err != nil {
			return fmt.Errorf("failed to create site directory: %w", err)
		}

		// Initialize Hugo site
		if err := hugo.InitializeSite(sitePath); err != nil {
			return fmt.Errorf("failed to initialize Hugo site: %w", err)
		}
		fmt.Printf("        %s Hugo site initialized\n", icons.Check)

		// Create walgo.yaml config
		if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "        %s Warning: Could not create walgo.yaml: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("        %s Walgo config created\n", icons.Check)
		}

		// [3/4] Set up site (config, archetypes, theme)
		fmt.Println("\n  [3/4] Setting up site...")

		// Blog is the default site type - use Ananke theme
		siteType := hugo.SiteTypeBlog
		themeInfo := hugo.GetThemeInfo(siteType)

		// Setup hugo.toml with blog configuration and site name
		if err := hugo.SetupSiteConfigWithName(sitePath, siteType, siteName); err != nil {
			fmt.Fprintf(os.Stderr, "        %s Warning: Could not set up config: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("        %s Config set up (blog)\n", icons.Check)
		}

		// Setup archetypes
		if err := hugo.SetupArchetypes(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "        %s Warning: Could not set up archetypes: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("        %s Archetypes set up\n", icons.Check)
		}

		// Setup favicon
		if err := hugo.SetupFavicon(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "        %s Warning: Could not set up favicon: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("        %s Favicon set up\n", icons.Check)
		}

		// Install theme (Ananke for blog)
		fmt.Printf("        %s Installing theme %s...\n", icons.Spinner, themeInfo.Name)
		if err := hugo.InstallTheme(sitePath, siteType); err != nil {
			fmt.Fprintf(os.Stderr, "        %s Warning: Could not install theme: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("        %s Theme %s installed\n", icons.Check, themeInfo.Name)
		}

		// Setup theme-specific overrides (e.g., favicon fix for business theme)
		if siteType == hugo.SiteTypeBusiness {
			if err := hugo.SetupBusinessThemeOverrides(sitePath); err != nil {
				fmt.Fprintf(os.Stderr, "        %s Warning: Could not set up theme overrides: %v\n", icons.Warning, err)
			} else {
				fmt.Printf("        %s Theme overrides set up\n", icons.Check)
			}
		}

		// [4/4] Create sample content
		if !skipBuild {
			fmt.Println("\n  [4/4] Creating sample content...")
		} else {
			fmt.Println("\n  Creating sample content...")
		}

		contentDir := filepath.Join(sitePath, "content", "posts")
		if err := os.MkdirAll(contentDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "        %s Warning: Could not create content directory: %v\n", icons.Warning, err)
		} else {
			welcomePath := filepath.Join(contentDir, "welcome.md")
			content := `---
title: "Welcome to Walrus Sites"
date: 2024-01-01T00:00:00Z
draft: false
---

Welcome to your new decentralized website powered by Walrus!

This site is hosted on the Walrus decentralized storage network, making it censorship-resistant and always available.

## Next Steps

1. Edit this content in ` + "`content/posts/welcome.md`" + `
2. Add more posts to ` + "`content/posts/`" + `
3. Customize your theme
4. Deploy with ` + "`walgo launch`" + `

Happy building! ` + icons.Rocket + `
`
			if err := os.WriteFile(welcomePath, []byte(content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "        %s Warning: Could not create welcome post: %v\n", icons.Warning, err)
			} else {
				fmt.Printf("        %s Sample content created\n", icons.Check)
			}
		}

		// Build site
		if !skipBuild {
			fmt.Println("\n  Building site...")
			if err := hugo.BuildSite(sitePath); err != nil {
				fmt.Fprintf(os.Stderr, "\n%s Error: Build failed: %v\n", icons.Error, err)
				return fmt.Errorf("failed to build site: %w", err)
			}
			fmt.Printf("        %s Site built\n", icons.Check)
		}

		manager, err := projects.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create project manager: %w", err)
		}
		defer manager.Close()

		// Create draft project
		if err := manager.CreateDraftProject(siteName, sitePath); err != nil {
			return fmt.Errorf("failed to create draft project: %w", err)
		}
		fmt.Printf("   %s Created draft project\n", icons.Check)

		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%s Quick start complete!\n", icons.Success)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%s Site directory: %s/\n", icons.Folder, siteName)
		if !skipBuild {
			fmt.Printf("%s Built files: public/\n", icons.Package)
		}
		fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
		fmt.Println("   - cd " + siteName)
		fmt.Println("   - walgo serve      # Preview your site locally")
		fmt.Println("   - walgo launch     # Deploy with interactive wizard")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(quickstartCmd)

	quickstartCmd.Flags().Bool("skip-build", false, "Skip build step")
}
