package cmd

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// findCommand searches rootCmd's subcommands (and nested subcommands) for the
// given path (e.g. "projects", "projects show").
func findCommand(root *cobra.Command, names ...string) *cobra.Command {
	cmd := root
	for _, name := range names {
		var found *cobra.Command
		for _, child := range cmd.Commands() {
			if child.Name() == name {
				found = child
				break
			}
		}
		if found == nil {
			return nil
		}
		cmd = found
	}
	return cmd
}

// --- Projects root command ---

func TestProjectsCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Projects command help",
			Args:        []string{"projects", "--help"},
			ExpectError: false,
			Contains: []string{
				"View, manage, and redeploy your Walrus site projects",
				"list",
				"show",
				"edit",
				"update",
				"delete",
				"archive",
			},
		},
		{
			Name:        "Projects command with invalid flag",
			Args:        []string{"projects", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestProjectsCommandDescription(t *testing.T) {
	projectsCommand := findCommand(rootCmd, "projects")
	if projectsCommand == nil {
		t.Fatal("projects command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if projectsCommand.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if projectsCommand.Long == "" {
			t.Error("Long description is empty")
		}
	})

	t.Run("Long description mentions examples", func(t *testing.T) {
		if !containsStr(projectsCommand.Long, "Examples") {
			t.Error("Long description should mention examples")
		}
	})

	t.Run("Long description mentions Project Identification", func(t *testing.T) {
		if !containsStr(projectsCommand.Long, "Project Identification") {
			t.Error("Long description should mention Project Identification")
		}
	})
}

func TestProjectsCommandFlags(t *testing.T) {
	projectsCommand := findCommand(rootCmd, "projects")
	if projectsCommand == nil {
		t.Fatal("projects command not found")
	}

	flagTests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"network flag", "network", "n", ""},
		{"status flag", "status", "s", ""},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := projectsCommand.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("Expected shorthand '%s', got '%s'", tt.shorthand, flag.Shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Expected default value '%s', got '%s'", tt.defValue, flag.DefValue)
			}
		})
	}
}

// --- Projects list subcommand ---

func TestProjectsListCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Projects list help",
			Args:        []string{"projects", "list", "--help"},
			ExpectError: false,
			Contains: []string{
				"List all projects",
				"--network",
				"--status",
			},
		},
		{
			Name:        "Projects list with invalid flag",
			Args:        []string{"projects", "list", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestProjectsListCommandDescription(t *testing.T) {
	listCmd := findCommand(rootCmd, "projects", "list")
	if listCmd == nil {
		t.Fatal("projects list command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if listCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})
}

func TestProjectsListCommandFlags(t *testing.T) {
	listCmd := findCommand(rootCmd, "projects", "list")
	if listCmd == nil {
		t.Fatal("projects list command not found")
	}

	flagTests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"network flag", "network", "n", ""},
		{"status flag", "status", "s", ""},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := listCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("Expected shorthand '%s', got '%s'", tt.shorthand, flag.Shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Expected default value '%s', got '%s'", tt.defValue, flag.DefValue)
			}
		})
	}
}

// --- Projects show subcommand ---

func TestProjectsShowCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Projects show help",
			Args:        []string{"projects", "show", "--help"},
			ExpectError: false,
			Contains: []string{
				"Show detailed information about a project",
				"--id",
				"--name",
			},
		},
		{
			Name:        "Projects show with invalid flag",
			Args:        []string{"projects", "show", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestProjectsShowCommandDescription(t *testing.T) {
	showCmd := findCommand(rootCmd, "projects", "show")
	if showCmd == nil {
		t.Fatal("projects show command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if showCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if showCmd.Long == "" {
			t.Error("Long description is empty")
		}
	})

	t.Run("Long description mentions Project Identification", func(t *testing.T) {
		if !containsStr(showCmd.Long, "Project Identification") {
			t.Error("Long description should mention Project Identification")
		}
	})
}

func TestProjectsShowCommandFlags(t *testing.T) {
	showCmd := findCommand(rootCmd, "projects", "show")
	if showCmd == nil {
		t.Fatal("projects show command not found")
	}

	flagTests := []struct {
		name     string
		flagName string
		defValue string
	}{
		{"id flag", "id", "0"},
		{"name flag", "name", ""},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := showCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Expected default value '%s', got '%s'", tt.defValue, flag.DefValue)
			}
		})
	}
}

