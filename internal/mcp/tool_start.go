package mcputil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	vembed "github.com/bhaskarjha-dev/vivechak/internal/embed"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	versionPkg "github.com/bhaskarjha-dev/vivechak/internal/version"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StartInput holds the canonical arguments for vivechak_start.
type StartInput struct {
	ProjectRoot string   `json:"project_root,omitempty"`
	Scope       string   `json:"scope"`
	Question    string   `json:"question"`
	Options     []string `json:"options,omitempty"`
	Criteria    string   `json:"criteria,omitempty"`
	Part        int      `json:"part,omitempty"`
}

func registerStart(server *sdkmcp.Server) {
	registerTool(server, ToolSpec{
		Name:        "vivechak_start",
		Description: "Start technical research (comparison, decision, or project). Returns a research brief to run with web search and an ID. Not for coding.",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint:    false,
			IdempotentHint:  false,
			DestructiveHint: BoolPtr(true),
		},
		Params: []ParamSpec{
			{Name: "scope", Type: "string", Enum: []string{"comparison", "decision", "project"}, Required: true, Doc: "Research scope: comparison (2-5 options), decision (ADR), or project (pipeline -> FAD)."},
			{Name: "question", Type: "string", Required: true, Description: "Choice or vision", Doc: "Technical dilemma, options to compare, or project vision."},
			{Name: "options", Type: "array", ItemType: "string", Doc: "Named options for comparison scope."},
			{Name: "criteria", Type: "string", Doc: "Optional evaluation criteria or constraints."},
			{Name: "part", Type: "integer", Description: "Brief part", Doc: "1-based part index for multi-part project briefs."},
			{Name: "project_root", Type: "string", Doc: "Absolute path to the target project directory."},
		},
	}, handleStart)
}

