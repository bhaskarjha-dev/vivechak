# Sample Adoption Walkthrough: Decision Scope (PostgreSQL to ClickHouse Migration)

> **Scope:** Decision Scope (`ScopeDecision`)  
> **Investment:** 1–3 Sessions (~30 minutes)  
> **Overhead:** Zero Project DAG, Zero FAD  
> **Deliverables:** `research/D-001-decision.md` (Locked ADR) and `research/DECISIONS.md` (Registry)

This walkthrough shows an AI coding agent using Vivechak to resolve an architectural fork in an active codebase: deciding whether to migrate analytical event storage from PostgreSQL to ClickHouse.

---

## Step 1: The Architectural Dilemma

The developer prompts the agent:
> *"Vivechak decision: Should we migrate our primary event log storage from PostgreSQL to ClickHouse? We currently ingest 25M events/day, query latency on 30-day aggregations is degrading, but our operations team has zero ClickHouse experience."*

The agent calls `vivechak_start` with `scope="decision"`:

**Tool Call:** `vivechak_start`
```json
{
  "scope": "decision",
  "question": "Migrate event log storage from PostgreSQL to ClickHouse. Ingesting 25M events/day, query latency on 30-day aggregations degrading. Ops team has zero ClickHouse experience."
}
```

**Tool Response:**
```json
{
  "ok": true,
  "message": "Decision workspace ready (D-001). Synthesize the 1-3 session decision plan and save it.",
  "data": {
    "scope": "decision",
    "id": "D-001",
    "brief": "# GENERATE: Architectural Decision Plan (ADR)\n\n## DILEMMA\nMigrate event log storage from PostgreSQL to ClickHouse...\n\nRoutes to 2 Sessions: C (Comparison & Benchmark) + F (Falsification & Failure Modes)..."
  },
  "next": {
    "tool": "vivechak_save",
    "args": {
      "kind": "plan",
      "scope": "decision",
      "id": "D-001"
    },
    "why": "Synthesize the decision research plan with a yaml vivechak-plan block and save it."
  }
}
```

---

## Step 2: Saving the Decision Plan

The agent executes the generator, producing a tailored 2-session research plan: one `yaml vivechak-plan` block listing the sessions (`D-001-S1`, then `D-001-S2`, which depends on it) and, under `## Sessions`, one `prompt` fence per session. It calls `vivechak_save` (the server normalizes `D-01` to `D-001`):

