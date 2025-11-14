package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"walgo/internal/config"
)

// convertWikilinksEnhanced converts Obsidian [[wikilinks]] with enhanced support for:
// - Aliases: [[link|display text]]
// - Headings: [[note#heading]]
// - Blocks: [[note^block-id]]
// - Transclusions: ![[note]] or ![[note#heading]]
func convertWikilinksEnhanced(content, attachmentDir string) string {
	// Handle transclusions first (![[...]])
	// Regex for transclusion: ![[target]] or ![[target#heading]]
	transclusionRegex := regexp.MustCompile(`!\[\[([^\]|#^]+)(#[^\]|]+)?(\|([^\]]*))?\]\]`)

	content = transclusionRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatch := transclusionRegex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}

		target := strings.TrimSpace(submatch[1])
		heading := ""
		if len(submatch) > 2 && submatch[2] != "" {
			heading = strings.TrimSpace(strings.TrimPrefix(submatch[2], "#"))
		}

		// Check if it's an attachment
		if isAttachment(target) {
			// Transclusion of image/media - embed it
			return fmt.Sprintf("![%s](/%s/%s)", target, attachmentDir, filepath.Base(target))
		}

		// For markdown transclusions, we can't directly embed in Hugo
		// Convert to a link with a note about transclusion
		if heading != "" {
			return fmt.Sprintf("\n> **Transcluded from [[%s#%s]]**\n> _(Original content was transcluded here)_\n", target, heading)
		}
		return fmt.Sprintf("\n> **Transcluded from [[%s]]**\n> _(Original content was transcluded here)_\n", target)
	})

	// Handle regular wikilinks: [[link]] or [[link|display text]] or [[link#heading]]
	// This regex captures: [[target#heading|display]]
	wikilinkRegex := regexp.MustCompile(`\[\[([^\]|#^]+)(#[^\]|^]+)?(\^[^\]|]+)?(\|([^\]]*))?\]\]`)

	content = wikilinkRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatch := wikilinkRegex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}

		target := strings.TrimSpace(submatch[1])
		heading := ""
		blockID := ""
		displayText := target

		// Extract heading if present (#heading)
		if len(submatch) > 2 && submatch[2] != "" {
			heading = strings.TrimSpace(strings.TrimPrefix(submatch[2], "#"))
		}

		// Extract block ID if present (^block-id)
		if len(submatch) > 3 && submatch[3] != "" {
			blockID = strings.TrimSpace(strings.TrimPrefix(submatch[3], "^"))
		}

		// Use custom display text if provided (|text)
		if len(submatch) >= 6 && submatch[5] != "" {
			displayText = strings.TrimSpace(submatch[5])
		} else if heading != "" {
			displayText = fmt.Sprintf("%s - %s", target, heading)
		}

		// Handle attachments
		if isAttachment(target) {
			return fmt.Sprintf("![%s](/%s/%s)", displayText, attachmentDir, filepath.Base(target))
		}

		// Handle regular page links
		linkPath := strings.ToLower(strings.ReplaceAll(target, " ", "-"))

		// Add heading anchor if present
		anchor := ""
		if heading != "" {
			anchor = "#" + strings.ToLower(strings.ReplaceAll(heading, " ", "-"))
		} else if blockID != "" {
			// Block references become anchors
			anchor = "#" + blockID
		}

		// Use Hugo's relref for internal links
		return fmt.Sprintf("[%s]({{< relref \"%s.md%s\" >}})", displayText, linkPath, anchor)
	})

	return content
}

// parseObsidianFrontmatter extracts and normalizes Obsidian frontmatter
func parseObsidianFrontmatter(content string) (frontmatter map[string]string, body string, hasFrontmatter bool) {
	frontmatter = make(map[string]string)

	// Check if content starts with YAML frontmatter
	if !strings.HasPrefix(content, "---") {
		return frontmatter, content, false
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return frontmatter, content, false
	}

	// Find end of frontmatter
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return frontmatter, content, false
	}

	// Parse frontmatter
	for i := 1; i < endIdx; i++ {
		line := lines[i]
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				// Remove quotes if present
				value = strings.Trim(value, "\"'")
				frontmatter[key] = value
			}
		}
	}

	// Return body without frontmatter
	body = strings.Join(lines[endIdx+1:], "\n")
	return frontmatter, body, true
}

