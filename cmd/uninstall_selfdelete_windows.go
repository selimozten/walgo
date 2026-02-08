//go:build windows

package cmd

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/selimozten/walgo/internal/ui"
)

// windowsDeleteBinary spawns a detached cmd.exe process that waits for
// walgo.exe to exit, then deletes the binary. This is needed because a
// running Windows executable cannot delete itself.
func windowsDeleteBinary(binaryPath string) error {
	icons := ui.GetIcons()

	// ping -n 3 provides ~2 second delay for walgo.exe to fully exit,
	// then del /f /q force-deletes the binary.
	script := fmt.Sprintf(`ping -n 3 127.0.0.1 >nul & del /f /q "%s"`, binaryPath)

	cmd := exec.Command("cmd.exe", "/C", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000 | 0x00000200, // CREATE_NO_WINDOW | CREATE_NEW_PROCESS_GROUP
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to schedule binary deletion: %w", err)
	}

	fmt.Printf("%s CLI binary will be removed after exit\n", icons.Check)
	return nil
}
