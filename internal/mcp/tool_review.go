package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ReviewInput holds the canonical arguments for vivechak_review.
type ReviewInput struct {
	ProjectRoot  string `json:"project_root,omitempty"`
	Mode         string `json:"mode"`
	ID           string `json:"id,omitempty"`
	Content      string `json:"content,omitempty"`
	ArtifactType string `json:"artifact_type,omitempty"`
}

var (
	reDecisionBaseName = regexp.MustCompile(`^(?i)(D-[A-Za-z0-9]+)`)
	reDecisionAnchored = regexp.MustCompile(`(?s)<!--\s*DECISION:\s*([A-Za-z0-9_-]+)\s*-->\s*(.*?)\s*<!--\s*/DECISION:\s*([A-Za-z0-9_-]+)\s*-->`)
)

func registerReview(server *sdkmcp.Server) {
	registerTool(server, ToolSpec{
		Name:        "vivechak_review",
		Description: "Check research: validate (content or id), evidence (citation audit), red_team (writes challenge record), or consistency.",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint:   false,
			IdempotentHint: false,
		},
		Params: []ParamSpec{
			{Name: "mode", Type: "string", Enum: []string{"validate", "evidence", "red_team", "consistency"}, Required: true, Doc: "Review mode: validate, evidence, red_team, or consistency."},
			{Name: "id", Type: "string", Description: "Target ID", Doc: "Target session, decision, comparison, or plan ID."},
			{Name: "content", Type: "string", Description: "Inline Markdown", Doc: "Inline Markdown content for dry-run validate mode."},
			{Name: "artifact_type", Type: "string", Enum: []string{"session", "decision", "comparison", "plan", "fad", "conflict-resolution"}, Doc: "Artifact type when validating inline content."},
			{Name: "project_root", Type: "string", Doc: "Absolute path to the target project directory."},
		},
	}, handleReview)
}

func handleReview(ctx context.Context, req *sdkmcp.CallToolRequest, in ReviewInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_review"

	mode := strings.ToLower(strings.TrimSpace(in.Mode))
	switch mode {
	case "", "validate":
		return handleReviewValidate(ctx, req, tool, in)
	case "red_team", "challenge":
		return handleReviewAudit(ctx, req, in, core.ModeRedTeam, true)
	case "evidence":
		return handleReviewAudit(ctx, req, in, core.ModeEvidenceAudit, false)
	case "consistency":
		return handleReviewAudit(ctx, req, in, core.ModeCrossSession, false)
	default:
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: invalid mode %q: use 'validate', 'evidence', 'red_team', or 'consistency'", in.Mode),
			"Specify mode 'validate', 'evidence', 'red_team', or 'consistency'.")
	}
}

