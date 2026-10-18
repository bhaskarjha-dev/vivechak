# Sample Adoption Walkthrough: Comparison Scope (Pino vs Winston)

> **Scope:** Comparison Scope (`ScopeComparison`)  
> **Investment:** 1 Session (~5 minutes)  
> **Overhead:** Zero DAG, Zero Phase 0 Gate  
> **Deliverable:** `research/C-01-comparison.md` (Weighted Evaluation Protocol Matrix)

This walkthrough shows an AI coding agent using Vivechak's Model Context Protocol (MCP) server to execute a fast, rigorous options comparison between two Node.js logging libraries.

---

## Step 1: Prompt & Generator Preparation

The user prompts their coding agent in chat:
> *"Vivechak compare: Pino vs Winston for our high-throughput Express.js microservice handling 15k req/sec."*

The agent calls `vivechak_prepare_generator` with `scope="comparison"`:

**Tool Call:** `vivechak_prepare_generator`
```json
{
  "scope": "comparison",
  "context": "Pino vs Winston for high-throughput Express.js microservice handling 15k req/sec."
}
```

**Tool Response:**
```json
{
  "success": true,
  "message": "Comparison generator prompt prepared.",
  "data": {
    "scope": "comparison",
    "prompt": "# GENERATE: Bounded Options Comparison\n\n## OPTIONS\n- Pino\n- Winston\n\n## CRITERIA\n- Throughput & Latency overhead\n- Structured JSON streaming performance\n- Ecosystem & transport plugin reliability\n...",
    "template_file": "COMPARISON-SESSION.template.md"
  },
  "next_step": "Execute the comparison research prompt using frontier LLM capabilities with web search, then call vivechak_save_plan with the generated Weighted Evaluation Protocol (WEP) matrix."
}
```

---

## Step 2: Research Execution

The agent executes the generator prompt using live web search, benchmarking data, and official documentation to populate the **Weighted Evaluation Protocol (WEP)** matrix with evidence grades:

- **Throughput:** Pino uses asynchronous worker thread logging (`pino/file`) reducing event loop blockage to sub-millisecond overhead [Grade A: https://getpino.io/docs/benchmarks/]. Winston formats synchronously on the main thread [Grade B: reproducible benchmark].
- **Memory footprint:** Pino allocates minimal objects per log call; Winston instantiates multiple format transforms per write [Grade A: source inspect].
- **Ecosystem:** Winston has extensive legacy transports (CloudWatch, Datadog); Pino uses decoupled pipeline transports (`pino-pretty`, `pino-loki`) outside the app process.

---

## Step 3: Persisting the Comparison Plan

The agent saves the completed comparison artifact directly via `vivechak_save_plan`. The server automatically handles on-the-fly workspace initialization if needed, copies only `COMPARISON-SESSION.template.md` (zero bloat), and bypasses DAG cycle constraints:

**Tool Call:** `vivechak_save_plan`
```json
{
  "scope": "comparison",
  "decision_id": "C-01",
  "content": "---\nid: C-01-cmp-pino-vs-winston\ntitle: Pino vs Winston for High-Throughput Node.js Microservice\ndate: \"2026-10-18\"\nstatus: complete\n---\n\n# C-01: Pino vs Winston Comparison\n\n## Weighted Evaluation Protocol (WEP) Matrix\n| Criterion | Weight | Pino Score (1-5) | Winston Score (1-5) | Evidence Grade & Provenance |\n|---|---|---|---|---|\n| Event Loop Latency | 0.40 | 5 (0.3ms/log) | 3 (1.8ms/log) | [Grade A: https://getpino.io/docs/benchmarks/] |\n| Memory Allocation | 0.30 | 5 (Minimal GC) | 3 (Object churn) | [Grade B: V8 heap profiler] |\n| Transport Decoupling | 0.30 | 5 (Worker thread) | 4 (In-process streams) | [Grade A: Pino transport docs] |\n| **Weighted Total** | **1.00** | **5.00** | **3.30** | **Pino Wins (+51% margin)** |\n\n## Sensitivity Check\nEven if transport weight increases to 0.50, Pino retains the winning score (4.70 vs 3.60).\n\n## Failure Modes\n- **Pino:** In unbuffered crashes, in-flight worker logs may be lost unless `pino.final()` handler is registered [Grade A: docs].\n- **Winston:** High volume synchronously saturates Node.js libuv event loop threads, creating backpressure lag on HTTP routes [Grade B: issue reports].\n\n## Recommendation\nAdopt **Pino** with externalized log shipping via worker thread. Register `pino.final()` on `SIGTERM`."
}
```

**Tool Response:**
```json
{
  "success": true,
  "message": "Comparison C-01 saved successfully in research/C-01-cmp-pino-vs-winston.md",
  "data": {
    "decision_id": "C-01",
    "scope": "comparison",
    "path": "research/C-01-cmp-pino-vs-winston.md"
  },
  "next_step": "Comparison complete! The Weighted Evaluation Protocol matrix is saved in research/C-01-cmp-pino-vs-winston.md. Zero gate or DAG overhead required. You can now proceed to implementation."
}
```

---

## Step 4: Optional Visualization

To inspect or present the decision to human reviewers, the agent or user can call `vivechak_visualize`:

**Tool Call:** `vivechak_visualize`
```json
{
  "format": "mermaid"
}
```

**Tool Response:**
```json
{
  "success": true,
  "message": "Rendered comparison visualization",
  "data": {
    "format": "mermaid",
    "rendering": "flowchart TD\n    subgraph C01[\"C-01: Pino vs Winston\"]\n        A[\"Pino: Score 5.00 ★ Winner\"]:::winner\n        B[\"Winston: Score 3.30\"]:::alternative\n    end\n    classDef winner fill:#22c55e,stroke:#15803d,color:#ffffff;\n    classDef alternative fill:#6b7280,stroke:#374151,color:#ffffff;"
  }
}
```

---

## Key Takeaways
1. **Total Time:** ~5 minutes.
2. **Zero DAG Bloat:** No `RESEARCH-PIPELINE.md`, no multi-tier execution tree.
3. **Zero Gate Overhead:** Comparison Scope is exempt from the pre-codebase exit gate.
4. **Epistemic Traceability:** The choice of Pino is grounded in official benchmark URLs (Grade A) and failure mode analysis rather than subjective chat recall.
