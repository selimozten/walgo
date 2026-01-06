// Package cmd provides CLI commands for walgo.
// This file defines the uninstall command for removing walgo from the system.
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	uninstallForce     bool
	uninstallDesktop   bool
	uninstallCLI       bool
	uninstallAll       bool
	uninstallKeepCache bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall walgo CLI and/or desktop app",
	Long: `Uninstall walgo CLI binary and/or desktop application.

This command helps you cleanly remove walgo from your system.

By default, it will ask what you want to uninstall.
Use flags to skip prompts and specify exactly what to remove.`,
	Example: `  # Interactive uninstall
  walgo uninstall

  # Uninstall everything without prompts
  walgo uninstall --all --force

  # Uninstall only CLI
  walgo uninstall --cli --force

  # Uninstall only desktop app
  walgo uninstall --desktop --force

  # Uninstall everything but keep cache
  walgo uninstall --all --keep-cache --force`,
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().BoolVarP(&uninstallForce, "force", "f", false, "Skip confirmation prompts")
	uninstallCmd.Flags().BoolVar(&uninstallDesktop, "desktop", false, "Uninstall desktop app only")
	uninstallCmd.Flags().BoolVar(&uninstallCLI, "cli", false, "Uninstall CLI only")
	uninstallCmd.Flags().BoolVarP(&uninstallAll, "all", "a", false, "Uninstall both CLI and desktop app")
	uninstallCmd.Flags().BoolVar(&uninstallKeepCache, "keep-cache", false, "Keep cache and data files")
}
