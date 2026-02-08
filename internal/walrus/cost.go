// Package walrus provides cost estimation for Walrus storage operations.
package walrus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/sui"
)

// Default RPC endpoints for Sui networks
const (
	SuiTestnetRPC = "https://fullnode.testnet.sui.io:443"
	SuiMainnetRPC = "https://fullnode.mainnet.sui.io:443"
)

// GetWalrusContext returns the walrus context based on the active Sui environment
// Returns "mainnet" or "testnet" to match the Sui network
func GetWalrusContext() string {
	env, err := sui.GetActiveEnv()
	if err != nil {
		return "testnet" // Default to testnet
	}
	env = strings.ToLower(strings.TrimSpace(env))
	if strings.Contains(env, "mainnet") {
		return "mainnet"
	}
	return "testnet"
}

// SuiRPCRequest represents a JSON-RPC 2.0 request to Sui
type SuiRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// SuiRPCResponse represents a JSON-RPC 2.0 response from Sui
type SuiRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *SuiRPCError    `json:"error,omitempty"`
}

// SuiRPCError represents an error from Sui RPC
type SuiRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// WalrusInfoJSON represents the JSON response from 'walrus info --json'
type WalrusInfoJSON struct {
	EpochInfo struct {
		CurrentEpoch        int `json:"currentEpoch"`
		StartOfCurrentEpoch struct {
			DateTime string `json:"DateTime"`
		} `json:"startOfCurrentEpoch"`
		EpochDuration struct {
			Secs  int `json:"secs"`
			Nanos int `json:"nanos"`
		} `json:"epochDuration"`
		MaxEpochsAhead int `json:"maxEpochsAhead"`
	} `json:"epochInfo"`
	StorageInfo struct {
		NShards int `json:"nShards"`
		NNodes  int `json:"nNodes"`
	} `json:"storageInfo"`
	SizeInfo struct {
		StorageUnitSize int64 `json:"storageUnitSize"` // 1 MiB = 1048576
		MaxBlobSize     int64 `json:"maxBlobSize"`
	} `json:"sizeInfo"`
	PriceInfo struct {
		StoragePricePerUnitSize    uint64 `json:"storagePricePerUnitSize"` // FROST per storage unit per epoch
		WritePricePerUnitSize      uint64 `json:"writePricePerUnitSize"`   // FROST per storage unit (one-time)
		EncodingDependentPriceInfo []struct {
			MarginalSize  int64  `json:"marginalSize"`  // 1 MiB
			MetadataPrice uint64 `json:"metadataPrice"` // Fixed metadata cost in FROST
			MarginalPrice uint64 `json:"marginalPrice"` // Per MiB cost for unencoded data
			EncodingType  string `json:"encodingType"`  // "RS2" for Reed-Solomon
			ExampleBlobs  []struct {
				UnencodedSize int64  `json:"unencodedSize"`
				EncodedSize   int64  `json:"encodedSize"`
				Price         uint64 `json:"price"`
				EncodingType  string `json:"encodingType"`
			} `json:"exampleBlobs"`
		} `json:"encodingDependentPriceInfo"`
	} `json:"priceInfo"`
}

// StorageInfo contains parsed walrus info storage parameters
type StorageInfo struct {
	CurrentEpoch       int     `json:"current_epoch"`
	EpochDuration      int     `json:"epoch_duration_secs"` // Duration in seconds
	StoragePrice       uint64  `json:"storage_price"`       // Price per encoded MiB per epoch in FROST
	WritePrice         uint64  `json:"write_price"`         // Write price per encoded MiB in FROST
	MetadataPrice      uint64  `json:"metadata_price"`      // Fixed metadata cost in FROST
	MarginalPrice      uint64  `json:"marginal_price"`      // Per unencoded MiB cost in FROST
	MaxBlobSize        int64   `json:"max_blob_size"`       // Maximum blob size in bytes
	StorageUnitSize    int64   `json:"storage_unit_size"`   // Storage unit size (1 MiB)
	NumShards          int     `json:"num_shards"`          // Number of storage shards
	MaxEpochsAhead     int     `json:"max_epochs_ahead"`    // Max epochs for storage
	EncodingMultiplier float64 `json:"encoding_multiplier"` // Encoding expansion factor (~5-8x depending on size)
}

