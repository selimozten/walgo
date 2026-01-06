package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

func TestServeCommand(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "Serve command help",
			Args:        []string{"serve", "--help"},
			ExpectError: false,
			Contains: []string{
				"Builds and serves",
				"hugo server",
				"--port",
				"--drafts",
			},
		},
	}

	runTestCases(t, rootCmd, tests)
}

func TestServeCommandFlags(t *testing.T) {
	// Find the serve command
	var serveCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "serve" {
			serveCommand = cmd
			break
		}
	}

	if serveCommand == nil {
		t.Fatal("serve command not found")
	}

	t.Run("port flag exists", func(t *testing.T) {
		portFlag := serveCommand.Flags().Lookup("port")
		if portFlag == nil {
			t.Error("port flag not found")
		} else {
			if portFlag.Shorthand != "p" {
				t.Errorf("Expected shorthand 'p', got '%s'", portFlag.Shorthand)
			}
			if portFlag.DefValue != "0" {
				t.Errorf("Expected default value '0', got '%s'", portFlag.DefValue)
			}
		}
	})

	t.Run("drafts flag exists", func(t *testing.T) {
		draftsFlag := serveCommand.Flags().Lookup("drafts")
		if draftsFlag == nil {
			t.Error("drafts flag not found")
		} else {
			if draftsFlag.Shorthand != "D" {
				t.Errorf("Expected shorthand 'D', got '%s'", draftsFlag.Shorthand)
			}
			if draftsFlag.DefValue != "false" {
				t.Errorf("Expected default value 'false', got '%s'", draftsFlag.DefValue)
			}
		}
	})

	t.Run("expired flag exists", func(t *testing.T) {
		expiredFlag := serveCommand.Flags().Lookup("expired")
		if expiredFlag == nil {
			t.Error("expired flag not found")
		} else {
			if expiredFlag.Shorthand != "E" {
				t.Errorf("Expected shorthand 'E', got '%s'", expiredFlag.Shorthand)
			}
		}
	})

	t.Run("future flag exists", func(t *testing.T) {
		futureFlag := serveCommand.Flags().Lookup("future")
		if futureFlag == nil {
			t.Error("future flag not found")
		} else {
			if futureFlag.Shorthand != "F" {
				t.Errorf("Expected shorthand 'F', got '%s'", futureFlag.Shorthand)
			}
		}
	})

	t.Run("unknown flags are allowed", func(t *testing.T) {
		// The serve command allows unknown flags to be passed to hugo server
		if !serveCommand.FParseErrWhitelist.UnknownFlags {
			t.Error("Unknown flags should be whitelisted for pass-through to hugo")
		}
	})
}

func TestFilterHugoOutput(t *testing.T) {
	icons := ui.GetIcons()

	tests := []struct {
		name     string
		input    string
		expected string
		contains []string
		excludes []string
	}{
		{
			name:     "Filter verbose output",
			input:    "Watching for changes\nBuilt in 100ms\nEnvironment: production",
			expected: "",
			excludes: []string{"Watching for", "Built in", "Environment:"},
		},
		{
			name:     "Keep error messages",
			input:    "error: something went wrong",
			contains: []string{icons.Error, "error: something went wrong"},
		},
		{
			name:     "Keep web server available",
			input:    "Web Server is available at http://localhost:1313/",
			contains: []string{icons.Success, "Web Server is available"},
		},
		{
			name:     "Keep change detected",
			input:    "Change detected, rebuilding site",
			contains: []string{icons.Info, "Change detected"},
		},
		{
			name:     "Filter empty lines",
			input:    "\n\n\n",
			expected: "",
		},
		{
			name:     "Filter hugo version line",
			input:    "hugo v0.123.0+extended darwin/amd64",
			expected: "",
			excludes: []string{"hugo v"},
		},
		{
			name:     "Filter table borders",
			input:    "│ Template │ Count │\n──────────────────",
			expected: "",
			excludes: []string{"│", "──────"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var writer bytes.Buffer

			filterHugoOutput(reader, &writer, icons)

			output := writer.String()

			if tt.expected != "" && output != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, output)
			}

			for _, contain := range tt.contains {
				if !strings.Contains(output, contain) {
					t.Errorf("Expected output to contain %q, got %q", contain, output)
				}
			}

			for _, exclude := range tt.excludes {
				if strings.Contains(output, exclude) {
					t.Errorf("Expected output to NOT contain %q, got %q", exclude, output)
				}
			}
		})
	}
}

