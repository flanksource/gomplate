package celext_test

import (
	"fmt"
	"slices"
	"strings"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/cost"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/ext"
	"github.com/flanksource/gomplate/v4/celext"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type parityLibrary struct {
	name       string
	options    [3]cel.EnvOption
	expression string
	inputType  *cel.Type
}

func parityLibraries() []parityLibrary {
	return []parityLibrary{
		{
			name: "Strings",
			options: [3]cel.EnvOption{
				ext.Strings(), celext.Strings(), celext.Strings(celext.StringsWithoutCostModels()),
			},
			expression: "input.indexOf('x')", inputType: cel.StringType,
		},
		{
			name: "Encoders",
			options: [3]cel.EnvOption{
				ext.Encoders(), celext.Encoders(), celext.Encoders(celext.EncodersWithoutCostModels()),
			},
			expression: "base64.encode(input)", inputType: cel.BytesType,
		},
		{
			name: "Lists",
			options: [3]cel.EnvOption{
				ext.Lists(), celext.Lists(), celext.Lists(celext.ListsWithoutCostModels()),
			},
			expression: "input.reverse()", inputType: cel.ListType(cel.IntType),
		},
		{
			name: "Math",
			options: [3]cel.EnvOption{
				ext.Math(), celext.Math(), celext.Math(celext.MathWithoutCostModels()),
			},
			expression: "math.least(input)", inputType: cel.ListType(cel.IntType),
		},
		{
			name: "Sets",
			options: [3]cel.EnvOption{
				ext.Sets(), celext.Sets(), celext.Sets(celext.SetsWithoutCostModels()),
			},
			expression: "sets.contains(input, [1])", inputType: cel.ListType(cel.IntType),
		},
	}
}

func parityEnv(options ...cel.EnvOption) *cel.Env {
	GinkgoHelper()
	env, err := cel.NewEnv(append([]cel.EnvOption{cel.StdLib()}, options...)...)
	Expect(err).NotTo(HaveOccurred())
	return env
}

func parityDeclarations(env *cel.Env) string {
	var description []string
	for name, function := range env.Functions() {
		description = append(description, "function "+name)
		for _, overload := range function.OverloadDecls() {
			arguments := make([]string, len(overload.ArgTypes()))
			for i, argument := range overload.ArgTypes() {
				arguments[i] = argument.String()
			}
			description = append(description, fmt.Sprintf("  %s: %s(%s) -> %s member=%t",
				name, overload.ID(), strings.Join(arguments, ", "), overload.ResultType(), overload.IsMemberFunction()))
		}
	}
	for _, macro := range env.Macros() {
		description = append(description, "macro "+macro.MacroKey())
	}
	for _, library := range env.Libraries() {
		description = append(description, "library "+library)
	}
	slices.Sort(description)
	return strings.Join(description, "\n")
}

func parityCompile(env *cel.Env, expression string) *cel.Ast {
	GinkgoHelper()
	checked, issues := env.Compile(expression)
	Expect(issues.Err()).NotTo(HaveOccurred(), "compile %s", expression)
	return checked
}

func parityEvaluate(env *cel.Env, expression string) (ref.Val, bool) {
	GinkgoHelper()
	program, err := env.Program(parityCompile(env, expression))
	Expect(err).NotTo(HaveOccurred(), "program %s", expression)
	value, _, err := program.Eval(map[string]any{})
	return value, err != nil || types.IsError(value)
}

type parityEstimator struct{}

func (parityEstimator) EstimateSize(node cost.AstNode) *cost.SizeEstimate {
	if slices.Equal(node.Path(), []string{"input"}) {
		return &cost.SizeEstimate{Min: 64, Max: 64}
	}
	return nil
}

func (parityEstimator) EstimateCallCost(string, string, *cost.AstNode, []cost.AstNode) *cost.CallEstimate {
	return nil
}

var _ = Describe("Upstream extension parity", func() {
	for _, library := range parityLibraries() {
		It("matches upstream function names, overload signatures, macro keys and library names for "+library.name+" with and without cost models", func() {
			var descriptions [3]string
			for i, option := range library.options {
				descriptions[i] = parityDeclarations(parityEnv(option))
			}
			Expect(descriptions[1:]).To(Equal([]string{descriptions[0], descriptions[0]}))
		})

		It("matches upstream cost estimates by default and changes the estimate when "+library.name+" cost models are disabled", func() {
			var estimates [3]cost.CostEstimate
			for i, option := range library.options {
				env := parityEnv(option, cel.Variable("input", library.inputType))
				var err error
				estimates[i], err = env.EstimateCost(parityCompile(env, library.expression), parityEstimator{})
				Expect(err).NotTo(HaveOccurred(), "estimate %s", library.expression)
			}
			Expect(estimates[1]).To(Equal(estimates[0]))
			Expect(estimates[2]).NotTo(Equal(estimates[0]))
		})
	}

	Describe("evaluation with all five libraries at their default versions", func() {
		var environments [3]*cel.Env

		BeforeEach(func() {
			var options [3][]cel.EnvOption
			for _, library := range parityLibraries() {
				for i, option := range library.options {
					options[i] = append(options[i], option)
				}
			}
			for i := range environments {
				environments[i] = parityEnv(options[i]...)
			}
		})

		DescribeTable("matches upstream value types and values or error presence with and without cost models",
			func(expression string, wantError bool) {
				upstream, upstreamError := parityEvaluate(environments[0], expression)
				Expect(upstreamError).To(Equal(wantError), "upstream evaluation of %s", expression)
				for i, env := range environments[1:] {
					actual, actualError := parityEvaluate(env, expression)
					Expect(actualError).To(Equal(upstreamError), "environment %d: %s", i+1, expression)
					if !upstreamError {
						Expect(actual).NotTo(BeNil())
						Expect(actual.Type()).To(Equal(upstream.Type()), "environment %d: %s", i+1, expression)
						Expect(upstream.Equal(actual)).To(Equal(types.True), "environment %d: %s; upstream=%v actual=%v", i+1, expression, upstream, actual)
					}
				}
			},
			func(expression string, _ bool) string { return expression },
			Entry(nil, "'a,b'.split(',')", false),
			Entry(nil, "['a','b'].join('-')", false),
			Entry(nil, "'%s=%d'.format(['x', 1])", false),
			Entry(nil, "'hello'.charAt(1)", false),
			Entry(nil, "'hello'.indexOf('l')", false),
			Entry(nil, "'hello'.lastIndexOf('l')", false),
			Entry(nil, "'Hello'.lowerAscii()", false),
			Entry(nil, "'hello'.upperAscii()", false),
			Entry(nil, "'aaa'.replace('a','b',2)", false),
			Entry(nil, "'hello'.substring(1,3)", false),
			Entry(nil, "'  x '.trim()", false),
			Entry(nil, "'abc'.reverse()", false),
			Entry(nil, `strings.quote('a"b')`, false),
			Entry(nil, "base64.encode(b'hi')", false),
			Entry(nil, "base64.decode('aGk=')", false),
			Entry(nil, "base64.encodeUrl(b'??')", false),
			Entry(nil, "base64.decodeUrl('Pz8=')", false),
			Entry(nil, "json.encode({'a': 1})", false),
			Entry(nil, `json.parse('{"a":1}')`, false),
			Entry(nil, "yaml.encode({'a': 1})", false),
			Entry(nil, "yaml.parse('a: 1')", false),
			Entry(nil, "[1,[2,[3]]].flatten()", false),
			Entry(nil, "[3,1,2].sort()", false),
			Entry(nil, "[1,1,2].distinct()", false),
			Entry(nil, "[1,2,3].reverse()", false),
			Entry(nil, "[1,2,3,4].slice(1,3)", false),
			Entry(nil, "lists.range(3)", false),
			Entry(nil, "[1,2].hasAll([1])", false),
			Entry(nil, "[1,2].hasAny([3])", false),
			Entry(nil, "[1,2].hasOnly([1,2,3])", false),
			Entry(nil, "[1,2].hasExactly([2,1])", false),
			Entry(nil, "[{'n':2},{'n':1}].sortBy(e, e.n)", false),
			Entry(nil, "math.greatest(1, 5, 3)", false),
			Entry(nil, "math.least([4, 2])", false),
			Entry(nil, "math.abs(-2)", false),
			Entry(nil, "math.ceil(1.2)", false),
			Entry(nil, "math.floor(1.8)", false),
			Entry(nil, "math.round(1.5)", false),
			Entry(nil, "math.trunc(-1.5)", false),
			Entry(nil, "math.sign(-3)", false),
			Entry(nil, "math.sqrt(16)", false),
			Entry(nil, "math.isNaN(0.0/0.0)", false),
			Entry(nil, "math.bitAnd(6, 3)", false),
			Entry(nil, "math.bitShiftLeft(1, 3)", false),
			Entry(nil, "sets.contains([1,2], [1])", false),
			Entry(nil, "sets.equivalent([1,2], [2,1])", false),
			Entry(nil, "sets.intersects([1], [2])", false),
			Entry(nil, "'hello'.substring(-1)", true),
			Entry(nil, "base64.decode('!')", true),
			Entry(nil, "json.parse('{')", false),
			Entry(nil, "lists.range(-1)", true),
			Entry(nil, "math.abs(-9223372036854775808)", true),
		)
	})
})
