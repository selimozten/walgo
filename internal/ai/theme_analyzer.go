package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// THEME ANALYZER
// =============================================================================
//
// This file provides dynamic theme analysis capabilities for Hugo sites.
// Instead of hardcoded theme rules, the system analyzes:
//
// 1. Theme layouts/ directory - to find supported sections
// 2. Theme archetypes/ - to find frontmatter templates
// 3. Theme exampleSite/ - to find recommended configuration
// 4. Theme README.md - to find documented params
// 5. Existing site content - to learn patterns
//
// This allows the AI to work with ANY Hugo theme automatically.
// =============================================================================

// =============================================================================
// TYPES
// =============================================================================

// ThemeAnalysis holds dynamically detected theme information
type ThemeAnalysis struct {
	Name              string
	Description       string
	Sections          []string            // Detected from layouts/
	Archetypes        map[string]string   // section -> frontmatter template
	FrontmatterFields map[string][]string // section -> required fields
	HasTaxonomies     bool
	HasSearch         bool
	MinHugoVersion    string
}

// ThemeConfigAnalysis holds configuration recommendations for a theme
type ThemeConfigAnalysis struct {
	Theme             string
	Description       string
	RecommendedParams map[string]interface{}
	RequiredParams    []string
	OptionalParams    []string
	PageParams        []string // Params used in page templates (.Params.X)
	Taxonomies        map[string]string
	Menus             []MenuConfig
	Outputs           map[string][]string
	ExampleConfig     string
}

// MenuConfig holds menu configuration from theme
type MenuConfig struct {
	Name  string
	Items []MenuItem
}

// MenuItem holds a menu item
type MenuItem struct {
	Name   string
	URL    string
	Weight int
}

// ContentPatterns holds learned patterns from existing content
type ContentPatterns struct {
	SectionPatterns map[string]*SectionPattern
}

// SectionPattern holds learned patterns for a section
type SectionPattern struct {
	Name         string
	CommonFields []string
	FieldCounts  map[string]int
	FileCount    int // Total files analyzed in this section
	UsesBundle   bool
	ExampleFiles []string
}

// =============================================================================
// MAIN ANALYSIS FUNCTIONS
// =============================================================================

// AnalyzeTheme dynamically analyzes a Hugo theme to extract structure info
func AnalyzeTheme(sitePath, themeName string) *ThemeAnalysis {
	analysis := &ThemeAnalysis{
		Name:              themeName,
		Sections:          []string{},
		Archetypes:        make(map[string]string),
		FrontmatterFields: make(map[string][]string),
	}

	if themeName == "" {
		return analysis
	}

	themeDir := filepath.Join(sitePath, "themes", themeName)
	if _, err := os.Stat(themeDir); os.IsNotExist(err) {
		return analysis
	}

	// 1. Read theme.toml/theme.yaml for description
	analysis.Description, analysis.MinHugoVersion = readThemeMetadata(themeDir)

	// 2. Analyze layouts/ directory for supported sections
	analysis.Sections = detectSectionsFromLayouts(themeDir)

	// 3. Analyze archetypes/ for frontmatter templates
	analysis.Archetypes, analysis.FrontmatterFields = analyzeArchetypes(themeDir)

	// 4. Also check site's own archetypes (they override theme)
	siteArchetypes, siteFrontmatter := analyzeArchetypes(sitePath)
	for k, v := range siteArchetypes {
		analysis.Archetypes[k] = v
	}
	for k, v := range siteFrontmatter {
		analysis.FrontmatterFields[k] = v
	}

	// 5. Check if theme has taxonomy and search templates
	analysis.HasTaxonomies = hasTaxonomySupport(themeDir)
	analysis.HasSearch = hasSearchSupport(themeDir)

	return analysis
}

// AnalyzeThemeConfig analyzes a theme to extract recommended hugo.toml configuration
func AnalyzeThemeConfig(sitePath, themeName string) *ThemeConfigAnalysis {
	analysis := &ThemeConfigAnalysis{
		Theme:             themeName,
		RecommendedParams: make(map[string]interface{}),
		Taxonomies:        make(map[string]string),
		Outputs:           make(map[string][]string),
	}

	if themeName == "" {
		return analysis
	}

	themeDir := filepath.Join(sitePath, "themes", themeName)
	if _, err := os.Stat(themeDir); os.IsNotExist(err) {
		return analysis
	}

	// 1. Read theme description
	analysis.Description, _ = readThemeMetadata(themeDir)

	// 2. Look for exampleSite config
	analysis.ExampleConfig = findExampleConfig(themeDir)

	// 3. Parse example config for params
	if analysis.ExampleConfig != "" {
		parseExampleConfig(analysis)
	}

	// 4. Analyze theme layouts for required params
	analyzeLayoutsForParams(themeDir, analysis)

	// 5. Analyze theme.toml for features
	analyzeThemeToml(themeDir, analysis)

	// 6. Check README for configuration hints
	analyzeReadmeForConfig(themeDir, analysis)

	return analysis
}

