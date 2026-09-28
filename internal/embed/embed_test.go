package embed

import "testing"

func TestEmbedIntegrity_Generators(t *testing.T) {
	expected := []string{"GENERATOR.md", "GENERATOR-DECISION.md", "GENERATOR-COMPARISON.md"}

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
