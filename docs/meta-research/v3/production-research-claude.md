I built this from public material: engineering posts from Anthropic, LangChain, Cognition and OpenAI, plus papers and audits (STORM, TTD-DR, CRAG, DeepTRACE, DREAM, and a few 2026 preprints). Two caveats.

- **Sources:** The detailed methods of commercial systems remain proprietary, so for those I'm relying on what vendors have said publicly. I can't speak to Claude's own Research feature beyond Anthropic's published posts. I lean on those posts because they're among the most detailed accounts available, so weigh that bias.
- **Evidence quality:** Much of the quantitative evidence is vendor-internal evals, LLM-judged benchmarks, or single-task preprints. Treat effect sizes as directional. Where I'm extrapolating rather than reporting, I say so.

**The short version**

- **Define "done" before searching.** Plan around coverage objectives and per-step completion criteria, not around tools.
- **Parallelize reads, serialize decisions.** Workers explore in isolated contexts and return claim-level evidence records. One agent writes the report.
- **Verification is its own stage and needs its own retrieval.** Citations that resolve and look on-topic routinely fail to support the claim attached to them.
- **Depth isn't monotonic.** More searching finds harder facts but can degrade grounded synthesis unless what the synthesizer sees stays small and verified.
- **Evaluate stages separately** (planning, retrieval, source quality, claim support, synthesis) rather than judging final prose.

## A. Research planning architecture

**Decomposition.** Step zero is scoping: LangChain's open implementation clarifies the request and generates a brief through user interaction before anything is delegated. After that, the evidence on good decomposition comes mostly from failures.

- Anthropic's early lead agents gave subagents short, vague instructions, and the subagents duplicated searches or misread the task. The fix was that every delegation specifies an objective, output format, tool and source guidance, and task boundaries.
- Salesforce's enterprise system goes further. It enumerates information objectives as an outline before any retrieval, on the theory that early retrieval pulls plans toward whatever is easy to find, then runs a reflection pass for missing fields. Removing the outline dropped their actionability score from 82.1 to 67.3.
- STORM approaches it differently. It surveys articles on similar topics to discover perspectives, then has the model ask questions from each one.

My synthesis is to decompose on two axes, what must be known (coverage) and who would disagree (perspectives), and to give every leaf an explicit deliverable.

There's a tradeoff between planning from the model's priors (fast and coherent, but may bake in stale beliefs) and planning retrieval-first (grounded, but prone to drift). Google's TTD-DR starts from a rough draft and revises it using retrieved information. A cheap hybrid I'd suggest is to outline from the task alone, run a short broad scan, revise the outline, then fan out. Anthropic similarly prompts agents to begin with short, broad queries before narrowing.

**Plan structure.** The best-evidenced middle path is a DAG with bounded replanning.

- In Salesforce's system, dependencies determine what runs in parallel and which prior outputs each step sees. After each wave, a progress reflection compares results to the outline and may add or revise steps. Replanning runs a fixed number of iterations, then the plan freezes.
- A variant without the dependency graph, running steps one at a time with cumulative context, took 222 minutes instead of 47 and scored lower.
- Anthropic uses a more dynamic lead-agent loop but saves the plan to external memory, since truncation at the context limit would otherwise lose it.
- OpenAI describes the opposite philosophy. Its agent is trained end-to-end on hard browsing tasks rather than built as a graph of operations with models at the nodes. That's not available to a framework built on API models. One RL survey notes that training whole stacks end-to-end remains impractical, so most work trains a single planner against core tools.

**Parallel vs. sequential.** Parallelize independent reads. Serialize anything where one step consumes another's conclusions, or where outputs must stay mutually consistent.

- Anthropic found multi-agent setups excel at breadth-first, heavily parallel work and fit poorly when agents need shared context or have many dependencies.
- Cognition's critique is that actions carry implicit decisions, so independent workers can make conflicting ones. LangChain's reconciliation is that conflicting writes cause far worse outcomes than conflicting reads. Its own research agent confines multi-agent work to research and writes the report in one shot, after parallel section-writing performed poorly.
- The cost is real. Anthropic measured about 4× chat's token use for single agents and about 15× for multi-agent systems, so it only pays off on high-value queries.

