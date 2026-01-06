package hugo

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/selimozten/walgo/internal/compress"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/optimizer"
	"github.com/selimozten/walgo/internal/projects"
	"gopkg.in/yaml.v3"
)

//go:embed tomls/*.toml
var tomlsFS embed.FS

//go:embed archetypes/*.md
var archetypesFS embed.FS

//go:embed assets/favicon.svg
var faviconSVG []byte

// ContentType represents a detected content type in a Hugo site
type ContentType struct {
	Name      string // Name of the content type (e.g., "posts", "pages")
	Path      string // Path to the content type directory
	FileCount int    // Number of files in this content type
}

// DetectContentTypes scans the content directory and returns found content types
func DetectContentTypes(sitePath string) ([]ContentType, error) {
	contentDir := filepath.Join(sitePath, "content")

	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, fmt.Errorf("reading content directory: %w", err)
	}

	var types []ContentType
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories
		if entry.Name()[0] == '.' {
			continue
		}

		// Count files in this directory
		typePath := filepath.Join(contentDir, entry.Name())
		fileCount := countMarkdownFiles(typePath)

		types = append(types, ContentType{
			Name:      entry.Name(),
			Path:      typePath,
			FileCount: fileCount,
		})
	}

	// Sort by file count (descending)
	sort.Slice(types, func(i, j int) bool {
		return types[i].FileCount > types[j].FileCount
	})

	return types, nil
}

// GetDefaultContentType returns the default content type for a Hugo site
func GetDefaultContentType(sitePath string) string {
	types, err := DetectContentTypes(sitePath)
	if err != nil || len(types) == 0 {
		return "posts"
	}

	// Return the most used content type
	return types[0].Name
}

// countMarkdownFiles counts .md files recursively in a directory
func countMarkdownFiles(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			count++
		}
		return nil
	})
	return count
}

// InitializeSite creates a new Hugo site at the given path.
func InitializeSite(sitePath string) error {
	if _, err := exec.LookPath("hugo"); err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	cmd := exec.Command("hugo", "new", "site", ".", "--format", "toml")
	cmd.Dir = sitePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize Hugo site at %s: %v\nOutput: %s", sitePath, err, string(output))
	}

	fmt.Printf("Hugo `new site` command output:\n%s\n", string(output))

	return nil
}

// cleanPublicDir removes unnecessary files from the public directory
func cleanPublicDir(publicDir string) error {
	var deletedCount int

	err := filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file should be deleted
		ext := filepath.Ext(path)
		baseName := filepath.Base(path)

		shouldDelete := false
		switch {
		case ext == ".map":
			// Source map files
			shouldDelete = true
		case baseName == ".DS_Store":
			// macOS system files
			shouldDelete = true
		case ext == ".br", ext == ".gz":
			// Compressed files (.br, .gz)
			shouldDelete = true
		}

		if shouldDelete {
			if err := os.Remove(path); err != nil {
				// Log but don't fail the entire operation
				fmt.Printf("Warning: failed to delete %s: %v\n", path, err)
			} else {
				deletedCount++
			}
		}

		return nil
	})

	if deletedCount > 0 {
		fmt.Printf("Cleaned up %d unnecessary files from public directory\n", deletedCount)
	}

	return err
}

