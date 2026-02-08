package optimizer

import (
	"fmt"
	"regexp"
	"strings"
)

// Pre-compiled regexes for CSS operations.
var (
	cssCommentRegex     = regexp.MustCompile(`/\*[^*]*\*+(?:[^/*][^*]*\*+)*/`)
	cssDoubleQuoteRegex = regexp.MustCompile(`"(?:\\.|[^"\\])*"`)
	cssSingleQuoteRegex = regexp.MustCompile(`'(?:\\.|[^'\\])*'`)
	cssHexColorRegex    = regexp.MustCompile(`#([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])`)

	// Minification regexes — only safe whitespace removal around unambiguous tokens.
	// Deliberately omits >, +, ~ (CSS combinators that also appear inside calc(),
	// media range queries, unicode-range, etc.) and URL quote removal (breaks
	// URLs containing spaces or special characters).
	cssBraceOpenRegex   = regexp.MustCompile(`\s*{\s*`)
	cssBraceCloseRegex  = regexp.MustCompile(`\s*}\s*`)
	cssSemicolonRegex   = regexp.MustCompile(`\s*;\s*`)
	cssColonRegex       = regexp.MustCompile(`\s*:\s*`)
	cssCommaRegex       = regexp.MustCompile(`\s*,\s*`)
	cssMultiSpaceRegex  = regexp.MustCompile(`\s+`)
	cssLastSemiRegex    = regexp.MustCompile(`;\s*}`)
	cssZeroUnitRegex    = regexp.MustCompile(`\b0+(px|em|rem|ex|pt|pc|in|cm|mm|%|vh|vw|vmin|vmax)\b`)
	cssLeadingZeroRegex = regexp.MustCompile(`\b0(\.\d+)`)
)

// CSSOptimizer handles CSS optimization
type CSSOptimizer struct {
	config CSSConfig
}

// NewCSSOptimizer creates a new CSS optimizer
func NewCSSOptimizer(config CSSConfig) *CSSOptimizer {
	return &CSSOptimizer{
		config: config,
	}
}

// Optimize optimizes CSS content
func (c *CSSOptimizer) Optimize(content []byte) ([]byte, error) {
	if !c.config.Enabled {
		return content, nil
	}

	result := content

	// Remove CSS comments
	if c.config.RemoveComments {
		result = c.removeComments(result)
	}

	// Compress colors
	if c.config.CompressColors {
		result = c.compressColors(result)
	}

	// Minify CSS
	if c.config.MinifyCSS {
		result = c.minifyCSS(result)
	}

	return result, nil
}

// removeComments removes CSS comments
func (c *CSSOptimizer) removeComments(content []byte) []byte {
	return cssCommentRegex.ReplaceAll(content, []byte(""))
}

// compressColors compresses CSS hex color values while preserving string contents.
// Only performs safe 6-digit → 3-digit hex shortening (#aabbcc → #abc).
// Named color replacement (white → #fff) is intentionally omitted because it
// corrupts CSS custom property names (e.g. --black-text → --#000-text).
func (c *CSSOptimizer) compressColors(content []byte) []byte {
	css := string(content)

	// Preserve strings (content: "...", etc.) before processing
	stringPlaceholders := make(map[string]string)
	placeholderCounter := 0

	// Extract and preserve double-quoted strings
	css = cssDoubleQuoteRegex.ReplaceAllStringFunc(css, func(match string) string {
		placeholder := fmt.Sprintf("__WALGO_CSS_STR_%d__", placeholderCounter)
		stringPlaceholders[placeholder] = match
		placeholderCounter++
		return placeholder
	})

	// Extract and preserve single-quoted strings
	css = cssSingleQuoteRegex.ReplaceAllStringFunc(css, func(match string) string {
		placeholder := fmt.Sprintf("__WALGO_CSS_STR_%d__", placeholderCounter)
		stringPlaceholders[placeholder] = match
		placeholderCounter++
		return placeholder
	})

	// Compress 6-digit hex colors to 3-digit when possible
	// e.g., #aabbcc -> #abc
	css = cssHexColorRegex.ReplaceAllStringFunc(css, func(match string) string {
		if len(match) == 7 && match[1] == match[2] && match[3] == match[4] && match[5] == match[6] {
			return "#" + match[1:2] + match[3:4] + match[5:6]
		}
		return match
	})

	// Restore preserved strings
	for placeholder, original := range stringPlaceholders {
		css = strings.Replace(css, placeholder, original, 1)
	}

	return []byte(css)
}

// minifyCSS minifies CSS by removing unnecessary whitespace and formatting.
// Only strips whitespace around tokens that are unambiguous ({, }, ;, :, ,).
// Deliberately does NOT strip whitespace around >, +, ~ because these CSS
// combinators also appear inside calc(), media range queries, and unicode-range
// where removing whitespace changes meaning or breaks parsing.
func (c *CSSOptimizer) minifyCSS(content []byte) []byte {
	css := string(content)

	// Remove unnecessary whitespace around unambiguous special characters
	css = cssBraceOpenRegex.ReplaceAllString(css, "{")
	css = cssBraceCloseRegex.ReplaceAllString(css, "}")
	css = cssSemicolonRegex.ReplaceAllString(css, ";")
	css = cssColonRegex.ReplaceAllString(css, ":")
	css = cssCommaRegex.ReplaceAllString(css, ",")

	// Remove leading and trailing whitespace
	css = strings.TrimSpace(css)

	// Collapse multiple spaces/newlines into a single space
	css = cssMultiSpaceRegex.ReplaceAllString(css, " ")

	// Remove last semicolon in declarations
	css = cssLastSemiRegex.ReplaceAllString(css, "}")

	// Remove zero units (0px -> 0, 0em -> 0, etc.)
	// Intentionally excludes time units (s, ms) where unitless 0 is invalid per CSS spec.
	css = cssZeroUnitRegex.ReplaceAllString(css, "0")

	// Remove leading zero from decimal numbers (0.5 -> .5)
	css = cssLeadingZeroRegex.ReplaceAllString(css, "$1")

	return []byte(css)
}
