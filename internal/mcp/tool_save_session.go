package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	store "github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SaveSessionInput holds the arguments for vivechak_save_session.
type SaveSessionInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path"`
	SessionID   string `json:"session_id"             jsonschema:"session identifier (e.g. R-01, R-02, SYN-01)"`
	Content     string `json:"content"                 jsonschema:"completed session output (Markdown with YAML frontmatter)"`
}

func registerSaveSession(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_save_session",
			Title: "Save Session Output",
			Description: "Validate and persist a completed research session output. " +
				"Runs 4-level validation ladder: checks YAML frontmatter, required fields, " +
				"evidence grades, and section structure. Does NOT reject for formatting variation — " +
				"validates structure, not style. Saves as draft if blocking issues found.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleSaveSession,
	)
}

func handleSaveSession(ctx context.Context, _ *sdkmcp.CallToolRequest, in SaveSessionInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save_session"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err,
			"Initialize a workspace first with vivechak_init.")
	}

	if strings.TrimSpace(in.SessionID) == "" {
		return ErrorResult(tool, fmt.Errorf("session_id is required"),
			"Provide the session identifier (e.g., 'R-01', 'R-02', 'SYN-01').")
	}

	// Validate session ID format to prevent path traversal
	if !isValidID(in.SessionID) {
		return ErrorResult(tool, fmt.Errorf("invalid session_id %q — must be alphanumeric with optional hyphens, underscores, dots", in.SessionID),
			"Use a simple ID like 'T1-01', 'SYN-01', or 'FAD'.")
	}

	if strings.TrimSpace(in.Content) == "" {
		return ErrorResult(tool, fmt.Errorf("content is required"),
			"Provide the session output content (Markdown with YAML frontmatter).")
	}

	if err := validateContentSize(in.Content); err != nil {
		return ErrorResult(tool, err, "Reduce content size or split into multiple artifacts.")
	}

	// FAD (Founding Architecture Document) writes to research/FAD.md
	isSynthesis := core.IsSynthesisSession(in.SessionID)

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err), "Provide a valid workspace.")
	}
	defer ws.Close()

	// Load DAG if available to inspect session door type
	var dag *core.DAG
	isOneWay := false
	if pipeData, err := ws.ReadFile(core.PipelineFile); err == nil {
		if parsedDAG, err := core.ParsePipeline(pipeData); err == nil {
			dag = parsedDAG
			if s := dag.SessionByID(in.SessionID); s != nil {
				isOneWay = strings.EqualFold(s.DoorType, "one-way")
			}
		}
	}

	// Validate the session content (use ValidateFAD for FAD/synthesis sessions)
	var validation *core.ValidationResult
	if isSynthesis {
		validation = core.ValidateFAD([]byte(in.Content))
	} else {
		validation = core.ValidateSessionWithContext([]byte(in.Content), isOneWay)
	}

	// Determine filename
	filename := in.SessionID + ".md"

	// Ensure sessions directory exists
	if err := ws.MkdirAll(core.SessionsDir, 0o755); err != nil {
		return ErrorResult(tool, fmt.Errorf("creating sessions directory: %w", err),
			"Check filesystem permissions.")
	}

	var relPath string
	if isSynthesis {
		relPath = core.FADFile
	} else {
		relPath = filepath.Join(core.SessionsDir, filename)
	}
	unlock, err := store.LockFile(ctx, filepath.Join(root, relPath), 5*time.Second)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("could not acquire lock: %w", err), "Another process may be writing. Try again.")
	}
	defer func() { _ = unlock() }()

	// L1 Construct: auto-remedy missing status in frontmatter via in-place insertion
	contentToSave := core.EnsureFrontmatterField([]byte(in.Content), "status", "draft")

	// Save the file (even with warnings — L2 saves as draft)
	if err := store.WriteFileAtomic(ws.Root(), relPath, contentToSave, 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("writing session %s: %w", filename, err),
			"Check filesystem permissions.")
	}

	status := validation.Status
	warningCount := validation.WarningCount()
	errorCount := validation.ErrorCount()

	// Build warnings list from validation issues
	var warnings []string
	for _, issue := range validation.Issues {
		warnings = append(warnings, issue.String())
	}

	var observations []core.QualityObservation
	if !isSynthesis {
		observations = core.ObserveSessionQuality([]byte(in.Content))
		for _, o := range observations {
			if o.Severity == "consideration" || o.Severity == "suggestion" {
				warnings = append(warnings, fmt.Sprintf("Q-%s: %s", strings.ToUpper(o.Category), o.Message))
			}
		}
	}

	nextStep := buildNextStep(in.SessionID, errorCount, observations, dag)

	dataMap := map[string]any{
		"workspace_root":    root,
		"session_id":        in.SessionID,
		"file_path":         relPath,
		"status":            status,
		"validation_passed": errorCount == 0,
		"validation":        validation,
	}

	if isSynthesis {
		// Also copy to project root for human consumption
		rootFAD := "FOUNDING-ARCHITECTURE.md"
		if cpErr := store.WriteFileAtomic(ws.Root(), rootFAD, contentToSave, 0o644); cpErr != nil {
			warnings = append(warnings, fmt.Sprintf(
				"W-FAD-COPY: could not copy FAD to project root: %v", cpErr))
		} else {
			dataMap["root_copy"] = rootFAD
		}
	}

	// Quality observations (advisory, never blocking)
	if !isSynthesis && len(observations) > 0 {
		obsData := make([]map[string]string, len(observations))
		for i, o := range observations {
			obsData[i] = map[string]string{
				"category": o.Category,
				"message":  o.Message,
				"severity": o.Severity,
			}
		}
		dataMap["quality_observations"] = obsData
	}

	// Mini-status (eliminates need for separate status calls)
	completedSessions, _ := scanCompletedSessions(ws)
	gateDecisions := collectGateDecisions(ws)
	progress := map[string]any{
		"sessions_completed": len(completedSessions),
		"decisions_recorded": len(gateDecisions),
	}
	info := core.InspectWorkspace(root)
	if info.HasPipeline {
		if pipeData, err := ws.ReadFile(core.PipelineFile); err == nil {
			if dag, err := core.ParsePipeline(pipeData); err == nil {
				progress["sessions_total"] = len(dag.Sessions)
			}
		}
	}
	dataMap["progress"] = progress

	env := Envelope{
		Success: true,
		Message: fmt.Sprintf("Session %s saved as %s (%d warnings, %d errors)", in.SessionID, status, warningCount, errorCount),
		Data:    dataMap,
		Warnings: warnings,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

func buildNextStep(sessionID string, errorCount int, observations []core.QualityObservation, dag *core.DAG) string {
	if errorCount > 0 {
		return fmt.Sprintf("Session saved as draft (%d blocking issues). Fix the issues and re-save, "+
			"or run vivechak_validate for a dry-run check.", errorCount)
	}

	var parts []string

	// 1. Surface quality advisories directly in next_step
	var advisories []string
	for _, obs := range observations {
		if obs.Severity == "consideration" || obs.Severity == "suggestion" {
			advisories = append(advisories, fmt.Sprintf("[Q-%s] %s", strings.ToUpper(obs.Category), obs.Message))
		}
	}
	if len(advisories) > 0 {
		parts = append(parts, "⚠️ QUALITY ADVISORIES: "+strings.Join(advisories, " | "))
	}

	// 2. Door type check from DAG
	isOneWay := false
	if dag != nil {
		if s := dag.SessionByID(sessionID); s != nil {
			isOneWay = strings.EqualFold(s.DoorType, "one-way")
		}
	}

	// 3. Route based on door type
	if isOneWay {
		parts = append(parts, fmt.Sprintf(
			"🛑 ONE-WAY DOOR: Session '%s' informs an irreversible decision. "+
				"Run 'vivechak_challenge(session_id=\"%s\", mode=\"red_team\")' to stress-test "+
				"your findings before recording the decision. Amend with challenge results "+
				"via vivechak_amend_session.",
			sessionID, sessionID))
	} else {
		parts = append(parts, "Run vivechak_next_session for the next session, or "+
			"vivechak_record_decision to record decisions from this session's findings.")
	}

	return strings.Join(parts, "\n\n")
}

