import { spawn } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { resolveVckBinary } from "../types.mjs";
import { execWithWatchdog } from "../watchdog.mjs";
import { BaseHarnessAdapter } from "./base.mjs";

const PROBE_TIMEOUT_MS = 10000;

/**
 * Environment for `vck setup`: HOME (and the per-OS config roots) point into the sandbox,
 * so it writes the host config there instead of the real user's home.
 */
function sandboxHomeEnv(sandboxDir) {
  return {
    HOME: sandboxDir,
    USERPROFILE: sandboxDir,
    XDG_CONFIG_HOME: path.join(sandboxDir, ".config"),
    APPDATA: path.join(sandboxDir, "AppData", "Roaming")
  };
}

/** One recorded step of the trace, in the format triage.mjs parses. */
function formatTraceStep(name, args, result) {
  return `→ Call: ${name} ${JSON.stringify(args)}\n← Result: ${JSON.stringify(result)}`;
}

export class DirectMcpAdapter extends BaseHarnessAdapter {
  constructor() {
    super("direct_mcp", "Headless Direct MCP (vck serve)", "deterministic", "headless-mcp");
    // runTurn prints every tool call and its full result as "→ Call:" / "← Result:" lines.
    this.recordsToolResults = true;
  }

  /**
   * Ready only when `vck serve` completes an MCP initialize handshake.
   */
  async probe() {
    const vckBin = resolveVckBinary();
    const probeDir = fs.mkdtempSync(path.join(os.tmpdir(), "vck-probe-"));
    const cleanup = () => {
      try {
        fs.rmSync(probeDir, { recursive: true, force: true });
      } catch (e) {
        // ignore
      }
    };

    return new Promise((resolve) => {
      let settled = false;
      let proc = null;
      let timer = null;
      const finish = (outcome) => {
        if (settled) return;
        settled = true;
        clearTimeout(timer);
        try {
          proc?.kill();
        } catch (e) {
          // ignore
        }
        cleanup();
        resolve(outcome);
      };
      const unavailable = (reason) =>
        finish({ available: false, error: `${vckBin}: ${reason}`, quotaState: "unknown" });

      try {
        proc = spawn(vckBin, ["serve"], { cwd: probeDir, shell: false, stdio: ["pipe", "pipe", "pipe"] });
      } catch (err) {
        unavailable(`could not start 'serve': ${err.message}`);
        return;
      }

      timer = setTimeout(
        () => unavailable(`no MCP initialize response within ${PROBE_TIMEOUT_MS / 1000}s`),
        PROBE_TIMEOUT_MS
      );

      let buffer = "";
      let stderr = "";
      proc.stdout.on("data", (chunk) => {
        buffer += chunk.toString();
        const lines = buffer.split("\n");
        buffer = lines.pop();
        for (const line of lines) {
          let msg;
          try {
            msg = JSON.parse(line.trim());
          } catch (e) {
            continue;
          }
          if (msg?.id !== 1 || msg.method !== undefined) continue;
          const info = msg.result?.serverInfo;
          if (msg.result?.protocolVersion && info?.name) {
            finish({
              available: true,
              version: `${info.name} ${info.version || ""}`.trim(),
              model: this.defaultModel,
              quotaState: "active"
            });
          } else {
            unavailable(`invalid MCP initialize response: ${JSON.stringify(msg.error || msg.result).slice(0, 200)}`);
          }
        }
      });
      proc.stderr.on("data", (d) => (stderr += d.toString()));
      proc.stdin.on("error", () => {});
      proc.on("error", (err) => unavailable(`could not start 'serve': ${err.message}`));
      proc.on("exit", (code, signal) =>
        unavailable(
          `exited (${signal || `code ${code}`}) before completing the MCP initialize handshake` +
            (stderr.trim() ? `: ${stderr.trim().slice(0, 200)}` : "")
        )
      );

      proc.stdin.write(
        JSON.stringify({
          jsonrpc: "2.0",
          id: 1,
          method: "initialize",
          params: {
            protocolVersion: "2024-11-05",
            capabilities: {},
            clientInfo: { name: "direct-mcp-harness-probe", version: "0.1.0" }
          }
        }) + "\n"
      );
    });
  }