// CostBreakdown provides detailed cost breakdown for storage operations
// Separates WAL (storage) and SUI (transaction) costs
type CostBreakdown struct {
	// SUI costs (transaction gas)
	GasUnits   uint64  `json:"gas_units"`    // Total gas units
	GasPrice   uint64  `json:"gas_price"`    // Gas price in MIST
	GasCostSUI float64 `json:"gas_cost_sui"` // Transaction gas cost in SUI

	// WAL costs (storage)
	StorageCostWAL float64 `json:"storage_cost_wal"` // Storage cost in WAL
	WriteCostWAL   float64 `json:"write_cost_wal"`   // Write cost in WAL
	TotalWAL       float64 `json:"total_wal"`        // Total WAL cost

	// Calculated values
	EncodedSize  int64 `json:"encoded_size"`  // Encoded blob size (after Reed-Solomon)
	OriginalSize int64 `json:"original_size"` // Original data size
	FileCount    int   `json:"file_count"`    // Number of files
	Epochs       int   `json:"epochs"`        // Storage duration in epochs

	// Estimates (min/max range)
	MinTotalWAL float64 `json:"min_total_wal"`
	MaxTotalWAL float64 `json:"max_total_wal"`
	MinTotalSUI float64 `json:"min_total_sui"`
	MaxTotalSUI float64 `json:"max_total_sui"`
}

// CostOptions contains parameters for cost estimation
type CostOptions struct {
	SiteSize  int64  // Total site size in bytes
	Epochs    int    // Number of epochs for storage
	FileCount int    // Actual number of files (if known, 0 to estimate)
	RPCURL    string // Sui RPC endpoint for gas price query
	GasPrice  uint64 // Manual gas price override (0 to fetch)
	Network   string // "testnet" or "mainnet"
	WalrusBin string // Path to walrus binary (optional)
}

