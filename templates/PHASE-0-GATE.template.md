# Phase 0 Gate: Research → Codebase Transition Protocol

> **Rule:** Do NOT write a single line of application code until every box below is checked.

---

### Step 1: Research Artifact Verification
- [ ] All Tier 1 Markdown Artifacts exist in `research/T1-*.md`
- [ ] All Tier 2 Decision Artifacts exist in `research/T2-*.md`
- [ ] All Tier 3 Blueprint Artifacts exist in `research/T3-*.md`
- [ ] `SYN-01` Grand Synthesis exists in `research/SYN-01-founding-architecture.md`

### Step 2: Documentation Harmonization
- [ ] `docs/ARCHITECTURE.md` updated with pinned tech stack from FAD.
- [ ] `DECISIONS.md` updated: all `D-001` through `D-NNN` marked `RESOLVED`.
- [ ] `docs/PRODUCT.md` updated to reflect any assumptions refuted by research.
- [ ] `PRINCIPLES.md` initialized with builder's constitution.

### Step 3: Scaffold & First Commit
- [ ] Repository initialized with chosen monorepo / framework tool.
- [ ] Linter, formatter (Biome/ESLint), and TypeScript strict mode configured.
- [ ] Initial build and test commands execute cleanly on empty scaffold (`pnpm build`, `pnpm check`).
- [ ] Initial Git commit created:  
  `git commit -m "feat: initial project scaffold based on Founding Architecture Document"`
