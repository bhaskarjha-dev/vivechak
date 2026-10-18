# vck — Shorthand CLI for Vivechak (विवेचक)

> **Evidence-grounded research for technical decisions — Autonomous Model Context Protocol (MCP) server.**

Official shorthand alias package for **Vivechak**. Automatically proxies to `vivechak`, fetching and executing the native pre-compiled Go engine.

---

## Quickstart

### 1. Run Instantly with `npx` (Zero Pre-Installation)

```bash
# Run MCP server over stdio for AI hosts
npx -y vck serve
```

### 2. Configure in Your AI Host / IDE

Add to your MCP host configuration:

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

You can run developer subcommands directly:

```bash
# Workspace integrity checker
npx -y vck doctor

# Headless Phase 0 exit gate evaluation (CI-ready)
npx -y vck gate --track all

# 1-second host auto-configuration across 21 presets
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
