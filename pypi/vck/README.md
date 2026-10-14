# vck — Shorthand CLI for Vivechak (विवेचक)

Evidence-grounded research for technical decisions — Autonomous Model Context Protocol (MCP) server.

## Quickstart

Run directly via `uvx`:
```bash
uvx vck serve
# or
uvx vivechak serve
```

Configure in Claude Desktop, Cursor, VS Code, Windsurf, or Antigravity:
```json
{
  "mcpServers": {
    "vivechak": {
      "command": "uvx",
      "args": ["vck", "serve"]
    }
  }
}
```

For full documentation, visit [GitHub Repository](https://github.com/bhaskarjha-dev/vivechak).
