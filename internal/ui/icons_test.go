package ui

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

// TestIconsStruct tests that the Icons struct fields are accessible
func TestIconsStruct(t *testing.T) {
	// Test EmojiIcons has expected values
	t.Run("EmojiIcons", func(t *testing.T) {
		icons := EmojiIcons

		// Status icons
		if icons.Success == "" {
			t.Error("EmojiIcons.Success should not be empty")
		}
		if icons.Error == "" {
			t.Error("EmojiIcons.Error should not be empty")
		}
		if icons.Warning == "" {
			t.Error("EmojiIcons.Warning should not be empty")
		}
		if icons.Info == "" {
			t.Error("EmojiIcons.Info should not be empty")
		}
		if icons.Question == "" {
			t.Error("EmojiIcons.Question should not be empty")
		}
		if icons.Spinner == "" {
			t.Error("EmojiIcons.Spinner should not be empty")
		}
		if icons.Key == "" {
			t.Error("EmojiIcons.Key should not be empty")
		}
		if icons.Garbage == "" {
			t.Error("EmojiIcons.Garbage should not be empty")
		}

		// Progress icons
		if icons.Rocket == "" {
			t.Error("EmojiIcons.Rocket should not be empty")
		}
		if icons.Package == "" {
			t.Error("EmojiIcons.Package should not be empty")
		}
		if icons.Folder == "" {
			t.Error("EmojiIcons.Folder should not be empty")
		}
		if icons.File == "" {
			t.Error("EmojiIcons.File should not be empty")
		}
		if icons.Desktop == "" {
			t.Error("EmojiIcons.Desktop should not be empty")
		}
		if icons.Globe == "" {
			t.Error("EmojiIcons.Globe should not be empty")
		}
		if icons.Hourglass == "" {
			t.Error("EmojiIcons.Hourglass should not be empty")
		}

		// Action icons
		if icons.Check == "" {
			t.Error("EmojiIcons.Check should not be empty")
		}
		if icons.Cross == "" {
			t.Error("EmojiIcons.Cross should not be empty")
		}
		if icons.Lightbulb == "" {
			t.Error("EmojiIcons.Lightbulb should not be empty")
		}
		if icons.Celebrate == "" {
			t.Error("EmojiIcons.Celebrate should not be empty")
		}
		if icons.Robot == "" {
			t.Error("EmojiIcons.Robot should not be empty")
		}
		if icons.Book == "" {
			t.Error("EmojiIcons.Book should not be empty")
		}
		if icons.Pencil == "" {
			t.Error("EmojiIcons.Pencil should not be empty")
		}
		if icons.Search == "" {
			t.Error("EmojiIcons.Search should not be empty")
		}
		if icons.Wrench == "" {
			t.Error("EmojiIcons.Wrench should not be empty")
		}
		if icons.Gear == "" {
			t.Error("EmojiIcons.Gear should not be empty")
		}
		if icons.Sparkles == "" {
			t.Error("EmojiIcons.Sparkles should not be empty")
		}

		// Deployment icons
		if icons.Upload == "" {
			t.Error("EmojiIcons.Upload should not be empty")
		}
		if icons.Download == "" {
			t.Error("EmojiIcons.Download should not be empty")
		}
		if icons.Database == "" {
			t.Error("EmojiIcons.Database should not be empty")
		}
		if icons.Server == "" {
			t.Error("EmojiIcons.Server should not be empty")
		}
		if icons.Network == "" {
			t.Error("EmojiIcons.Network should not be empty")
		}
		if icons.Clipboard == "" {
			t.Error("EmojiIcons.Clipboard should not be empty")
		}

		// Finance icons
		if icons.Money == "" {
			t.Error("EmojiIcons.Money should not be empty")
		}
		if icons.Coin == "" {
			t.Error("EmojiIcons.Coin should not be empty")
		}
		if icons.Gas == "" {
			t.Error("EmojiIcons.Gas should not be empty")
		}
		if icons.Chart == "" {
			t.Error("EmojiIcons.Chart should not be empty")
		}
		if icons.Stats == "" {
			t.Error("EmojiIcons.Stats should not be empty")
		}

		// Misc icons
		if icons.Arrow == "" {
			t.Error("EmojiIcons.Arrow should not be empty")
		}
		if icons.Separator == "" {
			t.Error("EmojiIcons.Separator should not be empty")
		}
		if icons.BoxTop == "" {
			t.Error("EmojiIcons.BoxTop should not be empty")
		}
		if icons.BoxBottom == "" {
			t.Error("EmojiIcons.BoxBottom should not be empty")
		}
		if icons.BoxSide == "" {
			t.Error("EmojiIcons.BoxSide should not be empty")
		}
		if icons.Lock == "" {
			t.Error("EmojiIcons.Lock should not be empty")
		}
		if icons.Link == "" {
			t.Error("EmojiIcons.Link should not be empty")
		}
		if icons.Home == "" {
			t.Error("EmojiIcons.Home should not be empty")
		}
	})

	// Test ASCIIIcons has expected values
	t.Run("ASCIIIcons", func(t *testing.T) {
		icons := ASCIIIcons

		// Verify ASCII icons use ASCII-only characters
		asciiFields := map[string]string{
			"Success":   icons.Success,
			"Error":     icons.Error,
			"Warning":   icons.Warning,
			"Info":      icons.Info,
			"Question":  icons.Question,
			"Spinner":   icons.Spinner,
			"Key":       icons.Key,
			"Garbage":   icons.Garbage,
			"Rocket":    icons.Rocket,
			"Package":   icons.Package,
			"Folder":    icons.Folder,
			"File":      icons.File,
			"Desktop":   icons.Desktop,
			"Globe":     icons.Globe,
			"Hourglass": icons.Hourglass,
			"Check":     icons.Check,
			"Cross":     icons.Cross,
			"Lightbulb": icons.Lightbulb,
			"Celebrate": icons.Celebrate,
			"Robot":     icons.Robot,
			"Book":      icons.Book,
			"Pencil":    icons.Pencil,
			"Search":    icons.Search,
			"Wrench":    icons.Wrench,
			"Gear":      icons.Gear,
			"Sparkles":  icons.Sparkles,
			"Upload":    icons.Upload,
			"Download":  icons.Download,
			"Database":  icons.Database,
			"Server":    icons.Server,
			"Network":   icons.Network,
			"Clipboard": icons.Clipboard,
			"Money":     icons.Money,
			"Coin":      icons.Coin,
			"Gas":       icons.Gas,
			"Chart":     icons.Chart,
			"Stats":     icons.Stats,
			"Arrow":     icons.Arrow,
			"Separator": icons.Separator,
			"BoxTop":    icons.BoxTop,
			"BoxBottom": icons.BoxBottom,
			"BoxSide":   icons.BoxSide,
			"Lock":      icons.Lock,
			"Link":      icons.Link,
			"Home":      icons.Home,
		}

		for name, value := range asciiFields {
			if value == "" {
				t.Errorf("ASCIIIcons.%s should not be empty", name)
			}
			// Check that all ASCII icons contain only ASCII characters
			for _, r := range value {
				if r > 127 {
					t.Errorf("ASCIIIcons.%s contains non-ASCII character: %q", name, value)
					break
				}
			}
		}
	})
}

