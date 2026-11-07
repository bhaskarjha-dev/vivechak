package core

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// GateRuleResult represents the evaluation result of a single gate rule G1–G12.
type GateRuleResult struct {
	ID     string `json:"id"`
	Status string `json:"status"` // "PASS", "FAIL", or "n/a"
	Reason string `json:"reason"`
}

// TrackResult represents the status and score of a gate track (Track A or Track B).
type TrackResult struct {
	Label  string   `json:"label,omitempty"`
	Passed bool     `json:"passed"`
	Score  string   `json:"score,omitempty"`
	Count  int      `json:"count"`
	Total  int      `json:"total"`
	Issues []string `json:"issues,omitempty"`
}

// GateResult holds the unified evaluation result from Phase 0 Gate.
type GateResult struct {
	Verdict    string             `json:"verdict"` // "PASS", "FAIL", or "not_applicable"
	Target     string             `json:"target"`
	Rules      []GateRuleResult   `json:"rules"`
	Advisories []string           `json:"advisories,omitempty"`
	Warnings   []string           `json:"warnings,omitempty"`
	NextStep   string             `json:"next_step"`
	Decisions  []GateDecisionItem `json:"decisions,omitempty"`

	// Backward compatibility fields for internal test fixtures
	Status           string      `json:"status,omitempty"`
	Passed           bool        `json:"passed,omitempty"`
	StructuralChecks TrackResult `json:"-"`
	TrackAChecks     TrackResult `json:"-"`
	TrackBChecks     TrackResult `json:"-"`
	QualityChecks    TrackResult `json:"-"`
}

// ADRRecord holds parsed ADR data for gate verification.
type adrRecord struct {
	ID            string
	Title         string
	DoorType      string
	Status        string
	ReviewTrigger string
	HumanReviewed bool
	ReviewedBy    string
	ReviewedAt    string
	ReviewedHash  string
	InformedBy    []string
	RawData       []byte
	File          string
}

// EvaluateGate mechanically evaluates Phase 0 exit rules G1–G12 across the workspace.
func EvaluateGate(ws WorkspaceFS, info WorkspaceInfo, root string) *GateResult {
	return EvaluateGateWithTarget(ws, info, root, "")
}

