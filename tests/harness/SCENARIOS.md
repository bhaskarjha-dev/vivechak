# Vivechak Harness Stress Test Scenarios for AI Agents & Small Models
### Human-Realistic Prompts with Typos, Slang, Vague Phrasings, Universal Domains & Ground-Truth Expectations
**Updated for Vivechak v0.1.0** · Covers All 13 MCP Tools · Universal 4-Scope Mental Model · Out-of-Scope Guardrails · AST Code-Fence Immunity · Two-Tier Lookups · Track-Filtered Exit Gate · 21 AI Host Harnesses

This test suite is designed to verify how **small models** (e.g. Claude 3.5 Haiku, Gemini 1.5/2.0 Flash, GPT-4o-mini, Qwen-2.5-Coder 7B/14B) running inside **Kiro IDE**, **Cursor**, **Windsurf**, **Cline**, **Claude Code**, or **Antigravity** interact with the Vivechak MCP server when prompted with realistic, messy, unpolished human requests across **technical, algorithmic, AI agent, and strategic business domains**.

---

## 🚀 Quick Copy-Paste Cheat Sheet (21 Comprehensive Scenarios)

| # | Scope & Target Capability | Realistic Messy User Prompt to Copy-Paste | Target MCP Tool & Key Feature |
|---|---|---|---|
| **01** | **Universal Option Comparison**<br>*(Tech, AI Models, Algorithms)* | `vivechek compare winston vs pino for nodejs api.. need very fast logging bcz traffic is high`<br>*(or: `compare claude 3.5 sonnet vs gpt-4o for coding agents`)* | `vivechak_prepare_generator(scope="comparison")`<br>Auto-scope inference, WEP matrix generation, read-only brief |
| **02** | **Skipped Init Auto-Creation**<br>*(Resilient On-the-Fly Workspace)* | `hey save this plan for our comparison c-01`<br>+ raw comparison markdown plan | `vivechak_save_plan`<br>On-the-fly auto-init (writes `research/.gitignore` with `*.lock`), skips DAG cycle checks, terminal completion `next_step` for comparisons (E-38, E-39) |
| **03** | **Sequential Guided Worker**<br>*(Context Injection & DAG Resolution)* | `ok do next step pls`<br>*(or simply `continue`)* | `vivechak_next_session`<br>Automated DAG traversal, upstream context injection, respects draft status |
| **04** | **Fast Architectural Decision**<br>*(Two-Way Door Fast Spike)* | `bro shuld i use react 19 server actions or tanstack react query... team is 2 junior devs` | `vivechak_prepare_generator(scope="decision")`<br>Two-Way Door ($R=1, N=1, B=1$), routes to 1-session Fast Spike or Confirm & Commit |
| **05** | **High-Stakes One-Way Door**<br>*(Multi-Session Deep Research)* | `vivechak decide: mongo vs postgres for our clinic app... storing patient medical records and lab reports`<br>*(or strategic: `strategic decision: direct enterprise sales vs partner reseller`)* | `vivechak_prepare_generator(scope="decision")`<br>Regulatory/HIPAA $X=3$ or $R=3$, priority of decision markers over project words, routes to Deep Research |
| **06** | **Full Venture / Project Pipeline**<br>*(Multi-Tier Research DAG → FAD)* | `hey vivechak plan my new startup idea. its basically uber but for dog walkers. need gps and stripe`<br>*(or venture: `new venture: decentralized renewable energy peer-to-peer microgrid trading marketplace`)* | `vivechak_prepare_generator(scope="project")`<br>Disentangles 2-col metadata tables from 4-col session tables, outputs `RESEARCH-PIPELINE.md` + initial `DECISIONS.md` |
| **07** | **Quality Coaching & 50-Char Gate**<br>*(Non-blocking `Q-*` Observations)* | `save this session output for T0-01: Gin is fastest framework [Grade A], zero memory, no downsides...` | `vivechak_save_session`<br>Surfaces `Q-GRADE-INFLATION`, `Q-MISSING-CONCERNS`, `Q-LOW-FINDINGS` in `next_step`, enforces $\ge 50$ char threshold |
| **08** | **Adversarial Red-Team Challenge**<br>*(Two-Tier Lookup & Resilient Aliases)* | `bro can u roast my decision? we chose microservices with kafka for 3-person team. use vivechak challenge` | `vivechak_challenge`<br>Accepts `decision_id` and `challenge_type="premortem"` aliases, resolves `research/sessions/` and standalone `research/D-*.md` / `research/C-*.md` |
| **09** | **Incremental Session / ADR Amendment**<br>*(Two-Tier Lookup & Auto-Registry Sync)* | `update session T0-01 with this benchmark finding: pino latency is 12ms under load`<br>*(or: amend standalone ADR `D-01` or comparison `C-01`)* | `vivechak_amend_session`<br>Accepts `content` and `reason` aliases, two-tier lookup patches files in `research/sessions/` or `research/`, auto-recompiles `DECISIONS.md` on ADR amendment (E-41) |
| **10** | **Formal ADR Recording**<br>*(Canonical ADR Compilation)* | `lock decision D-01 as accepted with postgres. reversal trigger is >50k concurrent writes. rejected was mongo` | `vivechak_record_decision`<br>Compiles canonical `research/D-01-*.md` into consolidated `research/DECISIONS.md` registry |
| **11** | **Dynamic DAG Re-planning**<br>*(Requirement Pivot / Invalidation)* | `wait stop! boss said no postgres allowed, oracle only! can vivechak update our plan?` | `vivechak_replan`<br>Guided overview mode or direct mutation with `reason` and `impact` metadata |
| **12** | **Mechanical Phase 0 Exit Gate**<br>*(Active Track Filtering & Decoupling)* | `yo can i start coding the backend now?? run the gate thing in vivechak`<br>*(or: `vck gate --track all`)* | `vivechak_run_gate` / `vck gate`<br>Active track filtering (`track="track_a"`, `"track_b"`, `"all"`), headless CI CLI, isolated blocking issues, decoupled Two-Way Door pass (E-40, E-42) |
| **13** | **Status & Graph Visualization**<br>*(Scope-Aware Diagnostics & Visuals)* | `vivechak status? show me visual graph if u can` | `vivechak_status` + `vivechak_visualize`<br>Scope-aware status: counts standalone `C-*.md` files, terminal `next_step` for comparisons (E-38, E-42), Mermaid/ASCII graph |
| **14** | **Multi-Tier Workspace Audit**<br>*(Automated Dry-Run Discovery)* | `check all our research files for any formatting or evidentiary errors` | `vivechak_validate`<br>Auto-discovers and validates sessions, decision plans (`*-plan.md`), standalone ADRs (`D-*.md`), comparisons (`C-*.md`), and FAD |
| **15** | **AST Code Fence Truncation Immunity**<br>*(CommonMark Fence Tracking E-28)* | `save this session with code block containing bash comments # in discovered concerns` | `vivechak_save_session`<br>`FenceTracker` prevents code `#` comments from truncating sections or triggering `V-OWD-NO-CONCERNS` |
| **16** | **CLI Shorthand Scope Recognition**<br>*(Command Prefix Detection E-34)* | `vck compare winston vs pino` or `vck decide: mongo vs postgres` or `vck plan new b2b saas` | `InferScope` via `vck setup` / `prepare_generator`<br>Instantly routes to comparison, decision, or project without regex ambiguity |
| **17** | **Dual-Source Decision Resolution**<br>*(Zero Split-Brain E-29)* | `run gate on workspace with standalone D-01.md and registry DECISIONS.md` | `vivechak_run_gate` & `vivechak_validate`<br>Scans both standalone files and consolidated registry concurrently with ID deduplication |
| **18** | **Evidence Ledger URL Cross-Referencing**<br>*(Quality Engine Provenance E-32)* | `save session citing CLAIM-01 with Grade A where URL is in Sources Ledger table` | `vivechak_save_session`<br>Cross-references ledger table anchors, excludes table rows from claim counts, 0 false alarms |
| **19** | **Universal Host Auto-Configuration**<br>*(21 Desktop & CLI Harnesses)* | `vck setup cursor`, `vck setup claude-code`, `vck setup continue`, or `vck mcp-config` | `vck setup` / `vck mcp-config`<br>Zero-loss JSON merging, number preservation, and instant host configuration |
| **20** | **Pre-Coding Workspace Integrity Check**<br>*(Automated CLI Doctor)* | `vck doctor` or `vivechak doctor /path/to/project` | `vck doctor`<br>Workspace layout audit, frontmatter checks, orphan detection, and active scope diagnostics |
| **21** | **Out-of-Scope Anti-Pattern Guardrail**<br>*(Direct Code Authoring & Bug Fixing)* | `Write a python function to compute HMAC-SHA256` or `Fix this typescript type error` | **Zero Tool Calls (Skip Vivechak)**<br>Agent directly answers/writes code without calling research tools |

