package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"walgo/internal/config"
	"walgo/internal/deployer"
	sb "walgo/internal/deployer/sitebuilder"
	"walgo/internal/hugo"

	"github.com/spf13/viper"
)

// CreateSite initializes a new Hugo site with Walrus configuration
func CreateSite(parentDir string, name string) error {
	sitePath := filepath.Join(parentDir, name)

	// 1. Create site directory
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		return fmt.Errorf("error creating site directory: %w", err)
	}

	// 2. Initialize Hugo site
	if err := hugo.InitializeSite(sitePath); err != nil {
		return fmt.Errorf("error initializing Hugo site: %w", err)
	}

	// 3. Create Walrus configuration
	if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
		return fmt.Errorf("error creating default config: %w", err)
	}

	return nil
}

// BuildSite builds the Hugo site at the given path
func BuildSite(sitePath string) error {
	// Change to site directory to ensure Hugo finds everything
	if err := os.Chdir(sitePath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Load config to ensure it exists and is valid
	viper.Reset()
	viper.SetConfigFile(filepath.Join(sitePath, "walgo.yaml"))
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read walgo.yaml: %w", err)
	}

	if err := hugo.BuildSite(sitePath); err != nil {
		return fmt.Errorf("hugo build failed: %w", err)
	}

	return nil
}

// DeployResult holds the result of a deployment
type DeployResult struct {
	Success  bool   `json:"success"`
	ObjectID string `json:"objectId"`
	Error    string `json:"error"`
}

// DeploySite deploys the site to Walrus
func DeploySite(sitePath string, epochs int) DeployResult {
	// Change to site directory
	if err := os.Chdir(sitePath); err != nil {
		return DeployResult{Error: fmt.Sprintf("failed to change directory: %v", err)}
	}

	// Load config
	viper.Reset()
	viper.SetConfigFile(filepath.Join(sitePath, "walgo.yaml"))
	if err := viper.ReadInConfig(); err != nil {
		return DeployResult{Error: fmt.Sprintf("failed to read walgo.yaml: %v", err)}
	}

	walgoCfg, err := config.LoadConfig()
	if err != nil {
		return DeployResult{Error: fmt.Sprintf("failed to load config: %v", err)}
	}

	// Check if public directory exists
	publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
	if _, err := os.Stat(publishDir); os.IsNotExist(err) {
		return DeployResult{Error: fmt.Sprintf("publish directory not found: %s. Please build first.", publishDir)}
	}

	// Deploy
	d := sb.New()
	// Use a reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 300000000000) // 5 minutes (approx)
	defer cancel()

	result, err := d.Deploy(ctx, publishDir, deployer.DeployOptions{
		Epochs:    epochs,
		Verbose:   true,
		WalrusCfg: walgoCfg.WalrusConfig,
	})

	if err != nil {
		return DeployResult{Error: fmt.Sprintf("deployment failed: %v", err)}
	}

	return DeployResult{
		Success:  result.Success,
		ObjectID: result.ObjectID,
	}
}
