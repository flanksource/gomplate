package funcs

import (
	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
)

var urlEncodeGen = cel.Function("urlencode",
	cel.Overload("urlencode.string",
		[]*cel.Type{
			cel.StringType,
		},
		cel.StringType,
		cel.UnaryBinding(func(arg ref.Val) ref.Val {
			var x EncodeFuncs
			result := x.URLEncode(arg.Value().(string))
			return types.String(result)
		}),
	),
)

var urlDecodeGen = cel.Function("urldecode",
	cel.Overload("urldecode.string",
		[]*cel.Type{
			cel.StringType,
		},
		cel.StringType,
		cel.UnaryBinding(func(arg ref.Val) ref.Val {
			var x EncodeFuncs
			result, _ := x.URLDecode(arg.Value().(string))
			return types.String(result)
		}),
	),
)
