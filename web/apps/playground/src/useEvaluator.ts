import { useCallback, useEffect, useRef, useState } from "react";
import { evaluate, type EvalRequest, type EvalResponse } from "./api";

const AUTO_RUN_STORAGE_KEY = "gomplate-playground:auto-run";

/** How long typing settles before an automatic run fires. */
const DEBOUNCE_MS = 250;

export interface Evaluator {
  response: EvalResponse | null;
  /** A request is in flight or a debounce is pending. */
  pending: boolean;
  /** The source or input has changed since the shown result was produced. */
  stale: boolean;
  autoRun: boolean;
  setAutoRun: (next: boolean) => void;
  /** Evaluates now, skipping the debounce. */
  run: () => void;
}

type Payload = Pick<EvalRequest, "language" | "source" | "input">;

/**
 * Runs an expression against the Go evaluator.
 *
 * Automatic evaluation is convenient for a one-line expression and a nuisance
 * for anything longer: a debounce fires mid-keystroke and reports errors for
 * half-written input. So it is a toggle, and an explicit run is always
 * available -- which is also the only way to re-run an expression whose value
 * changes on its own (`time.Now()`, `uuid.V4()`, `random.*`).
 */
export function useEvaluator(payload: Payload): Evaluator {
  const [response, setResponse] = useState<EvalResponse | null>(null);
  const [pending, setPending] = useState(false);
  const [autoRun, setAutoRunState] = useState(readAutoRun);
  const [evaluated, setEvaluated] = useState<Payload | null>(null);

  const abortRef = useRef<AbortController | undefined>(undefined);
  // The current payload, so `run` stays referentially stable: it is wired into
  // a Monaco action registered once at mount, which would otherwise capture the
  // payload from the first render forever.
  const payloadRef = useRef(payload);
  payloadRef.current = payload;

  const evaluateNow = useCallback(() => {
    const current = payloadRef.current;
    if (!current.source.trim()) {
      abortRef.current?.abort();
      setResponse(null);
      setEvaluated(current);
      setPending(false);
      return;
    }

    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    setPending(true);
    evaluate(current, controller.signal)
      .then((next) => {
        setResponse(next);
        setEvaluated(current);
        setPending(false);
      })
      .catch(() => {
        // Superseded by a newer run; that one owns the state.
      });
  }, []);

  useEffect(() => {
    if (!autoRun) return;
    const timer = setTimeout(evaluateNow, DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [payload.language, payload.source, payload.input, autoRun, evaluateNow]);

  // Switching language changes what the source even means, so show nothing
  // rather than the previous language's result.
  useEffect(() => {
    setResponse(null);
    setEvaluated(null);
  }, [payload.language]);

  const setAutoRun = useCallback(
    (next: boolean) => {
      setAutoRunState(next);
      window.localStorage.setItem(AUTO_RUN_STORAGE_KEY, String(next));
      if (next) evaluateNow();
    },
    [evaluateNow],
  );

  return {
    response,
    pending,
    stale: isStale(evaluated, payload),
    autoRun,
    setAutoRun,
    run: evaluateNow,
  };
}

function isStale(evaluated: Payload | null, current: Payload): boolean {
  if (!current.source.trim()) return false;
  if (!evaluated) return true;
  return (
    evaluated.source !== current.source ||
    evaluated.input !== current.input ||
    evaluated.language !== current.language
  );
}

function readAutoRun(): boolean {
  if (typeof window === "undefined") return true;
  // Default on: the playground should evaluate as soon as it opens.
  return window.localStorage.getItem(AUTO_RUN_STORAGE_KEY) !== "false";
}
