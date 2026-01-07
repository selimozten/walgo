package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

const suiupInstallScript = "https://raw.githubusercontent.com/MystenLabs/suiup/main/install.sh"

// runSetupDeps handles the main logic for installing dependencies via suiup.
func runSetupDeps(cmd *cobra.Command, args []string) error {
	icons := ui.GetIcons()

	withSiteBuilder, _ := cmd.Flags().GetBool("with-site-builder")
	withWalrus, _ := cmd.Flags().GetBool("with-walrus")
	withSui, _ := cmd.Flags().GetBool("with-sui")
	withHugo, _ := cmd.Flags().GetBool("with-hugo")
	network, _ := cmd.Flags().GetString("network")
	legacy, _ := cmd.Flags().GetBool("legacy")

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              Walgo Dependency Installer                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if legacy {
		return runLegacyInstall(cmd, withSiteBuilder, withWalrus, withHugo, network)
	}

	fmt.Printf("%s Using suiup (recommended method)\n", icons.Package)
	fmt.Printf("  Network: %s\n", network)
	fmt.Println()

	suiupPath, err := deps.LookPath("suiup")
	if err != nil {
		fmt.Printf("%s suiup not found. Installing suiup first...\n", icons.Info)
		fmt.Println()
		if err := installSuiup(); err != nil {
			fmt.Printf("%s Failed to install suiup: %v\n", icons.Error, err)
			fmt.Println()
			fmt.Printf("%s You can install suiup manually:\n", icons.Lightbulb)
			fmt.Printf("   curl -sSfL %s | sh\n", suiupInstallScript)
			fmt.Println()
			fmt.Printf("   Or use legacy method: walgo setup-deps --legacy\n")
			return fmt.Errorf("suiup installation failed: %w", err)
		}
		fmt.Printf("  %s suiup installed successfully\n", icons.Check)
		fmt.Println()

		suiupPath, _ = deps.LookPath("suiup")
		if suiupPath == "" {
			fmt.Printf("%s suiup installed but not in PATH\n", icons.Warning)
			fmt.Println("   Please restart your terminal or run:")
			fmt.Println("   source ~/.bashrc  # or ~/.zshrc")
			fmt.Println()
			fmt.Println("   Then run: walgo setup-deps")
			return nil
		}
	} else {
		fmt.Printf("  %s suiup found at %s\n", icons.Check, suiupPath)
	}

	fmt.Println()
	fmt.Printf("%s Installing tools via suiup...\n", icons.Gear)
	fmt.Println()

	if withSui {
		fmt.Printf("  [1/3] Installing sui@%s...\n", network)
		if err := runSuiup(suiupPath, "install", fmt.Sprintf("sui@%s", network)); err != nil {
			fmt.Printf("        %s Failed: %v\n", icons.Cross, err)
		} else {
			fmt.Printf("        %s sui installed\n", icons.Check)
		}
	}

	if withWalrus {
		fmt.Printf("  [2/3] Installing walrus@%s...\n", network)
		if err := runSuiup(suiupPath, "install", fmt.Sprintf("walrus@%s", network)); err != nil {
			fmt.Printf("        %s Failed: %v\n", icons.Cross, err)
		} else {
			fmt.Printf("        %s walrus installed\n", icons.Check)
		}
	}

	if withSiteBuilder {
		fmt.Printf("  [3/3] Installing site-builder@mainnet...\n")
		if err := runSuiup(suiupPath, "install", "site-builder@mainnet"); err != nil {
			fmt.Printf("        %s Failed: %v\n", icons.Cross, err)
		} else {
			fmt.Printf("        %s site-builder installed\n", icons.Check)
		}
	}

	if withHugo {
		fmt.Println()
		fmt.Printf("%s Checking Hugo...\n", icons.Info)
		if _, err := deps.LookPath("hugo"); err == nil {
			version := strings.TrimSpace(runQuiet("hugo", "version"))
			fmt.Printf("  %s Hugo already installed\n", icons.Check)
			if strings.Contains(strings.ToLower(version), "extended") {
				fmt.Printf("  %s Extended version detected\n", icons.Check)
			} else {
				fmt.Printf("  %s Note: Extended version recommended for SCSS support\n", icons.Warning)
			}
		} else {
			fmt.Printf("  %s Hugo not found\n", icons.Warning)
			fmt.Printf("\n%s Install Hugo Extended:\n", icons.Lightbulb)
			fmt.Println("   macOS:  brew install hugo")
			fmt.Println("   Linux:  apt install hugo / snap install hugo")
			fmt.Println("   Or: https://gohugo.io/installation/")
		}
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("%s Dependencies installed successfully!\n", icons.Success)
	fmt.Println()
	fmt.Printf("%s Useful commands:\n", icons.Lightbulb)
	fmt.Println("   suiup list              # Show installed tools")
	fmt.Println("   suiup update            # Update all tools")
	fmt.Println("   walgo doctor            # Verify your environment")
	fmt.Println("   walgo setup --network " + network + "  # Configure wallet")
	fmt.Println("═══════════════════════════════════════════════════════════")
	return nil
}

// installSuiup downloads and runs the official suiup install script.
// Downloads to a temp file first for security (avoids curl|sh pattern).
func installSuiup() error {
	// Download script to temp file first (safer than curl|sh)
	resp, err := http.Get(suiupInstallScript)
	if err != nil {
		return fmt.Errorf("failed to download suiup installer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download suiup installer: HTTP %d", resp.StatusCode)
	}

	// Create temp file for the script
	tmpDir := os.TempDir()
	scriptPath := filepath.Join(tmpDir, "suiup-install.sh")
	scriptFile, err := os.OpenFile(scriptPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		return fmt.Errorf("failed to create temp script file: %w", err)
	}

	// Copy script content to temp file
	_, err = io.Copy(scriptFile, resp.Body)
	scriptFile.Close()
	if err != nil {
		os.Remove(scriptPath)
		return fmt.Errorf("failed to write install script: %w", err)
	}

	// Clean up temp file when done
	defer os.Remove(scriptPath)

	// Execute the downloaded script
	cmd := exec.Command("sh", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runSuiup executes a suiup command with the given arguments.
func runSuiup(suiupPath string, args ...string) error {
	cmd := exec.Command(suiupPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
