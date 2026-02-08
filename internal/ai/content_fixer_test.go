package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewContentFixer(t *testing.T) {
	fixer := NewContentFixer("/path/to/site", SiteTypeBlog)

	if fixer.sitePath != "/path/to/site" {
		t.Errorf("expected sitePath '/path/to/site', got %s", fixer.sitePath)
	}
	if fixer.siteType != SiteTypeBlog {
		t.Errorf("expected siteType blog, got %s", fixer.siteType)
	}
}

func TestContentFixer_FixAll_NoContentDir(t *testing.T) {
	tempDir := t.TempDir()
	// Don't create content directory

	fixer := NewContentFixer(tempDir, SiteTypeBlog)
	err := fixer.FixAll()

	if err != nil {
		t.Errorf("FixAll should not error when content dir doesn't exist: %v", err)
	}
}

func TestContentFixer_FixAll_Blog(t *testing.T) {
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	os.MkdirAll(contentDir, 0755)

	// Create a file that needs fixing
	aboutPath := filepath.Join(contentDir, "about.md")
	originalContent := `---
title: About
---

# About Me

This is the about page.`

	os.WriteFile(aboutPath, []byte(originalContent), 0644)

	fixer := NewContentFixer(tempDir, SiteTypeBlog)
	err := fixer.FixAll()

	if err != nil {
		t.Fatalf("FixAll failed: %v", err)
	}

	// Read and verify
	content, _ := os.ReadFile(aboutPath)
	contentStr := string(content)

	// Should have description added
	if !strings.Contains(contentStr, "description:") {
		t.Error("description should be added")
	}
	// Should have featured_image added
	if !strings.Contains(contentStr, "featured_image:") {
		t.Error("featured_image should be added")
	}
}

func TestContentFixer_FixBlogContent(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		content        string
		expectedChecks []string
	}{
		{
			name: "home page",
			path: "content/_index.md",
			content: `---
title: My Blog
---

Content here.`,
			expectedChecks: []string{"description:", "featured_image:"},
		},
		{
			name: "about page",
			path: "content/about.md",
			content: `---
title: About
---

About content.`,
			expectedChecks: []string{"description:", "featured_image:"},
		},
		{
			name: "blog post without date",
			path: "content/posts/welcome/index.md",
			content: `---
title: Welcome
---

Post content.`,
			expectedChecks: []string{"date:", "draft: false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixer := NewContentFixer("", SiteTypeBlog)
			result, changed := fixer.fixBlogContent(tt.path, tt.content)

			if !changed {
				t.Error("expected content to be changed")
			}

			for _, check := range tt.expectedChecks {
				if !strings.Contains(result, check) {
					t.Errorf("expected %q in result:\n%s", check, result)
				}
			}
		})
	}
}

func TestContentFixer_FixDocsContent(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		content        string
		expectedChecks []string
	}{
		{
			name: "home page",
			path: "content/_index.md",
			content: `---
title: Docs
---

Content.`,
			expectedChecks: []string{"draft: false", "weight:"},
		},
		{
			name: "docs index",
			path: "content/docs/_index.md",
			content: `---
title: Documentation
---

Docs content.`,
			expectedChecks: []string{"draft: false", "weight:"},
		},
		{
			name: "doc page",
			path: "content/docs/intro/installation.md",
			content: `---
title: Installation
---

Install guide.`,
			expectedChecks: []string{"draft: false", "weight:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixer := NewContentFixer("", SiteTypeDocs)
			result, changed := fixer.fixDocsContent(tt.path, tt.content)

			if !changed {
				t.Error("expected content to be changed")
			}

			for _, check := range tt.expectedChecks {
				if !strings.Contains(result, check) {
					t.Errorf("expected %q in result:\n%s", check, result)
				}
			}
		})
	}
}

