package gomplate

import (
	"fmt"
	"sort"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

const (
	foldInitListFn       = "cel.@foldInitList"
	foldInitMapFn        = "cel.@foldInitMap"
	foldSortedMapEntries = "cel.@foldSortedMapEntries"
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
		cel.Function(foldSortedMapEntries,
			cel.Overload("fold_sorted_map_entries_dyn", []*cel.Type{cel.DynType}, cel.ListType(cel.ListType(cel.DynType)),
				cel.UnaryBinding(sortedMapEntries)),
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

	entryVar := "__fold_entry__"
	if entryVar == keyVar || entryVar == valVar || entryVar == accuVar {
		entryVar = "__fold_entry2__"
	}
	entry := mef.NewIdent(entryVar)
	key := mef.NewCall(operators.Index, entry, mef.NewLiteral(types.IntZero))
	value := mef.NewCall(operators.Index, mef.NewIdent(entryVar), mef.NewLiteral(types.Int(1)))
	step := mef.NewComprehensionTwoVar(
		mef.NewMap(mef.NewMapEntry(key, value, false)),
		keyVar,
		valVar,
		accuVar,
		mef.NewIdent(accuVar),
		mef.NewLiteral(types.True),
		args[3],
		mef.NewIdent(accuVar),
	)

	return mef.NewComprehension(
		mef.NewCall(foldSortedMapEntries, mef.Copy(target)),
		entryVar,
		accuVar,
		mef.NewCall(foldInitMapFn, mef.Copy(target)),
		mef.NewLiteral(types.True),
		step,
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

func sortedMapEntries(collection ref.Val) ref.Val {
	m, ok := collection.(traits.Mapper)
	if !ok {
		return types.NewErr("fold target is not a map")
	}

	keys := []ref.Val{}
	for it := m.Iterator(); it.HasNext() == types.True; {
		keys = append(keys, it.Next())
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i].Value()) < fmt.Sprint(keys[j].Value())
	})

	entries := make([]ref.Val, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, types.NewRefValList(types.DefaultTypeAdapter, []ref.Val{key, m.Get(key)}))
	}
	return types.NewRefValList(types.DefaultTypeAdapter, entries)
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
