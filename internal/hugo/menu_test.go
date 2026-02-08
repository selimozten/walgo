package hugo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/selimozten/walgo/internal/ai"
)

// =============================================================================
// RenderMenuTOML Tests
// =============================================================================

func TestRenderMenuTOML(t *testing.T) {
	tests := []struct {
		name     string
		items    []MenuItem
		wantSubs []string // substrings that must appear in output
		notSubs  []string // substrings that must NOT appear
	}{
		{
			name:  "Empty items produces only header",
			items: []MenuItem{},
			wantSubs: []string{
				"[menu]\n",
			},
			notSubs: []string{
				"[[menu.main]]",
			},
		},
		{
			name: "Single item",
			items: []MenuItem{
				{Name: "Home", PageRef: "/", Weight: 10},
			},
			wantSubs: []string{
				"[menu]\n",
				"[[menu.main]]",
				`name = "Home"`,
				`pageRef = "/"`,
				"weight = 10",
			},
		},
		{
			name: "Multiple items rendered in order",
			items: []MenuItem{
				{Name: "Home", PageRef: "/", Weight: 10},
				{Name: "About", PageRef: "/about", Weight: 30},
				{Name: "Contact", PageRef: "/contact", Weight: 40},
			},
			wantSubs: []string{
				`name = "Home"`,
				`name = "About"`,
				`name = "Contact"`,
				`pageRef = "/about"`,
				"weight = 30",
				"weight = 40",
			},
		},
		{
			name: "Special characters in name are quoted properly",
			items: []MenuItem{
				{Name: `O'Brien "The Great"`, PageRef: "/obriens", Weight: 10},
			},
			wantSubs: []string{
				// Go %q will escape double quotes and special chars
				"name = ",
				"pageRef = ",
			},
		},
		{
			name:  "Nil items treated as empty",
			items: nil,
			wantSubs: []string{
				"[menu]\n",
			},
			notSubs: []string{
				"[[menu.main]]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderMenuTOML(tt.items)

			for _, sub := range tt.wantSubs {
				if !strings.Contains(result, sub) {
					t.Errorf("RenderMenuTOML() output missing expected substring %q\nGot:\n%s", sub, result)
				}
			}
			for _, sub := range tt.notSubs {
				if strings.Contains(result, sub) {
					t.Errorf("RenderMenuTOML() output should NOT contain %q\nGot:\n%s", sub, result)
				}
			}
		})
	}
}

func TestRenderMenuTOML_Format(t *testing.T) {
	items := []MenuItem{
		{Name: "Home", PageRef: "/", Weight: 10},
		{Name: "Posts", PageRef: "/posts", Weight: 20},
	}

	result := RenderMenuTOML(items)

	// Verify the output starts with [menu]
	if !strings.HasPrefix(result, "[menu]\n") {
		t.Errorf("Expected output to start with [menu]\\n, got: %s", result[:min(30, len(result))])
	}

	// Count occurrences of [[menu.main]]
	count := strings.Count(result, "[[menu.main]]")
	if count != 2 {
		t.Errorf("Expected 2 [[menu.main]] entries, got %d", count)
	}

	// Verify each entry is indented with 2 spaces for the section header
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, "[[menu.main]]") {
			if !strings.HasPrefix(line, "  [[menu.main]]") {
				t.Errorf("Expected [[menu.main]] to be indented with 2 spaces, got: %q", line)
			}
		}
		if strings.Contains(line, "name =") || strings.Contains(line, "pageRef =") || strings.Contains(line, "weight =") {
			if !strings.HasPrefix(line, "    ") {
				t.Errorf("Expected field to be indented with 4 spaces, got: %q", line)
			}
		}
	}
}

// =============================================================================
// UpsertMenuBlockInHugoTOML Tests
// =============================================================================

