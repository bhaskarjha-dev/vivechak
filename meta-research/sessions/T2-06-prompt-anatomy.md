# Deep Research Prompt Architecture: An Evidence-Based Audit

*What actually improves output quality in research prompts for frontier deep-research systems — versus what is cargo cult.*

---

## Executive Summary

The evidence converges on a clear pattern, and it isn't "more structure" or "less structure" — it's **structure in the right place and freedom in the right place**, and these are different places for different parts of a research prompt.

Direct answers to the six questions posed:

1. **Which of the 8 sections are empirically justified?** Three are solidly justified (`temporal_enforcement`, `context`, `output_format`), one is justified only in a reframed form (`unbiased_constraint`), and four should be cut or substantially rewritten (`system` persona, `bias_resistance`, `web_searches`, `output_spec`). Full verdicts below.
2. **Exact queries vs. search freedom?** The evidence is one-sided. All three frontier labs' own production systems, plus the foundational agentic-search literature, treat a fixed query list as actively counterproductive for adaptive research. Effort should scale to complexity; queries should not be pre-registered.
3. **Do rigid output sections help or hurt?** Mixed, but leans toward hurt when the sections are mandatory regardless of findings. The fix is a *coverage checklist* rather than a *skeleton*.
4. **What's missing?** Source-quality heuristics, an effort/budget-scaling rule, an explicit assumption-declaration instruction, and a citation-integrity check are all better-evidenced than several elements currently in the template.
5. **Does this vary by model?** Yes — reasoning-native models (current GPT, Claude, and Gemini flagships all now reason internally) specifically penalize procedural over-specification in a way older non-reasoning models didn't. The direction of the effect is consistent across vendors; the magnitude differs.
6. **Minimal effective anatomy?** Five blocks, not eight — detailed in the recommendations section, with a worked example.

The "competing philosophy" described in the brief — define WHAT and WHY, leave HOW open — is **well supported**, but not because structure is bad in general. It's because three *specific* sections of the current template (`web_searches`, `bias_resistance`, and the rigid parts of `output_spec`) encode *procedure* rather than *goal*, and procedural over-specification is exactly the category the evidence says backfires. The sections that encode context, scope, and intent are under-weighted by comparison, and the evidence says those should stay or grow.

---

## Evidence Base

This audit draws on three kinds of sources, weighted differently:

- **Primary lab documentation** (highest weight): Anthropic's engineering write-up on its production multi-agent Research system, OpenAI's Deep Research API documentation and reasoning-model prompting guides, and Google's Gemini Deep Research/Deep Research Max documentation. These describe what three competing labs, each running large-scale internal evals, actually shipped — a form of revealed preference backed by real evaluation data, even where the underlying ablations aren't public.
- **Peer-reviewed and preprint research** (high weight for the specific claim tested, lower for generalization): persona-prompting studies, the format-restriction study, context-length degradation research, negation/suppression studies, and the foundational chain-of-thought and agentic-search papers. Most of these test single-turn QA or classification, not long-horizon report writing, so extrapolation to "deep research" specifically is reasonable but not proven.
- **Industry analysis and secondary reporting** (used only to corroborate, never as sole support): benchmark write-ups and practitioner analyses, cited only where a claim is either independently corroborated by multiple outlets or explicitly flagged as a single-source data point.

Where a verdict rests on inference rather than a direct ablation of "this exact prompt section," that's stated explicitly rather than implied.

---

## The Framework: Three Kinds of "Structure"

The current 8-section template treats "structure" as one thing, and the debate in the brief treats "prescriptive vs. directional" as one axis. Neither holds up. The evidence separates cleanly into three categories of structure with different risk profiles:

| Type | What it does | Evidence on risk | Verdict |
|---|---|---|---|
| **Framing structure** — delimiters and labels that separate context from instructions from constraints (e.g., XML tags, section headers organizing *input*) | Organizes what the model is looking at; doesn't tell it how to think | Low risk. OpenAI's own reasoning-model guidance recommends delimiters *even while telling developers to strip out chain-of-thought prescription* — clarity of input and freedom of process are treated as separable, not opposed (OpenAI, Reasoning Best Practices) | Keep |
| **Procedural structure** — mandating the exact steps, exact search queries, or exact reasoning order the model must follow | Dictates *how* the model works through the problem | High risk for reasoning-heavy and agentic tasks. This is the category the format-restriction literature and reasoning-model guidance both flag (Tam et al., 2024; OpenAI Reasoning Best Practices) | Cut, or convert to heuristics |
| **Output-shape structure** — mandating the exact skeleton of the *final deliverable*, independent of what was found | Dictates what the answer must look like, applied after the thinking is largely done | Moderate risk, and rigidity-dependent. A *required-coverage checklist* is well supported; a *mandatory fixed skeleton* is not | Replace skeleton with checklist |

