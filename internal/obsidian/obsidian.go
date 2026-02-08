package obsidian

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/config"
)

// ImportStats tracks statistics for the Obsidian vault import operation.
type ImportStats struct {
	FilesProcessed    int
	FilesSkipped      int
	FilesError        int
	AttachmentsCopied int
}

// ImportVault imports markdown content from an Obsidian vault to Hugo content directory.
func ImportVault(vaultPath, hugoContentDir string, cfg config.ObsidianConfig) (*ImportStats, error) {
	stats := &ImportStats{}
	var processingErrors []string

	// Validate and sanitize vault path
	absVaultPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("invalid vault path: %w", err)
	}
	vaultPath = filepath.Clean(absVaultPath)

	// Validate vault exists and is a directory
	vaultInfo, err := os.Stat(vaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("obsidian vault path does not exist: %s", vaultPath)
		}
		return nil, fmt.Errorf("cannot access vault path: %w", err)
	}
	if !vaultInfo.IsDir() {
		return nil, fmt.Errorf("vault path is not a directory: %s", vaultPath)
	}

	// Determine site root from Hugo content directory
	var siteRoot string
	if filepath.Base(hugoContentDir) == "content" {
		// hugoContentDir is directly the content directory
		siteRoot = filepath.Dir(hugoContentDir)
	} else {
		// hugoContentDir is a subdirectory within content
		siteRoot = filepath.Dir(filepath.Dir(hugoContentDir))
	}

	// Create content directory if it doesn't exist
	// #nosec G301 - content directory needs standard permissions
	if err := os.MkdirAll(hugoContentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create Hugo content directory %s: %w", hugoContentDir, err)
	}

	// Determine and create attachments directory
	attachmentDir := cfg.AttachmentDir
	if attachmentDir == "" {
		attachmentDir = "attachments"
	}
	staticDir := filepath.Join(siteRoot, "static", attachmentDir)

	// #nosec G301 - static directory needs standard permissions
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create attachments directory %s: %w", staticDir, err)
	}

	// Walk through the vault directory
	err = filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}

		// Skip directories (handled implicitly by walk)
		if d.IsDir() {
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(filepath.Base(path), ".") {
			stats.FilesSkipped++
			return nil
		}

		// Process markdown files
		if filepath.Ext(path) == ".md" {
			if err := processMarkdownFile(path, vaultPath, hugoContentDir, cfg); err != nil {
				errMsg := fmt.Sprintf("failed to process %s: %v", path, err)
				processingErrors = append(processingErrors, errMsg)
				stats.FilesError++
			} else {
				stats.FilesProcessed++
			}
			return nil
		}

		// Process attachments (images, PDFs, etc.)
		if isAttachment(path) {
			if err := copyAttachment(path, vaultPath, staticDir, attachmentDir); err != nil {
				errMsg := fmt.Sprintf("failed to copy attachment %s: %v", path, err)
				processingErrors = append(processingErrors, errMsg)
			} else {
				stats.AttachmentsCopied++
			}
			return nil
		}

		// Skip other files
		stats.FilesSkipped++
		return nil
	})

	if err != nil {
		return stats, fmt.Errorf("error walking vault directory: %w", err)
	}

	// If there were processing errors but no fatal error, return stats with error context
	if len(processingErrors) > 0 {
		// Return stats but with an error indicating some files failed
		err := fmt.Errorf("import completed with %d errors", len(processingErrors))
		return stats, err
	}

	return stats, nil
}

// isAttachment determines whether a file is an attachment (image, PDF, video, audio).
func isAttachment(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	attachmentExts := []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".avif", ".svg", ".ico", ".pdf", ".mp4", ".mov", ".webm", ".mp3", ".wav", ".ogg"}

	for _, attachExt := range attachmentExts {
		if ext == attachExt {
			return true
		}
	}
	return false
}

