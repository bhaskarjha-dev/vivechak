# Sample Adoption Walkthrough: Decision Scope (PostgreSQL to ClickHouse Migration)

> **Scope:** Decision Scope (`ScopeDecision`)  
> **Investment:** 1–3 Sessions (~30 minutes)  
> **Overhead:** Zero Project DAG, Zero FAD  
> **Deliverables:** `research/D-01-clickhouse-migration.md` (Locked ADR) and `research/DECISIONS.md` (Registry)

This walkthrough shows an AI coding agent using Vivechak to resolve an architectural fork in an active codebase: deciding whether to migrate analytical event storage from PostgreSQL to ClickHouse.

---

## Step 1: The Architectural Dilemma

The developer prompts the agent:
> *"Vivechak decision: Should we migrate our primary event log storage from PostgreSQL to ClickHouse? We currently ingest 25M events/day, query latency on 30-day aggregations is degrading, but our operations team has zero ClickHouse experience."*

The agent calls `vivechak_prepare_generator` with `scope="decision"`:

**Tool Call:** `vivechak_prepare_generator`
```json
{
  "scope": "decision",
  "context": "Migrate event log storage from PostgreSQL to ClickHouse. Ingesting 25M events/day, query latency on 30-day aggregations degrading. Ops team has zero ClickHouse experience."
}
```

**Tool Response:**
```json
{
  "success": true,
  "message": "Decision generator prompt prepared.",
  "data": {
    "scope": "decision",
    "prompt": "# GENERATE: Architectural Decision Plan (ADR)\n\n## DILEMMA\nMigrate event log storage from PostgreSQL to ClickHouse...\n\n## COMPLEXITY PROFILING\n- Reversibility (R): 2 (High data migration cost)\n- Novelty (N): 2 (Team lacks ClickHouse operational expertise)\n- Blast Radius (B): 2 (Core analytics pipeline)\n- Regulatory/Compliance (X): 0 (Standard internal logs)\n\nRoutes to 2 Sessions: L (Landscape & Benchmark) + F (Falsification & Failure Modes)...",
    "generator_file": "GENERATOR-DECISION.md"
  },
  "next_step": "Execute the decision generator prompt using frontier LLM capabilities with web search, then call vivechak_save_plan to persist the decision plan."
}
```

---

## Step 2: Saving the Decision Plan

The agent executes the generator, producing a tailored 2-session research plan (`D-01-S1` and `D-01-S2`) and calls `vivechak_save_plan`:

**Tool Call:** `vivechak_save_plan`
```json
{
  "scope": "decision",
  "decision_id": "D-01",
  "slug": "clickhouse-migration",
  "content": "# Decision Plan: D-01\n| Decision | D-01 |\n| Scope | decision |\n| Door Type | one-way |\n\n## Sessions\n1. D-01-S1: ClickHouse Ingestion Throughput & Partitioning Benchmark\n2. D-01-S2: Operational Failure Modes, Small-Batch Pitfalls & Reversibility"
}
```

**Tool Response:**
```json
{
  "success": true,
  "message": "Decision plan saved in research/D-01-clickhouse-migration-plan.md. 2 templates copied (DECISIONS, CONFLICT-RESOLUTION). Zero project DAG overhead.",
  "next_step": "Call vivechak_next_session to retrieve the first research prompt (D-01-S1)."
}
```

---

## Step 3: Executing Research Sessions

The agent executes `D-01-S1` and `D-01-S2` using `vivechak_next_session` and `vivechak_save_session`:

