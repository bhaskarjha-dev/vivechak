# Host Setup Guide

Configure Vivechak as an MCP server in your AI host.

Vivechak is a **universal Model Context Protocol (MCP) server**. It communicates over standard `stdio` using JSON-RPC 2.0. Any IDE, agent CLI, or runtime that supports the open MCP standard can connect to Vivechak out of the box.

> **💡 Shorthand CLI Alias (`vck`):** Vivechak installs the official 3-letter shorthand **`vck`** alongside `vivechak`. Both names are 100% identical and interchangeable across all subcommands (`vck setup`, `vck mcp-config`, `vck version`, `vck doctor`, `vck serve`).

---

## 1. Quick Host Setup (`vck setup`)

The fastest way to configure your AI host is the streamlined **`setup`** subcommand (aliased as `install`):

```sh
# Configure any supported desktop host or harness in 1 second:
vck setup cursor                               # Cursor IDE
vck setup vscode                               # VS Code (Copilot Agent mode)
vck setup claude                               # Claude Desktop
vck setup agy                                  # Google Antigravity (IDE, 2.0, CLI)

# Inside a project with an active host folder (.cursor, .agents, .vscode, .trae, etc.):
vck setup                                      # Auto-detects workspace host!

# Target any custom JSON file directly:
vck setup /path/to/mcp-settings.json

# Preview changes without modifying files:
vck setup cursor --dry-run
```

`vck setup` automatically resolves the target configuration file, detects whether global or workspace scope is active, formats Windows/macOS/Linux paths cleanly, and merges the Vivechak MCP server block without overwriting existing servers.

---

## 2. Universal Setup & Manual Configuration

### Universal MCP JSON (`vck mcp-config`)

To generate the exact JSON snippet containing the absolute path of your current binary (for any host or agent that supports standard MCP), run:

```sh
vck mcp-config
```

Output:
```json
{
  "mcpServers": {
    "vivechak": {
      "command": "/usr/local/bin/vivechak",
      "args": ["serve"]
    }
  }
}
```

### Supported Desktop Hosts & Harness Presets

| Environment / Harness | Fast Setup Command | Preset Shortcut | Default Config Location | Schema Key |
|---|---|---|---|---|
| **Cursor** | `vck setup cursor` | `cursor` | `~/.cursor/mcp.json` (global) or `.cursor/mcp.json` (workspace) | `mcpServers` |
| **VS Code** | `vck setup vscode` | `vscode` | `.vscode/mcp.json` (workspace) | `servers` |
| **Claude Desktop** | `vck setup claude` | `claude`, `claude-desktop` | `claude_desktop_config.json` (OS-specific application storage) | `mcpServers` |
| **Windsurf** | `vck setup windsurf` | `windsurf` | `~/.codeium/windsurf/mcp_config.json` | `mcpServers` |
| **Google Antigravity** | `vck setup agy` | `agy`, `antigravity` | `~/.gemini/config/mcp_config.json` (global) or `.agents/mcp_config.json` (workspace) | `mcpServers` |
| **Zed Editor** | `vck setup zed` | `zed` | `~/.config/zed/settings.json` (or `%APPDATA%\Zed\settings.json`) | `context_servers` |
| **AWS Kiro** | `vck setup kiro` | `kiro` | `~/.kiro/settings/mcp.json` (global) or `.kiro/settings/mcp.json` (workspace) | `mcpServers` |
| **ByteDance Trae** | `vck setup trae` | `trae` | `.trae/mcp.json` (workspace) or `~/.trae/mcp.json` (global) | `mcpServers` |
| **Oh My Pi (OMP)** | `vck setup omp` | `omp` | `~/.omp/agent/mcp.json` (global) or `.omp/mcp.json` (workspace) | `mcpServers` |
| **OpenHands (OpenDevin)** | `vck setup openhands` | `openhands` | `~/.openhands/mcp.json` | `mcpServers` |
| **Factory Droid** | `vck setup droid` | `droid` | `.factory/mcp.json` (workspace) or `~/.factory/mcp.json` (global) | `mcpServers` |
| **Cline (VS Code)** | `vck setup cline` | `cline` | `saoudrizwan.claude-dev/.../cline_mcp_settings.json` (or `.cline/mcp.json`) | `mcpServers` |
| **Roo Code (VS Code)** | `vck setup roo` | `roo` | `rooveterinaryinc.roo-cline/.../cline_mcp_settings.json` (or `.roo/mcp.json`) | `mcpServers` |
| **Devin (Cognition)** | `vck setup devin` | `devin` | `.devin/mcp_config.local.json` | `mcpServers` |

