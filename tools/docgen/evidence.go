package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/method"
)

// The evidence definition (grades, modifiers, verification methods, rules) has one source,
// internal/method/evidence.json. Every document that states it carries a generated block
// between markers, and this file rewrites those blocks from the source:
//
//	<!-- docgen:begin <id> (...) -->   ...generated text...   <!-- docgen:end <id> -->
//
// The generators keep their CORE:BEGIN/CORE:END markers for the same purpose. Generated lines take
// the indentation of the begin marker, so a block can sit inside a list item. Files with an
// embedded twin under internal/embed/ get the same bytes written to the twin.

// evidenceBlock is one generated block: its marker id ("" means the generators' CORE block) and renderer.
type evidenceBlock struct {
	id     string
	render func() string
}

// evidenceTarget is a file carrying generated evidence blocks, with its embedded twin if it has one.
type evidenceTarget struct {
	path   string // slash-separated, relative to the repository root
	twin   string // embedded twin, or ""
	blocks []evidenceBlock
}

var (
	coreEvidenceBlock         = evidenceBlock{id: "", render: method.GeneratorCoreMarkdown}
	gradeListBlock            = evidenceBlock{id: "evidence-grades", render: method.GradeListMarkdown}
	skillEvidenceBlock        = evidenceBlock{id: "evidence-skill", render: method.SkillMarkdown}
	gradesTableBlock          = evidenceBlock{id: "evidence-grades-table", render: method.GradesTableMarkdown}
	modifiersTableBlock       = evidenceBlock{id: "evidence-modifiers-table", render: method.ModifiersTableMarkdown}
	verificationTableBlock    = evidenceBlock{id: "evidence-verification-table", render: method.VerificationTableMarkdown}
	coreBeginRe               = regexp.MustCompile(`^\s*<!-- CORE:BEGIN\b.*-->\s*$`)
	coreEndRe                 = regexp.MustCompile(`^\s*<!-- CORE:END -->\s*$`)
	leadingWhitespaceRe       = regexp.MustCompile(`^[ \t]*`)
	docgenBeginMarkerTemplate = `^\s*<!-- docgen:begin %s(\s.*)?-->\s*$`
	docgenEndMarkerTemplate   = `^\s*<!-- docgen:end %s -->\s*$`
)

func evidenceTargets() []evidenceTarget {
	return []evidenceTarget{
		{"GENERATOR-PROJECT.md", "internal/embed/generators/GENERATOR-PROJECT.md", []evidenceBlock{coreEvidenceBlock}},
		{"GENERATOR-DECISION.md", "internal/embed/generators/GENERATOR-DECISION.md", []evidenceBlock{coreEvidenceBlock}},
		{"GENERATOR-COMPARISON.md", "internal/embed/generators/GENERATOR-COMPARISON.md", []evidenceBlock{coreEvidenceBlock}},
		{"skills/vivechak/SKILL.md", "internal/embed/skills/SKILL.md", []evidenceBlock{skillEvidenceBlock}},
		{"templates/rules/antigravity.md", "internal/embed/rules/antigravity.md", []evidenceBlock{gradeListBlock}},
		{"templates/rules/claude.md", "internal/embed/rules/claude.md", []evidenceBlock{gradeListBlock}},
		{"templates/rules/cursor.mdc", "internal/embed/rules/cursor.mdc", []evidenceBlock{gradeListBlock}},
		{"README.md", "", []evidenceBlock{gradeListBlock}},
		{"docs/QUICKSTART.md", "", []evidenceBlock{gradeListBlock}},
		{"FRAMEWORK.md", "", []evidenceBlock{gradesTableBlock, modifiersTableBlock, verificationTableBlock}},
	}
}

// generateEvidenceBlocks rewrites every generated evidence block (and embedded twin) under rootDir.
func generateEvidenceBlocks(rootDir string) error {
	for _, t := range evidenceTargets() {
		path := filepath.Join(rootDir, filepath.FromSlash(t.path))
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out, err := renderEvidenceBlocks(data, t.blocks)
		if err != nil {
			return fmt.Errorf("%s: %w", t.path, err)
		}
		if err := writeIfChanged(path, out); err != nil {
			return err
		}
		if t.twin != "" {
			if err := writeIfChanged(filepath.Join(rootDir, filepath.FromSlash(t.twin)), out); err != nil {
				return err
			}
		}
	}
	return nil
}

// renderEvidenceBlocks returns data with each of the blocks rewritten from the evidence definition.
func renderEvidenceBlocks(data []byte, blocks []evidenceBlock) ([]byte, error) {
	out := data
	for _, b := range blocks {
		var err error
		if out, err = replaceBlock(out, b); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// replaceBlock replaces the lines between the block's begin and end markers with its rendered text.
// The begin marker must appear exactly once and be followed by an end marker.
func replaceBlock(data []byte, b evidenceBlock) ([]byte, error) {
	beginRe, endRe, name := coreBeginRe, coreEndRe, "CORE:BEGIN/CORE:END"
	if b.id != "" {
		beginRe = regexp.MustCompile(fmt.Sprintf(docgenBeginMarkerTemplate, regexp.QuoteMeta(b.id)))
		endRe = regexp.MustCompile(fmt.Sprintf(docgenEndMarkerTemplate, regexp.QuoteMeta(b.id)))
		name = "docgen:begin/end " + b.id
	}
	nl := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		nl = "\r\n"
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	trailingNewline := strings.HasSuffix(text, "\n")
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")

	begin, end := -1, -1
	for i, l := range lines {
		if beginRe.MatchString(l) {
			if begin != -1 {
				return nil, fmt.Errorf("%s marker appears more than once (lines %d and %d)", name, begin+1, i+1)
			}
			begin = i
		}
	}
	if begin == -1 {
		return nil, fmt.Errorf("missing %s begin marker", name)
	}
	for i := begin + 1; i < len(lines); i++ {
		if endRe.MatchString(lines[i]) {
			end = i
			break
		}
	}
	if end == -1 {
		return nil, fmt.Errorf("%s begin marker on line %d has no end marker", name, begin+1)
	}

	indent := leadingWhitespaceRe.FindString(lines[begin])
	var generated []string
	for _, l := range strings.Split(strings.TrimSuffix(b.render(), "\n"), "\n") {
		if l != "" {
			l = indent + l
		}
		generated = append(generated, l)
	}
	merged := make([]string, 0, len(lines)-(end-begin-1)+len(generated))
	merged = append(merged, lines[:begin+1]...)
	merged = append(merged, generated...)
	merged = append(merged, lines[end:]...)
	result := strings.Join(merged, "\n")
	if trailingNewline {
		result += "\n"
	}
	return []byte(strings.ReplaceAll(result, "\n", nl)), nil
}

func writeIfChanged(path string, data []byte) error {
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, data) {
		return nil
	}
	return os.WriteFile(path, data, 0o644)
}
