package core

import (
	"fmt"
	"regexp"
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

// SessionByID returns the session with the given ID, or nil if not found.
func (d *DAG) SessionByID(id string) *Session {
	for i := range d.Sessions {
		if d.Sessions[i].ID == id {
			return &d.Sessions[i]
		}
	}
	return nil
}

// ValidateDAG checks the pipeline graph for cycles, unknown dependencies, and duplicates.
func (d *DAG) ValidateDAG() error {
	ids := make(map[string]bool)
	for _, s := range d.Sessions {
		if s.ID == "" {
			return fmt.Errorf("session with empty ID found")
		}
		if ids[s.ID] {
			return fmt.Errorf("duplicate session ID: %s", s.ID)
		}
		ids[s.ID] = true
	}

	// Warn about non-canonical IDs (no hyphen) — these will fail filename-based matching
	for _, s := range d.Sessions {
		if !strings.Contains(s.ID, "-") {
			return fmt.Errorf("session ID %q lacks a hyphen — use canonical form like 'T1-01' or 'SYN-01' for reliable filename matching", s.ID)
		}
	}

	// Check for dangling dependencies
	for _, s := range d.Sessions {
		for _, dep := range s.Dependencies {
			if !ids[dep] {
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
		// Skip already completed
		if completedIDs[s.ID] {
			continue
		}

		// Check if all dependencies are met
		allDeps := true
		for _, dep := range s.Dependencies {
			if !completedIDs[dep] {
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

// Patterns for parsing the pipeline format
var (
	// Matches session headers like: #### T1-01: Graph Persistence Landscape, ### Session T1-01: Spike, #### D-001-S1 — Spike
	sessionHeaderRe = regexp.MustCompile(`^#{2,4}\s+(?:Session\s+)?([A-Za-z0-9]+(?:-[A-Za-z0-9]+)+):?\s*(.*)`)

	// Matches metadata table rows like: | **ID** | T1-01 |
	metaFieldRe = regexp.MustCompile(`\|\s*\*\*([^*]+)\*\*\s*\|\s*(.+?)\s*\|`)

	// Matches the prompt code block
	promptStartRe = regexp.MustCompile("^```prompt")
	promptEndRe   = regexp.MustCompile("^```$")
)

// ParsePipeline parses a RESEARCH-PIPELINE.md file into a DAG structure.
func ParsePipeline(data []byte) (*DAG, error) {
	dag := &DAG{}
	lines := strings.Split(string(data), "\n")

	var currentSession *Session
	inPrompt := false
	var promptLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Check for session header
		if matches := sessionHeaderRe.FindStringSubmatch(trimmed); matches != nil {
			// Save previous session if any
			if currentSession != nil {
				if inPrompt {
					currentSession.Prompt = strings.Join(promptLines, "\n")
					inPrompt = false
					promptLines = nil
				}
				dag.Sessions = append(dag.Sessions, *currentSession)
			}

			currentSession = &Session{
				ID:    strings.TrimSpace(matches[1]),
				Title: strings.TrimSpace(strings.TrimPrefix(matches[2], "—")),
			}
			continue
		}

		// Collect prompt block
		if inPrompt {
			if promptEndRe.MatchString(trimmed) {
				if currentSession != nil {
					currentSession.Prompt = strings.Join(promptLines, "\n")
				}
				inPrompt = false
				promptLines = nil
			} else {
				promptLines = append(promptLines, line)
			}
			continue
		}

		if promptStartRe.MatchString(trimmed) {
			inPrompt = true
			promptLines = nil
			continue
		}

		// Parse metadata table rows
		if currentSession != nil && strings.Contains(line, "**") && strings.Contains(line, "|") {
			if matches := metaFieldRe.FindStringSubmatch(line); matches != nil {
				field := strings.TrimSpace(matches[1])
				value := strings.TrimSpace(matches[2])

				switch strings.ToLower(field) {
				case "id":
					currentSession.ID = strings.TrimSpace(value)
				case "layer":
					fmt.Sscanf(value, "%d", &currentSession.Layer)
				case "door type":
					currentSession.DoorType = value
				case "decision":
					currentSession.DecisionRef = value
				case "dependencies":
					currentSession.Dependencies = parseDependencies(value)
				case "output file":
					currentSession.OutputFile = strings.Trim(value, "`")
				}
			}
		}

		// Parse complexity score
		if strings.Contains(trimmed, "Complexity Score") && strings.Contains(trimmed, "→") {
			// Extract score and tier
			if idx := strings.Index(trimmed, "→"); idx > 0 {
				dag.Tier = strings.TrimSpace(trimmed[idx+len("→"):])
				beforeArrow := trimmed[:idx]
				re := regexp.MustCompile(`(\d+)(?:\s*/\s*\d+)?`)
				if matches := re.FindStringSubmatch(beforeArrow); len(matches) > 1 {
					fmt.Sscanf(matches[1], "%d", &dag.ComplexityScore)
				}
			}
		}

		// Parse archetype
		if strings.Contains(trimmed, "Primary Archetype") {
			if idx := strings.LastIndex(trimmed, "**"); idx > 0 {
				// Try to extract the value after the last **
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					dag.Archetype = strings.TrimSpace(strings.Trim(parts[1], "* "))
				}
			}
		}
	}

	// Save the last session
	if currentSession != nil {
		if inPrompt {
			currentSession.Prompt = strings.Join(promptLines, "\n")
		}
		dag.Sessions = append(dag.Sessions, *currentSession)
	}

	if len(dag.Sessions) == 0 {
		return dag, fmt.Errorf("no sessions found in pipeline")
	}

	return dag, nil
}

// Session IDs are used exactly as they appear in the pipeline.
// No normalization is performed. IDs must be consistent between
// the header, metadata table, and dependency lists.
// Canonical format uses a hyphen (e.g., "T1-01", "SYN-01", "D-001-S1").
// ValidateDAG warns if IDs lack hyphens.

// parseDependencies parses a dependency list from the metadata table.
// Handles formats like: "None (parallel)", "T1-01", "T1-01, T1-02", "T1-01 (soft)"
func parseDependencies(value string) []string {
	lower := strings.ToLower(value)
	if lower == "none" || strings.Contains(lower, "none") || lower == "—" || lower == "-" || lower == "n/a" {
		return nil
	}

	parts := strings.Split(value, ",")
	var deps []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, "("); idx > 0 {
			p = strings.TrimSpace(p[:idx])
		}
		if p != "" && !strings.EqualFold(p, "none") && !strings.EqualFold(p, "n/a") {
			deps = append(deps, p)
		}
	}
	return deps
}
