package walrus

import (
	"context"
	"fmt"
	"math"
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

// latestTransactionQuery reads the most recent transaction sent by an address
// together with its balance changes. Replaces the retired
// suix_queryTransactionBlocks JSON-RPC method; note that the filter key is
// sentAddress (was FromAddress) and status is an uppercase enum (was "success").
const latestTransactionQuery = `query($address: SuiAddress!) {
  transactions(last: 1, filter: { sentAddress: $address }) {
    nodes {
      digest
      effects {
        status
        balanceChanges { nodes { amount coinType { repr } } }
      }
    }
  }
}`

// latestTransactionResult mirrors the GraphQL response for latestTransactionQuery.
type latestTransactionResult struct {
	Transactions struct {
		Nodes []struct {
			Digest  string `json:"digest"`
			Effects struct {
				Status         string `json:"status"`
				BalanceChanges struct {
					Nodes []struct {
						Amount   string `json:"amount"`
						CoinType struct {
							Repr string `json:"repr"`
						} `json:"coinType"`
					} `json:"nodes"`
				} `json:"balanceChanges"`
			} `json:"effects"`
		} `json:"nodes"`
	} `json:"transactions"`
}

// GetLatestTransactionGas queries the Sui GraphQL API for the latest transaction
// from a wallet and returns the gas information
func GetLatestTransactionGas(walletAddress, network string) (*TransactionGasInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result latestTransactionResult
	err := queryGraphQL(
		ctx,
		GetGraphQLEndpoint(network),
		latestTransactionQuery,
		map[string]any{"address": walletAddress},
		&result,
	)
	if err != nil {
		return nil, err
	}

	if len(result.Transactions.Nodes) == 0 {
		return nil, fmt.Errorf("no transactions found for wallet %s", walletAddress)
	}

	tx := result.Transactions.Nodes[0]

	// Extract costs from balance changes.
	// A single transaction may produce multiple balance changes per coin type
	// (e.g., gas payment + storage rebate), so we accumulate all spends.
	var totalSUI, totalWAL float64
	for _, bc := range tx.Effects.BalanceChanges.Nodes {
		amount, err := strconv.ParseInt(bc.Amount, 10, 64)
		if err != nil {
			continue // Skip malformed amounts
		}
		if amount >= 0 {
			continue // Skip non-spend (only negative amounts are outflows)
		}

		// Check coin type and accumulate spent amount
		coinTypeLower := strings.ToLower(bc.CoinType.Repr)
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
		Success:     strings.EqualFold(tx.Effects.Status, "SUCCESS"),
	}, nil
}
