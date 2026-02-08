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
	Path   string
	BlobID string
}
