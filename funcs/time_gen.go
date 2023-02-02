package funcs

import (
	gotime "time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var ZoneNametimeGen = cel.Function("time.ZoneName",
	cel.Overload("time.ZoneName_",
		nil,
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ZoneName())

		}),
	),
)

var ZoneOffsettimeGen = cel.Function("time.ZoneOffset",
	cel.Overload("time.ZoneOffset_",
		nil,
		cel.IntType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ZoneOffset())

		}),
	),
)

var ParsetimeGen = cel.Function("time.Parse",
	cel.Overload("time.Parse_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs
			a0, a1 := x.Parse(args[0].Value().(string), args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var ParseLocaltimeGen = cel.Function("time.ParseLocal",
	cel.Overload("time.ParseLocal_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs
			a0, a1 := x.ParseLocal(args[0].Value().(string), args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var ParseInLocationtimeGen = cel.Function("time.ParseInLocation",
	cel.Overload("time.ParseInLocation_string_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs
			a0, a1 := x.ParseInLocation(args[0].Value().(string), args[1].Value().(string), args[2])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var NowtimeGen = cel.Function("time.Now",
	cel.Overload("time.Now_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Now())

		}),
	),
)

var UnixtimeGen = cel.Function("time.Unix",
	cel.Overload("time.Unix_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs
			a0, a1 := x.Unix(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var NanosecondtimeGen = cel.Function("time.Nanosecond",
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

var MicrosecondtimeGen = cel.Function("time.Microsecond",
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

var MillisecondtimeGen = cel.Function("time.Millisecond",
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

var SecondtimeGen = cel.Function("time.Second",
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

var MinutetimeGen = cel.Function("time.Minute",
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

var HourtimeGen = cel.Function("time.Hour",
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

var ParseDurationtimeGen = cel.Function("time.ParseDuration",
	cel.Overload("time.ParseDuration_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs
			a0, a1 := x.ParseDuration(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var SincetimeGen = cel.Function("time.Since",
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

var UntiltimeGen = cel.Function("time.Until",
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
