package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployment"
	"github.com/selimozten/walgo/internal/metrics"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/sui"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/selimozten/walgo/internal/version"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your Hugo site to Walrus Sites.",
	Long: `Deploys your Hugo site to Walrus Sites decentralized storage.
This command builds your site and uploads it to the Walrus network.

The site will be stored for the specified number of epochs (default: 1).
After deployment, you'll receive an object ID that you can use to access
your site and configure domain names.

Example: walgo deploy --epochs 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		quiet, _ := cmd.Flags().GetBool("quiet")

		telemetry, _ := cmd.Flags().GetBool("telemetry")
		var collector *metrics.Collector
		var startTime time.Time
		var deployMetrics metrics.DeployMetrics
		success := false
		if telemetry {
			collector = metrics.New(true)
			startTime = collector.Start()
			defer func() {
				if collector != nil {
					_ = collector.RecordDeploy(startTime, &deployMetrics, success)
				}
			}()
		}

		icons := ui.GetIcons()

		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
			return fmt.Errorf("error getting current directory: %w", err)
		}

		walgoCfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			fmt.Fprintf(os.Stderr, "\n%s Tip: Run 'walgo init <site-name>' to create a new site\n", icons.Lightbulb)
			return fmt.Errorf("error loading config: %w", err)
		}

		epochs, _ := cmd.Flags().GetInt("epochs")
		force, _ := cmd.Flags().GetBool("force")
		verbose, _ := cmd.Flags().GetBool("verbose")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		forceNew, _ := cmd.Flags().GetBool("force-new")
		skipVersionCheck, _ := cmd.Flags().GetBool("skip-version-check")
		saveProject, _ := cmd.Flags().GetBool("save-project")
		projectName, _ := cmd.Flags().GetString("project-name")
		category, _ := cmd.Flags().GetString("category")
		description, _ := cmd.Flags().GetString("description")
		imageURL, _ := cmd.Flags().GetString("image-url")

		// Check if project should be saved and validate project name
		if saveProject || projectName != "" {
			// Determine project name
			if projectName == "" {
				projectName = filepath.Base(sitePath)
				if projectName == "" || projectName == "." || projectName == "/" {
					projectName = "my-walgo-site"
				}
			}

			// Check if project name already exists
			pm, err := projects.NewManager()
			if err == nil {
				defer pm.Close()
				exists, err := pm.ProjectNameExists(projectName)
				if err == nil && exists {
					return fmt.Errorf("project name '%s' already exists. Use 'walgo projects' to view existing projects or choose a different name with --project-name", projectName)
				}
			}
		}

		publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
		if _, err := os.Stat(publishDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s Error: Build directory '%s' not found\n\n", icons.Error, publishDir)
			fmt.Fprintf(os.Stderr, "%s Run this first:\n", icons.Lightbulb)
			fmt.Fprintf(os.Stderr, "   walgo build\n")
			if !force {
				return fmt.Errorf("publish directory not found: %s", publishDir)
			}
		}

		if !quiet {
			fmt.Printf("%s Deploying to Walrus Sites...\n", icons.Rocket)
			fmt.Println("  [1/5] Verifying site...")
		}

		if !quiet {
			fmt.Println("  [2/5] Checking environment...")
		}
		if !skipVersionCheck {
			if err := version.CheckAndUpdateVersions(quiet); err != nil && !quiet {
				fmt.Fprintf(os.Stderr, "%s Warning: Version check failed: %v\n", icons.Warning, err)
				fmt.Fprintf(os.Stderr, "  Continuing with deployment...\n")
			}
		}

		opts := deployment.DeploymentOptions{
			SitePath:    sitePath,
			PublishDir:  publishDir,
			Epochs:      epochs,
			WalgoCfg:    walgoCfg,
			Quiet:       quiet,
			Verbose:     verbose,
			ForceNew:    forceNew,
			DryRun:      dryRun,
			SaveProject: saveProject || cmd.Flags().Changed("project-name"),
			ProjectName: projectName,
			Category:    category,
			Description: description,
			ImageURL:    imageURL,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		result, err := deployment.PerformDeployment(ctx, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nDeployment failed: %v\n", err)
			return fmt.Errorf("deployment failed: %w", err)
		}

		success = result.Success
		if telemetry {
			deployMetrics.TotalFiles = 0
			deployMetrics.ChangedFiles = 0
			deployMetrics.UploadDuration = 0
		}

		if !quiet {
			fmt.Println()
			if result.IsUpdate {
				ui.PrintBox(icons.Success + " Site Updated Successfully!")
			} else {
				ui.PrintBox(icons.Celebrate + " Deployment Successful!")
			}
			fmt.Println()
			fmt.Printf("%s Site Object ID: %s\n", icons.File, result.ObjectID)
			fmt.Println()

			network, err := sui.GetActiveEnv()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Failed to get active network: %v\n", icons.Warning, err)
				network = "testnet"
			}

			// Show explorer links
			ui.PrintHeader(icons.Link, "View Your Object on the Sui Network")
			fmt.Println()
			suiscanURL := sui.GetSuiscanURL(network, result.ObjectID)
			suivisionURL := sui.GetSuivisionURL(network, result.ObjectID)
			fmt.Printf("   • Suiscan:    %s\n", suiscanURL)
			fmt.Printf("   • Suivision:  %s\n", suivisionURL)
			fmt.Println()

			// Show SuiNS configuration instructions
			ui.PrintHeader(icons.Globe, "Next: Configure SuiNS for Public Access")
			fmt.Println()
			if network == "mainnet" {
				fmt.Printf("   1. Visit: https://suins.io\n")
				fmt.Printf("   2. Connect wallet & purchase domain\n")
				fmt.Printf("   3. Select 'Link To Walrus Site'\n")
				fmt.Printf("   4. Enter Object ID: %s\n", result.ObjectID)
				fmt.Printf("   5. Access at: https://your-domain.wal.app\n")
				fmt.Println()
			}

			commands := map[string]string{
				"walgo status":   "Check deployment status",
				"walgo update":   "Update your site",
				"walgo projects": "View all projects",
			}
			ui.PrintCommands("Useful Commands", commands)
		} else {
			fmt.Printf("Site Object ID: %s\n", result.ObjectID)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().IntP("epochs", "e", 1, "Number of epochs to store the site")
	deployCmd.Flags().BoolP("force", "f", false, "Deploy even if public directory doesn't exist")
	deployCmd.Flags().BoolP("verbose", "v", false, "Show detailed output for debugging")
	deployCmd.Flags().BoolP("quiet", "q", false, "Suppress output (used internally by quickstart)")
	deployCmd.Flags().Bool("dry-run", false, "Preview deployment plan without actually deploying")
	deployCmd.Flags().Bool("telemetry", false, "Record deployment metrics to local JSON file (~/.walgo/metrics.json)")
	deployCmd.Flags().Bool("skip-version-check", false, "Skip version checking and updating (not recommended for mainnet)")
	deployCmd.Flags().Bool("save-project", false, "Save deployment as a project with default name (directory name)")
	deployCmd.Flags().String("project-name", "", "Custom project name (default: directory name)")
	deployCmd.Flags().String("category", "", "Project category (default: website)")
	deployCmd.Flags().String("description", "", "Site description for metadata")
	deployCmd.Flags().String("image-url", "", "Site image URL for metadata")
	deployCmd.Flags().Bool("force-new", false, "Force deployment as new site (ignore existing objectID)")
}
