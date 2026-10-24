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

// AmendInput holds the arguments for vivechak_amend.
type AmendInput struct {
	ProjectRoot       string `json:"project_root,omitempty"       jsonschema:"workspace root path"`
	SessionID         string `json:"session_id,omitempty"        jsonschema:"session ID, decision ID, or comparison ID to amend (e.g. R-01, D-001)"`
	Target            string `json:"target,omitempty"            jsonschema:"alias for session_id"`
	AmendingSessionID string `json:"amending_session_id,omitempty" jsonschema:"session ID that discovered the amendment (optional, for traceability)"`
	Amendment         string `json:"amendment,omitempty"           jsonschema:"amendment content to append"`
	Content           string `json:"content,omitempty"             jsonschema:"alias for amendment content"`
	Reason            string `json:"reason,omitempty"              jsonschema:"optional reason or summary for the amendment"`
}

// AmendSessionInput provides backward compatibility for unit tests calling handleAmendSession.
type AmendSessionInput = AmendInput

func registerAmend(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_amend",
			Title: "Amend Research Session or Decision",
			Description: "Append a post-hoc amendment note to an existing session, decision, or comparison matrix. " +
				"Preserves original findings while documenting evolved understanding and notifying downstream stale sessions. " +
				"Use when later sessions reveal earlier findings were incomplete or superseded.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleAmend,
	)
}