// GetReferenceGasPrice queries Sui RPC for current reference gas price
// Uses the suix_getReferenceGasPrice method
// Returns gas price in MIST (1 SUI = 1e9 MIST)
func GetReferenceGasPrice(rpcURL string) (uint64, error) {
	if rpcURL == "" {
		rpcURL = SuiTestnetRPC // Default to testnet
	}

	// Create JSON-RPC request
	request := SuiRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "suix_getReferenceGasPrice",
		Params:  []interface{}{},
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make the request
	resp, err := client.Post(rpcURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return 0, fmt.Errorf("RPC request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var rpcResp SuiRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return 0, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	// Parse result (returned as a string number)
	var gasPriceStr string
	if err := json.Unmarshal(rpcResp.Result, &gasPriceStr); err != nil {
		return 0, fmt.Errorf("failed to parse gas price: %w", err)
	}

	gasPrice, err := strconv.ParseUint(gasPriceStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid gas price format: %w", err)
	}

	return gasPrice, nil
}

// GetStorageInfo fetches storage parameters from 'walrus info --json'
// It automatically uses the correct context based on the active Sui environment
func GetStorageInfo(walrusBin string) (*StorageInfo, error) {
	if walrusBin == "" {
		walrusBin = "walrus"
		// Try to find walrus using deps.LookPath
		if path, err := deps.LookPath("walrus"); err == nil {
			walrusBin = path
		}
	}

	// Get the context from Sui active environment
	walrusCtx := GetWalrusContext()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := execCommandContext(ctx, walrusBin, "info", "--json", "--context", walrusCtx)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run walrus info --json --context %s: %w", walrusCtx, err)
	}

	return ParseStorageInfoJSON(output)
}

// stripANSI removes ANSI escape codes (color codes) from byte slice
// ANSI codes follow pattern: ESC [ <params> <letter>
// Example: \x1b[32m (green), \x1b[0m (reset)
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(data []byte) []byte {
	return ansiRegex.ReplaceAll(data, []byte{})
}

// ParseStorageInfoJSON parses the walrus info JSON output
func ParseStorageInfoJSON(output []byte) (*StorageInfo, error) {
	// Strip ANSI color codes that walrus CLI may include
	output = stripANSI(output)

	// The walrus CLI outputs log lines before the JSON, so we need to find the JSON
	// Look for a '{' that starts on a new line (actual JSON object, not log text like "run{command=...}")
	// We look for "\n{" pattern to find the JSON object start
	lines := bytes.Split(output, []byte("\n"))
	var jsonStart int = -1
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 && trimmed[0] == '{' {
			// Found a line that starts with '{', this is likely our JSON
			jsonStart = 0
			for j := 0; j < i; j++ {
				jsonStart += len(lines[j]) + 1 // +1 for newline
			}
			break
		}
	}

	if jsonStart == -1 {
		return nil, fmt.Errorf("no JSON found in walrus info output")
	}

	var walrusInfo WalrusInfoJSON
	if err := json.Unmarshal(output[jsonStart:], &walrusInfo); err != nil {
		return nil, fmt.Errorf("failed to parse walrus info JSON: %w", err)
	}

	info := &StorageInfo{
		CurrentEpoch:       walrusInfo.EpochInfo.CurrentEpoch,
		EpochDuration:      walrusInfo.EpochInfo.EpochDuration.Secs,
		StoragePrice:       walrusInfo.PriceInfo.StoragePricePerUnitSize,
		WritePrice:         walrusInfo.PriceInfo.WritePricePerUnitSize,
		MaxBlobSize:        walrusInfo.SizeInfo.MaxBlobSize,
		StorageUnitSize:    walrusInfo.SizeInfo.StorageUnitSize,
		NumShards:          walrusInfo.StorageInfo.NShards,
		MaxEpochsAhead:     walrusInfo.EpochInfo.MaxEpochsAhead,
		EncodingMultiplier: 5.0, // Default, will be calculated more accurately per-blob
	}

	// Get encoding-specific pricing (RS2 encoding)
	if len(walrusInfo.PriceInfo.EncodingDependentPriceInfo) > 0 {
		encInfo := walrusInfo.PriceInfo.EncodingDependentPriceInfo[0]
		info.MetadataPrice = encInfo.MetadataPrice
		info.MarginalPrice = encInfo.MarginalPrice

		// Calculate more accurate encoding multiplier from example blobs
		if len(encInfo.ExampleBlobs) > 0 {
			// Use the first example to estimate encoding factor
			example := encInfo.ExampleBlobs[0]
			if example.UnencodedSize > 0 {
				info.EncodingMultiplier = float64(example.EncodedSize) / float64(example.UnencodedSize)
			}
		}
	}

	return info, nil
}

// encodingMultiplierForSize returns the estimated Reed-Solomon expansion factor
// based on blob size. Smaller blobs have higher overhead due to fixed metadata.
// Based on walrus info examples:
//   - 16 MiB -> 134 MiB (~8.4x)
//   - 512 MiB -> 2.31 GiB (~4.6x)
//   - 13.6 GiB -> 61.2 GiB (~4.5x)
func encodingMultiplierForSize(originalSize int64) float64 {
	sizeMiB := float64(originalSize) / (1024 * 1024)

	switch {
	case sizeMiB < 1:
		return 10.0 // Very small files have high overhead
	case sizeMiB < 16:
		return 8.5
	case sizeMiB < 100:
		return 6.0
	case sizeMiB < 500:
		return 5.0
	default:
		return 4.5
	}
}

