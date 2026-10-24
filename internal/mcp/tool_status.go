package mcputil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StatusInput holds the arguments for vivechak_status.
type StatusInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path (optional)"`
	Format      string `json:"format,omitempty"       jsonschema:"output format: 'summary' (default), 'mermaid' (visual flowchart), or 'json'"`
}

// VisualizeInput provides backward compatibility for unit tests calling handleVisualize.
type VisualizeInput struct {
	ProjectRoot string `json:"project_root,omitempty"`
	Root        string `json:"root,omitempty"`
	Format      string `json:"format,omitempty"`
}

func registerStatus(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_status",
			Title: "Vivechak Research Status & Visualization",
			Description: "Scan workspace and report research progress or render visual DAG flowcharts. " +
				"Format: 'summary' (default overview of scope, sessions, and decisions), " +
				"'mermaid' (renders session DAG flowchart or comparison matrix diagram), " +
				"or 'json' (structured metrics). Read-only.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleStatus,
	)
}

func handleStatus(_ context.Context, _ *sdkmcp.CallToolRequest, in StatusInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_status"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		env := Envelope{
			Ok:      true,
			Success: true,
			Message: "No workspace found. Vivechak is ready to initialize a new workspace.",
			Data: map[string]any{
				"initialized": false,
				"status":      "uninitialized",
			},
			Next: &NextAction{
				Tool: "vivechak_start",
				Why:  "Run vivechak_start with your project directory and scope to begin research.",
			},
			NextStep: "Run vivechak_start with your project directory and scope to begin research.",
			Meta:     NewMeta(tool),
		}
		return env.ToResult()
	}

	format := strings.ToLower(strings.TrimSpace(in.Format))
	if format == "mermaid" || format == "visualize" {
		return handleStatusRender(tool, root, "mermaid")
	}
	if format == "table" {
		return handleStatusRender(tool, root, "table")
	}
	if format != "" && format != "summary" {
		return ErrorResult(tool, fmt.Errorf("unsupported format %q: use 'summary', 'mermaid', or 'table'", in.Format),
			"Specify 'mermaid', 'table', or omit format for workspace summary.")
	}

	ws, wsErr := store.OpenWorkspace(root)
	if wsErr != nil {
		return ErrorResult(tool, wsErr, "Ensure workspace directory exists and is accessible.")
	}
	defer ws.Close()

	info := core.InspectWorkspace(root)
	data := map[string]any{
		"workspace_root": root,
		"root":           root,
		"initialized":    info.Initialized,
		"scope":          string(info.Scope),
		"has_pipeline":   info.HasPipeline,
		"has_decisions":  info.HasDecisions,
		"session_count":  info.SessionCount,
		"template_count": info.TemplateCount,
	}

	var next *NextAction
	var msg string
	if !info.Initialized {
		msg = "Workspace is uninitialized."
		next = &NextAction{
			Tool: "vivechak_start",
			Why:  "Run vivechak_start to begin research.",
		}
	} else if info.Scope == core.ScopeComparison && (info.HasPipeline || info.SessionCount >= 1) {
		msg = fmt.Sprintf("Workspace active at %s: comparison complete (%d sessions).", root, info.SessionCount)
		next = &NextAction{
			Tool: "vivechak_save",
			Why:  "Comparison scope is finished. Comparison research and WEP matrix are complete. No further sessions or gate checks required.",
		}
	} else if info.Scope == core.ScopeDecision && info.HasDecisions {
		msg = fmt.Sprintf("Workspace active at %s: decision locked (%d sessions).", root, info.SessionCount)
		next = &NextAction{
			Tool: "vivechak_record_decision",
			Why:  "Decision scope is complete. Decision research and ADR are finalized. Call vivechak_next if additional sessions remain.",
		}
	} else if !info.HasPipeline && info.SessionCount == 0 {
		msg = fmt.Sprintf("Workspace active at %s (Scope: %s).", root, info.Scope)
		next = &NextAction{
			Tool: "vivechak_save",
			Args: map[string]any{"kind": "plan"},
			Why:  "Research pipeline not yet saved. Save pipeline with vivechak_save(kind='plan').",
		}
	} else {
		hasFAD := false
		if _, err := os.Stat(filepath.Join(root, core.FADFile)); err == nil {
			hasFAD = true
		}
		msg = fmt.Sprintf("Workspace active at %s (%d sessions, %d templates).", root, info.SessionCount, info.TemplateCount)
		if hasFAD {
			next = &NextAction{
				Tool: "vivechak_gate",
				Why:  "Research synthesis is complete (FAD.md exists). Run Phase 0 Gate via vivechak_gate to verify pre-coding criteria.",
			}
		} else {
			next = &NextAction{
				Tool: "vivechak_next",
				Why:  "Call vivechak_next for the next actionable session, or vivechak_record_decision to record decisions from completed sessions.",
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

func handleStatusRender(tool string, root string, format string) (*sdkmcp.CallToolResult, Envelope, error) {
	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, err, "Ensure workspace directory exists.")
	}
	defer ws.Close()

	info := core.InspectWorkspace(root)
	if info.Scope == core.ScopeComparison {
		var compFile string
		var compData []byte
		if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				name := e.Name()
				if !e.IsDir() && strings.HasSuffix(name, ".md") &&
					(strings.HasPrefix(strings.ToUpper(name), "C-") ||
						strings.HasSuffix(name, "-comparison.md") ||
						strings.Contains(name, "-cmp-") ||
						strings.EqualFold(name, "COMPARISON.md")) {
					compFile = filepath.Join(core.ResearchDir, name)
					if d, rErr := ws.ReadFile(compFile); rErr == nil && len(d) > 0 {
						compData = d
						break
					}
				}
			}
		}
		if len(compData) == 0 {
			return ErrorResult(tool, fmt.Errorf("comparison file not found in %s", core.ResearchDir),
				"Save a comparison matrix first using vivechak_save(kind='comparison') before visualizing it.")
		}

		var rendered string
		if format == "table" {
			rendered = core.RenderComparisonTable(compData)
		} else {
			rendered = core.RenderComparisonMermaid(compData)
		}
		env := Envelope{
			Ok:      true,
			Success: true,
			Message: "Rendered comparison matrix visualization.",
			Data: map[string]any{
				"scope":    core.ScopeComparison,
				"format":   format,
				"rendered": rendered,
				"mermaid":  rendered,
				"table":    rendered,
				"file":     filepath.Base(compFile),
			},
			Next: &NextAction{
				Why: "Comparison visualization complete. Proceed to finalize decision or start implementation.",
			},
			NextStep: "Comparison visualization complete. Proceed to finalize decision or start implementation.",
			Meta:     NewMeta(tool),
		}
		return env.ToResult()
	}

	pipelineFile := core.PipelineFile
	pipelineData, err := ws.ReadFile(pipelineFile)
	if err != nil {
		found := false
		if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), "-plan.md") {
					pipelineFile = filepath.Join(core.ResearchDir, e.Name())
					pipelineData, err = ws.ReadFile(pipelineFile)
					if err == nil {
						found = true
						break
					}
				}
			}
		}
		if !found {
			return ErrorResult(tool, fmt.Errorf("pipeline file not found in %s", core.ResearchDir),
				"Create a research plan first using vivechak_save(kind='plan').")
		}
	}

	dag, err := core.ParsePipeline(pipelineData)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("parsing pipeline: %w", err), "Fix pipeline format.")
	}

	var knownIDs []string
	for _, s := range dag.Sessions {
		knownIDs = append(knownIDs, s.ID)
	}
	completedIDs, _ := scanCompletedSessions(ws, knownIDs...)

	mermaid := dag.ToMermaid(completedIDs)
	table := dag.ToStatusTable(completedIDs)

	var rendered string
	if format == "table" {
		rendered = table
	} else {
		rendered = mermaid
	}

	data := map[string]any{
		"workspace_root": root,
		"format":         format,
		"rendered":       rendered,
		"mermaid":        mermaid,
		"status_table":   table,
		"total_sessions": len(dag.Sessions),
		"completed":      len(completedIDs),
	}

	next := &NextAction{
		Tool: "vivechak_next",
		Why:  "Diagram rendered. Proceed with next ready session using vivechak_next.",
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Pipeline DAG rendered (%d sessions, %d completed).", len(dag.Sessions), len(completedIDs)),
		Data:     data,
		Next:     next,
		NextStep: next.Why,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// handleVisualize provides backward compatibility for unit tests calling handleVisualize.
func handleVisualize(_ context.Context, _ *sdkmcp.CallToolRequest, in VisualizeInput) (*sdkmcp.CallToolResult, Envelope, error) {
	root := in.ProjectRoot
	if root == "" {
		root = in.Root
	}
	format := in.Format
	if format == "" {
		format = "mermaid"
	}
	return handleStatus(context.Background(), nil, StatusInput{
		ProjectRoot: root,
		Format:      format,
	})
}
