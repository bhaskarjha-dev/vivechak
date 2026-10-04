package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ReplanInput holds arguments for vivechak_replan.
type ReplanInput struct {
	ProjectRoot  string `json:"project_root,omitempty"      jsonschema:"workspace root path (optional)"`
	Operation    string `json:"operation"                 jsonschema:"operation: 'add_session', 'remove_session', 'update_deps', 'update_prompt'"`
	SessionID    string `json:"session_id"                 jsonschema:"target session ID (e.g. T1-02)"`
	Layer        string `json:"layer,omitempty"            jsonschema:"for add_session: layer number (e.g. '1')"`
	Topic        string `json:"topic,omitempty"            jsonschema:"for add_session: session title/topic"`
	Prompt       string `json:"prompt,omitempty"           jsonschema:"for add_session: prompt markdown"`
	Dependencies string `json:"dependencies,omitempty"     jsonschema:"for add_session: comma-separated dependencies"`
	DoorType     string `json:"door_type,omitempty"        jsonschema:"for add_session: 'one-way' or 'two-way'"`
	DecisionRef  string `json:"decision_ref,omitempty"     jsonschema:"for add_session: informed decision ID (e.g. 'D-001')"`
	NewPrompt    string `json:"new_prompt,omitempty"       jsonschema:"for update_prompt: replacement prompt text"`
	NewDeps      string `json:"new_dependencies,omitempty" jsonschema:"for update_deps: replacement comma-separated dependencies"`
}

func registerReplan(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_replan",
			Title: "Replan Active Research Pipeline",
			Description: "Modify an active research pipeline DAG mid-flight. Operations: " +
				"'add_session' (add a new research session with dependencies), " +
				"'remove_session' (remove an uncompleted session), " +
				"'update_deps' (change dependencies), 'update_prompt' (replace prompt text). " +
				"Validates DAG integrity: no cycles, no orphaned dependencies, and protects " +
				"completed sessions from modification. Re-serializes RESEARCH-PIPELINE.md atomically.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleReplan,
	)
}

