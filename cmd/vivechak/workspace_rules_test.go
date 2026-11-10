package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// startLiveMCPServerWithEnvAndRoots launches a live server subprocess with custom env and client roots.
func startLiveMCPServerWithEnvAndRoots(t *testing.T, ctx context.Context, binPath, dir string, customEnv []string, roots []string) (*mcp.ClientSession, func()) {
	t.Helper()
	cmd := exec.CommandContext(ctx, binPath, "serve")
	if dir != "" {
		cmd.Dir = dir
	}
	if len(customEnv) > 0 {
		cmd.Env = append(os.Environ(), customEnv...)
	}
	transport := &mcp.CommandTransport{Command: cmd}
	client := mcp.NewClient(&mcp.Implementation{Name: "vck-test-client", Version: "1.0.0"}, nil)
	for _, r := range roots {
		fileURI := "file://" + filepath.ToSlash(r)
		if !strings.HasPrefix(r, "/") && !strings.HasPrefix(fileURI, "file:///") {
			fileURI = "file:///" + filepath.ToSlash(r)
		}
		//nolint:staticcheck // SA1019: AddRoots tested specifically for protocol 2024-11-05 compatibility
		client.AddRoots(&mcp.Root{URI: fileURI})
	}

	opts := &mcp.ClientSessionOptions{
		ProtocolVersion: "2024-11-05",
	}
	cs, err := client.Connect(ctx, transport, opts)
	if err != nil {
		t.Fatalf("connecting to live MCP server: %v", err)
	}
	cleanup := func() { _ = cs.Close() }
	return cs, cleanup
}

func sameWorkspacePath(a, b string) bool {
	ca := filepath.Clean(a)
	cb := filepath.Clean(b)
	if strings.EqualFold(ca, cb) {
		return true
	}
	ra, errA := filepath.EvalSymlinks(ca)
	rb, errB := filepath.EvalSymlinks(cb)
	return errA == nil && errB == nil && strings.EqualFold(filepath.Clean(ra), filepath.Clean(rb))
}

