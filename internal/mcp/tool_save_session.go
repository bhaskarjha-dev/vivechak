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

	// Validate the session content
	validation := core.ValidateSession([]byte(in.Content))

	// Determine filename
	filename := in.SessionID + ".md"

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err), "Provide a valid workspace.")
	}
	defer ws.Close()

	// Ensure sessions directory exists
	if err := ws.MkdirAll(core.SessionsDir, 0o755); err != nil {
		return ErrorResult(tool, fmt.Errorf("creating sessions directory: %w", err),
			"Check filesystem permissions.")
	}

	// FAD (Founding Architecture Document) writes to research/FAD.md
	isSynthesis := in.SessionID == "FAD" || strings.HasPrefix(strings.ToUpper(in.SessionID), "SYN")
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
	defer unlock()

	// L1 Construct: auto-remedy missing status in frontmatter
	contentToSave := []byte(in.Content)
	if fm, body, err := core.ParseFrontmatter(contentToSave); err == nil && fm != nil {
		if !fm.Has("status") {
			fm.Set("status", "draft")
			if composed, err := core.ComposeFrontmatter(fm, body); err == nil {
				contentToSave = composed
			}
		}
	}

	// Save the file (even with warnings — L2 saves as draft)
	if err := store.WriteFileAtomic(ws.Root(), relPath, contentToSave, 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("writing session %s: %w", filename, err),
			"Check filesystem permissions.")
	}

	// For synthesis sessions saved with session ID (e.g. SYN-01), also save to sessions/ directory
	if isSynthesis && in.SessionID != "FAD" {
		sessPath := filepath.Join(core.SessionsDir, filename)
		_ = store.WriteFileAtomic(ws.Root(), sessPath, contentToSave, 0o644)
	}

	status := validation.Status
	warningCount := validation.WarningCount()
	errorCount := validation.ErrorCount()

	// Build warnings list from validation issues
	var warnings []string
	for _, issue := range validation.Issues {
		warnings = append(warnings, issue.String())
	}

	var nextStep string
	if errorCount > 0 {
		nextStep = fmt.Sprintf("Session saved as draft (%d blocking issues). Fix the issues and re-save, "+
			"or run vivechak_validate for a dry-run check.", errorCount)
	} else {
		nextStep = "Run vivechak_next_session for the next session, or " +
			"vivechak_record_decision to record decisions from this session's findings."
	}

	env := Envelope{
		Success: true,
		Message: fmt.Sprintf("Session %s saved as %s (%d warnings, %d errors)", in.SessionID, status, warningCount, errorCount),
		Data: map[string]any{
			"workspace_root":    root,
			"session_id":        in.SessionID,
			"file_path":         relPath,
			"status":            status,
			"validation_passed": errorCount == 0,
			"validation":        validation,
		},
		Warnings: warnings,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}
