package gomplate

import (
	"reflect"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

// CELValue unwraps CEL-created aggregates without converting numbers through
// JSON. Host-provided native objects retain their original type.
func CELValue(out ref.Val) any {
	if out == types.NullValue {
		return nil
	}

	if !holdsCELValues(out.Value()) {
		return out.Value()
	}
	switch value := out.(type) {
	case traits.Lister:
		rows := make([]any, 0, int(value.Size().(types.Int)))
		for iterator := value.Iterator(); iterator.HasNext() == types.True; {
			rows = append(rows, CELValue(iterator.Next()))
		}
		return rows
	case traits.Mapper:
		strings, keys := map[string]any{}, map[any]any{}
		for iterator := value.Iterator(); iterator.HasNext() == types.True; {
			key := iterator.Next()
			nativeKey, nativeValue := CELValue(key), CELValue(value.Get(key))
			keys[nativeKey] = nativeValue
			if name, ok := nativeKey.(string); ok {
				strings[name] = nativeValue
			}
		}
		if len(strings) == len(keys) {
			return strings
		}
		return keys
	default:
		return out.Value()
	}
}

var refValType = reflect.TypeOf((*ref.Val)(nil)).Elem()

// holdsCELValues reports the aggregates cel-go built itself, which are the ones
// carrying ref.Vals rather than Go values. A value that is neither a Go map nor
// a Go slice is one of cel-go's own structs -- the map literal whose Value() is
// a mapAccessor, and which marshals to {"Adapter":{}}.
func holdsCELValues(value any) bool {
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Slice, reflect.Array:
		return reflected.Type().Elem().Implements(refValType)
	case reflect.Map:
		return reflected.Type().Key().Implements(refValType) ||
			reflected.Type().Elem().Implements(refValType)
	default:
		return true
	}
}
