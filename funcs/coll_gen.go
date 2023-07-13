// Code generated by gencel. DO NOT EDIT.

package funcs

import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/common/types/ref"

var collSliceGen = cel.Function("Slice",
	cel.Overload("Slice_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			return types.DefaultTypeAdapter.NativeToValue(x.Slice(list...))

		}),
	),
)

var collHasGen = cel.Function("Has",
	cel.Overload("Has_interface{}_string",

		[]*cel.Type{
			cel.DynType, cel.StringType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Has(args[0], args[1].Value().(string)))

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

			var x CollFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.Dict(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collKeysGen = cel.Function("Keys",
	cel.Overload("Keys_map[string]interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs
			list := transferSlice[map[string]interface{}](args[0].(ref.Val))

			a0, a1 := x.Keys(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collValuesGen = cel.Function("Values",
	cel.Overload("Values_map[string]interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs
			list := transferSlice[map[string]interface{}](args[0].(ref.Val))

			a0, a1 := x.Values(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

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

			var x CollFuncs

			a0, a1 := x.Append(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

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

			var x CollFuncs

			a0, a1 := x.Prepend(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collUniqGen = cel.Function("Uniq",
	cel.Overload("Uniq_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs

			a0, a1 := x.Uniq(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collReverseGen = cel.Function("Reverse",
	cel.Overload("Reverse_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs

			a0, a1 := x.Reverse(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collMergeGen = cel.Function("Merge",
	cel.Overload("Merge_map[string]interface{}_map[string]interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs
			list := transferSlice[map[string]interface{}](args[1].(ref.Val))

			a0, a1 := x.Merge(args[0].Value().(map[string]interface{}), list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collSortGen = cel.Function("Sort",
	cel.Overload("Sort_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.Sort(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collJQGen = cel.Function("JQ",
	cel.Overload("JQ_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs

			a0, a1 := x.JQ(args[0].Value().(string), args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collFlattenGen = cel.Function("Flatten",
	cel.Overload("Flatten_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.Flatten(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collPickGen = cel.Function("Pick",
	cel.Overload("Pick_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.Pick(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var collOmitGen = cel.Function("Omit",
	cel.Overload("Omit_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x CollFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.Omit(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)
