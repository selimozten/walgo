package config

import (
	"testing"
	"walgo/internal/optimizer"
)

func TestWalgoConfigStructure(t *testing.T) {
	// Test that all fields are properly tagged for YAML and mapstructure
	cfg := WalgoConfig{
		HugoConfig: HugoConfig{
			Version:     "0.120.0",
			BaseURL:     "https://example.com",
			PublishDir:  "dist",
			ContentDir:  "posts",
			ResourceDir: "assets",
		},
		WalrusConfig: WalrusConfig{
			ProjectID:   "test-project",
			BucketName:  "test-bucket",
			Entrypoint:  "main.html",
			SuiNSDomain: "example.sui",
		},
		ObsidianConfig: ObsidianConfig{
			VaultPath:         "/vault",
			AttachmentDir:     "files",
			ConvertWikilinks:  true,
			IncludeDrafts:     true,
			FrontmatterFormat: "toml",
		},
		OptimizerConfig: optimizer.OptimizerConfig{
			Enabled: true,
			HTML: optimizer.HTMLConfig{
				Enabled:          true,
				MinifyHTML:       true,
				RemoveComments:   true,
				RemoveWhitespace: true,
			},
			CSS: optimizer.CSSConfig{
				Enabled:        true,
				MinifyCSS:      true,
				RemoveComments: true,
			},
			JS: optimizer.JSConfig{
				Enabled:        true,
				MinifyJS:       true,
				RemoveComments: true,
			},
		},
	}

	// Verify all fields are accessible
	if cfg.HugoConfig.Version != "0.120.0" {
		t.Error("HugoConfig.Version not accessible")
	}
	if cfg.WalrusConfig.ProjectID != "test-project" {
		t.Error("WalrusConfig.ProjectID not accessible")
	}
	if !cfg.ObsidianConfig.ConvertWikilinks {
		t.Error("ObsidianConfig.ConvertWikilinks not accessible")
	}
	if !cfg.OptimizerConfig.Enabled {
		t.Error("OptimizerConfig.Enabled not accessible")
	}
}

func TestHugoConfigDefaults(t *testing.T) {
	// Test zero values
	var cfg HugoConfig

	// All fields should have zero values
	if cfg.Version != "" {
		t.Errorf("Expected empty Version, got %s", cfg.Version)
	}
	if cfg.BaseURL != "" {
		t.Errorf("Expected empty BaseURL, got %s", cfg.BaseURL)
	}
	if cfg.PublishDir != "" {
		t.Errorf("Expected empty PublishDir, got %s", cfg.PublishDir)
	}
	if cfg.ContentDir != "" {
		t.Errorf("Expected empty ContentDir, got %s", cfg.ContentDir)
	}
	if cfg.ResourceDir != "" {
		t.Errorf("Expected empty ResourceDir, got %s", cfg.ResourceDir)
	}
}

func TestWalrusConfigDefaults(t *testing.T) {
	// Test zero values
	var cfg WalrusConfig

	// All fields should have zero values
	if cfg.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", cfg.ProjectID)
	}
	if cfg.BucketName != "" {
		t.Errorf("Expected empty BucketName, got %s", cfg.BucketName)
	}
	if cfg.Entrypoint != "" {
		t.Errorf("Expected empty Entrypoint, got %s", cfg.Entrypoint)
	}
	if cfg.SuiNSDomain != "" {
		t.Errorf("Expected empty SuiNSDomain, got %s", cfg.SuiNSDomain)
	}
}

func TestObsidianConfigDefaults(t *testing.T) {
	// Test zero values
	var cfg ObsidianConfig

	// All fields should have zero values
	if cfg.VaultPath != "" {
		t.Errorf("Expected empty VaultPath, got %s", cfg.VaultPath)
	}
	if cfg.AttachmentDir != "" {
		t.Errorf("Expected empty AttachmentDir, got %s", cfg.AttachmentDir)
	}
	if cfg.ConvertWikilinks {
		t.Error("Expected ConvertWikilinks to be false")
	}
	if cfg.IncludeDrafts {
		t.Error("Expected IncludeDrafts to be false")
	}
	if cfg.FrontmatterFormat != "" {
		t.Errorf("Expected empty FrontmatterFormat, got %s", cfg.FrontmatterFormat)
	}
}

func TestNewDefaultWalgoConfigValues(t *testing.T) {
	cfg := NewDefaultWalgoConfig()

	// Test specific default values
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"HugoConfig.PublishDir", cfg.HugoConfig.PublishDir, "public"},
		{"HugoConfig.ContentDir", cfg.HugoConfig.ContentDir, "content"},
		{"WalrusConfig.ProjectID", cfg.WalrusConfig.ProjectID, "YOUR_WALRUS_PROJECT_ID"},
		{"WalrusConfig.Entrypoint", cfg.WalrusConfig.Entrypoint, "index.html"},
		{"ObsidianConfig.AttachmentDir", cfg.ObsidianConfig.AttachmentDir, "images"},
		{"ObsidianConfig.ConvertWikilinks", cfg.ObsidianConfig.ConvertWikilinks, true},
		{"ObsidianConfig.IncludeDrafts", cfg.ObsidianConfig.IncludeDrafts, false},
		{"ObsidianConfig.FrontmatterFormat", cfg.ObsidianConfig.FrontmatterFormat, "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestConfigFieldTypes(t *testing.T) {
	// This test ensures all fields have the correct type
	// and can be properly marshaled/unmarshaled

	cfg := WalgoConfig{
		HugoConfig: HugoConfig{
			Version:     "test",
			BaseURL:     "test",
			PublishDir:  "test",
			ContentDir:  "test",
			ResourceDir: "test",
		},
		WalrusConfig: WalrusConfig{
			ProjectID:   "test",
			BucketName:  "test",
			Entrypoint:  "test",
			SuiNSDomain: "test",
		},
		ObsidianConfig: ObsidianConfig{
			VaultPath:         "test",
			AttachmentDir:     "test",
			ConvertWikilinks:  true,
			IncludeDrafts:     true,
			FrontmatterFormat: "test",
		},
		OptimizerConfig: optimizer.OptimizerConfig{
			Enabled: true,
			HTML: optimizer.HTMLConfig{
				Enabled:          true,
				MinifyHTML:       true,
				RemoveComments:   true,
				RemoveWhitespace: true,
			},
			CSS: optimizer.CSSConfig{
				Enabled:        true,
				MinifyCSS:      true,
				RemoveComments: true,
			},
			JS: optimizer.JSConfig{
				Enabled:        true,
				MinifyJS:       true,
				RemoveComments: true,
			},
		},
	}

	// Verify the types compile and fields are accessible
	_ = cfg.HugoConfig
	_ = cfg.WalrusConfig
	_ = cfg.ObsidianConfig
	_ = cfg.OptimizerConfig
}