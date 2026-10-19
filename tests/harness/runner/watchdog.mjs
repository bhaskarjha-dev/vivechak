import { spawn } from "node:child_process";

/**
 * Executes a subprocess with an ironclad watchdog that guarantees zero stalling.
 * @param {string} cmd The executable or command.
 * @param {string[]} args Command line arguments.
 * @param {object} options Options including cwd, timeoutMs, stdio.
 * @returns {Promise<{ stdout: string, stderr: string, exitCode: number, timedOut: boolean, killed: boolean }>}
 */
export function execWithWatchdog(cmd, args = [], options = {}) {
  const timeoutMs = options.timeoutMs || 90000; // 90 second default
  const maxOutputChars = options.maxOutputChars || 60000;
  const cwd = options.cwd || process.cwd();

  return new Promise((resolve) => {
    let stdout = "";
    let stderr = "";
    let finished = false;
    let timedOut = false;
    let loopKilled = false;

    // Disconnect stdin by default to prevent blocking on interactive TTY/OAuth prompts
    const commandLine = args.length > 0 ? `${cmd} ${args.join(" ")}` : cmd;
    const proc = spawn(commandLine, {
      cwd,
      shell: true,
      stdio: ["ignore", "pipe", "pipe"],
      env: process.env
    });

    const killProcessTree = () => {
      try {
        if (process.platform === "win32") {
          spawn("taskkill", ["/F", "/T", "/PID", proc.pid.toString()], { stdio: "ignore" });
        } else {
          proc.kill("SIGKILL");
        }
      } catch (e) {
        // ignore
      }
    };

    const timer = setTimeout(() => {
      if (!finished) {
        finished = true;
        timedOut = true;
        console.warn(`\n[WATCHDOG] ⚠️ Timeout reached (${timeoutMs}ms). Terminating process tree PID ${proc.pid}...`);
        killProcessTree();
        resolve({
          stdout,
          stderr: stderr + `\n[WATCHDOG_TIMEOUT] Process exceeded ${timeoutMs}ms and was killed.`,
          exitCode: -1,
          timedOut: true,
          killed: true
        });
      }
    }, timeoutMs);

    proc.stdout?.on("data", (chunk) => {
      const text = chunk.toString();
      stdout += text;
      if (options.onStdout) options.onStdout(text);

      if (stdout.length > maxOutputChars && !finished) {
        finished = true;
        loopKilled = true;
        clearTimeout(timer);
        console.warn(`\n[WATCHDOG] ⚠️ Output exceeded ${maxOutputChars} chars. Terminating runaway loop PID ${proc.pid}...`);
        killProcessTree();
        resolve({
          stdout,
          stderr: stderr + `\n[WATCHDOG_LOOP] Process exceeded ${maxOutputChars} chars and was killed.`,
          exitCode: -2,
          timedOut: false,
          killed: true
        });
      }
    });

    proc.stderr?.on("data", (chunk) => {
      const text = chunk.toString();
      stderr += text;
      if (options.onStderr) options.onStderr(text);
    });

    proc.on("close", (code) => {
      if (!finished) {
        finished = true;
        clearTimeout(timer);
        resolve({
          stdout,
          stderr,
          exitCode: code ?? 0,
          timedOut: false,
          killed: false
        });
      }
    });

    proc.on("error", (err) => {
      if (!finished) {
        finished = true;
        clearTimeout(timer);
        resolve({
          stdout,
          stderr: stderr + `\n[PROCESS_ERROR] ${err.message}`,
          exitCode: -3,
          timedOut: false,
          killed: false
        });
      }
    });
  });
}
