I'll start with the Vivechak repo to ground the 9 tools, then survey the CLI landscape.Those are derivative summaries. I'll go to the primary source, clig.dev.clig.dev gives me the baseline conventions. Now the tool-specific surveys: `gh` formatting and environment behavior, plus `kubectl` and `docker` output negotiation.`gh` gives me a clear model: explicit field lists, built-in `--jq`, Go templates, and documented exit codes. Next I'll check its environment controls, then `kubectl` and `docker`.`gh` also shows help topics as a progressive-disclosure mechanism, plus explicit env-var overrides. Now `kubectl` and `docker`.`kubectl`'s format negotiation is the widest in the survey, and it strips `managedFields` from structured output by default. Now `docker`, including its output-stability history.Supabase's CLI now has a global `--agent <auto|yes|no>` flag for "agent detection", which bears directly on the dual-audience question. I'll dig into what it does.Supabase's in-flight PR and its `--agent` flag are significant. I'll read the PR itself to see how they define "agent" versus CI.Google's Workspace CLI (March 2026) is a strong dual-surface case: one binary exposing both CLI and `gws mcp`. I'll read the author's rationale post directly.This is a primary design-rationale source from a tool author. Its follow-up, "The MCP Abstraction Tax", addresses the MCP-versus-CLI layering question directly, so I'll read it too.I'll pull the Anthropic primary source directly rather than rely on the third-party summaries.That covers the "what makes the MCP side effective" baseline. Now the empirical question of whether CLI or MCP is better for agents, including the 2026 Playwright CLI-versus-MCP data.Clipboard behavior is a known fragility, especially over SSH and on headless machines. I'll check the OSC 52 escape-sequence fallback and its real-world limits.Railway is a strong dual-surface case: it reworked its CLI to be agent-ready, runs MCP through or bundled with the CLI, and uses a plan/apply pattern for destructive changes. Now pnpm and Turso.pnpm changes reporter by TTY and also changes some defaults by CI detection, a semantic shift worth noting. Turso next, alongside the question of how tools detect agent environments.I have NN/g's definition and its two conditions for getting progressive disclosure right. Now the dual-surface anchors: `gh api` as the raw escape hatch, and how Docker's CLI sits on its Engine API.`gh api` is confirmed as the curated-commands-plus-raw-escape-hatch model. One quick source check on Docker's CLI and Engine API relationship and its `--format` model.One last implementation-risk check: since `vck`'s language is still an open question in the roadmap (Python, Go, Rust or TypeScript), I'll verify the broken-pipe behavior per runtime.## D-015: CLI design for a dual-audience `vck` (as of 1 Oct 2026)

### Bottom line

Make `vck` a **peer adapter over the same core as the MCP server**: one package, two thin surfaces, with `vck mcp` serving the agent side. It should not wrap the MCP server, and the MCP server should not wrap it.

- **Tree:** keep it flat and 1:1 with the tools.
- **Prompts and plans:** commands that emit them produce byte-exact stdout, with additive `--output` and `--clipboard` destinations.
- **`--json`:** an explicit, versioned envelope identical to MCP `structuredContent`.
- **Interactivity:** gate it on TTY, `--no-input` and CI, not on guesses about who is calling.
- **Defer:** agent auto-detection, output templating and grouped subcommands. Each has a reversal trigger below.

**Premise check that limits confidence.** I couldn't find the nine MCP tools or `vck` in the public repo. `main` at v1.1.0 (22 Sep 2026) is a markdown framework. The README, ROADMAP and AGENTS.md describe a prompt-paste workflow, and the roadmap puts the CLI and engine in future phases with the implementation language still undecided. GitHub blocked automated access to the branch list, so a private branch is possible. The tree below is derived from the public five-step workflow, not from your tool list.

### Evidence scale and limits

- **A:** primary docs or spec, with behavior corroborated in at least two independent mature tools.
- **B:** a primary source from a single tool or author (docs, design post, PR).
- **C:** a secondary report, a single benchmark, or a PR/issue description.
- **D:** my inference.

This approximates your A–E convention; I didn't retrieve the FRAMEWORK.md §5 definitions, so use those if they differ.

I found no controlled studies of agent-versus-human CLI ergonomics. Most findings rest on docs, tool-author rationale and PR or incident evidence. The 2026 items (agent detection, CLI-versus-MCP benchmarks) are months old at most and still moving.

### Survey

