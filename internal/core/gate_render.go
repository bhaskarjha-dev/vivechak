package core

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// DetectProjectName extracts the project name from RESEARCH-PIPELINE.md or falls back to workspace folder name.
func DetectProjectName(ws WorkspaceFS, root string) string {
	if data, err := ws.ReadFile(PipelineFile); err == nil {
		if fm, _, err := ParseFrontmatter(data); err == nil && fm != nil {
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

// RenderGateArtifact renders the PHASE-0-GATE.md artifact using the template and gate results.
func RenderGateArtifact(ws WorkspaceFS, tmplBytes []byte, root, projectName, gateStatus string, trackAPass, trackBPass bool, trackAIssues, trackBIssues []string, decisions []GateDecisionItem, advisories ...string) []byte {
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
	hasTwoWay := false
	allTwoWayAccepted := true
	for _, d := range decisions {
		if strings.EqualFold(d.DoorType, "two-way") {
			hasTwoWay = true
			if !strings.EqualFold(d.Status, "PASS") && !strings.EqualFold(d.Status, "ACCEPTED") {
				allTwoWayAccepted = false
			}
		}
	}

	if !hasTwoWay {
		tmpl = strings.Replace(tmpl, "**Track A Result:** `[PASS / FAIL]`", "**Track A Result:** PASS (0 two-way doors)", 1)
	} else if trackAPass && allTwoWayAccepted {
		tmpl = strings.Replace(tmpl, "- [ ] Decision logged in ADR", "- [x] Decision logged in ADR", 1)
		tmpl = strings.Replace(tmpl, "- [ ] Reversibility confirmed", "- [x] Reversibility confirmed", 1)
		tmpl = strings.Replace(tmpl, "- [ ] At least one corroborated source", "- [x] At least one corroborated source", 1)
		tmpl = strings.Replace(tmpl, "**Track A Result:** `[PASS / FAIL]`", "**Track A Result:** PASS", 1)
	} else {
		tmpl = strings.Replace(tmpl, "**Track A Result:** `[PASS / FAIL]`", "**Track A Result:** FAIL", 1)
	}

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
			if strings.Contains(adv, "B5-ALTERNATIVES") || strings.Contains(adv, "B-DECISION-STUB") {
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

		// B7: Premortem Protocol (check if FAD documents substantive failure scenarios)
		if fadBytes, fErr := ws.ReadFile(FADFile); fErr == nil {
			fadStr := strings.ToLower(string(fadBytes))
			hasSubstantivePremortem := false
			if strings.Contains(fadStr, "premortem") || strings.Contains(fadStr, "failure scenario") || strings.Contains(fadStr, "risk register") {
				if !strings.Contains(fadStr, "[risk 1]") && !strings.Contains(fadStr, "[action]") && len(fadStr) > 300 {
					hasSubstantivePremortem = true
				}
			}
			if hasSubstantivePremortem {
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
		if _, statErr := ws.Stat(FADFile); statErr == nil {
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
