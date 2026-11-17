package mcputil

import (
	"context"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerPrompts registers the 4 canonical MCP prompts (no legacy prefix aliases).
func registerPrompts(server *sdkmcp.Server) {
	server.AddPrompt(
		&sdkmcp.Prompt{
			Name:        "compare",
			Title:       "Compare Options (WEP Matrix)",
			Description: "Compare 2-5 options into a WEP matrix.",
			Arguments: []*sdkmcp.PromptArgument{
				{Name: "options", Description: "2-5 options", Required: true},
				{Name: "criteria", Description: "Criteria", Required: false},
				{Name: "context", Description: "Context", Required: false},
			},
		},
		handlePromptCompare,
	)

	server.AddPrompt(
		&sdkmcp.Prompt{
			Name:        "decision",
			Title:       "Architectural Decision (ADR)",
			Description: "Resolve an architectural dilemma into an ADR.",
			Arguments: []*sdkmcp.PromptArgument{
				{Name: "question", Description: "Question", Required: true},
				{Name: "context", Description: "Context", Required: false},
			},
		},
		handlePromptDecision,
	)

	server.AddPrompt(
		&sdkmcp.Prompt{
			Name:        "project",
			Title:       "Project Pipeline (Full FAD)",
			Description: "Plan and research a system into a FAD.",
			Arguments: []*sdkmcp.PromptArgument{
				{Name: "vision", Description: "Vision", Required: true},
				{Name: "constraints", Description: "Constraints", Required: false},
			},
		},
		handlePromptProject,
	)

	server.AddPrompt(
		&sdkmcp.Prompt{
			Name:        "gate",
			Title:       "Phase 0 Pre-Codebase Exit Gate",
			Description: "Evaluate the Phase 0 pre-coding gate.",
			Arguments: []*sdkmcp.PromptArgument{
				{Name: "decision", Description: "D-ID", Required: false},
			},
		},
		handlePromptGate,
	)
}

func truncatePromptArg(s string, limit int) string {
	s = strings.TrimSpace(s)
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}

func handlePromptCompare(_ context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
	var options, criteria, contextStr string
	if req.Params != nil && req.Params.Arguments != nil {
		options = truncatePromptArg(req.Params.Arguments["options"], 140)
		criteria = truncatePromptArg(req.Params.Arguments["criteria"], 80)
		contextStr = truncatePromptArg(req.Params.Arguments["context"], 80)
	}
	if options == "" {
		options = "[2-5 options]"
	}
	q := options
	if contextStr != "" {
		q += " (" + contextStr + ")"
	}
	critArg := ""
	if criteria != "" {
		critArg = fmt.Sprintf(", criteria=%q", criteria)
	}
	text := fmt.Sprintf(
		"Compare %s using Vivechak:\n"+
			"1. Call vivechak_start(scope=\"comparison\", question=%q%s).\n"+
			"2. Execute the returned brief with live web search, grading every claim A-E.\n"+
			"3. Follow `next` in the response: save the WEP matrix with vivechak_save(kind=\"comparison\", id=<returned id>, content=<matrix>).",
		options, q, critArg,
	)
	return &sdkmcp.GetPromptResult{
		Description: fmt.Sprintf("Comparison workflow for %s", options),
		Messages: []*sdkmcp.PromptMessage{
			{Role: sdkmcp.Role("user"), Content: &sdkmcp.TextContent{Text: text}},
		},
	}, nil
}

func handlePromptDecision(_ context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
	var question, contextStr string
	if req.Params != nil && req.Params.Arguments != nil {
		question = truncatePromptArg(req.Params.Arguments["question"], 160)
		contextStr = truncatePromptArg(req.Params.Arguments["context"], 100)
	}
	if question == "" {
		question = "[architectural question]"
	}
	q := question
	if contextStr != "" {
		q += " — " + contextStr
	}
	text := fmt.Sprintf(
		"Decide %q using Vivechak:\n"+
			"1. Call vivechak_start(scope=\"decision\", question=%q); the server allocates the decision id.\n"+
			"2. Run the brief with web search, save the plan with vivechak_save(kind=\"plan\"), and follow `next` across vivechak_next, vivechak_save, vivechak_record_decision, and vivechak_gate.",
		question, q,
	)
	return &sdkmcp.GetPromptResult{
		Description: fmt.Sprintf("Decision workflow for %s", question),
		Messages: []*sdkmcp.PromptMessage{
			{Role: sdkmcp.Role("user"), Content: &sdkmcp.TextContent{Text: text}},
		},
	}, nil
}

func handlePromptProject(_ context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
	var vision, constraints string
	if req.Params != nil && req.Params.Arguments != nil {
		vision = truncatePromptArg(req.Params.Arguments["vision"], 160)
		constraints = truncatePromptArg(req.Params.Arguments["constraints"], 100)
	}
	if vision == "" {
		vision = "[project vision]"
	}
	q := vision
	if constraints != "" {
		q += " (constraints: " + constraints + ")"
	}
	text := fmt.Sprintf(
		"Architect %q using Vivechak:\n"+
			"1. Call vivechak_start(scope=\"project\", question=%q).\n"+
			"2. Decompose the vision using the returned brief and save research/RESEARCH-PIPELINE.md via vivechak_save(kind=\"plan\").\n"+
			"3. Follow `next` on each response: run vivechak_next with web search, save sessions, record ADRs, synthesize FAD.md, and run vivechak_gate.",
		vision, q,
	)
	return &sdkmcp.GetPromptResult{
		Description: "Project architecture workflow",
		Messages: []*sdkmcp.PromptMessage{
			{Role: sdkmcp.Role("user"), Content: &sdkmcp.TextContent{Text: text}},
		},
	}, nil
}

func handlePromptGate(_ context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
	var decision string
	if req.Params != nil && req.Params.Arguments != nil {
		decision = truncatePromptArg(req.Params.Arguments["decision"], 40)
	}
	callStr := "vivechak_gate()"
	if decision != "" {
		callStr = fmt.Sprintf("vivechak_gate(decision=%q)", decision)
	}
	text := fmt.Sprintf(
		"Run the Phase 0 Gate:\n"+
			"1. Call %s.\n"+
			"2. Inspect `data.verdict` and `data.rules` (Track A: G12; Track B: G1-G11).\n"+
			"3. Follow `next` to resolve any FAIL rules; for G11 ask the user to run `vck signoff <D-id>` in a terminal.",
		callStr,
	)
	return &sdkmcp.GetPromptResult{
		Description: "Phase 0 Exit Gate check",
		Messages: []*sdkmcp.PromptMessage{
			{Role: sdkmcp.Role("user"), Content: &sdkmcp.TextContent{Text: text}},
		},
	}, nil
}
