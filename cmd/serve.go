package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// killExistingHugoProcesses is implemented in serve_unix.go and serve_windows.go

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the Hugo site locally using Hugo's built-in server.",
	Long: `Builds and serves your Hugo site locally using 'hugo server'.
This command is a wrapper around 'hugo server' and supports common flags.
The server will typically be available at http://localhost:1313 (or the port you specify).
Any unrecognized flags will be passed directly to 'hugo server'.
Press Ctrl+C to stop the server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		fmt.Printf("%s Starting local Hugo development server...\n", icons.Rocket)

		// Kill any existing Hugo serve processes
		if err := killExistingHugoProcesses(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Error cleaning up existing Hugo processes: %v\n", icons.Warning, err)
		}

		// Check if Hugo is installed
		hugoPath, err := deps.LookPath("hugo")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Hugo is not installed or not found in PATH\n", icons.Error)
			fmt.Fprintf(os.Stderr, "\n%s Install Hugo: https://gohugo.io/installation/\n", icons.Lightbulb)
			return fmt.Errorf("hugo is not installed or not found in PATH")
		}

		sitePath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: Cannot determine current directory: %v\n", icons.Error, err)
			return fmt.Errorf("cannot determine current directory: %w", err)
		}

		hugoArgs := []string{"server"}

		// Append flags recognized by walgo serve
		drafts, err := cmd.Flags().GetBool("drafts")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading drafts flag: %v\n", err)
			return fmt.Errorf("error reading drafts flag: %w", err)
		}
		if drafts {
			hugoArgs = append(hugoArgs, "-D")
		}
		expired, err := cmd.Flags().GetBool("expired")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading expired flag: %v\n", err)
			return fmt.Errorf("error reading expired flag: %w", err)
		}
		if expired {
			hugoArgs = append(hugoArgs, "-E")
		}
		future, err := cmd.Flags().GetBool("future")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading future flag: %v\n", err)
			return fmt.Errorf("error reading future flag: %w", err)
		}
		if future {
			hugoArgs = append(hugoArgs, "-F")
		}
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading port flag: %v\n", err)
			return fmt.Errorf("error reading port flag: %w", err)
		}
		if port != 0 {
			hugoArgs = append(hugoArgs, "--port", strconv.Itoa(port))
		}

		err = hugo.BuildSite(sitePath)
		if err != nil {
			return fmt.Errorf("failed to build site: %w", err)
		}

		// Append any remaining arguments (including unrecognized flags) to be passed to hugo server
		hugoArgs = append(hugoArgs, args...)

		hugoCmd := exec.Command(hugoPath, hugoArgs...)
		hugoCmd.Dir = sitePath

		// Capture stdout and stderr to filter verbose output
		stdout, err := hugoCmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}
		stderr, err := hugoCmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to create stderr pipe: %w", err)
		}

		if err := hugoCmd.Start(); err != nil {
			return fmt.Errorf("failed to start hugo server: %w", err)
		}

		// Filter and display output
		go filterHugoOutput(stdout, os.Stdout, icons)
		go filterHugoOutput(stderr, os.Stderr, icons)

		if err := hugoCmd.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "\n%s Hugo server stopped: %v\n", icons.Warning, err)
		} else {
			fmt.Printf("\n%s Hugo server stopped\n", icons.Success)
		}

		return nil
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

// filterHugoOutput filters Hugo server output to show only essential info
func filterHugoOutput(r io.Reader, w io.Writer, icons *ui.Icons) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip verbose lines
		if strings.HasPrefix(line, "Watching for") ||
			strings.HasPrefix(line, "Start building") ||
			strings.HasPrefix(line, "hugo v") ||
			strings.Contains(line, "──────") ||
			strings.Contains(line, "│") ||
			strings.HasPrefix(line, "Built in") ||
			strings.HasPrefix(line, "Environment:") ||
			strings.HasPrefix(line, "Serving pages") ||
			strings.HasPrefix(line, "Running in Fast") ||
			strings.TrimSpace(line) == "" {
			continue
		}

		// Show important lines with icons
		if strings.Contains(line, "Web Server is available at") {
			fmt.Fprintf(w, "%s %s\n", icons.Success, line)
			fmt.Fprintf(w, "%s Press Ctrl+C to stop\n", icons.Lightbulb)
		} else if strings.Contains(line, "error") || strings.Contains(line, "Error") {
			fmt.Fprintf(w, "%s %s\n", icons.Error, line)
		} else if strings.Contains(line, "Change detected") || strings.Contains(line, "Syncing") {
			fmt.Fprintf(w, "%s %s\n", icons.Info, line)
		} else {
			fmt.Fprintln(w, line)
		}
	}
}
