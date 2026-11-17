package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceBlock_IndentsNestedBlockAndIsIdempotent(t *testing.T) {
	block := evidenceBlock{id: "evidence-grades", render: func() string { return "- one\n\n- two\n" }}
	in := "intro\n" +
		"5. **Grades:**\n" +
		"   <!-- docgen:begin evidence-grades (generated; do not edit) -->\n" +
		"   - stale hand-edited line\n" +
		"   <!-- docgen:end evidence-grades -->\n" +
		"6. next\n"
	want := "intro\n" +
		"5. **Grades:**\n" +
		"   <!-- docgen:begin evidence-grades (generated; do not edit) -->\n" +
		"   - one\n" +
		"\n" +
		"   - two\n" +
		"   <!-- docgen:end evidence-grades -->\n" +
		"6. next\n"
	got, err := replaceBlock([]byte(in), block)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("replaceBlock:\n%s\nwant:\n%s", got, want)
	}
	again, err := replaceBlock(got, block)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != want {
		t.Fatalf("second run changed the file:\n%s", again)
	}
}

func TestReplaceBlock_CoreMarkersAndCRLF(t *testing.T) {
	in := "a\r\n<!-- CORE:BEGIN — generated -->\r\nold\r\n<!-- CORE:END -->\r\nb\r\n"
	got, err := replaceBlock([]byte(in), evidenceBlock{render: func() string { return "new 1\nnew 2\n" }})
	if err != nil {
		t.Fatal(err)
	}
	want := "a\r\n<!-- CORE:BEGIN — generated -->\r\nnew 1\r\nnew 2\r\n<!-- CORE:END -->\r\nb\r\n"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReplaceBlock_RefusesMissingOrRepeatedMarkers(t *testing.T) {
	block := evidenceBlock{id: "evidence-grades", render: func() string { return "x\n" }}
	for name, in := range map[string]string{
		"no markers":   "text\n",
		"no end":       "<!-- docgen:begin evidence-grades -->\ntext\n",
		"end only":     "text\n<!-- docgen:end evidence-grades -->\n",
		"twice":        "<!-- docgen:begin evidence-grades -->\n<!-- docgen:end evidence-grades -->\n<!-- docgen:begin evidence-grades -->\n<!-- docgen:end evidence-grades -->\n",
		"other id":     "<!-- docgen:begin evidence-grades-table -->\n<!-- docgen:end evidence-grades-table -->\n",
		"end of other": "<!-- docgen:begin evidence-grades -->\n<!-- docgen:end evidence-skill -->\n",
	} {
		if out, err := replaceBlock([]byte(in), block); err == nil {
			t.Errorf("%s: expected an error, got %q", name, out)
		}
	}
}

// TestEvidenceBlocksUpToDate fails when internal/method/evidence.json changed and `go run ./tools/docgen`
// was not run, or when a generated block or an embedded twin was edited by hand.
func TestEvidenceBlocksUpToDate(t *testing.T) {
	root := filepath.Join("..", "..")
	norm := func(b []byte) string { return strings.ReplaceAll(string(b), "\r\n", "\n") }
	for _, target := range evidenceTargets() {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target.path)))
		if err != nil {
			t.Fatal(err)
		}
		want, err := renderEvidenceBlocks(data, target.blocks)
		if err != nil {
			t.Fatalf("%s: %v", target.path, err)
		}
		if norm(want) != norm(data) {
			t.Errorf("%s: generated evidence block is out of date; run: go run ./tools/docgen", target.path)
		}
		if target.twin == "" {
			continue
		}
		twin, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target.twin)))
		if err != nil {
			t.Fatal(err)
		}
		if norm(twin) != norm(want) {
			t.Errorf("%s does not match %s; run: go run ./tools/docgen", target.twin, target.path)
		}
	}
}
