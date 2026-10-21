package core

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
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

	// Skip leading whitespace and HTML comments (e.g. <!-- DECISION: D-001 -->)
	start := skipLeadingWhitespaceAndComments(d)

	trimmed := d[start:]
	hasOuterFence := false

	// Check if wrapped in <details> block (e.g. registry decision entries with collapsible metadata)
	if bytes.HasPrefix(trimmed, []byte("<details>")) || bytes.HasPrefix(trimmed, []byte("<details ")) || bytes.HasPrefix(trimmed, []byte("<details\n")) || bytes.HasPrefix(trimmed, []byte("<details\r\n")) {
		detailsClose := bytes.Index(trimmed, []byte("</details>"))
		if detailsClose != -1 {
			detailsContent := trimmed[:detailsClose]
			// 1. Look for ```yaml inside <details>...</details>
			yamlStart := bytes.Index(detailsContent, []byte("```yaml"))
			if yamlStart != -1 {
				yamlStartLine := yamlStart + 7
				if nl := bytes.IndexByte(detailsContent[yamlStartLine:], '\n'); nl != -1 {
					yamlContentStart := yamlStartLine + nl + 1
					yamlEnd := bytes.Index(detailsContent[yamlContentStart:], []byte("```"))
					if yamlEnd != -1 {
						yamlBlock := detailsContent[yamlContentStart : yamlContentStart+yamlEnd]
						fm := make(Frontmatter)
						if err := yaml.Unmarshal(yamlBlock, &fm); err != nil {
							return nil, nil, fmt.Errorf("malformed frontmatter in details block: %w", err)
						}
						if len(fm) > 0 {
							fenceEnd := yamlContentStart + yamlEnd + 3
							for fenceEnd < len(detailsContent) && detailsContent[fenceEnd] != '\n' {
								fenceEnd++
							}
							if fenceEnd < len(detailsContent) {
								fenceEnd++
							}
							innerBody := bytes.TrimSpace(detailsContent[fenceEnd:])
							outerBody := []byte{}
							if detailsClose+10 < len(trimmed) {
								outerBody = bytes.TrimSpace(trimmed[detailsClose+10:])
							}
							var body []byte
							if len(innerBody) > 0 && len(outerBody) > 0 {
								body = append(innerBody, append([]byte("\n\n"), outerBody...)...)
							} else if len(innerBody) > 0 {
								body = innerBody
							} else {
								body = outerBody
							}
							return fm, body, nil
						}
					}
				}
			} else {
				// 2. Look for standard --- delimited frontmatter inside <details>...</details>
				// (e.g. <summary>...</summary> followed by --- frontmatter)
				openIdx := -1
				pos := 0
				for pos < len(detailsContent) {
					lineStart := pos
					lineEnd := pos
					for lineEnd < len(detailsContent) && detailsContent[lineEnd] != '\n' {
						lineEnd++
					}
					lineTrim := bytes.TrimRight(detailsContent[lineStart:lineEnd], " \t\r")
					if bytes.Equal(lineTrim, []byte("---")) {
						openIdx = lineStart
						break
					}
					if lineEnd < len(detailsContent) {
						pos = lineEnd + 1
					} else {
						break
					}
				}

				if openIdx != -1 {
					openEnd := openIdx + 3
					for openEnd < len(detailsContent) && detailsContent[openEnd] != '\n' {
						openEnd++
					}
					if openEnd < len(detailsContent) {
						openEnd++
					}

					closeIdx := -1
					pos = openEnd
					for pos < len(detailsContent) {
						lineStart := pos
						lineEnd := pos
						for lineEnd < len(detailsContent) && detailsContent[lineEnd] != '\n' {
							lineEnd++
						}
						lineTrim := bytes.TrimRight(detailsContent[lineStart:lineEnd], " \t\r")
						if bytes.Equal(lineTrim, []byte("---")) {
							closeIdx = lineStart
							break
						}
						if lineEnd < len(detailsContent) {
							pos = lineEnd + 1
						} else {
							break
						}
					}

					if closeIdx != -1 {
						yamlBlock := detailsContent[openEnd:closeIdx]
						fm := make(Frontmatter)
						if err := yaml.Unmarshal(yamlBlock, &fm); err != nil {
							return nil, nil, fmt.Errorf("malformed frontmatter in details block: %w", err)
						}
						if len(fm) > 0 {
							closeEnd := closeIdx + 3
							for closeEnd < len(detailsContent) && detailsContent[closeEnd] != '\n' {
								closeEnd++
							}
							if closeEnd < len(detailsContent) {
								closeEnd++
							}
							innerBody := bytes.TrimSpace(detailsContent[closeEnd:])
							outerBody := []byte{}
							if detailsClose+10 < len(trimmed) {
								outerBody = bytes.TrimSpace(trimmed[detailsClose+10:])
							}
							var body []byte
							if len(innerBody) > 0 && len(outerBody) > 0 {
								body = append(innerBody, append([]byte("\n\n"), outerBody...)...)
							} else if len(innerBody) > 0 {
								body = innerBody
							} else {
								body = outerBody
							}
							return fm, body, nil
						}
					}
				}
			}
		}
	}

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
			} else if item != nil {
				result = append(result, fmt.Sprintf("%v", item))
			}
		}
		return result
	case []string:
		return val
	}
	return nil
}

