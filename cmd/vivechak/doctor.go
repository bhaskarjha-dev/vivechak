package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func runDoctor() {
	var explicit string
	if len(os.Args) > 2 {
		explicit = os.Args[2]
	}

	workspace, err := core.ResolveWorkspace(explicit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Workspace resolution failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Checking workspace at %s\n", workspace)
	hasErrors := false

	// 1. Workspace exists
	if !core.WorkspaceExists(workspace) {
		fmt.Fprintln(os.Stderr, "✗ Workspace (research/ directory) does not exist")
		hasErrors = true
	} else {
		fmt.Fprintln(os.Stderr, "✓ Workspace exists")
	}

	// 2. Templates complete
	templatesDir := filepath.Join(workspace, core.TemplatesDir)
	missingTemplates := []string{}
	for _, tpl := range core.TemplatesToCopy {
		if _, err := os.Stat(filepath.Join(templatesDir, tpl)); err != nil {
			missingTemplates = append(missingTemplates, tpl)
		}
	}
	if len(missingTemplates) > 0 {
		fmt.Fprintf(os.Stderr, "✗ Missing templates: %v\n", missingTemplates)
		hasErrors = true
	} else {
		fmt.Fprintln(os.Stderr, "✓ All templates present")
	}

	// Read pipeline and decisions for cross-checks
	pipelineBytes, _ := os.ReadFile(filepath.Join(workspace, core.PipelineFile))
	pipelineContent := string(pipelineBytes)
	
	decisionsBytes, _ := os.ReadFile(filepath.Join(workspace, core.DecisionsFile))
	decisionsContent := string(decisionsBytes)

	// 3. Frontmatter valid, 4. Orphaned sessions, 5. Stale cross-references, 6. No corruption
	sessionsDir := filepath.Join(workspace, core.SessionsDir)
	entries, err := os.ReadDir(sessionsDir)
	if err == nil {
		invalidFrontmatter := 0
		orphaned := 0
		staleRefs := 0
		corruptCount := 0

		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
				continue
			}
			path := filepath.Join(sessionsDir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			if !utf8.Valid(data) {
				fmt.Fprintf(os.Stderr, "✗ Corrupted UTF-8 in: %s\n", e.Name())
				corruptCount++
				hasErrors = true
			}

			fm, _, err := core.ParseFrontmatter(data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "✗ Invalid frontmatter in %s: %v\n", e.Name(), err)
				invalidFrontmatter++
				hasErrors = true
			}

			// Check orphaned
			if fm != nil && fm.Has("id") {
				id := fm.GetString("id")
				if id != "" && !strings.Contains(pipelineContent, id) {
					fmt.Fprintf(os.Stderr, "✗ Orphaned session (not in pipeline): %s\n", e.Name())
					orphaned++
					hasErrors = true
				}
			} else {
				// if no frontmatter id, fallback to filename
				id := strings.TrimSuffix(e.Name(), ".md")
				if !strings.Contains(pipelineContent, id) {
					fmt.Fprintf(os.Stderr, "✗ Orphaned session (not in pipeline): %s\n", e.Name())
					orphaned++
					hasErrors = true
				}
			}

			// Check stale decisions references
			re := regexp.MustCompile(`DECISION-\d+`)
			matches := re.FindAllString(string(data), -1)
			for _, match := range matches {
				if !strings.Contains(decisionsContent, match) {
					fmt.Fprintf(os.Stderr, "✗ Stale decision ref in %s: %s\n", e.Name(), match)
					staleRefs++
					hasErrors = true
				}
			}
		}

		if invalidFrontmatter == 0 {
			fmt.Fprintln(os.Stderr, "✓ All frontmatters valid")
		}
		if orphaned == 0 {
			fmt.Fprintln(os.Stderr, "✓ No orphaned sessions")
		}
		if staleRefs == 0 {
			fmt.Fprintln(os.Stderr, "✓ No stale cross-references")
		}
		if corruptCount == 0 {
			fmt.Fprintln(os.Stderr, "✓ No corrupted files (valid UTF-8)")
		}

	} else {
		// no sessions dir yet, which is fine
		fmt.Fprintln(os.Stderr, "✓ No frontmatters valid (no sessions)")
		fmt.Fprintln(os.Stderr, "✓ No orphaned sessions (no sessions)")
		fmt.Fprintln(os.Stderr, "✓ No stale cross-references (no sessions)")
		fmt.Fprintln(os.Stderr, "✓ No corrupted files (no sessions)")
	}

	if hasErrors {
		os.Exit(1)
	} else {
		fmt.Fprintln(os.Stderr, "Workspace is healthy.")
		os.Exit(0)
	}
}
