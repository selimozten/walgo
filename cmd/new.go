package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [content-type/slug]",
	Short: "Create new content in your Hugo site.",
	Long: `Creates a new content file (e.g., post, project) in your Hugo site.
This command is a wrapper around 'hugo new content ...'.
Example: walgo new posts/my-first-post.md`,
	Args: cobra.ExactArgs(1), // Expects exactly one argument: content-type/slug
	Run: func(cmd *cobra.Command, args []string) {
		contentPathArg := args[0]

		// Validate content path to prevent command injection
		if !isValidContentPath(contentPathArg) {
			fmt.Fprintf(os.Stderr, "Error: Invalid content path. Use only alphanumeric, hyphens, underscores, slashes and .md extension\n")
			os.Exit(1)
		}

		fmt.Printf("Creating new content: %s\n", contentPathArg)

		// Determine site path (current directory by default)
		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		// This is a simple wrapper around `hugo new ...`
		hugoCmd := exec.Command("hugo", "new", contentPathArg)
		hugoCmd.Dir = sitePath
		// Capture output to determine the actual file created by Hugo
		output, err := hugoCmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating new content with Hugo: %v\nOutput:\n%s\n", err, string(output))
			os.Exit(1)
		}

		// TODO: Parse Hugo's output to reliably determine the exact path of the created file.
		// Hugo's output for `hugo new my-section/my-post.md` is usually like:
		// "/path/to/site/content/my-section/my-post.md created"
		// For now, we print the raw output and a best guess.
		fmt.Println(string(output)) // Print Hugo's direct output

		// Best guess for the created file path for a simple success message
		// This might not be perfectly accurate if Hugo's output changes or for complex archetypes.
		createdFilePath := filepath.Join(sitePath, "content", contentPathArg)
		if filepath.Ext(createdFilePath) == "" {
			createdFilePath += ".md" // Common default
		}
		fmt.Printf("Successfully initiated content creation (see Hugo output above for exact path, e.g., %s).\n", createdFilePath)
		fmt.Println("Remember to edit the new file!")
	},
}

// isValidContentPath validates that content path only contains safe characters
func isValidContentPath(path string) bool {
	// Allow alphanumeric, hyphens, underscores, slashes, and .md/.html extensions
	validPath := regexp.MustCompile(`^[a-zA-Z0-9/_-]+\.(md|html|htm)?$`)
	return validPath.MatchString(path) && len(path) > 0 && len(path) < 200
}

func init() {
	rootCmd.AddCommand(newCmd)
	// newCmd.Flags().StringP("kind", "k", "", "Content type to create (archetype)")
}
