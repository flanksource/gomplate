import { describe, expect, it } from "vitest";
import { mergeSpec, spec } from "@flanksource/gomplate-lang";

import { functionCatalogue, functionCatalogueFlavour } from "./functionCatalogue";

describe("playground function catalogues", () => {
  it.each(["jsonpath", "javascript"] as const)(
    "does not claim Go-template functions are available in %s",
    (language) => {
      expect(functionCatalogueFlavour(language)).toBeNull();
      expect(functionCatalogue(language, spec)).toEqual([]);
    },
  );

  it.each(["cel", "gotemplate"] as const)("uses the generated %s catalogue", (language) => {
    expect(functionCatalogueFlavour(language)).toBe(language);
    expect(functionCatalogue(language, spec).length).toBeGreaterThan(100);
  });

  it("browses a host's own functions, not only the ones baked in", () => {
    // The tab count is the quickest signal that the server's spec arrived, so
    // it has to read the merged catalogue rather than the packaged one.
    const served = mergeSpec(spec, {
      ...spec,
      cel: {
        ...spec.cel,
        functions: [...spec.cel.functions, { name: "catalog.query", namespace: "catalog" }],
      },
    });
    const names = functionCatalogue("cel", served).map((fn) => fn.name);
    expect(names).toContain("catalog.query");
    expect(names.length).toBe(functionCatalogue("cel", spec).length + 1);
  });
});
