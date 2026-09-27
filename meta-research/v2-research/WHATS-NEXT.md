# Vivechak Strategic Deep Dive: Scope Expansion, MCP Architecture & Competitive Reality

> **Date:** 2026-09-23
> **Scope:** Response to three interconnected questions about Vivechak's future
> **Principle:** Nothing is Sacred. First-principles only.

---

## Part 1: Vivechak as a General-Purpose Research Framework

### Your Instinct Is Correct — But It's Bigger Than Repositioning

You're not just suggesting a name change from "pre-development" to "any-stage." You're identifying that **Vivechak's methodology works at multiple granularities**, and the current framing artificially constrains it to one use case.

Let me lay out what the methodology actually provides, stripped of its "pre-development" packaging:

| Capability | What It Actually Does | Stage-Independent? |
|---|---|---|
| Complexity scoring (0–24) | Sizes the research effort proportionally | ✅ Yes — works for a project OR a single decision |
| Session DAG | Structures parallel/sequential research | ✅ Yes — 1 session for a comparison, 30 for a project |
| 5-block prompt anatomy | Structures unbiased research questions | ✅ Yes — works at any granularity |
| Evidence grading (A–E) | Prevents hallucination from driving decisions | ✅ Yes — always relevant |
| One-Way / Two-Way Door classification | Calibrates rigor to reversibility | ✅ Yes — applies at any stage |
| Bounded Exploration Mandate | Prevents premature narrowing | ✅ Yes — always relevant |
| Weighted Evaluation Protocol | De-biases comparisons | ✅ Yes — works for any comparison |
| ADR with reversal triggers | Tracks decisions with decay conditions | ✅ Yes — any decision, any time |
| Conflict Resolution (ACH) | Resolves contradictory findings | ✅ Yes — always relevant |
| FAD synthesis | Compiles findings into architecture doc | ⚠️ Partially — the "founding" framing is stage-specific |
| Phase 0 Gate | Verifies readiness before coding | ⚠️ Partially — the gate concept works but the name implies pre-coding |

**The methodology is genuinely stage-independent.** The only things that need adapting are:

1. **The GENERATOR.md prompt** — Currently expects a "project vision." Should also accept a "decision context" or "research question" at a specific scope level.
2. **The output format** — A full FAD makes sense for project-level research. For a single mid-development decision (like "MCP vs Agent Plugins"), the output should be a well-grounded ADR, not a full architecture document.
3. **The complexity scoring** — Some dimensions (Expected Longevity, Investment Horizon) are project-level. For a single decision, you'd naturally score lower, which correctly produces a Tier 1 (1–3 sessions) pipeline. The scoring *already handles this* — you just need to apply it to the decision's scope, not the project's scope.

### The Three Natural Scope Levels

```
┌─────────────────────────────────────────────────────┐
│  LEVEL 1: PROJECT SCOPE                             │
│  "I'm building a new SaaS platform"                 │
│  Input: Project vision dump                         │
│  Output: Full FAD + Decision Registry               │
│  Sessions: 4–30 (Tier 2–4)                         │
│  Current Vivechak: ✅ Fully supported               │
├─────────────────────────────────────────────────────┤
│  LEVEL 2: DECISION SCOPE                            │
│  "Should we adopt MCP or build a custom plugin       │
│   system for Vivechak?"                             │
│  Input: Decision context + constraints              │
│  Output: Grounded ADR with evidence                 │
│  Sessions: 1–3 (Tier 1)                            │
│  Current Vivechak: ⚠️ Works but feels heavyweight    │
├─────────────────────────────────────────────────────┤
│  LEVEL 3: COMPARISON SCOPE                          │
│  "PostgreSQL 17 vs CockroachDB for our write-heavy  │
│   workload with these specific requirements"         │
│  Input: Specific options + criteria                  │
│  Output: Weighted evaluation matrix + recommendation │
│  Sessions: 1 (single focused session)               │
│  Current Vivechak: ⚠️ The WEP exists but no direct   │
│  entry point for it                                 │
└─────────────────────────────────────────────────────┘
```

