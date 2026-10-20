package core

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// ValidationLevel represents the severity of a validation finding.
// Per FINAL-PLAN.md: L1 Construct → L2 Block → L3 Warn → L4 Gate
type ValidationLevel int

const (
	// L1Construct auto-fills missing fields. Lowest severity.
	L1Construct ValidationLevel = 1
	// L2Block saves as draft with issues. Prevents marking as complete.
	L2Block ValidationLevel = 2
	// L3Warn saves but flags issues. Comply-or-explain.
	L3Warn ValidationLevel = 3
	// L4Gate is project-wide structural completeness. Highest severity.
	L4Gate ValidationLevel = 4
)

// String returns the level name.
func (l ValidationLevel) String() string {
	switch l {
	case L1Construct:
		return "L1-CONSTRUCT"
	case L2Block:
		return "L2-BLOCK"
	case L3Warn:
		return "L3-WARN"
	case L4Gate:
		return "L4-GATE"
	default:
		return fmt.Sprintf("L%d", l)
	}
}

// ValidationIssue is a single finding from validation.
type ValidationIssue struct {
	// Prefix is an optional scope/entity prefix (e.g. "DECISION D-001: ").
	Prefix string `json:"prefix,omitempty"`

	// Level is the severity of this issue.
	Level ValidationLevel `json:"level"`

	// Code is a machine-readable issue code (e.g., "V-MISSING-FRONTMATTER").
	Code string `json:"code"`

	// Message is a human-readable description.
	Message string `json:"message"`

	// Field is the specific field or section with the issue (optional).
	Field string `json:"field,omitempty"`

	// FixHint suggests how to resolve this issue (optional).
	FixHint string `json:"fix_hint,omitempty"`
}

func (v ValidationIssue) String() string {
	s := fmt.Sprintf("[%s] %s: %s", v.Level, v.Code, v.Message)
	if v.FixHint != "" {
		s += " (fix: " + v.FixHint + ")"
	}
	if v.Prefix != "" {
		s = v.Prefix + s
	}
	return s
}

// ValidationResult holds the complete set of findings from validation.
type ValidationResult struct {
	// Issues is the list of all findings.
	Issues []ValidationIssue `json:"issues,omitempty"`

	// Status summarizes the result.
	Status string `json:"status"`
}

// MaxLevel returns the highest severity level found.
func (r *ValidationResult) MaxLevel() ValidationLevel {
	var max ValidationLevel
	for _, issue := range r.Issues {
		if issue.Level > max {
			max = issue.Level
		}
	}
	return max
}

// HasBlocking reports whether any blocking (L2Block or L4Gate) issues were found.
func (r *ValidationResult) HasBlocking() bool {
	for _, issue := range r.Issues {
		if issue.Level == L2Block || issue.Level == L4Gate {
			return true
		}
	}
	return false
}

// WarningCount returns the number of L3 warnings.
func (r *ValidationResult) WarningCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Level == L3Warn {
			count++
		}
	}
	return count
}

// ErrorCount returns the number of L2 blocking issues.
func (r *ValidationResult) ErrorCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Level == L2Block {
			count++
		}
	}
	return count
}

// BlockingIssues returns all issues with Level == L2Block or L4Gate.
func (r *ValidationResult) BlockingIssues() []ValidationIssue {
	var blocking []ValidationIssue
	for _, issue := range r.Issues {
		if issue.Level == L2Block || issue.Level == L4Gate {
			blocking = append(blocking, issue)
		}
	}
	return blocking
}

// AddIssue appends a validation issue.
func (r *ValidationResult) AddIssue(level ValidationLevel, code, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Level:   level,
		Code:    code,
		Message: message,
	})
}

// AddIssueWithHint appends a validation issue with a fix hint.
func (r *ValidationResult) AddIssueWithHint(level ValidationLevel, code, message, fixHint string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Level:   level,
		Code:    code,
		Message: message,
		FixHint: fixHint,
	})
}

// AddFieldIssueWithHint appends a validation issue with a field and a fix hint.
func (r *ValidationResult) AddFieldIssueWithHint(level ValidationLevel, code, field, message, fixHint string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Level:   level,
		Code:    code,
		Field:   field,
		Message: message,
		FixHint: fixHint,
	})
}

