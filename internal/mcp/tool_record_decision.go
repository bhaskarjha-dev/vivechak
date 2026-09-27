package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	store "github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RecordDecisionInput holds the arguments for vivechak_record_decision.
type RecordDecisionInput struct {
	ProjectRoot  string `json:"project_root,omitempty"  jsonschema:"workspace root path"`
	ArtifactType string `json:"artifact_type"            jsonschema:"type of artifact: decision | conflict-resolution"`
	DecisionID   string `json:"decision_id"              jsonschema:"decision identifier (e.g. D-015)"`
	Content      string `json:"content"                   jsonschema:"decision record or conflict resolution content (Markdown with YAML frontmatter)"`
}

func registerRecordDecision(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_record_decision",
			Title: "Record Decision",
			Description: "Save an Architectural Decision Record (ADR) or conflict resolution. " +
				"Validates against DECISIONS.template.md or CONFLICT-RESOLUTION.template.md schema. " +
				"Use artifact_type='decision' for ADRs and 'conflict-resolution' for ACH analysis. " +
				"Classifies decisions as one-way or two-way door per P2.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleRecordDecision,
	)
}

func handleRecordDecision(_ context.Context, _ *sdkmcp.CallToolRequest, in RecordDecisionInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_record_decision"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err,
			"Initialize a workspace first with vivechak_init.")
	}

	if strings.TrimSpace(in.DecisionID) == "" {
		return ErrorResult(tool, fmt.Errorf("decision_id is required"),
			"Provide a decision identifier like 'D-015'.")
	}

	if strings.TrimSpace(in.Content) == "" {
		return ErrorResult(tool, fmt.Errorf("content is required"),
			"Provide the decision record content (Markdown with YAML frontmatter).")
	}

	// Validate artifact type
	artifactType := in.ArtifactType
	if artifactType == "" {
		artifactType = "decision"
	}
	if artifactType != "decision" && artifactType != "conflict-resolution" {
		return ErrorResult(tool, fmt.Errorf("invalid artifact_type %q", artifactType),
			"Use 'decision' for ADRs or 'conflict-resolution' for ACH analysis.")
	}

	// Validate content
	validation := core.ValidateDecision([]byte(in.Content))

	// Determine filename
	var filename string
	switch artifactType {
	case "decision":
		filename = in.DecisionID + "-decision.md"
	case "conflict-resolution":
		filename = in.DecisionID + "-conflict-resolution.md"
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err), "Provide a valid workspace.")
	}
	defer ws.Close()

	// Save to research directory
	relPath := filepath.Join(core.ResearchDir, filename)
	unlock, err := store.LockFile(filepath.Join(root, relPath), 5*time.Second)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("could not acquire lock: %w", err), "Another process may be writing. Try again.")
	}
	defer unlock()

	if err := store.WriteFileAtomic(ws.Root(), relPath, []byte(in.Content), 0o644); err != nil {
		return ErrorResult(tool, fmt.Errorf("writing %s: %w", filename, err),
			"Check filesystem permissions.")
	}

	var warnings []string
	for _, issue := range validation.Issues {
		warnings = append(warnings, issue.String())
	}

	env := Envelope{
		Success: true,
		Message: fmt.Sprintf("Saved %s %s as %s", artifactType, in.DecisionID, validation.Status),
		Data: map[string]any{
			"workspace_root": root,
			"decision_id":    in.DecisionID,
			"artifact_type":  artifactType,
			"file_path":      filepath.Join(core.ResearchDir, filename),
			"status":         validation.Status,
			"validation":     validation,
		},
		Warnings: warnings,
		NextStep: "Run vivechak_next_session for the next research session, or " +
			"vivechak_status to review overall progress.",
		Meta: NewMeta(tool),
	}
	return env.ToResult()
}
