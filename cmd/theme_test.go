package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestThemeCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Theme command help",
			Args:        []string{"theme", "--help"},
			ExpectError: false,
			Contains: []string{
				"Install, list, and manage Hugo themes",
				"install",
				"list",
				"new",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestThemeCommandDescription(t *testing.T) {
	themeCommand := findCommand(rootCmd, "theme")
	if themeCommand == nil {
		t.Fatal("theme command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if themeCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if themeCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})
}

// --- Theme install subcommand ---

func TestThemeInstallCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Theme install help",
			Args:        []string{"theme", "install", "--help"},
			ExpectError: false,
			Contains: []string{
				"Install a Hugo theme from a GitHub repository URL",
				"github-url",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestThemeInstallCommandDescription(t *testing.T) {
	installCmd := findCommand(rootCmd, "theme", "install")
	if installCmd == nil {
		t.Fatal("theme install command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if installCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if installCmd.Long == "" {
			t.Error("Long description is empty")
		}
	})

	t.Run("Long description mentions GitHub", func(t *testing.T) {
		if !containsStr(installCmd.Long, "GitHub") {
			t.Error("Long description should mention GitHub")
		}
	})
}

func TestThemeInstallCommandArgsValidation(t *testing.T) {
	installCmd := findCommand(rootCmd, "theme", "install")
	if installCmd == nil {
		t.Fatal("theme install command not found")
	}

	t.Run("No arguments returns error", func(t *testing.T) {
		err := installCmd.Args(installCmd, []string{})
		if err == nil {
			t.Error("Expected error for no arguments")
		}
	})

	t.Run("One argument is valid", func(t *testing.T) {
		err := installCmd.Args(installCmd, []string{"https://github.com/user/theme"})
		if err != nil {
			t.Error("Should accept one argument")
		}
	})

	t.Run("Two arguments returns error", func(t *testing.T) {
		err := installCmd.Args(installCmd, []string{"url1", "url2"})
		if err == nil {
			t.Error("Expected error for too many arguments")
		}
	})
}

// --- Theme list subcommand ---

func TestThemeListCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Theme list help",
			Args:        []string{"theme", "list", "--help"},
			ExpectError: false,
			Contains: []string{
				"List all themes currently installed",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestThemeListCommandDescription(t *testing.T) {
	listCmd := findCommand(rootCmd, "theme", "list")
	if listCmd == nil {
		t.Fatal("theme list command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if listCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if listCmd.Long == "" {
			t.Error("Long description is empty")
		}
	})
}

// --- Theme new subcommand ---

func TestThemeNewCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Theme new help",
			Args:        []string{"theme", "new", "--help"},
			ExpectError: false,
			Contains: []string{
				"Create a new Hugo theme using Hugo",
				"hugo new theme",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestThemeNewCommandDescription(t *testing.T) {
	newCmd := findCommand(rootCmd, "theme", "new")
	if newCmd == nil {
		t.Fatal("theme new command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if newCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if newCmd.Long == "" {
			t.Error("Long description is empty")
		}
	})

	t.Run("Long description mentions hugo new theme", func(t *testing.T) {
		if !containsStr(newCmd.Long, "hugo new theme") {
			t.Error("Long description should mention 'hugo new theme'")
		}
	})
}

func TestThemeNewCommandArgsValidation(t *testing.T) {
	newCmd := findCommand(rootCmd, "theme", "new")
	if newCmd == nil {
		t.Fatal("theme new command not found")
	}

	t.Run("No arguments returns error", func(t *testing.T) {
		err := newCmd.Args(newCmd, []string{})
		if err == nil {
			t.Error("Expected error for no arguments")
		}
	})

	t.Run("One argument is valid", func(t *testing.T) {
		err := newCmd.Args(newCmd, []string{"my-theme"})
		if err != nil {
			t.Error("Should accept one argument")
		}
	})

	t.Run("Two arguments returns error", func(t *testing.T) {
		err := newCmd.Args(newCmd, []string{"theme1", "theme2"})
		if err == nil {
			t.Error("Expected error for too many arguments")
		}
	})
}

// --- Theme subcommand registration ---

func TestThemeSubcommandsRegistered(t *testing.T) {
	themeCommand := findCommand(rootCmd, "theme")
	if themeCommand == nil {
		t.Fatal("theme command not found")
	}

	expectedSubcommands := []string{"install", "list", "new"}

	subcommands := make(map[string]bool)
	for _, child := range themeCommand.Commands() {
		subcommands[child.Name()] = true
	}

	for _, expected := range expectedSubcommands {
		t.Run("Subcommand "+expected+" is registered", func(t *testing.T) {
			if !subcommands[expected] {
				t.Errorf("Expected subcommand '%s' to be registered under 'theme'", expected)
			}
		})
	}
}

func TestThemeSubcommandsHaveRunE(t *testing.T) {
	themeCommand := findCommand(rootCmd, "theme")
	if themeCommand == nil {
		t.Fatal("theme command not found")
	}

	for _, child := range themeCommand.Commands() {
		t.Run(child.Name()+" has RunE", func(t *testing.T) {
			if child.RunE == nil {
				t.Errorf("Subcommand '%s' should have RunE set", child.Name())
			}
		})
	}
}

// --- isHugoSite helper function ---

func TestIsHugoSite(t *testing.T) {
	t.Run("Directory with hugo.toml", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tempDir, "hugo.toml"), []byte(`title = "Test"`), 0644); err != nil {
			t.Fatal(err)
		}
		if !isHugoSite(tempDir) {
			t.Error("Should detect hugo.toml as a Hugo site")
		}
	})

	t.Run("Directory with config.toml", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tempDir, "config.toml"), []byte(`title = "Test"`), 0644); err != nil {
			t.Fatal(err)
		}
		if !isHugoSite(tempDir) {
			t.Error("Should detect config.toml as a Hugo site")
		}
	})

	t.Run("Directory with hugo.yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tempDir, "hugo.yaml"), []byte(`title: Test`), 0644); err != nil {
			t.Fatal(err)
		}
		if !isHugoSite(tempDir) {
			t.Error("Should detect hugo.yaml as a Hugo site")
		}
	})

	t.Run("Directory with config.yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(`title: Test`), 0644); err != nil {
			t.Fatal(err)
		}
		if !isHugoSite(tempDir) {
			t.Error("Should detect config.yaml as a Hugo site")
		}
	})

	t.Run("Empty directory is not a Hugo site", func(t *testing.T) {
		tempDir := t.TempDir()
		if isHugoSite(tempDir) {
			t.Error("Empty directory should not be detected as a Hugo site")
		}
	})

	t.Run("Directory with unrelated files", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tempDir, "readme.md"), []byte("# Readme"), 0644); err != nil {
			t.Fatal(err)
		}
		if isHugoSite(tempDir) {
			t.Error("Directory without Hugo config files should not be detected as a Hugo site")
		}
	})

	t.Run("Non-existent directory", func(t *testing.T) {
		if isHugoSite("/nonexistent/path/that/does/not/exist") {
			t.Error("Non-existent directory should not be detected as a Hugo site")
		}
	})
}

// --- Theme install execution tests ---

func TestThemeInstallCommandExecution(t *testing.T) {
	t.Run("Install in non-Hugo directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute theme install in non-Hugo directory
		_, err := executeCommand(rootCmd, "theme", "install", "https://github.com/user/theme")
		if err == nil {
			t.Error("Expected error when not in a Hugo site directory")
		}
	})
}

// --- Theme new execution tests ---

func TestThemeNewCommandExecution(t *testing.T) {
	t.Run("New theme in non-Hugo directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		// Execute theme new in non-Hugo directory
		_, err := executeCommand(rootCmd, "theme", "new", "my-theme")
		if err == nil {
			t.Error("Expected error when not in a Hugo site directory")
		}
	})
}

// --- Theme command via executeCommand with help ---

func TestThemeCommandViaRootCmd(t *testing.T) {
	// Verify theme is accessible through root command
	var themeCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "theme" {
			themeCommand = cmd
			break
		}
	}

	if themeCommand == nil {
		t.Fatal("theme command not found in root command")
	}

	t.Run("Theme is registered under root", func(t *testing.T) {
		if themeCommand.Parent() != rootCmd {
			t.Error("Theme command should be a direct child of root")
		}
	})
}
