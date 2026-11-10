import { spawn } from "node:child_process";
import { createInterface } from "node:readline";
import type { Logger, Plugin } from "vite";

export interface EvalServerOptions {
  /** Repository root, where `go run` is invoked. */
  repoRoot: string;
  /** Loopback host the Go server listens on. */
  host: string;
  /** Preferred port; Go retains its listener or allocates another on conflict. */
  port: number;
}

/** Starts Go and resolves only after it reports its retained listener address. */
export function startEvalServer({ repoRoot, host, port, logger }: EvalServerOptions & {
  logger: Pick<Logger, "info" | "error">;
}): Promise<{ port: number; stop: () => void }> {
  const child = spawn("go", ["run", "./cmd/playground", "-addr", `${host}:${port}`, "-dynamic-port", "-watch-stdin"], {
    cwd: repoRoot,
    stdio: ["pipe", "pipe", "pipe"],
  });
  const stop = () => {
    // Closing stdin reaches the compiled grandchild as well as `go run`.
    child.stdin.end();
  };
  return new Promise((resolve, reject) => {
    let ready = false;
    const timeout = setTimeout(() => {
      stop();
      reject(new Error("eval server did not bind within 120 seconds"));
    }, 120_000);
    const lines = createInterface({ input: child.stdout });
    lines.on("line", (line) => {
      logger.info(`[gomplate] ${line}`);
      const prefix = `playground: listening on ${host}:`;
      if (!line.startsWith(prefix)) return;
      const boundPort = Number(line.slice(prefix.length));
      if (!Number.isInteger(boundPort) || boundPort < 1 || boundPort > 65535) {
        clearTimeout(timeout);
        stop();
        reject(new Error(`invalid eval server bound address: ${line}`));
        return;
      }
      ready = true;
      clearTimeout(timeout);
      resolve({ port: boundPort, stop });
    });
    child.stderr.on("data", (chunk: Buffer) => logger.error(`[gomplate] ${chunk.toString().trimEnd()}`));
    child.once("error", (error) => {
      clearTimeout(timeout);
      reject(error);
    });
    child.once("exit", (code, signal) => {
      clearTimeout(timeout);
      lines.close();
      if (!ready) reject(new Error(`eval server exited before binding (code ${code}, signal ${signal})`));
      else if (code !== 0) logger.error(`[gomplate] eval server exited (code ${code}, signal ${signal})`);
    });
  });
}

/** Runs the evaluator during development and points Vite at its bound port.
 * Set GOMPLATE_PLAYGROUND_SERVER=0 to manage the evaluator yourself.
 */
export function evalServer(options: EvalServerOptions): Plugin {
  let stopChild: (() => void) | undefined;
  const stop = () => {
    stopChild?.();
    stopChild = undefined;
    for (const signal of ["SIGINT", "SIGTERM", "exit"] as const) process.off(signal, stop);
  };
  return {
    name: "gomplate-eval-server",
    apply: "serve",
    async configureServer(server) {
      if (process.env.VITEST) return;
      if (process.env.GOMPLATE_PLAYGROUND_SERVER === "0") {
        server.config.logger.info(`[gomplate] eval server not started; expecting one on :${options.port}`);
        return;
      }
      const proxy = server.config.server.proxy?.["/api"];
      if (!proxy || typeof proxy === "string") throw new Error("eval server requires the /api proxy configuration");
      const running = await startEvalServer({ ...options, logger: server.config.logger });
      stopChild = running.stop;
      proxy.target = `http://${options.host}:${running.port}`;
      for (const signal of ["SIGINT", "SIGTERM", "exit"] as const) process.once(signal, stop);
      server.httpServer?.once("close", stop);
    },
    closeBundle: stop,
  };
}
