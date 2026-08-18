package gomplate

import (
	"math"
	"reflect"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"google.golang.org/protobuf/types/known/structpb"
)

// celResultValue is what an expression hands back to its caller.
//
// A map or list *constructed* inside CEL is held in cel-go's own representation,
// and its Value() is a map[ref.Val]ref.Val -- which marshals to {"Adapter":{}}
// and is useless to a caller that wants to write the result somewhere. A value
// merely *selected* out of the environment was never converted in the first
// place and comes back as whatever Go type it went in as, which is why the
// conversion is applied to aggregates rather than to everything: a registered
// native type must still round-trip as its own struct, not as a map.
func celResultValue(out ref.Val) any {
	// A CEL null's Value() is structpb.NullValue, a protobuf enum whose value is
	// 0. A caller checking for nil sees a number, and a caller writing the result
	// records a zero where the expression said nothing at all.
	if out == types.NullValue {
		return nil
	}

	switch out.(type) {
	case traits.Mapper, traits.Lister:
	default:
		return out.Value()
	}
	// An aggregate is not automatically cel-go's own: a function returning
	// []string, or a map read straight out of the environment, is wrapped in a
	// Lister or Mapper while its Value() stays the Go value it always was.
	// Converting those would widen `"open-source".split("-")` from []string to
	// []any for no gain, so only the representations holding ref.Vals are
	// rewritten.
	if !holdsCELValues(out.Value()) {
		return out.Value()
	}

	native, err := out.ConvertToNative(types.JSONValueType)
	if err != nil {
		// Prior behaviour, deliberately. A value CEL cannot render as JSON is one
		// this function has nothing better to say about, and returning what the
		// caller used to get is strictly no worse than failing the evaluation.
		return out.Value()
	}
	json, ok := native.(*structpb.Value)
	if !ok {
		return out.Value()
	}
	return celWholeNumbers(json.AsInterface())
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

// celWholeNumbers narrows integral floats back to integers. structpb's only
// numeric type is a float64, so a count of two would otherwise come back as
// `2.0` and a schema declaring that field an integer would be describing
// something the value does not look like. A genuinely fractional value -- a
// score of 6.9 -- is left alone.
func celWholeNumbers(value any) any {
	switch typed := value.(type) {
	case float64:
		if typed == math.Trunc(typed) && !math.IsInf(typed, 0) && math.Abs(typed) < 1<<53 {
			return int64(typed)
		}
		return typed
	case []any:
		for index, item := range typed {
			typed[index] = celWholeNumbers(item)
		}
		return typed
	case map[string]any:
		for key, item := range typed {
			typed[key] = celWholeNumbers(item)
		}
		return typed
	default:
		return value
	}
}
