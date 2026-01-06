package sitebuilder

import (
	"context"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployer"
	"testing"
)

func TestNew(t *testing.T) {
	adapter := New()
	if adapter == nil {
		t.Fatal("New() returned nil")
	}
}

func TestAdapter_Deploy(t *testing.T) {
	// Skip this test as it requires the actual site-builder binary
	// which causes timeout issues in CI/CD
	t.Skip("Skipping site-builder integration test - requires actual binary")

	tests := []struct {
		name    string
		siteDir string
		opts    deployer.DeployOptions
		wantErr bool
	}{
		{
			name:    "Deploy with valid options",
			siteDir: "/tmp/test-site",
			opts: deployer.DeployOptions{
				Verbose: true,
				Epochs:  5,
				WalrusCfg: config.WalrusConfig{
					ProjectID:  "test-project",
					Entrypoint: "index.html",
				},
			},
			wantErr: true, // Will error without actual site-builder
		},
		{
			name:    "Deploy with minimal options",
			siteDir: "/tmp/test-site",
			opts: deployer.DeployOptions{
				WalrusCfg: config.WalrusConfig{
					ProjectID: "test-project",
				},
			},
			wantErr: true, // Will error without actual site-builder
		},
		{
			name:    "Deploy with verbose flag",
			siteDir: "/tmp/test-site",
			opts: deployer.DeployOptions{
				Verbose: true,
				WalrusCfg: config.WalrusConfig{
					ProjectID: "test-project",
				},
			},
			wantErr: true, // Will error without actual site-builder
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := New()
			ctx := context.Background()

			result, err := adapter.Deploy(ctx, tt.siteDir, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Deploy() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result == nil {
				t.Error("Deploy() returned nil result without error")
			}
		})
	}
}

