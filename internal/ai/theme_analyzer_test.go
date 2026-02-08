package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// extractFrontmatterFields tests
// =============================================================================

func TestExtractFrontmatterFields(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter string
		expected    []string
	}{
		{
			name:        "simple yaml fields",
			frontmatter: "title: Hello\ndate: 2024-01-01\ndraft: false\n",
			expected:    []string{"title", "date", "draft"},
		},
		{
			name:        "fields with empty values",
			frontmatter: "title: \"\"\ndescription: \"\"\n",
			expected:    []string{"title", "description"},
		},
		{
			name:        "skips comments",
			frontmatter: "# This is a comment\ntitle: Hello\n# Another comment\ndate: 2024-01-01\n",
			expected:    []string{"title", "date"},
		},
		{
			name:        "skips empty lines",
			frontmatter: "\ntitle: Hello\n\ndate: 2024-01-01\n\n",
			expected:    []string{"title", "date"},
		},
		{
			name:        "empty frontmatter",
			frontmatter: "",
			expected:    []string{},
		},
		{
			name:        "only comments and blanks",
			frontmatter: "# comment\n\n# another\n",
			expected:    []string{},
		},
		{
			name:        "skips curly brace fields",
			frontmatter: "title: Hello\n{bad}: value\ndate: 2024-01-01\n",
			expected:    []string{"title", "date"},
		},
		{
			name:        "fields with template values",
			frontmatter: "title: \"{{ replace .Name \"-\" \" \" | title }}\"\ndate: {{ .Date }}\n",
			expected:    []string{"title", "date"},
		},
		{
			name:        "nested field ignored (colon at position 0)",
			frontmatter: ":invalid\ntitle: OK\n",
			expected:    []string{"title"},
		},
		{
			name:        "field with complex value",
			frontmatter: "tags: [\"go\", \"hugo\"]\ncategories: [\"tech\"]\n",
			expected:    []string{"tags", "categories"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFrontmatterFields(tt.frontmatter)
			if len(result) != len(tt.expected) {
				t.Errorf("extractFrontmatterFields() returned %d fields, want %d\nGot: %v\nWant: %v",
					len(result), len(tt.expected), result, tt.expected)
				return
			}
			for i, field := range result {
				if field != tt.expected[i] {
					t.Errorf("field[%d] = %q, want %q", i, field, tt.expected[i])
				}
			}
		})
	}
}

// =============================================================================
// containsString tests
// =============================================================================

