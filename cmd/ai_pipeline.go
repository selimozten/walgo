package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/ai"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	aiPipelineVerbose bool
	aiPipelineDryRun  bool
)

// aiPipelineCmd executes the full AI content generation pipeline: plan then generate.
var aiPipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Generate a complete site using AI (plan + generate)",
	Long: `Run the full AI content generation pipeline:
1. Plan: AI creates a site structure plan (JSON)
2. Generate: AI creates each page sequentially

The plan is saved to .walgo/plan.json for resumability.
If interrupted, run 'walgo ai resume' to continue.

Example:
  walgo ai pipeline`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("%s AI Site Pipeline\n", icons.Robot)
		fmt.Println()

		client, provider, model, err := ai.LoadClient(ai.LongRequestTimeout)
		if err != nil {
			return err
		}
		fmt.Printf("%s Using %s (%s)\n", icons.Check, provider, model)

		fmt.Println()
		fmt.Printf("Site name: ")
		siteName, _ := reader.ReadString('\n')
		siteName = strings.TrimSpace(siteName)
		if siteName == "" {
			return fmt.Errorf("site name is required")
		}

		fmt.Println()
		fmt.Println("Site type:")
		fmt.Println("  1) Blog")
		fmt.Println("  2) Portfolio")
		fmt.Println("  3) Docs")
		fmt.Println("  4) Business")
		fmt.Print("Select [1]: ")
		siteTypeChoice, _ := reader.ReadString('\n')
		siteTypeChoice = strings.TrimSpace(siteTypeChoice)
		if siteTypeChoice == "" {
			siteTypeChoice = "1"
		}

		var siteType ai.SiteType
		switch siteTypeChoice {
		case "1":
			siteType = ai.SiteTypeBlog
		case "2":
			siteType = ai.SiteTypePortfolio
		case "3":
			siteType = ai.SiteTypeDocs
		case "4":
			siteType = ai.SiteTypeBusiness
		default:
			return fmt.Errorf("invalid site type: %s", siteTypeChoice)
		}

		fmt.Println()
		fmt.Printf("Describe your site (1-2 sentences): ")
		description, _ := reader.ReadString('\n')
		description = strings.TrimSpace(description)

		fmt.Println()
		fmt.Printf("Target audience: ")
		audience, _ := reader.ReadString('\n')
		audience = strings.TrimSpace(audience)

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}

		sitePath = filepath.Join(sitePath, siteName)

		if err := os.MkdirAll(sitePath, 0755); err != nil {
			return fmt.Errorf("failed to create site directory: %w", err)
		}

		walgoConfigPath := filepath.Join(sitePath, config.DefaultConfigFileName)
		if _, err := os.Stat(walgoConfigPath); os.IsNotExist(err) {
			fmt.Printf("\n%s walgo.yaml not found, creating default configuration...\n", icons.Info)
			if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
				return fmt.Errorf("failed to create walgo.yaml: %w", err)
			}
			fmt.Printf("   %s Created walgo.yaml configuration\n", icons.Check)
		}

		hugoSiteType := hugo.SiteType(siteType)

		fmt.Printf("\n%s Setting up Hugo site...\n", icons.Spinner)

		if err := hugo.SetupSiteConfigWithName(sitePath, hugoSiteType, siteName); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not set up hugo.toml: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Configured hugo.toml for %s\n", icons.Check, siteType)
		}

		if err := hugo.SetupArchetypes(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not set up archetypes: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Set up archetypes\n", icons.Check)
		}

		if err := hugo.SetupFaviconForTheme(sitePath, hugoSiteType); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not set up favicon: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Set up favicon\n", icons.Check)
		}

		themeInfo := hugo.GetThemeInfo(hugoSiteType)
		fmt.Printf("   %s Installing theme %s...\n", icons.Spinner, themeInfo.Name)
		if err := hugo.InstallTheme(sitePath, hugoSiteType); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not install theme: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Installed theme %s\n", icons.Check, themeInfo.Name)
		}

		if hugoSiteType == hugo.SiteTypeBusiness {
			if err := hugo.SetupBusinessThemeOverrides(sitePath); err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Could not set up theme overrides: %v\n", icons.Warning, err)
			} else {
				fmt.Printf("   %s Set up theme overrides\n", icons.Check)
			}
		}
		if hugoSiteType == hugo.SiteTypePortfolio {
			if err := hugo.SetupPortfolioThemeOverrides(sitePath); err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Could not set up theme overrides: %v\n", icons.Warning, err)
			} else {
				fmt.Printf("   %s Set up theme overrides\n", icons.Check)
			}
		}
		if hugoSiteType == hugo.SiteTypeDocs {
			if err := hugo.SetupDocsThemeOverrides(sitePath); err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Could not set up theme overrides: %v\n", icons.Warning, err)
			} else {
				fmt.Printf("   %s Set up theme overrides\n", icons.Check)
			}
		}

		pipelineConfig := ai.DefaultPipelineConfig()
		pipelineConfig.Verbose = aiPipelineVerbose
		pipelineConfig.DryRun = aiPipelineDryRun
		// Set absolute paths to ensure content is created in the site directory
		pipelineConfig.ContentDir = filepath.Join(sitePath, "content")
		pipelineConfig.PlanPath = filepath.Join(sitePath, ".walgo", "plan.json")

		pipeline := ai.NewPipeline(client, pipelineConfig)
		pipeline.SetProgressHandler(ai.ConsoleProgressHandler(aiPipelineVerbose))

		input := &ai.PlannerInput{
			SiteName:    siteName,
			SiteType:    siteType,
			Description: description,
			Audience:    audience,
			Theme:       themeInfo.Name,
		}

		fmt.Printf("\n%s Generating content...\n", icons.Robot)
		ctx := cmd.Context()
		result, err := pipeline.Run(ctx, input)

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
			return fmt.Errorf("pipeline error: %w", err)
		}

		if result.Plan != nil {
			if err := applyMenuToConfig(result.Plan); err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: %v\n", icons.Warning, err)
			}

			runPostPipelineFixes(sitePath, siteType, result, icons)
		}

		if err := hugo.BuildSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Build failed: %v\n", icons.Warning, err)
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

		fmt.Printf("\n%s Site generated successfully!\n", icons.Celebrate)
		fmt.Printf("Run 'cd %s' to navigate to the site directory.\n", siteName)
		fmt.Printf("Run 'walgo build' to build the site.\n")
		fmt.Printf("Run 'walgo serve' to serve the site.\n")
		return nil
	},
}

