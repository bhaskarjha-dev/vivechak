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
| **JSON Configuration File** | **Cline**, **OpenCode**, **OMP (Oh My Pi)**, **Continue**, **Roo Code** | Run `vivechak mcp-config` and copy the JSON snippet into your agent's config file (e.g. `cline_mcp_settings.json`, `opencode.json`, `~/.omp/agent/mcp.json`). |
| **CLI Registration Command** | **Claude Code**, **Block Goose** | Register via the agent's CLI:<br>• Claude Code: `claude mcp add vivechak -- /path/to/vivechak serve`<br>• Goose: `goose configure` (add extension → `stdio` → command: `vivechak`, args: `serve`) |
| **Interactive Settings Panel** | **Factory Droid**, **Desktop AI apps**, **Agent Dashboards** | In the agent's MCP settings panel, add a server with Command: `/path/to/vivechak` (run `vivechak mcp-config` to copy it) and Arguments: `serve`. |

Once connected, simply prompt your agent:
> *"Use Vivechak to research the architecture for our project. We need to evaluate our primary backend and datastore trade-offs."*

---

## 2. Auto-Writing Configuration

Vivechak can automatically merge its server definition directly into existing JSON configuration files without overwriting other server settings.

### Universal File Writing (Any Agent Harness)

To configure any agent harness that reads an MCP configuration file (e.g. **Cline**, **OpenCode**, **OMP**, **Continue**, **Roo Code**, or custom pipelines), pass the path to its file:

```sh
vivechak mcp-config --path /path/to/mcp-settings.json --write
```

Vivechak automatically inspects the target file and applies the correct schema key:
- If the target is in `.vscode/` or already uses `"servers"`, it merges under `"servers"`.
- If the target is a Zed settings file or uses `"context_servers"`, it merges under `"context_servers"`.
- Otherwise, it merges under the standard `"mcpServers"` key.

### Optional Desktop Path Shortcuts (Convenience)

For common desktop IDEs, Vivechak provides built-in path shortcuts (`--preset <shortcut>`) so you don't have to manually look up and type their default configuration paths:

```sh
vivechak mcp-config --preset <shortcut> --write
# (--client <shortcut> is also supported as a backward-compatible alias)
```

| Desktop IDE | Preset Shortcut | Default Config Location | Schema Key |
|---|---|---|---|
| **Cursor** | `cursor` | `~/.cursor/mcp.json` | `mcpServers` |
| **VS Code** | `vscode` | `.vscode/mcp.json` (workspace) | `servers` |
| **Claude Desktop** | `claude-desktop` | `claude_desktop_config.json` | `mcpServers` |
| **Windsurf** | `windsurf` | `~/.codeium/windsurf/mcp_config.json` | `mcpServers` |
| **Antigravity** | `antigravity` | `.gemini/settings.json` (workspace) | `mcpServers` |
| **Zed** | `zed` | `~/.config/zed/settings.json` | `context_servers` |
| **AWS Kiro** | `kiro` | `~/.aws/.kiro/mcp.json` | `mcpServers` |

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

### Google Antigravity / Gemini CLI

#### Automatic
```sh
vivechak mcp-config --preset antigravity --write
```

#### Manual
Edit `.gemini/settings.json` (or `.agents/mcp_config.json`) in your project root:

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

AWS Kiro supports MCP servers natively.

#### Automatic
```sh
vivechak mcp-config --preset kiro --write
```

#### Manual
Edit `~/.aws/.kiro/mcp.json`:

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
