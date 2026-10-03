package core

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Session represents a single research session in the pipeline DAG.
type Session struct {
	// ID is the session identifier (e.g., "T1-01", "T2-01", "SYN-01").
	ID string `json:"id"`

	// Title is the human-readable session title.
	Title string `json:"title,omitempty"`

	// Layer is the pipeline layer (0 = landscape, 1 = deep-dive, etc.)
	Layer int `json:"layer"`

	// DoorType classifies the decision as one-way or two-way.
	DoorType string `json:"door_type,omitempty"`

	// DecisionRef links to the decision this session informs (e.g., "D-001").
	DecisionRef string `json:"decision_ref,omitempty"`

	// Dependencies lists session IDs that must complete before this one.
	Dependencies []string `json:"dependencies,omitempty"`

	// OutputFile is the expected output filename.
	OutputFile string `json:"output_file,omitempty"`

	// Prompt is the copy-paste-ready research prompt from the pipeline.
	Prompt string `json:"prompt,omitempty"`
}

// DAG represents the research pipeline as a directed acyclic graph of sessions.
type DAG struct {
	// Sessions is the ordered list of sessions in the pipeline.
	Sessions []Session `json:"sessions"`

	// Archetype is the inferred project archetype.
	Archetype string `json:"archetype,omitempty"`

	// ComplexityScore is the total complexity score.
	ComplexityScore int `json:"complexity_score,omitempty"`

	// Tier is the inferred tier (e.g., "Tier 1 (4-8 Sessions)").
	Tier string `json:"tier,omitempty"`
}

// SessionByID returns the session with the given ID, or nil if not found (case-insensitive).
func (d *DAG) SessionByID(id string) *Session {
	for i := range d.Sessions {
		if strings.EqualFold(d.Sessions[i].ID, id) {
			return &d.Sessions[i]
		}
	}
	return nil
}

// IsSynthesisSession reports whether id represents a synthesis or Founding Architecture Document
// session (e.g. "FAD", "SYN", "SYN-01", "SYN-02").
func IsSynthesisSession(id string) bool {
	upper := strings.ToUpper(strings.TrimSpace(id))
	return upper == "FAD" || strings.HasPrefix(upper, "SYN")
}


var validSessionIDRe = regexp.MustCompile(`^[A-Za-z0-9]+(?:[-_][A-Za-z0-9]+)*$`)

