# Embedding the expression playground

`playground` serves the API behind the language playground: evaluate an expression, fetch the function catalogue, fetch the sample documents. It runs everything through the same entry points a gomplate caller uses, so what an author sees in the editor is what production does.

A host embeds it to give its own authors a playground over its **own** language. Everything a host registers on top of gomplate — mission-control's `catalog.query`, `gitops.source` — flows through `Options`, so the catalogue, the highlighting and the evaluator all agree with what that binary can actually run.

## Mounting

```go
handler, err := playground.NewHandler(playground.Options{
    Timeout: 5 * time.Second,
})
if err != nil {
    return err
}
mux.Handle("/playground/", http.StripPrefix("/playground", handler.Mux()))
```

`Mux()` is an `http.Handler`, so an echo host wraps it:

```go
group.Any("/playground/*", echo.WrapHandler(
    http.StripPrefix("/playground", handler.Mux()),
))
```

Routes: `POST /api/eval`, `GET /api/spec`, `GET /api/examples`, `GET /api/health`.

## Standalone binary and browser evaluation

Build the complete binary with the Go version in `go.mod`, Node 24 and the pnpm version pinned in `web/package.json`:

```sh
make build
./bin/ceval serve
```

Open `http://127.0.0.1:8321`. The binary embeds the production `dist`, including Monaco, the generated language catalogue, the Go WASM runtime and the compressed evaluator. Serving a release requires no Node, pnpm, Go installation or source checkout.

`ceval serve -addr 127.0.0.1:9000 -timeout 5s` selects the listen address and the server evaluation timeout. `ceval version` prints the release version and commit. Existing `ceval -f env.yaml -e 'count * 2'` evaluation remains available.

The **Evaluate in** selector chooses **Server** or **Browser (WASM)**. The selection is preserved in the URL as `?eval=wasm`, so shared links retain it. WASM runs CEL, Go templates (including the embedded YAML/JSON/text modes), JSONPath and JavaScript through the same Go evaluator in a Web Worker. Source and input stay in the browser during WASM evaluation. WASM uses Gomplate's built-in functions; host extensions and host resources require Server mode. Browser network and filesystem restrictions still apply to functions that use those resources.

WASM loads lazily on the first browser evaluation. Its compressed module is decompressed using the browser's `DecompressionStream` API before streaming compilation. An evaluation, including initialization, has a ten-second ceiling; cancellation or timeout terminates the worker and the next run creates a new one. This releases runaway browser work as well as returning an error.

For development, `make playground` builds WASM and starts Vite with the Go API. `make wasm` refreshes the module and its matching `wasm_exec.js` after changing Go code. `make wasm-smoke` runs all four evaluators in the compiled WASM module under Node; `make web-build` runs that check before building the production UI. `make build-binary` compiles a binary from an already-built UI. Source checkouts must build the UI before compiling packages that embed it; `make build`, `make test` and `make lint` do that automatically.

The v4 release workflow follows Gavel and commons-db: it builds `dist` and adds it to a release-only commit before tagging, while main keeps generated bundles ignored. Release assets contain `ceval` archives for Linux, macOS and Windows on amd64 and arm64, plus SHA-256 checksums. Tagged source supports `go install github.com/flanksource/gomplate/v4/cmd/ceval@<tag>` without building the frontend locally. The `3.x` maintenance branch retains its existing semantic-release policy and downstream dependency updates.

Standalone servers pass the generated catalogue through `Options.Spec`. Hosts that register functions can omit it to extract a live catalogue, or supply a catalogue generated for their own registrations. A supplied spec must describe the functions granted by the other options.

## ⚠️ It carries no authorization

`/api/eval` runs arbitrary expressions with whatever `Options` grants them. In a host whose functions reach a database or a repository, **that is arbitrary execution against real data** — a `catalog.query` an author can write is a `catalog.query` anyone reaching the endpoint can write.

Mount it inside an already-authenticated route group, under the same authorization you would put any other query endpoint behind. The package deliberately does not offer a half-measure of its own.

## Supplying your own functions

Both fields are factories rather than plain slices, matching how hosts already register — duty keeps `map[string]func(Context) cel.EnvOption` because a function like `catalog.query` closes over the database handle it queries through.

```go
playground.Options{
    CelEnvs: func(ctx context.Context) []cel.EnvOption {
        opts := make([]cel.EnvOption, 0, len(duty.CelEnvFuncs))
        for _, f := range duty.CelEnvFuncs {
            opts = append(opts, f(dutyContext(ctx)))
        }
        return opts
    },
    Functions: func(ctx context.Context) map[string]any {
        out := map[string]any{}
        for name, f := range duty.TemplateFuncs {
            out[name] = f(dutyContext(ctx))
        }
        return out
    },
}
```

`CelEnvs` reaches three places at once, which is the point:

- the **evaluator**, through `gomplate.Template.CelEnvs`;
- the **compile check** that gives errors a source position, through `gomplate.CompileEnvOptions` — miss it there and a host's own function reports "undeclared reference" before the evaluator ever sees it;
- the **catalogue** at `GET /api/spec`, through `genmonarch.ExtractCEL`, which reads a live `cel.Env` rather than a maintained list. A `cel.Function("catalog.query", cel.Overload(...))` shows up there with its typed overloads, and the editor highlights and completes it without any change to the grammar.

`Functions` is exposed to both CEL and go templates, subject to gomplate's existing constraint: a CEL-visible entry must be a `func() any`. Anything with real arguments belongs in `CelEnvs`.

The spec is extracted once, at `NewHandler`, against `context.Background()`. The factories are per-request because a function's *binding* closes over a request; the *declarations* it registers — names, overloads, types — are the same every time.

## Sample data

```go
playground.Options{
    Examples: []playground.Example{{
        Name:     "Unhealthy config items",
        Language: playground.LanguageCEL,
        Source:   `catalog.query("health=unhealthy").size() > 0`,
        Input:    "…",
    }},
}
```

Served from `GET /api/examples`, always as an array.

## Bounding an evaluation

`Options.Timeout` bounds the **response**, not the work. gomplate honours no context deadline while evaluating — there is no `cel.ContextEval` and no deadline check in `RunTemplateContext` — so a runaway expression keeps its goroutine after the caller has been answered.

That is still the right trade for a shared endpoint, where a hung request is the worse failure. But it is not cancellation, and a host that expects to see CPU released on timeout will be disappointed.

## Known limits

- **Your functions will have thin documentation.** The catalogue reads `decl.Description()`; declare `cel.FunctionDocs` and overload examples to get prose and examples in hovers. Worth knowing: gomplate's own `gencel`-generated functions do not set them either, so this is a shared gap rather than a tax on hosts.
- **Go-template functions extract less well than CEL ones.** gomplate's own readable signatures come from parsing its source with `go/packages`; a host's `map[string]any` closure yields reflection types only, with no parameter names.
- **The conformance corpus does not cover host vocabulary.** It round-trips snippets from gomplate's docs through the real lexers. Hosts inherit the grammar guarantees, but need their own snippets for the same guard on their own functions.
