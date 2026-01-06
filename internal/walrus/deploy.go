package walrus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/ui"
)

// DeploySite manages the site deployment to Walrus decentralized storage.
// Executes the `site-builder publish` command for new deployments.
// Context parameter enables cancellation and timeout control for the operation.
func DeploySite(ctx context.Context, deployDir string, walrusCfg config.WalrusConfig, epochs int) (*SiteBuilderOutput, error) {
	if epochs <= 0 {
		return nil, fmt.Errorf("epochs must be greater than 0, got %d", epochs)
	}

	icons := ui.GetIcons()

	fmt.Printf("%s Analyzing deployment directory...\n", icons.Chart)
	fileCount := 0
	totalSize := int64(0)

	err := filepath.Walk(deployDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		fmt.Printf("   %s Warning: Could not analyze directory: %v\n", icons.Warning, err)
	} else {
		icons := ui.GetIcons()
		fmt.Printf("%s Calculating deployment costs...\n", icons.Chart)

		network := "testnet"
		if walrusCfg.Network != "" {
			network = walrusCfg.Network
		}

		options := CostOptions{
			SiteSize:  totalSize,
			Epochs:    epochs,
			Network:   network,
			FileCount: fileCount,
			RPCURL:    "",
		}

		breakdown, err := CalculateCost(options)
		if err != nil {
			fmt.Printf("   %s Warning: Cost estimation failed: %v\n", icons.Warning, err)
			sizeMB := float64(totalSize) / (1024 * 1024)
			fallbackCost := sizeMB * 0.01 * float64(epochs)
			fmt.Printf("   %s Using fallback estimate: ~%.4f SUI\n", icons.Info, fallbackCost)
		} else {
			fmt.Printf("%s\n", FormatCostBreakdown(*breakdown))
			fmt.Printf("\n%s Summary: %s\n\n", icons.Info,
				FormatCostSummary(breakdown.GasCostSUI+breakdown.TotalWAL, breakdown.FileCount, epochs))

			if breakdown.GasCostSUI+breakdown.TotalWAL > 0.5 {
				fmt.Printf("%s %s Tip: Consider using `update-resources` for small changes\n", icons.Lightbulb, icons.Info)
				fmt.Printf("%s %s Tip: Use longer epochs for storage duration efficiency\n", icons.Lightbulb, icons.Info)
			}
		}
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found in PATH", siteBuilderCmd)
	}

	siteBuilderContext := GetWalrusContext()
	args := []string{
		"--context", siteBuilderContext,
		"publish",
		deployDir,
		"--epochs", fmt.Sprintf("%d", epochs),
	}

	if isVerbose() {
		fmt.Printf("%s Verbose mode enabled\n", icons.Wrench)
		fmt.Printf("%s Builder path: %s\n", icons.Wrench, builderPath)
		fmt.Printf("%s Arguments: %v\n", icons.Wrench, args)
	}

	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, strings.Join(args, " "))
	fmt.Printf("%s Uploading site files to Walrus...\n", icons.Upload)
	fmt.Printf("%s This may take several minutes depending on file count and network conditions...\n", icons.Hourglass)
	fmt.Printf("   (timeout: %v)\n", DefaultCommandTimeout)
	fmt.Println()

	stdoutStr, stderrStr, err := runCommandWithTimeout(ctx, builderPath, args, true)
	if err != nil {
		return nil, handleSiteBuilderError(err, stderrStr)
	}

	fmt.Printf("\n%s Site deployment command executed successfully.\n", icons.Success)

	combinedOutput := stdoutStr + "\n" + stderrStr
	output := parseSiteBuilderOutput(combinedOutput)
	output.Success = true

	if output.ObjectID != "" {
		fmt.Printf("\n%s Deployment successful!\n", icons.Celebrate)
		fmt.Printf("%s Site Object ID: %s\n", icons.Clipboard, output.ObjectID)
		fmt.Printf("\n%s Next steps:\n", icons.Pencil)
		fmt.Printf("1. Save this Object ID in your walgo.yaml:\n")
		fmt.Printf("   walrus:\n")
		fmt.Printf("     projectID: \"%s\"\n", output.ObjectID)
		fmt.Printf("2. Configure a SuiNS domain: walgo domain <your-domain>\n")
		fmt.Printf("3. Update your site: walgo update\n")
		fmt.Printf("4. Check status: walgo status\n")

		if len(output.BrowseURLs) > 0 {
			fmt.Printf("\n%s Browse your site:\n", icons.Globe)
			for _, url := range output.BrowseURLs {
				fmt.Printf("   %s\n", url)
			}
		}
	}

	return output, nil
}