**Depth.** Agents are poor at calibrating their own effort.

- Anthropic embedded explicit scaling rules in prompts, roughly 1 agent with 3–10 tool calls for fact-finding, 2–4 subagents with 10–15 calls each for comparisons, and 10+ subagents for complex research, after early versions spawned as many as 50 subagents for simple queries.
- A stronger idea is Salesforce's step-level termination criteria, declared before execution. Ablating them dropped actionability from 82.1 to 73.7, and agents made far fewer tool calls (327 → 224 on public search), consistent with stopping early.
- I'd combine three stop conditions: criteria met, hard budget hit, or no materially new information in the last few searches. The last is my suggestion.
- Depth cuts both ways. Token usage alone explained about 80% of performance variance on BrowseComp in Anthropic's analysis, yet diagnostic work on search agents finds both over-search (redundant searching) and under-search (premature termination).

## B. The agentic search loop

**The loop.** In practice it runs plan, fan out, then reason over every result before the next call within each worker.

- Anthropic's subagents use interleaved thinking after tool results to evaluate quality, identify gaps and refine the next query. The lead synthesizes and decides whether to spawn more subagents or refine its strategy.
- TTD-DR keeps an evolving draft that guides research direction while new retrieval revises it.
- The well-evidenced trigger for another iteration is an unmet completion criterion or a gap found in progress reflection (the Salesforce design above). I'd add three triggers, as suggestions rather than findings: an unresolved contradiction, a load-bearing claim resting on one source, and a claim supported only by low-tier sources.

**Avoiding context stuffing.** The governing principle is the smallest set of high-signal tokens, since recall accuracy degrades as context grows ("context rot"). Four patterns work:

- **Isolated workers.** A subagent may use tens of thousands of tokens exploring but returns a distilled summary, often 1,000–2,000 tokens. LangChain's sub-agents prune irrelevant material before reporting back.
- **A reading step between raw pages and reasoning.** Search-o1 analyzes retrieved documents in a separate module before injecting them into the reasoning chain. CRAG uses a lightweight evaluator to grade retrieval as correct, ambiguous or incorrect. Good retrievals are refined into key knowledge strips; bad ones are discarded in favor of web search.
- **Handles, not payloads.** Anthropic has subagents write outputs to external storage and return lightweight references, which prevents information loss and cuts token overhead.
- **Recall-first compaction.** Tune the summarizer for recall first, then precision. Clearing stale tool results is one of the safest light-touch forms.

The risk is that compression is lossy, and the summary becomes the new ground truth. The evidence on depth is sobering. In a 2026 preprint, fact-check accuracy of cited statements fell about 42% on average as the tool-call allowance rose from 2 to 150. One model went from 79% to 17% and another from 80% to 58%, while link validity and relevance stayed above 92%. That is two models and LLM judges, but it's the failure shape you'd predict from stuffing.

This doesn't contradict Anthropic's BrowseComp finding, because BrowseComp rewards locating a hard fact while the preprint measures whether claims in a long report are supported. My reading is to spend tokens on parallel, isolated exploration but keep what the synthesizer sees small and verified. Concretely, have workers return claim–evidence records (claim, URL, extracted span, date, source type) rather than free prose, so the synthesizer can re-check instead of trust.

**Contradictions.** Two findings make naive handling dangerous.

- Models overrode correct prior knowledge in favor of wrong retrieved content over 60% of the time in ClashEval, though less often as the content became more blatantly unrealistic.
- Retrieval-augmented models tend to follow majority rule and favor evidence consistent with their own memory.

So don't resolve conflicts by counting sources, since syndication and SEO duplication inflate counts. Dedupe by origin instead (my addition). MADAM-RAG offers a useful taxonomy: conflict can come from ambiguity, misinformation or noise, and the right behavior differs. Ambiguous queries can get multiple answers while misinformation and noise are filtered, though a substantial gap remains when evidence is imbalanced.