// AnalyzeSiteContent analyzes existing site content to learn structure
func AnalyzeSiteContent(sitePath string) *ContentPatterns {
	patterns := &ContentPatterns{
		SectionPatterns: make(map[string]*SectionPattern),
	}

	contentDir := filepath.Join(sitePath, "content")
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return patterns
	}

	// Walk content directory
	_ = filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(contentDir, path)
		parts := strings.Split(relPath, string(os.PathSeparator))

		// Determine section
		section := ""
		if len(parts) > 1 {
			section = parts[0]
		}

		// Read and analyze frontmatter
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		if strings.HasPrefix(contentStr, "---") {
			fmParts := strings.SplitN(contentStr[3:], "---", 2)
			if len(fmParts) >= 1 {
				fields := extractFrontmatterFields(fmParts[0])

				if section != "" {
					if patterns.SectionPatterns[section] == nil {
						patterns.SectionPatterns[section] = &SectionPattern{
							Name:         section,
							FieldCounts:  make(map[string]int),
							UsesBundle:   false,
							ExampleFiles: []string{},
						}
					}

					sp := patterns.SectionPatterns[section]
					sp.FileCount++
					for _, f := range fields {
						sp.FieldCounts[f]++
					}

					// Check if uses bundles
					if info.Name() == "index.md" {
						sp.UsesBundle = true
					}

					// Add example file
					if len(sp.ExampleFiles) < 3 {
						sp.ExampleFiles = append(sp.ExampleFiles, relPath)
					}
				}
			}
		}

		return nil
	})

	// Calculate common fields per section
	for _, sp := range patterns.SectionPatterns {
		sp.CommonFields = []string{}
		threshold := sp.FileCount / 2
		if threshold < 1 {
			threshold = 1
		}

		for field, count := range sp.FieldCounts {
			if count >= threshold {
				sp.CommonFields = append(sp.CommonFields, field)
			}
		}
	}

	return patterns
}

// =============================================================================
// CONTEXT BUILDERS (for AI prompts)
// =============================================================================

// BuildDynamicThemeContext creates a comprehensive context string for AI.
// Convenience wrapper that runs all analyzers internally.
// If you already have analysis results, use BuildThemeContextFromAnalysis instead.
func BuildDynamicThemeContext(sitePath, themeName string) string {
	if themeName == "" {
		return ""
	}
	themeAnalysis := AnalyzeTheme(sitePath, themeName)
	configAnalysis := AnalyzeThemeConfig(sitePath, themeName)
	contentPatterns := AnalyzeSiteContent(sitePath)
	return BuildThemeContextFromAnalysis(themeName, themeAnalysis, configAnalysis, contentPatterns)
}

