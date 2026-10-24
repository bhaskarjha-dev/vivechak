package method

import (
	"fmt"
	"strings"
)

// EvidenceGrade defines an authoritative grade level in the Vivechak evidence hierarchy.
type EvidenceGrade struct {
	Grade         string   `json:"grade"`
	Label         string   `json:"label"`
	Description   string   `json:"description"`
	Provenance    []string `json:"provenance"`
	GateSatisfies []string `json:"gate_satisfies,omitempty"`
}

// Grades is the canonical registry of evidence grades (A–E) in Vivechak.
// This is the single source of truth across CLI, MCP resources, prompts, and gate rules.
var Grades = []EvidenceGrade{
	{
		Grade:       "A",
		Label:       "Primary Empirical",
		Description: "Official specifications, canonical source code, reproducible benchmarks, or primary documentation with explicit URL anchor.",
		Provenance:  []string{"fetched", "verified"},
		GateSatisfies: []string{"G6", "G7"},
	},
	{
		Grade:       "B",
		Label:       "Secondary Empirical",
		Description: "Peer-reviewed papers, reputable third-party benchmarks, published postmortems, or corroborated engineering reports.",
		Provenance:  []string{"fetched", "cached"},
		GateSatisfies: []string{"G6"},
	},
	{
		Grade:       "C",
		Label:       "Vendor & Practitioner",
		Description: "Vendor marketing, product whitepapers, conference talks, official tutorials, or non-reproduced claims.",
		Provenance:  []string{"fetched", "cached"},
		GateSatisfies: nil,
	},
	{
		Grade:       "D",
		Label:       "Informal & Opinion",
		Description: "Blog posts, forum discussions, StackOverflow answers, or model memory recall (capped at Grade D).",
		Provenance:  []string{"fetched", "cached", "recalled"},
		GateSatisfies: nil,
	},
	{
		Grade:       "E",
		Label:       "Speculative & Unsubstantiated",
		Description: "Unverified hypotheses, uncorroborated parametric recall, or ungrounded assertions.",
		Provenance:  []string{"recalled", "unverified"},
		GateSatisfies: nil,
	},
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

// FormatTaxonomyMarkdown renders the canonical evidence taxonomy card as Markdown.
func FormatTaxonomyMarkdown() string {
	var b strings.Builder
	b.WriteString("# Vivechak Canonical Evidence Taxonomy (Grades A–E)\n\n")
	b.WriteString("| Grade | Label | Description | Allowed Provenance | Gate Impact |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, g := range Grades {
		prov := strings.Join(g.Provenance, ", ")
		gate := "Advisory"
		if len(g.GateSatisfies) > 0 {
			gate = strings.Join(g.GateSatisfies, ", ")
		}
		b.WriteString(fmt.Sprintf("| **Grade %s** | %s | %s | `%s` | %s |\n",
			g.Grade, g.Label, g.Description, prov, gate))
	}
	b.WriteString("\n*Rule*: Parametric memory recall is capped at **Grade D**. Grade A requires canonical URL or specification anchor.\n")
	return b.String()
}
