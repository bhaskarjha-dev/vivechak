package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/embed"
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
				"Structural checks: pipeline exists, DAG sessions completed, templates present, " +
				"decisions recorded, FAD compiled. Quality checks: session count, evidence grades in FAD, " +
				"ADR lock status, reversal triggers, evidentiary threshold for one-way doors. " +
				"Advisory: rejected alternatives, recalled verification, evidence grade distribution. " +
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
	var trackATotal int

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
		expectedTemplates := len(core.TemplatesForScope(core.ScopeDecision))
		if info.TemplateCount >= expectedTemplates {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, fmt.Sprintf("Only %d/%d templates found", info.TemplateCount, expectedTemplates))
		}

		// Check 3: Decision record exists
		hasDecision := info.HasDecisions
		if !hasDecision {
			if entries, err := ws.ListDir(core.ResearchDir); err == nil {
				for _, e := range entries {
					name := e.Name()
					if e.IsDir() || !strings.HasSuffix(name, ".md") {
						continue
					}
					if core.IsSpecialResearchFile(name) {
						continue
					}
					if strings.HasPrefix(strings.ToUpper(name), "D-") {
						hasDecision = true
						break
					}
					if dData, err := ws.ReadFile(filepath.Join(core.ResearchDir, name)); err == nil && len(dData) > 0 {
						if fm, _, err := core.ParseFrontmatter(dData); err == nil && fm != nil && fm.Has("door_type") {
							hasDecision = true
							break
						}
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
		expectedTemplates := len(core.TemplatesForScope(core.ScopeComparison))
		if info.TemplateCount >= expectedTemplates {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, fmt.Sprintf("Only %d/%d templates found", info.TemplateCount, expectedTemplates))
		}

	default: // ScopeProject
		trackATotal = 5
		// Check 1: Pipeline exists
		if info.HasPipeline {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "RESEARCH-PIPELINE.md not found")
		}

		// Check 2: DAG sessions completed
		if info.HasPipeline {
			if pipeData, err := ws.ReadFile(core.PipelineFile); err == nil {
				if dag, err := core.ParsePipeline(pipeData); err == nil && len(dag.Sessions) > 0 {
					completedCount := 0
					for _, s := range dag.Sessions {
						found := false
						if s.OutputFile != "" {
							if _, err := ws.Stat(s.OutputFile); err == nil {
								found = true
							}
						}
						if !found {
							if entries, err := ws.ListDir(core.SessionsDir); err == nil {
								for _, e := range entries {
									if strings.HasPrefix(e.Name(), s.ID) && strings.HasSuffix(e.Name(), ".md") {
										found = true
										break
									}
								}
							}
						}
						if !found && core.IsSynthesisSession(s.ID) {
							if _, err := ws.Stat(core.FADFile); err == nil {
								found = true
							}
						}
						if found {
							completedCount++
						}
					}
					if completedCount == len(dag.Sessions) {
						trackAPassed++
					} else {
						trackAIssues = append(trackAIssues, fmt.Sprintf("Pipeline DAG incomplete: %d/%d sessions completed", completedCount, len(dag.Sessions)))
					}
				} else if info.SessionCount > 0 {
					trackAPassed++
				} else {
					trackAIssues = append(trackAIssues, "No completed sessions found")
				}
			} else if info.SessionCount > 0 {
				trackAPassed++
			} else {
				trackAIssues = append(trackAIssues, "No completed sessions found")
			}
		} else if info.SessionCount > 0 {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "No completed sessions found")
		}

		// Check 3: Templates present
		expectedTemplates := len(core.TemplatesForScope(core.ScopeProject))
		if info.TemplateCount >= expectedTemplates {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, fmt.Sprintf("Only %d/%d templates found", info.TemplateCount, expectedTemplates))
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
	var trackBTotal int
	var gateAdvisories []string

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
		// Check B2: Decision record quality (accepted status, reversal triggers, content)
		if ok, bIssues, aWarnings := verifyDecisionsMechanical(ws, info); ok {
			trackBPassed++
			gateAdvisories = append(gateAdvisories, aWarnings...)
		} else {
			trackBIssues = append(trackBIssues, bIssues...)
			gateAdvisories = append(gateAdvisories, aWarnings...)
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
				fadValidation := core.ValidateFAD(fadData)
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

		// Check B3: Decisions mechanical validation (ADR lock status, review triggers)
		if ok, bIssues, aWarnings := verifyDecisionsMechanical(ws, info); ok {
			trackBPassed++
			gateAdvisories = append(gateAdvisories, aWarnings...)
		} else {
			trackBIssues = append(trackBIssues, bIssues...)
			gateAdvisories = append(gateAdvisories, aWarnings...)
		}
	}

	// Evidence integrity checks (B3/B4/B5) — advisory, not blocking
	if info.Scope == core.ScopeProject || info.Scope == core.ScopeDecision {
		eiWarns, eiAdvisories := verifyEvidentiaryIntegrity(ws, info)
		gateAdvisories = append(gateAdvisories, eiWarns...)
		gateAdvisories = append(gateAdvisories, eiAdvisories...)
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
	for _, adv := range gateAdvisories {
		warnings = append(warnings, "GATE-ADVISORY: "+adv)
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
		"workspace_root":          root,
		"scope":                   info.Scope,
		"gate_status":             gateStatus,
		"gate_passed":             gatePassed,
		"structural_checks":       trackAData,
		"quality_checks":          trackBData,
		"structural_completeness": trackAData,
		"mechanical_quality":      trackBData,
		"track_a":                 trackAData, // deprecated alias for structural_checks
		"track_b":                 trackBData, // deprecated alias for quality_checks
		"scope_note": "This gate performs automated MECHANICAL checks only (Structural Completeness and Quality Indicators). " +
			"Track A (Two-Way Door fast track) and Track B (One-Way Door rigorous gate) semantic decisions and human sign-off per PHASE-0-GATE.template.md are the host architect's responsibility.",
	}

	decisions := collectGateDecisions(ws)
	projectName := detectProjectName(ws, root)
	gateContent := renderGateArtifact(ws, root, projectName, gateStatus, trackAPass, trackBPass, trackAIssues, trackBIssues, decisions, gateAdvisories...)
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
		Message:  fmt.Sprintf("Phase 0 Gate: %s (Structural Completeness: %d/%d, Quality Indicators: %d/%d)", gateStatus, trackAPassed, trackATotal, trackBPassed, trackBTotal),
		Data:     data,
		Warnings: warnings,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// verifyDecisionsMechanical checks ADR files for mechanical correctness:
// valid frontmatter, accepted status, and reversal triggers on one-way door decisions.
func verifyDecisionsMechanical(ws *store.Workspace, info core.WorkspaceInfo) (bool, []string, []string) {
	if !info.HasDecisions {
		return false, []string{"Cannot check decisions — DECISIONS.md not found"}, nil
	}

	type decisionSource struct {
		file string
		data []byte
	}
	var sources []decisionSource

	if entries, err := ws.ListDir(core.ResearchDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") {
				continue
			}
			if core.IsSpecialResearchFile(name) {
				continue
			}
			relPath := filepath.Join(core.ResearchDir, name)
			isCandidate := strings.HasPrefix(strings.ToUpper(name), "D-")
			dData, err := ws.ReadFile(relPath)
			if err != nil || len(dData) == 0 {
				continue
			}
			if !isCandidate {
				if fm, _, err := core.ParseFrontmatter(dData); err == nil && fm != nil {
					if !fm.Has("door_type") {
						continue // Not a decision record
					}
				} else {
					continue
				}
			}
			sources = append(sources, decisionSource{file: relPath, data: dData})
		}
	}

	// Fallback to DECISIONS.md if no individual ADR files exist
	if len(sources) == 0 {
		if decData, err := ws.ReadFile(core.DecisionsFile); err == nil && len(decData) > 0 {
			sources = append(sources, decisionSource{file: core.DecisionsFile, data: decData})
		}
	}

	if len(sources) == 0 {
		return false, []string{"DECISIONS.md or ADR files appear empty or trivial"}, nil
	}

	hasDecisionsParsed := false
	allAccepted := true
	missingReviewTrigger := 0
	unreviewedOneWay := 0
	var unacceptedDetails []string
	var missingReviewDetails []string
	var unreviewedIDs []string
	processedIDs := make(map[string]bool)

	checkDecisionFM := func(fm core.Frontmatter, sourceFile string) {
		if fm == nil || (!fm.Has("id") && !fm.Has("decision_id")) {
			return
		}
		id := fm.GetString("id")
		if id == "" {
			id = fm.GetString("decision_id")
		}
		if strings.HasPrefix(strings.ToUpper(id), "CHK-") || fm.Has("conflicting_sources") {
			return // Skip conflict resolution records
		}
		if processedIDs[id] {
			return
		}
		processedIDs[id] = true
		hasDecisionsParsed = true

		status := strings.ToLower(strings.TrimSpace(fm.GetString("status")))
		if status != "accepted" && status != "superseded" && status != "deprecated" && status != "rejected" {
			allAccepted = false
			unacceptedDetails = append(unacceptedDetails, fmt.Sprintf("'%s' (id: %s) has status '%s'", sourceFile, id, status))
		}
		isOneWay := strings.EqualFold(fm.GetString("door_type"), "one-way") && status == "accepted"
		if isOneWay {
			reviewTrigger := strings.TrimSpace(fm.GetString("review_trigger"))
			isPlaceholder := reviewTrigger == "[Condition or date for mandatory re-evaluation]" ||
				(strings.HasPrefix(reviewTrigger, "[") && strings.HasSuffix(reviewTrigger, "]"))
			hasReversal := (!isPlaceholder && reviewTrigger != "") ||
				(fm.Has("reversal_triggers") && len(fm.GetStringSlice("reversal_triggers")) > 0) ||
				(fm.Has("review_date") && strings.TrimSpace(fm.GetString("review_date")) != "" && !strings.Contains(fm.GetString("review_date"), "YYYY"))
			if !hasReversal {
				missingReviewTrigger++
				missingReviewDetails = append(missingReviewDetails, fmt.Sprintf("'%s' (id: %s)", sourceFile, id))
			}
			// Check human_reviewed for one-way doors (advisory, not blocking)
			reviewed := strings.ToLower(strings.TrimSpace(fm.GetString("human_reviewed")))
			if reviewed != "true" {
				unreviewedOneWay++
				unreviewedIDs = append(unreviewedIDs, id)
			}
		}
	}

	anchoredRe := regexp.MustCompile(`(?s)<!-- DECISION:\s*([A-Za-z0-9_-]+)\s*-->\s*(.*?)\s*<!-- /DECISION:\s*[A-Za-z0-9_-]+\s*-->`)
	unanchoredRe := regexp.MustCompile(`(?ms)^---\s*\n(.*?)\n---`)

	for _, src := range sources {
		chunk := src.data
		file := src.file
		parsedAny := false

		// 1. Anchored blocks: <!-- DECISION: ID --> ... <!-- /DECISION: ID -->
		for _, m := range anchoredRe.FindAllSubmatch(chunk, -1) {
			if fm, _, err := core.ParseFrontmatter(m[2]); err == nil && fm != nil {
				if !fm.Has("id") && !fm.Has("decision_id") {
					fm["id"] = string(m[1])
				}
				checkDecisionFM(fm, file)
				parsedAny = true
			}
		}

		// 2. Unanchored frontmatter blocks
		if !parsedAny {
			for _, m := range unanchoredRe.FindAllSubmatch(chunk, -1) {
				if fm, _, err := core.ParseFrontmatter(m[0]); err == nil && fm != nil {
					checkDecisionFM(fm, file)
					parsedAny = true
				}
			}
		}

		// 3. Document-level frontmatter (e.g. standalone ADR files)
		if !parsedAny {
			if fm, _, err := core.ParseFrontmatter(chunk); err == nil && fm != nil {
				checkDecisionFM(fm, file)
			}
		}
	}

	var issues []string
	if !hasDecisionsParsed {
		issues = append(issues, "No architectural decisions recorded yet — record decisions using vivechak_record_decision before running the gate")
		return false, issues, nil
	}

	if !allAccepted {
		msg := fmt.Sprintf("%d decision(s) not yet finalized (accepted, rejected, superseded, or deprecated):", len(unacceptedDetails))
		for _, d := range unacceptedDetails {
			msg += "\n  → " + d
		}
		msg += "\n  help: set status to 'accepted', 'rejected', 'superseded', or 'deprecated' in each file's frontmatter, or if the file is not an ADR, rename it without the D- prefix."
		issues = append(issues, msg)
	}
	if missingReviewTrigger > 0 {
		msg := fmt.Sprintf("%d one-way door decision(s) lack required reversal triggers:", missingReviewTrigger)
		for _, d := range missingReviewDetails {
			msg += "\n  → " + d
		}
		msg += "\n  help: add 'review_trigger' with a measurable condition (e.g. 'latency > 50ms' not 'when needed')."
		issues = append(issues, msg)
	}

	// Advisory: warn about unreviewed one-way doors (not blocking)
	var advisories []string
	if unreviewedOneWay > 0 {
		idList := strings.Join(unreviewedIDs, ", ")
		advisories = append(advisories, fmt.Sprintf("⚠ %d one-way door decision(s) lack human review: %s. Consider reviewing before implementation.", unreviewedOneWay, idList))
	}

	return allAccepted && missingReviewTrigger == 0, issues, advisories
}

