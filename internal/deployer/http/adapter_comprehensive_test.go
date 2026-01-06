package httpdeployer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/selimozten/walgo/internal/deployer"
)

func TestStatus(t *testing.T) {
	// Create a mock aggregator server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && strings.Contains(r.URL.Path, "/v1/blobs/") {
			blobID := strings.TrimPrefix(r.URL.Path, "/v1/blobs/")
			if blobID == "existing-blob" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	tests := []struct {
		name        string
		objectID    string
		opts        deployer.DeployOptions
		wantErr     bool
		wantSuccess bool
	}{
		{
			name:     "Status requires AggregatorBaseURL",
			objectID: "test-object-123",
			opts:     deployer.DeployOptions{},
			wantErr:  true, // Should fail without AggregatorBaseURL
		},
		{
			name:     "Status with empty objectID",
			objectID: "",
			opts: deployer.DeployOptions{
				AggregatorBaseURL: srv.URL,
			},
			wantErr:     false,
			wantSuccess: false, // Empty objectID should return success=false
		},
		{
			name:     "Status for existing blob",
			objectID: "existing-blob",
			opts: deployer.DeployOptions{
				AggregatorBaseURL: srv.URL,
			},
			wantErr:     false,
			wantSuccess: true,
		},
		{
			name:     "Status for nonexistent blob",
			objectID: "nonexistent-blob",
			opts: deployer.DeployOptions{
				AggregatorBaseURL: srv.URL,
			},
			wantErr:     false,
			wantSuccess: false,
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
				}
				return
			}

			if err != nil {
				t.Errorf("Status() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Errorf("Status() result = nil, expected non-nil result")
				return
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("Status() result.Success = %v, expected %v", result.Success, tt.wantSuccess)
			}

			if result.ObjectID != tt.objectID {
				t.Errorf("Status() result.ObjectID = %v, expected %v", result.ObjectID, tt.objectID)
			}
		})
	}
}

func TestNew(t *testing.T) {
	adapter := New()
	if adapter == nil {
		t.Error("New() returned nil")
	}

	// Verify it implements interface
	var _ deployer.WalrusDeployer = adapter
}
