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

		// Get possible install locations based on platform (check multiple locations)
		var possiblePaths []string
		var installLocation string

		homeDir, _ := os.UserHomeDir()

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
		var binaryPath string
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				binaryPath = path
				installLocation = filepath.Dir(path)
				break
			}
		}

		// Check if binary exists
		if binaryPath == "" {
			fmt.Fprintf(os.Stderr, "%s Error: Desktop app not found\n", icons.Error)
			fmt.Fprintf(os.Stderr, "  %s Checked locations:\n", icons.File)
			for _, path := range possiblePaths {
				fmt.Fprintf(os.Stderr, "    - %s\n", path)
			}
			fmt.Fprintf(os.Stderr, "\n%s Install desktop app:\n", icons.Lightbulb)
			fmt.Fprintf(os.Stderr, "  Run installer: curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash\n")
			return fmt.Errorf("desktop app not found in any standard location")
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
