package mcputil

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPromptListing(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}

	expectedPrompts := map[string]string{
		"compare":  "Compare Options (WEP Matrix)",
		"decision": "Architectural Decision (ADR)",
		"project":  "Project Pipeline (Full FAD)",
		"gate":     "Phase 0 Pre-Codebase Exit Gate",
	}

	if len(res.Prompts) != len(expectedPrompts) {
		t.Fatalf("expected %d prompts, got %d", len(expectedPrompts), len(res.Prompts))
	}

	for _, p := range res.Prompts {
		expectedTitle, ok := expectedPrompts[p.Name]
		if !ok {
			t.Errorf("unexpected prompt registered: %s", p.Name)
			continue
		}
		if p.Title != expectedTitle {
			t.Errorf("prompt %s: expected title %q, got %q", p.Name, expectedTitle, p.Title)
		}
		if p.Description == "" {
			t.Errorf("prompt %s: description is empty", p.Name)
		}
	}
}

func TestPromptAliasesRejected(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	removed := []string{
		"vivechak_compare", "vck_compare",
		"vivechak_decision", "vck_decision",
		"vivechak_project", "vck_project",
		"vivechak_gate", "vck_gate",
	}
	for _, name := range removed {
		if _, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: name}); err == nil {
			t.Errorf("expected removed prompt alias %q to return an error", name)
		}
	}
}

func TestPromptGet_Compare(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "compare",
		Arguments: map[string]string{
			"options":  "Pino, Winston",
			"criteria": "Throughput, ESM support",
			"context":  "High-throughput microservices in Node.js",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt(compare): %v", err)
	}
	if len(res.Messages) != 1 || res.Messages[0].Role != mcp.Role("user") {
		t.Fatalf("unexpected prompt messages: %+v", res.Messages)
	}
	text := res.Messages[0].Content.(*mcp.TextContent).Text
	if len(text) > 600 {
		t.Errorf("compare prompt exceeds 600 chars: %d", len(text))
	}
	for _, want := range []string{"Pino, Winston", "Throughput, ESM support", "High-throughput microservices", "vivechak_start", "vivechak_save", "next"} {
		if !strings.Contains(text, want) {
			t.Errorf("compare prompt missing %q: %s", want, text)
		}
	}
}

func TestPromptGet_Decision(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "decision",
		Arguments: map[string]string{
			"question": "Migrate event log from Postgres to ClickHouse",
			"context":  "100k events/sec, multi-tenant SaaS",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt(decision): %v", err)
	}
	text := res.Messages[0].Content.(*mcp.TextContent).Text
	if len(text) > 600 {
		t.Errorf("decision prompt exceeds 600 chars: %d", len(text))
	}
	for _, want := range []string{"Migrate event log from Postgres to ClickHouse", "100k events/sec", "vivechak_start", "next"} {
		if !strings.Contains(text, want) {
			t.Errorf("decision prompt missing %q: %s", want, text)
		}
	}
	if strings.Contains(text, "D-001") {
		t.Errorf("decision prompt must not hardcode D-001: %s", text)
	}
}

func TestPromptGet_Project(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "project",
		Arguments: map[string]string{
			"vision":      "Build a real-time collaborative code review engine",
			"constraints": "Go backend, WebSockets, Redis cache",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt(project): %v", err)
	}
	text := res.Messages[0].Content.(*mcp.TextContent).Text
	if len(text) > 600 {
		t.Errorf("project prompt exceeds 600 chars: %d", len(text))
	}
	for _, want := range []string{"Build a real-time collaborative code review engine", "Go backend, WebSockets", "vivechak_start", "next"} {
		if !strings.Contains(text, want) {
			t.Errorf("project prompt missing %q: %s", want, text)
		}
	}
}

func TestPromptGet_Gate(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "gate",
		Arguments: map[string]string{
			"decision": "D-001",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt(gate): %v", err)
	}
	text := res.Messages[0].Content.(*mcp.TextContent).Text
	if len(text) > 600 {
		t.Errorf("gate prompt exceeds 600 chars: %d", len(text))
	}
	for _, want := range []string{"vivechak_gate", "D-001", "vck signoff"} {
		if !strings.Contains(text, want) {
			t.Errorf("gate prompt missing %q: %s", want, text)
		}
	}
}

func TestPromptGet_UnknownPrompt(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	if _, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: "vivechak_unknown"}); err == nil {
		t.Fatal("expected error getting unknown prompt, got nil")
	}
}

func TestPromptGet_EmptyArgumentsFallback(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	for _, name := range []string{"compare", "decision", "project", "gate"} {
		res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: name})
		if err != nil {
			t.Fatalf("GetPrompt(%s, nil): %v", name, err)
		}
		text := res.Messages[0].Content.(*mcp.TextContent).Text
		if text == "" || len(text) > 600 {
			t.Errorf("prompt %s empty or >600 chars (%d): %s", name, len(text), text)
		}
	}
}

func TestPromptGet_SpecialCharactersAndMultiline(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	resComp, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "compare",
		Arguments: map[string]string{
			"options":  "Option \"A\" (v1.2) vs Option 'B' & Option C (देवनागरी)",
			"criteria": "Throughput > 10k req/s & latency < 5ms",
			"context":  "Multi-line\ncontext\nwith $symbols",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt compare with special chars: %v", err)
	}
	textComp := resComp.Messages[0].Content.(*mcp.TextContent).Text
	if len(textComp) > 600 || !strings.Contains(textComp, "Option \"A\" (v1.2)") || !strings.Contains(textComp, "Throughput > 10k req/s") {
		t.Errorf("special characters not preserved or >600 chars: %s", textComp)
	}
}
