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

func TestPromptAliases(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	pairs := []struct {
		canonical string
		legacy    string
		shorthand string
		args      map[string]string
	}{
		{
			canonical: "compare",
			legacy:    "vivechak_compare",
			shorthand: "vck_compare",
			args:      map[string]string{"options": "Pino, Winston"},
		},
		{
			canonical: "decision",
			legacy:    "vivechak_decision",
			shorthand: "vck_decision",
			args:      map[string]string{"dilemma": "Migrate DB"},
		},
		{
			canonical: "project",
			legacy:    "vivechak_project",
			shorthand: "vck_project",
			args:      map[string]string{"vision": "Build Engine"},
		},
		{
			canonical: "gate",
			legacy:    "vivechak_gate",
			shorthand: "vck_gate",
			args:      map[string]string{"track": "all"},
		},
	}

	for _, p := range pairs {
		rCanon, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: p.canonical, Arguments: p.args})
		if err != nil {
			t.Fatalf("GetPrompt(%s): %v", p.canonical, err)
		}
		rLeg, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: p.legacy, Arguments: p.args})
		if err != nil {
			t.Fatalf("GetPrompt(%s): %v", p.legacy, err)
		}
		rShort, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: p.shorthand, Arguments: p.args})
		if err != nil {
			t.Fatalf("GetPrompt(%s): %v", p.shorthand, err)
		}

		cCanon := rCanon.Messages[0].Content.(*mcp.TextContent).Text
		cLeg := rLeg.Messages[0].Content.(*mcp.TextContent).Text
		cShort := rShort.Messages[0].Content.(*mcp.TextContent).Text

		if cCanon != cLeg {
			t.Errorf("prompt %s != %s", p.canonical, p.legacy)
		}
		if cCanon != cShort {
			t.Errorf("prompt %s != %s", p.canonical, p.shorthand)
		}
	}
}

func TestPromptGet_Compare(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// 1. With explicit arguments
	res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_compare",
		Arguments: map[string]string{
			"options":  "Pino, Winston",
			"criteria": "Throughput, ESM support",
			"context":  "High-throughput microservices in Node.js",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt(vivechak_compare): %v", err)
	}

	if len(res.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(res.Messages))
	}
	msg := res.Messages[0]
	if msg.Role != mcp.Role("user") {
		t.Errorf("expected role 'user', got %q", msg.Role)
	}
	tc, ok := msg.Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", msg.Content)
	}
	text := tc.Text

	if !strings.Contains(text, "OPTIONS: Pino, Winston") {
		t.Errorf("missing injected options in prompt text: %s", text)
	}
	if !strings.Contains(text, "CRITERIA: Throughput, ESM support") {
		t.Errorf("missing injected criteria in prompt text: %s", text)
	}
	if !strings.Contains(text, "High-throughput microservices") {
		t.Errorf("missing injected context in prompt text: %s", text)
	}
	if !strings.Contains(text, "vivechak_save_plan") {
		t.Errorf("missing save_plan directive in prompt text")
	}
	if strings.Contains(text, "From the context block, output a SCOPE CHECK line") {
		t.Errorf("meta-generator instruction was not stripped from prompt output: %s", text)
	}
	if !strings.Contains(text, "Execute this evidence-calibrated research brief with live web search to compare the options below and produce the Weighted Evaluation Protocol (WEP) matrix.") {
		t.Errorf("direct research execution directive was not injected into prompt text")
	}

	// 2. With empty / omitted optional arguments
	resDefault, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_compare",
		Arguments: map[string]string{
			"options": "A vs B",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt(vivechak_compare default): %v", err)
	}
	tcDef, _ := resDefault.Messages[0].Content.(*mcp.TextContent)
	if !strings.Contains(tcDef.Text, "OPTIONS: A vs B") {
		t.Errorf("missing options in default call: %s", tcDef.Text)
	}
}

func TestPromptGet_Decision(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_decision",
		Arguments: map[string]string{
			"dilemma": "Migrate event log from Postgres to ClickHouse",
			"context": "100k events/sec, multi-tenant SaaS",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt(vivechak_decision): %v", err)
	}

	tc, ok := res.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Messages[0].Content)
	}

	if !strings.Contains(tc.Text, "DECISION: Migrate event log from Postgres to ClickHouse") {
		t.Errorf("missing injected dilemma: %s", tc.Text)
	}
	if !strings.Contains(tc.Text, "100k events/sec") {
		t.Errorf("missing injected context: %s", tc.Text)
	}
	if !strings.Contains(tc.Text, "vivechak_save_plan") {
		t.Errorf("missing save_plan directive")
	}
}

func TestPromptGet_Project(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_project",
		Arguments: map[string]string{
			"vision":  "Build a real-time collaborative code review engine",
			"context": "Go backend, WebSockets, Redis cache",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt(vivechak_project): %v", err)
	}

	tc, ok := res.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Messages[0].Content)
	}

	if !strings.Contains(tc.Text, "Build a real-time collaborative code review engine") {
		t.Errorf("missing injected vision: %s", tc.Text)
	}
	if !strings.Contains(tc.Text, "Go backend, WebSockets") {
		t.Errorf("missing injected context: %s", tc.Text)
	}
	if !strings.Contains(tc.Text, "pipeline_content") {
		t.Errorf("missing pipeline_content directive: %s", tc.Text)
	}
}

