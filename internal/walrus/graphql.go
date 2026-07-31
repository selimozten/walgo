package walrus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Sui GraphQL endpoints. Sui Foundation disabled JSON-RPC on its public fullnodes
// on 2026-07-31 (full decommission mid-October 2026); GraphQL replaces it for the
// chain reads walgo needs. See https://docs.sui.io/develop/accessing-data/json-rpc-migration
const (
	SuiTestnetGraphQL = "https://graphql.testnet.sui.io/graphql"
	SuiMainnetGraphQL = "https://graphql.mainnet.sui.io/graphql"
)

// GetGraphQLEndpoint returns the Sui GraphQL endpoint for the network.
// Unknown networks fall back to testnet.
func GetGraphQLEndpoint(network string) string {
	switch strings.ToLower(network) {
	case "mainnet":
		return SuiMainnetGraphQL
	default:
		return SuiTestnetGraphQL
	}
}

// graphQLRequest is a GraphQL over HTTP request body.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// graphQLResponse is a GraphQL over HTTP response body.
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// queryGraphQL posts query to endpoint and unmarshals the response "data" field
// into out. Errors reported by the server are returned as a single error.
func queryGraphQL(ctx context.Context, endpoint, query string, variables map[string]any, out any) error {
	reqBody, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create GraphQL request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GraphQL request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("failed to read GraphQL response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GraphQL request failed with status %d", resp.StatusCode)
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		messages := make([]string, 0, len(gqlResp.Errors))
		for _, e := range gqlResp.Errors {
			messages = append(messages, e.Message)
		}
		return fmt.Errorf("GraphQL error: %s", strings.Join(messages, "; "))
	}

	if len(gqlResp.Data) == 0 {
		return fmt.Errorf("GraphQL response contained no data")
	}

	if err := json.Unmarshal(gqlResp.Data, out); err != nil {
		return fmt.Errorf("failed to parse GraphQL data: %w", err)
	}

	return nil
}
