package ai

import (
	"time"
)

// =============================================================================
// Site Types
// =============================================================================

// SiteType represents the category of site being generated.
type SiteType string

const (
	SiteTypeBlog      SiteType = "blog"
	SiteTypePortfolio SiteType = "portfolio"
	SiteTypeDocs      SiteType = "docs"
	SiteTypeBusiness  SiteType = "business"
)

// ValidSiteTypes returns all valid site types.
func ValidSiteTypes() []SiteType {
	return []SiteType{SiteTypeBlog, SiteTypePortfolio, SiteTypeDocs, SiteTypeBusiness}
}

// IsValid checks if the site type is valid.
func (s SiteType) IsValid() bool {
	switch s {
	case SiteTypeBlog, SiteTypePortfolio, SiteTypeDocs, SiteTypeBusiness:
		return true
	default:
		return false
	}
}

// =============================================================================
// Page Types
// =============================================================================

// PageType represents the type of page being generated.
type PageType string

const (
	PageTypeHome    PageType = "home"    // Home page (_index.md)
	PageTypePage    PageType = "page"    // Static page (about.md, contact.md)
	PageTypePost    PageType = "post"    // Blog post
	PageTypeSection PageType = "section" // Section index (_index.md in subdirectory)
	PageTypeDocs    PageType = "docs"    // Documentation page
)

// =============================================================================
// Status Types
// =============================================================================

// PlanStatus represents the current state of the plan execution.
type PlanStatus string

const (
	PlanStatusPending    PlanStatus = "pending"
	PlanStatusInProgress PlanStatus = "in_progress"
	PlanStatusCompleted  PlanStatus = "completed"
	PlanStatusFailed     PlanStatus = "failed"
	PlanStatusPartial    PlanStatus = "partial" // Some pages completed, some failed
)

// PageStatus represents the generation state of a single page.
type PageStatus string

const (
	PageStatusPending    PageStatus = "pending"
	PageStatusInProgress PageStatus = "in_progress"
	PageStatusCompleted  PageStatus = "completed"
	PageStatusFailed     PageStatus = "failed"
	PageStatusSkipped    PageStatus = "skipped"
)

// =============================================================================
// Site Plan
// =============================================================================

// SitePlan represents the complete site generation plan produced by the Planner.
// This is persisted to .walgo/plan.json for resumability.
type SitePlan struct {
	// Metadata
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Site Information
	SiteName    string   `json:"site_name"`
	SiteType    SiteType `json:"site_type"`
	Description string   `json:"description"`
	Audience    string   `json:"audience"`
	Theme       string   `json:"theme,omitempty"`
	BaseURL     string   `json:"base_url,omitempty"`
	Tone        string   `json:"tone,omitempty"`

	// Generation Plan
	Pages []PageSpec `json:"pages"`

	// Execution State
	Status      PlanStatus `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Statistics
	Stats PlanStats `json:"stats"`
}

// PlanStats holds statistics about plan execution.
type PlanStats struct {
	TotalPages     int `json:"total_pages"`
	CompletedPages int `json:"completed_pages"`
	FailedPages    int `json:"failed_pages"`
	SkippedPages   int `json:"skipped_pages"`
}

// PendingPages returns the count of pages not yet processed.
func (s PlanStats) PendingPages() int {
	return s.TotalPages - s.CompletedPages - s.FailedPages - s.SkippedPages
}

// Progress returns completion percentage (0.0 to 1.0).
func (s PlanStats) Progress() float64 {
	if s.TotalPages == 0 {
		return 0
	}
	return float64(s.CompletedPages+s.FailedPages+s.SkippedPages) / float64(s.TotalPages)
}

// =============================================================================
// Page Specification
// =============================================================================

// PageSpec defines a single page to be generated.
type PageSpec struct {
	// Identification
	ID   string `json:"id"`
	Path string `json:"path"` // Relative path (e.g., "content/about.md")

	// Content Specification
	Title       string   `json:"title"`
	PageType    PageType `json:"page_type"`
	ContentType string   `json:"content_type,omitempty"` // Hugo content type (posts, docs, etc.)
	Description string   `json:"description"`            // What the page should contain
	Keywords    []string `json:"keywords,omitempty"`
	WordCount   int      `json:"word_count,omitempty"` // Target word count (0 = default)

	// Hugo Frontmatter
	Frontmatter map[string]any `json:"frontmatter,omitempty"`

	// Internal Links (page IDs this page should link to)
	InternalLinks []string `json:"internal_links,omitempty"`

	// Execution State
	Status      PageStatus `json:"status"`
	Attempts    int        `json:"attempts"`
	Error       string     `json:"error,omitempty"`
	GeneratedAt *time.Time `json:"generated_at,omitempty"`
}

// IsCompleted returns true if the page was successfully generated.
func (p PageSpec) IsCompleted() bool {
	return p.Status == PageStatusCompleted
}

// IsPending returns true if the page is waiting to be generated.
func (p PageSpec) IsPending() bool {
	return p.Status == PageStatusPending
}

// NeedsGeneration returns true if the page should be processed.
func (p PageSpec) NeedsGeneration() bool {
	return p.Status == PageStatusPending || p.Status == PageStatusInProgress
}

// =============================================================================
// Pipeline Input/Output
// =============================================================================

// PlannerInput contains all inputs needed for the Planner phase.
type PlannerInput struct {
	SiteName    string   `json:"site_name"`
	SiteType    SiteType `json:"site_type"`
	Description string   `json:"description"`
	Audience    string   `json:"audience"`
	Features    string   `json:"features,omitempty"` // Specific pages/features requested
	Theme       string   `json:"theme,omitempty"`
	BaseURL     string   `json:"base_url,omitempty"`
	Tone        string   `json:"tone,omitempty"`
}

// GeneratorOutput contains the result of generating a single page.
type GeneratorOutput struct {
	PageID   string        `json:"page_id"`
	Path     string        `json:"path"`
	Content  string        `json:"content,omitempty"`
	Success  bool          `json:"success"`
	Error    error         `json:"-"`
	ErrorMsg string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Attempts int           `json:"attempts"`
	Skipped  bool          `json:"skipped,omitempty"`
}

// PipelineResult contains the complete result of a pipeline execution.
type PipelineResult struct {
	Plan       *SitePlan         `json:"plan"`
	Pages      []GeneratorOutput `json:"pages"`
	Success    bool              `json:"success"`
	Error      error             `json:"-"`
	ErrorMsg   string            `json:"error,omitempty"`
	Duration   time.Duration     `json:"duration"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt time.Time         `json:"finished_at"`
}

