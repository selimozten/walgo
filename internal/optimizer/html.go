package optimizer

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// HTMLOptimizer handles HTML optimization
type HTMLOptimizer struct {
	config HTMLConfig
}

// NewHTMLOptimizer creates a new HTML optimizer
func NewHTMLOptimizer(config HTMLConfig) *HTMLOptimizer {
	return &HTMLOptimizer{
		config: config,
	}
}

// Optimize optimizes HTML content
func (h *HTMLOptimizer) Optimize(content []byte) ([]byte, error) {
	if !h.config.Enabled {
		return content, nil
	}

	result := content

	// Remove HTML comments
	if h.config.RemoveComments {
		result = h.removeComments(result)
	}

	// Compress inline CSS
	if h.config.CompressInlineCSS {
		result = h.compressInlineCSS(result)
	}

	// Compress inline JavaScript
	if h.config.CompressInlineJS {
		result = h.compressInlineJS(result)
	}

	// Minify HTML
	if h.config.MinifyHTML {
		result = h.minifyHTML(result)
	}

	return result, nil
}

// removeComments removes HTML comments while preserving conditional comments
func (h *HTMLOptimizer) removeComments(content []byte) []byte {
	// Preserve conditional comments (<!--[if ...]>...< ![endif]-->)
	conditionalCommentRegex := regexp.MustCompile(`<!--\[if[^>]*>.*?<!\[endif\]-->`)
	conditionalComments := conditionalCommentRegex.FindAll(content, -1)

	// Replace conditional comments with placeholders
	placeholders := make(map[string][]byte)
	for i, comment := range conditionalComments {
		placeholder := fmt.Sprintf("__CONDITIONAL_COMMENT_%d__", i)
		placeholders[placeholder] = comment
		content = bytes.Replace(content, comment, []byte(placeholder), 1)
	}

	// Remove regular HTML comments
	commentRegex := regexp.MustCompile(`<!--[^>]*-->`)
	content = commentRegex.ReplaceAll(content, []byte(""))

	// Restore conditional comments
	for placeholder, comment := range placeholders {
		content = bytes.Replace(content, []byte(placeholder), comment, 1)
	}

	return content
}

// compressInlineCSS compresses CSS within <style> tags
func (h *HTMLOptimizer) compressInlineCSS(content []byte) []byte {
	styleRegex := regexp.MustCompile(`(?s)<style[^>]*>(.*?)</style>`)

	return styleRegex.ReplaceAllFunc(content, func(match []byte) []byte {
		styleMatch := styleRegex.FindSubmatch(match)
		if len(styleMatch) < 2 {
			return match
		}

		cssContent := styleMatch[1]
		cssOptimizer := NewCSSOptimizer(CSSConfig{
			Enabled:        true,
			MinifyCSS:      true,
			RemoveComments: true,
			CompressColors: true,
		})

		optimizedCSS, err := cssOptimizer.Optimize(cssContent)
		if err != nil {
			return match // Return original on error
		}

		// Reconstruct the style tag
		return bytes.Replace(match, cssContent, optimizedCSS, 1)
	})
}

// compressInlineJS compresses JavaScript within <script> tags
func (h *HTMLOptimizer) compressInlineJS(content []byte) []byte {
	scriptRegex := regexp.MustCompile(`(?s)<script(?:[^>]*)>(.*?)</script>`)

	return scriptRegex.ReplaceAllFunc(content, func(match []byte) []byte {
		scriptMatch := scriptRegex.FindSubmatch(match)
		if len(scriptMatch) < 2 {
			return match
		}

		// Check if script has src attribute (external script)
		if bytes.Contains(match, []byte("src=")) {
			return match // Don't process external scripts
		}

		jsContent := scriptMatch[1]
		jsOptimizer := NewJSOptimizer(JSConfig{
			Enabled:        true,
			MinifyJS:       true,
			RemoveComments: true,
		})

		optimizedJS, err := jsOptimizer.Optimize(jsContent)
		if err != nil {
			return match // Return original on error
		}

		// Reconstruct the script tag
		return bytes.Replace(match, jsContent, optimizedJS, 1)
	})
}

// minifyHTML minifies HTML by removing unnecessary whitespace
func (h *HTMLOptimizer) minifyHTML(content []byte) []byte {
	if !h.config.RemoveWhitespace {
		return content
	}

	// Convert to string for easier manipulation
	html := string(content)

	// Preserve content of <pre>, <code>, <textarea>, and <script> tags
	preserveTags := []string{"pre", "code", "textarea", "script"}
	preserved := make(map[string]string)
	placeholderCounter := 0

	for _, tag := range preserveTags {
		tagRegex := regexp.MustCompile(fmt.Sprintf(`(?s)<%s[^>]*>.*?</%s>`, tag, tag))
		html = tagRegex.ReplaceAllStringFunc(html, func(match string) string {
			placeholder := fmt.Sprintf("__PRESERVE_%d__", placeholderCounter)
			preserved[placeholder] = match
			placeholderCounter++
			return placeholder
		})
	}

	// Remove whitespace between tags
	html = regexp.MustCompile(`>\s+<`).ReplaceAllString(html, "><")

	// Remove leading/trailing whitespace from lines
	lines := strings.Split(html, "\n")
	var processedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			processedLines = append(processedLines, trimmed)
		}
	}
	html = strings.Join(processedLines, "")

	// Restore preserved content
	for placeholder, original := range preserved {
		html = strings.Replace(html, placeholder, original, 1)
	}

	return []byte(html)
}

// GetFileExtensions returns the file extensions this optimizer handles
func (h *HTMLOptimizer) GetFileExtensions() []string {
	return []string{".html", ".htm"}
}
