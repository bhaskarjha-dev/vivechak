# Vivechak (`vck`)

**Evidence-graded technical decisions for AI coding agents — researched, challenged, and gated in Markdown before you write code.**

Official Node.js runner for **Vivechak** (exposing both `vivechak` and shorthand CLI `vck`). Automatically downloads, SHA-256-verifies, and caches the pre-compiled static Go binary for your OS and architecture (`darwin`, `linux`, `windows`, `freebsd`) while keeping `stdout` 100% clean for MCP JSON-RPC over `stdio`.

---

## 1. Quickstart (`npx`)

```bash
# Configure your AI coding host in one command (10 verified presets)
npx -y vivechak setup claude-code   # or: cursor, vscode, claude, agy, windsurf, zed, roo, opencode, goose

# Optional: install the Universal Agent Skill playbook
npx skills add bhaskarjha-dev/vivechak

# Check workspace integrity and evaluate the pre-code Phase 0 Gate (exits 1 on FAIL)
npx -y vivechak doctor
npx -y vivechak gate
```

If installed globally (`npm install -g vivechak`), both `vck` and `vivechak` are linked into your `PATH`.

---

## 2. Zero-Install Host Configuration (`npx -y vivechak serve`)

If you prefer running Vivechak ephemerally without installing a local binary:

### Claude Code (`~/.claude.json` or `.mcp.json`) · Cursor (`.cursor/mcp.json`) · Claude Desktop
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

### VS Code Copilot Agent Mode (`.vscode/mcp.json`)
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

Full documentation, methodology (`FRAMEWORK.md`), 9 MCP tool schemas, and Gate rules (`G1`–`G12`): **[github.com/bhaskarjha-dev/vivechak](https://github.com/bhaskarjha-dev/vivechak)** · MIT License.
