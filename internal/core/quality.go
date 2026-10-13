package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// QualityObservation is an advisory finding from content analysis.
type QualityObservation struct {
	Category string `json:"category"` // structure, evidence, coherence, freshness
	Message  string `json:"message"`
	Severity string `json:"severity"` // info, suggestion, consideration
}

// buildStaleDatePattern constructs a regex matching 4-digit years from 2010 up to cutoffYear - 1.
func buildStaleDatePattern(cutoffYear int) *regexp.Regexp {
	var years []string
	for y := 2010; y < cutoffYear; y++ {
		years = append(years, strconv.Itoa(y))
	}
	if len(years) == 0 {
		return regexp.MustCompile(`[^\s\S]`) // matches nothing
	}
	return regexp.MustCompile(`\b(?:` + strings.Join(years, "|") + `)\b`)
}

// staleDatePattern matches year references that may be outdated (more than 2 years old).
var staleDatePattern = func() *regexp.Regexp {
	cutoff := time.Now().Year() - 2
	return buildStaleDatePattern(cutoff)
}()

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
	// Verify Grade A claims have explicit URL or canonical RFC anchors (Task 4.2 / ADR D-UX-02)
	gradeALines := findLinesWithGradeA(bodyStr)
	unanchoredGradeA := 0
	for _, l := range gradeALines {
		lower := strings.ToLower(l)
		if !strings.Contains(lower, "http://") &&
			!strings.Contains(lower, "https://") &&
			!strings.Contains(lower, "rfc ") &&
			!strings.Contains(lower, "rfc-") &&
			!strings.Contains(lower, "doi:") {
			unanchoredGradeA++
		}
	}
	if unanchoredGradeA > 0 {
		observations = append(observations, QualityObservation{
			Category: "evidence",
			Message: fmt.Sprintf("%d Grade A claim(s) lack canonical URL or RFC provenance anchors. Primary evidence requires verifiable sources.", unanchoredGradeA),
			Severity: "consideration",
		})
	}

	// Citation diversity check for substantive research sessions (Task 4.5)
	if len(bodyStr) > 1000 {
		urlCount := countDistinctURLs(bodyStr)
		if urlCount < 2 {
			observations = append(observations, QualityObservation{
				Category: "evidence",
				Message:  "Session cites fewer than 2 distinct external sources. Citing multiple independent primary/secondary sources strengthens evidentiary grounding (Principle P3).",
				Severity: "consideration",
			})
		}
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
	findingsSection := ExtractSection(bodyStr, "key findings")
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
	evidenceLedger := ExtractSection(bodyStr, "evidence")
	if evidenceLedger == "" {
		evidenceLedger = ExtractSection(bodyStr, "sources")
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
		deltaSection := ExtractSection(bodyStr, "delta")
		lowerDelta := strings.ToLower(deltaSection)
		hasContradiction := strings.Contains(lowerDelta, "contradict") ||
			strings.Contains(lowerDelta, "revised") ||
			strings.Contains(lowerDelta, "overturned") ||
			strings.Contains(lowerDelta, "falsified") ||
			strings.Contains(lowerDelta, "disproved") ||
			(strings.Contains(lowerDelta, "updated") && !strings.Contains(lowerDelta, "unknown"))
		if !hasContradiction {
			observations = append(observations, QualityObservation{
				Category: "bias",
				Message:  "Delta section has no contradicted or updated beliefs. If every prior was confirmed or started from 'Unknown', consider whether the research genuinely challenged initial assumptions (P8).",
				Severity: "consideration",
			})
		}
	}

	// Recommendation & alternatives rigor check
	recSection := ExtractSection(bodyStr, "recommend", "verdict")
	altSection := ExtractSection(bodyStr, "alternative", "evaluated option", "rejected option")

	combinedRationale := strings.ToLower(recSection + "\n" + altSection)
	hasAlternatives := strings.Contains(combinedRationale, "alternative") ||
		strings.Contains(combinedRationale, "rejected") ||
		strings.Contains(combinedRationale, "instead of") ||
		strings.Contains(combinedRationale, "over ") ||
		len(strings.TrimSpace(altSection)) >= 30

	if recSection != "" && !hasAlternatives {
		observations = append(observations, QualityObservation{
			Category: "rigor",
			Message:  "Recommendation doesn't reference rejected alternatives. Stating why alternatives were rejected strengthens the decision rationale.",
			Severity: "suggestion",
		})
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

var gradeALineRe = regexp.MustCompile(`(?i)(?:\[(?:Grade\s+)?A[^\]\n]*\]|\((?:Grade\s+)?A[^)\n]*\)|\bGrade\s+A\b|\bGrade:A\b)`)

func findLinesWithGradeA(text string) []string {
	lines := strings.Split(text, "\n")
	var res []string
	for _, l := range lines {
		if gradeALineRe.MatchString(l) {
			res = append(res, l)
		}
	}
	return res
}

func countDistinctURLs(text string) int {
	matches := urlPattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return 0
	}
	seen := make(map[string]bool)
	for _, m := range matches {
		seen[strings.ToLower(m)] = true
	}
	return len(seen)
}
