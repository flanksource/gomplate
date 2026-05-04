package gomplate

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

const (
	foldInitListFn = "cel.@foldInitList"
	foldInitMapFn  = "cel.@foldInitMap"
)

func getFoldCelLibrary() cel.EnvOption {
	return cel.Lib(&foldCelLibrary{})
}

type foldCelLibrary struct{}

func (l *foldCelLibrary) LibraryName() string {
	return "gomplate.fold"
}

func (l *foldCelLibrary) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Macros(
			cel.ReceiverMacro("fold", 3, foldListMacro,
				cel.MacroDocs("Folds a list using an element variable, an accumulator variable, and a step expression."),
				cel.MacroExamples("[1, 2, 3].fold(e, acc, acc + e) // 6")),
			cel.ReceiverMacro("fold", 4, foldMapMacro,
				cel.MacroDocs("Folds a map using key/value variables, an accumulator variable, and a step expression."),
				cel.MacroExamples(`{"a": "apple", "b": "banana"}.fold(k, v, acc, acc + v) // "applebanana"`)),
		),
		cel.Function(foldInitListFn,
			cel.Overload("fold_init_list_dyn", []*cel.Type{cel.DynType}, cel.DynType,
				cel.UnaryBinding(func(collection ref.Val) ref.Val {
					return foldInitialValue(collection, false)
				})),
		),
		cel.Function(foldInitMapFn,
			cel.Overload("fold_init_map_dyn", []*cel.Type{cel.DynType}, cel.DynType,
				cel.UnaryBinding(func(collection ref.Val) ref.Val {
					return foldInitialValue(collection, true)
				})),
		),
		cel.Function("merge",
			cel.Overload("merge_map_map", []*cel.Type{cel.MapType(cel.DynType, cel.DynType), cel.MapType(cel.DynType, cel.DynType)}, cel.MapType(cel.DynType, cel.DynType),
				cel.BinaryBinding(mergeMaps)),
		),
	}
}

func (*foldCelLibrary) ProgramOptions() []cel.ProgramOption {
	return nil
}

func foldListMacro(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	iterVar, err := extractFoldIdent(mef, args[0])
	if err != nil {
		return nil, err
	}
	accuVar, err := extractFoldIdent(mef, args[1])
	if err != nil {
		return nil, err
	}
	if iterVar == accuVar {
		return nil, mef.NewError(args[1].ID(), fmt.Sprintf("duplicate variable name: %s", accuVar))
	}

	return mef.NewComprehension(
		target,
		iterVar,
		accuVar,
		mef.NewCall(foldInitListFn, mef.Copy(target)),
		mef.NewLiteral(types.True),
		args[2],
		mef.NewIdent(accuVar),
	), nil
}

func foldMapMacro(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	keyVar, err := extractFoldIdent(mef, args[0])
	if err != nil {
		return nil, err
	}
	valVar, err := extractFoldIdent(mef, args[1])
	if err != nil {
		return nil, err
	}
	accuVar, err := extractFoldIdent(mef, args[2])
	if err != nil {
		return nil, err
	}
	if keyVar == valVar || keyVar == accuVar || valVar == accuVar {
		return nil, mef.NewError(args[2].ID(), "fold variable names must be unique")
	}

	return mef.NewComprehensionTwoVar(
		target,
		keyVar,
		valVar,
		accuVar,
		mef.NewCall(foldInitMapFn, mef.Copy(target)),
		mef.NewLiteral(types.True),
		args[3],
		mef.NewIdent(accuVar),
	), nil
}

func extractFoldIdent(mef cel.MacroExprFactory, expr ast.Expr) (string, *cel.Error) {
	if expr.Kind() != ast.IdentKind {
		return "", mef.NewError(expr.ID(), "argument must be a simple name")
	}
	return expr.AsIdent(), nil
}

func foldInitialValue(collection ref.Val, mapValues bool) ref.Val {
	var first ref.Val
	if mapValues {
		m, ok := collection.(traits.Mapper)
		if !ok {
			return types.NewErr("fold target is not a map")
		}
		it := m.Iterator()
		if it.HasNext() != types.True {
			return types.DefaultTypeAdapter.NativeToValue(nil)
		}
		first = m.Get(it.Next())
	} else {
		l, ok := collection.(traits.Lister)
		if !ok {
			return types.NewErr("fold target is not a list")
		}
		if l.Size().Equal(types.IntZero) == types.True {
			return types.DefaultTypeAdapter.NativeToValue(nil)
		}
		first = l.Get(types.IntZero)
	}
	return zeroValueForFold(first)
}

func mergeMaps(lhs, rhs ref.Val) ref.Val {
	left, ok := lhs.(traits.Mapper)
	if !ok {
		return types.NewErr("left operand is not a map")
	}
	right, ok := rhs.(traits.Mapper)
	if !ok {
		return types.NewErr("right operand is not a map")
	}

	out := map[ref.Val]ref.Val{}
	for it := left.Iterator(); it.HasNext() == types.True; {
		key := it.Next()
		out[key] = left.Get(key)
	}
	for it := right.Iterator(); it.HasNext() == types.True; {
		key := it.Next()
		out[key] = right.Get(key)
	}
	return types.NewRefValMap(types.DefaultTypeAdapter, out)
}

func zeroValueForFold(v ref.Val) ref.Val {
	if types.IsError(v) {
		return v
	}
	switch v.(type) {
	case types.Int:
		return types.IntZero
	case types.Uint:
		return types.Uint(0)
	case types.Double:
		return types.Double(0)
	case types.String:
		return types.String("")
	case types.Bytes:
		return types.Bytes([]byte{})
	case traits.Mapper:
		return types.NewRefValMap(types.DefaultTypeAdapter, map[ref.Val]ref.Val{})
	case traits.Lister:
		return types.NewRefValList(types.DefaultTypeAdapter, []ref.Val{})
	default:
		return types.DefaultTypeAdapter.NativeToValue(nil)
	}
}