This distinction is what makes the section-by-section verdicts below coherent rather than a list of unrelated yes/no calls: sections built from framing structure survive the audit, sections built from procedural structure don't, and output-shape structure needs to be loosened rather than removed.

---

## Section-by-Section Audit

### 1. `<system>` — Expert persona / role assignment — **Cut**

This is the element with the most direct, most recent, and most convergent counter-evidence of the eight.

Zheng et al. (2024, EMNLP Findings) tested 162 personas across 2,410 factual questions and four LLM families and found that **adding a persona to the system prompt does not improve performance over no persona at all**; where an effect exists, it's driven by essentially random interaction between persona and question rather than genuine expertise transfer. A 2026 University of Southern California preprint (Hu, Rostami & Thomason) went further: expert personas *reduced* factual accuracy on a standard knowledge benchmark, from a 71.6% baseline to roughly 66%, with the effect worsening as the persona description got more elaborate. Their hypothesis — replicated in spirit by a Wharton Generative AI Labs study using GPQA Diamond and MMLU-Pro (Basil, Shapiro, Shapiro, Mollick, Mollick & Meincke) — is that a persona activates instruction-following and stylistic behavior at the expense of the factual-recall pathway. The Wharton study additionally found that *mismatched* personas (wrong domain) sometimes hurt further, and that low-knowledge personas ("you are a layperson") reliably hurt.

The common thread across all three studies: personas can shift *tone and alignment with human stylistic expectations* — which is a real, separate effect — but they don't improve, and often measurably damage, accuracy on **knowledge-heavy tasks**. A research report is precisely a knowledge-heavy task. Whatever benefit `<system>` was meant to provide (depth, rigor, "sounding like an analyst") is better achieved by stating those as explicit output requirements — "written for a decision-maker who needs defensible figures, not a general-audience explainer" — than by asking the model to role-play. That framing is a `context`/`output_spec` concern, not a persona concern, and folding it there also avoids the accuracy cost.

### 2. `<temporal_enforcement>` — Current date + recency verification — **Keep**

This is the one section in the template that is close to logically necessary rather than merely empirically nice-to-have, so the case for it rests less on ablation studies and more on structural reasoning — worth flagging honestly, since it's a different kind of justification than most of the other verdicts here. A model's parametric knowledge has a training cutoff; a search-augmented agent without explicit temporal grounding has no reliable way to know whether a retrieved page is current, superseded, or describing a since-reversed decision, and will default to training-era assumptions about what's true "now." Every production deep-research system reviewed here (Anthropic's Research feature, OpenAI's `o3`/`o4-mini`-deep-research models, Gemini Deep Research) is built around live, dated tool calls specifically because this can't be solved from parametric knowledge alone. The instruction to anchor on the current date and to verify recency isn't decoration — it's the thing that keeps a search-grounded system from confidently reporting stale information as current. Keep it, but keep it short: a stated date plus an instruction to prefer and flag recency is sufficient; it doesn't need elaboration.

### 3. `<context>` — Project background and constraints — **Keep, and this is the section to expand, not shrink**

This is the single most consistently supported element in the entire template, across every source examined.

OpenAI's own production pipeline for ChatGPT's Deep Research is built *entirely* around this: before the actual research model ever runs, an intermediate model runs a **clarification** pass and then a **prompt-rewriting** pass whose explicit job is to maximize context — stated preferences, constraints, sources to prioritize, audience, desired depth — while being equally explicit that *unstated* dimensions should be marked open-ended rather than invented (OpenAI, Deep Research API guide). The Responses API version of deep research skips this only because it "expects fully-formed prompts up front and will not ask for additional context... it simply starts researching based on the input it receives" — which is itself evidence that the context has to come from somewhere, and if the model can't ask for it, the prompt has to supply it. Anthropic's equivalent finding, from watching its production agents fail: giving subagents a "simple, short instruction like 'research the semiconductor shortage'" reliably caused **duplicated work and misinterpreted tasks** — one subagent investigated the 2021 chip crisis while two others redundantly covered 2025 supply chains — until the lead agent was taught to give each subagent "an objective, an output format, guidance on the tools and sources to use, and clear task boundaries" (Anthropic Engineering, *How We Built Our Multi-Agent Research System*). That's context, not persona, and not a fixed procedure. The industry term of art that emerged from this same period — "context engineering," popularized around the same June 2025 window by Cognition's "Don't Build Multi-Agents" post and adopted by Anthropic's own subsequent engineering post on the subject — exists precisely because labs converged on context curation, not prompt phrasing or role assignment, as the dominant lever for agent quality.

