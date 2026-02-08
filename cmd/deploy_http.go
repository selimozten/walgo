package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployer"
	httpdep "github.com/selimozten/walgo/internal/deployer/http"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

var deployHTTPCmd = &cobra.Command{
	Use:   "deploy-http",
	Short: "Deploy site via Walrus HTTP APIs (publisher/aggregator) on testnet.",
	Long: `Uploads the Hugo publish directory as a Quilt to a Walrus publisher and prints
the resulting quiltId and patchIds. This path does not require on-chain funds.

Choose from available publishers and aggregators:
  https://docs.wal.app/docs/usage/web-api#public-services

Example (Testnet):
  walgo deploy-http --publisher https://publisher.walrus-testnet.walrus.space \
    --aggregator https://aggregator.walrus-testnet.walrus.space --epochs 1

Example (Mainnet):
  walgo deploy-http --publisher https://walrus-mainnet-publisher-1.staketab.org:443 \
    --aggregator https://aggregator.walrus-mainnet.walrus.space --epochs 1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		var err error
		publisher, err := cmd.Flags().GetString("publisher")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading publisher flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading publisher flag: %w", err)
		}
		aggregator, err := cmd.Flags().GetString("aggregator")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading aggregator flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading aggregator flag: %w", err)
		}
		epochs, err := cmd.Flags().GetInt("epochs")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading epochs flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading epochs flag: %w", err)
		}
		mode, err := cmd.Flags().GetString("mode")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading mode flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading mode flag: %w", err)
		}
		workers, err := cmd.Flags().GetInt("workers")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading workers flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading workers flag: %w", err)
		}
		retries, err := cmd.Flags().GetInt("retries")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading retries flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading retries flag: %w", err)
		}
		jsonLogs, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading json flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading json flag: %w", err)
		}
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: reading verbose flag: %v\n", icons.Error, err)
			return fmt.Errorf("error reading verbose flag: %w", err)
		}

		if publisher == "" || aggregator == "" {
			fmt.Fprintf(os.Stderr, "%s Error: --publisher and --aggregator are required\n", icons.Error)
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintf(os.Stderr, "%s Choose from available endpoints:\n", icons.Info)
			fmt.Fprintln(os.Stderr, "  https://docs.wal.app/docs/usage/web-api#public-services")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintf(os.Stderr, "%s Testnet example:\n", icons.Lightbulb)
			fmt.Fprintln(os.Stderr, "  --publisher https://publisher.walrus-testnet.walrus.space \\")
			fmt.Fprintln(os.Stderr, "  --aggregator https://aggregator.walrus-testnet.walrus.space")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintf(os.Stderr, "%s Mainnet example:\n", icons.Lightbulb)
			fmt.Fprintln(os.Stderr, "  --publisher https://walrus-mainnet-publisher-1.staketab.org:443 \\")
			fmt.Fprintln(os.Stderr, "  --aggregator https://aggregator.walrus-mainnet.walrus.space")
			return fmt.Errorf("--publisher and --aggregator are required")
		}
		if epochs <= 0 {
			epochs = 1
		}

		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
			return fmt.Errorf("error getting cwd: %w", err)
		}

		cfg, err := config.LoadConfigFrom(sitePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("error loading config: %w", err)
		}

		publishDir := filepath.Join(sitePath, cfg.HugoConfig.PublishDir)
		if _, err := os.Stat(publishDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s Error: Publish directory '%s' not found\n", icons.Error, publishDir)
			fmt.Fprintf(os.Stderr, "%s Run 'walgo build' first\n", icons.Lightbulb)
			return fmt.Errorf("publish directory not found: %s", publishDir)
		}

		err = hugo.BuildSite(sitePath)
		if err != nil {
			return fmt.Errorf("failed to build site: %w", err)
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
		if workers > 50 {
			workers = 50 // Cap to prevent resource exhaustion
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
			fmt.Fprintf(os.Stderr, "%s Error: HTTP deploy failed: %v\n", icons.Error, err)
			return fmt.Errorf("HTTP deploy failed: %w", err)
		}

		fmt.Printf("\n%s HTTP deploy complete!\n", icons.Success)
		if res.ObjectID != "" {
			fmt.Printf("%s Quilt ID: %s\n", icons.Package, res.ObjectID)
		}

		// Show per-file info and aggregator fetch hints when available
		if len(res.QuiltPatches) > 0 {
			fmt.Printf("\n%s Files stored on Walrus:\n", icons.Folder)
			for ident := range res.QuiltPatches {
				display := ident
				display = strings.ReplaceAll(display, "__", "/")
				display = strings.ReplaceAll(display, "_", " ")
				fmt.Printf("  %s %s\n", icons.Check, display)
			}

			fmt.Printf("\n%s To fetch individual files (for testing):\n", icons.Download)
			for ident, patch := range res.QuiltPatches {
				display := ident
				display = strings.ReplaceAll(display, "__", "/")
				url := fmt.Sprintf("%s/v1/blobs/by-quilt-patch-id/%s", strings.TrimRight(aggregator, "/"), patch)
				fmt.Printf("  %s curl %s > %s\n", icons.Download, url, display)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployHTTPCmd)
	deployHTTPCmd.Flags().String("publisher", "", "Walrus publisher base URL (see https://docs.wal.app/docs/usage/web-api#public-services)")
	deployHTTPCmd.Flags().String("aggregator", "", "Walrus aggregator base URL (see https://docs.wal.app/docs/usage/web-api#public-services)")
	deployHTTPCmd.Flags().IntP("epochs", "e", 1, "Number of epochs to store the quilt")
	deployHTTPCmd.Flags().String("mode", "quilt", "HTTP deploy mode: quilt or blobs")
	deployHTTPCmd.Flags().Int("workers", 10, "Concurrent workers for blobs mode")
	deployHTTPCmd.Flags().Int("retries", 5, "Max retries per file for transient errors")
	deployHTTPCmd.Flags().Bool("json", false, "Emit structured JSON logs")
	deployHTTPCmd.Flags().BoolP("verbose", "v", false, "Verbose logging")
}
