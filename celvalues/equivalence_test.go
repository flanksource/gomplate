package celvalues_test

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
	"github.com/flanksource/commons/decimal"
	"github.com/flanksource/gomplate/v4/celvalues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type comparisonResult struct {
	value, failure string
	cost           uint64
	nodeID         int64
}

func comparisonProgram(expression string, options []cel.EnvOption, optimize bool) cel.Program {
	GinkgoHelper()
	env, err := cel.NewEnv(append(options, cel.Variable("x", cel.DynType), cel.Variable("y", cel.DynType))...)
	Expect(err).NotTo(HaveOccurred())
	tree, issues := env.Compile(expression)
	Expect(issues.Err()).NotTo(HaveOccurred())
	programOptions := []cel.ProgramOption{cel.CostTracking(nil)}
	if optimize {
		programOptions = append(programOptions, cel.EvalOptions(cel.OptOptimize))
	}
	program, err := env.Program(tree, programOptions...)
	Expect(err).NotTo(HaveOccurred())
	return program
}

func comparisonOutcome(program cel.Program, bindings map[string]any) comparisonResult {
	GinkgoHelper()
	value, details, err := program.Eval(bindings)
	result := comparisonResult{value: fmt.Sprintf("%T:%v", value, value), cost: *details.ActualCost()}
	if err != nil {
		result.failure = err.Error()
	}
	if failure, ok := value.(*types.Err); ok {
		result.nodeID = failure.NodeID()
	}
	if unknown, ok := value.(*types.Unknown); ok {
		result.value = fmt.Sprintf("%T:%v", unknown, unknown.IDs())
	}
	return result
}

var _ = Describe("ordinary CEL comparison compatibility", func() {
	It("matches stock CEL results, errors, unknowns and costs with optimization enabled and disabled", func() {
		day := time.Date(2026, time.November, 10, 0, 0, 0, 0, time.UTC)
		pairs := []struct {
			name        string
			left, right any
		}{
			{"integers", int64(-1), int64(1)},
			{"signed unsigned", int64(math.MaxInt64), uint64(math.MaxInt64) + 1},
			{"unsigned double", uint64(math.MaxUint64), float64(math.MaxUint64)},
			{"integer double precision", int64(9007199254740993), float64(9007199254740992)},
			{"negative unsigned", int64(-1), uint64(0)},
			{"nan", math.NaN(), float64(1)},
			{"nan pair", math.NaN(), math.NaN()},
			{"infinities", math.Inf(-1), math.Inf(1)},
			{"null pair", nil, nil},
			{"null number", nil, int64(0)},
			{"strings", "a", "b"},
			{"booleans", false, true},
			{"timestamps", day, day.Add(time.Hour)},
			{"durations", time.Second, time.Minute},
			{"different types", "1", int64(1)},
			{"lists", []int{1}, []int{2}},
			{"maps", map[string]int{"a": 1}, map[string]int{"a": 2}},
			{"left unknown", types.NewUnknown(11, nil), int64(1)},
			{"both unknown", types.NewUnknown(11, nil), types.NewUnknown(22, nil)},
			{"left error", types.NewErr("left failure"), int64(1)},
			{"both error", types.NewErr("left failure"), types.NewErr("right failure")},
			{"unknown error", types.NewUnknown(11, nil), types.NewErr("right failure")},
			{"receiver", comparisonReceiver{}, int64(1)},
		}
		for _, optimize := range []bool{false, true} {
			for _, operator := range []string{"==", "!=", "<", "<=", ">", ">="} {
				expression := "x " + operator + " y"
				stock := comparisonProgram(expression, nil, optimize)
				library := comparisonProgram(expression, []cel.EnvOption{celvalues.Library()}, optimize)
				for _, pair := range pairs {
					for _, bindings := range []map[string]any{{"x": pair.left, "y": pair.right}, {"x": pair.right, "y": pair.left}} {
						Expect(comparisonOutcome(library, bindings)).To(Equal(comparisonOutcome(stock, bindings)), "%s %s optimize=%t", pair.name, expression, optimize)
					}
				}
			}
		}
	})

	It("evaluates ordinary and typed operands exactly once and preserves error short circuiting", func() {
		money, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "125", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		for _, first := range []ref.Val{types.Int(125), celvalues.NewAdapter(types.DefaultTypeAdapter).NativeToValue(money), types.NewErr("stop")} {
			for _, optimize := range []bool{false, true} {
				for _, operator := range []string{"==", "!=", "<", "<=", ">", ">="} {
					counts := map[string]int{}
					options := make([]cel.EnvOption, 0, 3)
					options = append(options, celvalues.Library())
					for name, value := range map[string]ref.Val{"left": first, "right": types.Int(125)} {
						options = append(options, cel.Function(name, cel.Overload(name+"_value", nil, cel.DynType,
							cel.FunctionBinding(func(...ref.Val) ref.Val { counts[name]++; return value }))))
					}
					program := comparisonProgram("left() "+operator+" right()", options, optimize)
					_, _, _ = program.Eval(map[string]any{})
					want := map[string]int{"left": 1, "right": 1}
					if types.IsError(first) {
						want = map[string]int{"left": 1}
					}
					Expect(counts).To(Equal(want), "%s optimize=%t first=%T", operator, optimize, first)
				}
			}
		}
	})
})

type comparisonReceiver struct{}

func (comparisonReceiver) Type() ref.Type {
	return types.NewObjectType("test.ComparisonReceiver", traits.ReceiverType)
}
func (comparisonReceiver) Value() any                                  { return "receiver" }
func (comparisonReceiver) Equal(ref.Val) ref.Val                       { return types.NewErr("receiver equality failure") }
func (r comparisonReceiver) ConvertToNative(reflect.Type) (any, error) { return r, nil }
func (comparisonReceiver) ConvertToType(ref.Type) ref.Val {
	return types.NewErr("unsupported conversion")
}
func (comparisonReceiver) Receive(string, string, []ref.Val) ref.Val { return types.True }
