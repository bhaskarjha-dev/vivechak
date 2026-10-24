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
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StartInput holds the arguments for vivechak_start.
type StartInput struct {
	ProjectRoot string   `json:"project_root,omitempty" jsonschema:"workspace root path (optional; auto-discovered if omitted)"`
	Scope       string   `json:"scope,omitempty"        jsonschema:"research scope: 'comparison' (comparing 2-5 options) | 'decision' (single architectural decision/ADR) | 'project' (full system architecture pipeline)"`
	Question    string   `json:"question,omitempty"     jsonschema:"the technical dilemma, comparison options, or project vision"`
	Options     []string `json:"options,omitempty"      jsonschema:"optional list of named options to compare (e.g. ['Pino', 'Winston'])"`
	Criteria    string   `json:"criteria,omitempty"     jsonschema:"optional key evaluation criteria or constraints"`
	Part        int      `json:"part,omitempty"         jsonschema:"optional 1-based part index for multi-part delivery of large project briefs"`
}

// InitInput provides backward compatibility for unit tests calling handleInit.
type InitInput struct {
	ProjectRoot string `json:"project_root,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// PrepareGeneratorInput provides backward compatibility for unit tests calling handlePrepareGenerator.
type PrepareGeneratorInput struct {
	ProjectRoot string `json:"project_root,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Context     string `json:"context,omitempty"`
}

func registerStart(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_start",
			Title: "Start Research Workspace & Brief",
			Description: "Entry point for all Vivechak operations. Initializes research workspace in the user's project " +
				"and returns an executable research brief with an allocated identifier. " +
				"Scopes: 'comparison' (rapid evaluation between 2-5 options -> WEP matrix), " +
				"'decision' (architectural decision in an active codebase -> ADR), " +
				"or 'project' (greenfield architecture -> full session DAG pipeline). " +
				"Pass the user's technical dilemma or options as 'question'.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleStart,
	)
}

