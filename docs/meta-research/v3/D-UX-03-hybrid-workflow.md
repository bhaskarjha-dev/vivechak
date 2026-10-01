# Hybrid human–agent session execution for Vivechak

Oct 1, 2026 · @Bhaskar

Hybrid execution is Vivechak’s native mode, so the gap is a contract for the manual leg: hand off per session on a named trigger, return by file-drop plus one idempotent register call, and grade relayed claims as `secondhand` until the pipeline re-fetches the critical sources. Manual execution wins on access, API gaps, cost and cross-vendor triangulation, not on research quality, and as a default it costs more than it returns. Decision confidence is medium, and the public repo has no MCP surface yet, so MCP-level advice here is a design to diff, not a patch.

**Evidence tags.** `(S12: A, corr, aging, ind, Fx)` = ledger source · grade A–E · corroboration (`single`, `corr`, `contested`) · recency (`fresh`, `aging`, `stale`; FRAMEWORK §5.2 half-lives, measured from 2026-10-01) · directness (`dir`, `ind`) · read depth (`F` page opened in full, `Fx` retrieved as a search excerpt of the page, `SH` secondhand). This is the §5.5 composite format with commas in place of pipes. In tables, S-numbers point to the ledger. **\[Proposal\]** and **\[Inference\]** mark my own reasoning, not evidence.

## Research question

When an MCP-connected agent runs a Vivechak pipeline but a human runs some sessions in a browser AI tool and pastes the results back, what should the workflow look like, and what must Vivechak change in `next_step` text, tool parameters and documentation?

**Baseline caveat.** The public repo (main, v1.1.0, ROADMAP dated Sep 2026) has no MCP server, tool descriptions or `next_step` strings in the files I could read: README, ROADMAP, AGENTS, FRAMEWORK and GENERATOR (S1–S5: A, single, fresh, dir, F). The automated Engine is listed as “Next”, and only its v1.0 milestone mentions human-in-the-loop gates. I could not open `templates/`, `examples/` or `meta-research/`. Tool names below are placeholders to diff against the real server.

## Key findings

1. **Hybrid execution is Vivechak’s native mode; the manual leg lacks a contract.** The README places Vivechak above Gemini, ChatGPT and Perplexity deep research: prompts are pasted in and outputs saved by filename. FRAMEWORK §3.2 defines a session as one prompt sent to a vendor deep-research mode. The ROADMAP presents removing the copy-paste-save loop as the point of the first runner. Nothing public says how an agent should request, receive or grade a human-run session (S1, S2, S4: A, single, fresh, dir, F).
2. **Manual execution is justified by capability gaps, not quality.** In a 2026 blind benchmark, three general coding agents outscored a purpose-built research product (S27: C, single, fresh, ind, F). On expert-written consulting tasks, Claude with search, o3-deep-research and Gemini deep research all reached only 13–16% decision-grade acceptance, and failed differently (S28: B, single, fresh, ind, Fx). What holds up: access that lives in the human’s session; products with no usable API path (OpenAI listed its deep-research API models for shutdown on 2026-07-23, S31); cost and quota structure; cross-vendor triangulation (P7); and policy limits on automating consumer apps (S33).
3. **Humans are weak verifiers: use them as executors with access, and mechanize verification.** A 106-experiment meta-analysis found human–AI combinations did worse than the best of human or AI alone (g = −0.23), with losses where the AI was stronger (S12: A, corr, aging, ind, Fx). Anthropic reports users approved about 93% of permission prompts (S13: C, single, aging, ind, Fx). A vendor-run study, reported secondhand, found humans caught 13.6% of planted dangerous commands, falling to about 5% after 50+ prompts (S14: C, single, fresh, ind, SH). Deep-research systems showed 40–80% citation accuracy (S29: A, corr, aging, ind, Fx).
4. **Mature precedents agree on what a handoff contains.** CI/CD gates, data-pipeline HITL operators and agent-framework interrupts converge on six parts: a durable record with a stable ID; a stated reason and instruction; a typed response; explicit allowed responses, including decline; defined timeout behavior; notification where the human already is. They also separate blocking from optional steps (S19–S24: A, corr, fresh, ind, Fx). Vivechak has the prompt-in-a-fence and the filename; it lacks the rest.
5. **Re-integration decides whether hybrid mode pays.** Naive copying tends to drop source URLs and leave tool-specific citation markers (S34: D, corr, aging, dir, Fx). Gemini’s Deep Research agent documents no structured-output support (S30: A, single, fresh, dir, F), so the pipeline cannot rely on the browser tool to emit YAML. Large payloads should not pass through model output tokens. Relayed claims should enter as `secondhand` until the pipeline re-fetches the critical sources \[Proposal\].
6. **MCP has the right primitives, with uneven client support.** The 2026-07-28 revision reworked elicitation as multi round-trip requests and has an official Tasks extension with an `input_required` state (S6: A, single, fresh, dir, F; S7, S8: A, corr, fresh, dir, Fx). Form elicitation takes flat primitive fields, which cannot carry a 5,000-word report. Reports on client support are patchy (S9: D, contested, fresh, dir, Fx). Make the contract plain tool results with explicit state, and treat elicitation and Tasks as capability-gated upgrades \[Proposal\].
7. **Friction finding: as a default, hybrid mode costs more than it returns.** It pays when each manual session has a named trigger, is cheap to decline, and limits the human to executing, copying and saving. Without that, expect rubber-stamping, garbage-in pastes and stalled DAGs \[Inference\]. Detailed finding 6 gives kill criteria.

