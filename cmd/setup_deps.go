package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/ui"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

type walrusSitesConfig struct {
	Contexts map[string]struct {
		Package string `yaml:"package"`
		General struct {
			RPCURL       string `yaml:"rpc_url"`
			Wallet       string `yaml:"wallet"`
			WalrusBinary string `yaml:"walrus_binary"`
			WalrusConfig string `yaml:"walrus_config"`
			GasBudget    int    `yaml:"gas_budget"`
		} `yaml:"general"`
	} `yaml:"contexts"`
	DefaultContext string `yaml:"default_context"`
}

const suiupInstallScript = "https://raw.githubusercontent.com/MystenLabs/suiup/main/install.sh"

var setupDepsCmd = &cobra.Command{
	Use:   "setup-deps",
	Short: "Install required binaries (sui, walrus, site-builder) using suiup.",
	Long: `Installs Sui ecosystem tools using suiup, the official version manager.

suiup is the recommended way to install and manage sui, walrus, and site-builder.
It handles versioning, updates, and network-specific binaries automatically.

Examples:
  walgo setup-deps                              # Install all tools for testnet
  walgo setup-deps --network mainnet            # Install for mainnet
  walgo setup-deps --with-hugo                  # Also check Hugo installation
  walgo setup-deps --legacy                     # Use legacy direct download method`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()

		withSiteBuilder, err := cmd.Flags().GetBool("with-site-builder")
		if err != nil {
			return err
		}
		withWalrus, err := cmd.Flags().GetBool("with-walrus")
		if err != nil {
			return err
		}
		withSui, err := cmd.Flags().GetBool("with-sui")
		if err != nil {
			return err
		}
		withHugo, err := cmd.Flags().GetBool("with-hugo")
		if err != nil {
			return err
		}
		network, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}
		legacy, err := cmd.Flags().GetBool("legacy")
		if err != nil {
			return err
		}

		fmt.Println("╔═══════════════════════════════════════════════════════════╗")
		fmt.Println("║              Walgo Dependency Installer                   ║")
		fmt.Println("╚═══════════════════════════════════════════════════════════╝")
		fmt.Println()

		// Use legacy method if requested
		if legacy {
			return runLegacyInstall(cmd, withSiteBuilder, withWalrus, withHugo, network)
		}

		// Modern suiup-based installation
		fmt.Printf("%s Using suiup (recommended method)\n", icons.Package)
		fmt.Printf("  Network: %s\n", network)
		fmt.Println()

		// Check if suiup is installed
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

			// Refresh PATH to find suiup
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

		// Install Sui
		if withSui {
			fmt.Printf("  [1/3] Installing sui@%s...\n", network)
			if err := runSuiup(suiupPath, "install", fmt.Sprintf("sui@%s", network)); err != nil {
				fmt.Printf("        %s Failed: %v\n", icons.Cross, err)
			} else {
				fmt.Printf("        %s sui installed\n", icons.Check)
			}
		}

		// Install Walrus
		if withWalrus {
			fmt.Printf("  [2/3] Installing walrus@%s...\n", network)
			if err := runSuiup(suiupPath, "install", fmt.Sprintf("walrus@%s", network)); err != nil {
				fmt.Printf("        %s Failed: %v\n", icons.Cross, err)
			} else {
				fmt.Printf("        %s walrus installed\n", icons.Check)
			}
		}

		// Install site-builder (always mainnet - no testnet version exists)
		if withSiteBuilder {
			fmt.Printf("  [3/3] Installing site-builder@mainnet...\n")
			if err := runSuiup(suiupPath, "install", "site-builder@mainnet"); err != nil {
				fmt.Printf("        %s Failed: %v\n", icons.Cross, err)
			} else {
				fmt.Printf("        %s site-builder installed\n", icons.Check)
			}
		}

		// Check Hugo if requested
		if withHugo {
			fmt.Println()
			fmt.Printf("%s Checking Hugo...\n", icons.Info)
			if _, err := exec.LookPath("hugo"); err == nil {
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
	},
}

