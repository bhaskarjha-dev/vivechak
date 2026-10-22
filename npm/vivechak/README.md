# Vivechak (विवेचक) & vck

> **Evidence-grounded research for technical decisions — Autonomous Model Context Protocol (MCP) server.**

Official Node.js binary runner for **Vivechak** and shorthand alias **`vck`**. Automatically fetches and caches the pre-compiled, zero-dependency native Go binary for your OS and architecture (`darwin`, `linux`, `windows`, `freebsd`).

---

## Quickstart

### 1. Run Instantly with `npx` (Zero Pre-Installation)

```bash
# Run MCP server over stdio for AI hosts
npx -y vivechak serve
# or using official 3-letter shorthand
npx -y vck serve

# Or install Universal Agent Skill (Cursor, Claude Code, Antigravity, OpenCode)
npx skills add bhaskarjha-dev/vivechak
```

### 2. Configure in Your AI Host / IDE

#### Cursor IDE (`~/.cursor/mcp.json` or `.cursor/mcp.json`)
```json
{
  "mcpServers": {
    "vivechak": {
      "command": "npx",
      "args": ["-y", "vck", "serve"]
    }
  }
}
```

#### Claude Desktop (`claude_desktop_config.json`)
```json
{
  "mcpServers": {
    "vivechak": {
      "command": "npx",
      "args": ["-y", "vck", "serve"]
    }
  }
}
```

#### VS Code (Copilot Agent Mode in `.vscode/mcp.json`)
```json
{
  "servers": {
    "vivechak": {
      "command": "npx",
      "args": ["-y", "vck", "serve"]
    }
  }
}
```

#### Google Antigravity (`~/.gemini/config/mcp_config.json` or `.agents/mcp_config.json`)
```json
{
  "mcpServers": {
    "vivechak": {
      "command": "npx",
      "args": ["-y", "vck", "serve"]
    }
  }
}
```

---

## CLI Commands

You can also run developer subcommands directly:

```bash
# Workspace integrity checker
npx -y vck doctor

# Headless Phase 0 exit gate evaluation (CI-ready)
npx -y vck gate --track all

# 1-second host auto-configuration (21 presets)
npx -y vck setup cursor
npx -y vck setup claude
npx -y vck setup agy

# Output universal MCP JSON configuration
npx -y vck mcp-config
```

---

## Documentation & Repository

For complete documentation, methodology, 13 MCP tools reference, and templates, visit the [Vivechak GitHub Repository](https://github.com/bhaskarjha-dev/vivechak).

## License

[MIT](https://github.com/bhaskarjha-dev/vivechak/blob/main/LICENSE) © [Bhaskar Jha](https://bhaskarjha.dev)
