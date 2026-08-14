import { useMemo, useState } from "react";
import { ObjectGraph } from "@flanksource/clicky-ui/data";
import { createLazyJSONPathTree, literalSegments } from "@flanksource/clicky-ui/components";
import type { JSONPathNode, LazyJSONPathTree } from "@flanksource/clicky-ui/components";
import { pathExpression } from "@flanksource/gomplate-lang";
import { toGraphNode } from "./graphNodes";
import type { GraphNode } from "./graphNodes";

interface GraphPanelProps {
  /** The parsed input document, or undefined when there is nothing to show. */
  document: unknown;
  /** Monaco language id, which decides the syntax a click inserts. */
  languageId: string;
  /** Writes an expression at the source editor's cursor. */
  onInsert: (expression: string) => void;
}

/**
 * The shape of the document being evaluated against.
 *
 * The input pane shows the document as text; this shows it as paths. Clicking a
 * row writes that path into the expression, in the syntax of the language being
 * written, which is the step that otherwise means reading YAML and retyping it
 * by hand.
 */
export function GraphPanel({ document, languageId, onInsert }: GraphPanelProps) {
  const [selectedId, setSelectedId] = useState<string>();
  const [note, setNote] = useState<string>();

  const tree = useMemo<LazyJSONPathTree | null>(
    () =>
      document === undefined || document === null
        ? null
        : createLazyJSONPathTree(document, { keyPrefix: "input" }),
    [document],
  );

  const roots = useMemo(() => (tree ? tree.roots.map(toGraphNode) : []), [tree]);

  if (!tree) {
    return (
      <div className="flex h-full items-center justify-center p-6 text-center text-sm text-muted-foreground">
        Nothing to show yet — write a YAML or JSON document in the input pane and its shape
        appears here.
      </div>
    );
  }

  const select = (node: GraphNode) => {
    setSelectedId(node.id);
    const segments = node.path ? literalSegments(node.path) : undefined;
    if (!segments) {
      setNote("That row has no addressable path.");
      return;
    }
    const expression = pathExpression(languageId, segments);
    if (expression === null) {
      setNote(
        "A go template reaches a list element or a non-identifier key through `index`, " +
          "which is a call rather than a path — so there is nothing to insert.",
      );
      return;
    }
    if (expression === "") {
      setNote("The whole document has no name in this language — pick a key under it.");
      return;
    }
    setNote(undefined);
    onInsert(expression);
  };

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="min-h-0 flex-1 overflow-auto p-2">
        <ObjectGraph
          roots={roots}
          showControls
          defaultOpenDepth={2}
          selectedId={selectedId}
          onNodeSelect={select}
          loadChildren={async (node) => {
            const source = node.metadata?.node as JSONPathNode | undefined;
            if (!source) return [];
            return (await tree.loadChildren(source)).map(toGraphNode);
          }}
          empty="The document is empty."
        />
      </div>
      <footer className="shrink-0 border-t border-border px-4 py-2 text-xs text-muted-foreground">
        {note ?? "Click a key to insert its path at the cursor."}
      </footer>
    </div>
  );
}

