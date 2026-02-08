package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// yamlDatePattern matches ISO 8601 dates commonly used in Hugo frontmatter.
// Examples: 2023-10-27, 2023-10-27T09:00:00Z, 2023-10-27T09:00:00+02:00
var yamlDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}(:\d{2})?([Zz]|[+-]\d{2}:\d{2})?)?$`)

// isYAMLDate returns true if the string looks like a YAML-safe date value.
func isYAMLDate(s string) bool {
	return yamlDatePattern.MatchString(s)
}

// ContentFixer validates and fixes Hugo content for theme-specific requirements.
type ContentFixer struct {
	sitePath  string
	siteType  SiteType
	themeName string // Theme name for dynamic frontmatter analysis
}

// NewContentFixer initializes and returns a new ContentFixer instance.
// Deprecated: Use NewContentFixerWithTheme for dynamic theme support.
func NewContentFixer(sitePath string, siteType SiteType) *ContentFixer {
	return &ContentFixer{
		sitePath:  sitePath,
		siteType:  siteType,
		themeName: "", // Will fall back to generic fixes
	}
}

// NewContentFixerWithTheme initializes a ContentFixer with dynamic theme support.
// The themeName is used for dynamic frontmatter analysis based on the actual theme.
func NewContentFixerWithTheme(sitePath string, siteType SiteType, themeName string) *ContentFixer {
	return &ContentFixer{
		sitePath:  sitePath,
		siteType:  siteType,
		themeName: themeName,
	}
}

// FixAll validates and fixes all content files in the site.
func (cf *ContentFixer) FixAll() error {
	contentDir := filepath.Join(cf.sitePath, "content")

	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return nil // No content directory
	}

	return filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		return cf.fixFile(path)
	})
}

// fixFile validates and fixes a single content file.
func (cf *ContentFixer) fixFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", path, err)
	}

	fixed, changed := cf.fixContent(path, string(content))
	if !changed {
		return nil
	}

	// Validate fixed content still has valid frontmatter structure
	if err := validateFrontmatterStructure(fixed); err != nil {
		return fmt.Errorf("post-fix validation failed for %s: %w (skipping write)", path, err)
	}

	if err := os.WriteFile(path, []byte(fixed), 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", path, err)
	}

	return nil
}

// validateFrontmatterStructure checks that content has properly delimited frontmatter.
func validateFrontmatterStructure(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil // Empty content is valid
	}

	// YAML frontmatter: must start and end with ---
	if strings.HasPrefix(trimmed, "---") {
		rest := trimmed[3:]
		endIdx := strings.Index(rest, "\n---")
		if endIdx == -1 {
			return fmt.Errorf("YAML frontmatter has opening '---' but no closing '---'")
		}
		return nil
	}

	// TOML frontmatter: must start and end with +++
	if strings.HasPrefix(trimmed, "+++") {
		rest := trimmed[3:]
		endIdx := strings.Index(rest, "\n+++")
		if endIdx == -1 {
			return fmt.Errorf("TOML frontmatter has opening '+++' but no closing '+++'")
		}
		return nil
	}

	// JSON frontmatter: must start with {
	if strings.HasPrefix(trimmed, "{") {
		return nil // JSON validation is more complex; basic check is sufficient
	}

	// No frontmatter is also valid (content without metadata)
	return nil
}

// fixContent fixes content based on site type and file path.
func (cf *ContentFixer) fixContent(path, content string) (string, bool) {
	// If we have a theme name, use dynamic frontmatter fixing
	if cf.themeName != "" {
		return cf.fixContentDynamic(path, content)
	}

	// Fall back to legacy site-type based fixing
	switch cf.siteType {
	case SiteTypeBlog:
		return cf.fixBlogContent(path, content)
	case SiteTypeDocs:
		return cf.fixDocsContent(path, content)
	case SiteTypeWhitepaper:
		return cf.fixWhitepaperContent(path, content)
	default:
		return content, false
	}
}

// fixContentDynamic fixes content using dynamic theme analysis.
// This replaces the static theme-specific fix functions.
func (cf *ContentFixer) fixContentDynamic(path, content string) (string, bool) {
	relPath := strings.TrimPrefix(path, cf.sitePath)
	relPath = strings.TrimPrefix(relPath, "/content/")
	relPath = strings.TrimPrefix(relPath, "content/")

	changed := false
	result := content

	// Generic fixes that apply to all themes

	// Fix YAML quotes (apostrophes in single-quoted strings)
	result, c := fixYAMLQuotes(result)
	if c {
		changed = true
	}

	// Fix invalid frontmatter start
	result, c = fixFrontmatterStart(result)
	if c {
		changed = true
	}

	// Remove duplicate H1 (most themes generate H1 from title)
	result, c = removeDuplicateH1(result)
	if c {
		changed = true
	}

	// Determine the section from the path for dynamic frontmatter
	section := determineSectionFromPath(relPath)

	// Use dynamic frontmatter fixing based on theme analysis
	result, c = EnsureDynamicFrontmatter(result, cf.sitePath, cf.themeName, section)
	if c {
		changed = true
	}

	return result, changed
}

// determineSectionFromPath extracts the content section from a relative path.
func determineSectionFromPath(relPath string) string {
	// Handle root level files
	if !strings.Contains(relPath, "/") {
		if relPath == "_index.md" {
			return "home"
		}
		// Root level pages like about.md, contact.md
		return "pages"
	}

	// Extract section from path (first directory)
	parts := strings.Split(relPath, "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return "default"
}

// fixYAMLQuotes fixes YAML frontmatter values that need proper quoting.
// YAML special characters like : (colon) and ' (apostrophe) must be properly quoted.
// This function converts single-quoted strings to double-quoted strings and ensures
// all values with special characters are properly escaped.
// It also handles malformed quotes (unclosed quotes) and YAML arrays.
func fixYAMLQuotes(content string) (string, bool) {
	// Find frontmatter
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return content, false
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return content, false
	}

	frontmatter := parts[1]
	body := parts[2]
	changed := false

	lines := strings.Split(frontmatter, "\n")
	var newLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and lines without colons
		if trimmed == "" || !strings.Contains(trimmed, ":") {
			newLines = append(newLines, line)
			continue
		}

		// Check if this is a key-value pair (not an array or comment)
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}

		// Find the first colon (key separator)
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx == -1 {
			newLines = append(newLines, line)
			continue
		}

		key := trimmed[:colonIdx]
		rest := strings.TrimSpace(trimmed[colonIdx+1:])

		// Skip if no value
		if rest == "" {
			newLines = append(newLines, line)
			continue
		}

		// Skip boolean and numeric values
		if rest == "true" || rest == "false" || rest == "null" {
			newLines = append(newLines, line)
			continue
		}
		// Check if it's a number
		if _, err := strconv.ParseFloat(rest, 64); err == nil {
			newLines = append(newLines, line)
			continue
		}
		// Skip ISO 8601 dates (e.g., 2023-10-27T09:00:00Z)
		if isYAMLDate(rest) {
			newLines = append(newLines, line)
			continue
		}

		// Handle YAML arrays: ['item1', 'item2']
		if strings.HasPrefix(rest, "[") {
			fixedArray, arrayChanged := fixYAMLArray(rest)
			if arrayChanged {
				indent := ""
				for _, ch := range line {
					if ch == ' ' || ch == '\t' {
						indent += string(ch)
					} else {
						break
					}
				}
				newLines = append(newLines, fmt.Sprintf("%s%s: %s", indent, key, fixedArray))
				changed = true
				continue
			}
			newLines = append(newLines, line)
			continue
		}

		// Check if value is already properly quoted with double quotes
		if strings.HasPrefix(rest, "\"") && strings.HasSuffix(rest, "\"") && len(rest) > 1 {
			// Already double-quoted, but verify it's properly escaped
			newLines = append(newLines, line)
			continue
		}

		// Determine if we need to fix quoting
		needsQuoting := false
		var value string

		// Handle malformed single quotes (unclosed or mismatched)
		if strings.HasPrefix(rest, "'") {
			// Check if properly closed
			if strings.HasSuffix(rest, "'") && len(rest) > 1 {
				// Properly closed single quote
				value = rest[1 : len(rest)-1]
			} else {
				// Malformed: unclosed single quote - take everything after the opening quote
				value = rest[1:]
			}
			// ALWAYS convert single quotes to double quotes for consistency
			needsQuoting = true
		} else {
			value = rest
			// Check for special characters that require quoting
			// YAML special chars: : # ' " [ ] { } | > & * ! % @ `
			if strings.Contains(value, ":") || strings.Contains(value, "#") ||
				strings.Contains(value, "'") || strings.Contains(value, "\"") ||
				strings.Contains(value, "[") || strings.Contains(value, "]") ||
				strings.Contains(value, "{") || strings.Contains(value, "}") ||
				strings.Contains(value, "|") || strings.Contains(value, ">") ||
				strings.Contains(value, "&") || strings.Contains(value, "*") ||
				strings.Contains(value, "!") || strings.Contains(value, "%") ||
				strings.Contains(value, "@") || strings.Contains(value, "`") {
				needsQuoting = true
			}
		}

		if needsQuoting {
			// Escape any double quotes and backslashes in the value
			escapedValue := strings.ReplaceAll(value, "\\", "\\\\")
			escapedValue = strings.ReplaceAll(escapedValue, "\"", "\\\"")

			// Get original indentation
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			newLine := fmt.Sprintf("%s%s: \"%s\"", indent, key, escapedValue)
			newLines = append(newLines, newLine)
			changed = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if changed {
		newFrontmatter := strings.Join(newLines, "\n")
		return "---" + newFrontmatter + "---" + body, true
	}

	return content, false
}