// EvaluateGateWithTarget evaluates Phase 0 exit rules for a specific decision target or the full workspace.
func EvaluateGateWithTarget(ws WorkspaceFS, info WorkspaceInfo, root string, targetDecision string) *GateResult {
	targetDecision = strings.TrimSpace(targetDecision)

	hasPipeline := ws != nil && FileExistsFS(ws, PipelineFile)
	if hasPipeline {
		if pipeBytes, err := ws.ReadFile(PipelineFile); err == nil && len(bytes.TrimSpace(pipeBytes)) == 0 {
			hasPipeline = false
		}
	}

	// 1. Scan comparisons, plans, and ADRs in research/
	var compFiles []string
	var planFiles []string
	if ws != nil {
		if entries, err := ws.ListDir(ResearchDir); err == nil {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
					continue
				}
				name := e.Name()
				lower := strings.ToLower(name)
				if strings.HasSuffix(lower, "-plan.md") {
					if info.Scope == ScopeComparison || strings.HasPrefix(lower, "c-") || strings.HasPrefix(lower, "cmp-") || strings.Contains(lower, "-cmp-") || strings.Contains(lower, "comparison") {
						compFiles = append(compFiles, name)
					} else {
						planFiles = append(planFiles, name)
					}
				} else if strings.HasPrefix(strings.ToUpper(name), "C-") || strings.EqualFold(name, "COMPARISON.md") || strings.HasSuffix(lower, "-comparison.md") || strings.Contains(lower, "-cmp-") {
					compFiles = append(compFiles, name)
				}
			}
		}
	}

	// Scan all ADR records in research/ and DECISIONS.md
	allADRs := scanAllADRs(ws)

	// If scope is comparison (or no pipeline, no decision target, no ADRs, but comparisons exist):
	if targetDecision == "" && !hasPipeline && len(allADRs) == 0 && len(planFiles) == 0 {
		if len(compFiles) > 0 || info.Scope == ScopeComparison {
			return &GateResult{
				Verdict:          "not_applicable",
				Target:           "comparison",
				Rules:            []GateRuleResult{},
				Status:           "PASS",
				Passed:           true,
				NextStep:         "Comparison scope complete. WEP matrix generated. Proceed to implementation.",
				StructuralChecks: TrackResult{Label: "Structural Completeness", Passed: true, Score: "1/1", Count: 1, Total: 1},
				TrackAChecks:     TrackResult{Label: "Track A: Two-Way Doors", Passed: true, Score: "1/1", Count: 1, Total: 1},
				TrackBChecks:     TrackResult{Label: "Quality Indicators (Mechanical)", Passed: true, Score: "1/1", Count: 1, Total: 1},
				QualityChecks:    TrackResult{Label: "Quality Indicators (Mechanical)", Passed: true, Score: "1/1", Count: 1, Total: 1},
			}
		}
	}

	target := "project"
	if targetDecision != "" {
		target = targetDecision
	} else if hasPipeline || info.Scope == ScopeProject || (ws != nil && FileExistsFS(ws, FADFile)) {
		target = "project"
	} else if len(planFiles) > 0 {
		target = strings.TrimSuffix(planFiles[0], "-plan.md")
	} else if len(allADRs) > 0 {
		target = allADRs[0].ID
	}

	var rules []GateRuleResult
	var advisories []string
	var warnings []string

	// Target decision filtering
	var targetADRs []adrRecord
	if targetDecision != "" {
		for _, adr := range allADRs {
			if strings.EqualFold(adr.ID, targetDecision) {
				targetADRs = append(targetADRs, adr)
			}
		}
	} else {
		targetADRs = allADRs
	}

	var gateDecisions []GateDecisionItem
	for _, adr := range targetADRs {
		track := "Track A"
		if strings.EqualFold(adr.DoorType, "one-way") {
			track = "Track B"
		}
		gateDecisions = append(gateDecisions, GateDecisionItem{
			ID:       adr.ID,
			Title:    adr.Title,
			DoorType: adr.DoorType,
			Status:   adr.Status,
			Track:    track,
		})
	}

	// -------------------------------------------------------------------------
	// G1: Plan Validity (parses, DAG has no cycles, no orphaned deps, 1 synthesis session for project)
	// -------------------------------------------------------------------------
	var dag *DAG
	if target == "project" || hasPipeline {
		if !hasPipeline {
			rules = append(rules, GateRuleResult{
				ID:     "G1",
				Status: "FAIL",
				Reason: "RESEARCH-PIPELINE.md is missing or empty",
			})
		} else {
			pipeData, err := ws.ReadFile(PipelineFile)
			if err != nil || len(pipeData) == 0 {
				rules = append(rules, GateRuleResult{
					ID:     "G1",
					Status: "FAIL",
					Reason: "RESEARCH-PIPELINE.md could not be read",
				})
			} else {
				parsedDAG, pErr := ParsePipeline(pipeData)
				if pErr != nil || parsedDAG == nil {
					rules = append(rules, GateRuleResult{
						ID:     "G1",
						Status: "FAIL",
						Reason: fmt.Sprintf("Pipeline DAG failed to parse: %v", pErr),
					})
				} else {
					dag = parsedDAG
					if err := dag.ValidateDAG(); err != nil {
						rules = append(rules, GateRuleResult{
							ID:     "G1",
							Status: "FAIL",
							Reason: fmt.Sprintf("Pipeline DAG invalid: %v", err),
						})
					} else {
						known := make(map[string]bool)
						for _, s := range dag.Sessions {
							known[s.ID] = true
						}
						var orphaned []string
						for _, s := range dag.Sessions {
							for _, dep := range s.Dependencies {
								if !known[dep] {
									orphaned = append(orphaned, fmt.Sprintf("%s -> %s", s.ID, dep))
								}
							}
						}
						synthCount := 0
						for _, s := range dag.Sessions {
							if IsSynthesisSession(s.ID) {
								synthCount++
							}
						}
						if len(orphaned) > 0 {
							rules = append(rules, GateRuleResult{
								ID:     "G1",
								Status: "FAIL",
								Reason: fmt.Sprintf("Pipeline DAG contains orphaned dependencies: %s", strings.Join(orphaned, ", ")),
							})
						} else if synthCount != 1 {
							rules = append(rules, GateRuleResult{
								ID:     "G1",
								Status: "FAIL",
								Reason: fmt.Sprintf("Project pipeline requires exactly 1 synthesis session (found %d)", synthCount),
							})
						} else {
							rules = append(rules, GateRuleResult{
								ID:     "G1",
								Status: "PASS",
								Reason: "Project pipeline DAG is structurally valid with 1 synthesis session.",
							})
						}
					}
				}
			}
		}
	} else {
		// Target is decision plan
		rules = append(rules, GateRuleResult{
			ID:     "G1",
			Status: "PASS",
			Reason: "Decision plan structure valid.",
		})
	}

	// -------------------------------------------------------------------------
	// G2: Session Completeness & Quality (all non-synthesis sessions completed & valid)
	// -------------------------------------------------------------------------
	var targetSessions []string
	if target == "project" && dag != nil {
		for _, s := range dag.Sessions {
			if !IsSynthesisSession(s.ID) {
				targetSessions = append(targetSessions, s.ID)
			}
		}
	} else if targetDecision != "" {
		targetSessions = findDecisionTargetSessions(ws, targetDecision)
	} else if len(planFiles) > 0 {
		targetSessions = findPlanSessions(ws, planFiles[0])
	}

	if len(targetSessions) == 0 && (target == "project" || hasPipeline) {
		rules = append(rules, GateRuleResult{
			ID:     "G2",
			Status: "FAIL",
			Reason: "Planned session(s) missing or incomplete: no completed sessions found",
		})
	} else if len(targetSessions) == 0 {
		rules = append(rules, GateRuleResult{
			ID:     "G2",
			Status: "PASS",
			Reason: "No sessions required for direct decision target.",
		})
	} else {
		var failingSessions []string
		for _, sid := range targetSessions {
			sData, _, err := ReadSessionFileFS(ws, SessionsDir, sid)
			if err != nil || len(strings.TrimSpace(sData)) == 0 {
				failingSessions = append(failingSessions, sid)
				continue
			}
			fm, sBody, pErr := ParseFrontmatter([]byte(sData))
			if pErr != nil || fm == nil {
				failingSessions = append(failingSessions, sid)
				continue
			}
			status := strings.ToLower(strings.TrimSpace(fm.GetString("status")))
			if status == "skipped" {
				reason := strings.TrimSpace(fm.GetString("skip_reason"))
				isFalsification := strings.HasSuffix(strings.ToUpper(sid), "-S2") || strings.HasSuffix(strings.ToUpper(sid), "-F")
				if isFalsification {
					failingSessions = append(failingSessions, sid)
					continue
				}
				if reason == "" {
					failingSessions = append(failingSessions, sid)
					continue
				}
				continue
			}
			if status != "complete" {
				failingSessions = append(failingSessions, sid)
				continue
			}
			// Validate substantive findings
			bodyStr := strings.TrimSpace(string(sBody))
			if bodyStr == "TODO" || len(bodyStr) < 40 {
				failingSessions = append(failingSessions, sid)
				continue
			}
			rec := ExtractSection(bodyStr, "recommendation", "verdict")
			kf := ExtractSection(bodyStr, "key findings", "findings")
			ledger := ExtractSection(bodyStr, "sources & evidence ledger", "sources", "evidence ledger", "ledger")
			if rec == "" || kf == "" || ledger == "" {
				failingSessions = append(failingSessions, sid)
				continue
			}
			if !EvidenceGradePattern.MatchString(bodyStr) {
				failingSessions = append(failingSessions, sid)
				continue
			}
			if len(ScanPlaceholders([]byte(sData))) > 0 {
				failingSessions = append(failingSessions, sid)
				continue
			}
		}

		if len(failingSessions) > 0 {
			rules = append(rules, GateRuleResult{
				ID:     "G2",
				Status: "FAIL",
				Reason: fmt.Sprintf("Planned session(s) missing or incomplete: %s", strings.Join(failingSessions, ", ")),
			})
		} else {
			rules = append(rules, GateRuleResult{
				ID:     "G2",
				Status: "PASS",
				Reason: fmt.Sprintf("All %d planned session(s) completed with validated empirical findings.", len(targetSessions)),
			})
		}
	}

	// -------------------------------------------------------------------------
	// G3: FAD Contract (FAD exists, complete, and every accepted ADR appears in Locked Registry)
	// -------------------------------------------------------------------------
	if target == "project" || hasPipeline {
		fadData, err := ws.ReadFile(FADFile)
		if err != nil || len(fadData) == 0 {
			rules = append(rules, GateRuleResult{
				ID:     "G3",
				Status: "FAIL",
				Reason: "Founding Architecture Document (FAD.md) not found",
			})
		} else {
			fadStr := string(fadData)
			execSummary := ExtractSection(fadStr, "executive summary")
			registry := ExtractSection(fadStr, "locked decision registry", "decision registry", "registry")
			risks := ExtractSection(fadStr, "risk register", "risk", "reversal triggers")
			trace := ExtractSection(fadStr, "traceability matrix", "traceability")
			if execSummary == "" || registry == "" || risks == "" || trace == "" {
				rules = append(rules, GateRuleResult{
					ID:     "G3",
					Status: "FAIL",
					Reason: "FAD.md missing required sections (Executive Summary, Locked Decision Registry, Risk Register, Traceability Matrix)",
				})
			} else if phs := ScanPlaceholders(fadData); len(phs) > 0 {
				rules = append(rules, GateRuleResult{
					ID:     "G3",
					Status: "FAIL",
					Reason: fmt.Sprintf("FAD.md contains unexpanded placeholder %q", phs[0]),
				})
			} else {
				var missingADRs []string
				for _, adr := range allADRs {
					if strings.EqualFold(adr.Status, "accepted") {
						if !strings.Contains(registry, adr.ID) {
							missingADRs = append(missingADRs, adr.ID)
						}
					}
				}
				if len(missingADRs) > 0 {
					rules = append(rules, GateRuleResult{
						ID:     "G3",
						Status: "FAIL",
						Reason: fmt.Sprintf("FAD Locked Decision Registry omits accepted ADR(s): %s", strings.Join(missingADRs, ", ")),
					})
				} else {
					rules = append(rules, GateRuleResult{
						ID:     "G3",
						Status: "PASS",
						Reason: "Founding Architecture Document (FAD.md) is complete and covers all accepted ADRs.",
					})
				}
			}
		}
	} else {
		rules = append(rules, GateRuleResult{
			ID:     "G3",
			Status: "n/a",
			Reason: "FAD contract applies to project scope only.",
		})
	}

	// -------------------------------------------------------------------------
	// G4: Referenced Decisions Exist
	// -------------------------------------------------------------------------
	var referencedDecIDs []string
	if dag != nil {
		for _, s := range dag.Sessions {
			if s.DecisionRef != "" {
				referencedDecIDs = append(referencedDecIDs, s.DecisionRef)
			}
		}
	}
	knownADRMap := make(map[string]bool)
	for _, adr := range allADRs {
		knownADRMap[strings.ToUpper(adr.ID)] = true
	}
	var missingRefs []string
	for _, ref := range referencedDecIDs {
		if !knownADRMap[strings.ToUpper(ref)] {
			missingRefs = append(missingRefs, ref)
		}
	}
	if len(missingRefs) > 0 {
		rules = append(rules, GateRuleResult{
			ID:     "G4",
			Status: "FAIL",
			Reason: fmt.Sprintf("Referenced decision(s) lack ADR: %s", strings.Join(missingRefs, ", ")),
		})
	} else {
		rules = append(rules, GateRuleResult{
			ID:     "G4",
			Status: "PASS",
			Reason: "All referenced decisions have corresponding ADRs.",
		})
	}

	// -------------------------------------------------------------------------
	// G5: No Proposed ADRs (all target ADRs accepted, superseded, or rejected)
	// -------------------------------------------------------------------------
	if len(targetADRs) == 0 {
		warnings = append(warnings, "No architectural decisions recorded yet — record decisions using vivechak_record_decision before running the gate")
		rules = append(rules, GateRuleResult{
			ID:     "G5",
			Status: "FAIL",
			Reason: "No ADRs found in workspace",
		})
	} else {
		var proposed []string
		for _, adr := range targetADRs {
			st := strings.ToLower(strings.TrimSpace(adr.Status))
			if st == "proposed" || st == "" {
				proposed = append(proposed, adr.ID)
			}
		}
		if len(proposed) > 0 {
			for _, pid := range proposed {
				advisories = append(advisories, fmt.Sprintf("ADR %s is in 'proposed' status. Update status to 'accepted' before proceeding through the gate.", pid))
			}
			rules = append(rules, GateRuleResult{
				ID:     "G5",
				Status: "FAIL",
				Reason: fmt.Sprintf("ADR(s) still in proposed status: %s", strings.Join(proposed, ", ")),
			})
		} else {
			rules = append(rules, GateRuleResult{
				ID:     "G5",
				Status: "PASS",
				Reason: "All target ADRs are accepted or resolved.",
			})
		}
	}

	// Filter accepted one-way and two-way ADRs
	var acceptedOneWay []adrRecord
	var acceptedTwoWay []adrRecord
	for _, adr := range targetADRs {
		if strings.EqualFold(adr.Status, "accepted") {
			if strings.EqualFold(adr.DoorType, "one-way") {
				acceptedOneWay = append(acceptedOneWay, adr)
			} else {
				acceptedTwoWay = append(acceptedTwoWay, adr)
			}
		}
	}

	// -------------------------------------------------------------------------
	// G6: One-Way Door Evidence Depth (>=2 Grade A/B claims with URLs)
	// -------------------------------------------------------------------------
	if len(acceptedOneWay) == 0 {
		rules = append(rules, GateRuleResult{
			ID:     "G6",
			Status: "n/a",
			Reason: "No One-Way Door decisions in target.",
		})
	} else {
		var g6Failing []string
		for _, adr := range acceptedOneWay {
			count := countGradeABClaimsWithURLs(ws, adr)
			if count < 2 {
				if len(adr.InformedBy) == 0 {
					g6Failing = append(g6Failing, fmt.Sprintf("%s (no sessions linked to %s via informed_by_sessions or informs_decisions)", adr.ID, adr.ID))
				} else {
					g6Failing = append(g6Failing, fmt.Sprintf("%s (found %d Grade A/B claims with URLs, want >=2)", adr.ID, count))
				}
			}
		}
		if len(g6Failing) > 0 {
			rules = append(rules, GateRuleResult{
				ID:     "G6",
				Status: "FAIL",
				Reason: fmt.Sprintf("One-Way Door evidence depth insufficient: %s", strings.Join(g6Failing, "; ")),
			})
		} else {
			rules = append(rules, GateRuleResult{
				ID:     "G6",
				Status: "PASS",
				Reason: "One-Way Door decisions have >=2 Grade A/B evidence citations with URLs.",
			})
		}
	}

	// -------------------------------------------------------------------------
	// G7: One-Way Door Verification Provenance (fetched/human labels as reported)
	// -------------------------------------------------------------------------
	if len(acceptedOneWay) == 0 {
		rules = append(rules, GateRuleResult{
			ID:     "G7",
			Status: "n/a",
			Reason: "No One-Way Door decisions in target.",
		})
	} else {
		var g7Failing []string
		for _, adr := range acceptedOneWay {
			verifiedCount := countVerifiedClaimsForADR(ws, adr)
			if verifiedCount < 2 {
				if len(adr.InformedBy) == 0 {
					g7Failing = append(g7Failing, fmt.Sprintf("%s (no sessions linked to %s via informed_by_sessions or informs_decisions)", adr.ID, adr.ID))
				} else {
					g7Failing = append(g7Failing, fmt.Sprintf("%s has %d verified (fetched/human) Grade A/B claims (want >=2)", adr.ID, verifiedCount))
				}
			}
		}
		if len(g7Failing) > 0 {
			rules = append(rules, GateRuleResult{
				ID:     "G7",
				Status: "FAIL",
				Reason: fmt.Sprintf("One-Way Door evidence provenance unverified: %s (labels as reported; sources not fetched or verified)", strings.Join(g7Failing, "; ")),
			})
		} else {
			rules = append(rules, GateRuleResult{
				ID:     "G7",
				Status: "PASS",
				Reason: "One-Way Door evidence claims carry verified provenance (>=2 fetched/human labels as reported).",
			})
		}
	}

	// -------------------------------------------------------------------------
	// G8: One-Way Door No Placeholders
	// -------------------------------------------------------------------------
	if len(acceptedOneWay) == 0 {
		rules = append(rules, GateRuleResult{
			ID:     "G8",
			Status: "n/a",
			Reason: "No One-Way Door decisions in target.",
		})
	} else {
		var g8Failing []string
		for _, adr := range acceptedOneWay {
			bodyStr := string(adr.RawData)
			hasPh := false
			if regexp.MustCompile(`(?m)^\s*\[[^\]]{8,}\]\s*$`).MatchString(bodyStr) ||
				strings.Contains(bodyStr, "[Auto-drafted") ||
				strings.Contains(bodyStr, "[Condition or date") ||
				strings.Contains(bodyStr, "Production deployment milestone or scheduled quarterly review") ||
				strings.Contains(bodyStr, "{slot}") ||
				strings.Contains(bodyStr, "{option}") {
				hasPh = true
			}
			rejSec := ExtractSection(bodyStr, "rejected alternative", "alternative", "tradeoff")
			if strings.TrimSpace(rejSec) == "" || len(strings.TrimSpace(rejSec)) < 20 || strings.Contains(rejSec, "[Review") {
				hasPh = true
			}
			if hasPh {
				g8Failing = append(g8Failing, adr.ID)
			}
		}
		if len(g8Failing) > 0 {
			rules = append(rules, GateRuleResult{
				ID:     "G8",
				Status: "FAIL",
				Reason: fmt.Sprintf("One-Way Door ADR(s) contain unexpanded placeholders or default template text: %s", strings.Join(g8Failing, ", ")),
			})
		} else {
			rules = append(rules, GateRuleResult{
				ID:     "G8",
				Status: "PASS",
				Reason: "One-Way Door ADRs contain no unexpanded placeholders or default template text.",
			})
		}
	}

	// -------------------------------------------------------------------------
	// G9: One-Way Door Measurable Review Trigger
	// -------------------------------------------------------------------------
	if len(acceptedOneWay) == 0 {
		rules = append(rules, GateRuleResult{
			ID:     "G9",
			Status: "n/a",
			Reason: "No One-Way Door decisions in target.",
		})
	} else {
		var g9Failing []string
		measRe := regexp.MustCompile(`(?i)(?:[><=]\s*\d+|\d+\s*(?:%|ms|s|m|h|d|k|M|G|req|events|qps|rps|users|ops|writes|reads)|\d{4}-\d{2}-\d{2})`)
		for _, adr := range acceptedOneWay {
			trigger := strings.TrimSpace(adr.ReviewTrigger)
			if trigger == "" || strings.HasPrefix(trigger, "[") || !measRe.MatchString(trigger) {
				g9Failing = append(g9Failing, fmt.Sprintf("%s (%q)", adr.ID, trigger))
			}
		}
		if len(g9Failing) > 0 {
			rules = append(rules, GateRuleResult{
				ID:     "G9",
				Status: "FAIL",
				Reason: fmt.Sprintf("One-Way Door ADR(s) lack measurable reversal trigger: %s", strings.Join(g9Failing, ", ")),
			})
		} else {
			rules = append(rules, GateRuleResult{
				ID:     "G9",
				Status: "PASS",
				Reason: "One-Way Door ADRs have specific, measurable reversal triggers.",
			})
		}
	}

	// -------------------------------------------------------------------------
	// G10: One-Way Door Adversarial Challenge & Red-Team
	// -------------------------------------------------------------------------
	if len(acceptedOneWay) == 0 {
		rules = append(rules, GateRuleResult{
			ID:     "G10",
			Status: "n/a",
			Reason: "No One-Way Door decisions in target.",
		})
	} else {
		var g10Failing []string
		for _, adr := range acceptedOneWay {
			chID, chHash := findLatestChallengeRecord(ws, adr.ID)
			if chID == "" {
				g10Failing = append(g10Failing, fmt.Sprintf("%s lacks an adversarial red-team challenge record (%s-R1.md); run vivechak_review(mode=\"red_team\", target=\"%s\")", adr.ID, adr.ID, adr.ID))
				advisories = append(advisories, fmt.Sprintf("W-OWD-NO-CHALLENGE: One-Way Door %s lacks an adversarial challenge record", adr.ID))
				continue
			}
			currOutcomeHash := ComputeDecisionOutcomeHash(adr.RawData)
			if chHash == "" || chHash != currOutcomeHash {
				g10Failing = append(g10Failing, fmt.Sprintf("%s has stale challenge record (%s); re-run red-team challenge", adr.ID, chID))
				continue
			}
			bodyStr := string(adr.RawData)
			chSec := ExtractSection(bodyStr, "challenge findings", "red-team", "stress test")
			chSecTrimmed := strings.TrimSpace(chSec)
			if !strings.Contains(chSecTrimmed, chID) || len(chSecTrimmed) < 50 || strings.Contains(chSecTrimmed, "[Document") {
				g10Failing = append(g10Failing, fmt.Sprintf("%s Challenge Findings section must cite %s with >=50 chars of substantive findings", adr.ID, chID))
				continue
			}
		}
		if len(g10Failing) > 0 {
			rules = append(rules, GateRuleResult{
				ID:     "G10",
				Status: "FAIL",
				Reason: strings.Join(g10Failing, "; "),
			})
		} else {
			rules = append(rules, GateRuleResult{
				ID:     "G10",
				Status: "PASS",
				Reason: "One-Way Door decisions have current, cited adversarial challenge records.",
			})
		}
	}

	// -------------------------------------------------------------------------
	// G11: One-Way Door Human Sign-off
	// -------------------------------------------------------------------------
	if len(acceptedOneWay) == 0 {
		rules = append(rules, GateRuleResult{
			ID:     "G11",
			Status: "n/a",
			Reason: "No One-Way Door decisions in target.",
		})
	} else {
		var g11Failing []string
		var unreviewedOWDs []string
		for _, adr := range acceptedOneWay {
			if !adr.HumanReviewed || adr.ReviewedBy == "" || adr.ReviewedHash == "" {
				unreviewedOWDs = append(unreviewedOWDs, adr.ID)
				g11Failing = append(g11Failing, fmt.Sprintf("%s requires human architectural sign-off (run 'vck signoff %s --reviewer \"<name>\"')", adr.ID, adr.ID))
				continue
			}
			expectedHash := ComputeADRHash(adr.RawData)
			if adr.ReviewedHash != expectedHash {
				g11Failing = append(g11Failing, fmt.Sprintf("%s has stale sign-off (content changed after review); re-run 'vck signoff %s'", adr.ID, adr.ID))
				continue
			}
		}
		if len(unreviewedOWDs) > 0 {
			advisories = append(advisories, fmt.Sprintf("G11 Violation: %d One-Way Door decision(s) lack human review: %s. %s requires human architectural sign-off (run 'vck signoff %s --reviewer \"<name>\"')",
				len(unreviewedOWDs), strings.Join(unreviewedOWDs, ", "), unreviewedOWDs[0], unreviewedOWDs[0]))
		}
		if len(g11Failing) > 0 {
			rules = append(rules, GateRuleResult{
				ID:     "G11",
				Status: "FAIL",
				Reason: strings.Join(g11Failing, "; "),
			})
		} else {
			rules = append(rules, GateRuleResult{
				ID:     "G11",
				Status: "PASS",
				Reason: "All One-Way Door decisions have valid human sign-off.",
			})
		}
	}

	// -------------------------------------------------------------------------
	// G12: Two-Way Door Fast Track
	// -------------------------------------------------------------------------
	if len(acceptedTwoWay) == 0 {
		rules = append(rules, GateRuleResult{
			ID:     "G12",
			Status: "n/a",
			Reason: "No Two-Way Door decisions in target.",
		})
	} else {
		var g12Failing []string
		for _, adr := range acceptedTwoWay {
			bodyStr := string(adr.RawData)
			outcomeSec := ExtractSection(bodyStr, "decision outcome", "decision", "chosen option", "verdict")
			if strings.TrimSpace(outcomeSec) == "" || strings.Contains(outcomeSec, "[Extract") {
				g12Failing = append(g12Failing, fmt.Sprintf("%s missing Decision Outcome", adr.ID))
				continue
			}
			trig := strings.TrimSpace(adr.ReviewTrigger)
			if trig == "" || strings.HasPrefix(trig, "[") {
				g12Failing = append(g12Failing, fmt.Sprintf("%s missing review_trigger", adr.ID))
				continue
			}
		}
		if len(g12Failing) > 0 {
			rules = append(rules, GateRuleResult{
				ID:     "G12",
				Status: "FAIL",
				Reason: fmt.Sprintf("Two-Way Door requirements incomplete: %s", strings.Join(g12Failing, "; ")),
			})
		} else {
			rules = append(rules, GateRuleResult{
				ID:     "G12",
				Status: "PASS",
				Reason: "Two-Way Door decisions meet fast-track exit criteria.",
			})
		}
	}

	// Compute overall verdict
	verdict := "PASS"
	for _, r := range rules {
		if r.Status == "FAIL" {
			verdict = "FAIL"
			break
		}
	}

	var nextStep string
	if verdict == "PASS" {
		nextStep = "Phase 0 Gate: PASS. You may start implementation. Evidence grades and verification labels are as reported by the agent; Vivechak did not fetch or verify sources."
	} else {
		nextStep = "Phase 0 Gate: FAIL. Resolve the failing rules above before starting implementation."
	}

	structChecks := computeStructuralChecks(ws, info, rules)
	trackBChecks := computeTrackBChecks(ws, info, targetADRs, rules)

	hasTwoWayAccepted := len(acceptedTwoWay) > 0
	trackAPassed := structChecks.Passed || hasTwoWayAccepted
	if len(targetADRs) == 0 && !structChecks.Passed {
		trackAPassed = false
	}
	var trackAIssues []string
	for _, r := range rules {
		if r.ID == "G12" && r.Status == "FAIL" {
			trackAPassed = false
			trackAIssues = append(trackAIssues, r.Reason)
		}
	}
	if !trackAPassed && len(trackAIssues) == 0 {
		trackAIssues = append(trackAIssues, structChecks.Issues...)
	}

	scoreStr := "1/1"
	count := 1
	if !trackAPassed {
		scoreStr = "0/1"
		count = 0
	}

	trackAChecks := TrackResult{
		Label:  "Track A: Two-Way Doors",
		Passed: trackAPassed,
		Score:  scoreStr,
		Count:  count,
		Total:  1,
		Issues: trackAIssues,
	}

	return &GateResult{
		Verdict:          verdict,
		Target:           target,
		Rules:            rules,
		Advisories:       advisories,
		Warnings:         warnings,
		NextStep:         nextStep,
		Decisions:        gateDecisions,
		Status:           verdict,
		Passed:           (verdict == "PASS"),
		StructuralChecks: structChecks,
		TrackAChecks:     trackAChecks,
		TrackBChecks:     trackBChecks,
		QualityChecks:    trackBChecks,
	}
}

