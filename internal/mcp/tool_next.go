package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NextInput holds the arguments for vivechak_next.
type NextInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path (optional)"`
	SessionID   string `json:"session_id,omitempty"   jsonschema:"specific session ID to retrieve (optional; returns next ready session if omitted)"`
	Verbose     bool   `json:"verbose,omitempty"      jsonschema:"if true, include full upstream findings in response"`
}

// NextSessionInput provides backward compatibility for unit tests calling handleNextSession.
type NextSessionInput = NextInput

func registerNext(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_next",
			Title: "Get Next Research Session",
			Description: "Return the next actionable research session prompt with upstream findings " +
				"automatically injected into context slots. Resolves DAG dependencies and topological order. " +
				"For synthesis sessions (SYN-01/FAD), aggregates all finalized session findings. " +
				"Eliminates manual copy-paste between sessions. Read-only.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleNext,
	)
}

func handleNext(_ context.Context, _ *sdkmcp.CallToolRequest, in NextInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_next"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start.")
	}

	if !core.WorkspaceExists(root) {
		return ErrorResult(tool, fmt.Errorf("workspace not initialized at %s", root), "Run vivechak_start first.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err), "Provide a valid workspace.")
	}
	defer ws.Close()

	info := core.InspectWorkspace(root)
	if info.Scope == core.ScopeComparison {
		var compFile string
		var compContent []byte
		if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				if !e.IsDir() && (strings.HasSuffix(e.Name(), "-comparison.md") || strings.Contains(e.Name(), "-cmp-") || strings.EqualFold(e.Name(), "COMPARISON.md")) {
					compFile = filepath.Join(core.ResearchDir, e.Name())
					if data, rErr := ws.ReadFile(compFile); rErr == nil {
						compContent = data
						break
					}
				}
			}
		}

		compSessionID := core.SessionPrefix(info.Scope) + "01"
		if strings.Contains(compFile, "CMP-") || strings.Contains(string(compContent), "CMP-01") || strings.Contains(compFile, "COMP-") {
			compSessionID = "CMP-01"
		} else if idMatch := regexp.MustCompile(`(?m)^\|\s*(?:\*\*)?ID(?:\*\*)?\s*\|\s*([A-Za-z0-9_-]+)\s*\|`).FindSubmatch(compContent); len(idMatch) > 1 {
			compSessionID = string(idMatch[1])
		}

		promptStr := string(compContent)
		dataMap := map[string]any{
			"workspace_root":       root,
			"scope":                info.Scope,
			"session_id":           compSessionID,
			"prompt":               promptStr,
			"prompt_char_count":    len(promptStr),
			"prompt_approx_tokens": len(promptStr) / 4,
		}
		if compFile != "" {
			dataMap["source_file"] = compFile
		}

		msg := "Comparison scope research prompt ready."
		if promptStr == "" {
			msg = "Comparison scope workspace initialized. No comparison plan found yet."
		}

		env := Envelope{
			Ok:       true,
			Success:  true,
			Message:  msg,
			Data:     dataMap,
			Next:     &NextAction{Tool: "vivechak_save", Args: map[string]any{"kind": "comparison"}, Why: "Conduct research and grade claims with verifiable evidence. Save comparison matrix with vivechak_save(kind='comparison')."},
			NextStep: "Conduct research and grade claims with verifiable evidence. Save comparison matrix with vivechak_save(kind='comparison').",
			Meta:     NewMeta(tool),
		}
		return env.ToResult()
	}

	pipelineData, err := ws.ReadFile(core.PipelineFile)
	if err != nil {
		// Look for decision plan file (*-plan.md) in research/
		if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), "-plan.md") {
					if data, rErr := ws.ReadFile(filepath.Join(core.ResearchDir, e.Name())); rErr == nil {
						pipelineData = data
						err = nil
						break
					}
				}
			}
		}
		if err != nil {
			return ErrorResult(tool, fmt.Errorf("reading %s: %w", core.PipelineFile, err),
				"Create a research plan first using vivechak_save(kind='plan').")
		}
	}

	dag, err := core.ParsePipeline(pipelineData)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("parsing %s: %w", core.PipelineFile, err),
			"Fix the pipeline format and try again.")
	}

	knownIDs := make([]string, 0, len(dag.Sessions))
	for _, s := range dag.Sessions {
		knownIDs = append(knownIDs, s.ID)
	}
	completedIDs, _ := scanCompletedSessions(ws, knownIDs...)

	if in.SessionID != "" {
		target := strings.TrimSpace(in.SessionID)
		session := dag.SessionByID(target)
		if session == nil {
			var avail []string
			for _, s := range dag.Sessions {
				avail = append(avail, s.ID)
			}
			return ErrorResult(tool, fmt.Errorf("session %s not found in pipeline", target),
				fmt.Sprintf("Available sessions: %s. Check session ID against RESEARCH-PIPELINE.md.", strings.Join(avail, ", ")))
		}

		injected, injErr := core.InjectContextFS(*session, ws, core.SessionsDir, completedIDs)
		var prompt string
		var injWarnings []string
		if injErr == nil && injected != nil {
			prompt = injected.InjectedPrompt
			if injected.Warning != "" {
				injWarnings = append(injWarnings, injected.Warning)
			}
		} else {
			prompt = session.Prompt
			if injErr != nil {
				injWarnings = append(injWarnings, fmt.Sprintf(
					"W-INJECT-FAILED: context injection failed for %s: %v — using raw prompt without upstream findings",
					session.ID, injErr))
			}
		}

		if core.IsSynthesisSession(session.ID) {
			if adv := checkPreSynthesisAdvisory(dag, completedIDs, ws); adv != "" {
				injWarnings = append(injWarnings, adv)
			}
		}

		return buildSessionResponse(tool, *session, prompt, completedIDs, dag, in.Verbose, injWarnings, ws)
	}

	ready := dag.NextSessions(completedIDs)

	if len(ready) == 0 {
		completedCount := 0
		for _, s := range dag.Sessions {
			if completedIDs[s.ID] {
				completedCount++
			}
		}

		if completedCount == len(dag.Sessions) {
			next := &NextAction{
				Tool: "vivechak_gate",
				Why:  "All research sessions are complete. Run Phase 0 Gate via vivechak_gate to verify pre-coding criteria.",
			}
			env := Envelope{
				Ok:       true,
				Success:  true,
				Message:  fmt.Sprintf("All %d sessions complete!", len(dag.Sessions)),
				Data: map[string]any{
					"workspace_root":     root,
					"total_sessions":     len(dag.Sessions),
					"completed_sessions": completedCount,
					"all_complete":       true,
				},
				Next:     next,
				NextStep: next.Why,
				Meta:     NewMeta(tool),
			}
			return env.ToResult()
		}

		var blockedInfo []map[string]any
		for _, s := range dag.Sessions {
			if completedIDs[s.ID] {
				continue
			}
			var missingDeps []string
			for _, dep := range s.Dependencies {
				if !completedIDs[dep] {
					missingDeps = append(missingDeps, formatMissingDep(ws, dep))
				}
			}
			blockedInfo = append(blockedInfo, map[string]any{
				"session_id": s.ID,
				"blocked_by": missingDeps,
			})
		}

		env := Envelope{
			Ok:       true,
			Success:  true,
			Message:  fmt.Sprintf("%d sessions remaining but all blocked by dependencies", len(dag.Sessions)-completedCount),
			Data: map[string]any{
				"workspace_root":     root,
				"total_sessions":     len(dag.Sessions),
				"completed_sessions": completedCount,
				"blocked_sessions":   blockedInfo,
			},
			Next: &NextAction{
				Why: "Some sessions are blocked by incomplete dependencies. Complete the blocking sessions first.",
			},
			NextStep: "Some sessions are blocked by incomplete dependencies. Complete the blocking sessions first.",
			Meta:     NewMeta(tool),
		}
		return env.ToResult()
	}

	nextSession := ready[0]
	injected, injErr := core.InjectContextFS(nextSession, ws, core.SessionsDir, completedIDs)

	var prompt string
	var injWarnings []string
	if injErr == nil && injected != nil {
		prompt = injected.InjectedPrompt
		if injected.Warning != "" {
			injWarnings = append(injWarnings, injected.Warning)
		}
	} else {
		prompt = nextSession.Prompt
		if injErr != nil {
			injWarnings = append(injWarnings, fmt.Sprintf(
				"W-INJECT-FAILED: context injection failed for %s: %v — using raw prompt without upstream findings",
				nextSession.ID, injErr))
		}
	}

	var otherReady []string
	for _, s := range ready[1:] {
		otherReady = append(otherReady, s.ID)
	}

	if dag != nil {
		highestCompletedLayer := -1
		for _, s := range dag.Sessions {
			if completedIDs[s.ID] && s.Layer > highestCompletedLayer {
				highestCompletedLayer = s.Layer
			}
		}
		if nextSession.Layer > highestCompletedLayer && highestCompletedLayer >= 0 {
			injWarnings = append(injWarnings, fmt.Sprintf(
				"LAYER-TRANSITION: Layer %d complete. Before proceeding: consider whether any "+
					"remaining sessions should be adjusted based on Layer %d findings.", highestCompletedLayer, highestCompletedLayer))
		}
	}

	if core.IsSynthesisSession(nextSession.ID) {
		if adv := checkPreSynthesisAdvisory(dag, completedIDs, ws); adv != "" {
			injWarnings = append(injWarnings, adv)
		}
	}

	return buildSessionResponse(tool, nextSession, prompt, completedIDs, dag, in.Verbose, injWarnings, ws, otherReady...)
}

