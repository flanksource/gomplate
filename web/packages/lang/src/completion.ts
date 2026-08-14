import type * as monaco from "monaco-editor";
import type { GomplateSpec, Monaco, SpecFunction, SpecMacro } from "./types";
import { functionDocumentation, macroDocumentation } from "./hover";
import { childEntries, pathExpression, pathFlavour, resolvePath } from "./environment";
import { environmentPrefixAt } from "./prefix";

/** Supplies the document expressions are evaluated against, on every request. */
export type EnvironmentSource = () => unknown;

export interface CompletionOptions {
  /** The catalogue to complete from — gomplate's, or a host's merged over it. */
  spec: GomplateSpec;
  environment?: EnvironmentSource | undefined;
}

/**
 * Registers completion for one language.
 *
 * The catalogue half is built once per registration, so a host that supplies
 * its own spec re-registers rather than mutating in place. The document half is
 * rebuilt per request, because the document is being edited alongside the
 * expression.
 */
export function registerCompletion(
  monaco: Monaco,
  languageId: string,
  options: CompletionOptions,
) {
  return monaco.languages.registerCompletionItemProvider(
    languageId,
    completionProvider(monaco, languageId, options),
  );
}

/**
 * The provider `registerCompletion` installs, exposed so it can be driven
 * directly by a test over a real model rather than through Monaco's registry.
 */
export function completionProvider(
  monaco: Monaco,
  languageId: string,
  { spec, environment }: CompletionOptions,
) {
  const items = catalogueItems(monaco, languageId, spec);
  return {
    // A dot must retrigger, or `k8s.` offers nothing until another key is hit.
    triggerCharacters: ["."],
    provideCompletionItems(model: monaco.editor.ITextModel, position: monaco.Position) {
      const suggestions: monaco.languages.CompletionItem[] = environmentItems(
        monaco,
        model,
        position,
        languageId,
        environment,
      );
      const range = wordRange(model, position);
      for (const item of items) suggestions.push({ ...item, range });
      return { suggestions };
    },
  };
}

/**
 * A completion item without its range. The range depends on the cursor, so it
 * is attached per request while the rest of the item is built once.
 */
type Item = Omit<monaco.languages.CompletionItem, "range">;

function catalogueItems(monaco: Monaco, languageId: string, spec: GomplateSpec): Item[] {
  switch (pathFlavour(languageId)) {
    case "cel":
      return celCompletionItems(monaco, spec);
    case "gotemplate":
      return goTemplateCompletionItems(monaco, spec);
    default:
      // JSONPath's dialect has no catalogue to generate from; its completions
      // come entirely from the document.
      return [];
  }
}

/**
 * The keys reachable from the cursor's position in the document.
 *
 * Each item replaces the whole path typed so far with a freshly rendered one,
 * rather than appending to it, so what lands in the editor is always valid in
 * that language — quoting an awkward key, or dropping the suggestion entirely
 * where the language cannot express the path.
 */
function environmentItems(
  monaco: Monaco,
  model: monaco.editor.ITextModel,
  position: monaco.Position,
  languageId: string,
  environment: EnvironmentSource | undefined,
): monaco.languages.CompletionItem[] {
  if (!environment) return [];
  const document = environment();
  if (document === undefined || document === null) return [];

  const prefix = environmentPrefixAt(model, position, languageId);
  if (!prefix) return [];

  const parent = resolvePath(document, prefix.segments);
  if (parent === undefined) return [];

  const kinds = monaco.languages.CompletionItemKind;
  const range = {
    startLineNumber: position.lineNumber,
    endLineNumber: position.lineNumber,
    startColumn: prefix.startColumn,
    endColumn: prefix.endColumn,
  };

  // Monaco filters a candidate against the model text from the range start to
  // the cursor, so `filterText` has to continue what was typed rather than
  // repeat the rendered path: `pod.items[0]` does not match `pod.items.`, and
  // the subscripted keys would be filtered straight back out.
  const typedHead = prefix.typed.slice(0, prefix.typed.length - prefix.leaf.length);

  const items: monaco.languages.CompletionItem[] = [];
  for (const entry of childEntries(parent)) {
    const insertText = pathExpression(languageId, [...prefix.segments, entry.segment]);
    if (insertText === null) continue;
    items.push({
      label: entry.key,
      kind: entry.container ? kinds.Folder : kinds.Field,
      insertText,
      filterText: `${typedHead}${entry.key}`,
      detail: `${entry.kind} · ${entry.summary}`,
      // A key of the document being evaluated beats any catalogue entry: it is
      // what the author came to the editor to write.
      sortText: `0${entry.key}`,
      range,
    });
  }
  return items;
}

