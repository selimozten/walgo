package ai

import (
	"errors"
	"fmt"
	"strings"
)

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// Plan errors
	ErrPlanNotFound        = errors.New("plan not found")
	ErrPlanInvalid         = errors.New("plan is invalid or corrupted")
	ErrPlanVersionMismatch = errors.New("plan version mismatch")
	ErrEmptyPlan           = errors.New("plan contains no pages")

	// Page errors
	ErrPageNotFound = errors.New("page not found in plan")

	// Generation errors
	ErrGenerationFailed   = errors.New("content generation failed")
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")

	// Context errors
	ErrContextCancelled = errors.New("operation cancelled")
	ErrTimeout          = errors.New("operation timed out")

	// Configuration errors
	ErrAIClientNotConfigured = errors.New("AI client not configured")
	ErrInvalidSiteType       = errors.New("invalid site type")
	ErrInvalidInput          = errors.New("invalid input")
)

// =============================================================================
// Planner Error
// =============================================================================

// PlannerError represents an error that occurred during the planning phase.
type PlannerError struct {
	Input   *PlannerInput
	Cause   error
	Message string
}

func (e *PlannerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("planner error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("planner error: %s", e.Message)
}

func (e *PlannerError) Unwrap() error {
	return e.Cause
}

// NewPlannerError creates a new PlannerError.
func NewPlannerError(input *PlannerInput, cause error, message string) *PlannerError {
	return &PlannerError{
		Input:   input,
		Cause:   cause,
		Message: message,
	}
}

// =============================================================================
// Generator Error
// =============================================================================

// GeneratorError represents an error that occurred during page generation.
type GeneratorError struct {
	PageSpec *PageSpec
	Attempt  int
	Cause    error
	Message  string
}

func (e *GeneratorError) Error() string {
	pagePath := ""
	if e.PageSpec != nil {
		pagePath = e.PageSpec.Path
	}
	if e.Cause != nil {
		return fmt.Sprintf("generator error for %s (attempt %d): %s: %v",
			pagePath, e.Attempt, e.Message, e.Cause)
	}
	return fmt.Sprintf("generator error for %s (attempt %d): %s",
		pagePath, e.Attempt, e.Message)
}

func (e *GeneratorError) Unwrap() error {
	return e.Cause
}

// NewGeneratorError creates a new GeneratorError.
func NewGeneratorError(page *PageSpec, attempt int, cause error, message string) *GeneratorError {
	return &GeneratorError{
		PageSpec: page,
		Attempt:  attempt,
		Cause:    cause,
		Message:  message,
	}
}

// =============================================================================
// Pipeline Error
// =============================================================================

// PipelineError represents a pipeline-level error.
type PipelineError struct {
	Phase   PipelinePhase
	Cause   error
	Message string
	Partial bool // True if some pages were generated before failure
}

func (e *PipelineError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("pipeline error in %s phase: %s: %v", e.Phase, e.Message, e.Cause)
	}
	return fmt.Sprintf("pipeline error in %s phase: %s", e.Phase, e.Message)
}

func (e *PipelineError) Unwrap() error {
	return e.Cause
}

// NewPipelineError creates a new PipelineError.
func NewPipelineError(phase PipelinePhase, cause error, message string, partial bool) *PipelineError {
	return &PipelineError{
		Phase:   phase,
		Cause:   cause,
		Message: message,
		Partial: partial,
	}
}

// =============================================================================
// Validation Error
// =============================================================================

// ValidationError represents a validation failure.
type ValidationError struct {
	Field   string
	Value   any
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (got: %v)", e.Field, e.Message, e.Value)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field string, value any, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// =============================================================================
// Parse Error
// =============================================================================

// ParseError represents a JSON parsing error.
type ParseError struct {
	Content string
	Cause   error
	Message string
}

func (e *ParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("parse error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

// NewParseError creates a new ParseError.
func NewParseError(content string, cause error, message string) *ParseError {
	// Truncate content for logging
	if len(content) > 200 {
		content = content[:200] + "..."
	}
	return &ParseError{
		Content: content,
		Cause:   cause,
		Message: message,
	}
}

// =============================================================================
// Error Helpers
// =============================================================================

// IsRetryable returns true if the error is potentially transient and worth retrying.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation (not retryable)
	if errors.Is(err, ErrContextCancelled) {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Retryable patterns
	retryablePatterns := []string{
		"timeout",
		"deadline exceeded",
		"connection refused",
		"connection reset",
		"temporarily unavailable",
		"rate limit",
		"too many requests",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
		"429",
		"502",
		"503",
		"504",
		"eof",
		"network",
		"socket",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// IsValidationError checks if the error is a validation error.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// IsPlannerError checks if the error is a planner error.
func IsPlannerError(err error) bool {
	var pe *PlannerError
	return errors.As(err, &pe)
}

// IsGeneratorError checks if the error is a generator error.
func IsGeneratorError(err error) bool {
	var ge *GeneratorError
	return errors.As(err, &ge)
}

// IsPipelineError checks if the error is a pipeline error.
func IsPipelineError(err error) bool {
	var pe *PipelineError
	return errors.As(err, &pe)
}

// IsPartialError checks if the error indicates partial completion.
func IsPartialError(err error) bool {
	var pe *PipelineError
	if errors.As(err, &pe) {
		return pe.Partial
	}
	return false
}

// WrapError wraps an error with additional context.
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
