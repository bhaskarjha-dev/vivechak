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

	// KnownContextSlot is the placeholder for the KNOWN block in 8-block prompts.
	KnownContextSlot = "[KNOWN_CONTEXT]"

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
	isSynthesis := IsSynthesisSession(session.ID)

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
		excerpt := extractFindings(string(body), id, true)
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
		excerpt := extractFindings(string(body), depID, false)
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
// Buffers headings so they are only emitted if substantive content follows,
// and ensures substantive body content is measured before skipping the raw body fallback.
// When isSynthesis is true, rejected alternatives & tradeoffs are retained so FAD
// synthesis (Section 5) has access to them (Principle P8). When false, they are excluded
// to prevent context contamination in intermediate research prompts (Principle P6).
func extractFindings(body string, sessionID string, isSynthesis bool) string {
	if body == "" {
		return ""
	}

	var substantiveContent strings.Builder
	type headingEntry struct {
		level int
		line  string
	}
	var pendingHeadings []headingEntry
	substantiveBytes := 0
	hasRelevantHeading := false

	lines := strings.Split(body, "\n")
	inRelevantSection := false
	inRejectedSection := false
	fenceLen := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track code fence boundaries
		nTicks := countLeadingBackticks(trimmed)
		if fenceLen == 0 {
			if nTicks >= 3 {
				fenceLen = nTicks
				if inRelevantSection && !inRejectedSection {
					for _, h := range pendingHeadings {
						substantiveContent.WriteString(h.line + "\n")
					}
					pendingHeadings = nil
					substantiveContent.WriteString(line + "\n")
					substantiveBytes += len(trimmed)
				}
				continue
			}
		} else {
			if nTicks >= fenceLen && strings.TrimSpace(trimmed[nTicks:]) == "" {
				fenceLen = 0
				if inRelevantSection && !inRejectedSection {
					substantiveContent.WriteString(line + "\n")
					substantiveBytes += len(trimmed)
				}
				continue
			}
			// Line is inside code block; preserve if in relevant section without parsing as headings
			if inRelevantSection && !inRejectedSection && trimmed != "" {
				for _, h := range pendingHeadings {
					substantiveContent.WriteString(h.line + "\n")
				}
				pendingHeadings = nil
				substantiveContent.WriteString(line + "\n")
				substantiveBytes += len(trimmed)
			}
			continue
		}

		// Always check headings (only outside code blocks)
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			if !isSynthesis && (strings.Contains(lower, "rejected") || strings.Contains(lower, "alternatives considered")) {
				inRejectedSection = true
				inRelevantSection = false
				pendingHeadings = nil
				continue
			}
			inRejectedSection = false
			inRelevantSection = strings.Contains(lower, "recommend") ||
				strings.Contains(lower, "finding") ||
				strings.Contains(lower, "conclusion") ||
				strings.Contains(lower, "decision") ||
				strings.Contains(lower, "verdict") ||
				strings.Contains(lower, "summary") ||
				strings.Contains(lower, "result") ||
				strings.Contains(lower, "shortlist") ||
				strings.Contains(lower, "candidate") ||
				strings.Contains(lower, "chosen") ||
				strings.Contains(lower, "propos") ||
				strings.Contains(lower, "takeaway") ||
				strings.Contains(lower, "architecture") ||
				strings.Contains(lower, "matrix") ||
				strings.Contains(lower, "evaluat") ||
				strings.Contains(lower, "risk") ||
				strings.Contains(lower, "concern") ||
				strings.Contains(lower, "tradeoff") ||
				strings.Contains(lower, "failure") ||
				strings.Contains(lower, "premortem") ||
				strings.Contains(lower, "sensitiv") ||
				strings.Contains(lower, "criteri") ||
				(isSynthesis && (strings.Contains(lower, "rejected") || strings.Contains(lower, "alternatives considered")))

			if inRelevantSection {
				hasRelevantHeading = true
			}

			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			for len(pendingHeadings) > 0 && pendingHeadings[len(pendingHeadings)-1].level >= level {
				pendingHeadings = pendingHeadings[:len(pendingHeadings)-1]
			}
			pendingHeadings = append(pendingHeadings, headingEntry{level: level, line: line})
			continue
		}

		if inRejectedSection {
			continue
		}

		includeLine := false
		if inRelevantSection && trimmed != "" {
			includeLine = true
		} else if evidenceGradePattern.MatchString(trimmed) {
			includeLine = true
		}

		if includeLine {
			for _, h := range pendingHeadings {
				substantiveContent.WriteString(h.line + "\n")
			}
			pendingHeadings = nil
			substantiveContent.WriteString(line + "\n")
			substantiveBytes += len(trimmed)
		}
	}

	// If no substantive content was extracted, or if no relevant headings were matched
	// and only minimal stray text (< 150 bytes) was extracted, fall back to raw body (up to 4000 chars)
	// so downstream prompts are not starved.
	if (substantiveBytes == 0 || (!hasRelevantHeading && substantiveBytes < 150)) && len(body) > 0 {
		maxLen := 4000
		if len(body) < maxLen {
			maxLen = len(body)
		}
		result := fmt.Sprintf("## Findings from %s\n\n%s", sessionID, body[:maxLen])
		if len(body) > maxLen {
			result += "\n\n... [truncated for context injection]"
		}
		return result
	}

	return fmt.Sprintf("## Findings from %s\n\n%s", sessionID, substantiveContent.String())
}

// injectIntoPrompt replaces a slot placeholder with the given content.
// If the primary slot doesn't exist, checks for KNOWN_CONTEXT, decision generator
// placeholders (e.g. [PASTE S1 SHORTLIST]), and falls back to appending.
func injectIntoPrompt(prompt, slot, content string) string {
	if content == "" {
		return prompt
	}

	if strings.Contains(prompt, slot) {
		result := strings.Replace(prompt, slot, content, 1)
		// Also fill KNOWN_CONTEXT if present
		result = strings.Replace(result, KnownContextSlot, content, 1)
		return result
	}

	// Try KNOWN_CONTEXT slot
	if strings.Contains(prompt, KnownContextSlot) {
		return strings.Replace(prompt, KnownContextSlot, content, 1)
	}

	// Support decision generator candidate shortlist placeholder
	const decisionShortlistSlot = "[PASTE S1 SHORTLIST]"
	if strings.Contains(prompt, decisionShortlistSlot) {
		return strings.Replace(prompt, decisionShortlistSlot, content, 1)
	}

	// Append upstream context if no slot found
	return prompt + "\n\n## UPSTREAM RESEARCH CONTEXT\n\n" + content
}

// countLeadingBackticks returns the number of consecutive leading backticks.
func countLeadingBackticks(s string) int {
	count := 0
	for count < len(s) && s[count] == '`' {
		count++
	}
	return count
}

