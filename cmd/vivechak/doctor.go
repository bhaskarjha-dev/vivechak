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

var decisionRefRe = regexp.MustCompile(`(?i)\[(D-[A-Za-z0-9_-]+)\]|\b(?:ADR|decision|informs|refers?\s+to)\s+(D-[A-Za-z0-9_-]+)\b`)

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
	info := core.InspectWorkspace(workspace)
	expectedTemplates := core.TemplatesForScope(info.Scope)
	var missingTemplates []string
	for _, tpl := range expectedTemplates {
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

	// Read pipeline or decision plans for session cross-checks
	var pipelineChunks []string
	if data, err := os.ReadFile(filepath.Join(workspace, core.PipelineFile)); err == nil {
		pipelineChunks = append(pipelineChunks, string(data))
	}
	if matches, err := filepath.Glob(filepath.Join(workspace, core.ResearchDir, "*-plan.md")); err == nil {
		for _, m := range matches {
			if data, err := os.ReadFile(m); err == nil {
				pipelineChunks = append(pipelineChunks, string(data))
			}
		}
	}
	if matches, err := filepath.Glob(filepath.Join(workspace, core.ResearchDir, "*-comparison.md")); err == nil {
		for _, m := range matches {
			if data, err := os.ReadFile(m); err == nil {
				pipelineChunks = append(pipelineChunks, string(data))
			}
		}
	}
	var pipelineContent string
	if len(pipelineChunks) > 0 {
		pipelineContent = strings.Join(pipelineChunks, "\n\n")
	}

	// Read decisions registry and/or standalone ADRs for cross-checks
	var decisionsContent string
	var adrContents []string
	if data, err := os.ReadFile(filepath.Join(workspace, core.DecisionsFile)); err == nil {
		adrContents = append(adrContents, string(data))
	}
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

			if pipelineContent != "" && id != "" {
				inPipeline := strings.Contains(pipelineContent, id)
				if !inPipeline {
					// Check progressively shorter hyphen-delimited prefixes for slugged filenames
					// e.g. "D-001-S1-slug" -> "D-001-S1", "T1-01-database-selection" -> "T1-01", "S1-database" -> "S1"
					parts := strings.Split(id, "-")
					for k := len(parts) - 1; k >= 1; k-- {
						prefix := strings.Join(parts[:k], "-")
						if strings.Contains(pipelineContent, prefix) {
							inPipeline = true
							break
						}
					}
				}
				if !inPipeline {
					errs = append(errs, fmt.Sprintf("✗ Orphaned session (not in pipeline): %s", e.Name()))
					orphaned++
					hasErrors = true
				}
			}

			// Check stale decisions references from structured frontmatter and explicit references
			if decisionsContent != "" {
				var refs []string
				if fm != nil {
					refs = append(refs, fm.GetStringSlice("informs_decisions")...)
					refs = append(refs, fm.GetStringSlice("decisions")...)
					if s := fm.GetString("decision"); s != "" {
						refs = append(refs, s)
					}
					if s := fm.GetString("informs_decision"); s != "" {
						refs = append(refs, s)
					}
				}
				for _, match := range decisionRefRe.FindAllStringSubmatch(string(data), -1) {
					for k := 1; k < len(match); k++ {
						if match[k] != "" {
							refs = append(refs, match[k])
						}
					}
				}

				seenMatches := make(map[string]bool)
				for _, match := range refs {
					match = strings.TrimSpace(match)
					if match == "" || seenMatches[match] {
						continue
					}
					seenMatches[match] = true
					if !strings.Contains(decisionsContent, match) {
						if pipelineContent != "" && strings.Contains(pipelineContent, match) {
							successes = append(successes, fmt.Sprintf("ℹ In-flight planned decision in %s: %s (declared in pipeline, awaiting ADR)", e.Name(), match))
						} else {
							errs = append(errs, fmt.Sprintf("✗ Stale decision ref in %s: %s", e.Name(), match))
							staleRefs++
							hasErrors = true
						}
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
