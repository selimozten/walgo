//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/selimozten/walgo/internal/ui"
)

// hideConsoleWindow sets Windows-specific flags to hide console window
func hideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

// killExistingHugoProcesses finds and kills any existing 'hugo serve' processes on Windows
func killExistingHugoProcesses() error {
	icons := ui.GetIcons()

	// On Windows, use tasklist to find hugo processes
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq hugo.exe", "/FO", "CSV", "/NH")
	hideConsoleWindow(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Warning: Could not check for running Hugo processes: %v\n", icons.Warning, err)
		return nil // Don't fail, just continue
	}

	lines := strings.Split(string(output), "\n")
	var killedCount int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(strings.ToLower(line), "hugo") {
			continue
		}

		// Extract PID from CSV format: "hugo.exe","PID","..."
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}

		pid := strings.Trim(fields[1], "\"")

		// Kill the process using taskkill
		killCmd := exec.Command("taskkill", "/PID", pid, "/F")
		hideConsoleWindow(killCmd)
		if err := killCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Could not kill Hugo process %s: %v\n", icons.Warning, pid, err)
		} else {
			killedCount++
		}
	}

	if killedCount > 0 {
		fmt.Printf("%s Killed %d existing Hugo serve process(es)\n", icons.Info, killedCount)
	}

	return nil
}
