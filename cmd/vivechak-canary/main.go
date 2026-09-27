// Spike 0: Client Compatibility Canary
//
// Throwaway 2-tool MCP server to test against all 7 target hosts.
// DO NOT use this in production — the full server lives in cmd/vivechak/.
//
// Probes to verify with this canary:
//   1. Tool listing — do both tools appear in each client's tool picker?
//   2. Annotations — does readOnlyHint: true suppress approval prompts?
//   3. Response channels — structuredContent AND text content both work?
//   4. alwaysLoad — does the flag work for vivechak_ping?
//   5. PATH resolution — absolute-path command vs bare command
//   6. Workspace binding — VIVECHAK_PROJECT_ROOT / CWD resolution
//   7. Output size — what happens at 10K and 25K tokens?
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is set at build time via ldflags.
var version = "0.0.1-canary"

// --- Response Envelope (matches FINAL-PLAN.md §Standardized Response Envelope) ---

// Envelope is the standardized response every tool returns.
type Envelope struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     any         `json:"data,omitempty"`
	Warnings []string    `json:"warnings,omitempty"`
	NextStep string      `json:"next_step"`
	Meta     *EnvMeta    `json:"meta,omitempty"`
}

// EnvMeta carries non-functional metadata about the response.
type EnvMeta struct {
	API       int    `json:"api"`
	Tool      string `json:"tool"`
	Truncated bool   `json:"truncated"`
}

// --- Tool Inputs ---

// PingInput is the (empty) input for vivechak_ping.
type PingInput struct{}

// EchoInput holds the arguments for vivechak_echo.
type EchoInput struct {
	Text        string `json:"text,omitempty"         jsonschema:"text to echo back in the response envelope"`
	RepeatCount int    `json:"repeat_count,omitempty"  jsonschema:"number of times to repeat the text (for output-size testing; default 1)"`
	ProjectRoot string `json:"project_root,omitempty"  jsonschema:"explicit workspace root (optional; falls back to VIVECHAK_PROJECT_ROOT then CWD)"`
}

// --- Helpers ---

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool { return &b }

// resolveWorkspace implements the 4-step workspace resolution chain:
//   1. Explicit project_root argument
//   2. VIVECHAK_PROJECT_ROOT env
//   3. CLAUDE_PROJECT_DIR env
//   4. CWD
func resolveWorkspace(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if v := os.Getenv("VIVECHAK_PROJECT_ROOT"); v != "" {
		return v
	}
	if v := os.Getenv("CLAUDE_PROJECT_DIR"); v != "" {
		return v
	}
	cwd, _ := os.Getwd()
	return cwd
}

// envelopeToResult packages an Envelope into a CallToolResult with both
// structuredContent AND text content (dual-channel delivery per FINAL-PLAN.md).
func envelopeToResult(env Envelope) (*mcp.CallToolResult, Envelope, error) {
	// Marshal for the text channel
	jsonBytes, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return nil, env, fmt.Errorf("marshaling envelope: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
		// StructuredContent is auto-populated from the Out value by the SDK
	}, env, nil
}