// BuildThemeContextFromAnalysis builds the theme context string from pre-computed analysis results.
// Use this when you already have analysis results to avoid redundant filesystem scans.
func BuildThemeContextFromAnalysis(themeName string, themeAnalysis *ThemeAnalysis, configAnalysis *ThemeConfigAnalysis, contentPatterns *ContentPatterns) string {
	var sb strings.Builder

	sb.WriteString("=== DYNAMIC THEME ANALYSIS ===\n\n")

	// Theme info
	sb.WriteString(fmt.Sprintf("THEME: %s\n", themeName))
	if themeAnalysis.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", themeAnalysis.Description))
	}
	sb.WriteString("\n")

	// Sections from layouts
	if len(themeAnalysis.Sections) > 0 {
		sb.WriteString("SUPPORTED SECTIONS (from theme layouts/):\n")
		for _, section := range themeAnalysis.Sections {
			sb.WriteString(fmt.Sprintf("- %s/\n", section))
		}
		sb.WriteString("\n")
	}

	// Archetypes and frontmatter
	if len(themeAnalysis.FrontmatterFields) > 0 {
		sb.WriteString("FRONTMATTER BY SECTION (from archetypes/):\n")
		for section, fields := range themeAnalysis.FrontmatterFields {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", section, strings.Join(fields, ", ")))
		}
		sb.WriteString("\n")
	}

	// Required and optional params
	if len(configAnalysis.RequiredParams) > 0 || len(configAnalysis.OptionalParams) > 0 {
		sb.WriteString("SITE PARAMS (from theme templates):\n")
		if len(configAnalysis.RequiredParams) > 0 {
			sb.WriteString(fmt.Sprintf("- Required: %s\n", strings.Join(configAnalysis.RequiredParams, ", ")))
		}
		if len(configAnalysis.OptionalParams) > 0 {
			sb.WriteString(fmt.Sprintf("- Optional: %s\n", strings.Join(configAnalysis.OptionalParams, ", ")))
		}
		sb.WriteString("\n")
	}

	// Page params
	if len(configAnalysis.PageParams) > 0 {
		sb.WriteString(fmt.Sprintf("PAGE PARAMS (from .Params.X in templates): %s\n\n",
			strings.Join(configAnalysis.PageParams, ", ")))
	}

	// Taxonomies
	if themeAnalysis.HasTaxonomies {
		sb.WriteString("TAXONOMIES: Supported (tags, categories)\n")
		if len(configAnalysis.Taxonomies) > 0 {
			for k, v := range configAnalysis.Taxonomies {
				sb.WriteString(fmt.Sprintf("  - %s -> %s\n", k, v))
			}
		}
		sb.WriteString("\n")
	}

	// Search
	if themeAnalysis.HasSearch {
		sb.WriteString("SEARCH: Supported\n\n")
	}

	// Example menus from exampleSite
	if len(configAnalysis.Menus) > 0 {
		sb.WriteString("MENU STRUCTURE (from exampleSite):\n")
		for _, menu := range configAnalysis.Menus {
			sb.WriteString(fmt.Sprintf("- menu.%s:\n", menu.Name))
			for _, item := range menu.Items {
				sb.WriteString(fmt.Sprintf("  * %s -> %s\n", item.Name, item.URL))
			}
		}
		sb.WriteString("\n")
	}

	// Recommended params
	if len(configAnalysis.RecommendedParams) > 0 {
		sb.WriteString("RECOMMENDED PARAMS (from exampleSite):\n")
		for key, value := range configAnalysis.RecommendedParams {
			switch v := value.(type) {
			case string:
				if len(v) < 60 {
					sb.WriteString(fmt.Sprintf("- %s: %q\n", key, v))
				}
			case bool, int, int64, float64:
				sb.WriteString(fmt.Sprintf("- %s: %v\n", key, v))
			}
		}
		sb.WriteString("\n")
	}

	// Learned patterns from existing content
	if len(contentPatterns.SectionPatterns) > 0 {
		sb.WriteString("LEARNED FROM EXISTING CONTENT:\n")
		for _, sp := range contentPatterns.SectionPatterns {
			bundleInfo := ""
			if sp.UsesBundle {
				bundleInfo = " [page bundles]"
			}
			sb.WriteString(fmt.Sprintf("- %s/%s\n", sp.Name, bundleInfo))
			if len(sp.CommonFields) > 0 {
				sb.WriteString(fmt.Sprintf("  Fields: %s\n", strings.Join(sp.CommonFields, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GetRecommendedFrontmatter returns recommended frontmatter for a section
func GetRecommendedFrontmatter(analysis *ThemeAnalysis, patterns *ContentPatterns, section string) []string {
	fields := []string{"title", "draft"} // Always required

	if analysis == nil {
		return fields
	}

	// First check theme archetypes
	if themeFields, ok := analysis.FrontmatterFields[section]; ok {
		for _, f := range themeFields {
			if f != "title" && f != "draft" && !containsString(fields, f) {
				fields = append(fields, f)
			}
		}
		return fields
	}

	// Check default archetype
	if defaultFields, ok := analysis.FrontmatterFields["default"]; ok {
		for _, f := range defaultFields {
			if f != "title" && f != "draft" && !containsString(fields, f) {
				fields = append(fields, f)
			}
		}
	}

	// Then check learned patterns
	if patterns != nil {
		if sp, ok := patterns.SectionPatterns[section]; ok {
			for _, f := range sp.CommonFields {
				if !containsString(fields, f) {
					fields = append(fields, f)
				}
			}
		}
	}

	return fields
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func readThemeMetadata(themeDir string) (description, minVersion string) {
	// Try theme.toml
	tomlPath := filepath.Join(themeDir, "theme.toml")
	if content, err := os.ReadFile(tomlPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "description") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					description = strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				}
			}
			if strings.HasPrefix(trimmed, "min_version") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					minVersion = strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				}
			}
		}
		return
	}

	// Try theme.yaml
	yamlPath := filepath.Join(themeDir, "theme.yaml")
	if content, err := os.ReadFile(yamlPath); err == nil {
		var data map[string]interface{}
		if yaml.Unmarshal(content, &data) == nil {
			if desc, ok := data["description"].(string); ok {
				description = desc
			}
			if ver, ok := data["min_version"].(string); ok {
				minVersion = ver
			}
		}
	}

	return
}

func detectSectionsFromLayouts(themeDir string) []string {
	layoutsDir := filepath.Join(themeDir, "layouts")
	sections := []string{}

	ignoreDirs := map[string]bool{
		"_default": true, "partials": true, "shortcodes": true,
		"_markup": true, "_headers": true,
	}

	entries, err := os.ReadDir(layoutsDir)
	if err != nil {
		return sections
	}

	for _, entry := range entries {
		if entry.IsDir() && !ignoreDirs[entry.Name()] {
			sectionDir := filepath.Join(layoutsDir, entry.Name())
			if hasLayoutFiles(sectionDir) {
				sections = append(sections, entry.Name())
			}
		}
	}

	return sections
}

func hasLayoutFiles(dir string) bool {
	layoutFiles := []string{"list.html", "single.html", "section.html"}
	for _, f := range layoutFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			return true
		}
	}
	return false
}

func analyzeArchetypes(baseDir string) (map[string]string, map[string][]string) {
	archetypes := make(map[string]string)
	frontmatterFields := make(map[string][]string)

	archetypesDir := filepath.Join(baseDir, "archetypes")
	if _, err := os.Stat(archetypesDir); os.IsNotExist(err) {
		return archetypes, frontmatterFields
	}

	_ = filepath.Walk(archetypesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(archetypesDir, path)
		sectionName := strings.TrimSuffix(filepath.Base(relPath), ".md")

		if info.Name() == "index.md" {
			sectionName = filepath.Dir(relPath)
		}

		contentStr := string(content)
		if strings.HasPrefix(contentStr, "---") {
			parts := strings.SplitN(contentStr[3:], "---", 2)
			if len(parts) >= 1 {
				frontmatter := parts[0]
				archetypes[sectionName] = frontmatter
				fields := extractFrontmatterFields(frontmatter)
				if len(fields) > 0 {
					frontmatterFields[sectionName] = fields
				}
			}
		}

		return nil
	})

	return archetypes, frontmatterFields
}