func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "found in slice",
			slice:    []string{"a", "b", "c"},
			value:    "b",
			expected: true,
		},
		{
			name:     "not found in slice",
			slice:    []string{"a", "b", "c"},
			value:    "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			value:    "a",
			expected: false,
		},
		{
			name:     "nil slice",
			slice:    nil,
			value:    "a",
			expected: false,
		},
		{
			name:     "case sensitive",
			slice:    []string{"Hello"},
			value:    "hello",
			expected: false,
		},
		{
			name:     "empty string in slice",
			slice:    []string{"", "a"},
			value:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.slice, tt.value)
			if result != tt.expected {
				t.Errorf("containsString(%v, %q) = %v, want %v", tt.slice, tt.value, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// containsField tests
// =============================================================================

func TestContainsField(t *testing.T) {
	tests := []struct {
		name     string
		fields   []string
		field    string
		expected bool
	}{
		{
			name:     "exact match",
			fields:   []string{"title", "date", "draft"},
			field:    "date",
			expected: true,
		},
		{
			name:     "case insensitive match",
			fields:   []string{"Title", "Date", "Draft"},
			field:    "title",
			expected: true,
		},
		{
			name:     "not found",
			fields:   []string{"title", "date"},
			field:    "tags",
			expected: false,
		},
		{
			name:     "empty fields",
			fields:   []string{},
			field:    "title",
			expected: false,
		},
		{
			name:     "nil fields",
			fields:   nil,
			field:    "title",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsField(tt.fields, tt.field)
			if result != tt.expected {
				t.Errorf("containsField(%v, %q) = %v, want %v", tt.fields, tt.field, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// GetPageBundleType tests
// =============================================================================

func TestGetPageBundleType(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "branch bundle at root",
			path:     "_index.md",
			expected: "branch",
		},
		{
			name:     "branch bundle in content",
			path:     "content/_index.md",
			expected: "branch",
		},
		{
			name:     "branch bundle in subdirectory",
			path:     "posts/_index.md",
			expected: "branch",
		},
		{
			name:     "branch bundle deeply nested",
			path:     "docs/guides/_index.md",
			expected: "branch",
		},
		{
			name:     "leaf bundle at root",
			path:     "index.md",
			expected: "leaf",
		},
		{
			name:     "leaf bundle in subdirectory",
			path:     "posts/my-post/index.md",
			expected: "leaf",
		},
		{
			name:     "single page",
			path:     "about.md",
			expected: "single",
		},
		{
			name:     "single page in subdirectory",
			path:     "posts/my-post.md",
			expected: "single",
		},
		{
			name:     "file named similar but not index",
			path:     "posts/my-index.md",
			expected: "single",
		},
		{
			name:     "file named similar but not _index",
			path:     "posts/my_index.md",
			expected: "single",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPageBundleType(tt.path)
			if result != tt.expected {
				t.Errorf("GetPageBundleType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// getFieldDefaultValue tests
// =============================================================================

func TestGetFieldDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{
			name:     "image field",
			field:    "featured_image",
			expected: `""`,
		},
		{
			name:     "img field",
			field:    "thumbnail_img",
			expected: `""`,
		},
		{
			name:     "url field",
			field:    "project_url",
			expected: `""`,
		},
		{
			name:     "link field",
			field:    "external_link",
			expected: `""`,
		},
		{
			name:     "tags field",
			field:    "tags",
			expected: "[]",
		},
		{
			name:     "categories field",
			field:    "categories",
			expected: "[]",
		},
		{
			name:     "weight field",
			field:    "weight",
			expected: "10",
		},
		{
			name:     "order field",
			field:    "sort_order",
			expected: "10",
		},
		{
			name:     "toc field",
			field:    "toc",
			expected: "true",
		},
		{
			name:     "bookToc field",
			field:    "bookToc",
			expected: "true",
		},
		{
			name:     "collapse field",
			field:    "bookCollapseSection",
			expected: "false",
		},
		{
			name:     "price field",
			field:    "price",
			expected: `""`,
		},
		{
			name:     "tech_stack field",
			field:    "tech_stack",
			expected: "[]",
		},
		{
			name:     "author field",
			field:    "author",
			expected: `""`,
		},
		{
			name:     "unknown field defaults to empty string",
			field:    "custom_field",
			expected: `""`,
		},
		{
			name:     "case insensitive IMAGE",
			field:    "IMAGE_URL",
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFieldDefaultValue(tt.field)
			if result != tt.expected {
				t.Errorf("getFieldDefaultValue(%q) = %q, want %q", tt.field, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// GetRecommendedFrontmatter tests
// =============================================================================

func TestGetRecommendedFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		analysis *ThemeAnalysis
		patterns *ContentPatterns
		section  string
		check    func(t *testing.T, fields []string)
	}{
		{
			name: "always includes title and draft",
			analysis: &ThemeAnalysis{
				FrontmatterFields: map[string][]string{},
			},
			patterns: nil,
			section:  "posts",
			check: func(t *testing.T, fields []string) {
				if len(fields) < 2 {
					t.Error("expected at least title and draft")
					return
				}
				if fields[0] != "title" {
					t.Errorf("first field should be 'title', got %q", fields[0])
				}
				if fields[1] != "draft" {
					t.Errorf("second field should be 'draft', got %q", fields[1])
				}
			},
		},
		{
			name: "uses section-specific fields from theme",
			analysis: &ThemeAnalysis{
				FrontmatterFields: map[string][]string{
					"posts": {"title", "date", "tags", "featured_image"},
				},
			},
			patterns: nil,
			section:  "posts",
			check: func(t *testing.T, fields []string) {
				if !containsString(fields, "title") {
					t.Error("should contain title")
				}
				if !containsString(fields, "draft") {
					t.Error("should contain draft")
				}
				if !containsString(fields, "date") {
					t.Error("should contain date from theme")
				}
				if !containsString(fields, "tags") {
					t.Error("should contain tags from theme")
				}
				if !containsString(fields, "featured_image") {
					t.Error("should contain featured_image from theme")
				}
			},
		},
		{
			name: "falls back to default archetype",
			analysis: &ThemeAnalysis{
				FrontmatterFields: map[string][]string{
					"default": {"title", "date", "description"},
				},
			},
			patterns: nil,
			section:  "posts",
			check: func(t *testing.T, fields []string) {
				if !containsString(fields, "date") {
					t.Error("should contain date from default archetype")
				}
				if !containsString(fields, "description") {
					t.Error("should contain description from default archetype")
				}
			},
		},
		{
			name: "uses learned patterns when no theme archetypes",
			analysis: &ThemeAnalysis{
				FrontmatterFields: map[string][]string{},
			},
			patterns: &ContentPatterns{
				SectionPatterns: map[string]*SectionPattern{
					"posts": {
						Name:         "posts",
						CommonFields: []string{"author", "category"},
					},
				},
			},
			section: "posts",
			check: func(t *testing.T, fields []string) {
				if !containsString(fields, "author") {
					t.Error("should contain author from patterns")
				}
				if !containsString(fields, "category") {
					t.Error("should contain category from patterns")
				}
			},
		},
		{
			name: "does not duplicate title and draft",
			analysis: &ThemeAnalysis{
				FrontmatterFields: map[string][]string{
					"posts": {"title", "draft", "date"},
				},
			},
			patterns: nil,
			section:  "posts",
			check: func(t *testing.T, fields []string) {
				titleCount := 0
				draftCount := 0
				for _, f := range fields {
					if f == "title" {
						titleCount++
					}
					if f == "draft" {
						draftCount++
					}
				}
				if titleCount != 1 {
					t.Errorf("title appears %d times, want 1", titleCount)
				}
				if draftCount != 1 {
					t.Errorf("draft appears %d times, want 1", draftCount)
				}
			},
		},
		{
			name: "nil patterns handled gracefully",
			analysis: &ThemeAnalysis{
				FrontmatterFields: map[string][]string{},
			},
			patterns: nil,
			section:  "unknown",
			check: func(t *testing.T, fields []string) {
				if len(fields) < 2 {
					t.Error("should have at least title and draft")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRecommendedFrontmatter(tt.analysis, tt.patterns, tt.section)
			tt.check(t, result)
		})
	}
}

// =============================================================================
// writeParamValue tests
// =============================================================================

func TestWriteParamValue(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected string
	}{
		{
			name:     "string value",
			key:      "description",
			value:    "My site",
			expected: "  description = \"My site\"\n",
		},
		{
			name:     "bool value true",
			key:      "enableSearch",
			value:    true,
			expected: "  enableSearch = true\n",
		},
		{
			name:     "bool value false",
			key:      "enableSearch",
			value:    false,
			expected: "  enableSearch = false\n",
		},
		{
			name:     "int value",
			key:      "maxPosts",
			value:    10,
			expected: "  maxPosts = 10\n",
		},
		{
			name:     "float64 value",
			key:      "ratio",
			value:    3.14,
			expected: "  ratio = 3.14\n",
		},
		{
			name:     "string array",
			key:      "languages",
			value:    []interface{}{"en", "fr", "de"},
			expected: "  languages = [\"en\", \"fr\", \"de\"]\n",
		},
		{
			name:     "mixed array",
			key:      "items",
			value:    []interface{}{"hello", 42},
			expected: "  items = [\"hello\", 42]\n",
		},
		{
			name:     "empty array",
			key:      "tags",
			value:    []interface{}{},
			expected: "  tags = []\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			writeParamValue(&sb, tt.key, tt.value)
			result := sb.String()
			if result != tt.expected {
				t.Errorf("writeParamValue(%q, %v):\nGot:  %q\nWant: %q", tt.key, tt.value, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// BuildThemeContextFromAnalysis tests
// =============================================================================

func TestBuildThemeContextFromAnalysis(t *testing.T) {
	t.Run("empty theme name returns empty", func(t *testing.T) {
		result := BuildThemeContextFromAnalysis("", &ThemeAnalysis{}, &ThemeConfigAnalysis{}, &ContentPatterns{})
		// Even with empty theme name, BuildThemeContextFromAnalysis still builds a context
		if !strings.Contains(result, "=== DYNAMIC THEME ANALYSIS ===") {
			t.Error("expected header in output")
		}
	})

	t.Run("full analysis", func(t *testing.T) {
		themeAnalysis := &ThemeAnalysis{
			Name:        "ananke",
			Description: "A fast theme",
			Sections:    []string{"posts", "pages"},
			FrontmatterFields: map[string][]string{
				"posts": {"title", "date", "tags"},
			},
			HasTaxonomies: true,
			HasSearch:     true,
		}

		configAnalysis := &ThemeConfigAnalysis{
			Theme:          "ananke",
			RequiredParams: []string{"description"},
			OptionalParams: []string{"logo"},
			PageParams:     []string{"featured_image"},
			Taxonomies: map[string]string{
				"tag": "tags",
			},
			Menus: []MenuConfig{
				{
					Name: "main",
					Items: []MenuItem{
						{Name: "Home", URL: "/"},
					},
				},
			},
			RecommendedParams: map[string]interface{}{
				"description": "My site",
				"showDate":    true,
				"maxItems":    10,
			},
		}

		contentPatterns := &ContentPatterns{
			SectionPatterns: map[string]*SectionPattern{
				"posts": {
					Name:         "posts",
					CommonFields: []string{"title", "date", "author"},
					UsesBundle:   true,
				},
			},
		}

		result := BuildThemeContextFromAnalysis("ananke", themeAnalysis, configAnalysis, contentPatterns)

		checks := []string{
			"THEME: ananke",
			"Description: A fast theme",
			"SUPPORTED SECTIONS",
			"posts/",
			"pages/",
			"FRONTMATTER BY SECTION",
			"SITE PARAMS",
			"Required: description",
			"Optional: logo",
			"PAGE PARAMS",
			"featured_image",
			"TAXONOMIES: Supported",
			"SEARCH: Supported",
			"MENU STRUCTURE",
			"menu.main",
			"Home -> /",
			"RECOMMENDED PARAMS",
			"LEARNED FROM EXISTING CONTENT",
			"[page bundles]",
		}

		for _, check := range checks {
			if !strings.Contains(result, check) {
				t.Errorf("expected %q in context output\nGot:\n%s", check, result)
			}
		}
	})

	t.Run("no sections or params", func(t *testing.T) {
		themeAnalysis := &ThemeAnalysis{
			Name:              "minimal",
			Sections:          []string{},
			FrontmatterFields: map[string][]string{},
		}
		configAnalysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
		}
		contentPatterns := &ContentPatterns{
			SectionPatterns: map[string]*SectionPattern{},
		}

		result := BuildThemeContextFromAnalysis("minimal", themeAnalysis, configAnalysis, contentPatterns)

		if !strings.Contains(result, "THEME: minimal") {
			t.Error("expected theme name in output")
		}
		if strings.Contains(result, "SUPPORTED SECTIONS") {
			t.Error("should not have sections block when empty")
		}
		if strings.Contains(result, "TAXONOMIES") {
			t.Error("should not have taxonomies block when disabled")
		}
		if strings.Contains(result, "SEARCH") {
			t.Error("should not have search block when disabled")
		}
	})

	t.Run("recommended params filtering", func(t *testing.T) {
		themeAnalysis := &ThemeAnalysis{
			Name:              "test",
			Sections:          []string{},
			FrontmatterFields: map[string][]string{},
		}
		configAnalysis := &ThemeConfigAnalysis{
			RecommendedParams: map[string]interface{}{
				"short":    "ok",
				"longstr":  strings.Repeat("x", 100), // >60 chars, should be excluded
				"boolval":  true,
				"intval":   42,
				"floatval": 3.14,
			},
		}
		contentPatterns := &ContentPatterns{
			SectionPatterns: map[string]*SectionPattern{},
		}

		result := BuildThemeContextFromAnalysis("test", themeAnalysis, configAnalysis, contentPatterns)

		if !strings.Contains(result, "short") {
			t.Error("expected short string param")
		}
		if strings.Contains(result, strings.Repeat("x", 100)) {
			t.Error("should not include long string params")
		}
		if !strings.Contains(result, "boolval: true") {
			t.Error("expected bool param")
		}
		if !strings.Contains(result, "intval: 42") {
			t.Error("expected int param")
		}
	})
}

// =============================================================================
// generateArchetypeContent tests
// =============================================================================

func TestGenerateArchetypeContent(t *testing.T) {
	t.Run("minimal archetype", func(t *testing.T) {
		config := ArchetypeConfig{
			Section:      "posts",
			Fields:       []string{},
			CustomFields: make(map[string]string),
		}

		result := generateArchetypeContent(config)

		if !strings.HasPrefix(result, "---\n") {
			t.Error("should start with frontmatter delimiter")
		}
		if !strings.HasSuffix(result, "---\n") {
			t.Error("should end with frontmatter delimiter")
		}
		if !strings.Contains(result, `title: "{{ replace .Name "-" " " | title }}"`) {
			t.Error("should contain title template")
		}
		if !strings.Contains(result, "draft: false") {
			t.Error("should contain draft: false")
		}
		if !strings.Contains(result, `description: ""`) {
			t.Error("should contain description")
		}
	})

	t.Run("with date", func(t *testing.T) {
		config := ArchetypeConfig{
			Section:      "posts",
			Fields:       []string{},
			IncludeDate:  true,
			CustomFields: make(map[string]string),
		}

		result := generateArchetypeContent(config)

		if !strings.Contains(result, "date: {{ .Date }}") {
			t.Error("should contain date template")
		}
	})

	t.Run("with weight", func(t *testing.T) {
		config := ArchetypeConfig{
			Section:       "docs",
			Fields:        []string{},
			IncludeWeight: true,
			CustomFields:  make(map[string]string),
		}

		result := generateArchetypeContent(config)

		if !strings.Contains(result, "weight: 10") {
			t.Error("should contain weight")
		}
	})

	t.Run("with tags", func(t *testing.T) {
		config := ArchetypeConfig{
			Section:      "posts",
			Fields:       []string{},
			IncludeTags:  true,
			CustomFields: make(map[string]string),
		}

		result := generateArchetypeContent(config)

		if !strings.Contains(result, "tags: []") {
			t.Error("should contain tags")
		}
	})

	t.Run("with extra fields", func(t *testing.T) {
		config := ArchetypeConfig{
			Section:      "posts",
			Fields:       []string{"featured_image", "author"},
			CustomFields: make(map[string]string),
		}

		result := generateArchetypeContent(config)

		if !strings.Contains(result, "featured_image:") {
			t.Error("should contain featured_image")
		}
		if !strings.Contains(result, "author:") {
			t.Error("should contain author")
		}
	})

	t.Run("does not duplicate standard fields", func(t *testing.T) {
		config := ArchetypeConfig{
			Section:      "posts",
			Fields:       []string{"title", "draft", "description"},
			IncludeDate:  true,
			CustomFields: make(map[string]string),
		}

		result := generateArchetypeContent(config)

		titleCount := strings.Count(result, "title:")
		if titleCount != 1 {
			t.Errorf("title appears %d times, want 1", titleCount)
		}
		draftCount := strings.Count(result, "draft:")
		if draftCount != 1 {
			t.Errorf("draft appears %d times, want 1", draftCount)
		}
	})

	t.Run("with custom fields", func(t *testing.T) {
		config := ArchetypeConfig{
			Section: "products",
			Fields:  []string{},
			CustomFields: map[string]string{
				"sku":      `""`,
				"in_stock": "true",
			},
		}

		result := generateArchetypeContent(config)

		if !strings.Contains(result, "sku:") {
			t.Error("should contain custom field sku")
		}
		if !strings.Contains(result, "in_stock:") {
			t.Error("should contain custom field in_stock")
		}
	})
}

// =============================================================================
// generateDefaultArchetype tests
// =============================================================================

func TestGenerateDefaultArchetype(t *testing.T) {
	t.Run("without featured_image param", func(t *testing.T) {
		theme := &ThemeAnalysis{
			FrontmatterFields: map[string][]string{},
		}
		config := &ThemeConfigAnalysis{
			PageParams: []string{},
		}

		result := generateDefaultArchetype(theme, config)

		if !strings.HasPrefix(result, "---\n") {
			t.Error("should start with frontmatter delimiter")
		}
		if !strings.Contains(result, `title: "{{ replace .Name "-" " " | title }}"`) {
			t.Error("should contain title template")
		}
		if !strings.Contains(result, "date: {{ .Date }}") {
			t.Error("should contain date")
		}
		if !strings.Contains(result, "draft: false") {
			t.Error("should contain draft: false")
		}
		if strings.Contains(result, "featured_image") {
			t.Error("should not contain featured_image when not in page params")
		}
	})

	t.Run("with featured_image param", func(t *testing.T) {
		theme := &ThemeAnalysis{
			FrontmatterFields: map[string][]string{},
		}
		config := &ThemeConfigAnalysis{
			PageParams: []string{"featured_image"},
		}

		result := generateDefaultArchetype(theme, config)

		if !strings.Contains(result, `featured_image: ""`) {
			t.Error("should contain featured_image when in page params")
		}
	})

	t.Run("with image param", func(t *testing.T) {
		theme := &ThemeAnalysis{
			FrontmatterFields: map[string][]string{},
		}
		config := &ThemeConfigAnalysis{
			PageParams: []string{"image"},
		}

		result := generateDefaultArchetype(theme, config)

		if !strings.Contains(result, `featured_image: ""`) {
			t.Error("should contain featured_image when 'image' is in page params")
		}
	})
}

// =============================================================================
// buildArchetypeConfigs tests
// =============================================================================

func TestBuildArchetypeConfigs(t *testing.T) {
	t.Run("uses theme frontmatter fields", func(t *testing.T) {
		theme := &ThemeAnalysis{
			FrontmatterFields: map[string][]string{
				"posts":   {"title", "date", "tags"},
				"default": {"title", "draft"},
			},
			Sections: []string{"posts"},
		}
		config := &ThemeConfigAnalysis{
			PageParams: []string{},
		}

		configs := buildArchetypeConfigs(theme, config)

		if len(configs) != 1 {
			t.Fatalf("expected 1 config (default skipped), got %d", len(configs))
		}
		if configs[0].Section != "posts" {
			t.Errorf("expected section 'posts', got %q", configs[0].Section)
		}
		if !configs[0].IncludeDate {
			t.Error("expected IncludeDate to be true for posts with 'date' field")
		}
		if !configs[0].IncludeTags {
			t.Error("expected IncludeTags to be true for posts with 'tags' field")
		}
	})

	t.Run("infers from sections when no archetypes", func(t *testing.T) {
		theme := &ThemeAnalysis{
			FrontmatterFields: map[string][]string{},
			Sections:          []string{"posts", "docs"},
		}
		config := &ThemeConfigAnalysis{
			PageParams: []string{},
		}

		configs := buildArchetypeConfigs(theme, config)

		if len(configs) != 2 {
			t.Fatalf("expected 2 inferred configs, got %d", len(configs))
		}
	})

	t.Run("empty theme returns empty configs", func(t *testing.T) {
		theme := &ThemeAnalysis{
			FrontmatterFields: map[string][]string{},
			Sections:          []string{},
		}
		config := &ThemeConfigAnalysis{
			PageParams: []string{},
		}

		configs := buildArchetypeConfigs(theme, config)

		if len(configs) != 0 {
			t.Errorf("expected 0 configs, got %d", len(configs))
		}
	})
}

// =============================================================================
// inferArchetypeConfig tests
// =============================================================================

func TestInferArchetypeConfig(t *testing.T) {
	tests := []struct {
		name       string
		section    string
		pageParams []string
		check      func(t *testing.T, cfg ArchetypeConfig)
	}{
		{
			name:       "blog section gets date and tags",
			section:    "blog",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !cfg.IncludeDate {
					t.Error("blog should include date")
				}
				if !cfg.IncludeTags {
					t.Error("blog should include tags")
				}
			},
		},
		{
			name:       "posts section gets date and tags",
			section:    "posts",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !cfg.IncludeDate {
					t.Error("posts should include date")
				}
				if !cfg.IncludeTags {
					t.Error("posts should include tags")
				}
			},
		},
		{
			name:       "articles section gets date and tags",
			section:    "articles",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !cfg.IncludeDate {
					t.Error("articles should include date")
				}
				if !cfg.IncludeTags {
					t.Error("articles should include tags")
				}
			},
		},
		{
			name:       "docs section gets weight",
			section:    "docs",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !cfg.IncludeWeight {
					t.Error("docs should include weight")
				}
				if cfg.IncludeDate {
					t.Error("docs should not include date")
				}
			},
		},
		{
			name:       "guides section gets weight",
			section:    "guides",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !cfg.IncludeWeight {
					t.Error("guides should include weight")
				}
			},
		},
		{
			name:       "tutorial section gets weight",
			section:    "tutorials",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !cfg.IncludeWeight {
					t.Error("tutorials should include weight")
				}
			},
		},
		{
			name:       "projects section gets date",
			section:    "projects",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !cfg.IncludeDate {
					t.Error("projects should include date")
				}
			},
		},
		{
			name:       "portfolio section gets date",
			section:    "portfolio",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !cfg.IncludeDate {
					t.Error("portfolio should include date")
				}
			},
		},
		{
			name:       "services section gets date",
			section:    "services",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !cfg.IncludeDate {
					t.Error("services should include date")
				}
			},
		},
		{
			name:       "pages section minimal fields",
			section:    "pages",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if cfg.IncludeDate {
					t.Error("pages should not include date")
				}
				if cfg.IncludeTags {
					t.Error("pages should not include tags")
				}
				if cfg.IncludeWeight {
					t.Error("pages should not include weight")
				}
			},
		},
		{
			name:       "blog with featured_image page param",
			section:    "blog",
			pageParams: []string{"featured_image"},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !containsString(cfg.Fields, "featured_image") {
					t.Error("blog with featured_image page param should include featured_image field")
				}
			},
		},
		{
			name:       "docs with bookToc page param",
			section:    "docs",
			pageParams: []string{"bookToc", "bookCollapseSection"},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !containsString(cfg.Fields, "bookToc") {
					t.Error("docs with bookToc param should include bookToc field")
				}
			},
		},
		{
			name:       "projects with tech_stack param",
			section:    "projects",
			pageParams: []string{"tech_stack"},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !containsString(cfg.Fields, "tech_stack") {
					t.Error("projects with tech_stack param should include tech_stack field")
				}
			},
		},
		{
			name:       "services with price param",
			section:    "services",
			pageParams: []string{"price"},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if !containsString(cfg.Fields, "price") {
					t.Error("services with price param should include price field")
				}
			},
		},
		{
			name:       "unknown section gets defaults",
			section:    "foobar",
			pageParams: []string{},
			check: func(t *testing.T, cfg ArchetypeConfig) {
				if cfg.Section != "foobar" {
					t.Errorf("expected section 'foobar', got %q", cfg.Section)
				}
				if !containsString(cfg.Fields, "title") {
					t.Error("unknown section should have title")
				}
				if !containsString(cfg.Fields, "description") {
					t.Error("unknown section should have description")
				}
				if !containsString(cfg.Fields, "draft") {
					t.Error("unknown section should have draft")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ThemeConfigAnalysis{
				PageParams: tt.pageParams,
			}
			result := inferArchetypeConfig(tt.section, config)
			tt.check(t, result)
		})
	}
}

// =============================================================================
// parseTomlManually tests
// =============================================================================

func TestParseTomlManually(t *testing.T) {
	t.Run("extracts params", func(t *testing.T) {
		config := `
baseURL = "https://example.com"
title = "My Site"

[params]
description = "A great site"
author = "John"
enableSearch = true

[menu]
[[menu.main]]
name = "Home"
`
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
		}

		parseTomlManually(config, analysis)

		if v, ok := analysis.RecommendedParams["description"]; !ok || v != "A great site" {
			t.Errorf("expected description='A great site', got %v", v)
		}
		if v, ok := analysis.RecommendedParams["author"]; !ok || v != "John" {
			t.Errorf("expected author='John', got %v", v)
		}
		if v, ok := analysis.RecommendedParams["enableSearch"]; !ok || v != "true" {
			t.Errorf("expected enableSearch='true', got %v", v)
		}
	})

	t.Run("stops at next section", func(t *testing.T) {
		config := `
[params]
key1 = "value1"

[other]
key2 = "value2"
`
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
		}

		parseTomlManually(config, analysis)

		if _, ok := analysis.RecommendedParams["key1"]; !ok {
			t.Error("expected key1 in params")
		}
		if _, ok := analysis.RecommendedParams["key2"]; ok {
			t.Error("should not include key2 from [other] section")
		}
	})

	t.Run("handles Params case variant", func(t *testing.T) {
		config := `
[Params]
myKey = "myValue"
`
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
		}

		parseTomlManually(config, analysis)

		if v, ok := analysis.RecommendedParams["myKey"]; !ok || v != "myValue" {
			t.Errorf("expected myKey='myValue', got %v", v)
		}
	})

	t.Run("skips comments and empty lines", func(t *testing.T) {
		config := `
[params]
# This is a comment
key1 = "value1"

# Another comment
key2 = "value2"
`
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
		}

		parseTomlManually(config, analysis)

		if len(analysis.RecommendedParams) != 2 {
			t.Errorf("expected 2 params, got %d", len(analysis.RecommendedParams))
		}
	})

	t.Run("empty config", func(t *testing.T) {
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
		}

		parseTomlManually("", analysis)

		if len(analysis.RecommendedParams) != 0 {
			t.Errorf("expected 0 params, got %d", len(analysis.RecommendedParams))
		}
	})
}