### 4. `<unbiased_constraint>` — Catalog without premature filtering — **Keep, but reframe as a sequencing instruction, not a debiasing label**

This one survives, but for a narrower reason than its name suggests. The instruction "catalog broadly before filtering" is a *procedural sequencing* instruction (do exploration, then do evaluation), not an *abstract debiasing* instruction, and the evidence treats those very differently.

Two independent lines of support. First, Anthropic found its production agents' default failure mode was the opposite of this instruction: agents "often default to overly long, specific queries that return few results," and had to be explicitly prompted to "start with short, broad queries, evaluate what's available, then progressively narrow focus" — mirroring, in their words, how expert human researchers work. Second, at the reasoning-strategy level, Wang et al.'s self-consistency method (ICLR 2023) — one of the most cited results in LLM reasoning — shows that sampling multiple diverse reasoning paths *before* converging on an answer substantially outperforms committing early to a single greedy path (+17.9% on GSM8K, +12.2% on AQuA, and consistent gains across other reasoning benchmarks). Premature convergence is a real, measurable failure mode, and "explore before you commit" measurably fixes it. Keep this instruction, but frame it as *when to filter*, not as a claim about bias per se — that distinction matters for how it interacts with the next section.

### 5. `<bias_resistance>` — Naming a specific cognitive bias to resist — **Cut, or fold into #4 as a positive instruction**

This is the second-clearest cut in the template, and the mechanism is specific enough to be worth explaining rather than just asserted.

Telling a model to avoid a *named* bias is a negation instruction ("don't let confirmation bias affect this"), and negation instructions in LLMs have a documented failure mode closely analogous to ironic process theory in human cognition — the "don't think of a white bear" effect. A November 2025 study built specifically to test this in transformers ("Don't Think of the White Bear: Ironic Negation in Transformer Models Under Cognitive Load") found that LLMs show measurable **ironic rebound**: instructing a model not to produce a concept can paradoxically raise that concept's activation and make it more likely to surface, an effect that intensifies under longer or more semantically loaded prompts — exactly the conditions a multi-section research template creates. A related 2024 paper ("Suppressing Pink Elephants with Direct Principle Feedback") documents the same pattern and calls it, aptly, the Pink Elephant effect. Separately, the cognitive-debiasing literature draws a sharp distinction between a bare "avoid bias X" statement — which correlates weakly with actual debiasing and scales unreliably with model capability — and a *structured* debiasing process (determine, analyze, correct), which reliably outperforms the bare statement (Lyu et al., *Cognitive Debiasing Large Language Models for Decision-Making*). A one-line named-bias instruction is the weaker of the two forms, and it has a plausible specific downside beyond just "not helping."

The fix isn't to drop bias mitigation — it's to drop the *naming*. Fold the underlying intent into `<unbiased_constraint>` as positive procedural guidance: "before drawing a conclusion, note what evidence would change it" or "surface disagreements between sources rather than resolving them silently" are actionable, bias-resistant instructions that don't require naming and activating the concept you're trying to suppress.

### 6. `<web_searches>` — 10–14 exact search queries — **Cut. This is the single largest deviation from evidence in the current template.**

Every source examined — from the foundational agentic-search paradigm to all three frontier labs' shipped systems — points the same direction.

The theoretical foundation for modern search agents is the ReAct paradigm (Yao et al., 2022/2023): an interleaved *reason → act → observe* loop where each action is chosen in light of what the *previous* action returned. A pre-registered list of 10–14 queries is structurally incompatible with this loop — it cannot incorporate the "observe" step, because every query is fixed before the first search result is ever seen. This isn't a minor inefficiency; it's the difference in kind between an adaptive agent and a static pipeline, and the field's own survey literature frames it exactly this way, describing "dynamic" workflows as offering "significantly greater autonomy, continual and deep reasoning... and adaptive real-time interaction" precisely *in contrast to* systems that "heavily depend on pre-defined workflows" (Huang et al., *Deep Research Agents: A Systematic Examination and Roadmap*).

