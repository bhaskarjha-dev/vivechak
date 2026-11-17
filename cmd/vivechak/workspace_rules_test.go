package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// rootURI turns a directory into the file:// URI an MCP client advertises as a root.
func rootURI(dir string) string {
	uri := "file://" + filepath.ToSlash(dir)
	if !strings.HasPrefix(dir, "/") && !strings.HasPrefix(uri, "file:///") {
		uri = "file:///" + filepath.ToSlash(dir)
	}
	return uri
}

// rawMCPClient speaks newline-delimited JSON-RPC to a live `vck serve`, so a test controls the client side
// completely: whether it advertises the roots capability, whether it answers the server's roots/list requests at
// all, and how many of them the server sends (HIGH-14).
type rawMCPClient struct {
	t             *testing.T
	stdin         io.WriteCloser
	lines         chan []byte
	nextID        int
	roots         []string
	answerRoots   bool
	rootsRequests int
}

func startRawMCPClient(t *testing.T, binPath, dir string, env []string, advertiseRoots, answerRoots bool, roots []string) *rawMCPClient {
	t.Helper()
	cmd := exec.Command(binPath, "serve")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting vck serve: %v", err)
	}
	c := &rawMCPClient{t: t, stdin: stdin, lines: make(chan []byte, 64), roots: roots, answerRoots: answerRoots}
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 1<<20), 16<<20)
		for sc.Scan() {
			c.lines <- append([]byte(nil), sc.Bytes()...)
		}
		close(c.lines)
	}()
	t.Cleanup(func() {
		_ = stdin.Close()
		done := make(chan struct{})
		go func() { _ = cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
	})
	caps := map[string]any{}
	if advertiseRoots {
		caps["roots"] = map[string]any{"listChanged": true}
	}
	c.request("initialize", map[string]any{
		"protocolVersion": "2025-06-18",
		"capabilities":    caps,
		"clientInfo":      map[string]any{"name": "raw-test-client", "version": "0"},
	})
	c.send(map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	return c
}

func (c *rawMCPClient) send(msg any) {
	b, err := json.Marshal(msg)
	if err != nil {
		c.t.Fatal(err)
	}
	if _, err := c.stdin.Write(append(b, '\n')); err != nil {
		c.t.Fatalf("writing to server: %v", err)
	}
}

// request sends one request and returns its result, handling the server's own requests meanwhile.
func (c *rawMCPClient) request(method string, params any) json.RawMessage {
	c.t.Helper()
	c.nextID++
	id := strconv.Itoa(c.nextID)
	c.send(map[string]any{"jsonrpc": "2.0", "id": c.nextID, "method": method, "params": params})
	deadline := time.After(20 * time.Second)
	for {
		select {
		case line, ok := <-c.lines:
			if !ok {
				c.t.Fatalf("server closed stdout while waiting for %s", method)
			}
			var m struct {
				ID     json.RawMessage `json:"id"`
				Method string          `json:"method"`
				Result json.RawMessage `json:"result"`
				Error  json.RawMessage `json:"error"`
			}
			if json.Unmarshal(line, &m) != nil {
				continue
			}
			if m.Method != "" {
				if m.Method == "roots/list" && len(m.ID) > 0 {
					c.rootsRequests++
					if c.answerRoots {
						var list []map[string]any
						for _, r := range c.roots {
							list = append(list, map[string]any{"uri": rootURI(r), "name": filepath.Base(r)})
						}
						c.send(map[string]any{"jsonrpc": "2.0", "id": m.ID, "result": map[string]any{"roots": list}})
					}
				}
				continue
			}
			if string(m.ID) == id {
				if len(m.Error) > 0 {
					c.t.Fatalf("%s failed: %s", method, m.Error)
				}
				return m.Result
			}
		case <-deadline:
			c.t.Fatalf("no reply to %s", method)
		}
	}
}

// callTool calls a tool and returns its envelope.
func (c *rawMCPClient) callTool(name string, args map[string]any) testEnvelope {
	c.t.Helper()
	var res struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(c.request("tools/call", map[string]any{"name": name, "arguments": args}), &res); err != nil || len(res.Content) == 0 {
		c.t.Fatalf("%s: unreadable result (%v)", name, err)
	}
	var env testEnvelope
	if err := json.Unmarshal([]byte(res.Content[0].Text), &env); err != nil {
		c.t.Fatalf("%s: unreadable envelope: %v", name, err)
	}
	return env
}

// isolatedEnv pins HOME inside base (the marker walk-up stops there) and clears the workspace variables.
func isolatedEnv(base string) []string {
	return []string{"HOME=" + base, "USERPROFILE=" + base, "VIVECHAK_PROJECT_ROOT=", "VCK_PROJECT_ROOT="}
}

func treeSnapshot(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			rel, _ := filepath.Rel(dir, path)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	return out
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
		planContent := livePlan("D-001-PLAN: Event Storage", "D-001", "", []string{
			"{id: D-001-S1, title: Ingestion Engine, role: C, depends: []}",
		}, liveSection("D-001-S1", "Ingestion Engine", "Compare ingestion engines."))
		csSetup, cleanupSetup := startLiveMCPServer(t, ctx, binPath, app)
		setupRes, err := csSetup.CallTool(ctx, &mcp.CallToolParams{
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
		if setupRes.IsError {
			t.Fatalf("setup save plan refused: %s", parseTestEnvelope(t, setupRes).Message)
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
				// The gate reports the workspace it judged: the client's root, by rule roots.
				if tc.name == "vivechak_gate" {
					wsMap, _ := env.Data["workspace"].(map[string]any)
					root, _ := wsMap["root"].(string)
					if !sameWorkspacePath(root, app) || wsMap["rule"] != "roots" {
						t.Errorf("vivechak_gate: expected the client's root %s by rule roots, got %v (%s)", app, wsMap, env.Message)
					}
				}
			})
		}
		if entries, _ := os.ReadDir(neutral); len(entries) != 0 {
			t.Errorf("the server's working directory was written to: %v", treeSnapshot(t, neutral))
		}
	})
	// 5. Started in a subfolder with no project anchor and no marker, vivechak_start refuses with
	// E-ROOT-UNRESOLVED naming project_root, and writes nothing (no looser cwd fallback; HIGH-14).
	t.Run("SubfolderWithoutAnchorOrMarker_StartRefusesAndWritesNothing", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		base := t.TempDir()
		proj := filepath.Join(base, "proj")
		sub := filepath.Join(proj, "internal", "sub")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, "go.mod"), []byte("module example.com/proj\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		before := treeSnapshot(t, base)

		cs, cleanup := startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, sub, isolatedEnv(base), nil)
		defer cleanup()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_start",
			Arguments: map[string]any{"scope": "comparison", "question": "A vs B"},
		})
		if err != nil {
			t.Fatalf("vivechak_start transport error: %v", err)
		}
		env := parseTestEnvelope(t, res)
		if !res.IsError || !strings.Contains(env.Message, "E-ROOT-UNRESOLVED") || !strings.Contains(env.Message+env.Guidance(), "project_root") {
			t.Errorf("expected E-ROOT-UNRESOLVED naming project_root, got isError=%v %q / %q", res.IsError, env.Message, env.Guidance())
		}
		if after := treeSnapshot(t, base); strings.Join(after, "\n") != strings.Join(before, "\n") {
			t.Errorf("vivechak_start wrote files: before %v, after %v", before, after)
		}
	})

	// 6. roots/list goes only to clients that advertise the roots capability, at most once per session, and again
	// after notifications/roots/list_changed; a client that never answers costs one timeout, not one per call.
	t.Run("RootsAskedOnlyWhenAdvertised_OncePerSession", func(t *testing.T) {
		base := t.TempDir()
		app := filepath.Join(base, "app")
		neutral := filepath.Join(base, "neutral")
		for _, d := range []string{app, neutral} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(app, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		noCaps := startRawMCPClient(t, binPath, app, isolatedEnv(base), false, false, nil)
		for i := 0; i < 3; i++ {
			if env := noCaps.callTool("vivechak_status", map[string]any{}); !env.Ok {
				t.Fatalf("status without roots capability failed: %s", env.Message)
			}
		}
		if noCaps.rootsRequests != 0 {
			t.Errorf("a client without the roots capability got %d roots/list requests, want 0", noCaps.rootsRequests)
		}

		withRoots := startRawMCPClient(t, binPath, neutral, isolatedEnv(base), true, true, []string{app})
		for i := 0; i < 3; i++ {
			env := withRoots.callTool("vivechak_status", map[string]any{})
			wsMap, _ := env.Data["workspace"].(map[string]any)
			root, _ := wsMap["root"].(string)
			if !sameWorkspacePath(root, app) || wsMap["rule"] != "roots" {
				t.Fatalf("call %d: expected the client's root %s (rule roots), got %v", i+1, app, wsMap)
			}
		}
		if withRoots.rootsRequests != 1 {
			t.Errorf("a client advertising roots got %d roots/list requests for 3 calls, want 1", withRoots.rootsRequests)
		}
		withRoots.send(map[string]any{"jsonrpc": "2.0", "method": "notifications/roots/list_changed"})
		withRoots.callTool("vivechak_status", map[string]any{})
		if withRoots.rootsRequests != 2 {
			t.Errorf("after notifications/roots/list_changed expected one more roots/list request (2 in total), got %d", withRoots.rootsRequests)
		}

		silent := startRawMCPClient(t, binPath, app, isolatedEnv(base), true, false, nil)
		start := time.Now()
		for i := 0; i < 10; i++ {
			if env := silent.callTool("vivechak_status", map[string]any{}); !env.Ok {
				t.Fatalf("status from a client that never answers roots/list failed: %s", env.Message)
			}
		}
		t.Logf("10 vivechak_status calls from a client that never answers roots/list: %s", time.Since(start))
		if silent.rootsRequests != 1 {
			t.Errorf("a client that never answers got %d roots/list requests for 10 calls, want 1", silent.rootsRequests)
		}
	})

	// 7. No walk-up to a project anchor (round 2's regression): from a subfolder of a project that has an anchor
	// (go.mod) but no research/ marker, no workspace resolves (vivechak_gate and vivechak_next: E-ROOT-UNRESOLVED;
	// vivechak_status: no workspace) and nothing is written. Once the project holds research/.vivechak.json, the same
	// subfolder resolves to it by the marker rule.
	t.Run("NoAnchorWalkUp_MarkerWalkUpResolves", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		base := t.TempDir()
		proj := filepath.Join(base, "proj")
		sub := filepath.Join(proj, "internal", "sub")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(proj, "go.mod"), []byte("module example.com/proj\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		before := treeSnapshot(t, base)

		cs, cleanup := startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, sub, isolatedEnv(base), nil)
		for _, tool := range []string{"vivechak_gate", "vivechak_next"} {
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: map[string]any{}})
			if err != nil {
				t.Fatalf("%s transport error: %v", tool, err)
			}
			if env := parseTestEnvelope(t, res); !res.IsError || !strings.Contains(env.Message, "E-ROOT-UNRESOLVED") {
				t.Errorf("%s from a subfolder without a marker: expected E-ROOT-UNRESOLVED (no walk-up to go.mod), got isError=%v %q", tool, res.IsError, env.Message)
			}
		}
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_status", Arguments: map[string]any{}})
		cleanup()
		if err != nil {
			t.Fatalf("vivechak_status transport error: %v", err)
		}
		if env := parseTestEnvelope(t, res); !res.IsError || !strings.Contains(env.Message, "E-ROOT-UNRESOLVED") {
			t.Errorf("vivechak_status from a subfolder without a marker: expected E-ROOT-UNRESOLVED, got isError=%v %q %v", res.IsError, env.Message, env.Data)
		}
		if after := treeSnapshot(t, base); strings.Join(after, "\n") != strings.Join(before, "\n") {
			t.Errorf("the calls wrote files: before %v, after %v", before, after)
		}

		// Initialize the project from its own root (anchor in the working directory), then retry from the subfolder.
		csProj, cleanupProj := startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, proj, isolatedEnv(base), nil)
		resStart, err := csProj.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_start", Arguments: map[string]any{"scope": "comparison", "question": "Pino vs Winston"}})
		cleanupProj()
		if err != nil || resStart.IsError {
			t.Fatalf("vivechak_start in the project root failed: %v %s", err, parseTestEnvelope(t, resStart).Message)
		}
		csSub, cleanupSub := startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, sub, isolatedEnv(base), nil)
		defer cleanupSub()
		res, err = csSub.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_status", Arguments: map[string]any{}})
		if err != nil || res.IsError {
			t.Fatalf("vivechak_status from the subfolder of an initialized project failed: %v %s", err, parseTestEnvelope(t, res).Message)
		}
		wsMap, _ := parseTestEnvelope(t, res).Data["workspace"].(map[string]any)
		root, _ := wsMap["root"].(string)
		if !sameWorkspacePath(root, proj) || wsMap["rule"] != "marker" {
			t.Errorf("expected the project root %s by rule marker, got %v", proj, wsMap)
		}
	})

	// 8. An explicit project_root that does not exist is refused (E-ROOT-INVALID; vivechak_status reports no
	// workspace) and never created.
	t.Run("NonexistentProjectRoot_RefusedAndNotCreated", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		base := t.TempDir()
		app := filepath.Join(base, "app")
		if err := os.MkdirAll(app, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(app, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		missing := filepath.Join(base, "missing", "project")
		cs, cleanup := startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, app, isolatedEnv(base), nil)
		defer cleanup()

		for _, call := range []struct {
			tool string
			args map[string]any
		}{
			{"vivechak_start", map[string]any{"project_root": missing, "scope": "comparison", "question": "Pino vs Winston"}},
			{"vivechak_save", map[string]any{"project_root": missing, "kind": "comparison", "content": wgComparison("Pino vs Winston", "Pino", "Winston")}},
			{"vivechak_gate", map[string]any{"project_root": missing}},
		} {
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: call.tool, Arguments: call.args})
			if err != nil {
				t.Fatalf("%s transport error: %v", call.tool, err)
			}
			if env := parseTestEnvelope(t, res); !res.IsError || !strings.Contains(env.Message, "E-ROOT-INVALID") {
				t.Errorf("%s with a nonexistent project_root: expected E-ROOT-INVALID, got isError=%v %q", call.tool, res.IsError, env.Message)
			}
		}
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_status", Arguments: map[string]any{"project_root": missing}})
		if err != nil {
			t.Fatalf("vivechak_status transport error: %v", err)
		}
		if env := parseTestEnvelope(t, res); !res.IsError || !strings.Contains(env.Message, "E-ROOT-INVALID") {
			t.Errorf("vivechak_status with a nonexistent project_root: expected E-ROOT-INVALID, got isError=%v %q %v", res.IsError, env.Message, env.Data)
		}
		if _, err := os.Stat(filepath.Join(base, "missing")); !os.IsNotExist(err) {
			t.Errorf("a nonexistent project_root was created (stat err=%v): %v", err, treeSnapshot(t, base))
		}
		if entries, _ := os.ReadDir(app); len(entries) != 1 {
			t.Errorf("the server's working directory was written to: %v", treeSnapshot(t, app))
		}
	})

	// 9. vivechak_record_decision, vivechak_amend and vivechak_replan resolve the workspace from the client's root like
	// every other tool: each is the first call of a fresh session started in a folder that is not a project, and each
	// works on the root's research/ (rule roots), writing nothing in the server's working directory.
	t.Run("ClientRootUsedByRecordDecisionAmendReplan", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		base := t.TempDir()
		app := filepath.Join(base, "app")
		neutral := filepath.Join(base, "neutral")
		for _, d := range []string{app, neutral} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(app, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		csSetup, cleanupSetup := startLiveMCPServer(t, ctx, binPath, app)
		for _, args := range []map[string]any{
			{"project_root": app, "kind": "plan", "id": "D-001", "content": wgDecisionPlan("D-001", "Primary datastore")},
			{"project_root": app, "kind": "session", "id": "D-001-S1", "content": wgSession("D-001-S1", "D-001", "Landscape")},
		} {
			if env, isErr := wgCall(t, ctx, csSetup, "vivechak_save", args); isErr {
				t.Fatalf("setup save %v failed: %s", args["id"], env.Message)
			}
		}
		cleanupSetup()

		fresh := func() (*mcp.ClientSession, func()) {
			return startLiveMCPServerWithEnvAndRoots(t, ctx, binPath, neutral, isolatedEnv(base), []string{app})
		}

		cs, cleanup := fresh()
		rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{
			"id":      "D-001",
			"content": wgADR("D-001", "Primary datastore", "proposed", "two-way", "D-001-S1"),
		})
		cleanup()
		wsMap, _ := rec.Data["workspace"].(map[string]any)
		root, _ := wsMap["root"].(string)
		if isErr || !sameWorkspacePath(root, app) || wsMap["rule"] != "roots" {
			t.Errorf("vivechak_record_decision: expected the client's root %s by rule roots, got isError=%v workspace=%v (%s)", app, isErr, wsMap, rec.Message)
		}
		if _, err := os.Stat(filepath.Join(app, "research", "D-001-decision.md")); err != nil {
			t.Errorf("vivechak_record_decision did not write the ADR in the client's root: %v", err)
		}

		cs, cleanup = fresh()
		const amendment = "Load test confirmed the write-amplification concern at 5k writes/s."
		amend, isErr := wgCall(t, ctx, cs, "vivechak_amend", map[string]any{"id": "D-001-S1", "note": amendment})
		cleanup()
		sess, _ := os.ReadFile(filepath.Join(app, "research", "sessions", "D-001-S1.md"))
		if isErr || !strings.Contains(string(sess), amendment) {
			t.Errorf("vivechak_amend: expected the amendment in the client's root (%s), got isError=%v (%s)", filepath.Join(app, "research", "sessions", "D-001-S1.md"), isErr, amend.Message)
		}

		cs, cleanup = fresh()
		const prompt = "Find failure modes, reversal triggers and exit costs of the leading datastore."
		replan, isErr := wgCall(t, ctx, cs, "vivechak_replan", map[string]any{"plan": "D-001", "op": "update_prompt", "session_id": "D-001-S2", "prompt": prompt})
		cleanup()
		plan, _ := os.ReadFile(filepath.Join(app, "research", "D-001-plan.md"))
		if isErr || !strings.Contains(string(plan), prompt) {
			t.Errorf("vivechak_replan: expected the new prompt in the client's root plan, got isError=%v (%s)", isErr, replan.Message)
		}

		if entries, _ := os.ReadDir(neutral); len(entries) != 0 {
			t.Errorf("the server's working directory was written to: %v", treeSnapshot(t, neutral))
		}
	})
}
