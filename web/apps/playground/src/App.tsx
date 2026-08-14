import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  AppShell,
  Combobox,
  DensityProvider,
  Tabs,
  ThemeProvider,
} from "@flanksource/clicky-ui";
import type { AppShellNavSection } from "@flanksource/clicky-ui";
import { RouterProvider } from "@flanksource/clicky-ui/rpc";
import { MonacoEditor, MonacoProvider } from "@flanksource/clicky-ui/monaco";
import type { Monaco } from "@flanksource/clicky-ui/monaco";
import { mergeSpec, registerGomplateLanguages, spec } from "@flanksource/gomplate-lang";
import type { GomplateSpec, RegisteredLanguages } from "@flanksource/gomplate-lang";
import * as monacoEditor from "monaco-editor";

import { fetchSpec } from "./api";
import type { EvalResponse } from "./api";
import { defaultExample, examplesFor } from "./examples";
import { functionCatalogue, functionCatalogueFlavour } from "./functionCatalogue";
import { LANGUAGES, SECTIONS, languageById } from "./languages";
import { getMonacoWorker } from "./monaco-setup";
import { GraphPanel } from "./panels/GraphPanel";
import { ResultPanel } from "./panels/ResultPanel";
import { SpecPanel } from "./panels/SpecPanel";
import { TokensPanel } from "./panels/TokensPanel";
import { RunControls } from "./RunControls";
import { registerRunAction } from "./runAction";
import { useEvaluator } from "./useEvaluator";
import { useParsedInput } from "./useParsedInput";
import { VerticalSplit, rowsToPaneHeight } from "./VerticalSplit";
import { useHashRouter } from "./hashRouter";
import { useEditorTheme } from "./useEditorTheme";
import { stateHref, useUrlState } from "./useUrlState";

const SOURCE_MODEL_PATH = "inmemory://playground/source";
const INPUT_MODEL_PATH = "inmemory://playground/input.yaml";
/** Owner of the markers this app sets, so clearing them leaves other providers' alone. */
const MARKER_OWNER = "gomplate";

export function App() {
  const router = useHashRouter();

  return (
    <ThemeProvider>
      <DensityProvider>
        <RouterProvider adapter={router}>
          <MonacoProvider getWorker={getMonacoWorker}>
            <Playground />
          </MonacoProvider>
        </RouterProvider>
      </DensityProvider>
    </ThemeProvider>
  );
}