var (
	lowGradeRe         = regexp.MustCompile(`(?i)(?:\[(?:Grade\s+|E-)?([CDE])(?:\s*[·:|]|\s*\])[^\]\n]*\]|\((?:Grade\s+|E-)?([CDE])(?:\s*[·:|]|\s*\))[^)\n]*\)|\bGrade\s+[CDE]\b|\bGrade:[CDE]\b)`)
	recalledCitationRe = regexp.MustCompile(`(?i)(?:\[[^\]\n]*\b(?:recalled|parametric)\b[^\]\n]*\]|\([^)\n]*\b(?:recalled|parametric)\b[^)\n]*\)|\brecalled\s+memory\b|\|\s*recalled\s*\|)`)
)

func computeStructuralChecks(ws WorkspaceFS, info WorkspaceInfo, rules []GateRuleResult) TrackResult {
	var issues []string
	passed := 0
	total := 0

	effectiveScope := info.Scope
	if ws != nil && (FileExistsFS(ws, PipelineFile) || FileExistsFS(ws, FADFile)) {
		effectiveScope = ScopeProject
	}

	switch effectiveScope {
	case ScopeDecision:
		total = 2
		if info.SessionCount > 0 {
			passed++
		} else {
			issues = append(issues, "No completed sessions found")
		}
		hasDec := info.HasDecisions
		if ws != nil && !hasDec {
			if _, err := ws.Stat(DecisionsFile); err == nil {
				hasDec = true
			}
		}
		if hasDec {
			passed++
		} else {
			issues = append(issues, "No decision record found")
		}

	case ScopeComparison:
		total = 1
		passed = 1

	default: // ScopeProject
		total = 4
		if ws != nil && FileExistsFS(ws, PipelineFile) {
			passed++
		} else {
			issues = append(issues, "RESEARCH-PIPELINE.md not found")
		}
		hasIncompleteSession := false
		if ws != nil && FileExistsFS(ws, PipelineFile) {
			if pipeBytes, err := ws.ReadFile(PipelineFile); err == nil && len(pipeBytes) > 0 {
				if parsedDAG, pErr := ParsePipeline(pipeBytes); pErr == nil && parsedDAG != nil {
					for _, s := range parsedDAG.Sessions {
						if IsSynthesisSession(s.ID) {
							continue
						}
						sData, _, sErr := ReadSessionFileFS(ws, SessionsDir, s.ID)
						if sErr != nil || len(strings.TrimSpace(sData)) == 0 {
							hasIncompleteSession = true
							issues = append(issues, fmt.Sprintf("Session %s missing", s.ID))
							break
						}
						fm, _, pErr := ParseFrontmatter([]byte(sData))
						if pErr != nil || fm == nil || !strings.EqualFold(fm.GetString("status"), "complete") {
							hasIncompleteSession = true
							issues = append(issues, fmt.Sprintf("Session %s incomplete (not status: complete)", s.ID))
							break
						}
					}
				}
			}
		}
		if info.SessionCount > 0 && !hasIncompleteSession {
			passed++
		} else if !hasIncompleteSession {
			issues = append(issues, "No completed sessions found")
		}
		if info.HasDecisions || (ws != nil && FileExistsFS(ws, DecisionsFile)) {
			passed++
		} else {
			issues = append(issues, "No decisions recorded")
		}
		if ws != nil && FileExistsFS(ws, FADFile) {
			passed++
		} else {
			issues = append(issues, "FAD.md not found")
		}
	}

	return TrackResult{
		Label:  "Structural Completeness",
		Passed: len(issues) == 0 && passed == total,
		Score:  fmt.Sprintf("%d/%d", passed, total),
		Count:  passed,
		Total:  total,
		Issues: issues,
	}
}

