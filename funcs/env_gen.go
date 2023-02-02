package funcs

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var GetenvenvGen = cel.Function("Getenv",
	cel.Overload("Getenv_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x EnvFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Getenv(args[0], args[1]))

		}),
	),
)

var ExpandEnvenvGen = cel.Function("ExpandEnv",
	cel.Overload("ExpandEnv_interface{}",

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
