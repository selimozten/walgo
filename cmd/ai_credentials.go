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

// aiGetCmd shows current AI credentials.
var aiGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show current AI credentials",
	Long: `Display the currently configured AI provider credentials.

Shows provider, model, and base URL (API key is masked for security).

Example:
  walgo ai get`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()

		providers, err := ai.ListProviders()
		if err != nil {
			return fmt.Errorf("failed to list providers: %w", err)
		}

		if len(providers) == 0 {
			fmt.Printf("%s No AI credentials configured\n", icons.Warning)
			fmt.Printf("\n%s Run 'walgo ai configure' to set up AI credentials\n", icons.Lightbulb)
			return nil
		}

		fmt.Printf("%s AI Credentials\n", icons.Robot)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for _, provider := range providers {
			creds, err := ai.GetProviderCredentials(provider)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Warning: Could not load credentials for %s: %v\n", icons.Warning, provider, err)
				continue
			}

			fmt.Printf("\n%s Provider: %s\n", icons.Check, provider)
			fmt.Printf("   Model:    %s\n", creds.Model)
			fmt.Printf("   Base URL: %s\n", creds.BaseURL)

			maskedKey := "****"
			if len(creds.APIKey) > 8 {
				maskedKey = creds.APIKey[:4] + "..." + creds.APIKey[len(creds.APIKey)-4:]
			}
			fmt.Printf("   API Key:  %s\n", maskedKey)
		}

		credPath, _ := ai.GetCredentialsPath()
		fmt.Printf("\n%s Credentials file: %s\n", icons.File, credPath)

		return nil
	},
}

// aiRemoveCmd removes AI credentials.
var aiRemoveCmd = &cobra.Command{
	Use:   "remove [provider]",
	Short: "Remove AI credentials",
	Long: `Remove AI provider credentials.

If no provider is specified, removes all credentials.
If a provider is specified, removes only that provider's credentials.

Examples:
  walgo ai remove           # Remove all credentials
  walgo ai remove openai    # Remove only OpenAI credentials
  walgo ai remove openrouter # Remove only OpenRouter credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		reader := bufio.NewReader(os.Stdin)

		if len(args) > 0 {
			provider := args[0]
			fmt.Printf("Remove credentials for %s? [y/N]: ", provider)
			confirm, err := readLine(reader)
			if err != nil {
				return fmt.Errorf("reading confirmation: %w", err)
			}
			confirm = strings.ToLower(confirm)

			if confirm != "y" && confirm != "yes" {
				fmt.Printf("%s Cancelled\n", icons.Cross)
				return nil
			}

			if err := ai.RemoveProviderCredentials(provider); err != nil {
				return fmt.Errorf("failed to remove credentials: %w", err)
			}

			fmt.Printf("%s Removed credentials for %s\n", icons.Success, provider)
		} else {
			providers, err := ai.ListProviders()
			if err != nil {
				return fmt.Errorf("failed to list providers: %w", err)
			}

			if len(providers) == 0 {
				fmt.Printf("%s No credentials to remove\n", icons.Info)
				return nil
			}

			fmt.Printf("Remove all AI credentials (%s)? [y/N]: ", strings.Join(providers, ", "))
			confirm, err := readLine(reader)
			if err != nil {
				return fmt.Errorf("reading confirmation: %w", err)
			}
			confirm = strings.ToLower(confirm)

			if confirm != "y" && confirm != "yes" {
				fmt.Printf("%s Cancelled\n", icons.Cross)
				return nil
			}

			if err := ai.RemoveAllCredentials(); err != nil {
				return fmt.Errorf("failed to remove credentials: %w", err)
			}

			fmt.Printf("%s Removed all AI credentials\n", icons.Success)
		}

		return nil
	},
}
