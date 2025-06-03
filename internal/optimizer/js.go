package optimizer

import (
	"regexp"
	"strings"
)

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

	// Basic obfuscation (if enabled)
	if j.config.Obfuscate {
		result = j.obfuscateBasic(result)
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

		cleanLine := ""
		for i := 0; i < len(line); i++ {
			char := line[i]

			if escaped {
				cleanLine += string(char)
				escaped = false
				continue
			}

			if char == '\\' {
				escaped = true
				cleanLine += string(char)
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

			cleanLine += string(char)
		}

		if strings.TrimSpace(cleanLine) != "" {
			processedLines = append(processedLines, cleanLine)
		}
	}

	js = strings.Join(processedLines, "\n")

	// Remove multi-line comments /* ... */
	// This is a simplified approach - in production, you'd need a proper parser
	blockCommentRegex := regexp.MustCompile(`/\*[^*]*\*+(?:[^/*][^*]*\*+)*/`)
	js = blockCommentRegex.ReplaceAllString(js, "")

	return []byte(js)
}

// minifyJS performs basic JavaScript minification
func (j *JSOptimizer) minifyJS(content []byte) []byte {
	js := string(content)

	// Remove unnecessary whitespace while preserving string contents
	js = j.preserveStringsAndMinify(js)

	// Remove unnecessary semicolons before }
	js = regexp.MustCompile(`;\s*}`).ReplaceAllString(js, "}")

	// Remove unnecessary whitespace around operators and punctuation
	js = regexp.MustCompile(`\s*([{}();,])\s*`).ReplaceAllString(js, "$1")
	js = regexp.MustCompile(`\s*([=+\-*/%<>!&|^])\s*`).ReplaceAllString(js, "$1")

	// Compress consecutive newlines and spaces
	js = regexp.MustCompile(`\s+`).ReplaceAllString(js, " ")

	// Trim leading and trailing whitespace
	js = strings.TrimSpace(js)

	return []byte(js)
}

// preserveStringsAndMinify minifies JavaScript while preserving string contents
func (j *JSOptimizer) preserveStringsAndMinify(js string) string {
	var result strings.Builder
	inString := false
	inRegex := false
	stringChar := byte(0)
	i := 0

	for i < len(js) {
		char := js[i]

		// Handle escape sequences
		if i < len(js)-1 && char == '\\' {
			result.WriteByte(char)
			i++
			if i < len(js) {
				result.WriteByte(js[i])
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
				j := i + 1
				for j < len(js) && js[j] != '/' {
					if js[j] == '\\' && j+1 < len(js) {
						j += 2 // Skip escaped character
					} else {
						j++
					}
				}
				if j < len(js) {
					// Copy the entire regex
					result.WriteString(js[i : j+1])
					i = j + 1
					inRegex = false
					continue
				}
			}
		}

		// If we're in a string or regex, preserve everything
		if inString || inRegex {
			result.WriteByte(char)
			i++
			continue
		}

		// Handle whitespace outside of strings
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			// Only add space if the previous and next characters need separation
			if result.Len() > 0 && i < len(js)-1 {
				prevChar := result.String()[result.Len()-1]
				nextChar := js[i+1]

				// Check if space is needed between identifiers/keywords
				if isAlphaNumeric(prevChar) && isAlphaNumeric(nextChar) {
					result.WriteByte(' ')
				}
			}
			i++
			continue
		}

		result.WriteByte(char)
		i++
	}

	return result.String()
}

// obfuscateBasic performs basic variable name obfuscation
func (j *JSOptimizer) obfuscateBasic(content []byte) []byte {
	// This is a very basic obfuscation - in production, you'd use a proper parser
	js := string(content)

	// Map of common variable names to shorter versions
	// This is a simplified approach and may break code
	obfuscationMap := map[string]string{
		"function": "f",
		"variable": "v",
		"element":  "e",
		"document": "d",
		"window":   "w",
	}

	// Only obfuscate if explicitly enabled and user understands the risks
	for original, obfuscated := range obfuscationMap {
		// Use word boundaries to avoid partial matches
		wordRegex := regexp.MustCompile(`\b` + original + `\b`)
		js = wordRegex.ReplaceAllString(js, obfuscated)
	}

	return []byte(js)
}

// isAlphaNumeric checks if a character is alphanumeric or underscore
func isAlphaNumeric(char byte) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_' || char == '$'
}

// GetFileExtensions returns the file extensions this optimizer handles
func (j *JSOptimizer) GetFileExtensions() []string {
	return []string{".js", ".mjs"}
}
