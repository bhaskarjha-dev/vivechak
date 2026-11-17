package mcputil

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestWorkspaceMemoryIsPerSession checks HIGH-14's per-session memory: two MCP sessions on one server do not
// share the "session" rule. Session A anchors a project with project_root; session B, which never named a
// project, must not inherit A's workspace, while A's later calls still do.
func TestWorkspaceMemoryIsPerSession(t *testing.T) {
	ctx := context.Background()
	server := NewServer("0.1.0-test", slog.New(slog.NewTextHandler(io.Discard, nil)))
	connect := func() *mcp.ClientSession {
		st, ct := mcp.NewInMemoryTransports()
		if _, err := server.Connect(ctx, st, nil); err != nil {
			t.Fatalf("server connect: %v", err)
		}
		// No roots capability: resolution must come from project_root, session memory or the working directory.
		client := mcp.NewClient(&mcp.Implementation{Name: "session-test", Version: "0"}, &mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{}})
		cs, err := client.Connect(ctx, ct, nil)
		if err != nil {
			t.Fatalf("client connect: %v", err)
		}
		t.Cleanup(func() { _ = cs.Close() })
		return cs
	}
	status := func(cs *mcp.ClientSession) (map[string]any, bool) {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_status", Arguments: map[string]any{}})
		if err != nil {
			t.Fatalf("vivechak_status: %v", err)
		}
		data, _ := parseEnvelope(t, res).Data.(map[string]any)
		ws, _ := data["workspace"].(map[string]any)
		return ws, res.IsError
	}

	project := t.TempDir()
	a, b := connect(), connect()
	res, err := a.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_start", Arguments: map[string]any{
		"project_root": project, "scope": "comparison", "question": "Pino vs Winston",
	}})
	if err != nil || res.IsError {
		t.Fatalf("session A vivechak_start failed: %v %v", err, res)
	}

	if ws, _ := status(a); !strings.EqualFold(stringOf(ws["root"]), project) || ws["rule"] != "session" {
		t.Errorf("session A should remember %s (rule session), got %v", project, ws)
	}
	if ws, isErr := status(b); !isErr && strings.EqualFold(stringOf(ws["root"]), project) {
		t.Errorf("session B inherited session A's workspace %s (rule %v); session memory must be per MCP session", project, ws["rule"])
	}
}

func stringOf(v any) string {
	s, _ := v.(string)
	return s
}
