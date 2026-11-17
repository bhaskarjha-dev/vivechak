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

// GateInput defines the canonical input parameters for vivechak_gate.
type GateInput struct {
	ProjectRoot string `json:"project_root,omitempty"`
	Decision    string `json:"decision,omitempty"`
	Track       string `json:"track,omitempty"`
	Verbose     bool   `json:"verbose,omitempty"`
}

func registerGate(server *sdkmcp.Server) {
	registerTool(server, ToolSpec{
		Name:        "vivechak_gate",
		Description: "Evaluate Phase 0 exit criteria (G1-G12) for the workspace or one ADR and write PHASE-0-GATE.md (one-way needs vck signoff).",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint:    false,
			IdempotentHint:  false,
			DestructiveHint: BoolPtr(true),
		},
		Params: []ParamSpec{
			{Name: "decision", Type: "string", Description: "D-ID", Doc: "Specific decision ID to evaluate (e.g. D-001)."},
			{Name: "track", Type: "string", Enum: []string{"all", "track_a", "track_b"}, Doc: "Track filter: all (default), track_a, or track_b."},
			{Name: "verbose", Type: "boolean", Doc: "Include verbose rule diagnostics."},
			{Name: "project_root", Type: "string", Doc: "Absolute path to the target project directory."},
		},
	}, handleGate)
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
		return ErrorResult(tool, fmt.Errorf("E-NOT-INITIALIZED: workspace not initialized at %s", root), "Run vivechak_start first.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: opening workspace: %w", err), "Provide a valid workspace.")
	}
	defer ws.Close()

	trackFilter := strings.ToLower(strings.TrimSpace(in.Track))
	if trackFilter == "" {
		trackFilter = "all"
	}
	if trackFilter != "track_a" && trackFilter != "track_b" && trackFilter != "all" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: invalid track filter %q: must be 'track_a', 'track_b', or 'all'", in.Track),
			"Specify 'track_a', 'track_b', or omit for all.")
	}

	var rawWarnings []string
	targetDecision := strings.TrimSpace(in.Decision)
	if targetDecision != "" {
		if resolved, mapped, idErr := resolvePlaceholderRef(sessionOf(req), ws, root, targetDecision, true); idErr != nil {
			return ErrorResult(tool, idErr, placeholderRefFix)
		} else if mapped {
			rawWarnings = append(rawWarnings, fmt.Sprintf("W-ID-ALLOCATED: decision %s resolved to %s", targetDecision, resolved))
			targetDecision = resolved
		} else {
			targetDecision = core.NormalizeDecisionID(targetDecision)
		}
	}

	info := core.InspectWorkspace(root)
	gateResult := core.EvaluateGateWithTarget(ws, info, root, targetDecision)

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

	listedRules := make([]core.GateRuleResult, 0, len(gateResult.Rules))
	for _, r := range gateResult.Rules {
		if (trackFilter == "track_a" && r.ID != "G12") || (trackFilter == "track_b" && r.ID == "G12") {
			continue
		}
		listedRules = append(listedRules, r)
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
		"rules":             listedRules,
		"advisories":        gateResult.Advisories,
		"decisions":         gateResult.Decisions,
		"track_filter":      trackFilter,
		"track_a":           gateResult.TrackAChecks,
		"structural_checks": gateResult.StructuralChecks,
		"scope_note":        "Evidence grades and verification labels are as reported by the agent; Vivechak did not fetch or verify sources.",
	}

	rawWarnings = append(rawWarnings, gateResult.Warnings...)
	rawWarnings = append(rawWarnings, gateResult.Advisories...)
	rawWarnings = append(rawWarnings, gateResult.TrackBChecks.Issues...)
	rawWarnings = append(rawWarnings, gateResult.StructuralChecks.Issues...)

	warnings := ParseWarnings(rawWarnings)
	if gateResult.Verdict == "not_applicable" {
		warnings = append(warnings, NewWarning("W-CMP-GATE-NA", "Comparison scope: zero gate requirements. Proceed to implementation.", ""))
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
				warnings = append(warnings, NewWarning("W-GATE-WRITE", fmt.Sprintf("failed to write gate artifact: %v", writeErr), ""))
			} else {
				data["gate_artifact"] = core.GateFile
			}
		}
	}

	if !pass {
		var failingRules []string
		nextTool := "vivechak_status"
		nextArgs := map[string]any{"format": "summary"}
		for _, r := range listedRules {
			if r.Status == "FAIL" {
				failingRules = append(failingRules, fmt.Sprintf("%s: %s", r.ID, r.Reason))
				if nextTool == "vivechak_status" {
					switch r.ID {
					case "G1":
						nextTool = "vivechak_save"
						nextArgs = map[string]any{"kind": "plan"}
					case "G2":
						nextTool = "vivechak_next"
						nextArgs = map[string]any{}
					case "G3":
						nextTool = "vivechak_save"
						nextArgs = map[string]any{"kind": "fad"}
					case "G4", "G5", "G8", "G9", "G12":
						nextTool = "vivechak_record_decision"
						did := "D-NEW"
						if len(gateResult.Decisions) > 0 {
							did = gateResult.Decisions[0].ID
						}
						nextArgs = map[string]any{"id": did, "revise": true}
					case "G6", "G7", "G10":
						nextTool = "vivechak_review"
						did := "D-001"
						if len(gateResult.Decisions) > 0 {
							did = gateResult.Decisions[0].ID
						}
						nextArgs = map[string]any{"mode": "red_team", "id": did}
					}
				}
			}
		}
		if len(failingRules) == 0 && trackFilter == "track_a" && len(gateResult.TrackAChecks.Issues) > 0 {
			failingRules = append(failingRules, gateResult.TrackAChecks.Issues...)
			if info.Scope == core.ScopeProject {
				nextTool = "vivechak_save"
				nextArgs = map[string]any{"kind": "plan"}
			}
		}
		issueCount := len(failingRules)
		msg := fmt.Sprintf("Phase 0 Gate: FAIL (%d failing rule%s). Implementation is prohibited until resolved.", issueCount, plural(issueCount))
		switch trackFilter {
		case "track_a":
			msg = fmt.Sprintf("Phase 0 Gate: Track A FAIL (%d failing rule%s). Implementation is prohibited until resolved.", issueCount, plural(issueCount))
		case "track_b":
			msg = fmt.Sprintf("Phase 0 Gate: Track B FAIL (%d failing rule%s). Implementation is prohibited until resolved.", issueCount, plural(issueCount))
		}
		env := Envelope{
			Ok:       true,
			Message:  msg,
			Data:     data,
			Warnings: warnings,
			Next: &NextAction{
				Tool: nextTool,
				Args: nextArgs,
				Why:  fmt.Sprintf("DO NOT START IMPLEMENTATION. Resolve %d failing rule(s): %s", issueCount, strings.Join(failingRules, "; ")),
			},
		}
		return env.ToResult()
	}

	if gateResult.Verdict == "not_applicable" {
		env := Envelope{
			Ok:       true,
			Message:  "Phase 0 Gate: not applicable (Comparison scope only; no architecture gate required).",
			Data:     data,
			Warnings: warnings,
			Next: &NextAction{
				Tool: "vivechak_status",
				Args: map[string]any{"format": "summary"},
				Why:  "Comparison scope complete. WEP matrix generated. Proceed to implementation.",
			},
		}
		return env.ToResult()
	}

	// PASS: gateResult.NextStep already begins with "Phase 0 Gate: PASS." — do not repeat it (HIGH-06).
	passMsg := gateResult.NextStep
	if !strings.HasPrefix(passMsg, "Phase 0 Gate: PASS") {
		passMsg = "Phase 0 Gate: PASS. " + passMsg
	}
	env := Envelope{
		Ok:       true,
		Message:  passMsg,
		Data:     data,
		Warnings: warnings,
		Next: &NextAction{
			Tool: "vivechak_status",
			Args: map[string]any{"format": "summary"},
			Why:  passMsg,
		},
	}
	return env.ToResult()
}
