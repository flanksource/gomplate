package celvalues

import (
	"fmt"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/functions"
	"cel.dev/cel-go/common/operators"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/interpreter"
	"github.com/flanksource/commons/bigdecimal"
)

func arithmeticDeclarations() []cel.EnvOption {
	numbers := []*cel.Type{cel.IntType, cel.UintType, cel.DoubleType}
	options := []cel.EnvOption{
		cel.Function(operators.Negate, cel.Overload("gomplate_negate_bigmoney", []*cel.Type{bigMoneyType}, bigMoneyType)),
		cel.Function(operators.Add, cel.Overload("gomplate_add_bigmoney_bigmoney", []*cel.Type{bigMoneyType, bigMoneyType}, bigMoneyType)),
		cel.Function(operators.Subtract, cel.Overload("gomplate_subtract_bigmoney_bigmoney", []*cel.Type{bigMoneyType, bigMoneyType}, bigMoneyType)),
	}
	for _, number := range numbers {
		for _, operator := range []string{operators.Add, operators.Subtract, operators.Multiply, operators.Divide} {
			pairs := [][]*cel.Type{{bigMoneyType, number}}
			if operator != operators.Divide {
				pairs = append(pairs, []*cel.Type{number, bigMoneyType})
			}
			for _, pair := range pairs {
				id := fmt.Sprintf("gomplate_%s_%s_%s", operator, pair[0].TypeName(), pair[1].TypeName())
				options = append(options, cel.Function(operator, cel.Overload(id, pair, bigMoneyType)))
			}
		}
	}
	return options
}

func isArithmetic(operator string, arity int) bool {
	switch operator {
	case operators.Add, operators.Subtract, operators.Multiply, operators.Divide:
		return arity == 2
	case operators.Negate:
		return arity == 1
	default:
		return false
	}
}

// arithmeticCall evaluates Money arithmetic exactly and leaves every other operand to CEL's own binding.
type arithmeticCall struct {
	interpreter.InterpretableCall
	native *functions.Overload
}

func (c *arithmeticCall) Eval(activation interpreter.Activation) ref.Val {
	return interpreter.EvalActivation(activation, c.Exec)
}

func (c *arithmeticCall) Exec(frame *interpreter.ExecutionFrame) ref.Val {
	operands := make([]ref.Val, len(c.Args()))
	var unknown *types.Unknown
	for i, arg := range c.Args() {
		operands[i] = arg.Exec(frame)
		if types.IsError(operands[i]) {
			return operands[i]
		}
		unknown, _ = types.MaybeMergeUnknowns(operands[i], unknown)
	}
	if unknown != nil {
		return unknown
	}
	result := c.apply(operands)
	if tracker := frame.CostTracker(); tracker != nil {
		if len(operands) == 1 {
			tracker.EvalUnary(frame, c.ID(), c, operands[0], result)
		} else {
			tracker.EvalBinary(frame, c.ID(), c, operands[0], operands[1], result)
		}
	}
	return result
}

func (c *arithmeticCall) apply(operands []ref.Val) ref.Val {
	if !hasMoney(operands) {
		return c.nativeArithmetic(operands)
	}
	result, err := moneyArithmetic(c.Function(), operands)
	if err != nil {
		return types.LabelErrNode(c.ID(), types.WrapErr(err))
	}
	return moneyValue{native: result}
}

func (c *arithmeticCall) nativeArithmetic(operands []ref.Val) ref.Val {
	if c.native.OperandTrait != 0 && !operands[0].Type().HasTrait(c.native.OperandTrait) {
		return types.NewErrWithNodeID(c.ID(), "no such overload: %s", c.Function())
	}
	var result ref.Val
	if len(operands) == 1 {
		if types.IsBool(operands[0]) {
			return types.NewErrWithNodeID(c.ID(), "no such overload: %s", c.Function())
		}
		result = c.native.Unary(operands[0])
	} else {
		result = c.native.Binary(operands[0], operands[1])
	}
	if failure, ok := result.(*types.Err); ok && failure.NodeID() == 0 {
		return types.LabelErrNode(c.ID(), failure)
	}
	return result
}

