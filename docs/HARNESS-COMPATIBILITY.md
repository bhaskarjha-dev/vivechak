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
| **Single-Agent IDEs** | Cursor, VS Code, Google Antigravity, Windsurf, Zed, Void Editor, ByteDance Trae | ✅ Full (Production) | `vck setup` (21 presets) or manual config | Native tool calling, live diffs, inline research reviews. |
| **CLI Terminal Agents** | Claude Code, OpenCode, Block Goose, `agy` CLI, OpenHands, Aider | ✅ Full (Production) | `vck setup <preset>` or stdio launch | High-speed terminal workflows; supports script piping. |
| **Open-Source Code Assistants** | Continue.dev (VS Code & JetBrains), Cline, Roo Code | ✅ Full (Production) | `vck setup continue`, `vck setup cline`, etc. | Modular MCP configuration; zero cloud lock-in. |
| **Cloud Background Agents** | Devin, Factory Droid, GitHub Workspaces | ✅ Full (Production) | Pre-installed binary via `install.sh` + `vck mcp-config` | Autonomous headless execution through Phase 0 Gate. |
| **Parallel-Agent Swarms** | Subagent spawning (Antigravity subagents, Claude Code task agents) | ✅ Full (Production) | Inherited MCP server or spawned processes | Safe concurrent writes via advisory locking; `⚡` parallel hints. |
| **Multi-Agent Frameworks** | CrewAI, LangGraph, AutoGen | ✅ Compatible | Custom MCP client connector | Map research sessions to agent roles (e.g., Critic runs `challenge`). |
| **CI/CD Pipelines** | GitHub Actions, GitLab CI | ✅ Compatible | Headless CLI (`vck doctor`, `vivechak_run_gate`) | Fail CI builds if Phase 0 Gate criteria are not satisfied. |
| **Agent SDKs & Runtimes** | Google Genkit, Microsoft ADK, Mastra | ✅ Compatible | SDK MCP client connectors | Programmatic pipeline execution and decision recording. |
| **Web-Based Agents** | LibreChat, ChatGPT (Web MCP), Gemini Advanced | ⚠️ Rollout-Dependent | Web MCP connector / `librechat.yaml` | Functionality matches host's MCP rollout status. |

---

## 3. Host Setup & Configuration

Vivechak provides instant auto-configuration for 21 leading agent environments via `vck setup`.

### Quick Setup

```bash
# Auto-detect all installed harnesses and configure MCP automatically
vck setup

# Or target a specific host harness:
vck setup cursor                             # Cursor IDE
vck setup claude-code                        # Claude Code CLI (~/.claude.json or .mcp.json)
vck setup mcp                                # Universal portable standard (.mcp.json)
vck setup opencode                           # OpenCode terminal agent
vck setup continue                           # Continue.dev (.continue/mcpServers/)
vck setup goose                              # Block Goose (config.yaml)
vck setup copilot                            # GitHub Copilot CLI (~/.copilot/)
vck setup void                               # Void Editor (.void/ or ~/.void-editor/)
vck setup claude-desktop                     # Claude Desktop GUI
vck setup cline                              # Cline VS Code extension
vck setup windsurf                           # Windsurf Editor
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

### Universal Portable MCP (`.mcp.json`)
The emerging root standard recognized by Claude Code, Cursor, Windsurf, and other modern harnesses. Place at project root:
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

### Claude Code
Claude Code supports both user-global (`~/.claude.json`) and project-root (`.mcp.json`) configurations. To configure globally:
```bash
vck setup claude-code
```
Or manually in `~/.claude.json` (state and session metadata are preserved):
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

### OpenCode
OpenCode requires the root `"mcp"` key and an array-formatted command with local execution flags. Configure via `vck setup opencode` or in `opencode.json`:
```json
{
  "mcp": {
    "vivechak": {
      "command": ["vck", "serve"],
      "type": "local",
      "enabled": true
    }
  }
}
```

### Continue.dev
Continue loads modular MCP server definition files from `.continue/mcpServers/`. Configure via `vck setup continue` or create `.continue/mcpServers/vivechak.json`:
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

### Block Goose
Block Goose uses YAML-based configuration with stdio extensions. Configure via `vck setup goose` or add to `~/.config/goose/config.yaml`:
```yaml
extensions:
  vivechak:
    cmd: vck
    args:
      - serve
    type: stdio
    enabled: true
```

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

### Void Editor & GitHub Copilot CLI
Add to `.void/mcp.json` (Void) or `~/.copilot/mcp-config.json` (Copilot):
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

### Google Antigravity
Vivechak binary is detected on `PATH`. Add to `.gemini/antigravity-ide/mcp_config.json`:
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
