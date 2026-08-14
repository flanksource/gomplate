import { describe, expect, it } from "vitest";
// The package root entry is the browser bundle, which needs a real DOM to even
// load. The editor API entry is the headless surface, and it is all the
// tokenizer needs.
import * as monaco from "monaco-editor/esm/vs/editor/editor.api";
import { LANGUAGE_IDS, registerGomplateLanguages } from "../src";

registerGomplateLanguages(monaco, { completions: false, hovers: false });

/**
 * Tokenizes with real Monaco and returns `text=token` pairs, dropping
 * whitespace so the assertions stay about the interesting tokens.
 *
 * Monaco reports a token's start offset only, so each token's text runs to the
 * start of the next one.
 */
function tokenize(text: string, languageId: string): string[] {
  const lines = monaco.editor.tokenize(text, languageId);
  const out: string[] = [];

  text.split(/\r\n|\r|\n/).forEach((line, index) => {
    const tokens = lines[index] ?? [];
    tokens.forEach((token, i) => {
      const end = i + 1 < tokens.length ? tokens[i + 1]!.offset : line.length;
      const value = line.slice(token.offset, end);
      if (value.trim() === "") return;
      // Monaco appends the language's tokenPostfix to every token type.
      out.push(`${value}=${token.type.replace(/\.[a-z-]+$/, "")}`);
    });
  });
  return out;
}

describe("registration", () => {
  it("registers every generated language", () => {
    const registered = new Set(monaco.languages.getLanguages().map((l) => l.id));
    for (const id of LANGUAGE_IDS) expect(registered).toContain(id);
  });

  it("is idempotent, so several editors can register independently", () => {
    expect(() => registerGomplateLanguages(monaco, { completions: false, hovers: false })).not.toThrow();
  });

  it("rejects an unknown language instead of silently registering nothing", () => {
    expect(() =>
      // @ts-expect-error -- deliberately outside the union
      registerGomplateLanguages(monaco, { languages: ["klingon"] }),
    ).toThrow(/unknown gomplate language/);
  });
});

describe("cel", () => {
  it("colours a namespaced call, distinguishing namespace from function", () => {
    expect(tokenize("k8s.isHealthy(pod)", "cel")).toEqual([
      "k8s=namespace",
      ".=delimiter",
      "isHealthy=function",
      "(=delimiter.parenthesis",
      "pod=identifier",
      ")=delimiter.parenthesis",
    ]);
  });

  it("does not treat an ordinary field access as a namespace", () => {
    expect(tokenize("pod.metadata", "cel")).toEqual([
      "pod=identifier",
      ".=delimiter",
      "metadata=variable.field",
    ]);
  });

  it("colours a member-only function but not the same word used as a variable", () => {
    expect(tokenize("[1,2].sum()", "cel")).toContain("sum=function.member");
    expect(tokenize("sum + 1", "cel")).toEqual([
      "sum=identifier",
      "+=operator",
      "1=number",
    ]);
  });

  it("highlights macros only in call position", () => {
    expect(tokenize("has(a.b)", "cel")).toContain("has=keyword.macro");
    // Receiver-style macros are macros, not member functions.
    expect(tokenize("[1].fold(e, acc, acc + e)", "cel")).toContain("fold=keyword.macro");
    expect(tokenize("items.fold(e, acc, acc + e)", "cel")).toContain("fold=keyword.macro");
    // The same word outside call position is an ordinary field.
    expect(tokenize("a.fold", "cel")).toContain("fold=variable.field");
  });

  it("keeps a triple-quoted string whole", () => {
    expect(tokenize('"""a "b" c"""', "cel")).toEqual(['"""a "b" c"""=string']);
  });

  it("recognises raw and bytes string prefixes", () => {
    expect(tokenize('r"a\\db"', "cel")).toEqual(['r"a\\db"=string']);
    expect(tokenize('b"abc"', "cel")).toEqual(['b"abc"=string.bytes']);
  });

  it("recognises back-tick escaped identifiers", () => {
    expect(tokenize("`a.b-c`", "cel")).toEqual(["`a.b-c`=identifier.escaped"]);
  });

  it("separates uint and float literals from plain ints", () => {
    expect(tokenize("123u", "cel")).toEqual(["123u=number.uint"]);
    expect(tokenize("1.5e-3", "cel")).toEqual(["1.5e-3=number.float"]);
    expect(tokenize("0x1f", "cel")).toEqual(["0x1f=number"]);
  });

  it("distinguishes optional access from the ternary operator", () => {
    // The name after `.?` is a field, exactly as it is after a plain `.`.
    expect(tokenize("a.?b", "cel")).toEqual([
      "a=identifier",
      ".?=operator.optional",
      "b=variable.field",
    ]);
    expect(tokenize('a.?b.orValue("x")', "cel")).toEqual([
      "a=identifier",
      ".?=operator.optional",
      "b=variable.field",
      ".=delimiter",
      "orValue=function.member",
      "(=delimiter.parenthesis",
      '"x"=string',
      ")=delimiter.parenthesis",
    ]);
    expect(tokenize("a ? b : c", "cel")).toEqual([
      "a=identifier",
      "?=operator",
      "b=identifier",
      ":=operator",
      "c=identifier",
    ]);
  });

  it("colours comments and constants", () => {
    expect(tokenize("// note\ntrue", "cel")).toEqual([
      "// note=comment",
      "true=keyword.constant",
    ]);
  });
});

