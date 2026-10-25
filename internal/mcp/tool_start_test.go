package mcputil

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestStart_ProjectScope_MultipartChunking(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Request Part 1
	res1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"question":     "Architect a real-time event streaming pipeline with ClickHouse and Kafka",
			"part":         1,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_start part 1 failed: %v", err)
	}
	env1 := parseEnvelope(t, res1)
	if !env1.Success {
		t.Fatalf("expected success on part 1, got error: %s", env1.Message)
	}
	data1, ok := env1.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected env1.Data to be map[string]any, got %T", env1.Data)
	}

	part1Num, ok := data1["part"].(float64)
	if !ok || int(part1Num) != 1 {
		t.Errorf("expected part 1, got %v", data1["part"])
	}
	totalParts, ok := data1["total_parts"].(float64)
	if !ok || int(totalParts) < 2 {
		t.Errorf("expected total_parts >= 2, got %v", data1["total_parts"])
	}
	hasMore1, ok := data1["has_more"].(bool)
	if !ok || !hasMore1 {
		t.Errorf("expected has_more == true for part 1, got %v", data1["has_more"])
	}

	brief1, ok := data1["brief"].(string)
	if !ok || len(brief1) == 0 {
		t.Fatalf("brief1 is empty")
	}

	if env1.Next == nil || env1.Next.Tool != "vivechak_start" {
		t.Errorf("expected NextAction tool to be vivechak_start, got %v", env1.Next)
	}
	if nextPart, ok := env1.Next.Args["part"].(float64); !ok || int(nextPart) != 2 {
		t.Errorf("expected NextAction part arg to be 2, got %v", env1.Next.Args["part"])
	}

	// 2. Request Part 2
	res2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"question":     "Architect a real-time event streaming pipeline with ClickHouse and Kafka",
			"part":         2,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_start part 2 failed: %v", err)
	}
	env2 := parseEnvelope(t, res2)
	if !env2.Success {
		t.Fatalf("expected success on part 2, got error: %s", env2.Message)
	}
	data2, ok := env2.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected env2.Data to be map[string]any, got %T", env2.Data)
	}

	part2Num, ok := data2["part"].(float64)
	if !ok || int(part2Num) != 2 {
		t.Errorf("expected part 2, got %v", data2["part"])
	}
	hasMore2, ok := data2["has_more"].(bool)
	if !ok || hasMore2 {
		t.Errorf("expected has_more == false for final part 2, got %v", data2["has_more"])
	}

	brief2, ok := data2["brief"].(string)
	if !ok || len(brief2) == 0 {
		t.Fatalf("brief2 is empty")
	}

	if env2.Next == nil || env2.Next.Tool != "vivechak_save" {
		t.Errorf("expected NextAction tool to be vivechak_save on final part, got %v", env2.Next)
	}

	// Combined brief contains both vision question and later pipeline instructions
	combined := brief1 + "\n" + brief2
	if !strings.Contains(combined, "ClickHouse and Kafka") {
		t.Errorf("combined brief missing project context question")
	}
}

func TestStart_ProjectScope_DefaultOmittedPart(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"question":     "Vision: High-throughput ingestion engine",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_start failed: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("expected success, got error: %s", env.Message)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected env.Data to be map[string]any, got %T", env.Data)
	}

	totalParts, ok := data["total_parts"].(float64)
	if !ok || int(totalParts) < 2 {
		t.Errorf("expected total_parts >= 2, got %v", data["total_parts"])
	}

	brief, ok := data["brief"].(string)
	if !ok || len(brief) == 0 {
		t.Fatalf("brief is empty")
	}

	// Full brief should be returned when part is omitted
	totalCharCount, ok := data["total_char_count"].(float64)
	if !ok || int(totalCharCount) != len(brief) {
		t.Errorf("expected brief length %d to match total_char_count %v", len(brief), data["total_char_count"])
	}
	if env.Next == nil || env.Next.Tool != "vivechak_save" {
		t.Errorf("expected NextAction tool to be vivechak_save, got %v", env.Next)
	}
}

func TestStart_InvalidPart(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"question":     "Test",
			"part":         99,
		},
	})
	if err != nil {
		t.Fatalf("call tool failed: %v", err)
	}
	env := parseEnvelope(t, res)
	if env.Success {
		t.Errorf("expected failure for out-of-range part 99, got success")
	}
	if !strings.Contains(env.Message, "invalid part 99") {
		t.Errorf("expected error message to mention invalid part 99, got: %s", env.Message)
	}
}

func TestSplitBriefIntoParts(t *testing.T) {
	short := "Brief under target size."
	parts := splitBriefIntoParts(short, 100)
	if len(parts) != 1 || parts[0] != short {
		t.Errorf("expected 1 part for short brief, got %d", len(parts))
	}

	// Build a synthetic large markdown text with headings
	section1 := "## Section 1\n" + strings.Repeat("Content in section 1. ", 50) + "\n\n"
	section2 := "### Section 2\n" + strings.Repeat("Content in section 2. ", 50) + "\n\n"
	section3 := "## Section 3\n" + strings.Repeat("Content in section 3. ", 50)
	largeText := section1 + section2 + section3

	targetSize := len(section1) + 50
	split := splitBriefIntoParts(largeText, targetSize)
	if len(split) < 2 {
		t.Fatalf("expected at least 2 parts, got %d", len(split))
	}

	for i, p := range split {
		if len(p) == 0 {
			t.Errorf("part %d is empty", i+1)
		}
	}
}

func TestStart_FilesystemRootAndHomeGuards(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// 1. Filesystem root / drive root
	resRoot, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"project_root": "/",
			"scope":        "comparison",
			"question":     "Pino vs Winston",
		},
	})
	if err != nil {
		t.Fatalf("call tool failed: %v", err)
	}
	envRoot := parseEnvelope(t, resRoot)
	if envRoot.Success {
		t.Errorf("expected failure for filesystem root '/', got success")
	}
	if !strings.Contains(envRoot.Message, "cannot be a filesystem root") {
		t.Errorf("expected 'cannot be a filesystem root', got: %s", envRoot.Message)
	}

	// 2. User home directory
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		resHome, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_start",
			Arguments: map[string]any{
				"project_root": home,
				"scope":        "comparison",
				"question":     "Pino vs Winston",
			},
		})
		if err != nil {
			t.Fatalf("call tool failed: %v", err)
		}
		envHome := parseEnvelope(t, resHome)
		if envHome.Success {
			t.Errorf("expected failure for user home dir %q, got success", home)
		}
		if !strings.Contains(envHome.Message, "refusing to initialize workspace directly in user home directory") {
			t.Errorf("expected home directory refusal message, got: %s", envHome.Message)
		}
	}
}

func TestStart_ScopeInference(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"question":     "Compare Pino vs Winston for structured logging",
		},
	})
	if err != nil {
		t.Fatalf("call tool failed: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("expected success, got error: %s", env.Message)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected env.Data to be map[string]any, got %T", env.Data)
	}
	if data["scope"] != "comparison" {
		t.Errorf("expected inferred scope 'comparison', got: %v", data["scope"])
	}
}