1. **Session S1 Findings:**
   - ClickHouse achieves 10x–50x faster column scans than PostgreSQL TimescaleDB on billion-row datasets [Grade A: https://clickhouse.com/docs/benchmark].
   - Compression ratio achieves 4:1 to 6:1 using ZSTD [Grade A: primary benchmark].
2. **Session S2 Discovered Concerns & Falsification:**
   - **Critical Failure Mode:** ClickHouse degrades severely under single-row inserts or micro-batches (<1,000 rows/insert), creating massive "Too many parts" merging errors [Grade A: https://clickhouse.com/docs/en/troubleshooting/too-many-parts].
   - **Operational Overhead:** Requires dedicated buffer tier (Kafka or Vector buffer) before ClickHouse to ensure batches of 10,000+ rows.

---

## Step 4: Adversarial Stress-Testing (Red Team)

Before locking the decision, the agent subjects the recommendation to adversarial challenge:

**Tool Call:** `vivechak_challenge`
```json
{
  "decision_id": "D-01",
  "mode": "red_team"
}
```

**Tool Response:**
```json
{
  "success": true,
  "message": "Generated red-team challenge prompts for D-01",
  "data": {
    "challenge_prompts": [
      {
        "type": "premortem",
        "prompt": "Assume it is 12 months after migration and ClickHouse clusters are crashing weekly due to ZooKeeper/Keeper synchronization lag and part explosion. What specific operational guardrails did we fail to implement?"
      }
    ]
  },
  "next_step": "Run the challenge prompt to stress-test your assumptions, or proceed to vivechak_record_decision with explicit mitigations and reversal triggers."
}
```

---

## Step 5: Locking the Architectural Decision Record (ADR)

The agent formalizes the ADR with **explicit reversal triggers** and **rejected alternatives**:

**Tool Call:** `vivechak_record_decision`
```json
{
  "decision_id": "D-01",
  "content": "---\nid: D-01\ntitle: Migrate Analytics Event Storage to ClickHouse with Vector Ingestion Buffer\ndate: \"2026-10-18\"\nstatus: accepted\ndoor_type: one-way\nconfidence: \"85%\"\nreview_trigger: \"If ClickHouse operational management exceeds 5 hours/week of on-call time over a 30-day rolling window, or if ingestion lag exceeds 5 minutes for >3 incidents/month, trigger rollback evaluation.\"\n---\n\n# D-01: Migrate Analytics Event Storage to ClickHouse\n\n## Context\nPostgreSQL aggregation queries over 30-day ranges currently exceed 8 seconds, causing timeouts on dashboard metrics.\n\n## Decision\nMigrate to ClickHouse Cloud (managed) using a Vector ingestion buffer to batch writes into >=10,000 row chunks.\n\n## Rejected Alternatives\n1. **PostgreSQL TimescaleDB:** Retains single-node write saturation and high disk footprint [Grade B: benchmark].\n2. **AWS DynamoDB + Athena:** Query latency on ad-hoc scans is too high (3–15 seconds) and cost scales linearly with scan volume [Grade B: AWS documentation].\n\n## Reversal Trigger\nIf managed ClickHouse operational overhead exceeds 5 engineering hours/week or cluster maintenance costs exceed $1,200/month, replicate data back to PostgreSQL partitioning."
}
```

**Tool Response:**
```json
{
  "success": true,
  "message": "Decision D-01 recorded as canonical ADR in research/D-01-clickhouse-migration.md and compiled into consolidated research/DECISIONS.md.",
  "data": {
    "decision_id": "D-01",
    "door_type": "one-way",
    "status": "accepted"
  },
  "next_step": "Decision D-01 is locked! The canonical ADR is saved in research/D-01-clickhouse-migration.md and registry research/DECISIONS.md is up to date. You can now proceed to implement the Vector buffer and ClickHouse schema."
}
```

---

## Key Takeaways
1. **Total Time:** ~30 minutes across 2 targeted sessions.
2. **Zero Project Overhead:** No 20-session DAG, no grand Founding Architecture Document (`FAD.md`).
3. **One-Way Door Rigor:** Falsification identified the "Too many parts" disaster before coding; ADR locked with measurable reversal triggers.
4. **Permanent Traceability:** Canonical record stored in `research/D-01-clickhouse-migration.md` and indexed in `research/DECISIONS.md`.