// CalculateEncodedSize calculates the encoded blob size after Reed-Solomon encoding.
// The encoded size is approximately 5-8x the original size depending on blob size.
func CalculateEncodedSize(originalSize int64) int64 {
	return int64(float64(originalSize) * encodingMultiplierForSize(originalSize))
}

// calculateEncodedSizeWithMultiplier calculates the encoded blob size using the
// given multiplier from live walrus data instead of the heuristic sliding scale.
func calculateEncodedSizeWithMultiplier(originalSize int64, multiplier float64) int64 {
	return int64(float64(originalSize) * multiplier)
}

// DryRunResult represents the JSON output from 'walrus store --dry-run --json'
type DryRunResult struct {
	UnencodedSize int64  `json:"unencodedSize"`
	EncodedSize   int64  `json:"encodedSize"`
	Cost          uint64 `json:"cost"` // Cost in FROST
}

// GetEncodedSizeFromDryRun uses 'walrus store --dry-run --json' for accurate encoded size
// It automatically uses the correct context based on the active Sui environment
func GetEncodedSizeFromDryRun(filePath string, walrusBin string) (int64, error) {
	if walrusBin == "" {
		walrusBin = "walrus"
		if path, err := deps.LookPath("walrus"); err == nil {
			walrusBin = path
		}
	}

	// Get the context from Sui active environment
	walrusCtx := GetWalrusContext()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := execCommandContext(ctx, walrusBin, "store", "--dry-run", "--json", "--context", walrusCtx, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("dry-run failed: %w", err)
	}

	// Strip ANSI color codes that walrus CLI may include
	output = stripANSI(output)

	// Find JSON in output (walrus outputs log lines before JSON)
	jsonStart := bytes.IndexByte(output, '{')
	if jsonStart == -1 {
		return 0, fmt.Errorf("no JSON found in dry-run output")
	}

	var result DryRunResult
	if err := json.Unmarshal(output[jsonStart:], &result); err != nil {
		return 0, fmt.Errorf("failed to parse dry-run JSON: %w", err)
	}

	return result.EncodedSize, nil
}

// GetRPCEndpoint returns the appropriate RPC endpoint for the network
func GetRPCEndpoint(network string) string {
	switch strings.ToLower(network) {
	case "mainnet":
		return SuiMainnetRPC
	case "testnet":
		return SuiTestnetRPC
	default:
		return SuiTestnetRPC
	}
}

// DefaultGasPrice returns the fallback gas price for the network
func DefaultGasPrice(network string) uint64 {
	switch strings.ToLower(network) {
	case "testnet":
		return 750 // Lower gas prices on testnet
	case "mainnet":
		return 1000 // Higher gas prices on mainnet
	default:
		return 750
	}
}

