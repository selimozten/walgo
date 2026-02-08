package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Pipeline orchestrates the complete site generation workflow.
type Pipeline struct {
	client    *Client
	config    PipelineConfig
	planner   *Planner
	generator *Generator
	progress  ProgressHandler
}

// NewPipeline initializes and returns a new Pipeline instance with the provided AI client and configuration.
func NewPipeline(client *Client, config PipelineConfig) *Pipeline {
	p := &Pipeline{
		client:    client,
		config:    config,
		planner:   NewPlanner(client, config),
		generator: NewGenerator(client, config),
	}
	return p
}

// SetProgressHandler configures the progress callback handler for all pipeline phases.
func (p *Pipeline) SetProgressHandler(handler ProgressHandler) {
	p.progress = handler
	p.generator.SetProgressHandler(handler)
}

// Full Pipeline Execution

// Run executes the complete pipeline: plan, then generate all pages.
func (p *Pipeline) Run(ctx context.Context, input *PlannerInput) (*PipelineResult, error) {
	startTime := time.Now()
	result := &PipelineResult{
		StartedAt: startTime,
		Pages:     make([]GeneratorOutput, 0),
	}

	// Check for existing plan to resume
	existingPlan, err := p.LoadPlan()
	if err == nil && existingPlan != nil && existingPlan.Status != PlanStatusCompleted {
		// Resume from existing plan
		p.emitProgress(ProgressStart, PhasePlanning, "resuming from existing plan", nil, existingPlan)
		result.Plan = existingPlan
	} else {
		// Phase 1: Planning
		p.emitProgress(ProgressStart, PhasePlanning, "creating site plan", nil, nil)

		plan, err := p.planner.Plan(ctx, input)
		if err != nil {
			result.Error = err
			result.ErrorMsg = err.Error()
			result.FinishedAt = time.Now()
			result.Duration = time.Since(startTime)
			return result, NewPipelineError(PhasePlanning, err, "planning failed", false)
		}

		// Save the plan
		if err := p.SavePlan(plan); err != nil {
			// Non-fatal, continue with generation
			p.emitProgress(ProgressError, PhasePlanning,
				fmt.Sprintf("failed to save plan: %v", err), nil, plan)
		}

		result.Plan = plan
		p.emitProgress(ProgressComplete, PhasePlanning,
			fmt.Sprintf("plan created with %d pages", len(plan.Pages)), nil, plan)
	}

	// Phase 2: Generation & Routes
	result.Plan.Status = PlanStatusInProgress
	now := time.Now()
	result.Plan.StartedAt = &now

	p.emitProgress(ProgressStart, PhaseGenerating,
		fmt.Sprintf("generating %d pages", len(result.Plan.Pages)), nil, result.Plan)

	outputs, err := p.generator.GenerateAll(ctx, result.Plan)
	result.Pages = outputs

	// Update plan status
	if err != nil {
		var pipeErr *PipelineError
		if errors.As(err, &pipeErr) && pipeErr.Partial {
			result.Plan.Status = PlanStatusPartial
		} else {
			result.Plan.Status = PlanStatusFailed
		}
	} else {
		// Check if all pages succeeded
		if result.Plan.Stats.FailedPages > 0 || result.Plan.Stats.SkippedPages > 0 {
			result.Plan.Status = PlanStatusPartial
		} else {
			result.Plan.Status = PlanStatusCompleted
		}
	}

	completedAt := time.Now()
	result.Plan.CompletedAt = &completedAt
	result.Plan.UpdatedAt = completedAt
	result.FinishedAt = completedAt
	result.Duration = time.Since(startTime)
	result.Success = result.Plan.Status == PlanStatusCompleted

	if err != nil {
		result.Error = err
		result.ErrorMsg = err.Error()
	}

	// Save final plan state
	if saveErr := p.SavePlan(result.Plan); saveErr != nil {
		p.emitProgress(ProgressError, PhaseCompleted,
			fmt.Sprintf("failed to save final plan: %v", saveErr), nil, result.Plan)
	}

	// Emit completion
	summary := fmt.Sprintf("completed: %d/%d pages (%d skipped, %d failed)",
		result.Plan.Stats.CompletedPages,
		result.Plan.Stats.TotalPages,
		result.Plan.Stats.SkippedPages,
		result.Plan.Stats.FailedPages)
	p.emitProgress(ProgressComplete, PhaseCompleted, summary, nil, result.Plan)

	return result, err
}

// Individual Phase Execution

