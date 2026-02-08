package ai

import (
	"strings"
	"testing"
)

func TestBuildSitePlannerPrompt(t *testing.T) {
	result := BuildSitePlannerPrompt(
		"My Blog",
		"blog",
		"A tech blog about programming",
		"developers",
		"professional",
		"https://myblog.com",
	)

	// Verify all fields are included
	if !strings.Contains(result, "My Blog") {
		t.Error("expected site name in prompt")
	}
	if !strings.Contains(result, "blog") {
		t.Error("expected site type in prompt")
	}
	if !strings.Contains(result, "A tech blog about programming") {
		t.Error("expected description in prompt")
	}
	if !strings.Contains(result, "developers") {
		t.Error("expected audience in prompt")
	}
	if !strings.Contains(result, "professional") {
		t.Error("expected tone in prompt")
	}
	if !strings.Contains(result, "https://myblog.com") {
		t.Error("expected base URL in prompt")
	}
	if !strings.Contains(result, "Create the JSON plan now") {
		t.Error("expected instruction in prompt")
	}
}

func TestBuildSinglePageUserPrompt(t *testing.T) {
	tests := []struct {
		name           string
		plan           *SitePlan
		page           *PageSpec
		expectedChecks []string
	}{
		{
			name: "blog page",
			plan: &SitePlan{
				SiteName:    "My Blog",
				SiteType:    SiteTypeBlog,
				Description: "A blog about tech",
				Audience:    "developers",
			},
			page: &PageSpec{
				ID:       "about",
				Path:     "content/about.md",
				PageType: PageTypePage,
				Title:    "About Me",
			},
			expectedChecks: []string{
				"My Blog",
				"blog",
				"content/about.md",
				"About Me",
			},
		},
		{
			name: "blog post",
			plan: &SitePlan{
				SiteName:    "My Blog",
				SiteType:    SiteTypeBlog,
				Description: "A blog about tech",
				Audience:    "developers",
			},
			page: &PageSpec{
				ID:       "post1",
				Path:     "content/posts/welcome/index.md",
				PageType: PageTypePost,
				Title:    "Welcome",
			},
			expectedChecks: []string{
				"content/posts/welcome/index.md",
			},
		},
		{
			name: "docs homepage",
			plan: &SitePlan{
				SiteName:    "My Docs",
				SiteType:    SiteTypeDocs,
				Description: "Documentation site",
				Audience:    "users",
			},
			page: &PageSpec{
				ID:   "home",
				Path: "content/_index.md",
			},
			expectedChecks: []string{
				"docs",
			},
		},
		{
			name: "docs section index",
			plan: &SitePlan{
				SiteName:    "My Docs",
				SiteType:    SiteTypeDocs,
				Description: "Documentation site",
				Audience:    "users",
			},
			page: &PageSpec{
				ID:   "docs-index",
				Path: "content/docs/_index.md",
			},
			expectedChecks: []string{
				"content/docs/_index.md",
			},
		},
		{
			name: "docs subsection index",
			plan: &SitePlan{
				SiteName:    "My Docs",
				SiteType:    SiteTypeDocs,
				Description: "Documentation site",
				Audience:    "users",
			},
			page: &PageSpec{
				ID:   "getting-started",
				Path: "content/docs/intro/_index.md",
			},
			expectedChecks: []string{
				"content/docs/intro/_index.md",
			},
		},
		{
			name: "docs page",
			plan: &SitePlan{
				SiteName:    "My Docs",
				SiteType:    SiteTypeDocs,
				Description: "Documentation site",
				Audience:    "users",
			},
			page: &PageSpec{
				ID:   "installation",
				Path: "content/docs/intro/installation.md",
			},
			expectedChecks: []string{
				"content/docs/intro/installation.md",
			},
		},
		{
			name: "page with keywords and word count",
			plan: &SitePlan{
				SiteName:    "My Blog",
				SiteType:    SiteTypeBlog,
				Description: "A blog",
				Audience:    "readers",
			},
			page: &PageSpec{
				ID:        "about",
				Path:      "content/about.md",
				Keywords:  []string{"about", "me", "developer"},
				WordCount: 500,
			},
			expectedChecks: []string{
				"about, me, developer",
				"~500",
			},
		},
		{
			name: "page with internal links",
			plan: &SitePlan{
				SiteName:    "My Blog",
				SiteType:    SiteTypeBlog,
				Description: "A blog",
				Audience:    "readers",
			},
			page: &PageSpec{
				ID:            "about",
				Path:          "content/about.md",
				InternalLinks: []string{"/contact/", "/posts/"},
			},
			expectedChecks: []string{
				"/contact/, /posts/",
			},
		},
		{
			name: "page with description",
			plan: &SitePlan{
				SiteName:    "My Blog",
				SiteType:    SiteTypeBlog,
				Description: "A blog",
				Audience:    "readers",
			},
			page: &PageSpec{
				ID:          "about",
				Path:        "content/about.md",
				Description: "This page should contain information about me.",
			},
			expectedChecks: []string{
				"PAGE REQUIREMENTS",
				"This page should contain information about me",
			},
		},
		{
			name: "plan with tone",
			plan: &SitePlan{
				SiteName:    "My Blog",
				SiteType:    SiteTypeBlog,
				Description: "A blog",
				Audience:    "readers",
				Tone:        "casual and friendly",
			},
			page: &PageSpec{
				ID:   "about",
				Path: "content/about.md",
			},
			expectedChecks: []string{
				"casual and friendly",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSinglePageUserPrompt(tt.plan, tt.page, nil)

			for _, check := range tt.expectedChecks {
				if !strings.Contains(result, check) {
					t.Errorf("expected %q in prompt\nGot:\n%s", check, result)
				}
			}
		})
	}
}

