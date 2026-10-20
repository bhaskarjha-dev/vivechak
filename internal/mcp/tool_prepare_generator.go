package mcputil

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	vembed "github.com/bhaskarjha-dev/vivechak/internal/embed"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// PrepareGeneratorInput holds the arguments for vivechak_prepare_generator.
type PrepareGeneratorInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path (optional; uses resolution chain if omitted)"`
	Scope       string `json:"scope,omitempty"        jsonschema:"research scope (REQUIRED): 'comparison' (comparing 2+ options, models, algorithms, vendors, strategies) | 'decision' (single choice or dilemma with trade-offs/ADR) | 'project' (full system architecture, new venture, or comprehensive research pipeline)"`
	Context     string `json:"context"                jsonschema:"the user's question, options being compared, decision dilemma, or project vision. Required."`
}

func registerPrepareGenerator(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_prepare_generator",
			Title: "Prepare Generator Prompt",
			Description: "MANDATORY research entrypoint: Call this tool whenever the user asks to compare " +
				"two or more options, alternatives, approaches, technologies, models, algorithms, or strategies (e.g., 'A vs B', 'which should I choose'), " +
				"or asks for an evaluative decision, trade-off analysis, new project research, or system architecture plan. " +
				"Do NOT answer comparisons, option evaluations, decisions, or system planning from ungrounded memory — ALWAYS invoke this tool " +
				"to prepare an evidence-calibrated research brief. Skip for direct code implementation, debugging, or syntax lookup. " +
				"Requires 'scope' ('comparison', 'decision', or 'project') " +
				"and the user's question, options, decision dilemma, or project vision as 'context'. Read-only — does not modify workspace.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handlePrepareGenerator,
	)
}

