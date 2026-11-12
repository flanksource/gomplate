package celvalues

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"

	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
	"github.com/flanksource/commons/bigdecimal"
	"github.com/flanksource/commons/decimal"
	"github.com/google/uuid"
	"github.com/ohler55/ojg/alt"
	"google.golang.org/protobuf/proto"
)

// AsMapper is implemented by host values that present themselves to CEL as a map.
type AsMapper interface {
	AsMap(fields ...string) map[string]any
}

// hostValue converts the Go values cel-go's own adapter has no CEL form for, lazily:
// a struct becomes a map of its fields only when an expression reaches it.
func (a moneyAdapter) hostValue(value any) (ref.Val, bool) {
	switch v := value.(type) {
	case uuid.UUID:
		return types.String(v.String()), true
	case *uuid.UUID:
		if v == nil {
			return types.NullValue, true
		}
		return types.String(v.String()), true
	case AsMapper:
		return a.NativeToValue(v.AsMap()), true
	case alt.Simplifier:
		return a.NativeToValue(v.Simplify()), true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Pointer:
		if reflected.IsNil() {
			return types.NullValue, true
		}
		return a.NativeToValue(reflected.Elem().Interface()), true
	case reflect.Struct:
		return hostMap{Mapper: types.NewStringInterfaceMap(a, structFields(reflected)), host: value}, true
	case reflect.Map:
		return hostMap{Mapper: a.hostMapper(value), host: value}, true
	case reflect.Slice, reflect.Array:
		if reflected.Type().Elem().Kind() == reflect.Uint8 {
			return nil, false
		}
		return hostList{Lister: types.NewDynamicList(a, value), host: value}, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// JSON-shaped hosts carry counts as unsigned; CEL arithmetic needs them as ints.
		if unsigned := reflected.Uint(); unsigned <= math.MaxInt64 {
			return types.Int(int64(unsigned)), true
		}
	}
	return nil, false
}

// hostMapper uses cel-go's typed map for the common JSON-shaped maps, reflection otherwise.
func (a moneyAdapter) hostMapper(value any) traits.Mapper {
	switch v := value.(type) {
	case map[string]any:
		return types.NewStringInterfaceMap(a, v)
	case map[string]string:
		return types.NewStringStringMap(a, v)
	}
	return types.NewDynamicMap(a, value)
}

// hostMap and hostList read a host map, list or struct lazily. Their Value is the plain Go
// form, so a function or result given one receives maps and lists, never the host's structs.
type hostMap struct {
	traits.Mapper
	host any
}

func (m hostMap) Value() any        { return Plain(m.host) }
func (m hostMap) IsZeroValue() bool { return m.Size() == types.IntZero }

type hostList struct {
	traits.Lister
	host any
}

func (l hostList) Value() any        { return Plain(l.host) }
func (l hostList) IsZeroValue() bool { return l.Size() == types.IntZero }

// Plain converts a host value to the Go form CEL reads it as: structs become maps keyed like
// encoding/json, maps become map[string]any, lists []any, integers int64 and UUIDs strings.
// Dates, durations, Money, bytes and CEL values are kept.
func Plain(value any) any {
	switch v := value.(type) {
	case nil, string, bool, int64, float64, []byte, time.Time, time.Duration,
		decimal.Money, bigdecimal.Money, *decimal.Money, *bigdecimal.Money, proto.Message:
		return value
	case hostMap:
		return Plain(v.host)
	case hostList:
		return Plain(v.host)
	case ref.Val:
		return value
	case uuid.UUID:
		return v.String()
	case *uuid.UUID:
		if v == nil {
			return nil
		}
		return v.String()
	case AsMapper:
		return Plain(v.AsMap())
	case alt.Simplifier:
		return Plain(v.Simplify())
	}
	return plainReflected(reflect.ValueOf(value))
}

