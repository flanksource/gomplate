# @flanksource/gomplate-lang

Monaco language support for the expression languages [gomplate](https://github.com/flanksource/gomplate) evaluates: CEL, Go templates, templated YAML/JSON, and JSONPath.

Nothing in this package is written by hand. The tokenizers are generated from the grammars gomplate's own parsers use, and the function catalogue is read out of a live `cel.Env` and the `text/template` FuncMap gomplate installs — so the editor cannot disagree with the evaluator.

| Layer | Source |
|---|---|
| CEL lexical rules | `CEL.g4`, cel-go's ANTLR grammar |
| CEL functions, macros, types | `Env.Functions()`, `Env.Macros()` on a live environment |
| Template keywords, builtins, delimiters | the `text/template` lexer, read from its AST |
| gomplate functions | reflection over `gomplate.CreateFuncs` |

## Install

```bash
pnpm add @flanksource/gomplate-lang
```

`monaco-editor` is a peer dependency (`>=0.48 <1`). This package never imports it: you pass your Monaco instance in, so the host app keeps ownership of its version and bundling.

## Usage

```ts
import * as monaco from "monaco-editor";
import { registerGomplateLanguages } from "@flanksource/gomplate-lang";

registerGomplateLanguages(monaco);
```

With `@monaco-editor/react`, register in `beforeMount` — a model created before registration resolves to plaintext and is never revisited:

```tsx
<Editor language="cel" beforeMount={(monaco) => registerGomplateLanguages(monaco)} />
```

The call is idempotent, so several editors can register independently without coordinating. It returns a disposable that removes the completion and hover providers it added.

### Options

```ts
registerGomplateLanguages(monaco, {
  languages: ["cel", "yaml-gomplate"], // default: all
  completions: true,
  hovers: true,
  themes: true,
  environment: () => currentDocument, // default: none
});
```

### Completing the document

`environment` is what turns the function catalogue into an editor that knows the payload being evaluated against: typing `pod.` then offers the keys that document actually has, ahead of the 267 CEL functions.

It is a **getter**, not a value. Registration happens once, before the first editor mounts, while the document keeps being edited afterwards — a value would freeze at first mount. It is called on every completion request, so keep it cheap (read a ref; do not re-parse).

Each suggestion replaces the whole path typed so far rather than appending to it, so what lands in the editor is valid in that language: `pod["app.kubernetes.io/name"]` for an awkward key, `pod.items[0]` where a list needs an index, and nothing at all where a go template would need `index` instead of a path.

Paths render per language — `pod.metadata.name` in CEL, `.pod.metadata.name` inside a `{{ }}` action, `$.pod.metadata.name` in JSONPath. `pathExpression(languageId, segments)` is exported so a host UI can insert the same syntax the editor completes.

### Completing a host's own functions

The catalogue baked into this package is gomplate's. A host binary — mission-control, commons-db — registers more on top and serves the result from `GET /api/spec`. Fold it in with `setSpec`:

```ts
const languages = registerGomplateLanguages(monaco, { environment });

fetch("/api/spec")
  .then((r) => r.json())
  .then((served) => languages.setSpec(served));
```

Registration has to happen in `beforeMount`, before the first model exists, while the catalogue arrives over the network afterwards — so gating registration on the fetch would stall the editor. Register with the baked catalogue and update it when the response lands. (Which of the two lands first is a race, so a host that already has the spec can pass it as `spec` at registration; both paths are safe.)

There is no new grammar involved. The generated tokenizers match any dotted call and dispatch on word lists — `namespaces`, `globalFunctions`, `memberFunctions`, `macros` — so `catalog.query(…)` highlights the moment `catalog` joins `namespaces`. `setSpec` re-applies those lists and re-registers completion and hover against the merged functions.

`mergeSpec(base, incoming)` is exported for anyone who needs the merged catalogue themselves — a function browser, say. Incoming wins on a name collision, since its binary is what evaluates.

### Languages

| Id | For |
|---|---|
| `cel` | CEL expressions (`Template.Expression`) |
| `gomplate` | a bare Go template (`Template.Template`) |
| `yaml-gomplate` | YAML with `{{ }}` in it — how most configuration is written |
| `json-gomplate` | JSON with `{{ }}` in it |
| `text-gomplate` | plain text with `{{ }}` in it |
| `jsonpath` | JSONPath, the dialect `ojg` evaluates |

JavaScript (`Template.Javascript`) uses Monaco's own `javascript` language and is not generated here.

### Themes

`gomplate-light` and `gomplate-dark` colour the tokens the generated tokenizers emit — namespaces, member functions, macros, template delimiters, optional access. They inherit from `vs` and `vs-dark`, so anything they do not name keeps its usual colour.

```ts
import { GOMPLATE_DARK_THEME, GOMPLATE_LIGHT_THEME } from "@flanksource/gomplate-lang";

monaco.editor.setTheme(isDark ? GOMPLATE_DARK_THEME : GOMPLATE_LIGHT_THEME);
```

`setTheme` is global to a Monaco instance, and `@monaco-editor/react` re-applies its `theme` prop whenever an editor mounts. If a wrapper hardcodes that prop, re-assert the theme from `onMount`.

### The catalogue

The generated spec is the accurate function reference for this fork — more so than the checked-in Markdown, which documents functions that live in other repositories and omits whole namespaces.

```ts
import { spec, celFunction, celNamespace } from "@flanksource/gomplate-lang/spec";

spec.cel.namespaces;              // k8s, aws, math, time, filepath, ...
celFunction("k8s.isHealthy");     // overloads, argument and result types
celNamespace("k8s");              // everything under k8s.*
```

Note that the CEL and go-template catalogues genuinely differ: `conv.*` and `path.*` are registered for templates but not for CEL, and the same helper is `k8s.IsHealthy` in a template and `k8s.isHealthy` in CEL.

## Regenerating

From the gomplate repository root, after changing any registered function:

```bash
make monarch        # regenerate
make monarch-check  # fail if the checked-in files are stale (CI runs this)
```

## Testing

```bash
pnpm test
```

Two suites run against real Monaco. `tokenize.test.ts` asserts exact token streams for each language. `conformance.test.ts` replays a corpus generated by running cel-go's own ANTLR lexer over snippets from the reference docs, and asserts the tokenizer never ends a token part-way through a real one — the failure mode behind a triple-quoted string cut short, `0x1f` truncated to `0`, or `123u` split in two.
