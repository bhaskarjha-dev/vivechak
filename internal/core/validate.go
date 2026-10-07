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

// EvidenceGradePattern matches inline evidence grades like "A (source)", "[Grade A]", "(Grade B · ...)", "(A · corroborated · fresh | fetched)", "[E-01]", "E-001", etc.
var EvidenceGradePattern = regexp.MustCompile(`(?:\[?[Gg]rade\s+[A-E][^\]\)\n]*\]?|\b[A-E]\s*\([^)]+\)|\([Gg]rade\s+[A-E][^)]*\)|\([A-E]\s*[·|][^)]*\)|\[[A-E]\s*[·|][^\]]*\]|\[E-\d+\]|\bE-\d+\b)`)

var evidenceGradePattern = EvidenceGradePattern

// recalledHighGradePattern matches Grade A, B, or C claims that rely on recalled/parametric memory.
// Per Principle P3 (Evidentiary Grounding), unverified recall must be capped at Grade D.
var recalledHighGradePattern = regexp.MustCompile(`(?i)(?:\[(?:Grade\s+)?[ABC]\s*[·|:,][^\]\n]*\b(?:recalled|memory|parametric\s+memory|model\s+memory|ai\s+memory)\b[^\]\n]*\]|\((?:Grade\s+)?[ABC]\s*[·|:,][^)\n]*\b(?:recalled|memory|parametric\s+memory|model\s+memory|ai\s+memory)\b[^)\n]*\)|\b(?:Grade\s+)?[ABC]\s*\([^)\n]*\b(?:recalled|memory|parametric\s+memory|model\s+memory|ai\s+memory)\b[^)\n]*\)|\b(?:Grade\s+)?[ABC]\s*\[[^\]\n]*\b(?:recalled|memory|parametric\s+memory|model\s+memory|ai\s+memory)\b[^\]\n]*\]|\[(?:Grade\s+)[ABC][^\]\n]*\b(?:recalled|memory|parametric\s+memory|model\s+memory|ai\s+memory)\b[^\]\n]*\])`)

var (
	qualifiedSourceRe = regexp.MustCompile(
		`(?i)(https?://\S+|` + // Canonical URLs
			`[a-z0-9][-a-z0-9]*\.(?:org|com|io|dev|gov|edu|net|ai)\b|` + // Domain anchors
			`RFC\s*\d+|ISO\s*\d+|` + // Standard specifications
			`(?:github|gitlab)\.com/\S+)`, // Repository anchors
	)

	vagueSourceRe = regexp.MustCompile(
		`(?i)^(?:official\s*docs?|documentation|online|web|search|` +
			`the\s*internet|various\s*sources|vendor\s*(?:docs?|website)|` +
			`primary\s*sources?|community\s*(?:reports?|benchmarks?))$`,
	)
)

// ValidateEvidenceProvenance verifies that Grade A fetched claims have qualified provenance anchors.
func ValidateEvidenceProvenance(bodyStr string, isOneWay bool, result *ValidationResult) {
	lines := strings.Split(bodyStr, "\n")
	inSourcesTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## Sources") || strings.HasPrefix(trimmed, "## Evidence") {
			inSourcesTable = true
			continue
		}
		if inSourcesTable && strings.HasPrefix(trimmed, "## ") {
			inSourcesTable = false
		}

		if inSourcesTable && strings.HasPrefix(trimmed, "|") && !strings.Contains(trimmed, "---") {
			parts := strings.Split(trimmed, "|")
			if len(parts) >= 4 {
				isGradeA := false
				isFetched := false
				var sourceText string

				for idx, part := range parts {
					cell := strings.TrimSpace(part)
					cellUpper := strings.ToUpper(cell)
					if cellUpper == "A" || cellUpper == "GRADE A" || cellUpper == "[A]" || cellUpper == "[GRADE A]" {
						isGradeA = true
					}
					if strings.Contains(strings.ToLower(cell), "fetched") {
						isFetched = true
					}
					if idx == 2 && cell != "" && cellUpper != "SOURCE" && cellUpper != "A" && cellUpper != "GRADE A" {
						sourceText = cell
					} else if idx == 1 && cell != "" && cellUpper != "#" && cellUpper != "SOURCE" && cellUpper != "A" && cellUpper != "GRADE A" && sourceText == "" {
						sourceText = cell
					}
				}

				if isGradeA && (isFetched || strings.Contains(strings.ToLower(trimmed), "fetched")) {
					cleanSource := strings.Trim(strings.TrimSpace(sourceText), "[]()\"'`")
					isQualified := qualifiedSourceRe.MatchString(cleanSource)
					isVague := vagueSourceRe.MatchString(cleanSource) || cleanSource == "" || (!isQualified && len(cleanSource) < 15)

					if isVague {
						if isOneWay {
							result.AddIssueWithHint(L2Block, "V-OWD-VAGUE-PROVENANCE",
								fmt.Sprintf("One-way door session has Grade A claim with vague provenance: %q", cleanSource),
								"Provide a canonical URL, domain anchor (e.g. wails.io), RFC spec, or specific document title")
						} else {
							result.AddIssueWithHint(L3Warn, "W-VAGUE-PROVENANCE",
								fmt.Sprintf("Grade A claim has vague provenance: %q", cleanSource),
								"Provide a canonical URL, domain anchor, RFC spec, or specific document title")
						}
					}
				}
			}
		}
	}
}

