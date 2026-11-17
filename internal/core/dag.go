package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// The plan format (HIGH-07). Every plan, research/RESEARCH-PIPELINE.md and research/D-NNN-plan.md alike, carries
// exactly one fenced ```yaml vivechak-plan``` block, the only source of its sessions and their dependencies:
//
//	```yaml vivechak-plan
//	plan: PROJECT          # the project pipeline; a decision plan names its decision: plan: D-001
//	sessions:
//	  - {id: T1-01, title: Datastore, layer: 1, door: one-way, informs: [D-001], depends: []}
//	  - {id: T1-02, title: Queue, layer: 1, door: two-way, depends: [T1-01]}
//	  - {id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}
//	decisions:             # project plans list the decisions their sessions inform
//	  - {id: D-001, title: Primary datastore, door: one-way}
//	```
//
// Decision plans add role: L|C|F to each session. depends: [ALL] means every non-synthesis session listed before it.
// A session is the synthesis session only when it says kind: synthesis, and a project plan has exactly one. Each
// session's prompt is the one prompt fence under its "### <id>: <title>" heading in the plan's "## Sessions" section.
// Tables and other prose are for people and are never parsed.

// PlanMachineBlock is a plan's ```yaml vivechak-plan``` block.
type PlanMachineBlock struct {
	Plan      string             `yaml:"plan"`
	Sessions  []PlanSessionItem  `yaml:"sessions"`
	Decisions []PlanDecisionItem `yaml:"decisions"`
}

// PlanSessionItem is one session entry of a PlanMachineBlock.
type PlanSessionItem struct {
	ID      string   `yaml:"id"`
	Title   string   `yaml:"title"`
	Layer   int      `yaml:"layer"`
	Door    string   `yaml:"door"`
	Informs []string `yaml:"informs"`
	Depends []string `yaml:"depends"`
	Kind    string   `yaml:"kind"`
	Role    string   `yaml:"role"`
}

// PlanDecisionItem is one decision a project plan's block lists.
type PlanDecisionItem struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
	Door  string `yaml:"door"`
}

// SynthesisKind is the kind of a plan's synthesis session (kind: synthesis), whose output is research/FAD.md.
const SynthesisKind = "synthesis"

// Session represents a single research session in the pipeline DAG.
type Session struct {
	// ID is the session identifier (e.g., "T1-01", "D-001-S1", "SYN-01").
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

	// Kind is "synthesis" for the plan's synthesis session and empty for every other session.
	Kind string `json:"kind,omitempty"`

	// Role is a decision plan session's role: L (landscape), C (comparison) or F (falsify).
	Role string `json:"role,omitempty"`

	// declaredDeps are the dependencies the plan names by id; dependsOnAll records depends: [ALL].
	declaredDeps []string
	dependsOnAll bool
}

// IsSynthesis reports whether the plan marks this session kind: synthesis.
func (s Session) IsSynthesis() bool {
	return strings.EqualFold(strings.TrimSpace(s.Kind), SynthesisKind)
}

// DAG represents the research pipeline as a directed acyclic graph of sessions.
type DAG struct {
	// Sessions is the ordered list of sessions in the pipeline.
	Sessions []Session `json:"sessions"`

	// Plan is the block's plan field: PROJECT (or DAG) for the project pipeline, the decision id for a decision plan.
	Plan string `json:"plan,omitempty"`
}

// SessionByID returns the session with the given ID, or nil if not found (case-insensitive).
func (d *DAG) SessionByID(id string) *Session {
	if d == nil {
		return nil
	}
	id = strings.TrimSpace(id)
	for i := range d.Sessions {
		if strings.EqualFold(d.Sessions[i].ID, id) {
			return &d.Sessions[i]
		}
	}
	return nil
}

// IsSynthesisID reports whether the plan lists id as its kind: synthesis session.
func (d *DAG) IsSynthesisID(id string) bool {
	s := d.SessionByID(id)
	return s != nil && s.IsSynthesis()
}

// SynthesisSessions returns the sessions the plan marks kind: synthesis.
func (d *DAG) SynthesisSessions() []Session {
	if d == nil {
		return nil
	}
	var out []Session
	for _, s := range d.Sessions {
		if s.IsSynthesis() {
			out = append(out, s)
		}
	}
	return out
}

// IsProject reports whether the plan is the project pipeline (plan: PROJECT or DAG) rather than a decision plan.
func (d *DAG) IsProject() bool {
	return d != nil && isProjectPlanValue(d.Plan)
}

func isProjectPlanValue(v string) bool {
	v = strings.ToUpper(strings.TrimSpace(v))
	return v == "PROJECT" || v == "DAG"
}

func isDecisionPlanValue(v string) bool {
	v = strings.ToUpper(strings.TrimSpace(v))
	return strings.HasPrefix(v, "D-") && len(v) > 2
}

