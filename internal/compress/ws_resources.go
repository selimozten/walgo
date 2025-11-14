package compress

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WSResource represents a single resource configuration for Walrus Sites
type WSResource struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

// WSResourcesConfig represents the ws-resources.json file structure
type WSResourcesConfig struct {
	Resources []WSResource                 `json:"resources,omitempty"`
	Routes    map[string]string            `json:"routes,omitempty"`
	Headers   map[string]map[string]string `json:"headers,omitempty"`
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

// GenerateWSResourcesConfig creates a ws-resources.json configuration
func GenerateWSResourcesConfig(siteDir string, compressionStats *DirectoryCompressionStats, cacheConfig CacheControlConfig) (*WSResourcesConfig, error) {
	config := &WSResourcesConfig{
		Headers: make(map[string]map[string]string),
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
		if compressionStats != nil {
			// Remove leading slash for lookup
			lookupPath := strings.TrimPrefix(relPath, "/")
			if result, ok := compressionStats.Files[lookupPath]; ok && result.Compressed {
				fileWasCompressed = true
			}
		}

		// Set Content-Encoding if compressed
		if fileWasCompressed {
			headers["Content-Encoding"] = "br"
		}

		// Set Cache-Control headers
		if cacheConfig.Enabled {
			cacheControl := getCacheControl(relPath, cacheConfig)
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
