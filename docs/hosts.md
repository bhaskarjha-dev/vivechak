# Supported AI Editors & Agent Harness Presets

Vivechak auto-detects and configures 21 AI coding editors, agent harnesses, and runtime environments in 1 second using `vck setup`.

## 1. Quick Setup Commands

| Host / IDE | Shortcut | Target Configuration File | Server Key | 1-Second Setup Command |
|---|---|---|---|---|
| **Cursor** | `cursor` | `.cursor/mcp.json` or `~/.cursor/mcp.json` | `mcpServers` | `vck setup cursor` |
| **VS Code** | `vscode` | `.vscode/mcp.json` (Copilot Agent mode) | `servers` | `vck setup vscode` |
| **Google Antigravity** | `agy` | `~/.gemini/config/mcp_config.json` or `.agents/mcp_config.json` | `mcpServers` | `vck setup agy` |
| **Claude Code CLI** | `claude-code` | `.mcp.json` or `~/.claude.json` | `mcpServers` | `vck setup claude-code` |
| **Claude Desktop** | `claude` | OS-specific `claude_desktop_config.json` | `mcpServers` | `vck setup claude` |
| **OpenCode** | `opencode` | `opencode.json` or `~/.config/opencode/opencode.json` | `mcp` | `vck setup opencode` |
| **Zed** | `zed` | `~/.config/zed/settings.json` (or AppData on Windows) | `context_servers` | `vck setup zed` |
| **AWS Kiro** | `kiro` | `.kiro/settings/mcp.json` or `~/.kiro/settings/mcp.json` | `mcpServers` | `vck setup kiro` |
| **Windsurf** | `windsurf` | `~/.codeium/windsurf/mcp_config.json` | `mcpServers` | `vck setup windsurf` |
| **Block Goose** | `goose` | `~/.config/goose/config.yaml` (YAML format) | `extensions` | `vck setup goose` |
| **GitHub Copilot** | `copilot` | `~/.copilot/mcp-config.json` | `mcpServers` | `vck setup copilot` |
| **ByteDance Trae** | `trae` | `.trae/mcp.json` or `~/.trae/mcp.json` | `mcpServers` | `vck setup trae` |
| **Oh My Pi (OMP)** | `omp` | `.omp/mcp.json` or `~/.omp/agent/mcp.json` | `mcpServers` | `vck setup omp` |
| **Factory Droid** | `droid` | `.factory/mcp.json` or `~/.factory/mcp.json` | `mcpServers` | `vck setup droid` |
| **Cline** | `cline` | VS Code globalStorage `cline_mcp_settings.json` | `mcpServers` | `vck setup cline` |
| **Roo Code** | `roo` | VS Code globalStorage `cline_mcp_settings.json` | `mcpServers` | `vck setup roo` |
| **Cognition Devin** | `devin` | `.devin/mcp_config.local.json` | `mcpServers` | `vck setup devin` |
| **OpenHands** | `openhands` | `~/.openhands/mcp.json` | `mcpServers` | `vck setup openhands` |
| **Void Editor** | `void` | `.void/mcp.json` or `~/.void-editor/mcp.json` | `mcpServers` | `vck setup void` |
| **Continue.dev** | `continue` | `.continue/mcpServers/vivechak.json` | `mcpServers` | `vck setup continue` |
| **Universal Portable** | `mcp` | `.mcp.json` in workspace root | `mcpServers` | `vck setup mcp` |

---

## 2. Universal Auto-Detection

Run without arguments in any project workspace:
```sh
vck setup
```
Vivechak inspects directory markers (`.cursor`, `.agents`, `.vscode`, `.trae`, etc.) and configures the active environment automatically.
