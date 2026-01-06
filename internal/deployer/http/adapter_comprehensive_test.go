package httpdeployer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployer"
)

func TestUpdate(t *testing.T) {
	t.Skip("TestUpdate skipped: httptest.NewServer requires network/socket access not available in CI")
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test file
	indexHTML := filepath.Join(tmpDir, "index.html")
	if err := os.WriteFile(indexHTML, []byte("<html><body>Test Site</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		siteDir     string
		objectID    string
		opts        deployer.DeployOptions
		wantErr     bool
		errMsg      string
		setupServer func() *httptest.Server
	}{
		{
			name:     "Update performs fresh deploy",
			siteDir:  tmpDir,
			objectID: "test-object-123",
			opts: deployer.DeployOptions{
				PublisherBaseURL: "", // Will be set from server
				Epochs:           1,
			},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/v1/quilts" {
						resp := map[string]interface{}{
							"blobStoreResult": map[string]interface{}{
								"newlyCreated": map[string]interface{}{
									"blobObject": map[string]interface{}{
										"blobId": "new-blob-123",
									},
								},
							},
						}
						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(resp)
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			wantErr: false,
		},
		{
			name:     "Update with non-existent directory",
			siteDir:  "/non/existent/path",
			objectID: "test-object-123",
			opts: deployer.DeployOptions{
				PublisherBaseURL: "http://localhost:9999",
				Epochs:           1,
			},
			wantErr: true,
			errMsg:  "no such file or directory",
		},
	}

	adapter := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var srv *httptest.Server
			if tt.setupServer != nil {
				srv = tt.setupServer()
				defer srv.Close()
				if tt.opts.PublisherBaseURL == "" {
					tt.opts.PublisherBaseURL = srv.URL
				}
			}

			ctx := context.Background()
			result, err := adapter.Update(ctx, tt.siteDir, tt.objectID, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Update() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Update() error = %v, should contain %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Update() unexpected error = %v", err)
				}
				if result == nil {
					t.Errorf("Update() result = nil, expected non-nil result")
				}
			}
		})
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name     string
		objectID string
		opts     deployer.DeployOptions
		wantErr  bool
	}{
		{
			name:     "Status returns success stub",
			objectID: "test-object-123",
			opts:     deployer.DeployOptions{},
			wantErr:  false,
		},
		{
			name:     "Status with empty objectID",
			objectID: "",
			opts:     deployer.DeployOptions{},
			wantErr:  false,
		},
		{
			name:     "Status with verbose flag",
			objectID: "test-object-123",
			opts: deployer.DeployOptions{
				Verbose: true,
			},
			wantErr: false,
		},
	}

	adapter := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := adapter.Status(ctx, tt.objectID, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Status() error = nil, wantErr %v", tt.wantErr)
					return
				}
			} else {
				if err != nil {
					t.Errorf("Status() unexpected error = %v", err)
				}
				if result == nil {
					t.Errorf("Status() result = nil, expected non-nil result")
				} else {
					// Verify the result matches what Status() returns
					if !result.Success {
						t.Errorf("Status() result.Success = false, expected true")
					}
					if result.ObjectID != tt.objectID {
						t.Errorf("Status() result.ObjectID = %v, expected %v", result.ObjectID, tt.objectID)
					}
				}
			}
		})
	}
}

