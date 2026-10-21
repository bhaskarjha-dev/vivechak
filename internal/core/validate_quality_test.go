package core

import (
	"strings"
	"testing"
)

// TestObserveSessionQuality_NoGrades verifies observation when no evidence grades present.
func TestObserveSessionQuality_NoGrades(t *testing.T) {
	content := []byte("This is a session with no evidence grades at all. Just plain text.")
	obs := ObserveSessionQuality(content)

	found := false
	for _, o := range obs {
		if o.Category == "evidence" && strings.Contains(o.Message, "No inline evidence grades") {
			found = true
			if o.Severity != "suggestion" {
				t.Errorf("expected severity 'suggestion', got %q", o.Severity)
			}
		}
	}
	if !found {
		t.Error("expected evidence observation for content without grades")
	}
}

// TestObserveSessionQuality_FewGrades verifies observation when only 1-2 grades present.
func TestObserveSessionQuality_FewGrades(t *testing.T) {
	content := []byte("PostgreSQL is recommended. Grade A (official docs). Another claim here.")
	obs := ObserveSessionQuality(content)

	found := false
	for _, o := range obs {
		if o.Category == "evidence" && strings.Contains(o.Message, "1 evidence grade") {
			found = true
			if o.Severity != "info" {
				t.Errorf("expected severity 'info', got %q", o.Severity)
			}
		}
	}
	if !found {
		t.Error("expected info observation for content with few grades")
	}
}

// TestObserveSessionQuality_ManyGrades verifies no evidence observation when 3+ grades present with proper anchors.
func TestObserveSessionQuality_ManyGrades(t *testing.T) {
	content := []byte(`
Grade A (official docs: https://docs.example.com) for stability claim.
Grade B (benchmarks) for performance claim.
Grade C (vendor claims) for scalability.
Grade A (RFC 7636) for security.
`)
	obs := ObserveSessionQuality(content)

	for _, o := range obs {
		if o.Category == "evidence" {
			t.Errorf("unexpected evidence observation when 4 grades present: %v", o)
		}
	}
}

// TestObserveSessionQuality_NoPrior verifies observation when Prior section is missing.
func TestObserveSessionQuality_NoPrior(t *testing.T) {
	content := []byte("# Research\nSome findings here. Grade A (docs)")
	obs := ObserveSessionQuality(content)

	found := false
	for _, o := range obs {
		if o.Category == "structure" && strings.Contains(o.Message, "No Prior section") {
			found = true
			if o.Severity != "suggestion" {
				t.Errorf("expected severity 'suggestion', got %q", o.Severity)
			}
		}
	}
	if !found {
		t.Error("expected structure observation for missing Prior section")
	}
}

// TestObserveSessionQuality_NoDelta verifies observation when Delta section is missing.
func TestObserveSessionQuality_NoDelta(t *testing.T) {
	content := []byte("## Prior\n- I believe X because Y\n\n# Research\nSome findings. Grade A (docs)")
	obs := ObserveSessionQuality(content)

	foundDelta := false
	foundPrior := false
	for _, o := range obs {
		if o.Category == "structure" && strings.Contains(o.Message, "No Delta section") {
			foundDelta = true
		}
		if o.Category == "structure" && strings.Contains(o.Message, "No Prior section") {
			foundPrior = true
		}
	}
	if !foundDelta {
		t.Error("expected observation for missing Delta section")
	}
	if foundPrior {
		t.Error("should NOT observe missing Prior when it's present")
	}
}

// TestObserveSessionQuality_HasBothPriorAndDelta verifies no structure observations when both present.
func TestObserveSessionQuality_HasBothPriorAndDelta(t *testing.T) {
	content := []byte(`
## Prior
- I believe PostgreSQL is best because it handles JSON well

## Delta
| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| PostgreSQL handles JSON | Confirmed | Grade A (docs) | high |
`)
	obs := ObserveSessionQuality(content)

	for _, o := range obs {
		if o.Category == "structure" {
			t.Errorf("unexpected structure observation when Prior and Delta present: %v", o)
		}
	}
}

// TestObserveSessionQuality_StaleDates verifies observation for potentially stale date references.
func TestObserveSessionQuality_StaleDates(t *testing.T) {
	content := []byte("This library was released in 2019 and last updated in 2022. Grade A (docs)")
	obs := ObserveSessionQuality(content)

	found := false
	for _, o := range obs {
		if o.Category == "freshness" {
			found = true
			if o.Severity != "consideration" {
				t.Errorf("expected severity 'consideration', got %q", o.Severity)
			}
			if !strings.Contains(o.Message, "2019") || !strings.Contains(o.Message, "2022") {
				t.Errorf("expected stale dates mentioned, got: %s", o.Message)
			}
		}
	}
	if !found {
		t.Error("expected freshness observation for pre-2024 date references")
	}
}