### What Needs to Change (Honest Assessment)

**The methodology needs almost no changes.** The adaptive scaling (Tier 1 = 1–3 sessions) already handles decision-level research. The Weighted Evaluation Protocol already handles comparisons. The ADR template already handles any-stage decisions.

**What needs to change is the ENTRY POINT.** Currently there's one door into Vivechak: paste an 18KB generator prompt with a project vision. For decision-level and comparison-level research, you need simpler entry points:

| Level | Current Entry | Needed Entry |
|---|---|---|
| Project | 18KB generator prompt → full pipeline | Works as-is |
| Decision | Same 18KB prompt (overkill) | A shorter "decision research" prompt or tool that generates 1–3 session prompts + an ADR template |
| Comparison | No direct entry | A "structured comparison" prompt or tool that generates a single session with WEP |

### Your MCP vs Agent Plugins Example

Let's trace how this would actually work:

You're mid-development of Vivechak. You want to research: "Should we adopt MCP, or should we also support the Agent Plugins spec?"

**With current Vivechak (awkward):**
1. Paste the full 18KB generator prompt
2. Describe "Vivechak Engine" as the project
3. Get a full 4–9 session pipeline (overkill — you only need 1–2 sessions for this decision)
4. Most sessions are irrelevant to this specific question

**With scope-aware Vivechak (natural):**
1. Describe the decision context: "I'm building a Vivechak MCP server. I need to decide whether to also support Agent Plugins 1.0.0. Context: [constraints, requirements]"
2. The decision scores low on complexity (most dimensions are 0–1), yielding Tier 1 (1–3 sessions)
3. Get a focused pipeline: Session 1 = landscape survey (MCP vs A2A vs Agent Protocol), Session 2 = deep comparison with WEP
4. Output: A single, well-grounded ADR

**This is how Vivechak SHOULD work.** The methodology already supports it; the packaging doesn't.

### Honest Pushback on This Idea

**Risk: scope creep.** Expanding from "pre-development research" to "any-stage research" dramatically broadens the tool's surface area. Every expansion needs testing.

**Risk: dilution.** Vivechak's clearest value proposition is "before you code, research your architecture." Making it "research anything, anytime" makes it harder to explain what it IS.

**Risk: the GENERATOR.md prompt is calibrated for project-level input.** The complexity scoring dimensions (Investment Horizon, Expected Longevity, Coordination Complexity) don't always make sense for a single mid-project decision. You'd need either a separate scoring rubric or explicit guidance on how to score a decision-scope input.

**Mitigation:** Start with a second, shorter generator prompt for decision-level research. Don't modify the existing project-level generator. This is additive, not destructive.

---

## Part 2: MCP Server — First Principles End-to-End Design

### What MCP Actually Is (For Someone With Zero Knowledge)

MCP is a standardized way for AI agents to use tools. Think of it as a **USB standard for AI tools**:

```
Without MCP:                          With MCP:
┌──────────┐  custom code  ┌──────┐   ┌──────────┐  standard  ┌──────┐
│  Claude   │─────────────→│ Tool │   │  Claude   │──protocol──│ Tool │
└──────────┘              └──────┘   │  Cursor   │──────────→│Server│
                                      │  VS Code  │           └──────┘
┌──────────┐  different    ┌──────┐   │  AGY      │
│  Cursor   │─────────────→│ Tool │   └──────────┘
└──────────┘  custom code  └──────┘   Any agent talks to any server
```

An MCP server is a small program that says: "I have these tools. Here's what each one does. Here's what input each needs. Call me when you need them."

The AI agent (Claude, Cursor, Antigravity, etc.) sees the available tools, understands their descriptions, and decides when to call them based on what the user asks.

### The Fundamental Design Question

There are two architecturally different approaches to a Vivechak MCP server:

#### Approach A: "The Worker" — MCP Server Does the LLM Work

```
User → "Research architecture for my project"
  → Agent sees vivechak_generate_pipeline tool
  → Agent calls it with {vision: "..."}
  → MCP Server internally calls an LLM API with GENERATOR.md prompt
  → MCP Server parses response, saves files
  → Returns: "Pipeline generated with 6 sessions"
  → Agent calls vivechak_execute_session({session: "T2-01"})
  → MCP Server internally calls an LLM with the session prompt + web search
  → MCP Server saves output
  → Returns: "Session T2-01 complete"
  ... and so on
```

**Pros:** Clean abstraction. The host agent doesn't need to understand Vivechak internals.
**Cons:** Needs LLM API keys configured in the MCP server. Duplicates capabilities the host agent already has. Expensive (every tool call makes LLM API calls). The host agent can't use its own superior research capabilities (e.g., Gemini Deep Research Max).

#### Approach B: "The Guide" — MCP Server Manages Workflow, Agent Does the Work

```
User → "Research architecture for my project"
  → Agent sees vivechak tools
  → Agent calls vivechak_init({project_path: "..."})
  → MCP Server creates directories, copies templates
  → Returns: "Workspace ready"
  → Agent calls vivechak_get_generator_prompt({vision: "..."})
  → MCP Server reads GENERATOR.md, inserts vision, returns the complete prompt
  → Agent runs this prompt ITSELF (using its own LLM capabilities)
  → Agent gets the pipeline + decisions output
  → Agent calls vivechak_save_pipeline({content: "..."})
  → MCP Server parses, validates, saves to correct files
  → Agent calls vivechak_next_session()
  → MCP Server checks DAG, returns: "T2-01 is ready. Here's the prompt: ..."
  → Agent executes the research session ITSELF (web search, synthesis)
  → Agent calls vivechak_save_session({session_id: "T2-01", content: "..."})
  → MCP Server validates (evidence grades present? YAML frontmatter valid?)
  → ... and so on
```

**Pros:** Leverages the host agent's full capabilities. No API key duplication. The host agent can use Gemini Deep Research, its own web search, whatever it has. Cheaper. More capable.
**Cons:** More tool calls. Requires the host agent to be capable enough to execute research sessions well.

#### Approach C: "The Hybrid" — Best of Both

```
MCP Server provides BOTH approaches:
- Structural tools (init, save, validate, status, next) — always available
- Execution tools (generate_pipeline, execute_session) — available IF API keys configured
- The host agent picks which approach to use based on its capabilities
```

### My Recommendation: Approach B ("The Guide")

**Why:** The host agents (Claude, Antigravity, Cursor) are already incredibly capable. They have web search, URL reading, deep reasoning, and file management. Making the MCP server do LLM calls would duplicate what the host agent does better. The MCP server's value should be **methodology enforcement and workflow management**, not LLM orchestration.

The MCP server should be the **project manager** that knows the Vivechak methodology, tracks state, validates outputs, and guides the agent through the workflow. The agent does the actual intellectual work.

### End-to-End Walkthrough: All Three Scenarios

#### Scenario 1: Agent Sets Up, User Researches in Browser