func TestDeployQuilt(t *testing.T) {
	t.Skip("TestDeployQuilt skipped: httptest.NewServer requires network/socket access not available in CI")
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test HTML files
	indexHTML := filepath.Join(tmpDir, "index.html")
	if err := os.WriteFile(indexHTML, []byte("<html><body>Test Site</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	aboutHTML := filepath.Join(tmpDir, "about.html")
	if err := os.WriteFile(aboutHTML, []byte("<html><body>About Page</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a CSS file
	cssDir := filepath.Join(tmpDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatal(err)
	}
	cssFile := filepath.Join(cssDir, "style.css")
	if err := os.WriteFile(cssFile, []byte("body { margin: 0; }"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		siteDir      string
		publisher    string
		epochs       int
		setupServer  func() *httptest.Server
		serverURL    string
		wantErr      bool
		errContains  string
		validateResp func(*deployer.Result) error
	}{
		{
			name:      "Successful quilt deployment",
			siteDir:   tmpDir,
			publisher: "test-publisher",
			epochs:    5,
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/v1/quilts" {
						// Check method
						if r.Method != http.MethodPut {
							w.WriteHeader(http.StatusMethodNotAllowed)
							return
						}

						// Read the body to validate it's a proper tar.gz
						body, err := io.ReadAll(r.Body)
						if err != nil {
							w.WriteHeader(http.StatusBadRequest)
							return
						}

						// Simple validation that we got some data
						if len(body) == 0 {
							w.WriteHeader(http.StatusBadRequest)
							_, _ = w.Write([]byte("Empty request body"))
							return
						}

						// Return success response matching the actual expected format
						resp := map[string]interface{}{
							"blobStoreResult": map[string]interface{}{
								"newlyCreated": map[string]interface{}{
									"blobObject": map[string]interface{}{
										"blobId": "quilt-123",
									},
								},
							},
							"storedQuiltBlobs": []map[string]interface{}{
								{
									"identifier":   "index.html",
									"quiltPatchId": "patch-1",
								},
								{
									"identifier":   "about.html",
									"quiltPatchId": "patch-2",
								},
							},
						}
						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(resp)
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			wantErr: false,
			validateResp: func(result *deployer.Result) error {
				if result == nil {
					return fmt.Errorf("expected result, got nil")
				}
				if !result.Success {
					return fmt.Errorf("expected success=true")
				}
				if result.ObjectID != "quilt-123" {
					return fmt.Errorf("expected ObjectID=quilt-123, got %s", result.ObjectID)
				}
				if len(result.QuiltPatches) != 2 {
					return fmt.Errorf("expected 2 quilt patches, got %d", len(result.QuiltPatches))
				}
				return nil
			},
		},
		{
			name:        "Empty site directory",
			siteDir:     "",
			publisher:   "test-publisher",
			epochs:      1,
			setupServer: func() *httptest.Server { return nil },
			serverURL:   "http://localhost:9999",
			wantErr:     true,
			errContains: "no such file or directory",
		},
		{
			name:        "Non-existent site directory",
			siteDir:     "/non/existent/path",
			publisher:   "test-publisher",
			epochs:      1,
			setupServer: func() *httptest.Server { return nil },
			serverURL:   "http://localhost:9999",
			wantErr:     true,
			errContains: "no such file or directory",
		},
		{
			name:      "Server returns error",
			siteDir:   tmpDir,
			publisher: "test-publisher",
			epochs:    1,
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("Internal server error"))
				}))
			},
			wantErr:     true,
			errContains: "500",
		},
		{
			name:      "Invalid JSON response",
			siteDir:   tmpDir,
			publisher: "test-publisher",
			epochs:    1,
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/v1/quilts" {
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte("invalid json"))
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
			},
			wantErr:     true,
			errContains: "invalid character",
		},
		{
			name:      "Timeout during request",
			siteDir:   tmpDir,
			publisher: "test-publisher",
			epochs:    1,
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Simulate a slow server
					time.Sleep(2 * time.Second)
					w.WriteHeader(http.StatusOK)
				}))
			},
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := tt.serverURL
			if tt.setupServer != nil {
				srv := tt.setupServer()
				if srv != nil {
					defer srv.Close()
					// Use test server URL as publisher if not already set
					if publisher == "" {
						publisher = srv.URL
					}
				}
			}

			adapter := New()

			// Create context with timeout for timeout test
			ctx := context.Background()
			if tt.name == "Timeout during request" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()
			}

			result, err := adapter.deployQuilt(ctx, tt.siteDir, publisher, tt.epochs)

			if tt.wantErr {
				if err == nil {
					t.Errorf("deployQuilt() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("deployQuilt() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("deployQuilt() unexpected error = %v", err)
					return
				}
				if tt.validateResp != nil {
					if err := tt.validateResp(result); err != nil {
						t.Errorf("Response validation failed: %v", err)
					}
				}
			}
		})
	}
}

func TestDeploy_WithBlobStrategy(t *testing.T) {
	t.Skip("TestDeploy_WithBlobStrategy skipped: httptest.NewServer requires network/socket access not available in CI")
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test HTML file
	indexHTML := filepath.Join(tmpDir, "index.html")
	if err := os.WriteFile(indexHTML, []byte("<html><body>Test Site</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test server for quilt strategy (default)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/quilts" {
			// Return success response for quilt deployment
			resp := map[string]interface{}{
				"blobStoreResult": map[string]interface{}{
					"newlyCreated": map[string]interface{}{
						"blobObject": map[string]interface{}{
							"blobId": "blob-789",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		} else if strings.HasPrefix(r.URL.Path, "/v1/blobs") {
			// Also handle blob requests if mode is changed
			resp := map[string]interface{}{
				"newlyCreated": map[string]interface{}{
					"blobObject": map[string]interface{}{
						"blobId": "blob-individual",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Set environment variable for publisher endpoint
	os.Setenv("WALRUS_PUBLISHER_URL", srv.URL)
	defer os.Unsetenv("WALRUS_PUBLISHER_URL")

	adapter := New()
	ctx := context.Background()

	// Test with default blob strategy
	opts := deployer.DeployOptions{
		WalrusCfg: config.WalrusConfig{
			ProjectID: "test-project",
		},
		PublisherBaseURL: srv.URL,
		Verbose:          true,
		Epochs:           1,
	}

	result, err := adapter.Deploy(ctx, tmpDir, opts)
	if err != nil {
		t.Errorf("Deploy() with blob strategy unexpected error = %v", err)
		return
	}

	if result == nil {
		t.Error("Deploy() with blob strategy returned nil result")
		return
	}

	if !result.Success {
		t.Error("Deploy() with blob strategy expected success=true")
	}
}

func TestNew(t *testing.T) {
	adapter := New()
	if adapter == nil {
		t.Error("New() returned nil")
	}

	// Verify it implements the interface
	var _ deployer.WalrusDeployer = adapter
}
