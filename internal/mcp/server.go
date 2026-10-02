// Package mcputil provides the Vivechak MCP server and tool registration.
package mcputil

import (
	"log/slog"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer creates and configures the Vivechak MCP server with all 10 tools.
func NewServer(version string, logger *slog.Logger) *sdkmcp.Server {
	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "vivechak", Version: version},
		&sdkmcp.ServerOptions{
			Instructions: "Vivechak: Evidence-grounded research for technical decisions. " +
				"Start with vivechak_status to orient, or vivechak_init to create a new workspace. " +
				"This server provides 10 tools for running Vivechak research pipelines at three scope levels: " +
				"project (full pipeline → FAD), decision (1-3 sessions → ADR), or comparison (1 session → WEP matrix). " +
				"Every response includes a next_step field guiding you to the appropriate next action.",
			Logger: logger,
		},
	)

	// Register all 10 tools in build order
	registerInit(server)
	registerPrepareGenerator(server)
	registerSavePlan(server)
	registerStatus(server)
	registerNextSession(server)
	registerSaveSession(server)
	registerRecordDecision(server)
	registerValidate(server)
	registerRunGate(server)
	registerAmendSession(server)

	return server
}
