package walrus

import (
	"context"
	"fmt"
	"time"

	"github.com/ganbitlabs/walgo/internal/ui"
)

// runSitemap runs `site-builder sitemap` and returns its stdout. It prints
// nothing, so it is safe to call from pre-flight checks.
func runSitemap(objectID string) (string, error) {
	if err := validateObjectID(objectID); err != nil {
		return "", fmt.Errorf("invalid object ID: %w", err)
	}

	builderPath, _, err := checkSiteBuilderSetupQuiet()
	if err != nil {
		return "", fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	// Find walrus binary path to pass to site-builder
	walrusPath, err := execLookPath("walrus")
	if err != nil {
		return "", fmt.Errorf("'walrus' CLI not found in PATH. Please install it using:\n  suiup install walrus@mainnet\n  Or run: walgo setup-deps")
	}

	args := []string{
		"--context", GetWalrusContext(),
		"--walrus-binary", walrusPath,
		"sitemap",
		objectID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	stdoutStr, stderrStr, err := runCommandWithTimeout(ctx, builderPath, args, false)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to execute %s: %v", siteBuilderCmd, err)
		if stderrStr != "" {
			errorMsg += fmt.Sprintf("\nstderr:\n%s", stderrStr)
		}
		if stdoutStr != "" {
			errorMsg += fmt.Sprintf("\nstdout:\n%s", stdoutStr)
		}
		return "", fmt.Errorf("%s", errorMsg)
	}

	return stdoutStr, nil
}

// GetSiteStatus checks the status of a Walrus site.
// Note: The site-builder doesn't have a direct "status" command, but we can use sitemap.
func GetSiteStatus(objectID string) (*SiteBuilderOutput, error) {
	icons := ui.GetIcons()
	fmt.Printf("%s Reading site resources from chain...\n", icons.Info)

	stdoutStr, err := runSitemap(objectID)
	if err != nil {
		return nil, err
	}

	fmt.Println("Site status retrieved successfully.")

	output := parseSitemapOutput(stdoutStr)
	output.Success = true
	output.ObjectID = objectID

	if stdoutStr != "" {
		fmt.Printf("Site resources:\n%s\n", stdoutStr)
	}

	return output, nil
}
