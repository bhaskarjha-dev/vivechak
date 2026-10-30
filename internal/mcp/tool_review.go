package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ReviewInput holds the arguments for vivechak_review.
type ReviewInput struct {
	ProjectRoot   string `json:"project_root,omitempty"   jsonschema:"workspace root path (optional)"`
	Mode          string `json:"mode,omitempty"           jsonschema:"review mode: 'validate' (dry-run quality validation) | 'challenge' (adversarial red-team stress test)"`
	ChallengeType string `json:"challenge_type,omitempty"  jsonschema:"optional specific challenge angle: 'premortem' | 'falsification' | 'red_team'"`
	Target        string `json:"target,omitempty"         jsonschema:"target session ID (e.g. T1-01) or decision ID (e.g. D-001) or artifact path"`
	ID            string `json:"id,omitempty"             jsonschema:"alias for target (e.g. T1-01, D-001)"`
	SessionID     string `json:"session_id,omitempty"     jsonschema:"alias for target"`
	DecisionID    string `json:"decision_id,omitempty"    jsonschema:"alias for target"`
	ArtifactType  string `json:"artifact_type,omitempty"  jsonschema:"what to validate when mode='validate': 'session' | 'decision' | 'plan' | 'comparison' | 'fad' (omit with empty content for workspace validation)"`
	Content       string `json:"content,omitempty"        jsonschema:"content to validate (dry run); omit for workspace-wide validation or saved session validation"`
}

// ValidateInput provides backward compatibility for unit tests calling handleValidate.
type ValidateInput struct {
	ProjectRoot  string `json:"project_root,omitempty"`
	ArtifactType string `json:"artifact_type,omitempty"`
	Content      string `json:"content,omitempty"`
}

