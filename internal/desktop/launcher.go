package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetBinaryPath returns the path to the installed desktop binary for the current platform
func GetBinaryPath() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Applications/Walgo.app
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "Applications", "Walgo.app")
	case "windows":
		// Windows: %LOCALAPPDATA%\Programs\Walgo\walgo-desktop.exe
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, "Programs", "Walgo", "walgo-desktop.exe")
		}
		// Fallback
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs", "Walgo", "walgo-desktop.exe")
	default: // linux
		// Linux: ~/.local/bin/walgo-desktop
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".local", "bin", "walgo-desktop")
	}
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
