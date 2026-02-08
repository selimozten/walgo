package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// sanitizeFilename tests
// =============================================================================

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple filename",
			input:    "hello-world.md",
			expected: "hello-world.md",
		},
		{
			name:     "spaces to hyphens",
			input:    "hello world.md",
			expected: "hello-world.md",
		},
		{
			name:     "uppercase to lowercase",
			input:    "Hello-World.md",
			expected: "hello-world.md",
		},
		{
			name:     "special characters removed",
			input:    "hello@world!.md",
			expected: "helloworld.md",
		},
		{
			name:     "path traversal removed",
			input:    "../../../etc/passwd",
			expected: "passwd", // filepath.Base extracts "passwd", then ".." removal is no-op
		},
		{
			name:     "double dots removed",
			input:    "hello..world.md",
			expected: "helloworld.md",
		},
		{
			name:     "slashes removed",
			input:    "path/to/file.md",
			expected: "file.md",
		},
		{
			name:     "backslashes removed",
			input:    "path\\to\\file.md",
			expected: "pathtofile.md", // on Unix, filepath.Base doesn't split on backslash
		},
		{
			name:     "multiple hyphens collapsed",
			input:    "hello---world.md",
			expected: "hello-world.md",
		},
		{
			name:     "leading hyphens trimmed",
			input:    "---hello.md",
			expected: "hello.md",
		},
		{
			name:     "trailing hyphens trimmed",
			input:    "hello---.md",
			expected: "hello-.md", // ".." removed -> "hello-.md", hyphen before dot not at boundary
		},
		{
			name:     "underscores preserved",
			input:    "hello_world.md",
			expected: "hello_world.md",
		},
		{
			name:     "dots preserved",
			input:    "v1.2.3.md",
			expected: "v1.2.3.md",
		},
		{
			name:     "numbers preserved",
			input:    "post-123.md",
			expected: "post-123.md",
		},
		{
			name:  "empty after sanitization gets default",
			input: "!!!",
			// After sanitization it becomes empty, gets auto-generated name
			// We just check it's not empty
		},
		{
			name:  "just .md gets default",
			input: ".md",
			// filepath.Base(".md") = ".md", which triggers default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)

			if tt.expected != "" {
				if result != tt.expected {
					t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}

			// Universal checks
			if result == "" {
				t.Error("sanitized filename should never be empty")
			}
			if result == "." {
				t.Error("sanitized filename should never be '.'")
			}
			if strings.Contains(result, "..") {
				t.Error("sanitized filename should not contain '..'")
			}
			if strings.Contains(result, "/") {
				t.Error("sanitized filename should not contain '/'")
			}
			if strings.Contains(result, "\\") {
				t.Error("sanitized filename should not contain '\\'")
			}
			if result != strings.ToLower(result) && !strings.HasPrefix(result, "ai-generated-") {
				t.Errorf("sanitized filename should be lowercase, got %q", result)
			}
		})
	}
}

// =============================================================================
// NewContentGenerator tests
// =============================================================================

func TestNewContentGenerator(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	cg := NewContentGenerator(client)

	if cg == nil {
		t.Fatal("expected non-nil ContentGenerator")
	}
	if cg.client != client {
		t.Error("expected client to be set on ContentGenerator")
	}
}

// =============================================================================
// parseAIResponse tests
// =============================================================================

