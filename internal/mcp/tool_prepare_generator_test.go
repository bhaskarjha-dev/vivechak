package mcputil

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestExtractComparisonOptions_Prepositions(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Pino vs Winston on Node 22", "Pino, Winston"},
		{"Kafka vs RabbitMQ as an event bus", "Kafka, RabbitMQ"},
		{"Postgres vs MySQL to store session tokens", "Postgres, MySQL"},
		{"FastAPI vs Go across microservices", "FastAPI, Go"},
		{"Pino vs Winston for Node.js logging", "Pino, Winston"},
		{"Postgres vs Mongo in production", "Postgres, Mongo"},
		{"compare Redis vs Memcached under high throughput", "Redis, Memcached"},
		{"compare Kafka vs Pulsar in distributed pipelines", "Kafka, Pulsar"},
		{"evaluate Groß vs Klein under load", "Groß, Klein"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := extractComparisonOptions(tc.input)
			if got != tc.expected {
				t.Errorf("extractComparisonOptions(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestStart_AnchorsProjectRootWithScope(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"question":     "Pino vs Winston on Node 22",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_start: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("expected success, got: %s", env.Message)
	}

	// Verify project_root was anchored in this session's memory: a later call without project_root resolves it
	statusRes, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_status", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("vivechak_status: %v", err)
	}
	statusData, _ := parseEnvelope(t, statusRes).Data.(map[string]any)
	wsInfo, _ := statusData["workspace"].(map[string]any)
	if root, _ := wsInfo["root"].(string); !strings.EqualFold(root, tmpDir) || wsInfo["rule"] != "session" {
		t.Errorf("expected the session to remember %q (rule session), got %v", tmpDir, wsInfo)
	}

	// Verify the generated prompt contains the cleaned options
	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map in envelope")
	}
	prompt, _ := dataMap["brief"].(string)
	if !strings.Contains(prompt, "Pino, Winston") {
		t.Errorf("expected generator prompt to contain extracted options 'Pino, Winston', got:\n%s", prompt)
	}
}
