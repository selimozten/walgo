package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"walgo/internal/config"
	"walgo/internal/hugo"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [site-name]",
	Short: "Initialize a new Hugo site with Walrus Sites configuration.",
	Long: `Initializes a new Hugo site in a directory specified by [site-name].
It sets up the basic Hugo structure and creates a walgo.yaml configuration 
file tailored for Walrus Sites deployment.`,
	Args: cobra.ExactArgs(1), // Expects exactly one argument: site-name
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]

		fmt.Printf("Initializing new Walgo site: %s\n", siteName)

		// Get current working directory to resolve absolute path for siteName
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		sitePath := filepath.Join(cwd, siteName)

		// 1. Create site directory
		// #nosec G301 - site directory needs standard permissions
		if err := os.MkdirAll(sitePath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating site directory %s: %v\n", sitePath, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully created directory: %s\n", sitePath)

		// 2. Initialize Hugo site
		if err := hugo.InitializeSite(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing Hugo site in %s: %v\n", sitePath, err)
			os.Exit(1)
		}
		fmt.Println("Hugo site initialized successfully.")

		// 3. Create Walrus configuration (walgo.yaml)
		if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating default walgo.yaml in %s: %v\n", sitePath, err)
			os.Exit(1)
		}
		fmt.Printf("Default walgo.yaml created in %s\n", sitePath)

		// 4. (Optional) Modify Hugo's config.toml for Walrus (or guide user)
		// For now, we'll just print a message.
		fmt.Println("\nNext steps:")
		fmt.Println("- Customize your Hugo site configuration (e.g., config.toml or hugo.toml in themes). A hugo.toml has been created.")
		fmt.Printf("- Review and update %s/walgo.yaml with your Walrus project details.\n", siteName)
		fmt.Printf("- cd %s\n", siteName)
		fmt.Println("- Start adding content and then run 'walgo build'.")
		fmt.Println("- Option A (HTTP): 'walgo deploy-http --publisher https://publisher.walrus-testnet.walrus.space --aggregator https://aggregator.walrus-testnet.walrus.space' (no funds required)")
		fmt.Println("- Option B (On-chain): 'walgo setup --network testnet' then 'walgo deploy' (requires funded Sui wallet)")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.
	// Example:
	// initCmd.Flags().StringP("theme", "t", "", "Hugo theme to use")
}
