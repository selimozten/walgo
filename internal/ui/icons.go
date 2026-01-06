package ui

import (
	"os"
	"runtime"
	"strings"
)

// Icons provides terminal-compatible icons/emojis with ASCII fallbacks
type Icons struct {
	// Status
	Success  string
	Error    string
	Warning  string
	Info     string
	Question string
	Spinner  string
	Key      string
	Garbage  string

	// Progress
	Rocket    string
	Package   string
	Folder    string
	File      string
	Desktop   string
	Globe     string
	Hourglass string

	// Actions
	Check     string
	Cross     string
	Lightbulb string
	Celebrate string
	Robot     string
	Book      string
	Pencil    string
	Search    string
	Wrench    string
	Gear      string
	Sparkles  string

	// Deployment
	Upload    string
	Download  string
	Database  string
	Server    string
	Network   string
	Clipboard string

	// Finance/Resources
	Money string
	Coin  string
	Gas   string
	Chart string
	Stats string

	// Misc
	Arrow     string
	Separator string
	BoxTop    string
	BoxBottom string
	BoxSide   string
	Lock      string
	Link      string
	Home      string
}

var (
	// DefaultIcons is the global icon set (auto-detected)
	DefaultIcons *Icons

	// EmojiIcons uses Unicode emojis (modern terminals)
	EmojiIcons = &Icons{
		Success:  "✅",
		Error:    "❌",
		Warning:  "⚠️ ",
		Info:     "ℹ️ ",
		Question: "❓",
		Spinner:  "🔄",
		Key:      "🔑",
		Garbage:  "🗑️ ",

		Rocket:    "🚀",
		Package:   "📦",
		Folder:    "📂",
		File:      "📄",
		Desktop:   "🖥️ ",
		Globe:     "🌐",
		Hourglass: "⏳",

		Check:     "✓",
		Cross:     "✗",
		Lightbulb: "💡",
		Celebrate: "🎉",
		Robot:     "🤖",
		Book:      "📚",
		Pencil:    "📝",
		Search:    "🔍",
		Wrench:    "🔧",
		Gear:      "⚙️ ",
		Sparkles:  "✨",

		Upload:    "📤",
		Download:  "📥",
		Database:  "💾",
		Server:    "🖥️ ",
		Network:   "🌐",
		Clipboard: "📋",

		Money: "💰",
		Coin:  "🪙",
		Gas:   "⛽",
		Chart: "📊",
		Stats: "📈",

		Arrow:     "→",
		Separator: "━",
		BoxTop:    "╔",
		BoxBottom: "╚",
		BoxSide:   "║",
		Lock:      "🔒",
		Link:      "🔗",
		Home:      "🏠",
	}

	// ASCIIIcons uses ASCII characters (compatible with all terminals)
	ASCIIIcons = &Icons{
		Success:  "[OK]",
		Error:    "[ERROR]",
		Warning:  "[WARN]",
		Info:     "[INFO]",
		Question: "[?]",
		Spinner:  "[...]",
		Key:      "[KEY]",
		Garbage:  "[DEL]",

		Rocket:    "[*]",
		Package:   "[PKG]",
		Folder:    "[DIR]",
		File:      "[FILE]",
		Desktop:   "[APP]",
		Globe:     "[WEB]",
		Hourglass: "[...]",

		Check:     "[+]",
		Cross:     "[X]",
		Lightbulb: "[!]",
		Celebrate: "[*]",
		Robot:     "[AI]",
		Book:      "[DOC]",
		Pencil:    "[EDIT]",
		Search:    "[?]",
		Wrench:    "[CFG]",
		Gear:      "[CFG]",
		Sparkles:  "[*]",

		Upload:    "[UP]",
		Download:  "[DOWN]",
		Database:  "[DB]",
		Server:    "[SRV]",
		Network:   "[NET]",
		Clipboard: "[>]",

		Money: "[$]",
		Coin:  "[C]",
		Gas:   "[GAS]",
		Chart: "[~]",
		Stats: "[~]",

		Arrow:     "->",
		Separator: "-",
		BoxTop:    "+",
		BoxBottom: "+",
		BoxSide:   "|",
		Lock:      "[*]",
		Link:      "[>]",
		Home:      "[~]",
	}
)