*(The legacy syntax `vck mcp-config --preset <shortcut> --write` and `vck mcp-config --path <file> --write` remains 100% backward-compatible for scripts and automation).*

> **Protocol Neutrality Note:** These presets are strictly convenience shortcuts for filesystem path resolution across macOS, Windows, and Linux. At runtime, every MCP harness and editor interacts with Vivechak through the exact same JSON-RPC 2.0 wire protocol over `stdio`.

---

## 3. Desktop IDE Quick Reference

The following subsections document the default locations and shortcut commands for desktop editors:

### Cursor

#### Automatic
```sh
vck setup cursor
# or: vivechak mcp-config --preset cursor --write
```

#### Manual
Edit `~/.cursor/mcp.json` (or `.cursor/mcp.json` in your workspace):

```json
{
  "mcpServers": {
    "vivechak": {
      "command": "/usr/local/bin/vivechak",
      "args": ["serve"]
    }
  }
}
```

| OS | Config Path |
|---|---|
| macOS | `~/.cursor/mcp.json` |
| Linux | `~/.cursor/mcp.json` |
| Windows | `%USERPROFILE%\.cursor\mcp.json` |

---

### VS Code

Native VS Code (via GitHub Copilot Agent mode) discovers MCP servers defined in your project's `.vscode/mcp.json`. Note that **native VS Code uses the `"servers"` key** rather than `"mcpServers"`.

#### Automatic
```sh
vck setup vscode
# or: vivechak mcp-config --preset vscode --write
```

#### Manual
Edit `.vscode/mcp.json` in your workspace root:

```json
{
  "servers": {
    "vivechak": {
      "command": "/usr/local/bin/vivechak",
      "args": ["serve"]
    }
  }
}
```

> **Note for Extension Users (Cline, Roo Code, Continue):** If you use VS Code extensions like Cline or Roo Code, they use the generic `"mcpServers"` format in their extension settings tabs. You can paste the output of `vck mcp-config` directly into their MCP settings panel.

---

### Claude Desktop

#### Automatic
```sh
vck setup claude
# or: vivechak mcp-config --preset claude-desktop --write
```

#### Manual
Edit `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "vivechak": {
      "command": "/usr/local/bin/vivechak",
      "args": ["serve"]
    }
  }
}
```

| OS | Config Path |
|---|---|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |

> **⚠️ macOS Note:** Claude Desktop runs as a GUI application with a truncated `$PATH` (`/usr/bin:/bin:/usr/sbin:/sbin`). The `command` field **must** use an absolute path to the Vivechak binary. Running `vck setup claude` handles this automatically.

---

### Windsurf

#### Automatic
```sh
vck setup windsurf
# or: vivechak mcp-config --preset windsurf --write
```

#### Manual
Edit `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "vivechak": {
      "command": "/usr/local/bin/vivechak",
      "args": ["serve"]
    }
  }
}
```

| OS | Config Path |
|---|---|
| macOS | `~/.codeium/windsurf/mcp_config.json` |
| Linux | `~/.codeium/windsurf/mcp_config.json` |
| Windows | `%USERPROFILE%\.codeium\windsurf\mcp_config.json` |

> **Tip:** After editing the configuration, click the **Refresh** button in Windsurf's Cascade MCP panel to load the server.

---

### Google Antigravity (Antigravity IDE, Antigravity 2.0, & Antigravity CLI `agy`)

