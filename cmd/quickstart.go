package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
1. Ask which site type to create
2. Initialize a new Hugo site with the chosen theme
3. Add sample content (blog) or copy theme example site (docs, biolink, whitepaper)
4. Build the site

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

		// Ask user to select site type
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Site type:")
		fmt.Println("  1) Biolink")
		fmt.Println("  2) Blog")
		fmt.Println("  3) Docs")
		fmt.Println("  4) Whitepaper")
		fmt.Print("Select [1]: ")
		siteTypeChoice, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading site type: %w", err)
		}
		siteTypeChoice = strings.TrimSpace(siteTypeChoice)
		if siteTypeChoice == "" {
			siteTypeChoice = "1"
		}

		var siteType hugo.SiteType
		switch siteTypeChoice {
		case "1":
			siteType = hugo.SiteTypeBiolink
		case "2":
			siteType = hugo.SiteTypeBlog
		case "3":
			siteType = hugo.SiteTypeDocs
		case "4":
			siteType = hugo.SiteTypeWhitepaper
		default:
			return fmt.Errorf("invalid site type: %s", siteTypeChoice)
		}
		fmt.Println()

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
				// Resolve symlinks before removal to avoid deleting unexpected targets
				realPath, err := filepath.EvalSymlinks(sitePath)
				if err != nil {
					return // Can't resolve path, skip cleanup to be safe
				}
				// Verify resolved path is still under the working directory
				cwd, err := os.Getwd()
				if err != nil {
					return
				}
				if !strings.HasPrefix(realPath, cwd+string(os.PathSeparator)) {
					return // Path escaped working directory, skip cleanup
				}
				os.RemoveAll(realPath)
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

		themeInfo := hugo.GetThemeInfo(siteType)

		if siteType == hugo.SiteTypeBlog {
			// Blog: use our embedded TOML template + archetypes
			if err := hugo.SetupSiteConfigWithName(sitePath, siteType, siteName); err != nil {
				fmt.Fprintf(os.Stderr, "        %s Warning: Could not set up config: %v\n", icons.Warning, err)
			} else {
				fmt.Printf("        %s Config set up (%s)\n", icons.Check, siteType)
			}

			if err := hugo.SetupArchetypes(sitePath, themeInfo.DirName); err != nil {
				fmt.Fprintf(os.Stderr, "        %s Warning: Could not set up archetypes: %v\n", icons.Warning, err)
			} else {
				fmt.Printf("        %s Archetypes set up\n", icons.Check)
			}
		}

		// Setup favicon (theme-aware placement)
		if err := hugo.SetupFaviconForTheme(sitePath, themeInfo.DirName); err != nil {
			fmt.Fprintf(os.Stderr, "        %s Warning: Could not set up favicon: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("        %s Favicon set up\n", icons.Check)
		}

		// Install theme
		fmt.Printf("        %s Installing theme %s...\n", icons.Spinner, themeInfo.Name)
		if err := hugo.InstallTheme(sitePath, siteType); err != nil {
			fmt.Fprintf(os.Stderr, "        %s Warning: Could not install theme: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("        %s Theme %s installed\n", icons.Check, themeInfo.Name)
		}

		// Docs theme overrides
		if siteType == hugo.SiteTypeDocs {
			if err := hugo.SetupDocsThemeOverrides(sitePath); err != nil {
				fmt.Fprintf(os.Stderr, "        %s Warning: Could not set up docs theme overrides: %v\n", icons.Warning, err)
			}
		}

		// [4/4] Create sample content
		if !skipBuild {
			fmt.Println("\n  [4/4] Creating sample content...")
		} else {
			fmt.Println("\n  Creating sample content...")
		}

		switch siteType {
		case hugo.SiteTypeBlog:
			// Blog: use inline quickstart content
			contentDir := filepath.Join(sitePath, "content")
			if err := os.MkdirAll(contentDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "        %s Warning: Could not create content directory: %v\n", icons.Warning, err)
			} else {
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

` + "✓" + ` **Permanent** - Once published, it's always accessible
` + "✓" + ` **Distributed** - No single point of failure
` + "✓" + ` **Verifiable** - All content is cryptographically verified
` + "✓" + ` **Fast** - Delivered through a global network

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

		default:
			// Docs, Biolink, Whitepaper: use exampleSite directly (including config)
			if err := hugo.CopyExampleSiteWithConfig(sitePath, siteType, siteName); err != nil {
				fmt.Fprintf(os.Stderr, "        %s Warning: Could not copy example site: %v\n", icons.Warning, err)
				// Create a minimal homepage as fallback
				contentDir := filepath.Join(sitePath, "content")
				if mkErr := os.MkdirAll(contentDir, 0755); mkErr == nil {
					indexPath := filepath.Join(contentDir, "_index.md")
					fallbackContent := "---\ntitle: \"" + siteName + "\"\ndraft: false\n---\n\n# " + siteName + "\n"
					if wErr := os.WriteFile(indexPath, []byte(fallbackContent), 0644); wErr != nil {
						fmt.Fprintf(os.Stderr, "        %s Warning: Could not create fallback homepage: %v\n", icons.Warning, wErr)
					}
				}
				fmt.Printf("        %s Created minimal homepage (fallback)\n", icons.Check)
			} else {
				fmt.Printf("        %s Example site content copied\n", icons.Check)
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
