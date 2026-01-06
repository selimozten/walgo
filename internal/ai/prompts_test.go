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
		expectedTheme  string
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
			expectedTheme: "Ananke",
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
			expectedTheme: "Ananke",
			expectedChecks: []string{
				"content/posts/welcome/index.md",
				"Ananke blog post",
			},
		},
		{
			name: "business page",
			plan: &SitePlan{
				SiteName:    "My Business",
				SiteType:    SiteTypeBusiness,
				Description: "A business site",
				Audience:    "customers",
			},
			page: &PageSpec{
				ID:       "home",
				Path:     "content/_index.md",
				PageType: PageTypeHome,
			},
			expectedTheme: "Ananke",
			expectedChecks: []string{
				"business",
				"Ananke",
			},
		},
		{
			name: "business service page",
			plan: &SitePlan{
				SiteName:    "My Business",
				SiteType:    SiteTypeBusiness,
				Description: "A business site",
				Audience:    "customers",
			},
			page: &PageSpec{
				ID:   "service1",
				Path: "content/services/consulting.md",
			},
			expectedTheme: "Ananke",
			expectedChecks: []string{
				"Ananke service",
				"date (ISO 8601)",
			},
		},
		{
			name: "portfolio page",
			plan: &SitePlan{
				SiteName:    "My Portfolio",
				SiteType:    SiteTypePortfolio,
				Description: "A portfolio site",
				Audience:    "employers",
			},
			page: &PageSpec{
				ID:   "home",
				Path: "content/_index.md",
			},
			expectedTheme: "Ananke",
			expectedChecks: []string{
				"Ananke page",
			},
		},
		{
			name: "portfolio project",
			plan: &SitePlan{
				SiteName:    "My Portfolio",
				SiteType:    SiteTypePortfolio,
				Description: "A portfolio site",
				Audience:    "employers",
			},
			page: &PageSpec{
				ID:   "project1",
				Path: "content/projects/my-app.md",
			},
			expectedTheme: "Ananke",
			expectedChecks: []string{
				"Ananke project/portfolio entry",
				"date (ISO 8601)",
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
			expectedTheme: "Book",
			expectedChecks: []string{
				"Hugo Book homepage",
				"documentation landing page",
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
			expectedTheme: "Book",
			expectedChecks: []string{
				"Hugo Book docs section index",
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
			expectedTheme: "Book",
			expectedChecks: []string{
				"Hugo Book section index",
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
			expectedTheme: "Book",
			expectedChecks: []string{
				"Hugo Book documentation page",
				"code examples",
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
			expectedTheme: "Ananke",
			expectedChecks: []string{
				"about, me, developer",
				"approximately 500 words",
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
			expectedTheme: "Ananke",
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
			expectedTheme: "Ananke",
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
			expectedTheme: "Ananke",
			expectedChecks: []string{
				"casual and friendly",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSinglePageUserPrompt(tt.plan, tt.page)

			if !strings.Contains(result, tt.expectedTheme) {
				t.Errorf("expected theme %s in prompt", tt.expectedTheme)
			}

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
		if !strings.Contains(result, "Additional Context") {
			t.Error("expected 'Additional Context' label")
		}
	})

	t.Run("without context", func(t *testing.T) {
		result := BuildUserPrompt("Create about page", "")

		if !strings.Contains(result, "Create about page") {
			t.Error("expected instruction in prompt")
		}
		if strings.Contains(result, "Additional Context") {
			t.Error("should not include 'Additional Context' when empty")
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
	if !strings.Contains(result, "---START OF FILE---") {
		t.Error("expected file markers in prompt")
	}
	if !strings.Contains(result, "---END OF FILE---") {
		t.Error("expected file markers in prompt")
	}
}

func TestBuildSiteGenerationPrompt(t *testing.T) {
	result := BuildSiteGenerationPrompt(
		"My Blog",
		"blog",
		"Tech blog",
		"developers",
		"about, contact, posts",
	)

	if !strings.Contains(result, "My Blog") {
		t.Error("expected site name in prompt")
	}
	if !strings.Contains(result, "blog") {
		t.Error("expected site type in prompt")
	}
	if !strings.Contains(result, "Tech blog") {
		t.Error("expected description in prompt")
	}
	if !strings.Contains(result, "developers") {
		t.Error("expected audience in prompt")
	}
	if !strings.Contains(result, "about, contact, posts") {
		t.Error("expected features in prompt")
	}
	if !strings.Contains(result, "===FILE:") {
		t.Error("expected file format instruction in prompt")
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
				"SITE PLANNER",
				"JSON",
				"pages",
			},
		},
		"SystemPromptSinglePageGenerator": {
			prompt: SystemPromptSinglePageGenerator,
			contains: []string{
				"CONTENT GENERATOR",
				"Markdown",
				"frontmatter",
			},
		},
		"SystemPromptContentGeneration": {
			prompt: SystemPromptContentGeneration,
			contains: []string{
				"Hugo",
				"draft: false",
			},
		},
		"SystemPromptContentUpdate": {
			prompt: SystemPromptContentUpdate,
			contains: []string{
				"update",
				"frontmatter",
			},
		},
		"SystemPromptBlogPost": {
			prompt: SystemPromptBlogPost,
			contains: []string{
				"BLOG POST",
				"title",
			},
		},
		"SystemPromptPageGeneration": {
			prompt: SystemPromptPageGeneration,
			contains: []string{
				"HUGO PAGES",
				"title",
			},
		},
		"SystemPromptSiteGeneration": {
			prompt: SystemPromptSiteGeneration,
			contains: []string{
				"Hugo site architect",
				"===FILE:",
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

func TestSystemPromptBaseRules(t *testing.T) {
	// The base rules should be included in prompts that use it
	baseRulesContent := []string{
		"GLOBAL RULES",
		"Hugo static site",
		"Markdown",
		"frontmatter",
		"draft: false",
	}

	for _, content := range baseRulesContent {
		if !strings.Contains(systemPromptBaseRules, content) {
			t.Errorf("expected %q in systemPromptBaseRules", content)
		}
	}
}
