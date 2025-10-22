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

	"walgo/internal/deployer"
)

// TestDeployBlobs_ConcurrencyAndRetry validates worker cap and retry policy.
func TestDeployBlobs_ConcurrencyAndRetry(t *testing.T) {
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

		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		sig := string(body)
		v, _ := attempts.LoadOrStore(sig, new(int32))
		ctr := v.(*int32)
		n := atomic.AddInt32(ctr, 1)

		// First attempt -> 429, second -> 200
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"newlyCreated": map[string]any{
				"blobObject": map[string]any{
					"blobId": "blob-" + sig,
				},
			},
		})
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
		_, _ = w.Write([]byte(`{"alreadyCertified":{"blobId":"x"}}`))
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
