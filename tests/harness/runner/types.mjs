import fs from "node:fs";
import os from "node:os";
import path from "node:path";

export const TURN2_PROMPT = `Give me a concise debrief of your actions:
1. Tools Called: List each tool name called and key arguments passed (including aliases).
2. Workspace State: List all files/folders created, modified, or confirmed untouched.
3. Final Outcome: State the recommendation, decision, or artifact generated.
4. Next Step: What did Vivechak's next_step field recommend doing next?`;

export const TURN3_PROMPT = `Reflect on this session: Compare how you answered this question using Vivechak versus how you would have answered it without Vivechak (from pure parametric memory). What specific errors, ungrounded assumptions, or blind spots did the research catch or overturn? What are the honest tradeoffs (speed, token cost, rigor)?`;

export const TURN4_PROMPT = `Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?`;

/**
 * Dynamically resolves the path to the vck / vivechak binary.
 * Priority: VCK_BIN env var -> local project root build -> system PATH.
 */
export function resolveVckBinary() {
  if (process.env.VCK_BIN && fs.existsSync(process.env.VCK_BIN)) {
    return process.env.VCK_BIN;
  }

  const isWin = process.platform === "win32";
  const binaryName = isWin ? "vck.exe" : "vck";
  const localProjectBin = path.resolve(process.cwd(), binaryName);
  if (fs.existsSync(localProjectBin)) {
    return localProjectBin;
  }

  // Fallback to checking vivechak binary name
  const altName = isWin ? "vivechak.exe" : "vivechak";
  const localAlt = path.resolve(process.cwd(), altName);
  if (fs.existsSync(localAlt)) {
    return localAlt;
  }

  return binaryName;
}

/**
 * Dynamically resolves the base directory for isolated test sandboxes.
 * Priority: VIVECHAK_SANDBOX_DIR env var -> OS temp dir.
 */
export function getSandboxBaseDir() {
  if (process.env.VIVECHAK_SANDBOX_DIR) {
    return process.env.VIVECHAK_SANDBOX_DIR;
  }
  return path.join(os.tmpdir(), "vivechak-sandboxes");
}