---

## 🔍 Universal Debrief / Audit Follow-Up Prompts

Paste any of these follow-up prompts into the harness chat **immediately after** running a scenario to make the model report exactly what it did:

### Option A: The Fast 1-Liner (Quickest to paste)
```text
Briefly debrief: (1) what exact tools did you call and with what args, (2) what files/folders were created or changed in the workspace, and (3) what is the final outcome or recommendation?
```

### Option B: The Structured Forensic Audit Card (Recommended for small models)
```text
Give me a concise debrief of your actions:
1. Tools Called: List each tool name called and key arguments passed (including aliases).
2. Workspace State: List all files/folders created, modified, or confirmed untouched.
3. Final Outcome: State the recommendation, decision, or artifact generated.
4. Next Step: What did Vivechak's next_step field recommend doing next?
```

### Option C: Casual Human Phrasing
```text
hey give me a quick debrief: what tools did u just run, what files did u touch or create in my workspace, and what's the final verdict?
```

### Option D: Compliance & Quality Gate Audit
```text
Run a compliance check on our research: did you use qualified provenance [Grade A/B], document discovered concerns (>=50 chars), record reversal triggers, and run the adversarial challenge?
```

---

## Scenario 01: Universal Option Comparison (Scope Inference & Bounded Isolation)

### Intent & Engineering Context
The user wants to compare two or more options. The model should call `vivechak_prepare_generator` with `scope: "comparison"`.
Even if the small model omits `scope`, Vivechak's `core.InferScope` automatically detects comparison triggers (`" vs "`, `"compare "`, `"which is faster"`, `"which model"`, `"which algorithm"`, `"which library"`).
`vivechak_prepare_generator` is 100% read-only and will NOT modify the filesystem.
When the comparison matrix is saved via `vivechak_save_plan`, Vivechak bypasses DAG cycle checks because comparisons are standalone Weighted Evaluation Protocol (WEP) matrices.

### 🧪 Test Prompt 1A: Software Library Shootout (Typo & Slang)
```text
vivechek compare winston vs pino for nodejs api.. need very fast logging bcz traffic is high
```

### 🧪 Test Prompt 1B: Zero Mention of Vivechak (Harness Rule Trigger)
```text
hey im confused bw winston and pino for my express app which one shud i use?? need fast json logging pls help
```

### 🧪 Test Prompt 1C: AI Model Comparison (Universal AI Domain)
```text
compare claude 3.5 sonnet vs gpt-4o for complex multi-turn coding agent workflows in our ci pipeline.. tool calling accuracy and context utilization matter most
```

### 🧪 Test Prompt 1D: Algorithm & Data Structure Shootout (Universal Engineering Domain)
```text
which algorithm is faster between b-tree vs lsm-tree for write heavy storage engine with 80% writes?
```

### Expected Agent Behavior
1. **Tool Invocation:** Calls `vivechak_prepare_generator`.
   - `scope`: `"comparison"` (Explicitly passed or automatically inferred).
   - `context`: The user's query and options.
2. **Read-Only Invariant:** Tool returns the generator prompt with injected context. No files or folders are created.
3. **Execution:** Model executes the brief with web search, producing a single comparison session with:
   - Summary & problem statement.
   - WEP Scoring Matrix (1–5 weighted criteria).
   - Failure modes for each candidate.
   - Evidence grades with verified sources (`[Grade A, URL/RFC]`).
4. **Persist (Optional / Minimal):** Calls `vivechak_save_plan` with `scope: "comparison"` and `decision_id: "C-01"` (or auto-recovered).
   - **Zero-Bloat Guarantee:** The server creates only `research/C-01-comparison.md` and copies ONLY `COMPARISON-SESSION.template.md`. It does NOT scaffold the 6-template project pipeline or empty DAG folders.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the model avoid vomiting raw parametric memory without calling the tool?
- [ ] Did it select or infer `comparison` scope instead of defaulting to `project`?
- [ ] Did it avoid creating 6 project templates or multi-session DAG files?
- [ ] Does the comparison matrix include weighted criteria and failure modes?

---

## Scenario 02: Skipped Init Auto-Creation (Resilient Auto-Init & ID Recovery)

### Intent & Engineering Context
Small models frequently skip calling `vivechak_init` and jump straight from generating a plan or comparison matrix to calling `vivechak_save_plan`.
Vivechak automatically initializes the workspace on-the-fly, auto-detects or recovers the decision identifier (`C-01`, `D-01`) from table headers or section titles, copies only scope-appropriate templates, and bypasses DAG cycle constraints for standalone comparison files.

### 🧪 Test Prompt
*(Run this in a brand new empty folder where `vivechak_init` was NOT called yet)*
```text
use vivechak save plan for this comparison between pino and winston:
# Comparison Plan: C-01
| Decision | C-01 |
| Status | complete |

## Summary
Compare Pino and Winston for high-throughput JSON logging.

## WEP Matrix
| Criterion | Weight | Pino | Winston |
|---|---|---|---|
| Throughput | 0.50 | 5 [A, https://github.com/pinojs/pino] | 3 [B, benchmark docs] |
| Ecosystem | 0.50 | 4 [A] | 5 [A] |

## Recommendation
Pino is recommended for high throughput.
```

### Expected Agent Behavior
1. Model calls `vivechak_save_plan` directly with `content: "..."` and `scope: "comparison"` (or omitting `decision_id`).
2. **On-the-Fly Initialization:**
   - Detects missing `.vivechak.json`.
   - Creates `research/`, `research/sessions/`, `research/templates/`.
   - Automatically writes `research/.gitignore` containing `*.lock` (E-39) to protect git repositories from ephemeral advisory lock files.
   - Copies only `COMPARISON-SESSION.template.md` (zero bloat).
   - Auto-extracts `decision_id: "C-01"` from `| Decision | C-01 |`.
   - Writes `research/C-01-comparison.md`.
   - Skips DAG cycle checks for comparison scope.