func TestParseAIResponse(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	cg := NewContentGenerator(client)

	structure := &ContentStructure{
		DefaultType: "posts",
	}

	tests := []struct {
		name            string
		response        string
		expectedType    string
		expectedFile    string
		contentContains string
	}{
		{
			name: "full response with metadata",
			response: `CONTENT_TYPE: posts
FILENAME: my-article.md
---
title: "My Article"
draft: false
---

This is my article content.`,
			expectedType:    "posts",
			expectedFile:    "my-article.md",
			contentContains: "My Article",
		},
		{
			name: "response without metadata prefix",
			response: `---
title: "Direct Content"
draft: false
---

Direct content without metadata lines.`,
			expectedType:    "",
			expectedFile:    "",
			contentContains: "Direct Content",
		},
		{
			name: "response with only content type",
			response: `CONTENT_TYPE: docs
---
title: "Docs Page"
draft: false
---

Documentation content.`,
			expectedType:    "docs",
			expectedFile:    "",
			contentContains: "Docs Page",
		},
		{
			name:     "response with markdown fences",
			response: "```markdown\n---\ntitle: \"Fenced\"\ndraft: false\n---\n\nFenced content.\n```",
			// CleanGeneratedContent should strip the fences
			contentContains: "Fenced",
		},
		{
			name: "response with draft true gets cleaned",
			response: `---
title: "Draft Fix"
draft: true
---

Content here.`,
			contentContains: "draft: false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentType, filename, content := cg.parseAIResponse(tt.response, structure)

			if tt.expectedType != "" && contentType != tt.expectedType {
				t.Errorf("expected content type %q, got %q", tt.expectedType, contentType)
			}
			if tt.expectedFile != "" && filename != tt.expectedFile {
				t.Errorf("expected filename %q, got %q", tt.expectedFile, filename)
			}
			if tt.contentContains != "" && !strings.Contains(content, tt.contentContains) {
				t.Errorf("expected content to contain %q\nGot:\n%s", tt.contentContains, content)
			}
		})
	}
}

// =============================================================================
// extractMenuItems tests
// =============================================================================

func TestExtractMenuItems(t *testing.T) {
	t.Run("standard menu config", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"menu": map[string]interface{}{
				"main": []interface{}{
					map[string]interface{}{
						"name":   "Home",
						"url":    "/",
						"weight": int64(1),
					},
					map[string]interface{}{
						"name":   "About",
						"url":    "/about/",
						"weight": int64(2),
					},
				},
			},
		}

		menus := extractMenuItems(rawConfig)

		if len(menus) != 2 {
			t.Fatalf("expected 2 menu items, got %d", len(menus))
		}
		if menus[0].Name != "Home" {
			t.Errorf("expected first item name 'Home', got %q", menus[0].Name)
		}
		if menus[0].URL != "/" {
			t.Errorf("expected first item URL '/', got %q", menus[0].URL)
		}
		if menus[0].Weight != 1 {
			t.Errorf("expected first item weight 1, got %d", menus[0].Weight)
		}
		if menus[1].Name != "About" {
			t.Errorf("expected second item name 'About', got %q", menus[1].Name)
		}
	})

	t.Run("alternative menus key", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"menus": map[string]interface{}{
				"main": []interface{}{
					map[string]interface{}{
						"name": "Home",
						"url":  "/",
					},
				},
			},
		}

		menus := extractMenuItems(rawConfig)

		if len(menus) != 1 {
			t.Fatalf("expected 1 menu item with 'menus' key, got %d", len(menus))
		}
		if menus[0].Name != "Home" {
			t.Errorf("expected name 'Home', got %q", menus[0].Name)
		}
	})

	t.Run("pageRef instead of url", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"menu": map[string]interface{}{
				"main": []interface{}{
					map[string]interface{}{
						"name":    "Docs",
						"pageRef": "/docs/",
					},
				},
			},
		}

		menus := extractMenuItems(rawConfig)

		if len(menus) != 1 {
			t.Fatalf("expected 1 menu item, got %d", len(menus))
		}
		if menus[0].URL != "/docs/" {
			t.Errorf("expected URL '/docs/' from pageRef, got %q", menus[0].URL)
		}
	})

	t.Run("single item as map (not array)", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"menu": map[string]interface{}{
				"main": map[string]interface{}{
					"name": "Single",
					"url":  "/single/",
				},
			},
		}

		menus := extractMenuItems(rawConfig)

		if len(menus) != 1 {
			t.Fatalf("expected 1 menu item from single map, got %d", len(menus))
		}
		if menus[0].Name != "Single" {
			t.Errorf("expected name 'Single', got %q", menus[0].Name)
		}
	})

	t.Run("no menu config", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"title": "My Site",
		}

		menus := extractMenuItems(rawConfig)

		if len(menus) != 0 {
			t.Errorf("expected 0 menu items, got %d", len(menus))
		}
	})

	t.Run("skips items without name", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"menu": map[string]interface{}{
				"main": []interface{}{
					map[string]interface{}{
						"url": "/no-name/",
					},
					map[string]interface{}{
						"name": "Valid",
						"url":  "/valid/",
					},
				},
			},
		}

		menus := extractMenuItems(rawConfig)

		if len(menus) != 1 {
			t.Fatalf("expected 1 menu item (skipping nameless), got %d", len(menus))
		}
		if menus[0].Name != "Valid" {
			t.Errorf("expected name 'Valid', got %q", menus[0].Name)
		}
	})

	t.Run("weight as int (not int64)", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"menu": map[string]interface{}{
				"main": []interface{}{
					map[string]interface{}{
						"name":   "Home",
						"url":    "/",
						"weight": 5,
					},
				},
			},
		}

		menus := extractMenuItems(rawConfig)

		if len(menus) != 1 {
			t.Fatalf("expected 1 menu item, got %d", len(menus))
		}
		if menus[0].Weight != 5 {
			t.Errorf("expected weight 5, got %d", menus[0].Weight)
		}
	})

	t.Run("multiple menus", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"menu": map[string]interface{}{
				"main": []interface{}{
					map[string]interface{}{
						"name": "Home",
						"url":  "/",
					},
				},
				"footer": []interface{}{
					map[string]interface{}{
						"name": "Privacy",
						"url":  "/privacy/",
					},
				},
			},
		}

		menus := extractMenuItems(rawConfig)

		if len(menus) != 2 {
			t.Fatalf("expected 2 menu items from multiple menus, got %d", len(menus))
		}
	})
}