In production, this plays out concretely. Anthropic's own agents' default failure mode was *the opposite* of what a 10–14-query list would encode — agents defaulted to "overly long, specific queries that return few results" and had to be prompted toward starting broad and narrowing progressively, a strategy that cannot be specified as a fixed list because the right second query depends on what the first one returned. Rather than a query count, Anthropic embedded **effort-scaling rules keyed to task complexity**: simple fact-finding gets one agent and 3–10 tool calls; direct comparisons get 2–4 subagents at 10–15 calls each; complex, breadth-first research gets 10+ subagents with divided responsibilities. Critically, Anthropic's own analysis of the BrowseComp benchmark found that **three factors explain 95% of performance variance — token usage alone explains 80% of it, with tool-call count and model choice explaining the rest.** Quality tracks a *budget matched to complexity*, not adherence to a specific query list. OpenAI's deep-research API exposes the same idea from the other direction: `max_tool_calls` is described as "the primary tool available to you to constrain cost and latency," i.e., a ceiling, not a script. Gemini Deep Research resolves the same tension with a third mechanism — the agent proposes its own multi-step plan and the user can edit it before autonomous execution begins, which preserves user control without the plan being written by the user in advance.

The recommended replacement isn't "no guidance on search" — it's guidance on *strategy and budget* (start broad, narrow progressively, scale tool calls to complexity, stop when the marginal search stops changing the answer) rather than a *content* prescription (these specific 10–14 strings). If example queries are useful at all, they belong as a hint of the expected granularity, explicitly marked as illustrative rather than exhaustive — not as a checklist to execute.

### 7. `<output_spec>` — Detailed section-by-section output structure — **Modify: replace the mandatory skeleton with a required-coverage checklist**

The clearest direct evidence here is a 2024 EMNLP Industry Track paper, "Let Me Speak Freely? A Study on the Impact of Format Restrictions on Performance of Large Language Models" (Tam et al.), which found a **significant, and stricter-is-worse, decline in LLM reasoning ability when generation is constrained to a structured format versus left free-form.** One important caveat, in the interest of not overclaiming: that paper's strongest evidence concerns machine-parseable formats (JSON/XML) imposed *during* the token-by-token reasoning process, which is a stronger constraint than "organize the final written report under three broad headings." The two aren't identical, but the mechanism generalizes partially: forcing content into a predetermined shape while it's still being figured out competes with the reasoning itself, whereas applying structure to a report *after* the underlying analysis is largely settled is cheaper. This is exactly why Anthropic's system routes structuring to a separate downstream pass (a dedicated CitationAgent that runs after research is complete) rather than baking a rigid skeleton into the live research prompt, and why OpenAI's prompt-rewriting guidance is conditional rather than universal: *"if the user is asking for content that would be best returned in a structured format... ask the researcher to format as a report with appropriate headers,"* and separately, tables are requested only "if you determine that including a table will help illustrate, organize, or enhance the information" — not by default.

The practical failure mode of a rigid mandatory skeleton is Procrustean: findings that don't map onto the five preset sections get force-fit or padded, and genuinely emergent structure (a theme the research surfaced that the template's authors didn't anticipate) has nowhere to go. The fix that the evidence actually supports is a **coverage checklist** — name the questions that must be answered and the elements that must be present (a comparison table where genuinely comparative, inline citations, explicit flagging of unresolved contradictions) — while leaving the organizing structure to be discovered from what was actually found. This gets the consistency benefit (nothing important gets skipped) without the Procrustean cost (nothing gets force-fit).

### 8. `<output_format>` — Markdown / artifact formatting requirements — **Keep as-is**

This is a presentation-layer requirement, not a reasoning-quality lever, and the evidence above doesn't really bear on it either way — it's orthogonal to the prescriptive/directional debate rather than a data point in it. Specifying markdown, artifact packaging, or similar rendering requirements doesn't constrain *how* the model researches or *what* it concludes; it only affects how the already-finished answer is packaged for delivery. Keep it, but keep it minimal and mechanical — this is the one section where more specification carries essentially no risk, precisely because it sits downstream of everything that determines quality.

---

## The Prescriptive–Directional Spectrum: What the Evidence Actually Shows

The brief poses this as roughly binary — prescriptive templates vs. lean directional prompts. The evidence supports a more specific claim: **the optimal amount of specification is not constant across a prompt; it's high at the goal/context layer and low at the execution/procedure layer, with output-shape somewhere in between.**

The strongest single piece of evidence for this is convergence: three competing labs, running their own large-scale internal evals with no incentive to agree with each other, independently arrived at architectures that specify goal and constraints heavily while leaving search execution almost entirely to the model:

