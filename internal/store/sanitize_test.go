package store

import (
	"strings"
	"testing"
)

func TestValidateSlug(t *testing.T) {
	validSlugs := []string{
		"auth-flow",
		"database_migration",
		"pino-vs-winston",
		"v2",
		"cache123",
	}
	for _, s := range validSlugs {
		if err := ValidateSlug(s); err != nil {
			t.Errorf("expected %q to be valid, got: %v", s, err)
		}
	}

	invalidSlugs := []string{
		"-starts-with-dash",
		"_starts-with-underscore",
		"contains/slash",
		"contains\\backslash",
		"contains:colon",
		"CON",
		"con",
		"prn",
		"aux",
		"nul",
		"com1",
		"lpt9",
		"ends-with-dot.",
		"ends-with-space ",
		strings.Repeat("a", 129),
	}
	for _, s := range invalidSlugs {
		if err := ValidateSlug(s); err == nil {
			t.Errorf("expected %q to be rejected, got nil error", s)
		}
	}
}

func TestValidateSafePath(t *testing.T) {
	validPaths := []string{
		"research/decisions/D-001.md",
		"research/sessions/T0-01-landscape.md",
		"research/DECISIONS.md",
		".gitignore",
	}
	for _, p := range validPaths {
		if err := ValidateSafePath(p); err != nil {
			t.Errorf("expected %q to be valid, got: %v", p, err)
		}
	}

	invalidPaths := []string{
		"../escapes",
		"research/../escapes",
		"/absolute/path",
		"C:/windows/path",
		"research/con.md",
		"research/nul",
		"research/com1/foo.md",
		"research/foo.md::$DATA",
		"research/stream:name",
		"research/ends-with-dot./file.md",
	}
	for _, p := range invalidPaths {
		if err := ValidateSafePath(p); err == nil {
			t.Errorf("expected %q to be rejected, got nil error", p)
		}
	}
}
