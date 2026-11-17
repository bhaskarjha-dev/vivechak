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

// AmendInput holds the canonical arguments for vivechak_amend.
type AmendInput struct {
	ProjectRoot       string `json:"project_root,omitempty"`
	ID                string `json:"id"`
	Note              string `json:"note"`
	Reason            string `json:"reason,omitempty"`
	AmendingSessionID string `json:"amending_session_id,omitempty"`
}

func registerAmend(server *sdkmcp.Server) {
	registerTool(server, ToolSpec{
		Name:        "vivechak_amend",
		Description: "Append a dated amendment to a saved session, ADR, FAD, or comparison and flag stale dependents.",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint:    false,
			IdempotentHint:  false,
			DestructiveHint: BoolPtr(true),
		},
		Params: []ParamSpec{
			{Name: "id", Type: "string", Required: true, Description: "Target ID", Doc: "Session, decision, comparison, or synthesis ID to amend."},
			{Name: "note", Type: "string", Required: true, Description: "Amendment note", Doc: "Amendment Markdown note to append."},
			{Name: "reason", Type: "string", Doc: "Short summary of why the artifact is being amended."},
			{Name: "amending_session_id", Type: "string", Description: "Source session", Doc: "Session ID that discovered the amendment."},
			{Name: "project_root", Type: "string", Doc: "Absolute path to the target project directory."},
		},
	}, handleAmend)
}

