package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/embed"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// GateInput defines the input parameters for vivechak_gate.
type GateInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"root directory of the project (defaults to working directory)"`
	Verbose     bool   `json:"verbose,omitempty"       jsonschema:"if true, return full gate details instead of summary"`
	Track       string `json:"track,omitempty"         jsonschema:"optional track filter ('track_a', 'track_b', or 'all')"`
}

// RunGateInput provides backward compatibility for unit tests calling handleRunGate.
type RunGateInput = GateInput

func registerGate(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_gate",
			Title: "Run Phase 0 Exit Gate",
			Description: "Mechanically evaluate Phase 0 pre-coding exit criteria across Track A (Two-Way Doors) and Track B (One-Way Doors). " +
				"Evaluates structural pipeline completeness, FAD compilation, evidence grades, ADR lock status, and adversarial stress-testing. " +
				"Persists research/PHASE-0-GATE.md preserving human sign-offs. " +
				"Must PASS before codebase implementation begins.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleGate,
	)
}

func handleGate(_ context.Context, _ *sdkmcp.CallToolRequest, in GateInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_gate"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start.")
	}

	if !core.WorkspaceExists(root) {
		return ErrorResult(tool, fmt.Errorf("workspace not initialized at %s", root), "Run vivechak_start first.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err), "Provide a valid workspace.")
	}
	defer ws.Close()

	trackFilter := strings.ToLower(strings.TrimSpace(in.Track))
	if trackFilter == "" {
		trackFilter = "all"
	}
	if trackFilter != "track_a" && trackFilter != "track_b" && trackFilter != "all" {
		return ErrorResult(tool, fmt.Errorf("invalid track filter %q: must be 'track_a', 'track_b', or 'all'", in.Track),
			"Specify 'track_a', 'track_b', or omit for all.")
	}

	info := core.InspectWorkspace(root)
	gateResult := core.EvaluateGate(ws, info, root)

	structuralData := map[string]any{
		"passed": gateResult.StructuralChecks.Passed,
		"score":  gateResult.StructuralChecks.Score,
		"count":  gateResult.StructuralChecks.Count,
		"total":  gateResult.StructuralChecks.Total,
	}

	trackAData := map[string]any{
		"passed": gateResult.TrackAChecks.Passed,
		"score":  gateResult.TrackAChecks.Score,
		"count":  gateResult.TrackAChecks.Count,
		"total":  gateResult.TrackAChecks.Total,
	}

	trackBData := map[string]any{
		"passed": gateResult.TrackBChecks.Passed,
		"score":  gateResult.TrackBChecks.Score,
		"count":  gateResult.TrackBChecks.Count,
		"total":  gateResult.TrackBChecks.Total,
	}

	if in.Verbose {
		trackAData["issues"] = gateResult.TrackAChecks.Issues
		trackBData["issues"] = gateResult.TrackBChecks.Issues
		structuralData["issues"] = gateResult.StructuralChecks.Issues
	}

	trackAResult := "FAIL"
	if gateResult.TrackAChecks.Passed {
		trackAResult = "PASS"
	}
	trackBResult := "FAIL"
	if gateResult.TrackBChecks.Passed {
		trackBResult = "PASS"
	}

	gatePassed := gateResult.Passed
	gateStatus := gateResult.Status
	switch trackFilter {
	case "track_a":
		gatePassed = gateResult.TrackAChecks.Passed
		if gatePassed {
			gateStatus = "PASS"
		} else {
			gateStatus = "FAIL"
		}
	case "track_b":
		gatePassed = gateResult.TrackBChecks.Passed
		if gatePassed {
			gateStatus = "PASS"
		} else {
			gateStatus = "FAIL"
		}
	}

	data := map[string]any{
		"workspace_root":          root,
		"scope":                   info.Scope,
		"gate_status":             gateStatus,
		"gate_passed":             gatePassed,
		"track_filter":            trackFilter,
		"track_a_result":          trackAResult,
		"track_b_result":          trackBResult,
		"structural_checks":       structuralData,
		"quality_checks":          trackBData,
		"track_a_checks":          trackAData,
		"track_b_checks":          trackBData,
		"structural_completeness": structuralData,
		"mechanical_quality":      trackBData,
		"track_a":                 trackAData,
		"track_b":                 trackBData,
		"scope_note": "This gate performs automated MECHANICAL checks only (Track A Two-Way Doors, Track B One-Way Doors, and Structural Completeness). " +
			"Semantic decisions and human sign-off per PHASE-0-GATE.template.md are the host architect's responsibility.",
	}

	decisions := gateResult.Decisions
	data["decisions"] = decisions
	projectName := core.DetectProjectName(ws, root)

	var tmplBytes []byte
	if d, err := ws.ReadFile(filepath.Join(core.TemplatesDir, "PHASE-0-GATE.template.md")); err == nil {
		tmplBytes = d
	} else if d, err := embed.ReadTemplate("PHASE-0-GATE.template.md"); err == nil {
		tmplBytes = d
	}

	warnings := gateResult.Warnings
	allAdvisories := append(append([]string{}, gateResult.Advisories...), warnings...)
	gateContent := core.RenderGateArtifact(ws, tmplBytes, root, projectName, gateResult.Status, gateResult.TrackAChecks.Passed, gateResult.TrackBChecks.Passed, gateResult.TrackAChecks.Issues, gateResult.TrackBChecks.Issues, decisions, allAdvisories...)
	if len(gateContent) > 0 {
		var existingGate []byte
		if d, err := ws.ReadFile(core.GateFile); err == nil {
			existingGate = d
		}
		mergedContent := core.MergeGateArtifactContent(existingGate, gateContent)
		_ = ws.MkdirAll(core.ResearchDir, 0o755)
		if writeErr := store.WriteFileAtomic(ws.Root(), core.GateFile, mergedContent, 0o644); writeErr != nil {
			warnings = append(warnings, fmt.Sprintf("W-GATE-WRITE: failed to write gate artifact: %v", writeErr))
		} else {
			data["gate_artifact"] = core.GateFile
		}
	}

	if !gatePassed {
		var blockingIssues []string
		failureScope := "Track B"
		switch trackFilter {
		case "track_a":
			blockingIssues = append(blockingIssues, gateResult.TrackAChecks.Issues...)
			failureScope = "Track A"
		case "track_b":
			blockingIssues = append(blockingIssues, gateResult.TrackBChecks.Issues...)
			failureScope = "Track B"
		default:
			blockingIssues = append(blockingIssues, gateResult.StructuralChecks.Issues...)
			blockingIssues = append(blockingIssues, gateResult.TrackAChecks.Issues...)
			blockingIssues = append(blockingIssues, gateResult.TrackBChecks.Issues...)
			if len(gateResult.StructuralChecks.Issues) > 0 && len(gateResult.TrackBChecks.Issues) > 0 {
				failureScope = "Structural and Quality"
			} else if len(gateResult.StructuralChecks.Issues) > 0 {
				failureScope = "Structural Completeness"
			} else if len(gateResult.TrackBChecks.Issues) > 0 {
				failureScope = "Track B Quality"
			} else if len(gateResult.TrackAChecks.Issues) > 0 {
				failureScope = "Track A Two-Way Door"
			}
		}
		issueCount := len(blockingIssues)
		env := Envelope{
			Ok:      true,
			Success: true,
			Message: fmt.Sprintf("❌ PHASE 0 GATE FAILED: %s requirements not met (%d blocking issue%s). Implementation is strictly prohibited.",
				failureScope, issueCount, plural(issueCount)),
			Data:     data,
			Warnings: warnings,
			Next: &NextAction{
				Why: fmt.Sprintf("⛔ DO NOT PROCEED TO CODE. Resolve %d blocking issue(s): %s", issueCount, strings.Join(blockingIssues, "; ")),
			},
			NextStep: fmt.Sprintf("⛔ DO NOT PROCEED TO CODE. Phase 0 exit criteria NOT MET. "+
				"Resolve the %d blocking issue(s) before code scaffolding: %s",
				issueCount, strings.Join(blockingIssues, "; ")),
			Meta: NewMeta(tool),
		}
		return env.ToResult()
	}

	next := &NextAction{
		Why: "Phase 0 Gate PASSED. All Track A and Track B exit criteria satisfied. Ready for codebase implementation.",
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Phase 0 Gate: %s (Structural Completeness: %d/%d, Quality Indicators: %d/%d)", gateStatus, gateResult.StructuralChecks.Count, gateResult.StructuralChecks.Total, gateResult.QualityChecks.Count, gateResult.QualityChecks.Total),
		Data:     data,
		Warnings: warnings,
		Next:     next,
		NextStep: gateResult.NextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// handleRunGate provides backward compatibility for unit tests calling handleRunGate.
func handleRunGate(ctx context.Context, req *sdkmcp.CallToolRequest, in RunGateInput) (*sdkmcp.CallToolResult, Envelope, error) {
	return handleGate(ctx, req, in)
}