// =============================================================================
// Pipeline Configuration
// =============================================================================

// PipelineConfig configures the pipeline behavior.
type PipelineConfig struct {
	// Retry Configuration
	MaxRetries        int           `json:"max_retries"`
	RetryDelay        time.Duration `json:"retry_delay"`
	RetryBackoffMulti float64       `json:"retry_backoff_multi"`

	// Timeout Configuration
	PlannerTimeout   time.Duration `json:"planner_timeout"`
	GeneratorTimeout time.Duration `json:"generator_timeout"`

	// Behavior
	ContinueOnError   bool `json:"continue_on_error"`
	OverwriteExisting bool `json:"overwrite_existing"`
	DryRun            bool `json:"dry_run"`

	// Output Paths
	PlanPath   string `json:"plan_path"`
	ContentDir string `json:"content_dir"`

	// Verbosity
	Verbose bool `json:"verbose"`
}

// DefaultPipelineConfig returns the default pipeline configuration.
func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		MaxRetries:        3,
		RetryDelay:        2 * time.Second,
		RetryBackoffMulti: 2.0,
		PlannerTimeout:    5 * time.Minute,
		GeneratorTimeout:  2 * time.Minute,
		ContinueOnError:   true,
		OverwriteExisting: false,
		DryRun:            false,
		PlanPath:          ".walgo/plan.json",
		ContentDir:        "content",
		Verbose:           false,
	}
}

// =============================================================================
// Progress Tracking
// =============================================================================

// PipelinePhase indicates which phase of the pipeline is executing.
type PipelinePhase string

const (
	PhasePlanning   PipelinePhase = "planning"
	PhaseGenerating PipelinePhase = "generating"
	PhaseCompleted  PipelinePhase = "completed"
)

// ProgressType indicates the type of progress event.
type ProgressType string

const (
	ProgressStart     ProgressType = "start"
	ProgressUpdate    ProgressType = "update"
	ProgressPageStart ProgressType = "page_start"
	ProgressPageDone  ProgressType = "page_done"
	ProgressRetry     ProgressType = "retry"
	ProgressSkip      ProgressType = "skip"
	ProgressComplete  ProgressType = "complete"
	ProgressError     ProgressType = "error"
)

// ProgressEvent represents a progress update during pipeline execution.
type ProgressEvent struct {
	Timestamp time.Time     `json:"timestamp"`
	Phase     PipelinePhase `json:"phase"`
	EventType ProgressType  `json:"event_type"`
	PageID    string        `json:"page_id,omitempty"`
	PagePath  string        `json:"page_path,omitempty"`
	Message   string        `json:"message"`
	Progress  float64       `json:"progress"` // 0.0 to 1.0
	Current   int           `json:"current"`
	Total     int           `json:"total"`
}

