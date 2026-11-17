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

// ---------------------------------------------------------------------------------------------
// Verdicts come only from recorded evidence, never from words in the transcript:
//   1. every key call of the scenario is recorded with a successful result (an MCP tools/call
//      result with isError false; a CLI step with exit code 0), and
//   2. where the scenario names a file, that file is in the run's sandbox inventory.
// Adapters record a call as a "→ Call: <name> <json args>" line followed by a
// "← Result: <json result>" line in Turn 1 (see DirectMcpAdapter); runs without such records
// can't be verified and FAIL.
// ---------------------------------------------------------------------------------------------

const CALL_LINE = /^→ Call: ([A-Za-z0-9_.:/-]+)(?:\((.*)\)| (.*))?$/;
const RESULT_LINE = /^← Result: (.*)$/;
const CHALLENGE_MODES = new Set(["challenge", "red_team", "audit", "premortem"]);

const tool = (name, desc = name, pred = () => true) => ({
  desc,
  match: (r) => r.name === name && pred(r.args || {})
});
const cli = (subcommand) => ({
  desc: `vck ${subcommand}`,
  match: (r) => r.name === "vck" && Array.isArray(r.args?.argv) && r.args.argv[0] === subcommand
});
const reviewMode = (args) => {
  const mode = String(args.mode ?? "").trim().toLowerCase();
  if (!mode && args.challenge_type) return "challenge";
  return CHALLENGE_MODES.has(mode) ? "challenge" : "validate";
};
const file = (desc, test) => ({ desc, test });

const COMPARISON_FILE = file("research/C-NNN-comparison.md", (f) => /^research\/C-\d+-(?:comparison|cmp-[^/]+)\.md$/.test(f));
const DECISION_PLAN_FILE = file("research/D-NNN-plan.md", (f) => /^research\/D-\d+(?:-[^/]+)?-plan\.md$/.test(f));
const PIPELINE_FILE = file("research/RESEARCH-PIPELINE.md", (f) => f === "research/RESEARCH-PIPELINE.md");
const ADR_FILE = file(
  "research/D-NNN ADR",
  (f) => /^research\/D-\d+(?:-[^/]+)?\.md$/.test(f) && !/-(?:plan|R\d+|S\d+(?:-[^/]*)?)\.md$/.test(f)
);
const DECISIONS_FILE = file("research/DECISIONS.md", (f) => f === "research/DECISIONS.md");
const GATE_FILE = file("research/PHASE-0-GATE.md", (f) => f === "research/PHASE-0-GATE.md");
const CURSOR_CONFIG_FILE = file(".cursor/mcp.json", (f) => f === ".cursor/mcp.json");

/**
 * What each scenario must show in its recorded results (tests/harness/SCENARIOS.md):
 * `calls` are the key calls, `files` the files the scenario names. Scenario 21 passes only when
 * the run recorded its tool results and none of them is a Vivechak tool.
 */
