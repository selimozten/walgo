// Package cmd provides CLI commands for walgo.
// This file defines the setup-deps command for installing Sui ecosystem tools.
package cmd

import (
	"github.com/spf13/cobra"
)

var setupDepsCmd = &cobra.Command{
	Use:   "setup-deps",
	Short: "Install required binaries (sui, walrus, site-builder) using suiup",
	Long: `Installs Sui ecosystem tools using suiup, the official version manager.

suiup is the recommended way to install and manage sui, walrus, and site-builder.
It handles versioning, updates, and network-specific binaries automatically.

Examples:
  walgo setup-deps                              # Install all tools for testnet
  walgo setup-deps --network mainnet            # Install for mainnet
  walgo setup-deps --with-hugo                  # Also check Hugo installation
  walgo setup-deps --legacy                     # Use legacy direct download method`,
	RunE: runSetupDeps,
}

func init() {
	rootCmd.AddCommand(setupDepsCmd)

	setupDepsCmd.Flags().String("bin-dir", "", "Directory for legacy installs (default: ~/.config/walgo/bin)")
	setupDepsCmd.Flags().Bool("with-sui", true, "Install sui CLI")
	setupDepsCmd.Flags().Bool("with-site-builder", true, "Install site-builder")
	setupDepsCmd.Flags().Bool("with-walrus", true, "Install walrus client")
	setupDepsCmd.Flags().Bool("with-hugo", false, "Check Hugo installation (prints guidance if missing)")
	setupDepsCmd.Flags().String("network", "testnet", "Network to target (testnet or mainnet)")
	setupDepsCmd.Flags().Bool("legacy", false, "Use legacy direct download instead of suiup")
}
