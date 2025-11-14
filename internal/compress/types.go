package compress

import "fmt"

// CompressionResult holds the result of compressing a single file
type CompressionResult struct {
	OriginalSize   int
	CompressedSize int
	SavingsPercent float64
	Compressed     bool // True if file was actually compressed, false if original was kept
}

// DirectoryCompressionStats holds statistics for compressing a directory
type DirectoryCompressionStats struct {
	Files                  map[string]*CompressionResult
	Compressed             int
	NotWorthCompressing    int
	Skipped                int
	Errors                 int
	TotalOriginalSize      int
	TotalCompressedSize    int
	OverallSavingsPercent  float64
}

// PrintSummary prints a human-readable summary of compression stats
func (s *DirectoryCompressionStats) PrintSummary() {
	total := s.Compressed + s.NotWorthCompressing + s.Skipped

	fmt.Println("\n📦 Compression Summary:")
	fmt.Printf("  Total files processed: %d\n", total)
	fmt.Printf("  ✅ Compressed: %d\n", s.Compressed)
	fmt.Printf("  ⏭️  Skipped (already optimal): %d\n", s.NotWorthCompressing)
	fmt.Printf("  🚫 Skipped (binary/media): %d\n", s.Skipped)

	if s.Errors > 0 {
		fmt.Printf("  ❌ Errors: %d\n", s.Errors)
	}

	fmt.Printf("\n  💾 Original size: %.2f MB\n", float64(s.TotalOriginalSize)/(1024*1024))
	fmt.Printf("  📦 Compressed size: %.2f MB\n", float64(s.TotalCompressedSize)/(1024*1024))
	fmt.Printf("  💰 Total savings: %.1f%% (%.2f MB)\n",
		s.OverallSavingsPercent,
		float64(s.TotalOriginalSize-s.TotalCompressedSize)/(1024*1024))
}

// PrintVerboseSummary prints detailed file-by-file compression stats
func (s *DirectoryCompressionStats) PrintVerboseSummary() {
	s.PrintSummary()

	if len(s.Files) > 0 {
		fmt.Println("\n  📄 File details:")
		count := 0
		for path, result := range s.Files {
			if count >= 20 { // Limit to first 20 files
				fmt.Printf("    ... and %d more files\n", len(s.Files)-20)
				break
			}

			if result.Compressed {
				fmt.Printf("    ✓ %s: %.1f%% savings (%.1f KB → %.1f KB)\n",
					path,
					result.SavingsPercent,
					float64(result.OriginalSize)/1024,
					float64(result.CompressedSize)/1024)
			}
			count++
		}
	}
}
