// Package sui provides centralized Sui CLI operations for Walgo.
package sui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/executil"
)

// getSuiPath returns the path to the sui binary
func getSuiPath() (string, error) {
	return deps.LookPath("sui")
}

// runCommand executes a sui command and returns the output
// It filters out warning messages from stderr
func runCommand(args ...string) (string, error) {
	suiPath, err := getSuiPath()
	if err != nil {
		return "", fmt.Errorf("sui CLI not found: %w", err)
	}

	cmd := executil.Command(suiPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("sui command failed: %w\nOutput: %s", err, string(output))
	}

	// Filter out warning lines from output
	result := filterWarnings(string(output))
	return strings.TrimSpace(result), nil
}

// filterWarnings removes warning lines from command output
func filterWarnings(output string) string {
	lines := strings.Split(output, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip warning lines
		if strings.HasPrefix(strings.ToLower(trimmed), "[warning]") ||
			strings.HasPrefix(strings.ToLower(trimmed), "warning:") ||
			strings.HasPrefix(strings.ToLower(trimmed), "warn:") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

// runCommandJSON executes a sui command with --json flag and returns only the JSON output
// It strips any warning messages or non-JSON content that may precede the JSON
func runCommandJSON(args ...string) (string, error) {
	args = append(args, "--json")
	output, err := runCommand(args...)
	if err != nil {
		return output, err
	}

	// Extract JSON from the output - find first { or [ character
	return extractJSON(output), nil
}

// extractJSON extracts valid JSON from output that may contain warnings or other text
func extractJSON(output string) string {
	output = strings.TrimSpace(output)

	// Find the start of JSON (first { or [)
	startObj := strings.Index(output, "{")
	startArr := strings.Index(output, "[")

	start := -1
	if startObj >= 0 && startArr >= 0 {
		if startObj < startArr {
			start = startObj
		} else {
			start = startArr
		}
	} else if startObj >= 0 {
		start = startObj
	} else if startArr >= 0 {
		start = startArr
	}

	if start == -1 {
		return output // No JSON found, return original
	}

	// Find the matching end bracket
	jsonStr := output[start:]

	// For simple cases, just return from the start of JSON to the end
	// The JSON parser will handle any trailing content
	return strings.TrimSpace(jsonStr)
}

// GetActiveEnv returns the current active Sui environment (testnet/mainnet)
func GetActiveEnv() (string, error) {
	return runCommand("client", "active-env")
}

// GetActiveAddress returns the current active wallet address
func GetActiveAddress() (string, error) {
	return runCommand("client", "active-address")
}

// SwitchEnv switches to the specified network environment
func SwitchEnv(network string) error {
	_, err := runCommand("client", "switch", "--env", network)
	return err
}

// SwitchAddress switches to the specified wallet address
func SwitchAddress(address string) error {
	_, err := runCommand("client", "switch", "--address", address)
	return err
}

// AddressesResponse represents the JSON output from `sui client addresses --json`
type AddressesResponse struct {
	ActiveAddress string     `json:"activeAddress"`
	Addresses     [][]string `json:"addresses"` // [[alias, address], ...]
}

// GetAddresses returns all wallet addresses
func GetAddresses() (*AddressesResponse, error) {
	output, err := runCommandJSON("client", "addresses")
	if err != nil {
		return nil, err
	}

	var result AddressesResponse
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse addresses: %w", err)
	}

	return &result, nil
}

// GetAddressList returns just the list of addresses (without aliases)
func GetAddressList() ([]string, error) {
	resp, err := GetAddresses()
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, pair := range resp.Addresses {
		if len(pair) >= 2 {
			addresses = append(addresses, pair[1])
		}
	}
	return addresses, nil
}

// BalanceInfo represents parsed balance information
type BalanceInfo struct {
	SUI float64
	WAL float64
}

// GetBalance returns the balance for the active address
func GetBalance() (*BalanceInfo, error) {
	output, err := runCommandJSON("client", "balance")
	if err != nil {
		return nil, err
	}

	return parseBalanceJSON(output)
}

// parseBalanceJSON parses the JSON output from `sui client balance --json`
func parseBalanceJSON(jsonOutput string) (*BalanceInfo, error) {
	var result interface{}
	if err := json.Unmarshal([]byte(jsonOutput), &result); err != nil {
		return nil, fmt.Errorf("failed to parse balance JSON: %w", err)
	}

	// The response is [[token_entries...], boolean]
	outerArray, ok := result.([]interface{})
	if !ok || len(outerArray) == 0 {
		return nil, fmt.Errorf("unexpected JSON format: expected outer array")
	}

	// First element contains the token entries
	entries, ok := outerArray[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected JSON format: expected array of token entries")
	}

	info := &BalanceInfo{}

	for _, entry := range entries {
		entryArray, ok := entry.([]interface{})
		if !ok || len(entryArray) < 2 {
			continue
		}

		// First element is token info
		tokenInfoMap, ok := entryArray[0].(map[string]interface{})
		if !ok {
			continue
		}

		symbol, _ := tokenInfoMap["symbol"].(string)
		decimals := 9 // Default decimals for SUI and WAL

		if d, ok := tokenInfoMap["decimals"].(float64); ok {
			decimals = int(d)
		}

		// Second element is array of coin balances
		coinsArray, ok := entryArray[1].([]interface{})
		if !ok {
			continue
		}

		// Sum up all coins for this token
		var totalBalance float64
		for _, coin := range coinsArray {
			coinMap, ok := coin.(map[string]interface{})
			if !ok {
				continue
			}

			balanceStr, _ := coinMap["balance"].(string)
			var balance float64
			if _, err := fmt.Sscanf(balanceStr, "%f", &balance); err != nil {
				// Skip this balance if parsing fails
				continue
			}
			totalBalance += balance
		}

		// Convert to token units (divide by 10^decimals)
		tokenBalance := totalBalance / pow10(decimals)

		switch symbol {
		case "SUI":
			info.SUI = tokenBalance
		case "WAL":
			info.WAL = tokenBalance
		}
	}

	return info, nil
}

// pow10 calculates 10^n
func pow10(n int) float64 {
	result := float64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

// NewAddressResponse represents the JSON output from `sui client new-address --json`
type NewAddressResponse struct {
	Alias          string `json:"alias"`
	Address        string `json:"address"`
	KeyScheme      string `json:"keyScheme"`
	RecoveryPhrase string `json:"recoveryPhrase"`
}

// CreateAddressResult contains the result of creating a new address
type CreateAddressResult struct {
	Address        string
	Alias          string
	KeyScheme      string
	RecoveryPhrase string
}

// CreateAddressWithDetails creates a new address and returns full details including recovery phrase
func CreateAddressWithDetails(keyScheme string, alias string) (*CreateAddressResult, error) {
	if keyScheme == "" {
		keyScheme = "ed25519"
	}

	// Build command arguments
	// sui client new-address <KEY_SCHEME> [ALIAS]
	args := []string{"client", "new-address", keyScheme}
	if alias != "" {
		args = append(args, alias)
	}

	output, err := runCommandJSON(args...)
	if err != nil {
		return nil, err
	}

	var result NewAddressResponse
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		// Fallback to parsing text output for older CLI versions
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "0x") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "0x") {
						return &CreateAddressResult{Address: part}, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("failed to parse new address from output")
	}

	return &CreateAddressResult{
		Address:        result.Address,
		Alias:          result.Alias,
		KeyScheme:      result.KeyScheme,
		RecoveryPhrase: result.RecoveryPhrase,
	}, nil
}