Antigravity discovers MCP servers via **`mcp_config.json`**. This single configuration is shared seamlessly across all Antigravity surfaces:
- **Antigravity IDE**: Standalone in-editor AI pair programmer (VS Code-based).
- **Antigravity 2.0**: The desktop application (chat canvas, agent orchestrator).
- **Antigravity CLI (`agy`)**: The lightweight terminal interface / TUI.

#### Automatic Setup
Run either shortcut (both `vck setup agy` and `vck setup antigravity` are supported):
```sh
vck setup agy
# or: vivechak mcp-config --preset antigravity --write
```
*Writes to `~/.gemini/config/mcp_config.json` (global, available across all projects and CLI sessions) or `.agents/mcp_config.json` if run inside a project with an `.agents/` directory.*

#### Antigravity CLI (`agy`) Native Command
You can also register Vivechak directly via the `agy` CLI command:
```sh
agy mcp add vivechak /path/to/vivechak
```
*(Or inspect configured servers in `agy` via `agy mcp list`, or press `/mcp` inside the interactive TUI).*

#### Manual JSON Configuration
Add Vivechak to `~/.gemini/config/mcp_config.json` (global) or `<project-root>/.agents/mcp_config.json` (workspace):

```json
{
  "mcpServers": {
    "vivechak": {
      "command": "/usr/local/bin/vivechak",
      "args": ["serve"]
    }
  }
}
```

| Scope | Location | Applies To |
|---|---|---|
| **Global** | `~/.gemini/config/mcp_config.json` | All Antigravity surfaces (IDE, 2.0, CLI `agy`) across all projects |
| **Workspace** | `.agents/mcp_config.json` | Project-specific workspace |

> **Tip:** In Antigravity IDE or Antigravity 2.0, you can verify active MCP servers anytime in the chat panel via **"..." (More Options) > "Manage MCP Servers" > "View raw config"**. Inside the `agy` terminal, use `/mcp` or `agy mcp list`.
>
> **Workspace Location Best Practice:** When prompting agents in IDE environments, ensure agents pass `project_root` explicitly to `vivechak_init` (e.g. `project_root: "d:/path/to/project"`). Vivechak includes active guards against host CWD inheritance and rejects any attempts to initialize workspaces inside application program folders (`AppData\Local\Programs\...`). You may also optionally supply `"env": { "VIVECHAK_PROJECT_ROOT": "${workspaceFolder}" }` in workspace configurations.

---

### Zed Editor

Zed supports MCP natively via its Assistant panel and sandboxes MCP servers. Zed uses the **`"context_servers"`** configuration key.

#### Automatic
```sh
vck setup zed
# or: vivechak mcp-config --preset zed --write
```

#### Manual
Edit `settings.json` (accessible via Command Palette: `zed: open settings`):

```json
{
  "context_servers": {
    "vivechak": {
      "command": "/usr/local/bin/vivechak",
      "args": ["serve"]
    }
  }
}
```

| OS | Config Path |
|---|---|
| macOS | `~/.config/zed/settings.json` |
| Linux | `~/.config/zed/settings.json` |
| Windows | `%APPDATA%\Zed\settings.json` |

---

### AWS Kiro

AWS Kiro supports MCP servers natively at both user and workspace levels.

#### Automatic
```sh
vck setup kiro
# or: vivechak mcp-config --preset kiro --write
```
*Auto-detects: If `.kiro` directory exists in the workspace, writes to `.kiro/settings/mcp.json`. Otherwise writes to global `~/.kiro/settings/mcp.json` (or `~/.aws/.kiro/mcp.json` if existing).*

#### Manual
Edit `~/.kiro/settings/mcp.json` (global) or `.kiro/settings/mcp.json` (workspace):

```json
{
  "mcpServers": {
    "vivechak": {
      "command": "/usr/local/bin/vivechak",
      "args": ["serve"]
    }
  }
}
```

> **Tip:** You can verify your active MCP servers in Kiro chat by running `/mcp`.

---

### ByteDance Trae IDE

Trae reads project-scoped MCP configurations from `.trae/mcp.json`.

