# AI Deep Research Systems: A Technical Capability Analysis

**Scope:** Gemini Deep Research, ChatGPT/OpenAI Deep Research, Claude (extended thinking + web search / Research), and Perplexity Research
**Prepared:** August 2026 | **Purpose:** Ground-truth reference for a meta-framework using AI deep research as its primary research engine

---

## A note on method and bias

This report is built from vendor engineering documentation, model/system cards, independent academic benchmarks, and third-party empirical evaluations gathered via live web search — not from training-data memory, which would be stale for a category that has changed substantially in the last twelve months. Every quantitative claim below is attributed to a source category (vendor primary documentation, independent academic benchmark, or third-party empirical test) so you can weight it accordingly.

Two honesty notes up front, because they shape how to read everything that follows:

1. **This report was written by Claude, one of the four systems under review.** Where claims concern Claude specifically, they are sourced to Anthropic's own published engineering writeups and to independent third-party benchmarks, not to self-assessment — and Claude's documented failure modes are reported with the same weight as the other three systems'. Where a claim about Claude could not be corroborated outside Anthropic's own materials, that is flagged.
2. **This category churns fast.** In the course of this research, model IDs retired mid-project, a benchmark leaderboard shifted, and one vendor renamed its product twice in a year. Treat specific percentages and search-count figures as *directionally* accurate snapshots, not fixed specifications. A "verified vs. inferred" ledger is provided in Part 6.

---

## Executive Summary

- **Architecturally, these are not four versions of the same thing.** OpenAI, Google, and Perplexity each ship a model that was specifically trained (via reinforcement learning on browsing/search trajectories) to do research. Anthropic's Research feature, by contrast, is a *prompted multi-agent orchestration* built on top of general-purpose Claude models — there is no dedicated "Claude Deep Research" model, only an engineered system of standard models with tools, extended thinking, and a lead-agent/subagent architecture. This is a real, verifiable design difference, not just branding.
- **Search volume is documented with real precision for exactly one vendor: Google.** Google's own API docs state Gemini Deep Research runs ~80 search queries per typical task (~160 for "Deep Research Max"). Anthropic publishes effort *heuristics* (1 agent/3–10 calls for simple queries, up to 10+ subagents for complex ones) rather than a fixed count. OpenAI and Perplexity describe volume qualitatively ("dozens" of searches); the only precise OpenAI numbers come from third-party benchmarking (27–125 searches observed per task).
- **"Search budget dilution" from bundling multiple questions is a plausible, mechanistically well-supported inference — not a directly proven, isolated finding.** Anthropic's own data shows token/tool-call budget explaining ~80–95% of research quality variance, and its published scaling rules show effort is explicitly rationed by *perceived* task complexity. No vendor or academic study we found ran a controlled test of "identical topic asked alone vs. bundled with three others" across these four commercial products. The inference is reasonable; treat it as engineering prior, not proven law, until you test it yourself (Part 2 gives you the test design).
- **Deep research tools are genuinely strong at breadth-first, decomposable, verifiable-fact tasks** (OpenAI's Deep Research solves roughly half of BrowseComp's deliberately hard-to-find factual questions, versus near-zero for non-agentic baselines) **and genuinely weak at exhaustive-recall tasks.** In a controlled comparison against manual PubMed systematic review, ChatGPT's deep research function found only 47.8% of the studies a human search protocol found — consistent with a broader 2025 methodological review reporting AI tools miss a median of ~91% of relevant studies in literature search specifically (as distinct from screening/classification, where AI performs much better).
- **All four systems fabricate under pressure.** In the same PubMed study, when asked to account for articles it had missed, ChatGPT's deep research invented two nonexistent papers rather than admitting the gap. Independent benchmarking finds citation "groundedness" (claims that carry no citation at all) is a weak point industry-wide, not a single-vendor problem.
- **No single tool wins on every axis**, and the axes that matter (citation accuracy vs. citation volume, speed vs. depth, precision-on-a-known-source vs. breadth-on-an-unknown-landscape) trade off differently by vendor. This has direct design implications for a multi-tool framework, covered in Part 5.

---

## Part 1 — How Each System Actually Works

### 1.1 Architecture at a glance

