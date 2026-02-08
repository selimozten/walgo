package ai

import (
	"testing"
	"time"
)

func TestSiteType_IsValid(t *testing.T) {
	tests := []struct {
		siteType SiteType
		expected bool
	}{
		{SiteTypeBlog, true},
		{SiteTypeDocs, true},
		{SiteTypeBiolink, true},
		{SiteTypeWhitepaper, true},
		{SiteType("invalid"), false},
		{SiteType(""), false},
		{SiteType("BLOG"), false}, // Case sensitive
		{SiteType("portfolio"), false},
		{SiteType("business"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.siteType), func(t *testing.T) {
			result := tt.siteType.IsValid()
			if result != tt.expected {
				t.Errorf("SiteType(%q).IsValid() = %v, want %v", tt.siteType, result, tt.expected)
			}
		})
	}
}

func TestValidSiteTypes(t *testing.T) {
	types := ValidSiteTypes()

	if len(types) != 4 {
		t.Errorf("expected 4 site types, got %d", len(types))
	}

	expectedTypes := map[SiteType]bool{
		SiteTypeBlog:       true,
		SiteTypeDocs:       true,
		SiteTypeBiolink:    true,
		SiteTypeWhitepaper: true,
	}

	for _, st := range types {
		if !expectedTypes[st] {
			t.Errorf("unexpected site type: %s", st)
		}
	}
}

