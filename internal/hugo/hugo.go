package hugo

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
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

	configFile := filepath.Join(sitePath, "hugo.toml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configFile = filepath.Join(sitePath, "config.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("hugo configuration file (hugo.toml/config.toml) not found in %s. Are you in a Hugo site directory?", sitePath)
		}
	}

	fmt.Printf("Building Hugo site in %s...\n", sitePath)
	cmd := exec.Command("hugo", "build", "--minify", "--gc", "--cleanDestinationDir")
	cmd.Dir = sitePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build Hugo site: %v", err)
	}

	fmt.Println("Hugo site built successfully.")
	publicDir := filepath.Join(sitePath, "public")
	fmt.Printf("Static files generated in: %s\n", publicDir)

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
	cmd := exec.Command("hugo", "server", "-D")
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

	// Replace site name placeholder
	configStr := string(configData)
	if siteName == "" {
		// Use directory name as default site name
		siteName = filepath.Base(sitePath)
	}
	configStr = strings.Replace(configStr, "{{SITE_NAME}}", siteName, 1)

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
			Name:    "Coder",
			DirName: "hugo-coder",
			RepoURL: "https://github.com/luizdepra/hugo-coder.git",
		}
	case SiteTypeDocs:
		return ThemeInfo{
			Name:    "Book",
			DirName: "hugo-book",
			RepoURL: "https://github.com/alex-shpak/hugo-book.git",
		}
	case SiteTypeBusiness:
		return ThemeInfo{
			Name:    "Starter",
			DirName: "hugo-starter-theme",
			RepoURL: "https://github.com/zerostaticthemes/hugo-starter-theme.git",
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
