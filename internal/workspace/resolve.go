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

// ResolveWorkspace resolves the workspace root using the standard 7-step resolution ladder.
func ResolveWorkspace(explicit string) (string, error) {
	root, _, err := ResolveWorkspaceWithRule(explicit)
	return root, err
}

// ResolveWorkspaceWithRule resolves the workspace root and reports the rule that matched.
func ResolveWorkspaceWithRule(explicit string) (string, string, error) {
	return ResolveWorkspaceWithRoots(context.Background(), explicit, nil)
}

// ResolveWorkspaceWithRoots resolves the workspace root evaluating MCP client roots at step 2.
// Rules returned: "explicit", "roots", "env", "marker", "cwd", "session".
func ResolveWorkspaceWithRoots(ctx context.Context, explicit string, roots []string) (string, string, error) {
	trimmedExplicit := strings.TrimSpace(explicit)

	// Step 1: Explicit project_root argument (rule: "explicit")
	if trimmedExplicit != "" && !strings.Contains(trimmedExplicit, "${") {
		// Expand leading ~
		if trimmedExplicit == "~" {
			if home, err := os.UserHomeDir(); err == nil && home != "" {
				trimmedExplicit = home
			}
		} else if strings.HasPrefix(trimmedExplicit, "~/") || strings.HasPrefix(trimmedExplicit, `~\`) {
			if home, err := os.UserHomeDir(); err == nil && home != "" {
				trimmedExplicit = filepath.Join(home, trimmedExplicit[2:])
			}
		}
		// Expand $VAR and ${VAR}
		trimmedExplicit = os.ExpandEnv(trimmedExplicit)

		abs, err := filepath.Abs(trimmedExplicit)
		if err != nil {
			return "", "", fmt.Errorf("E-ROOT-INVALID: resolving explicit path %q: %w", explicit, err)
		}

		// Verify SafeRoot first
		if err := SafeRoot(abs); err != nil {
			return "", "", fmt.Errorf("E-ROOT-INVALID: %w", err)
		}

		// Check if explicit path exists and is a directory (never create it)
		fi, err := os.Stat(abs)
		if err != nil || !fi.IsDir() {
			return "", "", fmt.Errorf("E-ROOT-INVALID: explicit project_root %q does not exist or is not a directory", explicit)
		}

		// If explicit path points directly to research directory, normalize to parent workspace root
		if filepath.Base(abs) == ResearchDir {
			abs = filepath.Dir(abs)
		}

		SetLastResolvedWorkspace(abs)
		return abs, "explicit", nil
	}

	// Step 2: MCP Roots protocol (rule: "roots")
	var validRoots []string
	for _, rootURI := range roots {
		parsedPath := parseRootURI(rootURI)
		if parsedPath != "" {
			if abs, err := filepath.Abs(parsedPath); err == nil {
				if fi, err := os.Stat(abs); err == nil && fi.IsDir() && SafeRoot(abs) == nil {
					validRoots = append(validRoots, abs)
				}
			}
		}
	}
	if len(validRoots) == 1 {
		SetLastResolvedWorkspace(validRoots[0])
		return validRoots[0], "roots", nil
	} else if len(validRoots) > 1 {
		for _, vr := range validRoots {
			marker := filepath.Join(vr, MetadataFile)
			if _, err := os.Stat(marker); err == nil {
				SetLastResolvedWorkspace(vr)
				return vr, "roots", nil
			}
		}
		return "", "", fmt.Errorf("E-ROOT-AMBIGUOUS: multiple MCP client roots provided (%s); specify project_root explicitly", strings.Join(validRoots, ", "))
	}

	// Step 3: Environment variables (rule: "env")
	envLadder := []string{
		"VIVECHAK_PROJECT_ROOT",
		"VCK_PROJECT_ROOT",
	}
	for _, ev := range envLadder {
		val := strings.TrimSpace(os.Getenv(ev))
		if val != "" && !strings.Contains(val, "${") {
			if val == "~" {
				if home, err := os.UserHomeDir(); err == nil && home != "" {
					val = home
				}
			} else if strings.HasPrefix(val, "~/") || strings.HasPrefix(val, `~\`) {
				if home, err := os.UserHomeDir(); err == nil && home != "" {
					val = filepath.Join(home, val[2:])
				}
			}
			val = os.ExpandEnv(val)
			if abs, err := filepath.Abs(val); err == nil {
				if fi, err := os.Stat(abs); err == nil && fi.IsDir() && SafeRoot(abs) == nil {
					if filepath.Base(abs) == ResearchDir {
						abs = filepath.Dir(abs)
					}
					SetLastResolvedWorkspace(abs)
					return abs, "env", nil
				}
			}
		}
	}

	// Step 4: In-process session memory (rule: "session")
	if cached := GetLastResolvedWorkspace(); cached != "" {
		if fi, err := os.Stat(cached); err == nil && fi.IsDir() && SafeRoot(cached) == nil {
			return cached, "session", nil
		}
	}

	// Step 5: Marker walk-up from CWD (rule: "marker")
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("getting working directory: %w", err)
	}

	homeDir, _ := os.UserHomeDir()
	cleanHome := ""
	if homeDir != "" {
		cleanHome = filepath.Clean(homeDir)
	}

	dir := cwd
	for {
		cleanDir := filepath.Clean(dir)
		// Stop before reaching HOME or outside HOME
		if cleanHome != "" && (cleanDir == cleanHome || strings.EqualFold(cleanDir, cleanHome)) {
			break
		}
		if !IsSystemOrAppDir(cleanDir) && SafeRoot(cleanDir) == nil {
			markerPath := filepath.Join(cleanDir, MetadataFile)
			if fi, err := os.Stat(markerPath); err == nil && !fi.IsDir() {
				SetLastResolvedWorkspace(cleanDir)
				return cleanDir, "marker", nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Step 6: Guarded working directory (rule: "cwd")
	if !IsSystemOrAppDir(cwd) && HasProjectAnchor(cwd) && SafeRoot(cwd) == nil {
		SetLastResolvedWorkspace(cwd)
		return cwd, "cwd", nil
	}

	// Step 7: Fail-closed with E-ROOT-UNRESOLVED or E-ROOT-SYSTEM
	if IsSystemOrAppDir(cwd) {
		return "", "", fmt.Errorf("E-ROOT-SYSTEM: %q is an application or system directory, not a research workspace. Pass project_root=<absolute path of the user's project>", cwd)
	}

	return "", "", fmt.Errorf(
		"E-ROOT-UNRESOLVED: could not resolve workspace root (no explicit path, no MCP roots, "+
			"no VIVECHAK_PROJECT_ROOT, no research/ or project anchors walking up from %s, "+
			"and no prior workspace in session memory). "+
			"Pass project_root=<absolute path of the user's project>", cwd)
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