// CompletedSessions returns the plan's completed sessions (see ScanCompletedSessions). A non-draft research/FAD.md
// completes the plan's kind: synthesis session, whatever id the FAD file carries.
func (d *DAG) CompletedSessions(ws WorkspaceFS) (map[string]bool, error) {
	var ids []string
	if d != nil {
		for _, s := range d.Sessions {
			ids = append(ids, s.ID)
		}
	}
	completed, err := ScanCompletedSessions(ws, ids...)
	if err != nil || d == nil {
		return completed, err
	}
	if completed["FAD"] {
		for _, s := range d.SynthesisSessions() {
			completed[s.ID] = true
			completed[strings.ToUpper(s.ID)] = true
		}
	}
	return completed, nil
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
			if dep == "__ALL_PRIOR__" {
				continue
			}
			if !ids[dep] && !ids[strings.ToUpper(dep)] {
				return fmt.Errorf("session %s depends on nonexistent session %s", s.ID, dep)
			}
		}
	}

	// Cycle detection using DFS (0=unvisited, 1=visiting, 2=visited)
	visited := make(map[string]int)
	var dfs func(id string, path []string) error
	dfs = func(id string, path []string) error {
		upperID := strings.ToUpper(id)
		visited[upperID] = 1 // visiting (gray)
		path = append(path, id)

		session := d.SessionByID(id)
		if session != nil {
			for _, dep := range session.Dependencies {
				if dep == "__ALL_PRIOR__" {
					continue
				}
				upperDep := strings.ToUpper(dep)
				if visited[upperDep] == 1 {
					return fmt.Errorf("cycle detected in pipeline DAG: %s -> %s", strings.Join(path, " -> "), dep)
				}
				if visited[upperDep] == 0 {
					if err := dfs(dep, path); err != nil {
						return err
					}
				}
			}
		}

		visited[upperID] = 2 // visited (black)
		return nil
	}

	for _, s := range d.Sessions {
		if visited[strings.ToUpper(s.ID)] == 0 {
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
			// Terminal Synthesis Guard: the kind: synthesis session is ONLY actionable
			// when ALL non-synthesis sessions in the DAG have been completed.
			if s.IsSynthesis() {
				allNonSynthesisDone := true
				for _, other := range d.Sessions {
					if !other.IsSynthesis() {
						if !completedIDs[other.ID] && !completedIDs[strings.ToUpper(other.ID)] && !completedIDs[strings.ToLower(other.ID)] {
							allNonSynthesisDone = false
							break
						}
					}
				}
				if !allNonSynthesisDone {
					continue // Hold back synthesis session until all research sessions complete
				}
			}
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

// Plan validation codes. Each names a reason a plan can't be used, so saving such a plan fails.
const (
	CodePlanMissingBlock      = "V-PLAN-MISSING-BLOCK"
	CodePlanInvalidBlock      = "V-PLAN-INVALID-BLOCK"
	CodePlanNoSessions        = "V-PLAN-NO-SESSIONS"
	CodePlanMissingPrompt     = "V-PLAN-MISSING-PROMPT"
	CodePlanDuplicatePrompt   = "V-PLAN-DUPLICATE-PROMPT"
	CodePlanUnlistedSession   = "V-PLAN-UNLISTED-SESSION"
	CodePlanUnknownDependency = "V-PLAN-UNKNOWN-DEPENDENCY"
	CodePlanCycle             = "V-PLAN-CYCLE"
	CodePlanSynthesis         = "V-PLAN-SYNTHESIS"
)

// PlanIssue is one reason a plan can't be used: a V-PLAN-* code, what is wrong, and the fix in one sentence.
type PlanIssue struct {
	Code    string
	Message string
	Fix     string
}

func (p *PlanIssue) Error() string { return p.Code + ": " + p.Message }

// PlanFix returns the fix sentence of a plan error from ParsePipeline or ReplanInPlace, or "" for any other error.
func PlanFix(err error) string {
	var pi *PlanIssue
	if errors.As(err, &pi) {
		return pi.Fix
	}
	return ""
}

const fixMissingBlock = "Add one ```yaml vivechak-plan``` block that lists every session " +
	"(plan: PROJECT or the decision id; sessions with id, title and depends) and save the plan again."

// fenceRange is a fenced block in a plan, as line indexes: the opening fence line and the closing fence line
// (len(lines) when the fence is never closed).
type fenceRange struct {
	open  int
	close int
}

// planSection is a "### " section inside the plan's "## Sessions" section.
type planSection struct {
	heading int    // the heading line
	end     int    // the first line after the section
	token   string // the heading's first word: the session id it names
	title   string // the heading text after "<id>:"
	prompts []fenceRange
}

// planHeading is a "#### " (or deeper) heading inside "## Sessions".
type planHeading struct {
	line int
	text string
}

// planLayout is where a plan's block, "## Sessions" section and session prompts are, line by line.
type planLayout struct {
	lines       []string
	blocks      []fenceRange // ```yaml vivechak-plan``` fences
	hasSessions bool         // the plan has a "## Sessions" heading
	sessionsEnd int          // the first line after the last "## Sessions" section
	sections    []planSection
	subheadings []planHeading
}

// promptStartRe matches a prompt fence's opening line.
var promptStartRe = regexp.MustCompile("^`{3,}prompt")

// promptFenceResult represents the outcome of checking a fence delimiter inside a prompt block.
type promptFenceResult struct {
	isClosingFence bool
	inInnerBlock   bool
}

// checkPromptFence determines whether a line inside a prompt block is a closing fence.
// It handles nested code fences (e.g. inner ```json or ```go blocks) and boundary heuristics.
func checkPromptFence(trimmed string, outerFenceLen int, inInnerBlock bool, lines []string, currentLine int) promptFenceResult {
	numBackticks := 0
	for j := 0; j < len(trimmed) && trimmed[j] == '`'; j++ {
		numBackticks++
	}

	if outerFenceLen >= 4 {
		if numBackticks >= outerFenceLen && len(strings.TrimSpace(trimmed[numBackticks:])) == 0 {
			return promptFenceResult{isClosingFence: true, inInnerBlock: inInnerBlock}
		}
		return promptFenceResult{isClosingFence: false, inInnerBlock: inInnerBlock}
	}

	if numBackticks >= 3 {
		afterTicks := strings.TrimSpace(trimmed[numBackticks:])
		if afterTicks != "" {
			return promptFenceResult{isClosingFence: false, inInnerBlock: !inInnerBlock}
		}
		if inInnerBlock {
			return promptFenceResult{isClosingFence: false, inInnerBlock: false}
		}
		nextIsBoundary := true
		for k := currentLine + 1; k < len(lines); k++ {
			kTrim := strings.TrimSpace(lines[k])
			if kTrim == "" {
				continue
			}
			// A heading, a rule, or another tagged fence (a prompt, the vivechak-plan block) after a bare fence
			// means the bare fence closed the prompt.
			_, _, info, isFence := fenceOpening(kTrim)
			if strings.HasPrefix(kTrim, "---") ||
				strings.HasPrefix(kTrim, "###") ||
				strings.HasPrefix(kTrim, "##") ||
				promptStartRe.MatchString(kTrim) ||
				(isFence && info != "") {
				nextIsBoundary = true
			} else {
				nextIsBoundary = false
			}
			break
		}
		if nextIsBoundary {
			return promptFenceResult{isClosingFence: true, inInnerBlock: false}
		}
		return promptFenceResult{isClosingFence: false, inInnerBlock: true}
	}

	return promptFenceResult{isClosingFence: false, inInnerBlock: inInnerBlock}
}

func layoutPlan(data []byte) *planLayout {
	lines := strings.Split(string(data), "\n")
	lay := &planLayout{lines: lines}
	var (
		fenceChar    byte
		fenceLen     int
		fenceOpen    = -1
		fenceIsBlock bool
		promptOpen   = -1
		promptLen    int
		innerBlock   bool
		inSessions   bool
		cur          = -1
	)
	endSection := func(i int) {
		if cur >= 0 {
			lay.sections[cur].end = i
			cur = -1
		}
	}
	for i, raw := range lines {
		t := strings.TrimSpace(raw)
		if promptOpen >= 0 {
			res := checkPromptFence(t, promptLen, innerBlock, lines, i)
			innerBlock = res.inInnerBlock
			if res.isClosingFence {
				lay.sections[cur].prompts = append(lay.sections[cur].prompts, fenceRange{promptOpen, i})
				promptOpen = -1
			}
			continue
		}
		if fenceOpen >= 0 {
			if isFenceClose(t, fenceChar, fenceLen) {
				if fenceIsBlock {
					lay.blocks = append(lay.blocks, fenceRange{fenceOpen, i})
				}
				fenceOpen = -1
			}
			continue
		}
		if cur >= 0 && promptStartRe.MatchString(t) {
			promptOpen, promptLen, innerBlock = i, leadingRun(t, '`'), false
			continue
		}
		if c, n, info, ok := fenceOpening(t); ok {
			fenceChar, fenceLen, fenceOpen, fenceIsBlock = c, n, i, isPlanBlockInfo(info)
			continue
		}
		switch level, text := atxHeading(t); {
		case level == 1 || level == 2:
			endSection(i)
			if inSessions {
				lay.sessionsEnd = i
			}
			inSessions = level == 2 && strings.EqualFold(text, "Sessions")
			lay.hasSessions = lay.hasSessions || inSessions
		case level == 3 && inSessions:
			endSection(i)
			token, title := splitSessionHeading(text)
			lay.sections = append(lay.sections, planSection{heading: i, end: len(lines), token: token, title: title})
			cur = len(lay.sections) - 1
		case level > 3 && inSessions:
			lay.subheadings = append(lay.subheadings, planHeading{line: i, text: text})
		}
	}
	if promptOpen >= 0 {
		lay.sections[cur].prompts = append(lay.sections[cur].prompts, fenceRange{promptOpen, len(lines)})
	}
	if fenceOpen >= 0 && fenceIsBlock {
		lay.blocks = append(lay.blocks, fenceRange{fenceOpen, len(lines)})
	}
	endSection(len(lines))
	if inSessions {
		lay.sessionsEnd = len(lines)
	}
	return lay
}

func leadingRun(t string, c byte) int {
	n := 0
	for n < len(t) && t[n] == c {
		n++
	}
	return n
}

// fenceOpening reports whether t opens a fenced code block, with its fence character, length and info string.
func fenceOpening(t string) (byte, int, string, bool) {
	if len(t) < 3 || (t[0] != '`' && t[0] != '~') {
		return 0, 0, "", false
	}
	c := t[0]
	n := leadingRun(t, c)
	if n < 3 {
		return 0, 0, "", false
	}
	info := strings.TrimSpace(t[n:])
	if c == '`' && strings.Contains(info, "`") {
		return 0, 0, "", false
	}
	return c, n, info, true
}

func isFenceClose(t string, c byte, n int) bool {
	m := leadingRun(t, c)
	return m >= n && strings.TrimSpace(t[m:]) == ""
}

func isPlanBlockInfo(info string) bool {
	f := strings.Fields(strings.ToLower(info))
	return (len(f) == 1 && f[0] == "vivechak-plan") || (len(f) == 2 && f[0] == "yaml" && f[1] == "vivechak-plan")
}

// atxHeading returns the level and text of a Markdown heading line, or level 0.
func atxHeading(t string) (int, string) {
	n := leadingRun(t, '#')
	if n == 0 || n > 6 || (n < len(t) && t[n] != ' ' && t[n] != '\t') {
		return 0, ""
	}
	return n, strings.TrimSpace(t[n:])
}

// splitSessionHeading splits "T1-01: Datastore" into the id it names and its title.
func splitSessionHeading(text string) (string, string) {
	end := strings.IndexFunc(text, func(r rune) bool { return r == ':' || unicode.IsSpace(r) })
	if end < 0 {
		return text, ""
	}
	rest := strings.TrimSpace(text[end:])
	rest = strings.TrimSpace(strings.TrimLeft(rest, ":—–-"))
	return text[:end], rest
}

func (lay *planLayout) fenceContent(f fenceRange) string {
	end := f.close
	if end > len(lay.lines) {
		end = len(lay.lines)
	}
	var out []string
	for _, l := range lay.lines[f.open+1 : end] {
		out = append(out, strings.TrimRight(l, "\r"))
	}
	return strings.Join(out, "\n")
}

// decodeBlock reads the plan's one vivechak-plan block, checking its shape key by key.
func (lay *planLayout) decodeBlock() (*PlanMachineBlock, []PlanIssue) {
	switch {
	case len(lay.blocks) == 0:
		return nil, []PlanIssue{{CodePlanMissingBlock, "the plan has no ```yaml vivechak-plan``` block", fixMissingBlock}}
	case len(lay.blocks) > 1:
		return nil, []PlanIssue{{CodePlanInvalidBlock, fmt.Sprintf("the plan has %d ```yaml vivechak-plan``` blocks", len(lay.blocks)),
			"Keep exactly one vivechak-plan block that lists every session."}}
	case lay.blocks[0].close >= len(lay.lines):
		return nil, []PlanIssue{{CodePlanInvalidBlock, "the ```yaml vivechak-plan``` block is never closed",
			"End the block with a ``` line after its last entry."}}
	}
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(lay.fenceContent(lay.blocks[0])), &root); err != nil {
		return nil, []PlanIssue{{CodePlanInvalidBlock, "the vivechak-plan block is not valid YAML: " + err.Error(),
			"Fix the block's YAML; write each session as one entry like - {id: T1-01, title: Datastore, depends: []}."}}
	}
	if issues := checkBlockShape(&root); len(issues) > 0 {
		return nil, issues
	}
	var block PlanMachineBlock
	if err := root.Decode(&block); err != nil {
		return nil, []PlanIssue{{CodePlanInvalidBlock, "the vivechak-plan block can't be read: " + err.Error(),
			"Fix the block so every value has the type the plan format gives it."}}
	}
	return &block, nil
}

var (
	planTopKeys      = []string{"plan", "sessions", "decisions"}
	planSessionKeys  = []string{"id", "title", "layer", "door", "informs", "depends", "kind", "role"}
	planDecisionKeys = []string{"id", "title", "door"}
	// planKeyHints names the block key to use for keys planners commonly write instead.
	planKeyHints = map[string]string{
		"dependencies": "depends", "depends_on": "depends", "deps": "depends", "door_type": "door",
		"decision": "informs", "decision_ref": "informs", "decision_id": "informs", "type": "kind",
	}
)

func resolveAlias(n *yaml.Node) *yaml.Node {
	for n != nil && n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}

func isNullNode(n *yaml.Node) bool {
	return n == nil || (n.Kind == yaml.ScalarNode && n.Tag == "!!null")
}

func scalarText(n *yaml.Node) (string, bool) {
	if n == nil || n.Kind != yaml.ScalarNode {
		return "", false
	}
	if n.Tag == "!!null" {
		return "", true
	}
	return strings.TrimSpace(n.Value), true
}

func invalidBlock(msg, fix string) PlanIssue {
	return PlanIssue{Code: CodePlanInvalidBlock, Message: msg, Fix: fix}
}

func unknownKeyIssue(where, key string, allowed []string, hints map[string]string) PlanIssue {
	fix := fmt.Sprintf("Use only these keys there: %s.", strings.Join(allowed, ", "))
	if use, ok := hints[strings.ToLower(key)]; ok {
		fix = fmt.Sprintf("Write %s instead of %s.", use, key)
	}
	return invalidBlock(fmt.Sprintf("%s has an unknown key %q", where, key), fix)
}

// checkBlockShape checks the block's keys and value types against the plan format, naming each problem.
func checkBlockShape(root *yaml.Node) []PlanIssue {
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return []PlanIssue{invalidBlock("the vivechak-plan block is empty", "Write plan: and a sessions: list in the block.")}
	}
	doc := resolveAlias(root.Content[0])
	if doc.Kind != yaml.MappingNode {
		return []PlanIssue{invalidBlock("the vivechak-plan block is not a mapping with plan: and sessions:",
			"Start the block with plan: PROJECT (or the decision id) followed by a sessions: list.")}
	}
	var issues []PlanIssue
	var planNode, sessionsNode, decisionsNode *yaml.Node
	for i := 0; i+1 < len(doc.Content); i += 2 {
		k, v := doc.Content[i], resolveAlias(doc.Content[i+1])
		switch k.Value {
		case "plan":
			planNode = v
		case "sessions":
			sessionsNode = v
		case "decisions":
			decisionsNode = v
		default:
			issues = append(issues, unknownKeyIssue("the block", k.Value, planTopKeys, nil))
		}
	}
	switch v, ok := scalarText(planNode); {
	case !ok || v == "":
		issues = append(issues, invalidBlock("the block has no plan: value",
			"Add plan: PROJECT for the project pipeline, or plan: D-NNN (the decision id) for a decision plan."))
	case !isProjectPlanValue(v) && !isDecisionPlanValue(v):
		issues = append(issues, invalidBlock(fmt.Sprintf("plan: %q is neither PROJECT nor a decision id", v),
			"Set plan: PROJECT for the project pipeline, or plan: D-NNN (the decision id) for a decision plan."))
	}
	switch {
	case isNullNode(sessionsNode) || (sessionsNode.Kind == yaml.SequenceNode && len(sessionsNode.Content) == 0):
		issues = append(issues, PlanIssue{CodePlanNoSessions, "the vivechak-plan block lists no sessions",
			"List every session under sessions:, one entry each, e.g. - {id: T1-01, title: Datastore, depends: []}."})
	case sessionsNode.Kind != yaml.SequenceNode:
		issues = append(issues, invalidBlock("sessions: is not a list",
			"Write sessions: as a list with one - {id: ..., title: ..., depends: [...]} entry per session."))
	default:
		for n, item := range sessionsNode.Content {
			issues = append(issues, checkSessionEntry(n+1, resolveAlias(item))...)
		}
	}
	if !isNullNode(decisionsNode) {
		if decisionsNode.Kind != yaml.SequenceNode {
			issues = append(issues, invalidBlock("decisions: is not a list",
				"Write decisions: as a list with one - {id: D-001, title: ..., door: one-way} entry per decision."))
		} else {
			for n, item := range decisionsNode.Content {
				issues = append(issues, checkDecisionEntry(n+1, resolveAlias(item))...)
			}
		}
	}
	return issues
}

func mappingValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return resolveAlias(m.Content[i+1])
		}
	}
	return nil
}

func isScalarList(n *yaml.Node) bool {
	if isNullNode(n) {
		return true
	}
	if n.Kind != yaml.SequenceNode {
		return false
	}
	for _, c := range n.Content {
		if c = resolveAlias(c); c.Kind != yaml.ScalarNode {
			return false
		}
	}
	return true
}

func checkSessionEntry(n int, item *yaml.Node) []PlanIssue {
	if item.Kind != yaml.MappingNode {
		return []PlanIssue{invalidBlock(fmt.Sprintf("sessions entry %d is not a mapping", n),
			"Write each session as - {id: T1-01, title: ..., depends: [...]}.")}
	}
	name := fmt.Sprintf("sessions entry %d", n)
	id, ok := scalarText(mappingValue(item, "id"))
	if ok && id != "" {
		name = "session " + id
	}
	var issues []PlanIssue
	if !ok || id == "" {
		issues = append(issues, invalidBlock(fmt.Sprintf("sessions entry %d has no id", n), "Give every session an id, e.g. id: T1-01."))
	}
	for i := 0; i+1 < len(item.Content); i += 2 {
		key, v := item.Content[i].Value, resolveAlias(item.Content[i+1])
		text, isScalar := scalarText(v)
		switch key {
		case "id":
		case "title":
			if !isScalar {
				issues = append(issues, invalidBlock(name+" has a title that is not text", "Write the title as plain text, e.g. title: Datastore."))
			}
		case "layer":
			if !isNullNode(v) && (v.Kind != yaml.ScalarNode || v.Tag != "!!int") {
				issues = append(issues, invalidBlock(name+" has a layer that is not a whole number", "Write the layer as a number, e.g. layer: 1."))
			}
		case "door":
			if !isScalar || (text != "" && !strings.EqualFold(text, "one-way") && !strings.EqualFold(text, "two-way")) {
				issues = append(issues, invalidBlock(fmt.Sprintf("%s has door %q", name, text), "Write door: one-way or door: two-way."))
			}
		case "informs", "depends":
			if !isScalarList(v) {
				issues = append(issues, invalidBlock(fmt.Sprintf("%s has a %s value that is not a list of ids", name, key),
					fmt.Sprintf("Write %s as a list, e.g. %s: [%s] or %s: [].", key, key, map[string]string{"informs": "D-001", "depends": "T1-01"}[key], key)))
			}
		case "kind":
			if !isScalar || (text != "" && !strings.EqualFold(text, SynthesisKind)) {
				issues = append(issues, invalidBlock(fmt.Sprintf("%s has kind %q", name, text),
					"Write kind: synthesis on the plan's synthesis session and leave kind out everywhere else."))
			}
		case "role":
			if !isScalar || (text != "" && !strings.EqualFold(text, "L") && !strings.EqualFold(text, "C") && !strings.EqualFold(text, "F")) {
				issues = append(issues, invalidBlock(fmt.Sprintf("%s has role %q", name, text),
					"Write role: L (landscape), role: C (comparison) or role: F (falsify)."))
			}
		default:
			issues = append(issues, unknownKeyIssue(name, key, planSessionKeys, planKeyHints))
		}
	}
	return issues
}