func handleStart(_ context.Context, _ *sdkmcp.CallToolRequest, in StartInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_start"

	// Resolve workspace: explicit argument -> environment variable -> CWD
	root := in.ProjectRoot
	if root == "" {
		if env := os.Getenv("VIVECHAK_PROJECT_ROOT"); env != "" && !strings.HasPrefix(env, "${") {
			root = env
		} else if env := os.Getenv("VCK_PROJECT_ROOT"); env != "" && !strings.HasPrefix(env, "${") {
			root = env
		} else if env := os.Getenv("VIVECHAK_DEFAULT_ROOT"); env != "" && !strings.HasPrefix(env, "${") {
			root = env
		} else if env := os.Getenv("VCK_DEFAULT_ROOT"); env != "" && !strings.HasPrefix(env, "${") {
			root = env
		} else if env := os.Getenv("WORKSPACE"); env != "" && !strings.HasPrefix(env, "${") {
			root = env
		} else if env := os.Getenv("PROJECT_ROOT"); env != "" && !strings.HasPrefix(env, "${") {
			root = env
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return ErrorResult(tool, fmt.Errorf("E-ROOT-UNRESOLVED: cannot determine working directory: %w", err),
					"Pass project_root explicitly with your target project directory path.")
			}
			root = cwd
		}
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("invalid path %q: %w", root, err),
			"Provide a valid absolute or relative path as project_root.")
	}

	// Guard against initializing at a filesystem root
	if absRoot == "/" || absRoot == "\\" || filepath.Dir(absRoot) == absRoot {
		return ErrorResult(tool, fmt.Errorf("project_root cannot be a filesystem root: %q", absRoot),
			"Provide a project directory path, not a filesystem root.")
	}

	// Guard against unguided initialization directly in user's root home directory (F-05, MED-06)
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if absHome, err := filepath.Abs(home); err == nil && filepath.Clean(absRoot) == filepath.Clean(absHome) {
			return ErrorResult(tool, fmt.Errorf("refusing to initialize workspace directly in user home directory %q", absRoot),
				"Provide a dedicated project subdirectory path as project_root, not the root home directory.")
		}
	}

	// Guard against unguided initialization in application or system directories (e.g. host IDE CWD inheritance)
	if core.IsSystemOrAppDir(root) || core.IsSystemOrAppDir(absRoot) {
		return ErrorResult(tool, fmt.Errorf("refusing to initialize workspace inside application or system directory %q", absRoot),
			"The host IDE process spawned Vivechak without setting an active workspace directory. Call vivechak_start with 'project_root' explicitly set to your project folder path (e.g., 'project_root': 'd:/path/to/project').")
	}

	root = absRoot

	rawScope := strings.ToLower(strings.TrimSpace(in.Scope))
	if rawScope == "" {
		if strings.TrimSpace(in.Question) != "" {
			rawScope = string(core.InferScope(in.Question))
		} else {
			rawScope = "project"
		}
	}
	scope, ok := core.ScopeFromString(rawScope)
	if !ok {
		return ErrorResult(tool, fmt.Errorf("invalid scope %q", in.Scope),
			"Provide 'scope': 'comparison' (comparing 2-5 options), 'decision' (single decision/ADR), or 'project' (full research pipeline).")
	}

	// Ensure root directory exists
	if err := os.MkdirAll(root, 0o755); err != nil {
		return ErrorResult(tool, fmt.Errorf("creating directory %q: %w", root, err),
			"Check filesystem permissions.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace %q: %w", root, err),
			"Ensure workspace directory is accessible.")
	}
	defer ws.Close()

	core.SetLastResolvedWorkspace(root)

	// Initialize workspace directory structure and copy scope templates
	dirs := []string{
		core.ResearchDir,
		core.SessionsDir,
		core.TemplatesDir,
	}
	for _, dir := range dirs {
		if err := ws.MkdirAll(dir, 0o755); err != nil {
			return ErrorResult(tool, fmt.Errorf("creating %s: %w", dir, err),
				"Check filesystem permissions and try again.")
		}
	}

	var tmplsCopied []string
	for _, tmpl := range core.TemplatesForScope(scope) {
		relPath := filepath.Join(core.TemplatesDir, tmpl)
		if !ws.FileExists(relPath) {
			if data, err := vembed.ReadTemplate(tmpl); err == nil {
				_ = store.WriteFileAtomic(ws.Root(), relPath, data, 0o644)
			}
		}
		tmplsCopied = append(tmplsCopied, tmpl)
	}

	gitignoreRel := filepath.Join(core.ResearchDir, ".gitignore")
	if !ws.FileExists(gitignoreRel) {
		_ = store.WriteFileAtomic(ws.Root(), gitignoreRel, []byte("# Vivechak operational files\n*.lock\n"), 0o644)
	}

	meta := core.WorkspaceMeta{
		Scope:     scope,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if metaBytes, err := json.Marshal(meta); err == nil {
		_ = store.WriteFileAtomic(ws.Root(), core.MetadataFile, metaBytes, 0o644)
	}

	question := strings.TrimSpace(in.Question)
	data := map[string]any{
		"workspace_root":   root,
		"scope":            scope,
		"initialized":      true,
		"templates_copied": tmplsCopied,
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

						newBlock := blockContent
						switch scope {
						case core.ScopeDecision:
							if strings.Contains(newBlock, placeholderDecision) {
								newBlock = strings.Replace(newBlock, placeholderDecision, "DECISION: "+question, 1)
							} else if strings.Contains(newBlock, "DECISION:") {
								re := regexp.MustCompile(`(?m)^DECISION:.*$`)
								newBlock = re.ReplaceAllString(newBlock, "DECISION: "+question)
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
						}

						if strings.Contains(newBlock, placeholderContextDecision) {
							newBlock = strings.Replace(newBlock, placeholderContextDecision, "CONTEXT: "+question, 1)
						} else if strings.Contains(newBlock, placeholderContextComparison) {
							newBlock = strings.Replace(newBlock, placeholderContextComparison, "CONTEXT: "+question, 1)
						} else if strings.Contains(newBlock, "CONTEXT:") {
							re := regexp.MustCompile(`(?m)^CONTEXT:.*$`)
							newBlock = re.ReplaceAllString(newBlock, "CONTEXT: "+question)
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
			return ErrorResult(tool, fmt.Errorf("invalid part %d: brief has %d total parts", in.Part, totalParts),
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
	data["prompt"] = activeBrief
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
			cid = "C-001"
		}
		data["id"] = cid

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
				"kind":    "comparison",
				"id":      cid,
				"content": "<matrix markdown>",
			},
			Why: fmt.Sprintf("Run web search to evaluate %s, then save the WEP comparison matrix using vivechak_save (or vivechak_save_plan(kind='comparison', id='%s', content=...)). Do NOT write to local artifact/brain files or external paths first.", optsStr, cid),
		}

	case core.ScopeDecision:
		did, allocErr := store.AllocateID(ws, "D")
		if allocErr != nil {
			did = "D-001"
		}
		data["id"] = did

		topic := question
		if topic == "" {
			topic = "Architectural Decision"
		}

		msg = fmt.Sprintf("Decision workspace initialized. Allocated ID %s for %s.", did, topic)
		next = &NextAction{
			Tool: "vivechak_record_decision",
			Args: map[string]any{
				"decision_id": did,
				"content":     "<ADR markdown>",
			},
			Why: fmt.Sprintf("Execute the brief with web search, then create the decision plan and ADR skeleton using vivechak_record_decision (or vivechak_save_plan(decision_id='%s', ...)). Do NOT write to local artifact/brain files or external paths first.", did),
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
					"scope": "project",
					"part":  nextPart,
				},
				Why: fmt.Sprintf("Retrieve part %d of %d of the project brief before generating the research pipeline DAG.", nextPart, totalParts),
			}
		} else {
			msg = fmt.Sprintf("Project workspace initialized for %s. All 6 templates ready.", projName)
			next = &NextAction{
				Tool: "vivechak_save",
				Args: map[string]any{
					"kind":    "plan",
					"content": "<RESEARCH-PIPELINE.md content>",
				},
				Why: "Decompose the project vision to generate your research pipeline DAG, then save via vivechak_save (or vivechak_save_plan(kind='plan', content=...)). Do NOT write to local artifact/brain files or workspace files outside research/. (For large Tier 2/3 pipelines where output exceeds single-turn token limits, you may assemble sections in internal scratch before passing to vivechak_save_plan).",
			}
		}
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  msg,
		Data:     data,
		Next:     next,
		NextStep: next.Why,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// handleInit provides backward compatibility for existing tests.
