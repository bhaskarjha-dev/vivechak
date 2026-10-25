# Vivechak (विवेचक)

> **Evidence-grounded research for technical decisions — Autonomous Model Context Protocol (MCP) server.**

Official Node.js binary runner for **Vivechak** (providing CLI binaries `vivechak` and alias `vck`). Automatically fetches and caches the pre-compiled, zero-dependency native Go binary for your OS and architecture (`darwin`, `linux`, `windows`, `freebsd`).

---

## Quickstart

### 1. Run Instantly with `npx` (Zero Pre-Installation)

```bash
# Run MCP server over stdio for AI hosts
npx -y vivechak serve

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
      "args": ["-y", "vivechak", "serve"]
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
      "args": ["-y", "vivechak", "serve"]
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
      "args": ["-y", "vivechak", "serve"]
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
      "args": ["-y", "vivechak", "serve"]
    }
  }
}
```

---

## CLI Commands

You can run developer subcommands directly with `npx`:

```bash
# Workspace integrity checker
npx -y vivechak doctor

# Headless Phase 0 exit gate evaluation (CI-ready)
npx -y vivechak gate --track all

# 1-second host auto-configuration (21 presets)
npx -y vivechak setup cursor
npx -y vivechak setup claude
npx -y vivechak setup agy

# Output universal MCP JSON configuration
npx -y vivechak mcp-config
```

*(Note: If you install globally via `npm install -g vivechak`, both `vivechak` and shorthand `vck` binaries are linked into your `$PATH`, allowing you to run `vck doctor`, `vck gate`, etc. directly).*

---

## Documentation & Repository

For complete documentation, methodology, MCP tools reference, and templates, visit the [Vivechak GitHub Repository](https://github.com/bhaskarjha-dev/vivechak).

## License

[MIT](https://github.com/bhaskarjha-dev/vivechak/blob/main/LICENSE) © [Bhaskar Jha](https://bhaskarjha.dev)
