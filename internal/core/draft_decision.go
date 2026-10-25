package core

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var canonicalEvidenceIDRe = regexp.MustCompile(`\b(E-\d{1,4})\b`)

func extractCanonicalEvidenceIDs(sessionFM Frontmatter, sessionBody string) []string {
	seen := make(map[string]bool)
	var ids []string

	if sessionFM != nil {
		for _, id := range sessionFM.GetStringSlice("evidence_refs") {
			id = strings.TrimSpace(id)
			if id != "" && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}

	for _, m := range canonicalEvidenceIDRe.FindAllStringSubmatch(sessionBody, -1) {
		if len(m) > 1 {
			id := m[1]
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}

	return ids
}

// DraftDecisionFromSession extracts key sections from a session and assembles
// a draft decision record. This eliminates the ~60% content redundancy between
// session findings and decision records.
//
// Returns the draft as markdown content with YAML frontmatter, ready for
// review and editing before saving via vivechak_record_decision.
func DraftDecisionFromSession(sessionContent []byte, decisionID string) (string, error) {
	return DraftDecisionFromSessionWithExisting(nil, sessionContent, decisionID)
}

// DraftDecisionFromSessionWithExisting extracts key sections from a session and non-destructively
// merges them with an existing decision record (if any), preserving user-edited titles, review triggers,
// and substantive content sections while filling placeholders.
func DraftDecisionFromSessionWithExisting(existingADR []byte, sessionContent []byte, decisionID string) (string, error) {
	if len(sessionContent) == 0 {
		return "", fmt.Errorf("session content is empty")
	}

	fm, body, _ := ParseFrontmatter(sessionContent)
	bodyStr := string(body)

	// Extract session metadata
	sessionID := ""
	sessionTitle := ""
	if fm != nil {
		sessionID = fm.GetString("session_id")
		if sessionID == "" {
			sessionID = fm.GetString("id")
		}
		sessionTitle = fm.GetString("title")
	}

	// Extract sections from session body
	recommendation := ExtractSection(bodyStr, "recommend", "decision outcome", "chosen", "verdict", "executive summary", "summary")
	alternatives := ExtractSection(bodyStr, "evaluated option", "alternatives", "candidate", "shortlist", "option")
	concerns := ExtractSection(bodyStr, "concern", "failure", "risk", "premortem", "reversal")
	evidenceRefs := extractEvidenceGrades(bodyStr)

	// If existing ADR is provided, extract its existing frontmatter and body for non-destructive merge
	var existFM Frontmatter
	var existBodyStr string
	if len(existingADR) > 0 {
		efm, ebody, _ := ParseFrontmatter(existingADR)
		existFM = efm
		existBodyStr = string(ebody)
	}

	isPlaceholder := func(val string) bool {
		val = strings.TrimSpace(val)
		return val == "" || (strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]")) ||
			strings.EqualFold(val, "null") || strings.EqualFold(val, "tbd") || strings.EqualFold(val, "todo")
	}

	// Determine fields: preserve existing non-placeholder values
	title := sessionTitle
	if existFM != nil && !isPlaceholder(existFM.GetString("title")) {
		title = existFM.GetString("title")
	} else if title == "" {
		title = "[Decision Title]"
	}

	status := "proposed"
	if existFM != nil && !isPlaceholder(existFM.GetString("status")) {
		status = existFM.GetString("status")
	}

	doorType := "one-way"
	contentLower := strings.ToLower(bodyStr)
	if strings.Contains(contentLower, "door: two-way") ||
		strings.Contains(contentLower, "door_type: two-way") ||
		strings.Contains(contentLower, "door type: two-way") ||
		strings.Contains(contentLower, "two-way door") {
		doorType = "two-way"
	}
	if existFM != nil && !isPlaceholder(existFM.GetString("door_type")) {
		doorType = existFM.GetString("door_type")
	}

	date := time.Now().UTC().Format("2006-01-02")
	if existFM != nil && !isPlaceholder(existFM.GetString("date")) {
		date = existFM.GetString("date")
	}

	confidence := "medium"
	if existFM != nil && !isPlaceholder(existFM.GetString("confidence")) {
		confidence = existFM.GetString("confidence")
	}

	reviewTrigger := "Production deployment milestone or scheduled quarterly review"
	if existFM != nil && !isPlaceholder(existFM.GetString("review_trigger")) {
		reviewTrigger = existFM.GetString("review_trigger")
	}

	// Auto-drafted decisions always require interactive human signoff
	humanReviewed := "false"

	// Canonical evidence refs from session
	canonicalIDs := extractCanonicalEvidenceIDs(fm, bodyStr)
	if existFM != nil {
		for _, id := range existFM.GetStringSlice("evidence_refs") {
			id = strings.TrimSpace(id)
			has := false
			for _, cid := range canonicalIDs {
				if cid == id {
					has = true
					break
				}
			}
			if id != "" && !has {
				canonicalIDs = append(canonicalIDs, id)
			}
		}
	}

	// Informed by sessions: combine existing + current session
	var informedSessions []string
	seenSessions := make(map[string]bool)
	if existFM != nil {
		for _, s := range existFM.GetStringSlice("informed_by_sessions") {
			s = strings.TrimSpace(s)
			if s != "" && !seenSessions[s] {
				seenSessions[s] = true
				informedSessions = append(informedSessions, s)
			}
		}
	}
	if sessionID != "" && !seenSessions[sessionID] {
		informedSessions = append(informedSessions, sessionID)
	}

	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", decisionID)
	escapedTitle := strings.ReplaceAll(strings.ReplaceAll(title, `\`, `\\`), `"`, `\"`)
	fmt.Fprintf(&b, "title: \"%s\"\n", escapedTitle)
	fmt.Fprintf(&b, "status: %s\n", status)
	fmt.Fprintf(&b, "door_type: %s\n", doorType)
	fmt.Fprintf(&b, "date: %s\n", date)
	fmt.Fprintf(&b, "confidence: %s\n", confidence)
	if len(canonicalIDs) > 0 {
		var quoted []string
		for _, id := range canonicalIDs {
			quoted = append(quoted, fmt.Sprintf("%q", id))
		}
		fmt.Fprintf(&b, "evidence_refs: [%s]\n", strings.Join(quoted, ", "))
	} else {
		b.WriteString("evidence_refs: []\n")
	}
	if len(informedSessions) > 0 {
		var quoted []string
		for _, s := range informedSessions {
			quoted = append(quoted, fmt.Sprintf("%q", s))
		}
		fmt.Fprintf(&b, "informed_by_sessions: [%s]\n", strings.Join(quoted, ", "))
	} else {
		b.WriteString("informed_by_sessions: []\n")
	}
	b.WriteString("supersedes: null\n")
	b.WriteString("superseded_by: null\n")
	b.WriteString("amends: null\n")
	if !isPlaceholder(reviewTrigger) {
		escapedTrigger := strings.ReplaceAll(strings.ReplaceAll(reviewTrigger, `\`, `\\`), `"`, `\"`)
		fmt.Fprintf(&b, "review_trigger: \"%s\"\n", escapedTrigger)
	} else {
		b.WriteString("review_trigger: \"Production deployment milestone or scheduled quarterly review\"\n")
	}
	b.WriteString("review_date: null\n")
	b.WriteString("prediction: null\n")
	b.WriteString("tags: []\n")
	b.WriteString("authored_by: \"auto-drafted from session\"\n")
	fmt.Fprintf(&b, "human_reviewed: %s\n", humanReviewed)
	b.WriteString("schema_version: \"0.1.0\"\n")
	b.WriteString("---\n\n")

	// Body
	fmt.Fprintf(&b, "# %s: %s\n\n", decisionID, title)

	isSubstantiveSection := func(secContent string) bool {
		t := strings.TrimSpace(secContent)
		if len(t) < 20 {
			return false
		}
		if strings.HasPrefix(t, "[Auto-drafted") || strings.HasPrefix(t, "[Extract") ||
			strings.HasPrefix(t, "[Review") || strings.HasPrefix(t, "[Document") {
			return false
		}
		return true
	}

	// Context
	b.WriteString("## Context & Problem Statement\n\n")
	existContext := ExtractSection(existBodyStr, "context", "problem")
	if isSubstantiveSection(existContext) {
		b.WriteString(existContext + "\n\n")
	} else {
		b.WriteString("[Auto-drafted — review and refine the context from session findings.]\n\n")
	}

	// Evaluated Options
	b.WriteString("## Evaluated Options\n\n")
	existOptions := ExtractSection(existBodyStr, "evaluated option", "option", "candidate")
	if isSubstantiveSection(existOptions) {
		b.WriteString(existOptions + "\n\n")
	} else if alternatives != "" {
		b.WriteString(alternatives + "\n\n")
	} else {
		b.WriteString("[Extract evaluated options from session findings.]\n\n")
	}

	// Decision Outcome
	b.WriteString("## Decision Outcome\n\n")
	existOutcome := ExtractSection(existBodyStr, "decision outcome", "chosen option", "chosen", "verdict")
	if isSubstantiveSection(existOutcome) {
		b.WriteString(existOutcome + "\n\n")
	} else if recommendation != "" {
		b.WriteString(recommendation + "\n\n")
	} else {
		b.WriteString("**Chosen Option:** [Extract from session recommendation.]\n\n")
	}

	// Rejected Alternatives
	b.WriteString("## Rejected Alternatives & Tradeoffs\n\n")
	existAlt := ExtractSection(existBodyStr, "rejected alternative", "tradeoff")
	if isSubstantiveSection(existAlt) {
		b.WriteString(existAlt + "\n\n")
	} else {
		b.WriteString("[Review session findings and document why alternatives were rejected.]\n\n")
	}

	// Challenge Findings & Red-Team Stress Test
	b.WriteString("## Challenge Findings & Red-Team Stress Test\n\n")
	existChallenge := ExtractSection(existBodyStr, "challenge findings", "red-team", "stress test")
	if isSubstantiveSection(existChallenge) {
		b.WriteString(existChallenge + "\n\n")
	} else {
		b.WriteString("> Required for One-Way Doors: run `vivechak_review(target=\"...\", mode=\"challenge\")` before commitment.\n")
		b.WriteString("[Document disconfirming search results, premortem failure modes, and adversarial stress-test findings.]\n\n")
	}

	// Failure Modes & Reversal Triggers
	b.WriteString("## Failure Modes & Reversal Triggers\n\n")
	existConcerns := ExtractSection(existBodyStr, "failure mode", "reversal trigger")
	if isSubstantiveSection(existConcerns) {
		b.WriteString(existConcerns + "\n\n")
	} else if concerns != "" {
		b.WriteString(concerns + "\n\n")
	} else {
		b.WriteString("[Extract failure modes and reversal triggers from session concerns.]\n\n")
	}

	// Evidence summary
	if len(evidenceRefs) > 0 {
		b.WriteString("## Evidence Summary (from session)\n\n")
		for _, ref := range evidenceRefs {
			b.WriteString("- " + ref + "\n")
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}


func extractEvidenceGrades(body string) []string {
	matches := EvidenceGradePattern.FindAllString(body, 20)
	if len(matches) == 0 {
		return nil
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, m := range matches {
		cleaned := strings.TrimSpace(m)
		if !seen[cleaned] {
			seen[cleaned] = true
			unique = append(unique, cleaned)
		}
	}
	return unique
}
