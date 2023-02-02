// Code generated by "gencel";
// DO NOT EDIT.

package funcs

import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types/ref"

var BoolconvGen = cel.Function("Bool",
	cel.Overload("Bool_interface{}",

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

var ToBoolconvGen = cel.Function("ToBool",
	cel.Overload("ToBool_interface{}",

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

var ToBoolsconvGen = cel.Function("ToBools",
	cel.Overload("ToBools_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToBools(args[0]))

		}),
	),
)

var SliceconvGen = cel.Function("Slice",
	cel.Overload("Slice_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Slice(args[0]))

		}),
	),
)

var JoinconvGen = cel.Function("Join",
	cel.Overload("Join_interface{}_string",

		[]*cel.Type{
			cel.DynType, cel.StringType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			// Need to figure this out
			name := "Flanksource"
			return types.DefaultTypeAdapter.NativeToValue([]string{name, name + "suffix"})

		}),
	),
)

var HasconvGen = cel.Function("Has",
	cel.Overload("Has_interface{}_string",

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

var ParseIntconvGen = cel.Function("ParseInt",
	cel.Overload("ParseInt_interface{}_int_int",

		[]*cel.Type{
			cel.DynType, cel.IntType, cel.IntType,
		},
		cel.IntType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ParseInt(args[0], args[1].Value().(int), args[2].Value().(int)))

		}),
	),
)

var ParseFloatconvGen = cel.Function("ParseFloat",
	cel.Overload("ParseFloat_interface{}_int",

		[]*cel.Type{
			cel.DynType, cel.IntType,
		},
		cel.DoubleType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ParseFloat(args[0], args[1].Value().(int)))

		}),
	),
)

var ParseUintconvGen = cel.Function("ParseUint",
	cel.Overload("ParseUint_interface{}_int_int",

		[]*cel.Type{
			cel.DynType, cel.IntType, cel.IntType,
		},
		cel.UintType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ParseUint(args[0], args[1].Value().(int), args[2].Value().(int)))

		}),
	),
)

var AtoiconvGen = cel.Function("Atoi",
	cel.Overload("Atoi_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.IntType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Atoi(args[0]))

		}),
	),
)

var URLconvGen = cel.Function("URL",
	cel.Overload("URL_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			// Need to figure this out
			name := "Flanksource"
			return types.DefaultTypeAdapter.NativeToValue([]string{name, name + "suffix"})

		}),
	),
)

var ToInt64convGen = cel.Function("ToInt64",
	cel.Overload("ToInt64_interface{}",

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

var ToIntconvGen = cel.Function("ToInt",
	cel.Overload("ToInt_interface{}",

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

var ToInt64sconvGen = cel.Function("ToInt64s",
	cel.Overload("ToInt64s_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToInt64s(args[0]))

		}),
	),
)

var ToIntsconvGen = cel.Function("ToInts",
	cel.Overload("ToInts_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToInts(args[0]))

		}),
	),
)

var ToFloat64convGen = cel.Function("ToFloat64",
	cel.Overload("ToFloat64_interface{}",

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

var ToFloat64sconvGen = cel.Function("ToFloat64s",
	cel.Overload("ToFloat64s_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToFloat64s(args[0]))

		}),
	),
)

var ToStringconvGen = cel.Function("ToString",
	cel.Overload("ToString_interface{}",

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

var ToStringsconvGen = cel.Function("ToStrings",
	cel.Overload("ToStrings_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x ConvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToStrings(args[0]))

		}),
	),
)

var DefaultconvGen = cel.Function("Default",
	cel.Overload("Default_interface{}_interface{}",

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

var DictconvGen = cel.Function("Dict",
	cel.Overload("Dict_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			// Need to figure this out
			name := "Flanksource"
			return types.DefaultTypeAdapter.NativeToValue([]string{name, name + "suffix"})

		}),
	),
)
