package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	store "github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RecordDecisionInput holds the arguments for vivechak_record_decision.
type RecordDecisionInput struct {
	ProjectRoot   string `json:"project_root,omitempty"   jsonschema:"workspace root path"`
	ArtifactType  string `json:"artifact_type,omitempty"  jsonschema:"type of artifact: decision | conflict-resolution (default: decision)"`
	DecisionID    string `json:"decision_id"              jsonschema:"decision identifier (e.g. D-015)"`
	Slug          string `json:"slug,omitempty"            jsonschema:"slug for decision filename (optional)"`
	Content       string `json:"content,omitempty"         jsonschema:"decision record or conflict resolution content (Markdown with YAML frontmatter); omit with auto_draft_from to generate a draft"`
	AutoDraftFrom string `json:"auto_draft_from,omitempty" jsonschema:"session ID to auto-draft decision from (e.g. R-01); omit content to get a draft for review"`
}

func registerRecordDecision(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_record_decision",
			Title: "Record Decision",
			Description: "Save an Architectural Decision Record (ADR) or conflict resolution. " +
				"Validates against DECISIONS.template.md or CONFLICT-RESOLUTION.template.md schema. " +
				"Use artifact_type='decision' for ADRs and 'conflict-resolution' for ACH analysis. " +
				"Classifies decisions as one-way or two-way door per P2. " +
				"Supports auto_draft_from: provide a session ID (omit content) to auto-generate a " +
				"draft decision from session findings for review before saving.",
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

	// Auto-draft mode: generate draft from session if auto_draft_from is set and content is empty
	if strings.TrimSpace(in.Content) == "" && strings.TrimSpace(in.AutoDraftFrom) != "" {
		ws, err := store.OpenWorkspace(root)
		if err != nil {
			return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err), "Provide a valid workspace.")
		}
		defer ws.Close()

		sessionsDir := filepath.Join(root, core.SessionsDir)
		content, _, err := core.ReadSessionFilePublic(sessionsDir, in.AutoDraftFrom)
		if err != nil {
			return ErrorResult(tool, fmt.Errorf("reading session %s: %w", in.AutoDraftFrom, err),
				"Ensure the session has been saved with vivechak_save_session first.")
		}
		if content == "" {
			return ErrorResult(tool, fmt.Errorf("session file for %q not found", in.AutoDraftFrom),
				fmt.Sprintf("Ensure session %s has been saved first.", in.AutoDraftFrom))
		}

		draft, err := core.DraftDecisionFromSession([]byte(content), in.DecisionID)
		if err != nil {
			return ErrorResult(tool, fmt.Errorf("auto-drafting decision: %w", err),
				"Check session content and try again.")
		}

		env := Envelope{
			Success: true,
			Message: fmt.Sprintf("Auto-drafted decision %s from session %s", in.DecisionID, in.AutoDraftFrom),
			Data: map[string]any{
				"workspace_root":  root,
				"decision_id":     in.DecisionID,
				"source_session":  in.AutoDraftFrom,
				"draft_content":   draft,
			},
			NextStep: "Review the auto-drafted decision below. Edit as needed, " +
				"then call vivechak_record_decision with the final content.",
			Meta: NewMeta(tool),
		}
		return env.ToResult()
	}

	if strings.TrimSpace(in.Content) == "" {
		return ErrorResult(tool, fmt.Errorf("content is required (or use auto_draft_from to generate a draft)"),
			"Provide the decision record content, or use auto_draft_from with a session ID to generate a draft.")
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

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace: %w", err), "Provide a valid workspace.")
	}
	defer ws.Close()

	// Determine filename, harmonizing with existing ADR files in research/
	var filename string
	switch artifactType {
	case "decision":
		filename = resolveDecisionFilename(ws, in.DecisionID, in.Slug)
	case "conflict-resolution":
		filename = in.DecisionID + "-conflict-resolution.md"
	}

	// Save to research directory
	relPath := filepath.Join(core.ResearchDir, filename)
	unlock, err := store.LockFile(ctx, filepath.Join(root, relPath), 5*time.Second)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("could not acquire lock: %w", err), "Another process may be writing. Try again.")
	}
	defer func() { _ = unlock() }()

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
			defer func() { _ = decUnlock() }()
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

	// Warn if content contains unexpected or nested decision boundary markers that could confuse the registry
	startMarker := fmt.Sprintf("<!-- DECISION: %s -->", in.DecisionID)
	endMarker := fmt.Sprintf("<!-- /DECISION: %s -->", in.DecisionID)
	innerContent := strings.TrimSpace(in.Content)
	innerContent = strings.TrimPrefix(innerContent, startMarker)
	innerContent = strings.TrimSuffix(innerContent, endMarker)
	if strings.Contains(innerContent, "<!-- DECISION:") || strings.Contains(innerContent, "<!-- /DECISION:") {
		warnings = append(warnings, "W-MARKER-CONFLICT: content contains unexpected nested HTML decision markers that may conflict with the registry format")
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
			"vivechak_record_decision with auto_draft_from=[session_id] to auto-draft " +
			"a decision from session findings. Use vivechak_status to review progress.",
		Meta: NewMeta(tool),
	}
	return env.ToResult()
}

