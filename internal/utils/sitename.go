package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// SanitizeSiteName cleans a site name to be safe for directory creation
// It removes spaces, quotes, emojis, and other special characters,
// replacing them with hyphens while keeping the name readable.
func SanitizeSiteName(name string) string {
	if name == "" {
		return "site"
	}

	// Step 1: Remove quotes (single and double)
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, "'", "")

	// Step 2: Replace spaces and underscores with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Step 3: Remove emojis and other special characters (keep alphanumeric, hyphens, and basic punctuation)
	// We'll keep: letters (any language), numbers, hyphens, and dots
	var sanitized strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '.' {
			sanitized.WriteRune(r)
		}
	}

	name = sanitized.String()

	// Step 4: Collapse multiple consecutive hyphens into single hyphen
	multiHyphenRegex := regexp.MustCompile(`-+`)
	name = multiHyphenRegex.ReplaceAllString(name, "-")

	// Step 5: Remove hyphens and dots from start and end
	name = strings.Trim(name, "-.")

	// Step 6: Ensure it's not empty after sanitization
	if name == "" {
		return "site"
	}

	// Step 7: Limit length (max 100 characters)
	if len(name) > 100 {
		name = name[:100]
		// Make sure we don't cut in the middle of a character
		// and trim any trailing hyphens again
		name = strings.TrimRight(name, "-.")
	}

	// Step 8: Convert to lowercase for consistency
	name = strings.ToLower(name)

	return name
}

// ValidateSiteName checks if a site name is valid without sanitizing it
func ValidateSiteName(name string) error {
	if name == "" {
		return fmt.Errorf("site name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("site name too long (max 100 characters)")
	}

	// Check for invalid characters that could cause issues
	// NOTE: Dots are NOT allowed for site names to avoid path confusion
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("site name contains invalid characters (use only letters, numbers, hyphens, and underscores)")
	}

	// Check for reserved names
	reservedNames := []string{"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4",
		"com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3",
		"lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}
	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return fmt.Errorf("'%s' is a reserved system name", name)
		}
	}

	return nil
}

// IsValidSiteName is a convenience function that returns true if the name is valid
func IsValidSiteName(name string) bool {
	return ValidateSiteName(name) == nil
}