func TestBuildUserPrompt(t *testing.T) {
	t.Run("with context", func(t *testing.T) {
		result := BuildUserPrompt("Create about page", "Site is about technology")

		if !strings.Contains(result, "Create about page") {
			t.Error("expected instruction in prompt")
		}
		if !strings.Contains(result, "Site is about technology") {
			t.Error("expected context in prompt")
		}
		if !strings.Contains(result, "CONTEXT:") {
			t.Error("expected 'CONTEXT:' label")
		}
	})

	t.Run("without context", func(t *testing.T) {
		result := BuildUserPrompt("Create about page", "")

		if !strings.Contains(result, "Create about page") {
			t.Error("expected instruction in prompt")
		}
		if strings.Contains(result, "CONTEXT:") {
			t.Error("should not include 'CONTEXT:' when empty")
		}
	})
}

func TestBuildUpdatePrompt(t *testing.T) {
	existingContent := `---
title: "About"
---

Some existing content here.`

	result := BuildUpdatePrompt("Add more details about experience", existingContent)

	if !strings.Contains(result, "Add more details about experience") {
		t.Error("expected instruction in prompt")
	}
	if !strings.Contains(result, existingContent) {
		t.Error("expected existing content in prompt")
	}
	if !strings.Contains(result, "---START---") {
		t.Error("expected file markers in prompt")
	}
	if !strings.Contains(result, "---END---") {
		t.Error("expected file markers in prompt")
	}
}

