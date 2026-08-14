import type { CelSpec, GoTemplateSpec, GomplateSpec, SpecFunction, SpecMacro } from "./types";

/**
 * Folds a host's catalogue into the one this package ships.
 *
 * A host binary registers functions on top of gomplate's — mission-control's
 * `catalog.query`, `gitops.source` — and serves the result from `/api/spec`.
 * That response already *contains* gomplate's own functions, so merging is
 * mostly a union; the interesting case is a host that overrides a name, where
 * the host wins because its binary is what will actually evaluate.
 */
export function mergeSpec(base: GomplateSpec, incoming: GomplateSpec | undefined): GomplateSpec {
  if (!incoming) return base;
  return {
    cel: mergeCel(base.cel, incoming.cel),
    gotemplate: mergeGoTemplate(base.gotemplate, incoming.gotemplate),
  };
}

function mergeCel(base: CelSpec, incoming: CelSpec): CelSpec {
  return {
    namespaces: union(base.namespaces, incoming.namespaces),
    keywords: union(base.keywords, incoming.keywords),
    types: union(base.types, incoming.types),
    variables: union(base.variables ?? [], incoming.variables ?? []),
    // Keyed by arity as well as name: `map` is registered twice, at 2 and 3
    // arguments, and folding those together would drop an overload the hover
    // list already renders separately.
    macros: keyed<SpecMacro>(
      base.macros,
      incoming.macros,
      (macro) => `${macro.name}/${macro.argCount}/${macro.receiverStyle}`,
    ),
    functions: byName<SpecFunction>(base.functions, incoming.functions),
  };
}

function mergeGoTemplate(base: GoTemplateSpec, incoming: GoTemplateSpec): GoTemplateSpec {
  return {
    namespaces: union(base.namespaces, incoming.namespaces),
    keywords: union(base.keywords, incoming.keywords),
    builtins: union(base.builtins, incoming.builtins),
    // Delimiters are a property of the binary's parser, not a list to merge:
    // a host that has changed them means it, and half of a pair would be worse
    // than either.
    delimiters: incoming.delimiters ?? base.delimiters,
    functions: byName<SpecFunction>(base.functions, incoming.functions),
  };
}

function union(base: readonly string[], incoming: readonly string[]): string[] {
  return [...new Set([...base, ...incoming])].sort();
}

/** Keyed by name, incoming wins, order stable and alphabetical. */
function byName<T extends { name: string }>(base: readonly T[], incoming: readonly T[]): T[] {
  return keyed(base, incoming, (item) => item.name);
}

function keyed<T extends { name: string }>(
  base: readonly T[],
  incoming: readonly T[],
  key: (item: T) => string,
): T[] {
  const merged = new Map<string, T>();
  for (const item of base) merged.set(key(item), item);
  for (const item of incoming) merged.set(key(item), item);
  return [...merged.values()].sort((a, b) => (a.name < b.name ? -1 : a.name > b.name ? 1 : 0));
}