func TestFixYAMLQuotes(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedChange bool
		expectedOutput string
	}{
		{
			name:           "simple title - no special chars",
			input:          "---\ntitle: Simple Title\n---\nContent",
			expectedChange: false,
			expectedOutput: "",
		},
		{
			name:           "no frontmatter",
			input:          "Just content without frontmatter",
			expectedChange: false,
			expectedOutput: "",
		},
		{
			name:           "proper frontmatter without issues",
			input:          "---\ntitle: My Title\ndescription: A description\n---\nContent",
			expectedChange: false,
			expectedOutput: "",
		},
		{
			name:           "single quotes with apostrophe - MUST FIX",
			input:          "---\ntitle: 'Contact test1000'\ndescription: 'Get in touch with test1000 for collaborations, feedback, or general inquiries. Let's start a conversation today.'\ndraft: false\n---\nContent",
			expectedChange: true,
			expectedOutput: `---
title: "Contact test1000"
description: "Get in touch with test1000 for collaborations, feedback, or general inquiries. Let's start a conversation today."
draft: false
---
Content`,
		},
		{
			name:           "single quotes with colon - MUST FIX",
			input:          "---\ntitle: 'Welcome: A New Beginning'\n---\nContent",
			expectedChange: true,
			expectedOutput: `---
title: "Welcome: A New Beginning"
---
Content`,
		},
		{
			name:           "already double quoted - no change",
			input:          "---\ntitle: \"Contact test1000\"\ndescription: \"Let's connect today\"\n---\nContent",
			expectedChange: false,
			expectedOutput: "",
		},
		{
			name:           "unquoted with colon - MUST FIX",
			input:          "---\ntitle: Welcome: A New Beginning\n---\nContent",
			expectedChange: true,
			expectedOutput: `---
title: "Welcome: A New Beginning"
---
Content`,
		},
		{
			name:           "unquoted with apostrophe - MUST FIX",
			input:          "---\ndescription: Let's start today\n---\nContent",
			expectedChange: true,
			expectedOutput: `---
description: "Let's start today"
---
Content`,
		},
		{
			name:           "boolean values - no change",
			input:          "---\ntitle: Test\ndraft: false\nfeatured: true\n---\nContent",
			expectedChange: false,
			expectedOutput: "",
		},
		{
			name:           "numeric values - no change",
			input:          "---\ntitle: Test\nweight: 10\nprice: 99.99\n---\nContent",
			expectedChange: false,
			expectedOutput: "",
		},
		{
			name:           "malformed single quote - unclosed",
			input:          "---\ntitle: 'TechCorp Increased Test Coverage by 40% - Heres How\ndate: 2023-10-27T09:00:00Z\n---\nContent",
			expectedChange: true,
			expectedOutput: `---
title: "TechCorp Increased Test Coverage by 40% - Heres How"
date: 2023-10-27T09:00:00Z
---
Content`,
		},
		{
			name:           "single quoted array - MUST FIX",
			input:          "---\ntitle: Test\ntags: ['testing', 'qa', 'case study']\n---\nContent",
			expectedChange: true,
			expectedOutput: `---
title: Test
tags: ["testing", "qa", "case study"]
---
Content`,
		},
		{
			name:           "already double quoted array - no change",
			input:          "---\ntitle: Test\ntags: [\"testing\", \"qa\"]\n---\nContent",
			expectedChange: false,
			expectedOutput: "",
		},
		{
			name:           "complex case - multiple issues",
			input:          "---\ntitle: 'TechCorp: A Case Study'\ndescription: 'Here's how we did it'\ntags: ['test', 'qa']\ndraft: false\n---\nContent",
			expectedChange: true,
			expectedOutput: `---
title: "TechCorp: A Case Study"
description: "Here's how we did it"
tags: ["test", "qa"]
draft: false
---
Content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := fixYAMLQuotes(tt.input)

			if changed != tt.expectedChange {
				t.Errorf("expected change=%v, got %v", tt.expectedChange, changed)
			}

			if tt.expectedOutput != "" && result != tt.expectedOutput {
				t.Errorf("output mismatch:\nexpected:\n%s\n\ngot:\n%s", tt.expectedOutput, result)
			}
		})
	}
}

func TestFixYAMLArray(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		expectedChange bool
	}{
		{
			name:           "single quoted array",
			input:          "['item1', 'item2', 'item3']",
			expectedOutput: "[\"item1\", \"item2\", \"item3\"]",
			expectedChange: true,
		},
		{
			name:           "array with special chars",
			input:          "['testing', 'qa', 'case study']",
			expectedOutput: "[\"testing\", \"qa\", \"case study\"]",
			expectedChange: true,
		},
		{
			name:           "already double quoted",
			input:          "[\"item1\", \"item2\"]",
			expectedOutput: "[\"item1\", \"item2\"]",
			expectedChange: false,
		},
		{
			name:           "empty array",
			input:          "[]",
			expectedOutput: "[]",
			expectedChange: false,
		},
		{
			name:           "not an array",
			input:          "simple string",
			expectedOutput: "simple string",
			expectedChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := fixYAMLArray(tt.input)

			if changed != tt.expectedChange {
				t.Errorf("expected change=%v, got %v", tt.expectedChange, changed)
			}

			if result != tt.expectedOutput {
				t.Errorf("output mismatch:\nexpected: %s\ngot: %s", tt.expectedOutput, result)
			}
		})
	}
}

func TestFixFrontmatterStart(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedChange bool
	}{
		{
			name:           "already starts with ---",
			input:          "---\ntitle: Test\n---\nContent",
			expectedChange: false,
		},
		{
			name:           "starts with garbage",
			input:          "markdown\n---\ntitle: Test\n---\nContent",
			expectedChange: true,
		},
		{
			name:           "empty content",
			input:          "",
			expectedChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, changed := fixFrontmatterStart(tt.input)

			if changed != tt.expectedChange {
				t.Errorf("expected change=%v, got %v", tt.expectedChange, changed)
			}
		})
	}
}

func TestRemoveDuplicateH1(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedChange bool
		notContains    string
	}{
		{
			name: "has duplicate H1",
			input: `---
title: Test
---

# Test

Content here.`,
			expectedChange: true,
			notContains:    "# Test",
		},
		{
			name: "no H1",
			input: `---
title: Test
---

Content here.`,
			expectedChange: false,
			notContains:    "",
		},
		{
			name:           "no frontmatter",
			input:          "# Test\n\nContent",
			expectedChange: false,
			notContains:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := removeDuplicateH1(tt.input)

			if changed != tt.expectedChange {
				t.Errorf("expected change=%v, got %v", tt.expectedChange, changed)
			}
			if tt.notContains != "" && strings.Contains(result, tt.notContains) {
				t.Errorf("should not contain %q in result:\n%s", tt.notContains, result)
			}
		})
	}
}

func TestExtractFrontmatterField(t *testing.T) {
	content := `---
title: My Title
description: 'A description'
author: John Doe
---

Content`

	tests := []struct {
		field    string
		expected string
	}{
		{"title", "My Title"},
		{"description", "A description"},
		{"author", "John Doe"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := extractFrontmatterField(content, tt.field)
			if result != tt.expected {
				t.Errorf("extractFrontmatterField(%s) = %s, want %s", tt.field, result, tt.expected)
			}
		})
	}
}

func TestAddFrontmatterField(t *testing.T) {
	content := `---
title: Test
---

Content`

	tests := []struct {
		field    string
		value    string
		expected string
	}{
		{"description", "A test", "description: 'A test'"},
		{"weight", "1", "weight: 1"},
		{"featured", "true", "featured: true"},
		{"draft", "false", "draft: false"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := addFrontmatterField(content, tt.field, tt.value)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("expected %q in result:\n%s", tt.expected, result)
			}
		})
	}
}

func TestAddFrontmatterField_NoFrontmatter(t *testing.T) {
	content := "Just content without frontmatter"
	result := addFrontmatterField(content, "title", "Test")

	if result != content {
		t.Error("should not modify content without frontmatter")
	}
}

func TestValidateBlogContent(t *testing.T) {
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	postsDir := filepath.Join(contentDir, "posts")
	os.MkdirAll(postsDir, 0755)

	// Create minimal valid structure
	os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte("---\ntitle: Blog\ndescription: A blog\n---"), 0644)
	os.WriteFile(filepath.Join(contentDir, "about.md"), []byte("---\ntitle: About\ndescription: About\n---"), 0644)
	os.WriteFile(filepath.Join(contentDir, "contact.md"), []byte("---\ntitle: Contact\n---"), 0644)
	os.WriteFile(filepath.Join(postsDir, "welcome.md"), []byte("---\ntitle: Welcome\ndate: 2024-01-01\n---"), 0644)

	issues := ValidateBlogContent(tempDir)

	if len(issues) != 0 {
		t.Errorf("expected no issues for valid blog, got: %v", issues)
	}
}

func TestValidateBlogContent_MissingFiles(t *testing.T) {
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	os.MkdirAll(contentDir, 0755)

	// Only create _index.md
	os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte("---\ntitle: Blog\n---"), 0644)

	issues := ValidateBlogContent(tempDir)

	// Should have issues for missing about.md, contact.md, posts
	if len(issues) == 0 {
		t.Error("expected issues for missing files")
	}

	hasAboutIssue := false
	hasContactIssue := false
	hasPostsIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "about.md") {
			hasAboutIssue = true
		}
		if strings.Contains(issue, "contact.md") {
			hasContactIssue = true
		}
		if strings.Contains(issue, "posts") {
			hasPostsIssue = true
		}
	}

	if !hasAboutIssue {
		t.Error("expected issue for missing about.md")
	}
	if !hasContactIssue {
		t.Error("expected issue for missing contact.md")
	}
	if !hasPostsIssue {
		t.Error("expected issue for missing posts")
	}
}

func TestValidateDocsContent(t *testing.T) {
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	docsDir := filepath.Join(contentDir, "docs")
	introDir := filepath.Join(docsDir, "intro")
	os.MkdirAll(introDir, 0755)

	// Create valid structure
	os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte("---\ntitle: Docs\ndescription: Documentation\n---"), 0644)
	os.WriteFile(filepath.Join(docsDir, "_index.md"), []byte("---\ntitle: Documentation\n---"), 0644)
	os.WriteFile(filepath.Join(introDir, "_index.md"), []byte("---\ntitle: Introduction\n---"), 0644)

	issues := ValidateDocsContent(tempDir)

	if len(issues) != 0 {
		t.Errorf("expected no issues for valid docs site, got: %v", issues)
	}
}

func TestValidateDocsContent_MissingDocsIndex(t *testing.T) {
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	os.MkdirAll(contentDir, 0755)

	// Only create root _index.md
	os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte("---\ntitle: Docs\n---"), 0644)

	issues := ValidateDocsContent(tempDir)

	hasDocsIndexIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "docs/_index.md") {
			hasDocsIndexIssue = true
		}
	}

	if !hasDocsIndexIssue {
		t.Error("expected issue for missing docs/_index.md")
	}
}

func TestContentFixer_FixContent_UnknownSiteType(t *testing.T) {
	fixer := NewContentFixer("", SiteType("unknown"))
	content := "---\ntitle: Test\n---\nContent"

	result, changed := fixer.fixContent("content/test.md", content)

	if changed {
		t.Error("should not change content for unknown site type")
	}
	if result != content {
		t.Error("should return original content unchanged")
	}
}

func TestEnsureDocsFrontmatter_AddsTitleIfMissing(t *testing.T) {
	content := `---
draft: false
---

Content without title.`

	result, changed := ensureDocsFrontmatter(content, "doc")

	if !changed {
		t.Error("expected content to be changed")
	}
	if !strings.Contains(result, "title:") {
		t.Error("expected title to be added")
	}
}

func TestEnsureAnankePostFrontmatter_ChangeDraftTrue(t *testing.T) {
	content := `---
title: My Post
draft: true
---

Post content.`

	result, changed := ensureAnankePostFrontmatter(content)

	if !changed {
		t.Error("expected content to be changed")
	}
	if strings.Contains(result, "draft: true") {
		t.Error("draft should be changed to false")
	}
	if !strings.Contains(result, "draft: false") {
		t.Error("draft: false should be present")
	}
}

func TestContentFixer_FixWhitepaperContent(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		content        string
		expectedChecks []string
	}{
		{
			name: "root index",
			path: "content/_index.md",
			content: `---
title: My Whitepaper
---

Content here.`,
			expectedChecks: []string{"draft: false"},
		},
		{
			name: "whitepaper section index",
			path: "content/whitepaper/_index.md",
			content: `---
title: Whitepaper
---

Section index.`,
			expectedChecks: []string{"draft: false"},
		},
		{
			name: "whitepaper section page without weight",
			path: "content/whitepaper/01-abstract.md",
			content: `---
title: Abstract
---

Abstract content.`,
			expectedChecks: []string{"draft: false", "weight:"},
		},
		{
			name: "whitepaper section page with draft true",
			path: "content/whitepaper/03-problem.md",
			content: `---
title: Problem Statement
draft: true
weight: 3
---

Problem content.`,
			expectedChecks: []string{"draft: false", "weight: 3"},
		},
		{
			name: "appendix page",
			path: "content/appendix/glossary.md",
			content: `---
title: Glossary
---

Glossary content.`,
			expectedChecks: []string{"draft: false", "weight:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixer := NewContentFixer("", SiteTypeWhitepaper)
			result, changed := fixer.fixWhitepaperContent(tt.path, tt.content)

			if !changed {
				t.Error("expected content to be changed")
			}

			for _, check := range tt.expectedChecks {
				if !strings.Contains(result, check) {
					t.Errorf("expected %q in result:\n%s", check, result)
				}
			}
		})
	}
}

func TestContentFixer_FixWhitepaperContent_NoChange(t *testing.T) {
	fixer := NewContentFixer("", SiteTypeWhitepaper)

	content := `---
title: Abstract
draft: false
weight: 1
---

Complete content.`

	_, changed := fixer.fixWhitepaperContent("content/whitepaper/01-abstract.md", content)
	if changed {
		t.Error("expected no change for already-valid whitepaper section")
	}
}

func TestEnsureWhitepaperFrontmatter_SectionPage(t *testing.T) {
	content := `---
title: Introduction
---

Intro content.`

	result, changed := ensureWhitepaperFrontmatter(content, "section-page")

	if !changed {
		t.Error("expected content to be changed")
	}
	if !strings.Contains(result, "draft: false") {
		t.Error("draft: false should be added")
	}
	if !strings.Contains(result, "weight:") {
		t.Error("weight should be added")
	}
}

func TestEnsureWhitepaperFrontmatter_Home(t *testing.T) {
	content := `---
title: Home
---

Home content.`

	result, changed := ensureWhitepaperFrontmatter(content, "home")

	if !changed {
		t.Error("expected content to be changed")
	}
	if !strings.Contains(result, "draft: false") {
		t.Error("draft: false should be added")
	}
	// Home pages should NOT get weight
	if strings.Contains(result, "weight:") {
		t.Error("home page should not have weight")
	}
}

func TestValidateWhitepaperContent(t *testing.T) {
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	wpDir := filepath.Join(contentDir, "whitepaper")
	os.MkdirAll(wpDir, 0755)

	// Create valid structure
	os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte("---\ntitle: My Project\n---"), 0644)
	os.WriteFile(filepath.Join(wpDir, "_index.md"), []byte("---\ntitle: Whitepaper\n---"), 0644)
	os.WriteFile(filepath.Join(wpDir, "01-abstract.md"), []byte("---\ntitle: Abstract\nweight: 1\ndraft: false\n---\nAbstract."), 0644)
	os.WriteFile(filepath.Join(wpDir, "02-introduction.md"), []byte("---\ntitle: Introduction\nweight: 2\ndraft: false\n---\nIntro."), 0644)
	os.WriteFile(filepath.Join(wpDir, "03-problem.md"), []byte("---\ntitle: Problem\nweight: 3\ndraft: false\n---\nProblem."), 0644)

	issues := ValidateWhitepaperContent(tempDir)

	if len(issues) != 0 {
		t.Errorf("expected no issues for valid whitepaper, got: %v", issues)
	}
}

func TestValidateWhitepaperContent_MissingFiles(t *testing.T) {
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	os.MkdirAll(contentDir, 0755)

	// Only create root _index.md
	os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte("---\ntitle: Test\n---"), 0644)

	issues := ValidateWhitepaperContent(tempDir)

	hasWpIndexIssue := false
	hasWpDirIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "whitepaper/_index.md") {
			hasWpIndexIssue = true
		}
		if strings.Contains(issue, "whitepaper directory") {
			hasWpDirIssue = true
		}
	}

	if !hasWpIndexIssue {
		t.Error("expected issue for missing whitepaper/_index.md")
	}
	if !hasWpDirIssue {
		t.Error("expected issue for missing whitepaper directory")
	}
}

func TestValidateWhitepaperContent_TooFewSections(t *testing.T) {
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	wpDir := filepath.Join(contentDir, "whitepaper")
	os.MkdirAll(wpDir, 0755)

	os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte("---\ntitle: Test\n---"), 0644)
	os.WriteFile(filepath.Join(wpDir, "_index.md"), []byte("---\ntitle: Whitepaper\n---"), 0644)
	// Only one section
	os.WriteFile(filepath.Join(wpDir, "01-abstract.md"), []byte("---\ntitle: Abstract\nweight: 1\n---\nAbstract."), 0644)

	issues := ValidateWhitepaperContent(tempDir)

	hasSectionCountIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "at least 2 sections") {
			hasSectionCountIssue = true
		}
	}

	if !hasSectionCountIssue {
		t.Error("expected issue for too few sections")
	}
}

func TestValidateWhitepaperContent_MissingWeight(t *testing.T) {
	tempDir := t.TempDir()
	contentDir := filepath.Join(tempDir, "content")
	wpDir := filepath.Join(contentDir, "whitepaper")
	os.MkdirAll(wpDir, 0755)

	os.WriteFile(filepath.Join(contentDir, "_index.md"), []byte("---\ntitle: Test\n---"), 0644)
	os.WriteFile(filepath.Join(wpDir, "_index.md"), []byte("---\ntitle: Whitepaper\n---"), 0644)
	os.WriteFile(filepath.Join(wpDir, "01-abstract.md"), []byte("---\ntitle: Abstract\n---\nAbstract."), 0644)
	os.WriteFile(filepath.Join(wpDir, "02-intro.md"), []byte("---\ntitle: Introduction\nweight: 2\n---\nIntro."), 0644)

	issues := ValidateWhitepaperContent(tempDir)

	hasWeightIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "weight") && strings.Contains(issue, "01-abstract") {
			hasWeightIssue = true
		}
	}

	if !hasWeightIssue {
		t.Error("expected issue for missing weight in 01-abstract.md")
	}
}

func TestEnsureAnankeFrontmatter_AddsDescriptionAndFeaturedImage(t *testing.T) {
	content := `---
title: My Page
---

Page content.`

	result, changed := ensureAnankeFrontmatter(content, "page")

	if !changed {
		t.Error("expected content to be changed")
	}
	if !strings.Contains(result, "description:") {
		t.Error("description should be added")
	}
	if !strings.Contains(result, "featured_image:") {
		t.Error("featured_image should be added")
	}
}
