import { execWithWatchdog } from "../watchdog.mjs";
import { BaseHarnessAdapter } from "./base.mjs";

export class ClaudeCodeAdapter extends BaseHarnessAdapter {
  constructor() {
    super("claude_code", "Claude Code CLI", "claude-3-5-sonnet", "cli");
  }

  async probe() {
    const res = await execWithWatchdog("claude", ["--version"], { timeoutMs: 5000 });
    if (res.timedOut || res.exitCode !== 0) {
      return {
        available: false,
        error: res.stderr || "claude CLI not found in PATH or timed out",
        quotaState: "unknown"
      };
    }

    const version = res.stdout.trim();

    // Check for API key in environment
    const hasKey = !!process.env.ANTHROPIC_API_KEY;

    return {
      available: true,
      version,
      model: this.defaultModel,
      quotaState: hasKey ? "active" : "unknown"
    };
  }

  async runTurn(prompt, sandboxDir, model = this.defaultModel, timeoutMs = 90000) {
    return await execWithWatchdog(
      "claude",
      ["--dangerously-skip-permissions", "-p", JSON.stringify(prompt)],
      { cwd: sandboxDir, timeoutMs }
    );
  }

  async continueTurn(prompt, sandboxDir, timeoutMs = 90000) {
    return await execWithWatchdog(
      "claude",
      ["--dangerously-skip-permissions", "--continue", "-p", JSON.stringify(prompt)],
      { cwd: sandboxDir, timeoutMs }
    );
  }
}
