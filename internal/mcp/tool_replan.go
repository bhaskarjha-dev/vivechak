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
	Operation    string `json:"operation,omitempty"        jsonschema:"operation: 'add_session', 'remove_session', 'update_deps', 'update_prompt' (optional for guided overview)"`
	SessionID    string `json:"session_id,omitempty"       jsonschema:"target session ID (e.g. T1-02)"`
	Reason       string `json:"reason,omitempty"           jsonschema:"rationale or finding prompting the replan"`
	Impact       string `json:"impact,omitempty"           jsonschema:"intended impact on DAG topology or affected sessions"`
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

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, err, "Ensure the workspace directory exists.")
	}
	defer ws.Close()

	pipelineFile := core.PipelineFile
	if _, statErr := ws.Stat(pipelineFile); statErr != nil {
		found := false
		if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), "-plan.md") {
					pipelineFile = filepath.Join(core.ResearchDir, e.Name())
					found = true
					break
				}
			}
		}
		if !found {
			return ErrorResult(tool, fmt.Errorf("pipeline file %s not found", core.PipelineFile),
				"Save a plan first using vivechak_save_plan before mutating it.")
		}
	}

	op := strings.ToLower(strings.TrimSpace(in.Operation))
	if op == "" {
		data, err := ws.ReadFile(pipelineFile)
		if err != nil {
			return ErrorResult(tool, fmt.Errorf("reading pipeline: %w", err), "Verify pipeline file exists.")
		}
		dag, err := core.ParsePipeline(data)
		if err != nil {
			return ErrorResult(tool, fmt.Errorf("parsing pipeline: %w", err), "Pipeline file is malformed.")
		}

		knownIDs := make([]string, 0, len(dag.Sessions))
		for _, s := range dag.Sessions {
			knownIDs = append(knownIDs, s.ID)
		}
		completedMap, _ := core.ScanCompletedSessions(ws, knownIDs...)
		var candidates []string
		for _, s := range dag.Sessions {
			if !completedMap[s.ID] {
				candidates = append(candidates, fmt.Sprintf("%s (%s)", s.ID, s.Title))
			}
		}

		env := Envelope{
			Success: true,
			Message: "To replan, specify the operation ('add_session', 'remove_session', 'update_deps', or 'update_prompt') and target session_id.",
			Data: map[string]any{
				"status":               "guided",
				"reason":               in.Reason,
				"impact":               in.Impact,
				"available_operations": []string{"add_session", "remove_session", "update_deps", "update_prompt"},
				"uncompleted_sessions": candidates,
				"pipeline_file":        pipelineFile,
			},
			NextStep: fmt.Sprintf("Call vivechak_replan with operation ('update_prompt', 'remove_session', 'add_session', or 'update_deps') and session_id (uncompleted candidates: %s).", strings.Join(candidates, ", ")),
			Meta:     NewMeta(tool),
		}
		return env.ToResult()
	}

	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" {
		return ErrorResult(tool, fmt.Errorf("session_id is required"),
			"Provide the target session ID.")
	}

	// Acquire lock eagerly before reading and mutating pipeline to prevent TOCTOU races
	unlock, lockErr := store.LockFile(ctx, filepath.Join(root, pipelineFile), 5*time.Second)
	if lockErr != nil {
		return ErrorResult(tool, fmt.Errorf("acquiring lock on %s: %w", pipelineFile, lockErr), "Another process may be writing. Try again.")
	}
	defer func() { _ = unlock() }()

	pipelineData, err := ws.ReadFile(pipelineFile)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("reading pipeline file %s: %w", pipelineFile, err),
			"Check file permissions.")
	}

	dag, err := core.ParsePipeline(pipelineData)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("parsing pipeline: %w", err),
			"Ensure the pipeline markdown is formatted correctly.")
	}

	var knownIDs []string
	for _, s := range dag.Sessions {
		knownIDs = append(knownIDs, s.ID)
	}
	completedSessions, _ := scanCompletedSessions(ws, knownIDs...)
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

	// In-place replan if machine block is present (BUG-11 fix), else fallback to Serialize
	var newPipelineContent []byte
	replanParams := core.ReplanParams{
		SessionID:   sessionID,
		Topic:       in.Topic,
		DoorType:    in.DoorType,
		DecisionRef: in.DecisionRef,
		Prompt:      in.Prompt,
		Reason:      in.Reason,
		Impact:      in.Impact,
	}
	rawDeps := in.Dependencies
	if in.NewDeps != "" {
		rawDeps = in.NewDeps
	}
	if rawDeps != "" {
		for _, d := range strings.Split(rawDeps, ",") {
			d = strings.TrimSpace(d)
			if d != "" && !strings.EqualFold(d, "none") {
				replanParams.Dependencies = append(replanParams.Dependencies, d)
			}
		}
	}
	if in.NewPrompt != "" {
		replanParams.Prompt = in.NewPrompt
	}
	if in.Layer != "" {
		if l, err := strconv.Atoi(strings.TrimSpace(in.Layer)); err == nil {
			replanParams.Layer = l
		}
	}

	if inPlaceBytes, err := core.ReplanInPlace(pipelineData, op, replanParams); err == nil && len(inPlaceBytes) > 0 {
		newPipelineContent = inPlaceBytes
	} else {
		newPipelineContent = dag.Serialize()
	}

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
	if in.Reason != "" {
		resData["reason"] = in.Reason
	}
	if in.Impact != "" {
		resData["impact"] = in.Impact
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
