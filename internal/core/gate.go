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
	StructuralChecks TrackResult        `json:"structural_checks"`
	QualityChecks    TrackResult        `json:"quality_checks"`
	Advisories       []string           `json:"advisories,omitempty"`
	Warnings         []string           `json:"warnings,omitempty"`
	NextStep         string             `json:"next_step"`
	Decisions        []GateDecisionItem `json:"decisions,omitempty"`
}

// EvaluateGate performs mechanical verification of structural completeness and quality indicators.
func EvaluateGate(ws WorkspaceFS, info WorkspaceInfo, root string) *GateResult {
	// Structural completeness checks (Track A)
	var trackAIssues []string
	var trackAPassed int
	var trackATotal int

	switch info.Scope {
	case ScopeDecision:
		trackATotal = 3
		// Check 1: At least 1 session completed
		if info.SessionCount > 0 {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "No completed sessions found")
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
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, fmt.Sprintf("%d/%d required templates missing", missingDecTemplates, len(expectedDecTemplates)))
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
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "No decision record found")
		}

	case ScopeComparison:
		trackATotal = 2
		// Check 1: At least 1 session completed
		if info.SessionCount > 0 {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "No completed sessions found")
		}

		// Check 2: Templates present
		expectedTemplates := len(TemplatesForScope(ScopeComparison))
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
			if pipeData, err := ws.ReadFile(PipelineFile); err == nil {
				if dag, err := ParsePipeline(pipeData); err == nil && len(dag.Sessions) > 0 {
					completedCount := 0
					for _, s := range dag.Sessions {
						found := false
						if s.OutputFile != "" {
							if _, err := ws.Stat(s.OutputFile); err == nil {
								found = true
							}
						}
						if !found {
							if entries, err := ws.ListDir(SessionsDir); err == nil {
								for _, e := range entries {
									if strings.HasPrefix(e.Name(), s.ID) && strings.HasSuffix(e.Name(), ".md") {
										found = true
										break
									}
								}
							}
						}
						if !found && IsSynthesisSession(s.ID) {
							if _, err := ws.Stat(FADFile); err == nil {
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
		expectedProjTemplates := TemplatesForScope(ScopeProject)
		missingProjTemplates := 0
		for _, tmpl := range expectedProjTemplates {
			if _, err := ws.Stat(filepath.Join(TemplatesDir, tmpl)); err != nil {
				missingProjTemplates++
			}
		}
		if missingProjTemplates == 0 && info.TemplateCount >= len(expectedProjTemplates) {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, fmt.Sprintf("%d/%d required templates missing", missingProjTemplates, len(expectedProjTemplates)))
		}

		// Check 4: Decisions exist
		if info.HasDecisions {
			trackAPassed++
		} else {
			trackAIssues = append(trackAIssues, "DECISIONS.md not found")
		}

		// Check 5: FAD exists
		if _, err := ws.Stat(FADFile); err == nil {
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
	case ScopeComparison:
		trackBTotal = 2
		// Check B1: Session count
		if info.SessionCount >= 1 {
			trackBPassed++
		} else {
			trackBIssues = append(trackBIssues, "No comparison session found")
		}
		// Check B2: Evidence grades in session files
		hasGrades := false
		if entries, err := ws.ListDir(SessionsDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					if data, err := ws.ReadFile(filepath.Join(SessionsDir, e.Name())); err == nil {
						v := ValidateSession(data)
						hasNoGrades := false
						for _, issue := range v.Issues {
							if issue.Code == "W-NO-EVIDENCE-GRADES" {
								hasNoGrades = true
								break
							}
						}
						if !hasNoGrades {
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

	// Evidence integrity checks (B3/B4/B5) — strict blocking for One-Way Doors
	if info.Scope == ScopeProject || info.Scope == ScopeDecision {
		eiBlocking, eiAdvisories := verifyEvidentiaryIntegrity(ws, info)
		trackBIssues = append(trackBIssues, eiBlocking...)
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
		nextStep = "Gate PASSED! The research phase is complete. All architectural decisions and evidence gates satisfied. Proceed to repository implementation and code scaffolding."
	case "WARN":
		nextStep = "Gate passed with quality advisories. Review the warnings above before proceeding to code scaffolding."
	case "FAIL":
		nextStep = "Gate FAILED. All blocking issues must be resolved before proceeding to code. Fix the listed issues and re-run vivechak_run_gate."
	}

	decisions := CollectGateDecisions(ws)

	return &GateResult{
		Status: gateStatus,
		Passed: gatePassed,
		StructuralChecks: TrackResult{
			Label:  "Structural Completeness",
			Passed: trackAPass,
			Score:  fmt.Sprintf("%d/%d", trackAPassed, trackATotal),
			Count:  trackAPassed,
			Total:  trackATotal,
			Issues: trackAIssues,
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

// verifyDecisionsMechanical checks ADR files for mechanical correctness:
// valid frontmatter, accepted status, and reversal triggers on one-way door decisions.
func verifyDecisionsMechanical(ws WorkspaceFS, info WorkspaceInfo) (bool, []string, []string) {
	if !info.HasDecisions {
		return false, []string{"Cannot check decisions — DECISIONS.md not found"}, nil
	}

	type decisionSource struct {
		file string
		data []byte
	}
	var sources []decisionSource

	if entries, err := ws.ListDir(ResearchDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") {
				continue
			}
			if IsSpecialResearchFile(name) {
				continue
			}
			relPath := filepath.Join(ResearchDir, name)
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
	lowGradeRe := regexp.MustCompile(`(?i)(?:\[(?:Grade\s+)?[CDE][^\]\n]*\]|\((?:Grade\s+)?[CDE][^)\n]*\)|\bGrade\s+[CDE]\b|\bGrade:[CDE]\b)`)

	for _, d := range decisions {
		if !strings.EqualFold(d.DoorType, "one-way") {
			continue
		}

		// Read decision file
		var decData []byte
		if dData, err := ws.ReadFile(filepath.Join(ResearchDir, d.ID+".md")); err == nil {
			decData = dData
		} else if dData, err := ws.ReadFile(filepath.Join(ResearchDir, d.ID+"-decision.md")); err == nil {
			decData = dData
		} else if entries, lErr := ws.ListDir(ResearchDir); lErr == nil {
			for _, e := range entries {
				name := e.Name()
				if (strings.HasPrefix(name, d.ID+"-") || strings.HasPrefix(name, d.ID+"_") || strings.EqualFold(name, d.ID+".md")) && strings.HasSuffix(name, ".md") && !IsSpecialResearchFile(name) {
					decData, _ = ws.ReadFile(filepath.Join(ResearchDir, name))
					break
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
