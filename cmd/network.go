package cmd

import (
	"fmt"
	"strings"

	"github.com/ganbitlabs/walgo/internal/sui"
)

// normalizeNetwork reduces a Sui environment name to "mainnet" or "testnet".
// Anything unrecognized is reported as-is so callers can surface it.
func normalizeNetwork(env string) string {
	env = strings.ToLower(strings.TrimSpace(env))
	switch {
	case strings.Contains(env, "mainnet"):
		return "mainnet"
	case strings.Contains(env, "testnet"):
		return "testnet"
	default:
		return env
	}
}

// resolveDeployNetwork returns the network a deployment will actually run against.
//
// site-builder and walrus follow the active Sui environment, not walgo.yaml, so a
// mismatch between the two silently deploys to the wrong network (and spends real
// funds when that network is mainnet). Rather than guess, refuse and explain.
func resolveDeployNetwork(configuredNetwork string) (string, error) {
	activeEnv, err := sui.GetActiveEnv()
	if err != nil {
		return "", fmt.Errorf("could not determine the active Sui network: %w\n\n"+
			"  Check your Sui CLI setup:  sui client envs", err)
	}

	active := normalizeNetwork(activeEnv)
	if active != "mainnet" && active != "testnet" {
		return "", fmt.Errorf("unsupported active Sui network %q\n\n"+
			"  Switch to a supported network:  sui client switch --env mainnet", activeEnv)
	}

	configured := normalizeNetwork(configuredNetwork)
	if configured == "" || configured == active {
		return active, nil
	}

	return "", fmt.Errorf("network mismatch: walgo.yaml says %s but the active Sui environment is %s\n\n"+
		"  Deployments follow the Sui environment, so this would deploy to %s.\n\n"+
		"  To fix, either:\n"+
		"    1. Switch Sui:      sui client switch --env %s\n"+
		"    2. Or edit walgo.yaml:  walrus.network: %s",
		configured, active, active, configured, active)
}
