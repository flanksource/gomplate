import { describe, expect, it } from "vitest";
// The package root entry is the browser bundle; the editor API entry is the
// headless surface, which is all a model and a completion provider need.
import * as monaco from "monaco-editor/esm/vs/editor/editor.api";
import { pathExpression, registerGomplateLanguages, spec } from "../src";
import { completionProvider } from "../src/completion";

registerGomplateLanguages(monaco, { completions: false, hovers: false });

/** A payload with the shapes that make path rendering interesting. */
const DOCUMENT = {
  pod: {
    metadata: {
      name: "web-7d4f",
      labels: { app: "web", "app.kubernetes.io/name": "web" },
    },
    spec: { containers: [{ name: "app", image: "nginx:1.27" }] },
    status: { phase: "Running" },
  },
  count: 3,
};

const environment = () => DOCUMENT as unknown;

const FIELD_KINDS = [
  monaco.languages.CompletionItemKind.Field,
  monaco.languages.CompletionItemKind.Folder,
];

/**
 * Runs the real provider with the cursor at the end of `text`, and returns only
 * the items that came from the document — the catalogue items are asserted on
 * separately.
 */
function documentItems(languageId: string, text: string, source = environment) {
  return suggest(languageId, text, source).filter((item) => FIELD_KINDS.includes(item.kind));
}

function suggest(languageId: string, text: string, source?: () => unknown) {
  const model = monaco.editor.createModel(text, languageId);
  const lines = text.split("\n");
  const position = new monaco.Position(lines.length, lines[lines.length - 1]!.length + 1);
  try {
    return completionProvider(monaco, languageId, {
      spec,
      environment: source,
    }).provideCompletionItems(model, position).suggestions;
  } finally {
    model.dispose();
  }
}

function labels(languageId: string, text: string) {
  return documentItems(languageId, text).map((item) => String(item.label));
}

function itemFor(languageId: string, text: string, label: string) {
  return documentItems(languageId, text).find((candidate) => candidate.label === label)!;
}

function insertFor(languageId: string, text: string, label: string) {
  return itemFor(languageId, text, label)?.insertText;
}

describe("completing keys of the document", () => {
  it.each([
    ["cel", "pod."],
    ["gomplate", "{{ .pod."],
    ["jsonpath", "$.pod."],
  ])("offers the children of a map in %s", (languageId, text) => {
    expect(labels(languageId, text)).toEqual(["metadata", "spec", "status"]);
  });

  it.each([
    ["cel", "pod.metadata.", "name", "pod.metadata.name"],
    ["gomplate", "{{ .pod.metadata.", "name", ".pod.metadata.name"],
    ["jsonpath", "$.pod.metadata.", "name", "$.pod.metadata.name"],
  ])("inserts a whole %s path, not just the leaf", (languageId, text, label, expected) => {
    expect(insertFor(languageId, text, label)).toBe(expected);
  });

  it("completes a partially typed leaf", () => {
    expect(labels("cel", "pod.metadata.na")).toContain("name");
  });

  it("offers the top-level keys at the root of an expression", () => {
    expect(labels("cel", "")).toEqual(["pod", "count"]);
  });

  it("replaces the whole path typed so far, so nothing is glued onto it", () => {
    const item = documentItems("cel", "pod.metadata.").find((i) => i.label === "name")!;
    const range = item.range as monaco.IRange;
    expect(range.startColumn).toBe(1);
    expect(range.endColumn).toBe("pod.metadata.".length + 1);
  });

  it.each([
    ["cel", "pod.spec.containers.", "0", "pod.spec.containers.0"],
    ["cel", "pod.metadata.la", "labels", "pod.metadata.labels"],
    ["cel", "pod.metadata.labels.", "app.kubernetes.io/name", "pod.metadata.labels.app.kubernetes.io/name"],
    ["gomplate", "{{ .pod.", "metadata", ".pod.metadata"],
    ["jsonpath", "$.pod.", "metadata", "$.pod.metadata"],
  ])(
    "filters on the text typed, not the rendered path, in %s",
    (languageId, text, label, expected) => {
      // Monaco matches a candidate against the model text from the range start
      // to the cursor. A `filterText` of `pod.items[0]` never matches what the
      // author typed to get there — `pod.items.` — so the item disappears.
      const item = itemFor(languageId, text, label);
      expect(item.filterText).toBe(expected);
      // Whatever the author typed to reach this item still leads it.
      expect(text).toContain(String(item.filterText).slice(0, -label.length));
    },
  );

  it("carries the type and a sample so the shape is readable without running", () => {
    const items = documentItems("cel", "pod.metadata.");
    expect(items.find((i) => i.label === "name")!.detail).toBe('string · "web-7d4f"');
    expect(items.find((i) => i.label === "labels")!.detail).toBe("object · 2 keys");
    expect(documentItems("cel", "pod.spec.")[0]!.detail).toBe("array · 1 item");
  });

  it("sorts document keys ahead of the function catalogue", () => {
    const suggestions = suggest("cel", "pod.", environment);
    const key = suggestions.find((item) => item.label === "metadata")!;
    const catalogue = suggestions.filter((item) => !FIELD_KINDS.includes(item.kind));
    expect(catalogue.length).toBeGreaterThan(0);
    for (const item of catalogue) {
      expect(String(key.sortText) < String(item.sortText)).toBe(true);
    }
  });
});

