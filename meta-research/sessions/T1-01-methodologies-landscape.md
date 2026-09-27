# The Pre-Development Research Landscape
### A Comprehensive Catalog of Structured Decision-Making Methodologies in Software Engineering and Adjacent Fields

*Prepared as a landscape survey to inform the design of an AI-assisted pre-development research meta-framework. Compiled August 2026.*

---

## How to Use This Document

This catalog surveys the methods that exist today for doing rigorous thinking *before* code gets written: how to frame a technical question, gather evidence, weigh alternatives, surface disagreement, and leave a record that survives the people who made the decision. It draws on four kinds of sources: (1) software engineering's own documented practices, (2) publicly described processes at specific engineering organizations, (3) formal decision-making disciplines from medicine, intelligence analysis, military planning, and law, and (4) the AI-native research and specification practices that emerged through 2025–2026.

It is organized in ten parts:

1. A taxonomy for comparing any method on the same terms
2. Software-native pre-development methodologies
3. How named engineering organizations structure this work
4. What medicine, intelligence analysis, military planning, and law do
5. Decision-science frameworks that bridge all of the above
6. Emerging AI-assisted methodologies (2025–2026)
7. A comparative matrix across ~25 methods
8. What the evidence actually shows about outcomes (not just popularity)
9. Gaps and opportunities — especially for solo developers and lean teams
10. Design implications for a meta-framework

Throughout, methods are marked with an evidence-strength indicator (🟢 strong controlled/quasi-experimental evidence, 🟡 case-study or practitioner-consensus evidence, ⚪ theoretical/emerging, evidence still thin) so that popularity is never mistaken for proof of effectiveness.

---

## Part 1 — A Taxonomy for Comparing Any Method

Every method below can be located on the same six axes. This is the scaffolding the rest of the document hangs on.

| Axis | What it captures |
|---|---|
| **Weight** | How much process overhead the method imposes — light (minutes/solo), medium (hours–days/team), heavy (days–weeks/dedicated facilitation) |
| **Reversibility fit** | Whether the method is suited to Amazon's "one-way door" (consequential, hard-to-reverse) decisions, "two-way door" (cheap-to-reverse) decisions, or both |
| **Team-size fit** | Solo developer, small team, or institution/enterprise |
| **Core function** | Recording a rationale, generating options, testing a hypothesis, surfacing disagreement, building consensus, or calibrating confidence |
| **Source domain** | Where the method originated — software, business, medicine, intelligence, military, law, or cognitive science |
| **Evidence base** | Whether effectiveness has been measured, or is simply widely practiced |

The single most useful cross-cutting idea in this entire landscape is **calibrating research effort to the cost of being wrong** — a principle every mature domain reinvents under a different name: Amazon's one-way/two-way doors, medicine's GRADE strength-of-recommendation, law's tiered evidentiary standards, and the military's distinction between deliberate planning and rapid OODA cycling. Section 9 argues this is the single most transferable — and most under-implemented — idea for a software pre-development framework.

---

## Part 2 — Software-Native Pre-Development Methodologies

### 2.1 Overview table

| Method | Origin | Weight | Core Artifact | Primary Function |
|---|---|---|---|---|
| Architecture Decision Record (ADR) | Michael Nygard, 2011 | Light | Short markdown file per decision | Record rationale for posterity |
| RFC (Request for Comments) | Steve Crocker/ARPANET, 1969; adapted by industry | Medium | Proposal doc + threaded comments | Surface disagreement before commitment |
| Design Doc | Long-standing industry practice, formalized publicly by Google | Medium–Heavy | Structured narrative doc | Force specificity, build consensus |
| Technical/Design Spike | Kent Beck, Extreme Programming, late 1990s | Light | Throwaway code + answered question | Reduce technical uncertainty, timeboxed |
| Proof of Concept / Prototype / MVP | General engineering practice | Light–Medium | Working (partial) artifact | Test feasibility or desirability |
| Architecture Tradeoff Analysis Method (ATAM) | SEI, Carnegie Mellon, late 1990s | Heavy | Utility tree + risk/tradeoff/sensitivity list | Evaluate an existing architecture against quality attributes |
| Quality Attribute Workshop (QAW) | SEI, Carnegie Mellon | Heavy | Prioritized quality-attribute scenarios | Elicit non-functional requirements *before* an architecture exists |
| Cost Benefit Analysis Method (CBAM) / ARID / ADD | SEI, Carnegie Mellon | Heavy | Cost/benefit-weighted architectural options | Extend ATAM with economics; validate partial designs; guide attribute-driven design |
| C4 Model | Simon Brown, ~2011 | Light | Context/Container/Component/Code diagrams | Shared vocabulary for describing architecture at multiple zoom levels |
| Wardley Mapping | Simon Wardley, Fotango, 2005 | Medium | Value-chain × evolution map | Situational awareness of what to build vs. buy vs. commoditize |
| Google/GV Design Sprint | Jake Knapp, Google, 2010; popularized via Google Ventures | Heavy | Tested prototype in 5 days | Rapid, cross-functional convergence on a product/design question |
| Impact Mapping / Event Storming / Domain Storytelling | Gojko Adzic; Alberto Brandolini; various, 2010s | Medium | Visual/collaborative maps | Connect business goals to technical scope; model a domain collaboratively |
| Fitness Functions / Evolutionary Architecture | Neal Ford, Rebecca Parsons, Patrick Kua (ThoughtWorks) | Medium | Automated architectural checks | Keep architecture decisions enforced continuously, not just recorded once |
| arc42 template | Peter Hruschka, Gernot Starke (German-speaking SE community) | Medium | 12-section architecture documentation template | Standardize what an architecture document should cover |
| DACI / RAPID | Atlassian (DACI); Bain & Company (RAPID) | Light | Named decision roles | Clarify *who* decides, not what evidence is needed |
| One-way door / two-way door triage | Jeff Bezos, Amazon, 1997/2015 shareholder letters | Light | A single up-front classification | Calibrate how much process a decision deserves before applying any other method |

### 2.2 Documentation-first methods