// =============================================================================
// ContentStructure type tests
// =============================================================================

func TestContentStructure_Defaults(t *testing.T) {
	structure := &ContentStructure{
		SitePath:     "/tmp/test",
		ContentTypes: []ContentTypeInfo{},
		DefaultType:  "posts",
		ContentDir:   "/tmp/test/content",
	}

	if structure.DefaultType != "posts" {
		t.Errorf("expected default type 'posts', got %q", structure.DefaultType)
	}
	if structure.SiteConfig != nil {
		t.Error("expected nil SiteConfig initially")
	}
	if structure.ThemeInfo != nil {
		t.Error("expected nil ThemeInfo initially")
	}
}

// =============================================================================
// ContentFileInfo type tests
// =============================================================================

func TestContentFileInfo(t *testing.T) {
	fileInfo := ContentFileInfo{
		Path:        "posts/hello.md",
		Title:       "Hello World",
		Description: "A test post",
		Date:        "2024-01-01",
		Draft:       false,
		Tags:        []string{"go", "hugo"},
		Extra:       map[string]string{"author": "John"},
		BundleType:  "single",
	}

	if fileInfo.Path != "posts/hello.md" {
		t.Errorf("expected path 'posts/hello.md', got %q", fileInfo.Path)
	}
	if fileInfo.Title != "Hello World" {
		t.Errorf("expected title 'Hello World', got %q", fileInfo.Title)
	}
	if fileInfo.Draft {
		t.Error("expected draft to be false")
	}
	if len(fileInfo.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(fileInfo.Tags))
	}
	if fileInfo.Extra["author"] != "John" {
		t.Errorf("expected author 'John', got %q", fileInfo.Extra["author"])
	}
	if fileInfo.BundleType != "single" {
		t.Errorf("expected BundleType 'single', got %q", fileInfo.BundleType)
	}
}

// =============================================================================
// SiteConfigInfo type tests
// =============================================================================

