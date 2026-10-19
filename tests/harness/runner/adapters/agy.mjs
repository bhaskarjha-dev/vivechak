import { execWithWatchdog } from "../watchdog.mjs";
import { BaseHarnessAdapter } from "./base.mjs";

export class AgyAdapter extends BaseHarnessAdapter {
  constructor() {
    super("agy", "Google Antigravity CLI", "gemini-3.8-flash-high", "cli");
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
    return await execWithWatchdog(
      "agy",
      ["--dangerously-skip-permissions", "--model", model, "--print", JSON.stringify(prompt)],
      { cwd: sandboxDir, timeoutMs }
    );
  }

  async continueTurn(prompt, sandboxDir, timeoutMs = 90000) {
    return await execWithWatchdog(
      "agy",
      ["--continue", "--dangerously-skip-permissions", "--print", JSON.stringify(prompt)],
      { cwd: sandboxDir, timeoutMs }
    );
  }
}
