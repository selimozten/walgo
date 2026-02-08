package optimizer

import (
	"regexp"
	"strings"
)

// Pre-compiled regexes for JS operations.
var jsBlockCommentRegex = regexp.MustCompile(`/\*[^*]*\*+(?:[^/*][^*]*\*+)*/`)

// JSOptimizer handles JavaScript optimization
type JSOptimizer struct {
	config JSConfig
}

// NewJSOptimizer creates a new JavaScript optimizer
func NewJSOptimizer(config JSConfig) *JSOptimizer {
	return &JSOptimizer{
		config: config,
	}
}

// Optimize optimizes JavaScript content
func (j *JSOptimizer) Optimize(content []byte) ([]byte, error) {
	if !j.config.Enabled {
		return content, nil
	}

	result := content

	// Remove comments
	if j.config.RemoveComments {
		result = j.removeComments(result)
	}

	// Minify JavaScript
	if j.config.MinifyJS {
		result = j.minifyJS(result)
	}

	return result, nil
}

// removeComments removes JavaScript comments
func (j *JSOptimizer) removeComments(content []byte) []byte {
	js := string(content)

	// Remove single-line comments, but preserve URLs and regex patterns
	lines := strings.Split(js, "\n")
	var processedLines []string

	for _, line := range lines {
		// Check if the line contains a string or regex that might have //
		inString := false
		inRegex := false
		escaped := false
		quote := byte(0)

		var cleanLine strings.Builder
		for i := 0; i < len(line); i++ {
			char := line[i]

			if escaped {
				cleanLine.WriteByte(char)
				escaped = false
				continue
			}

			if char == '\\' {
				escaped = true
				cleanLine.WriteByte(char)
				continue
			}

			if !inString && !inRegex {
				// Check for start of comment
				if i < len(line)-1 && line[i:i+2] == "//" {
					break // Rest of line is comment
				}
				// Check for start of string
				if char == '"' || char == '\'' || char == '`' {
					inString = true
					quote = char
				}
				// Check for start of regex (simplified)
				if char == '/' && i > 0 && (line[i-1] == '=' || line[i-1] == '(' || line[i-1] == '[' || line[i-1] == ',' || line[i-1] == ':' || line[i-1] == ';' || line[i-1] == '!' || line[i-1] == '&' || line[i-1] == '|' || line[i-1] == '?' || line[i-1] == '+' || line[i-1] == '-' || line[i-1] == '*' || line[i-1] == '%' || line[i-1] == '{' || line[i-1] == '}') {
					inRegex = true
				}
			} else if inString && char == quote {
				inString = false
				quote = 0
			} else if inRegex && char == '/' {
				inRegex = false
			}

			cleanLine.WriteByte(char)
		}

		// Always keep the line (even if empty after comment removal) to preserve
		// line structure. The minifier handles whitespace collapse.
		processedLines = append(processedLines, cleanLine.String())
	}

	js = strings.Join(processedLines, "\n")

	// Remove multi-line comments /* ... */
	js = jsBlockCommentRegex.ReplaceAllString(js, "")

	return []byte(js)
}

// minifyJS performs basic JavaScript minification
func (j *JSOptimizer) minifyJS(content []byte) []byte {
	js := string(content)

	// Remove unnecessary whitespace while preserving string contents (including template literals)
	// This function handles all minification in a string-aware manner
	js = j.preserveStringsAndMinify(js)

	// Trim leading and trailing whitespace
	js = strings.TrimSpace(js)

	return []byte(js)
}

// preserveStringsAndMinify minifies JavaScript while preserving string contents.
// Newlines are preserved (collapsed to a single \n) to maintain JavaScript's
// Automatic Semicolon Insertion (ASI) semantics. Without this, constructs like
// "return\nvalue" would become "return value" (returns value) instead of the
// correct "return;\nvalue" (returns undefined).
func (j *JSOptimizer) preserveStringsAndMinify(js string) string {
	var result strings.Builder
	inString := false
	inRegex := false
	stringChar := byte(0)
	var lastWritten byte
	i := 0

	for i < len(js) {
		char := js[i]

		// Handle escape sequences
		if i < len(js)-1 && char == '\\' {
			result.WriteByte(char)
			lastWritten = char
			i++
			if i < len(js) {
				result.WriteByte(js[i])
				lastWritten = js[i]
			}
			i++
			continue
		}

		// Handle strings
		if !inRegex && (char == '"' || char == '\'' || char == '`') {
			if !inString {
				inString = true
				stringChar = char
			} else if char == stringChar {
				inString = false
				stringChar = 0
			}
			result.WriteByte(char)
			lastWritten = char
			i++
			continue
		}

		// Handle regex (simplified detection)
		if !inString && char == '/' {
			// Look ahead to see if this might be a regex
			if i > 0 {
				prevChar := js[i-1]
				if prevChar == '=' || prevChar == '(' || prevChar == '[' || prevChar == ',' {
					inRegex = true
				}
			}
			if inRegex && i < len(js)-1 {
				// Look for end of regex
				k := i + 1
				for k < len(js) && js[k] != '/' {
					if js[k] == '\\' && k+1 < len(js) {
						k += 2 // Skip escaped character
					} else {
						k++
					}
				}
				if k < len(js) {
					// Copy the entire regex
					result.WriteString(js[i : k+1])
					lastWritten = js[k]
					i = k + 1
					inRegex = false
					continue
				}
			}
		}

		// If we're in a string or regex, preserve everything
		if inString || inRegex {
			result.WriteByte(char)
			lastWritten = char
			i++
			continue
		}

		// Handle newlines — preserve them to maintain ASI behavior.
		// Collapse consecutive newlines and \r\n into a single \n.
		if char == '\n' || char == '\r' {
			if char == '\r' && i+1 < len(js) && js[i+1] == '\n' {
				i++ // Skip \r in \r\n
			}
			// Only emit newline if previous output wasn't already a newline
			if lastWritten != '\n' {
				result.WriteByte('\n')
				lastWritten = '\n'
			}
			i++
			continue
		}

		// Handle horizontal whitespace (spaces, tabs) outside of strings
		if char == ' ' || char == '\t' {
			// Only add space if the previous and next characters need separation
			if result.Len() > 0 && i < len(js)-1 {
				nextChar := js[i+1]

				// Check if space is needed between identifiers/keywords
				if isAlphaNumeric(lastWritten) && isAlphaNumeric(nextChar) {
					result.WriteByte(' ')
					lastWritten = ' '
				}
			}
			i++
			continue
		}

		result.WriteByte(char)
		lastWritten = char
		i++
	}

	return result.String()
}

// isAlphaNumeric checks if a character is alphanumeric or underscore
func isAlphaNumeric(char byte) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_' || char == '$'
}
