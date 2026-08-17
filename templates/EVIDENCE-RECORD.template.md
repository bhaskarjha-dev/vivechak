# Evidence Record Template — URP v3.0
### Standalone E-NNN Record with Blast-Radius Tracking

> **Usage:** Create one file per evidence node: `evidence/E-NNN-[slug].md`
> Evidence records are reusable across projects and decisions.

---

```yaml
---
id: E-NNN
title: "[Descriptive title of the evidence]"
grade: B                       # A | B | C | D | E
modifiers:
  corroboration: single        # single | corroborated | contested
  recency: fresh               # fresh | aging | stale
  directness: direct           # direct | indirect
verification:
  method: fetched              # fetched | cached | recalled | secondhand | human-provided
  fetched_date: YYYY-MM-DD
  fetched_by: "[agent-id or human]"
source:
  url: "[URL]"
  type: "[official-docs | engineering-postmortem | peer-reviewed | vendor-whitepaper | blog | forum]"
  author: "[Author or Organization]"
tags: []
used_by: []                    # D-NNN IDs that depend on this evidence
---
```

# E-NNN: [Evidence Title]

## Claim Summary

[1–3 sentences stating the specific factual claim this evidence supports.
Be precise about what was measured, under what conditions, and what the
quantitative result was.]

## Corroborating Sources

1. [Source 1 with URL] (Grade [X] · [directness])
2. [Source 2 with URL] (Grade [X] · [directness])

## Methodological Notes & Caveats

[Describe the conditions under which this evidence was produced. Include
hardware specs, workload characteristics, sample sizes, and known
limitations. Flag any conditions that could invalidate the claim for
your specific use case.]