func handleReplan(ctx context.Context, _ *sdkmcp.CallToolRequest, in ReplanInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_replan"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_init.")
	}

	op := strings.ToLower(strings.TrimSpace(in.Operation))
	if op == "" {
		return ErrorResult(tool, fmt.Errorf("operation is required"),
			"Specify an operation: 'add_session', 'remove_session', 'update_deps', or 'update_prompt'.")
	}

	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" {
		return ErrorResult(tool, fmt.Errorf("session_id is required"),
			"Provide the target session ID.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, err, "Ensure the workspace directory exists.")
	}
	defer ws.Close()

	pipelineFile := core.PipelineFile
	pipelineData, err := ws.ReadFile(pipelineFile)
	if err != nil {
		// Check for decision or comparison plan files in research/
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
			return ErrorResult(tool, fmt.Errorf("pipeline file %s not found", core.PipelineFile),
				"Save a plan first using vivechak_save_plan before mutating it.")
		}
	}

	dag, err := core.ParsePipeline(pipelineData)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("parsing pipeline: %w", err),
			"Ensure the pipeline markdown is formatted correctly.")
	}

	completedSessions, _ := scanCompletedSessions(ws)
	isCompleted := completedSessions[sessionID] || completedSessions[strings.ToUpper(sessionID)] || completedSessions[strings.ToLower(sessionID)]

	switch op {
	case "add_session":
		layer := 1
		if strings.TrimSpace(in.Layer) != "" {
			if l, err := strconv.Atoi(strings.TrimSpace(in.Layer)); err == nil {
				layer = l
			}
		}

		var deps []string
		depStr := in.Dependencies
		if depStr != "" {
			for _, d := range strings.Split(depStr, ",") {
				d = strings.TrimSpace(d)
				if d != "" && !strings.EqualFold(d, "none") {
					deps = append(deps, d)
				}
			}
		}

		door := strings.ToLower(strings.TrimSpace(in.DoorType))
		if door == "" {
			door = "two-way"
		}

		topic := strings.TrimSpace(in.Topic)
		if topic == "" {
			topic = sessionID
		}

		prompt := strings.TrimSpace(in.Prompt)
		if prompt == "" {
			prompt = fmt.Sprintf("# %s: %s\n\nBRIEF:\nConduct research for %s.\n\nAPPROACH:\n- Investigate architectural options with evidence grades.\n", sessionID, topic, topic)
		}

		s := core.Session{
			ID:           sessionID,
			Title:        topic,
			Layer:        layer,
			DoorType:     door,
			DecisionRef:  strings.TrimSpace(in.DecisionRef),
			Dependencies: deps,
			OutputFile:   fmt.Sprintf("sessions/%s.md", sessionID),
			Prompt:       prompt,
		}

		if err := dag.AddSession(s); err != nil {
			return ErrorResult(tool, err, "Ensure dependencies exist in the pipeline and no cyclic dependencies are created.")
		}

	case "remove_session":
		if isCompleted {
			return ErrorResult(tool, fmt.Errorf("cannot remove completed session %s", sessionID),
				"Completed sessions cannot be removed. Use vivechak_amend_session if findings have evolved.")
		}
		if err := dag.RemoveSession(sessionID); err != nil {
			return ErrorResult(tool, err, "Check if other sessions depend on this session.")
		}

	case "update_deps":
		if isCompleted {
			return ErrorResult(tool, fmt.Errorf("cannot update dependencies of completed session %s", sessionID),
				"Completed sessions cannot have their dependencies altered mid-flight.")
		}
		rawDeps := in.NewDeps
		if rawDeps == "" {
			rawDeps = in.Dependencies
		}
		var deps []string
		if strings.TrimSpace(rawDeps) != "" {
			for _, d := range strings.Split(rawDeps, ",") {
				d = strings.TrimSpace(d)
				if d != "" && !strings.EqualFold(d, "none") {
					deps = append(deps, d)
				}
			}
		}
		if err := dag.UpdateDependencies(sessionID, deps); err != nil {
			return ErrorResult(tool, err, "Ensure all new dependencies exist and no cyclic dependencies are created.")
		}

	case "update_prompt":
		if isCompleted {
			return ErrorResult(tool, fmt.Errorf("cannot update prompt of completed session %s", sessionID),
				"Session is already completed. Use vivechak_amend_session to add findings.")
		}
		newPrompt := in.NewPrompt
		if newPrompt == "" {
			newPrompt = in.Prompt
		}
		if strings.TrimSpace(newPrompt) == "" {
			return ErrorResult(tool, fmt.Errorf("new_prompt content is required"),
				"Provide replacement prompt content.")
		}
		if err := dag.UpdatePrompt(sessionID, newPrompt); err != nil {
			return ErrorResult(tool, err, "Verify the session ID exists in the pipeline.")
		}

	default:
		return ErrorResult(tool, fmt.Errorf("unsupported operation %q", op),
			"Supported operations: 'add_session', 'remove_session', 'update_deps', 'update_prompt'.")
	}

	// Re-serialize and atomically save with lock
	newPipelineContent := dag.Serialize()
	unlock, lockErr := store.LockFile(ctx, filepath.Join(root, pipelineFile), 5*time.Second)
	if lockErr != nil {
		return ErrorResult(tool, fmt.Errorf("acquiring lock on %s: %w", pipelineFile, lockErr), "Another process may be writing. Try again.")
	}
	defer func() { _ = unlock() }()

	if err := store.WriteFileAtomic(ws.Root(), pipelineFile, newPipelineContent, 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("saving mutated pipeline: %w", err),
			"Retry the operation.")
	}

	resData := map[string]any{
		"operation":      op,
		"session_id":     sessionID,
		"total_sessions": len(dag.Sessions),
		"pipeline_file":  pipelineFile,
	}

	nextStep := "Pipeline updated. Run vivechak_next_session to get the next research prompt, or vivechak_visualize to inspect the DAG."
	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Pipeline successfully updated: %s %s (total sessions: %d)", op, sessionID, len(dag.Sessions)),
		Data:     resData,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}

	return env.ToResult()
}
