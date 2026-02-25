package nilsafe

import (
	"time"

	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/cel-go/interpreter"
)

func zeroValueFor(t ref.Type) ref.Val {
	switch t {
	case types.IntType:
		return types.IntZero
	case types.UintType:
		return types.Uint(0)
	case types.DoubleType:
		return types.Double(0)
	case types.BoolType:
		return types.False
	case types.StringType:
		return types.String("")
	case types.BytesType:
		return types.Bytes{}
	case types.DurationType:
		return types.Duration{Duration: 0}
	case types.TimestampType:
		return types.Timestamp{Time: time.Unix(0, 0)}
	default:
		return types.NullValue
	}
}

func inferTypeFromPeers(vals []ref.Val, nullIdx int) ref.Type {
	for i, v := range vals {
		if i != nullIdx && v != types.NullValue {
			return v.Type()
		}
	}
	return types.NullType
}

type zeroValueCall struct {
	interpreter.InterpretableCall
}

func (c *zeroValueCall) Eval(ctx interpreter.Activation) ref.Val {
	args := c.InterpretableCall.Args()
	vals := make([]ref.Val, len(args))
	hasNull := false
	for i, arg := range args {
		vals[i] = arg.Eval(ctx)
		if vals[i] == types.NullValue {
			hasNull = true
		}
	}
	if !hasNull {
		return c.InterpretableCall.Eval(ctx)
	}

	fn := c.InterpretableCall.Function()
	if !isOperator(fn) {
		return types.NullValue
	}

	for i, v := range vals {
		if v == types.NullValue {
			vals[i] = zeroValueFor(inferTypeFromPeers(vals, i))
		}
	}
	return dispatchOp(fn, vals)
}

type zeroValueEq struct {
	interpreter.InterpretableCall
}

func (c *zeroValueEq) Eval(ctx interpreter.Activation) ref.Val {
	args := c.InterpretableCall.Args()
	lhs, rhs := args[0].Eval(ctx), args[1].Eval(ctx)

	lNull := lhs == types.NullValue
	rNull := rhs == types.NullValue
	if !lNull && !rNull {
		return c.InterpretableCall.Eval(ctx)
	}
	if lNull && rNull {
		return types.True
	}

	if lNull {
		lhs = zeroValueFor(rhs.Type())
	} else {
		rhs = zeroValueFor(lhs.Type())
	}
	return lhs.Equal(rhs)
}

func (c *zeroValueEq) Function() string   { return operators.Equals }
func (c *zeroValueEq) OverloadID() string { return "" }
func (c *zeroValueEq) Args() []interpreter.Interpretable {
	return c.InterpretableCall.Args()
}

type zeroValueNe struct {
	interpreter.InterpretableCall
}

func (c *zeroValueNe) Eval(ctx interpreter.Activation) ref.Val {
	args := c.InterpretableCall.Args()
	lhs, rhs := args[0].Eval(ctx), args[1].Eval(ctx)

	lNull := lhs == types.NullValue
	rNull := rhs == types.NullValue
	if !lNull && !rNull {
		return c.InterpretableCall.Eval(ctx)
	}
	if lNull && rNull {
		return types.False
	}

	if lNull {
		lhs = zeroValueFor(rhs.Type())
	} else {
		rhs = zeroValueFor(lhs.Type())
	}

	eq := lhs.Equal(rhs)
	if eq == types.True {
		return types.False
	}
	return types.True
}

func (c *zeroValueNe) Function() string   { return operators.NotEquals }
func (c *zeroValueNe) OverloadID() string { return "" }
func (c *zeroValueNe) Args() []interpreter.Interpretable {
	return c.InterpretableCall.Args()
}

func isOperator(fn string) bool {
	switch fn {
	case operators.Add, operators.Subtract, operators.Multiply,
		operators.Divide, operators.Modulo, operators.Less,
		operators.LessEquals, operators.Greater, operators.GreaterEquals,
		operators.Negate, operators.LogicalAnd, operators.LogicalOr:
		return true
	}
	return false
}

func dispatchOp(fn string, vals []ref.Val) ref.Val {
	switch fn {
	case operators.Add:
		if a, ok := vals[0].(traits.Adder); ok {
			return a.Add(vals[1])
		}
	case operators.Subtract:
		if a, ok := vals[0].(traits.Subtractor); ok {
			return a.Subtract(vals[1])
		}
	case operators.Multiply:
		if a, ok := vals[0].(traits.Multiplier); ok {
			return a.Multiply(vals[1])
		}
	case operators.Divide:
		if a, ok := vals[0].(traits.Divider); ok {
			return a.Divide(vals[1])
		}
	case operators.Modulo:
		if a, ok := vals[0].(traits.Modder); ok {
			return a.Modulo(vals[1])
		}
	case operators.Less:
		if a, ok := vals[0].(traits.Comparer); ok {
			return types.Bool(a.Compare(vals[1]) == types.IntNegOne)
		}
	case operators.LessEquals:
		if a, ok := vals[0].(traits.Comparer); ok {
			cmp := a.Compare(vals[1])
			return types.Bool(cmp == types.IntNegOne || cmp == types.IntZero)
		}
	case operators.Greater:
		if a, ok := vals[0].(traits.Comparer); ok {
			return types.Bool(a.Compare(vals[1]) == types.IntOne)
		}
	case operators.GreaterEquals:
		if a, ok := vals[0].(traits.Comparer); ok {
			cmp := a.Compare(vals[1])
			return types.Bool(cmp == types.IntOne || cmp == types.IntZero)
		}
	case operators.Negate:
		if a, ok := vals[0].(traits.Negater); ok {
			return a.Negate()
		}
	}
	return types.NewErr("unsupported operator: %s", fn)
}
