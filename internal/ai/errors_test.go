package ai

import (
	"errors"
	"testing"
)

func TestPlannerError(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		input := &PlannerInput{SiteName: "test"}
		err := NewPlannerError(input, cause, "test message")

		if err.Input != input {
			t.Error("expected Input to be set")
		}
		if err.Cause != cause {
			t.Error("expected Cause to be set")
		}
		if err.Message != "test message" {
			t.Errorf("expected Message 'test message', got %s", err.Message)
		}

		errStr := err.Error()
		if errStr != "planner error: test message: underlying error" {
			t.Errorf("unexpected error string: %s", errStr)
		}

		// Test Unwrap
		if err.Unwrap() != cause {
			t.Error("Unwrap should return cause")
		}
	})

	t.Run("without cause", func(t *testing.T) {
		err := NewPlannerError(nil, nil, "test message")

		errStr := err.Error()
		if errStr != "planner error: test message" {
			t.Errorf("unexpected error string: %s", errStr)
		}

		if err.Unwrap() != nil {
			t.Error("Unwrap should return nil when no cause")
		}
	})
}

func TestGeneratorError(t *testing.T) {
	t.Run("with all fields", func(t *testing.T) {
		cause := errors.New("generation failed")
		page := &PageSpec{Path: "content/about.md"}
		err := NewGeneratorError(page, 2, cause, "test message")

		if err.PageSpec != page {
			t.Error("expected PageSpec to be set")
		}
		if err.Attempt != 2 {
			t.Errorf("expected Attempt 2, got %d", err.Attempt)
		}
		if err.Cause != cause {
			t.Error("expected Cause to be set")
		}
		if err.Message != "test message" {
			t.Errorf("expected Message 'test message', got %s", err.Message)
		}

		errStr := err.Error()
		expectedStr := "generator error for content/about.md (attempt 2): test message: generation failed"
		if errStr != expectedStr {
			t.Errorf("expected: %s\ngot: %s", expectedStr, errStr)
		}

		// Test Unwrap
		if err.Unwrap() != cause {
			t.Error("Unwrap should return cause")
		}
	})

	t.Run("without cause", func(t *testing.T) {
		page := &PageSpec{Path: "content/index.md"}
		err := NewGeneratorError(page, 1, nil, "test message")

		errStr := err.Error()
		expectedStr := "generator error for content/index.md (attempt 1): test message"
		if errStr != expectedStr {
			t.Errorf("expected: %s\ngot: %s", expectedStr, errStr)
		}
	})

	t.Run("nil PageSpec", func(t *testing.T) {
		err := NewGeneratorError(nil, 1, nil, "test message")

		errStr := err.Error()
		if errStr != "generator error for  (attempt 1): test message" {
			t.Errorf("unexpected error string: %s", errStr)
		}
	})
}

func TestPipelineError(t *testing.T) {
	t.Run("with all fields", func(t *testing.T) {
		cause := errors.New("pipeline failed")
		err := NewPipelineError(PhasePlanning, cause, "test message", true)

		if err.Phase != PhasePlanning {
			t.Errorf("expected Phase PhasePlanning, got %s", err.Phase)
		}
		if err.Cause != cause {
			t.Error("expected Cause to be set")
		}
		if err.Message != "test message" {
			t.Errorf("expected Message 'test message', got %s", err.Message)
		}
		if !err.Partial {
			t.Error("expected Partial to be true")
		}

		errStr := err.Error()
		expectedStr := "pipeline error in planning phase: test message: pipeline failed"
		if errStr != expectedStr {
			t.Errorf("expected: %s\ngot: %s", expectedStr, errStr)
		}

		// Test Unwrap
		if err.Unwrap() != cause {
			t.Error("Unwrap should return cause")
		}
	})

	t.Run("without cause", func(t *testing.T) {
		err := NewPipelineError(PhaseGenerating, nil, "test message", false)

		errStr := err.Error()
		expectedStr := "pipeline error in generating phase: test message"
		if errStr != expectedStr {
			t.Errorf("expected: %s\ngot: %s", expectedStr, errStr)
		}
	})
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("site_name", "", "site name is required")

	if err.Field != "site_name" {
		t.Errorf("expected Field 'site_name', got %s", err.Field)
	}
	if err.Value != "" {
		t.Errorf("expected Value '', got %v", err.Value)
	}
	if err.Message != "site name is required" {
		t.Errorf("expected Message 'site name is required', got %s", err.Message)
	}

	errStr := err.Error()
	expectedStr := "validation error for field 'site_name': site name is required (got: )"
	if errStr != expectedStr {
		t.Errorf("expected: %s\ngot: %s", expectedStr, errStr)
	}
}