func TestFilterHugoOutputSpecialCases(t *testing.T) {
	icons := ui.GetIcons()

	t.Run("Handle Syncing message", func(t *testing.T) {
		reader := strings.NewReader("Syncing files to /public/")
		var writer bytes.Buffer

		filterHugoOutput(reader, &writer, icons)

		output := writer.String()
		if !strings.Contains(output, icons.Info) {
			t.Error("Syncing message should have info icon")
		}
	})

	t.Run("Handle Error with capital E", func(t *testing.T) {
		reader := strings.NewReader("Error building site: template error")
		var writer bytes.Buffer

		filterHugoOutput(reader, &writer, icons)

		output := writer.String()
		if !strings.Contains(output, icons.Error) {
			t.Error("Error message should have error icon")
		}
	})

	t.Run("Pass through unknown lines", func(t *testing.T) {
		reader := strings.NewReader("Some random output line")
		var writer bytes.Buffer

		filterHugoOutput(reader, &writer, icons)

		output := writer.String()
		if !strings.Contains(output, "Some random output line") {
			t.Error("Unknown lines should be passed through")
		}
	})

	t.Run("Handle empty reader", func(t *testing.T) {
		reader := strings.NewReader("")
		var writer bytes.Buffer

		filterHugoOutput(reader, &writer, icons)

		output := writer.String()
		if output != "" {
			t.Errorf("Expected empty output, got %q", output)
		}
	})
}

func TestFilterHugoOutputWithLargeInput(t *testing.T) {
	icons := ui.GetIcons()

	t.Run("Handle large input", func(t *testing.T) {
		// Create a large input with many lines
		var lines []string
		for i := 0; i < 1000; i++ {
			lines = append(lines, "Watching for changes in /path/to/site")
			lines = append(lines, "Some output line")
		}
		input := strings.Join(lines, "\n")

		reader := strings.NewReader(input)
		var writer bytes.Buffer

		filterHugoOutput(reader, &writer, icons)

		// Should not panic and should filter appropriately
		output := writer.String()
		if strings.Contains(output, "Watching for") {
			t.Error("Watching lines should be filtered")
		}
	})
}

func TestServeCommandDescription(t *testing.T) {
	var serveCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "serve" {
			serveCommand = cmd
			break
		}
	}

	if serveCommand == nil {
		t.Fatal("serve command not found")
	}

	t.Run("Short description mentions Hugo server", func(t *testing.T) {
		if !strings.Contains(serveCommand.Short, "Hugo") {
			t.Error("Short description should mention Hugo")
		}
	})

	t.Run("Long description mentions Ctrl+C", func(t *testing.T) {
		if !strings.Contains(serveCommand.Long, "Ctrl+C") {
			t.Error("Long description should mention Ctrl+C to stop")
		}
	})
}

// mockReader implements io.Reader with controlled behavior
type mockReader struct {
	data  string
	index int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.index >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.index:])
	m.index += n
	return n, nil
}

func TestFilterHugoOutputWithMockReader(t *testing.T) {
	icons := ui.GetIcons()

	t.Run("Read from mock reader", func(t *testing.T) {
		reader := &mockReader{data: "Web Server is available at http://localhost:1313/\n"}
		var writer bytes.Buffer

		filterHugoOutput(reader, &writer, icons)

		output := writer.String()
		if !strings.Contains(output, "Web Server") {
			t.Error("Should contain web server message")
		}
	})
}