func TestProjectsShowCommandArgsValidation(t *testing.T) {
	showCmd := findCommand(rootCmd, "projects", "show")
	if showCmd == nil {
		t.Fatal("projects show command not found")
	}

	t.Run("Accepts no arguments", func(t *testing.T) {
		if showCmd.Args(showCmd, []string{}) != nil {
			t.Error("Should accept no arguments (uses flags)")
		}
	})

	t.Run("Accepts one argument", func(t *testing.T) {
		if showCmd.Args(showCmd, []string{"myproject"}) != nil {
			t.Error("Should accept one positional argument")
		}
	})

	t.Run("Rejects two arguments", func(t *testing.T) {
		if err := showCmd.Args(showCmd, []string{"proj1", "proj2"}); err == nil {
			t.Error("Should reject more than one argument")
		}
	})
}

// --- Projects edit subcommand ---

func TestProjectsEditCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Projects edit help",
			Args:        []string{"projects", "edit", "--help"},
			ExpectError: false,
			Contains: []string{
				"Edit project metadata",
				"--id",
				"--name",
				"--new-name",
				"--category",
				"--description",
				"--image-url",
				"--suins",
			},
		},
		{
			Name:        "Projects edit with invalid flag",
			Args:        []string{"projects", "edit", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestProjectsEditCommandDescription(t *testing.T) {
	editCmd := findCommand(rootCmd, "projects", "edit")
	if editCmd == nil {
		t.Fatal("projects edit command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if editCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if editCmd.Long == "" {
			t.Error("Long description is empty")
		}
	})

	t.Run("Long description mentions ws-resources.json", func(t *testing.T) {
		if !containsStr(editCmd.Long, "ws-resources.json") {
			t.Error("Long description should mention ws-resources.json")
		}
	})

	t.Run("Long description mentions --new-name", func(t *testing.T) {
		if !containsStr(editCmd.Long, "--new-name") {
			t.Error("Long description should mention --new-name")
		}
	})
}

func TestProjectsEditCommandFlags(t *testing.T) {
	editCmd := findCommand(rootCmd, "projects", "edit")
	if editCmd == nil {
		t.Fatal("projects edit command not found")
	}

	flagTests := []struct {
		name     string
		flagName string
		defValue string
	}{
		{"id flag", "id", "0"},
		{"name flag", "name", ""},
		{"new-name flag", "new-name", ""},
		{"category flag", "category", ""},
		{"description flag", "description", ""},
		{"image-url flag", "image-url", ""},
		{"suins flag", "suins", ""},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := editCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Expected default value '%s', got '%s'", tt.defValue, flag.DefValue)
			}
		})
	}
}

func TestProjectsEditCommandArgsValidation(t *testing.T) {
	editCmd := findCommand(rootCmd, "projects", "edit")
	if editCmd == nil {
		t.Fatal("projects edit command not found")
	}

	t.Run("Accepts no arguments", func(t *testing.T) {
		if editCmd.Args(editCmd, []string{}) != nil {
			t.Error("Should accept no arguments (uses flags)")
		}
	})

	t.Run("Accepts one argument", func(t *testing.T) {
		if editCmd.Args(editCmd, []string{"myproject"}) != nil {
			t.Error("Should accept one positional argument")
		}
	})

	t.Run("Rejects two arguments", func(t *testing.T) {
		if err := editCmd.Args(editCmd, []string{"proj1", "proj2"}); err == nil {
			t.Error("Should reject more than one argument")
		}
	})
}

// --- Projects delete subcommand ---

func TestProjectsDeleteCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Projects delete help",
			Args:        []string{"projects", "delete", "--help"},
			ExpectError: false,
			Contains: []string{
				"Delete a project",
				"--id",
				"--name",
			},
		},
		{
			Name:        "Projects delete with invalid flag",
			Args:        []string{"projects", "delete", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestProjectsDeleteCommandDescription(t *testing.T) {
	deleteCmd := findCommand(rootCmd, "projects", "delete")
	if deleteCmd == nil {
		t.Fatal("projects delete command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if deleteCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if deleteCmd.Long == "" {
			t.Error("Long description is empty")
		}
	})

	t.Run("Long description mentions Walrus blockchain", func(t *testing.T) {
		if !containsStr(deleteCmd.Long, "Walrus blockchain") {
			t.Error("Long description should mention Walrus blockchain")
		}
	})
}

func TestProjectsDeleteCommandFlags(t *testing.T) {
	deleteCmd := findCommand(rootCmd, "projects", "delete")
	if deleteCmd == nil {
		t.Fatal("projects delete command not found")
	}

	flagTests := []struct {
		name     string
		flagName string
		defValue string
	}{
		{"id flag", "id", "0"},
		{"name flag", "name", ""},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := deleteCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Expected default value '%s', got '%s'", tt.defValue, flag.DefValue)
			}
		})
	}
}

