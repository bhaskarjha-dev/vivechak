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

func handleRecordDecision(ctx context.Context, _ *sdkmcp.CallToolRequest, in RecordDecisionInput) (*sdkmcp.CallToolResult, Envelope, error) {
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

	// Validate decision ID format to prevent path traversal
	if !isValidID(in.DecisionID) {
		return ErrorResult(tool, fmt.Errorf("invalid decision_id %q — must be alphanumeric with optional hyphens, underscores, dots", in.DecisionID),
			"Use a simple ID like 'D-01'.")
	}

	if strings.TrimSpace(in.Content) == "" {
		return ErrorResult(tool, fmt.Errorf("content is required"),
			"Provide the decision record content (Markdown with YAML frontmatter).")
	}

	if err := validateContentSize(in.Content); err != nil {
		return ErrorResult(tool, err, "Reduce content size or split into multiple artifacts.")
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
	var validation *core.ValidationResult
	switch artifactType {
	case "conflict-resolution":
		validation = core.ValidateConflictResolution([]byte(in.Content))
	default:
		validation = core.ValidateDecision([]byte(in.Content))
	}

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
	unlock, err := store.LockFile(ctx, filepath.Join(root, relPath), 5*time.Second)
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

	// Dual-write to DECISIONS.md registry for decision artifacts
	if artifactType == "decision" {
		decRelPath := core.DecisionsFile
		decUnlock, decErr := store.LockFile(ctx, filepath.Join(root, decRelPath), 5*time.Second)
		if decErr != nil {
			warnings = append(warnings, fmt.Sprintf("W-DECISIONS-LOCK: Could not acquire lock on %s: %v", decRelPath, decErr))
		} else {
			defer decUnlock()
			var existing []byte
			if data, err := ws.ReadFile(decRelPath); err == nil {
				existing = data
			}
			newDecContent := updateOrAppendDecision(existing, in.DecisionID, in.Content)
			if err := store.WriteFileAtomic(ws.Root(), decRelPath, newDecContent, 0o644); err != nil {
				warnings = append(warnings, fmt.Sprintf("W-DECISIONS-WRITE: Failed to update %s: %v", decRelPath, err))
			}
		}
	}

	// Warn if content contains decision boundary markers that could confuse the registry
	if strings.Contains(in.Content, "<!-- DECISION:") || strings.Contains(in.Content, "<!-- /DECISION:") {
		warnings = append(warnings, "W-MARKER-CONFLICT: content contains HTML decision markers that may conflict with the registry format")
	}

	dataMap := map[string]any{
		"workspace_root": root,
		"decision_id":    in.DecisionID,
		"artifact_type":  artifactType,
		"file_path":      filepath.Join(core.ResearchDir, filename),
		"status":         validation.Status,
		"validation":     validation,
	}
	if artifactType == "decision" {
		dataMap["decisions_file"] = core.DecisionsFile
	}

	env := Envelope{
		Success: true,
		Message: fmt.Sprintf("Saved %s %s as %s", artifactType, in.DecisionID, validation.Status),
		Data:    dataMap,
		Warnings: warnings,
		NextStep: "Run vivechak_next_session for the next research session, or " +
			"vivechak_status to review overall progress.",
		Meta: NewMeta(tool),
	}
	return env.ToResult()
}

// updateOrAppendDecision updates or appends a decision entry into DECISIONS.md.
func updateOrAppendDecision(existing []byte, decisionID string, content string) []byte {
	startMarker := fmt.Sprintf("<!-- DECISION: %s -->", decisionID)
	endMarker := fmt.Sprintf("<!-- /DECISION: %s -->", decisionID)
	entry := fmt.Sprintf("%s\n%s\n%s", startMarker, strings.TrimSpace(content), endMarker)

	str := string(existing)
	startIdx := strings.Index(str, startMarker)
	if startIdx >= 0 {
		endRel := strings.Index(str[startIdx:], endMarker)
		if endRel >= 0 {
			endIdx := startIdx + endRel + len(endMarker)
			newStr := str[:startIdx] + entry + str[endIdx:]
			return []byte(newStr)
		}
	}

	if len(strings.TrimSpace(str)) == 0 {
		header := "# Architectural Decisions\n\n"
		return []byte(header + entry + "\n")
	}

	return []byte(strings.TrimRight(str, "\n") + "\n\n---\n\n" + entry + "\n")
}
