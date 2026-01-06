package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Default timeout configurations for AI API requests
const (
	DefaultTimeout     = 120 * time.Second // 2 minutes for simple requests
	LongRequestTimeout = 300 * time.Second // 5 minutes for site generation
	MaxRetries         = 3
)

// Client manages communication with AI provider APIs.
type Client struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
	client   *http.Client
	Timeout  time.Duration
}

// Message represents a single chat message in the conversation.
type Message struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // Message content
}

// ChatRequest represents the API request payload for chat completion.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

// ChatResponse represents the API response from chat completion endpoint.
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewClient initializes and returns a new AI client instance.
func NewClient(provider, apiKey, baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = GetDefaultBaseURL(provider)
	}

	return &Client{
		Provider: provider,
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
		Timeout:  DefaultTimeout,
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// NewClientWithTimeout initializes a new AI client instance with custom timeout duration.
func NewClientWithTimeout(provider, apiKey, baseURL, model string, timeout time.Duration) *Client {
	client := NewClient(provider, apiKey, baseURL, model)
	client.Timeout = timeout
	client.client.Timeout = timeout
	return client
}

// Chat sends a chat completion request to the AI provider with automatic retry logic.
func (c *Client) Chat(messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/chat/completions", c.BaseURL)

	var lastErr error
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		result, err := c.doRequest(endpoint, jsonData)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return "", err
		}

		// Wait before retry (exponential backoff)
		if attempt < MaxRetries {
			backoff := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoff)
		}
	}

	return "", fmt.Errorf("request failed after %d attempts: %w", MaxRetries, lastErr)
}

// doRequest executes a single HTTP request to the AI provider.
func (c *Client) doRequest(endpoint string, jsonData []byte) (string, error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// OpenRouter-specific headers
	if c.Provider == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://github.com/selimozten/walgo")
		req.Header.Set("X-Title", "Walgo AI Content Generator")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		// Check for timeout
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			return "", fmt.Errorf("request timed out after %v - try increasing timeout or check your network", c.Timeout)
		}
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Success, continue parsing
	case http.StatusTooManyRequests:
		return "", fmt.Errorf("rate limited - please wait and try again: %s", string(body))
	case http.StatusUnauthorized:
		return "", fmt.Errorf("invalid API key - run 'walgo ai configure' to update credentials")
	case http.StatusBadRequest:
		return "", fmt.Errorf("bad request (check model name): %s", string(body))
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		return "", fmt.Errorf("service temporarily unavailable (retryable): %s", string(body))
	default:
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI - the model may have rejected the request")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// isRetryableError determines whether an error should trigger a retry attempt.
func isRetryableError(err error) bool {
	errStr := err.Error()
	retryablePatterns := []string{
		"timeout",
		"deadline exceeded",
		"connection refused",
		"connection reset",
		"temporarily unavailable",
		"rate limited",
		"502", "503", "504",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// GenerateContent creates content based on the provided system and user prompts.
func (c *Client) GenerateContent(systemPrompt, userPrompt string) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	return c.Chat(messages)
}

// Context-Aware Methods

// GenerateContentWithContext creates content with context support for cancellation and timeout.
func (c *Client) GenerateContentWithContext(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	return c.ChatWithContext(ctx, messages)
}

// ChatWithContext sends a chat completion request with context support for cancellation.
func (c *Client) ChatWithContext(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/chat/completions", c.BaseURL)

	var lastErr error
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		result, err := c.doRequestWithContext(ctx, endpoint, jsonData)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !isRetryableError(err) {
			return "", err
		}

		if attempt < MaxRetries {
			backoff := time.Duration(attempt*attempt) * time.Second
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return "", fmt.Errorf("request failed after %d attempts: %w", MaxRetries, lastErr)
}

// doRequestWithContext executes a single HTTP request with context for cancellation.
func (c *Client) doRequestWithContext(ctx context.Context, endpoint string, jsonData []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// OpenRouter-specific headers
	if c.Provider == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://github.com/selimozten/walgo")
		req.Header.Set("X-Title", "Walgo AI Content Generator")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		// Check for context cancellation
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		// Check for timeout
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			return "", fmt.Errorf("request timed out after %v - try increasing timeout or check your network", c.Timeout)
		}
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Success, continue parsing
	case http.StatusTooManyRequests:
		return "", fmt.Errorf("rate limited - please wait and try again: %s", string(body))
	case http.StatusUnauthorized:
		return "", fmt.Errorf("invalid API key - run 'walgo ai configure' to update credentials")
	case http.StatusBadRequest:
		return "", fmt.Errorf("bad request (check model name): %s", string(body))
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		return "", fmt.Errorf("service temporarily unavailable (retryable): %s", string(body))
	default:
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI - the model may have rejected the request")
	}

	return chatResp.Choices[0].Message.Content, nil
}
