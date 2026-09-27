package core

import (
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

// HasBlocking reports whether any L2+ issues were found.
func (r *ValidationResult) HasBlocking() bool {
	return r.MaxLevel() >= L2Block
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

// evidenceGradePattern matches inline evidence grades like "A (source)", "B (source)", etc.
var evidenceGradePattern = regexp.MustCompile(`\b[A-E]\s*\([^)]+\)`)

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
		if !fm.Has(field) {
			result.AddIssueWithHint(L2Block, "V-MISSING-FIELD",
				fmt.Sprintf("Required frontmatter field %q is missing", field),
				fmt.Sprintf("Add '%s: <value>' to the frontmatter block", field))
		}
	}

	// L2: Body must not be empty
	bodyStr := strings.TrimSpace(string(body))
	if bodyStr == "" {
		result.AddIssue(L2Block, "V-EMPTY-BODY", "Session body is empty")
	}

	// L3: Evidence grades should be present
	if bodyStr != "" && !evidenceGradePattern.MatchString(bodyStr) {
		result.AddIssueWithHint(L3Warn, "W-NO-EVIDENCE-GRADES",
			"No inline evidence grades found (expected A-E grades per P3)",
			"Add evidence grades like 'A (official docs)' or 'B (peer-reviewed study)' to claims")
	}

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
		if !fm.Has(field) {
			result.AddIssueWithHint(L2Block, "V-MISSING-FIELD",
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

	if result.HasBlocking() {
		result.Status = "draft"
	} else if result.WarningCount() > 0 {
		result.Status = "valid-with-warnings"
	}

	return result
}