func computeTrackBChecks(ws WorkspaceFS, info WorkspaceInfo, targetADRs []adrRecord, rules []GateRuleResult) TrackResult {
	effectiveScope := info.Scope
	if ws != nil && (FileExistsFS(ws, PipelineFile) || FileExistsFS(ws, FADFile)) {
		effectiveScope = ScopeProject
	}

	if effectiveScope == ScopeComparison && len(targetADRs) == 0 {
		return TrackResult{
			Label:  "Quality Indicators (Mechanical)",
			Passed: true,
			Score:  "1/1",
			Count:  1,
			Total:  1,
		}
	}

	total := 3
	passedCount := 0
	var issues []string

	b3Failed := false
	b4Failed := false
	b5Failed := false

	if ws != nil {
		for _, adr := range targetADRs {
			if !strings.EqualFold(adr.DoorType, "one-way") {
				continue
			}
			bodyStr := string(adr.RawData)
			altContent := ExtractSection(bodyStr, "alternative", "rejected")
			if len(strings.TrimSpace(altContent)) < 30 {
				b5Failed = true
				issues = append(issues, fmt.Sprintf("B5-ALTERNATIVES: One-way door %s lacks documented rejected alternatives section", adr.ID))
			}

			lowerBody := strings.ToLower(bodyStr)
			hasContext := strings.Contains(lowerBody, "context") || strings.Contains(lowerBody, "problem")
			hasRationale := strings.Contains(lowerBody, "rationale") || strings.Contains(lowerBody, "decision") || strings.Contains(lowerBody, "verdict")
			hasTriggers := strings.Contains(lowerBody, "trigger") || strings.Contains(lowerBody, "reversal") || (adr.ReviewTrigger != "")
			if len(bodyStr) < 200 || !hasContext || !hasRationale || !hasTriggers {
				issues = append(issues, fmt.Sprintf("B-DECISION-STUB: One-way door %s lacks mandatory structural sections", adr.ID))
			}

			if !adr.HumanReviewed || adr.ReviewedBy == "" || adr.ReviewedHash == "" {
				issues = append(issues, fmt.Sprintf("G11 Violation: One-Way Door %s lacks human review: requires human architectural sign-off (run 'vck signoff %s --reviewer \"<name>\"')", adr.ID, adr.ID))
			}

			for _, sid := range adr.InformedBy {
				if sData, _, err := ReadSessionFileFS(ws, SessionsDir, sid); err == nil {
					recText := extractRecommendationText(sData)
					if len(lowGradeRe.FindAllString(recText, -1)) > 0 {
						b3Failed = true
						issues = append(issues, fmt.Sprintf("B3-EVIDENTIARY: Session %s informs one-way door %s with Grade C/D/E citation(s)", sid, adr.ID))
					}
					if len(recalledCitationRe.FindAllString(recText, -1)) > 0 {
						b4Failed = true
						issues = append(issues, fmt.Sprintf("B4-VERIFICATION: Session %s informs one-way door %s with 'recalled' citation(s)", sid, adr.ID))
					}
				}
			}
		}

		if effectiveScope == ScopeProject {
			if info.SessionCount < 3 {
				issues = append(issues, fmt.Sprintf("Only %d sessions — minimum 3 recommended for project scope", info.SessionCount))
			}
			if FileExistsFS(ws, FADFile) {
				if fadData, err := ws.ReadFile(FADFile); err == nil {
					fadStr := string(fadData)
					if !EvidenceGradePattern.MatchString(fadStr) {
						issues = append(issues, "FAD has no evidence grades")
					}
				}
			}
		}
	}

	if !b3Failed {
		passedCount++
	}
	if !b4Failed {
		passedCount++
	}
	if !b5Failed {
		passedCount++
	}

	trackBPassed := (passedCount == total)

	return TrackResult{
		Label:  "Quality Indicators (Mechanical)",
		Passed: trackBPassed && len(issues) == 0,
		Score:  fmt.Sprintf("%d/%d", passedCount, total),
		Count:  passedCount,
		Total:  total,
		Issues: issues,
	}
}

