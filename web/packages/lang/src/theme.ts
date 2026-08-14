import type { Monaco } from "./types";

export const GOMPLATE_LIGHT_THEME = "gomplate-light";
export const GOMPLATE_DARK_THEME = "gomplate-dark";

/**
 * Token rules shared by both themes. Only the colours differ, so the token
 * vocabulary stays in one place.
 *
 * The token names are the ones the generated tokenizers emit; anything not
 * listed falls back to the base theme, which is why the themes inherit from
 * `vs` and `vs-dark` rather than starting from nothing.
 */
const TOKENS = [
  "namespace",
  "function",
  "function.member",
  "function.builtin",
  "keyword.macro",
  "keyword.constant",
  "keyword.directive",
  "operator.optional",
  "operator.pipe",
  "delimiter.template",
  "variable",
  "variable.field",
  "variable.anchor",
  "variable.root",
  "variable.current",
  "identifier.escaped",
  "string.bytes",
  "number.uint",
  "number.float",
  "type.yaml",
  "type.json",
  "comment.directive",
] as const;

type Palette = Record<(typeof TOKENS)[number], string>;

const LIGHT: Palette = {
  namespace: "267F99",
  function: "795E26",
  "function.member": "795E26",
  "function.builtin": "0000FF",
  "keyword.macro": "AF00DB",
  "keyword.constant": "0000FF",
  "keyword.directive": "AF00DB",
  "operator.optional": "AF00DB",
  "operator.pipe": "AF00DB",
  "delimiter.template": "AF00DB",
  variable: "001080",
  "variable.field": "001080",
  "variable.anchor": "267F99",
  "variable.root": "AF00DB",
  "variable.current": "AF00DB",
  "identifier.escaped": "001080",
  "string.bytes": "A31515",
  "number.uint": "098658",
  "number.float": "098658",
  "type.yaml": "0451A5",
  "type.json": "0451A5",
  "comment.directive": "008000",
};

const DARK: Palette = {
  namespace: "4EC9B0",
  function: "DCDCAA",
  "function.member": "DCDCAA",
  "function.builtin": "569CD6",
  "keyword.macro": "C586C0",
  "keyword.constant": "569CD6",
  "keyword.directive": "C586C0",
  "operator.optional": "C586C0",
  "operator.pipe": "C586C0",
  "delimiter.template": "C586C0",
  variable: "9CDCFE",
  "variable.field": "9CDCFE",
  "variable.anchor": "4EC9B0",
  "variable.root": "C586C0",
  "variable.current": "C586C0",
  "identifier.escaped": "9CDCFE",
  "string.bytes": "CE9178",
  "number.uint": "B5CEA8",
  "number.float": "B5CEA8",
  "type.yaml": "9CDCFE",
  "type.json": "9CDCFE",
  "comment.directive": "6A9955",
};

/** Defines the gomplate themes. Idempotent -- Monaco overwrites by name. */
export function defineThemes(monaco: Monaco) {
  monaco.editor.defineTheme(GOMPLATE_LIGHT_THEME, {
    base: "vs",
    inherit: true,
    rules: rulesFor(LIGHT),
    colors: {},
  });
  monaco.editor.defineTheme(GOMPLATE_DARK_THEME, {
    base: "vs-dark",
    inherit: true,
    rules: rulesFor(DARK),
    colors: {},
  });
}

function rulesFor(palette: Palette) {
  return TOKENS.map((token) => ({ token, foreground: palette[token] }));
}
