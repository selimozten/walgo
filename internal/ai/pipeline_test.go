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

func TestNewPipeline(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()

	pipeline := NewPipeline(client, config)

	if pipeline.client != client {
		t.Error("expected client to be set")
	}
	if pipeline.planner == nil {
		t.Error("expected planner to be initialized")
	}
	if pipeline.generator == nil {
		t.Error("expected generator to be initialized")
	}
}

func TestPipeline_SetProgressHandler(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	pipeline := NewPipeline(client, config)

	events := make([]ProgressEvent, 0)
	handler := func(event ProgressEvent) {
		events = append(events, event)
	}

	pipeline.SetProgressHandler(handler)

	if pipeline.progress == nil {
		t.Error("progress handler should be set on pipeline")
	}
}

func TestPipeline_PlanOnly_Success(t *testing.T) {
	planResponse := AIPlanResponse{
		Site: AISiteInfo{Type: "blog"},
		Pages: []AIPageInfo{
			{ID: "home", Path: "content/_index.md", PageType: "home"},
			{ID: "about", Path: "content/about.md", PageType: "page"},
			{ID: "contact", Path: "content/contact.md", PageType: "page"},
			{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
			{ID: "post2", Path: "content/posts/start/index.md", PageType: "post"},
			{ID: "post3", Path: "content/posts/another/index.md", PageType: "post"},
			{ID: "post4", Path: "content/posts/last/index.md", PageType: "post"},
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
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "plan.json")
	pipeline := NewPipeline(client, config)

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	plan, err := pipeline.PlanOnly(ctx, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == nil {
		t.Fatal("expected plan to be returned")
	}
	if len(plan.Pages) != 7 {
		t.Errorf("expected 7 pages, got %d", len(plan.Pages))
	}

	// Verify plan was saved
	if !pipeline.HasPlan() {
		t.Error("expected plan to be saved")
	}
}

func TestPipeline_Run_Success(t *testing.T) {
	callCount := 0
	planResponse := AIPlanResponse{
		Site: AISiteInfo{Type: "blog"},
		Pages: []AIPageInfo{
			{ID: "home", Path: "content/_index.md", PageType: "home"},
			{ID: "about", Path: "content/about.md", PageType: "page"},
			{ID: "contact", Path: "content/contact.md", PageType: "page"},
			{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
			{ID: "post2", Path: "content/posts/start/index.md", PageType: "post"},
			{ID: "post3", Path: "content/posts/another/index.md", PageType: "post"},
			{ID: "post4", Path: "content/posts/last/index.md", PageType: "post"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// First call is for planning
		if callCount == 1 {
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
			return
		}

		// Subsequent calls are for page generation
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
				}{Content: "---\ntitle: Test\ndraft: false\n---\nContent"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, ".walgo", "plan.json")
	config.ContentDir = tempDir
	pipeline := NewPipeline(client, config)

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	result, err := pipeline.Run(ctx, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.ErrorMsg)
	}
	if result.Plan.Status != PlanStatusCompleted {
		t.Errorf("expected completed status, got %s", result.Plan.Status)
	}
	if len(result.Pages) != 7 {
		t.Errorf("expected 7 page results, got %d", len(result.Pages))
	}
	// 1 planning call + 7 generation calls
	if callCount != 8 {
		t.Errorf("expected 8 API calls, got %d", callCount)
	}
}

func TestPipeline_GenerateFromPlan(t *testing.T) {
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
				}{Content: "---\ntitle: Test\ndraft: false\n---"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "plan.json")
	config.ContentDir = tempDir
	pipeline := NewPipeline(client, config)

	plan := &SitePlan{
		ID:          "test-plan",
		Version:     "1.0",
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog",
		Audience:    "developers",
		Pages: []PageSpec{
			{ID: "home", Path: "content/_index.md", Status: PageStatusPending},
			{ID: "about", Path: "content/about.md", Status: PageStatusPending},
		},
		Stats: PlanStats{TotalPages: 2},
	}

	ctx := context.Background()
	result, err := pipeline.GenerateFromPlan(ctx, plan)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.ErrorMsg)
	}
	if len(result.Pages) != 2 {
		t.Errorf("expected 2 page results, got %d", len(result.Pages))
	}
}

func TestPipeline_SaveAndLoadPlan(t *testing.T) {
	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "plan.json")
	pipeline := NewPipeline(client, config)

	plan := &SitePlan{
		ID:          "test-plan",
		Version:     "1.0",
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog",
		Audience:    "developers",
		Pages: []PageSpec{
			{ID: "home", Path: "content/_index.md", Status: PageStatusPending},
			{ID: "about", Path: "content/about.md", Status: PageStatusPending},
		},
		Stats: PlanStats{TotalPages: 2},
	}

	// Save plan
	err := pipeline.SavePlan(plan)
	if err != nil {
		t.Fatalf("SavePlan failed: %v", err)
	}

	// Verify HasPlan
	if !pipeline.HasPlan() {
		t.Error("HasPlan should return true")
	}

	// Load plan
	loaded, err := pipeline.LoadPlan()
	if err != nil {
		t.Fatalf("LoadPlan failed: %v", err)
	}

	if loaded.ID != plan.ID {
		t.Errorf("expected ID %s, got %s", plan.ID, loaded.ID)
	}
	if loaded.SiteName != plan.SiteName {
		t.Errorf("expected SiteName %s, got %s", plan.SiteName, loaded.SiteName)
	}
	if len(loaded.Pages) != len(plan.Pages) {
		t.Errorf("expected %d pages, got %d", len(plan.Pages), len(loaded.Pages))
	}
}

func TestPipeline_LoadPlan_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "nonexistent.json")
	pipeline := NewPipeline(client, config)

	_, err := pipeline.LoadPlan()
	if err == nil {
		t.Error("expected error for missing plan")
	}
	if err != ErrPlanNotFound {
		t.Errorf("expected ErrPlanNotFound, got %v", err)
	}
}

