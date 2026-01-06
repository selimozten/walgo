package ai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetCredentialsPath(t *testing.T) {
	path, err := GetCredentialsPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".walgo", "ai-credentials.yaml")

	if path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, path)
	}
}

func TestCredentialsRoundTrip(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .walgo directory
	walgoDir := filepath.Join(tempDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Test saving credentials
	creds := &CredentialsFile{
		Providers: map[string]Credentials{
			"openai": {
				Provider: "openai",
				APIKey:   "test-openai-key",
				BaseURL:  "https://api.openai.com/v1",
				Model:    "gpt-4",
			},
			"openrouter": {
				Provider: "openrouter",
				APIKey:   "test-openrouter-key",
				BaseURL:  "https://openrouter.ai/api/v1",
				Model:    "openai/gpt-4",
			},
		},
	}

	err := SaveCredentials(creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	// Verify file was created with correct permissions
	path, _ := GetCredentialsPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("credentials file not created: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected file permissions 0600, got %o", info.Mode().Perm())
	}

	// Test loading credentials
	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	if len(loaded.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(loaded.Providers))
	}

	openai, exists := loaded.Providers["openai"]
	if !exists {
		t.Fatal("openai provider not found")
	}
	if openai.APIKey != "test-openai-key" {
		t.Errorf("expected APIKey 'test-openai-key', got %s", openai.APIKey)
	}
	if openai.Model != "gpt-4" {
		t.Errorf("expected Model 'gpt-4', got %s", openai.Model)
	}
}

func TestLoadCredentials_NoFile(t *testing.T) {
	// Create temp directory with no credentials file
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .walgo directory
	walgoDir := filepath.Join(tempDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Should return empty credentials file, not error
	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials should not error for missing file: %v", err)
	}

	if creds.Providers == nil {
		t.Error("Providers map should not be nil")
	}
	if len(creds.Providers) != 0 {
		t.Errorf("expected empty providers, got %d", len(creds.Providers))
	}
}

