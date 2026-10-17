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

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err),
			"Provide a valid workspace.")
	}
	defer ws.Close()

	info := core.InspectWorkspace(root)

	// If workspace is comparison scope, there is no multi-session pipeline DAG
	if info.Scope == core.ScopeComparison {
		var compFile string
		var compContent []byte
		if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				if !e.IsDir() && (strings.HasSuffix(e.Name(), "-comparison.md") || strings.Contains(e.Name(), "-cmp-")) {
					compFile = filepath.Join(core.ResearchDir, e.Name())
					if data, rErr := ws.ReadFile(compFile); rErr == nil {
						compContent = data
						break
					}
				}
			}
		}

		compSessionID := core.SessionPrefix(info.Scope) + "01" // C-01
		if strings.Contains(compFile, "CMP-") || strings.Contains(string(compContent), "CMP-01") || strings.Contains(compFile, "COMP-") {
			compSessionID = "CMP-01"
		} else if idMatch := regexp.MustCompile(`(?m)^\|\s*\*\*ID\*\*\s*\|\s*([A-Za-z0-9_-]+)\s*\|`).FindSubmatch(compContent); len(idMatch) > 1 {
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
			Success: true,
			Message: msg,
			Data:    dataMap,
			NextStep: fmt.Sprintf("Execute your comparison research prompt. Ground claims with current, "+
				"verifiable evidence. Save the output with vivechak_save_session using session_id=%q.", compSessionID),
			Meta: NewMeta(tool),
		}
		return env.ToResult()
	}

	// Read pipeline file (or decision plan file)
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
			return ErrorResult(tool,
				fmt.Errorf("no research pipeline found — run vivechak_save_plan first"),
				"Run vivechak_prepare_generator to get the generator prompt, execute it, "+
					"then save with vivechak_save_plan.")
		}
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
	var knownIDs []string
	if dag != nil {
		for _, s := range dag.Sessions {
			knownIDs = append(knownIDs, s.ID)
		}
	}
	completedIDs, err := scanCompletedSessions(ws, knownIDs...)
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
			return buildSessionResponse(tool, *session, "", completedIDs, dag, in.Verbose, nil, ws)
		}
		var injWarnings []string
		if injected.Warning != "" {
			injWarnings = append(injWarnings, injected.Warning)
		}
		if core.IsSynthesisSession(session.ID) {
			if adv := checkPreSynthesisAdvisory(dag, completedIDs, sessionsDir); adv != "" {
				injWarnings = append(injWarnings, adv)
			}
		}
		return buildSessionResponse(tool, injected.Session, injected.InjectedPrompt, completedIDs, dag, in.Verbose, injWarnings, ws)
	}

	// Find next actionable sessions
	ready := dag.NextSessions(completedIDs)

	if len(ready) == 0 {
		// Count completed sessions that belong to the DAG
		completedCount := 0
		for _, s := range dag.Sessions {
			if completedIDs[s.ID] {
				completedCount++
			}
		}

		// All DAG sessions complete
		if completedCount == len(dag.Sessions) {
			env := Envelope{
				Success: true,
				Message: fmt.Sprintf("All %d sessions complete!", len(dag.Sessions)),
				Data: map[string]any{
					"workspace_root":     root,
					"total_sessions":     len(dag.Sessions),
					"completed_sessions": completedCount,
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
					missingDeps = append(missingDeps, formatMissingDep(ws, dep))
				}
			}
			blockedInfo = append(blockedInfo, map[string]any{
				"session_id": s.ID,
				"blocked_by": missingDeps,
			})
		}

		env := Envelope{
			Success: true,
			Message: fmt.Sprintf("%d sessions remaining but all blocked by dependencies", len(dag.Sessions)-completedCount),
			Data: map[string]any{
				"workspace_root":     root,
				"total_sessions":     len(dag.Sessions),
				"completed_sessions": completedCount,
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
	injected, injErr := core.InjectContext(nextSession, sessionsDir, completedIDs)

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
				"W-INJECT-FAILED: context injection failed for %s: %v \u2014 using raw prompt without upstream findings",
				nextSession.ID, injErr))
		}
	}

	// Build response with all ready sessions listed
	var otherReady []string
	for _, s := range ready[1:] {
		otherReady = append(otherReady, s.ID)
	}

	// Detect layer transition and add replan prompt
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
					"remaining sessions should be adjusted based on Layer %d findings. If a session "+
					"is no longer relevant, you may skip it by saving an empty session with a note "+
					"explaining why.", highestCompletedLayer, highestCompletedLayer))
		}
	}

	// Detect pre-synthesis evidentiary defects on One-Way Doors
	if core.IsSynthesisSession(nextSession.ID) {
		if adv := checkPreSynthesisAdvisory(dag, completedIDs, sessionsDir); adv != "" {
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

func checkPreSynthesisAdvisory(dag *core.DAG, completedIDs map[string]bool, sessionsDir string) string {
	if dag == nil {
		return ""
	}
	var owdIssues []string
	recalledCitationRe := regexp.MustCompile(`(?i)(?:\[[^\]\n]*\b(?:recalled|parametric)\b[^\]\n]*\]|\([^)\n]*\b(?:recalled|parametric)\b[^)\n]*\)|\brecalled\s+memory\b|\|\s*recalled\s*\|)`)

	for _, s := range dag.Sessions {
		if strings.EqualFold(s.DoorType, "one-way") && completedIDs[s.ID] {
			if sData, _, err := core.ReadSessionFilePublic(sessionsDir, s.ID); err == nil && sData != "" {
				// Check for recalled verification in recommendation
				rec := core.ExtractSection(sData, "recommend", "verdict", "decision")
				if len(rec) >= 20 && recalledCitationRe.MatchString(rec) {
					owdIssues = append(owdIssues, fmt.Sprintf("%s (recommendation relies on Grade D recalled claims)", s.ID))
				}
				// Check for challenge section
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
				"You may proceed with synthesis now or amend weak sessions first with vivechak_amend_session.",
			len(owdIssues), strings.Join(owdIssues, "; "))
	}
	return ""
}

const calibrationBlock = "\n\n---\n### RESEARCH CALIBRATION (Active Rules)\n" +
	"1. **Memory \u2260 Evidence:** Claims recalled from model weights cap at Grade D. Verify with web search.\n" +
	"2. **Upstream \u2260 Exhaustive:** Injected upstream findings provide high-level architectural context, NOT implementation specifications. Exact schemas, library bundle sizes, API rate limits, and security configurations must be verified with current external sources.\n" +
	"3. **Qualified Provenance:** Grade A claims require canonical URLs, domain anchors, RFCs, or doc titles.\n" +
	"4. **Falsification Required:** Actively search for failure modes, gotchas, and production incident postmortems.\n" +
	"5. **Honest Gaps:** Explicitly stating \"searched X and found no evidence\" is valued. Never hallucinate sources.\n"

// buildSessionResponse creates the response envelope for a session.
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

	// Inject research calibration rules into non-synthesis session prompts
	if !core.IsSynthesisSession(session.ID) && prompt != "" && !strings.Contains(prompt, "### RESEARCH CALIBRATION") {
		prompt += calibrationBlock
	}

	// Progressive disclosure: truncate if prompt is very large
	var warnings []string
	warnings = append(warnings, preWarnings...)
	displayPrompt := prompt
	approxTokens := len(prompt) / 4
	truncated := false
	if !verbose && approxTokens > 10000 && !core.IsSynthesisSession(session.ID) {
		warnings = append(warnings, "W-OUTPUT-SIZE: Prompt exceeds 10K tokens and was truncated. Use verbose=true for full prompt.")
		truncated = true
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

	// Add blocked sessions with their blockers for orchestrator awareness
	if dag != nil {
		var blocked []map[string]any
		for _, s := range dag.Sessions {
			if completedIDs[s.ID] || s.ID == session.ID {
				continue
			}
			// Check if this session is blocked (not in ready list)
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

	var nextStep string
	if completedIDs[session.ID] {
		nextStep = fmt.Sprintf("Session %s is already completed. Run vivechak_next_session without session_id for the next actionable session.", session.ID)
	} else {
		if core.IsSynthesisSession(session.ID) {
			nextStep = fmt.Sprintf(
				"Synthesize findings from all %d completed sessions into session %s. "+
					"Every recommendation must trace to session evidence. Conflicts between "+
					"sessions must be surfaced, not silently averaged. "+
					"The Belief Evolution section below shows how priors were confirmed or "+
					"contradicted — use this to identify the research's strongest and weakest signals. "+
					"Save with vivechak_save_session using session_id='%s'.",
				len(completedIDs), session.ID, session.ID)
			for _, w := range preWarnings {
				if strings.HasPrefix(w, "PRE-SYNTHESIS ADVISORY") {
					nextStep = "⚠️ " + w + "\n\n" + nextStep
					break
				}
			}
		} else {
			nextStep = fmt.Sprintf(
				"Execute the prompt for session %s. Ground every significant claim "+
					"with current, verifiable evidence — an architect should be able to "+
					"make the decision based solely on your findings. "+
					"Save with vivechak_save_session using session_id='%s'.",
				session.ID, session.ID)
		}
		if len(otherReady) > 0 {
			parallelPrefix := fmt.Sprintf(
				"⚡ %d sessions ready in parallel (%s). ",
				len(otherReady)+1, strings.Join(append([]string{session.ID}, otherReady...), ", "))
			nextStep = parallelPrefix + nextStep
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

// scanCompletedSessions forwards to core.ScanCompletedSessions.
func scanCompletedSessions(ws *store.Workspace, knownIDs ...string) (map[string]bool, error) {
	return core.ScanCompletedSessions(ws, knownIDs...)
}

// sessionIDList returns a comma-separated list of session IDs in the DAG.
func sessionIDList(dag *core.DAG) string {
	ids := make([]string, len(dag.Sessions))
	for i, s := range dag.Sessions {
		ids[i] = s.ID
	}
	return strings.Join(ids, ", ")
}
