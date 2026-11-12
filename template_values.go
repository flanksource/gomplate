package gomplate

import (
	"reflect"
	"time"

	"cel.dev/cel-go/common/types/ref"
	"github.com/flanksource/commons/bigdecimal"
	"github.com/flanksource/commons/decimal"
	"github.com/flanksource/gomplate/v4/celvalues"
	"github.com/flanksource/gomplate/v4/texttemplate"
	"github.com/ohler55/ojg/alt"
	"google.golang.org/protobuf/proto"
)

// DefaultValuesProvider exposes a go template environment without serializing it up front:
// struct fields are read by their encoding/json key as the template reaches them, and only the
// values a template prints, ranges over, tests or passes to a function are serialized, exactly as
// Serialize would have converted them.
type DefaultValuesProvider struct{}

var _ texttemplate.ValuesProvider = DefaultValuesProvider{}

func (DefaultValuesProvider) Field(receiver reflect.Value, name string) (reflect.Value, bool, error) {
	if !receiver.CanInterface() {
		return reflect.Value{}, false, nil
	}
	switch v := receiver.Interface().(type) {
	case AsMapper:
		return reflect.ValueOf(v.AsMap()[name]), true, nil
	case alt.Simplifier:
		simplified := reflect.ValueOf(v.Simplify())
		if simplified.Kind() != reflect.Map {
			return reflect.Value{}, false, nil
		}
		return simplified.MapIndex(reflect.ValueOf(name)), true, nil
	}
	switch {
	case keepsNative(receiver):
		return reflect.Value{}, false, nil
	case receiver.Kind() == reflect.Map && receiver.Type().Key().Kind() == reflect.String:
		// Serialized, every map was a map[string]any: a missing key is missing, never the zero element.
		return receiver.MapIndex(reflect.ValueOf(name).Convert(receiver.Type().Key())), true, nil
	case receiver.Kind() == reflect.Struct:
		value, found := celvalues.Field(receiver, name)
		return value, found, nil
	}
	return reflect.Value{}, false, nil
}

func (DefaultValuesProvider) Value(v reflect.Value) (reflect.Value, error) {
	if !v.IsValid() || plainKind(v.Kind()) || keepsNative(v) {
		return v, nil
	}
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v, nil
		}
		return DefaultValuesProvider{}.Value(v.Elem())
	}
	out, err := Serialize(map[string]any{"v": v.Interface()})
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(out["v"]), nil
}

func plainKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Func, reflect.Chan:
		return true
	}
	return false
}

// keepsNative reports the values Serialize hands to templates unchanged.
func keepsNative(v reflect.Value) bool {
	if !v.CanInterface() {
		return false
	}
	switch v.Interface().(type) {
	case time.Time, time.Duration, ref.Val, proto.Message,
		decimal.Money, bigdecimal.Money, *decimal.Money, *bigdecimal.Money:
		return true
	}
	return false
}
