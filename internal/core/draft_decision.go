package core

import (
	"fmt"
	"regexp"
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
	recommendation := extractSection(bodyStr, "recommend", "decision outcome", "chosen", "verdict")
	alternatives := extractSection(bodyStr, "evaluated option", "alternatives", "candidate", "shortlist", "option")
	concerns := extractSection(bodyStr, "concern", "failure", "risk", "premortem", "reversal")
	evidenceRefs := extractEvidenceGrades(bodyStr)

	// Build the draft
	var b strings.Builder

	// YAML frontmatter
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("id: %s\n", decisionID))
	if sessionTitle != "" {
		b.WriteString(fmt.Sprintf("title: \"%s\"\n", sessionTitle))
	} else {
		b.WriteString("title: \"[Decision Title]\"\n")
	}
	b.WriteString("status: proposed\n")
	b.WriteString("door_type: one-way             # one-way | two-way\n")
	b.WriteString(fmt.Sprintf("date: %s\n", time.Now().UTC().Format("2006-01-02")))
	b.WriteString("confidence: medium             # high | medium | low\n")
	b.WriteString("evidence_refs: []              # Optional: E-NNN IDs if you maintain a separate evidence ledger\n")
	if sessionID != "" {
		b.WriteString(fmt.Sprintf("informed_by_sessions: [\"%s\"]\n", sessionID))
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
	b.WriteString(fmt.Sprintf("# %s: %s\n\n", decisionID, sessionTitle))

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

// extractSection extracts the first matching section from body content.
// Searches for headings containing any of the keywords.
func extractSection(body string, keywords ...string) string {
	lines := strings.Split(body, "\n")
	var result strings.Builder
	inSection := false
	sectionLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}

			if inSection && level <= sectionLevel {
				// Hit a same-level or higher heading — stop
				break
			}

			for _, kw := range keywords {
				if strings.Contains(lower, kw) {
					inSection = true
					sectionLevel = level
					break
				}
			}
			if inSection {
				continue // skip the heading itself
			}
		}

		if inSection && trimmed != "" {
			result.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(result.String())
}

// extractEvidenceGrades finds all inline evidence grade markers in the body.
var draftEvidencePattern = regexp.MustCompile(`(?:\[?[Gg]rade\s+[A-E][^\]\)\n]*\]?|\b[A-E]\s*\([^)]+\)|\([Gg]rade\s+[A-E][^)]*\)|\([A-E]\s*[·|][^)]*\))`)

func extractEvidenceGrades(body string) []string {
	matches := draftEvidencePattern.FindAllString(body, 20)
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
