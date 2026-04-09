package funcs

import (
	"context"
	"strings"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"

	"github.com/flanksource/gomplate/v3/coll"
	"github.com/pkg/errors"
)

// CollNS -
//
// Deprecated: don't use
func CollNS() *CollFuncs {
	return &CollFuncs{}
}

// AddCollFuncs -
//
// Deprecated: use CreateCollFuncs instead
func AddCollFuncs(f map[string]interface{}) {
	for k, v := range CreateCollFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateCollFuncs -
func CreateCollFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}

	ns := &CollFuncs{ctx}
	f["coll"] = func() interface{} { return ns }

	f["has"] = ns.Has
	f["slice"] = ns.Slice
	f["dict"] = ns.Dict
	f["keys"] = ns.Keys
	f["values"] = ns.Values
	f["append"] = ns.Append
	f["prepend"] = ns.Prepend
	f["uniq"] = ns.Uniq
	f["reverse"] = ns.Reverse
	f["merge"] = ns.Merge
	f["sort"] = ns.Sort
	f["jq"] = ns.JQ
	f["flatten"] = ns.Flatten

	f["coalesce"] = ns.Coalesce
	f["first"] = ns.First
	f["last"] = ns.Last
	f["matchLabel"] = coll.MatchLabel
	f["mapToKeyVal"] = coll.MapToKeyVal[any]
	f["keyValToMap"] = coll.KeyValToMap
	f["jsonpath"] = coll.JSONPath
	f["jmespath"] = coll.JMESPath
	f["xpath"] = coll.XPath
	return f
}

// CollFuncs -
type CollFuncs struct {
	ctx context.Context
}

// Slice -
func (CollFuncs) Slice(args ...interface{}) []interface{} {
	return coll.Slice(args...)
}

// Has -
func (CollFuncs) Has(in interface{}, key string) bool {
	return coll.Has(in, key)
}

// Dict -
func (CollFuncs) Dict(in ...interface{}) (map[string]interface{}, error) {
	return coll.Dict(in...)
}

// Keys -
func (CollFuncs) Keys(in map[string]any) []string {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	return keys
}

// Values -
func (CollFuncs) Values(in map[string]any) []any {
	values := make([]any, 0, len(in))
	for _, v := range in {
		values = append(values, v)
	}
	return values
}

// Append -
func (CollFuncs) Append(v interface{}, list interface{}) ([]interface{}, error) {
	return coll.Append(v, list)
}

// Prepend -
func (CollFuncs) Prepend(v interface{}, list interface{}) ([]interface{}, error) {
	return coll.Prepend(v, list)
}

// Uniq -
func (CollFuncs) Uniq(in interface{}) ([]interface{}, error) {
	return coll.Uniq(in)
}

// Reverse -
func (CollFuncs) Reverse(in interface{}) ([]interface{}, error) {
	return coll.Reverse(in)
}

// Merge -
func (CollFuncs) Merge(dst map[string]interface{}, src ...map[string]interface{}) (map[string]interface{}, error) {
	return coll.Merge(dst, src...)
}

// Sort -
func (CollFuncs) Sort(args ...interface{}) ([]interface{}, error) {
	var (
		key  string
		list interface{}
	)
	if len(args) == 0 || len(args) > 2 {
		return nil, errors.Errorf("wrong number of args: wanted 1 or 2, got %d", len(args))
	}
	if len(args) == 1 {
		list = args[0]
	}
	if len(args) == 2 {
		key = conv.ToString(args[0])
		list = args[1]
	}
	return coll.Sort(key, list)
}

// JQ -
func (f *CollFuncs) JQ(jqExpr string, in interface{}) (interface{}, error) {
	return coll.JQ(f.ctx, jqExpr, in)
}

// Flatten -
func (CollFuncs) Flatten(args ...interface{}) ([]interface{}, error) {
	if len(args) == 0 || len(args) > 2 {
		return nil, errors.Errorf("wrong number of args: wanted 1 or 2, got %d", len(args))
	}
	list := args[0]
	depth := -1
	if len(args) == 2 {
		depth = conv.ToInt(args[0])
		list = args[1]
	}
	return coll.Flatten(list, depth)
}

func pickOmitArgs(args ...interface{}) (map[string]interface{}, []string, error) {
	if len(args) <= 1 {
		return nil, nil, errors.Errorf("wrong number of args: wanted 2 or more, got %d", len(args))
	}

	m, ok := args[len(args)-1].(map[string]interface{})
	if !ok {
		return nil, nil, errors.Errorf("wrong map type: must be map[string]interface{}, got %T", args[len(args)-1])
	}

	keys := make([]string, len(args)-1)
	for i, v := range args[0 : len(args)-1] {
		k, ok := v.(string)
		if !ok {
			return nil, nil, errors.Errorf("wrong key type: must be string, got %T (%+v)", args[i], args[i])
		}
		keys[i] = k
	}
	return m, keys, nil
}

// Pick -
func (CollFuncs) Pick(args ...interface{}) (map[string]interface{}, error) {
	m, keys, err := pickOmitArgs(args...)
	if err != nil {
		return nil, err
	}
	return coll.Pick(m, keys...), nil
}

