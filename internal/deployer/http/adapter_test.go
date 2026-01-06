package httpdeployer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/selimozten/walgo/internal/deployer"
)

// TestDeployBlobs_FailsWhenAllUploadsFail verifies that deployment fails when all uploads fail.
// This was a critical bug - previously, failed uploads were silently ignored.
func TestDeployBlobs_FailsWhenAllUploadsFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return 500 Internal Server Error
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("server error")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.html"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	a := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := a.Deploy(ctx, dir, deployer.DeployOptions{
		PublisherBaseURL: srv.URL,
		Mode:             "blobs",
		Workers:          1,
		MaxRetries:       1,
		Epochs:           1,
	})

	// This MUST fail - uploads failed
	if err == nil {
		t.Fatal("Expected error when all uploads fail, but got success")
	}

	// Error should mention the failed file
	if !strings.Contains(err.Error(), "test.html") {
		t.Errorf("Error should mention failed file, got: %v", err)
	}
}

// TestDeployBlobs_FailsWhenSomeUploadsFail verifies partial failures are reported.
func TestDeployBlobs_FailsWhenSomeUploadsFail(t *testing.T) {
	failedFiles := map[string]bool{"fail-me.html": true}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()

		// Fail specific files
		content := string(body)
		for failFile := range failedFiles {
			if strings.Contains(content, failFile) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// Succeed others
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"newlyCreated": map[string]any{
				"blobObject": map[string]any{"blobId": "blob-123"},
			},
		}); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "success.html"), []byte("works"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fail-me.html"), []byte("fail-me.html"), 0644); err != nil {
		t.Fatal(err)
	}

	a := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := a.Deploy(ctx, dir, deployer.DeployOptions{
		PublisherBaseURL: srv.URL,
		Mode:             "blobs",
		Workers:          1,
		MaxRetries:       1,
		Epochs:           1,
	})

	if err == nil {
		t.Fatal("Expected error when some uploads fail")
	}

	if !strings.Contains(err.Error(), "fail-me.html") {
		t.Errorf("Error should mention the failed file, got: %v", err)
	}
}

// TestDeployBlobs_EmptyDirectory fails for empty directories.
func TestDeployBlobs_EmptyDirectory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not be called for empty directory")
	}))
	defer srv.Close()

	dir := t.TempDir() // Empty directory

	a := New()
	_, err := a.Deploy(context.Background(), dir, deployer.DeployOptions{
		PublisherBaseURL: srv.URL,
		Mode:             "blobs",
		Epochs:           1,
	})

	if err == nil {
		t.Fatal("Expected error for empty directory")
	}

	if !strings.Contains(err.Error(), "no files found") {
		t.Errorf("Error should mention no files found, got: %v", err)
	}
}

// TestStatus_RequiresAggregatorURL verifies Status requires aggregator URL.
func TestStatus_RequiresAggregatorURL(t *testing.T) {
	a := New()
	_, err := a.Status(context.Background(), "some-id", deployer.DeployOptions{})

	if err == nil {
		t.Fatal("Expected error when AggregatorBaseURL is not set")
	}

	if !strings.Contains(err.Error(), "AggregatorBaseURL") {
		t.Errorf("Error should mention AggregatorBaseURL, got: %v", err)
	}
}