| Tool | What matters here | Lesson |
|---|---|---|
| **gh** (+ GitHub REST/MCP) | Curated commands plus `gh api`, a raw authenticated-request escape hatch. Default output is line-based text; `--json` takes an explicit field list (omit it to see the options), with built-in `--jq` and Go `--template`. Documented exit codes: 0, 1, 2 (canceled), 4 (auth needed). The REST API has 600+ operations; the official MCP server exposes 51 tools. | Self-describing fields, documented exit codes, curated surface plus raw hatch |
| **docker** | The CLI is a REST client of the daemon over a Unix socket, with API version negotiation. `--format` takes `json` or a Go template. Case study: Compose v2.21 changed `ps --format json` from an array to line-delimited objects. Its own PR called this breaking, and a maintainer said Compose doesn't follow strict semver. | Machine output is a contract |
| **kubectl** | `-o` accepts json, yaml, name, wide, jsonpath, go-template and custom-columns; `managedFields` is stripped from JSON/YAML unless requested. | Widest format menu; noise removed by default |
| **cargo** | `cargo metadata` emits JSON with an explicit `--format-version`; compatibility is promised within a version and output may otherwise change. | Versioned machine contract |
| **pnpm** | The reporter defaults by TTY (interactive on a TTY, append-only otherwise), `ndjson` serves machines, and `silent` suppresses even fatal errors. CI detection also changes install strictness. | Switch presentation by TTY; be wary of switching semantics |
| **Turso** | One command is a REPL or a one-shot depending on whether SQL is passed. Scalar flags like `--url` print only the value. | Bare-scalar mode for shell substitution |
| **Supabase** | Global `-o env/pretty/json/toml/yaml`, `--yes`, and `--agent auto/yes/no`. An open PR separates "agent" from "CI" detection and describes JSON auto-switching for agents as follow-up work. It also shows two CLI shells with different format enums. | Detection needs an override and a single source |
| **Railway** | Reworked the CLI for agents (`--json`, non-interactive flags, explicit project/service/environment scoping), with MCP bundled in the CLI plus a remote MCP. The local MCP runs through the CLI. Destructive changes in non-interactive sessions need a dedicated confirm flag after plan review. | Scoped confirmation beats blanket `--yes` |
| **Google Workspace CLI** (Mar 2026) | One binary serves the CLI, a stdio MCP server and agent skills from a single schema source. It also offers raw JSON payloads, `schema` introspection and `--dry-run`. | Closest precedent for one source, two surfaces |

### Cross-cutting findings

**Dual-surface shapes [B].** Three shapes recur:

- **Shared core:** gws, as above.
- **MCP over CLI:** Railway's earlier design, which it has since moved away from.
- **CLI over MCP:** FastMCP, MCPorter and mcp2cli turn MCP servers into list/call commands or generated subcommands, on the premise that tool schemas already hold what a CLI framework needs.

My read [D]: generated bridges make a good stopgap and a `gh api`-style raw hatch. They produce schema-shaped UX (no TTY-aware presentation, no clipboard, no example-led help), so they shouldn't be the primary human surface.

**Scale and the CLI-versus-MCP evidence [C].** The context-cost argument against MCP is framed around CLIs with hundreds of commands. At nine tools it is weak, and one review of a ~25-command CLI judged full schema introspection over-engineered. The token evidence is thinner than the discourse suggests:

- Microsoft's Playwright README asserts the CLI is leaner for coding agents.
- One benchmark reported ~114k versus ~27k tokens on a task; another found ~16% versus ~18% of context.
- The mechanism is large artifacts written to disk, and no schema load. That transfers: return a reference, not the body.

**What "effective MCP" means here.** This is Anthropic's guidance, and Anthropic is MCP's steward, so it is primary but not neutral. It calls for consolidated workflow-shaped tools, high-signal responses with readable identifiers, and a concise/detailed switch (their example went from 206 to 72 tokens). It also calls for pagination or truncation defaults with steering messages (Claude Code caps tool responses at 25,000 tokens by default) and actionable errors. The CLI must carry these properties over.

**Output and naming [A baseline; C on `-o`].** clig.dev sets the baseline: primary and machine-readable output on stdout, messaging on stderr, TTY checks to choose presentation, `--json` and `--plain`, and stable machine formats while human output may change. WorkOS turns JSON on automatically when stdout isn't a TTY. That suits status-style commands but would break `vck prompt | pbcopy` [D]. The `-o` flag is split:

- kubectl and Supabase use it for format.
- clig.dev, Repomix and code2prompt use it for an output file.

Agents reason by analogy, so `-o json` against a file-meaning flag would silently create a file named `json` [D].

**Prompt output [B].** The closest analogs:

- Repomix defaults to a file, offers `--stdout` (which suppresses logging), `--stdin`, `--split-output` and token counts.
- Its `--copy` adds a clipboard copy on top of the other destination.
- code2prompt accepts `-` for stdout.
- Its README describes auto-copy. The 4.2.0 argument source has an explicit clipboard flag alongside a deprecated "disable" flag. That reads like a retreat from implicit copying, though I didn't check its changelog.

