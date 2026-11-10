package celvalues

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/functions"
	"cel.dev/cel-go/common/operators"
	"cel.dev/cel-go/common/stdlib"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
	"cel.dev/cel-go/interpreter"
)

// Library enables Date and Money comparisons and string conversions.
func Library() cel.EnvOption { return cel.Lib(library{}) }

type library struct{}

func (library) LibraryName() string { return "gomplate.celvalues" }

func (library) CompileOptions() []cel.EnvOption {
	options := []cel.EnvOption{cel.Types(dateType), moneyTypes()}
	for _, scalar := range []*cel.Type{dateType, moneyType, bigMoneyType} {
		options = append(options, cel.Function("string", cel.Overload("string_"+scalar.TypeName(), []*cel.Type{scalar}, cel.StringType,
			cel.UnaryBinding(func(value ref.Val) ref.Val { return value.ConvertToType(types.StringType) }))))
		peers := []*cel.Type{scalar}
		if scalar == dateType {
			peers = append(peers, cel.StringType, cel.TimestampType)
		} else {
			peers = append(peers, cel.IntType, cel.UintType, cel.DoubleType)
		}
		for _, peer := range peers {
			options = append(options, comparisonDeclarations(scalar, peer)...)
		}
	}
	return append(options, arithmeticDeclarations()...)
}

func comparisonDeclarations(scalar, peer *cel.Type) []cel.EnvOption {
	options := make([]cel.EnvOption, 0, 8)
	for _, operator := range []string{operators.Less, operators.LessEquals, operators.Greater, operators.GreaterEquals} {
		pairs := [][]*cel.Type{{scalar, peer}}
		if scalar != peer {
			pairs = append(pairs, []*cel.Type{peer, scalar})
		}
		for _, pair := range pairs {
			id := strings.Join([]string{"gomplate", operator, pair[0].TypeName(), pair[1].TypeName()}, "_")
			options = append(options, cel.Function(operator, cel.Overload(id, pair, cel.BoolType)))
		}
	}
	return options
}

func (library) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{cel.CustomDecoratorV2(func(value interpreter.InterpretableV2) (interpreter.InterpretableV2, error) {
		call, ok := value.(interpreter.InterpretableCall)
		if !ok {
			return value, nil
		}
		arithmetic := isArithmetic(call.Function(), len(call.Args()))
		if !arithmetic && (len(call.Args()) != 2 || !isComparison(call.Function())) {
			return value, nil
		}
		bindings, err := nativeOperators()
		if err != nil {
			return nil, err
		}
		if arithmetic {
			return &arithmeticCall{InterpretableCall: call, native: bindings[call.Function()]}, nil
		}
		return &comparisonCall{InterpretableCall: call, native: bindings[call.Function()]}, nil
	})}
}

// nativeOperators holds CEL's own ordering and arithmetic bindings, used for operands that are not Date or Money.
var nativeOperators = sync.OnceValues(func() (map[string]*functions.Overload, error) {
	bindings := make(map[string]*functions.Overload, 9)
	for _, function := range stdlib.Functions() {
		name := function.Name()
		unary := isArithmetic(name, 1)
		if !unary && !isArithmetic(name, 2) && (!isComparison(name) || isEquality(name)) {
			continue
		}
		overloads, err := function.Bindings()
		if err != nil {
			return nil, err
		}
		for _, overload := range overloads {
			if overload.Operator == name && ((unary && overload.Unary != nil) || (!unary && overload.Binary != nil)) {
				bindings[name] = overload
			}
		}
		if bindings[name] == nil {
			return nil, fmt.Errorf("CEL operator %s has no singleton binding", name)
		}
	}
	return bindings, nil
})

type comparisonCall struct {
	interpreter.InterpretableCall
	native *functions.Overload
}

func (c *comparisonCall) Eval(activation interpreter.Activation) ref.Val {
	return interpreter.EvalActivation(activation, c.Exec)
}

func (c *comparisonCall) Exec(frame *interpreter.ExecutionFrame) ref.Val {
	left := c.Args()[0].Exec(frame)
	if types.IsError(left) {
		if isEquality(c.Function()) {
			c.track(frame, left, types.UnknownType, left)
		}
		return left
	}
	right := c.Args()[1].Exec(frame)
	result := c.compare(left, right)
	if !types.IsError(right) || (types.IsUnknown(left) && !isEquality(c.Function())) || isEquality(c.Function()) {
		c.track(frame, left, right, result)
	}
	return result
}

