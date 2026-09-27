---
id: R-05
title: Agent Plugins 1.0.0 Specification & Distribution Strategy for Vivechak's MCP Server
date: 2026-09-23
status: draft
topic: distribution
informs_decisions: [D-004]
---

# Agent Plugins 1.0.0 Specification & Distribution Strategy for Vivechak's MCP Server

**Evidence grading used throughout this document:**
**[A]** primary source (official spec text, vendor docs/support pages, direct repository inspection) · **[B]** strong secondary (multiple independent corroborating sources, or a detailed community source that itself cites primary docs) · **[C]** single secondary/community source, directionally useful but not independently cross-checked · **[D]** this document's own inference/synthesis, not a direct claim from any source.

---

## Research Question

Given the Agent Plugins 1.0.0 specification (agent-plugins.org) and the current MCP-configuration surfaces of seven agent hosts — Claude Desktop, Cursor, VS Code, Antigravity, ChatGPT, GitHub Copilot, and Kiro — what distribution strategy for Vivechak's Go-binary MCP server minimizes installation friction while maximizing OS coverage (Windows/macOS/Linux) and host coverage, and where does Agent Plugins 1.0.0 fall short of solving that problem on its own?

---

## Key Findings

### 1. Scope reality-check: what "Vivechak's MCP server" actually is today

Before the spec work, the brief's premise needs to be checked against the actual repository, because it does not hold up cleanly.

`github.com/bhaskarjha-dev/vivechak` **[A]** is, as of this research date, a markdown-only "evidence-grounded pre-development research meta-framework" — a generator prompt, four ADR-style templates, and a sealed meta-research corpus. There is no Go code, no `cmd/`, no `go.mod`, and no MCP server anywhere in the current tree. The repository's own `ROADMAP.md` **[A]** places an MCP server ("Vivechak Engine") in **Phase 5, explicitly labeled "Research Phase,"** and lists two items as open, unresolved **One-Way-Door research questions**, not settled facts:

> *"Language choice — Python vs Go vs Rust vs TypeScript... Not all frameworks support all languages."*
> *"Deployment model — CLI tool vs IDE plugin vs web service vs hybrid... Affects who can use it and how."*

In other words, the repository's own tracked roadmap says "Go" and "MCP server" have not yet been decided through Vivechak's own methodology — the same methodology this research session is an instance of. That is not necessarily a contradiction: doing exactly this kind of distribution research *before* the language/deployment decision is locked is consistent with the framework's stated purpose ("before writing a single line of implementation begins"). But it does mean:

- This document treats "Go binary, MCP server" as a **working premise supplied by the brief**, not a verified decision, and produces **contingent** research: *if* Go and *if* MCP-server-as-deployment-model, *then* this is the distribution strategy.
- D-004 should either point back to (or spawn) a decision record that actually locks the language/deployment-model choice — otherwise D-004 risks becoming the kind of "premature architectural decision made on unverified assumptions" the framework itself was built to prevent. See **Open Questions**, item 1.

### 2. Agent Plugins 1.0.0: what the spec is, and its central limitation for this brief

Agent Plugins is real and current **[A]**. Timeline, corroborated across the spec site and secondary coverage **[A][B]**: the effort began in April 2026 as Vercel's internal "Open Plugin Specification," was renamed and put under a multi-vendor Technical Steering Committee (TSC) in mid-July 2026, and **v1.0.0 published July 24, 2026** — roughly two months before this research. TSC composition is reported consistently as **Amazon, Cursor, Microsoft, OpenAI, and Vercel** **[A][B]**; one secondary source additionally lists Google **[C]**, which this document could not independently verify against the primary governance document and flags as an open discrepancy rather than resolving it either way.

What the spec actually standardizes, read from the canonical text at `agent-plugins.org/specification` **[A]**:

- A plugin is a directory with a required `plugin.json` manifest at its root.
- Exactly **two** portable component types exist in v1: **Agent Skills** (discovered non-recursively from `skills/<name>/SKILL.md`, per the separate Agent Skills spec) and **MCP servers** (declared in a root `mcp.json`).
- Everything else a client might want — slash commands, hooks, subagents, LSP servers, permissions — is explicitly **not standardized** and lives in client-owned, reverse-domain namespaces (e.g., `com.github.copilot/`) that other clients are required to ignore without validating.
- The spec is deliberately a **"small interoperability floor."**

That last point is the crux of this research. The spec's own design-philosophy section says it plainly **[A]**:

> *"Distribution, installation, permissions, user experience, and client-specific capabilities remain entirely under each client's control... It is an interoperability floor, not a package manager."*

**This is the limitation the brief asked this session to surface, and it is real and material.** Agent Plugins defines what a plugin *looks like* once it is sitting in a directory on a machine. It says nothing about how a Go binary gets onto that machine, how a client finds out the plugin exists in the first place, how updates are pushed, or how the binary is signed or verified. Adopting Agent Plugins does not reduce the distribution work for Vivechak by one channel — it adds a packaging *format* on top of channels that still have to exist.

There is a second, more specific limitation for a **Go binary** specifically. The `mcp.json` `stdio` schema has exactly one field for the executable — `command` — and it must be either a bare executable name or a plugin-relative path beginning with `./`; there is no `os`/`arch`-conditional field anywhere in the schema **[A]**. The spec's own design-decisions section resolves the ambiguity by recommending: *"A plugin that bundles an executable in the package MUST use a plugin-relative `command`"* **[A]** — and separately states that a conformant plugin **cannot rely on PATH resolution**: *"Whether a configured PATH environment value participates in resolving a bare command is client-defined. Plugins claiming conformance MUST NOT depend on that behavior"* **[A]**.

Put together: a single Go binary artifact is platform- and architecture-specific, but a single `mcp.json` has no way to say "use this binary on Windows, that one on macOS ARM." A genuinely conformant Agent Plugin bundling its own binary therefore cannot be one universal package — it has to be **N platform-specific plugin bundles**, each with its own binary pre-selected and pre-named at the fixed relative path. This is a real, spec-level consequence of Go's compile-once-per-platform model meeting a package format built primarily around interpreted-language servers (`npx`, `uvx`) where a single `command: "npx"` works everywhere. Section 3 of the Distribution Strategy below designs around this directly.

