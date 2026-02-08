package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/selimozten/walgo/internal/config"
	"gopkg.in/yaml.v3"
)

// ContentStructure holds comprehensive information about the site for AI context
type ContentStructure struct {
	// Site path (root directory)
	SitePath string

	// Site configuration from hugo.toml
	SiteConfig *SiteConfigInfo

	// Theme layout information
	ThemeInfo *ThemeLayoutInfo

	// Content directory sections (posts, pages, docs, etc.)
	ContentTypes []ContentTypeInfo

	// All content files with frontmatter
	ContentFiles []ContentFileInfo

	// Default content type
	DefaultType string

	// Content directory path
	ContentDir string
}

// SiteConfigInfo holds site configuration for AI context (from hugo.toml)
type SiteConfigInfo struct {
	// Core fields
	Title       string
	BaseURL     string
	Language    string
	Theme       string
	Description string
	Author      string

	// Dynamic configuration
	Menu       []MenuInfo
	Params     map[string]interface{} // Site-wide params
	Taxonomies map[string]string      // e.g., {"tag": "tags", "category": "categories"}
	Permalinks map[string]string      // URL patterns per section
	Markup     map[string]interface{} // Markup configuration
	Outputs    map[string][]string    // Output formats per kind

	// Raw config for full context
	RawConfig map[string]interface{}
}

// MenuInfo holds menu item info
type MenuInfo struct {
	Name   string
	URL    string
	Weight int
}

// ThemeLayoutInfo holds theme layout information for AI context
type ThemeLayoutInfo struct {
	Name              string
	SupportedSections []string
	FrontmatterFields []string
	Description       string
}

// ContentTypeInfo holds information about a content type
type ContentTypeInfo struct {
	Name      string
	Path      string
	FileCount int
	Files     []string // All files in this section
}

// ContentFileInfo holds information about a content file with frontmatter
type ContentFileInfo struct {
	Path        string
	Title       string
	Description string
	Date        string
	Draft       bool
	Tags        []string
	Extra       map[string]string
	BundleType  string // "branch", "leaf", or "single"
}

// ContentGenerationParams holds parameters for smart content generation
type ContentGenerationParams struct {
	SitePath     string
	Instructions string
	Context      context.Context
}

// ContentGenerationResult holds the result of content generation
type ContentGenerationResult struct {
	Success      bool
	Content      string
	FilePath     string
	ContentType  string
	Filename     string
	Error        error
	ErrorMessage string
}

// ContentGenerator handles intelligent content generation with structure awareness
type ContentGenerator struct {
	client *Client
}

// NewContentGenerator creates a new content generator
func NewContentGenerator(client *Client) *ContentGenerator {
	return &ContentGenerator{
		client: client,
	}
}

// GetContentStructure scans and returns comprehensive site information for AI context
func GetContentStructure(sitePath string) (*ContentStructure, error) {
	cfg, err := config.LoadConfigFrom(sitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	contentDir := filepath.Join(sitePath, cfg.HugoConfig.ContentDir)
	structure := &ContentStructure{
		SitePath:     sitePath,
		ContentTypes: []ContentTypeInfo{},
		DefaultType:  "posts",
		ContentDir:   contentDir,
	}

	// Load site configuration from hugo.toml
	siteConfig := loadSiteConfig(sitePath)
	if siteConfig != nil {
		structure.SiteConfig = siteConfig

		// Get theme layout info if theme is set
		if siteConfig.Theme != "" {
			structure.ThemeInfo = getThemeInfo(sitePath, siteConfig.Theme)
		}
	}

	// Get all content files with frontmatter
	structure.ContentFiles = getAllContentFiles(sitePath)

	// Read content directory for sections
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		// If content directory doesn't exist, return structure with what we have
		if os.IsNotExist(err) {
			return structure, nil
		}
		return nil, fmt.Errorf("failed to read content directory: %w", err)
	}

	// Scan each subdirectory
	for _, entry := range entries {
		if entry.IsDir() {
			typePath := filepath.Join(contentDir, entry.Name())

			// Collect ALL markdown files recursively in this section
			var mdFiles []string
			_ = filepath.Walk(typePath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
					relPath, _ := filepath.Rel(typePath, path)
					mdFiles = append(mdFiles, relPath)
				}
				return nil
			})

			structure.ContentTypes = append(structure.ContentTypes, ContentTypeInfo{
				Name:      entry.Name(),
				Path:      typePath,
				FileCount: len(mdFiles),
				Files:     mdFiles, // All files, not just first 5
			})
		}
	}

	// Set default type based on common conventions or theme info
	if structure.ThemeInfo != nil && len(structure.ThemeInfo.SupportedSections) > 0 {
		// Prefer the first supported section from theme
		for _, section := range structure.ThemeInfo.SupportedSections {
			for _, ct := range structure.ContentTypes {
				if ct.Name == section {
					structure.DefaultType = section
					break
				}
			}
			if structure.DefaultType != "posts" {
				break
			}
		}
	}

	// Fallback to common conventions
	if structure.DefaultType == "posts" {
		for _, ct := range structure.ContentTypes {
			if ct.Name == "posts" || ct.Name == "post" || ct.Name == "blog" {
				structure.DefaultType = ct.Name
				break
			}
		}
	}

	return structure, nil
}

