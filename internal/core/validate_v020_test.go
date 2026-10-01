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

// TestObserveSessionQuality_ManyGrades verifies no evidence observation when 3+ grades present.
func TestObserveSessionQuality_ManyGrades(t *testing.T) {
	content := []byte(`
Grade A (official docs) for stability claim.
Grade B (benchmarks) for performance claim.
Grade C (vendor claims) for scalability.
Grade A (peer-reviewed) for security.
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