// CalculateCost calculates the full cost for deploying a site
// Based on Sui gas docs: https://docs.sui.io/concepts/tokenomics/gas-in-sui
// And Walrus cost docs: https://docs.wal.app/docs/dev-guide/costs
func CalculateCost(options CostOptions) (*CostBreakdown, error) {
	// Validate inputs
	if options.SiteSize <= 0 {
		return nil, fmt.Errorf("site size must be positive")
	}
	if options.Epochs <= 0 {
		return nil, fmt.Errorf("epochs must be greater than 0")
	}

	// Get RPC endpoint
	rpcURL := options.RPCURL
	if rpcURL == "" {
		rpcURL = GetRPCEndpoint(options.Network)
	}

	// Fetch real gas price from Sui RPC
	gasPrice := options.GasPrice
	if gasPrice == 0 {
		var err error
		gasPrice, err = GetReferenceGasPrice(rpcURL)
		if err != nil {
			// Fall back to default if RPC fails
			gasPrice = DefaultGasPrice(options.Network)
		}
	}

	// Try to get real walrus storage info from 'walrus info --json'
	var storageInfo *StorageInfo
	storageInfo, storageErr := GetStorageInfo(options.WalrusBin)
	if storageErr != nil && isVerbose() {
		fmt.Printf("   Note: Could not get live storage pricing (%v), using defaults\n", storageErr)
	}
	if storageInfo == nil {
		// Use defaults based on actual Walrus pricing (Dec 2025)
		// Mainnet is ~10-11x more expensive than testnet
		// From `walrus info --json`: 1 storage unit = 1 MiB
		network := options.Network
		if network == "" {
			network = GetWalrusContext()
		}

		if strings.ToLower(network) == "mainnet" {
			// Mainnet pricing (Dec 2025)
			storageInfo = &StorageInfo{
				StoragePrice:       11000,   // 11,000 FROST per MiB per epoch
				WritePrice:         20000,   // 20,000 FROST per MiB (one-time)
				MetadataPrice:      682000,  // Fixed metadata cost in FROST
				MarginalPrice:      66000,   // Per unencoded MiB cost in FROST
				StorageUnitSize:    1048576, // 1 MiB
				EpochDuration:      1209600, // 14 days
				EncodingMultiplier: 8.0,     // Reed-Solomon ~8x expansion for small sites
			}
		} else {
			// Testnet pricing (Dec 2025)
			storageInfo = &StorageInfo{
				StoragePrice:       1000,    // 1,000 FROST per MiB per epoch
				WritePrice:         2000,    // 2,000 FROST per MiB (one-time)
				MetadataPrice:      62000,   // Fixed metadata cost in FROST
				MarginalPrice:      6000,    // Per unencoded MiB cost in FROST
				StorageUnitSize:    1048576, // 1 MiB
				EpochDuration:      86400,   // 1 day
				EncodingMultiplier: 8.0,     // Reed-Solomon ~8x expansion for small sites
			}
		}
	}

	// Calculate encoded size in MiB (storage units)
	// Use live encoding multiplier from walrus info when available,
	// otherwise fall back to the size-based heuristic.
	var encodedSizeBytes int64
	if storageInfo.EncodingMultiplier > 0 {
		encodedSizeBytes = calculateEncodedSizeWithMultiplier(options.SiteSize, storageInfo.EncodingMultiplier)
	} else {
		encodedSizeBytes = CalculateEncodedSize(options.SiteSize)
	}
	storageUnitSize := storageInfo.StorageUnitSize
	if storageUnitSize <= 0 {
		storageUnitSize = 1048576 // Default 1 MiB
	}
	encodedSizeMiB := math.Ceil(float64(encodedSizeBytes) / float64(storageUnitSize))
	if encodedSizeMiB < 1 {
		encodedSizeMiB = 1 // Minimum 1 storage unit
	}

	// Estimate file count if not provided
	fileCount := options.FileCount
	if fileCount <= 0 {
		// Average web file is ~50KB
		fileCount = int(math.Ceil(float64(options.SiteSize) / (50 * 1024)))
		if fileCount < 1 {
			fileCount = 1
		}
	}

	// Calculate WAL storage cost (per walrus info --json pricing)
	// Formula: metadata_price + (encoded_storage_units × storage_price × epochs) + (encoded_storage_units × write_price)
	// 1 WAL = 1,000,000,000 FROST
	metadataCostFrost := float64(storageInfo.MetadataPrice)
	if metadataCostFrost == 0 {
		metadataCostFrost = 62000 // Default from walrus info
	}
	storageCostFrost := encodedSizeMiB * float64(storageInfo.StoragePrice) * float64(options.Epochs)
	writeCostFrost := encodedSizeMiB * float64(storageInfo.WritePrice)

	totalFrost := metadataCostFrost + storageCostFrost + writeCostFrost
	storageCostWAL := storageCostFrost / 1e9
	writeCostWAL := (writeCostFrost + metadataCostFrost) / 1e9
	totalWAL := totalFrost / 1e9

	// Calculate SUI gas cost for transactions
	// Based on Sui docs: total_gas = computation_units × gas_price + storage_units × storage_price
	// Computation buckets: 1,000 to 5,000,000 units
	// Storage: 100 units per byte of on-chain storage
	//
	// site-builder typically runs 3 transactions:
	// 1. Reserve space (if needed)
	// 2. Register blob and assign blob ID
	// 3. Certify blob availability
	baseComputationUnits := uint64(500000)   // Base computation for site creation (mid-range bucket)
	perFileComputationUnits := uint64(10000) // Per-file computation
	totalComputationUnits := baseComputationUnits + uint64(fileCount)*perFileComputationUnits

	// On-chain storage for site metadata (estimated ~1KB per file for object storage)
	onChainStorageBytes := uint64(1024 * fileCount)
	storageUnits := onChainStorageBytes * 100 // 100 units per byte

	// Storage price is typically 76 MIST per unit on mainnet (governance-set)
	storageUnitPrice := uint64(76)
	if strings.ToLower(options.Network) == "testnet" {
		storageUnitPrice = 76 // Same on testnet
	}

	// Total gas = computation + storage
	computationCost := totalComputationUnits * gasPrice
	storageCost := storageUnits * storageUnitPrice
	totalGasUnits := totalComputationUnits + storageUnits
	gasCostSUI := float64(computationCost+storageCost) / 1e9

	// Calculate ranges (accounting for network variability)
	minWAL := totalWAL * 0.8
	maxWAL := totalWAL * 1.2
	minSUI := gasCostSUI * 0.7
	maxSUI := gasCostSUI * 1.5

	return &CostBreakdown{
		GasUnits:       totalGasUnits,
		GasPrice:       gasPrice,
		GasCostSUI:     gasCostSUI,
		StorageCostWAL: storageCostWAL,
		WriteCostWAL:   writeCostWAL,
		TotalWAL:       totalWAL,
		EncodedSize:    encodedSizeBytes,
		OriginalSize:   options.SiteSize,
		FileCount:      fileCount,
		Epochs:         options.Epochs,
		MinTotalWAL:    minWAL,
		MaxTotalWAL:    maxWAL,
		MinTotalSUI:    minSUI,
		MaxTotalSUI:    maxSUI,
	}, nil
}