func TestPipeline_LoadPlan_InvalidVersion(t *testing.T) {
	tempDir := t.TempDir()
	planPath := filepath.Join(tempDir, "plan.json")

	// Write plan with invalid version
	plan := &SitePlan{
		ID:       "test",
		Version:  "2.0", // Invalid version
		SiteName: "Test",
		SiteType: SiteTypeBlog,
	}
	data, _ := json.Marshal(plan)
	os.WriteFile(planPath, data, 0600)

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = planPath
	pipeline := NewPipeline(client, config)

	_, err := pipeline.LoadPlan()
	if err == nil {
		t.Error("expected error for invalid version")
	}
}

func TestPipeline_LoadPlan_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	planPath := filepath.Join(tempDir, "plan.json")

	// Write invalid JSON
	os.WriteFile(planPath, []byte("not json"), 0600)

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = planPath
	pipeline := NewPipeline(client, config)

	_, err := pipeline.LoadPlan()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestPipeline_DeletePlan(t *testing.T) {
	tempDir := t.TempDir()
	planPath := filepath.Join(tempDir, "plan.json")

	// Create plan file
	os.WriteFile(planPath, []byte("{}"), 0600)

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = planPath
	pipeline := NewPipeline(client, config)

	if !pipeline.HasPlan() {
		t.Error("expected plan to exist")
	}

	err := pipeline.DeletePlan()
	if err != nil {
		t.Fatalf("DeletePlan failed: %v", err)
	}

	if pipeline.HasPlan() {
		t.Error("expected plan to be deleted")
	}
}

func TestPipeline_DeletePlan_NotExists(t *testing.T) {
	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "nonexistent.json")
	pipeline := NewPipeline(client, config)

	// Should not error when file doesn't exist
	err := pipeline.DeletePlan()
	if err != nil {
		t.Errorf("DeletePlan should not error for non-existent file: %v", err)
	}
}

func TestPipeline_Resume(t *testing.T) {
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
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "plan.json")
	config.ContentDir = tempDir
	pipeline := NewPipeline(client, config)

	// Create a partial plan
	plan := &SitePlan{
		ID:          "test-plan",
		Version:     "1.0",
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog",
		Audience:    "developers",
		Status:      PlanStatusPartial, // Partial - needs resuming
		Pages: []PageSpec{
			{ID: "home", Path: "content/_index.md", Status: PageStatusCompleted},
			{ID: "about", Path: "content/about.md", Status: PageStatusPending}, // Not done
		},
		Stats: PlanStats{TotalPages: 2, CompletedPages: 1},
	}
	if err := pipeline.SavePlan(plan); err != nil {
		t.Fatalf("Failed to save plan: %v", err)
	}

	ctx := context.Background()
	result, err := pipeline.Resume(ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Pages) != 1 { // Only about.md should be processed
		t.Errorf("expected 1 page result, got %d", len(result.Pages))
	}
}

