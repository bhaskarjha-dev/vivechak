// Vivechak MCP Server — Evidence-grounded research for technical decisions.
//
// Usage:
//
//	vivechak                    # Start MCP server on stdio (default)
//	vivechak serve              # Explicit serve subcommand
//	vivechak mcp-config         # Universal MCP config (or --path <path> / --preset <shortcut> --write)
//	vivechak doctor             # Check workspace integrity
//	vivechak version            # Print version
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	mcputil "github.com/bhaskarjha-dev/vivechak/internal/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is set at build time via ldflags, e.g. -ldflags "-X main.version=vX.Y.Z".
// During development (go run/go build without ldflags), this default is used.
// During release, GoReleaser injects the git tag version via ldflags.
var version = "0.1.0-dev"

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "vivechak %s\n\n", version)
	fmt.Fprintf(w, "Usage: vivechak <command>\n\n")
	fmt.Fprintf(w, "Commands:\n")
	fmt.Fprintf(w, "  serve        Start MCP server over stdio (default — connects to ANY MCP client)\n")
	fmt.Fprintf(w, "  version      Print version\n")
	fmt.Fprintf(w, "  mcp-config   Output universal MCP JSON configuration (or write via --path <path> / --preset <shortcut> --write)\n")
	fmt.Fprintf(w, "  doctor       Check workspace integrity\n")
}

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
		printUsage(os.Stdout)
		os.Exit(0)
	default:
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func runServe(logger *slog.Logger) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := mcputil.NewServer(version, logger)

	logger.Info("starting vivechak MCP server", "version", version)
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil && err != context.Canceled {
		logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}