func formatMissingDep(ws core.WorkspaceFS, dep string) string {
	if ws != nil {
		if entries, err := ws.ListDir(core.SessionsDir); err == nil {
			for _, e := range entries {
				stem := strings.TrimSuffix(e.Name(), ".md")
				if strings.EqualFold(stem, dep) || strings.HasPrefix(strings.ToLower(stem), strings.ToLower(dep)+"-") {
					if dData, rErr := ws.ReadFile(filepath.Join(core.SessionsDir, e.Name())); rErr == nil {
						if fm, _, pErr := core.ParseFrontmatter(dData); pErr == nil && fm != nil {
							if strings.EqualFold(fm.GetString("status"), "draft") {
								return dep + " (saved as draft — resolve blocking issues and re-save)"
							}
						}
					}
					break
				}
			}
		}
	}
	return dep
}

func checkPreSynthesisAdvisory(dag *core.DAG, completedIDs map[string]bool, ws core.WorkspaceFS) string {
	if dag == nil {
		return ""
	}
	var owdIssues []string
	recalledCitationRe := regexp.MustCompile(`(?i)(?:\[[^\]\n]*\b(?:recalled|parametric)\b[^\]\n]*\]|\([^)\n]*\b(?:recalled|parametric)\b[^)\n]*\)|\brecalled\s+memory\b|\|\s*recalled\s*\|)`)

	for _, s := range dag.Sessions {
		if strings.EqualFold(s.DoorType, "one-way") && completedIDs[s.ID] {
			if sData, _, err := core.ReadSessionFileFS(ws, core.SessionsDir, s.ID); err == nil && sData != "" {
				rec := core.ExtractSection(sData, "recommend", "verdict", "decision")
				if len(rec) >= 20 && recalledCitationRe.MatchString(rec) {
					owdIssues = append(owdIssues, fmt.Sprintf("%s (recommendation relies on Grade D recalled claims)", s.ID))
				}
				lower := strings.ToLower(sData)
				if !strings.Contains(lower, "challenge findings") && !strings.Contains(lower, "red-team stress test") {
					owdIssues = append(owdIssues, fmt.Sprintf("%s (lacks adversarial challenge)", s.ID))
				}
			}
		}
	}
	if len(owdIssues) > 0 {
		return fmt.Sprintf(
			"PRE-SYNTHESIS ADVISORY: %d One-Way Door session(s) have evidentiary gaps: %s. "+
				"Note: Phase 0 Gate will FAIL after synthesis unless resolved. "+
				"You may proceed with synthesis now or amend weak sessions first with vivechak_amend.",
			len(owdIssues), strings.Join(owdIssues, "; "))
	}
	return ""
}

