import { spawn, type ChildProcess } from "node:child_process";
import type { Plugin } from "vite";

export interface EvalServerOptions {
  /** Repository root, where `go run` is invoked. */
  repoRoot: string;
  /** Port the Go server listens on. */
  port: number;
}

/**
 * Runs `go run ./cmd/playground` alongside the dev server, so evaluation goes
 * through the real gomplate engine rather than a reimplementation in the
 * browser.
 *
 * Set `GOMPLATE_PLAYGROUND_SERVER=0` to manage the process yourself -- useful
 * when attaching a debugger, or when iterating on Go code that would otherwise
 * be recompiled on every Vite restart.
 */
export function evalServer({ repoRoot, port }: EvalServerOptions): Plugin {
  let child: ChildProcess | undefined;

  const stop = () => {
    if (!child || child.killed) return;
    child.kill("SIGTERM");
    child = undefined;
  };

  return {
    name: "gomplate-eval-server",
    apply: "serve",

    configureServer(server) {
      // Vitest stands up a Vite server of its own; compiling and running the Go
      // binary for a unit test would be minutes of nothing useful.
      if (process.env.VITEST) return;
      if (process.env.GOMPLATE_PLAYGROUND_SERVER === "0") {
        server.config.logger.info(
          `[gomplate] eval server not started; expecting one on :${port}`,
        );
        return;
      }

      child = spawn("go", ["run", "./cmd/playground", "-addr", `:${port}`], {
        cwd: repoRoot,
        stdio: ["ignore", "pipe", "pipe"],
      });

      child.stdout?.on("data", (chunk: Buffer) => {
        server.config.logger.info(`[gomplate] ${chunk.toString().trimEnd()}`);
      });
      child.stderr?.on("data", (chunk: Buffer) => {
        // `go run` reports compile errors here; they are the single most useful
        // thing to surface, so do not swallow them.
        server.config.logger.error(`[gomplate] ${chunk.toString().trimEnd()}`);
      });
      child.on("exit", (code) => {
        if (code !== 0 && code !== null) {
          server.config.logger.error(`[gomplate] eval server exited with code ${code}`);
        }
        child = undefined;
      });

      // `go run` leaves the compiled binary as a grandchild, so kill on every
      // way the dev server can end rather than relying on process-group death.
      for (const signal of ["SIGINT", "SIGTERM", "exit"] as const) {
        process.once(signal, stop);
      }
      server.httpServer?.once("close", stop);
    },

    closeBundle: stop,
  };
}