// GetBool returns a boolean value from frontmatter, or false if missing/non-boolean.
func (fm Frontmatter) GetBool(key string) bool {
	if fm == nil {
		return false
	}
	v, ok := fm[key]
	if !ok || v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return strings.EqualFold(val, "true") || val == "1" || strings.EqualFold(val, "yes")
	default:
		return false
	}
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
// SetFrontmatterField sets or updates the given key and value in YAML frontmatter.
// If the key is already present, its value is replaced in-place.
// If valid frontmatter exists and the key is missing, the key-value pair is inserted
// directly before the closing delimiter ("---" or closing code fence) without re-serializing
// or altering comments, key order, and formatting.
// If no frontmatter is found, data is returned unchanged.
func SetFrontmatterField(data []byte, key, value string) []byte {
	fm, _, err := ParseFrontmatter(data)
	if err != nil || fm == nil {
		return data
	}

	d := data
	bomLen := 0
	if len(d) >= 3 && d[0] == 0xEF && d[1] == 0xBB && d[2] == 0xBF {
		d = d[3:]
		bomLen = 3
	}

	// Skip leading whitespace and HTML comments (e.g. <!-- DECISION: D-001 -->)
	start := skipLeadingWhitespaceAndComments(d)

	trimmed := d[start:]

	// Check if wrapped in <details> block
	if bytes.HasPrefix(trimmed, []byte("<details>")) || bytes.HasPrefix(trimmed, []byte("<details ")) || bytes.HasPrefix(trimmed, []byte("<details\n")) || bytes.HasPrefix(trimmed, []byte("<details\r\n")) {
		detailsClose := bytes.Index(trimmed, []byte("</details>"))
		if detailsClose != -1 {
			detailsContent := trimmed[:detailsClose]
			yamlStart := bytes.Index(detailsContent, []byte("```yaml"))
			if yamlStart != -1 {
				yamlStartLine := yamlStart + 7
				if nl := bytes.IndexByte(detailsContent[yamlStartLine:], '\n'); nl != -1 {
					yamlContentStart := yamlStartLine + nl + 1
					yamlEnd := bytes.Index(detailsContent[yamlContentStart:], []byte("```"))
					if yamlEnd != -1 {
						blockStart := start + yamlContentStart
						blockEnd := start + yamlContentStart + yamlEnd
						return replaceOrInsertInBlock(data, d, bomLen, blockStart, blockEnd, key, value)
					}
				}
			} else {
				// 2. Standard --- delimited frontmatter inside <details>
				openIdx := -1
				pos := 0
				for pos < len(detailsContent) {
					lineStart := pos
					lineEnd := pos
					for lineEnd < len(detailsContent) && detailsContent[lineEnd] != '\n' {
						lineEnd++
					}
					lineTrim := bytes.TrimRight(detailsContent[lineStart:lineEnd], " \t\r")
					if bytes.Equal(lineTrim, []byte("---")) {
						openIdx = lineStart
						break
					}
					if lineEnd < len(detailsContent) {
						pos = lineEnd + 1
					} else {
						break
					}
				}

				if openIdx != -1 {
					openEnd := openIdx + 3
					for openEnd < len(detailsContent) && detailsContent[openEnd] != '\n' {
						openEnd++
					}
					if openEnd < len(detailsContent) {
						openEnd++
					}

					closeIdx := -1
					pos = openEnd
					for pos < len(detailsContent) {
						lineStart := pos
						lineEnd := pos
						for lineEnd < len(detailsContent) && detailsContent[lineEnd] != '\n' {
							lineEnd++
						}
						lineTrim := bytes.TrimRight(detailsContent[lineStart:lineEnd], " \t\r")
						if bytes.Equal(lineTrim, []byte("---")) {
							closeIdx = lineStart
							break
						}
						if lineEnd < len(detailsContent) {
							pos = lineEnd + 1
						} else {
							break
						}
					}

					if closeIdx != -1 {
						blockStart := start + openEnd
						blockEnd := start + closeIdx
						return replaceOrInsertInBlock(data, d, bomLen, blockStart, blockEnd, key, value)
					}
				}
			}
		}
	}

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
						blockStart := start + nl + 1
						blockEnd := start + closeIdx
						return replaceOrInsertInBlock(data, d, bomLen, blockStart, blockEnd, key, value)
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

	blockStart := yamlStart
	blockEnd := closeLineStart
	return replaceOrInsertInBlock(data, d, bomLen, blockStart, blockEnd, key, value)
}

