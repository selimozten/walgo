package ui

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

// TestGetIcon tests the GetIcon function with table-driven tests
func TestGetIcon(t *testing.T) {
	// Save original
	origIcons := DefaultIcons
	defer func() {
		DefaultIcons = origIcons
	}()

	testCases := []struct {
		name      string
		iconName  string
		useEmoji  bool
		wantEmpty bool
		wantEmoji string // expected value in emoji mode
		wantASCII string // expected value in ASCII mode
	}{
		// Status icons
		{name: "success", iconName: "success", wantEmoji: EmojiIcons.Success, wantASCII: ASCIIIcons.Success},
		{name: "error", iconName: "error", wantEmoji: EmojiIcons.Error, wantASCII: ASCIIIcons.Error},
		{name: "warning", iconName: "warning", wantEmoji: EmojiIcons.Warning, wantASCII: ASCIIIcons.Warning},
		{name: "info", iconName: "info", wantEmoji: EmojiIcons.Info, wantASCII: ASCIIIcons.Info},
		{name: "question", iconName: "question", wantEmoji: EmojiIcons.Question, wantASCII: ASCIIIcons.Question},
		{name: "spinner", iconName: "spinner", wantEmoji: EmojiIcons.Spinner, wantASCII: ASCIIIcons.Spinner},
		{name: "key", iconName: "key", wantEmoji: EmojiIcons.Key, wantASCII: ASCIIIcons.Key},
		{name: "garbage", iconName: "garbage", wantEmoji: EmojiIcons.Garbage, wantASCII: ASCIIIcons.Garbage},
		{name: "delete alias", iconName: "delete", wantEmoji: EmojiIcons.Garbage, wantASCII: ASCIIIcons.Garbage},

		// Progress icons
		{name: "rocket", iconName: "rocket", wantEmoji: EmojiIcons.Rocket, wantASCII: ASCIIIcons.Rocket},
		{name: "package", iconName: "package", wantEmoji: EmojiIcons.Package, wantASCII: ASCIIIcons.Package},
		{name: "folder", iconName: "folder", wantEmoji: EmojiIcons.Folder, wantASCII: ASCIIIcons.Folder},
		{name: "file", iconName: "file", wantEmoji: EmojiIcons.File, wantASCII: ASCIIIcons.File},
		{name: "desktop", iconName: "desktop", wantEmoji: EmojiIcons.Desktop, wantASCII: ASCIIIcons.Desktop},
		{name: "globe", iconName: "globe", wantEmoji: EmojiIcons.Globe, wantASCII: ASCIIIcons.Globe},
		{name: "hourglass", iconName: "hourglass", wantEmoji: EmojiIcons.Hourglass, wantASCII: ASCIIIcons.Hourglass},
		{name: "wait alias", iconName: "wait", wantEmoji: EmojiIcons.Hourglass, wantASCII: ASCIIIcons.Hourglass},

		// Action icons
		{name: "check", iconName: "check", wantEmoji: EmojiIcons.Check, wantASCII: ASCIIIcons.Check},
		{name: "cross", iconName: "cross", wantEmoji: EmojiIcons.Cross, wantASCII: ASCIIIcons.Cross},
		{name: "lightbulb", iconName: "lightbulb", wantEmoji: EmojiIcons.Lightbulb, wantASCII: ASCIIIcons.Lightbulb},
		{name: "tip alias", iconName: "tip", wantEmoji: EmojiIcons.Lightbulb, wantASCII: ASCIIIcons.Lightbulb},
		{name: "celebrate", iconName: "celebrate", wantEmoji: EmojiIcons.Celebrate, wantASCII: ASCIIIcons.Celebrate},
		{name: "robot", iconName: "robot", wantEmoji: EmojiIcons.Robot, wantASCII: ASCIIIcons.Robot},
		{name: "book", iconName: "book", wantEmoji: EmojiIcons.Book, wantASCII: ASCIIIcons.Book},
		{name: "pencil", iconName: "pencil", wantEmoji: EmojiIcons.Pencil, wantASCII: ASCIIIcons.Pencil},
		{name: "search", iconName: "search", wantEmoji: EmojiIcons.Search, wantASCII: ASCIIIcons.Search},
		{name: "wrench", iconName: "wrench", wantEmoji: EmojiIcons.Wrench, wantASCII: ASCIIIcons.Wrench},
		{name: "gear", iconName: "gear", wantEmoji: EmojiIcons.Gear, wantASCII: ASCIIIcons.Gear},
		{name: "settings alias", iconName: "settings", wantEmoji: EmojiIcons.Gear, wantASCII: ASCIIIcons.Gear},
		{name: "sparkles", iconName: "sparkles", wantEmoji: EmojiIcons.Sparkles, wantASCII: ASCIIIcons.Sparkles},

		// Deployment icons
		{name: "upload", iconName: "upload", wantEmoji: EmojiIcons.Upload, wantASCII: ASCIIIcons.Upload},
		{name: "download", iconName: "download", wantEmoji: EmojiIcons.Download, wantASCII: ASCIIIcons.Download},
		{name: "database", iconName: "database", wantEmoji: EmojiIcons.Database, wantASCII: ASCIIIcons.Database},
		{name: "server", iconName: "server", wantEmoji: EmojiIcons.Server, wantASCII: ASCIIIcons.Server},
		{name: "network", iconName: "network", wantEmoji: EmojiIcons.Network, wantASCII: ASCIIIcons.Network},
		{name: "clipboard", iconName: "clipboard", wantEmoji: EmojiIcons.Clipboard, wantASCII: ASCIIIcons.Clipboard},

		// Finance icons
		{name: "money", iconName: "money", wantEmoji: EmojiIcons.Money, wantASCII: ASCIIIcons.Money},
		{name: "coin", iconName: "coin", wantEmoji: EmojiIcons.Coin, wantASCII: ASCIIIcons.Coin},
		{name: "gas", iconName: "gas", wantEmoji: EmojiIcons.Gas, wantASCII: ASCIIIcons.Gas},
		{name: "chart", iconName: "chart", wantEmoji: EmojiIcons.Chart, wantASCII: ASCIIIcons.Chart},
		{name: "stats", iconName: "stats", wantEmoji: EmojiIcons.Stats, wantASCII: ASCIIIcons.Stats},

		// Misc icons
		{name: "lock", iconName: "lock", wantEmoji: EmojiIcons.Lock, wantASCII: ASCIIIcons.Lock},
		{name: "link", iconName: "link", wantEmoji: EmojiIcons.Link, wantASCII: ASCIIIcons.Link},
		{name: "home", iconName: "home", wantEmoji: EmojiIcons.Home, wantASCII: ASCIIIcons.Home},

		// Unknown icon returns empty string
		{name: "unknown", iconName: "unknown", wantEmoji: "", wantASCII: "", wantEmpty: true},
		{name: "empty name", iconName: "", wantEmoji: "", wantASCII: "", wantEmpty: true},
		{name: "case sensitive", iconName: "SUCCESS", wantEmoji: "", wantASCII: "", wantEmpty: true},
		{name: "with spaces", iconName: " success ", wantEmoji: "", wantASCII: "", wantEmpty: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name+" emoji mode", func(t *testing.T) {
			UseEmoji()
			got := GetIcon(tc.iconName)
			if tc.wantEmpty {
				if got != "" {
					t.Errorf("GetIcon(%q) = %q, want empty string", tc.iconName, got)
				}
			} else if got != tc.wantEmoji {
				t.Errorf("GetIcon(%q) = %q, want %q", tc.iconName, got, tc.wantEmoji)
			}
		})

		t.Run(tc.name+" ASCII mode", func(t *testing.T) {
			UseASCII()
			got := GetIcon(tc.iconName)
			if tc.wantEmpty {
				if got != "" {
					t.Errorf("GetIcon(%q) = %q, want empty string", tc.iconName, got)
				}
			} else if got != tc.wantASCII {
				t.Errorf("GetIcon(%q) = %q, want %q", tc.iconName, got, tc.wantASCII)
			}
		})
	}
}

