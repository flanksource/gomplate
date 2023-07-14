// Code generated by gencel. DO NOT EDIT.

package funcs

import (
	"github.com/flanksource/gomplate/v3/conv"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var stringsHumanDurationGen = cel.Function("HumanDuration",
	cel.Overload("HumanDuration_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			a0, a1 := x.HumanDuration(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var stringsHumanSizeGen = cel.Function("HumanSize",
	cel.Overload("HumanSize_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			a0, _ := x.HumanSize(conv.ToString(args[0])) // Never returns error
			return types.String(a0)
		}),
	),
)

var stringsSemverGen = cel.Function("Semver",
	cel.Overload("Semver_string",

		[]*cel.Type{
			cel.StringType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			a0, a1 := x.SemverMap(args[0].Value().(string))
			if a1 != nil {
				return types.String("")
			}


			return types.DefaultTypeAdapter.NativeToValue(a0)
		}),
	),
)

var stringsSemverCompareGen = cel.Function("SemverCompare",
	cel.Overload("SemverCompare_string_string",

		[]*cel.Type{
			cel.StringType, cel.StringType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs

			a0, a1 := x.SemverCompare(args[0].Value().(string), args[1].Value().(string))
			if a1 != nil {
				return types.Bool(false)
			}

			return types.DefaultTypeAdapter.NativeToValue(a0)

		}),
	),
)

var stringsAbbrevGen = cel.Function("Abbrev",
	cel.Overload("Abbrev_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.Abbrev(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var stringsReplaceAllGen = cel.Function("ReplaceAll",
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

var stringsContainsGen = cel.Function("Contains",
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

var stringsHasPrefixGen = cel.Function("HasPrefix",
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

var stringsHasSuffixGen = cel.Function("HasSuffix",
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

var stringsRepeatGen = cel.Function("Repeat",
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

var stringsSortGen = cel.Function("Sort",
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

var stringsSplitGen = cel.Function("Split",
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

var stringsSplitNGen = cel.Function("SplitN",
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

var stringsTrimGen = cel.Function("Trim",
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

var stringsTrimPrefixGen = cel.Function("TrimPrefix",
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

var stringsTrimSuffixGen = cel.Function("TrimSuffix",
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

var stringsTitleGen = cel.Function("Title",
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

var stringsToUpperGen = cel.Function("ToUpper",
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

var stringsToLowerGen = cel.Function("ToLower",
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

var stringsTrimSpaceGen = cel.Function("TrimSpace",
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

var stringsTruncGen = cel.Function("Trunc",
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

var stringsIndentGen = cel.Function("Indent",
	cel.Overload("Indent_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.Indent(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var stringsSlugGen = cel.Function("Slug",
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

var stringsQuoteGen = cel.Function("Quote",
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

var stringsShellQuoteGen = cel.Function("ShellQuote",
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

var stringsSquoteGen = cel.Function("Squote",
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

var stringsSnakeCaseGen = cel.Function("SnakeCase",
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

var stringsCamelCaseGen = cel.Function("CamelCase",
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

var stringsKebabCaseGen = cel.Function("KebabCase",
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

var stringsWordWrapGen = cel.Function("WordWrap",
	cel.Overload("WordWrap_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.WordWrap(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var stringsRuneCountGen = cel.Function("RuneCount",
	cel.Overload("RuneCount_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x StringFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.RuneCount(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)