func extractFrontmatterFields(frontmatter string) []string {
	fields := []string{}
	lines := strings.Split(frontmatter, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if idx := strings.Index(trimmed, ":"); idx > 0 {
			field := strings.TrimSpace(trimmed[:idx])
			if !strings.Contains(field, "{") && !strings.Contains(field, "}") {
				fields = append(fields, field)
			}
		}
	}

	return fields
}

func hasTaxonomySupport(themeDir string) bool {
	paths := []string{
		filepath.Join(themeDir, "layouts", "_default", "taxonomy.html"),
		filepath.Join(themeDir, "layouts", "_default", "term.html"),
		filepath.Join(themeDir, "layouts", "taxonomy"),
		filepath.Join(themeDir, "layouts", "tags"),
		filepath.Join(themeDir, "layouts", "categories"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func hasSearchSupport(themeDir string) bool {
	paths := []string{
		filepath.Join(themeDir, "layouts", "search"),
		filepath.Join(themeDir, "layouts", "_default", "search.html"),
		filepath.Join(themeDir, "static", "js", "search.js"),
		filepath.Join(themeDir, "assets", "js", "search.js"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func findExampleConfig(themeDir string) string {
	configPaths := []string{
		filepath.Join(themeDir, "exampleSite", "hugo.toml"),
		filepath.Join(themeDir, "exampleSite", "config.toml"),
		filepath.Join(themeDir, "exampleSite", "hugo.yaml"),
		filepath.Join(themeDir, "exampleSite", "config.yaml"),
		filepath.Join(themeDir, "exampleSite", "config", "_default", "config.toml"),
		filepath.Join(themeDir, "exampleSite", "config", "_default", "params.toml"),
	}

	for _, path := range configPaths {
		if content, err := os.ReadFile(path); err == nil {
			return string(content)
		}
	}
	return ""
}

func parseExampleConfig(analysis *ThemeConfigAnalysis) {
	config := analysis.ExampleConfig

	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(config), &data); err == nil {
		extractConfigData(data, analysis)
		return
	}

	parseTomlManually(config, analysis)
}

func extractConfigData(data map[string]interface{}, analysis *ThemeConfigAnalysis) {
	if params, ok := data["params"].(map[string]interface{}); ok {
		for k, v := range params {
			analysis.RecommendedParams[k] = v
		}
	}

	if taxonomies, ok := data["taxonomies"].(map[string]interface{}); ok {
		for k, v := range taxonomies {
			if str, ok := v.(string); ok {
				analysis.Taxonomies[k] = str
			}
		}
	}

	if menu, ok := data["menu"].(map[string]interface{}); ok {
		for menuName, items := range menu {
			menuConfig := MenuConfig{Name: menuName}
			if itemsList, ok := items.([]interface{}); ok {
				for _, item := range itemsList {
					if itemMap, ok := item.(map[string]interface{}); ok {
						mi := MenuItem{}
						if name, ok := itemMap["name"].(string); ok {
							mi.Name = name
						}
						if url, ok := itemMap["url"].(string); ok {
							mi.URL = url
						}
						if weight, ok := itemMap["weight"].(int); ok {
							mi.Weight = weight
						}
						menuConfig.Items = append(menuConfig.Items, mi)
					}
				}
			}
			analysis.Menus = append(analysis.Menus, menuConfig)
		}
	}

	if outputs, ok := data["outputs"].(map[string]interface{}); ok {
		for kind, formats := range outputs {
			if arr, ok := formats.([]interface{}); ok {
				var fmts []string
				for _, f := range arr {
					if str, ok := f.(string); ok {
						fmts = append(fmts, str)
					}
				}
				analysis.Outputs[kind] = fmts
			}
		}
	}
}

func parseTomlManually(config string, analysis *ThemeConfigAnalysis) {
	lines := strings.Split(config, "\n")
	inParams := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "[params]") || strings.HasPrefix(trimmed, "[Params]") {
			inParams = true
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			inParams = false
			continue
		}

		if strings.Contains(trimmed, "=") && inParams {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"'`)
				analysis.RecommendedParams[key] = value
			}
		}
	}
}

var (
	siteParamPattern = regexp.MustCompile(`\.Site\.Params\.(\w+)`)
	pageParamPattern = regexp.MustCompile(`\.Params\.(\w+)`)
)

func analyzeLayoutsForParams(themeDir string, analysis *ThemeConfigAnalysis) {
	layoutsDir := filepath.Join(themeDir, "layouts")

	_ = filepath.Walk(layoutsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".html") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)

		// Site params
		matches := siteParamPattern.FindAllStringSubmatch(contentStr, -1)
		for _, match := range matches {
			if len(match) > 1 {
				paramName := strings.ToLower(match[1])
				if !containsString(analysis.OptionalParams, paramName) &&
					!containsString(analysis.RequiredParams, paramName) {
					if strings.Contains(contentStr, "if .Site.Params."+match[1]) ||
						strings.Contains(contentStr, "with .Site.Params."+match[1]) {
						analysis.OptionalParams = append(analysis.OptionalParams, paramName)
					} else {
						analysis.RequiredParams = append(analysis.RequiredParams, paramName)
					}
				}
			}
		}

		// Page params
		matches = pageParamPattern.FindAllStringSubmatch(contentStr, -1)
		for _, match := range matches {
			if len(match) > 1 {
				paramName := strings.ToLower(match[1])
				if !containsString(analysis.PageParams, paramName) {
					analysis.PageParams = append(analysis.PageParams, paramName)
				}
			}
		}

		return nil
	})
}