| | Gemini Deep Research | OpenAI/ChatGPT Deep Research | Claude Research | Perplexity Research |
|---|---|---|---|---|
| **Core approach** | Single agent, RL fine-tuned for long-horizon planning | Single agent, RL fine-tuned (originally on o3) for browsing/reasoning | Multi-agent orchestration (lead + subagents) of *general-purpose* Claude models — no dedicated research model | Retrieval-first "answer engine": owned index + reranker + generation model |
| **Dedicated research-trained model?** | Yes | Yes | **No** — prompt/tool engineering on Claude 4-class models | Yes (Sonar / Sonar Deep Research) |
| **Planning style** | Interactive: proposes a plan, user can edit before execution | Implicit plan-act-observe (ReAct-style) with backtracking | Explicit: lead agent writes a plan to memory, then delegates to parallel subagents | Query decomposition → parallel retrieval → synthesis, less user-visible planning |
| **Reported search volume** | ~80 queries typical / ~160 "Max" (vendor API docs) | Not published as a fixed figure; 27–125 observed per task in third-party testing | Heuristic: 3–10 calls (simple) to 10+ subagents each making multiple calls (complex) | "Dozens" of searches (vendor); ~20–50 estimated by third parties |
| **Typical turnaround** | Minutes to ~tens of minutes; independent tests found it markedly slower than rivals in one head-to-head | 5–30+ minutes | Minutes; one independent test measured ~6 minutes for 261 sources | Fastest of the four — 2–4 minutes (vendor-stated design goal) |
| **Synthesis mechanism** | Multi-pass self-critique before final report | Single-pass synthesis with inline citation generation | Lead agent synthesizes subagent outputs, then a **separate CitationAgent pass** maps claims to sources | Strict-grounding generation — designed to answer only from retrieved passages |
| **Distinctive integration** | Gmail/Drive/Docs/Chat, MCP servers, chart/infographic generation | File upload, code execution for data analysis | Google Workspace, large document context alongside web search | Deep index ownership; no dependency on a third-party search API |

### 1.2 Gemini Deep Research

Google's own description of the system frames it as an iterative loop: at each step the model grounds itself in everything gathered so far, identifies gaps, and decides whether to keep researching — explicitly trading comprehensiveness against compute and wait time. The user reviews and can edit the research plan before execution starts, which is a genuine point of difference from the other three (none of which expose an editable plan by default).

The clearest, most precise public data point in this entire category comes from Google's current API documentation: a standard Deep Research task uses **approximately 80 search queries, ~250K input tokens (50–70% cached), and ~60K output tokens**; the "Deep Research Max" variant scales to **~160 queries and ~900K input tokens** for due-diligence-grade tasks. This is a real, dated, vendor-published number — not a marketing estimate — and it's the only such precise figure available for any of the four systems.

Academic benchmarking (DeepResearch Bench, a 100-task PhD-level evaluation using paired RACE/FACT frameworks) found Gemini's Deep Research led on overall report quality (RACE) and produced by far the most citations per report (~111 "effective citations" on average) — but its *citation accuracy* trailed Perplexity's. In other words: Gemini cites the most, and gets the most individual facts right in absolute terms, but a larger share of any given citation is imprecise compared to Perplexity's tighter, more conservative sourcing. Independent (non-benchmark) testing produced a genuinely conflicting data point worth flagging: one head-to-head test found Gemini slower and covering fewer sources (62 sources, >15 minutes) than Claude's Research mode (261 sources, ~6 minutes) on the same task type — which sits oddly next to the ~80-query figure in Google's own docs and likely reflects mode/settings differences we could not fully resolve. We report both rather than picking one.

### 1.3 OpenAI / ChatGPT Deep Research

OpenAI's system, launched February 2025, was originally a fine-tuned variant of o3 trained end-to-end with reinforcement learning on browsing and reasoning trajectories — both auto-graded tasks with verifiable answers and open-ended tasks graded against rubrics. It follows a plan–act–observe loop with explicit backtracking when a line of inquiry proves unproductive. OpenAI's own BrowseComp writeup (a benchmark deliberately built from questions designed to survive at least five simple searches without being answered) reports Deep Research **solving around half of the problems** — a large, genuine capability jump over non-agentic baselines, which OpenAI attributes to the ability to adapt search strategy mid-task and synthesize across many sources rather than stopping at the first plausible answer.

At launch it scored 26.6% on Humanity's Last Exam (HLE) and 67.36% (pass@1) on GAIA — both were record-setting at the time, though HLE scores across the industry have moved substantially since (see Part 3). OpenAI upgraded the underlying model to GPT-5.2 in February 2026 and retired the original o3-based "legacy" mode on March 26, 2026; the specific API model IDs (`o3-deep-research`, `o4-mini-deep-research`) are themselves being phased out through mid-to-late 2026 per OpenAI's deprecation schedule, with the capability continuing under newer model generations. If you are calling this via API, budget engineering time for model-ID churn independent of any change in underlying quality.

No OpenAI documentation we could find publishes a target search-call count comparable to Google's. The only precise numbers come from independent testing: one factual-accuracy benchmark observed OpenAI's o3/o4-mini deep research models running **27–125 web searches per task** — yet scoring **75.8–81.8% accuracy, the lowest of six tools tested**, despite the highest search volume and 2–6× the cost of a comparable agent. That benchmark's qualitative finding is important: the tools that won navigated directly to primary sources (an SEC filing, an official docs page) and read them carefully, while OpenAI's models "searched broadly but extracted less precise information from the pages they found" — a volume-without-precision pattern worth remembering when designing your framework's tool-selection logic.

### 1.4 Claude Research (web search + extended thinking)