func TestProjectsDeleteCommandArgsValidation(t *testing.T) {
	deleteCmd := findCommand(rootCmd, "projects", "delete")
	if deleteCmd == nil {
		t.Fatal("projects delete command not found")
	}

	t.Run("Accepts no arguments", func(t *testing.T) {
		if deleteCmd.Args(deleteCmd, []string{}) != nil {
			t.Error("Should accept no arguments (uses flags)")
		}
	})

	t.Run("Accepts one argument", func(t *testing.T) {
		if deleteCmd.Args(deleteCmd, []string{"myproject"}) != nil {
			t.Error("Should accept one positional argument")
		}
	})

	t.Run("Rejects two arguments", func(t *testing.T) {
		if err := deleteCmd.Args(deleteCmd, []string{"proj1", "proj2"}); err == nil {
			t.Error("Should reject more than one argument")
		}
	})
}

// --- Projects archive subcommand ---

func TestProjectsArchiveCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Projects archive help",
			Args:        []string{"projects", "archive", "--help"},
			ExpectError: false,
			Contains: []string{
				"Archive a project",
				"--id",
				"--name",
			},
		},
		{
			Name:        "Projects archive with invalid flag",
			Args:        []string{"projects", "archive", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestProjectsArchiveCommandDescription(t *testing.T) {
	archiveCmd := findCommand(rootCmd, "projects", "archive")
	if archiveCmd == nil {
		t.Fatal("projects archive command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if archiveCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if archiveCmd.Long == "" {
			t.Error("Long description is empty")
		}
	})

	t.Run("Long description mentions Project Identification", func(t *testing.T) {
		if !containsStr(archiveCmd.Long, "Project Identification") {
			t.Error("Long description should mention Project Identification")
		}
	})
}

func TestProjectsArchiveCommandFlags(t *testing.T) {
	archiveCmd := findCommand(rootCmd, "projects", "archive")
	if archiveCmd == nil {
		t.Fatal("projects archive command not found")
	}

	flagTests := []struct {
		name     string
		flagName string
		defValue string
	}{
		{"id flag", "id", "0"},
		{"name flag", "name", ""},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := archiveCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Expected default value '%s', got '%s'", tt.defValue, flag.DefValue)
			}
		})
	}
}

func TestProjectsArchiveCommandArgsValidation(t *testing.T) {
	archiveCmd := findCommand(rootCmd, "projects", "archive")
	if archiveCmd == nil {
		t.Fatal("projects archive command not found")
	}

	t.Run("Accepts no arguments", func(t *testing.T) {
		if archiveCmd.Args(archiveCmd, []string{}) != nil {
			t.Error("Should accept no arguments (uses flags)")
		}
	})

	t.Run("Accepts one argument", func(t *testing.T) {
		if archiveCmd.Args(archiveCmd, []string{"myproject"}) != nil {
			t.Error("Should accept one positional argument")
		}
	})

	t.Run("Rejects two arguments", func(t *testing.T) {
		if err := archiveCmd.Args(archiveCmd, []string{"proj1", "proj2"}); err == nil {
			t.Error("Should reject more than one argument")
		}
	})
}

// --- Projects update subcommand ---

func TestProjectsUpdateCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Projects update help",
			Args:        []string{"projects", "update", "--help"},
			ExpectError: false,
			Contains: []string{
				"Update a project's site on Walrus blockchain",
				"--id",
				"--name",
				"--epochs",
			},
		},
		{
			Name:        "Projects update with invalid flag",
			Args:        []string{"projects", "update", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestProjectsUpdateCommandDescription(t *testing.T) {
	updateCmd := findCommand(rootCmd, "projects", "update")
	if updateCmd == nil {
		t.Fatal("projects update command not found")
	}

	t.Run("Short description is non-empty", func(t *testing.T) {
		if updateCmd.Short == "" {
			t.Error("Short description is empty")
		}
	})

	t.Run("Long description is non-empty", func(t *testing.T) {
		if updateCmd.Long == "" {
			t.Error("Long description is empty")
		}
	})

	t.Run("Long description mentions Walrus blockchain", func(t *testing.T) {
		if !containsStr(updateCmd.Long, "Walrus") {
			t.Error("Long description should mention Walrus")
		}
	})

	t.Run("Short description mentions on-chain", func(t *testing.T) {
		if !containsStr(updateCmd.Short, "on-chain") {
			t.Error("Short description should mention on-chain")
		}
	})
}

func TestProjectsUpdateCommandFlags(t *testing.T) {
	updateCmd := findCommand(rootCmd, "projects", "update")
	if updateCmd == nil {
		t.Fatal("projects update command not found")
	}

	flagTests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"id flag", "id", "", "0"},
		{"name flag", "name", "", ""},
		{"epochs flag", "epochs", "e", "0"},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			flag := updateCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("Expected shorthand '%s', got '%s'", tt.shorthand, flag.Shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Expected default value '%s', got '%s'", tt.defValue, flag.DefValue)
			}
		})
	}
}

func TestProjectsUpdateCommandArgsValidation(t *testing.T) {
	updateCmd := findCommand(rootCmd, "projects", "update")
	if updateCmd == nil {
		t.Fatal("projects update command not found")
	}

	t.Run("Accepts no arguments", func(t *testing.T) {
		if updateCmd.Args(updateCmd, []string{}) != nil {
			t.Error("Should accept no arguments (uses flags)")
		}
	})

	t.Run("Accepts one argument", func(t *testing.T) {
		if updateCmd.Args(updateCmd, []string{"myproject"}) != nil {
			t.Error("Should accept one positional argument")
		}
	})

	t.Run("Rejects two arguments", func(t *testing.T) {
		if err := updateCmd.Args(updateCmd, []string{"proj1", "proj2"}); err == nil {
			t.Error("Should reject more than one argument")
		}
	})
}

// --- Subcommand registration tests ---

func TestProjectsSubcommandsRegistered(t *testing.T) {
	projectsCommand := findCommand(rootCmd, "projects")
	if projectsCommand == nil {
		t.Fatal("projects command not found")
	}

	expectedSubcommands := []string{"list", "show", "edit", "update", "delete", "archive"}

	subcommands := make(map[string]bool)
	for _, child := range projectsCommand.Commands() {
		subcommands[child.Name()] = true
	}

	for _, expected := range expectedSubcommands {
		t.Run("Subcommand "+expected+" is registered", func(t *testing.T) {
			if !subcommands[expected] {
				t.Errorf("Expected subcommand '%s' to be registered under 'projects'", expected)
			}
		})
	}
}

func TestProjectsSubcommandsHaveRunE(t *testing.T) {
	projectsCommand := findCommand(rootCmd, "projects")
	if projectsCommand == nil {
		t.Fatal("projects command not found")
	}

	t.Run("projects root has RunE", func(t *testing.T) {
		if projectsCommand.RunE == nil {
			t.Error("projects root command should have RunE set (defaults to list)")
		}
	})

	for _, child := range projectsCommand.Commands() {
		t.Run(child.Name()+" has RunE", func(t *testing.T) {
			if child.RunE == nil {
				t.Errorf("Subcommand '%s' should have RunE set", child.Name())
			}
		})
	}
}

// --- Project identifier flags consistency ---

func TestProjectIdentifierFlagsConsistency(t *testing.T) {
	// All subcommands that take project identifiers should have both --id and --name
	subcommandsWithIdentifiers := []string{"show", "edit", "update", "delete", "archive"}

	for _, subcmdName := range subcommandsWithIdentifiers {
		subcmd := findCommand(rootCmd, "projects", subcmdName)
		if subcmd == nil {
			t.Fatalf("projects %s command not found", subcmdName)
		}

		t.Run(subcmdName+" has --id flag", func(t *testing.T) {
			flag := subcmd.Flags().Lookup("id")
			if flag == nil {
				t.Errorf("projects %s should have --id flag", subcmdName)
			}
		})

		t.Run(subcmdName+" has --name flag", func(t *testing.T) {
			flag := subcmd.Flags().Lookup("name")
			if flag == nil {
				t.Errorf("projects %s should have --name flag", subcmdName)
			}
		})

		t.Run(subcmdName+" accepts MaximumNArgs(1)", func(t *testing.T) {
			if err := subcmd.Args(subcmd, []string{}); err != nil {
				t.Errorf("projects %s should accept 0 args", subcmdName)
			}
			if err := subcmd.Args(subcmd, []string{"arg1"}); err != nil {
				t.Errorf("projects %s should accept 1 arg", subcmdName)
			}
			if err := subcmd.Args(subcmd, []string{"arg1", "arg2"}); err == nil {
				t.Errorf("projects %s should reject 2 args", subcmdName)
			}
		})
	}
}