func handleAmend(ctx context.Context, req *sdkmcp.CallToolRequest, in AmendInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_amend"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start.")
	}

	targetID := strings.TrimSpace(in.SessionID)
	if targetID == "" {
		targetID = strings.TrimSpace(in.Target)
	}

	if targetID == "" {
		return ErrorResult(tool, fmt.Errorf("session_id (or target) is required"),
			"Provide the ID to amend (e.g. 'T1-01', 'D-001').")
	}

	if !isValidID(targetID) {
		return ErrorResult(tool, fmt.Errorf("invalid identifier %q", targetID),
			"Use a valid ID like 'T1-01' or 'D-001'.")
	}

	amendmentContent := in.Amendment
	if amendmentContent == "" {
		amendmentContent = in.Content
	}
	if strings.TrimSpace(amendmentContent) == "" {
		return ErrorResult(tool, fmt.Errorf("amendment content is required"),
			"Provide the amendment text explaining the new finding or correction.")
	}

	if err := core.CheckHumanOnlyField([]byte(amendmentContent)); err != nil {
		return ErrorResult(tool, err, "Ask the human architect to run 'vck signoff <decision-id>' in an interactive terminal.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err),
			"Provide a valid workspace.")
	}
	defer ws.Close()

	// Locate target file (sessions, FAD, decisions, or comparisons)
	var relPath string
	if core.IsSynthesisSession(targetID) {
		relPath = core.FADFile
	} else {
		// Look in sessions first
		matches, _ := ws.ListDir(core.SessionsDir)
		for _, m := range matches {
			stem := strings.TrimSuffix(m.Name(), ".md")
			if strings.EqualFold(stem, targetID) || strings.HasPrefix(strings.ToLower(stem), strings.ToLower(targetID)+"-") {
				relPath = filepath.Join(core.SessionsDir, m.Name())
				break
			}
		}

		// If not in sessions, check research root (for D-*.md or C-*.md)
		if relPath == "" {
			rootMatches, _ := ws.ListDir(core.ResearchDir)
			for _, m := range rootMatches {
				stem := strings.TrimSuffix(m.Name(), ".md")
				if strings.EqualFold(stem, targetID) || strings.HasPrefix(strings.ToLower(stem), strings.ToLower(targetID)+"-") {
					relPath = filepath.Join(core.ResearchDir, m.Name())
					break
				}
			}
		}

		// Also check research/decisions/
		if relPath == "" {
			decMatches, _ := ws.ListDir(core.DecisionsDir)
			for _, m := range decMatches {
				stem := strings.TrimSuffix(m.Name(), ".md")
				if strings.EqualFold(stem, targetID) || strings.HasPrefix(strings.ToLower(stem), strings.ToLower(targetID)+"-") {
					relPath = filepath.Join(core.DecisionsDir, m.Name())
					break
				}
			}
		}
	}

	if relPath == "" {
		return ErrorResult(tool, fmt.Errorf("target file for %s not found in workspace", targetID),
			"Check the ID against existing sessions or decisions.")
	}

	existingBytes, err := ws.ReadFile(relPath)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("reading %s: %w", relPath, err), "Check file permissions.")
	}

	// Format amendment section
	now := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	by := in.AmendingSessionID
	if by == "" {
		by = "Post-hoc Review"
	}

	var sb strings.Builder
	sb.Write(existingBytes)
	if !strings.HasSuffix(string(existingBytes), "\n") {
		sb.WriteString("\n")
	}
	amendSource := ""
	if by != "" && by != "Post-hoc Review" {
		amendSource = fmt.Sprintf(" (appended by %s)", by)
	}
	fmt.Fprintf(&sb, "\n\n---\n\n## Post-Hoc Amendment%s\n\n", amendSource)
	if in.Reason != "" {
		fmt.Fprintf(&sb, "**Reason:** %s\n\n", in.Reason)
	}
	sb.WriteString(strings.TrimSpace(amendmentContent))
	fmt.Fprintf(&sb, "\n\n*Added: %s*\n", now)

	updatedContent := []byte(sb.String())

	unlock, lockErr := store.LockFile(ctx, filepath.Join(root, relPath), 5*time.Second)
	if lockErr != nil {
		return ErrorResult(tool, fmt.Errorf("could not acquire lock for %s: %w", relPath, lockErr), "Try again.")
	}
	defer func() { _ = unlock() }()

	if err := store.WriteFileAtomic(ws.Root(), relPath, updatedContent, 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("writing amended %s: %w", relPath, err), "Check file permissions.")
	}

	dataMap := map[string]any{
		"target_id":    targetID,
		"amended_file": relPath,
	}
	var warnings []string

	// If amending an ADR, recompile DECISIONS.md
	isADR := (filepath.Clean(filepath.Dir(relPath)) == filepath.Clean(core.ResearchDir) || filepath.Clean(filepath.Dir(relPath)) == filepath.Clean(core.DecisionsDir)) && strings.HasPrefix(strings.ToUpper(filepath.Base(relPath)), "D-")
	if isADR {
		if compileErr := compileDecisionsRegistry(ctx, ws, root); compileErr != nil {
			warnings = append(warnings, fmt.Sprintf("W-DECISIONS-COMPILE: failed to recompile DECISIONS.md: %v", compileErr))
		} else {
			dataMap["registry_recompiled"] = core.DecisionsFile
		}
	}

	// Check downstream sessions for staleness
	if pipelineData, err := ws.ReadFile(core.PipelineFile); err == nil {
		if dag, err := core.ParsePipeline(pipelineData); err == nil && dag != nil {
			dependents := dag.TransitiveDependents(targetID)
			if len(dependents) > 0 {
				var knownIDs []string
				for _, s := range dag.Sessions {
					knownIDs = append(knownIDs, s.ID)
				}
				completedIDs, _ := scanCompletedSessions(ws, knownIDs...)
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

	next := &NextAction{
		Tool: "vivechak_next",
		Why:  "Amendment recorded. Continue with next ready session using vivechak_next.",
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Amendment appended to %s (%s).", targetID, relPath),
		Data:     dataMap,
		Warnings: warnings,
		Next:     next,
		NextStep: next.Why,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// handleAmendSession provides backward compatibility for unit tests calling handleAmendSession.
func handleAmendSession(ctx context.Context, req *sdkmcp.CallToolRequest, in AmendSessionInput) (*sdkmcp.CallToolResult, Envelope, error) {
	return handleAmend(ctx, req, in)
}