// ProgressHandler is a callback function for progress updates.
type ProgressHandler func(event ProgressEvent)

// =============================================================================
// AI Response Parsing
// =============================================================================

// AIPlanResponse represents the JSON structure returned by the AI planner.
type AIPlanResponse struct {
	Site  AISiteInfo   `json:"site"`
	Pages []AIPageInfo `json:"pages"`
}

// AISiteInfo represents site information in AI response.
type AISiteInfo struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Language string `json:"language"`
	Tone     string `json:"tone"`
	BaseURL  string `json:"base_url"`
}

// AIPageInfo represents a single page in AI response.
type AIPageInfo struct {
	ID            string         `json:"id"`
	Path          string         `json:"path"`
	PageType      string         `json:"page_type"`
	Frontmatter   map[string]any `json:"frontmatter"`
	Outline       []string       `json:"outline"`
	InternalLinks []string       `json:"internal_links"`
}

// =============================================================================
// Minimum Page Requirements
// =============================================================================

// MinimumPageSet returns the minimum pages required for a site type.
func MinimumPageSet(siteType SiteType) []string {
	switch siteType {
	case SiteTypeBlog:
		return []string{
			"content/_index.md",
			"content/about.md",
			"content/contact.md",
			"content/posts/welcome/index.md",
			"content/posts/getting-started/index.md",
			"content/posts/latest-insights/index.md",
			"content/posts/case-study/index.md",
		}
	case SiteTypePortfolio:
		return []string{
			"content/_index.md",
			"content/about.md",
			"content/contact.md",
			"content/projects/_index.md",
			"content/projects/project-1.md",
			"content/projects/project-2.md",
			"content/projects/featured-work.md",
			"content/projects/tech-stack.md",
		}
	case SiteTypeDocs:
		return []string{
			"content/_index.md",
			"content/docs/_index.md",
			"content/docs/intro/index.md",
			"content/docs/install/index.md",
			"content/docs/usage/index.md",
			"content/docs/faq/index.md",
			"content/docs/quick-start/index.md",
			"content/docs/best-practices/index.md",
		}
	case SiteTypeBusiness:
		return []string{
			"content/_index.md",
			"content/about.md",
			"content/contact.md",
			"content/services/_index.md",
			"content/services/service-1.md",
			"content/services/service-2.md",
			"content/case-studies/index.md",
			"content/testimonials/index.md",
		}
	default:
		return []string{
			"content/_index.md",
			"content/about.md",
			"content/contact.md",
		}
	}
}

// MinimumPageCount returns the minimum number of pages for a site type.
func MinimumPageCount(siteType SiteType) int {
	return len(MinimumPageSet(siteType))
}

// =============================================================================
// Theme Configuration
// =============================================================================

// ThemeInfo contains information about a Hugo theme.
type ThemeInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RepoURL     string `json:"repo_url"`
	License     string `json:"license"`
}

// DefaultThemes maps site types to their default themes.
var DefaultThemes = map[SiteType]ThemeInfo{
	SiteTypeBlog: {
		Name:        "Ananke",
		Description: "Fast, clean, responsive theme for blogs and more",
		RepoURL:     "https://github.com/theNewDynamic/gohugo-theme-ananke",
		License:     "MIT",
	},
	SiteTypePortfolio: {
		Name:        "Ananke",
		Description: "Fast, clean, responsive theme for portfolios",
		RepoURL:     "https://github.com/theNewDynamic/gohugo-theme-ananke",
		License:     "MIT",
	},
	SiteTypeDocs: {
		Name:        "Book",
		Description: "Book-style documentation theme",
		RepoURL:     "https://github.com/alex-shpak/hugo-book",
		License:     "MIT",
	},
	SiteTypeBusiness: {
		Name:        "Ananke",
		Description: "Professional business theme",
		RepoURL:     "https://github.com/theNewDynamic/gohugo-theme-ananke",
		License:     "MIT",
	},
}

// GetDefaultTheme returns the default theme for a site type.
func GetDefaultTheme(siteType SiteType) ThemeInfo {
	if theme, ok := DefaultThemes[siteType]; ok {
		return theme
	}
	// Default to blog theme
	return DefaultThemes[SiteTypeBlog]
}

// GetThemeDirName returns the expected directory name for a theme.
func GetThemeDirName(themeName string) string {
	// Some themes have specific expected directory names
	switch themeName {
	case "Ananke":
		return "ananke"
	case "Book":
		return "hugo-book"
	default:
		return themeName
	}
}
