package mcputil

import (
	"context"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
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

func TestPrepareGenerator_AnchorsProjectRootWithScope(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"context":      "Pino vs Winston on Node 22",
		},
	})
	if err != nil {
		t.Fatalf("prepare_generator: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("expected success, got: %s", env.Message)
	}

	// Verify project_root was anchored in session memory
	lastResolved := core.GetLastResolvedWorkspace()
	if !strings.EqualFold(lastResolved, tmpDir) {
		t.Errorf("core.GetLastResolvedWorkspace() = %q, want %q", lastResolved, tmpDir)
	}

	// Verify the generated prompt contains the cleaned options
	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map in envelope")
	}
	prompt, _ := dataMap["prompt"].(string)
	if !strings.Contains(prompt, "Pino, Winston") {
		t.Errorf("expected generator prompt to contain extracted options 'Pino, Winston', got:\n%s", prompt)
	}
}
