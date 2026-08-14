import type { CelSpec, GoTemplateSpec, GomplateSpec } from "./types";
import { pathFlavour } from "./environment";

/**
 * The Monarch word lists a tokenizer rule refers to as `@name`.
 *
 * The tokenizers themselves are fixed: their rules match any dotted call and
 * then dispatch on these lists (`$1@namespaces`, `$1@globalFunctions`, …). So a
 * host's own `catalog.query` needs no new grammar — only `catalog` in
 * `namespaces` and `query` in `globalFunctions`.
 *
 * This mirrors the derivation in `genmonarch/lang_cel.go` and
 * `lang_gotemplate.go`. `attributes.test.ts` recomputes the generated bundle
 * from the generated spec and asserts the two agree, so the mirror cannot drift
 * silently.
 */
export type Attributes = Record<string, string[]>;

/** CEL's word lists, minus `operators`, which comes from the grammar. */
export function celAttributes(spec: CelSpec): Attributes {
  const global: string[] = [];
  const member: string[] = [];
  for (const fn of spec.functions) {
    if (fn.memberOnly) member.push(fn.name);
    else global.push(leafOf(fn.name));
  }

  return {
    // `true`/`false`/`null` are constants, and colouring them as keywords too
    // would let whichever rule ran first decide.
    keywords: spec.keywords.filter((word) => !CONSTANTS.includes(word)),
    constants: [...CONSTANTS],
    typeKeywords: spec.types,
    macros: spec.macros.map((macro) => macro.name),
    namespaces: spec.namespaces,
    globalFunctions: global,
    memberFunctions: member,
  };
}

/** The go-template word lists, shared by the bare and embedded languages. */
export function goTemplateAttributes(spec: GoTemplateSpec): Attributes {
  return {
    keywords: spec.keywords,
    builtins: spec.builtins,
    namespaces: spec.namespaces,
    functions: spec.functions.filter((fn) => !fn.namespace).map((fn) => fn.name),
  };
}

/** The word lists for one language id, or null where a spec does not drive it. */
export function attributesFor(languageId: string, spec: GomplateSpec): Attributes | null {
  switch (pathFlavour(languageId)) {
    case "cel":
      return normalize(celAttributes(spec.cel));
    case "gotemplate":
      return normalize(goTemplateAttributes(spec.gotemplate));
    default:
      // JSONPath's vocabulary is the grammar's own, with nothing to extend.
      return null;
  }
}

const CONSTANTS = ["true", "false", "null"];

/**
 * Sorted and deduplicated, matching `Language.MarshalJSON` on the Go side, so a
 * recomputed list compares equal to a generated one.
 */
function normalize(attributes: Attributes): Attributes {
  const out: Attributes = {};
  for (const [name, words] of Object.entries(attributes)) {
    out[name] = [...new Set(words)].sort();
  }
  return out;
}

function leafOf(name: string) {
  const dot = name.lastIndexOf(".");
  return dot < 0 ? name : name.slice(dot + 1);
}
