package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// runUninstall executes the uninstall process based on user options.
func runUninstall(cmd *cobra.Command, args []string) error {
	icons := ui.GetIcons()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              Walgo Uninstaller                            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	showBlockchainWarning()

	removeCLI := uninstallCLI || uninstallAll
	removeDesktop := uninstallDesktop || uninstallAll

	if !uninstallCLI && !uninstallDesktop && !uninstallAll {
		var err error
		removeCLI, removeDesktop, err = promptUninstallOptions()
		if err != nil {
			return err
		}
	}

	if !uninstallForce {
		if !confirmUninstall(removeCLI, removeDesktop) {
			fmt.Printf("%s Uninstall cancelled\n", icons.Cross)
			return nil
		}
	}

	var removed []string
	var failed []string

	fmt.Println()
	fmt.Println("Starting uninstall process...")
	fmt.Println()

	if removeCLI {
		fmt.Println("=== Uninstalling Walgo CLI ===")
		if err := uninstallCLIBinary(); err != nil {
			failed = append(failed, fmt.Sprintf("CLI binary: %v", err))
		} else {
			removed = append(removed, "CLI binary")
		}
		fmt.Println()
	}

	if removeDesktop {
		fmt.Println("=== Uninstalling Walgo Desktop ===")
		if err := uninstallDesktopApp(); err != nil {
			failed = append(failed, fmt.Sprintf("Desktop app: %v", err))
		} else {
			removed = append(removed, "Desktop app")
		}
		fmt.Println()
	}

	if (removeCLI || removeDesktop) && !uninstallKeepCache {
		fmt.Println("=== Cleaning Up Data ===")
		if err := cleanupWalgoData(); err != nil {
			failed = append(failed, fmt.Sprintf("Data cleanup: %v", err))
		} else {
			removed = append(removed, "Cache and data files")
		}
		fmt.Println()
	}

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("Uninstall Summary")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	if len(removed) > 0 {
		fmt.Printf("%s Successfully removed:\n", icons.Check)
		for _, item := range removed {
			fmt.Printf("  - %s\n", item)
		}
		fmt.Println()
	}

	if len(failed) > 0 {
		fmt.Printf("%s Failed to remove:\n", icons.Cross)
		for _, item := range failed {
			fmt.Printf("  - %s\n", item)
		}
		fmt.Println()
	}

	if len(removed) > 0 && len(failed) == 0 {
		fmt.Printf("%s Walgo has been successfully uninstalled!\n", icons.Check)
		fmt.Println()
		fmt.Println("Thank you for using Walgo. We hope to see you again!")
	}

	return nil
}

// promptUninstallOptions displays an interactive menu for uninstall choices.
func promptUninstallOptions() (bool, bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("What would you like to uninstall?")
	fmt.Println("  1) CLI only")
	fmt.Println("  2) Desktop app only")
	fmt.Println("  3) Both CLI and Desktop app")
	fmt.Println("  4) Cancel")
	fmt.Println()
	fmt.Print("Choose option [1-4]: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, false, err
	}

	input = strings.TrimSpace(input)

	switch input {
	case "1":
		return true, false, nil
	case "2":
		return false, true, nil
	case "3":
		return true, true, nil
	case "4":
		return false, false, fmt.Errorf("uninstall cancelled by user")
	default:
		return false, false, fmt.Errorf("invalid option")
	}
}

// confirmUninstall asks for final confirmation before proceeding.
func confirmUninstall(removeCLI, removeDesktop bool) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("This will remove:")
	if removeCLI {
		fmt.Println("  - Walgo CLI binary")
	}
	if removeDesktop {
		fmt.Println("  - Walgo Desktop application")
	}
	if !uninstallKeepCache {
		fmt.Println("  - Cache and data files")
	}
	fmt.Println()
	fmt.Print("Are you sure you want to continue? [y/N]: ")

	input, err := readLine(reader)
	if err != nil {
		return false
	}
	input = strings.ToLower(input)

	return input == "y" || input == "yes"
}
