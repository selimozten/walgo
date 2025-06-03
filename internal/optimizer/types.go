package optimizer

import "time"

// OptimizerConfig holds configuration for the optimization engine
type OptimizerConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// HTML optimization settings
	HTML HTMLConfig `mapstructure:"html" yaml:"html"`

	// CSS optimization settings
	CSS CSSConfig `mapstructure:"css" yaml:"css"`

	// JavaScript optimization settings
	JS JSConfig `mapstructure:"js" yaml:"js"`

	// General settings
	OutputDir    string   `mapstructure:"outputDir" yaml:"outputDir,omitempty"`       // Override output directory
	SkipPatterns []string `mapstructure:"skipPatterns" yaml:"skipPatterns,omitempty"` // Files/patterns to skip
	Verbose      bool     `mapstructure:"verbose" yaml:"verbose"`
}

// HTMLConfig holds HTML-specific optimization settings
type HTMLConfig struct {
	Enabled           bool `mapstructure:"enabled" yaml:"enabled"`
	MinifyHTML        bool `mapstructure:"minifyHTML" yaml:"minifyHTML"`
	RemoveComments    bool `mapstructure:"removeComments" yaml:"removeComments"`
	RemoveWhitespace  bool `mapstructure:"removeWhitespace" yaml:"removeWhitespace"`
	CompressInlineCSS bool `mapstructure:"compressInlineCSS" yaml:"compressInlineCSS"`
	CompressInlineJS  bool `mapstructure:"compressInlineJS" yaml:"compressInlineJS"`
}

// CSSConfig holds CSS-specific optimization settings
type CSSConfig struct {
	Enabled        bool `mapstructure:"enabled" yaml:"enabled"`
	MinifyCSS      bool `mapstructure:"minifyCSS" yaml:"minifyCSS"`
	RemoveComments bool `mapstructure:"removeComments" yaml:"removeComments"`
	RemoveUnused   bool `mapstructure:"removeUnused" yaml:"removeUnused"`     // Remove unused CSS rules
	Autoprefixer   bool `mapstructure:"autoprefixer" yaml:"autoprefixer"`     // Add vendor prefixes
	CompressColors bool `mapstructure:"compressColors" yaml:"compressColors"` // Compress color values
}

// JSConfig holds JavaScript-specific optimization settings
type JSConfig struct {
	Enabled        bool `mapstructure:"enabled" yaml:"enabled"`
	MinifyJS       bool `mapstructure:"minifyJS" yaml:"minifyJS"`
	RemoveComments bool `mapstructure:"removeComments" yaml:"removeComments"`
	Obfuscate      bool `mapstructure:"obfuscate" yaml:"obfuscate"`   // Basic variable name obfuscation
	SourceMaps     bool `mapstructure:"sourceMaps" yaml:"sourceMaps"` // Generate source maps
}

// OptimizationStats holds statistics about the optimization process
type OptimizationStats struct {
	FilesProcessed int
	FilesOptimized int
	FilesSkipped   int
	FilesError     int

	// Size statistics
	OriginalSize   int64
	OptimizedSize  int64
	SavingsBytes   int64
	SavingsPercent float64

	// File type breakdown
	HTMLFiles HTMLStats
	CSSFiles  CSSStats
	JSFiles   JSStats

	// Timing
	Duration time.Duration
}

// HTMLStats holds HTML-specific optimization statistics
type HTMLStats struct {
	FilesProcessed int
	OriginalSize   int64
	OptimizedSize  int64
	SavingsBytes   int64
	SavingsPercent float64
}

// CSSStats holds CSS-specific optimization statistics
type CSSStats struct {
	FilesProcessed int
	OriginalSize   int64
	OptimizedSize  int64
	SavingsBytes   int64
	SavingsPercent float64
	RulesRemoved   int // Number of unused CSS rules removed
}

// JSStats holds JavaScript-specific optimization statistics
type JSStats struct {
	FilesProcessed int
	OriginalSize   int64
	OptimizedSize  int64
	SavingsBytes   int64
	SavingsPercent float64
}

// NewDefaultOptimizerConfig creates an OptimizerConfig with sensible defaults
func NewDefaultOptimizerConfig() OptimizerConfig {
	return OptimizerConfig{
		Enabled: true,
		HTML: HTMLConfig{
			Enabled:           true,
			MinifyHTML:        true,
			RemoveComments:    true,
			RemoveWhitespace:  true,
			CompressInlineCSS: true,
			CompressInlineJS:  true,
		},
		CSS: CSSConfig{
			Enabled:        true,
			MinifyCSS:      true,
			RemoveComments: true,
			RemoveUnused:   false, // Can be aggressive, disabled by default
			Autoprefixer:   false, // Requires additional dependencies
			CompressColors: true,
		},
		JS: JSConfig{
			Enabled:        true,
			MinifyJS:       true,
			RemoveComments: true,
			Obfuscate:      false, // Can break code, disabled by default
			SourceMaps:     false, // Usually not needed for static sites
		},
		SkipPatterns: []string{
			"*.min.js",
			"*.min.css",
			"*.min.html",
			"*/.git/*",
			"*/node_modules/*",
		},
		Verbose: false,
	}
}
