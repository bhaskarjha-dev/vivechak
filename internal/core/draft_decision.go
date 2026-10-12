package core

import (
	"fmt"
	"strings"
	"time"
)

// DraftDecisionFromSession extracts key sections from a session and assembles
// a draft decision record. This eliminates the ~60% content redundancy between
// session findings and decision records.
//
// Returns the draft as markdown content with YAML frontmatter, ready for
// review and editing before saving via vivechak_record_decision.
func DraftDecisionFromSession(sessionContent []byte, decisionID string) (string, error) {
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
	recommendation := ExtractSection(bodyStr, "recommend", "decision outcome", "chosen", "verdict")
	alternatives := ExtractSection(bodyStr, "evaluated option", "alternatives", "candidate", "shortlist", "option")
	concerns := ExtractSection(bodyStr, "concern", "failure", "risk", "premortem", "reversal")
	evidenceRefs := extractEvidenceGrades(bodyStr)

	// Build the draft
	var b strings.Builder

	// YAML frontmatter
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", decisionID)
	if sessionTitle != "" {
		escapedTitle := strings.ReplaceAll(strings.ReplaceAll(sessionTitle, `\`, `\\`), `"`, `\"`)
		fmt.Fprintf(&b, "title: \"%s\"\n", escapedTitle)
	} else {
		b.WriteString("title: \"[Decision Title]\"\n")
	}
	b.WriteString("status: proposed\n")
	doorType := "one-way"
	contentLower := strings.ToLower(bodyStr)
	if strings.Contains(contentLower, "door: two-way") ||
		strings.Contains(contentLower, "door_type: two-way") ||
		strings.Contains(contentLower, "door type: two-way") ||
		strings.Contains(contentLower, "two-way door") {
		doorType = "two-way"
	}
	fmt.Fprintf(&b, "door_type: %s             # one-way | two-way\n", doorType)
	fmt.Fprintf(&b, "date: %s\n", time.Now().UTC().Format("2006-01-02"))
	b.WriteString("confidence: medium             # high | medium | low\n")
	b.WriteString("evidence_refs: []              # Optional: E-NNN IDs if you maintain a separate evidence ledger\n")
	if sessionID != "" {
		fmt.Fprintf(&b, "informed_by_sessions: [\"%s\"]\n", sessionID)
	} else {
		b.WriteString("informed_by_sessions: []\n")
	}
	b.WriteString("supersedes: null\n")
	b.WriteString("superseded_by: null\n")
	b.WriteString("amends: null\n")
	b.WriteString("review_trigger: \"[Condition or date for mandatory re-evaluation]\"\n")
	b.WriteString("review_date: null\n")
	b.WriteString("prediction: null\n")
	b.WriteString("tags: []\n")
	b.WriteString("authored_by: \"auto-drafted from session\"\n")
	b.WriteString("human_reviewed: false\n")
	b.WriteString("schema_version: \"0.1.0\"\n")
	b.WriteString("---\n\n")

	// Body
	fmt.Fprintf(&b, "# %s: %s\n\n", decisionID, sessionTitle)

	// Context
	b.WriteString("## Context & Problem Statement\n\n")
	b.WriteString("[Auto-drafted — review and refine the context from session findings.]\n\n")

	// Evaluated Options
	b.WriteString("## Evaluated Options\n\n")
	if alternatives != "" {
		b.WriteString(alternatives + "\n\n")
	} else {
		b.WriteString("[Extract evaluated options from session findings.]\n\n")
	}

	// Decision Outcome
	b.WriteString("## Decision Outcome\n\n")
	if recommendation != "" {
		b.WriteString(recommendation + "\n\n")
	} else {
		b.WriteString("**Chosen Option:** [Extract from session recommendation.]\n\n")
	}

	// Rejected Alternatives
	b.WriteString("## Rejected Alternatives & Tradeoffs\n\n")
	b.WriteString("[Review session findings and document why alternatives were rejected.]\n\n")

	// Failure Modes & Reversal Triggers
	b.WriteString("## Failure Modes & Reversal Triggers\n\n")
	if concerns != "" {
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
