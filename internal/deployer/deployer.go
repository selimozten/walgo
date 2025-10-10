package deployer

import (
	"context"
)

// Result captures the outcome of a deployment/update/status operation.
type Result struct {
	Success      bool
	ObjectID     string            // For site objects (site-builder path)
	BrowseURLs   []string          // Optional links returned by underlying tool
	FileToBlobID map[string]string // For HTTP per-blob uploads: relative path -> blobId
	Message      string
}

// DeployOptions configures deploy behavior.
type DeployOptions struct {
	// Generic
	Epochs   int
	Verbose  bool
	JSONLogs bool

	// HTTP-specific
	PublisherBaseURL  string // e.g., https://publisher.walrus-testnet.walrus.space
	AggregatorBaseURL string // e.g., https://aggregator.walrus-testnet.walrus.space
	Mode              string // "quilt" or "blobs"
	Workers           int    // number of concurrent workers for blobs mode
	MaxRetries        int    // per-file max retries
}

// WalrusDeployer provides a common interface across deployment backends.
type WalrusDeployer interface {
	Deploy(ctx context.Context, siteDir string, opts DeployOptions) (*Result, error)
	Update(ctx context.Context, siteDir string, objectID string, opts DeployOptions) (*Result, error)
	Status(ctx context.Context, objectID string, opts DeployOptions) (*Result, error)
}
