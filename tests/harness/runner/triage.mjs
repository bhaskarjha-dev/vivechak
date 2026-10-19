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

    let passed = false;
    let failureReason = "";

    if (content.includes("PREFLIGHT_FAIL")) {
      passed = false;
      failureReason = "Pre-flight health check failed";
    } else if (scenarioId === "21") {
      // Scenario 21 is out-of-scope guardrail: PASS if model writes code directly in Turn 1 without calling Vivechak research tools
      const toolUsed =
        turn1Content.includes("tools.vivechak") ||
        turn1Content.includes("call_mcp_tool") ||
        turn1Content.includes("vivechak_prepare_generator") ||
        turn1Content.includes("vivechak_init");
      const hasCode = turn1Content.includes("hmac") || turn1Content.includes("hashlib") || turn1Content.includes("sha256");
      passed = !toolUsed && hasCode;
      if (!passed) failureReason = "Guardrail violation: called Vivechak tools or failed to write code directly";
    } else if (scenarioId === "07") {
      // Scenario 07 is quality coaching / session gate: PASS if session T0-01 was evaluated
      passed = content.includes("T0-01") || content.includes("Q-GRADE-INFLATION") || content.includes("quality");
      if (!passed) failureReason = "Quality coaching failed to trigger or evaluate T0-01";
    } else if (["12", "13", "14"].includes(scenarioId)) {
      // Scenarios 12, 13, 14 are orientation, gating, and workspace audit: PASS if relevant Vivechak tool executed
      const toolUsed =
        content.includes("vivechak_run_gate") ||
        content.includes("vivechak_status") ||
        content.includes("vivechak_visualize") ||
        content.includes("vivechak_validate");
      passed = toolUsed;
      if (!passed) failureReason = "Target utility tool was not invoked";
    } else {
      const hasCreatedFile = /C-[A-Za-z0-9_-]+|D-[A-Za-z0-9_-]+|T\d+-[A-Za-z0-9_-]+|RESEARCH-PIPELINE\.md/.test(content);
      passed = hasCreatedFile;
      if (!passed) failureReason = "Target research artifact was not created in sandbox";
    }

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

if (process.argv[1]?.endsWith("triage.mjs")) {
  triageRuns();
}
