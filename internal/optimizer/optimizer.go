package optimizer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/ui"
)

// Engine is the main optimization engine
type Engine struct {
	config        OptimizerConfig
	htmlOptimizer *HTMLOptimizer
	cssOptimizer  *CSSOptimizer
	jsOptimizer   *JSOptimizer
	htmlContent   map[string][]byte // Store HTML content for CSS unused rule detection
}

// NewEngine creates a new optimization engine
func NewEngine(config OptimizerConfig) *Engine {
	return &Engine{
		config:        config,
		htmlOptimizer: NewHTMLOptimizer(config.HTML),
		cssOptimizer:  NewCSSOptimizer(config.CSS),
		jsOptimizer:   NewJSOptimizer(config.JS),
		htmlContent:   make(map[string][]byte),
	}
}

// OptimizeDirectory optimizes all files in a directory recursively
func (e *Engine) OptimizeDirectory(sourceDir string) (*OptimizationStats, error) {
	if !e.config.Enabled {
		return &OptimizationStats{}, nil
	}

	stats := &OptimizationStats{}
	startTime := time.Now()

	// First pass: collect HTML content for CSS optimization
	if e.config.CSS.RemoveUnused {
		err := e.collectHTMLContent(sourceDir)
		if err != nil {
			return stats, fmt.Errorf("failed to collect HTML content: %w", err)
		}
	}

	// Second pass: optimize all files
	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Check if file should be skipped
		if e.shouldSkipFile(path) {
			stats.FilesSkipped++
			return nil
		}

		// Process the file
		if err := e.optimizeFile(path, stats); err != nil {
			if e.config.Verbose {
				fmt.Printf("Error optimizing %s: %v\n", path, err)
			}
			stats.FilesError++
		}

		stats.FilesProcessed++
		return nil
	})

	stats.Duration = time.Since(startTime)
	e.calculateStats(stats)

	return stats, err
}

// optimizeFile optimizes a single file based on its extension
func (e *Engine) optimizeFile(filePath string, stats *OptimizationStats) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Read the file
	originalContent, err := os.ReadFile(filePath) // #nosec G304 - filePath comes from controlled directory walk
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	originalSize := int64(len(originalContent))
	var optimizedContent []byte

	// Optimize based on file type
	switch ext {
	case ".html", ".htm":
		optimizedContent, err = e.htmlOptimizer.Optimize(originalContent)
		if err != nil {
			return fmt.Errorf("HTML optimization failed: %w", err)
		}
		e.updateHTMLStats(stats, originalSize, int64(len(optimizedContent)))

	case ".css":
		optimizedContent, err = e.cssOptimizer.Optimize(originalContent)
		if err != nil {
			return fmt.Errorf("CSS optimization failed: %w", err)
		}

		// Apply unused rule removal if enabled
		if e.config.CSS.RemoveUnused {
			optimizedContent = e.applyCSSUnusedRuleRemoval(optimizedContent)
		}

		e.updateCSSStats(stats, originalSize, int64(len(optimizedContent)))

	case ".js", ".mjs":
		optimizedContent, err = e.jsOptimizer.Optimize(originalContent)
		if err != nil {
			return fmt.Errorf("JavaScript optimization failed: %w", err)
		}
		e.updateJSStats(stats, originalSize, int64(len(optimizedContent)))

	default:
		// File type not supported for optimization
		return nil
	}

	// Only write if content changed
	if len(optimizedContent) != len(originalContent) || string(optimizedContent) != string(originalContent) {
		err = os.WriteFile(filePath, optimizedContent, 0644) // #nosec G306 - HTML/CSS/JS files need to be readable by web servers
		if err != nil {
			return fmt.Errorf("failed to write optimized file: %w", err)
		}
		stats.FilesOptimized++
	}

	return nil
}

// collectHTMLContent collects all HTML content for CSS unused rule detection
func (e *Engine) collectHTMLContent(sourceDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".html" || ext == ".htm" {
			content, err := os.ReadFile(path) // #nosec G304 - path comes from controlled directory walk
			if err != nil {
				return err
			}
			e.htmlContent[path] = content
		}

		return nil
	})
}

// applyCSSUnusedRuleRemoval applies unused CSS rule removal using collected HTML content
func (e *Engine) applyCSSUnusedRuleRemoval(cssContent []byte) []byte {
	// Combine all HTML content
	var allHTML []byte
	for _, htmlContent := range e.htmlContent {
		allHTML = append(allHTML, htmlContent...)
		allHTML = append(allHTML, []byte(" ")...) // Add separator
	}

	if len(allHTML) > 0 {
		return e.cssOptimizer.RemoveUnusedRules(cssContent, allHTML)
	}

	return cssContent
}

// shouldSkipFile determines if a file should be skipped based on patterns
func (e *Engine) shouldSkipFile(filePath string) bool {
	// Check skip patterns
	for _, pattern := range e.config.SkipPatterns {
		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err == nil && matched {
			return true
		}

		// Also check against full path for directory patterns
		if strings.Contains(pattern, "/") {
			matched, err := filepath.Match(pattern, filePath)
			if err == nil && matched {
				return true
			}
		}
	}

	// Skip already minified files
	fileName := filepath.Base(filePath)
	return strings.Contains(fileName, ".min.")
}

