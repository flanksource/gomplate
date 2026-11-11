package gomplate

import (
	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func greetFunction(greeting string) CelFunction {
	return CelFunction{
		Name: "greet",
		Overloads: []CelOverload{{
			ID:     "greet_string",
			Args:   []*cel.Type{cel.StringType},
			Result: cel.StringType,
		}},
		Impl: func(_ string, args ...ref.Val) ref.Val {
			return types.String(greeting + " " + string(args[0].(types.String)))
		},
	}
}

var _ = Describe("CelFunctions", func() {
	environment := map[string]any{"name": "ada"}

	run := func(template Template) (any, error) {
		return RunExpression(environment, template)
	}

	It("evaluates each call with that call's implementation, even when the program is cached", func() {
		expression := "greet(name) + ' (late-bound)'"
		first, err := run(Template{Expression: expression, CelFunctions: []CelFunction{greetFunction("hello")}})
		Expect(err).NotTo(HaveOccurred())
		second, err := run(Template{Expression: expression, CelFunctions: []CelFunction{greetFunction("bye")}})
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{first, second}).To(Equal([]any{"hello ada (late-bound)", "bye ada (late-bound)"}))
	})

	It("supports receiver-style overloads", func() {
		shout := CelFunction{
			Name: "shout",
			Overloads: []CelOverload{{
				ID: "string_shout", Member: true, Args: []*cel.Type{cel.StringType}, Result: cel.StringType,
			}},
			Impl: func(_ string, args ...ref.Val) ref.Val { return types.String(string(args[0].(types.String)) + "!") },
		}
		Expect(run(Template{Expression: "name.shout()", CelFunctions: []CelFunction{shout}})).To(Equal("ada!"))
	})

	It("builds the extended environment once for templates with the same declarations", func() {
		before := extendedCelEnvBuilds.Load()
		for _, expression := range []string{"greet(name) + ' env-reuse-1'", "greet(name) + ' env-reuse-2'"} {
			_, err := run(Template{Expression: expression, CelFunctions: []CelFunction{greetFunction("hi")}})
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(extendedCelEnvBuilds.Load() - before).To(BeNumerically("<=", 1))
	})

	It("does not share a compiled program between templates whose overloads differ only in argument type", func() {
		expression := "greet('ada') + ' type-change'"
		_, err := run(Template{Expression: expression, CelFunctions: []CelFunction{greetFunction("hi")}})
		Expect(err).NotTo(HaveOccurred())

		intOnly := greetFunction("hi")
		intOnly.Overloads = []CelOverload{{ID: "greet_int", Args: []*cel.Type{cel.IntType}, Result: cel.StringType}}
		_, err = run(Template{Expression: expression, CelFunctions: []CelFunction{intOnly}})
		Expect(err).To(MatchError(ContainSubstring("found no matching overload for 'greet'")))
	})

	It("rejects a declared function without an implementation", func() {
		missing := greetFunction("hi")
		missing.Impl = nil
		_, err := run(Template{Expression: "greet(name)", CelFunctions: []CelFunction{missing}})
		Expect(err).To(MatchError(ContainSubstring("cel function greet has no implementation")))
	})

	It("keeps Functions calls bound to their own call", func() {
		expression := "fixedValue() + ' functions-cache'"
		results := make([]any, 0, 2)
		for _, value := range []string{"one", "two"} {
			out, err := run(Template{Expression: expression, Functions: map[string]any{"fixedValue": func() any { return value }}})
			Expect(err).NotTo(HaveOccurred())
			results = append(results, out)
		}
		Expect(results).To(Equal([]any{"one functions-cache", "two functions-cache"}))
	})

	It("treats templates with CelEnvs as uncacheable and those with only functions as cacheable", func() {
		Expect([]bool{
			Template{CelFunctions: []CelFunction{greetFunction("hi")}}.IsCacheable(),
			Template{Functions: map[string]any{"f": func() any { return 1 }}}.IsCacheable(),
			Template{CelEnvs: []cel.EnvOption{cel.Variable("x", cel.IntType)}}.IsCacheable(),
		}).To(Equal([]bool{true, true, false}))
	})
})
