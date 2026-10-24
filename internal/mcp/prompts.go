package mcputil

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	vembed "github.com/bhaskarjha-dev/vivechak/internal/embed"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// normalizePromptName strips legacy or shorthand prefixes from prompt names for backward compatibility.
func normalizePromptName(name string) string {
	trimmed := strings.TrimSpace(name)
	trimmed = strings.TrimPrefix(trimmed, "vivechak_")
	trimmed = strings.TrimPrefix(trimmed, "vck_")
	return trimmed
}

// registerPrompts registers the 4 native MCP prompt templates with the server.
func registerPrompts(server *sdkmcp.Server) {
	// Intercept prompts/get to transparently support legacy "vivechak_*" and "vck_*" aliases
	server.AddReceivingMiddleware(func(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
		return func(ctx context.Context, method string, req sdkmcp.Request) (sdkmcp.Result, error) {
			if method == "prompts/get" {
				if p, ok := req.GetParams().(*sdkmcp.GetPromptParams); ok && p != nil {
					p.Name = normalizePromptName(p.Name)
				}
			}
			return next(ctx, method, req)
		}
	})

	// 1. Compare Options Prompt
	server.AddPrompt(
		&sdkmcp.Prompt{
			Name:        "compare",
			Title:       "Compare Options (WEP Matrix)",
			Description: "Evaluate 2–5 options with an evidence-graded Weighted Evaluation Protocol (WEP) matrix (~5 minutes, 0 DAG, 0 gate).",
			Arguments: []*sdkmcp.PromptArgument{
				{
					Name:        "options",
					Title:       "Options to Compare",
					Description: "2–5 named options to evaluate (e.g. 'Pino vs Winston', 'ClickHouse vs TimescaleDB')",
					Required:    true,
				},
				{
					Name:        "criteria",
					Title:       "Evaluation Criteria",
					Description: "Optional key criteria or constraints (e.g. 'Throughput, memory footprint, ESM support')",
					Required:    false,
				},
				{
					Name:        "context",
					Title:       "Architectural Context",
					Description: "Optional workload, scale, team skills, or environment constraints",
					Required:    false,
				},
			},
		},
		handlePromptCompare,
	)

	// 2. Architectural Decision Prompt
	server.AddPrompt(
		&sdkmcp.Prompt{
			Name:        "decision",
			Title:       "Architectural Decision (ADR)",
			Description: "Resolve an architectural fork or migration in an active codebase into a locked ADR (~30 minutes, 1–3 sessions).",
			Arguments: []*sdkmcp.PromptArgument{
				{
					Name:        "dilemma",
					Title:       "Architectural Dilemma",
					Description: "The architectural fork to resolve (e.g. 'Migrate event storage from PostgreSQL to ClickHouse')",
					Required:    true,
				},
				{
					Name:        "context",
					Title:       "Production Context",
					Description: "Optional workload constraints, current stack, team skills, or SLAs",
					Required:    false,
				},
			},
		},
		handlePromptDecision,
	)

	// 3. Project Architecture Pipeline Prompt
	server.AddPrompt(
		&sdkmcp.Prompt{
			Name:        "project",
			Title:       "Project Pipeline (Full FAD)",
			Description: "Architect a greenfield system or major platform rewrite across an adaptive dependency DAG to a Founding Architecture Document (FAD).",
			Arguments: []*sdkmcp.PromptArgument{
				{
					Name:        "vision",
					Title:       "Project Vision",
					Description: "High-level system vision, functional requirements, expected scale, target stack",
					Required:    true,
				},
				{
					Name:        "context",
					Title:       "Constraints & Stack Preferences",
					Description: "Optional cloud provider, compliance (GDPR/HIPAA), team size, timeline",
					Required:    false,
				},
			},
		},
		handlePromptProject,
	)

	// 4. Phase 0 Exit Gate Prompt
	server.AddPrompt(
		&sdkmcp.Prompt{
			Name:        "gate",
			Title:       "Phase 0 Pre-Codebase Exit Gate",
			Description: "Evaluate Track A (reversibility) and Track B (checks B1–B11) pre-codebase exit criteria before coding.",
			Arguments: []*sdkmcp.PromptArgument{
				{
					Name:        "track",
					Title:       "Gate Track Filter",
					Description: "Evaluation track: 'all' (default), 'track_a', or 'track_b'",
					Required:    false,
				},
				{
					Name:        "project_root",
					Title:       "Project Root",
					Description: "Optional target workspace path (default: current workspace)",
					Required:    false,
				},
			},
		},
		handlePromptGate,
	)
}

