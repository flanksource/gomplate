import { useMemo } from "react";
import { Badge, DataTable } from "@flanksource/clicky-ui";
import type { DataTableColumn } from "@flanksource/clicky-ui";
import type { GomplateSpec, SpecFunction } from "@flanksource/gomplate-lang";

interface SpecPanelProps {
  /** Which catalogue to browse. */
  flavour: "cel" | "gotemplate";
  /** The catalogue itself — the server's when it has one, gomplate's otherwise. */
  spec: GomplateSpec;
}

interface FunctionRow extends Record<string, unknown> {
  name: string;
  namespace: string;
  signature: string;
  kind: string;
  doc: string;
}

/**
 * Browses the generated function catalogue.
 *
 * This is the first accurate reference that exists for this fork: CEL.md
 * documents functions that live in `duty` and omits whole namespaces, while
 * docs-src carries the upstream list. What is shown here is read out of the
 * live registries, so it is what the evaluator will actually accept.
 */
export function SpecPanel({ flavour, spec }: SpecPanelProps) {
  const rows = useMemo(() => {
    const functions = flavour === "cel" ? spec.cel.functions : spec.gotemplate.functions;
    return functions.map(toRow);
  }, [flavour, spec]);

  const columns: DataTableColumn<FunctionRow>[] = [
    {
      key: "name",
      label: "Name",
      sortable: true,
      grow: true,
      cellClassName: "font-mono",
    },
    {
      key: "namespace",
      label: "Namespace",
      sortable: true,
      filterable: true,
      shrink: true,
      render: (value) =>
        value ? String(value) : <span className="text-muted-foreground">—</span>,
    },
    {
      key: "kind",
      label: "Call",
      filterable: true,
      shrink: true,
      // `x.sum()` is legal where a bare `sum(x)` is not, and nothing else in
      // the catalogue records that -- so it is worth a column of its own.
      render: (value) =>
        value === "member" ? (
          <Badge variant="outline">member only</Badge>
        ) : (
          <span className="text-muted-foreground">global</span>
        ),
    },
    {
      key: "signature",
      label: "Signature",
      grow: true,
      cellClassName: "font-mono text-muted-foreground",
    },
    { key: "doc", label: "Description", grow: true },
  ];

  return (
    <DataTable
      data={rows}
      columns={columns}
      autoFilter
      showGlobalFilter
      globalFilterPlaceholder="Filter functions…"
      defaultSort={{ key: "name", dir: "asc" }}
      emptyMessage="No functions in this catalogue."
      className="h-full"
    />
  );
}

function toRow(fn: SpecFunction): FunctionRow {
  return {
    name: fn.name,
    namespace: fn.namespace ?? "",
    signature: signatureOf(fn),
    kind: fn.memberOnly ? "member" : "global",
    doc: fn.doc ?? "",
  };
}

function signatureOf(fn: SpecFunction): string {
  if (fn.signature) return fn.signature;

  const [overload] = fn.overloads ?? [];
  if (!overload) return "";

  // A member overload carries its receiver as the first argument; an author
  // writes it before the dot, not inside the parentheses.
  const args = overload.member ? overload.args.slice(1) : overload.args;
  return `(${args.join(", ")}) -> ${overload.result}`;
}