- **Anthropic** explicitly frames this as a design choice: *"Our prompting strategy focuses on instilling good heuristics rather than rigid rules."* Heuristics (start broad, scale effort, evaluate source quality) govern behavior; nothing in the published account resembles a fixed query list.
- **OpenAI**'s own forum guidance states the tradeoff almost as a punchline: *"clearly state your research objectives while allowing the model creative freedom to find the best solution... though counterintuitive, [this] enhances the model's performance by leveraging its inherent strengths"* — offered as a correction to the instinct (which the same practitioner describes having had) to write 1,000-word, maximally detailed meta-prompts.
- **Google**'s Gemini Deep Research resolves the tension procedurally rather than philosophically: the model proposes the plan, the person edits it, then execution is autonomous. Direction comes from the person; method comes from the model, checked but not authored by the person.

This isn't just three products converging by coincidence — it's the same underlying finding restated three ways: once a system is built around an adaptive search loop (which all three are, following the ReAct paradigm), a pre-specified procedure is not merely unnecessary but actively works against the architecture, because the architecture's entire value proposition is the ability to change course based on what it finds. A 90.2% improvement of Anthropic's multi-agent system over its own single-agent baseline was concentrated specifically in **breadth-first tasks** — the category where not knowing in advance what to search for is the whole point.

Where the evidence pushes back against the "just leave it open" reading of the competing philosophy: none of the three labs' systems is actually low-specification at the *context* layer. OpenAI runs two full model calls (clarify, then rewrite) before research even starts, specifically to maximize the amount of stated context and constraint. That is real, deliberate over-specification of goal and scope — it's just not procedural over-specification of method. The "paint by numbers" risk the competing philosophy warns about is real, but it's a risk of over-specifying *how*, not of over-specifying *what* and *why*. Collapsing those into one "structure vs. freedom" axis, as the current template implicitly does by treating `context` and `web_searches` as peers, is the actual design error worth fixing.

---

## Does This Vary by Model?

Yes, in a way that matters for how the template should be written, though the *direction* of the effect is consistent — only the intensity differs.

| Model family | Documented behavior relevant to prompt structure | Source |
|---|---|---|
| **OpenAI (`o3`/`o4-mini`-deep-research, GPT-5.x reasoning models)** | Explicit guidance to avoid chain-of-thought prescription ("think step by step" is called "unnecessary" and can hinder a model that already reasons internally); recommends delimiters for *input* clarity while cutting *process* prescription; internal coding-agent evals reportedly showed leaner system prompts improving evaluation scores by ~10–15% while cutting tokens 41–66% and cost 33–67% — OpenAI's own guidance labels these figures directional and workload-dependent, not universal | OpenAI, *Reasoning Best Practices*; OpenAI, *Model Guidance* (primary source, hedged by OpenAI itself) |
| **Anthropic (Claude with extended/interleaved thinking)** | Extended thinking is explicitly used as a "controllable scratchpad" for the lead agent to plan tool use and subagent count *before* acting, and subagents use interleaved thinking to evaluate results and adjust their next query — i.e., structure is provided *around* the thinking process (clear task boundaries, defined roles) rather than *inside* it | Anthropic Engineering, *How We Built Our Multi-Agent Research System* |
| **Gemini (Deep Research / Deep Research Max)** | Uses an explicit, user-editable planning stage as its mechanism for reconciling user direction with model autonomy, rather than either a fixed procedure or a fully blank check | Google, *Gemini Deep Research* / *Deep Research Max* documentation |

