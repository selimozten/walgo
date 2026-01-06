package walrus

import (
	"regexp"
	"strings"
)

// parseSiteBuilderOutput extracts key information from site-builder command output.
// This parser is designed to be resilient to format changes in site-builder output.
func parseSiteBuilderOutput(output string) *SiteBuilderOutput {
	result := &SiteBuilderOutput{
		BrowseURLs: make([]string, 0),
		Resources:  make([]Resource, 0),
	}

	lines := strings.Split(output, "\n")

	siteObjectPatterns := []string{
		"New site object ID:",
		"Site object ID:",
		"site object ID:",
	}

	objectIDRegex := regexp.MustCompile(`0x[0-9a-fA-F]{64}`)

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])

		if result.ObjectID == "" {
			for _, pattern := range siteObjectPatterns {
				if strings.Contains(line, pattern) {
					if match := objectIDRegex.FindString(line); match != "" {
						result.ObjectID = match
						break
					}
				}
			}
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "http://") || strings.Contains(line, "https://") {
			urlRegex := regexp.MustCompile(`https?://[^\s\]\)\"\']+`)
			urls := urlRegex.FindAllString(line, -1)
			for _, url := range urls {
				url = strings.TrimRight(url, ".,;:")
				result.BrowseURLs = append(result.BrowseURLs, url)
			}
		}
	}

	return result
}

// parseSitemapOutput extracts resources from sitemap command output.
func parseSitemapOutput(output string) *SiteBuilderOutput {
	result := &SiteBuilderOutput{
		Resources: make([]Resource, 0),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "blob ID") {
			parts := strings.Fields(line)
			var path, blobID string

			for i, part := range parts {
				if part == "resource" && i+1 < len(parts) {
					path = parts[i+1]
				}
				if part == "ID" && i+1 < len(parts) {
					blobID = parts[i+1]
				}
			}

			if path != "" && blobID != "" {
				result.Resources = append(result.Resources, Resource{
					Path:   path,
					BlobID: blobID,
				})
			}
		}
	}

	return result
}
