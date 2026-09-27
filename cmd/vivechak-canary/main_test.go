package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newTestServer creates a canary server wired to in-memory transports.
// Returns (server, clientSession) — the client can call tools on the server.
func newTestServer(t *testing.T) (*mcp.Server, *mcp.ClientSession) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	ctx := context.Background()

	server := mcp.NewServer(
		&mcp.Implementation{Name: "vivechak", Version: "0.0.1-canary-test"},
		&mcp.ServerOptions{
			Instructions: "Test canary server",
			Logger:       logger,
		},
	)

	// Register the same tools as main.go
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "vivechak_ping",
			Title:       "Vivechak Ping",
			Description: "Check Vivechak server connectivity and version.",
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: boolPtr(false),
				OpenWorldHint:   boolPtr(false),
			},
		},
		func(ctx context.Context, req *mcp.CallToolRequest, in PingInput) (*mcp.CallToolResult, Envelope, error) {
			env := Envelope{
				Success: true,
				Message: "Vivechak canary server v0.0.1-canary-test is running",
				Data: map[string]any{
					"version": "0.0.1-canary-test",
					"canary":  true,
				},
				NextStep: "Run vivechak_echo with test text.",
				Meta:     &EnvMeta{API: 1, Tool: "vivechak_ping"},
			}
			return envelopeToResult(env)
		},
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "vivechak_echo",
			Title:       "Vivechak Echo",
			Description: "Echo back provided text in the standardized response envelope.",
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  true,
				DestructiveHint: boolPtr(false),
				OpenWorldHint:   boolPtr(false),
			},
		},
		func(ctx context.Context, req *mcp.CallToolRequest, in EchoInput) (*mcp.CallToolResult, Envelope, error) {
			count := in.RepeatCount
			if count <= 0 {
				count = 1
			}
			if count > 500 {
				count = 500
			}
			workspace := resolveWorkspace(in.ProjectRoot)
			var text string
			if count == 1 {
				text = in.Text
			} else {
				text = strings.Repeat(in.Text+"\n", count)
			}

			// Same output-size warning logic as main.go
			approxTokens := len(text) / 4
			var warnings []string
			truncated := false

			if approxTokens > 25000 {
				warnings = append(warnings, "W-OUTPUT-SIZE: Response exceeds 25K token hard cap threshold")
				truncated = true
				if len(text) > 100000 {
					text = text[:100000] + "\n... [TRUNCATED]"
				}
			} else if approxTokens > 10000 {
				warnings = append(warnings, "W-OUTPUT-SIZE: Response exceeds 10K token warning threshold")
			}

			env := Envelope{
				Success: true,
				Message: "Echo complete",
				Data: map[string]any{
					"echo_text":      text,
					"repeat_count":   count,
					"byte_length":    len(text),
					"approx_tokens":  approxTokens,
					"workspace_root": workspace,
				},
				Warnings: warnings,
				NextStep: "Verify response.",
				Meta:     &EnvMeta{API: 1, Tool: "vivechak_echo", Truncated: truncated},
			}
			return envelopeToResult(env)
		},
	)

	// Connect via in-memory transports
	st, ct := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "test-client", Version: "1.0.0"},
		nil,
	)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	return server, cs
}

// --- Probe 1: Tool Listing ---
func TestProbe_ToolListing(t *testing.T) {
	_, cs := newTestServer(t)
	ctx := context.Background()

	var tools []*mcp.Tool
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		tools = append(tools, tool)
	}

	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name] = true
	}

	if !names["vivechak_ping"] {
		t.Error("missing vivechak_ping in tool listing")
	}
	if !names["vivechak_echo"] {
		t.Error("missing vivechak_echo in tool listing")
	}
}

// --- Probe 2: Annotations ---
func TestProbe_Annotations(t *testing.T) {
	_, cs := newTestServer(t)
	ctx := context.Background()

	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}

		if tool.Annotations == nil {
			t.Errorf("%s: annotations are nil", tool.Name)
			continue
		}

		switch tool.Name {
		case "vivechak_ping":
			if !tool.Annotations.ReadOnlyHint {
				t.Errorf("vivechak_ping: expected readOnlyHint=true")
			}
			if !tool.Annotations.IdempotentHint {
				t.Errorf("vivechak_ping: expected idempotentHint=true")
			}
		case "vivechak_echo":
			if tool.Annotations.ReadOnlyHint {
				t.Errorf("vivechak_echo: expected readOnlyHint=false")
			}
		}
	}
}

