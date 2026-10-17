package core

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTemplatesForScope verifies scope-aware template selection (T1-01).
func TestTemplatesForScope(t *testing.T) {
	tests := []struct {
		scope    Scope
		expected []string
	}{
		{
			scope: ScopeProject,
			expected: []string{
				"DECISIONS.template.md",
				"CONFLICT-RESOLUTION.template.md",
				"COMPARISON-SESSION.template.md",
				"FOUNDING-ARCHITECTURE.template.md",
				"PHASE-0-GATE.template.md",
				"SESSION.template.md",
			},
		},
		{
			scope: ScopeDecision,
			expected: []string{
				"DECISIONS.template.md",
				"CONFLICT-RESOLUTION.template.md",
			},
		},
		{
			scope: ScopeComparison,
			expected: []string{
				"COMPARISON-SESSION.template.md",
			},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.scope), func(t *testing.T) {
			result := TemplatesForScope(tt.scope)
			if len(result) != len(tt.expected) {
				t.Fatalf("TemplatesForScope(%q): got %d templates, want %d: %v", tt.scope, len(result), len(tt.expected), result)
			}
			for i, want := range tt.expected {
				if result[i] != want {
					t.Errorf("TemplatesForScope(%q)[%d] = %q, want %q", tt.scope, i, result[i], want)
				}
			}
		})
	}
}

// TestTemplatesForScope_ProjectEqualsAll verifies project scope returns the full set.
func TestTemplatesForScope_ProjectEqualsAll(t *testing.T) {
	all := TemplatesToCopy
	project := TemplatesForScope(ScopeProject)
	if len(project) != len(all) {
		t.Fatalf("project scope should return all %d templates, got %d", len(all), len(project))
	}
	for i, tmpl := range all {
		if project[i] != tmpl {
			t.Errorf("TemplatesForScope(project)[%d] = %q, want %q", i, project[i], tmpl)
		}
	}
}

// TestTemplatesForScope_DecisionExcludes verifies decision scope excludes project-only templates.
func TestTemplatesForScope_DecisionExcludes(t *testing.T) {
	decision := TemplatesForScope(ScopeDecision)
	excluded := map[string]bool{
		"COMPARISON-SESSION.template.md":      true,
		"FOUNDING-ARCHITECTURE.template.md":   true,
		"PHASE-0-GATE.template.md":            true,
	}
	for _, tmpl := range decision {
		if excluded[tmpl] {
			t.Errorf("decision scope should not include %q", tmpl)
		}
	}
}

// TestTemplatesForScope_ComparisonMinimal verifies comparison scope has minimum templates.
func TestTemplatesForScope_ComparisonMinimal(t *testing.T) {
	comparison := TemplatesForScope(ScopeComparison)
	if len(comparison) != 1 {
		t.Fatalf("comparison scope should have exactly 1 template, got %d: %v", len(comparison), comparison)
	}
	if comparison[0] != "COMPARISON-SESSION.template.md" {
		t.Errorf("comparison scope template = %q, want COMPARISON-SESSION.template.md", comparison[0])
	}
}

// TestTemplatesForScope_InvalidDefaultsToProject verifies unknown scope defaults to project.
func TestTemplatesForScope_InvalidDefaultsToProject(t *testing.T) {
	result := TemplatesForScope(Scope("invalid"))
	expected := TemplatesToCopy
	if len(result) != len(expected) {
		t.Fatalf("invalid scope should default to project (%d templates), got %d", len(expected), len(result))
	}
}

