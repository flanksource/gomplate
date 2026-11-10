package celvalues

import (
	"reflect"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
	"github.com/flanksource/commons/bigdecimal"
	"github.com/flanksource/commons/decimal"
)

// NewAdapter extends an existing CEL adapter with Commons Money support.
// Use this adapter for custom function results as well as host input.
func NewAdapter(base types.Adapter) types.Adapter {
	if base == nil {
		panic("CEL base adapter must not be nil")
	}
	return moneyAdapter{base: base}
}

type moneyAdapter struct{ base types.Adapter }

type moneyDescriptor struct {
	*types.Type
	native reflect.Type
}

func (d moneyDescriptor) ReflectType() reflect.Type                     { return d.native }
func (d moneyDescriptor) FieldNames() []string                          { return nil }
func (d moneyDescriptor) FindFieldType(string) (*types.FieldType, bool) { return nil, false }
func (d moneyDescriptor) NewValue(types.Adapter, map[string]ref.Val) ref.Val {
	return types.NewErr("construct Money with commons NewMoney")
}
func (d moneyDescriptor) Adapt(base types.Adapter, value any) ref.Val {
	return NewAdapter(base).NativeToValue(value)
}
func moneyTypes() cel.EnvOption {
	return cel.Types(moneyDescriptor{Type: types.NewObjectType(moneyType.TypeName(), traits.FieldTesterType, traits.IndexerType), native: reflect.TypeFor[decimal.Money]()}, moneyDescriptor{Type: types.NewObjectType(bigMoneyType.TypeName(), traits.FieldTesterType, traits.IndexerType), native: reflect.TypeFor[bigdecimal.Money]()})
}

func (a moneyAdapter) NativeToValue(value any) ref.Val {
	switch v := value.(type) {
	case decimal.Money:
		return moneyValue{native: v}
	case bigdecimal.Money:
		return moneyValue{native: v}
	case *decimal.Money:
		if v == nil {
			return types.NullValue
		}
		return moneyValue{native: *v}
	case *bigdecimal.Money:
		if v == nil {
			return types.NullValue
		}
		return moneyValue{native: *v}
	case ref.Val:
		return v
	}
	if value != nil {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Map:
			return types.NewDynamicMap(a, value)
		case reflect.Slice, reflect.Array:
			if _, ok := value.([]byte); !ok {
				return types.NewDynamicList(a, value)
			}
		}
	}
	return a.base.NativeToValue(value)
}
