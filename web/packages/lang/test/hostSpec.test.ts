import { describe, expect, it } from "vitest";
import * as monaco from "monaco-editor/esm/vs/editor/editor.api";
import { attributesFor, definitions, mergeSpec, registerGomplateLanguages, spec } from "../src";
import { celHoverProvider } from "../src/hover";
import type { GomplateSpec, LanguageId } from "../src";

registerGomplateLanguages(monaco, { completions: false, hovers: false });

/**
 * The catalogue a host serves from its own `GET /api/spec`.
 *
 * Shaped like duty's `catalog.query`: a namespaced global function with one
 * typed overload. A host's spec response contains gomplate's own functions too,
 * so the realistic input is gomplate's spec plus the host's.
 */
function hostSpec(): GomplateSpec {
  return {
    ...spec,
    cel: {
      ...spec.cel,
      namespaces: [...spec.cel.namespaces, "catalog"],
      functions: [
        ...spec.cel.functions,
        {
          name: "catalog.query",
          namespace: "catalog",
          doc: "Queries the config catalogue.",
          overloads: [{ id: "catalog.query_string", args: ["string"], result: "dyn" }],
        },
      ],
    },
  };
}

function tokensOf(text: string, languageId: string) {
  const lines = monaco.editor.tokenize(text, languageId);
  return (lines[0] ?? []).map((token) => token.type.replace(/\.[a-z-]+$/, ""));
}

describe("the generated attributes and the runtime derivation agree", () => {
  // The word lists are derived twice: in Go, when the bundle is generated, and
  // here, when a host's spec is merged in. This pins the two together — if
  // genmonarch's derivation changes, this fails rather than the highlighting
  // quietly going wrong for hosts only.
  it.each([
    ["cel", ["keywords", "constants", "typeKeywords", "macros", "namespaces", "globalFunctions", "memberFunctions"]],
    ["gomplate", ["keywords", "builtins", "namespaces", "functions"]],
    ["yaml-gomplate", ["keywords", "builtins", "namespaces", "functions"]],
  ] as const)("recomputes %s's word lists exactly", (languageId, names) => {
    const generated = definitions[languageId as LanguageId]!.monarch as unknown as Record<
      string,
      string[]
    >;
    const derived = attributesFor(languageId, spec)!;

    for (const name of names) {
      expect(derived[name], `${languageId}.${name}`).toEqual(generated[name]);
    }
  });

  it("leaves the grammar's own lists alone", () => {
    // `operators` comes from CEL.g4, not from the spec, so the derivation must
    // not claim to produce it — spreading a partial set over the definition is
    // what keeps it.
    expect(attributesFor("cel", spec)).not.toHaveProperty("operators");
    expect(definitions.cel.monarch).toHaveProperty("operators");
  });

  it("has nothing to derive for jsonpath", () => {
    expect(attributesFor("jsonpath", spec)).toBeNull();
  });
});

describe("merging a host's catalogue", () => {
  it("keeps gomplate's functions and adds the host's", () => {
    const merged = mergeSpec(spec, hostSpec());
    const names = merged.cel.functions.map((fn) => fn.name);
    expect(names).toContain("catalog.query");
    expect(names).toContain("k8s.cpuAsMillicores");
    expect(merged.cel.namespaces).toContain("catalog");
  });

  it("lets the host win a name it redefines", () => {
    // The host's binary is what evaluates, so its declaration is the true one.
    const overridden = mergeSpec(spec, {
      ...spec,
      cel: {
        ...spec.cel,
        functions: [{ name: "k8s.cpuAsMillicores", namespace: "k8s", doc: "host override" }],
      },
    });
    const fn = overridden.cel.functions.find((f) => f.name === "k8s.cpuAsMillicores");
    expect(fn?.doc).toBe("host override");
  });

  it("keeps macro overloads that share a name", () => {
    // `map` is registered at both 2 and 3 arguments; keying by name alone
    // silently drops one and the hover stops listing it.
    const merged = mergeSpec(spec, spec);
    const maps = merged.cel.macros.filter((macro) => macro.name === "map");
    expect(maps.length).toBe(spec.cel.macros.filter((m) => m.name === "map").length);
    expect(maps.length).toBeGreaterThan(1);
  });

  it("returns the base untouched when there is nothing to merge", () => {
    expect(mergeSpec(spec, undefined)).toBe(spec);
  });
});

describe("documenting a host's function", () => {
  const hover = (text: string, column: number) => {
    const model = monaco.editor.createModel(text, "cel");
    try {
      return celHoverProvider(mergeSpec(spec, hostSpec())).provideHover(model, {
        lineNumber: 1,
        column,
      });
    } finally {
      model.dispose();
    }
  };

  it("hovers a host's function with its signature and docs", () => {
    // Column 4 is inside `catalog`, so this also covers the dotted-word lookup
    // widening past the namespace separator.
    const contents = hover(`catalog.query("x")`, 4)?.contents;
    const markdown = contents?.map((c) => c.value).join("\n") ?? "";
    expect(markdown).toContain("catalog.query");
    expect(markdown).toContain("Queries the config catalogue.");
    expect(markdown).toContain("string");
  });

  it("still hovers gomplate's own functions", () => {
    const markdown = hover(`k8s.cpuAsMillicores("500m")`, 6)
      ?.contents.map((c) => c.value)
      .join("\n");
    expect(markdown).toContain("cpuAsMillicores");
  });

  it("says nothing about a name in neither catalogue", () => {
    expect(hover(`nonesuch(1)`, 4)).toBeNull();
  });
});

describe("setSpec", () => {
  it("tokenizes a host's namespaced function only after the spec arrives", () => {
    const languages = registerGomplateLanguages(monaco, { languages: ["cel"] });
    try {
      // Before: `catalog` is not a namespace this binary knows.
      expect(tokensOf(`catalog.query("x")`, "cel")).not.toContain("namespace");

      languages.setSpec(hostSpec());
      const after = tokensOf(`catalog.query("x")`, "cel");
      expect(after).toContain("namespace");
      expect(after).toContain("function");
    } finally {
      languages.dispose();
    }
  });

  it("still tokenizes gomplate's own functions afterwards", () => {
    const languages = registerGomplateLanguages(monaco, { languages: ["cel"] });
    try {
      languages.setSpec(hostSpec());
      expect(tokensOf(`k8s.cpuAsMillicores("500m")`, "cel")).toContain("namespace");
    } finally {
      languages.dispose();
    }
  });

  it("reverts to the baked catalogue when passed nothing", () => {
    const languages = registerGomplateLanguages(monaco, { languages: ["cel"] });
    try {
      languages.setSpec(hostSpec());
      languages.setSpec(undefined);
      expect(tokensOf(`catalog.query("x")`, "cel")).not.toContain("namespace");
    } finally {
      languages.dispose();
    }
  });

  it("does not throw when called before any model exists", () => {
    const languages = registerGomplateLanguages(monaco, { languages: ["cel"] });
    expect(() => languages.setSpec(hostSpec())).not.toThrow();
    languages.dispose();
  });
});
