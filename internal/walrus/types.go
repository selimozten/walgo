package walrus

import "time"

const siteBuilderCmd = "site-builder"

// DefaultCommandTimeout is the maximum time allowed for site-builder operations.
// Deployments can take a while for large sites, so we use a generous timeout.
const DefaultCommandTimeout = 10 * time.Minute

// TokenInfo represents information about a token type.
type TokenInfo struct {
	Decimals    int    `json:"decimals"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	ID          string `json:"id"`
}

// CoinBalance represents a coin balance.
type CoinBalance struct {
	CoinType            string `json:"coinType"`
	CoinObjectId        string `json:"coinObjectId"`
	Version             string `json:"version"`
	Digest              string `json:"digest"`
	Balance             string `json:"balance"`
	PreviousTransaction string `json:"previousTransaction"`
}

// BalanceEntry is an entry in the balance array containing token info and coins.
type BalanceEntry struct {
	TokenInfo TokenInfo     `json:"-"`
	Coins     []CoinBalance `json:"-"`
}

// WalletBalance represents the full wallet balance response.
type WalletBalance struct {
	Entries []BalanceEntry
	HasMore bool `json:"hasMore"`
}

// SiteBuilderOutput contains the result of site-builder operations.
type SiteBuilderOutput struct {
	ObjectID   string
	SiteURL    string
	BrowseURLs []string
	Resources  []Resource
	Base36ID   string
	Success    bool
}

// Resource represents a deployed site resource.
type Resource struct {
	Path   string
	BlobID string
}
