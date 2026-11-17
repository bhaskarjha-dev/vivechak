package method

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The renderers below turn the evidence definition (evidence.json) into the text that
// tools/docgen writes between generated-block markers. Every document that states the
// definition gets it from here, so the wording can't drift between them.

// GeneratorCoreMarkdown renders the evidence lines inside the generators' CORE:BEGIN/END block.
func GeneratorCoreMarkdown() string {
	grades := make([]string, 0, len(Grades))
	for _, g := range Grades {
		grades = append(grades, fmt.Sprintf("%s (%s)", g.Grade, inlineDescription(g.Description)))
	}
	mods := make([]string, 0, len(Modifiers))
	for _, m := range Modifiers {
		vals := make([]string, 0, len(m.Values))
		for _, v := range m.Values {
			vals = append(vals, v.Value)
		}
		mods = append(mods, fmt.Sprintf("%s (%s)", strings.ToLower(m.Name), strings.Join(vals, "/")))
	}
	methods := make([]string, 0, len(VerificationMethods))
	var caps []string
	for _, v := range VerificationMethods {
		methods = append(methods, v.Method)
		if v.GradeCap != "" {
			caps = append(caps, fmt.Sprintf("%s claims capped at Grade %s regardless of apparent source", v.Method, v.GradeCap))
		}
	}
	verification := strings.Join(methods, " | ")
	if len(caps) > 0 {
		verification += " (" + strings.Join(caps, "; ") + ")"
	}
	var b strings.Builder
	b.WriteString("For evidence grading, every factual claim should carry:\n")
	b.WriteString("- Base grade: " + strings.Join(grades, " | ") + "\n")
	b.WriteString("- Modifiers: " + strings.Join(mods, ", ") + "\n")
	b.WriteString("- Verification: " + verification + "\n")
	return b.String()
}

// GradeListMarkdown renders the grades as a Markdown bullet list, one grade per line.
func GradeListMarkdown() string {
	var b strings.Builder
	for _, g := range Grades {
		fmt.Fprintf(&b, "- **Grade %s (%s):** %s\n", g.Grade, g.Label, g.Description)
	}
	return b.String()
}

// SkillMarkdown renders the grades, modifiers, verification labels and rules for the agent skill.
func SkillMarkdown() string {
	var b strings.Builder
	b.WriteString(GradeListMarkdown())
	mods := make([]string, 0, len(Modifiers))
	for _, m := range Modifiers {
		vals := make([]string, 0, len(m.Values))
		for _, v := range m.Values {
			vals = append(vals, "`"+v.Value+"`")
		}
		mods = append(mods, fmt.Sprintf("%s (%s)", strings.ToLower(m.Name), strings.Join(vals, " | ")))
	}
	fmt.Fprintf(&b, "\n**Modifiers:** %s.\n", strings.Join(mods, ", "))
	fmt.Fprintf(&b, "\n**Verification:** %s.\n", strings.Join(verificationNames(), " | "))
	fmt.Fprintf(&b, "\n**Rules:** %s\n", strings.Join(Rules, " "))
	return b.String()
}

// GradesTableMarkdown renders the grades table followed by the grading rules.
func GradesTableMarkdown() string {
	var b strings.Builder
	b.WriteString("| Grade | Label | Description | Allowed Provenance | Gate Impact |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, g := range Grades {
		gate := "Advisory"
		if len(g.GateSatisfies) > 0 {
			gate = strings.Join(g.GateSatisfies, ", ")
		}
		fmt.Fprintf(&b, "| **Grade %s** | %s | %s | `%s` | %s |\n",
			g.Grade, g.Label, g.Description, strings.Join(g.Provenance, ", "), gate)
	}
	fmt.Fprintf(&b, "\n*Rule*: %s\n", strings.Join(Rules, " "))
	return b.String()
}

// ModifiersTableMarkdown renders the contextual modifiers table.
func ModifiersTableMarkdown() string {
	var b strings.Builder
	b.WriteString("| Modifier | Values | Purpose |\n")
	b.WriteString("|---|---|---|\n")
	for _, m := range Modifiers {
		vals := make([]string, 0, len(m.Values))
		for _, v := range m.Values {
			if v.Meaning != "" {
				vals = append(vals, fmt.Sprintf("`%s` (%s)", v.Value, v.Meaning))
			} else {
				vals = append(vals, "`"+v.Value+"`")
			}
		}
		fmt.Fprintf(&b, "| **%s** | %s | %s |\n", m.Name, strings.Join(vals, " · "), m.Purpose)
	}
	return b.String()
}

// VerificationTableMarkdown renders the verification methods table.
func VerificationTableMarkdown() string {
	var b strings.Builder
	b.WriteString("| Method | Definition | Grade Cap |\n")
	b.WriteString("|---|---|---|\n")
	for _, v := range VerificationMethods {
		capText := "None"
		if v.GradeCap != "" {
			capText = fmt.Sprintf("**Capped at Grade %s**", v.GradeCap)
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", verificationName(v), v.Definition, capText)
	}
	return b.String()
}

func verificationNames() []string {
	names := make([]string, 0, len(VerificationMethods))
	for _, v := range VerificationMethods {
		names = append(names, verificationName(v))
	}
	return names
}

// verificationName renders a method with its accepted aliases: `human-provided` (or `human`).
func verificationName(v VerificationMethod) string {
	name := "`" + v.Method + "`"
	for _, a := range v.Aliases {
		name += " (or `" + a + "`)"
	}
	return name
}

// inlineDescription turns "Official specifications, ... anchor." into "official specifications, ... anchor"
// for use inside a sentence. A leading acronym or proper noun ("RFCs", "StackOverflow") keeps its case.
func inlineDescription(s string) string {
	s = strings.TrimSuffix(strings.TrimSpace(s), ".")
	first, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return s
	}
	word := s[size:]
	if i := strings.IndexAny(word, " ,;"); i >= 0 {
		word = word[:i]
	}
	if strings.IndexFunc(word, unicode.IsUpper) >= 0 {
		return s
	}
	return string(unicode.ToLower(first)) + s[size:]
}
