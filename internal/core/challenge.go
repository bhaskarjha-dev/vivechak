package core

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ChallengeMode represents the type of adversarial challenge to generate.
type ChallengeMode string

const (
	// ModeRedTeam generates disconfirming searches, assumption audits, alternative steel-manning, and premortems.
	ModeRedTeam ChallengeMode = "red_team"
	// ModeEvidenceAudit generates per-citation URL and primary source verification prompts.
	ModeEvidenceAudit ChallengeMode = "evidence_audit"
	// ModeCrossSession generates cross-session compatibility and evidence independence audits.
	ModeCrossSession ChallengeMode = "cross_session"
)

// ChallengePrompt represents an actionable adversarial prompt for the agent to execute.
type ChallengePrompt struct {
	Type      string `json:"type"`       // disconfirming_search, assumption_audit, steel_man_alternative, premortem, etc.
	Prompt    string `json:"prompt"`     // The adversarial prompt text
	Rationale string `json:"rationale"` // Why this challenge matters (cites P8, P3, etc.)
}

// ChallengeResult encapsulates the challenge prompts and parsed session context.
type ChallengeResult struct {
	SessionID string            `json:"session_id"`
	Mode      ChallengeMode     `json:"mode"`
	Prompts   []ChallengePrompt `json:"challenge_prompts"`
	Context   map[string]any    `json:"context"`
}

var urlPattern = regexp.MustCompile(`https?://[^\s\)\]]+`)

// GenerateChallenges parses session content and produces structured adversarial challenge prompts.
// It executes purely deterministic parsing and templating without calling an external LLM.
func GenerateChallenges(sessionContent []byte, sessionID string, mode ChallengeMode, allSessions map[string][]byte) (*ChallengeResult, error) {
	if len(sessionContent) == 0 {
		return nil, fmt.Errorf("session content is empty")
	}

	if mode == "" {
		mode = ModeRedTeam
	}

	fm, body, _ := ParseFrontmatter(sessionContent)
	bodyStr := string(body)

	// Extract context
	title := ""
	doorType := ""
	if fm != nil {
		title = fm.GetString("title")
		doorType = fm.GetString("door_type")
	}

	rec := extractRecommendationSnippet(bodyStr)
	if rec == "" && title != "" {
		rec = title
	} else if rec == "" {
		rec = sessionID
	}

	rejectedAlts := extractRejectedAlternatives(bodyStr)
	gradeCounts := countEvidenceGrades(bodyStr)

	res := &ChallengeResult{
		SessionID: sessionID,
		Mode:      mode,
		Prompts:   make([]ChallengePrompt, 0),
		Context: map[string]any{
			"recommendation":    rec,
			"door_type":         doorType,
			"rejected_alts":     rejectedAlts,
			"evidence_grades":   gradeCounts,
			"total_sessions":    len(allSessions),
		},
	}

	switch mode {
	case ModeRedTeam:
		res.Prompts = generateRedTeamPrompts(rec, rejectedAlts)
	case ModeEvidenceAudit:
		res.Prompts = generateEvidenceAuditPrompts(bodyStr)
	case ModeCrossSession:
		res.Prompts = generateCrossSessionPrompts(sessionID, rec, allSessions)
	default:
		return nil, fmt.Errorf("unknown challenge mode: %q (supported: red_team, evidence_audit, cross_session)", mode)
	}

	return res, nil
}

func extractRecommendationSnippet(body string) string {
	lines := strings.Split(body, "\n")
	inRec := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "recommend") {
				inRec = true
				continue
			} else if inRec {
				break
			}
		}
		if inRec && trimmed != "" && !strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(trimmed, "```") {
			// Extract first sentence or up to 120 chars
			if idx := strings.Index(trimmed, "."); idx > 0 && idx < 150 {
				return strings.TrimSpace(trimmed[:idx])
			}
			if len(trimmed) > 120 {
				return strings.TrimSpace(trimmed[:120]) + "..."
			}
			return trimmed
		}
	}
	return ""
}

func extractRejectedAlternatives(body string) []string {
	var rejected []string
	lines := strings.Split(body, "\n")
	inAlt := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "alternative") {
				inAlt = true
				continue
			} else if inAlt {
				break
			}
		}
		if inAlt && strings.HasPrefix(trimmed, "|") {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "reject") {
				cells := strings.Split(trimmed, "|")
				if len(cells) >= 3 {
					alt := strings.TrimSpace(cells[1])
					alt = strings.Trim(alt, "*_`[]")
					if alt != "" && !strings.EqualFold(alt, "option") && !strings.EqualFold(alt, "candidate") {
						rejected = append(rejected, alt)
					}
				}
			}
		}
	}
	return rejected
}

