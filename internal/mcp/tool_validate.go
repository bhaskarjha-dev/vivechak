package mcputil

import (
	"context"
	"fmt"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ValidateInput holds the arguments for vivechak_validate.
type ValidateInput struct {
	ProjectRoot  string `json:"project_root,omitempty" jsonschema:"workspace root path"`
	ArtifactType string `json:"artifact_type"           jsonschema:"what to validate: session | decision | plan | fad"`
	Content      string `json:"content"                  jsonschema:"content to validate (Markdown)"`
}

func registerValidate(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_validate",
			Title: "Validate Artifact",
			Description: "Dry-run validation on any Vivechak artifact (session output, plan, " +
				"decision record, or FAD). Returns validation issues at all levels without " +
				"saving anything. Use this to check content before saving. Read-only — no side effects.",
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

	if strings.TrimSpace(in.Content) == "" {
		return ErrorResult(tool, fmt.Errorf("content is required"),
			"Provide the artifact content to validate.")
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
	case "plan", "fad":
		// Plans and FADs have different structures than sessions.
		// Validate frontmatter presence and non-empty body without
		// requiring session-specific fields like session_id.
		validation = core.ValidateArtifact([]byte(in.Content))
	default:
		return ErrorResult(tool, fmt.Errorf("unknown artifact_type %q", artifactType),
			"Use 'session', 'decision', 'plan', or 'fad'.")
	}

	var warnings []string
	for _, issue := range validation.Issues {
		warnings = append(warnings, issue.String())
	}

	var nextStep string
	if validation.HasBlocking() {
		nextStep = "Fix the blocking issues listed above and re-validate, or save as draft."
	} else if validation.WarningCount() > 0 {
		nextStep = fmt.Sprintf("Content is valid with %d warnings. You can save it using the appropriate save tool.", validation.WarningCount())
	} else {
		nextStep = "Content is valid. Save it using vivechak_save_session, vivechak_save_plan, or vivechak_record_decision."
	}

	env := Envelope{
		Success: true,
		Message: fmt.Sprintf("Validation complete: %s (%d issues)", validation.Status, len(validation.Issues)),
		Data: map[string]any{
			"artifact_type": artifactType,
			"status":        validation.Status,
			"validation":    validation,
			"error_count":   validation.ErrorCount(),
			"warning_count": validation.WarningCount(),
		},
		Warnings: warnings,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}