// loadSiteConfig reads hugo.toml and extracts site configuration dynamically
func loadSiteConfig(sitePath string) *SiteConfigInfo {
	// Try hugo.toml first, then config.toml, then yaml variants
	configFiles := []string{
		filepath.Join(sitePath, "hugo.toml"),
		filepath.Join(sitePath, "config.toml"),
		filepath.Join(sitePath, "hugo.yaml"),
		filepath.Join(sitePath, "config.yaml"),
	}

	var configPath string
	var isYAML bool
	for _, cf := range configFiles {
		if _, err := os.Stat(cf); err == nil {
			configPath = cf
			isYAML = strings.HasSuffix(cf, ".yaml") || strings.HasSuffix(cf, ".yml")
			break
		}
	}

	if configPath == "" {
		return nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	// Parse into generic map
	var rawConfig map[string]interface{}
	if isYAML {
		if err := yaml.Unmarshal(content, &rawConfig); err != nil {
			return nil
		}
	} else {
		if err := toml.Unmarshal(content, &rawConfig); err != nil {
			return nil
		}
	}

	cfg := &SiteConfigInfo{
		Language:   "en",
		RawConfig:  rawConfig,
		Params:     make(map[string]interface{}),
		Taxonomies: make(map[string]string),
		Permalinks: make(map[string]string),
		Markup:     make(map[string]interface{}),
		Outputs:    make(map[string][]string),
	}

	// Extract core fields
	if v, ok := rawConfig["title"].(string); ok {
		cfg.Title = v
	}
	if v, ok := rawConfig["baseURL"].(string); ok {
		cfg.BaseURL = v
	}
	if v, ok := rawConfig["languageCode"].(string); ok {
		cfg.Language = v
	} else if v, ok := rawConfig["language"].(string); ok {
		cfg.Language = v
	}
	if v, ok := rawConfig["theme"].(string); ok {
		cfg.Theme = v
	}

	// Extract params (site-wide parameters)
	if params, ok := rawConfig["params"].(map[string]interface{}); ok {
		cfg.Params = params
		// Also extract common fields from params
		if desc, ok := params["description"].(string); ok {
			cfg.Description = desc
		}
		if author, ok := params["author"].(string); ok {
			cfg.Author = author
		}
	}

	// Extract taxonomies
	if taxonomies, ok := rawConfig["taxonomies"].(map[string]interface{}); ok {
		for k, v := range taxonomies {
			if str, ok := v.(string); ok {
				cfg.Taxonomies[k] = str
			}
		}
	}

	// Extract permalinks
	if permalinks, ok := rawConfig["permalinks"].(map[string]interface{}); ok {
		for k, v := range permalinks {
			if str, ok := v.(string); ok {
				cfg.Permalinks[k] = str
			}
		}
	}

	// Extract markup settings
	if markup, ok := rawConfig["markup"].(map[string]interface{}); ok {
		cfg.Markup = markup
	}

	// Extract outputs
	if outputs, ok := rawConfig["outputs"].(map[string]interface{}); ok {
		for k, v := range outputs {
			if arr, ok := v.([]interface{}); ok {
				var formats []string
				for _, item := range arr {
					if str, ok := item.(string); ok {
						formats = append(formats, str)
					}
				}
				cfg.Outputs[k] = formats
			}
		}
	}

	// Extract menus (can be under "menu" or "menus")
	cfg.Menu = extractMenuItems(rawConfig)

	// If theme not found in config, check themes directory
	if cfg.Theme == "" {
		themesDir := filepath.Join(sitePath, "themes")
		if entries, err := os.ReadDir(themesDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					cfg.Theme = entry.Name()
					break
				}
			}
		}
	}

	return cfg
}