// =============================================================================
// extractConfigData tests
// =============================================================================

func TestExtractConfigData(t *testing.T) {
	t.Run("extracts params", func(t *testing.T) {
		data := map[string]interface{}{
			"params": map[string]interface{}{
				"description": "My site",
				"author":      "John",
			},
		}
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
			Taxonomies:        make(map[string]string),
			Outputs:           make(map[string][]string),
		}

		extractConfigData(data, analysis)

		if v, ok := analysis.RecommendedParams["description"]; !ok || v != "My site" {
			t.Errorf("expected description param, got %v", v)
		}
		if v, ok := analysis.RecommendedParams["author"]; !ok || v != "John" {
			t.Errorf("expected author param, got %v", v)
		}
	})

	t.Run("extracts taxonomies", func(t *testing.T) {
		data := map[string]interface{}{
			"taxonomies": map[string]interface{}{
				"tag":      "tags",
				"category": "categories",
			},
		}
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
			Taxonomies:        make(map[string]string),
			Outputs:           make(map[string][]string),
		}

		extractConfigData(data, analysis)

		if v, ok := analysis.Taxonomies["tag"]; !ok || v != "tags" {
			t.Errorf("expected tag taxonomy, got %v", v)
		}
		if v, ok := analysis.Taxonomies["category"]; !ok || v != "categories" {
			t.Errorf("expected category taxonomy, got %v", v)
		}
	})

	t.Run("extracts menus", func(t *testing.T) {
		data := map[string]interface{}{
			"menu": map[string]interface{}{
				"main": []interface{}{
					map[string]interface{}{
						"name": "Home",
						"url":  "/",
					},
				},
			},
		}
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
			Taxonomies:        make(map[string]string),
			Outputs:           make(map[string][]string),
		}

		extractConfigData(data, analysis)

		if len(analysis.Menus) != 1 {
			t.Fatalf("expected 1 menu, got %d", len(analysis.Menus))
		}
		if analysis.Menus[0].Name != "main" {
			t.Errorf("expected menu name 'main', got %q", analysis.Menus[0].Name)
		}
		if len(analysis.Menus[0].Items) != 1 {
			t.Fatalf("expected 1 menu item, got %d", len(analysis.Menus[0].Items))
		}
		if analysis.Menus[0].Items[0].Name != "Home" {
			t.Errorf("expected menu item name 'Home', got %q", analysis.Menus[0].Items[0].Name)
		}
	})

	t.Run("extracts outputs", func(t *testing.T) {
		data := map[string]interface{}{
			"outputs": map[string]interface{}{
				"home": []interface{}{"HTML", "RSS", "JSON"},
			},
		}
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
			Taxonomies:        make(map[string]string),
			Outputs:           make(map[string][]string),
		}

		extractConfigData(data, analysis)

		if formats, ok := analysis.Outputs["home"]; !ok {
			t.Error("expected home outputs")
		} else if len(formats) != 3 {
			t.Errorf("expected 3 output formats, got %d", len(formats))
		}
	})

	t.Run("handles missing sections gracefully", func(t *testing.T) {
		data := map[string]interface{}{}
		analysis := &ThemeConfigAnalysis{
			RecommendedParams: make(map[string]interface{}),
			Taxonomies:        make(map[string]string),
			Outputs:           make(map[string][]string),
		}

		extractConfigData(data, analysis)

		if len(analysis.RecommendedParams) != 0 {
			t.Error("should have no params")
		}
		if len(analysis.Taxonomies) != 0 {
			t.Error("should have no taxonomies")
		}
		if len(analysis.Menus) != 0 {
			t.Error("should have no menus")
		}
	})
}

