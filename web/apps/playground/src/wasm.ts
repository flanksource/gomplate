import type { EvalRequest, EvalResponse } from "./api";

let worker: Worker | undefined;
const TIMEOUT_MS = 10_000;

export function evaluateWasm(request: EvalRequest, signal?: AbortSignal): Promise<EvalResponse> {
  signal?.throwIfAborted();
  const current = worker ??= new Worker(new URL("./wasm.worker.ts", import.meta.url), { type: "module" });

  return new Promise((resolve, reject) => {
    const cleanup = () => {
      clearTimeout(timer);
      current.removeEventListener("message", onMessage);
      current.removeEventListener("error", onError);
      signal?.removeEventListener("abort", onAbort);
    };
    const terminate = () => {
      current.terminate();
      if (worker === current) worker = undefined;
    };
    const fail = (message: string) => {
      cleanup();
      terminate();
      resolve({ result: "", durationMs: 0, error: { message } });
    };
    const onMessage = (event: MessageEvent<EvalResponse | { fatal: string }>) => {
      if ("fatal" in event.data) {
        fail(event.data.fatal);
        return;
      }
      cleanup();
      resolve(event.data);
    };
    const onError = (event: ErrorEvent) => fail(`WASM evaluation failed: ${event.message}`);
    const onAbort = () => {
      cleanup();
      terminate();
      reject(signal?.reason);
    };
    const timer = setTimeout(() => fail("WASM evaluation exceeded 10 seconds"), TIMEOUT_MS);
    current.addEventListener("message", onMessage);
    current.addEventListener("error", onError);
    signal?.addEventListener("abort", onAbort, { once: true });
    current.postMessage(request);
  });
}
