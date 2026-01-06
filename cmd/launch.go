package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployment"
	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/launch"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/sui"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/selimozten/walgo/internal/version"
	"github.com/spf13/cobra"
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Interactive wizard to deploy your site to Walrus",
	Long: `Launch provides an interactive step-by-step wizard to deploy your site to Walrus.

The wizard guides you through:
  • Choosing network (testnet/mainnet)
  • Selecting/adding wallet
  • Naming your project
  • Setting storage duration (epochs)
  • Reviewing gas fees
  • Publishing your site
  • Configuring SuiNS for public access (post-deployment)

All deployments are saved and can be managed with 'walgo projects'.

Example:
  walgo launch`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()

		// Ensure readline is properly cleaned up at the end
		defer launch.CloseReadline()

		fmt.Println()
		fmt.Println("╔═══════════════════════════════════════════════════════════╗")
		fmt.Printf("║              %s Walrus Site Launch Wizard                 ║\n", icons.Rocket)
		fmt.Println("╚═══════════════════════════════════════════════════════════╝")
		fmt.Println()

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		err = hugo.BuildSite(sitePath)
		if err != nil {
			return fmt.Errorf("failed to build site: %w", err)
		}

		// Step 1: Choose Network
		fmt.Println("Step 1: Choose Network")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
		network, err := launch.SelectNetwork()
		if err != nil {
			return err
		}

		netConfig := projects.GetNetworkConfig(network)
		fmt.Printf("\n%s Network: %s\n", icons.Check, network)
		fmt.Printf("  %s Epoch duration: %s\n", icons.Arrow, netConfig.EpochDuration)
		fmt.Printf("  %s Maximum epochs: %d\n", icons.Arrow, netConfig.MaxEpochs)
		fmt.Printf("  %s SuiNS available for public access (configure after deployment)\n", icons.Arrow)
		fmt.Println()

		// Check required tools
		fmt.Printf("%s Checking required tools...\n", icons.Info)
		missingTools := deps.GetMissingTools()
		if len(missingTools) > 0 {
			fmt.Printf("\n%s Missing required tools: %s\n", icons.Error, strings.Join(missingTools, ", "))
			fmt.Printf("\n%s %s", icons.Lightbulb, deps.InstallInstructions(network))
			return fmt.Errorf("missing required tools: %s", strings.Join(missingTools, ", "))
		}
		fmt.Printf("%s All required tools found\n", icons.Check)
		fmt.Println()

		if err := version.CheckAndUpdateVersions(false); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Version check failed: %v\n", icons.Warning, err)
			fmt.Fprintf(os.Stderr, "  Continuing with deployment...\n")
		}

		// Step 2: Check wallet
		fmt.Println("Step 2: Wallet & Balance")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
		walletAddr, suiBalance, walBalance, err := launch.CheckWallet(network)
		if err != nil {
			return err
		}
		fmt.Printf("\n%s Wallet: %s\n", icons.Check, walletAddr)
		fmt.Printf("  • Balance: %s SUI | %s WAL\n", suiBalance, walBalance)
		fmt.Println()

		// Step 3: Project details
		fmt.Println("Step 3: Project Details")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
		projectDetails, err := launch.GetProjectDetails()
		if err != nil {
			return err
		}
		fmt.Printf("\n%s Project: %s\n", icons.Check, projectDetails.Name)
		fmt.Printf("  • Category: %s\n", projectDetails.Category)
		fmt.Printf("  • Description: %s\n", projectDetails.Description)
		fmt.Println()

		// Step 4: Storage duration
		fmt.Println("Step 4: Storage Duration")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
		epochs, err := launch.SelectEpochs(netConfig)
		if err != nil {
			return err
		}
		duration := projects.CalculateStorageDuration(epochs, network)
		fmt.Printf("\n%s Storage: %d epochs (%s)\n", icons.Check, epochs, duration)
		fmt.Println()

		// Step 5: Verify site is built
		fmt.Println("Step 5: Verify Site")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
		_, publishDir, siteSize, err := launch.VerifySite()
		if err != nil {
			return err
		}
		fmt.Printf("\n%s Site ready\n", icons.Check)
		fmt.Printf("  • Location: %s\n", publishDir)
		fmt.Printf("  • Size: %.2f MB\n", float64(siteSize)/(1024*1024))
		fmt.Println()

		// Step 6: Review & confirm
		fmt.Println("Step 6: Review & Confirm")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Get detailed cost estimate with epochs
		estimatedGas := projects.EstimateGasFeeWithEpochs(network, siteSize, epochs)

		// Count files for detailed breakdown
		var fileCount int
		filepath.Walk(publishDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				fileCount++
			}
			return nil
		})

		fmt.Printf("\n%s Deployment Summary:\n", icons.Info)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("  Network:          %s\n", network)
		fmt.Printf("  Project:          %s\n", projectDetails.Name)
		fmt.Printf("  Category:         %s\n", projectDetails.Category)
		fmt.Printf("  Wallet:           %s\n", walletAddr)
		fmt.Printf("  Balance:          %s SUI | %s WAL\n", suiBalance, walBalance)
		fmt.Printf("  Storage:          %d epochs (%s)\n", epochs, duration)
		fmt.Printf("  Site size:        %.2f MB (%d files)\n", float64(siteSize)/(1024*1024), fileCount)
		fmt.Printf("  Estimated cost:   %s\n", estimatedGas)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Show detailed cost breakdown if verbose
		fmt.Println()
		fmt.Printf("%s Cost Breakdown:\n", icons.Info)
		costEstimate, err := projects.EstimateGasFeeDetailed(network, siteSize, epochs, fileCount)
		if err == nil {
			walDisplay := fmt.Sprintf("%.4f", costEstimate.WAL)
			if costEstimate.WAL > 0 && costEstimate.WAL < 0.0001 {
				walDisplay = "< 0.0001"
			}
			fmt.Printf("  WAL (storage):    %s WAL (range: %s)\n", walDisplay, costEstimate.WALRange)
			fmt.Printf("  SUI (gas):        %.4f SUI (range: %s)\n", costEstimate.SUI, costEstimate.SUICostRange)
			fmt.Println()
			fmt.Printf("  %s WAL is used for Walrus storage, SUI for Sui transactions\n", icons.Info)
			fmt.Printf("  %s Use https://costcalculator.wal.app for official estimates\n", icons.Info)
		}
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		confirm := readlineConfirm(fmt.Sprintf("\n%s Ready to deploy? [Y/n]: ", icons.Rocket))

		if confirm != "" && confirm != "y" && confirm != "yes" {
			fmt.Printf("\n%s Deployment cancelled\n", icons.Cross)
			return nil
		}

		// Step 7: Deploy
		fmt.Printf("\n\n%s Launching deployment...\n", icons.Rocket)
		fmt.Println()

		sitePath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		walgoCfg, err := config.LoadConfigFrom(sitePath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		err = hugo.BuildSite(sitePath)
		if err != nil {
			return fmt.Errorf("failed to build site: %w", err)
		}

		// Prepare deployment options
		opts := deployment.DeploymentOptions{
			SitePath:    sitePath,
			PublishDir:  publishDir,
			Epochs:      epochs,
			WalgoCfg:    walgoCfg,
			Quiet:       false,
			Verbose:     true,
			ForceNew:    false,
			DryRun:      false,
			SaveProject: true,
			ProjectName: projectDetails.Name,
			Category:    projectDetails.Category,
			Network:     network,
			WalletAddr:  walletAddr,
			// Metadata for ws-resources.json (project name is used as site_name)
			Description: projectDetails.Description,
			ImageURL:    projectDetails.ImageURL,
		}

		// Perform deployment using common function
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		result, err := deployment.PerformDeployment(ctx, opts)
		if err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}

		if !result.Success {
			return fmt.Errorf("deployment failed: no object ID returned")
		}

		// Success!
		fmt.Println("╔═══════════════════════════════════════════════════════════╗")
		fmt.Printf("║              %s Deployment Successful!                    ║\n", icons.Celebrate)
		fmt.Println("╚═══════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("%s Site Object ID: %s\n", icons.Info, result.ObjectID)
		fmt.Println()

		// Show explorer links
		fmt.Printf("%s View your object on the Sui network:\n", icons.Link)
		fmt.Println()
		suiscanURL := sui.GetSuiscanURL(network, result.ObjectID)
		suivisionURL := sui.GetSuivisionURL(network, result.ObjectID)
		fmt.Printf("   • Suiscan:    %s\n", suiscanURL)
		fmt.Printf("   • Suivision:  %s\n", suivisionURL)
		fmt.Println()

		// SuiNS configuration instructions (mainnet only)
		if network == "mainnet" {
			fmt.Printf("%s To access your site publicly via SuiNS:\n", icons.Globe)
			fmt.Println()
			fmt.Println("   Step 1: Get a SuiNS domain")
			fmt.Println("   • Visit: https://suins.io")
			fmt.Println("   • Connect your wallet and purchase a domain")
			fmt.Println("   • Names must use only letters (a-z) and numbers (0-9)")
			fmt.Println()
			fmt.Println("   Step 2: Link SuiNS to your Walrus Site")
			fmt.Println("   • Go to https://suins.io → 'Names You Own'")
			fmt.Println("   • Click the three dots menu on your domain")
			fmt.Println("   • Select 'Link To Walrus Site'")
			fmt.Printf("   • Paste your Object ID: %s\n", result.ObjectID)
			fmt.Println("   • Approve the transaction")
			fmt.Println()
			fmt.Println("   Step 3: Access your site")
			fmt.Println("   • Your site will be at: https://your-domain.wal.app")
			fmt.Println()
			fmt.Printf("   %s Detailed guide: https://docs.wal.app/docs/walrus-sites/tutorial-suins\n", icons.Info)
			fmt.Println()
		}

		fmt.Printf("%s Next steps:\n", icons.Book)
		fmt.Println("   • View project: walgo projects")
		fmt.Println("   • Check status: walgo status")
		fmt.Println("   • Update site:  walgo projects update \"" + projectDetails.Name + "\"")
		fmt.Println()

		return nil
	},
}

// readlineConfirm reads a confirmation prompt using the shared readline from launch package
func readlineConfirm(prompt string) string {
	// Use launch package's readline helper which manages shared state
	result := launch.ReadlineInput(prompt)
	return strings.ToLower(result)
}

func init() {
	rootCmd.AddCommand(launchCmd)
}