```
User to Antigravity: "I want to research architecture for a new project.
Here's my idea: [brain dump about a real-time collaborative document editor]"

Agent: I'll set up a Vivechak research pipeline for you.

→ Agent calls vivechak_init({project_path: "D:/dev/pro/my-project"})
  ← "Workspace created at my-project/research/ with 4 templates"

→ Agent calls vivechak_get_generator_prompt({
    vision: "[user's brain dump]",
    scope: "project"  // vs "decision" or "comparison"
  })
  ← Returns the complete 18KB generator prompt with vision inserted

→ Agent executes this prompt itself (it IS an LLM, after all)
  ← Agent produces RESEARCH-PIPELINE.md + DECISIONS.md

→ Agent calls vivechak_save_pipeline({
    pipeline_content: "[the generated pipeline]",
    decisions_content: "[the generated decisions]"
  })
  ← MCP Server parses, validates YAML, saves files, returns summary:
     "Pipeline saved: 7 sessions (Tier 2), 5 proposed decisions"

Agent to user: "Your research pipeline is ready at my-project/research/.
Here are the 7 sessions you need to run:
  Layer 0 (run in parallel):
    T2-01: Primary Datastore Selection
    T2-02: Real-Time Sync Architecture
    T2-03: Auth & Multi-Tenancy
  Layer 1:
    T2-04: Conflict Resolution Protocol (depends on T2-01, T2-02)
    ...
  
Copy each session prompt from RESEARCH-PIPELINE.md and run it in
Gemini Deep Research or ChatGPT Deep Research. Save outputs to
research/sessions/."

── User goes to browser, runs sessions manually ──
── User saves outputs to research/sessions/ ──

User: "I've completed all sessions. Now what?"

→ Agent calls vivechak_status({project_path: "..."})
  ← MCP Server scans sessions/ directory, checks files exist,
     validates evidence grades present, returns:
     "7/7 sessions complete. 3 decisions still in 'proposed' status.
      Next: Record decisions for D-001, D-002, D-003"

→ Agent reads session outputs, formulates decisions
→ Agent calls vivechak_record_decision({...}) for each
  ← MCP Server validates schema, saves to DECISIONS.md

... continues through synthesis and gate ...
```

#### Scenario 2: Agent Generates Pipeline AND Research, User Reviews

```
User to Antigravity: "Use Vivechak to research architecture for
this project: [brain dump]. Generate the pipeline and do the research."

→ Agent calls vivechak_init(...)
→ Agent calls vivechak_get_generator_prompt({vision: "...", scope: "project"})
→ Agent executes the generator prompt itself
→ Agent calls vivechak_save_pipeline(...)

Now the agent iterates through sessions:

→ Agent calls vivechak_next_session()
  ← "Ready sessions: T2-01, T2-02, T2-03 (Layer 0, parallel)"
     Returns session prompts

→ Agent executes T2-01's research prompt ITSELF
   (using its own web search, URL reading, reasoning capabilities)
→ Agent calls vivechak_save_session({session_id: "T2-01", content: "..."})
  ← MCP Server validates: "⚠️ Warning: 2 claims have no evidence grade.
     Fix before proceeding." OR "✅ Valid. Evidence density: 4 Grade A/B claims."

→ Agent fixes issues, re-saves
→ Agent repeats for T2-02, T2-03 (in parallel if agent supports it)

→ Agent calls vivechak_next_session()
  ← "Layer 0 complete. Ready sessions: T2-04, T2-05 (Layer 1)"

... continues until all sessions done ...

→ Agent calls vivechak_status()
  ← "All sessions complete. Ready for decision recording."

→ Agent reads sessions, formulates and records decisions
→ Agent calls vivechak_synthesize({}) 
  ← Returns the FAD template with all session summaries injected,
     ready for the agent to synthesize
→ Agent synthesizes FAD
→ Agent calls vivechak_run_gate({})
  ← MCP Server checks: all decisions locked? Evidence grades sufficient?
     Conflicts resolved? Returns structured gate report.

Agent to user: "Research complete. Here's your FAD and gate report.
Key decisions: [summary]. Please review."
```

#### Scenario 3: Fully Autonomous (Zero Manual Research)

Same as Scenario 2, but the agent handles everything including web research.

**The key difference is in the agent's capability, not the MCP server.** The MCP server is identical in all three scenarios. What changes is how much the host agent does vs. the user.

This is the beauty of Approach B: **the MCP server doesn't care who does the work.** It manages the workflow, validates outputs, and tracks state. Whether a human, an agent, or Gemini Deep Research produces the session output — the MCP server validates it the same way.

