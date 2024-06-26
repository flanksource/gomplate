// Code generated by gencel. DO NOT EDIT.

package funcs

import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/common/types/ref"
import gotime "time"

var timeZoneNameGen = cel.Function("time.ZoneName",
	cel.Overload("time.ZoneName_",
		nil,
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.String(x.ZoneName())

		}),
	),
)

var timeZoneOffsetGen = cel.Function("time.ZoneOffset",
	cel.Overload("time.ZoneOffset_",
		nil,
		cel.IntType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.Int(x.ZoneOffset())

		}),
	),
)

var timeParseGen = cel.Function("time.Parse",
	cel.Overload("time.Parse_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			result, err := x.Parse(args[0].Value().(string), args[1])
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var timeParseLocalGen = cel.Function("time.ParseLocal",
	cel.Overload("time.ParseLocal_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			result, err := x.ParseLocal(args[0].Value().(string), args[1])
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)


		}),
	),
)

var timeParseInLocationGen = cel.Function("time.ParseInLocation",
	cel.Overload("time.ParseInLocation_string_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			result, err := x.ParseInLocation(args[0].Value().(string), args[1].Value().(string), args[2])
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var timeNowGen = cel.Function("time.Now",
	cel.Overload("time.Now_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Now())

		}),
	),
)

var timeUnixGen = cel.Function("time.Unix",
	cel.Overload("time.Unix_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			result, err := x.Unix(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				result, err,
			})

		}),
	),
)

var timeNanosecondGen = cel.Function("time.Nanosecond",
	cel.Overload("time.Nanosecond_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Nanosecond(args[0]))

		}),
	),
)

var timeMicrosecondGen = cel.Function("time.Microsecond",
	cel.Overload("time.Microsecond_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Microsecond(args[0]))

		}),
	),
)

var timeMillisecondGen = cel.Function("time.Millisecond",
	cel.Overload("time.Millisecond_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Millisecond(args[0]))

		}),
	),
)

var timeSecondGen = cel.Function("time.Second",
	cel.Overload("time.Second_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Second(args[0]))

		}),
	),
)

var timeMinuteGen = cel.Function("time.Minute",
	cel.Overload("time.Minute_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Minute(args[0]))

		}),
	),
)

var timeHourGen = cel.Function("time.Hour",
	cel.Overload("time.Hour_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Hour(args[0]))

		}),
	),
)


var timeParseDurationGen2= cel.Function("duration",
	cel.Overload("duration_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			result, err := x.ParseDuration(args[0])
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)
var timeParseDurationGen = cel.Function("time.ParseDuration",
	cel.Overload("time.ParseDuration_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			result, err := x.ParseDuration(args[0])
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)


		}),
	),
)

var timeSinceGen = cel.Function("time.Since",
	cel.Overload("time.Since_gotime.Time",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Since(args[0].Value().(gotime.Time)))

		}),
	),
)

var timeUntilGen = cel.Function("time.Until",
	cel.Overload("time.Until_gotime.Time",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Until(args[0].Value().(gotime.Time)))

		}),
	),
)
