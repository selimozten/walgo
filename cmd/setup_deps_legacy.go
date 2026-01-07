package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// runLegacyInstall performs direct download installation without suiup.
// This is the fallback method for environments where suiup doesn't work.
func runLegacyInstall(cmd *cobra.Command, withSiteBuilder, withWalrus, withHugo bool, network string) error {
	icons := ui.GetIcons()

	binDir, _ := cmd.Flags().GetString("bin-dir")

	if binDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		binDir = filepath.Join(home, ".config", "walgo", "bin")
	}

	// #nosec G301 - bin directory needs standard permissions
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("failed to create bin dir: %w", err)
	}

	osStr, archStr, err := mapOSArch()
	if err != nil {
		return err
	}

	fmt.Printf("%s Using legacy direct download method\n", icons.Warning)
	fmt.Printf("  Target directory: %s\n", binDir)
	fmt.Printf("  Network: %s\n", network)
	fmt.Println()
	fmt.Printf("  Note: Consider using 'walgo setup-deps' (without --legacy)\n")
	fmt.Printf("  for the recommended suiup-based installation.\n")
	fmt.Println()

	if withSiteBuilder {
		fmt.Println("  [1/2] Installing site-builder...")
		url, _ := siteBuilderURL(osStr, archStr, network)
		dest := filepath.Join(binDir, "site-builder")
		if err := downloadAndInstall(url, dest); err != nil {
			return fmt.Errorf("site-builder install failed: %w", err)
		}
		fmt.Printf("        %s Installed: %s\n", icons.Check, dest)
	}

	if withWalrus {
		fmt.Println("  [2/2] Installing walrus client...")
		url, _ := walrusURL(osStr, archStr, network)
		dest := filepath.Join(binDir, "walrus")
		if err := downloadAndInstall(url, dest); err != nil {
			return fmt.Errorf("walrus install failed: %w", err)
		}
		fmt.Printf("        %s Installed: %s\n", icons.Check, dest)
	}

	if withHugo {
		fmt.Println()
		fmt.Printf("%s Checking Hugo...\n", icons.Info)
		if _, err := deps.LookPath("hugo"); err == nil {
			fmt.Printf("  %s Hugo already present in PATH\n", icons.Check)
		} else {
			fmt.Printf("  %s Hugo not found\n", icons.Warning)
			fmt.Printf("\n%s Install Hugo via your package manager:\n", icons.Lightbulb)
			fmt.Println("   macOS:  brew install hugo")
			fmt.Println("   Linux:  apt install hugo / yum install hugo")
			fmt.Println("   Or: https://gohugo.io/installation/")
		}
	}

	fmt.Println()
	fmt.Printf("%s Configuring...\n", icons.Gear)
	if err := wireWalrusBinary(binDir); err != nil {
		fmt.Printf("  %s Warning: %v\n", icons.Warning, err)
	} else {
		fmt.Printf("  %s Updated walrus_binary path in sites-config.yaml\n", icons.Check)
	}

	fmt.Println()
	fmt.Printf("%s Dependencies installed successfully!\n", icons.Success)
	fmt.Printf("\n%s Next step:\n", icons.Lightbulb)
	fmt.Println("   walgo doctor     # Verify your environment")
	return nil
}

// mapOSArch returns the OS and architecture strings used in download URLs.
func mapOSArch() (string, string, error) {
	var osStr, archStr string
	switch runtime.GOOS {
	case "darwin":
		osStr = "macos"
	case "linux":
		osStr = "ubuntu"
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	switch runtime.GOARCH {
	case "arm64":
		archStr = "arm64"
	case "amd64":
		archStr = "x86_64"
	default:
		return "", "", fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
	}
	return osStr, archStr, nil
}

func baseBucket() string { return "https://storage.googleapis.com/mysten-walrus-binaries" }

func siteBuilderURL(osStr, archStr, network string) (string, string) {
	if network == "" {
		network = "testnet"
	}
	name := fmt.Sprintf("site-builder-%s-latest-%s-%s", network, osStr, archStr)
	return fmt.Sprintf("%s/%s", baseBucket(), name), name
}

func walrusURL(osStr, archStr, network string) (string, string) {
	name := fmt.Sprintf("walrus-latest-%s-%s", osStr, archStr)
	if network == "mainnet" || network == "testnet" {
		candidate := fmt.Sprintf("walrus-%s-latest-%s-%s", network, osStr, archStr)
		return fmt.Sprintf("%s/%s", baseBucket(), candidate), candidate
	}
	return fmt.Sprintf("%s/%s", baseBucket(), name), name
}

// downloadAndInstall fetches a binary from URL and installs it to dest path.
func downloadAndInstall(url, dest string) error {
	resp, err := http.Get(url) // #nosec G107 - URL constructed from hardcoded base
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed (%d): %s", resp.StatusCode, url)
	}

	tmp := dest + ".tmp"
	f, err := os.Create(tmp) // #nosec G304 - dest is managed bin directory
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	f.Close()

	// #nosec G302 - binary files need execute permissions
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}

	return os.Rename(tmp, dest)
}
