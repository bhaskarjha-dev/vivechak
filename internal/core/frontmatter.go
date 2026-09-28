package core

import (
	"bytes"
	"fmt"

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
//
// Handles UTF-8 BOM, \r\n line endings, and leading whitespace robustly
// by tracking byte offsets in the original data rather than string indexing.
func ParseFrontmatter(data []byte) (Frontmatter, []byte, error) {
	d := data

	// Skip UTF-8 BOM if present
	if len(d) >= 3 && d[0] == 0xEF && d[1] == 0xBB && d[2] == 0xBF {
		d = d[3:]
	}

	// Skip leading whitespace
	start := 0
	for start < len(d) && (d[start] == ' ' || d[start] == '\t' || d[start] == '\r' || d[start] == '\n') {
		start++
	}

	// Must start with "---"
	if len(d)-start < 3 || !bytes.Equal(d[start:start+3], []byte("---")) {
		return nil, data, nil
	}

	// Find end of opening "---" line
	openEnd := start + 3
	for openEnd < len(d) && d[openEnd] != '\n' {
		openEnd++
	}
	if openEnd < len(d) {
		openEnd++ // skip the \n
	}

	// Scan for closing "---" on a line by itself
	yamlStart := openEnd
	closeLineStart := -1
	closeLineEnd := -1
	pos := yamlStart
	for pos < len(d) {
		lineStart := pos
		lineEnd := pos
		for lineEnd < len(d) && d[lineEnd] != '\n' {
			lineEnd++
		}

		trimmedLine := bytes.TrimRight(d[lineStart:lineEnd], " \t\r")
		if bytes.Equal(trimmedLine, []byte("---")) {
			closeLineStart = lineStart
			closeLineEnd = lineEnd
			if closeLineEnd < len(d) {
				closeLineEnd++ // include the \n
			}
			break
		}

		if lineEnd < len(d) {
			pos = lineEnd + 1
		} else {
			pos = lineEnd
			break
		}
	}

	if closeLineStart == -1 {
		return nil, data, nil
	}

	// Extract YAML block and body
	yamlBlock := d[yamlStart:closeLineStart]
	body := d[closeLineEnd:]

	// Parse YAML
	fm := make(Frontmatter)
	if err := yaml.Unmarshal(yamlBlock, &fm); err != nil {
		return nil, data, fmt.Errorf("parsing YAML frontmatter: %w", err)
	}

	return fm, body, nil
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
