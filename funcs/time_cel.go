package funcs

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// timeInTimeRangeGen exposes time.InTimeRange(t, start, end) → bool.
// Returns true when the time of day of t falls within [start, end] inclusive.
// start and end are "HH:MM" or "HH:MM:SS" strings.
// Accepts a timestamp (time.Time) or an RFC3339 string as the first argument.
//
// Examples:
//
//	time.InTimeRange(t, "09:00", "17:00")
//	time.InTimeRange(t, "09:30:00", "17:30:00")
var timeInTimeRangeGen = cel.Function("time.InTimeRange",
	cel.Overload("time.InTimeRange_any_string_string",
		[]*cel.Type{cel.AnyType, cel.StringType, cel.StringType},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			var x TimeFuncs
			start := args[1].Value().(string)
			end := args[2].Value().(string)
			result, err := x.InTimeRange(args[0].Value(), start, end)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.Bool(result)
		}),
	),
)
