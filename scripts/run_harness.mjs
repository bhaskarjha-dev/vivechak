#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { AdapterRegistry } from "../tests/harness/runner/adapters/registry.mjs";
import { runOrchestration } from "../tests/harness/runner/orchestrator.mjs";
import { triageRuns } from "../tests/harness/runner/triage.mjs";
import { getSandboxBaseDir, SCENARIOS } from "../tests/harness/runner/types.mjs";

const CORE_BATCH = ["01", "02", "04", "05", "08", "21"];
const EXTENDED_BATCH = ["03", "06", "07", "12", "13", "14", "16"];
const ALL_BATCH = Object.keys(SCENARIOS);

const registry = new AdapterRegistry();

function printHelp() {
  console.log(`
Vivechak Universal Multi-Harness Autonomous Test Runner

USAGE:
  node scripts/run_harness.mjs [options]

COMMAND MODES:
  --probe                   Run pre-flight health & quota probes across all registered harnesses
  --triage                  Ingest and score all ephemeral runs in tests/harness/runs/
  --clean                   Wipe ephemeral run markdown files from tests/harness/runs/
  --clean-sandboxes         Wipe all isolated sandbox directories from the OS temp directory
  --archive <version>       Compile active runs into tests/harness/reports/audit-<version>-<date>.md and clean

EXECUTION OPTIONS:
  --auto                    Auto-select all harnesses that are currently available and have active quota
  --harness <name>          Target specific harness (direct_mcp, agy, opencode, claude_code) [default: direct_mcp]
  --exclude <name>          Exclude specific harness from --auto execution (e.g. --exclude agy)
  --scenario <id[,id2]>     Target specific scenario ID(s) (e.g. --scenario 01,02) [default: 01]
  --batch <core|extended|all> Run a preset batch of scenarios [default: core]
  --model <model>           Override default LLM model for the target harness
  --help                    Show this help message
`);
}

async function handleProbe() {
  console.log("\n🔍 Running Pre-Flight Health & Token/Quota Probes across all harnesses...\n");
  const probes = await registry.probeAll();

  console.log(`| Harness | Display Name | Type | Status | Quota / Token State | Version / Error |`);
  console.log(`|---|---|---|---|---|---|`);
  for (const p of probes) {
    const statusIcon = p.available ? "✅ READY" : "❌ UNAVAILABLE";
    const quota = p.quotaState || "unknown";
    const detail = p.available ? p.version || "ok" : p.error || "unavailable";
    console.log(`| \`${p.name}\` | **${p.displayName}** | \`${p.type}\` | ${statusIcon} | \`${quota}\` | ${detail} |`);
  }
  console.log("");
}

function handleClean() {
  const runsDir = path.join(process.cwd(), "tests", "harness", "runs");
  if (!fs.existsSync(runsDir)) {
    console.log("No runs directory found.");
    return;
  }
  const files = fs.readdirSync(runsDir).filter((f) => f.endsWith(".md"));
  for (const file of files) {
    fs.unlinkSync(path.join(runsDir, file));
  }
  console.log(`\n🧹 Clean complete: Removed ${files.length} ephemeral run files from tests/harness/runs/\n`);
}

function handleCleanSandboxes() {
  const baseDir = getSandboxBaseDir();
  if (!fs.existsSync(baseDir)) {
    console.log(`\n🧹 No sandbox directory found at ${baseDir}\n`);
    return;
  }
  const entries = fs.readdirSync(baseDir);
  let count = 0;
  for (const entry of entries) {
    if (entry.startsWith("s-")) {
      const full = path.join(baseDir, entry);
      try {
        fs.rmSync(full, { recursive: true, force: true, maxRetries: 5, retryDelay: 200 });
        count++;
      } catch (err) {
        console.warn(`Could not remove ${entry}: ${err.message}`);
      }
    }
  }
  console.log(`\n🧹 Sandbox Clean complete: Removed ${count} sandbox directories from ${baseDir}\n`);
}

