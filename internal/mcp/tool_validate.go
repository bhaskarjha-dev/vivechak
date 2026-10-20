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

// ValidateInput holds the arguments for vivechak_validate.
type ValidateInput struct {
	ProjectRoot  string `json:"project_root,omitempty" jsonschema:"workspace root path"`
	ArtifactType string `json:"artifact_type,omitempty" jsonschema:"what to validate: session | decision | decisions | comparison | plan | fad | conflict (omit for workspace-wide validation)"`
	Content      string `json:"content,omitempty"        jsonschema:"content to validate (Markdown); omit with project_root for workspace-wide validation"`
}

func registerValidate(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_validate",
			Title: "Validate Artifact",
			Description: "Dry-run validation on any Vivechak artifact (session output, plan, " +
				"decision record, comparison matrix, or FAD). Use for dry-run validation of in-progress content only — " +
				"save_session validates automatically, so calling validate before save doubles token cost. " +
				"Can also validate an entire workspace: omit content and provide only project_root " +
				"to validate all artifacts in the workspace. Read-only — no side effects.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleValidate,
	)
}

func handleValidate(_ context.Context, _ *sdkmcp.CallToolRequest, in ValidateInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_validate"

	// Workspace-wide validation: content empty
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
		// Auto-detect multi-decision registries (such as DECISIONS.md with introductory preambles or HTML comment blocks)
		if artifactType == "decisions" || strings.Contains(in.Content, "<!-- DECISION:") || len(core.ExtractUnanchoredDecisionBlocks(contentBytes)) > 1 || (strings.Contains(in.Content, "#") && !strings.HasPrefix(strings.TrimSpace(in.Content), "---")) {
			validation = core.ValidateDecisionRegistry(contentBytes)
		} else {
			validation = core.ValidateDecision(contentBytes)
		}
	case "conflict-resolution", "conflict":
		validation = core.ValidateConflictResolution([]byte(in.Content))
	case "comparison", "comp":
		validation = core.ValidateComparison([]byte(in.Content))
	case "plan":
		validation = core.ValidatePlan([]byte(in.Content))
	case "fad":
		validation = core.ValidateFAD([]byte(in.Content))
	default:
		return ErrorResult(tool, fmt.Errorf("unknown artifact_type %q", artifactType),
			"Use 'session', 'decision', 'decisions', 'comparison', 'plan', 'fad', or 'conflict'.")
	}

	var warnings []string
	for _, issue := range validation.Issues {
		warnings = append(warnings, issue.String())
	}

	// Gather advisory quality observations for session artifacts
	var observations []core.QualityObservation
	if artifactType == "session" {
		observations = core.ObserveSessionQuality([]byte(in.Content))
	}

	var nextStep string
	if validation.HasBlocking() {
		nextStep = "Review the structural issues above. Address any that affect correctness, then save."
	} else if validation.WarningCount() > 0 {
		nextStep = fmt.Sprintf("Content is structurally valid with %d observations. "+
			"Review the suggestions if you want to strengthen the output, then save.", validation.WarningCount())
	} else {
		nextStep = "Content is valid. Save it using vivechak_save_session, vivechak_save_plan, or vivechak_record_decision."
	}

	responseData := map[string]any{
		"artifact_type": artifactType,
		"status":        validation.Status,
		"validation":    validation,
		"error_count":   validation.ErrorCount(),
		"warning_count": validation.WarningCount(),
	}
	if len(observations) > 0 {
		responseData["observations"] = observations
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Validation complete: %s (%d issues)", validation.Status, len(validation.Issues)),
		Data:     responseData,
		Warnings: warnings,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// handleWorkspaceValidate validates all artifacts in a workspace and returns aggregated results.
func handleWorkspaceValidate(tool string, projectRoot string) (*sdkmcp.CallToolResult, Envelope, error) {
	root, err := core.ResolveWorkspace(projectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_init.")
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
	// DECISIONS.md is a compiled view and is not validated directly to prevent double-counting.
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
			// Whitelist: only D-* prefixed files or files with door_type in frontmatter
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

		// Also check unanchored decision blocks in DECISIONS.md
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

	// Validate comparisons if C-*.md, *-comparison.md, *-cmp-*.md, or COMPARISON.md exist
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
		results["comparison"] = compResults[0] // backward-compatible alias for first comparison
	}

	totalIssues := len(allWarnings)
	var nextStep string
	if totalIssues > 0 {
		nextStep = fmt.Sprintf("Workspace has %d issue(s). Review the warnings above and address any blocking issues.", totalIssues)
	} else {
		nextStep = "Workspace validation passed. All artifacts are structurally valid."
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Workspace validation complete (%d issue(s) found)", totalIssues),
		Data:     results,
		Warnings: allWarnings,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}