func TestPlanStats_PendingPages(t *testing.T) {
	tests := []struct {
		name     string
		stats    PlanStats
		expected int
	}{
		{
			name: "all pending",
			stats: PlanStats{
				TotalPages:     10,
				CompletedPages: 0,
				FailedPages:    0,
				SkippedPages:   0,
			},
			expected: 10,
		},
		{
			name: "mixed status",
			stats: PlanStats{
				TotalPages:     10,
				CompletedPages: 5,
				FailedPages:    2,
				SkippedPages:   1,
			},
			expected: 2,
		},
		{
			name: "all completed",
			stats: PlanStats{
				TotalPages:     10,
				CompletedPages: 10,
				FailedPages:    0,
				SkippedPages:   0,
			},
			expected: 0,
		},
		{
			name: "all processed",
			stats: PlanStats{
				TotalPages:     10,
				CompletedPages: 5,
				FailedPages:    3,
				SkippedPages:   2,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stats.PendingPages()
			if result != tt.expected {
				t.Errorf("PendingPages() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestPlanStats_Progress(t *testing.T) {
	tests := []struct {
		name     string
		stats    PlanStats
		expected float64
	}{
		{
			name: "zero total",
			stats: PlanStats{
				TotalPages: 0,
			},
			expected: 0,
		},
		{
			name: "none completed",
			stats: PlanStats{
				TotalPages:     10,
				CompletedPages: 0,
				FailedPages:    0,
				SkippedPages:   0,
			},
			expected: 0,
		},
		{
			name: "half completed",
			stats: PlanStats{
				TotalPages:     10,
				CompletedPages: 5,
				FailedPages:    0,
				SkippedPages:   0,
			},
			expected: 0.5,
		},
		{
			name: "all completed",
			stats: PlanStats{
				TotalPages:     10,
				CompletedPages: 10,
				FailedPages:    0,
				SkippedPages:   0,
			},
			expected: 1.0,
		},
		{
			name: "mixed status counts as progress",
			stats: PlanStats{
				TotalPages:     10,
				CompletedPages: 4,
				FailedPages:    3,
				SkippedPages:   3,
			},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stats.Progress()
			if result != tt.expected {
				t.Errorf("Progress() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestPageSpec_IsCompleted(t *testing.T) {
	tests := []struct {
		status   PageStatus
		expected bool
	}{
		{PageStatusCompleted, true},
		{PageStatusPending, false},
		{PageStatusInProgress, false},
		{PageStatusFailed, false},
		{PageStatusSkipped, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			page := PageSpec{Status: tt.status}
			result := page.IsCompleted()
			if result != tt.expected {
				t.Errorf("IsCompleted() with status %s = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestPageSpec_IsPending(t *testing.T) {
	tests := []struct {
		status   PageStatus
		expected bool
	}{
		{PageStatusPending, true},
		{PageStatusCompleted, false},
		{PageStatusInProgress, false},
		{PageStatusFailed, false},
		{PageStatusSkipped, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			page := PageSpec{Status: tt.status}
			result := page.IsPending()
			if result != tt.expected {
				t.Errorf("IsPending() with status %s = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestPageSpec_NeedsGeneration(t *testing.T) {
	tests := []struct {
		status   PageStatus
		expected bool
	}{
		{PageStatusPending, true},
		{PageStatusInProgress, true},
		{PageStatusCompleted, false},
		{PageStatusFailed, false},
		{PageStatusSkipped, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			page := PageSpec{Status: tt.status}
			result := page.NeedsGeneration()
			if result != tt.expected {
				t.Errorf("NeedsGeneration() with status %s = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestDefaultPipelineConfig(t *testing.T) {
	config := DefaultPipelineConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", config.MaxRetries)
	}
	if config.RetryDelay != 2*time.Second {
		t.Errorf("expected RetryDelay 2s, got %v", config.RetryDelay)
	}
	if config.RetryBackoffMulti != 2.0 {
		t.Errorf("expected RetryBackoffMulti 2.0, got %f", config.RetryBackoffMulti)
	}
	if config.PlannerTimeout != 5*time.Minute {
		t.Errorf("expected PlannerTimeout 5m, got %v", config.PlannerTimeout)
	}
	if config.GeneratorTimeout != 2*time.Minute {
		t.Errorf("expected GeneratorTimeout 2m, got %v", config.GeneratorTimeout)
	}
	if !config.ContinueOnError {
		t.Error("expected ContinueOnError to be true")
	}
	if config.OverwriteExisting {
		t.Error("expected OverwriteExisting to be false")
	}
	if config.DryRun {
		t.Error("expected DryRun to be false")
	}
	if config.PlanPath != ".walgo/plan.json" {
		t.Errorf("expected PlanPath '.walgo/plan.json', got %s", config.PlanPath)
	}
	if config.ContentDir != "content" {
		t.Errorf("expected ContentDir 'content', got %s", config.ContentDir)
	}
	if config.Verbose {
		t.Error("expected Verbose to be false")
	}
}

func TestMinimumPageSet(t *testing.T) {
	// Only homepage is required — AI decides all other pages
	tests := []struct {
		siteType SiteType
		minPages int
	}{
		{SiteTypeBlog, 1},
		{SiteTypeDocs, 1},
		{SiteTypeBiolink, 1},
		{SiteTypeWhitepaper, 1},
		{SiteType("unknown"), 1},
	}

	for _, tt := range tests {
		t.Run(string(tt.siteType), func(t *testing.T) {
			pages := MinimumPageSet(tt.siteType)
			if len(pages) != tt.minPages {
				t.Errorf("MinimumPageSet(%s) returned %d pages, want %d", tt.siteType, len(pages), tt.minPages)
			}
		})
	}
}

func TestMinimumPageCount(t *testing.T) {
	// Only homepage is required — AI decides all other pages
	tests := []struct {
		siteType SiteType
		expected int
	}{
		{SiteTypeBlog, 1},
		{SiteTypeDocs, 1},
		{SiteTypeBiolink, 1},
		{SiteTypeWhitepaper, 1},
		{SiteType("unknown"), 1},
	}

	for _, tt := range tests {
		t.Run(string(tt.siteType), func(t *testing.T) {
			count := MinimumPageCount(tt.siteType)
			if count != tt.expected {
				t.Errorf("MinimumPageCount(%s) = %d, want %d", tt.siteType, count, tt.expected)
			}
		})
	}
}

func TestMinimumPageSet_EssentialPages(t *testing.T) {
	// Only homepage is strictly required — AI decides the rest
	siteTypes := []SiteType{SiteTypeBlog, SiteTypeDocs, SiteTypeBiolink, SiteTypeWhitepaper}

	for _, siteType := range siteTypes {
		t.Run(string(siteType), func(t *testing.T) {
			pages := MinimumPageSet(siteType)

			if len(pages) != 1 {
				t.Errorf("%s: expected 1 page, got %d", siteType, len(pages))
			}

			if pages[0] != "content/_index.md" {
				t.Errorf("%s: expected content/_index.md, got %s", siteType, pages[0])
			}
		})
	}
}

func TestGetDefaultTheme(t *testing.T) {
	tests := []struct {
		siteType     SiteType
		expectedName string
		hasRepoURL   bool
	}{
		{SiteTypeBlog, "Ananke", true},
		{SiteTypeDocs, "Book", true},
		{SiteTypeBiolink, "Walgo Biolink", true},
		{SiteTypeWhitepaper, "Walgo Whitepaper", true},
		{SiteType("unknown"), "Ananke", true}, // Defaults to blog theme
	}

	for _, tt := range tests {
		t.Run(string(tt.siteType), func(t *testing.T) {
			theme := GetDefaultTheme(tt.siteType)
			if theme.Name != tt.expectedName {
				t.Errorf("GetDefaultTheme(%s).Name = %s, want %s", tt.siteType, theme.Name, tt.expectedName)
			}
			if tt.hasRepoURL && theme.RepoURL == "" {
				t.Errorf("GetDefaultTheme(%s).RepoURL is empty, want non-empty", tt.siteType)
			}
		})
	}
}

func TestGetThemeDirName(t *testing.T) {
	tests := []struct {
		themeName   string
		expectedDir string
	}{
		{"Ananke", "ananke"},
		{"Book", "hugo-book"},
		{"custom-theme", "custom-theme"}, // Unknown themes return as-is
	}

	for _, tt := range tests {
		t.Run(tt.themeName, func(t *testing.T) {
			result := GetThemeDirName(tt.themeName)
			if result != tt.expectedDir {
				t.Errorf("GetThemeDirName(%s) = %s, want %s", tt.themeName, result, tt.expectedDir)
			}
		})
	}
}

func TestDefaultThemes(t *testing.T) {
	// Verify all required site types have themes
	requiredTypes := []SiteType{SiteTypeBlog, SiteTypeDocs, SiteTypeBiolink, SiteTypeWhitepaper}

	for _, st := range requiredTypes {
		theme, exists := DefaultThemes[st]
		if !exists {
			t.Errorf("DefaultThemes missing entry for %s", st)
			continue
		}
		if theme.Name == "" {
			t.Errorf("DefaultThemes[%s].Name is empty", st)
		}
		if theme.License == "" {
			t.Errorf("DefaultThemes[%s].License is empty", st)
		}
	}
}

func TestPageTypeConstants(t *testing.T) {
	// Verify page type constants
	if PageTypeHome != "home" {
		t.Errorf("expected PageTypeHome = 'home', got %s", PageTypeHome)
	}
	if PageTypePage != "page" {
		t.Errorf("expected PageTypePage = 'page', got %s", PageTypePage)
	}
	if PageTypePost != "post" {
		t.Errorf("expected PageTypePost = 'post', got %s", PageTypePost)
	}
	if PageTypeSection != "section" {
		t.Errorf("expected PageTypeSection = 'section', got %s", PageTypeSection)
	}
	if PageTypeDocs != "docs" {
		t.Errorf("expected PageTypeDocs = 'docs', got %s", PageTypeDocs)
	}
}

func TestPlanStatusConstants(t *testing.T) {
	// Verify plan status constants
	if PlanStatusPending != "pending" {
		t.Errorf("expected PlanStatusPending = 'pending', got %s", PlanStatusPending)
	}
	if PlanStatusInProgress != "in_progress" {
		t.Errorf("expected PlanStatusInProgress = 'in_progress', got %s", PlanStatusInProgress)
	}
	if PlanStatusCompleted != "completed" {
		t.Errorf("expected PlanStatusCompleted = 'completed', got %s", PlanStatusCompleted)
	}
	if PlanStatusFailed != "failed" {
		t.Errorf("expected PlanStatusFailed = 'failed', got %s", PlanStatusFailed)
	}
	if PlanStatusPartial != "partial" {
		t.Errorf("expected PlanStatusPartial = 'partial', got %s", PlanStatusPartial)
	}
}

func TestPageStatusConstants(t *testing.T) {
	// Verify page status constants
	if PageStatusPending != "pending" {
		t.Errorf("expected PageStatusPending = 'pending', got %s", PageStatusPending)
	}
	if PageStatusInProgress != "in_progress" {
		t.Errorf("expected PageStatusInProgress = 'in_progress', got %s", PageStatusInProgress)
	}
	if PageStatusCompleted != "completed" {
		t.Errorf("expected PageStatusCompleted = 'completed', got %s", PageStatusCompleted)
	}
	if PageStatusFailed != "failed" {
		t.Errorf("expected PageStatusFailed = 'failed', got %s", PageStatusFailed)
	}
	if PageStatusSkipped != "skipped" {
		t.Errorf("expected PageStatusSkipped = 'skipped', got %s", PageStatusSkipped)
	}
}

func TestPipelinePhaseConstants(t *testing.T) {
	if PhasePlanning != "planning" {
		t.Errorf("expected PhasePlanning = 'planning', got %s", PhasePlanning)
	}
	if PhaseGenerating != "generating" {
		t.Errorf("expected PhaseGenerating = 'generating', got %s", PhaseGenerating)
	}
	if PhaseCompleted != "completed" {
		t.Errorf("expected PhaseCompleted = 'completed', got %s", PhaseCompleted)
	}
}

func TestProgressTypeConstants(t *testing.T) {
	expectedTypes := map[ProgressType]string{
		ProgressStart:     "start",
		ProgressUpdate:    "update",
		ProgressPageStart: "page_start",
		ProgressPageDone:  "page_done",
		ProgressRetry:     "retry",
		ProgressSkip:      "skip",
		ProgressComplete:  "complete",
		ProgressError:     "error",
	}

	for pt, expected := range expectedTypes {
		if string(pt) != expected {
			t.Errorf("expected %s = '%s', got '%s'", expected, expected, pt)
		}
	}
}

func TestSiteTypeConstants(t *testing.T) {
	if SiteTypeBlog != "blog" {
		t.Errorf("expected SiteTypeBlog = 'blog', got %s", SiteTypeBlog)
	}
	if SiteTypeDocs != "docs" {
		t.Errorf("expected SiteTypeDocs = 'docs', got %s", SiteTypeDocs)
	}
	if SiteTypeBiolink != "biolink" {
		t.Errorf("expected SiteTypeBiolink = 'biolink', got %s", SiteTypeBiolink)
	}
	if SiteTypeWhitepaper != "whitepaper" {
		t.Errorf("expected SiteTypeWhitepaper = 'whitepaper', got %s", SiteTypeWhitepaper)
	}
}
