package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/selimozten/walgo/internal/ai"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	aiGenerateNoBuild bool
	aiGenerateServe   bool
)

// aiCmd represents the root AI command group.
var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI-powered content generation and site creation for Hugo",
	Long: `Use AI to generate content, update files, and create complete Hugo sites.

Supports OpenAI and OpenRouter providers. Configure your API credentials first using 'walgo ai configure'.

Examples:
  walgo ai configure          # Set up AI provider credentials
  walgo ai generate           # Generate new content with auto-detection
  walgo ai update <file>      # Update existing content with AI
  walgo ai pipeline           # Create a complete site using AI pipeline`,
}

// applyMenuToConfig applies Hugo menu configuration from the site plan.
func applyMenuToConfig(plan *ai.SitePlan) error {
	sitePath, getwdErr := os.Getwd()
	if getwdErr != nil {
		return fmt.Errorf("failed to get current directory: %w", getwdErr)
	}

	hugoTomlPath := filepath.Join(sitePath, "hugo.toml")
	if _, err := os.Stat(hugoTomlPath); os.IsNotExist(err) {
		hugoTomlPath = filepath.Join(sitePath, "config.toml")
	}

	if _, err := os.Stat(hugoTomlPath); err == nil {
		fmt.Printf("\n%s Updating Hugo menu configuration...\n", ui.GetIcons().Gear)
		if err := hugo.ApplyMenuFromSitePlan(plan, hugoTomlPath); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: failed to apply menu: %v\n", ui.GetIcons().Warning, err)
		} else {
			fmt.Printf("%s Menu configuration updated\n", ui.GetIcons().Check)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(aiCmd)
	aiCmd.AddCommand(aiConfigureCmd)
	aiCmd.AddCommand(aiGenerateCmd)
	aiCmd.AddCommand(aiUpdateCmd)
	aiCmd.AddCommand(aiGetCmd)
	aiCmd.AddCommand(aiRemoveCmd)
	aiCmd.AddCommand(aiPipelineCmd)
	aiCmd.AddCommand(aiPlanCmd)
	aiCmd.AddCommand(aiResumeCmd)

	aiGenerateCmd.Flags().BoolVar(&aiGenerateNoBuild, "no-build", false, "Skip automatic build after generating")
	aiGenerateCmd.Flags().BoolVar(&aiGenerateServe, "serve", false, "Start development server after generating")

	aiPipelineCmd.Flags().BoolVarP(&aiPipelineVerbose, "verbose", "v", false, "Show verbose output")
	aiPipelineCmd.Flags().BoolVar(&aiPipelineDryRun, "dry-run", false, "Plan and generate without writing files")
	aiPlanCmd.Flags().BoolVarP(&aiPipelineVerbose, "verbose", "v", false, "Show verbose output")
	aiResumeCmd.Flags().BoolVarP(&aiPipelineVerbose, "verbose", "v", false, "Show verbose output")
	aiResumeCmd.Flags().BoolVar(&aiPipelineDryRun, "dry-run", false, "Generate without writing files")
}