func TestCleanMarkdownFences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no fences",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "markdown fence",
			input:    "```\nHello world\n```",
			expected: "Hello world",
		},
		{
			name:     "markdown fence with language",
			input:    "```markdown\nHello world\n```",
			expected: "Hello world",
		},
		{
			name:     "md fence",
			input:    "```md\nHello world\n```",
			expected: "Hello world",
		},
		{
			name:     "with whitespace",
			input:    "  ```\nHello world\n```  ",
			expected: "Hello world",
		},
		{
			name:     "only prefix",
			input:    "```\nHello world",
			expected: "Hello world",
		},
		{
			name:     "only suffix",
			input:    "Hello world\n```",
			expected: "Hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanMarkdownFences(tt.input)
			if result != tt.expected {
				t.Errorf("CleanMarkdownFences(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanGeneratedContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no changes needed",
			input:    "---\ntitle: Test\ndraft: false\n---\nContent",
			expected: "---\ntitle: Test\ndraft: false\n---\nContent",
		},
		{
			name:     "with markdown fence",
			input:    "```markdown\n---\ntitle: Test\n---\nContent\n```",
			expected: "---\ntitle: Test\n---\nContent",
		},
		{
			name:     "draft true to false",
			input:    "---\ntitle: Test\ndraft: true\n---\nContent",
			expected: "---\ntitle: Test\ndraft: false\n---\nContent",
		},
		{
			name:     "draft:true (no space)",
			input:    "---\ntitle: Test\ndraft:true\n---\nContent",
			expected: "---\ntitle: Test\ndraft: false\n---\nContent",
		},
		{
			name:     "draft: True (capitalized)",
			input:    "---\ntitle: Test\ndraft: True\n---\nContent",
			expected: "---\ntitle: Test\ndraft: false\n---\nContent",
		},
		{
			name:     "draft:True (capitalized, no space)",
			input:    "---\ntitle: Test\ndraft:True\n---\nContent",
			expected: "---\ntitle: Test\ndraft: false\n---\nContent",
		},
		{
			name:     "draft: TRUE (uppercase)",
			input:    "---\ntitle: Test\ndraft: TRUE\n---\nContent",
			expected: "---\ntitle: Test\ndraft: false\n---\nContent",
		},
		{
			name:     "combined: fence and draft",
			input:    "```markdown\n---\ntitle: Test\ndraft: true\n---\nContent\n```",
			expected: "---\ntitle: Test\ndraft: false\n---\nContent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanGeneratedContent(tt.input)
			if result != tt.expected {
				t.Errorf("CleanGeneratedContent():\nGot:\n%s\nWant:\n%s", result, tt.expected)
			}
		})
	}
}

func TestSystemPrompts(t *testing.T) {
	// Verify system prompts are non-empty and contain expected content
	prompts := map[string]struct {
		prompt   string
		contains []string
	}{
		"SystemPromptSitePlanner": {
			prompt: SystemPromptSitePlanner,
			contains: []string{
				"SITE ARCHITECT",
				"JSON",
				"pages",
			},
		},
		"SystemPromptContentUpdate": {
			prompt: SystemPromptContentUpdate,
			contains: []string{
				"UPDATE RULES",
				"frontmatter",
			},
		},
	}

	for name, tt := range prompts {
		t.Run(name, func(t *testing.T) {
			if len(tt.prompt) == 0 {
				t.Error("prompt is empty")
			}
			for _, expected := range tt.contains {
				if !strings.Contains(tt.prompt, expected) {
					t.Errorf("expected %q in prompt", expected)
				}
			}
		})
	}
}

func TestPromptComponents(t *testing.T) {
	// Test that core prompt components contain expected content
	tests := []struct {
		name     string
		prompt   string
		contains []string
	}{
		{
			name:   "OutputFormatRules",
			prompt: OutputFormatRules,
			contains: []string{
				"OUTPUT FORMAT",
				"frontmatter",
				"---",
			},
		},
		{
			name:   "YAMLSyntaxRules",
			prompt: YAMLSyntaxRules,
			contains: []string{
				"YAML",
				"double quotes",
			},
		},
		{
			name:   "ContentQualityRules",
			prompt: ContentQualityRules,
			contains: []string{
				"Hook",
				"call-to-action",
			},
		},
		{
			name:   "HugoRules",
			prompt: HugoRules,
			contains: []string{
				"title",
				"draft",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expected := range tt.contains {
				if !strings.Contains(tt.prompt, expected) {
					t.Errorf("expected %q in %s", expected, tt.name)
				}
			}
		})
	}
}