// ValidateDiscoveredConcerns verifies that Discovered Concerns contains substantive analysis on one-way door sessions.
func ValidateDiscoveredConcerns(bodyStr string, isOneWay bool, result *ValidationResult) {
	hasHeading := hasSectionHeading(bodyStr, "discovered concerns", "discovered concern")
	content := ExtractSection(bodyStr, "discovered concerns", "discovered concern")
	trimmed := strings.TrimSpace(content)

	isEmpty := !hasHeading || len(trimmed) < 50 ||
		strings.EqualFold(trimmed, "None.") ||
		strings.EqualFold(trimmed, "None identified.") ||
		strings.EqualFold(trimmed, "N/A") ||
		strings.EqualFold(trimmed, "No additional concerns.")

	if isEmpty && isOneWay {
		result.AddIssueWithHint(L2Block, "V-OWD-NO-CONCERNS",
			"One-way door sessions require substantive Discovered Concerns (≥50 chars).",
			"Document at least one unexpected constraint, trade-off, or failure mode discovered outside the original research scope.")
	} else if hasHeading && (len(trimmed) == 0 || strings.EqualFold(trimmed, "None.") || strings.EqualFold(trimmed, "N/A") || strings.EqualFold(trimmed, "None identified.")) {
		result.AddIssueWithHint(L3Warn, "W-NO-CONCERNS",
			"Discovered Concerns section is empty or trivial.",
			"Consider documenting unexpected findings, integration risks, or edge cases.")
	}
}

// ValidateSession checks a research session output against the validation ladder.
// Returns issues at levels L1-L3 (L4 is project-wide, not per-session).
func ValidateSession(data []byte) *ValidationResult {
	return ValidateSessionWithContext(data, false)
}

// ValidateSessionWithContext validates a session with explicit one-way door context.
func ValidateSessionWithContext(data []byte, isOneWay bool) *ValidationResult {
	result := &ValidationResult{Status: "valid"}

	fm, body, err := ParseFrontmatter(data)
	if err != nil {
		result.AddIssue(L2Block, "V-INVALID-FRONTMATTER", "YAML frontmatter is malformed: "+err.Error())
		result.Status = "invalid"
		return result
	}

	// L2: Frontmatter must exist
	if fm == nil {
		result.AddIssueWithHint(L2Block, "V-MISSING-FRONTMATTER",
			"Session output has no YAML frontmatter",
			"Add frontmatter with at least: session_id, title, date, status")
		result.Status = "draft"
		return result
	}

	if fm.Has("door_type") && strings.EqualFold(fm.GetString("door_type"), "one-way") {
		isOneWay = true
	}

	// L2: Required frontmatter fields
	requiredFields := []string{"session_id", "title", "date"}
	for _, field := range requiredFields {
		has := fm.Has(field)
		// Accept 'id' as alias for 'session_id'
		if !has && field == "session_id" {
			has = fm.Has("id")
		}
		// Accept 'synthesis_date' as alias for 'date'
		if !has && field == "date" {
			has = fm.Has("synthesis_date")
		}
		if !has {
			hint := fmt.Sprintf("Add '%s: <value>' to the frontmatter block", field)
			if field == "session_id" {
				hint = "Add 'session_id: <value>' (or 'id: <value>') to the frontmatter block"
			}
			result.AddFieldIssueWithHint(L2Block, "V-MISSING-FIELD", field,
				fmt.Sprintf("Required frontmatter field %q is missing", field),
				hint)
		}
	}

	// L3: Check date format (YYYY-MM-DD)
	dateKey := "date"
	if !fm.Has("date") && fm.Has("synthesis_date") {
		dateKey = "synthesis_date"
	}
	if fm.Has(dateKey) {
		dateStr := fm.GetString(dateKey)
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr); !matched {
			result.AddIssueWithHint(L3Warn, "W-INVALID-DATE-FORMAT",
				fmt.Sprintf("%s field is not in YYYY-MM-DD format", dateKey),
				fmt.Sprintf("Use ISO 8601 format: %s: 2026-09-29", dateKey))
		}
	}

	// L2: Body must not be empty
	bodyStr := strings.TrimSpace(string(body))
	if bodyStr == "" {
		result.AddIssue(L2Block, "V-EMPTY-BODY", "Session body is empty")
	}

	// L3: Evidence grades and recalled grade cap checks
	checkEvidenceGrades(bodyStr, result)
	checkRecalledGradeCap(bodyStr, result)

	// L2/L3: Provenance and concerns checks
	ValidateEvidenceProvenance(bodyStr, isOneWay, result)
	ValidateDiscoveredConcerns(bodyStr, isOneWay, result)

	// L3: Check for status field
	if !fm.Has("status") {
		result.AddIssueWithHint(L1Construct, "V-MISSING-STATUS",
			"No 'status' field in frontmatter — will default to 'draft'",
			"Add 'status: draft' or 'status: complete' to frontmatter")
	}

	// Update status based on findings
	if result.HasBlocking() {
		result.Status = "draft"
	} else if result.WarningCount() > 0 {
		result.Status = "valid-with-warnings"
	}

	return result
}