func hasMoney(operands []ref.Val) bool {
	for _, operand := range operands {
		if _, ok := operand.(moneyValue); ok {
			return true
		}
	}
	return false
}

func moneyArithmetic(operator string, operands []ref.Val) (bigdecimal.Money, error) {
	if operator == operators.Negate {
		value, err := arithmeticMoney(operands[0])
		if err != nil {
			return bigdecimal.Money{}, err
		}
		return value.Neg()
	}
	left, leftIsMoney := operands[0].(moneyValue)
	right, rightIsMoney := operands[1].(moneyValue)
	switch {
	case leftIsMoney && rightIsMoney:
		return moneyWithMoney(operator, left, right)
	case leftIsMoney:
		return moneyWithNumber(operator, left, operands[1])
	default:
		return numberWithMoney(operator, operands[0], right)
	}
}

func moneyWithMoney(operator string, left, right moneyValue) (bigdecimal.Money, error) {
	a, err := arithmeticMoney(left)
	if err != nil {
		return bigdecimal.Money{}, err
	}
	b, err := arithmeticMoney(right)
	if err != nil {
		return bigdecimal.Money{}, err
	}
	switch operator {
	case operators.Add:
		return a.Add(b)
	case operators.Subtract:
		return a.Sub(b)
	case operators.Multiply:
		return bigdecimal.Money{}, fmt.Errorf("cannot multiply Money by Money")
	default:
		return bigdecimal.Money{}, fmt.Errorf("cannot divide Money by Money")
	}
}

func moneyWithNumber(operator string, left moneyValue, right ref.Val) (bigdecimal.Money, error) {
	a, err := arithmeticMoney(left)
	if err != nil {
		return bigdecimal.Money{}, err
	}
	number, err := arithmeticNumber(right)
	if err != nil {
		return bigdecimal.Money{}, err
	}
	switch operator {
	case operators.Multiply:
		return a.Mul(number)
	case operators.Divide:
		return a.Quo(number)
	}
	b, err := sameCurrency(a, number)
	if err != nil {
		return bigdecimal.Money{}, err
	}
	if operator == operators.Add {
		return a.Add(b)
	}
	return a.Sub(b)
}

func numberWithMoney(operator string, left ref.Val, right moneyValue) (bigdecimal.Money, error) {
	b, err := arithmeticMoney(right)
	if err != nil {
		return bigdecimal.Money{}, err
	}
	number, err := arithmeticNumber(left)
	if err != nil {
		return bigdecimal.Money{}, err
	}
	switch operator {
	case operators.Multiply:
		return b.Mul(number)
	case operators.Divide:
		return bigdecimal.Money{}, fmt.Errorf("cannot divide a number by Money")
	}
	a, err := sameCurrency(b, number)
	if err != nil {
		return bigdecimal.Money{}, err
	}
	if operator == operators.Add {
		return a.Add(b)
	}
	return a.Sub(b)
}

// sameCurrency treats a plain number as an amount in the Money's own currency.
func sameCurrency(money bigdecimal.Money, number bigdecimal.Decimal) (bigdecimal.Money, error) {
	return bigdecimal.NewMoney(bigdecimal.MoneyOptions{Amount: number.String(), Currency: money.Currency(), DecimalPlaces: money.DecimalPlaces()})
}

func arithmeticMoney(value ref.Val) (bigdecimal.Money, error) {
	money := value.(moneyValue)
	if err := money.native.Validate(); err != nil {
		return bigdecimal.Money{}, err
	}
	large, ok := money.native.(bigdecimal.Money)
	if !ok {
		return bigdecimal.Money{}, fmt.Errorf("money arithmetic requires commons/bigdecimal.Money, got %T", money.native)
	}
	return large, nil
}

func arithmeticNumber(value ref.Val) (bigdecimal.Decimal, error) {
	text, err := numericOperand(value)
	if err != nil {
		return bigdecimal.Decimal{}, err
	}
	return bigdecimal.Parse(text)
}
