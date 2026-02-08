package obsidian

import (
	"testing"
)

func TestGenerateTitle(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "Simple filename",
			filePath: "my-post.md",
			expected: "My Post",
		},
		{
			name:     "Filename with path",
			filePath: "posts/my-awesome-post.md",
			expected: "My Awesome Post",
		},
		{
			name:     "Filename with underscores",
			filePath: "my_cool_article.md",
			expected: "My Cool Article",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateTitle(tt.filePath)
			if result != tt.expected {
				t.Errorf("generateTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsAttachment(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "PNG image",
			path:     "image.png",
			expected: true,
		},
		{
			name:     "JPG image",
			path:     "photo.jpg",
			expected: true,
		},
		{
			name:     "PDF document",
			path:     "document.pdf",
			expected: true,
		},
		{
			name:     "Markdown file",
			path:     "article.md",
			expected: false,
		},
		{
			name:     "Text file",
			path:     "notes.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAttachment(tt.path)
			if result != tt.expected {
				t.Errorf("isAttachment() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnsureFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		format   string
		hasfront bool
	}{
		{
			name:     "Content without frontmatter YAML",
			content:  "# My Article\n\nThis is content.",
			format:   "yaml",
			hasfront: false,
		},
		{
			name:     "Content with existing frontmatter",
			content:  "---\ntitle: Test\n---\n\n# Content",
			format:   "yaml",
			hasfront: true,
		},
		{
			name:     "Content without frontmatter TOML",
			content:  "# My Article\n\nThis is content.",
			format:   "toml",
			hasfront: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureFrontmatter(tt.content, "test.md", tt.format)

			if tt.hasfront {
				// Should return unchanged if frontmatter exists
				if result != tt.content {
					t.Errorf("ensureFrontmatter() should not modify content with existing frontmatter")
				}
			} else {
				// Should add frontmatter
				if result == tt.content {
					t.Errorf("ensureFrontmatter() should add frontmatter to content without it")
				}
				// Check format-specific delimiters
				switch tt.format {
				case "yaml":
					if result[0:3] != "---" {
						t.Errorf("ensureFrontmatter() should add YAML frontmatter starting with ---")
					}
				case "toml":
					if result[0:3] != "+++" {
						t.Errorf("ensureFrontmatter() should add TOML frontmatter starting with +++")
					}
				}
			}
		})
	}
}
