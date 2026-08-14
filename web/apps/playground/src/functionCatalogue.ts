import type { GomplateSpec, SpecFunction } from "@flanksource/gomplate-lang";

import type { EvalLanguage } from "./languages";

export type FunctionCatalogueFlavour = "cel" | "gotemplate";

export function functionCatalogueFlavour(
  language: EvalLanguage,
): FunctionCatalogueFlavour | null {
  if (language === "cel" || language === "gotemplate") return language;
  return null;
}

/**
 * The functions to browse for a language.
 *
 * Takes the catalogue rather than reading the baked one, so a host binary's own
 * functions show up in the browser as well as in completion — the tab count is
 * the quickest way to see whether the server's spec actually arrived.
 */
export function functionCatalogue(language: EvalLanguage, spec: GomplateSpec): SpecFunction[] {
  const flavour = functionCatalogueFlavour(language);
  if (!flavour) return [];
  return spec[flavour].functions;
}