func main() {
	// CRITICAL: slog MUST write to stderr; stdout is protocol-only.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	server := mcp.NewServer(
		&mcp.Implementation{Name: "vivechak", Version: version},
		&mcp.ServerOptions{
			Instructions: "Vivechak Canary Server — Spike 0 compatibility test. " +
				"Two tools: vivechak_ping (read-only status check) and vivechak_echo (echo with workspace resolution). " +
				"Start with vivechak_ping to verify connectivity.",
			Logger: logger,
		},
	)

	// --- Tool 1: vivechak_ping ---
	// Read-only, always-loaded orientation tool.
	// Tests probes: tool listing, annotations, response channels, alwaysLoad.
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "vivechak_ping",
			Title: "Vivechak Ping",
			Description: "Check Vivechak server connectivity and version. Returns server status in the " +
				"standardized response envelope with both structured and text content. Use this tool first " +
				"to verify the MCP connection is working. Read-only — no filesystem changes.",
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
				Message: fmt.Sprintf("Vivechak canary server v%s is running", version),
				Data: map[string]any{
					"version":          version,
					"go_version":       "1.27",
					"sdk_version":      "1.8.0",
					"canary":           true,
					"workspace_cwd":    resolveWorkspace(""),
				},
				NextStep: "Run vivechak_echo with test text to verify response handling.",
				Meta: &EnvMeta{
					API:  1,
					Tool: "vivechak_ping",
				},
			}
			return envelopeToResult(env)
		},
	)

	// --- Tool 2: vivechak_echo ---
	// Mutating (writes nothing, but not read-only to test annotation behavior).
	// Tests probes: response channels, workspace binding, output size.
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "vivechak_echo",
			Title: "Vivechak Echo",
			Description: "Echo back provided text in the standardized response envelope. " +
				"Supports repeat_count for output-size stress testing (Claude Code warns at ~10K tokens, " +
				"hard caps at ~25K). Resolves workspace root via 4-step precedence chain: " +
				"explicit project_root > VIVECHAK_PROJECT_ROOT > CLAUDE_PROJECT_DIR > CWD. " +
				"Not read-only — tests whether annotation-based approval prompts fire correctly.",
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  true,
				DestructiveHint: boolPtr(false),
				OpenWorldHint:   boolPtr(false),
			},
		},
		func(ctx context.Context, req *mcp.CallToolRequest, in EchoInput) (*mcp.CallToolResult, Envelope, error) {
			// Default repeat count
			count := in.RepeatCount
			if count <= 0 {
				count = 1
			}
			// Cap at 500 to avoid extreme payloads
			if count > 500 {
				count = 500
			}

			workspace := resolveWorkspace(in.ProjectRoot)

			// Build repeated text
			var text string
			if count == 1 {
				text = in.Text
			} else {
				text = strings.Repeat(in.Text+"\n", count)
			}

			// Calculate approximate token count (rough: 1 token ≈ 4 chars)
			approxTokens := len(text) / 4
			var warnings []string
			truncated := false

			if approxTokens > 25000 {
				warnings = append(warnings, "W-OUTPUT-SIZE: Response exceeds 25K token hard cap threshold")
				truncated = true
				// Truncate to ~25K tokens worth
				if len(text) > 100000 {
					text = text[:100000] + "\n... [TRUNCATED]"
				}
			} else if approxTokens > 10000 {
				warnings = append(warnings, "W-OUTPUT-SIZE: Response exceeds 10K token warning threshold")
			}

			env := Envelope{
				Success:  true,
				Message:  fmt.Sprintf("Echoed %d bytes (%d repeats) from workspace %s", len(text), count, workspace),
				Data: map[string]any{
					"echo_text":       text,
					"repeat_count":    count,
					"byte_length":     len(text),
					"approx_tokens":   approxTokens,
					"workspace_root":  workspace,
					"resolution_chain": map[string]string{
						"explicit":             in.ProjectRoot,
						"VIVECHAK_PROJECT_ROOT": os.Getenv("VIVECHAK_PROJECT_ROOT"),
						"CLAUDE_PROJECT_DIR":    os.Getenv("CLAUDE_PROJECT_DIR"),
						"cwd":                  func() string { d, _ := os.Getwd(); return d }(),
					},
				},
				Warnings: warnings,
				NextStep: "Verify this response appeared correctly in your client. " +
					"Check: (1) structured content visible, (2) text content visible, (3) warnings shown if present.",
				Meta: &EnvMeta{
					API:       1,
					Tool:      "vivechak_echo",
					Truncated: truncated,
				},
			}
			return envelopeToResult(env)
		},
	)

	// Run the server on stdio transport
	logger.Info("starting canary server", "version", version)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Error("canary server exited with error", "err", err)
		os.Exit(1)
	}
}