A practical pattern is a contradiction register of (claim, sources, conflict type, resolution). Resolve by primary source, date or scope, and report both sides with a confidence note if it stays open. The default is worse: audits find deep-research systems frequently give one-sided, highly confident answers to debate-style queries. That's why STORM's perspective-first design is structurally useful.

## C. Quality mechanisms

**Preventing training-data-only answers.** The failure is documented.

- Outcome-only supervision tends to produce search agents with low retrieval recall, whose correct answers come from parametric memory rather than retrieved evidence. They look fine in familiar domains and fail when their knowledge runs out.
- Search agents also struggle to recognize the limits of their own knowledge, so searching needlessly and underusing search are both common.

An "always search" instruction is therefore weak. Structural enforcement would look like this (my suggestions):

1. Make the unit of work a claim–evidence record, and treat any claim without a retrieved span as "recalled," which must be verified or flagged.
2. Route time-sensitive claims (officeholders, versions, prices, "latest") through retrieval by rule.
3. Borrow ClashEval's protocol (observe whether the model prefers modified evidence or its own prior) as a cheap diagnostic. Where prior and evidence disagree is where verification effort should go.

**Internal quality checks.**

- Anthropic's LLM judge scored factual accuracy, citation accuracy, completeness, source quality and tool efficiency. A single call with one prompt, returning 0.0–1.0 scores and a pass/fail grade, was more consistent than multiple judges. They started with about 20 queries and used human testers for what automation missed.
- TTD-DR applies judge-and-revise loops to each component (plan, query, answer, final writing), not just the end product.
- The most important idea comes from DREAM: *capability parity*. A verifier that can't search can't catch stale facts or well-cited falsehoods. In a small controlled test, a citation-alignment metric stayed near 100% as false-but-cited claims were swapped in, while an evaluator that searched independently tracked the error rate. It also issues neutral search queries and extracts both supporting and opposing passages to avoid confirmation bias. A self-check without tools mostly re-confirms the model's beliefs.

**Source reliability.** I found no widely adopted analogue of clinical evidence grading in the public literature, only heuristics.

- OpenAI has acknowledged that its deep research may struggle to distinguish authoritative information from rumors and is weak at conveying uncertainty accurately.
- Anthropic found early agents consistently preferred SEO content farms over authoritative but lower-ranked sources like academic PDFs, and fixed it with source-quality heuristics in prompts.
- DREAM scores cited domains into authority tiers using an LLM judge, but LLM credibility ratings of news outlets align only moderately with human experts.

I'd grade by role as well as domain: primary or official record, peer-reviewed, reputable secondary, vendor or marketing, forum or blog, and recalled-from-memory. Record the date and the number of independent origins, and surface the grade in the report. "Single vendor source" is itself useful information.

**Fluent but ungrounded.** One audit found some deep-research systems with over 70% of statements unsupported by their own cited sources, and citation accuracy ranging from 40–80%. The 2026 attribution study found links nearly always work and pages are on-topic, yet fact-check accuracy ranged only 39–77%. Surface signals are nearly useless as quality gates. The gate that works is claim-level entailment against fetched source text, plus an external check for mutable facts, run by a stage separate from the writer.

## D. Synthesis and report generation

**The transition.** Leave the loop when per-step criteria are met, or when budget or iteration caps hit.

- Salesforce freezes the plan after a fixed number of replanning rounds and executes the rest.
- Anthropic's lead exits when it judges information sufficient and hands everything to a citation stage.
- LangChain's supervisor moves to report generation once it deems the responses sufficient.

One addition from me: if you exit on budget, carry unmet criteria into the report as stated gaps rather than papering over them.

**Draft vs. one-shot.** STORM separates a pre-writing stage (research and outline) from a writing stage with citations. LangChain writes once from the brief and findings. TTD-DR keeps an evolving draft as the backbone, reporting more coherent writing and less information loss. These combine well. Write once (no parallel section writers) from an outline and evidence ledger, then run a revision pass targeted at flagged gaps.