## Recommendation

**Adopt hybrid as a per-session execution mode: opt-in by trigger, with a delegate-back exit.** Decision confidence is **medium** (FRAMEWORK §5.6). The mechanisms are well precedented, but no source tests Vivechak’s setting, and several load-bearing facts are single-source or about products that change every six months.

1. **Route per session.** Add `execution_hint` (agent, manual, either), `manual_reasons`, `blocking`, `triangulation_group`, estimated active and wall-clock minutes, `prompt_sha256` and `expires` to the session metadata table.
2. **Hand off with a Manual Session Card:** one durable artifact with the reason, the exact one-paste prompt, a per-platform copy recipe, the return path and the decline path (detailed finding 3).
3. **Return by file-drop plus register.** The human saves the raw output to a known path, and the agent calls one idempotent register tool. The raw file stays immutable; a normalized copy is derived (detailed finding 4).
4. **Grade conservatively.** Relayed claims enter as `secondhand`. A verify step re-fetches critical sources and upgrades only what it confirms, so Phase 0 step 4 holds unchanged.
5. **Replace prose `next_step` with templated, versioned messages** that name the owner, forbid self-execution, say what to continue in parallel and give the exit ramps (detailed finding 5).
6. **Document hybrid mode** in README, AGENTS.md, FRAMEWORK §3 and a dated recipes page.
7. **Run a paired pilot, with kill criteria set in advance, before any manual path becomes a default** (detailed finding 6).

**Review (P8).** `review_date: 2027-04-01`. Review triggers: a change in MCP client support, a vendor changing deep-research access, or pilot results. Prediction to falsify \[Proposal\]: at most about 25% of sessions in a Tier 2 pipeline will carry `execution_hint: manual`, and most of those will be `ACCESS_SESSION` or `TRIANGULATE`.

### Handoff decision rule \[Proposal\]

Apply in order. Stop at the first match.

| # | Test | Outcome |
| --- | --- | --- |
| 1 | Two-way door with a known pattern (the framework’s skip-or-default tier) | Never hand off |
| 2 | An automated path meets every session requirement within budget | Agent runs it |
| 3 | A non-substitutable trigger holds: `ACCESS_SESSION`, `NO_API_PATH`, or `TRIANGULATE` with fewer than two available vendor backends | `manual`; `blocking` if the session is the sole evidence for a one-way door or a required triangulation leg |
| 4 | Only a substitutable trigger holds: `COST_CAP` or `POLICY` (direction depends on the organization) | `either`; ask once; delegate-back allowed |
| 5 | `STEER` alone | Optional. Offer the vendor’s plan-review step, never a mid-run chat |
| 6 | “The browser tool is better” with no measured gap | Not a trigger until the pilot shows one |

## Alternatives considered

