package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/selimozten/walgo/internal/ai"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// aiUpdateCmd updates existing Hugo content using AI.
var aiUpdateCmd = &cobra.Command{
	Use:   "update <file>",
	Short: "Update existing Hugo content using AI",
	Long: `Update an existing Hugo content file using AI.

The AI will modify the file based on your instructions while preserving the original structure and style.

Example:
  walgo ai update content/posts/my-post.md

Note: This command will update the content of the specified file using AI.
The file will be saved with the updated content.
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		reader := bufio.NewReader(os.Stdin)
		icons := ui.GetIcons()

		client, provider, model, err := ai.LoadClient(ai.LongRequestTimeout)
		if err != nil {
			fmt.Printf("\n%s Run 'walgo ai configure' to set up AI features\n", icons.Lightbulb)
			return err
		}

		existingContent, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		fmt.Printf("%s AI Content Updater (%s: %s)\n", icons.Robot, provider, model)
		fmt.Printf("%s File: %s\n", icons.File, filePath)
		fmt.Println()

		fmt.Print("Describe what changes you want to make: ")
		instruction, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("reading instruction: %w", err)
		}

		if instruction == "" {
			return fmt.Errorf("instruction cannot be empty")
		}

		fmt.Printf("\n%s Updating content...\n", icons.Spinner)

		userPrompt := ai.BuildUpdatePrompt(instruction, string(existingContent))

		updatedContent, err := client.GenerateContent(ai.SystemPromptContentUpdate, userPrompt)
		if err != nil {
			return fmt.Errorf("updating content: %w", err)
		}

		updatedContent = ai.CleanGeneratedContent(updatedContent)

		fmt.Printf("\n%s Content updated!\n", icons.Success)
		fmt.Println()
		fmt.Print("Save changes to file? [Y/n]: ")
		confirm, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("reading confirmation: %w", err)
		}
		confirm = strings.ToLower(confirm)

		if confirm != "" && confirm != "y" && confirm != "yes" {
			fmt.Printf("%s Changes not saved\n", icons.Info)
			return nil
		}

		if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
			return fmt.Errorf("saving file: %w", err)
		}

		fmt.Printf("\n%s File updated: %s\n", icons.Success, filePath)

		// Apply content fixer to ensure YAML frontmatter is correct
		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		fmt.Printf("\n%s Fixing YAML frontmatter...\n", icons.Spinner)
		themeName := hugo.GetThemeName(sitePath)
		fixer := ai.NewContentFixerWithTheme(sitePath, hugo.DetectSiteType(sitePath), themeName)
		if err := fixer.FixAll(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Content fix failed: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("%s YAML frontmatter validated\n", icons.Check)
		}

		err = hugo.BuildSite(sitePath)
		if err != nil {
			return fmt.Errorf("failed to build site: %w", err)
		}

		fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
		fmt.Println("   - Preview: walgo serve")
		fmt.Println("   - Deploy: walgo launch")

		return nil
	},
}
