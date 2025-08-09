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

var setupDepsCmd = &cobra.Command{
	Use:   "setup-deps",
	Short: "Download and install required binaries (site-builder, walrus) to a managed bin dir.",
	Long: `Detects OS/arch and installs selected tools under ~/.config/walgo/bin (or a custom dir).
Also updates sites-config.yaml to point walrus_binary to the managed path.

Examples:
  walgo setup-deps --with-site-builder --with-walrus --network testnet
  walgo setup-deps --bin-dir ~/.local/bin --with-hugo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		binDir, _ := cmd.Flags().GetString("bin-dir")
		withSiteBuilder, _ := cmd.Flags().GetBool("with-site-builder")
		withWalrus, _ := cmd.Flags().GetBool("with-walrus")
		withHugo, _ := cmd.Flags().GetBool("with-hugo")
		network, _ := cmd.Flags().GetString("network")

		if binDir == "" {
			home, _ := os.UserHomeDir()
			binDir = filepath.Join(home, ".config", "walgo", "bin")
		}
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			return fmt.Errorf("failed to create bin dir: %w", err)
		}

		osStr, archStr, err := mapOSArch()
		if err != nil {
			return err
		}

		if withSiteBuilder {
			url, _ := siteBuilderURL(osStr, archStr, network)
			dest := filepath.Join(binDir, "site-builder")
			if err := downloadAndInstall(url, dest); err != nil {
				return fmt.Errorf("site-builder install failed: %w", err)
			}
			fmt.Println("✓ installed site-builder:", dest)
		}

		if withWalrus {
			url, _ := walrusURL(osStr, archStr, network)
			dest := filepath.Join(binDir, "walrus")
			if err := downloadAndInstall(url, dest); err != nil {
				return fmt.Errorf("walrus install failed: %w", err)
			}
			fmt.Println("✓ installed walrus:", dest)
		}

		if withHugo {
			// Try system package first; otherwise attempt a basic download hint
			if _, err := exec.LookPath("hugo"); err == nil {
				fmt.Println("✓ hugo already present in PATH")
			} else {
				fmt.Println("ℹ️  Please install Hugo via your package manager (e.g., brew install hugo).")
			}
		}

		// Update sites-config.yaml walrus_binary to managed path if present
		if err := wireWalrusBinary(binDir); err != nil {
			fmt.Println("Warning:", err)
		} else {
			fmt.Println("✓ wired walrus_binary path in sites-config.yaml")
		}

		fmt.Println("Done. You can run 'walgo doctor' next.")
		return nil
	},
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
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed (%d): %s", resp.StatusCode, url)
	}
	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	f.Close()
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return err
	}
	return nil
}

func wireWalrusBinary(binDir string) error {
	home, _ := os.UserHomeDir()
	scPath := filepath.Join(home, ".config", "walrus", "sites-config.yaml")
	data, err := os.ReadFile(scPath)
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
	if err := os.WriteFile(scPath, out, 0o644); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(setupDepsCmd)
	setupDepsCmd.Flags().String("bin-dir", "", "Directory to install tools (default: ~/.config/walgo/bin)")
	setupDepsCmd.Flags().Bool("with-site-builder", true, "Install site-builder")
	setupDepsCmd.Flags().Bool("with-walrus", true, "Install walrus client")
	setupDepsCmd.Flags().Bool("with-hugo", false, "Ensure Hugo is installed (prints guidance if missing)")
	setupDepsCmd.Flags().String("network", "testnet", "Network to target for downloads (testnet or mainnet)")
}
