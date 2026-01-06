package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigFileName = "walgo.yaml"
)

// CreateDefaultWalgoConfig creates a default walgo.yaml file in the specified site path.
func CreateDefaultWalgoConfig(sitePath string) error {
	cfg := NewDefaultWalgoConfig()
	configFilePath := filepath.Join(sitePath, DefaultConfigFileName)

	if _, err := os.Stat(configFilePath); err == nil {
		return fmt.Errorf("configuration file %s already exists", configFilePath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking for config file %s: %w", configFilePath, err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal default config to YAML: %w", err)
	}

	// #nosec G306 - config file needs to be readable
	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write default config file %s: %w", configFilePath, err)
	}

	fmt.Printf("Created default configuration file: %s\n", configFilePath)
	return nil
}

// LoadConfig loads the Walgo configuration from walgo.yaml.
func LoadConfig() (*WalgoConfig, error) {
	if viper.ConfigFileUsed() == "" {
		return nil, fmt.Errorf("walgo.yaml configuration file not found or failed to load. Ensure you are in a Walgo project directory (where walgo.yaml or .walgo.yaml exists in CWD/home), or use the --config flag to specify the path. You can create a default config with 'walgo init'")
	}

	var cfg WalgoConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling configuration from %s: %w. Please check the file format and structure", viper.ConfigFileUsed(), err)
	}

	if cfg.HugoConfig.PublishDir == "" {
		cfg.HugoConfig.PublishDir = "public"
	}
	if cfg.HugoConfig.ContentDir == "" {
		cfg.HugoConfig.ContentDir = "content"
	}
	if cfg.HugoConfig.ResourceDir == "" {
		cfg.HugoConfig.ResourceDir = "resources" // Hugo's default
	}
	if cfg.WalrusConfig.Entrypoint == "" {
		cfg.WalrusConfig.Entrypoint = "index.html"
	}

	return &cfg, nil
}

// LoadConfigFrom loads the Walgo configuration from a specific directory.
// This is useful when you need to load config from a path different from CWD.
func LoadConfigFrom(sitePath string) (*WalgoConfig, error) {
	configPath := filepath.Join(sitePath, DefaultConfigFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("walgo.yaml not found in %s", sitePath)
	}

	// Read and parse the config file directly
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var cfg WalgoConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Apply defaults
	if cfg.HugoConfig.PublishDir == "" {
		cfg.HugoConfig.PublishDir = "public"
	}
	if cfg.HugoConfig.ContentDir == "" {
		cfg.HugoConfig.ContentDir = "content"
	}
	if cfg.HugoConfig.ResourceDir == "" {
		cfg.HugoConfig.ResourceDir = "resources"
	}
	if cfg.WalrusConfig.Entrypoint == "" {
		cfg.WalrusConfig.Entrypoint = "index.html"
	}

	return &cfg, nil
}

// SaveConfig saves the given WalgoConfig to walgo.yaml in the specified directory.
func SaveConfig(configDir string, cfg *WalgoConfig) error {
	configFilePath := filepath.Join(configDir, DefaultConfigFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// #nosec G306 - config file needs to be readable
	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configFilePath, err)
	}
	return nil
}
