package store

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var validSlugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_\-]*$`)

var reservedDeviceNames = map[string]bool{
	"CON":     true,
	"PRN":     true,
	"AUX":     true,
	"NUL":     true,
	"COM1":    true,
	"COM2":    true,
	"COM3":    true,
	"COM4":    true,
	"COM5":    true,
	"COM6":    true,
	"COM7":    true,
	"COM8":    true,
	"COM9":    true,
	"LPT1":    true,
	"LPT2":    true,
	"LPT3":    true,
	"LPT4":    true,
	"LPT5":    true,
	"LPT6":    true,
	"LPT7":    true,
	"LPT8":    true,
	"LPT9":    true,
	"CONIN$":  true,
	"CONOUT$": true,
	"CLOCK$":  true,
}

// IsReservedWindowsName checks if a base name (with or without extension) is a Windows device name.
func IsReservedWindowsName(name string) bool {
	upper := strings.ToUpper(strings.TrimSpace(name))
	if reservedDeviceNames[upper] {
		return true
	}
	// Check before extension: e.g., "con.txt", "NUL.md"
	if idx := strings.IndexByte(upper, '.'); idx != -1 {
		base := upper[:idx]
		if reservedDeviceNames[base] {
			return true
		}
	}
	return false
}

// ValidateSlug verifies that a slug is URL-safe and cannot be used for path traversal or device collision.
func ValidateSlug(slug string) error {
	if slug == "" {
		return nil
	}
	if len(slug) > 128 {
		return fmt.Errorf("slug exceeds maximum length of 128 characters")
	}
	if !validSlugPattern.MatchString(slug) {
		return fmt.Errorf("slug %q contains invalid characters: must start with alphanumeric and contain only alphanumeric, hyphens, or underscores", slug)
	}
	if IsReservedWindowsName(slug) {
		return fmt.Errorf("slug %q is a reserved Windows device name", slug)
	}
	if strings.HasSuffix(slug, ".") || strings.HasSuffix(slug, " ") {
		return fmt.Errorf("slug %q cannot end with a dot or space", slug)
	}
	return nil
}

// ValidateSafePath verifies that a relative path does not escape, does not contain alternate data streams,
// and does not reference Windows reserved device names.
func ValidateSafePath(relPath string) error {
	if !filepath.IsLocal(relPath) || filepath.IsAbs(relPath) {
		return fmt.Errorf("E-CONFINEMENT: path escapes workspace: %s", relPath)
	}
	// Check for any ".." path segments before or after clean
	normalizedSlash := filepath.ToSlash(relPath)
	for _, seg := range strings.Split(normalizedSlash, "/") {
		if seg == ".." {
			return fmt.Errorf("E-CONFINEMENT: path contains traversal segment '..': %s", relPath)
		}
	}
	// Disallow Alternate Data Streams (colon) and null bytes
	if strings.ContainsAny(relPath, ":\x00") {
		return fmt.Errorf("E-CONFINEMENT: path contains forbidden characters (colon or null): %s", relPath)
	}
	cleanSlash := filepath.ToSlash(filepath.Clean(relPath))
	parts := strings.Split(cleanSlash, "/")
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if IsReservedWindowsName(part) {
			return fmt.Errorf("E-CONFINEMENT: path component %q is a reserved Windows device name: %s", part, relPath)
		}
		if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return fmt.Errorf("E-CONFINEMENT: path component %q ends with dot or space: %s", part, relPath)
		}
	}
	return nil
}
