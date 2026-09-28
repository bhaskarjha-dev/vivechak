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

	// MaxInjectedBytesWarning is the threshold (100KB) above which an injection size warning is emitted.
	MaxInjectedBytesWarning = 100 * 1024
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

	// Warning is set if injected context exceeds size guidelines.
	Warning string `json:"warning,omitempty"`
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
		findings, sessions, totalBytes, err := gatherAllFindings(sessionsDir, completedSessions)
		if err != nil {
			return nil, fmt.Errorf("gathering all findings: %w", err)
		}
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
		findings, sessions, totalBytes, err := gatherDependencyFindings(sessionsDir, session.Dependencies)
		if err != nil {
			return nil, fmt.Errorf("gathering dependency findings: %w", err)
		}
		prompt = injectIntoPrompt(prompt, UpstreamFindingsSlot, findings)
		result.UpstreamSessions = sessions
		result.InjectedBytes = totalBytes
	}

	result.InjectedPrompt = prompt
	if result.InjectedBytes > MaxInjectedBytesWarning {
		result.Warning = fmt.Sprintf("W-INJECTION-SIZE: Injected upstream context is %d KB (exceeds %d KB threshold). Consider consolidating upstream sessions.", result.InjectedBytes/1024, MaxInjectedBytesWarning/1024)
	}
	return result, nil
}

// gatherAllFindings reads all completed session files and assembles their
// findings into a single context block.
func gatherAllFindings(sessionsDir string, completedSessions map[string]bool) (string, []string, int, error) {
	var findings []string
	var sessionIDs []string
	totalBytes := 0
	seenFiles := make(map[string]bool)

	// Collect and sort IDs for deterministic ordering
	var sortedIDs []string
	for id := range completedSessions {
		sortedIDs = append(sortedIDs, id)
	}
	sort.Strings(sortedIDs)

	for _, id := range sortedIDs {
		content, fileName, err := readSessionFile(sessionsDir, id)
		if err != nil {
			return "", nil, 0, err
		}
		if content == "" || seenFiles[fileName] {
			continue
		}
		seenFiles[fileName] = true

		sessionIDs = append(sessionIDs, id)

		// Extract key findings (frontmatter body)
		_, body, _ := ParseFrontmatter([]byte(content))
		excerpt := extractFindings(string(body), id)
		findings = append(findings, excerpt)
		totalBytes += len(excerpt)
	}

	return strings.Join(findings, "\n\n---\n\n"), sessionIDs, totalBytes, nil
}

// gatherDependencyFindings reads only the direct dependency session files.
func gatherDependencyFindings(sessionsDir string, dependencies []string) (string, []string, int, error) {
	var findings []string
	var sessionIDs []string
	totalBytes := 0
	seenFiles := make(map[string]bool)

	for _, depID := range dependencies {
		content, fileName, err := readSessionFile(sessionsDir, depID)
		if err != nil {
			return "", nil, 0, err
		}
		if content == "" || seenFiles[fileName] {
			continue
		}
		seenFiles[fileName] = true

		sessionIDs = append(sessionIDs, depID)

		_, body, _ := ParseFrontmatter([]byte(content))
		excerpt := extractFindings(string(body), depID)
		findings = append(findings, excerpt)
		totalBytes += len(excerpt)
	}

	return strings.Join(findings, "\n\n---\n\n"), sessionIDs, totalBytes, nil
}

// readSessionFile reads a session file from the sessions directory.
// Tries exact filename patterns first, then prefix-based matching.
// Returns empty content with nil error if file not found (normal condition).
func readSessionFile(sessionsDir string, sessionID string) (string, string, error) {
	patterns := []string{
		sessionID + ".md",
		strings.ToLower(sessionID) + ".md",
		strings.ReplaceAll(sessionID, "-", "_") + ".md",
	}

	for _, pattern := range patterns {
		path := filepath.Join(sessionsDir, pattern)
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), pattern, nil
		}
		if !os.IsNotExist(err) {
			return "", "", fmt.Errorf("reading session file %s: %w", path, err)
		}
	}

	// Try to find a file whose name starts with the session ID (prefix match only).
	// This handles cases like "T1-01-database-selection.md" for session "T1-01".
	// We do NOT use Contains — that would match "T1-02.md" for session "T1".
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return "", "", fmt.Errorf("reading sessions directory: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		nameStem := strings.TrimSuffix(e.Name(), ".md")
		nameUpper := strings.ToUpper(nameStem)
		idUpper := strings.ToUpper(sessionID)

		matchesPrefix := false
		if nameUpper == idUpper {
			matchesPrefix = true
		} else if strings.HasPrefix(nameUpper, idUpper+"-") || strings.HasPrefix(nameUpper, idUpper+"_") {
			// Ensure it's not a compound ID like T1 matching T1-02.
			// If sessionID has no hyphen and the suffix starts with a digit, it's a sub-session, not a slug.
			suffix := nameUpper[len(idUpper)+1:]
			if strings.Contains(sessionID, "-") || (len(suffix) > 0 && (suffix[0] < '0' || suffix[0] > '9')) {
				matchesPrefix = true
			}
		}

		if matchesPrefix {
			path := filepath.Join(sessionsDir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return "", "", fmt.Errorf("reading session file %s: %w", path, err)
			}
			// If file has frontmatter with an explicit session ID, that ID is authoritative.
			if fm, _, fmErr := ParseFrontmatter(data); fmErr == nil && fm != nil {
				fileSID := fm.GetString("session_id")
				if fileSID == "" {
					fileSID = fm.GetString("id")
				}
				if fileSID != "" && !strings.EqualFold(fileSID, sessionID) {
					// Frontmatter explicitly declares a different session ID. Skip.
					continue
				}
			}
			return string(data), e.Name(), nil
		}
	}

	// Not an error — session file simply doesn't exist yet
	return "", "", nil
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

		// Always include headings, except rejected alternatives sections
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "rejected") || strings.Contains(lower, "alternatives considered") {
				inRelevantSection = false
				continue
			}
			excerpt.WriteString(line + "\n")
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

	// If the excerpt is too small, supplement with raw body (up to 4000 chars)
	if len(result) < 200 && len(body) > 0 {
		maxLen := 4000
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
