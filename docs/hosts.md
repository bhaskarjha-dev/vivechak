# Supported AI Editors & Agent Harness Presets

> Generated from `internal/hosts/presets.go` (`go run ./tools/docgen`). Do not edit by hand.

`vck setup <preset>` configures 10 hosts whose MCP config file and key are verified against each host's own documentation (URLs in `cmd/vivechak/presets.go` and `internal/hosts/presets.go`). Any other host: `vck setup <path-to-its-mcp-config>`.

## 1. Quick Setup Commands

| Host / IDE | Shortcut | Target Configuration File | Server Key | Setup Command | Host Docs |
|---|---|---|---|---|---|
| **Cursor** | `cursor` | `.cursor/mcp.json` or `~/.cursor/mcp.json` | `mcpServers` | `vck setup cursor` | [Docs](https://cursor.com/docs/context/mcp) |
| **VS Code** | `vscode` | `.vscode/mcp.json` (legacy file VS Code documents as deprecated; removal planned for February 2027) | `servers` | `vck setup vscode` | [Docs](https://code.visualstudio.com/docs/copilot/reference/mcp-configuration) |
| **Claude Desktop** | `claude` | OS-specific `claude_desktop_config.json` | `mcpServers` | `vck setup claude` | [Docs](https://modelcontextprotocol.io/docs/develop/connect-local-servers) |
| **Google Antigravity** | `agy` | `~/.gemini/config/mcp_config.json` or `.agents/mcp_config.json` | `mcpServers` | `vck setup agy` | [Docs](https://antigravity.google/docs/mcp) |
| **Windsurf** | `windsurf` | `~/.codeium/windsurf/mcp_config.json` (read only when `chat.mcp.discovery.enabled` includes the windsurf source) | `mcpServers` | `vck setup windsurf` | [Docs](https://docs.devin.ai/windsurf/plugins/cascade/mcp) |
| **Zed** | `zed` | `~/.config/zed/settings.json` (or AppData on Windows) | `context_servers` | `vck setup zed` | [Docs](https://zed.dev/docs/ai/mcp) |
| **Roo Code** | `roo` | `.roo/mcp.json` (project) | `mcpServers` | `vck setup roo` | [Docs](https://docs.roocode.com/features/mcp/using-mcp-in-roo) |
| **Claude Code CLI** | `claude-code` | `.mcp.json` or `~/.claude.json` | `mcpServers` | `vck setup claude-code` | [Docs](https://code.claude.com/docs/en/mcp) |
| **OpenCode** | `opencode` | `opencode.json`, `.opencode/opencode.json` or `~/.config/opencode/opencode.json` | `mcp` | `vck setup opencode` | [Docs](https://opencode.ai/docs/config/) |
| **Block Goose** | `goose` | `~/.config/goose/config.yaml` (YAML format) | `extensions` | `vck setup goose` | [Docs](https://goose-docs.ai/docs/guides/config-files/) |

Hosts without a preset (for example Cline, Kiro, Continue, Trae, GitHub Copilot CLI): `vck setup <path-to-its-mcp-config>`.

---

## 2. Universal Auto-Detection

Run without arguments in any project workspace:
```sh
vck setup
```
Vivechak inspects directory markers (`.cursor`, `.agents`, `.vscode`, `.mcp.json`, `opencode.json`, `.opencode`, `.goose`, `.roo`) and configures the detected host.
