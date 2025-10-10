package http

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
		rel, _ := filepath.Rel(siteDir, path)
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
		return nil, err
	}
	writer.Close()

	endpoint := fmt.Sprintf("%s/v1/quilts?epochs=%d", strings.TrimRight(publisher, "/"), epochs)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("publisher responded %d: %s", resp.StatusCode, string(respBytes))
	}

	var qResp struct {
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
	}
	if err := json.Unmarshal(respBytes, &qResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	quiltID := qResp.BlobStoreResult.NewlyCreated.BlobObject.BlobId
	if quiltID == "" {
		quiltID = qResp.BlobStoreResult.AlreadyCertified.BlobId
	}
	return &deployer.Result{Success: true, ObjectID: quiltID}, nil
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
		rel, _ := filepath.Rel(siteDir, path)
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
		blobID, err := putBlob(ctx, endpoint, filePath)
		if err == nil && blobID != "" {
			return blobID, nil
		}
		// Retry on error or empty blobID
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

func putBlob(ctx context.Context, endpoint, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, f)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("publisher responded %d: %s", resp.StatusCode, string(b))
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
		return "", err
	}
	id := parsed.NewlyCreated.BlobObject.BlobId
	if id == "" {
		id = parsed.AlreadyCertified.BlobId
	}
	return id, nil
}
