import type * as monacoEditor from "monaco-editor";
import type { Monaco } from "@flanksource/clicky-ui/monaco";

/**
 * Registers Cmd/Ctrl+Enter inside a Monaco editor.
 *
 * A window-level listener is not enough: Monaco captures keystrokes while it
 * has focus, which is exactly where the reader is when they want to run. The
 * action also puts "Run expression" in the editor's command palette, so the
 * accelerator is discoverable rather than folklore.
 *
 * `run` is read through a ref by the caller, because the action is registered
 * once per editor and must not capture the first render's closure.
 */
export function registerRunAction(
  editor: monacoEditor.editor.IStandaloneCodeEditor,
  monaco: Monaco,
  run: () => void,
): monacoEditor.IDisposable {
  return editor.addAction({
    id: "gomplate.run",
    label: "Run expression",
    keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
    run,
  });
}
