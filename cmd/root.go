package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// readLine reads a line from the reader, trimming whitespace.
// Returns an error if stdin is closed or broken.
func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	if err == io.EOF && line == "" {
		return "", io.EOF
	}
	return strings.TrimSpace(line), nil
}

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

Docs: https://github.com/ganbitlabs/walgo`,
	// A runtime failure is not a usage mistake; printing the full flag list after
	// one buries the actual error message.
	SilenceUsage: true,
}

// Execute runs the root command and returns any error encountered.
// Cobra already reports the error, so it is only passed up for the exit code.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.walgo.yaml or ./walgo.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	viper.SetConfigType("yaml")

	if cfgFile == "" {
		found := findConfigFile()
		if found == "" {
			viper.AutomaticEnv()
			return // no config file is a normal situation
		}
		viper.SetConfigFile(found)
	} else {
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read config file %s: %v\n", viper.ConfigFileUsed(), err)
		return
	}

	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
}

// findConfigFile returns the first existing config file, preferring the working
// directory over the home directory. Candidates are matched by full filename
// rather than by basename, so an executable named "walgo" sitting next to the
// site is never mistaken for a config file.
func findConfigFile() string {
	candidates := []string{"walgo.yaml", "walgo.yml"}

	for _, name := range candidates {
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			return name
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	for _, name := range []string{".walgo.yaml", ".walgo.yml"} {
		path := filepath.Join(home, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}
