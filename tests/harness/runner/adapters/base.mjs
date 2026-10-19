/**
 * BaseHarnessAdapter
 * Abstract contract for any AI agent or MCP client test harness.
 */
export class BaseHarnessAdapter {
  /**
   * @param {string} name Short identifier (e.g. "agy", "opencode", "claude_code", "direct_mcp")
   * @param {string} displayName Human-readable label (e.g. "Google Antigravity CLI")
   * @param {string} defaultModel Default LLM model identifier
   * @param {string[]} candidateModels Prioritized fallback models for failover
   * @param {"cli" | "headless-mcp" | "ide"} type Architectural classification
   */
  constructor(name, displayName, defaultModel, candidateModels = [], type = "cli") {
    this.name = name;
    this.displayName = displayName;
    this.defaultModel = defaultModel;
    this.candidateModels = candidateModels.length > 0 ? candidateModels : [defaultModel];
    this.type = type;
  }

  /**
   * Returns the next candidate model in the cascade, or null if exhausted.
   * @param {string} currentModel
   * @returns {string | null}
   */
  getNextCandidate(currentModel) {
    const idx = this.candidateModels.indexOf(currentModel);
    if (idx !== -1 && idx + 1 < this.candidateModels.length) {
      return this.candidateModels[idx + 1];
    }
    return null;
  }

  /**
   * Detects whether command output indicates quota exhaustion or rate limits.
   * @param {string} output
   * @returns {boolean}
   */
  isQuotaOrRateLimitError(output) {
    const lower = (output || "").toLowerCase();
    return (
      lower.includes("credit expired") ||
      lower.includes("quota exceeded") ||
      lower.includes("out of credits") ||
      lower.includes("rate limit") ||
      lower.includes("resource_exhausted") ||
      lower.includes("too many requests") ||
      lower.includes("429")
    );
  }

  /**
   * Probes harness availability, CLI binary, authentication, and token/quota status.
   * @returns {Promise<{ available: boolean, error?: string, version?: string, model?: string, quotaState?: "active"|"exhausted"|"unknown" }>}
   */
  async probe() {
    throw new Error(`probe() not implemented for adapter: ${this.name}`);
  }

  /**
   * Executes Turn 1 (initial scenario prompt) in the specified sandbox workspace.
   * @param {string} prompt The scenario prompt string
   * @param {string} sandboxDir Pristine isolated workspace directory path
   * @param {string} model Target model override (or this.defaultModel)
   * @param {number} timeoutMs Execution timeout before watchdog clamps
   * @returns {Promise<{ stdout: string, stderr: string, exitCode: number, timedOut: boolean, killed: boolean }>}
   */
  async runTurn(prompt, sandboxDir, model = this.defaultModel, timeoutMs = 90000) {
    throw new Error(`runTurn() not implemented for adapter: ${this.name}`);
  }

  /**
   * Continues the multi-turn session in the existing sandbox workspace.
   * @param {string} prompt Follow-up prompt string (debrief, reflection, or meta-probe)
   * @param {string} sandboxDir The existing sandbox workspace
   * @param {number} timeoutMs Execution timeout before watchdog clamps
   * @returns {Promise<{ stdout: string, stderr: string, exitCode: number, timedOut: boolean, killed: boolean }>}
   */
  async continueTurn(prompt, sandboxDir, timeoutMs = 90000) {
    throw new Error(`continueTurn() not implemented for adapter: ${this.name}`);
  }
}