### Proposed MCP Tools (Revised from First Principles)

| Tool | Purpose | Input | Output |
|---|---|---|---|
| **`vivechak_init`** | Create workspace, copy templates | `project_path` | Confirmation + structure created |
| **`vivechak_get_generator_prompt`** | Return the complete generator prompt with vision inserted | `vision`, `scope` (project/decision/comparison), `date` | The ready-to-execute prompt text |
| **`vivechak_save_pipeline`** | Parse, validate, and save generated pipeline + decisions | `pipeline_content`, `decisions_content`, `project_path` | Summary (session count, tier, decision count) + validation warnings |
| **`vivechak_status`** | Full pipeline status report | `project_path` | Sessions done/pending, decisions status, conflicts, next steps |
| **`vivechak_next_session`** | Get the next executable session(s) based on DAG | `project_path` | Session IDs, prompts, and dependency status |
| **`vivechak_save_session`** | Validate and save a session's research output | `session_id`, `content`, `project_path` | Validation report (evidence grades, YAML, quality rubric) |
| **`vivechak_record_decision`** | Record/update a decision in DECISIONS.md | Decision fields (YAML) + body content | Validation report |
| **`vivechak_synthesize`** | Prepare synthesis context (all sessions + decisions) | `project_path` | Template + injected session summaries for agent to synthesize |
| **`vivechak_run_gate`** | Execute Phase 0 Gate checks programmatically | `project_path` | Structured gate report (Track A + Track B results) |
| **`vivechak_validate`** | Validate any Vivechak artifact (session, ADR, FAD) | `file_path`, `artifact_type` | Validation results with specific issues |

### What the MCP Server DOESN'T Do

- **Doesn't call LLM APIs** — the host agent does the thinking
- **Doesn't do web research** — the host agent uses its own search capabilities
- **Doesn't decide what to research** — it provides the methodology, the agent/user decides
- **Does validate** — it enforces the Vivechak methodology (evidence grades, YAML schemas, DAG dependencies, gate criteria)
- **Does manage state** — it tracks what's done, what's pending, what's next
- **Does provide the right prompt at the right time** — it's the project manager

### MCP Server Technology Stack

| Component | Choice | Rationale |
|---|---|---|
| Language | Python | Widest MCP SDK support, largest LLM ecosystem |
| MCP SDK | `mcp` (official Python SDK) | Official, maintained by Anthropic |
| Transport | stdio (primary) + streamable HTTP (optional) | stdio works with all local agents; HTTP enables remote |
| State | File-based (`.vivechak/state.json` in project) | No database needed; state lives with the project |
| Validation | Pydantic models | Type-safe YAML/JSON validation |
| Template engine | Jinja2 or string formatting | For injecting vision into generator prompt |

---

## Part 3: The Competitor Question — Why the Gap Exists

### You're Right to Question This

The "zero competitors" finding is real but needs context. The reason isn't that nobody has thought of this — it's that **the competitive landscape operates at different layers:**

### Layer 1: Behavioral Competitors (Biggest Threat)

These are what developers actually DO instead of using Vivechak:

| Behavior | "Market Share" | Quality | Cost |
|---|---|---|---|
| **"Just ask ChatGPT"** | ~80% of developers | Low — no structure, no evidence grading, framing bias | Free |
| **Single Deep Research session** | ~10% | Medium — better than ad-hoc, but single-session, no cross-session synthesis | Low |
| **Write a design doc / RFC** | ~5% (mostly larger teams) | Medium — records decisions but doesn't generate research | Medium (time) |
| **Hire an architecture consultant** | ~2% (funded startups) | High | High (\$\$\$) |
| **Do nothing (vibe code)** | ~3% | None | Free |

Vivechak's real competition is **the habit of asking ChatGPT once and going with whatever it says.** This behavior is free, fast, and feels adequate — until it isn't.

### Layer 2: Tool Competitors (Partial Overlap)