// ValidateDecision checks an ADR (Architectural Decision Record) output.
func ValidateDecision(data []byte) *ValidationResult {
	result := &ValidationResult{Status: "valid"}

	fm, body, err := ParseFrontmatter(data)
	if err != nil {
		result.AddIssue(L2Block, "V-INVALID-FRONTMATTER", "YAML frontmatter is malformed: "+err.Error())
		result.Status = "invalid"
		return result
	}

	if fm == nil {
		result.AddIssueWithHint(L2Block, "V-MISSING-FRONTMATTER",
			"Decision record has no YAML frontmatter",
			"Add frontmatter with at least: decision_id, title, status, door_type")
		result.Status = "draft"
		return result
	}

	// L2: Required fields for decisions
	requiredFields := []string{"decision_id", "title", "status"}
	for _, field := range requiredFields {
		has := fm.Has(field)
		// Accept 'id' as alias for 'decision_id'
		if !has && field == "decision_id" {
			has = fm.Has("id")
		}
		if !has {
			result.AddFieldIssueWithHint(L2Block, "V-MISSING-FIELD", field,
				fmt.Sprintf("Required field %q missing from decision", field),
				fmt.Sprintf("Add '%s: <value>' to the frontmatter", field))
		}
	}

	// L3: door_type should be present
	if !fm.Has("door_type") {
		result.AddIssueWithHint(L3Warn, "W-MISSING-DOOR-TYPE",
			"Decision lacks door_type classification",
			"Add 'door_type: one-way' or 'door_type: two-way' per P2")
	}

	// L3: Body should not be trivially short
	bodyStr := strings.TrimSpace(string(body))
	if len(bodyStr) < 100 {
		result.AddIssueWithHint(L3Warn, "W-SHORT-DECISION",
			"Decision body is very short — may lack sufficient context",
			"Include Context, Decision, Consequences, and Evidence sections")
	}

	// L3: Evidence grades and recalled grade cap checks (only for researched/accepted decisions)
	status := strings.ToLower(fm.GetString("status"))
	if status != "proposed" {
		checkEvidenceGrades(bodyStr, result)
		checkRecalledGradeCap(bodyStr, result)
	}

	if result.HasBlocking() {
		result.Status = "draft"
	} else if result.WarningCount() > 0 {
		result.Status = "valid-with-warnings"
	}

	return result
}

