package compress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WSResource represents a single resource configuration for Walrus Sites
type WSResource struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

// WSMetadata represents the metadata section for displaying site info on wallets/explorers
// Fields correspond to the Sui Display object standard
type WSMetadata struct {
	Link        string `json:"link,omitempty"`        // Link to your site's homepage
	ImageURL    string `json:"image_url,omitempty"`   // URL to your site's logo/favicon
	Description string `json:"description,omitempty"` // Brief description of your site
	ProjectURL  string `json:"project_url,omitempty"` // Link to GitHub repository
	Creator     string `json:"creator,omitempty"`     // Name of creator/company
	Category    string `json:"category,omitempty"`    // Site category (e.g., website, blog, portfolio)
}

// WSResourcesConfig represents the ws-resources.json file structure
// Note: All field names use snake_case as required by site-builder
type WSResourcesConfig struct {
	Headers   map[string]map[string]string `json:"headers,omitempty"`   // Custom HTTP response headers per resource
	Routes    map[string]string            `json:"routes,omitempty"`    // Client-side routing rules for SPAs
	Metadata  *WSMetadata                  `json:"metadata,omitempty"`  // Site metadata for wallets/explorers
	SiteName  string                       `json:"site_name,omitempty"` // Name for your Walrus Site
	ObjectID  string                       `json:"object_id,omitempty"` // Sui Object ID of deployed site
	Ignore    []string                     `json:"ignore,omitempty"`    // Files/folders to exclude from upload
	Resources []WSResource                 `json:"resources,omitempty"` // Legacy field for compatibility
}

// CacheControlConfig holds cache control settings
type CacheControlConfig struct {
	Enabled bool
	// MaxAge for immutable assets (hashed filenames)
	ImmutableMaxAge int // Default: 31536000 (1 year)
	// MaxAge for HTML and other mutable content
	MutableMaxAge int // Default: 300 (5 minutes)
	// File patterns that should have immutable caching
	ImmutablePatterns []string
}

// DefaultCacheControlConfig returns default cache control settings
func DefaultCacheControlConfig() CacheControlConfig {
	return CacheControlConfig{
		Enabled:         true,
		ImmutableMaxAge: 31536000, // 1 year
		MutableMaxAge:   300,      // 5 minutes
		ImmutablePatterns: []string{
			// Patterns for files with content hashes in filename
			"*.*.min.js",  // e.g., app.abc123.min.js
			"*.*.min.css", // e.g., style.def456.min.css
			"*.*.js",
			"*.*.css",
			// Font files (rarely change)
			"*.woff2",
			"*.woff",
			"*.ttf",
			"*.eot",
		},
	}
}

// DefaultIgnorePatterns returns patterns for files that should not be uploaded to Walrus Sites
func DefaultIgnorePatterns() []string {
	return []string{
		"/.DS_Store",     // macOS metadata
		"/Thumbs.db",     // Windows thumbnails
		"/.gitkeep",      // Git placeholder files
		"/.gitignore",    // Git ignore file
		"/.git/*",        // Git directory (if somehow in publish dir)
		"/*.map",         // Source maps in root
		"/js/*.map",      // JavaScript source maps
		"/css/*.map",     // CSS source maps
		"/assets/*.map",  // Asset source maps
		"/.well-known/*", // Usually not needed for static sites
	}
}

// WSResourcesOptions holds options for generating ws-resources.json
type WSResourcesOptions struct {
	SiteName         string
	Description      string
	ImageURL         string
	Link             string
	ProjectURL       string
	Creator          string
	Category         string
	CompressionStats *DirectoryCompressionStats
	CacheConfig      CacheControlConfig
	CustomRoutes     map[string]string
	CustomIgnore     []string
}