func handleStart(ctx context.Context, req *sdkmcp.CallToolRequest, in StartInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_start"

	var roots []string
	if strings.TrimSpace(in.ProjectRoot) == "" {
		roots = getClientRoots(ctx, req)
	}
	root, rule, err := core.ResolveWorkspaceWithRoots(ctx, in.ProjectRoot, roots)
	if err != nil {
		if in.ProjectRoot == "" {
			return ErrorResult(tool, err, "Pass project_root=<absolute path of the user's project> explicitly.")
		}
		return ErrorResult(tool, err, "Pass a valid project_root directory path.")
	}

	question := strings.TrimSpace(in.Question)

	rawScope := strings.ToLower(strings.TrimSpace(in.Scope))
	if rawScope == "" {
		if question != "" {
			rawScope = string(core.InferScope(question))
		} else {
			rawScope = "project"
		}
	}
	scope, ok := core.ScopeFromString(rawScope)
	if !ok {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: invalid scope %q", in.Scope),
			"Provide scope: 'comparison', 'decision', or 'project'.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: opening workspace %q: %w", root, err),
			"Ensure workspace directory is accessible.")
	}
	defer ws.Close()

	// If a project pipeline already exists and no explicit part was requested, return its status instead of a fresh planning brief (HIGH-06).
	if scope == core.ScopeProject && in.Part == 0 && ws.FileExists(core.PipelineFile) {
		if pipeData, rErr := ws.ReadFile(core.PipelineFile); rErr == nil && len(strings.TrimSpace(string(pipeData))) > 0 {
			dag, pErr := core.ParsePipeline(pipeData)
			totalSessions := 0
			completedCount := 0
			if pErr == nil && dag != nil {
				totalSessions = len(dag.Sessions)
				completedIDs, _ := dag.CompletedSessions(ws)
				completedCount = len(completedIDs)
			}
			hasFAD := false
			if _, sErr := os.Stat(filepath.Join(root, core.FADFile)); sErr == nil {
				hasFAD = true
			}
			nextTool := "vivechak_next"
			nextWhy := "Pipeline already exists. Call vivechak_next to execute the next ready session."
			if hasFAD && totalSessions > 0 && completedCount == totalSessions {
				nextTool = "vivechak_gate"
				nextWhy = "All pipeline sessions and FAD.md exist. Call vivechak_gate to verify Phase 0 exit criteria."
			}
			env := Envelope{
				Ok:      true,
				Message: fmt.Sprintf("Project pipeline already exists at %s (%d/%d sessions completed).", core.PipelineFile, completedCount, totalSessions),
				Data: map[string]any{
					"workspace":          map[string]any{"root": root, "rule": rule},
					"workspace_root":     root,
					"scope":              scope,
					"initialized":        true,
					"existing_pipeline":  true,
					"pipeline_file":      core.PipelineFile,
					"total_sessions":     totalSessions,
					"completed_sessions": completedCount,
					"has_fad":            hasFAD,
				},
				Next: &NextAction{
					Tool: nextTool,
					Args: map[string]any{},
					Why:  nextWhy,
				},
			}
			return env.ToResult()
		}
	}

	dirs := []string{
		core.ResearchDir,
		core.SessionsDir,
	}
	for _, dir := range dirs {
		if err := ws.MkdirAll(dir, 0o755); err != nil {
			return ErrorResult(tool, fmt.Errorf("E-IO: creating %s: %w", dir, err),
				"Check filesystem permissions and try again.")
		}
	}

	gitignoreRel := filepath.Join(core.ResearchDir, ".gitignore")
	if !ws.FileExists(gitignoreRel) {
		_ = store.WriteFileAtomic(ws.Root(), gitignoreRel, []byte("# Vivechak operational files\n*.lock\n.locks/\n.history/\n"), 0o644)
	}

	meta := core.WorkspaceMeta{
		Schema:    1,
		Scope:     scope,
		CreatedBy: fmt.Sprintf("vivechak %s", versionPkg.Version),
		Counters:  map[string]int{"C": 0, "D": 0},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if ws.FileExists(core.MetadataFile) {
		if existingData, err := ws.ReadFile(core.MetadataFile); err == nil {
			var existingMeta core.WorkspaceMeta
			if err := json.Unmarshal(existingData, &existingMeta); err == nil {
				if existingMeta.CreatedAt != "" {
					meta.CreatedAt = existingMeta.CreatedAt
				}
				if existingMeta.Counters != nil {
					meta.Counters = existingMeta.Counters
				}
				if existingMeta.Scope == core.ScopeProject || existingMeta.Scope == core.ScopeDecision {
					meta.Scope = existingMeta.Scope
				}
				meta.Schema = 1
			}
		}
	}
	if metaBytes, err := json.MarshalIndent(meta, "", "  "); err == nil {
		_ = store.WriteFileAtomic(ws.Root(), core.MetadataFile, metaBytes, 0o644)
	}

	data := map[string]any{
		"workspace": map[string]any{
			"root": root,
			"rule": rule,
		},
		"workspace_root": root,
		"scope":          scope,
		"initialized":    true,
	}

	genFile := core.GeneratorFile(scope)
	var genPrompt string
	if genBytes, err := vembed.ReadGenerator(genFile); err == nil {
		genPrompt = string(genBytes)
		switch scope {
		case core.ScopeProject:
			if startIdx := strings.Index(genPrompt, "````markdown"); startIdx >= 0 {
				afterFence := startIdx + len("````markdown")
				for afterFence < len(genPrompt) && genPrompt[afterFence] != '\n' {
					afterFence++
				}
				if afterFence < len(genPrompt) {
					afterFence++
				}
				if endIdx := strings.LastIndex(genPrompt[afterFence:], "````"); endIdx >= 0 {
					genPrompt = strings.TrimSpace(genPrompt[afterFence : afterFence+endIdx])
				}
			}
			placeholder := "[PASTE YOUR PROJECT DESCRIPTION HERE]"
			if strings.Contains(genPrompt, placeholder) {
				genPrompt = strings.Replace(genPrompt, placeholder, question, 1)
			} else if question != "" {
				genPrompt = genPrompt + "\n\n## Context\n\n" + question
			}
		case core.ScopeDecision, core.ScopeComparison:
			contextBlockStart := "```context"
			contextBlockEnd := "```"
			startIdx := strings.Index(genPrompt, contextBlockStart)
			if startIdx >= 0 {
				afterStart := startIdx + len(contextBlockStart)
				endIdx := strings.Index(genPrompt[afterStart:], contextBlockEnd)
				if endIdx >= 0 {
					endIdx += afterStart
					blockContent := genPrompt[afterStart:endIdx]
					today := time.Now().UTC().Format("2006-01-02")

					if strings.Contains(question, "DECISION:") || strings.Contains(question, "OPTIONS:") || strings.Contains(question, "CONTEXT:") {
						newBlock := "\n" + strings.TrimSpace(question) + "\n"
						newBlock = strings.Replace(newBlock, "[today]", today, 1)
						genPrompt = genPrompt[:afterStart] + newBlock + genPrompt[endIdx:]
					} else {
						placeholderDecision := "DECISION: [the question, as you'd ask it]"
						placeholderContextDecision := "CONTEXT: [what's being built; workload, team, constraints, what's already decided]"
						placeholderContextComparison := "CONTEXT: [what's being built, compared, or decided; workload, requirements, team, constraints, stage]"
						placeholderOptions := "OPTIONS: [2-5 named options]"
						placeholderCriteria := "CRITERIA: [optional; weights if known]"

						newBlock := blockContent
						switch scope {
						case core.ScopeDecision:
							if strings.Contains(newBlock, placeholderDecision) {
								newBlock = strings.Replace(newBlock, placeholderDecision, "DECISION: "+question, 1)
							} else if strings.Contains(newBlock, "DECISION:") {
								newBlock = reDecisionLine.ReplaceAllString(newBlock, "DECISION: "+question)
							}
						case core.ScopeComparison:
							if strings.Contains(newBlock, placeholderOptions) {
								opts := extractComparisonOptions(question)
								if len(in.Options) > 0 {
									opts = strings.Join(in.Options, ", ")
								}
								if opts != "" {
									newBlock = strings.Replace(newBlock, placeholderOptions, "OPTIONS: "+opts, 1)
								}
							}
							if strings.TrimSpace(in.Criteria) != "" && strings.Contains(newBlock, placeholderCriteria) {
								newBlock = strings.Replace(newBlock, placeholderCriteria, "CRITERIA: "+strings.TrimSpace(in.Criteria), 1)
							}
						}

						if strings.Contains(newBlock, placeholderContextDecision) {
							newBlock = strings.Replace(newBlock, placeholderContextDecision, "CONTEXT: "+question, 1)
						} else if strings.Contains(newBlock, placeholderContextComparison) {
							newBlock = strings.Replace(newBlock, placeholderContextComparison, "CONTEXT: "+question, 1)
						} else if strings.Contains(newBlock, "CONTEXT:") {
							newBlock = reContextLine.ReplaceAllString(newBlock, "CONTEXT: "+question)
						}
						newBlock = strings.Replace(newBlock, "[today]", today, 1)
						genPrompt = genPrompt[:afterStart] + newBlock + genPrompt[endIdx:]
					}
				}
			} else if question != "" {
				genPrompt = genPrompt + "\n\n## Context\n\n" + question
			}
		}
	}

	parts := []string{genPrompt}
	if scope == core.ScopeProject && len(genPrompt) > 18000 {
		parts = splitBriefIntoParts(genPrompt, 18000)
	}
	totalParts := len(parts)

	var activeBrief string
	var currentPart int
	var hasMore bool

	if in.Part > 0 {
		if in.Part > totalParts {
			return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: invalid part %d: brief has %d total parts", in.Part, totalParts),
				fmt.Sprintf("Request a part between 1 and %d.", totalParts))
		}
		currentPart = in.Part
		activeBrief = parts[currentPart-1]
		hasMore = currentPart < totalParts
	} else {
		currentPart = 1
		activeBrief = genPrompt
		hasMore = false
	}

	data["brief"] = activeBrief
	data["part"] = currentPart
	data["total_parts"] = totalParts
	data["has_more"] = hasMore
	data["char_count"] = len(activeBrief)
	data["approx_tokens"] = len(activeBrief) / 4
	data["total_char_count"] = len(genPrompt)
	data["generator_file"] = genFile

	var next *NextAction
	var msg string

	switch scope {
	case core.ScopeComparison:
		cid, allocErr := store.AllocateID(ws, "C")
		if allocErr != nil {
			return ErrorResult(tool, fmt.Errorf("E-IO: allocating comparison ID: %w", allocErr), "Ensure research/.vivechak.json is accessible.")
		}
		if st := stateForSession(sessionOf(req)); st != nil {
			st.reserveID(root, "C", cid)
		}
		data["id"] = cid

		metaInstruction := "From the context block, output a SCOPE CHECK line, then ONE research prompt in a `prompt` code fence: fill {slots} ({a | b} = pick one; {filename} = {ID}-cmp-{slug}.md), keep the rest verbatim. No plan, ADR, personas, search queries or search-count minimums."
		directInstruction := "Execute this evidence-calibrated research brief with live web search to compare the options below and produce the Weighted Evaluation Protocol (WEP) matrix. Do not output a prompt template; execute the comparison research directly."
		if strings.Contains(activeBrief, metaInstruction) {
			activeBrief = strings.Replace(activeBrief, metaInstruction, directInstruction, 1)
		} else {
			reMeta := regexp.MustCompile(`(?m)^From the context block, output a SCOPE CHECK line.*$`)
			activeBrief = reMeta.ReplaceAllString(activeBrief, directInstruction)
		}
		activeBrief = strings.Replace(activeBrief, "ID: [optional; default C-01]", fmt.Sprintf("ID: %s (persisted to research/%s-comparison.md)", cid, cid), 1)
		activeBrief = strings.Replace(activeBrief, "{filename} = {ID}-cmp-{slug}.md", fmt.Sprintf("filename = research/%s-comparison.md", cid), 1)
		data["brief"] = activeBrief

		options := in.Options
		if len(options) == 0 && question != "" {
			extracted := extractComparisonOptions(question)
			if extracted != "" {
				for _, opt := range strings.Split(extracted, ",") {
					trimmed := strings.TrimSpace(opt)
					if trimmed != "" {
						options = append(options, trimmed)
					}
				}
			}
		}
		data["options"] = options

		optsStr := "Options to evaluate"
		if len(options) > 0 {
			optsStr = strings.Join(options, " vs ")
		} else if question != "" {
			optsStr = question
		}

		msg = fmt.Sprintf("Comparison workspace initialized. Allocated ID %s for %s.", cid, optsStr)
		next = &NextAction{
			Tool: "vivechak_save",
			Args: map[string]any{
				"kind": "comparison",
				"id":   cid,
			},
			Why: fmt.Sprintf("Run web search to evaluate %s, then save the WEP comparison matrix using vivechak_save(kind=\"comparison\", id=%q, content=...).", optsStr, cid),
		}

	case core.ScopeDecision:
		did, allocErr := store.AllocateID(ws, "D")
		if allocErr != nil {
			return ErrorResult(tool, fmt.Errorf("E-IO: allocating decision ID: %w", allocErr), "Ensure research/.vivechak.json is accessible.")
		}
		if st := stateForSession(sessionOf(req)); st != nil {
			st.reserveID(root, "D", did)
		}
		data["id"] = did

		activeBrief = strings.Replace(activeBrief, "IDS: [optional: decision ID (default D-NEW), session prefix]", fmt.Sprintf("IDS: %s (session prefix %s-S)", did, did), 1)
		data["brief"] = activeBrief

		topic := question
		if topic == "" {
			topic = "Architectural Decision"
		}

		msg = fmt.Sprintf("Decision workspace initialized. Allocated ID %s for %s.", did, topic)
		next = &NextAction{
			Tool: "vivechak_save",
			Args: map[string]any{
				"kind":  "plan",
				"scope": "decision",
				"id":    did,
			},
			Why: fmt.Sprintf("Generate the decision plan from the brief and save it via vivechak_save(kind=\"plan\", id=%q, content=...).", did),
		}

	case core.ScopeProject:
		projName := filepath.Base(root)
		if question != "" {
			projName = question
		}

		if hasMore {
			nextPart := currentPart + 1
			msg = fmt.Sprintf("Project workspace initialized for %s. Serving brief part %d of %d.", projName, currentPart, totalParts)
			next = &NextAction{
				Tool: "vivechak_start",
				Args: map[string]any{
					"scope":    "project",
					"question": question,
					"part":     nextPart,
				},
				Why: fmt.Sprintf("Retrieve part %d of %d of the project brief before generating the research pipeline DAG.", nextPart, totalParts),
			}
		} else {
			msg = fmt.Sprintf("Project workspace initialized for %s.", projName)
			next = &NextAction{
				Tool: "vivechak_save",
				Args: map[string]any{
					"kind":  "plan",
					"scope": "project",
				},
				Why: "Decompose the project vision into a research pipeline DAG and save it via vivechak_save(kind=\"plan\", content=...).",
			}
		}
	}

	env := Envelope{
		Ok:      true,
		Message: msg,
		Data:    data,
		Next:    next,
	}
	return env.ToResult()
}

