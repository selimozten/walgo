package hugo

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/selimozten/walgo/internal/ai"
)

// =============================================================================
// Menu Configuration Types and Constants
// =============================================================================

const (
	// menuBlockStart marks the beginning of the WALGO-generated menu section in hugo.toml
	menuBlockStart = "# BEGIN WALGO MENU\n"
	// menuBlockEnd marks the end of the WALGO-generated menu section in hugo.toml
	menuBlockEnd = "# END WALGO MENU\n"
)

// MenuItem represents a single menu item for Hugo configuration.
//
// Fields:
//   - Name: Display text for the menu item
//   - PageRef: Hugo page reference (absolute URL path, e.g., "/about/")
//   - Weight: Display order (lower numbers appear first)
type MenuItem struct {
	Name    string
	PageRef string
	Weight  int
}

// =============================================================================
// Menu TOML Rendering
// =============================================================================

// RenderMenuTOML renders menu items as TOML configuration suitable for Hugo.
// The output uses Hugo's menu.main syntax with name, pageRef, and weight attributes.
//
// Example output:
//
//	[menu]
//	  [[menu.main]]
//	    name = "About"
//	    pageRef = "/about/"
//	    weight = 30
//
// Parameters:
//   - items: Slice of MenuItem to render. Must not be nil.
//
// Returns: TOML-formatted string ready for inclusion in hugo.toml.
func RenderMenuTOML(items []MenuItem) string {
	var b strings.Builder
	b.WriteString("[menu]\n")

	for _, it := range items {
		b.WriteString("  [[menu.main]]\n")
		b.WriteString(fmt.Sprintf("    name = %q\n", it.Name))
		b.WriteString(fmt.Sprintf("    pageRef = %q\n", it.PageRef))
		b.WriteString(fmt.Sprintf("    weight = %d\n\n", it.Weight))
	}

	return b.String()
}

// =============================================================================
// Hugo Configuration File Management
// =============================================================================

// hasExistingMenuSection checks whether the TOML content already contains a
// [menu] or [[menu.*]] section outside of WALGO-managed markers.
func hasExistingMenuSection(content string) bool {
	// Strip WALGO block so we only inspect user-authored content
	stripped := content
	if start := strings.Index(stripped, menuBlockStart); start >= 0 {
		if end := strings.Index(stripped, menuBlockEnd); end > start {
			stripped = stripped[:start] + stripped[end+len(menuBlockEnd):]
		}
	}

	for _, line := range strings.Split(stripped, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == "[menu]" || strings.HasPrefix(trimmed, "[[menu.") {
			return true
		}
	}
	return false
}

// UpsertMenuBlockInHugoTOML writes menu TOML into hugo.toml using markers.
//
// This function implements an "upsert" (insert-or-update) pattern:
//   - If a WALGO MENU block already exists: replace it with new content
//   - If the config already contains a [menu] or [[menu.*]] section (not
//     managed by WALGO): skip — the existing menu is preserved untouched
//   - Otherwise: append a new block to the end of the file
//
// Parameters:
//   - hugoTomlPath: Path to the Hugo configuration file (hugo.toml or config.toml)
//   - menuTOML: Menu TOML content to write (excluding block markers)
//
// Returns:
//   - nil on success (including the "skip" case — no error when menu already exists)
//   - error on file I/O failure (wrapped with context about what operation failed)
//
// Note: The function preserves all existing content in hugo.toml outside the menu block.
func UpsertMenuBlockInHugoTOML(hugoTomlPath string, menuTOML string) error {
	// Read existing config
	data, err := os.ReadFile(hugoTomlPath)
	if err != nil {
		return fmt.Errorf("failed to read hugo config from %s: %w", hugoTomlPath, err)
	}

	content := string(data)
	block := menuBlockStart + menuTOML + menuBlockEnd

	// Replace existing block if present, otherwise append
	if strings.Contains(content, menuBlockStart) && strings.Contains(content, menuBlockEnd) {
		startIdx := strings.Index(content, menuBlockStart)
		endIdx := strings.Index(content, menuBlockEnd)
		if startIdx >= 0 && endIdx > startIdx {
			before := content[:startIdx]
			after := content[endIdx+len(menuBlockEnd):]
			content = before + block + after
		} else {
			// Markers in unexpected order; append new block instead
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += "\n" + block
		}
	} else {
		// No WALGO block — skip if config already has a menu section
		if hasExistingMenuSection(content) {
			return nil
		}
		// Ensure proper newline before appending
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + block
	}

	// Write back to file
	if err := os.WriteFile(hugoTomlPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write menu config to %s: %w", hugoTomlPath, err)
	}

	return nil
}

// =============================================================================
// Menu Building
// =============================================================================

