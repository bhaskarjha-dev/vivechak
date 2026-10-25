import fs from "node:fs";
import path from "node:path";

export function triageRuns() {
  const runsDir = path.join(process.cwd(), "tests", "harness", "runs");
  if (!fs.existsSync(runsDir)) {
    console.log("No runs directory found at " + runsDir);
    return { total: 0, passed: 0, failed: 0, passRate: 0, runs: [] };
  }

  const files = fs.readdirSync(runsDir).filter((f) => f.endsWith(".md") && !f.startsWith("."));
  if (files.length === 0) {
    console.log("No active run files found in " + runsDir);
    return { total: 0, passed: 0, failed: 0, passRate: 0, runs: [] };
  }

  console.log(`\n======================================================`);
  console.log(`📊 FORENSIC HARNESS RUNS TRIAGE (${files.length} Runs Ingested)`);
  console.log(`======================================================\n`);

  const summary = [];
  const categorizedSuggestions = {
    dx: [],
    prompt: [],
    layout: [],
    general: []
  };

  for (const file of files) {
    const fullPath = path.join(runsDir, file);
    const content = fs.readFileSync(fullPath, "utf8");

    const scenarioMatch = content.match(/\* \*\*(?:Scenario ID|Scenario):\*\* (?:Scenario )?([0-9A-Za-z-]+)/);
    const harnessMatch = content.match(/\* \*\*Harness:\*\* `(.*)`/);
    const modelMatch = content.match(/\* \*\*Model:\*\* `(.*)`/);
    const scopeMatch = content.match(/\* \*\*Target Scope:\*\* `?(.*?)`?$/m);

    const scenarioId = scenarioMatch ? scenarioMatch[1].trim() : "Unknown";
    const harness = harnessMatch ? harnessMatch[1].trim() : "unknown";
    const model = modelMatch ? modelMatch[1].trim() : "unknown";
    const scope = scopeMatch ? scopeMatch[1].trim() : "comparison";

    const turn1Start = content.indexOf("## 1. Turn 1: Primary Scenario Execution");
    const turn2Start = content.indexOf("## 2. Turn 2: Structured Forensic Audit Card");
    const turn1Content = turn1Start !== -1 && turn2Start !== -1 ? content.slice(turn1Start, turn2Start) : content;

    const evalResult = evaluateScenarioPass(scenarioId, content, turn1Content);
    const passed = evalResult.passed;
    const failureReason = evalResult.reason;

    summary.push({
      file,
      scenarioId,
      harness,
      model,
      scope,
      status: passed ? "PASS" : "FAIL",
      failureReason
    });

    // Extract Turn 4
    const turn4Index = content.indexOf("## 4. Turn 4: Meta-Improvement Probe");
    if (turn4Index !== -1) {
      const endTurn4 = content.indexOf("## 5. Sandbox Filesystem Inventory", turn4Index);
      const turn4Section = endTurn4 !== -1 ? content.slice(turn4Index, endTurn4) : content.slice(turn4Index);

      const sourceTag = `${harness} (${model}) on S-${scenarioId}`;
      const lower = turn4Section.toLowerCase();

      if (lower.includes("dx") || lower.includes("parameter") || lower.includes("alias") || lower.includes("friction")) {
        categorizedSuggestions.dx.push({ source: sourceTag, text: turn4Section });
      }
      if (lower.includes("prompt") || lower.includes("brief") || lower.includes("ambiguity") || lower.includes("next_step")) {
        categorizedSuggestions.prompt.push({ source: sourceTag, text: turn4Section });
      }
      if (lower.includes("file") || lower.includes("layout") || lower.includes("artifact") || lower.includes("directory")) {
        categorizedSuggestions.layout.push({ source: sourceTag, text: turn4Section });
      }
    }
  }

  // Print scorecard
  console.log(`| Run File | Scenario | Harness | Model | Scope | Status |`);
  console.log(`|---|---|---|---|---|---|`);
  for (const s of summary) {
    console.log(
      `| \`${s.file}\` | **${s.scenarioId}** | \`${s.harness}\` | \`${s.model}\` | \`${s.scope}\` | ${
        s.status === "PASS" ? "✅ PASS" : "❌ FAIL: " + s.failureReason
      } |`
    );
  }

  const passedCount = summary.filter((s) => s.status === "PASS").length;
  const passRate = summary.length > 0 ? ((passedCount / summary.length) * 100).toFixed(1) : 0;

  console.log(`\n======================================================`);
  console.log(`📈 AGGREGATE SCORECARD`);
  console.log(`======================================================`);
  console.log(`Total Runs Ingested: ${summary.length}`);
  console.log(`Passed:              ${passedCount}`);
  console.log(`Failed:              ${summary.length - passedCount}`);
  console.log(`Pass Rate:           ${passRate}%\n`);

  console.log(`======================================================`);
  console.log(`💡 DYNAMIC TURN 4 META-IMPROVEMENTS SUMMARY`);
  console.log(`======================================================`);
  console.log(`• DX / Parameter Suggestions:        ${categorizedSuggestions.dx.length} items surfaced`);
  console.log(`• Prompt & Brief Clarity:           ${categorizedSuggestions.prompt.length} items surfaced`);
  console.log(`• File & Layout Ergonomics:         ${categorizedSuggestions.layout.length} items surfaced\n`);

  return {
    total: summary.length,
    passed: passedCount,
    failed: summary.length - passedCount,
    passRate,
    runs: summary,
    suggestions: categorizedSuggestions
  };
}
 
