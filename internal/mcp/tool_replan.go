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

// ReplanInput holds canonical arguments for vivechak_replan.
type ReplanInput struct {
	ProjectRoot  string `json:"project_root,omitempty"`
	Plan         string `json:"plan"`
	Op           string `json:"op"`
	SessionID    string `json:"session_id,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Layer        string `json:"layer,omitempty"`
	Topic        string `json:"topic,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Dependencies string `json:"dependencies,omitempty"`
	DoorType     string `json:"door_type,omitempty"`
	DecisionRef  string `json:"decision_ref,omitempty"`
}

func registerReplan(server *sdkmcp.Server) {
	registerTool(server, ToolSpec{
		Name:        "vivechak_replan",
		Description: "Edit a plan DAG in place (add_session, remove_session, update_deps, update_prompt); protects completed sessions.",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint:    false,
			IdempotentHint:  false,
			DestructiveHint: BoolPtr(true),
		},
		Params: []ParamSpec{
			{Name: "plan", Type: "string", Required: true, Description: "PROJECT or D-ID", Doc: "Target plan: PROJECT or decision ID (e.g. D-001)."},
			{Name: "op", Type: "string", Enum: []string{"add_session", "remove_session", "update_deps", "update_prompt"}, Required: true, Doc: "Replan operation to perform."},
			{Name: "session_id", Type: "string", Doc: "Target session ID (e.g. T1-02)."},
			{Name: "reason", Type: "string", Doc: "Rationale for the plan change."},
			{Name: "layer", Type: "string", Doc: "DAG layer number for add_session."},
			{Name: "topic", Type: "string", Doc: "Session title for add_session."},
			{Name: "prompt", Type: "string", Doc: "Session prompt text for add_session or update_prompt."},
			{Name: "dependencies", Type: "string", Description: "Comma-sep IDs", Doc: "Comma-separated dependency session IDs for add_session or update_deps."},
			{Name: "door_type", Type: "string", Enum: []string{"one-way", "two-way"}, Doc: "Door classification for add_session."},
			{Name: "decision_ref", Type: "string", Description: "Informs D-ID", Doc: "Informed decision ID for add_session."},
			{Name: "project_root", Type: "string", Doc: "Absolute path to the target project directory."},
		},
	}, handleReplan)
}

func resolvePlanFile(ss *sdkmcp.ServerSession, ws *store.Workspace, root, planArg string) (string, []string, error) {
	p := strings.TrimSpace(planArg)
	var notes []string
	if p == "" {
		if _, err := ws.Stat(core.PipelineFile); err == nil {
			return core.PipelineFile, nil, nil
		}
		if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), "-plan.md") {
					return filepath.Join(core.ResearchDir, e.Name()), nil, nil
				}
			}
		}
		return "", nil, fmt.Errorf("E-NOT-FOUND: pipeline file %s not found", core.PipelineFile)
	}
	if strings.EqualFold(p, "PROJECT") || strings.EqualFold(p, "DAG") {
		if _, err := ws.Stat(core.PipelineFile); err != nil {
			return "", nil, fmt.Errorf("E-NOT-FOUND: pipeline file %s not found", core.PipelineFile)
		}
		return core.PipelineFile, nil, nil
	}
	resolved, mapped, idErr := resolvePlaceholderRef(ss, ws, root, p, true)
	if idErr != nil {
		return "", nil, idErr
	}
	if mapped {
		notes = append(notes, fmt.Sprintf("W-ID-ALLOCATED: plan %s resolved to %s", p, resolved))
		p = resolved
	}
	norm := core.NormalizeDecisionID(p)
	upper := strings.ToUpper(norm)
	if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			nameUpper := strings.ToUpper(e.Name())
			if nameUpper == upper+"-PLAN.MD" || nameUpper == upper+".PLAN.MD" ||
				(strings.HasPrefix(nameUpper, upper+"-") && strings.HasSuffix(nameUpper, "-PLAN.MD")) {
				return filepath.Join(core.ResearchDir, e.Name()), notes, nil
			}
		}
	}
	return "", nil, fmt.Errorf("E-NOT-FOUND: plan %s not found in %s", p, core.ResearchDir)
}

