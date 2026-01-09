package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [site-name]",
	Short: "Initialize a new Hugo site with Walrus Sites configuration.",
	Long: `Initializes a new Hugo site in a directory specified by [site-name].
It sets up the basic Hugo structure and creates a walgo.yaml configuration
file tailored for Walrus Sites deployment.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		siteName := args[0]
		quiet, _ := cmd.Flags().GetBool("quiet")

		if !quiet {
			fmt.Printf("%s Initializing new Walgo site: %s\n\n", icons.Rocket, siteName)
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
			return fmt.Errorf("cannot determine current directory: %w", err)
		}
		siteName = strings.TrimSpace(siteName)
		if siteName == "" {
			return fmt.Errorf("site name is required")
		}
		sitePath := filepath.Join(cwd, siteName)

		// Check if directory already exists before creating
		dirExistedBefore := false
		if _, err := os.Stat(sitePath); err == nil {
			dirExistedBefore = true
		}

		// #nosec G301 - site directory needs standard permissions
		if err := os.MkdirAll(sitePath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Failed to create site directory %s: %v\n", icons.Error, sitePath, err)
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

		if !quiet {
			fmt.Printf("  %s Created directory: %s\n", icons.Check, sitePath)
		}

		if err := hugo.InitializeSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Failed to initialize Hugo site in %s: %v\n", icons.Error, sitePath, err)
			return fmt.Errorf("failed to initialize Hugo site: %w", err)
		}
		if !quiet {
			fmt.Printf("  %s Hugo site initialized\n", icons.Check)
		}

		if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Failed to create walgo.yaml in %s: %v\n", icons.Error, sitePath, err)
			return fmt.Errorf("failed to create walgo.yaml: %w", err)
		}
		if !quiet {
			fmt.Printf("  %s Created walgo.yaml configuration\n", icons.Check)
		}

		err = hugo.BuildSite(sitePath)
		if err != nil {
			return fmt.Errorf("failed to build site: %w", err)
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
		if !quiet {
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Printf("%s Site initialized successfully!\n", icons.Success)
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
			fmt.Printf("   - cd %s\n", siteName)
			fmt.Println("   - Add content: walgo new posts/my-first-post.md")
			fmt.Println("   - Build site: walgo build")
			fmt.Println("   - Preview: walgo serve")
			fmt.Printf("\n%s Deploy:\n", icons.Rocket)
			fmt.Println("   walgo launch    # Interactive wizard (recommended)")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolP("quiet", "q", false, "Suppress output (used internally by quickstart)")
}