function Playground() {
  const initial = defaultExample("cel");
  const [state, setState] = useUrlState({
    language: "cel",
    source: initial.source,
    input: initial.input,
  });

  const language = useMemo(() => languageById(state.language), [state.language]);
  const [outputTab, setOutputTab] = useState("result");

  const evaluator = useEvaluator({
    language: language.evalLanguage,
    source: state.source,
    input: state.input,
  });

  // Completion reads the document through a ref: languages are registered once,
  // before the first editor mounts, while the input keeps being edited after.
  const parsedInput = useParsedInput(state.input);
  const environmentRef = useRef<unknown>(undefined);
  environmentRef.current = parsedInput.value;

  // The catalogue the server can actually evaluate, which for a host binary is
  // wider than the one this package ships. Everything that reads a catalogue
  // reads this one, so the Functions tab counts what completion offers.
  const [served, setServed] = useState<GomplateSpec>();
  const servedRef = useRef<GomplateSpec>(undefined);
  servedRef.current = served;
  const activeSpec = useMemo(() => mergeSpec(spec, served), [served]);

  const catalogue = functionCatalogue(language.evalLanguage, activeSpec);
  const catalogueFlavour = functionCatalogueFlavour(language.evalLanguage);

  // Registration must happen before the first model is created, or Monaco
  // resolves the language id to plaintext and never revisits it. Which of the
  // two lands first is a race -- Monaco loads slowly, the fetch returns fast --
  // so both paths apply the catalogue: `beforeMount` reads whatever has already
  // arrived, and the effect updates whatever is already registered.
  const languages = useRef<RegisteredLanguages | null>(null);
  const registerLanguages = useCallback((monaco: Monaco) => {
    languages.current = registerGomplateLanguages(monaco, {
      environment: () => environmentRef.current,
      spec: servedRef.current,
    });
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void fetchSpec(controller.signal).then((spec) => {
      if (!controller.signal.aborted) setServed(spec);
    });
    return () => controller.abort();
  }, []);

  useEffect(() => {
    if (served) languages.current?.setSpec(served);
  }, [served]);

  const [sourceModel, setSourceModel] = useState<monacoEditor.editor.ITextModel | null>(null);
  useMarkers(evaluator.response, sourceModel);
  const applyTheme = useEditorTheme();

  useEffect(() => {
    if (outputTab === "spec" && !catalogueFlavour) setOutputTab("result");
  }, [catalogueFlavour, outputTab]);

  // The Monaco action is registered once per editor, so it has to reach the
  // current `run` through a ref rather than capturing it.
  const runRef = useRef(evaluator.run);
  runRef.current = evaluator.run;
  const sourceEditor = useRef<Parameters<typeof registerRunAction>[0] | null>(null);
  const onEditorMount = useCallback(
    (editor: Parameters<typeof registerRunAction>[0], monaco: Monaco) => {
      applyTheme();
      // Hovers are content widgets, and Monaco renders them inside its own DOM
      // unless told otherwise. An error on line 2 has no room above it inside a
      // short pane, so the hover -- the only place the marker's message is
      // readable -- was drawn over the editor's top edge and clipped away by the
      // container's `overflow-hidden`. Fixed positioning lets it escape to the
      // viewport, which is what every editor embedded in a pane does.
      editor.updateOptions({
        fixedOverflowWidgets: true,
        // Monaco's defaults size the gutter for a source file: five digits of
        // line number and a decorations strip nothing here draws in. A
        // playground document is tens of lines, so three digits is generous and
        // the strip is dead space. What is reclaimed pays for the glyph margin
        // below several times over.
        lineNumbersMinChars: 3,
        // Enough to keep the number off the code; the default 10 exists to hold
        // decorations, and folding would add 16 more for controls a document
        // this short has no use for.
        lineDecorationsWidth: 6,
        folding: false,
      });
      registerRunAction(editor, monaco, () => runRef.current());
      const model = editor.getModel();
      if (model?.uri.toString() === SOURCE_MODEL_PATH) {
        sourceEditor.current = editor;
        setSourceModel(model);
        // The glyph margin is where the error icon goes, and it is also the only
        // gutter column Monaco will show a hover over. Reserved up front rather
        // than toggled with the error, so the text does not jump sideways
        // mid-keystroke as evaluation succeeds and fails.
        editor.updateOptions({ glyphMargin: true });
      }
    },
    [applyTheme],
  );

  /** Writes a path from the object graph where the reader last left the caret. */
  const insertExpression = useCallback((expression: string) => {
    const editor = sourceEditor.current;
    const selection = editor?.getSelection();
    if (!editor || !selection) return;
    editor.executeEdits("object-graph", [
      { range: selection, text: expression, forceMoveMarkers: true },
    ]);
    editor.focus();
  }, []);

  const examples = examplesFor(state.language);

  return (
    <AppShell
      brand={<Brand />}
      navSections={navSections(state.language)}
      collapsedStorageKey="gomplate-playground:rail"
      sidebarFooter={<CatalogueSummary />}
      bodyHeader={
        <div>
          <h1 className="text-sm font-semibold">{language.label}</h1>
          <p className="text-xs text-muted-foreground">{language.description}</p>
        </div>
      }
      bodyActions={
        <div className="flex items-center gap-3">
          {examples.length > 0 ? (
            <Combobox
              options={examples.map((example) => ({
                value: example.name,
                label: example.name,
              }))}
              value=""
              allowCustomValue={false}
              placeholder="Load an example…"
              ariaLabel="Load an example"
              className="w-64"
              onChange={(name) => {
                const example = examples.find((candidate) => candidate.name === name);
                if (example) setState({ source: example.source, input: example.input });
              }}
            />
          ) : null}
          <RunControls evaluator={evaluator} />
        </div>
      }
      bodySidebar={
        <VerticalSplit
          storageKey="gomplate-playground:editor-split"
          defaultTopHeight={rowsToPaneHeight(language.editorRows)}
          top={
            <EditorPane label="Expression">
              <MonacoEditor
                value={state.source}
                onChange={(source) => setState({ source })}
                language={language.editorLanguage}
                path={SOURCE_MODEL_PATH}
                height="100%"
                beforeMount={registerLanguages}
                onMount={onEditorMount}
              />
            </EditorPane>
          }
          bottom={
            <EditorPane
              label="Input"
              hint={parsedInput.error ?? "YAML or JSON"}
              hintTone={parsedInput.error ? "error" : "muted"}
            >
              <MonacoEditor
                value={state.input}
                onChange={(input) => setState({ input })}
                language="yaml"
                path={INPUT_MODEL_PATH}
                height="100%"
                onMount={onEditorMount}
              />
            </EditorPane>
          }
        />
      }
      bodySplit={52}
      contentClassName="p-0"
    >
      <div className="flex h-full flex-col">
        <div className="border-b border-border px-4">
          <Tabs
            tabs={[
              { id: "result", label: "Result" },
              { id: "graph", label: "Object graph" },
              { id: "tokens", label: "Tokens" },
              ...(catalogueFlavour
                ? [{ id: "spec", label: "Functions", count: catalogue.length }]
                : []),
            ]}
            value={outputTab}
            onChange={setOutputTab}
          />
        </div>

        <div className="min-h-0 flex-1">
          {outputTab === "result" ? (
            <ResultPanel
              response={evaluator.response}
              pending={evaluator.pending}
              stale={evaluator.stale}
              onRun={evaluator.run}
            />
          ) : null}
          {outputTab === "graph" ? (
            <GraphPanel
              document={parsedInput.value}
              languageId={language.editorLanguage}
              onInsert={insertExpression}
            />
          ) : null}
          {outputTab === "tokens" ? (
            <TokensPanel source={state.source} languageId={language.editorLanguage} />
          ) : null}
          {outputTab === "spec" && catalogueFlavour ? (
            <SpecPanel flavour={catalogueFlavour} spec={activeSpec} />
          ) : null}
        </div>
      </div>
    </AppShell>
  );
}

