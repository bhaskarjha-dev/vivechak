# Vivechak (विवेचक) & vck

<!-- mcp-name: io.github.bhaskarjha-dev/vivechak -->

> **Evidence-grounded research for technical decisions — Autonomous Model Context Protocol (MCP) server.**

Official Python binary runner for **Vivechak** and shorthand alias **`vck`**. Automatically fetches and caches the pre-compiled native Go binary for your OS and architecture (`darwin`, `linux`, `windows`, `freebsd`).

---

## Quickstart

### 1. Run Directly via `uvx` (Zero Pre-Installation)

```bash
# Run MCP server over stdio for AI hosts
uvx vivechak serve
# or using official shorthand alias
uvx vck serve

# Or install Universal Agent Skill (Cursor, Claude Code, Antigravity, OpenCode)
npx skills add bhaskarjha-dev/vivechak
```

### 2. Configure in Your AI Host / IDE

#### Cursor IDE (`~/.cursor/mcp.json` or `.cursor/mcp.json`)
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

#### Claude Desktop (`claude_desktop_config.json`)
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

#### VS Code (Copilot Agent Mode in `.vscode/mcp.json`)
```json
{
  "servers": {
    "vivechak": {
      "command": "uvx",
      "args": ["vck", "serve"]
    }
  }
}
```

#### Google Antigravity (`~/.gemini/config/mcp_config.json` or `.agents/mcp_config.json`)
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

---

## CLI Commands

You can run developer subcommands directly:

```bash
# Workspace integrity checker
uvx vck doctor

# Headless Phase 0 exit gate evaluation (CI-ready)
uvx vck gate --track all

# 1-second host auto-configuration (21 presets)
uvx vck setup cursor
uvx vck setup claude
uvx vck setup agy

# Output universal MCP JSON configuration
uvx vck mcp-config
```

---

## Documentation & Repository

For complete documentation, methodology, 13 MCP tools reference, and templates, visit the [Vivechak GitHub Repository](https://github.com/bhaskarjha-dev/vivechak).

## License

[MIT](https://github.com/bhaskarjha-dev/vivechak/blob/main/LICENSE) © [Bhaskar Jha](https://bhaskarjha.dev)
