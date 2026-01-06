package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestBuildCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Build command help",
			Args:        []string{"build", "--help"},
			ExpectError: false,
			Contains: []string{
				"Builds the Hugo site",
			},
		},
		{
			Name:        "Build command with invalid flag",
			Args:        []string{"build", "--invalid-flag"},
			ExpectError: true,
			Contains: []string{
				"unknown flag",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestBuildCommandFlags(t *testing.T) {
	// Find the build command
	var buildCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "build" {
			buildCommand = cmd
			break
		}
	}

	if buildCommand == nil {
		t.Fatal("build command not found")
	}
}
