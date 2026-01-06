package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name            string
		provider        string
		apiKey          string
		baseURL         string
		model           string
		expectedBaseURL string
	}{
		{
			name:            "openai provider with default baseURL",
			provider:        "openai",
			apiKey:          "test-key",
			baseURL:         "",
			model:           "gpt-4",
			expectedBaseURL: "https://api.openai.com/v1",
		},
		{
			name:            "openrouter provider with default baseURL",
			provider:        "openrouter",
			apiKey:          "test-key",
			baseURL:         "",
			model:           "openai/gpt-4",
			expectedBaseURL: "https://openrouter.ai/api/v1",
		},
		{
			name:            "custom baseURL",
			provider:        "custom",
			apiKey:          "test-key",
			baseURL:         "https://custom.api.com/v1",
			model:           "custom-model",
			expectedBaseURL: "https://custom.api.com/v1",
		},
		{
			name:            "unknown provider with empty baseURL",
			provider:        "unknown",
			apiKey:          "test-key",
			baseURL:         "",
			model:           "model",
			expectedBaseURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.provider, tt.apiKey, tt.baseURL, tt.model)

			if client.Provider != tt.provider {
				t.Errorf("expected provider %s, got %s", tt.provider, client.Provider)
			}
			if client.APIKey != tt.apiKey {
				t.Errorf("expected apiKey %s, got %s", tt.apiKey, client.APIKey)
			}
			if client.BaseURL != tt.expectedBaseURL {
				t.Errorf("expected baseURL %s, got %s", tt.expectedBaseURL, client.BaseURL)
			}
			if client.Model != tt.model {
				t.Errorf("expected model %s, got %s", tt.model, client.Model)
			}
			if client.Timeout != DefaultTimeout {
				t.Errorf("expected timeout %v, got %v", DefaultTimeout, client.Timeout)
			}
		})
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	timeout := 5 * time.Minute
	client := NewClientWithTimeout("openai", "test-key", "", "gpt-4", timeout)

	if client.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.Timeout)
	}
}

func TestClient_Chat_Success(t *testing.T) {
	expectedContent := "Hello, this is a test response."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected /chat/completions path, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("expected Authorization header with Bearer token")
		}

		// Verify request body
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Model != "gpt-4" {
			t.Errorf("expected model gpt-4, got %s", req.Model)
		}

		// Send response
		resp := ChatResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: expectedContent,
					},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")

	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	result, err := client.Chat(messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expectedContent {
		t.Errorf("expected content %s, got %s", expectedContent, result)
	}
}

func TestClient_Chat_OpenRouterHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify OpenRouter-specific headers
		if r.Header.Get("HTTP-Referer") != "https://github.com/selimozten/walgo" {
			t.Errorf("expected HTTP-Referer header for OpenRouter")
		}
		if r.Header.Get("X-Title") != "Walgo AI Content Generator" {
			t.Errorf("expected X-Title header for OpenRouter")
		}

		// Send minimal response
		resp := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{Content: "OK"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("openrouter", "test-key", server.URL, "openai/gpt-4")
	_, err := client.Chat([]Message{{Role: "user", Content: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Chat_ErrorResponses(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedErrMsg string
	}{
		{
			name:           "unauthorized",
			statusCode:     http.StatusUnauthorized,
			responseBody:   "invalid token",
			expectedErrMsg: "invalid API key",
		},
		{
			name:           "rate limited",
			statusCode:     http.StatusTooManyRequests,
			responseBody:   "rate limit exceeded",
			expectedErrMsg: "rate limited",
		},
		{
			name:           "bad request",
			statusCode:     http.StatusBadRequest,
			responseBody:   "invalid model",
			expectedErrMsg: "bad request",
		},
		{
			name:           "service unavailable",
			statusCode:     http.StatusServiceUnavailable,
			responseBody:   "service down",
			expectedErrMsg: "service temporarily unavailable",
		},
		{
			name:           "bad gateway",
			statusCode:     http.StatusBadGateway,
			responseBody:   "gateway error",
			expectedErrMsg: "service temporarily unavailable",
		},
		{
			name:           "gateway timeout",
			statusCode:     http.StatusGatewayTimeout,
			responseBody:   "gateway timeout",
			expectedErrMsg: "service temporarily unavailable",
		},
		{
			name:           "internal server error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   "internal error",
			expectedErrMsg: "API request failed with status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("openai", "test-key", server.URL, "gpt-4")
			_, err := client.Chat([]Message{{Role: "user", Content: "test"}})

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("expected error containing %q, got %q", tt.expectedErrMsg, err.Error())
			}
		})
	}
}

func TestClient_Chat_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{}, // Empty choices
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	_, err := client.Chat([]Message{{Role: "user", Content: "test"}})

	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if !strings.Contains(err.Error(), "no response from AI") {
		t.Errorf("expected 'no response from AI' error, got: %v", err)
	}
}