**Architecture Decision Records (ADR).** Nygard's original 2011 proposal was deliberately minimal: a numbered markdown file per decision with five fields — Title, Context, Decision, Status (proposed/accepted/deprecated/superseded), and Consequences — stored in version control next to the code it describes. The format's popularity rests almost entirely on its low friction: it requires no facilitator, no meeting, and no tooling beyond a text editor. ThoughtWorks moved ADRs into the "Adopt" ring of its Technology Radar roughly seven years after Nygard's post, and the community has since layered on variants — the Y-statement format (Zdun et al.), MADR (Markdown Any Decision Records), and AWS's own prescriptive-guidance template. ADRs record a decision *after* it has effectively been made; they are a memory device, not a deliberation device.

**RFCs.** The RFC format predates software-industry adoption by decades — Steve Crocker wrote the first one in 1969 to document a meeting of ARPANET researchers, deliberately choosing a tentative title ("Request for Comments") to invite correction rather than assert authority. Its modern software use (popularized in different lineages by Rust, Python's PEPs, and countless companies) borrows that same posture: a written proposal — problem, proposed solution, rejected alternatives, and open questions — circulated for asynchronous comment before a synchronous meeting is called only if disagreement remains unresolved. The throughline across every company that documents its RFC culture publicly (Shopify, GitLab, Stripe, Squarespace, Klarna, and dozens more) is the same two-phase pattern: async review first, meeting only as an escalation path.

**Design Docs.** Widely practiced long before it was written up, the format was popularized outside Google by a widely shared 2020 essay from a Google engineer, and separately documented in the book *Software Engineering at Google*. The essay's central claim is that design docs earn their keep less through the document itself and more through the *writing process* — forcing an author to make a design's trade-offs explicit before code exists, when redirection is cheap. Design docs sit one notch heavier than ADRs: they are pre-decision deliberation artifacts (typically requiring a team meeting to discuss), where ADRs are post-decision memory artifacts. Many organizations use both — a design doc to decide, an ADR to remember.

### 2.3 Exploration and de-risking methods

**Technical/Design Spikes.** Coined by Kent Beck during Extreme Programming's formative Chrysler C3 project in the late 1990s, a spike is a strict, timeboxed, throwaway exploration meant to answer exactly one question — "driving a spike through" a specific unknown rather than building toward a shippable feature. The Scaled Agile Framework later formalized this as an "enabler story." The discipline that separates a spike from ordinary exploratory coding is the pre-commitment to throw the code away and the narrow scope of the question being answered.

**Proof of Concept, Prototype, and MVP** are frequently conflated but answer different questions: a PoC tests *can this be built at all* (technical feasibility, often thrown away), a prototype tests *what should this look/feel like* (often for stakeholders or usability testing, not production-bound), and an MVP tests *will anyone want this* (a real, shippable slice aimed at market/user validation). Treating them interchangeably is a common source of wasted effort — building MVP-grade polish to answer a PoC-grade feasibility question, or vice versa.

**Design Sprints.** Jake Knapp began running five-day, Monday-through-Friday sprints inside Google (on Search, Chrome, and Google X) starting in 2010, then refined the format with Braden Kowitz and John Zeratsky after moving to Google Ventures in 2012; the 2016 book *Sprint* made the format broadly popular (GV reports adoption at over 100 companies, including Slack, Uber, and Airbnb). The structure — map the problem, sketch solutions individually, decide, prototype, test with real users — is one of the few methods on this list that mandates real user contact *before* build begins, which distinguishes it from the otherwise similar internal-facing methods above it.

### 2.4 Heavyweight architecture evaluation: the SEI family

Carnegie Mellon's Software Engineering Institute produced a coordinated family of architecture methods, most from the late 1990s, aimed at large, often defense- or safety-critical systems:

- **ATAM (Architecture Tradeoff Analysis Method)** — a facilitated, multi-day workshop that stress-tests a *proposed* architecture against prioritized quality-attribute scenarios (performance, security, modifiability, etc.), explicitly to surface trade-offs and risk, not to produce a "correct" score.
- **QAW (Quality Attribute Workshop)** — designed to run *before* ATAM, when no architecture exists yet, to elicit and prioritize the quality-attribute scenarios ATAM will later test against.
- **CBAM (Cost Benefit Analysis Method)** — extends ATAM with explicit cost/benefit weighting of architectural options.
- **ARID (Active Reviews for Intermediate Designs)** and **ADD (Attribute-Driven Design)** — round out the family for partial designs and attribute-driven design generation, respectively.

This family is thorough and well-documented, but its cost (multi-day facilitated workshops, a trained external facilitator, a room full of stakeholders) makes it essentially inaccessible below the scale of a well-resourced enterprise team — a gap discussed further in Part 9.

### 2.5 Strategy, mapping, and reversibility framing

**C4 Model** (Simon Brown) gives teams a shared, zoom-level vocabulary — System Context, Container, Component, Code — for architecture diagrams, addressing the common failure mode where "the architecture diagram" means something different to every person in the room.

**Wardley Mapping** plots the components a business needs (a value chain, anchored at user need) against an evolution axis — genesis, custom-built, product/rental, commodity/utility — that describes how predictably a component's supply is changing. Its main pre-development use is deciding what's worth building in-house (genesis/custom) versus what should simply be bought as a commodity, a question most architecture methods assume away.

**Fitness Functions / Evolutionary Architecture** (associated with ThoughtWorks) reframes architecture governance as continuous and automated rather than a one-time review, encoding architectural constraints (e.g., "service X may never call database Y directly") as executable checks in CI. This matters for pre-development research because it changes what a "decision" needs to produce: not just a document, but something a machine can verify was actually followed — a theme that resurfaces sharply in Part 6's discussion of AI coding agents.

**DACI (Driver, Approver, Contributors, Informed) and RAPID (Recommend, Agree, Perform, Input, Decide)** are governance frameworks, not research frameworks — they clarify *who* has authority over a decision, which most software methods leave implicit and which becomes a real bottleneck once an RFC or design doc process scales past a single team.

**One-way door / two-way door triage.** Amazon's own contribution to this space is not a research method at all but a *triage* heuristic, introduced in Bezos's 1997 and 2015 shareholder letters: classify a decision as a "one-way door" (consequential, expensive or impossible to reverse — warrants slow, deliberate, highly-consulted analysis) or a "two-way door" (cheap to reverse — warrants fast action on roughly 70% of the information you'd ideally want, made by small groups or individuals). Because it's applied *before* choosing a research method, it functions as the dial that should set how heavy a process from the rest of this catalog gets applied — the taxonomy's reversibility axis exists because of this idea.

---

## Part 3 — How Leading Engineering Organizations Structure This Work

### 3.1 Comparison table

| Organization | Primary pre-dev artifact | Process weight | Review gate | Distinguishing trait |
|---|---|---|---|---|
| **Google** | Design doc | Medium–heavy | Team design review meeting; senior-engineer scrutiny | Docs are the primary knowledge-sharing surface (`go/` links); design review feedback is reported to differ substantially from code review feedback |
| **Amazon** | PR/FAQ (Working Backwards) + six-page narrative memo | Heavy for one-way doors, deliberately light for two-way doors | "Narrative meeting" — 20 minutes silent reading, then discussion | Explicitly customer-first framing (fictional press release written *before* the FAQ); door-type triage sets process weight up front |
| **Stripe** | RFC / long-form internal doc | Medium–heavy | API changes specifically pass a strict, code-review-plus review process | An unusually strong, explicitly cultivated "writing culture" — leadership itself writes and models long-form docs; historically no dedicated PM role, pushing product thinking onto engineers |
| **Shopify** | RFC + decision log (for large cross-cutting programs) | Medium, scaling to heavy for multi-team "Programs" | Escalation-based: disagreements become logged decisions once stakeholders align | Publishes an explicit "engineering program" playbook with defined artifacts for alignment, status, and decisions |
| **Vercel (Next.js)** | RFC via public GitHub Discussions | Medium, but concentrated | Community comment period; final call stays with the core team | Governance is nominally open ("large architecture decisions start as an RFC"), but independent observers note that in practice very few community-originated RFCs have been adopted relative to internally-driven ones — a useful caution against assuming a documented process is a fully lived one |
| **Netflix** | None mandated — minimal formal artifact | Deliberately light | Peer influence and "context, not control" culture substitute for formal gates | The clearest counter-example in this catalog: a "Freedom & Responsibility" culture that explicitly avoids heavyweight process, relying instead on hiring judgment and informal norms — illustrates that heavy documentation is a choice, not a requirement, even at extreme scale |
| **Spotify, Klarna, GitLab, and others** | RFC and/or ADR, often both together | Medium | Varies | Several organizations reporting on their own practice (per industry surveys) explicitly pair a *design doc/RFC* for the deliberation with a separate, lighter-weight *ADR* purely for the historical record — treating them as complementary rather than redundant |

### 3.2 What this comparison shows

No two companies solve this identically, but three patterns recur:

1. **Async-first, sync-as-escalation.** Every documented process above — regardless of company — puts a written proposal in front of stakeholders before calling a meeting, and reserves the meeting for genuine disagreement rather than status-reporting. This is close to a universal convention among engineering organizations that write publicly about their process.
2. **Deliberation and memory are treated as separate artifacts by the more mature practices.** Google/Amazon/Stripe-style design docs and PR/FAQs do the deliberating; a shorter, more mechanical record (ADR or decision log) does the remembering. Conflating the two — trying to make one document do both jobs — is a common failure mode reported anecdotally across many of these write-ups.
3. **Stated process and lived process diverge**, sometimes significantly (Next.js's RFC governance is the clearest publicly documented example). A catalog of *documented* processes will always overstate how universally and faithfully those processes are followed — worth remembering when using this section as a benchmark.

---

## Part 4 — What Other Fields Do

### 4.1 Medicine: grading evidence, not just gathering it

Medicine's contribution is not a single method but an entire **discipline of grading how much to trust what you know**, developed because clinical decisions are irreversible in a way most software decisions are not.

- **The evidence hierarchy** ranks study types by susceptibility to bias — systematic reviews/meta-analyses and randomized controlled trials sit above cohort studies, which sit above case-control studies, case series, and expert opinion. The point is not that lower tiers are worthless, but that a recommendation should say *which* tier its evidence came from.
- **PICO** (Population, Intervention, Comparison, Outcome) is medicine's answer to the "garbage in, garbage out" problem: before searching for evidence, force the question itself into a structure specific enough to be answerable and searchable.
- **GRADE (Grading of Recommendations Assessment, Development and Evaluation)**, developed by an international working group starting around 2000, separates two judgments that are routinely conflated elsewhere: *how certain is the evidence* (rated high/moderate/low/very-low, based on risk of bias, consistency, and precision) and *how strong is the resulting recommendation* (strong vs. conditional/weak). GRADE is now used or endorsed by over 100 organizations worldwide, including WHO and the Cochrane Collaboration, and has been extended into "Evidence to Decision" (EtD) frameworks that structure how a panel moves from graded evidence to an actual policy or clinical decision.
- **Evidence-Based Software Engineering (EBSE).** This exact cross-domain transplant has already been attempted: researchers (notably Kitchenham, Dybå, and Jørgensen) explicitly ported medicine's evidence-based paradigm into software engineering, proposing the same five-step cycle — convert a problem into an answerable question, search for the best evidence, critically appraise it, integrate it with practical experience to decide, then evaluate the outcome — with the Systematic Literature Review (SLR) as its core tool. EBSE is well established as an *academic research methodology* (used to synthesize what the empirical software engineering literature says about a practice), but it has seen very limited adoption as a *practitioner* decision tool for day-to-day architecture choices — a gap discussed in Part 9.

### 4.2 Intelligence analysis: structured techniques for reasoning under deception and incomplete information

- **Analysis of Competing Hypotheses (ACH)**, developed by CIA veteran Richards Heuer between 1978–1986 and formalized in his 1999 CIA-published monograph *Psychology of Intelligence Analysis*, inverts the natural instinct to look for evidence that confirms a favored explanation. Instead, analysts lay out *every* plausible hypothesis in a matrix against *all* available evidence, and — critically — try to disconfirm each hypothesis rather than confirm it; the hypothesis with the least evidence against it, not the most evidence for it, wins. This single reframe (falsification over confirmation) is the most direct, underused transplant candidate for architecture decisions, which are almost always written up as "why we chose X" rather than "why X survived every attempt to disprove it."
- **The Delphi Method**, developed at RAND by Olaf Helmer and Norman Dalkey in the 1950s for Cold War military forecasting, structures group judgment through anonymous, iterative rounds: each expert answers privately, sees an anonymized summary of the group's reasoning, and revises — repeating until responses converge (typically measured by shrinking interquartile range) or clearly won't. Anonymity specifically defeats status-based anchoring (a senior voice dominating the room), which most software design-review meetings do nothing to prevent.
- **The broader family of Structured Analytic Techniques** (of which ACH and Delphi are members) also includes the **Key Assumptions Check** (list every assumption a conclusion depends on and test each one explicitly) and formalized **Team A/Team B** exercises — both directly portable to architecture decisions with almost no modification.

### 4.3 Military planning: tempo, and adversarial stress-testing

- **The OODA Loop** (Observe–Orient–Decide–Act), developed by USAF Colonel John Boyd from his experience as a fighter pilot in the early 1970s, is less a checklist than a claim about competitive advantage: the side that can cycle through observation and re-orientation *faster* than its opponent can out-maneuver a materially stronger but slower one. Its software-relevant descendant is the entire case for two-way-door decisions and reversible experimentation: speed of the *loop*, not correctness of any single pass, is the lever.
- **Red teaming** has a genuinely old lineage — the Roman Catholic Church's *advocatus diaboli* ("Devil's Advocate"), formally tasked since roughly the 16th–17th century with arguing against a candidate's canonization; 19th-century Prussian *Kriegsspiel* wargaming, where a designated opposing force actively tried to defeat a plan rather than merely critique it; and the CIA's post-9/11 "Red Cell." The important distinction, drawn out clearly in current red-teaming literature, is that **devil's advocacy is internal and contrarian** (someone in the room argues the opposite case), while **red teaming is external and structured** (a team reasons *as* a specific adversary, from that adversary's actual incentives and constraints, not just from generic skepticism) — a materially harder and more valuable exercise. The Israeli intelligence community's "Tenth Man Doctrine" — if nine analysts agree, the tenth is obligated to build the case that they're wrong — operationalizes the same idea as a standing rule rather than an occasional exercise.
- **Premortems**, formalized by cognitive psychologist Gary Klein in a widely cited 2007 *Harvard Business Review* article, are the clearest evidence-backed technique in this entire catalog (see Part 8) and sit at the intersection of military after-action thinking and behavioral decision research: imagine the project has already failed and work backward to explain why, exploiting "prospective hindsight" rather than uncertain foresight.

### 4.4 Legal reasoning: tiering confidence to stakes, and arguing both sides on purpose

- **IRAC (Issue, Rule, Application, Conclusion)** is the standard scaffold taught in American legal education for structuring analysis: identify the precise question, state the governing rule, apply the rule to the specific facts, and only then conclude. One often-cited origin story traces an early version to the U.S. Army's WWII-era effort to teach fast, structured problem-solving to newly drafted recruits — a reminder that "structure the question before you argue the answer" is a genuinely cross-domain habit, not a software-specific insight waiting to be discovered.
- **Tiered evidentiary standards** are law's most directly transplantable idea: *preponderance of the evidence* (more likely than not — a bar just over 50%, used for ordinary civil disputes), *clear and convincing evidence* (substantially more likely true than not — a "firm belief or conviction," used for higher-stakes civil matters like fraud or parental rights), and *beyond a reasonable doubt* (used only where the state can deprive someone of liberty). The structural insight — that the *bar for how convinced you need to be* should rise explicitly with the stakes and irreversibility of the decision, and that this bar should be named and agreed on in advance rather than argued about after the fact — has no equivalent in mainstream software decision-making, where an ADR for a caching layer and an ADR for a data-residency architecture are typically expected to clear the same informal bar.
- **The adversarial process itself** — assigning one party to argue for a position and another to argue against it, in front of a neutral decision-maker — is essentially institutionalized red-teaming, and is one of humanity's oldest formalized debiasing mechanisms.

---

## Part 5 — Decision-Science Bridge Frameworks

These sit between the domain-specific methods above and general-purpose applicability to almost any pre-development question.

- **Cynefin**, developed by Dave Snowden starting in 1999 and popularized through a 2007 *Harvard Business Review* article with Mary Boone, sorts a situation into one of several domains — clear/obvious, complicated, complex, chaotic, and "disorder" (not knowing which domain applies) — based on how perceptible the relationship between cause and effect is. Its central, software-relevant claim is that *the correct method depends on which domain you're in*: clear and complicated problems reward analysis and best/good practice (the SEI/ATAM family fits here); complex problems don't have a knowable "right" architecture in advance and instead reward safe-to-fail probing (spikes, prototypes, two-way-door experimentation fit here); chaotic problems reward acting first and analyzing after (an incident response, not an architecture review, fits here). A recurring failure mode Cynefin names directly: applying complicated-domain rigor (heavyweight review boards) to a genuinely complex problem, or applying complex-domain experimentation to a problem that actually has a knowable right answer.
- **Superforecasting.** Philip Tetlock and Barbara Mellers's Good Judgment Project, the winning entry in IARPA's 2011–2015 forecasting tournament, is one of the largest controlled studies of judgment quality ever run: hundreds of thousands of forecasts, tracked against real-world outcomes using Brier scores. Its most transferable, and most counter-intuitive, finding is that a specific pool of practices — the top ~2% of forecasters, dubbed "superforecasters" — beat career intelligence analysts working with classified information by a wide margin (commonly cited at roughly 30%), and that a short training intervention measurably improved ordinary participants' accuracy. Notably, "**comparison classes**" (reference-class forecasting — grounding an estimate in how similar past situations actually turned out, the "outside view") was one of the training elements most strongly *correlated with accuracy*, while, in a genuinely useful caveat, "post-mortem analysis" training was mildly *correlated with inaccuracy* in that same dataset — a reminder that not every plausible-sounding debiasing technique earns its keep equally.
- **Decision journaling**, most visibly associated with former professional poker player and decision researcher Annie Duke, targets a specific, well-documented bias: "**resulting**," the tendency to judge a decision's quality by its outcome rather than by the process and information available at the time. The practice — record your reasoning, an explicit confidence percentage, and a scheduled future review date, all *before* the outcome is known — is one of the lightest-weight methods in this catalog (minutes per decision) and directly targets the same calibration goal as GRADE and superforecasting, without requiring any institutional machinery at all.

---

## Part 6 — Emerging AI-Assisted Methodologies (2025–2026)

This is the fastest-moving section of the landscape, and the one most directly relevant to the meta-framework this catalog is meant to inform.

### 6.1 Spec-Driven Development (SDD)

SDD emerged through 2025 as organized pushback against "vibe coding" — a term popularized by Andrej Karpathy in early 2025 for loosely prompting an AI coding agent and accepting whatever it returns — after teams found that fast code generation was outpacing teams' ability to verify the generated code actually solved the intended problem. The core move is to make a precise, version-controlled specification (not the code) the source of truth: an agent derives an implementation plan and task breakdown from the spec, and when requirements change, the spec is edited first and code is regenerated from it. GitHub's Spec Kit (which popularized a `specify → plan → implement` workflow in late 2025), AWS's Kiro, and comparable capabilities across Claude Code, Cursor, OpenSpec, BMAD-METHOD, and others had all shipped some version of this by mid-2026. A companion syntax, **EARS (Easy Approach to Requirements Syntax)**, has emerged specifically for writing acceptance criteria precise enough for an agent to verify against.

This is directly a pre-development research practice by another name — the spec-writing phase *is* the research phase, now made mandatory because an under-specified prompt produces plausible-looking but wrong code at a speed that makes the old "we'll figure it out as we build" fallback much more expensive than it used to be. It is worth noting, however, that ThoughtWorks' Technology Radar placed SDD in its "Assess" ring (not yet "Adopt") in 2025 and explicitly flagged the risk of the practice reintroducing heavyweight, big-upfront-specification anti-patterns that agile methods spent two decades trying to eliminate — a caution worth taking seriously rather than treating SDD's rapid tooling adoption as proof of its effectiveness.

### 6.2 Architecture documentation is bifurcating into human-readable and agent-executable layers

A distinct, closely related shift: teams are increasingly maintaining architecture guidance in two parallel forms — an `AGENTS.md`-style file of constraints an AI coding agent is expected to *obey* on every change, versus a traditional ADR that *explains* a decision to a human reader. Several practitioner write-ups in early-to-mid 2026 describe extending ADR templates with an explicit "Agent Context" block — recording which model proposed or influenced a decision, what alternatives it considered, its stated confidence, and the human reviewer who accepted or modified it — specifically for audit-trail purposes as agents take a larger role in day-to-day architectural choices. A related and increasingly common pattern is **continuous ADR generation**: an agent is instructed to draft a proposed ADR automatically whenever a pull request touches architecturally significant code (service boundaries, data models, authentication, infrastructure), with the draft going to a human architect for review rather than being merged automatically. An arXiv paper circulating in this period, *"Architecture Without Architects: How AI Coding Agents Shape Software Architecture,"* frames this shift directly: the artifact that most shapes what actually gets built is increasingly the one written for the agent, not the one written for the historical record — which raises a real risk that the *human-legible* rationale (the traditional purpose of an ADR) quietly degrades even as the *machine-enforceable* rules improve.

### 6.3 Multi-agent deliberation and the "generate–critique–rank–evolve" pattern

Structured disagreement between multiple AI agents — rather than a single model asked once — has a real academic lineage (Irving et al.'s 2018 "AI safety via debate" proposal; Du et al.'s 2023 finding that multi-agent debate measurably reduces hallucination and improves factual accuracy) and by 2025–2026 had matured into a specific, transferable template most clearly demonstrated by Google DeepMind's **Co-Scientist** system (published in *Nature*, built on Gemini): specialized agents *generate* candidate hypotheses grounded in literature, a *reflection*/*critique* agent stress-tests them, a *ranking* process runs a tournament between competing hypotheses, and the surviving hypotheses are *evolved* into refined versions for the next round — explicitly mirroring the human scientific method rather than replacing it, with a scientist remaining in the loop throughout. Related systems (MIT's SciAgents, FutureHouse's Robin, and Sakana AI's "AI Scientist" line) apply the same generate–critique–rank–evolve shape to different scientific domains.

This maps unusually well onto pre-development architecture research: "which datastore should we use," "should this be a monolith or split into services," and "what's the right hypothesis for why our current design won't scale" are all naturally framed as competing-hypotheses questions rather than single-shot Q&A — closer to ACH (Part 4.2) than to a simple prompt-and-answer pattern. Current evidence on *when* multi-agent structures actually help is more nuanced than the hype suggests, however: 2026 research comparing multi-agent to well-engineered single-agent baselines finds the advantage concentrated in **parallelizable, read-heavy tasks** (fanning out several independent research sub-queries and merging results), with little to no advantage — and real added cost and latency — on genuinely **sequential reasoning** tasks, where a single agent given the same budget often performs just as well. A separate, notable observation from Kent Beck (creator of Extreme Programming and the technical spike, still active in this space) is that he sees AI coding tools as accelerating a return to XP's original small-team, tightly-coupled, customer-proximate practices — a direct, practitioner-level echo of this catalog's broader observation that the AI era rewards *lightweight, fast-cycling* methods over heavyweight, once-a-quarter review processes.

---

## Part 7 — Comparative Matrix

Weight, time cost, and evidence strength are necessarily approximate; they describe typical/reported practice, not a universal constant.

| Method | Source Domain | Weight | Typical Time Cost | Team-Size Fit | Reversibility Fit | Core Function | Evidence Strength |
|---|---|---|---|---|---|---|---|
| ADR | Software | Light | Minutes–hours | Solo → Enterprise | Both | Record rationale | 🟡 Strong practitioner consensus; little controlled outcome data |
| RFC | Software | Medium | Days | Team → Enterprise | Both | Surface disagreement pre-commit | 🟡 Strong practitioner consensus |
| Design Doc | Software | Medium–Heavy | Days–weeks | Team → Enterprise | Leans one-way | Force specificity, build consensus | 🟡 Practitioner consensus; some internal-company claims, not independently verified |
| Technical Spike | Software (XP) | Light | Hours–days, timeboxed | Solo → Team | Two-way | Reduce technical uncertainty | 🟡 Practitioner consensus |
| PoC / Prototype / MVP | Software | Light–Medium | Days | Solo → Team | Two-way | Test feasibility/desirability | 🟡 Practitioner consensus |
| ATAM | SEI / Software | Heavy | Days (facilitated workshop) | Enterprise | One-way | Evaluate architecture vs. quality attributes | 🟡 DoD/enterprise case studies; no RCTs |
| QAW / CBAM / ARID / ADD | SEI / Software | Heavy | Days | Enterprise | One-way | Elicit NFRs; cost/benefit; validate partial designs | 🟡 Case-study evidence |
| C4 Model | Software | Light | Hours | Solo → Enterprise | Both | Shared architecture vocabulary | ⚪ Widely adopted; effectiveness largely anecdotal |
| Wardley Mapping | Strategy | Medium | Hours–days | Team → Enterprise | Both | Value-chain/evolution situational awareness | ⚪ Practitioner consensus; limited formal study |
| Design Sprint | Product/Design | Heavy | 5 days | Team | Two-way (validate before build) | Rapid convergence + user testing | 🟡 100+ company case studies; no controlled comparison |
| One-way/two-way door triage | Business | n/a (heuristic) | Minutes | Any | Is the axis itself | Calibrate process weight to stakes | 🟡 Strong internal Amazon testimony; not independently studied |
| EBM hierarchy / PICO | Medicine | Light (framing) | Minutes–hours | Any | Any | Frame answerable question; rank evidence types | 🟢 Decades of methodological research |
| GRADE | Medicine | Heavy | Weeks (panel process) | Institutional | One-way | Grade evidence certainty + recommendation strength | 🟢 RCT-backed; used/endorsed by 100+ orgs, WHO, Cochrane |
| Evidence-Based Software Engineering | Medicine → Software | Heavy | Weeks | Institutional/research | Any | Port EBM's 5-step cycle into SE via systematic review | 🟡 Well-established academic method; minimal day-to-day practitioner adoption |
| ACH | Intelligence | Medium | Hours | Solo → Team | Both | Falsify hypotheses against an evidence matrix | 🟡 Strong cognitive-science rationale; limited direct outcome studies |
| Delphi | Intelligence/Forecasting | Heavy | Weeks (multi-round) | Expert panel | Both | Converge expert judgment anonymously | 🟢 Strong in health-guideline contexts; mixed for general forecasting |
| Structured Analytic Techniques (Key Assumptions Check, Team A/B) | Intelligence | Medium | Hours | Team | Both | Test assumptions explicitly; institutionalize dissent | 🟡 Institutional track record; limited controlled comparison |
| OODA Loop | Military | Light (mental model) | Continuous | Any | Two-way (fast iteration) | Win via faster observe–act cycling | 🟡 Strong in origin domain; thin controlled evidence elsewhere |
| Red Teaming | Military/Intelligence | Heavy | Days; dedicated team | Team → Enterprise | One-way | Adversarial stress-test from outside perspective | 🟡 Strong institutional track record; limited controlled comparison |
| Premortem | Cognitive psychology | Light | 20–30 minutes | Any | Both | Surface failure modes via prospective hindsight | 🟢 Multiple controlled studies; ~30% improvement in failure-mode identification |
| IRAC | Legal | Light | n/a (mental template) | Any | Any | Structure issue → rule → application → conclusion | 🟡 Strong within legal pedagogy; untested outside law |
| Tiered evidentiary standards | Legal | n/a (calibration heuristic) | n/a | Any | Is the axis itself | Calibrate confidence bar to stakes | 🟡 Strong within legal system; novel transfer to software |
| Cynefin | Decision science | Light | Minutes (sensemaking) | Any | Any | Match method to problem's causal structure | 🟡 Practitioner consensus; limited RCTs |
| Superforecasting practices | Decision science | Medium | Ongoing | Solo or crowd | Any | Calibrated probabilistic forecasting | 🟢 Large government-funded tournament (IARPA), replicated |
| Decision Journal | Decision science | Light | Minutes per decision | Solo → Team | Any | Separate decision quality from outcome; enable calibration | 🟡 Strong theoretical grounding; thinner large-scale outcome data |
| Spec-Driven Development | Software/AI, 2025–26 | Medium | Hours–days | Solo → Enterprise | Both | Make the spec the executable source of truth for AI agents | ⚪ Emerging; vendor case studies; ThoughtWorks flags antipattern risk |
| Multi-agent debate / AI Co-Scientist pattern | AI research | Medium–Heavy (compute) | Minutes–hours | Solo → Enterprise | Both | Structured generate–critique–rank–evolve loop | ⚪ Emerging; strong on parallel-read/science tasks, weak-to-null on sequential reasoning |

---

## Part 8 — What the Evidence Actually Shows

The user's brief specifically asked for evidence of *outcomes*, not just popularity. Sorting the methods above by evidence strength surfaces an uncomfortable but important pattern:

**The most rigorously validated methods in this entire catalog are not software-native.** The strongest controlled evidence belongs to premortems (a ~30% improvement in identifying failure causes, first demonstrated in a 1989 Wharton/Cornell/Colorado study on "prospective hindsight" by Mitchell, Russo, and Pennington, popularized by Klein's 2007 article, and independently replicated in a 2010 178-participant study on pandemic-response planning), to superforecasting practices (validated across a multi-year, IARPA-funded tournament against professional intelligence analysts, with superforecasters outperforming analysts working with classified access by a wide margin), and to GRADE (built on and continuously validated against the broader evidence-based medicine literature, and now used by over 100 institutions worldwide). All three came from psychology, intelligence analysis, and medicine, respectively — none from software engineering.

**Software engineering's own pre-development practices — ADRs, RFCs, design docs — rest almost entirely on practitioner testimony, not controlled measurement.** This is not a knock on the practices; the "Design Docs at Google" essay's claim that writing forces specificity is intuitively compelling and consistent with everything decision science knows about externalizing reasoning. But it is worth being honest that no one has run a controlled study comparing teams that write ADRs against teams that don't and measured downstream architectural rework, incident rates, or decision reversal frequency. The evidence for these practices is closer in kind to case-study/testimonial evidence in medicine's hierarchy — real, useful, but several tiers below GRADE's standard.

**Evidence-Based Software Engineering (EBSE) shows the transplant from medicine is possible but has not taken hold as a practitioner tool.** EBSE has succeeded as an *academic research methodology* (systematic literature reviews are now a standard tool for synthesizing what the SE research literature says about a given practice) but has seen little uptake as something an individual engineer or team reaches for when facing a live architecture decision — largely because the systematic-review machinery is heavy, slow, and built for answering "what does 20 years of research say" rather than "what should we do by Friday."

**Not every plausible debiasing technique earns its keep equally**, and the superforecasting research is unusually direct about this: reference-class forecasting ("the outside view" — grounding an estimate in how similar past situations actually turned out) was strongly correlated with forecasting accuracy in Tetlock's data, while a "post-mortem analysis" training component was mildly *anti*-correlated with accuracy in the same study — a useful caution against assuming any technique that sounds rigorous actually improves outcomes without checking.

**Multi-agent AI deliberation shows genuine but narrow, task-dependent benefit as of 2026.** It measurably reduces hallucination and improves factual grounding on parallelizable, read-heavy research tasks (exactly the shape of "survey what's out there before deciding"), but current research finds no consistent advantage — and real added latency/cost — on genuinely sequential reasoning tasks when compared against a well-engineered single agent given the same compute budget. This is a meaningfully different, more measured picture than the framing in most vendor material about multi-agent systems.

**The net takeaway:** if the goal is to borrow what's *proven* to improve decision outcomes rather than what's merely popular, the highest-confidence imports are premortems, structured falsification (ACH-style), reference-class/outside-view forecasting, and reversibility-calibrated rigor (GRADE's strength-grading logic, Amazon's door typing, and law's tiered evidentiary standards) — not any specific software-native document format, which functions more as a *container* for good thinking than as a guarantee of it.

---

## Part 9 — Gaps and Opportunities

Ten gaps emerge from placing every method in this catalog side by side. They are ordered from most specific to most structural.

**1. No stakes-calibrated confidence grading in mainstream software practice.** GRADE forces a guideline author to separately state *how certain* the evidence is and *how strongly* to act on it. Legal standards force a fact-finder to name which confidence bar applies before weighing evidence. Nothing in the ADR/RFC/design-doc world does this: a decision to switch a caching library and a decision to change a data-residency architecture are typically expected to clear the same informal bar, in the same lightweight template, even though the cost of being wrong differs by orders of magnitude.

**2. No default falsification discipline.** Design docs and RFCs are structurally built to answer "why did we choose X" — they invite justification, not disconfirmation. ACH's core discipline — lay out every plausible alternative and actively try to kill each one, including the favorite — has no equivalent default step in any mainstream software template. Teams that do this well are relying on individual habit, not the format.

**3. No default adversarial step, especially for solo developers and small teams.** Red teaming and devil's advocacy work because someone is *structurally assigned* to argue against the plan. A solo developer or two-person team has no natural "tenth man" — and no framework in this catalog currently assigns that role to anything. This is arguably the single most direct opportunity for AI assistance: an agent explicitly instructed to argue the adversary's case, in the adversary's own terms (not merely list generic risks), is a close, practical stand-in for the CIA Red Cell or the Devil's Advocate role that solo developers and lean teams structurally lack today.

**4. Enterprise-weight processes do not scale down.** ATAM, QAW, CBAM, and Delphi all assume a facilitator, a room of stakeholders, and days of dedicated time. None of that exists for a solo developer or a two-person startup, yet the underlying *questions* those methods answer — which quality attributes actually matter here, what are we assuming, where do experts disagree — are exactly as relevant at small scale. The gap is not that small teams don't need this thinking; it's that no one has repackaged it at solo-developer cost.

**5. The evidence base for software-native methods is thin, and no one is closing that gap.** As Part 8 lays out, ADRs, RFCs, and design docs are supported by strong practitioner consensus and almost no controlled measurement, in sharp contrast to premortems, GRADE, or superforecasting. A meta-framework that simply repackages ADRs and RFCs is repackaging the same unvalidated assumptions software engineering has run on for over a decade.

**6. Decisions are recorded, not revisited.** ADRs have a "superseded" status field, but nothing in mainstream practice schedules a return date the way Annie Duke's decision journal does (confidence percentage + explicit review date, set *before* the outcome is known). Most ADRs are written once and never checked against what actually happened — which means the "resulting" bias Duke identifies is structurally invited rather than guarded against, and there is no feedback loop by which a team's calibration could even in principle improve over time.

**7. Architecture documentation is fragmenting into two unreconciled audiences.** As Part 6.2 describes, teams are now maintaining human-legible ADRs *and* agent-executable constraint files (AGENTS.md and similar) as increasingly separate artifacts, generated and consumed differently. No current practice treats these as a single decision record with two views — creating a real risk that the version an AI agent actually obeys quietly diverges from the version a human reads to understand why.

**8. The generate–critique–rank–evolve pattern proven in AI-assisted science has not been packaged for architecture questions.** Google's Co-Scientist and its relatives are a well-evidenced template for exactly the kind of competing-hypotheses reasoning pre-development architecture research needs ("should this be Postgres or DynamoDB," "monolith or services") — but as of this writing, that pattern lives in scientific-discovery tooling, not in anything purpose-built for "should we build this and how."

**9. Spec-driven development is optimizing implementation fidelity, not decision quality.** The 2025–2026 SDD wave (Spec Kit, Kiro, and peers) is very good at making sure an AI agent builds *what the spec says* — but a precisely executed spec for the wrong architecture is still the wrong architecture. ThoughtWorks' own caution about SDD reintroducing big-upfront-specification anti-patterns is a direct warning that the current AI-tooling wave is solving *execution* fidelity while leaving the harder *pre-development research* question — is this the right thing to build, on the right foundation — almost entirely unaddressed.

**10. No single throughline exists from question-framing to evidence-gathering to hypothesis-testing to decision to revisiting.** This is the structural gap underneath all the others. Right now, a rigorous team has to hand-assemble its own pipeline: PICO-style question framing (borrowed, if at all, from nowhere in software's own toolkit), an ACH-style hypothesis matrix (borrowed from intelligence analysis), a premortem (borrowed from cognitive psychology), a reversibility triage (borrowed from Amazon), a written record (native to software), and a scheduled revisit (borrowed from decision journaling) — with no existing framework connecting these into one coherent, right-sized flow. Every piece above has been invented, validated (to varying degrees), and documented *somewhere*. Nothing currently stitches them together end to end, calibrated by stakes, and sized for one person with an AI collaborator rather than a facilitator and a conference room.

---

## Part 10 — Design Implications for an AI-Native Meta-Framework

This catalog does not attempt to design the meta-framework itself, but four implications follow directly from the gap analysis above and are worth stating plainly:

- **Reversibility should be the first question the framework asks, not an afterthought.** Every heavyweight method in this catalog (ATAM, GRADE, Delphi, red teaming) is justified by high stakes and low reversibility; every lightweight one (spikes, ADRs, premortems) works precisely because it doesn't try to be those things. A framework that asks "one-way door or two-way door?" before selecting which subsequent steps to run is applying Part 1's taxonomy as an actual routing mechanism rather than a descriptive label.
- **An AI collaborator is well-suited to fill exactly the roles solo developers structurally lack** — the Devil's Advocate, the ACH hypothesis-killer, Delphi's anonymous second opinion, the Tenth Man — provided it is explicitly instructed to occupy those adversarial roles rather than defaulting to helpful agreement. The generate–critique–rank–evolve pattern from Co-Scientist-style systems is a validated template for structuring that, already proven on structurally similar hypothesis-competition problems.
- **Calibration needs a closing loop, not just an opening ritual.** Of everything surveyed here, only decision journaling and GRADE's evidence-certainty grading build in a mechanism for finding out later whether the confidence expressed at decision time was actually justified. A framework that records a decision without scheduling its own review is repeating software engineering's most consistent blind spot.
- **The evidence base matters as much as the artifact.** Given how much of software's own methodology (Part 8) is validated by consensus rather than controlled measurement, a meta-framework has a genuine opportunity to be more honest than its predecessors about which of its components are evidence-backed (premortems, reference-class forecasting, reversibility triage) versus merely well-established convention (ADRs, RFCs) versus actively unproven (spec-driven development, multi-agent architecture debate) — and to let that distinction shape how much weight each component is given, rather than treating every borrowed technique as equally trustworthy.

---

## Sources

**Software-native methods**
- Nygard, "Documenting Architecture Decisions" (2011) — cognitect.com/blog/2011/11/15/documenting-architecture-decisions
- ADR community hub — adr.github.io
- ThoughtWorks Technology Radar — thoughtworks.com/radar
- arc42 documentation — docs.arc42.org
- Increment, "Planning for change with RFCs" — increment.com/planning/planning-with-requests-for-comments
- Spike (software development) — en.wikipedia.org/wiki/Spike_(software_development); xp.c2.com/SpikeSolution.html
- SEI ATAM technical reports — sei.cmu.edu; en.wikipedia.org/wiki/Architecture_tradeoff_analysis_method
- SEI Quality Attribute Workshop collection — resources.sei.cmu.edu
- Wardley map — en.wikipedia.org/wiki/Wardley_map; wardleymaps.com
- Design sprint — en.wikipedia.org/wiki/Design_sprint; gv.com/sprint

**Company practices**
- "Design Docs at Google" — industrialempathy.com/posts/design-docs-at-google; abseil.io/resources/swe-book (Software Engineering at Google)
- Bryar & Carr, *Working Backwards*; workingbackwards.com/concepts/working-backwards-pr-faq-process
- AWS Executive Insights, "Elements of Amazon's Day 1 Culture" — aws.amazon.com/executive-insights
- Pragmatic Engineer, "Inside Stripe's Engineering Culture" and "Companies Using RFCs or Design Docs" — newsletter.pragmaticengineer.com; blog.pragmaticengineer.com/rfcs-and-design-docs
- Shopify Engineering, "A Guide to Running an Engineering Program" — shopify.engineering
- Next.js governance — nextjs.org/governance; Netlify, "Next.js Deployment Challenges" — netlify.com/blog/how-we-run-nextjs
- Pragmatic Engineer, "Netflix's Engineering Culture" — newsletter.pragmaticengineer.com/p/netflix

**Cross-domain frameworks**
- GRADE Working Group overview — publications.aap.org; ncbi.nlm.nih.gov/pmc/articles/PMC3546302; PMC5975536
- Heuer, *Psychology of Intelligence Analysis*; Wikipedia, "Richards Heuer" and "Analysis of competing hypotheses"
- RAND, "Delphi Method" and "Methodological Guidance for Conducting and Critically Appraising Delphi Panels" — rand.org
- Wikipedia, "OODA loop"; coljohnboyd.com
- "Old Advocacy, New Algorithms" — royapakzad.substack.com; spacestrategies.org, "Red Team Analysis"
- Klein, "Performing a Project Premortem," *Harvard Business Review* (2007); Mitchell, Russo & Pennington (1989); Veinott, Klein & Pierce (2010) — idl.iscram.org/files/veinott/2010
- American Bar Association, "Legal Reasoning? It's All about IRAC" — americanbar.org
- Cornell Legal Information Institute, "preponderance of the evidence" and "clear and convincing evidence" — law.cornell.edu/wex
- Wikipedia, "Cynefin framework"; Snowden & Boone, "A Leader's Framework for Decision Making," *Harvard Business Review* (2007)
- Tetlock & Gardner, *Superforecasting*; Good Judgment Project — goodjudgment.com; AI Impacts, "Evidence on good forecasting practices from the Good Judgment Project"
- Kitchenham, Dybå & Jørgensen on Evidence-Based Software Engineering — sciencedirect.com; en.wikipedia.org/wiki/Empirical_software_engineering
- Duke, *Thinking in Bets* / *How to Decide* — summarized via grahammann.net/book-notes; transactionintelligence.net, "Decision Journals"

**Emerging AI-assisted methodologies**
- DEV Community / Augment Code / BCMS, overviews of Spec-Driven Development, 2026 — dev.to; augmentcode.com; thebcms.com
- "AGENTS.md vs Architecture Decision Records" — ai.gopubby.com
- "Architecture Decision Records with Codex CLI" — codex.danielvaughan.com; arXiv:2604.04990, "Architecture Without Architects"
- Irving et al., "AI safety via debate" (2018); Du et al., "Improving factuality and reasoning in language models through multiagent debate" (2023)
- FlowHunt, "Multi-Agent AI Systems in 2026: What the Research Actually Says" — flowhunt.io
- Google DeepMind, "Co-Scientist: A multi-agent AI partner to accelerate research" — deepmind.google/blog; Gottweis et al., arXiv:2502.18864
- Wikipedia, "Kent Beck"

*This document synthesizes publicly available material as of August 2026. Company-specific practices are based on public engineering blogs, books, and third-party reporting rather than internal verification, and may not reflect current internal practice in full. Effectiveness claims are attributed to their original studies where available; where only vendor or practitioner testimony exists, this is noted explicitly rather than presented as independently verified.*
