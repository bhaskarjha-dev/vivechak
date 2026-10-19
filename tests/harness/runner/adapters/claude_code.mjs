import { execWithWatchdog } from "../watchdog.mjs";
import { BaseHarnessAdapter } from "./base.mjs";

export class ClaudeCodeAdapter extends BaseHarnessAdapter {
  constructor() {
    super(
      "claude_code",
      "Claude Code CLI",
      "claude-3-5-sonnet",
      ["claude-3-5-sonnet", "claude-3-5-haiku"],
      "cli"
    );
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
    let res = await execWithWatchdog(
      "claude",
      ["--dangerously-skip-permissions", "-p", prompt],
      { cwd: sandboxDir, timeoutMs }
    );

    if (this.isQuotaOrRateLimitError(res.stdout + " " + res.stderr)) {
      const fallback = this.getNextCandidate(model);
      if (fallback) {
        console.warn(`\n[MODEL_FAILOVER] ⚠️ Claude Code model ${model} quota/rate limited. Auto-failing over to ${fallback}...`);
        this.defaultModel = fallback;
        res = await execWithWatchdog(
          "claude",
          ["--dangerously-skip-permissions", "-p", prompt],
          { cwd: sandboxDir, timeoutMs }
        );
      }
    }

    return res;
  }

  async continueTurn(prompt, sandboxDir, timeoutMs = 90000) {
    return await execWithWatchdog(
      "claude",
      ["--dangerously-skip-permissions", "--continue", "-p", prompt],
      { cwd: sandboxDir, timeoutMs }
    );
  }
}
