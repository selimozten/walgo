package walrus

import (
	"testing"
	"time"
)

// sitemapOutput is real site-builder 2.12 output, ANSI codes and all.
const sitemapOutput = "\x1b[38;5;11m\x1b[1m\nMultiple blob objects found for blob_id wPGB08NRNWait8SjnfuNElFnjqOuVxa5UunWuwNr24o. " +
	"Using 0x0205b83382f37220e9db8a41bb17abf8f8fa3086001ed459b81d4a77e24e8150 (end_epoch 73) instead.\n\x1b[0m" +
	"-------------------------------------------------------------------------------\n" +
	" Resource path          Blob / Quilt Patch ID                               Owned blob object ID (if any)                                       Earliest Expiration Date \n" +
	"-------------------------------------------------------------------------------\n" +
	" /categories/index.xml  IRaHhCENFBOmPjxQKh7W4Ob4anB8_JEZRb86FtKK110BAQACAA  0x75988d9991bf3975aa362b8798aa7d26532c3db55f3694f599677a0f3ccb63ff  2026-08-11 \n" +
	" /index.html            0nLObHz-oI39cAeXPyIJJ1n9p0XK6csAtG2d1viFke8BAQACAA  0x69f3aff948a9ffead94dd5ea6cf48907367cfe032a97f06808af8881ad845230  2026-08-11 \n" +
	"-------------------------------------------------------------------------------\n"

func TestParseSitemapOutputTable(t *testing.T) {
	result := parseSitemapOutput(sitemapOutput)

	if len(result.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d: %+v", len(result.Resources), result.Resources)
	}

	first := result.Resources[0]
	if first.Path != "/categories/index.xml" {
		t.Errorf("path = %q, want /categories/index.xml", first.Path)
	}
	if first.BlobID != "IRaHhCENFBOmPjxQKh7W4Ob4anB8_JEZRb86FtKK110BAQACAA" {
		t.Errorf("blob id = %q", first.BlobID)
	}
	if first.BlobObjectID != "0x75988d9991bf3975aa362b8798aa7d26532c3db55f3694f599677a0f3ccb63ff" {
		t.Errorf("blob object id = %q", first.BlobObjectID)
	}
	want := time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC)
	if !first.Expiry.Equal(want) {
		t.Errorf("expiry = %v, want %v", first.Expiry, want)
	}
}

func TestParseSitemapOutputLegacy(t *testing.T) {
	legacy := "resource /index.html with blob ID abc123\nresource /style.css with blob ID def456\n"

	result := parseSitemapOutput(legacy)

	if len(result.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(result.Resources))
	}
	if result.Resources[1].Path != "/style.css" || result.Resources[1].BlobID != "def456" {
		t.Errorf("unexpected resource: %+v", result.Resources[1])
	}
}

func TestParseSitemapOutputIgnoresNoise(t *testing.T) {
	noise := "Multiple blob objects found for blob_id wPGB0\n---------\n Resource path  Blob / Quilt Patch ID \n"

	if resources := parseSitemapOutput(noise).Resources; len(resources) != 0 {
		t.Errorf("expected no resources, got %+v", resources)
	}
}

func TestSummarizeExpiry(t *testing.T) {
	now := time.Date(2026, 8, 12, 15, 0, 0, 0, time.UTC)

	resources := []Resource{
		{Path: "/expired.html", Expiry: time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC)},
		{Path: "/today.html", Expiry: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)},
		{Path: "/later.html", Expiry: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)},
		{Path: "/unknown.html"},
	}

	summary := SummarizeExpiry(resources, now)

	if summary.Total != 4 {
		t.Errorf("total = %d, want 4", summary.Total)
	}
	if len(summary.Expired) != 1 || summary.Expired[0] != "/expired.html" {
		t.Errorf("expired = %v, want [/expired.html]", summary.Expired)
	}
	if summary.Unknown != 1 {
		t.Errorf("unknown = %d, want 1", summary.Unknown)
	}
	// A resource expiring today is still live, so it sets the earliest date.
	if want := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC); !summary.Earliest.Equal(want) {
		t.Errorf("earliest = %v, want %v", summary.Earliest, want)
	}
	if !summary.HasExpired() {
		t.Error("HasExpired() = false, want true")
	}
}

func TestSummarizeExpiryAllHealthy(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	resources := []Resource{
		{Path: "/a.html", Expiry: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)},
		{Path: "/b.html", Expiry: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)},
	}

	summary := SummarizeExpiry(resources, now)

	if summary.HasExpired() {
		t.Error("HasExpired() = true, want false")
	}
	if want := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC); !summary.Earliest.Equal(want) {
		t.Errorf("earliest = %v, want %v", summary.Earliest, want)
	}
	if got := summary.Describe(now); got != "expires in 14 days (2026-08-15)" {
		t.Errorf("Describe() = %q", got)
	}
}

func TestSummarizeExpiryDescribeExpired(t *testing.T) {
	now := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	resources := []Resource{
		{Path: "/a.html", Expiry: time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)},
		{Path: "/b.html", Expiry: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)},
	}

	if got := SummarizeExpiry(resources, now).Describe(now); got != "1 of 2 resources expired" {
		t.Errorf("Describe() = %q", got)
	}
}

func TestSummarizeExpiryEmpty(t *testing.T) {
	summary := SummarizeExpiry(nil, time.Now())

	if summary.HasExpired() {
		t.Error("HasExpired() = true for empty site")
	}
	if got := summary.Describe(time.Now()); got != "no resources found" {
		t.Errorf("Describe() = %q", got)
	}
}
