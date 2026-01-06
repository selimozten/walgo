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
		{SiteTypePortfolio, true},
		{SiteTypeDocs, true},
		{SiteTypeBusiness, true},
		{SiteType("invalid"), false},
		{SiteType(""), false},
		{SiteType("BLOG"), false}, // Case sensitive
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
		SiteTypeBlog:      true,
		SiteTypePortfolio: true,
		SiteTypeDocs:      true,
		SiteTypeBusiness:  true,
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
	tests := []struct {
		siteType SiteType
		minPages int
	}{
		{SiteTypeBlog, 6},
		{SiteTypePortfolio, 8},
		{SiteTypeDocs, 8},
		{SiteTypeBusiness, 8},
		{SiteType("unknown"), 3}, // Default
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
	tests := []struct {
		siteType SiteType
		expected int
	}{
		{SiteTypeBlog, 6},
		{SiteTypePortfolio, 8},
		{SiteTypeDocs, 8},
		{SiteTypeBusiness, 8},
		{SiteType("unknown"), 3},
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

func TestMinimumPageSet_BlogNewPages(t *testing.T) {
	blogPages := MinimumPageSet(SiteTypeBlog)
	requiredBlogPaths := []string{
		"content/_index.md",
		"content/about.md",
		"content/contact.md",
		"content/posts/welcome/index.md",
		"content/posts/getting-started/index.md",
		"content/posts/latest-insights/index.md",
		"content/posts/case-study/index.md",
	}

	for _, path := range requiredBlogPaths {
		found := false
		for _, p := range blogPages {
			if p == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("blog should include %s", path)
		}
	}
}

func TestMinimumPageSet_PortfolioNewPages(t *testing.T) {
	portfolioPages := MinimumPageSet(SiteTypePortfolio)
	requiredPortfolioPaths := []string{
		"content/_index.md",
		"content/about.md",
		"content/contact.md",
		"content/projects/_index.md",
		"content/projects/project-1.md",
		"content/projects/project-2.md",
		"content/projects/featured-work.md",
		"content/projects/tech-stack.md",
	}

	for _, path := range requiredPortfolioPaths {
		found := false
		for _, p := range portfolioPages {
			if p == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("portfolio should include %s", path)
		}
	}
}

func TestMinimumPageSet_BusinessNewPages(t *testing.T) {
	businessPages := MinimumPageSet(SiteTypeBusiness)
	requiredBusinessPaths := []string{
		"content/_index.md",
		"content/about.md",
		"content/contact.md",
		"content/services/_index.md",
		"content/services/service-1.md",
		"content/services/service-2.md",
		"content/case-studies/index.md",
		"content/testimonials/index.md",
	}

	for _, path := range requiredBusinessPaths {
		found := false
		for _, p := range businessPages {
			if p == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("business should include %s", path)
		}
	}
}

func TestMinimumPageSet_DocsNewPages(t *testing.T) {
	docsPages := MinimumPageSet(SiteTypeDocs)
	requiredDocsPaths := []string{
		"content/_index.md",
		"content/docs/_index.md",
		"content/docs/intro/index.md",
		"content/docs/install/index.md",
		"content/docs/usage/index.md",
		"content/docs/faq/index.md",
		"content/docs/quick-start/index.md",
		"content/docs/best-practices/index.md",
	}

	for _, path := range requiredDocsPaths {
		found := false
		for _, p := range docsPages {
			if p == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("docs should include %s", path)
		}
	}
}

func TestGetDefaultTheme(t *testing.T) {
	tests := []struct {
		siteType     SiteType
		expectedName string
	}{
		{SiteTypeBlog, "Ananke"},
		{SiteTypePortfolio, "Ananke"},
		{SiteTypeDocs, "Book"},
		{SiteTypeBusiness, "Ananke"},
		{SiteType("unknown"), "Ananke"}, // Defaults to blog theme
	}

	for _, tt := range tests {
		t.Run(string(tt.siteType), func(t *testing.T) {
			theme := GetDefaultTheme(tt.siteType)
			if theme.Name != tt.expectedName {
				t.Errorf("GetDefaultTheme(%s).Name = %s, want %s", tt.siteType, theme.Name, tt.expectedName)
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
	requiredTypes := []SiteType{SiteTypeBlog, SiteTypePortfolio, SiteTypeDocs, SiteTypeBusiness}

	for _, st := range requiredTypes {
		theme, exists := DefaultThemes[st]
		if !exists {
			t.Errorf("DefaultThemes missing entry for %s", st)
			continue
		}
		if theme.Name == "" {
			t.Errorf("DefaultThemes[%s].Name is empty", st)
		}
		if theme.RepoURL == "" {
			t.Errorf("DefaultThemes[%s].RepoURL is empty", st)
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
	if SiteTypePortfolio != "portfolio" {
		t.Errorf("expected SiteTypePortfolio = 'portfolio', got %s", SiteTypePortfolio)
	}
	if SiteTypeDocs != "docs" {
		t.Errorf("expected SiteTypeDocs = 'docs', got %s", SiteTypeDocs)
	}
	if SiteTypeBusiness != "business" {
		t.Errorf("expected SiteTypeBusiness = 'business', got %s", SiteTypeBusiness)
	}
}
