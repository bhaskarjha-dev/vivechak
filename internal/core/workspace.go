package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceLayout describes the directory structure of a Vivechak research workspace.
const (
	// ResearchDir is the root research directory within a project.
	ResearchDir = "research"
	// SessionsDir holds individual research session outputs.
	SessionsDir = "research/sessions"
	// TemplatesDir holds the copied template files.
	TemplatesDir = "research/templates"

	// PipelineFile is the research pipeline DAG (project scope).
	PipelineFile = "research/RESEARCH-PIPELINE.md"
	// DecisionsFile is the decision registry.
	DecisionsFile = "research/DECISIONS.md"
	// FADFile is the Founding Architecture Document (terminal synthesis).
	FADFile = "research/FAD.md"
	// GateFile is the Phase 0 Exit Gate record.
	GateFile = "research/PHASE-0-GATE.md"
	// MetadataFile stores workspace metadata (scope, init timestamp).
	MetadataFile = "research/.vivechak.json"
)

// WorkspaceMeta holds metadata written during vivechak_init.
type WorkspaceMeta struct {
	Scope     Scope  `json:"scope"`
	CreatedAt string `json:"created_at,omitempty"`
}

// TemplatesToCopy lists the template files that vivechak_init copies
// into the workspace's research/templates/ directory.
var TemplatesToCopy = []string{
	"DECISIONS.template.md",
	"CONFLICT-RESOLUTION.template.md",
	"COMPARISON-SESSION.template.md",
	"FOUNDING-ARCHITECTURE.template.md",
	"PHASE-0-GATE.template.md",
	"SESSION.template.md",
}

// TemplatesForScope returns the template files appropriate for the given scope.
func TemplatesForScope(scope Scope) []string {
	switch scope {
	case ScopeDecision:
		return []string{
			"DECISIONS.template.md",
			"CONFLICT-RESOLUTION.template.md",
		}
	case ScopeComparison:
		return []string{
			"COMPARISON-SESSION.template.md",
		}
	default: // ScopeProject
		return TemplatesToCopy // all 6
	}
}

// IsSpecialResearchFile returns true if name is a reserved pipeline or non-ADR artifact
// in research/ (e.g. DECISIONS.md, FAD.md, FOUNDING-ARCHITECTURE.md, PHASE-0-GATE.md,
// RESEARCH-PIPELINE.md, *-plan.md, *-comparison.md, *-conflict-resolution.md).
func IsSpecialResearchFile(name string) bool {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, "-plan.md") ||
		strings.HasSuffix(lower, "-comparison.md") ||
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

// ResolveWorkspace implements the 4-step workspace resolution chain
// per FINAL-PLAN.md:
//
//  1. Explicit project_root argument (from tool call)
//  2. VIVECHAK_PROJECT_ROOT environment variable
//  3. Discover research/ directory walking up from CWD
//  4. Error
//
// Returns the absolute path to the workspace root.
func ResolveWorkspace(explicit string) (string, error) {
	// Step 1: Explicit argument
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", fmt.Errorf("resolving explicit path %q: %w", explicit, err)
		}
		// If explicit path points directly to research directory, normalize to parent workspace root
		if filepath.Base(abs) == ResearchDir {
			if _, err := os.Stat(filepath.Join(abs, ResearchDir)); os.IsNotExist(err) {
				if fi, err := os.Stat(abs); err == nil && fi.IsDir() {
					abs = filepath.Dir(abs)
				}
			}
		}
		return abs, nil
	}

	// Step 2: Environment variable
	if v := os.Getenv("VIVECHAK_PROJECT_ROOT"); v != "" {
		abs, err := filepath.Abs(v)
		if err != nil {
			return "", fmt.Errorf("resolving VIVECHAK_PROJECT_ROOT %q: %w", v, err)
		}
		return abs, nil
	}

	// Step 3: Walk up from CWD looking for research/ directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	dir := cwd
	for {
		researchPath := filepath.Join(dir, ResearchDir)
		if info, err := os.Stat(researchPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding research/
			break
		}
		dir = parent
	}

	// Step 4: Error
	return "", fmt.Errorf(
		"could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, "+
			"and no research/ directory found walking up from %s. "+
			"Run vivechak_init first or pass project_root explicitly", cwd)
}

