package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/sui"
	"github.com/selimozten/walgo/internal/ui"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

type sitesConfig struct {
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

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check environment and configuration for on-chain and HTTP deploys.",
	Long: `Diagnose your Walgo environment and check for common issues.

This command checks:
- Required binaries (hugo, site-builder, walrus, sui)
- Sui client configuration and active address
- Wallet token balances (SUI, WAL, and others)
- Configuration files
- Provides auto-fix suggestions

Examples:
  walgo doctor              # Run diagnostics
  walgo doctor --fix-paths  # Fix tilde paths in config
  walgo doctor --fix-all    # Auto-fix all issues`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		fixAll, err := cmd.Flags().GetBool("fix-all")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading fix-all flag: %v\n", icons.Error, err)
			return fmt.Errorf("error getting fix-all flag: %w", err)
		}
		fixPaths, err := cmd.Flags().GetBool("fix-paths")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading fix-paths flag: %v\n", icons.Error, err)
			return fmt.Errorf("error getting fix-paths flag: %w", err)
		}
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading verbose flag: %v\n", icons.Error, err)
			return fmt.Errorf("error getting verbose flag: %w", err)
		}

		fmt.Println("╔═══════════════════════════════════════════════════════════╗")
		fmt.Println("║                     Walgo Doctor                          ║")
		fmt.Println("║             Environment Diagnostics                       ║")
		fmt.Println("╚═══════════════════════════════════════════════════════════╝")
		fmt.Println()

		issues := 0
		warnings := 0

		// Check binaries
		fmt.Printf("%s Checking dependencies...\n", icons.Package)
		fmt.Println()

		binaries := map[string]struct {
			name     string
			required bool
			purpose  string
			install  string
		}{
			"hugo":         {"hugo", true, "Static site generation", installHint("hugo")},
			"site-builder": {"site-builder", false, "On-chain deployment", installHint("site-builder")},
			"walrus":       {"walrus", false, "Walrus CLI operations", installHint("walrus")},
			"sui":          {"sui", false, "On-chain wallet management", installHint("sui")},
		}

		// Check for suiup first (tool manager)
		if _, err := deps.LookPath("suiup"); err == nil {
			fmt.Printf("  %s suiup found (Sui tool manager)\n", icons.Check)
			if verbose {
				version := strings.TrimSpace(runQuiet("suiup", "--version"))
				if version != "" {
					fmt.Printf("    Version: %s\n", version)
				}
			}
		} else {
			fmt.Printf("  %s suiup not found (recommended tool manager)\n", icons.Warning)
			fmt.Println("    Install: curl -sSfL https://raw.githubusercontent.com/MystenLabs/suiup/main/install.sh | sh")
			fmt.Println("    Or run: walgo setup-deps")
			warnings++
		}

		for _, bin := range []string{"hugo", "site-builder", "walrus", "sui"} {
			info := binaries[bin]
			if path, err := deps.LookPath(bin); err != nil {
				if info.required {
					fmt.Printf("  %s %s not found (REQUIRED)\n", icons.Cross, info.name)
					fmt.Printf("    Purpose: %s\n", info.purpose)
					fmt.Printf("    Install: %s\n", info.install)
					issues++
				} else {
					fmt.Printf("  %s %s not found (optional for %s)\n", icons.Warning, info.name, info.purpose)
					fmt.Printf("    Install: %s\n", info.install)
					warnings++
				}
			} else {
				fmt.Printf("  %s %s found", icons.Check, info.name)
				if verbose {
					fmt.Printf(" at %s", path)
				}
				fmt.Println()

				// Get version info and check for Hugo Extended
				switch bin {
				case "hugo":
					version := strings.TrimSpace(runQuiet("hugo", "version"))
					if version != "" {
						if verbose {
							fmt.Printf("    Version: %s\n", version)
						}
						// Check if Hugo Extended is installed
						if !strings.Contains(strings.ToLower(version), "extended") {
							fmt.Printf("  %s Hugo Extended is required but standard Hugo is installed\n", icons.Warning)
							fmt.Println("    Extended version is needed for SCSS/SASS support")
							fmt.Println("    Install: brew install hugo (macOS) or download 'extended' from https://github.com/gohugoio/hugo/releases")
							warnings++
						} else if verbose {
							fmt.Printf("    %s Extended version detected\n", icons.Check)
						}
					}
				case "sui":
					if verbose {
						version := strings.TrimSpace(runQuiet("sui", "--version"))
						if version != "" {
							fmt.Printf("    Version: %s\n", version)
						}
					}
				}
			}
		}

		fmt.Println()

		if _, err := deps.LookPath("sui"); err == nil {
			fmt.Printf("%s Checking Sui configuration...\n", icons.Info)
			fmt.Println()

			activeEnv, _ := sui.GetActiveEnv()
			if activeEnv != "" {
				fmt.Printf("  %s Active network: %s\n", icons.Check, activeEnv)
			}

			address, _ := sui.GetActiveAddress()
			if address != "" {
				fmt.Printf("  %s Active address: %s\n", icons.Check, address)

				// Check token balances (SUI and WAL)
				balance, err := sui.GetBalance()

				if err != nil || (balance.SUI == 0 && balance.WAL == 0) {
					fmt.Printf("  %s No token balances found\n", icons.Cross)
					fmt.Println("    On-chain deployment requires SUI for gas fees")
					if strings.Contains(activeEnv, "testnet") {
						fmt.Printf("    Get testnet tokens: https://faucet.sui.io/?address=%s\n", address)
					} else if strings.Contains(activeEnv, "devnet") {
						fmt.Println("    Get devnet tokens: sui client faucet")
					}
					issues++
				} else {
					// Show SUI balance
					if balance.SUI > 0 {
						fmt.Printf("  %s SUI balance: %.2f SUI\n", icons.Check, balance.SUI)
						if balance.SUI < 0.1 {
							fmt.Printf("    %s Low SUI balance (< 0.1 SUI), consider getting more tokens\n", icons.Warning)
						}
					} else {
						fmt.Printf("  %s No SUI tokens found\n", icons.Cross)
						fmt.Println("    On-chain deployment requires SUI for gas fees")
						if strings.Contains(activeEnv, "testnet") {
							fmt.Printf("    Get testnet tokens: https://faucet.sui.io/?address=%s\n", address)
						}
						issues++
					}

					// Show WAL balance
					if balance.WAL > 0 {
						fmt.Printf("  %s WAL balance: %.2f WAL\n", icons.Check, balance.WAL)
						if verbose {
							fmt.Println("    WAL tokens provide storage quota on Walrus network")
						}
					} else {
						fmt.Printf("  %s No WAL tokens found\n", icons.Info)
						fmt.Println("    WAL tokens provide extended storage quota (optional but recommended)")
						if verbose {
							fmt.Println("    Get WAL: walrus get-wal --context testnet (testnet)")
						}
					}
				}
			} else {
				fmt.Printf("  %s No active Sui address configured\n", icons.Cross)
				fmt.Println("    Run: sui client")
				issues++
			}

			fmt.Println()
		}

		fmt.Printf("%s Checking configuration files...\n", icons.Info)
		fmt.Println()

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Cannot determine home directory: %v\n", icons.Error, err)
			return fmt.Errorf("error getting home directory: %w", err)
		}
		scPath := filepath.Join(home, ".config", "walrus", "sites-config.yaml")

		if _, err := os.Stat(scPath); err == nil {
			fmt.Printf("  %s sites-config.yaml found at %s\n", icons.Check, scPath)

			data, err := os.ReadFile(scPath) // #nosec G304 - path is constructed from known directory
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Could not read sites-config.yaml: %v\n", icons.Warning, err)
			} else {
				if containsTildePath(string(data)) {
					fmt.Printf("  %s Configuration contains tilde paths (~)\n", icons.Warning)
					fmt.Println("    Run: walgo doctor --fix-paths")
					warnings++

					if fixPaths || fixAll {
						if err := ensureAbsolutePaths(scPath, home); err != nil {
							fmt.Printf("  %s Failed to fix paths: %v\n", icons.Cross, err)
							issues++
						} else {
							fmt.Printf("  %s Fixed tilde paths to absolute paths\n", icons.Check)
						}
					}
				}
			}
		} else {
			fmt.Printf("  %s sites-config.yaml not found\n", icons.Warning)
			fmt.Println("    For on-chain deployment, run: walgo setup --network testnet --force")
			warnings++
		}

		if _, err := os.Stat("walgo.yaml"); err == nil {
			fmt.Printf("  %s walgo.yaml found in current directory\n", icons.Check)
		} else {
			fmt.Printf("  %s walgo.yaml not found in current directory\n", icons.Warning)
			fmt.Println("    Initialize a site: walgo init my-site")
			warnings++
		}

		fmt.Println()

		// Show deployment options
		fmt.Printf("%s Deployment options:\n", icons.Rocket)
		fmt.Println()
		fmt.Println("  Recommended: Interactive wizard")
		fmt.Println("    walgo launch")
		fmt.Println()
		fmt.Println("  Alternative: HTTP Testnet (No wallet required)")
		fmt.Println("    walgo deploy-http \\")
		fmt.Println("      --publisher https://publisher.walrus-testnet.walrus.space \\")
		fmt.Println("      --aggregator https://aggregator.walrus-testnet.walrus.space")
		fmt.Println()

		// Summary
		fmt.Println("═══════════════════════════════════════════════════════════")
		if issues == 0 && warnings == 0 {
			fmt.Printf("%s All checks passed! Your environment is ready.\n", icons.Success)
		} else {
			fmt.Printf("Summary: %d issue(s), %d warning(s)\n", issues, warnings)
			if issues > 0 {
				fmt.Printf("\n%s Please fix the issues above before deploying on-chain.\n", icons.Error)
			}
			if warnings > 0 {
				fmt.Printf("\n%s Warnings indicate optional features that may not work.\n", icons.Warning)
			}
			if !fixAll && warnings > 0 {
				fmt.Printf("\n%s Tip: Run 'walgo doctor --fix-all' to auto-fix some issues\n", icons.Lightbulb)
			}
		}
		fmt.Println("═══════════════════════════════════════════════════════════")

		return nil
	},
}