### 3. Where implementations already diverge from the spec text

Because the spec is two months old, "supports Agent Plugins" claims should be treated as aspirational until checked against each vendor's own docs — and checking them surfaced concrete gaps:

- **Cursor** documents that it does *not* implement the placeholder expansion the spec requires. Cursor's own FAQ **[A]**: *"Cursor does not expand the standard's `${PLUGIN_ROOT}` and `${PLUGIN_DATA}` variables in `mcp.json`; use `${CURSOR_PLUGIN_ROOT}` for the plugin root."* Any `args`/`env`/`cwd` value in a strictly spec-conformant `mcp.json` that uses `${PLUGIN_DATA}` will pass through to Cursor as a literal, unexpanded string, not a path. `command` itself is unaffected (it resolves via direct relative-path handling, not placeholder substitution), so a bundled binary still launches — but anything else parameterized through those placeholders needs a Cursor-specific fallback.
- **Antigravity** (not an Agent-Plugins client at all — see next finding) has a related but separate problem at the native-config layer: its subprocess environment does not inherit the local shell `PATH`. A Google AI developer-forum thread hosted on `discuss.ai.google.dev` **[B]** shows a user hitting exactly this with a bare `npx` command, and the guidance given is to switch to an **absolute path** to the executable. This applies directly to a bare `vivechak-mcp` command in Antigravity's native config.
- **VS Code, GitHub Copilot, and Claude Code** each layer a same-shaped-but-different, non-Agent-Plugins manifest alongside (or instead of) the standard: VS Code's own docs table five coexisting plugin formats — Agent Plugins 1.0, "Copilot," "Claude" (`.claude-plugin/plugin.json`, using `${CLAUDE_PLUGIN_ROOT}`), and "Legacy OpenPlugin" (`.plugin/plugin.json`) **[A]**. GitHub Copilot's own docs separately describe a "legacy" MCP path via `.mcp.json`, `.github/mcp.json`, or an `mcpServers` manifest field **[A]** — three more variants of "where does MCP config live" inside one product family. None of this breaks anything for a plugin author who sticks to the pure Agent Plugins fields, but it means "Agent Plugins support" is best read as "one of several formats this client also happens to read," not as a guarantee that every path in the spec behaves identically everywhere.

### 4. Per-host reality across the required seven hosts

The single highest-leverage fact this research surfaced: **`agent-plugins.org/compatible-clients`** **[A]**, the spec's own authoritative client list, does **not** include Claude Desktop or Antigravity. Cross-referenced against the brief's seven required hosts:

| Host | Listed as Agent-Plugins-conformant? | Evidence |
|---|---|---|
| Cursor | Yes | **[A]** compatible-clients page + Cursor's own docs |
| VS Code | Yes | **[A]** compatible-clients page + VS Code's own docs |
| GitHub Copilot | Yes | **[A]** compatible-clients page + GitHub's own docs |
| ChatGPT & Codex | Yes | **[A]** compatible-clients page + OpenAI's own docs |
| Kiro | Yes (branded "Powers") | **[A]** compatible-clients page + kiro.dev docs |
| **Claude Desktop** | **No** | **[A]** absent from the list; uses its own `.mcpb` "Desktop Extensions" format instead |
| **Antigravity** | **No** | **[A]** absent from the list; uses its own `mcp_config.json` format instead |

**Anthropic is not participating in Agent Plugins on any surface** **[A][D]**: no TSC seat, Claude Desktop absent from the compatible-clients list, and Claude Code (a different Anthropic product, the CLI) uses its own proprietary `.claude-plugin/plugin.json` format that VS Code merely happens to also read. Given Claude Desktop is one of the seven hosts this brief explicitly requires, **Agent Plugins packaging cannot be the whole distribution strategy under any circumstances** — native configuration is mandatory for at least two of the seven hosts, full stop, independent of how the spec matures.

A second finding worth flagging because it changes a widely-held assumption: **Claude Desktop reached Linux in beta on June 30, 2026** — Ubuntu 22.04+ and Debian 12+, x86_64 and arm64, installed and updated through Anthropic's own apt repository, with MCP support included from day one **[B]**, corroborated across multiple independent outlets that cite the same official announcement and support docs. Before this, Linux users depended on an unofficial community repackaging project (`aaddrick/claude-desktop-debian`) **[B][C]**. This means Claude Desktop is now nominally available on all three required operating systems — but the Linux build is beta, Debian-family-only, and does not self-update the way the Mac/Windows builds do (`apt upgrade`, not in-app auto-update) **[B]**. Treat this as real but conditional coverage, not full parity yet.

A third piece worth naming explicitly because "ChatGPT" is not one distribution target: the consumer ChatGPT web app can only reach MCP-backed tools through OpenAI's reviewed "universal plugin directory" — a submission flow built around OAuth, optional UI, and commerce/checkout review, aimed at public-facing apps **[A]**. The low-friction path relevant to a devtools binary is the separate **"Codex host"** — the ChatGPT desktop app, Codex CLI, and Codex IDE extension — which **share one local config file, `~/.codex/config.toml`** (TOML, not JSON), editable by hand or via a GUI "Add server" flow with no review step **[A]**. Treating "ChatGPT" as a single review-gated target would massively overstate the friction; treating it as "config.toml, like everything else" would understate it, because the format is TOML and the key syntax (`[mcp_servers.name]`) differs from every JSON-based host.

### 5. A distribution/discovery layer the brief's scope didn't name: the official MCP Registry