| Alternative | Verdict |
| --- | --- |
| Agent only, calling vendor deep-research APIs where they exist | The best friction profile, and the right default when keys and budget exist. Gemini’s agent costs roughly $1–3 per standard task and $3–7 for Max (S30: A, single, fresh, dir, F). It falls short where the API path is gone (OpenAI, S31), where access lives in the human’s session, and on flat-fee economics. Keep it as the default arm. |
| Human-only manual loop (status quo) | Highest friction. The ROADMAP already treats copy-paste-save as the cost to remove (S2). |
| Agent drives the consumer web apps (computer use) | OpenAI’s consumer terms forbid automatically or programmatically extracting Output except through the API (S33: A, single, aging, dir, Fx). I did not check Google, Perplexity or Anthropic consumer terms. Brittle and policy-risky; not recommended. I am not a lawyer, so this needs a legal read. |
| Push from the browser tool through an MCP connector | ChatGPT deep research can connect to MCP apps (S32: A, corr, aging, dir, Fx). Permissions and availability vary by plan, and are reported as read/fetch-only on some tiers (S35: D, single, fresh, ind, Fx). A later experiment. |
| Elicitation-only handoff | Flat primitive fields, patchy client support, built for in-call round trips (S6, S9). Fine for the 30-second record, wrong for the report. |

## Detailed findings

### 1. Survey: how mixed human–machine workflows are built

The precedents agree on structure but differ in weight. Approval gates are cheap and binary; Vivechak’s manual leg is long and returns a large artifact.

**Workflow engines and CI/CD**

| System | Human step declared as | What the human sees | How the result returns | Timeout or default | Lesson for Vivechak |
| --- | --- | --- | --- | --- | --- |
| GitHub Actions environments (S19) | A job targets an environment with required reviewers | A notification, then “Review deployments”; approve or reject with an optional comment | Approval releases the waiting job | Optional wait timer; “prevent self-review”; admin bypass requires a comment | Separation of duties and a justified bypass are normal. Mirror the bypass as `waive` with a required reason |
| GitLab manual jobs (S20) | `when: manual`, optional (`allow_failure: true`) or blocking (`allow_failure: false`) | A play button; un-run manual jobs show as “skipped”; optional confirmation dialog and run-time variables | Running the job continues the pipeline | Optional jobs never block | Blocking versus optional is first-class. Never encode “needs a human” as a quiet state |
| Jenkins `input` (S21) | A step with `message`, stable `id`, `ok` label, `submitter` allow-list and `parameters` | A prompt with a message and form | Typed return: one parameter returns its value, several return a map. `submitterParameter` records who acted. The `id` lets external tools answer | Waits indefinitely unless wrapped in `timeout` | Stable ID, typed return, record the actor, always set a timeout |
| AWS Step Functions callback (S22) | `.waitForTaskToken` hands an opaque token to any external actor | Whatever your system sends: email, queue, chat | The caller sends success with output, or failure with error and cause | Heartbeats and timeouts prevent indefinite waits | A human step is an external task keyed by a correlation token, with success and failure channels |
| Airflow 3.1+ HITL operators (S23) | `HITLOperator`, `ApprovalOperator`, `HITLEntryOperator`: subject, options, a `params` form, assigned users, notifiers | A “Required Actions” inbox with a deep link | Chosen option and form values flow to downstream tasks | `defaults` apply on timeout; with none, the task fails | A central inbox, typed form fields, explicit timeout semantics, notification where people are |

**Agent frameworks, protocols and research practice**

