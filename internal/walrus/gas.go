package walrus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TransactionGasInfo contains gas information for a transaction
type TransactionGasInfo struct {
	Digest      string
	TotalGasSUI float64 // Total SUI spent (from balance changes)
	TotalWAL    float64 // Total WAL spent (from balance changes)
	Success     bool
}

// rpcRequest represents a JSON-RPC request
type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// rpcResponse represents a JSON-RPC response
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// balanceChange represents a coin balance change in a transaction
type balanceChange struct {
	Owner struct {
		AddressOwner string `json:"AddressOwner"`
	} `json:"owner"`
	CoinType string `json:"coinType"`
	Amount   string `json:"amount"`
}

// queryTransactionBlocksResult represents the result of suix_queryTransactionBlocks
type queryTransactionBlocksResult struct {
	Data []struct {
		Digest  string `json:"digest"`
		Effects struct {
			Status struct {
				Status string `json:"status"`
			} `json:"status"`
		} `json:"effects"`
		BalanceChanges []balanceChange `json:"balanceChanges"`
	} `json:"data"`
	HasNextPage bool   `json:"hasNextPage"`
	NextCursor  string `json:"nextCursor"`
}

// GetLatestTransactionGas queries the Sui RPC for the latest transaction from a wallet
// and returns the gas information
func GetLatestTransactionGas(walletAddress, network string) (*TransactionGasInfo, error) {
	rpcURL := GetRPCEndpoint(network)

	// Query the latest transaction from this wallet
	params := []interface{}{
		map[string]interface{}{
			"filter": map[string]string{
				"FromAddress": walletAddress,
			},
			"options": map[string]bool{
				"showEffects":        true,
				"showBalanceChanges": true,
			},
		},
		nil,  // cursor
		1,    // limit - just get the latest one
		true, // descending order (newest first)
	}

	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "suix_queryTransactionBlocks",
		Params:  params,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(rpcURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("RPC request failed: %w", err)
	}
	defer resp.Body.Close()

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	var result queryTransactionBlocksResult
	if err := json.Unmarshal(rpcResp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no transactions found for wallet %s", walletAddress)
	}

	tx := result.Data[0]

	// Extract costs from balance changes.
	// A single transaction may produce multiple balance changes per coin type
	// (e.g., gas payment + storage rebate), so we accumulate all spends.
	var totalSUI, totalWAL float64
	for _, bc := range tx.BalanceChanges {
		amount, err := strconv.ParseInt(bc.Amount, 10, 64)
		if err != nil {
			continue // Skip malformed amounts
		}
		if amount >= 0 {
			continue // Skip non-spend (only negative amounts are outflows)
		}

		// Check coin type and accumulate spent amount
		coinTypeLower := strings.ToLower(bc.CoinType)
		if strings.Contains(coinTypeLower, "sui::sui") {
			// SUI spent (1 SUI = 1e9 MIST)
			totalSUI += math.Abs(float64(amount)) / 1e9
		} else if strings.Contains(coinTypeLower, "wal::wal") {
			// WAL spent (1 WAL = 1e9 FROST)
			totalWAL += math.Abs(float64(amount)) / 1e9
		}
	}

	return &TransactionGasInfo{
		Digest:      tx.Digest,
		TotalGasSUI: totalSUI,
		TotalWAL:    totalWAL,
		Success:     tx.Effects.Status.Status == "success",
	}, nil
}