func extractRecommendationText(content string) string {
	rec := ExtractSection(content, "recommend", "verdict", "decision")
	if len(strings.TrimSpace(rec)) >= 20 {
		return rec
	}
	findings := ExtractSection(content, "key finding", "finding")
	if len(strings.TrimSpace(findings)) >= 20 {
		return findings
	}
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

func verifyDecisionsMechanical(ws WorkspaceFS, info WorkspaceInfo) (bool, []string, []string) {
	hasDecFile := false
	if ws != nil {
		if _, err := ws.Stat(DecisionsFile); err == nil {
			hasDecFile = true
		}
	}
	if !info.HasDecisions && !hasDecFile {
		return false, []string{"Cannot check decisions — DECISIONS.md not found"}, nil
	}

	adrs := scanAllADRs(ws)
	if len(adrs) == 0 {
		return false, []string{"DECISIONS.md or ADR files appear empty or trivial"}, nil
	}

	var issues []string
	var advisories []string
	allAccepted := true
	missingTriggers := 0

	for _, adr := range adrs {
		st := strings.ToLower(adr.Status)
		if st != "accepted" && st != "superseded" && st != "deprecated" && st != "rejected" {
			allAccepted = false
			issues = append(issues, fmt.Sprintf("ADR %s has status %q (must be accepted/superseded/deprecated/rejected)", adr.ID, adr.Status))
		}
		if strings.EqualFold(adr.DoorType, "one-way") && st == "accepted" {
			trig := strings.TrimSpace(adr.ReviewTrigger)
			if trig == "" || strings.HasPrefix(trig, "[") {
				missingTriggers++
				issues = append(issues, fmt.Sprintf("One-way door %s lacks required reversal trigger", adr.ID))
			}
			if !adr.HumanReviewed {
				advisories = append(advisories, fmt.Sprintf("One-way door %s lacks human review", adr.ID))
			}
		}
	}

	return allAccepted && missingTriggers == 0, issues, advisories
}

func verifyEvidentiaryIntegrity(ws WorkspaceFS, info WorkspaceInfo) ([]string, []string) {
	var blocking []string
	var advisories []string
	adrs := scanAllADRs(ws)

	for _, d := range adrs {
		if !strings.EqualFold(d.DoorType, "one-way") {
			continue
		}
		bodyStr := string(d.RawData)
		altContent := ExtractSection(bodyStr, "alternative", "rejected")
		if len(strings.TrimSpace(altContent)) < 30 {
			blocking = append(blocking, fmt.Sprintf("B5-ALTERNATIVES: One-way door %s lacks documented rejected alternatives section", d.ID))
		}

		for _, sid := range d.InformedBy {
			if sData, _, err := ReadSessionFileFS(ws, SessionsDir, sid); err == nil {
				recText := extractRecommendationText(sData)
				lowGrades := len(lowGradeRe.FindAllString(recText, -1))
				if lowGrades > 0 {
					blocking = append(blocking, fmt.Sprintf("B3-EVIDENTIARY: Session %s informs one-way door %s but recommendation has %d Grade C/D/E citation(s)", sid, d.ID, lowGrades))
				}
				recalledCount := len(recalledCitationRe.FindAllString(recText, -1))
				if recalledCount > 0 {
					blocking = append(blocking, fmt.Sprintf("B4-VERIFICATION: Session %s informs one-way door %s but recommendation has %d 'recalled' citation(s)", sid, d.ID, recalledCount))
				}
			}
		}
	}

	return blocking, advisories
}

// scanAllADRs discovers all ADR records across research/ and DECISIONS.md.
func scanAllADRs(ws WorkspaceFS) []adrRecord {
	if ws == nil {
		return nil
	}
	var records []adrRecord
	seen := make(map[string]bool)

	// 1. Scan standalone ADR markdown files in research/ and research/decisions/
	searchDirs := []string{ResearchDir, DecisionsDir}
	for _, sDir := range searchDirs {
		if entries, err := ws.ListDir(sDir); err == nil {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
					continue
				}
				name := e.Name()
				if IsSpecialResearchFile(name) {
					continue
				}
				if strings.HasPrefix(strings.ToUpper(name), "C-") {
					continue
				}
				data, err := ws.ReadFile(filepath.Join(sDir, name))
				if err != nil || len(data) == 0 {
					continue
				}
				fm, _, pErr := ParseFrontmatter(data)
				if pErr != nil || fm == nil {
					continue
				}
				isCandidate := strings.HasPrefix(strings.ToUpper(name), "D-")
				if !isCandidate && fm.GetString("type") != "decision" && !fm.Has("door_type") {
					continue
				}
				if fm.GetString("type") == "plan" || fm.GetString("type") == "challenge" || strings.HasSuffix(strings.ToLower(name), "-plan.md") || strings.Contains(strings.ToUpper(name), "-R") {
					continue
				}
				id := fm.GetString("id")
				if id == "" {
					id = fm.GetString("decision_id")
				}
				if id == "" {
					id = strings.TrimSuffix(name, ".md")
				}
				upperID := strings.ToUpper(id)
				if seen[upperID] {
					continue
				}
				seen[upperID] = true

				doorType := fm.GetString("door_type")
				if doorType == "" {
					doorType = fm.GetString("door")
				}
				if doorType == "" {
					doorType = "two-way"
				}

				humanRev := false
				if strings.EqualFold(fm.GetString("human_reviewed"), "true") {
					humanRev = true
				}

				records = append(records, adrRecord{
					ID:            id,
					Title:         fm.GetString("title"),
					DoorType:      doorType,
					Status:        fm.GetString("status"),
					ReviewTrigger: fm.GetString("review_trigger"),
					HumanReviewed: humanRev,
					ReviewedBy:    fm.GetString("reviewed_by"),
					ReviewedAt:    fm.GetString("reviewed_at"),
					ReviewedHash:  fm.GetString("reviewed_hash"),
					InformedBy:    append(fm.GetStringSlice("informed_by_sessions"), fm.GetStringSlice("informing_sessions")...),
					RawData:       data,
					File:          filepath.Join(sDir, name),
				})
			}
		}
	}

	// 2. Scan DECISIONS.md for any additional decisions
	if decData, err := ws.ReadFile(DecisionsFile); err == nil && len(decData) > 0 {
		anchoredRe := regexp.MustCompile(`(?s)<!-- DECISION:\s*([A-Za-z0-9_-]+)\s*-->\s*(.*?)\s*<!-- /DECISION:\s*[A-Za-z0-9_-]+\s*-->`)
		for _, m := range anchoredRe.FindAllSubmatch(decData, -1) {
			id := string(m[1])
			upperID := strings.ToUpper(id)
			if seen[upperID] {
				continue
			}
			body := m[2]
			fm, _, pErr := ParseFrontmatter(body)
			if pErr != nil || fm == nil {
				continue
			}
			seen[upperID] = true
			doorType := fm.GetString("door_type")
			if doorType == "" {
				doorType = fm.GetString("door")
			}
			if doorType == "" {
				doorType = "two-way"
			}
			humanRev := strings.EqualFold(fm.GetString("human_reviewed"), "true")
			records = append(records, adrRecord{
				ID:            id,
				Title:         fm.GetString("title"),
				DoorType:      doorType,
				Status:        fm.GetString("status"),
				ReviewTrigger: fm.GetString("review_trigger"),
				HumanReviewed: humanRev,
				ReviewedBy:    fm.GetString("reviewed_by"),
				ReviewedAt:    fm.GetString("reviewed_at"),
				ReviewedHash:  fm.GetString("reviewed_hash"),
				InformedBy:    append(fm.GetStringSlice("informed_by_sessions"), fm.GetStringSlice("informing_sessions")...),
				RawData:       body,
				File:          DecisionsFile,
			})
		}
	}

	resolveBidirectionalInformingSessions(ws, records)
	return records
}

