# Host Setup Guide

Configure Vivechak as an MCP server in your AI host.

> **Recommended:** Use the `mcp-config` command — it auto-detects the binary path and writes the correct config file:
>
> ```sh
> vivechak mcp-config --client <host> --write
> ```
>
> Manual configuration is shown below for reference.

---

## Cursor

### Automatic
```sh
vivechak mcp-config --client cursor --write
```

### Manual

Edit `~/.cursor/mcp.json`:

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

## VS Code

### Automatic
```sh
vivechak mcp-config --client vscode --write
```

### Manual

Edit `.vscode/mcp.json` in your project root:

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
| All | `.vscode/mcp.json` (project-level) |

---

## Claude Desktop

### Automatic
```sh
vivechak mcp-config --client claude-desktop --write
```

### Manual

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

> **⚠️ macOS Note:** Claude Desktop runs as a GUI app with a truncated PATH (`/usr/bin:/bin:/usr/sbin:/sbin`). The `command` field **must** use an absolute path to the vivechak binary. The `mcp-config` command handles this automatically.

---

## Antigravity (Google Gemini)

### Automatic
```sh
vivechak mcp-config --client antigravity --write
```

### Manual

Edit `.gemini/settings.json` in your project root:

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
| All | `.gemini/settings.json` (project-level) |

---

## ChatGPT (Codex)

### Automatic
```sh
vivechak mcp-config --client chatgpt
```

> **Note:** `--write` is not supported for ChatGPT — the config location varies. Copy the output JSON into your ChatGPT/Codex MCP configuration.

### Manual

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

## OpenAI Codex

### Automatic
```sh
vivechak mcp-config --client codex
```

> **Note:** `--write` is not supported for Codex — copy the output JSON into your Codex MCP configuration.

### Manual

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

## Kiro

### Automatic
```sh
vivechak mcp-config --client kiro
```

> **Note:** `--write` is not supported for Kiro — copy the output JSON into your Kiro MCP configuration.

### Manual

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

## Windows Users

On Windows, replace the command path with the full path to `vivechak.exe`. For example:

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

The `mcp-config --write` command always uses the absolute path to the currently-running binary, so it works correctly on all platforms.

---

## Troubleshooting

### Binary not found
Make sure the `command` field uses an **absolute path**. GUI applications on macOS truncate PATH to `/usr/bin:/bin:/usr/sbin:/sbin`.

### Verify installation
```sh
vivechak version
vivechak doctor /path/to/your/project
```

### Test MCP connection
Use the MCP Inspector to verify the server responds:
```sh
npx @modelcontextprotocol/inspector vivechak serve
```
