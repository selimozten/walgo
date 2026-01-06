package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "walgo",
	Short: "Walgo ships static sites to Walrus (on-chain and HTTP paths).",
	Long: `Walgo provides a seamless bridge for Hugo users to build and deploy
static sites to Walrus decentralized storage.

What you can do:
• init/new/build/serve
• optimize HTML/CSS/JS
• On-chain: deploy, update, status, domain
• HTTP (Testnet): deploy-http to publisher and fetch via aggregator (no wallet)
• doctor: diagnose config, gas, and PATH issues
• setup: write sites-config.yaml; setup-deps: install site-builder/walrus

Quick Start:
  walgo init my-site
  cd my-site
  walgo build
  walgo launch    # Interactive deployment wizard (recommended)

Alternative deployment methods:
  walgo deploy-http   # HTTP deployment (no wallet, testnet only)
  walgo deploy        # Direct on-chain deployment (advanced)

Docs: https://github.com/selimozten/walgo`,
}

// Execute runs the root command.
func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.walgo.yaml or ./walgo.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".walgo")

		viper.AddConfigPath(".")
		viper.SetConfigName("walgo")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if cfgFile != "" {
			fmt.Fprintf(os.Stderr, "Error: Failed to read specified config file %s: %v\n", cfgFile, err)
		} else {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				fmt.Fprintf(os.Stderr, "Error: Found config file but failed to read/parse it: %v\n", err)
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
