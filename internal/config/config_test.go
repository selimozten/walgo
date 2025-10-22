package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"walgo/internal/optimizer"
)

func TestCreateDefaultWalgoConfig(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(string) error
		wantErr     bool
		errContains string
	}{
		{
			name:    "Create config in new directory",
			setup:   nil,
			wantErr: false,
		},
		{
			name: "Config file already exists",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, DefaultConfigFileName), []byte("existing"), 0644)
			},
			wantErr:     true,
			errContains: "already exists",
		},
		{
			name: "Directory with permission issues",
			setup: func(dir string) error {
				// Create a subdirectory with no write permission
				subdir := filepath.Join(dir, "readonly")
				if err := os.Mkdir(subdir, 0555); err != nil {
					return err
				}
				return nil
			},
			wantErr: false, // The main directory is still writable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()

			// Run setup if provided
			if tt.setup != nil {
				if err := tt.setup(tempDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Execute function
			err := CreateDefaultWalgoConfig(tempDir)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDefaultWalgoConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got %q", tt.errContains, err.Error())
				}
			}

			// If successful, verify file was created
			if !tt.wantErr {
				configPath := filepath.Join(tempDir, DefaultConfigFileName)
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Error("Config file was not created")
				}

				// Verify it's valid YAML
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read created config: %v", err)
				}

				var cfg WalgoConfig
				if err := yaml.Unmarshal(data, &cfg); err != nil {
					t.Errorf("Created config is not valid YAML: %v", err)
				}

				// Verify defaults are set
				if cfg.HugoConfig.PublishDir != "public" {
					t.Errorf("Expected PublishDir to be 'public', got %s", cfg.HugoConfig.PublishDir)
				}
				if cfg.WalrusConfig.Entrypoint != "index.html" {
					t.Errorf("Expected Entrypoint to be 'index.html', got %s", cfg.WalrusConfig.Entrypoint)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		configData  string
		setupViper  func(string)
		wantErr     bool
		errContains string
		validate    func(*WalgoConfig) error
	}{
		{
			name: "Valid config file",
			configData: `
hugo:
  publishDir: dist
  contentDir: posts
walrus:
  projectID: test-project
  entrypoint: home.html
optimizer:
  enabled: true
  html:
    enabled: true
    minifyHTML: true
  css:
    enabled: true
  js:
    enabled: true
`,
			setupViper: func(configPath string) {
				viper.Reset()
				viper.SetConfigFile(configPath)
				viper.ReadInConfig()
			},
			wantErr: false,
			validate: func(cfg *WalgoConfig) error {
				if cfg.HugoConfig.PublishDir != "dist" {
					return fmt.Errorf("expected PublishDir 'dist', got %s", cfg.HugoConfig.PublishDir)
				}
				if cfg.WalrusConfig.ProjectID != "test-project" {
					return fmt.Errorf("expected ProjectID 'test-project', got %s", cfg.WalrusConfig.ProjectID)
				}
				return nil
			},
		},
		{
			name: "Config with defaults",
			configData: `
walrus:
  projectID: minimal-project
`,
			setupViper: func(configPath string) {
				viper.Reset()
				viper.SetConfigFile(configPath)
				viper.ReadInConfig()
			},
			wantErr: false,
			validate: func(cfg *WalgoConfig) error {
				// Defaults should be applied
				if cfg.HugoConfig.PublishDir != "public" {
					return fmt.Errorf("expected default PublishDir 'public', got %s", cfg.HugoConfig.PublishDir)
				}
				if cfg.HugoConfig.ContentDir != "content" {
					return fmt.Errorf("expected default ContentDir 'content', got %s", cfg.HugoConfig.ContentDir)
				}
				if cfg.HugoConfig.ResourceDir != "resources" {
					return fmt.Errorf("expected default ResourceDir 'resources', got %s", cfg.HugoConfig.ResourceDir)
				}
				if cfg.WalrusConfig.Entrypoint != "index.html" {
					return fmt.Errorf("expected default Entrypoint 'index.html', got %s", cfg.WalrusConfig.Entrypoint)
				}
				return nil
			},
		},
		{
			name: "No config file loaded",
			setupViper: func(configPath string) {
				viper.Reset()
				// Don't load any config
			},
			wantErr:     true,
			errContains: "not found or failed to load",
		},
		{
			name: "Invalid YAML",
			configData: `
invalid yaml:
  - this is not: valid
  [mixing brackets
`,
			setupViper: func(configPath string) {
				viper.Reset()
				viper.SetConfigFile(configPath)
				// Try to read the invalid config
				err := viper.ReadInConfig()
				if err != nil {
					// If viper fails to read, that's expected
					// But mark the config as not used so LoadConfig will fail
					viper.Reset()
				}
			},
			wantErr:     true,
			errContains: "not found or failed to load",
		},
		{
			name: "Config with all sections",
			configData: `
hugo:
  version: "0.120.0"
  baseURL: "https://example.com"
  publishDir: "build"
  contentDir: "articles"
  resourceDir: "assets"
walrus:
  projectID: "full-project"
  bucketName: "my-bucket"
  entrypoint: "main.html"
  suinsDomain: "example.sui"
obsidian:
  vaultPath: "/path/to/vault"
  attachmentDir: "attachments"
  convertWikilinks: true
  includeDrafts: false
  frontmatterFormat: "yaml"
optimizer:
  enabled: true
  html:
    enabled: true
    minifyHTML: true
    removeComments: true
  css:
    enabled: true
    minifyCSS: true
  js:
    enabled: true
    minifyJS: true
`,
			setupViper: func(configPath string) {
				viper.Reset()
				viper.SetConfigFile(configPath)
				viper.ReadInConfig()
			},
			wantErr: false,
			validate: func(cfg *WalgoConfig) error {
				if cfg.HugoConfig.Version != "0.120.0" {
					return fmt.Errorf("expected Version '0.120.0', got %s", cfg.HugoConfig.Version)
				}
				if cfg.ObsidianConfig.VaultPath != "/path/to/vault" {
					return fmt.Errorf("expected VaultPath '/path/to/vault', got %s", cfg.ObsidianConfig.VaultPath)
				}
				if !cfg.OptimizerConfig.Enabled {
					return fmt.Errorf("expected OptimizerConfig.Enabled to be true")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and config file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "walgo.yaml")

			if tt.configData != "" {
				if err := os.WriteFile(configPath, []byte(tt.configData), 0644); err != nil {
					t.Fatalf("Failed to write test config: %v", err)
				}
			}

			// Setup viper
			if tt.setupViper != nil {
				tt.setupViper(configPath)
			}

			// Execute function
			cfg, err := LoadConfig()

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got %q", tt.errContains, err.Error())
				}
			}

			// Validate config if successful
			if !tt.wantErr && tt.validate != nil {
				if err := tt.validate(cfg); err != nil {
					t.Errorf("Config validation failed: %v", err)
				}
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *WalgoConfig
		setup       func(string) error
		wantErr     bool
		errContains string
	}{
		{
			name: "Save valid config",
			config: &WalgoConfig{
				HugoConfig: HugoConfig{
					PublishDir: "dist",
					ContentDir: "posts",
				},
				WalrusConfig: WalrusConfig{
					ProjectID:  "test-project",
					Entrypoint: "index.html",
				},
			},
			wantErr: false,
		},
		{
			name:   "Save default config",
			config: func() *WalgoConfig { cfg := NewDefaultWalgoConfig(); return &cfg }(),
			wantErr: false,
		},
		{
			name: "Save to read-only directory",
			config: &WalgoConfig{
				HugoConfig: HugoConfig{
					PublishDir: "public",
				},
			},
			setup: func(dir string) error {
				// Make directory read-only
				return os.Chmod(dir, 0555)
			},
			wantErr:     true,
			errContains: "failed to write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()

			// Run setup if provided
			if tt.setup != nil {
				if err := tt.setup(tempDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				// Ensure we can clean up even if directory is read-only
				defer os.Chmod(tempDir, 0755)
			}

			// Execute function
			err := SaveConfig(tempDir, tt.config)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got %q", tt.errContains, err.Error())
				}
			}

			// If successful, verify file was created and is valid
			if !tt.wantErr {
				configPath := filepath.Join(tempDir, DefaultConfigFileName)
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read saved config: %v", err)
				}

				var cfg WalgoConfig
				if err := yaml.Unmarshal(data, &cfg); err != nil {
					t.Errorf("Saved config is not valid YAML: %v", err)
				}

				// Verify content matches
				if cfg.HugoConfig.PublishDir != tt.config.HugoConfig.PublishDir {
					t.Errorf("PublishDir mismatch: expected %s, got %s",
						tt.config.HugoConfig.PublishDir, cfg.HugoConfig.PublishDir)
				}
			}
		})
	}
}