function celCompletionItems(monaco: Monaco, spec: GomplateSpec) {
  const kinds = monaco.languages.CompletionItemKind;
  const items: Item[] = [];

  for (const fn of spec.cel.functions) {
    items.push(buildFunctionItem(monaco, fn, celInsertText(fn), kinds.Function));
  }
  for (const macro of spec.cel.macros) {
    items.push(buildMacroItem(monaco, macro));
  }
  for (const keyword of spec.cel.keywords) {
    items.push({
      label: keyword,
      kind: kinds.Keyword,
      insertText: keyword,
      detail: "CEL keyword",
    });
  }
  for (const type of spec.cel.types) {
    items.push({ label: type, kind: kinds.TypeParameter, insertText: type, detail: "CEL type" });
  }
  return items;
}

function goTemplateCompletionItems(monaco: Monaco, spec: GomplateSpec) {
  const kinds = monaco.languages.CompletionItemKind;
  const items: Item[] = [];

  for (const fn of spec.gotemplate.functions) {
    items.push(buildFunctionItem(monaco, fn, fn.name, kinds.Function));
  }
  for (const builtin of spec.gotemplate.builtins) {
    items.push({
      label: builtin,
      kind: kinds.Function,
      insertText: builtin,
      detail: "text/template builtin",
    });
  }
  for (const keyword of spec.gotemplate.keywords) {
    items.push({
      label: keyword,
      kind: kinds.Keyword,
      insertText: keyword,
      detail: "template keyword",
    });
  }
  return items;
}

function buildFunctionItem(
  monaco: Monaco,
  fn: SpecFunction,
  insertText: string,
  kind: monaco.languages.CompletionItemKind,
): Item {
  // A function with neither a Go signature nor an overload has no detail to
  // show; the key has to be absent rather than explicitly undefined.
  const detail = fn.signature ?? fn.overloads?.[0]?.result;
  return {
    label: fn.name,
    kind: fn.memberOnly ? monaco.languages.CompletionItemKind.Method : kind,
    insertText,
    ...(detail === undefined ? {} : { detail }),
    documentation: { value: functionDocumentation(fn) },
    filterText: fn.name,
    // Un-namespaced names first: they are the short, common ones, and a
    // namespace is easy to reach by typing its prefix.
    sortText: fn.namespace ? `2${fn.name}` : `1${fn.name}`,
  };
}

function buildMacroItem(monaco: Monaco, macro: SpecMacro): Item {
  return {
    label: macro.name,
    kind: monaco.languages.CompletionItemKind.Keyword,
    insertText: macro.name,
    detail: `macro (${macro.argCount === 0 ? "variadic" : `${macro.argCount} args`})`,
    documentation: { value: macroDocumentation(macro) },
    filterText: macro.name,
    sortText: `1${macro.name}`,
  };
}

/**
 * CEL functions take parentheses; a member-only function is offered without a
 * leading dot because the dot is already typed when completion fires.
 */
function celInsertText(fn: SpecFunction) {
  const arity = fn.overloads?.[0]?.args.length ?? 0;
  const receiver = fn.overloads?.[0]?.member ? 1 : 0;
  return arity - receiver === 0 ? `${fn.name}()` : `${fn.name}(`;
}

function wordRange(
  model: {
    getWordUntilPosition(p: { lineNumber: number; column: number }): {
      startColumn: number;
      endColumn: number;
    };
  },
  position: { lineNumber: number; column: number },
) {
  const word = model.getWordUntilPosition(position);
  return {
    startLineNumber: position.lineNumber,
    endLineNumber: position.lineNumber,
    startColumn: word.startColumn,
    endColumn: word.endColumn,
  };
}
