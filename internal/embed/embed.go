// Package embed provides access to Vivechak's embedded assets:
// generator prompts and output templates.
//
// Assets are embedded at compile time via go:embed, ensuring the server
// binary is fully self-contained with no external file dependencies.
package embed

import (
	"embed"
	"fmt"
	"io/fs"
)

// Generators embeds the three generator prompt files.
//
//go:embed generators/*.md
var Generators embed.FS

// Templates embeds the five output template files.
//
//go:embed templates/*.md
var Templates embed.FS

// ReadGenerator returns the contents of a generator file by name.
// Valid names: "GENERATOR.md", "GENERATOR-DECISION.md", "GENERATOR-COMPARISON.md"
func ReadGenerator(name string) ([]byte, error) {
	data, err := fs.ReadFile(Generators, "generators/"+name)
	if err != nil {
		return nil, fmt.Errorf("reading generator %q: %w", name, err)
	}
	return data, nil
}

// ReadTemplate returns the contents of a template file by name.
// Valid names: "DECISIONS.template.md", "CONFLICT-RESOLUTION.template.md",
// "COMPARISON-SESSION.template.md", "FOUNDING-ARCHITECTURE.template.md",
// "PHASE-0-GATE.template.md"
func ReadTemplate(name string) ([]byte, error) {
	data, err := fs.ReadFile(Templates, "templates/"+name)
	if err != nil {
		return nil, fmt.Errorf("reading template %q: %w", name, err)
	}
	return data, nil
}

// ListGenerators returns the names of all embedded generator files.
func ListGenerators() ([]string, error) {
	return listDir(Generators, "generators")
}

// ListTemplates returns the names of all embedded template files.
func ListTemplates() ([]string, error) {
	return listDir(Templates, "templates")
}

func listDir(fsys embed.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
