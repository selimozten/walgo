package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/ui"
)

// LinkStyle determines how wikilinks are converted
type LinkStyle string

const (
	// LinkStyleRelref uses Hugo's relref shortcode (strict, throws REF_NOT_FOUND if target missing)
	LinkStyleRelref LinkStyle = "relref"
	// LinkStyleMarkdown uses plain markdown links (permissive, works even if target missing)
	LinkStyleMarkdown LinkStyle = "markdown"

	// defaultLinkStyle is the default link conversion style
	// Using markdown by default to avoid REF_NOT_FOUND errors
	defaultLinkStyle LinkStyle = LinkStyleMarkdown
)

// ConvertWikilinksWithConfig converts wikilinks using the config's link style setting
func ConvertWikilinksWithConfig(content, attachmentDir, linkStyleStr string) string {
	style := defaultLinkStyle
	if linkStyleStr == "relref" {
		style = LinkStyleRelref
	}
	return convertWikilinksWithStyle(content, attachmentDir, style)
}

// convertWikilinksWithStyle converts wikilinks using the specified link style
func convertWikilinksWithStyle(content, attachmentDir string, style LinkStyle) string {
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
		// Validate submatch length (5 groups + full match = 6)
		if len(submatch) < 6 {
			return match
		}

		target := strings.TrimSpace(submatch[1])
		heading := ""
		blockID := ""
		displayText := target

		if submatch[2] != "" {
			heading = strings.TrimSpace(strings.TrimPrefix(submatch[2], "#"))
		}

		if submatch[3] != "" {
			blockID = strings.TrimSpace(strings.TrimPrefix(submatch[3], "^"))
		}

		if submatch[5] != "" {
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

		// Generate link based on style
		switch style {
		case LinkStyleRelref:
			// Use Hugo's relref for internal links (strict - throws error if target missing)
			return fmt.Sprintf("[%s]({{< relref \"%s.md%s\" >}})", displayText, linkPath, anchor)
		case LinkStyleMarkdown:
			// Use plain markdown links (permissive - works even if target missing)
			// This avoids REF_NOT_FOUND errors during Hugo build
			return fmt.Sprintf("[%s](%s.md%s)", displayText, linkPath, anchor)
		default:
			// Default to markdown style for safety (same as LinkStyleMarkdown)
			return fmt.Sprintf("[%s](%s.md%s)", displayText, linkPath, anchor)
		}
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
	TotalFiles         int
	MarkdownFiles      int
	Attachments        int
	WouldProcess       int
	WouldSkip          int
	WikilinksFound     int
	TransclusionsFound int
	EstimatedSize      int64
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
			content, err := os.ReadFile(path) // #nosec G304 - Reading user-specified Obsidian vault files is intended behavior
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
	icons := ui.GetIcons()
	fmt.Printf("\n%s Dry-run Import Summary:\n", icons.Info)
	fmt.Printf("  Total files found: %d (%.2f MB)\n", s.TotalFiles, float64(s.EstimatedSize)/(1024*1024))
	fmt.Printf("  %s Markdown files: %d\n", icons.File, s.MarkdownFiles)
	fmt.Printf("  %s Attachments: %d\n", icons.Package, s.Attachments)
	fmt.Printf("  %s Would skip: %d\n", icons.Cross, s.WouldSkip)
	fmt.Printf("\n  %s Wikilinks found: %d\n", icons.Book, s.WikilinksFound)
	fmt.Printf("  %s Transclusions found: %d\n", icons.Book, s.TransclusionsFound)
	fmt.Printf("\n%s Would process %d files\n", icons.Check, s.WouldProcess)
}
