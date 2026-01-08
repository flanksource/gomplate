package funcs

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
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
