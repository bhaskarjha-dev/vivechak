import { execWithWatchdog } from "../watchdog.mjs";
import { BaseHarnessAdapter } from "./base.mjs";

export class AgyAdapter extends BaseHarnessAdapter {
  constructor() {
    super(
      "agy",
      "Google Antigravity CLI",
      "gemini-3.8-flash-high",
      ["gemini-3.8-flash-high", "gemini-3.8-flash", "claude-sonnet-4-6"],
      "cli"
    );
  }

  async probe() {
    const res = await execWithWatchdog("agy", ["--version"], { timeoutMs: 5000 });
    if (res.timedOut || res.exitCode !== 0) {
      return {
        available: false,
        error: res.stderr || "agy CLI not found in PATH or timed out",
        quotaState: "unknown"
      };
    }

    const version = res.stdout.trim();
    return {
      available: true,
      version,
      model: this.defaultModel,
      quotaState: "active"
    };
  }

  async runTurn(prompt, sandboxDir, model = this.defaultModel, timeoutMs = 90000) {
    let res = await execWithWatchdog(
      "agy",
      ["--dangerously-skip-permissions", "--model", model, "--print", prompt],
      { cwd: sandboxDir, timeoutMs }
    );

    if (this.isQuotaOrRateLimitError(res.stdout + " " + res.stderr)) {
      const fallback = this.getNextCandidate(model);
      if (fallback) {
        console.warn(`\n[MODEL_FAILOVER] ⚠️ Antigravity model ${model} quota/rate limited. Auto-failing over to ${fallback}...`);
        this.defaultModel = fallback;
        res = await execWithWatchdog(
          "agy",
          ["--dangerously-skip-permissions", "--model", fallback, "--print", prompt],
          { cwd: sandboxDir, timeoutMs }
        );
      }
    }

    return res;
  }

  async continueTurn(prompt, sandboxDir, timeoutMs = 90000) {
    return await execWithWatchdog(
      "agy",
      ["--continue", "--dangerously-skip-permissions", "--print", prompt],
      { cwd: sandboxDir, timeoutMs }
    );
  }
}
