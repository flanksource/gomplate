package funcs

import (
	gotime "time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var ZoneNametimeGen = cel.Function("ZoneName",
	cel.Overload("ZoneName_",
		nil,
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ZoneName())

		}),
	),
)

var ZoneOffsettimeGen = cel.Function("ZoneOffset",
	cel.Overload("ZoneOffset_",
		nil,
		cel.IntType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ZoneOffset())

		}),
	),
)

var ParsetimeGen = cel.Function("Parse",
	cel.Overload("Parse_string_interface{}",

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

var ParseLocaltimeGen = cel.Function("ParseLocal",
	cel.Overload("ParseLocal_string_interface{}",

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

var ParseInLocationtimeGen = cel.Function("ParseInLocation",
	cel.Overload("ParseInLocation_string_string_interface{}",

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

var NowtimeGen = cel.Function("Now",
	cel.Overload("Now_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x TimeFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Now())

		}),
	),
)

var UnixtimeGen = cel.Function("Unix",
	cel.Overload("Unix_interface{}",

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

var NanosecondtimeGen = cel.Function("Nanosecond",
	cel.Overload("Nanosecond_interface{}",

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

var MicrosecondtimeGen = cel.Function("Microsecond",
	cel.Overload("Microsecond_interface{}",

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

var MillisecondtimeGen = cel.Function("Millisecond",
	cel.Overload("Millisecond_interface{}",

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

var SecondtimeGen = cel.Function("Second",
	cel.Overload("Second_interface{}",

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

var MinutetimeGen = cel.Function("Minute",
	cel.Overload("Minute_interface{}",

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

var HourtimeGen = cel.Function("Hour",
	cel.Overload("Hour_interface{}",

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

var ParseDurationtimeGen = cel.Function("ParseDuration",
	cel.Overload("ParseDuration_interface{}",

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

var SincetimeGen = cel.Function("Since",
	cel.Overload("Since_gotime.Time",

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

var UntiltimeGen = cel.Function("Until",
	cel.Overload("Until_gotime.Time",

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
