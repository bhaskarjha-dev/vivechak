package workspace

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	lastResolvedWorkspaceMu sync.RWMutex
	lastResolvedWorkspace   string
)

// SetLastResolvedWorkspace records the most recently resolved workspace root in memory.
func SetLastResolvedWorkspace(ws string) {
	lastResolvedWorkspaceMu.Lock()
	defer lastResolvedWorkspaceMu.Unlock()
	lastResolvedWorkspace = ws
}

// GetLastResolvedWorkspace returns the in-memory cached workspace root, if any.
func GetLastResolvedWorkspace() string {
	lastResolvedWorkspaceMu.RLock()
	defer lastResolvedWorkspaceMu.RUnlock()
	return lastResolvedWorkspace
}

// ResetLastResolvedWorkspace clears the in-memory cached workspace root (used in tests).
func ResetLastResolvedWorkspace() {
	lastResolvedWorkspaceMu.Lock()
	defer lastResolvedWorkspaceMu.Unlock()
	lastResolvedWorkspace = ""
}

// ProjectAnchorFiles lists manifest and anchor files identifying a software project root.
var ProjectAnchorFiles = []string{
	".git",
	"go.mod",
	"package.json",
	"Cargo.toml",
	"pyproject.toml",
	"requirements.txt",
	"pom.xml",
	"build.gradle",
	"CMakeLists.txt",
	"Makefile",
	"flake.nix",
}

// HasProjectAnchor reports whether the given directory contains a software project anchor.
func HasProjectAnchor(dir string) bool {
	for _, anchor := range ProjectAnchorFiles {
		target := filepath.Join(dir, anchor)
		if _, err := os.Stat(target); err == nil {
			return true
		}
	}
	return false
}

// ResolveWorkspace resolves the workspace root using the standard 7-step resolution ladder:
//  1. Explicit project_root argument
//  2. (Skipped in headless mode; see ResolveWorkspaceWithRoots)
//  3. Environment variables (VIVECHAK_PROJECT_ROOT -> VCK_PROJECT_ROOT -> VIVECHAK_DEFAULT_ROOT -> VCK_DEFAULT_ROOT)
//  4. Marker walk-up from CWD (checking research/.vivechak.json and research/)
//  5. Guarded CWD (project manifest anchor detection)
//  6. In-process session memory
//  7. Fail-closed with E-ROOT-UNRESOLVED
func ResolveWorkspace(explicit string) (string, error) {
	return ResolveWorkspaceWithRoots(context.Background(), explicit, nil)
}

// ResolveWorkspaceWithRoots resolves the workspace root evaluating MCP client roots at step 2.
func ResolveWorkspaceWithRoots(ctx context.Context, explicit string, roots []string) (string, error) {
	// Step 1: Explicit project_root argument
	trimmedExplicit := strings.TrimSpace(explicit)
	if trimmedExplicit != "" && !strings.HasPrefix(trimmedExplicit, "${") {
		abs, err := filepath.Abs(trimmedExplicit)
		if err != nil {
			return "", fmt.Errorf("resolving explicit path %q: %w", explicit, err)
		}
		// If explicit path points directly to research directory, normalize to parent workspace root
		if filepath.Base(abs) == ResearchDir {
			if _, err := os.Stat(filepath.Join(abs, ResearchDir)); os.IsNotExist(err) {
				if fi, err := os.Stat(abs); err == nil && fi.IsDir() {
					abs = filepath.Dir(abs)
				}
			}
		}
		if err := SafeRoot(abs); err != nil {
			return "", err
		}
		SetLastResolvedWorkspace(abs)
		return abs, nil
	}

	// Step 2: MCP Roots protocol
	for _, rootURI := range roots {
		parsedPath := parseRootURI(rootURI)
		if parsedPath != "" {
			abs, err := filepath.Abs(parsedPath)
			if err == nil {
				if err := SafeRoot(abs); err == nil {
					// Check if research marker or research dir exists
					marker := filepath.Join(abs, MetadataFile)
					research := filepath.Join(abs, ResearchDir)
					if _, err := os.Stat(marker); err == nil {
						SetLastResolvedWorkspace(abs)
						return abs, nil
					}
					if fi, err := os.Stat(research); err == nil && fi.IsDir() {
						SetLastResolvedWorkspace(abs)
						return abs, nil
					}
					// If anchor found or directory exists, record and return
					if HasProjectAnchor(abs) {
						SetLastResolvedWorkspace(abs)
						return abs, nil
					}
				}
			}
		}
	}

	// Step 3: Environment variables
	envLadder := []string{
		"VIVECHAK_PROJECT_ROOT",
		"VCK_PROJECT_ROOT",
		"VIVECHAK_DEFAULT_ROOT",
		"VCK_DEFAULT_ROOT",
	}
	for _, ev := range envLadder {
		val := strings.TrimSpace(os.Getenv(ev))
		if val != "" && !strings.HasPrefix(val, "${") {
			abs, err := filepath.Abs(val)
			if err == nil {
				if filepath.Base(abs) == ResearchDir {
					abs = filepath.Dir(abs)
				}
				if err := SafeRoot(abs); err == nil {
					SetLastResolvedWorkspace(abs)
					return abs, nil
				}
			}
		}
	}

	// Step 4: Walk up from CWD looking for research/.vivechak.json or research/ directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	dir := cwd
	for {
		if !IsSystemOrAppDir(dir) {
			markerPath := filepath.Join(dir, MetadataFile)
			if fi, err := os.Stat(markerPath); err == nil && !fi.IsDir() {
				if err := SafeRoot(dir); err == nil {
					SetLastResolvedWorkspace(dir)
					return dir, nil
				}
			}
			researchPath := filepath.Join(dir, ResearchDir)
			if info, err := os.Stat(researchPath); err == nil && info.IsDir() {
				if err := SafeRoot(dir); err == nil {
					SetLastResolvedWorkspace(dir)
					return dir, nil
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Step 5: Guarded CWD (project manifest anchor detection)
	if !IsSystemOrAppDir(cwd) && HasProjectAnchor(cwd) {
		if err := SafeRoot(cwd); err == nil {
			SetLastResolvedWorkspace(cwd)
			return cwd, nil
		}
	}

	// Step 6: Fall back to in-process session memory
	if cached := GetLastResolvedWorkspace(); cached != "" {
		if fi, err := os.Stat(cached); err == nil && fi.IsDir() {
			if err := SafeRoot(cached); err == nil {
				return cached, nil
			}
		}
	}

	// Step 7: Fail-closed with E-ROOT-UNRESOLVED
	return "", fmt.Errorf(
		"E-ROOT-UNRESOLVED: could not resolve workspace root (no explicit path, no MCP roots, "+
			"no VIVECHAK_PROJECT_ROOT, no research/ or project anchors walking up from %s, "+
			"and no prior workspace in session memory). "+
			"Start a research session in your AI editor or pass project_root explicitly", cwd)
}

func parseRootURI(uri string) string {
	trimmed := strings.TrimSpace(uri)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "file://") {
		u, err := url.Parse(trimmed)
		if err == nil {
			p := u.Path
			// On Windows, /C:/foo -> C:/foo
			if len(p) >= 3 && p[0] == '/' && p[2] == ':' {
				p = p[1:]
			}
			return filepath.FromSlash(p)
		}
	}
	return trimmed
}
