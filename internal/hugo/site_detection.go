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
// 2. Fallback to content structure analysis
//
// Returns:
//
//	ai.SiteType: Detected site type (defaults to Blog)
func DetectSiteType(sitePath string) ai.SiteType {
	// Strategy 1: Check config file for theme
	if siteType := detectFromConfig(sitePath); siteType != ai.SiteTypeBlog {
		return siteType
	}

	// Strategy 2: Fallback to content structure
	return detectFromContentStructure(sitePath)
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
func parseConfigForTheme(configContent string, sitePath string) ai.SiteType {
	configLower := strings.ToLower(configContent)

	// Check for Hugo Book theme (docs)
	if strings.Contains(configLower, "hugo-book") || strings.Contains(configLower, "theme = \"book\"") {
		return ai.SiteTypeDocs
	}

	// Check for Ananke theme (can be blog, business, or portfolio)
	if strings.Contains(configLower, "ananke") || strings.Contains(configLower, "theme = \"ananke\"") {
		// Check content structure to determine which Ananke-based site type
		if siteType := checkAnankeSiteType(sitePath); siteType != ai.SiteTypeBlog {
			return siteType
		}
		return ai.SiteTypeBlog
	}

	// Check for Coder theme (portfolio)
	if strings.Contains(configLower, "coder") || strings.Contains(configLower, "theme = \"coder\"") {
		return ai.SiteTypePortfolio
	}

	return ai.SiteTypeBlog
}

// detectFromContentStructure analyzes content directory structure to determine site type.
func detectFromContentStructure(sitePath string) ai.SiteType {
	// Get content structure (this uses ai package for now, could be moved to hugo)
	structure, err := ai.GetContentStructure(sitePath)
	if err != nil {
		return ai.SiteTypeBlog // Default to blog on error
	}

	// Check for type indicators
	for _, ct := range structure.ContentTypes {
		switch ct.Name {
		case "docs":
			return ai.SiteTypeDocs
		case "projects":
			return ai.SiteTypePortfolio
		case "services", "testimonials":
			return ai.SiteTypeBusiness
		}
	}

	return ai.SiteTypeBlog // Default
}

// checkAnankeSiteType determines specific Ananke-based site type by checking content structure.
func checkAnankeSiteType(sitePath string) ai.SiteType {
	// Check for services directory (business site)
	servicesDir := filepath.Join(sitePath, "content", "services")
	if info, err := os.Stat(servicesDir); err == nil && info.IsDir() {
		return ai.SiteTypeBusiness
	}

	// Check for projects directory (portfolio site)
	projectsDir := filepath.Join(sitePath, "content", "projects")
	if info, err := os.Stat(projectsDir); err == nil && info.IsDir() {
		return ai.SiteTypePortfolio
	}

	// Default to blog for Ananke
	return ai.SiteTypeBlog
}