// CalculateUpdateCost calculates cost for updating an existing site
func CalculateUpdateCost(changedSize int64, newFiles int, epochs int, network string) (*CostBreakdown, error) {
	if changedSize <= 0 && newFiles <= 0 {
		// No changes, just metadata update
		rpcURL := GetRPCEndpoint(network)
		gasPrice, err := GetReferenceGasPrice(rpcURL)
		if err != nil {
			gasPrice = DefaultGasPrice(network)
		}

		gasUnits := uint64(50000) // Single transaction
		gasCostSUI := float64(gasUnits) * float64(gasPrice) / 1e9

		return &CostBreakdown{
			GasUnits:    gasUnits,
			GasPrice:    gasPrice,
			GasCostSUI:  gasCostSUI,
			MinTotalSUI: gasCostSUI * 0.7,
			MaxTotalSUI: gasCostSUI * 1.5,
		}, nil
	}

	// Updates only pay for new/changed content
	return CalculateCost(CostOptions{
		SiteSize:  changedSize,
		Epochs:    epochs,
		FileCount: newFiles,
		Network:   network,
	})
}

// CalculateDestroyCost calculates cost for destroying a site
func CalculateDestroyCost(network string) (*CostBreakdown, error) {
	rpcURL := GetRPCEndpoint(network)
	gasPrice, err := GetReferenceGasPrice(rpcURL)
	if err != nil {
		gasPrice = DefaultGasPrice(network)
	}

	// Destroy is a single transaction
	gasUnits := uint64(100000) // Slightly higher for cleanup
	gasCostSUI := float64(gasUnits) * float64(gasPrice) / 1e9

	return &CostBreakdown{
		GasUnits:    gasUnits,
		GasPrice:    gasPrice,
		GasCostSUI:  gasCostSUI,
		MinTotalSUI: gasCostSUI * 0.7,
		MaxTotalSUI: gasCostSUI * 1.5,
	}, nil
}

