# Host Setup Guide

Configure Vivechak as an MCP server in your AI host.

Vivechak is a **universal Model Context Protocol (MCP) server**. It communicates over standard `stdio` using JSON-RPC 2.0. Any IDE, agent CLI, or runtime that supports the open MCP standard can connect to Vivechak out of the box.

---

## 1. The Universal Setup (Any MCP Host)

To connect Vivechak to **any** MCP-compliant host (including terminal agents, Neovim, Emacs, LibreChat, or custom runtimes), simply configure the server command:

- **Command:** `vivechak` (or absolute path, e.g. `/usr/local/bin/vivechak` or `C:\Users\you\AppData\Local\Programs\vivechak\vivechak.exe`)
- **Arguments:** `["serve"]`
- **Transport:** `stdio`

To generate the exact JSON snippet containing the absolute path of your current binary, run:

```sh
vivechak mcp-config
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

### Connecting Any Agent Harness (Cline, OpenCode, OMP, Claude Code, Goose, Droid, etc.)

Agent harnesses connect to MCP servers using one of three standard integration patterns:

| Integration Pattern | Harness Examples | How to Connect |
|---|---|---|
| **Preset Shortcut (`--preset`)** | **Cursor**, **VS Code**, **Claude Desktop**, **Windsurf**, **Antigravity**, **Zed**, **Kiro**, **Trae**, **OMP**, **OpenHands**, **Factory Droid**, **Cline**, **Roo Code**, **Devin** | Run `vivechak mcp-config --preset <name> --write` to automatically resolve and merge into the target file. |
| **Universal JSON Path (`--path`)** | **OpenCode**, **Continue**, **Warp**, **Qwen Code**, **Kilo Code**, custom pipelines | Run `vivechak mcp-config --path <file> --write` (supports `.mcp.json`, `opencode.json`, `config.json`, etc.). |
| **CLI Registration Command** | **Antigravity CLI (`agy`)**, **Claude Code**, **Block Goose**, **Grok Build**, **Factory Droid**, **OpenHands CLI**, **Copilot CLI**, **Sourcegraph Amp** | Register via the agent's CLI:<br>• Antigravity CLI: `agy mcp add vivechak /path/to/vivechak`<br>• Claude Code: `claude mcp add vivechak -- /path/to/vivechak serve`<br>• Grok Build: `grok mcp add vivechak -- /path/to/vivechak serve`<br>• Factory Droid: `droid mcp add vivechak /path/to/vivechak`<br>• OpenHands CLI: `openhands mcp add vivechak --transport stdio /path/to/vivechak serve`<br>• Copilot CLI: `copilot mcp add vivechak /path/to/vivechak serve`<br>• Sourcegraph Amp: `amp mcp add vivechak /path/to/vivechak`<br>• Goose: `goose configure` (add extension → `stdio` → command: `vivechak`, args: `serve`) |
| **Interactive Settings Panel** | **Antigravity IDE / 2.0**, **Warp Terminal**, **Agent Dashboards**, **Desktop AI apps** | In the agent's MCP settings panel, add a server with Command: `/path/to/vivechak` (run `vivechak mcp-config` to copy it) and Arguments: `serve`. |

Once connected, simply prompt your agent:
> *"Use Vivechak to research the architecture for our project. We need to evaluate our primary backend and datastore trade-offs."*

---

## 2. Auto-Writing Configuration

Vivechak can automatically merge its server definition directly into existing JSON configuration files without overwriting other server settings.

### Universal File Writing (Any Agent Harness)

To configure any agent harness that reads an MCP configuration file (e.g. **OpenCode**, **Continue**, **Warp**, or custom pipelines), pass the path to its file:

```sh
vivechak mcp-config --path /path/to/mcp-settings.json --write
```

Vivechak automatically inspects the target file and applies the correct schema key:
- If the target is in `.vscode/` or already uses `"servers"`, it merges under `"servers"`.
- If the target is a Zed settings file or uses `"context_servers"`, it merges under `"context_servers"`.
- Otherwise, it merges under the standard `"mcpServers"` key.

### Built-in Desktop & Harness Path Shortcuts

Vivechak provides built-in path shortcuts (`--preset <shortcut>`) so you don't have to manually look up and type default configuration paths across macOS, Windows, and Linux:

```sh
vivechak mcp-config --preset <shortcut> --write
# (--client <shortcut> is also supported as a backward-compatible alias)
```

| Environment / Harness | Preset Shortcut | Default Config Location | Schema Key |
|---|---|---|---|
| **Cursor** | `cursor` | `~/.cursor/mcp.json` (global) or `.cursor/mcp.json` (workspace) | `mcpServers` |
| **VS Code** | `vscode` | `.vscode/mcp.json` (workspace) | `servers` |
| **Claude Desktop** | `claude-desktop` | `claude_desktop_config.json` (OS-specific application storage) | `mcpServers` |
| **Windsurf** | `windsurf` | `~/.codeium/windsurf/mcp_config.json` | `mcpServers` |
| **Google Antigravity** | `antigravity`, `agy` | `~/.gemini/config/mcp_config.json` (global) or `.agents/mcp_config.json` (workspace) | `mcpServers` |
| **Zed Editor** | `zed` | `~/.config/zed/settings.json` (or `%APPDATA%\Zed\settings.json`) | `context_servers` |
| **AWS Kiro** | `kiro` | `~/.kiro/settings/mcp.json` (global) or `.kiro/settings/mcp.json` (workspace) | `mcpServers` |
| **ByteDance Trae** | `trae` | `.trae/mcp.json` (workspace) or `~/.trae/mcp.json` (global) | `mcpServers` |
| **Oh My Pi (OMP)** | `omp` | `~/.omp/agent/mcp.json` (global) or `.omp/mcp.json` (workspace) | `mcpServers` |
| **OpenHands (OpenDevin)** | `openhands` | `~/.openhands/mcp.json` | `mcpServers` |
| **Factory Droid** | `droid` | `.factory/mcp.json` (workspace) or `~/.factory/mcp.json` (global) | `mcpServers` |
| **Cline (VS Code)** | `cline` | `saoudrizwan.claude-dev/.../cline_mcp_settings.json` (or `.cline/mcp.json`) | `mcpServers` |
| **Roo Code (VS Code)** | `roo` | `rooveterinaryinc.roo-cline/.../cline_mcp_settings.json` (or `.roo/mcp.json`) | `mcpServers` |
| **Devin (Cognition)** | `devin` | `.devin/mcp_config.local.json` | `mcpServers` |

> **Protocol Neutrality Note:** These presets are strictly convenience shortcuts for filesystem path resolution across macOS, Windows, and Linux. At runtime, every MCP harness and editor interacts with Vivechak through the exact same JSON-RPC 2.0 wire protocol over `stdio`.

---

## 3. Desktop IDE Quick Reference

The following subsections document the default locations and shortcut commands for desktop editors:

### Cursor

#### Automatic
```sh
vivechak mcp-config --preset cursor --write
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
vivechak mcp-config --preset vscode --write
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

