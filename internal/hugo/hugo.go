package hugo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// InitializeSite creates a new Hugo site at the given path.
// It runs `hugo new site .` in the specified directory.
// For Phase 1, Walgo interacts with Hugo via `exec.Command`. This approach is chosen
// for its simplicity and broad compatibility, ensuring Walgo can work with various
// Hugo versions without direct dependency on Hugo's internal libraries.
// Future enhancements might explore direct Hugo library integration for more
// granular control or performance benefits where applicable.
func InitializeSite(sitePath string) error {
	// Check if Hugo is installed
	if _, err := exec.LookPath("hugo"); err != nil {
		return fmt.Errorf("Hugo is not installed or not found in PATH. Please install Hugo first.")
	}

	cmd := exec.Command("hugo", "new", "site", ".", "--format", "toml")
	cmd.Dir = sitePath
	// Capture output for better error reporting or logging
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize Hugo site at %s: %v\nOutput: %s", sitePath, err, string(output))
	}

	fmt.Printf("Hugo `new site` command output:\n%s\n", string(output))

	// Create a basic hugo.toml if it doesn't exist (Hugo might create config.toml by default)
	// Hugo 0.125.x creates hugo.toml by default.
	// We will ensure a basic one is there or create one.
	// For now, let's assume hugo new site . --format toml handles this sufficiently.

	return nil
}

// BuildSite runs the Hugo build process in the given site path.
// It assumes that the current working directory is the site's root or sitePath is absolute.
// For Phase 1, Walgo interacts with Hugo via `exec.Command`. This approach is chosen
// for its simplicity and broad compatibility. Future enhancements might explore direct
// Hugo library integration.
func BuildSite(sitePath string) error {
	// Check if Hugo is installed
	if _, err := exec.LookPath("hugo"); err != nil {
		return fmt.Errorf("Hugo is not installed or not found in PATH. Please install Hugo first.")
	}

	// Check if we are in a Hugo site directory (e.g., by checking for hugo.toml or go.mod if it's a module theme)
	// For simplicity, we assume `sitePath` is the root of a Hugo project.
	configFile := filepath.Join(sitePath, "hugo.toml") // or config.toml, config.yaml
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Try other common config file names
		configFile = filepath.Join(sitePath, "config.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("hugo configuration file (hugo.toml/config.toml) not found in %s. Are you in a Hugo site directory?", sitePath)
		}
	}

	fmt.Printf("Building Hugo site in %s...\n", sitePath)
	cmd := exec.Command("hugo")
	cmd.Dir = sitePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build Hugo site: %v", err)
	}

	fmt.Println("Hugo site built successfully.")
	publicDir := filepath.Join(sitePath, "public")
	fmt.Printf("Static files generated in: %s\n", publicDir)
	return nil
}