export function evaluateScenarioPass(scenarioId, content, turn1Content) {
  if (content.includes("PREFLIGHT_FAIL") || content.includes("TURN1_FAIL") || content.includes("(Skipped: Turn 1 did not complete successfully)")) {
    return { passed: false, reason: "Turn 1 execution failed or timed out" };
  }

  const normContent = content.toLowerCase();
  const normTurn1 = (turn1Content || "").toLowerCase();

  switch (scenarioId) {
    case "01": // Universal Option Comparison (Winston vs Pino)
    case "02": // Skipped Init Auto-Creation (Resilient Save Plan)
    case "16": // CLI Shorthand Scope Recognition (vck compare)
      {
        const hasComparisonArtifact =
          /c-[a-z0-9_-]+|comparison\.md/i.test(content) ||
          normContent.includes("wep matrix") ||
          content.includes("vivechak_start") ||
          content.includes("vivechak_save") ||
          content.includes("vivechak_prepare_generator") ||
          content.includes("vivechak_save_plan");
        const mentionsOptions =
          (normContent.includes("pino") && normContent.includes("winston")) ||
          normContent.includes("c-01");
        if (hasComparisonArtifact && mentionsOptions) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "Missing comparison matrix (WEP) or option comparison output" };
      }

    case "03": // Guided Worker Flow (DAG Resolution & Draft Sync)
    case "06": // Full Venture / Project Pipeline (Multi-Tier DAG -> FAD)
      {
        const hasProjectArtifact =
          content.includes("RESEARCH-PIPELINE.md") ||
          /t\d+-[a-z0-9_-]+/i.test(content) ||
          content.includes("vivechak_start") ||
          content.includes("vivechak_prepare_generator") ||
          content.includes("vivechak_init");
        if (hasProjectArtifact) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "Project DAG pipeline or session artifacts not generated" };
      }

    case "04": // Fast Architectural Decision (Two-Way Door Fast Spike)
    case "05": // High-Stakes One-Way Door (Regulatory X=3 Deep Research)
      {
        const hasAdrOrPlan =
          /d-[a-z0-9_-]+/i.test(content) ||
          content.includes("DECISIONS.md") ||
          content.includes("vivechak_start") ||
          content.includes("vivechak_prepare_generator") ||
          content.includes("scope=\"decision\"") ||
          normContent.includes("decision plan");
        if (hasAdrOrPlan) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "Decision scope ADR or decision plan artifact missing" };
      }

    case "07": // Quality Coaching & 50-Char Concern Gate
      {
        const caughtQuality =
          content.includes("T0-01") ||
          content.includes("Q-GRADE-INFLATION") ||
          normContent.includes("grade inflation") ||
          normContent.includes("quality") ||
          normContent.includes("draft");
        if (caughtQuality) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "Quality coaching failed to trigger or evaluate T0-01" };
      }

    case "08": // Adversarial Red-Team Challenge
      {
        const usedChallenge =
          content.includes("vivechak_review") ||
          content.includes("vivechak_challenge") ||
          normContent.includes("red_team") ||
          normContent.includes("challenge") ||
          normContent.includes("critique") ||
          normContent.includes("blast radius");
        if (usedChallenge) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "vivechak_review or vivechak_challenge was not invoked or evaluated" };
      }

    case "09": // Incremental Session / ADR Amendment
      {
        const usedAmend =
          content.includes("vivechak_amend") ||
          content.includes("vivechak_amend_session") ||
          normContent.includes("amend") ||
          normContent.includes("updated session") ||
          normContent.includes("12ms");
        if (usedAmend) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "vivechak_amend or vivechak_amend_session was not invoked or executed" };
      }

    case "10": // Formal ADR Recording (Canonical ADR Compilation)
      {
        const recordedAdr =
          content.includes("vivechak_record_decision") ||
          /d-01\.md|decisions\.md/i.test(content) ||
          normContent.includes("reversal trigger");
        if (recordedAdr) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "vivechak_record_decision was not invoked or ADR not recorded" };
      }

    case "11": // Dynamic DAG Re-planning (Requirement Pivot)
      {
        const usedReplan =
          content.includes("vivechak_replan") ||
          normContent.includes("replan") ||
          normContent.includes("invalidate") ||
          normContent.includes("oracle");
        if (usedReplan) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "vivechak_replan was not invoked or plan not updated" };
      }

    case "12": // Mechanical Phase 0 Exit Gate (Track Filtering)
    case "17": // Dual-Source Decision Resolution
      {
        const ranGate =
          content.includes("vivechak_gate") ||
          content.includes("vivechak_run_gate") ||
          normContent.includes("phase-0-gate") ||
          normContent.includes("track a") ||
          normContent.includes("track b") ||
          normContent.includes("exit gate");
        if (ranGate) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "Phase 0 Gate evaluation was not executed" };
      }

    case "13": // Status & Graph Visualization (Mermaid/ASCII)
      {
        const checkedStatus =
          content.includes("vivechak_status") ||
          content.includes("vivechak_visualize") ||
          normContent.includes("mermaid") ||
          normContent.includes("dag status") ||
          normContent.includes("pipeline");
        if (checkedStatus) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "vivechak_status or vivechak_visualize not invoked" };
      }

    case "14": // Multi-Tier Workspace Audit (Automated Discovery)
      {
        const ranValidate =
          content.includes("vivechak_review") ||
          content.includes("vivechak_validate") ||
          normContent.includes("workspace validation") ||
          normContent.includes("all artifacts valid");
        if (ranValidate) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "vivechak_review or vivechak_validate was not invoked" };
      }

    case "15": // AST Code Fence Truncation Immunity
    case "18": // Evidence Ledger URL Cross-Referencing
      {
        const savedSession =
          content.includes("vivechak_save") ||
          content.includes("vivechak_save_session") ||
          normContent.includes("session saved") ||
          normContent.includes("claim-01") ||
          normContent.includes("status: complete");
        if (savedSession) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "Session output was not processed or saved" };
      }

    case "19": // Universal Host Auto-Configuration
      {
        const ranSetup =
          normContent.includes("vck setup") ||
          normContent.includes("cursor") ||
          normContent.includes(".cursor/mcp.json") ||
          normContent.includes("mcp-config") ||
          normContent.includes("vivechak");
        if (ranSetup) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "vck setup or host MCP config not emitted" };
      }

    case "20": // Pre-Coding Workspace Integrity Check
      {
        const ranDoctor =
          normContent.includes("vck doctor") ||
          normContent.includes("doctor diagnostics") ||
          normContent.includes("workspace integrity") ||
          normContent.includes("clean");
        if (ranDoctor) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "vck doctor audit not performed" };
      }

    case "21": // Out-of-Scope Anti-Pattern Boundary (Direct Code Authoring)
      {
        const toolUsed =
          turn1Content.includes("tools.vivechak") ||
          turn1Content.includes("call_mcp_tool") ||
          turn1Content.includes("vivechak_prepare_generator") ||
          turn1Content.includes("vivechak_init");
        const hasCode =
          normTurn1.includes("hmac") ||
          normTurn1.includes("hashlib") ||
          normTurn1.includes("sha256") ||
          normTurn1.includes("def ");
        if (!toolUsed && hasCode) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "Guardrail violation: called Vivechak tools or failed to write code directly" };
      }

    default:
      {
        const hasArtifact = /C-[A-Za-z0-9_-]+|D-[A-Za-z0-9_-]+|T\d+-[A-Za-z0-9_-]+|RESEARCH-PIPELINE\.md/.test(content);
        if (hasArtifact) {
          return { passed: true, reason: "" };
        }
        return { passed: false, reason: "Target artifact was not found in execution output" };
      }
  }
}

if (process.argv[1]?.endsWith("triage.mjs")) {
  triageRuns();
}