// fixYAMLArray converts single-quoted array items to double-quoted.
// Input: ['item1', 'item2', 'item with: colon']
// Output: ["item1", "item2", "item with: colon"]
func fixYAMLArray(arrayStr string) (string, bool) {
	// Simple array format: ['a', 'b', 'c']
	if !strings.HasPrefix(arrayStr, "[") || !strings.HasSuffix(arrayStr, "]") {
		return arrayStr, false
	}

	// Check if it contains single quotes
	if !strings.Contains(arrayStr, "'") {
		return arrayStr, false
	}

	changed := false
	result := "["
	inside := arrayStr[1 : len(arrayStr)-1] // Remove [ and ]

	var items []string
	var currentItem strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for i, ch := range inside {
		if ch == '\'' || ch == '"' {
			if !inQuote {
				inQuote = true
				quoteChar = ch
				if ch == '\'' {
					changed = true
				}
			} else if ch == quoteChar {
				inQuote = false
				quoteChar = 0
			} else {
				currentItem.WriteRune(ch)
			}
		} else if ch == ',' && !inQuote {
			item := strings.TrimSpace(currentItem.String())
			if item != "" {
				items = append(items, item)
			}
			currentItem.Reset()
		} else if ch != ' ' || inQuote || (i > 0 && inside[i-1] != ',') {
			currentItem.WriteRune(ch)
		}
	}

	// Add last item
	item := strings.TrimSpace(currentItem.String())
	if item != "" {
		items = append(items, item)
	}

	// Rebuild array with double quotes
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		// Escape double quotes in item
		escapedItem := strings.ReplaceAll(item, "\\", "\\\\")
		escapedItem = strings.ReplaceAll(escapedItem, "\"", "\\\"")
		result += "\"" + escapedItem + "\""
	}
	result += "]"

	return result, changed
}