3. Returns `success: true` with terminal completion guidance in `next_step` ("Single-session comparison complete (no DAG or gate required)") instead of suggesting `vivechak_next_session`.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the tool succeed without returning "workspace not initialized" error?
- [ ] Did it auto-extract `C-01` without failing on missing `decision_id`?
- [ ] Did it copy only the 1 comparison template into `research/templates/`?
- [ ] Did it generate `research/.gitignore` containing `*.lock`?
- [ ] Did `next_step` provide terminal completion guidance rather than demanding `next_session`?

---

## Scenario 03: The "Keep Going" Test (Guided Worker Flow & Draft Sync)

### Intent & Engineering Context
Test whether a small model can autonomously traverse the research pipeline across multiple turns using Vivechak's `next_step` guidance without stalling.
Tests that `ScanCompletedSessions` respects `status: draft` (preventing premature DAG unlocking when upstream sessions have unresolved blocking defects) and resolves slugged filenames.

### 🧪 Test Prompt 3A (Initial Trigger)
```text
vivechak plan: minimal url shortener in golang with redis and postgres
```
*(Wait for model to save the plan with `vivechak_save_plan`)*

### 🧪 Test Prompt 3B (Follow-up Vague Prompt)
```text
ok do next step pls
```
*(Or simply:)*
```text
continue
```

### Expected Agent Behavior
1. Model inspects previous tool's `next_step`: `"Run vivechak_next_session to get the first research session prompt."`
2. Model calls `vivechak_next_session()`.
3. Vivechak resolves the DAG, selects the next unblocked node (e.g. `T0-01`), injects upstream context, and returns the prompt brief.
4. Model executes the brief with web search and calls `vivechak_save_session(session_id="T0-01", content=...)`.
5. When valid, `tool_save_session` marks `status: complete` and returns the next unblocked session in `next_step`.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the model call `vivechak_next_session` without needing manual explanation?
- [ ] Did it preserve the exact session ID (`T0-01`)?
- [ ] Did it avoid asking the user "What should I do next?"

---

## Scenario 04: Fast Architectural Decision (Two-Way Door Fast Spike)

### Intent & Engineering Context
The user asks an architectural question that is easily reversible (can be swapped later without database migration or enterprise contracts).
The model should invoke `vivechak_prepare_generator(scope="decision")`. The decision generator profiles door type: Reversal Cost $R=1$, Blast Radius $B=1$, Team Familiarity $N=1$, routing to **Fast Spike (1 session)** or **Confirm & Commit** rather than an unnecessary multi-session enterprise pipeline.

### 🧪 Test Prompt 4A: Frontend Framework Dilemma (Slang & Misspelling)
```text
bro shuld i use react 19 server actions or tanstack react query for fetching data in my nextjs dashboard? im totally confused
```

### 🧪 Test Prompt 4B: Caching Layer Spike (Direct Trigger with Constraints)
```text
use vivechak decide in-memory caching vs redis for dev spike... team is 2 junior devs and app is small saas
```

### Expected Agent Behavior
1. **Tool Invocation:** Calls `vivechak_prepare_generator` with `scope: "decision"`.
   - `scope`: `"decision"`
   - `context`: Injects user question, constraints ("2 junior devs", "small saas"), and options.
2. **Execution:** Produces `[ID]-PLAN.md` with:
   - Door Profile: **Two-Way Door**.
   - Lane: **Fast Spike** (1 session) or **Confirm & Commit**.
   - Pre-allocated proposed ADR `D-01-data-fetching.md`.
3. **Persist:** Saves plan via `vivechak_save_plan(scope="decision")`.
4. Returns concise next step for the single spike session.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the model correctly categorize this as a **Two-Way Door**?
- [ ] Did it create a lightweight 1-session plan rather than a bloated 6-session pipeline?
- [ ] Did it preserve team constraints in the decision context?

---

## Scenario 05: High-Stakes One-Way Door (Regulatory $X=3$ / Strategic Deep Research)

### Intent & Engineering Context
User asks an innocent-sounding database or business question that involves high reversibility cost ($R=3$) or regulatory compliance (HIPAA/PHI, enterprise contracts).
Tests that `core.InferScope` prioritizes explicit decision markers (e.g. `strategic decision:`, `business decision:`, `adr:`) over generic keywords, profiles the choice as a **One-Way Door ($R=3$)**, routes to **Deep Research (2–3 sessions: Landscape L, Comparison C, Falsification F)**, and mandates Phase 0 Gate **Track B** (B1–B11 checks + adversarial red-team challenge).

### 🧪 Test Prompt 5A: Healthcare Database Selection (Regulatory $X=3$)
```text
vivechak decide: mongo vs postgres for our new clinic app. we gonna store patient medical records, lab reports and billings. we need to start asap
```

### 🧪 Test Prompt 5B: Strategic Business Expansion (Universal Strategic Domain)
```text
strategic decision: direct enterprise sales vs partner reseller for our European market expansion. enterprise contracts are 7-figure ACV and GDPR compliance is mandatory
```

### Expected Agent Behavior
1. **Tool Invocation:** Calls `vivechak_prepare_generator` with `scope: "decision"`.
2. **Door Profiling:**
   - Detects `patient medical records` (Regulatory $X=3$, HIPAA/PHI) OR 7-figure enterprise contracts / GDPR ($R=3$).
   - Categorizes as **One-Way Door**.
   - Routes to **Deep Research Lane** (2–3 sessions: Landscape `L`, Comparison `C`, Falsification `F`).
   - Targets **Gate Track B** (rigorous checks, red-team challenge required).
3. **Persist:** Calls `vivechak_save_plan` saving `research/D-01-plan.md` and draft ADR in `research/DECISIONS.md`.
4. Auto-recovers decision ID `D-01` from plan routing tables.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the model flag the decision as a **One-Way Door**?
- [ ] Did it schedule a dedicated Falsification session (`F`) probing failure modes/compliance?
- [ ] Did it state that Phase 0 Gate Track B (with adversarial challenge) will be required?

---

## Scenario 06: Full Venture / Project Pipeline (Multi-Tier DAG → FAD)

### Intent & Engineering Context
User presents a broad product idea or comprehensive new system.
Calls `vivechak_prepare_generator(scope="project")`.
Tests that Vivechak's DAG parser safely disentangles 2-column key-value metadata tables (`| Field | Value |`) from 4-column horizontal session tables (`| ID | Layer | Door Type | Dependencies |`) and guards against rubric tables, successfully producing `RESEARCH-PIPELINE.md` and initial `DECISIONS.md`.

### 🧪 Test Prompt 6A: Startup Product Pitch ("Uber for Dogs")
```text
hey vivechak plan my new startup idea. its basically uber but for dog walkers. need mobile app, real time tracking gps, stripe payments and push notifications. what tech stack should we build with?
```

### 🧪 Test Prompt 6B: Clean Energy Microgrid Venture (Universal Systems Domain)
```text
new venture: decentralized renewable energy peer-to-peer microgrid trading marketplace. requires smart meter cryptographic verification, real-time matching engine, and regulatory grid settlement protocols
```

### Expected Agent Behavior
1. **Tool Invocation:** Calls `vivechak_prepare_generator` with `scope: "project"`.
2. **Pipeline Generation:** Produces `RESEARCH-PIPELINE.md` containing:
   - Layer 0: Foundational choices (Protocols / Tech stack).
   - Layer 1: Core engines (Matching / Real-time dispatch).
   - Layer 2: Supporting services (Payments, Auth, Notifications).
   - Terminal Synthesis: `SYN-01` (`research/FAD.md`).
