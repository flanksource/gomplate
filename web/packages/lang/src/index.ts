import { LANGUAGE_IDS, definitions, spec } from "./generated";
import type { LanguageId } from "./generated";
import { registerCompletion } from "./completion";
import type { EnvironmentSource } from "./completion";
import { registerCelHover, registerGoTemplateHover } from "./hover";
import { pathFlavour } from "./environment";
import { attributesFor } from "./attributes";
import { mergeSpec } from "./merge";
import { defineThemes } from "./theme";
import type { GomplateSpec, LanguageDefinition, Monaco } from "./types";

export { spec, LANGUAGE_IDS, definitions };
export type { LanguageId };
export * from "./types";
export { GOMPLATE_DARK_THEME, GOMPLATE_LIGHT_THEME } from "./theme";
export {
  childEntries,
  isIdentifier,
  kindOf,
  pathExpression,
  pathFlavour,
  resolvePath,
  summarize,
} from "./environment";
export type { EnvironmentEntry, PathSegment, ValueKind } from "./environment";
export { environmentPrefixAt } from "./prefix";
export type { EnvironmentPrefix } from "./prefix";
export type { EnvironmentSource } from "./completion";
export { mergeSpec } from "./merge";
export { attributesFor, celAttributes, goTemplateAttributes } from "./attributes";
export type { Attributes } from "./attributes";

export interface RegisterOptions {
  /** Which languages to register. Defaults to all of them. */
  languages?: readonly LanguageId[];
  /** Register completion providers. Defaults to true. */
  completions?: boolean;
  /** Register hover providers. Defaults to true. */
  hovers?: boolean;
  /** Define the gomplate colour themes. Defaults to true. */
  themes?: boolean;
  /**
   * The document expressions are evaluated against, so completion can offer the
   * key paths it actually contains.
   *
   * A getter rather than a value: registration happens once, before the first
   * editor mounts, while the document keeps being edited afterwards. It is
   * called on every completion request.
   */
  environment?: EnvironmentSource;
  /**
   * A host's own catalogue, merged over gomplate's.
   *
   * Usually left unset here and supplied later through `setSpec`, because it
   * arrives from the host's `GET /api/spec` after the editor has mounted.
   */
  spec?: GomplateSpec;
}

/** What `registerGomplateLanguages` hands back. */
export interface RegisteredLanguages {
  /**
   * Replaces the catalogue and re-applies it.
   *
   * Registration has to happen in `beforeMount`, before the first model exists,
   * while a host's spec arrives over the network afterwards — so gating
   * registration on the fetch would stall the editor. Register with the baked
   * catalogue instead and call this when the response lands: the tokenizers are
   * re-applied with the merged word lists, and completion and hover are
   * re-registered against the merged functions.
   */
  setSpec(spec: GomplateSpec | undefined): void;
  dispose(): void;
}

/**
 * Registers gomplate's languages with a Monaco instance.
 *
 * Safe to call more than once: a language already registered is left alone, so
 * a component tree with several editors does not need to coordinate. The
 * returned handle removes only the providers this call added.
 */
export function registerGomplateLanguages(
  monaco: Monaco,
  options: RegisterOptions = {},
): RegisteredLanguages {
  const { languages = LANGUAGE_IDS, completions = true, hovers = true, themes = true } = options;

  if (themes) defineThemes(monaco);

  const known = new Set(monaco.languages.getLanguages().map((l) => l.id));
  const selected: LanguageId[] = [];

  for (const id of languages) {
    const definition: LanguageDefinition | undefined = definitions[id];
    if (!definition) {
      throw new Error(
        `unknown gomplate language "${id}"; expected one of ${LANGUAGE_IDS.join(", ")}`,
      );
    }
    selected.push(id);

    // Only the registration itself is once-only. The tokenizer and the
    // providers are re-applied below for every selected language, whether or
    // not this call is the one that introduced it -- otherwise a second editor
    // would get a handle whose setSpec silently does nothing.
    if (known.has(id)) continue;
    monaco.languages.register({ id });
    monaco.languages.setLanguageConfiguration(id, definition.configuration);
  }

  const applySpec = (merged: GomplateSpec) => {
    for (const id of selected) {
      const definition = definitions[id]!;
      const attributes = attributesFor(id, merged);
      // Spread over the generated definition rather than replacing it: the
      // tokenizer rules and the grammar-derived lists (CEL's `operators`) are
      // not the spec's to change.
      monaco.languages.setMonarchTokensProvider(
        id,
        attributes ? { ...definition.monarch, ...attributes } : definition.monarch,
      );

      const flavour = pathFlavour(id);
      const installed: { dispose(): void }[] = [];
      if (completions) {
        installed.push(
          registerCompletion(monaco, id, { spec: merged, environment: options.environment }),
        );
      }
      if (hovers && flavour === "cel") installed.push(registerCelHover(monaco, id, merged));
      if (hovers && flavour === "gotemplate") {
        installed.push(registerGoTemplateHover(monaco, id, merged));
      }
      replaceProviders(monaco, id, installed);
    }
  };

  applySpec(mergeSpec(spec, options.spec));

  return {
    setSpec(next) {
      applySpec(mergeSpec(spec, next));
    },
    dispose() {
      for (const id of selected) replaceProviders(monaco, id, []);
    },
  };
}

/**
 * One set of completion and hover providers per language, per Monaco.
 *
 * Monaco stacks providers rather than replacing them, so registering twice for
 * a language shows every suggestion twice. Tracking them here means the latest
 * registration wins instead, and a component tree with several editors does not
 * have to coordinate.
 */
const providersByLanguage = new WeakMap<Monaco, Map<string, { dispose(): void }[]>>();

function replaceProviders(monaco: Monaco, id: string, installed: { dispose(): void }[]) {
  let byLanguage = providersByLanguage.get(monaco);
  if (!byLanguage) {
    byLanguage = new Map();
    providersByLanguage.set(monaco, byLanguage);
  }
  for (const disposable of byLanguage.get(id) ?? []) disposable.dispose();
  if (installed.length === 0) byLanguage.delete(id);
  else byLanguage.set(id, installed);
}
