package optimizer

import (
	"regexp"
	"strings"
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
	// Remove /* comment */ style comments
	commentRegex := regexp.MustCompile(`/\*[^*]*\*+(?:[^/*][^*]*\*+)*/`)
	return commentRegex.ReplaceAll(content, []byte(""))
}

// compressColors compresses CSS color values
func (c *CSSOptimizer) compressColors(content []byte) []byte {
	css := string(content)

	// Compress 6-digit hex colors to 3-digit when possible
	// e.g., #aabbcc -> #abc
	hexColorRegex := regexp.MustCompile(`#([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])`)
	css = hexColorRegex.ReplaceAllStringFunc(css, func(match string) string {
		if len(match) == 7 && match[1] == match[2] && match[3] == match[4] && match[5] == match[6] {
			return "#" + string(match[1]) + string(match[3]) + string(match[5])
		}
		return match
	})

	// Convert named colors to shorter equivalents
	colorMap := map[string]string{
		"black":   "#000",
		"white":   "#fff",
		"red":     "#f00",
		"green":   "#008000",
		"blue":    "#00f",
		"yellow":  "#ff0",
		"cyan":    "#0ff",
		"magenta": "#f0f",
		"silver":  "#c0c0c0",
		"gray":    "#808080",
		"maroon":  "#800000",
		"olive":   "#808000",
		"lime":    "#0f0",
		"aqua":    "#0ff",
		"teal":    "#008080",
		"navy":    "#000080",
		"fuchsia": "#f0f",
		"purple":  "#800080",
	}

	for colorName, hexValue := range colorMap {
		// Only replace if the hex value is shorter
		if len(hexValue) < len(colorName) {
			wordBoundaryRegex := regexp.MustCompile(`\b` + colorName + `\b`)
			css = wordBoundaryRegex.ReplaceAllString(css, hexValue)
		}
	}

	// Convert rgb(255,255,255) to shorter hex
	rgbRegex := regexp.MustCompile(`rgb\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)`)
	css = rgbRegex.ReplaceAllStringFunc(css, func(match string) string {
		// For now, keep original format to avoid complex conversion
		// In a production version, you'd implement RGB to hex conversion
		return match
	})

	return []byte(css)
}

// minifyCSS minifies CSS by removing unnecessary whitespace and formatting
func (c *CSSOptimizer) minifyCSS(content []byte) []byte {
	css := string(content)

	// Remove unnecessary whitespace around special characters
	css = regexp.MustCompile(`\s*{\s*`).ReplaceAllString(css, "{")
	css = regexp.MustCompile(`\s*}\s*`).ReplaceAllString(css, "}")
	css = regexp.MustCompile(`\s*;\s*`).ReplaceAllString(css, ";")
	css = regexp.MustCompile(`\s*:\s*`).ReplaceAllString(css, ":")
	css = regexp.MustCompile(`\s*,\s*`).ReplaceAllString(css, ",")
	css = regexp.MustCompile(`\s*>\s*`).ReplaceAllString(css, ">")
	css = regexp.MustCompile(`\s*\+\s*`).ReplaceAllString(css, "+")
	css = regexp.MustCompile(`\s*~\s*`).ReplaceAllString(css, "~")

	// Remove leading and trailing whitespace
	css = strings.TrimSpace(css)

	// Remove multiple spaces and newlines
	css = regexp.MustCompile(`\s+`).ReplaceAllString(css, " ")

	// Remove last semicolon in declarations
	css = regexp.MustCompile(`;\s*}`).ReplaceAllString(css, "}")

	// Remove unnecessary quotes from URLs (if safe)
	css = regexp.MustCompile(`url\(["']([^"']+)["']\)`).ReplaceAllString(css, "url($1)")

	// Remove zero units (0px -> 0, 0em -> 0, etc.) except for time values
	css = regexp.MustCompile(`\b0+(px|em|rem|ex|pt|pc|in|cm|mm|%|vh|vw|vmin|vmax)\b`).ReplaceAllString(css, "0")

	// Compress multiple zeros in decimal numbers
	css = regexp.MustCompile(`0\.0*(\d+)`).ReplaceAllString(css, ".$1")

	return []byte(css)
}

// findUnusedRules finds CSS rules that are not used in the provided HTML content
func (c *CSSOptimizer) findUnusedRules(cssContent []byte, htmlContent []byte) []string {
	if !c.config.RemoveUnused {
		return nil
	}

	css := string(cssContent)
	html := string(htmlContent)

	var unusedRules []string

	// Extract CSS selectors (simplified approach)
	selectorRegex := regexp.MustCompile(`([^{}]+)\s*{[^}]*}`)
	matches := selectorRegex.FindAllStringSubmatch(css, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		selector := strings.TrimSpace(match[1])

		// Skip media queries, keyframes, and other special rules
		if strings.HasPrefix(selector, "@") {
			continue
		}

		// Check if selector is used in HTML
		if !c.isSelectorUsed(selector, html) {
			unusedRules = append(unusedRules, match[0])
		}
	}

	return unusedRules
}

// isSelectorUsed checks if a CSS selector is used in the HTML content
func (c *CSSOptimizer) isSelectorUsed(selector, html string) bool {
	// Simplified check - in a production version, you'd use a proper CSS parser
	// and HTML parser to accurately determine usage

	// Split complex selectors
	parts := regexp.MustCompile(`[,\s]+`).Split(selector, -1)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for ID selectors
		if strings.HasPrefix(part, "#") {
			id := strings.TrimPrefix(part, "#")
			if strings.Contains(html, `id="`+id+`"`) || strings.Contains(html, `id='`+id+`'`) {
				return true
			}
		}

		// Check for class selectors
		if strings.HasPrefix(part, ".") {
			class := strings.TrimPrefix(part, ".")
			if strings.Contains(html, `class="`+class+`"`) ||
				strings.Contains(html, `class='`+class+`'`) ||
				strings.Contains(html, ` `+class+` `) {
				return true
			}
		}

		// Check for element selectors
		if regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(part) {
			tagRegex := regexp.MustCompile(`<` + part + `[\s>]`)
			if tagRegex.MatchString(html) {
				return true
			}
		}
	}

	return false
}

// RemoveUnusedRules removes unused CSS rules based on provided HTML content
func (c *CSSOptimizer) RemoveUnusedRules(cssContent []byte, htmlContent []byte) []byte {
	if !c.config.RemoveUnused {
		return cssContent
	}

	unusedRules := c.findUnusedRules(cssContent, htmlContent)
	css := string(cssContent)

	for _, rule := range unusedRules {
		css = strings.Replace(css, rule, "", 1)
	}

	return []byte(css)
}

// GetFileExtensions returns the file extensions this optimizer handles
func (c *CSSOptimizer) GetFileExtensions() []string {
	return []string{".css"}
}
