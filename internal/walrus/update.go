package walrus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/selimozten/walgo/internal/ui"
)

// UpdateSite handles updating an existing site on Walrus.
// It executes the `site-builder deploy` command which auto-detects updates via ws-resources.json.
// The context can be used to cancel or timeout the operation.
func UpdateSite(ctx context.Context, deployDir, objectID string, epochs int) (*SiteBuilderOutput, error) {
	if err := validateObjectID(objectID); err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}

	if epochs <= 0 {
		return nil, fmt.Errorf("epochs must be greater than 0, got %d", epochs)
	}

	if err := CheckSiteBuilderSetup(); err != nil {
		return nil, fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	// Find walrus binary path to pass to site-builder
	walrusPath, err := execLookPath("walrus")
	if err != nil {
		return nil, fmt.Errorf("'walrus' CLI not found in PATH. Please install it using:\n  suiup install walrus@mainnet\n  Or run: walgo setup-deps")
	}

	siteBuilderContext := GetWalrusContext()
	args := []string{
		"--context", siteBuilderContext,
		"--walrus-binary", walrusPath,
		"deploy",
		"--epochs", fmt.Sprintf("%d", epochs),
		deployDir,
	}

	icons := ui.GetIcons()
	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, strings.Join(args, " "))
	fmt.Printf("%s Updating site files on Walrus...\n", icons.Upload)

	var totalSize int64
	var fileCount int
	var walkErrors int
	_ = filepath.Walk(deployDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			walkErrors++
			return nil
		}
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})
	if walkErrors > 0 {
		fmt.Fprintf(os.Stderr, "%s Warning: Could not access %d file(s) during size calculation\n", icons.Warning, walkErrors)
	}

	sizeMB := float64(totalSize) / (1024 * 1024)
	simpleCost := sizeMB * 0.01 * float64(epochs)
	fmt.Printf("   %s Estimated cost: ~%.4f SUI (%d files, %.2f MB)\n", icons.Info, simpleCost, fileCount, sizeMB)

	fmt.Printf("%s This may take several minutes depending on file count and network conditions...\n", icons.Hourglass)
	fmt.Printf("   (timeout: %v)\n", DefaultCommandTimeout)
	fmt.Println()

	stdoutStr, stderrStr, err := runCommandWithTimeout(ctx, builderPath, args, true)
	if err != nil {
		combinedErr := stderrStr
		if combinedErr == "" && stdoutStr != "" {
			combinedErr = stdoutStr
		}
		fmt.Fprintf(os.Stderr, "\n%s Debug: command=%s args=%v\n", icons.Wrench, builderPath, args)
		if stdoutStr != "" {
			fmt.Fprintf(os.Stderr, "%s Debug stdout: %s\n", icons.Wrench, stdoutStr)
		}
		if stderrStr != "" {
			fmt.Fprintf(os.Stderr, "%s Debug stderr: %s\n", icons.Wrench, stderrStr)
		}
		return nil, handleSiteBuilderError(err, combinedErr)
	}

	fmt.Printf("\n%s Site update command executed successfully.\n", icons.Success)

	combinedOutput := stdoutStr + "\n" + stderrStr
	output := parseSiteBuilderOutput(combinedOutput)
	output.Success = true
	output.ObjectID = objectID

	fmt.Printf("\n%s Site updated successfully!\n", icons.Success)
	fmt.Printf("%s Object ID: %s\n", icons.Clipboard, objectID)
	fmt.Printf("%s Your site should be updated at the same URLs as before\n", icons.Globe)

	return output, nil
}