3. **Persist:** Calls `vivechak_save_plan` writing `research/RESEARCH-PIPELINE.md` and `research/DECISIONS.md`.
4. Returns total session count and points to `T0-01` via `next_step`.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the model output clean session IDs (`T0-01`, `T1-01`) without illegal prefixes?
- [ ] Did `vivechak_save_plan` pass validation with 0 blocking L2 errors?
- [ ] Does the plan include the terminal synthesis session (`SYN-01` → `FAD.md`)?

---

## Scenario 07: Quality Coaching & 50-Char Concern Gate (`core.ObserveSessionQuality`)

### Intent & Engineering Context
Small models frequently generate ungrounded, hyper-optimistic claims without URLs or citations, and forget to document failure modes.
When saving a session with `vivechak_save_session`, Vivechak runs `core.ObserveSessionQuality` and injects non-blocking `Q-*` coaching observations directly into `next_step`:
- `Q-GRADE-INFLATION`: Grade A claims lacking qualified provenance (canonical URL, RFC, or domain anchor).
- `Q-MISSING-CONCERNS`: Missing Discovered Concerns section or content under 50 characters (`V-OWD-NO-CONCERNS`). 49 characters is blocked in `status: draft`; $\ge 50$ characters is promoted to `status: complete`.
- `Q-DELTA-CONFIRMATION-BIAS`: Discovered concerns identical to or confirming preferred option.
- `Q-LOW-FINDINGS`: Fewer than 3 substantive findings.

### 🧪 Test Prompt (Feeding Sloppy Session Output)
```text
save this session output for T0-01:
---
id: T0-01
title: Framework Evaluation
status: complete
---
# Session T0-01: Framework Selection
We decided to use Gin framework. Gin is the fastest framework ever made [Grade A].
It handles 10 million requests per second with zero memory [Grade A].
Everyone uses Gin. No downsides found.
```

### Expected Agent Behavior
1. Model calls `vivechak_save_session(session_id="T0-01", content=...)`.
2. Vivechak evaluates the content:
   - Flags unqualified `[Grade A]` without URL/RFC (`Q-GRADE-INFLATION`).
   - Flags missing Discovered Concerns (`Q-MISSING-CONCERNS`).
   - Flags low findings count (`Q-LOW-FINDINGS`).
   - Demotes frontmatter to `status: draft` due to blocking defects.
3. Response returns `success: true` with actionable `Q-*` coaching list in `next_step`.
4. The agent reads the coaching notes and amends the session with qualified citations and $\ge 50$ chars of discovered concerns.

### 🔍 Forensic Evaluation Checklist
- [ ] Did Vivechak flag the ungrounded Grade A claim in `next_step`?
- [ ] Did it flag missing Discovered Concerns?
- [ ] Was the session kept in `draft` status until repaired?

---

## Scenario 08: Adversarial Red-Team Challenge (Two-Tier Lookup & Resilient Aliases)

### Intent & Engineering Context
Test `vivechak_challenge` to run an adversarial red-team probe.
Verifies parameter aliasing: accepts `decision_id` (alias for `session_id`) and `challenge_type="premortem"` (alias for `mode="red_team"`).
**Key Hardening Feature:** Two-Tier Lookup! The tool searches `research/sessions/` first, and if not found, searches `research/` directly for standalone ADR files (`research/D-01.md`, `research/D-01-database.md`) and comparison matrices (`research/C-01.md`, `research/C-01-comparison.md`).

### 🧪 Test Prompt 8A: Roast a Tech Decision (Slang / Casual)
```text
bro can u roast my decision? we chose microservices with kafka for our 3-person team mvp. use vivechak challenge
```

### 🧪 Test Prompt 8B: Challenge a Standalone ADR directly in `research/`
```text
run an adversarial red-team challenge on decision D-01 using vivechak challenge with challenge_type premortem
```

### 🧪 Test Prompt 8C: Challenge a Comparison Matrix directly in `research/`
```text
challenge comparison C-01: probe edge cases and vendor lock-in risks
```

### Expected Agent Behavior
1. Model calls `vivechak_challenge`:
   - `decision_id`: `"D-01"` (or `session_id: "D-01"`)
   - `challenge_type`: `"premortem"` (or `mode: "red_team"`)
2. Two-tier resolution finds the target file in `research/` or `research/sessions/`.
3. Returns structured data:
   - `recommendation`: Extracted from artifact.
   - `challenge_prompts`: Array of targeted adversarial prompts probing 12-month post-mortem failure, operational overhead, and migration traps.
   - `context`: Evidentiary context (door type, evidence grades, rejected alternatives).
4. The model executes the challenge prompt and reports the falsification critique.

### 🔍 Forensic Evaluation Checklist
- [ ] Did `vivechak_challenge` accept `decision_id` and `challenge_type` aliases without error?
- [ ] Did it successfully resolve standalone files directly in `research/`?
- [ ] Did the model avoid sycophantically agreeing with the user ("Yes, Kafka is great!")?

---

## Scenario 09: Incremental Session & ADR Amendment (`vivechak_amend_session`)

### Intent & Engineering Context
Test appending post-hoc findings or challenge results without overwriting existing files.
Verifies parameter aliases: `content` (alias for `amendment`) and `reason` (prepends `Reason: <reason>\n\n<content>`).
**Key Hardening Feature:** Two-Tier Lookup resolves target session IDs across both `research/sessions/` and `research/` (for standalone ADRs `D-*.md` and comparison matrices `C-*.md`). For synthesis sessions (`SYN-01`, `FAD`), the amendment atomically updates both `research/FAD.md` and the root `FOUNDING-ARCHITECTURE.md` mirror.

### 🧪 Test Prompt 9A: Amend a Session in `research/sessions/`
```text
hey vivechak update session T0-01: add this benchmark finding to detailed findings - pino handles 14000 req/sec while winston drops to 4500 req/sec under heavy json serialization [Grade A, https://github.com/pinojs/benchmarks]
```

### 🧪 Test Prompt 9B: Amend a Standalone ADR in `research/`
```text
amend decision D-01 with reason "Clarified enterprise EMEA pod structure": Add dedicated EMEA pod organization structure for localized customer onboarding.
```

### 🧪 Test Prompt 9C: Amend a Comparison Matrix in `research/`
```text
update comparison C-01 with updated pricing tier quotas and token rate limits
```

### Expected Agent Behavior
1. Model calls `vivechak_amend_session`:
   - `session_id`: `"T0-01"` (or `"D-01"`, `"C-01"`)
   - `content` or `amendment`: The finding text.
   - `reason`: (Optional) rationale summary.
2. Two-tier lookup resolves target file in `research/sessions/` or `research/`.
3. Appends formatted `## Post-Hoc Amendment` block with UTC timestamp under an advisory lock.
4. **Registry Sync (E-41):** If the amended target is a standalone ADR (`D-*.md`), Vivechak automatically recompiles `research/DECISIONS.md` so the consolidated decision registry stays strictly synchronized with the amended file.
5. Returns `success: true` with confirmed file path.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the tool append the amendment without wiping existing file content?
- [ ] Did it accept `content` and `reason` aliases?
- [ ] Did it successfully resolve standalone ADRs and comparison files directly in `research/`?
- [ ] When amending a standalone ADR (`D-01.md`), did it automatically recompile `research/DECISIONS.md`?

---

## Scenario 10: Formal ADR Recording & Single-Source Consolidation (`vivechak_record_decision`)

### Intent & Engineering Context
Test recording an architectural decision into a canonical standalone ADR file (`research/D-01-database.md`) and compiling the consolidated `research/DECISIONS.md` registry.
Verifies that all required fields for Phase 0 Gate Track B are properly populated: `door_type: one-way`, `review_trigger`, `rejected_alts`, `informed_by_sessions`, `evidence_refs`.
Supports `auto_draft_from` to draft ADR bodies directly from completed sessions.

