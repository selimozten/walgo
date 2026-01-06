package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/selimozten/walgo/internal/ai"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// aiConfigureCmd configures AI provider credentials.
var aiConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure AI provider credentials",
	Long: `Configure AI provider credentials for content generation.

Supported providers:
  - openai: OpenAI API (gpt-4, gpt-3.5-turbo, etc.)
  - openrouter: OpenRouter API (access to multiple models including Claude, GPT-4, etc.)

Credentials are stored securely in ~/.walgo/ai-credentials.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("%s AI Configuration Setup\n", icons.Robot)
		fmt.Println()

		fmt.Println("Select AI provider:")
		fmt.Println("  1) OpenAI (gpt-4, gpt-3.5-turbo, etc.)")
		fmt.Println("  2) OpenRouter (access to Claude, GPT-4, and more)")
		fmt.Println()

		providerChoice, err := ui.PromptLineOrDefault(reader, "Select [1]: ", "1")
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}

		var provider string
		switch providerChoice {
		case "1":
			provider = "openai"
		case "2":
			provider = "openrouter"
		default:
			return fmt.Errorf("invalid selection: %s", providerChoice)
		}

		fmt.Println()
		apiKey, err := ui.PromptLine(reader, fmt.Sprintf("Enter your %s API key: ", strings.ToUpper(provider)))
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		fmt.Println()
		defaultURL := ai.GetDefaultBaseURL(provider)
		baseURL, err := ui.PromptLineOrDefault(reader, fmt.Sprintf("Custom base URL [%s]: ", defaultURL), defaultURL)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}

		fmt.Println()
		defaultModel := "gpt-4"
		modelExamples := "gpt-4, gpt-4o, gpt-4-turbo, gpt-3.5-turbo"
		if provider == "openrouter" {
			defaultModel = "openai/gpt-4"
			modelExamples = "openai/gpt-4, anthropic/claude-3.5-sonnet, google/gemini-pro"
		}
		fmt.Printf("Enter model name (e.g., %s)\n", modelExamples)
		modelName, err := ui.PromptLineOrDefault(reader, fmt.Sprintf("Model [%s]: ", defaultModel), defaultModel)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}

		if err := ai.SetProviderCredentials(provider, apiKey, baseURL, modelName); err != nil {
			return fmt.Errorf("saving credentials: %w", err)
		}

		credPath, _ := ai.GetCredentialsPath()
		fmt.Printf("\n%s AI configuration complete!\n", icons.Success)
		fmt.Printf("   Provider: %s\n", provider)
		fmt.Printf("   Model: %s\n", modelName)
		fmt.Printf("   Credentials: %s\n", credPath)
		fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
		fmt.Printf("   - Generate content: walgo ai generate\n")
		fmt.Printf("   - Create new site: walgo ai pipeline\n")

		return nil
	},
}