func TestParseError(t *testing.T) {
	t.Run("with cause and short content", func(t *testing.T) {
		cause := errors.New("json syntax error")
		err := NewParseError("invalid json content", cause, "failed to parse")

		if err.Content != "invalid json content" {
			t.Errorf("expected Content 'invalid json content', got %s", err.Content)
		}
		if err.Cause != cause {
			t.Error("expected Cause to be set")
		}
		if err.Message != "failed to parse" {
			t.Errorf("expected Message 'failed to parse', got %s", err.Message)
		}

		errStr := err.Error()
		expectedStr := "parse error: failed to parse: json syntax error"
		if errStr != expectedStr {
			t.Errorf("expected: %s\ngot: %s", expectedStr, errStr)
		}

		// Test Unwrap
		if err.Unwrap() != cause {
			t.Error("Unwrap should return cause")
		}
	})

	t.Run("without cause", func(t *testing.T) {
		err := NewParseError("content", nil, "test message")

		errStr := err.Error()
		expectedStr := "parse error: test message"
		if errStr != expectedStr {
			t.Errorf("expected: %s\ngot: %s", expectedStr, errStr)
		}

		if err.Unwrap() != nil {
			t.Error("Unwrap should return nil when no cause")
		}
	})

	t.Run("long content truncation", func(t *testing.T) {
		longContent := ""
		for i := 0; i < 300; i++ {
			longContent += "x"
		}

		err := NewParseError(longContent, nil, "test")

		if len(err.Content) != 203 { // 200 + "..."
			t.Errorf("expected truncated content length 203, got %d", len(err.Content))
		}
		if err.Content[len(err.Content)-3:] != "..." {
			t.Error("expected truncated content to end with '...'")
		}
	})
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"timeout", errors.New("connection timeout"), true},
		{"deadline exceeded", errors.New("context deadline exceeded"), true},
		{"connection refused", errors.New("dial tcp: connection refused"), true},
		{"connection reset", errors.New("connection reset by peer"), true},
		{"temporarily unavailable", errors.New("service temporarily unavailable"), true},
		{"rate limit", errors.New("rate limit exceeded"), true},
		{"too many requests", errors.New("too many requests"), true},
		{"service unavailable", errors.New("service unavailable"), true},
		{"bad gateway", errors.New("bad gateway"), true},
		{"gateway timeout", errors.New("gateway timeout"), true},
		{"429 status", errors.New("status code 429"), true},
		{"502 status", errors.New("received 502"), true},
		{"503 status", errors.New("got 503 error"), true},
		{"504 status", errors.New("504 gateway timeout"), true},
		{"eof error", errors.New("unexpected EOF"), true},
		{"network error", errors.New("network is unreachable"), true},
		{"socket error", errors.New("socket connection error"), true},
		{"context cancelled", ErrContextCancelled, false},
		{"unauthorized", errors.New("unauthorized"), false},
		{"bad request", errors.New("bad request"), false},
		{"not found", errors.New("not found"), false},
		{"validation error", errors.New("validation failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		err := NewValidationError("field", "value", "message")
		if !IsValidationError(err) {
			t.Error("expected IsValidationError to return true")
		}
	})

	t.Run("wrapped validation error", func(t *testing.T) {
		innerErr := NewValidationError("field", "value", "message")
		wrappedErr := WrapError(innerErr, "additional context")
		if !IsValidationError(wrappedErr) {
			t.Error("expected IsValidationError to return true for wrapped error")
		}
	})

	t.Run("non-validation error", func(t *testing.T) {
		err := errors.New("regular error")
		if IsValidationError(err) {
			t.Error("expected IsValidationError to return false")
		}
	})
}

func TestIsPlannerError(t *testing.T) {
	t.Run("planner error", func(t *testing.T) {
		err := NewPlannerError(nil, nil, "message")
		if !IsPlannerError(err) {
			t.Error("expected IsPlannerError to return true")
		}
	})

	t.Run("wrapped planner error", func(t *testing.T) {
		innerErr := NewPlannerError(nil, nil, "message")
		wrappedErr := WrapError(innerErr, "context")
		if !IsPlannerError(wrappedErr) {
			t.Error("expected IsPlannerError to return true for wrapped error")
		}
	})

	t.Run("non-planner error", func(t *testing.T) {
		err := errors.New("regular error")
		if IsPlannerError(err) {
			t.Error("expected IsPlannerError to return false")
		}
	})
}

func TestIsGeneratorError(t *testing.T) {
	t.Run("generator error", func(t *testing.T) {
		err := NewGeneratorError(nil, 1, nil, "message")
		if !IsGeneratorError(err) {
			t.Error("expected IsGeneratorError to return true")
		}
	})

	t.Run("wrapped generator error", func(t *testing.T) {
		innerErr := NewGeneratorError(nil, 1, nil, "message")
		wrappedErr := WrapError(innerErr, "context")
		if !IsGeneratorError(wrappedErr) {
			t.Error("expected IsGeneratorError to return true for wrapped error")
		}
	})

	t.Run("non-generator error", func(t *testing.T) {
		err := errors.New("regular error")
		if IsGeneratorError(err) {
			t.Error("expected IsGeneratorError to return false")
		}
	})
}

