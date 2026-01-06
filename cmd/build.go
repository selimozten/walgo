package cmd

import (
	"fmt"
	"os"

	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/ui"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Hugo site.",
	Long: `Builds the Hugo site using the configuration found in the current directory
(or the directory specified by global --config flag if walgo.yaml is there).
This command runs the 'hugo' command to generate static files typically into the 'public' directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		fmt.Printf("%s Building site...\n", icons.Package)

		fmt.Printf("Running Hugo build...\n")
		if err := hugo.BuildSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "\n%s Troubleshooting:\n", icons.Lightbulb)
			fmt.Fprintf(os.Stderr, "  - Check that Hugo is installed: hugo version\n")
			fmt.Fprintf(os.Stderr, "  - Check hugo.toml for syntax errors\n")
			fmt.Fprintf(os.Stderr, "  - Run: hugo --verbose (for detailed output)\n")
			return fmt.Errorf("hugo build failed: %w", err)
		}

		fmt.Printf("\n%s Build complete! Output: %s\n", icons.Success, sitePath+"/public")
		fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
		fmt.Printf("  - Preview: walgo serve\n")
		fmt.Printf("  - Deploy:  walgo launch\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
