package mcputil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NextSessionInput holds the arguments for vivechak_next_session.
type NextSessionInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path"`
	SessionID   string `json:"session_id,omitempty"    jsonschema:"specific session ID to retrieve (optional; returns next actionable if omitted)"`
	Verbose     bool   `json:"verbose,omitempty"        jsonschema:"if true, include full upstream findings in response"`
}

func registerNextSession(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_next_session",
			Title: "Get Next Session",
			Description: "Return the next actionable research session prompt with upstream findings " +
				"injected into context slots. Parses the research pipeline DAG to determine session " +
				"ordering and dependency satisfaction. For SYN-01 (synthesis), returns the synthesis " +
				"prompt with ALL session findings aggregated. This is the core value proposition — " +
				"context injection eliminates manual copy-paste between sessions. Read-only.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleNextSession,
	)
}

func handleNextSession(_ context.Context, _ *sdkmcp.CallToolRequest, in NextSessionInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_next_session"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err,
			"Initialize a workspace first with vivechak_init.")
	}

	if !core.WorkspaceExists(root) {
		return ErrorResult(tool, fmt.Errorf("workspace not initialized at %s", root),
			"Run vivechak_init first.")
	}

	// Read pipeline file
	pipelinePath := filepath.Join(root, core.PipelineFile)
	pipelineData, err := os.ReadFile(pipelinePath)
	if err != nil {
		return ErrorResult(tool,
			fmt.Errorf("no research pipeline found — run vivechak_save_plan first"),
			"Run vivechak_prepare_generator to get the generator prompt, execute it, "+
				"then save with vivechak_save_plan.")
	}

	// Parse pipeline into DAG
	dag, err := core.ParsePipeline(pipelineData)
	if err != nil {
		return ErrorResult(tool,
			fmt.Errorf("failed to parse pipeline: %w", err),
			"The RESEARCH-PIPELINE.md may be malformed. Check its structure and try again.")
	}

	// Scan completed sessions
	sessionsDir := filepath.Join(root, core.SessionsDir)
	completedIDs, err := scanCompletedSessions(sessionsDir)
	if err != nil {
		return ErrorResult(tool, err, "Check workspace directory permissions.")
	}

	// If a specific session is requested, return it directly
	if in.SessionID != "" {
		session := dag.SessionByID(in.SessionID)
		if session == nil {
			return ErrorResult(tool,
				fmt.Errorf("session %q not found in pipeline", in.SessionID),
				fmt.Sprintf("Available sessions: %s", sessionIDList(dag)))
		}

		injected, err := core.InjectContext(*session, sessionsDir, completedIDs)
		if err != nil {
			// Session found but no prompt — return what we have
			return buildSessionResponse(tool, *session, "", completedIDs, dag, in.Verbose)
		}
		return buildSessionResponse(tool, injected.Session, injected.InjectedPrompt, completedIDs, dag, in.Verbose)
	}

	// Find next actionable sessions
	ready := dag.NextSessions(completedIDs)

	if len(ready) == 0 {
		// All sessions complete
		if len(completedIDs) == len(dag.Sessions) {
			env := Envelope{
				Success: true,
				Message: fmt.Sprintf("All %d sessions complete!", len(dag.Sessions)),
				Data: map[string]any{
					"workspace_root":     root,
					"total_sessions":     len(dag.Sessions),
					"completed_sessions": len(completedIDs),
					"all_complete":       true,
				},
				NextStep: "All research sessions are complete. Run vivechak_run_gate " +
					"to verify the Phase 0 exit criteria, or vivechak_record_decision " +
					"to record any outstanding decisions.",
				Meta: NewMeta(tool),
			}
			return env.ToResult()
		}

		// Some sessions remain but none are actionable (blocked by dependencies)
		var blockedInfo []map[string]any
		for _, s := range dag.Sessions {
			if completedIDs[s.ID] {
				continue
			}
			var missingDeps []string
			for _, dep := range s.Dependencies {
				if !completedIDs[dep] {
					missingDeps = append(missingDeps, dep)
				}
			}
			blockedInfo = append(blockedInfo, map[string]any{
				"session_id":   s.ID,
				"blocked_by":   missingDeps,
			})
		}

		env := Envelope{
			Success: true,
			Message: fmt.Sprintf("%d sessions remaining but all blocked by dependencies", len(dag.Sessions)-len(completedIDs)),
			Data: map[string]any{
				"workspace_root":     root,
				"total_sessions":     len(dag.Sessions),
				"completed_sessions": len(completedIDs),
				"blocked_sessions":   blockedInfo,
			},
			NextStep: "Some sessions are blocked by incomplete dependencies. " +
				"Complete the blocking sessions first.",
			Meta: NewMeta(tool),
		}
		return env.ToResult()
	}

	// Return the first ready session with context injected
	nextSession := ready[0]
	injected, err := core.InjectContext(nextSession, sessionsDir, completedIDs)

	var prompt string
	if err == nil && injected != nil {
		prompt = injected.InjectedPrompt
	} else {
		prompt = nextSession.Prompt
	}

	// Build response with all ready sessions listed
	var otherReady []string
	for _, s := range ready[1:] {
		otherReady = append(otherReady, s.ID)
	}

	return buildSessionResponse(tool, nextSession, prompt, completedIDs, dag, in.Verbose, otherReady...)
}