For clipboard over SSH:

- OSC 52 works in many modern terminals; tmux needs a setting.
- Tools emit it to the terminal rather than stdout.
- Tools built around it guard at roughly 65–75 KB.
- Terminals disable the read path as a leak vector. So pasting results back via clipboard can't rely on OSC 52; stdin must be the universal path.

**Interaction and agent detection [A for TTY gating; C for detection].** clig.dev: prompt only on a TTY, offer `--no-input`, never require a prompt, and tier confirmations (`--force`, or `--confirm=<name>` for severe cases). gh exposes `GH_PROMPT_DISABLED`, `GH_FORCE_TTY` and `NO_COLOR`. A counter-pattern: a March 2026 Tigris PR describes auto-confirming when stdin isn't a TTY. That removes the safety net exactly where agent callers live.

The agent-detection conventions:

- `AI_AGENT` and `AGENT` variables, plus harness-specific ones.
- An open-e2ee PR defaults to JSON under an agent, with `--agent no` restoring text.
- Slack's CLI uses detection only for user-agent and telemetry.
- Detectors need ordered, harness-specific heuristics. Some harnesses mirror others' variables. That is brittle.

**Pipes [A/B].** This is the sharpest failure mode for large text. Node writes to pipes asynchronously on POSIX, so `process.exit()` after writing can truncate output. Two open-source PRs found piped JSON silently cut at 512 bytes and 64 KiB, including for MCP-wrapping consumers. Python's default on an early-closed pipe is a noisy BrokenPipeError, with a documented handling pattern. I did not verify Go or Rust.

**Discovery [B].** Cobra does dynamic completion through a hidden command the completion script calls. NN/g says progressive disclosure needs a good primary/secondary split and an obvious path between levels. That is a GUI principle applied by analogy. Empirical CLI support is modest: a lab study found command suggestions gave faster task success and less documentation checking, and a CHI 2021 study found unstructured text, weak status indication and inaccessible errors.

### Recommendation

**Tree** (illustrative: verbs derived from the public workflow; replace with your real nine):

```
vck <verb> [args] [flags]

  init                       scaffold workspace                    mutator
  generate --from <file|->   emit generator prompt                 EMITTER
  status [--next]            DAG + session/decision state          read
  prompt <session>           emit copy-paste-ready prompt          EMITTER
  save <session> [--from -]  ingest output to research/sessions/   mutator
  decide <D-id>              create/update decision record         mutator
  resolve <conflict-id>      record conflict resolution            mutator
  synthesize                 emit FAD synthesis prompt             EMITTER
  gate                       Phase-0 gate → report + exit code     validator
  ─────
  tools [--json]             MCP tool manifest (schemas, annotations)
  mcp [install <agent>]      stdio server / config writer (explicit opt-in)
  completion <shell>         bash | zsh | fish | pwsh
  help <topic>               output | exit-codes | env | mcp
```

**Behavior contract:**

1. **Parity.** Tool `snake_case` maps to subcommand `kebab-case`, with identical parameter names. `--json` is byte-compatible with `structuredContent`. MCP annotations drive behavior: read-only never prompts; destructive requires `--yes` (prompting on a TTY). A CI test runs per tool. Keep one schema but two prose fields: a model-facing description and a human-facing summary with examples.
2. **Emitters.** Stdout is the artifact, byte-for-byte, regardless of TTY. Any destination flag replaces stdout, and destinations combine. Stderr gets one summary line (bytes, estimated tokens, destination) and one next-step hint, silenced by `-q`. On a TTY with no destination, page the artifact and hint `--clipboard`. Implicit clipboard only via explicit config.
3. **Clipboard.** Native tool first, then OSC 52 to the terminal (never stdout), with a ~64 KB guard. If the guard is exceeded, write a file, say so, and exit with a distinct code. `--from-clipboard` is native-only; stdin (`-`) is universal. Measure the generator and synthesis prompts against the guard before promising clipboard for them.
4. **JSON.** Envelope `{schema_version, ok, data | error{code,message,hint}}`, additive-only within a version. Human text is explicitly unstable. Read verbs show a table on a TTY, line-oriented plain text when piped, and JSON only on request. Offer `--fields` and a concise default with `--full`. Inline artifact bodies up to roughly 10k tokens (my number, under the 25k ceiling); above that, return path, bytes, hash and token estimate. No `-o`; use `--output` and `--format`, so `-o json` becomes an error with a suggestion.
5. **Interaction.** Prompt only when stdin is a TTY and no `--no-input`, `VCK_NO_INPUT` or `CI`. `--yes` skips confirmations only. Overwrites refuse by default, with `--force`. If execution verbs arrive, use plan then apply with a scoped confirm flag.
6. **Agent awareness.** v1 keys only on TTY, CI, `NO_COLOR` and `--no-input`. If you add detection later, keep it presentation-only with an `--agent auto|yes|no` override, never altering artifact bytes or safety prompts. Keep stderr terse, since some harnesses likely show the agent both streams [D].
7. **Exit codes** (documented under `help exit-codes`) [D]: 0 ok; 1 runtime error; 2 usage; 3 gate failed (not a crash); 4 artifact generated but destination failed.
8. **Hardening.** Validate session and decision IDs against strict patterns before building paths. gws treats agent input as adversarial, including traversal and control characters. The MCP adapter confines writes to the workspace.
9. **Pipe conformance suite.** Test `| head -c 100`, `| wc -c`, `| cat`, a closed stdout and a 1 MB artifact. On Node, set the exit code and drain rather than hard-exit. On Python, catch BrokenPipeError.
10. **Discovery.** Bare `vck` prints five lines of help. Each `--help` leads with examples. `vck tools --json` is the lightweight introspection, with no separate schema subsystem. `vck completion` completes session and decision IDs from workspace files, offline and silent on error.

