package walrus

import (
	"math"
	"strings"
	"testing"
)

func TestEncodingMultiplierForSize(t *testing.T) {
	tests := []struct {
		name         string
		originalSize int64
		wantMin      float64
		wantMax      float64
	}{
		{
			name:         "very small file (100 bytes)",
			originalSize: 100,
			wantMin:      10.0,
			wantMax:      10.0,
		},
		{
			name:         "sub-MiB file (500 KB)",
			originalSize: 500 * 1024,
			wantMin:      10.0,
			wantMax:      10.0,
		},
		{
			name:         "exactly 1 MiB boundary",
			originalSize: 1024 * 1024,
			wantMin:      8.5,
			wantMax:      8.5,
		},
		{
			name:         "small file (5 MiB)",
			originalSize: 5 * 1024 * 1024,
			wantMin:      8.5,
			wantMax:      8.5,
		},
		{
			name:         "16 MiB file",
			originalSize: 16 * 1024 * 1024,
			wantMin:      6.0,
			wantMax:      6.0,
		},
		{
			name:         "medium file (50 MiB)",
			originalSize: 50 * 1024 * 1024,
			wantMin:      6.0,
			wantMax:      6.0,
		},
		{
			name:         "100 MiB file",
			originalSize: 100 * 1024 * 1024,
			wantMin:      5.0,
			wantMax:      5.0,
		},
		{
			name:         "large file (300 MiB)",
			originalSize: 300 * 1024 * 1024,
			wantMin:      5.0,
			wantMax:      5.0,
		},
		{
			name:         "500 MiB file",
			originalSize: 500 * 1024 * 1024,
			wantMin:      4.5,
			wantMax:      4.5,
		},
		{
			name:         "very large file (1 GiB)",
			originalSize: 1024 * 1024 * 1024,
			wantMin:      4.5,
			wantMax:      4.5,
		},
		{
			name:         "zero size",
			originalSize: 0,
			wantMin:      10.0,
			wantMax:      10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodingMultiplierForSize(tt.originalSize)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("encodingMultiplierForSize(%d) = %v, want between %v and %v",
					tt.originalSize, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateEncodedSize(t *testing.T) {
	tests := []struct {
		name         string
		originalSize int64
		wantMin      int64
		wantMax      int64
	}{
		{
			name:         "zero bytes",
			originalSize: 0,
			wantMin:      0,
			wantMax:      0,
		},
		{
			name:         "1 byte",
			originalSize: 1,
			wantMin:      10, // 1 * 10.0
			wantMax:      10,
		},
		{
			name:         "1 KiB",
			originalSize: 1024,
			wantMin:      1024 * 10, // 10.0x for sub-MiB
			wantMax:      1024 * 10,
		},
		{
			name:         "1 MiB",
			originalSize: 1024 * 1024,
			wantMin:      int64(float64(1024*1024) * 8.5),
			wantMax:      int64(float64(1024*1024) * 8.5),
		},
		{
			name:         "100 MiB",
			originalSize: 100 * 1024 * 1024,
			wantMin:      int64(float64(100*1024*1024) * 5.0),
			wantMax:      int64(float64(100*1024*1024) * 5.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateEncodedSize(tt.originalSize)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateEncodedSize(%d) = %d, want between %d and %d",
					tt.originalSize, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateEncodedSizeWithMultiplier(t *testing.T) {
	tests := []struct {
		name         string
		originalSize int64
		multiplier   float64
		want         int64
	}{
		{
			name:         "zero size",
			originalSize: 0,
			multiplier:   5.0,
			want:         0,
		},
		{
			name:         "1 MiB with 5x multiplier",
			originalSize: 1024 * 1024,
			multiplier:   5.0,
			want:         5 * 1024 * 1024,
		},
		{
			name:         "1 MiB with 8.375x multiplier",
			originalSize: 1024 * 1024,
			multiplier:   8.375,
			want:         int64(float64(1024*1024) * 8.375),
		},
		{
			name:         "1 byte with 1x multiplier",
			originalSize: 1,
			multiplier:   1.0,
			want:         1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateEncodedSizeWithMultiplier(tt.originalSize, tt.multiplier)
			if got != tt.want {
				t.Errorf("calculateEncodedSizeWithMultiplier(%d, %v) = %d, want %d",
					tt.originalSize, tt.multiplier, got, tt.want)
			}
		})
	}
}

func TestGetRPCEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		network string
		want    string
	}{
		{"mainnet", "mainnet", SuiMainnetRPC},
		{"testnet", "testnet", SuiTestnetRPC},
		{"mainnet uppercase", "Mainnet", SuiMainnetRPC},
		{"testnet uppercase", "Testnet", SuiTestnetRPC},
		{"MAINNET", "MAINNET", SuiMainnetRPC},
		{"empty defaults to testnet", "", SuiTestnetRPC},
		{"unknown defaults to testnet", "devnet", SuiTestnetRPC},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRPCEndpoint(tt.network)
			if got != tt.want {
				t.Errorf("GetRPCEndpoint(%q) = %q, want %q", tt.network, got, tt.want)
			}
		})
	}
}

func TestDefaultGasPrice(t *testing.T) {
	tests := []struct {
		name    string
		network string
		want    uint64
	}{
		{"testnet", "testnet", 750},
		{"mainnet", "mainnet", 1000},
		{"Testnet uppercase", "Testnet", 750},
		{"Mainnet uppercase", "Mainnet", 1000},
		{"empty defaults to testnet price", "", 750},
		{"unknown defaults to testnet price", "devnet", 750},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultGasPrice(tt.network)
			if got != tt.want {
				t.Errorf("DefaultGasPrice(%q) = %d, want %d", tt.network, got, tt.want)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero bytes", 0, "0 B"},
		{"1 byte", 1, "1 B"},
		{"1023 bytes", 1023, "1023 B"},
		{"1 KiB", 1024, "1.00 KiB"},
		{"1.5 KiB", 1536, "1.50 KiB"},
		{"1 MiB", 1024 * 1024, "1.00 MiB"},
		{"1 GiB", 1024 * 1024 * 1024, "1.00 GiB"},
		{"10 MiB", 10 * 1024 * 1024, "10.00 MiB"},
		{"500 KiB", 500 * 1024, "500.00 KiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatCostSummary(t *testing.T) {
	tests := []struct {
		name      string
		walCost   float64
		suiCost   float64
		fileCount int
		epochs    int
		wantParts []string
	}{
		{
			name:      "basic summary",
			walCost:   0.5,
			suiCost:   0.001,
			fileCount: 10,
			epochs:    5,
			wantParts: []string{"0.5000 WAL", "0.0010 SUI", "10 files", "5 epochs"},
		},
		{
			name:      "zero cost",
			walCost:   0,
			suiCost:   0,
			fileCount: 0,
			epochs:    1,
			wantParts: []string{"0.0000 WAL", "0.0000 SUI", "0 files", "1 epochs"},
		},
		{
			name:      "large values",
			walCost:   123.4567,
			suiCost:   0.9876,
			fileCount: 1000,
			epochs:    100,
			wantParts: []string{"123.4567 WAL", "0.9876 SUI", "1000 files", "100 epochs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCostSummary(tt.walCost, tt.suiCost, tt.fileCount, tt.epochs)
			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("FormatCostSummary() = %q, missing %q", got, part)
				}
			}
		})
	}
}

func TestFormatCostBreakdown(t *testing.T) {
	t.Run("full breakdown with all fields", func(t *testing.T) {
		breakdown := CostBreakdown{
			GasUnits:       600000,
			GasPrice:       750,
			GasCostSUI:     0.00045,
			StorageCostWAL: 0.001,
			WriteCostWAL:   0.0005,
			TotalWAL:       0.0015,
			EncodedSize:    8 * 1024 * 1024,
			OriginalSize:   1024 * 1024,
			FileCount:      10,
			Epochs:         5,
			MinTotalWAL:    0.0012,
			MaxTotalWAL:    0.0018,
			MinTotalSUI:    0.000315,
			MaxTotalSUI:    0.000675,
		}

		got := FormatCostBreakdown(breakdown)

		expectedParts := []string{
			"Cost Breakdown",
			"Data Size:",
			"Original:",
			"Encoded:",
			"Reed-Solomon",
			"Storage Cost (WAL):",
			"Storage:",
			"Write:",
			"Total:",
			"Transaction Cost (SUI):",
			"Gas Units:",
			"Gas Price:",
			"MIST",
			"Files: 10",
			"WAL is used for Walrus storage",
		}

		for _, part := range expectedParts {
			if !strings.Contains(got, part) {
				t.Errorf("FormatCostBreakdown() missing %q in output:\n%s", part, got)
			}
		}
	})

	t.Run("breakdown without storage costs", func(t *testing.T) {
		breakdown := CostBreakdown{
			GasUnits:    100000,
			GasPrice:    750,
			GasCostSUI:  0.000075,
			MinTotalSUI: 0.0000525,
			MaxTotalSUI: 0.0001125,
		}

		got := FormatCostBreakdown(breakdown)

		// Should have transaction info but not storage info
		if !strings.Contains(got, "Transaction Cost (SUI):") {
			t.Error("should contain transaction cost section")
		}
		if strings.Contains(got, "Storage Cost (WAL):") {
			t.Error("should not contain storage cost section when TotalWAL is 0")
		}
		if strings.Contains(got, "Data Size:") {
			t.Error("should not contain data size section when OriginalSize is 0")
		}
	})

	t.Run("breakdown without file count", func(t *testing.T) {
		breakdown := CostBreakdown{
			GasUnits:   50000,
			GasPrice:   1000,
			GasCostSUI: 0.00005,
		}

		got := FormatCostBreakdown(breakdown)

		if strings.Contains(got, "Files:") {
			t.Error("should not contain Files line when FileCount is 0")
		}
	})
}

func TestParseStorageInfoJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(t *testing.T, info *StorageInfo)
	}{
		{
			name:    "valid walrus info output",
			input:   `{"epochInfo": {"currentEpoch": 42, "startOfCurrentEpoch": {"DateTime": "2025-01-01T00:00:00Z"}, "epochDuration": {"secs": 86400, "nanos": 0}, "maxEpochsAhead": 365}, "storageInfo": {"nShards": 1000, "nNodes": 100}, "sizeInfo": {"storageUnitSize": 1048576, "maxBlobSize": 13958643712}, "priceInfo": {"storagePricePerUnitSize": 1000, "writePricePerUnitSize": 2000, "encodingDependentPriceInfo": [{"marginalSize": 1048576, "metadataPrice": 62000, "marginalPrice": 6000, "encodingType": "RS2", "exampleBlobs": [{"unencodedSize": 16777216, "encodedSize": 134217728, "price": 100000, "encodingType": "RS2"}]}]}}`,
			wantErr: false,
			checkFunc: func(t *testing.T, info *StorageInfo) {
				if info.CurrentEpoch != 42 {
					t.Errorf("CurrentEpoch = %d, want 42", info.CurrentEpoch)
				}
				if info.EpochDuration != 86400 {
					t.Errorf("EpochDuration = %d, want 86400", info.EpochDuration)
				}
				if info.StoragePrice != 1000 {
					t.Errorf("StoragePrice = %d, want 1000", info.StoragePrice)
				}
				if info.WritePrice != 2000 {
					t.Errorf("WritePrice = %d, want 2000", info.WritePrice)
				}
				if info.MetadataPrice != 62000 {
					t.Errorf("MetadataPrice = %d, want 62000", info.MetadataPrice)
				}
				if info.MarginalPrice != 6000 {
					t.Errorf("MarginalPrice = %d, want 6000", info.MarginalPrice)
				}
				if info.MaxBlobSize != 13958643712 {
					t.Errorf("MaxBlobSize = %d, want 13958643712", info.MaxBlobSize)
				}
				if info.StorageUnitSize != 1048576 {
					t.Errorf("StorageUnitSize = %d, want 1048576", info.StorageUnitSize)
				}
				if info.NumShards != 1000 {
					t.Errorf("NumShards = %d, want 1000", info.NumShards)
				}
				if info.MaxEpochsAhead != 365 {
					t.Errorf("MaxEpochsAhead = %d, want 365", info.MaxEpochsAhead)
				}
				// Encoding multiplier should be calculated from example blob
				expectedMultiplier := float64(134217728) / float64(16777216) // 8.0
				if math.Abs(info.EncodingMultiplier-expectedMultiplier) > 0.01 {
					t.Errorf("EncodingMultiplier = %v, want %v", info.EncodingMultiplier, expectedMultiplier)
				}
			},
		},
		{
			name:    "output with log lines before JSON",
			input:   "INFO: Starting walrus daemon\nDEBUG: Connecting to network\n{\"epochInfo\": {\"currentEpoch\": 10, \"startOfCurrentEpoch\": {\"DateTime\": \"2025-01-01T00:00:00Z\"}, \"epochDuration\": {\"secs\": 3600, \"nanos\": 0}, \"maxEpochsAhead\": 100}, \"storageInfo\": {\"nShards\": 500, \"nNodes\": 50}, \"sizeInfo\": {\"storageUnitSize\": 1048576, \"maxBlobSize\": 1000000}, \"priceInfo\": {\"storagePricePerUnitSize\": 500, \"writePricePerUnitSize\": 1000, \"encodingDependentPriceInfo\": []}}",
			wantErr: false,
			checkFunc: func(t *testing.T, info *StorageInfo) {
				if info.CurrentEpoch != 10 {
					t.Errorf("CurrentEpoch = %d, want 10", info.CurrentEpoch)
				}
				if info.StoragePrice != 500 {
					t.Errorf("StoragePrice = %d, want 500", info.StoragePrice)
				}
				// No encoding info, should use default multiplier
				if info.EncodingMultiplier != 5.0 {
					t.Errorf("EncodingMultiplier = %v, want 5.0 (default)", info.EncodingMultiplier)
				}
			},
		},
		{
			name:    "no JSON in output",
			input:   "This is just log output with no JSON\nAnother log line",
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			input:   "{invalid json content",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "JSON with no encoding dependent price info",
			input:   `{"epochInfo": {"currentEpoch": 5, "startOfCurrentEpoch": {"DateTime": "2025-01-01T00:00:00Z"}, "epochDuration": {"secs": 43200, "nanos": 0}, "maxEpochsAhead": 50}, "storageInfo": {"nShards": 100, "nNodes": 10}, "sizeInfo": {"storageUnitSize": 1048576, "maxBlobSize": 5000000}, "priceInfo": {"storagePricePerUnitSize": 2000, "writePricePerUnitSize": 4000, "encodingDependentPriceInfo": []}}`,
			wantErr: false,
			checkFunc: func(t *testing.T, info *StorageInfo) {
				if info.MetadataPrice != 0 {
					t.Errorf("MetadataPrice = %d, want 0 (no encoding info)", info.MetadataPrice)
				}
				if info.MarginalPrice != 0 {
					t.Errorf("MarginalPrice = %d, want 0 (no encoding info)", info.MarginalPrice)
				}
				// Default encoding multiplier when no examples provided
				if info.EncodingMultiplier != 5.0 {
					t.Errorf("EncodingMultiplier = %v, want 5.0 (default)", info.EncodingMultiplier)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseStorageInfoJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("ParseStorageInfoJSON() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseStorageInfoJSON() unexpected error: %v", err)
			}
			if info == nil {
				t.Fatal("ParseStorageInfoJSON() returned nil info")
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, info)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	t.Run("valid cost calculation with manual gas price", func(t *testing.T) {
		options := CostOptions{
			SiteSize:  1024 * 1024, // 1 MiB
			Epochs:    5,
			FileCount: 10,
			GasPrice:  750, // Manual gas price to avoid network call
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test", // Force fallback to defaults
		}

		breakdown, err := CalculateCost(options)
		if err != nil {
			t.Fatalf("CalculateCost() error = %v", err)
		}

		if breakdown == nil {
			t.Fatal("CalculateCost() returned nil")
		}

		// Basic sanity checks
		if breakdown.OriginalSize != options.SiteSize {
			t.Errorf("OriginalSize = %d, want %d", breakdown.OriginalSize, options.SiteSize)
		}
		if breakdown.Epochs != options.Epochs {
			t.Errorf("Epochs = %d, want %d", breakdown.Epochs, options.Epochs)
		}
		if breakdown.FileCount != options.FileCount {
			t.Errorf("FileCount = %d, want %d", breakdown.FileCount, options.FileCount)
		}
		if breakdown.GasPrice != 750 {
			t.Errorf("GasPrice = %d, want 750", breakdown.GasPrice)
		}

		// Encoded size should be larger than original
		if breakdown.EncodedSize <= breakdown.OriginalSize {
			t.Errorf("EncodedSize (%d) should be larger than OriginalSize (%d)",
				breakdown.EncodedSize, breakdown.OriginalSize)
		}

		// Costs should be positive
		if breakdown.TotalWAL <= 0 {
			t.Errorf("TotalWAL = %v, should be positive", breakdown.TotalWAL)
		}
		if breakdown.GasCostSUI <= 0 {
			t.Errorf("GasCostSUI = %v, should be positive", breakdown.GasCostSUI)
		}
		if breakdown.StorageCostWAL <= 0 {
			t.Errorf("StorageCostWAL = %v, should be positive", breakdown.StorageCostWAL)
		}
		if breakdown.WriteCostWAL <= 0 {
			t.Errorf("WriteCostWAL = %v, should be positive", breakdown.WriteCostWAL)
		}

		// Ranges should bracket the total
		if breakdown.MinTotalWAL > breakdown.TotalWAL {
			t.Errorf("MinTotalWAL (%v) > TotalWAL (%v)", breakdown.MinTotalWAL, breakdown.TotalWAL)
		}
		if breakdown.MaxTotalWAL < breakdown.TotalWAL {
			t.Errorf("MaxTotalWAL (%v) < TotalWAL (%v)", breakdown.MaxTotalWAL, breakdown.TotalWAL)
		}
		if breakdown.MinTotalSUI > breakdown.GasCostSUI {
			t.Errorf("MinTotalSUI (%v) > GasCostSUI (%v)", breakdown.MinTotalSUI, breakdown.GasCostSUI)
		}
		if breakdown.MaxTotalSUI < breakdown.GasCostSUI {
			t.Errorf("MaxTotalSUI (%v) < GasCostSUI (%v)", breakdown.MaxTotalSUI, breakdown.GasCostSUI)
		}
	})

	t.Run("zero site size returns error", func(t *testing.T) {
		_, err := CalculateCost(CostOptions{
			SiteSize: 0,
			Epochs:   5,
			GasPrice: 750,
		})
		if err == nil {
			t.Error("CalculateCost() should error for zero site size")
		}
	})

	t.Run("negative site size returns error", func(t *testing.T) {
		_, err := CalculateCost(CostOptions{
			SiteSize: -100,
			Epochs:   5,
			GasPrice: 750,
		})
		if err == nil {
			t.Error("CalculateCost() should error for negative site size")
		}
	})

	t.Run("zero epochs returns error", func(t *testing.T) {
		_, err := CalculateCost(CostOptions{
			SiteSize: 1024,
			Epochs:   0,
			GasPrice: 750,
		})
		if err == nil {
			t.Error("CalculateCost() should error for zero epochs")
		}
	})

	t.Run("negative epochs returns error", func(t *testing.T) {
		_, err := CalculateCost(CostOptions{
			SiteSize: 1024,
			Epochs:   -1,
			GasPrice: 750,
		})
		if err == nil {
			t.Error("CalculateCost() should error for negative epochs")
		}
	})

	t.Run("file count estimation when not provided", func(t *testing.T) {
		options := CostOptions{
			SiteSize:  1024 * 1024, // 1 MiB
			Epochs:    1,
			FileCount: 0, // Should be estimated
			GasPrice:  750,
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		breakdown, err := CalculateCost(options)
		if err != nil {
			t.Fatalf("CalculateCost() error = %v", err)
		}

		// File count should be estimated (1 MiB / 50 KiB avg = ~20 files)
		if breakdown.FileCount <= 0 {
			t.Errorf("FileCount = %d, should be estimated to positive value", breakdown.FileCount)
		}

		expectedEstimate := int(math.Ceil(float64(1024*1024) / (50 * 1024)))
		if breakdown.FileCount != expectedEstimate {
			t.Errorf("FileCount = %d, want %d (estimated)", breakdown.FileCount, expectedEstimate)
		}
	})

	t.Run("very small site gets minimum 1 file", func(t *testing.T) {
		options := CostOptions{
			SiteSize:  100, // 100 bytes
			Epochs:    1,
			FileCount: 0,
			GasPrice:  750,
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		breakdown, err := CalculateCost(options)
		if err != nil {
			t.Fatalf("CalculateCost() error = %v", err)
		}

		if breakdown.FileCount < 1 {
			t.Errorf("FileCount = %d, should be at least 1", breakdown.FileCount)
		}
	})

	t.Run("more epochs increases WAL cost", func(t *testing.T) {
		baseOptions := CostOptions{
			SiteSize:  1024 * 1024,
			Epochs:    1,
			FileCount: 10,
			GasPrice:  750,
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		moreEpochsOptions := CostOptions{
			SiteSize:  1024 * 1024,
			Epochs:    10,
			FileCount: 10,
			GasPrice:  750,
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		breakdown1, err := CalculateCost(baseOptions)
		if err != nil {
			t.Fatalf("CalculateCost(1 epoch) error = %v", err)
		}

		breakdown10, err := CalculateCost(moreEpochsOptions)
		if err != nil {
			t.Fatalf("CalculateCost(10 epochs) error = %v", err)
		}

		if breakdown10.TotalWAL <= breakdown1.TotalWAL {
			t.Errorf("10 epochs cost (%v WAL) should be > 1 epoch cost (%v WAL)",
				breakdown10.TotalWAL, breakdown1.TotalWAL)
		}

		// Storage cost should scale with epochs, write cost should stay the same
		if breakdown10.StorageCostWAL <= breakdown1.StorageCostWAL {
			t.Errorf("10 epochs storage (%v) should be > 1 epoch storage (%v)",
				breakdown10.StorageCostWAL, breakdown1.StorageCostWAL)
		}
	})

	t.Run("larger site increases cost", func(t *testing.T) {
		smallOptions := CostOptions{
			SiteSize:  1024 * 1024, // 1 MiB
			Epochs:    5,
			FileCount: 10,
			GasPrice:  750,
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		largeOptions := CostOptions{
			SiteSize:  100 * 1024 * 1024, // 100 MiB
			Epochs:    5,
			FileCount: 10,
			GasPrice:  750,
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		small, err := CalculateCost(smallOptions)
		if err != nil {
			t.Fatalf("CalculateCost(small) error = %v", err)
		}

		large, err := CalculateCost(largeOptions)
		if err != nil {
			t.Fatalf("CalculateCost(large) error = %v", err)
		}

		if large.TotalWAL <= small.TotalWAL {
			t.Errorf("larger site cost (%v WAL) should be > smaller site cost (%v WAL)",
				large.TotalWAL, small.TotalWAL)
		}
	})

	t.Run("mainnet uses higher default pricing", func(t *testing.T) {
		testnetOptions := CostOptions{
			SiteSize:  1024 * 1024,
			Epochs:    1,
			FileCount: 10,
			GasPrice:  DefaultGasPrice("testnet"),
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		mainnetOptions := CostOptions{
			SiteSize:  1024 * 1024,
			Epochs:    1,
			FileCount: 10,
			GasPrice:  DefaultGasPrice("mainnet"),
			Network:   "mainnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		testnet, err := CalculateCost(testnetOptions)
		if err != nil {
			t.Fatalf("CalculateCost(testnet) error = %v", err)
		}

		mainnet, err := CalculateCost(mainnetOptions)
		if err != nil {
			t.Fatalf("CalculateCost(mainnet) error = %v", err)
		}

		// Mainnet should have higher WAL cost due to higher storage pricing
		if mainnet.TotalWAL <= testnet.TotalWAL {
			t.Errorf("mainnet WAL cost (%v) should be > testnet WAL cost (%v)",
				mainnet.TotalWAL, testnet.TotalWAL)
		}
	})

	t.Run("more files increases gas cost", func(t *testing.T) {
		fewFiles := CostOptions{
			SiteSize:  1024 * 1024,
			Epochs:    1,
			FileCount: 1,
			GasPrice:  750,
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		manyFiles := CostOptions{
			SiteSize:  1024 * 1024,
			Epochs:    1,
			FileCount: 100,
			GasPrice:  750,
			Network:   "testnet",
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		few, err := CalculateCost(fewFiles)
		if err != nil {
			t.Fatalf("CalculateCost(few files) error = %v", err)
		}

		many, err := CalculateCost(manyFiles)
		if err != nil {
			t.Fatalf("CalculateCost(many files) error = %v", err)
		}

		if many.GasCostSUI <= few.GasCostSUI {
			t.Errorf("100 files gas cost (%v SUI) should be > 1 file gas cost (%v SUI)",
				many.GasCostSUI, few.GasCostSUI)
		}
	})
}

func TestCostBreakdownRanges(t *testing.T) {
	options := CostOptions{
		SiteSize:  5 * 1024 * 1024, // 5 MiB
		Epochs:    3,
		FileCount: 20,
		GasPrice:  750,
		Network:   "testnet",
		WalrusBin: "/nonexistent/walrus-for-test",
	}

	breakdown, err := CalculateCost(options)
	if err != nil {
		t.Fatalf("CalculateCost() error = %v", err)
	}

	// MinTotalWAL should be 80% of TotalWAL
	expectedMinWAL := breakdown.TotalWAL * 0.8
	if math.Abs(breakdown.MinTotalWAL-expectedMinWAL) > 0.0001 {
		t.Errorf("MinTotalWAL = %v, want %v (80%% of TotalWAL)", breakdown.MinTotalWAL, expectedMinWAL)
	}

	// MaxTotalWAL should be 120% of TotalWAL
	expectedMaxWAL := breakdown.TotalWAL * 1.2
	if math.Abs(breakdown.MaxTotalWAL-expectedMaxWAL) > 0.0001 {
		t.Errorf("MaxTotalWAL = %v, want %v (120%% of TotalWAL)", breakdown.MaxTotalWAL, expectedMaxWAL)
	}

	// MinTotalSUI should be 70% of GasCostSUI
	expectedMinSUI := breakdown.GasCostSUI * 0.7
	if math.Abs(breakdown.MinTotalSUI-expectedMinSUI) > 0.0001 {
		t.Errorf("MinTotalSUI = %v, want %v (70%% of GasCostSUI)", breakdown.MinTotalSUI, expectedMinSUI)
	}

	// MaxTotalSUI should be 150% of GasCostSUI
	expectedMaxSUI := breakdown.GasCostSUI * 1.5
	if math.Abs(breakdown.MaxTotalSUI-expectedMaxSUI) > 0.0001 {
		t.Errorf("MaxTotalSUI = %v, want %v (150%% of GasCostSUI)", breakdown.MaxTotalSUI, expectedMaxSUI)
	}
}

func TestStripANSICodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no ANSI codes",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "single color code",
			input: "\x1b[32mgreen text\x1b[0m",
			want:  "green text",
		},
		{
			name:  "multiple color codes",
			input: "\x1b[1m\x1b[33mwarning\x1b[0m normal",
			want:  "warning normal",
		},
		{
			name:  "complex ANSI sequence",
			input: "\x1b[1;32;40mbold green on black\x1b[0m",
			want:  "bold green on black",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(stripANSI([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSuiRPCConstants(t *testing.T) {
	if SuiTestnetRPC == "" {
		t.Error("SuiTestnetRPC should not be empty")
	}
	if SuiMainnetRPC == "" {
		t.Error("SuiMainnetRPC should not be empty")
	}
	if !strings.HasPrefix(SuiTestnetRPC, "https://") {
		t.Errorf("SuiTestnetRPC should use HTTPS, got %q", SuiTestnetRPC)
	}
	if !strings.HasPrefix(SuiMainnetRPC, "https://") {
		t.Errorf("SuiMainnetRPC should use HTTPS, got %q", SuiMainnetRPC)
	}
	if !strings.Contains(SuiTestnetRPC, "testnet") {
		t.Errorf("SuiTestnetRPC should contain 'testnet', got %q", SuiTestnetRPC)
	}
	if !strings.Contains(SuiMainnetRPC, "mainnet") {
		t.Errorf("SuiMainnetRPC should contain 'mainnet', got %q", SuiMainnetRPC)
	}
}

func TestCostOptionsDefaults(t *testing.T) {
	t.Run("empty network defaults handled", func(t *testing.T) {
		options := CostOptions{
			SiteSize:  1024,
			Epochs:    1,
			GasPrice:  750,
			Network:   "", // empty network
			WalrusBin: "/nonexistent/walrus-for-test",
		}

		breakdown, err := CalculateCost(options)
		if err != nil {
			t.Fatalf("CalculateCost() error = %v", err)
		}

		if breakdown == nil {
			t.Fatal("CalculateCost() returned nil for empty network")
		}

		if breakdown.TotalWAL <= 0 {
			t.Error("TotalWAL should be positive even with empty network")
		}
	})
}

func TestEncodedSizeMinimum(t *testing.T) {
	// Very small site should still get minimum 1 storage unit for encoded size MiB calculation
	options := CostOptions{
		SiteSize:  1, // 1 byte
		Epochs:    1,
		FileCount: 1,
		GasPrice:  750,
		Network:   "testnet",
		WalrusBin: "/nonexistent/walrus-for-test",
	}

	breakdown, err := CalculateCost(options)
	if err != nil {
		t.Fatalf("CalculateCost() error = %v", err)
	}

	// Even 1 byte should produce a positive cost (minimum 1 storage unit)
	if breakdown.TotalWAL <= 0 {
		t.Errorf("TotalWAL = %v, should be positive even for 1 byte", breakdown.TotalWAL)
	}
}
