package walrus

import "time"

const siteBuilderCmd = "site-builder"

// DefaultCommandTimeout is the maximum time allowed for site-builder operations.
// Deployments can take a while for large sites, so we use a generous timeout.
const DefaultCommandTimeout = 10 * time.Minute

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
	Path string
	// BlobID is the blob ID or, for quilted sites, the quilt patch ID.
	BlobID string
	// BlobObjectID is the owned Sui blob object backing the resource, when the
	// wallet owns one. Empty for resources stored by somebody else.
	BlobObjectID string
	// Expiry is when the backing storage runs out. Zero when site-builder
	// reported no expiration date (typically because no owned blob was found).
	Expiry time.Time
}
