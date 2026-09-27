# Vivechak MCP Server: Python vs Go — First-Principles Evaluation

> **Date:** 2026-09-23
> **Decision Type:** Two-Way Door (language choice for MCP server)
> **Verdict:** Go is the stronger choice for Vivechak's MCP server.

---

## The Server's Job (What We're Actually Building)

Before comparing languages, be precise about what the MCP server does:

- Creates directories and copies template files
- Reads markdown files (GENERATOR.md, templates/) and injects text
- Parses YAML frontmatter from markdown files
- Validates JSON/YAML against schemas (ADR 18-field schema, pipeline structure)
- Manages file-based state (`.vivechak/state.json`)
- Tracks DAG dependencies (which sessions are done, what's next)
- Communicates via stdio (JSON-RPC 2.0)

**What it does NOT do:** Call LLM APIs, do web searches, run ML models, process natural language. The host agent does all of that.

This is a **file I/O + validation + state management** tool. Not an AI tool.

---

## Head-to-Head Comparison

| Criterion | Go | Python | Winner | Weight |
|---|---|---|---|---|
| **SDK maturity** | Official Tier 1 SDK (modelcontextprotocol/go-sdk). ~5K stars. Full protocol support: Tools, Resources, Prompts, stdio, SSE, streamable HTTP. | Reference implementation. ~24K stars. Full support + built-in `mcp dev` inspector tool. | Python (more mature, inspector tool) | Medium |
| **Distribution** | Single static binary. No runtime dependency. `go install` or download binary. Zero friction. | Requires Python 3.10+, pip/uv, virtual env recommended. Dependency conflicts common. | **Go** (decisive) | **High** |
| **Cross-platform** | `GOOS=windows/linux/darwin GOARCH=amd64/arm64 go build`. One command, 6 platforms. | Works on all platforms IF Python is installed and configured correctly. | **Go** (decisive) | **High** |
| **Startup time (stdio)** | 5–20ms cold start. Instant for stdio transport. | 50–150ms+ (interpreter + imports). Noticeable for frequent tool calls. | **Go** | Medium |
| **File I/O + validation** | Excellent. Strong typing via structs. YAML/JSON parsing via standard library + `gopkg.in/yaml.v3`. Struct tags auto-generate JSON schemas. | Excellent. Pydantic v2 for validation. `PyYAML` for YAML. More concise code for schema definition. | Tie (Go more type-safe, Python more concise) | Medium |
| **Development speed** | Slower initial development. More boilerplate. But strong compiler catches errors early. | Faster prototyping. Less code. But runtime errors more common. | Python (marginally) | Low |
| **Agent Plugins 1.0.0** | Fully supported. Language-agnostic spec. Go binary in plugin directory works. | Fully supported. | Tie | Low |
| **Future Engine path** | If Engine is ever built, Go is natural for long-running services, DAG execution, concurrency (goroutines). No migration needed. | Would need rewrite or Python-Go bridge if Engine moves to Go later. | **Go** | Medium |
| **Community/ecosystem** | Go MCP ecosystem is smaller but growing. Many developer tools in Go (Docker, K8s, Terraform, gh CLI). | Larger MCP community. More example servers. | Python (larger community) | Low |
| **User's dev machine** | Binary just works. No "do you have Python 3.11?" question. | Every developer has a different Python setup. Virtualenv wars. | **Go** (decisive) | **High** |
| **LLM library access** | Not needed (Approach B — server doesn't call LLMs). If ever needed, Go LLM libraries exist but are less mature. | Not needed. But if ever needed, Python is dominant. | Tie (irrelevant for Approach B) | N/A |

---

## The Decisive Factor: Distribution

The MCP server's primary adoption barrier is **installation friction.** Consider:

**Go installation experience:**
```
# Option 1: go install
go install github.com/vivechak/vivechak@latest

# Option 2: download binary
# Download from releases page, put in PATH

# Option 3: in Claude Desktop config
{
  "mcpServers": {
    "vivechak": {
      "command": "/usr/local/bin/vivechak",
      "args": ["serve"]
    }
  }
}
```

**Python installation experience:**
```
# Hope user has Python 3.11+
python --version  # 3.9? 3.10? Which python? python3?

# Install
pip install vivechak  # or pip3? or uv? 
# "ERROR: externally-managed-environment" on Ubuntu 23+
# Create venv?

# In Claude Desktop config
{
  "mcpServers": {
    "vivechak": {
      "command": "python",       # or python3? or /usr/bin/python3.11?
      "args": ["-m", "vivechak"]
    }
  }
}
```

For a developer tool targeting the "post-vibe" developer who just wants to run `vivechak` — a Go binary eliminates an entire class of installation problems.

---

## Honest Pushback Against Go

**1. "Python is the AI ecosystem language"**
True for ML/LLM work. Irrelevant for Vivechak's MCP server, which is a file I/O and validation tool (Approach B). If we were building Approach A (server calls LLMs), Python would win. We're not.

**2. "Python SDK has more community examples"**
True. More MCP servers are written in Python today. But Go servers exist in production, the SDK is Tier 1, and Vivechak's tools are straightforward enough that SDK maturity differences don't matter.

**3. "Go is slower to develop"**
Marginally true for initial development. But the compile-time type checking catches errors that Python would catch at runtime (or not at all). For a tool that validates YAML schemas and tracks state, Go's type system is an asset.

**4. "The ROADMAP-NEXT says Python"**
Yes, and this analysis updates that recommendation with new evidence (Go SDK is Tier 1, official, feature-complete). The decision was classified as a Two-Way Door — this is exactly when you update.

---

## Verdict

**Go is the stronger choice for Vivechak's MCP server.** The decisive factors are:

1. **Distribution friction is the #1 adoption barrier** — Go eliminates it with a single binary.
2. **The server does file I/O + validation, not AI** — Python's LLM ecosystem advantage is irrelevant.
3. **Cross-platform without runtime dependency** — one `go build` command, 6 platforms.
4. **Future Engine alignment** — if the Engine is ever built, Go is the natural choice.
5. **Go SDK is Tier 1, official, feature-complete** — no SDK maturity penalty.

### Recommendation: Update ROADMAP-NEXT

Change the technology stack from Python to Go:

| Component | Old Choice | New Choice | Rationale |
|---|---|---|---|
| Language | Python 3.11+ | Go 1.22+ | Single binary distribution, cross-platform, no runtime dependency |
| MCP SDK | `mcp` (Python) | `github.com/modelcontextprotocol/go-sdk` | Official Tier 1, full protocol support |
| Validation | Pydantic v2 | Go struct tags + `gopkg.in/yaml.v3` | Compile-time type safety, auto JSON schema from struct tags |
| Distribution | PyPI (`pip install`) | GitHub Releases + `go install` | Zero-dependency binary |
| Template engine | Jinja2 | `text/template` (stdlib) | No external dependency needed |
