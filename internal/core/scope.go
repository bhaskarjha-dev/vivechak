// Package core contains pure business logic for Vivechak operations.
// It has NO dependency on the MCP SDK — all MCP concerns live in internal/mcp.
// This separation allows unit testing without MCP transport overhead.
package core

import "strings"

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

// InferScope analyzes raw user context text to detect whether it describes a bounded
// options comparison, a single architectural decision, or a full project pipeline.
// Defaults to ScopeProject if no clear comparison or decision signals are found.
func InferScope(contextText string) Scope {
	trimmed := strings.TrimSpace(contextText)
	if trimmed == "" {
		return ScopeProject
	}

	lower := strings.ToLower(trimmed)

	// Explicit schema keys from generator templates
	if strings.Contains(trimmed, "OPTIONS:") && (strings.Contains(trimmed, "CRITERIA:") || !strings.Contains(trimmed, "DECISION:")) {
		return ScopeComparison
	}
	if strings.Contains(trimmed, "DECISION:") && (strings.Contains(trimmed, "STAGE:") || strings.Contains(trimmed, "LEANING:")) {
		return ScopeDecision
	}

	// Full project / application signals (override weak comparison/decision words)
	projectSignals := []string{
		"full architecture", "entire architecture", "system design", "system architecture",
		"founding architecture", "fad", "startup idea", "new startup", "build an app",
		"build a platform", "building a platform", "complete stack", "tech stack for",
		"research pipeline", "uber for", "pipeline for",
	}
	for _, sig := range projectSignals {
		if strings.Contains(lower, sig) {
			return ScopeProject
		}
	}

	// Single decision signals: "should we use", "should i use", "decision:", "choose between"
	decisionSignals := []string{
		"should we use", "should i use", "should we adopt", "should i adopt",
		"should we pick", "should i pick", "choose between", "decision between",
		"architectural decision", "adr:", "adr for", "decision:",
	}
	for _, sig := range decisionSignals {
		if strings.Contains(lower, sig) {
			return ScopeDecision
		}
	}

	// Bounded comparison signals: "vs", "versus", "compare X and Y", "which is better", "library shootout"
	compSignals := []string{
		" vs ", " vs. ", " versus ", "compare ", "comparison between", "comparison of",
		"which is better", "which one is better", "which one should", "which library",
		"which tool", "shootout", "benchmark between",
	}
	for _, sig := range compSignals {
		if strings.Contains(lower, sig) {
			return ScopeComparison
		}
	}


	return ScopeProject
}