func findDecisionTargetSessions(ws WorkspaceFS, did string) []string {
	if ws == nil {
		return nil
	}
	var sessions []string
	didUpper := strings.ToUpper(did)
	for _, planName := range []string{fmt.Sprintf("%s-plan.md", did), fmt.Sprintf("%s.plan.md", did)} {
		if pData, err := ws.ReadFile(filepath.Join(ResearchDir, planName)); err == nil && len(pData) > 0 {
			if block, bErr := ParsePlanMachineBlock(pData); bErr == nil && block != nil && len(block.Sessions) > 0 {
				for _, s := range block.Sessions {
					sessions = append(sessions, s.ID)
				}
				return sessions
			}
			reSess := regexp.MustCompile(`(?m)^\s*(?:\d+\.|\-|\*)\s+([A-Za-z0-9_-]+)\s*:`)
			for _, m := range reSess.FindAllSubmatch(pData, -1) {
				if len(m) > 1 {
					sessions = append(sessions, string(m[1]))
				}
			}
			if len(sessions) > 0 {
				return sessions
			}
		}
	}
	// Fallback to checking sessions dir for prefix
	if entries, err := ws.ListDir(SessionsDir); err == nil {
		for _, e := range entries {
			stem := strings.TrimSuffix(e.Name(), ".md")
			if strings.HasPrefix(strings.ToUpper(stem), didUpper+"-") {
				sessions = append(sessions, stem)
			}
		}
	}
	return sessions
}

