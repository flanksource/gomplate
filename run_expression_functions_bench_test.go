package gomplate

import (
	"testing"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types/ref"
)

// Head-only: CelFunction does not exist on the benchmark base, and the benchmark
// workflow copies only run_expression_bench_test.go into the base worktree.
// BenchmarkRunExpressionContextFunctions measures templates that carry functions: they are
// cached by their declarations, with the implementation bound on each evaluation.
func BenchmarkRunExpressionContextFunctions(b *testing.B) {
	lateBound := CelFunction{
		Name:      "bench_noop",
		Overloads: []CelOverload{{ID: "bench_noop_string", Args: []*cel.Type{cel.StringType}, Result: cel.StringType}},
		Impl:      func(_ string, args ...ref.Val) ref.Val { return args[0] },
	}
	cases := map[string]Template{
		"celFunctions": {Expression: `bench_noop(config_type) == "Kubernetes::Pod"`, CelFunctions: []CelFunction{lateBound}},
		"functions":    {Expression: `bench_fixed() == config_type`, Functions: map[string]any{"bench_fixed": func() any { return "Kubernetes::Pod" }}},
	}
	for name, template := range cases {
		b.Run(name+"/cache=hit/environment=large", func(b *testing.B) {
			env := benchExprEnv(true)
			assertBenchmarkExpression(b, env, template)
			b.ReportAllocs()
			for b.Loop() {
				if _, err := RunExpression(env, template); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