func handleInit(ctx context.Context, req *sdkmcp.CallToolRequest, in InitInput) (*sdkmcp.CallToolResult, Envelope, error) {
	return handleStart(ctx, req, StartInput{
		ProjectRoot: in.ProjectRoot,
		Scope:       in.Scope,
	})
}

// handlePrepareGenerator provides backward compatibility for existing tests.
func handlePrepareGenerator(ctx context.Context, req *sdkmcp.CallToolRequest, in PrepareGeneratorInput) (*sdkmcp.CallToolResult, Envelope, error) {
	return handleStart(ctx, req, StartInput{
		ProjectRoot: in.ProjectRoot,
		Scope:       in.Scope,
		Question:    in.Context,
	})
}

func extractComparisonOptions(ctx string) string {
	lower := strings.ToLower(ctx)
	delims := []string{" vs. ", " vs ", " versus "}
	for _, delim := range delims {
		if idx := strings.Index(lower, delim); idx != -1 {
			left := strings.TrimSpace(ctx[:idx])
			right := strings.TrimSpace(ctx[idx+len(delim):])
			prefixes := []string{
				"compare ", "evaluate ", "benchmark ", "choose between ", "decide between ",
				"which one is better: ", "which is better, ", "which is better: ",
			}
			for _, p := range prefixes {
				if strings.HasPrefix(strings.ToLower(left), p) {
					left = strings.TrimSpace(left[len(p):])
					break
				}
			}
			suffixes := []string{" for ", " in ", " with ", " when ", " under ", " on ", " as ", " to ", " across ", " over ", " within "}
			for _, s := range suffixes {
				if sIdx := strings.Index(strings.ToLower(right), s); sIdx != -1 {
					right = strings.TrimSpace(right[:sIdx])
					break
				}
			}

			// Normalize any additional intermediate comparison delimiters in right (e.g. A vs B vs C)
			for _, mid := range []string{" vs. ", " vs ", " versus ", " or "} {
				for {
					mIdx := strings.Index(strings.ToLower(right), mid)
					if mIdx == -1 {
						break
					}
					right = strings.TrimSpace(right[:mIdx]) + ", " + strings.TrimSpace(right[mIdx+len(mid):])
				}
			}

			left = strings.Trim(left, "?!.,:;")
			right = strings.Trim(right, "?!.,:;")
			if left != "" && right != "" {
				return left + ", " + right
			}
		}
	}
	return ""
}

// splitBriefIntoParts partitions a large brief into clean parts at markdown section boundaries.
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
