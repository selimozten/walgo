package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const ( // TODO: make walgo.yaml configurable
	DefaultConfigFileName = "walgo.yaml"
)

// CreateDefaultWalgoConfig creates a default walgo.yaml file in the specified site path.
func CreateDefaultWalgoConfig(sitePath string) error {
	cfg := NewDefaultWalgoConfig()
	configFilePath := filepath.Join(sitePath, DefaultConfigFileName)

	// Check if file already exists
	if _, err := os.Stat(configFilePath); err == nil {
		return fmt.Errorf("configuration file %s already exists", configFilePath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking for config file %s: %w", configFilePath, err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal default config to YAML: %w", err)
	}

	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write default config file %s: %w", configFilePath, err)
	}

	fmt.Printf("Created default configuration file: %s\n", configFilePath)
	return nil
}

// LoadConfig loads the Walgo configuration.
// It relies on Viper being pre-configured (e.g., by initConfig in cmd/root.go)
// and initConfig having attempted to read the configuration.
func LoadConfig() (*WalgoConfig, error) {
	// Viper is expected to be initialized by the caller (e.g., cmd.initConfig)
	// which handles the --config flag and default search paths, and calls viper.ReadInConfig().

	// Check if a configuration file was successfully read and is being used.
	if viper.ConfigFileUsed() == "" {
		// This indicates that initConfig did not successfully load/read a config file.
		// This could be because the file wasn't found, or because cfgFile was specified and couldn't be read (error already printed by initConfig).
		return nil, fmt.Errorf("walgo.yaml configuration file not found or failed to load. Ensure you are in a Walgo project directory (where walgo.yaml or .walgo.yaml exists in CWD/home), or use the --config flag to specify the path. You can create a default config with 'walgo init'")
	}

	var cfg WalgoConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling configuration from %s: %w. Please check the file format and structure", viper.ConfigFileUsed(), err)
	}

	// Apply defaults for fields that might be empty in the config file but have defaults
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

// SaveConfig saves the given WalgoConfig to walgo.yaml in the specified directory.
// Note: configDir here implies that SaveConfig needs to know where to save,
// which might be different from where LoadConfig loaded from if --config was used.
// For Phase 1, SaveConfig is not actively used by commands.
func SaveConfig(configDir string, cfg *WalgoConfig) error {
	configFilePath := filepath.Join(configDir, DefaultConfigFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configFilePath, err)
	}
	return nil
}