const calibrationBlock = "\n\n---\n### RESEARCH CALIBRATION (Active Rules)\n" +
	"1. **Memory ≠ Evidence:** Claims recalled from model weights cap at Grade D. Verify with web search.\n" +
	"2. **Upstream ≠ Exhaustive:** Injected upstream findings provide high-level architectural context, NOT implementation specifications. Exact schemas, library bundle sizes, API rate limits, and security configurations must be verified with current external sources.\n" +
	"3. **Qualified Provenance:** Grade A claims require canonical URLs, domain anchors, RFCs, or doc titles.\n" +
	"4. **Falsification Required:** Actively search for failure modes, gotchas, and production incident postmortems.\n" +
	"5. **Honest Gaps:** Explicitly stating \"searched X and found no evidence\" is valued. Never hallucinate sources.\n"

func buildSessionResponse(tool string, session core.Session, prompt string, completedIDs map[string]bool, dag *core.DAG, verbose bool, preWarnings []string, ws core.WorkspaceFS, otherReady ...string) (*sdkmcp.CallToolResult, Envelope, error) {
	data := map[string]any{
		"session_id":         session.ID,
		"title":              session.Title,
		"layer":              session.Layer,
		"door_type":          session.DoorType,
		"decision_ref":       session.DecisionRef,
		"dependencies":       session.Dependencies,
		"output_file":        session.OutputFile,
		"total_sessions":     len(dag.Sessions),
		"completed_sessions": len(completedIDs),
		"already_completed":  completedIDs[session.ID],
	}

	if !core.IsSynthesisSession(session.ID) && prompt != "" && !strings.Contains(prompt, "### RESEARCH CALIBRATION") {
		prompt += calibrationBlock
	}

	var warnings []string
	warnings = append(warnings, preWarnings...)
	displayPrompt := prompt
	approxTokens := len(prompt) / 4
	if !verbose && approxTokens > 10000 && !core.IsSynthesisSession(session.ID) {
		warnings = append(warnings, "W-OUTPUT-SIZE: Prompt exceeds 10K tokens and was truncated. Use verbose=true for full prompt.")
		if len(prompt) > 2000 {
			displayPrompt = prompt[:2000] + "... [truncated, use verbose=true for full prompt]"
		}
	}

	if displayPrompt != "" {
		data["prompt"] = displayPrompt
		data["prompt_char_count"] = len(prompt)
		data["prompt_approx_tokens"] = approxTokens
	}

	if len(otherReady) > 0 {
		data["other_ready_sessions"] = otherReady
		data["parallelism_hint"] = fmt.Sprintf(
			"PARALLEL EXECUTION AVAILABLE: %d independent sessions can run simultaneously. "+
				"If your runtime supports subagents, execute %s in parallel for %dx faster completion.",
			len(otherReady)+1, strings.Join(append([]string{session.ID}, otherReady...), ", "), len(otherReady)+1)
	}

	if dag != nil {
		var blocked []map[string]any
		for _, s := range dag.Sessions {
			if completedIDs[s.ID] || s.ID == session.ID {
				continue
			}
			isReady := false
			for _, r := range otherReady {
				if r == s.ID {
					isReady = true
					break
				}
			}
			if !isReady {
				var missingDeps []string
				for _, dep := range s.Dependencies {
					if !completedIDs[dep] {
						missingDeps = append(missingDeps, formatMissingDep(ws, dep))
					}
				}
				if len(missingDeps) > 0 {
					blocked = append(blocked, map[string]any{
						"session_id": s.ID,
						"blocked_by": missingDeps,
					})
				}
			}
		}
		if len(blocked) > 0 {
			data["blocked_sessions"] = blocked
		}
	}

	var next *NextAction
	var nextStep string
	if completedIDs[session.ID] {
		nextStep = fmt.Sprintf("Session %s is already completed. Run vivechak_next without session_id for the next actionable session.", session.ID)
		next = &NextAction{
			Tool: "vivechak_next",
			Why:  nextStep,
		}
	} else {
		if core.IsSynthesisSession(session.ID) {
			advisoryPrefix := ""
			for _, w := range warnings {
				if strings.HasPrefix(w, "PRE-SYNTHESIS ADVISORY:") {
					advisoryPrefix = "⚠️ " + w + "\n\n"
					break
				}
			}
			nextStep = fmt.Sprintf(
				"%sSynthesize findings from all %d completed sessions into session %s. "+
					"Every recommendation must trace to session evidence. Conflicts between "+
					"sessions must be surfaced, not silently averaged. "+
					"The Belief Evolution section below shows how priors were confirmed or "+
					"contradicted — use this to identify the research's strongest and weakest signals. "+
					"Save with vivechak_save(kind='session', id='%s', content=...).",
				advisoryPrefix, len(completedIDs), session.ID, session.ID)
		} else {
			nextStep = fmt.Sprintf(
				"Execute the prompt for session %s. Ground every significant claim "+
					"with current, verifiable evidence quality — an architect should be able to "+
					"make the decision based solely on your findings. "+
					"Save with vivechak_save(kind='session', id='%s', content=...).",
				session.ID, session.ID)
		}
		if len(otherReady) > 0 {
			parallelPrefix := fmt.Sprintf(
				"⚡ %d sessions ready in parallel (%s). ",
				len(otherReady)+1, strings.Join(append([]string{session.ID}, otherReady...), ", "))
			nextStep = parallelPrefix + nextStep
		}
		next = &NextAction{
			Tool: "vivechak_save",
			Args: map[string]any{
				"kind": "session",
				"id":   session.ID,
			},
			Why: nextStep,
		}
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Ready session %s: %s", session.ID, session.Title),
		Data:     data,
		Warnings: warnings,
		Next:     next,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// handleNextSession provides backward compatibility for unit tests calling handleNextSession.
func handleNextSession(ctx context.Context, req *sdkmcp.CallToolRequest, in NextSessionInput) (*sdkmcp.CallToolResult, Envelope, error) {
	return handleNext(ctx, req, in)
}