// ExtractUnanchoredDecisionBlocks finds all individual decision blocks in a multi-decision file (e.g. DECISIONS.md).
// Each block includes its YAML frontmatter and the associated body text up to the next block or EOF.
func ExtractUnanchoredDecisionBlocks(content []byte) [][]byte {
	lines := strings.Split(string(content), "\n")
	type decLoc struct {
		startLine int
		fmEndLine int
		id        string
	}
	var decs []decLoc

	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "---" {
			continue
		}
		// Look ahead for the closing fence and verify an id or decision_id exists
		closeLine := -1
		hasID := false
		rawID := ""
		for j := i + 1; j < len(lines); j++ {
			tr := strings.TrimSpace(lines[j])
			if tr == "---" {
				closeLine = j
				break
			}
			if strings.HasPrefix(tr, "id:") || strings.HasPrefix(tr, "decision_id:") {
				hasID = true
				parts := strings.SplitN(tr, ":", 2)
				if len(parts) == 2 {
					rawID = strings.TrimSpace(parts[1])
					rawID = strings.Trim(rawID, `"'`)
				}
			}
		}

		if closeLine != -1 && hasID {
			fmText := strings.Join(lines[i:closeLine+1], "\n")
			id := rawID
			if fm, _, err := ParseFrontmatter([]byte(fmText)); err == nil && fm != nil {
				if parsedID := fm.GetString("id"); parsedID != "" {
					id = parsedID
				} else if parsedID := fm.GetString("decision_id"); parsedID != "" {
					id = parsedID
				}
			}
			if id != "" && !strings.HasPrefix(strings.ToUpper(id), "CHK-") {
				decs = append(decs, decLoc{startLine: i, fmEndLine: closeLine, id: id})
				i = closeLine // advance past this frontmatter
			}
		}
	}

	if len(decs) == 0 {
		return nil
	}

	var blocks [][]byte
	for i := 0; i < len(decs); i++ {
		start := decs[i].startLine
		var end int
		if i+1 < len(decs) {
			end = decs[i+1].startLine
			// Look backwards from next decision startLine to trim preceding divider and heading
			for end > decs[i].fmEndLine+1 {
				tr := strings.TrimSpace(lines[end-1])
				if tr == "" || tr == "---" || strings.HasPrefix(tr, "#") {
					end--
				} else {
					break
				}
			}
		} else {
			end = len(lines)
		}
		blockStr := strings.Join(lines[start:end], "\n")
		blocks = append(blocks, []byte(blockStr))
	}
	return blocks
}