func countEvidenceGrades(body string) map[string]int {
	counts := map[string]int{"A": 0, "B": 0, "C": 0, "D": 0, "E": 0}
	matches := EvidenceGradePattern.FindAllString(body, -1)
	for _, m := range matches {
		upper := strings.ToUpper(m)
		for _, g := range []string{"A", "B", "C", "D", "E"} {
			if strings.Contains(upper, "GRADE "+g) || strings.Contains(upper, "["+g) || strings.Contains(upper, "("+g) {
				counts[g]++
				break
			}
		}
	}
	return counts
}

func generateRedTeamPrompts(rec string, rejectedAlts []string) []ChallengePrompt {
	var prompts []ChallengePrompt

	// 1. Disconfirming search
	cleanRec := cleanSnippetForPrompt(rec)
	prompts = append(prompts, ChallengePrompt{
		Type: "disconfirming_search",
		Prompt: fmt.Sprintf(
			"Execute a disconfirming search to challenge the recommendation (%q):\n"+
				"Search for: %q problems, %q failures, %q limitations, %q criticism, %q postmortem.\n"+
				"Target: Uncover at least 2 real-world production failure modes, scale ceilings, or operational regressions. "+
				"Document the findings with evidence grades.",
			cleanRec, cleanRec, cleanRec, cleanRec, cleanRec, cleanRec),
		Rationale: "Principle P8 (Structured Falsification): Disconfirming evidence must be prioritized. Meta-research indicates a 74% confirmation bias rate without adversarial search.",
	})

	// 2. Assumption audit
	prompts = append(prompts, ChallengePrompt{
		Type: "assumption_audit",
		Prompt: fmt.Sprintf(
			"Conduct a rigorous assumption audit for %q:\n"+
				"1. Identify the 3 most critical implicit assumptions underlying this choice (e.g. workload concurrency, resource constraints, deployment topology, dependency stability).\n"+
				"2. For each assumption, define what empirical observation or threshold would DISPROVE it.\n"+
				"3. Search for evidence that might invalidate these assumptions.",
			cleanRec),
		Rationale: "Principle P8 & P2: All architectural commitments rest on implicit hypotheses. Reversibility requires explicit falsification criteria.",
	})

	// 3. Steel-man rejected alternative
	altCase := "the leading alternative"
	if len(rejectedAlts) > 0 {
		altCase = fmt.Sprintf("the rejected alternative %q", rejectedAlts[0])
	}
	prompts = append(prompts, ChallengePrompt{
		Type: "steel_man_alternative",
		Prompt: fmt.Sprintf(
			"Steel-man %s:\n"+
				"Construct the strongest, most compelling architectural case FOR adopting it instead of %q. "+
				"Under what specific workload scale, team constraint, or deployment environment would it become the clearly superior choice? "+
				"Verify that the rejection rationale in the session was causal rather than superficial.",
			altCase, cleanRec),
		Rationale: "Principle P8: Document rejected alternatives with causal refutation. Steel-manning prevents superficial dismissal.",
	})

	// 4. Prospective Premortem
	prompts = append(prompts, ChallengePrompt{
		Type: "premortem",
		Prompt: fmt.Sprintf(
			"Execute a prospective hindsight premortem for %q:\n"+
				"\"Imagine it is 12 months from now, and this technical choice has catastrophically failed in production.\"\n"+
				"1. Write 3 concrete, architecture-specific failure scenarios detailing how and why it failed.\n"+
				"2. For each failure scenario, formulate an architectural mitigation or pre-condition to prevent it.\n"+
				"3. Document the top failure scenario and mitigation into the session's Open Questions & Risks section.",
			cleanRec),
		Rationale: "Principle P8 & Phase 0 Gate Check B7: Prospective hindsight mitigates optimism bias and uncovers catastrophic blind spots before implementation.",
	})

	return prompts
}

