package workspace

// Scope represents the resolution breadth of a Vivechak research engagement.
type Scope string

const (
	// ScopeComparison: Rapid evaluation between 2–5 options (WEP matrix, 1 session).
	ScopeComparison Scope = "comparison"
	// ScopeDecision: Architectural dilemma in an active codebase (ADR, 1–3 sessions).
	ScopeDecision Scope = "decision"
	// ScopeProject: Greenfield architecture or major rewrite (FAD, DAG sessions, Phase 0 gate).
	ScopeProject Scope = "project"
)

// ValidScope reports whether s is one of the three recognized Vivechak scopes.
func ValidScope(s Scope) bool {
	switch s {
	case ScopeComparison, ScopeDecision, ScopeProject:
		return true
	default:
		return false
	}
}

// WorkspaceLayout constants define standard directory structure and file paths.
const (
	ResearchDir    = "research"
	SessionsDir    = "research/sessions"
	ComparisonsDir = "research/comparisons"
	DecisionsDir   = "research/decisions"
	TemplatesDir   = "research/templates"
	HistoryDir     = "research/.history"
	LocksDir       = "research/.locks"

	PipelineFile  = "research/RESEARCH-PIPELINE.md"
	DecisionsFile = "research/DECISIONS.md"
	FADFile       = "research/FAD.md"
	GateFile      = "research/PHASE-0-GATE.md"
	MetadataFile  = "research/.vivechak.json"
)

// WorkspaceMeta holds metadata written to research/.vivechak.json.
type WorkspaceMeta struct {
	Schema    int            `json:"schema,omitempty"`
	Scope     Scope          `json:"scope"`
	CreatedAt string         `json:"created_at,omitempty"`
	Counters  map[string]int `json:"counters,omitempty"`
}

// TemplatesToCopy lists all 6 standard templates for project scope.
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
		return TemplatesToCopy
	}
}

// WorkspaceInfo describes the current state of a Vivechak workspace.
type WorkspaceInfo struct {
	Root          string `json:"root"`
	Initialized   bool   `json:"initialized"`
	HasPipeline   bool   `json:"has_pipeline,omitempty"`
	HasDecisions  bool   `json:"has_decisions,omitempty"`
	SessionCount  int    `json:"session_count"`
	TemplateCount int    `json:"template_count"`
	Scope         Scope  `json:"scope,omitempty"`
}