func handleAmend(ctx context.Context, req *sdkmcp.CallToolRequest, in AmendInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_amend"

	var roots []string
	if strings.TrimSpace(in.ProjectRoot) == "" {
		roots = getClientRoots(ctx, req)
	}
	root, rule, err := core.ResolveWorkspaceWithRoots(ctx, in.ProjectRoot, roots)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start.")
	}

	targetID := strings.TrimSpace(in.ID)
	if targetID == "" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: id is required"),
			"Provide the ID to amend in id (e.g. 'T1-01', 'D-001').")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: opening workspace: %w", err),
			"Provide a valid workspace.")
	}
	defer ws.Close()

	var warnings []Warning
	if resolved, mapped, idErr := resolvePlaceholderRef(sessionOf(req), ws, root, targetID, true); idErr != nil {
		return ErrorResult(tool, idErr, placeholderRefFix)
	} else if mapped {
		warnings = append(warnings, NewWarning("W-ID-ALLOCATED", fmt.Sprintf("id %s resolved to %s", targetID, resolved), ""))
		targetID = resolved
	}
	if in.AmendingSessionID != "" {
		if resolved, mapped, idErr := resolvePlaceholderRef(sessionOf(req), ws, root, in.AmendingSessionID, false); idErr != nil {
			return ErrorResult(tool, idErr, placeholderRefFix)
		} else if mapped {
			warnings = append(warnings, NewWarning("W-ID-ALLOCATED", fmt.Sprintf("amending_session_id %s resolved to %s", in.AmendingSessionID, resolved), ""))
			in.AmendingSessionID = resolved
		}
	}

	if !isValidID(targetID) {
		return ErrorResult(tool, fmt.Errorf("E-ID-INVALID: invalid identifier %q", targetID),
			"Use a valid ID like 'T1-01' or 'D-001'.")
	}

	amendmentContent := in.Note
	if strings.TrimSpace(amendmentContent) == "" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: note is required"),
			"Provide the amendment text in note.")
	}

	if err := core.CheckHumanOnlyField([]byte(amendmentContent)); err != nil {
		return ErrorResult(tool, err, "Ask the human architect to run 'vck signoff <decision-id>' in an interactive terminal.")
	}

	var relPath string
	if isSynthesisTarget(ws, targetID) || strings.EqualFold(targetID, "FAD") {
		relPath = core.FADFile
	} else {
		normID := core.NormalizeDecisionID(targetID)
		matchesID := func(stem string) bool {
			sLower := strings.ToLower(stem)
			tLower := strings.ToLower(targetID)
			nLower := strings.ToLower(normID)
			return strings.EqualFold(stem, targetID) || strings.HasPrefix(sLower, tLower+"-") ||
				(nLower != "" && (strings.EqualFold(stem, normID) || strings.HasPrefix(sLower, nLower+"-")))
		}

		matches, _ := ws.ListDir(core.SessionsDir)
		for _, m := range matches {
			stem := strings.TrimSuffix(m.Name(), ".md")
			if matchesID(stem) {
				relPath = filepath.Join(core.SessionsDir, m.Name())
				break
			}
		}

		if relPath == "" {
			if fadBytes, fErr := ws.ReadFile(core.FADFile); fErr == nil && len(fadBytes) > 0 {
				if fm, _, _ := core.ParseFrontmatter(fadBytes); fm != nil {
					fid := fm.GetString("id")
					if fid == "" {
						fid = fm.GetString("session_id")
					}
					if strings.EqualFold(fid, targetID) {
						relPath = core.FADFile
					}
				}
			}
		}

		if relPath == "" {
			rootMatches, _ := ws.ListDir(core.ResearchDir)
			for _, m := range rootMatches {
				lowerName := strings.ToLower(m.Name())
				if m.IsDir() || !strings.HasSuffix(lowerName, ".md") ||
					lowerName == "research-pipeline.md" || lowerName == "decisions.md" || lowerName == "phase-0-gate.md" ||
					strings.HasSuffix(lowerName, "-plan.md") || strings.HasSuffix(lowerName, ".plan.md") {
					continue
				}
				stem := strings.TrimSuffix(m.Name(), ".md")
				if matchesID(stem) && !reSessionOrReview.MatchString(strings.ToUpper(stem)) {
					relPath = filepath.Join(core.ResearchDir, m.Name())
					break
				}
			}
		}

		if relPath == "" {
			for _, subDir := range []string{core.DecisionsDir, core.ComparisonsDir} {
				subMatches, _ := ws.ListDir(subDir)
				for _, m := range subMatches {
					stem := strings.TrimSuffix(m.Name(), ".md")
					if matchesID(stem) {
						relPath = filepath.Join(subDir, m.Name())
						break
					}
				}
				if relPath != "" {
					break
				}
			}
		}
	}

	if relPath == "" {
		return ErrorResult(tool, fmt.Errorf("E-NOT-FOUND: target file for %s not found in workspace", targetID),
			"Check the ID against existing sessions or decisions.")
	}

	existingBytes, err := ws.ReadFile(relPath)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: reading %s: %w", relPath, err), "Check file permissions.")
	}

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
		return ErrorResult(tool, fmt.Errorf("E-LOCK-TIMEOUT: could not acquire lock for %s: %w", relPath, lockErr), "Try again.")
	}
	defer func() { _ = unlock() }()

	if err := ws.WriteArtifact(relPath, updatedContent, 0o644, true); err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: writing amended %s: %w", relPath, err), "Check file permissions.")
	}

	dataMap := map[string]any{
		"workspace":    map[string]any{"root": root, "rule": rule},
		"target_id":    targetID,
		"id":           targetID,
		"amended_file": filepath.ToSlash(relPath),
	}

	isADR := (filepath.Clean(filepath.Dir(relPath)) == filepath.Clean(core.ResearchDir) || filepath.Clean(filepath.Dir(relPath)) == filepath.Clean(core.DecisionsDir)) && strings.HasPrefix(strings.ToUpper(filepath.Base(relPath)), "D-")
	if isADR {
		if compileErr := compileDecisionsRegistry(ctx, ws, root); compileErr != nil {
			warnings = append(warnings, NewWarning("W-DECISIONS-COMPILE", fmt.Sprintf("failed to recompile DECISIONS.md: %v", compileErr), ""))
		} else {
			dataMap["registry_recompiled"] = core.DecisionsFile
		}
	}

	if pipelineData, err := ws.ReadFile(core.PipelineFile); err == nil {
		if dag, err := core.ParsePipeline(pipelineData); err == nil && dag != nil {
			dependents := dag.TransitiveDependents(targetID)
			if len(dependents) > 0 {
				completedIDs, _ := dag.CompletedSessions(ws)
				var staleDependents []string
				for _, depID := range dependents {
					if completedIDs[depID] {
						staleDependents = append(staleDependents, depID)
					}
				}
				if len(staleDependents) > 0 {
					dataMap["potentially_stale_sessions"] = staleDependents
					warnings = append(warnings, NewWarning(
						"W-STALE-DOWNSTREAM",
						fmt.Sprintf("%d downstream session(s) were completed before this amendment: %s.", len(staleDependents), strings.Join(staleDependents, ", ")),
						"",
					))
				}
			}
		}
	}

	env := Envelope{
		Ok:       true,
		Message:  fmt.Sprintf("Amendment appended to %s (%s).", targetID, filepath.ToSlash(relPath)),
		Data:     dataMap,
		Warnings: warnings,
		Next: &NextAction{
			Tool: "vivechak_next",
			Args: map[string]any{},
			Why:  "Amendment recorded. Continue with next ready session using vivechak_next.",
		},
	}
	return env.ToResult()
}