**Tool Call:** `vivechak_save`
```json
{
  "kind": "plan",
  "id": "D-01",
  "content": "# D-001-PLAN: Migrate event log storage from PostgreSQL to ClickHouse\n\n```yaml vivechak-plan\nplan: D-001\nsessions:\n  - {id: D-001-S1, title: ClickHouse Ingestion Throughput & Partitioning Benchmark, role: C, door: one-way, informs: [D-001], depends: []}\n  - {id: D-001-S2, title: \"Operational Failure Modes, Small-Batch Pitfalls & Reversibility\", role: F, door: one-way, informs: [D-001], depends: [D-001-S1]}\n```\n\n## Routing\n| Decision | D-001 |\n|---|---|\n| Scope | decision |\n| Door Type | one-way |\n\n## Sessions\n\n### D-001-S1: ClickHouse Ingestion Throughput & Partitioning Benchmark\n```prompt\n# RESEARCH BRIEF: ClickHouse ingestion throughput and partitioning\nCompare ClickHouse and PostgreSQL for 25M events/day and 30-day aggregations: ingestion throughput, partitioning, compression and query latency, citing reproducible benchmarks.\n```\n\n### D-001-S2: Operational Failure Modes, Small-Batch Pitfalls & Reversibility\n```prompt\n# RESEARCH BRIEF: how a ClickHouse migration could fail\nFind ClickHouse failure modes for a team without ClickHouse experience (small-batch inserts, merges, operations), the exit path back to PostgreSQL and its cost, and reversal triggers.\n```\n"
}
```

**Tool Response:**
```json
{
  "ok": true,
  "message": "Saved decision plan D-001 to research/D-001-plan.md (2 sessions).",
  "data": {
    "id": "D-001",
    "file_path": "research/D-001-plan.md"
  },
  "next": {
    "tool": "vivechak_next",
    "args": {},
    "why": "Call vivechak_next to retrieve the first unblocked session prompt."
  }
}
```

---

## Step 3: Executing Research Sessions

The agent gets each prompt with `vivechak_next()`, executes `D-001-S1` and `D-001-S2` using web search, and saves results with `vivechak_save(kind="session", id=..., content=...)`:

1. **Session S1 Findings (`vivechak_save(kind="session", id="D-001-S1", content="<session markdown>")`):**
   - ClickHouse achieves 10x–50x faster column scans than PostgreSQL TimescaleDB on billion-row datasets [Grade A: https://clickhouse.com/docs/benchmark].
   - Compression ratio achieves 4:1 to 6:1 using ZSTD [Grade A: primary benchmark].
2. **Session S2 Discovered Concerns & Falsification (`vivechak_save(kind="session", id="D-001-S2", content="<session markdown>")`):**
   - **Critical Failure Mode:** ClickHouse degrades severely under single-row inserts or micro-batches (<1,000 rows/insert), creating massive "Too many parts" merging errors [Grade A: https://clickhouse.com/docs/en/troubleshooting/too-many-parts].
   - **Operational Overhead:** Requires dedicated buffer tier (Kafka or Vector buffer) before ClickHouse to ensure batches of 10,000+ rows.

---

## Step 4: Quality Review & Stress-Testing

Before locking the decision, the agent validates workspace artifacts and checks quality observations:

**Tool Call:** `vivechak_review`
```json
{
  "mode": "validate"
}
```

**Tool Response:**
```json
{
  "ok": true,
  "message": "Workspace validation complete (0 blocking issues).",
  "next": {
    "tool": "vivechak_next",
    "args": {},
    "why": "Continue executing planned sessions or record decisions."
  }
}
```

---

## Step 5: Locking the Architectural Decision Record (ADR)

The agent formalizes the ADR with **explicit reversal triggers** and **rejected alternatives**:

**Tool Call:** `vivechak_record_decision`
```json
{
  "id": "D-01",
  "content": "---\nid: D-01\ntitle: Migrate Analytics Event Storage to ClickHouse with Vector Ingestion Buffer\ndate: \"2026-10-18\"\nstatus: accepted\ndoor_type: one-way\nconfidence: \"85%\"\nreview_trigger: \"If ClickHouse operational management exceeds 5 hours/week of on-call time over a 30-day rolling window, or if ingestion lag exceeds 5 minutes for >3 incidents/month, trigger rollback evaluation.\"\n---\n\n# D-01: Migrate Analytics Event Storage to ClickHouse\n\n## Context\nPostgreSQL aggregation queries over 30-day ranges currently exceed 8 seconds, causing timeouts on dashboard metrics.\n\n## Decision\nMigrate to ClickHouse Cloud (managed) using a Vector ingestion buffer to batch writes into >=10,000 row chunks.\n\n## Rejected Alternatives\n1. **PostgreSQL TimescaleDB:** Retains single-node write saturation and high disk footprint [Grade B: benchmark].\n2. **AWS DynamoDB + Athena:** Query latency on ad-hoc scans is too high (3–15 seconds) and cost scales linearly with scan volume [Grade B: AWS documentation].\n\n## Reversal Trigger\nIf managed ClickHouse operational overhead exceeds 5 engineering hours/week or cluster maintenance costs exceed $1,200/month, replicate data back to PostgreSQL partitioning."
}
```

**Tool Response:**
```json
{
  "ok": true,
  "message": "Saved decision D-001 as valid",
  "data": {
    "id": "D-001",
    "file_path": "research/D-001-decision.md",
    "decisions_file": "research/DECISIONS.md",
    "status": "valid"
  },
  "next": {
    "tool": "vivechak_review",
    "args": {
      "mode": "red_team",
      "id": "D-001"
    },
    "why": "Run adversarial red-team review for One-Way Door D-001 and cite the returned record_id before Phase 0 Gate."
  }
}
```

---

## Key Takeaways
1. **Total Time:** ~30 minutes across 2 targeted sessions.
2. **Zero Project Overhead:** No 20-session DAG, no grand Founding Architecture Document (`FAD.md`).
3. **One-Way Door Rigor:** Falsification identified the "Too many parts" disaster before coding; ADR locked with measurable reversal triggers.
4. **Permanent Traceability:** Canonical record stored in `research/D-001-decision.md` and indexed in `research/DECISIONS.md`.
