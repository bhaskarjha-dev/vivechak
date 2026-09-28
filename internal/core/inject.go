package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ContextSlot defines a placeholder in a session prompt that should be
// filled with findings from upstream sessions.
const (
	// UpstreamFindingsSlot is the placeholder for injected findings.
	UpstreamFindingsSlot = "[UPSTREAM_FINDINGS]"

	// AllFindingsSlot is the placeholder for ALL session findings (used in SYN-01).
	AllFindingsSlot = "[ALL_SESSION_FINDINGS]"
)

// InjectedSession is a session prompt with upstream context injected.
type InjectedSession struct {
	// Session is the session metadata.
	Session Session `json:"session"`

	// InjectedPrompt is the prompt with upstream findings filled in.
	InjectedPrompt string `json:"injected_prompt"`

	// UpstreamSessions lists which sessions' findings were injected.
	UpstreamSessions []string `json:"upstream_sessions,omitempty"`

	// InjectedBytes is the total size of injected context.
	InjectedBytes int `json:"injected_bytes"`
}

// InjectContext takes a session and fills its prompt with findings from
// completed upstream sessions. This is the core value proposition:
// eliminating manual copy-paste between research sessions.
func InjectContext(session Session, sessionsDir string, completedSessions map[string]bool) (*InjectedSession, error) {
	prompt := session.Prompt
	if prompt == "" {
		return nil, fmt.Errorf("session %s has no prompt", session.ID)
	}

	result := &InjectedSession{
		Session: session,
	}

	// Determine if this is a synthesis session
	isSynthesis := strings.HasPrefix(strings.ToUpper(session.ID), "SYN")

	if isSynthesis {
		// For synthesis: inject ALL completed session findings
		findings, sessions, totalBytes := gatherAllFindings(sessionsDir, completedSessions)
		// Try AllFindingsSlot first, then UpstreamFindingsSlot, then fallback append.
		// Only inject once to prevent doubling token consumption.
		if strings.Contains(prompt, AllFindingsSlot) {
			prompt = strings.Replace(prompt, AllFindingsSlot, findings, 1)
		} else if strings.Contains(prompt, UpstreamFindingsSlot) {
			prompt = strings.Replace(prompt, UpstreamFindingsSlot, findings, 1)
		} else if findings != "" {
			prompt = prompt + "\n\n## UPSTREAM RESEARCH CONTEXT\n\n" + findings
		}
		result.UpstreamSessions = sessions
		result.InjectedBytes = totalBytes
	} else if len(session.Dependencies) > 0 {
		// For regular sessions: inject only direct dependency findings
		findings, sessions, totalBytes := gatherDependencyFindings(sessionsDir, session.Dependencies)
		prompt = injectIntoPrompt(prompt, UpstreamFindingsSlot, findings)
		result.UpstreamSessions = sessions
		result.InjectedBytes = totalBytes
	}

	result.InjectedPrompt = prompt
	return result, nil
}

// gatherAllFindings reads all completed session files and assembles their
// findings into a single context block.
func gatherAllFindings(sessionsDir string, completedSessions map[string]bool) (string, []string, int) {
	var findings []string
	var sessionIDs []string
	totalBytes := 0

	// Collect and sort IDs for deterministic ordering
	var sortedIDs []string
	for id := range completedSessions {
		sortedIDs = append(sortedIDs, id)
	}
	sort.Strings(sortedIDs)

	for _, id := range sortedIDs {
		content := readSessionFile(sessionsDir, id)
		if content == "" {
			continue
		}

		sessionIDs = append(sessionIDs, id)

		// Extract key findings (frontmatter body)
		_, body, _ := ParseFrontmatter([]byte(content))
		excerpt := extractFindings(string(body), id)
		findings = append(findings, excerpt)
		totalBytes += len(excerpt)
	}

	return strings.Join(findings, "\n\n---\n\n"), sessionIDs, totalBytes
}

// gatherDependencyFindings reads only the direct dependency session files.
func gatherDependencyFindings(sessionsDir string, dependencies []string) (string, []string, int) {
	var findings []string
	var sessionIDs []string
	totalBytes := 0

	for _, depID := range dependencies {
		content := readSessionFile(sessionsDir, depID)
		if content == "" {
			continue
		}

		sessionIDs = append(sessionIDs, depID)

		_, body, _ := ParseFrontmatter([]byte(content))
		excerpt := extractFindings(string(body), depID)
		findings = append(findings, excerpt)
		totalBytes += len(excerpt)
	}

	return strings.Join(findings, "\n\n---\n\n"), sessionIDs, totalBytes
}

// readSessionFile reads a session file from the sessions directory.
// Tries multiple filename patterns.
func readSessionFile(sessionsDir string, sessionID string) string {
	patterns := []string{
		sessionID + ".md",
		strings.ToLower(sessionID) + ".md",
		strings.ReplaceAll(sessionID, "-", "_") + ".md",
	}

	for _, pattern := range patterns {
		path := filepath.Join(sessionsDir, pattern)
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data)
		}
	}

	// Try to find any file containing the session ID
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if strings.Contains(strings.ToUpper(e.Name()), strings.ToUpper(sessionID)) {
			path := filepath.Join(sessionsDir, e.Name())
			data, err := os.ReadFile(path)
			if err == nil {
				return string(data)
			}
		}
	}

	return ""
}

// extractFindings creates a compact context excerpt from a session body.
// Focuses on headings, recommendations, and evidence-graded claims.
func extractFindings(body string, sessionID string) string {
	if body == "" {
		return ""
	}

	var excerpt strings.Builder
	excerpt.WriteString(fmt.Sprintf("## Findings from %s\n\n", sessionID))

	lines := strings.Split(body, "\n")
	inRelevantSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Always include headings
		if strings.HasPrefix(trimmed, "#") {
			excerpt.WriteString(line + "\n")
			lower := strings.ToLower(trimmed)
			inRelevantSection = strings.Contains(lower, "recommendation") ||
				strings.Contains(lower, "finding") ||
				strings.Contains(lower, "conclusion") ||
				strings.Contains(lower, "decision") ||
				strings.Contains(lower, "verdict") ||
				strings.Contains(lower, "summary") ||
				strings.Contains(lower, "result")
			continue
		}

		// Include content from relevant sections
		if inRelevantSection && trimmed != "" {
			excerpt.WriteString(line + "\n")
		}

		// Always include lines with evidence grades
		if evidenceGradePattern.MatchString(trimmed) {
			if !inRelevantSection {
				excerpt.WriteString(line + "\n")
			}
		}
	}

	result := excerpt.String()

	// If the excerpt is too small, include the first 2000 chars of the body
	if len(result) < 200 && len(body) > 0 {
		maxLen := 2000
		if len(body) < maxLen {
			maxLen = len(body)
		}
		result = fmt.Sprintf("## Findings from %s\n\n%s", sessionID, body[:maxLen])
		if len(body) > maxLen {
			result += "\n\n... [truncated for context injection]"
		}
	}

	return result
}

// injectIntoPrompt replaces a slot placeholder with the given content.
// If the slot doesn't exist in the prompt, appends the content at the end.
func injectIntoPrompt(prompt, slot, content string) string {
	if content == "" {
		return prompt
	}

	if strings.Contains(prompt, slot) {
		return strings.Replace(prompt, slot, content, 1)
	}

	// Append upstream context if no slot found
	return prompt + "\n\n## UPSTREAM RESEARCH CONTEXT\n\n" + content
}
