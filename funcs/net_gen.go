package funcs

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var netContainsCIDRGen = cel.Function("net.ContainsCIDR",
	cel.Overload("net.ContainsCIDR_string_string",
		[]*cel.Type{
			cel.StringType, cel.StringType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			var x NetFuncs
			return types.Bool(x.ContainsCIDR(args[0].Value().(string), args[1].Value().(string)))
		}),
	),
)

var netIsValidIPGen = cel.Function("net.IsValidIP",
	cel.Overload("net.IsValidIP_string",
		[]*cel.Type{
			cel.StringType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			var x NetFuncs
			return types.Bool(x.IsValidIP(args[0].Value().(string)))
		}),
	),
)