func checkDecisionEntry(n int, item *yaml.Node) []PlanIssue {
	if item.Kind != yaml.MappingNode {
		return []PlanIssue{invalidBlock(fmt.Sprintf("decisions entry %d is not a mapping", n),
			"Write each decision as - {id: D-001, title: ..., door: one-way}.")}
	}
	name := fmt.Sprintf("decisions entry %d", n)
	if id, ok := scalarText(mappingValue(item, "id")); ok && id != "" {
		name = "decision " + id
	}
	var issues []PlanIssue
	for i := 0; i+1 < len(item.Content); i += 2 {
		key, v := item.Content[i].Value, resolveAlias(item.Content[i+1])
		text, isScalar := scalarText(v)
		switch key {
		case "id", "title":
			if !isScalar {
				issues = append(issues, invalidBlock(fmt.Sprintf("%s has a %s that is not text", name, key), "Write it as plain text."))
			}
		case "door":
			if !isScalar || (text != "" && !strings.EqualFold(text, "one-way") && !strings.EqualFold(text, "two-way")) {
				issues = append(issues, invalidBlock(fmt.Sprintf("%s has door %q", name, text), "Write door: one-way or door: two-way."))
			}
		default:
			issues = append(issues, unknownKeyIssue(name, key, planDecisionKeys, map[string]string{"door_type": "door"}))
		}
	}
	return issues
}