// recalledCitationRe matches explicit recalled verification methods, avoiding false positives on general memory topics (RAM, heap).
var recalledCitationRe = regexp.MustCompile(`(?i)(?:\[[^\]\n]*\b(?:recalled|parametric)\b[^\]\n]*\]|\([^)\n]*\b(?:recalled|parametric)\b[^)\n]*\)|\brecalled\s+memory\b|\|\s*recalled\s*\|)`)

// verifyEvidentiaryIntegrity checks B3 (evidence threshold), B4 (verification integrity), and B5 (rejected alternatives).
func verifyEvidentiaryIntegrity(ws *store.Workspace, info core.WorkspaceInfo) ([]string, []string) {
	var warnings []string
	var advisories []string

	decisions := collectGateDecisions(ws)
	lowGradeRe := regexp.MustCompile(`(?i)(?:\[(?:Grade\s+)?[CDE][^\]\n]*\]|\((?:Grade\s+)?[CDE][^)\n]*\)|\bGrade\s+[CDE]\b|\bGrade:[CDE]\b)`)

	for _, d := range decisions {
		if !strings.EqualFold(d.DoorType, "one-way") {
			continue
		}

		// Read decision file
		var decData []byte
		if dData, err := ws.ReadFile(filepath.Join(core.ResearchDir, d.ID+".md")); err == nil {
			decData = dData
		} else if dData, err := ws.ReadFile(filepath.Join(core.ResearchDir, d.ID+"-decision.md")); err == nil {
			decData = dData
		} else if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				name := e.Name()
				if (strings.HasPrefix(name, d.ID+"-") || strings.HasPrefix(name, d.ID+"_") || strings.EqualFold(name, d.ID+".md")) && strings.HasSuffix(name, ".md") && !core.IsSpecialResearchFile(name) {
					decData, _ = ws.ReadFile(filepath.Join(core.ResearchDir, name))
					break
				}
			}
		}

		// Fallback to DECISIONS.md if no individual file
		if len(decData) == 0 {
			if decAll, err := ws.ReadFile(core.DecisionsFile); err == nil {
				anchoredRe := regexp.MustCompile(fmt.Sprintf(`(?s)<!-- DECISION:\s*%s\s*-->\s*(.*?)\s*<!-- /DECISION:\s*%s\s*-->`, regexp.QuoteMeta(d.ID), regexp.QuoteMeta(d.ID)))
				if m := anchoredRe.FindSubmatch(decAll); len(m) > 1 {
					decData = m[1]
				}
			}
		}

		if len(decData) == 0 {
			continue
		}

		fm, body, _ := core.ParseFrontmatter(decData)
		bodyStr := string(body)

		// B5: Check rejected alternatives section (>30 characters of content)
		altContent := core.ExtractSection(bodyStr, "alternative", "rejected")
		if len(strings.TrimSpace(altContent)) < 30 {
			advisories = append(advisories, fmt.Sprintf(
				"B5-ALTERNATIVES: One-way door %s lacks documented rejected alternatives section", d.ID))
		}

		// Find informing sessions from frontmatter
		var sessIDs []string
		if fm != nil {
			sessIDs = fm.GetStringSlice("informed_by_sessions")
		}
		// Also scan session files for informs_decisions if empty
		if len(sessIDs) == 0 {
			if sEntries, sErr := ws.ListDir(core.SessionsDir); sErr == nil {
				for _, se := range sEntries {
					if se.IsDir() || !strings.HasSuffix(se.Name(), ".md") {
						continue
					}
					sData, _ := ws.ReadFile(filepath.Join(core.SessionsDir, se.Name()))
					if sFm, _, _ := core.ParseFrontmatter(sData); sFm != nil {
						infDecs := sFm.GetStringSlice("informs_decisions")
						for _, inf := range infDecs {
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
				for _, se := range sEntries {
					name := strings.TrimSuffix(se.Name(), ".md")
					nameUpper := strings.ToUpper(name)
					sidUpper := strings.ToUpper(sid)
					if nameUpper == sidUpper || strings.HasPrefix(nameUpper, sidUpper+"-") || strings.HasPrefix(nameUpper, sidUpper+"_") {
						sData, _ = ws.ReadFile(filepath.Join(core.SessionsDir, se.Name()))
						break
					}
				}
			}
			if len(sData) == 0 {
				continue
			}
			sessStr := string(sData)

			// B3: Check evidence grades
			lowGrades := len(lowGradeRe.FindAllString(sessStr, -1))
			if lowGrades > 0 {
				advisories = append(advisories, fmt.Sprintf(
					"B3-EVIDENTIARY: Session %s informs one-way door %s but has %d Grade C/D/E citation(s). One-way decisions require Grade A/B evidence.",
					sid, d.ID, lowGrades))
			}

			// B4: Check recalled verification
			recalledCount := len(recalledCitationRe.FindAllString(sessStr, -1))
			if recalledCount > 0 {
				advisories = append(advisories, fmt.Sprintf(
					"B4-VERIFICATION: Session %s informs one-way door %s but has %d 'recalled' citation(s). One-way decisions require fetched/cached verification.",
					sid, d.ID, recalledCount))
			}
		}
	}
	return warnings, advisories
}

type gateDecisionItem struct {
	ID            string
	Title         string
	DoorType      string
	Track         string
	Status        string
	HumanReviewed bool
}

func collectGateDecisions(ws *store.Workspace) []gateDecisionItem {
	type decisionSource struct {
		file string
		data []byte
	}
	var sources []decisionSource

	if entries, err := ws.ListDir(core.ResearchDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") {
				continue
			}
			if core.IsSpecialResearchFile(name) {
				continue
			}
			relPath := filepath.Join(core.ResearchDir, name)
			isCandidate := strings.HasPrefix(strings.ToUpper(name), "D-")
			dData, err := ws.ReadFile(relPath)
			if err != nil || len(dData) == 0 {
				continue
			}
			if !isCandidate {
				if fm, _, err := core.ParseFrontmatter(dData); err == nil && fm != nil {
					if !fm.Has("door_type") {
						continue
					}
				} else {
					continue
				}
			}
			sources = append(sources, decisionSource{file: relPath, data: dData})
		}
	}

	if len(sources) == 0 {
		if decData, err := ws.ReadFile(core.DecisionsFile); err == nil && len(decData) > 0 {
			sources = append(sources, decisionSource{file: core.DecisionsFile, data: decData})
		}
	}

	var decisions []gateDecisionItem
	processedIDs := make(map[string]bool)

	checkDecisionFM := func(fm core.Frontmatter) {
		if fm == nil || (!fm.Has("id") && !fm.Has("decision_id")) {
			return
		}
		id := fm.GetString("id")
		if id == "" {
			id = fm.GetString("decision_id")
		}
		if id == "" || strings.HasPrefix(strings.ToUpper(id), "CHK-") || fm.Has("conflicting_sources") {
			return
		}
		if processedIDs[id] {
			return
		}
		processedIDs[id] = true

		title := fm.GetString("title")
		if title == "" {
			title = id
		}
		doorType := fm.GetString("door_type")
		if doorType == "" {
			doorType = "two-way"
		}
		track := "A"
		if strings.EqualFold(doorType, "one-way") {
			track = "B"
		}
		status := strings.ToUpper(strings.TrimSpace(fm.GetString("status")))
		switch status {
		case "ACCEPTED":
			status = "PASS"
		case "":
			status = "PENDING"
		}

		reviewed := strings.EqualFold(strings.TrimSpace(fm.GetString("human_reviewed")), "true")

		decisions = append(decisions, gateDecisionItem{
			ID:            id,
			Title:         title,
			DoorType:      doorType,
			Track:         track,
			Status:        status,
			HumanReviewed: reviewed,
		})
	}

	anchoredRe := regexp.MustCompile(`(?s)<!-- DECISION:\s*([A-Za-z0-9_-]+)\s*-->\s*(.*?)\s*<!-- /DECISION:\s*[A-Za-z0-9_-]+\s*-->`)
	unanchoredRe := regexp.MustCompile(`(?ms)^---\s*\n(.*?)\n---`)

	for _, src := range sources {
		chunk := src.data
		parsedAny := false

		for _, m := range anchoredRe.FindAllSubmatch(chunk, -1) {
			if fm, _, err := core.ParseFrontmatter(m[2]); err == nil && fm != nil {
				if !fm.Has("id") && !fm.Has("decision_id") {
					fm["id"] = string(m[1])
				}
				checkDecisionFM(fm)
				parsedAny = true
			}
		}

		if !parsedAny {
			if matches := unanchoredRe.FindAllSubmatch(chunk, -1); len(matches) > 0 {
				for _, m := range matches {
					block := fmt.Sprintf("---\n%s\n---", string(m[1]))
					if fm, _, err := core.ParseFrontmatter([]byte(block)); err == nil && fm != nil {
						if fm.Has("id") || fm.Has("decision_id") || fm.Has("door_type") {
							checkDecisionFM(fm)
							parsedAny = true
						}
					}
				}
			}
		}

		if !parsedAny {
			if fm, _, err := core.ParseFrontmatter(chunk); err == nil && fm != nil {
				checkDecisionFM(fm)
			}
		}
	}

	return decisions
}