| System | Pattern | Lesson for Vivechak |
| --- | --- | --- |
| LangGraph `interrupt` and `Command(resume=…)` (S24) | A pause surfaces a payload; the resume value becomes the return of `interrupt()`; the node re-runs from its start on resume, so earlier side effects must be idempotent; a thread ID is the durable pointer | Idempotent resume keyed on a session ID. Validate returned input and re-ask with a specific error |
| LangChain Agent Inbox (S24) | The request declares allowed responses (`allow_accept`, `allow_edit`, `allow_respond`, `allow_ignore`) plus a Markdown description; the response is accept, edit, response or ignore | The request itself declares what the human may do. Map “ignore” to decline-or-delegate and “edit” to an edited prompt |
| 12-factor agents, factors 6, 7, 11 (S17) | Contact humans through typed tool calls such as `request_human_input`; pause and resume through simple APIs; reach people where they already are | Make “ask a human” a typed tool result, not prose |
| MCP elicitation and multi round-trip requests (S6, S7) | A call returns an input-required result with `inputRequests` and an opaque `requestState`; the client retries with `inputResponses`; user actions are accept, decline, cancel; form schemas are flat primitives | Right for short structured questions; not a paste box for reports |
| MCP Tasks extension (S8) | A durable handle with `working`, `input_required`, `completed`, `failed`, `cancelled`; `tasks/get` surfaces `inputRequests`; `tasks/update` answers | The closest native fit for a long human step; client support unproven |
| Gemini Deep Research API and ChatGPT deep research (S30, S32) | Gemini: `collaborative_planning` returns a plan for review and refinement before execution. ChatGPT: plan edit and interrupt while running | Vendors already put a human gate inside the tool, at plan time |
| Magentic-UI (S15) | Co-planning, co-tasking, action guards, answer verification | Timed human input at plan time and when stuck raised GAIA success from 30.3% to 51.9% with a simulated user who knew extra task information |
| Systematic-review platforms (S37, S38) | Rayyan and Covidence layer AI suggestions on a human-driven dual-screening workflow where the human still records each decision. Elicit screens and extracts itself and calls for manual verification. One study proposes the AI as second extractor, with the human reconciling discrepancies | Spend human attention on disagreements between two independent passes, not on re-reading everything |
| RAISE reporting guidance (S18) | Report AI tool, version, date of use, purpose and degree of human oversight; humans stay accountable | A ready provenance template for relayed AI output |

**Cross-cutting lessons \[Inference from the rows above\]**

- **Return-path patterns matter most:** typed return, correlation ID, idempotency, validate-and-re-ask. Approval patterns (reviewers, separation of duties) matter least, because the manual leg is expensive and open-ended.
- Make the pending step a durable record with a stable ID: a Jenkins `id`, a Step Functions token, a LangGraph thread, an Airflow task instance.
- Declare blocking versus optional, and let the DAG continue independent branches. FRAMEWORK §3.1 already prefers parallel Layer 0.
- Put the reason, instruction and response schema in the record the human sees.
- Declare allowed responses, including decline, and define timeout behavior up front.
- Record who or what acted, and when. Make resume idempotent. Provide a justified bypass.
- Keep gates rare, because they decay into rubber stamps (section 2.4).

### 2. When does manual execution beat automated?

Manual execution wins on access, API gaps, cost and triangulation. The quality evidence does not support a general claim that browser tools beat agents with search.

#### 2.1 What the quality evidence says

| Source | Date | Comparison | Finding | Tag |
| --- | --- | --- | --- | --- |
| DeepResearch Bench (S25) | 2025 | Deep-research products versus models with search tools | Products beat search-augmented models on effective citations. Gemini 2.5 Pro Deep Research led (about 111 per task); Perplexity Deep Research led on citation accuracy (about 90%) | B, single, stale, ind, Fx |
| FutureSearch Deep Research Bench (S26) | 2025 paper; 2026 leaderboard | Agents on a frozen web, plus commercial products | ChatGPT o3 with web research beat OpenAI’s o3-based Deep Research. Gemini Deep Research beat Gemini 2.5 Pro with search. Best agents sat well below the human-noise ceiling; standard-harness runs cost cents per task | B, single, aging, ind, Fx |
| AIMultiple DR-20 (S27) | Sep 2026 | Codex CLI, Claude Code, Grok CLI, Gemini CLI and Exa Agent on 20 business briefs; blind LLM judges | Codex 0.790, Claude Code 0.775 and Grok 0.699 beat Exa Agent 0.584, the purpose-built product. No tool reproduced half the required facts (best 202 of 434). 14–63 broken links per 20 reports. $0.58–$18.35 per task. Briefs unpublished; consumer deep-research apps not tested | C, single, fresh, ind, F |
| Consulting benchmark (S28) | May–Jun 2026 | Claude Opus 4.6 with search, o3-deep-research, Gemini 3.1 Pro deep research; 70 expert prompts, 210 responses | Acceptance 13–16% for all three, statistically indistinguishable. Rank order depends on the metric. Claude is strongest on long-document reasoning and weakest on structural obedience; Gemini is the reverse | B, single, fresh, ind, Fx |
| DeepTRACE (S29) | 2025; ICLR 2026 | Search engines and deep-research systems, audited claim by claim | Large fractions of statements unsupported by their own listed sources. Citation accuracy 40–80% across systems | A, corr, aging, ind, Fx |