// BuildSite runs the Hugo build process in the given site path.
func BuildSite(sitePath string) error {
	if _, err := exec.LookPath("hugo"); err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	walgoCfg := filepath.Join(sitePath, "walgo.yaml")
	if _, err := os.Stat(walgoCfg); os.IsNotExist(err) {
		return fmt.Errorf("walgo.yaml not found in %s", sitePath)
	}

	content, err := os.ReadFile(walgoCfg)
	if err != nil {
		return fmt.Errorf("failed to read walgo.yaml: %w", err)
	}

	var walgoCfgData config.WalgoConfig
	if err := yaml.Unmarshal(content, &walgoCfgData); err != nil {
		return fmt.Errorf("failed to unmarshal walgo.yaml: %w", err)
	}

	configFile := filepath.Join(sitePath, "hugo.toml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configFile = filepath.Join(sitePath, "config.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("hugo configuration file (hugo.toml/config.toml) not found in %s. Are you in a Hugo site directory?", sitePath)
		}
	}

	fmt.Printf("Building Hugo site in %s...\n", sitePath)
	cmd := exec.Command("hugo", "build", "--environment", "production", "--minify", "--gc", "--cleanDestinationDir")
	cmd.Dir = sitePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build Hugo site: %v", err)
	}

	fmt.Println("Hugo site built successfully.")
	publicDir := filepath.Join(sitePath, "public")
	fmt.Printf("Static files generated in: %s\n", publicDir)

	if walgoCfgData.OptimizerConfig.Enabled {
		fmt.Printf("Optimizing assets...\n")
		optimizerEngine := optimizer.NewEngine(walgoCfgData.OptimizerConfig)
		stats, err := optimizerEngine.OptimizeDirectory(publicDir)

		if err != nil {
			return fmt.Errorf("failed to optimize directory: %w", err)
		} else {
			optimizerEngine.PrintStats(stats)
		}
	}

	if walgoCfgData.CompressConfig.GenerateWSResources {
		fmt.Printf("Generating ws-resources.json...\n")

		cacheConfig := compress.CacheControlConfig{
			Enabled:         walgoCfgData.CacheConfig.Enabled,
			ImmutableMaxAge: walgoCfgData.CacheConfig.ImmutableMaxAge,
			MutableMaxAge:   walgoCfgData.CacheConfig.MutableMaxAge,
		}

		if cacheConfig.ImmutableMaxAge == 0 {
			cacheConfig.ImmutableMaxAge = 31536000 // 1 year default
		}
		if cacheConfig.MutableMaxAge == 0 {
			cacheConfig.MutableMaxAge = 300 // 5 minutes default
		}

		if len(cacheConfig.ImmutablePatterns) == 0 {
			cacheConfig.ImmutablePatterns = compress.DefaultCacheControlConfig().ImmutablePatterns
		}

		wsOptions := compress.WSResourcesOptions{
			CacheConfig:  cacheConfig,
			CustomRoutes: walgoCfgData.CompressConfig.CustomRoutes,
			CustomIgnore: walgoCfgData.CompressConfig.IgnorePatterns,
		}

		// Try to get project metadata (non-fatal if not found)
		var existingObjectID string
		pm, err := projects.NewManager()
		if err == nil {
			defer pm.Close()
			project, err := pm.GetProjectBySitePath(sitePath)
			if err == nil && project != nil {
				// Use project metadata if available
				wsOptions.SiteName = project.Name
				wsOptions.Description = project.Description
				wsOptions.ImageURL = project.ImageURL
				wsOptions.Link = compress.DefaultLink
				wsOptions.ProjectURL = compress.DefaultProjectURL
				wsOptions.Creator = compress.DefaultCreator
				wsOptions.Category = project.Category
				existingObjectID = project.ObjectID
			} else {
				// Project not found - use defaults
				wsOptions.SiteName = filepath.Base(sitePath)
				wsOptions.Description = ""
				wsOptions.ImageURL = compress.DefaultImageURL
				wsOptions.Link = compress.DefaultLink
				wsOptions.ProjectURL = compress.DefaultProjectURL
				wsOptions.Creator = compress.DefaultCreator
				wsOptions.Category = compress.DefaultCategory
			}
		} else {
			// Project manager failed - use defaults
			fmt.Fprintf(os.Stderr, "Warning: Could not access project database: %v (using defaults)\n", err)
			wsOptions.SiteName = filepath.Base(sitePath)
			wsOptions.Description = ""
			wsOptions.ImageURL = compress.DefaultImageURL
			wsOptions.Link = compress.DefaultLink
			wsOptions.ProjectURL = compress.DefaultProjectURL
			wsOptions.Creator = compress.DefaultCreator
			wsOptions.Category = compress.DefaultCategory
		}

		// Check for existing ObjectID in old ws-resources.json
		outputPath := filepath.Join(publicDir, "ws-resources.json")
		if existingObjectID == "" {
			if oldConfig, err := compress.ReadWSResourcesConfig(outputPath); err == nil && oldConfig.ObjectID != "" {
				existingObjectID = oldConfig.ObjectID
			}
		}

		// Generate new ws-resources.json
		wsConfig, err := compress.GenerateWSResourcesConfig(publicDir, wsOptions)
		if err != nil {
			return fmt.Errorf("failed to generate ws-resources.json: %w", err)
		}

		// Preserve existing ObjectID if found
		if existingObjectID != "" {
			wsConfig.ObjectID = existingObjectID
		}

		// Write ws-resources.json (critical operation)
		if err := compress.WriteWSResourcesConfig(wsConfig, outputPath); err != nil {
			return fmt.Errorf("failed to write ws-resources.json: %w", err)
		}
		fmt.Printf("Generated ws-resources.json (%d resources)\n", len(wsConfig.Headers))

		// Generate and merge routes
		routes, err := compress.GenerateRoutesFromPublic(publicDir)
		if err != nil {
			return fmt.Errorf("failed to generate routes: %w", err)
		}

		if err := compress.MergeRoutesIntoWSResources(outputPath, routes); err != nil {
			return fmt.Errorf("failed to merge routes into ws-resources.json: %w", err)
		}
	}

	// Clean up unnecessary files from public directory
	if err := cleanPublicDir(publicDir); err != nil {
		fmt.Printf("Warning: failed to clean public directory: %v\n", err)
	}

	return nil
}