// TestObserveSessionQuality_CurrentDates verifies no freshness observation for current dates.
func TestObserveSessionQuality_CurrentDates(t *testing.T) {
	content := []byte("This was released in 2025 and updated in 2026. Grade A (docs)")
	obs := ObserveSessionQuality(content)

	for _, o := range obs {
		if o.Category == "freshness" {
			t.Errorf("unexpected freshness observation for current dates: %v", o)
		}
	}
}

// TestObserveSessionQuality_DeduplicatesDates verifies stale dates are deduplicated.
func TestObserveSessionQuality_DeduplicatesDates(t *testing.T) {
	content := []byte("2019 was mentioned in 2019 and again 2019 was referenced. Grade A (docs)")
	obs := ObserveSessionQuality(content)

	for _, o := range obs {
		if o.Category == "freshness" {
			// Count occurrences of "2019" in message
			count := strings.Count(o.Message, "2019")
			if count > 1 {
				t.Errorf("expected deduplicated date in message, got %d occurrences of 2019: %s", count, o.Message)
			}
		}
	}
}

// TestObserveSessionQuality_EmptyContent verifies empty content produces no observations.
func TestObserveSessionQuality_EmptyContent(t *testing.T) {
	obs := ObserveSessionQuality([]byte(""))
	// Should have at least evidence observation (no grades)
	found := false
	for _, o := range obs {
		if o.Category == "evidence" {
			// Empty content shouldn't produce evidence observation (no body to grade)
			// Actually it should suggest adding grades
			found = true
		}
	}
	// Empty string won't match the grade pattern, so we should see a suggestion
	if !found {
		t.Error("expected evidence observation even for empty content")
	}
}

// TestObserveSessionQuality_PriorVariants verifies different Prior heading formats are recognized.
func TestObserveSessionQuality_PriorVariants(t *testing.T) {
	variants := []string{
		"## Prior\nBeliefs here",
		"## Prior (pre-research beliefs)\nBeliefs here",
	}
	for _, v := range variants {
		obs := ObserveSessionQuality([]byte(v))
		for _, o := range obs {
			if o.Category == "structure" && strings.Contains(o.Message, "No Prior section") {
				t.Errorf("Prior section was present but not recognized in: %q", v)
			}
		}
	}
}