describe("where a key path has no meaning", () => {
  it("offers nothing outside a template action", () => {
    // The same text inside `{{ }}` completes; as literal output it is prose.
    expect(labels("gomplate", "hello .pod.")).toEqual([]);
    expect(labels("gomplate", "{{ .pod.")).not.toEqual([]);
  });

  it("offers nothing inside a template comment", () => {
    expect(labels("gomplate", "{{/* .pod.")).toEqual([]);
  });

  it("requires the go-template leading dot", () => {
    expect(labels("gomplate", "{{ pod.")).toEqual([]);
  });

  it("requires a JSONPath root marker", () => {
    expect(labels("jsonpath", "pod.")).toEqual([]);
  });

  it("rejects a leading dot in CEL, which has no root object", () => {
    expect(labels("cel", ".pod.")).toEqual([]);
  });

  it("offers nothing for an unknown path", () => {
    expect(labels("cel", "nope.")).toEqual([]);
  });

  it("falls back to the catalogue when no document is supplied", () => {
    const suggestions = suggest("cel", "pod.", undefined);
    expect(suggestions.filter((item) => FIELD_KINDS.includes(item.kind))).toEqual([]);
    expect(suggestions.some((item) => item.label === "size")).toBe(true);
  });
});

describe("lists", () => {
  it("offers indices rather than the element's keys", () => {
    // `containers.name` parses in none of these languages. Offering the index
    // instead is what keeps the inserted expression evaluable: the item rewrites
    // the trailing dot into a subscript rather than appending to it.
    expect(labels("cel", "pod.spec.containers.")).toEqual(["0"]);
    expect(insertFor("cel", "pod.spec.containers.", "0")).toBe("pod.spec.containers[0]");
  });

  it("completes through an index", () => {
    expect(labels("cel", "pod.spec.containers[0].")).toEqual(["name", "image"]);
    expect(insertFor("cel", "pod.spec.containers[0].", "image")).toBe(
      "pod.spec.containers[0].image",
    );
  });

  it("has no index to offer in a go template, which reaches one through `index`", () => {
    expect(labels("gomplate", "{{ .pod.spec.containers.")).toEqual([]);
  });
});

describe("keys that are not identifiers", () => {
  it("subscripts them in CEL and JSONPath", () => {
    expect(insertFor("cel", "pod.metadata.labels.", "app.kubernetes.io/name")).toBe(
      'pod.metadata.labels["app.kubernetes.io/name"]',
    );
    expect(insertFor("jsonpath", "$.pod.metadata.labels.", "app.kubernetes.io/name")).toBe(
      '$.pod.metadata.labels["app.kubernetes.io/name"]',
    );
  });

  it("omits them for go templates, which need `index` rather than a path", () => {
    expect(labels("gomplate", "{{ .pod.metadata.labels.")).toEqual(["app"]);
  });
});

describe("pathExpression", () => {
  it.each([
    ["cel", ["pod", "metadata"], "pod.metadata"],
    ["cel", ["pod", 0], "pod[0]"],
    ["yaml-gomplate", ["pod", "metadata"], ".pod.metadata"],
    ["gomplate", [], "."],
    ["jsonpath", [], "$"],
    ["jsonpath", ["a b"], '$["a b"]'],
  ])("renders %s paths", (languageId, segments, expected) => {
    expect(pathExpression(languageId, segments as (string | number)[])).toBe(expected);
  });

  it("has no rendering for a go-template list element", () => {
    expect(pathExpression("gomplate", ["items", 0])).toBeNull();
  });

  it("has no rendering for an unknown language", () => {
    expect(pathExpression("klingon", ["a"])).toBeNull();
  });
});
