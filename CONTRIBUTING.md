# Contributing to Vivechak (विवेचक)

Thank you for your interest in improving Vivechak! We welcome contributions that maintain the rigor, empirical grounding, and philosophical integrity of the framework.

---

## The Golden Rule: Evidence-Grounded Evolution

Vivechak was created by applying its own methodology to itself — 11 meta-research sessions tested 10 founding hypotheses before v1.0 was written.

**Any proposed change to the core methodology (`FRAMEWORK.md`, `GENERATOR.md`, or `templates/`) must satisfy the Anti-Bias & Evidence Threshold:**

1. **Cite Grounded Evidence:** Modifications to prompt structure, evidence grading, scaling tiers, or gating criteria must cite empirical benchmarks, peer-reviewed literature, or documented research trajectories.
2. **No Uncalibrated Dogma:** Avoid introducing arbitrary rules (e.g. "always run 5 sessions" or "always use 3 models"). Rules must scale with risk (One-Way vs Two-Way doors).
3. **Preserve Provable Contracts:** Templates in `templates/` are machine-parseable contracts used by autonomous AI agents (Antigravity, Cursor, Claude Code). Do not break YAML frontmatter schemas without updating `schema_version`.

---

## Development & Contribution Workflow

1. **Fork & Branch:** Create a feature branch from `main` (e.g., `feat/dspy-generator-optimization` or `fix/template-frontmatter`).
2. **Check for Self-Sufficiency:** Ensure generated files remain self-contained. A project initialized with Vivechak templates should never need runtime dependencies on this repository.
3. **Verify Links & Fences:** Ensure all Markdown links resolve and code fences close properly.
4. **Submit a Pull Request:** Fill out the [Pull Request Template](.github/PULL_REQUEST_TEMPLATE.md) completely, including the Evidence Check.

---

## Release Checklist

When releasing a new version, update ALL of the following references:

- [ ] `VERSION` — bump the version number
- [ ] `CHANGELOG.md` — add new release entry
- [ ] `README.md` — header version, tree diagram version, roadmap table
- [ ] `AGENTS.md` — header version
- [ ] `GENERATOR.md` — title version, Design Notes version
- [ ] `FRAMEWORK.md` — header version (if applicable)
- [ ] `templates/` — header version in all 4 templates, `schema_version` in DECISIONS template if schema changed

---

## Change Propagation Map

**Why this matters:** FRAMEWORK.md is the specification. GENERATOR.md is the implementation (the prompt that produces actual output). Templates are the contracts users fill out. Changing the spec without updating the implementation means improvements exist only in documentation but never affect generated pipelines. This is the Vivechak equivalent of updating a language spec without updating the compiler.

**Rule: When you modify any file listed below, you MUST check every file in its propagation set.** Not every change will require propagation, but every change MUST be checked.

### FRAMEWORK.md → check these files

| If you changed... | Check GENERATOR.md | Check templates/ | Check README.md |
|---|---|---|---|
| Core Principles (§2) | ✅ If the principle affects how pipelines are generated | — | ✅ If it changes user-facing behavior |
| Pipeline Architecture (§3) — scaling, tiers, session budgets | ✅ Scoring rubric and tier mapping are duplicated in GENERATOR.md Step 2 | — | — |
| Pipeline Architecture (§3) — failure modes, termination, checkpoints | ✅ "How to Execute" section must reflect new guidance | — | — |
| Prompt Design (§4) — 5-block anatomy, output rubric | ✅ Step 4 writes prompts using this anatomy | — | — |
| Evidence System (§5) — grading, modifiers, citation format | ✅ Evidence grading is inlined in every generated prompt | — | — |
| Output Architecture (§6) — YAML frontmatter, body skeleton | ✅ FORMAT block in generated prompts | ✅ Template schemas must match | — |
| Phase 0 Gate (§7) | — | ✅ PHASE-0-GATE template | — |
| Generator Architecture (§8) — archetypes, domain adaptation | ✅ Step 1 archetype classification | — | — |

### templates/ → check these files

| If you changed... | Check GENERATOR.md | Check FRAMEWORK.md |
|---|---|---|
| DECISIONS.template.md — YAML schema fields | ✅ DECISIONS.md generation spec (DELIVERABLE §2) must list new fields | ✅ §5.7 ADR section must match |
| DECISIONS.template.md — body sections | ✅ If new sections should appear in generated DECISIONS.md | — |
| FOUNDING-ARCHITECTURE.template.md | — | ✅ §6.3 Map-Reduce workflow must match |
| PHASE-0-GATE.template.md | — | ✅ §7 Gate section must match |
| CONFLICT-RESOLUTION.template.md | — | — (self-contained) |

### GENERATOR.md → check these files

| If you changed... | Check FRAMEWORK.md | Check templates/ |
|---|---|---|
| Complexity scoring (Step 2) | ✅ §3.2 Scaling Model must match | — |
| Session matrix / routing (Step 3) | ✅ §3.1 Topology must match | — |
| Prompt anatomy (Step 4) | ✅ §4 Prompt Design must match | — |
| DECISIONS.md generation spec | — | ✅ DECISIONS.template.md schema must match |
| "How to Execute" references | — | ✅ Template filenames must match |

### Quick Reference: The Duplication Points

These elements exist in MORE THAN ONE file and MUST stay synchronized:

| Element | Lives In | Also Referenced In |
|---|---|---|
| 8-dimension complexity rubric | FRAMEWORK.md §3.2 | GENERATOR.md Step 2 (intentionally abbreviated labels for prompt economy — substantively identical) |
| Tier mapping (score → session budget) | FRAMEWORK.md §3.2 | GENERATOR.md Step 2 |
| Per-decision routing matrix | FRAMEWORK.md §3.2 | GENERATOR.md Step 3 |
| 5-block prompt anatomy | FRAMEWORK.md §4 | GENERATOR.md Step 4 |
| Evidence grading tiers (A-E) | FRAMEWORK.md §5.1 | GENERATOR.md Step 4 (inlined in every prompt) |
| ADR YAML schema (18 fields) | templates/DECISIONS.template.md (canonical) | GENERATOR.md DELIVERABLE §2 (full schema inlined), FRAMEWORK.md §5.7 (example) |
| Template filenames (4 templates) | templates/ directory | GENERATOR.md "How to Execute", README.md Step 2, AGENTS.md §2 |

## Reporting Issues

- **Bug Reports:** Use the [Bug Report Template](.github/ISSUE_TEMPLATE/bug_report.md).
- **Feature / Spec Proposals:** Use the [Feature Request Template](.github/ISSUE_TEMPLATE/feature_request.md).

Thank you for helping build sovereign, evidence-grounded tools!
