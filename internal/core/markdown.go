package core

import (
	"strings"
)

// ExtractSection extracts the first matching section from body content.
// Searches for headings containing any of the keywords.
// Once inside the section, child headings (higher level number, e.g. ### inside ##)
// are retained as content, while sibling or parent headings (<= sectionLevel) terminate extraction.
func ExtractSection(body string, keywords ...string) string {
	lines := strings.Split(body, "\n")
	var result strings.Builder
	inSection := false
	sectionLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}

			if inSection {
				if level <= sectionLevel {
					// Hit a same-level or higher heading — stop
					break
				}
				// Child heading inside active section — retain it in output
				if trimmed != "" {
					result.WriteString(line + "\n")
				}
				continue
			}

			for _, kw := range keywords {
				if strings.Contains(lower, kw) {
					inSection = true
					sectionLevel = level
					break
				}
			}
			if inSection {
				continue // skip the primary section heading itself
			}
		}

		if inSection && trimmed != "" {
			result.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(result.String())
}

// extractSection is an unexported package alias for ExtractSection.
func extractSection(body string, keywords ...string) string {
	return ExtractSection(body, keywords...)
}
