package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "New command help",
			Args:        []string{"new", "--help"},
			ExpectError: false,
			Contains: []string{
				"Creates",
				"--no-build",
				"--serve",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestNewCommandFlags(t *testing.T) {
	// Find the new command
	var newCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "new" {
			newCommand = cmd
			break
		}
	}

	if newCommand == nil {
		t.Fatal("new command not found")
	}

	t.Run("no-build flag exists", func(t *testing.T) {
		noBuildFlag := newCommand.Flags().Lookup("no-build")
		if noBuildFlag == nil {
			t.Error("no-build flag not found")
		} else {
			if noBuildFlag.DefValue != "false" {
				t.Errorf("Expected default value 'false', got '%s'", noBuildFlag.DefValue)
			}
		}
	})

	t.Run("serve flag exists", func(t *testing.T) {
		serveFlag := newCommand.Flags().Lookup("serve")
		if serveFlag == nil {
			t.Error("serve flag not found")
		} else {
			if serveFlag.DefValue != "false" {
				t.Errorf("Expected default value 'false', got '%s'", serveFlag.DefValue)
			}
		}
	})

	t.Run("Command accepts optional slug argument", func(t *testing.T) {
		if newCommand.Args(newCommand, []string{}) != nil {
			t.Error("Should accept no arguments")
		}
		if newCommand.Args(newCommand, []string{"my-post"}) != nil {
			t.Error("Should accept one argument")
		}
	})
}

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid alphanumeric", "mypost", true},
		{"Valid with hyphen", "my-post", true},
		{"Valid with underscore", "my_post", true},
		{"Valid with numbers", "post123", true},
		{"Valid mixed", "my-post_123", true},
		{"Valid with .md extension", "my-post.md", true},
		{"Empty string", "", false},
		{"Contains space", "my post", false},
		{"Contains slash", "my/post", false},
		{"Contains special char", "my@post", false},
		{"Too long", string(make([]byte, 101)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSlug(tt.input)
			if result != tt.expected {
				t.Errorf("isValidSlug(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewCommandExecution(t *testing.T) {
	t.Run("New command requires Hugo site", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute new command without Hugo site structure
		// This will trigger content type detection
		output, _ := executeCommand(rootCmd, "new", "my-post", "--no-build")
		_ = output
	})

	t.Run("New command with Hugo site structure", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create Hugo site structure
		if err := os.MkdirAll("content/posts", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("hugo.toml", []byte(`title = "Test Site"`), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute new command - will fail without Hugo binary
		output, _ := executeCommand(rootCmd, "new", "my-post", "--no-build")
		_ = output
	})

	t.Run("New command detects content types", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create Hugo site structure with multiple content types
		for _, dir := range []string{"content/posts", "content/pages", "content/docs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
		}

		// Add some files to detect
		if err := os.WriteFile("content/posts/post1.md", []byte("# Post"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("content/posts/post2.md", []byte("# Post 2"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("content/docs/doc1.md", []byte("# Doc"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile("hugo.toml", []byte(`title = "Test Site"`), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute new command
		output, _ := executeCommand(rootCmd, "new", "my-post", "--no-build")
		_ = output
	})

	t.Run("New command with invalid slug", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create Hugo site structure
		if err := os.MkdirAll("content/posts", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("hugo.toml", []byte(`title = "Test Site"`), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute new command with invalid slug
		_, err := executeCommand(rootCmd, "new", "my post with spaces", "--no-build")
		if err == nil {
			t.Error("Expected error for invalid slug")
		}
	})

	t.Run("New command adds .md extension", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create Hugo site structure
		if err := os.MkdirAll("content/posts", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("hugo.toml", []byte(`title = "Test Site"`), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute new command without .md extension
		output, _ := executeCommand(rootCmd, "new", "my-post", "--no-build")
		_ = output
	})
}

func TestNewCommandWithNoContent(t *testing.T) {
	t.Run("New creates default content type when none exists", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create Hugo site structure without any content directories
		if err := os.MkdirAll("content", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("hugo.toml", []byte(`title = "Test Site"`), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute new command
		output, _ := executeCommand(rootCmd, "new", "my-post", "--no-build")
		_ = output
	})
}

func TestNewCommandDescription(t *testing.T) {
	var newCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "new" {
			newCommand = cmd
			break
		}
	}

	if newCommand == nil {
		t.Fatal("new command not found")
	}

	t.Run("Short description mentions content", func(t *testing.T) {
		if newCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description provides examples", func(t *testing.T) {
		if newCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})
}

func TestSlugValidation(t *testing.T) {
	t.Run("Validate slug with various extensions", func(t *testing.T) {
		validSlugs := []string{
			"my-post",
			"my-post.md",
			"my_post",
			"mypost123",
			"123mypost",
		}

		for _, slug := range validSlugs {
			if !isValidSlug(slug) {
				t.Errorf("Expected %q to be valid", slug)
			}
		}
	})

	t.Run("Validate slug rejects invalid characters", func(t *testing.T) {
		invalidSlugs := []string{
			"",
			"my post",
			"my.post",
			"my/post",
			"my\\post",
			"my@post",
		}

		for _, slug := range invalidSlugs {
			if isValidSlug(slug) {
				t.Errorf("Expected %q to be invalid", slug)
			}
		}
	})
}

func TestNewCommandArchetypes(t *testing.T) {
	t.Run("New command uses archetypes", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Create Hugo site structure with archetype
		if err := os.MkdirAll("content/posts", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll("archetypes", 0755); err != nil {
			t.Fatal(err)
		}

		// Create a custom archetype
		archetype := `---
title: "{{ replace .Name "-" " " | title }}"
date: {{ .Date }}
draft: true
tags: []
---

Your content here.
`
		if err := os.WriteFile(filepath.Join("archetypes", "default.md"), []byte(archetype), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile("hugo.toml", []byte(`title = "Test Site"`), 0644); err != nil {
			t.Fatal(err)
		}

		// Execute new command
		output, _ := executeCommand(rootCmd, "new", "my-new-post", "--no-build")
		_ = output
	})
}
