package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

// quickstartCmd represents the quickstart command
var quickstartCmd = &cobra.Command{
	Use:   "quickstart <site-name>",
	Short: "Quick start: init, build, and deploy in one command",
	Long: `Creates a new Hugo site, adds sample content, builds it, and deploys to Walrus.

This command will:
1. Initialize a new Hugo site
2. Add sample content
3. Build the site
4. Deploy to Walrus HTTP testnet

Example:
  walgo quickstart my-blog
  walgo quickstart my-portfolio --skip-deploy`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]
		skipDeploy, _ := cmd.Flags().GetBool("skip-deploy")
		skipBuild, _ := cmd.Flags().GetBool("skip-build")

		// Validate site name to prevent command injection
		if !isValidSiteName(siteName) {
			fmt.Fprintf(os.Stderr, "❌ Invalid site name. Use only letters, numbers, hyphens and underscores\n")
			os.Exit(1)
		}

		fmt.Println("🚀 Walgo Quickstart")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Printf("Creating site: %s\n\n", siteName)

		// Check if Hugo is installed
		fmt.Println("[1/5] Checking dependencies...")
		if _, err := exec.LookPath("hugo"); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Hugo not found\n\n")
			fmt.Fprintf(os.Stderr, "Install Hugo first:\n")
			fmt.Fprintf(os.Stderr, "  macOS:  brew install hugo\n")
			fmt.Fprintf(os.Stderr, "  Linux:  sudo apt install hugo\n")
			fmt.Fprintf(os.Stderr, "  Or visit: https://gohugo.io/installation/\n")
			os.Exit(1)
		}
		fmt.Println("  ✓ Hugo found")

		// Step 2: Initialize site
		fmt.Println("\n[2/5] Creating Hugo site...")
		initCmd := exec.Command("walgo", "init", siteName) // #nosec G204 - siteName is validated by isValidSiteName() above
		initCmd.Stdout = os.Stdout
		initCmd.Stderr = os.Stderr
		if err := initCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Failed to create site: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("  ✓ Site created")

		// Change to site directory
		if err := os.Chdir(siteName); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Failed to enter site directory: %v\n", err)
			os.Exit(1)
		}

		// Step 3: Fetch and apply a theme
		fmt.Println("\n[3/5] Setting up theme...")
		const themeURL = "https://github.com/theNewDynamic/gohugo-theme-ananke.git"
		const themeName = "ananke"

		// Clone theme - hardcoded safe values
		cloneCmd := exec.Command("git", "clone", "--depth", "1", themeURL, filepath.Join("themes", themeName)) // #nosec G204 - hardcoded constants
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		if err := cloneCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ Warning: Failed to clone theme: %v\n", err)
			fmt.Fprintf(os.Stderr, "    You can add a theme manually later\n")
		} else {
			// Update hugo config to use theme
			fmt.Println("  ✓ Theme installed")
			// Write theme config safely
			configContent := fmt.Sprintf("\ntheme = \"%s\"\n", themeName)
			configPath := "hugo.toml"
			// #nosec G302 - config file needs to be writable
			f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ Warning: Could not open config: %v\n", err)
			} else {
				defer f.Close()
				if _, err := f.WriteString(configContent); err != nil {
					fmt.Fprintf(os.Stderr, "  ⚠ Warning: Could not update config: %v\n", err)
				}
			}
		}

		// Step 4: Create sample content
		fmt.Println("\n[4/5] Creating sample content...")
		contentDir := filepath.Join("content", "posts")
		if err := os.MkdirAll(contentDir, 0750); err != nil { // #nosec G301
			fmt.Fprintf(os.Stderr, "  ⚠ Warning: Could not create content directory: %v\n", err)
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
4. Deploy updates with ` + "`walgo deploy`" + `

Happy building! 🚀
`
			// #nosec G306 - content files need to be readable
			if err := os.WriteFile(welcomePath, []byte(content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ Warning: Could not create welcome post: %v\n", err)
			} else {
				fmt.Println("  ✓ Sample content created")
			}
		}

		if !skipBuild {
			// Step 5: Build the site
			fmt.Println("\n[5/5] Building site...")
			buildCmd := exec.Command("walgo", "build")
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr
			if err := buildCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "\n❌ Build failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("  ✓ Site built")
		}

		if !skipDeploy {
			// Step 6: Deploy
			fmt.Println("\n🌐 Deploying to Walrus (HTTP mode)...")
			deployCmd := exec.Command("walgo", "deploy-http")
			deployCmd.Stdout = os.Stdout
			deployCmd.Stderr = os.Stderr
			if err := deployCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "\n❌ Deploy failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "\nYou can try deploying manually with:\n")
				fmt.Fprintf(os.Stderr, "  cd %s && walgo deploy-http\n", siteName)
				os.Exit(1)
			}
		}

		// Success message
		fmt.Println("\n═══════════════════════════════════════════════════════════")
		fmt.Println("✨ Success! Your site is ready.")
		fmt.Printf("\n📁 Site location: %s/\n", siteName)
		if !skipBuild {
			fmt.Println("📦 Built files: public/")
		}
		if !skipDeploy {
			fmt.Println("🌐 Your site is now live on Walrus!")
		}
		fmt.Println("\n🎯 Next steps:")
		fmt.Println("  1. Edit content in content/posts/")
		fmt.Println("  2. Rebuild: walgo build")
		if skipDeploy {
			fmt.Println("  3. Deploy: walgo deploy-http")
		} else {
			fmt.Println("  3. Update: walgo update <object-id>")
		}
		fmt.Println("\n📖 Learn more: https://github.com/selimozten/walgo")
	},
}

// isValidSiteName validates that site name only contains safe characters
func isValidSiteName(name string) bool {
	// Only allow alphanumeric, hyphens, and underscores
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validName.MatchString(name) && len(name) > 0 && len(name) < 100
}

func init() {
	rootCmd.AddCommand(quickstartCmd)

	quickstartCmd.Flags().Bool("skip-deploy", false, "Skip deployment step")
	quickstartCmd.Flags().Bool("skip-build", false, "Skip build step")
}