// copyAttachment copies vault attachments to the Hugo static directory while preserving directory structure.
func copyAttachment(srcPath, vaultPath, staticDir, attachmentDir string) error {
	relPath, err := filepath.Rel(vaultPath, srcPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	destPath := filepath.Clean(filepath.Join(staticDir, relPath))
	if !strings.HasPrefix(destPath, filepath.Clean(staticDir)+string(os.PathSeparator)) &&
		destPath != filepath.Clean(staticDir) {
		return fmt.Errorf("path traversal detected: %s escapes static directory", relPath)
	}
	destDir := filepath.Dir(destPath)

	// #nosec G301 - attachment directory needs standard permissions
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create attachment directory %s: %w", destDir, err)
	}

	// #nosec G304 - srcPath comes from controlled directory walk
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read attachment %s: %w", srcPath, err)
	}

	// #nosec G306 - attachment files need to be readable by web servers
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write attachment %s: %w", destPath, err)
	}

	return nil
}

func processMarkdownFile(srcPath, vaultPath, hugoContentDir string, cfg config.ObsidianConfig) error {
	// #nosec G304 - srcPath comes from controlled directory walk
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", srcPath, err)
	}

	// Get relative path from vault
	relPath, err := filepath.Rel(vaultPath, srcPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path from vault: %w", err)
	}

	// Convert content
	convertedContent := string(content)

	// Convert wikilinks if enabled
	if cfg.ConvertWikilinks {
		// Use enhanced wikilink conversion with transclusion support
		// Use config's link style (defaults to "markdown" to avoid REF_NOT_FOUND errors)
		convertedContent = ConvertWikilinksWithConfig(convertedContent, cfg.AttachmentDir, cfg.LinkStyle)
	}

	// Parse existing frontmatter if present
	obsidianFM, body, hasFM := parseObsidianFrontmatter(convertedContent)

	// Generate Hugo frontmatter
	if hasFM {
		// Use enhanced frontmatter that preserves Obsidian metadata
		convertedContent = enhancedFrontmatter(obsidianFM, relPath, cfg.FrontmatterFormat) + body
	} else {
		// Add Hugo frontmatter if not present
		convertedContent = ensureFrontmatter(convertedContent, relPath, cfg.FrontmatterFormat)
	}

	// Create destination path and validate it stays within target directory
	destPath := filepath.Clean(filepath.Join(hugoContentDir, relPath))
	if !strings.HasPrefix(destPath, filepath.Clean(hugoContentDir)+string(os.PathSeparator)) &&
		destPath != filepath.Clean(hugoContentDir) {
		return fmt.Errorf("path traversal detected: %s escapes content directory", relPath)
	}
	destDir := filepath.Dir(destPath)

	// Create directory structure if needed
	// #nosec G301 - directory structure needs standard permissions
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	// Write converted content
	// #nosec G306 - markdown files need to be readable by web servers
	if err := os.WriteFile(destPath, []byte(convertedContent), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

// ensureFrontmatter adds Hugo frontmatter if it is not present in the file.
func ensureFrontmatter(content, filePath, format string) string {
	// Check if frontmatter already exists
	if strings.HasPrefix(content, "---") || strings.HasPrefix(content, "+++") || strings.HasPrefix(content, "{") {
		return content
	}

	// Generate frontmatter
	title := generateTitle(filePath)
	date := time.Now().Format("2006-01-02T15:04:05-07:00")

	var frontmatter string
	switch format {
	case "toml":
		frontmatter = fmt.Sprintf(`+++
title = "%s"
date = "%s"
draft = false
+++

`, title, date)
	case "json":
		frontmatter = fmt.Sprintf(`{
  "title": "%s",
  "date": "%s",
  "draft": false
}

`, title, date)
	default: // yaml
		frontmatter = fmt.Sprintf(`---
title: "%s"
date: %s
draft: false
---

`, title, date)
	}

	return frontmatter + content
}

// generateTitle creates a human-readable title from the file path.
func generateTitle(filePath string) string {
	// Get filename without extension
	filename := filepath.Base(filePath)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	// Replace hyphens and underscores with spaces
	title := strings.ReplaceAll(filename, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")

	// Capitalize first letter of each word
	words := strings.Fields(title)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}
