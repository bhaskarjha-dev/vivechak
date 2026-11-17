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

// StatusInput holds the canonical arguments for vivechak_status.
type StatusInput struct {
	ProjectRoot string `json:"project_root,omitempty"`
	Format      string `json:"format,omitempty"`
}

func registerStatus(server *sdkmcp.Server) {
	registerTool(server, ToolSpec{
		Name:        "vivechak_status",
		Description: "Show workspace status, sessions, decisions, and progress (summary, table, or mermaid). Read-only.",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
		Params: []ParamSpec{
			{Name: "format", Type: "string", Enum: []string{"summary", "table", "mermaid"}, Doc: "Output format: summary (default), table, or mermaid."},
			{Name: "project_root", Type: "string", Doc: "Absolute path to the target project directory."},
		},
	}, handleStatus)
}

func handleStatus(ctx context.Context, req *sdkmcp.CallToolRequest, in StatusInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_status"

	var roots []string
	if strings.TrimSpace(in.ProjectRoot) == "" {
		roots = getClientRoots(ctx, req)
	}
	root, rule, err := core.ResolveWorkspaceWithRoots(ctx, in.ProjectRoot, roots)
	if err != nil {
		if in.ProjectRoot == "" {
			return ErrorResult(tool, err, "Pass project_root=<absolute path of the user's project> explicitly.")
		}
		return ErrorResult(tool, err, "Pass a valid existing directory path in project_root.")
	}

	format := strings.ToLower(strings.TrimSpace(in.Format))
	if format != "" && format != "summary" && format != "table" && format != "mermaid" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: unsupported format %q: use 'summary', 'table', or 'mermaid'", in.Format),
			"Specify 'summary', 'table', or 'mermaid'.")
	}

	info := core.InspectWorkspace(root)
	if !info.Initialized {
		if format == "" {
			format = "summary"
		}
		env := Envelope{
			Ok:      true,
			Message: "Workspace is uninitialized.",
			Data: map[string]any{
				"workspace":      map[string]any{"root": root, "rule": rule},
				"workspace_root": root,
				"root":           root,
				"format":         format,
				"initialized":    false,
				"scope":          string(info.Scope),
				"has_pipeline":   false,
				"has_decisions":  false,
				"session_count":  0,
				"template_count": 0,
			},
			Next: &NextAction{
				Tool: "vivechak_start",
				Args: map[string]any{"scope": "comparison"},
				Why:  "Call vivechak_start with scope and question to initialize research.",
			},
		}
		return env.ToResult()
	}

	if format == "mermaid" || format == "table" {
		return handleStatusRender(tool, root, rule, format)
	}

	ws, wsErr := store.OpenWorkspace(root)
	if wsErr != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: %w", wsErr), "Ensure workspace directory exists and is accessible.")
	}
	defer ws.Close()

	data := map[string]any{
		"workspace":      map[string]any{"root": root, "rule": rule},
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
	switch {
	case info.Scope == core.ScopeComparison && (info.HasPipeline || info.SessionCount >= 1):
		msg = fmt.Sprintf("Workspace active at %s: comparison complete (%d sessions).", root, info.SessionCount)
		next = &NextAction{
			Tool: "vivechak_review",
			Args: map[string]any{"mode": "validate"},
			Why:  "Comparison scope is complete. Optionally run vivechak_review(mode=\"validate\") or start a decision.",
		}
	case info.Scope == core.ScopeDecision && info.HasDecisions:
		msg = fmt.Sprintf("Workspace active at %s: decision recorded (%d sessions).", root, info.SessionCount)
		next = &NextAction{
			Tool: "vivechak_gate",
			Args: map[string]any{},
			Why:  "Decision ADR is recorded. Call vivechak_gate to verify Phase 0 exit criteria.",
		}
	case !info.HasPipeline && info.SessionCount == 0:
		msg = fmt.Sprintf("Workspace active at %s (Scope: %s).", root, info.Scope)
		saveKind := "plan"
		if info.Scope == core.ScopeComparison {
			saveKind = "comparison"
		}
		next = &NextAction{
			Tool: "vivechak_save",
			Args: map[string]any{"kind": saveKind},
			Why:  fmt.Sprintf("Save the %s output with vivechak_save(kind=%q).", saveKind, saveKind),
		}
	default:
		hasFAD := false
		if _, err := os.Stat(filepath.Join(root, core.FADFile)); err == nil {
			hasFAD = true
		}
		msg = fmt.Sprintf("Workspace active at %s (%d sessions).", root, info.SessionCount)
		if hasFAD {
			next = &NextAction{
				Tool: "vivechak_gate",
				Args: map[string]any{},
				Why:  "FAD.md exists. Run Phase 0 Gate via vivechak_gate to verify pre-coding criteria.",
			}
		} else {
			next = &NextAction{
				Tool: "vivechak_next",
				Args: map[string]any{},
				Why:  "Call vivechak_next for the next ready session, or vivechak_record_decision to record ADRs.",
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

func handleStatusRender(tool, root, rule, format string) (*sdkmcp.CallToolResult, Envelope, error) {
	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: %w", err), "Ensure workspace directory exists.")
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
			return ErrorResult(tool, fmt.Errorf("E-NOT-FOUND: comparison file not found in %s", core.ResearchDir),
				"Save a comparison matrix first using vivechak_save(kind=\"comparison\").")
		}

		var rendered string
		if format == "table" {
			rendered = core.RenderComparisonTable(compData)
		} else {
			rendered = core.RenderComparisonMermaid(compData)
		}
		env := Envelope{
			Ok:      true,
			Message: "Rendered comparison matrix visualization.",
			Data: map[string]any{
				"workspace":      map[string]any{"root": root, "rule": rule},
				"workspace_root": root,
				"scope":          core.ScopeComparison,
				"format":         format,
				"rendered":       rendered,
				"mermaid":        rendered,
				"table":          rendered,
				"file":           filepath.Base(compFile),
			},
			Next: &NextAction{
				Tool: "vivechak_review",
				Args: map[string]any{"mode": "validate"},
				Why:  "Comparison visualization complete. Optionally run vivechak_review(mode=\"validate\").",
			},
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
			return ErrorResult(tool, fmt.Errorf("E-NOT-FOUND: pipeline file not found in %s", core.ResearchDir),
				"Create a research plan first using vivechak_save(kind=\"plan\").")
		}
	}

	dag, err := core.ParsePipeline(pipelineData)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("parsing pipeline: %w", err), planFixOr(err, "Fix pipeline format."))
	}

	completedIDs, _ := dag.CompletedSessions(ws)
	mermaid := dag.ToMermaid(completedIDs)
	table := dag.ToStatusTable(completedIDs)

	rendered := mermaid
	if format == "table" {
		rendered = table
	}

	data := map[string]any{
		"workspace":      map[string]any{"root": root, "rule": rule},
		"workspace_root": root,
		"format":         format,
		"rendered":       rendered,
		"mermaid":        mermaid,
		"status_table":   table,
		"total_sessions": len(dag.Sessions),
		"completed":      len(completedIDs),
	}

	env := Envelope{
		Ok:      true,
		Message: fmt.Sprintf("Pipeline DAG rendered (%d sessions, %d completed).", len(dag.Sessions), len(completedIDs)),
		Data:    data,
		Next: &NextAction{
			Tool: "vivechak_next",
			Args: map[string]any{},
			Why:  "Diagram rendered. Proceed with next ready session using vivechak_next.",
		},
	}
	return env.ToResult()
}