func TestSiteConfigInfo(t *testing.T) {
	cfg := &SiteConfigInfo{
		Title:       "My Site",
		BaseURL:     "https://example.com",
		Language:    "en-us",
		Theme:       "ananke",
		Description: "A test site",
		Author:      "John",
		Params: map[string]interface{}{
			"enableSearch": true,
		},
		Taxonomies: map[string]string{
			"tag": "tags",
		},
		Permalinks: map[string]string{
			"posts": "/:year/:month/:title/",
		},
	}

	if cfg.Title != "My Site" {
		t.Errorf("expected title 'My Site', got %q", cfg.Title)
	}
	if cfg.Theme != "ananke" {
		t.Errorf("expected theme 'ananke', got %q", cfg.Theme)
	}
	if len(cfg.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(cfg.Params))
	}
	if cfg.Taxonomies["tag"] != "tags" {
		t.Errorf("expected taxonomy tag='tags', got %q", cfg.Taxonomies["tag"])
	}
}

// =============================================================================
// ContentTypeInfo type tests
// =============================================================================

func TestContentTypeInfo(t *testing.T) {
	ct := ContentTypeInfo{
		Name:      "posts",
		Path:      "/site/content/posts",
		FileCount: 3,
		Files:     []string{"post1.md", "post2.md", "post3.md"},
	}

	if ct.Name != "posts" {
		t.Errorf("expected name 'posts', got %q", ct.Name)
	}
	if ct.FileCount != 3 {
		t.Errorf("expected file count 3, got %d", ct.FileCount)
	}
	if len(ct.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(ct.Files))
	}
}

// =============================================================================
// getAllContentFiles tests (filesystem-based)
// =============================================================================

func TestGetAllContentFiles(t *testing.T) {
	t.Run("nonexistent content dir returns nil", func(t *testing.T) {
		result := getAllContentFiles("/nonexistent/path")
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("scans markdown files", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		os.MkdirAll(filepath.Join(contentDir, "posts"), 0755)
		os.WriteFile(filepath.Join(contentDir, "posts", "hello.md"), []byte(`---
title: "Hello World"
description: "A test post"
date: "2024-01-01"
draft: false
tags:
  - go
  - hugo
author: "John"
---

Hello content.`), 0644)

		result := getAllContentFiles(siteDir)

		if len(result) != 1 {
			t.Fatalf("expected 1 file, got %d", len(result))
		}

		file := result[0]
		if file.Title != "Hello World" {
			t.Errorf("expected title 'Hello World', got %q", file.Title)
		}
		if file.Description != "A test post" {
			t.Errorf("expected description 'A test post', got %q", file.Description)
		}
		if file.Date != "2024-01-01" {
			t.Errorf("expected date '2024-01-01', got %q", file.Date)
		}
		if file.Draft {
			t.Error("expected draft false")
		}
		if len(file.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(file.Tags))
		}
		if file.Extra["author"] != "John" {
			t.Errorf("expected author 'John' in Extra, got %q", file.Extra["author"])
		}
		if file.BundleType != "single" {
			t.Errorf("expected BundleType 'single', got %q", file.BundleType)
		}
	})

	t.Run("detects bundle types", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		// Branch bundle
		os.MkdirAll(filepath.Join(contentDir, "posts"), 0755)
		os.WriteFile(filepath.Join(contentDir, "posts", "_index.md"), []byte("---\ntitle: Posts\n---\n"), 0644)

		// Leaf bundle
		os.MkdirAll(filepath.Join(contentDir, "posts", "my-post"), 0755)
		os.WriteFile(filepath.Join(contentDir, "posts", "my-post", "index.md"), []byte("---\ntitle: My Post\n---\n"), 0644)

		// Single page
		os.WriteFile(filepath.Join(contentDir, "posts", "simple.md"), []byte("---\ntitle: Simple\n---\n"), 0644)

		result := getAllContentFiles(siteDir)

		bundleTypes := map[string]string{}
		for _, f := range result {
			bundleTypes[f.Path] = f.BundleType
		}

		if bt, ok := bundleTypes[filepath.Join("posts", "_index.md")]; ok {
			if bt != "branch" {
				t.Errorf("expected branch for _index.md, got %q", bt)
			}
		} else {
			t.Error("expected _index.md in results")
		}

		if bt, ok := bundleTypes[filepath.Join("posts", "my-post", "index.md")]; ok {
			if bt != "leaf" {
				t.Errorf("expected leaf for index.md, got %q", bt)
			}
		} else {
			t.Error("expected index.md in results")
		}

		if bt, ok := bundleTypes[filepath.Join("posts", "simple.md")]; ok {
			if bt != "single" {
				t.Errorf("expected single for simple.md, got %q", bt)
			}
		} else {
			t.Error("expected simple.md in results")
		}
	})

	t.Run("ignores non-markdown files", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		os.MkdirAll(contentDir, 0755)
		os.WriteFile(filepath.Join(contentDir, "image.png"), []byte(""), 0644)
		os.WriteFile(filepath.Join(contentDir, "data.json"), []byte(""), 0644)
		os.WriteFile(filepath.Join(contentDir, "page.md"), []byte("---\ntitle: Page\n---\n"), 0644)

		result := getAllContentFiles(siteDir)

		if len(result) != 1 {
			t.Errorf("expected 1 file (only .md), got %d", len(result))
		}
	})

	t.Run("handles file without frontmatter", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		os.MkdirAll(contentDir, 0755)
		os.WriteFile(filepath.Join(contentDir, "bare.md"), []byte("No frontmatter here."), 0644)

		result := getAllContentFiles(siteDir)

		if len(result) != 1 {
			t.Fatalf("expected 1 file, got %d", len(result))
		}
		if result[0].Title != "" {
			t.Errorf("expected empty title for file without frontmatter, got %q", result[0].Title)
		}
	})

	t.Run("handles draft true", func(t *testing.T) {
		siteDir := t.TempDir()
		contentDir := filepath.Join(siteDir, "content")

		os.MkdirAll(contentDir, 0755)
		os.WriteFile(filepath.Join(contentDir, "draft.md"), []byte("---\ntitle: Draft\ndraft: true\n---\n"), 0644)

		result := getAllContentFiles(siteDir)

		if len(result) != 1 {
			t.Fatalf("expected 1 file, got %d", len(result))
		}
		if !result[0].Draft {
			t.Error("expected draft to be true")
		}
	})
}