// BuildProfessionalMainMenuFromPlan creates a theme-agnostic, professional menu structure
// based on the site plan. The menu structure varies by site type:
//
// Blog sites: Home → Posts → About → Contact
//   - Posts: included if any content/posts/ pages exist
//   - About/Contact: included if content/about.md and content/contact.md exist
//
// Docs sites: Home → Docs → Contact (optional)
//   - Docs: minimal top-level entry; sidebar handles detailed navigation
//   - Contact: only included if explicitly present
//
// Design principles:
//   - Only includes items that exist in the plan (prevents 404 links)
//   - Uses stable sorting by weight, then alphabetically by name
//   - Deduplicates by PageRef to handle malformed plans gracefully
//
// Parameters:
//   - plan: Site plan containing page definitions. Must not be nil.
//
// Returns:
//   - []MenuItem: Sorted, deduplicated menu items
//   - error: Always nil (kept for future extensibility)
func BuildProfessionalMainMenuFromPlan(plan *ai.SitePlan) ([]MenuItem, error) {
	// Build path lookup map for O(1) lookups
	pathMap := make(map[string]*ai.PageSpec)
	for i := range plan.Pages {
		pathMap[plan.Pages[i].Path] = &plan.Pages[i]
	}

	// Helper to check if a path exists in the plan
	pathExists := func(path string) bool {
		_, ok := pathMap[path]
		return ok
	}

	// Helper to check if any page exists with given prefix
	hasPrefix := func(prefix string) bool {
		for path := range pathMap {
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
		return false
	}

	// Helper to get title from page, with fallback
	titleFor := func(path, fallback string) string {
		if page, ok := pathMap[path]; ok && strings.TrimSpace(page.Title) != "" {
			return page.Title
		}
		return fallback
	}

	// Initialize with Home page (always present).
	// Always use "Home" as the label — the page title (often the site name)
	// is too long and duplicates the site branding already shown by the theme.
	items := []MenuItem{
		{Name: "Home", PageRef: "/", Weight: 10},
	}

	// Add type-specific menu items
	addIfExists := func(item MenuItem, condition bool) {
		if condition {
			items = append(items, item)
		}
	}

	switch plan.SiteType {
	case ai.SiteTypeBlog:
		addIfExists(MenuItem{Name: "Posts", PageRef: "/posts", Weight: 20},
			hasPrefix("content/posts/"))
		addIfExists(MenuItem{Name: "About", PageRef: "/about", Weight: 30},
			pathExists("content/about.md"))
		addIfExists(MenuItem{Name: "Contact", PageRef: "/contact", Weight: 40},
			pathExists("content/contact.md"))

	case ai.SiteTypeDocs:
		// Docs sites keep top menu minimal - sidebar handles detailed navigation
		docsTitle := titleFor("content/docs/_index.md", "Docs")
		addIfExists(MenuItem{Name: docsTitle, PageRef: "/docs", Weight: 20},
			pathExists("content/docs/_index.md") || hasPrefix("content/docs/"))
		// Contact is optional for docs sites
		addIfExists(MenuItem{Name: "Contact", PageRef: "/contact", Weight: 40},
			pathExists("content/contact.md"))

	default:
		// Fallback to safe minimal menu for unknown site types
		addIfExists(MenuItem{Name: "About", PageRef: "/about", Weight: 30},
			pathExists("content/about.md"))
		addIfExists(MenuItem{Name: "Contact", PageRef: "/contact", Weight: 40},
			pathExists("content/contact.md"))
	}

	// Sort: stable sort by weight, then alphabetically by name
	// This ensures menu remains consistent across runs and predictable for users
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Weight == items[j].Weight {
			return items[i].Name < items[j].Name
		}
		return items[i].Weight < items[j].Weight
	})

	// Deduplicate by PageRef to handle edge cases in plan data
	// This prevents duplicate menu items if plan contains redundant entries
	seen := make(map[string]bool, len(items))
	deduped := make([]MenuItem, 0, len(items))
	for _, item := range items {
		// Skip empty PageRef (malformed data) and already-seen items
		if item.PageRef == "" || seen[item.PageRef] {
			continue
		}
		seen[item.PageRef] = true
		deduped = append(deduped, item)
	}

	return deduped, nil
}

// ApplyMenuFromSitePlan builds and applies menu configuration from a SitePlan object.
// This is the preferred method when the plan is already in memory (e.g., from pipeline).
//
// Parameters:
//   - plan: Site plan containing page definitions. Must not be nil.
//   - hugoTomlPath: Path to hugo.toml configuration file
//
// Returns:
//   - nil on success
//   - error wrapped with context about what operation failed (build/write)
//
// Example:
//
//	plan, err := pipeline.Run(ctx, input)
//	if err != nil { return err }
//	if err := ApplyMenuFromSitePlan(plan.Plan, "hugo.toml"); err != nil {
//	    log.Printf("Warning: menu generation failed: %v", err)
//	}
func ApplyMenuFromSitePlan(plan *ai.SitePlan, hugoTomlPath string) error {
	items, err := BuildProfessionalMainMenuFromPlan(plan)
	if err != nil {
		return fmt.Errorf("failed to build menu: %w", err)
	}

	menuTOML := RenderMenuTOML(items)
	return UpsertMenuBlockInHugoTOML(hugoTomlPath, menuTOML)
}
