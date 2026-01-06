package httpdeployer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/selimozten/walgo/internal/deployer"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

// Deploy supports two modes:
// - quilt: single multipart PUT to /v1/quilts
// - blobs: per-file PUTs to /v1/blobs using a worker pool with retries
func (a *Adapter) Deploy(ctx context.Context, siteDir string, opts deployer.DeployOptions) (*deployer.Result, error) {
	workers := opts.Workers
	if workers <= 0 {
		workers = 10
	}
	maxRetries := opts.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 5
	}

	if strings.ToLower(opts.Mode) == "blobs" {
		return a.deployBlobs(ctx, siteDir, opts.PublisherBaseURL, opts.Epochs, workers, maxRetries)
	}
	return a.deployQuilt(ctx, siteDir, opts.PublisherBaseURL, opts.Epochs)
}

func (a *Adapter) Update(ctx context.Context, siteDir string, objectID string, opts deployer.DeployOptions) (*deployer.Result, error) {
	// HTTP path does not have native update semantics; perform a fresh deploy
	return a.Deploy(ctx, siteDir, opts)
}

func (a *Adapter) Status(ctx context.Context, objectID string, opts deployer.DeployOptions) (*deployer.Result, error) {
	// HTTP deployments store blobs, not site objects
	// We can verify blob existence by checking the aggregator
	if opts.AggregatorBaseURL == "" {
		return nil, fmt.Errorf("status check requires AggregatorBaseURL to verify blob existence")
	}

	if objectID == "" {
		return &deployer.Result{Success: false, Message: "no object ID provided"}, nil
	}

	// Check if blob exists on aggregator
	endpoint := fmt.Sprintf("%s/v1/blobs/%s", strings.TrimRight(opts.AggregatorBaseURL, "/"), objectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &deployer.Result{
			Success:  false,
			ObjectID: objectID,
			Message:  fmt.Sprintf("failed to reach aggregator: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
		return &deployer.Result{
			Success:  true,
			ObjectID: objectID,
			Message:  "blob exists on aggregator",
		}, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return &deployer.Result{
			Success:  false,
			ObjectID: objectID,
			Message:  "blob not found on aggregator",
		}, nil
	}

	return &deployer.Result{
		Success:  false,
		ObjectID: objectID,
		Message:  fmt.Sprintf("aggregator returned status %d", resp.StatusCode),
	}, nil
}

func (a *Adapter) Destroy(ctx context.Context, objectID string) error {
	// HTTP deployment path does not support site destruction
	// Files uploaded via HTTP cannot be deleted through the API
	return fmt.Errorf("destroy operation not supported for HTTP deployment mode - files must be managed manually")
}

// calculateUploadTimeout returns a dynamic timeout based on body size
// Minimum 30 seconds, scales with size (1MB = 30s additional), maximum 10 minutes
func calculateUploadTimeout(bodySize int) time.Duration {
	const (
		minTimeout     = 30 * time.Second
		maxTimeout     = 10 * time.Minute
		bytesPerSecond = 100 * 1024 // Assume 100KB/s as conservative upload speed
	)

	// Calculate time needed at conservative speed, plus buffer
	estimatedTime := time.Duration(bodySize/bytesPerSecond) * time.Second
	timeout := minTimeout + estimatedTime

	if timeout > maxTimeout {
		return maxTimeout
	}
	return timeout
}

// Quilt upload: single multipart request
func (a *Adapter) deployQuilt(ctx context.Context, siteDir, publisher string, epochs int) (*deployer.Result, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	var fileCount int

	// Walk files and add to multipart
	err := filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(siteDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		field := strings.ReplaceAll(rel, string(os.PathSeparator), "__")
		field = strings.ReplaceAll(field, " ", "_")

		part, err := writer.CreateFormFile(field, filepath.Base(path))
		if err != nil {
			return err
		}
		// #nosec G304 - path comes from filepath.Walk which is already constrained to siteDir
		f, err := os.Open(filepath.Clean(path))
		if err != nil {
			return err
		}
		// Close immediately to prevent FD leak in Walk loop
		if _, copyErr := io.Copy(part, f); copyErr != nil {
			f.Close()
			return copyErr
		}
		if closeErr := f.Close(); closeErr != nil {
			return fmt.Errorf("failed to close file %s: %w", path, closeErr)
		}
		fileCount++
		return nil
	})
	if err != nil {
		return nil, err
	}

	if fileCount == 0 {
		return nil, fmt.Errorf("no files found in directory: %s", siteDir)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1/quilts?epochs=%d", strings.TrimRight(publisher, "/"), epochs)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Use dynamic timeout based on body size
	timeout := calculateUploadTimeout(body.Len())
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			return nil, fmt.Errorf("upload timed out after %v (body size: %s) - try deploying with 'blobs' mode for large sites",
				timeout, formatBytes(int64(body.Len())))
		}
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("publisher responded %d: %s", resp.StatusCode, string(respBytes))
	}

	// Try to parse the response - handle both v1 and v2 API response formats
	quiltID, patches, err := parseQuiltResponse(respBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nRaw response: %s", err, string(respBytes))
	}

	return &deployer.Result{Success: true, ObjectID: quiltID, QuiltPatches: patches}, nil
}

