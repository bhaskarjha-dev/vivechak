package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

var decisionIDRe = regexp.MustCompile(`\bD-\d+\b`)

func runDoctor() {
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	os.Exit(runDoctorWithArgs(args, os.Stdout, os.Stderr))
}

func runDoctorWithArgs(args []string, stdout, stderr io.Writer) int {
	var explicit string
	if len(args) > 0 {
		if args[0] == "--help" || args[0] == "-h" {
			fmt.Fprintln(stdout, "Usage: vivechak doctor [workspace_path]  (or: vck doctor [workspace_path])")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "Check workspace integrity, templates, frontmatter validity, and cross-references.")
			return 0
		}
		explicit = args[0]
	}

	workspace, err := core.ResolveWorkspace(explicit)
	if err != nil {
		fmt.Fprintf(stderr, "Workspace resolution failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Checking workspace at %s\n", workspace)
	hasErrors, successes, errs := checkWorkspace(workspace)
	for _, s := range successes {
		fmt.Fprintln(stdout, s)
	}
	for _, e := range errs {
		fmt.Fprintln(stderr, e)
	}

	if hasErrors {
		return 1
	}
	fmt.Fprintln(stdout, "Workspace is healthy.")
	return 0
}

// checkWorkspace performs integrity and consistency checks on the given workspace path.
// Returns (hasErrors, successes, errors).
func checkWorkspace(workspace string) (bool, []string, []string) {
	hasErrors := false
	var successes []string
	var errs []string

	// 1. Workspace exists
	if !core.WorkspaceExists(workspace) {
		errs = append(errs, "✗ Workspace (research/ directory) does not exist")
		hasErrors = true
		return hasErrors, successes, errs
	}
	successes = append(successes, "✓ Workspace exists")

	// 2. Templates complete
	templatesDir := filepath.Join(workspace, core.TemplatesDir)
	var missingTemplates []string
	for _, tpl := range core.TemplatesToCopy {
		if _, err := os.Stat(filepath.Join(templatesDir, tpl)); err != nil {
			missingTemplates = append(missingTemplates, tpl)
		}
	}
	if len(missingTemplates) > 0 {
		errs = append(errs, fmt.Sprintf("✗ Missing templates: %v", missingTemplates))
		hasErrors = true
	} else {
		successes = append(successes, "✓ All templates present")
	}

	// Read pipeline or decision plan for session cross-checks
	var pipelineContent string
	if data, err := os.ReadFile(filepath.Join(workspace, core.PipelineFile)); err == nil {
		pipelineContent = string(data)
	} else if matches, err := filepath.Glob(filepath.Join(workspace, core.ResearchDir, "*-plan.md")); err == nil && len(matches) > 0 {
		if data, err := os.ReadFile(matches[0]); err == nil {
			pipelineContent = string(data)
		}
	}

	// Read decisions registry or ADRs for cross-checks
	var decisionsContent string
	if data, err := os.ReadFile(filepath.Join(workspace, core.DecisionsFile)); err == nil {
		decisionsContent = string(data)
	} else {
		var adrContents []string
		if entries, err := os.ReadDir(filepath.Join(workspace, core.ResearchDir)); err == nil {
			for _, e := range entries {
				name := e.Name()
				if !e.IsDir() && strings.HasSuffix(name, ".md") &&
					!strings.HasSuffix(name, "-plan.md") &&
					!strings.EqualFold(name, "RESEARCH-PIPELINE.md") &&
					!strings.EqualFold(name, "FAD.md") &&
					!strings.EqualFold(name, "DECISIONS.md") {
					if data, err := os.ReadFile(filepath.Join(workspace, core.ResearchDir, name)); err == nil {
						adrContents = append(adrContents, string(data))
					}
				}
			}
		}
		if len(adrContents) > 0 {
			decisionsContent = strings.Join(adrContents, "\n")
		}
	}

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
				errs = append(errs, fmt.Sprintf("✗ Corrupted UTF-8 in: %s", e.Name()))
				corruptCount++
				hasErrors = true
			}

			fm, _, err := core.ParseFrontmatter(data)
			if err != nil {
				errs = append(errs, fmt.Sprintf("✗ Invalid frontmatter in %s: %v", e.Name(), err))
				invalidFrontmatter++
				hasErrors = true
			}

			// Check orphaned
			id := core.ExtractSessionID(e.Name(), fm)

			if pipelineContent != "" && id != "" && !strings.Contains(pipelineContent, id) {
				errs = append(errs, fmt.Sprintf("✗ Orphaned session (not in pipeline): %s", e.Name()))
				orphaned++
				hasErrors = true
			}

			// Check stale decisions references
			if decisionsContent != "" {
				matches := decisionIDRe.FindAllString(string(data), -1)
				seenMatches := make(map[string]bool)
				for _, match := range matches {
					if seenMatches[match] {
						continue
					}
					seenMatches[match] = true
					if !strings.Contains(decisionsContent, match) {
						errs = append(errs, fmt.Sprintf("✗ Stale decision ref in %s: %s", e.Name(), match))
						staleRefs++
						hasErrors = true
					}
				}
			}
		}

		if invalidFrontmatter == 0 {
			successes = append(successes, "✓ All frontmatters valid")
		}
		if pipelineContent != "" {
			if orphaned == 0 {
				successes = append(successes, "✓ No orphaned sessions")
			}
		} else {
			successes = append(successes, "✓ Orphan check skipped (pipeline not applicable)")
		}
		if decisionsContent != "" {
			if staleRefs == 0 {
				successes = append(successes, "✓ No stale cross-references")
			}
		} else {
			successes = append(successes, "✓ Cross-reference check skipped (no decision registry)")
		}
		if corruptCount == 0 {
			successes = append(successes, "✓ No corrupted files (valid UTF-8)")
		}

	} else {
		successes = append(successes, "✓ Frontmatter check skipped (no sessions)")
		successes = append(successes, "✓ Orphan check skipped (no sessions)")
		successes = append(successes, "✓ Cross-reference check skipped (no sessions)")
		successes = append(successes, "✓ Integrity check skipped (no sessions)")
	}

	return hasErrors, successes, errs
}
