import { useCallback, useEffect } from "react";
import * as monaco from "monaco-editor";
import { GOMPLATE_DARK_THEME, GOMPLATE_LIGHT_THEME } from "@flanksource/gomplate-lang";

/**
 * Applies the gomplate colour themes and keeps them in step with clicky-ui's
 * theme switcher.
 *
 * `monaco.editor.setTheme` is global rather than per-editor, and
 * `@monaco-editor/react` calls it whenever an editor mounts, using the `theme`
 * prop clicky-ui hardcodes to `light`/`vs-dark`. So setting the theme once is
 * not enough: every editor that mounts afterwards resets it. Re-asserting from
 * each editor's `onMount` runs after that call and is deterministic, where a
 * plain effect would race the editor's lazy mount.
 *
 * clicky-ui's `MonacoEditor` gaining a `theme` prop makes this unnecessary;
 * until then this keeps the playground correct against the published package.
 */
export function useEditorTheme() {
  const applyTheme = useCallback(() => {
    const dark = document.documentElement.getAttribute("data-theme") === "dark";
    monaco.editor.setTheme(dark ? GOMPLATE_DARK_THEME : GOMPLATE_LIGHT_THEME);
  }, []);

  useEffect(() => {
    applyTheme();

    // clicky-ui's ThemeProvider writes `data-theme` on <html> and emits no
    // event, so observe the attribute.
    const observer = new MutationObserver(applyTheme);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["data-theme"],
    });
    return () => observer.disconnect();
  }, [applyTheme]);

  return applyTheme;
}
