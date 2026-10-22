package embed

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEmbedIntegrity_Generators(t *testing.T) {
	expected := []string{"GENERATOR-COMPARISON.md", "GENERATOR-DECISION.md", "GENERATOR-PROJECT.md"}

	names, err := ListGenerators()
	if err != nil {
		t.Fatalf("ListGenerators: %v", err)
	}

	if len(names) != len(expected) {
		t.Fatalf("expected %d generators, got %d: %v", len(expected), len(names), names)
	}

	for _, name := range expected {
		data, err := ReadGenerator(name)
		if err != nil {
			t.Errorf("ReadGenerator(%q): %v", name, err)
			continue
		}
		if len(data) < 100 {
			t.Errorf("generator %q seems too small (%d bytes)", name, len(data))
		}
	}
}

func TestEmbedIntegrity_Templates(t *testing.T) {
	expected := []string{
		"COMPARISON-SESSION.template.md",
		"CONFLICT-RESOLUTION.template.md",
		"DECISIONS.template.md",
		"FOUNDING-ARCHITECTURE.template.md",
		"PHASE-0-GATE.template.md",
		"SESSION.template.md",
	}

	names, err := ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}

	if len(names) != len(expected) {
		t.Fatalf("expected %d templates, got %d: %v", len(expected), len(names), names)
	}

	for _, name := range expected {
		data, err := ReadTemplate(name)
		if err != nil {
			t.Errorf("ReadTemplate(%q): %v", name, err)
			continue
		}
		if len(data) < 100 {
			t.Errorf("template %q seems too small (%d bytes)", name, len(data))
		}
	}
}

func TestEmbedIntegrity_Skills(t *testing.T) {
	skills, err := ListSkills()
	if err != nil {
		t.Fatalf("ListSkills: %v", err)
	}
	if len(skills) != 1 || skills[0] != "SKILL.md" {
		t.Fatalf("expected [SKILL.md], got %v", skills)
	}

	data, err := ReadSkill("SKILL.md")
	if err != nil {
		t.Fatalf("ReadSkill(SKILL.md): %v", err)
	}
	if len(data) < 100 {
		t.Errorf("skill seems too small (%d bytes)", len(data))
	}
	if !bytes.Contains(data, []byte("name: vivechak")) {
		t.Errorf("skill missing 'name: vivechak' header")
	}

	// Error path: non-existent skill
	if _, err := ReadSkill("nonexistent-skill.md"); err == nil {
		t.Error("expected error for nonexistent skill, got nil")
	}
}

func TestEmbedIntegrity_Rules(t *testing.T) {
	rules, err := ListRules()
	if err != nil {
		t.Fatalf("ListRules: %v", err)
	}
	expected := []string{"antigravity.md", "claude.md", "cursor.mdc"}
	if len(rules) != len(expected) {
		t.Fatalf("expected %d rules, got %d: %v", len(expected), len(rules), rules)
	}

	for _, name := range expected {
		data, err := ReadRule(name)
		if err != nil {
			t.Errorf("ReadRule(%q): %v", name, err)
			continue
		}
		if len(data) < 50 {
			t.Errorf("rule %q seems too small (%d bytes)", name, len(data))
		}
	}

	// Error path: non-existent rule
	if _, err := ReadRule("nonexistent-rule.mdc"); err == nil {
		t.Error("expected error for nonexistent rule, got nil")
	}
}