func TestPromptGet_Gate(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_gate",
		Arguments: map[string]string{
			"track":        "track_b",
			"project_root": "/custom/workspace",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt(vivechak_gate): %v", err)
	}

	tc, ok := res.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Messages[0].Content)
	}

	if !strings.Contains(tc.Text, "vivechak_run_gate") {
		t.Errorf("missing vivechak_run_gate instruction: %s", tc.Text)
	}
	if !strings.Contains(tc.Text, `"track_b"`) {
		t.Errorf("missing track_b parameter: %s", tc.Text)
	}
	if !strings.Contains(tc.Text, `project_root="/custom/workspace"`) {
		t.Errorf("missing project_root parameter: %s", tc.Text)
	}
}

func TestPromptGet_UnknownPrompt(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	_, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_unknown",
	})
	if err == nil {
		t.Fatal("expected error getting unknown prompt, got nil")
	}
}

func TestPromptGet_EmptyArgumentsFallback(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// 1. Compare with nil arguments
	resComp, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_compare",
	})
	if err != nil {
		t.Fatalf("GetPrompt(vivechak_compare, nil): %v", err)
	}
	tcComp := resComp.Messages[0].Content.(*mcp.TextContent)
	if !strings.Contains(tcComp.Text, "[options to compare]") {
		t.Errorf("expected fallback options placeholder, got: %s", tcComp.Text)
	}

	// 2. Decision with nil arguments
	resDec, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_decision",
	})
	if err != nil {
		t.Fatalf("GetPrompt(vivechak_decision, nil): %v", err)
	}
	tcDec := resDec.Messages[0].Content.(*mcp.TextContent)
	if !strings.Contains(tcDec.Text, "[architectural fork or dilemma]") {
		t.Errorf("expected fallback dilemma placeholder, got: %s", tcDec.Text)
	}

	// 3. Project with nil arguments
	resProj, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_project",
	})
	if err != nil {
		t.Fatalf("GetPrompt(vivechak_project, nil): %v", err)
	}
	tcProj := resProj.Messages[0].Content.(*mcp.TextContent)
	if !strings.Contains(tcProj.Text, "[system architecture vision and requirements]") {
		t.Errorf("expected fallback vision placeholder, got: %s", tcProj.Text)
	}

	// 4. Gate with nil arguments
	resGate, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_gate",
	})
	if err != nil {
		t.Fatalf("GetPrompt(vivechak_gate, nil): %v", err)
	}
	tcGate := resGate.Messages[0].Content.(*mcp.TextContent)
	if !strings.Contains(tcGate.Text, `"all"`) {
		t.Errorf("expected default track 'all', got: %s", tcGate.Text)
	}
}

func TestPromptGet_SpecialCharactersAndMultiline(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// 1. Compare with special chars, quotes, ampersands, and unicode
	resComp, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_compare",
		Arguments: map[string]string{
			"options":  "Option \"A\" (v1.2) vs Option 'B' & Option C (देवनागरी)",
			"criteria": "Throughput > 10k req/s & latency < 5ms",
			"context":  "Multi-line\ncontext\nwith `code blocks` and $symbols",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt compare with special chars: %v", err)
	}
	textComp := resComp.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(textComp, "Option \"A\" (v1.2) vs Option 'B' & Option C (देवनागरी)") {
		t.Errorf("special characters not preserved in compare options")
	}
	if !strings.Contains(textComp, "Throughput > 10k req/s & latency < 5ms") {
		t.Errorf("criteria with symbols not preserved")
	}

	// 2. Decision with markdown fences and multiline dilemma
	resDec, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_decision",
		Arguments: map[string]string{
			"dilemma": "Migrate `auth_service` from OAuth2 to Passkeys/WebAuthn",
			"context": "Stack: Go 1.25+\nScale: 10M users\nBudget: $0/month additional SaaS cost",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt decision with multiline: %v", err)
	}
	textDec := resDec.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(textDec, "Migrate `auth_service` from OAuth2 to Passkeys/WebAuthn") {
		t.Errorf("decision dilemma with code blocks not preserved")
	}
	if !strings.Contains(textDec, "Budget: $0/month") {
		t.Errorf("multiline context not preserved")
	}

	// 3. Project with nested markdown fences
	resProj, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "vivechak_project",
		Arguments: map[string]string{
			"vision":  "Build distributed KV store with raft consensus\n```go\ntype Node struct{}\n```",
			"context": "Security: mTLS on port 9090",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt project with nested code blocks: %v", err)
	}
	textProj := resProj.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(textProj, "Build distributed KV store with raft consensus") {
		t.Errorf("project vision not preserved")
	}
	if !strings.Contains(textProj, "mTLS on port 9090") {
		t.Errorf("project context not preserved")
	}
}

