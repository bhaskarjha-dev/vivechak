package method

import (
	"encoding/json"
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

// TestEvidenceDefinitionIsTheOneSource checks that the registry, the resource card and every
// rendering tools/docgen writes into the documents come from evidence.json, so a wording change
// there reaches all of them.
func TestEvidenceDefinitionIsTheOneSource(t *testing.T) {
	def, err := ParseDefinition(evidenceJSON)
	if err != nil {
		t.Fatalf("evidence.json: %v", err)
	}
	if len(def.Grades) != len(Grades) || len(def.Modifiers) != len(Modifiers) ||
		len(def.Verification) != len(VerificationMethods) || len(def.Rules) != len(Rules) {
		t.Fatal("registry was not loaded from evidence.json")
	}
	card := FormatTaxonomyMarkdown()
	core := GeneratorCoreMarkdown()
	skill := SkillMarkdown()
	list := GradeListMarkdown()
	table := GradesTableMarkdown()
	for _, g := range def.Grades {
		for name, text := range map[string]string{"resource": card, "skill": skill, "grade list": list, "grades table": table} {
			if !strings.Contains(text, g.Description) || !strings.Contains(text, g.Label) {
				t.Errorf("%s rendering lacks grade %s (%q, %q)", name, g.Grade, g.Label, g.Description)
			}
		}
		if !strings.Contains(core, g.Grade+" ("+inlineDescription(g.Description)+")") {
			t.Errorf("generator CORE rendering lacks grade %s's definition", g.Grade)
		}
	}
	for _, m := range def.Modifiers {
		for _, v := range m.Values {
			for name, text := range map[string]string{"resource": card, "skill": skill, "core": core, "modifiers table": ModifiersTableMarkdown()} {
				if !strings.Contains(text, v.Value) {
					t.Errorf("%s rendering lacks modifier value %q", name, v.Value)
				}
			}
		}
	}
	for _, v := range def.Verification {
		for name, text := range map[string]string{"resource": card, "skill": skill, "core": core, "verification table": VerificationTableMarkdown()} {
			if !strings.Contains(text, v.Method) {
				t.Errorf("%s rendering lacks verification method %q", name, v.Method)
			}
		}
	}
	if !strings.Contains(core, "recalled claims capped at Grade D") {
		t.Errorf("generator CORE rendering lacks the recall cap: %s", core)
	}
}

func TestParseDefinitionRejectsBrokenDefinitions(t *testing.T) {
	valid := string(evidenceJSON)
	cases := map[string]string{
		"unknown field": strings.Replace(valid, `"label": "Primary Empirical"`, `"lable": "Primary Empirical"`, 1),
		"grade F":       strings.Replace(valid, `"grade": "E"`, `"grade": "F"`, 1),
		"missing grade": strings.Replace(valid, `"grade": "C"`, `"grade": "B"`, 1),
		"empty label":   strings.Replace(valid, `"label": "Secondary Empirical"`, `"label": ""`, 1),
		"not JSON":      "{",
	}
	for name, data := range cases {
		if data == valid {
			t.Fatalf("%s: the replacement did not apply; update the test to the current evidence.json", name)
		}
		if _, err := ParseDefinition([]byte(data)); err == nil {
			t.Errorf("%s: expected ParseDefinition to fail", name)
		}
	}
	def, err := ParseDefinition(evidenceJSON)
	if err != nil {
		t.Fatal(err)
	}
	def.Verification = nil
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseDefinition(data); err == nil {
		t.Error("no verification methods: expected ParseDefinition to fail")
	}
}

func TestInlineDescription(t *testing.T) {
	for in, want := range map[string]string{
		"Official specifications, RFCs.": "official specifications, RFCs",
		"RFCs and specifications.":       "RFCs and specifications",
		"StackOverflow answers.":         "StackOverflow answers",
		"":                               "",
	} {
		if got := inlineDescription(in); got != want {
			t.Errorf("inlineDescription(%q) = %q, want %q", in, got, want)
		}
	}
}