func init() {
	DefaultIcons = detectIconSupport()
}

// detectIconSupport detects if the terminal supports Unicode/emojis
func detectIconSupport() *Icons {
	// Check environment variable override
	if os.Getenv("WALGO_ASCII") == "1" || os.Getenv("WALGO_ASCII") == "true" {
		return ASCIIIcons
	}
	if os.Getenv("WALGO_EMOJI") == "1" || os.Getenv("WALGO_EMOJI") == "true" {
		return EmojiIcons
	}

	// Check if running in CI/CD environment
	if isCI() {
		return ASCIIIcons
	}

	// Check terminal type
	term := strings.ToLower(os.Getenv("TERM"))

	// Dumb terminals don't support Unicode well
	if term == "dumb" || term == "" {
		return ASCIIIcons
	}

	// Windows-specific detection
	if runtime.GOOS == "windows" {
		// Windows Terminal and modern terminals support Unicode
		if os.Getenv("WT_SESSION") != "" || os.Getenv("TERM_PROGRAM") == "vscode" {
			return EmojiIcons
		}

		// Check Windows version for modern console support
		// Windows 10+ with UTF-8 support
		if isModernWindows() {
			return EmojiIcons
		}

		// Fallback to ASCII for older Windows/CMD
		return ASCIIIcons
	}

	// Modern terminals (iTerm2, VSCode, Alacritty, etc.)
	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram == "iTerm.app" || termProgram == "vscode" ||
		termProgram == "Apple_Terminal" || termProgram == "Hyper" {
		return EmojiIcons
	}

	// Check for modern terminal emulators
	if strings.Contains(term, "256color") || strings.Contains(term, "xterm") ||
		strings.Contains(term, "screen") || strings.Contains(term, "tmux") {
		return EmojiIcons
	}

	// Default to emojis on Unix-like systems (macOS, Linux)
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		return EmojiIcons
	}

	// Fallback to ASCII for unknown environments
	return ASCIIIcons
}

// isCI checks if running in a CI/CD environment
func isCI() bool {
	ciEnvVars := []string{
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

	for _, envVar := range ciEnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// isModernWindows checks if running on Windows 10+ with UTF-8 support
func isModernWindows() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// Check for Windows Terminal or modern console
	if os.Getenv("WT_SESSION") != "" {
		return true
	}

	// Check ConEmu
	if os.Getenv("ConEmuPID") != "" {
		return true
	}

	// Check if UTF-8 is supported via chcp
	// This is a simple heuristic - modern Windows should support it
	return os.Getenv("LANG") != "" || os.Getenv("LC_ALL") != ""
}

// UseASCII forces ASCII mode
func UseASCII() {
	DefaultIcons = ASCIIIcons
}

// UseEmoji forces emoji mode
func UseEmoji() {
	DefaultIcons = EmojiIcons
}

// GetIcons returns the current icon set
func GetIcons() *Icons {
	return DefaultIcons
}

// FormatBox creates a box with the given title
func FormatBox(title string) (top, middle, bottom string) {
	icons := GetIcons()

	if icons == EmojiIcons {
		// Unicode box drawing
		width := 63
		padding := (width - len(title) - 4) / 2
		leftPad := strings.Repeat(" ", padding)
		rightPad := strings.Repeat(" ", width-len(title)-4-padding)

		top = "╔" + strings.Repeat("═", width-2) + "╗"
		middle = "║  " + leftPad + title + rightPad + "║"
		bottom = "╚" + strings.Repeat("═", width-2) + "╝"
	} else {
		// ASCII box drawing
		width := 63
		padding := (width - len(title) - 4) / 2
		leftPad := strings.Repeat(" ", padding)
		rightPad := strings.Repeat(" ", width-len(title)-4-padding)

		top = "+" + strings.Repeat("-", width-2) + "+"
		middle = "|  " + leftPad + title + rightPad + "|"
		bottom = "+" + strings.Repeat("-", width-2) + "+"
	}

	return
}

// Separator returns a visual separator line
func Separator() string {
	icons := GetIcons()
	if icons == EmojiIcons {
		return strings.Repeat("━", 44)
	}
	return strings.Repeat("-", 44)
}
