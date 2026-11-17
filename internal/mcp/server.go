// Package mcputil provides the Vivechak MCP server and tool registration.
package mcputil

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerInstructions provides concise, neutral operational guidance (<= 600 bytes).
const ServerInstructions = "Vivechak runs evidence-graded research for technical decisions under research/ in the user's project.\n" +
	"Use for option comparisons, architectural ADRs, or system plans; not for writing or debugging code.\n" +
	"Flow: vivechak_start(scope, question) -> search -> vivechak_save. Projects: vivechak_next -> research -> vivechak_save -> vivechak_gate.\n" +
	"Follow `next` in every response; pass project_root if the workspace can't be resolved.\n" +
	"Grade claims A-E (vivechak://methodology/grades); recalled claims cap at Grade D.\n" +
	"One-way decisions need vck signoff <D-id>; never set review fields yourself."

// NewServer creates and configures the Vivechak MCP server with 9 canonical tools.
func NewServer(version string, logger *slog.Logger) *sdkmcp.Server {
	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "vivechak", Version: version},
		&sdkmcp.ServerOptions{
			Instructions: ServerInstructions,
			Logger:       logger,
			//nolint:staticcheck // SA1019: roots remain functional during the 12-month deprecation window (SEP-2577)
			RootsListChangedHandler: handleRootsListChanged,
		},
	)

	// Workspace resolution remembers the last workspace per MCP session, never across sessions (HIGH-14).
	server.AddReceivingMiddleware(func(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
		return func(ctx context.Context, method string, req sdkmcp.Request) (sdkmcp.Result, error) {
			if ss, ok := req.GetSession().(*sdkmcp.ServerSession); ok {
				if st := stateForSession(ss); st != nil {
					ctx = core.WithWorkspaceMemory(ctx, &st.workspace)
				}
			}
			return next(ctx, method, req)
		}
	})

	// Panic recovery middleware
	server.AddReceivingMiddleware(func(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
		return func(ctx context.Context, method string, req sdkmcp.Request) (res sdkmcp.Result, err error) {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					fmt.Fprintf(os.Stderr, "PANIC in %s: %v\n%s\n", method, r, buf[:n])
					if method == "tools/call" {
						env := ErrorEnvelope("server", fmt.Errorf("E-INTERNAL: internal error: %v", r), "Please report this issue.")
						jsonBytes, _ := json.Marshal(env)
						res = &sdkmcp.CallToolResult{
							Content: []sdkmcp.Content{
								&sdkmcp.TextContent{Text: string(jsonBytes)},
							},
							StructuredContent: json.RawMessage(jsonBytes),
							IsError:           true,
						}
						err = nil
					}
				}
			}()

			return next(ctx, method, req)
		}
	})

	// Register the 9 canonical tools
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