// =============================================================================
// mergeParams tests
// =============================================================================

func TestMergeParams(t *testing.T) {
	t.Run("merges into empty dest", func(t *testing.T) {
		dest := make(map[string]interface{})
		src := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		mergeParams(dest, src)

		if len(dest) != 2 {
			t.Errorf("expected 2 params, got %d", len(dest))
		}
		if dest["key1"] != "value1" {
			t.Errorf("expected key1='value1', got %v", dest["key1"])
		}
	})

	t.Run("overwrites existing keys", func(t *testing.T) {
		dest := map[string]interface{}{
			"key1": "old",
		}
		src := map[string]interface{}{
			"key1": "new",
		}

		mergeParams(dest, src)

		if dest["key1"] != "new" {
			t.Errorf("expected key1='new', got %v", dest["key1"])
		}
	})

	t.Run("preserves existing non-overlapping keys", func(t *testing.T) {
		dest := map[string]interface{}{
			"existing": "keep",
		}
		src := map[string]interface{}{
			"new": "added",
		}

		mergeParams(dest, src)

		if dest["existing"] != "keep" {
			t.Errorf("expected existing='keep', got %v", dest["existing"])
		}
		if dest["new"] != "added" {
			t.Errorf("expected new='added', got %v", dest["new"])
		}
	})
}

