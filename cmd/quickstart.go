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

		// Check if directory already exists before creating
		dirExistedBefore := false
		if _, err := os.Stat(sitePath); err == nil {
			dirExistedBefore = true
		}

		// Create site directory
		if err := os.MkdirAll(sitePath, 0755); err != nil {
			return fmt.Errorf("failed to create site directory: %w", err)
		}

		// Setup cleanup on failure
		success := false
		defer func() {
			if !success && !dirExistedBefore {
				// Clean up the directory if we created it and operation failed
				os.RemoveAll(sitePath)
			}
		}()

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

		contentDir := filepath.Join(sitePath, "content")
		if err := os.MkdirAll(contentDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "        %s Warning: Could not create content directory: %v\n", icons.Warning, err)
		} else {
			// Create homepage with detailed content
			indexPath := filepath.Join(contentDir, "_index.md")
			indexContent := `---
title: "` + siteName + `"
date: 2024-01-01T00:00:00Z
draft: false
---

# Welcome to ` + siteName + `

Your decentralized website powered by **Walrus** - the cutting-edge decentralized storage network built on the Sui blockchain.

## What is Walrus?

Walrus is a decentralized storage and data availability protocol designed for large binary files (blobs). Built on Sui, it provides:

- **Censorship Resistance**: Your content cannot be taken down or restricted
- **High Availability**: Data is distributed across multiple storage nodes
- **Cost Effective**: Optimized storage with efficient encoding
- **Fast Access**: Quick retrieval through CDN-like distribution

## About This Site

This site is hosted entirely on the Walrus network, making it:

✓ **Permanent** - Once published, it's always accessible
✓ **Distributed** - No single point of failure
✓ **Verifiable** - All content is cryptographically verified
✓ **Fast** - Delivered through a global network

## Getting Started

### Edit Your Content

Your site uses Hugo, a fast static site generator. All content is in Markdown format:

- Edit this page: ` + "`content/_index.md`" + `
- Add new pages to the ` + "`content/`" + ` directory
- Organize with subdirectories for complex sites

### Preview Locally

Test your changes before deploying:

` + "```bash" + `
walgo serve
` + "```" + `

This starts a local server at ` + "`http://localhost:1313`" + `

### Deploy to Walrus

When you're ready to publish:

` + "```bash" + `
walgo launch
` + "```" + `

Follow the interactive wizard to:
1. Configure your deployment
2. Select network (testnet/mainnet)
3. Set storage epochs
4. Publish to Walrus

## Next Steps

1. **Customize Your Theme**: Edit ` + "`hugo.toml`" + ` to change colors, fonts, and layout
2. **Add More Content**: Create new pages and blog posts
3. **Explore Hugo**: Learn more at [gohugo.io](https://gohugo.io)
4. **Join the Community**: Connect with other Walrus builders

## Resources

- **Walrus Documentation**: [docs.walrus.site](https://docs.walrus.site)
- **Walgo CLI**: [github.com/selimozten/walgo](https://github.com/selimozten/walgo)
- **Hugo Docs**: [gohugo.io/documentation](https://gohugo.io/documentation)
- **Sui Network**: [sui.io](https://sui.io)

---

**Ready to build the decentralized web?** Start editing this file and make it your own! ` + icons.Rocket + `
`
			if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "        %s Warning: Could not create homepage: %v\n", icons.Warning, err)
			} else {
				fmt.Printf("        %s Homepage created\n", icons.Check)
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

		success = true
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
