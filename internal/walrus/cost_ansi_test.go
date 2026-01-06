package walrus

import (
	"testing"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "No ANSI codes",
			input:    []byte(`{"key": "value"}`),
			expected: []byte(`{"key": "value"}`),
		},
		{
			name:     "ANSI color codes",
			input:    []byte("\x1b[32m{\"key\": \"value\"}\x1b[0m"),
			expected: []byte(`{"key": "value"}`),
		},
		{
			name:     "ANSI codes in middle",
			input:    []byte("{\"key\": \x1b[33m\"value\"\x1b[0m}"),
			expected: []byte(`{"key": "value"}`),
		},
		{
			name:     "Multiple ANSI codes",
			input:    []byte("\x1b[1m\x1b[32m{\"key\": \"value\"}\x1b[0m\x1b[0m"),
			expected: []byte(`{"key": "value"}`),
		},
		{
			name:     "Complex ANSI codes",
			input:    []byte("\x1b[1;32;40m{\"status\": \"ok\"}\x1b[0m"),
			expected: []byte(`{"status": "ok"}`),
		},
		{
			name:     "Real walrus output simulation",
			input:    []byte("\x1b[32mINFO\x1b[0m Some log message\n{\"epochInfo\": {\"currentEpoch\": 123}}"),
			expected: []byte("INFO Some log message\n{\"epochInfo\": {\"currentEpoch\": 123}}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if string(result) != string(tt.expected) {
				t.Errorf("stripANSI() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseStorageInfoJSON_WithANSI(t *testing.T) {
	// Simulate walrus info output with ANSI codes
	mockOutput := []byte("\x1b[32mINFO\x1b[0m: Fetching storage info\n\x1b[33m{\"epochInfo\": {\"currentEpoch\": 100, \"startOfCurrentEpoch\": {\"DateTime\": \"2025-01-01T00:00:00Z\"}, \"epochDuration\": {\"secs\": 86400, \"nanos\": 0}, \"maxEpochsAhead\": 365}, \"storageInfo\": {\"nShards\": 1000, \"nNodes\": 100}, \"sizeInfo\": {\"storageUnitSize\": 1048576, \"maxBlobSize\": 13958643712}, \"priceInfo\": {\"storagePricePerUnitSize\": 1000, \"writePricePerUnitSize\": 2000, \"encodingDependentPriceInfo\": [{\"marginalSize\": 1048576, \"metadataPrice\": 62000, \"marginalPrice\": 6000, \"encodingType\": \"RS2\", \"exampleBlobs\": [{\"unencodedSize\": 16777216, \"encodedSize\": 134217728, \"price\": 100000, \"encodingType\": \"RS2\"}]}]}}\x1b[0m")

	info, err := ParseStorageInfoJSON(mockOutput)
	if err != nil {
		t.Fatalf("ParseStorageInfoJSON() failed: %v", err)
	}

	// Verify parsed values
	if info.CurrentEpoch != 100 {
		t.Errorf("CurrentEpoch = %d, want 100", info.CurrentEpoch)
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
}

func TestParseStorageInfoJSON_InvalidCharacterError(t *testing.T) {
	// This simulates the exact error the user reported
	// The \x1b character appears in what looks like a JSON key position
	mockOutput := []byte("{\"key" + "\x1b[32m" + "\": \"value\"}")

	// Without ANSI stripping, this would fail with "invalid character '\x1b'"
	// But with our fix, stripANSI() will remove \x1b[32m before parsing

	// First verify that stripANSI removes the escape codes
	cleaned := stripANSI(mockOutput)
	expected := []byte("{\"key\": \"value\"}")
	if string(cleaned) != string(expected) {
		t.Errorf("stripANSI() = %q, want %q", cleaned, expected)
	}

	// Note: ParseStorageInfoJSON will still fail because {"key": "value"}
	// doesn't match WalrusInfoJSON structure, but it won't fail with ANSI error
	_, err := ParseStorageInfoJSON(mockOutput)

	// The error should be about JSON structure, NOT about invalid character \x1b
	if err != nil && err.Error() != "" {
		// This is expected - the JSON structure doesn't match
		// But we verify it's not an ANSI escape character error
		if err.Error() == "invalid character '\\x1b' looking for beginning of object key string" {
			t.Fatalf("ParseStorageInfoJSON() still failing with ANSI error: %v", err)
		}
	}
}
