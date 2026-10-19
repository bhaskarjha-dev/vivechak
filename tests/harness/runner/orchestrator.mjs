import fs from "node:fs";
import path from "node:path";
import { AdapterRegistry } from "./adapters/registry.mjs";
import {
  getSandboxBaseDir,
  resolveVckBinary,
  SCENARIOS,
  TURN2_PROMPT,
  TURN3_PROMPT,
  TURN4_PROMPT
} from "./types.mjs";
import { execWithWatchdog } from "./watchdog.mjs";

const registry = new AdapterRegistry();

export async function runOrchestration(scenarioId, harnessName = "direct_mcp", modelOverride = null) {
  const scenario = SCENARIOS[scenarioId];
  if (!scenario) {
    throw new Error(`Unknown scenario ID: ${scenarioId}. Available: ${Object.keys(SCENARIOS).join(", ")}`);
  }

  const adapter = registry.get(harnessName);
  if (!adapter) {
    const known = registry.getAll().map((a) => a.name).join(", ");
    throw new Error(`Unsupported harness: ${harnessName}. Registered harnesses: ${known}`);
  }

  const model = modelOverride || adapter.defaultModel;
  const modelSlug = model.replace(/[^a-zA-Z0-9_-]/g, "_");

  console.log(`\n======================================================`);
  console.log(`🚀 STARTING AUTONOMOUS RUN`);
  console.log(`Scenario: [${scenario.id}] ${scenario.title}`);
  console.log(`Harness:  ${adapter.displayName} (${adapter.name})`);
  console.log(`Model:    ${model}`);
  console.log(`Scope:    ${scenario.scope}`);
  console.log(`======================================================\n`);

  // Step 1: Pre-flight probe
  console.log(`[PROBE] Running pre-flight health probe for ${adapter.name}...`);
  const probe = await adapter.probe();
  if (!probe.available) {
    console.error(`\n❌ PRE-FLIGHT PROBE FAILED: ${probe.error}`);
    return {
      success: false,
      status: "PREFLIGHT_FAIL",
      error: probe.error,
      scenarioId,
      harnessName,
      model
    };
  }
  console.log(`[PROBE] ✅ Health check passed: ${JSON.stringify(probe)}`);

  // Step 2: Create pristine, isolated sandbox
  const timestamp = Date.now();
  const baseDir = getSandboxBaseDir();
  const sandboxDir = path.join(baseDir, `s-${scenario.id}-${harnessName}-${timestamp}`);
  fs.mkdirSync(sandboxDir, { recursive: true });
  console.log(`[SANDBOX] Created isolated workspace: ${sandboxDir}`);

  // Step 3: Turn 1 (Scenario Prompt)
  const isProjectScope = scenario.scope === "project";
  const turn1Timeout = isProjectScope ? 180000 : 90000;
  console.log(`\n[Turn 1] Dispatching Scenario Prompt (Timeout: ${turn1Timeout / 1000}s)...`);
  const turn1 = await adapter.runTurn(scenario.prompt, sandboxDir, model, turn1Timeout);
  console.log(`[Turn 1] Completed in ${turn1.timedOut ? "TIMEOUT" : "OK"} (Exit code: ${turn1.exitCode})`);

  // Step 4: Turn 2 (Forensic Debrief)
  console.log(`\n[Turn 2] Dispatching Forensic Audit Card...`);
  const turn2 = await adapter.continueTurn(TURN2_PROMPT, sandboxDir);
  console.log(`[Turn 2] Completed (Exit code: ${turn2.exitCode})`);

  // Step 5: Turn 3 (Comparative Reflection)
  console.log(`\n[Turn 3] Dispatching Comparative Reflection...`);
  const turn3 = await adapter.continueTurn(TURN3_PROMPT, sandboxDir);
  console.log(`[Turn 3] Completed (Exit code: ${turn3.exitCode})`);

  // Step 6: Turn 4 (Meta-Improvements Probe)
  console.log(`\n[Turn 4] Dispatching Meta-Improvements Probe...`);
  const turn4 = await adapter.continueTurn(TURN4_PROMPT, sandboxDir);
  console.log(`[Turn 4] Completed (Exit code: ${turn4.exitCode})`);

  // Step 7: Inventory sandbox files
  const sandboxFiles = [];
  function walkDir(dir) {
    try {
      const entries = fs.readdirSync(dir);
      for (const entry of entries) {
        const full = path.join(dir, entry);
        const stat = fs.statSync(full);
        if (stat.isDirectory()) {
          walkDir(full);
        } else {
          sandboxFiles.push(path.relative(sandboxDir, full).replace(/\\/g, "/"));
        }
      }
    } catch (e) {
      // ignore
    }
  }
  walkDir(sandboxDir);

  // Step 8: Post-test vck doctor audit
  const vckBin = resolveVckBinary();
  const doctorRes = await execWithWatchdog(vckBin, ["doctor", sandboxDir], { timeoutMs: 5000 });

  // Step 9: Assemble markdown report
  const report = `# Autonomous Execution Report: Scenario ${scenario.id} (${scenario.name})
* **Date:** ${new Date().toISOString().slice(0, 10)}
* **Scenario ID:** ${scenario.id}
* **Scenario Title:** ${scenario.title}
* **Harness:** \`${harnessName}\`
* **Model:** \`${model}\`
* **Target Scope:** \`${scenario.scope}\`
* **Sandbox Workspace:** \`${sandboxDir.replace(/\\/g, "/")}\`

---

## 1. Turn 1: Primary Scenario Execution
**User Prompt:**
\`\`\`text
${scenario.prompt}
\`\`\`

**Model Response:**
${turn1.stdout || "(No output captured)"}
${turn1.stderr ? `\n*Stderr:*\n\`\`\`text\n${turn1.stderr}\n\`\`\`` : ""}

---

## 2. Turn 2: Structured Forensic Audit Card
**Prompt:**
\`\`\`text
${TURN2_PROMPT}
\`\`\`

**Model Response:**
${turn2.stdout || "(No output captured)"}

---

## 3. Turn 3: Comparative Reflection ("Vivechak vs. Without Vivechak")
**Prompt:**
\`\`\`text
${TURN3_PROMPT}
\`\`\`

**Model Response:**
${turn3.stdout || "(No output captured)"}

---

## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
\`\`\`text
${TURN4_PROMPT}
\`\`\`

**Model Response:**
${turn4.stdout || "(No output captured)"}

---

## 5. Sandbox Filesystem Inventory
\`\`\`text
${sandboxFiles.length > 0 ? sandboxFiles.join("\n") : "(No files created in sandbox)"}
\`\`\`

---

## 6. Workspace Doctor Diagnostics
\`\`\`text
${doctorRes.stdout || doctorRes.stderr || "(Doctor check clean)"}
\`\`\`
`;

  const runsDir = path.join(process.cwd(), "tests", "harness", "runs");
  fs.mkdirSync(runsDir, { recursive: true });
  const runFile = path.join(runsDir, `run-${scenario.id}-${harnessName}-${modelSlug}.md`);
  fs.writeFileSync(runFile, report, "utf8");

  console.log(`\n======================================================`);
  console.log(`✅ RUN COMPLETE! Artifact written to:`);
  console.log(`   ${runFile}`);
  console.log(`======================================================\n`);

  return {
    success: true,
    runFile,
    scenarioId: scenario.id,
    harnessName,
    model,
    sandboxFilesCount: sandboxFiles.length
  };
}
