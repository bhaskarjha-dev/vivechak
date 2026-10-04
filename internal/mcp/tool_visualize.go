package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// VisualizeInput holds arguments for vivechak_visualize.
type VisualizeInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path (optional)"`
	Format      string `json:"format,omitempty"       jsonschema:"output format: 'mermaid' (default) or 'table'"`
}

func registerVisualize(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_visualize",
			Title: "Visualize Research Pipeline DAG",
			Description: "Render the active research pipeline DAG as a Mermaid flowchart or Markdown status table. " +
				"Shows session dependencies, completion status (green=completed, amber=ready, gray=blocked, blue=synthesis), " +
				"and parallel execution opportunities. Read-only.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleVisualize,
	)
}

func handleVisualize(ctx context.Context, _ *sdkmcp.CallToolRequest, in VisualizeInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_visualize"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_init.")
	}

	format := strings.ToLower(strings.TrimSpace(in.Format))
	if format == "" {
		format = "mermaid"
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, err, "Ensure the workspace directory exists.")
	}
	defer ws.Close()

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
				"Save a plan first using vivechak_save_plan before visualizing it.")
		}
	}

	dag, err := core.ParsePipeline(pipelineData)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("parsing pipeline: %w", err),
			"Ensure the pipeline markdown is formatted correctly.")
	}

	completedSessions, _ := scanCompletedSessions(ws)

	var rendered string
	switch format {
	case "mermaid":
		rendered = dag.ToMermaid(completedSessions)
	case "table":
		rendered = dag.ToStatusTable(completedSessions)
	default:
		return ErrorResult(tool, fmt.Errorf("unsupported format %q", format),
			"Supported formats: 'mermaid' (default) or 'table'.")
	}

	resData := map[string]any{
		"format":          format,
		"rendered":        rendered,
		"rendering":       rendered,
		format:            rendered,
		"total_sessions":  len(dag.Sessions),
		"completed_count": len(completedSessions),
	}

	nextStep := "Use vivechak_next_session to get the next ready research prompt, or vivechak_status to view project health."
	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Pipeline visualization rendered (%s format, %d sessions, %d completed)", format, len(dag.Sessions), len(completedSessions)),
		Data:     resData,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}

	return env.ToResult()
}