// TestReadLine tests the ReadLine function
func TestReadLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple input",
			input:    "hello\n",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "input with spaces",
			input:    "  hello world  \n",
			expected: "hello world",
			wantErr:  false,
		},
		{
			name:     "empty input",
			input:    "\n",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "only whitespace",
			input:    "   \n",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "input with tabs",
			input:    "\thello\t\n",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "input without newline (EOF)",
			input:    "hello",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "special characters",
			input:    "hello!@#$%^&*()\n",
			expected: "hello!@#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "unicode characters",
			input:    "hello \u4e16\u754c\n",
			expected: "hello \u4e16\u754c",
			wantErr:  false,
		},
		{
			name:     "multiple lines reads only first",
			input:    "first\nsecond\n",
			expected: "first",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			got, err := ReadLine(reader)

			if tc.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("ReadLine() = %q, want %q", got, tc.expected)
			}
		})
	}
}

// TestReadLineError tests ReadLine error handling
func TestReadLineError(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		errReader := &errorReader{err: io.ErrUnexpectedEOF}
		reader := bufio.NewReader(errReader)
		_, err := ReadLine(reader)
		if err == nil {
			t.Error("expected error but got nil")
		}
	})
}

// errorReader is a helper that always returns an error
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

// TestReadLineOrDefault tests the ReadLineOrDefault function
func TestReadLineOrDefault(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		defaultValue string
		expected     string
		wantErr      bool
	}{
		{
			name:         "uses input when provided",
			input:        "hello\n",
			defaultValue: "default",
			expected:     "hello",
			wantErr:      false,
		},
		{
			name:         "uses default when empty",
			input:        "\n",
			defaultValue: "default",
			expected:     "default",
			wantErr:      false,
		},
		{
			name:         "uses default when whitespace only",
			input:        "   \n",
			defaultValue: "default",
			expected:     "default",
			wantErr:      false,
		},
		{
			name:         "empty default with empty input",
			input:        "\n",
			defaultValue: "",
			expected:     "",
			wantErr:      false,
		},
		{
			name:         "input with spaces",
			input:        "  hello  \n",
			defaultValue: "default",
			expected:     "hello",
			wantErr:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			got, err := ReadLineOrDefault(reader, tc.defaultValue)

			if tc.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("ReadLineOrDefault() = %q, want %q", got, tc.expected)
			}
		})
	}
}

