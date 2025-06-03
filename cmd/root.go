package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "walgo",
	Short: "Walgo is a CLI tool to integrate Hugo with Walrus Sites.",
	Long: `Walgo provides a seamless bridge for Hugo users to build and deploy 
static sites to the Walrus decentralized storage protocol.

Key Features:
• Initialize Hugo sites pre-configured for Walrus deployment
• Build and serve Hugo sites locally  
• Deploy sites to Walrus decentralized storage
• Import content from Obsidian vaults
• Manage SuiNS domains for your sites
• Update existing sites efficiently

Quick Start:
  walgo init my-site      # Create a new Hugo site configured for Walrus
  cd my-site             # Navigate to your site directory
  walgo new posts/hello  # Create new content
  walgo build            # Build your site
  walgo serve            # Preview locally (optional)
  walgo deploy           # Deploy to Walrus Sites

Learn more: https://github.com/selimozten/walgo`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.walgo.yaml or ./walgo.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err) // Exits on error if home dir not found

		// Search config in home directory with name ".walgo" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml") // Important to set type before name if name has no extension
		viper.SetConfigName(".walgo")

		// Search config in current directory
		viper.AddConfigPath(".")
		// viper.SetConfigType("yaml") // Already set
		viper.SetConfigName("walgo") // Will look for walgo.yaml
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Attempt to read the configuration file.
	if err := viper.ReadInConfig(); err != nil {
		// If a specific config file was provided via --config flag, and it's not found or fails to load, this is a more direct error.
		if cfgFile != "" {
			// This branch handles errors when cfgFile is explicitly set.
			// It could be a viper.ConfigFileNotFoundError or some other read/parse error.
			fmt.Fprintf(os.Stderr, "Error: Failed to read specified config file %s: %v\n", cfgFile, err)
			// Depending on desired strictness, could os.Exit(1) here.
			// For now, we let commands that require config (via LoadConfig) handle the absence of a loaded config.
		} else {
			// If cfgFile was NOT set, and we encounter an error that is NOT ConfigFileNotFoundError,
			// it means one of the default locations might exist but is problematic (e.g., permission, malformed).
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				fmt.Fprintf(os.Stderr, "Error: Found config file but failed to read/parse it: %v\n", err)
				// viper.ConfigFileUsed() might give a clue here if Viper identified a file before failing to parse.
				// For example: fmt.Fprintf(os.Stderr, "Error reading config file %s: %v\n", viper.ConfigFileUsed(), err)
			}
			// If it IS a ConfigFileNotFoundError and cfgFile was not set, this is normal (e.g., before 'walgo init').
			// No message needed here; commands that need config will report it via LoadConfig.
		}
	} else {
		// If a config file is found and read successfully:
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