// fixFrontmatterStart ensures content starts with proper YAML frontmatter.
func fixFrontmatterStart(content string) (string, bool) {
	content = strings.TrimSpace(content)

	// If already starts with ---, it's fine
	if strings.HasPrefix(content, "---") {
		return content, false
	}

	// Check for common issues like starting with "markdown" or other words
	lines := strings.SplitN(content, "\n", 2)
	if len(lines) == 0 {
		return content, false
	}

	firstLine := strings.TrimSpace(lines[0])

	// If first line is not ---, try to find frontmatter
	if firstLine != "---" {
		// Look for frontmatter pattern
		fmRegex := regexp.MustCompile(`(?s)^[^\n]*?\n---\n(.+?)\n---`)
		if fmRegex.MatchString(content) {
			// Remove the garbage before the first ---
			idx := strings.Index(content, "---")
			if idx > 0 {
				content = content[idx:]
				return content, true
			}
		}
	}

	return content, false
}

// removeDuplicateH1 removes duplicate H1 headings from content.
func removeDuplicateH1(content string) (string, bool) {
	// Find frontmatter end
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return content, false
	}

	frontmatter := parts[1]
	body := parts[2]

	// Check for duplicate H1s at the start of body
	lines := strings.Split(body, "\n")
	var newLines []string
	h1Count := 0
	firstH1Removed := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "# ") && !strings.HasPrefix(trimmed, "## ") {
			h1Count++
			// Skip the first H1 (Ananke generates H1 from title)
			if h1Count == 1 && i < 5 { // Only check first few lines
				firstH1Removed = true
				continue
			}
		}
		newLines = append(newLines, line)
	}

	if firstH1Removed {
		newBody := strings.Join(newLines, "\n")
		// Clean up extra newlines at the start
		newBody = strings.TrimLeft(newBody, "\n")
		return "---" + frontmatter + "---\n\n" + newBody, true
	}

	return content, false
}