func (c *comparisonCall) track(frame *interpreter.ExecutionFrame, left, right, result ref.Val) {
	if tracker := frame.CostTracker(); tracker != nil {
		tracker.EvalBinary(frame, c.ID(), c, left, right, result)
	}
}

func (c *comparisonCall) compare(left, right ref.Val) ref.Val {
	operator := c.Function()
	if types.IsUnknown(left) && !isEquality(operator) {
		unknown, _ := types.MaybeMergeUnknowns(right, left.(*types.Unknown))
		return unknown
	}
	if types.IsError(right) {
		return right
	}
	unknown, _ := types.MaybeMergeUnknowns(left, nil)
	if unknown, _ = types.MaybeMergeUnknowns(right, unknown); unknown != nil {
		return unknown
	}
	if left == types.NullValue || right == types.NullValue {
		if isEquality(operator) {
			return equalityVerdict(operator, types.Equal(left, right))
		}
		return c.nativeComparison(left, right)
	}
	if _, ok := left.(interface{ scalarComparison() }); ok {
		return comparisonVerdict(operator, left.(traits.Comparer).Compare(right))
	}
	if _, ok := right.(interface{ scalarComparison() }); ok {
		comparison := right.(traits.Comparer).Compare(left)
		if value, ok := comparison.(types.Int); ok {
			comparison = -value
		}
		return comparisonVerdict(operator, comparison)
	}
	if isEquality(operator) {
		return equalityVerdict(operator, types.Equal(left, right))
	}
	return c.nativeComparison(left, right)
}

func (c *comparisonCall) nativeComparison(left, right ref.Val) ref.Val {
	var result ref.Val
	switch {
	case c.native != nil && (c.native.OperandTrait == 0 || left.Type().HasTrait(c.native.OperandTrait)):
		result = c.native.Binary(left, right)
	case left.Type().HasTrait(traits.ReceiverType):
		result = left.(traits.Receiver).Receive(c.Function(), c.OverloadID(), []ref.Val{right})
	default:
		return types.NewErrWithNodeID(c.ID(), "no such overload: %s", c.Function())
	}
	if failure, ok := result.(*types.Err); ok && failure.NodeID() == 0 {
		return types.LabelErrNode(c.ID(), failure)
	}
	return result
}

func (Date) scalarComparison()       {}
func (moneyValue) scalarComparison() {}

func comparisonVerdict(operator string, comparison ref.Val) ref.Val {
	value, ok := comparison.(types.Int)
	if !ok {
		return comparison
	}
	switch operator {
	case operators.Equals:
		return types.Bool(value == 0)
	case operators.NotEquals:
		return types.Bool(value != 0)
	case operators.Less:
		return types.Bool(value < 0)
	case operators.LessEquals:
		return types.Bool(value <= 0)
	case operators.Greater:
		return types.Bool(value > 0)
	case operators.GreaterEquals:
		return types.Bool(value >= 0)
	default:
		return types.NewErr("unsupported comparison %s", operator)
	}
}

func equalityVerdict(operator string, value ref.Val) ref.Val {
	if operator == operators.NotEquals {
		return types.Bool(value != types.True)
	}
	return value
}

func isEquality(operator string) bool {
	return operator == operators.Equals || operator == operators.NotEquals
}

func isComparison(operator string) bool {
	switch operator {
	case operators.Equals, operators.NotEquals, operators.Less, operators.LessEquals, operators.Greater, operators.GreaterEquals:
		return true
	default:
		return false
	}
}

func scalarNative(value ref.Val, target reflect.Type) (any, error) {
	if reflect.TypeOf(value.Value()).AssignableTo(target) {
		return value.Value(), nil
	}
	if target.Kind() == reflect.String {
		return reflect.ValueOf(value.(fmt.Stringer).String()).Convert(target).Interface(), nil
	}
	return nil, fmt.Errorf("cannot convert %s to %s", value.Type().TypeName(), target)
}