// GetVersion returns the Sui CLI version
func GetVersion() (string, error) {
	return runCommand("--version")
}

// ImportMethod represents the method for importing an address
type ImportMethod string

const (
	ImportFromMnemonic   ImportMethod = "mnemonic"
	ImportFromPrivateKey ImportMethod = "key"
)

// ImportAddressWithInput imports an address using provided mnemonic or private key
// This is suitable for GUI applications where user input is collected separately
// Correct format: sui keytool import "<input>" <keyScheme> [derivation_path]
func ImportAddressWithInput(method ImportMethod, keyScheme, input string) (string, error) {
	suiPath, err := getSuiPath()
	if err != nil {
		return "", fmt.Errorf("sui CLI not found: %w", err)
	}

	if keyScheme == "" {
		keyScheme = "ed25519"
	}

	// Correct command format: sui keytool import "<input>" <keyScheme>
	// Input is passed as argument, not via stdin
	cmd := executil.Command(suiPath, "keytool", "import", input, keyScheme, "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("import failed: %w - %s", err, string(output))
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		// Fallback to text parsing if JSON fails
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "0x") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "0x") {
						return part, nil
					}
				}
			}
		}
		return "", fmt.Errorf("could not parse output: %v", err)
	}

	// Extract address from JSON
	if addr, ok := result["suiAddress"].(string); ok {
		return addr, nil
	}

	return "", fmt.Errorf("could not extract address from JSON output")
}

// GetExplorerURL returns the explorer URL for a given object ID on the specified network
func GetExplorerURL(network, objectType, objectID string) string {
	switch objectType {
	case "suiscan":
		return fmt.Sprintf("https://suiscan.xyz/%s/object/%s", network, objectID)
	case "suivision":
		// Suivision uses /package/ path for Walrus sites
		baseDomain := "suivision.xyz"
		if network == "testnet" {
			baseDomain = "testnet.suivision.xyz"
		}
		return fmt.Sprintf("https://%s/package/%s", baseDomain, objectID)
	default:
		return ""
	}
}

// GetSuiscanURL returns the Suiscan URL for the given object ID on the specified network
func GetSuiscanURL(network, objectID string) string {
	return GetExplorerURL(network, "suiscan", objectID)
}

// GetSuivisionURL returns the Suivision URL for the given object ID on the specified network
func GetSuivisionURL(network, objectID string) string {
	return GetExplorerURL(network, "suivision", objectID)
}