func readArtifactByID(ws *store.Workspace, target string) (data string, inferredType string, found bool) {
	if isSynthesisTarget(ws, target) || strings.EqualFold(target, "FAD") {
		if d, err := ws.ReadFile(core.FADFile); err == nil && len(d) > 0 {
			return string(d), "fad", true
		}
	}
	if strings.EqualFold(target, "DAG") || strings.EqualFold(target, "PROJECT") {
		if d, err := ws.ReadFile(core.PipelineFile); err == nil && len(d) > 0 {
			return string(d), "plan", true
		}
	}
	if sData, _, err := core.ReadSessionFileFS(ws, core.SessionsDir, target); err == nil && sData != "" {
		return sData, "session", true
	}
	normID := core.NormalizeDecisionID(target)
	for _, dir := range []string{core.ResearchDir, core.DecisionsDir, core.ComparisonsDir} {
		entries, err := ws.ListDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			lowerName := strings.ToLower(e.Name())
			if e.IsDir() || !strings.HasSuffix(lowerName, ".md") ||
				lowerName == "research-pipeline.md" || lowerName == "decisions.md" || lowerName == "phase-0-gate.md" || lowerName == "fad.md" {
				continue
			}
			stem := strings.TrimSuffix(e.Name(), ".md")
			stemUpper := strings.ToUpper(stem)
			if reSessionOrReview.MatchString(stemUpper) {
				continue
			}
			if strings.HasSuffix(lowerName, "-plan.md") {
				continue
			}
			tUpper := strings.ToUpper(target)
			nUpper := strings.ToUpper(normID)
			if stemUpper == tUpper || strings.HasPrefix(stemUpper, tUpper+"-") || strings.HasPrefix(stemUpper, tUpper+"_") ||
				(nUpper != "" && (stemUpper == nUpper || strings.HasPrefix(stemUpper, nUpper+"-") || strings.HasPrefix(stemUpper, nUpper+"_"))) {
				raw, rErr := ws.ReadFile(filepath.Join(dir, e.Name()))
				if rErr != nil || len(raw) == 0 {
					continue
				}
				lower := strings.ToLower(e.Name())
				switch {
				case strings.HasPrefix(stemUpper, "C-") || strings.HasSuffix(lower, "-comparison.md") || strings.Contains(lower, "-cmp-"):
					return string(raw), "comparison", true
				case strings.HasSuffix(lower, "-conflict-resolution.md"):
					return string(raw), "conflict-resolution", true
				default:
					return string(raw), "decision", true
				}
			}
		}
	}
	// Fallback 1: check DECISIONS.md anchor
	if decData, rErr := ws.ReadFile(core.DecisionsFile); rErr == nil {
		matches := reDecisionAnchored.FindAllStringSubmatch(string(decData), -1)
		for _, m := range matches {
			if len(m) > 1 && (strings.EqualFold(m[1], target) || strings.EqualFold(core.NormalizeDecisionID(m[1]), normID)) {
				return m[0], "decision", true
			}
		}
	}
	// Fallback 2: check sessions that inform this decision
	if normID != "" {
		if entries, lErr := ws.ListDir(core.SessionsDir); lErr == nil {
			for _, e := range entries {
				if strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
					data, rErr := ws.ReadFile(filepath.Join(core.SessionsDir, e.Name()))
					if rErr == nil {
						fm, _, _ := core.ParseFrontmatter(data)
						if fm != nil {
							informs, _ := fm["informs_decisions"].([]any)
							for _, inf := range informs {
								if s, ok := inf.(string); ok && core.NormalizeDecisionID(s) == normID {
									return string(data), "session", true
								}
							}
						}
					}
				}
			}
		}
	}
	return "", "", false
}

func loadAllSessionsMap(ws *store.Workspace) map[string][]byte {
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
	return allSessions
}