// WorkspaceExists checks if a workspace has been initialized at the given path.
// Returns true if the research/ directory exists.
func WorkspaceExists(root string) bool {
	info, err := os.Stat(filepath.Join(root, ResearchDir))
	return err == nil && info.IsDir()
}

// WorkspaceInfo describes the current state of a Vivechak workspace.
type WorkspaceInfo struct {
	// Root is the absolute path to the workspace root.
	Root string `json:"root"`

	// Initialized reports whether research/ directory exists.
	Initialized bool `json:"initialized"`

	// HasPipeline reports whether RESEARCH-PIPELINE.md exists (project scope).
	HasPipeline bool `json:"has_pipeline,omitempty"`

	// HasDecisions reports whether DECISIONS.md exists.
	HasDecisions bool `json:"has_decisions,omitempty"`

	// SessionCount is the number of session files in research/sessions/.
	SessionCount int `json:"session_count"`

	// TemplateCount is the number of template files in research/templates/.
	TemplateCount int `json:"template_count"`

	// Scope is the detected scope (empty if not determinable).
	Scope Scope `json:"scope,omitempty"`
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

	// Try to detect scope from metadata file first
	hasMetadataScope := false
	if metaData, err := os.ReadFile(filepath.Join(root, MetadataFile)); err == nil {
		var meta WorkspaceMeta
		if json.Unmarshal(metaData, &meta) == nil && ValidScope(meta.Scope) {
			info.Scope = meta.Scope
			hasMetadataScope = true
		}
	}
	// If metadata did not provide a scope, infer scope from pipeline presence or manual workspace files
	if !hasMetadataScope {
		if info.HasPipeline {
			info.Scope = ScopeProject
		} else if entries, err := os.ReadDir(researchPath); err == nil {
			hasDecisionPlan := false
			hasComparisonPlan := false
			for _, e := range entries {
				name := e.Name()
				if !e.IsDir() && strings.HasSuffix(name, ".md") {
					if strings.HasSuffix(name, "-plan.md") {
						hasDecisionPlan = true
						info.HasPipeline = true
					} else if strings.HasSuffix(name, "-comparison.md") {
						hasComparisonPlan = true
						info.HasPipeline = true
					}
				}
			}
			if hasDecisionPlan {
				info.Scope = ScopeDecision
			} else if hasComparisonPlan {
				info.Scope = ScopeComparison
			} else {
				// Check sessions directory for session ID prefixes if still undetermined
				if sEntries, err := os.ReadDir(filepath.Join(root, SessionsDir)); err == nil {
					for _, se := range sEntries {
						sName := se.Name()
						if strings.HasPrefix(sName, "S1-") || strings.HasPrefix(sName, "D-") {
							info.Scope = ScopeDecision
							break
						} else if strings.HasPrefix(sName, "C1-") || strings.HasPrefix(sName, "COMP-") {
							info.Scope = ScopeComparison
							break
						}
					}
				}
				if info.Scope == "" {
					for _, e := range entries {
						name := e.Name()
						if !e.IsDir() && strings.HasSuffix(name, "-decision.md") {
							info.Scope = ScopeDecision
							break
						}
					}
				}
			}
		}
	}

	// Check for decisions (non-empty DECISIONS.md or ADR files in research/)
	if fi, err := os.Stat(filepath.Join(root, DecisionsFile)); err == nil && fi.Size() > 0 {
		info.HasDecisions = true
	} else if entries, err := os.ReadDir(researchPath); err == nil {
		for _, e := range entries {
			name := e.Name()
			if !e.IsDir() && strings.HasSuffix(name, ".md") && !IsSpecialResearchFile(name) && strings.HasPrefix(strings.ToUpper(name), "D-") {
				info.HasDecisions = true
				break
			}
		}
	}

	// Count sessions (research sessions in sessions/ plus synthesis session FAD.md if present and not in sessions/)
	sessionsPath := filepath.Join(root, SessionsDir)
	hasSynthesisInSessions := false
	if entries, err := os.ReadDir(sessionsPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
				info.SessionCount++
				stem := strings.TrimSuffix(e.Name(), ".md")
				if IsSynthesisSession(stem) {
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

	// Count templates
	templatesPath := filepath.Join(root, TemplatesDir)
	if entries, err := os.ReadDir(templatesPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				info.TemplateCount++
			}
		}
	}

	return info
}