// =============================================================================
// writeThemeParams tests
// =============================================================================

func TestWriteThemeParams(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		var sb strings.Builder
		params := map[string]interface{}{
			"key": "value",
		}
		writeThemeParams(&sb, params, "  ")
		if !strings.Contains(sb.String(), `  key = "value"`) {
			t.Errorf("expected string param, got %q", sb.String())
		}
	})

	t.Run("bool value", func(t *testing.T) {
		var sb strings.Builder
		params := map[string]interface{}{
			"enabled": true,
		}
		writeThemeParams(&sb, params, "  ")
		if !strings.Contains(sb.String(), "  enabled = true") {
			t.Errorf("expected bool param, got %q", sb.String())
		}
	})

	t.Run("int value", func(t *testing.T) {
		var sb strings.Builder
		params := map[string]interface{}{
			"count": 42,
		}
		writeThemeParams(&sb, params, "  ")
		if !strings.Contains(sb.String(), "  count = 42") {
			t.Errorf("expected int param, got %q", sb.String())
		}
	})

	t.Run("array value", func(t *testing.T) {
		var sb strings.Builder
		params := map[string]interface{}{
			"langs": []interface{}{"en", "fr"},
		}
		writeThemeParams(&sb, params, "  ")
		result := sb.String()
		if !strings.Contains(result, `  langs = ["en", "fr"]`) {
			t.Errorf("expected array param, got %q", result)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		var sb strings.Builder
		params := map[string]interface{}{
			"items": []interface{}{},
		}
		writeThemeParams(&sb, params, "  ")
		if !strings.Contains(sb.String(), "  items = []") {
			t.Errorf("expected empty array param, got %q", sb.String())
		}
	})

	t.Run("nested map creates section", func(t *testing.T) {
		var sb strings.Builder
		params := map[string]interface{}{
			"social": map[string]interface{}{
				"twitter": "@test",
			},
		}
		writeThemeParams(&sb, params, "  ")
		result := sb.String()
		if !strings.Contains(result, "[params.social]") {
			t.Errorf("expected nested section, got %q", result)
		}
	})
}