export const SCENARIOS = {
  "01": {
    id: "01",
    name: "universal-option-comparison",
    scope: "comparison",
    title: "Universal Option Comparison (Winston vs Pino)",
    prompt: "vivechek compare winston vs pino for nodejs api.. need very fast logging bcz traffic is high"
  },
  "02": {
    id: "02",
    name: "skipped-init-auto-creation",
    scope: "comparison",
    title: "Skipped Init Auto-Creation (Resilient Save Plan)",
    prompt: `use vivechak save plan for this comparison between pino and winston:
# Comparison Plan: C-01
| Decision | C-01 |
| Status | complete |

## Summary
Compare Pino and Winston for high-throughput JSON logging.

## WEP Matrix
| Criterion | Weight | Pino | Winston |
|---|---|---|---|
| Throughput | 0.50 | 5 [A, https://github.com/pinojs/pino] | 3 [B, benchmark docs] |
| Ecosystem | 0.50 | 4 [A] | 5 [A] |

## Recommendation
Pino is recommended for high throughput.`
  },
  "03": {
    id: "03",
    name: "guided-worker-flow",
    scope: "project",
    title: "The Keep Going Test (DAG Resolution & Draft Sync)",
    prompt: "vivechak plan: minimal url shortener in golang with redis and postgres"
  },
  "04": {
    id: "04",
    name: "fast-architectural-decision",
    scope: "decision",
    title: "Fast Architectural Decision (Two-Way Door Fast Spike)",
    prompt: "bro shuld i use react 19 server actions or tanstack react query for fetching data in my nextjs dashboard? im totally confused"
  },
  "05": {
    id: "05",
    name: "high-stakes-one-way-door",
    scope: "decision",
    title: "High-Stakes One-Way Door (Regulatory X=3 Deep Research)",
    prompt: "vivechak decide: mongo vs postgres for our new clinic app. we gonna store patient medical records, lab reports and billings. we need to start asap"
  },
  "06": {
    id: "06",
    name: "full-venture-project-pipeline",
    scope: "project",
    title: "Full Venture / Project Pipeline (Multi-Tier DAG → FAD)",
    prompt: "hey vivechak plan my new startup idea. its basically uber but for dog walkers. need mobile app, real time tracking gps, stripe payments and push notifications. what tech stack should we build with?"
  },
  "07": {
    id: "07",
    name: "quality-coaching-gate",
    scope: "project",
    title: "Quality Coaching & 50-Char Concern Gate",
    prompt: `save this session output for T0-01:
---
id: T0-01
title: Framework Evaluation
status: complete
---
# Session T0-01: Framework Selection
We decided to use Gin framework. Gin is the fastest framework ever made [Grade A].
It handles 10 million requests per second with zero memory [Grade A].
Everyone uses Gin. No downsides found.`
  },
  "08": {
    id: "08",
    name: "adversarial-red-team-challenge",
    scope: "decision",
    title: "Adversarial Red-Team Challenge (Two-Tier Lookup)",
    prompt: "bro can u roast my decision? we chose microservices with kafka for our 3-person team mvp. use vivechak challenge"
  },
  "09": {
    id: "09",
    name: "incremental-session-amendment",
    scope: "decision",
    title: "Incremental Session / ADR Amendment (Two-Tier Lookup)",
    prompt: "update session T0-01 with this benchmark finding: pino latency is 12ms under load"
  },
  "10": {
    id: "10",
    name: "formal-adr-recording",
    scope: "decision",
    title: "Formal ADR Recording (Canonical ADR Compilation)",
    prompt: "lock decision D-01 as accepted with postgres. reversal trigger is >50k concurrent writes. rejected was mongo"
  },
  "11": {
    id: "11",
    name: "dynamic-dag-replanning",
    scope: "project",
    title: "Dynamic DAG Re-planning (Requirement Pivot)",
    prompt: "wait stop! boss said no postgres allowed, oracle only! can vivechak update our plan?"
  },
  "12": {
    id: "12",
    name: "mechanical-exit-gate",
    scope: "project",
    title: "Mechanical Phase 0 Exit Gate (Track Filtering)",
    prompt: "yo can i start coding the backend now?? run the gate thing in vivechak"
  },
  "13": {
    id: "13",
    name: "quick-status-and-visual-graph",
    scope: "project",
    title: "Status & Graph Visualization (Mermaid/ASCII)",
    prompt: "vivechak status? show me visual graph if u can"
  },
  "14": {
    id: "14",
    name: "multi-tier-workspace-audit",
    scope: "project",
    title: "Multi-Tier Workspace Audit (Automated Discovery)",
    prompt: "run vivechak validate on the whole workspace to make sure all our research files, adrs, plans, and comparisons are valid"
  },
  "15": {
    id: "15",
    name: "ast-code-fence-truncation-immunity",
    scope: "project",
    title: "AST Code Fence Truncation Immunity (CommonMark Fence Tracking)",
    prompt: `save this session output for T0-01:
---
session_id: T0-01
title: Kernel Buffer Tuning
door_type: one-way
status: draft
date: "2026-10-18"
---
# T0-01: Kernel Buffer Tuning
## Key Findings
- FoundationDB achieves sub-5ms commit latency with Redwood engine [Grade A: https://apple.github.io/foundationdb/].

## Discovered Concerns
Here are the primary production operating risks and kernel configuration commands:
\`\`\`bash
# Check current system file descriptor limits
cat /proc/sys/fs/file-max
# Verify socket buffer limits
sysctl net.core.rmem_max
# Ensure no silent truncation occurs here
ulimit -n 65536
\`\`\`
These buffer configuration parameters must be carefully benchmarked under load.

## Recommendation
Adopt FoundationDB with Redwood engine [Grade A: https://apple.github.io/foundationdb/].`
  },
  "16": {
    id: "16",
    name: "cli-shorthand-scope-triggers",
    scope: "comparison",
    title: "CLI Shorthand Scope Recognition (vck command prefixes)",
    prompt: "vck compare winston vs pino for nodejs api logging"
  },
  "17": {
    id: "17",
    name: "dual-source-decision-resolution",
    scope: "decision",
    title: "Dual-Source Decision Resolution (Zero Split-Brain)",
    prompt: "run gate on workspace with standalone D-01.md and registry DECISIONS.md"
  },
  "18": {
    id: "18",
    name: "evidence-ledger-cross-referencing",
    scope: "project",
    title: "Evidence Ledger URL Cross-Referencing",
    prompt: `save this session output for T0-01:
---
session_id: T0-01
title: Postgres Storage Engine
door_type: two-way
status: draft
date: "2026-10-18"
---
# T0-01: Postgres Storage Engine
## Key Findings
- [CLAIM-01] PostgreSQL 16 introduces btree indexing speedups [Grade: A | corroborated | fetched].
- [CLAIM-02] Sequential scan memory overhead reduced by 30% [Grade: B | fresh | test].

## Discovered Concerns
Vacuum freeze overhead on multi-terabyte tables requires scheduled maintenance windows.

## 5. Sources & Evidence Ledger
| # | Claim | Grade | Modifiers | Status | Source |
|---|---|---|---|---|---|
| 1 | CLAIM-01 | Grade A | corroborated | fetched | https://www.postgresql.org/docs/16/release-16.html |
| 2 | CLAIM-02 | Grade B | fresh | test | Internal hardware benchmark |

## Recommendation
Adopt PostgreSQL 16 [Grade A: https://www.postgresql.org/docs/16/release-16.html].`
  },
  "19": {
    id: "19",
    name: "universal-host-autoconfiguration",
    scope: "comparison",
    title: "Universal Host Auto-Configuration",
    prompt: "vck setup cursor"
  },
  "20": {
    id: "20",
    name: "pre-coding-workspace-integrity",
    scope: "project",
    title: "Pre-Coding Workspace Integrity Check",
    prompt: "vck doctor"
  },
  "21": {
    id: "21",
    name: "out-of-scope-anti-pattern",
    scope: "boundary",
    title: "Out-of-Scope Anti-Pattern Boundary (Direct Code Authoring)",
    prompt: "Write a python function to compute HMAC-SHA256 signature of a payload given a secret key."
  }
};
