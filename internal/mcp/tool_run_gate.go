package mcputil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
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

	info := core.InspectWorkspace(root)

	// Track A: Structural completeness
	var trackAIssues []string
	var trackAPassed int
	trackATotal := 5

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
	fadPath := filepath.Join(root, "research", "FAD.md")
	hasFAD := false
	if _, err := os.Stat(fadPath); err == nil {
		hasFAD = true
		trackAPassed++
	} else {
		trackAIssues = append(trackAIssues, "FAD.md not found — synthesis not complete")
	}

	// Track B: Quality indicators (mechanical only)
	var trackBIssues []string
	var trackBPassed int
	trackBTotal := 3

	// Check B1: Session count meets minimum for scope
	minSessions := 3 // project scope default
	switch info.Scope {
	case core.ScopeComparison:
		minSessions = 1
	case core.ScopeDecision:
		minSessions = 1
	}
	if info.SessionCount >= minSessions {
		trackBPassed++
	} else {
		trackBIssues = append(trackBIssues, fmt.Sprintf(
			"Only %d sessions — minimum %d recommended for %s scope",
			info.SessionCount, minSessions, info.Scope))
	}

	// Check B2: FAD has evidence grades if it exists
	if hasFAD {
		fadData, err := os.ReadFile(fadPath)
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
		decData, err := os.ReadFile(filepath.Join(root, core.DecisionsFile))
		if err == nil && len(decData) > 100 {
			trackBPassed++
		} else {
			trackBIssues = append(trackBIssues, "DECISIONS.md appears empty or trivial")
		}
	} else {
		trackBIssues = append(trackBIssues, "Cannot check decisions — DECISIONS.md not found")
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
		warnings = append(warnings, "GATE-A: "+issue)
	}
	for _, issue := range trackBIssues {
		warnings = append(warnings, "GATE-B: "+issue)
	}

	var nextStep string
	switch gateStatus {
	case "PASS":
		nextStep = "Gate passed! The research phase is complete. You can now begin implementation. " +
			"Note: this gate checks structural completeness only — semantic quality " +
			"(premortem substance, alternative genuineness) is YOUR responsibility."
	case "WARN":
		nextStep = "Track A (structural) passed but Track B (quality) has warnings. " +
			"Review the warnings above. You may proceed if the warnings are acceptable, " +
			"or address them and re-run the gate."
	case "FAIL":
		nextStep = "Gate failed — structural issues must be addressed. " +
			"Fix the Track A issues listed above and re-run vivechak_run_gate."
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
		"workspace_root": root,
		"gate_status":    gateStatus,
		"gate_passed":    gatePassed,
		"track_a":        trackAData,
		"track_b":        trackBData,
		"scope_note": "This gate performs MECHANICAL checks only. " +
			"Semantic quality assessment is the host agent's responsibility.",
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Phase 0 Gate: %s (Track A: %d/%d, Track B: %d/%d)", gateStatus, trackAPassed, trackATotal, trackBPassed, trackBTotal),
		Data:     data,
		Warnings: warnings,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}
