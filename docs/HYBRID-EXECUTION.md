# Hybrid Execution Guide

> **Native mode.** Vivechak assumes hybrid execution — some sessions run
> through an MCP agent, some through browser-based AI tools, some by pasting
> prompts manually. This guide explains how to move between modes seamlessly.

---

## 1. Hybrid Is the Native Mode

There is no "pure automated" or "pure manual" path in Vivechak. Every real
pipeline is hybrid:

- **MCP-automated sessions** run through `vivechak_next_session` →
  agent execution → `vivechak_save_session`. The server handles context
  injection, DAG resolution, and validation.
- **Manual sessions** run by pasting the prompt into a browser AI tool
  (Gemini Deep Research, ChatGPT with browsing, Claude with web search),
  copying the output, and saving the file to `research/sessions/`.
- **Assisted sessions** are a mix: the agent retrieves the prompt with
  `vivechak_next_session`, the user runs it manually, then tells the
  agent to save the result.

All three produce the same artifact format. The DAG doesn't care how a
session was executed — only that the output file exists with valid
frontmatter.

---

## 2. When to Run Manually

Manual execution is recommended when any of these conditions apply:

### Cross-Vendor Triangulation (P7)

For contested one-way door decisions, running the same prompt on a
second model provides independent verification. If your MCP agent uses
Claude, run the comparison session manually on Gemini or ChatGPT (or
vice versa). This is the strongest form of triangulation available
without specialized infrastructure.

### Authenticated or Gated Sources

Some sources require login or API access that your MCP agent may not
have: internal wikis, paid databases, vendor consoles, private
repositories. Run the session manually where you can authenticate, then
save the output.

### Deep Research Mode

Several AI providers offer "deep research" modes that perform extended
multi-step investigations (Gemini Deep Research, ChatGPT's research
mode). These modes often produce higher-quality outputs for complex
one-way door sessions than standard chat. They're typically only
available through browser interfaces.

### Agent Capability Gaps

If your MCP agent's harness doesn't support web search, or returns
summaries without URLs, manual execution lets you use a tool that does.
The prompt works identically either way — Vivechak's prompts are
harness-agnostic by design (P4).

### Preference or Comfort

Sometimes you want to watch the research unfold and steer it
interactively. That's a valid reason. There's no penalty for manual
execution.

---

## 3. How to Hand Off: The Manual Session Card

When the MCP agent encounters a session that should run manually, it
should present a **Manual Session Card** with these elements:

```
┌─────────────────────────────────────────────────┐
│ MANUAL SESSION CARD                             │
├─────────────────────────────────────────────────┤
│ Session:    T1-03 — Authentication Architecture │
│ Reason:     Cross-vendor triangulation (P7)     │
│ Prompt:     [full prompt text or file path]      │
│                                                 │
│ Return path:                                    │
│   Save output to:                               │
│     research/sessions/T1-03-auth-architecture.md│
│   Then tell the agent: "T1-03 is complete"      │
│                                                 │
│ Decline path:                                   │
│   Tell the agent: "Run T1-03 yourself"          │
│   The agent will execute it via MCP instead.    │
└─────────────────────────────────────────────────┘
```

### Card Elements

| Element | Purpose |
|---|---|
| **Session** | ID and title so the user knows what to run |
| **Reason** | Why manual is recommended (one of the triggers above) |
| **Prompt** | The complete prompt text, or a path to where it's stored |
| **Return path** | Exactly where to save the output and what to tell the agent |
| **Decline path** | How to skip manual execution and let the agent handle it |

The decline path is important: manual execution is always a
recommendation, never a requirement. The user can always say "run it
yourself" and the agent should proceed with automated execution.

---

## 4. How to Return: Completing a Manual Session

After running a session manually:

### Option A: Save the File Directly

1. Copy the AI's output
2. Add YAML frontmatter if the AI didn't include it:
   ```yaml
   ---
   session_id: T1-03
   title: Authentication Architecture Research
   date: 2026-10-01
   status: complete
   ---
   ```
3. Save to `research/sessions/T1-03-auth-architecture.md`
   (use the filename from the session metadata)
4. Tell the agent: "T1-03 is done" or "I've completed T1-03"

### Option B: Let the Agent Save It