// extractMenuItems extracts menu items from config
func extractMenuItems(rawConfig map[string]interface{}) []MenuInfo {
	var menus []MenuInfo

	// Try "menu" first (Hugo's standard)
	menuConfig, ok := rawConfig["menu"].(map[string]interface{})
	if !ok {
		// Try "menus" (alternative)
		menuConfig, ok = rawConfig["menus"].(map[string]interface{})
	}
	if !ok {
		return menus
	}

	// Process each menu (main, footer, etc.)
	for _, menuItems := range menuConfig {
		items, ok := menuItems.([]interface{})
		if !ok {
			// Could be a map for single item
			if item, ok := menuItems.(map[string]interface{}); ok {
				items = []interface{}{item}
			} else {
				continue
			}
		}

		for _, item := range items {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			mi := MenuInfo{}
			if name, ok := itemMap["name"].(string); ok {
				mi.Name = name
			}
			if url, ok := itemMap["url"].(string); ok {
				mi.URL = url
			} else if pageRef, ok := itemMap["pageRef"].(string); ok {
				mi.URL = pageRef
			}
			if weight, ok := itemMap["weight"].(int64); ok {
				mi.Weight = int(weight)
			} else if weight, ok := itemMap["weight"].(int); ok {
				mi.Weight = weight
			}

			if mi.Name != "" {
				menus = append(menus, mi)
			}
		}
	}

	return menus
}

// getThemeInfo returns theme-specific layout information using dynamic analysis.
func getThemeInfo(sitePath, themeName string) *ThemeLayoutInfo {
	analysis := AnalyzeTheme(sitePath, themeName)

	// Convert ThemeAnalysis to ThemeLayoutInfo
	info := &ThemeLayoutInfo{
		Name:              analysis.Name,
		Description:       analysis.Description,
		SupportedSections: analysis.Sections,
	}

	// Collect unique frontmatter fields across all sections
	fieldSet := make(map[string]bool)
	for _, fields := range analysis.FrontmatterFields {
		for _, f := range fields {
			fieldSet[f] = true
		}
	}
	for f := range fieldSet {
		info.FrontmatterFields = append(info.FrontmatterFields, f)
	}

	// Sensible defaults if dynamic analysis found nothing
	if len(info.SupportedSections) == 0 {
		info.SupportedSections = []string{"posts", "pages"}
	}
	if len(info.FrontmatterFields) == 0 {
		info.FrontmatterFields = []string{"title", "date", "draft", "description"}
	}
	if info.Description == "" {
		info.Description = "Custom theme."
	}

	return info
}