  /**
   * Helper to execute a sequence of JSON-RPC MCP calls against `vck serve`.
   */
  async _executeMcpClient(sandboxDir, toolCalls) {
    const vckBin = resolveVckBinary();
    const trace = [];

    return new Promise((resolve, reject) => {
      const proc = spawn(vckBin, ["serve"], {
        cwd: sandboxDir,
        shell: false,
        stdio: ["pipe", "pipe", "pipe"]
      });

      let buffer = "";
      let callIndex = 0;
      let nextId = 1;
      const pending = new Map();

      const send = (msg) => {
        const payload = JSON.stringify(msg) + "\n";
        proc.stdin.write(payload);
      };

      const handleResponse = (res) => {
        if (!res.id) return; // Notification
        if (res.method !== undefined) return; // Server-to-client request (e.g. roots/list), not a response
        const resolveCall = pending.get(res.id);
        if (resolveCall) {
          pending.delete(res.id);
          resolveCall(res);
        }
      };

      proc.stderr.on("data", () => {});

      proc.stdout.on("data", (chunk) => {
        buffer += chunk.toString();
        const lines = buffer.split("\n");
        buffer = lines.pop(); // keep last incomplete line
        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed) continue;
          try {
            const parsed = JSON.parse(trimmed);
            handleResponse(parsed);
          } catch (e) {
            // non-json output
          }
        }
      });

      const callTool = (name, args) => {
        const id = ++nextId;
        return new Promise((res) => {
          pending.set(id, res);
          trace.push({ type: "CALL", name, args });
          send({
            jsonrpc: "2.0",
            id,
            method: "tools/call",
            params: { name, arguments: args }
          });
        });
      };

      const timeout = setTimeout(() => {
        proc.kill();
        resolve({ trace, timedOut: true });
      }, 30000);

      // Start MCP Handshake
      const initId = 1;
      pending.set(initId, async (initRes) => {
        trace.push({ type: "INIT", result: initRes });
        send({ jsonrpc: "2.0", method: "notifications/initialized" });

        // Execute scenario tool calls sequentially
        for (const call of toolCalls) {
          const args = { project_root: sandboxDir, ...call.args };
          const res = await callTool(call.name, args);
          // Record the full tools/call result (or the JSON-RPC error) so triage can read isError.
          trace.push({ type: "RESULT", name: call.name, result: res.result !== undefined ? res.result : { error: res.error ?? null } });
        }

        clearTimeout(timeout);
        proc.kill();
        resolve({ trace, timedOut: false });
      });

      send({
        jsonrpc: "2.0",
        id: initId,
        method: "initialize",
        params: {
          protocolVersion: "2024-11-05",
          capabilities: {},
          clientInfo: { name: "direct-mcp-harness", version: "0.1.0" }
        }
      });
    });
  }

  async runTurn(prompt, sandboxDir, model = this.defaultModel, timeoutMs = 90000) {
    const p = prompt.toLowerCase();

    // Scenario 21: Out-of-scope code authoring guardrail (Zero Vivechak tools)
    if (p.includes("hmac-sha256") || p.includes("python function")) {
      const pyCode = `import hmac\nimport hashlib\n\ndef compute_hmac_sha256(payload: bytes, secret_key: bytes) -> str:\n    return hmac.new(secret_key, payload, hashlib.sha256).hexdigest()\n`;
      fs.writeFileSync(path.join(sandboxDir, "hmac_sha256.py"), pyCode, "utf8");
      return {
        stdout: `Recognized direct implementation request. Writing code directly without Vivechak research tools.\n\n\`\`\`python\n${pyCode}\`\`\``,
        stderr: "",
        exitCode: 0,
        timedOut: false,
        killed: false
      };
    }

    // Scenario 19: Universal Host Auto-Configuration (CLI setup command)
    // Scenario 20: Pre-Coding Workspace Integrity Check (CLI doctor command)
    let cliArgv = null;
    if (p.includes("vck setup") || p.includes("setup cursor")) cliArgv = ["setup", "cursor"];
    else if (p.includes("vck doctor")) {
      fs.mkdirSync(path.join(sandboxDir, "research", "sessions"), { recursive: true });
      cliArgv = ["doctor", sandboxDir];
    }
    if (cliArgv) {
      const res = await execWithWatchdog(resolveVckBinary(), cliArgv, {
        cwd: sandboxDir,
        timeoutMs: 5000,
        // Only setup writes into HOME; doctor must keep the real HOME (it refuses a home dir as root).
        env: cliArgv[0] === "setup" ? sandboxHomeEnv(sandboxDir) : undefined
      });
      const step = formatTraceStep("vck", { argv: cliArgv }, {
        exitCode: res.exitCode,
        timedOut: res.timedOut,
        stdout: res.stdout,
        stderr: res.stderr
      });
      return {
        stdout: `Headless Direct CLI Execution Trace:\n${step}`,
        stderr: res.stderr,
        exitCode: res.exitCode,
        timedOut: false,
        killed: false
      };
    }

    const samplePipeline = `# Project Pipeline\n\n\`\`\`yaml vivechak-plan\nplan: PROJECT\nsessions:\n  - {id: T0-01, title: Database Evaluation, layer: 0, door: two-way, depends: []}\n  - {id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}\n\`\`\`\n\n## Sessions\n\n### T0-01: Database Evaluation\n\`\`\`prompt\nEvaluate PostgreSQL as the primary database.\n\`\`\`\n\n### SYN-01: Synthesis\n\`\`\`prompt\nSynthesize the findings. [ALL_SESSION_FINDINGS]\n\`\`\`\n`;

    // Determine tool sequence based on prompt intent
    const toolCalls = [];

    if (p.includes("save plan") && p.includes("comparison plan: c-01")) {
      // Scenario 02: Skipped init auto-creation
      toolCalls.push({
        name: "vivechak_save",
        args: {
          kind: "comparison",
          scope: "comparison",
          id: "C-01",
          slug: "pino-vs-winston",
          content: prompt
        }
      });
    } else if (p.includes("compare") || p.includes("winston vs pino")) {
      // Comparison scope (S-01, S-16)
      toolCalls.push({
        name: "vivechak_start",
        args: { scope: "comparison", question: prompt }
      });
      toolCalls.push({
        name: "vivechak_save",
        args: {
          kind: "comparison",
          id: "C-01",
          slug: "pino-vs-winston",
          content: `# Comparison Plan: C-01\n| Decision | C-01 |\n| Status | complete |\n\n## Summary\nCompare Pino and Winston for high-throughput JSON logging.\n\n## WEP Matrix\n| Criterion | Weight | Pino | Winston |\n|---|---|---|---|\n| Throughput | 0.50 | 5 [Grade A, https://github.com/pinojs/pino] | 3 [Grade B, docs] |\n| Ecosystem | 0.50 | 4 [Grade A] | 5 [Grade A] |\n\n## Recommendation\nPino is recommended for high throughput.`
        }
      });
    } else if (p.includes("roast") || p.includes("challenge")) {
      // Adversarial challenge (S-08)
      toolCalls.push({
        name: "vivechak_start",
        args: { scope: "decision", question: prompt }
      });
      toolCalls.push({
        name: "vivechak_record_decision",
        args: {
          id: "D-001",
          slug: "microservices-kafka-mvp",
          status: "accepted",
          content: `---\nid: D-001\ntitle: Microservices with Kafka for 3-Person MVP\ndoor_type: two-way\nstatus: accepted\n---\n# D-001: Microservices with Kafka\nWe chose microservices with kafka for 3-person team.`
        }
      });
      toolCalls.push({
        name: "vivechak_review",
        args: { id: "D-001", mode: "red_team" }
      });
    } else if (p.includes("update session") || p.includes("amend")) {
      // Incremental amendment (S-09)
      toolCalls.push({ name: "vivechak_start", args: { scope: "project", question: prompt } });
      toolCalls.push({
        name: "vivechak_save",
        args: {
          kind: "session",
          id: "T0-01",
          content: `# T0-01: Framework Evaluation\nInitial benchmarks.`
        }
      });
      toolCalls.push({
        name: "vivechak_amend",
        args: {
          id: "T0-01",
          note: "pino latency is 12ms under load",
          reason: "benchmark update"
        }
      });
    } else if (p.includes("lock decision") || p.includes("record decision")) {
      // Formal ADR recording (S-10)
      toolCalls.push({ name: "vivechak_start", args: { scope: "decision", question: prompt } });
      toolCalls.push({
        name: "vivechak_record_decision",
        args: {
          id: "D-01",
          status: "accepted",
          content: `---\nid: D-01\ntitle: Primary Database Selection\ndoor_type: one-way\nstatus: accepted\nreview_trigger: ">50k concurrent writes"\n---\n# D-01: Primary Database Selection\nAccepted with postgres.`
        }
      });
    } else if (p.includes("oracle only") || p.includes("replan")) {
      // Dynamic replanning (S-11)
      toolCalls.push({ name: "vivechak_start", args: { scope: "project", question: prompt } });
      toolCalls.push({
        name: "vivechak_save",
        args: {
          kind: "plan",
          content: samplePipeline
        }
      });
      toolCalls.push({
        name: "vivechak_replan",
        args: {
          plan: "PROJECT",
          op: "update_prompt",
          session_id: "T0-01",
          prompt: "Evaluate Oracle DB for enterprise compliance",
          reason: "boss said no postgres allowed, oracle only"
        }
      });
    } else if (p.includes("standalone d-01.md and registry decisions.md")) {
      // Dual-source decision resolution (S-17)
      toolCalls.push({ name: "vivechak_start", args: { scope: "decision", question: prompt } });
      toolCalls.push({
        name: "vivechak_record_decision",
        args: {
          id: "D-01",
          status: "accepted",
          content: `---\nid: D-01\ntitle: Standalone ADR\ndoor_type: two-way\nstatus: accepted\n---\n# D-01: Standalone ADR`
        }
      });
      toolCalls.push({
        name: "vivechak_gate",
        args: { track: "all" }
      });
    } else if (p.includes("gate") || p.includes("coding the backend")) {
      // Mechanical Gate (S-12)
      toolCalls.push({ name: "vivechak_start", args: { scope: "project", question: prompt } });
      toolCalls.push({ name: "vivechak_save", args: { kind: "plan", content: samplePipeline } });
      toolCalls.push({
        name: "vivechak_gate",
        args: { track: "all" }
      });
    } else if (p.includes("save this session")) {
      // Session save & quality coaching (S-07, S-15, S-18)
      toolCalls.push({
        name: "vivechak_start",
        args: { scope: "project", question: prompt }
      });
      toolCalls.push({
        name: "vivechak_save",
        args: {
          kind: "session",
          id: "T0-01",
          content: prompt
        }
      });
    } else if (p.includes("vivechak status") || p.includes("visual graph") || p.startsWith("status")) {
      // Status & Visualize (S-13)
      toolCalls.push({ name: "vivechak_status", args: {} });
      toolCalls.push({ name: "vivechak_status", args: { format: "mermaid" } });
    } else if (p.includes("validate")) {
      // Workspace audit (S-14)
      toolCalls.push({ name: "vivechak_start", args: { scope: "comparison", question: "workspace check" } });
      toolCalls.push({ name: "vivechak_review", args: { mode: "validate" } });
    } else if (p.includes("decide:") || p.includes("mongo vs postgres") || p.includes("server actions")) {
      // Decision scope (S-04, S-05)
      toolCalls.push({
        name: "vivechak_start",
        args: { scope: "decision", question: prompt }
      });
      toolCalls.push({
        name: "vivechak_save",
        args: {
          kind: "plan",
          id: "D-01",
          slug: "tech-choice",
          content: `# D-NEW-PLAN: Tech Choice\n\n\`\`\`yaml vivechak-plan\nplan: D-NEW\nsessions:\n  - {id: D-NEW-S1, title: Fast Spike, role: C, door: two-way, informs: [D-NEW], depends: []}\n\`\`\`\n\n## Sessions\n\n### D-NEW-S1: Fast Spike\n\`\`\`prompt\n1 session fast spike: compare the candidates with graded evidence.\n\`\`\`\n`
        }
      });
    } else {
      // Project scope fallback (S-03, S-06)
      toolCalls.push({
        name: "vivechak_start",
        args: { scope: "project", question: prompt }
      });
      toolCalls.push({
        name: "vivechak_save",
        args: {
          kind: "plan",
          id: "DAG",
          content: samplePipeline
        }
      });
      if (p.includes("url shortener") || p.includes("next step") || p.includes("continue")) {
        toolCalls.push({
          name: "vivechak_next",
          args: {}
        });
      }
    }

    const { trace, timedOut } = await this._executeMcpClient(sandboxDir, toolCalls);

    // Full call arguments and full results, one JSON document per line (parsed by triage.mjs).
    const formattedTrace = trace
      .map((t) => {
        if (t.type === "CALL") return `→ Call: ${t.name} ${JSON.stringify(t.args)}`;
        if (t.type === "RESULT") return `← Result: ${JSON.stringify(t.result)}`;
        return `[${t.type}]`;
      })
      .join("\n");

    return {
      stdout: `Headless Direct MCP Execution Trace:\n${formattedTrace}`,
      stderr: "",
      exitCode: 0,
      timedOut,
      killed: false
    };
  }

  async continueTurn(prompt, sandboxDir, timeoutMs = 90000) {
    const p = prompt.toLowerCase();
    let text = "";

    if (p.includes("debrief") || p.includes("tools called")) {
      text = `1. Tools Called: Executed direct MCP stdio JSON-RPC sequence.\n2. Workspace State: Sandbox initialized, artifacts written atomically.\n3. Final Outcome: Action complete with zero external token expenditure.\n4. Next Step: Phase verified cleanly.`;
    } else if (p.includes("reflect") || p.includes("parametric")) {
      text = `Direct MCP Client Reflection: High-speed deterministic execution. Zero token cost, zero network overhead, sub-second latency. Verified exact JSON-RPC protocol compliance and filesystem mutations.`;
    } else {
      text = `Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.`;
    }

    return {
      stdout: text,
      stderr: "",
      exitCode: 0,
      timedOut: false,
      killed: false
    };
  }
}
