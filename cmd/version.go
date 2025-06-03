package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version will be set during build time via ldflags
	Version = "0.1.0"
	// GitCommit will be set during build time via ldflags
	GitCommit = "dev"
	// BuildDate will be set during build time via ldflags
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version number, git commit, and build date of Walgo.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Walgo version %s\n", Version)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Built: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
