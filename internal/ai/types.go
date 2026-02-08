package ai

import (
	"strings"
	"time"
)

// =============================================================================
// Site Types
// =============================================================================

// SiteType represents the category of site being generated.
type SiteType string

const (
	SiteTypeBlog       SiteType = "blog"
	SiteTypeDocs       SiteType = "docs"
	SiteTypeBiolink    SiteType = "biolink"
	SiteTypeWhitepaper SiteType = "whitepaper"
)

// ValidSiteTypes returns all valid site types.
func ValidSiteTypes() []SiteType {
	return []SiteType{SiteTypeBlog, SiteTypeDocs, SiteTypeBiolink, SiteTypeWhitepaper}
}

// IsValid checks if the site type is valid.
func (s SiteType) IsValid() bool {
	switch s {
	case SiteTypeBlog, SiteTypeDocs, SiteTypeBiolink, SiteTypeWhitepaper:
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
	SitePath    string   `json:"site_path,omitempty"` // For dynamic theme analysis
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
	SitePath    string   `json:"site_path,omitempty"` // For dynamic theme analysis
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

	// Parallel Generation Configuration
	ParallelMode      ParallelMode  `json:"parallel_mode"`       // auto, sequential, parallel
	MaxConcurrent     int           `json:"max_concurrent"`      // Max concurrent generations (default: 5)
	RequestsPerMinute int           `json:"requests_per_minute"` // Rate limit (default: 30)
	RateLimitBackoff  time.Duration `json:"rate_limit_backoff"`  // Backoff on 429 (default: 30s)

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

// ParallelMode defines how page generation should be parallelized
type ParallelMode string

const (
	ParallelModeAuto       ParallelMode = "auto"       // Automatically determine based on page count
	ParallelModeSequential ParallelMode = "sequential" // Generate pages one by one
	ParallelModeParallel   ParallelMode = "parallel"   // Generate pages in parallel
)

// DefaultPipelineConfig returns the default pipeline configuration.
func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		MaxRetries:        3,
		RetryDelay:        2 * time.Second,
		RetryBackoffMulti: 2.0,
		PlannerTimeout:    5 * time.Minute,
		GeneratorTimeout:  2 * time.Minute,
		ParallelMode:      ParallelModeAuto,
		MaxConcurrent:     5,
		RequestsPerMinute: 30,
		RateLimitBackoff:  30 * time.Second,
		ContinueOnError:   true,
		OverwriteExisting: false,
		DryRun:            false,
		PlanPath:          ".walgo/plan.json",
		ContentDir:        "content",
		Verbose:           false,
	}
}

// DetermineParallelism returns the optimal concurrency level based on page count, mode, and RPM.
func (c *PipelineConfig) DetermineParallelism(pageCount int) int {
	switch c.ParallelMode {
	case ParallelModeSequential:
		return 1
	case ParallelModeParallel:
		return c.MaxConcurrent
	case ParallelModeAuto:
		fallthrough
	default:
		// Auto mode: determine based on page count
		var concurrency int
		if pageCount <= 3 {
			return 1 // Sequential - overhead not worth it
		}
		if pageCount <= 10 {
			concurrency = min(3, c.MaxConcurrent)
		} else if pageCount <= 20 {
			concurrency = min(5, c.MaxConcurrent)
		} else {
			concurrency = c.MaxConcurrent
		}
		// Cap concurrency so workers don't starve the rate limiter.
		// With RPM=30 (0.5 rps), more than 3 workers means each waits 6s+
		// between requests, which is wasteful. Cap at RPM/10 (minimum 2).
		if c.RequestsPerMinute > 0 {
			rpmCap := c.RequestsPerMinute / 10
			if rpmCap < 2 {
				rpmCap = 2
			}
			if concurrency > rpmCap {
				concurrency = rpmCap
			}
		}
		return concurrency
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

// MinimumPageSet returns the minimum essential pages required for any site type.
// Only a homepage is strictly required — the AI decides all other pages.
func MinimumPageSet(siteType SiteType) []string {
	return []string{
		"content/_index.md",
	}
}

// MinimumPageCount returns the minimum number of pages for a site type.
// Only a homepage is required — the AI decides the actual page count.
func MinimumPageCount(siteType SiteType) int {
	return 1 // only homepage is mandatory
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
	SiteTypeDocs: {
		Name:        "Book",
		Description: "Book-style documentation theme",
		RepoURL:     "https://github.com/alex-shpak/hugo-book",
		License:     "MIT",
	},
	SiteTypeBiolink: {
		Name:        "Walgo Biolink",
		Description: "Link-in-bio theme for blockchain professionals",
		RepoURL:     "https://github.com/ganbitlabs/walgo-biolink",
		License:     "MIT",
	},
	SiteTypeWhitepaper: {
		Name:        "Walgo Whitepaper",
		Description: "Academic whitepaper theme for blockchain projects",
		RepoURL:     "https://github.com/ganbitlabs/walgo-whitepaper",
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
// This function converts theme display names to their typical directory names.
// For custom themes, it returns the lowercase version of the name.
func GetThemeDirName(themeName string) string {
	// Common theme name mappings (these are well-known themes)
	// The system will work with any theme - these are just convenience mappings
	knownThemes := map[string]string{
		"Ananke":           "ananke",
		"Book":             "hugo-book",
		"PaperMod":         "PaperMod",
		"Stack":            "hugo-theme-stack",
		"Docsy":            "docsy",
		"Blowfish":         "blowfish",
		"Congo":            "congo",
		"Terminal":         "hugo-theme-terminal",
		"Academic":         "academic",
		"Mainroad":         "mainroad",
		"Beautiful":        "beautifulhugo",
		"Coder":            "hugo-coder",
		"Walgo Biolink":    "walgo-biolink",
		"Walgo Whitepaper": "walgo-whitepaper",
	}

	if dirName, ok := knownThemes[themeName]; ok {
		return dirName
	}

	// For unknown themes, use lowercase version
	return strings.ToLower(themeName)
}