// TestStaleDatePattern verifies the stale date regex pattern.
func TestStaleDatePattern(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"2019", true},
		{"2020", true},
		{"2023", true},
		{"2024", false}, // 2024+ is current
		{"2025", false},
		{"2026", false},
		{"2030", false},
		{"2010", true},
		{"1999", false}, // not matched by pattern
	}
	for _, tt := range tests {
		got := staleDatePattern.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("staleDatePattern.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestObserveSessionQuality_ConfirmationBias(t *testing.T) {
	t.Run("Delta with only confirmed beliefs emits bias observation", func(t *testing.T) {
		content := []byte(`
## Prior
- Believe SQLite works.

## Delta
| Belief | Status |
|---|---|
| SQLite works | Confirmed |
`)
		obs := ObserveSessionQuality(content)
		foundBias := false
		for _, o := range obs {
			if o.Category == "bias" {
				foundBias = true
				if !strings.Contains(o.Message, "Delta section has no contradicted or updated beliefs") {
					t.Errorf("unexpected bias message: %s", o.Message)
				}
			}
		}
		if !foundBias {
			t.Error("expected bias observation when Delta contains only confirmed beliefs")
		}
	})

	t.Run("Delta with contradicted belief does not emit bias observation", func(t *testing.T) {
		content := []byte(`
## Prior
- Believe SQLite works.

## Delta
| Belief | Status |
|---|---|
| SQLite works | Contradicted - WAL lock issue |
`)
		obs := ObserveSessionQuality(content)
		for _, o := range obs {
			if o.Category == "bias" {
				t.Errorf("unexpected bias observation when belief was contradicted: %v", o)
			}
		}
	})
}

func TestObserveSessionQuality_RecommendationAlternatives(t *testing.T) {
	t.Run("Recommendation without alternatives emits rigor observation", func(t *testing.T) {
		content := []byte(`
## Recommendation
Adopt PostgreSQL 16 for all primary relational workloads.
`)
		obs := ObserveSessionQuality(content)
		foundRigor := false
		for _, o := range obs {
			if o.Category == "rigor" {
				foundRigor = true
				if !strings.Contains(o.Message, "Recommendation doesn't reference rejected alternatives") {
					t.Errorf("unexpected rigor message: %s", o.Message)
				}
			}
		}
		if !foundRigor {
			t.Error("expected rigor observation when Recommendation omits alternatives")
		}
	})

	t.Run("Recommendation with rejected alternatives emits no rigor observation", func(t *testing.T) {
		content := []byte(`
## Recommendation
Adopt PostgreSQL 16 instead of MySQL or MongoDB which were rejected due to lack of vector indexing.
`)
		obs := ObserveSessionQuality(content)
		for _, o := range obs {
			if o.Category == "rigor" {
				t.Errorf("unexpected rigor observation when alternatives mentioned: %v", o)
			}
		}
	})
}

func TestObserveSessionQuality_RefinedStaleDates(t *testing.T) {
	t.Run("False positives on ports, latencies, and throughput are ignored", func(t *testing.T) {
		content := []byte(`
Connected to localhost:2020 with latency of 2015 ms and throughput 2022 req/s under v2021.0.
`)
		obs := ObserveSessionQuality(content)
		for _, o := range obs {
			if o.Category == "freshness" {
				t.Errorf("unexpected freshness observation for port/latency/qps: %v", o)
			}
		}
	})

	t.Run("Actual stale year reference is detected", func(t *testing.T) {
		content := []byte(`
According to a survey conducted in 2021, most teams choose Go.
`)
		obs := ObserveSessionQuality(content)
		foundFreshness := false
		for _, o := range obs {
			if o.Category == "freshness" {
				foundFreshness = true
				if !strings.Contains(o.Message, "2021") {
					t.Errorf("expected 2021 in freshness message, got: %s", o.Message)
				}
			}
		}
		if !foundFreshness {
			t.Error("expected freshness observation for survey conducted in 2021")
		}
	})
}

func TestIsLikelyYearReference_EdgeCases(t *testing.T) {
	tests := []struct {
		text  string
		start int
		end   int
		want  bool
	}{
		// Valid year references
		{"published in 2021 by author", 13, 17, true},
		{"from 2019 to 2024", 5, 9, true},
		{"since 2020.", 6, 10, true},

		// Ports, versions, variables
		{"localhost:2020/api", 10, 14, false},
		{"release v2020.1", 9, 13, false},
		{"release V2020.1", 9, 13, false},
		{"issue #2020 was closed", 7, 11, false},
		{"cost was $2020 total", 10, 14, false},
		{"mention @2020 user", 9, 13, false},

		// Percentages and multipliers
		{"growth was 2020% year over year", 11, 15, false},
		{"scale factor 2020x speedup", 13, 17, false},

		// Latency and throughput units
		{"latency of 2020 ms under load", 11, 15, false},
		{"latency of 2020ms under load", 11, 15, false},
		{"rate is 2020 qps sustained", 8, 12, false},
		{"rate is 2020req/s sustained", 8, 12, false},
		{"rate is 2020 rps sustained", 8, 12, false},
		{"rate is 2020 ops/s sustained", 8, 12, false},
		{"rendering 2020 fps benchmark", 10, 14, false},

		// Frequency and memory units
		{"frequency 2020 mhz clock", 10, 14, false},
		{"frequency 2020 ghz clock", 10, 14, false},
		{"frequency 2020 hz clock", 10, 14, false},
		{"frequency 2020 khz clock", 10, 14, false},
		{"allocated 2020 mb ram", 10, 14, false},
		{"allocated 2020 gb ram", 10, 14, false},
		{"allocated 2020 kb cache", 10, 14, false},
		{"allocated 2020 tb storage", 10, 14, false},
		{"allocated 2020 bytes total", 10, 14, false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := isLikelyYearReference(tt.text, tt.start, tt.end)
			if got != tt.want {
				t.Errorf("isLikelyYearReference(%q, %d, %d) = %v; want %v",
					tt.text, tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestObserveSessionQuality_ProvenanceVerification(t *testing.T) {
	// Grade A claims without URL/RFC anchors trigger unanchored provenance warning
	contentUnanchored := []byte(`
## Prior
Initial assumption.
## Key Findings
- Finding 1 [Grade A] — unanchored claim without source
- Finding 2 [Grade A] — another unanchored claim
- Finding 3 [Grade B] — benchmarks
## Discovered Concerns
None major.
## Delta
Contradicted old belief.
`)
	obs := ObserveSessionQuality(contentUnanchored)
	foundProv := false
	for _, o := range obs {
		if o.Category == "evidence" && strings.Contains(o.Message, "lack canonical URL or RFC provenance anchors") {
			foundProv = true
			break
		}
	}
	if !foundProv {
		t.Error("expected provenance warning when Grade A claims lack URL/RFC anchors")
	}

	// 100% Grade A claims with valid URL or RFC anchors do NOT trigger warning (ADR D-UX-02)
	contentAnchored := []byte(`
## Prior
Initial assumption.
## Key Findings
- Finding 1 [Grade A] https://tools.ietf.org/html/rfc7636
- Finding 2 [Grade A] https://www.sqlite.org/wal.html
- Finding 3 [Grade A] RFC 7519 JWT Specification
- Finding 4 [Grade A] doi:10.1145/3318464.3389700
## Discovered Concerns
None major.
## Delta
Contradicted old belief.
`)
	obs2 := ObserveSessionQuality(contentAnchored)
	for _, o := range obs2 {
		if o.Category == "evidence" && strings.Contains(o.Message, "lack canonical URL or RFC provenance anchors") {
			t.Errorf("unexpected provenance warning when Grade A claims are properly anchored: %v", o)
		}
	}
}

func TestObserveSessionQuality_DiscoveredConcerns(t *testing.T) {
	contentNoConcerns := []byte(`
## Prior
Initial assumption.
## Key Findings
- Finding 1 [Grade A]
- Finding 2 [Grade B]
- Finding 3 [Grade B]
## Delta
Contradicted old belief.
`)
	obs := ObserveSessionQuality(contentNoConcerns)
	foundConcerns := false
	for _, o := range obs {
		if o.Category == "concerns" && strings.Contains(o.Message, "Discovered Concerns") {
			foundConcerns = true
			break
		}
	}
	if !foundConcerns {
		t.Error("expected Discovered Concerns observation when section missing")
	}

	contentWithConcerns := []byte(`
## Prior
Initial assumption.
## Key Findings
- Finding 1 [Grade A]
- Finding 2 [Grade B]
- Finding 3 [Grade B]
## Discovered Concerns
Found latency spikes under load.
## Delta
Contradicted old belief.
`)
	obs2 := ObserveSessionQuality(contentWithConcerns)
	for _, o := range obs2 {
		if o.Category == "concerns" && strings.Contains(o.Message, "Discovered Concerns") {
			t.Errorf("unexpected concerns observation when section present: %v", o)
		}
	}
}

func TestObserveSessionQuality_KeyFindingsCount(t *testing.T) {
	contentFewFindings := []byte(`
## Prior
Initial assumption.
## Key Findings
- Only one finding [Grade A]
## Discovered Concerns
Concerns.
## Delta
Contradicted old belief.
`)
	obs := ObserveSessionQuality(contentFewFindings)
	foundDepth := false
	for _, o := range obs {
		if o.Category == "depth" && strings.Contains(o.Message, "Only 1 key finding") {
			foundDepth = true
			break
		}
	}
	if !foundDepth {
		t.Error("expected depth observation when fewer than 3 findings")
	}
}

func TestObserveSessionQuality_GradeAURLs(t *testing.T) {
	contentNoURL := []byte(`
## Prior
Initial assumption.
## Key Findings
- Primary claim [Grade A]
## Discovered Concerns
Concerns.
## Delta
Contradicted old belief.
## Evidence Ledger
| Claim | Grade A · fetched | Missing URL |
`)
	obs := ObserveSessionQuality(contentNoURL)
	foundURLWarn := false
	for _, o := range obs {
		if o.Category == "evidence" && strings.Contains(o.Message, "lack URLs") {
			foundURLWarn = true
			break
		}
	}
	if !foundURLWarn {
		t.Error("expected URL warning for Grade A fetched citation without URL")
	}

	contentWithURL := []byte(`
## Prior
Initial assumption.
## Key Findings
- Primary claim [Grade A]
## Discovered Concerns
Concerns.
## Delta
Contradicted old belief.
## Evidence Ledger
| Claim | Grade A · fetched | https://docs.example.com/spec |
`)
	obs2 := ObserveSessionQuality(contentWithURL)
	for _, o := range obs2 {
		if o.Category == "evidence" && strings.Contains(o.Message, "lack URLs") {
			t.Errorf("unexpected URL warning when URL present: %v", o)
		}
	}
}

func TestHasSectionHeading(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		keywords []string
		want     bool
	}{
		{"level 1 heading", "# Prior Beliefs\nSome text.", []string{"prior"}, true},
		{"level 2 heading", "## Prior\nSome text.", []string{"prior"}, true},
		{"level 3 heading", "### Discovered Concerns & Landmines\nDetails.", []string{"discovered concerns", "discovered concern"}, true},
		{"level 4 heading", "#### Delta and Changes\nTable.", []string{"delta"}, true},
		{"compound keywords", "## Key Findings & Benchmarks\n- A", []string{"key findings"}, true},
		{"narrative text not heading", "In the prior era, we used monoliths.\nNow microservices.", []string{"prior"}, false},
		{"subheading match", "## Context\n### Prior Knowledge\nNotes.", []string{"prior"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasSectionHeading(tt.body, tt.keywords...)
			if got != tt.want {
				t.Errorf("hasSectionHeading(%q, %v) = %v; want %v", tt.body, tt.keywords, got, tt.want)
			}
		})
	}
}

func TestBuildStaleDatePattern_EmptyYears(t *testing.T) {
	re := buildStaleDatePattern(2005)
	if re.MatchString("2005") || re.MatchString("anything") || re.MatchString("") {
		t.Errorf("expected empty pattern to match nothing")
	}
}

func TestObserveSessionQuality_GradeAInLedgerTable(t *testing.T) {
	content := []byte(`---
session_id: T0-01
status: draft
---
## Prior
Initial assumption.

## Key Findings
- [CLAIM-01] Postgres 16 introduces btree indexing speedups (Grade A · corroborated | fetched).

## Discovered Concerns
Memory pressure during maintenance operations.

## Delta
Performance gains observed.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Status | Supports |
|---|---|---|---|---|---|
| 1 | https://www.postgresql.org/docs/16/release-16.html | Grade A | corroborated | fetched | CLAIM-01 |
`)
	obs := ObserveSessionQuality(content)
	for _, o := range obs {
		if strings.Contains(o.Message, "lack canonical URL") {
			t.Fatalf("unexpected false Grade A warning when URL is in ledger table: %s", o.Message)
		}
	}
}

func TestHasSectionHeading_CaseInsensitive(t *testing.T) {
	doc := `
## Discovered Concerns
Here are unexpected findings.
`
	if !hasSectionHeading(doc, "DISCOVERED CONCERNS") {
		t.Errorf("expected hasSectionHeading to match uppercase keywords")
	}
	if !hasSectionHeading(doc, "Discovered Concerns") {
		t.Errorf("expected hasSectionHeading to match mixed-case keywords")
	}
	if hasSectionHeading(doc, "NONEXISTENT") {
		t.Errorf("expected false for nonexistent heading")
	}
}

func TestGradeALineRe_ImmunityToProperNouns(t *testing.T) {
	// Proper nouns starting with 'A' in parentheses must NOT be identified as Grade A lines
	falsePositives := []string{
		"Using Apache Kafka (Apache Kafka) for ingestion bus.",
		"Offloading log segments to Amazon S3 (AWS S3) storage.",
		"Real-time stream evaluation engine (Arroyo streaming engine).",
		"Anomaly detector (Anomaly detection worker).",
		"Distributed consensus layer (All previous sessions).",
	}

	for _, line := range falsePositives {
		if gradeALineRe.MatchString(line) {
			t.Errorf("gradeALineRe falsely matched proper noun: %q", line)
		}
	}

	// Genuine Grade A citations MUST match
	truePositives := []string{
		"Verified against official documentation [Grade A: official docs].",
		"Empirical benchmark results (Grade A · verified).",
		"Protocol wire format is Grade A (https://kafka.apache.org).",
		"Memory footprint measured at Grade: A.",
		"Primary spec verified [A · official spec].",
		"Direct benchmark corroborated (A | official).",
	}

	for _, line := range truePositives {
		if !gradeALineRe.MatchString(line) {
			t.Errorf("gradeALineRe failed to match genuine Grade A citation: %q", line)
		}
	}
}

func TestObserveSessionQuality_GradeDistributionSkew(t *testing.T) {
	// 4 out of 5 citations are Grade A (80% Grade A) -> triggers distribution skew observation
	content := []byte(`
Grade A (official docs: https://docs.example.com) for stability.
Grade A (RFC 7636) for authentication.
Grade A (benchmark https://bench.example.com) for throughput.
Grade A (spec https://spec.example.com) for protocol.
Grade B (informal blog) for latency.
`)
	obs := ObserveSessionQuality(content)

	foundSkew := false
	for _, o := range obs {
		if o.Category == "evidence" && strings.Contains(o.Message, "Grade distribution is") {
			foundSkew = true
			if !strings.Contains(o.Message, "80% Grade A") {
				t.Errorf("expected 80%% Grade A in observation message, got: %s", o.Message)
			}
		}
	}
	if !foundSkew {
		t.Error("expected Grade distribution skew observation when Grade A > 70%")
	}
}