// =============================================================================
// writeNestedParams tests
// =============================================================================

func TestWriteNestedParams(t *testing.T) {
	t.Run("writes various types", func(t *testing.T) {
		var sb strings.Builder
		params := map[string]interface{}{
			"name":    "test",
			"enabled": true,
			"count":   5,
		}
		writeNestedParams(&sb, params, "    ")
		result := sb.String()

		if !strings.Contains(result, `    name = "test"`) {
			t.Errorf("expected string in nested params, got %q", result)
		}
		if !strings.Contains(result, "    enabled = true") {
			t.Errorf("expected bool in nested params, got %q", result)
		}
		if !strings.Contains(result, "    count = 5") {
			t.Errorf("expected int in nested params, got %q", result)
		}
	})

	t.Run("inline table for nested maps", func(t *testing.T) {
		var sb strings.Builder
		params := map[string]interface{}{
			"inner": map[string]interface{}{
				"key": "val",
			},
		}
		writeNestedParams(&sb, params, "  ")
		result := sb.String()

		if !strings.Contains(result, "inner = {") {
			t.Errorf("expected inline table, got %q", result)
		}
		if !strings.Contains(result, `key = "val"`) {
			t.Errorf("expected key=val in inline table, got %q", result)
		}
	})
}

// =============================================================================
// AnalyzeTheme tests (filesystem-based, using temp dirs)
// =============================================================================

func TestAnalyzeTheme(t *testing.T) {
	t.Run("empty theme name returns empty analysis", func(t *testing.T) {
		result := AnalyzeTheme("/nonexistent", "")
		if result.Name != "" {
			t.Errorf("expected empty name, got %q", result.Name)
		}
		if len(result.Sections) != 0 {
			t.Errorf("expected 0 sections, got %d", len(result.Sections))
		}
	})

	t.Run("nonexistent theme dir returns empty analysis", func(t *testing.T) {
		result := AnalyzeTheme("/nonexistent/path", "mytheme")
		if result.Name != "mytheme" {
			t.Errorf("expected name 'mytheme', got %q", result.Name)
		}
		if len(result.Sections) != 0 {
			t.Errorf("expected 0 sections, got %d", len(result.Sections))
		}
	})

	t.Run("full analysis with temp dir", func(t *testing.T) {
		siteDir := t.TempDir()
		themeDir := filepath.Join(siteDir, "themes", "test-theme")

		// Create layout sections
		os.MkdirAll(filepath.Join(themeDir, "layouts", "posts"), 0755)
		os.WriteFile(filepath.Join(themeDir, "layouts", "posts", "single.html"), []byte("<html>{{ .Content }}</html>"), 0644)
		os.MkdirAll(filepath.Join(themeDir, "layouts", "docs"), 0755)
		os.WriteFile(filepath.Join(themeDir, "layouts", "docs", "list.html"), []byte("<html>list</html>"), 0644)

		// Create ignored directories
		os.MkdirAll(filepath.Join(themeDir, "layouts", "_default"), 0755)
		os.MkdirAll(filepath.Join(themeDir, "layouts", "partials"), 0755)

		// Create taxonomy support
		os.MkdirAll(filepath.Join(themeDir, "layouts", "_default"), 0755)
		os.WriteFile(filepath.Join(themeDir, "layouts", "_default", "taxonomy.html"), []byte("taxonomy"), 0644)

		// Create search support
		os.MkdirAll(filepath.Join(themeDir, "layouts", "search"), 0755)

		// Create archetypes
		os.MkdirAll(filepath.Join(themeDir, "archetypes"), 0755)
		os.WriteFile(filepath.Join(themeDir, "archetypes", "posts.md"), []byte("---\ntitle: \"\"\ndate: \"\"\ntags: []\n---\n"), 0644)
		os.WriteFile(filepath.Join(themeDir, "archetypes", "default.md"), []byte("---\ntitle: \"\"\ndraft: true\n---\n"), 0644)

		// Create theme.toml
		os.WriteFile(filepath.Join(themeDir, "theme.toml"), []byte(`description = "A test theme"
min_version = "0.80.0"
`), 0644)

		result := AnalyzeTheme(siteDir, "test-theme")

		if result.Name != "test-theme" {
			t.Errorf("expected name 'test-theme', got %q", result.Name)
		}
		if result.Description != "A test theme" {
			t.Errorf("expected description 'A test theme', got %q", result.Description)
		}
		if result.MinHugoVersion != "0.80.0" {
			t.Errorf("expected min version '0.80.0', got %q", result.MinHugoVersion)
		}
		if !containsString(result.Sections, "posts") {
			t.Error("expected 'posts' in sections")
		}
		if !containsString(result.Sections, "docs") {
			t.Error("expected 'docs' in sections")
		}
		if containsString(result.Sections, "_default") {
			t.Error("should not include _default in sections")
		}
		if containsString(result.Sections, "partials") {
			t.Error("should not include partials in sections")
		}
		if !result.HasTaxonomies {
			t.Error("expected HasTaxonomies to be true")
		}
		if !result.HasSearch {
			t.Error("expected HasSearch to be true")
		}
		if _, ok := result.Archetypes["posts"]; !ok {
			t.Error("expected 'posts' archetype")
		}
		if fields, ok := result.FrontmatterFields["posts"]; ok {
			if !containsString(fields, "title") {
				t.Error("expected 'title' in posts frontmatter fields")
			}
			if !containsString(fields, "date") {
				t.Error("expected 'date' in posts frontmatter fields")
			}
			if !containsString(fields, "tags") {
				t.Error("expected 'tags' in posts frontmatter fields")
			}
		} else {
			t.Error("expected frontmatter fields for 'posts'")
		}
	})

	t.Run("site archetypes override theme archetypes", func(t *testing.T) {
		siteDir := t.TempDir()
		themeDir := filepath.Join(siteDir, "themes", "test-theme")

		// Create minimal theme
		os.MkdirAll(filepath.Join(themeDir, "layouts", "_default"), 0755)
		os.MkdirAll(filepath.Join(themeDir, "archetypes"), 0755)
		os.WriteFile(filepath.Join(themeDir, "archetypes", "posts.md"), []byte("---\ntitle: \"\"\ndate: \"\"\n---\n"), 0644)

		// Create site archetypes that override
		os.MkdirAll(filepath.Join(siteDir, "archetypes"), 0755)
		os.WriteFile(filepath.Join(siteDir, "archetypes", "posts.md"), []byte("---\ntitle: \"\"\ndate: \"\"\nauthor: \"\"\ncategory: \"\"\n---\n"), 0644)

		result := AnalyzeTheme(siteDir, "test-theme")

		// Site archetype should override theme archetype for "posts"
		if fields, ok := result.FrontmatterFields["posts"]; ok {
			if !containsString(fields, "author") {
				t.Error("expected 'author' from site archetype override")
			}
			if !containsString(fields, "category") {
				t.Error("expected 'category' from site archetype override")
			}
		} else {
			t.Error("expected frontmatter fields for 'posts'")
		}
	})
}

// =============================================================================
// AnalyzeThemeConfig tests (filesystem-based, using temp dirs)
// =============================================================================

