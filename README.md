# <img src="docs/assets/logo.svg" alt="" width="40"> Vivechak (विवेचक)

[![CI](https://github.com/bhaskarjha-dev/vivechak/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/bhaskarjha-dev/vivechak/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)

**Evidence-graded technical decisions for AI coding agents, made before the code is written.** Vivechak is an MCP server and CLI (`vck`, also installed as `vivechak`): your agent researches a hard-to-reverse choice with its own web search, Vivechak structures and grades what it finds and saves it as Markdown in your repository, and `vck gate` checks the decision trail before anyone writes code.

## Who it is for, and why
Developers who work with an AI coding agent in an MCP host (Claude Code, Cursor, VS Code, Windsurf, Zed and others) and face a choice that is expensive to undo: a datastore, a protocol, a core dependency, a public API, a security boundary. Asked directly, an agent answers from memory and leaves no record. With Vivechak it compares the options with graded, sourced claims, records the decision and the alternatives it rejected as an ADR, and stops for a person on one-way doors.

## How it fits
- **Deep-research tools** write a one-off report; nothing checks it or keeps it with your code.
- **ADR tools** (MADR, log4brains) give you a template, not the research behind it.
- **Spec kits** (GitHub Spec Kit) plan the build once the choices are made.
- **Vivechak** is the evidence-graded decision step before the plan, kept in `research/` next to your code.

## Proof
The gate refuses an ADR that was auto-drafted through MCP (`vivechak_record_decision` with `auto_save=true`) and then marked `accepted` by editing the file, placeholders still in. Real output of a 0.1.0 development build on 2026-11-16 (excerpt; exit code 1):
```text
$ vck gate
Phase 0 Gate: FAIL
  • Track B (One-Way Doors): FAIL
  …
  • G8: FAIL - One-Way Door ADR(s) contain unexpanded placeholders or default template text: D-001
  • G9: FAIL - One-Way Door ADR(s) lack measurable reversal trigger: D-001 ("[Condition or date for mandatory re-evaluation]")
  • G10: FAIL - D-001 lacks an adversarial red-team challenge record (D-001-R1.md); run vivechak_review(mode="red_team", id="D-001")
  • G11: FAIL - D-001 requires human architectural sign-off (run 'vck signoff D-001 --reviewer "<name>"')
```

## Quickstart
```sh
curl -fsSL https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.sh | sh
vck version
vck setup claude-code   # run in your project; or cursor, vscode, claude-desktop, agy, windsurf, zed, roo, opencode, goose
# then ask your agent: "Use Vivechak to compare Pino and Winston for our Node.js service."
vck gate                # NOT_APPLICABLE after a comparison; PASS or FAIL with reasons once there is a decision
```
Windows: `irm https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.ps1 | iex`. Also available via Homebrew (`brew install --cask bhaskarjha-dev/tap/vivechak`), npm (`npx -y vivechak …`), PyPI (`uvx vck …`), Scoop (`scoop bucket add vivechak https://github.com/bhaskarjha-dev/scoop-bucket && scoop install vivechak`), WinGet (`winget install bhaskarjha-dev.Vivechak`), Docker (`ghcr.io/bhaskarjha-dev/vivechak`), Nix (`nix run github:bhaskarjha-dev/vivechak`), Agent Skills (`npx skills add bhaskarjha-dev/vivechak`), Smithery, the [release archives](https://github.com/bhaskarjha-dev/vivechak/releases), deb/rpm/apk/Arch packages, and `go install github.com/bhaskarjha-dev/vivechak/cmd/vivechak@latest` (see [Quickstart](docs/QUICKSTART.md)). Any other MCP host: `vck setup <path-to-its-mcp-config>` ([hosts](docs/hosts.md)).

## How it works
- **Three scopes.** `vivechak_start` gives the agent a research brief for a *comparison* (2–5 named options → weighted matrix in `research/C-NNN-comparison.md`), a *decision* (1–3 sessions → ADR in `research/D-NNN-decision.md`) or a *project* (session plan → `research/FAD.md`). Hosts that show MCP prompts also offer `compare`, `decision`, `project` and `gate`.
- **The agent does the research** with its own web search and saves Markdown through `vivechak_save`. Vivechak validates structure and IDs, refuses to replace a saved artifact with different content unless the call is a revision, and keeps the replaced version in `research/.history/`.
- **Every claim carries a grade**, with modifiers and a verification label:
  <!-- docgen:begin evidence-grades (generated from internal/method/evidence.json by tools/docgen; do not edit by hand) -->
  - **Grade A (Primary Empirical):** Official specifications, canonical source code, RFCs, or primary documentation with explicit URL anchor.
  - **Grade B (Secondary Empirical):** Peer-reviewed papers, reproducible third-party benchmarks, published postmortems, or corroborated engineering reports.
  - **Grade C (Vendor & Practitioner):** Vendor marketing, product whitepapers, vendor benchmarks, conference talks, or interested-party documentation.
  - **Grade D (Informal & Opinion):** Blog posts, forum discussions, StackOverflow answers, or model memory recall (capped at Grade D).
  - **Grade E (Speculative & Unsubstantiated):** Unverified hypotheses, uncorroborated parametric recall, or ungrounded assertions.
  <!-- docgen:end evidence-grades -->
- **Decisions become ADRs** with rejected alternatives and review triggers. A one-way-door decision also needs a red-team record (`vivechak_review`) and a person's sign-off, typed in a terminal with `vck signoff <D-id> --reviewer "<name>"`.
- **`vck gate` checks the trail** against rules G1–G12 before you code, and exits 1 on FAIL, so it can run in CI.

## Requirements and limits
- An MCP host whose agent can search the web. macOS, Linux or Windows (amd64 or arm64); no runtime dependencies (Go 1.26+ only for `go install`).
- Vivechak **does not fetch or verify** sources: grades and `fetched` labels are the agent's report, and the gate says so.
- The gate is necessary, not sufficient: it checks structure and the agent's own labels, not the truth of the sources. The sign-off is a typed attestation by the named reviewer; it raises the bar but is not a security boundary.
- The server makes no network calls and sends no telemetry. Its tools write only under `research/` in the project root they resolve (pass `project_root` when your host doesn't report one). `vck setup` edits your host's MCP config and backs up an existing file first.
- Not for writing code, fixing bugs or syntax questions: ask your agent directly.

## Upgrade and uninstall
| Channel | Upgrade | Uninstall |
|---|---|---|
| `install.sh` (macOS, Linux) | Re-run the install command | `sudo rm /usr/local/bin/vivechak /usr/local/bin/vck` (or the same two files in `~/.local/bin`) |
| `install.ps1` (Windows) | Re-run the install command | Delete `%LOCALAPPDATA%\Programs\vivechak` and remove it from your user `PATH` |
| Homebrew | `brew update && brew upgrade --cask bhaskarjha-dev/tap/vivechak` | `brew uninstall --cask bhaskarjha-dev/tap/vivechak` |
| WinGet / Scoop | `winget upgrade bhaskarjha-dev.Vivechak` or `scoop update vivechak` | `winget uninstall bhaskarjha-dev.Vivechak` or `scoop uninstall vivechak` |
| `npx` / `uvx` / Docker / Nix | Ephemeral runners pull the latest release (`docker pull ghcr.io/bhaskarjha-dev/vivechak:latest` for cached images) | Remove cached image/cache if desired (`~/.cache/vivechak` or `%LOCALAPPDATA%\vivechak`) |
| Release archive | Replace both binaries with the new archive's | Delete both binaries |
| deb, rpm, apk, Arch package | Install the new package file (`sudo dpkg -i`, `sudo rpm -U`, `sudo apk add --allow-untrusted`, `sudo pacman -U`) | `sudo apt remove vivechak`, `sudo dnf remove vivechak`, `sudo apk del vivechak` or `sudo pacman -R vivechak` |
| `go install` | Run the `go install …@latest` command again | `rm "$(go env GOPATH)/bin/vivechak"` |

Then delete the `vivechak` entry from your host's MCP config; if `vck setup` changed an existing file, the original is next to it as `<config>.vivechak-bak-<timestamp>`. Your `research/` folders stay; they are your files.

## Links
- **Docs** ([docs/](docs/)): [Quickstart](docs/QUICKSTART.md) · [Host setup](docs/HOST-SETUP.md) · [Hosts and presets](docs/hosts.md) · [Manual workflow](docs/MANUAL-WORKFLOW.md) · [MCP tools](docs/MCP-TOOLS.md) · [Troubleshooting](docs/TROUBLESHOOTING.md) · [Architecture](docs/ARCHITECTURE.md) · [Gate rules](docs/reference/gate.md)
- **Examples** ([examples/](examples/)): illustrative walkthroughs of the [comparison](examples/SAMPLE-COMPARISON.md), [decision](examples/SAMPLE-DECISION.md) and [project](examples/SAMPLE-PROJECT.md) scopes, and Vivechak's own decision records in [examples/dogfood](examples/dogfood/research/)
- **Project:** [Methodology (FRAMEWORK.md)](FRAMEWORK.md) · [Contributing](CONTRIBUTING.md) · [Security](SECURITY.md) · [Changelog](CHANGELOG.md)

*Vivechak* (विवेचक): vi- (apart) + √vic (to sift) + -aka (the one who does), "the discerning analyst". MIT License · Copyright (c) 2026 Bhaskar Jha.
