package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"walgo/internal/config"
	"walgo/internal/deployer"
	httpdep "walgo/internal/deployer/http"

	"github.com/spf13/cobra"
)

type quiltUploadResponse struct {
	BlobStoreResult struct {
		NewlyCreated struct {
			BlobObject struct {
				BlobId string `json:"blobId"`
			} `json:"blobObject"`
		} `json:"newlyCreated"`
		AlreadyCertified struct {
			BlobId string `json:"blobId"`
		} `json:"alreadyCertified"`
	} `json:"blobStoreResult"`
	StoredQuiltBlobs []struct {
		Identifier   string `json:"identifier"`
		QuiltPatchId string `json:"quiltPatchId"`
	} `json:"storedQuiltBlobs"`
}

var deployHTTPCmd = &cobra.Command{
	Use:   "deploy-http",
	Short: "Deploy site via Walrus HTTP APIs (publisher/aggregator) on testnet.",
	Long: `Uploads the Hugo publish directory as a Quilt to a Walrus publisher and prints
the resulting quiltId and patchIds. This path does not require on-chain funds.

Example:
  walgo deploy-http --publisher https://publisher.walrus-testnet.walrus.space \
    --aggregator https://aggregator.walrus-testnet.walrus.space --epochs 1`,
	Run: func(cmd *cobra.Command, args []string) {
		publisher, _ := cmd.Flags().GetString("publisher")
		aggregator, _ := cmd.Flags().GetString("aggregator")
		epochs, _ := cmd.Flags().GetInt("epochs")

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

		// Use new HTTP deployer with quilt (single request) by default
		d := httpdep.New()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		res, err := d.Deploy(ctx, publishDir, deployer.DeployOptions{
			Epochs:            epochs,
			PublisherBaseURL:  publisher,
			AggregatorBaseURL: aggregator,
			Mode:              "quilt",
			Workers:           10,
			MaxRetries:        5,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "HTTP deploy failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✅ HTTP deploy complete!")
		if res.ObjectID != "" {
			fmt.Printf("📦 Quilt ID: %s\n", res.ObjectID)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployHTTPCmd)
	deployHTTPCmd.Flags().String("publisher", "", "Walrus publisher base URL (e.g., https://publisher.walrus-testnet.walrus.space)")
	deployHTTPCmd.Flags().String("aggregator", "", "Walrus aggregator base URL (e.g., https://aggregator.walrus-testnet.walrus.space)")
	deployHTTPCmd.Flags().IntP("epochs", "e", 1, "Number of epochs to store the quilt")
}
