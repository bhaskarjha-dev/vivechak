// Vivechak MCP Server — Evidence-grounded research for technical decisions.
//
// Usage:
//
//	vivechak                    # Start MCP server on stdio (default)
//	vivechak serve              # Explicit serve subcommand
//	vivechak mcp-config --client <host> --write   # Generate client config
//	vivechak doctor             # Check workspace integrity
//	vivechak version            # Print version
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	mcputil "github.com/bhaskarjha-dev/vivechak/internal/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is set at build time via ldflags: -ldflags "-X main.version=v1.0.0"
var version = "0.0.1-dev"

func main() {
	// CRITICAL: slog MUST write to stderr; stdout is MCP protocol-only.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Parse subcommand
	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "serve", "":
		runServe(logger)
	case "version":
		fmt.Fprintf(os.Stderr, "vivechak %s\n", version)
	case "mcp-config":
		runMCPConfig()
	case "doctor":
		runDoctor()
	default:
		// If no recognized subcommand, default to serve (for MCP stdio)
		runServe(logger)
	}
}

func runServe(logger *slog.Logger) {
	server := mcputil.NewServer(version, logger)

	logger.Info("starting vivechak MCP server", "version", version)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}

