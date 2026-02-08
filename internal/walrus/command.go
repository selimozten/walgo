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
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
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
