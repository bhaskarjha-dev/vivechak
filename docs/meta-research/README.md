# Meta-Research — Vivechak Design Evidence

This directory contains the architectural decision records and pipeline definitions
that justify every design choice in Vivechak. It is the **audit trail** — proof that
the framework practices what it preaches.

**You do NOT need to read any of this to use Vivechak.** Start with the [README](../../README.md).

---

## Contents

| File | What |
|---|---|
| [`DECISIONS.md`](DECISIONS.md) | All 14 locked architectural decisions (D-001–D-014) in Vivechak ADR format |
| [`RESEARCH-PIPELINE-v1.md`](RESEARCH-PIPELINE-v1.md) | Foundational meta-research pipeline — methodology genesis (D-001–D-010) |
| [`v2/FINAL-PLAN.md`](v2/FINAL-PLAN.md) | The definitive implementation plan for v0.1.0 — 5 phases, risk register, kill criteria |
| [`v2/RESEARCH-PIPELINE-v2.md`](v2/RESEARCH-PIPELINE-v2.md) | MCP server & multi-scope research pipeline (D-011–D-014) |
| [`v2/README.md`](v2/README.md) | MCP server & multi-scope research context and methodology notes |

## Raw Research Sessions

The original research sessions (~1.4MB, 27 files) that produced these decisions
were removed from the working tree for v1.0.0. They are preserved in git history:

```sh
# v1.0 sessions (15 files, 642KB)
git log --all --diff-filter=D -- 'meta-research/sessions/'

# v2.0 sessions (12 files, 751KB)
git log --all --diff-filter=D -- 'meta-research/v2-research/with-repo-link/'
git log --all --diff-filter=D -- 'meta-research/v2-research/without-repo-link/'
```

## Why This Exists

Vivechak's Principle P3 (Evidentiary Grounding) requires every claim to carry an
evidence grade. This directory is the framework eating its own dogfood — applying
the same standard to its own design decisions.
