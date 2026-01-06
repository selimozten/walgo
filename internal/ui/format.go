package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// PrintSuccess prints a success message with icon
func PrintSuccess(message string) {
	icons := GetIcons()
	fmt.Printf("%s %s\n", icons.Success, message)
}

// PrintError prints an error message with icon
func PrintError(message string) {
	icons := GetIcons()
	fmt.Printf("%s %s\n", icons.Error, message)
}

// PrintWarning prints a warning message with icon
func PrintWarning(message string) {
	icons := GetIcons()
	fmt.Printf("%s %s\n", icons.Warning, message)
}

// PrintInfo prints an info message with icon
func PrintInfo(message string) {
	icons := GetIcons()
	fmt.Printf("%s %s\n", icons.Info, message)
}

// PrintStep prints a step indicator
func PrintStep(current, total int, message string) {
	fmt.Printf("  [%d/%d] %s\n", current, total, message)
}

// PrintCheck prints a checkmark with message
func PrintCheck(message string) {
	icons := GetIcons()
	fmt.Printf("  %s %s\n", icons.Check, message)
}

// PrintTip prints a helpful tip
func PrintTip(message string) {
	icons := GetIcons()
	fmt.Printf("\n%s %s\n", icons.Lightbulb, message)
}

// PrintSeparator prints a visual separator
func PrintSeparator() {
	fmt.Println(Separator())
}

// PrintBox prints a titled box
func PrintBox(title string) {
	top, middle, bottom := FormatBox(title)
	fmt.Println(top)
	fmt.Println(middle)
	fmt.Println(bottom)
}

// PrintHeader prints a section header
func PrintHeader(icon, title string) {
	if icon == "" {
		icon = GetIcons().Package
	}
	fmt.Printf("%s %s\n", icon, title)
	PrintSeparator()
}

// PrintNextSteps prints a list of next steps
func PrintNextSteps(steps []string) {
	icons := GetIcons()
	fmt.Printf("\n%s Next steps:\n", icons.Lightbulb)
	for _, step := range steps {
		fmt.Printf("   %s %s\n", icons.Arrow, step)
	}
	fmt.Println()
}

// PrintCommands prints a list of useful commands
func PrintCommands(title string, commands map[string]string) {
	icons := GetIcons()
	fmt.Printf("%s %s:\n", icons.Lightbulb, title)

	// Find max command length for alignment
	maxLen := 0
	for cmd := range commands {
		if len(cmd) > maxLen {
			maxLen = len(cmd)
		}
	}

	for cmd, desc := range commands {
		padding := strings.Repeat(" ", maxLen-len(cmd))
		fmt.Printf("   %s %s%s  # %s\n", icons.Arrow, cmd, padding, desc)
	}
	fmt.Println()
}

// GetIcon returns a specific icon by name
func GetIcon(name string) string {
	icons := GetIcons()

	switch name {
	case "success":
		return icons.Success
	case "error":
		return icons.Error
	case "warning":
		return icons.Warning
	case "info":
		return icons.Info
	case "question":
		return icons.Question
	case "spinner":
		return icons.Spinner
	case "key":
		return icons.Key
	case "garbage", "delete":
		return icons.Garbage
	case "rocket":
		return icons.Rocket
	case "package":
		return icons.Package
	case "folder":
		return icons.Folder
	case "file":
		return icons.File
	case "desktop":
		return icons.Desktop
	case "globe":
		return icons.Globe
	case "hourglass", "wait":
		return icons.Hourglass
	case "check":
		return icons.Check
	case "cross":
		return icons.Cross
	case "lightbulb", "tip":
		return icons.Lightbulb
	case "celebrate":
		return icons.Celebrate
	case "robot":
		return icons.Robot
	case "book":
		return icons.Book
	case "pencil":
		return icons.Pencil
	case "search":
		return icons.Search
	case "wrench":
		return icons.Wrench
	case "gear", "settings":
		return icons.Gear
	case "sparkles":
		return icons.Sparkles
	case "upload":
		return icons.Upload
	case "download":
		return icons.Download
	case "database":
		return icons.Database
	case "server":
		return icons.Server
	case "network":
		return icons.Network
	case "clipboard":
		return icons.Clipboard
	case "money":
		return icons.Money
	case "coin":
		return icons.Coin
	case "gas":
		return icons.Gas
	case "chart":
		return icons.Chart
	case "stats":
		return icons.Stats
	case "lock":
		return icons.Lock
	case "link":
		return icons.Link
	case "home":
		return icons.Home
	default:
		return ""
	}
}

// ReadLine reads a line of input from the reader and returns the trimmed result.
// Returns an error if reading fails (e.g., EOF, I/O error).
func ReadLine(reader *bufio.Reader) (string, error) {
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(input), nil
}

// ReadLineOrDefault reads a line of input, returning defaultValue if input is empty.
// Returns an error if reading fails.
func ReadLineOrDefault(reader *bufio.Reader, defaultValue string) (string, error) {
	input, err := ReadLine(reader)
	if err != nil {
		return "", err
	}
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

// PromptLine prints a prompt and reads a line of input.
// Returns an error if reading fails.
func PromptLine(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	return ReadLine(reader)
}

// PromptLineOrDefault prints a prompt and reads a line, returning defaultValue if empty.
// Returns an error if reading fails.
func PromptLineOrDefault(reader *bufio.Reader, prompt string, defaultValue string) (string, error) {
	fmt.Print(prompt)
	return ReadLineOrDefault(reader, defaultValue)
}