1. Copy the AI's output
2. Tell the agent: "Save this as T1-03" and paste the content
3. The agent calls `vivechak_save_session` with the content,
   which handles frontmatter construction and validation

### Option C: Hybrid Save

1. Save the raw output to the sessions directory
2. Ask the agent to validate it: "Validate T1-03"
3. The agent calls `vivechak_validate` and reports any observations
4. Address any structural issues, or proceed as-is

### What Happens Next

Once the session file exists, the DAG automatically unblocks any
downstream sessions that depended on it. The next call to
`vivechak_next_session` will pick up the completed session and
return the next actionable one with upstream context injected.

---

## 5. Provenance and Evidence Handling

### Externally-Run Claims

When a session is run manually through a different AI tool, the
findings carry a different provenance than agent-executed sessions:

- Claims from the manual session are **secondhand** from the MCP
  agent's perspective — the agent didn't perform the research itself
- For informational sessions and two-way door decisions, this is
  fine — the evidence grade and verification method in the output
  document the provenance
- For critical one-way door decisions, consider having the agent
  independently verify key claims from the manual session by checking
  the cited sources

### Provenance Markers

The session output's frontmatter and evidence grades already encode
provenance through the verification field:

| Verification | Meaning |
|---|---|
| `fetched` | The executing model retrieved the source during research |
| `cached` | The executing model had a recent cached copy |
| `recalled` | The executing model used parametric memory (capped at Grade D) |
| `secondhand` | Claim comes from another session or external source |
| `human-provided` | Information supplied directly by the user |

When the MCP agent processes findings from a manually-run session for
context injection into downstream sessions, the inherited claims
naturally carry appropriate provenance. The KNOWN block in downstream
prompts tags them as `INHERITED`, signaling that they should be
stress-tested rather than accepted at face value.

### When to Verify Manually-Run Findings

For most sessions, no additional verification is needed. The evidence
grades in the output document the quality of each claim. Consider
independent verification when:

- A one-way door decision rests primarily on findings from a single
  manually-run session
- The manual session's output lacks evidence grades or sources
- Key claims are marked as recalled or secondhand
- The findings conflict with other sessions' results (use the
  Conflict Resolution template)

---

## 6. Common Patterns

### Pattern: Agent + One Manual Deep Research

The most common hybrid pattern:

1. Agent runs all Layer 0 (landscape) sessions via MCP
2. User runs one or two critical Layer 1 sessions manually using
   a deep research tool for maximum depth
3. Agent runs remaining Layer 1 and Layer 2 sessions with the
   manually-produced findings injected as upstream context
4. Agent runs synthesis (SYN-01)

### Pattern: Agent Plans, User Executes

For users who prefer hands-on research:

1. Agent calls `vivechak_prepare_generator` and generates the pipeline
2. Agent calls `vivechak_save_plan` to store it
3. User executes each session manually, saving outputs to the
   sessions directory
4. Agent calls `vivechak_next_session` between sessions to verify
   DAG progress and inject upstream context into the next prompt
5. Agent runs validation and gate at the end

### Pattern: Cross-Model Triangulation

For high-stakes one-way door decisions:

1. Agent runs the primary session via MCP (Model A)
2. User runs the same prompt manually on Model B
3. If findings agree: strong confidence, proceed
4. If findings conflict: use the Conflict Resolution template
   to resolve through structured analysis (ACH matrix)

---

## 7. Troubleshooting

### "The agent doesn't know my manual session is complete"

The agent discovers completed sessions by scanning `research/sessions/`
for `.md` files. Ensure:
- The file is saved in the correct directory
- The filename starts with the session ID (e.g., `T1-03-*.md`)
- The file has valid YAML frontmatter with `session_id`

### "Context from my manual session isn't being injected"

Context injection reads the session file and extracts key findings
sections. Ensure your manual output includes headings like:
- `## Recommendation` / `## Findings` / `## Key Findings`
- `## Conclusion` / `## Summary` / `## Results`

If the output uses different headings, the injector falls back to
including the first 4000 characters of the body.

### "The agent wants to re-run a session I already completed"

Check that:
1. The file exists in `research/sessions/`
2. The filename matches the expected session ID pattern
3. The frontmatter `session_id` field matches the DAG session ID

Run `vivechak_status` to see which sessions the server considers
complete.
