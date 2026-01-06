package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// UpdateWalgoYAMLProjectID updates the projectID field in walgo.yaml
// This function preserves the YAML structure and comments while updating specific field
func UpdateWalgoYAMLProjectID(sitePath, objectID string) error {
	// Read existing walgo.yaml
	data, err := os.ReadFile(filepath.Join(sitePath, "walgo.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read walgo.yaml: %w", err)
	}

	// Parse YAML as a generic map to preserve comments and structure
	var yamlMap map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlMap); err != nil {
		return fmt.Errorf("failed to parse walgo.yaml: %w", err)
	}

	// Navigate to walrus.projectID
	walrusMap, ok := yamlMap["walrus"].(map[string]interface{})
	if !ok {
		walrusMap = make(map[string]interface{})
		yamlMap["walrus"] = walrusMap
	}

	// Update projectID
	walrusMap["projectID"] = objectID

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(yamlMap)
	if err != nil {
		return fmt.Errorf("failed to marshal walgo.yaml: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(filepath.Join(sitePath, "walgo.yaml"), updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write walgo.yaml: %w", err)
	}

	return nil
}
