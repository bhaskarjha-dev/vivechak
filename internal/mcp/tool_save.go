package mcputil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	vembed "github.com/bhaskarjha-dev/vivechak/internal/embed"
	store "github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SaveInput holds the arguments for vivechak_save.
type SaveInput struct {
	ProjectRoot      string `json:"project_root,omitempty"       jsonschema:"workspace root path (optional)"`
	Kind             string `json:"kind"                         jsonschema:"artifact kind (REQUIRED): 'plan' (RESEARCH-PIPELINE.md) | 'session' (research session) | 'comparison' (WEP matrix)"`
	Scope            string `json:"scope,omitempty"              jsonschema:"optional scope hint ('comparison' | 'decision' | 'project')"`
	ID               string `json:"id,omitempty"                 jsonschema:"identifier: session_id (e.g. T1-01, SYN-01) or comparison_id (e.g. C-001)"`
	Content          string `json:"content"                      jsonschema:"markdown artifact content to save"`
	DecisionsContent string `json:"decisions_content,omitempty" jsonschema:"initial decision registry content when saving plan (optional)"`
	Slug             string `json:"slug,omitempty"               jsonschema:"URL-safe slug for the artifact (optional)"`
}

// SavePlanInput provides backward compatibility for unit tests calling handleSavePlan.
type SavePlanInput struct {
	ProjectRoot      string `json:"project_root,omitempty"`
	Scope            string `json:"scope,omitempty"`
	Content          string `json:"content,omitempty"`
	PipelineContent  string `json:"pipeline_content,omitempty"`
	DecisionsContent string `json:"decisions_content,omitempty"`
	ID               string `json:"id,omitempty"`
	DecisionID       string `json:"decision_id,omitempty"`
	ComparisonID     string `json:"comparison_id,omitempty"`
	Slug             string `json:"slug,omitempty"`
}

// SaveSessionInput provides backward compatibility for unit tests calling handleSaveSession.
type SaveSessionInput struct {
	ProjectRoot string `json:"project_root,omitempty"`
	SessionID   string `json:"session_id"`
	Content     string `json:"content"`
}

func registerSave(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_save",
			Title: "Save Research Artifact",
			Description: "Validate and persist research artifacts into the workspace. " +
				"Pass 'kind': 'plan' (persists RESEARCH-PIPELINE.md and compiles initial decision stubs), " +
				"'session' (validates evidence and persists research/sessions/<id>.md or research/FAD.md for synthesis), " +
				"or 'comparison' (validates and persists research/<id>-comparison.md). " +
				"Validates artifacts against the 4-level quality ladder before saving.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleSave,
	)
}

func handleSave(ctx context.Context, req *sdkmcp.CallToolRequest, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start.")
	}

	if err := validateContentSize(in.Content); err != nil {
		return ErrorResult(tool, err, "Reduce content size to under 10 MB.")
	}

	if !core.WorkspaceExists(root) {
		_ = os.MkdirAll(root, 0o755)
		_ = os.MkdirAll(filepath.Join(root, core.ResearchDir), 0o755)
		_ = os.MkdirAll(filepath.Join(root, core.SessionsDir), 0o755)
		_ = os.MkdirAll(filepath.Join(root, core.TemplatesDir), 0o755)

		gitignorePath := filepath.Join(root, core.ResearchDir, ".gitignore")
		if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
			_ = os.WriteFile(gitignorePath, []byte("# Vivechak operational files\n*.lock\n"), 0o644)
		}

		scope := core.ScopeProject
		if in.Kind == "comparison" || strings.EqualFold(in.Scope, "comparison") || (in.ID != "" && strings.HasPrefix(strings.ToUpper(in.ID), "C-")) {
			scope = core.ScopeComparison
		} else if in.Kind == "decision" || strings.EqualFold(in.Scope, "decision") || (in.ID != "" && strings.HasPrefix(strings.ToUpper(in.ID), "D-")) {
			scope = core.ScopeDecision
		}

		for _, tmpl := range core.TemplatesForScope(scope) {
			relPath := filepath.Join(root, core.TemplatesDir, tmpl)
			if data, err := vembed.ReadTemplate(tmpl); err == nil {
				_ = os.WriteFile(relPath, data, 0o644)
			}
		}

		meta := core.WorkspaceMeta{
			Scope:     scope,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if metaBytes, err := json.Marshal(meta); err == nil {
			_ = os.WriteFile(filepath.Join(root, core.MetadataFile), metaBytes, 0o644)
		}
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, err, "Ensure workspace directory exists and is accessible.")
	}
	defer ws.Close()

	kind := strings.ToLower(strings.TrimSpace(in.Kind))
	if kind == "" {
		// Infer kind if possible
		if in.ID != "" && (strings.HasPrefix(in.ID, "C-") || strings.HasPrefix(in.ID, "c-")) {
			kind = "comparison"
		} else if in.ID != "" {
			kind = "session"
		} else {
			kind = "plan"
		}
	}

	switch kind {
	case "plan":
		return handleSavePlanKind(ctx, ws, root, in)
	case "comparison":
		return handleSaveComparisonKind(ctx, ws, root, in)
	case "session":
		return handleSaveSessionKind(ctx, ws, root, in)
	default:
		return ErrorResult(tool, fmt.Errorf("invalid kind %q", in.Kind),
			"Use kind 'plan', 'session', or 'comparison'.")
	}
}

