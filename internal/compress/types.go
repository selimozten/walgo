package compress

import (
	"fmt"

	"github.com/selimozten/walgo/internal/ui"
)

// CompressionResult holds the result of compressing a single file
type CompressionResult struct {
	OriginalSize   int
	CompressedSize int
	SavingsPercent float64
	Compressed     bool // True if file was actually compressed, false if original was kept
}

// DirectoryCompressionStats holds statistics for compressing a directory
type DirectoryCompressionStats struct {
	Files                 map[string]*CompressionResult
	Compressed            int
	NotWorthCompressing   int
	Skipped               int
	Errors                int
	TotalOriginalSize     int
	TotalCompressedSize   int
	OverallSavingsPercent float64
}

// PrintSummary prints a human-readable summary of compression stats
func (s *DirectoryCompressionStats) PrintSummary() {
	icons := ui.GetIcons()
	total := s.Compressed + s.NotWorthCompressing + s.Skipped

	fmt.Printf("\n%s Compression Summary:\n", icons.Package)
	fmt.Printf("  Total files processed: %d\n", total)
	fmt.Printf("  %s Compressed: %d\n", icons.Success, s.Compressed)
	fmt.Printf("  %s Skipped (already optimal): %d\n", icons.Arrow, s.NotWorthCompressing)
	fmt.Printf("  %s Skipped (binary/media): %d\n", icons.Cross, s.Skipped)

	if s.Errors > 0 {
		fmt.Printf("  %s Errors: %d\n", icons.Error, s.Errors)
	}

	fmt.Printf("\n  %s Original size: %.2f MB\n", icons.Database, float64(s.TotalOriginalSize)/(1024*1024))
	fmt.Printf("  %s Compressed size: %.2f MB\n", icons.Package, float64(s.TotalCompressedSize)/(1024*1024))
	fmt.Printf("  %s Total savings: %.1f%% (%.2f MB)\n", icons.Money,
		s.OverallSavingsPercent,
		float64(s.TotalOriginalSize-s.TotalCompressedSize)/(1024*1024))
}

// PrintVerboseSummary prints detailed file-by-file compression stats
func (s *DirectoryCompressionStats) PrintVerboseSummary() {
	s.PrintSummary()

	icons := ui.GetIcons()
	if len(s.Files) > 0 {
		fmt.Printf("\n  %s File details:\n", icons.File)
		count := 0
		for path, result := range s.Files {
			if count >= 20 { // Limit to first 20 files
				fmt.Printf("    ... and %d more files\n", len(s.Files)-20)
				break
			}

			if result.Compressed {
				fmt.Printf("    %s %s: %.1f%% savings (%.1f KB %s %.1f KB)\n",
					icons.Check, path,
					result.SavingsPercent,
					float64(result.OriginalSize)/1024, icons.Arrow,
					float64(result.CompressedSize)/1024)
			}
			count++
		}
	}
}