Compositions this enables:

```
vck prompt T1-01 --clipboard          # paste into a research platform
pbpaste | vck save T1-01              # bring the result back via stdin
vck gate --json | jq -e '.data.blocking == 0'
```

### Tradeoffs

| # | Choice (vs alternative) | Grade | What you give up |
|---|---|---|---|
| 1 | Shared core, two adapters (vs bridge or wrapper) | B | Core extraction if handlers are MCP-coupled; superiority over alternatives is untested |
| 2 | Flat verbs, 1:1 with tools (vs noun-verb groups) | C | Stops scaling past ~12 (my threshold); Docker kept flat aliases when it grouped |
| 3 | Byte-exact emitter stdout (vs auto-JSON when piped) | A/C | Agents must pass `--json` explicitly |
| 4 | Additive `--output` + `--clipboard`, no implicit copy (vs default clipboard) | B | One extra flag for humans; config opt-in available |
| 5 | Versioned envelope = `structuredContent`; no templates or jq in v1 (vs docker/kubectl/gh-style) | B / D | Power users pipe to `jq` |
| 6 | No `-o` (vs kubectl-style) | C | Less convenient; safer under the ecosystem split |
| 7 | Native, then OSC 52, then file, with guard and distinct exit code | B | Platform matrix (macOS, Linux, WSL, Windows) and tmux/SSH to test |
| 8 | Concise default, reference for big bodies (vs always inline) | B/C | Extra read round trip; `--full` escape |
| 9 | Non-TTY never prompts; narrow `--yes` (vs Tigris-style auto-confirm) | A | Scripts must pass flags |
| 10 | Defer agent detection (vs Supabase/open-e2ee-style auto-JSON) | C | Agents that forget `--json` get text |
| 11 | Pipe conformance tests | A | Harness effort |
| 12 | Three-tier help plus `tools --json` (vs schema subsystem) | B | Less introspection depth |

### Open risks and reversal triggers

| Risk | Trigger to re-open |
|---|---|
| **Unverified premise:** the nine tools and `vck` aren't in the public repo, and the language is undecided | More than three tools turn out not to be 1:1 workflow verbs, or any is long-running or streaming |
| **CLI-versus-MCP efficacy for your tasks is unknown:** the benchmarks are one vendor's tool, reported secondhand | An Anthropic-style eval on ≥20 real tasks shows MCP matching CLI on success and tokens → cut agent-facing CLI polish, keep human features |
| **Surface drift:** Supabase's two shells show the cost | Parity test fails twice in a quarter → generate flags from tool schemas (FastMCP-style) |
| **Output-contract break:** Compose v2.21 precedent | Any `--json` shape change without a `schema_version` bump → freeze snapshot tests in CI |
| **Clipboard fragility** | More than one nominal artifact exceeds ~64 KB, or clipboard failures dominate issue reports → file output becomes the documented default for those verbs |
| **Runtime pipe traps:** Go and Rust unverified | Language decision lands → add runtime-specific tests |
| **Agent-detection churn:** `AI_AGENT` versus `AGENT` unsettled | ≥2 major harnesses reliably set one variable, and evals show agents omitting `--json` |
| **Execution verbs** (the roadmap's engine phases) change the contract: streaming, cost, resume, concurrency | Execution lands → re-open D-015 for NDJSON events, cost-estimating `--dry-run`, plan/apply |
| **`-o` habit collisions** | "unknown flag -o" recurs in transcripts → pick one meaning and document it |
| **Thin evidence base:** vendor-adjacent, no controlled studies | Your own eval contradicts a B/C row → revisit that row only |

Want me to turn this into a markdown decision record (ADR-style, tables intact) for `DECISIONS.md`?