function Brand() {
  return (
    <div className="flex flex-col leading-tight">
      <span className="text-sm font-semibold">gomplate</span>
      <span className="text-[10px] uppercase tracking-wide opacity-60">playground</span>
    </div>
  );
}

/**
 * The catalogue sizes double as a sanity check: an empty namespace list means
 * the generated spec did not load.
 */
function CatalogueSummary() {
  return (
    <div className="px-1 text-[11px] leading-relaxed opacity-70">
      <div>{spec.cel.functions.length} CEL functions</div>
      <div>{spec.gotemplate.functions.length} template functions</div>
    </div>
  );
}

/**
 * Builds the rail. Each language is an anchor to the hash that selects it, so
 * the browser treats a language switch as navigation; `useUrlState` picks the
 * change up from `hashchange`.
 */
function navSections(activeId: string): AppShellNavSection[] {
  return SECTIONS.map((section) => ({
    label: section,
    items: LANGUAGES.filter((language) => language.section === section).map((language) => {
      const example = defaultExample(language.id);
      return {
        key: language.id,
        label: language.label,
        icon: language.icon,
        active: language.id === activeId,
        to: stateHref({
          language: language.id,
          source: example.source,
          input: example.input,
        }),
      };
    }),
  }));
}

function EditorPane({
  label,
  hint,
  hintTone = "muted",
  children,
}: {
  label: string;
  hint?: string;
  /** A parse failure belongs next to the text that caused it, not in a panel. */
  hintTone?: "muted" | "error";
  children: React.ReactNode;
}) {
  return (
    <section className="flex h-full min-h-0 flex-col bg-background">
      <header className="flex h-[33px] shrink-0 items-baseline gap-2 px-4 py-2">
        <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          {label}
        </h2>
        {hint ? (
          <span
            className={`truncate text-[11px] ${
              hintTone === "error" ? "text-destructive" : "text-muted-foreground/70"
            }`}
            title={hint}
          >
            {hint}
          </span>
        ) : null}
      </header>
      <div className="min-h-0 flex-1 [&_[data-slot=monaco-editor]]:h-full [&_[data-slot=monaco-editor]]:rounded-none [&_[data-slot=monaco-editor]]:border-0 [&_[data-slot=monaco-editor]]:border-t">
        {children}
      </div>
    </section>
  );
}