// ValidateDiscoveredConcerns verifies that Discovered Concerns contains substantive analysis on one-way door sessions.
func ValidateDiscoveredConcerns(bodyStr string, isOneWay bool, result *ValidationResult) {
	hasHeading := hasSectionHeading(bodyStr, "discovered concerns", "discovered concern")
	content := extractSection(bodyStr, "discovered concerns", "discovered concern")
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
	if bodyStr == "" {
		return
	}
	matches := recalledHighGradePattern.FindAllString(bodyStr, -1)
	for _, match := range matches {
		matchLower := strings.ToLower(match)
		// If the citation cites documentation, URLs, RFCs, or benchmarks, technical computing memory (RAM, in-memory) must not be flagged
		if strings.Contains(matchLower, "official") ||
			strings.Contains(matchLower, "doc") ||
			strings.Contains(matchLower, "http") ||
			strings.Contains(matchLower, "rfc") ||
			strings.Contains(matchLower, "benchmark") ||
			strings.Contains(matchLower, "in-memory") ||
			strings.Contains(matchLower, "shared memory") ||
			strings.Contains(matchLower, "memory safety") {
			// Check if it explicitly indicates unverified recall
			if !strings.Contains(matchLower, "recalled") {
				continue
			}
		}
		result.AddIssueWithHint(L3Warn, "W-RECALLED-GRADE-CAP",
			"Recalled knowledge must be capped at Grade D per Principle P3",
			"Downgrade recalled claims to Grade D or corroborate them with live fetched/cached sources")
		break
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

	// Grade distribution analysis
	gradeACount := 0
	gradeBCount := 0
	gradeCPlusCount := 0
	for _, g := range gradeMatches {
		upper := strings.ToUpper(g)
		if strings.Contains(upper, "GRADE A") || strings.Contains(upper, "[A]") || strings.Contains(upper, "(A ") {
			gradeACount++
		} else if strings.Contains(upper, "GRADE B") || strings.Contains(upper, "[B]") || strings.Contains(upper, "(B ") {
			gradeBCount++
		} else {
			gradeCPlusCount++
		}
	}
	if gradeCount >= 4 && gradeACount > int(float64(gradeCount)*0.7) {
		observations = append(observations, QualityObservation{
			Category: "evidence",
			Message: fmt.Sprintf("Grade distribution is %d A / %d B / %d C+ (%d%% Grade A). "+
				"Web-based research typically produces ~25%% A, ~50%% B, ~25%% C. "+
				"Verify Grade A citations include URLs to primary sources.",
				gradeACount, gradeBCount, gradeCPlusCount,
				gradeACount*100/gradeCount),
			Severity: "consideration",
		})
	}

	// Discovered Concerns check
	hasConcerns := hasSectionHeading(bodyStr, "discovered concerns", "discovered concern")
	if !hasConcerns {
		observations = append(observations, QualityObservation{
			Category: "concerns",
			Message: "No Discovered Concerns section found. Every research session should " +
				"uncover at least one unexpected finding beyond the stated scope.",
			Severity: "suggestion",
		})
	}

	// Key Findings count check
	findingsSection := extractSection(bodyStr, "key findings")
	if findingsSection != "" {
		bulletCount := strings.Count(findingsSection, "\n- ") + strings.Count(findingsSection, "\n* ")
		trimmedFindings := strings.TrimSpace(findingsSection)
		if len(trimmedFindings) > 0 && (trimmedFindings[0] == '-' || trimmedFindings[0] == '*') {
			bulletCount++
		}
		if bulletCount > 0 && bulletCount < 3 {
			observations = append(observations, QualityObservation{
				Category: "depth",
				Message: fmt.Sprintf("Only %d key finding(s). Sessions typically benefit "+
					"from 4-7 graded findings for adequate decision support.", bulletCount),
				Severity: "suggestion",
			})
		}
	}

	// Grade A URL check: Grade A "fetched" citations require URLs
	evidenceLedger := extractSection(bodyStr, "evidence")
	if evidenceLedger == "" {
		evidenceLedger = extractSection(bodyStr, "sources")
	}
	if evidenceLedger != "" {
		gradeAFetched := regexp.MustCompile(`(?i)grade\s*a.*?\bfetched\b`)
		urlPattern := regexp.MustCompile(`https?://`)
		rows := strings.Split(evidenceLedger, "\n")
		gradeANoURL := 0
		for _, row := range rows {
			if gradeAFetched.MatchString(row) && !urlPattern.MatchString(row) {
				gradeANoURL++
			}
		}
		if gradeANoURL > 0 {
			observations = append(observations, QualityObservation{
				Category: "evidence",
				Message: fmt.Sprintf("%d Grade A 'fetched' citation(s) lack URLs. "+
					"Grade A requires verifiable access to primary sources — "+
					"include the source URL or downgrade to Grade B.", gradeANoURL),
				Severity: "consideration",
			})
		}
	}

	// Check for Prior/Delta sections
	hasPrior := hasSectionHeading(bodyStr, "prior")
	hasDelta := hasSectionHeading(bodyStr, "delta")
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
	} else {
		// Confirmation bias check: does Delta challenge any beliefs?
		deltaSection := extractSection(bodyStr, "delta")
		lowerDelta := strings.ToLower(deltaSection)
		hasContradiction := strings.Contains(lowerDelta, "contradict") ||
			strings.Contains(lowerDelta, "updated") ||
			strings.Contains(lowerDelta, "revised") ||
			strings.Contains(lowerDelta, "overturned")
		if !hasContradiction {
			observations = append(observations, QualityObservation{
				Category: "bias",
				Message:  "Delta section has no contradicted or updated beliefs. If every prior was confirmed, consider whether the research genuinely challenged initial assumptions (P8).",
				Severity: "consideration",
			})
		}
	}

	// Recommendation alternatives check
	recSection := extractSection(bodyStr, "recommend")
	if recSection != "" {
		lowerRec := strings.ToLower(recSection)
		hasAlternatives := strings.Contains(lowerRec, "alternative") ||
			strings.Contains(lowerRec, "rejected") ||
			strings.Contains(lowerRec, "instead of") ||
			strings.Contains(lowerRec, "over ")
		if !hasAlternatives {
			observations = append(observations, QualityObservation{
				Category: "rigor",
				Message:  "Recommendation doesn't reference rejected alternatives. Stating why alternatives were rejected strengthens the decision rationale.",
				Severity: "suggestion",
			})
		}
	}

	// Check for stale date references (refined to avoid false positives on ports, latencies, QPS)
	staleIndices := staleDatePattern.FindAllStringIndex(bodyStr, -1)
	if len(staleIndices) > 0 {
		seen := make(map[string]bool)
		var unique []string
		for _, idx := range staleIndices {
			start, end := idx[0], idx[1]
			if isLikelyYearReference(bodyStr, start, end) {
				m := bodyStr[start:end]
				if !seen[m] {
					seen[m] = true
					unique = append(unique, m)
				}
			}
		}
		if len(unique) > 0 {
			observations = append(observations, QualityObservation{
				Category: "freshness",
				Message:  fmt.Sprintf("Found references to potentially stale dates: %s. Verify these are still current.", strings.Join(unique, ", ")),
				Severity: "consideration",
			})
		}
	}

	return observations
}