func handleReviewAudit(ctx context.Context, req *sdkmcp.CallToolRequest, in ReviewInput, chMode core.ChallengeMode, writeRecord bool) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_review"

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
		return ErrorResult(tool, fmt.Errorf("E-IO: opening workspace: %w", err), "Provide a valid workspace.")
	}
	defer ws.Close()

	allSessions := loadAllSessionsMap(ws)

	target := strings.TrimSpace(in.ID)
	if target == "" && chMode == core.ModeCrossSession && len(allSessions) > 0 {
		var ids []string
		for k := range allSessions {
			ids = append(ids, k)
		}
		sort.Strings(ids)
		target = ids[0]
	}
	if target == "" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: id is required for %s mode", in.Mode),
			"Provide the target session or decision ID in id (e.g. 'T1-01' or 'D-001').")
	}

	var idWarnings []Warning
	if resolved, mapped, idErr := resolvePlaceholderRef(sessionOf(req), ws, root, target, true); idErr != nil {
		return ErrorResult(tool, idErr, placeholderRefFix)
	} else if mapped {
		idWarnings = append(idWarnings, NewWarning("W-ID-ALLOCATED", fmt.Sprintf("id %s resolved to %s", target, resolved), ""))
		target = resolved
	}

	sessionData := strings.TrimSpace(in.Content)
	if sessionData == "" {
		var found bool
		var artType string
		sessionData, artType, found = readArtifactByID(ws, target)
		if !found || artType == "plan" {
			return ErrorResult(tool, fmt.Errorf("E-NOT-FOUND: target %s not found in %s or %s", target, core.SessionsDir, core.ResearchDir),
				"Ensure the session or decision file exists before running review.")
		}
	}

	challengeResult, cErr := core.GenerateChallenges([]byte(sessionData), target, chMode, allSessions)
	if cErr != nil {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: %w", cErr), "Verify artifact content is non-empty and formatted correctly.")
	}

	if challengeResult.Context == nil {
		challengeResult.Context = make(map[string]any)
	}
	challengeResult.Context["workspace"] = map[string]any{"root": root, "rule": rule}

	if writeRecord {
		var promptStrings []string
		for _, p := range challengeResult.Prompts {
			promptStrings = append(promptStrings, p.Prompt)
		}

		var recordID string
		var decHash string
		if strings.HasPrefix(strings.ToUpper(target), "D-") && !reDecisionSessionID.MatchString(strings.ToUpper(target)) {
			recID, h, wErr := core.WriteDecisionChallengeRecord(ws, target, []byte(sessionData), promptStrings)
			if wErr == nil {
				recordID = recID
				decHash = h
			}
		} else {
			rec := core.ChallengeRecord{
				TargetID:   target,
				TargetHash: core.ComputeContentHash([]byte(sessionData)),
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				Mode:       string(core.ModeRedTeam),
				Prompts:    promptStrings,
			}
			recordID, _ = core.SaveChallengeRecord(ws, rec)
			decHash = rec.TargetHash
		}

		challengeResult.RecordID = recordID
		challengeResult.DecisionHash = decHash
		challengeResult.Context["challenge_ref"] = recordID
		challengeResult.Context["record_id"] = recordID
		challengeResult.Context["target_hash"] = decHash
		challengeResult.Context["decision_hash"] = decHash
	}

	var msg string
	var nextWhy string
	switch chMode {
	case core.ModeEvidenceAudit:
		msg = fmt.Sprintf("Generated %d citation-audit prompts for %s.", len(challengeResult.Prompts), target)
		nextWhy = fmt.Sprintf("Verify cited sources with web search and append any corrections via vivechak_amend(id=%q, note=...).", target)
	case core.ModeCrossSession:
		msg = fmt.Sprintf("Generated %d consistency-check prompts for %s across %d session(s).", len(challengeResult.Prompts), target, len(allSessions))
		nextWhy = fmt.Sprintf("Resolve cross-session conflicts via vivechak_amend(id=%q, note=...) or record a conflict-resolution artifact.", target)
	default:
		msg = fmt.Sprintf("Generated %d red-team challenge prompts for %s.", len(challengeResult.Prompts), target)
		nextWhy = fmt.Sprintf("Execute the adversarial challenge with web search, then append findings using vivechak_amend(id=%q, note=...).", target)
	}

	env := Envelope{
		Ok:       true,
		Message:  msg,
		Data:     challengeResult,
		Warnings: idWarnings,
		Next: &NextAction{
			Tool: "vivechak_amend",
			Args: map[string]any{"id": target},
			Why:  nextWhy,
		},
	}
	return env.ToResult()
}

func saveKindForArtifactType(artifactType string) string {
	switch artifactType {
	case "comparison":
		return "comparison"
	case "plan":
		return "plan"
	case "fad":
		return "fad"
	default:
		return "session"
	}
}

