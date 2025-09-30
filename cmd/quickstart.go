package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var quickstartCmd = &cobra.Command{
	Use:   "quickstart <site-name>",
	Short: "Create and deploy a site in one command (HTTP testnet, no wallet needed)",
	Long: `Quickstart creates a new Hugo site, builds it, and deploys to Walrus HTTP testnet.
This is the fastest way to get started - no wallet or on-chain setup required.

The command will:
1. Create a new Hugo site
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
		initCmd := exec.Command("walgo", "init", siteName)
		initCmd.Stdout = os.Stdout
		initCmd.Stderr = os.Stderr
		if err := initCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Failed to create site: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("  ✓ Site created")

		// Change to site directory
		siteDir := siteName
		if err := os.Chdir(siteDir); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to enter site directory: %v\n", err)
			os.Exit(1)
		}

		// Step 3: Create sample content
		fmt.Println("\n[3/5] Creating sample content...")

		// Create a welcome post
		welcomeContent := `---
title: "Welcome to Walrus Sites"
date: %s
draft: false
---

# Welcome!

This is your first post on Walrus decentralized storage.

## What is Walrus?

Walrus is a decentralized storage network that provides:
- **Permanent storage**: Your content persists without relying on centralized servers
- **Censorship resistance**: No single entity can take down your site
- **Fast delivery**: Content is distributed across the network

## Next Steps

1. Edit this post: ` + "`content/posts/welcome.md`" + `
2. Add more content: ` + "`walgo new posts/my-post.md`" + `
3. Customize your site: Edit ` + "`hugo.toml`" + `
4. Rebuild and deploy: ` + "`walgo build && walgo deploy-http ...`" + `

Happy publishing! 🚀
`
		// Create content directory
		contentDir := filepath.Join("content", "posts")
		if err := os.MkdirAll(contentDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ Warning: Could not create content directory: %v\n", err)
		} else {
			welcomePath := filepath.Join(contentDir, "welcome.md")
			date := exec.Command("date", "+%Y-%m-%d")
			dateOut, _ := date.Output()
			content := fmt.Sprintf(welcomeContent, string(dateOut))

			if err := os.WriteFile(welcomePath, []byte(content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ Warning: Could not create welcome post: %v\n", err)
			} else {
				fmt.Println("  ✓ Sample content created")
			}
		}

		if skipBuild {
			fmt.Println("\n✅ Site created successfully!")
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  cd %s\n", siteName)
			fmt.Printf("  walgo build\n")
			fmt.Printf("  walgo serve  # Preview locally\n")
			return
		}

		// Step 4: Build site
		fmt.Println("\n[4/5] Building site...")
		buildCmd := exec.Command("walgo", "build")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Build failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "\nYou can manually build later with: walgo build\n")
			os.Exit(1)
		}

		if skipDeploy {
			fmt.Println("\n✅ Site created and built successfully!")
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  cd %s\n", siteName)
			fmt.Printf("  walgo serve  # Preview locally\n")
			fmt.Printf("  walgo deploy-http --publisher https://publisher.walrus-testnet.walrus.space \\\n")
			fmt.Printf("    --aggregator https://aggregator.walrus-testnet.walrus.space --epochs 1\n")
			return
		}

		// Step 5: Deploy via HTTP (no wallet needed)
		fmt.Println("\n[5/5] Deploying to Walrus HTTP testnet...")
		fmt.Println("  (This may take a minute...)")

		deployCmd := exec.Command("walgo", "deploy-http",
			"--publisher", "https://publisher.walrus-testnet.walrus.space",
			"--aggregator", "https://aggregator.walrus-testnet.walrus.space",
			"--epochs", "1")
		deployCmd.Stdout = os.Stdout
		deployCmd.Stderr = os.Stderr

		if err := deployCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Deployment failed: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "💡 You can deploy manually:\n")
			fmt.Fprintf(os.Stderr, "  cd %s\n", siteName)
			fmt.Fprintf(os.Stderr, "  walgo deploy-http --publisher https://publisher.walrus-testnet.walrus.space \\\n")
			fmt.Fprintf(os.Stderr, "    --aggregator https://aggregator.walrus-testnet.walrus.space --epochs 1\n")
			os.Exit(1)
		}

		fmt.Println("\n═══════════════════════════════════════════════════════════")
		fmt.Println("🎉 Quickstart complete!")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Printf("\nYour site is now live on Walrus!\n\n")
		fmt.Printf("💡 What's next?\n")
		fmt.Printf("  • Preview locally: walgo serve\n")
		fmt.Printf("  • Add content: walgo new posts/my-post.md\n")
		fmt.Printf("  • Rebuild & redeploy: walgo build && walgo deploy-http ...\n")
		fmt.Printf("  • On-chain deploy: walgo setup --network testnet --force\n")
		fmt.Printf("\n📚 Documentation: https://github.com/selimozten/walgo\n")
	},
}

func init() {
	rootCmd.AddCommand(quickstartCmd)
	quickstartCmd.Flags().Bool("skip-deploy", false, "Skip deployment step")
	quickstartCmd.Flags().Bool("skip-build", false, "Skip build and deploy steps")
}