async function handleArchive(version) {
  if (!version) {
    console.error("Error: --archive requires a version tag (e.g. --archive v0.1.0)");
    process.exit(1);
  }

  const triage = triageRuns();
  if (triage.total === 0) {
    console.error("No active runs to archive.");
    return;
  }

  const date = new Date().toISOString().slice(0, 10);
  const reportsDir = path.join(process.cwd(), "tests", "harness", "reports");
  fs.mkdirSync(reportsDir, { recursive: true });

  const reportFileName = `audit-${version}-${date}.md`;
  const reportFilePath = path.join(reportsDir, reportFileName);

  let reportContent = `# Vivechak Release Audit Report: ${version}
* **Release Version:** \`${version}\`
* **Date:** ${date}
* **Total Runs Evaluated:** ${triage.total}
* **Pass Rate:** ${triage.passRate}% (${triage.passed} / ${triage.total} PASS)

---

## 1. Execution Scorecard
| Run File | Scenario | Harness | Model | Scope | Status |
|---|---|---|---|---|---|
`;

  for (const r of triage.runs) {
    reportContent += `| \`${r.file}\` | **${r.scenarioId}** | \`${r.harness}\` | \`${r.model}\` | \`${r.scope}\` | ${
      r.status === "PASS" ? "✅ PASS" : "❌ FAIL: " + r.failureReason
    } |\n`;
  }

  reportContent += `\n---\n\n## 2. Dynamic Turn 4 Meta-Improvements\n`;
  reportContent += `### DX Friction Items (${triage.suggestions.dx.length})\n`;
  for (const item of triage.suggestions.dx) {
    reportContent += `* **${item.source}:**\n${item.text}\n\n`;
  }

  reportContent += `### Prompt / Brief Clarifications (${triage.suggestions.prompt.length})\n`;
  for (const item of triage.suggestions.prompt) {
    reportContent += `* **${item.source}:**\n${item.text}\n\n`;
  }

  reportContent += `### File & Layout Ergonomics (${triage.suggestions.layout.length})\n`;
  for (const item of triage.suggestions.layout) {
    reportContent += `* **${item.source}:**\n${item.text}\n\n`;
  }

  fs.writeFileSync(reportFilePath, reportContent, "utf8");
  console.log(`\n📦 Successfully compiled milestone audit report:`);
  console.log(`   ${reportFilePath}`);

  // Update reports README
  const indexFile = path.join(reportsDir, "README.md");
  let indexContent = fs.existsSync(indexFile) ? fs.readFileSync(indexFile, "utf8") : "# Vivechak Release Audit Reports\n\n";
  const entry = `- **[${reportFileName}](${reportFileName}):** Release ${version} milestone audit (${triage.total} runs, ${triage.passRate}% pass).\n`;
  if (!indexContent.includes(reportFileName)) {
    indexContent += entry;
    fs.writeFileSync(indexFile, indexContent, "utf8");
  }

  // Clean ephemeral runs
  handleClean();
}

async function main() {
  const args = process.argv.slice(2);

  if (args.includes("--help") || args.includes("-h")) {
    printHelp();
    return;
  }

  if (args.includes("--probe")) {
    await handleProbe();
    return;
  }

  if (args.includes("--triage")) {
    triageRuns();
    return;
  }

  if (args.includes("--clean")) {
    handleClean();
    return;
  }

  if (args.includes("--clean-sandboxes")) {
    handleCleanSandboxes();
    return;
  }

  const archiveIdx = args.indexOf("--archive");
  if (archiveIdx !== -1) {
    const version = args[archiveIdx + 1];
    await handleArchive(version);
    return;
  }

  let scenarioArg = "01";
  let harnessArg = "direct_mcp";
  let modelArg = null;
  let isAuto = false;
  let isBatch = false;
  let batchMode = "core";
  const excludeList = [];

  for (let i = 0; i < args.length; i++) {
    if (args[i] === "--scenario" && args[i + 1]) scenarioArg = args[++i];
    if (args[i] === "--harness" && args[i + 1]) harnessArg = args[++i];
    if (args[i] === "--model" && args[i + 1]) modelArg = args[++i];
    if (args[i] === "--auto") isAuto = true;
    if (args[i] === "--exclude" && args[i + 1]) excludeList.push(args[++i]);
    if (args[i] === "--batch") {
      isBatch = true;
      if (args[i + 1] && !args[i + 1].startsWith("--")) {
        batchMode = args[++i];
      }
    }
  }

  let targets = [scenarioArg.trim().padStart(2, "0")];
  if (isBatch) {
    if (batchMode === "extended") targets = EXTENDED_BATCH;
    else if (batchMode === "all") targets = ALL_BATCH;
    else targets = CORE_BATCH;
  } else if (scenarioArg.includes(",")) {
    targets = scenarioArg.split(",").map((s) => s.trim().padStart(2, "0"));
  }

  let harnesses = [harnessArg];
  if (isAuto) {
    console.log(`🔍 Probing available and funded harnesses (excluding: ${excludeList.join(", ") || "none"})...`);
    const availableAdapters = await registry.getAvailableAdapters({ exclude: excludeList });
    if (availableAdapters.length === 0) {
      console.warn("⚠️ No external AI harnesses available with active quota. Falling back to direct_mcp.");
      harnesses = ["direct_mcp"];
    } else {
      harnesses = availableAdapters.map((a) => a.name);
    }
  }

  console.log(`\n🎯 TARGET EXECUTION PLAN:`);
  console.log(`Scenarios: ${targets.join(", ")}`);
  console.log(`Harnesses: ${harnesses.join(", ")}\n`);

  for (const h of harnesses) {
    for (const s of targets) {
      try {
        await runOrchestration(s, h, modelArg);
      } catch (err) {
        console.error(`❌ Scenario ${s} failed on ${h}:`, err.message);
      }
    }
  }
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
