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
	versionPkg "github.com/bhaskarjha-dev/vivechak/internal/version"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SaveInput holds the arguments for vivechak_save.
type SaveInput struct {
	ProjectRoot      string `json:"project_root,omitempty"       jsonschema:"workspace root path (optional)"`
	Kind             string `json:"kind"                         jsonschema:"required,artifact kind: 'plan' (RESEARCH-PIPELINE.md) | 'session' (research session) | 'comparison' (WEP matrix) | 'fad' (FAD.md)"`
	Scope            string `json:"scope,omitempty"              jsonschema:"optional scope hint ('comparison' | 'decision' | 'project')"`
	ID               string `json:"id,omitempty"                 jsonschema:"identifier: session_id (e.g. T1-01, SYN-01) or comparison_id (e.g. C-001)"`
	SessionID        string `json:"session_id,omitempty"         jsonschema:"alias for id"`
	DecisionID       string `json:"decision_id,omitempty"        jsonschema:"alias for id in decision scope"`
	Content          string `json:"content"                      jsonschema:"required,markdown artifact content to save"`
	PipelineContent  string `json:"pipeline_content,omitempty"   jsonschema:"alias for content when saving a research pipeline"`
	DecisionsContent string `json:"decisions_content,omitempty" jsonschema:"initial decision registry content when saving plan (optional)"`
	Slug             string `json:"slug,omitempty"               jsonschema:"URL-safe slug for the artifact (optional)"`
	Revise           bool   `json:"revise,omitempty"             jsonschema:"force revision of existing artifact with history archiving"`
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
	Revise      bool   `json:"revise,omitempty"`
}

var (
	reSaveTableRow  = regexp.MustCompile(`(?m)^\|\s*(?:\*\*)?(?:ID|Decision|Decision ID|Comparison ID)(?:\*\*)?\s*\|\s*(D-\d+|C-\d+|[A-Za-z]+-\d+)\s*\|`)
	reSaveHeadingID = regexp.MustCompile(`(?m)^#+\s*(?:Session\s+)?([A-Za-z]+-\d+[A-Za-z0-9_-]*)`)
	reSaveDec       = regexp.MustCompile(`\b(D-\d{1,4})\b`)
	reSaveDNew      = regexp.MustCompile(`\bD-NEW\b`)
	reSaveDNewDash  = regexp.MustCompile(`\bD-NEW-`)
)

func registerSave(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_save",
			Title: "Save Research Artifact",
			Description: "Validate and persist research artifacts into the workspace. " +
				"Pass 'kind': 'plan' (persists RESEARCH-PIPELINE.md and compiles initial decision stubs), " +
				"'session' (validates evidence and persists research/sessions/<id>.md), " +
				"'comparison' (validates and persists research/<id>-comparison.md), " +
				"or 'fad' (validates and persists research/FAD.md for synthesis). " +
				"Validates artifacts against the 4-level quality ladder before saving.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: BoolPtr(true),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleSave,
	)
}

