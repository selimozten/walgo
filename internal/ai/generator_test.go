package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()

	generator := NewGenerator(client, config)

	if generator.client != client {
		t.Error("expected client to be set")
	}
	if generator.config.MaxRetries != config.MaxRetries {
		t.Error("expected config to be set")
	}
}

func TestGenerator_SetProgressHandler(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	generator := NewGenerator(client, config)

	handler := func(_ ProgressEvent) {
		// Handler set successfully
	}

	generator.SetProgressHandler(handler)

	// Verify progress handler is set
	if generator.progress == nil {
		t.Error("progress handler should be set")
	}
}

func TestGenerator_GeneratePage_Success(t *testing.T) {
	expectedContent := "---\ntitle: About\ndraft: false\n---\n\n## About Me\n\nThis is the about page."

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

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	config.DryRun = false
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Stats: PlanStats{
			TotalPages: 1,
		},
	}

	page := &PageSpec{
		ID:       "about",
		Path:     "content/about.md",
		PageType: PageTypePage,
		Title:    "About",
		Status:   PageStatusPending,
	}

	ctx := context.Background()
	output := generator.GeneratePage(ctx, plan, page)

	if !output.Success {
		t.Errorf("expected success, got error: %s", output.ErrorMsg)
	}
	if output.PageID != "about" {
		t.Errorf("expected PageID 'about', got %s", output.PageID)
	}
	if output.Skipped {
		t.Error("page should not be skipped")
	}
	if output.Duration <= 0 {
		t.Error("duration should be positive")
	}

	// Verify file was written
	filePath := filepath.Join(tempDir, "about.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("failed to read generated file: %v", err)
	}
	if string(content) != expectedContent {
		t.Errorf("file content mismatch:\ngot: %s\nwant: %s", string(content), expectedContent)
	}
}

func TestGenerator_GeneratePage_DryRun(t *testing.T) {
	expectedContent := "---\ntitle: About\n---\nContent"

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

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	config.DryRun = true // Dry run enabled
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Stats:       PlanStats{TotalPages: 1},
	}

	page := &PageSpec{
		ID:       "about",
		Path:     "content/about.md",
		PageType: PageTypePage,
		Status:   PageStatusPending,
	}

	ctx := context.Background()
	output := generator.GeneratePage(ctx, plan, page)

	if !output.Success {
		t.Errorf("expected success, got error: %s", output.ErrorMsg)
	}

	// File should NOT be written in dry run mode
	filePath := filepath.Join(tempDir, "about.md")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("file should not be created in dry run mode")
	}
}

func TestGenerator_GeneratePage_SkipExisting(t *testing.T) {
	tempDir := t.TempDir()

	// Create existing file
	existingPath := filepath.Join(tempDir, "about.md")
	if err := os.WriteFile(existingPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	client := NewClient("openai", "test-key", "http://unused.test", "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	config.OverwriteExisting = false // Don't overwrite
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Stats:       PlanStats{TotalPages: 1},
	}

	page := &PageSpec{
		ID:       "about",
		Path:     "content/about.md",
		PageType: PageTypePage,
		Status:   PageStatusPending,
	}

	ctx := context.Background()
	output := generator.GeneratePage(ctx, plan, page)

	if !output.Success {
		t.Errorf("expected success, got error: %s", output.ErrorMsg)
	}
	if !output.Skipped {
		t.Error("page should be skipped when file exists")
	}

	// Verify existing content was not modified
	content, _ := os.ReadFile(existingPath)
	if string(content) != "existing content" {
		t.Error("existing file should not be modified")
	}
}

func TestGenerator_GeneratePage_OverwriteExisting(t *testing.T) {
	newContent := "---\ntitle: New\n---\nNew content"

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
				}{Content: newContent}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tempDir := t.TempDir()

	// Create existing file
	existingPath := filepath.Join(tempDir, "about.md")
	if err := os.WriteFile(existingPath, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	config.OverwriteExisting = true // Overwrite enabled
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Stats:       PlanStats{TotalPages: 1},
	}

	page := &PageSpec{
		ID:       "about",
		Path:     "content/about.md",
		PageType: PageTypePage,
		Status:   PageStatusPending,
	}

	ctx := context.Background()
	output := generator.GeneratePage(ctx, plan, page)

	if !output.Success {
		t.Errorf("expected success, got error: %s", output.ErrorMsg)
	}
	if output.Skipped {
		t.Error("page should not be skipped when overwrite is enabled")
	}

	// Verify content was overwritten
	content, _ := os.ReadFile(existingPath)
	if string(content) != newContent {
		t.Error("file should be overwritten")
	}
}

func TestGenerator_GeneratePage_Retry(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("temporarily unavailable"))
			return
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
				}{Content: "---\ntitle: Test\n---"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	config.RetryDelay = 10 * time.Millisecond
	config.RetryBackoffMulti = 1.0
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Stats:       PlanStats{TotalPages: 1},
	}

	page := &PageSpec{
		ID:       "test",
		Path:     "content/test.md",
		PageType: PageTypePage,
		Status:   PageStatusPending,
	}

	ctx := context.Background()
	output := generator.GeneratePage(ctx, plan, page)

	if !output.Success {
		t.Errorf("expected success after retry, got error: %s", output.ErrorMsg)
	}
	// The client has its own internal retry loop that handles 503 errors.
	// The generator's Attempts counts generator-level attempts, not HTTP requests.
	// Since the client successfully retries internally, the generator only makes 1 attempt.
	if output.Attempts != 1 {
		t.Errorf("expected 1 generator attempt (client handles retries internally), got %d", output.Attempts)
	}
}