var (
	comparisonDelimRe  = regexp.MustCompile(`(?i)\s+(?:vs\.?|versus)\s+`)
	comparisonPrefixRe = regexp.MustCompile(`(?i)^(?:use\s+vivechak\s+to\s+|please\s+)?(?:compare|evaluate|benchmark|choose\s+between|decide\s+between|which\s+one\s+is\s+better:?|which\s+is\s+better[:,]?)\s+`)
	comparisonSuffixRe = regexp.MustCompile(`(?i)\s+(?:for|in|with|when|under|on|as|to|across|over|within)\s+.*$`)
	comparisonMidRe    = regexp.MustCompile(`(?i)(?:,\s*(?:and|or)\s+|\s+(?:vs\.?|versus|and|or)\s+|,\s+)`)
	comparisonFromToRe = regexp.MustCompile(`(?i)\bfrom\s+(.+?)\s+to\s+(.+)$`)
	reDecisionLine     = regexp.MustCompile(`(?m)^DECISION:.*$`)
	reContextLine      = regexp.MustCompile(`(?m)^CONTEXT:.*$`)
)

func extractComparisonOptions(ctx string) string {
	trimmed := strings.TrimSpace(ctx)
	if trimmed == "" {
		return ""
	}

	// 1. Explicit "A vs B [vs C]" / "A versus B"
	if loc := comparisonDelimRe.FindStringIndex(trimmed); loc != nil {
		left := strings.TrimSpace(comparisonPrefixRe.ReplaceAllString(strings.TrimSpace(trimmed[:loc[0]]), ""))
		right := strings.TrimSpace(comparisonSuffixRe.ReplaceAllString(strings.TrimSpace(trimmed[loc[1]:]), ""))
		parts := splitAndCleanOptions(left + ", " + comparisonMidRe.ReplaceAllString(right, ", "))
		if len(parts) >= 2 {
			return strings.Join(parts, ", ")
		}
	}

	// 2. "compare A and B", "compare A, B, and C", "choose between A and B"
	if comparisonPrefixRe.MatchString(trimmed) {
		body := strings.TrimSpace(comparisonPrefixRe.ReplaceAllString(trimmed, ""))
		body = strings.TrimSpace(comparisonSuffixRe.ReplaceAllString(body, ""))
		if comparisonMidRe.MatchString(body) {
			parts := splitAndCleanOptions(comparisonMidRe.ReplaceAllString(body, ", "))
			if len(parts) >= 2 && len(parts) <= 5 {
				return strings.Join(parts, ", ")
			}
		}
	}

	// 3. "from A to B"
	if m := comparisonFromToRe.FindStringSubmatch(trimmed); len(m) == 3 {
		left := strings.TrimSpace(m[1])
		right := strings.TrimSpace(comparisonSuffixRe.ReplaceAllString(strings.TrimSpace(m[2]), ""))
		parts := splitAndCleanOptions(left + ", " + right)
		if len(parts) >= 2 {
			return strings.Join(parts, ", ")
		}
	}

	return ""
}

