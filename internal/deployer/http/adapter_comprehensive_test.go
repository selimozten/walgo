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

	"walgo/internal/config"
	"walgo/internal/deployer"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		name     string
		siteDir  string
		objectID string
		opts     deployer.DeployOptions
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Update not implemented",
			siteDir:  "/tmp/test-site",
			objectID: "test-object-123",
			opts:     deployer.DeployOptions{},
			wantErr:  true,
			errMsg:   "HTTP deployer does not support updates",
		},
		{
			name:     "Update with empty objectID",
			siteDir:  "/tmp/test-site",
			objectID: "",
			opts:     deployer.DeployOptions{},
			wantErr:  true,
			errMsg:   "HTTP deployer does not support updates",
		},
		{
			name:     "Update with epochs",
			siteDir:  "/tmp/test-site",
			objectID: "test-object-123",
			opts: deployer.DeployOptions{
				Epochs: 10,
			},
			wantErr:  true,
			errMsg:   "HTTP deployer does not support updates",
		},
	}

	adapter := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := adapter.Update(ctx, tt.siteDir, tt.objectID, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Update() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Update() error = %v, wantErr %v", err.Error(), tt.errMsg)
				}
				if result != nil {
					t.Errorf("Update() result = %v, want nil", result)
				}
			} else {
				if err != nil {
					t.Errorf("Update() unexpected error = %v", err)
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
		errMsg   string
	}{
		{
			name:     "Status not implemented",
			objectID: "test-object-123",
			opts:     deployer.DeployOptions{},
			wantErr:  true,
			errMsg:   "HTTP deployer does not support status checks",
		},
		{
			name:     "Status with empty objectID",
			objectID: "",
			opts:     deployer.DeployOptions{},
			wantErr:  true,
			errMsg:   "HTTP deployer does not support status checks",
		},
		{
			name:     "Status with verbose flag",
			objectID: "test-object-123",
			opts: deployer.DeployOptions{
				Verbose: true,
			},
			wantErr:  true,
			errMsg:   "HTTP deployer does not support status checks",
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
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Status() error = %v, wantErr %v", err.Error(), tt.errMsg)
				}
				if result != nil {
					t.Errorf("Status() result = %v, want nil", result)
				}
			} else {
				if err != nil {
					t.Errorf("Status() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestDeployQuilt(t *testing.T) {
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
						if r.Method != http.MethodPost {
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
							w.Write([]byte("Empty request body"))
							return
						}

						// Return success response
						resp := map[string]interface{}{
							"id":     "quilt-123",
							"status": "stored",
							"urls": []string{
								"https://example.walrus.site",
								"https://backup.walrus.site",
							},
						}
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(resp)
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
				if len(result.BrowseURLs) != 2 {
					return fmt.Errorf("expected 2 browse URLs, got %d", len(result.BrowseURLs))
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
			errContains: "siteDir cannot be empty",
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
					w.Write([]byte("Internal server error"))
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
						w.Write([]byte("invalid json"))
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
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test HTML file
	indexHTML := filepath.Join(tmpDir, "index.html")
	if err := os.WriteFile(indexHTML, []byte("<html><body>Test Site</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test server for blob strategy (default)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/blobs") {
			resp := map[string]interface{}{
				"id": "blob-789",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
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
		Verbose: true,
		Epochs:  1,
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