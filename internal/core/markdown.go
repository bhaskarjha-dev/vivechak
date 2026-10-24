package core

import (
	"strings"
)

// FenceTracker tracks CommonMark code fence state across lines.
// Supports both backtick (```) and tilde (~~~) fences of arbitrary length (>= 3).
type FenceTracker struct {
	inFence   bool
	fenceChar byte
	fenceLen  int
}

// ProcessLine updates the fence state for a given trimmed line.
// Returns true if the line is inside a code block (including opening and closing fence lines).
func (ft *FenceTracker) ProcessLine(trimmed string) bool {
	if !ft.inFence {
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			ft.inFence = true
			ft.fenceChar = trimmed[0]
			ft.fenceLen = 0
			for ft.fenceLen < len(trimmed) && trimmed[ft.fenceLen] == ft.fenceChar {
				ft.fenceLen++
			}
			return true
		}
		return false
	}

	// Inside fence — check for closing fence (matching char, at least opening length, trailing spaces only)
	if len(trimmed) >= ft.fenceLen && trimmed[0] == ft.fenceChar {
		currLen := 0
		for currLen < len(trimmed) && trimmed[currLen] == ft.fenceChar {
			currLen++
		}
		if currLen >= ft.fenceLen && strings.TrimSpace(trimmed[currLen:]) == "" {
			ft.inFence = false
			return true // the closing fence line itself is a fence boundary
		}
	}
	return true
}

// InFence returns whether the current state is inside a code block.
func (ft *FenceTracker) InFence() bool {
	return ft.inFence
}

// ExtractSection extracts the first matching section from body content.
// Searches for headings containing any of the keywords.
// Headings inside code blocks (``` or ~~~) are ignored to prevent premature truncation.
// Once inside the section, child headings (higher level number, e.g. ### inside ##)
// are retained as content, while sibling or parent headings (<= sectionLevel) terminate extraction.
func ExtractSection(body string, keywords ...string) string {
	lines := strings.Split(body, "\n")
	var result strings.Builder
	inSection := false
	sectionLevel := 0
	var ft FenceTracker

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track code fence boundaries
		inCodeBlock := ft.ProcessLine(trimmed)

		// Only inspect headings outside code fences
		if !inCodeBlock && strings.HasPrefix(trimmed, "#") {
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
				if strings.Contains(lower, strings.ToLower(kw)) {
					inSection = true
					sectionLevel = level
					break
				}
			}
			if inSection {
				continue // skip the primary section heading itself
			}
		}

		if inSection {
			result.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(result.String())
}

// TruncateRunes truncates a string to at most maxRunes, appending "..." if truncated,
// without slicing across multi-byte UTF-8 boundaries.
func TruncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}