func handleReviewValidate(ctx context.Context, req *sdkmcp.CallToolRequest, tool string, in ReviewInput) (*sdkmcp.CallToolResult, Envelope, error) {
	content := in.Content
	artifactType := strings.TrimSpace(in.ArtifactType)
	var wsInfo map[string]any
	var idWarnings []Warning

	if strings.TrimSpace(content) == "" {
		if strings.TrimSpace(in.ID) == "" {
			return handleWorkspaceValidate(ctx, req, tool, in.ProjectRoot)
		}
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
			return ErrorResult(tool, fmt.Errorf("E-IO: opening workspace: %w", err), "Provide a valid workspace.")
		}
		defer ws.Close()

		target := strings.TrimSpace(in.ID)
		if resolved, mapped, idErr := resolvePlaceholderRef(sessionOf(req), ws, root, target, true); idErr != nil {
			return ErrorResult(tool, idErr, placeholderRefFix)
		} else if mapped {
			idWarnings = append(idWarnings, NewWarning("W-ID-ALLOCATED", fmt.Sprintf("id %s resolved to %s", target, resolved), ""))
			target = resolved
		}
		loaded, inferred, found := readArtifactByID(ws, target)
		if !found {
			return ErrorResult(tool, fmt.Errorf("E-NOT-FOUND: artifact %s not found in workspace", target),
				"Check the ID against existing artifacts with vivechak_status.")
		}
		content = loaded
		if artifactType == "" {
			artifactType = inferred
		}
		wsInfo = map[string]any{"root": root, "rule": rule}
	}

	if artifactType == "" {
		artifactType = "session"
	}

	var validation *core.ValidationResult
	switch artifactType {
	case "session":
		validation = core.ValidateSession([]byte(content))
	case "decision", "decisions":
		contentBytes := []byte(content)
		if artifactType == "decisions" || strings.Contains(content, "<!-- DECISION:") || len(core.ExtractUnanchoredDecisionBlocks(contentBytes)) > 1 || (strings.Contains(content, "#") && !strings.HasPrefix(strings.TrimSpace(content), "---")) {
			validation = core.ValidateDecisionRegistry(contentBytes)
		} else {
			validation = core.ValidateDecision(contentBytes)
		}
	case "comparison", "comp":
		validation = core.ValidateComparison([]byte(content))
	case "plan":
		validation = core.ValidatePlan([]byte(content))
	case "fad":
		validation = core.ValidateFAD([]byte(content))
	case "conflict", "conflict-resolution":
		validation = core.ValidateConflictResolution([]byte(content))
	default:
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: invalid artifact_type %q", in.ArtifactType),
			"Use 'session', 'decision', 'comparison', 'plan', 'fad', or 'conflict-resolution'.")
	}

	var issuesList []map[string]any
	warnings := append([]Warning{}, idWarnings...)
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
		warnings = append(warnings, WarningFromIssue(issue))
	}

	var observations []core.QualityObservation
	if artifactType == "session" {
		observations = core.ObserveSessionQuality([]byte(content))
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
	if wsInfo != nil {
		data["workspace"] = wsInfo
	}
	if len(observations) > 0 {
		data["observations"] = observations
	}

	next := &NextAction{
		Tool: "vivechak_save",
		Args: map[string]any{"kind": saveKindForArtifactType(artifactType)},
		Why:  "Review any validation findings, then persist the artifact with vivechak_save.",
	}
	if artifactType == "decision" || artifactType == "conflict-resolution" {
		next = &NextAction{
			Tool: "vivechak_record_decision",
			Args: map[string]any{"id": strings.TrimSpace(in.ID)},
			Why:  "Review any validation findings, then record the decision with vivechak_record_decision.",
		}
		if strings.TrimSpace(in.ID) == "" {
			next.Args["id"] = "D-NEW"
		}
	} else if strings.TrimSpace(in.Content) == "" && !validation.HasBlocking() {
		next = &NextAction{
			Tool: "vivechak_gate",
			Args: map[string]any{},
			Why:  "Saved artifact is valid. Run vivechak_gate when ready to verify Phase 0 exit criteria.",
		}
	}

	env := Envelope{
		Ok:       true,
		Message:  fmt.Sprintf("Validation complete: %s (%d blocking, %d warnings).", validation.Status, blockingCount, validation.WarningCount()),
		Data:     data,
		Warnings: warnings,
		Next:     next,
	}
	return env.ToResult()
}

