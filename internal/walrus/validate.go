package walrus

import (
	"fmt"
	"regexp"
	"strings"
)

// validateObjectID ensures objectID is hexadecimal to prevent command injection.
func validateObjectID(objectID string) error {
	if objectID == "" {
		return fmt.Errorf("object ID cannot be empty")
	}

	validObjectID := regexp.MustCompile(`^(0x)?[0-9a-fA-F]{64}$`)
	if !validObjectID.MatchString(objectID) {
		return fmt.Errorf("invalid object ID format: %s (must be 64 hex characters, optionally prefixed with 0x)", objectID)
	}

	if strings.ContainsAny(objectID, "\x00\n\r;|&$`(){}[]<>") {
		return fmt.Errorf("object ID contains invalid characters")
	}

	return nil
}