// GenerateWSResourcesConfig creates a ws-resources.json configuration
func GenerateWSResourcesConfig(siteDir string, opts WSResourcesOptions) (*WSResourcesConfig, error) {
	// Start with default ignore patterns and merge custom ones
	ignorePatterns := DefaultIgnorePatterns()
	if len(opts.CustomIgnore) > 0 {
		ignorePatterns = append(ignorePatterns, opts.CustomIgnore...)
	}

	// Set defaults for metadata fields
	creator := opts.Creator
	if creator == "" {
		creator = DefaultCreator
	}
	link := opts.Link
	if link == "" {
		link = DefaultLink
	}
	projectURL := opts.ProjectURL
	if projectURL == "" {
		projectURL = DefaultProjectURL
	}
	imageURL := opts.ImageURL
	if imageURL == "" {
		imageURL = DefaultImageURL
	}

	config := &WSResourcesConfig{
		Headers:  make(map[string]map[string]string),
		Ignore:   ignorePatterns,
		SiteName: opts.SiteName,
		Metadata: &WSMetadata{
			Description: opts.Description,
			ImageURL:    imageURL,
			Link:        link,
			ProjectURL:  projectURL,
			Creator:     creator,
			Category:    opts.Category,
		},
	}

	// Walk through the site directory
	err := filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(siteDir, path)
		if err != nil {
			return err
		}

		// Normalize path for Walrus (use forward slashes)
		relPath = filepath.ToSlash(relPath)
		if !strings.HasPrefix(relPath, "/") {
			relPath = "/" + relPath
		}

		headers := make(map[string]string)

		// Set Content-Type based on extension
		contentType := getContentType(path)
		if contentType != "" {
			headers["Content-Type"] = contentType
		}

		// Check if file was compressed
		fileWasCompressed := false
		if opts.CompressionStats != nil {
			// Remove leading slash for lookup
			lookupPath := strings.TrimPrefix(relPath, "/")
			if result, ok := opts.CompressionStats.Files[lookupPath]; ok && result.Compressed {
				fileWasCompressed = true
			}
		}

		// Set Content-Encoding if compressed
		if fileWasCompressed {
			headers["Content-Encoding"] = "br"
		}

		// Set Cache-Control headers
		if opts.CacheConfig.Enabled {
			cacheControl := getCacheControl(relPath, opts.CacheConfig)
			if cacheControl != "" {
				headers["Cache-Control"] = cacheControl
			}
		}

		// Only add headers if there are any
		if len(headers) > 0 {
			config.Headers[relPath] = headers
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Only add custom routes if user specified them (for SPA sites)
	// Hugo static sites don't need routes - each page has its own HTML file
	if len(opts.CustomRoutes) > 0 {
		config.Routes = make(map[string]string)
		for pattern, target := range opts.CustomRoutes {
			config.Routes[pattern] = target
		}
	}

	return config, nil
}

// WriteWSResourcesConfig writes the configuration to ws-resources.json
func WriteWSResourcesConfig(config *WSResourcesConfig, outputPath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// #nosec G306 - config file needs to be readable
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func GenerateRoutesFromPublic(publicDir string) (map[string]string, error) {
	routes := map[string]string{}

	// Root routes: always useful
	routes["/"] = "/index.html"
	routes["/index.html"] = "/index.html"

	// Discover all index.html under public/
	err := filepath.WalkDir(publicDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "index.html" {
			return nil
		}

		rel, err := filepath.Rel(publicDir, p)
		if err != nil {
			return err
		}

		rel = filepath.ToSlash(rel)                   // e.g. "docs/intro/index.html"
		dir := strings.TrimSuffix(rel, "/index.html") // e.g. "docs/intro"
		if dir == "" || dir == "index.html" {
			// This is the root index.html; already handled above
			return nil
		}

		// Web path for directory
		webDir := "/" + dir                 // "/docs/intro"
		target := "/" + dir + "/index.html" // "/docs/intro/index.html"

		// Add routes: clean URL and wildcard for trailing slash/subpaths
		addRoute(routes, webDir, target)
		addRoute(routes, webDir+"/*", target)

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Add fallback only if 404.html exists
	if _, err := os.Stat(filepath.Join(publicDir, "404.html")); err == nil {
		routes["*"] = "/404.html"
	}

	return routes, nil
}

func addRoute(routes map[string]string, key, target string) {
	// Normalize: no double slashes (except "http://", irrelevant here)
	key = strings.ReplaceAll(key, "//", "/")
	if _, exists := routes[key]; !exists {
		routes[key] = target
	}
}

func MergeRoutesIntoWSResources(wsResourcesPath string, newRoutes map[string]string) error {
	// Read existing file
	raw, err := os.ReadFile(wsResourcesPath)
	if err != nil {
		return fmt.Errorf("read ws-resources: %w", err)
	}

	// Parse into generic map to preserve unknown fields
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fmt.Errorf("parse ws-resources json: %w", err)
	}

	// Set/overwrite routes
	routeObj := make(map[string]any, len(newRoutes))
	for k, v := range newRoutes {
		routeObj[k] = v
	}
	obj["routes"] = routeObj

	// Write deterministically with ordered "routes" keys (and "*" last)
	out, err := marshalDeterministicWSResources(obj)
	if err != nil {
		return err
	}

	if err := os.WriteFile(wsResourcesPath, out, 0o644); err != nil {
		return fmt.Errorf("write ws-resources: %w", err)
	}
	return nil
}

func marshalDeterministicWSResources(obj map[string]any) ([]byte, error) {
	// Extract routes for special handling (sorted keys, "*" last)
	var routes map[string]any
	if r, ok := obj["routes"]; ok {
		if rm, ok := r.(map[string]any); ok {
			routes = rm
		}
	}

	// Define the desired field order
	fieldOrder := []string{"headers", "ignore", "routes", "metadata", "object_id", "site_name"}

	// Collect any extra keys not in our predefined order
	extraKeys := []string{}
	for k := range obj {
		found := false
		for _, f := range fieldOrder {
			if k == f {
				found = true
				break
			}
		}
		if !found {
			extraKeys = append(extraKeys, k)
		}
	}
	sort.Strings(extraKeys)

	// Build ordered list of keys that exist in obj
	orderedKeys := []string{}
	for _, k := range fieldOrder {
		if _, exists := obj[k]; exists {
			orderedKeys = append(orderedKeys, k)
		}
	}
	orderedKeys = append(orderedKeys, extraKeys...)

	var buf bytes.Buffer
	buf.WriteString("{\n")

	for i, k := range orderedKeys {
		buf.WriteString(fmt.Sprintf("  %q: ", k))

		if k == "routes" && routes != nil {
			// Write routes with sorted keys and "*" last
			routeKeys := make([]string, 0, len(routes))
			for rk := range routes {
				if rk != "*" {
					routeKeys = append(routeKeys, rk)
				}
			}
			sort.Strings(routeKeys)

			buf.WriteString("{\n")
			for j, rk := range routeKeys {
				buf.WriteString(fmt.Sprintf("    %q: %q", rk, routes[rk].(string)))
				if j < len(routeKeys)-1 || routes["*"] != nil {
					buf.WriteString(",")
				}
				buf.WriteString("\n")
			}
			// "*" last if present
			if star, ok := routes["*"]; ok {
				buf.WriteString(fmt.Sprintf("    %q: %q\n", "*", star.(string)))
			}
			buf.WriteString("  }")
		} else {
			// Marshal with indentation for other values
			vBytes, err := json.MarshalIndent(obj[k], "  ", "  ")
			if err != nil {
				return nil, fmt.Errorf("marshal field %s: %w", k, err)
			}
			buf.Write(vBytes)
		}

		if i < len(orderedKeys)-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteString("\n")
		}
	}

	buf.WriteString("}\n")
	return buf.Bytes(), nil
}

// wsResourcesLegacy is used for backward compatibility with old objectId field
type wsResourcesLegacy struct {
	ObjectIDLegacy string `json:"objectId,omitempty"` // Legacy camelCase field
}

// ReadWSResourcesConfig reads ws-resources.json from disk
// Supports both snake_case (object_id) and legacy camelCase (objectId) formats
func ReadWSResourcesConfig(path string) (*WSResourcesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ws-resources.json: %w", err)
	}

	var config WSResourcesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse ws-resources.json: %w", err)
	}

	// Check for legacy objectId field if object_id is empty
	if config.ObjectID == "" {
		var legacy wsResourcesLegacy
		if err := json.Unmarshal(data, &legacy); err == nil && legacy.ObjectIDLegacy != "" {
			config.ObjectID = legacy.ObjectIDLegacy
		}
	}

	return &config, nil
}

