# Vivechak v0.1.0 — Re-Audit 4 Resolution & Verification Report

**Baseline Commit:** `12b768f7524bfd1d5c9d851f5f3f906105edbeba` (`overhaul/0.1.0`)
**Audit Work Order:** `.agents/audits/vivechak-v0.1.0-reaudit-4.changes.md` & `.agents/audits/vivechak-v0.1.0-reaudit-4.analysis.md`
**Owner Decisions Applied:** `D1 = A` (9 canonical MCP tools, zero aliases, real schemas, canonical envelope, `<= 10 KB` always-loaded / `<= 7 KB` `tools/list`), `D2` (keep distribution channels with `"coming soon"` labels where prerequisites are unverified; fail-closed release preflight; keep PyPI companion `vck`), `D3 = A` (`vck signoff` interactive terminal attestation + red-team challenge record `D-NNN-R<k>.md`), flat `research/` layout, `FUT-01..FUT-05` deferred, `go.mod`/`go.sum` unchanged.
**Engineering Status:** All 17 open re-audit 4 items (`6 CRIT`, `11 HIGH`) implemented and verified by executed checks (`checks.py all`: **218 PASS / 0 FAIL**; `go test -race ./...`: **PASS**; `golangci-lint`: **0 issues**; `direct_mcp` triage: **21/21 PASS**).
**Release Readiness:** Code and CI changes complete; tagging `v0.1.0` remains gated on the external **Owner Actions** listed in §5.

---

## 1. Item-by-Item Resolution Accounting

