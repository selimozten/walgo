package cmd

import (
	"fmt"

	"github.com/ganbitlabs/walgo/internal/ui"
	"github.com/ganbitlabs/walgo/internal/version"
)

// printToolchainReport prints the resolved path and version of every tool walgo
// shells out to, then flags any that are too old to talk to Sui.
//
// Installers report what they wrote to disk, which is not necessarily what runs:
// an older sui, walrus or site-builder earlier in PATH keeps winning, and the
// resulting "incompatible versions" reports are confusing. Printing the resolved
// path next to the version makes that immediately obvious.
func printToolchainReport() {
	icons := ui.GetIcons()

	fmt.Println()
	fmt.Printf("%s Verifying installed tools...\n", icons.Info)

	toolchain := version.InspectToolchain()
	for _, status := range toolchain {
		switch {
		case !status.Installed:
			fmt.Printf("  %s %s not found in PATH\n", icons.Warning, status.Tool)
		case status.Version == "":
			fmt.Printf("  %s %s version unknown (%s)\n", icons.Warning, status.Tool, status.Path)
		case status.Outdated:
			fmt.Printf("  %s %s v%s at %s (need v%s+)\n",
				icons.Cross, status.Tool, status.Version, status.Path, status.Minimum)
		default:
			fmt.Printf("  %s %s v%s (%s)\n", icons.Check, status.Tool, status.Version, status.Path)
		}
	}

	if err := version.ToolchainProblems(toolchain); err != nil {
		fmt.Println()
		fmt.Printf("  %s %v\n", icons.Cross, err)
	}
	fmt.Println()
}