export const SCENARIO_EXPECTATIONS = {
  "01": { calls: [tool("vivechak_start"), tool("vivechak_save")], files: [COMPARISON_FILE] },
  "02": { calls: [tool("vivechak_save")], files: [COMPARISON_FILE] },
  "03": { calls: [tool("vivechak_save"), tool("vivechak_next")], files: [PIPELINE_FILE] },
  "04": { calls: [tool("vivechak_start"), tool("vivechak_save")], files: [DECISION_PLAN_FILE] },
  "05": { calls: [tool("vivechak_start"), tool("vivechak_save")], files: [DECISION_PLAN_FILE] },
  "06": { calls: [tool("vivechak_start"), tool("vivechak_save")], files: [PIPELINE_FILE] },
  "07": { calls: [tool("vivechak_save")] },
  "08": { calls: [tool("vivechak_review", "vivechak_review (challenge)", (a) => reviewMode(a) === "challenge")] },
  "09": { calls: [tool("vivechak_amend")] },
  "10": { calls: [tool("vivechak_record_decision")], files: [ADR_FILE, DECISIONS_FILE] },
  "11": { calls: [tool("vivechak_replan")] },
  "12": { calls: [tool("vivechak_gate")], files: [GATE_FILE] },
  "13": {
    calls: [
      tool("vivechak_status"),
      tool("vivechak_status", "vivechak_status (format=mermaid or table)", (a) => ["mermaid", "table"].includes(String(a.format ?? "").toLowerCase()))
    ]
  },
  "14": { calls: [tool("vivechak_review", "vivechak_review (validate)", (a) => reviewMode(a) === "validate")] },
  "15": { calls: [tool("vivechak_save")] },
  "16": { calls: [tool("vivechak_start")] },
  "17": { calls: [tool("vivechak_gate")], files: [GATE_FILE] },
  "18": { calls: [tool("vivechak_save")] },
  "19": { calls: [cli("setup")], files: [CURSOR_CONFIG_FILE] },
  "20": { calls: [cli("doctor")] },
  "21": { forbiddenCall: /vivechak_/, forbiddenFile: /^research\// }
};

/**
 * Parses the recorded tool trace: each "→ Call:" line opens a record and the next
 * "← Result:" line holds its result (undefined when none was recorded, { unparsed } when the
 * line is not JSON).
 * @returns {Array<{ name: string, args: object, result: any }>}
 */
export function parseToolTrace(text) {
  const records = [];
  for (const raw of (text || "").split(/\r?\n/)) {
    const line = raw.trimEnd();
    const call = line.match(CALL_LINE);
    if (call) {
      const argsText = (call[2] ?? call[3] ?? "").trim();
      let args = {};
      if (argsText) {
        try {
          args = JSON.parse(argsText);
        } catch (e) {
          args = { unparsed: argsText };
        }
      }
      records.push({ name: call[1], args, result: undefined });
      continue;
    }
    const result = line.match(RESULT_LINE);
    if (result && records.length > 0 && records[records.length - 1].result === undefined) {
      try {
        records[records.length - 1].result = JSON.parse(result[1]);
      } catch (e) {
        records[records.length - 1].result = { unparsed: result[1].slice(0, 200) };
      }
    }
  }
  return records;
}

/** True for an MCP tools/call result with isError false, or a CLI step that exited 0. */
export function stepSucceeded(result) {
  if (!result || typeof result !== "object" || Array.isArray(result) || result.unparsed !== undefined) return false;
  if ("exitCode" in result) return result.exitCode === 0 && result.timedOut !== true;
  return Array.isArray(result.content) && result.isError !== true;
}

function clip(text) {
  return String(text).replace(/\s+/g, " ").trim().slice(0, 160).replace(/\|/g, "\\|");
}

function describeFailure(result) {
  if (result === undefined) return "no result was recorded";
  if (!result || typeof result !== "object" || Array.isArray(result)) return "the result is not a JSON object";
  if (result.unparsed !== undefined) return "the recorded result is not valid JSON";
  if ("exitCode" in result) return `exit code ${result.exitCode}${result.timedOut ? " (timed out)" : ""}`;
  if (result.error) return `JSON-RPC error: ${clip(result.error.message || JSON.stringify(result.error))}`;
  if (result.isError === true) {
    const text = (result.content || []).find((c) => c?.type === "text")?.text || "";
    let message = text;
    try {
      message = JSON.parse(text).message || text;
    } catch (e) {
      // plain-text error
    }
    return `isError: true (${clip(message)})`;
  }
  return "the result has no content";
}

function parseSandboxInventory(content) {
  const start = content.indexOf("## 5. Sandbox Filesystem Inventory");
  if (start === -1) return null;
  const block = content.slice(start).match(/```text\r?\n([\s\S]*?)\r?\n```/);
  if (!block) return null;
  return block[1]
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter((l) => l && l !== "(No files created in sandbox)");
}

function toolResultsFlag(content) {
  const m = content.match(/^\* \*\*Tool Results:\*\* `([^`]*)`/m);
  return m ? m[1] : null;
}

const fail = (reason) => ({ passed: false, reason });

export function evaluateScenarioPass(scenarioId, content, turn1Content) {
  const text = content || "";
  if (text.includes("PREFLIGHT_FAIL") || text.includes("TURN1_FAIL") || text.includes("(Skipped: Turn 1 did not complete successfully)")) {
    return fail("Turn 1 execution failed or timed out");
  }

  const spec = SCENARIO_EXPECTATIONS[scenarioId];
  if (!spec) return fail(`No expectation is defined for scenario ${scenarioId}`);

  const records = parseToolTrace(turn1Content || text);
  if (records.length === 0 && toolResultsFlag(text) !== "recorded") {
    return fail("No tool results were recorded for this run, so the scenario can't be verified");
  }

  for (const call of spec.calls || []) {
    const made = records.filter(call.match);
    if (made.length === 0) return fail(`Key call ${call.desc} was not recorded`);
    if (!made.some((r) => stepSucceeded(r.result))) {
      return fail(`Key call ${call.desc} failed: ${describeFailure(made[made.length - 1].result)}`);
    }
  }
  if (spec.forbiddenCall) {
    const used = records.find((r) => spec.forbiddenCall.test(r.name));
    if (used) return fail(`Vivechak tool ${used.name} was called for an out-of-scope request`);
  }

  if (spec.files || spec.forbiddenFile) {
    const inventory = parseSandboxInventory(text);
    if (inventory === null) return fail("The run has no sandbox file inventory, so its files can't be checked");
    for (const f of spec.files || []) {
      if (!inventory.some(f.test)) return fail(`Expected file ${f.desc} was not written`);
    }
    const created = spec.forbiddenFile ? inventory.find((f) => spec.forbiddenFile.test(f)) : undefined;
    if (created) return fail(`The out-of-scope request created ${created}`);
  }

  return { passed: true, reason: "" };
}

if (process.argv[1]?.endsWith("triage.mjs")) {
  triageRuns();
}
