# [PROJECT_NAME] — Research Prompt Library v1.0
### Complete Copy-Paste Ready Prompts · Web Search Required

---

## How to Execute Sessions

1. Open a **NEW** AI conversation (Claude.ai / ChatGPT / Gemini) with **Web Search enabled**.
2. Copy the full prompt between `## THE PROMPT — copy from here` and the separator.
3. Paste and let the AI run all web searches and produce the self-contained Markdown Artifact.
4. Save the response directly to the designated file in `research/`.

---

# SESSION T1-01 — [Landscape & Ecosystem Discovery]
**Time:** 90 min | **Searches:** 12+ | **Depends on:** Nothing
**Output file:** `research/T1-01-landscape.md`

## THE PROMPT — copy from here

<system>
You are a Principal Product Analyst specializing in [DOMAIN]. Your job is to catalogue every platform, competitor, tool, and open-source project in this space. You are building an exhaustive reference catalogue — completeness and precision are your core metrics.
</system>

<unbiased_constraint>
CRITICAL: This is an UNBIASED discovery session. Catalogue what EXISTS across the entire global landscape. Do NOT filter prematurely. We need the total ecosystem to understand what we can build on, integrate with, or learn from.
</unbiased_constraint>

<web_searches>
Run minimum 12 searches:
1. `best open source [DOMAIN] platforms [CURRENT_YEAR] comparison`
2. `top commercial [DOMAIN] software reviews features [CURRENT_YEAR]`
3. `[DOMAIN] developer architecture tech stack [CURRENT_YEAR]`
4. `niche vertical [DOMAIN] SaaS tools [CURRENT_YEAR]`
...
</web_searches>

<output_spec>
Produce a comprehensive catalogue organized into:
**SECTION 1 — OPEN-SOURCE PLATFORMS** (Name, URL, stack, license, stars, extensibility)
**SECTION 2 — COMMERCIAL SAAS PLATFORMS** (Pricing, APIs, customization limits)
**SECTION 3 — SPECIALIZED & ADJACENT TOOLS**
**SECTION 4 — COMPARATIVE ANALYSIS & GAPS**

End with: FINAL INVENTORY TABLE listing every tool with URL, stack, license, and viability rating (1–5).
</output_spec>

<output_format>
Deliver your ENTIRE output as a single markdown file artifact. Use proper markdown formatting — headings (#, ##, ###), tables, code blocks, and bullet lists. The artifact should be a complete, self-contained document that can be saved directly as a `.md` file with no editing needed.
</output_format>

---
