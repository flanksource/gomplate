import { describe, expect, it } from "vitest";
import { defaultExample, examplesFor } from "../src/examples";
import { LANGUAGES } from "../src/languages";

describe("examples", () => {
  // They are keyed by the playground language id, which is not always the
  // evaluator name — the go template language is `gomplate` here and
  // `gotemplate` on the wire. Getting that wrong silently opens the editor
  // empty, because the lookup falls through to a blank fallback.
  it.each(LANGUAGES.map((language) => [language.id]))("has an example for %s", (id) => {
    expect(examplesFor(id).length).toBeGreaterThan(0);
    expect(defaultExample(id).source).not.toBe("");
  });

  it("gives every example an input to evaluate against", () => {
    for (const language of LANGUAGES) {
      for (const example of examplesFor(language.id)) {
        expect(example.name, `${language.id}: ${example.name}`).not.toBe("");
      }
    }
  });
});
