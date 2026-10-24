package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/tailscale/hujson"
)

// WorkspaceExists checks if a workspace has been initialized at the given path.
// Returns true if the research/ directory exists.
func WorkspaceExists(root string) bool {
	info, err := os.Stat(filepath.Join(root, ResearchDir))
	return err == nil && info.IsDir()
}

// IsSpecialResearchFile returns true if name is a reserved pipeline or non-ADR artifact
// in research/ (e.g. DECISIONS.md, FAD.md, FOUNDING-ARCHITECTURE.md, PHASE-0-GATE.md,
// RESEARCH-PIPELINE.md, *-plan.md, *-comparison.md, *-conflict-resolution.md).
func IsSpecialResearchFile(name string) bool {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, "-plan.md") ||
		strings.HasSuffix(lower, "-comparison.md") ||
		strings.Contains(lower, "-cmp-") ||
		strings.EqualFold(name, "COMPARISON.md") ||
		strings.HasSuffix(lower, "-conflict-resolution.md") ||
		strings.EqualFold(name, "DECISIONS.md") ||
		strings.EqualFold(name, "FAD.md") ||
		strings.EqualFold(name, "FOUNDING-ARCHITECTURE.md") ||
		strings.EqualFold(name, "PHASE-0-GATE.md") ||
		strings.EqualFold(name, "RESEARCH-PIPELINE.md") {
		return true
	}
	return false
}

// InspectWorkspace scans a workspace and returns its current state.
func InspectWorkspace(root string) WorkspaceInfo {
	info := WorkspaceInfo{Root: root}

	researchPath := filepath.Join(root, ResearchDir)
	if fi, err := os.Stat(researchPath); err == nil && fi.IsDir() {
		info.Initialized = true
	} else {
		return info
	}

	// Check for pipeline
	if _, err := os.Stat(filepath.Join(root, PipelineFile)); err == nil {
		info.HasPipeline = true
	}

	// Detect scope from metadata file first
	hasMetadataScope := false
	if metaData, err := os.ReadFile(filepath.Join(root, MetadataFile)); err == nil {
		var meta WorkspaceMeta
		if std, sErr := hujson.Standardize(metaData); sErr == nil {
			metaData = std
		}
		if json.Unmarshal(metaData, &meta) == nil && ValidScope(meta.Scope) {
			info.Scope = meta.Scope
			hasMetadataScope = true
		}
	}

	// Scan research directory and subdirectories for plan/comparison/decision files
	researchEntries, _ := os.ReadDir(researchPath)
	var hasDecisionPlan, hasComparisonPlan bool
	for _, e := range researchEntries {
		name := e.Name()
		if !e.IsDir() && strings.HasSuffix(name, ".md") {
			if strings.HasSuffix(name, "-plan.md") {
				hasDecisionPlan = true
				info.HasPipeline = true
			} else if strings.HasSuffix(name, "-comparison.md") || strings.Contains(name, "-cmp-") {
				hasComparisonPlan = true
				info.HasPipeline = true
			}
		}
	}

	// Check subdirectories
	if decEntries, err := os.ReadDir(filepath.Join(root, DecisionsDir)); err == nil && len(decEntries) > 0 {
		hasDecisionPlan = true
		info.HasDecisions = true
	}
	if compEntries, err := os.ReadDir(filepath.Join(root, ComparisonsDir)); err == nil && len(compEntries) > 0 {
		hasComparisonPlan = true
	}

	// If metadata did not provide a scope, infer scope
	if !hasMetadataScope {
		if hasDecisionPlan {
			info.Scope = ScopeDecision
		} else if hasComparisonPlan {
			info.Scope = ScopeComparison
		} else if info.HasPipeline {
			info.Scope = ScopeProject
		} else {
			if sEntries, err := os.ReadDir(filepath.Join(root, SessionsDir)); err == nil {
				for _, se := range sEntries {
					sName := se.Name()
					if strings.HasPrefix(sName, "S1-") || strings.HasPrefix(sName, "D-") {
						info.Scope = ScopeDecision
						break
					} else if strings.HasPrefix(sName, "C-") || strings.HasPrefix(sName, "C1-") || strings.HasPrefix(sName, "CMP-") || strings.HasPrefix(sName, "COMP-") {
						info.Scope = ScopeComparison
						break
					}
				}
			}
			if info.Scope == "" {
				for _, e := range researchEntries {
					name := e.Name()
					if !e.IsDir() && strings.HasSuffix(name, "-decision.md") {
						info.Scope = ScopeDecision
						break
					}
				}
			}
		}
	}

	// Check for decisions (substantive DECISIONS.md or ADR files in research/ or research/decisions/)
	if decData, err := os.ReadFile(filepath.Join(root, DecisionsFile)); err == nil && len(decData) > 0 {
		decStr := string(decData)
		if strings.Contains(decStr, "<!-- DECISION:") ||
			strings.Contains(decStr, "door_type:") ||
			strings.Contains(decStr, "| D-") ||
			strings.Contains(decStr, "| [D-") {
			info.HasDecisions = true
		}
	}
	if !info.HasDecisions {
		for _, e := range researchEntries {
			name := e.Name()
			if !e.IsDir() && strings.HasSuffix(name, ".md") && !IsSpecialResearchFile(name) {
				if strings.HasPrefix(strings.ToUpper(name), "D-") {
					info.HasDecisions = true
					break
				}
				if dData, err := os.ReadFile(filepath.Join(root, ResearchDir, name)); err == nil && len(dData) > 0 {
					if strings.Contains(string(dData), "door_type:") {
						info.HasDecisions = true
						break
					}
				}
			}
		}
	}

	// Count sessions (research sessions in sessions/ plus synthesis session FAD.md)
	sessionsPath := filepath.Join(root, SessionsDir)
	hasSynthesisInSessions := false
	if entries, err := os.ReadDir(sessionsPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
				info.SessionCount++
				stem := strings.TrimSuffix(e.Name(), ".md")
				upper := strings.ToUpper(stem)
				if upper == "FAD" || upper == "SYN" || strings.HasPrefix(upper, "SYN-") {
					hasSynthesisInSessions = true
				}
			}
		}
	}
	if !hasSynthesisInSessions {
		if _, err := os.Stat(filepath.Join(root, FADFile)); err == nil {
			info.SessionCount++
		}
	}
	if info.Scope == ScopeComparison && info.SessionCount == 0 {
		if compEntries, err := os.ReadDir(filepath.Join(root, ComparisonsDir)); err == nil {
			for _, e := range compEntries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					info.SessionCount++
				}
			}
		}
		for _, e := range researchEntries {
			name := e.Name()
			if !e.IsDir() && strings.HasSuffix(name, ".md") && (strings.HasSuffix(name, "-comparison.md") || strings.Contains(name, "-cmp-") || strings.HasPrefix(strings.ToUpper(name), "C-")) {
				info.SessionCount++
			}
		}
	}

	// Count templates (only non-hidden .md files)
	templatesPath := filepath.Join(root, TemplatesDir)
	if entries, err := os.ReadDir(templatesPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && !strings.HasPrefix(e.Name(), ".") {
				info.TemplateCount++
			}
		}
	}

	return info
}