func handleWorkspaceValidate(ctx context.Context, req *sdkmcp.CallToolRequest, tool string, projectRoot string) (*sdkmcp.CallToolResult, Envelope, error) {
	var roots []string
	if strings.TrimSpace(projectRoot) == "" {
		roots = getClientRoots(ctx, req)
	}
	root, rule, err := core.ResolveWorkspaceWithRoots(ctx, projectRoot, roots)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start.")
	}

	if !core.WorkspaceExists(root) {
		return ErrorResult(tool, fmt.Errorf("E-NOT-INITIALIZED: workspace not initialized at %s", root),
			"Run vivechak_start first.")
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("E-IO: opening workspace: %w", err),
			"Provide a valid workspace.")
	}
	defer ws.Close()

	var allWarnings []Warning
	results := map[string]any{
		"workspace":      map[string]any{"root": root, "rule": rule},
		"workspace_root": root,
	}

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
			if dag.IsSynthesisID(stem) {
				continue
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
						w := WarningFromIssue(issue)
						w.Message = fmt.Sprintf("SESSION %s: %s", e.Name(), w.Message)
						allWarnings = append(allWarnings, w)
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

	decisionsValid := 0
	decisionsWarnings := 0
	decisionsErrors := 0
	seenIDs := make(map[string]bool)

	if entries, err := ws.ListDir(core.ResearchDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") || core.IsSpecialResearchFile(name) {
				continue
			}
			if strings.HasSuffix(strings.ToLower(name), "-plan.md") || reSessionOrReview.MatchString(strings.ToUpper(name)) {
				continue
			}
			isCandidate := strings.HasPrefix(strings.ToUpper(name), "D-") && isDecisionArtifact(ws, name)
			relPath := filepath.Join(core.ResearchDir, name)
			data, err := ws.ReadFile(relPath)
			if err != nil || len(data) == 0 {
				continue
			}
			if !isCandidate {
				if fm, _, err := core.ParseFrontmatter(data); err == nil && fm != nil {
					if !fm.Has("door_type") {
						continue
					}
				} else {
					continue
				}
			}

			if fm, _, err := core.ParseFrontmatter(data); err == nil && fm != nil {
				id := fm.GetString("id")
				if id == "" {
					id = fm.GetString("decision_id")
				}
				if id != "" {
					seenIDs[strings.ToUpper(id)] = true
					seenIDs[strings.ToUpper(core.NormalizeDecisionID(id))] = true
				}
			}
			baseName := strings.TrimSuffix(name, ".md")
			seenIDs[strings.ToUpper(baseName)] = true
			seenIDs[strings.ToUpper(core.NormalizeDecisionID(baseName))] = true
			if m := reDecisionBaseName.FindStringSubmatch(baseName); len(m) > 1 {
				seenIDs[strings.ToUpper(m[1])] = true
				seenIDs[strings.ToUpper(core.NormalizeDecisionID(m[1]))] = true
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
					w := WarningFromIssue(issue)
					w.Message = fmt.Sprintf("DECISION %s: %s", name, w.Message)
					allWarnings = append(allWarnings, w)
				}
			} else if v.WarningCount() > 0 {
				decisionsWarnings++
			} else {
				decisionsValid++
			}
		}
	}

	if data, err := ws.ReadFile(core.DecisionsFile); err == nil && len(data) > 0 {
		matches := reDecisionAnchored.FindAllSubmatch(data, -1)
		for _, m := range matches {
			id := string(m[1])
			closeID := string(m[3])
			if !strings.EqualFold(id, closeID) {
				continue
			}
			normID := strings.ToUpper(core.NormalizeDecisionID(id))
			if seenIDs[strings.ToUpper(id)] || seenIDs[normID] {
				continue
			}
			seenIDs[strings.ToUpper(id)] = true
			seenIDs[normID] = true
			v := core.ValidateDecision(m[2])
			if v.HasBlocking() {
				decisionsErrors++
				for _, issue := range v.BlockingIssues() {
					w := WarningFromIssue(issue)
					w.Message = fmt.Sprintf("DECISION %s: %s", id, w.Message)
					allWarnings = append(allWarnings, w)
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
				for _, line := range strings.Split(string(block), "\n") {
					tr := strings.TrimSpace(line)
					if strings.HasPrefix(tr, "id:") || strings.HasPrefix(tr, "decision_id:") {
						parts := strings.SplitN(tr, ":", 2)
						if len(parts) == 2 {
							id = strings.Trim(strings.TrimSpace(parts[1]), `"'`)
							break
						}
					}
				}
			}
			if id != "" && !seenIDs[strings.ToUpper(id)] && !seenIDs[strings.ToUpper(core.NormalizeDecisionID(id))] && !strings.HasPrefix(strings.ToUpper(id), "CHK-") {
				seenIDs[strings.ToUpper(id)] = true
				seenIDs[strings.ToUpper(core.NormalizeDecisionID(id))] = true
				v := core.ValidateDecision(block)
				if v.HasBlocking() {
					decisionsErrors++
					for _, issue := range v.BlockingIssues() {
						w := WarningFromIssue(issue)
						w.Message = fmt.Sprintf("DECISION %s: %s", id, w.Message)
						allWarnings = append(allWarnings, w)
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

	if data, err := ws.ReadFile(core.FADFile); err == nil && len(data) > 0 {
		v := core.ValidateFAD(data)
		results["fad"] = map[string]any{
			"status":        v.Status,
			"error_count":   v.ErrorCount(),
			"warning_count": v.WarningCount(),
		}
		if v.HasBlocking() {
			for _, issue := range v.BlockingIssues() {
				w := WarningFromIssue(issue)
				w.Message = fmt.Sprintf("FAD.md: %s", w.Message)
				allWarnings = append(allWarnings, w)
			}
		}
	}

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
		results["pipeline"] = map[string]any{
			"status":        v.Status,
			"error_count":   v.ErrorCount(),
			"warning_count": v.WarningCount(),
		}
		if v.HasBlocking() {
			for _, issue := range v.BlockingIssues() {
				w := WarningFromIssue(issue)
				w.Message = fmt.Sprintf("PLAN %s: %s", filepath.Base(planFile), w.Message)
				allWarnings = append(allWarnings, w)
			}
		}
	}

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
					compResults = append(compResults, map[string]any{
						"file":          name,
						"status":        v.Status,
						"error_count":   v.ErrorCount(),
						"warning_count": v.WarningCount(),
					})
					if v.HasBlocking() {
						for _, issue := range v.BlockingIssues() {
							w := WarningFromIssue(issue)
							w.Message = fmt.Sprintf("COMPARISON %s: %s", name, w.Message)
							allWarnings = append(allWarnings, w)
						}
					}
				}
			}
		}
	}
	if len(compResults) > 0 {
		results["comparisons"] = compResults
		results["comparison"] = compResults[0]
	}

	totalIssues := len(allWarnings)
	nextTool := "vivechak_gate"
	nextWhy := "Workspace validation passed. Run vivechak_gate to evaluate Phase 0 exit criteria."
	if totalIssues > 0 {
		nextTool = "vivechak_status"
		nextWhy = fmt.Sprintf("Workspace has %d blocking issue(s). Inspect warnings and re-save affected artifacts.", totalIssues)
	}

	env := Envelope{
		Ok:       true,
		Message:  fmt.Sprintf("Workspace validation complete (%d issue(s) found)", totalIssues),
		Data:     results,
		Warnings: allWarnings,
		Next: &NextAction{
			Tool: nextTool,
			Args: map[string]any{},
			Why:  nextWhy,
		},
	}
	return env.ToResult()
}
