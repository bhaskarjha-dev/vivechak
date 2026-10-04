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

		// Check for horizontal session table row (e.g. decision plans: | ID | Topic | Dependencies | OutputFile |)
		if strings.Contains(line, "|") && !strings.Contains(line, "**") && !strings.Contains(line, "---") && strings.Count(trimmed, "|") == 5 {
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
					if currentSession != nil {
						parsedDeps := parseDependencies(value)
						if len(parsedDeps) > 0 {
							currentSession.Dependencies = parsedDeps
						}
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

// AddSession appends a new research session to the DAG after verifying that
// dependencies exist and no cycles are created.
func (d *DAG) AddSession(s Session) error {
	s.ID = strings.TrimSpace(s.ID)
	if s.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if !validSessionIDRe.MatchString(s.ID) {
		return fmt.Errorf("session ID %q is invalid — use alphanumeric characters with optional hyphens or underscores", s.ID)
	}
	if d.SessionByID(s.ID) != nil {
		return fmt.Errorf("session %s already exists in pipeline", s.ID)
	}

	// Validate dependencies exist
	for _, dep := range s.Dependencies {
		dep = strings.TrimSpace(dep)
		if strings.EqualFold(dep, s.ID) {
			return fmt.Errorf("session %s cannot depend on itself", s.ID)
		}
		if d.SessionByID(dep) == nil {
			return fmt.Errorf("dependency %s does not exist in pipeline", dep)
		}
	}

	if s.OutputFile == "" {
		s.OutputFile = fmt.Sprintf("sessions/%s.md", s.ID)
	}

	d.Sessions = append(d.Sessions, s)
	if err := d.ValidateDAG(); err != nil {
		d.Sessions = d.Sessions[:len(d.Sessions)-1]
		return err
	}
	return nil
}

// RemoveSession removes a session from the DAG if no other sessions depend on it.
func (d *DAG) RemoveSession(id string) error {
	id = strings.TrimSpace(id)
	idx := -1
	for i, s := range d.Sessions {
		if strings.EqualFold(s.ID, id) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("session %s not found in pipeline", id)
	}

	// Check if any other session depends on id
	for _, s := range d.Sessions {
		for _, dep := range s.Dependencies {
			if strings.EqualFold(dep, id) {
				return fmt.Errorf("cannot remove session %s: session %s depends on it", id, s.ID)
			}
		}
	}

	d.Sessions = append(d.Sessions[:idx], d.Sessions[idx+1:]...)
	return nil
}

// UpdateDependencies changes dependencies for a session and verifies no cycles are formed.
func (d *DAG) UpdateDependencies(id string, deps []string) error {
	id = strings.TrimSpace(id)
	s := d.SessionByID(id)
	if s == nil {
		return fmt.Errorf("session %s not found in pipeline", id)
	}

	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if strings.EqualFold(dep, id) {
			return fmt.Errorf("session %s cannot depend on itself", id)
		}
		if d.SessionByID(dep) == nil {
			return fmt.Errorf("dependency %s does not exist in pipeline", dep)
		}
	}

	oldDeps := s.Dependencies
	s.Dependencies = deps
	if err := d.ValidateDAG(); err != nil {
		s.Dependencies = oldDeps
		return err
	}
	return nil
}

// UpdatePrompt replaces the prompt text of an existing session.
func (d *DAG) UpdatePrompt(id string, prompt string) error {
	id = strings.TrimSpace(id)
	s := d.SessionByID(id)
	if s == nil {
		return fmt.Errorf("session %s not found in pipeline", id)
	}
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("prompt content cannot be empty")
	}
	s.Prompt = prompt
	return nil
}

// Serialize re-renders the DAG into Markdown format suitable for RESEARCH-PIPELINE.md,
// ensuring round-trip fidelity through ParsePipeline.
func (d *DAG) Serialize() []byte {
	var sb strings.Builder

	title := "Research Pipeline"
	if d.Archetype != "" {
		title = fmt.Sprintf("Research Pipeline: %s", d.Archetype)
	}
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	if d.Tier != "" || d.ComplexityScore > 0 {
		sb.WriteString(fmt.Sprintf("> **Tier:** %s · **Complexity Score:** %d\n\n", d.Tier, d.ComplexityScore))
	}

	sb.WriteString("## Pipeline Topology\n\n")
	sb.WriteString("| ID | Topic | Dependencies | Output File |\n")
	sb.WriteString("|---|---|---|---|\n")
	for _, s := range d.Sessions {
		deps := "none"
		if len(s.Dependencies) > 0 {
			deps = strings.Join(s.Dependencies, ", ")
		}
		sTitle := s.Title
		if sTitle == "" {
			sTitle = s.ID
		}
		out := s.OutputFile
		if out == "" {
			out = fmt.Sprintf("sessions/%s.md", s.ID)
		}
		fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n", s.ID, sTitle, deps, out)
	}
	sb.WriteString("\n---\n\n## Session Prompts\n\n")

	for _, s := range d.Sessions {
		sTitle := s.Title
		if sTitle == "" {
			sTitle = s.ID
		}
		fmt.Fprintf(&sb, "### Session %s: %s\n\n", s.ID, sTitle)
		sb.WriteString("| **Field** | **Value** |\n")
		sb.WriteString("|---|---|\n")
		fmt.Fprintf(&sb, "| **ID** | %s |\n", s.ID)
		fmt.Fprintf(&sb, "| **Layer** | %d |\n", s.Layer)
		deps := "none"
		if len(s.Dependencies) > 0 {
			deps = strings.Join(s.Dependencies, ", ")
		}
		fmt.Fprintf(&sb, "| **Dependencies** | %s |\n", deps)
		if s.DecisionRef != "" {
			fmt.Fprintf(&sb, "| **Decision** | %s |\n", s.DecisionRef)
		}
		out := s.OutputFile
		if out == "" {
			out = fmt.Sprintf("sessions/%s.md", s.ID)
		}
		fmt.Fprintf(&sb, "| **Output** | %s |\n", out)
		if s.DoorType != "" {
			fmt.Fprintf(&sb, "| **Door Type** | %s |\n", s.DoorType)
		}
		sb.WriteString("\n")

		if s.Prompt != "" {
			sb.WriteString("```prompt\n")
			sb.WriteString(strings.TrimSpace(s.Prompt))
			sb.WriteString("\n```\n\n")
		} else {
			sb.WriteString("```prompt\n")
			fmt.Fprintf(&sb, "# %s: %s\n\nBRIEF:\nConduct research for %s.\n", s.ID, sTitle, sTitle)
			sb.WriteString("```\n\n")
		}
		sb.WriteString("---\n\n")
	}

	return []byte(sb.String())
}