// updateHTMLStats updates HTML-specific statistics
func (e *Engine) updateHTMLStats(stats *OptimizationStats, originalSize, optimizedSize int64) {
	stats.HTMLFiles.FilesProcessed++
	stats.HTMLFiles.OriginalSize += originalSize
	stats.HTMLFiles.OptimizedSize += optimizedSize
	stats.HTMLFiles.SavingsBytes += originalSize - optimizedSize

	if originalSize > 0 {
		stats.HTMLFiles.SavingsPercent = float64(stats.HTMLFiles.SavingsBytes) / float64(stats.HTMLFiles.OriginalSize) * 100
	}
}

// updateCSSStats updates CSS-specific statistics
func (e *Engine) updateCSSStats(stats *OptimizationStats, originalSize, optimizedSize int64) {
	stats.CSSFiles.FilesProcessed++
	stats.CSSFiles.OriginalSize += originalSize
	stats.CSSFiles.OptimizedSize += optimizedSize
	stats.CSSFiles.SavingsBytes += originalSize - optimizedSize

	if originalSize > 0 {
		stats.CSSFiles.SavingsPercent = float64(stats.CSSFiles.SavingsBytes) / float64(stats.CSSFiles.OriginalSize) * 100
	}
}

// updateJSStats updates JavaScript-specific statistics
func (e *Engine) updateJSStats(stats *OptimizationStats, originalSize, optimizedSize int64) {
	stats.JSFiles.FilesProcessed++
	stats.JSFiles.OriginalSize += originalSize
	stats.JSFiles.OptimizedSize += optimizedSize
	stats.JSFiles.SavingsBytes += originalSize - optimizedSize

	if originalSize > 0 {
		stats.JSFiles.SavingsPercent = float64(stats.JSFiles.SavingsBytes) / float64(stats.JSFiles.OriginalSize) * 100
	}
}

// calculateStats calculates overall statistics
func (e *Engine) calculateStats(stats *OptimizationStats) {
	stats.OriginalSize = stats.HTMLFiles.OriginalSize + stats.CSSFiles.OriginalSize + stats.JSFiles.OriginalSize
	stats.OptimizedSize = stats.HTMLFiles.OptimizedSize + stats.CSSFiles.OptimizedSize + stats.JSFiles.OptimizedSize
	stats.SavingsBytes = stats.OriginalSize - stats.OptimizedSize

	if stats.OriginalSize > 0 {
		stats.SavingsPercent = float64(stats.SavingsBytes) / float64(stats.OriginalSize) * 100
	}
}

// PrintStats prints optimization statistics
func (e *Engine) PrintStats(stats *OptimizationStats) {
	icons := ui.GetIcons()
	fmt.Printf("\n%s Optimization Results\n", icons.Sparkles)
	fmt.Println(ui.Separator())
	fmt.Printf("Files processed: %d\n", stats.FilesProcessed)
	fmt.Printf("Files optimized: %d\n", stats.FilesOptimized)
	fmt.Printf("Files skipped: %d\n", stats.FilesSkipped)
	fmt.Printf("Files with errors: %d\n", stats.FilesError)
	fmt.Printf("Duration: %v\n", stats.Duration)

	if stats.OriginalSize > 0 {
		fmt.Printf("\n%s Size Reduction\n", icons.Chart)
		fmt.Printf("Original size: %s\n", formatBytes(stats.OriginalSize))
		fmt.Printf("Optimized size: %s\n", formatBytes(stats.OptimizedSize))
		fmt.Printf("Bytes saved: %s (%.1f%%)\n", formatBytes(stats.SavingsBytes), stats.SavingsPercent)
	}

	if stats.HTMLFiles.FilesProcessed > 0 {
		fmt.Printf("\n%s HTML Files\n", icons.File)
		fmt.Printf("Files: %d, Saved: %s (%.1f%%)\n",
			stats.HTMLFiles.FilesProcessed,
			formatBytes(stats.HTMLFiles.SavingsBytes),
			stats.HTMLFiles.SavingsPercent)
	}

	if stats.CSSFiles.FilesProcessed > 0 {
		fmt.Printf("\n%s CSS Files\n", icons.File)
		fmt.Printf("Files: %d, Saved: %s (%.1f%%)\n",
			stats.CSSFiles.FilesProcessed,
			formatBytes(stats.CSSFiles.SavingsBytes),
			stats.CSSFiles.SavingsPercent)
		if stats.CSSFiles.RulesRemoved > 0 {
			fmt.Printf("Unused rules removed: %d\n", stats.CSSFiles.RulesRemoved)
		}
	}

	if stats.JSFiles.FilesProcessed > 0 {
		fmt.Printf("\n%s JavaScript Files\n", icons.File)
		fmt.Printf("Files: %d, Saved: %s (%.1f%%)\n",
			stats.JSFiles.FilesProcessed,
			formatBytes(stats.JSFiles.SavingsBytes),
			stats.JSFiles.SavingsPercent)
	}
}

// formatBytes formats byte count into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