// --- Probe 3: Response Channels (structured + text content) ---
func TestProbe_ResponseChannels(t *testing.T) {
	_, cs := newTestServer(t)
	ctx := context.Background()

	// Call vivechak_ping
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_ping"})
	if err != nil {
		t.Fatalf("calling vivechak_ping: %v", err)
	}

	// Check text content channel
	if len(result.Content) == 0 {
		t.Fatal("vivechak_ping returned no text content")
	}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	// The text content should be a valid JSON envelope
	var textEnvelope Envelope
	if err := json.Unmarshal([]byte(textContent.Text), &textEnvelope); err != nil {
		t.Fatalf("text content is not valid JSON envelope: %v", err)
	}
	if !textEnvelope.Success {
		t.Error("text envelope: success should be true")
	}
	if textEnvelope.Message == "" {
		t.Error("text envelope: message should not be empty")
	}

	// Check structured content channel
	if result.StructuredContent == nil {
		t.Fatal("vivechak_ping returned no structured content")
	}

	// Marshal structured content and verify it's a valid envelope
	scBytes, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshaling structured content: %v", err)
	}
	var scEnvelope Envelope
	if err := json.Unmarshal(scBytes, &scEnvelope); err != nil {
		t.Fatalf("structured content is not valid envelope: %v", err)
	}
	if !scEnvelope.Success {
		t.Error("structured envelope: success should be true")
	}

	// Both channels should carry the same data
	if textEnvelope.Message != scEnvelope.Message {
		t.Errorf("channel mismatch: text=%q structured=%q", textEnvelope.Message, scEnvelope.Message)
	}
}

// --- Probe 3b: Response Envelope structure ---
func TestProbe_ResponseEnvelope(t *testing.T) {
	_, cs := newTestServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_ping"})
	if err != nil {
		t.Fatalf("calling vivechak_ping: %v", err)
	}

	scBytes, _ := json.Marshal(result.StructuredContent)
	var env Envelope
	json.Unmarshal(scBytes, &env)

	// Verify envelope fields per FINAL-PLAN.md
	if env.NextStep == "" {
		t.Error("envelope missing next_step — violates Guided Worker pattern")
	}
	if env.Meta == nil {
		t.Error("envelope missing meta")
	} else {
		if env.Meta.API != 1 {
			t.Errorf("expected meta.api=1, got %d", env.Meta.API)
		}
		if env.Meta.Tool != "vivechak_ping" {
			t.Errorf("expected meta.tool=vivechak_ping, got %s", env.Meta.Tool)
		}
	}
}

// --- Probe 6: Workspace Resolution ---
func TestProbe_WorkspaceResolution(t *testing.T) {
	_, cs := newTestServer(t)
	ctx := context.Background()

	// Test with explicit project_root
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_echo",
		Arguments: map[string]any{
			"text":         "hello",
			"project_root": "/tmp/test-workspace",
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_echo: %v", err)
	}

	scBytes, _ := json.Marshal(result.StructuredContent)
	var env Envelope
	json.Unmarshal(scBytes, &env)

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("envelope data is not a map")
	}

	workspace, _ := data["workspace_root"].(string)
	if workspace != "/tmp/test-workspace" {
		t.Errorf("explicit project_root not respected: got %q", workspace)
	}

	// Test without explicit project_root (should fall back to env/CWD)
	result2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_echo",
		Arguments: map[string]any{
			"text": "hello",
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_echo (no root): %v", err)
	}

	scBytes2, _ := json.Marshal(result2.StructuredContent)
	var env2 Envelope
	json.Unmarshal(scBytes2, &env2)

	data2, _ := env2.Data.(map[string]any)
	workspace2, _ := data2["workspace_root"].(string)
	if workspace2 == "" {
		t.Error("workspace resolution should not return empty string")
	}
}

// --- Probe 7: Output Size Warning ---
func TestProbe_OutputSizeWarning(t *testing.T) {
	_, cs := newTestServer(t)
	ctx := context.Background()

	// Generate text that exceeds ~10K tokens (40K chars)
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_echo",
		Arguments: map[string]any{
			"text":         strings.Repeat("A", 100),
			"repeat_count": 500,
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_echo with large output: %v", err)
	}

	// Use the text content channel (more reliable for large payloads)
	if len(result.Content) == 0 {
		t.Fatal("no text content returned")
	}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var env Envelope
	if err := json.Unmarshal([]byte(textContent.Text), &env); err != nil {
		t.Fatalf("text content is not valid JSON envelope: %v", err)
	}

	if len(env.Warnings) == 0 {
		data, ok2 := env.Data.(map[string]any)
		if !ok2 {
			t.Fatalf("Data is %T, not map[string]any", env.Data)
		}
		// List all keys
		var keys []string
		for k := range data {
			keys = append(keys, k)
		}
		t.Logf("Data map keys: %v", keys)
		approxTokens, _ := data["approx_tokens"]
		repeatCount, _ := data["repeat_count"]
		byteLen, _ := data["byte_length"]
		t.Errorf("expected output size warning for large response, got none. approx_tokens=%v, repeat_count=%v, byte_length=%v", approxTokens, repeatCount, byteLen)
	}

	foundWarning := false
	for _, w := range env.Warnings {
		if strings.Contains(w, "W-OUTPUT-SIZE") {
			foundWarning = true
			break
		}
	}
	if !foundWarning && len(env.Warnings) > 0 {
		t.Errorf("expected W-OUTPUT-SIZE warning, got: %v", env.Warnings)
	}
}

// --- Probe: IsError should be false for successful calls ---
func TestProbe_NotErrorOnSuccess(t *testing.T) {
	_, cs := newTestServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_ping"})
	if err != nil {
		t.Fatalf("calling vivechak_ping: %v", err)
	}
	if result.IsError {
		t.Error("vivechak_ping should not return isError=true on success")
	}
}
