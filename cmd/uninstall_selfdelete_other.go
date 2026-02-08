//go:build !windows

package cmd

// windowsDeleteBinary is a no-op on non-Windows platforms.
// On Unix systems, a running binary can be deleted directly via os.Remove.
func windowsDeleteBinary(_ string) error {
	return nil
}
