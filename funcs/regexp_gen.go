// Code generated by gencel. DO NOT EDIT.

package funcs

import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/common/types/ref"

var regexpFindGen = cel.Function("regexp.Find",
	cel.Overload("regexp.Find_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ReFuncs

			result, err := x.Find(args[0].Value(), args[1].Value())
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var regexpFindAllGen = cel.Function("regexp.FindAll",
	cel.Overload("regexp.FindAll_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ReFuncs
			list, err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := x.FindAll(list...)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var regexpMatchGen = cel.Function("regexp.Match",
	cel.Overload("regexp.Match_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ReFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Match(args[0], args[1]))

		}),
	),
)

var regexpQuoteMetaGen = cel.Function("regexp.QuoteMeta",
	cel.Overload("regexp.QuoteMeta_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ReFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.QuoteMeta(args[0]))

		}),
	),
)

var regexpReplaceGen = cel.Function("regexp.Replace",
	cel.Overload("regexp.Replace_interface{}_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType, cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ReFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Replace(args[0], args[1], args[2]))

		}),
	),
)

var regexpReplaceLiteralGen = cel.Function("regexp.ReplaceLiteral",
	cel.Overload("regexp.ReplaceLiteral_interface{}_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ReFuncs

			a0, a1 := x.ReplaceLiteral(args[0], args[1], args[2])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var regexpSplitGen = cel.Function("regexp.Split",
	cel.Overload("regexp.Split_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ReFuncs
			list, err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := x.Split(list...)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)
