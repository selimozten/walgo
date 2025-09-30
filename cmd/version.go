package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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

const (
	githubReleasesAPI = "https://api.github.com/repos/selimozten/walgo/releases/latest"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version number, git commit, and build date of Walgo.`,
	Run: func(cmd *cobra.Command, args []string) {
		checkUpdates, _ := cmd.Flags().GetBool("check-updates")
		short, _ := cmd.Flags().GetBool("short")

		if short {
			fmt.Printf("v%s\n", Version)
			return
		}

		fmt.Printf("Walgo v%s\n", Version)
		fmt.Printf("Commit:  %s\n", GitCommit)
		fmt.Printf("Built:   %s\n", BuildDate)

		if checkUpdates {
			fmt.Println()
			checkForUpdates()
		}
	},
}

func checkForUpdates() {
	fmt.Print("Checking for updates... ")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(githubReleasesAPI)
	if err != nil {
		fmt.Println("failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("failed")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed")
		return
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		fmt.Println("failed")
		return
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(Version, "v")

	if latestVersion == currentVersion {
		fmt.Println("✓")
		fmt.Println("\n✓ You are using the latest version!")
	} else if latestVersion > currentVersion {
		fmt.Println("✓")
		fmt.Printf("\n⚠ New version available: v%s (you have v%s)\n", latestVersion, currentVersion)
		fmt.Printf("\nUpdate with:\n")
		fmt.Printf("  curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash\n")
		fmt.Printf("\nRelease notes: %s\n", release.HTMLURL)
	} else {
		fmt.Println("✓")
		fmt.Println("\n✓ You are using the latest version (or a development build)")
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().Bool("check-updates", false, "Check for available updates")
	versionCmd.Flags().Bool("short", false, "Print version number only")
}
