import { useCallback, useEffect, useState } from "react";

export interface PlaygroundState {
  language: string;
  source: string;
  input: string;
}

/**
 * Keeps the whole playground state in the URL hash so a snippet is shareable by
 * copying the address bar -- the thing anyone actually wants from a playground.
 *
 * Base64 rather than percent-encoding: expressions are full of `%`, `#` and
 * `&`, and a hash full of escapes is unreadable and easy to truncate by hand.
 */
export function useUrlState(initial: PlaygroundState) {
  const [state, setState] = useState<PlaygroundState>(() => decode(window.location.hash) ?? initial);

  useEffect(() => {
    const onHashChange = () => {
      const decoded = decode(window.location.hash);
      if (decoded) setState(decoded);
    };
    window.addEventListener("hashchange", onHashChange);
    return () => window.removeEventListener("hashchange", onHashChange);
  }, []);

  const update = useCallback((patch: Partial<PlaygroundState>) => {
    setState((previous) => {
      const next = { ...previous, ...patch };
      // replaceState, not a hash assignment: editing must not push a history
      // entry per keystroke.
      window.history.replaceState(null, "", `#${encode(next)}`);
      return next;
    });
  }, []);

  return [state, update] as const;
}

/**
 * The hash a link should point at to open the playground in a given state.
 *
 * The language rail is built from these, so switching language is a real
 * anchor -- middle-clickable, copyable, and reachable by Back -- rather than a
 * button that mutates state behind the URL's back.
 */
export function stateHref(state: PlaygroundState): string {
  return `#${encode(state)}`;
}

function encode(state: PlaygroundState): string {
  const json = JSON.stringify(state);
  // btoa is latin1-only; percent-encode first so non-ASCII survives.
  return btoa(unescape(encodeURIComponent(json)));
}

function decode(hash: string): PlaygroundState | null {
  const raw = hash.replace(/^#/, "");
  if (!raw) return null;
  try {
    const parsed = JSON.parse(decodeURIComponent(escape(atob(raw)))) as Partial<PlaygroundState>;
    if (typeof parsed.language !== "string" || typeof parsed.source !== "string") return null;
    return { language: parsed.language, source: parsed.source, input: parsed.input ?? "" };
  } catch {
    // A hand-edited or truncated link should drop back to the default state
    // rather than break the page.
    return null;
  }
}