**Reading the evidence \[Inference\]**

- **No source tests the case that matters.** No benchmark compares a human operating a browser deep-research app against an agent with search on architecture-decision research. The closest are business, consulting and general web tasks, run through APIs or CLIs.
- **Three arms are routinely conflated:** (A) an agent with generic web search, (B) an agent calling a vendor deep-research API, (C) a human running the vendor’s app. B and C use the same vendor agent family, but parity between the app’s agent and the API’s preview agent is undocumented in what I read.
- **Product wrappers sometimes add value and sometimes do not.** Gemini Deep Research beat Gemini 2.5 Pro with search in 2025, while ChatGPT o3 with web research beat OpenAI’s own product (S26). Both findings are vendor- and time-specific.
- **Reliability binds every arm.** No tool reproduced half the required facts (S27), acceptance sat at 13–16% (S28), and citation accuracy ranged 40–80% (S29). S27 cautions that polished output discourages checking.
- **Cost depends on depth far more than on arm:** cents per task on a standard harness (S26), roughly $1–7 for Gemini’s API agent (S30), $13–18 for the best coding-agent runs (S27).

#### 2.2 Where manual execution genuinely wins

| Trigger | Why a human helps | Evidence | How the pipeline can detect it |
| --- | --- | --- | --- |
| `ACCESS_SESSION` | Entitlements live in the human’s browser: SSO, institutional subscriptions, paywalled outlets, personal workspaces. APIs can take file stores and MCP credentials, but not a logged-in session | S30, S32 | The generator tags `requires_sources`; an agent fetch returns 401, 403 or a paywall |
| `NO_API_PATH` | A product has no supported programmatic path for this person or tier, or keys and budget are missing. OpenAI’s dedicated deep-research API models were listed for shutdown on 2026-07-23, while Gemini’s agent is now API-callable. The trigger is vendor-specific and moving | S30, S31 | The executor registry lists available backends; none match |
| `TRIANGULATE` | P7 asks for identical prompts on two or three models for contested one-way doors. Agents fail distinctively (S28), and systems vary widely on citation accuracy (S29). A human can reach a second vendor without keys | S4, S28, S29 | A `triangulation_group` with fewer than two available backends |
| `COST_CAP` | Per-task API cost for deep runs ranged from about $1 to $18, and a flat subscription changes the marginal cost. Plan quotas change often and were not pinned down here | S27, S30 | Estimated cost above the session budget |
| `STEER` | Vendor UIs offer plan review and mid-run interrupts, and timed human input helped in Magentic-UI’s simulations. But mid-run follow-ups conflict with Vivechak’s front-load rule (FRAMEWORK cites a 39% multi-turn drop; not re-verified here). A weak trigger | S4, S15, S30, S32 | High scope ambiguity |
| `POLICY` | Consumer-app terms restrict automated extraction (S33), and data-handling rules may require an approved tool. The direction depends on the organization | S33 | The Regulatory Exposure = 3 override (FRAMEWORK §3.2) |

#### 2.3 Where manual execution loses \[Inference\]

- **Structured or machine-readable output.** Gemini’s agent documents no structured-output support (S30), and FutureSearch reported being unable to obtain valid JSON from Gemini Deep Research in 2025 (S26).
- **Hard-dependency sessions.** Each dependency adds a human round trip. Keep them with the agent unless a trigger applies.
- **Prompts still being iterated.** Iteration latency dominates.
- **Wide fan-out.** Layer 0 can run in parallel, but human bandwidth then becomes the critical path.

#### 2.4 Human-factors evidence

