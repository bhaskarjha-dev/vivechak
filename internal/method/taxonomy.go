package method

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

// evidenceJSON is the one source of Vivechak's evidence definition: grades, modifiers,
// verification methods and rules. The registry below is parsed from it, the
// vivechak://methodology/grades resource renders it, and tools/docgen writes it into every
// document that states the definition. Edit evidence.json, then run `go run ./tools/docgen`.
//
//go:embed evidence.json
var evidenceJSON []byte

// EvidenceGrade defines an authoritative grade level in the Vivechak evidence hierarchy.
type EvidenceGrade struct {
	Grade         string   `json:"grade"`
	Label         string   `json:"label"`
	Description   string   `json:"description"`
	Provenance    []string `json:"provenance"`
	GateSatisfies []string `json:"gate_satisfies,omitempty"`
}

// ModifierValue is one value of a contextual modifier, with its optional meaning.
type ModifierValue struct {
	Value   string `json:"value"`
	Meaning string `json:"meaning,omitempty"`
}

// Modifier is a contextual evidence modifier (corroboration, recency, directness).
type Modifier struct {
	Name    string          `json:"name"`
	Purpose string          `json:"purpose"`
	Values  []ModifierValue `json:"values"`
}

// VerificationMethod is a verification label a claim carries (fetched, cached, ...).
type VerificationMethod struct {
	Method     string   `json:"method"`
	Definition string   `json:"definition"`
	GradeCap   string   `json:"grade_cap,omitempty"`
	Aliases    []string `json:"aliases,omitempty"`
}

// Definition is the parsed evidence definition (evidence.json).
type Definition struct {
	Comment      string               `json:"$comment,omitempty"`
	Grades       []EvidenceGrade      `json:"grades"`
	Modifiers    []Modifier           `json:"modifiers"`
	Verification []VerificationMethod `json:"verification"`
	Rules        []string             `json:"rules"`
}

var (
	// Grades is the canonical registry of evidence grades (A–E), parsed from evidence.json.
	Grades []EvidenceGrade
	// Modifiers lists the contextual modifiers, parsed from evidence.json.
	Modifiers []Modifier
	// VerificationMethods lists the verification labels, parsed from evidence.json.
	VerificationMethods []VerificationMethod
	// Rules lists the hard grading rules, parsed from evidence.json.
	Rules []string
)

func init() {
	def, err := ParseDefinition(evidenceJSON)
	if err != nil {
		panic(fmt.Sprintf("internal/method/evidence.json: %v", err))
	}
	Grades, Modifiers, VerificationMethods, Rules = def.Grades, def.Modifiers, def.Verification, def.Rules
}

// ParseDefinition parses and checks an evidence definition. The rest of the code (the gate's
// claim recognizer, the validators) assumes exactly the grades A to E, so anything else is an error.
func ParseDefinition(data []byte) (Definition, error) {
	var def Definition
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&def); err != nil {
		return Definition{}, err
	}
	letters := make([]string, 0, len(def.Grades))
	for _, g := range def.Grades {
		if strings.TrimSpace(g.Label) == "" || strings.TrimSpace(g.Description) == "" || len(g.Provenance) == 0 {
			return Definition{}, fmt.Errorf("grade %q needs a label, a description and provenance", g.Grade)
		}
		letters = append(letters, g.Grade)
	}
	if strings.Join(letters, "") != "ABCDE" {
		return Definition{}, fmt.Errorf("grades must be A, B, C, D, E in that order (got %v)", letters)
	}
	if len(def.Modifiers) == 0 || len(def.Verification) == 0 || len(def.Rules) == 0 {
		return Definition{}, fmt.Errorf("modifiers, verification and rules must not be empty")
	}
	for _, m := range def.Modifiers {
		if strings.TrimSpace(m.Name) == "" || len(m.Values) == 0 {
			return Definition{}, fmt.Errorf("modifier %q needs a name and values", m.Name)
		}
	}
	for _, v := range def.Verification {
		if strings.TrimSpace(v.Method) == "" || strings.TrimSpace(v.Definition) == "" {
			return Definition{}, fmt.Errorf("verification method %q needs a name and a definition", v.Method)
		}
	}
	return def, nil
}

// GetGrade returns the EvidenceGrade for a given grade letter, or nil if not found.
func GetGrade(grade string) *EvidenceGrade {
	upper := strings.ToUpper(strings.TrimSpace(grade))
	for _, g := range Grades {
		if g.Grade == upper {
			return &g
		}
	}
	return nil
}

// ValidGrade reports whether grade is one of A, B, C, D, or E.
func ValidGrade(grade string) bool {
	return GetGrade(grade) != nil
}

// FormatTaxonomyMarkdown renders the canonical evidence taxonomy card (the
// vivechak://methodology/grades resource) as Markdown.
func FormatTaxonomyMarkdown() string {
	var b strings.Builder
	b.WriteString("# Vivechak Canonical Evidence Taxonomy (Grades A–E)\n\n")
	b.WriteString(GradesTableMarkdown())
	b.WriteString("\n*Gate*: G6 counts Grade A/B claims that sit on a line with their URL, written as a tag that starts with the grade " +
		"(`(Grade A · corroborated · fresh | fetched)`, `[A · single · fresh · direct · fetched](url)`, `[Grade A: url | fetched]`) " +
		"or an evidence-ledger row whose Grade cell is A or B; G7 counts those whose verification label is exactly `fetched` or `human` (`human-provided`).\n")
	b.WriteString("\n## Contextual Modifiers\n\n")
	b.WriteString(ModifiersTableMarkdown())
	b.WriteString("\n## Verification Methods\n\n")
	b.WriteString(VerificationTableMarkdown())
	return b.String()
}
