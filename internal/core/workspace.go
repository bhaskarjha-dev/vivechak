package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	if metaData, err := os.ReadFile(filepath.Join(root, MetadataFile)); err == nil {
		var meta WorkspaceMeta
		if json.Unmarshal(metaData, &meta) == nil && ValidScope(meta.Scope) {
			info.Scope = meta.Scope
		}
	}
	// Pipeline presence overrides — project scope is authoritative when pipeline exists
	if info.HasPipeline {
		info.Scope = ScopeProject
	}

	// Check for decisions
	if _, err := os.Stat(filepath.Join(root, DecisionsFile)); err == nil {
		info.HasDecisions = true
	}

	// Count sessions
	sessionsPath := filepath.Join(root, SessionsDir)
	if entries, err := os.ReadDir(sessionsPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
				info.SessionCount++
			}
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
