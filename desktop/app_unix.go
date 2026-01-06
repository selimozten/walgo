//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// killExistingHugoProcesses kills any existing Hugo serve processes (Unix/Linux/macOS)
func killExistingHugoProcesses() error {
	// Use ps to find hugo processes
	cmd := exec.Command("ps", "-eo", "pid,command")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var killedPIDs []int

	for _, line := range lines {
		if strings.Contains(line, "hugo") && strings.Contains(line, "server") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			pid, err := strconv.Atoi(fields[0])
			if err != nil {
				continue
			}

			// Skip self (this process)
			if pid == os.Getpid() || pid == os.Getppid() {
				continue
			}

			// Kill the process
			if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
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