func TestClient_Chat_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	_, err := client.Chat([]Message{{Role: "user", Content: "test"}})

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestClient_Chat_RetryOnTransientError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("temporarily unavailable"))
			return
		}

		// Success on third attempt
		resp := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{Content: "success"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithTimeout("openai", "test-key", server.URL, "gpt-4", 30*time.Second)
	result, err := client.Chat([]Message{{Role: "user", Content: "test"}})

	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %s", result)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_Chat_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("temporarily unavailable"))
	}))
	defer server.Close()

	client := NewClientWithTimeout("openai", "test-key", server.URL, "gpt-4", 30*time.Second)
	_, err := client.Chat([]Message{{Role: "user", Content: "test"}})

	if err == nil {
		t.Fatal("expected error after max retries")
	}
	if !strings.Contains(err.Error(), "request failed after") {
		t.Errorf("expected 'request failed after' error, got: %v", err)
	}
	if attempts != MaxRetries {
		t.Errorf("expected %d attempts, got %d", MaxRetries, attempts)
	}
}

func TestClient_Chat_NoRetryOnNonTransientError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid token"))
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	_, err := client.Chat([]Message{{Role: "user", Content: "test"}})

	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retries for 401), got %d", attempts)
	}
}

func TestClient_GenerateContent(t *testing.T) {
	expectedContent := "Generated content"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify messages structure
		if len(req.Messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != "system" {
			t.Errorf("expected first message role 'system', got %s", req.Messages[0].Role)
		}
		if req.Messages[1].Role != "user" {
			t.Errorf("expected second message role 'user', got %s", req.Messages[1].Role)
		}

		resp := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{Content: expectedContent}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	result, err := client.GenerateContent("system prompt", "user prompt")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expectedContent {
		t.Errorf("expected %s, got %s", expectedContent, result)
	}
}

func TestClient_ChatWithContext_Cancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.ChatWithContext(ctx, []Message{{Role: "user", Content: "test"}})

	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestClient_GenerateContentWithContext(t *testing.T) {
	expectedContent := "Generated with context"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{Content: expectedContent}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	ctx := context.Background()
	result, err := client.GenerateContentWithContext(ctx, "system", "user")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expectedContent {
		t.Errorf("expected %s, got %s", expectedContent, result)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"timeout error", "request timeout", true},
		{"deadline exceeded", "context deadline exceeded", true},
		{"connection refused", "connection refused", true},
		{"connection reset", "connection reset by peer", true},
		{"temporarily unavailable", "service temporarily unavailable", true},
		{"rate limited", "rate limited by server", true},
		{"502 error", "received 502 from server", true},
		{"503 error", "status 503", true},
		{"504 error", "gateway timeout 504", true},
		{"unauthorized", "unauthorized access", false},
		{"bad request", "bad request", false},
		{"not found", "resource not found", false},
		{"invalid api key", "invalid API key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &testError{msg: tt.errMsg}
			result := isRetryableError(err)
			if result != tt.expected {
				t.Errorf("isRetryableError(%q) = %v, want %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestDefaultTimeout(t *testing.T) {
	if DefaultTimeout != 120*time.Second {
		t.Errorf("expected DefaultTimeout to be 120s, got %v", DefaultTimeout)
	}
}

func TestLongRequestTimeout(t *testing.T) {
	if LongRequestTimeout != 300*time.Second {
		t.Errorf("expected LongRequestTimeout to be 300s, got %v", LongRequestTimeout)
	}
}

func TestMaxRetries(t *testing.T) {
	if MaxRetries != 3 {
		t.Errorf("expected MaxRetries to be 3, got %d", MaxRetries)
	}
}