func analyzeThemeToml(themeDir string, analysis *ThemeConfigAnalysis) {
	tomlPath := filepath.Join(themeDir, "theme.toml")
	content, err := os.ReadFile(tomlPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	inFeatures := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[features]") {
			inFeatures = true
			continue
		}
		if strings.HasPrefix(trimmed, "[") && inFeatures {
			inFeatures = false
			continue
		}

		if inFeatures && strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				feature := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if value == "true" {
					analysis.RecommendedParams["feature_"+feature] = true
				}
			}
		}
	}
}

func analyzeReadmeForConfig(themeDir string, analysis *ThemeConfigAnalysis) {
	readmePaths := []string{
		filepath.Join(themeDir, "README.md"),
		filepath.Join(themeDir, "readme.md"),
	}

	var readmeContent string
	for _, path := range readmePaths {
		if content, err := os.ReadFile(path); err == nil {
			readmeContent = string(content)
			break
		}
	}

	if readmeContent == "" {
		return
	}

	paramMentions := regexp.MustCompile(`(?i)params\.(\w+)`)
	matches := paramMentions.FindAllStringSubmatch(readmeContent, -1)
	for _, match := range matches {
		if len(match) > 1 {
			paramName := strings.ToLower(match[1])
			if !containsString(analysis.OptionalParams, paramName) &&
				!containsString(analysis.RequiredParams, paramName) {
				analysis.OptionalParams = append(analysis.OptionalParams, paramName)
			}
		}
	}
}

func containsString(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func writeParamValue(sb *strings.Builder, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		sb.WriteString(fmt.Sprintf("  %s = %q\n", key, v))
	case bool:
		sb.WriteString(fmt.Sprintf("  %s = %v\n", key, v))
	case int, int64, float64:
		sb.WriteString(fmt.Sprintf("  %s = %v\n", key, v))
	case []interface{}:
		sb.WriteString(fmt.Sprintf("  %s = [", key))
		for i, item := range v {
			if i > 0 {
				sb.WriteString(", ")
			}
			if str, ok := item.(string); ok {
				sb.WriteString(fmt.Sprintf("%q", str))
			} else {
				sb.WriteString(fmt.Sprintf("%v", item))
			}
		}
		sb.WriteString("]\n")
	}
}

// =============================================================================
// PAGE BUNDLE HELPERS
// =============================================================================

// GetPageBundleType determines if a path should be a leaf or branch bundle
func GetPageBundleType(path string) string {
	if path == "_index.md" || path == "content/_index.md" || strings.HasSuffix(path, "/_index.md") {
		return "branch"
	}
	if path == "index.md" || strings.HasSuffix(path, "/index.md") {
		return "leaf"
	}
	return "single"
}

// =============================================================================
// THEME CONFIG UPDATE
// =============================================================================

// UpdateSiteConfigForTheme merges theme-specific [params] from the theme's
// exampleSite config into the existing hugo.toml. All existing settings
// (baseURL, title, [menu], [outputs], [markup], etc.) are preserved.
// Only new params from the theme are added; existing params are NOT overwritten.
func UpdateSiteConfigForTheme(sitePath, themeName string) error {
	configPath := filepath.Join(sitePath, "hugo.toml")

	existingContent, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	// Extract theme params from exampleSite
	themeDir := filepath.Join(sitePath, "themes", themeName)
	themeParams := extractThemeParams(themeDir)
	if len(themeParams) == 0 {
		return nil // Nothing to merge
	}

	// Parse existing config to check which params already exist
	var existingConfig map[string]interface{}
	if err := toml.Unmarshal(existingContent, &existingConfig); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	// Get existing params (if any)
	existingParams, _ := existingConfig["params"].(map[string]interface{})

	// Find new params that don't already exist in the config
	newParams := make(map[string]interface{})
	for k, v := range themeParams {
		if existingParams != nil {
			if _, exists := existingParams[k]; exists {
				continue // Don't overwrite existing param
			}
		}
		newParams[k] = v
	}

	if len(newParams) == 0 {
		return nil // All params already present
	}

	// Append new params to the existing config file
	content := string(existingContent)

	// Build TOML snippet for new params
	var sb strings.Builder
	sb.WriteString("\n# Theme-specific params (from theme exampleSite)\n")

	// Check if [params] section already exists in file
	hasParamsSection := strings.Contains(content, "[params]") || strings.Contains(content, "[Params]")
	if !hasParamsSection {
		sb.WriteString("[params]\n")
	}

	writeThemeParams(&sb, newParams, "  ")

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += sb.String()

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// extractThemeParams extracts params from theme's exampleSite config
func extractThemeParams(themeDir string) map[string]interface{} {
	params := make(map[string]interface{})

	// Find example config
	configPaths := []string{
		filepath.Join(themeDir, "exampleSite", "hugo.yaml"),
		filepath.Join(themeDir, "exampleSite", "config.yaml"),
		filepath.Join(themeDir, "exampleSite", "hugo.toml"),
		filepath.Join(themeDir, "exampleSite", "config.toml"),
		filepath.Join(themeDir, "exampleSite", "config", "_default", "params.yaml"),
		filepath.Join(themeDir, "exampleSite", "config", "_default", "params.toml"),
	}

	for _, path := range configPaths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var data map[string]interface{}

		// Try YAML first
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			if err := yaml.Unmarshal(content, &data); err == nil {
				if p, ok := data["params"].(map[string]interface{}); ok {
					mergeParams(params, p)
				}
			}
		} else {
			// Try TOML
			if err := toml.Unmarshal(content, &data); err == nil {
				if p, ok := data["params"].(map[string]interface{}); ok {
					mergeParams(params, p)
				}
			}
		}
	}

	return params
}

