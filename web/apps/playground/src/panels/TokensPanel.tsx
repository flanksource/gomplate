import { useMemo } from "react";
import { DataTable } from "@flanksource/clicky-ui";
import type { DataTableColumn } from "@flanksource/clicky-ui";
import * as monaco from "monaco-editor";

interface TokensPanelProps {
  source: string;
  languageId: string;
}

interface TokenRow extends Record<string, unknown> {
  line: number;
  text: string;
  token: string;
}

/**
 * Shows the token stream Monarch produces.
 *
 * This is the panel that makes a highlighting bug legible: colours tell you
 * something is off, the token stream tells you which rule matched. It is also
 * the fastest way to check a newly registered function is classified as
 * `function` rather than falling through to `identifier`.
 */
export function TokensPanel({ source, languageId }: TokensPanelProps) {
  const rows = useMemo(() => tokenize(source, languageId), [source, languageId]);

  const columns: DataTableColumn<TokenRow>[] = [
    { key: "line", label: "Line", align: "right", shrink: true, sortable: true },
    {
      key: "text",
      label: "Text",
      grow: true,
      cellClassName: "font-mono whitespace-pre",
    },
    {
      key: "token",
      label: "Token",
      sortable: true,
      filterable: true,
      cellClassName: "font-mono",
      // An `identifier` is the fallback every unmatched word lands on, so it is
      // the one class worth de-emphasising: what stands out is what matched.
      render: (value) => (
        <span className={String(value).startsWith("identifier") ? "text-muted-foreground" : ""}>
          {String(value)}
        </span>
      ),
    },
  ];

  return (
    <DataTable
      data={rows}
      columns={columns}
      autoFilter
      showGlobalFilter
      globalFilterPlaceholder="Filter tokens…"
      emptyMessage="Nothing to tokenize."
      className="h-full"
    />
  );
}

function tokenize(source: string, languageId: string): TokenRow[] {
  if (!source.trim()) return [];

  const lines = monaco.editor.tokenize(source, languageId);
  const rows: TokenRow[] = [];

  source.split(/\r\n|\r|\n/).forEach((line, index) => {
    const tokens = lines[index] ?? [];
    tokens.forEach((token, i) => {
      // Monaco reports a start offset only; each token runs to the next start.
      const end = i + 1 < tokens.length ? tokens[i + 1]!.offset : line.length;
      const text = line.slice(token.offset, end);
      if (text.trim() === "") return;
      rows.push({ line: index + 1, text, token: token.type });
    });
  });
  return rows;
}