func generateEvidenceAuditPrompts(body string) []ChallengePrompt {
	var prompts []ChallengePrompt
	lines := strings.Split(body, "\n")

	type auditedClaim struct {
		grade string
		line  string
		url   string
	}
	var claims []auditedClaim

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if EvidenceGradePattern.MatchString(trimmed) && (strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "|")) {
			g := "A"
			upper := strings.ToUpper(trimmed)
			for _, check := range []string{"GRADE A", "GRADE B", "GRADE C", "GRADE D", "GRADE E", "(A", "(B", "(C", "(D", "(E"} {
				if strings.Contains(upper, check) {
					g = string(check[len(check)-1])
					break
				}
			}
			url := urlPattern.FindString(trimmed)
			claims = append(claims, auditedClaim{grade: g, line: trimmed, url: url})
		}
	}

	for _, c := range claims {
		if c.grade == "A" {
			if c.url != "" {
				prompts = append(prompts, ChallengePrompt{
					Type: "verify_primary_url",
					Prompt: fmt.Sprintf(
						"Verify primary source URL validity:\n"+
							"URL: %s\n"+
							"Claim: %s\n"+
							"Confirm that this URL still directly supports the claim and does not contain contradictory caveats or deprecation notices.",
						c.url, cleanSnippetForPrompt(c.line)),
					Rationale: "Principle P3 (Evidentiary Grounding): Grade A requires verified primary documentation directly supporting the claim.",
				})
			} else {
				prompts = append(prompts, ChallengePrompt{
					Type: "missing_grade_a_url",
					Prompt: fmt.Sprintf(
						"Provide primary source URL for Grade A claim:\n"+
							"Claim: %s\n"+
							"This claim is marked Grade A but lacks an authoritative primary URL. Find and add the official documentation, RFC, or specification link.",
						cleanSnippetForPrompt(c.line)),
					Rationale: "Principle P3 & Gate Check B3: Grade A claims require fetched primary source verification with URLs.",
				})
			}
		} else if c.grade == "C" || c.grade == "D" || strings.Contains(strings.ToLower(c.line), "recalled") {
			prompts = append(prompts, ChallengePrompt{
				Type: "upgrade_secondary_citation",
				Prompt: fmt.Sprintf(
					"Upgrade or corroborate secondary/recalled claim:\n"+
						"Claim: %s\n"+
						"Search official specifications or reproducible benchmarks to upgrade this claim from Grade %s/recalled to Grade A/B.",
					cleanSnippetForPrompt(c.line), c.grade),
				Rationale: "Principle P3: Recalled knowledge is capped at Grade D. One-way decisions require Grade A/B corroboration.",
			})
		}

		if len(prompts) >= 4 {
			break
		}
	}

	if len(prompts) == 0 {
		prompts = append(prompts, ChallengePrompt{
			Type: "systematic_evidence_audit",
			Prompt: "Audit all factual statements in this session. Ensure every substantive technical claim carries an inline evidence grade [Grade A–E], modifiers (corroborated, fresh), and verification method (fetched with URL, cached, recalled).",
			Rationale: "Principle P3: Every technical claim must be explicitly graded.",
		})
	}

	return prompts
}

func generateCrossSessionPrompts(currentSessionID, currentRec string, allSessions map[string][]byte) []ChallengePrompt {
	var prompts []ChallengePrompt

	if len(allSessions) < 3 {
		prompts = append(prompts, ChallengePrompt{
			Type: "cross_session_prerequisite",
			Prompt: fmt.Sprintf(
				"Cross-session analysis currently has %d session(s) available. Inter-session coherence audits are most effective with 3+ completed sessions. "+
					"Review available sessions for direct interface coupling or conflicting runtime assumptions with %s.",
				len(allSessions), currentSessionID),
			Rationale: "Principle P1 & P7: Cross-session triangulation audits inter-session consistency across multiple completed research nodes.",
		})
		return prompts
	}

	// Collect session IDs
	var sessionIDs []string
	for sid := range allSessions {
		sessionIDs = append(sessionIDs, sid)
	}
	sort.Strings(sessionIDs)
	sessionList := strings.Join(sessionIDs, ", ")

	prompts = append(prompts, ChallengePrompt{
		Type: "technology_coherence",
		Prompt: fmt.Sprintf(
			"Cross-Session Technology Coherence Audit across (%s):\n"+
				"1. Do the technology choices across these sessions work harmoniously together at runtime?\n"+
				"2. Are there conflicting operational prerequisites (e.g. CGo dependencies, conflicting memory models, mismatched async event loops)?\n"+
				"3. Does %s make assumptions about upstream systems that contradict other sessions?",
			sessionList, currentSessionID),
		Rationale: "Principle P1 (Context Architecture Law): Synthesize technology choices across decomposed sessions to catch composability failures early.",
	})

	prompts = append(prompts, ChallengePrompt{
		Type: "evidence_independence",
		Prompt: fmt.Sprintf(
			"Evidence Independence Audit across (%s):\n"+
				"1. Do the recommendations across these sessions rely on truly independent evidence bases?\n"+
				"2. Are there circular citations or shared vendor whitepapers influencing multiple separate decisions?\n"+
				"3. Confirm that one-way door decisions do not rely on single-source authority.",
			sessionList),
		Rationale: "Principle P7 (Staged Triangulation): Corroboration requires independent evidence sources, not multiple citations of the same root marketing source.",
	})

	prompts = append(prompts, ChallengePrompt{
		Type: "coupling_boundary_audit",
		Prompt: fmt.Sprintf(
			"Coupling & Dependency Audit:\n"+
				"Check whether session %s creates an unintended tight coupling with any other session in (%s). "+
				"Ensure each decision boundary adheres to Principle P5 (Commodity-Maximized Composition) with clean interfaces.",
			currentSessionID, sessionList),
		Rationale: "Principle P1 & P5: Maintain strict boundary encapsulation between researched subsystems.",
	})

	return prompts
}

func cleanSnippetForPrompt(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > 80 {
		return s[:80] + "..."
	}
	return s
}