func handlePrepareGenerator(_ context.Context, _ *sdkmcp.CallToolRequest, in PrepareGeneratorInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_prepare_generator"

	// Parse scope
	scope := core.ScopeProject
	if in.Scope != "" {
		s, ok := core.ScopeFromString(in.Scope)
		if !ok {
			return ErrorResult(tool, fmt.Errorf("invalid scope %q", in.Scope),
				"Use scope 'project', 'decision', or 'comparison'.")
		}
		scope = s
	} else {
		// Internal fallback if caller omitted scope
		if ws, err := core.ResolveWorkspace(in.ProjectRoot); err == nil {
			info := core.InspectWorkspace(ws)
			if info.Scope != "" {
				scope = info.Scope
			}
		}
		if scope == core.ScopeProject && strings.TrimSpace(in.Context) != "" {
			if inf := core.InferScope(in.Context); inf != core.ScopeProject {
				scope = inf
			}
		}
	}

	// Validate context
	if strings.TrimSpace(in.Context) == "" {
		return ErrorResult(tool,
			fmt.Errorf("context is required — provide your %s description", scope),
			fmt.Sprintf("Call vivechak_prepare_generator again with the 'context' field filled in with your %s description.", scope))
	}

	// Read the generator template
	genFile := core.GeneratorFile(scope)
	genBytes, err := vembed.ReadGenerator(genFile)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("reading generator %s: %w", genFile, err),
			"This is an internal error — the generator file should be embedded in the binary.")
	}

	// Inject context into the generator prompt.
	// NOTE: in.Context is injected via string replacement. This is safe because the
	// output is returned as text for human/agent review, not executed. If this function
	// is ever extended to directly execute the prompt, sanitization would be required.
	genPrompt := string(genBytes)
	switch scope {
	case core.ScopeProject:
		// Extract inner prompt from ````markdown code fences if present
		if startIdx := strings.Index(genPrompt, "````markdown"); startIdx >= 0 {
			afterFence := startIdx + len("````markdown")
			for afterFence < len(genPrompt) && genPrompt[afterFence] != '\n' {
				afterFence++
			}
			if afterFence < len(genPrompt) {
				afterFence++ // skip '\n'
			}
			if endIdx := strings.LastIndex(genPrompt[afterFence:], "````"); endIdx >= 0 {
				genPrompt = strings.TrimSpace(genPrompt[afterFence : afterFence+endIdx])
			}
		}

		placeholder := "[PASTE YOUR PROJECT DESCRIPTION HERE]"
		if strings.Contains(genPrompt, placeholder) {
			genPrompt = strings.Replace(genPrompt, placeholder, in.Context, 1)
		} else {
			genPrompt = genPrompt + "\n\n## Context\n\n" + in.Context
		}
	case core.ScopeDecision, core.ScopeComparison:
		// Decision and comparison generators use a ```context block
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

				// If input context already provides structured key-value fields, use it directly
				if strings.Contains(in.Context, "DECISION:") || strings.Contains(in.Context, "OPTIONS:") || strings.Contains(in.Context, "CONTEXT:") {
					newBlock := "\n" + strings.TrimSpace(in.Context) + "\n"
					newBlock = strings.Replace(newBlock, "[today]", today, 1)
					genPrompt = genPrompt[:afterStart] + newBlock + genPrompt[endIdx:]
				} else {
					// Otherwise, preserve schema keys and inject into DECISION and CONTEXT slots, auto-filling DATE
					placeholderDecision := "DECISION: [the question, as you'd ask it]"
					placeholderContextDecision := "CONTEXT: [what's being built; workload, team, constraints, what's already decided]"
					placeholderContextComparison := "CONTEXT: [what's being built, compared, or decided; workload, requirements, team, constraints, stage]"
					placeholderOptions := "OPTIONS: [2-5 named options]"

					newBlock := blockContent
					switch scope {
					case core.ScopeDecision:
						if strings.Contains(newBlock, placeholderDecision) {
							newBlock = strings.Replace(newBlock, placeholderDecision, "DECISION: "+in.Context, 1)
						} else if strings.Contains(newBlock, "DECISION:") {
							re := regexp.MustCompile(`(?m)^DECISION:.*$`)
							newBlock = re.ReplaceAllString(newBlock, "DECISION: "+in.Context)
						}
					case core.ScopeComparison:
						if strings.Contains(newBlock, placeholderOptions) {
							if opts := extractComparisonOptions(in.Context); opts != "" {
								newBlock = strings.Replace(newBlock, placeholderOptions, "OPTIONS: "+opts, 1)
							}
						}
					}

					if strings.Contains(newBlock, placeholderContextDecision) {
						newBlock = strings.Replace(newBlock, placeholderContextDecision, "CONTEXT: "+in.Context, 1)
					} else if strings.Contains(newBlock, placeholderContextComparison) {
						newBlock = strings.Replace(newBlock, placeholderContextComparison, "CONTEXT: "+in.Context, 1)
					} else if strings.Contains(newBlock, "CONTEXT:") {
						re := regexp.MustCompile(`(?m)^CONTEXT:.*$`)
						newBlock = re.ReplaceAllString(newBlock, "CONTEXT: "+in.Context)
					}
					newBlock = strings.Replace(newBlock, "[today]", today, 1)
					genPrompt = genPrompt[:afterStart] + newBlock + genPrompt[endIdx:]
				}
			}
		} else {
			genPrompt = genPrompt + "\n\n## Context\n\n" + in.Context
		}
	}

	// Calculate approximate token count for the prompt
	approxTokens := len(genPrompt) / 4

	msg := fmt.Sprintf("Prepared %s generator prompt (%d chars, ~%d tokens)", scope, len(genPrompt), approxTokens)

	dataMap := map[string]any{
		"scope":          scope,
		"generator_file": genFile,
		"prompt":         genPrompt,
		"char_count":     len(genPrompt),
		"approx_tokens":  approxTokens,
	}

	var nextStep string
	switch scope {
	case core.ScopeComparison:
		nextStep = "Execute this prompt in a research session with web search. " +
			"Save the resulting WEP comparison matrix using vivechak_save_plan."
	case core.ScopeDecision:
		nextStep = "Execute this prompt in a research session with web search. " +
			"Save the decision plan and ADR skeleton using vivechak_save_plan."
	default:
		nextStep = "Execute the generated prompt (either directly in your next turn or by delegating to a research subagent) to generate your research pipeline. " +
			"Save the output using vivechak_save_plan."
	}

	env := Envelope{
		Success:  true,
		Message:  msg,
		Data:     dataMap,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// extractComparisonOptions attempts to extract 2-5 candidate options from comparison prompts like "Pino vs Winston"
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
			suffixes := []string{" for ", " in ", " with ", " when ", " under "}
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

