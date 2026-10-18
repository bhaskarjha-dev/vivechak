package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
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

// RunGateInput defines the input parameters for vivechak_run_gate.
type RunGateInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"root directory of the project (defaults to working directory)"`
	Verbose     bool   `json:"verbose,omitempty"       jsonschema:"if true, return full gate details instead of summary"`
	Track       string `json:"track,omitempty"         jsonschema:"optional track filter ('track_a', 'track_b', or 'all')"`
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
				"Supports progressive disclosure via verbose parameter and track filtering via track parameter ('track_a', 'track_b', or 'all').",
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

	trackFilter := strings.ToLower(strings.TrimSpace(in.Track))
	if trackFilter == "" {
		trackFilter = "all"
	}
	if trackFilter != "track_a" && trackFilter != "track_b" && trackFilter != "all" {
		return ErrorResult(tool, fmt.Errorf("invalid track filter %q: must be 'track_a', 'track_b', or 'all'", in.Track),
			"Specify 'track_a', 'track_b', or omit for all.")
	}

	trackAData := map[string]any{
		"label":  gateResult.TrackAChecks.Label,
		"passed": gateResult.TrackAChecks.Passed,
		"score":  gateResult.TrackAChecks.Score,
	}
	trackBData := map[string]any{
		"label":  gateResult.TrackBChecks.Label,
		"passed": gateResult.TrackBChecks.Passed,
		"score":  gateResult.TrackBChecks.Score,
	}
	structuralData := map[string]any{
		"label":  gateResult.StructuralChecks.Label,
		"passed": gateResult.StructuralChecks.Passed,
		"score":  gateResult.StructuralChecks.Score,
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

	// Track B Check B11 (Red-Team Challenge Verification):
	// For any One-Way Door decision, inspect whether the session markdown or ADR
	// includes a "## Challenge Findings" or "## Red-Team Stress Test" section.
	for _, d := range decisions {
		if !strings.EqualFold(d.DoorType, "one-way") {
			continue
		}
		hasChallenge := false
		var adrData []byte
		if b, err := ws.ReadFile(filepath.Join(core.ResearchDir, d.ID+".md")); err == nil {
			adrData = b
		} else if b, err := ws.ReadFile(filepath.Join(core.ResearchDir, d.ID+"-decision.md")); err == nil {
			adrData = b
		} else if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				name := e.Name()
				if (strings.HasPrefix(name, d.ID+"-") || strings.HasPrefix(name, d.ID+"_") || strings.EqualFold(name, d.ID+".md")) && strings.HasSuffix(name, ".md") && !core.IsSpecialResearchFile(name) {
					adrData, _ = ws.ReadFile(filepath.Join(core.ResearchDir, name))
					break
				}
			}
		}
		if len(adrData) == 0 {
			if decAll, err := ws.ReadFile(core.DecisionsFile); err == nil {
				anchoredRe := regexp.MustCompile(fmt.Sprintf(`(?s)<!-- DECISION:\s*%s\s*-->\s*(.*?)\s*<!-- /DECISION:\s*%s\s*-->`, regexp.QuoteMeta(d.ID), regexp.QuoteMeta(d.ID)))
				if m := anchoredRe.FindSubmatch(decAll); len(m) > 1 {
					adrData = m[1]
				}
			}
		}

		checkChallengeText := func(text string) bool {
			lower := strings.ToLower(text)
			return strings.Contains(lower, "## challenge findings") ||
				strings.Contains(lower, "## red-team stress test") ||
				strings.Contains(lower, "## red team stress test") ||
				strings.Contains(lower, "challenge findings") ||
				strings.Contains(lower, "red-team stress test")
		}

		if len(adrData) > 0 && checkChallengeText(string(adrData)) {
			hasChallenge = true
		}

		if !hasChallenge {
			fm, _, _ := core.ParseFrontmatter(adrData)
			var sessIDs []string
			if fm != nil {
				sessIDs = fm.GetStringSlice("informed_by_sessions")
			}
			if len(sessIDs) == 0 {
				if sEntries, sErr := ws.ListDir(core.SessionsDir); sErr == nil {
					for _, se := range sEntries {
						if se.IsDir() || !strings.HasSuffix(se.Name(), ".md") {
							continue
						}
						sData, _ := ws.ReadFile(filepath.Join(core.SessionsDir, se.Name()))
						if sFm, _, _ := core.ParseFrontmatter(sData); sFm != nil {
							for _, inf := range sFm.GetStringSlice("informs_decisions") {
								if strings.EqualFold(inf, d.ID) {
									sessIDs = append(sessIDs, core.ExtractSessionID(se.Name(), sFm))
								}
							}
						}
					}
				}
			}
			for _, sid := range sessIDs {
				var sData []byte
				if sd, err := ws.ReadFile(filepath.Join(core.SessionsDir, sid+".md")); err == nil {
					sData = sd
				} else if sEntries, lErr := ws.ListDir(core.SessionsDir); lErr == nil {
					sidUpper := strings.ToUpper(sid)
					for _, se := range sEntries {
						nameUpper := strings.ToUpper(strings.TrimSuffix(se.Name(), ".md"))
						if nameUpper == sidUpper || strings.HasPrefix(nameUpper, sidUpper+"-") || strings.HasPrefix(nameUpper, sidUpper+"_") {
							sData, _ = ws.ReadFile(filepath.Join(core.SessionsDir, se.Name()))
							break
						}
					}
				}
				if len(sData) > 0 && checkChallengeText(string(sData)) {
					hasChallenge = true
					break
				}
			}
		}

		if !hasChallenge {
			warnings = append(warnings, fmt.Sprintf("W-OWD-NO-CHALLENGE: One-Way Door decision %s was not subjected to adversarial red-team challenge before commitment.", d.ID))
		}
	}

	allAdvisories := append(append([]string{}, gateResult.Advisories...), warnings...)
	gateContent := core.RenderGateArtifact(ws, tmplBytes, root, projectName, gateResult.Status, gateResult.TrackAChecks.Passed, gateResult.TrackBChecks.Passed, gateResult.TrackAChecks.Issues, gateResult.TrackBChecks.Issues, decisions, allAdvisories...)
	if len(gateContent) > 0 {
		_ = ws.MkdirAll(core.ResearchDir, 0o755)
		if writeErr := store.WriteFileAtomic(ws.Root(), core.GateFile, gateContent, 0o644); writeErr != nil {
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
			Success: true,
			Message: fmt.Sprintf("❌ PHASE 0 GATE FAILED: %s requirements not met (%d blocking issue%s). Implementation is strictly prohibited.",
				failureScope, issueCount, plural(issueCount)),
			Data:     data,
			Warnings: warnings,
			NextStep: fmt.Sprintf("⛔ DO NOT PROCEED TO CODE. Phase 0 exit criteria NOT MET. "+
				"Resolve the %d blocking issue(s) before code scaffolding: %s",
				issueCount, strings.Join(blockingIssues, "; ")),
			Meta: NewMeta(tool),
		}
		return env.ToResult()
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Phase 0 Gate: %s (Structural Completeness: %d/%d, Quality Indicators: %d/%d)", gateStatus, gateResult.StructuralChecks.Count, gateResult.StructuralChecks.Total, gateResult.QualityChecks.Count, gateResult.QualityChecks.Total),
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
