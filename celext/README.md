# celext

`celext` is a generated copy of cel-go's Strings, Encoders, Lists, Math, and Sets extension libraries. cel-go compiles registered overload cost models during every environment configuration, including `Env.Extend`. Callers that never estimate cost can disable these models per library and avoid that compilation work.

Defaults match upstream, including cost model registration. The upstream function and option names, signatures, and `LibraryName()` strings are preserved. Each library adds one opt-out option: `StringsWithoutCostModels`, `EncodersWithoutCostModels`, `ListsWithoutCostModels`, `MathWithoutCostModels`, or `SetsWithoutCostModels`.

```go
env, err := cel.NewEnv(
    celext.Strings(celext.StringsWithoutCostModels()),
    celext.Encoders(celext.EncodersWithoutCostModels()),
    celext.Lists(celext.ListsWithoutCostModels()),
    celext.Math(celext.MathWithoutCostModels()),
    celext.Sets(celext.SetsWithoutCostModels()),
)
```

Import `cel.dev/cel-go/cel` and `github.com/flanksource/gomplate/v4/celext` for this example, and handle `err` before using `env`. Disable cost models only when their cost estimates are not needed. Library identities remain upstream identities, so choose upstream or celext options for a given library in an environment.

## Regeneration and drift

Install the [Flanksource gopatch fork](https://github.com/flanksource/gopatch), then run `GOWORK=off make celext`. Set `GOPATCH=/absolute/path/to/gopatch` to use a particular binary; otherwise the generator requires `gopatch` on PATH. The generator discovers the selected cel-go module through `go list -m`, copies the five library files and their three required helpers, retains Apache-2.0 headers, and uses gopatch for field additions, the sets receiver rename, and cost model rewrites. Generated files must not be edited by hand.

Run `GOWORK=off make celext-check` to regenerate and reject tracked differences under `celext`. The generator checks every generated file for remaining upstream cost registrations, checks replacement counts against the original source, and checks gofmt formatting. The Ginkgo SourceVersion spec compares the generated version with the cel-go dependency in the test binary and tells readers to run `make celext` on drift.

The generator rejects `init` functions and unreviewed package initializer calls so an upstream update cannot silently introduce global registration collisions with the original `ext` package. The copied state consists of local type values, an error value, and an encoder cache; no names are globally registered.

Remove this package and return callers to upstream `ext` once cel-go compiles cost models once per model set.