// installSuiup installs suiup using the official install script
func installSuiup() error {
	// Download and run the install script
	cmd := exec.Command("bash", "-c", fmt.Sprintf("curl -sSfL %s | sh", suiupInstallScript))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runSuiup executes a suiup command
func runSuiup(suiupPath string, args ...string) error {
	cmd := exec.Command(suiupPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runLegacyInstall performs the legacy direct download installation
func runLegacyInstall(cmd *cobra.Command, withSiteBuilder, withWalrus, withHugo bool, network string) error {
	icons := ui.GetIcons()

	binDir, err := cmd.Flags().GetString("bin-dir")
	if err != nil {
		return err
	}

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
		if _, err := exec.LookPath("hugo"); err == nil {
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

func mapOSArch() (string, string, error) {
	var osStr, archStr string
	switch runtime.GOOS {
	case "darwin":
		osStr = "macos"
	case "linux":
		// Use ubuntu builds as the common denominator
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
	// Examples:
	// site-builder-mainnet-latest-macos-arm64
	// site-builder-testnet-latest-ubuntu-x86_64
	name := fmt.Sprintf("site-builder-%s-latest-%s-%s", network, osStr, archStr)
	return fmt.Sprintf("%s/%s", baseBucket(), name), name
}

func walrusURL(osStr, archStr, network string) (string, string) {
	// Walrus client does not usually depend on network in filename; use stable name pattern
	// Examples seen: walrus-mainnet-latest-macos-arm64 (if available) or just walrus-*
	// Fallback to generic walrus-latest-<os>-<arch>
	name := fmt.Sprintf("walrus-latest-%s-%s", osStr, archStr)
	// Try network-pinned first
	if network == "mainnet" || network == "testnet" {
		candidate := fmt.Sprintf("walrus-%s-latest-%s-%s", network, osStr, archStr)
		return fmt.Sprintf("%s/%s", baseBucket(), candidate), candidate
	}
	return fmt.Sprintf("%s/%s", baseBucket(), name), name
}

func downloadAndInstall(url, dest string) error {
	resp, err := http.Get(url) // #nosec G107 - URL is constructed from hardcoded base bucket and known patterns
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed (%d): %s", resp.StatusCode, url)
	}
	tmp := dest + ".tmp"
	f, err := os.Create(tmp) // #nosec G304 - dest is a known binary path under managed bin directory
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close() // #nosec G104 - cleanup close, primary error already handled
		return err
	}
	f.Close() // #nosec G104 - error not critical, file already written
	// #nosec G302 - binary files need execute permissions
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return err
	}
	return nil
}

func wireWalrusBinary(binDir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.New("failed to get home directory")
	}
	scPath := filepath.Join(home, ".config", "walrus", "sites-config.yaml")
	data, err := os.ReadFile(scPath) // #nosec G304 - path is constructed from known directory
	if err != nil {
		return errors.New("sites-config.yaml not found; run walgo setup first")
	}
	var sc walrusSitesConfig
	if err := yaml.Unmarshal(data, &sc); err != nil {
		return fmt.Errorf("failed to parse sites-config.yaml: %w", err)
	}
	walrusPath := filepath.Join(binDir, "walrus")
	for k, ctx := range sc.Contexts {
		// Only set if empty to avoid clobbering custom paths
		if strings.TrimSpace(ctx.General.WalrusBinary) == "" {
			ctx.General.WalrusBinary = walrusPath
			sc.Contexts[k] = ctx
		}
	}
	out, err := yaml.Marshal(&sc)
	if err != nil {
		return err
	}
	// #nosec G306 - config file needs to be readable for site-builder
	if err := os.WriteFile(scPath, out, 0o644); err != nil {
		return err
	}
	return nil
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
