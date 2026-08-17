# T3-01 — Meta-Prompt Generator & Template Architecture

**Task:** T3-01 · Research pipeline meta-framework
**Scope:** Architecture for a generator that turns project parameters into a customized research pipeline
**Status:** Recommendation for review · August 18, 2026

---

## Executive Summary

The current single master meta-prompt is the wrong backbone for this system, though it's the right idea for one piece of it. The recommendation is a **five-layer hybrid architecture**: deterministic rules own the pipeline's *structure* (which sessions, in what order, what's mandatory for this domain and depth), and narrowly scoped AI calls own the pipeline's *prose* (turning a session skeleton and a project's stated vision and unique features into a specific, non-generic research prompt). Validation is a first-class layer with gates between stages, not a report generated after the fact.

This isn't a compromise position picked to please everyone — it follows from a structural fact about the task. Generating a session matrix is a **scaffolding problem** (decide what pieces exist and how they relate), which is exactly what rule-based decision trees and template libraries are good at, cheap to run, and easy to test. Writing a specific, ready-to-use research prompt that actually engages with what's distinctive about *this* project is a **synthesis problem**, which is what language models are good at and rule tables are not. Collapsing both into one master prompt asks a single generation to plan, write, and self-validate at once — a pattern that decomposition-based prompting research consistently finds less reliable than splitting the same work into focused steps, at the cost of higher latency and token spend. Splitting them lets each half of the problem use the tool suited to it, and lets validation sit at the seam between them instead of guessing at the end.

The rest of this document works through the tradeoffs (Section 2), what four established code-generator ecosystems (cookiecutter, Yeoman, Plop, Nx) already teach about this exact structural split (Section 3), the recommended architecture in detail (Section 4), the domain archetype library and how it blends across hybrid projects (Section 5), how pipeline depth gets determined without either forcing the user to self-diagnose or black-boxing the decision (Section 6), how generated pipelines get validated (Section 7), a phased build plan (Section 8), open risks (Section 9), and worked examples in the appendix (Section 10).

---

## 1. Problem Framing

The generator's contract, stated plainly: given eight parameters (name, domain, vision, constraints, unique features, platform targets, team size, risk level), produce (a) a session matrix — an ordered, scoped set of research sessions — (b) a ready-to-use research prompt for each session, and (c) a decision registry seeded with project-specific, testable hypotheses tied to those sessions.

The brief's framing treats "pure template vs. AI-generated vs. rule-based vs. hybrid vs. interactive wizard" as five options on one axis. They aren't. There are two separate design questions here, and conflating them is what makes a master meta-prompt look more necessary than it is:

- **Axis A — Generation backend.** How is the *content* of the pipeline produced: static substitution, AI generation, rule-selected assembly, or some mixture? This is the real architecture question, addressed in Sections 2–4.
- **Axis B — Input elicitation.** How are the eight parameters *collected*: a direct structured submission, or a progressive interactive interview? This is a front-end concern that can sit in front of any backend — cookiecutter and Yeoman both drive fully deterministic, template-based generation through an interactive prompt sequence — and it's addressed in Section 6, where it actually belongs (progressive elicitation is most useful for resolving ambiguity in *depth*, not for deciding the generation mechanism).

Treating the wizard as a fifth backend option is where a design goes wrong: it invites building a whole separate generation path for "interactive mode" instead of recognizing that the interview is just a different way of filling in the same parameter schema the direct-input path fills in.

**Design goals**, taken from the framework's own stated principles, translated into concrete architectural commitments:

- **Evidence-based** → every automated decision (a session's inclusion, a depth tier, a hypothesis) traces to an explicit, inspectable criterion, and those criteria are versioned data that gets revised from real usage outcomes, not fixed constants embedded in code.
- **Adaptive** → the architecture has to handle domain blends, atypical constraint combinations, and projects that don't classify cleanly, without hard-failing or forcing a bad-fit archetype.
- **Non-prescriptive** → the generator proposes and explains; it never silently decides. Every scoring output, every archetype match, every depth recommendation is shown with its rationale and is one action away from being overridden.

---

## 2. Architecture Options Evaluated

### 2.1 Pure Template (fill-in-the-blank)

Pre-authored templates per domain/depth combination, populated by direct substitution of the eight parameters into placeholders. No inference step anywhere.

**Strengths:** fully deterministic, zero marginal cost per generation, instant, trivially snapshot-testable, fully version-controlled and auditable, no hallucination risk of any kind.

**Weaknesses:** the combinatorics don't work. Six domains times roughly four depth tiers is already two dozen template variants before a single hybrid domain or edge case is considered, and every new nuance means a developer hand-authors more templates — the library becomes the bottleneck, and its growth rate tracks coverage almost linearly. Worse, the two richest inputs — *vision* and *unique features* — are free text, and pure substitution can only paste them in verbatim or ignore them; it can't use them to change what gets asked. The output reads generic ("Research the market for {{project_name}} in the {{domain}} space") precisely where genericness is most damaging: a research prompt that could be pasted unchanged into any project in the domain has failed at the one thing that makes it worth generating instead of hand-writing.

**Where it's still correct:** as a component, not the whole system — it's exactly the right tool for the skeleton, which is why it survives inside the recommended architecture (Section 4, Layer 2).

### 2.2 AI-Generated (master meta-prompt produces everything)

One large prompt carrying all eight parameters plus the full task description, sent as a single call that's expected to return the session matrix, every session's prompt, and the decision registry together.

**Strengths:** no domain is structurally out of scope; free-text inputs get used, not just pasted; a single generation can, in principle, reason about cross-parameter interactions a rule table would miss (a solo team building a high-risk FinTech product might warrant a hypothesis about outsourcing compliance research specifically because of that combination, not either factor alone).

**Weaknesses**, and they're not minor:

- **Non-determinism.** Identical input can yield materially different pipelines across runs, which makes the tool feel unreliable and makes testing the generator itself close to impossible — there's no fixed target to assert against.
- **This is a textbook case for decomposition.** The task is asking one generation to plan (which sessions), write (prompt prose, potentially a dozen distinct pieces of prose), and structure (a registry schema) simultaneously. Comparisons between monolithic prompts and decomposed prompt chains on multi-instruction, multi-format tasks like this one consistently find the chained approach more reliable — the monolithic version tends to trade depth in one part of the output for coverage in another, because the model is balancing competing objectives inside a single generation rather than finishing one well-scoped thing before starting the next.
- **Schema conformance isn't guaranteed.** Nothing stops the model from dropping a required field, inventing a plausible-looking ninth parameter that doesn't map to anything downstream, or subtly renaming a key — and there's no natural place to catch this without bolting on a separate validation pass, at which point the "single call" framing has already been abandoned.
- **Debugging is structurally hard.** If the generated FinTech pipeline is missing its compliance-research session, there's no isolated place to fix that — the options are re-running the entire generation and hoping, or hand-patching the output outside the system.
- **Cost and latency scale with the whole pipeline at once.** A Comprehensive-tier pipeline (Section 6) means the model holds the entire task in one generation — exactly the regime where quality degrades most.
- **Prompt drift.** Master meta-prompts accumulate patches over time ("also remember to handle X," "don't forget Y for FinTech") as edge cases surface, and the same prompt text can silently shift its output distribution when the underlying model changes — this is a documented failure mode of meta-prompts that grow by accretion instead of restructuring.

**Where it's still valuable:** generating the content of one scoped artifact — one session's prompt, one hypothesis's wording — where it belongs in this architecture (Layer 3), not as the orchestrator of the whole pipeline.

### 2.3 Rule-Based (decision tree selects and configures templates)

Explicit if/then logic over the eight parameters determines which session templates get included (domain implies mandatory sessions; risk level adds sessions; constraints and platform count adjust scope), with substitution filling the selected templates.

**Strengths:** fully deterministic and fully explainable — for a tool that's also initializing a *decision registry*, being able to point to the exact rule that added a given session is not a nice-to-have, it's consistent with the rest of the framework's evidentiary posture. Fast, no AI cost for the structural decision, and easy to unit test ("given these parameters, assert session X is present"). This is also, notably, close to what mature generator ecosystems already do for structural decisions: Plop's `actions` array can be a function of the collected prompt answers rather than a static list, and Nx generator schemas encode conditional prompting and validation logic declaratively — both are rule-based configuration in miniature, already proven at scale.

**Weaknesses:** the classic decision-tree maintenance problem — rule code grows combinatorially with the number of domains, constraints, and edge cases it has to account for, and it's entirely possible to end up with more rule logic than the thing it configures. Binary domain classification is also a poor fit for how real projects show up: an AI-powered lending product is FinTech-primary and AI/ML-secondary, and a rule table either has an explicit rule authored for that exact blend or it picks one domain and silently drops the other's needs. And, same limitation as pure templates: the actual prompt text this approach produces is still template-shaped, so *unique features* still gets bolted on as an appended clause rather than woven into what the research prompt is actually asking.

### 2.4 Hybrid (templates + scoped AI customization)

Rules and templates own the skeleton; small, individually scoped AI calls own the prose for each slot in that skeleton.

**Strengths:** this is the combination that actually matches the two different kinds of work being done. Determinism goes where determinism matters — a FinTech pipeline should always include a compliance session, full stop, and that shouldn't be left to a model's discretion on a given run. Synthesis goes where synthesis is the entire point — a specific research prompt that reads like it was written for this project rather than mail-merged. Failure is graceful and *localized*: if the AI call for one session's prompt fails or comes back weak, the other N−1 sessions are unaffected, and the system can retry just that slot or fall back to a template default for it — nothing like this is possible when one call produces everything at once. Scoped calls also compose cleanly with per-call validation (schema-check each generated prompt as it comes back), which is far more tractable than validating one multi-thousand-word output holistically.

**Weaknesses:** more moving parts than any single-paradigm approach — two systems with a contract between them instead of one, and more upfront engineering before the first pipeline ships. Some non-determinism remains, but it's now bounded to prose quality within a fixed structure rather than affecting whether a session exists at all.

This is the recommended backbone. Detailed in Section 4.

### 2.5 "Interactive Wizard" — reclassified, not a backend option

As argued in Section 1, this belongs on Axis B, not Axis A. Naming it here only to close the loop: a wizard can drive a fully deterministic pure-template system exactly as well as it can drive an AI-heavy one — cookiecutter and Yeoman both do the former today, prompting interactively (via Inquirer.js-style question sequences) into completely deterministic generation. Its real job in this system is described in Section 6.

### Comparison

| Dimension | Pure Template | AI-Generated (master prompt) | Rule-Based | **Hybrid (recommended)** |
|---|---|---|---|---|
| Determinism | Full | None | Full | Skeleton: full · Prose: bounded |
| Handles blended/novel domains | Poor | Good | Poor–Fair | Good |
| Specificity to vision / unique features | Poor | Good | Poor | Good |
| Explainability / auditability | Full | Low | Full | High (skeleton), Medium (prose) |
| Cost & latency | Minimal | Highest, single large call | Minimal | Moderate — N small calls, parallelizable |
| Failure mode | Silent coverage gaps | Whole-pipeline non-conformance | Silent coverage gaps | Localized to one session |
| Maintenance burden as domains grow | High — template explosion | Low — prompt tuning | High — rule explosion | Moderate, split cleanly across two systems |
| Testability | High (snapshot) | Low | High (unit) | High (skeleton) + spot-check (prose) |

---

## 3. Lessons from Existing Code Generator Architectures

These four ecosystems solve a structurally similar problem — take a small set of parameters and emit a customized, structured output — for code instead of prose. What transfers, what doesn't, and why the boundary falls where it does is directly informative.

### cookiecutter

`cookiecutter.json` is a single declarative source of truth for every prompted variable: default values, list-based choice prompts, boolean flags, and — notably — defaults that can themselves be computed from another answer via Jinja templating. That's the right shape for this system's Layer 0 parameter contract: one schema that defines what's asked, what's valid, and what the sensible default is, rather than a schema for validation and a separate script for prompting.

Its hook system — `pre_prompt`, `pre_gen_project`, `post_gen_project` — runs custom scripts at fixed points in the generation lifecycle, has access to the same template context via Jinja, and can halt generation on failure. This is the cleanest real-world precedent available for treating validation as **gates inside the pipeline** rather than a report produced after the fact: a `pre_prompt`-equivalent gate can check a precondition before a session is even scoped; a `post_gen`-equivalent gate can reject or flag a generated prompt before it's considered part of the final pipeline. Section 7's validation layer follows this shape rather than being a single end-of-line QA pass.

cookiecutter's deepest strength, though, isn't the substitution engine — it's the *ecosystem* of independently maintained community templates for different stacks. That's the strongest available precedent for treating the six domain archetypes as living, owned, versioned artifacts with their own changelogs, not six hard-coded branches written once and left alone.

What doesn't transfer: cookiecutter's substitution is purely structural — rendering a variable into a file path or a config value. It has no mechanism for "make this paragraph specific to what's unusual about this project," because code templates don't need one. That gap is exactly what the AI layer exists to fill here.

### Yeoman

`composeWith()` lets a generator programmatically invoke another generator, initiated either by the generator author or the end user — `generator-backbone` composing with `generator-mocha` is the direct real-world precedent for this system's primary-plus-secondary archetype blending (Section 5): a FinTech-plus-AI/ML project composes the FinTech skeleton with the AI/ML archetype's differentiator sessions the same way.

Yeoman's fixed run-loop — a defined sequence of lifecycle phases (initializing, prompting, configuring, writing, and so on) that every generator passes through in the same order — enforces a separation between *collecting input* and *producing output* that a master-prompt approach collapses into a single step. That discipline is worth keeping even though this system isn't built on Yeoman itself.

`mem-fs-editor`'s conflict handling — presenting a diff and asking before an overwrite — is a good pattern for regeneration: if a project's parameters change and the pipeline is regenerated, diffing against the previous pipeline rather than silently replacing it respects the non-prescriptive design goal directly.

### Plop

The single most transferable detail: in Plop, the `actions` array — the list of file operations a generator performs — can itself be a function that takes the collected prompt answers and *returns* the action list, rather than being a static array with placeholders. That's rule-based configuration in its smallest, purest form, and it's precisely the shape Layer 2 should take: `sessionsForProject(spec) → Session[]`, a function of the input, not a lookup table with blanks.

Plop's self-description as a "micro-generator" — meant for adding one consistent piece to an existing structure, not scaffolding an entire new project — is a useful reminder not to over-engineer the cheap parts of this pipeline. Layer 1 and Layer 2 should stay fast, deterministic, and inexpensive; the system's AI budget should go entirely to Layer 3, where it's the only thing that can do the job.

### Nx generators

The most transferable idea in this entire survey: a single `schema.json` simultaneously serves as the input-validation contract, the interactive-prompt definition (via the `x-prompt` extension), and the self-documentation source that Nx Console reads to render help text and autocomplete. **One artifact, three consumers.** That's the model for this system's project-parameter schema — author it once, and derive the wizard's questions, the direct-input validator, and (per Section 7) the completeness rules from that same definition, instead of maintaining three definitions that inevitably drift apart from each other.

Nx generators are also tested against an in-memory virtual file system rather than real disk I/O, keeping the generation logic pure and independently testable. The equivalent discipline here: Layers 1 and 2 (classification, skeleton assembly) should be pure functions of the parameter object — no AI call, no I/O, fully unit-testable with fixtures.

### create-*-app pattern (create-react-app and peers)

Worth naming as the explicit counter-example. These tools deliberately support exactly one archetype and refuse general configurability — "eject" is the escape hatch, not a config flag — and that's a genuinely good design when a tool serves one well-understood use case extremely well. It's the wrong model *here* specifically because the brief calls for a system spanning six named domains plus open-ended unique features: the single-archetype simplicity that makes create-react-app excellent at its one job is the same property that would make a single-archetype version of this generator bad at its job. Ruled out explicitly because "just have great defaults and skip the branching" is a real proposal a stakeholder could reasonably make, and it deserves a direct answer rather than being ignored.

### Synthesis

None of these four tools generate prose that requires judgment — they all stop at exactly the boundary where a human would start writing something domain-specific, because code templates don't need that capability. A session matrix skeleton is a scaffolding problem, shaped like what cookiecutter, Yeoman, and Nx already solve well. A specific, non-generic research prompt is a writing problem, which is outside what any of these four tools attempt. That's why no pure scaffolding-tool architecture is sufficient alone, and why the AI layer in the recommended design isn't an enhancement — it's doing the one part of this job that has no precedent in the systems this framework should otherwise borrow most heavily from.

---

## 4. Recommended Architecture: Layered Hybrid System

```
INPUT: 8 project parameters (name, domain, vision, constraints,
       unique features, platform targets, team size, risk level)

  ┌─────────────────────────────────────────────────────────┐
  │ LAYER 0 · Input Contract & Elicitation                  │
  │ direct params ──┬──►  one parameter schema  ──► spec    │
  │ wizard/interview ┘    (validates + prompts, one source) │
  └─────────────────────────────────────────────────────────┘
                               ▼
  ┌────────────────────────────────────────────────────────┐
  │ LAYER 1 · Classification & Complexity Scoring  [rules] │
  │ domain match + confidence · complexity score · tier    │
  └────────────────────────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────┐
  │ LAYER 2 · Archetype Resolution & Skeleton Assembly  [rules] │
  │ primary (+ secondary blend, or fallback) ──► ordered        │
  │ session skeleton: topic, objective, mandatory flag          │
  └─────────────────────────────────────────────────────────────┘
                               ▼
  ┌────────────────────────────────────────────────────────┐
  │ LAYER 3 · Scoped Prompt Synthesis  [N small AI calls]  │
  │ skeleton slot + spec ──► one ready-to-use research     │
  │ prompt per session; independently retryable / fallback │
  └────────────────────────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────┐
  │ LAYER 4 · Decision Registry Init  [rule cats + AI text] │
  │ archetype hypothesis categories ──► project-specific    │
  │ hypothesis wording per category                         │
  └─────────────────────────────────────────────────────────┘
                               ▼
  ┌────────────────────────────────────────────────────────────┐
  │ LAYER 5 · Validation & QA  [gates, not a final report]     │
  │ schema · completeness · depth budget · optional judge pass │
  │ redundancy check · human gate (high-risk)                  │
  │ ──► usage feedback loop back to Layer 1 / Layer 2          │
  └────────────────────────────────────────────────────────────┘

OUTPUT: session matrix + per-session research prompts +
        seeded decision registry
```

**Layer 0 — Input Contract & Elicitation.** One schema defines all eight parameters: type, validity constraints, and — critically, following the Nx pattern directly — the same schema drives both a direct structured-input path and the interactive wizard's question sequence. There is exactly one definition of what a valid `ProjectSpec` looks like; the wizard is a UI over it, not a parallel system. Output: a validated `ProjectSpec` object.

**Layer 1 — Classification & Complexity Scoring.** Pure, deterministic, no AI call. Matches the `ProjectSpec` against the domain archetype library to produce a primary domain and, where the signal supports it, a secondary domain with a confidence score (Section 5 details the blending and fallback logic). Computes a complexity score from explicit, inspectable weights and maps it to a depth tier (Section 6 gives the full formula). Both outputs — the domain match and the depth tier — carry their reasoning forward; nothing here is a black box.

**Layer 2 — Archetype Resolution & Skeleton Assembly.** Also deterministic. Given the domain match and depth tier, assembles an ordered list of session *skeletons* — topic, one-line objective, and a mandatory/optional flag — by combining the primary archetype's session library with, if applicable, the secondary archetype's differentiator sessions (Section 5), pruned or expanded to match the depth tier's session-count budget. This is Plop's "actions as a function of answers" pattern, generalized: `assembleSkeleton(domainMatch, depthTier) → SessionSkeleton[]`. No prose is written here — every skeleton entry is still a topic and an objective, not a finished prompt.

**Layer 3 — Scoped Prompt Synthesis.** This is where AI enters, and it enters narrowly. Each session skeleton, together with the parts of the `ProjectSpec` relevant to it (never the whole spec dumped in — a competitive-positioning session doesn't need the platform-targets list), goes to a separate, small, focused generation whose only job is to turn that skeleton into one specific, ready-to-use research prompt. This is prompt-chaining applied to the generator's own internals: N focused calls instead of one call carrying the whole task, which is the decomposition that reliability research on multi-instruction generation consistently favors. Each call is independently retryable, independently validatable (Layer 5 checks it as it returns, not after the whole batch completes), and independently replaceable with a template-default fallback if the AI call fails outright — none of which is achievable when one generation produces the entire pipeline.

**Layer 4 — Decision Registry Initialization.** Hybrid in the same spirit as the whole system: which *categories* of hypothesis a project needs is a rule-based property of its archetype (a FinTech pipeline always seeds a regulatory-exposure hypothesis category; a Consumer Mobile pipeline always seeds a retention-loop hypothesis category), while the specific, testable wording of each hypothesis is generated the same scoped way session prompts are in Layer 3 — informed by the project's actual vision and unique features, not a template with the project name dropped in.

**Layer 5 — Validation & QA.** Not a single end-of-pipeline report — gates positioned at the seams between the layers above, following cookiecutter's hook model directly. Full detail in Section 7; briefly: structural schema validation on every AI-generated artifact as it returns from Layer 3 or 4, archetype-completeness checks against Layer 2's mandatory-session list, a depth-budget check that the assembled pipeline's size actually matches its tier, an optional rubric-based quality pass, a redundancy check across sessions, and a human review gate that activates automatically at the high-risk tier. Its outputs also feed a usage-tracking loop back into Layers 1 and 2's weights and archetype definitions — the mechanism that keeps this system evidence-based over time rather than evidence-based only at design time.

### Why this satisfies the framework's own principles

*Evidence-based*, because every structural decision (Layers 1 and 2) is a rule with an inspectable weight or condition behind it, not model discretion, and Section 7's feedback loop means those weights get revised against what actually happens to generated pipelines rather than staying fixed at whatever they were set to at launch. *Adaptive*, because the AI layer absorbs the parts of the problem — free-text vision, unique features, domain blends, novel combinations — that a fixed rule table structurally cannot, without that adaptability contaminating the parts (mandatory sessions, schema conformance) where it shouldn't be allowed to vary. *Non-prescriptive*, because nothing in this architecture is presented as a final answer: the domain match, the depth tier, and every generated artifact carry their rationale and sit one action away from being overridden — detailed concretely in Section 6.

---

## 5. Domain Archetype Library

Each archetype below lists what's *distinctive* about that domain's research needs — not the baseline research every project needs (target users, core value proposition, and so on), which lives in a shared cross-domain session set that every archetype includes regardless of match.

**B2B SaaS.** The buyer and the user are routinely different people, which the research pipeline has to treat as two separate questions rather than one. Distinctive needs: buyer-vs-user decision-path research; sales-assisted vs. self-serve go-to-market motion; the integration and ecosystem surface the product will be expected to plug into (Salesforce, Slack, SSO providers); pricing model research (seat-based vs. usage-based vs. tiered); multi-tenancy and data-isolation architecture; enterprise procurement readiness (SOC 2, SSO/SAML expectations) if the vision implies upmarket motion.

**Developer Tools.** Distribution and first-run experience carry unusual weight relative to other domains. Distinctive needs: developer-experience research focused specifically on time-to-first-value and friction points; distribution-channel research (package managers, IDE marketplaces, GitHub-driven discovery); open-source-vs-commercial licensing strategy; API/SDK design-convention research scoped to the target language ecosystem's idioms; documentation-as-product strategy; competitive research against open-source alternatives specifically, which behave differently from commercial competitors.

**FinTech.** The domain where getting research scope wrong carries the highest downstream cost, so its mandatory-session list is the least negotiable of the six. Distinctive needs: jurisdiction-specific regulatory and licensing landscape (money-transmitter licensing, KYC/AML, PCI-DSS, and region-specific frameworks such as PSD2 where relevant); banking or payment-rail integration approach (card networks, ACH, open-banking APIs); security and fraud architecture research; audit-trail and data-retention requirements; trust and credibility positioning research, which functions differently in FinTech than in most other domains.

**AI/ML.** The domain most likely to be under-scoped by a static template, because what "AI/ML research" means shifts faster here than in any other archetype. Distinctive needs: build-vs-buy model-selection research (foundation model, fine-tune, or custom); data acquisition and labeling strategy; evaluation and benchmarking methodology specific to the task; inference cost and latency architecture; AI-specific UX research covering trust, explainability, and failure-mode handling; responsible-AI and bias research; competitive-landscape research with an explicit recency requirement, since this domain's competitive set moves faster than the pipeline's own shelf life.

**Consumer Mobile.** Distinctive needs: app-store discovery and optimization research; platform review-guideline compliance (Apple and Google review requirements specifically, not generic app-store advice); retention-loop and notification-strategy research; monetization-model research (freemium, subscription, ads, in-app purchase) matched to the stated vision; onboarding-friction research, since mobile drop-off during onboarding is unusually steep relative to other platforms; privacy and tracking-consent research (App Tracking Transparency and equivalent).

**Real-Time/IoT.** Distinctive needs: connectivity and protocol research (MQTT, CoAP, BLE, or whatever the platform targets imply); hardware-constraint research (power, memory, offline operation); latency and reliability architecture, specifically the edge-vs-cloud processing tradeoff; device-fleet management and over-the-air update strategy; security research scoped to embedded/networked attack surfaces, which differ meaningfully from server-side security research; degraded-operation and network-partition handling.

### Archetype blending

Real projects routinely don't sit inside one archetype — an AI-powered lending product is FinTech-primary and AI/ML-secondary. Layer 1's classifier produces a primary domain and, where the signal supports it, a secondary domain with its own confidence score. Layer 2 then composes the primary archetype's *full* session skeleton with only the secondary archetype's *differentiator* sessions (the ones distinctive to that domain, listed above) — not its full skeleton, which would duplicate the shared baseline sessions both archetypes already include. This is Yeoman's `composeWith()` pattern applied to session skeletons rather than file generators: one archetype composing with another, triggered by the classification rather than hand-authored per combination.

### Fallback for unclassified or low-confidence projects

When Layer 1's primary-domain confidence falls below a threshold — a genuinely novel project, or one that doesn't map cleanly to any of the six — the system does not force a best-fit archetype onto it. It falls back to a minimal generic skeleton (the shared cross-domain baseline only) and shifts more weight onto Layer 3's AI synthesis to fill the gap the missing archetype would otherwise have filled, since free-text synthesis is the one part of this architecture that isn't structurally limited to the six named domains. This fallback path is also the system's own signal for when it needs a seventh archetype: if "Custom" triggers often, that's evidence a domain is missing from the library, not proof the fallback is working as intended (revisited in Section 9).

---

## 6. Complexity & Depth Determination

The brief poses this as a choice between explicit user input, complexity scoring, interactive interview, or some combination. The right answer is a specific combination, not an equal blend of all three: **complexity scoring is the default engine, its result is always shown with its reasoning, and the user always has final say — with the interview mode as an alternative way to fill in the same scoring inputs, not a separate depth-setting mechanism.**

Pure explicit input (just ask the user how deep they want the pipeline) puts the burden on exactly the judgment this tool exists to provide — most users asking for a research pipeline generator don't yet know what "right-sized" looks like for their specific combination of domain, risk, and constraints, which is the whole reason to build the scoring logic in the first place. Pure automated scoring, with no visibility or override, violates the non-prescriptive principle outright — an opaque tier assignment is exactly the kind of "black box" decision the framework says to avoid. The combination resolves both problems: scoring does the work, but never silently.

### An illustrative scoring formula

The weights below are a starting point for calibration, not a claim of correctness — a real-world precedent worth following here is a published government project-complexity-and-risk framework that scores multiple weighted dimensions into a composite mapped to oversight tiers, and separately, research on adaptive risk-scoring has specifically flagged static, hand-set weights as a source of inflexibility that degrades accuracy as conditions change. The formula should be config, not a constant baked into code, and Section 7's feedback loop is what keeps it honest.

```
score = 0
score += DOMAIN_INTENSITY[primary_domain] * 3      # 1–3, static per-archetype lookup
score += RISK_WEIGHT[risk_level] * 3               # low=1, medium=2, high=3
score += platform_breadth_score(platform_targets) * 2   # 0–2
score += constraint_density_score(constraints) * 2       # 0–2
score += feature_novelty_score(unique_features) * 2       # 0–2, heuristic or AI-assisted
                                                            # (does this match a known
                                                            # pattern in the domain, or not?)

tier = score_to_tier(score)          # Lite / Standard / Deep / Comprehensive
tier = apply_team_size_ceiling(tier, team_size)   # caps, never raises
return { score, tier, breakdown }     # breakdown always shown to the user
```

Illustrative tiers over a 0–30 range: **Lite** (0–8, three to four sessions), **Standard** (9–16, five to eight sessions), **Deep** (17–23, nine to thirteen sessions), **Comprehensive** (24–30, fourteen-plus sessions, presented as sequential waves rather than one flat list).

**Team size deliberately doesn't add to the score.** A large team doesn't make a domain inherently riskier, and a solo founder doesn't make FinTech's regulatory exposure any smaller — team size measures execution capacity, not problem complexity, and folding it into the same additive score as risk or domain intensity would conflate two different things. Instead it acts as a **ceiling**: a two-person team scoring into Comprehensive gets shown that score honestly, alongside a capped recommendation and an explicit note of why the cap exists, with one click to proceed at full scope anyway. This is the non-prescriptive principle applied concretely — the system states its concern and gets out of the way.

### Where the interactive interview actually helps

Not as a separate depth-selection mechanism, but for two things a static form does poorly: progressively eliciting the same eight parameters for a user who doesn't want to fill in a structured form up front, and resolving genuinely ambiguous or conflicting signals by asking a targeted follow-up — a project marked solo-team but high-risk is worth one clarifying question (is the risk regulatory, technical, or market-driven?) before scoring commits to an interpretation, since the three answers imply different sessions. Because Layer 0's schema is the single source for both the direct-input path and the interview (Section 4, following Nx's pattern), the interview never diverges from what direct input would have produced for the same answers — it's a different route to the same `ProjectSpec`, not a different generation path.

---

## 7. Quality Validation Framework

Validation is distributed across the pipeline as gates, not concentrated in a single pass at the end — following cookiecutter's hook model (Section 3) directly. Five checks, roughly in order of where they sit in the pipeline:

**Structural validation.** Every artifact Layer 3 or Layer 4 returns is schema-checked immediately — required fields present, types correct, no invented fields that don't map to anything downstream. This is standard practice in current structured-generation tooling (schema-constrained decoding, or a validate-and-retry loop that feeds a parsing failure back to the model with instructions to correct it) and it's cheap enough to run on every single generated artifact rather than sampling.

**Completeness validation.** Checked against Layer 2's skeleton, not against the AI output in isolation: every session the archetype marked mandatory is present in the final matrix. This is where a classic failure mode — a FinTech pipeline quietly missing its compliance session because a generation dropped it — gets caught mechanically rather than relying on someone noticing.

**Appropriateness / depth-budget validation.** The assembled pipeline's actual session count is checked against its assigned tier's expected range (Section 6). A Lite-tier project that somehow assembled fourteen sessions, or a Comprehensive-tier project that assembled four, fails this check and triggers a re-assembly at Layer 2 rather than shipping a pipeline that doesn't match the depth it claims.

**Content-quality validation (optional, cost-gated).** A rubric-based pass — not an open-ended "is this good?" judgment, which research on LLM-based evaluation consistently finds less reliable than scoring against explicit, named criteria. The rubric for a generated research prompt: *specificity* (does it engage with what's actually distinctive about this project, or would it read identically for any project in the domain — the direct test for the genericness failure mode pure templates can't avoid), *alignment* (does it address what the skeleton's stated objective for this session actually asks for), and *actionability* (does it produce a concrete research task rather than a restated topic label). Two known biases are worth designing around explicitly rather than discovering later: verbosity bias, where a judge model rewards length independent of quality, so the rubric should score against the three named criteria rather than any holistic or length-correlated impression; and self-evaluation bias, where a model grading its own output tends toward leniency, so the judge pass should run as a distinct, decontextualized call rather than the same call that produced the content marking its own work. Given the added cost and latency, this pass is worth gating to the Deep and Comprehensive tiers and to any high-risk project, rather than running unconditionally.

**Redundancy check.** A lighter check across the assembled session prompts for near-duplicate scope — two sessions that would produce substantially the same research — which can happen at the seam where a blended archetype's differentiator sessions overlap with the primary archetype's baseline.

**Human review gate.** Automatic at the high-risk tier, and available on request otherwise: the generated pipeline is presented for review before being treated as final, with a diff against the relevant archetype's default skeleton so a reviewer can see exactly what was customized rather than re-reading the whole thing from scratch.

**The feedback loop.** Every one of these checks, plus downstream signal about which generated sessions actually got used, skipped, or flagged as unhelpful once a pipeline is in someone's hands, feeds back into Layer 1's scoring weights and Layer 2's archetype definitions. This is what keeps the word "evidence-based" true on an ongoing basis rather than only at design time — the formula in Section 6 is a hypothesis, and this loop is how it gets tested against reality and revised.

---

## 8. Implementation Roadmap

**Phase 1 — Deterministic core.** Build Layers 0–2 only, for two archetypes (B2B SaaS and Developer Tools are reasonable starting points — well-understood domains with clear session needs), with no AI call anywhere yet. The goal is validating that classification, scoring, and skeleton assembly produce sensible session matrices on their own merits before any AI budget is spent on prose — if the skeleton is wrong, fixing it after the AI layer is built means debugging through a nondeterministic layer to find a deterministic bug.

**Phase 2 — AI synthesis and structural validation.** Add Layer 3 (scoped prompt synthesis) and the deterministic parts of Layer 5 — schema, completeness, and depth-budget checks. No rubric-based judge pass yet; the goal here is proving the scoped-call pattern (independent retry, independent fallback) actually behaves the way Section 4 claims before adding the more expensive quality-judgment layer on top.

**Phase 3 — Full domain coverage and blending.** The remaining four archetypes, the blending logic (Section 5), the low-confidence fallback path, and the interactive wizard as an alternative Layer 0 front end (built against the same schema Phase 1 already defined, per Section 4's Nx-derived design).

**Phase 4 — Quality loop closure.** Layer 4's decision registry generation, the optional rubric-based judge pass, the human review gate at the high-risk tier, and the usage-feedback telemetry that closes the loop back into Layer 1 and 2's weights. This is deliberately last: it's the layer that makes the system evidence-based *over time*, and it needs the other layers stable and producing real usage data before recalibration is meaningful.

---

## 9. Risks & Open Questions

**Archetype staleness.** Domain research needs shift — what AI/ML research needed to cover changed substantially even across the last few years — and an archetype library that's treated as finished code rather than versioned, owned data will drift out of date silently. Mitigation is organizational as much as architectural: each archetype needs an owner and a revision cadence, echoing cookiecutter's community-maintained-template model rather than a one-time hard-coded branch.

**AI layer cost at scale.** Decomposing one generation into N scoped calls is well-supported as a reliability improvement, but it's not free — the same research that favors chaining over monolithic prompts also notes it typically increases total token spend and latency. That's a reasonable trade for reliability, but it should be measured, not assumed, and Layer 2's session-count budget per tier is the main lever for keeping a Lite-tier pipeline from paying for AI calls a Comprehensive-tier pipeline needs and it doesn't.

**Overfitting the six archetypes to the projects used to define them.** The six named domains were presumably chosen because they cover the framework's current users, and the fallback path (Section 5) exists so a seventh doesn't hit a wall — but that path needs monitoring, not just building. A fallback that triggers often is a signal a new archetype is needed, not evidence the fallback is doing its job well.

**Non-prescriptive as a UX footgun, not just an architecture concern.** Showing a recommendation with its rationale and an override only functions as genuinely non-prescriptive if people actually use the override. If accepting the default is one click and adjusting it is five, the architecture's intent gets undermined by interface friction that has nothing to do with the architecture itself — worth flagging even though it sits slightly outside this document's scope.

---

## 10. Appendix

### 10.1 Illustrative parameter schema (Layer 0)

```json
{
  "project_name": "string",
  "domain": {
    "primary": "b2b_saas | dev_tools | fintech | ai_ml | consumer_mobile | realtime_iot | custom",
    "secondary": "same enum, optional",
    "confidence": "0.0–1.0, set by Layer 1"
  },
  "vision": "free text",
  "constraints": ["free text", "..."],
  "unique_features": ["free text", "..."],
  "platform_targets": ["ios", "android", "web", "desktop", "embedded", "..."],
  "team_size": "solo | small_2_5 | mid_6_20 | large_20plus",
  "risk_level": "low | medium | high"
}
```

### 10.2 Illustrative FinTech skeleton, Standard tier (Layer 2 output, before Layer 3)

```
FinTech — Standard tier (target: 5–8 sessions), * = mandatory for this archetype
1. * Regulatory & licensing landscape (jurisdiction-specific)
2. * Payment rail / banking integration approach
3. * Security & fraud architecture
4.   Trust & credibility positioning
5.   Pricing & unit economics
   [if secondary domain confidence clears threshold, append that
    archetype's differentiator sessions here before Layer 3 runs]
```

### 10.3 Further reading

Generator architectures referenced in Section 3:
- cookiecutter hooks: https://cookiecutter.readthedocs.io/en/stable/advanced/hooks.html
- cookiecutter core docs: https://cookiecutter.readthedocs.io/en/stable/cookiecutter.html
- Yeoman composability: https://yeoman.io/authoring/composability.html
- Plop documentation: https://plopjs.com/documentation/
- Nx local generators: https://nx.dev/docs/kb/local-generators
- Nx generator options (schema-driven prompts): https://nx.dev/recipes/generators/generator-options

Prompting and evaluation techniques referenced in Sections 2, 4, and 7:
- Meta-prompting overview: https://www.prompthub.us/blog/a-complete-guide-to-meta-prompting
- Prompt chaining technique: https://www.promptingguide.ai/techniques/prompt_chaining
- LLM-as-a-judge methodology: https://langfuse.com/docs/evaluation/evaluation-methods/llm-as-a-judge

Complexity/risk scoring precedent referenced in Section 6:
- Project Complexity and Risk Assessment guide (Canada Treasury Board): https://www.canada.ca/en/treasury-board-secretariat/services/information-technology-project-management/project-management/guide-using-project-complexity-risk-assessment-tool.html
