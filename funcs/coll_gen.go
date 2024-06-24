// Code generated by gencel. DO NOT EDIT.

package funcs

import (
	"context"
	"reflect"
	"sort"

	"encoding/json"

	"github.com/flanksource/gomplate/v3/coll"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var typeMapStringAny = reflect.TypeOf(map[string]interface{}{})

// NOTE: Slice Makes no sense in cel-go because
// cel-go doesn't support variadic functions.
// https://github.com/google/cel-go/issues/476#issuecomment-1029172709
// var collSliceGen = cel.Function("Slice",
// 	cel.Overload("Slice_interface{}",

// 		[]*cel.Type{
// 			cel.DynType,
// 		},
// 		cel.DynType,
// 		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
// 			list, err := sliceToNative[interface{}](args[0].(ref.Val))
// 			if err != nil {
// 				return types.WrapErr(err)
// 			}

// 			return types.DefaultTypeAdapter.NativeToValue(coll.Slice(list...))

// 		}),
// 	),
// )

var collHasGen = cel.Function("Has",
	cel.Overload("Has_interface{}_any",
		[]*cel.Type{
			cel.DynType, cel.StringType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			list, err := args[0].ConvertToNative(reflect.TypeOf([]string{}))
			if err != nil {
				return types.WrapErr(err)
			}

			return types.DefaultTypeAdapter.NativeToValue(coll.Has(list, args[1].Value().(string)))
		}),
	),
)

var collDictGen = cel.Function("Dict",
	cel.Overload("Dict_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			list, err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := coll.Dict(list...)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var collKeysGen = cel.Function("keys",
	cel.MemberOverload("map_keys",

		[]*cel.Type{
			cel.MapType(cel.StringType, cel.AnyType),
		},
		cel.ListType(cel.StringType),
		cel.UnaryBinding(func(arg ref.Val) ref.Val {
			_map, err := convertMap(arg)
			if err != nil {
				return types.WrapErr(err)
			}
			var result []string
			for k := range _map {
				result = append(result, k)
			}
			sort.Strings(result)
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var collValuesGen = cel.Function("values",
	cel.MemberOverload("map_values",

		[]*cel.Type{
			cel.MapType(cel.StringType, cel.AnyType),
		},
		cel.ListType(cel.AnyType),
		cel.UnaryBinding(func(arg ref.Val) ref.Val {
			_map, err := convertMap(arg)
			if err != nil {
				return types.WrapErr(err)
			}

			var result []any

			for _, v := range _map {
				result = append(result, v)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var collAppendGen = cel.Function("Append",
	cel.Overload("Append_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			result, err := coll.Append(args[0], args[1])
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var collPrependGen = cel.Function("Prepend",
	cel.Overload("Prepend_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			result, err := coll.Prepend(args[0], args[1])
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var collUniqGen = cel.Function("uniq",
	cel.MemberOverload("uniq_interface{}",
		[]*cel.Type{
			cel.ListType(cel.DynType),
		},
		cel.ListType(cel.DynType),
		cel.UnaryBinding(func(arg ref.Val) ref.Val {
			list, _ := arg.ConvertToNative(reflect.TypeOf([]any{}))

			result, err := coll.Uniq(list)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var collReverseGen = cel.Function("reverse",
	cel.MemberOverload("Reverse_interface{}",

		[]*cel.Type{
			cel.ListType(cel.AnyType),
		},
		cel.ListType(cel.AnyType),
		cel.UnaryBinding(func(arg ref.Val) ref.Val {
			list, err := sliceToNative[any](arg)
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := coll.Reverse(list)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var collMergeGen = cel.Function("merge",
	cel.MemberOverload("merge_map[string]interface{}",
		[]*cel.Type{
			cel.MapType(cel.StringType, cel.DynType),
			cel.MapType(cel.StringType, cel.DynType),
		},
		cel.MapType(cel.StringType, cel.DynType),
		cel.BinaryBinding(func(into, from ref.Val) ref.Val {
			_into, err := into.ConvertToNative(typeMapStringAny)
			if err != nil {
				return types.WrapErr(err)
			}

			_from, err := from.ConvertToNative(typeMapStringAny)
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := coll.Merge(_from.(map[string]interface{}), _into.(map[string]interface{}))
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var collSortGen = cel.Function("sort",
	cel.MemberOverload("sort_string_interface{}",
		[]*cel.Type{
			cel.ListType(cel.DynType),
		},
		cel.ListType(cel.DynType),
		cel.UnaryBinding(func(arg ref.Val) ref.Val {
			list, err := sliceToNative[any](arg)
			if err != nil {
				return types.WrapErr(err)
			}

			sorted, err := coll.Sort("", list)
			if err != nil {
				return types.WrapErr(err)
			}

			return types.DefaultTypeAdapter.NativeToValue(sorted)
		}),
	),
)

var collSortByGen = cel.Function("sortBy",
	cel.MemberOverload("sort_interface{}",

		[]*cel.Type{
			cel.ListType(cel.AnyType),
			cel.StringType,
		},
		cel.ListType(cel.AnyType),
		cel.BinaryBinding(func(arg ref.Val, key ref.Val) ref.Val {
			var x CollFuncs

			list, err := sliceToNative[any](arg)
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := x.Sort(list, key.Value().(string))
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var collJQGen = cel.Function("jq",
	cel.Overload("jq_string_interface{}",
		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.BinaryBinding(func(query, object ref.Val) ref.Val {

			var input interface{}
			switch v := object.Value().(type) {
			case string, []byte:
				input = v
			default:
				data, err := json.Marshal(v)
				if err != nil {
					return types.WrapErr(err)
				}
				if err := json.Unmarshal(data, &input); err != nil {
					return types.WrapErr(err)
				}
			}

			result, err := coll.JQ(context.Background(), query.Value().(string), input)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var collFlattenGen = cel.Function("flatten",
	cel.MemberOverload("flatten_list{}",

		[]*cel.Type{
			cel.ListType(cel.AnyType),
		},
		cel.ListType(cel.AnyType),
		cel.UnaryBinding(func(arg ref.Val) ref.Val {

			var x CollFuncs
			list, err := sliceToNative[any](arg)
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := x.Flatten(list)

			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var collPickGen = cel.Function("pick",
	cel.MemberOverload("pick_interface{}",
		[]*cel.Type{
			cel.AnyType, cel.ListType(cel.StringType),
		},
		cel.AnyType,
		cel.OverloadIsNonStrict(),
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			m, err := args[0].ConvertToNative(typeMapStringAny)
			if err != nil {
				return types.WrapErr(err)
			}

			list, err := sliceToNative[string](args[1:]...)
			if err != nil {
				return types.WrapErr(err)
			}
			result := coll.Pick(m.(map[string]interface{}), list...)

			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var collOmitGen = cel.Function("omit",
	cel.MemberOverload("omit_interface{}",
		[]*cel.Type{
			cel.MapType(cel.StringType, cel.AnyType),
			cel.ListType(cel.StringType),
		},
		cel.MapType(cel.StringType, cel.AnyType),
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			m, err := args[0].ConvertToNative(typeMapStringAny)
			if err != nil {
				return types.WrapErr(err)
			}

			list, err := sliceToNative[string](args[1:]...)
			if err != nil {
				return types.WrapErr(err)
			}

			result := coll.Omit(m.(map[string]interface{}), list...)

			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)