> **Note for Extension Users (Cline, Roo Code, Continue):** If you use VS Code extensions like Cline or Roo Code, they use the generic `"mcpServers"` format in their extension settings tabs. You can paste the output of `vivechak mcp-config` directly into their MCP settings panel.

---

### Claude Desktop

#### Automatic
```sh
vivechak mcp-config --preset claude-desktop --write
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

> **⚠️ macOS Note:** Claude Desktop runs as a GUI application with a truncated `$PATH` (`/usr/bin:/bin:/usr/sbin:/sbin`). The `command` field **must** use an absolute path to the Vivechak binary. Running `vivechak mcp-config --preset claude-desktop --write` handles this automatically.

---

### Windsurf

#### Automatic
```sh
vivechak mcp-config --preset windsurf --write
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
Run either shortcut (both `--preset antigravity` and `--preset agy` are supported):
```sh
vivechak mcp-config --preset antigravity --write
# or: vivechak mcp-config --preset agy --write
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

---

### Zed Editor

Zed supports MCP natively via its Assistant panel and sandboxes MCP servers. Zed uses the **`"context_servers"`** configuration key.

#### Automatic
```sh
vivechak mcp-config --preset zed --write
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
vivechak mcp-config --preset kiro --write
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
vivechak mcp-config --preset trae --write
```
*Auto-writes to `.trae/mcp.json` in your workspace (or `~/.trae/mcp.json` if global).*

---

### Oh My Pi (OMP)

Oh My Pi provides native terminal MCP agent orchestration:

```sh
vivechak mcp-config --preset omp --write
```
*Writes to `~/.omp/agent/mcp.json` (or workspace `.omp/mcp.json` if `.omp` exists). In the OMP terminal, use `/mcp` to inspect loaded servers.*

---

### OpenHands (OpenDevin)

OpenHands CLI and Agent Canvas read MCP server definitions from `~/.openhands/mcp.json`:

```sh
vivechak mcp-config --preset openhands --write
```
*Or register directly via OpenHands CLI:*
```sh
openhands mcp add vivechak --transport stdio /path/to/vivechak serve
```

---

### Factory Droid

Factory Droid connects via `~/.factory/mcp.json` or `.factory/mcp.json`:

```sh
vivechak mcp-config --preset droid --write
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
vivechak mcp-config --preset cline --write

# For Roo Code:
vivechak mcp-config --preset roo --write
```
*Auto-resolves the correct OS-specific VS Code storage directory on Windows, macOS, and Linux (or workspace `.cline/mcp.json` / `.roo/mcp.json` if present).*

---

### Cognition Devin

Devin local and CLI configurations read from `.devin/mcp_config.local.json`:

```sh
vivechak mcp-config --preset devin --write
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
- **Command:** Absolute path to `vivechak` (run `vivechak mcp-config` to output the exact binary path)
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

The `vivechak mcp-config --write` command automatically detects and formats Windows paths correctly.

---

## 6. Troubleshooting & Diagnostics

### Binary Not Found
Make sure the `command` field uses an **absolute path**. GUI applications on macOS do not inherit terminal `$PATH`.

### Verify Installation & Workspace
```sh
vivechak version
vivechak doctor /path/to/your/project
```

### Test MCP Connection Interactively
Verify that the Vivechak MCP server responds over stdio using the official MCP Inspector:
```sh
npx @modelcontextprotocol/inspector vivechak serve
```
