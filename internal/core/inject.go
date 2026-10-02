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

const synthesisContextBudget = 20 * 1024 // 20KB ~5000 tokens

// SessionTechRow represents a row in the synthesized technology matrix.
type SessionTechRow struct {
	SessionID string
	Topic     string
	Tech      string
	Tags      string
}

// extractTags returns frontmatter tags as a comma-separated string.
func extractTags(fm Frontmatter) string {
	if fm == nil {
		return ""
	}
	if slice := fm.GetStringSlice("tags"); len(slice) > 0 {
		return strings.Join(slice, ", ")
	}
	return fm.GetString("tags")
}

func cleanTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", "/")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

// extractRecommendation extracts the primary recommendation from session body.
func extractRecommendation(body string) string {
	lines := strings.Split(body, "\n")
	inRec := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "recommend") {
				if idx := strings.Index(trimmed, ":"); idx != -1 && idx+1 < len(trimmed) {
					candidate := strings.TrimSpace(trimmed[idx+1:])
					if candidate != "" {
						return cleanTableCell(candidate)
					}
				}
				if idx := strings.Index(trimmed, "—"); idx != -1 && idx+len("—") < len(trimmed) {
					candidate := strings.TrimSpace(trimmed[idx+len("—"):])
					if candidate != "" {
						return cleanTableCell(candidate)
					}
				}
				inRec = true
				continue
			} else if inRec {
				break
			}
		}
		if inRec && trimmed != "" {
			val := strings.TrimLeft(trimmed, "-*# ")
			val = strings.Trim(val, "*_`")
			val = cleanTableCell(val)
			if val != "" {
				if len(val) > 120 {
					val = val[:117] + "..."
				}
				return val
			}
		}
	}
	return "See session findings"
}

// buildTechnologyMatrix creates a markdown table summarizing technology choices.
func buildTechnologyMatrix(rows []SessionTechRow) string {
	if len(rows) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## TECHNOLOGY CHOICES ACROSS SESSIONS\n\n")
	sb.WriteString("| Session | Topic | Chosen Technology | Tags |\n")
	sb.WriteString("|---|---|---|---|\n")
	for _, r := range rows {
		topic := cleanTableCell(r.Topic)
		if topic == "" {
			topic = "Research"
		}
		tech := cleanTableCell(r.Tech)
		tags := cleanTableCell(r.Tags)
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", r.SessionID, topic, tech, tags))
	}
	sb.WriteString("\n**Check for coherence:** Do these technology choices work together? Flag any\nruntime conflicts (e.g. CGo requirements across multiple sessions, conflicting\nlanguage runtimes, incompatible dependency versions).\n")
	return sb.String()
}

// budgetFindings progressively trims findings if total exceeds the byte budget.
func budgetFindings(findings []string, sessionIDs []string, budget int) []string {
	total := 0
	for _, f := range findings {
		total += len(f)
	}
	if total <= budget || len(findings) == 0 {
		return findings
	}
	perSession := budget / len(findings)
	trimmed := make([]string, len(findings))
	for i, f := range findings {
		if len(f) <= perSession {
			trimmed[i] = f
		} else {
			id := "session"
			if i < len(sessionIDs) {
				id = sessionIDs[i]
			}
			trimmed[i] = f[:perSession] + "\n\n... [trimmed for synthesis context budget — see sessions/" + id + ".md for full findings]\n"
		}
	}
	return trimmed
}

// gatherAllFindings reads all completed session files and assembles their
// findings into a single context block. For synthesis, also extracts ## Delta
// sections and aggregates them into a belief evolution block.
func gatherAllFindings(sessionsDir string, completedSessions map[string]bool) (string, []string, int, error) {
	var findings []string
	var deltaEntries []string
	var sessionIDs []string
	var techRows []SessionTechRow
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
		fm, body, _ := ParseFrontmatter([]byte(content))
		excerpt := extractFindings(string(body), id, true)
		findings = append(findings, excerpt)

		// Technology row for cross-session coherence
		topic := id
		if fm != nil {
			if t := fm.GetString("title"); t != "" {
				topic = t
			}
		}
		tags := extractTags(fm)
		rec := extractRecommendation(string(body))
		techRows = append(techRows, SessionTechRow{
			SessionID: id,
			Topic:     topic,
			Tech:      rec,
			Tags:      tags,
		})

		// Extract Delta section for belief evolution aggregation
		if delta := extractDeltaSection(string(body)); delta != "" {
			title := id
			if fm != nil {
				if t := fm.GetString("title"); t != "" {
					title = id + " — " + t
				}
			}
			deltaEntries = append(deltaEntries, fmt.Sprintf("### %s\n%s", title, delta))
		}
	}

	// Apply synthesis context density budgeting if needed
	findings = budgetFindings(findings, sessionIDs, synthesisContextBudget)

	var parts []string
	if len(techRows) > 0 {
		parts = append(parts, buildTechnologyMatrix(techRows))
	}
	if len(findings) > 0 {
		parts = append(parts, strings.Join(findings, "\n\n---\n\n"))
	}
	if len(deltaEntries) > 0 {
		beliefEvolution := "## BELIEF EVOLUTION ACROSS SESSIONS\n\n" +
			strings.Join(deltaEntries, "\n\n")
		parts = append(parts, beliefEvolution)
	}

	result := strings.Join(parts, "\n\n---\n\n")
	totalBytes := len(result)

	return result, sessionIDs, totalBytes, nil
}


