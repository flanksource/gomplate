import type { GomplateSpec, Monaco, SpecFunction, SpecMacro } from "./types";

/**
 * Indexes are built per registration, not once per module.
 *
 * A host's catalogue arrives from its `/api/spec` after the editor has already
 * mounted, so anything computed at module scope is frozen before the host can
 * speak. `setSpec` re-registers with a fresh index instead.
 */
export function registerCelHover(monaco: Monaco, languageId: string, spec: GomplateSpec) {
  return monaco.languages.registerHoverProvider(languageId, celHoverProvider(spec));
}

export function registerGoTemplateHover(monaco: Monaco, languageId: string, spec: GomplateSpec) {
  return monaco.languages.registerHoverProvider(languageId, goTemplateHoverProvider(spec));
}

/**
 * The providers the register functions install, exposed so a test can drive
 * them over a real model rather than through Monaco's registry, which offers no
 * way to enumerate what is registered.
 */
export function celHoverProvider(spec: GomplateSpec) {
  const functions = new Map<string, SpecFunction>(spec.cel.functions.map((f) => [f.name, f]));
  const macrosByName = new Map<string, SpecMacro[]>();
  for (const macro of spec.cel.macros) {
    macrosByName.set(macro.name, [...(macrosByName.get(macro.name) ?? []), macro]);
  }

  return {
    provideHover(model: Model, position: Position) {
      const word = dottedWordAt(model, position);
      if (!word) return null;

      const fn = functions.get(word.text) ?? functions.get(word.leaf);
      if (fn) return { range: word.range, contents: [{ value: functionDocumentation(fn) }] };

      const macros = macrosByName.get(word.leaf);
      if (macros) {
        return {
          range: word.range,
          contents: macros.map((macro) => ({ value: macroDocumentation(macro) })),
        };
      }
      return null;
    },
  };
}

export function goTemplateHoverProvider(spec: GomplateSpec) {
  const functions = new Map<string, SpecFunction>(
    spec.gotemplate.functions.map((f) => [f.name, f]),
  );

  return {
    provideHover(model: Model, position: Position) {
      const word = dottedWordAt(model, position);
      const fn = word && functions.get(word.text);
      if (!fn || !word) return null;
      return { range: word.range, contents: [{ value: functionDocumentation(fn) }] };
    },
  };
}

/** Renders a function as markdown: signature, overloads, docs, examples. */
export function functionDocumentation(fn: SpecFunction): string {
  const lines: string[] = [];
  lines.push("```", fn.signature ? `${fn.name}${fn.signature}` : fn.name, "```");

  if (fn.memberOnly) {
    lines.push("", "_Callable only in member position: `value." + leafOf(fn.name) + "(...)`._");
  }
  if (fn.doc) lines.push("", fn.doc);

  if (fn.overloads?.length) {
    lines.push("", "**Overloads**", "");
    for (const overload of fn.overloads) {
      const receiver = overload.member && overload.args.length > 0 ? `${overload.args[0]}.` : "";
      const args = overload.member ? overload.args.slice(1) : overload.args;
      lines.push(`- \`${receiver}${leafOf(fn.name)}(${args.join(", ")}) -> ${overload.result}\``);
    }
  }
  if (fn.examples?.length) {
    lines.push("", "**Examples**", "", "```cel", ...fn.examples, "```");
  }
  return lines.join("\n");
}

/** Renders a macro as markdown. */
export function macroDocumentation(macro: SpecMacro): string {
  const shape = macro.receiverStyle ? `value.${macro.name}(...)` : `${macro.name}(...)`;
  const arity = macro.argCount === 0 ? "variadic" : `${macro.argCount} arguments`;
  const lines = ["```", shape, "```", "", `_Macro, ${arity}. Expanded at parse time._`];
  if (macro.doc) lines.push("", macro.doc);
  if (macro.examples?.length) lines.push("", "```cel", ...macro.examples, "```");
  return lines.join("\n");
}

function leafOf(name: string) {
  const dot = name.lastIndexOf(".");
  return dot < 0 ? name : name.slice(dot + 1);
}

interface Model {
  getLineContent(line: number): string;
}

interface Position {
  lineNumber: number;
  column: number;
}

/**
 * Monaco's own word lookup stops at a dot, so `k8s.isHealthy` would be read as
 * `isHealthy`. The dotted name is what the spec is keyed by, so widen it here.
 */
function dottedWordAt(model: Model, position: Position) {
  const line = model.getLineContent(position.lineNumber);
  const isWord = (c: string) => /[A-Za-z0-9_.]/.test(c);

  let start = position.column - 1;
  while (start > 0 && isWord(line[start - 1]!)) start--;
  let end = position.column - 1;
  while (end < line.length && isWord(line[end]!)) end++;
  if (start === end) return null;

  const text = line.slice(start, end).replace(/^\.+|\.+$/g, "");
  if (!text) return null;

  return {
    text,
    leaf: leafOf(text),
    range: {
      startLineNumber: position.lineNumber,
      endLineNumber: position.lineNumber,
      startColumn: start + 1,
      endColumn: end + 1,
    },
  };
}
