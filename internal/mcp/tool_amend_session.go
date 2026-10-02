package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// AmendSessionInput holds the arguments for vivechak_amend_session.
type AmendSessionInput struct {
	ProjectRoot      string `json:"project_root,omitempty"       jsonschema:"workspace root path"`
	SessionID        string `json:"session_id"                    jsonschema:"session ID to amend (e.g. R-01)"`
	AmendingSessionID string `json:"amending_session_id,omitempty" jsonschema:"session ID that discovered the amendment (optional, for traceability)"`
	Amendment        string `json:"amendment"                     jsonschema:"amendment content to append"`
}

func registerAmendSession(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_amend_session",
			Title: "Amend Session",
			Description: "Append a post-hoc amendment note to an existing session file. " +
				"Preserves the original content while documenting evolved understanding. " +
				"Use when later sessions reveal earlier sessions were incomplete or wrong.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleAmendSession,
	)
}

func handleAmendSession(ctx context.Context, _ *sdkmcp.CallToolRequest, in AmendSessionInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_amend_session"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err,
			"Initialize a workspace first with vivechak_init.")
	}

	if strings.TrimSpace(in.SessionID) == "" {
		return ErrorResult(tool, fmt.Errorf("session_id is required"),
			"Provide the session ID to amend (e.g. 'R-01').")
	}

	if !isValidID(in.SessionID) {
		return ErrorResult(tool, fmt.Errorf("invalid session_id %q", in.SessionID),
			"Use a valid session ID like 'R-01' or 'T1-01'.")
	}

	if strings.TrimSpace(in.Amendment) == "" {
		return ErrorResult(tool, fmt.Errorf("amendment content is required"),
			"Provide the amendment text to append to the session.")
	}

	if err := validateContentSize(in.Amendment); err != nil {
		return ErrorResult(tool, err, "Reduce amendment size.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err),
			"Provide a valid workspace.")
	}
	defer ws.Close()

	// Find the session file
	var relPath string
	var displayName string
	if core.IsSynthesisSession(in.SessionID) {
		relPath = core.FADFile
		displayName = filepath.Base(core.FADFile)
		if _, err := ws.ReadFile(relPath); err != nil {
			return ErrorResult(tool,
				fmt.Errorf("synthesis file %s not found", relPath),
				"Ensure the synthesis session (FAD) has been saved with vivechak_save_session first.")
		}
	} else {
		sessionsDir := core.SessionsDir
		var sessionFile string
		if entries, err := ws.ListDir(sessionsDir); err == nil {
			for _, e := range entries {
				if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
					continue
				}
				name := strings.TrimSuffix(e.Name(), ".md")
				nameUpper := strings.ToUpper(name)
				idUpper := strings.ToUpper(in.SessionID)
				if nameUpper == idUpper ||
					strings.HasPrefix(nameUpper, idUpper+"-") ||
					strings.HasPrefix(nameUpper, idUpper+"_") {
					sessionFile = e.Name()
					break
				}
			}
		}

		if sessionFile == "" {
			return ErrorResult(tool,
				fmt.Errorf("session file for %q not found in %s", in.SessionID, sessionsDir),
				fmt.Sprintf("Ensure session %s has been saved with vivechak_save_session first.", in.SessionID))
		}
		relPath = filepath.Join(sessionsDir, sessionFile)
		displayName = sessionFile
	}

	// Acquire lock
	unlock, err := store.LockFile(ctx, filepath.Join(root, relPath), 5*time.Second)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("could not acquire lock: %w", err),
			"Another process may be writing. Try again.")
	}
	defer func() { _ = unlock() }()

	// Read existing content
	existing, err := ws.ReadFile(relPath)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("reading session file: %w", err),
			"Check workspace integrity.")
	}

	// Build amendment block
	amendSource := ""
	if in.AmendingSessionID != "" {
		amendSource = fmt.Sprintf(" (appended by %s)", in.AmendingSessionID)
	}
	timestamp := time.Now().UTC().Format(time.RFC3339)
	amendmentBlock := fmt.Sprintf("\n\n---\n\n## Post-Hoc Amendment%s\n\n%s\n\n*Added: %s*\n",
		amendSource, strings.TrimSpace(in.Amendment), timestamp)

	// Append to file
	newContent := append(existing, []byte(amendmentBlock)...)
	if err := store.WriteFileAtomic(ws.Root(), relPath, newContent, 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("writing amended session: %w", err),
			"Check filesystem permissions.")
	}

	dataMap := map[string]any{
		"workspace_root":      root,
		"session_id":          in.SessionID,
		"file_path":           relPath,
		"amendment_timestamp": timestamp,
	}
	if in.AmendingSessionID != "" {
		dataMap["amending_session_id"] = in.AmendingSessionID
	}

	// Check for completed downstream sessions that may now be stale
	var warnings []string
	if pipelineData, err := ws.ReadFile(core.PipelineFile); err == nil {
		if dag, err := core.ParsePipeline(pipelineData); err == nil && dag != nil {
			dependents := dag.TransitiveDependents(in.SessionID)
			if len(dependents) > 0 {
				var knownIDs []string
				for _, s := range dag.Sessions {
					knownIDs = append(knownIDs, s.ID)
				}
				if completedIDs, err := scanCompletedSessions(ws, knownIDs...); err == nil {
					var staleDependents []string
					for _, depID := range dependents {
						if completedIDs[depID] {
							staleDependents = append(staleDependents, depID)
						}
					}
					if len(staleDependents) > 0 {
						dataMap["potentially_stale_sessions"] = staleDependents
						warnings = append(warnings, fmt.Sprintf(
							"W-STALE-DOWNSTREAM: %d downstream session(s) were completed before this amendment: %s. "+
								"Review whether their conclusions still hold given the amended findings.",
							len(staleDependents), strings.Join(staleDependents, ", ")))
					}
				}
			}
		}
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Amendment appended to session %s (%s)", in.SessionID, displayName),
		Data:     dataMap,
		Warnings: warnings,
		NextStep: "Continue with vivechak_next_session for the next research session, " +
			"or vivechak_status to review progress.",
		Meta: NewMeta(tool),
	}
	return env.ToResult()
}