// TestReadLineOrDefaultError tests error propagation
func TestReadLineOrDefaultError(t *testing.T) {
	errReader := &errorReader{err: io.ErrUnexpectedEOF}
	reader := bufio.NewReader(errReader)
	_, err := ReadLineOrDefault(reader, "default")
	if err == nil {
		t.Error("expected error but got nil")
	}
}

// TestPromptLine tests the PromptLine function
func TestPromptLine(t *testing.T) {
	testCases := []struct {
		name     string
		prompt   string
		input    string
		expected string
	}{
		{
			name:     "simple prompt",
			prompt:   "Enter name: ",
			input:    "John\n",
			expected: "John",
		},
		{
			name:     "empty prompt",
			prompt:   "",
			input:    "test\n",
			expected: "test",
		},
		{
			name:     "prompt with special chars",
			prompt:   "Enter [value]: ",
			input:    "answer\n",
			expected: "answer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			got, err := PromptLine(reader, tc.prompt)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("PromptLine() = %q, want %q", got, tc.expected)
			}
		})
	}
}

// TestPromptLineOrDefault tests the PromptLineOrDefault function
func TestPromptLineOrDefault(t *testing.T) {
	testCases := []struct {
		name         string
		prompt       string
		defaultValue string
		input        string
		expected     string
	}{
		{
			name:         "uses input",
			prompt:       "Enter value: ",
			defaultValue: "default",
			input:        "custom\n",
			expected:     "custom",
		},
		{
			name:         "uses default on empty",
			prompt:       "Enter value [default]: ",
			defaultValue: "default",
			input:        "\n",
			expected:     "default",
		},
		{
			name:         "uses default on whitespace",
			prompt:       "Enter value: ",
			defaultValue: "default",
			input:        "   \n",
			expected:     "default",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			got, err := PromptLineOrDefault(reader, tc.prompt, tc.defaultValue)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("PromptLineOrDefault() = %q, want %q", got, tc.expected)
			}
		})
	}
}

// TestPrintFunctions tests print functions (output validation)
// Note: These functions print to stdout, so we're primarily testing they don't panic
func TestPrintFunctions(t *testing.T) {
	// Save original
	origIcons := DefaultIcons
	defer func() {
		DefaultIcons = origIcons
	}()

	// Test in both modes
	for _, mode := range []string{"emoji", "ascii"} {
		t.Run(mode+" mode", func(t *testing.T) {
			if mode == "emoji" {
				UseEmoji()
			} else {
				UseASCII()
			}

			// These should not panic
			t.Run("PrintSuccess", func(t *testing.T) {
				PrintSuccess("test message")
			})

			t.Run("PrintError", func(t *testing.T) {
				PrintError("test message")
			})

			t.Run("PrintWarning", func(t *testing.T) {
				PrintWarning("test message")
			})

			t.Run("PrintInfo", func(t *testing.T) {
				PrintInfo("test message")
			})

			t.Run("PrintStep", func(t *testing.T) {
				PrintStep(1, 5, "test step")
			})

			t.Run("PrintCheck", func(t *testing.T) {
				PrintCheck("test check")
			})

			t.Run("PrintTip", func(t *testing.T) {
				PrintTip("test tip")
			})

			t.Run("PrintSeparator", func(t *testing.T) {
				PrintSeparator()
			})

			t.Run("PrintBox", func(t *testing.T) {
				PrintBox("Test Box")
			})

			t.Run("PrintHeader with icon", func(t *testing.T) {
				PrintHeader(GetIcons().Rocket, "Test Header")
			})

			t.Run("PrintHeader without icon", func(t *testing.T) {
				PrintHeader("", "Test Header")
			})

			t.Run("PrintNextSteps", func(t *testing.T) {
				PrintNextSteps([]string{"Step 1", "Step 2", "Step 3"})
			})

			t.Run("PrintNextSteps empty", func(t *testing.T) {
				PrintNextSteps([]string{})
			})

			t.Run("PrintCommands", func(t *testing.T) {
				PrintCommands("Available Commands", map[string]string{
					"walgo init":   "Initialize project",
					"walgo deploy": "Deploy website",
				})
			})

			t.Run("PrintCommands empty", func(t *testing.T) {
				PrintCommands("No Commands", map[string]string{})
			})
		})
	}
}