// extractFrontmatterField extracts a field value from frontmatter.
func extractFrontmatterField(content, field string) string {
	// Match quoted values (double or single) or unquoted values
	pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^%s:\s*(?:"([^"\n]*)"|'([^'\n]*)'|([^\n]*))`, field))
	matches := pattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		// Return the first non-empty capture group
		for _, m := range matches[1:] {
			if m != "" {
				return strings.TrimSpace(m)
			}
		}
	}
	return ""
}

// addFrontmatterField adds a field to the frontmatter.
func addFrontmatterField(content, field, value string) string {
	// Find the end of frontmatter
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return content
	}

	frontmatter := parts[1]
	body := parts[2]

	// Add the new field
	var newField string
	if value == "true" || value == "false" {
		newField = fmt.Sprintf("%s: %s\n", field, value)
	} else if _, err := strconv.Atoi(value); err == nil {
		newField = fmt.Sprintf("%s: %s\n", field, value)
	} else {
		newField = fmt.Sprintf("%s: '%s'\n", field, value)
	}

	frontmatter = strings.TrimRight(frontmatter, "\n") + "\n" + newField

	return "---" + frontmatter + "---" + body
}

// =============================================================================
// BLOG (Ananke Theme) Content Fixer and Validator
// =============================================================================

// fixBlogContent fixes Ananke theme specific issues.
func (cf *ContentFixer) fixBlogContent(path, content string) (string, bool) {
	relPath := strings.TrimPrefix(path, cf.sitePath)
	relPath = strings.TrimPrefix(relPath, "/content/")
	relPath = strings.TrimPrefix(relPath, "content/")

	changed := false
	result := content

	// Fix YAML quotes (apostrophes in single-quoted strings)
	result, c := fixYAMLQuotes(result)
	if c {
		changed = true
	}

	// Fix invalid frontmatter start
	result, c = fixFrontmatterStart(result)
	if c {
		changed = true
	}

	// Remove duplicate H1 (Ananke generates H1 from title)
	result, c = removeDuplicateH1(result)
	if c {
		changed = true
	}

	// Add required frontmatter based on file type
	switch {
	case relPath == "_index.md":
		result, c = ensureAnankeFrontmatter(result, "home")
		if c {
			changed = true
		}
	case relPath == "about.md":
		result, c = ensureAnankeFrontmatter(result, "page")
		if c {
			changed = true
		}
	case relPath == "contact.md":
		result, c = ensureAnankeFrontmatter(result, "page")
		if c {
			changed = true
		}
	case strings.HasPrefix(relPath, "posts/"):
		result, c = ensureAnankePostFrontmatter(result)
		if c {
			changed = true
		}
	}

	return result, changed
}

