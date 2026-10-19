import { execWithWatchdog } from "../watchdog.mjs";
import { BaseHarnessAdapter } from "./base.mjs";

export class OpenCodeAdapter extends BaseHarnessAdapter {
  constructor() {
    super("opencode", "OpenCode CLI", "opencode/beam-free", "cli");
  }

  async probe() {
    const versionRes = await execWithWatchdog("opencode", ["--version"], { timeoutMs: 5000 });
    if (versionRes.timedOut || versionRes.exitCode !== 0) {
      return {
        available: false,
        error: versionRes.stderr || "opencode CLI not found in PATH or timed out",
        quotaState: "unknown"
      };
    }

    const version = versionRes.stdout.trim();

    // Pre-flight check with beam-free model to verify authentication and credit availability
    const testRes = await execWithWatchdog(
      "opencode",
      ["run", "--model", this.defaultModel, "--auto", "ping"],
      { timeoutMs: 10000 }
    );

    if (testRes.timedOut) {
      return {
        available: true,
        version,
        model: this.defaultModel,
        quotaState: "unknown",
        error: "opencode model test timed out (daemon or network delay)"
      };
    }

    const combinedOutput = (testRes.stdout + " " + testRes.stderr).toLowerCase();
    if (combinedOutput.includes("credit expired") || combinedOutput.includes("quota exceeded") || combinedOutput.includes("out of credits")) {
      return {
        available: false,
        version,
        model: this.defaultModel,
        quotaState: "exhausted",
        error: "OpenCode free credits/quota exhausted for model " + this.defaultModel
      };
    }
    if (combinedOutput.includes("unauthenticated") || combinedOutput.includes("sign in required") || combinedOutput.includes("401")) {
      return {
        available: false,
        version,
        model: this.defaultModel,
        quotaState: "unknown",
        error: "OpenCode requires authentication: " + testRes.stderr.trim()
      };
    }

    return {
      available: true,
      version,
      model: this.defaultModel,
      quotaState: "active"
    };
  }

  async runTurn(prompt, sandboxDir, model = this.defaultModel, timeoutMs = 90000) {
    return await execWithWatchdog(
      "opencode",
      ["run", "--model", model, "--auto", JSON.stringify(prompt)],
      { cwd: sandboxDir, timeoutMs }
    );
  }

  async continueTurn(prompt, sandboxDir, timeoutMs = 90000) {
    return await execWithWatchdog(
      "opencode",
      ["run", "--continue", "--auto", JSON.stringify(prompt)],
      { cwd: sandboxDir, timeoutMs }
    );
  }
}
