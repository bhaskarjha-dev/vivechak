package core

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

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

// ExtractSessionID extracts the authoritative session ID from frontmatter (session_id or id fields),
// falling back to the filename stem (without .md extension).
func ExtractSessionID(filename string, fm Frontmatter) string {
	if fm != nil {
		if sid := fm.GetString("session_id"); sid != "" {
			return sid
		}
		if id := fm.GetString("id"); id != "" {
			return id
		}
	}
	base := filepath.Base(filename)
	return strings.TrimSuffix(base, ".md")
}

// ParseFrontmatter extracts YAML frontmatter from a Markdown document.
// Returns the frontmatter (may be nil if none found) and the remaining body.
//
// Handles UTF-8 BOM, \r\n line endings, leading whitespace, and markdown code fence
// wrappers (e.g. ```yaml\n---\n...\n---\n```) robustly.
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

	trimmed := d[start:]
	hasOuterFence := false

	// Check if wrapped in code fence, e.g. ```yaml or ```
	if bytes.HasPrefix(trimmed, []byte("```yaml")) || bytes.HasPrefix(trimmed, []byte("```")) {
		nl := bytes.IndexByte(trimmed, '\n')
		if nl != -1 {
			fenceHeader := bytes.TrimRight(trimmed[:nl], " \t\r")
			if bytes.Equal(fenceHeader, []byte("```yaml")) || bytes.Equal(fenceHeader, []byte("```")) {
				afterFence := trimmed[nl+1:]
				// Check if followed by --- frontmatter
				skipSpaces := 0
				for skipSpaces < len(afterFence) && (afterFence[skipSpaces] == ' ' || afterFence[skipSpaces] == '\t' || afterFence[skipSpaces] == '\r' || afterFence[skipSpaces] == '\n') {
					skipSpaces++
				}
				if bytes.HasPrefix(afterFence[skipSpaces:], []byte("---")) {
					hasOuterFence = true
					start += nl + 1 + skipSpaces
				} else if bytes.Equal(fenceHeader, []byte("```yaml")) {
					// pure ```yaml block without --- delimiters
					closeIdx := -1
					pos := nl + 1
					for pos < len(trimmed) {
						lineStart := pos
						lineEnd := pos
						for lineEnd < len(trimmed) && trimmed[lineEnd] != '\n' {
							lineEnd++
						}
						lineTrim := bytes.TrimRight(trimmed[lineStart:lineEnd], " \t\r")
						if bytes.Equal(lineTrim, []byte("```")) {
							closeIdx = lineStart
							break
						}
						if lineEnd < len(trimmed) {
							pos = lineEnd + 1
						} else {
							break
						}
					}
					if closeIdx != -1 {
						yamlBlock := trimmed[nl+1 : closeIdx]
						fm := make(Frontmatter)
						if err := yaml.Unmarshal(yamlBlock, &fm); err == nil && len(fm) > 0 {
							bodyStart := closeIdx + 3
							for bodyStart < len(trimmed) && trimmed[bodyStart] != '\n' {
								bodyStart++
							}
							if bodyStart < len(trimmed) {
								bodyStart++
							}
							return fm, trimmed[bodyStart:], nil
						}
					}
				}
			}
		}
	}

	// Must start with "---"
	if len(d)-start < 3 || !bytes.Equal(d[start:start+3], []byte("---")) {
		return nil, data, nil
	}

	// Find end of opening "---" line
	openEnd := start + 3
	for openEnd < len(d) && d[openEnd] != '\n' {
		// If characters on the opening line after "---" are not whitespace/CR, it's not frontmatter (e.g. horizontal rule "------" or "--- Title ---")
		if d[openEnd] != ' ' && d[openEnd] != '\t' && d[openEnd] != '\r' {
			return nil, data, nil
		}
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
			break
		}
	}

	if closeLineStart == -1 {
		return nil, data, fmt.Errorf("unclosed frontmatter block")
	}

	// Extract YAML block and body
	yamlBlock := d[yamlStart:closeLineStart]
	body := d[closeLineEnd:]

	// If wrapped in outer code fence, strip closing fence line if present
	if hasOuterFence {
		bodyTrimmed := bytes.TrimLeft(body, " \t\r\n")
		if bytes.HasPrefix(bodyTrimmed, []byte("```")) {
			if nl := bytes.IndexByte(bodyTrimmed, '\n'); nl != -1 {
				body = bodyTrimmed[nl+1:]
			} else {
				body = nil
			}
		}
	}

	// Parse YAML
	fm := make(Frontmatter)
	if err := yaml.Unmarshal(yamlBlock, &fm); err != nil {
		return nil, data, fmt.Errorf("parsing YAML frontmatter: %w", err)
	}

	return fm, body, nil
}