/**
 * Mirrors an evaluation error onto the editor as a marker, so the position the
 * Go side reported is underlined where the mistake is -- with Monaco's own
 * squiggle, hover and F8 navigation -- rather than only named in the result
 * panel.
 *
 * The squiggle alone underlines one token in a pane the reader may not be
 * looking at, so the same line is also called out in the gutter -- an icon and a
 * red line number -- and the icon carries the message too, so hovering either
 * the gutter or the token explains the failure.
 *
 * `model` is passed rather than looked up because Monaco loads asynchronously:
 * the first evaluation regularly finishes before the editor exists, and a
 * lookup that missed would leave that error unmarked until the next run.
 */
function useMarkers(response: EvalResponse | null, model: monacoEditor.editor.ITextModel | null) {
  const decorations = useRef<string[]>([]);

  useEffect(() => {
    if (!model || model.isDisposed()) return;

    const error = response?.error;
    if (!error?.line) {
      monacoEditor.editor.setModelMarkers(model, MARKER_OWNER, []);
      decorations.current = model.deltaDecorations(decorations.current, []);
      return;
    }

    const line = Math.min(Math.max(error.line, 1), model.getLineCount());
    const column = Math.max(error.column ?? 1, 1);
    monacoEditor.editor.setModelMarkers(model, MARKER_OWNER, [
      {
        severity: monacoEditor.MarkerSeverity.Error,
        message: error.message,
        startLineNumber: line,
        startColumn: column,
        endLineNumber: line,
        endColumn: markerEndColumn(model, line, column),
      },
    ]);
    decorations.current = model.deltaDecorations(decorations.current, [
      {
        range: new monacoEditor.Range(line, 1, line, 1),
        options: {
          glyphMarginClassName: "playground-error-glyph",
          // A fenced block, not bare text: the message is full of `<...>` and
          // `{}`, which the hover's markdown renderer would otherwise eat.
          glyphMarginHoverMessage: { value: "```\n" + error.message + "\n```" },
          lineNumberClassName: "playground-error-line-number",
          // Survives edits to the line rather than growing to swallow what the
          // reader types next to it.
          stickiness: monacoEditor.editor.TrackedRangeStickiness.NeverGrowsWhenTypingAtEdges,
        },
      },
    ]);
  }, [response, model]);
}

/**
 * Underlines the whole token at the reported position. A compiler points at one
 * character; a one-character squiggle is easy to miss and unpleasant to hover.
 * Monaco clamps a range that runs past the end of the line, so a position at
 * end-of-input still underlines something.
 */
function markerEndColumn(
  model: monacoEditor.editor.ITextModel,
  line: number,
  column: number,
): number {
  const word = model.getWordAtPosition({ lineNumber: line, column });
  return Math.max(word?.endColumn ?? 0, column + 1);
}