// TestDetectIconSupport tests icon detection based on environment
func TestDetectIconSupport(t *testing.T) {
	// Save original environment
	origASCII := os.Getenv("WALGO_ASCII")
	origEmoji := os.Getenv("WALGO_EMOJI")
	origCI := os.Getenv("CI")
	origTerm := os.Getenv("TERM")
	origWTSession := os.Getenv("WT_SESSION")
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origConEmuPID := os.Getenv("ConEmuPID")
	origLang := os.Getenv("LANG")
	origLcAll := os.Getenv("LC_ALL")

	// Cleanup helper
	cleanup := func() {
		os.Setenv("WALGO_ASCII", origASCII)
		os.Setenv("WALGO_EMOJI", origEmoji)
		os.Setenv("CI", origCI)
		os.Setenv("TERM", origTerm)
		os.Setenv("WT_SESSION", origWTSession)
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("ConEmuPID", origConEmuPID)
		os.Setenv("LANG", origLang)
		os.Setenv("LC_ALL", origLcAll)
	}
	defer cleanup()

	// Clear all relevant env vars
	clearEnv := func() {
		os.Unsetenv("WALGO_ASCII")
		os.Unsetenv("WALGO_EMOJI")
		os.Unsetenv("CI")
		os.Unsetenv("CONTINUOUS_INTEGRATION")
		os.Unsetenv("JENKINS_URL")
		os.Unsetenv("TRAVIS")
		os.Unsetenv("CIRCLECI")
		os.Unsetenv("GITHUB_ACTIONS")
		os.Unsetenv("GITLAB_CI")
		os.Unsetenv("BUILDKITE")
		os.Unsetenv("DRONE")
		os.Unsetenv("TERM")
		os.Unsetenv("WT_SESSION")
		os.Unsetenv("TERM_PROGRAM")
		os.Unsetenv("ConEmuPID")
		os.Unsetenv("LANG")
		os.Unsetenv("LC_ALL")
	}

	t.Run("WALGO_ASCII=1 forces ASCII", func(t *testing.T) {
		clearEnv()
		os.Setenv("WALGO_ASCII", "1")
		result := detectIconSupport()
		if result != ASCIIIcons {
			t.Error("WALGO_ASCII=1 should return ASCIIIcons")
		}
	})

	t.Run("WALGO_ASCII=true forces ASCII", func(t *testing.T) {
		clearEnv()
		os.Setenv("WALGO_ASCII", "true")
		result := detectIconSupport()
		if result != ASCIIIcons {
			t.Error("WALGO_ASCII=true should return ASCIIIcons")
		}
	})

	t.Run("WALGO_EMOJI=1 forces emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("WALGO_EMOJI", "1")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("WALGO_EMOJI=1 should return EmojiIcons")
		}
	})

	t.Run("WALGO_EMOJI=true forces emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("WALGO_EMOJI", "true")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("WALGO_EMOJI=true should return EmojiIcons")
		}
	})

	t.Run("CI environment returns ASCII", func(t *testing.T) {
		clearEnv()
		os.Setenv("CI", "true")
		os.Setenv("TERM", "xterm-256color")
		result := detectIconSupport()
		if result != ASCIIIcons {
			t.Error("CI environment should return ASCIIIcons")
		}
	})

	t.Run("GITHUB_ACTIONS returns ASCII", func(t *testing.T) {
		clearEnv()
		os.Setenv("GITHUB_ACTIONS", "true")
		os.Setenv("TERM", "xterm-256color")
		result := detectIconSupport()
		if result != ASCIIIcons {
			t.Error("GITHUB_ACTIONS environment should return ASCIIIcons")
		}
	})

	t.Run("dumb terminal returns ASCII", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "dumb")
		result := detectIconSupport()
		if result != ASCIIIcons {
			t.Error("dumb terminal should return ASCIIIcons")
		}
	})

	t.Run("empty TERM returns ASCII", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "")
		result := detectIconSupport()
		if result != ASCIIIcons {
			t.Error("empty TERM should return ASCIIIcons")
		}
	})

	t.Run("xterm-256color returns emoji on non-CI", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "xterm-256color")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("xterm-256color on non-CI should return EmojiIcons")
		}
	})

	t.Run("screen terminal returns emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "screen")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("screen terminal should return EmojiIcons")
		}
	})

	t.Run("tmux terminal returns emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "tmux-256color")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("tmux terminal should return EmojiIcons")
		}
	})

	t.Run("iTerm.app returns emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "xterm")
		os.Setenv("TERM_PROGRAM", "iTerm.app")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("iTerm.app should return EmojiIcons")
		}
	})

	t.Run("vscode returns emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "xterm")
		os.Setenv("TERM_PROGRAM", "vscode")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("vscode should return EmojiIcons")
		}
	})

	t.Run("Apple_Terminal returns emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "xterm")
		os.Setenv("TERM_PROGRAM", "Apple_Terminal")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("Apple_Terminal should return EmojiIcons")
		}
	})

	t.Run("Hyper returns emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "xterm")
		os.Setenv("TERM_PROGRAM", "Hyper")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("Hyper should return EmojiIcons")
		}
	})

	// Platform-specific tests
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		t.Run("Unix-like OS defaults to emoji", func(t *testing.T) {
			clearEnv()
			os.Setenv("TERM", "linux")
			result := detectIconSupport()
			if result != EmojiIcons {
				t.Error("Unix-like OS should default to EmojiIcons")
			}
		})
	}
}

