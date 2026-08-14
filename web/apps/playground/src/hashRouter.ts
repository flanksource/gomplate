import { createElement, useCallback, useSyncExternalStore } from "react";
import type { RouterAdapter, RenderLinkArgs } from "@flanksource/clicky-ui";

/**
 * A RouterAdapter whose location is the URL hash.
 *
 * clicky-ui's default adapter intercepts a plain left-click on a nav link and
 * navigates via `history.pushState`. That is right for a path-routed app, but
 * `pushState` does not fire `hashchange`, so a hash-routed playground would see
 * the URL change and never hear about it -- the rail would highlight while the
 * page kept showing the previous language.
 *
 * Letting the browser handle the anchor natively fixes that and costs nothing:
 * the whole state already lives in the hash, so a link is genuinely a link --
 * middle-clickable, copyable and reachable by Back.
 */
export function useHashRouter(): RouterAdapter {
  const pathname = useSyncExternalStore(
    subscribeToHash,
    () => window.location.hash,
    () => "",
  );

  const navigate = useCallback((to: string, opts?: { replace?: boolean }) => {
    const hash = to.startsWith("#") ? to : `#${to}`;
    if (opts?.replace) {
      window.history.replaceState(null, "", hash);
      // replaceState is silent, so tell the subscribers ourselves.
      window.dispatchEvent(new HashChangeEvent("hashchange"));
      return;
    }
    window.location.hash = hash;
  }, []);

  return { pathname, navigate, renderLink: anchorLink };
}

function subscribeToHash(onChange: () => void): () => void {
  window.addEventListener("hashchange", onChange);
  return () => window.removeEventListener("hashchange", onChange);
}

/** A plain anchor: no interception, so the browser fires `hashchange`. */
function anchorLink({ to, className, children, title }: RenderLinkArgs) {
  return createElement("a", { href: to, className, title }, children);
}
