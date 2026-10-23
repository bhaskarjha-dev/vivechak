package mcputil

import (
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