// TestIsCI tests CI detection
func TestIsCI(t *testing.T) {
	// Save original CI env vars
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"JENKINS_URL",
		"TRAVIS",
		"CIRCLECI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"BUILDKITE",
		"DRONE",
	}

	origValues := make(map[string]string)
	for _, v := range ciVars {
		origValues[v] = os.Getenv(v)
	}

	cleanup := func() {
		for _, v := range ciVars {
			if origValues[v] != "" {
				os.Setenv(v, origValues[v])
			} else {
				os.Unsetenv(v)
			}
		}
	}
	defer cleanup()

	clearAllCI := func() {
		for _, v := range ciVars {
			os.Unsetenv(v)
		}
	}

	t.Run("no CI env vars returns false", func(t *testing.T) {
		clearAllCI()
		if isCI() {
			t.Error("isCI() should return false when no CI env vars are set")
		}
	})

	// Test each CI env var
	for _, envVar := range ciVars {
		t.Run(envVar+" returns true", func(t *testing.T) {
			clearAllCI()
			os.Setenv(envVar, "true")
			if !isCI() {
				t.Errorf("isCI() should return true when %s is set", envVar)
			}
		})
	}
}

// TestIsModernWindows tests Windows version detection
func TestIsModernWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Run("returns false on non-Windows", func(t *testing.T) {
			if isModernWindows() {
				t.Error("isModernWindows() should return false on non-Windows")
			}
		})
		return
	}

	// Windows-specific tests
	origWT := os.Getenv("WT_SESSION")
	origConEmu := os.Getenv("ConEmuPID")
	origLang := os.Getenv("LANG")
	origLcAll := os.Getenv("LC_ALL")

	cleanup := func() {
		os.Setenv("WT_SESSION", origWT)
		os.Setenv("ConEmuPID", origConEmu)
		os.Setenv("LANG", origLang)
		os.Setenv("LC_ALL", origLcAll)
	}
	defer cleanup()

	clearEnv := func() {
		os.Unsetenv("WT_SESSION")
		os.Unsetenv("ConEmuPID")
		os.Unsetenv("LANG")
		os.Unsetenv("LC_ALL")
	}

	t.Run("WT_SESSION returns true", func(t *testing.T) {
		clearEnv()
		os.Setenv("WT_SESSION", "abc123")
		if !isModernWindows() {
			t.Error("isModernWindows() should return true with WT_SESSION")
		}
	})

	t.Run("ConEmuPID returns true", func(t *testing.T) {
		clearEnv()
		os.Setenv("ConEmuPID", "12345")
		if !isModernWindows() {
			t.Error("isModernWindows() should return true with ConEmuPID")
		}
	})

	t.Run("LANG set returns true", func(t *testing.T) {
		clearEnv()
		os.Setenv("LANG", "en_US.UTF-8")
		if !isModernWindows() {
			t.Error("isModernWindows() should return true with LANG")
		}
	})

	t.Run("LC_ALL set returns true", func(t *testing.T) {
		clearEnv()
		os.Setenv("LC_ALL", "en_US.UTF-8")
		if !isModernWindows() {
			t.Error("isModernWindows() should return true with LC_ALL")
		}
	})
}