// ParsePlanMachineBlock returns the plan's ```yaml vivechak-plan``` block: (nil, nil) when the plan has none and an
// error when the block can't be read. It does not check the block against the plan format; ParsePipeline does.
func ParsePlanMachineBlock(data []byte) (*PlanMachineBlock, error) {
	lay := layoutPlan(data)
	if len(lay.blocks) == 0 {
		return nil, nil
	}
	if lay.blocks[0].close >= len(lay.lines) {
		return nil, fmt.Errorf("the vivechak-plan block is never closed")
	}
	var block PlanMachineBlock
	if err := yaml.Unmarshal([]byte(lay.fenceContent(lay.blocks[0])), &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// ParsePipeline parses a plan (the project pipeline or a decision plan) into its DAG. It fails with the first
// V-PLAN-* problem (a *PlanIssue; PlanFix gives its fix) when the plan can't be used: no block or an invalid one, no
// sessions, a session without exactly one prompt, an unknown dependency, or a cycle. How many synthesis sessions a
// project plan has is checked by ValidatePlan and the gate.
func ParsePipeline(data []byte) (*DAG, error) {
	dag, issues := parsePlan(data)
	if len(issues) > 0 {
		return nil, &issues[0]
	}
	return dag, nil
}

// parsePlan builds the plan's DAG from its block and collects every V-PLAN-* problem. The DAG is nil when the
// block is missing or invalid.
func parsePlan(data []byte) (*DAG, []PlanIssue) {
	lay := layoutPlan(data)
	block, issues := lay.decodeBlock()
	if block == nil {
		return nil, issues
	}

	dag := &DAG{Plan: strings.TrimSpace(block.Plan)}
	index := map[string]int{}
	var rawDeps [][]string
	for _, item := range block.Sessions {
		id := strings.TrimSpace(item.ID)
		switch _, dup := index[strings.ToUpper(id)]; {
		case !validSessionIDRe.MatchString(id):
			issues = append(issues, invalidBlock(fmt.Sprintf("session id %q is not a valid id", id),
				"Use letters and digits joined by single hyphens or underscores, e.g. T1-01."))
			continue
		case dup:
			issues = append(issues, invalidBlock(fmt.Sprintf("session %s is listed twice in the block", id),
				"Give each session exactly one entry in the block."))
			continue
		}
		s := Session{
			ID:       id,
			Title:    strings.TrimSpace(item.Title),
			Layer:    item.Layer,
			DoorType: strings.ToLower(strings.TrimSpace(item.Door)),
			Role:     strings.ToUpper(strings.TrimSpace(item.Role)),
		}
		if strings.EqualFold(strings.TrimSpace(item.Kind), SynthesisKind) {
			s.Kind = SynthesisKind
		}
		for _, inf := range item.Informs {
			if inf = strings.TrimSpace(inf); inf != "" {
				s.DecisionRef = inf
				break
			}
		}
		if s.IsSynthesis() {
			s.OutputFile = FADFile
		} else {
			s.OutputFile = filepath.ToSlash(filepath.Join(SessionsDir, id+".md"))
		}
		index[strings.ToUpper(id)] = len(dag.Sessions)
		dag.Sessions = append(dag.Sessions, s)
		rawDeps = append(rawDeps, item.Depends)
	}

	// Dependencies: ids from the block; ALL is every non-synthesis session listed before this one.
	for i := range dag.Sessions {
		s := &dag.Sessions[i]
		seen := map[string]bool{}
		add := func(id string) {
			if !seen[strings.ToUpper(id)] {
				seen[strings.ToUpper(id)] = true
				s.Dependencies = append(s.Dependencies, id)
			}
		}
		for _, dep := range rawDeps[i] {
			dep = strings.TrimSpace(dep)
			if dep == "" {
				continue
			}
			if isAllPriorSessions(dep) {
				s.dependsOnAll = true
				for _, prior := range dag.Sessions[:i] {
					if !prior.IsSynthesis() {
						add(prior.ID)
					}
				}
				continue
			}
			s.declaredDeps = append(s.declaredDeps, dep)
			k, ok := index[strings.ToUpper(dep)]
			switch {
			case !ok:
				issues = append(issues, PlanIssue{CodePlanUnknownDependency,
					fmt.Sprintf("session %s depends on %s, which the block does not list", s.ID, dep),
					"List only session ids from the block in depends, or ALL for every session listed before it."})
			case k == i:
				issues = append(issues, PlanIssue{CodePlanCycle, fmt.Sprintf("session %s depends on itself", s.ID),
					fmt.Sprintf("Remove %s from its own depends list.", s.ID)})
			case dag.Sessions[k].IsSynthesis() && !s.IsSynthesis():
				issues = append(issues, PlanIssue{CodePlanCycle,
					fmt.Sprintf("session %s depends on the synthesis session %s, which runs after every other session", s.ID, dep),
					fmt.Sprintf("Remove %s from %s's depends list.", dep, s.ID)})
			default:
				add(dag.Sessions[k].ID)
			}
		}
	}
	if cycle := findCycle(dag); cycle != nil {
		issues = append(issues, PlanIssue{CodePlanCycle,
			"sessions depend on each other in a cycle: " + strings.Join(cycle, " -> "),
			"Remove one of these dependencies so the sessions can run in order."})
	}

	issues = append(issues, attachPrompts(lay, dag)...)
	return dag, issues
}

// findCycle returns a dependency cycle as a path of session ids (first id repeated at the end), or nil.
func findCycle(dag *DAG) []string {
	state := map[string]int{}
	var path []string
	var visit func(id string) []string
	visit = func(id string) []string {
		u := strings.ToUpper(id)
		state[u] = 1
		path = append(path, id)
		if s := dag.SessionByID(id); s != nil {
			for _, dep := range s.Dependencies {
				switch state[strings.ToUpper(dep)] {
				case 1:
					for j, p := range path {
						if strings.EqualFold(p, dep) {
							return append(append([]string{}, path[j:]...), dep)
						}
					}
				case 0:
					if c := visit(dep); c != nil {
						return c
					}
				}
			}
		}
		state[u] = 2
		path = path[:len(path)-1]
		return nil
	}
	for _, s := range dag.Sessions {
		if state[strings.ToUpper(s.ID)] == 0 {
			if c := visit(s.ID); c != nil {
				return c
			}
		}
	}
	return nil
}

// sessionNamedIn returns the id of a listed session that heading mentions as a whole word, or "".
func sessionNamedIn(heading string, dag *DAG) string {
	upper := strings.ToUpper(heading)
	isIDChar := func(b byte) bool {
		return b == '-' || b == '_' || (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
	}
	for _, s := range dag.Sessions {
		id := strings.ToUpper(s.ID)
		for from := 0; ; {
			k := strings.Index(upper[from:], id)
			if k < 0 {
				break
			}
			k += from
			end := k + len(id)
			if (k == 0 || !isIDChar(upper[k-1])) && (end == len(upper) || !isIDChar(upper[end])) {
				return s.ID
			}
			from = k + 1
		}
	}
	return ""
}

// attachPrompts gives each session the one prompt fence under its "### <id>: <title>" heading in "## Sessions".
func attachPrompts(lay *planLayout, dag *DAG) []PlanIssue {
	if len(dag.Sessions) == 0 {
		return nil
	}
	if !lay.hasSessions {
		return []PlanIssue{{CodePlanMissingPrompt,
			fmt.Sprintf("the plan has no ## Sessions section, so none of its %d sessions has a prompt", len(dag.Sessions)),
			"Add a ## Sessions section with a ### <id>: <title> heading and one ```prompt fence for each session."}}
	}
	var issues []PlanIssue
	bySession := map[string][]int{}
	misplaced := map[string]bool{}
	// A deeper heading that names a listed session ("#### T1-02: ...") is a session heading at the wrong level: the
	// prompts after it are not the enclosing session's, and the session it names gets a fix for its heading.
	named := func(text string) string {
		if token, _ := splitSessionHeading(text); dag.SessionByID(token) != nil {
			return dag.SessionByID(token).ID
		}
		return sessionNamedIn(text, dag)
	}
	for _, h := range lay.subheadings {
		if id := named(h.text); id != "" {
			misplaced[strings.ToUpper(id)] = true
		}
	}
	promptsOf := func(sec planSection) []fenceRange {
		cut := sec.end
		for _, h := range lay.subheadings {
			if h.line > sec.heading && h.line < cut {
				if id := named(h.text); id != "" && !strings.EqualFold(id, sec.token) {
					cut = h.line
				}
			}
		}
		var out []fenceRange
		for _, f := range sec.prompts {
			if f.open < cut {
				out = append(out, f)
			}
		}
		return out
	}
	for k, sec := range lay.sections {
		if s := dag.SessionByID(sec.token); s != nil {
			bySession[strings.ToUpper(s.ID)] = append(bySession[strings.ToUpper(s.ID)], k)
			continue
		}
		if len(sec.prompts) == 0 {
			continue // a prose section, not a session
		}
		heading := strings.TrimSpace(lay.lines[sec.heading])
		if id := sessionNamedIn(heading, dag); id != "" {
			misplaced[strings.ToUpper(id)] = true
			continue
		}
		issues = append(issues, PlanIssue{CodePlanUnlistedSession,
			fmt.Sprintf("%q has a prompt but names no session the block lists", heading),
			fmt.Sprintf("Add %s to sessions: in the vivechak-plan block, or delete its section.", sec.token)})
	}
	for i := range dag.Sessions {
		s := &dag.Sessions[i]
		secs := bySession[strings.ToUpper(s.ID)]
		switch {
		case len(secs) == 0 && misplaced[strings.ToUpper(s.ID)]:
			issues = append(issues, PlanIssue{CodePlanMissingPrompt,
				fmt.Sprintf("session %s's heading is not a ### heading that starts with its id", s.ID),
				fmt.Sprintf("Write the heading as ### %s: <title> under ## Sessions.", s.ID)})
		case len(secs) == 0:
			issues = append(issues, PlanIssue{CodePlanMissingPrompt,
				fmt.Sprintf("session %s has no ### %s heading under ## Sessions", s.ID, s.ID),
				fmt.Sprintf("Add ### %s: <title> under ## Sessions, followed by one ```prompt fence with its prompt.", s.ID)})
		case len(secs) > 1:
			issues = append(issues, PlanIssue{CodePlanDuplicatePrompt,
				fmt.Sprintf("session %s has %d ### %s headings under ## Sessions", s.ID, len(secs), s.ID),
				fmt.Sprintf("Keep one ### %s section with one ```prompt fence.", s.ID)})
		default:
			sec := lay.sections[secs[0]]
			prompts := promptsOf(sec)
			switch len(prompts) {
			case 0:
				issues = append(issues, PlanIssue{CodePlanMissingPrompt,
					fmt.Sprintf("session %s has no ```prompt fence under its heading", s.ID),
					fmt.Sprintf("Put %s's prompt in one ```prompt fence under ### %s.", s.ID, s.ID)})
			case 1:
				s.Prompt = lay.fenceContent(prompts[0])
				if strings.TrimSpace(s.Prompt) == "" {
					issues = append(issues, PlanIssue{CodePlanMissingPrompt, fmt.Sprintf("session %s's prompt fence is empty", s.ID),
						fmt.Sprintf("Write %s's prompt inside its ```prompt fence.", s.ID)})
				}
				if s.Title == "" {
					s.Title = sec.title
				}
			default:
				issues = append(issues, PlanIssue{CodePlanDuplicatePrompt,
					fmt.Sprintf("session %s has %d ```prompt fences under its heading", s.ID, len(prompts)),
					fmt.Sprintf("Merge %s's prompts into exactly one ```prompt fence.", s.ID)})
			}
		}
	}
	return issues
}

// planScopeIssues checks the plan against the scope it is saved as ("project", "decision", or "" for the block's
// plan field): the plan field must agree, and a project plan has exactly one kind: synthesis session.
func planScopeIssues(dag *DAG, scope string) []PlanIssue {
	var issues []PlanIssue
	isProject := dag.IsProject()
	switch scope {
	case "project":
		if !isProject {
			issues = append(issues, invalidBlock(fmt.Sprintf("this plan is saved as the project pipeline, but its block says plan: %s", dag.Plan),
				"Set plan: PROJECT in the block, or save it as a decision plan with its decision id."))
		}
		isProject = true
	case "decision":
		if isProject {
			issues = append(issues, invalidBlock(fmt.Sprintf("this plan is saved as a decision plan, but its block says plan: %s", dag.Plan),
				"Set plan: to the decision id (e.g. plan: D-001), or save it as the project pipeline with id DAG."))
		}
		isProject = false
	}
	if isProject {
		var synth []string
		for _, s := range dag.SynthesisSessions() {
			synth = append(synth, s.ID)
		}
		switch len(synth) {
		case 1:
		case 0:
			issues = append(issues, PlanIssue{CodePlanSynthesis, "the project plan has no session with kind: synthesis",
				"Mark the session that writes research/FAD.md with kind: synthesis (exactly one)."})
		default:
			issues = append(issues, PlanIssue{CodePlanSynthesis,
				fmt.Sprintf("the project plan has %d sessions with kind: synthesis (%s)", len(synth), strings.Join(synth, ", ")),
				"Keep kind: synthesis on exactly one session, the one that writes research/FAD.md."})
		}
	}
	return issues
}

// isAllPriorSessions reports whether a dependency means every session before this one: ALL (or "all prior
// sessions", "*", "__ALL_PRIOR__", with an optional "(hard)"-style qualifier).
func isAllPriorSessions(s string) bool {
	raw := strings.TrimSpace(depCommentRe.ReplaceAllString(s, ""))
	if raw == "*" {
		return true
	}
	s = strings.Trim(raw, " \t\r\n.,;:[]`'\"*")
	lower := strings.ToLower(s)
	return lower == "all" ||
		lower == "all prior" ||
		lower == "all prior sessions" ||
		lower == "all previous" ||
		lower == "all previous sessions" ||
		lower == "all preceding" ||
		lower == "all preceding sessions" ||
		lower == "all other sessions" ||
		lower == "all non-synthesis sessions" ||
		lower == "all sessions" ||
		lower == "__all_prior__" ||
		lower == "__all__"
}

// depCommentRe matches parenthetical qualifiers in dependency lists like: (constrains), (hard), (soft)
var depCommentRe = regexp.MustCompile(`\s*\([^)]*\)`)

// AddSession appends a new research session to the DAG after verifying that dependencies exist and no cycles are
// created. ALL in its dependencies means every non-synthesis session (the new session runs before synthesis).
func (d *DAG) AddSession(s Session) error {
	s.ID = strings.Trim(strings.TrimSpace(s.ID), "*_`[]\"'")
	if s.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if !validSessionIDRe.MatchString(s.ID) {
		return fmt.Errorf("session ID %q is invalid — use alphanumeric characters with optional hyphens or underscores", s.ID)
	}
	if d.SessionByID(s.ID) != nil {
		return fmt.Errorf("session %s already exists in pipeline", s.ID)
	}

	// Validate and expand dependencies
	var expandedDeps, declared []string
	seenDeps := make(map[string]bool)
	for _, dep := range s.Dependencies {
		dep = strings.TrimSpace(dep)
		if isAllPriorSessions(dep) {
			s.dependsOnAll = true
			for _, prior := range d.Sessions {
				if !prior.IsSynthesis() && !seenDeps[prior.ID] {
					seenDeps[prior.ID] = true
					expandedDeps = append(expandedDeps, prior.ID)
				}
			}
		} else {
			if strings.EqualFold(dep, s.ID) {
				return fmt.Errorf("session %s cannot depend on itself", s.ID)
			}
			prior := d.SessionByID(dep)
			if prior == nil {
				return fmt.Errorf("dependency %s does not exist in pipeline", dep)
			}
			if prior.IsSynthesis() && !s.IsSynthesis() {
				return fmt.Errorf("session %s cannot depend on the synthesis session %s, which runs after every other session", s.ID, dep)
			}
			declared = append(declared, dep)
			if !seenDeps[dep] {
				seenDeps[dep] = true
				expandedDeps = append(expandedDeps, dep)
			}
		}
	}
	s.Dependencies = expandedDeps
	s.declaredDeps = declared

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

// RemoveSession removes a session from the DAG if no other session names it as a dependency. A dependency that
// comes only from depends: [ALL] does not hold the session back.
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
		deps := s.Dependencies
		if s.dependsOnAll {
			deps = s.declaredDeps
		}
		for _, dep := range deps {
			if strings.EqualFold(dep, id) {
				return fmt.Errorf("cannot remove session %s: session %s depends on it", id, s.ID)
			}
		}
	}

	d.Sessions = append(d.Sessions[:idx], d.Sessions[idx+1:]...)
	for i := range d.Sessions {
		var kept []string
		for _, dep := range d.Sessions[i].Dependencies {
			if !strings.EqualFold(dep, id) {
				kept = append(kept, dep)
			}
		}
		d.Sessions[i].Dependencies = kept
	}
	return nil
}

// UpdateDependencies changes dependencies for a session and verifies no cycles are formed. ALL means every
// non-synthesis session listed before it.
func (d *DAG) UpdateDependencies(id string, deps []string) error {
	id = strings.TrimSpace(id)
	s := d.SessionByID(id)
	if s == nil {
		return fmt.Errorf("session %s not found in pipeline", id)
	}

	var expandedDeps, declared []string
	dependsOnAll := false
	seenDeps := make(map[string]bool)
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if isAllPriorSessions(dep) {
			dependsOnAll = true
			for _, prior := range d.Sessions {
				if strings.EqualFold(prior.ID, id) {
					break
				}
				if !prior.IsSynthesis() && !seenDeps[prior.ID] {
					seenDeps[prior.ID] = true
					expandedDeps = append(expandedDeps, prior.ID)
				}
			}
		} else {
			if strings.EqualFold(dep, id) {
				return fmt.Errorf("session %s cannot depend on itself", id)
			}
			prior := d.SessionByID(dep)
			if prior == nil {
				return fmt.Errorf("dependency %s does not exist in pipeline", dep)
			}
			if prior.IsSynthesis() && !s.IsSynthesis() {
				return fmt.Errorf("session %s cannot depend on the synthesis session %s, which runs after every other session", id, dep)
			}
			declared = append(declared, dep)
			if !seenDeps[dep] {
				seenDeps[dep] = true
				expandedDeps = append(expandedDeps, dep)
			}
		}
	}

	oldDeps, oldDeclared, oldAll := s.Dependencies, s.declaredDeps, s.dependsOnAll
	s.Dependencies, s.declaredDeps, s.dependsOnAll = expandedDeps, declared, dependsOnAll
	if err := d.ValidateDAG(); err != nil {
		s.Dependencies, s.declaredDeps, s.dependsOnAll = oldDeps, oldDeclared, oldAll
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

// ReplanParams holds the parameters for in-place DAG replanning.
type ReplanParams struct {
	SessionID    string
	Topic        string
	Layer        int
	DoorType     string
	DecisionRef  string
	Dependencies []string
	Prompt       string
	Reason       string
	Impact       string
}

// ReplanInPlace applies a replan operation to a plan's text. It edits only the session's entry in the
// vivechak-plan block (in the style and with the keys the planner wrote) and the session's one prompt section under
// "## Sessions"; every other byte of the plan is kept. The edited plan is refused (with its V-PLAN-* problem) if it
// is no longer a valid plan, for example when the operation removes a project plan's synthesis session.
func ReplanInPlace(pipelineData []byte, op string, p ReplanParams) ([]byte, error) {
	if _, issues := parsePlan(pipelineData); len(issues) > 0 {
		return nil, &issues[0]
	}
	lay := layoutPlan(pipelineData)
	bl, err := locateBlockEntries(lay)
	if err != nil {
		return nil, err
	}
	targetID := strings.TrimSpace(p.SessionID)
	lines := append([]string{}, lay.lines...)
	entry := bl.entry(targetID)
	section := -1
	for k, sec := range lay.sections {
		if strings.EqualFold(sec.token, targetID) {
			section = k
			break
		}
	}

	switch op {
	case "add_session":
		if entry != nil {
			return nil, fmt.Errorf("session %s already exists in pipeline", targetID)
		}
		topic := strings.TrimSpace(p.Topic)
		if topic == "" {
			topic = targetID
		}
		prompt := strings.TrimSpace(p.Prompt)
		if prompt == "" {
			prompt = fmt.Sprintf("# %s: %s\n\nBRIEF:\nConduct research for %s.", targetID, topic, topic)
		}
		// The new session goes before the synthesis session (entry and section), else after the last one.
		entryAt, sectionAt := bl.entries[len(bl.entries)-1].end, lay.sessionsEnd
		if sectionAt > 0 && sectionAt == len(lines) && strings.TrimSpace(lines[sectionAt-1]) == "" {
			sectionAt-- // keep the plan's final newline last
		}
		for _, e := range bl.entries {
			if e.synthesis {
				entryAt = e.start
				for _, sec := range lay.sections {
					if strings.EqualFold(sec.token, e.id) {
						sectionAt = sec.heading
						break
					}
				}
				break
			}
		}
		newSection := []string{fmt.Sprintf("### %s: %s", targetID, topic), ""}
		fence := fenceFor(prompt, 3)
		newSection = append(newSection, fence+"prompt")
		newSection = append(newSection, strings.Split(prompt, "\n")...)
		newSection = append(newSection, fence, "")
		if sectionAt > 0 && strings.TrimSpace(lines[sectionAt-1]) != "" {
			newSection = append([]string{""}, newSection...)
		}
		lines = spliceEdits(lines, lineEdit{sectionAt, sectionAt, newSection}, lineEdit{entryAt, entryAt, bl.renderEntry(targetID, topic, p)})

	case "remove_session":
		if entry == nil {
			return nil, fmt.Errorf("session %s not found in pipeline", targetID)
		}
		edits := []lineEdit{{entry.start, entry.end, nil}}
		if section >= 0 {
			edits = append(edits, lineEdit{lay.sections[section].heading, lay.sections[section].end, nil})
		}
		lines = spliceEdits(lines, edits...)

	case "update_deps":
		if entry == nil {
			return nil, fmt.Errorf("session %s not found in pipeline", targetID)
		}
		updated, err := bl.withDepends(entry, p.Dependencies)
		if err != nil {
			return nil, err
		}
		lines = spliceLines(lines, entry.start, entry.end, updated)

	case "update_prompt":
		if section < 0 || len(lay.sections[section].prompts) != 1 {
			return nil, fmt.Errorf("session %s has no prompt section to update", targetID)
		}
		prompt := strings.TrimSpace(p.Prompt)
		if prompt == "" {
			return nil, fmt.Errorf("prompt content cannot be empty")
		}
		f := lay.sections[section].prompts[0]
		open := lines[f.open]
		n := leadingRun(strings.TrimSpace(open), '`')
		fence := fenceFor(prompt, n)
		closeLine := []string{fence}
		if f.close < len(lines) {
			closeLine = []string{lines[f.close]}
		}
		if len(fence) > n {
			open = strings.Replace(open, strings.Repeat("`", n), fence, 1)
			closeLine = []string{strings.Repeat(" ", len(lines[f.open])-len(strings.TrimLeft(lines[f.open], " \t"))) + fence}
		}
		body := append([]string{open}, strings.Split(prompt, "\n")...)
		end := f.close + 1
		if end > len(lines) {
			end = len(lines)
		}
		lines = spliceLines(lines, f.open, end, append(body, closeLine...))

	default:
		return nil, fmt.Errorf("unsupported replan operation %q", op)
	}

	// A plan written with CRLF line endings keeps them on the lines the edit adds.
	if usesCRLF(lay.lines) {
		for i := 0; i < len(lines)-1; i++ {
			if !strings.HasSuffix(lines[i], "\r") {
				lines[i] += "\r"
			}
		}
	}
	out := []byte(strings.Join(lines, "\n"))
	if dag, issues := parsePlan(out); len(issues) > 0 {
		return nil, &issues[0]
	} else if scope := planScopeIssues(dag, ""); len(scope) > 0 {
		return nil, &scope[0]
	}
	return out, nil
}

// usesCRLF reports whether every line of a plan (split on "\n") but the last ends with "\r": a CRLF file.
func usesCRLF(lines []string) bool {
	if len(lines) < 2 {
		return false
	}
	for _, l := range lines[:len(lines)-1] {
		if !strings.HasSuffix(l, "\r") {
			return false
		}
	}
	return true
}

func spliceLines(lines []string, from, to int, with []string) []string {
	out := make([]string, 0, len(lines)-(to-from)+len(with))
	out = append(out, lines[:from]...)
	out = append(out, with...)
	return append(out, lines[to:]...)
}

// lineEdit replaces lines [from, to) with its lines.
type lineEdit struct {
	from, to int
	lines    []string
}

// spliceEdits applies non-overlapping edits, the later one first so every index still points at the original lines.
func spliceEdits(lines []string, edits ...lineEdit) []string {
	sort.Slice(edits, func(i, j int) bool { return edits[i].from > edits[j].from })
	for _, e := range edits {
		lines = spliceLines(lines, e.from, e.to, e.lines)
	}
	return lines
}

// fenceFor returns a backtick fence at least min long and longer than any backtick run that starts a line of prompt.
func fenceFor(prompt string, min int) string {
	n := min
	for _, l := range strings.Split(prompt, "\n") {
		if r := leadingRun(strings.TrimSpace(l), '`'); r >= n {
			n = r + 1
		}
	}
	return strings.Repeat("`", n)
}

// blockEntry is one session entry of the plan's block, as document lines [start, end).
type blockEntry struct {
	id        string
	start     int
	end       int
	node      *yaml.Node
	synthesis bool
}

type blockEntries struct {
	open    int // the block's opening fence line
	lines   []string
	entries []blockEntry
}

func (b *blockEntries) entry(id string) *blockEntry {
	for i := range b.entries {
		if strings.EqualFold(b.entries[i].id, id) {
			return &b.entries[i]
		}
	}
	return nil
}

// locateBlockEntries finds each session entry of the block in the plan's lines.
func locateBlockEntries(lay *planLayout) (*blockEntries, error) {
	blk := lay.blocks[0]
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(lay.fenceContent(blk)), &root); err != nil {
		return nil, invalidBlockError("the vivechak-plan block is not valid YAML: " + err.Error())
	}
	seq := mappingValue(resolveAlias(root.Content[0]), "sessions")
	if seq == nil || seq.Kind != yaml.SequenceNode || seq.Style&yaml.FlowStyle != 0 {
		return nil, &PlanIssue{CodePlanInvalidBlock, "replan edits a sessions: list written one entry per line",
			"Write sessions: as one - {id: ...} entry per line and run the replan again."}
	}
	b := &blockEntries{open: blk.open, lines: lay.lines}
	for k, item := range seq.Content {
		item = resolveAlias(item)
		start := blk.open + item.Line
		if start >= blk.close || strings.TrimSpace(runePrefix(lay.lines[start], item.Column-1)) != "-" {
			return nil, &PlanIssue{CodePlanInvalidBlock, fmt.Sprintf("sessions entry %d does not start on its - line", k+1),
				"Start each session entry on its own line with - and run the replan again."}
		}
		id, _ := scalarText(mappingValue(item, "id"))
		kind, _ := scalarText(mappingValue(item, "kind"))
		b.entries = append(b.entries, blockEntry{id: id, start: start, node: item, synthesis: strings.EqualFold(kind, SynthesisKind)})
	}
	for k := range b.entries {
		limit := blk.close
		if k+1 < len(b.entries) {
			limit = b.entries[k+1].start
		}
		dash := indentOf(lay.lines[b.entries[k].start])
		end := b.entries[k].start + 1
		for end < limit {
			t := strings.TrimSpace(lay.lines[end])
			if t != "" && indentOf(lay.lines[end]) <= dash && k+1 == len(b.entries) {
				break
			}
			end++
		}
		for end > b.entries[k].start+1 && strings.TrimSpace(lay.lines[end-1]) == "" {
			end--
		}
		b.entries[k].end = end
	}
	return b, nil
}

func invalidBlockError(msg string) error {
	return &PlanIssue{CodePlanInvalidBlock, msg, "Fix the block's YAML and run the replan again."}
}

func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

// runePrefix returns the first n runes of s (yaml.v3 columns count runes).
func runePrefix(s string, n int) string {
	return s[:runeOffset(s, n)]
}

func runeOffset(s string, n int) int {
	off := 0
	for i := 0; i < n && off < len(s); i++ {
		_, size := utf8.DecodeRuneInString(s[off:])
		off += size
	}
	return off
}

// renderEntry writes a new session entry in the style (flow or block) and indentation of the block's entries.
func (b *blockEntries) renderEntry(id, title string, p ReplanParams) []string {
	ref := b.entries[len(b.entries)-1]
	door := strings.ToLower(strings.TrimSpace(p.DoorType))
	if door == "" {
		door = "two-way"
	}
	pairs := [][2]string{{"id", id}, {"title", yamlScalar(title)}}
	if p.Layer != 0 {
		pairs = append(pairs, [2]string{"layer", fmt.Sprint(p.Layer)})
	}
	pairs = append(pairs, [2]string{"door", door})
	if ref := strings.TrimSpace(p.DecisionRef); ref != "" {
		pairs = append(pairs, [2]string{"informs", "[" + yamlScalar(ref) + "]"})
	}
	pairs = append(pairs, [2]string{"depends", flowList(p.Dependencies)})

	refLine := b.lines[ref.start]
	dashCol := indentOf(refLine)
	if ref.node.Style&yaml.FlowStyle != 0 {
		var parts []string
		for _, kv := range pairs {
			parts = append(parts, kv[0]+": "+kv[1])
		}
		return []string{refLine[:dashCol] + "- {" + strings.Join(parts, ", ") + "}"}
	}
	keyCol := len(runePrefix(refLine, ref.node.Column-1))
	if keyCol <= dashCol {
		keyCol = dashCol + 2
	}
	out := []string{refLine[:dashCol] + "-" + strings.Repeat(" ", keyCol-dashCol-1) + pairs[0][0] + ": " + pairs[0][1]}
	for _, kv := range pairs[1:] {
		out = append(out, strings.Repeat(" ", keyCol)+kv[0]+": "+kv[1])
	}
	return out
}

// flowList writes dependency ids as a flow list; any form of ALL is written as ALL.
func flowList(ids []string) string {
	return "[" + strings.Join(depTokens(ids), ", ") + "]"
}

func depTokens(ids []string) []string {
	var out []string
	for _, id := range ids {
		if id = strings.TrimSpace(id); id == "" {
			continue
		}
		if isAllPriorSessions(id) {
			id = "ALL"
		}
		out = append(out, yamlScalar(id))
	}
	return out
}

var plainYAMLRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9 ._/()+&$'-]*$`)

// yamlScalar writes s as a plain YAML scalar when that is safe in a flow mapping, else double-quoted.
func yamlScalar(s string) string {
	switch strings.ToLower(s) {
	case "true", "false", "yes", "no", "on", "off", "null", "y", "n":
		return fmt.Sprintf("%q", s)
	}
	if plainYAMLRe.MatchString(s) && !strings.HasSuffix(s, " ") {
		return s
	}
	return fmt.Sprintf("%q", s)
}

// withDepends returns the entry's lines with its depends value set to deps, changing nothing else in the entry.
func (b *blockEntries) withDepends(e *blockEntry, deps []string) ([]string, error) {
	text := strings.Join(b.lines[e.start:e.end], "\n")
	lineStart := []int{0}
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lineStart = append(lineStart, i+1)
		}
	}
	offset := func(n *yaml.Node) int {
		l := b.open + n.Line - e.start
		if l < 0 || l >= len(lineStart) {
			return -1
		}
		lineEnd := len(text)
		if l+1 < len(lineStart) {
			lineEnd = lineStart[l+1] - 1
		}
		return lineStart[l] + runeOffset(text[lineStart[l]:lineEnd], n.Column-1)
	}
	list := flowList(deps)
	var key, val *yaml.Node
	for i := 0; i+1 < len(e.node.Content); i += 2 {
		if e.node.Content[i].Value == "depends" {
			key, val = e.node.Content[i], e.node.Content[i+1]
		}
	}
	var out string
	switch {
	case key == nil && e.node.Style&yaml.FlowStyle != 0:
		open := offset(e.node)
		closeAt := matchBracket(text, open)
		if open < 0 || closeAt < 0 {
			return nil, invalidBlockError(fmt.Sprintf("session %s's entry can't be edited", e.id))
		}
		at := closeAt
		for at > open+1 && (text[at-1] == ' ' || text[at-1] == '\t') {
			at--
		}
		sep := ", "
		if strings.TrimSpace(text[open+1:at]) == "" {
			sep = ""
		}
		out = text[:at] + sep + "depends: " + list + text[at:]
	case key == nil:
		keyCol := strings.Repeat(" ", len(runePrefix(b.lines[e.start], e.node.Column-1)))
		out = text + "\n" + keyCol + "depends: " + list
	case val.Kind == yaml.SequenceNode && val.Style&yaml.FlowStyle != 0:
		open := offset(val)
		closeAt := matchBracket(text, open)
		if open < 0 || closeAt < 0 {
			return nil, invalidBlockError(fmt.Sprintf("session %s's depends list can't be edited", e.id))
		}
		out = text[:open] + list + text[closeAt+1:]
	case val.Kind == yaml.SequenceNode:
		// A block list: rewrite its item lines in the same indentation.
		first, last := b.open+val.Content[0].Line, b.open+val.Content[len(val.Content)-1].Line
		itemIndent := b.lines[first][:indentOf(b.lines[first])]
		var items []string
		for _, id := range depTokens(deps) {
			items = append(items, itemIndent+"- "+id)
		}
		lines := append([]string{}, b.lines[e.start:e.end]...)
		if len(items) == 0 {
			keyLine := b.open + key.Line - e.start
			lines[keyLine] = strings.TrimRight(lines[keyLine], " \t\r") + " []"
		}
		return spliceLines(lines, first-e.start, last-e.start+1, items), nil
	default: // depends: with no value, ~ or null
		at := offset(val)
		if at < 0 {
			return nil, invalidBlockError(fmt.Sprintf("session %s's depends value can't be edited", e.id))
		}
		end := at + len(val.Value)
		if val.Value == "" {
			out = text[:at] + " " + list + text[at:]
		} else {
			out = text[:at] + list + text[end:]
		}
	}
	return strings.Split(out, "\n"), nil
}

// matchBracket returns the index of the bracket that closes the one at text[open], skipping quoted text, or -1.
func matchBracket(text string, open int) int {
	if open < 0 || open >= len(text) || (text[open] != '[' && text[open] != '{') {
		return -1
	}
	depth := 0
	for i := open; i < len(text); i++ {
		switch text[i] {
		case '[', '{':
			depth++
		case ']', '}':
			depth--
			if depth == 0 {
				return i
			}
		case '"':
			for i++; i < len(text) && text[i] != '"'; i++ {
				if text[i] == '\\' {
					i++
				}
			}
		case '\'':
			for i++; i < len(text); i++ {
				if text[i] == '\'' {
					if i+1 < len(text) && text[i+1] == '\'' {
						i++
						continue
					}
					break
				}
			}
		}
	}
	return -1
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
		} else if s.IsSynthesis() {
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
