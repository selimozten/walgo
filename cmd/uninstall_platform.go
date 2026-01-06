package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/selimozten/walgo/internal/ui"
)

// uninstallCLIBinary removes the walgo binary from the system.
// It detects if sudo is required and handles both cases.
func uninstallCLIBinary() error {
	icons := ui.GetIcons()
	binaryPath, err := exec.LookPath("walgo")
	if err != nil {
		return fmt.Errorf("walgo binary not found in PATH")
	}

	fmt.Printf("Found walgo at: %s\n", binaryPath)

	if realPath, err := filepath.EvalSymlinks(binaryPath); err == nil {
		binaryPath = realPath
		fmt.Printf("Resolved to: %s\n", binaryPath)
	}

	needSudo := !isWritable(binaryPath)

	if needSudo {
		fmt.Println("Removing binary (requires sudo)...")
		cmd := exec.Command("sudo", "rm", "-f", binaryPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to remove binary: %w", err)
		}
	} else {
		fmt.Println("Removing binary...")
		if err := os.Remove(binaryPath); err != nil {
			return fmt.Errorf("failed to remove binary: %w", err)
		}
	}

	fmt.Printf("%s CLI binary removed\n", icons.Check)
	return nil
}

// uninstallDesktopApp removes the desktop application based on the current OS.
func uninstallDesktopApp() error {
	switch runtime.GOOS {
	case "darwin":
		return uninstallDesktopMacOS()
	case "windows":
		return uninstallDesktopWindows()
	case "linux":
		return uninstallDesktopLinux()
	default:
		return fmt.Errorf("desktop uninstall not supported on %s", runtime.GOOS)
	}
}

func uninstallDesktopMacOS() error {
	icons := ui.GetIcons()
	appPaths := []string{
		"/Applications/Walgo.app",
		"/Applications/walgo-desktop.app",
		filepath.Join(os.Getenv("HOME"), "Applications", "Walgo.app"),
		filepath.Join(os.Getenv("HOME"), "Applications", "walgo-desktop.app"),
	}

	found := false
	for _, appPath := range appPaths {
		if _, err := os.Stat(appPath); err == nil {
			found = true
			fmt.Printf("Found desktop app at: %s\n", appPath)

			needSudo := !isWritable(appPath)

			if needSudo {
				fmt.Println("Removing app (requires sudo)...")
				cmd := exec.Command("sudo", "rm", "-rf", appPath)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to remove app: %w", err)
				}
			} else {
				fmt.Println("Removing app...")
				if err := os.RemoveAll(appPath); err != nil {
					return fmt.Errorf("failed to remove app: %w", err)
				}
			}

			fmt.Printf("%s Removed: %s\n", icons.Check, appPath)
		}
	}

	if !found {
		return fmt.Errorf("desktop app not found")
	}

	return nil
}

func uninstallDesktopWindows() error {
	icons := ui.GetIcons()
	homePath := os.Getenv("LOCALAPPDATA")
	if homePath == "" {
		homePath = os.Getenv("USERPROFILE")
	}

	appPaths := []string{
		// New naming (v0.3.0+)
		filepath.Join(homePath, "Programs", "Walgo", "Walgo.exe"),
		filepath.Join(os.Getenv("PROGRAMFILES"), "Walgo", "Walgo.exe"),
		filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Walgo", "Walgo.exe"),
		// Old naming (backward compatibility)
		filepath.Join(homePath, "Programs", "Walgo", "walgo-desktop.exe"),
		filepath.Join(os.Getenv("PROGRAMFILES"), "Walgo", "walgo-desktop.exe"),
		filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Walgo", "walgo-desktop.exe"),
	}

	found := false
	for _, appPath := range appPaths {
		if _, err := os.Stat(appPath); err == nil {
			found = true
			appDir := filepath.Dir(appPath)
			fmt.Printf("Found desktop app at: %s\n", appDir)

			fmt.Println("Removing app...")
			if err := os.RemoveAll(appDir); err != nil {
				return fmt.Errorf("failed to remove app: %w", err)
			}

			fmt.Printf("%s Removed: %s\n", icons.Check, appDir)
		}
	}

	startMenu := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	shortcutPath := filepath.Join(startMenu, "Walgo.lnk")
	if _, err := os.Stat(shortcutPath); err == nil {
		os.Remove(shortcutPath)
		fmt.Printf("%s Removed Start Menu shortcut\n", icons.Check)
	}

	if !found {
		return fmt.Errorf("desktop app not found")
	}

	return nil
}

func uninstallDesktopLinux() error {
	icons := ui.GetIcons()
	homeDir := os.Getenv("HOME")

	binaryPaths := []string{
		// New naming (v0.3.0+)
		filepath.Join(homeDir, ".local", "bin", "Walgo"),
		"/usr/local/bin/Walgo",
		"/usr/bin/Walgo",
		// Old naming (backward compatibility)
		filepath.Join(homeDir, ".local", "bin", "walgo-desktop"),
		"/usr/local/bin/walgo-desktop",
		"/usr/bin/walgo-desktop",
	}

	found := false
	for _, binaryPath := range binaryPaths {
		if _, err := os.Stat(binaryPath); err == nil {
			found = true
			fmt.Printf("Found desktop binary at: %s\n", binaryPath)

			needSudo := !isWritable(binaryPath)

			if needSudo {
				fmt.Println("Removing binary (requires sudo)...")
				cmd := exec.Command("sudo", "rm", "-f", binaryPath)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to remove binary: %w", err)
				}
			} else {
				fmt.Println("Removing binary...")
				if err := os.Remove(binaryPath); err != nil {
					return fmt.Errorf("failed to remove binary: %w", err)
				}
			}

			fmt.Printf("%s Removed: %s\n", icons.Check, binaryPath)
		}
	}

	desktopFiles := []string{
		filepath.Join(homeDir, ".local", "share", "applications", "walgo-desktop.desktop"),
		"/usr/share/applications/walgo-desktop.desktop",
		"/usr/local/share/applications/walgo-desktop.desktop",
	}

	for _, desktopFile := range desktopFiles {
		if _, err := os.Stat(desktopFile); err == nil {
			fmt.Printf("Found desktop entry at: %s\n", desktopFile)

			needSudo := !isWritable(desktopFile)

			if needSudo {
				cmd := exec.Command("sudo", "rm", "-f", desktopFile)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				_ = cmd.Run() // Best-effort cleanup
			} else {
				_ = os.Remove(desktopFile) // Best-effort cleanup
			}

			fmt.Printf("%s Removed: %s\n", icons.Check, desktopFile)

			if exec.Command("which", "update-desktop-database").Run() == nil {
				desktopDir := filepath.Dir(desktopFile)
				_ = exec.Command("update-desktop-database", desktopDir).Run() // Best-effort cleanup
			}
		}
	}

	if !found {
		return fmt.Errorf("desktop app not found")
	}

	return nil
}