func TestAdapter_Update(t *testing.T) {
	// Skip this test as it requires the actual site-builder binary
	t.Skip("Skipping site-builder integration test - requires actual binary")

	tests := []struct {
		name     string
		siteDir  string
		objectID string
		opts     deployer.DeployOptions
		wantErr  bool
	}{
		{
			name:     "Update with valid object ID",
			siteDir:  "/tmp/test-site",
			objectID: "0x123abc",
			opts: deployer.DeployOptions{
				Epochs: 5,
			},
			wantErr: true, // Will error without actual site-builder
		},
		{
			name:     "Update with minimal options",
			siteDir:  "/tmp/test-site",
			objectID: "0xabc123",
			opts:     deployer.DeployOptions{},
			wantErr:  true, // Will error without actual site-builder
		},
		{
			name:     "Update with empty object ID",
			siteDir:  "/tmp/test-site",
			objectID: "",
			opts:     deployer.DeployOptions{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := New()
			ctx := context.Background()

			result, err := adapter.Update(ctx, tt.siteDir, tt.objectID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result == nil {
				t.Error("Update() returned nil result without error")
			}
		})
	}
}

func TestAdapter_Status(t *testing.T) {
	// Skip this test as it requires the actual site-builder binary
	t.Skip("Skipping site-builder integration test - requires actual binary")

	tests := []struct {
		name     string
		objectID string
		opts     deployer.DeployOptions
		wantErr  bool
	}{
		{
			name:     "Status with valid object ID",
			objectID: "0x123abc",
			opts:     deployer.DeployOptions{},
			wantErr:  true, // Will error without actual site-builder
		},
		{
			name:     "Status with empty object ID",
			objectID: "",
			opts:     deployer.DeployOptions{},
			wantErr:  true,
		},
		{
			name:     "Status with verbose option",
			objectID: "0xdef456",
			opts: deployer.DeployOptions{
				Verbose: true,
			},
			wantErr: true, // Will error without actual site-builder
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := New()
			ctx := context.Background()

			result, err := adapter.Status(ctx, tt.objectID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Status() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result == nil {
				t.Error("Status() returned nil result without error")
			}
		})
	}
}

func TestAdapter_ContextCancellation(t *testing.T) {
	// Skip this test as it requires the actual site-builder binary
	t.Skip("Skipping site-builder integration test - requires actual binary")

	adapter := New()

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	t.Run("Deploy with cancelled context", func(t *testing.T) {
		// The current implementation doesn't check context, but test anyway
		opts := deployer.DeployOptions{
			WalrusCfg: config.WalrusConfig{
				ProjectID: "test",
			},
		}
		_, err := adapter.Deploy(ctx, "/tmp/test", opts)
		// Will error due to missing site-builder, not context cancellation
		if err == nil {
			t.Error("Expected error with cancelled context")
		}
	})

	t.Run("Update with cancelled context", func(t *testing.T) {
		opts := deployer.DeployOptions{}
		_, err := adapter.Update(ctx, "/tmp/test", "0x123", opts)
		// Will error due to missing site-builder, not context cancellation
		if err == nil {
			t.Error("Expected error with cancelled context")
		}
	})

	t.Run("Status with cancelled context", func(t *testing.T) {
		opts := deployer.DeployOptions{}
		_, err := adapter.Status(ctx, "0x123", opts)
		// Will error due to missing site-builder, not context cancellation
		if err == nil {
			t.Error("Expected error with cancelled context")
		}
	})
}

func TestAdapter_ResultHandling(t *testing.T) {
	// Skip this test as it requires the actual site-builder binary
	t.Skip("Skipping site-builder integration test - requires actual binary")

	// Test that the adapter properly converts walrus results to deployer results
	// This would require mocking the walrus package functions,
	// which would need refactoring for dependency injection

	adapter := New()
	ctx := context.Background()

	// Test Deploy result handling
	t.Run("Deploy result conversion", func(t *testing.T) {
		opts := deployer.DeployOptions{
			WalrusCfg: config.WalrusConfig{
				ProjectID: "test",
			},
		}
		result, err := adapter.Deploy(ctx, "/tmp/nonexistent", opts)

		// Should error without site-builder
		if err == nil {
			t.Error("Expected error without site-builder")
		}

		if result != nil {
			t.Error("Expected nil result on error")
		}
	})

	// Test Update result handling
	t.Run("Update result conversion", func(t *testing.T) {
		opts := deployer.DeployOptions{}
		result, err := adapter.Update(ctx, "/tmp/nonexistent", "0x123", opts)

		// Should error without site-builder
		if err == nil {
			t.Error("Expected error without site-builder")
		}

		if result != nil {
			t.Error("Expected nil result on error")
		}
	})

	// Test Status result handling
	t.Run("Status result conversion", func(t *testing.T) {
		opts := deployer.DeployOptions{}
		result, err := adapter.Status(ctx, "0x123", opts)

		// Should error without site-builder
		if err == nil {
			t.Error("Expected error without site-builder")
		}

		if result != nil {
			t.Error("Expected nil result on error")
		}
	})
}

func TestAdapter_VerboseMode(t *testing.T) {
	// Skip this test as it requires the actual site-builder binary
	t.Skip("Skipping site-builder integration test - requires actual binary")

	adapter := New()
	ctx := context.Background()

	// Test that verbose flag is properly passed through
	opts := deployer.DeployOptions{
		Verbose: true,
		WalrusCfg: config.WalrusConfig{
			ProjectID: "test",
		},
	}

	// This will set verbose mode even though the actual deploy will fail
	_, _ = adapter.Deploy(ctx, "/tmp/test", opts)
	// The SetVerbose call is made before the actual deployment attempt
}

func TestAdapter_EpochsHandling(t *testing.T) {
	// Skip this test as it requires the actual site-builder binary
	t.Skip("Skipping site-builder integration test - requires actual binary")

	adapter := New()
	ctx := context.Background()

	tests := []struct {
		name   string
		epochs int
	}{
		{
			name:   "Zero epochs",
			epochs: 0,
		},
		{
			name:   "Positive epochs",
			epochs: 5,
		},
		{
			name:   "Large number of epochs",
			epochs: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Deploy with epochs
			deployOpts := deployer.DeployOptions{
				Epochs: tt.epochs,
				WalrusCfg: config.WalrusConfig{
					ProjectID: "test",
				},
			}
			_, _ = adapter.Deploy(ctx, "/tmp/test", deployOpts)

			// Test Update with epochs
			updateOpts := deployer.DeployOptions{
				Epochs: tt.epochs,
			}
			_, _ = adapter.Update(ctx, "/tmp/test", "0x123", updateOpts)
		})
	}
}

// Benchmark tests
func BenchmarkAdapter_Deploy(b *testing.B) {
	adapter := New()
	ctx := context.Background()
	opts := deployer.DeployOptions{
		WalrusCfg: config.WalrusConfig{
			ProjectID: "bench-test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Will error but we're testing performance of the adapter layer
		_, _ = adapter.Deploy(ctx, "/tmp/bench", opts)
	}
}

func BenchmarkAdapter_Update(b *testing.B) {
	adapter := New()
	ctx := context.Background()
	opts := deployer.DeployOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Will error but we're testing performance of the adapter layer
		_, _ = adapter.Update(ctx, "/tmp/bench", "0x123", opts)
	}
}

func BenchmarkAdapter_Status(b *testing.B) {
	adapter := New()
	ctx := context.Background()
	opts := deployer.DeployOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Will error but we're testing performance of the adapter layer
		_, _ = adapter.Status(ctx, "0x123", opts)
	}
}
