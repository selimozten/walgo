package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/config"
)

// ContentStructure holds information about the content directory structure
type ContentStructure struct {
	ContentTypes []ContentTypeInfo
	DefaultType  string
	ContentDir   string
}

// ContentTypeInfo holds information about a content type
type ContentTypeInfo struct {
	Name      string
	Path      string
	FileCount int
	Files     []string
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

// GetContentStructure scans and returns the content directory structure
func GetContentStructure(sitePath string) (*ContentStructure, error) {
	cfg, err := config.LoadConfigFrom(sitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	contentDir := filepath.Join(sitePath, cfg.HugoConfig.ContentDir)
	structure := &ContentStructure{
		ContentTypes: []ContentTypeInfo{},
		DefaultType:  "posts",
		ContentDir:   contentDir,
	}

	// Read content directory
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		// If content directory doesn't exist, return empty structure
		if os.IsNotExist(err) {
			return structure, nil
		}
		return nil, fmt.Errorf("failed to read content directory: %w", err)
	}

	// Scan each subdirectory
	for _, entry := range entries {
		if entry.IsDir() {
			typePath := filepath.Join(contentDir, entry.Name())
			files, err := os.ReadDir(typePath)
			if err != nil {
				continue
			}

			// Count markdown files and collect filenames
			var mdFiles []string
			fileCount := 0
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
					mdFiles = append(mdFiles, file.Name())
					fileCount++
				}
			}

			structure.ContentTypes = append(structure.ContentTypes, ContentTypeInfo{
				Name:      entry.Name(),
				Path:      typePath,
				FileCount: fileCount,
				Files:     mdFiles,
			})
		}
	}

	// Set default type based on common conventions
	for _, ct := range structure.ContentTypes {
		if ct.Name == "posts" || ct.Name == "post" || ct.Name == "blog" {
			structure.DefaultType = ct.Name
			break
		}
	}

	return structure, nil
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
	if !strings.HasPrefix(absSavePath, absContentDir) {
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

// buildSmartSystemPrompt creates a system prompt with content structure awareness
func (cg *ContentGenerator) buildSmartSystemPrompt(structure *ContentStructure) string {
	// Convert ContentTypeInfo to the format expected by BuildContentStructureInfo
	contentTypes := make([]struct {
		Name      string
		FileCount int
		Files     []string
	}, len(structure.ContentTypes))

	for i, ct := range structure.ContentTypes {
		contentTypes[i].Name = ct.Name
		contentTypes[i].FileCount = ct.FileCount
		contentTypes[i].Files = ct.Files
	}

	structureInfo := BuildContentStructureInfo(contentTypes)
	return BuildSmartContentPrompt(structureInfo)
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
