## Summary of Changes

<!-- Brief 1-3 sentence summary of what this PR introduces or fixes. -->

## Type of Change

- [ ] MCP Server / CLI (`cmd/vivechak/`, `internal/`)
- [ ] Core Framework update (`FRAMEWORK.md`, `internal/method/`)
- [ ] Generator Prompt update (`GENERATOR-PROJECT.md`, `GENERATOR-DECISION.md`, `GENERATOR-COMPARISON.md`)
- [ ] Template / Skill / Rule contract update (`templates/`, `skills/`)
- [ ] Packaging / Distribution / CI (`.goreleaser.yml`, `Dockerfile`, `npm/`, `pypi/`, `winget/`, `flake.nix`)
- [ ] Documentation / Example update
- [ ] Bug fix

## Evidence & Rationale

<!--
If modifying core principles, prompt anatomy, scaling tiers, or gating logic:
Cite the empirical evidence, benchmark, paper, or trial results justifying the change.
-->

## Verification Checklist

- [ ] `go test -race ./...` passes cleanly.
- [ ] `make embed-sync` and `go run ./tools/docgen` produce zero uncommitted drift.
- [ ] All Markdown links resolve correctly.
- [ ] Operational templates in `templates/` remain valid (`schema_version: "0.1.0"`).
- [ ] Anti-Bias Guard: No uncalibrated dogma or uncorroborated assumptions introduced.