func TestAnalyzeThemeConfig(t *testing.T) {
	t.Run("empty theme name returns empty analysis", func(t *testing.T) {
		result := AnalyzeThemeConfig("/nonexistent", "")
		if result.Theme != "" {
			t.Errorf("expected empty theme, got %q", result.Theme)
		}
	})

	t.Run("nonexistent theme dir returns empty analysis", func(t *testing.T) {
		result := AnalyzeThemeConfig("/nonexistent/path", "mytheme")
		if result.Theme != "mytheme" {
			t.Errorf("expected theme 'mytheme', got %q", result.Theme)
		}
	})

	t.Run("with example config", func(t *testing.T) {
		siteDir := t.TempDir()
		themeDir := filepath.Join(siteDir, "themes", "test-theme")

		// Create exampleSite config
		os.MkdirAll(filepath.Join(themeDir, "exampleSite"), 0755)
		os.WriteFile(filepath.Join(themeDir, "exampleSite", "hugo.toml"), []byte(`
baseURL = "https://example.com"
title = "Example"

[params]
description = "An example site"
author = "Test"
`), 0644)

		// Create minimal layout for param detection
		os.MkdirAll(filepath.Join(themeDir, "layouts", "_default"), 0755)
		os.WriteFile(filepath.Join(themeDir, "layouts", "_default", "baseof.html"), []byte(`
{{ .Site.Params.description }}
{{ if .Site.Params.logo }}logo{{ end }}
{{ .Params.featured_image }}
`), 0644)

		result := AnalyzeThemeConfig(siteDir, "test-theme")

		if result.ExampleConfig == "" {
			t.Error("expected example config to be loaded")
		}
		if len(result.RequiredParams) == 0 && len(result.OptionalParams) == 0 {
			t.Error("expected some params to be detected from layouts")
		}
	})

	t.Run("analyzes layouts for params", func(t *testing.T) {
		siteDir := t.TempDir()
		themeDir := filepath.Join(siteDir, "themes", "test-theme")

		os.MkdirAll(filepath.Join(themeDir, "layouts", "_default"), 0755)
		os.WriteFile(filepath.Join(themeDir, "layouts", "_default", "single.html"), []byte(`
{{ .Site.Params.siteName }}
{{ with .Site.Params.optionalLogo }}{{ . }}{{ end }}
{{ .Params.pageImage }}
`), 0644)

		result := AnalyzeThemeConfig(siteDir, "test-theme")

		if !containsString(result.RequiredParams, "sitename") {
			t.Errorf("expected 'sitename' in required params, got %v", result.RequiredParams)
		}
		if !containsString(result.OptionalParams, "optionallogo") {
			t.Errorf("expected 'optionallogo' in optional params, got %v", result.OptionalParams)
		}
		if !containsString(result.PageParams, "pageimage") {
			t.Errorf("expected 'pageimage' in page params, got %v", result.PageParams)
		}
	})
}

// =============================================================================
// AnalyzeSiteContent tests (filesystem-based, using temp dirs)
// =============================================================================

func TestAnalyzeSiteContent(t *testing.T) {
	t.Run("nonexistent content dir returns empty", func(t *testing.T) {
		result := AnalyzeSiteContent("/nonexistent/path")
		if len(result.SectionPatterns) != 0 {
			t.Errorf("expected 0 section patterns, got %d", len(result.SectionPatterns))
		}
	})

	t.Run("analyzes content sections", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		// Create posts section
		os.MkdirAll(filepath.Join(contentDir, "posts"), 0755)
		os.WriteFile(filepath.Join(contentDir, "posts", "post1.md"), []byte("---\ntitle: Post 1\ndate: 2024-01-01\nauthor: John\n---\nContent 1"), 0644)
		os.WriteFile(filepath.Join(contentDir, "posts", "post2.md"), []byte("---\ntitle: Post 2\ndate: 2024-01-02\nauthor: Jane\n---\nContent 2"), 0644)
		os.WriteFile(filepath.Join(contentDir, "posts", "post3.md"), []byte("---\ntitle: Post 3\ndate: 2024-01-03\n---\nContent 3"), 0644)

		result := AnalyzeSiteContent(siteDir)

		if sp, ok := result.SectionPatterns["posts"]; !ok {
			t.Error("expected 'posts' section pattern")
		} else {
			if sp.Name != "posts" {
				t.Errorf("expected section name 'posts', got %q", sp.Name)
			}
			if !containsString(sp.CommonFields, "title") {
				t.Error("expected 'title' in common fields")
			}
			if !containsString(sp.CommonFields, "date") {
				t.Error("expected 'date' in common fields")
			}
			if len(sp.ExampleFiles) != 3 {
				t.Errorf("expected 3 example files, got %d", len(sp.ExampleFiles))
			}
		}
	})

	t.Run("detects page bundles", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		// Create bundle content
		os.MkdirAll(filepath.Join(contentDir, "posts", "my-post"), 0755)
		os.WriteFile(filepath.Join(contentDir, "posts", "my-post", "index.md"), []byte("---\ntitle: My Post\n---\nBundle content"), 0644)

		result := AnalyzeSiteContent(siteDir)

		if sp, ok := result.SectionPatterns["posts"]; ok {
			if !sp.UsesBundle {
				t.Error("expected UsesBundle to be true for posts with index.md")
			}
		} else {
			t.Error("expected 'posts' section pattern")
		}
	})

	t.Run("limits example files to 3", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		os.MkdirAll(filepath.Join(contentDir, "posts"), 0755)
		for i := 0; i < 5; i++ {
			filename := filepath.Join(contentDir, "posts", strings.Replace("post-NUM.md", "NUM", string(rune('1'+i)), 1))
			os.WriteFile(filename, []byte("---\ntitle: Post\n---\n"), 0644)
		}

		result := AnalyzeSiteContent(siteDir)

		if sp, ok := result.SectionPatterns["posts"]; ok {
			if len(sp.ExampleFiles) > 3 {
				t.Errorf("expected max 3 example files, got %d", len(sp.ExampleFiles))
			}
		}
	})

	t.Run("ignores non-markdown files", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		os.MkdirAll(filepath.Join(contentDir, "posts"), 0755)
		os.WriteFile(filepath.Join(contentDir, "posts", "image.png"), []byte("fake image"), 0644)
		os.WriteFile(filepath.Join(contentDir, "posts", "post.md"), []byte("---\ntitle: Post\n---\n"), 0644)

		result := AnalyzeSiteContent(siteDir)

		if sp, ok := result.SectionPatterns["posts"]; ok {
			if len(sp.ExampleFiles) != 1 {
				t.Errorf("expected 1 example file (only .md), got %d", len(sp.ExampleFiles))
			}
		}
	})

	t.Run("ignores root-level files (no section)", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		os.MkdirAll(contentDir, 0755)
		os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte("---\ntitle: Home\n---\n"), 0644)

		result := AnalyzeSiteContent(siteDir)

		// Root-level files have empty section string, which is not stored
		if len(result.SectionPatterns) != 0 {
			t.Errorf("expected 0 section patterns for root-level files, got %d", len(result.SectionPatterns))
		}
	})
}

// =============================================================================
// readThemeMetadata tests (filesystem-based)
// =============================================================================