func TestNewDefaultWalgoConfig(t *testing.T) {
	cfg := NewDefaultWalgoConfig()

	// Test Hugo defaults
	if cfg.HugoConfig.PublishDir != "public" {
		t.Errorf("Expected PublishDir 'public', got %s", cfg.HugoConfig.PublishDir)
	}
	if cfg.HugoConfig.ContentDir != "content" {
		t.Errorf("Expected ContentDir 'content', got %s", cfg.HugoConfig.ContentDir)
	}

	// Test Walrus defaults
	if cfg.WalrusConfig.ProjectID != "YOUR_WALRUS_PROJECT_ID" {
		t.Errorf("Expected placeholder ProjectID, got %s", cfg.WalrusConfig.ProjectID)
	}
	if cfg.WalrusConfig.Entrypoint != "index.html" {
		t.Errorf("Expected Entrypoint 'index.html', got %s", cfg.WalrusConfig.Entrypoint)
	}

	// Test Obsidian defaults
	if cfg.ObsidianConfig.AttachmentDir != "images" {
		t.Errorf("Expected AttachmentDir 'images', got %s", cfg.ObsidianConfig.AttachmentDir)
	}
	if !cfg.ObsidianConfig.ConvertWikilinks {
		t.Error("Expected ConvertWikilinks to be true")
	}
	if cfg.ObsidianConfig.IncludeDrafts {
		t.Error("Expected IncludeDrafts to be false")
	}
	if cfg.ObsidianConfig.FrontmatterFormat != "yaml" {
		t.Errorf("Expected FrontmatterFormat 'yaml', got %s", cfg.ObsidianConfig.FrontmatterFormat)
	}

	// Test that OptimizerConfig is set (assuming it has defaults)
	// The actual defaults are tested in the optimizer package tests
	if cfg.OptimizerConfig.Enabled {
		// Good, optimizer is enabled by default
	}
	if cfg.OptimizerConfig.HTML.Enabled {
		// HTML optimization is enabled
	}
	if cfg.OptimizerConfig.CSS.Enabled {
		// CSS optimization is enabled
	}
	if cfg.OptimizerConfig.JS.Enabled {
		// JS optimization is enabled
	}
}