// updateOrAppendDecision updates or appends a decision entry into DECISIONS.md.
func updateOrAppendDecision(existing []byte, decisionID string, content string) []byte {
	startMarker := fmt.Sprintf("<!-- DECISION: %s -->", decisionID)
	endMarker := fmt.Sprintf("<!-- /DECISION: %s -->", decisionID)

	cleanContent := strings.TrimSpace(content)
	cleanContent = strings.TrimPrefix(cleanContent, startMarker)
	cleanContent = strings.TrimSuffix(cleanContent, endMarker)
	cleanContent = strings.TrimSpace(cleanContent)

	// Wrap raw YAML frontmatter in <details> block for correct rendering
	cleanContent = wrapFrontmatterForRegistry(cleanContent)

	entry := fmt.Sprintf("%s\n%s\n%s", startMarker, cleanContent, endMarker)

	str := string(existing)
	startIdx := strings.Index(str, startMarker)
	if startIdx >= 0 {
		endRel := strings.Index(str[startIdx:], endMarker)
		if endRel >= 0 && endRel > len(startMarker) {
			endIdx := startIdx + endRel + len(endMarker)
			newStr := str[:startIdx] + entry + str[endIdx:]
			return []byte(newStr)
		}
	}

	// Fallback: look for an unanchored decision entry matching decisionID
	if sIdx, eIdx, ok := findUnanchoredDecision(str, decisionID); ok {
		newStr := str[:sIdx] + entry + str[eIdx:]
		return []byte(newStr)
	}

	if len(strings.TrimSpace(str)) == 0 {
		header := "# Architectural Decisions\n\n"
		return []byte(header + entry + "\n")
	}

	return []byte(strings.TrimRight(str, "\n") + "\n\n---\n\n" + entry + "\n")
}

// wrapFrontmatterForRegistry transforms raw YAML frontmatter (--- delimited)
// into a collapsible <details> block with a YAML code fence inside.
// This prevents markdown renderers from treating mid-file --- as horizontal rules.
func wrapFrontmatterForRegistry(content string) string {
	// Check if content starts with YAML frontmatter
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return content
	}

	// Find the closing ---
	lines := strings.SplitN(trimmed, "\n", 2)
	if len(lines) < 2 {
		return content
	}
	rest := lines[1]
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return content
	}

	yamlBlock := rest[:endIdx]
	afterFrontmatter := strings.TrimSpace(rest[endIdx+4:]) // skip "\n---"

	// Extract summary fields from YAML
	id := extractYAMLField(yamlBlock, "id")
	if id == "" {
		id = extractYAMLField(yamlBlock, "decision_id")
	}
	title := extractYAMLField(yamlBlock, "title")
	doorType := extractYAMLField(yamlBlock, "door_type")
	status := extractYAMLField(yamlBlock, "status")

	// Build summary line
	summaryParts := []string{}
	if id != "" {
		summaryParts = append(summaryParts, "<strong>"+id+"</strong>")
	}
	if title != "" {
		summaryParts = append(summaryParts, title)
	}
	if doorType != "" {
		summaryParts = append(summaryParts, "<code>"+doorType+"</code>")
	}
	if status != "" {
		summaryParts = append(summaryParts, status)
	}
	summaryLine := strings.Join(summaryParts, " · ")

	// Build the <details> block
	var b strings.Builder
	b.WriteString("<details>\n")
	b.WriteString("<summary>" + summaryLine + "</summary>\n\n")
	b.WriteString("```yaml\n")
	b.WriteString(strings.TrimSpace(yamlBlock) + "\n")
	b.WriteString("```\n\n")
	b.WriteString("</details>\n")

	if afterFrontmatter != "" {
		b.WriteString("\n" + afterFrontmatter)
	}

	return b.String()
}

