import { describe, expect, it } from "vitest";
import { createLazyJSONPathTree } from "@flanksource/clicky-ui/components";
import { toGraphNode } from "../src/panels/graphNodes";

const DOCUMENT = {
  pod: { metadata: { name: "web-7d4f" } },
  replicas: 3,
  ready: true,
  note: null,
  tags: ["a", "b"],
};

const tree = createLazyJSONPathTree(DOCUMENT, { keyPrefix: "test" });
const root = tree.roots[0]!;

async function childrenOf(node: { metadata?: { node: Parameters<typeof toGraphNode>[0] } }) {
  return (await tree.loadChildren(node.metadata!.node)).map(toGraphNode);
}

describe("mapping a document node onto a graph row", () => {
  it("labels the root with the path root rather than its internal key", () => {
    // `key` namespaces the node for caching (`test$`); it is not a display name.
    expect(toGraphNode(root).label).toBe("$");
    expect(toGraphNode(root).id).toBe("test$");
  });

  it("labels a child with the last segment of its path", async () => {
    const children = await childrenOf(toGraphNode(root));
    expect(children.map((child) => child.label)).toEqual([
      "pod",
      "replicas",
      "ready",
      "note",
      "tags",
    ]);
    expect(children.map((child) => child.path)).toEqual([
      "$.pod",
      "$.replicas",
      "$.ready",
      "$.note",
      "$.tags",
    ]);
  });

  it("labels a list element with its index", async () => {
    const children = await childrenOf(toGraphNode(root));
    const tags = children.find((child) => child.label === "tags")!;
    expect((await childrenOf(tags)).map((child) => child.path)).toEqual(["$.tags[0]", "$.tags[1]"]);
  });

  it("reports the scalar's own type, not the tree's structural kind", async () => {
    const byLabel = new Map((await childrenOf(toGraphNode(root))).map((c) => [c.label, c]));
    expect(byLabel.get("replicas")!.type).toBe("number");
    expect(byLabel.get("ready")!.type).toBe("boolean");
    expect(byLabel.get("note")!.type).toBe("null");
    expect(byLabel.get("pod")!.type).toBe("object");
    expect(byLabel.get("tags")!.type).toBe("array");
  });

  it("shows a scalar's value and a container's summary, never both", async () => {
    const byLabel = new Map((await childrenOf(toGraphNode(root))).map((c) => [c.label, c]));
    expect(byLabel.get("replicas")!.value).toBe(3);
    expect(byLabel.get("replicas")!.raw).toBeUndefined();
    expect(byLabel.get("pod")!.value).toBeUndefined();
    expect(byLabel.get("pod")!.raw).toBeTruthy();
  });

  it("marks only containers with children as expandable", async () => {
    const byLabel = new Map((await childrenOf(toGraphNode(root))).map((c) => [c.label, c]));
    expect(byLabel.get("pod")!.expandable).toBe(true);
    expect(byLabel.get("tags")!.expandable).toBe(true);
    expect(byLabel.get("replicas")!.expandable).toBe(false);
  });

  it("gives every row a distinct id, which is what selection is keyed by", async () => {
    const children = await childrenOf(toGraphNode(root));
    const ids = children.map((child) => child.id);
    expect(new Set(ids).size).toBe(ids.length);
  });
});