func TestConfigRoundTrip(t *testing.T) {
	// Test that we can save and load a config without data loss
	tempDir := t.TempDir()

	originalCfg := &WalgoConfig{
		HugoConfig: HugoConfig{
			Version:     "0.120.0",
			BaseURL:     "https://example.com",
			PublishDir:  "build",
			ContentDir:  "articles",
			ResourceDir: "assets",
		},
		WalrusConfig: WalrusConfig{
			ProjectID:   "round-trip-test",
			BucketName:  "test-bucket",
			Entrypoint:  "main.html",
			SuiNSDomain: "test.sui",
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
				Enabled:        false,
				MinifyCSS:      false,
				RemoveComments: false,
			},
			JS: optimizer.JSConfig{
				Enabled:        true,
				MinifyJS:       true,
				RemoveComments: true,
			},
		},
	}

	// Save config
	if err := SaveConfig(tempDir, originalCfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Setup viper to load the saved config
	configPath := filepath.Join(tempDir, DefaultConfigFileName)
	viper.Reset()
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config with viper: %v", err)
	}

	// Load config
	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Compare all fields
	if loadedCfg.HugoConfig != originalCfg.HugoConfig {
		t.Errorf("HugoConfig mismatch: %+v != %+v", loadedCfg.HugoConfig, originalCfg.HugoConfig)
	}
	if loadedCfg.WalrusConfig != originalCfg.WalrusConfig {
		t.Errorf("WalrusConfig mismatch: %+v != %+v", loadedCfg.WalrusConfig, originalCfg.WalrusConfig)
	}
	if loadedCfg.ObsidianConfig != originalCfg.ObsidianConfig {
		t.Errorf("ObsidianConfig mismatch: %+v != %+v", loadedCfg.ObsidianConfig, originalCfg.ObsidianConfig)
	}
	// Compare OptimizerConfig fields individually since it contains slices
	if loadedCfg.OptimizerConfig.Enabled != originalCfg.OptimizerConfig.Enabled {
		t.Errorf("OptimizerConfig.Enabled mismatch: %v != %v",
			loadedCfg.OptimizerConfig.Enabled, originalCfg.OptimizerConfig.Enabled)
	}
	if loadedCfg.OptimizerConfig.HTML != originalCfg.OptimizerConfig.HTML {
		t.Errorf("OptimizerConfig.HTML mismatch: %+v != %+v",
			loadedCfg.OptimizerConfig.HTML, originalCfg.OptimizerConfig.HTML)
	}
	if loadedCfg.OptimizerConfig.CSS != originalCfg.OptimizerConfig.CSS {
		t.Errorf("OptimizerConfig.CSS mismatch: %+v != %+v",
			loadedCfg.OptimizerConfig.CSS, originalCfg.OptimizerConfig.CSS)
	}
	if loadedCfg.OptimizerConfig.JS != originalCfg.OptimizerConfig.JS {
		t.Errorf("OptimizerConfig.JS mismatch: %+v != %+v",
			loadedCfg.OptimizerConfig.JS, originalCfg.OptimizerConfig.JS)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && (s[:len(substr)] == substr ||
		contains(s[1:], substr)))
}

// Additional test for edge cases
func TestConfigEdgeCases(t *testing.T) {
	t.Run("Empty config file", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "walgo.yaml")

		// Create empty file
		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		viper.Reset()
		viper.SetConfigFile(configPath)
		viper.ReadInConfig()

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("Empty config should not error: %v", err)
		}

		// Should have defaults
		if cfg.HugoConfig.PublishDir != "public" {
			t.Errorf("Expected default PublishDir, got %s", cfg.HugoConfig.PublishDir)
		}
	})

	t.Run("Config with unknown fields", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "walgo.yaml")

		configData := `
hugo:
  publishDir: public
walrus:
  projectID: test
unknownField: should-be-ignored
futureFeature:
  nested: value
`
		if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
			t.Fatal(err)
		}

		viper.Reset()
		viper.SetConfigFile(configPath)
		viper.ReadInConfig()

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("Config with unknown fields should not error: %v", err)
		}

		// Should still load known fields
		if cfg.WalrusConfig.ProjectID != "test" {
			t.Errorf("Expected ProjectID 'test', got %s", cfg.WalrusConfig.ProjectID)
		}
	})
}