This is the one system in the comparison set with no dedicated research-trained model. Per Anthropic's own engineering writeup, the Research feature is an **orchestrator-worker multi-agent system**: a "LeadResearcher" agent (Claude Opus in Anthropic's reference implementation) analyzes the query, writes a plan to persistent memory (to survive context truncation past 200K tokens), and spawns 3–5 subagents (Claude Sonnet in the reference implementation) that search in parallel, each running its own observe–orient–decide–act loop with interleaved extended thinking between tool calls. The lead agent synthesizes subagent findings and, if gaps remain, can spawn additional subagents. A separate **CitationAgent** then does a dedicated pass mapping every claim in the draft report to a specific source — architecturally distinct from the other three systems, which generate citations inline during synthesis rather than as a separate verification step.

Anthropic publishes explicit effort-scaling heuristics baked into the lead agent's prompt: **simple fact-finding gets 1 agent and 3–10 tool calls; direct comparisons get 2–4 subagents with 10–15 calls each; complex, open-ended research gets 10+ subagents with explicitly divided responsibilities.** This is the most transparent public documentation of "how much search effort does a system decide to spend" of any vendor in this set, and it is directly relevant to the multi-topic question in Part 2.

The quantified cost of this design, also disclosed by Anthropic: general agentic tasks use roughly 4× the tokens of a simple chat turn, and the full multi-agent Research system uses roughly **15× the tokens of a simple chat turn**. Anthropic's own regression analysis of the BrowseComp benchmark found that **token usage alone explains about 80% of performance variance**, with tool-call count and model choice explaining most of the rest (95% combined) — i.e., for this category, "how much compute you're willing to spend" is close to the single dominant lever, more than any particular clever prompting trick. Anthropic reports a 90.2% win-rate improvement for the multi-agent system over a single Opus agent on their internal research evaluation, illustrated with an S&P 500 IT-sector board-membership enumeration task that the single agent failed via slow sequential search and the multi-agent system solved via decomposition.

Two caveats worth flagging. First, access requires web search to be enabled and then Research toggled on separately, and the feature sits behind paid tiers (Pro/Max/Team/Enterprise) — it is not available on Claude's free tier. Second, when academic benchmarks test "Claude" in a deep-research comparison, they often test a bare model-with-search-tool baseline (e.g., "Claude Sonnet w/ Search") rather than the full orchestrated Research product, since the latter isn't exposed via a stable API endpoint the way OpenAI's and Google's are. In the one major academic benchmark we found with a Claude entry (DeepResearch Bench), the Claude-with-search baseline scored best-in-class among *non-agentic* baselines on overall report quality and had strong citation accuracy — but this is not a clean apples-to-apples test against the full multi-agent Research product, and we did not find one that was.

### 1.5 Perplexity Research

Perplexity's architecture is retrieval-first rather than agent-first: the company owns its own crawler and index (reported at somewhere between "hundreds of billions" and 200 billion+ unique URLs depending on source and date), runs a hybrid lexical-plus-embedding retrieval funnel ending in cross-encoder reranking, and only then hands the top passages to a generation model (Sonar) instructed to answer strictly from what was retrieved, with a citation on every claim. This "grounding-first" design is the most architecturally distinct of the four and plausibly explains its benchmark profile: strong precision, comparatively less independent reasoning/planning visible to the user.

One genuine point of public uncertainty: Sonar's current base model is inconsistently reported across sources. Perplexity's own documentation (via a technical third-party deep-dive) describes Sonar Large as built on Llama 3.1 70B; separately, marketing-oriented sources describe a newer Sonar as built on unspecified "GPT-5 Chat architecture," which reads as unverified and is not corroborated anywhere authoritative we could find. We flag rather than resolve this — it matters if you're deciding whether to route cost-sensitive traffic through Sonar based on assumptions about its underlying model family.

Deep Research launched free-for-all on February 14, 2025 (5 queries/day free tier, higher limits Pro) and was renamed simply "Research" around May 2025, positioned between "Search" and the more agentic "Labs" tier. Perplexity's own description: **"dozens" of automatic searches, reading "hundreds of sources," typically finishing in 2–4 minutes** — explicitly the fastest of the four by design intent, and independent testing generally confirms this. At launch it scored 21.1% on HLE and a strong 93.9% on SimpleQA (a short-form factuality benchmark that plays directly to a strict-grounding retrieval architecture's strengths).