```sh
vck setup trae
# or: vivechak mcp-config --preset trae --write
```
*Auto-writes to `.trae/mcp.json` in your workspace (or `~/.trae/mcp.json` if global).*

---

### Oh My Pi (OMP)

Oh My Pi provides native terminal MCP agent orchestration:

```sh
vck setup omp
# or: vivechak mcp-config --preset omp --write
```
*Writes to `~/.omp/agent/mcp.json` (or workspace `.omp/mcp.json` if `.omp` exists). In the OMP terminal, use `/mcp` to inspect loaded servers.*

---

### OpenHands (OpenDevin)

OpenHands CLI and Agent Canvas read MCP server definitions from `~/.openhands/mcp.json`:

```sh
vck setup openhands
# or: vivechak mcp-config --preset openhands --write
```
*Or register directly via OpenHands CLI:*
```sh
openhands mcp add vivechak --transport stdio /path/to/vivechak serve
```

---

### Factory Droid

Factory Droid connects via `~/.factory/mcp.json` or `.factory/mcp.json`:

```sh
vck setup droid
# or: vivechak mcp-config --preset droid --write
```
*Or register via Droid CLI:*
```sh
droid mcp add vivechak /path/to/vivechak
```

---

### Cline & Roo Code (VS Code Extensions)

Both Cline and Roo Code store global MCP server settings in VS Code extension storage (`saoudrizwan.claude-dev` and `rooveterinaryinc.roo-cline`):

```sh
# For Cline:
vck setup cline
# or: vivechak mcp-config --preset cline --write

# For Roo Code:
vck setup roo
# or: vivechak mcp-config --preset roo --write
```
*Auto-resolves the correct OS-specific VS Code storage directory on Windows, macOS, and Linux (or workspace `.cline/mcp.json` / `.roo/mcp.json` if present).*

---

### Cognition Devin

Devin local and CLI configurations read from `.devin/mcp_config.local.json`:

```sh
vck setup devin
# or: vivechak mcp-config --preset devin --write
```

---

### CLI-Native Agent Harnesses

For terminal agents with built-in MCP management commands, you can register Vivechak directly without manually touching configuration files:

- **Claude Code**:
  ```sh
  claude mcp add vivechak -- /path/to/vivechak serve
  ```
- **xAI Grok Build**:
  ```sh
  grok mcp add vivechak -- /path/to/vivechak serve
  ```
- **GitHub Copilot CLI**:
  ```sh
  copilot mcp add vivechak /path/to/vivechak serve
  ```
- **Block Goose**:
  ```sh
  goose configure
  # Select stdio -> name: vivechak -> command: /path/to/vivechak -> args: serve
  ```
- **Sourcegraph Amp**:
  ```sh
  amp mcp add vivechak /path/to/vivechak
  ```

---

## 4. Graphical & Interactive Settings

For clients or applications that configure MCP servers through an interactive settings interface (such as desktop AI apps or custom agent dashboards), add the server directly via their configuration panel:

- **Name:** `vivechak`
- **Command:** Absolute path to `vivechak` (run `vck mcp-config` to output the exact binary path)
- **Arguments:** `serve`
- **Transport:** `stdio`

---

## 5. Windows Configuration

On Windows, use escaped backslashes (`\\`) in configuration files:

```json
{
  "mcpServers": {
    "vivechak": {
      "command": "C:\\Users\\you\\AppData\\Local\\Programs\\vivechak\\vivechak.exe",
      "args": ["serve"]
    }
  }
}
```

The `vck setup` (and `vck mcp-config --write`) command automatically detects and formats Windows paths correctly.

---

## 6. Troubleshooting & Diagnostics

### Binary Not Found
Make sure the `command` field uses an **absolute path**. GUI applications on macOS do not inherit terminal `$PATH`.

### Verify Installation & Workspace
```sh
vck version
vck doctor /path/to/your/project
```

### Test MCP Connection Interactively
Verify that the Vivechak MCP server responds over stdio using the official MCP Inspector:
```sh
npx @modelcontextprotocol/inspector vck serve
```