func handleSave(ctx context.Context, req *sdkmcp.CallToolRequest, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	var roots []string
	if strings.TrimSpace(in.ProjectRoot) == "" {
		roots = getClientRoots(ctx, req)
	}
	root, rule, err := core.ResolveWorkspaceWithRoots(ctx, in.ProjectRoot, roots)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_start or pass project_root.")
	}

	if err := validateContentSize(in.Content); err != nil {
		return ErrorResult(tool, err, "Reduce content size to under 10 MB.")
	}

	if in.Content != "" {
		if err := core.CheckHumanOnlyField([]byte(in.Content)); err != nil {
			return ErrorResult(tool, err, "Ask the human architect to run 'vck signoff <decision-id>' in an interactive terminal.")
		}
	}
	if in.DecisionsContent != "" {
		if err := core.CheckHumanOnlyField([]byte(in.DecisionsContent)); err != nil {
			return ErrorResult(tool, err, "Ask the human architect to run 'vck signoff <decision-id>' in an interactive terminal.")
		}
	}

	if in.Slug != "" {
		if err := store.ValidateSlug(in.Slug); err != nil {
			return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: %w", err), "Provide a URL-safe alphanumeric slug without path traversal or reserved names.")
		}
	}

	_ = os.MkdirAll(root, 0o755)
	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, err, "Ensure workspace directory exists and is accessible.")
	}
	defer ws.Close()

	if !core.WorkspaceExists(root) {
		_ = ws.MkdirAll(core.ResearchDir, 0o755)
		_ = ws.MkdirAll(core.SessionsDir, 0o755)

		gitignoreRel := filepath.Join(core.ResearchDir, ".gitignore")
		if !ws.FileExists(gitignoreRel) {
			_ = store.WriteFileAtomic(ws.Root(), gitignoreRel, []byte("# Vivechak operational files\n*.lock\n.locks/\n.history/\n"), 0o644)
		}

		if !ws.FileExists(core.MetadataFile) {
			meta := map[string]any{
				"schema":     1,
				"created_by": fmt.Sprintf("vivechak %s", versionPkg.Version),
				"counters":   map[string]int{"C": 0, "D": 0},
				"created_at": time.Now().UTC().Format(time.RFC3339),
			}
			if metaBytes, err := json.MarshalIndent(meta, "", "  "); err == nil {
				_ = store.WriteFileAtomic(ws.Root(), core.MetadataFile, metaBytes, 0o644)
			}
		}
	}

	if in.ID == "" {
		if in.SessionID != "" {
			in.ID = in.SessionID
		} else if in.DecisionID != "" {
			in.ID = in.DecisionID
		}
	}

	if in.Content == "" && in.PipelineContent != "" {
		in.Content = in.PipelineContent
	}
	if in.Content == "" {
		return ErrorResult(tool, fmt.Errorf("content is required"), "Provide content to save.")
	}

	kind := strings.ToLower(strings.TrimSpace(in.Kind))
	if kind == "" {
		if in.SessionID != "" {
			kind = "session"
		} else if in.PipelineContent != "" {
			kind = "plan"
		} else if in.Scope == "comparison" || strings.HasPrefix(strings.ToUpper(in.ID), "C-") {
			kind = "comparison"
		}
	}
	if kind == "" {
		return ErrorResult(tool, fmt.Errorf("kind is required: use 'plan', 'session', 'comparison', or 'fad'"),
			"Provide the 'kind' parameter indicating what artifact type is being saved.")
	}

	switch kind {
	case "plan":
		return handleSavePlanKind(ctx, ws, root, rule, in)
	case "comparison":
		return handleSaveComparisonKind(ctx, ws, root, rule, in)
	case "session", "fad":
		if kind == "fad" && strings.TrimSpace(in.ID) == "" {
			in.ID = "SYN-01"
		}
		return handleSaveSessionKind(ctx, ws, root, rule, in)
	default:
		return ErrorResult(tool, fmt.Errorf("invalid kind %q", in.Kind),
			"Use kind 'plan', 'session', 'comparison', or 'fad'.")
	}
}

func ensureWorkspaceTemplates(ws *store.Workspace, scope core.Scope) {
	if scope == "" {
		scope = core.ScopeProject
	}
	_ = ws.MkdirAll(core.TemplatesDir, 0o755)
	for _, tmpl := range core.TemplatesForScope(scope) {
		relPath := filepath.Join(core.TemplatesDir, tmpl)
		if !ws.FileExists(relPath) {
			if tmplBytes, err := vembed.ReadTemplate(tmpl); err == nil {
				_ = store.WriteFileAtomic(ws.Root(), relPath, tmplBytes, 0o644)
			}
		}
	}
}