### 🧪 Test Prompt 10A: Technical Database ADR
```text
vivechak record decision D-01: Title is "Primary Database Selection". We chose PostgreSQL. Door is one-way door. Reversal trigger is write throughput exceeding 50k sustained queries per sec. Rejected option was MongoDB because of lack of relational joins and ACID compliance. Informed by session T0-01. Evidence: https://www.postgresql.org/docs/current/
```

### 🧪 Test Prompt 10B: Strategic Enterprise Sales ADR (Universal Business Domain)
```text
record decision D-01: Title is "Direct Enterprise Sales Model". Status is accepted. Door is one-way door. Reversal trigger is enterprise CAC exceeding 18 months payback period. Rejected partner reseller due to channel conflict. Informed by session D-01-S1.
```

### Expected Agent Behavior
1. Model calls `vivechak_record_decision`:
   - `decision_id`: `"D-01"`
   - `title`: `"Primary Database Selection"`
   - `door_type`: `"one-way"`
   - `status`: `"accepted"`
   - `review_trigger`: `"Write throughput exceeding 50k sustained queries per sec"`
   - `content`: Markdown body with Decision Outcome, Rejected Alternatives, and Reversal Triggers.
2. Server writes canonical `research/D-01-database.md` (or slugged filename) with YAML frontmatter.
3. Automatically updates and compiles `research/DECISIONS.md` registry table.
4. Returns `success: true` with `next_step` recommending next session or exit gate.

### 🔍 Forensic Evaluation Checklist
- [ ] Did `research/DECISIONS.md` update with the new decision entry?
- [ ] Does the ADR frontmatter contain `door_type: one-way`, `status: accepted`, and `review_trigger`?
- [ ] Does the document record explicit rejected alternatives?

---

## Scenario 11: The Scope Pivot (Dynamic DAG Re-planning)

### Intent & Engineering Context
Midway through research, the user changes a fundamental constraint.
The model must use `vivechak_replan` to adapt the DAG without corrupting completed, independent sessions.
Vivechak identifies downstream affected nodes, flags invalid dependencies, and updates the plan.

### 🧪 Test Prompt
```text
wait wait stop everything! boss said no postgres allowed, company has strict contract with oracle only! can vivechak update our plan?
```

### Expected Agent Behavior
1. Model calls `vivechak_replan`:
   - Either directly with `operation`: `"update_prompt"`, `session_id`: target session (e.g. `T0-01`), `new_prompt`: `"Evaluate Oracle DB..."`, `reason`: `"Company restriction: PostgreSQL disallowed, Oracle database mandatory"`, and `impact`: `"Updates database evaluation and ORM selection sessions"`.
   - Or in guided mode: omitting `operation` returns available operations (`["add_session", "remove_session", "update_deps", "update_prompt"]`), candidate uncompleted sessions, and guidance via `next_step` directing the agent to specify the operation and target session.
2. Vivechak validates acyclicity, updates the DAG, and preserves unrelated completed sessions (e.g. auth, frontend).
3. Informs the user which specific sessions were invalidated or scheduled for re-execution.

### 🔍 Forensic Evaluation Checklist
- [ ] Did Vivechak preserve completed sessions unrelated to database storage?
- [ ] Did the model report which specific sessions were affected?
- [ ] Was the pipeline document updated without corruption?

---

## Scenario 12: Mechanical Phase 0 Exit Gate (`vivechak_run_gate` & Track Filtering)

### Intent & Engineering Context
Test `vivechak_run_gate` when the user asks to start writing code.
The gate decouples validation into two distinct tracks:
- **Track A (Two-Way Doors):** Fast verification of reversibility and review triggers.
- **Track B (One-Way Doors):** Rigorous mechanical evaluation of 11 checks (B1–B11):
  - B1: Status accepted
  - B2: Valid door type
  - B3: Grade A/B evidence threshold
  - B4: Verification method present (fetched or cached; recalled capped at Grade D)
  - B5: Rejected alternatives documented
  - B6: Explicit reversal trigger (metric/event condition)
  - B7: Schema version valid
  - B8: Confidence level present
  - B9: Informed by sessions referenced
  - B10: Substantive content ($\ge 50$ chars)
  - B11: Adversarial red-team challenge performed
Supports `track` parameter: `"track_a"`, `"track_b"`, or `"all"`.
- **Active Track Filtering (E-40):** When filtering by `track="track_a"`, failing Track B criteria do not pollute the output or block Track A prototyping; blocking issues are scoped strictly to the selected track.
- **Decoupled Clearance (E-42):** Track A explicitly passes and authorizes prototyping when all Two-Way Doors satisfy reversibility, even if Track B (One-Way Doors) is still pending or incomplete.
Returns structured JSON fields: `track_filter: "track_a"|"track_b"|"all"`, `track_a_result: "PASS"|"FAIL"`, `track_b_result: "PASS"|"FAIL"`, and a `decisions` array. Automatically persists `research/PHASE-0-GATE.md`.

### 🧪 Test Prompt 12A: General Exit Gate Check
```text
yo can i start coding the backend now?? run the gate thing in vivechak
```

### 🧪 Test Prompt 12B: Two-Way Door Fast Prototyping Filter
```text
can i start prototyping my two-way door frontend parts now? run vivechak gate for track a only
```

### 🧪 Test Prompt 12C: High-Stakes One-Way Door Filter
```text
check if our one-way architectural decisions pass the exit gate. run vivechak gate for track b
```

### 🧪 Test Prompt 12D: Headless Terminal / CI Evaluation (`vck gate`)
```bash
# In terminal, CI pipeline, or shell script:
vck gate --track all
vck gate --track track_a
vck gate --track track_b --format json
```

### Expected Agent Behavior
1. Model calls `vivechak_run_gate(track="all")` (or `track="track_a"` / `"track_b"`), or user runs `vck gate` from terminal or CI.
2. Vivechak mechanically inspects all ADRs, sessions, and FAD against decoupled Track A and Track B criteria.
3. For `track="track_a"`, if all Two-Way Doors pass, it returns `track_a_result: "PASS"` and issues no blocking failures from Track B, giving clearance for prototyping.
4. If evaluated checks fail, returns binary punch list of blockers scoped to the active track (e.g. `[FAIL] B11: D-01 has no red-team challenge`).
5. Persists `research/PHASE-0-GATE.md` with rendered gate status.
6. Model reports exact pass/fail status per track to the user.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the tool prevent premature coding when critical checks failed?
- [ ] Did it support the `track` filter when asked for Track A or Track B, returning `track_filter` in JSON?
- [ ] For `track="track_a"`, did Two-Way Doors pass without Track B failures polluting the result?
- [ ] Was the output an objective, binary PASS/FAIL checklist rather than vague AI prose?

---

## Scenario 13: Quick Status & Visual Graph (`vivechak_status` & `vivechak_visualize`)

### Intent & Engineering Context
User wants a quick pulse check on where the research stands.
Calls `vivechak_status` to get completed/pending session counts and gate readiness.
- **Scope-Aware Accounting (E-42):** `InspectWorkspace` discovers and counts standalone comparison matrices (`research/C-*.md`) as completed sessions even when `research/sessions/` is empty.
- **Terminal Completion Guidance (E-38):** When all sessions are finished (or for Comparison Scope), `vivechak_status` emits a terminal completion `next_step` instead of prompting for `vivechak_next_session`.
Calls `vivechak_visualize` to generate an ASCII tree or Mermaid diagram representing the research DAG.