// getAllContentFiles scans the content directory and returns all markdown files with frontmatter
func getAllContentFiles(sitePath string) []ContentFileInfo {
	contentDir := filepath.Join(sitePath, "content")

	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return nil
	}

	var files []ContentFileInfo

	_ = filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(contentDir, path)

		fileInfo := ContentFileInfo{
			Path:       relPath,
			Extra:      make(map[string]string),
			BundleType: GetPageBundleType(relPath),
		}

		content, err := os.ReadFile(path)
		if err != nil {
			files = append(files, fileInfo)
			return nil
		}

		// Parse frontmatter (YAML between --- markers)
		contentStr := string(content)
		if strings.HasPrefix(contentStr, "---") {
			parts := strings.SplitN(contentStr[3:], "---", 2)
			if len(parts) >= 1 {
				frontmatter := parts[0]

				var fm map[string]interface{}
				if err := yaml.Unmarshal([]byte(frontmatter), &fm); err == nil {
					if title, ok := fm["title"].(string); ok {
						fileInfo.Title = title
					}
					if desc, ok := fm["description"].(string); ok {
						fileInfo.Description = desc
					}
					if date, ok := fm["date"].(string); ok {
						fileInfo.Date = date
					}
					if draft, ok := fm["draft"].(bool); ok {
						fileInfo.Draft = draft
					}
					if tags, ok := fm["tags"].([]interface{}); ok {
						for _, t := range tags {
							if tag, ok := t.(string); ok {
								fileInfo.Tags = append(fileInfo.Tags, tag)
							}
						}
					}
					for k, v := range fm {
						if k != "title" && k != "description" && k != "date" && k != "draft" && k != "tags" {
							if strVal, ok := v.(string); ok {
								fileInfo.Extra[k] = strVal
							}
						}
					}
				}
			}
		}

		files = append(files, fileInfo)
		return nil
	})

	return files
}

// GenerateContent creates content based on natural language instructions
func (cg *ContentGenerator) GenerateContent(params ContentGenerationParams) *ContentGenerationResult {
	result := &ContentGenerationResult{
		Success: false,
	}

	// Get content structure
	structure, err := GetContentStructure(params.SitePath)
	if err != nil {
		result.Error = err
		result.ErrorMessage = fmt.Sprintf("failed to get content structure: %v", err)
		return result
	}

	// Build system prompt with content structure
	systemPrompt := cg.buildSmartSystemPrompt(structure)

	// Build user prompt
	userPrompt := fmt.Sprintf("User instructions: %s", params.Instructions)

	// Generate content
	ctx := params.Context
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), LongRequestTimeout)
		defer cancel()
	}

	response, err := cg.client.GenerateContentWithContext(ctx, systemPrompt, userPrompt)
	if err != nil {
		result.Error = err
		result.ErrorMessage = fmt.Sprintf("AI generation failed: %v", err)
		return result
	}

	// Parse AI response
	contentType, filename, content := cg.parseAIResponse(response, structure)

	// Validate and sanitize
	if contentType == "" {
		contentType = structure.DefaultType
	}
	if filename == "" {
		filename = fmt.Sprintf("ai-generated-%d.md", time.Now().Unix())
	}
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	// Sanitize filename
	filename = sanitizeFilename(filename)

	// Load config to get content directory
	cfg, err := config.LoadConfigFrom(params.SitePath)
	if err != nil {
		result.Error = err
		result.ErrorMessage = fmt.Sprintf("failed to load config: %v", err)
		return result
	}

	// Build save path using absolute path from sitePath
	// Ensure ContentDir is resolved relative to sitePath, not current working directory
	contentDir := cfg.HugoConfig.ContentDir
	if !filepath.IsAbs(contentDir) {
		contentDir = filepath.Join(params.SitePath, contentDir)
	}
	savePath := filepath.Join(contentDir, contentType, filename)

	// Security check - ensure path is within content directory
	absContentDir, err := filepath.Abs(contentDir)
	if err != nil {
		result.Error = err
		result.ErrorMessage = "invalid content directory"
		return result
	}
	absSavePath, err := filepath.Abs(savePath)
	if err != nil {
		result.Error = err
		result.ErrorMessage = "invalid save path"
		return result
	}
	relPath, err := filepath.Rel(absContentDir, absSavePath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		result.Error = fmt.Errorf("path traversal detected")
		result.ErrorMessage = "invalid path - security check failed"
		return result
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		result.Error = err
		result.ErrorMessage = fmt.Sprintf("failed to create directory: %v", err)
		return result
	}

	// Save file
	if err := os.WriteFile(savePath, []byte(content), 0644); err != nil {
		result.Error = err
		result.ErrorMessage = fmt.Sprintf("failed to save file: %v", err)
		return result
	}

	// Success!
	result.Success = true
	result.Content = content
	result.FilePath = savePath
	result.ContentType = contentType
	result.Filename = filename

	return result
}

