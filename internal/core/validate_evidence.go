package core

import (
	"fmt"
	"regexp"
	"strings"
)

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