**Combinations are not automatically better.** The Nature Human Behaviour meta-analysis (106 experiments, 370 effect sizes) found combinations did worse than the best of human or AI alone (g = −0.23, 95% CI −0.39 to −0.07). Decision tasks lost and content-creation tasks gained. Where the human was the stronger party the combination gained; where the AI was stronger it lost (S12: A, corr, aging, ind, Fx).

**Gates decay.** Anthropic’s telemetry shows about 93% of permission prompts approved. Experienced users auto-approve roughly twice as often as new ones while interrupting more (S13). The vendor-run study reports detection falling with exposure (S14; vendor-motivated, reported secondhand, one planted command). Anthropic’s answer was to change the environment so fewer questions are needed, not to ask better ones \[Inference from S13\].

**Well-timed input helps.** Magentic-UI’s simulated-user runs improved GAIA success from 30.3% to 51.9% when the user helped at plan time and when the agent was stuck (S15: B, single, aging, ind, Fx). The users were simulated, so treat this as directional.

**Implication \[Inference\].** Use the human where the human is the stronger party: access, entitlements and execution inside a vendor UI. Do not ask the human to out-read a model on a polished report. Keep human accountability (S18), but spend attention on discrepancies and unreachable sources, not on a full read.

### 3. Handoff communication: how to signal “do this manually”

A good handoff names the owner, the reason and the exit in one record. The principles below combine the CI/CD, pipeline and human–AI interaction precedents.

#### 3.1 Principles

1. **Name the owner and the blocking state first.** GitLab shows un-run manual jobs as “skipped” by default (S20). A needed human step must never look like a skipped one.
2. **Give a reason code and one plain sentence.** Humans ask “why me?” first, and the HAX guidelines ask systems to explain why they acted as they did (S16: A, corr, stale, ind, Fx; used as a design heuristic). Codes such as `ACCESS_SESSION` and `TRIANGULATE` let you measure which triggers pay off.
3. **State the cost honestly:** active minutes versus waiting minutes \[Inference\]. Gemini’s agent runs up to 60 minutes, most within 20 (S30).
4. **Put everything in one record:** reason, exact prompt, steps, return path, decline path. Airflow’s subject plus form params and Agent Inbox’s description play this role (S23, S24). Use one paste, because Vivechak bans drip-feeding (S4).
5. **Declare allowed responses, including decline.** Agent Inbox’s request carries `allow_*` flags (S24), and the HAX guidelines ask for efficient dismissal (S16).
6. **Pre-answer the vendor’s own questions.** Gemini offers plan review (S30), and ChatGPT shows a plan and accepts interrupts (S32). Tell the human to approve the plan unchanged, record any edit, and answer clarifying questions with “proceed with the scope as written” \[Inference\].
7. **Set an expiry and its consequence.** Jenkins waits forever unless wrapped in a timeout (S21). Airflow applies defaults or fails (S23).
8. **Keep gates rare.** Every unnecessary request teaches the user to ignore the next (S13, S14). Ask once per session, and never re-ask on the same blocker.
9. **Write tool descriptions like onboarding a new hire:** make implicit context explicit, name parameters unambiguously, return actionable errors (S11: B, single, aging, dir, F).

#### 3.2 The Manual Session Card \[Proposal\]

The card is the human-facing artifact. The agent shows it verbatim and does not paraphrase it.

```markdown
# Manual session T1-01: Primary datastore selection
Owner: you · Blocking: yes (D-001 stays unconfirmed without it) · About 8 min of your time plus about 20 min of waiting · Expires 2026-10-08

Why you? (TRIANGULATE) D-001 is a one-way door and the agent holds one single-vendor result. Running the identical prompt in a second vendor’s deep-research mode is the cheapest independent check.

Do this
1. Open Gemini (Deep Research) or ChatGPT (deep research) in a new chat.
2. Paste the whole prompt below as one message. Add nothing.
3. If a plan appears, approve it unchanged. If you edit it, note what you changed.
4. If it asks a question, reply: Proceed with the scope as written.
5. When it finishes, copy the report with its sources (recipe for your platform below).
6. Save it as research/sessions/raw/T1-01.raw.md and put a 30-second note on top: platform, mode, model shown, date, plan edited?, anything odd.
7. Tell your agent: T1-01 saved.

Can’t or won’t? Tell your agent: delegate T1-01 (the agent runs it with its own search and labels the result single-vendor), or: waive T1-01 because …

Data check: the prompt contains project details. Run it only in a tool you are allowed to use for this project.

[generated prompt block goes here, fenced as a prompt block]
[platform copy-with-sources recipe goes here]
```

