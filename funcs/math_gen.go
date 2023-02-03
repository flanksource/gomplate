package funcs

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var mathAddGen = cel.Function("math.Add",
	cel.Overload("math.Add_interface{}",
		[]*cel.Type{
			cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			converted := transferSlice[interface{}](args[0])
			var x MathFuncs
			return types.DefaultTypeAdapter.NativeToValue(x.Add(converted...))
		}),
	),
)

var mathIsIntGen = cel.Function("math.IsInt",
	cel.Overload("math.IsInt_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.IsInt(args[0]))

		}),
	),
)

var mathIsFloatGen = cel.Function("math.IsFloat",
	cel.Overload("math.IsFloat_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.IsFloat(args[0]))

		}),
	),
)

var mathIsNumGen = cel.Function("math.IsNum",
	cel.Overload("math.IsNum_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.IsNum(args[0]))

		}),
	),
)

var mathAbsGen = cel.Function("math.Abs",
	cel.Overload("math.Abs_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Abs(args[0]))

		}),
	),
)

var mathSubGen = cel.Function("math.Sub",
	cel.Overload("math.Sub_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Sub(args[0], args[1]))

		}),
	),
)

var mathDivGen = cel.Function("math.Div",
	cel.Overload("math.Div_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs
			a0, a1 := x.Div(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var mathRemGen = cel.Function("math.Rem",
	cel.Overload("math.Rem_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Rem(args[0], args[1]))

		}),
	),
)

var mathPowGen = cel.Function("math.Pow",
	cel.Overload("math.Pow_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Pow(args[0], args[1]))

		}),
	),
)

var mathCeilGen = cel.Function("math.Ceil",
	cel.Overload("math.Ceil_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Ceil(args[0]))

		}),
	),
)

var mathFloorGen = cel.Function("math.Floor",
	cel.Overload("math.Floor_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Floor(args[0]))

		}),
	),
)

var mathRoundGen = cel.Function("math.Round",
	cel.Overload("math.Round_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x MathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Round(args[0]))

		}),
	),
)
