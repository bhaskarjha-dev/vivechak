# Vivechak — Agent Harness Compatibility Matrix
### Universal Integration Guide across IDEs, CLI Agents, Cloud Runtimes, and Agent Swarms

---

## 1. Overview

Vivechak is built from the ground up to be **harness-agnostic**. It communicates exclusively via the standard [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) over standard input/output (`stdio`). Any AI agent environment that supports MCP can orchestrate Vivechak research pipelines natively.

```
┌────────────────────────────────────────────────────────┐
│ Host Harness (Cursor, Claude Code, Antigravity, etc.)  │
└──────────────────────────┬─────────────────────────────┘
                           │ Standard MCP (JSON-RPC over stdio)
┌──────────────────────────▼─────────────────────────────┐
│ vivechak MCP Server (Single Go Binary)                 │
│  - 13 atomic tools                                     │
│  - Advisory file locking & atomic writes               │
│  - Workspace isolation via Go 1.25+ os.Root            │
└──────────────────────────┬─────────────────────────────┘
                           │ Filesystem I/O
┌──────────────────────────▼─────────────────────────────┐
│ Project Workspace (research/ directory)                │
└────────────────────────────────────────────────────────┘
```

---

## 2. Compatibility Matrix

| Harness Category | Representative Examples | Compatibility | Configuration Method | Key Capabilities & Notes |
|---|---|---|---|---|
| **Single-Agent IDEs** | Cursor, VS Code (Cline/Roo-Code), Claude Desktop, Google Antigravity, Windsurf | ✅ Full (Production) | `vck setup` (14 presets) or manual config | Native tool calling, live diffs, inline research reviews. |
| **CLI Terminal Agents** | Claude Code, `agy` CLI, OpenHands, Aider | ✅ Full (Production) | Direct stdio launch (`vivechak serve`) | High-speed terminal workflows; supports script piping. |
| **Cloud Background Agents** | Devin, Factory Droid, GitHub Workspaces | ✅ Full (Production) | Pre-installed binary via `install.sh` + `vck mcp-config` | Autonomous headless execution through Phase 0 Gate. |
| **Parallel-Agent Swarms** | Subagent spawning (Antigravity subagents, Claude Code task agents) | ✅ Full (Production) | Inherited MCP server or spawned processes | Safe concurrent writes via advisory locking; `⚡` parallel hints. |
| **Multi-Agent Frameworks** | CrewAI, LangGraph, AutoGen | ✅ Compatible | Custom MCP client connector | Map research sessions to agent roles (e.g., Critic runs `challenge`). |
| **CI/CD Pipelines** | GitHub Actions, GitLab CI | ✅ Compatible | Headless CLI (`vck doctor`, `vivechak_run_gate`) | Fail CI builds if Phase 0 Gate criteria are not satisfied. |
| **Agent SDKs & Runtimes** | Google Genkit, Microsoft ADK, Mastra | ✅ Compatible | SDK MCP client connectors | Programmatic pipeline execution and decision recording. |
| **Web-Based Agents** | ChatGPT (Web MCP), Gemini Advanced | ⚠️ Rollout-Dependent | Web MCP connector (when host supports local stdio bridge) | Functionality matches host's MCP rollout status. |

---

## 3. Host Setup & Configuration

Vivechak provides instant auto-configuration for 14 leading agent environments via `vck setup`.

### Quick Setup

```bash
# Auto-detect all installed harnesses and configure MCP automatically
vck setup

# Or target a specific host harness:
vck setup cursor
vck setup claude-desktop
vck setup cline
vck setup windsurf
```

### Universal JSON Configuration (`vck mcp-config`)

If your harness uses a custom JSON configuration file, run:

```bash
vck mcp-config
```

This outputs a drop-in configuration block:

```json
{
  "mcpServers": {
    "vivechak": {
      "command": "vivechak",
      "args": ["serve"]
    }
  }
}
```

---

## 4. Concurrency & Parallel Execution Model

Frontier agent harnesses frequently spawn parallel subagents to execute multiple research tasks simultaneously. Vivechak is engineered for high-concurrency safety:

1. **Advisory File Locks (`store.LockFile`):**
   - Every mutating operation (`save_session`, `record_decision`, `amend_session`, `replan`) acquires an advisory lock on the target file with exponential backoff and timeouts.
   - Recording decisions acquires an exclusive advisory lock on `research/DECISIONS.md` to prevent compilation race conditions.
2. **Atomic Writes (`store.WriteFileAtomic`):**
   - Files are written to temporary files and atomically renamed into place.
   - Employs a process-wide monotonic counter combined with high-resolution timestamps to guarantee zero collisions on Windows filesystem handles.
3. **Directory Handles Confinement (`os.Root`):**
   - All workspace reads and writes are strictly confined within the workspace root, eliminating path-traversal vulnerabilities across concurrent worker processes.

---

## 5. Harness-Specific Integration Recipes

### Cursor
Add to `.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "vivechak": {
      "command": "vck",
      "args": ["serve"]
    }
  }
}
```

### Claude Desktop
Add to `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "vivechak": {
      "command": "vck",
      "args": ["serve"]
    }
  }
}
```

### Google Antigravity / Claude Code
Vivechak binary is detected on `PATH`. Add to your global or project agent MCP settings:
```json
{
  "mcpServers": {
    "vivechak": {
      "command": "vck",
      "args": ["serve"]
    }
  }
}
```

### Headless GitHub Actions CI
```yaml
name: Verify Phase 0 Gate
on: [pull_request]
jobs:
  gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Vivechak
        run: curl -sSL https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.sh | bash
      - name: Run Workspace Doctor
        run: vck doctor
```
