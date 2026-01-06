//go:build windows
// +build windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// killExistingHugoProcesses kills any existing Hugo serve processes (Windows)
func killExistingHugoProcesses() error {
	// Use tasklist to find hugo processes
	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var killedPIDs []int

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "hugo") {
			// Parse CSV format: "hugo.exe","1234","Console","1","12,345 K"
			fields := strings.Split(line, ",")
			if len(fields) < 2 {
				continue
			}

			// Remove quotes from PID field
			pidStr := strings.Trim(fields[1], `"`)
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				continue
			}

			// Kill the process using taskkill
			killCmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
			if err := killCmd.Run(); err != nil {
				fmt.Printf("⚠️  Warning: Could not kill Hugo process %d: %v\n", pid, err)
			} else {
				killedPIDs = append(killedPIDs, pid)
			}
		}
	}

	if len(killedPIDs) > 0 {
		fmt.Printf("ℹ️  Killed %d existing Hugo serve process(es)\n", len(killedPIDs))
		for _, pid := range killedPIDs {
			fmt.Printf("  - PID: %d\n", pid)
		}
	}

	return nil
}
