package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// walrusSitesConfig represents the sites-config.yaml structure.
type walrusSitesConfig struct {
	Contexts map[string]struct {
		Package string `yaml:"package"`
		General struct {
			RPCURL       string `yaml:"rpc_url"`
			Wallet       string `yaml:"wallet"`
			WalrusBinary string `yaml:"walrus_binary"`
			WalrusConfig string `yaml:"walrus_config"`
			GasBudget    int    `yaml:"gas_budget"`
		} `yaml:"general"`
	} `yaml:"contexts"`
	DefaultContext string `yaml:"default_context"`
}

// wireWalrusBinary updates sites-config.yaml to point to the installed walrus binary.
// Only sets the path if currently empty to avoid overwriting custom configurations.
func wireWalrusBinary(binDir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.New("failed to get home directory")
	}

	scPath := filepath.Join(home, ".config", "walrus", "sites-config.yaml")
	data, err := os.ReadFile(scPath) // #nosec G304 - known config path
	if err != nil {
		return errors.New("sites-config.yaml not found; run walgo setup first")
	}

	var sc walrusSitesConfig
	if err := yaml.Unmarshal(data, &sc); err != nil {
		return fmt.Errorf("failed to parse sites-config.yaml: %w", err)
	}

	walrusPath := filepath.Join(binDir, "walrus")
	for k, ctx := range sc.Contexts {
		if strings.TrimSpace(ctx.General.WalrusBinary) == "" {
			ctx.General.WalrusBinary = walrusPath
			sc.Contexts[k] = ctx
		}
	}

	out, err := yaml.Marshal(&sc)
	if err != nil {
		return err
	}

	// #nosec G306 - config file needs to be readable
	return os.WriteFile(scPath, out, 0o644)
}
