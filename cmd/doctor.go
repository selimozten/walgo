package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
- Wallet gas balance
- Configuration files
- Provides auto-fix suggestions

Examples:
  walgo doctor              # Run diagnostics
  walgo doctor --fix-paths  # Fix tilde paths in config
  walgo doctor --fix-all    # Auto-fix all issues`,
	Run: func(cmd *cobra.Command, args []string) {
		fixAll, err := cmd.Flags().GetBool("fix-all")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting fix-all flag: %v\n", err)
			return
		}
		fixPaths, err := cmd.Flags().GetBool("fix-paths")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting fix-paths flag: %v\n", err)
			return
		}
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting verbose flag: %v\n", err)
			return
		}

		fmt.Println("╔═══════════════════════════════════════════════════════════╗")
		fmt.Println("║                     Walgo Doctor                          ║")
		fmt.Println("║             Environment Diagnostics                       ║")
		fmt.Println("╚═══════════════════════════════════════════════════════════╝")
		fmt.Println()

		issues := 0
		warnings := 0

		// Check binaries
		fmt.Println("📦 Checking dependencies...")
		fmt.Println()

		binaries := map[string]struct {
			name     string
			required bool
			purpose  string
			install  string
		}{
			"hugo":         {"hugo", true, "Static site generation", "brew install hugo (macOS) or https://gohugo.io/installation/"},
			"site-builder": {"site-builder", false, "On-chain deployment", "walgo setup-deps --with-site-builder"},
			"walrus":       {"walrus", false, "Walrus CLI operations", "walgo setup-deps --with-walrus"},
			"sui":          {"sui", false, "On-chain wallet management", "https://docs.sui.io/guides/developer/getting-started/sui-install"},
		}

		for _, bin := range []string{"hugo", "site-builder", "walrus", "sui"} {
			info := binaries[bin]
			if path, err := exec.LookPath(bin); err != nil {
				if info.required {
					fmt.Printf("  ✗ %s not found (REQUIRED)\n", info.name)
					fmt.Printf("    Purpose: %s\n", info.purpose)
					fmt.Printf("    Install: %s\n", info.install)
					issues++
				} else {
					fmt.Printf("  ⚠ %s not found (optional for %s)\n", info.name, info.purpose)
					fmt.Printf("    Install: %s\n", info.install)
					warnings++
				}
			} else {
				fmt.Printf("  ✓ %s found", info.name)
				if verbose {
					fmt.Printf(" at %s", path)
				}
				fmt.Println()

				// Get version info if available
				if verbose {
					switch bin {
					case "hugo":
						version := strings.TrimSpace(runQuiet("hugo", "version"))
						if version != "" {
							fmt.Printf("    Version: %s\n", version)
						}
					case "sui":
						version := strings.TrimSpace(runQuiet("sui", "--version"))
						if version != "" {
							fmt.Printf("    Version: %s\n", version)
						}
					}
				}
			}
		}

		fmt.Println()

		// Check Sui environment if sui is available
		if _, err := exec.LookPath("sui"); err == nil {
			fmt.Println("🔐 Checking Sui configuration...")
			fmt.Println()

			activeEnv := runQuiet("sui", "client", "active-env")
			activeEnv = strings.TrimSpace(activeEnv)
			if activeEnv != "" {
				fmt.Printf("  ✓ Active network: %s\n", activeEnv)
			}

			address := strings.TrimSpace(runQuiet("sui", "client", "active-address"))
			if address != "" {
				fmt.Printf("  ✓ Active address: %s\n", address)

				// Check gas balance
				gas := runQuiet("sui", "client", "gas")
				if strings.Contains(gas, "No gas coins are owned") || strings.Contains(gas, "Error") {
					fmt.Println("  ✗ No SUI gas coins found")
					fmt.Println("    On-chain deployment requires SUI for gas fees")
					if strings.Contains(activeEnv, "testnet") {
						fmt.Printf("    Get testnet tokens: https://faucet.sui.io/?address=%s\n", address)
					} else if strings.Contains(activeEnv, "devnet") {
						fmt.Println("    Get devnet tokens: sui client faucet")
					}
					issues++
				} else {
					fmt.Println("  ✓ SUI gas coins available")
					if verbose {
						fmt.Printf("    %s\n", strings.TrimSpace(gas))
					}
				}
			} else {
				fmt.Println("  ✗ No active Sui address configured")
				fmt.Println("    Run: sui client")
				issues++
			}

			fmt.Println()
		}

		// Check configuration files
		fmt.Println("⚙️  Checking configuration files...")
		fmt.Println()

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			return
		}
		scPath := filepath.Join(home, ".config", "walrus", "sites-config.yaml")

		if _, err := os.Stat(scPath); err == nil {
			fmt.Printf("  ✓ sites-config.yaml found at %s\n", scPath)

			// Check for tilde paths
			data, err := os.ReadFile(scPath) // #nosec G304 - path is constructed from known directory
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not read sites-config.yaml: %v\n", err)
			} else {
				if strings.Contains(string(data), "~/") {
					fmt.Println("  ⚠ Configuration contains tilde paths (~)")
					fmt.Println("    Run: walgo doctor --fix-paths")
					warnings++

					if fixPaths || fixAll {
						if err := ensureAbsolutePaths(scPath, home); err != nil {
							fmt.Printf("  ✗ Failed to fix paths: %v\n", err)
							issues++
						} else {
							fmt.Println("  ✓ Fixed tilde paths to absolute paths")
						}
					}
				}
			}
		} else {
			fmt.Println("  ⚠ sites-config.yaml not found")
			fmt.Println("    For on-chain deployment, run: walgo setup --network testnet --force")
			warnings++
		}

		// Check for walgo.yaml in current directory
		if _, err := os.Stat("walgo.yaml"); err == nil {
			fmt.Println("  ✓ walgo.yaml found in current directory")
		} else {
			fmt.Println("  ⚠ walgo.yaml not found in current directory")
			fmt.Println("    Initialize a site: walgo init my-site")
			warnings++
		}

		fmt.Println()

		// Show deployment options
		fmt.Println("🚀 Deployment options:")
		fmt.Println()
		fmt.Println("  Option 1: HTTP Testnet (No wallet required)")
		fmt.Println("    walgo deploy-http \\")
		fmt.Println("      --publisher https://publisher.walrus-testnet.walrus.space \\")
		fmt.Println("      --aggregator https://aggregator.walrus-testnet.walrus.space \\")
		fmt.Println("      --epochs 1")
		fmt.Println()
		fmt.Println("  Option 2: On-chain (Requires wallet and SUI)")
		fmt.Println("    walgo setup --network testnet --force")
		fmt.Println("    walgo doctor --fix-paths")
		fmt.Println("    walgo deploy --epochs 5")
		fmt.Println()

		// Summary
		fmt.Println("═══════════════════════════════════════════════════════════")
		if issues == 0 && warnings == 0 {
			fmt.Println("✅ All checks passed! Your environment is ready.")
		} else {
			fmt.Printf("Summary: %d issue(s), %d warning(s)\n", issues, warnings)
			if issues > 0 {
				fmt.Println("\n❌ Please fix the issues above before deploying on-chain.")
			}
			if warnings > 0 {
				fmt.Println("\n⚠️  Warnings indicate optional features that may not work.")
			}
			if !fixAll && warnings > 0 {
				fmt.Println("\n💡 Tip: Run 'walgo doctor --fix-all' to auto-fix some issues")
			}
		}
		fmt.Println("═══════════════════════════════════════════════════════════")
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
		if strings.HasPrefix(ctx.General.Wallet, "~/") {
			ctx.General.Wallet = filepath.Join(home, ctx.General.Wallet[2:])
		}
		if strings.HasPrefix(ctx.General.WalrusConfig, "~/") {
			ctx.General.WalrusConfig = filepath.Join(home, ctx.General.WalrusConfig[2:])
		}
		sc.Contexts[k] = ctx
	}
	out, err := yaml.Marshal(&sc)
	if err != nil {
		return err
	}
	return os.WriteFile(scPath, out, 0644) // #nosec G306 - config file needs to be readable for site-builder
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().Bool("fix-paths", false, "Rewrite tildes in sites-config.yaml to absolute paths")
	doctorCmd.Flags().Bool("fix-all", false, "Automatically fix all detected issues")
	doctorCmd.Flags().BoolP("verbose", "v", false, "Show detailed output including versions and paths")
}
