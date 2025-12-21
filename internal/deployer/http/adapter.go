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

	"walgo/internal/deployer"
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
	// HTTP path has no status; return success stub
	return &deployer.Result{Success: true, ObjectID: objectID}, nil
}

// Quilt upload: single multipart request
func (a *Adapter) deployQuilt(ctx context.Context, siteDir, publisher string, epochs int) (*deployer.Result, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
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
		defer f.Close()
		if _, err := io.Copy(part, f); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
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

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
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

// parseQuiltResponse handles both v1 and v2 Walrus API response formats
func parseQuiltResponse(respBytes []byte) (string, map[string]string, error) {
	// Try v1 format first (current)
	var v1Resp struct {
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

	if err := json.Unmarshal(respBytes, &v1Resp); err == nil {
		quiltID := v1Resp.BlobStoreResult.NewlyCreated.BlobObject.BlobId
		if quiltID == "" {
			quiltID = v1Resp.BlobStoreResult.AlreadyCertified.BlobId
		}
		if quiltID != "" {
			patches := make(map[string]string, len(v1Resp.StoredQuiltBlobs))
			for _, p := range v1Resp.StoredQuiltBlobs {
				patches[p.Identifier] = p.QuiltPatchId
			}
			return quiltID, patches, nil
		}
	}

	// Try v2 format (potential new structure)
	var v2Resp struct {
		// v2 might have a flatter structure
		BlobId   string `json:"blobId"`
		QuiltId  string `json:"quiltId"`
		ObjectId string `json:"objectId"`
		Blobs    []struct {
			Path   string `json:"path"`
			BlobId string `json:"blobId"`
		} `json:"blobs"`
		// Also try nested result
		Result struct {
			BlobId  string `json:"blobId"`
			QuiltId string `json:"quiltId"`
		} `json:"result"`
	}

	if err := json.Unmarshal(respBytes, &v2Resp); err == nil {
		quiltID := v2Resp.BlobId
		if quiltID == "" {
			quiltID = v2Resp.QuiltId
		}
		if quiltID == "" {
			quiltID = v2Resp.ObjectId
		}
		if quiltID == "" {
			quiltID = v2Resp.Result.BlobId
		}
		if quiltID == "" {
			quiltID = v2Resp.Result.QuiltId
		}

		patches := make(map[string]string, len(v2Resp.Blobs))
		for _, b := range v2Resp.Blobs {
			patches[b.Path] = b.BlobId
		}

		if quiltID != "" {
			return quiltID, patches, nil
		}
	}

	// If both fail, try to extract any blob ID from the raw response
	var generic map[string]interface{}
	if err := json.Unmarshal(respBytes, &generic); err != nil {
		return "", nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	// Look for common ID field names recursively
	if id := findBlobID(generic); id != "" {
		return id, nil, nil
	}

	return "", nil, fmt.Errorf("could not extract blob ID from response")
}

// findBlobID recursively searches for a blob ID in a map
func findBlobID(m map[string]interface{}) string {
	idFields := []string{"blobId", "blob_id", "quiltId", "quilt_id", "objectId", "object_id", "id"}

	for _, field := range idFields {
		if val, ok := m[field]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}

	// Recurse into nested objects
	for _, val := range m {
		if nested, ok := val.(map[string]interface{}); ok {
			if id := findBlobID(nested); id != "" {
				return id
			}
		}
	}

	return ""
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

	endpointBase := strings.TrimRight(publisher, "/") + "/v1/blobs?epochs=" + fmt.Sprint(epochs)
	fileToBlob := make(map[string]string, len(files))
	var mu sync.Mutex
	jobs := make(chan job)
	wg := sync.WaitGroup{}

	workerFn := func() {
		defer wg.Done()
		for j := range jobs {
			blobID, err := uploadWithRetry(ctx, endpointBase, j.abs, maxRetries)
			if err == nil && blobID != "" {
				mu.Lock()
				fileToBlob[j.rel] = blobID
				mu.Unlock()
			}
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, f)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
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
