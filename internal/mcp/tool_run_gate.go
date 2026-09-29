package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RunGateInput holds the arguments for vivechak_run_gate.
type RunGateInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path"`
	Verbose     bool   `json:"verbose,omitempty"       jsonschema:"if true, return full gate details instead of summary"`
}

func registerRunGate(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_run_gate",
			Title: "Run Phase 0 Gate",
			Description: "Execute the Phase 0 exit gate check (Track A + Track B). " +
				"Performs MECHANICAL VALIDATION ONLY: ADR lock status, reversal triggers present, " +
				"evidence density, FAD completeness. Semantic quality assessment (is the premortem " +
				"substantive?) is the host agent's responsibility. " +
				"Supports progressive disclosure via verbose parameter.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
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

	// Structural completeness checks (Track A)
	var trackAIssues []string
	var trackAPassed int
	trackATotal := 5

	switch info.Scope {
	case core.ScopeDecision:
		trackATotal = 3
		// Check 1: At least 1 session completed
		if info.SessionCount > 0 {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "No completed sessions found")
		}

		// Check 2: Templates present
		if info.TemplateCount >= 5 {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, fmt.Sprintf("Only %d/5 templates found", info.TemplateCount))
		}

		// Check 3: Decision record exists
		hasDecision := info.HasDecisions
		if !hasDecision {
			if entries, err := ws.ListDir(core.ResearchDir); err == nil {
				for _, e := range entries {
					if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && !strings.HasSuffix(e.Name(), "-plan.md") && !strings.HasPrefix(e.Name(), ".") {
						hasDecision = true
						break
					}
				}
			}
		}
		if hasDecision {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "No decision record found")
		}

	case core.ScopeComparison:
		trackATotal = 2
		// Check 1: At least 1 session completed
		if info.SessionCount > 0 {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "No completed sessions found")
		}

		// Check 2: Templates present
		if info.TemplateCount >= 5 {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, fmt.Sprintf("Only %d/5 templates found", info.TemplateCount))
		}

	default: // ScopeProject
		trackATotal = 5
		// Check 1: Pipeline exists
		if info.HasPipeline {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "RESEARCH-PIPELINE.md not found")
		}

		// Check 2: At least 1 session completed
		if info.SessionCount > 0 {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "No completed sessions found")
		}

		// Check 3: Templates present
		if info.TemplateCount >= 5 {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, fmt.Sprintf("Only %d/5 templates found", info.TemplateCount))
		}

		// Check 4: Decisions exist
		if info.HasDecisions {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "DECISIONS.md not found")
		}

		// Check 5: FAD exists
		if _, err := ws.Stat(core.FADFile); err == nil {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "FAD.md not found — synthesis not complete")
		}
	}

	// Quality indicators (Track B)
	var trackBIssues []string
	var trackBPassed int
	trackBTotal := 3

	switch info.Scope {
	case core.ScopeComparison:
		trackBTotal = 2
		// Check B1: Session count
		if info.SessionCount >= 1 {
			trackBPassed++
		} else {
			trackBIssues = append(trackBIssues, "No comparison session found")
		}
		// Check B2: Evidence grades in session files
		hasGrades := false
		if entries, err := ws.ListDir(core.SessionsDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					if data, err := ws.ReadFile(filepath.Join(core.SessionsDir, e.Name())); err == nil {
						v := core.ValidateSession(data)
						if v.WarningCount() == 0 {
							hasGrades = true
							break
						}
					}
				}
			}
		}
		if hasGrades {
			trackBPassed++
		} else {
			trackBIssues = append(trackBIssues, "No evidence grades found in comparison session")
		}

	case core.ScopeDecision:
		trackBTotal = 2
		// Check B1: Session count
		if info.SessionCount >= 1 {
			trackBPassed++
		} else {
			trackBIssues = append(trackBIssues, "No decision session found")
		}
		// Check B2: Decision record has substantive content
		decisionSubstantive := false
		if info.HasDecisions {
			if decData, err := ws.ReadFile(core.DecisionsFile); err == nil && len(decData) > 100 {
				decisionSubstantive = true
			}
		}
		if !decisionSubstantive {
			if entries, err := ws.ListDir(core.ResearchDir); err == nil {
				for _, e := range entries {
					if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && !strings.HasSuffix(e.Name(), "-plan.md") && !strings.HasPrefix(e.Name(), ".") {
						if dData, err := ws.ReadFile(filepath.Join(core.ResearchDir, e.Name())); err == nil && len(dData) > 100 {
							decisionSubstantive = true
							break
						}
					}
				}
			}
		}
		if decisionSubstantive {
			trackBPassed++
		} else {
			trackBIssues = append(trackBIssues, "Decision record appears empty or trivial (< 100 chars)")
		}

	default: // ScopeProject
		trackBTotal = 3
		// Check B1: Session count meets minimum for scope
		minSessions := 3
		if info.SessionCount >= minSessions {
			trackBPassed++
		} else {
			trackBIssues = append(trackBIssues, fmt.Sprintf(
				"Only %d sessions — minimum %d recommended for %s scope",
				info.SessionCount, minSessions, info.Scope))
		}

		// Check B2: FAD has evidence grades if it exists
		if _, err := ws.Stat(core.FADFile); err == nil {
			fadData, err := ws.ReadFile(core.FADFile)
			if err == nil {
				fadValidation := core.ValidateArtifact(fadData)
				if fadValidation.WarningCount() == 0 {
					trackBPassed++
				} else {
					for _, issue := range fadValidation.Issues {
						if issue.Level == core.L3Warn {
							trackBIssues = append(trackBIssues, issue.String())
						}
					}
				}
			}
		} else {
			trackBIssues = append(trackBIssues, "Cannot check FAD quality — FAD not yet created")
		}

		// Check B3: Decisions file has content
		if info.HasDecisions {
			decData, err := ws.ReadFile(core.DecisionsFile)
			if err == nil && len(decData) > 100 {
				trackBPassed++
			} else {
				trackBIssues = append(trackBIssues, "DECISIONS.md appears empty or trivial")
			}
		} else {
			trackBIssues = append(trackBIssues, "Cannot check decisions — DECISIONS.md not found")
		}
	}

	// Overall gate result
	trackAPass := len(trackAIssues) == 0
	trackBPass := len(trackBIssues) == 0
	gatePassed := trackAPass && trackBPass

	var gateStatus string
	if gatePassed {
		gateStatus = "PASS"
	} else if trackAPass {
		gateStatus = "WARN"
	} else {
		gateStatus = "FAIL"
	}

	var warnings []string
	for _, issue := range trackAIssues {
		warnings = append(warnings, "GATE-STRUCTURAL: "+issue)
	}
	for _, issue := range trackBIssues {
		warnings = append(warnings, "GATE-QUALITY: "+issue)
	}

	var nextStep string
	switch gateStatus {
	case "PASS":
		nextStep = "Gate passed! The research phase is complete. You can now begin implementation. " +
			"Note: this gate checks structural completeness only — semantic quality " +
			"(premortem substance, alternative genuineness) is YOUR responsibility."
	case "WARN":
		nextStep = "Structural checks passed but quality indicators have warnings. " +
			"Review the warnings above. You may proceed if the warnings are acceptable, " +
			"or address them and re-run the gate."
	case "FAIL":
		nextStep = "Gate failed — structural issues must be addressed. " +
			"Fix the issues listed above and re-run vivechak_run_gate."
	}

	trackAData := map[string]any{
		"label":  "Structural Completeness",
		"passed": trackAPass,
		"score":  fmt.Sprintf("%d/%d", trackAPassed, trackATotal),
	}
	trackBData := map[string]any{
		"label":  "Quality Indicators (Mechanical)",
		"passed": trackBPass,
		"score":  fmt.Sprintf("%d/%d", trackBPassed, trackBTotal),
	}

	if in.Verbose {
		trackAData["issues"] = trackAIssues
		trackBData["issues"] = trackBIssues
	}

	data := map[string]any{
		"workspace_root":    root,
		"scope":             info.Scope,
		"gate_status":       gateStatus,
		"gate_passed":       gatePassed,
		"structural_checks": trackAData,
		"quality_checks":    trackBData,
		"track_a":           trackAData, // backward compatibility
		"track_b":           trackBData, // backward compatibility
		"scope_note": "This gate performs automated MECHANICAL checks only (Structural Completeness and Quality Indicators). " +
			"Track A/Track B semantic decisions and human sign-off per PHASE-0-GATE.template.md are the host architect's responsibility.",
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Phase 0 Gate: %s (Structural Completeness: %d/%d, Quality Indicators: %d/%d)", gateStatus, trackAPassed, trackATotal, trackBPassed, trackBTotal),
		Data:     data,
		Warnings: warnings,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}