describe("gomplate", () => {
  it("separates literal text from an action", () => {
    // Monaco merges adjacent tokens of the same type, so the literal run keeps
    // its trailing space.
    expect(tokenize("Hello {{ .name }}!", "gomplate")).toEqual([
      "Hello =source",
      "{{=delimiter.template",
      ".=delimiter",
      "name=variable.field",
      "}}=delimiter.template",
      "!=source",
    ]);
  });

  it("colours a pipeline into a namespaced function", () => {
    expect(tokenize("{{ .name | strings.ToUpper }}", "gomplate")).toEqual([
      "{{=delimiter.template",
      ".=delimiter",
      "name=variable.field",
      "|=operator.pipe",
      "strings=namespace",
      ".=delimiter",
      "ToUpper=function",
      "}}=delimiter.template",
    ]);
  });

  it("handles trim markers", () => {
    expect(tokenize("a{{- if .x -}}b{{- end -}}c", "gomplate")).toContain("{{-=delimiter.template");
    expect(tokenize("a{{- if .x -}}b", "gomplate")).toContain("if=keyword");
  });

  it("treats a template comment as a comment, not an action", () => {
    // The whole comment is one merged token; what matters is that `hidden` is
    // not tokenized as a keyword or a function.
    expect(tokenize("{{/* hidden */}}", "gomplate")).toEqual(["{{/* hidden */}}=comment"]);
    expect(tokenize("{{/* if range */}}", "gomplate")).toEqual(["{{/* if range */}}=comment"]);
  });

  it("colours variables and assignment separately from fields", () => {
    expect(tokenize("{{ $x := .y }}", "gomplate")).toEqual([
      "{{=delimiter.template",
      "$x=variable",
      ":==operator",
      ".=delimiter",
      "y=variable.field",
      "}}=delimiter.template",
    ]);
  });

  it("marks the delimiter directive header", () => {
    expect(tokenize("# gotemplate: left-delim=$[[ right-delim=]]", "gomplate")).toEqual([
      "# gotemplate: left-delim=$[[ right-delim=]]=comment.directive",
    ]);
  });

  it("colours builtins distinctly from gomplate functions", () => {
    const tokens = tokenize("{{ printf \"%s\" (toJSON .x) }}", "gomplate");
    expect(tokens).toContain("printf=function.builtin");
    expect(tokens).toContain("toJSON=function");
  });
});

describe("yaml-gomplate", () => {
  it("colours YAML structure and the template inside a value", () => {
    expect(tokenize("name: {{ .app }}", "yaml-gomplate")).toEqual([
      "name=type.yaml",
      ":=delimiter",
      "{{=delimiter.template",
      ".=delimiter",
      "app=variable.field",
      "}}=delimiter.template",
    ]);
  });

  it("keeps YAML comments and constants", () => {
    expect(tokenize("# note\nenabled: true", "yaml-gomplate")).toEqual([
      "# note=comment",
      "enabled=type.yaml",
      ":=delimiter",
      "true=keyword.constant",
    ]);
  });

  it("colours a template inside a list item", () => {
    const tokens = tokenize("items:\n  - {{ .a }}", "yaml-gomplate");
    expect(tokens).toContain("{{=delimiter.template");
    expect(tokens).toContain("a=variable.field");
  });
});

describe("json-gomplate", () => {
  it("colours object keys and an embedded template", () => {
    expect(tokenize('{"k": "{{ .v }}"}', "json-gomplate")).toEqual([
      "{=delimiter.curly",
      '"k"=type.json',
      ":=delimiter",
      '"=string',
      "{{=delimiter.template",
      ".=delimiter",
      "v=variable.field",
      "}}=delimiter.template",
      '"=string',
      "}=delimiter.curly",
    ]);
  });
});

describe("jsonpath", () => {
  it("colours root, descent and filters", () => {
    expect(tokenize("$..book[?(@.price < 10)]", "jsonpath")).toEqual([
      "$=variable.root",
      "..=operator.descendant",
      "book=variable.field",
      "[=delimiter.square",
      "?(=keyword.filter",
      "@=variable.current",
      ".=delimiter",
      "price=variable.field",
      "<=operator",
      "10=number",
      ")=delimiter.parenthesis",
      "]=delimiter.square",
    ]);
  });

  it("colours a wildcard and a slice", () => {
    const tokens = tokenize("$.items[*][0:2]", "jsonpath");
    expect(tokens).toContain("*=operator.wildcard");
    expect(tokens).toContain(":=operator.slice");
  });
});
