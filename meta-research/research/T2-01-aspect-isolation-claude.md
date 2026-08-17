# Aspect-Isolation vs. Integrated Research: An Evidence-Based Verdict

**Document:** T2-01
**Question under review:** Does mandating one dedicated session per research topic actually produce better outputs than integrated, multi-topic research — and under what conditions?
**Status of the current rule:** The "Aspect-Isolation Law" was derived from observational experience across 6 projects. It has not been empirically tested. This report evaluates it against the wider evidence base.

---

## Bottom line

**The absolute version of the Aspect-Isolation Law — every distinct topic gets its own dedicated session, no exceptions — is not supported by the evidence and should be retired.** The concern behind it is real: finite search/attention budget is a genuine, measurable constraint on AI research quality, not a superstition. But the specific remedy the Law prescribes (full session-level isolation, applied uniformly) is a blunt instrument for a problem that the best-documented evidence says is actually about *context architecture* — how much irrelevant material competes for attention, how clearly a task is scoped, and whether findings get synthesized — not about how many separate conversations you open.

The most direct evidence available is not a lab study of "isolated vs. bundled chat sessions" (that study doesn't exist yet — see §8). It is the revealed design choice of every frontier lab that has built a serious AI research agent. Anthropic, OpenAI, Google, and the open-source "deep research" ecosystem have all converged on the same architecture: decompose a query into clearly-bounded, independent sub-investigations, run them with protected context (often in parallel), and then **explicitly synthesize** the results into one report — all inside a single overarching task. None of them isolate topics into disconnected sessions with no synthesis step. That convergence is itself strong evidence about what actually works at the frontier, and it points toward *conditional decomposition with synthesis*, not blanket isolation.

**Confidence level:** High confidence that an unconditional isolation rule is wrong. Moderate confidence in the specific conditioning variables and decision framework proposed below. Low confidence in any precise numeric threshold. See §9 for the full graded breakdown.

---

## 1. How this was investigated

This report draws on four literatures that rarely get put in the same room: cognitive load theory and task-switching research (psychology), information foraging theory (information science), long-context and multi-task performance research (AI/ML, largely 2023–2026), and the published engineering practices of production AI research systems. Where evidence is direct and quantified, it's presented as such. Where it's analogical — a finding about human working memory, or about scientific citation patterns, offered as a *lens* rather than a *proof* — that's flagged explicitly. The goal is to keep the evidence tiers visible rather than blending a Stanford NLP benchmark and a business-book anecdote into equal-looking bullet points.

One framing choice matters throughout: **"session" is not the same variable as "context."** A single chat session can be internally well-architected (clear sub-scopes, bounded context per sub-task, an explicit synthesis pass) or a mess (everything dumped in one undifferentiated stream). Conversely, "separate sessions" can still be a well-designed decomposition, or can just be fragmentation with extra steps. Most of the disagreement in this space collapses once that distinction is made explicit — see §4.

---

## 2. The case *for* isolation: what the Law gets right

The Aspect-Isolation Law's three claimed mechanisms each have real evidentiary support. This section steelmans them.

### 2.1 Search/token budget is a measured performance driver, not a hunch

This is the strongest piece of evidence in the Law's favor, and it comes from the most authoritative possible source: Anthropic's own account of building its multi-agent Research system. In their internal analysis of the BrowseComp evaluation (a benchmark for agents that must locate hard-to-find information), three factors explained 95% of performance variance — and **token usage by itself explained 80% of that variance**, with tool-call count and model choice accounting for the rest. Put plainly: how much budget an agent gets to spend on a question is, empirically, the single largest determinant of whether it finds the answer. This directly validates the Law's premise #1 — finite budget divided across topics is a real constraint with a measured effect, not an assumption.

### 2.2 LLMs measurably lose the plot in long, cluttered contexts

"Context window attention diffuses across multiple topics" turns out to be an understatement of a well-documented, actively-researched class of phenomena:

- **Lost in the Middle** (Liu et al., Stanford/University of Washington, published in *TACL*): across multi-document QA and key-value retrieval tasks, LLMs show a U-shaped accuracy curve — strong at the start and end of context, and more than 30% worse when the relevant information sits in the middle. This has been replicated across many model families and persists even in models built for long context.
- **Context rot** (Chroma Research, testing 18 frontier models including GPT-4.1, Claude 4, Gemini 2.5, and Qwen3): performance degrades non-uniformly and non-linearly as input length grows — well before the context window is anywhere near full. Critically, the researchers note their tests used *simple* tasks, and explicitly predict that degradation would be **worse** for tasks requiring synthesis or multi-step reasoning — i.e., exactly the kind of evaluative research work this framework cares about.
- **Multi-needle retrieval** (extending Greg Kamradt's "Needle in a Haystack" benchmark; later formalized in the RULER and "Needle Threading" benchmarks): accuracy degrades as the *number* of facts an LLM must simultaneously track increases, and the degradation sets in at *shorter* context lengths as needle count grows (in one test, ~25K tokens with multiple needles vs. ~73K tokens with a single needle). One evaluation vendor reports multi-needle accuracy dropping from ~95% to ~60%. Reasoning over retrieved facts is consistently harder-hit than simple retrieval.

These are AI-specific, mechanistic findings — not an analogy borrowed from human psychology. They support a real version of "attention dilution," though note the mechanism is about **context volume and simultaneous item-count**, not literally about topic bundling (more on this distinction in §4).

### 2.3 Bundling instructions into one prompt measurably dilutes focus

A separate, growing body of 2025–2026 research tests exactly the scenario the Law worries about — multiple tasks or instructions in a single prompt:

- A six-task, cross-model study found that combining NLP tasks incrementally into one prompt causes performance degradation — though critically, this was **highly architecture-dependent**: one model stayed stable across all tasks while others showed severe collapse on fine-grained tasks (see §5.4).
- Work on multi-instruction following (building on the "BatchPrompt" line of research) found GPT-3.5-turbo and GPT-4-turbo answered multiple simultaneous instructions measurably worse than the same instructions answered one at a time, with missed instructions increasing further when instructions were given in randomized rather than sequential order.
- A 2026 study of prompt optimization for LLM judges found that asking a model to give feedback on multiple criteria jointly cut feedback specificity by 59%, and that combining independently-optimized single-task instructions into one prompt measurably hurt downstream correlation with ground truth.
- Separately, multi-*turn* degradation is also documented: one study found LLMs performed 39% worse and were 112% less reliable when a fully-specified task was revealed piecemeal across turns rather than given all at once; another found instruction-adherence fell monotonically across turns (even a frontier reasoning model dropped from 88% to 71% accuracy between turn one and turn three).

**Direct answer to "is search budget dilution documented or assumed":** It is genuinely documented — across at least four independent research lines (long-context position effects, multi-fact retrieval, multi-instruction prompting, and multi-turn instruction-following) — that overloading a single context with more simultaneous demands than it can cleanly track degrades output quality. That part of the Law's logic is empirically real. What is *not* documented is the specific causal chain the Law assumes — that this is best fixed by session-count isolation specifically, rather than by other forms of context management. See §4.

---

## 3. The case *against* blanket isolation

### 3.1 Isolation is not free — and the Law's rationale doesn't price the cost in

The same Anthropic engineering account that validates the budget-dilution concern also quantifies isolation's cost: agents that decompose work across separate contexts use roughly **4x** the tokens of a standard chat interaction, and full multi-agent decomposition uses roughly **15x**. Anthropic's own conclusion is that this overhead means decomposition "requires tasks where the value of the task is high enough to pay for the increased performance" — i.e., it's a targeted tool for high-stakes, complex work, not a default applied uniformly. Their early prototypes also surfaced a concrete over-isolation failure mode: agents spawning 50 sub-agents for simple queries, burning budget on decomposition the task never needed. A rule that isolates *every* distinct topic regardless of its complexity or stakes will systematically recreate this failure mode on the simple end of the distribution.

### 3.2 Splitting content that needs joint interpretation makes things worse, not better

Cognitive load theory contains a finding that cuts directly against naive fragmentation: the **split-attention effect**. When two information sources that must be mentally integrated to be understood (classically, text and a diagram) are physically or temporally *separated*, learners perform worse than when the same information is presented together — because separation forces costly extra work to mentally recombine what should have been unified. The fix researchers recommend is integration, not further separation. This is a direct, well-replicated finding (going back to Chandler & Sweller's original 1992 studies, with ongoing replication work since) about the cost of fragmenting *related* material. If two research topics are genuinely coupled — one's findings change how you'd investigate the other — isolating them into separate sessions with no automatic cross-reference recreates exactly the condition this literature warns against: the human (or downstream system) now has to do the integration work manually, later, with less context than the AI had in the moment.

### 3.3 "Attention residue" is about incompleteness, not about bundling

Task-switching costs are real and well-established (see §2), but the specific mechanism identified by Sophie Leroy's foundational 2009 research on "attention residue" is more precise than "switching is bad": residue is driven overwhelmingly by **leaving a task incomplete or unresolved** before moving to the next one. When a task is brought to a clear stopping point before switching, the residue effect — and its performance cost — is substantially reduced. This is an important correction to how the Law's rationale is usually stated. It doesn't argue for isolating every topic into its own session; it argues for **finishing (or reaching a clean checkpoint on) one line of inquiry before moving to the next**, which is entirely achievable *within* a single well-structured session, and is in fact exactly what a good sequential or parallel-with-synthesis workflow already does.

### 3.4 Decomposition trades one failure mode for another: error accumulation

The multi-hop question-answering literature — a research area built entirely around decomposing complex questions into sub-questions — has documented a specific cost of decomposition itself: aggregating answers from independently-solved sub-parts introduces **error accumulation** that compounds across the pipeline, and decomposition-based approaches frequently underperform more holistic (non-decomposed) approaches on raw accuracy despite being more interpretable. A 2026 study on decomposed prompting formalizes this as two competing effects: decomposition helps when a model has the relevant knowledge but lacks the executive capacity to organize it unprompted (the "scaffolding" case), but **hurts** when a model was reaching correct answers through an associative or holistic route that explicit decomposition actually disrupts (the "shortcut" case). The practical implication: isolating and later stitching together topic-by-topic findings is not a strictly safer default than integrated investigation — it introduces its own, different risk of dropped or miscombined information at the seams.

### 3.5 Over-isolation risks the systemic blind spot

This is more analogy than controlled study, but it's a well-known one and it maps precisely onto a question this review was asked to investigate directly. In scientometric research, Uzzi, Mukherjee, Stringer, and Jones (*Science*, 2013; 17.9 million papers analyzed) found that the highest-impact scientific work is not the most novel work in isolation — it's work that combines a foundation of conventional, well-established grounding with an *intrusion of unusual, cross-domain combinations*; such papers were roughly twice as likely to be highly cited. The organizational-behavior literature on "silo effects" tells the same story from the failure side: a canonical account of the 2008 financial crisis (Gillian Tett's *The Silo Effect*) describes how regulators and economists, each looking reasonably within their own narrowly-scoped domain, collectively missed an emergent, cross-cutting systemic risk that was only visible by looking across domains. Neither of these is a study of AI research pipelines. But they're directly on point for the question this review was asked to answer — "does fragmenting research into isolated sessions miss systemic interactions, cross-cutting insights, or emergent patterns?" — and the answer from both literatures is: that is exactly the known failure mode of fragmentation, independent of whether the fragments themselves are individually well-researched. Tellingly, Anthropic's own write-up reports the opposite, synthesis-preserving design paying off in practice: users of their multi-agent Research system reported it surfaced "research connections they wouldn't have found alone" — a benefit that depends on the system *not* leaving each sub-investigation permanently isolated.

---

## 4. The reconciling insight: the Law is solving for the wrong variable

Put the two sections above side by side and a pattern emerges: **the real, well-documented mechanism is context overload (too much simultaneous, undifferentiated, or poorly-scoped material competing for a finite attention/token budget) — not session count.** "One topic per session" is *one* crude way to bound context volume. It is not the only way, and the evidence suggests it is not the way the most sophisticated systems actually use.

Look at what Anthropic actually built when they set out to solve precisely this problem — a system that has to research complex, multi-part queries under a real, measured token/attention budget constraint (§2.1). Their architecture is **not** "isolate every distinct sub-topic into a disconnected user-facing session." It's an orchestrator-worker pattern, all within one task: a lead agent decomposes the query into clearly-scoped, non-overlapping sub-investigations; spins up several sub-agents in parallel, each operating in its *own protected context window* (this is where the real isolation happens — at the context-architecture level, not the session level); each sub-agent runs its own tool calls; and then the lead agent **explicitly synthesizes** the distilled findings into one report, with a final pass to verify citations. Every other serious "deep research" system — OpenAI's Deep Research, Google's Gemini Deep Research, Alibaba's Tongyi DeepResearch, and open implementations like STORM and GPT-Researcher — follows the same basic shape: decompose, investigate (often in parallel, sometimes issuing dozens of searches), synthesize, all inside one overarching task. This is not a coincidence or a shared blind spot; it's convergent evolution toward what the underlying constraints actually reward, and it's the closest thing available to a real-world, high-stakes, expensively-tested answer to this review's question.

Two further data points from the same source sharpen this. First, Anthropic explicitly names the condition under which their decomposition approach does *not* work well: **"domains that require all agents to share the same context or involve many dependencies between agents"** — their example is that most coding tasks have far fewer genuinely parallelizable sub-tasks than research does. This is a direct, first-party statement that decomposition (their word for what the isolation framework calls "aspect separation") is conditional on the *interdependency* of the sub-tasks, not a universal law. Second, their single clearest documented failure mode was not under-isolation — it was **poor scoping**: when the lead agent gave a vague instruction like "research the semiconductor shortage," sub-agents duplicated each other's work (one investigated the 2021 automotive chip crisis while two others redundantly covered the 2025 supply chain) because the task boundaries weren't clear. The fix was better task descriptions with explicit, non-overlapping objectives — not more or fewer sessions. **The quality of the decomposition (how clearly each piece is scoped) is doing more work than the mere fact of separation.** A framework that mandates isolation without also mandating clear, non-overlapping scope definitions is solving half the problem; a framework that integrates topics without any internal structure is solving none of it.

---

## 5. What actually determines the answer: five conditioning variables

The evidence converges on isolation and integration each being correct under different, identifiable conditions — not on either one being universally right. Five variables recur across the literatures reviewed:

**1. Relatedness / interdependency of the topics.** This is the variable Anthropic's engineering team names explicitly as the deciding factor for their own architecture. Low-interdependency topics (truly independent facts, entities, or questions) are the case multi-agent decomposition was built for and where it delivered a 90.2% improvement over a single agent on breadth-first queries. High-interdependency topics — where investigating one changes what you'd look for in another, or where a shared context is genuinely needed — are the case Anthropic says decomposition is a poor fit for, and it's also where the split-attention-effect research (§3.2) says forced separation actively hurts.

**2. Task type: convergent/factual vs. evaluative/synthetic.** Simple fact-finding tolerates bundling far better than judgment work. This shows up twice in the evidence: the context-rot researchers explicitly predict worse degradation for synthesis and multi-step reasoning than for the simple retrieval tasks they tested, and the multi-needle literature independently finds that *reasoning* over retrieved facts degrades faster than *retrieving* them. If the deliverable is "list five facts," bundling costs little. If it's "form a judgment about how these things interact," bundling costs more — but so does isolating them with no synthesis step, since forming that judgment is the part that requires seeing them together.

**3. Topic count and overall task complexity.** Anthropic's system encodes an explicit scaling rule, calibrated against their own evaluations: simple fact-finding gets one agent with 3–10 tool calls; direct comparisons get 2–4 sub-agents with 10–15 calls each; complex, many-faceted research gets 10+ sub-agents with clearly divided responsibilities. The shape of this rule — effort and decomposition scale *with* complexity, rather than being fixed at "one topic, one unit" regardless of how small or large that topic is — is a more defensible design pattern than a flat per-topic rule, and it's backed by measured performance data rather than the 6-project observational base the current Law rests on.

**4. Model capability.** The multi-task-prompting degradation research is unambiguous on this point: it is **architecture- and capability-dependent**, not universal. In one controlled six-task study, one model family stayed essentially stable as tasks were added while others collapsed on fine-grained tasks. A September 2025 study explicitly formalizing "cognitive load" for LLMs found the same pattern at the extreme: small open-source models scored 0% even on *clean, single-topic* high-complexity reasoning, while a more capable model showed strong baseline performance with only gradual, if statistically real, degradation under added load. A rule tuned to the framework's weakest supported model will be too conservative for its strongest one, and vice versa.

**5. Stakes vs. overhead cost.** Isolation and decomposition are not free (§3.1) — realistically 4–15x more expensive in tokens/effort than a single unified pass, per Anthropic's own production data. That overhead is justified for complex, high-value research where the performance gain pays for it, and not justified for quick, low-stakes lookups. A blanket rule that isolates a two-minute fact-check exactly as aggressively as a multi-week strategic research project is spending the same resource multiplier on both, which the evidence suggests is a poor trade on the cheap end.

---

## 6. Where's the line? Optimal granularity

Information foraging theory (Pirolli & Card) offers the cleanest theoretical frame for this, even though it was developed to describe human web-browsing behavior rather than AI agents. Its core result, the **marginal value theorem**, says a forager should stay in an information "patch" as long as the patch's rate of return exceeds the average rate available elsewhere in the environment — and critically, that decision depends on the *travel cost* of switching to a new patch. When switching is cheap, the theory predicts (and observes) that foragers sample many patches briefly — what later researchers called "information snacking." When switching is expensive, foragers should stay put longer and extract more from where they already are.

This reframes the granularity question productively: **the right question is not "how many topics is too many for one session," but "what does switching (i.e., isolating into a new session or sub-agent) actually cost in this system, and does the topic's value exceed that cost?"** If your framework's session-switch cost is low — cheap to spin up, trivial to re-inject any needed shared context, and you always run an automatic synthesis pass afterward — you can afford to isolate fairly liberally, closer to Anthropic's high end of 10+ parallel sub-investigations for genuinely complex, many-faceted work. If switching is expensive in your system — no automatic synthesis, each new session needs manual re-briefing, related context has to be copy-pasted by a human — the same theory says you should isolate far more conservatively, because you're paying real "travel cost" every time and eating a proportionally bigger chunk of the value.

The two failure modes bookend the range and are both documented above: **too integrated** looks like the split-attention effect and multi-instruction dilution — related-but-crammed content that never gets the focused pass it needs, or genuinely independent topics fighting for the same limited budget (§2). **Too fragmented** looks like Anthropic's 50-subagents-for-a-simple-query anti-pattern, or the semiconductor-shortage duplication failure — resources burned on decomposition the task didn't need, or coordination failures from unclear scope boundaries (§3, §4). The optimal point sits between them and moves based on the five variables in §5, not at a fixed topic count.

---

## 7. Actionable decision framework

| Signal | Favors **ISOLATION** (separate session / sub-agent, own context) | Favors **INTEGRATION** (one session, internally structured) |
|---|---|---|
| Topic interdependency | Topics are genuinely independent — no shared entities, no finding in one that changes the approach to another | Topics share entities, need cross-referencing, or one's findings should redirect the other |
| Task type | Convergent: fact-finding, enumeration, verification | Divergent: synthesis, judgment, "how do X and Y interact / conflict" |
| Number of topics | Many (roughly 4+), each substantial enough to be its own line of inquiry | Few (2–3), naturally read as facets of one question |
| Stakes / value | High enough to justify a real resource multiplier (Anthropic's own data: ~4–15x tokens) | Low-to-moderate; overhead isn't worth paying |
| Model capability in use | A model shown to be robust to long/cluttered context and multi-task load | A smaller or more interference-prone model, where a focused single pass outperforms a crowded one |
| Switch/setup cost in your system | Cheap to spin up a new session and carry over any needed shared context | Expensive — no automatic context carryover, manual re-briefing required |
| Need for a unified judgment or narrative | Low — a set of separate answers *is* the deliverable | High — the deliverable is one coherent view, and the cross-cutting insight is the point |

**Two rules apply regardless of which side wins:**

1. **Always scope each piece clearly and non-overlappingly before starting.** The single best-documented failure mode in this review (§4) was ambiguous task boundaries causing duplicated work — a decomposition-quality problem, not an isolation-quantity problem. This matters whether you end up with one session or ten.
2. **Always add an explicit synthesis step**, even — especially — when topics are isolated. Every high-performing real-world system reviewed here treats decomposition and synthesis as a pair, never decomposition alone. If the current framework isolates topics into separate sessions with no mechanism that reads across them afterward, that absence is very likely costing more than the isolation itself is saving, per §3.5 and §4.

---

## 8. What we don't know — and how the framework could find out for itself

In the interest of the rigor this review was asked for: **no controlled, peer-reviewed study directly tests "fully isolated single-topic AI research sessions" against "integrated multi-topic sessions" as this framework operationalizes that choice.** The evidence assembled here is real, but every piece of it is either (a) analogical — drawn from human cognitive psychology or scientometrics and offered as a lens, not a proof; (b) mechanistically adjacent — direct LLM research on context length, multi-fact retrieval, and multi-instruction prompting, which bears on the *underlying* dilution mechanism but doesn't test session-boundary-as-such; or (c) revealed-preference — the architecture choices of production systems, which is unusually strong practical evidence (these systems are tested against real usage at scale) but is still not a controlled experiment, and reflects those teams' specific constraints, not necessarily this framework's.

The 6-project observational base behind the current Law sits in the same evidence tier as (c) — real experience, but uncontrolled, unreplicated, and vulnerable to confounds (were the "successful" isolated sessions also the ones on simpler or less-related topics? Were the "diluted" bundled sessions also the ones attempted with less scoping effort?). Neither this report nor the Law it's evaluating can currently answer that.

**A concrete way to close the gap:** take a reasonable sample of past or upcoming research topics — ideally including both clearly-related pairs/sets and clearly-unrelated ones — and run each set both ways: once fully isolated (current default), once integrated with explicit internal scoping and a synthesis pass. Score both against a rubric like the one Anthropic used to evaluate their own system — factual accuracy, completeness, source quality, tool efficiency — plus one criterion the standard rubric misses: a cross-cutting-insight score, rating whether the output surfaced any connection, contradiction, or pattern that only becomes visible when the topics are viewed together. That last criterion is the one most likely to separate the two approaches, since it's precisely what integration can produce and isolation-without-synthesis structurally cannot.

---

## 9. Recommendation and confidence level

**Recommendation:** Retire the unconditional Aspect-Isolation Law. Replace it with a conditional decomposition policy using the five variables in §5 and the framework in §7, defaulting toward isolation for high-value, low-interdependency, factual, many-topic work, and toward integration for related, judgment-oriented, few-topic work — with clear scope boundaries and an explicit synthesis step mandatory in either case.

**Graded confidence:**

- **High confidence** — An absolute, exceptionless "every topic gets its own session" rule is not supported by the evidence and should not remain a hard law.
- **High confidence** — Finite search/attention budget is a real, measured constraint on AI research quality (Anthropic's own data: token usage explains ~80% of performance variance in a hard search benchmark). The Law's founding concern is empirically real, not a myth to be dismissed.
- **High confidence** — Isolation carries real, non-trivial resource and coordination costs (measured at 4–15x token overhead in production multi-agent systems, plus documented duplication/coordination failure modes) that the current framework's rationale does not appear to price in.
- **Moderate confidence** — The five conditioning variables identified (relatedness, task type, topic count/complexity, model capability, stakes) are the right ones and are reasonably prioritized in that order.
- **Low confidence** — Any specific numeric threshold (e.g., "isolate at 4+ topics," "bundle up to 3 related ones"). The evidence supports the general shape of this line, not a validated number for this framework's specific models, tools, and topic types.
- **Acknowledged gap** — No study directly tests this framework's exact choice (disconnected sessions vs. integrated sessions) head-to-head. §8 proposes how to generate that evidence directly rather than continuing to infer it from adjacent literatures.

---

## References

**Cognitive science**
- Chandler, P. & Sweller, J. (1992). The split-attention effect as a factor in the design of instruction. *British Journal of Educational Psychology*.
- Rogers, R. & Monsell, S. (1995). The costs of a predictable switch between simple cognitive tasks. *Journal of Experimental Psychology: General*.
- Rubinstein, J., Meyer, D., & Evans, J. (2001). Executive control of cognitive processes in task switching. *Journal of Experimental Psychology: Human Perception and Performance*.
- Monsell, S. (2003). Task switching. *Trends in Cognitive Sciences*.
- Leroy, S. (2009). Why is it so hard to do my work? The challenge of attention residue when switching between work tasks. *Organizational Behavior and Human Decision Processes*.

**Information science**
- Pirolli, P. & Card, S. (1999). Information foraging. *Psychological Review*.
- Pirolli, P. (2007). *Information Foraging Theory: Adaptive Interaction with Information*.

**AI / long-context and multi-task research (2023–2026)**
- Liu, N. F. et al. (2023/2024). Lost in the Middle: How Language Models Use Long Contexts. *Transactions of the Association for Computational Linguistics*.
- Hong, K., Troynikov, A., & Huber, J. (2025). Context Rot: How Increasing Input Tokens Impacts LLM Performance. Chroma Research.
- Martin, L. / LangChain (2024). Multi Needle in a Haystack (extension of Kamradt's Needle-in-a-Haystack benchmark).
- Hsieh, C. et al. (2024). RULER: What's the Real Context Size of Your Long-Context Language Models?
- Needle Threading: Can LLMs Follow Threads through Near-Million-Scale Haystacks? (arXiv 2411.05000)
- Degradation of Multi-Task Prompting Across Six NLP Tasks and LLM Families (2025), *Electronics* (MDPI).
- Mosaic-IT: Cost-Free Compositional Data Synthesis for Instruction Tuning (arXiv 2405.13326) — multi-instruction degradation findings.
- When Gradients Collide: Failure Modes of Multi-Objective Prompt Optimization for LLM Judges (arXiv 2605.26046).
- Laban, P. et al. (2025). Sharded simulation / multi-turn instruction degradation studies; Multi-IF benchmark (He et al., 2024).
- Cognitive Load Limits in Large Language Models: Benchmarking Multi-Hop Reasoning (arXiv 2509.19517).
- Decomposed Prompting Does Not Fix Knowledge Gaps, But Helps Models Say "I Don't Know" (arXiv 2602.04853).

**AI research-agent architecture (industry practice)**
- Anthropic (June 13, 2025). How we built our multi-agent research system. *Anthropic Engineering*. https://www.anthropic.com/engineering/multi-agent-research-system
- OpenAI (2025). Introducing Deep Research. https://openai.com/index/introducing-deep-research/
- Deep Research: A Systematic Survey (arXiv 2512.02038) — comparative overview including Gemini Deep Research and Tongyi DeepResearch.

**Cross-domain synthesis and knowledge fragmentation**
- Uzzi, B., Mukherjee, S., Stringer, M., & Jones, B. (2013). Atypical Combinations and Scientific Impact. *Science*, 342(6157).
- Tett, G. (2015). *The Silo Effect: The Peril of Expertise and the Promise of Breaking Down Barriers*.