// extractDeltaSection extracts the content under a ## Delta heading from session body.
// Returns the delta content (without the heading itself), or empty string if not found.
func extractDeltaSection(body string) string {
	lines := strings.Split(body, "\n")
	var deltaContent strings.Builder
	inDelta := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "delta") && !inDelta {
				inDelta = true
				continue
			} else if inDelta {
				// Hit another heading at same or higher level — stop
				level := 0
				for level < len(trimmed) && trimmed[level] == '#' {
					level++
				}
				if level <= 2 { // ## or # heading ends the delta section
					break
				}
				// Sub-heading within delta — include it
				deltaContent.WriteString(line + "\n")
				continue
			}
		}
		if inDelta {
			deltaContent.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(deltaContent.String())
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

// ReadSessionFilePublic is the exported wrapper around readSessionFile.
// Used by the MCP layer for auto-drafting decisions from session content.
func ReadSessionFilePublic(sessionsDir string, sessionID string) (string, string, error) {
	return readSessionFile(sessionsDir, sessionID)
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
// Concern/risk/failure headings are tracked separately and emitted as a distinct
// "Upstream Concerns" subsection to make them visible for stress-testing.
func extractFindings(body string, sessionID string, isSynthesis bool) string {
	if body == "" {
		return ""
	}

	var substantiveContent strings.Builder
	var concernContent strings.Builder
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
	inConcernSection := false
	inRecPara := false
	recParaHasText := false
	fenceLen := 0

	// isConcernHeading checks if a lowercase heading indicates a concern/risk section
	isConcernHeading := func(lower string) bool {
		return strings.Contains(lower, "concern") ||
			strings.Contains(lower, "risk") ||
			strings.Contains(lower, "failure") ||
			strings.Contains(lower, "premortem") ||
			strings.Contains(lower, "open question")
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track code fence boundaries
		nTicks := countLeadingBackticks(trimmed)
		if fenceLen == 0 {
			if nTicks >= 3 {
				fenceLen = nTicks
				if inConcernSection {
					concernContent.WriteString(line + "\n")
				} else if inRelevantSection && !inRejectedSection {
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
				if inConcernSection {
					concernContent.WriteString(line + "\n")
				} else if inRelevantSection && !inRejectedSection {
					substantiveContent.WriteString(line + "\n")
					substantiveBytes += len(trimmed)
				}
				continue
			}
			// Line is inside code block; preserve if in relevant section without parsing as headings
			if inConcernSection && trimmed != "" {
				concernContent.WriteString(line + "\n")
			} else if inRelevantSection && !inRejectedSection && trimmed != "" {
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
				inConcernSection = false
				pendingHeadings = nil
				continue
			}
			inRejectedSection = false

			// Check if this is a concern heading — track separately
			if isConcernHeading(lower) {
				inConcernSection = true
				inRelevantSection = false
				inRecPara = false
				pendingHeadings = nil
				// Don't add concern heading to pending — we emit concerns separately
				continue
			}

			inConcernSection = false
			if isSynthesis {
				inRelevantSection = strings.Contains(lower, "recommend") ||
					strings.Contains(lower, "delta") ||
					strings.Contains(lower, "alternative") ||
					strings.Contains(lower, "rejected") ||
					strings.Contains(lower, "amend") ||
					strings.Contains(lower, "post-hoc") ||
					strings.Contains(lower, "update") ||
					strings.Contains(lower, "correction") ||
					strings.Contains(lower, "retract")
				if strings.Contains(lower, "recommend") {
					inRecPara = true
					recParaHasText = false
				} else {
					inRecPara = false
				}
			} else {
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
					strings.Contains(lower, "tradeoff") ||
					strings.Contains(lower, "sensitiv") ||
					strings.Contains(lower, "criteri") ||
					strings.Contains(lower, "amend") ||
					strings.Contains(lower, "post-hoc") ||
					strings.Contains(lower, "update") ||
					strings.Contains(lower, "correction") ||
					strings.Contains(lower, "retract")
			}

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

		// In synthesis mode, limit recommendation to the first paragraph
		if isSynthesis && inRecPara {
			if trimmed == "" {
				if recParaHasText {
					inRelevantSection = false
					inRecPara = false
					continue
				}
				continue
			} else {
				recParaHasText = true
			}
		}

		// Concern section content — track separately
		if inConcernSection && trimmed != "" {
			if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
				concernContent.WriteString(trimmed + "\n")
			} else {
				concernContent.WriteString("- " + trimmed + "\n")
			}
			continue
		}

		includeLine := false
		if inRelevantSection && trimmed != "" {
			includeLine = true
		} else if !isSynthesis && evidenceGradePattern.MatchString(trimmed) {
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
		// Append concerns even in fallback
		if concerns := strings.TrimSpace(concernContent.String()); concerns != "" {
			result += "\n\n### Upstream Concerns (stress-test these)\n" + concerns
		}
		return result
	}

	result := fmt.Sprintf("## Findings from %s\n\n%s", sessionID, substantiveContent.String())

	// Append discovered concerns as a distinct subsection
	if concerns := strings.TrimSpace(concernContent.String()); concerns != "" {
		result += "\n### Upstream Concerns (stress-test these)\n" + concerns + "\n"
	}

	return result
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

