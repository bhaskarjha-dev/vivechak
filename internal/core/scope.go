// Package core contains pure business logic for Vivechak operations.
// It has NO dependency on the MCP SDK — all MCP concerns live in internal/mcp.
// This separation allows unit testing without MCP transport overhead.
package core

// Scope represents the research scope level.
// Vivechak works at three scope levels per FRAMEWORK.md §9.
type Scope string

const (
	// ScopeProject is a full research pipeline (→ FAD).
	// Uses GENERATOR.md, produces RESEARCH-PIPELINE.md + DECISIONS.md.
	ScopeProject Scope = "project"

	// ScopeDecision is a single architectural decision (→ ADR).
	// Uses GENERATOR-DECISION.md, produces 1-3 session plan.
	ScopeDecision Scope = "decision"

	// ScopeComparison is a bounded options comparison (→ WEP matrix).
	// Uses GENERATOR-COMPARISON.md, produces single session.
	ScopeComparison Scope = "comparison"
)

// ValidScope reports whether s is a recognized scope level.
func ValidScope(s Scope) bool {
	switch s {
	case ScopeProject, ScopeDecision, ScopeComparison:
		return true
	}
	return false
}

// ScopeFromString converts a string to a Scope, returning false if invalid.
func ScopeFromString(s string) (Scope, bool) {
	scope := Scope(s)
	return scope, ValidScope(scope)
}

// GeneratorFile returns the embedded generator filename for a scope.
func GeneratorFile(scope Scope) string {
	switch scope {
	case ScopeProject:
		return "GENERATOR.md"
	case ScopeDecision:
		return "GENERATOR-DECISION.md"
	case ScopeComparison:
		return "GENERATOR-COMPARISON.md"
	default:
		return ""
	}
}

// SessionPrefix returns the canonical prefix for session IDs for a given scope.
// ScopeComparison uses "C-" (e.g. C-01).
// ScopeDecision uses "D-" (e.g. D-01 or D-015-S1).
// ScopeProject uses "T" (e.g. T0-01, T1-01).
func SessionPrefix(scope Scope) string {
	switch scope {
	case ScopeComparison:
		return "C-"
	case ScopeDecision:
		return "D-"
	default:
		return "T"
	}
}
