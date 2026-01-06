package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/selimozten/walgo/internal/ui"
)

// cleanupWalgoData removes walgo configuration and cache directories.
func cleanupWalgoData() error {
	icons := ui.GetIcons()
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = os.Getenv("USERPROFILE")
	}

	dataPaths := []string{
		filepath.Join(homeDir, ".walgo"),
		filepath.Join(homeDir, ".config", "walgo"),
	}

	found := false
	for _, path := range dataPaths {
		if _, err := os.Stat(path); err == nil {
			found = true
			fmt.Printf("Removing: %s\n", path)

			if err := os.RemoveAll(path); err != nil {
				fmt.Printf("%s Failed to remove %s: %v\n", icons.Warning, path, err)
			} else {
				fmt.Printf("%s Removed: %s\n", icons.Check, path)
			}
		}
	}

	if !found {
		fmt.Printf("%s No data files found\n", icons.Info)
	}

	return nil
}

// isWritable checks if a file path is writable by the current user.
func isWritable(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// showBlockchainWarning displays information about blockchain data preservation.
// This reassures users that their on-chain assets are safe during uninstall.
func showBlockchainWarning() {
	icons := ui.GetIcons()
	fmt.Printf("%s IMPORTANT: Your blockchain data is safe!\n", icons.Warning)
	fmt.Println()
	fmt.Println("Uninstalling Walgo will NOT delete:")
	fmt.Printf("  %s Your SUI balance\n", icons.Check)
	fmt.Printf("  %s Your deployed sites on Walrus\n", icons.Check)
	fmt.Printf("  %s Your NFTs and on-chain objects\n", icons.Check)
	fmt.Println()
	fmt.Println("These live on the Sui blockchain, not on your computer.")
	fmt.Println()

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = os.Getenv("USERPROFILE")
	}

	suiConfigDir := filepath.Join(homeDir, ".sui")
	if _, err := os.Stat(suiConfigDir); err == nil {
		fmt.Printf("%s WALLET BACKUP REMINDER:\n", icons.Warning)
		fmt.Println()
		fmt.Println("Your Sui wallet keys are stored in:")
		fmt.Printf("  %s%s\n", icons.File, suiConfigDir)
		fmt.Println()
		fmt.Println("Walgo will NOT delete this directory, but if you manually")
		fmt.Println("delete it later, make sure you have backed up:")
		fmt.Println("  - Your seed phrase/recovery phrase")
		fmt.Println("  - Or export your private keys")
		fmt.Println()
		fmt.Println("To backup your wallet:")
		fmt.Println("  sui keytool export --key-identity <address>")
		fmt.Println()
	}

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
}