func handleSavePlanKind(ctx context.Context, ws *store.Workspace, root, rule string, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	if strings.TrimSpace(in.Content) == "" {
		return ErrorResult(tool, fmt.Errorf("content is required when saving a plan"),
			"Provide the RESEARCH-PIPELINE.md content.")
	}

	info := core.InspectWorkspace(root)
	scope := strings.ToLower(strings.TrimSpace(in.Scope))
	if scope == "" {
		if info.Scope == core.ScopeComparison || (in.ID != "" && strings.HasPrefix(strings.ToUpper(in.ID), "C-")) {
			scope = "comparison"
		} else if in.DecisionID != "" || (in.ID != "" && strings.HasPrefix(strings.ToUpper(in.ID), "D-")) {
			scope = "decision"
		} else {
			scope = "project"
		}
	}
	if scope == "comparison" {
		return handleSaveComparisonKind(ctx, ws, root, rule, in)
	}
	if scope != "project" && scope != "decision" && scope != "comparison" {
		return ErrorResult(tool, fmt.Errorf("invalid plan scope %q: use 'project', 'decision', or 'comparison'", in.Scope),
			"Provide 'scope': 'project', 'decision', or 'comparison'.")
	}

	ensureWorkspaceTemplates(ws, core.Scope(scope))

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

	_ = ws.MkdirAll(core.ResearchDir, 0o755)

	pipelinePath := core.PipelineFile
	planID := ""
	switch scope {
	case "comparison":
		compID := in.ID
		if compID == "" {
			compID = in.DecisionID
		}
		if compID == "" {
			compID = "C-001"
		}
		filename := fmt.Sprintf("%s-comparison.md", compID)
		if in.Slug != "" {
			filename = fmt.Sprintf("%s-cmp-%s.md", compID, in.Slug)
		}
		pipelinePath = filepath.Join(core.ResearchDir, filename)
		planID = compID
	case "decision":
		decID := in.ID
		if decID == "" {
			decID = in.DecisionID
		}
		if decID == "" {
			if m := reSaveTableRow.FindStringSubmatch(in.Content); len(m) > 1 {
				val := strings.TrimSpace(m[1])
				lower := strings.ToLower(val)
				if val != "D-NEW" && val != "C-NEW" && val != "" &&
					lower != "role" && lower != "depends" && lower != "filename" &&
					lower != "topic" && lower != "description" && lower != "status" && lower != "title" {
					decID = val
				}
			}
			if decID == "" {
				if m := reSaveHeadingID.FindStringSubmatch(in.Content); len(m) > 1 {
					val := strings.TrimSpace(m[1])
					if strings.Contains(val, "-S") {
						parts := strings.Split(val, "-S")
						decID = parts[0]
					} else {
						decID = val
					}
				}
			}
			if decID == "" {
				if m := reSaveDec.FindStringSubmatch(in.Content); len(m) > 1 && m[1] != "D-NEW" {
					decID = m[1]
				}
			}
		}
		isAllocated := false
		if decID == "" || strings.EqualFold(decID, "D-NEW") || strings.EqualFold(decID, "NEW") {
			var allocErr error
			decID, allocErr = store.AllocateID(ws, "D")
			if allocErr != nil {
				return ErrorResult(tool, fmt.Errorf("allocating decision ID: %w", allocErr), "Ensure research/.vivechak.json is accessible.")
			}
			isAllocated = true
		} else {
			decID = core.NormalizeDecisionID(decID)
		}

		if !isValidID(decID) {
			return ErrorResult(tool, fmt.Errorf("E-ID-INVALID: invalid plan ID %q", decID), "Provide a valid alphanumeric decision ID.")
		}
		planID = decID
		if isAllocated {
			warnings = append(warnings, fmt.Sprintf("W-ID-ALLOCATED: allocated decision ID %s", planID))
			in.Content = reSaveDNew.ReplaceAllString(in.Content, planID)
			in.Content = reSaveDNewDash.ReplaceAllString(in.Content, planID+"-")
		}

		filename := fmt.Sprintf("%s-plan.md", planID)
		if in.Slug != "" {
			filename = fmt.Sprintf("%s-%s-plan.md", planID, in.Slug)
		}
		pipelinePath = filepath.Join(core.ResearchDir, filename)
	default:
		planID = "DAG"
	}

	contentBytes := []byte(in.Content)
	fm, body, _ := core.ParseFrontmatter(contentBytes)
	if fm == nil {
		fm = make(core.Frontmatter)
	}
	fm["id"] = planID
	fm["type"] = "plan"
	fm["vivechak"] = 1
	if !fm.Has("title") || fm.GetString("title") == "" {
		fm["title"] = "Research Plan"
	}
	if !fm.Has("date") {
		fm["date"] = time.Now().UTC().Format("2006-01-02")
	}
	if !fm.Has("status") {
		if planValidation.HasBlocking() {
			fm["status"] = "draft"
		} else {
			fm["status"] = "complete"
		}
	}
	if composed, err := core.ComposeFrontmatter(fm, body); err == nil {
		contentBytes = composed
	}

	if err := ws.WriteArtifact(pipelinePath, contentBytes, 0o644, in.Revise); err != nil {
		return ErrorResult(tool, err, "Pass revise=true if updating an existing plan.")
	}

	savedFiles := []string{filepath.ToSlash(pipelinePath)}
	var decValResult any
	if strings.TrimSpace(in.DecisionsContent) != "" {
		decVal := core.ValidateDecisionRegistry([]byte(in.DecisionsContent))
		decValResult = decVal
		for _, issue := range decVal.Issues {
			warnings = append(warnings, fmt.Sprintf("%s[%s] %s: %s (fix: %s)", issue.Prefix, issue.Level, issue.Code, issue.Message, issue.FixHint))
		}
		if scope == "decision" {
			filename := fmt.Sprintf("%s.md", planID)
			if in.Slug != "" {
				filename = fmt.Sprintf("%s-%s.md", planID, in.Slug)
			}
			adrPath := filepath.Join(core.ResearchDir, filename)
			adrBytes := []byte(in.DecisionsContent)
			adrFM, adrBody, _ := core.ParseFrontmatter(adrBytes)
			if adrFM == nil {
				adrFM = make(core.Frontmatter)
			}
			adrFM["id"] = planID
			adrFM["type"] = "decision"
			adrFM["vivechak"] = 1
			if !adrFM.Has("status") {
				adrFM["status"] = "proposed"
			}
			if !adrFM.Has("date") {
				adrFM["date"] = time.Now().UTC().Format("2006-01-02")
			}
			if composed, err := core.ComposeFrontmatter(adrFM, adrBody); err == nil {
				adrBytes = composed
			}
			if err := ws.WriteArtifact(adrPath, adrBytes, 0o644, in.Revise); err == nil {
				savedFiles = append(savedFiles, filepath.ToSlash(adrPath))
			}
		} else {
			decBytes := core.NormalizeDecisionRegistry([]byte(in.DecisionsContent))
			decPath := core.DecisionsFile
			if err := ws.WriteArtifact(decPath, decBytes, 0o644, in.Revise); err == nil {
				savedFiles = append(savedFiles, filepath.ToSlash(decPath))
			}
		}
	}

	data := map[string]any{
		"workspace": map[string]any{
			"root": root,
			"rule": rule,
		},
		"id":          planID,
		"file_path":   filepath.ToSlash(pipelinePath),
		"saved_files": savedFiles,
		"valid":       !planValidation.HasBlocking(),
		"validation":  planValidation,
	}
	if decValResult != nil {
		data["decisions_validation"] = decValResult
	}

	msg := fmt.Sprintf("Research plan saved to %s.", filepath.ToSlash(pipelinePath))
	next := &NextAction{
		Tool: "vivechak_next",
		Why:  fmt.Sprintf("Research plan saved to %s. Call vivechak_next to begin the first research session.", filepath.ToSlash(pipelinePath)),
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  msg,
		Data:     data,
		Warnings: warnings,
		Next:     next,
		NextStep: next.Why,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

func handleSaveComparisonKind(_ context.Context, ws *store.Workspace, root, rule string, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	ensureWorkspaceTemplates(ws, core.ScopeComparison)

	cid := in.ID
	isAllocated := false
	if cid == "" || strings.EqualFold(cid, "C-NEW") || strings.EqualFold(cid, "NEW") {
		var allocErr error
		cid, allocErr = store.AllocateID(ws, "C")
		if allocErr != nil {
			return ErrorResult(tool, fmt.Errorf("allocating comparison ID: %w", allocErr), "Ensure research/.vivechak.json is accessible.")
		}
		isAllocated = true
	} else {
		cid = core.NormalizeComparisonID(cid)
	}

	if !isValidID(cid) {
		return ErrorResult(tool, fmt.Errorf("E-ID-INVALID: invalid ID %q", cid), "Use a valid identifier like 'C-001'.")
	}

	var warnings []string
	if isAllocated {
		warnings = append(warnings, fmt.Sprintf("W-ID-ALLOCATED: allocated comparison ID %s", cid))
	}

	val := core.ValidateComparison([]byte(in.Content))
	for _, issue := range val.Issues {
		warnings = append(warnings, fmt.Sprintf("[%s] %s: %s (fix: %s)", issue.Level, issue.Code, issue.Message, issue.FixHint))
	}

	contentBytes := []byte(in.Content)
	fm, body, _ := core.ParseFrontmatter(contentBytes)
	if fm == nil {
		fm = make(core.Frontmatter)
	}
	fm["id"] = cid
	fm["type"] = "comparison"
	fm["vivechak"] = 1
	if !fm.Has("title") || fm.GetString("title") == "" {
		fm["title"] = "Comparison"
	}
	if !fm.Has("date") {
		fm["date"] = time.Now().UTC().Format("2006-01-02")
	}
	if val.HasBlocking() {
		fm["status"] = "draft"
	} else {
		fm["status"] = "complete"
	}
	if composed, err := core.ComposeFrontmatter(fm, body); err == nil {
		contentBytes = composed
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

	if err := ws.WriteArtifact(relPath, contentBytes, 0o644, in.Revise); err != nil {
		return ErrorResult(tool, err, "Pass revise=true if modifying an existing comparison.")
	}

	data := map[string]any{
		"workspace": map[string]any{
			"root": root,
			"rule": rule,
		},
		"id":         cid,
		"file_path":  filepath.ToSlash(relPath),
		"saved_file": filepath.ToSlash(relPath),
		"valid":      val.Status == "valid",
		"validation": val,
	}

	next := &NextAction{
		Tool: "vivechak_review",
		Args: map[string]any{"mode": "validate", "target": cid},
		Why:  fmt.Sprintf("Comparison WEP matrix successfully saved to %s. Comparison scope is complete. No DAG sessions required. (Optional: run vivechak_review to audit empirical claims).", filepath.ToSlash(relPath)),
	}

	env := Envelope{
		Ok:       true,
		Success:  true,
		Message:  fmt.Sprintf("Comparison saved to %s.", filepath.ToSlash(relPath)),
		Data:     data,
		Warnings: warnings,
		Next:     next,
		NextStep: next.Why,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

func handleSaveSessionKind(ctx context.Context, ws *store.Workspace, root, rule string, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	ensureWorkspaceTemplates(ws, core.ScopeProject)

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
	fm, sBody, _ := core.ParseFrontmatter(contentBytes)
	if fm == nil {
		fm = make(core.Frontmatter)
	}
	fm["id"] = sessionID
	if isSynthesis {
		fm["type"] = "fad"
	} else {
		fm["type"] = "session"
	}
	fm["vivechak"] = 1
	if !fm.Has("title") || fm.GetString("title") == "" {
		fm["title"] = sessionID
	}
	if !fm.Has("date") {
		fm["date"] = time.Now().UTC().Format("2006-01-02")
	}
	if val.HasBlocking() {
		fm["status"] = "draft"
	} else {
		fm["status"] = "complete"
	}
	if comp, err := core.ComposeFrontmatter(fm, sBody); err == nil {
		contentBytes = comp
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

	if err := ws.WriteArtifact(relPath, contentBytes, 0o644, in.Revise); err != nil {
		return ErrorResult(tool, err, "Pass revise=true if revising an existing session.")
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
		"workspace": map[string]any{
			"root": root,
			"rule": rule,
		},
		"session_id":           sessionID,
		"id":                   sessionID,
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
		Revise:      in.Revise,
	})
}

func buildNextStep(sessionID string, errorCount int, observations []core.QualityObservation, dag *core.DAG) string {
	if errorCount > 0 {
		return fmt.Sprintf("Session saved as draft (%d blocking issues). Fix the issues and re-save, "+
			"or run vivechak_review for quality coaching.", errorCount)
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
			"➡️ NEXT REQUIRED STEP: Run 'vivechak_gate' to evaluate Phase 0 exit criteria "+
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
				"1. Run 'vivechak_review(session_id=\"%s\", mode=\"challenge\")' to stress-test findings.\n"+
				"2. Amend findings via 'vivechak_amend'.\n"+
				"3. Auto-draft and record canonical ADR via 'vivechak_record_decision(decision_id=\"<D-ID>\", auto_draft_from=\"%s\")'.",
			sessionID, sessionID, sessionID))
	} else {
		parts = append(parts, fmt.Sprintf(
			"Next: Run 'vivechak_next' for the next session, or "+
				"record an ADR via 'vivechak_record_decision(decision_id=\"<D-ID>\", auto_draft_from=\"%s\")'.",
			sessionID))
	}

	return strings.Join(parts, "\n\n")
}
