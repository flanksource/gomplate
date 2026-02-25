package coll

import (
	"reflect"
	"sort"
)

// First returns the first element of a list/slice, the first character of a
// string, or the value at the lexicographically smallest key of a map.
// Returns nil for nil input, "" for an empty string, and nil for an empty
// list or map.
func First(in any) any {
	return indexCollection(in, true)
}

// Last returns the last element of a list/slice, the last character of a
// string, or the value at the lexicographically largest key of a map.
// Returns nil for nil input, "" for an empty string, and nil for an empty
// list or map.
func Last(in any) any {
	return indexCollection(in, false)
}

func indexCollection(in any, first bool) any {
	if in == nil {
		return nil
	}
	switch v := in.(type) {
	case string:
		return indexString(v, first)
	case map[string]any:
		return indexMap(v, first)
	}
	rv := reflect.ValueOf(in)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if rv.Len() == 0 {
			return nil
		}
		if first {
			return rv.Index(0).Interface()
		}
		return rv.Index(rv.Len() - 1).Interface()
	case reflect.String:
		return indexString(rv.String(), first)
	}
	return nil
}

func indexString(s string, first bool) any {
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}
	if first {
		return string(runes[0])
	}
	return string(runes[len(runes)-1])
}

func indexMap(m map[string]any, first bool) any {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if first {
		return m[keys[0]]
	}
	return m[keys[len(keys)-1]]
}
