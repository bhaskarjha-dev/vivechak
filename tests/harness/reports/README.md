# Vivechak Release Audit Reports
### Permanent & Immutable Multi-Harness Empirical Baselines

This directory contains versioned milestone audit reports compiled at major release points. Each audit report freezes empirical performance, tool compliance, pass rates, and model reflection data across heterogeneous agent runtimes.

## Release Audit Catalog

| Milestone Version | Audit Report Date | Scenarios Evaluated | Harnesses Tested | Pass Rate | Audit Document / Notes |
|---|---|---|---|---|---|
| **v0.1.0** | 2026-10-19 | 12 Scenarios (Phase 1 Baseline, 25 Runs) | Google Antigravity (`agy`), OpenCode (`opencode`), Direct MCP | **100.0%** (25/25) | [audit-v0.1.0-2026-10-19.md](audit-v0.1.0-2026-10-19.md) |
| **v0.1.0 (Full)** | 2026-10-19 | 21 Scenarios (Run 27 Milestone, 64 Runs) | Google Antigravity (`agy`), OpenCode (`opencode`), Direct MCP | **100.0%** (21/21 Deterministic Direct MCP) / **81.3%** Aggregate | Documented in [`DEFECTS.md`](../defects/DEFECTS.md#run-27-full-21-scenario-autonomous-multi-harness-campaign-release-v010-audit) |

## Archiving New Release Audits
When a development/hardening cycle completes:
```bash
node scripts/run_harness.mjs --archive <version>
# Example: node scripts/run_harness.mjs --archive v0.1.1
```
This automatically compiles active runs in `tests/harness/runs/` into an immutable audit report in this directory, updates this catalog index, and clears the local testbed for the next iteration.
