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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Walgo doctor")

		// Check binaries
		checkBin("hugo")
		checkBin("site-builder")
		checkBin("walrus")
		checkBin("sui")

		// Sui client: active env and address
		activeEnv := runQuiet("sui", "client", "envs")
		fmt.Print(activeEnv)
		address := strings.TrimSpace(runQuiet("sui", "client", "active-address"))
		if address != "" {
			fmt.Println("Active Sui address:", address)
		}
		gas := runQuiet("sui", "client", "gas")
		if strings.Contains(gas, "No gas coins are owned") {
			fmt.Println("⚠️  No SUI gas coins detected for the active address.")
			if strings.Contains(activeEnv, "testnet") {
				fmt.Printf("Get testnet tokens: https://faucet.sui.io/?address=%s\n", address)
			} else if strings.Contains(activeEnv, "devnet") {
				fmt.Println("Try: sui client faucet")
			}
		} else {
			fmt.Println("✓ SUI gas coins found (or gas query returned data).")
		}

		// Fix sites-config.yaml tildes if requested
		fixPaths, _ := cmd.Flags().GetBool("fix-paths")
		home, _ := os.UserHomeDir()
		scPath := filepath.Join(home, ".config", "walrus", "sites-config.yaml")
		if _, err := os.Stat(scPath); err == nil {
			if fixPaths {
				if err := ensureAbsolutePaths(scPath, home); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to fix paths in %s: %v\n", scPath, err)
				} else {
					fmt.Println("✓ Updated sites-config.yaml to use absolute paths")
				}
			} else {
				fmt.Println("Found:", scPath)
				fmt.Println("Tip: run 'walgo doctor --fix-paths' to expand any '~/' to absolute paths.")
			}
		} else {
			fmt.Println("ℹ️  sites-config.yaml not found; run 'walgo setup' to create it.")
		}

		fmt.Println("\nHTTP Testnet publish (no wallet needed):")
		fmt.Println("  walgo deploy-http --publisher https://publisher.walrus-testnet.walrus.space \\")
		fmt.Println("    --aggregator https://aggregator.walrus-testnet.walrus.space --epochs 1")
	},
}

func checkBin(name string) {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("✗ %s not found in PATH\n", name)
	} else {
		fmt.Printf("✓ %s found\n", name)
	}
}

func runQuiet(name string, args ...string) string {
	out, _ := exec.Command(name, args...).CombinedOutput()
	return string(out)
}

func ensureAbsolutePaths(scPath, home string) error {
	data, err := os.ReadFile(scPath)
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
	return os.WriteFile(scPath, out, 0644)
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().Bool("fix-paths", false, "Rewrite tildes in sites-config.yaml to absolute paths")
}