// replaceOrInsertInBlock scans the YAML lines between blockStart and blockEnd.
// If a line defining key is found, it is replaced in-place.
// If the key is not found, key: value is inserted before blockEnd.
func replaceOrInsertInBlock(data, d []byte, bomLen, blockStart, blockEnd int, key, value string) []byte {
	pos := blockStart
	for pos < blockEnd {
		lineStart := pos
		lineEnd := pos
		for lineEnd < blockEnd && d[lineEnd] != '\n' {
			lineEnd++
		}

		lineBytes := d[lineStart:lineEnd]
		trimmedLine := bytes.TrimRight(lineBytes, " \t\r")
		if len(trimmedLine) > 0 && trimmedLine[0] != '#' {
			trimmedLeft := bytes.TrimLeft(trimmedLine, " \t")
			indent := trimmedLine[:len(trimmedLine)-len(trimmedLeft)]
			if bytes.HasPrefix(trimmedLeft, []byte(key)) {
				remainder := trimmedLeft[len(key):]
				remainderTrimmed := bytes.TrimLeft(remainder, " \t")
				if bytes.HasPrefix(remainderTrimmed, []byte(":")) {
					// Found existing key! Replace line in-place.
					lineBreak := "\n"
					if lineEnd > lineStart && d[lineEnd-1] == '\r' {
						lineBreak = "\r\n"
					}
					replacement := string(indent) + key + ": " + value + lineBreak
					var res bytes.Buffer
					res.Grow(len(data) + len(replacement))
					res.Write(data[:bomLen+lineStart])
					res.WriteString(replacement)
					endOffset := bomLen + lineEnd
					if lineEnd < len(d) && d[lineEnd] == '\n' {
						endOffset = bomLen + lineEnd + 1
					}
					res.Write(data[endOffset:])
					return res.Bytes()
				}
			}
		}

		if lineEnd < blockEnd {
			pos = lineEnd + 1
		} else {
			break
		}
	}

	// Key not found in frontmatter; insert before blockEnd
	insertPos := bomLen + blockEnd
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
	return SetFrontmatterField(data, key, value)
}

// skipLeadingWhitespaceAndComments skips leading whitespace, HTML comments (e.g. <!-- DECISION: D-001 -->),
// and optional leading markdown headings (e.g. # Document Title) before frontmatter.
func skipLeadingWhitespaceAndComments(d []byte) int {
	start := 0
	for {
		for start < len(d) && (d[start] == ' ' || d[start] == '\t' || d[start] == '\r' || d[start] == '\n') {
			start++
		}
		if bytes.HasPrefix(d[start:], []byte("<!--")) {
			endComment := bytes.Index(d[start:], []byte("-->"))
			if endComment != -1 {
				start += endComment + 3
				continue
			}
		}
		// Skip top-level Markdown title before frontmatter if followed by ---
		if start < len(d) && d[start] == '#' {
			lineEnd := bytes.IndexByte(d[start:], '\n')
			if lineEnd != -1 {
				rest := d[start+lineEnd+1:]
				trimmedRest := bytes.TrimLeft(rest, " \t\r\n")
				if bytes.HasPrefix(trimmedRest, []byte("---")) {
					start += lineEnd + 1
					continue
				}
				// Look ahead: check if subsequent heading lines lead to ---
				lines := bytes.Split(rest, []byte("\n"))
				hasFM := false
				for _, l := range lines {
					tl := bytes.TrimSpace(l)
					if len(tl) == 0 || (len(tl) > 0 && tl[0] == '#') || bytes.HasPrefix(tl, []byte("<!--")) {
						continue
					}
					if bytes.Equal(tl, []byte("---")) {
						hasFM = true
					}
					break
				}
				if hasFM {
					start += lineEnd + 1
					continue
				}
			}
		}
		break
	}
	return start
}