// UpdateObjectID updates the objectID in an existing ws-resources.json file
func UpdateObjectID(wsResourcesPath string, objectID string) error {
	config, err := ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		// If file doesn't exist, create minimal config
		if os.IsNotExist(err) {
			config = &WSResourcesConfig{
				Headers: make(map[string]map[string]string),
			}
		} else {
			return err
		}
	}

	config.ObjectID = objectID
	return WriteWSResourcesConfig(config, wsResourcesPath)
}

// MetadataOptions holds options for updating site metadata
type MetadataOptions struct {
	ObjectID    string
	SiteName    string
	Description string
	ImageURL    string
	Link        string // Site URL (e.g., https://yoursite.wal.app)
	ProjectURL  string // GitHub or project page URL
	Creator     string // Creator name (defaults to "Walgo")
	Category    string // Site category (e.g., website, blog, portfolio)
}

// DefaultCreator is the default creator name for sites deployed with Walgo
const DefaultCreator = "Walgo"
const DefaultLink = "https://walgo.xyz"
const DefaultProjectURL = "https://github.com/selimozten/walgo"
const DefaultImageURL = "https://walgo.xyz/walgo-logo.png"

// UpdateMetadata updates all metadata fields in ws-resources.json
func UpdateMetadata(wsResourcesPath string, opts MetadataOptions) error {
	config, err := ReadWSResourcesConfig(wsResourcesPath)
	if err != nil {
		// If file doesn't exist, create minimal config
		if os.IsNotExist(err) {
			config = &WSResourcesConfig{
				Headers: make(map[string]map[string]string),
				Ignore:  DefaultIgnorePatterns(),
			}
		} else {
			return err
		}
	}

	// Update object_id if provided
	if opts.ObjectID != "" {
		config.ObjectID = opts.ObjectID
	}

	// Update site_name if provided
	if opts.SiteName != "" {
		config.SiteName = opts.SiteName
	}

	// Initialize metadata if nil
	if config.Metadata == nil {
		config.Metadata = &WSMetadata{}
	}

	// Update metadata fields
	if opts.Description != "" {
		config.Metadata.Description = opts.Description
	}
	if opts.ImageURL != "" {
		config.Metadata.ImageURL = opts.ImageURL
	} else if config.Metadata.ImageURL == "" {
		config.Metadata.ImageURL = DefaultImageURL
	}
	if opts.Link != "" {
		config.Metadata.Link = opts.Link
	} else if config.Metadata.Link == "" {
		config.Metadata.Link = DefaultLink
	}
	if opts.ProjectURL != "" {
		config.Metadata.ProjectURL = opts.ProjectURL
	} else if config.Metadata.ProjectURL == "" {
		config.Metadata.ProjectURL = DefaultProjectURL
	}
	if opts.Category != "" {
		config.Metadata.Category = opts.Category
	}

	// Set creator (default to "Walgo" if not specified)
	if opts.Creator != "" {
		config.Metadata.Creator = opts.Creator
	} else if config.Metadata.Creator == "" {
		config.Metadata.Creator = DefaultCreator
	}

	return WriteWSResourcesConfig(config, wsResourcesPath)
}

