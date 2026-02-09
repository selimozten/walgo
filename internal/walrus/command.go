package walrus

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync/atomic"

	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/executil"
)

// Test hooks for dependency injection.
var (
	execLookPath = deps.LookPath
	execCommand  = executil.Command
	verboseFlag  atomic.Bool // Thread-safe verbose mode flag
	osStat       = os.Stat
)

// execCommandContext is a test hook for creating context-aware commands.
var execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return executil.CommandContext(ctx, name, args...)
}

// SetVerbose enables or disables verbose output mode.
// Thread-safe for concurrent access.
func SetVerbose(verbose bool) {
	verboseFlag.Store(verbose)
}

// isVerbose returns the current verbose mode setting.
// Thread-safe for concurrent access.
func isVerbose() bool {
	return verboseFlag.Load()
}

// safeWriter wraps a writer and silently ignores write errors.
// This prevents broken os.Stdout/os.Stderr in GUI apps from
// disrupting pipe-based output capture.
type safeWriter struct {
	w io.Writer
}

func (s safeWriter) Write(p []byte) (int, error) {
	n, err := s.w.Write(p)
	if err != nil {
		return len(p), nil // Pretend success
	}
	return n, nil
}

// runCommandWithTimeout executes a command with a timeout context.
// Returns stdout, stderr, and any error.
func runCommandWithTimeout(ctx context.Context, name string, args []string, streamOutput bool) (string, string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), DefaultCommandTimeout)
		defer cancel()
	}

	cmd := execCommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer

	if streamOutput {
		// Use safeWriter for os.Stdout/os.Stderr to prevent errors when
		// running from a GUI app (e.g. Wails) where standard handles may
		// be invalid. The buffer always captures output reliably.
		cmd.Stdout = io.MultiWriter(safeWriter{os.Stdout}, &stdout)
		cmd.Stderr = io.MultiWriter(safeWriter{os.Stderr}, &stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return stdout.String(), stderr.String(), fmt.Errorf("command timed out after %v - the operation took too long", DefaultCommandTimeout)
	}

	if ctx.Err() == context.Canceled {
		return stdout.String(), stderr.String(), fmt.Errorf("command was cancelled")
	}

	return stdout.String(), stderr.String(), err
}