// ChallengeInput provides backward compatibility for unit tests calling handleChallenge.
type ChallengeInput struct {
	ProjectRoot   string `json:"project_root,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
	DecisionID    string `json:"decision_id,omitempty"`
	Mode          string `json:"mode,omitempty"`
	ChallengeType string `json:"challenge_type,omitempty"`
}

func registerReview(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_review",
			Title: "Review Research (Validate or Challenge)",
			Description: "Inspect research artifacts for quality validation or adversarial stress-testing. " +
				"Modes: 'validate' (runs 4-level quality ladder on content or entire workspace without saving), " +
				"or 'challenge' (generates adversarial red-team critique prompts for One-Way Doors, premortems, and disconfirming search). Read-only.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleReview,
	)
}

func handleReview(ctx context.Context, req *sdkmcp.CallToolRequest, in ReviewInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_review"

	mode := strings.ToLower(strings.TrimSpace(in.Mode))
	if in.ChallengeType != "" && mode == "" {
		mode = "challenge"
	}
	if mode == "challenge" || mode == "red_team" || mode == "audit" || mode == "premortem" {
		return handleReviewChallenge(ctx, in)
	}

	// Default to validate
	return handleReviewValidate(tool, in)
}

func handleReviewChallenge(_ context.Context, in ReviewInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_review"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start.")
	}

	target := strings.TrimSpace(in.Target)
	if target == "" && in.ID != "" {
		target = strings.TrimSpace(in.ID)
	}
	if target == "" && in.SessionID != "" {
		target = strings.TrimSpace(in.SessionID)
	}
	if target == "" && in.DecisionID != "" {
		target = strings.TrimSpace(in.DecisionID)
	}
	if target == "" {
		return ErrorResult(tool, fmt.Errorf("target (session_id or decision_id) is required for challenge mode"),
			"Provide the target ID to challenge (e.g. 'T1-01' or 'D-001').")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err), "Provide a valid workspace.")
	}
	defer ws.Close()

	sessionData, _, err := core.ReadSessionFileFS(ws, core.SessionsDir, target)
	if err != nil || sessionData == "" {
		found := false
		if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			idUpper := strings.ToUpper(target)
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
					continue
				}
				name := strings.TrimSuffix(e.Name(), ".md")
				nameUpper := strings.ToUpper(name)
				if nameUpper == idUpper || strings.HasPrefix(nameUpper, idUpper+"-") || strings.HasPrefix(nameUpper, idUpper+"_") {
					if dData, rErr := ws.ReadFile(filepath.Join(core.ResearchDir, e.Name())); rErr == nil && len(dData) > 0 {
						sessionData = string(dData)
						found = true
						break
					}
				}
			}
		}
		if !found {
			return ErrorResult(tool, fmt.Errorf("target %s not found in %s or %s", target, core.SessionsDir, core.ResearchDir),
				"Ensure the session or decision file exists before challenging.")
		}
	}

	allSessions := make(map[string][]byte)
	if entries, lErr := ws.ListDir(core.SessionsDir); lErr == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
				continue
			}
			data, rErr := ws.ReadFile(filepath.Join(core.SessionsDir, e.Name()))
			if rErr != nil || len(data) == 0 {
				continue
			}
			sid := strings.TrimSuffix(e.Name(), ".md")
			if fm, _, pErr := core.ParseFrontmatter(data); pErr == nil && fm != nil {
				if idVal := fm.GetString("id"); idVal != "" {
					sid = idVal
				} else if sessID := fm.GetString("session_id"); sessID != "" {
					sid = sessID
				}
			}
			allSessions[sid] = data
		}
	}

	challengeResult, cErr := core.GenerateChallenges([]byte(sessionData), target, core.ModeRedTeam, allSessions)
	if cErr != nil {
		return ErrorResult(tool, cErr, "Verify session content is non-empty and formatted correctly.")
	}

	// Persist challenge record into research/.challenges/
	rec := core.ChallengeRecord{
		TargetID:   target,
		TargetHash: core.ComputeContentHash([]byte(sessionData)),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Mode:       string(core.ModeRedTeam),
	}
	for _, p := range challengeResult.Prompts {
		rec.Prompts = append(rec.Prompts, p.Prompt)
	}
	challengeRef, _ := core.SaveChallengeRecord(ws, rec)
	if challengeResult.Context == nil {
		challengeResult.Context = make(map[string]any)
	}
	challengeResult.Context["challenge_ref"] = challengeRef
	challengeResult.Context["target_hash"] = rec.TargetHash

	next := &NextAction{
		Tool: "vivechak_amend",
		Args: map[string]any{"session_id": target},
		Why:  fmt.Sprintf("Execute the adversarial challenge with web search, then append findings using vivechak_amend(session_id='%s', amendment=...).", target),
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Generated %d red-team challenge prompts for %s.", len(challengeResult.Prompts), target),
		Data:     challengeResult,
		Next:     next,
		NextStep: next.Why,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

func handleReviewValidate(tool string, in ReviewInput) (*sdkmcp.CallToolResult, Envelope, error) {
	if strings.TrimSpace(in.Content) == "" {
		return handleWorkspaceValidate(tool, in.ProjectRoot)
	}

	artifactType := in.ArtifactType
	if artifactType == "" {
		artifactType = "session"
	}

	var validation *core.ValidationResult
	switch artifactType {
	case "session":
		validation = core.ValidateSession([]byte(in.Content))
	case "decision", "decisions":
		contentBytes := []byte(in.Content)
		if artifactType == "decisions" || strings.Contains(in.Content, "<!-- DECISION:") || len(core.ExtractUnanchoredDecisionBlocks(contentBytes)) > 1 || (strings.Contains(in.Content, "#") && !strings.HasPrefix(strings.TrimSpace(in.Content), "---")) {
			validation = core.ValidateDecisionRegistry(contentBytes)
		} else {
			validation = core.ValidateDecision(contentBytes)
		}
	case "comparison", "comp":
		validation = core.ValidateComparison([]byte(in.Content))
	case "plan":
		validation = core.ValidatePlan([]byte(in.Content))
	case "fad":
		validation = core.ValidateFAD([]byte(in.Content))
	case "conflict", "conflict-resolution":
		validation = core.ValidateConflictResolution([]byte(in.Content))
	default:
		return ErrorResult(tool, fmt.Errorf("invalid artifact_type %q", in.ArtifactType),
			"Use 'session', 'decision', 'comparison', 'plan', 'fad', or 'conflict'.")
	}

	var issuesList []map[string]any
	var warnings []string
	blockingCount := 0
	for _, issue := range validation.Issues {
		if issue.Level == core.L2Block || issue.Level == core.L4Gate {
			blockingCount++
		}
		issuesList = append(issuesList, map[string]any{
			"level":   issue.Level.String(),
			"code":    issue.Code,
			"field":   issue.Field,
			"message": issue.Message,
			"fix":     issue.FixHint,
		})
		warnings = append(warnings, issue.String())
	}

	var observations []core.QualityObservation
	if artifactType == "session" {
		observations = core.ObserveSessionQuality([]byte(in.Content))
	}

	data := map[string]any{
		"artifact_type":  artifactType,
		"status":         validation.Status,
		"valid":          !validation.HasBlocking(),
		"blocking_count": blockingCount,
		"error_count":    validation.ErrorCount(),
		"warning_count":  validation.WarningCount(),
		"issues":         issuesList,
	}
	if len(observations) > 0 {
		data["observations"] = observations
	}

	next := &NextAction{
		Why: "Review validation output. Correct any blocking defects before saving.",
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Validation complete: %s (%d blocking, %d warnings).", validation.Status, blockingCount, validation.WarningCount()),
		Data:     data,
		Warnings: warnings,
		Next:     next,
		NextStep: next.Why,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

func handleWorkspaceValidate(tool string, projectRoot string) (*sdkmcp.CallToolResult, Envelope, error) {
	root, err := core.ResolveWorkspace(projectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start.")
	}

	if !core.WorkspaceExists(root) {
		return ErrorResult(tool, fmt.Errorf("workspace not initialized at %s", root),
			"Run vivechak_start first.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err),
			"Provide a valid workspace.")
	}
	defer ws.Close()

	var allWarnings []string
	results := map[string]any{}

	// Validate sessions
	sessionsValid := 0
	sessionsWarnings := 0
	sessionsErrors := 0
	var dag *core.DAG
	if pipeData, err := ws.ReadFile(core.PipelineFile); err == nil {
		dag, _ = core.ParsePipeline(pipeData)
	}
	if entries, err := ws.ListDir(core.SessionsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
				continue
			}
			stem := strings.TrimSuffix(e.Name(), ".md")
			if core.IsSynthesisSession(stem) {
				continue // Skip synthesis sessions — validated separately as FAD
			}
			relPath := filepath.Join(core.SessionsDir, e.Name())
			if data, err := ws.ReadFile(relPath); err == nil {
				sessionID := stem
				if fm, _, err := core.ParseFrontmatter(data); err == nil && fm != nil {
					if id := fm.GetString("session_id"); id != "" {
						sessionID = id
					} else if id := fm.GetString("id"); id != "" {
						sessionID = id
					}
				}
				isOneWay := false
				if dag != nil {
					if s := dag.SessionByID(sessionID); s != nil {
						isOneWay = strings.EqualFold(s.DoorType, "one-way")
					}
				}
				v := core.ValidateSessionWithContext(data, isOneWay)
				if v.HasBlocking() {
					sessionsErrors++
					for _, issue := range v.BlockingIssues() {
						allWarnings = append(allWarnings, fmt.Sprintf("SESSION %s: %s", e.Name(), issue.String()))
					}
				} else if v.WarningCount() > 0 {
					sessionsWarnings++
				} else {
					sessionsValid++
				}
			}
		}
	}
	results["sessions"] = map[string]any{
		"valid":    sessionsValid,
		"warnings": sessionsWarnings,
		"errors":   sessionsErrors,
	}

	// Validate decisions: individual files in research/ are the canonical source.
	decisionsValid := 0
	decisionsWarnings := 0
	decisionsErrors := 0
	seenIDs := make(map[string]bool)

	if entries, err := ws.ListDir(core.ResearchDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") {
				continue
			}
			if core.IsSpecialResearchFile(name) {
				continue
			}
			isCandidate := strings.HasPrefix(strings.ToUpper(name), "D-")
			relPath := filepath.Join(core.ResearchDir, name)
			data, err := ws.ReadFile(relPath)
			if err != nil || len(data) == 0 {
				continue
			}
			if !isCandidate {
				if fm, _, err := core.ParseFrontmatter(data); err == nil && fm != nil {
					if !fm.Has("door_type") {
						continue // Not a decision record
					}
				} else {
					continue
				}
			}

			// Track ID for deduplication against DECISIONS.md
			if fm, _, err := core.ParseFrontmatter(data); err == nil && fm != nil {
				id := fm.GetString("id")
				if id == "" {
					id = fm.GetString("decision_id")
				}
				if id != "" {
					seenIDs[strings.ToUpper(id)] = true
				}
			}
			baseName := strings.TrimSuffix(name, ".md")
			seenIDs[strings.ToUpper(baseName)] = true
			if m := regexp.MustCompile(`^(?i)(D-[A-Za-z0-9]+)`).FindStringSubmatch(baseName); len(m) > 1 {
				seenIDs[strings.ToUpper(m[1])] = true
			}
			if idx := strings.Index(baseName, "-"); idx != -1 {
				if nextIdx := strings.Index(baseName[idx+1:], "-"); nextIdx != -1 {
					shortID := baseName[:idx+1+nextIdx]
					seenIDs[strings.ToUpper(shortID)] = true
				}
			}

			v := core.ValidateDecision(data)
			if v.HasBlocking() {
				decisionsErrors++
				for _, issue := range v.BlockingIssues() {
					allWarnings = append(allWarnings, fmt.Sprintf("DECISION %s: %s", name, issue.String()))
				}
			} else if v.WarningCount() > 0 {
				decisionsWarnings++
			} else {
				decisionsValid++
			}
		}
	}

	// Also validate decisions in consolidated DECISIONS.md (deduplicating already validated IDs)
	if data, err := ws.ReadFile(core.DecisionsFile); err == nil && len(data) > 0 {
		anchoredRe := regexp.MustCompile(`(?s)<!--\s*DECISION:\s*([A-Za-z0-9_-]+)\s*-->\s*(.*?)\s*<!--\s*/DECISION:\s*([A-Za-z0-9_-]+)\s*-->`)

		matches := anchoredRe.FindAllSubmatch(data, -1)
		for _, m := range matches {
			id := string(m[1])
			closeID := string(m[3])
			if !strings.EqualFold(id, closeID) {
				continue
			}
			if seenIDs[strings.ToUpper(id)] {
				continue
			}
			seenIDs[strings.ToUpper(id)] = true
			v := core.ValidateDecision(m[2])
			if v.HasBlocking() {
				decisionsErrors++
				for _, issue := range v.BlockingIssues() {
					allWarnings = append(allWarnings, fmt.Sprintf("DECISION %s: %s", id, issue.String()))
				}
			} else if v.WarningCount() > 0 {
				decisionsWarnings++
			} else {
				decisionsValid++
			}
		}

		for _, block := range core.ExtractUnanchoredDecisionBlocks(data) {
			id := ""
			if fm, _, err := core.ParseFrontmatter(block); err == nil && fm != nil {
				id = fm.GetString("id")
				if id == "" {
					id = fm.GetString("decision_id")
				}
			}
			if id == "" {
				lines := strings.Split(string(block), "\n")
				for _, line := range lines {
					tr := strings.TrimSpace(line)
					if strings.HasPrefix(tr, "id:") || strings.HasPrefix(tr, "decision_id:") {
						parts := strings.SplitN(tr, ":", 2)
						if len(parts) == 2 {
							id = strings.TrimSpace(parts[1])
							id = strings.Trim(id, `"'`)
							break
						}
					}
				}
			}
			if id != "" && !seenIDs[strings.ToUpper(id)] && !strings.HasPrefix(strings.ToUpper(id), "CHK-") {
				seenIDs[strings.ToUpper(id)] = true
				v := core.ValidateDecision(block)
				if v.HasBlocking() {
					decisionsErrors++
					for _, issue := range v.BlockingIssues() {
						allWarnings = append(allWarnings, fmt.Sprintf("DECISION %s: %s", id, issue.String()))
					}
				} else if v.WarningCount() > 0 {
					decisionsWarnings++
				} else {
					decisionsValid++
				}
			}
		}
	}

	results["decisions"] = map[string]any{
		"valid":    decisionsValid,
		"warnings": decisionsWarnings,
		"errors":   decisionsErrors,
	}

	// Validate FAD if it exists
	if data, err := ws.ReadFile(core.FADFile); err == nil && len(data) > 0 {
		v := core.ValidateFAD(data)
		fadResult := map[string]any{
			"status":        v.Status,
			"error_count":   v.ErrorCount(),
			"warning_count": v.WarningCount(),
		}
		if v.HasBlocking() {
			for _, issue := range v.BlockingIssues() {
				allWarnings = append(allWarnings, fmt.Sprintf("FAD.md: %s", issue.String()))
			}
		}
		results["fad"] = fadResult
	}

	// Validate pipeline or decision plan if it exists
	planFile := core.PipelineFile
	planData, err := ws.ReadFile(planFile)
	if err != nil || len(planData) == 0 {
		if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), "-plan.md") {
					planFile = filepath.Join(core.ResearchDir, e.Name())
					planData, _ = ws.ReadFile(planFile)
					break
				}
			}
		}
	}
	if len(planData) > 0 {
		v := core.ValidatePlan(planData)
		pipeResult := map[string]any{
			"status":        v.Status,
			"error_count":   v.ErrorCount(),
			"warning_count": v.WarningCount(),
		}
		if v.HasBlocking() {
			for _, issue := range v.BlockingIssues() {
				allWarnings = append(allWarnings, fmt.Sprintf("PLAN %s: %s", filepath.Base(planFile), issue.String()))
			}
		}
		results["pipeline"] = pipeResult
	}

	// Validate comparisons
	var compResults []map[string]any
	if entries, lErr := ws.ListDir(core.ResearchDir); lErr == nil {
		for _, e := range entries {
			name := e.Name()
			if !e.IsDir() && strings.HasSuffix(name, ".md") &&
				(strings.HasPrefix(strings.ToUpper(name), "C-") ||
					strings.HasSuffix(name, "-comparison.md") ||
					strings.Contains(name, "-cmp-") ||
					strings.EqualFold(name, "COMPARISON.md")) {
				compPath := filepath.Join(core.ResearchDir, name)
				if compData, rErr := ws.ReadFile(compPath); rErr == nil && len(compData) > 0 {
					v := core.ValidateComparison(compData)
					compResult := map[string]any{
						"file":          name,
						"status":        v.Status,
						"error_count":   v.ErrorCount(),
						"warning_count": v.WarningCount(),
					}
					if v.HasBlocking() {
						for _, issue := range v.BlockingIssues() {
							allWarnings = append(allWarnings, fmt.Sprintf("COMPARISON %s: %s", name, issue.String()))
						}
					}
					compResults = append(compResults, compResult)
				}
			}
		}
	}
	if len(compResults) > 0 {
		results["comparisons"] = compResults
		results["comparison"] = compResults[0]
	}

	totalIssues := len(allWarnings)
	var nextStep string
	if totalIssues > 0 {
		nextStep = fmt.Sprintf("Workspace has %d issue(s). Review the warnings above and address any blocking issues.", totalIssues)
	} else {
		nextStep = "Workspace validation passed. All artifacts are structurally valid."
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Workspace validation complete (%d issue(s) found)", totalIssues),
		Data:     results,
		Warnings: allWarnings,
		Next: &NextAction{
			Why: nextStep,
		},
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// handleValidate provides backward compatibility for unit tests calling handleValidate.
func handleValidate(ctx context.Context, req *sdkmcp.CallToolRequest, in ValidateInput) (*sdkmcp.CallToolResult, Envelope, error) {
	return handleReview(ctx, req, ReviewInput{
		ProjectRoot:  in.ProjectRoot,
		Mode:         "validate",
		ArtifactType: in.ArtifactType,
		Content:      in.Content,
	})
}

// handleChallenge provides backward compatibility for unit tests calling handleChallenge.
func handleChallenge(ctx context.Context, req *sdkmcp.CallToolRequest, in ChallengeInput) (*sdkmcp.CallToolResult, Envelope, error) {
	target := in.SessionID
	if target == "" {
		target = in.DecisionID
	}
	return handleReview(ctx, req, ReviewInput{
		ProjectRoot: in.ProjectRoot,
		Mode:        "challenge",
		Target:      target,
	})
}