// getContentType returns the appropriate Content-Type for a file
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	contentTypes := map[string]string{
		".html": "text/html; charset=utf-8",
		".htm":  "text/html; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".js":   "application/javascript; charset=utf-8",
		".mjs":  "application/javascript; charset=utf-8",
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain; charset=utf-8",
		".md":   "text/markdown; charset=utf-8",

		// Images
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",

		// Fonts
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".ttf":   "font/ttf",
		".eot":   "application/vnd.ms-fontobject",

		// Media
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",

		// Archives
		".pdf": "application/pdf",
		".zip": "application/zip",
		".tar": "application/x-tar",
		".gz":  "application/gzip",

		// WebAssembly
		".wasm": "application/wasm",
	}

	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}

	return "application/octet-stream"
}

// getCacheControl returns the appropriate Cache-Control header value
func getCacheControl(path string, config CacheControlConfig) string {
	if !config.Enabled {
		return ""
	}

	// Check if file matches immutable patterns
	filename := filepath.Base(path)
	for _, pattern := range config.ImmutablePatterns {
		if matchPattern(filename, pattern) {
			return fmt.Sprintf("public, max-age=%d, immutable", config.ImmutableMaxAge)
		}
	}

	// HTML files and other mutable content
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".html" || ext == ".htm" {
		return fmt.Sprintf("public, max-age=%d, must-revalidate", config.MutableMaxAge)
	}

	// Default: moderate caching for other files
	return fmt.Sprintf("public, max-age=%d", config.MutableMaxAge)
}

// matchPattern checks if a filename matches a glob-like pattern
// Simplified version - only supports * wildcard
func matchPattern(filename, pattern string) bool {
	// Simple wildcard matching
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return filename == pattern
	}

	// Check if filename starts with first part
	if !strings.HasPrefix(filename, parts[0]) {
		return false
	}

	// Check if filename ends with last part
	if !strings.HasSuffix(filename, parts[len(parts)-1]) {
		return false
	}

	// For more complex patterns, check if all parts exist in order
	pos := 0
	for i, part := range parts {
		if part == "" {
			continue
		}

		if i == 0 {
			pos = len(part)
			continue
		}

		idx := strings.Index(filename[pos:], part)
		if idx == -1 {
			return false
		}
		pos += idx + len(part)
	}

	return true
}