// ExtractYAMLField extracts a simple scalar value from a YAML block by key name.
// Handles quoted and unquoted values and strips inline comments.
func ExtractYAMLField(yamlStr string, key string) string {
	for _, line := range strings.Split(yamlStr, "\n") {
		trimmed := strings.TrimSpace(line)
		prefix := key + ":"
		if strings.HasPrefix(trimmed, prefix) {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			if idx := strings.Index(val, "#"); idx >= 0 {
				val = strings.TrimSpace(val[:idx])
			}
			val = strings.Trim(val, `"'`)
			return val
		}
	}
	return ""
}

// WrapDecisionFrontmatter transforms raw YAML frontmatter (--- delimited)
// into a collapsible <details> block with a YAML code fence inside.
// This prevents markdown renderers from treating mid-file --- as horizontal rules.
// If the content is already wrapped in a <details> block or does not contain frontmatter,
// it is returned unchanged (idempotent).
func WrapDecisionFrontmatter(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	trimmed := strings.TrimSpace(normalized)

	// If already wrapped in a details block, preserve as-is
	if strings.HasPrefix(trimmed, "<details>") || strings.HasPrefix(trimmed, "<details ") || strings.HasPrefix(trimmed, "<details\n") {
		return content
	}

	prefix := ""
	suffix := ""
	if strings.HasPrefix(trimmed, "<!-- DECISION:") {
		if endIdx := strings.Index(trimmed, "-->"); endIdx != -1 {
			prefix = trimmed[:endIdx+3] + "\n"
			trimmed = strings.TrimSpace(trimmed[endIdx+3:])
		}
	}
	if strings.HasSuffix(trimmed, "-->") {
		if startIdx := strings.LastIndex(trimmed, "<!-- /DECISION:"); startIdx != -1 {
			suffix = "\n" + trimmed[startIdx:]
			trimmed = strings.TrimSpace(trimmed[:startIdx])
		}
	}

	if strings.HasPrefix(trimmed, "<details>") || strings.HasPrefix(trimmed, "<details ") || strings.HasPrefix(trimmed, "<details\n") {
		return content
	}

	if !strings.HasPrefix(trimmed, "---") {
		return content
	}

	// Find the closing ---
	lines := strings.SplitN(trimmed, "\n", 2)
	if len(lines) < 2 {
		return content
	}
	rest := lines[1]
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return content
	}

	yamlBlock := rest[:endIdx]
	afterFrontmatter := strings.TrimSpace(rest[endIdx+4:]) // skip "\n---"

	id := ExtractYAMLField(yamlBlock, "id")
	if id == "" {
		id = ExtractYAMLField(yamlBlock, "decision_id")
	}
	title := ExtractYAMLField(yamlBlock, "title")
	doorType := ExtractYAMLField(yamlBlock, "door_type")
	status := ExtractYAMLField(yamlBlock, "status")

	summaryParts := []string{}
	if id != "" {
		summaryParts = append(summaryParts, "<strong>"+id+"</strong>")
	}
	if title != "" {
		summaryParts = append(summaryParts, title)
	}
	if doorType != "" {
		summaryParts = append(summaryParts, "<code>"+doorType+"</code>")
	}
	if status != "" {
		summaryParts = append(summaryParts, status)
	}
	summaryLine := strings.Join(summaryParts, " · ")

	var b strings.Builder
	if prefix != "" {
		b.WriteString(prefix)
	}
	b.WriteString("<details>\n")
	b.WriteString("<summary>" + summaryLine + "</summary>\n\n")
	b.WriteString("```yaml\n")
	b.WriteString(strings.TrimSpace(yamlBlock) + "\n")
	b.WriteString("```\n\n")
	b.WriteString("</details>\n")
	if afterFrontmatter != "" {
		b.WriteString("\n" + afterFrontmatter + "\n")
	}
	if suffix != "" {
		b.WriteString(suffix)
	}

	return b.String()
}