// TestStatus_ReturnsFalseForNonexistent verifies Status returns false for missing blobs.
func TestStatus_ReturnsFalseForNonexistent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && strings.Contains(r.URL.Path, "/v1/blobs/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	a := New()
	result, err := a.Status(context.Background(), "nonexistent-blob-id", deployer.DeployOptions{
		AggregatorBaseURL: srv.URL,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Success {
		t.Fatal("Status should return false for nonexistent blob")
	}

	if !strings.Contains(result.Message, "not found") {
		t.Errorf("Message should indicate not found, got: %s", result.Message)
	}
}

// TestStatus_ReturnsTrueForExisting verifies Status returns true for existing blobs.
func TestStatus_ReturnsTrueForExisting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && strings.Contains(r.URL.Path, "/v1/blobs/") {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	a := New()
	result, err := a.Status(context.Background(), "existing-blob-id", deployer.DeployOptions{
		AggregatorBaseURL: srv.URL,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Success {
		t.Fatal("Status should return true for existing blob")
	}
}

// TestCalculateUploadTimeout verifies dynamic timeout calculation.
func TestCalculateUploadTimeout(t *testing.T) {
	tests := []struct {
		name     string
		bodySize int
		wantMin  time.Duration
		wantMax  time.Duration
	}{
		{"small file", 1024, 30 * time.Second, 35 * time.Second},
		{"1MB file", 1024 * 1024, 30 * time.Second, 45 * time.Second},
		{"10MB file", 10 * 1024 * 1024, 100 * time.Second, 150 * time.Second},
		{"100MB file", 100 * 1024 * 1024, 10 * time.Minute, 10 * time.Minute}, // capped at max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateUploadTimeout(tt.bodySize)
			if got < tt.wantMin {
				t.Errorf("Timeout too short: got %v, want >= %v", got, tt.wantMin)
			}
			if got > tt.wantMax {
				t.Errorf("Timeout too long: got %v, want <= %v", got, tt.wantMax)
			}
		})
	}
}

// TestDeployBlobs_ConcurrencyAndRetry validates worker cap and retry policy.
func TestDeployBlobs_ConcurrencyAndRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	t.Parallel()

	// Track per-file attempts by content signature
	var maxConcurrent int32
	var currentConcurrent int32
	attempts := sync.Map{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/blobs") {
			http.NotFound(w, r)
			return
		}
		atomic.AddInt32(&currentConcurrent, 1)
		defer atomic.AddInt32(&currentConcurrent, -1)
		for {
			// update max concurrently observed
			cur := atomic.LoadInt32(&currentConcurrent)
			old := atomic.LoadInt32(&maxConcurrent)
			if cur <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, cur) {
				break
			}
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}
		if err := r.Body.Close(); err != nil {
			t.Errorf("Failed to close request body: %v", err)
		}
		sig := string(body)
		v, _ := attempts.LoadOrStore(sig, new(int32))
		ctr := v.(*int32)
		n := atomic.AddInt32(ctr, 1)

		// First attempt -> 429, second -> 200
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write([]byte("rate limited")); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"newlyCreated": map[string]any{
				"blobObject": map[string]any{
					"blobId": "blob-" + sig,
				},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer srv.Close()

	// Prepare a temp dir with files whose content identifies them
	dir := t.TempDir()
	fileCount := 20
	for i := 0; i < fileCount; i++ {
		name := filepath.Join(dir, "file-"+itoa(i)+".txt")
		if err := os.WriteFile(name, []byte("file-"+itoa(i)), 0644); err != nil {
			t.Fatalf("write temp file: %v", err)
		}
	}

	a := New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := a.Deploy(ctx, dir, deployer.DeployOptions{
		PublisherBaseURL: srv.URL,
		Mode:             "blobs",
		Workers:          10,
		MaxRetries:       3,
		Epochs:           1,
	})
	if err != nil {
		t.Fatalf("deploy error: %v", err)
	}
	if len(res.FileToBlobID) != fileCount {
		t.Fatalf("expected %d uploads, got %d", fileCount, len(res.FileToBlobID))
	}
	// Ensure concurrency cap respected (allow small slop on slow CI)
	if got := atomic.LoadInt32(&maxConcurrent); got > 10 {
		t.Fatalf("concurrency exceeded: got %d, want <= 10", got)
	}
	// Ensure each file retried once after 429
	attemptsOk := true
	attempts.Range(func(key, value any) bool {
		if atomic.LoadInt32(value.(*int32)) != 2 {
			attemptsOk = false
			return false
		}
		return true
	})
	if !attemptsOk {
		t.Fatalf("unexpected attempt counts: %+v", dumpAttempts(&attempts))
	}
}

// TestDeployBlobs_Cancel ensures cancellation stops further uploads.
func TestDeployBlobs_Cancel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	t.Parallel()
	blocker := make(chan struct{})
	var received int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/blobs") {
			http.NotFound(w, r)
			return
		}
		atomic.AddInt32(&received, 1)
		<-blocker // block to simulate slow server; cancellation should stop client
		w.WriteHeader(200)
		if _, err := w.Write([]byte(`{"alreadyCertified":{"blobId":"x"}}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	for i := 0; i < 50; i++ {
		name := filepath.Join(dir, "f-"+itoa(i))
		if err := os.WriteFile(name, []byte(strings.Repeat("x", 8)), 0644); err != nil {
			t.Fatal(err)
		}
	}

	a := New()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_, _ = a.Deploy(ctx, dir, deployer.DeployOptions{PublisherBaseURL: srv.URL, Mode: "blobs", Workers: 10, MaxRetries: 1, Epochs: 1})
		close(done)
	}()
	// Let a few requests start
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("deploy did not return after cancellation")
	}
	close(blocker)
	if atomic.LoadInt32(&received) == 0 {
		t.Fatal("server did not receive any requests; test invalid")
	}
}

// Helpers
func itoa(i int) string { return strconv.Itoa(i) }

func dumpAttempts(m *sync.Map) map[string]int32 {
	out := make(map[string]int32)
	m.Range(func(k, v any) bool {
		out[k.(string)] = *v.(*int32)
		return true
	})
	return out
}
