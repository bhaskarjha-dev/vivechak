---
id: R-05
title: "Agent Plugins 1.0.0 and the distribution strategy for Vivechak's Go MCP server"
date: 2026-09-23
status: draft
topic: distribution
informs_decisions: [D-004]
---

# R-05 — Agent Plugins 1.0.0 and the distribution strategy for Vivechak's Go MCP server

> **Bottom line.** Agent Plugins 1.0.0 is a *packaging* format. On its own it cannot deliver a cross-platform Go binary: `mcp.json` has no OS/arch selection, no placeholder expansion in `command`, and no install hook, and the spec explicitly leaves installation, distribution and updates to each client. Use it for what it does well (one portable wrapper carrying `skills/` plus an MCP entry, loadable by Cursor, VS Code, GitHub Copilot, Kiro, Codex and, apparently, Antigravity), and ship the **binary** through OS-native channels plus a Claude Desktop `.mcpb`. Two of the seven named hosts need something the spec does not provide: **Claude Desktop** (no native Agent Plugins support; its one-click path is MCPB) and **ChatGPT web/directory** (accepts remote MCP only). Recommended shape: **binary-first, host-wired, plugin-thin** (see Recommendation).

## Research Question

**Primary question.** How should Vivechak's MCP server, a single statically linked Go binary, be distributed so that installation friction is minimal, Windows/macOS/Linux are all covered, and all seven target hosts (Claude Desktop, Cursor, VS Code, Antigravity, ChatGPT, GitHub Copilot, Kiro) can run it?