### 🧪 Test Prompt
```text
vivechak status? show me visual graph if u can
```

### Expected Agent Behavior
1. Model calls `vivechak_status()`:
   - Returns detected scope (`"project"`, `"decision"`, or `"comparison"`).
   - Accurately counts completed sessions (including `research/C-*.md` in comparison scope).
   - If pending sessions exist, identifies next session; if all complete, provides synthesis/gate/terminal completion `next_step`.
2. Model calls `vivechak_visualize(format="mermaid")` (or `format="table"`).
3. Model displays the status summary and renders the diagram directly in chat.

### 🔍 Forensic Evaluation Checklist
- [ ] Did it clearly display completed vs pending sessions?
- [ ] For comparison scope, did it count standalone `C-*.md` in completed sessions?
- [ ] When all sessions are done (or in comparison scope), did `next_step` avoid prompting for `next_session`?
- [ ] Did it display a valid Mermaid or Table diagram?
- [ ] Did it state the exact next actionable step or exit gate readiness?

---

## Scenario 14: Multi-Tier Workspace Audit (`vivechak_validate`)

### Intent & Engineering Context
Test running a full dry-run validation across the entire workspace.
When `vivechak_validate` is called with no content (workspace mode), it automatically discovers and validates all artifacts across all scope tiers:
- Research sessions in `research/sessions/*.md`
- Decision plans (`research/*-plan.md`) and project pipelines (`research/RESEARCH-PIPELINE.md`)
- Standalone ADRs (`research/D-*.md`)
- Consolidated decision registry (`research/DECISIONS.md`)
- Standalone comparison matrices (`research/C-*-comparison.md`)
- Founding Architecture Document (`research/FAD.md`)
Deduplicates IDs across standalone files and registry using regex prefix matching `^(?i)(D-[A-Za-z0-9]+)`.

### 🧪 Test Prompt
```text
run vivechak validate on the whole workspace to make sure all our research files, adrs, plans, and comparisons are valid
```

### Expected Agent Behavior
1. Model calls `vivechak_validate()`.
2. Vivechak scans the workspace, discovering:
   - `pipeline`: Validates DAG structure, cycles, dangling references.
   - `sessions`: Validates frontmatter, evidence grades, discovered concerns.
   - `decisions`: Validates ADR schema, door types, reversal triggers.
   - `comparison`: Validates standalone comparison matrix format.
   - `fad`: Validates synthesis document structure (if present).
3. Returns structured summary with remediation hints if any warnings or blocking defects exist.

### 🔍 Forensic Evaluation Checklist
- [ ] Did `vivechak_validate` run across the workspace without requiring manual file paths?
- [ ] Did it discover standalone ADRs, comparison matrices, and decision plans?
- [ ] Did it avoid double-counting decisions present in both `D-*.md` and `DECISIONS.md`?

---

## Scenario 15: AST Code-Fence Parsing & Section Truncation Immunity (`E-28`)

