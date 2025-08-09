package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"walgo/internal/config"

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

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		err = filepath.Walk(publishDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(publishDir, path)
			field := strings.ReplaceAll(rel, string(os.PathSeparator), "__")
			field = strings.ReplaceAll(field, " ", "_")

			part, err := writer.CreateFormFile(field, filepath.Base(path))
			if err != nil {
				return err
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(part, f); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error preparing upload: %v\n", err)
			os.Exit(1)
		}
		writer.Close()

		endpoint := fmt.Sprintf("%s/v1/quilts?epochs=%d", strings.TrimRight(publisher, "/"), epochs)
		req, err := http.NewRequest(http.MethodPut, endpoint, &body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
			os.Exit(1)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "HTTP error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		respBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Fprintf(os.Stderr, "Publisher responded %d: %s\n", resp.StatusCode, string(respBytes))
			os.Exit(1)
		}

		var qResp quiltUploadResponse
		if err := json.Unmarshal(respBytes, &qResp); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse response: %v\nBody: %s\n", err, string(respBytes))
			os.Exit(1)
		}

		quiltID := qResp.BlobStoreResult.NewlyCreated.BlobObject.BlobId
		if quiltID == "" {
			quiltID = qResp.BlobStoreResult.AlreadyCertified.BlobId
		}

		fmt.Println("\n✅ HTTP deploy complete!")
		if quiltID != "" {
			fmt.Printf("quiltId: %s\n", quiltID)
		}
		if len(qResp.StoredQuiltBlobs) > 0 {
			fmt.Println("patchIds:")
			for _, e := range qResp.StoredQuiltBlobs {
				fmt.Printf("  %s: %s\n", e.Identifier, e.QuiltPatchId)
			}
		}

		fmt.Println("\nFetch examples:")
		fmt.Printf("  Aggregator: %s/v1/blobs/by-quilt-patch-id/<patchId>\n", strings.TrimRight(aggregator, "/"))
		fmt.Println("  (Use a patchId printed above to fetch a file)")
	},
}

func init() {
	rootCmd.AddCommand(deployHTTPCmd)
	deployHTTPCmd.Flags().String("publisher", "", "Walrus publisher base URL (e.g., https://publisher.walrus-testnet.walrus.space)")
	deployHTTPCmd.Flags().String("aggregator", "", "Walrus aggregator base URL (e.g., https://aggregator.walrus-testnet.walrus.space)")
	deployHTTPCmd.Flags().IntP("epochs", "e", 1, "Number of epochs to store the quilt")
}
