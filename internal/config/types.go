package config

import "walgo/internal/optimizer"

// WalgoConfig is the top-level configuration for Walgo.
// It will be stored in walgo.yaml in the site root.
type WalgoConfig struct {
	HugoConfig      HugoConfig                `mapstructure:"hugo" yaml:"hugo"`
	WalrusConfig    WalrusConfig              `mapstructure:"walrus" yaml:"walrus"`
	ObsidianConfig  ObsidianConfig            `mapstructure:"obsidian" yaml:"obsidian,omitempty"`
	OptimizerConfig optimizer.OptimizerConfig `mapstructure:"optimizer" yaml:"optimizer,omitempty"`
	CompressConfig  CompressConfig            `mapstructure:"compress" yaml:"compress,omitempty"`
	CacheConfig     CacheConfig               `mapstructure:"cache" yaml:"cache,omitempty"`
	// Future: Additional integrations
}

// HugoConfig holds Hugo-specific settings relevant to Walgo.
// Note: Most Hugo configurations are in hugo.toml (or config.toml/yaml/json).
// This section is for settings Walgo might need to override or know about.
type HugoConfig struct {
	Version     string `mapstructure:"version" yaml:"version,omitempty"`         // Hugo version to target/ensure
	BaseURL     string `mapstructure:"baseURL" yaml:"baseURL,omitempty"`         // Overrides Hugo's baseURL for specific deployments
	PublishDir  string `mapstructure:"publishDir" yaml:"publishDir,omitempty"`   // Default: "public"
	ContentDir  string `mapstructure:"contentDir" yaml:"contentDir,omitempty"`   // Default: "content"
	ResourceDir string `mapstructure:"resourceDir" yaml:"resourceDir,omitempty"` // Default: "resources"
}

// WalrusConfig holds settings for deploying to Walrus Sites.
type WalrusConfig struct {
	ProjectID   string `mapstructure:"projectID" yaml:"projectID"`               // Walrus Project ID or name
	BucketName  string `mapstructure:"bucketName" yaml:"bucketName,omitempty"`   // Optional: specific bucket if not default
	Entrypoint  string `mapstructure:"entrypoint" yaml:"entrypoint,omitempty"`   // Default: "index.html"
	SuiNSDomain string `mapstructure:"suinsDomain" yaml:"suinsDomain,omitempty"` // SuiNS domain to associate
	// Future: API keys, access tokens (consider secure storage/env vars)
}

// ObsidianConfig holds settings for importing from Obsidian vaults.
type ObsidianConfig struct {
	VaultPath         string `mapstructure:"vaultPath" yaml:"vaultPath,omitempty"`         // Default Obsidian vault path
	AttachmentDir     string `mapstructure:"attachmentDir" yaml:"attachmentDir,omitempty"` // Where to put attachments (relative to static/)
	ConvertWikilinks  bool   `mapstructure:"convertWikilinks" yaml:"convertWikilinks"`     // Convert [[wikilinks]] to [markdown](links)
	IncludeDrafts     bool   `mapstructure:"includeDrafts" yaml:"includeDrafts"`           // Include files marked as drafts
	FrontmatterFormat string `mapstructure:"frontmatterFormat" yaml:"frontmatterFormat"`   // yaml, toml, json
}

// CompressConfig holds settings for Brotli compression
type CompressConfig struct {
	Enabled     bool `mapstructure:"enabled" yaml:"enabled"`           // Enable compression
	Level       int  `mapstructure:"level" yaml:"level,omitempty"`     // Brotli level 0-11, default: 6
	GenerateWSResources bool `mapstructure:"generateWSResources" yaml:"generateWSResources"` // Generate ws-resources.json
}

// CacheConfig holds settings for caching and cache-control headers
type CacheConfig struct {
	Enabled         bool `mapstructure:"enabled" yaml:"enabled"`                   // Enable cache-control headers
	ImmutableMaxAge int  `mapstructure:"immutableMaxAge" yaml:"immutableMaxAge,omitempty"` // Max-age for immutable assets (default: 31536000)
	MutableMaxAge   int  `mapstructure:"mutableMaxAge" yaml:"mutableMaxAge,omitempty"`     // Max-age for HTML (default: 300)
}

// NewDefaultWalgoConfig creates a WalgoConfig with sensible defaults.
func NewDefaultWalgoConfig() WalgoConfig {
	return WalgoConfig{
		HugoConfig: HugoConfig{
			PublishDir: "public",
			ContentDir: "content",
		},
		WalrusConfig: WalrusConfig{
			ProjectID:  "YOUR_WALRUS_PROJECT_ID", // User needs to fill this
			Entrypoint: "index.html",
		},
		ObsidianConfig: ObsidianConfig{
			AttachmentDir:     "images",
			ConvertWikilinks:  true,
			IncludeDrafts:     false,
			FrontmatterFormat: "yaml",
		},
		OptimizerConfig: optimizer.NewDefaultOptimizerConfig(),
		CompressConfig: CompressConfig{
			Enabled:             true,
			Level:               6, // Balanced compression
			GenerateWSResources: true,
		},
		CacheConfig: CacheConfig{
			Enabled:         true,
			ImmutableMaxAge: 31536000, // 1 year
			MutableMaxAge:   300,       // 5 minutes
		},
	}
}