func TestReadThemeMetadata(t *testing.T) {
	t.Run("reads theme.toml", func(t *testing.T) {
		themeDir := t.TempDir()
		os.WriteFile(filepath.Join(themeDir, "theme.toml"), []byte(`
name = "My Theme"
description = "A wonderful theme"
min_version = "0.80.0"
`), 0644)

		desc, minVer := readThemeMetadata(themeDir)

		if desc != "A wonderful theme" {
			t.Errorf("expected description 'A wonderful theme', got %q", desc)
		}
		if minVer != "0.80.0" {
			t.Errorf("expected min_version '0.80.0', got %q", minVer)
		}
	})

	t.Run("reads theme.yaml", func(t *testing.T) {
		themeDir := t.TempDir()
		os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(`
description: "A YAML theme"
min_version: "0.90.0"
`), 0644)

		desc, minVer := readThemeMetadata(themeDir)

		if desc != "A YAML theme" {
			t.Errorf("expected description 'A YAML theme', got %q", desc)
		}
		if minVer != "0.90.0" {
			t.Errorf("expected min_version '0.90.0', got %q", minVer)
		}
	})

	t.Run("prefers theme.toml over theme.yaml", func(t *testing.T) {
		themeDir := t.TempDir()
		os.WriteFile(filepath.Join(themeDir, "theme.toml"), []byte(`description = "TOML"`), 0644)
		os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(`description: "YAML"`), 0644)

		desc, _ := readThemeMetadata(themeDir)

		if desc != "TOML" {
			t.Errorf("expected 'TOML' (from theme.toml), got %q", desc)
		}
	})

	t.Run("nonexistent theme dir returns empty", func(t *testing.T) {
		desc, minVer := readThemeMetadata("/nonexistent/path")

		if desc != "" {
			t.Errorf("expected empty description, got %q", desc)
		}
		if minVer != "" {
			t.Errorf("expected empty min_version, got %q", minVer)
		}
	})
}

// =============================================================================
// detectSectionsFromLayouts tests (filesystem-based)
// =============================================================================

func TestDetectSectionsFromLayouts(t *testing.T) {
	t.Run("detects sections with layout files", func(t *testing.T) {
		themeDir := t.TempDir()

		// Section with single.html
		os.MkdirAll(filepath.Join(themeDir, "layouts", "posts"), 0755)
		os.WriteFile(filepath.Join(themeDir, "layouts", "posts", "single.html"), []byte(""), 0644)

		// Section with list.html
		os.MkdirAll(filepath.Join(themeDir, "layouts", "docs"), 0755)
		os.WriteFile(filepath.Join(themeDir, "layouts", "docs", "list.html"), []byte(""), 0644)

		// Section with section.html
		os.MkdirAll(filepath.Join(themeDir, "layouts", "projects"), 0755)
		os.WriteFile(filepath.Join(themeDir, "layouts", "projects", "section.html"), []byte(""), 0644)

		// Dir without layout files should be ignored
		os.MkdirAll(filepath.Join(themeDir, "layouts", "empty-section"), 0755)

		result := detectSectionsFromLayouts(themeDir)

		if !containsString(result, "posts") {
			t.Error("expected 'posts' in sections")
		}
		if !containsString(result, "docs") {
			t.Error("expected 'docs' in sections")
		}
		if !containsString(result, "projects") {
			t.Error("expected 'projects' in sections")
		}
		if containsString(result, "empty-section") {
			t.Error("should not include section without layout files")
		}
	})

	t.Run("ignores special directories", func(t *testing.T) {
		themeDir := t.TempDir()

		ignoreDirs := []string{"_default", "partials", "shortcodes", "_markup", "_headers"}
		for _, dir := range ignoreDirs {
			os.MkdirAll(filepath.Join(themeDir, "layouts", dir), 0755)
			os.WriteFile(filepath.Join(themeDir, "layouts", dir, "single.html"), []byte(""), 0644)
		}

		result := detectSectionsFromLayouts(themeDir)

		for _, dir := range ignoreDirs {
			if containsString(result, dir) {
				t.Errorf("should not include ignored directory %q", dir)
			}
		}
	})

	t.Run("nonexistent layouts dir returns empty", func(t *testing.T) {
		result := detectSectionsFromLayouts("/nonexistent")

		if len(result) != 0 {
			t.Errorf("expected 0 sections, got %d", len(result))
		}
	})
}

// =============================================================================
// hasLayoutFiles tests (filesystem-based)
// =============================================================================

func TestHasLayoutFiles(t *testing.T) {
	t.Run("dir with single.html", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "single.html"), []byte(""), 0644)

		if !hasLayoutFiles(dir) {
			t.Error("expected true for dir with single.html")
		}
	})

	t.Run("dir with list.html", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "list.html"), []byte(""), 0644)

		if !hasLayoutFiles(dir) {
			t.Error("expected true for dir with list.html")
		}
	})

	t.Run("dir with section.html", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "section.html"), []byte(""), 0644)

		if !hasLayoutFiles(dir) {
			t.Error("expected true for dir with section.html")
		}
	})

	t.Run("empty dir", func(t *testing.T) {
		dir := t.TempDir()

		if hasLayoutFiles(dir) {
			t.Error("expected false for empty dir")
		}
	})

	t.Run("dir with other files only", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "baseof.html"), []byte(""), 0644)

		if hasLayoutFiles(dir) {
			t.Error("expected false for dir with only baseof.html")
		}
	})
}

// =============================================================================
// EnsureDynamicFrontmatter tests (pure logic, no FS for the function itself)
// =============================================================================

func TestEnsureDynamicFrontmatter_PureLogic(t *testing.T) {
	// Note: EnsureDynamicFrontmatter calls GetDynamicFrontmatterFields internally,
	// which accesses the filesystem. We test the core logic by setting up temp dirs.

	t.Run("adds missing draft field fix", func(t *testing.T) {
		// Set up a minimal site with theme that has no archetypes
		siteDir := t.TempDir()
		themeDir := filepath.Join(siteDir, "themes", "testtheme")
		os.MkdirAll(filepath.Join(themeDir, "layouts", "_default"), 0755)

		content := "---\ntitle: Test\ndraft: true\n---\nBody"
		result, changed := EnsureDynamicFrontmatter(content, siteDir, "testtheme", "posts")

		if !changed {
			t.Error("expected content to be changed (draft: true -> draft: false)")
		}
		if strings.Contains(result, "draft: true") {
			t.Error("draft: true should be replaced with draft: false")
		}
		if !strings.Contains(result, "draft: false") {
			t.Error("should contain draft: false")
		}
	})

	t.Run("no frontmatter returns unchanged", func(t *testing.T) {
		siteDir := t.TempDir()
		themeDir := filepath.Join(siteDir, "themes", "testtheme")
		os.MkdirAll(filepath.Join(themeDir, "layouts", "_default"), 0755)

		content := "Just some content without frontmatter"
		result, changed := EnsureDynamicFrontmatter(content, siteDir, "testtheme", "posts")

		if changed {
			t.Error("expected no changes for content without frontmatter")
		}
		if result != content {
			t.Error("content should be unchanged")
		}
	})

	t.Run("incomplete frontmatter returns unchanged", func(t *testing.T) {
		siteDir := t.TempDir()
		themeDir := filepath.Join(siteDir, "themes", "testtheme")
		os.MkdirAll(filepath.Join(themeDir, "layouts", "_default"), 0755)

		content := "---\ntitle: Test\n"
		result, changed := EnsureDynamicFrontmatter(content, siteDir, "testtheme", "posts")

		if changed {
			t.Error("expected no changes for incomplete frontmatter")
		}
		if result != content {
			t.Error("content should be unchanged")
		}
	})
}