// ToMermaid generates a Mermaid flowchart representation of the DAG,
// color-coding sessions by their completion status.
func (d *DAG) ToMermaid(completedIDs map[string]bool) string {
	if d == nil || len(d.Sessions) == 0 {
		return "graph TD\n    empty[\"No sessions in pipeline\"]\n"
	}
	if completedIDs == nil {
		completedIDs = make(map[string]bool)
	}

	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Node definitions
	for _, s := range d.Sessions {
		label := s.ID
		if s.Title != "" {
			label = fmt.Sprintf("%s: %s", s.ID, s.Title)
		}
		label = strings.ReplaceAll(label, "\"", "'")
		fmt.Fprintf(&sb, "    %s[\"%s\"]\n", s.ID, label)
	}

	// Directed edges
	for _, s := range d.Sessions {
		for _, dep := range s.Dependencies {
			dep = strings.TrimSpace(dep)
			if dep != "" {
				fmt.Fprintf(&sb, "    %s --> %s\n", dep, s.ID)
			}
		}
	}

	// Status styles
	for _, s := range d.Sessions {
		isDone := completedIDs[s.ID] || completedIDs[strings.ToUpper(s.ID)] || completedIDs[strings.ToLower(s.ID)]
		if isDone {
			fmt.Fprintf(&sb, "    style %s fill:#22c55e,stroke:#16a34a,color:#ffffff\n", s.ID)
		} else if IsSynthesisSession(s.ID) {
			fmt.Fprintf(&sb, "    style %s fill:#3b82f6,stroke:#2563eb,color:#ffffff\n", s.ID)
		} else {
			allDepsDone := true
			for _, dep := range s.Dependencies {
				if !completedIDs[dep] && !completedIDs[strings.ToUpper(dep)] && !completedIDs[strings.ToLower(dep)] {
					allDepsDone = false
					break
				}
			}
			if allDepsDone {
				fmt.Fprintf(&sb, "    style %s fill:#f59e0b,stroke:#d97706,color:#ffffff\n", s.ID)
			} else {
				fmt.Fprintf(&sb, "    style %s fill:#6b7280,stroke:#4b5563,color:#ffffff\n", s.ID)
			}
		}
	}

	return sb.String()
}

// ToStatusTable generates a formatted Markdown status table of all sessions.
func (d *DAG) ToStatusTable(completedIDs map[string]bool) string {
	if d == nil || len(d.Sessions) == 0 {
		return "No sessions in pipeline.\n"
	}
	if completedIDs == nil {
		completedIDs = make(map[string]bool)
	}

	var sb strings.Builder
	sb.WriteString("| Session ID | Title | Layer | Status | Dependencies | Decision |\n")
	sb.WriteString("|---|---|---|---|---|---|\n")

	for _, s := range d.Sessions {
		isDone := completedIDs[s.ID] || completedIDs[strings.ToUpper(s.ID)] || completedIDs[strings.ToLower(s.ID)]
		status := "Blocked"
		if isDone {
			status = "Completed"
		} else {
			allDepsDone := true
			for _, dep := range s.Dependencies {
				if !completedIDs[dep] && !completedIDs[strings.ToUpper(dep)] && !completedIDs[strings.ToLower(dep)] {
					allDepsDone = false
					break
				}
			}
			if allDepsDone {
				status = "Ready (⚡)"
			}
		}

		deps := "none"
		if len(s.Dependencies) > 0 {
			deps = strings.Join(s.Dependencies, ", ")
		}
		dec := s.DecisionRef
		if dec == "" {
			dec = "—"
		}
		sTitle := s.Title
		if sTitle == "" {
			sTitle = s.ID
		}
		fmt.Fprintf(&sb, "| %s | %s | %d | %s | %s | %s |\n", s.ID, sTitle, s.Layer, status, deps, dec)
	}

	return sb.String()
}