// ensureAnankeFrontmatter ensures Ananke theme frontmatter fields exist.
func ensureAnankeFrontmatter(content, pageType string) (string, bool) {
	changed := false

	// Ensure description exists
	if !strings.Contains(content, "description:") {
		title := extractFrontmatterField(content, "title")
		if title != "" {
			content = addFrontmatterField(content, "description", title)
			changed = true
		}
	}

	// Ensure featured_image exists (can be empty)
	if !strings.Contains(content, "featured_image:") {
		content = addFrontmatterField(content, "featured_image", "")
		changed = true
	}

	return content, changed
}

// ensureAnankePostFrontmatter ensures blog post frontmatter fields exist.
func ensureAnankePostFrontmatter(content string) (string, bool) {
	changed := false

	// First ensure base Ananke fields
	content, c := ensureAnankeFrontmatter(content, "post")
	if c {
		changed = true
	}

	// Ensure date exists
	if !strings.Contains(content, "date:") {
		content = addFrontmatterField(content, "date", "2024-01-01T00:00:00Z")
		changed = true
	}

	// Ensure draft: false
	if strings.Contains(content, "draft: true") || strings.Contains(content, "draft:true") {
		content = strings.Replace(content, "draft: true", "draft: false", 1)
		content = strings.Replace(content, "draft:true", "draft: false", 1)
		changed = true
	}
	if !strings.Contains(content, "draft:") {
		content = addFrontmatterField(content, "draft", "false")
		changed = true
	}

	return content, changed
}

// ValidateBlogContent validates content for Ananke theme requirements.
// Returns a list of issues found.
func ValidateBlogContent(sitePath string) []string {
	var issues []string
	contentDir := filepath.Join(sitePath, "content")

	// Required files for blog
	requiredFiles := []string{
		"_index.md",
		"about.md",
		"contact.md",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(contentDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Missing required file: content/%s", file))
		}
	}

	// Check for at least one blog post
	postsDir := filepath.Join(contentDir, "posts")
	if _, err := os.Stat(postsDir); err == nil {
		entries, _ := os.ReadDir(postsDir)
		postCount := 0
		for _, entry := range entries {
			if filepath.Ext(entry.Name()) == ".md" || entry.IsDir() {
				postCount++
			}
		}
		if postCount == 0 {
			issues = append(issues, "No blog posts found in content/posts/")
		}
	} else {
		issues = append(issues, "Missing posts directory: content/posts/")
	}

	// Validate frontmatter in key files
	keyFiles := map[string][]string{
		"_index.md": {"title", "description"},
		"about.md":  {"title", "description"},
	}

	for file, fields := range keyFiles {
		path := filepath.Join(contentDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		for _, field := range fields {
			if !strings.Contains(string(content), field+":") {
				issues = append(issues, fmt.Sprintf("Missing '%s' in %s", field, file))
			}
		}
	}

	// Validate blog posts have required fields
	if _, err := os.Stat(postsDir); err == nil {
		_ = filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(path) != ".md" {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			contentStr := string(content)
			relPath := strings.TrimPrefix(path, contentDir+"/")

			if !strings.Contains(contentStr, "date:") {
				issues = append(issues, fmt.Sprintf("Missing 'date' in %s", relPath))
			}
			if !strings.Contains(contentStr, "title:") {
				issues = append(issues, fmt.Sprintf("Missing 'title' in %s", relPath))
			}
			if strings.Contains(contentStr, "draft: true") {
				issues = append(issues, fmt.Sprintf("Post is draft: %s", relPath))
			}

			return nil
		})
	}

	return issues
}