The common thread: as models have moved to reasoning-native architectures (all three vendors' current flagships), the cost of procedural prescription has gone up, because prescribing steps to a model that already plans internally either duplicates that planning in the visible response (wasted tokens) or actively overrides it with a worse, human-authored plan. This is a meaningfully different regime than the GPT-4-class, non-reasoning models most "10-step prompt template" folklore was written for. A template built for those models will over-specify procedure by current standards even if it wasn't over-specified when written.

One caveat on precision here: claims about *which specific formatting dialect* each current model family rewards (e.g., XML-tagged content for one vendor, JSON schemas for another) come from practitioner synthesis rather than each lab's own controlled ablation, and should be treated as directionally useful rather than precisely verified.

---

## Does This Vary by Task Complexity?

Yes, and the same sources that justify cutting `web_searches` as a fixed list also supply a usable decision rule, because they were built to answer exactly this question operationally.

| Task shape | Recommended posture | Why |
|---|---|---|
| **Narrow, single-fact lookup** ("What is X's current market share?") | Tighter is fine — low ambiguity means extra specification has little to constrain away. Low effort budget (Anthropic: ~1 agent, 3–10 tool calls) | There's little "unexpected" left to discover; the cost of prescription is close to zero because there's not much divergent path to lose |
| **Structured comparison** ("Compare X, Y, and Z on criteria A, B, C") | Moderate specification of the *comparison dimensions* (context layer) + moderate effort budget (Anthropic: 2–4 subagents, 10–15 calls each); still no fixed query list | The dimensions of comparison are worth pinning down explicitly (this is context, not procedure); the search path to fill them in is not |
| **Open-ended, breadth-first exploration** ("What are the emerging risks in X industry we haven't considered?") | Maximum context/goal specification, minimum procedural specification, largest effort budget (Anthropic: 10+ subagents, clearly divided) | This is exactly the category where Anthropic's multi-agent architecture beat single-agent by 90.2% — the entire value is discovering what wasn't anticipated, which a pre-specified procedure forecloses by construction |
| **High-stakes / low-tolerance-for-error** (legal, medical, financial decisions riding on the output) | Add an explicit verification/citation-integrity requirement (see Gaps, below) regardless of the above | Citation and factual-accuracy failure rates in current systems are non-trivial even on default settings (see below) — stakes should raise the verification bar, not the procedural-prescription bar |

The template as given makes no distinction along this axis — it applies the same 8 sections regardless of whether the underlying query is a single-fact lookup or an open-ended exploration. That's arguably a bigger issue than any single section's wording: **a template with no complexity-sensitivity will always be either over-specified for simple queries or under-specified for open-ended ones.**

---

## Gaps: What's Missing From the Current Anatomy

Several elements with real evidentiary support are absent from the current 8-section template, or are present only implicitly:

**Source-quality heuristics.** Anthropic's account is specific and instructive: human evaluators (not automated evals) caught that early production agents "consistently chose SEO-optimized content farms over authoritative but less highly-ranked sources like academic PDFs or personal blogs," and this was fixed only after source-quality heuristics were explicitly added to the prompt. OpenAI's prompt-rewriting guidance independently converges on the same fix: prefer primary/official sources over aggregators, original papers over survey summaries. This is currently nowhere in the 8-section template and shouldn't be assumed to be covered by `unbiased_constraint` — source ranking and premature filtering are different failure modes with different fixes.

**An effort/budget-scaling rule.** The natural replacement for the fixed query count isn't "no guidance" — it's a stated rule that ties effort to complexity, of the kind Anthropic embeds directly (roughly: simple → 1 agent/few calls; comparative → several subagents/moderate calls; open-ended → many subagents/high call budget). Without this, a model has no way to judge "have I done enough," which was Anthropic's own named failure mode before the rule was added (agents either stopped too early or kept searching after sufficient results were already in hand).

**Explicit assumption-declaration for unstated dimensions.** OpenAI's prompt-rewriting guidelines contain a specific, well-reasoned instruction that has no analogue in the current template: fill unstated-but-necessary dimensions in as *explicitly open-ended*, and never silently invent a constraint the user didn't give. A single-shot deep-research prompt (as this template is) doesn't get the benefit of OpenAI's separate clarification turn, which makes this instruction more important here, not less — the model needs to be told to flag its own assumptions rather than quietly resolve ambiguity in whatever direction seems reasonable.

**A citation-integrity / verification step.** This is the best-evidenced gap. Independent benchmarking puts citation accuracy for current commercial deep-research systems at roughly 78–94% depending on system (DRBench's FACT framework, as reported in recent citation-hallucination research), and a separate production-usage benchmark found the best-performing system achieving only 65% citation quality and 68% factual accuracy on realistic tasks, with law and medicine — the highest-stakes domains — receiving the harshest penalties. This is consistent with why Anthropic routes final output through a dedicated CitationAgent rather than trusting the research agent's own citations, and it's a real limitation of self-checking: a model reflecting on its own draft cannot detect an error that originates from its own knowledge gap, because it has no independent ground truth to check against. A prompt-level instruction can't fully solve this, but a cheap, honest version is achievable: require the model to flag single-sourced or low-confidence claims explicitly rather than presenting everything with uniform confidence. That's a coverage requirement, not a procedural one, so it belongs in a revised `output_spec`, not as a new heavyweight section.

---

## Recommended Minimal Anatomy

Collapsing the audit above into a working template: five blocks instead of eight, each justified by a specific piece of evidence rather than by convention.

| New block | Replaces | Core content | Why it survives the audit |
|---|---|---|---|
| **`brief`** | `system` + top of `context` | Goal, deliverable, audience, decision the output will inform, expected depth/rigor | Context/goal specification is the most consistently supported lever across all sources; persona is dropped, its intended function (calibrating depth and tone) is achieved directly instead |
| **`scope`** | Rest of `context` + source-quality guidance (new) | Boundaries (time window, domain, in/out of scope), source-type preferences and exclusions, any internal data available, current date | Context engineering literature and OpenAI's rewriting guidelines both treat this as the highest-leverage section; folds in the source-quality gap identified above |
| **`approach`** | `unbiased_constraint` + `bias_resistance` + `web_searches` | Heuristics only: explore broadly before narrowing; scale effort to complexity (state the budget rule); state assumptions explicitly rather than inventing constraints; surface contradictions between sources rather than silently resolving them. No fixed query list, no named bias | This is where the strongest and most convergent evidence sits — replacing procedural prescription with heuristics matches all three labs' production systems and the ReAct paradigm underlying adaptive search |
| **`deliverable`** | `output_spec` | A required-coverage checklist (questions that must be answered, elements that must appear — comparison tables where genuinely comparative, inline citations, explicit confidence flags on single-sourced claims), with organizing structure left to emerge from findings | Converts a Procrustean skeleton into a checklist, consistent with the format-restriction evidence and OpenAI's own conditional ("if it would help") formatting guidance |
| **`format`** | `output_format` | Markdown/artifact packaging mechanics | Presentation-layer only; unchanged because the evidence doesn't implicate it either way |

A worked example, condensed:

```
BRIEF: Assess whether mid-market SaaS companies (50-500 employees) are
adopting usage-based pricing over seat-based pricing, and why. This will
inform a pricing-strategy recommendation for our board; assume a technical
but non-specialist reader who needs defensible figures, not a primer.

SCOPE: Focus on the last 24 months. Prioritize primary sources (company
pricing pages, earnings calls, analyst reports from Gartner/Forrester-tier
firms) over aggregator blogs. Today's date is [DATE] — flag anything you
can't confirm is still current.

APPROACH: Start broad to map the landscape before narrowing to specific
vendors. Scale your effort to what you find — a handful of searches may
settle a simple sub-question, but trace genuine disagreement across
sources rather than picking a side after one query. State explicitly
where you've had to assume something I didn't specify. Note contradictions
between sources rather than silently resolving them in the write-up.

DELIVERABLE: Must address: (1) prevalence of the shift with figures where
available, (2) the 2-3 most-cited reasons driving it, (3) at least one
counter-example of a company moving the other direction. Use a comparison
table if you find enough companies with comparable data; cite inline;
flag any claim resting on a single source.

FORMAT: Markdown, headers as needed for what you found.
```

This is shorter than the current 8-section template, not because less thought went into it, but because the thought is concentrated where the evidence says it pays off — goal, scope, and required coverage — and removed from where the evidence says it doesn't — a pre-written search script and a mandatory skeleton.

---

## Migration Reference

| Current section | Disposition | Goes to |
|---|---|---|
| `<system>` | Cut | Intent absorbed into `brief` (stated depth/audience, not role-play) |
| `<temporal_enforcement>` | Keep | `scope` |
| `<context>` | Keep, expand | Split across `brief` and `scope` |
| `<unbiased_constraint>` | Keep, reframe | `approach`, as sequencing not labeling |
| `<bias_resistance>` | Cut | Intent folded into `approach` as positive instruction |
| `<web_searches>` | Cut | Replaced by heuristics + effort-budget rule in `approach` |
| `<output_spec>` | Modify | `deliverable`, as coverage checklist not skeleton |
| `<output_format>` | Keep | `format`, unchanged |
| *(new)* source-quality guidance | Add | `scope` |
| *(new)* assumption-declaration | Add | `approach` |
| *(new)* citation-confidence flagging | Add | `deliverable` |

---

## Limitations of This Analysis

Honest caveats, in the interest of the same evidentiary standard applied throughout:

- Most of the strongest controlled evidence (persona studies, the format-restriction study) tests single-turn QA or classification benchmarks, not long-horizon multi-step report generation. The extrapolation to deep-research tasks specifically is reasonable — the mechanisms (instruction-following competing with recall; generation constraints competing with reasoning) are general — but it is extrapolation, not a direct test of "deep research prompt with `<system>` tag" vs. without.
- The frontier-lab evidence is revealed preference from shipped products backed by internal evals, not published ablations isolating prompt structure from every other system-engineering choice (tool design, orchestration, RL training) those labs also made simultaneously. The convergence across three independent labs is meaningful, but it's suggestive rather than a controlled proof that prompt structure alone, holding everything else constant, caused the results.
- A few figures here (the OpenAI internal token/eval-improvement percentages, the per-model "formatting dialect" claims) rest on sources with limited independent cross-verification and are flagged as such at point of use rather than presented with false precision.
- This field changes fast — model releases, vendor guidance, and benchmark numbers cited here are current as of this research but should be expected to shift within months, particularly the specific percentage figures.

---

## References

**Primary lab sources**
- Anthropic Engineering. "How we built our multi-agent research system." June 13, 2025. https://www.anthropic.com/engineering/multi-agent-research-system
- Anthropic Engineering. "Effective context engineering for AI agents." https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents
- OpenAI. "Deep research | OpenAI API." https://developers.openai.com/api/docs/guides/deep-research
- OpenAI. "Reasoning best practices | OpenAI API." https://developers.openai.com/api/docs/guides/reasoning-best-practices
- OpenAI. "Model guidance | OpenAI API." https://developers.openai.com/api/docs/guides/latest-model
- Google. "Gemini Deep Research agent | Gemini API." https://ai.google.dev/gemini-api/docs/deep-research
- Google. "Deep Research Max: a step change for autonomous research agents." https://blog.google/innovation-and-ai/models-and-research/gemini-models/next-generation-gemini-deep-research/
- Google. "Gemini Deep Research — your personal research assistant." https://gemini.google/overview/deep-research/

**Peer-reviewed / preprint research**
- Zheng, M., Pei, J., Logeswaran, L., Lee, M., & Jurgens, D. (2024). "When 'A Helpful Assistant' Is Not Really Helpful: Personas in System Prompts Do Not Improve Performances of Large Language Models." *Findings of EMNLP 2024*. https://arxiv.org/abs/2311.10054
- Hu, Z., Rostami, M., & Thomason, J. (2026). "Expert Personas Improve LLM Alignment but Damage Accuracy: Bootstrapping Intent-Based Persona Routing with PRISM." University of Southern California preprint.
- Basil, S., Shapiro, I., Shapiro, D., Mollick, E., Mollick, L., & Meincke, L. "Prompting Science Report 4: Playing Pretend: Expert Personas Don't Improve Factual Accuracy." Generative AI Labs, Wharton School. https://arxiv.org/pdf/2512.05858
- Tam, Z.R., Wu, C-K., Tsai, Y-L., Lin, C-Y., Lee, H., & Chen, Y-N. (2024). "Let Me Speak Freely? A Study on the Impact of Format Restrictions on Performance of Large Language Models." *EMNLP 2024 Industry Track*. https://arxiv.org/abs/2408.02442
- Chroma Research. "Context Rot: How Increasing Input Tokens Impacts LLM Performance." (2025). https://www.trychroma.com/research/context-rot
- Liu, N.F. et al. (2024). "Lost in the Middle: How Language Models Use Long Contexts." *TACL*.
- "Don't Think of the White Bear: Ironic Negation in Transformer Models Under Cognitive Load." (2025). https://arxiv.org/abs/2511.12381
- "Suppressing Pink Elephants with Direct Principle Feedback." (2024). https://arxiv.org/abs/2402.07896
- Lyu, Y. et al. "Cognitive Debiasing Large Language Models for Decision-Making."
- Wang, X. et al. (2023). "Self-Consistency Improves Chain of Thought Reasoning in Language Models." *ICLR 2023*. https://arxiv.org/abs/2203.11171
- Yao, S. et al. (2023). "ReAct: Synergizing Reasoning and Acting in Language Models." *ICLR 2023*. https://arxiv.org/abs/2210.03629
- Huang, Y. et al. "Deep Research Agents: A Systematic Examination and Roadmap." (2025). https://arxiv.org/abs/2506.18096
- "Detecting and Correcting Reference Hallucinations in Commercial LLMs and Deep Research Agents" (cites DRBench [Du et al., 2025] and DRACO [Zhong et al., 2026] benchmark figures). https://arxiv.org/pdf/2604.03173

**Corroborating industry analysis** (used only where independently corroborated across multiple outlets)
- The Register, Search Engine Journal, and ALM Corp coverage of the USC persona-accuracy preprint (March 2026).
- OpenAI Forum. "Exploring Deep Research: Three Tips for Better AI-Assisted Inquiry." April 2025.
- LangChain Blog. "How and when to build multi-agent systems" (discussing Cognition's "Don't Build Multi-Agents" and the origin of "context engineering" as a term).