func TestGenerator_GeneratePage_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	config.GeneratorTimeout = 5 * time.Second
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Stats:       PlanStats{TotalPages: 1},
	}

	page := &PageSpec{
		ID:       "test",
		Path:     "content/test.md",
		PageType: PageTypePage,
		Status:   PageStatusPending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	output := generator.GeneratePage(ctx, plan, page)

	if output.Success {
		t.Error("expected failure due to context cancellation")
	}
	if output.Error == nil {
		t.Error("expected error to be set")
	}
}

func TestGenerator_GenerateAll(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
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
				}{Content: "---\ntitle: Test\ndraft: false\n---"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Pages: []PageSpec{
			{ID: "home", Path: "content/_index.md", Status: PageStatusPending},
			{ID: "about", Path: "content/about.md", Status: PageStatusPending},
			{ID: "contact", Path: "content/contact.md", Status: PageStatusPending},
		},
		Stats: PlanStats{TotalPages: 3},
	}

	ctx := context.Background()
	results, err := generator.GenerateAll(ctx, plan)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	if callCount != 3 {
		t.Errorf("expected 3 API calls, got %d", callCount)
	}

	// Verify all succeeded
	for _, result := range results {
		if !result.Success {
			t.Errorf("page %s failed: %s", result.PageID, result.ErrorMsg)
		}
	}
}

func TestGenerator_GenerateAll_SkipsCompleted(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
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
				}{Content: "---\ntitle: Test\n---"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Pages: []PageSpec{
			{ID: "home", Path: "content/_index.md", Status: PageStatusCompleted}, // Already done
			{ID: "about", Path: "content/about.md", Status: PageStatusPending},
			{ID: "contact", Path: "content/contact.md", Status: PageStatusSkipped}, // Already skipped
		},
		Stats: PlanStats{TotalPages: 3},
	}

	ctx := context.Background()
	results, err := generator.GenerateAll(ctx, plan)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 { // Only about.md should be processed
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if callCount != 1 {
		t.Errorf("expected 1 API call, got %d", callCount)
	}
}

func TestGenerator_GenerateAll_ContinueOnError(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad request"))
			return
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
				}{Content: "---\ntitle: Test\n---"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	config.ContinueOnError = true
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Pages: []PageSpec{
			{ID: "home", Path: "content/_index.md", Status: PageStatusPending},
			{ID: "about", Path: "content/about.md", Status: PageStatusPending},
		},
		Stats: PlanStats{TotalPages: 2},
	}

	ctx := context.Background()
	results, err := generator.GenerateAll(ctx, plan)

	if err != nil {
		t.Fatalf("unexpected error with ContinueOnError: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	// First page should fail, second should succeed
	if results[0].Success {
		t.Error("first page should have failed")
	}
	if !results[1].Success {
		t.Error("second page should have succeeded")
	}
}

func TestGenerator_GenerateAll_StopOnError(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	config.ContinueOnError = false // Stop on first error
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Pages: []PageSpec{
			{ID: "home", Path: "content/_index.md", Status: PageStatusPending},
			{ID: "about", Path: "content/about.md", Status: PageStatusPending},
		},
		Stats: PlanStats{TotalPages: 2},
	}

	ctx := context.Background()
	results, err := generator.GenerateAll(ctx, plan)

	if err == nil {
		t.Error("expected error when ContinueOnError is false")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result before stopping, got %d", len(results))
	}
	if callCount != 1 {
		t.Errorf("expected 1 API call before stopping, got %d", callCount)
	}
}

func TestGenerator_GenerateAll_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
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
				}{Content: "---\ntitle: Test\n---"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	generator := NewGenerator(client, config)

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Pages: []PageSpec{
			{ID: "home", Path: "content/_index.md", Status: PageStatusPending},
			{ID: "about", Path: "content/about.md", Status: PageStatusPending},
			{ID: "contact", Path: "content/contact.md", Status: PageStatusPending},
		},
		Stats: PlanStats{TotalPages: 3},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := generator.GenerateAll(ctx, plan)

	if err == nil {
		t.Error("expected error due to context cancellation")
	}
}

func TestGenerator_WriteFile_CreatesDirectories(t *testing.T) {
	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	generator := NewGenerator(client, config)

	// Write to nested directory that doesn't exist
	filePath := filepath.Join(tempDir, "a", "b", "c", "test.md")
	err := generator.writeFile(filePath, "test content")

	if err != nil {
		t.Fatalf("writeFile failed: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("unexpected content: %s", string(content))
	}
}

func TestGenerator_ProgressHandler(t *testing.T) {
	events := make([]ProgressEvent, 0)

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
				}{Content: "---\ntitle: Test\n---"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.ContentDir = tempDir
	generator := NewGenerator(client, config)

	generator.SetProgressHandler(func(event ProgressEvent) {
		events = append(events, event)
	})

	plan := &SitePlan{
		SiteName:    "Test Site",
		SiteType:    SiteTypeBlog,
		Description: "A test site",
		Audience:    "testers",
		Pages: []PageSpec{
			{
				ID:       "test",
				Path:     "content/test.md",
				PageType: PageTypePage,
				Status:   PageStatusPending,
			},
		},
		Stats: PlanStats{TotalPages: 1},
	}

	ctx := context.Background()
	generator.GenerateAll(ctx, plan)

	if len(events) < 2 {
		t.Errorf("expected at least 2 events, got %d", len(events))
	}

	// Check for page start and done events
	hasStart := false
	hasDone := false
	for _, e := range events {
		if e.EventType == ProgressPageStart {
			hasStart = true
		}
		if e.EventType == ProgressPageDone {
			hasDone = true
		}
	}

	if !hasStart {
		t.Error("expected ProgressPageStart event")
	}
	if !hasDone {
		t.Error("expected ProgressPageDone event")
	}
}