Separate from both Agent Plugins packaging and each host's own marketplace, there is a third, protocol-level layer: the **official MCP Registry** (`registry.modelcontextprotocol.io`), originally launched in preview by Anthropic in September 2025 and **handed to the Linux Foundation's Agentic AI Foundation for governance in December 2025** **[B]**. It indexes MCP *servers* specifically (not full plugins) via a `server.json` manifest, with namespace ownership proved through GitHub OAuth or DNS **[B]**. By September 10, 2026 it held **30,375 unique servers**, having roughly tripled since May 2026, with publication heavily concentrated (the top 50 publishers account for a quarter of all entries) and a documented pattern of **62% of servers published once and never updated** **[B]**. Downstream community directories (Smithery, Glama, PulseMCP) and, per one analysis, an increasing number of agent hosts themselves read from this registry rather than requiring a separate listing per client **[B]**.

This matters for the brief's question of "how agent hosts discover... plugins": publishing a `server.json` to the official registry is a low-cost, complementary **discovery** step, independent of which binary-delivery channel is used and independent of whether Agent Plugins packaging is pursued. It does not replace anything below — a registry entry still has to point at an actual installable artifact — but omitting it means Vivechak is invisible to the aggregator layer a meaningful share of the ecosystem now searches first.

### 6. The Go-binary distribution toolchain is mature and largely automatable from one CI job

None of the above changes the fact that the *foundation* of any of this — getting cross-compiled, checksummed, signed binaries out at all — is a solved problem. **`goreleaser`**, driven by a single `.goreleaser.yaml` and a tagged Git release, can in one CI run produce: cross-compiled binaries across the practical GOOS/GOARCH matrix, `tar.gz`/`zip` archives, a `checksums.txt` signed with `cosign` via GitHub's OIDC-based keyless signing, a Homebrew formula pushed to a tap repository, a Scoop manifest pushed to a bucket repository, and (via the bundled `nfpm` packager) `.deb`, `.rpm`, and `.apk` packages **[A][B]**. A CLI-generation vendor's own documentation independently confirms this exact "full stack" (goreleaser config, GitHub Actions workflow, `install.sh`/`install.ps1`, checksums + detached signature, optional Homebrew, optional WinGet, optional nfpm) as standard industry practice for 2026 **[B]**, which is good corroboration that this isn't a niche approach.

One concrete, dated correction to older tutorials: **Go 1.24 dropped the `windows/arm` (32-bit ARM) build target, and Go 1.25 removed it entirely** **[B]**, per a February 2026 shared-CI-config pull request that documents the resulting `goreleaser` failure and fix. Any cross-compilation matrix written from a pre-2026 tutorial should be checked against this.

On Windows package managers specifically — the brief asked about Scoop and Chocolatey, but the research surfaced a channel that deserves equal billing: **`winget`**, Microsoft's own package manager, now ships **pre-installed on Windows 11** and is reachable via the Microsoft Store's "App Installer" on most Windows 10 machines, so it is the only Windows channel with **zero bootstrap step** for the end user **[B][C]**. Scoop remains the better fit for Vivechak's specific persona (installs to the user's home directory, no admin/UAC, purpose-built for exactly this class of small CLI tool, and — unlike `winget`'s community-moderated central `winget-pkgs` repo — a maintainer-owned Scoop **bucket** needs no external review to publish or update) **[B][C]**. Chocolatey typically requires administrator rights and is oriented toward broader/enterprise software catalogs **[B]**; for a single Go CLI tool aimed at individual developers, it is the least leveraged of the three. This is reported here per the brief's explicit instruction to flag better channels prominently.

---

## Recommendation

Ship in three tiers. The guiding principle: **the format is the same binary and the same six release artifacts everywhere; only the wrapper around them differs per host, and native config-file editing is the one channel that works on all seven hosts on day one with zero packaging investment.**

### Tier 1 — ship on day one (foundation + zero-investment reach)