// ServeSite starts the Hugo development server
func ServeSite(sitePath string) error {
	if _, err := exec.LookPath("hugo"); err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	fmt.Printf("Starting Hugo development server...\n")
	cmd := exec.Command("hugo", "server", "--environment", "development", "--bind", "0.0.0.0",
		"--port", "1313", "--buildDrafts", "--buildFuture", "--disableFastRender", "--noHTTPCache")
	cmd.Dir = sitePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// CreateContent creates new content using Hugo's new command
func CreateContent(sitePath, contentPath string) error {
	if _, err := exec.LookPath("hugo"); err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	cmd := exec.Command("hugo", "new", contentPath)
	cmd.Dir = sitePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("creating content: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// SiteType represents the type of Hugo site
type SiteType string

const (
	SiteTypeBlog      SiteType = "blog"
	SiteTypePortfolio SiteType = "portfolio"
	SiteTypeDocs      SiteType = "docs"
	SiteTypeBusiness  SiteType = "business"
)

// DetermineSiteTypeFromPath analyzes a Hugo site and determines its type
func DetermineSiteTypeFromPath(sitePath string) SiteType {
	// Check for hugo.toml or config.toml
	configPath := filepath.Join(sitePath, "hugo.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join(sitePath, "config.toml")
	}

	// Read config file if it exists
	if content, err := os.ReadFile(configPath); err == nil {
		configStr := string(content)

		// Check for theme indicators
		if strings.Contains(configStr, "hugo-book") || strings.Contains(configStr, "BookTheme") {
			return SiteTypeDocs
		}
		if strings.Contains(configStr, "coder") || strings.Contains(configStr, "portfolio") {
			return SiteTypePortfolio
		}
		if strings.Contains(configStr, "business") {
			return SiteTypeBusiness
		}
	}

	// Check content structure
	contentDir := filepath.Join(sitePath, "content")
	if entries, err := os.ReadDir(contentDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				name := strings.ToLower(entry.Name())
				if name == "docs" || name == "documentation" {
					return SiteTypeDocs
				}
				if name == "portfolio" || name == "projects" {
					return SiteTypePortfolio
				}
				if name == "services" || name == "products" {
					return SiteTypeBusiness
				}
			}
		}
	}

	// Default to blog
	return SiteTypeBlog
}

// GetSiteConfig returns the embedded TOML config for a site type
func GetSiteConfig(siteType SiteType) ([]byte, error) {
	filename := fmt.Sprintf("tomls/%s.toml", siteType)
	return tomlsFS.ReadFile(filename)
}

// SetupSiteConfig copies the appropriate TOML config to the site directory
func SetupSiteConfig(sitePath string, siteType SiteType) error {
	return SetupSiteConfigWithName(sitePath, siteType, "")
}

// SetupSiteConfigWithName copies the appropriate TOML config to the site directory with site name
func SetupSiteConfigWithName(sitePath string, siteType SiteType, siteName string) error {
	configData, err := GetSiteConfig(siteType)
	if err != nil {
		return fmt.Errorf("getting config for %s: %w", siteType, err)
	}

	// Replace site name placeholder (all occurrences)
	configStr := string(configData)
	if siteName == "" {
		// Use directory name as default site name
		siteName = filepath.Base(sitePath)
	}
	configStr = strings.ReplaceAll(configStr, "{{SITE_NAME}}", siteName)

	configPath := filepath.Join(sitePath, "hugo.toml")
	if err := os.WriteFile(configPath, []byte(configStr), 0644); err != nil {
		return fmt.Errorf("writing hugo.toml: %w", err)
	}

	return nil
}

// SetupArchetypes copies the embedded archetypes to the site directory
func SetupArchetypes(sitePath string) error {
	archetypesDir := filepath.Join(sitePath, "archetypes")

	// Create archetypes directory if it doesn't exist
	if err := os.MkdirAll(archetypesDir, 0755); err != nil {
		return fmt.Errorf("creating archetypes directory: %w", err)
	}

	// Read all embedded archetypes
	entries, err := archetypesFS.ReadDir("archetypes")
	if err != nil {
		return fmt.Errorf("reading embedded archetypes: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Read the archetype content
		content, err := archetypesFS.ReadFile(filepath.Join("archetypes", entry.Name()))
		if err != nil {
			return fmt.Errorf("reading archetype %s: %w", entry.Name(), err)
		}

		// Write to site's archetypes directory
		destPath := filepath.Join(archetypesDir, entry.Name())
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("writing archetype %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// ThemeInfo contains information about a Hugo theme
type ThemeInfo struct {
	Name    string // Theme name
	DirName string // Directory name in themes/
	RepoURL string // Git repository URL
}

// GetThemeInfo returns theme information for a site type
func GetThemeInfo(siteType SiteType) ThemeInfo {
	switch siteType {
	case SiteTypeBlog:
		return ThemeInfo{
			Name:    "Ananke",
			DirName: "ananke",
			RepoURL: "https://github.com/theNewDynamic/gohugo-theme-ananke.git",
		}
	case SiteTypePortfolio:
		return ThemeInfo{
			Name:    "Ananke",
			DirName: "ananke",
			RepoURL: "https://github.com/theNewDynamic/gohugo-theme-ananke.git",
		}
	case SiteTypeDocs:
		return ThemeInfo{
			Name:    "Book",
			DirName: "hugo-book",
			RepoURL: "https://github.com/alex-shpak/hugo-book.git",
		}
	case SiteTypeBusiness:
		return ThemeInfo{
			Name:    "Ananke",
			DirName: "ananke",
			RepoURL: "https://github.com/theNewDynamic/gohugo-theme-ananke.git",
		}
	default:
		// Default to blog theme
		return ThemeInfo{
			Name:    "Ananke",
			DirName: "ananke",
			RepoURL: "https://github.com/theNewDynamic/gohugo-theme-ananke.git",
		}
	}
}

// InstallTheme clones a theme to the site's themes directory
func InstallTheme(sitePath string, siteType SiteType) error {
	theme := GetThemeInfo(siteType)
	themesDir := filepath.Join(sitePath, "themes")

	// Create themes directory if it doesn't exist
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return fmt.Errorf("creating themes directory: %w", err)
	}

	themePath := filepath.Join(themesDir, theme.DirName)

	// Check if theme already exists
	if _, err := os.Stat(themePath); err == nil {
		return nil // Theme already installed
	}

	// Clone the theme
	cmd := exec.Command("git", "clone", "--depth", "1", theme.RepoURL, themePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cloning theme %s: %w\nOutput: %s", theme.Name, err, string(output))
	}

	return nil
}

// SetupFavicon copies the embedded favicon to the site's static directory
func SetupFavicon(sitePath string) error {
	staticDir := filepath.Join(sitePath, "static")

	// Create static directory if it doesn't exist
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("creating static directory: %w", err)
	}

	// Write favicon to static directory
	faviconPath := filepath.Join(staticDir, "favicon.svg")
	if err := os.WriteFile(faviconPath, faviconSVG, 0644); err != nil {
		return fmt.Errorf("writing favicon: %w", err)
	}

	return nil
}

// UpdatePortfolioParams updates the hugo.toml params for portfolio sites with dynamic values
func UpdatePortfolioParams(sitePath, description, audience string) error {
	configPath := filepath.Join(sitePath, "hugo.toml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading hugo.toml: %w", err)
	}

	configStr := string(content)

	// Update description if provided
	if description != "" {
		// Replace the default description
		configStr = strings.Replace(configStr, `description = "Portfolio website"`, fmt.Sprintf(`description = "%s"`, description), 1)
	}

	// Generate dynamic info lines based on audience/description
	if audience != "" || description != "" {
		// Create info lines from audience or description
		infoLine1 := "Creative Professional"
		infoLine2 := "Building digital experiences"

		if audience != "" {
			// Use audience as first info line hint
			infoLine1 = audience
		}
		if description != "" && len(description) < 50 {
			infoLine2 = description
		}

		// Replace default info array
		oldInfo := `info = ["Designer & Developer", "Creating digital experiences"]`
		newInfo := fmt.Sprintf(`info = ["%s", "%s"]`, infoLine1, infoLine2)
		configStr = strings.Replace(configStr, oldInfo, newInfo, 1)
	}

	if err := os.WriteFile(configPath, []byte(configStr), 0644); err != nil {
		return fmt.Errorf("writing hugo.toml: %w", err)
	}

	return nil
}

// SetupFaviconForTheme copies the embedded favicon to the appropriate location for the theme
func SetupFaviconForTheme(sitePath string, siteType SiteType) error {
	var targetDir string

	switch siteType {
	case SiteTypePortfolio:
		// Coder theme expects favicon in /images/ folder
		targetDir = filepath.Join(sitePath, "static", "images")
	default:
		// Default location is /static/
		targetDir = filepath.Join(sitePath, "static")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("creating favicon directory: %w", err)
	}

	// Write favicon
	faviconPath := filepath.Join(targetDir, "favicon.svg")
	if err := os.WriteFile(faviconPath, faviconSVG, 0644); err != nil {
		return fmt.Errorf("writing favicon: %w", err)
	}

	return nil
}

// SetupPortfolioThemeOverrides creates necessary layout overrides for Ananke theme.
// This function is kept for backward compatibility but Ananke doesn't require overrides.
func SetupPortfolioThemeOverrides(sitePath string) error {
	// Ananke theme doesn't require specific layout overrides
	return nil
}

// SetupBusinessThemeOverrides creates necessary layout overrides for Ananke theme.
// This function is kept for backward compatibility but Ananke doesn't require overrides.
func SetupBusinessThemeOverrides(sitePath string) error {
	// Ananke theme doesn't require specific layout overrides
	return nil
}

// SetupDocsThemeOverrides creates necessary layout overrides for hugo-book theme
// This improves SVG favicon support with proper type specification
func SetupDocsThemeOverrides(sitePath string) error {
	layoutsDir := filepath.Join(sitePath, "layouts", "_partials", "docs")

	// Create layouts directory if it doesn't exist
	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		return fmt.Errorf("creating layouts directory: %w", err)
	}

	// Check if html-head-favicon.html already exists (don't overwrite user customizations)
	faviconPath := filepath.Join(layoutsDir, "html-head-favicon.html")
	if _, err := os.Stat(faviconPath); err == nil {
		return nil // Already exists, skip
	}

	// Create html-head-favicon.html override with proper SVG type
	faviconContent := `{{- $favicon := .Site.Params.BookFavicon | default "favicon.svg" -}}
<link rel="icon" href="{{ $favicon | relURL }}" type="image/svg+xml">
<link rel="apple-touch-icon" href="{{ $favicon | relURL }}">
`

	if err := os.WriteFile(faviconPath, []byte(faviconContent), 0644); err != nil {
		return fmt.Errorf("writing html-head-favicon.html: %w", err)
	}

	return nil
}

// UpdateDocsParams updates hugo-book theme params based on AI plan
func UpdateDocsParams(sitePath, description string) error {
	configPath := filepath.Join(sitePath, "hugo.toml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	configStr := string(content)

	// Update description in homepage _index.md if it exists
	indexPath := filepath.Join(sitePath, "content", "_index.md")
	if _, err := os.Stat(indexPath); err == nil {
		indexContent, err := os.ReadFile(indexPath)
		if err == nil {
			indexStr := string(indexContent)
			// Update description in frontmatter if empty or generic
			if strings.Contains(indexStr, "description: ''") || strings.Contains(indexStr, `description: ""`) {
				indexStr = strings.Replace(indexStr, "description: ''", fmt.Sprintf("description: '%s'", description), 1)
				indexStr = strings.Replace(indexStr, `description: ""`, fmt.Sprintf(`description: "%s"`, description), 1)
				if err := os.WriteFile(indexPath, []byte(indexStr), 0644); err != nil {
					return fmt.Errorf("writing index: %w", err)
				}
			}
		}
	}

	// Write back config if changed
	if configStr != string(content) {
		if err := os.WriteFile(configPath, []byte(configStr), 0644); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	}

	return nil
}