func TestPipeline_Resume_NoPlan(t *testing.T) {
	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "nonexistent.json")
	pipeline := NewPipeline(client, config)

	ctx := context.Background()
	_, err := pipeline.Resume(ctx)

	if err == nil {
		t.Error("expected error when no plan exists")
	}
}

func TestPipeline_Resume_AlreadyCompleted(t *testing.T) {
	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "plan.json")
	pipeline := NewPipeline(client, config)

	// Create a completed plan
	plan := &SitePlan{
		ID:          "test-plan",
		Version:     "1.0",
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog",
		Audience:    "developers",
		Status:      PlanStatusCompleted, // Already done
		Pages:       []PageSpec{},
		Stats:       PlanStats{},
	}
	if err := pipeline.SavePlan(plan); err != nil {
		t.Fatalf("Failed to save plan: %v", err)
	}

	ctx := context.Background()
	_, err := pipeline.Resume(ctx)

	if err == nil {
		t.Error("expected error when plan is already completed")
	}
}

func TestPipeline_Run_ResumesExistingPlan(t *testing.T) {
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
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "plan.json")
	config.ContentDir = tempDir
	pipeline := NewPipeline(client, config)

	// Create an existing partial plan
	existingPlan := &SitePlan{
		ID:          "existing-plan",
		Version:     "1.0",
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog",
		Audience:    "developers",
		Status:      PlanStatusPartial,
		Pages: []PageSpec{
			{ID: "home", Path: "content/_index.md", Status: PageStatusCompleted},
			{ID: "about", Path: "content/about.md", Status: PageStatusPending},
		},
		Stats: PlanStats{TotalPages: 2, CompletedPages: 1},
	}
	if err := pipeline.SavePlan(existingPlan); err != nil {
		t.Fatalf("Failed to save plan: %v", err)
	}

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	result, err := pipeline.Run(ctx, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use existing plan, not create new one
	if result.Plan.ID != "existing-plan" {
		t.Error("should resume from existing plan")
	}

	// Only 1 page should be generated (about.md)
	if callCount != 1 {
		t.Errorf("expected 1 API call, got %d", callCount)
	}
}

func TestPipeline_WithConfig(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config1 := DefaultPipelineConfig()
	config1.Verbose = false

	pipeline1 := NewPipeline(client, config1)

	config2 := DefaultPipelineConfig()
	config2.Verbose = true

	pipeline2 := pipeline1.WithConfig(config2)

	if pipeline1.config.Verbose != false {
		t.Error("original pipeline config should not change")
	}
	if pipeline2.config.Verbose != true {
		t.Error("new pipeline should have new config")
	}
}

func TestPipeline_Config(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.Verbose = true

	pipeline := NewPipeline(client, config)
	returnedConfig := pipeline.Config()

	if returnedConfig.Verbose != true {
		t.Error("Config() should return the pipeline's config")
	}
}

func TestPipeline_Client(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()

	pipeline := NewPipeline(client, config)
	returnedClient := pipeline.Client()

	if returnedClient != client {
		t.Error("Client() should return the pipeline's client")
	}
}