func handleReplan(ctx context.Context, req *sdkmcp.CallToolRequest, in ReplanInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_replan"

	var roots []string
	if strings.TrimSpace(in.ProjectRoot) == "" {
		roots = getClientRoots(ctx, req)
	}
	root, rule, err := core.ResolveWorkspaceWithRoots(ctx, in.ProjectRoot, roots)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: %w", err), "Ensure the workspace directory exists.")
	}
	defer ws.Close()

	ss := sessionOf(req)
	pipelineFile, idWarnings, planErr := resolvePlanFile(ss, ws, root, in.Plan)
	if planErr != nil {
		return ErrorResult(tool, planErr, "Save a plan first using vivechak_save(kind=\"plan\") before mutating it.")
	}

	op := strings.ToLower(strings.TrimSpace(in.Op))
	if op == "" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: op is required"),
			"Specify op: 'add_session', 'remove_session', 'update_deps', or 'update_prompt'.")
	}

	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: session_id is required"),
			"Provide the target session ID in session_id.")
	}

	if resolved, mapped, idErr := resolvePlaceholderRef(ss, ws, root, sessionID, false); idErr != nil {
		return ErrorResult(tool, idErr, placeholderRefFix)
	} else if mapped {
		idWarnings = append(idWarnings, fmt.Sprintf("W-ID-ALLOCATED: session id %s resolved to %s", sessionID, resolved))
		sessionID = resolved
	}
	resolvedDeps, notes, idErr := resolvePlaceholderRefList(ss, ws, root, in.Dependencies)
	if idErr != nil {
		return ErrorResult(tool, idErr, placeholderRefFix)
	}
	in.Dependencies = resolvedDeps
	idWarnings = append(idWarnings, notes...)
	if in.DecisionRef != "" {
		if resolved, mapped, idErr := resolvePlaceholderRef(ss, ws, root, in.DecisionRef, true); idErr != nil {
			return ErrorResult(tool, idErr, placeholderRefFix)
		} else if mapped {
			idWarnings = append(idWarnings, fmt.Sprintf("W-ID-ALLOCATED: decision_ref %s resolved to %s", in.DecisionRef, resolved))
			in.DecisionRef = resolved
		}
	}

	unlock, lockErr := store.LockFile(ctx, filepath.Join(root, pipelineFile), 5*time.Second)
	if lockErr != nil {
		return ErrorResult(tool, fmt.Errorf("E-LOCK-TIMEOUT: acquiring lock on %s: %w", pipelineFile, lockErr), "Another process may be writing. Try again.")
	}
	defer func() { _ = unlock() }()

	pipelineData, err := ws.ReadFile(pipelineFile)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: reading pipeline file %s: %w", pipelineFile, err),
			"Check file permissions.")
	}

	dag, err := core.ParsePipeline(pipelineData)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("parsing pipeline: %w", err), planFixOr(err,
			"Ensure the pipeline markdown is formatted correctly."))
	}

	completedSessions, _ := dag.CompletedSessions(ws)
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
		if in.Dependencies != "" {
			for _, d := range strings.Split(in.Dependencies, ",") {
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
			return ErrorResult(tool, fmt.Errorf("E-SESSION-COMPLETED: cannot remove completed session %s", sessionID),
				"Completed sessions cannot be removed. Use vivechak_amend if findings have evolved.")
		}
		if err := dag.RemoveSession(sessionID); err != nil {
			return ErrorResult(tool, err, "Check if other sessions depend on this session.")
		}

	case "update_deps":
		if isCompleted {
			return ErrorResult(tool, fmt.Errorf("E-SESSION-COMPLETED: cannot update dependencies of completed session %s", sessionID),
				"Completed sessions cannot have their dependencies altered mid-flight.")
		}
		var deps []string
		if strings.TrimSpace(in.Dependencies) != "" {
			for _, d := range strings.Split(in.Dependencies, ",") {
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
			return ErrorResult(tool, fmt.Errorf("E-SESSION-COMPLETED: cannot update prompt of completed session %s", sessionID),
				"Session is already completed. Use vivechak_amend to add findings.")
		}
		if strings.TrimSpace(in.Prompt) == "" {
			return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: prompt content is required"),
				"Provide replacement prompt content in prompt.")
		}
		if err := dag.UpdatePrompt(sessionID, in.Prompt); err != nil {
			return ErrorResult(tool, err, "Verify the session ID exists in the pipeline.")
		}

	default:
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: unsupported op %q", op),
			"Supported operations: 'add_session', 'remove_session', 'update_deps', 'update_prompt'.")
	}

	replanParams := core.ReplanParams{
		SessionID:   sessionID,
		Topic:       in.Topic,
		DoorType:    in.DoorType,
		DecisionRef: in.DecisionRef,
		Prompt:      in.Prompt,
		Reason:      in.Reason,
	}
	if in.Dependencies != "" {
		for _, d := range strings.Split(in.Dependencies, ",") {
			d = strings.TrimSpace(d)
			if d != "" && !strings.EqualFold(d, "none") {
				replanParams.Dependencies = append(replanParams.Dependencies, d)
			}
		}
	}
	if s := dag.SessionByID(sessionID); op == "add_session" && s != nil {
		replanParams.Layer = s.Layer
	}

	newPipelineContent, err := core.ReplanInPlace(pipelineData, op, replanParams)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("replanning %s: %w", pipelineFile, err), planFixOr(err,
			"Check the session id and the plan's vivechak-plan block, then retry."))
	}

	if err := store.WriteFileAtomic(ws.Root(), pipelineFile, newPipelineContent, 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: saving mutated pipeline: %w", err),
			"Retry the operation.")
	}

	resData := map[string]any{
		"workspace":      map[string]any{"root": root, "rule": rule},
		"op":             op,
		"operation":      op,
		"session_id":     sessionID,
		"total_sessions": len(dag.Sessions),
		"pipeline_file":  filepath.ToSlash(pipelineFile),
	}
	if in.Reason != "" {
		resData["reason"] = in.Reason
	}

	nextWhy := "Plan updated. Call vivechak_next to retrieve the next ready session."
	env := Envelope{
		Ok:       true,
		Message:  fmt.Sprintf("Plan updated: %s %s (total sessions: %d).", op, sessionID, len(dag.Sessions)),
		Data:     resData,
		Warnings: ParseWarnings(idWarnings),
		Next: &NextAction{
			Tool: "vivechak_next",
			Args: map[string]any{},
			Why:  nextWhy,
		},
	}

	return env.ToResult()
}