// --- Pure function tests ---

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"Less than a minute", 30 * time.Second, "just now"},
		{"Exactly one minute", 1 * time.Minute, "1 minute"},
		{"Five minutes", 5 * time.Minute, "5 minutes"},
		{"Exactly one hour", 1 * time.Hour, "1 hour"},
		{"Three hours", 3 * time.Hour, "3 hours"},
		{"Exactly one day", 24 * time.Hour, "1 day"},
		{"Two days", 48 * time.Hour, "2 days"},
		{"Seven days", 7 * 24 * time.Hour, "7 days"},
		{"Thirty seconds", 30 * time.Second, "just now"},
		{"59 minutes", 59 * time.Minute, "59 minutes"},
		{"23 hours", 23 * time.Hour, "23 hours"},
		{"100 days", 100 * 24 * time.Hour, "100 days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.input)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculateExpiryDate(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		lastDeploy   time.Time
		epochs       int
		network      string
		expectedDays int
	}{
		{"Mainnet 1 epoch", baseTime, 1, "mainnet", 14},
		{"Mainnet 5 epochs", baseTime, 5, "mainnet", 70},
		{"Testnet 1 epoch", baseTime, 1, "testnet", 1},
		{"Testnet 10 epochs", baseTime, 10, "testnet", 10},
		{"Mainnet 0 epochs", baseTime, 0, "mainnet", 0},
		{"Testnet 0 epochs", baseTime, 0, "testnet", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateExpiryDate(tt.lastDeploy, tt.epochs, tt.network)
			expected := baseTime.Add(time.Duration(tt.expectedDays) * 24 * time.Hour)
			if !result.Equal(expected) {
				t.Errorf("calculateExpiryDate(%v, %d, %q) = %v, want %v",
					tt.lastDeploy, tt.epochs, tt.network, result, expected)
			}
		})
	}
}

func TestFormatExpiryDuration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		expiry   time.Time
		contains string
	}{
		{"Already expired", now.Add(-24 * time.Hour), "Expired"},
		{"Expiring soon (minutes)", now.Add(30 * time.Minute), "Expiring soon"},
		{"Hours away", now.Add(5 * time.Hour), "hours"},
		{"1 day away", now.Add(30 * time.Hour), "1 day"},
		{"Days away", now.Add(3 * 24 * time.Hour), "days"},
		{"Weeks away", now.Add(14 * 24 * time.Hour), "weeks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatExpiryDuration(tt.expiry)
			if !containsStr(result, tt.contains) {
				t.Errorf("formatExpiryDuration(%v) = %q, want it to contain %q", tt.expiry, result, tt.contains)
			}
		})
	}
}

func TestFormatExpiryDurationEdgeCases(t *testing.T) {
	now := time.Now()

	t.Run("Exactly 7 days plus buffer", func(t *testing.T) {
		// Add extra hour buffer to ensure we land solidly in the 7-day+ range
		result := formatExpiryDuration(now.Add(7*24*time.Hour + time.Hour))
		if !containsStr(result, "week") {
			t.Errorf("7+ days should show weeks, got %q", result)
		}
	})

	t.Run("8 days", func(t *testing.T) {
		result := formatExpiryDuration(now.Add(8*24*time.Hour + time.Hour))
		if !containsStr(result, "week") {
			t.Errorf("8 days should show weeks, got %q", result)
		}
	})

	t.Run("Exactly 1 day no extra hours", func(t *testing.T) {
		// Add exactly 24 hours + 1 minute to ensure it's 1 day
		result := formatExpiryDuration(now.Add(24*time.Hour + time.Minute))
		if !containsStr(result, "1 day") {
			t.Errorf("Expected '1 day', got %q", result)
		}
	})
}

// --- Helper ---

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
