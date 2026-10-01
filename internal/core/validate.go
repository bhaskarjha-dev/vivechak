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

// evidenceGradePattern matches inline evidence grades like "A (source)", "[Grade A]", "(Grade B · ...)", "(A · corroborated · fresh | fetched)", "[E-01]", "E-001", etc.
var evidenceGradePattern = regexp.MustCompile(`(?:\[?[Gg]rade\s+[A-E][^\]\)\n]*\]?|\b[A-E]\s*\([^)]+\)|\([Gg]rade\s+[A-E][^)]*\)|\([A-E]\s*[·|][^)]*\)|\[[A-E]\s*[·|][^\]]*\]|\[E-\d+\]|\bE-\d+\b)`)

// recalledHighGradePattern matches Grade A or B claims that rely on recalled/parametric memory.
// Per Principle P3 (Evidentiary Grounding), unverified recall must be capped at Grade D.
var recalledHighGradePattern = regexp.MustCompile(`(?i)(?:\[(?:Grade\s+)?[AB]\s*[·|:,][^\]\n]*\b(?:recalled|memory)\b[^\]\n]*\]|\((?:Grade\s+)?[AB]\s*[·|:,][^)\n]*\b(?:recalled|memory)\b[^)\n]*\)|\b(?:Grade\s+)?[AB]\s*\([^)\n]*\b(?:recalled|memory)\b[^)\n]*\)|\b(?:Grade\s+)?[AB]\s*\[[^\]\n]*\b(?:recalled|memory)\b[^\]\n]*\]|\[(?:Grade\s+)[AB][^\]\n]*\b(?:recalled|memory)\b[^\]\n]*\])`)

// ValidateSession checks a research session output against the validation ladder.
// Returns issues at levels L1-L3 (L4 is project-wide, not per-session).
func ValidateSession(data []byte) *ValidationResult {
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

// checkEvidenceGrades verifies that inline evidence grades (A-E) are present in the body.
func checkEvidenceGrades(bodyStr string, result *ValidationResult) {
	if bodyStr != "" && !evidenceGradePattern.MatchString(bodyStr) {
		result.AddIssueWithHint(L3Warn, "W-NO-EVIDENCE-GRADES",
			"No inline evidence grades found (expected A-E grades per P3)",
			"Add evidence grades like 'A (official docs)' or 'B (peer-reviewed study)' to claims")
	}
}

// checkRecalledGradeCap verifies that recalled knowledge claims are capped at Grade D per Principle P3.
func checkRecalledGradeCap(bodyStr string, result *ValidationResult) {
	if bodyStr != "" && recalledHighGradePattern.MatchString(bodyStr) {
		result.AddIssueWithHint(L3Warn, "W-RECALLED-GRADE-CAP",
			"Recalled knowledge must be capped at Grade D per Principle P3",
			"Downgrade recalled claims to Grade D or corroborate them with live fetched/cached sources")
	}
}

// QualityObservation is an advisory finding from content analysis.
type QualityObservation struct {
	Category string `json:"category"` // structure, evidence, coherence, freshness
	Message  string `json:"message"`
	Severity string `json:"severity"` // info, suggestion, consideration
}

// staleDatePattern matches year references that may be outdated (pre-2024).
var staleDatePattern = regexp.MustCompile(`\b20(?:1[0-9]|2[0-3])\b`)

// ObserveSessionQuality performs advisory content analysis and returns observations.
// These are informational suggestions, not validation errors. They help the author
// strengthen their output but never block saving.
func ObserveSessionQuality(content []byte) []QualityObservation {
	var observations []QualityObservation
	bodyStr := string(content)

	// Count inline evidence grade markers
	gradeMatches := evidenceGradePattern.FindAllString(bodyStr, -1)
	gradeCount := len(gradeMatches)
	if gradeCount == 0 {
		observations = append(observations, QualityObservation{
			Category: "evidence",
			Message:  "No inline evidence grades found. Consider adding grades (A-E) to significant claims.",
			Severity: "suggestion",
		})
	} else if gradeCount < 3 {
		observations = append(observations, QualityObservation{
			Category: "evidence",
			Message:  fmt.Sprintf("Found %d evidence grade(s). Key findings typically benefit from more graded claims.", gradeCount),
			Severity: "info",
		})
	}

	// Check for Prior/Delta sections
	hasPrior := strings.Contains(bodyStr, "## Prior") || strings.Contains(bodyStr, "## Prior (")
	hasDelta := strings.Contains(bodyStr, "## Delta") || strings.Contains(bodyStr, "## Delta (")
	if !hasPrior {
		observations = append(observations, QualityObservation{
			Category: "structure",
			Message:  "No Prior section found. Consider stating pre-research beliefs to make the Delta visible.",
			Severity: "suggestion",
		})
	}
	if !hasDelta {
		observations = append(observations, QualityObservation{
			Category: "structure",
			Message:  "No Delta section found. Consider adding a table showing what research confirmed, updated, or contradicted.",
			Severity: "suggestion",
		})
	}

	// Check for stale date references
	staleMatches := staleDatePattern.FindAllString(bodyStr, -1)
	if len(staleMatches) > 0 {
		// Deduplicate
		seen := make(map[string]bool)
		var unique []string
		for _, m := range staleMatches {
			if !seen[m] {
				seen[m] = true
				unique = append(unique, m)
			}
		}
		observations = append(observations, QualityObservation{
			Category: "freshness",
			Message:  fmt.Sprintf("Found references to potentially stale dates: %s. Verify these are still current.", strings.Join(unique, ", ")),
			Severity: "consideration",
		})
	}

	return observations
}

