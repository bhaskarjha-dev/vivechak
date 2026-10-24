package core

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

// SignoffDecision sets human_reviewed: true, reviewed_by, and reviewed_at on the specified decision,
// then updates the compiled DECISIONS.md registry.
func SignoffDecision(ws *store.Workspace, decisionID string, reviewer string) (string, error) {
	decisionID = strings.TrimSpace(decisionID)
	if decisionID == "" {
		return "", fmt.Errorf("decision ID is required")
	}

	// 1. Locate the ADR file in research/decisions or research/
	var targetRelPath string
	var adrData []byte

	candidates := []string{
		filepath.Join(DecisionsDir, decisionID+".md"),
		filepath.Join(ResearchDir, decisionID+".md"),
	}
	for _, cand := range candidates {
		if data, err := ws.ReadFile(cand); err == nil && len(data) > 0 {
			targetRelPath = cand
			adrData = data
			break
		}
	}

	if targetRelPath == "" {
		// Search by prefix
		searchDirs := []string{DecisionsDir, ResearchDir}
		for _, sDir := range searchDirs {
			if entries, err := ws.ListDir(sDir); err == nil {
				idUpper := strings.ToUpper(decisionID)
				for _, e := range entries {
					if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
						continue
					}
					name := strings.TrimSuffix(e.Name(), ".md")
					nameUpper := strings.ToUpper(name)
					if nameUpper == idUpper || strings.HasPrefix(nameUpper, idUpper+"-") || strings.HasPrefix(nameUpper, idUpper+"_") {
						cand := filepath.Join(sDir, e.Name())
						if data, rErr := ws.ReadFile(cand); rErr == nil && len(data) > 0 {
							targetRelPath = cand
							adrData = data
							break
						}
					}
				}
			}
			if targetRelPath != "" {
				break
			}
		}
	}

	if targetRelPath == "" {
		return "", fmt.Errorf("decision %s not found in %s or %s", decisionID, DecisionsDir, ResearchDir)
	}

	// 2. Parse frontmatter and update
	fm, body, err := ParseFrontmatter(adrData)
	if err != nil {
		return "", fmt.Errorf("parsing decision frontmatter: %w", err)
	}
	if fm == nil {
		fm = make(Frontmatter)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	fm["human_reviewed"] = true
	if reviewer != "" {
		fm["reviewed_by"] = reviewer
	}
	fm["reviewed_at"] = now

	// Render updated content
	updatedContent, err := ComposeFrontmatter(fm, body)
	if err != nil {
		return "", fmt.Errorf("composing updated frontmatter: %w", err)
	}

	// 3. Atomically write updated ADR
	if err := ws.WriteArtifact(targetRelPath, updatedContent, 0o644, true); err != nil {
		return "", fmt.Errorf("writing updated decision: %w", err)
	}

	// 4. Update DECISIONS.md registry if it exists
	decisionsPath := filepath.Join(ResearchDir, DecisionsFile)
	if ws.FileExists(decisionsPath) {
		if curDec, err := ws.ReadFile(decisionsPath); err == nil {
			// Update the row for decisionID in the table
			reRow := regexp.MustCompile(fmt.Sprintf(`(?m)^(\|[\s]*%s[\s]*\|[^|\n]*\|[^|\n]*\|[^|\n]*\|[\s]*)(false|true|no|yes)([\s]*\|.*)$`, regexp.QuoteMeta(decisionID)))
			if reRow.Match(curDec) {
				updatedRegistry := reRow.ReplaceAll(curDec, []byte("${1}true${3}"))
				_ = ws.WriteArtifact(decisionsPath, updatedRegistry, 0o644, false)
			}
		}
	}

	return targetRelPath, nil
}