// runPostPipelineFixes executes content validation and fixes based on site type.
func runPostPipelineFixes(sitePath string, siteType ai.SiteType, result *ai.PipelineResult, icons *ui.Icons) {
	switch siteType {
	case ai.SiteTypeBusiness:
		fmt.Printf("\n%s Validating and fixing content for Ananke theme...\n", icons.Gear)
		fixer := ai.NewContentFixer(sitePath, siteType)
		if err := fixer.FixAll(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Content fix failed: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Content validated and fixed\n", icons.Check)
		}

		issues := ai.ValidateBusinessContent(sitePath)
		if len(issues) > 0 {
			fmt.Printf("   %s Remaining issues:\n", icons.Warning)
			for _, issue := range issues {
				fmt.Printf("      - %s\n", issue)
			}
		}

	case ai.SiteTypeBlog:
		fmt.Printf("\n%s Validating and fixing content for Ananke theme...\n", icons.Gear)
		fixer := ai.NewContentFixer(sitePath, siteType)
		if err := fixer.FixAll(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Content fix failed: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Content validated and fixed\n", icons.Check)
		}

		issues := ai.ValidateBlogContent(sitePath)
		if len(issues) > 0 {
			fmt.Printf("   %s Remaining issues:\n", icons.Warning)
			for _, issue := range issues {
				fmt.Printf("      - %s\n", issue)
			}
		}

	case ai.SiteTypePortfolio:
		if err := hugo.UpdatePortfolioParams(sitePath, result.Plan.Description, result.Plan.Audience); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not update portfolio params: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Updated portfolio params\n", icons.Check)
		}

		fmt.Printf("\n%s Validating and fixing content for Ananke theme...\n", icons.Gear)
		fixer := ai.NewContentFixer(sitePath, siteType)
		if err := fixer.FixAll(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Content fix failed: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Content validated and fixed\n", icons.Check)
		}

		issues := ai.ValidatePortfolioContent(sitePath)
		if len(issues) > 0 {
			fmt.Printf("   %s Remaining issues:\n", icons.Warning)
			for _, issue := range issues {
				fmt.Printf("      - %s\n", issue)
			}
		}

	case ai.SiteTypeDocs:
		if err := hugo.UpdateDocsParams(sitePath, result.Plan.Description); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not update docs params: %v\n", icons.Warning, err)
		}

		fmt.Printf("\n%s Validating and fixing content for Hugo Book theme...\n", icons.Gear)
		fixer := ai.NewContentFixer(sitePath, siteType)
		if err := fixer.FixAll(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Content fix failed: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Content validated and fixed\n", icons.Check)
		}

		issues := ai.ValidateDocsContent(sitePath)
		if len(issues) > 0 {
			fmt.Printf("   %s Remaining issues:\n", icons.Warning)
			for _, issue := range issues {
				fmt.Printf("      - %s\n", issue)
			}
		}
	}
}