// buildSessionResponse creates the response envelope for a session.
func buildSessionResponse(tool string, session core.Session, prompt string, completedIDs map[string]bool, dag *core.DAG, verbose bool, otherReady ...string) (*sdkmcp.CallToolResult, Envelope, error) {
	data := map[string]any{
		"session_id":          session.ID,
		"title":               session.Title,
		"layer":               session.Layer,
		"door_type":           session.DoorType,
		"decision_ref":        session.DecisionRef,
		"dependencies":        session.Dependencies,
		"output_file":         session.OutputFile,
		"total_sessions":      len(dag.Sessions),
		"completed_sessions":  len(completedIDs),
		"already_completed":   completedIDs[session.ID],
	}

	if prompt != "" {
		data["prompt"] = prompt
		data["prompt_char_count"] = len(prompt)
		data["prompt_approx_tokens"] = len(prompt) / 4
	}

	if len(otherReady) > 0 {
		data["other_ready_sessions"] = otherReady
	}

	var warnings []string

	// Progressive disclosure: warn if prompt is very large
	approxTokens := len(prompt) / 4
	truncated := false
	if !verbose && approxTokens > 10000 {
		warnings = append(warnings, "W-OUTPUT-SIZE: Prompt exceeds 10K tokens and was truncated. Use verbose=true for full prompt.")
		truncated = true
		if len(prompt) > 2000 {
			prompt = prompt[:2000] + "... [truncated, use verbose=true for full prompt]"
		}
		data["prompt"] = prompt
	}

	var nextStep string
	if completedIDs[session.ID] {
		nextStep = fmt.Sprintf("Session %s is already completed. Run vivechak_next_session without session_id for the next actionable session.", session.ID)
	} else {
		nextStep = fmt.Sprintf("Execute the prompt for session %s in a fresh AI session with web search enabled. "+
			"Save the output with vivechak_save_session using session_id='%s'.", session.ID, session.ID)
		if len(otherReady) > 0 {
			nextStep += fmt.Sprintf(" Also ready: %s (these can run in parallel).", strings.Join(otherReady, ", "))
		}
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Session %s: %s (%d/%d complete)", session.ID, session.Title, len(completedIDs), len(dag.Sessions)),
		Data:     data,
		Warnings: warnings,
		NextStep: nextStep,
		Meta:     &EnvMeta{API: 1, Tool: tool, Truncated: truncated},
	}
	return env.ToResult()
}

// scanCompletedSessions reads the sessions directory and returns completed IDs.
func scanCompletedSessions(sessionsDir string) (map[string]bool, error) {
	completed := map[string]bool{}
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return completed, nil // Sessions dir not yet created — expected
		}
		return nil, fmt.Errorf("reading sessions directory: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		// Extract session ID from filename (e.g., "T1-01.md" → "T1-01")
		name := strings.TrimSuffix(e.Name(), ".md")
		completed[name] = true

		// Also try to match longer filenames like "T1-01-graph-persistence.md"
		// by extracting the ID prefix
		parts := strings.SplitN(name, "-", 3)
		if len(parts) >= 2 {
			// Reconstruct short ID: "T1" + "-" + "01" = "T1-01"
			shortID := parts[0] + "-" + parts[1]
			completed[shortID] = true
		}
	}

	return completed, nil
}

// sessionIDList returns a comma-separated list of session IDs in the DAG.
func sessionIDList(dag *core.DAG) string {
	ids := make([]string, len(dag.Sessions))
	for i, s := range dag.Sessions {
		ids[i] = s.ID
	}
	return strings.Join(ids, ", ")
}
