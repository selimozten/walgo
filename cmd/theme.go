package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage Hugo themes",
	Long:  `Install, list, and manage Hugo themes for your site.`,
}

var themeInstallCmd = &cobra.Command{
	Use:   "install <github-url>",
	Short: "Install a Hugo theme from GitHub",
	Long: `Install a Hugo theme from a GitHub repository URL.

This command will:
1. Remove any existing themes in the themes/ directory
2. Download the new theme from GitHub
3. Update hugo.toml with the new theme name

Examples:
  walgo theme install https://github.com/theNewDynamic/gohugo-theme-ananke
  walgo theme install https://github.com/alex-shpak/hugo-book
  walgo theme install https://github.com/panr/hugo-theme-terminal`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		githubURL := args[0]

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		// Verify this is a Hugo site
		if !isHugoSite(sitePath) {
			fmt.Fprintf(os.Stderr, "%s Error: Not a Hugo site directory\n", icons.Error)
			fmt.Fprintf(os.Stderr, "   Please run this command from within a Hugo site directory.\n")
			return fmt.Errorf("not a Hugo site directory")
		}

		fmt.Printf("%s Installing theme from: %s\n", icons.Package, githubURL)
		fmt.Println()

		themeName, err := hugo.InstallThemeFromURL(sitePath, githubURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return err
		}

		fmt.Println()
		fmt.Printf("%s Theme '%s' installed successfully!\n", icons.Check, themeName)
		fmt.Println()
		fmt.Printf("%s Next steps:\n", icons.Lightbulb)
		fmt.Printf("   1. Review theme documentation for configuration options\n")
		fmt.Printf("   2. Run 'walgo serve' to preview your site\n")
		fmt.Printf("   3. Run 'walgo build' to build for production\n")
		fmt.Println()

		return nil
	},
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed themes",
	Long:  `List all themes currently installed in the themes/ directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		themes, err := hugo.GetInstalledThemes(sitePath)
		if err != nil {
			return fmt.Errorf("listing themes: %w", err)
		}

		if len(themes) == 0 {
			fmt.Printf("%s No themes installed\n", icons.Info)
			fmt.Println()
			fmt.Printf("%s Install a theme with: walgo theme install <github-url>\n", icons.Lightbulb)
			return nil
		}

		fmt.Printf("%s Installed themes:\n", icons.Package)
		for _, theme := range themes {
			fmt.Printf("   • %s\n", theme)
		}
		fmt.Println()

		return nil
	},
}

var themeNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new Hugo theme scaffold",
	Long: `Create a new Hugo theme using Hugo's built-in theme scaffolding.

This command will:
1. Run 'hugo new theme <name>' in the current site directory
2. Update hugo.toml to use the new theme

Examples:
  walgo theme new my-theme
  walgo theme new custom-blog`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		themeName := args[0]

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		// Verify this is a Hugo site
		if !isHugoSite(sitePath) {
			fmt.Fprintf(os.Stderr, "%s Error: Not a Hugo site directory\n", icons.Error)
			fmt.Fprintf(os.Stderr, "   Please run this command from within a Hugo site directory.\n")
			return fmt.Errorf("not a Hugo site directory")
		}

		fmt.Printf("%s Creating theme '%s'...\n", icons.Package, themeName)

		if err := hugo.CreateTheme(sitePath, themeName); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return err
		}

		fmt.Printf("%s Theme '%s' created successfully!\n", icons.Check, themeName)
		fmt.Println()
		fmt.Printf("%s Next steps:\n", icons.Lightbulb)
		fmt.Printf("   1. Edit theme files in themes/%s/\n", themeName)
		fmt.Printf("   2. Run 'walgo serve' to preview your site\n")
		fmt.Println()

		return nil
	},
}

// isHugoSite checks if the given path contains a Hugo site
func isHugoSite(sitePath string) bool {
	// Check for hugo.toml or config.toml
	configFiles := []string{"hugo.toml", "config.toml", "hugo.yaml", "config.yaml"}
	for _, cf := range configFiles {
		if _, err := os.Stat(filepath.Join(sitePath, cf)); err == nil {
			return true
		}
	}
	return false
}

func init() {
	themeCmd.AddCommand(themeInstallCmd)
	themeCmd.AddCommand(themeListCmd)
	themeCmd.AddCommand(themeNewCmd)
	rootCmd.AddCommand(themeCmd)
}
