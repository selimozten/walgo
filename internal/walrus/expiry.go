package walrus

import (
	"fmt"
	"time"
)

// SiteExpiry summarizes when a deployed site's storage runs out.
//
// A Walrus Site is two things with different lifetimes: the Sui object, which
// never expires, and the blobs holding the content, which do. Once the blobs
// expire the object still exists but serves nothing, and the storage cannot be
// extended any more — the walrus contract rejects extending an expired blob
// (blob.move: assert_certified_not_expired). The only way back is re-uploading
// the content from a local copy, at full price.
type SiteExpiry struct {
	// Total is the number of resources reported for the site.
	Total int
	// Expired lists resource paths whose storage has already run out.
	Expired []string
	// Earliest is the soonest expiration among resources that still have one.
	Earliest time.Time
	// Unknown counts resources site-builder reported without an expiration
	// date, which happens when the wallet owns no blob object for them.
	Unknown int
}

// HasExpired reports whether at least one resource is past its expiration.
func (e *SiteExpiry) HasExpired() bool {
	return e != nil && len(e.Expired) > 0
}

// TimeLeft returns how long until the earliest expiration. It is negative when
// that moment has already passed and zero when no expiration date is known.
func (e *SiteExpiry) TimeLeft(now time.Time) time.Duration {
	if e == nil || e.Earliest.IsZero() {
		return 0
	}
	return e.Earliest.Sub(now)
}

// Describe renders a one-line human-readable summary of the site's lifetime.
func (e *SiteExpiry) Describe(now time.Time) string {
	if e == nil || e.Total == 0 {
		return "no resources found"
	}

	if e.HasExpired() {
		return fmt.Sprintf("%d of %d resources expired", len(e.Expired), e.Total)
	}

	if e.Earliest.IsZero() {
		return "expiration unknown"
	}

	days := int(e.TimeLeft(now).Hours() / 24)
	switch {
	case days < 1:
		return fmt.Sprintf("expires today (%s)", e.Earliest.Format(time.DateOnly))
	case days == 1:
		return fmt.Sprintf("expires tomorrow (%s)", e.Earliest.Format(time.DateOnly))
	default:
		return fmt.Sprintf("expires in %d days (%s)", days, e.Earliest.Format(time.DateOnly))
	}
}

// SummarizeExpiry folds site resources into a SiteExpiry.
//
// Expiration dates are day-granular, and a resource is only unusable once that
// day has passed, so a resource expiring today still counts as live.
func SummarizeExpiry(resources []Resource, now time.Time) *SiteExpiry {
	summary := &SiteExpiry{Total: len(resources)}
	today := now.UTC().Truncate(24 * time.Hour)

	for _, resource := range resources {
		switch {
		case resource.Expiry.IsZero():
			summary.Unknown++
		case resource.Expiry.Before(today):
			summary.Expired = append(summary.Expired, resource.Path)
		default:
			if summary.Earliest.IsZero() || resource.Expiry.Before(summary.Earliest) {
				summary.Earliest = resource.Expiry
			}
		}
	}

	return summary
}

// GetSiteExpiry reads a site's resources from chain and summarizes their
// lifetime. Unlike GetSiteStatus it prints nothing, so callers can use it as a
// pre-flight check.
func GetSiteExpiry(objectID string) (*SiteExpiry, error) {
	output, err := runSitemap(objectID)
	if err != nil {
		return nil, err
	}

	return SummarizeExpiry(parseSitemapOutput(output).Resources, time.Now()), nil
}