// Omit -
func (CollFuncs) Omit(args ...interface{}) (map[string]interface{}, error) {
	m, keys, err := pickOmitArgs(args...)
	if err != nil {
		return nil, err
	}
	return coll.Omit(m, keys...), nil
}

// Coalesce returns the first argument that is neither nil nor empty.
func (CollFuncs) Coalesce(args ...interface{}) interface{} {
	return coll.Coalesce(args...)
}

// First returns the first element of a list, first character of a string, or
// value at the lexicographically smallest key of a map.
func (CollFuncs) First(in interface{}) interface{} {
	return coll.First(in)
}

// Last returns the last element of a list, last character of a string, or
// value at the lexicographically largest key of a map.
func (CollFuncs) Last(in interface{}) interface{} {
	return coll.Last(in)
}

// celCoalesceFirst returns the first non-null, non-empty ref.Val from args,
// unwrapping CEL optional<T> values along the way.
// optional.none() and plain null are skipped; optional.of(v) is unwrapped and
// its inner value is checked for emptiness.
func celCoalesceFirst(args []ref.Val) ref.Val {
	for _, arg := range args {
		if opt, ok := arg.(*types.Optional); ok {
			if !opt.HasValue() {
				continue
			}
			arg = opt.GetValue()
		}
		if arg == types.NullValue {
			continue
		}
		if s, ok := arg.(types.String); ok && string(s) == "" {
			continue
		}
		if l, ok := arg.(traits.Lister); ok && l.Size() == types.IntZero {
			continue
		}
		if m, ok := arg.(traits.Mapper); ok && m.Size() == types.IntZero {
			continue
		}
		return arg
	}
	return types.NullValue
}

var celCoalesce = cel.Function("coalesce",
	cel.Overload("coalesce_1", []*cel.Type{cel.DynType}, cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val { return celCoalesceFirst(args) }),
	),
	cel.Overload("coalesce_2", []*cel.Type{cel.DynType, cel.DynType}, cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val { return celCoalesceFirst(args) }),
	),
	cel.Overload("coalesce_3", []*cel.Type{cel.DynType, cel.DynType, cel.DynType}, cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val { return celCoalesceFirst(args) }),
	),
	cel.Overload("coalesce_4", []*cel.Type{cel.DynType, cel.DynType, cel.DynType, cel.DynType}, cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val { return celCoalesceFirst(args) }),
	),
	cel.Overload("coalesce_5", []*cel.Type{cel.DynType, cel.DynType, cel.DynType, cel.DynType, cel.DynType}, cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val { return celCoalesceFirst(args) }),
	),
)

func celFirstLastImpl(arg ref.Val, first bool) ref.Val {
	if opt, ok := arg.(*types.Optional); ok {
		if !opt.HasValue() {
			return types.NullValue
		}
		arg = opt.GetValue()
	}
	if arg == types.NullValue {
		return types.NullValue
	}

	if l, ok := arg.(traits.Lister); ok {
		if l.Size() == types.IntZero {
			return types.NullValue
		}
		if first {
			return l.Get(types.IntZero)
		}
		return l.Get(l.Size().(types.Int) - 1)
	}

	if _, ok := arg.(traits.Mapper); ok {
		m, err := convertMap(arg)
		if err != nil {
			return types.WrapErr(err)
		}
		var result any
		if first {
			result = coll.First(m)
		} else {
			result = coll.Last(m)
		}
		if result == nil {
			return types.NullValue
		}
		return types.DefaultTypeAdapter.NativeToValue(result)
	}

	return types.NullValue
}

var celFirst = cel.Function("first",
	cel.Overload("first_dyn", []*cel.Type{cel.DynType}, cel.DynType,
		cel.OverloadIsNonStrict(),
		cel.UnaryBinding(func(arg ref.Val) ref.Val { return celFirstLastImpl(arg, true) }),
	),
	cel.MemberOverload("dyn_first", []*cel.Type{cel.DynType}, cel.DynType,
		cel.OverloadIsNonStrict(),
		cel.UnaryBinding(func(arg ref.Val) ref.Val { return celFirstLastImpl(arg, true) }),
	),
)

var celLast = cel.Function("last",
	cel.Overload("last_dyn", []*cel.Type{cel.DynType}, cel.DynType,
		cel.OverloadIsNonStrict(),
		cel.UnaryBinding(func(arg ref.Val) ref.Val { return celFirstLastImpl(arg, false) }),
	),
	cel.MemberOverload("dyn_last", []*cel.Type{cel.DynType}, cel.DynType,
		cel.OverloadIsNonStrict(),
		cel.UnaryBinding(func(arg ref.Val) ref.Val { return celFirstLastImpl(arg, false) }),
	),
)

var celLabelsMatch = cel.Function("matchLabel",
	cel.Overload("matchLabel_map_string_string",
		[]*cel.Type{
			cel.MapType(cel.StringType, cel.DynType), cel.StringType, cel.StringType,
		},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			key := conv.ToString(args[1])
			valuePatterns := strings.Split(conv.ToString(args[2]), ",")

			labels, err := convertMap(args[0])
			if err != nil {
				return types.WrapErr(errors.New("matchLabel expects the first argument to be a map[string]any"))
			}

			result := coll.MatchLabel(labels, key, valuePatterns...)
			return types.DefaultTypeAdapter.NativeToValue(result)
		}),
	),
)
