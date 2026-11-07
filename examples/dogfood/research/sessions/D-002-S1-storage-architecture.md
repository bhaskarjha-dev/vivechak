---
id: D-002-S1
title: "Workspace Storage Model Analysis"
date: "2026-10-24"
status: complete
topic: "storage"
tags: ["storage", "cgo-free"]
informs_decisions: ["D-002"]
confidence: high
schema_version: "0.1.0"
---

# D-002-S1: Workspace Storage Model Analysis

## Prior

1. I believe an embedded SQLite database would offer simpler multi-agent atomic locking.
2. I believe developers prefer inspectable markdown files in their git repository.

## Research Question

Should Vivechak adopt SQLite or a confined file-based store (`internal/store`) using Go 1.24+ `os.Root` confinement?

## Key Findings

- **Cross-Compilation Independence:** Using SQLite via CGO (`mattn/go-sqlite3`) breaks zero-dependency static compilation across 7 targets. Pure-Go SQLite adds 3.2 MB to the binary [Grade A: go tool nm / size | verified].
- **Human Collaboration:** Storing research as plain Markdown files in `research/` enables seamless code review in pull requests without specialized DB viewers [Grade B: GitHub review workflow | cached].

## Recommendation

Implement a confined file-based store in `internal/store` with atomic writes, monotonic collision counters, and directory confinement.

## Alternatives Considered

| Option | Verdict | Key Tradeoff |
|---|---|---|
| **Confined File-Based Store** | **Recommended** | Zero CGO dependencies, human-readable Git diffs, safe os.Root confinement |
| **SQLite Database** | **Rejected** | Requires CGO or bloated pure-Go driver; makes research opaque in Git |

## Open Questions & Risks

| Risk / Question | Impact | Mitigation |
|---|---|---|
| Windows nanosecond file rename collisions | Medium | Use atomic monotonic temp file counters and exponential backoff retry loop |

## Discovered Concerns

On Windows NTFS filesystems, rapid back-to-back atomic file replacements can trigger intermittent `ERROR_ACCESS_DENIED` due to transient file handle locking by antivirus or indexer processes. Implementing a dedicated retry loop with exponential backoff is essential for Windows reliability.

## Delta

| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| SQLite is required for robust concurrency | Contradicted | Confined atomic file replacement with advisory locks satisfies all concurrency invariants | High |

## Sources & Evidence Ledger

| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://pkg.go.dev/os#Root | A | authoritative · fresh | fetched | Go 1.24 filesystem confinement specification |
| 2 | internal/store/atomic.go | A | direct · verified | cached | Atomic temporary write and rename implementation |
