// Code generated by "gencel";
// DO NOT EDIT.

package funcs

import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types/ref"

var envGetenvGen = cel.Function("env.Getenv",
	cel.Overload("env.Getenv_interface{}_string",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x EnvFuncs
			list := transferSlice[string](args[1].(ref.Val))

			return types.DefaultTypeAdapter.NativeToValue(x.Getenv(args[0], list...))

		}),
	),
)

var envExpandEnvGen = cel.Function("env.ExpandEnv",
	cel.Overload("env.ExpandEnv_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x EnvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ExpandEnv(args[0]))

		}),
	),
)
