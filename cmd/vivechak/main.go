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

// version is set at build time via ldflags, e.g. -ldflags "-X main.version=vX.Y.Z"
var version = "0.1.0"

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
	case "version", "--version", "-v":
		fmt.Fprintf(os.Stdout, "vivechak %s\n", version)
	case "mcp-config":
		runMCPConfig()
	case "doctor":
		runDoctor()
	case "help", "--help", "-h":
		fmt.Fprintf(os.Stdout, "vivechak %s\n\n", version)
		fmt.Fprintf(os.Stdout, "Usage: vivechak <command>\n\n")
		fmt.Fprintf(os.Stdout, "Commands:\n")
		fmt.Fprintf(os.Stdout, "  serve        Start MCP server over stdio (default — connects to ANY MCP client)\n")
		fmt.Fprintf(os.Stdout, "  version      Print version\n")
		fmt.Fprintf(os.Stdout, "  mcp-config   Output universal MCP JSON configuration (or auto-write via --client <preset> / --path <path> --write)\n")
		fmt.Fprintf(os.Stdout, "  doctor       Check workspace integrity\n")
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "vivechak %s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: vivechak <command>\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  serve        Start MCP server over stdio (default — connects to ANY MCP client)\n")
		fmt.Fprintf(os.Stderr, "  version      Print version\n")
		fmt.Fprintf(os.Stderr, "  mcp-config   Output universal MCP JSON configuration (or auto-write via --client <preset> / --path <path> --write)\n")
		fmt.Fprintf(os.Stderr, "  doctor       Check workspace integrity\n")
		os.Exit(1)
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