// TestWorkspaceRules_LiveSubprocess tests workspace resolution rules:
// unnamespaced env vars ignored, unexpanded macros skipped, client roots used,
// and fresh sessions resolving client roots across next/gate/review.
func TestWorkspaceRules_LiveSubprocess(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)

	// 1. WORKSPACE and PROJECT_ROOT unnamespaced env vars are ignored
	t.Run("UnnamespacedEnvIgnored", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		base := t.TempDir()
		app := filepath.Join(base, "app")
		other := filepath.Join(base, "other")
		if err := os.MkdirAll(app, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(other, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(app, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		customEnv := []string{
			"WORKSPACE=" + other,
			"PROJECT_ROOT=" + other,
		}

		cs, cleanup := startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, app, customEnv, nil)
		defer cleanup()

		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_status",
		})
		if err != nil || res.IsError {
			t.Fatalf("vivechak_status failed: %v", err)
		}

		env := parseTestEnvelope(t, res)
		wsMap, ok := env.Data["workspace"].(map[string]any)
		if !ok || wsMap == nil {
			t.Fatalf("data.workspace missing in status reply: %+v", env.Data)
		}

		rootVal := wsMap["root"]
		cleanRoot := strings.TrimPrefix(strings.TrimPrefix(rootVal.(string), "file:///"), "file://")
		if !sameWorkspacePath(cleanRoot, app) {
			t.Errorf("expected workspace root %q, got %q", app, cleanRoot)
		}

		otherEntries, _ := os.ReadDir(other)
		if len(otherEntries) > 0 {
			t.Errorf("other directory should be untouched, found %d entries", len(otherEntries))
		}
	})

	// 2. Unexpanded ${workspaceFolder} macro in VIVECHAK_PROJECT_ROOT is skipped; VCK_PROJECT_ROOT is used
	t.Run("UnexpandedMacroSkipped", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		base := t.TempDir()
		app := filepath.Join(base, "app")
		neutral := filepath.Join(base, "neutral")
		if err := os.MkdirAll(app, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(neutral, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(app, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		customEnv := []string{
			"VIVECHAK_PROJECT_ROOT=${workspaceFolder}",
			"VCK_PROJECT_ROOT=" + app,
		}

		cs, cleanup := startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, neutral, customEnv, nil)
		defer cleanup()

		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_status",
		})
		if err != nil || res.IsError {
			t.Fatalf("vivechak_status failed: %v", err)
		}

		env := parseTestEnvelope(t, res)
		wsMap, ok := env.Data["workspace"].(map[string]any)
		if !ok || wsMap == nil {
			t.Fatalf("data.workspace missing in status reply: %+v", env.Data)
		}

		cleanRoot := strings.TrimPrefix(strings.TrimPrefix(wsMap["root"].(string), "file:///"), "file://")
		if !sameWorkspacePath(cleanRoot, app) {
			t.Errorf("expected workspace root %q, got %q", app, cleanRoot)
		}
		if wsMap["rule"] != "env" {
			t.Errorf("expected workspace rule 'env', got %v", wsMap["rule"])
		}
	})

	// 3. Single MCP root from client is used when in neutral folder
	t.Run("SingleMCPRootUsed", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		base := t.TempDir()
		app := filepath.Join(base, "app")
		neutral := filepath.Join(base, "neutral")
		if err := os.MkdirAll(app, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(neutral, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(app, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		cs, cleanup := startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, neutral, nil, []string{app})
		defer cleanup()

		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_status",
		})
		if err != nil || res.IsError {
			t.Fatalf("vivechak_status failed: %v", err)
		}

		env := parseTestEnvelope(t, res)
		wsMap, ok := env.Data["workspace"].(map[string]any)
		if !ok || wsMap == nil {
			t.Fatalf("data.workspace missing in status reply: %+v", env.Data)
		}

		cleanRoot := strings.TrimPrefix(strings.TrimPrefix(wsMap["root"].(string), "file:///"), "file://")
		if !sameWorkspacePath(cleanRoot, app) {
			t.Errorf("expected workspace root %q, got %q", app, cleanRoot)
		}
		if wsMap["rule"] != "roots" {
			t.Errorf("expected workspace rule 'roots', got %v", wsMap["rule"])
		}
	})

	// 4. Fresh session with client roots resolves next, gate, and review without E-ROOT-UNRESOLVED
	t.Run("FreshSessionClientRoot_NextGateReview", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		base := t.TempDir()
		app := filepath.Join(base, "app")
		neutral := filepath.Join(base, "neutral")
		if err := os.MkdirAll(app, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(neutral, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(app, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		// First, save a decision plan in app
		planContent := `# D-001-PLAN: Event Storage
| ID | Topic |
|---|---|
| D-001-S1 | Ingestion Engine |
`
		csSetup, cleanupSetup := startLiveMCPServer(t, ctx, binPath, app)
		_, err := csSetup.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "plan",
				"scope":   "decision",
				"id":      "D-001",
				"content": planContent,
			},
		})
		cleanupSetup()
		if err != nil {
			t.Fatalf("setup save plan failed: %v", err)
		}

		// Now launch a fresh session in neutral directory with client roots pointing to app
		csFresh, cleanupFresh := startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, neutral, nil, []string{app})
		defer cleanupFresh()

		toolsToTest := []struct {
			name string
			args map[string]any
		}{
			{name: "vivechak_next", args: map[string]any{}},
			{name: "vivechak_gate", args: map[string]any{}},
			{name: "vivechak_review", args: map[string]any{"mode": "validate"}},
		}

		for _, tc := range toolsToTest {
			t.Run(tc.name, func(t *testing.T) {
				callRes, callErr := csFresh.CallTool(ctx, &mcp.CallToolParams{
					Name:      tc.name,
					Arguments: tc.args,
				})
				if callErr != nil {
					t.Fatalf("%s call failed with transport error: %v", tc.name, callErr)
				}
				env := parseTestEnvelope(t, callRes)
				if strings.Contains(env.Message, "E-ROOT-UNRESOLVED") {
					t.Errorf("%s failed with E-ROOT-UNRESOLVED: %s", tc.name, env.Message)
				}
			})
		}
	})
}
