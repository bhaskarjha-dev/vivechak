# Decision Registry: Vivechak v2.0

## Generated following Vivechak v1.1 methodology

---

---
id: D-001
title: "MCP Server Architecture Pattern"
status: proposed
door_type: one-way
date: 2026-09-23
confidence: medium
evidence_refs: []
informed_by_sessions: [R-02]
supersedes: null
superseded_by: null
amends: null
review_trigger: "After first 10 real pipeline executions via MCP"
review_date: null
prediction: "Approach B ('The Guide') will be sufficient for v1.0"
tags: [architecture, mcp, server]
authored_by: "antigravity-agent"
human_reviewed: false
schema_version: "1.1"
---

### D-001: MCP Server Architecture Pattern

**Competing Hypotheses:**

| Hypothesis | Summary |
|---|---|
| **(A) "The Guide"** | Server manages workflow and validates outputs. The host agent does all LLM work and research. Server is a project manager. |
| **(B) "The Worker"** | Server calls LLM APIs directly. Self-contained but requires API keys and duplicates host agent capabilities. |
| **(C) "The Hybrid"** | Both structural and execution tools available. Agent picks approach based on capability. |

**Current Lean:** (A) — based on analysis that host agents are already capable, and API key management adds friction.

**Open Questions:**
- Do any successful workflow MCP servers use the "Worker" pattern? (R-02 will answer)
- Is Approach A sufficient for fully autonomous pipeline execution? (Needs testing)

---

---
id: D-002
title: "Multi-Scope Methodology Design"
status: proposed
door_type: two-way
date: 2026-09-23
confidence: medium
evidence_refs: []
informed_by_sessions: [R-03, R-06]
supersedes: null
superseded_by: null
amends: null
review_trigger: "After 5 decision-level and 5 comparison-level pipelines executed"
review_date: null
prediction: "Three fixed scope levels with separate generators will be more usable than a single adaptive generator"
tags: [methodology, scope, generators]
authored_by: "antigravity-agent"
human_reviewed: false
schema_version: "1.1"
---

### D-002: Multi-Scope Methodology Design

**Competing Hypotheses:**

| Hypothesis | Summary |
|---|---|
| **(A) Three fixed scope levels** | Separate generators for Project, Decision, and Comparison. Each has distinct input format and output format. |
| **(B) Continuous scope spectrum** | One generator with a scope parameter (1-10 depth) that scales the pipeline. |
| **(C) Single adaptive generator** | One generator that detects scope from input and adapts automatically. |

**Current Lean:** (A) — based on analysis that distinct entry points are more discoverable and easier for agents to select.

**Open Questions:**
- How do cross-domain methodologies (medical systematic reviews, intelligence analysis) handle scope? (R-03 will answer)
- Are there scope levels we haven't considered? (R-03 will answer)
- Does reducing scope compromise evidence quality? (R-03 will answer)

---

---
id: D-003
title: "MCP Tool API Surface"
status: proposed
door_type: one-way
date: 2026-09-23
confidence: low
evidence_refs: []
informed_by_sessions: [R-04]
supersedes: null
superseded_by: null
amends: null
review_trigger: "Before first public release — API surface must be finalized"
review_date: null
prediction: "8-10 fine-grained tools will be more usable than 5 coarse-grained tools"
tags: [api, mcp, tools, design]
authored_by: "antigravity-agent"
human_reviewed: false
schema_version: "1.1"
---

### D-003: MCP Tool API Surface

**Competing Hypotheses:**

| Hypothesis | Summary |
|---|---|
| **(A) 10 fine-grained tools** | Each step of the workflow gets its own tool. Clear, single-responsibility, but many tools for the agent to understand. |
| **(B) 5 coarse-grained tools** | Fewer tools, each covering multiple steps. Simpler for the agent but less granular control. |
| **(C) Dynamic tool registration** | Server exposes different tools based on pipeline state (e.g., `vivechak_run_gate` only appears when pipeline is complete). |

**Current Lean:** (A) — but confidence is low. Needs R-04 research on MCP tool discoverability patterns.

**Open Questions:**
- How many tools can an agent effectively reason about? (R-04 will investigate)
- Should validation be a separate tool or built into save tools? (R-04 will answer)
- Should MCP Resources and Prompts be used alongside Tools? (R-04 will answer)

---

---
id: D-004
title: "Distribution Strategy"
status: proposed
door_type: two-way
date: 2026-09-23
confidence: medium
evidence_refs: []
informed_by_sessions: [R-05]
supersedes: null
superseded_by: null
amends: null
review_trigger: "After first 3 months of distribution — track which channels get installs"
review_date: null
prediction: "Agent Plugins + go install + GitHub Releases will cover 90%+ of users"
tags: [distribution, packaging, agent-plugins]
authored_by: "antigravity-agent"
human_reviewed: false
schema_version: "1.1"
---

### D-004: Distribution Strategy

**Competing Hypotheses:**

| Hypothesis | Summary |
|---|---|
| **(A) Agent Plugins 1.0.0 only** | Minimal distribution. Relies on agent hosts supporting the spec. |
| **(B) Agent Plugins + go install + GitHub Releases** | Three channels covering plugin ecosystem, Go developers, and everyone else. |
| **(C) Full matrix** | Agent Plugins + go install + GitHub Releases + Homebrew + Scoop + platform packages. Maximum coverage but high maintenance. |

**Current Lean:** (B) — balanced coverage without maintenance burden.

**Open Questions:**
- What percentage of target users have Go installed? (R-05 may answer)
- Do all major agent hosts support Agent Plugins 1.0.0? (R-05 will answer)
- Is Homebrew/Scoop worth the maintenance? (R-05 will assess)
