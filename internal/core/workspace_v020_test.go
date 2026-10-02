package core

import (
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