func handleSavePlanKind(_ context.Context, ws *store.Workspace, root string, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	if strings.TrimSpace(in.Content) == "" {
		return ErrorResult(tool, fmt.Errorf("content is required when saving a plan"),
			"Provide the RESEARCH-PIPELINE.md content.")
	}

	planValidation := core.ValidatePlan([]byte(in.Content))
	if planValidation.HasBlocking() {
		var firstErr string
		var firstHint string
		for _, issue := range planValidation.Issues {
			if issue.Level == core.L2Block {
				firstErr = issue.Message
				firstHint = issue.FixHint
				break
			}
		}
		if firstHint == "" {
			firstHint = "Fix dependency cycles, dangling session references, or missing plan content."
		}
		return ErrorResult(tool, fmt.Errorf("plan validation failed: %s", firstErr), firstHint)
	}

	var warnings []string
	for _, issue := range planValidation.Issues {
		warnings = append(warnings, fmt.Sprintf("[%s] %s: %s (fix: %s)", issue.Level, issue.Code, issue.Message, issue.FixHint))
	}

	if strings.TrimSpace(in.DecisionsContent) != "" {
		decValidation := core.ValidateDecisionRegistry([]byte(in.DecisionsContent))
		for _, issue := range decValidation.Issues {
			warnings = append(warnings, issue.String())
		}
	}

	// Parse pipeline DAG
	pipeline, dagErr := core.ParsePipeline([]byte(in.Content))
	if dagErr != nil {
		warnings = append(warnings, fmt.Sprintf("W-DAG-PARSE: %v", dagErr))
	}
	machineBlock, _ := core.ParsePlanMachineBlock([]byte(in.Content))

	_ = ws.MkdirAll(core.ResearchDir, 0o755)

	scope := in.Scope
	if scope == "" {
		info := core.InspectWorkspace(root)
		scope = string(info.Scope)
	}

	pipelinePath := core.PipelineFile
	if strings.EqualFold(scope, "decision") {
		decID := in.ID
		if decID == "" {
			reTableRow := regexp.MustCompile(`(?m)^\|\s*(?:\*\*)?(?:ID|Decision|Decision ID|Comparison ID)(?:\*\*)?\s*\|\s*([A-Za-z0-9_-]+)\s*\|`)
			if m := reTableRow.FindStringSubmatch(in.Content); len(m) > 1 {
				val := strings.TrimSpace(m[1])
				if val != "D-NEW" && val != "C-NEW" && val != "" {
					if strings.Contains(val, "-S") {
						parts := strings.Split(val, "-S")
						decID = parts[0]
					} else {
						decID = val
					}
				}
			}
			if decID == "" {
				reDec := regexp.MustCompile(`\b(D-\d{1,4})\b`)
				if m := reDec.FindStringSubmatch(in.Content); len(m) > 1 && m[1] != "D-NEW" {
					decID = m[1]
				} else {
					decID = "D-01"
				}
			}
		}
		filename := fmt.Sprintf("%s-plan.md", decID)
		if in.Slug != "" {
			filename = fmt.Sprintf("%s-%s-plan.md", decID, in.Slug)
		}
		pipelinePath = filepath.Join(core.ResearchDir, filename)
	}

	if err := store.WriteFileAtomic(ws.Root(), pipelinePath, []byte(in.Content), 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("writing %s: %w", pipelinePath, err), "Check file permissions.")
	}

	// Persist initial DECISIONS.md
	decisionsContent := in.DecisionsContent
	if strings.TrimSpace(decisionsContent) == "" {
		if machineBlock != nil && len(machineBlock.Decisions) > 0 {
			var sb strings.Builder
			sb.WriteString("# Architectural Decisions Registry\n\n")
			for _, d := range machineBlock.Decisions {
				sb.WriteString(fmt.Sprintf("<!-- DECISION: %s -->\n", d.ID))
				sb.WriteString(fmt.Sprintf("---\nid: %s\ntitle: %q\nstatus: proposed\ndoor_type: %s\n---\n\n", d.ID, d.Title, d.Door))
				sb.WriteString(fmt.Sprintf("# %s: %s\n\n[Pending research]\n\n", d.ID, d.Title))
				sb.WriteString(fmt.Sprintf("<!-- /DECISION: %s -->\n\n", d.ID))
			}
			decisionsContent = sb.String()
		} else if !strings.EqualFold(scope, "decision") {
			if tmpl, err := vembed.ReadTemplate("DECISIONS.template.md"); err == nil {
				decisionsContent = string(tmpl)
			}
		}
	}

	if len(decisionsContent) > 0 {
		normalizedDecs := core.NormalizeDecisionRegistry([]byte(decisionsContent))
		_ = store.WriteFileAtomic(ws.Root(), core.DecisionsFile, normalizedDecs, 0o644)
	}

	savedFiles := []string{pipelinePath}
	if !strings.EqualFold(scope, "decision") {
		savedFiles = append(savedFiles, core.DecisionsFile)
	}

	data := map[string]any{
		"saved_files": savedFiles,
		"valid":       !planValidation.HasBlocking(),
		"validation":  planValidation,
	}
	if in.DecisionsContent != "" {
		decVal := core.ValidateDecisionRegistry([]byte(in.DecisionsContent))
		data["decisions_validation"] = decVal
	}
	if pipeline != nil {
		data["sessions_count"] = len(pipeline.Sessions)
	}
	if machineBlock != nil && len(machineBlock.Decisions) > 0 {
		data["decisions_count"] = len(machineBlock.Decisions)
	}

	next := &NextAction{
		Tool: "vivechak_next",
		Why:  "Research plan saved successfully. Call vivechak_next to begin the first research session.",
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  "Research plan saved to research/RESEARCH-PIPELINE.md.",
		Data:     data,
		Warnings: warnings,
		Next:     next,
		NextStep: next.Why,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

func handleSaveComparisonKind(_ context.Context, ws *store.Workspace, root string, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	cid := in.ID
	if cid == "" {
		var allocErr error
		cid, allocErr = store.AllocateID(ws, "C")
		if allocErr != nil {
			cid = "C-001"
		}
	}

	if !isValidID(cid) {
		return ErrorResult(tool, fmt.Errorf("invalid ID %q", cid), "Use a valid identifier like 'C-001'.")
	}

	val := core.ValidateComparison([]byte(in.Content))
	var warnings []string
	for _, issue := range val.Issues {
		warnings = append(warnings, fmt.Sprintf("[%s] %s: %s (fix: %s)", issue.Level, issue.Code, issue.Message, issue.FixHint))
	}

	_ = ws.MkdirAll(core.ResearchDir, 0o755)
	filename := fmt.Sprintf("%s-comparison.md", cid)
	if in.Slug != "" {
		if !strings.HasPrefix(in.Slug, "cmp-") {
			filename = fmt.Sprintf("%s-cmp-%s.md", cid, in.Slug)
		} else {
			filename = fmt.Sprintf("%s-%s.md", cid, in.Slug)
		}
	}
	relPath := filepath.Join(core.ResearchDir, filename)

	if err := store.WriteFileAtomic(ws.Root(), relPath, []byte(in.Content), 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("writing comparison: %w", err), "Check file permissions.")
	}

	data := map[string]any{
		"id":         cid,
		"saved_file": relPath,
		"valid":      val.Status == "valid",
	}

	next := &NextAction{
		Tool: "vivechak_review",
		Args: map[string]any{"mode": "validate", "target": cid},
		Why:  fmt.Sprintf("Comparison WEP matrix successfully saved to %s. Comparison scope is complete. No DAG sessions required. (Optional: run vivechak_review to audit empirical claims).", relPath),
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Comparison saved to %s.", relPath),
		Data:     data,
		Warnings: warnings,
		Next:     next,
		NextStep: next.Why,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

func handleSaveSessionKind(ctx context.Context, ws *store.Workspace, root string, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	sessionID := strings.TrimSpace(in.ID)
	if sessionID == "" {
		return ErrorResult(tool, fmt.Errorf("id (session_id) is required when saving a session"),
			"Provide the session identifier (e.g. 'T1-01', 'SYN-01').")
	}

	if !isValidID(sessionID) {
		return ErrorResult(tool, fmt.Errorf("invalid session id %q", sessionID), "Use an ID like 'T1-01'.")
	}

	isSynthesis := core.IsSynthesisSession(sessionID)
	isOneWay := false
	var dag *core.DAG
	if pipeData, pErr := ws.ReadFile(core.PipelineFile); pErr == nil {
		dag, _ = core.ParsePipeline(pipeData)
		if dag != nil {
			if s := dag.SessionByID(sessionID); s != nil {
				isOneWay = strings.EqualFold(s.DoorType, "one-way")
			}
		}
	}

	var val *core.ValidationResult
	if isSynthesis {
		val = core.ValidateFAD([]byte(in.Content))
	} else {
		val = core.ValidateSessionWithContext([]byte(in.Content), isOneWay)
	}

	var warnings []string
	for _, issue := range val.Issues {
		warnings = append(warnings, fmt.Sprintf("[%s] %s: %s (fix: %s)", issue.Level, issue.Code, issue.Message, issue.FixHint))
	}

	// Quality coaching observations (non-blocking Q-* codes)
	qualityObs := core.ObserveSessionQuality([]byte(in.Content))
	for _, q := range qualityObs {
		if q.Severity == "consideration" || q.Severity == "suggestion" {
			warnings = append(warnings, fmt.Sprintf("Q-%s: %s", strings.ToUpper(q.Category), q.Message))
		}
	}

	contentBytes := []byte(in.Content)
	if val.HasBlocking() {
		contentBytes = core.SetFrontmatterField(contentBytes, "status", "draft")
	} else {
		contentBytes = core.SetFrontmatterField(contentBytes, "status", "complete")
	}

	var relPath string
	if isSynthesis {
		relPath = core.FADFile
	} else {
		filename := fmt.Sprintf("%s.md", sessionID)
		if dag != nil {
			if s := dag.SessionByID(sessionID); s != nil && s.OutputFile != "" {
				filename = filepath.Base(s.OutputFile)
			}
		}
		relPath = filepath.Join(core.SessionsDir, filename)
	}
	relPath = filepath.ToSlash(relPath)

	_ = ws.MkdirAll(filepath.Dir(relPath), 0o755)

	unlock, lockErr := store.LockFile(ctx, filepath.Join(root, relPath), 5*time.Second)
	if lockErr != nil {
		return ErrorResult(tool, fmt.Errorf("could not acquire lock for %s: %w", relPath, lockErr), "Try again.")
	}
	defer func() { _ = unlock() }()

	if err := store.WriteFileAtomic(ws.Root(), relPath, contentBytes, 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("writing session %s: %w", sessionID, err), "Check permissions.")
	}

	var obsData []map[string]any
	for _, o := range qualityObs {
		obsData = append(obsData, map[string]any{
			"category": o.Category,
			"message":  o.Message,
			"severity": o.Severity,
		})
	}

	var knownIDs []string
	if dag != nil {
		for _, s := range dag.Sessions {
			knownIDs = append(knownIDs, s.ID)
		}
	}
	completedSessions, _ := scanCompletedSessions(ws, knownIDs...)
	gateDecisions := collectGateDecisions(ws)
	progress := map[string]any{
		"sessions_completed": len(completedSessions),
		"decisions_recorded": len(gateDecisions),
	}
	if dag != nil {
		progress["sessions_total"] = len(dag.Sessions)
	}

	data := map[string]any{
		"session_id":           sessionID,
		"saved_file":           relPath,
		"file_path":            relPath,
		"status":               val.Status,
		"valid":                !val.HasBlocking(),
		"validation_passed":    !val.HasBlocking(),
		"quality_observations": obsData,
		"progress":             progress,
	}

	nextStep := buildNextStep(sessionID, val.ErrorCount(), qualityObs, dag)
	var next *NextAction
	if isSynthesis {
		next = &NextAction{
			Tool: "vivechak_gate",
			Why:  nextStep,
		}
	} else {
		next = &NextAction{
			Tool: "vivechak_next",
			Why:  nextStep,
		}
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Session %s saved to %s.", sessionID, relPath),
		Data:     data,
		Warnings: warnings,
		Next:     next,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// handleSavePlan provides backward compatibility for unit tests calling handleSavePlan.
func handleSavePlan(ctx context.Context, req *sdkmcp.CallToolRequest, in SavePlanInput) (*sdkmcp.CallToolResult, Envelope, error) {
	content := in.Content
	if content == "" {
		content = in.PipelineContent
	}
	kind := "plan"
	id := in.ID
	if in.DecisionID != "" {
		id = in.DecisionID
	}
	if in.ComparisonID != "" {
		id = in.ComparisonID
		kind = "comparison"
	}
	if in.Scope == "comparison" {
		kind = "comparison"
	}

	return handleSave(ctx, req, SaveInput{
		ProjectRoot:      in.ProjectRoot,
		Kind:             kind,
		Scope:            in.Scope,
		ID:               id,
		Content:          content,
		DecisionsContent: in.DecisionsContent,
		Slug:             in.Slug,
	})
}

// handleSaveSession provides backward compatibility for unit tests calling handleSaveSession.
func handleSaveSession(ctx context.Context, req *sdkmcp.CallToolRequest, in SaveSessionInput) (*sdkmcp.CallToolResult, Envelope, error) {
	return handleSave(ctx, req, SaveInput{
		ProjectRoot: in.ProjectRoot,
		Kind:        "session",
		ID:          in.SessionID,
		Content:     in.Content,
	})
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

	// Dedicated synthesis exit gate routing
	if core.IsSynthesisSession(sessionID) {
		parts = append(parts, "🎉 FAD synthesis saved to research/FAD.md. "+
			"All research pipeline sessions are complete, but Phase 0 is NOT yet passed.\n\n"+
			"➡️ NEXT REQUIRED STEP: Run 'vivechak_run_gate' to evaluate Phase 0 exit criteria "+
			"before code scaffolding or implementation.")
		return strings.Join(parts, "\n\n")
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
			"🛑 ONE-WAY DOOR: Session '%s' informs an irreversible decision.\n"+
				"1. Run 'vivechak_challenge(session_id=\"%s\", mode=\"red_team\")' to stress-test findings.\n"+
				"2. Amend findings via 'vivechak_amend_session'.\n"+
				"3. Auto-draft and record canonical ADR via 'vivechak_record_decision(decision_id=\"<D-ID>\", auto_draft_from=\"%s\")'.",
			sessionID, sessionID, sessionID))
	} else {
		parts = append(parts, fmt.Sprintf(
			"Next: Run 'vivechak_next_session' for the next session, or "+
				"record an ADR via 'vivechak_record_decision(decision_id=\"<D-ID>\", auto_draft_from=\"%s\")'.",
			sessionID))
	}

	return strings.Join(parts, "\n\n")
}