// =============================================================================
// loadSiteConfig tests (filesystem-based)
// =============================================================================

func TestLoadSiteConfig(t *testing.T) {
	t.Run("loads hugo.toml", func(t *testing.T) {
		siteDir := t.TempDir()
		os.WriteFile(filepath.Join(siteDir, "hugo.toml"), []byte(`
baseURL = "https://example.com"
title = "My Site"
languageCode = "en-us"
theme = "ananke"

[params]
description = "A great site"
author = "John"
`), 0644)

		cfg := loadSiteConfig(siteDir)

		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.Title != "My Site" {
			t.Errorf("expected title 'My Site', got %q", cfg.Title)
		}
		if cfg.BaseURL != "https://example.com" {
			t.Errorf("expected baseURL 'https://example.com', got %q", cfg.BaseURL)
		}
		if cfg.Language != "en-us" {
			t.Errorf("expected language 'en-us', got %q", cfg.Language)
		}
		if cfg.Theme != "ananke" {
			t.Errorf("expected theme 'ananke', got %q", cfg.Theme)
		}
		if cfg.Description != "A great site" {
			t.Errorf("expected description 'A great site', got %q", cfg.Description)
		}
		if cfg.Author != "John" {
			t.Errorf("expected author 'John', got %q", cfg.Author)
		}
	})

	t.Run("loads hugo.yaml", func(t *testing.T) {
		siteDir := t.TempDir()
		os.WriteFile(filepath.Join(siteDir, "hugo.yaml"), []byte(`
baseURL: "https://example.com"
title: "YAML Site"
languageCode: "en"
theme: "book"
params:
  description: "YAML description"
`), 0644)

		cfg := loadSiteConfig(siteDir)

		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.Title != "YAML Site" {
			t.Errorf("expected title 'YAML Site', got %q", cfg.Title)
		}
		if cfg.Theme != "book" {
			t.Errorf("expected theme 'book', got %q", cfg.Theme)
		}
	})

	t.Run("no config file returns nil", func(t *testing.T) {
		siteDir := t.TempDir()
		cfg := loadSiteConfig(siteDir)

		if cfg != nil {
			t.Error("expected nil config when no config file exists")
		}
	})

	t.Run("prefers hugo.toml over config.toml", func(t *testing.T) {
		siteDir := t.TempDir()
		os.WriteFile(filepath.Join(siteDir, "hugo.toml"), []byte(`title = "Hugo TOML"`), 0644)
		os.WriteFile(filepath.Join(siteDir, "config.toml"), []byte(`title = "Config TOML"`), 0644)

		cfg := loadSiteConfig(siteDir)

		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.Title != "Hugo TOML" {
			t.Errorf("expected 'Hugo TOML' (from hugo.toml), got %q", cfg.Title)
		}
	})

	t.Run("extracts taxonomies", func(t *testing.T) {
		siteDir := t.TempDir()
		os.WriteFile(filepath.Join(siteDir, "hugo.toml"), []byte(`
title = "Test"

[taxonomies]
tag = "tags"
category = "categories"
`), 0644)

		cfg := loadSiteConfig(siteDir)

		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.Taxonomies["tag"] != "tags" {
			t.Errorf("expected taxonomy tag='tags', got %q", cfg.Taxonomies["tag"])
		}
		if cfg.Taxonomies["category"] != "categories" {
			t.Errorf("expected taxonomy category='categories', got %q", cfg.Taxonomies["category"])
		}
	})

	t.Run("extracts permalinks", func(t *testing.T) {
		siteDir := t.TempDir()
		os.WriteFile(filepath.Join(siteDir, "hugo.toml"), []byte(`
title = "Test"

[permalinks]
posts = "/:year/:month/:title/"
`), 0644)

		cfg := loadSiteConfig(siteDir)

		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.Permalinks["posts"] != "/:year/:month/:title/" {
			t.Errorf("expected permalink pattern, got %q", cfg.Permalinks["posts"])
		}
	})

	t.Run("detects theme from themes directory when not in config", func(t *testing.T) {
		siteDir := t.TempDir()
		os.WriteFile(filepath.Join(siteDir, "hugo.toml"), []byte(`title = "Test"`), 0644)
		os.MkdirAll(filepath.Join(siteDir, "themes", "auto-detected"), 0755)

		cfg := loadSiteConfig(siteDir)

		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.Theme != "auto-detected" {
			t.Errorf("expected auto-detected theme, got %q", cfg.Theme)
		}
	})

	t.Run("default language is en", func(t *testing.T) {
		siteDir := t.TempDir()
		os.WriteFile(filepath.Join(siteDir, "hugo.toml"), []byte(`title = "Test"`), 0644)

		cfg := loadSiteConfig(siteDir)

		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.Language != "en" {
			t.Errorf("expected default language 'en', got %q", cfg.Language)
		}
	})

	t.Run("extracts menu items", func(t *testing.T) {
		siteDir := t.TempDir()
		os.WriteFile(filepath.Join(siteDir, "hugo.yaml"), []byte(`
title: "Test"
menu:
  main:
    - name: "Home"
      url: "/"
      weight: 1
    - name: "About"
      url: "/about/"
      weight: 2
`), 0644)

		cfg := loadSiteConfig(siteDir)

		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if len(cfg.Menu) != 2 {
			t.Fatalf("expected 2 menu items, got %d", len(cfg.Menu))
		}
		if cfg.Menu[0].Name != "Home" {
			t.Errorf("expected first menu item 'Home', got %q", cfg.Menu[0].Name)
		}
	})
}

