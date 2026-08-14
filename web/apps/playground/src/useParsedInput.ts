import { useMemo, useRef } from "react";
import { parse } from "yaml";

export interface ParsedInput {
  /** The last document that parsed, so editing one does not blank the other. */
  value: unknown;
  /** The parse failure for the text as it stands, if there is one. */
  error: string | null;
}

/**
 * Parses the input pane on the client.
 *
 * The eval response carries only the result — it never echoes the environment
 * back — and completion has to work while the document is still being typed,
 * before any round trip. JSON is valid YAML, so one parser covers both, the same
 * way the Go side's `parseInput` does.
 *
 * A half-typed document is a parse error on most keystrokes. Holding the last
 * good value through those keeps the completion list and the shape tree from
 * flickering out from under the reader.
 */
export function useParsedInput(input: string): ParsedInput {
  const lastGood = useRef<unknown>(undefined);

  return useMemo(() => {
    if (input.trim() === "") {
      lastGood.current = undefined;
      return { value: undefined, error: null };
    }
    try {
      const value: unknown = parse(input);
      lastGood.current = value;
      return { value, error: null };
    } catch (error) {
      return { value: lastGood.current, error: (error as Error).message };
    }
  }, [input]);
}
