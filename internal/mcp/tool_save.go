package mcputil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	store "github.com/bhaskarjha-dev/vivechak/internal/store"
	versionPkg "github.com/bhaskarjha-dev/vivechak/internal/version"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SaveInput holds the canonical arguments for vivechak_save.
type SaveInput struct {
	ProjectRoot      string `json:"project_root,omitempty"`
	Kind             string `json:"kind"`
	Scope            string `json:"scope,omitempty"`
	ID               string `json:"id,omitempty"`
	Content          string `json:"content"`
	DecisionsContent string `json:"decisions_content,omitempty"`
	Slug             string `json:"slug,omitempty"`
	Revise           bool   `json:"revise,omitempty"`
}

var (
	reDecisionPlanHeading = regexp.MustCompile(`(?im)^#{1,3}\s+D-[A-Za-z0-9]+-PLAN\b`)
	reDecisionSessionID   = regexp.MustCompile(`(?i)^D-[A-Za-z0-9]+-S\d+$`)
)

func deriveDecisionIDFromPlan(content string) string {
	if block, err := core.ParsePlanMachineBlock([]byte(content)); err == nil && block != nil {
		id := strings.TrimSpace(block.Plan)
		if m := numberedIDRe.FindStringSubmatch(id); m != nil && strings.EqualFold(m[1], "D") {
			return id
		}
	}
	return ""
}

func planScopeFromContent(content string) string {
	if block, err := core.ParsePlanMachineBlock([]byte(content)); err == nil && block != nil {
		switch p := strings.ToUpper(strings.TrimSpace(block.Plan)); {
		case strings.HasPrefix(p, "D-"):
			return "decision"
		case p == "PROJECT" || p == "DAG":
			return "project"
		}
		if len(block.Sessions) > 0 {
			for _, s := range block.Sessions {
				if !reDecisionSessionID.MatchString(strings.TrimSpace(s.ID)) {
					return "project"
				}
			}
			return "decision"
		}
	}
	if reDecisionPlanHeading.MatchString(content) {
		return "decision"
	}
	return ""
}

func registerSave(server *sdkmcp.Server) {
	registerTool(server, ToolSpec{
		Name:        "vivechak_save",
		Description: "Validate and persist research under research/ (comparison, plan, session, or fad). Needs revise=true to overwrite; invalid saves as draft.",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint:    false,
			IdempotentHint:  false,
			DestructiveHint: BoolPtr(true),
		},
		Params: []ParamSpec{
			{Name: "kind", Type: "string", Enum: []string{"plan", "session", "comparison", "fad"}, Required: true, Doc: "Artifact kind: plan, session, comparison, or fad."},
			{Name: "content", Type: "string", Required: true, Description: "Markdown body", Doc: "Markdown content to validate and save."},
			{Name: "id", Type: "string", Description: "Artifact ID", Doc: "Artifact or session ID (allocated when omitted for comparison/plan)."},
			{Name: "scope", Type: "string", Enum: []string{"comparison", "decision", "project"}, Doc: "Optional scope hint when saving a plan."},
			{Name: "slug", Type: "string", Doc: "Optional URL-safe filename slug."},
			{Name: "revise", Type: "boolean", Doc: "Update an existing artifact, archiving the old version."},
			{Name: "decisions_content", Type: "string", Description: "Seeded ADR", Doc: "Initial ADR stub or registry Markdown when saving a plan."},
			{Name: "project_root", Type: "string", Doc: "Absolute path to the target project directory."},
		},
	}, handleSave)
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
		return ErrorResult(tool, fmt.Errorf("E-IO: %w", err), "Ensure workspace directory exists and is accessible.")
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

	if in.Content == "" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: content is required"), "Provide content to save.")
	}

	kind := strings.ToLower(strings.TrimSpace(in.Kind))
	if kind == "" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: kind is required: use 'plan', 'session', 'comparison', or 'fad'"),
			"Provide the 'kind' parameter indicating what artifact type is being saved.")
	}

	switch kind {
	case "plan":
		return handleSavePlanKind(req, ws, root, rule, in)
	case "comparison":
		return handleSaveComparisonKind(req, ws, root, rule, in)
	case "session", "fad":
		if kind == "fad" && strings.TrimSpace(in.ID) == "" {
			in.ID = "SYN-01"
			if synth := readPipelineDAG(ws).SynthesisSessions(); len(synth) == 1 {
				in.ID = synth[0].ID
			}
		}
		sessionID, mapped, idErr := resolvePlaceholderRef(sessionOf(req), ws, root, in.ID, false)
		if idErr != nil {
			return ErrorResult(tool, idErr, placeholderRefFix)
		}
		var idWarnings []Warning
		if mapped {
			did := sessionID[:strings.Index(strings.ToUpper(sessionID), "-S")]
			in.Content = placeholderTokenRe.ReplaceAllString(in.Content, did)
			idWarnings = append(idWarnings, NewWarning("W-ID-ALLOCATED", fmt.Sprintf("session id %s resolved to %s", strings.TrimSpace(in.ID), sessionID), ""))
			in.ID = sessionID
		}
		return handleSaveSessionKind(ctx, ws, root, rule, in, idWarnings...)
	default:
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: invalid kind %q", in.Kind),
			"Use kind 'plan', 'session', 'comparison', or 'fad'.")
	}
}