func TestIsPipelineError(t *testing.T) {
	t.Run("pipeline error", func(t *testing.T) {
		err := NewPipelineError(PhasePlanning, nil, "message", false)
		if !IsPipelineError(err) {
			t.Error("expected IsPipelineError to return true")
		}
	})

	t.Run("wrapped pipeline error", func(t *testing.T) {
		innerErr := NewPipelineError(PhaseGenerating, nil, "message", true)
		wrappedErr := WrapError(innerErr, "context")
		if !IsPipelineError(wrappedErr) {
			t.Error("expected IsPipelineError to return true for wrapped error")
		}
	})

	t.Run("non-pipeline error", func(t *testing.T) {
		err := errors.New("regular error")
		if IsPipelineError(err) {
			t.Error("expected IsPipelineError to return false")
		}
	})
}

func TestIsPartialError(t *testing.T) {
	t.Run("partial pipeline error", func(t *testing.T) {
		err := NewPipelineError(PhaseGenerating, nil, "message", true)
		if !IsPartialError(err) {
			t.Error("expected IsPartialError to return true")
		}
	})

	t.Run("non-partial pipeline error", func(t *testing.T) {
		err := NewPipelineError(PhasePlanning, nil, "message", false)
		if IsPartialError(err) {
			t.Error("expected IsPartialError to return false")
		}
	})

	t.Run("wrapped partial error", func(t *testing.T) {
		innerErr := NewPipelineError(PhaseGenerating, nil, "message", true)
		wrappedErr := WrapError(innerErr, "context")
		if !IsPartialError(wrappedErr) {
			t.Error("expected IsPartialError to return true for wrapped error")
		}
	})

	t.Run("non-pipeline error", func(t *testing.T) {
		err := errors.New("regular error")
		if IsPartialError(err) {
			t.Error("expected IsPartialError to return false for regular error")
		}
	})
}

func TestWrapError(t *testing.T) {
	t.Run("wrap error", func(t *testing.T) {
		original := errors.New("original error")
		wrapped := WrapError(original, "additional context")

		if wrapped == nil {
			t.Fatal("expected non-nil error")
		}

		errStr := wrapped.Error()
		if errStr != "additional context: original error" {
			t.Errorf("unexpected error string: %s", errStr)
		}

		// Test unwrapping
		if !errors.Is(wrapped, original) {
			t.Error("wrapped error should contain original")
		}
	})

	t.Run("wrap nil", func(t *testing.T) {
		wrapped := WrapError(nil, "context")
		if wrapped != nil {
			t.Error("wrapping nil should return nil")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors are properly defined
	sentinelErrors := []struct {
		name string
		err  error
	}{
		{"ErrPlanNotFound", ErrPlanNotFound},
		{"ErrPlanInvalid", ErrPlanInvalid},
		{"ErrPlanVersionMismatch", ErrPlanVersionMismatch},
		{"ErrEmptyPlan", ErrEmptyPlan},
		{"ErrPageNotFound", ErrPageNotFound},
		{"ErrGenerationFailed", ErrGenerationFailed},
		{"ErrMaxRetriesExceeded", ErrMaxRetriesExceeded},
		{"ErrContextCancelled", ErrContextCancelled},
		{"ErrTimeout", ErrTimeout},
		{"ErrAIClientNotConfigured", ErrAIClientNotConfigured},
		{"ErrInvalidSiteType", ErrInvalidSiteType},
		{"ErrInvalidInput", ErrInvalidInput},
	}

	for _, tt := range sentinelErrors {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s has empty message", tt.name)
			}
		})
	}
}

func TestErrorsUnwrapping(t *testing.T) {
	// Test that errors properly support unwrapping chains
	cause := errors.New("root cause")

	plannerErr := NewPlannerError(nil, cause, "planner failed")
	if !errors.Is(plannerErr, cause) {
		t.Error("PlannerError should unwrap to cause")
	}

	generatorErr := NewGeneratorError(nil, 1, cause, "generator failed")
	if !errors.Is(generatorErr, cause) {
		t.Error("GeneratorError should unwrap to cause")
	}

	pipelineErr := NewPipelineError(PhasePlanning, cause, "pipeline failed", false)
	if !errors.Is(pipelineErr, cause) {
		t.Error("PipelineError should unwrap to cause")
	}

	parseErr := NewParseError("content", cause, "parse failed")
	if !errors.Is(parseErr, cause) {
		t.Error("ParseError should unwrap to cause")
	}
}
