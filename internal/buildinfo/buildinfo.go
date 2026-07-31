// Package buildinfo is the single source of the version reported by the CLI,
// the API and the desktop app.
package buildinfo

import (
	"runtime/debug"
	"strings"
	"time"
)

// Values injected at build time. GoReleaser and the desktop build set these
// through -X flags; see .goreleaser.yml and .github/workflows/release.yml.
//
// Keep releaseVersion in sync with the current release: it is what a plain
// `go build ./...` reports, and what tags are cut from.
var (
	version   string
	commit    string
	buildDate string
)

const (
	releaseVersion = "0.4.1"
	unknownCommit  = "dev"
	unknownDate    = "unknown"
)

// Version returns the release version, without a leading "v".
func Version() string {
	if version != "" {
		return strings.TrimPrefix(version, "v")
	}

	// `go install github.com/ganbitlabs/walgo@v1.2.3` records the version in the
	// build info even though no ldflags were passed.
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}

	return releaseVersion
}

// Commit returns the git revision the binary was built from, suffixed with
// "-dirty" when the working tree had uncommitted changes.
func Commit() string {
	if commit != "" {
		return commit
	}

	revision, modified := vcsInfo()
	if revision == "" {
		return unknownCommit
	}
	if modified {
		return revision + "-dirty"
	}

	return revision
}

// BuildDate returns when the binary was built, in RFC 3339 form.
func BuildDate() string {
	if buildDate != "" {
		return buildDate
	}

	if _, _, buildTime := vcsSettings(); buildTime != "" {
		return buildTime
	}

	return unknownDate
}

// vcsInfo reports the revision recorded by the Go toolchain and whether the
// working tree was dirty. Both are empty for builds outside a repository.
func vcsInfo() (revision string, modified bool) {
	revision, dirty, _ := vcsSettings()
	return revision, dirty == "true"
}

// vcsSettings extracts the vcs.* build settings the toolchain stamps into
// binaries built from a repository.
func vcsSettings() (revision, modified, buildTime string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", "", ""
	}

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value
		case "vcs.time":
			// Normalize so it matches the format used by the ldflags path.
			if parsed, err := time.Parse(time.RFC3339, setting.Value); err == nil {
				buildTime = parsed.UTC().Format(time.RFC3339)
			} else {
				buildTime = setting.Value
			}
		}
	}

	return revision, modified, buildTime
}
