package core

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	GateMarkerBegin = "<!-- vivechak:begin -->"
	GateMarkerEnd   = "<!-- vivechak:end -->"
)

// DetectProjectName extracts the project name from RESEARCH-PIPELINE.md or falls back to workspace folder name.
func DetectProjectName(ws WorkspaceFS, root string) string {
	if ws != nil {
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
	}
	base := filepath.Base(root)
	if base != "" && base != "." && base != "/" && base != "\\" {
		return base
	}
	return "Project"
}

// RenderGateArtifact renders the PHASE-0-GATE.md artifact using the template and gate results (backward compatibility for test suites).
func RenderGateArtifact(ws WorkspaceFS, tmplBytes []byte, root, projectName, gateStatus string, trackAPass, trackBPass bool, trackAIssues, trackBIssues []string, decisions []GateDecisionItem, advisories ...string) []byte {
	if len(tmplBytes) == 0 {
		return nil
	}

	tmpl := strings.ReplaceAll(string(tmplBytes), "\r\n", "\n")
	today := time.Now().UTC().Format("2006-01-02")

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

	if trackBPass {
		tmpl = strings.Replace(tmpl, "- [ ] All required dependency paths", "- [x] All required dependency paths", 1)
		tmpl = strings.Replace(tmpl, "- [ ] No orphaned sessions remain", "- [x] No orphaned sessions remain", 1)
		tmpl = strings.Replace(tmpl, "- [ ] All cross-model divergences resolved", "- [x] All cross-model divergences resolved", 1)
		tmpl = strings.Replace(tmpl, "- [ ] No unresolved `contested` corroboration flags", "- [x] No unresolved `contested` corroboration flags", 1)

		hasB3Adv := false
		hasB4Adv := false
		hasB5Adv := false
		hasB10Adv := false
		hasB11Adv := false
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
			if strings.Contains(adv, "B10") || strings.Contains(adv, "B-DECISION-STUB") {
				hasB10Adv = true
			}
			if strings.Contains(adv, "B11") || strings.Contains(adv, "W-OWD-NO-CHALLENGE") {
				hasB11Adv = true
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
		if !hasB10Adv {
			tmpl = strings.Replace(tmpl, "- [ ] Decision record contains substantive causal rationale", "- [x] Decision record contains substantive causal rationale", 1)
		}
		if !hasB11Adv {
			tmpl = strings.Replace(tmpl, "- [ ] One-Way Door decisions subjected to adversarial challenge", "- [x] One-Way Door decisions subjected to adversarial challenge", 1)
		}
		tmpl = strings.Replace(tmpl, "- [ ] Every locked ADR contains an explicit `review_trigger`", "- [x] Every locked ADR contains an explicit `review_trigger`", 1)
		tmpl = strings.Replace(tmpl, "- [ ] Review triggers are specific and measurable", "- [x] Review triggers are specific and measurable", 1)

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
		if _, statErr := ws.Stat(FADFile); statErr == nil {
			tmpl = strings.Replace(tmpl, "- [ ] FAD compiled with full traceability", "- [x] FAD compiled with full traceability", 1)
			tmpl = strings.Replace(tmpl, "- [ ] FAD committed to repository root", "- [x] FAD committed to repository root", 1)
			tmpl = strings.Replace(tmpl, "- [ ] Repository scaffolding ready to generate", "- [x] Repository scaffolding ready to generate", 1)
		}
	}
	tmpl = strings.Replace(tmpl, "**Track B Result:** `[PASS / FAIL]`", fmt.Sprintf("**Track B Result:** %s", trackBResultStr), 1)

	return []byte(tmpl)
}

// RenderGateReport renders the honest Phase 0 gate report according to CRIT-12 & CRIT-13.
func RenderGateReport(ws WorkspaceFS, root, projectName string, res *GateResult) []byte {
	today := time.Now().UTC().Format("2006-01-02")
	target := res.Target
	if target == "" {
		target = "project"
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("title: \"Phase 0 Pre-Codebase Exit Gate Report\"\n")
	fmt.Fprintf(&b, "project: %q\n", projectName)
	fmt.Fprintf(&b, "date: %s\n", today)
	fmt.Fprintf(&b, "verdict: %s\n", res.Verdict)
	b.WriteString("status: complete\n")
	b.WriteString("vivechak: 1\n")
	b.WriteString("schema_version: \"0.1.0\"\n")
	b.WriteString("---\n\n")

	fmt.Fprintf(&b, "# Phase 0 Pre-Codebase Exit Gate: %s\n\n", projectName)
	fmt.Fprintf(&b, "**Gate Date:** %s  \n", today)
	fmt.Fprintf(&b, "**Target:** %s  \n", target)
	fmt.Fprintf(&b, "**Verdict:** **%s**\n\n", res.Verdict)

	b.WriteString(GateMarkerBegin + "\n")
	b.WriteString("## Automated Mechanical Rule Evaluation (G1–G12)\n\n")
	b.WriteString("| Rule | Status | Details |\n")
	b.WriteString("|---|---|---|\n")

	for _, r := range res.Rules {
		cleanReason := strings.ReplaceAll(r.Reason, "|", "\\|")
		cleanReason = strings.ReplaceAll(cleanReason, "\n", " ")
		fmt.Fprintf(&b, "| %s | %s | %s |\n", r.ID, r.Status, cleanReason)
	}

	b.WriteString("\n### Evidence Provenance Note\n")
	b.WriteString("Evidence grades and verification labels are as reported by the agent; Vivechak did not fetch or verify sources. Labels as reported; sources not fetched or verified.\n\n")

	b.WriteString("### Human checks (not evaluated by Vivechak)\n")
	b.WriteString("- ☐ not checked (human): source verification and evidence quality\n")
	b.WriteString("- ☐ not checked (human): premortem\n")
	b.WriteString("- ☐ not checked (human): contradiction resolution\n")
	b.WriteString(GateMarkerEnd + "\n")

	return []byte(b.String())
}

// MergeGateArtifactContent merges newly rendered gate content into an existing gate file,
// preserving any human notes or sign-offs outside the <!-- vivechak:begin --> and <!-- vivechak:end --> markers.
func MergeGateArtifactContent(existingContent, newGeneratedContent []byte) []byte {
	if len(existingContent) == 0 {
		return newGeneratedContent
	}

	existingStr := string(existingContent)
	endIdx := strings.Index(existingStr, GateMarkerEnd)

	if endIdx != -1 {
		suffix := existingStr[endIdx+len(GateMarkerEnd):]
		var sb strings.Builder
		sb.Write(newGeneratedContent)
		if !bytes.HasSuffix(newGeneratedContent, []byte("\n")) {
			sb.WriteString("\n")
		}
		if strings.TrimSpace(suffix) != "" {
			sb.WriteString(suffix)
			if !strings.HasSuffix(suffix, "\n") {
				sb.WriteString("\n")
			}
		}
		return []byte(sb.String())
	}

	// No markers found in existingContent: preserve any human lines (like Signed off by:) as suffix
	var humanSuffix strings.Builder
	for _, line := range strings.Split(existingStr, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Signed off by:") || strings.HasPrefix(trimmed, "HUMAN NOTE:") || strings.HasPrefix(trimmed, "Human Note:") {
			humanSuffix.WriteString("\n")
			humanSuffix.WriteString(line)
		}
	}

	var sb strings.Builder
	sb.Write(newGeneratedContent)
	if !bytes.HasSuffix(newGeneratedContent, []byte("\n")) {
		sb.WriteString("\n")
	}
	if humanSuffix.Len() > 0 {
		sb.WriteString(humanSuffix.String())
		sb.WriteString("\n")
	}
	return []byte(sb.String())
}
