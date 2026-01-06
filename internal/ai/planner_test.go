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

func TestNewPlanner(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()

	planner := NewPlanner(client, config)

	if planner.client != client {
		t.Error("expected client to be set")
	}
	if planner.config.MaxRetries != config.MaxRetries {
		t.Error("expected config to be set")
	}
}

func TestPlanner_Plan_Success(t *testing.T) {
	planResponse := AIPlanResponse{
		Site: AISiteInfo{
			Type:     "blog",
			Title:    "My Blog",
			Language: "en",
			Tone:     "professional",
			BaseURL:  "https://myblog.com",
		},
		Pages: []AIPageInfo{
			{
				ID:       "home",
				Path:     "content/_index.md",
				PageType: "home",
				Frontmatter: map[string]any{
					"title": "Welcome",
					"draft": false,
				},
				Outline:       []string{"Introduction", "Features"},
				InternalLinks: []string{"/about/", "/contact/"},
			},
			{
				ID:       "about",
				Path:     "content/about.md",
				PageType: "page",
				Frontmatter: map[string]any{
					"title": "About",
				},
			},
			{
				ID:       "contact",
				Path:     "content/contact.md",
				PageType: "page",
			},
			{
				ID:       "post1",
				Path:     "content/posts/welcome/index.md",
				PageType: "post",
			},
			{
				ID:       "post2",
				Path:     "content/posts/getting-started/index.md",
				PageType: "post",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respJSON, _ := json.Marshal(planResponse)
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
				}{Content: string(respJSON)}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	planner := NewPlanner(client, config)

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	plan, err := planner.Plan(ctx, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == nil {
		t.Fatal("expected plan to be returned")
	}
	if plan.SiteName != "My Blog" {
		t.Errorf("expected SiteName 'My Blog', got %s", plan.SiteName)
	}
	if plan.SiteType != SiteTypeBlog {
		t.Errorf("expected SiteType blog, got %s", plan.SiteType)
	}
	if len(plan.Pages) != 5 {
		t.Errorf("expected 5 pages, got %d", len(plan.Pages))
	}
	if plan.Status != PlanStatusPending {
		t.Errorf("expected status pending, got %s", plan.Status)
	}
	if plan.ID == "" {
		t.Error("expected plan ID to be set")
	}
	if plan.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", plan.Version)
	}
}

func TestPlanner_Plan_ValidationErrors(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	planner := NewPlanner(client, config)

	tests := []struct {
		name          string
		input         *PlannerInput
		expectedError string
	}{
		{
			name:          "nil input",
			input:         nil,
			expectedError: "input is required",
		},
		{
			name: "empty site name",
			input: &PlannerInput{
				SiteName:    "",
				SiteType:    SiteTypeBlog,
				Description: "A blog",
			},
			expectedError: "site name is required",
		},
		{
			name: "whitespace site name",
			input: &PlannerInput{
				SiteName:    "   ",
				SiteType:    SiteTypeBlog,
				Description: "A blog",
			},
			expectedError: "site name is required",
		},
		{
			name: "invalid site type",
			input: &PlannerInput{
				SiteName:    "My Site",
				SiteType:    SiteType("invalid"),
				Description: "A site",
			},
			expectedError: "invalid site type",
		},
		{
			name: "empty description",
			input: &PlannerInput{
				SiteName:    "My Site",
				SiteType:    SiteTypeBlog,
				Description: "",
			},
			expectedError: "description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := planner.Plan(ctx, tt.input)

			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error containing %q, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestPlanner_Plan_PlanValidation(t *testing.T) {
	tests := []struct {
		name          string
		planResponse  AIPlanResponse
		expectedError string
	}{
		{
			name: "empty pages",
			planResponse: AIPlanResponse{
				Site:  AISiteInfo{Type: "blog"},
				Pages: []AIPageInfo{},
			},
			expectedError: "plan contains no pages",
		},
		{
			name: "missing home page",
			planResponse: AIPlanResponse{
				Site: AISiteInfo{Type: "blog"},
				Pages: []AIPageInfo{
					{ID: "about", Path: "content/about.md", PageType: "page"},
					{ID: "contact", Path: "content/contact.md", PageType: "page"},
					{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
					{ID: "post2", Path: "content/posts/start/index.md", PageType: "post"},
					{ID: "extra", Path: "content/extra.md", PageType: "page"},
				},
			},
			expectedError: "home page",
		},
		{
			name: "too few pages for blog",
			planResponse: AIPlanResponse{
				Site: AISiteInfo{Type: "blog"},
				Pages: []AIPageInfo{
					{ID: "home", Path: "content/_index.md", PageType: "home"},
					{ID: "about", Path: "content/about.md", PageType: "page"},
				},
			},
			expectedError: "requires at least",
		},
		{
			name: "duplicate paths",
			planResponse: AIPlanResponse{
				Site: AISiteInfo{Type: "blog"},
				Pages: []AIPageInfo{
					{ID: "home", Path: "content/_index.md", PageType: "home"},
					{ID: "about", Path: "content/about.md", PageType: "page"},
					{ID: "about2", Path: "content/about.md", PageType: "page"}, // Duplicate
					{ID: "contact", Path: "content/contact.md", PageType: "page"},
					{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
				},
			},
			expectedError: "duplicate path",
		},
		{
			name: "path not starting with content/",
			planResponse: AIPlanResponse{
				Site: AISiteInfo{Type: "blog"},
				Pages: []AIPageInfo{
					{ID: "home", Path: "content/_index.md", PageType: "home"},
					{ID: "about", Path: "about.md", PageType: "page"}, // Missing content/
					{ID: "contact", Path: "content/contact.md", PageType: "page"},
					{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
					{ID: "post2", Path: "content/posts/start/index.md", PageType: "post"},
				},
			},
			expectedError: "must start with 'content/'",
		},
		{
			name: "empty path",
			planResponse: AIPlanResponse{
				Site: AISiteInfo{Type: "blog"},
				Pages: []AIPageInfo{
					{ID: "home", Path: "content/_index.md", PageType: "home"},
					{ID: "about", Path: "", PageType: "page"}, // Empty path
					{ID: "contact", Path: "content/contact.md", PageType: "page"},
					{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
					{ID: "post2", Path: "content/posts/start/index.md", PageType: "post"},
				},
			},
			expectedError: "path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				respJSON, _ := json.Marshal(tt.planResponse)
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
						}{Content: string(respJSON)}},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient("openai", "test-key", server.URL, "gpt-4")
			config := DefaultPipelineConfig()
			planner := NewPlanner(client, config)

			input := &PlannerInput{
				SiteName:    "My Blog",
				SiteType:    SiteTypeBlog,
				Description: "A blog about technology",
				Audience:    "developers",
			}

			ctx := context.Background()
			_, err := planner.Plan(ctx, input)

			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error containing %q, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestPlanner_Plan_AlternativeFormat(t *testing.T) {
	// Test parsing alternative JSON format
	alternativeResponse := `{
		"site_name": "My Blog",
		"site_type": "blog",
		"description": "A blog",
		"pages": [
			{"id": "home", "path": "content/_index.md", "page_type": "home"},
			{"id": "about", "path": "content/about.md", "page_type": "page"},
			{"id": "contact", "path": "content/contact.md", "page_type": "page"},
			{"id": "post1", "path": "content/posts/welcome/index.md", "page_type": "post"},
			{"id": "post2", "path": "content/posts/start/index.md", "page_type": "post"}
		]
	}`

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
				}{Content: alternativeResponse}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	planner := NewPlanner(client, config)

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	plan, err := planner.Plan(ctx, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Pages) != 5 {
		t.Errorf("expected 5 pages, got %d", len(plan.Pages))
	}
}

func TestPlanner_Plan_WithMarkdownFence(t *testing.T) {
	planResponse := AIPlanResponse{
		Site: AISiteInfo{Type: "blog"},
		Pages: []AIPageInfo{
			{ID: "home", Path: "content/_index.md", PageType: "home"},
			{ID: "about", Path: "content/about.md", PageType: "page"},
			{ID: "contact", Path: "content/contact.md", PageType: "page"},
			{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
			{ID: "post2", Path: "content/posts/start/index.md", PageType: "post"},
		},
	}

	planJSON, _ := json.Marshal(planResponse)
	wrappedResponse := "```json\n" + string(planJSON) + "\n```"

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
				}{Content: wrappedResponse}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	planner := NewPlanner(client, config)

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	plan, err := planner.Plan(ctx, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Pages) != 5 {
		t.Errorf("expected 5 pages, got %d", len(plan.Pages))
	}
}

func TestPlanner_Plan_InvalidJSON(t *testing.T) {
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
				}{Content: "not valid json at all"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	planner := NewPlanner(client, config)

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	_, err := planner.Plan(ctx, input)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestPlanner_Plan_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.PlannerTimeout = 100 * time.Millisecond
	planner := NewPlanner(client, config)

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	_, err := planner.Plan(ctx, input)

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no fences",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "json fence",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON fence uppercase",
			input:    "```JSON\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "plain fence",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "with whitespace",
			input:    "  ```json\n{\"key\": \"value\"}\n```  ",
			expected: `{"key": "value"}`,
		},
		{
			name:     "only opening fence",
			input:    "```json\n{\"key\": \"value\"}",
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanJSONResponse(tt.input)
			if result != tt.expected {
				t.Errorf("CleanJSONResponse():\ngot: %s\nwant: %s", result, tt.expected)
			}
		})
	}
}

func TestPlanner_DeterminePageType(t *testing.T) {
	planner := &Planner{}

	tests := []struct {
		pageType string
		path     string
		expected PageType
	}{
		{"home", "content/_index.md", PageTypeHome},
		{"post", "content/posts/test.md", PageTypePost},
		{"blog", "content/posts/test.md", PageTypePost},
		{"docs", "content/docs/intro.md", PageTypeDocs},
		{"documentation", "content/docs/intro.md", PageTypeDocs},
		{"section", "content/posts/_index.md", PageTypeSection},
		{"page", "content/about.md", PageTypePage},
		{"", "content/_index.md", PageTypeHome},
		{"", "content/posts/_index.md", PageTypeSection},
		{"", "content/docs/_index.md", PageTypeSection},
		{"", "content/posts/welcome/index.md", PageTypePost},
		{"", "content/blog/entry.md", PageTypePost},
		{"", "content/docs/intro.md", PageTypeDocs},
		{"", "content/about.md", PageTypePage},
	}

	for _, tt := range tests {
		t.Run(tt.pageType+"_"+tt.path, func(t *testing.T) {
			result := planner.determinePageType(tt.pageType, tt.path)
			if result != tt.expected {
				t.Errorf("determinePageType(%s, %s) = %s, want %s",
					tt.pageType, tt.path, result, tt.expected)
			}
		})
	}
}

func TestPlanner_DetermineContentType(t *testing.T) {
	planner := &Planner{}

	tests := []struct {
		path     string
		expected string
	}{
		{"content/posts/welcome.md", "posts"},
		{"content/docs/intro.md", "docs"},
		{"content/services/service-1.md", "services"},
		{"content/about.md", ""}, // Root level file
		{"content/_index.md", ""}, // Root level index
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := planner.determineContentType(tt.path)
			if result != tt.expected {
				t.Errorf("determineContentType(%s) = %s, want %s",
					tt.path, result, tt.expected)
			}
		})
	}
}

func TestPlanner_ValidatePlan(t *testing.T) {
	planner := &Planner{}

	t.Run("nil plan", func(t *testing.T) {
		err := planner.validatePlan(nil)
		if err == nil {
			t.Error("expected error for nil plan")
		}
	})

	t.Run("empty ID", func(t *testing.T) {
		plan := &SitePlan{
			ID:    "",
			Pages: []PageSpec{{Path: "content/_index.md"}},
		}
		err := planner.validatePlan(plan)
		if err == nil {
			t.Error("expected error for empty ID")
		}
	})

	t.Run("empty pages", func(t *testing.T) {
		plan := &SitePlan{
			ID:    "test",
			Pages: []PageSpec{},
		}
		err := planner.validatePlan(plan)
		if err == nil {
			t.Error("expected error for empty pages")
		}
	})
}
