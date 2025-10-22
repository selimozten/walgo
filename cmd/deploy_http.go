package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"walgo/internal/config"
	"walgo/internal/deployer"
	httpdep "walgo/internal/deployer/http"

	"github.com/spf13/cobra"
)

var deployHTTPCmd = &cobra.Command{
	Use:   "deploy-http",
	Short: "Deploy site via Walrus HTTP APIs (publisher/aggregator) on testnet.",
	Long: `Uploads the Hugo publish directory as a Quilt to a Walrus publisher and prints
the resulting quiltId and patchIds. This path does not require on-chain funds.

Example:
  walgo deploy-http --publisher https://publisher.walrus-testnet.walrus.space \
    --aggregator https://aggregator.walrus-testnet.walrus.space --epochs 1`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		publisher, err := cmd.Flags().GetString("publisher")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading publisher flag:", err)
			os.Exit(1)
		}
		aggregator, err := cmd.Flags().GetString("aggregator")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading aggregator flag:", err)
			os.Exit(1)
		}
		epochs, err := cmd.Flags().GetInt("epochs")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading epochs flag:", err)
			os.Exit(1)
		}
		mode, err := cmd.Flags().GetString("mode")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading mode flag:", err)
			os.Exit(1)
		}
		workers, err := cmd.Flags().GetInt("workers")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading workers flag:", err)
			os.Exit(1)
		}
		retries, err := cmd.Flags().GetInt("retries")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading retries flag:", err)
			os.Exit(1)
		}
		jsonLogs, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading json flag:", err)
			os.Exit(1)
		}
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading verbose flag:", err)
			os.Exit(1)
		}

		if publisher == "" || aggregator == "" {
			fmt.Fprintln(os.Stderr, "--publisher and --aggregator are required")
			os.Exit(1)
		}
		if epochs <= 0 {
			epochs = 1
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting cwd: %v\n", err)
			os.Exit(1)
		}
		publishDir := filepath.Join(sitePath, cfg.HugoConfig.PublishDir)
		if _, err := os.Stat(publishDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Publish directory '%s' not found. Run 'walgo build' first.\n", publishDir)
			os.Exit(1)
		}

		// Use HTTP deployer; default to quilt (single request)
		d := httpdep.New()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		if mode == "" {
			mode = "quilt"
		}
		if workers <= 0 {
			workers = 10
		}
		if retries <= 0 {
			retries = 5
		}

		res, err := d.Deploy(ctx, publishDir, deployer.DeployOptions{
			Epochs:            epochs,
			PublisherBaseURL:  publisher,
			AggregatorBaseURL: aggregator,
			Mode:              mode,
			Workers:           workers,
			MaxRetries:        retries,
			JSONLogs:          jsonLogs,
			Verbose:           verbose,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "HTTP deploy failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✅ HTTP deploy complete!")
		if res.ObjectID != "" {
			fmt.Printf("📦 Quilt ID: %s\n", res.ObjectID)
		}

		// Show per-file info and aggregator fetch hints when available
		if len(res.QuiltPatches) > 0 {
			fmt.Println()
			fmt.Println("📂 Files stored on Walrus:")
			for ident := range res.QuiltPatches {
				display := ident
				display = strings.ReplaceAll(display, "__", "/")
				display = strings.ReplaceAll(display, "_", " ")
				fmt.Printf("  ✓ %s\n", display)
			}

			fmt.Println()
			fmt.Println("📥 To fetch individual files (for testing):")
			for ident, patch := range res.QuiltPatches {
				display := ident
				display = strings.ReplaceAll(display, "__", "/")
				url := fmt.Sprintf("%s/v1/blobs/by-quilt-patch-id/%s", strings.TrimRight(aggregator, "/"), patch)
				fmt.Printf("    curl %s > %s\n", url, display)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(deployHTTPCmd)
	deployHTTPCmd.Flags().String("publisher", "", "Walrus publisher base URL (e.g., https://publisher.walrus-testnet.walrus.space)")
	deployHTTPCmd.Flags().String("aggregator", "", "Walrus aggregator base URL (e.g., https://aggregator.walrus-testnet.walrus.space)")
	deployHTTPCmd.Flags().IntP("epochs", "e", 1, "Number of epochs to store the quilt")
	deployHTTPCmd.Flags().String("mode", "quilt", "HTTP deploy mode: quilt or blobs")
	deployHTTPCmd.Flags().Int("workers", 10, "Concurrent workers for blobs mode")
	deployHTTPCmd.Flags().Int("retries", 5, "Max retries per file for transient errors")
	deployHTTPCmd.Flags().Bool("json", false, "Emit structured JSON logs")
	deployHTTPCmd.Flags().BoolP("verbose", "v", false, "Verbose logging")
}
