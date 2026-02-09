package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Planner manages the planning phase of the AI content generation pipeline.
type Planner struct {
	client *Client
	config PipelineConfig
}

// NewPlanner initializes and returns a new Planner instance with the provided client and configuration.
func NewPlanner(client *Client, config PipelineConfig) *Planner {
	return &Planner{
		client: client,
		config: config,
	}
}

// Plan generates a site plan based on the input.
func (p *Planner) Plan(ctx context.Context, input *PlannerInput) (*SitePlan, error) {
	// Validate input
	if err := p.validateInput(input); err != nil {
		return nil, NewPlannerError(input, err, "invalid input")
	}

	// Build the prompt with dynamic theme analysis if sitePath is available
	systemPrompt := SystemPromptSitePlanner
	userPrompt := BuildSitePlannerPromptWithTheme(
		input.SiteName,
		string(input.SiteType),
		input.Description,
		input.Audience,
		input.Tone,
		input.BaseURL,
		input.SitePath,
		input.Theme,
	)

	// Apply timeout
	if p.config.PlannerTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.PlannerTimeout)
		defer cancel()
	}

	// Generate plan via AI
	response, err := p.client.GenerateContentWithContext(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, NewPlannerError(input, err, "AI generation failed")
	}

	// Parse JSON response
	plan, err := p.parsePlanResponse(response, input)
	if err != nil {
		return nil, NewPlannerError(input, err, "failed to parse AI response")
	}

	// Validate the plan
	if err := p.validatePlan(plan); err != nil {
		return nil, NewPlannerError(input, err, "plan validation failed")
	}

	return plan, nil
}

// Input Validation

// validateInput validates the provided planner input parameters and returns an error if validation fails.
func (p *Planner) validateInput(input *PlannerInput) error {
	if input == nil {
		return NewValidationError("input", nil, "input is required")
	}

	if strings.TrimSpace(input.SiteName) == "" {
		return NewValidationError("site_name", input.SiteName, "site name is required")
	}

	if !input.SiteType.IsValid() {
		return NewValidationError("site_type", input.SiteType,
			fmt.Sprintf("invalid site type, must be one of: %v", ValidSiteTypes()))
	}

	if strings.TrimSpace(input.Description) == "" {
		return NewValidationError("description", input.Description, "description is required")
	}

	return nil
}

// Response Parsing

// parsePlanResponse converts the raw AI response string into a structured SitePlan object.
func (p *Planner) parsePlanResponse(response string, input *PlannerInput) (*SitePlan, error) {
	// Clean the response (remove markdown code fences if present)
	response = CleanJSONResponse(response)

	// Try to parse as our expected format
	var aiPlan AIPlanResponse
	if err := json.Unmarshal([]byte(response), &aiPlan); err != nil {
		// Try alternative parsing
		return p.parseAlternativeFormat(response, input, err)
	}

	return p.convertAIPlanToSitePlan(&aiPlan, input)
}

// parseAlternativeFormat attempts to parse AI responses in alternative JSON format structures.
func (p *Planner) parseAlternativeFormat(response string, input *PlannerInput, originalErr error) (*SitePlan, error) {
	// Try parsing as a simple pages array
	var simpleFormat struct {
		SiteName    string       `json:"site_name"`
		SiteType    string       `json:"site_type"`
		Description string       `json:"description"`
		Pages       []AIPageInfo `json:"pages"`
	}

	if err := json.Unmarshal([]byte(response), &simpleFormat); err != nil {
		return nil, NewParseError(response, originalErr, "could not parse AI response as JSON")
	}

	// Convert simple format to our structure
	aiPlan := &AIPlanResponse{
		Site: AISiteInfo{
			Type:  simpleFormat.SiteType,
			Title: simpleFormat.SiteName,
		},
		Pages: simpleFormat.Pages,
	}

	return p.convertAIPlanToSitePlan(aiPlan, input)
}

// convertAIPlanToSitePlan transforms the AI-generated plan structure into the internal SitePlan format.
func (p *Planner) convertAIPlanToSitePlan(aiPlan *AIPlanResponse, input *PlannerInput) (*SitePlan, error) {
	now := time.Now()

	plan := &SitePlan{
		ID:          uuid.New().String(),
		Version:     "1.0",
		CreatedAt:   now,
		UpdatedAt:   now,
		SiteName:    input.SiteName,
		SiteType:    input.SiteType,
		Description: input.Description,
		Audience:    input.Audience,
		Theme:       input.Theme,
		SitePath:    input.SitePath, // For dynamic theme analysis during generation
		BaseURL:     input.BaseURL,
		Tone:        input.Tone,
		Status:      PlanStatusPending,
		Pages:       make([]PageSpec, 0, len(aiPlan.Pages)),
		Stats: PlanStats{
			TotalPages: len(aiPlan.Pages),
		},
	}

	// Override with AI response if available
	if aiPlan.Site.Tone != "" {
		plan.Tone = aiPlan.Site.Tone
	}
	if aiPlan.Site.BaseURL != "" {
		plan.BaseURL = aiPlan.Site.BaseURL
	}

	// Convert pages
	for i, aiPage := range aiPlan.Pages {
		pageID := aiPage.ID
		if pageID == "" {
			pageID = fmt.Sprintf("page_%d", i+1)
		}

		// Determine page type from page type
		pageType := p.determinePageType(aiPage.PageType, aiPage.Path)

		// Build description from outline if not provided
		description := ""
		if len(aiPage.Outline) > 0 {
			description = strings.Join(aiPage.Outline, " ")
		}

		// Extract title from frontmatter
		title := ""
		if t, ok := aiPage.Frontmatter["title"].(string); ok {
			title = t
		}

		page := PageSpec{
			ID:            pageID,
			Path:          aiPage.Path,
			Title:         title,
			PageType:      pageType,
			ContentType:   p.determineContentType(aiPage.Path),
			Description:   description,
			Frontmatter:   aiPage.Frontmatter,
			InternalLinks: aiPage.InternalLinks,
			Status:        PageStatusPending,
			Attempts:      0,
		}

		plan.Pages = append(plan.Pages, page)
	}

	plan.Stats.TotalPages = len(plan.Pages)

	return plan, nil
}

