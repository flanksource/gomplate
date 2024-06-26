// Code generated by gencel. DO NOT EDIT.

package funcs

import (
	"github.com/google/cel-go/cel"
	"encoding/json"
	"reflect"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/protobuf/types/known/structpb"
	"github.com/flanksource/gomplate/v3/data"
)

var dataJSONMemberGen = cel.Function("JSON",
	cel.MemberOverload(".string.JSON()",
		[]*cel.Type{
			cel.StringType,
		},
		cel.DynType,
		cel.UnaryBinding(func(arg ref.Val) ref.Val {
			var x DataFuncs
			result, err := x.JSON(arg)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var dataJSONArrayMemberGen = cel.Function("JSONArray",
	cel.MemberOverload(".string.JSONArray()",
		[]*cel.Type{
			cel.StringType,
		},
		cel.DynType,
		cel.UnaryBinding(func(arg ref.Val) ref.Val {
			var x DataFuncs
			result, err := x.JSONArray(arg)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var dataYAMLGen = cel.Function("YAML",
	cel.Overload("YAML_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x DataFuncs

			result, err := x.YAML(args[0])
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var dataYAMLArrayGen = cel.Function("YAMLArray",
	cel.Overload("data.YAMLArray_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x DataFuncs

			result, err := x.YAMLArray(args[0].Value())
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var dataTOMLGen = cel.Function("TOML",
	cel.Overload("TOML_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x DataFuncs

			result, err := x.TOML(args[0])
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)


		}),
	),
)

var dataCSVGen = cel.Function("CSV",
	cel.Overload("CSV_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			list, err := sliceToNative[string](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := data.CSV(list...)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var dataCSVByRowGen = cel.Function("data.CSVByRow",
	cel.Overload("data.CSVByRow_string",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			list, err := sliceToNative[string](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := data.CSVByRow(list...)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)

		}),
	),
)

var dataCSVByColumnGen = cel.Function("data.CSVByColumn",
	cel.Overload("data.CSVByColumn_string",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			list, err := sliceToNative[string](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}
			result, err := data.CSVByColumn(list...)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)

var dataToCSVGen = cel.Function("toCSV",
	cel.Overload("toCSV_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			list, err := sliceToNative[interface{}](args[0].(ref.Val))
			if err != nil {
				return types.WrapErr(err)
			}

			result, err := data.ToCSV(list...)
			if err != nil {
				return types.WrapErr(err)
			}
			return types.String(result)

		}),
	),
)

var (
	listType  = reflect.TypeOf(&structpb.ListValue{})
	mapType   = reflect.TypeOf(&structpb.Struct{})
	mapStrDyn = cel.MapType(cel.StringType, cel.DynType)
)

func toJson(val ref.Val) ref.Val {
	var typeDesc reflect.Type

	switch val.Type() {
	case types.ListType:
		typeDesc = listType
	case types.MapType:
		typeDesc = mapType
	default:
		if result, err := json.Marshal(val.Value());  err != nil {
			return types.WrapErr(err)
		} else {
			return types.String(result)
		}
	}

	nativeVal, err := val.ConvertToNative(typeDesc)
	if err != nil {
		return types.NewErr("failed to convert to native: %w", err)
	}

	if result, err := json.Marshal(nativeVal); err != nil {
		return types.WrapErr(err)
	} else {
		return types.String(result)
	}
}

var dataToJSONGen = cel.Function("toJSON",
	cel.MemberOverload("dyn_toJSON",
		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.UnaryBinding(toJson),
	),
)

var dataToJSONPrettyGen = cel.Function("toJSONPretty",
	cel.MemberOverload("toJSONPretty_interface{}",
		[]*cel.Type{
			cel.DynType, cel.StringType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			result, err := data.ToJSONPretty(args[1].Value().(string), args[0].Value())
			if err != nil {
				return types.WrapErr(err)
			}
			return types.String(result)
		}),
	),
)

var dataToYAMLGen = cel.Function("toYAML",
	cel.Overload("toYAML_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			result, err := data.ToYAML(args[0].Value())
			if err != nil {
				return types.WrapErr(err)
			}
			return types.String(result)
		}),
	),
)

var dataToTOMLGen = cel.Function("toTOML",
	cel.Overload("toTOML_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			result, err := data.ToTOML(args[0].Value())
			if err != nil {
				return types.WrapErr(err)
			}
			return types.String(result)

		}),
	),
)
