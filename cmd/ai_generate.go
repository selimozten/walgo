package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/selimozten/walgo/internal/ai"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// aiGenerateCmd generates new Hugo content files using AI with automatic content type detection.
var aiGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate new Hugo content using AI",
	Long: `Generate new Hugo content files using AI with automatic content type detection.

The AI will create properly formatted Hugo markdown files with frontmatter based on your instructions.
Content type is automatically detected from your Hugo site structure.

Examples:
  walgo ai generate                    # Interactive generation with auto-detect
  walgo ai generate --serve            # Generate and start dev server
  walgo ai generate --no-build         # Generate without building`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		reader := bufio.NewReader(os.Stdin)

		client, provider, model, err := ai.LoadClient(ai.LongRequestTimeout)
		if err != nil {
			fmt.Printf("\n%s Run 'walgo ai configure' to set up AI features\n", icons.Lightbulb)
			return err
		}

		sitePath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		fmt.Printf("%s AI Content Generator (%s: %s)\n", icons.Robot, provider, model)
		fmt.Println()

		// Get content structure
		structure, err := ai.GetContentStructure(sitePath)
		if err != nil {
			return fmt.Errorf("failed to get content structure: %w", err)
		}

		// Display content structure
		fmt.Printf("%s Current content structure:\n", icons.Info)
		if len(structure.ContentTypes) > 0 {
			for _, ct := range structure.ContentTypes {
				marker := " "
				if ct.Name == structure.DefaultType {
					marker = "*"
				}
				fmt.Printf("  %s %s/ (%d files)\n", marker, ct.Name, ct.FileCount)
				if len(ct.Files) > 0 && len(ct.Files) <= 3 {
					for _, f := range ct.Files {
						fmt.Printf("      - %s\n", f)
					}
				}
			}
			fmt.Printf("\n  %s Default type: %s\n", icons.Check, structure.DefaultType)
		} else {
			fmt.Printf("  No content types found, will create appropriate structure\n")
		}
		fmt.Println()

		// Get instructions
		fmt.Printf("%s What content do you want to create?\n", icons.Lightbulb)
		fmt.Println("  Examples:")
		fmt.Println("  - Create a blog post about blockchain technology for beginners")
		fmt.Println("  - Write a tutorial on deploying Hugo sites to Walrus")
		fmt.Println("  - Generate an about page for my portfolio")
		fmt.Println()
		fmt.Print("Your instructions: ")
		instructions, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("reading instructions: %w", err)
		}

		if instructions == "" {
			return fmt.Errorf("instructions cannot be empty")
		}

		fmt.Printf("\n%s Generating content (AI will determine type and filename)...\n", icons.Spinner)

		// Use the new smart content generator
		generator := ai.NewContentGenerator(client)
		result := generator.GenerateContent(ai.ContentGenerationParams{
			SitePath:     sitePath,
			Instructions: instructions,
			Context:      context.Background(),
		})

		if !result.Success {
			return fmt.Errorf("generation failed: %s", result.ErrorMessage)
		}

		fmt.Printf("\n%s Content generated successfully!\n", icons.Success)
		fmt.Printf("   Type: %s\n", result.ContentType)
		fmt.Printf("   File: %s\n", result.Filename)
		fmt.Printf("   Path: %s\n", result.FilePath)

		// Apply content fixer to ensure YAML frontmatter is correct
		fmt.Printf("\n%s Fixing YAML frontmatter...\n", icons.Spinner)
		themeName := hugo.GetThemeName(sitePath)
		fixer := ai.NewContentFixerWithTheme(sitePath, hugo.DetectSiteType(sitePath), themeName)
		if err := fixer.FixAll(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Content fix failed: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("%s YAML frontmatter validated\n", icons.Check)
		}

		if !aiGenerateNoBuild {
			fmt.Printf("\n%s Building site...\n", icons.Spinner)
			if err := hugo.BuildSite(sitePath); err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Build failed: %v\n", icons.Warning, err)
			}
		}

		if aiGenerateServe {
			fmt.Printf("\n%s Starting development server...\n", icons.Globe)
			fmt.Println("   Press Ctrl+C to stop")
			fmt.Println()
			return hugo.ServeSite(sitePath)
		}

		fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
		fmt.Printf("   - Review: %s\n", result.FilePath)
		fmt.Println("   - Preview: walgo serve")
		fmt.Println("   - Deploy: walgo launch")

		return nil
	},
}
