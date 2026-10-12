package mcputil

import (
	"context"
	"fmt"
	"path/filepath"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/embed"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

// RunGateInput defines the input parameters for vivechak_run_gate.
type RunGateInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"root directory of the project (defaults to working directory)"`
	Verbose     bool   `json:"verbose,omitempty"       jsonschema:"if true, return full gate details instead of summary"`
}

func registerRunGate(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_run_gate",
			Title: "Run Phase 0 Gate",
			Description: "Execute the Phase 0 exit gate check (Track A + Track B). " +
				"Structural checks: pipeline exists, DAG sessions completed, templates present, " +
				"decisions recorded, FAD compiled. Quality checks: session count, evidence grades in FAD, " +
				"ADR lock status, reversal triggers, evidentiary threshold for one-way doors. " +
				"Advisory: rejected alternatives, recalled verification, evidence grade distribution. " +
				"Supports progressive disclosure via verbose parameter.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleRunGate,
	)
}

func handleRunGate(_ context.Context, _ *sdkmcp.CallToolRequest, in RunGateInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_run_gate"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err,
			"Initialize a workspace first with vivechak_init.")
	}

	if !core.WorkspaceExists(root) {
		return ErrorResult(tool, fmt.Errorf("workspace not initialized at %s", root),
			"Run vivechak_init first.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err),
			"Provide a valid workspace.")
	}
	defer ws.Close()

	info := core.InspectWorkspace(root)
	gateResult := core.EvaluateGate(ws, info, root)

	trackAData := map[string]any{
		"label":  gateResult.StructuralChecks.Label,
		"passed": gateResult.StructuralChecks.Passed,
		"score":  gateResult.StructuralChecks.Score,
	}
	trackBData := map[string]any{
		"label":  gateResult.QualityChecks.Label,
		"passed": gateResult.QualityChecks.Passed,
		"score":  gateResult.QualityChecks.Score,
	}

	if in.Verbose {
		trackAData["issues"] = gateResult.StructuralChecks.Issues
		trackBData["issues"] = gateResult.QualityChecks.Issues
	}

	data := map[string]any{
		"workspace_root":          root,
		"scope":                   info.Scope,
		"gate_status":             gateResult.Status,
		"gate_passed":             gateResult.Passed,
		"structural_checks":       trackAData,
		"quality_checks":          trackBData,
		"structural_completeness": trackAData,
		"mechanical_quality":      trackBData,
		"track_a":                 trackAData, // deprecated alias for structural_checks
		"track_b":                 trackBData, // deprecated alias for quality_checks
		"scope_note": "This gate performs automated MECHANICAL checks only (Structural Completeness and Quality Indicators). " +
			"Track A (Two-Way Door fast track) and Track B (One-Way Door rigorous gate) semantic decisions and human sign-off per PHASE-0-GATE.template.md are the host architect's responsibility.",
	}

	decisions := gateResult.Decisions
	projectName := core.DetectProjectName(ws, root)

	var tmplBytes []byte
	if d, err := ws.ReadFile(filepath.Join(core.TemplatesDir, "PHASE-0-GATE.template.md")); err == nil {
		tmplBytes = d
	} else if d, err := embed.ReadTemplate("PHASE-0-GATE.template.md"); err == nil {
		tmplBytes = d
	}

	warnings := gateResult.Warnings
	gateContent := core.RenderGateArtifact(ws, tmplBytes, root, projectName, gateResult.Status, gateResult.StructuralChecks.Passed, gateResult.QualityChecks.Passed, gateResult.StructuralChecks.Issues, gateResult.QualityChecks.Issues, decisions, gateResult.Advisories...)
	if len(gateContent) > 0 {
		_ = ws.MkdirAll(core.ResearchDir, 0o755)
		if writeErr := store.WriteFileAtomic(ws.Root(), core.GateFile, gateContent, 0o644); writeErr != nil {
			warnings = append(warnings, fmt.Sprintf("W-GATE-WRITE: failed to write gate artifact: %v", writeErr))
		} else {
			data["gate_artifact"] = core.GateFile
		}
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Phase 0 Gate: %s (Structural Completeness: %d/%d, Quality Indicators: %d/%d)", gateResult.Status, gateResult.StructuralChecks.Count, gateResult.StructuralChecks.Total, gateResult.QualityChecks.Count, gateResult.QualityChecks.Total),
		Data:     data,
		Warnings: warnings,
		NextStep: gateResult.NextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// collectGateDecisions forwards to core.CollectGateDecisions for mcp tools.
func collectGateDecisions(ws *store.Workspace) []core.GateDecisionItem {
	return core.CollectGateDecisions(ws)
}
