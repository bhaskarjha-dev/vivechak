import { AgyAdapter } from "./agy.mjs";
import { ClaudeCodeAdapter } from "./claude_code.mjs";
import { DirectMcpAdapter } from "./direct_mcp.mjs";
import { OpenCodeAdapter } from "./opencode.mjs";

export class AdapterRegistry {
  constructor() {
    this.adapters = new Map();
    // Default standard adapters
    this.register(new DirectMcpAdapter());
    this.register(new AgyAdapter());
    this.register(new OpenCodeAdapter());
    this.register(new ClaudeCodeAdapter());
  }

  /**
   * Registers a harness adapter.
   * @param {import("./base.mjs").BaseHarnessAdapter} adapter
   */
  register(adapter) {
    this.adapters.set(adapter.name, adapter);
  }

  /**
   * Returns adapter by name.
   * @param {string} name
   */
  get(name) {
    return this.adapters.get(name);
  }

  /**
   * Returns all registered adapters.
   */
  getAll() {
    return Array.from(this.adapters.values());
  }

  /**
   * Probes all registered adapters in parallel.
   * @returns {Promise<Array<{ name: string, displayName: string, type: string, available: boolean, version?: string, model?: string, quotaState?: string, error?: string }>>}
   */
  async probeAll() {
    const adapters = this.getAll();
    const probePromises = adapters.map(async (adapter) => {
      try {
        const res = await adapter.probe();
        return {
          name: adapter.name,
          displayName: adapter.displayName,
          type: adapter.type,
          ...res
        };
      } catch (err) {
        return {
          name: adapter.name,
          displayName: adapter.displayName,
          type: adapter.type,
          available: false,
          error: err.message,
          quotaState: "unknown"
        };
      }
    });

    const settled = await Promise.allSettled(probePromises);
    return settled.map((s, idx) => {
      if (s.status === "fulfilled") return s.value;
      const adapter = adapters[idx];
      return {
        name: adapter.name,
        displayName: adapter.displayName,
        type: adapter.type,
        available: false,
        error: s.reason?.message || "Probe error",
        quotaState: "unknown"
      };
    });
  }

  /**
   * Returns only harnesses that are currently available and have active token quota.
   * @param {{ exclude?: string[] }} options
   */
  async getAvailableAdapters({ exclude = [] } = {}) {
    const probes = await this.probeAll();
    const available = [];
    for (const p of probes) {
      if (p.available && !exclude.includes(p.name)) {
        available.push(this.get(p.name));
      }
    }
    return available;
  }
}