func TestEmbeddedFilesMatchRoot(t *testing.T) {
	rootGenDir := filepath.Join("..", "..")
	for _, name := range []string{"GENERATOR-PROJECT.md", "GENERATOR-DECISION.md", "GENERATOR-COMPARISON.md"} {
		rootPath := filepath.Join(rootGenDir, name)
		rootData, err := os.ReadFile(rootPath)
		if os.IsNotExist(err) {
			t.Skip("root files not found (running outside git repo)")
			return
		}
		if err != nil {
			t.Fatalf("reading %s: %v", rootPath, err)
		}
		embedData, err := ReadGenerator(name)
		if err != nil {
			t.Fatalf("ReadGenerator(%s): %v", name, err)
		}
		if !bytes.Equal(rootData, embedData) {
			t.Errorf("Embedded generator %s does not match root %s (bytes: %d embed vs %d root)", name, rootPath, len(embedData), len(rootData))
		}
	}

	rootTmplDir := filepath.Join("..", "..", "templates")
	for _, name := range []string{
		"COMPARISON-SESSION.template.md",
		"CONFLICT-RESOLUTION.template.md",
		"DECISIONS.template.md",
		"FOUNDING-ARCHITECTURE.template.md",
		"PHASE-0-GATE.template.md",
		"SESSION.template.md",
	} {
		rootPath := filepath.Join(rootTmplDir, name)
		rootData, err := os.ReadFile(rootPath)
		if err != nil {
			t.Fatalf("reading %s: %v", rootPath, err)
		}
		embedData, err := ReadTemplate(name)
		if err != nil {
			t.Fatalf("ReadTemplate(%s): %v", name, err)
		}
		if !bytes.Equal(rootData, embedData) {
			t.Errorf("Embedded template %s does not match root %s (bytes: %d embed vs %d root)", name, rootPath, len(embedData), len(rootData))
		}
	}

	// Verify Agent Skill matches root
	rootSkillPath := filepath.Join("..", "..", "skills", "vivechak", "SKILL.md")
	rootSkillData, err := os.ReadFile(rootSkillPath)
	if err != nil {
		t.Fatalf("reading %s: %v", rootSkillPath, err)
	}
	embedSkillData, err := ReadSkill("SKILL.md")
	if err != nil {
		t.Fatalf("ReadSkill(SKILL.md): %v", err)
	}
	if !bytes.Equal(rootSkillData, embedSkillData) {
		t.Errorf("Embedded skill SKILL.md does not match root %s", rootSkillPath)
	}

	// Verify Rule templates match root
	rootRulesDir := filepath.Join("..", "..", "templates", "rules")
	for _, name := range []string{"cursor.mdc", "claude.md", "antigravity.md"} {
		rootPath := filepath.Join(rootRulesDir, name)
		rootData, err := os.ReadFile(rootPath)
		if err != nil {
			t.Fatalf("reading %s: %v", rootPath, err)
		}
		embedData, err := ReadRule(name)
		if err != nil {
			t.Fatalf("ReadRule(%s): %v", name, err)
		}
		if !bytes.Equal(rootData, embedData) {
			t.Errorf("Embedded rule %s does not match root %s", name, rootPath)
		}
	}
}

func TestCoreMarkersMatchAcrossGenerators(t *testing.T) {
	generators := []string{"GENERATOR-PROJECT.md", "GENERATOR-DECISION.md", "GENERATOR-COMPARISON.md"}
	var extractedCores [][]byte

	startMarker := []byte("<!-- CORE:BEGIN")
	endMarker := []byte("<!-- CORE:END -->")

	for _, name := range generators {
		data, err := ReadGenerator(name)
		if err != nil {
			t.Fatalf("ReadGenerator(%s): %v", name, err)
		}
		sIdx := bytes.Index(data, startMarker)
		if sIdx == -1 {
			t.Fatalf("%s: missing CORE:BEGIN marker", name)
		}
		eIdx := bytes.Index(data[sIdx:], endMarker)
		if eIdx == -1 {
			t.Fatalf("%s: missing CORE:END marker", name)
		}
		block := data[sIdx : sIdx+eIdx+len(endMarker)]
		block = bytes.ReplaceAll(block, []byte("\r\n"), []byte("\n"))
		extractedCores = append(extractedCores, block)
	}

	for i := 1; i < len(generators); i++ {
		if !bytes.Equal(extractedCores[0], extractedCores[i]) {
			t.Errorf("CORE kernel drift between %s and %s:\n--- %s ---\n%s\n--- %s ---\n%s",
				generators[0], generators[i],
				generators[0], string(extractedCores[0]),
				generators[i], string(extractedCores[i]))
		}
	}
}