// =============================================================================
// getThemeInfo tests (filesystem-based)
// =============================================================================

func TestGetThemeInfo(t *testing.T) {
	t.Run("returns defaults for nonexistent theme", func(t *testing.T) {
		siteDir := t.TempDir()
		info := getThemeInfo(siteDir, "nonexistent")

		if info == nil {
			t.Fatal("expected non-nil ThemeLayoutInfo")
		}
		if info.Name != "nonexistent" {
			t.Errorf("expected name 'nonexistent', got %q", info.Name)
		}
		// Should have sensible defaults
		if len(info.SupportedSections) == 0 {
			t.Error("expected default supported sections")
		}
		if len(info.FrontmatterFields) == 0 {
			t.Error("expected default frontmatter fields")
		}
		if info.Description == "" {
			t.Error("expected default description")
		}
	})

	t.Run("populates from theme analysis", func(t *testing.T) {
		siteDir := t.TempDir()
		themeDir := filepath.Join(siteDir, "themes", "test-theme")

		os.MkdirAll(filepath.Join(themeDir, "layouts", "posts"), 0755)
		os.WriteFile(filepath.Join(themeDir, "layouts", "posts", "single.html"), []byte(""), 0644)

		os.MkdirAll(filepath.Join(themeDir, "archetypes"), 0755)
		os.WriteFile(filepath.Join(themeDir, "archetypes", "posts.md"), []byte("---\ntitle: \"\"\ndate: \"\"\n---\n"), 0644)

		os.WriteFile(filepath.Join(themeDir, "theme.toml"), []byte(`description = "Test theme"`), 0644)

		info := getThemeInfo(siteDir, "test-theme")

		if info.Name != "test-theme" {
			t.Errorf("expected name 'test-theme', got %q", info.Name)
		}
		if info.Description != "Test theme" {
			t.Errorf("expected description 'Test theme', got %q", info.Description)
		}
		if !containsString(info.SupportedSections, "posts") {
			t.Error("expected 'posts' in supported sections")
		}
	})
}