// skipParamKeys lists exampleSite param keys that reference the theme author's
// repository or demo-specific settings and should not be merged into user sites.
var skipParamKeys = map[string]bool{
	"BookRepo":           true,
	"BookEditLink":       true,
	"BookLastChangeLink": true,
}

// mergeParams merges source params into destination, skipping exampleSite-specific values.
// Params whose string values contain Hugo template delimiters ({{ }}) are skipped because
// they reference the theme author's example paths/repos and are meaningless for user sites.
// Additionally, known repo/edit-link keys are skipped since they always reference the
// theme author's repository.
func mergeParams(dest, src map[string]interface{}) {
	for k, v := range src {
		if skipParamKeys[k] {
			continue
		}
		if str, ok := v.(string); ok {
			if strings.Contains(str, "{{") {
				continue // Skip Hugo template expressions
			}
		}
		dest[k] = v
	}
}

// writeThemeParams writes theme params to string builder with proper TOML formatting
func writeThemeParams(sb *strings.Builder, params map[string]interface{}, indent string) {
	// First pass: write simple values and simple arrays
	for key, value := range params {
		switch v := value.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("%s%s = %q\n", indent, key, v))
		case bool:
			sb.WriteString(fmt.Sprintf("%s%s = %v\n", indent, key, v))
		case int, int64, float64:
			sb.WriteString(fmt.Sprintf("%s%s = %v\n", indent, key, v))
		case []interface{}:
			if len(v) == 0 {
				sb.WriteString(fmt.Sprintf("%s%s = []\n", indent, key))
			} else if _, isMap := v[0].(map[string]interface{}); isMap {
				// Array of tables — handled in third pass
				continue
			} else {
				// Simple inline array (strings, numbers, bools)
				sb.WriteString(fmt.Sprintf("%s%s = [", indent, key))
				for i, item := range v {
					if i > 0 {
						sb.WriteString(", ")
					}
					if s, ok := item.(string); ok {
						sb.WriteString(fmt.Sprintf("%q", s))
					} else {
						sb.WriteString(fmt.Sprintf("%v", item))
					}
				}
				sb.WriteString("]\n")
			}
		}
	}

	// Second pass: write nested tables
	for key, value := range params {
		if nested, ok := value.(map[string]interface{}); ok {
			sb.WriteString(fmt.Sprintf("\n%s[params.%s]\n", indent, key))
			writeNestedParams(sb, nested, indent+"  ")
		}
	}

	// Third pass: write arrays of tables ([[params.key]])
	for key, value := range params {
		if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
			if _, isMap := arr[0].(map[string]interface{}); isMap {
				for _, item := range arr {
					if m, ok := item.(map[string]interface{}); ok {
						sb.WriteString(fmt.Sprintf("\n[[params.%s]]\n", key))
						writeNestedParams(sb, m, indent)
					}
				}
			}
		}
	}
}

// writeNestedParams writes nested params
func writeNestedParams(sb *strings.Builder, params map[string]interface{}, indent string) {
	for key, value := range params {
		switch v := value.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("%s%s = %q\n", indent, key, v))
		case bool:
			sb.WriteString(fmt.Sprintf("%s%s = %v\n", indent, key, v))
		case int, int64, float64:
			sb.WriteString(fmt.Sprintf("%s%s = %v\n", indent, key, v))
		case []interface{}:
			if len(v) == 0 {
				sb.WriteString(fmt.Sprintf("%s%s = []\n", indent, key))
			} else if _, isMap := v[0].(map[string]interface{}); isMap {
				// Skip arrays of tables in nested context — not representable inline
			} else {
				sb.WriteString(fmt.Sprintf("%s%s = [", indent, key))
				for i, item := range v {
					if i > 0 {
						sb.WriteString(", ")
					}
					if s, ok := item.(string); ok {
						sb.WriteString(fmt.Sprintf("%q", s))
					} else {
						sb.WriteString(fmt.Sprintf("%v", item))
					}
				}
				sb.WriteString("]\n")
			}
		case map[string]interface{}:
			// Handle inline tables for simple nested objects
			sb.WriteString(fmt.Sprintf("%s%s = { ", indent, key))
			first := true
			for nk, nv := range v {
				if !first {
					sb.WriteString(", ")
				}
				first = false
				switch tv := nv.(type) {
				case string:
					sb.WriteString(fmt.Sprintf("%s = %q", nk, tv))
				case bool:
					sb.WriteString(fmt.Sprintf("%s = %v", nk, tv))
				default:
					sb.WriteString(fmt.Sprintf("%s = %v", nk, tv))
				}
			}
			sb.WriteString(" }\n")
		}
	}
}

// =============================================================================
// DYNAMIC ARCHETYPE GENERATION
// =============================================================================

// ArchetypeConfig holds configuration for generating an archetype
type ArchetypeConfig struct {
	Section       string            // Section name (e.g., "posts", "docs", "projects")
	Fields        []string          // Frontmatter fields to include
	IncludeDate   bool              // Whether to include date field
	IncludeWeight bool              // Whether to include weight field
	IncludeTags   bool              // Whether to include tags field
	CustomFields  map[string]string // Custom fields with default values
}