#### 3.3 Platform recipes: a template to fill by test, not by assertion

I could verify only fragments. Each recipe needs `verified_on` and `review_by` fields (six months) and a test pass on current accounts before it ships.

| Platform | Mode | Plan step | Copy that keeps sources | Known artifacts | Status |
| --- | --- | --- | --- | --- | --- |
| Gemini | Deep Research | Plan review exists in the API via `collaborative_planning` (S30); app behavior not verified | Not verified | Not verified | Unverified |
| ChatGPT | Deep research | Shows a plan; accepts edits and interrupts (S32) | Not verified | Plain copy tends to drop source URLs; inline markers such as 【3†L157-L164】 appear in copies (S34) | Partly reported |
| Perplexity | Deep Research | Not verified | Not verified | Copies keep numbered markers; Markdown export uses footnote-style references (S34) | Partly reported |
| Claude | Research | Not verified | Not verified | Not verified | Unverified |

### 4. Re-integration: accepting manual results into the pipeline

Return the result by file-drop, validate it mechanically, and grade it conservatively. The human should never do format work, name files or write YAML.

#### 4.1 One session contract, two executors \[Proposal\]

Treat the human as another executor behind the same session contract. Input: prompt plus context injection, identified by `prompt_sha256`. Output: a result that passes the validators. The DAG does not care who executed; the human executor differs in latency, failure profile and mandatory validation.

- **Server:** compiles prompt and context, writes the card, fixes filenames, normalizes, validates, re-fetches citations, diffs against sibling sessions, updates the DAG.
- **Human:** runs the vendor tool, approves the plan, copies with sources, saves one file, writes a 30-second note.
- **Vendor tool:** does the research.

#### 4.2 Land raw, then derive

Land the paste unchanged and immutable (`research/sessions/raw/<ID>.raw.md`, with a content hash). Then derive the normalized artifact under the existing filename contract (`research/sessions/<ID>-<slug>.md`), so decisions and synthesis downstream do not change. A practitioner skill that fans one question out to four vendor APIs uses the same raw-then-structured split (S36: D, single, fresh, ind, Fx). No agent should clean up the raw text.

#### 4.3 How the text gets from browser to pipeline

| Route | Fit | Notes |
| --- | --- | --- |
| File-drop, then register by path | Best default | Content never passes through model output tokens. Works when a local server reads the workspace. Matches the existing save-by-filename habit |
| Paste into chat; the agent passes `result_text` | Short results only | The model must re-emit every token. Claude Code caps tool responses at 25,000 tokens by default (S11). Risk of silent paraphrase |
| Elicitation form | Metadata only | Flat primitive fields; client support uneven (S6, S9). The answer returns as client-collected input, not model tokens |
| Own upload page via URL-mode elicitation | Later | Needs a UI, which is out of scope here |
| Push from the vendor tool through an MCP connector | Experiment | Connectors are reported read-only or plan-gated on some tiers (S32, S35) |

#### 4.4 Validation ladder

V0–V4 are deterministic and cheap. V5–V7 need a model or a human.

| Check | Catches | On failure |
| --- | --- | --- |
| V0 Integrity: non-empty, readable, not cut off mid-sentence, plausible length | Truncated copies | `needs_fix`: copy the whole report again |
| V1 Structure: seven sections found or mappable | Wrong output type; a chat transcript pasted | `needs_fix` naming the missing parts |
| V2 Provenance note present: platform, mode, date, plan edited | Unknown origin | Ask once for the 30-second note; do not block |
| V3 Sources present: distinct URLs counted, vendor markers detected | Lost sources; a model with no web access (FRAMEWORK §3.1 caps that at Grade D, recalled) | `needs_fix` pointing to the recipe step that keeps sources |
| V4 Liveness: HTTP check of cited URLs | Dead or invented links (14–63 per 20 reports in S27) | Warning listing dead links |
| V5 Temporal anchor: the report respects today’s date | Stale framing | Warning |
| V6 Support spot-check: sample claims against the fetched page, using a different model; paywalled items become excerpt requests to the human | Unsupported claims (S29) | Warning, plus per-claim downgrade |
| V7 Rubric (FRAMEWORK §4.3) and sibling diff | Quality gaps; disagreement inside a triangulation group | Show only the disagreements to the human (S38) |

