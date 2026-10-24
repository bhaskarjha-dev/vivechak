// Package mcputil provides the Vivechak MCP server and tool registration.
package mcputil

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerInstructions provides concise, neutral operational guidance (≤ 600 chars, AX-01/AX-02).
const ServerInstructions = "Vivechak structures evidence-graded research for technical choices under research/ in user project.\n" +
	"Use when asked to compare options, record an architectural decision (ADR), or plan system architecture. Not for code authoring or debugging.\n" +
	"Flow: vivechak_start(scope, question) -> run brief with search -> save with returned id. Projects: vivechak_next -> research -> vivechak_save, repeat -> vivechak_gate.\n" +
	"Workspace: discovered automatically; if E-ROOT-UNRESOLVED, pass project_root='<path>'.\n" +
	"Grade every claim A-E (vivechak://methodology/grades). Recalled claims are max Grade E."

// Legacy retired tool mappings for clear error guidance.
var legacyToolMap = map[string]string{
	"vivechak_init":              "Consolidated into 'vivechak_start'. Use vivechak_start(scope=..., question=...).",
	"vivechak_prepare_generator": "Consolidated into 'vivechak_start'. Use vivechak_start(scope=..., question=...).",
	"vivechak_save_plan":         "Consolidated into 'vivechak_save'. Use vivechak_save(kind='plan', content=...).",
	"vivechak_save_session":      "Consolidated into 'vivechak_save'. Use vivechak_save(kind='session', id=..., content=...).",
	"vivechak_next_session":      "Renamed to 'vivechak_next'. Use vivechak_next().",
	"vivechak_visualize":         "Consolidated into 'vivechak_status'. Use vivechak_status(format='mermaid').",
	"vivechak_validate":          "Consolidated into 'vivechak_review'. Use vivechak_review(mode='validate', ...).",
	"vivechak_challenge":         "Consolidated into 'vivechak_review'. Use vivechak_review(mode='challenge', ...).",
	"vivechak_amend_session":     "Renamed to 'vivechak_amend'. Use vivechak_amend(session_id=..., amendment=...).",
	"vivechak_run_gate":          "Renamed to 'vivechak_gate'. Use vivechak_gate(track=...).",
}

// NewServer creates and configures the Vivechak MCP server with 9 consolidated tools.
func NewServer(version string, logger *slog.Logger) *sdkmcp.Server {
	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "vivechak", Version: version},
		&sdkmcp.ServerOptions{
			Instructions: ServerInstructions,
			Logger:       logger,
		},
	)

	// Panic recovery & legacy tool middleware (BUG-23)
	server.AddReceivingMiddleware(func(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
		return func(ctx context.Context, method string, req sdkmcp.Request) (res sdkmcp.Result, err error) {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					fmt.Fprintf(os.Stderr, "PANIC in %s: %v\n%s\n", method, r, buf[:n])
					if method == "tools/call" {
						env := ErrorEnvelope("server", fmt.Errorf("E-INTERNAL: internal error: %v", r), "Please report this issue.")
						jsonBytes, _ := json.MarshalIndent(env, "", "  ")
						res = &sdkmcp.CallToolResult{
							Content: []sdkmcp.Content{
								&sdkmcp.TextContent{Text: string(jsonBytes)},
							},
							IsError: true,
						}
						err = nil
					}
				}
			}()

			if method == "tools/call" {
				if p, ok := req.GetParams().(*sdkmcp.CallToolParamsRaw); ok && p != nil {
					if hint, isOld := legacyToolMap[p.Name]; isOld {
						rawArgs := strings.TrimSpace(string(p.Arguments))
						if rawArgs == "" || rawArgs == "{}" || rawArgs == "null" {
							env := ErrorEnvelope(p.Name, fmt.Errorf("tool %q has been retired in v0.1.0: %s", p.Name, hint), hint)
							jsonBytes, _ := json.MarshalIndent(env, "", "  ")
							return &sdkmcp.CallToolResult{
								Content: []sdkmcp.Content{
									&sdkmcp.TextContent{Text: string(jsonBytes)},
								},
								IsError: true,
							}, nil
						}

						// Transparent backward compatibility: translate legacy tool calls with arguments
						var args map[string]any
						_ = json.Unmarshal(p.Arguments, &args)
						if args == nil {
							args = make(map[string]any)
						}

						switch p.Name {
						case "vivechak_init":
							p.Name = "vivechak_start"
						case "vivechak_prepare_generator":
							p.Name = "vivechak_start"
							if ctxVal, hasCtx := args["context"]; hasCtx {
								if args["question"] == nil {
									args["question"] = ctxVal
								}
								delete(args, "context")
							}
						case "vivechak_save_plan":
							p.Name = "vivechak_save"
							args["kind"] = "plan"
							if sVal, ok := args["scope"].(string); ok && strings.EqualFold(sVal, "comparison") {
								args["kind"] = "comparison"
							}
							if decID, hasDecID := args["decision_id"]; hasDecID {
								if args["id"] == nil {
									args["id"] = decID
								}
								delete(args, "decision_id")
							}
							if pc, hasPC := args["pipeline_content"]; hasPC {
								if args["content"] == nil {
									args["content"] = pc
								}
								delete(args, "pipeline_content")
							}
						case "vivechak_save_session":
							p.Name = "vivechak_save"
							args["kind"] = "session"
							if sid, hasSID := args["session_id"]; hasSID {
								if args["id"] == nil {
									args["id"] = sid
								}
								delete(args, "session_id")
							}
						case "vivechak_next_session":
							p.Name = "vivechak_next"
						case "vivechak_visualize":
							p.Name = "vivechak_status"
							if args["format"] == nil {
								args["format"] = "mermaid"
							}
						case "vivechak_validate":
							p.Name = "vivechak_review"
							args["mode"] = "validate"
						case "vivechak_challenge":
							p.Name = "vivechak_review"
							args["mode"] = "challenge"
							if sid, hasSID := args["session_id"]; hasSID {
								if args["target"] == nil {
									args["target"] = sid
								}
								delete(args, "session_id")
							}
							if did, hasDID := args["decision_id"]; hasDID {
								if args["target"] == nil {
									args["target"] = did
								}
								delete(args, "decision_id")
							}
							delete(args, "challenge_type")
						case "vivechak_amend_session":
							p.Name = "vivechak_amend"
						case "vivechak_run_gate":
							p.Name = "vivechak_gate"
						}

						if rootVal, hasRoot := args["root"]; hasRoot {
							if args["project_root"] == nil {
								args["project_root"] = rootVal
							}
							delete(args, "root")
						}

						if marshaled, err := json.Marshal(args); err == nil {
							p.Arguments = marshaled
						}
					}
				}
			}

			return next(ctx, method, req)
		}
	})

	// Register the 9 consolidated tools
	registerStart(server)
	registerSave(server)
	registerRecordDecision(server)
	registerNext(server)
	registerStatus(server)
	registerReview(server)
	registerAmend(server)
	registerReplan(server)
	registerGate(server)

	// Register 4 native MCP prompts
	registerPrompts(server)

	// Register 8 canonical MCP resources
	registerResources(server)

	return server
}
