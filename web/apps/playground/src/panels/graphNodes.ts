import type { ObjectGraphNode } from "@flanksource/clicky-ui/data";
import { literalSegments } from "@flanksource/clicky-ui/components";
import type { JSONPathNode } from "@flanksource/clicky-ui/components";
import { kindOf } from "@flanksource/gomplate-lang";

/** An `ObjectGraphNode` that keeps the tree node it came from, for lazy loading. */
export interface GraphNode extends ObjectGraphNode {
  metadata?: { node: JSONPathNode };
}

/**
 * Maps one lazy-tree node to a graph row.
 *
 * Two of the tree's fields do not mean what their names suggest here. `key` is a
 * namespaced identity, not a display name, so the label comes off the end of the
 * path instead. And `kind` is structural — object, array, scalar — so a scalar's
 * own type is read from the value: `string` against `number` is the distinction
 * that decides whether an expression needs quotes.
 */
export function toGraphNode(node: JSONPathNode): GraphNode {
  const container = node.kind === "object" || node.kind === "array";
  const scalar = node.kind === "scalar";
  return {
    id: node.key,
    label: labelOf(node),
    path: node.path,
    kind: node.kind,
    type: scalar || container ? kindOf(node.value) : undefined,
    value: scalar ? scalarValue(node.value) : undefined,
    raw: scalar ? undefined : node.summary,
    expandable: container && node.childCount > 0,
    metadata: { node },
  };
}

/** The last segment of the node's path — the root has none, and is `$`. */
function labelOf(node: JSONPathNode): string {
  if (node.kind === "more") return "…";
  const segments = literalSegments(node.path);
  const last = segments?.[segments.length - 1];
  return last === undefined ? node.path : String(last);
}

function scalarValue(value: unknown): string | number | boolean | null {
  if (value === null || value === undefined) return null;
  const type = typeof value;
  if (type === "string" || type === "number" || type === "boolean") {
    return value as string | number | boolean;
  }
  return String(value);
}
