package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// NormalizeLineEndings ensures LF line endings and strips any UTF-8 BOM.
func NormalizeLineEndings(data []byte) []byte {
	// Strip UTF-8 BOM if present
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	// Normalize CRLF to LF
	return bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
}

// ExtractFrontmatterID extracts the id, session_id, or decision_id from YAML frontmatter.
func ExtractFrontmatterID(content []byte) string {
	s := string(content)
	lines := strings.Split(s, "\n")
	inFM := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !inFM && i == 0 {
				inFM = true
				continue
			}
			if inFM {
				break
			}
		}
		if inFM {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				key := strings.ToLower(strings.TrimSpace(parts[0]))
				if key == "id" || key == "session_id" || key == "decision_id" {
					val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
					if val != "" {
						return val
					}
				}
			}
		}
	}
	return ""
}

// ArchiveRevision archives prior file content into research/.history/<path>@<UTC>.md
// and caps total revisions for this file at 10.
func ArchiveRevision(ws *Workspace, relPath string, existing []byte) error {
	cleanRel := filepath.ToSlash(filepath.Clean(relPath))
	subPath := strings.TrimPrefix(cleanRel, "research/")
	subPath = strings.TrimPrefix(subPath, "/")

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	archiveRel := fmt.Sprintf("research/.history/%s@%s.md", subPath, timestamp)

	archiveDir := filepath.Dir(archiveRel)
	if err := ws.MkdirAll(archiveDir, 0o755); err != nil {
		return fmt.Errorf("creating history archive directory %q: %w", archiveDir, err)
	}

	if err := WriteFileAtomic(ws.Root(), archiveRel, existing, 0o644); err != nil {
		return fmt.Errorf("writing history archive: %w", err)
	}

	// Cap at 10 revisions
	historyDir := filepath.Dir(archiveRel)
	entries, err := ws.ListDir(historyDir)
	if err == nil {
		basePrefix := filepath.Base(subPath) + "@"
		var matching []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), basePrefix) && strings.HasSuffix(e.Name(), ".md") {
				matching = append(matching, filepath.Join(historyDir, e.Name()))
			}
		}
		if len(matching) > 10 {
			sort.Strings(matching) // oldest timestamps first
			excess := len(matching) - 10
			for i := 0; i < excess; i++ {
				_ = ws.Root().Remove(matching[i])
			}
		}
	}

	return nil
}

// WriteArtifact writes an artifact with path confinement, LF line endings, conflict detection,
// and optional revision archiving.
func (w *Workspace) WriteArtifact(relPath string, data []byte, perm os.FileMode, revise bool) error {
	// 1. Path confinement check
	if err := ValidateSafePath(relPath); err != nil {
		return err
	}
	cleanSlash := filepath.ToSlash(filepath.Clean(relPath))
	if cleanSlash != "research" && !strings.HasPrefix(cleanSlash, "research/") && cleanSlash != ".gitignore" {
		return fmt.Errorf("E-CONFINEMENT: artifact path %q must be located under research/", relPath)
	}

	// 2. LF line endings normalization
	normalized := NormalizeLineEndings(data)

	// 3. Conflict detection and revision archiving
	if w.FileExists(relPath) {
		existing, err := w.ReadFile(relPath)
		if err == nil && len(existing) > 0 {
			oldID := ExtractFrontmatterID(existing)
			newID := ExtractFrontmatterID(normalized)
			if oldID != "" && newID != "" && !strings.EqualFold(oldID, newID) {
				return fmt.Errorf("E-ID-CONFLICT: cannot overwrite existing artifact %s (id: %s) with differing id %s", relPath, oldID, newID)
			}

			if revise {
				if err := ArchiveRevision(w, relPath, existing); err != nil {
					return fmt.Errorf("archiving prior version: %w", err)
				}
			}
		}
	}

	// 4. Ensure parent directories exist
	parent := filepath.Dir(relPath)
	if parent != "." && parent != "" {
		if err := w.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", parent, err)
		}
	}

	// 5. Atomic write
	return WriteFileAtomic(w.Root(), relPath, normalized, perm)
}
