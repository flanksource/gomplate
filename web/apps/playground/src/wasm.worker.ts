import type { EvalRequest } from "./api";

interface GoRuntime {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
}

const runtime = globalThis as typeof globalThis & {
  Go: new () => GoRuntime;
  gomplateReady: () => void;
  gomplateEvaluate: (request: string) => string;
};

async function initialize() {
  const base = new URL(`${import.meta.env.BASE_URL}wasm/`, self.location.origin);
  await import(/* @vite-ignore */ new URL("wasm_exec.js", base).href);
  const go = new runtime.Go();
  const response = await fetch(new URL("eval.wasm.gz", base));
  if (!response.ok) throw new Error(`WASM download failed: ${response.status}`);
  if (!response.body) throw new Error("WASM download returned an empty body");
  const module = new Response(response.body.pipeThrough(new DecompressionStream("gzip")), {
    headers: { "Content-Type": "application/wasm" },
  });
  const { instance } = await WebAssembly.instantiateStreaming(module, go.importObject);
  await new Promise<void>((resolve, reject) => {
    runtime.gomplateReady = resolve;
    void go.run(instance).then(() => reject(new Error("Go WASM runtime exited")), reject);
  });
}

let ready: Promise<void> | undefined;

self.onmessage = async (event: MessageEvent<EvalRequest>) => {
  try {
    await (ready ??= initialize());
    postMessage(JSON.parse(runtime.gomplateEvaluate(JSON.stringify(event.data))));
  } catch (cause) {
    postMessage({ fatal: `WASM evaluation failed: ${String(cause)}` });
    self.close();
  }
};
