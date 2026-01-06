// walrus_real_test.go - Tests that actually verify behavior, not mocks
//
// These tests are designed to EXPOSE bugs, not rubber-stamp existing code.
// Run with: go test -v -tags=integration ./internal/walrus/
//
//go:build integration

package walrus

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestSiteBuilderBinaryExists verifies the site-builder binary is actually installed
// and executable. This is the FIRST thing that should be tested - if this fails,
// nothing else matters.
func TestSiteBuilderBinaryExists(t *testing.T) {
	path, err := exec.LookPath("site-builder")
	if err != nil {
		t.Fatalf("site-builder binary not found in PATH: %v\n"+
			"This is a CRITICAL dependency. Install it with: walgo setup-deps", err)
	}

	// Verify it's actually executable
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("cannot stat site-builder at %s: %v", path, err)
	}

	if info.Mode()&0111 == 0 {
		t.Fatalf("site-builder at %s is not executable (mode: %v)", path, info.Mode())
	}

	// Verify it runs and returns version info
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("site-builder --version failed: %v\nOutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Fatal("site-builder --version returned empty output")
	}

	t.Logf("site-builder version: %s", strings.TrimSpace(string(output)))
}

// TestWalrusBinaryExists verifies the walrus CLI is installed
func TestWalrusBinaryExists(t *testing.T) {
	path, err := exec.LookPath("walrus")
	if err != nil {
		t.Fatalf("walrus binary not found in PATH: %v\n"+
			"Install it with: walgo setup-deps", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("walrus --version failed: %v\nOutput: %s", err, output)
	}

	t.Logf("walrus version: %s", strings.TrimSpace(string(output)))
}

// TestGetStorageInfoReturnsValidData tests that GetStorageInfo doesn't silently fail
func TestGetStorageInfoReturnsValidData(t *testing.T) {
	// Find walrus binary
	walrusBin, err := exec.LookPath("walrus")
	if err != nil {
		t.Skip("walrus binary not found, skipping storage info test")
	}

	info, err := GetStorageInfo(walrusBin)

	// The REAL test: if this returns an error, we should KNOW about it
	// Currently the code at cost.go:410 ignores this error
	if err != nil {
		t.Errorf("GetStorageInfo returned error (this is silently ignored in production!): %v", err)
	}

	if info == nil {
		t.Error("GetStorageInfo returned nil info - production code falls back to hardcoded defaults")
		t.Log("This means cost calculations may be inaccurate if walrus info fails")
	} else {
		// Validate the returned data makes sense
		if info.StorageUnit == 0 {
			t.Error("StorageUnit is 0, which is invalid")
		}
		t.Logf("Storage info: %+v", info)
	}
}

// TestDeployCreatesValidOutput tests that a real deployment to testnet works
func TestDeployCreatesValidOutput(t *testing.T) {
	if os.Getenv("WALGO_INTEGRATION_TEST") != "1" {
		t.Skip("Set WALGO_INTEGRATION_TEST=1 to run deployment tests")
	}

	// Create a minimal test site
	tempDir := t.TempDir()
	publicDir := filepath.Join(tempDir, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a minimal index.html
	indexPath := filepath.Join(publicDir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<!DOCTYPE html><html><body>Test</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	// Attempt deployment
	result, err := Deploy(DeployOptions{
		SiteDirectory: publicDir,
		Network:       "testnet",
		Epochs:        1, // Minimum storage time
	})

	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	// Validate the result structure
	if result.ObjectID == "" {
		t.Error("Deploy returned empty ObjectID - deployment may have failed silently")
	}

	if !strings.HasPrefix(result.ObjectID, "0x") {
		t.Errorf("ObjectID doesn't look like a valid Sui address: %s", result.ObjectID)
	}

	t.Logf("Deployed to: %s", result.ObjectID)
}

// TestPreflightCheckDetectsRealProblems tests that preflight actually catches issues
func TestPreflightCheckDetectsRealProblems(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError bool
		errorContains string
	}{
		{
			name: "empty directory should fail",
			setup: func(t *testing.T) string {
				return t.TempDir() // Empty dir
			},
			expectError: true,
			errorContains: "no files",
		},
		{
			name: "missing index.html should warn",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				// Create a file that's not index.html
				os.WriteFile(filepath.Join(dir, "style.css"), []byte("body{}"), 0644)
				return dir
			},
			expectError: true,
			errorContains: "index",
		},
		{
			name: "valid site should pass",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0644)
				return dir
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			err := PreflightCheck(PreflightOptions{SiteDirectory: dir})

			if tt.expectError && err == nil {
				t.Error("expected error but got nil - preflight is not catching issues")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("error should contain %q, got: %v", tt.errorContains, err)
			}
		})
	}
}

// TestCostCalculationIsReasonable ensures cost estimates aren't wildly off
func TestCostCalculationIsReasonable(t *testing.T) {
	// Create a known-size test directory
	tempDir := t.TempDir()

	// Create exactly 1MB of data
	oneMB := make([]byte, 1024*1024)
	for i := range oneMB {
		oneMB[i] = byte(i % 256)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "data.bin"), oneMB, 0644); err != nil {
		t.Fatal(err)
	}

	cost, err := EstimateCost(CostOptions{
		SiteDirectory: tempDir,
		Network:       "testnet",
		Epochs:        1,
	})

	if err != nil {
		t.Fatalf("EstimateCost failed: %v", err)
	}

	// Sanity checks - these are based on known Walrus pricing
	if cost.TotalBytes != 1024*1024 {
		t.Errorf("TotalBytes should be 1MB, got %d", cost.TotalBytes)
	}

	if cost.StorageCost == 0 {
		t.Error("StorageCost is 0, which is definitely wrong")
	}

	// Cost should be > 0 but not astronomical
	if cost.TotalCost > 1e18 { // More than 1 SUI seems wrong for 1MB on testnet
		t.Errorf("Cost seems unreasonably high: %v", cost.TotalCost)
	}

	t.Logf("Cost for 1MB on testnet: %+v", cost)
}