// fixDocsContent fixes hugo-book theme specific issues.
func (cf *ContentFixer) fixDocsContent(path, content string) (string, bool) {
	relPath := strings.TrimPrefix(path, cf.sitePath)
	relPath = strings.TrimPrefix(relPath, "/content/")
	relPath = strings.TrimPrefix(relPath, "content/")

	result := content
	changed := false

	// Fix YAML quotes first
	result, c := fixYAMLQuotes(result)
	if c {
		changed = true
	}

	// Apply different fixes based on file type
	switch {
	case relPath == "_index.md":
		result, c = ensureDocsFrontmatter(result, "home")
		if c {
			changed = true
		}
	case relPath == "docs/_index.md":
		result, c = ensureDocsFrontmatter(result, "section")
		if c {
			changed = true
		}
	case strings.HasPrefix(relPath, "docs/") && strings.HasSuffix(relPath, "/_index.md"):
		result, c = ensureDocsFrontmatter(result, "section")
		if c {
			changed = true
		}
	case strings.HasPrefix(relPath, "docs/"):
		result, c = ensureDocsFrontmatter(result, "doc")
		if c {
			changed = true
		}
	}

	return result, changed
}

// ensureDocsFrontmatter ensures hugo-book theme frontmatter fields exist.
func ensureDocsFrontmatter(content, pageType string) (string, bool) {
	changed := false

	// Ensure title exists
	if !strings.Contains(content, "title:") {
		content = addFrontmatterField(content, "title", "Untitled")
		changed = true
	}

	// Ensure draft: false
	if strings.Contains(content, "draft: true") || strings.Contains(content, "draft:true") {
		content = strings.Replace(content, "draft: true", "draft: false", 1)
		content = strings.Replace(content, "draft:true", "draft: false", 1)
		changed = true
	}
	if !strings.Contains(content, "draft:") {
		content = addFrontmatterField(content, "draft", "false")
		changed = true
	}

	// Ensure weight exists for proper ordering
	if !strings.Contains(content, "weight:") {
		content = addFrontmatterField(content, "weight", "10")
		changed = true
	}

	return content, changed
}

// ValidateDocsContent validates content for hugo-book theme requirements.
// Returns a list of issues found.
func ValidateDocsContent(sitePath string) []string {
	var issues []string
	contentDir := filepath.Join(sitePath, "content")

	// Required files
	requiredFiles := []string{
		"_index.md",
		"docs/_index.md",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(contentDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Missing required file: content/%s", file))
		}
	}

	// Check for at least one doc page in docs/
	docsDir := filepath.Join(contentDir, "docs")
	if _, err := os.Stat(docsDir); err == nil {
		entries, _ := os.ReadDir(docsDir)
		docCount := 0
		for _, entry := range entries {
			if entry.IsDir() {
				// Check for _index.md in subdirectory
				subIndex := filepath.Join(docsDir, entry.Name(), "_index.md")
				if _, err := os.Stat(subIndex); err == nil {
					docCount++
				}
			} else if entry.Name() != "_index.md" && filepath.Ext(entry.Name()) == ".md" {
				docCount++
			}
		}
		if docCount == 0 {
			issues = append(issues, "No documentation pages found in content/docs/")
		}
	}

	// Validate frontmatter in key files
	docsKeyFiles := map[string][]string{
		"_index.md":      {"title", "description"},
		"docs/_index.md": {"title"},
	}

	for file, fields := range docsKeyFiles {
		path := filepath.Join(contentDir, file)
		fileContent, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		for _, field := range fields {
			if !strings.Contains(string(fileContent), field+":") {
				issues = append(issues, fmt.Sprintf("Missing '%s' in %s", field, file))
			}
		}
	}

	return issues
}

// =============================================================================
// WHITEPAPER (walgo-whitepaper Theme) Content Fixer and Validator
// =============================================================================

