package method

import (
	"strings"
	"testing"
)

func TestTaxonomy(t *testing.T) {
	for _, letter := range []string{"A", "B", "C", "D", "E"} {
		if !ValidGrade(letter) {
			t.Errorf("expected grade %s to be valid", letter)
		}
		g := GetGrade(letter)
		if g == nil {
			t.Fatalf("expected GetGrade(%s) to be non-nil", letter)
		}
		if len(g.Provenance) == 0 {
			t.Errorf("grade %s has no provenance tags", letter)
		}
	}

	if ValidGrade("F") || ValidGrade("Z") {
		t.Error("unexpected valid grade for F or Z")
	}

	card := FormatTaxonomyMarkdown()
	if !strings.Contains(card, "Primary Empirical") || !strings.Contains(card, "Grade A") {
		t.Errorf("markdown card missing expected content: %s", card)
	}
}
