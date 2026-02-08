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
	"github.com/selimozten/walgo/internal/utils"
	"github.com/spf13/cobra"
)

var (
	aiPipelineVerbose    bool
	aiPipelineDryRun     bool
	aiPipelineParallel   string // auto, sequential, parallel
	aiPipelineConcurrent int
	aiPipelineRPM        int
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
		siteName, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("reading site name: %w", err)
		}
		if siteName == "" {
			return fmt.Errorf("site name is required")
		}

		fmt.Println()
		fmt.Println("Site type:")
		fmt.Println("  1) Blog")
		fmt.Println("  2) Docs")
		fmt.Print("Select [1]: ")
		siteTypeChoice, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("reading site type: %w", err)
		}
		if siteTypeChoice == "" {
			siteTypeChoice = "1"
		}

		var siteType ai.SiteType
		switch siteTypeChoice {
		case "1":
			siteType = ai.SiteTypeBlog
		case "2":
			siteType = ai.SiteTypeDocs
		default:
			return fmt.Errorf("invalid site type: %s", siteTypeChoice)
		}

		fmt.Println()
		fmt.Printf("Describe your site (1-2 sentences): ")
		description, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("reading description: %w", err)
		}

		fmt.Println()
		fmt.Printf("Target audience: ")
		audience, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("reading audience: %w", err)
		}

		// Sanitize site name for directory creation, keep original for config/DB
		sanitizedDirName := utils.SanitizeSiteName(siteName)
		if sanitizedDirName != siteName {
			fmt.Printf("Directory name sanitized: '%s' -> '%s'\n", siteName, sanitizedDirName)
		}

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}

		sitePath = filepath.Join(sitePath, sanitizedDirName)

		// Check if directory already exists before creating
		dirExistedBefore := false
		if _, err := os.Stat(sitePath); err == nil {
			dirExistedBefore = true
		}

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

		themeInfo := hugo.GetThemeInfo(hugoSiteType)

		if err := hugo.SetupSiteConfigWithName(sitePath, hugoSiteType, siteName); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not set up hugo.toml: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Configured hugo.toml for %s\n", icons.Check, siteType)
		}

		if err := hugo.SetupArchetypes(sitePath, themeInfo.DirName); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not set up archetypes: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Set up archetypes\n", icons.Check)
		}

		if err := hugo.SetupFaviconForTheme(sitePath, themeInfo.DirName); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not set up favicon: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Set up favicon\n", icons.Check)
		}
		fmt.Printf("   %s Installing theme %s...\n", icons.Spinner, themeInfo.Name)
		if err := hugo.InstallTheme(sitePath, hugoSiteType); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not install theme: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("   %s Installed theme %s\n", icons.Check, themeInfo.Name)
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

		// Apply parallel generation settings
		if aiPipelineParallel != "" {
			switch aiPipelineParallel {
			case "auto":
				pipelineConfig.ParallelMode = ai.ParallelModeAuto
			case "sequential":
				pipelineConfig.ParallelMode = ai.ParallelModeSequential
			case "parallel":
				pipelineConfig.ParallelMode = ai.ParallelModeParallel
			}
		}
		if aiPipelineConcurrent > 0 {
			pipelineConfig.MaxConcurrent = aiPipelineConcurrent
		}
		if aiPipelineRPM > 0 {
			pipelineConfig.RequestsPerMinute = aiPipelineRPM
		}

		pipeline := ai.NewPipeline(client, pipelineConfig)
		pipeline.SetProgressHandler(ai.ConsoleProgressHandler(aiPipelineVerbose))

		input := &ai.PlannerInput{
			SiteName:    siteName,
			SiteType:    siteType,
			Description: description,
			Audience:    audience,
			Theme:       themeInfo.Name,
			SitePath:    sitePath, // For dynamic theme analysis
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
			if err := applyMenuToConfig(result.Plan, sitePath); err != nil {
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

		success = true
		fmt.Printf("\n%s Site generated successfully!\n", icons.Celebrate)
		fmt.Printf("Run 'cd %s' to navigate to the site directory.\n", siteName)
		fmt.Printf("Run 'walgo build' to build the site.\n")
		fmt.Printf("Run 'walgo serve' to serve the site.\n")
		return nil
	},
}

// runPostPipelineFixes executes content validation and fixes based on site type.
func runPostPipelineFixes(sitePath string, siteType ai.SiteType, result *ai.PipelineResult, icons *ui.Icons) {
	// Get theme name from plan for dynamic theme support
	themeName := ""
	if result != nil && result.Plan != nil {
		themeName = result.Plan.Theme
	}

	switch siteType {
	case ai.SiteTypeBlog:
		fmt.Printf("\n%s Validating and fixing content for theme...\n", icons.Gear)
		fixer := ai.NewContentFixerWithTheme(sitePath, siteType, themeName)
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

	case ai.SiteTypeDocs:
		if err := hugo.UpdateDocsParams(sitePath, result.Plan.Description); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not update docs params: %v\n", icons.Warning, err)
		}

		fmt.Printf("\n%s Validating and fixing content for theme...\n", icons.Gear)
		fixer := ai.NewContentFixerWithTheme(sitePath, siteType, themeName)
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
