package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// Generator manages the content generation phase of the AI pipeline.
type Generator struct {
	client      *Client
	config      PipelineConfig
	progress    ProgressHandler
	rateLimiter *rate.Limiter
	mu          sync.Mutex // protects plan stats updates
}

// NewGenerator initializes and returns a new Generator instance with the provided client and configuration.
func NewGenerator(client *Client, config PipelineConfig) *Generator {
	// Create rate limiter: tokens per second = RequestsPerMinute / 60
	rps := float64(config.RequestsPerMinute) / 60.0
	if rps <= 0 {
		rps = 0.5 // default to 30 RPM
	}
	// Burst allows concurrent workers to start without unnecessary queuing,
	// while still respecting the sustained rate over time.
	burst := config.MaxConcurrent
	if burst < 1 {
		burst = 1
	}
	limiter := rate.NewLimiter(rate.Limit(rps), burst)

	return &Generator{
		client:      client,
		config:      config,
		rateLimiter: limiter,
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

	// Check if file already exists (strip "content/" prefix if present to avoid duplication)
	fullPath := filepath.Join(g.config.ContentDir, strings.TrimPrefix(page.Path, "content/"))

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
			msg := fmt.Sprintf("retry attempt %d/%d (last error: %v)", attempt, g.config.MaxRetries, lastErr)
			g.emitProgress(ProgressRetry, page, msg, plan)
		} else {
			g.emitProgress(ProgressPageStart, page, "generating", plan)
		}

		// Apply timeout for this generation attempt
		genCtx := ctx
		var cancelTimeout context.CancelFunc
		if g.config.GeneratorTimeout > 0 {
			genCtx, cancelTimeout = context.WithTimeout(ctx, g.config.GeneratorTimeout)
		}
		// Ensure cancel is called after each attempt to release resources
		cancelFunc := func() {
			if cancelTimeout != nil {
				cancelTimeout()
			}
		}

		// Generate content
		content, err := g.generatePageContent(genCtx, plan, page)
		cancelFunc() // Release timeout resources after each attempt
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

			// Use longer backoff for rate limit errors (429)
			backoffDelay := time.Duration(
				float64(g.config.RetryDelay) * float64(attempt) * g.config.RetryBackoffMulti,
			)
			if isRateLimitError(err) && g.config.RateLimitBackoff > backoffDelay {
				backoffDelay = g.config.RateLimitBackoff
			}

			select {
			case <-ctx.Done():
				output.Error = ctx.Err()
				output.ErrorMsg = ctx.Err().Error()
				output.Duration = time.Since(startTime)
				return output
			case <-time.After(backoffDelay):
			}

			// Wait for rate limiter before retry to avoid flooding
			if err := g.rateLimiter.Wait(ctx); err != nil {
				output.Error = ctx.Err()
				output.ErrorMsg = "rate limiter cancelled"
				output.Duration = time.Since(startTime)
				return output
			}
			continue
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

	return output
}

// Content Generation

// generatePageContent creates the actual content for a single page using the AI client.
func (g *Generator) generatePageContent(ctx context.Context, plan *SitePlan, page *PageSpec) (string, error) {
	// Run analysis once, reuse for both theme context and frontmatter fields
	themeContext := ""
	var frontmatterFields []string
	if plan.SitePath != "" && plan.Theme != "" {
		themeAnalysis := AnalyzeTheme(plan.SitePath, plan.Theme)
		configAnalysis := AnalyzeThemeConfig(plan.SitePath, plan.Theme)
		contentPatterns := AnalyzeSiteContent(plan.SitePath)

		themeContext = BuildThemeContextFromAnalysis(plan.Theme, themeAnalysis, configAnalysis, contentPatterns)

		section := determineSectionFromPath(page.Path)
		frontmatterFields = GetRecommendedFrontmatter(themeAnalysis, contentPatterns, section)
	}

	// Theme context only in system prompt; per-page fields in user prompt
	userPrompt := BuildSinglePageUserPrompt(plan, page, frontmatterFields)
	systemPrompt := ComposePageGeneratorPrompt(themeContext)

	// Generate via AI
	content, err := g.client.GenerateContentWithContext(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", NewGeneratorError(page, page.Attempts, err, "AI generation failed")
	}

	// Clean the content
	content = CleanGeneratedContent(content)

	return content, nil
}

// isRateLimitError checks if an error is a rate limit (HTTP 429) error.
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "429")
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

// GenerateAll processes all pending pages in the plan, using parallel or sequential mode based on config.
func (g *Generator) GenerateAll(ctx context.Context, plan *SitePlan) ([]GeneratorOutput, error) {
	// Count pending pages
	pendingCount := 0
	for _, page := range plan.Pages {
		if page.Status != PageStatusCompleted && page.Status != PageStatusSkipped {
			pendingCount++
		}
	}

	// Determine parallelism level
	concurrency := g.config.DetermineParallelism(pendingCount)

	if concurrency > 1 {
		return g.generateAllParallel(ctx, plan, concurrency)
	}
	return g.generateAllSequential(ctx, plan)
}

