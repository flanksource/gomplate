// Code generated by "gencel";
// DO NOT EDIT.

package funcs

import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types/ref"

var ReplaceAllstringsGen = cel.Function("strings.ReplaceAll",
	cel.Overload("strings.ReplaceAll_string_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.StringType, cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ReplaceAll(args[0].Value().(string), args[1].Value().(string), args[2]))

		}),
	),
)

var ContainsstringsGen = cel.Function("strings.Contains",
	cel.Overload("strings.Contains_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Contains(args[0].Value().(string), args[1]))

		}),
	),
)

var HasPrefixstringsGen = cel.Function("strings.HasPrefix",
	cel.Overload("strings.HasPrefix_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.HasPrefix(args[0].Value().(string), args[1]))

		}),
	),
)

var HasSuffixstringsGen = cel.Function("strings.HasSuffix",
	cel.Overload("strings.HasSuffix_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.HasSuffix(args[0].Value().(string), args[1]))

		}),
	),
)

var RepeatstringsGen = cel.Function("strings.Repeat",
	cel.Overload("strings.Repeat_int_interface{}",

		[]*cel.Type{
			cel.IntType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			a0, a1 := x.Repeat(args[0].Value().(int), args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var SortstringsGen = cel.Function("strings.Sort",
	cel.Overload("strings.Sort_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			a0, a1 := x.Sort(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var SplitstringsGen = cel.Function("strings.Split",
	cel.Overload("strings.Split_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Split(args[0].Value().(string), args[1]))

		}),
	),
)

var SplitNstringsGen = cel.Function("strings.SplitN",
	cel.Overload("strings.SplitN_string_int_interface{}",

		[]*cel.Type{
			cel.StringType, cel.IntType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.SplitN(args[0].Value().(string), args[1].Value().(int), args[2]))

		}),
	),
)

var TrimstringsGen = cel.Function("strings.Trim",
	cel.Overload("strings.Trim_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Trim(args[0].Value().(string), args[1]))

		}),
	),
)

var TrimPrefixstringsGen = cel.Function("strings.TrimPrefix",
	cel.Overload("strings.TrimPrefix_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.TrimPrefix(args[0].Value().(string), args[1]))

		}),
	),
)

var TrimSuffixstringsGen = cel.Function("strings.TrimSuffix",
	cel.Overload("strings.TrimSuffix_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.TrimSuffix(args[0].Value().(string), args[1]))

		}),
	),
)

var TitlestringsGen = cel.Function("strings.Title",
	cel.Overload("strings.Title_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Title(args[0]))

		}),
	),
)

var ToUpperstringsGen = cel.Function("strings.ToUpper",
	cel.Overload("strings.ToUpper_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToUpper(args[0]))

		}),
	),
)

var ToLowerstringsGen = cel.Function("strings.ToLower",
	cel.Overload("strings.ToLower_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToLower(args[0]))

		}),
	),
)

var TrimSpacestringsGen = cel.Function("strings.TrimSpace",
	cel.Overload("strings.TrimSpace_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.TrimSpace(args[0]))

		}),
	),
)

var TruncstringsGen = cel.Function("strings.Trunc",
	cel.Overload("strings.Trunc_int_interface{}",

		[]*cel.Type{
			cel.IntType, cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Trunc(args[0].Value().(int), args[1]))

		}),
	),
)

var SlugstringsGen = cel.Function("strings.Slug",
	cel.Overload("strings.Slug_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Slug(args[0]))

		}),
	),
)

var QuotestringsGen = cel.Function("strings.Quote",
	cel.Overload("strings.Quote_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Quote(args[0]))

		}),
	),
)

var ShellQuotestringsGen = cel.Function("strings.ShellQuote",
	cel.Overload("strings.ShellQuote_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ShellQuote(args[0]))

		}),
	),
)

var SquotestringsGen = cel.Function("strings.Squote",
	cel.Overload("strings.Squote_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Squote(args[0]))

		}),
	),
)

var SnakeCasestringsGen = cel.Function("strings.SnakeCase",
	cel.Overload("strings.SnakeCase_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			a0, a1 := x.SnakeCase(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var CamelCasestringsGen = cel.Function("strings.CamelCase",
	cel.Overload("strings.CamelCase_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			a0, a1 := x.CamelCase(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var KebabCasestringsGen = cel.Function("strings.KebabCase",
	cel.Overload("strings.KebabCase_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			a0, a1 := x.KebabCase(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)