func splitAndCleanOptions(joined string) []string {
	raw := strings.Split(joined, ",")
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		clean := strings.Trim(strings.TrimSpace(item), "?!.,:;\"'`")
		if clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func splitBriefIntoParts(prompt string, targetPartSize int) []string {
	if len(prompt) <= targetPartSize {
		return []string{prompt}
	}

	parts := make([]string, 0, 2)
	remaining := prompt
	for len(remaining) > targetPartSize {
		splitTarget := targetPartSize
		searchStart := splitTarget - 3000
		if searchStart < 0 {
			searchStart = 0
		}
		searchEnd := splitTarget + 3000
		if searchEnd > len(remaining) {
			searchEnd = len(remaining)
		}
		window := remaining[searchStart:searchEnd]

		splitIdx := -1
		if idx := strings.LastIndex(window, "\n### "); idx != -1 {
			splitIdx = searchStart + idx
		} else if idx := strings.LastIndex(window, "\n## "); idx != -1 {
			splitIdx = searchStart + idx
		} else if idx := strings.LastIndex(window, "\n\n"); idx != -1 {
			splitIdx = searchStart + idx + 1
		}

		if splitIdx <= 0 || splitIdx >= len(remaining) {
			splitIdx = targetPartSize
		}

		parts = append(parts, strings.TrimSpace(remaining[:splitIdx]))
		remaining = strings.TrimSpace(remaining[splitIdx:])
	}
	if len(remaining) > 0 {
		parts = append(parts, remaining)
	}
	return parts
}