// SetupDynamicArchetypes creates archetypes based on theme analysis
// This replaces static embedded archetypes with theme-aware dynamic ones
func SetupDynamicArchetypes(sitePath, themeName string) error {
	archetypesDir := filepath.Join(sitePath, "archetypes")

	// Create archetypes directory if it doesn't exist
	if err := os.MkdirAll(archetypesDir, 0755); err != nil {
		return fmt.Errorf("creating archetypes directory: %w", err)
	}

	// Analyze the theme to understand what archetypes it needs
	themeAnalysis := AnalyzeTheme(sitePath, themeName)
	configAnalysis := AnalyzeThemeConfig(sitePath, themeName)

	// Generate archetypes based on theme analysis
	archetypeConfigs := buildArchetypeConfigs(themeAnalysis, configAnalysis)

	for _, config := range archetypeConfigs {
		content := generateArchetypeContent(config)
		filename := config.Section + ".md"
		destPath := filepath.Join(archetypesDir, filename)

		if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing archetype %s: %w", filename, err)
		}
	}

	// Always create a default archetype
	defaultContent := generateDefaultArchetype(themeAnalysis, configAnalysis)
	if err := os.WriteFile(filepath.Join(archetypesDir, "default.md"), []byte(defaultContent), 0644); err != nil {
		return fmt.Errorf("writing default archetype: %w", err)
	}

	return nil
}

// buildArchetypeConfigs creates archetype configurations based on theme analysis
func buildArchetypeConfigs(theme *ThemeAnalysis, config *ThemeConfigAnalysis) []ArchetypeConfig {
	configs := []ArchetypeConfig{}

	// Check theme archetypes first - use them as base
	for section, fields := range theme.FrontmatterFields {
		if section == "default" {
			continue // Skip default, we handle it separately
		}

		cfg := ArchetypeConfig{
			Section:       section,
			Fields:        fields,
			IncludeDate:   containsField(fields, "date"),
			IncludeWeight: containsField(fields, "weight"),
			IncludeTags:   containsField(fields, "tags"),
			CustomFields:  make(map[string]string),
		}
		configs = append(configs, cfg)
	}

	// If no archetypes found in theme, infer from sections
	if len(configs) == 0 && len(theme.Sections) > 0 {
		for _, section := range theme.Sections {
			cfg := inferArchetypeConfig(section, config)
			configs = append(configs, cfg)
		}
	}

	return configs
}

// inferArchetypeConfig infers archetype configuration from section name and theme params
func inferArchetypeConfig(section string, config *ThemeConfigAnalysis) ArchetypeConfig {
	cfg := ArchetypeConfig{
		Section:      section,
		Fields:       []string{"title", "description", "draft"},
		CustomFields: make(map[string]string),
	}

	sectionLower := strings.ToLower(section)

	// Infer based on common section naming conventions
	switch {
	case strings.Contains(sectionLower, "post") || strings.Contains(sectionLower, "blog") || strings.Contains(sectionLower, "article"):
		cfg.IncludeDate = true
		cfg.IncludeTags = true
		cfg.Fields = append(cfg.Fields, "date", "tags")
		// Check if theme uses featured_image
		if containsString(config.PageParams, "featured_image") || containsString(config.PageParams, "image") {
			cfg.Fields = append(cfg.Fields, "featured_image")
		}

	case strings.Contains(sectionLower, "doc") || strings.Contains(sectionLower, "guide") || strings.Contains(sectionLower, "tutorial"):
		cfg.IncludeWeight = true
		cfg.Fields = append(cfg.Fields, "weight")
		// Check for bookToc type params
		for _, param := range config.PageParams {
			if strings.HasPrefix(strings.ToLower(param), "book") || strings.Contains(strings.ToLower(param), "toc") {
				cfg.Fields = append(cfg.Fields, param)
			}
		}

	case strings.Contains(sectionLower, "project") || strings.Contains(sectionLower, "portfolio") || strings.Contains(sectionLower, "work"):
		cfg.IncludeDate = true
		cfg.Fields = append(cfg.Fields, "date")
		if containsString(config.PageParams, "featured_image") || containsString(config.PageParams, "image") {
			cfg.Fields = append(cfg.Fields, "featured_image")
		}
		if containsString(config.PageParams, "tech_stack") || containsString(config.PageParams, "technologies") {
			cfg.Fields = append(cfg.Fields, "tech_stack")
		}
		if containsString(config.PageParams, "project_url") || containsString(config.PageParams, "url") {
			cfg.Fields = append(cfg.Fields, "project_url")
		}

	case strings.Contains(sectionLower, "service"):
		cfg.IncludeDate = true
		cfg.Fields = append(cfg.Fields, "date")
		if containsString(config.PageParams, "featured_image") || containsString(config.PageParams, "image") {
			cfg.Fields = append(cfg.Fields, "featured_image")
		}
		if containsString(config.PageParams, "price") || containsString(config.PageParams, "pricing") {
			cfg.Fields = append(cfg.Fields, "price")
		}

	case strings.Contains(sectionLower, "page"):
		// Generic pages - minimal frontmatter
		if containsString(config.PageParams, "featured_image") || containsString(config.PageParams, "image") {
			cfg.Fields = append(cfg.Fields, "featured_image")
		}
	}

	return cfg
}

