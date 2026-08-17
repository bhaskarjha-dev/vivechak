# Evidence Grading and Decision Traceability for AI-Assisted Research Pipelines
### A cross-domain evaluation of evidence classification systems, and a recommended design

---

## TL;DR — recommendation at a glance

1. **Keep the A–E scale.** It's the right order of magnitude and already has team buy-in. The problem isn't the number of tiers — it's that a single letter is being asked to carry information that belongs on several independent axes.
2. **Stop treating the letter as a fixed rank.** Following GRADE's model in medicine, treat A–E as a *starting point* derived from source type, then adjust it with a small number of lightweight modifiers — not a bigger scale.
3. **Add exactly two modifiers: corroboration and recency**, plus one borrowed GRADE flag, **directness**. Do not build a full orthogonal matrix like the intelligence community's Admiralty Code. Real-world use of two independent axes collapses under time pressure — trained analysts do it too (§3, §5).
4. **Add a `verification-method` field, separate from the grade.** This is the single highest-leverage change for a system where an AI does the research: it distinguishes "I fetched this just now" from "I'm recalling what I think this source says." Cap anything in the second category at Grade D, automatically.
5. **Un-conflate Grade E.** "Fabricated / untraceable" and "outdated but once-real" are different failure modes that need different fixes. Fabrication is a verification problem; staleness is a recency problem. Keep them on separate axes instead of dumping both into the bottom tier.
6. **Give every decision a door-type tag** (one-way / two-way, after Amazon's reversibility framework) and let it set the evidence bar. Two-way doors can proceed on a single corroborated Grade B/C source. One-way doors need corroborated Grade A/B evidence and an explicit comparison of alternatives.
7. **Separate decision confidence from evidence grade.** Medicine and intelligence tradecraft both learned this the hard way: a decision can rationally be high-confidence on middling evidence (reversible, time-sensitive) or low-confidence despite excellent evidence (contested trade-offs, high stakes). Don't let the letter grade mechanically set the confidence.
8. **Make decisions and evidence separate, linked, immutable-once-accepted records.** Never edit history — supersede. Auto-generate a searchable index so both humans and *future AI sessions with no memory of this one* can query it without anyone maintaining a wiki by hand.
9. **Grade inline, at the point of citation, never as a separate pass.** Keep required fields under eight. Scale rigor to door-type, not uniformly. The adoption research in §3 is fairly blunt about this: these aren't style preferences, they're what separates systems people keep using from systems abandoned within two years.

The rest of this document is the evidence for these nine calls, plus complete schemas and a worked example.

---

## 0. Where each of your questions gets answered

| You asked | Answered in |
|---|---|
| What can software engineering learn from medicine, intelligence analysis, and legal reasoning? | §2 |
| Is 5-tier the right granularity, or should it be simpler / more nuanced? | §5 |
| Does evidence grading actually change decision behavior? | §3 |
| What are the best ADR formats/tools in 2026, and how should AI-authored ADRs differ? | §2.4, §6.2 |
| How should decisions link to evidence, version over time, and stay searchable across projects? | §6.3–§6.6 |

---

## 1. Scope and method

This report draws on evidence-based medicine (the GRADE framework and the evidence pyramid literature), intelligence tradecraft (the Admiralty Code, ICD 203, and Heuer's Analysis of Competing Hypotheses), legal evidentiary reasoning (standards of proof and citation-validity services), and software engineering's own evidence-based movement and ADR tooling ecosystem. Sources are listed in Appendix B. Where a specific figure or study is cited, it's attributed by author/organization in the text; treat the numbers as illustrative of a pattern, not as precision this document doesn't have standing to claim.

One honest limitation up front: cross-domain analogy is suggestive, not dispositive. Medicine, intelligence, and law all evolved their systems under stakes and feedback loops (patient outcomes, national-security failures, appellate reversals) that software architecture decisions mostly don't share. Where a lesson transfers cleanly, this report says so and explains why. Where it doesn't, it says that too, rather than force-fitting the analogy.

---

## 2. Cross-domain landscape

### 2.1 Evidence-based medicine: GRADE and the evidence pyramid

**GRADE** (Grading of Recommendations Assessment, Development and Evaluation) is the dominant framework in modern clinical guideline-writing, having replaced a fragmented landscape of incompatible "Level I / Grade A"-style systems that didn't mean the same thing across organizations. GRADE's central move, and the one most worth importing, is separating two questions that most evidence systems — including the A–E framework this report is evaluating — quietly merge into one:

- **Certainty of evidence** (formerly "quality of evidence"): how much confidence do we have that the effect estimate is correct? Rated High / Moderate / Low / Very low.
- **Strength of recommendation**: given that certainty, how strongly should we act on it? Rated Strong / Conditional.

These are meant to diverge. GRADE guidance explicitly warns against strong recommendations built on low-certainty evidence, then just as explicitly enumerates specific paradigmatic exceptions — life-threatening situations chief among them — where a strong recommendation from thin evidence is still correct, because the cost of inaction outweighs the uncertainty. The reverse also happens: high-certainty evidence sometimes only supports a conditional recommendation, because patient values and resource trade-offs — not the evidence itself — are what's actually contested. A meta-epidemiological survey of top-tier medical journals found that only about a third of systematic reviews formally assessed certainty of evidence at all, though that share has been rising — a useful reminder that even a well-regarded, decades-old framework has an adoption curve, not a switch-flip.

GRADE also **starts** a rating from study design (randomized trials start high, observational studies start low) and then moves it up or down through a checklist of explicit factors — risk of bias, inconsistency across studies, indirectness, imprecision, and publication bias can each push it down; a large effect size, a dose-response gradient, or evidence that survives plausible confounding can push it up. The rating is a *conclusion*, not a lookup on a fixed table keyed to source type alone.

That last point connects to a second, more specific lesson: the **evidence pyramid is under active revision within medicine itself**, and the revision is directly relevant to how the A–E framework treats "primary sources" as categorically superior. The traditional pyramid put systematic reviews and meta-analyses at the apex, individual studies below, and expert opinion at the base — a clean ordinal ranking by study design. Murad and colleagues' 2016 critique in *Evidence-Based Medicine* argued this ranking breaks in practice: a credible systematic review can faithfully summarize biased underlying trials, and a sloppy systematic review can misrepresent excellent ones, so the *type* of study doesn't fix its trustworthiness — the pyramid's dividing lines should be "wavy," not straight, and systematic reviews are better understood as *tools for consuming* evidence than as a rung above it. A more recent line of critique goes further, arguing systematic reviews and original data-generating research shouldn't share one ranked pyramid at all, since treating synthesis as inherently superior to primary data collection has measurably discouraged investment in new primary research.

**What transfers:** a "primary source" (an official doc, an RFC, a maintainer's own benchmark) is a reasonable *starting point* for trust, exactly as an RCT is a reasonable starting point in GRADE — but it is not an automatic final answer. A four-year-old official doc describing a deprecated API should not outrank a rigorously reproduced, actively-corroborated engineering postmortem just because "official docs" sits atop the current A–E ordering. The fix GRADE model suggests isn't a bigger hierarchy; it's making the hierarchy adjustable.

### 2.2 Intelligence analysis: the Admiralty Code, ICD 203, and ACH

Three separate ideas from intelligence tradecraft are relevant, and they point in a genuinely interesting direction — because two of them pull toward more axes, and the third is a hard warning about what happens when you actually build more axes.

**The Admiralty Code (NATO System)**, developed by British Naval Intelligence in 1939 and since standardized by NATO (STANAG 2511, AJP-2.1), rates raw intelligence on two *independent* scales: **source reliability** (A "completely reliable" through F "cannot be judged," based on the source's track record) and **information credibility** (1 "confirmed by other sources" through 6 "cannot be judged," based on whether this specific claim is corroborated). The two combine into a digraph — B2, F6, A5 — that travels with the information through analysis. The design intent is important: a completely reliable source can still report something wrong (A5 is a real, meaningful combination), and a source with no track record can still happen to report something independently confirmed (E1). Keeping the axes separate is the entire point.

In practice, it doesn't stay separate. A widely cited 1968 analysis of over 1,400 U.S. Army field intelligence reports found that **87% of ratings fell on the diagonal** (A1, B2, C3...), meaning analysts were not actually treating the two scales independently — their judgment of the source contaminated their judgment of the specific claim, and vice versa. A more recent critique (Irwin and Mandel, 2019) goes further, arguing that information-evaluation scales like this one often *mask* analyst subjectivity behind a veneer of structure rather than genuinely disciplining it, across three failure points: how ratings are communicated, what determines them, and where they sit in the broader analytic process. This is not a system built by amateurs — it has 85 years of institutional use behind it — and it still fights a persistent halo effect. That is a strong caution against assuming that adding a second independent axis will actually function independently just because you declared it should.

**ICD 203**, the U.S. Intelligence Community's analytic tradecraft standard, was rewritten in its modern form directly in response to the 2002 Iraq WMD National Intelligence Estimate — a case where thin sourcing, unexamined assumptions, and dismissed alternative explanations produced confident-sounding conclusions that were wrong. Its most transferable idea, tracing back to CIA analyst Sherman Kent's 1964 essay "Words of Estimative Probability," is the same separation GRADE makes from a different direction: **probability language** (how likely is the claim — "likely," "even chance," "very unlikely," mapped to rough numeric ranges) is kept distinct from **confidence level** (how solid is the evidence and reasoning behind that probability judgment — high/moderate/low). ICD 203 also requires that analysts not silently mix vocabulary — if a product uses different scales in different places, it must say so explicitly, rather than let the reader assume consistency that isn't there. And it formally mandates "analysis of alternatives": products must identify and assess plausible competing explanations, not just build the case for one.

**Analysis of Competing Hypotheses (ACH)**, developed by CIA veteran Richards Heuer, operationalizes that alternatives mandate as a matrix: hypotheses across the top, evidence down the side, each cell marked consistent / inconsistent / not applicable. Heuer's central insight is about *diagnosticity*: evidence that's consistent with every hypothesis on the table tells you nothing, no matter how much of it you have, and analysts should prioritize evidence that actually discriminates between live options over evidence that simply piles up. The conclusion is the hypothesis with the *least* evidence against it, not the one with the most evidence for it — a deliberate inversion of the intuitive approach, designed specifically to counter confirmation bias.

Does it work? A 2019 randomized study (Dhami, *Applied Cognitive Psychology*) assigned 50 trained intelligence analysts to use ACH or not on a hypothesis-testing task with known ground truth. The finding worth sitting with: **analysts trained in ACH did not reliably follow all of its steps.** Structure was taught; structure was not fully executed under real task conditions. This is the same adoption gap GRADE shows in medicine and, as §2.4 covers, evidence-based software engineering shows in this field.

**What transfers:** the diagnosticity principle is genuinely valuable and cheap to adopt — reward evidence that discriminates between the options actually being weighed, not evidence volume. The probability/confidence split is the same lesson as GRADE's certainty/strength split, arrived at independently by a completely different field, which is a strong signal it's a real structural insight rather than a medicine-specific quirk. The Admiralty Code's diagonal-collapse finding is the strongest single piece of evidence in this whole report against building a full independent multi-axis grading matrix (see §5).

### 2.3 Legal reasoning: standards of proof and citation validity

Two ideas from law are worth importing, and neither is really about "grading evidence" in the medicine/intelligence sense — they're about how much evidence is *enough*, and how you know evidence is still good.

**Standards of proof scale with the cost of being wrong**, not with a fixed bar. The U.S. legal system runs a ladder of thresholds — reasonable suspicion, then probable cause, then preponderance of the evidence (just over 50%, "more likely than not," the default for ordinary civil cases), then clear and convincing evidence (something courts describe as requiring the claim be "highly probable," used for consequential matters like terminating parental rights or life-support decisions), then beyond a reasonable doubt for criminal conviction, where liberty is at stake. Legal scholarship is direct about the mechanism: as the consequences of an erroneous decision rise, so does the evidentiary bar required before acting — the standard is a function of stakes and reversibility, not a universal constant.

This is, independently, exactly Amazon's well-known **"one-way door / two-way door"** framework from Jeff Bezos's 2016 shareholder letter: irreversible, high-consequence decisions ("Type 1") warrant slow, deliberate, high-evidence processes; reversible, low-consequence decisions ("Type 2") should be made fast, by a small group, on incomplete information — Bezos's own guidance is to act on roughly 70% of the information you wish you had, because waiting for 90%+ is usually just slowness dressed up as diligence. One widely cited example applies this directly to software: choosing a core programming language or framework is a one-way door; adopting a new internal dev tool is usually a two-way door. Three independent domains — law, intelligence, and Amazon's own operating philosophy — converge on the identical principle: **scale rigor to the cost of being wrong, not to a fixed evidentiary bar.** That convergence is stronger evidence for the principle than any single domain's version of it.

**Citation-validity services (Shepard's Citations, Westlaw's KeyCite)** solve a problem this report's brief calls out directly: how do you know a piece of evidence you relied on is still good, once time has passed? Every case a lawyer cites can later be cited *itself* — "Shepardizing" a case produces a forward-citation report showing every later case that treated it, tagged with a simple signal (a red flag for overruled/no-longer-good-law, an orange flag for questioned validity, a yellow flag for limited or distinguished treatment, green for positively followed). Two features of this system are worth copying directly. First, **treatment is often partial, not all-or-nothing** — a case can be overruled on one point of law while remaining good authority on another, and the citator reflects that granularity rather than just flipping a single valid/invalid bit. Second, **the citator tracks forward**, not just backward: it's built to answer "what now depends on this, and has anything downstream challenged it?" — exactly the "blast radius" question a decision registry needs to answer when a piece of evidence turns out to be wrong. It's also worth noting, in the interest of not overselling this: an independent comparison of Shepard's, KeyCite, and a third citator (BCite) found real disagreement between them on what counts as negative treatment roughly a quarter of the time — even a business built on this exact problem for 150 years hasn't achieved perfect consistency. That's a realistic bar to hold this report's own recommendation to.

### 2.4 Software engineering: evidence-based practice and the ADR landscape

Software engineering already ran its own version of GRADE's founding story. **Evidence-based software engineering (EBSE)**, coined by Kitchenham, Dybå, and Jørgensen in 2004–2005, explicitly proposed adapting evidence-based medicine's procedures to technology-adoption decisions, on the observation that practitioners routinely adopt or reject techniques without objective evidence of efficacy — driven instead by what's already familiar. Their own diagnosis of the problem is blunt: **developers tend to prefer mature technologies and generally accepted methods** over what the evidence actually supports, and enthusiasm for a new technique often spreads on the strength of advocacy rather than data. Twenty years on, EBSE is a respected academic subfield with real influence on how systematic literature reviews are conducted in SE research — but it has not become how working engineering teams routinely make day-to-day technology choices. The gap between "well-regarded framework" and "actually changes daily practice" shows up again here, as it did with GRADE and ACH.

**Architecture Decision Records** are the closer analog to this report's D-NNN registry, and the format landscape as of 2026 has genuinely converged rather than fragmented further:

| Format | Origin | Shape | Best fit |
|---|---|---|---|
| **Nygard template** | Michael Nygard, 2011 (Cognitect), building on Kruchten's "decision view" of software architecture | Title / Status / Context / Decision / Consequences | Fastest to write; still the most common starting point |
| **MADR** (Markdown Architectural Decision Records) | ADR GitHub org; current major version 4.0.0, released Sept. 2024 | Context & Problem Statement / Decision Drivers / Considered Options (with pros/cons) / Decision Outcome / Consequences, plus optional metadata (decision-makers, confirmation date) | Teams that want the trade-off analysis visible, not just the outcome — ships in full, minimal, bare, and bare-minimal variants so teams can pick their own overhead level |
| **Y-Statement** | Olaf Zimmermann | One structured sentence: *"In the context of \<use case\>, facing \<concern\>, we decided for \<option\> to achieve \<quality\>, accepting \<downside\>."* | The genuine minimum-viable ADR — good for decisions too small to justify a full record but still worth a searchable trace |

All three remain in active use, and the honest answer to "which is best in 2026" is that the format matters less than the operational discipline around it. A useful, precisely-named failure mode from recent practitioner writing is **"Decision Documentation Theater"**: teams adopt ADRs, write them for a while, and stop within about two years — and even while they're active, they tend to document *trivial* decisions (framework bikeshedding) and *cosmic* decisions ("we will be cloud-native") while skipping the actual load-bearing, ambiguous, hard-to-reverse choices that are precisely what future engineers most need explained. The diagnosis: the format was never the differentiator between teams that sustain the practice and teams that don't — what differs operationally is where the records live, when they're written, who reviews them, and what happens when the world changes and nobody goes back to update them. This is the software-engineering-specific version of the same adoption gap GRADE, ACH, and EBSE all show, and it's the most directly relevant precedent for this report's brief, which explicitly worries about "bureaucratic overhead killing adoption."

On tooling, the field has settled hard on a **"docs-as-code"** pattern: ADRs as Markdown files, stored in the same git repository as the code they describe, so they version, diff, and review the same way code does. **Log4brains** (CLI + hot-reload local preview + static-site publishing + full-text search + a timeline view, with metadata auto-extracted from the file and git history) is the most complete open implementation of this pattern. **adr-tools** (Nat Pryce) is the lighter-weight bash-script equivalent for teams that want Nygard-format records with minimal tooling and supports typed cross-links between records (e.g., "Amends" / "Amended by"). Newer entrants — **dotnet-adr**, **ReflectRally** (adds structured review workflows and ownership), **adr.zone** (multi-format generator spanning Nygard/MADR/Y-Statement and even ISO/IEC/IEEE 42010-flavored output) — mostly compete on authoring ergonomics rather than changing the underlying model. One design rule shows up consistently across all of them and is worth stating plainly because it's load-bearing for §6: **an ADR is treated as immutable once accepted — only its status changes.** New information doesn't rewrite an old record; it produces a new one that supersedes the old, which is kept, not deleted. That is the same instinct behind Shepard's forward-citation model in §2.3, arrived at independently by software tooling authors.

### 2.5 Cross-domain synthesis

The same handful of structural ideas keep reappearing under different names. That convergence, more than any individual framework, is the actual case for adopting them here.

| Structural pattern | Medicine (GRADE) | Intelligence (Admiralty / ICD 203 / ACH) | Law | Software (this report's recommendation) |
|---|---|---|---|---|
| Separate evidence quality from action confidence | Certainty of evidence *vs.* strength of recommendation | Kent's probability language *vs.* confidence level | Standard of proof required *vs.* evidence actually marshaled | Evidence grade *vs.* decision confidence (§6.3) |
| Rank is a starting point, adjusted by explicit factors, not a fixed lookup | RCT/observational start point + up/down-grading domains | N/A directly, but ICD 203 forbids unexplained certainty | N/A directly | A–E source-type start point + modifiers (§6.1) |
| Track whether a claim is independently corroborated | "Consistency" domain (do studies agree?) | Credibility axis (1–6): confirmed by other sources? | Corroborating witnesses/evidence strengthen a claim | `corroboration` modifier (§6.1) |
| Rigor should scale with the cost of being wrong | Five paradigmatic exceptions for strong recs on weak evidence | Effort scales with decision stakes in practice | Standard of proof ladder scales with consequence | `door-type` field on decisions (§6.3) |
| Require explicit comparison of alternatives, not just justification of the choice | Evidence-to-Decision framework weighs benefits/harms/values explicitly | ACH matrix; ICD 203's "analysis of alternatives" mandate | Adversarial process forces both sides to be argued | Rejected-alternatives field (already present; formalized in §6.3) |
| Track whether evidence has gone stale | "Living guidelines" that get re-issued as evidence changes | Less developed here | Shepard's/KeyCite "still good law" signals, forward-citation graphs | `recency` modifier + `review_trigger` field (§6.1, §6.5) |
| Keep the notation compact enough to scan at a glance | High/Moderate/Low/Very low | A–F, 1–6 digraph | Color-coded citator flags (red/orange/yellow/green) | A–E retained, plus short tags (§6.1) |
| Even mature, well-resourced versions of these systems don't achieve full fidelity | ~34% of top-journal reviews formally rate certainty at all | 87% diagonal collapse on a two-axis scale; ACH steps skipped under load | ~25% disagreement rate between competing citators | Design for graceful degradation, not perfect compliance (§3, §6.7) |

That last row matters as much as the others. Every mature analog of this kind of system, in every field examined, is used inconsistently by trained, motivated professionals — and none of them are catastrophic failures for it. The design implication is not "give up on rigor," it's "design for the system to still be useful when it's used at 60% fidelity," because 60% fidelity is what every comparable system in this survey actually achieves in steady state.

---

## 3. Does evidence grading actually change decision behavior?

This is the question the brief is right to insist on answering with evidence rather than assumption, because the honest answer is genuinely mixed — and the mix is informative.

**The negative/mixed case is substantial:**

- **Clinical decision support alert overrides.** Systematic reviews of electronic drug-interaction and clinical alerts consistently report override rates in the **49–96% range**, with several large studies clustering around 85–93%. Crucially, this isn't simply clinicians being lazy: appropriateness reviews find a meaningful share of those overrides are *correct* — the alerts themselves are too broad, insufficiently patient-specific, and fire so often that clinicians can no longer distinguish informative signal from noise. Two mechanisms are named in the literature: **cognitive overload** (too much to process) and **desensitization** (repeated low-value alerts train people to dismiss the next one, including a genuinely important one).
- **ACH under real conditions.** As noted in §2.2, trained analysts randomly assigned to use ACH did not reliably execute all of its steps, even immediately after training, in a controlled study.
- **EBSE adoption.** Twenty years after its founding papers, evidence-based software engineering remains a respected research program more than a routine industry practice; the founders' own stated motivation — that practitioners lean on familiarity over evidence — largely still holds.
- **"Decision Documentation Theater."** As covered in §2.4, ADR practices are frequently adopted, applied to the wrong (easy, obvious) decisions, and abandoned within about two years — not because the template was wrong, but because nothing in the operational setup made maintaining it the path of least resistance.
- **The Admiralty Code's diagonal collapse.** Even a system built specifically to force two independent judgments produces one judgment, in practice, 87% of the time.

**The positive case is real, but narrower and more specific than "structure helps":**

The clearest positive result in this survey is the WHO Surgical Safety Checklist, studied by Gawande's team across eight hospitals on four continents. A **19-item checklist** reduced in-hospital mortality from 1.5% to 0.8% (a 47% relative reduction) and complications from 11% to 7% — and the effect held in both extremely well-resourced and extremely resource-constrained hospitals. This is about as strong a result as behavioral interventions in complex professional settings ever produce.

**What separates the checklist's success from alert fatigue's failure is not "structure" in the abstract — it's a specific, transferable set of design conditions:**

| Condition | WHO checklist (worked) | Drug-interaction alerts (didn't) |
|---|---|---|
| Length | 19 items, fixed | Effectively unbounded, grows with every new rule added |
| Placement | Embedded at a natural pause the team was already taking (before incision) | Interrupts an unrelated task mid-flow |
| Applies to whom | Everyone, including senior surgeons who initially resisted it | Everyone, undifferentiated by alert relevance to that patient |
| Specificity | Fixed set of high-value checks | Broad rule-based triggers with high false-positive rates |
| Forcing function | Requires the team to *say things out loud* to each other | Requires only a click to dismiss |

Translated into design rules for an evidence/decision system: **keep it short and bounded, embed it at a point in the workflow that already exists rather than bolting on a new step, apply it uniformly rather than letting "obviously fine" cases opt out, keep the signal specific enough that it's rarely wrong, and prefer a forcing function that requires a small explicit act (stating a rationale) over a system that can be silently dismissed.** These five properties recur, in one form or another, throughout §6.

**The AI-specific wrinkle.** This pipeline's context — an AI model doing the research and populating the grades — adds a failure mode none of the four domains above had to design around: the person (or model) doing the grading is also the one who might be generating the evidence. Research on **automation bias** — the well-documented tendency to over-trust automated output, studied originally in aviation and now increasingly in AI-assisted contexts — shows it persists even as AI tools improve, and that AI-generated explanations don't reliably fix it, because an explanation being present doesn't mean it's faithful to what actually happened. More concretely: studies of AI-assisted diagnostic reasoning found hallucination rates climbing to 50–82% once even a single incorrect detail entered the input, and separate work on AI-assisted writing and coding has repeatedly found people **skim rather than rigorously verify** AI output, especially once they've built up comfort with the tool. Citation fabrication specifically — an AI confidently attributing a claim to a source that doesn't say that, or doesn't exist — is one of the best-documented LLM failure modes and is exactly the scenario a naive "Grade A: official docs" label is supposed to rule out but, without a verification-method check, cannot.

The implication for this report's recommendation is direct and is the reason §6.2 makes `verification-method` a first-class field rather than a footnote: **in a system where the grader might also be the confabulator, "what tier is this evidence" and "did anyone actually check" have to be two different questions with two different answers, or the grade becomes exactly the kind of unearned confidence GRADE, ICD 203, and the alert-fatigue literature all warn against.**

---

## 4. Assessing the current A–E / D-NNN framework

Evaluated against §2's landscape, the existing framework gets the big call right and has one structural gap plus several smaller, fixable ones.

**What it already gets right:**

- **The order of magnitude is correct.** Five tiers is close to GRADE's four (High/Moderate/Low/Very low) and far more usable day-to-day than a 36-cell Admiralty matrix. The instinct to keep this lightweight was the right one — it just needs orthogonal decomposition, not more tiers (§5).
- **The source-type ordering is sound.** Primary > empirical > vendor > secondary opinion > speculation is a reasonable starting-point ranking, structurally similar to how GRADE ranks RCTs above observational studies as a starting point.
- **The D-NNN registry already does the hardest part of ACH and ICD 203's "analysis of alternatives" mandate** — it captures rejected alternatives and rationale, not just the winning option. Most teams don't do this at all. It needs structuring (linking rejected options to the evidence that ruled them out) rather than inventing from scratch.

**Where it breaks down:**

1. **It conflates source type with in-context reliability.** A Grade A doc that's stale describes something no longer true; the grade doesn't say so. This is precisely the "wavy lines" problem the medical evidence-pyramid literature identified: rank by type alone, without an adjustment mechanism, silently overstates confidence in aging Grade A/B material and understates confidence in fresh, well-corroborated Grade B/C material.
2. **Grade E bundles two unrelated failure modes.** "AI-generated unsupported claims" is a *provenance* problem — there's no real source at all. "Outdated forum posts" is a *recency* problem — there was a real source, it's just old. Bundling them means a grade of E tells you *that* something is unreliable but not *why*, which is exactly the information loss the Admiralty Code's two-axis design was built to prevent (even if it doesn't fully achieve it in practice).
3. **No corroboration signal.** One Grade B benchmark and five independent Grade B benchmarks that agree are indistinguishable today, even though multi-source agreement is precisely what the Admiralty credibility axis, GRADE's "consistency" domain, and ACH's diagnosticity principle are all designed to capture. Real confidence information is being discarded.
4. **No verification-method distinction.** As §3 argues, this is the most consequential gap given who's doing the research. A "Grade A" label currently can't distinguish "fetched the RFC this session" from "recalling what I believe the RFC says," and the second is exactly where citation fabrication lives.
5. **No decay mechanism.** Nothing in the current spec says when a D-NNN needs re-review. Contrast this with GRADE's living guidelines or Shepard's "still good law" signals — both fields treat currency as something that has to be actively tracked, not assumed. Without it, a three-year-old decision built on a framework that's since had two major breaking releases looks exactly as current as one made yesterday.
6. **Grading and deciding aren't formally bridged.** The framework grades evidence and separately tracks decisions but doesn't specify how the two relate — nothing currently stops a Grade E claim from silently underpinning a "high confidence" decision, which is the exact failure GRADE's certainty/strength split and Kent's probability/confidence split were both built to prevent.
7. **No stated mechanism for linkage, versioning, or cross-project search.** The brief describes what the registry *tracks* (context, options, rationale) but not how it's *stored, linked, or queried* — which, per §2.4's tooling review, is where most of the real difference between sustained and abandoned systems actually lives.

None of these are reasons to discard the framework — they're reasons to extend it along specific, narrow axes rather than replace it. That's the case for §5 and §6.

---

## 5. Granularity verdict: why not binary, why not full multi-axis

**Binary (verified/unverified) loses too much.** It collapses "a freshly fetched RFC" and "a vendor's comparison page" into the same bucket as long as both have *a* URL behind them, discarding exactly the source-type signal that lets a reader calibrate trust at a glance. It also has no room for graceful degradation — most real evidence in a fast-moving technical field is legitimately in-between "authoritative" and "worthless," and a binary system forces a false choice at exactly the point where nuance matters most.

**A full orthogonal multi-axis system (source type × recency × replicability, Admiralty-style) is over-engineered for this context, and §2's evidence explains precisely why:** even intelligence analysts with institutional training and 85 years of doctrine behind the method collapse two independent axes into one 87% of the time under real conditions. A third axis doesn't make that problem better — it makes the notation heavier while making genuine independence between axes *less* likely to hold in practice, not more. Every added axis is another judgment call per citation, and §3's evidence on alert fatigue and Decision Documentation Theater is blunt about what happens to systems that add judgment calls without adding proportional value: they get worked around, then abandoned. A full matrix looks more rigorous than a single letter grade. Whether it *is* more rigorous, once you account for how it's actually used, is a separate and much weaker claim — and the cross-domain evidence here suggests the answer is usually no.

**The resolution GRADE models — one ordinal starting grade, adjusted by a small number of checklist-style modifiers, rather than a second independent scale — captures most of the real benefit of multi-axis thinking without either failure mode.** Recency, corroboration, and directness all matter and are worth tracking. But they don't need to be a second full grading scale requiring fresh subjective judgment; they can be lightweight flags computed largely from observable, close-to-mechanical facts: *is there a second independent source that agrees? how old is this claim, relative to how fast this category of fact typically changes? does this evidence directly address the question at hand, or is it evidence about something adjacent?* Those are checklist items, not judgment calls on a whole new scale — which is exactly why they resist the diagonal-collapse failure mode that a second full letter-and-number grade would invite.

| Approach | Information captured | Judgment calls per citation | Failure mode seen in an analog field | Verdict |
|---|---|---|---|---|
| Binary (verified/unverified) | Source type; corroboration; recency; directness — all collapsed into one bit | 1 | Loses the exact signal GRADE, Admiralty, and Shepard's all treat as essential | Too coarse |
| Current 5-tier (A–E), single axis | Source type only | 1 | The evidence-pyramid "wavy lines" problem: fixed rank overstates confidence in aging Grade A material | Right order of magnitude, wrong structure |
| Full orthogonal multi-axis (Admiralty-style) | Source type × credibility × recency × replicability, each independently | 3–4+ | 87% diagonal collapse even among trained analysts; Decision Documentation Theater-style abandonment | Over-engineered for this context |
| **Recommended: tiered grade + lightweight modifiers (GRADE-style)** | Source type (starting point) + corroboration + recency + directness, as checklist flags rather than independent scales | 1 base judgment + up to 3 near-mechanical checks | — (this is what §6 specifies) | Best fit |

---

## 6. Recommended system

### 6.1 Evidence classification: revised tiers plus modifiers

Keep the letters. Redefine what they mean and add three lightweight tags.

**Base grade (starting point, adjustable):**

| Grade | Definition | Notes |
|---|---|---|
| **A** | Primary source — official docs, source code, RFCs, API specs | Starting point only; a stale Grade A should carry a `recency: stale` tag, not be silently treated as current |
| **B** | Empirical evidence — published benchmarks, engineering postmortems | Corroborated Grade B evidence is often more trustworthy in practice than uncorroborated Grade A — the modifiers, not the base letter, should reflect that |
| **C** | Vendor claims — marketing material, comparison pages | Structurally motivated to overstate; treat corroboration as close to mandatory before this supports a one-way-door decision |
| **D** | Secondary opinion — blog posts, tutorials, third-party write-ups | Includes anything recalled from an AI's training memory rather than freshly verified (§6.2) — regardless of what the *original* source would have graded |
| **E** | **Unverifiable / no traceable origin** — narrowed from the current definition | Reserved specifically for claims with no real source behind them at all: model confabulation, or a claim nobody can locate the origin of. **"Outdated" is explicitly removed from this tier's definition** — an old-but-real source keeps its original type-based grade and gets `recency: stale` instead. Conflating "fabricated" with "aged" was destroying diagnostic information (§4) |

**Modifiers (attach 0–3 per citation; computed from mostly observable facts, not fresh subjective judgment):**

- **`corroboration`**: `single` (default) · `corroborated` (≥2 independent sources agree) · `contested` (sources disagree — this itself is important information, not a reason to omit the citation)
- **`recency`**: `fresh` · `aging` · `stale` — computed against a per-category expected half-life that each project sets and tunes for itself. Illustrative starting defaults: package/library-version facts, ~6 months; framework or language behavioral claims, ~18 months; protocol/RFC/standard claims, several years or "until superseded"; architectural pattern/best-practice claims, ~2–3 years. These are deliberately starting defaults to tune, not an external standard — treat them the way GRADE treats its own domains: a documented default that a specific field can override with a stated reason.
- **`directness`** (borrowed directly from GRADE): `direct` (evidence addresses this exact question) · `indirect` (evidence is analogical — e.g., a benchmark of library X in language Y being used to infer something about library X in language Z)

A citation therefore reads, e.g., `B · corroborated · fresh · direct` — still scannable in one glance, but carrying four independent pieces of information instead of one.

### 6.2 Designing for AI authorship

This is the section that answers the brief's specific question about how ADR-style records should differ when an AI, not a human engineer, is doing the research. Four concrete changes:

**1. The `verification-method` field — the single highest-leverage addition in this report.** Every piece of evidence carries a method, independent of its grade:

- `fetched` — retrieved and read via a live tool call in this research session
- `cached` — read from a local/previously-saved copy, not fetched live this session
- `recalled` — asserted from the model's training-data memory, *not* checked this session
- `secondhand` — sourced from something else's description of a primary source (a blog post characterizing what an RFC says), rather than the primary source itself
- `human-provided` — a human supplied the evidence or link directly

**Hard rule: any citation with `verification-method: recalled` is automatically capped at Grade D, regardless of what tier the underlying source would otherwise warrant, and is flagged as needing a live fetch before it can support a one-way-door decision.** This single rule is doing most of the work against the hallucination risk documented in §3 — it makes "I'm confident this is what the docs probably say" structurally distinguishable from "I read this just now," which an unadorned A–E grade cannot do. It's also a direct, practical import of the legal "best evidence rule" (§2.3): the original beats a description of the original, and a court — or here, a research pipeline — should be able to tell which one it's actually looking at.

**2. A named-reviewer checkpoint for AI-authored one-way-door decisions.** Not a full human re-derivation of the research — that would defeat the point of an AI research pipeline — but a required, logged acknowledgment field before an irreversible, entirely-AI-sourced decision flips to `accepted`. Automation-bias research is consistent on this point: an explanation being present doesn't guarantee it's faithful to what actually happened, so the safeguard that matters is a real, if brief, second look by an accountable party — not a more elaborate self-report from the same system that might be wrong.

**3. Consistent confidence vocabulary within a project**, borrowed directly from ICD 203's rule against silently mixing probability language: if "likely" means roughly 60% in one decision record and 90% in another, the vocabulary has stopped communicating anything. Fix the mapping once per project and hold every record to it.

**4. Reward diagnosticity, not volume**, importing Heuer's central ACH lesson explicitly for the AI-research context it's arguably most relevant to. An AI research agent can generate a large volume of superficially relevant, mutually consistent evidence supporting whichever option it framed first — this is a known pattern in AI-assisted work, adjacent to sycophancy and confirmation-seeking. The registry should treat evidence that actually discriminates between the live options as more valuable than evidence that simply piles up in support of one, and decision records for contested or high-stakes choices should show what would have changed the answer, not just what supports it.

A quieter but important fifth point: the file format itself should assume **no persistent memory between sessions**, because that's the actual operating condition of an AI research pipeline unless something else is engineered to preserve it. Plain YAML frontmatter plus Markdown body (§6.6) is preferred over a bespoke database specifically because a future AI session with zero memory of this one needs to be able to parse it cold, and a human reviewer needs to be able to read it without tooling. This is the same "docs-as-code" instinct the 2026 ADR tooling landscape converged on (§2.4), for an overlapping reason — git-friendly plain text survives context loss, staff turnover, and tool changes about equally well.

### 6.3 Decision record schema

Extends D-NNN; every existing field is kept, nothing is removed, several are added and structured.

```yaml
---
id: D-118
title: "Primary datastore replication strategy for order service"
status: accepted              # proposed | accepted | rejected | deprecated | superseded
door_type: one-way            # one-way | two-way — sets the required evidence bar, see below
date: 2026-08-14
confidence: medium            # separate from evidence grade — see rationale in body
evidence_refs: [E-047, E-048, E-051]
supersedes: null
superseded_by: null
amends: null
review_trigger: "re-verify if Postgres major version changes, or by 2027-02"
tags: [postgres, replication, database, order-service]
authored_by: research-agent-3
human_reviewed: true          # required if door_type is one-way (see §6.2)
---

## Context
[the problem and forces at play]

## Options considered
1. Postgres logical replication — evidence: E-047 (B · corroborated · fresh · direct), E-048
2. Debezium/CDC pipeline — evidence: E-051 (B · single · fresh · direct)
3. Application-layer dual-write — no credible evidence located (searched; only Grade E hits)

## Decision
Chosen option: Postgres logical replication.

## Rationale
[why, referencing the evidence grades explicitly]

## Rejected alternatives
- Debezium/CDC: viable, but single-sourced at Grade B; kept as the documented fallback if the
  chosen option underperforms at production scale.
- Dual-write: rejected on evidence grounds alone — no traceable source describes its reliability
  characteristics at our scale, and the failure mode (split-brain) is exactly the kind of risk
  this framework requires corroborated evidence to accept for a one-way-door decision.

## Confidence note
Grade B, corroborated, direct, fresh — strong empirical footing, but not Grade A (no
vendor-neutral spec makes a performance claim). Because this is a one-way-door decision, confidence
is capped at "medium" until a production-scale trial confirms results at our actual data volume —
see review_trigger.
```

Two fields carry the design weight and deserve emphasis:

**`door_type`** operationalizes §2.3's convergent lesson from law and Amazon: it sets the required evidence bar rather than applying one bar uniformly.
- **Two-way door** (reversible, low blast radius): may proceed on a single corroborated Grade B/C source. Speed is the right trade-off, exactly as Bezos's "70% of the data" heuristic argues — a wrong two-way-door call costs a reversal, not a crisis.
- **One-way door** (hard/costly to reverse): requires corroborated Grade A/B evidence, an explicit options-considered comparison, and — if entirely AI-sourced — the human-reviewed checkpoint from §6.2.

**`confidence`** is deliberately *not* a mechanical function of the evidence grade, for the same reason GRADE keeps certainty and strength-of-recommendation separate and ICD 203 keeps probability and confidence-level separate: a decision can be rationally high-confidence on middling evidence when it's cheap to reverse and speed matters, or rationally low-confidence despite strong evidence when the real disagreement is about values or trade-offs the evidence can't settle (which technology is "better" often being exactly that kind of question). The field forces whoever's deciding to state that judgment explicitly rather than let the letter grade make it by default.

### 6.4 Linking and traceability

Evidence and decisions are separate, independently addressable entities — `E-NNN` for evidence, `D-NNN` for decisions, kept as parallel registries so a single well-sourced benchmark can be cited by five different decisions without being re-graded five times, the same way a single Shepard's-tracked case is referenced from every later case that cites it.

- Every evidence record carries a **`used_by`** backlink, auto-maintained by tooling rather than hand-updated, listing every decision that cites it. When a piece of evidence is later found to be wrong or superseded, this backlink is what answers the "blast radius" question instantly — exactly the function Shepard's forward-citation graph serves for case law (§2.3) — instead of requiring a manual search across every decision ever made.
- Decision-to-decision relationships are **typed**, not a single generic link, following adr-tools' precedent (§2.4): `supersedes` / `superseded_by`, `amends` (a partial update — see §6.5), `depends_on` (this decision assumes another one holds), `conflicts_with` (flagged, not silently allowed — see §6.5).

### 6.5 Versioning and decay

**Immutable once accepted.** This is the one point where every domain and every modern ADR tool examined in this report agrees without exception: don't edit history. If new evidence changes the picture, write a new record that supersedes the old one; the old one keeps its original content and gets its `status` updated to `superseded`, preserving the trail of what was believed and why at the time. This is Nygard's original guidance, Log4brains' explicit design rule, and the same instinct behind Shepard's refusal to simply delete overruled cases from the record.

**Partial supersession**, imported directly from Shepard's/KeyCite's most useful feature (§2.3): a decision is often invalidated for one of its stated reasons but not all of them. Use `amends` for this rather than forcing a binary `superseded_by`: D-118 chose logical replication for two reasons (throughput and operational simplicity); if a later decision only overturns the throughput claim, it should `amend` D-118's throughput rationale specifically, not blanket-supersede a decision that's still correct on its second ground.

**Staleness as a soft, non-blocking signal.** A decision's evidence can decay in place — an `E-NNN` flips from `fresh` to `stale` as its half-life elapses, or a `review_trigger` condition fires (a major version release, a stated date). This should surface in the generated index (§6.6) as a "needs re-verification" flag, visible but not blocking — a soft nudge, not a gate. §3's alert-fatigue evidence is the direct reason for that design choice: a system that blocks or interrupts on every stale flag will train people (and AI agents) to dismiss it exactly the way 49–96% of clinical alerts get dismissed. A system that surfaces "these 14 decisions haven't been re-checked against evidence older than their own half-life" in a weekly or per-project index is far more likely to get acted on, precisely because it isn't in anyone's way.

### 6.6 Searchability across projects

Store both registries as plain-text Markdown with YAML frontmatter — not a database — for the reason given in §6.2: the same files need to be cheaply parsed by a future AI session with no memory of this one, and legible to a human without tooling. This is also simply where the field has converged (§2.4).

```
/evidence/E-047-postgres-logical-replication-benchmark.md
/evidence/E-048-...
/decisions/D-118-order-service-replication-strategy.md
/decisions/D-119-...
/index/                      # generated, not hand-maintained
  by-tag.md
  by-status.md
  needs-review.md            # auto-populated from stale/triggered records
  graph.json                 # machine-readable node/edge graph for programmatic queries
```

The `/index` directory is rebuilt by a script the AI pipeline runs itself each session, not curated by a person — this is the direct fix for the adoption failure §2.4 and §3 both surface repeatedly: **systems that require a human to remember to separately maintain a wiki decay within about two years, in every domain surveyed here.** A generated index costs nothing to keep current because nothing about it depends on anyone remembering to update it.

For **cross-project reuse**, the cleanest structure is a shared, growing `/evidence` corpus referenced by many projects' local `/decisions` registries — a org-wide evidence library, project-specific decision logs. A new project asking "how should we handle Postgres replication" first searches the shared evidence tag index before commissioning new research, the same way a lawyer Shepardizes an existing case before building a new argument from scratch, and a new benchmark only needs to be graded once no matter how many projects eventually cite it.

### 6.7 Adoption safeguards — the non-negotiables

Everything in this subsection is a direct, specific response to §3's finding that structure only changes behavior under particular conditions, and drifts into Decision-Documentation-Theater-style abandonment or alert-fatigue-style dismissal otherwise.

1. **Grade inline, at the point of citation — never as a separate audit pass.** A separate grading step is exactly the kind of bolted-on task that gets skipped once deadlines bite; it has to be part of the act of citing, not a follow-up chore.
2. **Hard cap: one base letter plus up to three modifier tags, ever.** No expanding into a full independent matrix by default. The optional full ACH-style evidence/hypothesis matrix (§2.2) is available but opt-in, reserved for flagged one-way-door decisions where the extra rigor is worth the extra time.
3. **Keep required decision-record fields under eight** (context, options, decision, rationale, evidence_refs, status, door_type, rejected alternatives). Everything else — confidence notes, review triggers, human-reviewed flags — is populated when relevant, not force-filled on every record.
4. **Rigor scales with `door_type`, never applied uniformly.** Most decisions in most projects are two-way doors; forcing one-way-door-level rigor onto all of them is the single fastest way to produce the "trivial and cosmic ADRs, never the load-bearing ones" pattern §2.4 documents.
5. **The registry is queried by a script the AI runs itself, not a dashboard a human has to remember to check.** This follows directly from the adoption evidence, not from a tooling preference.
6. **Flags, not gates.** A Grade E or uncorroborated citation is allowed to exist — visibly flagged, forcing acknowledgment — rather than silently passed through *or* hard-blocked. GRADE's own guidance does exactly this: you can act on low-certainty evidence, in specific stated circumstances, as long as you say so. A hard block just pushes people to route around the system entirely, which is worse than an honestly-labeled weak citation.

---

## 7. Migration path from the current system

No renumbering, no re-grading sweep, no big-bang cutover — consistent with the operational lesson underlying this whole report.

1. **Phase 1 (immediate, ~zero cost):** Keep existing A–E letters and D-NNN IDs exactly as they are. Start adding `verification-method` and `door_type` to *new* records only. This alone captures most of §3's hallucination-risk mitigation.
2. **Phase 2:** Add the three modifiers (`corroboration`, `recency`, `directness`) to new evidence citations. Backfill old records opportunistically, not as a dedicated project — mark anything not backfilled as `verification-method: unknown-legacy` rather than spending effort re-auditing history that mostly still holds.
3. **Phase 3:** Build the generated `/index` (by-tag, by-status, needs-review, graph.json) as a script. This is what makes staleness tracking and cross-project search real instead of aspirational.
4. **Phase 4 (opt-in, as-needed):** Introduce the full ACH-style alternatives matrix as an optional appendix block for decisions explicitly flagged high-stakes/one-way-door — never as a default requirement.

Each phase is independently useful and none depends on completing the previous one across the whole existing registry, which matters given §3 and §4's shared lesson: **systems that require a complete migration before they pay off tend not to survive to complete it.**

---

## Appendix A: Copy-paste templates

**Evidence record**

```yaml
---
id: E-XXX
title: ""
grade: A|B|C|D|E
modifiers:
  corroboration: single|corroborated|contested
  recency: fresh|aging|stale
  directness: direct|indirect
verification:
  method: fetched|cached|recalled|secondhand|human-provided
  fetched_date: YYYY-MM-DD
  fetched_by: ""
source:
  url: ""
  type: ""
tags: []
used_by: []
---

## Claim
[paraphrased in your own words — do not paste verbatim source text]

## Corroborating source(s)
[if corroboration: corroborated]

## Notes / caveats
```

**Decision record**

```yaml
---
id: D-XXX
title: ""
status: proposed|accepted|rejected|deprecated|superseded
door_type: one-way|two-way
date: YYYY-MM-DD
confidence: high|medium|low
evidence_refs: []
supersedes: null
superseded_by: null
amends: null
review_trigger: ""
tags: []
authored_by: ""
human_reviewed: true|false
---

## Context

## Options considered

## Decision

## Rationale

## Rejected alternatives

## Confidence note
```

---

## Appendix B: Key sources consulted

**Evidence-based medicine / GRADE**
- GRADE approach — https://en.wikipedia.org/wiki/GRADE_approach
- Evaluating the Certainty of Evidence in EBM (ScienceDirect) — https://www.sciencedirect.com/science/article/pii/S2405456923002316
- Murad et al., "Proposed new evidence-based medicine pyramid" (U.S. GRADE Network) — https://usblog.gradeworkinggroup.org/2016/06/proposed-new-evidence-based-medicine.html
- "Strong recommendations from low certainty evidence" — https://www.ncbi.nlm.nih.gov/pmc/articles/PMC10039768/
- "Beyond the pyramid: reconsidering evidence synthesis" — https://pmc.ncbi.nlm.nih.gov/articles/PMC13054162/
- Certainty-of-evidence assessment prevalence survey — https://www.ncbi.nlm.nih.gov/pmc/articles/PMC12442678/
- WHO surgical safety checklist trial review — https://pmc.ncbi.nlm.nih.gov/articles/PMC3279961/

**Intelligence analysis**
- Admiralty code — https://en.wikipedia.org/wiki/Admiralty_code
- Critical review of the Admiralty Code (diagonal-collapse finding) — https://www.blockint.nl/intel-analysis/critical-review-of-the-admiralty-code/
- ICD 203 analytic standards — https://legalclarity.org/icd-203-analytic-standards-for-all-source-intelligence/
- Words of Estimative Probability (Kent) — https://www.cisecurity.org/ms-isac/services/words-of-estimative-probability-analytic-confidences-and-structured-analytic-techniques
- ACH steps — https://isc.sans.edu/diary/22460
- Dhami, randomized ACH fidelity study — https://onlinelibrary.wiley.com/doi/full/10.1002/acp.3550

**Legal reasoning**
- Standards of proof overview — https://www.nolo.com/legal-encyclopedia/legal-standards-proof.html
- "Legal Standards by the Numbers" (Duke Judicature) — https://judicature.duke.edu/articles/legal-standards-by-the-numbers/
- Shepard's Citations — https://en.wikipedia.org/wiki/Shepard's_Citations
- Citator signal codes — https://fiveable.me/advanced-legal-research/unit-2/shepardizing-updating-case-law/study-guide/OXfTJS5QBfcvkd8z
- Citator disagreement study — https://www.aallnet.org/wp-content/uploads/2018/12/LLJ_110n4_02_hellyer.pdf

**Software engineering**
- Kitchenham, Dybå, Jørgensen — Evidence-Based Software Engineering for Practitioners — https://web-backend.simula.no/sites/default/files/publications/Dyba.2005.1.pdf
- "Decision Documentation Theater" — https://hidekazu-konishi.com/entry/architecture_decision_records_templates_and_operations.html
- Architectural Decision Records hub — https://adr.github.io/
- MADR — https://adr.github.io/madr/
- MADR / Y-Statement primer — https://ozimmer.ch/practices/2022/11/22/MADRTemplatePrimer.html
- Log4brains — https://github.com/thomvaill/log4brains/
- ADR tooling comparison — https://adr.github.io/adr-tooling/

**Decision science / AI-specific**
- One-way vs. two-way door decisions — https://www.producttalk.org/glossary-discovery-one-way-door-decision/
- Software engineering example — https://medium.com/one-to-n/one-way-two-way-door-decisions-a0e29029e200
- CSET, AI Safety and Automation Bias — https://cset.georgetown.edu/publication/ai-safety-and-automation-bias
- International AI Safety Report 2026 — https://arxiv.org/pdf/2602.21012
- Automation bias in LLM-assisted diagnostic reasoning — https://www.medrxiv.org/content/10.1101/2025.08.23.25334280.full.pdf
