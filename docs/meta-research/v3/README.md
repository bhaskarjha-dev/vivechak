# v3 Research — Prompt Evolution, Hybrid Workflow & Production Orchestration Evidence Base

> **Sealed:** 2026-10-01. These files are historical records.

## Contents

| File | Purpose |
|---|---|
| [`D-UX-01-cli-design.md`](D-UX-01-cli-design.md) | CLI subcommand architecture, interactive vs headless modes, command taxonomy (informs D-015) |
| [`D-UX-02-validation-depth.md`](D-UX-02-validation-depth.md) | Research depth validation, advisory vs blocking enforcement, avoiding composite score gamification (informs D-016) |
| [`D-UX-03-hybrid-workflow.md`](D-UX-03-hybrid-workflow.md) | Hybrid MCP/Manual workflow, clipboard injection, Manual Session Card format, and HITL patterns |
| [`production-research-claude.md`](production-research-claude.md) | Deep research into production research agents (STORM, TTD-DR, CRAG, DREAM), 8-block prompt anatomy, Prior/Delta tracking |
| [`production-research-gemini.md`](production-research-gemini.md) | Production AI research orchestration framework, examiner questions, validation ladders, field guide |
| [`production-research-orchestration-design.md`](production-research-orchestration-design.md) | Orchestration architecture, Claim-Evidence Record schema, Directed Epistemic Graph model |

## Context

This research informed Vivechak v0.1.0:
1. **Prompt Architecture Evolution:** Expanding from 5 to 8 blocks (`DECISION`, `KNOWN`, `CALIBRATION`, `DONE`), introducing dynamic moves vocabulary (`DEEPEN`/`WIDEN`/`CORROBORATE`/`FALSIFY`/`PIVOT`), iterative research self-correction, and `Prior`/`Delta` epistemic tracking.
2. **Advisory Validation Philosophy:** Transitioning from rigid blocking enforcement to enablement and advisory quality observations, grounded in empirical evidence against composite scoring and gaming.
3. **Hybrid Execution Patterns:** Supporting both headless MCP agent loops and interactive manual/browser research sessions via the Manual Session Card pattern ([`HYBRID-EXECUTION.md`](../../HYBRID-EXECUTION.md)).
4. **Scope-Aware Orchestration:** Calibrating template sets and verification rigor dynamically to the selected scope (Project, Decision, Comparison).

These documents represent foundational synthesis and comparative analysis across state-of-the-art production research systems (Claude Deep Research, OpenAI Deep Research, Gemini Deep Research, academic frameworks) and user-experience architectural tradeoffs.