func handlePromptCompare(_ context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
	var options, criteria, contextStr string
	if req.Params != nil && req.Params.Arguments != nil {
		options = strings.TrimSpace(req.Params.Arguments["options"])
		criteria = strings.TrimSpace(req.Params.Arguments["criteria"])
		contextStr = strings.TrimSpace(req.Params.Arguments["context"])
	}

	if options == "" {
		options = "[options to compare]"
	}
	if criteria == "" {
		criteria = "[criteria derived from context, normalized to 1.0]"
	}
	if contextStr == "" {
		contextStr = "Evaluating " + options + " for production adoption."
	}

	genBytes, err := vembed.ReadGenerator("GENERATOR-COMPARISON.md")
	if err != nil {
		return nil, fmt.Errorf("reading comparison generator: %w", err)
	}

	today := time.Now().UTC().Format("2006-01-02")
	promptText := string(genBytes)

	// Inject placeholders
	promptText = strings.Replace(promptText, "OPTIONS: [2-5 named options]", "OPTIONS: "+options, 1)
	promptText = strings.Replace(promptText, "CRITERIA: [optional; weights if known]", "CRITERIA: "+criteria, 1)
	promptText = strings.Replace(promptText, "CONTEXT: [what's being built, compared, or decided; workload, requirements, team, constraints, stage]", "CONTEXT: "+contextStr, 1)
	promptText = strings.Replace(promptText, "DATE: [today]", "DATE: "+today, 1)

	// Clarify execution directive: replace the meta-generator instruction with direct research execution
	metaInstruction := "From the context block, output a SCOPE CHECK line, then ONE research prompt in a `prompt` code fence: fill {slots} ({a | b} = pick one; {filename} = {ID}-cmp-{slug}.md), keep the rest verbatim. No plan, ADR, personas, search queries or search-count minimums."
	directInstruction := "Execute this evidence-calibrated research brief with live web search to compare the options below and produce the Weighted Evaluation Protocol (WEP) matrix. Do not output a prompt template; execute the comparison research directly."
	if strings.Contains(promptText, metaInstruction) {
		promptText = strings.Replace(promptText, metaInstruction, directInstruction, 1)
	} else {
		reMeta := regexp.MustCompile(`(?m)^From the context block, output a SCOPE CHECK line.*$`)
		promptText = reMeta.ReplaceAllString(promptText, directInstruction)
	}

	// Append anti-spill persistence directive
	promptText += "\n\n---\n**CRITICAL PERSISTENCE DIRECTIVE FOR AI AGENT:**\n" +
		"Execute web research to evaluate these options with inline evidence grades (A–E) and produce the Weighted Evaluation Protocol (WEP) matrix.\n" +
		"Once your research is complete, invoke the MCP tool `vivechak_save` (or `vivechak_save_plan`) with `kind=\"comparison\"` and `content` set to your complete WEP markdown document to persist it to disk at `research/C-NNN-comparison.md`."

	return &sdkmcp.GetPromptResult{
		Description: fmt.Sprintf("Evidence-calibrated comparison brief for: %s", options),
		Messages: []*sdkmcp.PromptMessage{
			{
				Role: sdkmcp.Role("user"),
				Content: &sdkmcp.TextContent{
					Text: promptText,
				},
			},
		},
	}, nil
}