// extractYAMLField extracts a simple scalar value from a YAML block by key name.
// Handles quoted and unquoted values. Not a full YAML parser — sufficient for
// extracting known simple fields from decision frontmatter.
func extractYAMLField(yaml string, key string) string {
	for _, line := range strings.Split(yaml, "\n") {
		trimmed := strings.TrimSpace(line)
		prefix := key + ":"
		if strings.HasPrefix(trimmed, prefix) {
			val := strings.TrimSpace(trimmed[len(prefix):])
			// Strip surrounding quotes
			val = strings.TrimPrefix(val, "\"")
			val = strings.TrimSuffix(val, "\"")
			val = strings.TrimPrefix(val, "'")
			val = strings.TrimSuffix(val, "'")
			return val
		}
	}
	return ""
}

// findUnanchoredDecision locates an unanchored decision block (YAML frontmatter and optional markdown)
// for the given decisionID within DECISIONS.md content.
// It scans sequential frontmatter blocks to prevent greedy cross-entry deletion.
func findUnanchoredDecision(content string, decisionID string) (int, int, bool) {
	// 1. Look for a YAML frontmatter block for this decision.
	// Find pairs of '---' delimiters where the enclosed block contains id / decision_id.
	fmRegex := regexp.MustCompile(`(?ms)^---\s*\n(.*?)\n---\s*`)
	matches := fmRegex.FindAllStringSubmatchIndex(content, -1)

	idPattern := fmt.Sprintf(`(?m)^(?:id|decision_id):\s*["']?%s["']?\s*$`, regexp.QuoteMeta(decisionID))
	idRe := regexp.MustCompile(idPattern)

	for i, idxs := range matches {
		yamlInside := content[idxs[2]:idxs[3]]
		// Disregard if the inside contains a separate delimiter
		if strings.Contains(yamlInside, "\n---") {
			continue
		}

		if idRe.MatchString(yamlInside) {
			startIdx := idxs[0]

			endIdx := len(content)
			if i+1 < len(matches) {
				endIdx = matches[i+1][0]
			}

			// If an anchored decision appears before endIdx, that terminates this entry
			if anchorLoc := strings.Index(content[idxs[1]:endIdx], "<!-- DECISION:"); anchorLoc != -1 {
				endIdx = idxs[1] + anchorLoc
			}

			// If this is the last frontmatter block, check if a header for a DIFFERENT decision follows
			if i == len(matches)-1 {
				hdrRegex := regexp.MustCompile(`(?m)^#{1,4}\s+.*?\b(D-\d+|D-[A-Za-z0-9_-]+)\b`)
				for _, hLoc := range hdrRegex.FindAllStringSubmatchIndex(content[idxs[1]:endIdx], -1) {
					hdrID := content[idxs[1]+hLoc[2] : idxs[1]+hLoc[3]]
					if !strings.EqualFold(hdrID, decisionID) {
						endIdx = idxs[1] + hLoc[0]
						break
					}
				}
			}

			return startIdx, endIdx, true
		}
	}

	// 2. Fallback: look for a Markdown header like "# D-001:" or "## D-001:"
	headerPattern := fmt.Sprintf(`(?m)^#{1,4}\s+.*?\b%s\b.*?\n`, regexp.QuoteMeta(decisionID))
	headerRe := regexp.MustCompile(headerPattern)
	loc := headerRe.FindStringIndex(content)
	if loc == nil {
		return 0, 0, false
	}

	startIdx := loc[0]
	rest := content[loc[1]:]
	boundaryRegex := regexp.MustCompile(`(?m)(?:^<!-- DECISION:|^---\s*\n|^#{1,4}\s+.*?\b(?:D-\d+|D-[A-Za-z0-9_-]+)\b)`)
	if nextLoc := boundaryRegex.FindStringIndex(rest); nextLoc != nil {
		return startIdx, loc[1] + nextLoc[0], true
	}
	return startIdx, len(content), true
}

// resolveDecisionFilename determines the target filename for a decision artifact.
// It prioritizes explicit slugs, then matches existing [ID]-*.md files in the research directory,
// and falls back to [ID]-decision.md.
func resolveDecisionFilename(ws *store.Workspace, decisionID string, slug string) string {
	if slug != "" {
		return fmt.Sprintf("%s-%s.md", decisionID, slug)
	}
	if ws != nil {
		if entries, err := ws.ListDir(core.ResearchDir); err == nil {
			for _, e := range entries {
				name := e.Name()
				if !e.IsDir() && strings.HasPrefix(name, decisionID+"-") &&
					strings.HasSuffix(name, ".md") &&
					!strings.HasSuffix(name, "-plan.md") &&
					!strings.HasSuffix(name, "-conflict-resolution.md") {
					return name
				}
			}
		}
	}
	return decisionID + "-decision.md"
}
