package core

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// TrackResult represents the status and score of a gate track (Track A or Track B).
type TrackResult struct {
	Label  string   `json:"label"`
	Passed bool     `json:"passed"`
	Score  string   `json:"score"`
	Count  int      `json:"count"`
	Total  int      `json:"total"`
	Issues []string `json:"issues,omitempty"`
}

// GateResult holds the complete evaluation result from Phase 0 Gate.
type GateResult struct {
	Status           string             `json:"gate_status"`
	Passed           bool               `json:"gate_passed"`
	StructuralChecks TrackResult        `json:"structural_checks"` // Pipeline, templates, DAG, FAD
	TrackAChecks     TrackResult        `json:"track_a_checks"`     // Two-Way Door fast track (A1-A4)
	TrackBChecks     TrackResult        `json:"track_b_checks"`     // One-Way Door rigor (B1-B11)
	QualityChecks    TrackResult        `json:"quality_checks"`     // Backward-compatible alias for Track B
	Advisories       []string           `json:"advisories,omitempty"`
	Warnings         []string           `json:"warnings,omitempty"`
	NextStep         string             `json:"next_step"`
	Decisions        []GateDecisionItem `json:"decisions,omitempty"`
}

// EvaluateGate performs mechanical verification of structural completeness and quality indicators.
func EvaluateGate(ws WorkspaceFS, info WorkspaceInfo, root string) *GateResult {
	// Structural completeness checks
	var structuralIssues []string
	var structuralPassed int
	var structuralTotal int

	switch info.Scope {
	case ScopeDecision:
		structuralTotal = 3
		// Check 1: At least 1 session completed
		if info.SessionCount > 0 {
			structuralPassed++
		} else {
			structuralIssues = append(structuralIssues, "No completed sessions found")
		}

		// Check 2: Templates present
		expectedDecTemplates := TemplatesForScope(ScopeDecision)
		missingDecTemplates := 0
		for _, tmpl := range expectedDecTemplates {
			if _, err := ws.Stat(filepath.Join(TemplatesDir, tmpl)); err != nil {
				missingDecTemplates++
			}
		}
		if missingDecTemplates == 0 && info.TemplateCount >= len(expectedDecTemplates) {
			structuralPassed++
		} else {
			structuralIssues = append(structuralIssues, fmt.Sprintf("%d/%d required templates missing", missingDecTemplates, len(expectedDecTemplates)))
		}

		// Check 3: Decision record exists
		hasDecision := info.HasDecisions
		if !hasDecision {
			if entries, err := ws.ListDir(ResearchDir); err == nil {
				for _, e := range entries {
					name := e.Name()
					if e.IsDir() || !strings.HasSuffix(name, ".md") {
						continue
					}
					if IsSpecialResearchFile(name) {
						continue
					}
					if strings.HasPrefix(strings.ToUpper(name), "D-") {
						hasDecision = true
						break
					}
					if dData, err := ws.ReadFile(filepath.Join(ResearchDir, name)); err == nil && len(dData) > 0 {
						if fm, _, err := ParseFrontmatter(dData); err == nil && fm != nil && fm.Has("door_type") {
							hasDecision = true
							break
						}
					}
				}
			}
		}
		if hasDecision {
			structuralPassed++
		} else {
			structuralIssues = append(structuralIssues, "No decision record found")
		}

	case ScopeComparison:
		structuralTotal = 1
		hasComp := false
		if entries, err := ws.ListDir(ResearchDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && (strings.HasSuffix(e.Name(), "-comparison.md") || strings.Contains(e.Name(), "-cmp-") || strings.EqualFold(e.Name(), "COMPARISON.md")) {
					hasComp = true
					break
				}
			}
		}
		if hasComp || info.SessionCount > 0 {
			structuralPassed++
		} else {
			structuralIssues = append(structuralIssues, "No comparison session found")
		}

	default: // ScopeProject
		structuralTotal = 5
		// Check 1: Pipeline exists
		if info.HasPipeline {
			structuralPassed++
		} else {
			structuralIssues = append(structuralIssues, "RESEARCH-PIPELINE.md not found")
		}

		// Check 2: DAG sessions completed
		if info.HasPipeline {
			if pipeData, err := ws.ReadFile(PipelineFile); err == nil {
				if dag, err := ParsePipeline(pipeData); err == nil && len(dag.Sessions) > 0 {
					completedCount := 0
					for _, s := range dag.Sessions {
						var sData []byte
						found := false
						if s.OutputFile != "" {
							if data, err := ws.ReadFile(s.OutputFile); err == nil && len(data) > 0 {
								sData = data
								found = true
							}
						}
						if !found {
							if entries, err := ws.ListDir(SessionsDir); err == nil {
								for _, e := range entries {
									if (strings.HasPrefix(e.Name(), s.ID+"-") || strings.HasPrefix(e.Name(), s.ID+"_") || strings.EqualFold(e.Name(), s.ID+".md")) && strings.HasSuffix(e.Name(), ".md") {
										if data, err := ws.ReadFile(filepath.Join(SessionsDir, e.Name())); err == nil && len(data) > 0 {
											sData = data
											found = true
											break
										}
									}
								}
							}
						}
						if !found && IsSynthesisSession(s.ID) {
							if data, err := ws.ReadFile(FADFile); err == nil && len(data) > 0 {
								sData = data
								found = true
							}
						}

						if !found {
							structuralIssues = append(structuralIssues, fmt.Sprintf("G2-SESSION-MISSING: Session %s has not been completed", s.ID))
							continue
						}

						// Check frontmatter status: draft is NOT complete
						fm, _, _ := ParseFrontmatter(sData)
						if fm != nil && strings.EqualFold(strings.TrimSpace(fm.GetString("status")), "draft") {
							structuralIssues = append(structuralIssues, fmt.Sprintf("G2-SESSION-DRAFT: Session %s is still in draft status", s.ID))
							continue
						}

						completedCount++
					}
					if completedCount == len(dag.Sessions) {
						structuralPassed++
					} else {
						structuralIssues = append(structuralIssues, fmt.Sprintf("G2-INCOMPLETE: Pipeline DAG incomplete: %d/%d sessions completed", completedCount, len(dag.Sessions)))
					}
				} else if info.SessionCount > 0 {
					structuralPassed++
				} else {
					structuralIssues = append(structuralIssues, "No completed sessions found")
				}
			} else if info.SessionCount > 0 {
				structuralPassed++
			} else {
				structuralIssues = append(structuralIssues, "No completed sessions found")
			}
		} else if info.SessionCount > 0 {
			structuralPassed++
		} else {
			structuralIssues = append(structuralIssues, "No completed sessions found")
		}

		// Check 3: Templates present
		expectedProjTemplates := TemplatesForScope(ScopeProject)
		missingProjTemplates := 0
		for _, tmpl := range expectedProjTemplates {
			if _, err := ws.Stat(filepath.Join(TemplatesDir, tmpl)); err != nil {
				missingProjTemplates++
			}
		}
		if missingProjTemplates == 0 && info.TemplateCount >= len(expectedProjTemplates) {
			structuralPassed++
		} else {
			structuralIssues = append(structuralIssues, fmt.Sprintf("%d/%d required templates missing", missingProjTemplates, len(expectedProjTemplates)))
		}

		// Check 4: Decisions registry or ADR files exist
		hasDecFile := false
		if _, err := ws.Stat(DecisionsFile); err == nil {
			hasDecFile = true
		}
		if hasDecFile || info.HasDecisions {
			structuralPassed++
		} else {
			structuralIssues = append(structuralIssues, "DECISIONS.md not found")
		}

		// Check 5: FAD exists
		if _, err := ws.Stat(FADFile); err == nil {
			structuralPassed++
		} else {
			structuralIssues = append(structuralIssues, "FAD.md not found — synthesis not complete")
		}
	}

	// Quality indicators (Track B)
	var trackBIssues []string
	var trackBPassed int
	var trackBTotal int
	var gateAdvisories []string

	switch info.Scope {
	case ScopeComparison:
		trackBTotal = 1
		hasGrades := false
		if entries, err := ws.ListDir(SessionsDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					if data, err := ws.ReadFile(filepath.Join(SessionsDir, e.Name())); err == nil {
						if EvidenceGradePattern.Match(data) {
							hasGrades = true
							break
						}
					}
				}
			}
		}
		if !hasGrades {
			if entries, err := ws.ListDir(ResearchDir); err == nil {
				for _, e := range entries {
					if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && (strings.HasSuffix(e.Name(), "-comparison.md") || strings.Contains(e.Name(), "-cmp-") || strings.HasPrefix(strings.ToUpper(e.Name()), "C-")) {
						if data, err := ws.ReadFile(filepath.Join(ResearchDir, e.Name())); err == nil {
							if EvidenceGradePattern.Match(data) {
								hasGrades = true
								break
							}
						}
					}
				}
			}
		}
		if hasGrades || info.SessionCount > 0 {
			trackBPassed++
		} else {
			trackBIssues = append(trackBIssues, "No evidence grades found in comparison session")
		}

	case ScopeDecision:
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
		if _, err := ws.Stat(FADFile); err == nil {
			fadData, err := ws.ReadFile(FADFile)
			if err == nil {
				fadValidation := ValidateFAD(fadData)
				if fadValidation.WarningCount() == 0 {
					trackBPassed++
				} else {
					for _, issue := range fadValidation.Issues {
						if issue.Level == L3Warn {
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

	// Evidence integrity checks (B3/B4/B5/B11) — strict blocking for One-Way Doors
	if info.Scope == ScopeProject || info.Scope == ScopeDecision {
		trackBTotal++
		eiBlocking, eiAdvisories := verifyEvidentiaryIntegrity(ws, info)
		if len(eiBlocking) == 0 {
			trackBPassed++
		} else {
			trackBIssues = append(trackBIssues, eiBlocking...)
		}
		gateAdvisories = append(gateAdvisories, eiAdvisories...)
	}

	decisions := CollectGateDecisions(ws)
	trackAChecks := evaluateTrackAChecks(info, decisions)

	// Overall gate result
	structuralPass := len(structuralIssues) == 0
	trackAPass := trackAChecks.Passed
	trackBPass := len(trackBIssues) == 0

	var gatePassed bool
	switch info.Scope {
	case ScopeComparison:
		gatePassed = true
		gateAdvisories = append(gateAdvisories, "Comparison scope has zero gate requirements. Research is complete upon WEP matrix generation.")
	case ScopeDecision:
		hasOneWay := false
		for _, d := range decisions {
			if strings.EqualFold(d.DoorType, "one-way") || strings.EqualFold(d.DoorType, "one_way") {
				hasOneWay = true
				break
			}
		}
		if hasOneWay {
			gatePassed = trackAPass && trackBPass && structuralPass
		} else {
			// For purely Two-Way Door decisions, Track A reversibility confirmation is sufficient
			gatePassed = trackAPass && structuralPass
			if len(trackBIssues) > 0 {
				for _, issue := range trackBIssues {
					gateAdvisories = append(gateAdvisories, "Track B (advisory for two-way door): "+issue)
				}
				trackBIssues = nil
			}
		}
	default: // ScopeProject
		gatePassed = trackAPass && trackBPass && structuralPass
	}

	var gateStatus string
	if gatePassed {
		gateStatus = "PASS"
	} else if structuralPass {
		gateStatus = "WARN"
	} else {
		gateStatus = "FAIL"
	}

	var warnings []string
	for _, issue := range structuralIssues {
		warnings = append(warnings, "GATE-STRUCTURAL: "+issue)
	}
	for _, issue := range trackAChecks.Issues {
		warnings = append(warnings, "GATE-TRACK-A: "+issue)
	}
	for _, issue := range trackBIssues {
		warnings = append(warnings, "GATE-QUALITY: "+issue)
	}
	for _, adv := range gateAdvisories {
		warnings = append(warnings, "GATE-ADVISORY: "+adv)
	}

	var nextStep string
	if info.Scope == ScopeComparison {
		nextStep = "Comparison scope complete. WEP matrix generated. Proceed to implementation."
		gateStatus = "PASS"
	} else {
		switch gateStatus {
		case "PASS":
			nextStep = "Gate PASSED! The research phase is complete. All architectural decisions and evidence gates satisfied. Proceed to repository implementation and code scaffolding."
		case "WARN":
			nextStep = "Gate quality checks incomplete. Resolve the listed quality warnings before proceeding to code scaffolding."
		case "FAIL":
			nextStep = "Gate FAILED. All blocking issues must be resolved before proceeding to code. Fix the listed issues and re-run vivechak_run_gate."
		}
	}

	return &GateResult{
		Status: gateStatus,
		Passed: gatePassed,
		StructuralChecks: TrackResult{
			Label:  "Structural Completeness",
			Passed: structuralPass,
			Score:  fmt.Sprintf("%d/%d", structuralPassed, structuralTotal),
			Count:  structuralPassed,
			Total:  structuralTotal,
			Issues: structuralIssues,
		},
		TrackAChecks: trackAChecks,
		TrackBChecks: TrackResult{
			Label:  "Quality Indicators (Mechanical)",
			Passed: trackBPass,
			Score:  fmt.Sprintf("%d/%d", trackBPassed, trackBTotal),
			Count:  trackBPassed,
			Total:  trackBTotal,
			Issues: trackBIssues,
		},
		QualityChecks: TrackResult{
			Label:  "Quality Indicators (Mechanical)",
			Passed: trackBPass,
			Score:  fmt.Sprintf("%d/%d", trackBPassed, trackBTotal),
			Count:  trackBPassed,
			Total:  trackBTotal,
			Issues: trackBIssues,
		},
		Advisories: gateAdvisories,
		Warnings:   warnings,
		NextStep:   nextStep,
		Decisions:  decisions,
	}
}

// evaluateTrackAChecks mechanically validates Two-Way Door decisions (Checks A1–A4).
// Two-way doors are fast-track decisions requiring reversibility confirmation and review triggers.
func evaluateTrackAChecks(info WorkspaceInfo, decisions []GateDecisionItem) TrackResult {
	if info.SessionCount == 0 && !info.HasDecisions && !info.HasPipeline {
		return TrackResult{
			Label:  "Track A: Two-Way Doors",
			Passed: false,
			Score:  "0/1",
			Count:  0,
			Total:  1,
			Issues: []string{"A1-LOGGED: Workspace is empty (no sessions, pipeline, or decisions recorded)"},
		}
	}

	var issues []string

	twoWayCount := 0
	passedCount := 0

	for _, d := range decisions {
		if !strings.EqualFold(d.DoorType, "two-way") {
			continue
		}
		twoWayCount++
		// A1: Logged in ADR with status accepted or pass
		if !strings.EqualFold(d.Status, "accepted") && !strings.EqualFold(d.Status, "pass") {
			issues = append(issues, fmt.Sprintf("A1-STATUS: Two-Way Door %s has status %q (must be accepted or pass)", d.ID, d.Status))
			continue
		}
		passedCount++
	}

	passed := len(issues) == 0
	score := fmt.Sprintf("%d/%d", passedCount, twoWayCount)
	if twoWayCount == 0 {
		score = "0/0"
		passed = true
	}

	return TrackResult{
		Label:  "Track A: Two-Way Doors",
		Passed: passed,
		Score:  score,
		Count:  passedCount,
		Total:  twoWayCount,
		Issues: issues,
	}
}

// verifyDecisionsMechanical checks ADR files for mechanical correctness:
// valid frontmatter, accepted status, and reversal triggers on one-way door decisions.
func verifyDecisionsMechanical(ws WorkspaceFS, info WorkspaceInfo) (bool, []string, []string) {
	hasDecFile := false
	if _, err := ws.Stat(DecisionsFile); err == nil {
		hasDecFile = true
	}
	if !info.HasDecisions && !hasDecFile {
		return false, []string{"Cannot check decisions — DECISIONS.md not found"}, nil
	}

	type decisionSource struct {
		file string
		data []byte
	}
	var sources []decisionSource

	scanDir := func(dir string) {
		if entries, err := ws.ListDir(dir); err == nil {
			for _, e := range entries {
				name := e.Name()
				if e.IsDir() || !strings.HasSuffix(name, ".md") {
					continue
				}
				if IsSpecialResearchFile(name) {
					continue
				}
				relPath := filepath.Join(dir, name)
				isCandidate := strings.HasPrefix(strings.ToUpper(name), "D-")
				dData, err := ws.ReadFile(relPath)
				if err != nil || len(dData) == 0 {
					continue
				}
				if !isCandidate {
					if fm, _, err := ParseFrontmatter(dData); err == nil && fm != nil {
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
	}

	scanDir(ResearchDir)
	scanDir(DecisionsDir)

	// Scan consolidated DECISIONS.md alongside standalone files.
	// processedIDs guarantees standalone files take precedence on ID collisions.
	if decData, err := ws.ReadFile(DecisionsFile); err == nil && len(decData) > 0 {
		sources = append(sources, decisionSource{file: DecisionsFile, data: decData})
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

	checkDecisionFM := func(fm Frontmatter, sourceFile string) {
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
		upperID := strings.ToUpper(id)
		if processedIDs[upperID] {
			return
		}
		processedIDs[upperID] = true
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
			if fm, _, err := ParseFrontmatter(m[2]); err == nil && fm != nil {
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
				if fm, _, err := ParseFrontmatter(m[0]); err == nil && fm != nil {
					checkDecisionFM(fm, file)
					parsedAny = true
				}
			}
		}

		// 3. Document-level frontmatter (e.g. standalone ADR files)
		if !parsedAny {
			if fm, _, err := ParseFrontmatter(chunk); err == nil && fm != nil {
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

// verifyEvidentiaryIntegrity checks B3 (evidence threshold), B4 (verification integrity), B5 (rejected alternatives), and ADR structural shape.
// Returns (blockingIssues, advisories).
func verifyEvidentiaryIntegrity(ws WorkspaceFS, info WorkspaceInfo) ([]string, []string) {
	var blocking []string
	var advisories []string

	decisions := CollectGateDecisions(ws)
	lowGradeRe := regexp.MustCompile(`(?i)(?:\[(?:Grade\s*[:\s]\s*)[CDE][^\]\n]*\]|\((?:Grade\s*[:\s]\s*)[CDE][^)\n]*\)|\bGrade\s*[:\s]\s*[CDE]\b|\[[CDE]\s*[·|:,][^\]\n]*\]|\([CDE]\s*[·|:,][^)\n]*\))`)

	// G4: ADR parity check - verify that all decisions referenced by sessions actually exist
	decisionIDs := make(map[string]bool)
	for _, d := range decisions {
		decisionIDs[strings.ToUpper(strings.TrimSpace(d.ID))] = true
	}
	if sEntries, err := ws.ListDir(SessionsDir); err == nil {
		for _, se := range sEntries {
			if se.IsDir() || !strings.HasSuffix(se.Name(), ".md") {
				continue
			}
			sData, err := ws.ReadFile(filepath.Join(SessionsDir, se.Name()))
			if err != nil {
				continue
			}
			sFm, _, _ := ParseFrontmatter(sData)
			if sFm == nil {
				continue
			}
			var referencedDecs []string
			if sFm.Has("informs_decisions") {
				referencedDecs = append(referencedDecs, sFm.GetStringSlice("informs_decisions")...)
			}
			if sFm.Has("informs") {
				referencedDecs = append(referencedDecs, sFm.GetStringSlice("informs")...)
			}
			for _, ref := range referencedDecs {
				refUpper := strings.ToUpper(strings.TrimSpace(ref))
				if refUpper != "" && !decisionIDs[refUpper] {
					blocking = append(blocking, fmt.Sprintf(
						"G4-ADR-PARITY: Session %s references decision %s in informs_decisions, but %s has no corresponding ADR file or record",
						ExtractSessionID(se.Name(), sFm), ref, ref))
				}
			}
		}
	}

	for _, d := range decisions {
		if !strings.EqualFold(d.DoorType, "one-way") {
			continue
		}

		// Read decision file from DecisionsDir or ResearchDir
		var decData []byte
		if dData, err := ws.ReadFile(filepath.Join(DecisionsDir, d.ID+".md")); err == nil {
			decData = dData
		} else if dData, err := ws.ReadFile(filepath.Join(DecisionsDir, d.ID+"-decision.md")); err == nil {
			decData = dData
		} else if dData, err := ws.ReadFile(filepath.Join(ResearchDir, d.ID+".md")); err == nil {
			decData = dData
		} else if dData, err := ws.ReadFile(filepath.Join(ResearchDir, d.ID+"-decision.md")); err == nil {
			decData = dData
		} else if entries, lErr := ws.ListDir(DecisionsDir); lErr == nil {
			for _, e := range entries {
				name := e.Name()
				if (strings.HasPrefix(name, d.ID+"-") || strings.HasPrefix(name, d.ID+"_") || strings.EqualFold(name, d.ID+".md")) && strings.HasSuffix(name, ".md") && !IsSpecialResearchFile(name) {
					decData, _ = ws.ReadFile(filepath.Join(DecisionsDir, name))
					break
				}
			}
		}
		if len(decData) == 0 {
			if entries, lErr := ws.ListDir(ResearchDir); lErr == nil {
				for _, e := range entries {
					name := e.Name()
					if (strings.HasPrefix(name, d.ID+"-") || strings.HasPrefix(name, d.ID+"_") || strings.EqualFold(name, d.ID+".md")) && strings.HasSuffix(name, ".md") && !IsSpecialResearchFile(name) {
						decData, _ = ws.ReadFile(filepath.Join(ResearchDir, name))
						break
					}
				}
			}
		}

		// Fallback to DECISIONS.md if no individual file
		if len(decData) == 0 {
			if decAll, err := ws.ReadFile(DecisionsFile); err == nil {
				anchoredRe := regexp.MustCompile(fmt.Sprintf(`(?s)<!-- DECISION:\s*%s\s*-->\s*(.*?)\s*<!-- /DECISION:\s*%s\s*-->`, regexp.QuoteMeta(d.ID), regexp.QuoteMeta(d.ID)))
				if m := anchoredRe.FindSubmatch(decAll); len(m) > 1 {
					decData = m[1]
				}
			}
		}

		if len(decData) == 0 {
			continue
		}

		fm, body, _ := ParseFrontmatter(decData)
		bodyStr := string(body)

		// B5: Check rejected alternatives section (>30 characters of content)
		altContent := ExtractSection(bodyStr, "alternative", "rejected")
		if len(strings.TrimSpace(altContent)) < 30 {
			blocking = append(blocking, fmt.Sprintf(
				"B5-ALTERNATIVES: One-way door %s lacks documented rejected alternatives section (minimum 30 chars required). "+
					"(help: edit ADR to include substantive '## Alternatives Considered' section)", d.ID))
		}

		// B-DECISION-STUB: Structural check replacing naive character volume check (advisory)
		lowerBody := strings.ToLower(bodyStr)
		hasContext := strings.Contains(lowerBody, "context") || strings.Contains(lowerBody, "problem")
		hasRationale := strings.Contains(lowerBody, "rationale") || strings.Contains(lowerBody, "decision") || strings.Contains(lowerBody, "verdict")
		hasTriggers := strings.Contains(lowerBody, "trigger") || strings.Contains(lowerBody, "reversal") || (fm != nil && fm.GetString("review_trigger") != "")

		if !hasContext || !hasRationale || !hasTriggers {
			advisories = append(advisories, fmt.Sprintf(
				"B-DECISION-STUB: One-way door %s lacks mandatory structural sections (Context, Rationale, or Reversal Trigger)", d.ID))
		}

		// Find informing sessions from frontmatter
		var sessIDs []string
		if fm != nil {
			sessIDs = fm.GetStringSlice("informed_by_sessions")
		}
		// Also scan session files for informs_decisions if empty
		if len(sessIDs) == 0 {
			if sEntries, sErr := ws.ListDir(SessionsDir); sErr == nil {
				for _, se := range sEntries {
					if se.IsDir() || !strings.HasSuffix(se.Name(), ".md") {
						continue
					}
					sData, _ := ws.ReadFile(filepath.Join(SessionsDir, se.Name()))
					if sFm, _, _ := ParseFrontmatter(sData); sFm != nil {
						infDecs := sFm.GetStringSlice("informs_decisions")
						for _, inf := range infDecs {
							if strings.EqualFold(inf, d.ID) {
								sessIDs = append(sessIDs, ExtractSessionID(se.Name(), sFm))
							}
						}
					}
				}
			}
		}

		// B11: Check red-team challenge verification
		checkChallengeText := func(text string) bool {
			lower := strings.ToLower(text)
			return strings.Contains(lower, "## challenge findings") ||
				strings.Contains(lower, "## red-team stress test") ||
				strings.Contains(lower, "## red team stress test") ||
				strings.Contains(lower, "challenge findings") ||
				strings.Contains(lower, "red-team stress test")
		}
		hasChallenge := checkChallengeText(bodyStr)

		for _, sid := range sessIDs {
			var sData []byte
			if sd, err := ws.ReadFile(filepath.Join(SessionsDir, sid+".md")); err == nil {
				sData = sd
			} else if sEntries, lErr := ws.ListDir(SessionsDir); lErr == nil {
				for _, se := range sEntries {
					name := strings.TrimSuffix(se.Name(), ".md")
					nameUpper := strings.ToUpper(name)
					sidUpper := strings.ToUpper(sid)
					if nameUpper == sidUpper || strings.HasPrefix(nameUpper, sidUpper+"-") || strings.HasPrefix(nameUpper, sidUpper+"_") {
						sData, _ = ws.ReadFile(filepath.Join(SessionsDir, se.Name()))
						break
					}
				}
			}
			if len(sData) == 0 {
				continue
			}
			sessStr := string(sData)
			if !hasChallenge && checkChallengeText(sessStr) {
				hasChallenge = true
			}
			checkStr := extractRecommendationText(sessStr)
			if len(strings.TrimSpace(checkStr)) < 20 {
				advisories = append(advisories, fmt.Sprintf(
					"W-GATE-SECTION: Session %s recommendation section could not be isolated for evidentiary audit; bibliography scan omitted to prevent false B3 failure.", sid))
				continue
			}

			// B3: Check evidence grades supporting the recommendation
			lowGrades := len(lowGradeRe.FindAllString(checkStr, -1))
			if lowGrades > 0 {
				blocking = append(blocking, fmt.Sprintf(
					"B3-EVIDENTIARY: Session %s informs one-way door %s but recommendation has %d Grade C/D/E citation(s). "+
						"One-way decisions require Grade A/B evidence. (help: re-verify with web search and update via vivechak_amend_session(session_id=%q))",
					sid, d.ID, lowGrades, sid))
			}

			// B4: Check recalled verification supporting the recommendation
			recalledCount := len(recalledCitationRe.FindAllString(checkStr, -1))
			if recalledCount > 0 {
				blocking = append(blocking, fmt.Sprintf(
					"B4-VERIFICATION: Session %s informs one-way door %s but recommendation has %d 'recalled' citation(s). "+
						"One-way decisions require fetched/cached verification. (help: fetch primary evidence and update via vivechak_amend_session(session_id=%q))",
					sid, d.ID, recalledCount, sid))
			}
		}

		if !hasChallenge {
			advisories = append(advisories, fmt.Sprintf(
				"W-OWD-NO-CHALLENGE: One-Way Door decision %s was not subjected to adversarial red-team challenge before commitment.", d.ID))
		}
	}
	return blocking, advisories
}

// extractRecommendationText extracts text under Recommendation/Verdict/Findings in a session,
// or the session body excluding Sources/Bibliography/Ledger to avoid flagging secondary links.
func extractRecommendationText(content string) string {
	rec := ExtractSection(content, "recommend", "verdict", "decision")
	if len(strings.TrimSpace(rec)) >= 20 {
		return rec
	}
	findings := ExtractSection(content, "key finding", "finding")
	if len(strings.TrimSpace(findings)) >= 20 {
		return findings
	}
	// Fallback: use body but strip Sources / Bibliography / Evidence Ledger
	lines := strings.Split(content, "\n")
	var kept []string
	inSources := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "source") || strings.Contains(lower, "reference") ||
				strings.Contains(lower, "bibliography") || strings.Contains(lower, "evidence ledger") {
				inSources = true
				continue
			} else if inSources && strings.HasPrefix(trimmed, "## ") {
				inSources = false
			}
		}
		if !inSources {
			kept = append(kept, l)
		}
	}
	return strings.Join(kept, "\n")
}