// NormalizeDecisionRegistry inspects a multi-decision registry document (e.g. DECISIONS.md)
// and ensures all decision entries have their metadata wrapped in collapsible <details> blocks
// with ```yaml code fences, rather than raw mid-file --- delimiters.
// Preambles, overview tables, headings, and non-frontmatter markdown bodies are preserved.
func NormalizeDecisionRegistry(content []byte) []byte {
	if len(bytes.TrimSpace(content)) == 0 {
		return content
	}

	str := string(content)
	normalized := strings.ReplaceAll(str, "\r\n", "\n")

	// 1. Anchored blocks: <!-- DECISION: D-NNN --> ... <!-- /DECISION: D-NNN -->
	anchoredRe := regexp.MustCompile(`(?s)<!--\s*DECISION:\s*([A-Za-z0-9_-]+)\s*-->\s*(.*?)\s*<!--\s*/DECISION:\s*([A-Za-z0-9_-]+)\s*-->`)
	matches := anchoredRe.FindAllStringSubmatchIndex(normalized, -1)

	if len(matches) > 0 {
		var out strings.Builder
		lastPos := 0
		for _, m := range matches {
			out.WriteString(normalized[lastPos:m[0]])
			id := normalized[m[2]:m[3]]
			inner := normalized[m[4]:m[5]]
			closeID := normalized[m[6]:m[7]]

			wrappedInner := inner
			if !strings.Contains(inner, "<details>") {
				wrappedInner = WrapDecisionFrontmatter(inner)
			}
			fmt.Fprintf(&out, "<!-- DECISION: %s -->\n%s\n<!-- /DECISION: %s -->", id, strings.TrimSpace(wrappedInner), closeID)
			lastPos = m[1]
		}
		out.WriteString(normalized[lastPos:])
		normalized = out.String()
	}

	// 2. Unanchored blocks: Line-based scan for --- delimiters with id:/decision_id:
	lines := strings.Split(normalized, "\n")
	type blockLoc struct {
		startLine int
		endLine   int
		fmText    string
	}
	var unanchoredBlocks []blockLoc

	for i := 0; i < len(lines); i++ {
		trLine := strings.TrimSpace(lines[i])
		if trLine != "---" {
			continue
		}

		// Check if line i is inside an anchored <!-- DECISION: ... --> block
		inAnchor := false
		for back := i - 1; back >= 0; back-- {
			tb := strings.TrimSpace(lines[back])
			if strings.HasPrefix(tb, "<!-- /DECISION:") {
				break
			}
			if strings.HasPrefix(tb, "<!-- DECISION:") {
				inAnchor = true
				break
			}
		}
		if inAnchor {
			continue
		}

		// Check if line i is inside a <details> block
		inDetails := false
		for back := i - 1; back >= 0; back-- {
			tb := strings.TrimSpace(lines[back])
			if strings.HasPrefix(tb, "</details>") {
				break
			}
			if strings.HasPrefix(tb, "<details>") || strings.HasPrefix(tb, "<details ") {
				inDetails = true
				break
			}
		}
		if inDetails {
			continue
		}

		// Look ahead for closing "---" and verify id: or decision_id: exists
		closeLine := -1
		hasID := false
		validYAML := true
		for j := i + 1; j < len(lines); j++ {
			tr := strings.TrimSpace(lines[j])
			if tr == "---" {
				closeLine = j
				break
			}
			if strings.HasPrefix(tr, "<") || strings.HasPrefix(tr, "```") || strings.HasPrefix(tr, "# ") || strings.HasPrefix(tr, "## ") || strings.HasPrefix(tr, "### ") {
				validYAML = false
				break
			}
			if strings.HasPrefix(tr, "id:") || strings.HasPrefix(tr, "decision_id:") {
				hasID = true
			}
		}

		if !validYAML || closeLine == -1 || !hasID {
			continue
		}
		fmText := strings.Join(lines[i:closeLine+1], "\n")
		id := ExtractYAMLField(fmText, "id")
		if id == "" {
			id = ExtractYAMLField(fmText, "decision_id")
		}
		if id != "" && !strings.HasPrefix(strings.ToUpper(id), "CHK-") {
			unanchoredBlocks = append(unanchoredBlocks, blockLoc{
				startLine: i,
				endLine:   closeLine,
				fmText:    fmText,
			})
			i = closeLine // advance past this frontmatter block
		}
	}

	if len(unanchoredBlocks) > 0 {
		var outLines []string
		curr := 0
		for _, b := range unanchoredBlocks {
			outLines = append(outLines, lines[curr:b.startLine]...)
			wrapped := WrapDecisionFrontmatter(b.fmText)
			outLines = append(outLines, strings.TrimSpace(wrapped))
			curr = b.endLine + 1
		}
		outLines = append(outLines, lines[curr:]...)
		normalized = strings.Join(outLines, "\n")
	}

	if strings.Contains(str, "\r\n") {
		normalized = strings.ReplaceAll(normalized, "\n", "\r\n")
	}
	return []byte(normalized)
}

