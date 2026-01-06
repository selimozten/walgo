package desktop

import (
	"fmt"
	"os"
	"os/exec"
)

// CommandRunner is an interface for running commands
// This allows for testing both darwin and non-darwin code paths
type CommandRunner interface {
	Run(cmd *exec.Cmd) error
}

// DefaultRunner is the default command runner that executes commands directly
type DefaultRunner struct{}

// Run executes the command
func (r *DefaultRunner) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

// LaunchWithRunner launches the desktop application using the provided runner
// This is an internal function that allows testing both code paths
func LaunchWithRunner(binaryPath string, isDarwin bool, runner CommandRunner) error {
	if isDarwin {
		// Use 'open' command on macOS for .app bundles
		launchCmd := exec.Command("open", binaryPath)
		if err := runner.Run(launchCmd); err != nil {
			return fmt.Errorf("failed to launch: %w", err)
		}
	} else {
		// Direct execution on Linux/Windows
		launchCmd := exec.Command(binaryPath)
		launchCmd.Stdout = os.Stdout
		launchCmd.Stderr = os.Stderr

		if err := runner.Run(launchCmd); err != nil {
			return fmt.Errorf("failed to launch: %w", err)
		}
	}

	return nil
}