// FormatCostBreakdown formats the cost breakdown for display
func FormatCostBreakdown(breakdown CostBreakdown) string {
	var builder strings.Builder

	builder.WriteString("\nCost Breakdown\n")
	builder.WriteString("─────────────────────────────────────\n\n")

	if breakdown.OriginalSize > 0 {
		builder.WriteString("Data Size:\n")
		builder.WriteString(fmt.Sprintf("  Original: %s\n", formatBytes(breakdown.OriginalSize)))
		builder.WriteString(fmt.Sprintf("  Encoded:  %s (after Reed-Solomon)\n\n", formatBytes(breakdown.EncodedSize)))
	}

	if breakdown.TotalWAL > 0 {
		builder.WriteString("Storage Cost (WAL):\n")
		builder.WriteString(fmt.Sprintf("  Storage:  %.6f WAL (%d epochs)\n", breakdown.StorageCostWAL, breakdown.Epochs))
		builder.WriteString(fmt.Sprintf("  Write:    %.6f WAL\n", breakdown.WriteCostWAL))
		builder.WriteString(fmt.Sprintf("  Total:    %.6f WAL\n", breakdown.TotalWAL))
		builder.WriteString(fmt.Sprintf("  Range:    %.6f - %.6f WAL\n\n", breakdown.MinTotalWAL, breakdown.MaxTotalWAL))
	}

	builder.WriteString("Transaction Cost (SUI):\n")
	builder.WriteString(fmt.Sprintf("  Gas Units: %d\n", breakdown.GasUnits))
	builder.WriteString(fmt.Sprintf("  Gas Price: %d MIST\n", breakdown.GasPrice))
	builder.WriteString(fmt.Sprintf("  Total:     %.6f SUI\n", breakdown.GasCostSUI))
	builder.WriteString(fmt.Sprintf("  Range:     %.6f - %.6f SUI\n\n", breakdown.MinTotalSUI, breakdown.MaxTotalSUI))

	if breakdown.FileCount > 0 {
		builder.WriteString(fmt.Sprintf("Files: %d\n\n", breakdown.FileCount))
	}

	builder.WriteString("─────────────────────────────────────\n")
	builder.WriteString("Note: WAL is used for Walrus storage, SUI for Sui transactions.\n")
	builder.WriteString("Actual costs may vary based on network conditions.\n")
	builder.WriteString("Use https://costcalculator.wal.app for official estimates.\n")

	return builder.String()
}

// formatBytes formats bytes to human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	suffixes := "KMGTPE"
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit && exp < len(suffixes)-1; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), suffixes[exp])
}

// FormatCostSummary creates a user-friendly summary with separate WAL and SUI amounts.
func FormatCostSummary(walCost, suiCost float64, fileCount int, epochs int) string {
	return fmt.Sprintf(
		"Estimated: %.4f WAL + %.4f SUI for %d files stored for %d epochs",
		walCost, suiCost, fileCount, epochs,
	)
}

// EstimateCostSimple provides a quick estimation string
func EstimateCostSimple(siteSize int64, epochs int, network string) string {
	breakdown, err := CalculateCost(CostOptions{
		SiteSize: siteSize,
		Epochs:   epochs,
		Network:  network,
	})
	if err != nil {
		return "Unable to estimate cost"
	}

	return fmt.Sprintf("~%.4f WAL + ~%.4f SUI (range: %.4f-%.4f WAL, %.4f-%.4f SUI)",
		breakdown.TotalWAL, breakdown.GasCostSUI,
		breakdown.MinTotalWAL, breakdown.MaxTotalWAL,
		breakdown.MinTotalSUI, breakdown.MaxTotalSUI)
}
