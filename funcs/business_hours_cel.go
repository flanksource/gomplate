package funcs

import (
	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
)

var inBusinessHourGen = cel.Function("in_business_hours",
	cel.Overload("in_business_hours_string",
		[]*cel.Type{
			cel.StringType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			in, err := inBusinessHour(args[0].Value().(string))
			if err != nil {
				return types.WrapErr(err)
			}
			if in == nil {
				return types.NullValue
			}
			return types.Bool(*in)
		}),
	),
)
