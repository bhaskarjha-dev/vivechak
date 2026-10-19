import { execWithWatchdog } from "../watchdog.mjs";
import { BaseHarnessAdapter } from "./base.mjs";

export class OpenCodeAdapter extends BaseHarnessAdapter {
  constructor() {
    super(
      "opencode",
      "OpenCode CLI",
      "opencode/beam-free",
      [
        "opencode/beam-free",
        "opencode/big-pickle",
        "opencode/nemotron-3.5-lightning-free",
        "opencode/glm-5.3-flash-free",
        "opencode/space-bunny-free"
      ],
      "cli"
    );
  }

  async probe(targetModel = this.defaultModel) {
    const versionRes = await execWithWatchdog("opencode", ["--version"], { timeoutMs: 5000 });
    if (versionRes.timedOut || versionRes.exitCode !== 0) {
      return {
        available: false,
        error: versionRes.stderr || "opencode CLI not found in PATH or timed out",
        quotaState: "unknown"
      };
    }

    const version = versionRes.stdout.trim();

    // Check candidate models starting with targetModel
    const candidates = [targetModel, ...this.candidateModels.filter(m => m !== targetModel)];
    let lastError = "";

    for (const modelCandidate of candidates) {
      const testRes = await execWithWatchdog(
        "opencode",
        ["run", "--model", modelCandidate, "--auto", "ping"],
        { timeoutMs: 10000 }
      );

      if (testRes.timedOut) {
        lastError = `Model ${modelCandidate} ping timed out`;
        continue;
      }

      const combinedOutput = (testRes.stdout + " " + testRes.stderr).toLowerCase();
      if (combinedOutput.includes("credit expired") || combinedOutput.includes("quota exceeded") || combinedOutput.includes("out of credits")) {
        lastError = `Quota exhausted for ${modelCandidate}`;
        continue;
      }
      if (combinedOutput.includes("unauthenticated") || combinedOutput.includes("sign in required") || combinedOutput.includes("401")) {
        return {
          available: false,
          version,
          model: modelCandidate,
          quotaState: "unknown",
          error: "OpenCode requires authentication: " + testRes.stderr.trim()
        };
      }

      // Found a working model!
      this.defaultModel = modelCandidate;
      return {
        available: true,
        version,
        model: modelCandidate,
        quotaState: "active"
      };
    }

    return {
      available: false,
      version,
      model: targetModel,
      quotaState: "exhausted",
      error: `All OpenCode candidate models exhausted or timed out (${lastError})`
    };
  }

  async runTurn(prompt, sandboxDir, model = this.defaultModel, timeoutMs = 90000) {
    let res = await execWithWatchdog(
      "opencode",
      ["run", "--model", model, "--auto", prompt],
      { cwd: sandboxDir, timeoutMs }
    );

    if (this.isQuotaOrRateLimitError(res.stdout + " " + res.stderr)) {
      const fallback = this.getNextCandidate(model);
      if (fallback) {
        console.warn(`\n[MODEL_FAILOVER] ⚠️ Model ${model} credits exhausted. Auto-failing over to ${fallback}...`);
        this.defaultModel = fallback;
        res = await execWithWatchdog(
          "opencode",
          ["run", "--model", fallback, "--auto", prompt],
          { cwd: sandboxDir, timeoutMs }
        );
      }
    }

    return res;
  }

  async continueTurn(prompt, sandboxDir, timeoutMs = 90000) {
    return await execWithWatchdog(
      "opencode",
      ["run", "--continue", "--auto", prompt],
      { cwd: sandboxDir, timeoutMs }
    );
  }
}
