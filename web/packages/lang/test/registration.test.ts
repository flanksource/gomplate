import { describe, expect, it } from "vitest";
import { registerGomplateLanguages } from "../src";
import type { Monaco } from "../src";

/**
 * Counts what reaches Monaco.
 *
 * Real Monaco offers no way to enumerate the providers registered for a
 * language, and the invariant under test is the bookkeeping around it rather
 * than anything Monaco does — so this stands in for it. The tokenizer and
 * completion behaviour itself is covered against real Monaco elsewhere.
 */
function stubMonaco() {
  const state = {
    registered: [] as string[],
    tokenizers: [] as string[],
    providers: 0,
    disposed: 0,
  };

  const disposable = () => {
    state.providers += 1;
    return {
      dispose() {
        state.disposed += 1;
      },
    };
  };

  const monaco = {
    languages: {
      getLanguages: () => state.registered.map((id) => ({ id })),
      register: ({ id }: { id: string }) => state.registered.push(id),
      setLanguageConfiguration: () => ({ dispose() {} }),
      setMonarchTokensProvider: (id: string) => {
        state.tokenizers.push(id);
        return { dispose() {} };
      },
      registerCompletionItemProvider: disposable,
      registerHoverProvider: disposable,
      CompletionItemKind: new Proxy({}, { get: () => 0 }),
    },
    editor: { defineTheme: () => {} },
  } as unknown as Monaco;

  return { monaco, state };
}

describe("registering more than once", () => {
  it("registers the language itself only the first time", () => {
    const { monaco, state } = stubMonaco();
    registerGomplateLanguages(monaco, { languages: ["cel"] });
    registerGomplateLanguages(monaco, { languages: ["cel"] });
    expect(state.registered).toEqual(["cel"]);
  });

  it("replaces the previous providers rather than stacking them", () => {
    // The bug this guards: `beforeMount` fires once per editor, so a two-editor
    // page registered twice and every suggestion appeared twice.
    const { monaco, state } = stubMonaco();
    registerGomplateLanguages(monaco, { languages: ["cel"] });
    const installed = state.providers;

    registerGomplateLanguages(monaco, { languages: ["cel"] });
    expect(state.providers).toBe(installed * 2);
    expect(state.disposed).toBe(installed);
  });

  it("lets a later handle update the catalogue the earlier one registered", () => {
    // setSpec has to act on every language the call selected, not only those it
    // introduced, or the second editor's handle is inert.
    const { monaco, state } = stubMonaco();
    registerGomplateLanguages(monaco, { languages: ["cel"] });
    const second = registerGomplateLanguages(monaco, { languages: ["cel"] });

    const before = state.tokenizers.length;
    second.setSpec(undefined);
    expect(state.tokenizers.length).toBeGreaterThan(before);
  });

  it("disposes what it installed", () => {
    const { monaco, state } = stubMonaco();
    const languages = registerGomplateLanguages(monaco, { languages: ["cel"] });
    const installed = state.providers;
    languages.dispose();
    expect(state.disposed).toBe(installed);
  });

  it("rejects an unknown language instead of silently registering nothing", () => {
    const { monaco } = stubMonaco();
    expect(() =>
      // @ts-expect-error -- deliberately outside the union
      registerGomplateLanguages(monaco, { languages: ["klingon"] }),
    ).toThrow(/unknown gomplate language/);
  });
});
