// Code generated by "gencel";
// DO NOT EDIT.

package funcs

import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types/ref"

var BasefilepathGen = cel.Function("filepath.Base",
	cel.Overload("filepath.Base_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Base(args[0]))

		}),
	),
)

var CleanfilepathGen = cel.Function("filepath.Clean",
	cel.Overload("filepath.Clean_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Clean(args[0]))

		}),
	),
)

var DirfilepathGen = cel.Function("filepath.Dir",
	cel.Overload("filepath.Dir_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Dir(args[0]))

		}),
	),
)

var ExtfilepathGen = cel.Function("filepath.Ext",
	cel.Overload("filepath.Ext_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Ext(args[0]))

		}),
	),
)

var FromSlashfilepathGen = cel.Function("filepath.FromSlash",
	cel.Overload("filepath.FromSlash_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.FromSlash(args[0]))

		}),
	),
)

var IsAbsfilepathGen = cel.Function("filepath.IsAbs",
	cel.Overload("filepath.IsAbs_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.IsAbs(args[0]))

		}),
	),
)

var MatchfilepathGen = cel.Function("filepath.Match",
	cel.Overload("filepath.Match_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs
			a0, a1 := x.Match(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var RelfilepathGen = cel.Function("filepath.Rel",
	cel.Overload("filepath.Rel_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs
			a0, a1 := x.Rel(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var SplitfilepathGen = cel.Function("filepath.Split",
	cel.Overload("filepath.Split_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.Split(args[0]))

		}),
	),
)

var ToSlashfilepathGen = cel.Function("filepath.ToSlash",
	cel.Overload("filepath.ToSlash_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.ToSlash(args[0]))

		}),
	),
)

var VolumeNamefilepathGen = cel.Function("filepath.VolumeName",
	cel.Overload("filepath.VolumeName_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x FilePathFuncs

			return types.DefaultTypeAdapter.NativeToValue(x.VolumeName(args[0]))

		}),
	),
)