// buildSmartSystemPrompt creates a system prompt with comprehensive site context
func (cg *ContentGenerator) buildSmartSystemPrompt(structure *ContentStructure) string {
	var contextParts []string

	// Add site configuration
	if structure.SiteConfig != nil {
		cfg := structure.SiteConfig
		var sb strings.Builder
		fmt.Fprintf(&sb, "SITE CONFIGURATION:\n- Title: %s\n- Theme: %s\n- Language: %s",
			cfg.Title, cfg.Theme, cfg.Language)

		if cfg.Description != "" {
			fmt.Fprintf(&sb, "\n- Description: %s", cfg.Description)
		}
		if cfg.Author != "" {
			fmt.Fprintf(&sb, "\n- Author: %s", cfg.Author)
		}
		if cfg.BaseURL != "" {
			fmt.Fprintf(&sb, "\n- Base URL: %s", cfg.BaseURL)
		}

		// Add menu if present
		if len(cfg.Menu) > 0 {
			sb.WriteString("\n- Menu Items:")
			for _, item := range cfg.Menu {
				if item.Weight > 0 {
					fmt.Fprintf(&sb, "\n  * %s -> %s (weight: %d)", item.Name, item.URL, item.Weight)
				} else {
					fmt.Fprintf(&sb, "\n  * %s -> %s", item.Name, item.URL)
				}
			}
		}

		// Add taxonomies if configured
		if len(cfg.Taxonomies) > 0 {
			sb.WriteString("\n- Taxonomies:")
			for singular, plural := range cfg.Taxonomies {
				fmt.Fprintf(&sb, "\n  * %s -> %s", singular, plural)
			}
		}

		// Add permalinks if configured
		if len(cfg.Permalinks) > 0 {
			sb.WriteString("\n- Permalinks:")
			for section, pattern := range cfg.Permalinks {
				fmt.Fprintf(&sb, "\n  * %s: %s", section, pattern)
			}
		}

		// Add outputs if configured
		if len(cfg.Outputs) > 0 {
			sb.WriteString("\n- Outputs:")
			for kind, formats := range cfg.Outputs {
				fmt.Fprintf(&sb, "\n  * %s: [%s]", kind, strings.Join(formats, ", "))
			}
		}

		// Add site params (useful context for AI)
		if len(cfg.Params) > 0 {
			sb.WriteString("\n- Site Parameters:")
			for key, value := range cfg.Params {
				// Only show string/simple values, skip complex objects
				switch v := value.(type) {
				case string:
					if len(v) < 100 { // Avoid very long strings
						fmt.Fprintf(&sb, "\n  * %s: %q", key, v)
					}
				case bool:
					fmt.Fprintf(&sb, "\n  * %s: %v", key, v)
				case int, int64, float64:
					fmt.Fprintf(&sb, "\n  * %s: %v", key, v)
				}
			}
		}

		contextParts = append(contextParts, sb.String())
	}

	// Add theme-specific information (dynamic analysis preferred, compact fallback)
	themeContextAdded := false
	if structure.SiteConfig != nil && structure.SiteConfig.Theme != "" {
		themeContext := BuildDynamicThemeContext(structure.SitePath, structure.SiteConfig.Theme)
		if themeContext != "" {
			contextParts = append(contextParts, themeContext)
			themeContextAdded = true
		}
	}
	if !themeContextAdded && structure.ThemeInfo != nil {
		themeInfo := fmt.Sprintf(`THEME INFORMATION:
- Theme Name: %s
- Description: %s
- Supported Sections: %s
- Required Frontmatter Fields: %s`,
			structure.ThemeInfo.Name,
			structure.ThemeInfo.Description,
			strings.Join(structure.ThemeInfo.SupportedSections, ", "),
			strings.Join(structure.ThemeInfo.FrontmatterFields, ", "))

		contextParts = append(contextParts, themeInfo)
	}

	// Add content structure (all files)
	if len(structure.ContentTypes) > 0 {
		structureInfo := "EXISTING CONTENT STRUCTURE:\n"
		for _, ct := range structure.ContentTypes {
			structureInfo += fmt.Sprintf("- %s/ (%d files)\n", ct.Name, ct.FileCount)
			// List all files in this section
			for _, file := range ct.Files {
				structureInfo += fmt.Sprintf("  - %s\n", file)
			}
		}
		contextParts = append(contextParts, structureInfo)
	}

	// Add existing content files with their metadata
	if len(structure.ContentFiles) > 0 {
		filesInfo := "EXISTING CONTENT FILES:\n"
		for _, file := range structure.ContentFiles {
			// Show bundle type for clarity
			bundleLabel := ""
			switch file.BundleType {
			case "branch":
				bundleLabel = "[SECTION]"
			case "leaf":
				bundleLabel = "[BUNDLE]"
			default:
				bundleLabel = "[PAGE]"
			}

			fileEntry := fmt.Sprintf("- %s %s", file.Path, bundleLabel)
			if file.Title != "" {
				fileEntry += fmt.Sprintf(" | Title: %q", file.Title)
			}
			if file.Description != "" && len(file.Description) < 100 {
				fileEntry += fmt.Sprintf(" | Desc: %q", file.Description)
			}
			if len(file.Tags) > 0 {
				fileEntry += fmt.Sprintf(" | Tags: [%s]", strings.Join(file.Tags, ", "))
			}
			filesInfo += fileEntry + "\n"
		}
		contextParts = append(contextParts, filesInfo)
	}

	// Combine all context
	// Dynamic theme analysis already provides comprehensive Hugo/theme knowledge
	fullContext := "=== SITE CONTEXT ===\n\n" + strings.Join(contextParts, "\n\n")

	// Build the smart content prompt with this context
	return BuildSmartContentPrompt(fullContext)
}

