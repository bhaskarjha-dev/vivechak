import { spawn } from "node:child_process";

/**
 * Executes a subprocess with an ironclad watchdog that guarantees zero stalling.
 * Uses native execution (shell: false) by default to prevent argument mangling
 * and Windows cmd.exe pipe/special character splitting, falling back gracefully
 * if shell execution is required on Windows for batch scripts (.cmd/.bat).
 *
 * @param {string} cmd The executable or command.
 * @param {string[]} args Command line arguments.
 * @param {object} options Options including cwd, timeoutMs, stdio, shell.
 * @returns {Promise<{ stdout: string, stderr: string, exitCode: number, timedOut: boolean, killed: boolean }>}
 */
export function execWithWatchdog(cmd, args = [], options = {}) {
  const timeoutMs = options.timeoutMs || 90000; // 90 second default
  const maxOutputChars = options.maxOutputChars || 60000;
  const cwd = options.cwd || process.cwd();
  const isWin = process.platform === "win32";

  return new Promise((resolve) => {
    let stdout = "";
    let stderr = "";
    let finished = false;
    let timedOut = false;
    let loopKilled = false;
    let timer = null;
    let proc = null;

    const killProcessTree = () => {
      try {
        if (proc?.pid) {
          if (isWin) {
            spawn("taskkill", ["/F", "/T", "/PID", proc.pid.toString()], { stdio: "ignore" });
          } else {
            proc.kill("SIGKILL");
          }
        }
      } catch (e) {
        // ignore
      }
    };

    const setupProcEvents = (p) => {
      timer = setTimeout(() => {
        if (!finished) {
          finished = true;
          timedOut = true;
          console.warn(`\n[WATCHDOG] ⚠️ Timeout reached (${timeoutMs}ms). Terminating process tree PID ${p.pid}...`);
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

      p.stdout?.on("data", (chunk) => {
        const text = chunk.toString();
        stdout += text;
        if (options.onStdout) options.onStdout(text);

        if (stdout.length > maxOutputChars && !finished) {
          finished = true;
          loopKilled = true;
          clearTimeout(timer);
          console.warn(`\n[WATCHDOG] ⚠️ Output exceeded ${maxOutputChars} chars. Terminating runaway loop PID ${p.pid}...`);
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

      p.stderr?.on("data", (chunk) => {
        const text = chunk.toString();
        stderr += text;
        if (options.onStderr) options.onStderr(text);
      });

      p.on("close", (code) => {
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
    };

    const spawnWithFallback = () => {
      const initialShell = options.shell ?? false;
      try {
        proc = spawn(cmd, args, {
          cwd,
          shell: initialShell,
          windowsHide: true,
          stdio: ["ignore", "pipe", "pipe"],
          env: process.env
        });

        proc.on("error", (err) => {
          // If on Windows and initial spawn failed with EINVAL (batch script) or ENOENT (needs PATHEXT shell resolution), retry with shell: true
          if (isWin && !initialShell && (err.code === "EINVAL" || err.code === "ENOENT")) {
            if (timer) clearTimeout(timer);
            try {
              const cmdLine = args.length > 0 ? `${cmd} ${args.map((a) => (typeof a === "string" && (a.includes(" ") || a.includes("\n")) ? JSON.stringify(a) : a)).join(" ")}` : cmd;
              proc = spawn(cmdLine, {
                cwd,
                shell: true,
                windowsHide: true,
                stdio: ["ignore", "pipe", "pipe"],
                env: process.env
              });
              setupProcEvents(proc);
              proc.on("error", (err2) => {
                if (!finished) {
                  finished = true;
                  if (timer) clearTimeout(timer);
                  resolve({
                    stdout,
                    stderr: stderr + `\n[PROCESS_ERROR] ${err2.message}`,
                    exitCode: -3,
                    timedOut: false,
                    killed: false
                  });
                }
              });
              return;
            } catch (fallbackErr) {
              // fall through
            }
          }

          if (!finished) {
            finished = true;
            if (timer) clearTimeout(timer);
            resolve({
              stdout,
              stderr: stderr + `\n[PROCESS_ERROR] ${err.message}`,
              exitCode: -3,
              timedOut: false,
              killed: false
            });
          }
        });

        setupProcEvents(proc);
      } catch (err) {
        if (isWin && !initialShell) {
          try {
            const cmdLine = args.length > 0 ? `${cmd} ${args.map((a) => (typeof a === "string" && (a.includes(" ") || a.includes("\n")) ? JSON.stringify(a) : a)).join(" ")}` : cmd;
            proc = spawn(cmdLine, {
              cwd,
              shell: true,
              windowsHide: true,
              stdio: ["ignore", "pipe", "pipe"],
              env: process.env
            });
            setupProcEvents(proc);
            return;
          } catch (e) {
            // fall through
          }
        }
        resolve({
          stdout: "",
          stderr: `[SPAWN_ERROR] ${err.message}`,
          exitCode: -3,
          timedOut: false,
          killed: false
        });
      }
    };

    spawnWithFallback();
  });
}