// generateAllSequential processes pages one by one (original behavior)
func (g *Generator) generateAllSequential(ctx context.Context, plan *SitePlan) ([]GeneratorOutput, error) {
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

		// Wait for rate limiter
		if err := g.rateLimiter.Wait(ctx); err != nil {
			return results, NewPipelineError(PhaseGenerating, err,
				"rate limiter cancelled", len(results) > 0)
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
			page.Status = PageStatusFailed
			page.Error = output.ErrorMsg
			plan.Stats.FailedPages++
		}

		plan.UpdatedAt = now

		// Emit progress AFTER stats update so counts are accurate
		if output.Success {
			g.emitProgress(ProgressPageDone, page, "completed", plan)
		} else {
			g.emitProgress(ProgressError, page, output.ErrorMsg, plan)
			if !g.config.ContinueOnError {
				return results, NewPipelineError(PhaseGenerating, output.Error,
					fmt.Sprintf("page %s failed", page.Path), true)
			}
		}
	}

	return results, nil
}

// generateAllParallel processes pages concurrently with rate limiting
func (g *Generator) generateAllParallel(ctx context.Context, plan *SitePlan, concurrency int) ([]GeneratorOutput, error) {
	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, concurrency)

	// Collect pending pages
	var pendingPages []*PageSpec
	var pendingIndices []int
	for i := range plan.Pages {
		if plan.Pages[i].Status != PageStatusCompleted && plan.Pages[i].Status != PageStatusSkipped {
			pendingPages = append(pendingPages, &plan.Pages[i])
			pendingIndices = append(pendingIndices, i)
		}
	}

	// Results channel and slice
	resultsChan := make(chan struct {
		index  int
		output GeneratorOutput
	}, len(pendingPages))

	// Error tracking
	var firstError atomic.Value
	var cancelOnce sync.Once
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// WaitGroup for goroutines
	var wg sync.WaitGroup

	// Emit parallel start progress
	g.emitProgress(ProgressStart, nil,
		fmt.Sprintf("generating %d pages with %d workers", len(pendingPages), concurrency), plan)

	// Launch goroutines for each pending page
	for i, page := range pendingPages {
		wg.Add(1)
		go func(idx int, p *PageSpec, pageIndex int) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			// Wait for rate limiter
			if err := g.rateLimiter.Wait(ctx); err != nil {
				if ctx.Err() == nil {
					// Rate limit error, not cancellation
					g.mu.Lock()
					p.Status = PageStatusFailed
					p.Error = fmt.Sprintf("rate limit error: %v", err)
					plan.Stats.FailedPages++
					g.mu.Unlock()
				}
				return
			}

			// Check for cancellation before starting
			if ctx.Err() != nil {
				return
			}

			// Update page status (thread-safe)
			g.mu.Lock()
			p.Status = PageStatusInProgress
			g.mu.Unlock()

			// Generate the page
			output := g.GeneratePage(ctx, plan, p)

			// Update page and plan status (thread-safe)
			g.mu.Lock()
			now := time.Now()
			p.Attempts = output.Attempts

			if output.Success {
				if output.Skipped {
					p.Status = PageStatusSkipped
					plan.Stats.SkippedPages++
				} else {
					p.Status = PageStatusCompleted
					p.GeneratedAt = &now
					plan.Stats.CompletedPages++
				}
			} else {
				p.Status = PageStatusFailed
				p.Error = output.ErrorMsg
				plan.Stats.FailedPages++
				if !g.config.ContinueOnError {
					// Store first error and cancel remaining
					firstError.CompareAndSwap(nil, output.Error)
					cancelOnce.Do(func() { cancel() })
				}
			}
			plan.UpdatedAt = now
			g.mu.Unlock()

			// Emit progress AFTER stats update so counts are accurate
			if output.Success {
				g.emitProgress(ProgressPageDone, p, "completed", plan)
			} else {
				g.emitProgress(ProgressError, p, output.ErrorMsg, plan)
			}

			// Send result
			resultsChan <- struct {
				index  int
				output GeneratorOutput
			}{idx, *output}
		}(i, page, pendingIndices[i])
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results directly into ordered slice
	results := make([]GeneratorOutput, len(pendingPages))
	for r := range resultsChan {
		results[r.index] = r.output
	}

	// Check for errors
	if err := firstError.Load(); err != nil {
		if e, ok := err.(error); ok {
			return results, NewPipelineError(PhaseGenerating, e,
				"parallel generation failed", len(results) > 0)
		}
	}

	return results, nil
}

// Progress Emission

// emitProgress broadcasts a progress event if a progress handler has been configured.
// Thread-safe: acquires g.mu to read plan.Stats consistently during parallel generation.
func (g *Generator) emitProgress(eventType ProgressType, page *PageSpec, message string, plan *SitePlan) {
	if g.progress == nil {
		return
	}

	// Calculate progress under lock (plan.Stats is updated by parallel goroutines)
	var progress float64
	var current, total int

	if plan != nil {
		g.mu.Lock()
		total = plan.Stats.TotalPages
		current = plan.Stats.CompletedPages + plan.Stats.FailedPages + plan.Stats.SkippedPages
		g.mu.Unlock()
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
