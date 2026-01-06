package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/selimozten/walgo/internal/ai"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// aiResumeCmd resumes generation from an existing plan.
var aiResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume content generation from existing plan",
	Long: `Resume content generation from an existing plan.

If you have a plan.json file from a previous 'walgo ai plan' or
interrupted 'walgo ai pipeline' command, this will continue
generating the remaining pages.

Example:
  walgo ai resume`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()

		fmt.Printf("%s Resuming AI Generation\n", icons.Robot)
		fmt.Println()

		client, provider, model, err := ai.LoadClient(ai.LongRequestTimeout)
		if err != nil {
			return err
		}
		fmt.Printf("%s Using %s (%s)\n", icons.Check, provider, model)

		// Get current directory (should be the site directory when resuming)
		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		pipelineConfig := ai.DefaultPipelineConfig()
		pipelineConfig.Verbose = aiPipelineVerbose
		pipelineConfig.DryRun = aiPipelineDryRun
		// Set absolute paths to ensure content is created in the current directory
		pipelineConfig.ContentDir = filepath.Join(sitePath, "content")
		pipelineConfig.PlanPath = filepath.Join(sitePath, ".walgo", "plan.json")

		pipeline := ai.NewPipeline(client, pipelineConfig)
		pipeline.SetProgressHandler(ai.ConsoleProgressHandler(aiPipelineVerbose))

		if !pipeline.HasPlan() {
			return fmt.Errorf("no plan found - run 'walgo ai plan' or 'walgo ai pipeline' first")
		}

		ctx := cmd.Context()
		result, err := pipeline.Resume(ctx)

		fmt.Println()
		if result != nil && result.Plan != nil {
			fmt.Printf("%s Results:\n", icons.Chart)
			fmt.Printf("   Total pages: %d\n", result.Plan.Stats.TotalPages)
			fmt.Printf("   Completed:   %d\n", result.Plan.Stats.CompletedPages)
			fmt.Printf("   Skipped:     %d\n", result.Plan.Stats.SkippedPages)
			fmt.Printf("   Failed:      %d\n", result.Plan.Stats.FailedPages)
			fmt.Printf("   Duration:    %v\n", result.Duration.Round(time.Second))
		}

		if err != nil {
			return fmt.Errorf("resume error: %w", err)
		}

		if result.Plan != nil {
		if err := applyMenuToConfig(result.Plan); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: %v\n", icons.Warning, err)
		}

		if result.Plan.SiteType == ai.SiteTypeBusiness {
			fmt.Printf("\n%s Validating and fixing content for Ananke theme...\n", icons.Gear)
			fixer := ai.NewContentFixer(sitePath, ai.SiteTypeBusiness)
			if err := fixer.FixAll(); err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Content fix failed: %v\n", icons.Warning, err)
			} else {
				fmt.Printf("   %s Content validated and fixed\n", icons.Check)
			}
		}
		manager, err := projects.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create project manager: %w", err)
		}
		defer manager.Close()

		if err := manager.CreateDraftProject(result.Plan.SiteName, sitePath); err != nil {
			return fmt.Errorf("failed to create draft project: %w", err)
			}
		}
		fmt.Printf("   %s Created draft project\n", icons.Check)

		fmt.Printf("\n%s Site generated successfully!\n", icons.Celebrate)
		fmt.Printf("Run 'walgo build' to build the site.\n")
		fmt.Printf("Run 'walgo serve' to serve the site.\n")
		return nil
	},
}
