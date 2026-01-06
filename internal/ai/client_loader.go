package ai

import (
	"fmt"
	"time"
)

// LoadClient retrieves and initializes an AI client from stored credentials.
// This is the unified function that should be used by all commands.
// It checks OpenAI and OpenRouter providers in order and returns the first
// valid credential found.
//
// Parameters:
//
//	timeout: Custom timeout duration (use 0 for default timeout)
//
// Returns:
//
//	*Client: Initialized AI client
//	Provider: Name of the provider being used ("openai" or "openrouter")
//	Model: Model name being used
//	error: Error if no valid credentials found
func LoadClient(timeout time.Duration) (*Client, string, string, error) {
	providers := []string{"openai", "openrouter"}

	for _, provider := range providers {
		creds, err := GetProviderCredentials(provider)
		if err == nil && creds.APIKey != "" {
			// Resolve model name (use default if not specified)
			model := resolveModel(provider, creds.Model)

			// Create client with or without timeout
			var client *Client
			if timeout > 0 {
				client = NewClientWithTimeout(creds.Provider, creds.APIKey, creds.BaseURL, model, timeout)
			} else {
				client = NewClient(creds.Provider, creds.APIKey, creds.BaseURL, model)
			}

			return client, provider, model, nil
		}
	}

	return nil, "", "", fmt.Errorf("no AI credentials found - run 'walgo ai configure' first")
}

// resolveModel returns the appropriate model name based on provider and user configuration.
// If no model is specified, it returns a sensible default.
func resolveModel(provider, configuredModel string) string {
	if configuredModel != "" {
		return configuredModel
	}

	// Return defaults if no model configured
	switch provider {
	case "openrouter":
		return "openai/gpt-4"
	default: // openai
		return "gpt-4"
	}
}
