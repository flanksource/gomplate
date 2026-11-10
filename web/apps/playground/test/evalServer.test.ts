import { EventEmitter } from "node:events";
import { PassThrough } from "node:stream";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ViteDevServer } from "vite";

vi.mock("node:child_process", () => ({ spawn: vi.fn() }));
import { spawn } from "node:child_process";
import { evalServer, startEvalServer } from "../plugins/eval-server";

const logger = { info: vi.fn(), error: vi.fn() };

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllEnvs();
  vi.useRealTimers();
});

function childProcess() {
  return Object.assign(new EventEmitter(), {
    stdout: new PassThrough(), stderr: new PassThrough(), stdin: new PassThrough(),
  });
}

describe("evaluation server startup", () => {
  it("configures the proxy with the bound port and stops on Vite shutdown", async () => {
    vi.stubEnv("VITEST", "");
    vi.stubEnv("GOMPLATE_PLAYGROUND_SERVER", "1");
    const child = childProcess();
    vi.mocked(spawn).mockReturnValue(child as unknown as ReturnType<typeof spawn>);
    const proxy = { target: "http://127.0.0.1:8321" };
    const httpServer = new EventEmitter();
    const plugin = evalServer({ repoRoot: "/repo", host: "127.0.0.1", port: 8321 });
    if (typeof plugin.configureServer !== "function") throw new Error("expected a configureServer hook");
    const configured = plugin.configureServer.call({} as never, {
      config: { logger, server: { proxy: { "/api": proxy } } }, httpServer,
    } as unknown as ViteDevServer);
    try {
      child.stdout.write("playground: listening on 127.0.0.1:49153\n");
      await configured;
      expect(proxy.target).toBe("http://127.0.0.1:49153");
    } finally {
      httpServer.emit("close");
    }
    expect(child.stdin.writableEnded).toBe(true);
    child.emit("exit", 0);
  });

  it("waits for the bound address, including split stdout chunks", async () => {
    const child = childProcess();
    vi.mocked(spawn).mockReturnValue(child as unknown as ReturnType<typeof spawn>);
    const startup = startEvalServer({ repoRoot: "/repo", host: "127.0.0.1", port: 8321, logger });
    child.stdout.write("playground: listening on 127.0.0.1:");
    child.stdout.write("49153\n");
    const server = await startup;
    expect(server.port).toBe(49153);
    expect(spawn).toHaveBeenCalledWith("go", ["run", "./cmd/playground", "-addr", "127.0.0.1:8321", "-dynamic-port", "-watch-stdin"], expect.any(Object));
    server.stop();
    expect(child.stdin.writableEnded).toBe(true);
  });

  it("rejects startup when the process exits before binding", async () => {
    const child = childProcess();
    vi.mocked(spawn).mockReturnValue(child as unknown as ReturnType<typeof spawn>);
    const startup = startEvalServer({ repoRoot: "/repo", host: "127.0.0.1", port: 8321, logger });
    const rejected = expect(startup).rejects.toThrow("before binding");
    child.emit("exit", 1);
    await rejected;
  });

  it("rejects spawn errors", async () => {
    const child = childProcess();
    vi.mocked(spawn).mockReturnValue(child as unknown as ReturnType<typeof spawn>);
    const startup = startEvalServer({ repoRoot: "/repo", host: "127.0.0.1", port: 8321, logger });
    const rejected = expect(startup).rejects.toThrow("spawn failed");
    child.emit("error", new Error("spawn failed"));
    await rejected;
  });

  it("rejects invalid bound addresses and closes the child's input", async () => {
    const child = childProcess();
    vi.mocked(spawn).mockReturnValue(child as unknown as ReturnType<typeof spawn>);
    const startup = startEvalServer({ repoRoot: "/repo", host: "127.0.0.1", port: 8321, logger });
    const rejected = expect(startup).rejects.toThrow("invalid eval server bound address");
    child.stdout.write("playground: listening on 127.0.0.1:0\n");
    await rejected;
    expect(child.stdin.writableEnded).toBe(true);
  });

  it("rejects a startup timeout and closes the child's input", async () => {
    vi.useFakeTimers();
    const child = childProcess();
    vi.mocked(spawn).mockReturnValue(child as unknown as ReturnType<typeof spawn>);
    const startup = startEvalServer({ repoRoot: "/repo", host: "127.0.0.1", port: 8321, logger });
    const rejected = expect(startup).rejects.toThrow("did not bind within 120 seconds");
    await vi.advanceTimersByTimeAsync(120_000);
    await rejected;
    expect(child.stdin.writableEnded).toBe(true);
  });
});