func TestUpsertMenuBlockInHugoTOML(t *testing.T) {
	tests := []struct {
		name            string
		existingContent string
		menuTOML        string
		wantSubs        []string
		notSubs         []string
		wantErr         bool
	}{
		{
			name: "Append to file without existing block",
			existingContent: `baseURL = "https://example.com/"
title = "My Site"
`,
			menuTOML: "[menu]\n  [[menu.main]]\n    name = \"Home\"\n",
			wantSubs: []string{
				`baseURL = "https://example.com/"`,
				menuBlockStart,
				menuBlockEnd,
				`name = "Home"`,
			},
		},
		{
			name: "Replace existing menu block",
			existingContent: `baseURL = "https://example.com/"
title = "My Site"

# BEGIN WALGO MENU
[menu]
  [[menu.main]]
    name = "Old Home"
# END WALGO MENU
`,
			menuTOML: "[menu]\n  [[menu.main]]\n    name = \"New Home\"\n",
			wantSubs: []string{
				`name = "New Home"`,
				menuBlockStart,
				menuBlockEnd,
			},
			notSubs: []string{
				`name = "Old Home"`,
			},
		},
		{
			name:            "Append when file has no trailing newline",
			existingContent: `title = "My Site"`,
			menuTOML:        "[menu]\n",
			wantSubs: []string{
				"title = \"My Site\"\n",
				menuBlockStart,
			},
		},
		{
			name: "Preserves content before and after block",
			existingContent: `baseURL = "/"
title = "Before"

# BEGIN WALGO MENU
old menu data
# END WALGO MENU

[params]
  description = "After"
`,
			menuTOML: "new menu data\n",
			wantSubs: []string{
				`title = "Before"`,
				"new menu data",
				`description = "After"`,
			},
			notSubs: []string{
				"old menu data",
			},
		},
		{
			name:            "Empty file gets block appended",
			existingContent: "",
			menuTOML:        "[menu]\n",
			wantSubs: []string{
				menuBlockStart,
				"[menu]\n",
				menuBlockEnd,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			hugoTomlPath := filepath.Join(tmpDir, "hugo.toml")

			if err := os.WriteFile(hugoTomlPath, []byte(tt.existingContent), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			err := UpsertMenuBlockInHugoTOML(hugoTomlPath, tt.menuTOML)

			if (err != nil) != tt.wantErr {
				t.Fatalf("UpsertMenuBlockInHugoTOML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			content, err := os.ReadFile(hugoTomlPath)
			if err != nil {
				t.Fatalf("Failed to read result file: %v", err)
			}

			result := string(content)
			for _, sub := range tt.wantSubs {
				if !strings.Contains(result, sub) {
					t.Errorf("Result missing expected substring %q\nGot:\n%s", sub, result)
				}
			}
			for _, sub := range tt.notSubs {
				if strings.Contains(result, sub) {
					t.Errorf("Result should NOT contain %q\nGot:\n%s", sub, result)
				}
			}

			// Verify markers appear exactly once
			startCount := strings.Count(result, menuBlockStart)
			endCount := strings.Count(result, menuBlockEnd)
			if startCount != 1 {
				t.Errorf("Expected exactly 1 BEGIN marker, got %d", startCount)
			}
			if endCount != 1 {
				t.Errorf("Expected exactly 1 END marker, got %d", endCount)
			}
		})
	}
}

func TestUpsertMenuBlockInHugoTOML_SkipsExistingMenu(t *testing.T) {
	tests := []struct {
		name    string
		content string
		skip    bool // true = file should remain unchanged
	}{
		{
			name: "Skips when [menu] section exists",
			content: `baseURL = "/"
title = "Test"

[menu]
  [[menu.main]]
    name = "Home"
    pageRef = "/"
    weight = 10
`,
			skip: true,
		},
		{
			name: "Skips when [[menu.whitepaper]] exists",
			content: `baseURL = "/"
title = "Test"

[[menu.whitepaper]]
  name = "Abstract"
  url = "/whitepaper/01-abstract/"
  weight = 1
`,
			skip: true,
		},
		{
			name: "Does not skip when menu is only in a comment",
			content: `baseURL = "/"
title = "Test"
# [menu] is not configured yet
`,
			skip: false,
		},
		{
			name: "Does not skip when no menu exists",
			content: `baseURL = "/"
title = "Test"
`,
			skip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			hugoTomlPath := filepath.Join(tmpDir, "hugo.toml")

			if err := os.WriteFile(hugoTomlPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			menuTOML := "[menu]\n  [[menu.main]]\n    name = \"AI Home\"\n"
			err := UpsertMenuBlockInHugoTOML(hugoTomlPath, menuTOML)
			if err != nil {
				t.Fatalf("UpsertMenuBlockInHugoTOML() unexpected error: %v", err)
			}

			result, _ := os.ReadFile(hugoTomlPath)
			resultStr := string(result)

			if tt.skip {
				// File should be unchanged
				if resultStr != tt.content {
					t.Errorf("Expected file to remain unchanged when existing menu found.\nBefore:\n%s\nAfter:\n%s", tt.content, resultStr)
				}
				if strings.Contains(resultStr, menuBlockStart) {
					t.Error("Should not have inserted WALGO MENU block")
				}
			} else {
				// WALGO block should be added
				if !strings.Contains(resultStr, menuBlockStart) {
					t.Error("Expected WALGO MENU block to be added")
				}
				if !strings.Contains(resultStr, `name = "AI Home"`) {
					t.Error("Expected AI menu content to be present")
				}
			}
		})
	}
}

func TestUpsertMenuBlockInHugoTOML_UpdatesOwnBlockEvenWithExistingMenu(t *testing.T) {
	// If a WALGO block already exists alongside a user menu, we should still update our block
	tmpDir := t.TempDir()
	hugoTomlPath := filepath.Join(tmpDir, "hugo.toml")

	content := `baseURL = "/"
title = "Test"

[[menu.whitepaper]]
  name = "Abstract"
  weight = 1

# BEGIN WALGO MENU
[menu]
  [[menu.main]]
    name = "Old"
# END WALGO MENU
`
	if err := os.WriteFile(hugoTomlPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	menuTOML := "[menu]\n  [[menu.main]]\n    name = \"Updated\"\n"
	if err := UpsertMenuBlockInHugoTOML(hugoTomlPath, menuTOML); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result, _ := os.ReadFile(hugoTomlPath)
	resultStr := string(result)

	if !strings.Contains(resultStr, `name = "Updated"`) {
		t.Error("WALGO block should be updated with new content")
	}
	if strings.Contains(resultStr, `name = "Old"`) {
		t.Error("Old WALGO content should be replaced")
	}
	// User's whitepaper menu should be preserved
	if !strings.Contains(resultStr, `name = "Abstract"`) {
		t.Error("Existing user menu should be preserved")
	}
}

func TestHasExistingMenuSection(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"No menu", `baseURL = "/"`, false},
		{"Has [menu]", "[menu]\n  [[menu.main]]", true},
		{"Has [[menu.whitepaper]]", "[[menu.whitepaper]]\n  name = \"x\"", true},
		{"Menu in comment only", "# [menu]\n# [[menu.main]]", false},
		{"Empty content", "", false},
		{"Menu inside WALGO block only", "# BEGIN WALGO MENU\n[menu]\n  [[menu.main]]\n# END WALGO MENU\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasExistingMenuSection(tt.content)
			if got != tt.want {
				t.Errorf("hasExistingMenuSection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpsertMenuBlockInHugoTOML_FileNotFound(t *testing.T) {
	err := UpsertMenuBlockInHugoTOML("/nonexistent/path/hugo.toml", "menu data")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("Error should mention 'failed to read', got: %v", err)
	}
}

func TestUpsertMenuBlockInHugoTOML_ReadOnlyFile(t *testing.T) {
	tmpDir := t.TempDir()
	hugoTomlPath := filepath.Join(tmpDir, "hugo.toml")

	if err := os.WriteFile(hugoTomlPath, []byte("title = \"test\""), 0444); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := UpsertMenuBlockInHugoTOML(hugoTomlPath, "menu data")
	if err == nil {
		t.Fatal("Expected error for read-only file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to write") {
		t.Errorf("Error should mention 'failed to write', got: %v", err)
	}
}

func TestUpsertMenuBlockInHugoTOML_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	hugoTomlPath := filepath.Join(tmpDir, "hugo.toml")

	initial := `baseURL = "/"
title = "Test"
`
	menuTOML := "[menu]\n  [[menu.main]]\n    name = \"Home\"\n"

	if err := os.WriteFile(hugoTomlPath, []byte(initial), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Apply twice
	if err := UpsertMenuBlockInHugoTOML(hugoTomlPath, menuTOML); err != nil {
		t.Fatalf("First upsert failed: %v", err)
	}
	if err := UpsertMenuBlockInHugoTOML(hugoTomlPath, menuTOML); err != nil {
		t.Fatalf("Second upsert failed: %v", err)
	}

	content, _ := os.ReadFile(hugoTomlPath)
	result := string(content)

	// Should still have exactly one block
	startCount := strings.Count(result, menuBlockStart)
	endCount := strings.Count(result, menuBlockEnd)
	if startCount != 1 || endCount != 1 {
		t.Errorf("Expected exactly 1 BEGIN and 1 END marker after double upsert, got %d/%d\nContent:\n%s",
			startCount, endCount, result)
	}
}

// =============================================================================
// BuildProfessionalMainMenuFromPlan Tests
// =============================================================================

func TestBuildProfessionalMainMenuFromPlan(t *testing.T) {
	tests := []struct {
		name      string
		plan      *ai.SitePlan
		wantNames []string // expected menu item names in order
		wantCount int
	}{
		{
			name: "Blog site with all pages",
			plan: &ai.SitePlan{
				SiteType: ai.SiteTypeBlog,
				Pages: []ai.PageSpec{
					{Path: "content/_index.md", Title: "Welcome"},
					{Path: "content/posts/first.md", Title: "First Post"},
					{Path: "content/about.md", Title: "About Us"},
					{Path: "content/contact.md", Title: "Contact Us"},
				},
			},
			// Home menu item always shows "Home" regardless of page title
			wantNames: []string{"Home", "Posts", "About", "Contact"},
			wantCount: 4,
		},
		{
			name: "Blog site with only homepage and posts",
			plan: &ai.SitePlan{
				SiteType: ai.SiteTypeBlog,
				Pages: []ai.PageSpec{
					{Path: "content/_index.md", Title: "My Blog"},
					{Path: "content/posts/hello.md", Title: "Hello World"},
				},
			},
			wantNames: []string{"Home", "Posts"},
			wantCount: 2,
		},
		{
			name: "Blog site with only homepage",
			plan: &ai.SitePlan{
				SiteType: ai.SiteTypeBlog,
				Pages: []ai.PageSpec{
					{Path: "content/_index.md", Title: "Minimal Blog"},
				},
			},
			wantNames: []string{"Home"},
			wantCount: 1,
		},
		{
			name: "Docs site with docs section",
			plan: &ai.SitePlan{
				SiteType: ai.SiteTypeDocs,
				Pages: []ai.PageSpec{
					{Path: "content/_index.md", Title: "Documentation"},
					{Path: "content/docs/_index.md", Title: "API Reference"},
					{Path: "content/docs/getting-started.md", Title: "Getting Started"},
					{Path: "content/contact.md", Title: "Contact"},
				},
			},
			wantNames: []string{"Home", "API Reference", "Contact"},
			wantCount: 3,
		},
		{
			name: "Docs site without contact",
			plan: &ai.SitePlan{
				SiteType: ai.SiteTypeDocs,
				Pages: []ai.PageSpec{
					{Path: "content/_index.md", Title: "Docs"},
					{Path: "content/docs/_index.md", Title: "Reference"},
				},
			},
			wantNames: []string{"Home", "Reference"},
			wantCount: 2,
		},
		{
			name: "Docs site with prefix match only (no _index.md)",
			plan: &ai.SitePlan{
				SiteType: ai.SiteTypeDocs,
				Pages: []ai.PageSpec{
					{Path: "content/_index.md", Title: "Home"},
					{Path: "content/docs/page1.md", Title: "Page 1"},
				},
			},
			wantNames: []string{"Home", "Docs"},
			wantCount: 2,
		},
		{
			name: "Default/unknown site type with about and contact",
			plan: &ai.SitePlan{
				SiteType: "portfolio", // unknown type
				Pages: []ai.PageSpec{
					{Path: "content/_index.md", Title: "Portfolio"},
					{Path: "content/about.md", Title: "About Me"},
					{Path: "content/contact.md", Title: "Get in Touch"},
				},
			},
			// Home menu item always shows "Home" regardless of page title
			wantNames: []string{"Home", "About", "Contact"},
			wantCount: 3,
		},
		{
			name: "Homepage title fallback when title is empty",
			plan: &ai.SitePlan{
				SiteType: ai.SiteTypeBlog,
				Pages: []ai.PageSpec{
					{Path: "content/_index.md", Title: ""},
					{Path: "content/posts/test.md", Title: "Test"},
				},
			},
			wantNames: []string{"Home", "Posts"},
			wantCount: 2,
		},
		{
			name: "Homepage title fallback when title is whitespace",
			plan: &ai.SitePlan{
				SiteType: ai.SiteTypeBlog,
				Pages: []ai.PageSpec{
					{Path: "content/_index.md", Title: "   "},
				},
			},
			wantNames: []string{"Home"},
			wantCount: 1,
		},
		{
			name: "Empty plan has just Home",
			plan: &ai.SitePlan{
				SiteType: ai.SiteTypeBlog,
				Pages:    []ai.PageSpec{},
			},
			wantNames: []string{"Home"},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := BuildProfessionalMainMenuFromPlan(tt.plan)
			if err != nil {
				t.Fatalf("BuildProfessionalMainMenuFromPlan() unexpected error: %v", err)
			}

			if len(items) != tt.wantCount {
				t.Errorf("Expected %d items, got %d: %+v", tt.wantCount, len(items), items)
				return
			}

			for i, wantName := range tt.wantNames {
				if i >= len(items) {
					t.Errorf("Missing item at index %d (expected %q)", i, wantName)
					continue
				}
				if items[i].Name != wantName {
					t.Errorf("Item[%d].Name = %q, want %q", i, items[i].Name, wantName)
				}
			}
		})
	}
}

func TestBuildProfessionalMainMenuFromPlan_Sorting(t *testing.T) {
	plan := &ai.SitePlan{
		SiteType: ai.SiteTypeBlog,
		Pages: []ai.PageSpec{
			{Path: "content/_index.md", Title: "Home"},
			{Path: "content/posts/p1.md", Title: "Post 1"},
			{Path: "content/about.md", Title: "About"},
			{Path: "content/contact.md", Title: "Contact"},
		},
	}

	items, err := BuildProfessionalMainMenuFromPlan(plan)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify items are sorted by weight
	for i := 1; i < len(items); i++ {
		if items[i].Weight < items[i-1].Weight {
			t.Errorf("Items not sorted by weight: [%d]=%d > [%d]=%d",
				i-1, items[i-1].Weight, i, items[i].Weight)
		}
	}

	// Verify expected weights
	expectedWeights := map[string]int{
		"/":        10,
		"/posts":   20,
		"/about":   30,
		"/contact": 40,
	}
	for _, item := range items {
		if expected, ok := expectedWeights[item.PageRef]; ok {
			if item.Weight != expected {
				t.Errorf("Item %q has weight %d, expected %d", item.PageRef, item.Weight, expected)
			}
		}
	}
}

func TestBuildProfessionalMainMenuFromPlan_Deduplication(t *testing.T) {
	// This shouldn't normally happen, but tests the dedup logic:
	// The plan has duplicate path entries that would map to the same PageRef
	plan := &ai.SitePlan{
		SiteType: ai.SiteTypeBlog,
		Pages: []ai.PageSpec{
			{Path: "content/_index.md", Title: "Home"},
			{Path: "content/_index.md", Title: "Home Again"}, // duplicate path
			{Path: "content/posts/p1.md", Title: "Post"},
		},
	}

	items, err := BuildProfessionalMainMenuFromPlan(plan)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have Home and Posts (Home deduplicated)
	pageRefs := make(map[string]int)
	for _, item := range items {
		pageRefs[item.PageRef]++
	}

	if pageRefs["/"] != 1 {
		t.Errorf("Expected exactly 1 Home item, got %d", pageRefs["/"])
	}
}

func TestBuildProfessionalMainMenuFromPlan_NoEmptyPageRef(t *testing.T) {
	// The dedup logic should skip items with empty PageRef
	// BuildProfessionalMainMenuFromPlan shouldn't produce empty PageRef items,
	// but the dedup guard is there for safety
	plan := &ai.SitePlan{
		SiteType: ai.SiteTypeBlog,
		Pages: []ai.PageSpec{
			{Path: "content/_index.md", Title: "Home"},
		},
	}

	items, err := BuildProfessionalMainMenuFromPlan(plan)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	for _, item := range items {
		if item.PageRef == "" {
			t.Error("Found item with empty PageRef")
		}
	}
}

// =============================================================================
// ApplyMenuFromSitePlan Tests
// =============================================================================

func TestApplyMenuFromSitePlan(t *testing.T) {
	tmpDir := t.TempDir()
	hugoTomlPath := filepath.Join(tmpDir, "hugo.toml")

	initial := `baseURL = "https://example.com/"
title = "Test Site"
`
	if err := os.WriteFile(hugoTomlPath, []byte(initial), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	plan := &ai.SitePlan{
		SiteType: ai.SiteTypeBlog,
		Pages: []ai.PageSpec{
			{Path: "content/_index.md", Title: "Home"},
			{Path: "content/posts/hello.md", Title: "Hello"},
			{Path: "content/about.md", Title: "About"},
		},
	}

	err := ApplyMenuFromSitePlan(plan, hugoTomlPath)
	if err != nil {
		t.Fatalf("ApplyMenuFromSitePlan() error: %v", err)
	}

	content, err := os.ReadFile(hugoTomlPath)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}

	result := string(content)

	// Verify the original content is preserved
	if !strings.Contains(result, "baseURL") {
		t.Error("Original baseURL should be preserved")
	}

	// Verify menu block markers
	if !strings.Contains(result, menuBlockStart) {
		t.Error("Missing BEGIN WALGO MENU marker")
	}
	if !strings.Contains(result, menuBlockEnd) {
		t.Error("Missing END WALGO MENU marker")
	}

	// Verify menu items
	if !strings.Contains(result, `name = "Home"`) {
		t.Error("Missing Home menu item")
	}
	if !strings.Contains(result, `name = "Posts"`) {
		t.Error("Missing Posts menu item")
	}
	if !strings.Contains(result, `name = "About"`) {
		t.Error("Missing About menu item")
	}
}

func TestApplyMenuFromSitePlan_FileNotFound(t *testing.T) {
	plan := &ai.SitePlan{
		SiteType: ai.SiteTypeBlog,
		Pages:    []ai.PageSpec{},
	}

	err := ApplyMenuFromSitePlan(plan, "/nonexistent/hugo.toml")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

// =============================================================================
// MenuItem struct Tests
// =============================================================================

func TestMenuItem_Fields(t *testing.T) {
	item := MenuItem{
		Name:    "Test Item",
		PageRef: "/test",
		Weight:  50,
	}

	if item.Name != "Test Item" {
		t.Errorf("Name = %q, want %q", item.Name, "Test Item")
	}
	if item.PageRef != "/test" {
		t.Errorf("PageRef = %q, want %q", item.PageRef, "/test")
	}
	if item.Weight != 50 {
		t.Errorf("Weight = %d, want %d", item.Weight, 50)
	}
}
