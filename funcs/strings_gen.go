// Code generated by "gencel";
// DO NOT EDIT.

package funcs

import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types/ref"

var AbbrevstringsGen = cel.Function("Abbrev",
	cel.Overload("Abbrev_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			a0, a1 := x.Abbrev(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var ReplaceAllstringsGen = cel.Function("ReplaceAll",
	cel.Overload("ReplaceAll_string_string_interface{}",

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

var ContainsstringsGen = cel.Function("Contains",
	cel.Overload("Contains_string_interface{}",

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

var HasPrefixstringsGen = cel.Function("HasPrefix",
	cel.Overload("HasPrefix_string_interface{}",

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

var HasSuffixstringsGen = cel.Function("HasSuffix",
	cel.Overload("HasSuffix_string_interface{}",

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

var RepeatstringsGen = cel.Function("Repeat",
	cel.Overload("Repeat_int_interface{}",

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

var SortstringsGen = cel.Function("Sort",
	cel.Overload("Sort_interface{}",

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

var SplitstringsGen = cel.Function("Split",
	cel.Overload("Split_string_interface{}",

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

var SplitNstringsGen = cel.Function("SplitN",
	cel.Overload("SplitN_string_int_interface{}",

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

var TrimstringsGen = cel.Function("Trim",
	cel.Overload("Trim_string_interface{}",

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

var TrimPrefixstringsGen = cel.Function("TrimPrefix",
	cel.Overload("TrimPrefix_string_interface{}",

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

var TrimSuffixstringsGen = cel.Function("TrimSuffix",
	cel.Overload("TrimSuffix_string_interface{}",

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

var TitlestringsGen = cel.Function("Title",
	cel.Overload("Title_interface{}",

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

var ToUpperstringsGen = cel.Function("ToUpper",
	cel.Overload("ToUpper_interface{}",

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

var ToLowerstringsGen = cel.Function("ToLower",
	cel.Overload("ToLower_interface{}",

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

var TrimSpacestringsGen = cel.Function("TrimSpace",
	cel.Overload("TrimSpace_interface{}",

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

var TruncstringsGen = cel.Function("Trunc",
	cel.Overload("Trunc_int_interface{}",

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

var IndentstringsGen = cel.Function("Indent",
	cel.Overload("Indent_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			a0, a1 := x.Indent(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var SlugstringsGen = cel.Function("Slug",
	cel.Overload("Slug_interface{}",

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

var QuotestringsGen = cel.Function("Quote",
	cel.Overload("Quote_interface{}",

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

var ShellQuotestringsGen = cel.Function("ShellQuote",
	cel.Overload("ShellQuote_interface{}",

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

var SquotestringsGen = cel.Function("Squote",
	cel.Overload("Squote_interface{}",

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

var SnakeCasestringsGen = cel.Function("SnakeCase",
	cel.Overload("SnakeCase_interface{}",

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

var CamelCasestringsGen = cel.Function("CamelCase",
	cel.Overload("CamelCase_interface{}",

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

var KebabCasestringsGen = cel.Function("KebabCase",
	cel.Overload("KebabCase_interface{}",

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

var WordWrapstringsGen = cel.Function("WordWrap",
	cel.Overload("WordWrap_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			a0, a1 := x.WordWrap(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var RuneCountstringsGen = cel.Function("RuneCount",
	cel.Overload("RuneCount_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			a0, a1 := x.RuneCount(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)
