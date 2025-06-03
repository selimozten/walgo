package cmd

import (
	"fmt"
	"os"

	"walgo/internal/walrus"

	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert <object-id>",
	Short: "Convert a Walrus Site object ID from hex to Base36 format.",
	Long: `Converts a Walrus Site object ID from hexadecimal format to Base36 format.
This is useful to get the subdomain representation of your site for direct access
without a SuiNS domain.

Example: walgo convert 0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf

The Base36 format can be used to access your site directly via:
https://[base36-id].wal.app (if the portal supports Base36 subdomains)`,
	Args: cobra.ExactArgs(1), // Expects exactly one argument: object-id
	Run: func(cmd *cobra.Command, args []string) {
		objectID := args[0]
		fmt.Printf("Converting object ID: %s\n", objectID)

		base36ID, err := walrus.ConvertObjectID(objectID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting object ID: %v\n", err)
			os.Exit(1)
		}

		if base36ID != "" {
			fmt.Printf("\n✅ Conversion successful!\n")
			fmt.Printf("🔗 Base36 ID: %s\n", base36ID)
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
