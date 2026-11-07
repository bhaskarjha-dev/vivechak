# Vivechak (विवेचक) — Manual Dogfooding & Benchmarking Harness
### Multi-Scope Interactive Evaluation Prompt and 5-Stage Agent Audit Protocol

> **Purpose:** Use this document to manually test and benchmark any AI assistant, coding agent, or model (Claude, Cursor, Antigravity, OpenCode, VS Code Copilot) using Vivechak.  
> **Structure:**  
> - **Stage 1 (System Vision Prompt):** Initial prompt pasted into the agent to execute an end-to-end multi-scope project lifecycle.  
> - **Stage 2 (Post-Execution Audit Protocol):** 5-phase structured questionnaire pasted sequentially to evaluate the agent's output, epistemic quality, and Vivechak's developer experience.

---

## Stage 1: The End-to-End System Vision Prompt

*Copy and paste the prompt below into your AI coding assistant:*

```markdown
Hey, we are starting a brand new project: a distributed real-time IoT Telemetry Ingestion & Anomaly Detection Platform ("OmniPulse").

Here is our system vision and constraints:
- Scale: 50,000 IoT sensor devices streaming metrics (temperature, pressure, vibration) at 10 Hz (~500k events/sec peak).
- Requirements:
  1. Low latency anomaly detection (< 500ms from sensor emission to alert trigger).
  2. Hot timeseries storage for 30-day live querying, plus cold tier for archival.
  3. Ephemeral edge-to-cloud caching/buffering to handle network partitions.
- Core dilemmas we are facing:
  - Dilemma 1 (One-Way Door): Streaming message bus selection (Apache Kafka vs Redpanda vs NATS JetStream) — team is small (3 engineers), operational simplicity matters as much as raw throughput.
  - Dilemma 2 (Two-Way Door): Hot timeseries datastore (ClickHouse vs TimescaleDB).
  - Dilemma 3 (Rapid Comparison): In-memory edge cache (Redis vs Dragonfly).

I want to use Vivechak to systematically research and plan our architecture before we write any code. Please:
1. Plan the overall project architecture pipeline in Vivechak with initial candidate decisions.
2. Carry out the necessary research sessions and option comparisons.
3. Record the key architectural decisions (ADRs) with evidence grades and reversal triggers, and stress-test the high-stakes message bus decision with an adversarial red-team challenge.
4. Compile the Founding Architecture Document (FAD) and verify our pre-codebase exit criteria with the Phase 0 Gate.

Go ahead and drive this end-to-end using Vivechak.
```

---

## Stage 2: The 5-Phase Audit & Interrogation Protocol

*Once the agent completes the Stage 1 task, paste these audit prompts to inspect the agent's work and evaluate Vivechak's engine ergonomics:*

### Audit 1: Tool Sequence & Guidance Ergonomics
```markdown
Perform Audit 1 on our session:
1. List the exact chronological sequence of every Vivechak tool you called (with parameters passed).
2. For each tool call, did it succeed on the first attempt or did you encounter any error or retry?
3. Did the "next_step" guidance returned by Vivechak give you clear, actionable instructions on what to do next at every turn?
```

### Audit 2: Research Quality & Evidence Grading
```markdown
Perform Audit 2 on our research quality:
1. What evidence grades (A–E) did you assign to the findings, and were citations qualified with URLs/RFCs?
2. Did any session trigger quality coaching warnings (e.g., Q-GRADE-INFLATION, Q-MISSING-CONCERNS, or Q-LOW-FINDINGS)?
3. For the message bus One-Way Door decision, did you document substantive Discovered Concerns (>= 50 chars), explicit rejected alternatives, and quantifiable reversal triggers?
```

### Audit 3: Workspace Layout & Decision Hygiene
```markdown
Perform Audit 3 on our workspace layout:
1. Inspect research/DECISIONS.md: How were the decisions formatted (collapsible <details> blocks, YAML frontmatter, table overview)?
2. When candidate decisions were first saved in "status: proposed", did Vivechak handle them cleanly without throwing premature evidence warnings?
3. In the research DAG, how were the synthesis dependencies structured, and did Vivechak resolve all topologically preceding sessions cleanly?
```

### Audit 4: Verification & Exit Gate Clearance
```markdown
Perform Audit 4 on verification and exit gates:
1. What was the output when you ran workspace validation (vivechak_review with mode="validate")? Were there any false-positive errors or warnings?
2. When you evaluated the Phase 0 Gate (vivechak_gate / vck gate), how were Track A (Two-Way Doors) and Track B (One-Way Doors) scored?
3. Did the gate authorize proceeding to implementation, and did it save research/PHASE-0-GATE.md?
```

### Audit 5: Developer Experience (DX) & Engine Friction
```markdown
Perform Audit 5 on Developer Experience (DX) and Engine Friction:
1. As an AI agent using Vivechak, where did you experience any friction, hesitation, or ambiguity in tool descriptions or inputs?
2. Did any parameter naming feel unnatural or redundant?
3. What were the standout strengths that helped you generate high-confidence architecture, and what is your top recommendation to make Vivechak even smoother?
```
