package mcputil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StatusInput holds the arguments for vivechak_status.
type StatusInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path"`
}

func registerStatus(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_status",
			Title: "Vivechak Status",
			Description: "Scan workspace and report research progress. Shows workspace state, " +
				"session counts, decision status, and scope detection. Use this tool first to orient " +
				"before any other operation. Read-only — does not modify the workspace. " +
				"Always available for agent orientation.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleStatus,
	)
}

func handleStatus(_ context.Context, _ *sdkmcp.CallToolRequest, in StatusInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_status"

	// Resolve workspace (don't error if not found — just report)
	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		env := Envelope{
			Success: true,
			Message: "No workspace found. Vivechak is ready to initialize a new workspace.",
			Data: map[string]any{
				"initialized": false,
				"error":       err.Error(),
			},
			NextStep: "Run vivechak_init with your project directory to create a workspace, then " +
				"vivechak_prepare_generator to get the generator prompt.",
			Meta: NewMeta(tool),
		}
		return env.ToResult()
	}

	// Inspect workspace
	info := core.InspectWorkspace(root)

	var nextStep string
	if !info.Initialized {
		nextStep = fmt.Sprintf("Run vivechak_init at %s to create the workspace.", root)
	} else if !info.HasPipeline && info.SessionCount == 0 {
		nextStep = "Run vivechak_prepare_generator to get the generator prompt, execute it, " +
			"then save the output with vivechak_save_plan."
	} else if info.SessionCount == 0 {
		nextStep = "Run vivechak_next_session to get the first research session prompt."
	} else {
		hasFAD := false
		if _, err := os.Stat(filepath.Join(root, core.FADFile)); err == nil {
			hasFAD = true
		}

		if hasFAD {
			nextStep = "Research synthesis is complete (FAD.md exists). Run vivechak_run_gate to verify the Phase 0 exit gate before starting codebase construction."
		} else if info.Scope == core.ScopeComparison && info.SessionCount >= 1 {
			nextStep = "Comparison research session is complete. Run vivechak_run_gate to verify the exit gate."
		} else if info.Scope == core.ScopeDecision && info.HasDecisions && info.SessionCount >= 1 {
			nextStep = "Decision research and ADR are complete. Run vivechak_run_gate to verify the exit gate, or vivechak_next_session if more sessions remain."
		} else {
			nextStep = "Run vivechak_next_session for the next actionable session, or " +
				"vivechak_record_decision to record decisions from completed sessions."
		}
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Workspace at %s: %d sessions, %d templates", root, info.SessionCount, info.TemplateCount),
		Data:     info,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}