// ValidateDAG checks the pipeline graph for cycles, unknown dependencies, and duplicates.
func (d *DAG) ValidateDAG() error {
	ids := make(map[string]bool)
	for _, s := range d.Sessions {
		if strings.TrimSpace(s.ID) == "" {
			return fmt.Errorf("session with empty ID found")
		}
		upperID := strings.ToUpper(s.ID)
		if ids[upperID] {
			return fmt.Errorf("duplicate session ID: %s", s.ID)
		}
		ids[upperID] = true
		ids[s.ID] = true
	}

	// Verify session IDs have valid identifier format
	for _, s := range d.Sessions {
		if !validSessionIDRe.MatchString(s.ID) {
			return fmt.Errorf("session ID %q is invalid — use alphanumeric characters with optional hyphens or underscores (e.g. 'T1-01', 'SYN-01')", s.ID)
		}
	}

	// Check for dangling dependencies
	for _, s := range d.Sessions {
		for _, dep := range s.Dependencies {
			if !ids[dep] && !ids[strings.ToUpper(dep)] {
				return fmt.Errorf("session %s depends on nonexistent session %s", s.ID, dep)
			}
		}
	}

	// Cycle detection using DFS (0=unvisited, 1=visiting, 2=visited)
	visited := make(map[string]int)
	var dfs func(id string, path []string) error
	dfs = func(id string, path []string) error {
		visited[id] = 1 // visiting (gray)
		path = append(path, id)

		session := d.SessionByID(id)
		if session != nil {
			for _, dep := range session.Dependencies {
				if visited[dep] == 1 {
					return fmt.Errorf("cycle detected in pipeline DAG: %s -> %s", strings.Join(path, " -> "), dep)
				}
				if visited[dep] == 0 {
					if err := dfs(dep, path); err != nil {
						return err
					}
				}
			}
		}

		visited[id] = 2 // visited (black)
		return nil
	}

	for _, s := range d.Sessions {
		if visited[s.ID] == 0 {
			if err := dfs(s.ID, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

// NextSessions returns sessions whose dependencies are all satisfied.
// completedIDs is the set of session IDs that have been completed.
func (d *DAG) NextSessions(completedIDs map[string]bool) []Session {
	var ready []Session
	for _, s := range d.Sessions {
		// Skip already completed (case-insensitive)
		if completedIDs[s.ID] || completedIDs[strings.ToUpper(s.ID)] || completedIDs[strings.ToLower(s.ID)] {
			continue
		}

		// Check if all dependencies are met (case-insensitive)
		allDeps := true
		for _, dep := range s.Dependencies {
			if !completedIDs[dep] && !completedIDs[strings.ToUpper(dep)] && !completedIDs[strings.ToLower(dep)] {
				allDeps = false
				break
			}
		}

		if allDeps {
			ready = append(ready, s)
		}
	}
	return ready
}

// TransitiveDependents returns all session IDs that transitively depend on the given ID.
func (d *DAG) TransitiveDependents(id string) []string {
	if d == nil {
		return nil
	}
	// Build reverse adjacency map
	reverseDeps := make(map[string][]string)
	for _, s := range d.Sessions {
		for _, dep := range s.Dependencies {
			reverseDeps[dep] = append(reverseDeps[dep], s.ID)
		}
	}
	// BFS from id
	visited := map[string]bool{}
	queue := []string{id}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, downstream := range reverseDeps[current] {
			if !visited[downstream] {
				visited[downstream] = true
				queue = append(queue, downstream)
			}
		}
	}
	var result []string
	for k := range visited {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// Patterns for parsing the pipeline format
var (
	// Matches session headers like:
	// - #### T1-01: Graph Persistence Landscape
	// - ### Session T1-01: Spike
	// - ### Session S1: Spike
	// - #### S1: Landscape
	// - #### D-001-S1 — Spike
	sessionHeaderRe = regexp.MustCompile(`^#{1,4}\s+(?:(?:Session\s+([A-Za-z0-9]+(?:[-_][A-Za-z0-9]+)*))|([A-Za-z0-9]+(?:-[A-Za-z0-9]+)+)|([A-Za-z]+\d+))\s*[:—\-]\s*(.*)`)

	// Matches metadata table rows like: | **ID** | T1-01 |
	metaFieldRe = regexp.MustCompile(`\|\s*\*\*([^*]+)\*\*\s*\|\s*(.+?)\s*\|`)

	// Matches horizontal session table rows like: | D-015-S1 | Comparison | none | sessions/D-015-S1-cache-comparison.md |
	sessionTableRowRe = regexp.MustCompile(`^\|\s*([A-Za-z0-9]+(?:[-_][A-Za-z0-9]+)*)\s*\|\s*([^|]+)\|\s*([^|]+)\|\s*([^|]+)\|`)

	// Matches words matching session ID pattern inside a brief header
	briefIDRe = regexp.MustCompile(`[A-Za-z0-9]+(?:[-_][A-Za-z0-9]+)*`)

	// Matches the prompt code block
	promptStartRe = regexp.MustCompile("^`{3,}prompt")
)

// ParsePipeline parses a RESEARCH-PIPELINE.md or decision plan file into a DAG structure.
func ParsePipeline(data []byte) (*DAG, error) {
	dag := &DAG{}
	lines := strings.Split(string(data), "\n")

	sessionsByID := make(map[string]*Session)
	var sessionOrder []string

	getOrCreateSession := func(id string) *Session {
		id = strings.TrimSpace(id)
		if s, ok := sessionsByID[id]; ok {
			return s
		}
		s := &Session{ID: id}
		sessionsByID[id] = s
		sessionOrder = append(sessionOrder, id)
		return s
	}

	var currentSession *Session
	inPrompt := false
	var promptLines []string
	outerFenceLen := 0
	inInnerBlock := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Collect prompt block if inside prompt
		if inPrompt {
			numBackticks := 0
			for j := 0; j < len(trimmed) && trimmed[j] == '`'; j++ {
				numBackticks++
			}

			isClosingFence := false
			if outerFenceLen >= 4 {
				if numBackticks >= outerFenceLen && len(strings.TrimSpace(trimmed[numBackticks:])) == 0 {
					isClosingFence = true
				}
			} else {
				if numBackticks >= 3 {
					afterTicks := strings.TrimSpace(trimmed[numBackticks:])
					if afterTicks != "" {
						inInnerBlock = true
					} else {
						if inInnerBlock {
							inInnerBlock = false
						} else {
							isClosingFence = true
						}
					}
				}
			}

			if isClosingFence {
				inPrompt = false
				promptContent := strings.Join(promptLines, "\n")
				targetSession := currentSession

				// Check if prompt header mentions a specific session ID
				nonEmptyLines := 0
				for _, pLine := range promptLines {
					pTrim := strings.TrimSpace(pLine)
					if pTrim == "" {
						continue
					}
					nonEmptyLines++
					if strings.HasPrefix(pTrim, "#") {
						found := false
						for _, match := range briefIDRe.FindAllString(pTrim, -1) {
							if s, ok := sessionsByID[match]; ok {
								targetSession = s
								found = true
								break
							}
						}
						if found {
							break
						}
					}
					if nonEmptyLines >= 5 {
						break
					}
				}

				// If target has a prompt already, assign to next unprompted session
				if targetSession == nil || targetSession.Prompt != "" {
					for _, id := range sessionOrder {
						if sessionsByID[id].Prompt == "" {
							targetSession = sessionsByID[id]
							break
						}
					}
				}

				if targetSession != nil {
					targetSession.Prompt = promptContent
				}
				promptLines = nil
			} else {
				promptLines = append(promptLines, line)
			}
			continue
		}

		// Check for prompt block start
		if promptStartRe.MatchString(trimmed) {
			outerFenceLen = 0
			for j := 0; j < len(trimmed) && trimmed[j] == '`'; j++ {
				outerFenceLen++
			}
			inPrompt = true
			inInnerBlock = false
			promptLines = nil
			continue
		}

		// Check for session header
		if matches := sessionHeaderRe.FindStringSubmatch(trimmed); matches != nil {
			id := ""
			for k := 1; k <= 3; k++ {
				if matches[k] != "" {
					id = strings.TrimSpace(matches[k])
					break
				}
			}
			title := strings.TrimSpace(strings.TrimPrefix(matches[4], "—"))
			title = strings.TrimSpace(strings.TrimPrefix(title, "-"))
			title = strings.TrimSpace(strings.TrimPrefix(title, ":"))
			if id != "" {
				currentSession = getOrCreateSession(id)
				if title != "" && currentSession.Title == "" {
					currentSession.Title = title
				}
				continue
			}
		}

		// Check for horizontal session table row (e.g. decision plans)
		if strings.Contains(line, "|") && !strings.Contains(line, "**") && !strings.Contains(line, "---") {
			if matches := sessionTableRowRe.FindStringSubmatch(trimmed); matches != nil {
				idVal := strings.TrimSpace(matches[1])
				if !strings.EqualFold(idVal, "session-id") && !strings.EqualFold(idVal, "id") && !strings.EqualFold(idVal, "session") {
					currentSession = getOrCreateSession(idVal)
					if len(currentSession.Dependencies) == 0 {
						currentSession.Dependencies = parseDependencies(matches[3])
					}
					if currentSession.OutputFile == "" {
						currentSession.OutputFile = strings.Trim(strings.TrimSpace(matches[4]), "`")
					}
					continue
				}
			}
		}

		// Parse metadata table rows
		if strings.Contains(line, "**") && strings.Contains(line, "|") {
			if matches := metaFieldRe.FindStringSubmatch(line); matches != nil {
				field := strings.TrimSpace(matches[1])
				value := strings.TrimSpace(matches[2])

				switch strings.ToLower(field) {
				case "id", "session id":
					currentSession = getOrCreateSession(value)
				case "layer":
					if currentSession != nil {
						if l, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
							currentSession.Layer = l
						}
					}
				case "door type":
					if currentSession != nil && currentSession.DoorType == "" {
						currentSession.DoorType = value
					}
				case "decision":
					if currentSession != nil && currentSession.DecisionRef == "" {
						currentSession.DecisionRef = value
					}
				case "dependencies":
					if currentSession != nil && len(currentSession.Dependencies) == 0 {
						currentSession.Dependencies = parseDependencies(value)
					}
				case "output file":
					if currentSession != nil && currentSession.OutputFile == "" {
						currentSession.OutputFile = strings.Trim(value, "`")
					}
				}
			}
			continue
		}

		// Parse complexity score
		if strings.Contains(trimmed, "Complexity Score") && strings.Contains(trimmed, "→") {
			// Extract score and tier
			if idx := strings.Index(trimmed, "→"); idx > 0 {
				dag.Tier = strings.TrimSpace(trimmed[idx+len("→"):])
				beforeArrow := trimmed[:idx]
				re := regexp.MustCompile(`(\d+)(?:\s*/\s*\d+)?`)
				if matches := re.FindStringSubmatch(beforeArrow); len(matches) > 1 {
					if score, err := strconv.Atoi(matches[1]); err == nil {
						dag.ComplexityScore = score
					}
				}
			}
		}

		// Parse archetype
		if strings.Contains(trimmed, "Primary Archetype") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				dag.Archetype = strings.TrimSpace(strings.Trim(parts[1], "* "))
			}
		}
	}

	// Handle unclosed prompt block at EOF
	if inPrompt && len(promptLines) > 0 {
		promptContent := strings.Join(promptLines, "\n")
		targetSession := currentSession
		nonEmptyLines := 0
		for _, pLine := range promptLines {
			pTrim := strings.TrimSpace(pLine)
			if pTrim == "" {
				continue
			}
			nonEmptyLines++
			if strings.HasPrefix(pTrim, "#") {
				found := false
				for _, match := range briefIDRe.FindAllString(pTrim, -1) {
					if s, ok := sessionsByID[match]; ok {
						targetSession = s
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if nonEmptyLines >= 5 {
				break
			}
		}
		if targetSession == nil || targetSession.Prompt != "" {
			for _, id := range sessionOrder {
				if sessionsByID[id].Prompt == "" {
					targetSession = sessionsByID[id]
					break
				}
			}
		}
		if targetSession != nil {
			targetSession.Prompt = promptContent
		}
	}

	// Assemble deduplicated sessions in order of appearance
	for _, id := range sessionOrder {
		dag.Sessions = append(dag.Sessions, *sessionsByID[id])
	}

	if len(dag.Sessions) == 0 {
		return dag, fmt.Errorf("no sessions found in pipeline")
	}

	return dag, nil
}

// Session IDs are used exactly as they appear in the pipeline.
// No normalization is performed. IDs must be consistent between
// the header, metadata table, and dependency lists.
// Canonical format uses alphanumeric tokens with optional hyphens or underscores
// (e.g., "T1-01", "SYN-01", "D-001-S1", "S1").

var markdownLinkRe = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)

// parseDependencies parses a dependency list from the metadata table.
// Handles formats like: "None (parallel)", "T1-01", "T1-01, T1-02", "T1-01 (soft)",
// "[T1-01, T1-02]", "`T1-01` and `T1-02`", "T1-01 & T1-02",
// markdown links "[R-01](./R-01.md)", and bullet lists "- R-01\n- R-02".
func parseDependencies(value string) []string {
	value = strings.TrimSpace(value)

	// Strip markdown link syntax: [text](url) → text
	value = markdownLinkRe.ReplaceAllString(value, "$1")

	// Normalize newlines to commas (for bullet lists or multi-line dependencies)
	value = strings.ReplaceAll(value, "\\n", ", ")
	value = strings.ReplaceAll(value, "\r\n", ", ")
	value = strings.ReplaceAll(value, "\n", ", ")

	// Clean enclosing brackets, backticks, quotes
	value = strings.ReplaceAll(value, "[", "")
	value = strings.ReplaceAll(value, "]", "")
	value = strings.ReplaceAll(value, "`", "")
	value = strings.ReplaceAll(value, "\"", "")
	value = strings.ReplaceAll(value, "'", "")

	// Strip trailing punctuation from entire string
	value = strings.Trim(value, " \t\r\n.,;:[]`'\"")

	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "none" || strings.HasPrefix(lower, "none (") || strings.HasPrefix(lower, "none(") ||
		strings.HasPrefix(lower, "none.") || lower == "—" || lower == "-" || lower == "n/a" || lower == "nil" || lower == "null" || lower == "" {
		return nil
	}

	// Normalize conjunctions
	value = strings.ReplaceAll(value, " and ", ", ")
	value = strings.ReplaceAll(value, " & ", ", ")

	parts := strings.Split(value, ",")
	var deps []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		// Strip bullet markers (- or * or •)
		p = strings.TrimLeft(p, "-*• \t")
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, "("); idx > 0 {
			p = strings.TrimSpace(p[:idx])
		}
		p = strings.TrimLeft(p, "-*• \t")
		p = strings.Trim(p, " \t\r\n.,;:[]`'\"")
		if p != "" && !strings.EqualFold(p, "none") && !strings.EqualFold(p, "n/a") && !strings.EqualFold(p, "nil") && !strings.EqualFold(p, "null") && !strings.EqualFold(p, "and") {
			deps = append(deps, p)
		}
	}
	return deps
}
