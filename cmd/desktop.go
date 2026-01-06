package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/selimozten/walgo/internal/desktop"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// desktopCmd represents the desktop command
var desktopCmd = &cobra.Command{
	Use:   "desktop",
	Short: "Launch the Walgo desktop application",
	Long: `Launches the Walgo desktop GUI application.

The desktop app provides a graphical interface for:
• Creating and managing Hugo sites
• Building and deploying to Walrus
• AI-powered content generation
• Project management
• Obsidian vault import
• Site optimization and compression

Example:
  walgo desktop          # Launch the desktop app
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()

		// Get standard install location based on platform
		var binaryPath string
		var installLocation string

		switch runtime.GOOS {
		case "darwin":
			// macOS: ~/Applications/Walgo.app
			homeDir, _ := os.UserHomeDir()
			binaryPath = filepath.Join(homeDir, "Applications", "Walgo.app")
			installLocation = filepath.Join(homeDir, "Applications")
		case "windows":
			// Windows: %LOCALAPPDATA%\Programs\Walgo\walgo-desktop.exe
			localAppData := os.Getenv("LOCALAPPDATA")
			if localAppData != "" {
				binaryPath = filepath.Join(localAppData, "Programs", "Walgo", "walgo-desktop.exe")
				installLocation = filepath.Join(localAppData, "Programs", "Walgo")
			} else {
				// Fallback
				binaryPath = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs", "Walgo", "walgo-desktop.exe")
				installLocation = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs", "Walgo")
			}
		default: // linux
			// Linux: ~/.local/bin/walgo-desktop
			homeDir, _ := os.UserHomeDir()
			binaryPath = filepath.Join(homeDir, ".local", "bin", "walgo-desktop")
			installLocation = filepath.Join(homeDir, ".local", "bin")
		}

		// Check if binary exists
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s Error: Desktop app not found\n", icons.Error)
			fmt.Fprintf(os.Stderr, "  %s Looking for: %s\n", icons.File, binaryPath)
			fmt.Fprintf(os.Stderr, "\n%s Install desktop app:\n", icons.Lightbulb)
			fmt.Fprintf(os.Stderr, "  Run installer: curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash\n")
			return fmt.Errorf("desktop app not found: %s", binaryPath)
		}

		// Launch production binary
		fmt.Printf("%s Launching Walgo Desktop...\n", icons.Desktop)
		fmt.Printf("   Location: %s\n", installLocation)
		fmt.Println()

		// Launch the app
		if err := desktop.Launch(binaryPath); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Failed to launch desktop app\n", icons.Error)
			fmt.Fprintf(os.Stderr, "   %v\n", err)
			return fmt.Errorf("failed to launch desktop app: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(desktopCmd)
}
