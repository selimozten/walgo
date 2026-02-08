package walrus

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/selimozten/walgo/internal/ui"
)

// DestroySite handles destroying an existing site on Walrus.
// It executes the `site-builder destroy` command.
// The context can be used to cancel or timeout the operation.
func DestroySite(ctx context.Context, objectID string) error {
	if err := validateObjectID(objectID); err != nil {
		return fmt.Errorf("invalid object ID: %w", err)
	}

	if err := CheckSiteBuilderSetup(); err != nil {
		return fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	// Find walrus binary path to pass to site-builder
	walrusPath, err := execLookPath("walrus")
	if err != nil {
		return fmt.Errorf("'walrus' CLI not found in PATH. Please install it using:\n  suiup install walrus@mainnet\n  Or run: walgo setup-deps")
	}

	siteBuilderContext := GetWalrusContext()
	args := []string{
		"--context", siteBuilderContext,
		"--walrus-binary", walrusPath,
		"destroy",
		objectID,
	}

	icons := ui.GetIcons()
	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, args)
	fmt.Printf("%s Destroying site on Walrus...\n", icons.Garbage)
	fmt.Println()

	// Use CommandContext so the process is killed automatically on context cancellation
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := execCommandContext(ctx, builderPath, args...)
	var stdout, stderr bytes.Buffer

	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("destroy operation cancelled: %w", ctx.Err())
		}
		return handleSiteBuilderError(err, stderr.String())
	}

	fmt.Printf("\n%s Site destroyed successfully on Walrus!\n", icons.Success)

	return nil
}
