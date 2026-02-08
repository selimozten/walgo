//go:build !windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/selimozten/walgo/internal/ui"
)

// killExistingHugoProcesses finds and kills any existing 'hugo serve' processes on Unix systems
func killExistingHugoProcesses() error {
	icons := ui.GetIcons()

	// On Unix, use ps to find hugo processes
	cmd := exec.Command("ps", "ax")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Warning: Could not check for running Hugo processes: %v\n", icons.Warning, err)
		return nil // Don't fail, just continue
	}

	lines := strings.Split(string(output), "\n")
	var killedPIDs []int

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Check command field (5th column onward in ps ax output) for hugo server/serve
		isHugoServer := false
		for _, f := range fields[4:] {
			base := filepath.Base(f)
			if base == "hugo" {
				isHugoServer = true
				continue
			}
			if isHugoServer && (f == "server" || f == "serve") {
				break
			}
			if isHugoServer && !strings.HasPrefix(f, "-") {
				isHugoServer = false
			}
		}

		if !isHugoServer {
			continue
		}

		// Try to parse PID from first field
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
			fmt.Fprintf(os.Stderr, "%s Warning: Could not kill Hugo process %d: %v\n", icons.Warning, pid, err)
		} else {
			killedPIDs = append(killedPIDs, pid)
		}
	}

	if len(killedPIDs) > 0 {
		fmt.Printf("%s Killed %d existing Hugo serve process(es)\n", icons.Info, len(killedPIDs))
		for _, pid := range killedPIDs {
			fmt.Printf("  - PID: %d\n", pid)
		}
	}

	return nil
}
