package cmd

import "testing"

func TestVersionInit(t *testing.T) {
	// Test that init properly adds that command
	// This is mostly covered by the command execution tests,
	// but we can verify that command is registered

	// Find the version command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			found = true

			// Check flags are registered
			checkUpdatesFlag := cmd.Flags().Lookup("check-updates")
			if checkUpdatesFlag == nil {
				t.Error("check-updates flag not found")
			}

			shortFlag := cmd.Flags().Lookup("short")
			if shortFlag == nil {
				t.Error("short flag not found")
			}

			break
		}
	}

	if !found {
		t.Error("version command not registered with root command")
	}
}
