// Vivechak MCP Server — Evidence-grounded research for technical decisions.
//
// Usage:
//
//	vck setup <host|path>       # Set up Vivechak in an AI host or config file
//	vck                         # Start MCP server on stdio (default)
//	vck serve                   # Explicit serve subcommand
//	vck doctor                  # Check workspace integrity
//	vck mcp-config              # Universal MCP config (pipeable)
//	vck version                 # Print version
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
	fmt.Fprintf(w, "vivechak (vck) %s\n\n", version)
	fmt.Fprintf(w, "Usage: vivechak <command>  (or: vck <command>)\n\n")
	fmt.Fprintf(w, "Commands:\n")
	fmt.Fprintf(w, "  setup <host|path>  Set up Vivechak in an AI host or config file (alias: install)\n")
	fmt.Fprintf(w, "  serve              Start MCP server over stdio (default — connects to ANY MCP client)\n")
	fmt.Fprintf(w, "  doctor             Check workspace integrity\n")
	fmt.Fprintf(w, "  mcp-config         Output universal MCP JSON configuration (pipeable)\n")
	fmt.Fprintf(w, "  version            Print version\n")
}

func main() {
	// CRITICAL: slog MUST write to stderr; stdout is MCP protocol-only.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	os.Exit(runCLI(os.Args, os.Stdout, os.Stderr, logger))
}

func runCLI(args []string, stdout, stderr io.Writer, logger ...*slog.Logger) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return runCLIWithContext(ctx, args, stdout, stderr, logger...)
}

func runCLIWithContext(ctx context.Context, args []string, stdout, stderr io.Writer, logger ...*slog.Logger) int {
	// Parse subcommand
	cmd := "serve"
	if len(args) > 1 {
		cmd = args[1]
	}

	switch cmd {
	case "serve", "":
		var l *slog.Logger
		if len(logger) > 0 && logger[0] != nil {
			l = logger[0]
		} else {
			l = slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
		}
		return runServeWithContext(ctx, l)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "vivechak %s\n", version)
		return 0
	case "setup", "install":
		var subArgs []string
		if len(args) > 2 {
			subArgs = args[2:]
		}
		return runSetupWithArgs(subArgs, stdout, stderr)
	case "mcp-config":
		var subArgs []string
		if len(args) > 2 {
			subArgs = args[2:]
		}
		return runMCPConfigWithArgs(subArgs, stdout, stderr)
	case "doctor":
		var subArgs []string
		if len(args) > 2 {
			subArgs = args[2:]
		}
		return runDoctorWithArgs(subArgs, stdout, stderr)
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	default:
		printUsage(stderr)
		return 1
	}
}

func runServe(logger *slog.Logger) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return runServeWithContext(ctx, logger)
}

func runServeWithContext(ctx context.Context, logger *slog.Logger) int {
	server := mcputil.NewServer(version, logger)

	logger.Info("starting vivechak MCP server", "version", version)
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil && err != context.Canceled {
		logger.Error("server exited with error", "err", err)
		return 1
	}
	return 0
}
