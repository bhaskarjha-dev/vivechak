# Vivechak & `vck`

<!-- mcp-name: io.github.bhaskarjha-dev/vivechak -->

**Evidence-graded technical decisions for AI coding agents — researched, challenged, and gated in Markdown before you write code.**

Official Python runner for **Vivechak** and shorthand CLI **`vck`**. Automatically downloads, SHA-256-verifies, and caches the pre-compiled static Go binary for your OS and architecture (`darwin`, `linux`, `windows`, `freebsd`) while keeping `stdout` 100% clean for MCP JSON-RPC over `stdio`.

---

## 1. Quickstart (`uvx`)

```bash
# Configure your AI coding host in one command (10 verified presets)
uvx vck setup claude-code   # or: cursor, vscode, claude, agy, windsurf, zed, roo, opencode, goose

# Optional: install as a persistent CLI tool via uv
uv tool install vivechak    # provides both `vck` and `vivechak`

# Check workspace integrity and evaluate the pre-code Phase 0 Gate (exits 1 on FAIL)
uvx vck doctor
uvx vck gate
```

---

## 2. Zero-Install Host Configuration (`uvx vck serve`)

### Claude Code (`~/.claude.json` or `.mcp.json`) · Cursor (`.cursor/mcp.json`) · Claude Desktop
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

### VS Code Copilot Agent Mode (`.vscode/mcp.json`)
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

---

## 3. Three Reversibility-Calibrated Scopes

| Scope | Example Agent Prompt | Investment | Saved Artifact in `research/` |
|---|---|---|---|
| **Comparison** | *"Use Vivechak to compare Pino vs Winston for our Node.js service."* | 1 session (~5 min) | `C-001-comparison.md` (Weighted Evaluation Matrix) |
| **Decision** | *"Use Vivechak to decide whether to migrate events from Postgres to ClickHouse."* | 1–3 sessions (~30 min) | `D-001-decision.md` (ADR) + `DECISIONS.md` |
| **Project** | *"Use Vivechak to architect our multi-tenant telemetry pipeline."* | 4–30 DAG sessions | `RESEARCH-PIPELINE.md` → `FAD.md` |

For irreversible **One-Way Door** ADRs, `vck gate` requires an adversarial red-team challenge record (`vivechak_review(mode="red_team", id="D-001")`) and an out-of-band human terminal sign-off (`vck signoff D-001 --reviewer "Your Name"`).

---

## Documentation & Source

Full documentation, methodology (`FRAMEWORK.md`), 9 MCP tool schemas, and Gate rules (`G1`–`G12`): **[github.com/bhaskarjha-dev/vivechak](https://github.com/bhaskarjha-dev/vivechak)** · MIT License © [Bhaskar Jha](https://bhaskarjha.dev).
