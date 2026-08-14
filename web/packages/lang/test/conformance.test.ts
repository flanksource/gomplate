import { describe, expect, it } from "vitest";
import * as monaco from "monaco-editor/esm/vs/editor/editor.api";
import { conformance } from "../src/generated";
import { registerGomplateLanguages } from "../src";

registerGomplateLanguages(monaco, { completions: false, hovers: false });

/**
 * Token start offsets Monarch produces for a single-line snippet, excluding
 * whitespace-only tokens so the comparison matches the lexer's own view.
 */
function monarchBoundaries(source: string, languageId: string): number[] {
  const [tokens = []] = monaco.editor.tokenize(source, languageId);
  const out: number[] = [];

  tokens.forEach((token, i) => {
    const end = i + 1 < tokens.length ? tokens[i + 1]!.offset : source.length;
    if (source.slice(token.offset, end).trim() === "") return;
    out.push(token.offset);
  });
  return out;
}

const byLanguage = new Map<string, typeof conformance>();
for (const testCase of conformance) {
  byLanguage.set(testCase.language, [...(byLanguage.get(testCase.language) ?? []), testCase]);
}

describe("conformance corpus", () => {
  it("covers every language and is not trivially small", () => {
    expect(conformance.length).toBeGreaterThan(50);
    expect([...byLanguage.keys()].sort()).toEqual(["cel", "gomplate", "jsonpath"]);
  });

  describe("cel token boundaries agree with cel-go's own lexer", () => {
    const cases = (byLanguage.get("cel") ?? []).filter((c) => c.boundaries?.length);

    it("has boundaries for every CEL case", () => {
      expect(cases.length).toBe(byLanguage.get("cel")?.length);
    });

    for (const testCase of cases) {
      it(`${testCase.source} (${testCase.origin})`, () => {
        // Every boundary Monarch produces must be a boundary the real lexer
        // also has -- that is, the tokenizer never ends a token part-way
        // through a real one. This is where the subtle bugs live: a
        // triple-quoted string cut short after two quotes, `0x1f` truncated to
        // `0`, `123u` split into a number and an identifier.
        //
        // The reverse does not hold, and should not be asserted: Monaco merges
        // adjacent tokens of the same type, so `()` collapses into one token
        // and legitimately loses the lexer's boundary between them.
        const expected = new Set(testCase.boundaries!);
        const spurious = monarchBoundaries(testCase.source, "cel").filter(
          (offset) => !expected.has(offset),
        );

        expect(
          spurious,
          `tokenizer split ${JSON.stringify(testCase.source)} at offsets the CEL lexer does not`,
        ).toEqual([]);
      });
    }
  });

  it("would actually catch a tokenizer that splits a literal", () => {
    // A gate that cannot fail is worse than no gate. Feed the comparison a
    // boundary set missing the ones inside a triple-quoted string -- the shape
    // the tokenizer produced before the grammar translation was fixed to order
    // alternatives longest-first -- and confirm it reports the difference.
    const source = '"""a"""';
    const brokenLexerView = new Set([0]);
    const asIfSplit = [0, 2, 3, 5];
    const spurious = asIfSplit.filter((offset) => !brokenLexerView.has(offset));
    expect(spurious).not.toEqual([]);

    // And the real tokenizer keeps it whole.
    expect(monarchBoundaries(source, "cel")).toEqual([0]);
  });

  describe("every snippet tokenizes without an error token", () => {
    for (const testCase of conformance) {
      it(`${testCase.language}: ${testCase.source}`, () => {
        const [tokens = []] = monaco.editor.tokenize(testCase.source, testCase.language);
        expect(tokens.length).toBeGreaterThan(0);
        // `invalid` is what a Monarch definition emits when nothing matched.
        expect(tokens.map((t) => t.type).filter((t) => t.startsWith("invalid"))).toEqual([]);
      });
    }
  });
});
