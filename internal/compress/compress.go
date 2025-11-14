package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli"
)

// Compressor handles file compression
type Compressor struct {
	config Config
}

// Config holds compression configuration
type Config struct {
	Enabled        bool
	BrotliLevel    int // 0-11, default: 6
	GzipEnabled    bool
	GzipLevel      int      // 1-9, default: 6
	SkipExtensions []string // Extensions to skip (e.g., .min.js, .jpg)
}

// DefaultConfig returns sensible defaults for compression
func DefaultConfig() Config {
	return Config{
		Enabled:     true,
		BrotliLevel: 6,     // Balance between compression and speed
		GzipEnabled: false, // Only Brotli for now
		GzipLevel:   6,
		SkipExtensions: []string{
			".jpg", ".jpeg", ".png", ".gif", ".webp", // Already compressed images
			".mp4", ".mp3", ".wav", // Media files
			".zip", ".gz", ".br", ".tar", // Already compressed
			".woff", ".woff2", ".ttf", ".eot", // Fonts (already compressed)
		},
	}
}

// New creates a new compressor with the given config
func New(config Config) *Compressor {
	return &Compressor{config: config}
}

// ShouldCompress checks if a file should be compressed based on its extension
func (c *Compressor) ShouldCompress(filePath string) bool {
	if !c.config.Enabled {
		return false
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	// Check if extension is in skip list
	for _, skipExt := range c.config.SkipExtensions {
		if ext == skipExt {
			return false
		}
	}

	// Only compress text-based files
	compressibleExts := []string{
		".html", ".htm",
		".css",
		".js", ".mjs",
		".json",
		".xml",
		".txt",
		".md",
		".svg",
		".wasm",
	}

	for _, compExt := range compressibleExts {
		if ext == compExt {
			return true
		}
	}

	return false
}

// CompressFile compresses a file using Brotli
func (c *Compressor) CompressFile(inputPath, outputPath string) (*CompressionResult, error) {
	// Read input file
	data, err := os.ReadFile(inputPath) // #nosec G304 - Reading user-specified files for compression is intended behavior
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	originalSize := len(data)

	// Compress with Brotli
	var compressed bytes.Buffer
	writer := brotli.NewWriterLevel(&compressed, c.config.BrotliLevel)

	_, err = writer.Write(data)
	if err != nil {
		_ = writer.Close() // Ignore close error when write already failed
		return nil, fmt.Errorf("failed to compress: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close compressor: %w", err)
	}

	compressedData := compressed.Bytes()
	compressedSize := len(compressedData)

	// Only write compressed version if it's actually smaller
	if compressedSize < originalSize {
		// #nosec G306 - output file needs to be readable
		if err := os.WriteFile(outputPath, compressedData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write compressed file: %w", err)
		}

		return &CompressionResult{
			OriginalSize:   originalSize,
			CompressedSize: compressedSize,
			SavingsPercent: float64(originalSize-compressedSize) / float64(originalSize) * 100,
			Compressed:     true,
		}, nil
	}

	// If compressed version is larger, just copy the original
	// #nosec G306 - output file needs to be readable
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &CompressionResult{
		OriginalSize:   originalSize,
		CompressedSize: originalSize,
		SavingsPercent: 0,
		Compressed:     false,
	}, nil
}

// CompressDirectory compresses all eligible files in a directory
func (c *Compressor) CompressDirectory(dir string) (*DirectoryCompressionStats, error) {
	stats := &DirectoryCompressionStats{
		Files: make(map[string]*CompressionResult),
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Check if file should be compressed
		if !c.ShouldCompress(path) {
			stats.Skipped++
			return nil
		}

		// Create output path with .br extension
		outputPath := path + ".br"

		// Compress file
		result, err := c.CompressFile(path, outputPath)
		if err != nil {
			stats.Errors++
			return nil // Continue processing other files
		}

		stats.Files[relPath] = result
		stats.TotalOriginalSize += result.OriginalSize
		stats.TotalCompressedSize += result.CompressedSize

		if result.Compressed {
			stats.Compressed++
		} else {
			stats.NotWorthCompressing++
		}

		return nil
	})

	if err != nil {
		return stats, err
	}

	if stats.TotalOriginalSize > 0 {
		stats.OverallSavingsPercent = float64(stats.TotalOriginalSize-stats.TotalCompressedSize) / float64(stats.TotalOriginalSize) * 100
	}

	return stats, nil
}

// CompressInPlace compresses files and replaces originals with compressed versions
func (c *Compressor) CompressInPlace(dir string) (*DirectoryCompressionStats, error) {
	stats := &DirectoryCompressionStats{
		Files: make(map[string]*CompressionResult),
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Check if file should be compressed
		if !c.ShouldCompress(path) {
			stats.Skipped++
			return nil
		}

		// Compress to temporary file
		tmpPath := path + ".tmp.br"
		result, err := c.CompressFile(path, tmpPath)
		if err != nil {
			stats.Errors++
			return nil
		}

		// If compression was beneficial, replace original
		if result.Compressed {
			if err := os.Rename(tmpPath, path); err != nil {
				_ = os.Remove(tmpPath) // Best effort cleanup
				stats.Errors++
				return nil
			}
		} else {
			// Remove temp file if compression wasn't beneficial
			_ = os.Remove(tmpPath) // Best effort cleanup
		}

		stats.Files[relPath] = result
		stats.TotalOriginalSize += result.OriginalSize
		stats.TotalCompressedSize += result.CompressedSize

		if result.Compressed {
			stats.Compressed++
		} else {
			stats.NotWorthCompressing++
		}

		return nil
	})

	if err != nil {
		return stats, err
	}

	if stats.TotalOriginalSize > 0 {
		stats.OverallSavingsPercent = float64(stats.TotalOriginalSize-stats.TotalCompressedSize) / float64(stats.TotalOriginalSize) * 100
	}

	return stats, nil
}

// CompressBuffer compresses data in memory
func CompressBuffer(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	writer := brotli.NewWriterLevel(&buf, level)

	if _, err := writer.Write(data); err != nil {
		_ = writer.Close() // Ignore close error when write already failed
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DecompressBuffer decompresses Brotli-compressed data
func DecompressBuffer(data []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(reader)
}

// CompressGzip compresses data using gzip (fallback for compatibility)
func CompressGzip(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		_ = writer.Close() // Ignore close error when write already failed
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