// fixWhitepaperContent fixes walgo-whitepaper theme specific issues.
func (cf *ContentFixer) fixWhitepaperContent(path, content string) (string, bool) {
	relPath := strings.TrimPrefix(path, cf.sitePath)
	relPath = strings.TrimPrefix(relPath, "/content/")
	relPath = strings.TrimPrefix(relPath, "content/")

	changed := false
	result := content

	// Fix YAML quotes first
	result, c := fixYAMLQuotes(result)
	if c {
		changed = true
	}

	// Fix invalid frontmatter start
	result, c = fixFrontmatterStart(result)
	if c {
		changed = true
	}

	// Remove duplicate H1 (theme generates H1 from title)
	result, c = removeDuplicateH1(result)
	if c {
		changed = true
	}

	// Apply whitepaper-specific frontmatter fixes based on path
	switch {
	case relPath == "_index.md":
		// Root homepage — just needs title
		result, c = ensureWhitepaperFrontmatter(result, "home")
		if c {
			changed = true
		}
	case relPath == "whitepaper/_index.md":
		// Section index
		result, c = ensureWhitepaperFrontmatter(result, "section")
		if c {
			changed = true
		}
	case strings.HasPrefix(relPath, "whitepaper/"):
		// Individual whitepaper sections — need weight + draft: false
		result, c = ensureWhitepaperFrontmatter(result, "section-page")
		if c {
			changed = true
		}
	case strings.HasPrefix(relPath, "appendix/"):
		// Appendix pages — same requirements as section pages
		result, c = ensureWhitepaperFrontmatter(result, "section-page")
		if c {
			changed = true
		}
	}

	return result, changed
}

// ensureWhitepaperFrontmatter ensures walgo-whitepaper theme frontmatter fields exist.
func ensureWhitepaperFrontmatter(content, pageType string) (string, bool) {
	changed := false

	// Ensure title exists
	if !strings.Contains(content, "title:") {
		content = addFrontmatterField(content, "title", "Untitled")
		changed = true
	}

	// Ensure draft: false
	if strings.Contains(content, "draft: true") || strings.Contains(content, "draft:true") {
		content = strings.Replace(content, "draft: true", "draft: false", 1)
		content = strings.Replace(content, "draft:true", "draft: false", 1)
		changed = true
	}
	if !strings.Contains(content, "draft:") {
		content = addFrontmatterField(content, "draft", "false")
		changed = true
	}

	// Section pages need weight for ordering
	if pageType == "section-page" {
		if !strings.Contains(content, "weight:") {
			content = addFrontmatterField(content, "weight", "10")
			changed = true
		}
	}

	return content, changed
}

// ValidateWhitepaperContent validates content for walgo-whitepaper theme requirements.
// Returns a list of issues found.
func ValidateWhitepaperContent(sitePath string) []string {
	var issues []string
	contentDir := filepath.Join(sitePath, "content")

	// Required files
	requiredFiles := []string{
		"_index.md",
		"whitepaper/_index.md",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(contentDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Missing required file: content/%s", file))
		}
	}

	// Check for at least 2 whitepaper sections
	wpDir := filepath.Join(contentDir, "whitepaper")
	if _, err := os.Stat(wpDir); err == nil {
		entries, _ := os.ReadDir(wpDir)
		sectionCount := 0
		for _, entry := range entries {
			if entry.Name() != "_index.md" && filepath.Ext(entry.Name()) == ".md" {
				sectionCount++
			}
		}
		if sectionCount < 2 {
			issues = append(issues, fmt.Sprintf("Whitepaper needs at least 2 sections, found %d", sectionCount))
		}

		// Validate each section has weight and draft: false
		_ = filepath.Walk(wpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(path) != ".md" {
				return nil
			}
			if info.Name() == "_index.md" {
				return nil // Skip section index
			}

			fileContent, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			contentStr := string(fileContent)
			relPath := strings.TrimPrefix(path, contentDir+"/")

			if !strings.Contains(contentStr, "title:") {
				issues = append(issues, fmt.Sprintf("Missing 'title' in %s", relPath))
			}
			if !strings.Contains(contentStr, "weight:") {
				issues = append(issues, fmt.Sprintf("Missing 'weight' in %s", relPath))
			}
			if strings.Contains(contentStr, "draft: true") {
				issues = append(issues, fmt.Sprintf("Section is draft: %s", relPath))
			}

			return nil
		})
	} else {
		issues = append(issues, "Missing whitepaper directory: content/whitepaper/")
	}

	return issues
}