// parseAIResponse extracts content type, filename, and content from AI response
func (cg *ContentGenerator) parseAIResponse(response string, structure *ContentStructure) (contentType, filename, content string) {
	lines := strings.Split(response, "\n")
	contentStartIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "CONTENT_TYPE:") {
			contentType = strings.TrimSpace(strings.TrimPrefix(trimmed, "CONTENT_TYPE:"))
		} else if strings.HasPrefix(trimmed, "FILENAME:") {
			filename = strings.TrimSpace(strings.TrimPrefix(trimmed, "FILENAME:"))
		} else if trimmed == "---" && contentStartIdx == -1 {
			// Start of frontmatter
			contentStartIdx = i
			break
		}
	}

	// If we found the start of content, extract everything from there
	if contentStartIdx >= 0 {
		content = strings.Join(lines[contentStartIdx:], "\n")
		content = CleanGeneratedContent(content)
	} else {
		// Fallback: try to find frontmatter anywhere in response
		content = CleanGeneratedContent(response)
	}

	return contentType, filename, content
}

// sanitizeFilename removes dangerous characters from filenames
func sanitizeFilename(filename string) string {
	// Get base name only (no path components)
	filename = filepath.Base(filename)

	// Remove path traversal attempts
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// Convert to lowercase
	filename = strings.ToLower(filename)

	// Replace spaces with hyphens
	filename = strings.ReplaceAll(filename, " ", "-")

	// Remove special characters except hyphens, underscores, and dots
	reg := regexp.MustCompile(`[^a-z0-9\-_.]`)
	filename = reg.ReplaceAllString(filename, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	filename = reg.ReplaceAllString(filename, "-")

	// Trim hyphens from start and end
	filename = strings.Trim(filename, "-")

	// Ensure it's not empty
	if filename == "" || filename == "." || filename == ".md" {
		filename = fmt.Sprintf("ai-generated-%d", time.Now().Unix())
	}

	return filename
}
