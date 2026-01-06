package ai

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Credentials stores AI provider API credentials.
type Credentials struct {
	Provider string `yaml:"provider"` // "openai" or "openrouter"
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url,omitempty"` // Optional custom base URL
	Model    string `yaml:"model,omitempty"`    // Model to use (e.g., "gpt-4", "openai/gpt-4")
}

// CredentialsFile represents the structure of AI credentials file.
type CredentialsFile struct {
	Providers map[string]Credentials `yaml:"providers"`
}

// GetCredentialsPath returns the file system path for the AI credentials file.
func GetCredentialsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	walgoDir := filepath.Join(homeDir, ".walgo")
	if err := os.MkdirAll(walgoDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .walgo directory: %w", err)
	}

	return filepath.Join(walgoDir, "ai-credentials.yaml"), nil
}

// LoadCredentials retrieves AI credentials from ~/.walgo/ai-credentials.yaml file.
func LoadCredentials() (*CredentialsFile, error) {
	path, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &CredentialsFile{
			Providers: make(map[string]Credentials),
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds CredentialsFile
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	if creds.Providers == nil {
		creds.Providers = make(map[string]Credentials)
	}

	return &creds, nil
}

// SaveCredentials persists AI credentials to ~/.walgo/ai-credentials.yaml file.
func SaveCredentials(creds *CredentialsFile) error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// #nosec G306 - credentials file should be restrictive
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// GetProviderCredentials retrieves credentials for specified AI provider.
func GetProviderCredentials(provider string) (*Credentials, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	providerCreds, exists := creds.Providers[provider]
	if !exists {
		return nil, fmt.Errorf("no credentials found for provider: %s", provider)
	}

	return &providerCreds, nil
}

// SetProviderCredentials stores credentials for specified AI provider.
func SetProviderCredentials(provider, apiKey, baseURL, model string) error {
	creds, err := LoadCredentials()
	if err != nil {
		return err
	}

	creds.Providers[provider] = Credentials{
		Provider: provider,
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
	}

	return SaveCredentials(creds)
}

// RemoveProviderCredentials deletes credentials for specified AI provider.
func RemoveProviderCredentials(provider string) error {
	creds, err := LoadCredentials()
	if err != nil {
		return err
	}

	// If provider doesn't exist, that's fine - already removed
	if _, exists := creds.Providers[provider]; !exists {
		return nil // Success - provider has no credentials
	}

	delete(creds.Providers, provider)
	if err := SaveCredentials(creds); err != nil {
		return fmt.Errorf("failed to save credentials after deletion: %w", err)
	}

	return nil
}

// RemoveAllCredentials deletes all stored AI credentials.
func RemoveAllCredentials() error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("no credentials file found")
	}

	return os.Remove(path)
}

// ListProviders returns a list of all configured AI providers.
func ListProviders() ([]string, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	providers := make([]string, 0, len(creds.Providers))
	for p := range creds.Providers {
		providers = append(providers, p)
	}
	return providers, nil
}

// GetDefaultBaseURL returns the default base URL for a provider
func GetDefaultBaseURL(provider string) string {
	switch provider {
	case "openai":
		return "https://api.openai.com/v1"
	case "openrouter":
		return "https://openrouter.ai/api/v1"
	default:
		return ""
	}
}
