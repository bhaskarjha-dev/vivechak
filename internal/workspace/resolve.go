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

// Memory is the workspace one MCP session last resolved (rule "session"). Each MCP session has its own
// (see WithMemory); a resolution without one (CLI commands) remembers nothing. The zero value is ready to use.
type Memory struct {
	mu   sync.RWMutex
	root string
}

// Get returns the remembered workspace root, or "" (also for a nil *Memory).
func (m *Memory) Get() string {
	if m == nil {
		return ""
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.root
}

// Set remembers root; a nil *Memory ignores it.
func (m *Memory) Set(root string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.root = root
}

type memoryKey struct{}

// WithMemory returns a context whose workspace resolutions read and update m, the calling MCP session's memory.
func WithMemory(ctx context.Context, m *Memory) context.Context {
	return context.WithValue(ctx, memoryKey{}, m)
}

// memoryFrom returns the session memory carried by ctx, or nil.
func memoryFrom(ctx context.Context) *Memory {
	if ctx == nil {
		return nil
	}
	m, _ := ctx.Value(memoryKey{}).(*Memory)
	return m
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
// Rules returned: "explicit", "roots", "env", "marker", "cwd", "session". The "session" rule reads, and every
// successful resolution updates, the calling MCP session's memory carried by ctx (WithMemory); without one,
// nothing is remembered.
func ResolveWorkspaceWithRoots(ctx context.Context, explicit string, roots []string) (string, string, error) {
	mem := memoryFrom(ctx)
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

		mem.Set(abs)
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
		mem.Set(validRoots[0])
		return validRoots[0], "roots", nil
	} else if len(validRoots) > 1 {
		for _, vr := range validRoots {
			marker := filepath.Join(vr, MetadataFile)
			if _, err := os.Stat(marker); err == nil {
				mem.Set(vr)
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
					mem.Set(abs)
					return abs, "env", nil
				}
			}
		}
	}

	// Step 4: This MCP session's memory (rule: "session")
	if cached := mem.Get(); cached != "" {
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
				mem.Set(cleanDir)
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
		mem.Set(cwd)
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