| Tool | What It Does | Where Vivechak Differs |
|---|---|---|
| **MADR / adr-tools** | Records decisions in markdown | Vivechak GENERATES the research that INFORMS decisions. MADR just records them. |
| **Elicit / Semantic Scholar** | AI-powered academic literature research | Similar concept (structured multi-source research with evidence) but for academic papers, not software architecture |
| **Perplexity Spaces** | Organized multi-query research workspaces | No methodology, no evidence grading, no decision framework. But could evolve. |
| **Claude Projects / Gemini Gems** | Persistent AI workspaces with custom instructions | A user COULD set up a project with Vivechak templates as custom instructions. This is a "poor man's Vivechak." |
| **ThoughtWorks Technology Radar** | Industry-wide technology assessments | Not project-specific, not customizable |
| **StackShare** | Community technology comparisons | No structured research, no evidence grading |

### Layer 3: Emerging Threats

| Threat | Timeline | Risk Level |
|---|---|---|
| **Deep Research tools getting multi-session capability** | 6–12 months | **HIGH** — If Gemini Deep Research adds "research plan" mode that runs multiple sub-sessions and synthesizes, it eats Vivechak's orchestration value |
| **AI coding agents that do implicit architecture research** | Already happening | **MEDIUM** — Claude Code and Cursor already suggest architecture, but without structure. If they add evidence grading, that's a problem |
| **Someone building "Vivechak but as a product"** | Unknown | **MEDIUM** — The category is uncontested. First mover with a usable tool wins |

### Why Nobody Has Built This Yet (Honest Analysis)

1. **The market is invisible.** Developers don't search for "pre-development research frameworks." The category doesn't exist in anyone's mental model. You can't sell what people don't know they need.

2. **The pain is delayed.** The cost of bad architecture shows up 6 months later. Developers blame "scaling problems," not "inadequate research." The causal link between poor upfront research and downstream failure is real but not obvious.

3. **"Just ask ChatGPT" feels free.** The perceived cost of ad-hoc research is zero. The actual cost (biased evidence, missed alternatives, no decision tracking) is hidden.

4. **The methodology is hard to productize.** Multi-session research with evidence grading, cross-session synthesis, conflict resolution — these are genuinely hard problems. Easier to build yet another chatbot wrapper.

5. **The closest analogs are in domains with life-or-death stakes.** Medical systematic reviews (Covidence), intelligence analysis (ACH), legal research (Westlaw) — these domains have structured research because wrong answers kill people or cost millions. Software architecture hasn't reached that level of consequence awareness.

### The Strategic Implication

**Vivechak's moat is the methodology, not the delivery format.** If someone builds a slick tool that does evidence-graded multi-session research for software architecture, they'll cite the same underlying research (framing bias, GRADE, premortems). Vivechak's advantage is being first and having the deepest methodology.

**But a methodology without a usable tool is an academic paper, not a product.** The MCP server is the minimum viable delivery mechanism that converts the methodology into something developers can actually use.

---

## Part 4: Agent Plugins 1.0.0 — It's Real

**Agent Plugins 1.0.0 IS a real specification.** Released **August 6, 2026.** [Grade A — official spec]

### What It Is

An open, vendor-neutral specification for **packaging** AI agent extensions into a single, portable directory structure. It solves the "fork-and-drift" problem by standardizing how Agent Skills and MCP server configurations are bundled.

### How It Relates to MCP

**Complementary, not competitive.** Think of it this way:
- **MCP** = the runtime protocol (how the agent talks to the tool)
- **Agent Plugins** = the packaging format (how the tool is distributed and discovered)

A compliant plugin directory includes:
- `plugin.json` — manifest
- `mcp.json` — MCP server configuration (optional)
- `skills/` — Agent Skills directory

### Who's Behind It

Technical Steering Committee: **Vercel, Amazon (AWS), Cursor (Anysphere), Microsoft (GitHub), OpenAI, Google.**