// isLikelyYearReference distinguishes actual year references from ports (:2020),
// latencies (2015 ms), throughput (2022 req/s), versions (v2020), and percentages.
func isLikelyYearReference(text string, start, end int) bool {
	if start > 0 {
		prev := text[start-1]
		if prev == ':' || prev == 'v' || prev == 'V' || prev == '#' || prev == '$' || prev == '@' {
			return false
		}
	}
	if end < len(text) {
		next := text[end]
		if next == '%' || next == 'x' || next == 'X' {
			return false
		}
	}
	after := text[end:]
	if len(after) > 0 && (after[0] == ' ' || after[0] == '\t') {
		after = strings.TrimLeft(after, " \t")
	}
	afterLower := strings.ToLower(after)
	unitPrefixes := []string{"ms", "µs", "us", "ns", "qps", "req", "rps", "rpm", "fps", "hz", "khz", "mhz", "ghz", "kb", "mb", "gb", "tb", "bytes", "ops/s"}
	for _, u := range unitPrefixes {
		if strings.HasPrefix(afterLower, u) {
			rem := afterLower[len(u):]
			isAlnum := len(rem) > 0 && ((rem[0] >= 'a' && rem[0] <= 'z') || (rem[0] >= '0' && rem[0] <= '9'))
			if len(rem) == 0 || !isAlnum {
				return false
			}
		}
	}
	return true
}

// hasSectionHeading checks whether body contains any markdown heading (# ...) containing any of the keywords.
func hasSectionHeading(body string, keywords ...string) bool {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			for _, kw := range keywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
		}
	}
	return false
}


