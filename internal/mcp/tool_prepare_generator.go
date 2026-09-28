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
	Scope       string `json:"scope,omitempty"         jsonschema:"research scope: project | decision | comparison (default: project)"`
	Context     string `json:"context"                 jsonschema:"project vision, decision context, or comparison context to inject into the generator prompt"`
}

func registerPrepareGenerator(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_prepare_generator",
			Title: "Prepare Generator Prompt",
			Description: "Return the appropriate generator prompt with context slots filled in, ready for " +
				"execution. Accepts scope (project/decision/comparison) and your context description. " +
				"The returned prompt should be executed in a fresh AI session with web search enabled. " +
				"Read-only — does not modify the workspace.",
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
	} else if in.ProjectRoot != "" {
		if ws, err := core.ResolveWorkspace(in.ProjectRoot); err == nil {
			info := core.InspectWorkspace(ws)
			if info.Scope != "" {
				scope = info.Scope
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
					if strings.Contains(newBlock, "[today]") {
						newBlock = strings.Replace(newBlock, "[today]", today, 1)
					}
					genPrompt = genPrompt[:afterStart] + newBlock + genPrompt[endIdx:]
				} else {
					// Otherwise, preserve schema keys and inject into CONTEXT slot, auto-filling DATE
					placeholderContextDecision := "CONTEXT: [what's being built; workload, team, constraints, what's already decided]"
					placeholderContextComparison := "CONTEXT: [what's being built; workload, team, constraints, stage]"

					newBlock := blockContent
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

	env := Envelope{
		Success: true,
		Message: fmt.Sprintf("Prepared %s generator prompt (%d chars, ~%d tokens)", scope, len(genPrompt), approxTokens),
		Data: map[string]any{
			"scope":          scope,
			"generator_file": genFile,
			"prompt":         genPrompt,
			"char_count":     len(genPrompt),
			"approx_tokens":  approxTokens,
		},
		NextStep: "Execute this prompt in a fresh AI session with web search enabled. " +
			"Save the output using vivechak_save_plan.",
		Meta: NewMeta(tool),
	}
	return env.ToResult()
}
