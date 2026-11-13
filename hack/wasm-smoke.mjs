import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { pathToFileURL } from "node:url";
import { gunzipSync } from "node:zlib";

const assets = resolve("web/apps/playground/public/wasm");
await import(pathToFileURL(resolve(assets, "wasm_exec.js")).href);
const go = new globalThis.Go();
globalThis.gomplateReady = () => {
  for (const [language, source, input, expected] of [
    ["cel", "count * 2", "count: 3", "6"],
    ["gotemplate", "{{ .name }}", "name: acme", "acme"],
    ["jsonpath", "$.name", "name: acme", "acme"],
    ["javascript", "1 + 2", "", "3"],
  ]) {
    const response = JSON.parse(globalThis.gomplateEvaluate(JSON.stringify({ language, source, input })));
    assert.equal(response.error, undefined, JSON.stringify(response.error));
    assert.equal(response.result, expected, language);
  }
  process.stdout.write("WASM: all four evaluators passed\n");
  process.exit(0);
};
const module = gunzipSync(await readFile(resolve(assets, "eval.wasm.gz")));
const { instance } = await WebAssembly.instantiate(module, go.importObject);
await go.run(instance);
throw new Error("WASM runtime exited before completing evaluation");
