package hugo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/selimozten/walgo/internal/ai"
)

// DetectSiteType detects site type from Hugo configuration or content structure.
// This is a unified implementation that should be used everywhere.
//
// Detection strategy:
// 1. Check config file (hugo.toml or config.toml) for theme name
// 2. Fallback to content directory structure analysis
//
// Returns:
//
//	ai.SiteType: Detected site type (defaults to Blog)
func DetectSiteType(sitePath string) ai.SiteType {
	return detectFromConfig(sitePath)
}

// detectFromConfig attempts to determine site type from Hugo config file.
func detectFromConfig(sitePath string) ai.SiteType {
	// Try hugo.toml first
	hugoTomlPath := filepath.Join(sitePath, "hugo.toml")
	if content, err := os.ReadFile(hugoTomlPath); err == nil {
		if siteType := parseConfigForTheme(string(content), sitePath); siteType != ai.SiteTypeBlog {
			return siteType
		}
	}

	// Try config.toml as fallback
	configTomlPath := filepath.Join(sitePath, "config.toml")
	if content, err := os.ReadFile(configTomlPath); err == nil {
		if siteType := parseConfigForTheme(string(content), sitePath); siteType != ai.SiteTypeBlog {
			return siteType
		}
	}

	// No config or no theme detected
	return ai.SiteTypeBlog
}

// parseConfigForTheme extracts site type from config file content.
// This uses a combination of theme hints and content structure analysis.
func parseConfigForTheme(configContent string, sitePath string) ai.SiteType {
	configLower := strings.ToLower(configContent)

	// Check for walgo built-in themes
	if strings.Contains(configLower, "walgo-biolink") {
		return ai.SiteTypeBiolink
	}
	if strings.Contains(configLower, "walgo-whitepaper") {
		return ai.SiteTypeWhitepaper
	}

	// Check for documentation-focused themes (theme names containing "doc", "book", "learn")
	docThemePatterns := []string{"hugo-book", "docsy", "learn", "doks", "geekdoc", "techdoc"}
	for _, pattern := range docThemePatterns {
		if strings.Contains(configLower, pattern) {
			return ai.SiteTypeDocs
		}
	}

	// For other themes (general purpose like Ananke, PaperMod, etc.)
	// Determine site type from content structure
	return detectSiteTypeFromContent(sitePath)
}

// detectSiteTypeFromContent analyzes content directory to determine site type.
func detectSiteTypeFromContent(sitePath string) ai.SiteType {
	contentDir := filepath.Join(sitePath, "content")

	// Check for whitepaper directory (whitepaper site indicator)
	if dirExists(filepath.Join(contentDir, "whitepaper")) {
		return ai.SiteTypeWhitepaper
	}

	// Check for docs directory (documentation site indicator)
	if dirExists(filepath.Join(contentDir, "docs")) {
		return ai.SiteTypeDocs
	}

	// Check for biolink indicators (links/ directory or biolink-style config)
	if dirExists(filepath.Join(contentDir, "links")) || dirExists(filepath.Join(contentDir, "biolink")) {
		return ai.SiteTypeBiolink
	}

	// Default to blog
	return ai.SiteTypeBlog
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// GetThemeName extracts the theme name from Hugo config file.
// Returns empty string if no theme is found or config cannot be read.
func GetThemeName(sitePath string) string {
	// Try hugo.toml first
	hugoTomlPath := filepath.Join(sitePath, "hugo.toml")
	if content, err := os.ReadFile(hugoTomlPath); err == nil {
		if theme := extractThemeFromConfig(string(content)); theme != "" {
			return theme
		}
	}

	// Try config.toml as fallback
	configTomlPath := filepath.Join(sitePath, "config.toml")
	if content, err := os.ReadFile(configTomlPath); err == nil {
		if theme := extractThemeFromConfig(string(content)); theme != "" {
			return theme
		}
	}

	return ""
}

// extractThemeFromConfig parses config content to find the "theme" key.
// Only matches the standalone "theme" key (not "themeDir", "themeColor", etc.).
func extractThemeFromConfig(configContent string) string {
	lines := strings.Split(configContent, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Match "theme" followed by optional whitespace and "="
		if (trimmed == "theme" || strings.HasPrefix(trimmed, "theme ") || strings.HasPrefix(trimmed, "theme=")) &&
			strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				// Verify the key is exactly "theme" (not themeDir, themeColor, etc.)
				key := strings.TrimSpace(parts[0])
				if key != "theme" {
					continue
				}
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"'`)
				return value
			}
		}
	}
	return ""
}