func runQuiet(name string, args ...string) string {
	out, _ := exec.Command(name, args...).CombinedOutput()
	return string(out)
}

func ensureAbsolutePaths(scPath, home string) error {
	data, err := os.ReadFile(scPath) // #nosec G304 - scPath is a known config file path
	if err != nil {
		return err
	}
	var sc sitesConfig
	if err := yaml.Unmarshal(data, &sc); err != nil {
		return err
	}
	for k, ctx := range sc.Contexts {
		ctx.General.Wallet = expandTilde(ctx.General.Wallet, home)
		ctx.General.WalrusConfig = expandTilde(ctx.General.WalrusConfig, home)
		ctx.General.WalrusBinary = expandTilde(ctx.General.WalrusBinary, home)
		sc.Contexts[k] = ctx
	}
	out, err := yaml.Marshal(&sc)
	if err != nil {
		return err
	}
	return os.WriteFile(scPath, out, 0644) // #nosec G306 - config file needs to be readable for site-builder
}

func containsTildePath(content string) bool {
	return strings.Contains(content, "~/") || strings.Contains(content, "~\\")
}

func expandTilde(value, home string) string {
	if strings.HasPrefix(value, "~/") || strings.HasPrefix(value, "~\\") {
		return filepath.Join(home, value[2:])
	}
	return value
}

func installHint(tool string) string {
	switch tool {
	case "hugo":
		switch runtime.GOOS {
		case "darwin":
			return "brew install hugo (installs the Extended build)"
		case "linux":
			return "Install Hugo Extended via your package manager or download from https://gohugo.io/installation/"
		case "windows":
			return "Download the Hugo Extended release ZIP from https://gohugo.io/installation/ and add it to PATH"
		default:
			return "Install Hugo Extended from https://gohugo.io/installation/"
		}
	default:
		if runtime.GOOS == "windows" {
			return "Run: walgo setup-deps (PowerShell) to install via suiup"
		}
		return "Run: walgo setup-deps (uses suiup)"
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().Bool("fix-paths", false, "Rewrite tildes in sites-config.yaml to absolute paths")
	doctorCmd.Flags().Bool("fix-all", false, "Automatically fix all detected issues")
	doctorCmd.Flags().BoolP("verbose", "v", false, "Show detailed output including versions and paths")
}
