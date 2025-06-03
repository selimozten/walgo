package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the Hugo site locally using Hugo's built-in server.",
	Long: `Builds and serves the Hugo site locally using 'hugo server'.
This command is a wrapper around 'hugo server' and supports common flags.
The server will typically be available at http://localhost:1313 (or the port you specify).
Any unrecognized flags will be passed directly to 'hugo server'.
Press Ctrl+C to stop the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting local Hugo development server...")

		// Check if Hugo is installed
		if _, err := exec.LookPath("hugo"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Hugo is not installed or not found in PATH. Please install Hugo first.\n")
			os.Exit(1)
		}

		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		hugoArgs := []string{"server"}

		// Append flags recognized by walgo serve
		if drafts, _ := cmd.Flags().GetBool("drafts"); drafts {
			hugoArgs = append(hugoArgs, "-D")
		}
		if expired, _ := cmd.Flags().GetBool("expired"); expired {
			hugoArgs = append(hugoArgs, "-E")
		}
		if future, _ := cmd.Flags().GetBool("future"); future {
			hugoArgs = append(hugoArgs, "-F")
		}
		if port, _ := cmd.Flags().GetInt("port"); port != 0 {
			hugoArgs = append(hugoArgs, "--port", strconv.Itoa(port))
		}

		// Append any remaining arguments (including unrecognized flags) to be passed to hugo server
		hugoArgs = append(hugoArgs, args...)

		fmt.Printf("Executing: hugo %s\n", strings.Join(hugoArgs, " "))

		hugoCmd := exec.Command("hugo", hugoArgs...)
		hugoCmd.Dir = sitePath
		hugoCmd.Stdout = os.Stdout
		hugoCmd.Stderr = os.Stderr

		if err := hugoCmd.Run(); err != nil {
			// Hugo server usually runs until Ctrl+C.
			// A non-nil error here can be a normal termination (e.g., via signal) or an actual error.
			// We print the message, but don't os.Exit(1) as user-initiated stop is expected.
			fmt.Fprintf(os.Stderr, "Hugo server process ended: %v\n", err)
		} else {
			fmt.Println("Hugo server process ended.")
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().BoolP("drafts", "D", false, "Include content marked as draft (passed to 'hugo server -D')")
	serveCmd.Flags().BoolP("expired", "E", false, "Include content with expiry date in the past (passed to 'hugo server -E')")
	serveCmd.Flags().BoolP("future", "F", false, "Include content with publishdate in the future (passed to 'hugo server -F')")
	serveCmd.Flags().IntP("port", "p", 0, "Port for Hugo server (e.g., 1313). If 0 or not set, Hugo's default (usually 1313) is used.")

	// Allow unknown flags to be passed through to hugo server
	serveCmd.FParseErrWhitelist.UnknownFlags = true
}
