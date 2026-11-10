package gomplate

import (
	"fmt"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/ast"
	"cel.dev/cel-go/common/operators"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
)

func collectionCelOptions() []cel.EnvOption {
	options := make([]cel.EnvOption, 1, 9)
	options[0] = cel.Macros(cel.ReceiverMacro("joinBy", 5, joinByMacro))
	for name, binding := range map[string]func(ref.Val) ref.Val{
		"groupBy": groupRows, "distinctBy": distinctRows, "latestBy": latestRow,
	} {
		options = append(options,
			cel.Macros(cel.ReceiverMacro(name, 2, collectionKeyMacro(name))),
			cel.Function("gomplate.@"+name, cel.Overload("collection_"+name, []*cel.Type{cel.ListType(cel.DynType)}, cel.DynType, cel.UnaryBinding(binding))),
		)
	}
	return append(options,
		cel.Function("gomplate.@joinBy", cel.Overload("collection_joinBy", []*cel.Type{cel.ListType(cel.DynType), cel.ListType(cel.DynType)}, cel.ListType(cel.DynType), cel.BinaryBinding(joinRows))),
		cel.Function("one", cel.MemberOverload("collection_one", []*cel.Type{cel.ListType(cel.DynType)}, cel.DynType, cel.UnaryBinding(oneRow))),
	)
}

func collectionKeyMacro(name string) cel.MacroFactory {
	return func(factory cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
		projection, err := keyedRowsMacro(factory, target, args)
		if err != nil {
			return nil, err
		}
		return factory.NewCall("gomplate.@"+name, projection), nil
	}
}

func keyedRowsMacro(factory cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	variable, err := extractFoldIdent(factory, args[0])
	if err != nil {
		return nil, err
	}
	accumulator := "@collection:" + fmt.Sprint(target.ID())
	return factory.NewComprehension(target, variable, accumulator,
		factory.NewList(), factory.NewLiteral(types.True),
		factory.NewCall(operators.Add, factory.NewIdent(accumulator), factory.NewList(factory.NewList(factory.NewIdent(variable), args[1]))),
		factory.NewIdent(accumulator)), nil
}

func joinByMacro(factory cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	left, err := keyedRowsMacro(factory, target, args[1:3])
	if err != nil {
		return nil, err
	}
	right, err := keyedRowsMacro(factory, args[0], args[3:5])
	if err != nil {
		return nil, err
	}
	return factory.NewCall("gomplate.@joinBy", left, right), nil
}

type collectionGroup struct {
	key  ref.Val
	rows []ref.Val
}

func indexedGroups(value ref.Val) ([]collectionGroup, map[any]int, ref.Val) {
	list, ok := value.(traits.Lister)
	if !ok {
		return nil, nil, types.NewErr("groupBy requires a list")
	}
	groups, index := []collectionGroup{}, map[any]int{}
	for iterator := list.Iterator(); iterator.HasNext() == types.True; {
		pair := iterator.Next().(traits.Lister)
		row, key := pair.Get(types.IntZero), pair.Get(types.Int(1))
		if types.IsError(key) || types.IsUnknown(key) {
			return nil, nil, key
		}
		switch key.(type) {
		case types.String, types.Int, types.Uint, types.Bool:
		default:
			return nil, nil, types.NewErr("groupBy key must be a string, integer or boolean, got %s", key.Type())
		}
		identity := any(key)
		position, exists := index[identity]
		if !exists {
			position = len(groups)
			index[identity] = position
			groups = append(groups, collectionGroup{key: key})
		}
		groups[position].rows = append(groups[position].rows, row)
	}
	return groups, index, nil
}

func groupRows(value ref.Val) ref.Val {
	groups, _, err := indexedGroups(value)
	if err != nil {
		return err
	}
	rows := make([]ref.Val, len(groups))
	for index, group := range groups {
		rows[index] = types.NewRefValMap(types.DefaultTypeAdapter, map[ref.Val]ref.Val{
			types.String("key"): group.key, types.String("rows"): types.NewRefValList(types.DefaultTypeAdapter, group.rows),
		})
	}
	return types.NewRefValList(types.DefaultTypeAdapter, rows)
}

func distinctRows(value ref.Val) ref.Val {
	groups, _, err := indexedGroups(value)
	if err != nil {
		return err
	}
	rows := make([]ref.Val, len(groups))
	for index, group := range groups {
		rows[index] = group.rows[0]
	}
	return types.NewRefValList(types.DefaultTypeAdapter, rows)
}

func oneRow(value ref.Val) ref.Val {
	list := value.(traits.Lister)
	if list.Size() != types.Int(1) {
		return types.NewErr("one requires exactly one row, got %v", list.Size())
	}
	return list.Get(types.IntZero)
}

func latestRow(value ref.Val) ref.Val {
	list := value.(traits.Lister)
	if list.Size() == types.IntZero {
		return types.NewErr("latestBy requires at least one row")
	}
	var latest, key ref.Val
	tied := false
	for iterator := list.Iterator(); iterator.HasNext() == types.True; {
		pair := iterator.Next().(traits.Lister)
		candidate := pair.Get(types.Int(1))
		comparer, ok := candidate.(traits.Comparer)
		if !ok || types.IsError(candidate) || types.IsUnknown(candidate) {
			return types.NewErr("latestBy requires comparable keys, got %s", candidate.Type())
		}
		if key == nil {
			latest, key = pair.Get(types.IntZero), candidate
			continue
		}
		comparison := comparer.Compare(key)
		if types.IsError(comparison) {
			return comparison
		}
		if comparison == types.Int(1) {
			latest, key, tied = pair.Get(types.IntZero), candidate, false
		} else if comparison == types.IntZero {
			tied = true
		}
	}
	if tied {
		return types.NewErr("latestBy has multiple rows at the latest key %v", key)
	}
	return latest
}

func joinRows(left, right ref.Val) ref.Val {
	_, _, err := indexedGroups(left)
	if err != nil {
		return err
	}
	rightGroups, index, err := indexedGroups(right)
	if err != nil {
		return err
	}
	rows := []ref.Val{}
	for iterator := left.(traits.Lister).Iterator(); iterator.HasNext() == types.True; {
		pair := iterator.Next().(traits.Lister)
		position, exists := index[any(pair.Get(types.Int(1)))]
		if !exists {
			continue
		}
		for _, rightRow := range rightGroups[position].rows {
			rows = append(rows, types.NewRefValMap(types.DefaultTypeAdapter, map[ref.Val]ref.Val{types.String("left"): pair.Get(types.IntZero), types.String("right"): rightRow}))
		}
	}
	return types.NewRefValList(types.DefaultTypeAdapter, rows)
}