func TestPipeline_ProgressEmission(t *testing.T) {
	planResponse := AIPlanResponse{
		Site: AISiteInfo{Type: "blog"},
		Pages: []AIPageInfo{
			{ID: "home", Path: "content/_index.md", PageType: "home"},
			{ID: "about", Path: "content/about.md", PageType: "page"},
			{ID: "contact", Path: "content/contact.md", PageType: "page"},
			{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
			{ID: "post2", Path: "content/posts/start/index.md", PageType: "post"},
			{ID: "post3", Path: "content/posts/another/index.md", PageType: "post"},
			{ID: "post4", Path: "content/posts/last/index.md", PageType: "post"},
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
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "plan.json")
	config.DryRun = true
	pipeline := NewPipeline(client, config)

	events := make([]ProgressEvent, 0)
	pipeline.SetProgressHandler(func(event ProgressEvent) {
		events = append(events, event)
	})

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	pipeline.PlanOnly(ctx, input)

	// Verify we got planning events
	hasStart := false
	hasComplete := false
	for _, e := range events {
		if e.Phase == PhasePlanning && e.EventType == ProgressStart {
			hasStart = true
		}
		if e.Phase == PhasePlanning && e.EventType == ProgressComplete {
			hasComplete = true
		}
	}

	if !hasStart {
		t.Error("expected planning start event")
	}
	if !hasComplete {
		t.Error("expected planning complete event")
	}
}

func TestPipeline_Run_PartialStatus(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// First call is for planning
		if callCount == 1 {
			planResponse := AIPlanResponse{
				Site: AISiteInfo{Type: "blog"},
				Pages: []AIPageInfo{
					{ID: "home", Path: "content/_index.md", PageType: "home"},
					{ID: "about", Path: "content/about.md", PageType: "page"},
					{ID: "contact", Path: "content/contact.md", PageType: "page"},
					{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
					{ID: "post2", Path: "content/posts/start/index.md", PageType: "post"},
					{ID: "post3", Path: "content/posts/another/index.md", PageType: "post"},
					{ID: "post4", Path: "content/posts/last/index.md", PageType: "post"},
				},
			}
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
			return
		}

		// Fail the second page (about.md)
		if callCount == 3 {
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
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "plan.json")
	config.ContentDir = tempDir
	config.ContinueOnError = true
	pipeline := NewPipeline(client, config)

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	ctx := context.Background()
	result, err := pipeline.Run(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With ContinueOnError=true, should have result even with partial status
	// May or may not have error depending on implementation
	if result.Plan == nil {
		t.Fatal("expected result.Plan to not be nil")
	}
	if result.Plan.Status != PlanStatusPartial {
		t.Errorf("expected partial status, got %s", result.Plan.Status)
	}
	if result.Success {
		t.Error("expected Success to be false when partial")
	}
}

func TestPipeline_GetPlanPath_Absolute(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = "/absolute/path/plan.json"
	pipeline := NewPipeline(client, config)

	path := pipeline.getPlanPath()
	if path != "/absolute/path/plan.json" {
		t.Errorf("expected absolute path to be used as-is, got %s", path)
	}
}

func TestPipeline_SavePlan_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	planPath := filepath.Join(tempDir, "nested", "deep", "plan.json")

	client := NewClient("openai", "test-key", "", "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = planPath
	pipeline := NewPipeline(client, config)

	plan := &SitePlan{
		ID:          "test",
		Version:     "1.0",
		SiteName:    "Test",
		SiteType:    SiteTypeBlog,
		Description: "Test",
		Audience:    "testers",
		Pages:       []PageSpec{},
	}

	err := pipeline.SavePlan(plan)
	if err != nil {
		t.Fatalf("SavePlan should create directories: %v", err)
	}

	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Error("plan file should be created")
	}
}

func TestPipeline_Timestamps(t *testing.T) {
	planResponse := AIPlanResponse{
		Site: AISiteInfo{Type: "blog"},
		Pages: []AIPageInfo{
			{ID: "home", Path: "content/_index.md", PageType: "home"},
			{ID: "about", Path: "content/about.md", PageType: "page"},
			{ID: "contact", Path: "content/contact.md", PageType: "page"},
			{ID: "post1", Path: "content/posts/welcome/index.md", PageType: "post"},
			{ID: "post2", Path: "content/posts/start/index.md", PageType: "post"},
			{ID: "post3", Path: "content/posts/another/index.md", PageType: "post"},
			{ID: "post4", Path: "content/posts/last/index.md", PageType: "post"},
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
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	client := NewClient("openai", "test-key", server.URL, "gpt-4")
	config := DefaultPipelineConfig()
	config.PlanPath = filepath.Join(tempDir, "plan.json")
	config.ContentDir = tempDir
	config.DryRun = true
	pipeline := NewPipeline(client, config)

	input := &PlannerInput{
		SiteName:    "My Blog",
		SiteType:    SiteTypeBlog,
		Description: "A blog about technology",
		Audience:    "developers",
	}

	startTime := time.Now()
	ctx := context.Background()
	result, _ := pipeline.Run(ctx, input)
	endTime := time.Now()

	if result.StartedAt.Before(startTime) || result.StartedAt.After(endTime) {
		t.Error("StartedAt should be within test bounds")
	}
	if result.FinishedAt.Before(result.StartedAt) {
		t.Error("FinishedAt should be after StartedAt")
	}
	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}
