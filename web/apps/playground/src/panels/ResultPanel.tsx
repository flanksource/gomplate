import { Badge, Button, CodeBlock, JsonView } from "@flanksource/clicky-ui";
import type { EvalResponse } from "../api";
import { RUN_SHORTCUT_LABEL } from "../RunControls";

interface ResultPanelProps {
  response: EvalResponse | null;
  pending: boolean;
  /** The shown result predates the current source or input. */
  stale: boolean;
  onRun: () => void;
}

export function ResultPanel({ response, pending, stale, onRun }: ResultPanelProps) {
  if (!response) {
    return (
      <div className="p-4">
        <p className="text-sm text-muted-foreground">
          {pending ? "Evaluating…" : "Type an expression, then run it."}
        </p>
        {stale && !pending ? (
          <Button onClick={onRun} size="sm" className="mt-3">
            Run ({RUN_SHORTCUT_LABEL})
          </Button>
        ) : null}
      </div>
    );
  }

  if (response.error) {
    return (
      <div className="p-4">
        <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3">
          <p className="mb-2 text-xs font-medium uppercase tracking-wide text-destructive">
            {response.error.line
              ? `Error at line ${response.error.line}, column ${response.error.column}`
              : "Error"}
          </p>
          <CodeBlock source={response.error.message} bare />
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-border px-4 py-2 text-xs text-muted-foreground">
        <Badge variant="outline">{response.durationMs.toFixed(2)} ms</Badge>
        {response.type ? <Badge variant="outline">{response.type}</Badge> : null}
        {/* Without this, a result that no longer matches what is on screen is
            indistinguishable from one that does. */}
        {stale ? (
          <button
            type="button"
            onClick={onRun}
            className="ml-auto rounded border border-amber-500/40 bg-amber-500/10 px-2 py-0.5 text-amber-700 hover:bg-amber-500/20 dark:text-amber-400"
            title={`Run (${RUN_SHORTCUT_LABEL})`}
          >
            Out of date — run again
          </button>
        ) : null}
      </div>

      <div
        className={`min-h-0 flex-1 overflow-auto p-4 ${stale ? "opacity-50" : ""}`}
      >
        {/* The rendered string is what a gomplate caller receives. A structured
            result is more readable as a tree, so hand those to JsonView and
            keep the raw rendering below it. */}
        {isStructured(response.value) ? (
          <>
            <JsonView data={response.value} defaultOpenDepth={3} />
            <details className="mt-4">
              <summary className="cursor-pointer text-xs text-muted-foreground">
                Rendered string
              </summary>
              <div className="mt-2">
                <CodeBlock source={response.result} bare copyable={false} />
              </div>
            </details>
          </>
        ) : (
          <CodeBlock
            source={response.result || "(empty)"}
            language={languageOf(response.result)}
            copyable
          />
        )}
      </div>
    </div>
  );
}

function isStructured(value: unknown): boolean {
  return typeof value === "object" && value !== null;
}

/**
 * Templates render whole documents, so highlight the output when it is
 * recognisably YAML or JSON rather than showing a wall of grey.
 */
function languageOf(result: string): string | undefined {
  const trimmed = result.trim();
  if (!trimmed) return undefined;
  if (trimmed.startsWith("{") || trimmed.startsWith("[")) return "json";
  if (/^[A-Za-z_][\w.-]*:\s/m.test(trimmed)) return "yaml";
  return undefined;
}
