package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ValidateInput holds the arguments for vivechak_validate.
type ValidateInput struct {
	ProjectRoot  string `json:"project_root,omitempty" jsonschema:"workspace root path"`
	ArtifactType string `json:"artifact_type,omitempty" jsonschema:"what to validate: session | decision | plan | fad | conflict (omit for workspace-wide validation)"`
	Content      string `json:"content,omitempty"        jsonschema:"content to validate (Markdown); omit with project_root for workspace-wide validation"`
}

func registerValidate(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_validate",
			Title: "Validate Artifact",
			Description: "Dry-run validation on any Vivechak artifact (session output, plan, " +
				"decision record, or FAD). Returns validation issues at all levels without " +
				"saving anything. Use this to check content before saving. " +
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
	case "decision":
		validation = core.ValidateDecision([]byte(in.Content))
	case "conflict-resolution", "conflict":
		validation = core.ValidateConflictResolution([]byte(in.Content))
	case "plan":
		validation = core.ValidatePlan([]byte(in.Content))
	case "fad":
		validation = core.ValidateFAD([]byte(in.Content))
	default:
		return ErrorResult(tool, fmt.Errorf("unknown artifact_type %q", artifactType),
			"Use 'session', 'decision', 'plan', 'fad', or 'conflict'.")
	}

	var warnings []string
	for _, issue := range validation.Issues {
		warnings = append(warnings, issue.String())
	}

	// Gather advisory quality observations
	observations := core.ObserveSessionQuality([]byte(in.Content))

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
				v := core.ValidateSession(data)
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
	if entries, err := ws.ListDir(core.ResearchDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") {
				continue
			}
			if strings.HasSuffix(name, "-plan.md") ||
				strings.HasSuffix(name, "-comparison.md") ||
				strings.HasSuffix(name, "-conflict-resolution.md") ||
				strings.EqualFold(name, "DECISIONS.md") ||
				strings.EqualFold(name, "FAD.md") ||
				strings.EqualFold(name, "RESEARCH-PIPELINE.md") {
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

	// Validate pipeline if it exists
	if data, err := ws.ReadFile(core.PipelineFile); err == nil && len(data) > 0 {
		v := core.ValidatePlan(data)
		pipeResult := map[string]any{
			"status":        v.Status,
			"error_count":   v.ErrorCount(),
			"warning_count": v.WarningCount(),
		}
		if v.HasBlocking() {
			for _, issue := range v.BlockingIssues() {
				allWarnings = append(allWarnings, fmt.Sprintf("PIPELINE: %s", issue.String()))
			}
		}
		results["pipeline"] = pipeResult
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
