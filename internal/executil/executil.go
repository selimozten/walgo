// Package executil provides exec.Command wrappers that hide console windows
// on Windows. On non-Windows platforms, these are thin wrappers around the
// standard library functions.
package executil

import (
	"context"
	"os/exec"
)

// Command creates an exec.Cmd with platform-appropriate settings.
// On Windows, it sets CREATE_NO_WINDOW to prevent console window flashing.
func Command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	hideWindow(cmd)
	return cmd
}

// CommandContext creates a context-aware exec.Cmd with platform-appropriate settings.
// On Windows, it sets CREATE_NO_WINDOW to prevent console window flashing.
func CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	hideWindow(cmd)
	return cmd
}