**Citations.** Anthropic runs a dedicated citation agent after the research loop to identify where claims should be attributed. Post-hoc attribution is cheap, but the attribution literature reports a consistent coverage-versus-correctness tradeoff between generation-time and post-hoc citing. If you can afford it, carry provenance from the start (claim ID to source span), so citation is a lookup-and-check rather than a reconstruction. Measure two things jointly: how many verifiable claims are cited, and whether the cited text supports them. DREAM combines these as a harmonic mean, so citing nothing and citing everything wrongly both score badly. DeepResearch Bench's FACT framework likewise reports citation accuracy alongside effective citations.

**Complete vs. surface-level.** DeepResearch Bench scores comprehensiveness, insight, instruction-following and readability. Salesforce's rubric operationalizes "surface-level" well: a five-point scale running from no answer or generic themes up to complete, customer-specific answers with names, numbers and timelines, where "actionable" means a score of 4 or higher. A checklist, partly mine:

- Every sub-question is answered with specifics.
- Unfound items are stated.
- Disagreements are shown with their resolution.
- Fact is separated from inference.
- Mutable claims are dated.

DREAM's warning applies: fluent, well-cited reports can score well despite obsolete information or flawed logic, so "feels complete" is not a metric.

## E. Failure modes and edge cases

Common failure modes and what addresses them:

- **Over-investment, endless hunts for nonexistent sources, and duplicated work** (all seen in Anthropic's early versions). Use effort-scaling rules, hard budgets, a "not found after N distinct strategies" exit, and specific delegation briefs.
- **Premature stopping and uneven coverage** (per Salesforce). Use per-step completion criteria and coverage checklists.
- **Over- and under-search** (per the search-boundary work). Tie budgets to criteria rather than raw call counts.
- **Context rot and explosion** (per Anthropic's context-engineering guidance). Use isolation, handles and claim-level notes.
- **Low-tier source preference** (per Anthropic). Use source tiering and prefer primary sources.
- **One-sided, overconfident synthesis** (per DeepTRACE). Use perspective discovery and a contradiction register.
- **Cited-but-unsupported claims** (per the attribution study). Use a tool-equipped entailment gate.
- **Stale facts** (per DREAM). Date claims and use rule-based retrieval for mutable facts.
- **Compounding errors in long runs.** Anthropic stresses resuming from checkpoints, letting agents adapt to failing tools, and full production tracing.

**Sparse or conflicting current information.** Treat "not found" as a first-class result. Log the queries and strategies tried, and distinguish "no evidence exists" from "I couldn't find it." Tie confidence labels to evidence structure (independent primary sources, recency, agreement) rather than fluency. This matters because overconfidence is common. DeepTRACE counts an answer as overconfident when it is one-sided yet maximally confident, and OpenAI flagged calibration as a weakness. RL-trained research agents have shown emergent cross-validation and honesty when unable to find definitive answers. My inference is that your evals should reward honest negative results, or you'll push the system to always produce an answer.

**Training data vs. web evidence.** Both directions fail. ClashEval found models adopting wrong retrieved content over 60% of the time even when their prior was correct, while another study found stronger models persistently favoring faulty internal memory despite correct evidence. Neither "web always wins" nor "prior wins" is safe. A policy I'd implement:

1. For mutable facts, dated retrieved evidence outranks priors. DREAM's tests show what staleness costs: key-information coverage fell from about 79 to 45 with a January 2025 knowledge cutoff and to 22 with a January 2024 cutoff, while a standard report-quality benchmark barely registered it.
2. For stable facts, a lone low-tier page shouldn't override a strong prior without higher-tier corroboration.
3. Log every prior–evidence conflict, escalate to a primary source, and state in the report which side you took and why.

If it would help, I can turn this into a concrete reference architecture next: the claim–evidence record schema, stage interfaces, and an eval rubric per stage.