| Item | Summary of Implemented Change | Key Files Modified | Executed Verification |
|---|---|---|---|
| **CRIT-19** | `vivechak_save(kind="plan")` resolves decision vs. project scope from `id` (`D-*` vs. `DAG`/`PROJECT`) or plan block without misrouting to comparison after an earlier comparison in the same workspace; refuses `C-*` IDs or `scope="comparison"` on `kind="plan"`. | `internal/mcp/tool_save.go`, `internal/workspace/inspect.go` | `checks.py CRIT-19` (PASS) |
| **CRIT-13** | Tightened G6/G7 citation parser in `internal/core/gate.go` so only canonical grade tags or ledger Grade/Verification cells with `https://` URLs count, and only exact `fetched` / `human` (`human-provided`) labels satisfy G7 (`recalled`, `cached`, `verified`, `unverified` rejected). Enforced current `D-NNN-R<k>.md` challenge hash in G10 and interactive TTY `vck signoff` (`E-HUMAN-ONLY-FIELD` on MCP writes). | `internal/core/gate.go`, `internal/core/signoff.go`, `cmd/vivechak/signoff.go`, `skills/vivechak/SKILL.md` | `checks.py CRIT-13` (PASS), `checks.py CRIT-15` (PASS), `checks.py HIGH-16` (PASS) |
| **CRIT-04** | Updated `.goreleaser.yml` Homebrew cask hook (`system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/vck"]` gated on `OS.mac?`) and verified `install.sh` / `install.ps1` checksum matching against multi-artifact `checksums.txt` containing SBOMs. | `.goreleaser.yml`, `install.sh`, `install.ps1` | `checks.py CRIT-04` (PASS) |
| **CRIT-14** | Removed stale/false claims across docs, templates, and prompts; re-ran and captured real `examples/*.md` tool traces; tightened `tests/harness/runner/triage.mjs` so `isError: true` on key calls fails triage; updated `direct_mcp.mjs` to canonical tool/parameter names; updated `SECURITY.md` and `docs/TROUBLESHOOTING.md`. | `examples/*.md`, `SECURITY.md`, `docs/TROUBLESHOOTING.md`, `tests/harness/runner/triage.mjs`, `tests/harness/runner/adapters/direct_mcp.mjs` | `checks.py CRIT-14` (32/32 PASS), `run_harness.mjs --triage` (21/21 PASS) |
| **CRIT-01** | Added fail-closed `preflight` job in `.github/workflows/release.yml` verifying required secrets (`TAP_GITHUB_TOKEN`, `NPM_TOKEN`, `PYPI_API_TOKEN`) and ownership/HTTP reachability of npm `vivechak`, PyPI `vivechak`, and PyPI `vck` before any build or GitHub Release creation; kept version-synced `npm/` and `pypi/` (`vivechak` + `vck`) runners. | `.github/workflows/release.yml`, `npm/package.json`, `pypi/pyproject.toml`, `pypi/vck/pyproject.toml` | `checks.py CRIT-01` (PASS) |
| **CRIT-03** | Labelled unverified channels (`Homebrew`, `npm`/`npx`, `PyPI`/`uvx`, `WinGet`, `Scoop`, `Docker`/`GHCR`, `Nix`, `Smithery`, `npx skills add`) as `"coming soon"` across `README.md` and `docs/`; kept release preflight fail-closed before publishing any artifact. | `README.md`, `docs/QUICKSTART.md`, `docs/HOST-SETUP.md`, `.github/workflows/release.yml` | `checks.py CRIT-03` (PASS) |
| **HIGH-15** | Server-side sequential `C-NNN` / `D-NNN` allocation, session-scoped `D-NEW` / `D-NNN` placeholder resolution, `E-ID-INVALID` refusal on malformed IDs (`Role`), canonical `research/D-NNN-<slug>.md` naming on seeded ADRs, frontmatter stamping (`id`, `type`, `title`, `date`, `status`, `vivechak`), zero `research/templates/` copies, and dynamic ID guidance in the `decision` prompt. | `internal/core/ids.go`, `internal/mcp/id_session.go`, `internal/mcp/tool_save.go`, `internal/mcp/prompts.go` | `checks.py HIGH-15` (PASS) |
| **HIGH-13** | Seeded auto-drafted ADR titles from the decision plan heading or session title (never untitled when available) and kept placeholder review triggers bracketed so G8/G9 block unedited drafts. | `internal/core/draft_decision.go`, `internal/mcp/tool_record_decision.go` | `checks.py HIGH-13` (PASS) |
| **HIGH-14** | Unified workspace resolution across all 9 tools: ignores `WORKSPACE` / `PROJECT_ROOT` and unexpanded `${...}` values, requests `roots/list` only when the client advertises `roots` capability (at most once per session), and refuses unanchored subdirectories in `vivechak_start` with `E-ROOT-UNRESOLVED` without creating files. | `internal/workspace/resolve.go`, `internal/mcp/roots.go`, `internal/mcp/tool_*.go` | `checks.py HIGH-14` (PASS) |
| **HIGH-17** | Removed duplicate verdict and track keys (`gate_passed`, `gate_status`, `track_a_result`, `track_b_result`, `track_b`, `quality_checks`, `mechanical_quality`, `structural_completeness`) from `vivechak_gate` `data`, stating verdict once via `verdict` + `rules`. | `internal/mcp/tool_gate.go` | `checks.py HIGH-17` (PASS) |
| **HIGH-18** | Added live subprocess regression assertions across `gate_targets_test.go`, `write_guard_test.go`, `workspace_rules_test.go`, `artifact_ids_test.go`, `validate_comparison_test.go`, and HIGH-05 untrusted-data wrapper / red-team quote tests; removed loose two-value test assertions (`not_applicable` / `PASS`, `C-01` / `CMP-01`) and clean `vck serve` EOF exit (`serve_test.go`). | `cmd/vivechak/*_test.go`, `internal/core/*_test.go`, `internal/mcp/*_test.go` | `go test -count=1 ./...` (PASS); 0 two-value hits |
| **HIGH-11** | Added `ConfigSafety` subprocess tests in `cmd/vivechak/config_safety_test.go` covering a 50,000-line `~/.claude.json` `--dry-run` (`< 2s`, `+` vivechak diff) and single-line configs containing `GH_PAT`, `DATABASE_URL`, and `args` secrets (asserting zero secret leakage on stdout). | `cmd/vivechak/setup.go`, `cmd/vivechak/config_safety_test.go` | `checks.py HIGH-11` (PASS); `go test ./cmd/vivechak -run ConfigSafety -v` (PASS) |
| **HIGH-04** | Removed unanchored `cellar`/`caskroom`/`_npx` substring heuristics in `cmd/vivechak/config.go` so a non-`PATH` binary writes its absolute path and prints an upgrade warning; restricted presets to the 10 doc-verified hosts (`internal/hosts/presets.go`, `cmd/vivechak/presets.go`) with per-preset fixture and launcher tests (`preset_fixtures_test.go`). | `cmd/vivechak/config.go`, `cmd/vivechak/presets.go`, `internal/hosts/presets.go`, `cmd/vivechak/preset_fixtures_test.go` | `/dev/shm/.../Cellar/.../vck setup cursor` writes abs path + warning; `Preset` tests PASS |
| **HIGH-06** | Settled the 9-tool MCP surface: deleted all parameter aliases and prompt alias middleware (`vivechak_*` / `vck_*`); added explicit `required` + `enum` input schemas with `additionalProperties: true` (`W-UNKNOWN-ARG`) and zero `outputSchema`; standardized `Envelope` (`{ok, code?, message, fix?, data, warnings: [{code, message, fix}], next: {tool, args, why}, truncated?}`); deduplicated `brief`; returned existing pipeline status on repeat project `vivechak_start`; shrunk wire context to `9,063 B` (`tools/list = 6,115 B`). | `internal/mcp/schema.go`, `internal/mcp/envelope.go`, `internal/mcp/prompts.go`, `internal/mcp/tool_*.go`, `internal/core/codes.go`, ported tests & docs | `high-06.txt`: `9,063 B <= 10,000 B`, `6,115 B <= 7,000 B`, 28/28 calls have `next.tool` & `next.args` |
| **HIGH-07** | Required a single fenced ```` ```yaml vivechak-plan ```` block on every saved plan (`V-PLAN-MISSING-BLOCK` and `V-PLAN-*` codes), deleted Markdown table heuristics, selected synthesis solely via `kind: synthesis`, preserved planner text byte-for-byte in `ReplanInPlace`, and updated `GENERATOR-PROJECT.md` and `GENERATOR-DECISION.md` (+ embedded twins). | `internal/core/dag.go`, `internal/core/dag_plan_block_test.go`, `GENERATOR-PROJECT.md`, `GENERATOR-DECISION.md`, `internal/embed/generators/*` | `go test ./internal/core -run Plan -v` (PASS) |
| **HIGH-08** | Single-sourced evidence definitions in `internal/method/evidence.json` (`tools/docgen/evidence.go`), generated `docs/reference/{tools,codes,gate}.md` and `docs/hosts.md` from live Go registries (`78` emitted codes == `78` registered == `78` documented, enforced by `TestAllCodes_ExactSourceParity`), reduced `docs/MCP-TOOLS.md` and `docs/TROUBLESHOOTING.md` code catalogues to links, and extended `DocsFlows` (`cmd/vivechak/docs_flows_test.go`) to schema-check and live-replay all 70 inline `vivechak_*(...)` calls and 9 scope sequences. | `tools/docgen/main.go`, `tools/docgen/evidence.go`, `internal/core/codes_test.go`, `cmd/vivechak/docs_flows_test.go`, `docs/*` | `checks.py CRIT-16` (PASS); `DocsFlows` PASS + fails on scratch alias/missing-arg edits |
| **HIGH-09** | Rewrote `README.md` (`76` lines `<= 150`) in the specified order (pitch, who & why, how it fits, real proof excerpt + real gate FAIL output, quickstart, how it works, requirements & limits, upgrade/uninstall, links); corrected `examples/dogfood/research/` byte claims (`6,115 B` / `9,063 B`) and removed agent self-signed flags; updated `Makefile` `gate` target to evaluate `examples/dogfood` in a temporary copy so `git status` stays clean. | `README.md`, `Makefile`, `.github/workflows/ci.yml`, `examples/dogfood/research/*` | `wc -l < README.md` = `76`; `make gate && git status --porcelain` clean |

---

## 2. Executed Verification Summary

All commands below were executed in an isolated sandbox (`/tmp/vck-impl.DIS2Xf`, offline `GOMODCACHE`, `GOPROXY=off`, throwaway `HOME`):

1. **Audit Harness (`checks.py all` across all 25 active + regression-guard suites):**
   - `PASS=218`, `FAIL=0` (`== 0 failing assertion(s)`, `rc=0`).
2. **HIGH-06 Wire & Surface Measurements (`mcp.py` against compiled `vck` binary):**
   - `initialize.instructions`: `584` bytes (`<= 900` limit)
   - `json.dumps(tools/list)`: `6,115` bytes (`<= 7,000` / `7 KiB` limit; 9 tools, all listing `required`, 0 `outputSchema`, 0 alias parameters)
   - `json.dumps(prompts/list)`: `1,032` bytes (4 canonical prompts `compare`, `decision`, `project`, `gate`, each rendered `<= 600` chars; `prompts/get vivechak_compare` returns JSON-RPC `-32602`)
   - `json.dumps(resources/list)`: `1,332` bytes (8 canonical `vivechak://` resources)
   - **Total always-loaded wire payload:** **`9,063` bytes** (`<= 10,000` / `10 KiB` limit)
   - `vivechak_status(bogus_extra=1)` returns `isError: false`, `ok: true`, `warnings: [{code: "W-UNKNOWN-ARG", ...}]`
   - 28/28 successful tool calls across Comparison, Decision, and Project scopes return non-empty `next.tool` and object `next.args`.
3. **HIGH-08 Single-Source Documentation & Inline Replay:**
   - `go run ./tools/docgen && git diff --exit-code` exits `0`.
   - `TestAllCodes_ExactSourceParity`: `78` codes emitted across `cmd/` + `internal/` == `78` entries in `core.AllCodes()` == `78` rows in `docs/reference/codes.md` (0 undocumented, 0 dead).
   - `go test -count=1 ./cmd/vivechak -run DocsFlows -v`: replays `SAMPLE-COMPARISON.md`, `SAMPLE-DECISION.md`, `SAMPLE-PROJECT.md`, all `70` inline `vivechak_*(...)` calls, and `9` documented scope sequences over live `stdio`; verified in scratch copies that dropping `content=` in `docs/QUICKSTART.md` or changing `id=` to `decision_id=` in `AGENTS.md` fails `TestDocsFlows_InlineCallsAndScopeSequences`.
4. **Go Quality, Race Detector, Linter & Embed Parity:**
   - `gofmt -l .`: empty (`0` unformatted files).
   - `go vet ./...`: `PASS`.
   - `golangci-lint run ./...` (`v2.14.0`): `0 issues`.
   - `make embed-sync` (`TestEmbeddedFilesMatchRoot`): `PASS`.
   - `go test -count=1 ./...`: `PASS` across `cmd/vivechak`, `internal/core`, `internal/embed`, `internal/mcp`, `internal/method`, `internal/store`, `internal/version`, `internal/workspace`, `tools/docgen`.
   - `CGO_ENABLED=1 go test -count=1 -race ./...`: `PASS` across all packages (`0` data races).
   - `make gate`: `Phase 0 Gate: PASS` on `examples/dogfood` copy with `git status --porcelain` empty.
5. **Autonomous Multi-Harness Runner (`direct_mcp`):**
   - `node scripts/run_harness.mjs --harness direct_mcp --batch all` followed by `node scripts/run_harness.mjs --triage`: **`21 / 21 PASS` (`100.0%`)**, `0` schema rejections, `0` retired tool names.

---

## 3. Judgment Calls Logged

1. **Concise Wire `Description` vs. Rich `Doc` on `ParamSpec` (`internal/mcp/schema.go`):** To keep `json.dumps(tools/list)` at `6,115` bytes (`< 7 KB`) and total always-loaded context at `9,063` bytes (`< 10 KB`) without losing human documentation detail in `docs/reference/tools.md`, `ParamSpec` separates short wire `Description` (`<= 100` chars, omitted when self-explanatory from parameter name + `enum`) from full `Doc` text rendered by `tools/docgen`.
2. **Canonical `session_id` vs. `id` Across Tools (`HIGH-06`):** `vivechak_save`, `vivechak_record_decision`, `vivechak_review`, and `vivechak_amend` take `id` (with `session_id`/`decision_id`/`target` aliases deleted); `vivechak_next` and `vivechak_replan` take `session_id` (with `id`/`action`/`new_prompt`/`new_dependencies`/`impact` aliases deleted), matching the `HIGH-06` specification and `fixtures.py`.
3. **Shared `internal/hosts` Package (`HIGH-04` / `HIGH-08`):** Host preset definitions and documentation URLs live in `internal/hosts/presets.go` and are re-exported in `cmd/vivechak/presets.go` (preserving URL comments in `cmd/vivechak/presets.go`) so `vck setup` and `tools/docgen` (`docs/hosts.md`) share one source without introducing package cycles.
4. **Reduced `docs/MCP-TOOLS.md` and `docs/TROUBLESHOOTING.md` (`HIGH-08`):** Kept both files at their existing paths so inbound links and `CRIT-14` checks remain valid, while replacing their duplicated per-tool parameter tables and error/coaching code catalogues with links to the generated `docs/reference/{tools,codes,gate}.md` files.

---

## 4. Deferred Items (Per Owner Scope)

`FUT-01` (blind decision-quality study), `FUT-02` (`vck verify-links` opt-in citation checker), `FUT-03` (GitHub Action + review-due reminders), `FUT-04` (gradual re-enablement of `"coming soon"` channels with provenance/notarization after dated smokes), and `FUT-05` (MADR/log4brains export & Spec Kit hand-off) remain deferred post-`v0.1.0`.

---

## 5. Remaining Owner-Only Actions Before Tagging `v0.1.0`

These steps require external account credentials, GitHub repository administration, or interactive multi-OS/host sessions and were **not** performed here:

1. **Claim Package Names (`CRIT-01`):** Publish initial pre-releases under your accounts for PyPI `vck`, PyPI `vivechak`, and npm `vivechak` so `https://pypi.org/pypi/vck/json`, `https://pypi.org/pypi/vivechak/json`, and `https://registry.npmjs.org/vivechak` each return HTTP `200`.
2. **Configure Release Secrets (`CRIT-01`, `CRIT-03`):** Set `TAP_GITHUB_TOKEN` (`contents: write` on `bhaskarjha-dev/homebrew-tap` and, if Scoop is kept in `.goreleaser.yml`, `bhaskarjha-dev/scoop-bucket`), `NPM_TOKEN`, and `PYPI_API_TOKEN` (covering both `vivechak` and `vck`, or trusted publishing per project).
3. **Channel Prerequisites (`CRIT-03`):** Create `bhaskarjha-dev/scoop-bucket` (or remove Scoop from `.goreleaser.yml` until ready); keep `"coming soon"` labels on channels until smoked from live registries (`VCK_ONLINE=1 checks.py CRIT-03`).
4. **Repository Governance & Merge Order:** Enable GitHub Private Vulnerability Reporting, add a tag protection ruleset for `v*`, open a pull request from `overhaul/0.1.0` to `main`, verify PR checks (`Test`, `release-dryrun`, `Govulncheck`) pass, enable branch protection on `main`, and **merge to `main` before tagging** (so `main/install.sh` and `main/install.ps1` have the fixed checksum verification).
5. **Release-Candidate Rehearsal & OS Smoke Matrix (`CRIT-03`, `CRIT-04`):** Push `v0.1.0-rc.1` from `main` to rehearse preflight + prerelease artifacts; test missing-secret fail-closed behavior on a test fork; verify `brew install bhaskarjha-dev/tap/vivechak` on clean macOS arm64, `install.ps1` on Windows 11 x64, and `install.sh` on Ubuntu x64.
6. **Three Real Host-Agent Example Runs (`HIGH-09`):** Execute one comparison, one decision (through interactive `vck signoff`), and one small project in a live web-search-enabled agent host, commit the resulting artifacts under `examples/<scope>/research/` recording model and date, and add those paths to CI's `vck gate` step.