// TestPrintStepEdgeCases tests edge cases for PrintStep
func TestPrintStepEdgeCases(t *testing.T) {
	testCases := []struct {
		name    string
		current int
		total   int
		message string
	}{
		{"zero values", 0, 0, "test"},
		{"negative current", -1, 5, "test"},
		{"current exceeds total", 10, 5, "test"},
		{"large numbers", 999999, 1000000, "large step"},
		{"empty message", 1, 5, ""},
		{"message with special chars", 1, 5, "test <>&\"'"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			PrintStep(tc.current, tc.total, tc.message)
		})
	}
}

// TestPrintNextStepsEdgeCases tests edge cases for PrintNextSteps
func TestPrintNextStepsEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		steps []string
	}{
		{"nil slice", nil},
		{"empty slice", []string{}},
		{"single step", []string{"Only step"}},
		{"many steps", []string{"Step 1", "Step 2", "Step 3", "Step 4", "Step 5"}},
		{"steps with special chars", []string{"Step with <html>", "Step with &special;"}},
		{"empty step string", []string{"Valid", "", "Also valid"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			PrintNextSteps(tc.steps)
		})
	}
}

// TestPrintCommandsEdgeCases tests edge cases for PrintCommands
func TestPrintCommandsEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		title    string
		commands map[string]string
	}{
		{"nil map", "Title", nil},
		{"empty map", "Title", map[string]string{}},
		{"single command", "Title", map[string]string{"cmd": "description"}},
		{"commands with varying lengths", "Title", map[string]string{
			"a":                 "short",
			"very-long-command": "description",
		}},
		{"empty title", "", map[string]string{"cmd": "desc"}},
		{"special chars in commands", "Title", map[string]string{
			"cmd --flag=<value>": "description with 'quotes'",
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			PrintCommands(tc.title, tc.commands)
		})
	}
}

// TestPrintHeaderEdgeCases tests edge cases for PrintHeader
func TestPrintHeaderEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		icon  string
		title string
	}{
		{"empty icon uses default", "", "Title"},
		{"empty title", GetIcons().Rocket, ""},
		{"both empty", "", ""},
		{"long title", GetIcons().Rocket, strings.Repeat("Long", 50)},
		{"special chars in title", GetIcons().Rocket, "<html>&amp;"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			PrintHeader(tc.icon, tc.title)
		})
	}
}

// TestPrintBoxEdgeCases tests edge cases for PrintBox
func TestPrintBoxEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		title       string
		shouldPanic bool
	}{
		{"empty title", "", false},
		{"short title", "Hi", false},
		{"exact width title", strings.Repeat("x", 55), false},
		{"unicode title", "\u4e16\u754c\u4f60\u597d", false},
		{"special chars", "<html>&amp;\"quotes\"", false},
		// Note: The original FormatBox function panics when title is too long
		// (exceeds box width) because it results in negative Repeat count.
		// This is documented as expected behavior for now.
		{"long title causes panic", strings.Repeat("x", 100), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic for long title but did not panic")
					}
				}()
			}
			PrintBox(tc.title)
		})
	}
}

// Benchmark tests
func BenchmarkGetIcon(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetIcon("success")
	}
}

func BenchmarkReadLine(b *testing.B) {
	input := "test input\n"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader := bufio.NewReader(strings.NewReader(input))
		_, _ = ReadLine(reader)
	}
}

