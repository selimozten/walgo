package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	newNoBuild bool
	newServe   bool
)

var newCmd = &cobra.Command{
	Use:   "new [slug]",
	Short: "Create new content in your Hugo site",
	Long: `Creates a new content file in your Hugo site with automatic content type detection.

Walgo automatically detects your Hugo content structure and creates files in the appropriate directory.
After creation, it automatically builds the site (use --no-build to skip).

Examples:
  walgo new my-first-post           # Creates in detected content type (e.g., posts/)
  walgo new my-first-post --serve   # Creates, builds, and starts dev server
  walgo new my-first-post --no-build # Creates without building`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		reader := bufio.NewReader(os.Stdin)

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		// Detect available content types
		contentTypes, _ := hugo.DetectContentTypes(sitePath)
		defaultType := hugo.GetDefaultContentType(sitePath)

		// Show detected content types
		fmt.Printf("%s Detected content types:\n", icons.Info)
		if len(contentTypes) > 0 {
			for i, ct := range contentTypes {
				marker := " "
				if ct.Name == defaultType {
					marker = "*"
				}
				fmt.Printf("  %s %d) %s (%d files)\n", marker, i+1, ct.Name, ct.FileCount)
			}
		} else {
			fmt.Printf("  No content types found, will create '%s' directory\n", defaultType)
		}
		fmt.Println()

		// Get or prompt for content type
		var selectedType string
		if len(contentTypes) > 1 {
			fmt.Print("Select content type (or press Enter for default): ")
			input, err := readLine(reader)
			if err != nil {
				return fmt.Errorf("reading content type: %w", err)
			}

			if input == "" {
				selectedType = defaultType
			} else {
				// Try to match by number or name
				for i, ct := range contentTypes {
					if input == fmt.Sprintf("%d", i+1) || strings.EqualFold(input, ct.Name) {
						selectedType = ct.Name
						break
					}
				}
				if selectedType == "" {
					selectedType = input // Use as-is for new content type
				}
			}
		} else {
			selectedType = defaultType
		}

		// Get slug from args or prompt
		var slug string
		if len(args) > 0 {
			slug = args[0]
		} else {
			fmt.Print("Enter content slug (e.g., my-first-post): ")
			var err error
			slug, err = reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			slug = strings.TrimSpace(slug)
		}

		if slug == "" {
			return fmt.Errorf("slug cannot be empty")
		}

		// Validate slug
		if !isValidSlug(slug) {
			return fmt.Errorf("invalid slug: use only letters, numbers, hyphens, and underscores")
		}

		// Ensure .md extension
		if !strings.HasSuffix(slug, ".md") {
			slug += ".md"
		}

		// Build content path
		contentPath := filepath.Join(selectedType, slug)

		fmt.Printf("\n%s Creating: content/%s\n", icons.Pencil, contentPath)

		// Create content using Hugo
		if err := hugo.CreateContent(sitePath, contentPath); err != nil {
			return fmt.Errorf("failed to create content: %w", err)
		}

		createdFilePath := filepath.Join(sitePath, "content", contentPath)
		fmt.Printf("%s Content created: %s\n", icons.Success, createdFilePath)

		// Auto-build unless --no-build flag is set
		if !newNoBuild {
			fmt.Printf("\n%s Building site...\n", icons.Spinner)
			if err := hugo.BuildSite(sitePath); err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Build failed: %v\n", icons.Warning, err)
				fmt.Fprintf(os.Stderr, "   You can build manually with 'walgo build'\n")
			}
		}

		// Start dev server if --serve flag is set
		if newServe {
			fmt.Printf("\n%s Starting development server...\n", icons.Globe)
			fmt.Println("   Press Ctrl+C to stop")
			fmt.Println()
			return hugo.ServeSite(sitePath)
		}

		fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
		fmt.Printf("   - Edit: %s\n", createdFilePath)
		fmt.Println("   - Preview: walgo serve")
		fmt.Println("   - Deploy: walgo launch")

		return nil
	},
}

func isValidSlug(slug string) bool {
	// Remove .md extension for validation
	slug = strings.TrimSuffix(slug, ".md")
	validSlug := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validSlug.MatchString(slug) && len(slug) > 0 && len(slug) < 100
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().BoolVar(&newNoBuild, "no-build", false, "Skip automatic build after creating content")
	newCmd.Flags().BoolVar(&newServe, "serve", false, "Start development server after creating content")
}
