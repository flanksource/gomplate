import EditorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
import JsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker";
import TsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

/**
 * Worker factory for clicky-ui's `MonacoProvider`; Vite bundles each `?worker`
 * import.
 *
 * gomplate's own languages need no worker -- Monarch tokenizes on the main
 * thread and completion is served from the generated spec. The JSON worker
 * backs the input editor, and the TypeScript worker backs the JavaScript
 * language, which is Monaco's own rather than one this package generates.
 */
export function getMonacoWorker(label: string): Worker {
  if (label === "typescript" || label === "javascript") return new TsWorker();
  if (label === "json") return new JsonWorker();
  return new EditorWorker();
}
