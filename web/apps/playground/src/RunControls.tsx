import { useEffect } from "react";
import { Button, Switch } from "@flanksource/clicky-ui";
import { UiPlay } from "@flanksource/clicky-ui/icons";
import type { Evaluator } from "./useEvaluator";

/** The accelerator, spelled the way the platform spells it. */
export const RUN_SHORTCUT_LABEL = isApple() ? "⌘⏎" : "Ctrl+↵";

interface RunControlsProps {
  evaluator: Evaluator;
}

export function RunControls({ evaluator }: RunControlsProps) {
  useGlobalRunShortcut(evaluator.run);

  return (
    <div className="flex items-center gap-3">
      <Switch
        checked={evaluator.autoRun}
        onChange={evaluator.setAutoRun}
        label={<span className="text-xs text-muted-foreground">Auto-run</span>}
      />
      <Button
        onClick={evaluator.run}
        // No `loadingLabel`: Button renders it whenever it is defined, not only
        // while loading, so it would replace "Run" permanently.
        loading={evaluator.pending}
        // A stale result is the case the button exists for, so make it the one
        // that stands out; otherwise it is a quiet re-run control.
        variant={evaluator.stale ? "default" : "outline"}
        size="sm"
        className="[&_svg]:size-3.5"
        title={`Run (${RUN_SHORTCUT_LABEL})`}
      >
        <UiPlay className="mr-1.5" />
        Run
        <kbd className="ml-2 rounded border border-current/25 px-1 text-[10px] leading-4 opacity-70">
          {RUN_SHORTCUT_LABEL}
        </kbd>
      </Button>
    </div>
  );
}

/**
 * Cmd/Ctrl+Enter outside the editors.
 *
 * Monaco swallows keystrokes while it has focus, so the editors register the
 * same accelerator as a Monaco action of their own (see `runAction`). This
 * covers the rest of the page.
 */
function useGlobalRunShortcut(run: () => void) {
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Enter") return;
      if (!(event.metaKey || event.ctrlKey)) return;
      event.preventDefault();
      run();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [run]);
}

function isApple(): boolean {
  if (typeof navigator === "undefined") return false;
  return /mac|iphone|ipad/i.test(navigator.userAgent);
}