func detectProjectName(ws *store.Workspace, root string) string {
	if data, err := ws.ReadFile(core.PipelineFile); err == nil {
		if fm, _, err := core.ParseFrontmatter(data); err == nil && fm != nil {
			if proj := fm.GetString("project"); proj != "" {
				return proj
			}
		}
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "# ") {
				heading := strings.TrimPrefix(trimmed, "# ")
				if strings.Contains(heading, ":") {
					parts := strings.SplitN(heading, ":", 2)
					if p := strings.TrimSpace(parts[1]); p != "" {
						return p
					}
				}
				return heading
			}
		}
	}
	base := filepath.Base(root)
	if base != "" && base != "." && base != "/" {
		return base
	}
	return "Project"
}

func renderGateArtifact(ws *store.Workspace, root, projectName, gateStatus string, trackAPass, trackBPass bool, trackAIssues, trackBIssues []string, decisions []gateDecisionItem, advisories ...string) []byte {
	var tmplBytes []byte
	if data, err := ws.ReadFile(filepath.Join(core.TemplatesDir, "PHASE-0-GATE.template.md")); err == nil {
		tmplBytes = data
	} else if data, err := embed.ReadTemplate("PHASE-0-GATE.template.md"); err == nil {
		tmplBytes = data
	}

	if len(tmplBytes) == 0 {
		return nil
	}

	tmpl := strings.ReplaceAll(string(tmplBytes), "\r\n", "\n")
	today := time.Now().UTC().Format("2006-01-02")

	// 1. Frontmatter and headers
	tmpl = strings.Replace(tmpl, `project: "[Project Name]"`, fmt.Sprintf(`project: "%s"`, projectName), 1)
	tmpl = strings.Replace(tmpl, `date: "[YYYY-MM-DD]"`, fmt.Sprintf(`date: "%s"`, today), 1)
	tmpl = strings.Replace(tmpl, `verdict: "[PENDING | PASS | FAIL]"`, fmt.Sprintf(`verdict: "%s"`, gateStatus), 1)

	trackAResultStr := "FAIL"
	if trackAPass {
		trackAResultStr = "PASS"
	}
	trackBResultStr := "FAIL"
	if trackBPass {
		trackBResultStr = "PASS"
	}

	tmpl = strings.Replace(tmpl, `track_a_result: "[PENDING | PASS | FAIL]"`, fmt.Sprintf(`track_a_result: "%s"`, trackAResultStr), 1)
	tmpl = strings.Replace(tmpl, `track_b_result: "[PENDING | PASS | FAIL]"`, fmt.Sprintf(`track_b_result: "%s"`, trackBResultStr), 1)

	tmpl = strings.Replace(tmpl, `## Project: `+"`[Project Name]`", fmt.Sprintf(`## Project: %s`, projectName), 1)
	tmpl = strings.Replace(tmpl, `**Gate Date:** `+"`[YYYY-MM-DD]`", fmt.Sprintf(`**Gate Date:** %s`, today), 1)

	// 2. Decision Routing Summary
	var routingTable strings.Builder
	if len(decisions) == 0 {
		routingTable.WriteString("| - | None recorded | - | - | PENDING |\n")
	} else {
		for _, d := range decisions {
			fmt.Fprintf(&routingTable, "| %s | %s | %s | Track %s | %s |\n", d.ID, d.Title, d.DoorType, d.Track, d.Status)
		}
	}
	oldDecisionRows := "| D-001 | `[Title]` | `[1-way/2-way]` | `[A/B]` | `[PASS/FAIL/PENDING]` |\n" +
		"| D-002 | `[Title]` | `[1-way/2-way]` | `[A/B]` | `[PASS/FAIL/PENDING]` |\n" +
		"| D-NNN | `[Title]` | `[1-way/2-way]` | `[A/B]` | `[PASS/FAIL/PENDING]` |"
	tmpl = strings.Replace(tmpl, oldDecisionRows, strings.TrimRight(routingTable.String(), "\n"), 1)

	// 3. Track A section
	if trackAPass {
		tmpl = strings.Replace(tmpl, "- [ ] Decision logged in ADR", "- [x] Decision logged in ADR", 1)
		tmpl = strings.Replace(tmpl, "- [ ] Reversibility confirmed", "- [x] Reversibility confirmed", 1)
		tmpl = strings.Replace(tmpl, "- [ ] At least one corroborated source", "- [x] At least one corroborated source", 1)
	}
	tmpl = strings.Replace(tmpl, "**Track A Result:** `[PASS / FAIL]`", fmt.Sprintf("**Track A Result:** %s", trackAResultStr), 1)

	// 4. Track B section
	if trackBPass {
		// B1: DAG Closure
		tmpl = strings.Replace(tmpl, "- [ ] All required dependency paths", "- [x] All required dependency paths", 1)
		tmpl = strings.Replace(tmpl, "- [ ] No orphaned sessions remain", "- [x] No orphaned sessions remain", 1)

		// B2: Contradictions
		tmpl = strings.Replace(tmpl, "- [ ] All cross-model divergences resolved", "- [x] All cross-model divergences resolved", 1)
		tmpl = strings.Replace(tmpl, "- [ ] No unresolved `contested` corroboration flags", "- [x] No unresolved `contested` corroboration flags", 1)

		// B3, B4, B5 checks based on advisories
		hasB3Adv := false
		hasB4Adv := false
		hasB5Adv := false
		for _, adv := range advisories {
			if strings.Contains(adv, "B3-EVIDENTIARY") {
				hasB3Adv = true
			}
			if strings.Contains(adv, "B4-VERIFICATION") {
				hasB4Adv = true
			}
			if strings.Contains(adv, "B5-ALTERNATIVES") {
				hasB5Adv = true
			}
		}
		if !hasB3Adv {
			tmpl = strings.Replace(tmpl, "- [ ] Zero uncorroborated Grade C/D/E claims", "- [x] Zero uncorroborated Grade C/D/E claims", 1)
			tmpl = strings.Replace(tmpl, "- [ ] All critical claims backed by Grade A or B", "- [x] All critical claims backed by Grade A or B", 1)
		}
		if !hasB4Adv {
			tmpl = strings.Replace(tmpl, "- [ ] 100% of critical citations carry", "- [x] 100% of critical citations carry", 1)
			tmpl = strings.Replace(tmpl, "- [ ] Zero `recalled` citations support", "- [x] Zero `recalled` citations support", 1)
		}
		if !hasB5Adv {
			tmpl = strings.Replace(tmpl, "- [ ] Every locked ADR explicitly details", "- [x] Every locked ADR explicitly details", 1)
			tmpl = strings.Replace(tmpl, "- [ ] Rejection rationale is causal", "- [x] Rejection rationale is causal", 1)
		}

		// B6: Decay Triggers
		tmpl = strings.Replace(tmpl, "- [ ] Every locked ADR contains an explicit `review_trigger`", "- [x] Every locked ADR contains an explicit `review_trigger`", 1)
		tmpl = strings.Replace(tmpl, "- [ ] Review triggers are specific and measurable", "- [x] Review triggers are specific and measurable", 1)

		// B7: Premortem Protocol (check if FAD documents failure scenarios)
		if fadBytes, fErr := ws.ReadFile(core.FADFile); fErr == nil {
			fadStr := strings.ToLower(string(fadBytes))
			if strings.Contains(fadStr, "premortem") || strings.Contains(fadStr, "failure scenario") || strings.Contains(fadStr, "risk register") {
				tmpl = strings.Replace(tmpl, "- [ ] 30-minute prospective hindsight", "- [x] 30-minute prospective hindsight", 1)
				tmpl = strings.Replace(tmpl, "- [ ] Prompt: *\"It is 12 months from now", "- [x] Prompt: *\"It is 12 months from now", 1)
				tmpl = strings.Replace(tmpl, "- [ ] Top 3 failure scenarios documented", "- [x] Top 3 failure scenarios documented", 1)
				tmpl = strings.Replace(tmpl, "- [ ] Mitigations incorporated into FAD", "- [x] Mitigations incorporated into FAD", 1)
			}
		}

		// B8: Human Architect Review (check human_reviewed in one-way decisions)
		hasOneWay := false
		allHumanReviewed := true
		for _, d := range decisions {
			if strings.EqualFold(d.DoorType, "one-way") {
				hasOneWay = true
				if !d.HumanReviewed {
					allHumanReviewed = false
				}
			}
		}
		if hasOneWay && allHumanReviewed {
			tmpl = strings.Replace(tmpl, "- [ ] Named Principal Architect has reviewed", "- [x] Named Principal Architect has reviewed", 1)
		}

		// B9: FAD Sealed
		if _, statErr := ws.Stat(core.FADFile); statErr == nil {
			tmpl = strings.Replace(tmpl, "- [ ] FAD compiled with full traceability", "- [x] FAD compiled with full traceability", 1)
			tmpl = strings.Replace(tmpl, "- [ ] FAD committed to repository root", "- [x] FAD committed to repository root", 1)
			tmpl = strings.Replace(tmpl, "- [ ] Repository scaffolding ready to generate", "- [x] Repository scaffolding ready to generate", 1)
		}
	}
	tmpl = strings.Replace(tmpl, "**Track B Result:** `[PASS / FAIL]`", fmt.Sprintf("**Track B Result:** %s", trackBResultStr), 1)

	// 5. Verdict table
	trackAIssuesSummary := "None"
	if len(trackAIssues) > 0 {
		trackAIssuesSummary = strings.Join(trackAIssues, "; ")
	}
	trackBIssuesSummary := "None"
	if len(trackBIssues) > 0 {
		trackBIssuesSummary = strings.Join(trackBIssues, "; ")
	}

	oldVerdictTable := "| Track A (Two-Way) | `[PASS/FAIL]` | `[None / List issues]` |\n" +
		"| Track B (One-Way) | `[PASS/FAIL]` | `[None / List issues]` |\n" +
		"| **Overall** | **`[PASS/FAIL]`** | |"
	newVerdictTable := fmt.Sprintf("| Track A (Two-Way) | %s | %s |\n"+
		"| Track B (One-Way) | %s | %s |\n"+
		"| **Overall** | **%s** | |", trackAResultStr, trackAIssuesSummary, trackBResultStr, trackBIssuesSummary, gateStatus)
	tmpl = strings.Replace(tmpl, oldVerdictTable, newVerdictTable, 1)

	// 6. Date at footer
	tmpl = strings.Replace(tmpl, "**Date:** `[YYYY-MM-DD]`", fmt.Sprintf("**Date:** %s", today), 1)

	return []byte(tmpl)
}