func TestLoadCredentials_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .walgo directory and invalid credentials file
	walgoDir := filepath.Join(tempDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	credPath := filepath.Join(walgoDir, "ai-credentials.yaml")
	if err := os.WriteFile(credPath, []byte("not: valid: yaml: content:"), 0600); err != nil {
		t.Fatalf("failed to write invalid yaml: %v", err)
	}

	_, err := LoadCredentials()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestGetProviderCredentials(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .walgo directory
	walgoDir := filepath.Join(tempDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Save test credentials
	creds := &CredentialsFile{
		Providers: map[string]Credentials{
			"openai": {
				Provider: "openai",
				APIKey:   "test-key",
				Model:    "gpt-4",
			},
		},
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("failed to save credentials: %v", err)
	}

	// Test getting existing provider
	provCreds, err := GetProviderCredentials("openai")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provCreds.APIKey != "test-key" {
		t.Errorf("expected APIKey 'test-key', got %s", provCreds.APIKey)
	}

	// Test getting non-existent provider
	_, err = GetProviderCredentials("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent provider")
	}
}

func TestSetProviderCredentials(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .walgo directory
	walgoDir := filepath.Join(tempDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Set credentials
	err := SetProviderCredentials("openai", "new-key", "https://custom.url", "gpt-4-turbo")
	if err != nil {
		t.Fatalf("SetProviderCredentials failed: %v", err)
	}

	// Verify credentials were saved
	provCreds, err := GetProviderCredentials("openai")
	if err != nil {
		t.Fatalf("failed to get credentials: %v", err)
	}

	if provCreds.APIKey != "new-key" {
		t.Errorf("expected APIKey 'new-key', got %s", provCreds.APIKey)
	}
	if provCreds.BaseURL != "https://custom.url" {
		t.Errorf("expected BaseURL 'https://custom.url', got %s", provCreds.BaseURL)
	}
	if provCreds.Model != "gpt-4-turbo" {
		t.Errorf("expected Model 'gpt-4-turbo', got %s", provCreds.Model)
	}
}

func TestRemoveProviderCredentials(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .walgo directory
	walgoDir := filepath.Join(tempDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Set up credentials
	creds := &CredentialsFile{
		Providers: map[string]Credentials{
			"openai": {Provider: "openai", APIKey: "key1"},
			"openrouter": {Provider: "openrouter", APIKey: "key2"},
		},
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("failed to save credentials: %v", err)
	}

	// Remove openai credentials
	err := RemoveProviderCredentials("openai")
	if err != nil {
		t.Fatalf("RemoveProviderCredentials failed: %v", err)
	}

	// Verify removal
	_, err = GetProviderCredentials("openai")
	if err == nil {
		t.Error("expected error after removing provider")
	}

	// Verify other provider still exists
	_, err = GetProviderCredentials("openrouter")
	if err != nil {
		t.Error("openrouter should still exist")
	}

	// Try to remove non-existent provider
	err = RemoveProviderCredentials("nonexistent")
	if err == nil {
		t.Error("expected error for removing non-existent provider")
	}
}

func TestRemoveAllCredentials(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .walgo directory
	walgoDir := filepath.Join(tempDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create credentials file
	creds := &CredentialsFile{
		Providers: map[string]Credentials{
			"openai": {Provider: "openai", APIKey: "key1"},
		},
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("failed to save credentials: %v", err)
	}

	// Remove all credentials
	err := RemoveAllCredentials()
	if err != nil {
		t.Fatalf("RemoveAllCredentials failed: %v", err)
	}

	// Verify file is gone
	path, _ := GetCredentialsPath()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("credentials file should be removed")
	}

	// Try to remove again - should error
	err = RemoveAllCredentials()
	if err == nil {
		t.Error("expected error when no credentials file exists")
	}
}

func TestListProviders(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .walgo directory
	walgoDir := filepath.Join(tempDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Test empty list
	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders failed: %v", err)
	}
	if len(providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(providers))
	}

	// Add some providers
	creds := &CredentialsFile{
		Providers: map[string]Credentials{
			"openai":     {Provider: "openai", APIKey: "key1"},
			"openrouter": {Provider: "openrouter", APIKey: "key2"},
			"custom":     {Provider: "custom", APIKey: "key3"},
		},
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("failed to save credentials: %v", err)
	}

	// List providers
	providers, err = ListProviders()
	if err != nil {
		t.Fatalf("ListProviders failed: %v", err)
	}
	if len(providers) != 3 {
		t.Errorf("expected 3 providers, got %d", len(providers))
	}

	// Check that all providers are present
	providerSet := make(map[string]bool)
	for _, p := range providers {
		providerSet[p] = true
	}
	if !providerSet["openai"] {
		t.Error("openai should be in list")
	}
	if !providerSet["openrouter"] {
		t.Error("openrouter should be in list")
	}
	if !providerSet["custom"] {
		t.Error("custom should be in list")
	}
}

func TestGetDefaultBaseURL(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{"openai", "https://api.openai.com/v1"},
		{"openrouter", "https://openrouter.ai/api/v1"},
		{"unknown", ""},
		{"", ""},
		{"OPENAI", ""}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			result := GetDefaultBaseURL(tt.provider)
			if result != tt.expected {
				t.Errorf("GetDefaultBaseURL(%q) = %q, want %q", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestCredentials_EmptyProvidersMap(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create .walgo directory
	walgoDir := filepath.Join(tempDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Write credentials file with nil providers
	credPath := filepath.Join(walgoDir, "ai-credentials.yaml")
	content := `providers:
`
	if err := os.WriteFile(credPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write credentials file: %v", err)
	}

	// Load and verify providers map is initialized
	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}
	if creds.Providers == nil {
		t.Error("Providers map should be initialized, not nil")
	}
}
