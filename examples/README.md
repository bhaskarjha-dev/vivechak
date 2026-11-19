# Vivechak Adoption Examples & Dogfood Workspace

Concrete end-to-end walkthroughs demonstrating how AI coding agents use Vivechak across all three scope levels — replayed and schema-checked in CI by `TestDocsFlows_ReplaysExampleWalkthroughs` (`cmd/vivechak/docs_flows_test.go`).

---

## 1. Walkthroughs by Scope

| Scope | Walkthrough | Scenario | Sessions & Overhead | Output Artifacts |
|---|---|---|---|---|
| **1. Comparison** | [**`SAMPLE-COMPARISON.md`**](SAMPLE-COMPARISON.md) | Evaluating **Pino vs. Winston** for high-throughput Node.js JSON logging | 1 session (~5 min) · Zero DAG · Zero Gate | `research/C-001-comparison.md` (Weighted Evaluation Matrix + sensitivity check) |
| **2. Decision** | [**`SAMPLE-DECISION.md`**](SAMPLE-DECISION.md) | Deciding whether to migrate analytical event logs from **PostgreSQL to ClickHouse** at 25M events/day | 2 sessions (`C` + `F`, ~30 min) · Track B Gate | `research/D-001-plan.md`, `research/D-001-decision.md`, `research/D-001-R1.md`, `research/DECISIONS.md` |
| **3. Project** | [**`SAMPLE-PROJECT.md`**](SAMPLE-PROJECT.md) | Architecting **Katha**, an offline-first collaborative storytelling platform from scratch | Multi-layer DAG (`T1-*` → `SYN-01`) · Phase 0 Gate | `research/RESEARCH-PIPELINE.md`, `research/sessions/*.md`, `research/FAD.md`, `research/PHASE-0-GATE.md` |

---

## 2. Live Dogfood Workspace (`examples/dogfood/`)

[`dogfood/research/`](dogfood/research/) contains real architectural decision records (`D-001` through `D-003`) for Vivechak itself, audited continuously in CI via `vck doctor` and `vck gate`:

```sh
# Inspect workspace health and run the headless Phase 0 Gate on a scratch copy:
cp -R examples/dogfood /tmp/vivechak-dogfood
vck doctor /tmp/vivechak-dogfood
vck gate --track all /tmp/vivechak-dogfood
```
