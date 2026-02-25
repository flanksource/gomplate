package coll

import "reflect"

// Coalesce returns the first argument that is neither nil nor empty.
// Empty means: empty string, empty slice, or empty map.
// Zero (0, 0.0) and false are considered valid non-empty values.
// Returns nil when all arguments are nil/empty.
func Coalesce(args ...any) any {
	for _, v := range args {
		if !isEmpty(v) {
			return v
		}
	}
	return nil
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	}
	return false
}