// ValidateDecisionRegistry checks a multi-decision registry file (e.g. DECISIONS.md)
// containing one or more decision records (anchored in HTML comments or unanchored with YAML frontmatters).
// It validates each decision block individually and aggregates issues with "DECISION D-NNN: " prefixes.
func ValidateDecisionRegistry(data []byte) *ValidationResult {
	result := &ValidationResult{Status: "valid"}
	if len(strings.TrimSpace(string(data))) == 0 {
		result.AddIssue(L2Block, "V-EMPTY", "Decision registry content is empty")
		result.Status = "draft"
		return result
	}

	anchoredRe := regexp.MustCompile(`(?s)<!--\s*DECISION:\s*([A-Za-z0-9_-]+)\s*-->\s*(.*?)\s*<!--\s*/DECISION:\s*([A-Za-z0-9_-]+)\s*-->`)
	matches := anchoredRe.FindAllSubmatch(data, -1)
	seenDecIDs := make(map[string]bool)

	validateSingleDec := func(id string, block []byte) {
		v := ValidateDecision(block)
		for _, issue := range v.Issues {
			// Suppress short body warning during initial plan saving (ADR skeletons in DECISIONS.md are concise)
			if issue.Code == "W-SHORT-DECISION" {
				continue
			}
			prefix := ""
			if id != "" {
				prefix = fmt.Sprintf("DECISION %s: ", id)
			}
			issueWithPrefix := issue
			issueWithPrefix.Prefix = prefix
			result.Issues = append(result.Issues, issueWithPrefix)
		}
		if v.HasBlocking() {
			result.Status = "draft"
		} else if v.WarningCount() > 0 && result.Status != "draft" {
			result.Status = "valid-with-warnings"
		}
	}

	for _, m := range matches {
		id := string(m[1])
		closeID := string(m[3])
		if !strings.EqualFold(id, closeID) {
			continue
		}
		if seenDecIDs[strings.ToUpper(id)] {
			continue
		}
		seenDecIDs[strings.ToUpper(id)] = true
		validateSingleDec(id, m[2])
	}

	for _, block := range ExtractUnanchoredDecisionBlocks(data) {
		id := ""
		if fm, _, err := ParseFrontmatter(block); err == nil && fm != nil {
			id = fm.GetString("id")
			if id == "" {
				id = fm.GetString("decision_id")
			}
		}
		if id == "" {
			// Extract rawID if ParseFrontmatter failed (e.g. malformed YAML syntax)
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
		if id != "" && !seenDecIDs[strings.ToUpper(id)] && !strings.HasPrefix(strings.ToUpper(id), "CHK-") {
			seenDecIDs[strings.ToUpper(id)] = true
			validateSingleDec(id, block)
		}
	}

	// Fallback for single standalone ADR (e.g. in decision scope or single D-*.md)
	if len(seenDecIDs) == 0 {
		v := ValidateDecision(data)
		for _, issue := range v.Issues {
			if issue.Code == "W-SHORT-DECISION" {
				continue
			}
			result.Issues = append(result.Issues, issue)
		}
		result.Status = v.Status
	}

	if len(result.Issues) == 0 {
		result.Status = "valid"
	}
	return result
}

// ValidateConflictResolution checks a conflict resolution record (ACH matrix analysis).
func ValidateConflictResolution(data []byte) *ValidationResult {
	result := &ValidationResult{Status: "valid"}

	fm, body, err := ParseFrontmatter(data)
	if err != nil {
		result.AddIssue(L2Block, "V-INVALID-FRONTMATTER", "YAML frontmatter is malformed: "+err.Error())
		result.Status = "invalid"
		return result
	}

	if fm == nil {
		result.AddIssueWithHint(L2Block, "V-MISSING-FRONTMATTER",
			"Conflict resolution record has no YAML frontmatter",
			"Add frontmatter with at least: id (or conflict_id), decision_id, title, status, door_type")
		result.Status = "draft"
		return result
	}

	// Required fields: id (or conflict_id), decision_id, title, status
	requiredFields := []string{"id", "decision_id", "title", "status"}
	for _, field := range requiredFields {
		has := fm.Has(field)
		if !has && field == "id" {
			has = fm.Has("conflict_id")
		}
		if !has {
			result.AddFieldIssueWithHint(L2Block, "V-MISSING-FIELD", field,
				fmt.Sprintf("Required field %q missing from conflict resolution", field),
				fmt.Sprintf("Add '%s: <value>' to the frontmatter", field))
		}
	}

	// L3: door_type should be present
	if !fm.Has("door_type") {
		result.AddIssueWithHint(L3Warn, "W-MISSING-DOOR-TYPE",
			"Conflict resolution lacks door_type classification",
			"Add 'door_type: one-way' or 'door_type: two-way'")
	}

	// L3: Body should not be trivially short
	bodyStr := strings.TrimSpace(string(body))
	if len(bodyStr) < 100 {
		result.AddIssueWithHint(L3Warn, "W-SHORT-BODY",
			"Conflict resolution body is very short — may lack sufficient context",
			"Include Conflict Summary, ACH Matrix, and Resolution sections")
	}

	// L3: Evidence grades and recalled grade cap checks
	checkEvidenceGrades(bodyStr, result)
	checkRecalledGradeCap(bodyStr, result)

	if result.HasBlocking() {
		result.Status = "draft"
	} else if result.WarningCount() > 0 {
		result.Status = "valid-with-warnings"
	}

	return result
}

// ValidateArtifact performs minimal structural validation suitable for
// plans, FADs, and other non-session/non-decision artifacts.
func ValidateArtifact(data []byte) *ValidationResult {
	result := &ValidationResult{Status: "valid"}

	if len(bytes.TrimSpace(data)) == 0 {
		result.AddIssue(L2Block, "V-EMPTY", "Content is empty")
		result.Status = "invalid"
		return result
	}

	// Check frontmatter
	fm, body, err := ParseFrontmatter(data)
	if err != nil || fm == nil {
		result.AddIssueWithHint(L3Warn, "W-MISSING-FRONTMATTER",
			"No YAML frontmatter found",
			"Add a --- delimited YAML block at the top of the document")
	}

	// Check body length
	bodyStr := strings.TrimSpace(string(body))
	if len(bodyStr) < 100 {
		result.AddIssueWithHint(L3Warn, "W-SHORT-BODY",
			"Body is very short (< 100 characters)",
			"Ensure the artifact contains substantive content")
	}

	// L3: Evidence grades and recalled grade cap checks
	checkEvidenceGrades(bodyStr, result)
	checkRecalledGradeCap(bodyStr, result)

	if result.HasBlocking() {
		result.Status = "draft"
	} else if result.WarningCount() > 0 {
		result.Status = "valid-with-warnings"
	}

	return result
}

// ValidateFAD validates a Founding Architecture Document (FAD).
// It checks for FAD frontmatter (id/session_id, title, synthesis_date/date, status),
// substantive body content (>= 100 chars), and inline evidence grades.
func ValidateFAD(data []byte) *ValidationResult {
	result := &ValidationResult{Status: "valid"}

	if len(bytes.TrimSpace(data)) == 0 {
		result.AddIssue(L2Block, "V-EMPTY", "FAD content is empty")
		result.Status = "invalid"
		return result
	}

	fm, body, err := ParseFrontmatter(data)
	if err != nil || fm == nil {
		result.AddIssueWithHint(L3Warn, "W-MISSING-FRONTMATTER",
			"No YAML frontmatter found",
			"Add a --- delimited YAML block at the top of the FAD")
	} else {
		// Check ID (id or session_id)
		if !fm.Has("id") && !fm.Has("session_id") {
			result.AddIssueWithHint(L3Warn, "W-MISSING-FIELD",
				"Missing 'id' or 'session_id' in frontmatter",
				"Add 'id: FAD-001' or 'session_id: FAD' to frontmatter")
		}
		// Check Title
		if !fm.Has("title") || strings.TrimSpace(fm.GetString("title")) == "" {
			result.AddIssueWithHint(L3Warn, "W-MISSING-FIELD",
				"Missing 'title' in frontmatter",
				"Add 'title: Founding Architecture Document' to frontmatter")
		}
		// Check Date (synthesis_date or date)
		if !fm.Has("synthesis_date") && !fm.Has("date") {
			result.AddIssueWithHint(L3Warn, "W-MISSING-FIELD",
				"Missing 'synthesis_date' (or 'date') in frontmatter",
				"Add 'synthesis_date: YYYY-MM-DD' to frontmatter")
		}
	}

	// Check body length
	bodyStr := strings.TrimSpace(string(body))
	if len(bodyStr) < 100 {
		result.AddIssueWithHint(L3Warn, "W-SHORT-BODY",
			"FAD body is very short (< 100 characters)",
			"Ensure the FAD synthesizes all session findings into an architectural blueprint")
	}

	// L3: Evidence grades and recalled grade cap checks
	checkEvidenceGrades(bodyStr, result)
	checkRecalledGradeCap(bodyStr, result)

	if result.HasBlocking() {
		result.Status = "draft"
	} else if result.WarningCount() > 0 {
		result.Status = "valid-with-warnings"
	}

	return result
}

// ValidateComparison checks a comparison session or WEP matrix output.
// It verifies frontmatter (id/session_id, title, date, status), substantive body content,
// Weighted Evaluation Protocol (WEP) matrix presence, and inline evidence grades.
func ValidateComparison(data []byte) *ValidationResult {
	result := &ValidationResult{Status: "valid"}

	if len(bytes.TrimSpace(data)) == 0 {
		result.AddIssue(L2Block, "V-EMPTY", "Comparison content is empty")
		result.Status = "invalid"
		return result
	}

	fm, body, err := ParseFrontmatter(data)
	if err != nil {
		result.AddIssue(L2Block, "V-INVALID-FRONTMATTER", "YAML frontmatter is malformed: "+err.Error())
		result.Status = "invalid"
		return result
	}

	if fm == nil {
		result.AddIssueWithHint(L2Block, "V-MISSING-FRONTMATTER",
			"Comparison record has no YAML frontmatter",
			"Add frontmatter with at least: id (or session_id), title, date, status")
		result.Status = "draft"
		return result
	}

	// L2: Required frontmatter fields
	if !fm.Has("id") && !fm.Has("session_id") {
		result.AddFieldIssueWithHint(L2Block, "V-MISSING-FIELD", "id",
			"Required frontmatter field 'id' (or 'session_id') is missing",
			"Add 'id: C-01-cmp-<slug>' to frontmatter")
	}
	if !fm.Has("title") || strings.TrimSpace(fm.GetString("title")) == "" {
		result.AddFieldIssueWithHint(L2Block, "V-MISSING-FIELD", "title",
			"Required frontmatter field 'title' is missing",
			"Add 'title: [Comparison Title]' to frontmatter")
	}
	if !fm.Has("date") {
		result.AddFieldIssueWithHint(L2Block, "V-MISSING-FIELD", "date",
			"Required frontmatter field 'date' is missing",
			"Add 'date: YYYY-MM-DD' to frontmatter")
	}

	// Check body length
	bodyStr := strings.TrimSpace(string(body))
	if len(bodyStr) < 100 {
		result.AddIssueWithHint(L3Warn, "W-SHORT-BODY",
			"Comparison body is very short (< 100 characters)",
			"Ensure the comparison includes research question, recommendation, and evaluation matrix")
	}

	// Check for Weighted Evaluation Protocol or criteria matrix table
	hasWEPTable := strings.Contains(bodyStr, "| Criteria |") ||
		strings.Contains(bodyStr, "| Criterion |") ||
		strings.Contains(bodyStr, "| Option |") ||
		strings.Contains(bodyStr, "Weighted Evaluation") ||
		strings.Contains(bodyStr, "Weighted Total")
	if !hasWEPTable {
		result.AddIssueWithHint(L3Warn, "W-MISSING-WEP-MATRIX",
			"Comparison artifact lacks a structured Weighted Evaluation Protocol (WEP) matrix table",
			"Include a criteria and score matrix per COMPARISON-SESSION.template.md")
	}

	// L3: Evidence grades and recalled grade cap checks
	checkEvidenceGrades(bodyStr, result)
	checkRecalledGradeCap(bodyStr, result)

	if result.HasBlocking() {
		result.Status = "draft"
	} else if result.WarningCount() > 0 {
		result.Status = "valid-with-warnings"
	}

	return result
}

// ValidatePlan validates a research plan or pipeline DAG.
// Unlike sessions and decisions, plans do not require YAML frontmatter or evidence grades.
func ValidatePlan(data []byte) *ValidationResult {
	result := &ValidationResult{Status: "valid"}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		result.AddIssue(L2Block, "V-EMPTY", "Plan content is empty")
		result.Status = "invalid"
		return result
	}

	// Try parsing as a pipeline DAG
	dag, err := ParsePipeline(data)
	if err == nil && len(dag.Sessions) > 0 {
		// Validate DAG structure (cycles, dangling dependencies)
		if valErr := dag.ValidateDAG(); valErr != nil {
			result.AddIssueWithHint(L2Block, "V-INVALID-DAG",
				fmt.Sprintf("Pipeline DAG validation failed: %v", valErr),
				"Fix dependency cycles or dangling session references")
		}
		// Warn if any session lacks prompt instructions
		for _, s := range dag.Sessions {
			if strings.TrimSpace(s.Prompt) == "" {
				result.AddIssueWithHint(L3Warn, "W-EMPTY-PROMPT",
					fmt.Sprintf("Session %s has no prompt block", s.ID),
					"Include an execution prompt block for each session")
			}
		}

		// Cross-check declared topology sessions against parsed sessions
		declared := ExtractDeclaredSessionIDs(string(data))
		parsedMap := make(map[string]bool)
		for _, s := range dag.Sessions {
			parsedMap[strings.ToUpper(s.ID)] = true
		}
		for _, declID := range declared {
			if IsSynthesisSession(declID) && (parsedMap["SYN-01"] || parsedMap["FAD"]) {
				continue
			}
			if !parsedMap[declID] {
				result.AddIssueWithHint(L3Warn, "W-PARSER-SESSION-MISMATCH",
					fmt.Sprintf("Session %s is declared in pipeline topology or DAG diagram, but lacks a parsed session entry with prompt block", declID),
					"Ensure every declared session has a '### [ID]: [Title]' header and a complete '```prompt' block")
			}
		}
	} else {
		// Decision-level, comparison, or generic plan without full DAG sessions
		if len(trimmed) < 100 {
			result.AddIssueWithHint(L3Warn, "W-SHORT-PLAN",
				"Plan is very short (< 100 characters)",
				"Ensure the plan includes clear research questions, scope, and objectives")
		}
	}

	if result.HasBlocking() {
		result.Status = "draft"
	} else if result.WarningCount() > 0 {
		result.Status = "valid-with-warnings"
	}

	return result
}


