package mcputil

import (
	"context"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
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
				IdempotentHint:  false,
				DestructiveHint: BoolPtr(true),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleGate,
	)
}

func handleGate(ctx context.Context, req *sdkmcp.CallToolRequest, in GateInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_gate"

	var roots []string
	if strings.TrimSpace(in.ProjectRoot) == "" {
		roots = getClientRoots(ctx, req)
	}
	root, rule, err := core.ResolveWorkspaceWithRoots(ctx, in.ProjectRoot, roots)
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

	pass := gateResult.Passed
	verdict := gateResult.Verdict
	switch trackFilter {
	case "track_a":
		pass = gateResult.TrackAChecks.Passed
		if pass {
			verdict = "PASS"
		} else {
			verdict = "FAIL"
		}
	case "track_b":
		pass = gateResult.TrackBChecks.Passed
		if pass {
			verdict = "PASS"
		} else {
			verdict = "FAIL"
		}
	}

	trackAResult := "FAIL"
	if gateResult.TrackAChecks.Passed {
		trackAResult = "PASS"
	}
	trackBResult := "FAIL"
	if gateResult.TrackBChecks.Passed {
		trackBResult = "PASS"
	}

	data := map[string]any{
		"workspace": map[string]any{
			"root": root,
			"rule": rule,
		},
		"workspace_root":    root,
		"scope":             info.Scope,
		"verdict":           verdict,
		"target":            gateResult.Target,
		"rules":             gateResult.Rules,
		"advisories":        gateResult.Advisories,
		"decisions":         gateResult.Decisions,
		"gate_passed":       pass,
		"gate_status":       verdict,
		"track_filter":      trackFilter,
		"track_a_result":    trackAResult,
		"track_b_result":    trackBResult,
		"track_a":           gateResult.TrackAChecks,
		"structural_checks": gateResult.StructuralChecks,
		"scope_note":        "Evidence grades and verification labels are as reported by the agent; Vivechak did not fetch or verify sources.",
	}

	warnings := gateResult.Warnings
	warnings = append(warnings, gateResult.Advisories...)
	warnings = append(warnings, gateResult.TrackBChecks.Issues...)
	warnings = append(warnings, gateResult.StructuralChecks.Issues...)
	if gateResult.Verdict == "not_applicable" || info.Scope == core.ScopeComparison {
		warnings = append(warnings, "Comparison scope: zero gate requirements. Proceed to implementation.")
	}

	if gateResult.Verdict != "not_applicable" {
		projectName := core.DetectProjectName(ws, root)
		gateContent := core.RenderGateReport(ws, root, projectName, gateResult)
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
	}

	if !pass {
		var failingRules []string
		for _, r := range gateResult.Rules {
			if r.Status == "FAIL" {
				failingRules = append(failingRules, fmt.Sprintf("%s: %s", r.ID, r.Reason))
			}
		}
		issueCount := len(failingRules)
		msg := fmt.Sprintf("❌ PHASE 0 GATE FAILED (%d failing rule%s). Implementation is strictly prohibited until resolved.", issueCount, plural(issueCount))
		switch trackFilter {
		case "track_a":
			msg = fmt.Sprintf("❌ Phase 0 Gate: Track A requirements not met (%d failing rule%s). Implementation is strictly prohibited until resolved.", issueCount, plural(issueCount))
		case "track_b":
			msg = fmt.Sprintf("❌ Phase 0 Gate: Track B requirements not met (%d failing rule%s). Implementation is strictly prohibited until resolved.", issueCount, plural(issueCount))
		}
		nextStep := "⛔ DO NOT PROCEED TO CODE. Phase 0 Gate failed: resolve all failing rules above before beginning implementation."
		if gateResult.NextStep != "" {
			nextStep = fmt.Sprintf("⛔ DO NOT PROCEED TO CODE. %s", gateResult.NextStep)
		}
		env := Envelope{
			Ok:       true,
			Success:  true,
			Message:  msg,
			Data:     data,
			Warnings: warnings,
			Next: &NextAction{
				Why: fmt.Sprintf("DO NOT START IMPLEMENTATION. Resolve %d failing rule(s): %s", issueCount, strings.Join(failingRules, "; ")),
			},
			NextStep: nextStep,
			Meta:     NewMeta(tool),
		}
		return env.ToResult()
	}

	if gateResult.Verdict == "not_applicable" {
		env := Envelope{
			Ok:       true,
			Success:  true,
			Message:  "Phase 0 Gate: not applicable (Comparison scope only; no architecture gate required).",
			Data:     data,
			Warnings: warnings,
			Next: &NextAction{
				Why: "Comparison scope complete. WEP matrix generated. Proceed to implementation.",
			},
			NextStep: gateResult.NextStep,
			Meta:     NewMeta(tool),
		}
		return env.ToResult()
	}

	// PASS
	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Phase 0 Gate: PASS. %s", gateResult.NextStep),
		Data:     data,
		Warnings: warnings,
		Next: &NextAction{
			Why: gateResult.NextStep,
		},
		NextStep: gateResult.NextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// handleRunGate provides backward compatibility for unit tests calling handleRunGate.
func handleRunGate(ctx context.Context, req *sdkmcp.CallToolRequest, in RunGateInput) (*sdkmcp.CallToolResult, Envelope, error) {
	return handleGate(ctx, req, in)
}