// parseQuiltResponse extracts blob ID and patches from Walrus API response
func parseQuiltResponse(respBytes []byte) (string, map[string]string, error) {
	var resp struct {
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

	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return "", nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	quiltID := resp.BlobStoreResult.NewlyCreated.BlobObject.BlobId
	if quiltID == "" {
		quiltID = resp.BlobStoreResult.AlreadyCertified.BlobId
	}

	if quiltID == "" {
		return "", nil, fmt.Errorf("no blob ID found in response")
	}

	patches := make(map[string]string, len(resp.StoredQuiltBlobs))
	for _, p := range resp.StoredQuiltBlobs {
		patches[p.Identifier] = p.QuiltPatchId
	}

	return quiltID, patches, nil
}

// uploadError tracks a failed file upload
type uploadError struct {
	file string
	err  error
}

// Blobs upload: concurrent workers with exponential backoff
func (a *Adapter) deployBlobs(ctx context.Context, siteDir, publisher string, epochs, workers, maxRetries int) (*deployer.Result, error) {
	type job struct{ rel, abs string }
	files := make([]job, 0, 128)
	if err := filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(siteDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		files = append(files, job{rel: rel, abs: path})
		return nil
	}); err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in directory: %s", siteDir)
	}

	endpointBase := strings.TrimRight(publisher, "/") + "/v1/blobs?epochs=" + fmt.Sprint(epochs)
	fileToBlob := make(map[string]string, len(files))
	var uploadErrors []uploadError
	var mu sync.Mutex
	jobs := make(chan job)
	wg := sync.WaitGroup{}

	workerFn := func() {
		defer wg.Done()
		for j := range jobs {
			blobID, err := uploadWithRetry(ctx, endpointBase, j.abs, maxRetries)
			mu.Lock()
			if err != nil {
				uploadErrors = append(uploadErrors, uploadError{file: j.rel, err: err})
			} else if blobID == "" {
				uploadErrors = append(uploadErrors, uploadError{file: j.rel, err: fmt.Errorf("empty blob ID returned")})
			} else {
				fileToBlob[j.rel] = blobID
			}
			mu.Unlock()
		}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go workerFn()
	}
send:
	for _, f := range files {
		select {
		case jobs <- f:
		case <-ctx.Done():
			break send
		}
	}
	close(jobs)
	wg.Wait()

	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, fmt.Errorf("deployment cancelled: %w", ctx.Err())
	}

	// Check for upload failures
	mu.Lock()
	defer mu.Unlock()

	if len(uploadErrors) > 0 {
		// Build error message with failed files
		var errMsgs []string
		for _, ue := range uploadErrors {
			errMsgs = append(errMsgs, fmt.Sprintf("  - %s: %v", ue.file, ue.err))
		}
		return nil, fmt.Errorf("failed to upload %d of %d files:\n%s",
			len(uploadErrors), len(files), strings.Join(errMsgs, "\n"))
	}

	return &deployer.Result{Success: true, FileToBlobID: fileToBlob}, nil
}

func uploadWithRetry(ctx context.Context, endpoint, filePath string, maxRetries int) (string, error) {
	backoff := 250 * time.Millisecond
	for attempt := 0; attempt <= maxRetries; attempt++ {
		blobID, status, err := putBlob(ctx, endpoint, filePath)
		if err == nil && blobID != "" {
			return blobID, nil
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if err == nil && (status < 500 && status != 429) {
			return "", fmt.Errorf("non-retryable status %d", status)
		}
		select {
		case <-time.After(backoff + time.Duration(attempt)*50*time.Millisecond):
			backoff *= 2
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return "", fmt.Errorf("max retries reached for %s", filepath.Base(filePath))
}

func putBlob(ctx context.Context, endpoint, filePath string) (string, int, error) {
	// #nosec G304 - filePath comes from filepath.Walk which is already constrained to siteDir
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	// Get file size for dynamic timeout
	stat, err := f.Stat()
	if err != nil {
		return "", 0, fmt.Errorf("failed to stat file: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, f)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.ContentLength = stat.Size()

	// Use dynamic timeout based on file size
	timeout := calculateUploadTimeout(int(stat.Size()))
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", resp.StatusCode, fmt.Errorf("publisher responded %d: %s", resp.StatusCode, string(b))
	}
	var parsed struct {
		NewlyCreated struct {
			BlobObject struct {
				BlobId string `json:"blobId"`
			} `json:"blobObject"`
		} `json:"newlyCreated"`
		AlreadyCertified struct {
			BlobId string `json:"blobId"`
		} `json:"alreadyCertified"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", resp.StatusCode, err
	}
	id := parsed.NewlyCreated.BlobObject.BlobId
	if id == "" {
		id = parsed.AlreadyCertified.BlobId
	}
	return id, resp.StatusCode, nil
}

// formatBytes converts byte count to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