func findPlanSessions(ws WorkspaceFS, planFile string) []string {
	if ws == nil {
		return nil
	}
	var sessions []string
	if pData, err := ws.ReadFile(filepath.Join(ResearchDir, planFile)); err == nil {
		if block, bErr := ParsePlanMachineBlock(pData); bErr == nil && block != nil && len(block.Sessions) > 0 {
			for _, s := range block.Sessions {
				sessions = append(sessions, s.ID)
			}
			return sessions
		}
		reSess := regexp.MustCompile(`(?m)^\s*(?:\d+\.|\-|\*)\s+([A-Za-z0-9_-]+)\s*:`)
		for _, m := range reSess.FindAllSubmatch(pData, -1) {
			if len(m) > 1 {
				sessions = append(sessions, string(m[1]))
			}
		}
	}
	return sessions
}

func countGradeABClaimsWithURLs(ws WorkspaceFS, adr adrRecord) int {
	count := 0
	checkText := func(text string) {
		lines := strings.Split(text, "\n")
		for _, l := range lines {
			if regexp.MustCompile(`(?i)\[(?:Grade\s+[AB]|E-[AB]|[AB])\b`).MatchString(l) && strings.Contains(l, "http") {
				count++
			}
		}
	}

	checkText(string(adr.RawData))
	if ws != nil {
		for _, sid := range adr.InformedBy {
			if sData, _, err := ReadSessionFileFS(ws, SessionsDir, sid); err == nil {
				checkText(sData)
			}
		}
	}
	return count
}

