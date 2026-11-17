package mcputil

import (
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

// scanCompletedSessions forwards to core.ScanCompletedSessions.
func scanCompletedSessions(ws *store.Workspace, knownIDs ...string) (map[string]bool, error) {
	return core.ScanCompletedSessions(ws, knownIDs...)
}

// collectGateDecisions forwards to core.CollectGateDecisions for mcp tools.
func collectGateDecisions(ws *store.Workspace) []core.GateDecisionItem {
	return core.CollectGateDecisions(ws)
}

// planFixOr returns the fix sentence of a V-PLAN-* plan error, or fallback for any other error.
func planFixOr(err error, fallback string) string {
	if fix := core.PlanFix(err); fix != "" {
		return fix
	}
	return fallback
}

// readPipelineDAG parses research/RESEARCH-PIPELINE.md, or returns nil when there is none or it can't be used.
func readPipelineDAG(ws *store.Workspace) *core.DAG {
	data, err := ws.ReadFile(core.PipelineFile)
	if err != nil {
		return nil
	}
	dag, err := core.ParsePipeline(data)
	if err != nil {
		return nil
	}
	return dag
}

// isSynthesisTarget reports whether id names the synthesis session, whose output is research/FAD.md: the
// pipeline's kind: synthesis session, or the id research/FAD.md itself carries.
func isSynthesisTarget(ws *store.Workspace, id string) bool {
	if readPipelineDAG(ws).IsSynthesisID(id) {
		return true
	}
	data, err := ws.ReadFile(core.FADFile)
	if err != nil {
		return false
	}
	fm, _, _ := core.ParseFrontmatter(data)
	id = strings.TrimSpace(id)
	return fm != nil && id != "" &&
		(strings.EqualFold(strings.TrimSpace(fm.GetString("id")), id) || strings.EqualFold(strings.TrimSpace(fm.GetString("session_id")), id))
}
