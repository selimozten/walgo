package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Version will be set during build time via ldflags
	Version = "0.3.4"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		checkUpdates, err := cmd.Flags().GetBool("check-updates")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading check-updates flag: %v\n", err)
			return fmt.Errorf("error reading check-updates flag: %w", err)
		}
		short, err := cmd.Flags().GetBool("short")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading short flag: %v\n", err)
			return fmt.Errorf("error reading short flag: %w", err)
		}

		if short {
			fmt.Printf("v%s\n", Version)
			return nil
		}

		icons := ui.GetIcons()
		fmt.Printf("Walgo v%s\n", Version)
		fmt.Printf("Commit:  %s\n", GitCommit)
		fmt.Printf("Built:   %s\n", BuildDate)
		fmt.Printf("\n%s This is a beta release. Please report issues at:\n", icons.Warning)
		fmt.Printf("   https://github.com/selimozten/walgo/issues\n")

		if checkUpdates {
			fmt.Println()
			checkForUpdates()
		}

		return nil
	},
}

func checkForUpdates() {
	icons := ui.GetIcons()
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

	switch compareSemver(latestVersion, currentVersion) {
	case 0:
		fmt.Println(icons.Check)
		fmt.Printf("\n%s You are using the latest version!\n", icons.Check)
	case 1:
		fmt.Println(icons.Check)
		fmt.Printf("\n%s New version available: v%s (you have v%s)\n", icons.Warning, latestVersion, currentVersion)
		fmt.Printf("\nUpdate with:\n")
		fmt.Printf("  curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash\n")
		fmt.Printf("\nRelease notes: %s\n", release.HTMLURL)
	default:
		fmt.Println(icons.Check)
		fmt.Printf("\n%s You are using the latest version (or a development build)\n", icons.Check)
	}
}

func compareSemver(a, b string) int {
	parse := func(input string) [3]int {
		var result [3]int
		clean := strings.SplitN(input, "-", 2)[0]
		parts := strings.Split(clean, ".")
		for i := 0; i < len(result) && i < len(parts); i++ {
			if n, err := strconv.Atoi(parts[i]); err == nil {
				result[i] = n
			}
		}
		return result
	}
	av := parse(a)
	bv := parse(b)
	for i := 0; i < len(av); i++ {
		if av[i] > bv[i] {
			return 1
		}
		if av[i] < bv[i] {
			return -1
		}
	}
	return 0
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().Bool("check-updates", false, "Check for available updates")
	versionCmd.Flags().Bool("short", false, "Print version number only")
}