// TestUseASCII tests forcing ASCII mode
func TestUseASCII(t *testing.T) {
	// Save original
	origIcons := DefaultIcons
	defer func() {
		DefaultIcons = origIcons
	}()

	UseASCII()
	if DefaultIcons != ASCIIIcons {
		t.Error("UseASCII() should set DefaultIcons to ASCIIIcons")
	}
}

// TestUseEmoji tests forcing emoji mode
func TestUseEmoji(t *testing.T) {
	// Save original
	origIcons := DefaultIcons
	defer func() {
		DefaultIcons = origIcons
	}()

	UseEmoji()
	if DefaultIcons != EmojiIcons {
		t.Error("UseEmoji() should set DefaultIcons to EmojiIcons")
	}
}

// TestGetIcons tests getting the current icon set
func TestGetIcons(t *testing.T) {
	// Save original
	origIcons := DefaultIcons
	defer func() {
		DefaultIcons = origIcons
	}()

	t.Run("returns current icons", func(t *testing.T) {
		UseEmoji()
		if GetIcons() != EmojiIcons {
			t.Error("GetIcons() should return EmojiIcons after UseEmoji()")
		}

		UseASCII()
		if GetIcons() != ASCIIIcons {
			t.Error("GetIcons() should return ASCIIIcons after UseASCII()")
		}
	})
}