DeepResearch Bench found a genuine, useful trade-off: Perplexity had the **highest citation accuracy of the deep-research category (90.2%)** but one of the lower overall RACE (report-quality) scores — it says less, and covers less ground, but what it says is comparatively well-attributed. A separate independent benchmark corroborates the precision side: Perplexity Sonar scored 87.9% on a factual-accuracy test (second only to a specialized non-consumer tool, and ahead of OpenAI's deep research models), while producing far more words per task (5,253 on average) than terser competitors — and the same test caught a case where Perplexity's verbosity didn't correlate with correctness (it wrote the most about a specific technical detail and got it wrong, while a competitor wrote a fifth as much and got it right). In a separate benchmark, Perplexity's Research also failed an explicit table-formatting instruction outright, scoring zero on that requirement despite substantive underlying content — a concrete instruction-following miss.

---

## Part 2 — Search Budget Allocation: Single-Topic vs. Multi-Topic Prompts

This is the question most directly relevant to a framework built on these tools, so it gets dedicated treatment.

### 2.1 What is actually documented

Only Anthropic publishes an explicit effort-allocation policy, and it's worth quoting the structure precisely because it's the clearest evidence available anywhere in this category: the lead agent is instructed to give **1 agent and 3–10 tool calls to simple fact-finding, 2–4 subagents with 10–15 calls each to direct comparisons, and 10+ subagents with explicitly divided responsibilities to complex, open-ended research** — and Anthropic explicitly states this scaling logic exists *because* early versions of the system misjudged effort in both directions: sometimes spawning 50 subagents for a trivial query, sometimes running long sequential searches for something requiring real decomposition.

This tells us two things with confidence. First, these systems do not treat "more sub-questions in the prompt" as automatically meaning "proportionally more search calls" — the *lead agent's own judgment* of complexity determines the budget, and that judgment can be wrong in either direction. Second, a multi-topic prompt is not silently ignored or truncated to one topic; the documented behavior is that it gets classified into a complexity tier and subagents get dispatched against the tier, not against the literal count of sub-questions. Whether the tier-classification correctly scales with three bundled questions the way it would with one three-times-harder question is the open part.

Google's API documentation offers an indirect data point in the same direction: query volume (~80 vs. ~160) is explicitly described as scaling with "the depth of research required," determined autonomously by the agent per-task — again, a judgment call about perceived complexity, not a hard per-question allocation.

### 2.2 The mechanism that makes "dilution" plausible

Three separate, independently well-evidenced findings combine into a coherent (if not directly tested) causal story:

1. **Effort is the dominant lever.** Anthropic's BrowseComp regression found token/tool-call budget explains ~80–95% of quality variance. If a system allocates a fixed or coarsely-tiered budget to a request regardless of how many distinct questions are packed into it, each individual question inside a bundled prompt receives a smaller effective share of that budget than it would as a standalone request.
2. **A single continuous response reflects a single framing.** An analysis of multi-agent research design makes a structural point worth taking seriously: a single-pass answer to a multi-part prompt tends to be either broad-and-shallow (touching each part briefly) or narrow-and-deep (thoroughly answering the part that dominated the model's attention) — genuinely covering several questions with equal depth requires something closer to the subagent-per-facet architecture Anthropic and Google both use, not a single synthesis pass.
3. **Long-context synthesis has a real, physically-grounded weak spot.** The "lost in the middle" effect — degraded recall for information positioned in the middle of a long context, tied to positional-encoding attention decay (RoPE) — is well-replicated across model families and, per a 2026 follow-up study, **multi-step reasoning degrades more than simple factual retrieval** even in models with very large context windows. A bundled multi-topic research task inherently produces a longer, more heterogeneous context to synthesize over at the end; if the synthesis step underweights middle content, a topic that happens to land mid-context in the gathered material is at elevated risk regardless of how much search effort was spent finding it.

### 2.3 What is *not* proven, and honest counter-evidence

We did not find a published, controlled study that isolates this specific variable — identical topic, asked alone vs. bundled with two or three others, same tool, measured for depth/accuracy per topic — across Gemini, OpenAI, Claude, or Perplexity's actual products. The closest adjacent evidence is a 2025 study of small open-source language models (3–8B parameters) on unrelated NLP classification tasks (sentiment, NER, translation, etc.), which found multi-task prompt degradation is real and measurable but **highly architecture-dependent and not universal** — some models degraded severely, one (Qwen3 4B) showed almost no degradation, and two models actually *improved* on some tasks under multi-task conditions ("positive transfer"). This is genuinely adjacent evidence for the general phenomenon of prompt-complexity effects on LLMs, but it used small non-agentic classifier-style models on a battery of *unrelated* tasks (translation + JSON formatting + NER, etc.) — a different regime from a frontier agentic system decomposing several *related* research sub-questions with dedicated subagents. Extrapolating its exact numbers to Gemini/OpenAI/Claude/Perplexity deep research would overstate what's known.

Practitioner opinion is also genuinely split, which is itself informative. Community prompting guides for these tools commonly advise focusing a deep-research prompt on one primary objective and breaking multi-part questions into separate calls. But direct guidance surfaced from an OpenAI-affiliated prompting session argued close to the opposite: that ambitious, multi-part prompts are fine because the model "will strategize internally" during planning. Both claims are plausible-sounding and neither is backed by a controlled experiment we could locate.

### 2.4 Our assessment

**Verdict: plausible and mechanistically well-supported, not independently proven.** Given (a) Anthropic's explicit, documented tiered-effort allocation, (b) the dominance of effort/budget in explaining quality variance, and (c) known long-context synthesis weaknesses that get worse as the material to synthesize gets more heterogeneous, we think it's reasonable to design your framework as though dilution is real and to treat any specific quantitative claim about *how much* dilution occurs as unverified until you run your own test. Section 2.5 gives a test design; Part 5 gives the resulting design recommendation.

### 2.5 A test your framework should run (since the literature doesn't answer it for you)

Take a set of topics your framework actually needs researched. For each, run it (a) alone and (b) bundled with two unrelated topics in one prompt, on each tool you plan to use, and score the per-topic output on the same rubric (source count, specificity, presence of the single most load-bearing fact, citation accuracy). This is a small, cheap experiment relative to the cost of building framework architecture on an unverified assumption, and it will also reveal whether dilution is uniform across your four candidate tools or concentrated in one or two — which the existing literature cannot tell you.

---

## Part 3 — When Deep Research Excels vs. When It Fails

### 3.1 Benchmark snapshot (treat as directional — see caveats below)

| Benchmark | What it measures | OpenAI DR | Gemini DR | Claude* | Perplexity DR |
|---|---|---|---|---|---|
| Humanity's Last Exam (HLE) | Expert-level cross-discipline reasoning | 26.6% (Feb 2025 launch) | 18.8%–54.6% (range across snapshots/variants; 54.6% is a vendor-reported "Max" figure from Apr 2026) | Not separately benchmarked as a DR product | 21.1% (Feb 2025 launch) |
| GAIA (pass@1) | Real-world multi-step tool-use tasks | 67.4% | Not directly found for the DR product | Not directly found for the DR product | Not directly found for the DR product |
| BrowseComp | Deliberately hard-to-find factual questions | ~51.5% ("about half") | Not directly found for the DR product | 12–20% for base Claude models *with search tool* (not the orchestrated Research product) | Not directly found for the DR product |
| DeepResearch Bench — RACE (report quality, 0–100) | Overall research report quality vs. expert reference | ~47.0 | ~48.9 (leading) | Best-in-class *among LLM+search baselines*, not full DR agents | ~42.3 |
| DeepResearch Bench — Citation accuracy | % of citations that correctly support their claim | ~78% | ~81% | Strong, among LLM+search baselines | **90.2% (leading)** |

*\*Claude's full multi-agent Research product is not exposed via a benchmark-stable API the way OpenAI's and Google's are; academic benchmarks generally test a bare Claude model with a search tool instead, which understates what the orchestrated product does per Anthropic's own 90.2%-improvement claim above. This is a real gap in the independent-evaluation literature, not a judgment about Claude's capability.*

**Caveats on this table, stated plainly:** these numbers come from different snapshots across an 18-month period during which every vendor updated its underlying models multiple times; HLE and GAIA numbers in particular are frequently reported for a *base model*, not unambiguously for the *deep-research product specifically*, across the papers we reviewed; and benchmark design changes the ranking dramatically — a separate, harder benchmark using precise structured-fact questions (prices, exact specifications) found accuracy for the same category of tools in the 20–35% range, with Perplexity's Sonar Deep Research topping that harder test at only 34%. Read rankings as "roughly this tier," not "precisely this score, today."

### 3.2 Where deep research genuinely excels

- **Breadth-first, decomposable, independently-verifiable sub-facts.** Anthropic's own illustrative example — enumerate the board members of every company in the S&P 500 IT sector — is the archetype: many independent, verifiable atoms of fact that parallelize cleanly across subagents. BrowseComp's ~51.5% solve rate for OpenAI's system (against a benchmark specifically designed so five ordinary searches won't find the answer) is real, substantial capability that did not exist in non-agentic search two years ago.
- **Fast first-pass landscape mapping and competitive/business intelligence.** Anthropic's internal usage-clustering data (via their Clio research tool) shows real-world Research usage concentrated in developing business growth/revenue strategy, academic research assistance, and verifying information about people, places, and organizations — exactly the domains where "good enough, fast, with sources to check" beats "nothing, because a human hasn't had time yet."
- **Precision-on-a-known-source tasks**, when the correct primary document is identifiable — independent testing found agentic tools that navigated directly to a specific SEC filing or official documentation page outperformed broader, higher-search-volume deep-research models on exactly the same question, because they read one authoritative source carefully instead of skimming many.
- **Combining private/internal documents with live web context** — Claude's and Gemini's ability to reason over uploaded documents or connected Workspace content alongside web search is a genuine differentiator neither ChatGPT's nor Perplexity's default consumer product matches as directly.

### 3.3 Where it's mediocre or fails

- **Exhaustive-recall tasks.** This is the clearest, most consistently documented weakness across independent sources. A controlled comparison of ChatGPT's deep research against a manual PubMed systematic-review protocol on dental implantology found manual search identified 124 candidate articles (23 included after screening) while ChatGPT retrieved 114 candidates but only actually synthesized 11 — a sensitivity of 47.8%. A broader 2025 methodological review of 19 comparative studies found generative AI tools **missed a median of ~91% of relevant studies** in literature search specifically (recall averaging 13%, range 4–32%), while making incorrect inclusion decisions in up to 29% of cases and incorrect exclusions in up to 83%. The same body of literature draws an important distinction: AI is *much* better at screening/classifying a pre-identified set of candidates (one meta-analysis found title/abstract screening sensitivity of 99.2%) than at the initial exhaustive *search* that produces the candidate set in the first place. That's a completeness problem, not a comprehension problem.
- **Confabulation under scrutiny.** In the same dental-implant study, when researchers asked the model to account for studies it had missed relative to the manual search, it **fabricated two nonexistent articles** rather than reporting the gap honestly. This is a materially different (and more concerning) failure mode than an ordinary hallucinated fact — it's confabulation specifically triggered by being pressed on a known weakness.
- **Version- and time-specific precision, even when explicitly specified.** An independent technical-documentation benchmark asked six tools (including Claude Code and OpenAI's o3/o4-mini deep research) to produce a migration guide between two exactly-specified software versions. Three of six tools pulled requirements from the wrong, more prominent, older documentation page despite the correct version number being stated explicitly in the prompt — a concrete, reproducible illustration of "reads the most findable document, not necessarily the correct one."
- **Citation "groundedness."** Beyond citation accuracy (are the citations that exist correct?), a separate axis — groundedness (are all factual claims backed by *any* citation?) — scored much lower across systems in one open-ended-science benchmark (roughly 0.31–0.59) than citation faithfulness for claims that were cited (roughly 0.80–0.86). Read plainly: a meaningful fraction of assertions in an otherwise well-sourced report carry no citation at all.
- **Source-authority discrimination.** Anthropic's own human testers found early versions of their system systematically preferred SEO-optimized content-farm pages over less-visible but more authoritative sources (academic PDFs, primary blogs) until this was explicitly corrected with prompt-level source-quality heuristics — a reminder that "reads the whole web" doesn't imply "correctly weighs the web's reliability" without deliberate engineering.

### 3.4 Failure-mode taxonomy

| Failure mode | What it looks like | Concrete documented instance |
|---|---|---|
| Citation fabrication under pressure | Invents a plausible-sounding source rather than admitting a gap | ChatGPT deep research fabricated 2 nonexistent articles when asked about missed studies |
| Version/temporal drift | Defaults to the most prominent doc version, not the one specified | 3 of 6 tools used wrong software-version docs despite the exact version stated in-prompt |
| Low groundedness | Factual claims with no citation at all, even in a well-sourced report | Groundedness scores ~0.31–0.59 vs. ~0.80–0.86 citation faithfulness in one benchmark |
| Exhaustive-recall gaps | Systematically misses a large share of relevant items in open-ended search | Median ~91% miss rate for GenAI literature search across 19 comparative studies |
| Subagent duplication/gaps | Poor task division causes redundant or missed coverage between parallel agents | Anthropic's own account: two subagents both investigated 2025 supply chains while a third, unprompted, investigated the unrelated 2021 chip crisis |
| Effort miscalibration | Over- or under-scales search effort relative to true complexity | Anthropic's account of early versions spawning 50 subagents for trivial queries |
| Source-authority bias | Defaults to SEO-optimized/prominent sources over authoritative ones | Anthropic testers found this pattern pre-correction; required explicit prompt heuristics to fix |
| Instruction non-compliance | Ignores an explicit output-format constraint | Perplexity Research scored zero on an explicit table-format requirement in one benchmark despite good content |
| Single-document tunnel vision | Reads only the first/most findable matching document, not the full required set | None of 6 tools correctly read all four sequential upgrade guides a task required |

---

## Part 4 — AI Deep Research vs. Human Expert Research

### 4.1 The clearest available head-to-head: systematic review search

The most rigorous, directly comparable evidence we found is a peer-reviewed study (published October 2025) comparing ChatGPT's deep research function against a standard manual PubMed systematic-review protocol, same date, same search terms, same inclusion criteria, on dental implantology. Results: manual search returned 124 articles → 23 included; ChatGPT returned 114 articles → 13 selected → only 11 actually used in synthesis, for a sensitivity of 47.8% against the human-identified set. The authors' own conclusion, in substance: deep research can support but not replace manual systematic search and selection, offering real value for writing support and preliminary synthesis, but with reliability and sensitivity limitations serious enough to require cautious, transparently-disclosed use.

This single study's finding is corroborated at larger scale by a 2025 methodological review synthesizing 19 comparative studies across the evidence-synthesis literature: generative AI missed a median of ~91% of relevant studies in search, made incorrect data extractions in 4–31% of cases, and incorrect risk-of-bias assessments in 10–56% of cases — leading the reviewers to conclude current evidence does not support using generative AI for evidence synthesis without meaningful human involvement.

### 4.2 The important counter-finding: AI-as-secondary-reviewer works much better than AI-as-primary-searcher

The same literature contains a genuinely encouraging counter-example that matters for framework design: a 2024 multicenter study using AI as a *secondary* reviewer alongside a human-led primary search — rather than as the autonomous search mechanism — reduced missed-article rates to under 1%, comparable to human-reviewer accuracy. Separately, title/abstract screening specifically (classifying a pre-identified candidate set, as opposed to generating the candidate set from scratch) is a far more mature AI capability, with one meta-analysis reporting 99.2% sensitivity. The failure mode is concentrated in *exhaustive discovery*, not in *judgment applied to a bounded set* — a distinction with direct implications for how you'd architect a research pipeline (Part 5).

### 4.3 Domains and conditions favoring each side

| Condition | Favors AI deep research | Favors human expert research |
|---|---|---|
| Task structure | Decomposable into independent, verifiable sub-facts | Requires judgment about source quality/relevance that isn't yet well-specified |
| Completeness requirement | "Good enough coverage fast" is acceptable | Exhaustive/high-recall is the actual requirement (legal discovery, systematic review, regulatory filing) |
| Source landscape | The correct primary source is identifiable/nameable | Requires weighing many partially-authoritative, conflicting sources |
| Verification cost | Downstream human review of citations is planned anyway | Output will be trusted with minimal downstream verification |
| Time sensitivity | Speed matters more than marginal completeness | Stakes justify the multi-week/month timeline of rigorous human research |
| Novelty | Question has been asked before somewhere on the indexed web | Genuinely novel synthesis or judgment call with no close precedent |

### 4.4 A dissenting view worth naming

Cognitive scientist and AI critic Gary Marcus has argued publicly and repeatedly that deep research tools remain fundamentally limited by the same issues he's raised about neural-network-based AI generally: weak temporal/factual reasoning grounded in statistical pattern-matching rather than causal understanding, and poor calibration — confident-sounding output that doesn't reliably signal its own uncertainty. He has separately raised a longer-horizon concern about AI-generated research contributing to a decline in the reliability of the published scientific literature if used to mass-produce papers without adequate human oversight. This is opinion commentary rather than a controlled study, and we flag it as such — but it's a widely-cited skeptical position in this space and the underlying concerns (calibration, temporal reasoning) are consistent with the empirical failure modes documented in Part 3.

---

## Part 5 — Practical Optimal Use Patterns for Framework Design

These recommendations are derived directly from the evidence above, not generic prompting advice.

1. **Decompose upstream, at your orchestration layer, rather than trusting internal decomposition.** Given Part 2's findings — effort allocation is judgment-based and can misjudge complexity, and long-context synthesis measurably degrades on middle content — a framework that dispatches one well-scoped topic per research call gives you explicit control over effort allocation instead of hoping each vendor's internal classifier scales correctly with an unpredictable bundle of sub-questions. This is the single highest-leverage design decision available to you given what's actually documented.
2. **Provide known primary sources explicitly when you have them.** The evidence in 3.2–3.3 is consistent: tools that navigate to a specific, named authoritative source outperform tools doing a broad open crawl, on the same question. If your framework already knows the likely authoritative source (an official docs page, a specific filing), pass it in rather than relying on the tool to find it.
3. **State versions, dates, and scope constraints explicitly — then verify the output against them.** The Unity documentation case shows models will silently substitute the most prominent version for the one you specified. Treat any date- or version-sensitive claim as needing a post-hoc check against your stated constraint, not just against plausibility.
4. **Treat single-pass output as a first draft requiring a verification layer, especially for citation-critical or completeness-critical work.** The most consistently validated pattern in the literature we reviewed is AI-as-secondary-reviewer (< 1% miss rate) dramatically outperforming AI-as-autonomous-primary-searcher (median ~91% miss rate on the hardest version of this task). If your framework's use case resembles exhaustive discovery, architect a human or second-AI-pass verification step rather than trusting single-pass completeness.
5. **Build an automatic citation-groundedness check.** Because uncited-but-asserted claims are a documented weak point across all four vendors (not one), a generic post-processing pass that flags factual sentences with no attached citation — for human or secondary-model review — will catch a real, quantified failure mode rather than a hypothetical one.
6. **Match tool to task type using documented strengths, not brand preference:**
   - Enumerable, decomposable, many-independent-facts tasks → a parallel multi-agent approach (Claude's or Gemini's architecture) with a generous call budget.
   - A single hard-to-find fact behind a knowable primary source → a targeted agentic fetch tool rather than a broad "deep research" product; independent testing found this consistently beats higher-search-volume deep-research models on precision.
   - Fast, current, short factual lookups where speed and cost matter more than maximum depth → a retrieval-first, strict-grounding tool (Perplexity's design intent, and its benchmark profile).
   - Tasks requiring private/internal documents combined with the live web in one pass → Claude or Gemini's document-plus-search integration.
7. **Budget consciously for the real cost multiplier.** Anthropic's own disclosed figure — roughly 15× the token cost of a simple chat turn for the full multi-agent system — is a genuine economic constraint, and Anthropic explicitly frames multi-agent research as justified only when task value clears that multiplier. Apply the same discipline in your framework: don't route trivial lookups through your most expensive/most decomposed research pipeline.
8. **Use extended-thinking/reasoning modes where available.** Independent hallucination benchmarking found extended/reasoning modes roughly halved measured hallucination rates across the models tested in one 2026 study, via what researchers described as self-correction visible in the reasoning trace. Where a tool offers a reasoning-mode toggle, the evidence favors turning it on for anything citation-sensitive.
9. **For genuinely completeness-critical questions, don't rely on a single vendor.** The benchmark data in Part 3 shows different systems are best on different specific axes — Gemini on citation volume, Perplexity on citation accuracy, OpenAI on instruction-following, Claude's architecture on parallel decomposition — meaning single-tool reliance concentrates that tool's specific blind spot. Cross-running a high-stakes query through two systems and reconciling disagreements is a defensible design pattern given the evidence, not just belt-and-suspenders caution.
10. **Run the test in Section 2.5 before committing to an architecture.** This report can tell you what's plausible and well-evidenced; it cannot tell you the exact dilution curve for your specific topics on your specific chosen tools, because that experiment doesn't exist in the public literature yet.

---

## Part 6 — Epistemic Status: Verified, Inferred, and Genuinely Unknown

| Claim | Status | Basis |
|---|---|---|
| Anthropic's subagent effort-scaling rules (1 agent/3–10 calls → 2–4 subagents/10–15 calls each → 10+ subagents) | **Verified** | Primary source: Anthropic's own engineering publication |
| Gemini's ~80 / ~160 query counts | **Verified** | Primary source: Google's current API documentation |
| Token usage explains ~80% of BrowseComp performance variance for Claude's system | **Verified** | Primary source: Anthropic's own regression analysis |
| OpenAI's internal target search-call count | **Not publicly documented**; only third-party-observed (27–125/task) | Independent benchmark, not vendor disclosure |
| Perplexity Sonar's current base model | **Contested** — conflicting claims across secondary sources, no authoritative recent confirmation found | Mixed-quality secondary sourcing; flagged, not resolved |
| "Search budget dilution" occurs specifically when bundling unrelated topics in one prompt | **Reasonable, mechanistically-supported inference** | Composite of documented effort-tiering + long-context degradation research + adjacent (not directly on-topic) multi-task prompting study; **no controlled test across these four products found** |
| Whether Claude's full orchestrated Research product (vs. a bare model+search baseline) was tested in academic deep-research benchmarks | **Unclear/likely not**, based on available benchmark methodology sections | Inferred from benchmark descriptions, not confirmed by benchmark authors |
| Relative accuracy rankings across the four systems | **Verified directionally, not absolutely** — rankings shift meaningfully with benchmark design and model-version snapshot | Multiple independent benchmarks, partially conflicting |
| Which system is fastest / covers most sources | **Genuinely conflicting evidence** — vendor docs and independent tests disagree for at least one vendor (Gemini) | Direct contradiction between primary API docs and one independent head-to-head test; both reported above rather than adjudicated |

---

## Selected Sources

- Anthropic — ["How we built our multi-agent research system"](https://www.anthropic.com/engineering/multi-agent-research-system), Anthropic Engineering, June 2025
- Anthropic — ["When to use multi-agent systems (and when not to)"](https://claude.com/blog/building-multi-agent-systems-when-and-how-to-use-them), Claude blog, Jan 2026
- Anthropic — Claude Help Center: ["Use research on Claude"](https://support.claude.com/en/articles/11088861-use-research-on-claude); ["Enable and use web search"](https://support.claude.com/en/articles/10684626-enable-and-use-web-search)
- Google — ["Gemini Deep Research agent"](https://ai.google.dev/gemini-api/docs/deep-research), Gemini API developer docs
- Google — ["Gemini Deep Research — your personal research assistant"](https://gemini.google/overview/deep-research/)
- OpenAI — ["BrowseComp: a benchmark for browsing agents"](https://openai.com/index/browsecomp/)
- OpenAI Help Center — ["Deep research in ChatGPT"](https://help.openai.com/en/articles/10500283-deep-research-faq)
- Perplexity — ["Introducing Deep Research"](https://www.perplexity.ai/hub/blog/introducing-perplexity-deep-research); Perplexity Help Center, ["What is Research mode?"](https://www.perplexity.ai/help-center/en/articles/10738684-what-is-research-mode)
- Du et al., ["DeepResearch Bench: A Comprehensive Benchmark for Deep Research Agents"](https://arxiv.org/pdf/2506.11763), arXiv 2506.11763
- Bencze, Sokolowski, et al., "Comparing Manual and ChatGPT Deep Research on Systematic Search and Selection in the PubMed Database on the Topic of Dental Implantology," *International Journal of Dentistry*, Oct 2025
- Clark, Barton, Albarqouni, et al., "Generative artificial intelligence use in evidence synthesis: A systematic review," *Research Synthesis Methods*, 2025 (summarized via University of Cambridge Medical Library systematic-review guidance)
- AIMultiple, ["AI Deep Research: Claude vs ChatGPT vs Grok"](https://aimultiple.com/ai-deep-research) — independent benchmark methodology and results, updated June 2026
- Di Maio & Gozzi, "Degradation of Multi-Task Prompting Across Six NLP Tasks and LLM Families," *Electronics* 14(21), Nov 2025
- Liu et al. (2023/2024) "lost in the middle" long-context positional-bias research, as synthesized in multiple 2025–2026 follow-up papers
- Gary Marcus, ["Deep Research, Deep Bullshit, and the potential (model) collapse of science"](https://garymarcus.substack.com/p/deep-research-deep-bullshit-and-the), Marcus on AI, Feb 2025

*This report reflects publicly available information as of August 2026. Given the pace of change in this category, verify current search-volume figures, pricing, and model versions directly against vendor documentation before finalizing framework architecture decisions.*
