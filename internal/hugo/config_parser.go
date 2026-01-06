package hugo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetBaseURL extracts the production baseURL from Hugo config files.
// This is a unified implementation that should be used everywhere.
//
// It checks both hugo.toml and config.toml for baseURL setting,
// filtering out placeholder values (example.com, localhost).
//
// Parameters:
//
//	sitePath: Path to the Hugo site root directory
//
// Returns:
//
//	string: The production baseURL, or error if not found
//	error: Error if baseURL cannot be found in either config file
func GetBaseURL(sitePath string) (string, error) {
	// Try hugo.toml first
	hugoTomlPath := filepath.Join(sitePath, "hugo.toml")
	if baseURL, err := extractBaseURLFromConfig(hugoTomlPath); err == nil && baseURL != "" {
		return baseURL, nil
	}

	// Try config.toml as fallback
	configTomlPath := filepath.Join(sitePath, "config.toml")
	if baseURL, err := extractBaseURLFromConfig(configTomlPath); err == nil && baseURL != "" {
		return baseURL, nil
	}

	return "", fmt.Errorf("baseURL not found in hugo.toml or config.toml")
}

// extractBaseURLFromConfig extracts baseURL from a specific config file.
func extractBaseURLFromConfig(configPath string) (string, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse file line by line for baseURL
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for baseURL or baseurl (case-insensitive)
		if strings.HasPrefix(line, "baseURL") || strings.HasPrefix(line, "baseurl") {
			// Extract value after = sign
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				baseURL := strings.TrimSpace(parts[1])

				// Remove quotes (single or double)
				baseURL = strings.Trim(baseURL, `"`)
				baseURL = strings.Trim(baseURL, `'`)

				// Skip placeholder values
				if baseURL != "" &&
					!strings.Contains(baseURL, "example.") &&
					!strings.Contains(baseURL, "localhost") {
					return baseURL, nil
				}
			}
		}
	}

	return "", fmt.Errorf("baseURL not found in %s", configPath)
}
