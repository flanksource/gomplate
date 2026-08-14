import type { GomplateSpec } from "@flanksource/gomplate-lang";

import type { EvalLanguage } from "./languages";

export interface EvalRequest {
  language: EvalLanguage;
  source: string;
  input?: string;
  leftDelim?: string;
  rightDelim?: string;
}

export interface EvalError {
  message: string;
  line?: number;
  column?: number;
}

export interface EvalResponse {
  result: string;
  value?: unknown;
  type?: string;
  error?: EvalError;
  durationMs: number;
}

/**
 * Evaluates against the Go server the dev server proxies to.
 *
 * A transport failure is surfaced as an error result rather than thrown: the
 * usual cause is the eval server not running yet, and the playground should say
 * so rather than blank out.
 */
export async function evaluate(
  request: EvalRequest,
  signal?: AbortSignal,
): Promise<EvalResponse> {
  let response: Response;
  try {
    response = await fetch("/api/eval", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
      signal,
    });
  } catch (cause) {
    if (signal?.aborted) throw cause;
    return {
      result: "",
      durationMs: 0,
      error: {
        message:
          `could not reach the eval server: ${String(cause)}\n\n` +
          "Start it with `make playground-server`, or let the dev server manage it " +
          "by unsetting GOMPLATE_PLAYGROUND_SERVER.",
      },
    };
  }

  if (!response.ok && response.status !== 400) {
    return {
      result: "",
      durationMs: 0,
      error: { message: `eval server returned ${response.status} ${response.statusText}` },
    };
  }
  return (await response.json()) as EvalResponse;
}

/**
 * Fetches the catalogue the running server can actually evaluate.
 *
 * The package ships gomplate's own catalogue baked in, which is right for this
 * app but not for a host that registers its own functions. Reading it from the
 * server instead is the path a host takes, so the playground takes it too --
 * otherwise the interesting case only ever runs somewhere else.
 *
 * Returns undefined on any failure: the baked catalogue stays in place, which
 * is exactly right when the server is simply not up yet.
 */
export async function fetchSpec(signal?: AbortSignal): Promise<GomplateSpec | undefined> {
  try {
    const response = await fetch("/api/spec", { signal });
    if (!response.ok) return undefined;
    return (await response.json()) as GomplateSpec;
  } catch {
    return undefined;
  }
}