func BenchmarkReadLineOrDefault(b *testing.B) {
	input := "\n"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader := bufio.NewReader(strings.NewReader(input))
		_, _ = ReadLineOrDefault(reader, "default")
	}
}

// TestConcurrentAccess tests thread safety of icon access
func TestConcurrentAccess(t *testing.T) {
	done := make(chan bool)

	// Run multiple goroutines accessing icons
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = GetIcons()
				_ = GetIcon("success")
				_ = Separator()
				_, _, _ = FormatBox("Test")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestReadLineWithBuffer tests ReadLine with buffered content
func TestReadLineWithBuffer(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("line1\n")
	buf.WriteString("line2\n")
	buf.WriteString("line3\n")

	reader := bufio.NewReader(&buf)

	// Read first line
	line1, err := ReadLine(reader)
	if err != nil {
		t.Errorf("unexpected error reading line1: %v", err)
	}
	if line1 != "line1" {
		t.Errorf("line1 = %q, want %q", line1, "line1")
	}

	// Read second line
	line2, err := ReadLine(reader)
	if err != nil {
		t.Errorf("unexpected error reading line2: %v", err)
	}
	if line2 != "line2" {
		t.Errorf("line2 = %q, want %q", line2, "line2")
	}

	// Read third line
	line3, err := ReadLine(reader)
	if err != nil {
		t.Errorf("unexpected error reading line3: %v", err)
	}
	if line3 != "line3" {
		t.Errorf("line3 = %q, want %q", line3, "line3")
	}
}

// TestGetIconAllNames tests that GetIcon returns correct values for all valid icon names
func TestGetIconAllNames(t *testing.T) {
	// Save original
	origIcons := DefaultIcons
	defer func() {
		DefaultIcons = origIcons
	}()

	UseEmoji()
	icons := GetIcons()

	// Map of icon names to their expected values
	expectedIcons := map[string]string{
		"success":   icons.Success,
		"error":     icons.Error,
		"warning":   icons.Warning,
		"info":      icons.Info,
		"question":  icons.Question,
		"spinner":   icons.Spinner,
		"key":       icons.Key,
		"garbage":   icons.Garbage,
		"delete":    icons.Garbage,
		"rocket":    icons.Rocket,
		"package":   icons.Package,
		"folder":    icons.Folder,
		"file":      icons.File,
		"desktop":   icons.Desktop,
		"globe":     icons.Globe,
		"hourglass": icons.Hourglass,
		"wait":      icons.Hourglass,
		"check":     icons.Check,
		"cross":     icons.Cross,
		"lightbulb": icons.Lightbulb,
		"tip":       icons.Lightbulb,
		"celebrate": icons.Celebrate,
		"robot":     icons.Robot,
		"book":      icons.Book,
		"pencil":    icons.Pencil,
		"search":    icons.Search,
		"wrench":    icons.Wrench,
		"gear":      icons.Gear,
		"settings":  icons.Gear,
		"sparkles":  icons.Sparkles,
		"upload":    icons.Upload,
		"download":  icons.Download,
		"database":  icons.Database,
		"server":    icons.Server,
		"network":   icons.Network,
		"clipboard": icons.Clipboard,
		"money":     icons.Money,
		"coin":      icons.Coin,
		"gas":       icons.Gas,
		"chart":     icons.Chart,
		"stats":     icons.Stats,
		"lock":      icons.Lock,
		"link":      icons.Link,
		"home":      icons.Home,
	}

	for name, expected := range expectedIcons {
		got := GetIcon(name)
		if got != expected {
			t.Errorf("GetIcon(%q) = %q, want %q", name, got, expected)
		}
	}
}

// TestPrintFunctionsWithEmptyMessage tests print functions with empty messages
func TestPrintFunctionsWithEmptyMessage(t *testing.T) {
	// These should not panic
	PrintSuccess("")
	PrintError("")
	PrintWarning("")
	PrintInfo("")
	PrintCheck("")
	PrintTip("")
}

// TestPrintFunctionsWithSpecialChars tests print functions with special characters
func TestPrintFunctionsWithSpecialChars(t *testing.T) {
	specialMessages := []string{
		"<script>alert('xss')</script>",
		"message with\nnewline",
		"message with\ttab",
		"message with unicode: \u4e16\u754c",
		"message with emoji: \U0001f600",
		`message with "quotes" and 'apostrophes'`,
		"message with backslash: \\n\\t",
	}

	for _, msg := range specialMessages {
		// These should not panic
		PrintSuccess(msg)
		PrintError(msg)
		PrintWarning(msg)
		PrintInfo(msg)
		PrintCheck(msg)
		PrintTip(msg)
	}
}