1. **GitHub Releases via `goreleaser`**, producing signed, checksummed binaries for the full GOOS/GOARCH matrix (below). This is the substrate every other channel points at.
2. **`install.sh` / `install.ps1`** — one curl-or-`iwr`-and-run script per shell family, auto-detecting OS/arch, downloading the right release asset, verifying the checksum, and placing the binary on `PATH`. Lowest friction of any channel that doesn't require a pre-existing package manager.
3. **Native MCP config snippets for all seven hosts**, published in the README/docs as copy-paste blocks (exact syntax per host in the Distribution Strategy section below). This is the only channel that reaches Claude Desktop and Antigravity at all, and it works the day the first binary exists — no marketplace review, no packaging format decision required.
4. **Homebrew tap** (`goreleaser`'s `brews:` block). One tap formula covers macOS *and* Linux, because Homebrew runs on Linux too — this quietly closes most of the Linux gap without touching `apt`/`dnf` at all.
5. **Official MCP Registry listing** (`server.json`, GitHub-OAuth-verified namespace). Near-zero cost, and it's the aggregator layer several hosts and most third-party directories now read from.

### Tier 2 — add once Tier 1 is live and the binary is stable

6. **Agent Plugins 1.0.0 packages**, built as **platform-specific bundles** (see Finding 2 and the Distribution Strategy) for the five conformant hosts — Cursor, VS Code, GitHub Copilot, Kiro ("Powers"), and, if pursued, the ChatGPT/Codex ecosystem. Kiro is the standout here: its Powers marketplace is genuinely Agent-Plugins-native and one-click, making it the single lowest-friction *marketplace* install of the seven hosts once this work is done.
7. **Scoop bucket** (`goreleaser`'s `scoops:` block) — self-hosted, no moderation queue, matches the audience.
8. **`winget` manifest submission** to `microsoft/winget-pkgs` — highest-reach, zero-bootstrap Windows channel; accept the community PR-review lag as a one-time cost per version bump (or automate it — tooling like `wingetcreate` exists for this).
9. **Claude Desktop `.mcpb` Desktop Extension** — package with `manifest.json` and `mcpb pack`, distribute by direct sideload (Settings → Extensions → Advanced → Extension Developer → Install Extension) immediately; pursue the official Anthropic-reviewed directory listing in parallel as a slower-moving track.

### Tier 3 — lower priority / demand-gated

10. **`nfpm`-built `.deb`/`.rpm`/`.apk`** behind a self-hosted apt/yum repo (e.g., Cloudsmith, Fury.io, or a self-managed `reprepro`/GitHub Pages repo) — worthwhile once there's a large enough Debian/Fedora user base that wants `apt install`/`dnf install` update semantics; plain downloadable `.deb`/`.rpm` files on the GitHub Release are a fine zero-effort stopgap in the meantime.
11. **Chocolatey** — publish only if enterprise/Windows-admin demand actually shows up; the moderation queue and admin-rights requirement make it the weakest ROI of the Windows channels for this audience.
12. **`go install github.com/.../cmd/vivechak-mcp@latest`** — keep it working (it's free, since it just needs a `main` package at a stable import path), document it for Go developers specifically, but don't treat it as a primary channel: it requires a local Go toolchain, compiles from source on every install, and has no update mechanism beyond the user manually re-running the command.

### What not to do

- Don't treat Agent Plugins 1.0.0 as sufficient by itself. It cannot reach Claude Desktop or Antigravity, and even where it's "supported," the schema alone doesn't solve Go's platform-specific-binary problem (Finding 2) or guarantee identical runtime behavior (Finding 3).
- Don't wait for the spec to mature before shipping. It is two months old and adoption claims in press coverage are ahead of what individual client docs currently commit to **[B][D]** — Tier 1 gets Vivechak in front of users on all seven hosts today, independent of how Agent Plugins evolves.
- Don't rely on a bare command name (`vivechak-mcp`) in Antigravity's native config, or in any Agent-Plugins `mcp.json` where strict conformance matters — use absolute or plugin-relative paths per Finding 3.

---

## Distribution Strategy

### Vivechak `plugin.json`

Minimal, spec-valid manifest (required fields are only `$schema` and `name`; everything else below is optional metadata worth filling in) **[A]**. `vivechak` satisfies the name constraints (1–64 chars, lowercase alphanumeric/hyphen/period, starts/ends alphanumeric, no `--`/`..`) **[A]**. `MIT` matches the license already declared in the actual repository **[A]**.

```json
{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json",
  "name": "vivechak",
  "version": "1.0.0",
  "description": "Evidence-grounded pre-development research pipeline exposed as MCP tools: generate research sessions, grade evidence, and track architecture decisions.",
  "author": {
    "name": "Bhaskar Jha",
    "url": "https://github.com/bhaskarjha-dev"
  },
  "homepage": "https://github.com/bhaskarjha-dev/vivechak",
  "repository": "https://github.com/bhaskarjha-dev/vivechak",
  "license": "MIT",
  "keywords": ["research", "architecture-decisions", "evidence-grading", "adr", "pre-development"]
}
```

Notes:
- The schema is **closed** — any top-level field not in `{$schema, name, version, description, author, homepage, repository, license, keywords, extensions}` is reported and ignored (non-fatal), so there's no benefit to inventing extra top-level fields; anything Vivechak-specific belongs under `extensions` **[A]**.
- `keywords` does double duty in Kiro specifically: Kiro's "Powers" runtime uses conversation keywords to decide when to dynamically load a power's tools rather than loading everything up front **[A]** — so these values are worth choosing deliberately (terms a user would actually type, like `"research"`, `"architecture decision"`, `"evidence"`), not just SEO-style tags.

### Vivechak `mcp.json` — and why one file can't cover three operating systems

```json
{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json",
  "mcpServers": {
    "vivechak": {
      "type": "stdio",
      "command": "./bin/vivechak-mcp",
      "args": ["serve"],
      "env": {
        "VIVECHAK_DATA_DIR": "${PLUGIN_DATA}"
      }
    }
  }
}
```

This is spec-valid: `command` is a plugin-relative path beginning with `./` (not a placeholder — the spec explicitly forbids placeholder expansion *inside* `command` itself, and a `./`-relative path is resolved directly against the plugin root, which is a different, permitted mechanism) **[A]**; `${PLUGIN_DATA}` in `env` *is* eligible for expansion and gives the binary a client-managed, update-persistent directory to write state into, per the spec's own recommended use (`"installed dependencies... generated code, caches, and other plugin state that should persist across updates"`) **[A]**.

**The catch, restated from Finding 2:** `./bin/vivechak-mcp` has to be one specific file, and a Go binary compiled for `darwin/arm64` will not run on `windows/amd64`. Because the schema has no OS/arch conditional, this exact file is a **template** — the actual thing shipped to each of the five Agent-Plugins-conformant hosts must be **platform-specific bundle variants**, produced at release time:

```
vivechak-plugin-darwin-arm64.zip   → bin/vivechak-mcp   (darwin/arm64 binary)
vivechak-plugin-darwin-amd64.zip   → bin/vivechak-mcp   (darwin/amd64 binary)
vivechak-plugin-linux-amd64.zip    → bin/vivechak-mcp   (linux/amd64 binary)
vivechak-plugin-linux-arm64.zip    → bin/vivechak-mcp   (linux/arm64 binary)
vivechak-plugin-windows-amd64.zip  → bin/vivechak-mcp.exe  (windows/amd64 binary, command adjusted to "./bin/vivechak-mcp.exe")
```
Each variant carries an identical `plugin.json` and `skills/`, differing only in `bin/` and the `.exe` suffix in `mcp.json`'s `command` on Windows. This is the direct, practical answer to "if Agent Plugins 1.0.0 has limitations for Go binary distribution" — the limitation is real, and the workaround is release-time templating, not a spec feature **[D]**.

For **Cursor** specifically, ship a second `mcp.json` variant (or a build-time substitution) using `${CURSOR_PLUGIN_ROOT}` in any field beyond `command` that needs the plugin root, per Finding 3 — or simply avoid needing `${PLUGIN_ROOT}`/`${PLUGIN_DATA}` in anything except `command` (which doesn't need expansion at all) to sidestep the gap entirely. The latter is the simpler fix and is reflected in the example above, which only uses `${PLUGIN_DATA}` — worth testing on Cursor specifically before relying on it there too, since Cursor's docs only confirm `${CURSOR_PLUGIN_ROOT}`, not a Cursor-specific data-directory equivalent.

### `skills/` directory

Per spec **[A]**: non-recursive, immediate child directories of `skills/` containing a file literally named `SKILL.md` are each treated as one skill; the file format itself is governed by the separate Agent Skills specification, not Agent Plugins. A plausible pairing for Vivechak (illustrative only — server/skill implementation is out of scope for this session):

```
skills/
└── research-pipeline/
    ├── SKILL.md          # walks the agent through generating/executing a Vivechak session
    └── references/
        └── decision-template.md
```

### Full directory layout (one platform variant)

```
vivechak/                        # plugin root
├── plugin.json                  # REQUIRED — manifest (Distribution Strategy §1)
├── mcp.json                     # MCP server declaration (Distribution Strategy §2)
├── skills/
│   └── research-pipeline/
│       └── SKILL.md
├── bin/
│   └── vivechak-mcp[.exe]       # the platform-specific compiled binary
├── LICENSE
└── CHANGELOG.md
```

### Distribution channel matrix

**Channel × OS coverage** (what gets the binary onto disk):

| Channel | Windows | macOS | Linux | Maintainer effort | End-user friction |
|---|---|---|---|---|---|
| GitHub Releases (`goreleaser`) | ✅ | ✅ | ✅ | Low — one CI job | Medium — manual download + `PATH` |
| `install.sh` / `install.ps1` | ✅ (ps1) | ✅ (sh) | ✅ (sh) | Low | Low — one command |
| Homebrew tap | — | ✅ | ✅ (Linuxbrew) | Low (`goreleaser brews:`) | Very low |
| Scoop bucket | ✅ | — | — | Low (`goreleaser scoops:`) | Very low |
| `winget` | ✅ | — | — | Medium (community PR review) | Very low (pre-installed on Win11) |
| Chocolatey | ✅ | — | — | Medium (moderation queue, admin needed) | Low |
| `nfpm` → `.deb`/`.rpm`/`.apk` | — | — | ✅ | Medium (repo hosting for real `apt`/`dnf` UX) | Low–medium |
| `go install ...@latest` | ✅ | ✅ | ✅ | ~Zero | Low if Go toolchain present, else blocked |

**Config format × agent host** (what wires the installed binary into each host — this is the part that actually varies by AI agent, independent of which row above put the binary on disk):

| Host | Native config path/format | Agent Plugins support | Notes |
|---|---|---|---|
| Claude Desktop | `claude_desktop_config.json` **or** `.mcpb` Desktop Extension | **No** | Only native path reaches this host |
| Cursor | `.cursor/mcp.json` / `~/.cursor/mcp.json` | Yes, with `${PLUGIN_ROOT}` gap | Deeplink install also available |
| VS Code | `.vscode/mcp.json` (key: `servers`, not `mcpServers`) | Yes, full | Also reads Copilot/Claude/legacy formats |
| Antigravity | `~/.gemini/config/mcp_config.json` (or older `~/.gemini/antigravity/...`) | **No** | Needs absolute path — no `PATH` inheritance |
| ChatGPT / Codex | `~/.codex/config.toml` (TOML) — shared by ChatGPT desktop app, Codex CLI, Codex IDE ext | Yes (via reviewed plugin directory) | `config.toml` is the low-friction path; the plugin directory is the reviewed/heavy path |
| GitHub Copilot | Varies by surface: VS Code as above; CLI via `copilot plugin install` or `~/.copilot/settings.json`; cloud agent via `.github/copilot/settings.json` | Yes, full | Multiple install surfaces under one brand |
| Kiro | `~/.kiro/settings/mcp.json` or `<project>/.kiro/settings/mcp.json`, or `kiro-cli mcp add` | Yes, branded "Powers," one-click | Best marketplace UX of the seven once packaged |

### Installation UX per agent host

Each block assumes the binary `vivechak-mcp` has already been installed by one of the Tier 1/2 channels above and is either on `PATH` or at a known absolute path.

**Claude Desktop**
1. One-click (once a `.mcpb` exists): **Settings → Extensions → Browse extensions** (official directory) or **Settings → Extensions → Advanced settings → Extension Developer → Install Extension…**, then select the `.mcpb` file.
2. Manual fallback: edit `claude_desktop_config.json` — macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`; Windows: `%APPDATA%\Claude\claude_desktop_config.json`; Linux (beta): follows the same XDG-style convention as other Linux apps, verify against the installed build.
   ```json
   { "mcpServers": { "vivechak": { "command": "/absolute/path/to/vivechak-mcp" } } }
   ```
3. Fully quit and restart Claude Desktop; confirm via the "+" → **Connectors** panel in the chat box.

**Cursor**
1. Marketplace (once an Agent Plugin exists): **Customize → search "vivechak" → Install**, choose project or user scope.
2. One-click via deeplink: `cursor://anysphere.cursor-deeplink/mcp/install?name=vivechak&config=<base64-encoded-json>`.
3. Manual fallback: add to `.cursor/mcp.json` (project) or `~/.cursor/mcp.json` (global):
   ```json
   { "mcpServers": { "vivechak": { "command": "vivechak-mcp" } } }
   ```

**VS Code**
1. Marketplace (once an Agent Plugin exists): ensure `chat.plugins.enabled` is `true`, then install via the Extensions view filtered to `@agentPlugins`, or via **Chat: Install Plugin From Source** from the Command Palette.
2. Manual fallback: add to `.vscode/mcp.json` — note the top-level key is **`servers`**, not `mcpServers`, and `type` is explicit:
   ```json
   { "servers": { "vivechak": { "type": "stdio", "command": "vivechak-mcp" } } }
   ```

**Antigravity**
1. No Agent Plugins path currently — native config only.
2. Edit `~/.gemini/config/mcp_config.json` (check for the older `~/.gemini/antigravity/mcp_config.json` path if using an older build) or the project-scoped `.agents/mcp_config.json`:
   ```json
   { "mcpServers": { "vivechak": { "command": "/absolute/path/to/vivechak-mcp" } } }
   ```
   **Use an absolute path** — Antigravity's subprocess does not inherit shell `PATH`.
3. Refresh via **Settings → Customizations → Installed MCP Servers → Refresh** (Antigravity 2.0/IDE) or `agy` → `/mcp` (CLI).

**ChatGPT / Codex**
1. GUI (ChatGPT desktop app): **Settings → MCP servers → Add server** → name `vivechak`, transport **STDIO**, command `vivechak-mcp` → Save → Restart.
2. Config file (shared by ChatGPT desktop app, Codex CLI, Codex IDE extension): edit `~/.codex/config.toml` (or project-scoped `.codex/config.toml` for trusted projects):
   ```toml
   [mcp_servers.vivechak]
   command = "vivechak-mcp"
   ```
3. Consumer ChatGPT web plugin directory is a separate, reviewed submission path — deprioritized per the Recommendation.

**GitHub Copilot**
1. In VS Code: same as the VS Code block above.
2. Copilot CLI: `copilot plugin install vivechak@<marketplace>` or `/plugin install` in-chat, or declaratively via `enabledPlugins` in `~/.copilot/settings.json` (user) or `.github/copilot/settings.json` (repo).
3. Copilot cloud agent: declarative only, via `.github/copilot/settings.json`'s `enabledPlugins` (add `extraKnownMarketplaces` first if not using a default marketplace).
4. GitHub Copilot app: **Customize → Plugins** → browse → Install.

**Kiro**
1. One-click (once packaged as a "Power"): browse `kiro.dev/powers` or in-IDE, click **Install** — no JSON editing.
2. Install from a GitHub URL directly, without waiting for marketplace curation.
3. Manual fallback: edit `~/.kiro/settings/mcp.json` (user) or `<project>/.kiro/settings/mcp.json` (project):
   ```json
   { "mcpServers": { "vivechak": { "command": "vivechak-mcp", "disabled": false } } }
   ```
   or via CLI: `kiro-cli mcp add --name vivechak --scope global --command vivechak-mcp`.

### Cross-compilation build matrix

Standard, current (post Go 1.24) target set for a CLI/MCP-server binary **[A][B]**:

| GOOS | GOARCH | Ship? | Notes |
|---|---|---|---|
| `darwin` | `amd64` | ✅ | Intel Macs |
| `darwin` | `arm64` | ✅ | Apple Silicon |
| `linux` | `amd64` | ✅ | |
| `linux` | `arm64` | ✅ | incl. Raspberry Pi 4+, AWS Graviton, Apple Silicon under Linux |
| `windows` | `amd64` | ✅ | |
| `windows` | `arm64` | ✅ | Windows on ARM (Surface, Snapdragon laptops) |
| `windows` | `arm` (32-bit) | ❌ | dropped by Go 1.24, removed by Go 1.25 — do not build |

Minimal `.goreleaser.yaml` build block reflecting this, with static-binary and reproducibility flags confirmed against real-world CI examples **[B]**:

```yaml
builds:
  - id: vivechak-mcp
    main: ./cmd/vivechak-mcp
    binary: vivechak-mcp
    env:
      - CGO_ENABLED=0
    flags:
      - -trimpath
    ldflags:
      - -s -w -X main.version={{.Version}}
    goos: [darwin, linux, windows]
    goarch: [amd64, arm64]

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

nfpms:
  - id: vivechak-mcp
    package_name: vivechak-mcp
    formats: [deb, rpm, apk]
    goos: [linux]

brews:
  - repository:
      owner: bhaskarjha-dev
      name: homebrew-vivechak

scoops:
  - repository:
      owner: bhaskarjha-dev
      name: scoop-bucket
```

`CGO_ENABLED=0` guarantees a genuinely static binary on every target (important for the "single static binary" requirement in the brief, and a prerequisite for the `./bin/vivechak-mcp` bundling approach in the Agent Plugins packages, since a dynamically-linked binary could fail on a machine missing the right shared libraries). `-trimpath` and the version-stamped `ldflags` support reproducible, verifiable builds.

### Update & versioning strategy

**Single source of truth:** an annotated Git tag (`vX.Y.Z`, SemVer) drives the entire release. `goreleaser` reads it once and stamps every downstream artifact — the Homebrew formula, the Scoop manifest, the `winget` manifest, the `nfpm` packages, and (via `ldflags -X main.version=...`) the binary's own `--version` output — from that single value, so there is exactly one place a maintainer bumps a version **[A][D]**. `plugin.json`'s `version` field and the official MCP Registry's `server.json` version should be bumped in the same release step, ideally the same CI job, rather than as an afterthought — the registry data shows 62% of published servers are never updated again **[B]**, which is the specific failure mode to design against.

Update mechanics differ sharply by channel, and this determines how much "please update" documentation is needed per host:

| Channel | Update mechanism | User action required |
|---|---|---|
| Homebrew / Scoop / `winget` / Chocolatey | `brew upgrade` / `scoop update` / `winget upgrade` / `choco upgrade` | Manual, but a familiar habit for the audience |
| `nfpm` via a real hosted apt/yum repo | Rides normal `apt upgrade` / `dnf upgrade` | Manual, but bundled into routine OS maintenance |
| Plain GitHub Release binary / `install.sh` re-run | None automatic | Fully manual re-download |
| `go install ...@latest` | None automatic | Fully manual re-run |
| Claude Desktop `.mcpb`, official directory listing | Automatic | None |
| Claude Desktop `.mcpb`, sideloaded/private | None automatic | **[A]** Anthropic's own docs are explicit: *"For privately distributed extensions, users will need to install updated `.mcpb` files manually."* |
| VS Code Agent Plugin, git-repo marketplace source | Automatic check every 24h (or manual Command Palette check) if `extensions.autoUpdate` is on | Low — confirm, don't reinstall |
| VS Code Agent Plugin, npm/PyPI source | Manual "Update" button appears | Manual click (not applicable here — ship via a git-based marketplace to get the automatic path) |
| Cursor Marketplace | Marketplace re-indexes on push (Auto Refresh, ≤10 min batching) if enabled by the marketplace owner | Not fully confirmed whether an already-installed user's copy auto-updates versus only the marketplace listing refreshing — flagged in Open Questions |
| Official MCP Registry | Versioned, but does not push updates to consumers itself | Maintainer must re-publish `server.json` each release; consumers re-discover on their own cadence |

Practical recommendation **[D]**: lead with channels that self-update (Homebrew/Scoop/`winget`, and a real apt/yum repo if that channel is pursued) in all "how to install" documentation, since they eliminate the single biggest real-world failure mode (users stuck on an old version indefinitely). Reserve the plain-binary-download and `go install` instructions for users who have a specific reason to want them.

---

## Open Questions & Risks

1. **The Go/MCP-server premise is not a locked decision.** Vivechak's own `ROADMAP.md` frames language choice and deployment model as unresolved One-Way-Door research questions for a component that doesn't exist yet. Recommend confirming — or explicitly creating — a decision record for that choice, and marking D-004 as conditional on it, before treating this distribution strategy as final.
2. **Agent Plugins 1.0.0 is two months old.** Schema details, TSC composition, and per-client conformance could all still shift before any 1.1/2.0 revision. Treat AP packaging as additive polish (Tier 2), not load-bearing infrastructure, until it has a longer track record.
3. **Spec-vs-implementation gaps found here (Cursor's `${PLUGIN_ROOT}`, Antigravity's PATH behavior) were discovered by reading vendor docs, not by testing an actual package on each host.** Before shipping, validate the platform-specific bundles against each of the five Agent-Plugins-conformant hosts directly rather than trusting spec-conformance claims at face value.
4. **Claude Desktop's Linux build is beta and Debian-family-only.** Fedora/Arch/other-distro users on Claude Desktop still need the `go install`, manual-binary, or (if pursued) `nfpm` path — there is no first-party route for them yet.
5. **Official-directory review timelines are unknown and likely variable**: Claude Desktop's extension directory ("we're building a directory... complete our interest form"), `winget-pkgs`' community PR review, VS Code's default marketplaces, and Cursor's manual plugin review all gate on human review with no published SLA found in this research. Sideload/self-hosted paths should be the actual launch mechanism; official-directory listings should be tracked as a slower parallel workstream, not a blocker.
6. **Name collision risk was not checked.** "vivechak" was not verified for availability/collision across the Homebrew tap namespace, the Scoop bucket ecosystem, `winget-pkgs`, the official MCP Registry, npm (if a wrapper is ever added), or existing Agent Plugin names. Do this before locking any manifest.
7. **Antigravity's config path appears to have moved between product versions** (`~/.gemini/antigravity/mcp_config.json` on older builds vs. `~/.gemini/config/mcp_config.json` on "Antigravity 2.0," per community documentation, not an official migration note) — confirm the current path against the actually-installed Antigravity version at documentation time rather than hardcoding one.
8. **Cursor's user-side auto-update behavior for an installed Agent Plugin was not fully confirmed** — Cursor's docs describe marketplace re-indexing (Auto Refresh) clearly but were less explicit about whether an already-installed plugin on a user's machine updates automatically or requires a manual re-install/re-open. Worth a direct test before writing user-facing update guidance.
9. **The TSC-composition discrepancy** (whether Google holds a seat, per one secondary source, or not, per the majority of sources and the primary spec's own governance-doc reference which was not independently pulled) was not resolved and is noted rather than papered over.
10. **This document did not evaluate server-implementation questions** (what MCP tools Vivechak should expose, transport choice beyond stdio for a purely local tool, etc.) — explicitly out of scope per the brief, but worth flagging that the illustrative `skills/research-pipeline` example above is schematic, not a design commitment.

---

## Sources & Evidence Ledger

| # | Source | URL | Grade | Used for |
|---|---|---|---|---|
| 1 | GitHub — `bhaskarjha-dev/vivechak` (root) | github.com/bhaskarjha-dev/vivechak | A | Repo reality-check: markdown-only framework, no Go/MCP code |
| 2 | GitHub — `vivechak/ROADMAP.md` | github.com/bhaskarjha-dev/vivechak/blob/main/ROADMAP.md | A | "Vivechak Engine" Phase 5 research-phase status; open language/deployment questions |
| 3 | Agent Plugins Specification 1.0.0 (canonical) | agent-plugins.org/specification | A | Complete field reference: plugin.json, mcp.json, discovery, conformance, versioning |
| 4 | Agent Plugins — Compatible Clients | agent-plugins.org/compatible-clients | A | Authoritative host support list; Claude Desktop/Antigravity absence; per-client transport support |
| 5 | VS Code Docs — Agent plugins in VS Code | code.visualstudio.com/docs/agent-customization/agent-plugins | A | VS Code AP implementation, plugin-format comparison table, `chat.pluginLocations`, update behavior |
| 6 | Cursor Docs — Plugins | cursor.com/docs/plugins | A | Cursor AP support + `${CURSOR_PLUGIN_ROOT}` gap, Marketplace, deeplinks, local testing |
| 7 | GitHub Docs — About GitHub Copilot plugins | docs.github.com/en/copilot/concepts/agents/about-plugins | A | Copilot's two plugin formats, marketplace.json, install surfaces (CLI/cloud agent/app) |
| 8 | OpenAI Developers — Plugins | developers.openai.com/plugins | A | ChatGPT/Codex plugin architecture, reviewed "universal plugin directory" |
| 9 | ChatGPT Learn — Model Context Protocol (Codex) | developers.openai.com/codex/extend/mcp (learn.chatgpt.com/docs/extend/mcp) | A | `~/.codex/config.toml` schema, GUI flow, plugin-provided MCP servers |
| 10 | Kiro Docs — Powers | kiro.dev/docs/powers/ | A | Kiro's Agent-Plugins-native "Powers," dynamic loading, one-click install |
| 11 | Claude Help Center — Getting Started with Local MCP Servers on Claude Desktop | support.claude.com/en/articles/10949351 | A | `.mcpb` Desktop Extensions, manifest.json, languages supported (incl. binary), update behavior, sensitive-field encryption |
| 12 | pypi.org — `kubernetes-readonly-mcp` package docs | pypi.org/project/kubernetes-readonly-mcp/0.2.0/ | B | Cross-host native config comparison: Claude Code, Codex CLI, Kiro CLI, Antigravity |
| 13 | docs.payzu.com.br — MCP setup guide | docs.payzu.com.br/en/docs/pix-processamento/mcp | C | Cross-host config comparison incl. Antigravity path version migration, VS Code `servers` key |
| 14 | Google AI Developer Forum — "Mcp_config.json won't accept 'type' field" | discuss.ai.google.dev/t/mcp-config-json-wont-accept-type-field/120831 | B | Antigravity PATH-inheritance gotcha, official guidance to use absolute paths |
| 15 | Cua Docs — Antigravity CLI | cua.ai/docs/use-cua-with/antigravity-cli | B | Antigravity stdio config confirmation, project-scope `.agents/mcp_config.json` |
| 16 | SonarSource Docs — Antigravity quickstart | docs.sonarsource.com/sonarqube-mcp-server/setup/quickstart-guides/antigravity | A | Confirms Antigravity as "successor to Gemini CLI," stdio-transport guidance |
| 17 | pasqualepillitteri.it / itbrief.co.uk / itbrief.com.au / omgubuntu.co.uk / letsdatascience.com / technobezz.com (Claude Desktop Linux beta coverage, multiple outlets) | see individual URLs | B | Claude Desktop Linux beta (June 30, 2026): scope, distros, apt-based updates, missing features |
| 18 | GitHub Issue — `anthropics/claude-code#65697` ("Preflight Checklist") | github.com/anthropics/claude-code/issues/65697 | C | Historical context on the pre-beta Linux gap; used cautiously as it may predate the fix |
| 19 | dsebastien.net — "Agent Plugins: one package format..." | dsebastien.net/agent-plugins-one-package-format-for-skills-and-mcp-servers | B | AP overview, TSC members, adoption claims |
| 20 | ravichaganti.com — "Agent Plugins: Package reusable..." | ravichaganti.com/blog/agent-plugins-package-reusable-agent-components... | B | AP overview; sole source listing Google as a TSC member (unresolved discrepancy) |
| 21 | dreampixelforge.com — "agent plugins spec" | dreampixelforge.com/blog/agent-plugins-spec | B | Governance/timeline detail, AWS naming-collision warning, "ahead of the evidence" caveat |
| 22 | eesel.ai — "Agent Plugins: the new open standard" | eesel.ai/blog/agent-plugins | B | Component-support table (standardized vs. client-specific), minimal-manifest example |
| 23 | codex.danielvaughan.com — Agent Plugins 1.0 for Codex CLI | codex.danielvaughan.com/2026/08/08/... | B | Plugin-name constraints, early mcp.json examples (superseded by primary source) |
| 24 | GoReleaser Docs — Builds customization | goreleaser.com/customization/builds | A | GOOS/GOARCH/GOARM matrix generation, ignore rules |
| 25 | GitHub PR — cloudposse/.github#246 | github.com/cloudposse/.github/pull/246 | B | Go 1.24+ dropping `windows/arm`; dated Feb 2026 |
| 26 | GitHub PR — cloudoperators/cloudctl#46 | github.com/cloudoperators/cloudctl/pull/46/files | B | Real-world GH Actions matrix (6 targets), `CGO_ENABLED=0`, `-trimpath`, `-ldflags "-s -w"` |
| 27 | nfpm Docs — Install / Quick Start | nfpm.goreleaser.com/install, /docs/quick-start | A | nfpm install methods, `.deb`/`.rpm`/`.apk`/Arch/`.msix` packagers, `cosign` artifact verification |
| 28 | Speakeasy Docs — Distribute a CLI | speakeasy.com/docs/cli-generation/distribute-cli | B | Confirms "full stack" (goreleaser + install scripts + Homebrew + WinGet + nfpm) as standard practice |
| 29 | ScoopInstaller Wiki — Chocolatey and Winget Comparison | github.com/ScoopInstaller/Scoop/wiki/So-What/Chocolatey-and-Winget-Comparison | B | Scoop's dev-focused, no-admin design philosophy |
| 30 | dev.to — Chocolatey vs Scoop package managers | dev.to/bowmanjd/chocolatey-vs-scoop-package-managers-for-windows-2kik | B | Admin/UAC distinctions between Windows package managers |
| 31 | opper.ai — AI Roundtable: scoop vs winget vs choco | opper.ai/ai-roundtable/questions/scoop-vs-winget-vs-choco-8ffcccad | C | Directional signal that `winget` is the 2026 general-recommendation consensus; not primary evidence |
| 32 | dev.to — "The MCP registry by the numbers" | dev.to/amareswer/the-mcp-registry-by-the-numbers-38nc | B | Registry scale (30,375 servers, Sept 2026), growth curve, publisher concentration, update-neglect rate |
| 33 | qveris.ai — MCP Registry Guide 2026 | qveris.ai/guides/mcp-registry | B | Registry governance history (Anthropic → Linux Foundation, Dec 2025), `server.json` schema, namespace verification |
| 34 | RoxyAPI — MCP Registry Guide 2026 | roxyapi.com/blogs/mcp-registries-where-to-list-your-server | B | Registry as 2026 consolidation point; downstream directory propagation |

**Discrepancies logged rather than resolved:** TSC membership (Google included per source #20 only); Cursor's user-side plugin auto-update behavior (not explicit in source #6); Antigravity's exact current config path (sources #12–13 disagree by version).
