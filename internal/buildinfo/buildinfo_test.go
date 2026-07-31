package buildinfo

import (
	"strings"
	"testing"
)

func TestVersionPrefersLdflags(t *testing.T) {
	defer restore(version, commit, buildDate)

	version = "v1.2.3"
	if got := Version(); got != "1.2.3" {
		t.Errorf("Version() = %q, want 1.2.3 (leading v stripped)", got)
	}
}

func TestVersionFallsBack(t *testing.T) {
	defer restore(version, commit, buildDate)

	version = ""
	got := Version()
	if got == "" {
		t.Fatal("Version() is empty without ldflags")
	}
	if strings.HasPrefix(got, "v") {
		t.Errorf("Version() = %q, should not carry a leading v", got)
	}
}

func TestCommitPrefersLdflags(t *testing.T) {
	defer restore(version, commit, buildDate)

	commit = "abc1234"
	if got := Commit(); got != "abc1234" {
		t.Errorf("Commit() = %q, want abc1234", got)
	}
}

func TestCommitFallsBack(t *testing.T) {
	defer restore(version, commit, buildDate)

	commit = ""
	// Tests build from the repository, so the toolchain stamps a revision; a
	// build from outside one reports the placeholder instead.
	if got := Commit(); got == "" {
		t.Error("Commit() is empty without ldflags")
	}
}

func TestBuildDatePrefersLdflags(t *testing.T) {
	defer restore(version, commit, buildDate)

	buildDate = "2026-08-01T00:00:00Z"
	if got := BuildDate(); got != "2026-08-01T00:00:00Z" {
		t.Errorf("BuildDate() = %q", got)
	}
}

func TestBuildDateFallsBack(t *testing.T) {
	defer restore(version, commit, buildDate)

	buildDate = ""
	if got := BuildDate(); got == "" {
		t.Error("BuildDate() is empty without ldflags")
	}
}

// restore puts the injected build values back after a test mutates them.
func restore(v, c, d string) {
	version, commit, buildDate = v, c, d
}
