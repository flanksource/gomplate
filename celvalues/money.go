package celvalues

import (
	"fmt"
	"math"
	"reflect"
	"strconv"

	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
	"github.com/flanksource/commons/bigdecimal"
	"github.com/flanksource/commons/decimal"
)

var moneyType = types.NewObjectType("decimal.Money", traits.ComparerType)
var bigMoneyType = types.NewObjectType("bigdecimal.Money", traits.ComparerType)

type currencyValue interface {
	String() string
	Currency() string
	Validate() error
}
type moneyValue struct{ native currencyValue }

func (m moneyValue) Value() any     { return m.native }
func (m moneyValue) String() string { return m.native.String() }
func (m moneyValue) Type() ref.Type {
	if _, ok := m.native.(bigdecimal.Money); ok {
		return bigMoneyType
	}
	return moneyType
}
func (m moneyValue) ConvertToNative(target reflect.Type) (any, error) { return scalarNative(m, target) }
func (m moneyValue) ConvertToType(target ref.Type) ref.Val {
	if err := m.native.Validate(); err != nil {
		return types.WrapErr(err)
	}
	switch target {
	case m.Type():
		return m
	case types.TypeType:
		if m.Type() == bigMoneyType {
			return bigMoneyType
		}
		return moneyType
	case types.StringType:
		return types.String(m.String())
	default:
		return types.NewErr("cannot convert Money to %s", target.TypeName())
	}
}
func (m moneyValue) Equal(other ref.Val) ref.Val {
	if other == types.NullValue {
		return types.False
	}
	return comparisonVerdict("_==_", m.Compare(other))
}
func (m moneyValue) Compare(other ref.Val) ref.Val {
	if err := m.native.Validate(); err != nil {
		return types.WrapErr(err)
	}
	if peer, ok := other.(moneyValue); ok {
		switch left := m.native.(type) {
		case decimal.Money:
			if right, ok := peer.native.(decimal.Money); ok {
				v, err := left.Cmp(right)
				if err != nil {
					return types.WrapErr(err)
				}
				return types.Int(v)
			}
		case bigdecimal.Money:
			if right, ok := peer.native.(bigdecimal.Money); ok {
				v, err := left.Cmp(right)
				if err != nil {
					return types.WrapErr(err)
				}
				return types.Int(v)
			}
		}
		return types.NewErr("cannot compare Money with different decimal precision types")
	}
	text, err := numericOperand(other)
	if err != nil {
		return types.WrapErr(err)
	}
	switch left := m.native.(type) {
	case decimal.Money:
		right, err := decimal.Parse(text)
		if err != nil {
			return types.WrapErr(err)
		}
		return types.Int(left.Amount().Cmp(right))
	case bigdecimal.Money:
		right, err := bigdecimal.Parse(text)
		if err != nil {
			return types.WrapErr(err)
		}
		if right.IsNaN() || right.IsInf(0) {
			return types.NewErr("Money comparison requires a finite number")
		}
		return types.Int(left.Amount().Compare(right))
	default:
		return types.NewErr("unsupported Money type %T", m.native)
	}
}
func numericOperand(other ref.Val) (string, error) {
	switch v := other.(type) {
	case types.Int:
		return strconv.FormatInt(int64(v), 10), nil
	case types.Uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case types.Double:
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return "", fmt.Errorf("money operand must be a finite number")
		}
		return strconv.FormatFloat(float64(v), 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("money operand expects money or a number, got %s", other.Type().TypeName())
	}
}
