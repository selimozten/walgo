package walrus

import (
	"regexp"
	"strings"
	"time"
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

// sitemapColumnSeparator splits the columns of the sitemap table, which are
// padded apart with at least two spaces.
var sitemapColumnSeparator = regexp.MustCompile(`\s{2,}`)

// parseSitemapOutput extracts resources from sitemap command output.
//
// site-builder prints a table:
//
//	Resource path   Blob / Quilt Patch ID   Owned blob object ID (if any)   Earliest Expiration Date
//	/index.html     0nLObHz-...             0x69f3aff9...                   2026-08-11
//
// Older builds printed one "resource <path> with blob ID <id>" line per entry,
// which is still recognized.
func parseSitemapOutput(output string) *SiteBuilderOutput {
	result := &SiteBuilderOutput{
		Resources: make([]Resource, 0),
	}

	for _, line := range strings.Split(string(stripANSI([]byte(output))), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if resource, ok := parseSitemapTableRow(line); ok {
			result.Resources = append(result.Resources, resource)
			continue
		}

		if resource, ok := parseSitemapLegacyLine(line); ok {
			result.Resources = append(result.Resources, resource)
		}
	}

	return result
}

// parseSitemapTableRow parses one row of the sitemap table. Header, separator
// and non-resource lines are rejected.
func parseSitemapTableRow(line string) (Resource, bool) {
	if !strings.HasPrefix(line, "/") {
		return Resource{}, false // resource paths are absolute; headers are not
	}

	columns := sitemapColumnSeparator.Split(line, -1)
	if len(columns) < 2 {
		return Resource{}, false
	}

	resource := Resource{
		Path:   strings.TrimSpace(columns[0]),
		BlobID: strings.TrimSpace(columns[1]),
	}

	if len(columns) > 2 {
		if objectID := strings.TrimSpace(columns[2]); strings.HasPrefix(objectID, "0x") {
			resource.BlobObjectID = objectID
		}
	}

	if len(columns) > 3 {
		if expiry, err := time.Parse(time.DateOnly, strings.TrimSpace(columns[3])); err == nil {
			resource.Expiry = expiry
		}
	}

	if resource.Path == "" || resource.BlobID == "" {
		return Resource{}, false
	}

	return resource, true
}

// parseSitemapLegacyLine parses the pre-table sitemap format.
func parseSitemapLegacyLine(line string) (Resource, bool) {
	if !strings.Contains(line, "blob ID") {
		return Resource{}, false
	}

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

	if path == "" || blobID == "" {
		return Resource{}, false
	}

	return Resource{Path: path, BlobID: blobID}, true
}
