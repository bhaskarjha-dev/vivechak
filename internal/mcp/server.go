// Package mcputil provides the Vivechak MCP server and tool registration.
package mcputil

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerInstructions provides concise, neutral operational guidance (≤ 600 chars, AX-01/AX-02).
const ServerInstructions = "Vivechak structures evidence-graded research for technical choices under research/ in user project.\n" +
	"Use for option comparisons, architectural ADRs, or system architecture. Not for code authoring or debugging.\n" +
	"Flow: vivechak_start(scope, question) -> search -> vivechak_save. Projects: vivechak_next -> save -> vivechak_gate.\n" +
	"Grade claims A-E (vivechak://methodology/grades); recalled claims are max Grade D.\n" +
	"One-way decisions need user sign-off (vck signoff <D-id>); never set human_reviewed yourself."

// NewServer creates and configures the Vivechak MCP server with 9 consolidated tools.
func NewServer(version string, logger *slog.Logger) *sdkmcp.Server {
	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "vivechak", Version: version},
		&sdkmcp.ServerOptions{
			Instructions: ServerInstructions,
			Logger:       logger,
		},
	)

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
