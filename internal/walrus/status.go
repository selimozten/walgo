package walrus

import (
	"context"
	"fmt"
	"time"

	"github.com/selimozten/walgo/internal/ui"
)

// GetSiteStatus checks the status of a Walrus site.
// Note: The site-builder doesn't have a direct "status" command, but we can use sitemap.
func GetSiteStatus(objectID string) (*SiteBuilderOutput, error) {
	if err := validateObjectID(objectID); err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}

	if err := CheckSiteBuilderSetup(); err != nil {
		return nil, fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	siteBuilderContext := GetWalrusContext()
	args := []string{
		"--context", siteBuilderContext,
		"sitemap",
		objectID,
	}

	icons := ui.GetIcons()
	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, args)

	statusTimeout := 2 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), statusTimeout)
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
		return nil, fmt.Errorf("%s", errorMsg)
	}

	fmt.Println("Site status retrieved successfully.")

	output := parseSitemapOutput(stdoutStr)
	output.Success = true
	output.ObjectID = objectID

	if stdoutStr != "" {
		fmt.Printf("Site resources:\n%s\n", stdoutStr)
	}

	if stderrStr != "" {
		fmt.Printf("Stderr from %s:\n%s\n", siteBuilderCmd, stderrStr)
	}

	return output, nil
}