#### 4.5 Status model

```
ready -> agent_running -> returned
ready -> awaiting_human -> returned
awaiting_human -> delegated | waived | expired
returned -> validating -> accepted | accepted_with_warnings | needs_fix | rejected
needs_fix -> returned        (human fixes, registers again)
rejected -> awaiting_human | delegated   (other platform, split prompt, or agent)
```

| Status | Meaning | Agent does | Human does |
| --- | --- | --- | --- |
| `accepted` | Passed V0–V4, no warnings | Runs citation verification on critical citations; continues | Nothing |
| `accepted_with_warnings` | Usable, issues listed | Mentions them in one sentence; continues | Nothing unless asked |
| `needs_fix` | Not registered; one concrete problem | Relays the single action; never repairs the text | One action, then re-registers (safe to repeat) |
| `rejected` | Matches the FRAMEWORK §3.1 garbage-session pattern | Offers another platform, a split prompt, or delegation | Chooses |
| `delegated`, `waived`, `expired` | Closed without a human result | Records it; downgrades decision confidence; a waiver needs a reason | Decides |

#### 4.6 Idempotency and versioning

Key each submission on a hash of session ID plus normalized content, so a repeated call returns the earlier result. LangGraph re-runs a resumed node from its start (S24), and Step Functions ties each wait to a token with heartbeats and timeouts (S22). Duplicate delivery should therefore be harmless by design. A changed re-paste is a new version: `on_conflict: supersede` keeps the old raw file and links the two.

#### 4.7 Provenance and grading

RAISE asks reviews to report the tool, version, date, purpose and degree of human oversight (S18: A, corr, fresh, ind, Fx). Add an `execution` block to result frontmatter. This session’s own record is the worked example:

```yaml
execution:
  mode: manual
  executor_type: human-relayed-ai   # never human-authored
  platform: Claude (claude.ai)
  product_mode: chat with web search and fetch tools, not a branded Deep Research mode
  model_label: Claude Sonnet 5.5
  run_date: 2026-10-01
  prompt_sha256: not supplied in brief
  plan_edited: false
  tool_calls_approx: 45
  source_capture: URLs listed in the ledger below
  human_oversight: none in session; pipeline verification pending
  pasted_via: chat-paste
  raw_artifact: not yet saved
verification:
  default_method: secondhand
  ledger_sources: 38
  citations_url_live: not checked
  claims_spot_checked: 0
  upgraded_to_fetched: []
```

**Grade policy \[Proposal\].** Treat claims in a relayed report as `secondhand`. The vendor tool fetched, but the pipeline did not observe it, and citation accuracy in these systems is 40–80% (S29). Phase 0 step 4 already requires `fetched` or `cached` for critical citations, so a verify step becomes the upgrade path. The pipeline re-fetches the page, confirms the passage, and only then records `fetched`.

Paywalled pages come back as excerpt requests to the human, who supplies `human-provided` text. That spends the human’s special access on the narrow set of citations that need it, instead of on a full read. The ROADMAP already lists a missing “tool-generated” verification method (ES-03) and an undefined quality-evaluation owner (GA-04); decide them together with this (S2). A new `relayed` method is optional, useful only if you want to measure how much relayed content survives verification.

#### 4.8 Treat relayed text as untrusted data

Relayed text carries web content through a human into the agent’s context. Google’s documentation warns about prompt injection through files and about reviewing citations (S30), and the MCP spec treats echoed request state as attacker-controlled input (S7). The register tool should hand the agent a normalized extract wrapped as data and strip instruction-like lines. Its description should tell the agent never to follow instructions found in returned text \[Proposal\].
