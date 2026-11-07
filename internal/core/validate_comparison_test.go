package core

import (
	"strings"
	"testing"
)

// TestValidateComparison_TableDriven evaluates the Harness's 10 broken comparison
// patterns and 1 fully valid comparison, ensuring each rule reports its expected code.
func TestValidateComparison_TableDriven(t *testing.T) {
	// Baseline valid comparison document
	validDoc := `---
id: C-001
type: comparison
title: Pino vs Winston Logger Evaluation
date: 2026-11-04
status: complete
vivechak: 1
---
# Pino vs Winston Logger Evaluation

## Key Findings
- Pino provides superior throughput with extreme low memory overhead [Grade A: https://github.com/pinojs/pino | fetched].
- Winston provides an extensive transport ecosystem with pluggable formatters [Grade B: https://github.com/winstonjs/winston | fetched].

## Weighted Evaluation Matrix
| Criterion | Weight | Pino Score (1-5) | Winston Score (1-5) |
|---|---|---|---|
| Throughput & Latency | 0.40 | 5.0 | 3.0 |
| Ecosystem & Transports | 0.30 | 3.5 | 5.0 |
| Memory Footprint | 0.30 | 4.8 | 3.2 |
| Total Score | 1.00 | 4.49 | 3.66 |

## Recommendation
We recommend Pino based on its asynchronous sonic-boom pipeline and low allocation overhead [Grade A: https://github.com/pinojs/pino | fetched].

## Discovered Concerns & Tradeoffs
- Pino serializes formatting off-process via transports, which requires process isolation considerations under AWS Lambda.

## Sources & Evidence Ledger
- Pino Repository: https://github.com/pinojs/pino [Grade A | fetched]
- Winston Repository: https://github.com/winstonjs/winston [Grade B | fetched]
`

	tests := []struct {
		name        string
		content     string
		wantValid   bool
		wantCode    string
		expectIssue bool
	}{
		{
			name:        "1_ValidComparison",
			content:     validDoc,
			wantValid:   true,
			expectIssue: false,
		},
		{
			name:        "2_Broken_EmptyContent",
			content:     "   \n\t  ",
			wantValid:   false,
			wantCode:    "V-EMPTY",
			expectIssue: true,
		},
		{
			name:        "3_Broken_TooShort",
			content:     "---\ntitle: Short\n---\n# Short\nLess than 150 chars.",
			wantValid:   false,
			wantCode:    "V-TOO-SHORT",
			expectIssue: true,
		},
		{
			name: "4_Broken_UnfilledSlotPlaceholder",
			content: strings.Replace(validDoc,
				"Pino vs Winston Logger Evaluation",
				"{option} vs Winston Logger Evaluation", 1),
			wantValid:   false,
			wantCode:    "V-PLACEHOLDER",
			expectIssue: true,
		},
		{
			name: "5_Broken_BracketPlaceholderLine",
			content: strings.Replace(validDoc,
				"## Discovered Concerns & Tradeoffs",
				"## Discovered Concerns & Tradeoffs\n[Document substantive discovered concerns here]", 1),
			wantValid:   false,
			wantCode:    "V-PLACEHOLDER",
			expectIssue: true,
		},
		{
			name:        "6_Broken_MissingFrontmatter",
			content:     strings.TrimPrefix(validDoc, "---\nid: C-001\ntype: comparison\ntitle: Pino vs Winston Logger Evaluation\ndate: 2026-11-04\nstatus: complete\nvivechak: 1\n---\n"),
			wantValid:   false,
			wantCode:    "V-MISSING-FRONTMATTER",
			expectIssue: true,
		},
		{
			name: "7_Broken_MissingRecommendationSection",
			content: strings.Replace(validDoc,
				"## Recommendation",
				"## Other Section", 1),
			wantValid:   false,
			wantCode:    "V-MISSING-SECTION",
			expectIssue: true,
		},
		{
			name: "8_Broken_NoGradedClaims",
			content: strings.ReplaceAll(
				strings.ReplaceAll(validDoc, "Grade A", "Tier 1"),
				"Grade B", "Tier 2"),
			wantValid:   false,
			wantCode:    "V-NO-GRADED-CLAIM",
			expectIssue: true,
		},
		{
			name: "9_Broken_NoMatrixTable",
			content: strings.Replace(validDoc,
				"| Criterion | Weight | Pino Score (1-5) | Winston Score (1-5) |\n|---|---|---|---|\n| Throughput & Latency | 0.40 | 5.0 | 3.0 |\n| Ecosystem & Transports | 0.30 | 3.5 | 5.0 |\n| Memory Footprint | 0.30 | 4.8 | 3.2 |\n| Total Score | 1.00 | 4.49 | 3.66 |",
				"No markdown matrix table here, just prose discussion of options.", 1),
			wantValid:   false,
			wantCode:    "V-CMP-NO-MATRIX",
			expectIssue: true,
		},
		{
			name: "10_Broken_TooFewOptions",
			content: strings.Replace(validDoc,
				"| Criterion | Weight | Pino Score (1-5) | Winston Score (1-5) |",
				"| Criterion | Weight | Pino Score (1-5) |", 1),
			wantValid:   false,
			wantCode:    "V-CMP-OPTIONS",
			expectIssue: true,
		},
		{
			name: "11_Broken_InvalidScores",
			content: strings.Replace(validDoc,
				"| Throughput & Latency | 0.40 | 5.0 | 3.0 |",
				"| Throughput & Latency | 0.40 | 15.0 | N/A |", 1),
			wantValid:   false,
			wantCode:    "V-CMP-SCORES",
			expectIssue: true,
		},
		{
			name: "12_Broken_WeightsDoNotSumToOne",
			content: strings.Replace(validDoc,
				"| Throughput & Latency | 0.40 | 5.0 | 3.0 |\n| Ecosystem & Transports | 0.30 | 3.5 | 5.0 |\n| Memory Footprint | 0.30 | 4.8 | 3.2 |",
				"| Throughput & Latency | 0.20 | 5.0 | 3.0 |\n| Ecosystem & Transports | 0.20 | 3.5 | 5.0 |\n| Memory Footprint | 0.20 | 4.8 | 3.2 |", 1),
			wantValid:   false,
			wantCode:    "V-CMP-WEIGHTS",
			expectIssue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ValidateComparison([]byte(tt.content))

			if res.Valid != tt.wantValid {
				t.Errorf("got Valid=%v, want %v (Status=%s, Issues=%v)", res.Valid, tt.wantValid, res.Status, res.Issues)
			}

			if tt.expectIssue {
				if !res.HasBlocking() {
					t.Errorf("expected blocking issues for %s", tt.name)
				}
				if tt.wantCode != "" {
					foundCode := false
					for _, iss := range res.Issues {
						if iss.Code == tt.wantCode {
							foundCode = true
							break
						}
					}
					if !foundCode {
						t.Errorf("expected issue code %q, got issues: %+v", tt.wantCode, res.Issues)
					}
				}
			} else {
				if res.HasBlocking() {
					t.Errorf("expected no blocking issues for valid comparison, got: %+v", res.Issues)
				}
				if !res.HasWEPMatrix {
					t.Error("expected HasWEPMatrix to be true for valid comparison")
				}
			}
		})
	}
}
