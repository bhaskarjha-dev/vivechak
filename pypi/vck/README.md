# `vck` — Shorthand CLI for Vivechak

<!-- mcp-name: io.github.bhaskarjha-dev/vivechak -->

**Evidence-graded technical decisions for AI coding agents — researched, challenged, and gated in Markdown before you write code.**

Official shorthand companion package for [**`vivechak`**](https://pypi.org/project/vivechak/). Installing or invoking `vck` delegates directly to `vivechak`, downloading, SHA-256-verifying, and caching the native Go MCP server and CLI binary for your platform.

---

## 1. Quickstart (`uvx vck`)

```bash
# Configure your AI coding host in one command (10 verified presets)
uvx vck setup claude-code   # or: cursor, vscode, claude, agy, windsurf, zed, roo, opencode, goose

# Run workspace diagnostics and pre-code Phase 0 Gate checks (CI-ready; exits 1 on FAIL)
uvx vck doctor
uvx vck gate
```

---

## 2. MCP Host Configuration

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

## Documentation & License

Official Website: **[vivechak.bhaskarjha.dev](https://vivechak.bhaskarjha.dev)** · Documentation: **[vivechak.bhaskarjha.dev/docs](https://vivechak.bhaskarjha.dev/docs)** · Source: **[github.com/bhaskarjha-dev/vivechak](https://github.com/bhaskarjha-dev/vivechak)** · [MIT](https://github.com/bhaskarjha-dev/vivechak/blob/main/LICENSE) © [Bhaskar Jha](https://bhaskarjha.dev).
