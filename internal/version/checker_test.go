package version

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "semantic version with v prefix",
			input:    "sui 1.35.0-abc123",
			expected: "1.35.0-abc123",
		},
		{
			name:     "semantic version without v",
			input:    "walrus 2.1.3",
			expected: "2.1.3",
		},
		{
			name:     "version with v prefix",
			input:    "v1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "simple version",
			input:    "version 1.2",
			expected: "1.2",
		},
		{
			name:     "complex output",
			input:    "site-builder version: 1.0.5-beta+meta",
			expected: "1.0.5-beta",
		},
		{
			name:     "no version",
			input:    "no version info",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersion(tt.input)
			if result != tt.expected {
				t.Errorf("parseVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "v1 greater than v2",
			v1:       "1.2.3",
			v2:       "1.2.2",
			expected: 1,
		},
		{
			name:     "v1 less than v2",
			v1:       "1.2.1",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "v1 equals v2",
			v1:       "1.2.3",
			v2:       "1.2.3",
			expected: 0,
		},
		{
			name:     "major version difference",
			v1:       "2.0.0",
			v2:       "1.9.9",
			expected: 1,
		},
		{
			name:     "minor version difference",
			v1:       "1.5.0",
			v2:       "1.4.9",
			expected: 1,
		},
		{
			name:     "with pre-release tags",
			v1:       "1.2.3-beta",
			v2:       "1.2.2",
			expected: 1,
		},
		{
			name:     "simple versions",
			v1:       "1.2",
			v2:       "1.1",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestCheckCompatibility(t *testing.T) {
	tests := []struct {
		name       string
		walrus     string
		sb         string
		wantErr    bool
		errContain string
	}{
		{
			name:    "minimum supported versions",
			walrus:  "1.52.0",
			sb:      "2.12.0",
			wantErr: false,
		},
		{
			name:    "current versions",
			walrus:  "1.53.0-42bc13730da8",
			sb:      "2.12.0-1cd9e75a5654",
			wantErr: false,
		},
		{
			name:       "walrus too old",
			walrus:     "1.51.2",
			sb:         "2.12.0",
			wantErr:    true,
			errContain: "walrus v1.51.2",
		},
		{
			name:       "site-builder too old",
			walrus:     "1.52.1",
			sb:         "2.11.0",
			wantErr:    true,
			errContain: "site-builder v2.11.0",
		},
		{
			name:       "both too old",
			walrus:     "1.49.1",
			sb:         "2.10.0",
			wantErr:    true,
			errContain: "and",
		},
		{
			name:    "empty walrus version",
			walrus:  "",
			sb:      "2.12.0",
			wantErr: false,
		},
		{
			name:    "empty sb version",
			walrus:  "1.52.1",
			sb:      "",
			wantErr: false,
		},
		{
			name:    "both empty",
			walrus:  "",
			sb:      "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckCompatibility(tt.walrus, tt.sb)
			if tt.wantErr && err == nil {
				t.Errorf("CheckCompatibility(%q, %q) = nil, want error containing %q", tt.walrus, tt.sb, tt.errContain)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("CheckCompatibility(%q, %q) = %v, want nil", tt.walrus, tt.sb, err)
			}
			if tt.wantErr && err != nil && tt.errContain != "" {
				if !contains(err.Error(), tt.errContain) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContain)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseVersionParts(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected [3]int
	}{
		{
			name:     "full semantic version",
			version:  "1.2.3",
			expected: [3]int{1, 2, 3},
		},
		{
			name:     "with pre-release",
			version:  "1.2.3-beta",
			expected: [3]int{1, 2, 3},
		},
		{
			name:     "two part version",
			version:  "1.2",
			expected: [3]int{1, 2, 0},
		},
		{
			name:     "single digit",
			version:  "1",
			expected: [3]int{1, 0, 0},
		},
		{
			name:     "empty version",
			version:  "",
			expected: [3]int{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersionParts(tt.version)
			if result != tt.expected {
				t.Errorf("parseVersionParts(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}