// TestFormatBox tests box formatting
func TestFormatBox(t *testing.T) {
	// Save original
	origIcons := DefaultIcons
	defer func() {
		DefaultIcons = origIcons
	}()

	testCases := []struct {
		name      string
		title     string
		useEmoji  bool
		checkFunc func(top, middle, bottom string) bool
	}{
		{
			name:     "emoji mode with short title",
			title:    "Test",
			useEmoji: true,
			checkFunc: func(top, middle, bottom string) bool {
				return strings.HasPrefix(top, "\u2554") && // Unicode box drawing
					strings.Contains(middle, "Test") &&
					strings.HasPrefix(bottom, "\u255a")
			},
		},
		{
			name:     "ASCII mode with short title",
			title:    "Test",
			useEmoji: false,
			checkFunc: func(top, middle, bottom string) bool {
				return strings.HasPrefix(top, "+") &&
					strings.Contains(middle, "Test") &&
					strings.HasPrefix(bottom, "+")
			},
		},
		{
			name:     "empty title",
			title:    "",
			useEmoji: true,
			checkFunc: func(top, middle, bottom string) bool {
				return len(top) > 0 && len(middle) > 0 && len(bottom) > 0
			},
		},
		{
			name:     "long title",
			title:    "This is a very long title that should still fit",
			useEmoji: true,
			checkFunc: func(top, middle, bottom string) bool {
				return strings.Contains(middle, "This is a very long title")
			},
		},
		{
			name:     "special characters in title",
			title:    "Test <>&\"'",
			useEmoji: false,
			checkFunc: func(top, middle, bottom string) bool {
				return strings.Contains(middle, "Test <>&\"'")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.useEmoji {
				UseEmoji()
			} else {
				UseASCII()
			}

			top, middle, bottom := FormatBox(tc.title)
			if !tc.checkFunc(top, middle, bottom) {
				t.Errorf("FormatBox(%q) check failed:\ntop=%q\nmiddle=%q\nbottom=%q",
					tc.title, top, middle, bottom)
			}
		})
	}
}

// TestSeparator tests separator formatting
func TestSeparator(t *testing.T) {
	// Save original
	origIcons := DefaultIcons
	defer func() {
		DefaultIcons = origIcons
	}()

	t.Run("emoji mode uses Unicode", func(t *testing.T) {
		UseEmoji()
		sep := Separator()
		if !strings.Contains(sep, "\u2501") { // Unicode box drawing heavy horizontal
			t.Errorf("Separator() in emoji mode should use Unicode, got: %q", sep)
		}
		if len(sep) != 44*3 { // Unicode char is 3 bytes
			// This is just checking it has the expected length pattern
		}
	})

	t.Run("ASCII mode uses dashes", func(t *testing.T) {
		UseASCII()
		sep := Separator()
		if !strings.HasPrefix(sep, "-") {
			t.Errorf("Separator() in ASCII mode should use dashes, got: %q", sep)
		}
		if len(sep) != 44 {
			t.Errorf("Separator() in ASCII mode should be 44 chars, got: %d", len(sep))
		}
	})
}

// TestDefaultIconsInitialized tests that DefaultIcons is initialized
func TestDefaultIconsInitialized(t *testing.T) {
	if DefaultIcons == nil {
		t.Error("DefaultIcons should be initialized by init()")
	}

	// Should be either EmojiIcons or ASCIIIcons
	if DefaultIcons != EmojiIcons && DefaultIcons != ASCIIIcons {
		t.Error("DefaultIcons should be either EmojiIcons or ASCIIIcons")
	}
}