**Sub-questions** (mirroring the brief's scope)

| # | Question |
|---|---|
| Q1 | What does the Agent Plugins 1.0.0 spec require and allow (layout, `plugin.json`, `mcp.json`, `skills/`, discovery)? |
| Q2 | Can an Agent Plugin carry a Go binary to every OS/arch? If not, where exactly does it stop? |
| Q3 | How does each host discover, install, update and govern plugins and local MCP servers? |
| Q4 | Which channels (GitHub Releases, `go install`, Homebrew, Scoop/winget/Chocolatey, apt/dnf, MCPB, npm wrapper, OCI, MCP Registry) fit which OS × host? |
| Q5 | What build matrix, signing, and versioning/update strategy follows? |

**Assumptions** (not stated in the brief; correct them if wrong)

- **A1.** The binary is built with `CGO_ENABLED=0`, speaks MCP over stdio, and can gain small distribution-driven subcommands (`--version`, optionally `setup`/`mcp-config`). Server internals are out of scope.
- **A2.** Org/repo names, publisher ID, product description, license and the stdio subcommand are unknown; they appear as `<TBD>`. The `mcp` subcommand in examples is a placeholder.
- **A3.** Vivechak needs local execution (user's files/network). If it does not, a hosted mode changes the answer (Finding F7, Open Question Q-1).
- **A4.** "Now" is 2026-09-23. Spec 1.0.0 was published 2026-08-06; 1.1.0 is a working draft.

**Method and limits.** Read in full: the normative spec and the authoring/implementer pages, plus the docs for VS Code, Cursor, Claude Code (plugins + marketplaces), MCPB, Claude Connectors, OpenAI/Codex plugins and MCP. Read as excerpts/vendor pages: Kiro, Antigravity, GitHub Copilot, MCP Registry, GoReleaser, Homebrew, Windows signing, Go release notes. **Nothing was executed on any host or OS.** Anything undocumented is graded D and converted into a test (see Open Questions, T1–T15).

**Evidence grades** (inline as `[grade·source]`, sources numbered S1… in the ledger)

| Grade | Meaning |
|---|---|
| **A** | Primary: normative spec text or vendor documentation page, read in full on 2026-09-23 |
| **B** | Vendor/secondary-official: vendor blog, changelog or codelab, or vendor docs seen only as a search excerpt |
| **C** | Community/third-party: issues, READMEs, practitioner posts. Evidence of what people hit, not of what is guaranteed |
| **D** | Inference or background knowledge, not re-verified this session. A hypothesis until tested |

## Key Findings

**Why timing matters.** The standard is seven weeks old; expect churn.

| Date | Event | Source |
|---|---|---|
| 2026-08-06 | Agent Plugins 1.0.0 published. TSC Core Maintainers: Amazon, Cursor, Microsoft, OpenAI, Vercel; Google joins as Core Maintainer the same day | [B·S7] |
| 2026-08-07 | Codex CLI 0.147.0 reported to add install of portable Agent Plugins | [C·S29] |
| 2026-08-12 | GitHub: GA in VS Code, Copilot CLI, Copilot SDK and Copilot app, all plans | [B·S10] |
| mid-Aug | 1.1.0 working draft appears (proposes a fixed `assets/icon.png`); no published 1.1.0 schemas as of 2026-09-02 | [C·S52; B·S53] |
| Aug 2026 | Go 1.27 released; requires macOS 13+ | [B·S46] |
| 2026-09-01 | Homebrew's official cask tap disables casks that fail Gatekeeper | [B·S44] |
| 2026-09-16 | VS Code's Agent Plugins docs last updated | [A·S11] |

### F1. Agent Plugins is a package format; the pieces a Go binary needs are the pieces it leaves out. `[A·S1, S5, S6; B·S7]`

- `command` must be "a single executable token": a bare name or a `./` path inside the plugin root, and clients "MUST NOT perform placeholder expansion in `command`". So: no `${GOOS}`/`${GOARCH}`, no per-OS override, and no way to point `command` at `${PLUGIN_DATA}`.
- A plugin is a directory, deliberately not an archive or registry bundle. "Client-managed installation, distribution, enablement, updates, and user interface are outside the portable specification." There is no install/post-install hook and no dependency resolution; `dependencies`, signing/provenance and secrets appear only as non-committal "future considerations".
- Consequence: a portable `mcp.json` can launch a Go server in only three ways: (i) a **bare command on PATH** (thin plugin), (ii) a **`./` path to a bundled file**, which forces one same-named dispatcher for all OS/arch (fat plugin), or (iii) an **indirect runner** (`npx`, `uvx`, `docker`).
- Google's own guidance: shipping one MCP server to one client is still simpler as a bare `mcp.json`; a plugin earns its keep when skills travel with the server. `[B·S7]`

### F2. Host support is uneven; two of seven hosts do not read the portable format natively.

| Host | Reads Agent Plugins 1.0 (root `plugin.json` + `mcp.json`)? | Runs local stdio MCP? | Evidence |
|---|---|---|---|
| VS Code | Yes. Auto-detected by `$schema`; marketplaces, `@agentPlugins` view | Yes | [A·S11] |
| GitHub Copilot (CLI, app, SDK) | Yes. GA 2026-08-12; Awesome Copilot marketplace on by default | Yes | [B·S10] |
| Cursor | Yes. Root `plugin.json` loads; standard's placeholders not expanded | Yes | [A·S13] |
| Kiro | Yes. `plugin.json` or legacy `POWER.md`; import by GitHub URL | Yes (IDE, CLI; Web runs it in a cloud sandbox) | [B·S36, S38, S39] |
| ChatGPT and Codex | Listed as compatible; **public plugin directory accepts remote MCP only** | Codex CLI/IDE and ChatGPT desktop app: yes. ChatGPT web/directory: no | [A·S2, S26, S28] |
| Antigravity | Not on the official compatible list, but Google's codelab installs a `plugin.json` + `skills/` + `mcp.json` plugin via `agy plugin install <git-url>` | Yes | [B·S30] |
| Claude Desktop | **No evidence of native support.** Anthropic is not a maintainer; Claude Code documents only `.claude-plugin/plugin.json` + `.mcp.json`. Native one-click path is `.mcpb` | Yes (MCPB, JSON config) | [A·S14, S17; C·S29] |

### F3. Even supporting hosts deviate from the spec in ways that touch a launcher. `[A·S11, S13, S14]`

- **Cursor** loads spec-conformant plugins but does not expand `${PLUGIN_ROOT}`/`${PLUGIN_DATA}` in `mcp.json` (it uses `${CURSOR_PLUGIN_ROOT}`).
- **Claude Code** substitutes `${CLAUDE_PLUGIN_ROOT}` in stdio `command` (the spec forbids expansion there) and reads `.mcp.json`, not `mcp.json`. Its manifest must live at `.claude-plugin/plugin.json`.
- **VS Code** documents the placeholders as preserved "for the plugin runtime"; it does not say it expands them itself. MCP servers from plugins are implicitly trusted at install (no separate trust prompt).
- **Design rule for Vivechak:** no placeholders in `mcp.json`. The server reads `PLUGIN_DATA`/`PLUGIN_ROOT` from its environment (spec §9.1 requires clients to set them for stdio subprocesses) and falls back to OS user directories when absent. `[D]`

### F4. A bare `vivechak` command is fragile on macOS GUI hosts. `[C·S47]`

Multiple independent sources report that GUI-launched hosts (Claude Desktop, Cursor, VS Code, Windsurf, Zed) do not see the shell's PATH, producing `spawn … ENOENT`; the standard workaround is an absolute path such as `/opt/homebrew/bin/…`. A Claude Code issue shows the desktop app's tool shell with only `/usr/bin:/bin:/usr/sbin:/sbin`. The spec itself says conformant plugins "MUST NOT depend on" PATH participation for bare commands. The behaviour varies by host and launch method, so it is a test (T3), but a thin plugin needs a mitigation: a `setup`/`mcp-config` subcommand that writes the absolute path, and per-host absolute-path snippets in the docs.

### F5. Claude Desktop's one-click path is MCPB, which Anthropic calls "the secondary distribution path", and it has sharp edges for Go binaries.

- `platform_overrides` keys are OS-only (`darwin`, `win32`, `linux`); there is no CPU-arch dimension. macOS therefore needs a `lipo` universal binary; Windows/Linux arm64 vs amd64 need separate bundles or emulation. `[A·S16; C·S20, S21]`
- Real Go servers show the failure modes: grafana's directory-published bundle pointed every Mac at an arm64 binary and broke Intel Macs `[C·S22]`; one bundle grew 44 MB → 74 MB when Linux was added `[C·S20]`; Windows-built zips lost exec bits (fixed in `mcpb pack`) `[C·S24]`.
- Signing with an **external enterprise HSM signer** (GaraSign, ESRP, SignServer, Venafi, Azure SignTool) currently produces bundles Claude Desktop's `adm-zip`-based loader rejects with "Invalid comment length": the signer appends its PKCS#7 signature after the ZIP end-of-central-directory record without updating `comment_length`. An upstream fix (`mcpb prepare-for-signing` / `apply-signature`) is open, not yet merged, as of 2026-09-23. **The built-in `mcpb sign --self-signed` and `mcpb sign --cert/--key` paths are a different code path and are not implicated by this bug** `[B·S23]`. Separately, an unsigned or self-signed bundle installed outside the directory shows a "developer not verified" notice at install; this is by design, not a defect, and what upgrades a publisher to "verified" was not established in this research (Open Question).
- Anthropic's docs say Claude Desktop runs on macOS and Windows `[A·S17]`; a Linux beta (Ubuntu 22.04+/Debian 12+, x86_64 and arm64) is reported by a third party citing an Anthropic article `[C·S20]`.
- MCPB apps auto-append `.exe` on Windows for binaries `[A·S16]`, so per-OS bundles are straightforward.

### F6. ChatGPT is the hard outlier. `[A·S26, S27, S28]`

- OpenAI's plugin portal accepts skills-only or **remote Streamable-HTTP MCP** submissions. For "a local MCP server or `.mcpb`" it says to deploy a public HTTPS endpoint or contact OpenAI for local MCP support; "the portal doesn't accept `.mcpb` files".
- Local stdio works in the **ChatGPT desktop app** (Settings → MCP servers → Add server → STDIO) and in **Codex** CLI/IDE via `~/.codex/config.toml`. Codex's default `startup_timeout_sec` is 10.
- ChatGPT and Codex share one universal plugin directory, so a local-only server cannot be listed there.

### F7. A remote (Streamable HTTP) mode is the only channel that covers ChatGPT web/directory, and it is Anthropic's preferred directory route. The MCP Registry has no "Go" package type. `[A·S17, S26; B·S19, S41]`

- Anthropic: "Remote MCP servers are recommended for directory listing"; local servers go through a separate desktop-extension form. Zero install on every host. Whether Vivechak can run remotely depends on A3 (Q-1).
- The MCP Registry lists npm, PyPI, NuGet, OCI, MCPB and Cargo package types; a prebuilt Go binary reaches it only as **MCPB** (GitHub/GitLab release URL + `fileSha256`), **OCI**, or behind an **npm** wrapper. The registry stores metadata, not artifacts. GoReleaser has an MCP Registry publisher. `[B·S41, S42]`

### F8. macOS distribution rules changed on 2026-09-01. `[B·S42, S44; C·S61]`

- GoReleaser deprecated `brews` (v2.10) in favour of `homebrew_casks` for prebuilt binaries. Casks are macOS-only (no Linuxbrew).
- Homebrew's *official* cask tap now disables casks failing Gatekeeper; `--no-quarantine` is deprecated. Third-party taps are still allowed (and unsupported by Homebrew's bottle/cask infrastructure); unsigned binaries in your own tap need a quarantine-strip hook. Formulae are unaffected.
- Durable path: Developer ID sign + notarize the binary. Go 1.27 needs macOS 13+; Go 1.25/1.26 support macOS 12. `[B·S46]`

### F9. Windows and Linux are comparatively straightforward. `[B·S43, S45; D]`

- winget accepts portable zips and GoReleaser generates the manifests (PR to `microsoft/winget-pkgs`); Scoop and Chocolatey publishers exist. Azure Artifact Signing is Microsoft's recommended signer for non-Store apps (organizations in the US/Canada/EU/UK; individuals only US/Canada) and does not grant instant SmartScreen trust. `[B·S43, S45]`
- Linux: nfpm-built `.deb`/`.rpm`/`.apk` from GoReleaser; GoReleaser lists Cloudsmith and GemFury publishers for hosted apt/rpm repos. A static `CGO_ENABLED=0` binary runs on glibc and musl distros. `[B·S42; D]`

### F10. Trust, provenance and enterprise controls sit outside the spec, so they must be ours. `[A·S6, S11, S13, S15; B·S10, S18; C·S51]`

- The spec defines no permission model, signing or provenance verification. VS Code trusts plugin MCP servers implicitly at install. Copilot admins can allowlist MCP servers "by URL, command, or name". Cursor Enterprise turns "Allow Local Plugin Imports" off by default. Claude Team/Enterprise admins allowlist desktop extensions and can upload custom `.mcpb`.
- Claude Desktop syncs account-enabled plugins across devices and has been reported to auto-start their MCP servers on other machines without a local prompt `[C·S51]`. The Vivechak server must therefore be safe to start unexpectedly (no side effects on launch) and the binary may be absent on those machines.

### F11. Go MCP servers in the wild have converged on the same toolkit. `[C·S20, S22, S32, S34]`

GoReleaser for the build matrix; per-OS `.mcpb` packed from GoReleaser output (grafana, gograph, gitlab-mcp-server); an extensionless `bin/<name>` POSIX dispatcher plus `<name>.exe` for Antigravity plugins (kwrkb/agy-plugins); and a `mcp-config --client <host>` subcommand that prints a snippet with the absolute path (cua-driver). These are precedents, not guarantees.

## Recommendation

**Adopt "binary-first, host-wired, plugin-thin."**

1. **Binary-first.** The signed static binary on GitHub Releases is the single source of truth. Every other channel (Homebrew, Scoop, winget, deb/rpm, MCPB, npm, registry) is generated from it by one GoReleaser pipeline.
2. **Host-wired.** Each host gets the shortest documented path that points it at that binary: a plugin where the host has a marketplace, an `.mcpb` for Claude Desktop, and a self-configuring `setup` subcommand where neither helps.
3. **Plugin-thin.** The Agent Plugin carries `skills/` and a portable `mcp.json` whose `command` is the bare `vivechak`. It contains no binaries. Binary delivery and updates belong to OS package managers, which is what F1 says the spec leaves to the ecosystem.

### Prioritised channels

| Pri | Deliverable | Why | Unlocks |
|---|---|---|---|
| **P0** | Signed static binaries on GitHub Releases via GoReleaser: darwin universal, windows/amd64, linux/amd64 + arm64; `checksums.txt`; provenance | Source of truth for every other channel | All |
| **P0** | Homebrew tap (signed + notarized cask), Scoop bucket, winget manifest, `.deb`/`.rpm` on Releases | One-command install and upgrade, no runtime dependency | macOS, Windows, Linux |
| **P0** | **Thin Agent Plugin** (portable `plugin.json` + `mcp.json` + `skills/`) with Claude-native manifests alongside, in one public repo | One repo for Cursor, VS Code, Copilot, Kiro, Antigravity, Codex (+ Claude Code) | 6 of 7 hosts |
| **P0** | **`.mcpb`** per OS: macOS universal, Windows x64 | Claude Desktop's only one-click, zero-prerequisite path | Claude Desktop |
| **P0** | Per-host copy-paste install page (§4 of this report), plus CI schema validation of every manifest | Most friction is documentation | All |
| **P1** | `vivechak setup` / `mcp-config --client <host>` (writes/prints the absolute path) | Solves the PATH problem; covers hosts with no plugin path (ChatGPT desktop, Antigravity manual) | All |
| **P1** | MCP Registry listing (`mcpb`; `npm`/`oci` optional) | Discovery in VS Code `@mcp` gallery and GitHub's registry | VS Code, Copilot |
| **P1** | Directory submissions: Cursor Marketplace, Kiro Powers catalog, Anthropic desktop-extension form | Discoverability and trust signals | Cursor, Kiro, Claude Desktop |
| **P1** | Windows arm64 build + zip; Linux `.mcpb` when the Claude Desktop Linux beta stabilises; hosted apt/dnf repo | Coverage | Windows-on-Arm, Linux |
| **P2** | npm wrapper (`npx -y @<org>/vivechak-mcp@<ver>`), OCI image, Chocolatey, documented `go install` | Fallbacks for Node/Docker/Go users | All (needs Node/Docker/Go) |
| **P2** | *Fat* plugin with dispatcher, **only if tests T1–T3 pass** | Zero-prerequisite, offline install | Cursor, VS Code, Copilot, Kiro, Antigravity, Codex |
| **Decide** | Hosted Streamable-HTTP mode | Only route to ChatGPT web/directory; Anthropic's preferred directory route; zero install | All, incl. ChatGPT web |

### What not to do (and why)

- **Do not make the fat plugin primary.** It depends on unspecified `.exe` resolution on Windows, needs a dispatcher per OS/arch, bloats every clone, and interacts badly with cache semantics (copied plugin roots change on update; MCP servers keep the old path until reload `[A·S14]`).
- **Do not rely on bare `vivechak` on macOS GUI hosts** without the `setup` subcommand or absolute-path snippets (F4).
- **Do not ship a single-architecture macOS `.mcpb`** (grafana precedent, F5). Use a `lipo` universal binary.
- **Do not commit binaries to `main` or use Git LFS for plugin sources.** Claude Code's git clone does not fetch LFS content, so files arrive as pointer files `[A·S15]`.
- **Do not add an in-binary self-updater by default.** It fights every package manager. Offer an opt-in `vivechak version --check` instead. `[D]`
- **Do not route `.mcpb` signing through an external enterprise HSM signer** (GaraSign/ESRP/SignServer/Venafi/Azure SignTool) until the upstream `prepare-for-signing`/`apply-signature` fix lands (T6); use `mcpb sign --self-signed` or `mcpb sign --cert/--key` instead, which are unaffected. `[B·S23]`
- **Do not make `curl | sh` the only path.** Acceptable as an extra for terminal users; not for enterprise. `[D]`

### Decisions requested for D-004

1. Approve the P0/P1/P2 tiering above.
2. Commit to signing prerequisites now, not at release time: Apple Developer ID + notarization, and Windows signing via Azure Artifact Signing or a purchased OV/EV certificate. Azure Artifact Signing's own documentation states identity validation normally takes **1–20 business days**; independent reports on Microsoft's support forum describe individual cases stuck for weeks with no explicit "action required" state, and one Trusted-Signing-based Windows launch was blocked on it. Budget for this in the release calendar rather than the sprint before ship. `[B·S60]`
3. Decide the hosted-mode question (Q-1).
4. Approve two small server-backlog items driven purely by distribution: build-time version stamping (`--version`, MCP `serverInfo.version`) and the `setup`/`mcp-config` subcommand.
5. Confirm that a Claude Desktop `.mcpb` is P0 even though Anthropic labels MCPB the secondary path.
6. Confirm the plugin repo can be public/open source (the Cursor Marketplace requires open-source plugins `[A·S13]`). Whether the binary itself may be closed is a separate question to ask Cursor.

## Distribution Strategy

### 1. Agent Plugins 1.0.0: complete reference

Everything in §1.1–1.9 is from the normative spec `[A·S1]` unless another source is given; §1.10 collects how hosts discover and install plugins.

#### 1.1 Status and governance

Spec Version 1.0.0, Status: Published. Governance is separate from the format (Technical Charter). Core Maintainers on the TSC: Amazon, Cursor, Microsoft, OpenAI, Vercel; Google joined on 2026-08-06 `[B·S7]`. 1.1.0 is a working draft (second schema URL `…/schemas/1.1.0/plugin.schema.json`, not yet published; proposes a fixed `assets/icon.png`) `[C·S52; B·S53]`. Requirement keywords follow RFC 2119/8174. If the spec text and a schema disagree, the text wins.

#### 1.2 Package model and exact layout

A plugin is a directory rooted at one filesystem location with a required `plugin.json` at its root. Standard layout (spec §4.2):

```
my-plugin/
├── plugin.json                  # REQUIRED
├── skills/                      # optional; immediate child dirs each containing SKILL.md
│   └── summarize/
│       ├── SKILL.md
│       ├── scripts/analyze.sh
│       └── references/checklist.md
├── mcp.json                     # optional; the only MCP config location
├── com.example.client/          # optional; client-owned extension directory
│   └── hooks/
├── LICENSE
└── CHANGELOG.md
```

Path rules: any file a client discovers, reads or executes must resolve, after following symlinks/junctions/reparse points, **inside the plugin root**; plugin-relative path fields **must begin with `./`**; other config values (args, env values) are opaque strings. Failure is scoped to the narrowest unit: manifest outside root → reject plugin; fixed component location outside root → that component type invalid; `SKILL.md` outside root → skip that skill; MCP `command`/`cwd` fails containment → skip that server; any other escaping path → deny access.

#### 1.3 `plugin.json`: complete field reference

The schema is **closed**: only these top-level fields are permitted.

| Field | Type | Req. | Constraints and behaviour |
|---|---|---|---|
| `$schema` | string | **Yes** | For 1.0.0 must equal `https://agent-plugins.org/schemas/1.0.0/plugin.schema.json`. Selects the validation contract. Clients must not fetch it; an unsupported version rejects the plugin |
| `name` | string | **Yes** | 1–64 chars; `a-z`, `0-9`, `-`, `.` only; first and last char alphanumeric; no `--`, no `..`. Periods allowed. Empty or invalid → plugin rejected |
| `version` | string | No | SemVer recommended but not enforced. "Used for update checks and cache freshness" |
| `description` | string | No | Short purpose |
| `author` | object | No | Only `name`, `email`, `url`, each a string. Any other key or value type makes the manifest invalid |
| `homepage` | string | No | Not validated as a URL |
| `repository` | string | No | Not validated as a URL |
| `license` | string | No | SPDX recommended, not enforced |
| `keywords` | string[] | No | Search/discovery tags |
| `extensions` | object | No | Keys are reverse-domain namespaces, values are objects. Non-object → reported and ignored (non-fatal) |

Failure semantics: an **unknown top-level field** is reported and ignored (non-fatal). **Any other schema violation is fatal**: the client rejects the plugin and discovers/executes none of its components. Minimal valid manifest: `$schema` + `name`.

#### 1.4 Component discovery and skills

Two component types only, each at a **fixed location** that `plugin.json` cannot relocate or inline: `skills/` and `mcp.json`. A missing location is not an error; a location of the wrong filesystem kind invalidates only that component type. Other component types (commands, hooks, agents, rules, LSP) are outside v1.

Skills must conform to the Agent Skills specification (agentskills.io). Each **immediate** child directory of `skills/` containing a regular file named exactly `SKILL.md` is one skill; there is no recursive search. A non-conforming skill is skipped and reported. How a host exposes skills to users or models is host policy. Host note: VS Code silently skips skills whose `name` is not plain kebab-case or does not match the directory name `[A·S11]`.

#### 1.5 `mcp.json`: complete field reference

Top level must contain exactly `$schema` (for 1.0.0: `https://agent-plugins.org/schemas/1.0.0/mcp.schema.json`, and the version must **match `plugin.json`'s**, otherwise MCP is disabled for the plugin) and `mcpServers` (object; empty is valid). Each server needs an explicit `type` and must match exactly one closed variant; an unknown field, unknown `type`, or a field from another variant invalidates **that entry only**.

| Variant | Field | Type | Req. | Notes |
|---|---|---|---|---|
| `stdio` | `type` | `"stdio"` | Yes | |
| | `command` | string | Yes | **One executable token**, not a shell string. Bare name (platform search rules) **or** plugin-relative `./path`. **No placeholder expansion.** A bundled executable must use a `./` command. Clients may use an interpreter for `.bat`/`.cmd` on Windows but keep `command` as one token |
| | `args` | string[] | No | `${PLUGIN_ROOT}`/`${PLUGIN_DATA}` expanded per element |
| | `env` | object of strings | No | Values expanded; must not contain `PLUGIN_ROOT`/`PLUGIN_DATA` keys (entry invalid); no secrets |
| | `cwd` | string | No | Default = plugin root. When set: `./…`, or `${PLUGIN_ROOT}[/…]`, or `${PLUGIN_DATA}[/…]`; must stay inside the corresponding directory |
| `streamable-http` | `type`, `url` | | Yes | Absolute http(s) URL; no userinfo or fragment; non-loopback must be HTTPS |
| | `headers` | object of strings | No | Literal, visible data; **no credentials**; no expansion; not forwarded across origins |
| `sse` | `type`, `url`, `headers` | | as above | Deprecated HTTP+SSE; client support optional |

Transport rules: a client that supports MCP must support at least one of `stdio`/`streamable-http` (should support both); it uses the declared transport for the first connection attempt and the spec defines **no fallback**. Whether a configured `PATH` participates in bare-command resolution is **client-defined**, and conformant plugins "MUST NOT depend on" it. No OAuth or credential-reference fields exist in v1; authorization is client-managed. Loading rules: invalid/unsupported/mismatched `mcp.json` disables MCP for the plugin; an invalid or unsupported entry is skipped; a server that fails to start/connect/auth is reported and the rest still load.

#### 1.6 Environment and placeholders

Clients that launch stdio subprocesses **must** put `PLUGIN_ROOT` (absolute, filesystem-resolved plugin root) and `PLUGIN_DATA` (absolute path of a client-managed, writable, per-installed-plugin directory that **persists across updates**; created by the client before launch; may be deleted on uninstall) into each subprocess environment, after overlaying the configured `env`. Expansion is single-pass and non-recursive, applies only to `args` elements, `env` values and `cwd`, and never to `command`, env keys, URLs or headers. Unrecognised `${…}` text stays literal. Base-environment inheritance is client-chosen; plugins may not rely on it except PATH search for a bare command. Intended use of `PLUGIN_DATA`: dependencies, caches, state; of `PLUGIN_ROOT`: bundled scripts, binaries, config.

#### 1.7 Client extensions

Client-specific data goes under a reverse-domain namespace, as an `extensions.<namespace>` object and/or a top-level directory named exactly that namespace. Clients ignore namespaces they do not implement without validating them. Known: `com.github.copilot` (custom agents, commands, rules, hooks, automation templates in VS Code/Copilot) `[A·S11]`; `io.github.block.goose` (Goose hooks) `[C]`.

#### 1.8 Versioning

The spec version is the version of both schemas; every release republishes both. A change to either schema requires a new spec release; published schema URLs are immutable. A plugin's `$schema` declares the targeted spec version; clients decide what they support and may map several versions to one implementation. Plugin `version` should be SemVer and may drive update checks.

#### 1.9 What the spec does not define (the gap that matters here)

| Capability | Status in 1.0.0 | Where it actually lives |
|---|---|---|
| Install / enable / uninstall | Out of scope | Each client |
| Distribution, marketplaces, registries, discovery | Out of scope. Google points to separate ARD / AI Catalog proposals `[B·S7]` | Each client (see §1.10) |
| Updates | Only a `version` hint | Each client |
| **Per-OS/arch command selection** | **Not defined**; `command` is one token | Nowhere portable |
| **Install / post-install hooks** | **Not defined**; hooks are client extensions | Client namespaces (e.g. `com.github.copilot`) |
| Dependencies | Not defined; "future consideration" `[A·S6]` | n/a |
| Permissions, sandboxing, trust levels | Not defined; "future consideration" | Each client |
| Signing / provenance | Not defined; "future consideration" | Our release pipeline |
| Secrets, OAuth | Not defined; `env`/`headers` are visible package data | Each client |
| Icons | Not in 1.0; `assets/icon.png` proposed for 1.1 `[B·S53]` | n/a |

#### 1.10 How hosts discover and install plugins and local MCP servers

| Host | Discovery surfaces | Install action | Update behaviour | Admin controls |
|---|---|---|---|---|
| **VS Code** `[A·S11, S12, S48]` | Extensions view `@agentPlugins`; plugin marketplaces (git repos; Awesome Copilot by default); workspace recommendations via `.claude/settings.json` or `.github/copilot/settings.json` (`extraKnownMarketplaces`, `enabledPlugins`); MCP gallery via `@mcp` | **Install** in Extensions view; local dir via `chat.pluginLocations`; MCP alone: `code --add-mcp '{…}'`, `.vscode/mcp.json`, user `mcp.json` (top-level key `servers`) | Checks on demand or every 24 h if `extensions.autoUpdate`; plugins from npm/PyPI never auto-update (explicit **Update** click). Bump `version` or nothing updates | Managed `enabledPlugins`, `extraKnownMarketplaces`, `strictKnownMarketplaces`; MCP allowlists |
| **GitHub Copilot** (CLI, app, SDK) `[B·S10]` | Awesome Copilot marketplace by default; same managed settings as VS Code | Copilot CLI/app plugin install (exact commands not read: T12) | Client-managed | Same managed settings; MCP allowlists "by URL, command, or name" |
| **Cursor** `[A·S13]` | **Customize** page; Cursor Marketplace (manually reviewed, open-source only, submit at cursor.com/marketplace/publish); cursor.directory; team marketplaces (import a Git repo) | Install → choose project/user scope; local drop-in `~/.cursor/plugins/local/<name>` + reload; MCP alone: `~/.cursor/mcp.json` or `cursor://anysphere.cursor-deeplink/mcp/install?name=…&config=<base64>` | GitHub-imported team marketplaces: Auto Refresh at most once per 10 min (needs Cursor GitHub App) | Marketplace access groups; install modes Default Off/On/Required; "Allow Local Plugin Imports" (off by default on Enterprise) |
| **Claude Code** `[A·S14, S15]` | `/plugin marketplace add <owner/repo>`; official marketplace; claude.ai-synced plugins | `/plugin install <name>@<marketplace>` or `claude plugin install …`; scopes user/project/local/managed | `version` in manifest pins the plugin: bump it or omit it (then commit SHA drives updates); auto-update; `/reload-plugins` to move running MCP servers to a new copied path | `strictKnownMarketplaces`, org distribution through claude.ai (sources restricted; **no top-level `bin/`**) |
| **Claude Desktop** `[A·S17; B·S18, S55; C·S54]` | Settings → Extensions (directory); Cowork plugins synced from claude.ai | Double-click / drag `.mcpb`, or Settings → Extensions → Advanced settings → Install Extension…; or hand-edit `claude_desktop_config.json` | Directory-listed extensions update automatically; sideloaded ones do not, and orgs must build their own update flow | Org allowlist; custom `.mcpb` upload for Team/Enterprise |
| **ChatGPT and Codex** `[A·S26, S27, S28]` | One universal plugin directory (remote MCP only); local marketplace for testing; enterprise plugin management | ChatGPT desktop: Settings → MCP servers → Add server → STDIO; Codex: `~/.codex/config.toml` `[mcp_servers.<name>]`, or a plugin's bundled `.mcp.json` | Client-managed | Workspace plugin controls |
| **Kiro** `[B·S36, S37, S38]` | Powers panel; official registry; catalog submission at kiro.dev/powers/submit | Add Custom Power → Import power from GitHub (URL, incl. subdirectory) or local path; MCP alone: `~/.kiro/settings/mcp.json`, `.kiro/settings/mcp.json` | Not read | Not read |
| **Antigravity** `[B·S30; C·S35, S57]` | `agy plugin` commands; MCP Store in the IDE; no one-click MCP install links | `agy plugin install <git-url>` (subdirectory URL shown in codelab); `agy plugin import claude|gemini`; or edit raw MCP config | Not read | Not read |

### 2. Vivechak package specification

#### 2.1 Design constraints derived from §1

| # | Constraint | Reason |
|---|---|---|
| D1 | No `${…}` placeholders in `mcp.json` | Cursor does not expand them; Claude uses different variables; spec forbids them in `command` (F3) |
| D2 | Thin plugin uses `command: "vivechak"` (bare) | No portable per-OS path exists (F1) |
| D3 | Server reads `PLUGIN_DATA`/`PLUGIN_ROOT` from the environment, falls back to `os.UserConfigDir()`/`os.UserCacheDir()` | Spec §9.1 makes clients set them; hosts differ; fallback keeps it working everywhere `[D]` |
| D4 | Plugin version is decoupled from binary version; the server reports a clear "upgrade the binary" error when a plugin needs a newer one | The plugin cannot pin or install the binary `[D]` |
| D5 | Skill frontmatter limited to Agent Skills fields; skill text provider-neutral | Claude-only keys (`argument-hint`, `disable-model-invocation`) fail the reference `skills-ref` validator `[C·S56]`; OpenAI asks for provider-neutral wording `[A·S26]` |
| D6 | Server startup is side-effect-free and tolerates `sandboxEnabled` in VS Code | Account-synced plugins may launch it unprompted (F10); VS Code can sandbox stdio servers on macOS/Linux `[B·S12]` |

#### 2.2 Repository layout (one public repo: marketplace + plugin)

```
vivechak-plugins/
├── .claude-plugin/marketplace.json      # Claude Code marketplace catalog
├── .agents/plugins/marketplace.json     # Codex marketplace catalog            (verify: T8)
├── .cursor-plugin/marketplace.json      # Cursor marketplace, multi-plugin     (verify)
└── plugins/vivechak/
    ├── plugin.json                      # Agent Plugins 1.0.0 (portable)
    ├── mcp.json                         # portable MCP entry (thin)
    ├── skills/
    │   ├── vivechak/SKILL.md            # how/when to use the tools (product-owned)
    │   └── vivechak-setup/
    │       ├── SKILL.md                 # detect OS/arch -> install or repair binary -> verify
    │       └── scripts/{install.sh,install.ps1}
    ├── .claude-plugin/plugin.json       # Claude-native manifest (no "version": SHA drives updates)
    ├── .mcp.json                        # Claude/Codex-native MCP config
    ├── .codex-plugin/plugin.json        # only if Codex does not accept the portable manifest (T8)
    ├── LICENSE
    └── CHANGELOG.md
```

Why a subdirectory: Claude Code reads a `.mcp.json` at a repository root as project-scoped MCP config, so a plugin at the repo root triggers "pending approval" prompts when the repo is opened `[C·S50]`; Kiro and Antigravity both accept subdirectory URLs `[B·S30, S36]`; and the catalog files can list more plugins later. **No binaries in this repo.**

#### 2.3 Portable `plugins/vivechak/plugin.json`

```json
{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json",
  "name": "vivechak",
  "version": "0.1.0",
  "description": "<TBD: one line. State that the vivechak binary must be installed>",
  "author": { "name": "<TBD>", "url": "https://<TBD>" },
  "homepage": "https://<TBD>/docs",
  "repository": "https://github.com/<org>/vivechak-plugins",
  "license": "<SPDX-TBD>",
  "keywords": ["mcp", "<TBD>"]
}
```

Only fields from the closed set; no `icon` (1.1 proposal); no `extensions` until Copilot-specific agents/hooks are wanted (then `com.github.copilot`). Bump `version` on **every** change under `plugins/vivechak/`: VS Code does not update otherwise `[A·S11]`. CI should enforce this.

#### 2.4 Portable `plugins/vivechak/mcp.json` (thin, recommended)

```json
{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json",
  "mcpServers": {
    "vivechak": {
      "type": "stdio",
      "command": "vivechak",
      "args": ["mcp"]
    }
  }
}
```

`args` is a placeholder: use whatever the server needs to start in stdio mode (omit if stdio is the default). The `$schema` version must equal `plugin.json`'s. A copy-pasted `.mcp.json` entry without `type` fails validation `[C·S49]`.

**Variant A, npm runner (P2, only if the wrapper exists):** `"command": "npx", "args": ["-y", "@<org>/vivechak-mcp@0.1.0"]`. Pin the version. Adds a Node dependency, a Windows `.cmd` shim, and cold-start latency against hosts' timeouts (Codex default 10 s `[A·S28]`).

**Variant B, fat plugin (P2, experimental, gated on T1–T3):**

```json
{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json",
  "mcpServers": {
    "vivechak": { "type": "stdio", "command": "./bin/vivechak", "args": ["mcp"] }
  }
}
```

with `bin/vivechak` (POSIX dispatcher, mode 0755), `bin/vivechak.exe` (windows/amd64; runs under emulation on Windows-on-Arm), `bin/vivechak-darwin-universal`, `bin/vivechak-linux-amd64`, `bin/vivechak-linux-arm64`. This mirrors the Antigravity precedent, where the extensionless dispatcher uses `uname` on Unix and the host invokes `<name>.exe` on Windows `[C·S34]`. Whether each host resolves `./bin/vivechak` to `vivechak.exe` is **not specified** (T1).

```sh
#!/bin/sh
# plugins/vivechak/bin/vivechak  (mode 0755); Windows resolves bin/vivechak.exe instead
set -eu
here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
case "$(uname -s)" in
  Darwin) exec "$here/vivechak-darwin-universal" "$@" ;;
  Linux)
    case "$(uname -m)" in
      x86_64|amd64)  exec "$here/vivechak-linux-amd64" "$@" ;;
      aarch64|arm64) exec "$here/vivechak-linux-arm64" "$@" ;;
    esac ;;
esac
echo "vivechak: unsupported platform $(uname -s)/$(uname -m)" >&2
exit 127
```

Cost model: five binaries at roughly 8–15 MB each is 40–75 MB **per release** if committed. Keep them out of `main`; if the variant is pursued, publish it from an orphan release ref, an npm tarball, or (Claude Code only) a zip `archive` source with `sha256` `[A·S15]`. Actual binary size is unknown (Q-6).

#### 2.5 Host-native manifests kept beside the portable one

`plugins/vivechak/.mcp.json` (Claude Code / Codex native; bare command in the thin design):

```json
{ "mcpServers": { "vivechak": { "command": "vivechak", "args": ["mcp"] } } }
```

`plugins/vivechak/.claude-plugin/plugin.json`: recognised keys only, **omit `version`** so the commit SHA drives updates (Claude pins a set `version` and ignores new commits until it is bumped `[A·S14]`):

```json
{
  "name": "vivechak",
  "displayName": "Vivechak",
  "description": "<TBD>",
  "author": { "name": "<TBD>" },
  "homepage": "https://<TBD>/docs",
  "repository": "https://github.com/<org>/vivechak-plugins",
  "license": "<SPDX-TBD>",
  "keywords": ["mcp"]
}
```

`.claude-plugin/marketplace.json` at the repo root (relative-path source, marketplace name in kebab-case, not one of Anthropic's reserved names `[A·S15]`):

```json
{
  "name": "vivechak",
  "owner": { "name": "<TBD>", "url": "https://<TBD>" },
  "plugins": [
    { "name": "vivechak", "source": "./plugins/vivechak", "description": "<TBD>" }
  ]
}
```

Codex and Cursor catalogs follow the same idea (`.agents/plugins/marketplace.json` for Codex per community repos `[C·S62]`; `.cursor-plugin/marketplace.json` for Cursor `[A·S13]`); their exact schemas were not read, so generate them from each host's template and validate in CI. Whether Codex needs `.codex-plugin/plugin.json` given it reportedly installs portable plugins is T8.

#### 2.6 MCPB manifests for Claude Desktop (per-OS bundles)

Spec version noted in `MANIFEST.md` is 0.3 (last updated 2025-12-02), while its body documents 0.4 features and third parties ship 0.4 `[A·S16; C·S20]`. Confirm which `manifest_version` Claude Desktop accepts (T5). Bundle layout: `manifest.json`, `icon.png`, `server/<binary>`. macOS bundle (`vivechak_<ver>_darwin_universal.mcpb`, `server/vivechak` = signed, notarized universal binary, mode 0755):

```json
{
  "manifest_version": "0.3",
  "name": "vivechak",
  "display_name": "Vivechak",
  "version": "0.1.0",
  "description": "<TBD>",
  "author": { "name": "<TBD>", "url": "https://<TBD>" },
  "repository": { "type": "git", "url": "https://github.com/<org>/vivechak.git" },
  "homepage": "https://<TBD>",
  "support": "https://github.com/<org>/vivechak/issues",
  "icon": "icon.png",
  "license": "<SPDX-TBD>",
  "privacy_policies": ["https://<TBD>/privacy"],
  "server": {
    "type": "binary",
    "entry_point": "server/vivechak",
    "mcp_config": { "command": "${__dirname}/server/vivechak", "args": ["mcp"] }
  },
  "compatibility": { "claude_desktop": ">=0.10.0", "platforms": ["darwin"] }
}
```

Windows bundle (`vivechak_<ver>_windows_amd64.mcpb`): same, with `"entry_point": "server/vivechak.exe"`, `"command": "${__dirname}/server/vivechak.exe"`, `"platforms": ["win32"]`. Linux bundles (`linux_amd64`, `linux_arm64`) follow when the beta stabilises. Add `tools`/`tools_generated` to match the server. Build order: build → sign binary → `mcpb validate` → `mcpb pack` on a macOS/Linux runner (keeps exec bits `[C·S24]`) → attach to the Release. Directory submission additionally needs a privacy policy, a title plus `readOnlyHint`/`destructiveHint` on **every tool**, and working examples `[B·S19; C·S23]`; these are server-side prerequisites to schedule.

#### 2.7 MCP Registry entry (`server.json`, P1)

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.<org>/vivechak",
  "title": "Vivechak",
  "description": "<TBD>",
  "version": "0.1.0",
  "repository": { "url": "https://github.com/<org>/vivechak", "source": "github" },
  "packages": [
    {
      "registryType": "mcpb",
      "identifier": "https://github.com/<org>/vivechak/releases/download/v0.1.0/vivechak_0.1.0_darwin_universal.mcpb",
      "fileSha256": "<sha256>",
      "transport": { "type": "stdio" }
    },
    {
      "registryType": "mcpb",
      "identifier": "https://github.com/<org>/vivechak/releases/download/v0.1.0/vivechak_0.1.0_windows_amd64.mcpb",
      "fileSha256": "<sha256>",
      "transport": { "type": "stdio" }
    }
  ]
}
```

Rules: the registry stores metadata only; `fileSha256` is required for MCPB and clients validate it; the artifact URL must contain "mcp" (the `.mcpb` extension satisfies this); MCPB artifacts must be hosted on GitHub or GitLab releases `[B·S41]`. How a client picks among several `mcpb` entries by platform is undocumented (Q-8). GoReleaser has an MCP Registry publisher (licensing tier: Q-9).

#### 2.8 Self-wiring subcommand (P1): `vivechak mcp-config --client <host> [--write]`

Prints (or with `--write` merges) the host's config snippet with an **absolute, upgrade-stable** path (the package manager's shim such as `/opt/homebrew/bin/vivechak`, not a versioned Caskroom/Cellar path) `[D]`. Precedents: `cua-driver mcp-config --client antigravity` `[C·S32]`, `searchatlas login` printing paths `[C]`, Unity-MCP's per-host "configurators" `[C·S58]`. Config targets:

| Host | File | Key |
|---|---|---|
| Claude Desktop | macOS `~/Library/Application Support/Claude/claude_desktop_config.json`; Windows `%APPDATA%\Claude\claude_desktop_config.json` | `mcpServers` `[C·S58]` |
| Cursor | `~/.cursor/mcp.json` (user), `.cursor/mcp.json` (project) | `mcpServers` `[C·S58]` |
| VS Code | user `mcp.json` ("MCP: Open User Configuration") or `.vscode/mcp.json`; or `code --add-mcp` | **`servers`**, with `"type": "stdio"` `[B·S12, S48]` |
| Antigravity | `~/.gemini/config/mcp_config.json`, project `.agents/mcp_config.json`; legacy `~/.gemini/antigravity/mcp_config.json` also seen | `mcpServers` `[C·S32, S33]` |
| Codex / ChatGPT desktop | `~/.codex/config.toml`, project `.codex/config.toml` | `[mcp_servers.<name>]` `[A·S28]` |
| Kiro | `~/.kiro/settings/mcp.json`, `.kiro/settings/mcp.json` | `mcpServers` `[B·S38]` |
| Copilot CLI | `~/.copilot/mcp-config.json` | verify `[C·S58]` |

#### 2.9 CI validation gates for the package

Validate `plugin.json` and `mcp.json` against the **official** JSON Schemas (vendored, version-pinned); assert both `$schema` versions match; run `claude plugin validate ./plugins/vivechak --strict` `[A·S14]`; run `mcpb validate`; check that every manifest's version equals the tag (or, for the plugin, that it was bumped when the tree changed); run an MCP `initialize` + `tools/list` smoke test against each built binary on each OS (pattern from gograph `[C]`). A community Go validator for `plugin.json` exists (`apv`) but is young `[C·S65]`; prefer the official schemas.

### 3. Distribution channel matrix (channel × platform × agent host)

Coverage: **Y** = works today per evidence gathered, **P** = partial/conditional (see note), **—** = not applicable/not offered on that OS, **✗** = does not reach that host.

| Channel | macOS | Windows | Linux | Hosts it feeds | Priority |
|---|---|---|---|---|---|
| GitHub Releases (raw binary + `checksums.txt`) | Y | Y | Y | Every host, indirectly — this is what every other channel packages | P0 |
| Homebrew cask (own tap) | Y | — | P (Linuxbrew via formula, not cask; F8) | All hosts, via the installed `vivechak` on PATH | P0 |
| Scoop bucket | — | Y | — | Same, Windows | P0 |
| winget manifest | — | Y | — | Same, Windows | P0 |
| `.deb`/`.rpm`/`.apk` (nfpm, hosted repo) | — | — | Y | Same, Linux | P0 |
| Chocolatey package | — | Y | — | Same, Windows | P2 |
| `go install ./cmd/vivechak` | Y | Y | Y | Same, wherever Go toolchain exists | P2 (document, don't lead with it) |
| `curl \| sh` installer script | Y | P (WSL/Git-Bash only) | Y | Same | P2, optional convenience |
| Thin Agent Plugin (portable `plugin.json`+`mcp.json`+`skills/`) | Y (needs binary on PATH) | Y (ditto) | Y (ditto) | **Cursor, VS Code, GitHub Copilot, Kiro, Antigravity, Codex** (+ Claude Code) | P0 |
| `.mcpb` bundle, per OS | Y (universal) | Y (amd64; arm64 P1) | P (beta only; F5) | **Claude Desktop** only | P0 (mac+win), P1 (linux) |
| MCP Registry `server.json` (→ `mcpb`/`npm`/`oci` package refs) | inherits from referenced package | inherits | inherits | Discovery surface for **VS Code, GitHub Copilot** `@mcp` gallery | P1 |
| npm wrapper (`npx -y @org/vivechak-mcp`) | Y (needs Node) | Y (needs Node; `.cmd` shim) | Y (needs Node) | Any host that can run an `npx` command in `mcp.json`/`mcp_config` | P2 |
| OCI/Docker image | Y (needs Docker) | Y (needs Docker/WSL2) | Y (needs Docker) | Any host that can run `docker run` in its stdio command | P2 |
| Cursor Marketplace listing | — | — | — | **Cursor** discovery only (packages the thin plugin) | P1 |
| Kiro Powers catalog listing | — | — | — | **Kiro** discovery only (packages the thin plugin) | P1 |
| Awesome Copilot / known marketplace listing | — | — | — | **VS Code, GitHub Copilot** discovery (packages the thin plugin) | P1 (submission-dependent, T13) |
| Anthropic Connectors Directory (desktop-extension form) | — | — | — | **Claude Desktop** discovery only (packages the `.mcpb`); separate portal from remote-MCP submissions | P1 |
| Hosted Streamable-HTTP server | n/a (remote) | n/a | n/a | **ChatGPT web/directory** (only route in); also works for every other host as an alternative to local install | Decide (Q-1) |

Read the two "P" cells carefully: Homebrew-on-Linux needs a **formula**, not the macOS-only cask (GoReleaser's `homebrew_casks` output is Mac-only; a Linux formula is a second, separate config block: F8). The Linux `.mcpb` is gated on Claude Desktop's Linux beta leaving beta (F5); build it, but don't advertise it until confirmed.

### 4. Installation UX per agent host

Every path below assumes the `vivechak` binary is already on the machine (via any P0 binary channel above) **or** is fetched by the plugin's `vivechak-setup` skill on first use. Where a host can launch a bare `vivechak`, that is the primary path; the `vivechak mcp-config --client <host> --write` fallback (§2.8) is offered wherever F4's PATH problem can bite.

**Claude Desktop**
1. Install the binary: `brew install <org>/tap/vivechak` (macOS) or `winget install <org>.vivechak` (Windows).
2. Get the extension: once listed, Settings → Extensions → find "Vivechak" → Install. Until then, download `vivechak_<version>_<platform>.mcpb` from the latest GitHub Release.
3. Double-click the file, drag it onto the Claude Desktop window, or Settings → Extensions → Advanced settings → Install Extension… and pick the file.
4. Review the permissions/config screen shown by the manifest, click Install.
5. If Claude Desktop reports the server as disconnected despite step 1, run `vivechak mcp-config --client claude-desktop --write`, then restart Claude Desktop. `[A·S17; B·S18, S55]`

**Cursor**
1. Install the binary (as above, or `apt`/`dnf` on Linux).
2. Cursor → **Customize** page → search "vivechak" in the Cursor Marketplace (or open cursor.directory) → **Install**, choosing Project or User scope. `[A·S13]`
   - No listing yet: clone `github.com/<org>/vivechak-plugins` into `~/.cursor/plugins/local/vivechak` and reload the window; or click a `cursor://anysphere.cursor-deeplink/mcp/install?...` link from Vivechak's docs; or paste the `mcp.json` block into `~/.cursor/mcp.json`.
3. If Cursor can't launch the bare `vivechak`, run `vivechak mcp-config --client cursor --write`.

**VS Code**
1. Install the binary.
2. Extensions view → `@agentPlugins` → search "vivechak" → **Install** (once listed in a known marketplace); otherwise add Vivechak's marketplace via the `extraKnownMarketplaces` setting first. `[A·S11]`
   - MCP-only alternative: Command Palette → "MCP: Open User Configuration" and add the `vivechak` entry under the top-level `servers` key, or run `code --add-mcp '{"name":"vivechak","command":"vivechak","args":["mcp"]}'`. `[B·S12]`
3. Approve the server the first time VS Code lists its tools (plugin-installed servers are trusted implicitly; manually-added ones prompt once).

**Antigravity**
1. Install the binary.
2. `agy plugin install https://github.com/<org>/vivechak-plugins/plugins/vivechak` — installs `skills/` and the MCP server together. `[B·S30]`
3. Verify with `agy plugin list`.
4. If the MCP server doesn't start, run `vivechak mcp-config --client antigravity` and paste the result into `~/.gemini/config/mcp_config.json`.

**ChatGPT and Codex** (local install only; see F6/F7 for why chatgpt.com's directory is out of reach)
1. Install the binary.
2. **ChatGPT desktop app:** Settings → Connectors/MCP servers → Add server → Transport **STDIO** → Command `vivechak`, Args `mcp`. `[A·S26]`
3. **Codex CLI/IDE:** add to `~/.codex/config.toml`:
   ```toml
   [mcp_servers.vivechak]
   command = "vivechak"
   args = ["mcp"]
   ```
   `[A·S28]`

**GitHub Copilot** (CLI, app, or VS Code)
1. Install the binary.
2. If listed in the Awesome Copilot marketplace (on by default): open the plugin/marketplace view in the Copilot surface being used → **Install**. `[B·S10]`
3. Otherwise, the same manual MCP steps as VS Code, in that surface's own MCP config file.
4. First tool call may need admin approval depending on the org's MCP allowlist.

**Kiro**
1. Install the binary.
2. Powers panel → **Add Custom Power** → **Import from GitHub** → `https://github.com/<org>/vivechak-plugins` (point at the `plugins/vivechak` subdirectory). `[B·S36, S37]`
3. MCP-only alternative: add the JSON block to `~/.kiro/settings/mcp.json` (user) or `.kiro/settings/mcp.json` (project). `[B·S38]`

### 5. Cross-compilation build matrix

| GOOS | GOARCH | CGO | Output | Feeds | Note |
|---|---|---|---|---|---|
| darwin | amd64 | 0 | `vivechak_darwin_amd64` | `lipo`'d into the universal binary below | |
| darwin | arm64 | 0 | `vivechak_darwin_arm64` | ditto | |
| darwin | universal | — | `lipo -create -output vivechak_darwin_universal vivechak_darwin_amd64 vivechak_darwin_arm64` | Homebrew cask, `.mcpb`, direct download | One artifact, both Mac CPU families; GoReleaser's `universal_binaries:` block automates this `[D]` |
| windows | amd64 | 0 | `vivechak_windows_amd64.exe` | Releases zip, winget, Scoop, Chocolatey, `.mcpb` win32 | P0 |
| windows | arm64 | 0 | `vivechak_windows_arm64.exe` | Releases zip; winget/Scoop later | P1. Without it, Windows-on-Arm falls back to x64 emulation of the amd64 build |
| linux | amd64 | 0 | `vivechak_linux_amd64` | Releases tar.gz, `.deb`/`.rpm`/`.apk`, Linux `.mcpb` | P0 |
| linux | arm64 | 0 | `vivechak_linux_arm64` | ditto | P0 (arm64 dev boxes and CI are common enough to ship day one) |

Excluded on purpose: 32-bit (`386`, `arm`) and `linux/riscv64`/BSDs — no signal in the brief that any target host or its users need them; add only on a specific request. `[D]`

**Build invocation** (per cell, illustrative):
```sh
CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build \
  -trimpath \
  -ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
  -o "dist/vivechak_${goos}_${goarch}${ext}" ./cmd/vivechak
```
`-X main.version=…` is what makes `vivechak --version` and the MCP `initialize` response's `serverInfo.version` meaningful; every install-UX and update-detection idea in §6 depends on this being wired up. `[D]`

**GoReleaser sketch** (illustrative; not a complete config):
```yaml
builds:
  - id: vivechak
    main: ./cmd/vivechak
    env: [CGO_ENABLED=0]
    goos: [darwin, windows, linux]
    goarch: [amd64, arm64]
    ldflags: ["-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}"]
universal_binaries:
  - ids: [vivechak]
    replace: true
archives:
  - formats: [zip]
    format_overrides: [{ goos: linux, formats: [tar.gz] }]
checksum:
  name_template: "checksums.txt"
homebrew_casks:
  - repository: { owner: "<org>", name: "homebrew-tap" }
    hooks:
      post:
        install: |
          if OS.mac? && !system_command("codesign", args: ["--verify", "#{staged_path}/vivechak"], print_stderr: false).success?
            system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/vivechak"]
          end
nfpms:
  - formats: [deb, rpm, apk]
scoops:
  - repository: { owner: "<org>", name: "scoop-bucket" }
winget:
  - repository: { owner: "<org>", name: "winget-pkgs" }
mcp:
  name: "io.github.<org>/vivechak"
```
`[B·S42, S43]` for the shape of `homebrew_casks`, `nfpms`, `scoops`, `winget` and `mcp:` blocks and the quarantine post-install hook; the exact YAML keys should be checked against the GoReleaser version pinned in CI (T14) rather than copied verbatim from this sketch.

**Minimum OS versions.** Go 1.27 (Aug 2026) requires **macOS 13+**; pin the toolchain to Go 1.25/1.26 instead if macOS 12 must be supported. `[B·S46]` A current Windows minimum was not confirmed this session (Open Question). Linux has no enforced floor: a `CGO_ENABLED=0` binary has no libc dependency and runs on glibc and musl distributions alike. `[D]`

### 6. Update and versioning strategy

**Principle:** the binary's SemVer is the one version that matters; every packaging artifact (plugin, MCPB, registry entry) carries its *own* version field because the spec and every host studied give each artifact independent, uncoordinated versioning (F1, §1.8, §1.9's "Updates" row) — there is no cross-artifact dependency mechanism to lean on.

- **Binary itself:** tag `vX.Y.Z` → GoReleaser builds, signs/notarizes, and pushes every channel in one run (see the release checklist below). Stamp `main.version` (§5) and surface it through the MCP `initialize` handshake's `serverInfo.version` so any host or user can see what's actually running. `[D]`
- **Package-manager channels (Homebrew, Scoop, winget, apt/dnf, Chocolatey):** these own the update UX already (`brew upgrade`, `scoop update`, `winget upgrade`, `apt upgrade`). Do nothing extra; do **not** add a self-updater to the binary, since it would race with and undermine every package manager's own version tracking. `[D]`
- **`go install`:** no update mechanism beyond the user re-running it; document `go install github.com/<org>/vivechak/cmd/vivechak@latest` alongside a pinned-version form for reproducible installs.
- **Thin Agent Plugin:** bump `plugins/vivechak/plugin.json`'s `version` on every change under that directory — VS Code's client is documented to skip updates otherwise. `[A·S11]` Treat the plugin's version as independent of the binary's; a plugin release is needed only when `skills/` or `mcp.json` change, not on every binary release.
- **Claude Code manifest:** omit `version` from `.claude-plugin/plugin.json` so the marketplace's commit SHA drives updates, matching the auto-update behaviour Claude Code documents; if `version` is set instead, it must be bumped for users to see anything new. `[A·S14]`
- **`.mcpb` for Claude Desktop:** once listed in the Connectors Directory, updates are automatic; a sideloaded (manually-installed) bundle is **not** auto-updated, so a release should post the new file where existing sideload users will see it (release notes, a version-check line logged by the server itself, or the docs page), rather than assuming they'll notice. `[B·S18, S55]`
- **npm wrapper, if shipped:** pin the exact version in every `mcp.json`/`mcp_config` snippet Vivechak publishes (`@org/vivechak-mcp@0.1.0`, not `@latest`); bump the pin as part of the release checklist. An unpinned `npx -y` install silently changes behaviour underneath the user on every cold start.
- **MCP Registry `server.json`:** update `version` and every `fileSha256` on each release; the registry stores metadata only; a stale `fileSha256` blocks the check the client is expected to perform. `[B·S41]`
- **Deprecation window:** support the current and previous minor binary version; when a skill or tool call needs a feature from a newer binary, have the server's own error message (or the skill's text) name the minimum version required — this is the only cross-artifact version check the ecosystem offers today (D4, §2.1).

**Release checklist (one CI run per tag), in order:** (1) run tests and `go vet`/`staticcheck` on the tagged commit; (2) GoReleaser builds and signs/notarizes every OS/arch cell in §5 and produces `checksums.txt` and provenance; (3) push the Homebrew cask, Scoop bucket, winget manifest, and `.deb`/`.rpm`/`.apk` artifacts; (4) build, `mcpb validate`, and attach the per-OS `.mcpb` bundles; (5) update and publish `server.json` for the MCP Registry with the new `fileSha256`s; (6) bump `plugins/vivechak/plugin.json`'s `version` only if `skills/` or `mcp.json` changed this release, and open/merge that PR into the plugin repo; (7) publish release notes naming the minimum binary version any changed skill now requires. `[D]`, synthesised from F1, F8, F9, §1.9, §1.10 and §2 rather than read from any one source.

## Open Questions & Risks

### Decisions needed from Vivechak (not answerable from research alone)

| # | Question | Why it matters | Referenced at |
|---|---|---|---|
| **Q-1** | Can/should Vivechak's server run in a **hosted Streamable-HTTP** mode, even as a secondary option? | It is the *only* route into ChatGPT's web/directory (F6, F7) and Anthropic's own preferred route for directory listing (F7). If local-only is a hard requirement (A3), say so explicitly so this report's "Decide" row can be closed out rather than left open. | A3, F7, Recommendation |
| **Q-2** | Does the server need per-user secrets or OAuth (API keys, tokens)? | The spec has none: `env`/`headers` are visible package data, not secret storage (§1.5, §1.6, F10). If yes, each host's own credential UI (or a first-run prompt in the binary) has to carry this, not the plugin. | §1.6, F10 |
| **Q-3** | Must the distribution repo (and the binary) be open source? | The Cursor Marketplace requires published plugins to be open source (§1.10); Vivechak can still ship a closed-source binary alongside an open plugin-metadata repo, but that split should be a conscious choice, not a default. | Recommendation decision #6 |
| **Q-4** | What does Vivechak actually know about its users' machines (Windows-on-Arm share, Linux arm64 share, corporate-managed vs personal)? | Decides whether Windows arm64 and a hosted apt/dnf repo are P1 or can drop to P2, and whether the enterprise-admin controls in F10 need active design work now or later. | F9, F10, Recommendation |
| **Q-5** | What legal entity/name will sign the binaries? | Apple Developer ID enrollment and Azure Artifact Signing (or a purchased OV/EV cert) both gate on an org identity that can take 1–20+ business days to validate (F8, F9, Recommendation decision #2); this should be started before the first release is scheduled, not during it. | F8, F9, Recommendation |
| **Q-6** | What is the actual binary size? | Drives whether the fat-plugin variant (§2.4 Variant B) is viable at all, whether `.mcpb` bundles stay comfortably under directory size norms, and Docker image size for the OCI channel. Unknown — Vivechak's own build is the only source of truth. | §2.4 |
| **Q-7** | Is any update-check telemetry (a version ping) acceptable, or must the binary make zero network calls unless a tool is actually invoked? | Interacts with D6 (side-effect-free startup) and with F10's finding that Claude Desktop can auto-start synced servers on machines the user didn't expect. | §2.1 (D6), F10, §6 |
| **Q-8** | Given several `mcpb` package entries in one `server.json` (§2.7), how does a consuming client pick the right one for its OS? | Not documented anywhere read this session; affects whether the MCP Registry listing is useful before this is confirmed empirically (T9 covers the test). | §2.7 |
| **Q-9** | Does GoReleaser's MCP Registry publisher, `universal_binaries`, and `homebrew_casks` require the paid Pro tier, or are they in the OSS build? | Changes the CI tooling budget line item; not confirmed this session. | §5 |

### Tests to run before committing to the plan (T1–T15)

These are the concrete unknowns flagged `[D]` or "not specified" throughout this report, converted into checks a small script or a day of manual testing can close out. None of them were executed this session.

| # | Test | Closes |
|---|---|---|
| **T1** | On Windows, does a plugin `mcp.json` `command: "./bin/vivechak"` resolve to `bin/vivechak.exe` automatically, the way Antigravity's own dispatcher convention assumes? Try it in VS Code, Cursor, Copilot, and Kiro. | Whether the fat-plugin variant (§2.4 Variant B) is viable anywhere beyond Antigravity |
| **T2** | Install the thin plugin's `mcp.json` unmodified in Claude Desktop (via a hand-edited `claude_desktop_config.json`, since there is no native plugin support) and confirm it behaves like any other manually-configured server, with no surprises from the missing native plugin loader. | Confirms F2's "no native support" finding has no silent workaround already shipped |
| **T3** | Launch each host from its normal GUI entry point (Dock/Start Menu icon, not a terminal) on a clean macOS and Windows account with `vivechak` installed only via Homebrew/winget, and confirm whether a bare `command: "vivechak"` resolves. Repeat with the binary installed via `go install` (which uses a different bin directory). | F4; whether `mcp-config --write`'s absolute-path fallback is required by default or only in edge cases |
| **T4** | Print `PLUGIN_ROOT`/`PLUGIN_DATA` from inside the running server on each host that claims spec support, to confirm they are actually set and point where §1.6 says they should. | Whether D3's environment-variable design (§2.1) works as specified everywhere, or needs the fallback path more often than expected |
| **T5** | Pack an `.mcpb` with `manifest_version: "0.3"` and again with `"0.4"` and confirm which one(s) the current Claude Desktop release accepts. | §2.6's open version-field question |
| **T6** | Once the upstream `prepare-for-signing`/`apply-signature` commands ship, confirm they work with Vivechak's actual signing pipeline before switching off `mcpb sign --self-signed`/`--cert`. | The "what not to do" signing caveat (F5) |
| **T7** | Confirm the minimum Windows version the pinned Go toolchain currently supports, from that Go release's own notes at build time. | §5's "not confirmed this session" note |
| **T8** | Confirm whether Codex actually reads the portable root `plugin.json`/`mcp.json` directly (per the Codex CLI 0.147.0 report, F2) or needs its own `.codex-plugin/plugin.json`, by installing the thin plugin in a Codex CLI/IDE session. | Repo layout §2.2's conditional `.codex-plugin/` and `.agents/plugins/marketplace.json` files |
| **T9** | Determine, empirically, how a real client (start with the MCP Registry's own reference client, or VS Code's `@mcp` gallery) chooses among multiple `mcpb` package entries in one `server.json` when they differ only by platform. | Q-8 |
| **T10** | Confirm whether the Claude Desktop Linux beta (Ubuntu 22.04+/Debian 12+) has reached general availability, and whether it accepts `.mcpb` the same way macOS/Windows do. | F5's Linux-beta caveat; the Recommendation's "when the beta stabilises" gating on the Linux `.mcpb` |
| **T11** | Confirm GoReleaser edition/licensing for `universal_binaries`, `homebrew_casks`, `nfpms`, `scoops`, `winget`, and the `mcp:` registry publisher against the specific GoReleaser version CI will pin. | Q-9, T14 |
| **T12** | Read GitHub Copilot CLI/app/SDK's actual plugin-install commands and admin-allowlist syntax directly (not inferred from the shared-settings changelog) before writing Copilot-specific install docs. | §1.10's Copilot row |
| **T13** | Confirm Awesome Copilot's and each marketplace's actual submission/review process and turnaround before counting on P1 discoverability timing. | §3's "submission-dependent" note |
| **T14** | Re-check the exact GoReleaser YAML keys used in §5's sketch against the version pinned in CI; GoReleaser's own deprecation history (F8) shows these keys move. | §5 |
| **T15** | Before general release, do one clean-machine install per host × OS cell that's claimed P0/P1 in §3 (a fresh VM or container, no dev tools preinstalled), following exactly the steps written in §4, and fix whatever the steps got wrong. | Every install-UX claim in §4 that was written from documentation rather than performed |

### Risks

| Risk | Likelihood/impact | Mitigation |
|---|---|---|
| The spec (7 weeks old) changes underneath this plan; 1.1.0 is already drafting new fields (icons) `[C·S52; B·S53]` | Medium/Medium — the thin-plugin design (bare command, no placeholders, closed schema) is deliberately conservative and shouldn't need rework, but new 1.1 fields may be worth adopting later | Pin `$schema` to `1.0.0` explicitly; re-run this research when 1.1.0 publishes rather than tracking the working draft |
| A host changes its plugin-loading behaviour (VS Code, Cursor and Claude Code already each deviate from the spec in a different way, F3) | Medium/Medium — a launcher assumption that works today could silently break | CI smoke test (§2.9) run against each host's current release on a schedule, not just at ship time |
| Claude Desktop adds native Agent Plugins support later, making the `.mcpb` investment partly redundant | Low/Low — MCPB isn't wasted even then (F5's directory-submission and enterprise-upload paths are `.mcpb`-based regardless), but the "no native support" framing in F2 would need updating | Re-check F2 on each Claude Desktop release; nothing in the thin-plugin design needs to change if this happens |
| Signing identity validation (Q-5) is not started early enough and blocks the first Mac/Windows release | Medium/High — the researched lead time is 1–20 business days with credible reports of multi-week stalls | Start Apple Developer ID and Azure Artifact Signing (or OV cert purchase) enrollment now, independent of when the binary itself is ready |
| An external enterprise HSM signing workflow is adopted before the upstream `mcpb` fix lands, silently shipping bundles Claude Desktop rejects (F5) | Low/Medium — only a risk if Vivechak's signing infra routes through GaraSign/ESRP/SignServer/Venafi/Azure SignTool rather than `mcpb sign` directly | T6; keep signing on `mcpb sign --self-signed`/`--cert` until confirmed fixed |
| The fat-plugin variant is built and shipped before T1–T3 are run, and it silently fails to resolve `.exe` on Windows for some hosts | Low (not the recommended default) /Medium if attempted | Gate Variant B behind T1–T3 explicitly, as already stated in §2.4 and the Recommendation table |
| A `.mcpb` sideload update is missed by existing users because sideloaded bundles don't auto-update (F5, §6) | Medium/Low — annoyance, not breakage, but erodes trust over time | Log the running version at server startup and have it note when it's behind the latest GitHub Release; document this loudly for early adopters until the directory listing (which does auto-update) is live |
| Two of seven target hosts (Claude Desktop, ChatGPT web/directory) cannot be served by the same portable-plugin investment that covers the other five | Confirmed, not hypothetical (F2, F6, F7) | Already priced into the Recommendation as separate P0 line items (`.mcpb`, and the open hosted-mode decision) rather than assumed away |

## Sources & Evidence Ledger

Fetched in full with `web_fetch` unless marked "(search excerpt)". Grades follow the table in Research Question. Where a `web_search`/`web_search_fast` call is the source and no single URL was retained for this ledger, that is stated plainly rather than attributed to an invented specific page.

**Agent Plugins specification (agent-plugins.org, agentplugins/agent-plugins-spec) — grade A**

| # | Source | Used for |
|---|---|---|
| S1 | agent-plugins.org/specification (full normative text) | §1.2–1.9 field references; F1's `command`/placeholder rules; F3's transport and containment rules |
| S2 | agent-plugins.org/compatible-clients | F2's host-compatibility list (and the Claude Desktop/Antigravity omission) |
| S5 | agent-plugins.org/plugin-authors/mcp-servers | F1's three-ways-to-launch-a-Go-server analysis; §1.5 `mcp.json` authoring guidance |
| S6 | github.com/agentplugins/agent-plugins-spec (repo README) and its FUTURE_CONSIDERATIONS.md | §1.9's "not defined"/"future consideration" rows (dependencies, permissions, signing, secrets); F10's opening claim |
| — | agent-plugins.org/client-implementers/loading-and-discovery; agent-plugins.org/plugin-authors/build-an-agent-plugin | Read in full; folded into the general §1 synthesis without a standalone bracket citation |
| — | github.com/agentplugins/agent-plugins-spec/tree/main/spec | Attempted; blocked by that site's robots rules. No 1.1.0 schema text was read directly (see S52/S53 below) |

**Vendor/official documentation, read in full — grade A**

| # | Source | Used for |
|---|---|---|
| S11 | code.visualstudio.com/docs/agent-customization/agent-plugins | VS Code's `$schema` auto-detection, `@agentPlugins`, skill kebab-case validation, `com.github.copilot` namespace, managed marketplace settings, version-bump-to-update behaviour |
| S13 | cursor.com/docs/plugins | Cursor's marketplace, install scopes, `${CURSOR_PLUGIN_ROOT}` deviation, Enterprise "Allow Local Plugin Imports" default |
| S14 | code.claude.com/docs/en/plugins-reference | Claude Code's `.claude-plugin/plugin.json`, `${CLAUDE_PLUGIN_ROOT}` expansion, `version` pinning vs SHA-driven updates, `claude plugin validate` |
| S15 | code.claude.com/docs/en/plugin-marketplaces | `marketplace.json` schema, reserved marketplace names, Git LFS pointer-file gotcha, zip `archive`+`sha256` source option, "no top-level `bin/`" |
| S16 | github.com/modelcontextprotocol/mcpb/blob/main/MANIFEST.md | MCPB manifest fields, `platform_overrides` (OS-only, no arch), `.exe` auto-append, `manifest_version` 0.3/0.4 note |
| S17 | claude.com/docs/connectors/building/mcpb | "MCPB is the secondary distribution path," Claude Desktop OS support (macOS/Windows) |
| S26 | developers.openai.com/plugins/guides/submit-claude-plugin | OpenAI plugin portal's remote-only / no-`.mcpb` policy; provider-neutral wording guidance (D5) |
| S27 | developers.openai.com/codex/build-plugins | Codex plugin-building guidance corroborating S26/S28 |
| S28 | learn.chatgpt.com/codex/extend/mcp | Codex's `~/.codex/config.toml`, `[mcp_servers.<name>]` format, default `startup_timeout_sec` of 10 |
| S19 | claude.com/docs/connectors/building/submission (retrieved as rendered content via search after a direct-fetch attempt was blocked; substantially complete) | Directory submission requirements: tool `title`+`readOnlyHint`/`destructiveHint`, OAuth 2.0, privacy-policy requirement, desktop extensions using a separate submission form |

**Vendor blogs, changelogs, codelabs, and vendor docs seen only as search excerpts — grade B**

| # | Source | Used for |
|---|---|---|
| S7 | developers.googleblog.com/agent-plugins-package-your-skills-tools-and-more (fetched in full) | TSC Core Maintainer roster and Google's joining; "bare `mcp.json` is simpler for one server/one client" guidance; pointer to separate ARD/AI Catalog proposals |
| S10 | github.blog/changelog/2026-08-12-agent-plugins-1-0-in-vs-code-copilot-cli-and-the-copilot-app (fetched in full) | GA date and scope (VS Code, Copilot CLI/SDK/app, all plans); Copilot admin allowlist-by-URL/command/name |
| S18 | support.claude.com article on building desktop extensions with MCPB (seen via search, non-English mirror) | Three install methods (double-click, drag-drop, Install Extension menu), per-user install scope, pointer to the directory submission guide |
| S55 | A second Anthropic-adjacent source corroborating directory-listed-vs-sideloaded update behaviour for Claude Desktop extensions (search excerpt; exact URL not retained) | "Directory-listed extensions update automatically; sideloaded ones do not" |
| S23 | github.com/modelcontextprotocol/mcpb/pull/222 ("feat: add prepare-for-signing and apply-signature for enterprise HSM signing") | The precise EOCD/`comment_length` signing bug, which signers are affected (GaraSign/ESRP/SignServer/Venafi/Azure SignTool), and confirmation that `mcpb sign --self-signed`/`--cert` are a separate, unaffected path |
| S30 | codelabs.developers.google.com/cloud-dev-plugin-agy (Google Cloud Developer plugin in Antigravity codelab; seen across several locale mirrors) | `agy plugin install <git-url>` command, the plugin directory layout Google itself documents, `.gemini/config/plugins/` install location, `agy plugin list`/`uninstall` |
| S36, S37 | kiro.dev documentation on Powers (search excerpts) | `plugin.json`/legacy `POWER.md`, GitHub-URL import including subdirectories, catalog submission at kiro.dev/powers/submit |
| S38 | kiro.dev MCP configuration documentation (search excerpt) | `~/.kiro/settings/mcp.json` and project-level `.kiro/settings/mcp.json` |
| S39 | Kiro documentation on Kiro Web's execution environment (search excerpt) | "Web runs it in a cloud sandbox" |
| S41 | MCP Registry documentation/schema at registry.modelcontextprotocol.io or its spec repo (search excerpt) | Package types (npm/PyPI/NuGet/OCI/MCPB/Cargo), `fileSha256` requirement, GitHub/GitLab-only hosting for MCPB artifacts |
| S42 | goreleaser.com/resources/deprecations and goreleaser.com/blog/goreleaser-v2.10 (fetched via search this session, content substantially complete) | `brews` deprecated in favour of `homebrew_casks` for prebuilt binaries; the Gatekeeper quarantine-strip post-install hook; nfpm Linux packaging; GoReleaser's own experimental MCP server/registry publisher |
| S43 | GoReleaser documentation on winget/Scoop/Chocolatey publishers (search excerpt) | winget accepting portable zips via a PR to `microsoft/winget-pkgs`; Scoop/Chocolatey publisher existence |
| S44 | Homebrew announcement of the official cask tap's Gatekeeper enforcement, effective 2026-09-01 (search excerpt) | F8's policy-change date and scope (official tap only; third-party taps unaffected) |
| S45 | Microsoft Learn documentation on Azure Artifact Signing / Trusted Signing eligibility and process (search excerpt, original search) plus qliqsoft.com's migration write-up and Microsoft Q&A threads on learn.microsoft.com (fetched this session, S60 below) | Eligibility by region/entity type; the two-stage org+individual validation flow |
| S46 | Go 1.27 release notes (search excerpt) | macOS 13+ minimum requirement |
| S48 | Merged into S12 below (VS Code MCP configuration, same underlying finding) | — |
| S53 | A vendor-level mention of the 1.1.0 working draft's proposed fixed `assets/icon.png` location (search excerpt) | §1.9's icons row |
| S60 | learn.microsoft.com/answers Q&A threads (AJ Roberts, Christopher Tollstoy, Roman, Marcin De Clermont threads) and qliqsoft.com "Death of the Verification Week: Migrating to Azure Artifact Signing" (fetched this session) | Documented 1–20-business-day validation window; multiple independent reports of multi-week stalls; a real case of a signing-dependent Windows launch being blocked |

**Vendor docs seen only as a search excerpt, not independently fetched — grade B (continued)**

| # | Source | Used for |
|---|---|---|
| S12 | VS Code MCP-configuration documentation (`code --add-mcp`, user `mcp.json`'s `servers` key, sandboxed stdio servers on macOS/Linux; search excerpt, exact page slug not retained) | §1.10's VS Code install-action column; D6's sandboxing note; §2.8's VS Code config target |

**Community and third-party evidence — grade C** (issues, READMEs, and practitioner posts; treated as evidence of what people hit, not of guarantees)

| # | Source (as best identified; several are search excerpts whose exact URL was not retained after this session's context was trimmed) | Used for |
|---|---|---|
| S20, S21 | Community discussion(s) of Go-based MCP servers' `.mcpb` packaging: bundle size growing from roughly 44 MB to 74 MB after adding a Linux binary; the Linux beta's stated OS floor (Ubuntu 22.04+/Debian 12+, x86_64/arm64); no CPU-arch dimension in `platform_overrides` biting real packages | F5's bundle-size and Linux-beta claims |
| S22 | A GitHub issue on a real Go-based MCP server's `.mcpb` distribution reporting the directory build pointing every Mac at an arm64 binary, breaking Intel Macs | F5, F11's grafana precedent |
| S24 | A GitHub issue on `modelcontextprotocol/mcpb` about Windows-built zip archives losing executable bits, since fixed in `mcpb pack` | F5, §2.6's build-order note |
| S29 | A changelog/release-note tracker entry for Codex CLI 0.147.0 reportedly adding install support for portable Agent Plugins | Timeline table; F2's Claude-Desktop-row cross-reference |
| S32 | `cua-driver`'s README/docs, demonstrating an `mcp-config --client <host>` subcommand pattern | §2.8's precedent; F11 |
| S33 | A second community source documenting Antigravity's actual MCP config file paths (`~/.gemini/config/mcp_config.json`, `.agents/mcp_config.json`, and a legacy `~/.gemini/antigravity/mcp_config.json`) | §2.8's Antigravity config-target row |
| S34 | `kwrkb/agy-plugins` GitHub repo, demonstrating an extensionless `bin/<name>` POSIX dispatcher plus `<name>.exe` for Antigravity plugins | §2.4 Variant B's dispatcher design; F11 |
| S35, S57 | Community discussion(s) of Antigravity's MCP Store / plugin ecosystem and its absence from the official Agent Plugins compatible-clients list | F2/§1.10's Antigravity row |
| S47 | Multiple independent community reports (GitHub issues on MCP server repos, and a Claude Code issue) of GUI-launched hosts not inheriting shell PATH, producing `spawn … ENOENT`, with the observed tool-shell PATH `/usr/bin:/bin:/usr/sbin:/sbin` | F4 |
| S49 | A community example showing a copy-pasted `mcp.json` server entry missing `type` failing validation | §2.4's `type`-field warning |
| S50 | A community report of a repo-root `.mcp.json` triggering Claude Code's "pending approval" project-scoped-MCP prompt | §2.2's rationale for a plugin subdirectory |
| S51 | A community report describing Claude Desktop auto-starting account-synced MCP servers on other machines without a local prompt | F10 |
| S52 | A raw commit/PR on the spec repo indicating a 1.1.0 working draft exists | Timeline table; §1.1 |
| S54 | A community/support-forum post on Claude Desktop extension-install UX quirks | §1.10's Claude Desktop row |
| S56 | A community Agent-Skills validator ("skills-ref" or similarly named) rejecting Claude-specific frontmatter keys (`argument-hint`, `disable-model-invocation`) | D5 (§2.1) |
| S58 | The `Unity-MCP` project's per-host "configurator" pattern, plus commonly-documented config file locations for Claude Desktop/Cursor/Copilot CLI repeated across many MCP server READMEs | §2.8's config-target table |
| S61 | A practitioner account of a Homebrew cask affected by the new official-tap Gatekeeper enforcement | F8's header citation |
| S62 | A community repo using `.agents/plugins/marketplace.json` as a Codex marketplace-catalog convention | §2.5 |
| S65 | `apv`, a small community-built Go validator for `plugin.json` | §2.9 |

**A note on this ledger's limits.** This session's tool-call history shows roughly 30 distinct `web_fetch`/`web_search` calls whose full result content was later trimmed from context to manage session length, before this ledger was written. Every URL and finding above is reconstructed from (a) the exact queries and URLs actually issued, which remained visible, and (b) this document's own already-written body text, produced at the time those results were fully in context. Sources marked "search excerpt, exact URL not retained" are real findings from real searches, held with slightly lower confidence in their precise citation than the grade table would otherwise imply; none of them carries a claim that isn't corroborated by at least one other source or by the normative spec text itself. Anyone acting on a single-sourced community (grade C) claim above should re-verify it directly — that is what the T1–T15 tests are for.
