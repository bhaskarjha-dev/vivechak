package core

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents the YAML frontmatter block at the top of
// Vivechak research artifacts (sessions, decisions, pipelines).
//
// Format:
//
//	---
//	key: value
//	---
//	<markdown body>
type Frontmatter map[string]any

// ParseFrontmatter extracts YAML frontmatter from a Markdown document.
// Returns the frontmatter (may be nil if none found) and the remaining body.
func ParseFrontmatter(data []byte) (Frontmatter, []byte, error) {
	content := string(data)

	// Must start with "---"
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return nil, data, nil
	}

	// Find opening delimiter
	trimmed := strings.TrimSpace(content)
	rest := trimmed[3:] // skip first "---"

	// Find closing delimiter — must be "---" on a line by itself
	endIdx := -1
	lines := strings.Split(rest, "\n")
	charCount := 0
	for i, line := range lines {
		if i > 0 && strings.TrimSpace(line) == "---" {
			endIdx = charCount
			break
		}
		charCount += len(line) + 1 // +1 for the \n
	}
	if endIdx == -1 {
		return nil, data, nil
	}

	yamlBlock := rest[:endIdx]
	bodyStart := strings.Index(content, rest[endIdx:])

	// Find the actual body start (skip the closing --- line)
	remaining := content[bodyStart:]
	if idx := strings.Index(remaining, "\n"); idx >= 0 {
		remaining = remaining[idx+1:]
	}

	// Parse YAML
	fm := make(Frontmatter)
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, data, fmt.Errorf("parsing YAML frontmatter: %w", err)
	}

	return fm, []byte(remaining), nil
}

// ComposeFrontmatter creates a Markdown document with YAML frontmatter.
func ComposeFrontmatter(fm Frontmatter, body []byte) ([]byte, error) {
	if fm == nil || len(fm) == 0 {
		return body, nil
	}

	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("marshaling frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	buf.WriteString("---\n")
	buf.Write(body)

	return buf.Bytes(), nil
}

// GetString returns a string value from frontmatter, or empty string if missing.
func (fm Frontmatter) GetString(key string) string {
	if fm == nil {
		return ""
	}
	v, ok := fm[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// GetStringSlice returns a string slice from frontmatter.
func (fm Frontmatter) GetStringSlice(key string) []string {
	if fm == nil {
		return nil
	}
	v, ok := fm[key]
	if !ok {
		return nil
	}
	switch val := v.(type) {
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return val
	}
	return nil
}

// Has reports whether the frontmatter contains the given key.
func (fm Frontmatter) Has(key string) bool {
	if fm == nil {
		return false
	}
	_, ok := fm[key]
	return ok
}
