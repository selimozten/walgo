package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Generator manages the content generation phase of the AI pipeline.

// Generator manages the content generation phase of the AI pipeline.
type Generator struct {
	client   *Client
	config   PipelineConfig
	progress ProgressHandler
}

// NewGenerator initializes and returns a new Generator instance with the provided client and configuration.
func NewGenerator(client *Client, config PipelineConfig) *Generator {
	return &Generator{
		client: client,
		config: config,
	}
}

// SetProgressHandler configures the progress callback function for tracking generation progress.
func (g *Generator) SetProgressHandler(handler ProgressHandler) {
	g.progress = handler
}

// Single Page Generation

// GeneratePage creates content for a single page with automatic retry logic and error handling.
func (g *Generator) GeneratePage(ctx context.Context, plan *SitePlan, page *PageSpec) *GeneratorOutput {
	startTime := time.Now()
	output := &GeneratorOutput{
		PageID:   page.ID,
		Path:     page.Path,
		Attempts: 0,
	}

	// Check if file already exists
	fullPath := filepath.Join(g.config.ContentDir, page.Path)
	_ = strings.TrimPrefix(fullPath, "content/") // Remove duplicate content/ (not used)
	fullPath = filepath.Join(g.config.ContentDir, strings.TrimPrefix(page.Path, "content/"))

	if !g.config.OverwriteExisting {
		if _, err := os.Stat(fullPath); err == nil {
			output.Success = true
			output.Skipped = true
			output.Duration = time.Since(startTime)
			g.emitProgress(ProgressSkip, page, "file already exists", plan)
			return output
		}
	}

	// Retry loop
	var lastErr error
	for attempt := 1; attempt <= g.config.MaxRetries; attempt++ {
		output.Attempts = attempt

		// Emit progress for attempt start
		if attempt > 1 {
			g.emitProgress(ProgressRetry, page,
				fmt.Sprintf("retry attempt %d/%d", attempt, g.config.MaxRetries), plan)
		} else {
			g.emitProgress(ProgressPageStart, page, "generating", plan)
		}

		// Apply timeout for this generation
		genCtx := ctx
		if g.config.GeneratorTimeout > 0 {
			var cancel context.CancelFunc
			genCtx, cancel = context.WithTimeout(ctx, g.config.GeneratorTimeout)
			defer cancel()
		}

		// Generate content
		content, err := g.generatePageContent(genCtx, plan, page)
		if err != nil {
			lastErr = err

			// Check if context was cancelled
			if ctx.Err() != nil {
				output.Error = ctx.Err()
				output.ErrorMsg = ctx.Err().Error()
				output.Duration = time.Since(startTime)
				return output
			}

			// Check if retryable
			if !IsRetryable(err) || attempt >= g.config.MaxRetries {
				break
			}

			// Wait before retry with exponential backoff
			backoffDelay := time.Duration(
				float64(g.config.RetryDelay) * float64(attempt) * g.config.RetryBackoffMulti,
			)

			select {
			case <-ctx.Done():
				output.Error = ctx.Err()
				output.ErrorMsg = ctx.Err().Error()
				output.Duration = time.Since(startTime)
				return output
			case <-time.After(backoffDelay):
				continue
			}
		}

		// Success!
		output.Content = content
		output.Success = true
		output.Duration = time.Since(startTime)

		// Write file unless dry run
		if !g.config.DryRun {
			if err := g.writeFile(fullPath, content); err != nil {
				output.Success = false
				output.Error = err
				output.ErrorMsg = fmt.Sprintf("failed to write file: %v", err)
			}
		}

		g.emitProgress(ProgressPageDone, page, "completed", plan)
		return output
	}

	// All retries exhausted
	output.Error = lastErr
	if lastErr != nil {
		output.ErrorMsg = fmt.Sprintf("generation failed after %d attempts: %v", output.Attempts, lastErr)
	} else {
		output.ErrorMsg = fmt.Sprintf("generation failed after %d attempts", output.Attempts)
	}
	output.Duration = time.Since(startTime)
	g.emitProgress(ProgressError, page, output.ErrorMsg, plan)

	return output
}

// Content Generation

// generatePageContent creates the actual content for a single page using the AI client.
func (g *Generator) generatePageContent(ctx context.Context, plan *SitePlan, page *PageSpec) (string, error) {
	// Build user prompt with context from plan
	userPrompt := BuildSinglePageUserPrompt(plan, page)

	// Generate via AI
	content, err := g.client.GenerateContentWithContext(ctx, SystemPromptSinglePageGenerator, userPrompt)
	if err != nil {
		return "", NewGeneratorError(page, page.Attempts, err, "AI generation failed")
	}

	// Clean the content
	content = CleanGeneratedContent(content)

	return content, nil
}

// File Operations

// writeFile saves content to the specified path, creating any necessary directories.
func (g *Generator) writeFile(path, content string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// Batch Generation

// GenerateAll processes all pending pages in the plan sequentially, updating their status as they are completed.
func (g *Generator) GenerateAll(ctx context.Context, plan *SitePlan) ([]GeneratorOutput, error) {
	results := make([]GeneratorOutput, 0, len(plan.Pages))

	for i := range plan.Pages {
		page := &plan.Pages[i]

		// Skip already completed or skipped pages
		if page.Status == PageStatusCompleted || page.Status == PageStatusSkipped {
			continue
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			return results, NewPipelineError(PhaseGenerating, ctx.Err(),
				"generation cancelled", len(results) > 0)
		default:
		}

		// Update page status
		page.Status = PageStatusInProgress

		// Generate the page
		output := g.GeneratePage(ctx, plan, page)
		results = append(results, *output)

		// Update page and plan status
		now := time.Now()
		page.Attempts = output.Attempts

		if output.Success {
			if output.Skipped {
				page.Status = PageStatusSkipped
				plan.Stats.SkippedPages++
			} else {
				page.Status = PageStatusCompleted
				page.GeneratedAt = &now
				plan.Stats.CompletedPages++
			}
		} else {
			if g.config.ContinueOnError {
				page.Status = PageStatusSkipped
				page.Error = output.ErrorMsg
				plan.Stats.SkippedPages++
			} else {
				page.Status = PageStatusFailed
				page.Error = output.ErrorMsg
				plan.Stats.FailedPages++
				return results, NewPipelineError(PhaseGenerating, output.Error,
					fmt.Sprintf("page %s failed", page.Path), true)
			}
		}

		plan.UpdatedAt = now
	}

	return results, nil
}

// Progress Emission

// emitProgress broadcasts a progress event if a progress handler has been configured.
func (g *Generator) emitProgress(eventType ProgressType, page *PageSpec, message string, plan *SitePlan) {
	if g.progress == nil {
		return
	}

	// Calculate progress
	var progress float64
	var current, total int

	if plan != nil {
		total = plan.Stats.TotalPages
		current = plan.Stats.CompletedPages + plan.Stats.FailedPages + plan.Stats.SkippedPages
		if total > 0 {
			progress = float64(current) / float64(total)
		}
	}

	event := ProgressEvent{
		Timestamp: time.Now(),
		Phase:     PhaseGenerating,
		EventType: eventType,
		Message:   message,
		Progress:  progress,
		Current:   current,
		Total:     total,
	}

	if page != nil {
		event.PageID = page.ID
		event.PagePath = page.Path
	}

	g.progress(event)
}
