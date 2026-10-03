package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
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

	// Acquire decisions registry lock for ADR decisions to serialize write + registry compilation
	var decUnlock func() error
	if artifactType == "decision" {
		var lockErr error
		decUnlock, lockErr = store.LockFile(ctx, filepath.Join(root, core.DecisionsFile), 10*time.Second)
		if lockErr != nil {
			return ErrorResult(tool, fmt.Errorf("could not acquire decisions registry lock: %w", lockErr), "Another process is recording decisions. Try again.")
		}
		defer func() {
			if decUnlock != nil {
				_ = decUnlock()
			}
		}()
	}

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

	// Compile DECISIONS.md registry from all individual ADR files
	if artifactType == "decision" {
		if compileErr := compileDecisionsRegistry(ctx, ws, root); compileErr != nil {
			warnings = append(warnings, fmt.Sprintf(
				"W-DECISIONS-COMPILE: Failed to compile DECISIONS.md: %v", compileErr))
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

	// Mini-status (eliminates need for separate status calls)
	completedSessions, _ := scanCompletedSessions(ws)
	gateDecisions := collectGateDecisions(ws)
	progress := map[string]any{
		"sessions_completed": len(completedSessions),
		"decisions_recorded": len(gateDecisions),
	}
	info := core.InspectWorkspace(root)
	if info.HasPipeline {
		if pipeData, err := ws.ReadFile(core.PipelineFile); err == nil {
			if dag, err := core.ParsePipeline(pipeData); err == nil {
				progress["sessions_total"] = len(dag.Sessions)
			}
		}
	}
	dataMap["progress"] = progress

	nextStep := "Run vivechak_next_session for the next research session, or " +
		"vivechak_record_decision with auto_draft_from=[session_id] to auto-draft " +
		"a decision from session findings. Use vivechak_status to review progress."
	if fm, _, _ := core.ParseFrontmatter([]byte(in.Content)); fm != nil {
		if strings.EqualFold(fm.GetString("door_type"), "one-way") {
			nextStep += " (one-way door — requires Grade A/B evidence, specific reversal triggers, and documented rejected alternatives)"
		}
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Saved %s %s as %s", artifactType, in.DecisionID, validation.Status),
		Data:     dataMap,
		Warnings: warnings,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}
	return env.ToResult()
}

// compileDecisionsRegistry compiles DECISIONS.md from all individual ADR files in research/,
// while preserving planned/unmaterialized decisions already registered in DECISIONS.md.
func compileDecisionsRegistry(ctx context.Context, ws *store.Workspace, root string) error {
	decRelPath := core.DecisionsFile
	existingByID := make(map[string]string)
	var preamble string

	// 1. Read existing DECISIONS.md to preserve planned decisions and preamble
	if existingData, err := ws.ReadFile(decRelPath); err == nil && len(existingData) > 0 {
		anchoredRe := regexp.MustCompile(`(?s)<!-- DECISION:\s*([A-Za-z0-9_-]+)\s*-->\s*(.*?)\s*<!-- /DECISION:\s*[A-Za-z0-9_-]+\s*-->`)
		matches := anchoredRe.FindAllSubmatchIndex(existingData, -1)

		earliestDecisionStart := -1

		for _, loc := range matches {
			if earliestDecisionStart == -1 || loc[0] < earliestDecisionStart {
				earliestDecisionStart = loc[0]
			}
			id := string(existingData[loc[2]:loc[3]])
			body := string(existingData[loc[4]:loc[5]])
			existingByID[id] = body
		}

		isInsideAnchors := func(start, end int) bool {
			for _, loc := range matches {
				if start >= loc[0] && end <= loc[1] {
					return true
				}
			}
			return false
		}

		// Also check for any unanchored frontmatter blocks (whether or not anchored blocks exist)
		fmRegex := regexp.MustCompile(`(?ms)^---\s*\n(.*?)\n---\s*`)
		fmMatches := fmRegex.FindAllSubmatchIndex(existingData, -1)
		for _, loc := range fmMatches {
			if isInsideAnchors(loc[0], loc[1]) {
				continue
			}
			if fm, _, pErr := core.ParseFrontmatter(existingData[loc[0]:loc[1]]); pErr == nil && fm != nil {
				id := fm.GetString("id")
				if id == "" {
					id = fm.GetString("decision_id")
				}
				if id != "" {
					if earliestDecisionStart == -1 || loc[0] < earliestDecisionStart {
						earliestDecisionStart = loc[0]
					}
					// Find end of this unanchored decision: next anchor or next unanchored fm block or EOF
					endIdx := len(existingData)
					for _, aLoc := range matches {
						if aLoc[0] > loc[0] && aLoc[0] < endIdx {
							endIdx = aLoc[0]
						}
					}
					for _, otherFm := range fmMatches {
						if otherFm[0] > loc[0] && otherFm[0] < endIdx && !isInsideAnchors(otherFm[0], otherFm[1]) {
							endIdx = otherFm[0]
						}
					}
					existingByID[id] = string(existingData[loc[0]:endIdx])
				}
			}
		}

		if earliestDecisionStart > 0 {
			preamble = string(existingData[:earliestDecisionStart])
		} else if earliestDecisionStart == -1 {
			preamble = string(existingData)
		}
	}

	// 2. Scan individual ADR files in research/
	entries, err := ws.ListDir(core.ResearchDir)
	if err != nil {
		return fmt.Errorf("listing research directory: %w", err)
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		if core.IsSpecialResearchFile(name) {
			continue
		}
		// Whitelist: D-* files or files with door_type in frontmatter
		isCandidate := strings.HasPrefix(strings.ToUpper(name), "D-")
		var data []byte
		var readErr error
		for attempt := 0; attempt < 5; attempt++ {
			data, readErr = ws.ReadFile(filepath.Join(core.ResearchDir, name))
			if readErr == nil && len(data) > 0 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if readErr != nil || len(data) == 0 {
			continue
		}
		fm, _, err := core.ParseFrontmatter(data)
		if err != nil || fm == nil {
			continue
		}
		if !isCandidate && !fm.Has("door_type") {
			continue
		}
		id := fm.GetString("id")
		if id == "" {
			id = fm.GetString("decision_id")
		}
		if id == "" {
			stem := strings.TrimSuffix(name, ".md")
			id = stem
		}
		existingByID[id] = string(data)
	}

	type decisionEntry struct {
		id      string
		content string
	}
	var decisions []decisionEntry
	for id, content := range existingByID {
		decisions = append(decisions, decisionEntry{
			id:      id,
			content: content,
		})
	}

	// Sort decisions deterministically by natural ID
	sort.Slice(decisions, func(i, j int) bool {
		return naturalDecisionIDLess(decisions[i].id, decisions[j].id)
	})

	var b strings.Builder
	trimmedPreamble := strings.TrimSpace(preamble)
	if trimmedPreamble != "" {
		b.WriteString(trimmedPreamble)
		b.WriteString("\n\n")
	} else {
		b.WriteString("# Architectural Decisions\n\n")
	}
	for i, d := range decisions {
		wrapped := wrapFrontmatterForRegistry(d.content)
		fmt.Fprintf(&b, "<!-- DECISION: %s -->\n%s\n<!-- /DECISION: %s -->\n", d.id, wrapped, d.id)
		if i < len(decisions)-1 {
			b.WriteString("\n---\n\n")
		}
	}

	return store.WriteFileAtomic(ws.Root(), decRelPath, []byte(b.String()), 0o644)
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

// naturalDecisionIDLess sorts decision IDs naturally so D-2 sorts before D-10.
func naturalDecisionIDLess(a, b string) bool {
	var numA, numB int
	if n, err := fmt.Sscanf(strings.ToUpper(a), "D-%d", &numA); n == 1 && err == nil {
		if m, err := fmt.Sscanf(strings.ToUpper(b), "D-%d", &numB); m == 1 && err == nil {
			if numA != numB {
				return numA < numB
			}
		}
	}
	return a < b
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
					!core.IsSpecialResearchFile(name) {
					return name
				}
			}
		}
	}
	return decisionID + "-decision.md"
}