func plainReflected(reflected reflect.Value) any {
	switch reflected.Kind() {
	case reflect.Pointer, reflect.Interface:
		if reflected.IsNil() {
			return nil
		}
		return Plain(reflected.Elem().Interface())
	case reflect.Struct:
		fields := structFields(reflected)
		for key, field := range fields {
			fields[key] = Plain(field)
		}
		return fields
	case reflect.Map:
		entries := make(map[string]any, reflected.Len())
		for iterator := reflected.MapRange(); iterator.Next(); {
			entries[fmt.Sprint(iterator.Key().Interface())] = Plain(iterator.Value().Interface())
		}
		return entries
	case reflect.Slice, reflect.Array:
		if reflected.Kind() == reflect.Slice && reflected.IsNil() {
			return []any(nil)
		}
		items := make([]any, reflected.Len())
		for i := range items {
			items[i] = Plain(reflected.Index(i).Interface())
		}
		return items
	case reflect.String:
		return reflected.String()
	case reflect.Bool:
		return reflected.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflected.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if unsigned := reflected.Uint(); unsigned <= math.MaxInt64 {
			return int64(unsigned)
		}
		return reflected.Uint()
	case reflect.Float32, reflect.Float64:
		return reflected.Float()
	}
	return reflected.Interface()
}

type structField struct {
	key       string
	index     []int
	omitEmpty bool
}

var structFieldCache sync.Map // reflect.Type -> []structField

// structFields returns a struct's fields keyed the way encoding/json names them: json tag
// names, `-` and unexported fields skipped, omitempty zero values absent, and the fields of
// anonymous structs promoted. The values stay native; they are converted as they are read.
func structFields(value reflect.Value) map[string]any {
	fields := map[string]any{}
	for _, field := range fieldsOf(value.Type()) {
		fieldValue, err := value.FieldByIndexErr(field.index)
		if err != nil {
			continue // promoted through a nil embedded pointer
		}
		if field.omitEmpty && isEmptyValue(fieldValue) {
			continue
		}
		fields[field.key] = fieldValue.Interface()
	}
	return fields
}

// Field reads one field of a struct by the key encoding/json names it with, following the same
// rules as the CEL view of the struct except omitempty: a zero field is still present.
func Field(value reflect.Value, key string) (reflect.Value, bool) {
	for _, field := range fieldsOf(value.Type()) {
		if field.key != key {
			continue
		}
		fieldValue, err := value.FieldByIndexErr(field.index)
		if err != nil {
			return reflect.Value{}, false // promoted through a nil embedded pointer
		}
		return fieldValue, true
	}
	return reflect.Value{}, false
}

// isEmptyValue is encoding/json's omitempty rule: a struct is never empty.
func isEmptyValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Interface, reflect.Pointer:
		return value.IsZero()
	}
	return false
}

// fieldsOf caches a struct type's fields; where a promoted name clashes, the shallowest field wins.
func fieldsOf(structType reflect.Type) []structField {
	if cached, found := structFieldCache.Load(structType); found {
		return cached.([]structField)
	}
	shallowest := map[string]structField{}
	var order []string
	for _, field := range collectFields(structType, nil) {
		existing, found := shallowest[field.key]
		if !found {
			order = append(order, field.key)
		}
		if !found || len(field.index) < len(existing.index) {
			shallowest[field.key] = field
		}
	}
	fields := make([]structField, len(order))
	for i, key := range order {
		fields[i] = shallowest[key]
	}
	structFieldCache.Store(structType, fields)
	return fields
}

func collectFields(structType reflect.Type, parent []int) []structField {
	var fields []structField
	for i := range structType.NumField() {
		field := structType.Field(i)
		index := append(append([]int(nil), parent...), i)
		tag, tagged := field.Tag.Lookup("json")
		name, options, _ := strings.Cut(tag, ",")
		if name == "-" && options == "" {
			continue
		}
		embedded := field.Type
		if embedded.Kind() == reflect.Pointer {
			embedded = embedded.Elem()
		}
		if field.Anonymous && name == "" && embedded.Kind() == reflect.Struct {
			fields = append(fields, collectFields(embedded, index)...)
			continue
		}
		if !field.IsExported() {
			continue
		}
		if name == "" || !tagged {
			name = field.Name
		}
		fields = append(fields, structField{key: name, index: index, omitEmpty: strings.Contains(","+options+",", ",omitempty,")})
	}
	return fields
}