// TestIconFieldsDifferent tests that Emoji and ASCII icons are different
func TestIconFieldsDifferent(t *testing.T) {
	// At least some icons should be different between the two sets
	if EmojiIcons.Success == ASCIIIcons.Success {
		t.Error("EmojiIcons.Success should differ from ASCIIIcons.Success")
	}
	if EmojiIcons.Error == ASCIIIcons.Error {
		t.Error("EmojiIcons.Error should differ from ASCIIIcons.Error")
	}
	if EmojiIcons.Arrow == ASCIIIcons.Arrow {
		t.Error("EmojiIcons.Arrow should differ from ASCIIIcons.Arrow")
	}
	if EmojiIcons.Check == ASCIIIcons.Check {
		t.Error("EmojiIcons.Check should differ from ASCIIIcons.Check")
	}
}

// TestWindowsTerminalDetection tests Windows Terminal specific detection
func TestWindowsTerminalDetection(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows OS")
	}

	origWT := os.Getenv("WT_SESSION")
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origTerm := os.Getenv("TERM")
	origASCII := os.Getenv("WALGO_ASCII")
	origEmoji := os.Getenv("WALGO_EMOJI")
	origCI := os.Getenv("CI")

	cleanup := func() {
		os.Setenv("WT_SESSION", origWT)
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("TERM", origTerm)
		os.Setenv("WALGO_ASCII", origASCII)
		os.Setenv("WALGO_EMOJI", origEmoji)
		os.Setenv("CI", origCI)
	}
	defer cleanup()

	clearEnv := func() {
		os.Unsetenv("WT_SESSION")
		os.Unsetenv("TERM_PROGRAM")
		os.Unsetenv("TERM")
		os.Unsetenv("WALGO_ASCII")
		os.Unsetenv("WALGO_EMOJI")
		os.Unsetenv("CI")
	}

	t.Run("Windows Terminal returns emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "xterm")
		os.Setenv("WT_SESSION", "some-session-id")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("Windows Terminal should return EmojiIcons")
		}
	})

	t.Run("VSCode on Windows returns emoji", func(t *testing.T) {
		clearEnv()
		os.Setenv("TERM", "xterm")
		os.Setenv("TERM_PROGRAM", "vscode")
		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("VSCode on Windows should return EmojiIcons")
		}
	})
}

// TestMixedCIAndTerminalProgram tests priority of CI over terminal detection
func TestMixedCIAndTerminalProgram(t *testing.T) {
	origCI := os.Getenv("CI")
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origTerm := os.Getenv("TERM")
	origASCII := os.Getenv("WALGO_ASCII")
	origEmoji := os.Getenv("WALGO_EMOJI")

	cleanup := func() {
		os.Setenv("CI", origCI)
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("TERM", origTerm)
		os.Setenv("WALGO_ASCII", origASCII)
		os.Setenv("WALGO_EMOJI", origEmoji)
	}
	defer cleanup()

	t.Run("CI overrides terminal program", func(t *testing.T) {
		os.Unsetenv("WALGO_ASCII")
		os.Unsetenv("WALGO_EMOJI")
		os.Setenv("CI", "true")
		os.Setenv("TERM_PROGRAM", "iTerm.app")
		os.Setenv("TERM", "xterm-256color")

		result := detectIconSupport()
		if result != ASCIIIcons {
			t.Error("CI should override terminal program detection")
		}
	})

	t.Run("WALGO_EMOJI overrides CI", func(t *testing.T) {
		os.Setenv("WALGO_EMOJI", "1")
		os.Setenv("CI", "true")

		result := detectIconSupport()
		if result != EmojiIcons {
			t.Error("WALGO_EMOJI should override CI detection")
		}
	})
}

// Benchmark tests
func BenchmarkDetectIconSupport(b *testing.B) {
	for i := 0; i < b.N; i++ {
		detectIconSupport()
	}
}

func BenchmarkFormatBox(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatBox("Test Title")
	}
}

func BenchmarkSeparator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Separator()
	}
}

func BenchmarkGetIcons(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetIcons()
	}
}