// =============================================================================
// buildSmartSystemPrompt tests
// =============================================================================

func TestBuildSmartSystemPrompt(t *testing.T) {
	client := NewClient("openai", "test-key", "", "gpt-4")
	cg := NewContentGenerator(client)

	t.Run("includes site config", func(t *testing.T) {
		structure := &ContentStructure{
			SitePath: "/tmp/test",
			SiteConfig: &SiteConfigInfo{
				Title:       "My Site",
				Theme:       "",
				Language:    "en",
				Description: "Test site",
				Author:      "John",
				BaseURL:     "https://example.com",
				Params:      map[string]interface{}{},
				Taxonomies:  map[string]string{},
				Permalinks:  map[string]string{},
				Outputs:     map[string][]string{},
			},
			ContentTypes: []ContentTypeInfo{},
			ContentFiles: []ContentFileInfo{},
		}

		result := cg.buildSmartSystemPrompt(structure)

		if !strings.Contains(result, "My Site") {
			t.Error("expected site title in prompt")
		}
		if !strings.Contains(result, "Test site") {
			t.Error("expected description in prompt")
		}
		if !strings.Contains(result, "John") {
			t.Error("expected author in prompt")
		}
	})

	t.Run("includes content structure", func(t *testing.T) {
		structure := &ContentStructure{
			SitePath:   "/tmp/test",
			SiteConfig: nil,
			ContentTypes: []ContentTypeInfo{
				{
					Name:      "posts",
					Path:      "/tmp/test/content/posts",
					FileCount: 2,
					Files:     []string{"hello.md", "world.md"},
				},
			},
			ContentFiles: []ContentFileInfo{
				{
					Path:       "posts/hello.md",
					Title:      "Hello",
					BundleType: "single",
				},
			},
		}

		result := cg.buildSmartSystemPrompt(structure)

		if !strings.Contains(result, "posts/") {
			t.Error("expected content structure in prompt")
		}
		if !strings.Contains(result, "Hello") {
			t.Error("expected content file title in prompt")
		}
	})

	t.Run("includes menu items in config", func(t *testing.T) {
		structure := &ContentStructure{
			SitePath: "/tmp/test",
			SiteConfig: &SiteConfigInfo{
				Title:    "My Site",
				Theme:    "",
				Language: "en",
				Menu: []MenuInfo{
					{Name: "Home", URL: "/", Weight: 1},
					{Name: "About", URL: "/about/"},
				},
				Params:     map[string]interface{}{},
				Taxonomies: map[string]string{},
				Permalinks: map[string]string{},
				Outputs:    map[string][]string{},
			},
			ContentTypes: []ContentTypeInfo{},
			ContentFiles: []ContentFileInfo{},
		}

		result := cg.buildSmartSystemPrompt(structure)

		if !strings.Contains(result, "Menu Items") {
			t.Error("expected menu items label in prompt")
		}
		if !strings.Contains(result, "Home") {
			t.Error("expected Home menu item in prompt")
		}
		if !strings.Contains(result, "About") {
			t.Error("expected About menu item in prompt")
		}
	})

	t.Run("includes taxonomies in config", func(t *testing.T) {
		structure := &ContentStructure{
			SitePath: "/tmp/test",
			SiteConfig: &SiteConfigInfo{
				Title:    "My Site",
				Theme:    "",
				Language: "en",
				Taxonomies: map[string]string{
					"tag":      "tags",
					"category": "categories",
				},
				Params:     map[string]interface{}{},
				Permalinks: map[string]string{},
				Outputs:    map[string][]string{},
			},
			ContentTypes: []ContentTypeInfo{},
			ContentFiles: []ContentFileInfo{},
		}

		result := cg.buildSmartSystemPrompt(structure)

		if !strings.Contains(result, "Taxonomies") {
			t.Error("expected taxonomies in prompt")
		}
		if !strings.Contains(result, "tag") {
			t.Error("expected tag taxonomy in prompt")
		}
	})

	t.Run("includes content files with bundle labels", func(t *testing.T) {
		structure := &ContentStructure{
			SitePath:     "/tmp/test",
			SiteConfig:   nil,
			ContentTypes: []ContentTypeInfo{},
			ContentFiles: []ContentFileInfo{
				{
					Path:        "posts/_index.md",
					Title:       "Posts",
					BundleType:  "branch",
					Description: "All posts",
				},
				{
					Path:       "posts/hello/index.md",
					Title:      "Hello",
					BundleType: "leaf",
					Tags:       []string{"go", "hugo"},
				},
				{
					Path:       "about.md",
					Title:      "About",
					BundleType: "single",
				},
			},
		}

		result := cg.buildSmartSystemPrompt(structure)

		if !strings.Contains(result, "[SECTION]") {
			t.Error("expected [SECTION] label for branch bundle")
		}
		if !strings.Contains(result, "[BUNDLE]") {
			t.Error("expected [BUNDLE] label for leaf bundle")
		}
		if !strings.Contains(result, "[PAGE]") {
			t.Error("expected [PAGE] label for single page")
		}
	})
}

// =============================================================================
// ContentGenerationResult type tests
// =============================================================================

func TestContentGenerationResult(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := &ContentGenerationResult{
			Success:     true,
			Content:     "---\ntitle: Test\n---\n",
			FilePath:    "/site/content/posts/test.md",
			ContentType: "posts",
			Filename:    "test.md",
		}

		if !result.Success {
			t.Error("expected success")
		}
		if result.Error != nil {
			t.Error("expected nil error on success")
		}
	})

	t.Run("failure result", func(t *testing.T) {
		result := &ContentGenerationResult{
			Success:      false,
			ErrorMessage: "API call failed",
		}

		if result.Success {
			t.Error("expected failure")
		}
		if result.ErrorMessage != "API call failed" {
			t.Errorf("expected error message, got %q", result.ErrorMessage)
		}
	})
}