// generateArchetypeContent generates the content for an archetype file
func generateArchetypeContent(config ArchetypeConfig) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(`title: "{{ replace .Name "-" " " | title }}"` + "\n")

	// Add date if needed
	if config.IncludeDate {
		sb.WriteString("date: {{ .Date }}\n")
	}

	sb.WriteString("draft: false\n")
	sb.WriteString(`description: ""` + "\n")

	// Add weight if needed (for docs)
	if config.IncludeWeight {
		sb.WriteString("weight: 10\n")
	}

	// Add tags if needed
	if config.IncludeTags {
		sb.WriteString("tags: []\n")
	}

	// Add other fields from theme analysis
	addedFields := map[string]bool{
		"title": true, "date": true, "draft": true,
		"description": true, "weight": true, "tags": true,
	}

	for _, field := range config.Fields {
		if addedFields[field] {
			continue
		}
		addedFields[field] = true

		// Determine default value based on field name
		defaultValue := getFieldDefaultValue(field)
		sb.WriteString(fmt.Sprintf("%s: %s\n", field, defaultValue))
	}

	// Add custom fields
	for field, value := range config.CustomFields {
		if !addedFields[field] {
			sb.WriteString(fmt.Sprintf("%s: %s\n", field, value))
		}
	}

	sb.WriteString("---\n")

	return sb.String()
}

// generateDefaultArchetype generates a default archetype based on theme analysis
func generateDefaultArchetype(theme *ThemeAnalysis, config *ThemeConfigAnalysis) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(`title: "{{ replace .Name "-" " " | title }}"` + "\n")
	sb.WriteString("date: {{ .Date }}\n")
	sb.WriteString("draft: false\n")
	sb.WriteString(`description: ""` + "\n")

	// Check if theme commonly uses featured_image
	if containsString(config.PageParams, "featured_image") || containsString(config.PageParams, "image") {
		sb.WriteString(`featured_image: ""` + "\n")
	}

	sb.WriteString("---\n")

	return sb.String()
}

// getFieldDefaultValue returns an appropriate default value for a frontmatter field
func getFieldDefaultValue(field string) string {
	fieldLower := strings.ToLower(field)

	switch {
	case strings.Contains(fieldLower, "image") || strings.Contains(fieldLower, "img"):
		return `""`
	case strings.Contains(fieldLower, "url") || strings.Contains(fieldLower, "link"):
		return `""`
	case strings.Contains(fieldLower, "tags") || strings.Contains(fieldLower, "categories"):
		return "[]"
	case strings.Contains(fieldLower, "weight") || strings.Contains(fieldLower, "order"):
		return "10"
	case strings.Contains(fieldLower, "toc"):
		return "true"
	case strings.Contains(fieldLower, "collapse"):
		return "false"
	case strings.Contains(fieldLower, "price"):
		return `""`
	case strings.Contains(fieldLower, "tech") || strings.Contains(fieldLower, "stack"):
		return "[]"
	case strings.Contains(fieldLower, "author"):
		return `""`
	default:
		return `""`
	}
}

// containsField checks if a field exists in the fields slice
func containsField(fields []string, field string) bool {
	for _, f := range fields {
		if strings.EqualFold(f, field) {
			return true
		}
	}
	return false
}

// GetDynamicFrontmatterFields returns the required frontmatter fields for a section
// based on theme analysis (replaces static theme-specific functions)
func GetDynamicFrontmatterFields(sitePath, themeName, section string) []string {
	themeAnalysis := AnalyzeTheme(sitePath, themeName)
	configAnalysis := AnalyzeThemeConfig(sitePath, themeName)

	// Check theme's archetype fields first
	if fields, ok := themeAnalysis.FrontmatterFields[section]; ok {
		return fields
	}

	// Check default archetype
	if fields, ok := themeAnalysis.FrontmatterFields["default"]; ok {
		return fields
	}

	// Infer from section name
	config := inferArchetypeConfig(section, configAnalysis)
	return config.Fields
}

// EnsureDynamicFrontmatter ensures frontmatter has all required fields based on theme
// This replaces static ensureAnankeFrontmatter, ensureBookFrontmatter, etc.
func EnsureDynamicFrontmatter(content, sitePath, themeName, section string) (string, bool) {
	requiredFields := GetDynamicFrontmatterFields(sitePath, themeName, section)

	if len(requiredFields) == 0 {
		return content, false
	}

	// Parse existing frontmatter
	if !strings.HasPrefix(content, "---") {
		return content, false
	}

	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) < 2 {
		return content, false
	}

	frontmatter := parts[0]
	body := parts[1]
	changed := false

	// Check and add missing fields
	for _, field := range requiredFields {
		fieldLower := strings.ToLower(field)
		pattern := fieldLower + ":"

		if !strings.Contains(strings.ToLower(frontmatter), pattern) {
			defaultValue := getFieldDefaultValue(field)
			frontmatter = strings.TrimSuffix(frontmatter, "\n") + "\n" + field + ": " + defaultValue + "\n"
			changed = true
		}
	}

	// Ensure draft: false
	if strings.Contains(frontmatter, "draft: true") {
		frontmatter = strings.Replace(frontmatter, "draft: true", "draft: false", 1)
		changed = true
	}

	if changed {
		return "---" + frontmatter + "---" + body, true
	}

	return content, false
}