// determinePageType determines the page type from page type or path.
func (p *Planner) determinePageType(pageType, path string) PageType {
	switch strings.ToLower(pageType) {
	case "home":
		return PageTypeHome
	case "post", "blog":
		return PageTypePost
	case "docs", "documentation":
		return PageTypeDocs
	case "section":
		return PageTypeSection
	case "page":
		return PageTypePage
	}

	// Fallback: determine from path
	pathLower := strings.ToLower(path)

	if strings.HasSuffix(pathLower, "_index.md") {
		if pathLower == "content/_index.md" {
			return PageTypeHome
		}
		return PageTypeSection
	}

	if strings.Contains(pathLower, "/posts/") || strings.Contains(pathLower, "/blog/") {
		return PageTypePost
	}

	if strings.Contains(pathLower, "/docs/") {
		return PageTypeDocs
	}

	return PageTypePage
}

// determineContentType extracts the Hugo content type from the provided file path.
func (p *Planner) determineContentType(path string) string {
	pathLower := strings.ToLower(path)

	// Remove "content/" prefix
	pathLower = strings.TrimPrefix(pathLower, "content/")

	// Get first directory
	parts := strings.Split(pathLower, "/")
	if len(parts) > 1 {
		return parts[0]
	}

	return ""
}

// Plan Validation

// validatePlan performs comprehensive validation on the parsed site plan.
func (p *Planner) validatePlan(plan *SitePlan) error {
	if plan == nil {
		return NewValidationError("plan", nil, "plan is nil")
	}

	if plan.ID == "" {
		return NewValidationError("id", plan.ID, "plan ID is required")
	}

	if len(plan.Pages) == 0 {
		return ErrEmptyPlan
	}

	// Check minimum pages for site type
	minCount := MinimumPageCount(plan.SiteType)
	if len(plan.Pages) < minCount {
		return NewValidationError("pages", len(plan.Pages),
			fmt.Sprintf("site type %s requires at least %d pages, got %d",
				plan.SiteType, minCount, len(plan.Pages)))
	}

	// Validate each page
	pathsSeen := make(map[string]bool)
	for i, page := range plan.Pages {
		if page.Path == "" {
			return NewValidationError(fmt.Sprintf("pages[%d].path", i), page.Path, "path is required")
		}

		// Normalize path
		normalizedPath := strings.ToLower(page.Path)
		if pathsSeen[normalizedPath] {
			return NewValidationError(fmt.Sprintf("pages[%d].path", i), page.Path, "duplicate path")
		}
		pathsSeen[normalizedPath] = true

		// Check path starts with content/
		if !strings.HasPrefix(strings.ToLower(page.Path), "content/") {
			return NewValidationError(fmt.Sprintf("pages[%d].path", i), page.Path,
				"path must start with 'content/'")
		}

		// Title is optional at this stage - can be in frontmatter
	}

	// Verify home page exists
	hasHome := false
	for _, page := range plan.Pages {
		if page.PageType == PageTypeHome ||
			strings.ToLower(page.Path) == "content/_index.md" {
			hasHome = true
			break
		}
	}

	if !hasHome {
		return NewValidationError("pages", nil, "plan must include a home page (content/_index.md)")
	}

	// Verify section index files exist for multi-page sections
	sectionPageCount := map[string]int{}
	sectionHasIndex := map[string]bool{}

	for _, page := range plan.Pages {
		lower := strings.ToLower(page.Path)
		for _, section := range []string{"posts", "docs", "services", "projects"} {
			prefix := "content/" + section + "/"
			if strings.HasPrefix(lower, prefix) {
				sectionPageCount[section]++
				if strings.HasSuffix(lower, "/_index.md") {
					sectionHasIndex[section] = true
				}
			}
		}
	}

	for section := range sectionPageCount {
		if !sectionHasIndex[section] {
			// Auto-add missing section index instead of failing validation
			sectionTitle := strings.ToUpper(section[:1]) + section[1:]
			plan.Pages = append(plan.Pages, PageSpec{
				ID:          fmt.Sprintf("section_%s_index", section),
				Path:        fmt.Sprintf("content/%s/_index.md", section),
				Title:       sectionTitle,
				PageType:    PageTypeSection,
				ContentType: "section",
				Description: fmt.Sprintf("Index page for the %s section", section),
				Status:      PageStatusPending,
			})
		}
	}

	return nil
}

// Helper Functions

// CleanJSONResponse strips markdown code fence formatting from JSON responses.
func CleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)

	// Remove markdown code fences
	prefixes := []string{
		"```json\n",
		"```json",
		"```JSON\n",
		"```JSON",
		"```\n",
		"```",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(response, prefix) {
			response = strings.TrimPrefix(response, prefix)
			break
		}
	}

	suffixes := []string{
		"\n```",
		"```",
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(response, suffix) {
			response = strings.TrimSuffix(response, suffix)
			break
		}
	}

	return strings.TrimSpace(response)
}