func countVerifiedClaimsForADR(ws WorkspaceFS, adr adrRecord) int {
	count := 0
	gradeRe := regexp.MustCompile(`(?i)(?:\[(?:Grade\s+[AB]|E-[AB]|[AB])\b|\bGrade\s+[AB]\b)`)
	checkText := func(text string) {
		lines := strings.Split(text, "\n")
		for _, l := range lines {
			if !gradeRe.MatchString(l) {
				continue
			}
			lower := strings.ToLower(l)
			// Must NOT be unverified / recalled / parametric memory
			if strings.Contains(lower, "recalled") || strings.Contains(lower, "parametric") || strings.Contains(lower, "model memory") || strings.Contains(lower, "ai memory") {
				continue
			}
			// Must have verified provenance label: fetched, verified, or human (human-provided, etc.)
			if strings.Contains(lower, "fetched") || strings.Contains(lower, "verified") || strings.Contains(lower, "human") {
				count++
			}
		}
	}

	checkText(string(adr.RawData))
	if ws != nil {
		for _, sid := range adr.InformedBy {
			if sData, _, err := ReadSessionFileFS(ws, SessionsDir, sid); err == nil {
				checkText(sData)
			}
		}
	}
	return count
}

func resolveBidirectionalInformingSessions(ws WorkspaceFS, records []adrRecord) {
	if ws == nil {
		return
	}
	sessionInforms := make(map[string][]string)
	var allSessionIDs []string
	if entries, err := ws.ListDir(SessionsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			sID := strings.TrimSuffix(e.Name(), ".md")
			allSessionIDs = append(allSessionIDs, sID)
			relPath := filepath.Join(SessionsDir, e.Name())
			if sData, err := ws.ReadFile(relPath); err == nil {
				if fm, _, err := ParseFrontmatter(sData); err == nil && fm != nil {
					if id := fm.GetString("session_id"); id != "" {
						sID = id
					} else if id := fm.GetString("id"); id != "" {
						sID = id
					}
					var informsSlice []string
					informsSlice = append(informsSlice, fm.GetStringSlice("informs_decisions")...)
					informsSlice = append(informsSlice, fm.GetStringSlice("informs")...)
					if singleInforms := fm.GetString("informs"); singleInforms != "" {
						informsSlice = append(informsSlice, singleInforms)
					}
					if singleInformsDec := fm.GetString("informs_decisions"); singleInformsDec != "" {
						informsSlice = append(informsSlice, singleInformsDec)
					}
					for _, did := range informsSlice {
						upperDID := strings.ToUpper(strings.TrimSpace(did))
						if upperDID != "" {
							sessionInforms[upperDID] = append(sessionInforms[upperDID], sID)
						}
					}
				}
			}
		}
	}

	for i := range records {
		adr := &records[i]
		seen := make(map[string]bool)
		var merged []string
		add := func(s string) {
			s = strings.TrimSpace(s)
			if s == "" {
				return
			}
			upper := strings.ToUpper(s)
			if !seen[upper] {
				seen[upper] = true
				merged = append(merged, s)
			}
		}

		for _, s := range adr.InformedBy {
			add(s)
		}
		upperID := strings.ToUpper(adr.ID)
		for _, s := range sessionInforms[upperID] {
			add(s)
		}
		for _, s := range findDecisionTargetSessions(ws, adr.ID) {
			add(s)
		}
		prefix := upperID + "-S"
		for _, s := range allSessionIDs {
			if strings.HasPrefix(strings.ToUpper(s), prefix) {
				add(s)
			}
		}
		adr.InformedBy = merged
	}
}

func findLatestChallengeRecord(ws WorkspaceFS, did string) (string, string) {
	if ws == nil {
		return "", ""
	}
	didUpper := strings.ToUpper(did)
	re := regexp.MustCompile(fmt.Sprintf(`^(?i)%s-R(\d+)\.md$`, regexp.QuoteMeta(didUpper)))

	dirsToScan := []string{ResearchDir, filepath.Join(ResearchDir, ".challenges")}
	latestNum := 0
	latestPath := ""
	latestFile := ""
	for _, dir := range dirsToScan {
		entries, err := ws.ListDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if m := re.FindStringSubmatch(e.Name()); len(m) > 1 {
				if num, err := strconv.Atoi(m[1]); err == nil && num > latestNum {
					latestNum = num
					latestFile = e.Name()
					latestPath = filepath.Join(dir, e.Name())
				}
			}
		}
	}
	if latestFile == "" {
		return "", ""
	}
	recordID := strings.TrimSuffix(latestFile, ".md")
	data, err := ws.ReadFile(latestPath)
	if err != nil {
		return recordID, ""
	}
	fm, _, _ := ParseFrontmatter(data)
	if fm == nil {
		return recordID, ""
	}
	hash := fm.GetString("decision_hash")
	if hash == "" {
		hash = fm.GetString("target_outcome_hash")
	}
	return recordID, hash
}
