package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/selimozten/walgo/internal/ai"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// aiPlanCmd runs just the planning phase.
var aiPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create a site plan without generating content",
	Long: `Create a site structure plan using AI without generating content.

The plan is saved to .walgo/plan.json and can be reviewed before
running 'walgo ai resume' to generate the content.

Example:
  walgo ai plan
  cat .walgo/plan.json
  walgo ai resume`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("%s AI Site Planner\n", icons.Robot)
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
				// Clean up the directory if we created it and operation failed
				os.RemoveAll(sitePath)
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

		pipelineConfig := ai.DefaultPipelineConfig()
		pipelineConfig.Verbose = aiPipelineVerbose
		// Set absolute paths to ensure plan is created in the site directory
		pipelineConfig.ContentDir = filepath.Join(sitePath, "content")
		pipelineConfig.PlanPath = filepath.Join(sitePath, ".walgo", "plan.json")

		pipeline := ai.NewPipeline(client, pipelineConfig)
		pipeline.SetProgressHandler(ai.ConsoleProgressHandler(aiPipelineVerbose))

		input := &ai.PlannerInput{
			SiteName:    siteName,
			SiteType:    siteType,
			Description: description,
			Audience:    audience,
		}

		ctx := cmd.Context()
		plan, err := pipeline.PlanOnly(ctx, input)
		if err != nil {
			return fmt.Errorf("planning failed: %w", err)
		}

		if plan == nil {
			return fmt.Errorf("planning failed: no plan returned")
		}

		fmt.Println()
		fmt.Printf("%s Plan created with %d pages:\n", icons.Success, len(plan.Pages))
		for i, page := range plan.Pages {
			fmt.Printf("   %d. %s\n", i+1, page.Path)
		}

		success = true
		fmt.Printf("\n%s Plan saved to .walgo/plan.json\n", icons.File)
		fmt.Println("Run 'walgo ai resume' to generate content.")
		return nil
	},
}