// enhancedFrontmatter creates Hugo-compatible frontmatter with Obsidian metadata
func enhancedFrontmatter(obsidianFM map[string]string, filePath, format string) string {
	// Start with existing frontmatter or create new
	title := obsidianFM["title"]
	if title == "" {
		title = generateTitle(filePath)
	}

	date := obsidianFM["date"]
	if date == "" {
		date = getCurrentDate()
	}

	tags := obsidianFM["tags"]
	aliases := obsidianFM["aliases"]

	var frontmatter string
	switch format {
	case "toml":
		frontmatter = fmt.Sprintf(`+++
title = "%s"
date = "%s"
draft = false
`, title, date)
		if tags != "" {
			frontmatter += fmt.Sprintf(`tags = [%s]
`, formatTags(tags))
		}
		if aliases != "" {
			frontmatter += fmt.Sprintf(`aliases = [%s]
`, formatAliases(aliases))
		}
		frontmatter += "+++\n\n"

	case "json":
		frontmatter = `{
  "title": "` + title + `",
  "date": "` + date + `",
  "draft": false`
		if tags != "" {
			frontmatter += `,
  "tags": [` + formatTags(tags) + `]`
		}
		if aliases != "" {
			frontmatter += `,
  "aliases": [` + formatAliases(aliases) + `]`
		}
		frontmatter += "\n}\n\n"

	default: // yaml
		frontmatter = fmt.Sprintf(`---
title: "%s"
date: %s
draft: false
`, title, date)
		if tags != "" {
			frontmatter += fmt.Sprintf(`tags: [%s]
`, formatTags(tags))
		}
		if aliases != "" {
			frontmatter += fmt.Sprintf(`aliases: [%s]
`, formatAliases(aliases))
		}
		frontmatter += "---\n\n"
	}

	return frontmatter
}

// formatTags formats tags for Hugo frontmatter
func formatTags(tags string) string {
	// Handle various tag formats: "tag1, tag2" or "#tag1 #tag2" or "[tag1, tag2]"
	tags = strings.Trim(tags, "[]")
	tags = strings.ReplaceAll(tags, "#", "")

	parts := strings.Split(tags, ",")
	quoted := make([]string, 0, len(parts))
	for _, tag := range parts {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			quoted = append(quoted, fmt.Sprintf(`"%s"`, tag))
		}
	}
	return strings.Join(quoted, ", ")
}

// formatAliases formats aliases for Hugo
func formatAliases(aliases string) string {
	// Similar to tags
	aliases = strings.Trim(aliases, "[]")
	parts := strings.Split(aliases, ",")
	quoted := make([]string, 0, len(parts))
	for _, alias := range parts {
		alias = strings.TrimSpace(alias)
		if alias != "" {
			quoted = append(quoted, fmt.Sprintf(`"%s"`, alias))
		}
	}
	return strings.Join(quoted, ", ")
}

// getCurrentDate returns current date in Hugo format
func getCurrentDate() string {
	return time.Now().Format("2006-01-02T15:04:05-07:00")
}

// DryRunStats holds statistics for dry-run mode
type DryRunStats struct {
	TotalFiles      int
	MarkdownFiles   int
	Attachments     int
	WouldProcess    int
	WouldSkip       int
	WikilinksFound  int
	TransclusionsFound int
	EstimatedSize   int64
}

// DryRunImport simulates an import without actually copying files
func DryRunImport(vaultPath, hugoContentDir string, cfg config.ObsidianConfig) (*DryRunStats, error) {
	stats := &DryRunStats{}

	err := filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		stats.TotalFiles++
		stats.EstimatedSize += info.Size()

		if filepath.Ext(path) == ".md" {
			stats.MarkdownFiles++

			// Check for wikilinks and transclusions
			content, err := os.ReadFile(path)
			if err == nil {
				contentStr := string(content)
				stats.WikilinksFound += strings.Count(contentStr, "[[")
				stats.TransclusionsFound += strings.Count(contentStr, "![[")
			}

			stats.WouldProcess++
		} else if isAttachment(path) {
			stats.Attachments++
			stats.WouldProcess++
		} else {
			stats.WouldSkip++
		}

		return nil
	})

	return stats, err
}

// PrintDryRunStats prints dry-run statistics
func (s *DryRunStats) PrintSummary() {
	fmt.Println("\n🔍 Dry-run Import Summary:")
	fmt.Printf("  Total files found: %d (%.2f MB)\n", s.TotalFiles, float64(s.EstimatedSize)/(1024*1024))
	fmt.Printf("  📝 Markdown files: %d\n", s.MarkdownFiles)
	fmt.Printf("  📎 Attachments: %d\n", s.Attachments)
	fmt.Printf("  ⏭️  Would skip: %d\n", s.WouldSkip)
	fmt.Printf("\n  🔗 Wikilinks found: %d\n", s.WikilinksFound)
	fmt.Printf("  📋 Transclusions found: %d\n", s.TransclusionsFound)
	fmt.Printf("\n✅ Would process %d files\n", s.WouldProcess)
}
