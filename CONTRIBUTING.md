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

---

## Reporting Issues

- **Bug Reports:** Use the [Bug Report Template](.github/ISSUE_TEMPLATE/bug_report.md).
- **Feature / Spec Proposals:** Use the [Feature Request Template](.github/ISSUE_TEMPLATE/feature_request.md).

Thank you for helping build sovereign, evidence-grounded tools!
