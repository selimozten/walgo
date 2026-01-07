package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetBinaryPath returns the path to the installed desktop binary for the current platform
// It checks multiple possible install locations and returns the first one found
func GetBinaryPath() string {
	homeDir, _ := os.UserHomeDir()

	var possiblePaths []string

	switch runtime.GOOS {
	case "darwin":
		// macOS: Check both system and user Applications
		possiblePaths = []string{
			"/Applications/Walgo.app",                           // System-wide (preferred)
			filepath.Join(homeDir, "Applications", "Walgo.app"), // User-specific
		}
	case "windows":
		// Windows: Check LOCALAPPDATA and USERPROFILE
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			possiblePaths = append(possiblePaths, filepath.Join(localAppData, "Programs", "Walgo", "Walgo.exe"))
		}
		possiblePaths = append(possiblePaths, filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs", "Walgo", "Walgo.exe"))
	default: // linux
		// Linux: Check multiple possible locations
		possiblePaths = []string{
			filepath.Join(homeDir, ".local", "bin", "Walgo"),
			"/usr/local/bin/Walgo",
		}
	}

	// Find the first existing path
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// If nothing found, return the preferred default
	if len(possiblePaths) > 0 {
		return possiblePaths[0]
	}

	return ""
}

// Launch launches the desktop application
func Launch(binaryPath string) error {
	if runtime.GOOS == "darwin" {
		// Use 'open' command on macOS for .app bundles
		launchCmd := exec.Command("open", binaryPath)
		if err := launchCmd.Run(); err != nil {
			return fmt.Errorf("failed to launch: %w", err)
		}
	} else {
		// Direct execution on Linux/Windows
		launchCmd := exec.Command(binaryPath)
		launchCmd.Stdout = os.Stdout
		launchCmd.Stderr = os.Stderr

		if err := launchCmd.Run(); err != nil {
			return fmt.Errorf("failed to launch: %w", err)
		}
	}

	return nil
}
