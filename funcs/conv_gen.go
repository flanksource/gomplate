// Code generated by gencel. DO NOT EDIT.

package funcs

import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/common/types/ref"

var convBoolGen = cel.Function("conv.Bool",
	cel.Overload("conv.Bool_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Bool(args[0]))

		}),
	),
)

var convToBoolGen = cel.Function("conv.ToBool",
	cel.Overload("conv.ToBool_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToBool(args[0]))

		}),
	),
)

var convToBoolsGen = cel.Function("conv.ToBools",
	cel.Overload("conv.ToBools_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs
			list,err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}

			return types.DefaultTypeAdapter.NativeToValue(x.ToBools(list...))

		}),
	),
)

var convSliceGen = cel.Function("conv.Slice",
	cel.Overload("conv.Slice_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs
			list, err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}

			return types.DefaultTypeAdapter.NativeToValue(x.Slice(list...))

		}),
	),
)

var convHasGen = cel.Function("conv.Has",
	cel.Overload("conv.Has_interface{}_string",

		[]*cel.Type{
			cel.DynType, cel.StringType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Has(args[0], args[1].Value().(string)))

		}),
	),
)


var convToInt64Gen = cel.Function("conv.ToInt64",
	cel.Overload("conv.ToInt64_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.IntType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToInt64(args[0]))

		}),
	),
)

var convToIntGen = cel.Function("conv.ToInt",
	cel.Overload("conv.ToInt_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.IntType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToInt(args[0]))

		}),
	),
)

var convToInt64sGen = cel.Function("conv.ToInt64s",
	cel.Overload("conv.ToInt64s_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs
			list, err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}


			return types.DefaultTypeAdapter.NativeToValue(x.ToInt64s(list...))

		}),
	),
)

var convToIntsGen = cel.Function("conv.ToInts",
	cel.Overload("conv.ToInts_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs
			list, err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}


			return types.DefaultTypeAdapter.NativeToValue(x.ToInts(list...))

		}),
	),
)

var convToFloat64Gen = cel.Function("conv.ToFloat64",
	cel.Overload("conv.ToFloat64_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DoubleType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToFloat64(args[0]))

		}),
	),
)

var convToFloat64sGen = cel.Function("conv.ToFloat64s",
	cel.Overload("conv.ToFloat64s_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs
			list, err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}


			return types.DefaultTypeAdapter.NativeToValue(x.ToFloat64s(list...))

		}),
	),
)

var convToStringGen = cel.Function("conv.ToString",
	cel.Overload("conv.ToString_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToString(args[0]))

		}),
	),
)

var convToStringsGen = cel.Function("conv.ToStrings",
	cel.Overload("conv.ToStrings_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs
			list, err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(x.ToStrings(list...))

		}),
	),
)

var convDefaultGen = cel.Function("conv.Default",
	cel.Overload("conv.Default_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Default(args[0], args[1]))

		}),
	),
)