### Intent & Engineering Context
Small models and developers frequently paste bash, Python, or Dockerfile snippets inside research session sections (such as `## Discovered Concerns` or `## Recommendation`).
Before E-28, lines starting with `#` inside code blocks were parsed as Markdown H1 headings, immediately aborting section extraction and causing `len(concerns) < 50`, which triggered a fatal `V-OWD-NO-CONCERNS` error.
With CommonMark `FenceTracker`, Vivechak tracks arbitrary length fences (```` ``` ```` and `~~~`) and ignores comments inside code blocks.

### 🧪 Test Prompt (Code Block Inside Discovered Concerns)
```text
save this session output for T0-01:
---
session_id: T0-01
title: Kernel Buffer Tuning
door_type: one-way
status: draft
date: "2026-10-18"
---
# T0-01: Kernel Buffer Tuning
## Key Findings
- FoundationDB achieves sub-5ms commit latency with Redwood engine [Grade A: https://apple.github.io/foundationdb/].

## Discovered Concerns
Here are the primary production operating risks and kernel configuration commands:
```bash
# Check current system file descriptor limits
cat /proc/sys/fs/file-max
# Verify socket buffer limits
sysctl net.core.rmem_max
# Ensure no silent truncation occurs here
ulimit -n 65536
```
These buffer configuration parameters must be carefully benchmarked under load.

## Recommendation
Adopt FoundationDB with Redwood engine [Grade A: https://apple.github.io/foundationdb/].
```

### Expected Agent Behavior
1. Model calls `vivechak_save_session(session_id="T0-01", content=...)`.
2. `ExtractSection` uses `FenceTracker` to safely retain the code block and trailing text.
3. Because the extracted text exceeds 50 characters, `V-OWD-NO-CONCERNS` is **not triggered**.
4. Frontmatter is promoted from `status: draft` to `status: complete`.

### 🔍 Forensic Evaluation Checklist
- [ ] Did `ExtractSection` avoid truncating at `# Check current system...`?
- [ ] Was the session promoted to `status: complete` on disk?
- [ ] Did `next_step` confirm successful persistence without blocking concerns defects?

---

## Scenario 16: CLI Shorthand & Scope Auto-Inference Triggers (`E-34`)

### Intent & Engineering Context
Users frequently invoke Vivechak using natural command prefixes with the official shorthand `vck` or `vivechak` (e.g. `vck plan`, `vck decide`, `vck compare`, `vck vs`).
`core.InferScope` explicitly detects these prefixes with highest priority, bypassing ambiguous keyword collisions.

### 🧪 Test Prompt 16A: Shorthand Compare
```text
vck compare winston vs pino for nodejs api logging
```

### 🧪 Test Prompt 16B: Shorthand Decide
```text
vck decide: postgres vs mongodb for clinic records
```

### 🧪 Test Prompt 16C: Shorthand Plan
```text
vck plan new b2b payment platform with stripe and plaid
```

### Expected Agent Behavior
1. For Prompt 16A: Infers `ScopeComparison` (`comparison`) and calls `vivechak_prepare_generator(scope="comparison")`.
2. For Prompt 16B: Infers `ScopeDecision` (`decision`) and calls `vivechak_prepare_generator(scope="decision")`.
3. For Prompt 16C: Infers `ScopeProject` (`project`) and calls `vivechak_prepare_generator(scope="project")`.

### 🔍 Forensic Evaluation Checklist
- [ ] Did `vck compare` trigger `comparison` scope immediately?
- [ ] Did `vck decide` trigger `decision` scope immediately?
- [ ] Did `vck plan` trigger `project` scope immediately?

---

## Scenario 17: Dual-Source Decision Resolution & Deduplication (`E-29`)

### Intent & Engineering Context
In hybrid workflows, some ADRs are recorded directly as standalone files (`research/D-01.md`), while others are appended to the consolidated registry (`research/DECISIONS.md`).
Before E-29, the presence of any standalone ADR caused `CollectGateDecisions` and `verifyDecisionsMechanical` to ignore `research/DECISIONS.md`.
With dual-source resolution, both sources are scanned simultaneously with case-insensitive deduplication (`strings.ToUpper(id)`), ensuring zero split-brain behavior.

### 🧪 Test Prompt
```text
run vivechak gate on our project where decision D-01 is in research/D-01.md and decision D-02 is in research/DECISIONS.md
```

### Expected Agent Behavior
1. Model calls `vivechak_run_gate()`.
2. Vivechak scans both `research/D-01.md` and `research/DECISIONS.md`.
3. Discovers both decisions (`D-01` and `D-02`), evaluating Track A for Two-Way Doors and Track B for One-Way Doors.
4. Returns evaluation results for both decisions without omitting either source.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the gate evaluate decisions from both files?
- [ ] Did it deduplicate decisions present in both sources?
- [ ] Was `research/PHASE-0-GATE.md` updated with the combined decisions list?

---

## Scenario 18: Quality Coaching Provenance Cross-Referencing & Colon-Delimited Syntax (`E-32`, `E-35`, `E-37`)

### Intent & Engineering Context
Research sessions frequently separate finding bullets from formal citation tables. An agent might state a claim `[CLAIM-01]` in bullet points and provide the canonical URL in the `## Sources & Evidence Ledger` table (with headings like `## Sources & Evidence Ledger`, `## 5. Sources & Evidence Ledger`, or `### References`).
- **Table Anchor Provenance:** `quality.go` cross-references table anchors against Grade A claims, excludes table rows (`|`) from claim line counts, and uses `sourcesHeadingRegex` to match numbered headings, avoiding false `Q-INFLATION` warnings.
- **Colon & Parenthetical Syntax Support (E-37):** `extractEvidenceGrades` seamlessly parses both colon-delimited `[Grade: A]` / `(Grade: B)` and space-delimited `[Grade A]` notation.

### 🧪 Test Prompt (Grade A Anchored in Ledger Table & Colon Syntax)
```text
save this session output for T0-01:
---
session_id: T0-01
title: Postgres Storage Engine
door_type: two-way
status: draft
date: "2026-10-18"
---
# T0-01: Postgres Storage Engine
## Key Findings
- [CLAIM-01] PostgreSQL 16 introduces btree indexing speedups [Grade: A | corroborated | fetched].
- [CLAIM-02] Sequential scan memory overhead reduced by 30% [Grade: B | fresh | test].

## Discovered Concerns
Vacuum freeze overhead on multi-terabyte tables requires scheduled maintenance windows.

## 5. Sources & Evidence Ledger
| # | Claim | Grade | Modifiers | Status | Source |
|---|---|---|---|---|---|
| 1 | CLAIM-01 | Grade A | corroborated | fetched | https://www.postgresql.org/docs/16/release-16.html |
| 2 | CLAIM-02 | Grade B | fresh | test | Internal hardware benchmark |

## Recommendation
Adopt PostgreSQL 16 [Grade A: https://www.postgresql.org/docs/16/release-16.html].
```

### Expected Agent Behavior
1. Model calls `vivechak_save_session(session_id="T0-01", content=...)`.
2. Vivechak parses the findings and ledger table:
   - Accurately recognizes `[Grade: A]` and `[Grade: B]` via colon-aware evidence grade extraction (`E-37`).
   - Finds URL `https://www.postgresql.org/docs/16/release-16.html` in the ledger.
   - Matches `## 5. Sources & Evidence Ledger` via `sourcesHeadingRegex`.
   - Excludes table rows starting with `|` from claim line counts.
   - Detects that Grade A claim has a corresponding table anchor.
3. No false `Q-INFLATION` warning is raised. Frontmatter is promoted to `status: complete`.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the quality engine accept the canonical URL inside the ledger table?
- [ ] Did it correctly extract colon-delimited evidence grades (`[Grade: A]`, `[Grade: B]`)?
- [ ] Did it recognize the numbered heading `## 5. Sources & Evidence Ledger`?
- [ ] Was the session promoted to `status: complete` on disk?

---

## Scenario 19: Universal Host Auto-Configuration Across 21 AI Harnesses (`vck setup` & `vck mcp-config`)

### Intent & Engineering Context
Users and agents configure their IDE or CLI agent using Vivechak's setup subcommands.
`vck setup` supports 21 distinct presets (`cursor`, `vscode`, `claude-code`, `mcp`, `opencode`, `continue`, `goose`, `copilot`, `void`, `claude-desktop`, `windsurf`, `antigravity`, `kiro`, `trae`, `omp`, `openhands`, `droid`, `cline`, `roo`, `devin`, `zed`).
Uses number-preserving JSON decoding (`dec.UseNumber()`) to guarantee zero data loss in existing host configuration files (such as `~/.claude.json`).

### 🧪 Test Commands
```bash
# 1. Auto-configure specific host harnesses:
vck setup cursor
vck setup claude-code
vck setup continue
vck setup mcp

# 2. Output universal pipeable MCP configuration:
vck mcp-config

# 3. Dry-run preview:
vck setup vscode --dry-run
```

### Expected Behavior
1. `vck setup cursor` creates or merges `.cursor/mcp.json` or `~/.cursor/mcp.json` with the `mcpServers` key.
2. `vck setup claude-code` merges with number preservation into `~/.claude.json` or `.mcp.json`.
3. `vck mcp-config` outputs valid JSON containing the absolute binary path and `["serve"]` arguments.

### 🔍 Forensic Evaluation Checklist
- [ ] Did `vck setup` merge without wiping other configured servers?
- [ ] Did `vck mcp-config` output valid JSON to stdout?
- [ ] Were existing token strings and numeric IDs preserved without truncation?

---

## Scenario 20: Pre-Coding Workspace Integrity Check (`vck doctor`)

### Intent & Engineering Context
Before running the Phase 0 Gate or starting codebase development, users can run `vck doctor` to audit workspace integrity.
Checks:
- Workspace folder existence and layout.
- Required templates for the active scope (6 for project, 2 for decision, 1 for comparison).
- YAML frontmatter completeness across all session and decision files.
- Orphaned or unregistered sessions in `research/sessions/`.
- Stale ADR references (`D-OLD`) after supersession.
- In-flight decisions planned in `RESEARCH-PIPELINE.md` that lack locked ADRs.
- Active workspace scope detection (`✓ Scope: <project|decision|comparison>`).

### 🧪 Test Command
```bash
vck doctor
# Or from outside the workspace:
vck doctor /path/to/project
```

### Expected Behavior
1. Prints clear diagnostic checklist:
   ```text
   ✓ Workspace exists
   ✓ Scope: project (or decision / comparison)
   ✓ 6/6 required templates present
   ✓ Frontmatter valid across all files
   ✓ No orphaned sessions found
   ✓ All planned decisions locked
   ```
2. Returns exit code `0` if healthy, non-zero if blocking defects are found.

### 🔍 Forensic Evaluation Checklist
- [ ] Did `vck doctor` report the detected research scope?
- [ ] Did it flag any missing required templates?
- [ ] Did it check for orphaned sessions and unaccepted decisions?

---

## Scenario 21: Out-of-Scope Anti-Pattern Boundary (Direct Code Authoring & Bug Fixing)

### Intent & Engineering Context
Vivechak is an epistemic grounding framework for architectural, comparative, and strategic decisions. It must **never** be invoked for direct code authoring, local bug fixing, compiler error diagnosis, micro-refactoring, or syntax lookups.
Per AGENTS.md §2.4, README.md, and `ServerInstructions`, models must skip Vivechak research tools completely when the user's intent is routine software engineering.

### 🧪 Test Prompt 21A: Direct Function Implementation
```text
Write a python function to compute HMAC-SHA256 signature of a payload given a secret key.
```

### 🧪 Test Prompt 21B: Local Compiler / Type Error
```text
Fix this typescript type error in my component: Type 'null' is not assignable to type 'string'.
```

### 🧪 Test Prompt 21C: Micro-Refactoring / Variable Renaming
```text
Rename variable user_id to account_id in src/auth.ts and update references.
```

### Expected Agent Behavior
1. **Tool Invocation:** **Zero tool calls.** The agent must NOT call `vivechak_init`, `vivechak_prepare_generator`, or any other Vivechak tool.
2. **Direct Execution:** The agent immediately writes the requested Python function, diagnoses/fixes the TypeScript error, or performs the refactoring directly.

### 🔍 Forensic Evaluation Checklist
- [ ] Did the model skip calling `vivechak_prepare_generator` or other research tools?
- [ ] Did it directly answer the user's code, syntax, or bug-fix request?
- [ ] Was the answer delivered without creating any `research/` directories or templates?

---

## Summary Matrix for AI Agent Testing

| Test | User Prompt Snippet | Expected First Tool Call | Success Indicator |
|---|---|---|---|
| **#1** | `vivechek compare winston vs pino...` | `vivechak_prepare_generator(scope="comparison")` | Scope auto-inferred, read-only brief, generates WEP matrix |
| **#2** | `save plan for comparison c-01...` | `vivechak_save_plan(scope="comparison")` | Auto-initializes on-the-fly, auto-recovers ID, writes `research/.gitignore` (`*.lock`), terminal completion `next_step` |
| **#3** | `ok do next step pls` | `vivechak_next_session()` | Seamless sequential execution via `next_step`, respects draft status |
| **#4** | `use vivechak decide react 19 vs react query...` | `vivechak_prepare_generator(scope="decision")` | Profiles Two-Way Door, produces 1-session Fast Spike ADR |
| **#5** | `vivechak decide: mongo vs postgres clinic app...` | `vivechak_prepare_generator(scope="decision")` | Profiles One-Way Door ($X=3$), routes Deep Research (L, C, F) |
| **#6** | `hey vivechak plan my startup... uber for dogs` | `vivechak_prepare_generator(scope="project")` | Disentangles tables, outputs multi-tier DAG in `RESEARCH-PIPELINE.md` |
| **#7** | `save this session: Gin is fastest [Grade A]...` | `vivechak_save_session()` | Triggers `Q-*` coaching on ungrounded claims, sets `status: draft` |
| **#8** | `bro roast my decision... use vivechak challenge` | `vivechak_challenge(decision_id="D-01")` | Accepts `decision_id` & `premortem` aliases, two-tier lookup resolves ADR |
| **#9** | `update session T0-01 with benchmark finding...` | `vivechak_amend_session(session_id="T0-01")` | Accepts `content` & `reason` aliases, two-tier lookup patches file, auto-recompiles `DECISIONS.md` on ADR amendments |
| **#10**| `record decision D-01: Title is Primary DB...` | `vivechak_record_decision(decision_id="D-01")` | Compiles canonical ADR to `research/D-01-*.md` and updates `DECISIONS.md` |
| **#11**| `wait stop! boss said no postgres... update plan` | `vivechak_replan(reason="...")` | Guided overview mode or modifies affected nodes with `reason` and `impact` metadata |
| **#12**| `yo can i start coding now?? run gate track a` | `vivechak_run_gate(track="track_a")` or `vck gate` | Evaluates mechanical criteria, headless CLI (`vck gate`), returns structured `track_filter`, decouples Two-Way Door pass |
| **#13**| `vivechak status? show me visual graph if u can` | `vivechak_status()` + `vivechak_visualize()` | Returns structured status counts (counts standalone `C-*.md`), terminal completion `next_step`, renders Mermaid/ASCII graph |
| **#14**| `run vivechak validate on the whole workspace` | `vivechak_validate()` | Multi-tier auto-discovery validates sessions, plans, ADRs, comparisons |
| **#15**| `save session with bash comments # in code block` | `vivechak_save_session()` | `FenceTracker` prevents `#` comment from truncating section |
| **#16**| `vck compare winston vs pino` | `InferScope` via `prepare_generator` | Command shorthand triggers comparison scope immediately |
| **#17**| `run gate with standalone D-01 and DECISIONS.md` | `vivechak_run_gate()` | Evaluates dual sources without omission, deduplicates IDs |
| **#18**| `save session citing Grade A with URL in table` | `vivechak_save_session()` | Cross-references ledger table anchors, parses `[Grade: A]` colon syntax, 0 false alarms |
| **#19**| `vck setup cursor` or `vck mcp-config` | CLI setup command | Number-preserving merge across 21 presets, valid JSON |
| **#20**| `vck doctor` | CLI doctor command | Reports active scope, template layout, and frontmatter health |
| **#21**| `Write a python function to compute HMAC...` | None (Direct Execution) | Agent directly writes code/fixes bug, 0 Vivechak tool calls |

---

## 🛡️ Troubleshooting Heuristics for Small Models in Kiro IDE

| Observed Model Failure Mode | Why It Happens in Small Models | Vivechak Built-in Defense & Remediation |
|---|---|---|
| **Memory Hallucination**<br>*(Answering comparison from training weights without calling tool)* | Small models have high bias to immediately generate answer text. | Front-loaded `ServerInstructions`: `"MANDATORY TRIGGER RULES: ALWAYS invoke vivechak_prepare_generator for comparisons..."` |
| **Skipping `vivechak_init`**<br>*(Calling `save_plan` on an uninitialized directory)* | Small models forget multi-step setup sequences. | `tool_save_plan` detects missing `.vivechak.json` and auto-initializes workspace on-the-fly, generating `research/.gitignore` (`*.lock`). |
| **Argument Misnaming**<br>*(Passing `decision_id` instead of `session_id`, or `content` instead of `amendment`)* | Small models confuse argument names across similar tools. | `tool_challenge` and `tool_amend_session` accept aliases (`decision_id`, `challenge_type`, `content`, `reason`). |
| **Skipping Decision Recording**<br>*(Leaving decisions as generator stubs in `DECISIONS.md`)* | Models assume generated plan stubs are finalized decisions. | `tool_save_session` generates `auto_draft_from` one-liners in `next_step` to guide immediate ADR recording. |
| **Skipping the Exit Gate**<br>*(Finishing synthesis and telling user "Done! Ready to code")* | Models don't realize the Phase 0 Gate is mandatory before coding. | `tool_save_session` upon `SYN-01` save explicitly directs the agent to call `vivechak_run_gate` in `next_step`. |
| **Track Contamination at Exit Gate**<br>*(Blocked from coding fast Two-Way Door spikes due to unrelated One-Way Door checks)* | Gate monolithic evaluation blocks all coding regardless of reversibility. | Active track filtering (`track="track_a"`, `E-40`) and decoupled clearance (`E-42`) allow reversible tasks to pass independently. |
| **Registry Drift on Amendment**<br>*(Amending standalone `D-*.md` without updating `research/DECISIONS.md`)* | Models don't remember to patch both the ADR file and consolidated registry table. | `tool_amend_session` automatically recompiles `research/DECISIONS.md` whenever an ADR file is amended (`E-41`). |
| **Grade Inflation**<br>*(Assigning `[Grade A]` to unverified claims from memory)* | Models use Grade A as a rhetorical emphasis tool. | `core.ObserveSessionQuality` flags `Q-GRADE-INFLATION` in `next_step`, keeping session in `status: draft`. |
| **Code Comment Truncation**<br>*(Discovered concerns truncated on `#` comments)* | Section extractors treating code comments as H1 headings. | `markdown.FenceTracker` tracks CommonMark code fences, ignoring `#` inside code blocks (`E-28`). |
| **Split-Brain Fallback**<br>*(Omitting `DECISIONS.md` when standalone ADRs exist)* | Fallback logic checking `len(sources) == 0`. | Dual-source resolution scans both standalone files and consolidated registry concurrently (`E-29`). |
