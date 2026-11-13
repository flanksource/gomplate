# texttemplate

A fork of the Go standard library's `text/template` execution engine, taken from **go1.27.4** (`$GOROOT/src/text/template`). Parsing is unchanged and still uses `text/template/parse`; `internal/fmtsort` is vendored under `internal/`. Licensed under the Go BSD license in [LICENSE](LICENSE).

With `Execute`, or `ExecuteWith` and zero `ExecuteOptions`, it behaves exactly like `text/template`: the upstream `exec_test.go`, `multi_test.go` and examples run against it unmodified apart from the package name in expected error strings.

## Additions

```go
tpl := texttemplate.Must(texttemplate.New("t").Parse("{{ .item.metadata.name }}"))
err := tpl.ExecuteWith(w, env, texttemplate.ExecuteOptions{
	Values:  provider, // texttemplate.ValuesProvider; nil = text/template semantics
	NilSafe: true,
})
```

`ValuesProvider` has two methods:

- `Field(receiver, name)` resolves `.name` and string `index` keys. It is consulted before methods, struct fields and map keys; `found=false` falls back to them. `found=true` with an invalid value is a missing key and follows `missingkey`. With a provider set, a key nothing resolves on a struct is a missing key rather than `can't evaluate field`.
- `Value(v)` converts a value just before it is printed, ranged over, truth-tested by `if`/`with`, or passed to a function or method (including the pipeline's final value). Field-chain intermediates, `index` receivers and `$variable` assignments are not converted, so a provider can keep host values lazy until they are consumed.

`NilSafe` makes field access through a nil interface or nil pointer, and `index` of a nil value, a nil pointer, a missing key or an out-of-range integer position, yield the missing value instead of an error, and an action whose value is missing prints empty text instead of `<no value>`. Type errors stay errors and `missingkey=error` still wins. Options apply to nested `{{ template }}` and `{{ block }}` calls.

Options are per execution, so one parsed template can be shared by executions with different providers.

## Upstream sync

Every deviation from upstream is tagged `gomplate:`; `rg "gomplate:" texttemplate` lists them. The additions themselves live in `values.go`. Changes tagged `gomplate: lint` only satisfy this repository's golangci-lint rules (error naming, camelCase, `else if`, no `log.Fatal` after `defer`) and do not change behaviour. To sync, copy the upstream files over, re-apply the tagged hunks and run `go test ./texttemplate`.