Supported by: ChatGPT, Codex, Cursor, GitHub Copilot, Kiro, VS Code.

> **Note:** Anthropic created MCP and Agent Skills but is NOT on the Agent Plugins steering committee. This suggests a potential ecosystem split worth watching.

### Implication for Vivechak

**A Vivechak MCP server should be packaged as an Agent Plugin.** This gives it maximum distribution — installable by any agent that supports the spec (which is most of them). The MCP server provides the runtime tools; the Agent Plugin packaging provides the discovery and installation.

This is a **Two-Way Door** — the packaging format is a thin wrapper that can be changed easily.

---

## Part 4b: Updated Competitor Findings

The latest research found additional tools in adjacent spaces:

| Tool | What It Does | Overlap with Vivechak | Threat Level |
|---|---|---|---|
| **Taskade** | Agentic workspace: generates component diagrams, task backlogs, and AI agents that critique designs for trade-offs | Generates architecture artifacts with AI critique | **Medium** — does generation but no structured multi-session research or evidence grading |
| **Workik** | AI assistance for cloud, database, and backend planning. Generates ADRs and API docs | ADR generation | **Low** — focused on single-pass generation |
| **Eraser.io (DiagramGPT)** | Translates natural language to system architecture diagrams | Architecture visualization | **Low** — different focus (diagrams, not research) |
| **Elicit** | Systematic reviews of technical literature | Structured multi-source research with evidence synthesis | **Medium** — similar concept but for academic papers |
| **CatchAll** | Web search monitoring API for enterprise research automation | Automated research tracking | **Low** — monitoring, not decision research |

**Updated assessment:** The competitive landscape is slightly more populated than "zero competitors," but none of these tools do what Vivechak does: **multi-session evidence-graded research → decision tracking → synthesis → exit gate.** Taskade and Elicit are the closest conceptual neighbors, but they operate differently.

---

## Part 5: Synthesis — What This All Means

### The Three Moves

```
┌─────────────────────────────────────────────────────┐
│  MOVE 1: Scope the Methodology (weeks)              │
│  Add decision-level and comparison-level entry       │
│  points. A second, shorter generator prompt for      │
│  "I need to research ONE decision" use cases.        │
│  This is a methodology change, not a software one.   │
├─────────────────────────────────────────────────────┤
│  MOVE 2: Build the MCP Server (weeks)               │
│  "The Guide" pattern — manages workflow, validates   │
│  outputs, tracks state. The host agent does the      │
│  intellectual work. 10 tools. Python. File-based     │
│  state. Works with any MCP-compatible agent.         │
├─────────────────────────────────────────────────────┤
│  MOVE 3: Prove Demand (ongoing)                     │
│  Publish real case studies, measure adoption,        │
│  collect feedback. Without this, everything else     │
│  is building in the dark.                           │
└─────────────────────────────────────────────────────┘
```

### What Move 1 + Move 2 Enables

After these two moves, a developer using ANY MCP-compatible AI agent could say:

> "I'm building a real-time collaborative editor. Use Vivechak to research the architecture."

And the agent would autonomously:
1. Initialize the workspace (MCP: `vivechak_init`)
2. Get and execute the generator prompt (MCP: `vivechak_get_generator_prompt` → agent executes)
3. Save the pipeline (MCP: `vivechak_save_pipeline`)
4. Execute each session in DAG order (MCP: `vivechak_next_session` → agent researches → `vivechak_save_session`)
5. Record decisions (MCP: `vivechak_record_decision`)
6. Synthesize the FAD (MCP: `vivechak_synthesize` → agent synthesizes)
7. Run the gate (MCP: `vivechak_run_gate`)

OR the same developer could say:

> "I'm mid-project and need to decide between MCP and Agent Plugins. Use Vivechak to research this decision."

And the agent would run a focused 1–2 session investigation and produce a grounded ADR.

**Same methodology. Same tools. Different scope. This is what Vivechak should be.**