// TestIsSpecialResearchFile verifies detection of reserved non-ADR research files.
func TestIsSpecialResearchFile(t *testing.T) {
	specialFiles := []string{
		"DECISIONS.md",
		"decisions.md",
		"FAD.md",
		"fad.md",
		"FOUNDING-ARCHITECTURE.md",
		"founding-architecture.md",
		"PHASE-0-GATE.md",
		"phase-0-gate.md",
		"RESEARCH-PIPELINE.md",
		"research-pipeline.md",
		"D-001-plan.md",
		"my-project-plan.md",
		"D-002-comparison.md",
		"database-comparison.md",
		"D-001-conflict-resolution.md",
	}

	for _, name := range specialFiles {
		if !IsSpecialResearchFile(name) {
			t.Errorf("IsSpecialResearchFile(%q) = false, want true", name)
		}
	}

	adrFiles := []string{
		"D-001.md",
		"D-001-database-selection.md",
		"d-002-cache-layer.md",
		"custom-adr.md",
		"NOTES.md",
		"README.md",
	}

	for _, name := range adrFiles {
		if IsSpecialResearchFile(name) {
			t.Errorf("IsSpecialResearchFile(%q) = true, want false", name)
		}
	}
}

// TestInferScope verifies context-based scope auto-detection.
func TestInferScope(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Scope
	}{
		{
			name:     "Empty context defaults to project",
			input:    "",
			expected: ScopeProject,
		},
		{
			name:     "Winston vs Pino library comparison",
			input:    "Winston vs Pino for high throughput logging",
			expected: ScopeComparison,
		},
		{
			name:     "Explicit compare verb",
			input:    "compare fastify and express for new api gateway",
			expected: ScopeComparison,
		},
		{
			name:     "Which is better question",
			input:    "which is better react query or swr",
			expected: ScopeComparison,
		},
		{
			name:     "Template OPTIONS schema",
			input:    "OPTIONS: PostgreSQL, TimescaleDB, ClickHouse\nCRITERIA: write throughput",
			expected: ScopeComparison,
		},
		{
			name:     "Single decision question",
			input:    "should we use mongo or postgres for patient medical records",
			expected: ScopeDecision,
		},
		{
			name:     "Decision prefix",
			input:    "decision: graphql vs rest for mobile clients",
			expected: ScopeDecision,
		},
		{
			name:     "Template DECISION schema",
			input:    "DECISION: Datastore selection\nSTAGE: pre-build\nOPTIONS: PG, Mongo",
			expected: ScopeDecision,
		},
		{
			name:     "Full project startup pitch",
			input:    "hey plan my new startup idea. its basically uber for dog walkers. need mobile app, gps tracking, stripe payments",
			expected: ScopeProject,
		},
		{
			name:     "Full architecture design",
			input:    "system architecture for a multi-tenant B2B analytics platform",
			expected: ScopeProject,
		},
		{
			name:     "New business venture project",
			input:    "we are planning a new business venture for global freight brokerage and logistics",
			expected: ScopeProject,
		},
		{
			name:     "Market entry plan",
			input:    "market entry research pipeline for SEA healthtech telemedicine platform",
			expected: ScopeProject,
		},
		{
			name:     "Strategic business decision",
			input:    "strategic decision: build in-house sales team vs partner agency",
			expected: ScopeDecision,
		},
		{
			name:     "Trade-off between business models",
			input:    "trade-off between product-led growth and enterprise outbound sales",
			expected: ScopeDecision,
		},
		{
			name:     "AI model comparison",
			input:    "Claude 3.5 Sonnet vs GPT-4o for multi-turn autonomous coding agents",
			expected: ScopeComparison,
		},
		{
			name:     "Algorithm comparison",
			input:    "which algorithm is faster for distributed consensus in low latency networks",
			expected: ScopeComparison,
		},
		{
			name:     "Strategic option choice",
			input:    "which strategy for cross-border currency settlement and hedging",
			expected: ScopeComparison,
		},
		{
			name:     "vck compare prefix",
			input:    "vck compare winston vs pino",
			expected: ScopeComparison,
		},
		{
			name:     "vck decide prefix with colon",
			input:    "vck decide: postgres vs clickhouse for analytics",
			expected: ScopeDecision,
		},
		{
			name:     "vck plan prefix",
			input:    "vck plan new b2b payment platform",
			expected: ScopeProject,
		},
		{
			name:     "vivechak adr prefix",
			input:    "vivechak adr event streaming architecture",
			expected: ScopeDecision,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferScope(tt.input)
			if got != tt.expected {
				t.Errorf("InferScope(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestSessionIDPrefixRegex_Comparison verifies that C-01 and CMP-01 session IDs are matched.
func TestSessionIDPrefixRegex_Comparison(t *testing.T) {
	stems := []struct {
		stem     string
		expected string
	}{
		{"C-01", "C-01"},
		{"C-01-winston-vs-pino", "C-01"},
		{"c-02_logging", "c-02"},
		{"CMP-01-database", "CMP-01"},
		{"COMP-01", "COMP-01"},
		{"T1-01-database", "T1-01"},
		{"T0-01", "T0-01"},
		{"SYN-01-synthesis", "SYN-01"},
		{"D-001-S1-landscape", "D-001-S1"},
		{"FAD", "FAD"},
	}

	for _, tt := range stems {
		m := sessionIDPrefixRe.FindStringSubmatch(tt.stem)
		if len(m) < 2 {
			t.Errorf("sessionIDPrefixRe failed to match %q", tt.stem)
		} else if m[1] != tt.expected {
			t.Errorf("sessionIDPrefixRe matched %q, want %q for stem %q", m[1], tt.expected, tt.stem)
		}
	}
}

// TestInspectWorkspace_ComparisonScopeDetection verifies that InspectWorkspace detects
// comparison scope from -cmp- plan files and C- session IDs.
func TestInspectWorkspace_ComparisonScopeDetection(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, ResearchDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, SessionsDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, TemplatesDir), 0o755)

	// Add comparison template
	_ = os.WriteFile(filepath.Join(tmpDir, TemplatesDir, "COMPARISON-SESSION.template.md"), []byte("# Template"), 0o644)

	// Add slugged comparison plan
	_ = os.WriteFile(filepath.Join(tmpDir, ResearchDir, "D-001-cmp-logging.md"), []byte("# Plan"), 0o644)

	// Add C-01 session
	_ = os.WriteFile(filepath.Join(tmpDir, SessionsDir, "C-01.md"), []byte("---\nsession_id: C-01\ntitle: Test\nstatus: complete\n---\n# Content"), 0o644)

	info := InspectWorkspace(tmpDir)
	if info.Scope != ScopeComparison {
		t.Errorf("InspectWorkspace scope = %q, want %q", info.Scope, ScopeComparison)
	}
	if !info.HasPipeline {
		t.Error("InspectWorkspace HasPipeline = false, want true")
	}
	if info.SessionCount != 1 {
		t.Errorf("InspectWorkspace SessionCount = %d, want 1", info.SessionCount)
	}
}

func TestInspectWorkspace_DecisionsFrontmatterDoorType(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, ResearchDir), 0o755)

	// File named "storage-choice.md" (does not start with D-)
	adrContent := `---
id: D-01
title: Storage Engine Choice
door_type: two-way
status: accepted
---
# Content
`
	_ = os.WriteFile(filepath.Join(tmpDir, ResearchDir, "storage-choice.md"), []byte(adrContent), 0o644)

	info := InspectWorkspace(tmpDir)
	if !info.HasDecisions {
		t.Errorf("expected InspectWorkspace to detect decision file via door_type frontmatter")
	}
}

func TestCollectGateDecisions_CaseInsensitiveDeduplication(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, ResearchDir), 0o755)

	// Standalone ADR file with lowercase id "d-01"
	adrContent := `---
id: d-01
title: Storage Engine
door_type: two-way
status: accepted
---
# Content
`
	_ = os.WriteFile(filepath.Join(tmpDir, ResearchDir, "D-01.md"), []byte(adrContent), 0o644)

	// Consolidated DECISIONS.md with uppercase ID "D-01"
	decisionsContent := `# Decisions

<!-- DECISION: D-01 -->
---
id: D-01
title: Storage Engine
door_type: two-way
status: accepted
---
# Content
<!-- /DECISION: D-01 -->
`
	_ = os.WriteFile(filepath.Join(tmpDir, DecisionsFile), []byte(decisionsContent), 0o644)

	ws := &testWorkspaceFS{root: tmpDir}
	decisions := CollectGateDecisions(ws)
	if len(decisions) != 1 {
		t.Fatalf("expected exactly 1 deduplicated decision, got %d: %+v", len(decisions), decisions)
	}
}