func handleSavePlanKind(req *sdkmcp.CallToolRequest, ws *store.Workspace, root, rule string, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	if strings.TrimSpace(in.Content) == "" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: content is required when saving a plan"),
			"Provide the plan Markdown content.")
	}

	scope := strings.ToLower(strings.TrimSpace(in.Scope))
	planIDArg := strings.ToUpper(strings.TrimSpace(in.ID))
	if strings.HasPrefix(planIDArg, "C-") {
		return ErrorResult(tool, fmt.Errorf("E-ID-INVALID: %q is a comparison id; kind=\"plan\" saves a decision plan (D-NNN) or the project pipeline (DAG)", in.ID),
			"Save comparisons with vivechak_save(kind=\"comparison\", id=\"C-NNN\", content=...).")
	}
	if scope == "comparison" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: kind=\"plan\" saves a decision plan or the project pipeline, not a comparison"),
			"Save comparisons with vivechak_save(kind=\"comparison\", content=...).")
	}
	if scope == "" {
		switch {
		case planIDArg == "DAG" || planIDArg == "PROJECT":
			scope = "project"
		case strings.HasPrefix(planIDArg, "D-"):
			scope = "decision"
		default:
			scope = planScopeFromContent(in.Content)
			if scope == "" {
				if info := core.InspectWorkspace(root); info.Scope == core.ScopeDecision {
					scope = "decision"
				} else {
					scope = "project"
				}
			}
		}
	}
	if scope != "project" && scope != "decision" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: invalid plan scope %q: use 'project' or 'decision'", in.Scope),
			"Provide scope 'project' or 'decision' (comparisons are saved with kind=\"comparison\").")
	}

	if scope == "decision" {
		raw := strings.TrimSpace(in.ID)
		if err := checkArtifactIDSyntax("D", raw); err != nil {
			return ErrorResult(tool, err, "Use a decision id like D-001, or omit id (or pass D-NEW) to get the next free one.")
		}
	}

	planValidation := core.ValidatePlanForScope([]byte(in.Content), scope)
	if planValidation.HasBlocking() {
		var problems, fixes []string
		for _, issue := range planValidation.BlockingIssues() {
			problems = append(problems, issue.Code+": "+issue.Message)
			if issue.FixHint != "" && !slices.Contains(fixes, issue.FixHint) {
				fixes = append(fixes, issue.FixHint)
			}
		}
		return ErrorResult(tool, fmt.Errorf("plan validation failed: %s", strings.Join(problems, "; ")), strings.Join(fixes, " "))
	}

	var warnings []Warning
	for _, issue := range planValidation.Issues {
		warnings = append(warnings, WarningFromIssue(issue))
	}

	_ = ws.MkdirAll(core.ResearchDir, 0o755)

	pipelinePath := core.PipelineFile
	planID := ""
	switch scope {
	case "decision":
		decID := strings.TrimSpace(in.ID)
		if decID == "" {
			decID = deriveDecisionIDFromPlan(in.Content)
		}
		resolvedID, isAllocated, idErr := resolveArtifactID(sessionOf(req), ws, root, "D", decID,
			func(id string) bool { return !decisionPlanExists(ws, id) })
		if idErr != nil {
			return ErrorResult(tool, idErr, "Use a decision id like D-001, or omit id (or pass D-NEW) to get the next free one.")
		}
		planID = resolvedID
		if isAllocated {
			warnings = append(warnings, NewWarning("W-ID-ALLOCATED", fmt.Sprintf("allocated decision ID %s", planID), ""))
		}
		in.Content = placeholderTokenRe.ReplaceAllString(in.Content, planID)
		in.DecisionsContent = placeholderTokenRe.ReplaceAllString(in.DecisionsContent, planID)

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
	if !fm.Has("title") || strings.TrimSpace(fm.GetString("title")) == "" {
		if t := core.PlanDecisionTitle(contentBytes); t != "" {
			fm["title"] = t
		} else {
			fm["title"] = "[Decision title]"
		}
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
		return ErrorResult(tool, err, "Pass revise=true only if this updates that same plan; for a different decision, save without an id (or with id D-NEW) to get a new one.")
	}
	if scope == "decision" {
		if st := stateForSession(sessionOf(req)); st != nil {
			st.notePlanDecision(root, planID)
		}
	}

	savedFiles := []string{filepath.ToSlash(pipelinePath)}
	var decValResult any
	if strings.TrimSpace(in.DecisionsContent) != "" {
		decVal := core.ValidateDecisionRegistry([]byte(in.DecisionsContent))
		decValResult = decVal
		for _, issue := range decVal.Issues {
			warnings = append(warnings, WarningFromIssue(issue))
		}
		if scope == "decision" {
			filename := resolveDecisionFilename(ws, planID, in.Slug)
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
		Args: map[string]any{},
		Why:  fmt.Sprintf("Research plan saved to %s. Call vivechak_next to begin the first research session.", filepath.ToSlash(pipelinePath)),
	}

	env := Envelope{
		Ok:       true,
		Message:  msg,
		Data:     data,
		Warnings: warnings,
		Next:     next,
	}
	return env.ToResult()
}

func handleSaveComparisonKind(req *sdkmcp.CallToolRequest, ws *store.Workspace, root, rule string, in SaveInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	cid, isAllocated, idErr := resolveArtifactID(sessionOf(req), ws, root, "C", in.ID,
		func(id string) bool { return !artifactIDInUse(ws, id) })
	if idErr != nil {
		return ErrorResult(tool, idErr, "Use a comparison id like C-001, or omit id to get the next free one.")
	}

	var warnings []Warning
	if isAllocated {
		warnings = append(warnings, NewWarning("W-ID-ALLOCATED", fmt.Sprintf("allocated comparison ID %s", cid), ""))
	}

	val := core.ValidateComparison([]byte(in.Content))
	for _, issue := range val.Issues {
		warnings = append(warnings, WarningFromIssue(issue))
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
		Args: map[string]any{"mode": "validate", "id": cid},
		Why:  fmt.Sprintf("Comparison WEP matrix saved to %s. Comparison scope is complete (optional: run vivechak_review to audit claims).", filepath.ToSlash(relPath)),
	}
	if val.HasBlocking() {
		next = &NextAction{
			Tool: "vivechak_save",
			Args: map[string]any{"kind": "comparison", "id": cid, "revise": true},
			Why:  "Comparison saved as draft with blocking validation issues; fix the listed issues and re-save with revise=true.",
		}
	}

	env := Envelope{
		Ok:       true,
		Message:  fmt.Sprintf("Comparison saved to %s.", filepath.ToSlash(relPath)),
		Data:     data,
		Warnings: warnings,
		Next:     next,
	}
	return env.ToResult()
}

func handleSaveSessionKind(ctx context.Context, ws *store.Workspace, root, rule string, in SaveInput, idWarnings ...Warning) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_save"

	sessionID := strings.TrimSpace(in.ID)
	if sessionID == "" {
		return ErrorResult(tool, fmt.Errorf("E-INVALID-ARGUMENT: id is required when saving a session"),
			"Provide the session identifier in id (e.g. 'T1-01', 'SYN-01').")
	}

	if !isValidID(sessionID) {
		return ErrorResult(tool, fmt.Errorf("E-ID-INVALID: invalid session id %q", sessionID), "Use an ID like 'T1-01'.")
	}

	dag := readPipelineDAG(ws)
	isSynthesis := strings.EqualFold(strings.TrimSpace(in.Kind), "fad") || dag.IsSynthesisID(sessionID)
	isOneWay := false
	if s := dag.SessionByID(sessionID); s != nil {
		isOneWay = strings.EqualFold(s.DoorType, "one-way")
	}

	var val *core.ValidationResult
	if isSynthesis {
		val = core.ValidateFAD([]byte(in.Content))
	} else {
		val = core.ValidateSessionWithContext([]byte(in.Content), isOneWay)
	}

	warnings := append([]Warning{}, idWarnings...)
	for _, issue := range val.Issues {
		warnings = append(warnings, WarningFromIssue(issue))
	}

	qualityObs := core.ObserveSessionQuality([]byte(in.Content))
	for _, q := range qualityObs {
		if q.Severity == "consideration" || q.Severity == "suggestion" {
			warnings = append(warnings, NewWarning("Q-"+strings.ToUpper(q.Category), q.Message, ""))
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
		return ErrorResult(tool, fmt.Errorf("E-LOCK-TIMEOUT: could not acquire lock for %s: %w", relPath, lockErr), "Try again.")
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

	completedSessions, _ := dag.CompletedSessions(ws)
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

	nextWhy := buildNextStep(sessionID, isSynthesis, val.ErrorCount(), qualityObs, dag)
	var next *NextAction
	switch {
	case val.HasBlocking():
		saveKind := "session"
		if isSynthesis {
			saveKind = "fad"
		}
		next = &NextAction{
			Tool: "vivechak_save",
			Args: map[string]any{"kind": saveKind, "id": sessionID, "revise": true},
			Why:  nextWhy,
		}
	case isSynthesis:
		next = &NextAction{
			Tool: "vivechak_gate",
			Args: map[string]any{},
			Why:  nextWhy,
		}
	default:
		next = &NextAction{
			Tool: "vivechak_next",
			Args: map[string]any{},
			Why:  nextWhy,
		}
	}

	env := Envelope{
		Ok:       true,
		Message:  fmt.Sprintf("Session %s saved to %s.", sessionID, relPath),
		Data:     data,
		Warnings: warnings,
		Next:     next,
	}
	return env.ToResult()
}

func buildNextStep(sessionID string, isSynthesis bool, errorCount int, observations []core.QualityObservation, dag *core.DAG) string {
	if errorCount > 0 {
		return fmt.Sprintf("Session saved as draft (%d blocking issues). Fix the issues and re-save with revise=true, "+
			"or run vivechak_review(mode=\"validate\", id=%q).", errorCount, sessionID)
	}

	var parts []string
	var advisories []string
	for _, obs := range observations {
		if obs.Severity == "consideration" || obs.Severity == "suggestion" {
			advisories = append(advisories, fmt.Sprintf("[Q-%s] %s", strings.ToUpper(obs.Category), obs.Message))
		}
	}
	if len(advisories) > 0 {
		parts = append(parts, "QUALITY ADVISORIES: "+strings.Join(advisories, " | "))
	}

	if isSynthesis {
		parts = append(parts, "FAD synthesis saved to research/FAD.md. Run vivechak_gate to evaluate Phase 0 exit criteria before coding.")
		return strings.Join(parts, "\n\n")
	}

	isOneWay := false
	if dag != nil {
		if s := dag.SessionByID(sessionID); s != nil {
			isOneWay = strings.EqualFold(s.DoorType, "one-way")
		}
	}

	if isOneWay {
		parts = append(parts, fmt.Sprintf(
			"ONE-WAY DOOR: Session %s informs an irreversible decision. "+
				"Record the ADR via vivechak_record_decision(id=\"<D-ID>\", auto_draft_from=%q), "+
				"run vivechak_review(mode=\"red_team\", id=\"<D-ID>\"), and cite the returned record_id.",
			sessionID, sessionID))
	} else {
		parts = append(parts, fmt.Sprintf(
			"Run vivechak_next for the next session, or record an ADR via vivechak_record_decision(id=\"<D-ID>\", auto_draft_from=%q).",
			sessionID))
	}

	return strings.Join(parts, "\n\n")
}