// ComposeFrontmatter creates a Markdown document with YAML frontmatter.
func ComposeFrontmatter(fm Frontmatter, body []byte) ([]byte, error) {
	if len(fm) == 0 {
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
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case time.Time:
		return val.Format("2006-01-02")
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
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

// Set sets a key-value pair in frontmatter.
func (fm Frontmatter) Set(key string, value any) {
	if fm != nil {
		fm[key] = value
	}
}

// EnsureFrontmatterField ensures that the given key and value exist in the frontmatter.
// If the key is already present, data is returned unchanged.
// If valid frontmatter exists and the key is missing, the key-value pair is inserted
// directly before the closing delimiter ("---" or closing code fence) without re-serializing
// or altering comments, key order, and formatting.
func EnsureFrontmatterField(data []byte, key, value string) []byte {
	fm, _, err := ParseFrontmatter(data)
	if err != nil || fm == nil || fm.Has(key) {
		return data
	}

	d := data
	bomLen := 0
	if len(d) >= 3 && d[0] == 0xEF && d[1] == 0xBB && d[2] == 0xBF {
		d = d[3:]
		bomLen = 3
	}

	// Skip leading whitespace
	start := 0
	for start < len(d) && (d[start] == ' ' || d[start] == '\t' || d[start] == '\r' || d[start] == '\n') {
		start++
	}

	trimmed := d[start:]

	// Check if wrapped in code fence, e.g. ```yaml or ```
	if bytes.HasPrefix(trimmed, []byte("```yaml")) || bytes.HasPrefix(trimmed, []byte("```")) {
		nl := bytes.IndexByte(trimmed, '\n')
		if nl != -1 {
			fenceHeader := bytes.TrimRight(trimmed[:nl], " \t\r")
			if bytes.Equal(fenceHeader, []byte("```yaml")) || bytes.Equal(fenceHeader, []byte("```")) {
				afterFence := trimmed[nl+1:]
				skipSpaces := 0
				for skipSpaces < len(afterFence) && (afterFence[skipSpaces] == ' ' || afterFence[skipSpaces] == '\t' || afterFence[skipSpaces] == '\r' || afterFence[skipSpaces] == '\n') {
					skipSpaces++
				}
				if bytes.HasPrefix(afterFence[skipSpaces:], []byte("---")) {
					start += nl + 1 + skipSpaces
				} else if bytes.Equal(fenceHeader, []byte("```yaml")) {
					// pure ```yaml block without --- delimiters
					closeIdx := -1
					pos := nl + 1
					for pos < len(trimmed) {
						lineStart := pos
						lineEnd := pos
						for lineEnd < len(trimmed) && trimmed[lineEnd] != '\n' {
							lineEnd++
						}
						lineTrim := bytes.TrimRight(trimmed[lineStart:lineEnd], " \t\r")
						if bytes.Equal(lineTrim, []byte("```")) {
							closeIdx = lineStart
							break
						}
						if lineEnd < len(trimmed) {
							pos = lineEnd + 1
						} else {
							break
						}
					}
					if closeIdx != -1 {
						insertPos := bomLen + start + closeIdx
						lineBreak := "\n"
						if insertPos > 1 && data[insertPos-2] == '\r' && data[insertPos-1] == '\n' {
							lineBreak = "\r\n"
						}
						insertText := key + ": " + value + lineBreak
						if insertPos > 0 && data[insertPos-1] != '\n' {
							insertText = lineBreak + key + ": " + value + lineBreak
						}
						var res bytes.Buffer
						res.Grow(len(data) + len(insertText))
						res.Write(data[:insertPos])
						res.WriteString(insertText)
						res.Write(data[insertPos:])
						return res.Bytes()
					}
				}
			}
		}
	}

	// Must start with "---"
	if len(d)-start < 3 || !bytes.Equal(d[start:start+3], []byte("---")) {
		return data
	}

	// Find end of opening "---" line
	openEnd := start + 3
	for openEnd < len(d) && d[openEnd] != '\n' {
		if d[openEnd] != ' ' && d[openEnd] != '\t' && d[openEnd] != '\r' {
			return data
		}
		openEnd++
	}
	if openEnd < len(d) {
		openEnd++
	}

	// Scan for closing "---" on a line by itself
	yamlStart := openEnd
	closeLineStart := -1
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
			break
		}

		if lineEnd < len(d) {
			pos = lineEnd + 1
		} else {
			break
		}
	}

	if closeLineStart == -1 {
		return data
	}

	insertPos := bomLen + closeLineStart
	lineBreak := "\n"
	if insertPos > 1 && data[insertPos-2] == '\r' && data[insertPos-1] == '\n' {
		lineBreak = "\r\n"
	}
	insertText := key + ": " + value + lineBreak
	if insertPos > 0 && data[insertPos-1] != '\n' {
		insertText = lineBreak + key + ": " + value + lineBreak
	}

	var res bytes.Buffer
	res.Grow(len(data) + len(insertText))
	res.Write(data[:insertPos])
	res.WriteString(insertText)
	res.Write(data[insertPos:])
	return res.Bytes()
}
