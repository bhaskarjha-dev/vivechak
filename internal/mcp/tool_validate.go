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
	ArtifactType string `json:"artifact_type"           jsonschema:"what to validate: session | decision | plan | fad | conflict"`
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