// PlanOnly executes only the planning phase without generating any content.
func (p *Pipeline) PlanOnly(ctx context.Context, input *PlannerInput) (*SitePlan, error) {
	p.emitProgress(ProgressStart, PhasePlanning, "creating site plan", nil, nil)

	plan, err := p.planner.Plan(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := p.SavePlan(plan); err != nil {
		return plan, fmt.Errorf("plan created but failed to save: %w", err)
	}

	p.emitProgress(ProgressComplete, PhasePlanning,
		fmt.Sprintf("plan created with %d pages", len(plan.Pages)), nil, plan)

	return plan, nil
}

// GenerateFromPlan executes content generation using an existing site plan.
func (p *Pipeline) GenerateFromPlan(ctx context.Context, plan *SitePlan) (*PipelineResult, error) {
	startTime := time.Now()
	result := &PipelineResult{
		StartedAt: startTime,
		Plan:      plan,
		Pages:     make([]GeneratorOutput, 0),
	}

	plan.Status = PlanStatusInProgress
	now := time.Now()
	plan.StartedAt = &now

	p.emitProgress(ProgressStart, PhaseGenerating,
		fmt.Sprintf("generating %d pages", len(plan.Pages)), nil, plan)

	outputs, err := p.generator.GenerateAll(ctx, plan)
	result.Pages = outputs

	// Update status
	if err != nil {
		var pipeErr *PipelineError
		if errors.As(err, &pipeErr) && pipeErr.Partial {
			plan.Status = PlanStatusPartial
		} else {
			plan.Status = PlanStatusFailed
		}
	} else {
		if plan.Stats.FailedPages > 0 || plan.Stats.SkippedPages > 0 {
			plan.Status = PlanStatusPartial
		} else {
			plan.Status = PlanStatusCompleted
		}
	}

	completedAt := time.Now()
	plan.CompletedAt = &completedAt
	plan.UpdatedAt = completedAt
	result.FinishedAt = completedAt
	result.Duration = time.Since(startTime)
	result.Success = plan.Status == PlanStatusCompleted

	if err != nil {
		result.Error = err
		result.ErrorMsg = err.Error()
	}

	// Save final plan state
	if saveErr := p.SavePlan(plan); saveErr != nil {
		p.emitProgress(ProgressError, PhaseCompleted,
			fmt.Sprintf("failed to save final plan: %v", saveErr), nil, plan)
	}

	// Emit completion
	summary := fmt.Sprintf("completed: %d/%d pages (%d skipped, %d failed)",
		plan.Stats.CompletedPages,
		plan.Stats.TotalPages,
		plan.Stats.SkippedPages,
		plan.Stats.FailedPages)
	p.emitProgress(ProgressComplete, PhaseCompleted, summary, nil, plan)

	return result, err
}

// Resume attempts to continue generation from a partially completed plan.
func (p *Pipeline) Resume(ctx context.Context) (*PipelineResult, error) {
	plan, err := p.LoadPlan()
	if err != nil {
		return nil, fmt.Errorf("no plan found to resume: %w", err)
	}

	if plan.Status == PlanStatusCompleted {
		return nil, fmt.Errorf("plan is already completed")
	}

	return p.GenerateFromPlan(ctx, plan)
}

// Plan Persistence

// LoadPlan loads a plan from the configured path.
func (p *Pipeline) LoadPlan() (*SitePlan, error) {
	planPath := p.getPlanPath()

	data, err := os.ReadFile(planPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("failed to read plan: %w", err)
	}

	var plan SitePlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPlanInvalid, err)
	}

	// Version check
	if plan.Version != "1.0" {
		return nil, fmt.Errorf("%w: expected 1.0, got %s", ErrPlanVersionMismatch, plan.Version)
	}

	return &plan, nil
}

// SavePlan saves a plan to the configured path.
func (p *Pipeline) SavePlan(plan *SitePlan) error {
	planPath := p.getPlanPath()

	// Ensure directory exists
	dir := filepath.Dir(planPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create plan directory: %w", err)
	}

	// Update timestamp
	plan.UpdatedAt = time.Now()

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	if err := os.WriteFile(planPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write plan: %w", err)
	}

	return nil
}

// DeletePlan removes the persisted plan file from disk.
func (p *Pipeline) DeletePlan() error {
	planPath := p.getPlanPath()

	if err := os.Remove(planPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	return nil
}

// HasPlan returns true if a plan file exists at the configured path.
func (p *Pipeline) HasPlan() bool {
	planPath := p.getPlanPath()
	_, err := os.Stat(planPath)
	return err == nil
}

// getPlanPath constructs and returns the full file system path for the plan file.
func (p *Pipeline) getPlanPath() string {
	if filepath.IsAbs(p.config.PlanPath) {
		return p.config.PlanPath
	}
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to home directory if cwd is unavailable
		if home, homeErr := os.UserHomeDir(); homeErr == nil {
			return filepath.Join(home, p.config.PlanPath)
		}
		return p.config.PlanPath
	}
	return filepath.Join(cwd, p.config.PlanPath)
}

// Progress Emission

// emitProgress broadcasts a progress event if a progress handler has been configured.
func (p *Pipeline) emitProgress(eventType ProgressType, phase PipelinePhase, message string, page *PageSpec, plan *SitePlan) {
	if p.progress == nil {
		return
	}

	event := ProgressEvent{
		Timestamp: time.Now(),
		Phase:     phase,
		EventType: eventType,
		Message:   message,
	}

	if page != nil {
		event.PageID = page.ID
		event.PagePath = page.Path
	}

	if plan != nil {
		event.Total = plan.Stats.TotalPages
		event.Current = plan.Stats.CompletedPages + plan.Stats.FailedPages + plan.Stats.SkippedPages
		if event.Total > 0 {
			event.Progress = float64(event.Current) / float64(event.Total)
		}
	}

	p.progress(event)
}

// Configuration Helpers

// WithConfig creates and returns a new Pipeline instance with the specified configuration overrides.
func (p *Pipeline) WithConfig(config PipelineConfig) *Pipeline {
	return NewPipeline(p.client, config)
}

// Config returns the current pipeline configuration.
func (p *Pipeline) Config() PipelineConfig {
	return p.config
}

// Client returns the AI client instance used by this pipeline.
func (p *Pipeline) Client() *Client {
	return p.client
}