func handlePromptDecision(_ context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
	var dilemma, contextStr string
	if req.Params != nil && req.Params.Arguments != nil {
		dilemma = strings.TrimSpace(req.Params.Arguments["dilemma"])
		contextStr = strings.TrimSpace(req.Params.Arguments["context"])
	}

	if dilemma == "" {
		dilemma = "[architectural fork or dilemma]"
	}
	if contextStr == "" {
		contextStr = "Resolving architectural fork for active codebase."
	}

	genBytes, err := vembed.ReadGenerator("GENERATOR-DECISION.md")
	if err != nil {
		return nil, fmt.Errorf("reading decision generator: %w", err)
	}

	today := time.Now().UTC().Format("2006-01-02")
	promptText := string(genBytes)

	promptText = strings.Replace(promptText, "DECISION: [the question, as you'd ask it]", "DECISION: "+dilemma, 1)
	promptText = strings.Replace(promptText, "CONTEXT: [what's being built or decided; project/venture context, workload, team, constraints, what's already decided]", "CONTEXT: "+contextStr, 1)
	promptText = strings.Replace(promptText, "DATE: [today]", "DATE: "+today, 1)

	promptText += "\n\n---\n**CRITICAL PERSISTENCE DIRECTIVE FOR AI AGENT:**\n" +
		"Output the decision research plan and proposed ADR. Once drafted, call the MCP tool `vivechak_save` (or `vivechak_save_plan`) with `kind=\"plan\"` and `content` set to your output to save the plan, or invoke `vivechak_record_decision` to lock the final ADR."

	return &sdkmcp.GetPromptResult{
		Description: fmt.Sprintf("Architectural decision research plan for: %s", dilemma),
		Messages: []*sdkmcp.PromptMessage{
			{
				Role: sdkmcp.Role("user"),
				Content: &sdkmcp.TextContent{
					Text: promptText,
				},
			},
		},
	}, nil
}

func handlePromptProject(_ context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
	var vision, contextStr string
	if req.Params != nil && req.Params.Arguments != nil {
		vision = strings.TrimSpace(req.Params.Arguments["vision"])
		contextStr = strings.TrimSpace(req.Params.Arguments["context"])
	}

	if vision == "" {
		vision = "[system architecture vision and requirements]"
	}
	if contextStr != "" {
		vision = vision + "\n\nConstraints & Preferences:\n" + contextStr
	}

	genBytes, err := vembed.ReadGenerator("GENERATOR-PROJECT.md")
	if err != nil {
		return nil, fmt.Errorf("reading project generator: %w", err)
	}

	genPrompt := string(genBytes)

	// Extract inner prompt from ````markdown code fences if present
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
		genPrompt = strings.Replace(genPrompt, placeholder, vision, 1)
	} else {
		genPrompt = genPrompt + "\n\n## Context\n\n" + vision
	}

	genPrompt += "\n\n---\n**CRITICAL PERSISTENCE DIRECTIVE FOR AI AGENT:**\n" +
		"Decompose this project architecture into a structured research pipeline DAG.\n" +
		"Once synthesized, call the MCP tool `vivechak_save` (or `vivechak_save_plan`) with `kind=\"plan\"` and `content` (or `pipeline_content`) set to your complete RESEARCH-PIPELINE.md document to persist it into `research/`."

	return &sdkmcp.GetPromptResult{
		Description: "Greenfield system architecture decomposition brief",
		Messages: []*sdkmcp.PromptMessage{
			{
				Role: sdkmcp.Role("user"),
				Content: &sdkmcp.TextContent{
					Text: genPrompt,
				},
			},
		},
	}, nil
}

func handlePromptGate(_ context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
	track := "all"
	projectRoot := ""
	if req.Params != nil && req.Params.Arguments != nil {
		if t := strings.TrimSpace(req.Params.Arguments["track"]); t != "" {
			track = strings.ToLower(t)
		}
		projectRoot = strings.TrimSpace(req.Params.Arguments["project_root"])
	}

	promptText := fmt.Sprintf(
		"# Phase 0 Pre-Codebase Exit Gate Audit\n\n"+
			"Please audit the current research workspace against the Vivechak Phase 0 Gate criteria before writing any codebase implementation.\n\n"+
			"**Instructions for AI Agent:**\n"+
			"1. Invoke the MCP tool `vivechak_gate` (or `vivechak_run_gate`) with `track=%q`", track)

	if projectRoot != "" {
		promptText += fmt.Sprintf(" and `project_root=%q`", projectRoot)
	}
	promptText += ".\n" +
		"2. Review the resulting `track_a_result` (Reversibility confirmation for Two-Way Doors) and `track_b_result` (checks B1–B11 for One-Way Doors).\n" +
		"3. Summarize all passing checks, and for any failed checks, provide specific recommendations on how to remediate them before starting implementation.\n"

	return &sdkmcp.GetPromptResult{
		Description: fmt.Sprintf("Phase 0 Exit Gate evaluation (track: %s)", track),
		Messages: []*sdkmcp.PromptMessage{
			{
				Role: sdkmcp.Role("user"),
				Content: &sdkmcp.TextContent{
					Text: promptText,
				},
			},
		},
	}, nil
}
