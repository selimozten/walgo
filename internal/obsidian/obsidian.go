package obsidian

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"walgo/internal/config"
)

// ImportStats holds statistics about the import operation
type ImportStats struct {
	FilesProcessed    int
	FilesSkipped      int
	FilesError        int
	AttachmentsCopied int
}

// ImportVault imports content from an Obsidian vault to Hugo content directory
func ImportVault(vaultPath, hugoContentDir string, cfg config.ObsidianConfig) (*ImportStats, error) {
	stats := &ImportStats{}

	// Validate vault path
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("obsidian vault path does not exist: %s", vaultPath)
	}

	// Create content directory if it doesn't exist
	// #nosec G301 - content directory needs standard permissions
	if err := os.MkdirAll(hugoContentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create Hugo content directory: %w", err)
	}

	// Create attachments directory
	staticDir := filepath.Join(filepath.Dir(hugoContentDir), "static", cfg.AttachmentDir)
	// #nosec G301 - static directory needs standard permissions
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create attachments directory: %w", err)
	}

	// Walk through the vault directory
	err := filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Process only markdown files
		if filepath.Ext(path) != ".md" {
			// Check if it's an attachment (image, pdf, etc.)
			if isAttachment(path) {
				if err := copyAttachment(path, vaultPath, staticDir, cfg.AttachmentDir); err != nil {
					fmt.Printf("Warning: Failed to copy attachment %s: %v\n", path, err)
				} else {
					stats.AttachmentsCopied++
				}
			}
			return nil
		}

		// Process markdown file
		if err := processMarkdownFile(path, vaultPath, hugoContentDir, cfg); err != nil {
			fmt.Printf("Error processing file %s: %v\n", path, err)
			stats.FilesError++
		} else {
			stats.FilesProcessed++
		}

		return nil
	})

	if err != nil {
		return stats, fmt.Errorf("error walking vault directory: %w", err)
	}

	return stats, nil
}

// isAttachment checks if a file is an attachment (image, pdf, etc.)
func isAttachment(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	attachmentExts := []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".pdf", ".mp4", ".mov", ".mp3", ".wav"}

	for _, attachExt := range attachmentExts {
		if ext == attachExt {
			return true
		}
	}
	return false
}

// copyAttachment copies an attachment file to the static directory
func copyAttachment(srcPath, vaultPath, staticDir, attachmentDir string) error {
	// Get relative path from vault
	relPath, err := filepath.Rel(vaultPath, srcPath)
	if err != nil {
		return err
	}

	// Create destination path
	destPath := filepath.Join(staticDir, filepath.Base(relPath))

	// Read source file
	data, err := os.ReadFile(srcPath) // #nosec G304 - srcPath comes from controlled directory walk
	if err != nil {
		return err
	}

	// Write to destination
	return os.WriteFile(destPath, data, 0644) // #nosec G306 - attachment files need to be readable by web servers
}

// processMarkdownFile processes a single markdown file
func processMarkdownFile(srcPath, vaultPath, hugoContentDir string, cfg config.ObsidianConfig) error {
	// Read the file
	content, err := os.ReadFile(srcPath) // #nosec G304 - srcPath comes from controlled directory walk
	if err != nil {
		return err
	}

	// Get relative path from vault
	relPath, err := filepath.Rel(vaultPath, srcPath)
	if err != nil {
		return err
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

	// Create destination path
	destPath := filepath.Join(hugoContentDir, relPath)
	destDir := filepath.Dir(destPath)

	// Create directory structure if needed
	// #nosec G301 - directory structure needs standard permissions
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Write converted content
	return os.WriteFile(destPath, []byte(convertedContent), 0644) // #nosec G306 - markdown files need to be readable by web servers
}

// convertWikilinks converts Obsidian [[wikilinks]] to Hugo markdown links
func convertWikilinks(content, attachmentDir string) string {
	// Regex for [[link]] or [[link|display text]]
	wikilinkRegex := regexp.MustCompile(`\[\[([^|\]]+)(\|([^\]]*))?\]\]`)

	return wikilinkRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatch := wikilinkRegex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}

		target := strings.TrimSpace(submatch[1])
		displayText := target

		// If there's a display text (|text), use it
		if len(submatch) >= 4 && submatch[3] != "" {
			displayText = strings.TrimSpace(submatch[3])
		}

		// Handle attachments
		if isAttachment(target) {
			return fmt.Sprintf("![%s](/%s/%s)", displayText, attachmentDir, filepath.Base(target))
		}

		// Handle regular page links
		// Convert to Hugo-style link (assuming content structure)
		linkPath := strings.ToLower(strings.ReplaceAll(target, " ", "-"))
		if !strings.HasSuffix(linkPath, ".md") {
			linkPath += ".md"
		}

		return fmt.Sprintf("[%s]({{< relref \"%s\" >}})", displayText, linkPath)
	})
}

// ensureFrontmatter adds Hugo frontmatter if not present
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

// generateTitle generates a title from the file path
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
