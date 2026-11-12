package texttemplate

import (
	"fmt"
	"io"
	"reflect"
	"text/template/parse"
)

// ValuesProvider lets a caller expose data to a template without converting it up front.
type ValuesProvider interface {
	// Field resolves .name (and string-keyed index) on receiver. found=false hands the lookup back to
	// the standard method, struct-field and map-key resolution; found=true with an invalid value is a
	// missing key and follows the missingkey option.
	Field(receiver reflect.Value, name string) (value reflect.Value, found bool, err error)
	// Value shapes a value about to be printed, ranged over, truth-tested or passed to a function.
	// Intermediate values of a field chain and $variable assignments are not passed through Value.
	Value(v reflect.Value) (reflect.Value, error)
}

// ExecuteOptions configures a single execution. The zero value is identical to text/template.
type ExecuteOptions struct {
	// Values resolves fields and shapes consumed values; nil uses text/template semantics.
	Values ValuesProvider
	// NilSafe makes field access through a nil or missing receiver, and index of a nil collection,
	// a missing key or an out-of-range position, yield the missing value instead of an error, and a
	// missing value prints as empty text instead of "<no value>". missingkey=error still reports
	// missing map keys.
	NilSafe bool
}

// ExecuteWith applies a parsed template to data like Execute, with per-execution options.
func (t *Template) ExecuteWith(wr io.Writer, data any, opts ExecuteOptions) error {
	return t.execute(wr, data, opts)
}

// consume passes a value that is about to be printed, ranged, truth-tested or handed to a
// function through the provider.
func (s *state) consume(v reflect.Value) reflect.Value {
	if s.values == nil {
		return v
	}
	out, err := s.values.Value(v)
	if err != nil {
		s.errorf("%v", err)
	}
	return out
}

// argValue consumes a function argument. A nil type marks the receiver of a chain such as
// (pipe).Field, which is an intermediate and stays raw.
func (s *state) argValue(v reflect.Value, typ reflect.Type) reflect.Value {
	if typ == nil {
		return v
	}
	return s.consume(v)
}

func (s *state) providerField(receiver reflect.Value, name string) (reflect.Value, bool) {
	value, found, err := s.values.Field(receiver, name)
	if err != nil {
		s.errorf("%v", err)
	}
	return value, found
}

// missingField is the result of a key no lookup resolved, following the missingkey option.
func (s *state) missingField(name string) reflect.Value {
	if s.tmpl.option.missingKey == mapError {
		s.errorf("map has no entry for key %q", name)
	}
	return zero
}

// rawArg evaluates the argument nodes that produce data rather than literals, unconverted.
func (s *state) rawArg(dot reflect.Value, n parse.Node) (reflect.Value, bool) {
	switch arg := n.(type) {
	case *parse.DotNode:
		return dot, true
	case *parse.FieldNode:
		return s.evalFieldNode(dot, arg, []parse.Node{n}, missingVal), true
	case *parse.VariableNode:
		return s.evalVariableNode(dot, arg, nil, missingVal), true
	case *parse.PipeNode:
		return s.evalPipeline(dot, arg), true
	case *parse.IdentifierNode:
		return s.evalFunction(dot, arg, arg, nil, missingVal), true
	case *parse.ChainNode:
		return s.evalChainNode(dot, arg, nil, missingVal), true
	}
	return reflect.Value{}, false
}

// evalIndex runs the index builtin on raw arguments, so a provider resolves string keys lazily
// and nil-safety applies.
func (s *state) evalIndex(dot reflect.Value, node parse.Node, args []parse.Node, final reflect.Value) reflect.Value {
	argv := make([]reflect.Value, 0, len(args)+1)
	for _, arg := range args {
		v, ok := s.rawArg(dot, arg)
		if !ok {
			v = s.evalArg(dot, reflectValueType, arg).Interface().(reflect.Value)
		}
		argv = append(argv, reflect.ValueOf(v))
	}
	if !isMissing(final) {
		argv = append(argv, reflect.ValueOf(final))
	}
	v, err := safeCall(reflect.ValueOf(s.index), argv)
	if err != nil {
		s.at(node)
		s.errorf("error calling index: %w", err)
	}
	return v.Interface().(reflect.Value)
}

// index is the index builtin with provider lookups for string keys and nil-safety.
func (s *state) index(item reflect.Value, indexes ...reflect.Value) (reflect.Value, error) {
	item = indirectInterface(item)
	for _, index := range indexes {
		if !item.IsValid() {
			if s.nilSafe {
				return zero, nil
			}
			return zero, fmt.Errorf("index of untyped nil")
		}
		index = indirectInterface(index)
		var isNil bool
		if item, isNil = indirect(item); isNil {
			if s.nilSafe {
				return zero, nil
			}
			return zero, fmt.Errorf("index of nil pointer")
		}
		next, err := s.indexOne(item, index)
		if err != nil {
			return zero, err
		}
		item = next
	}
	if !item.IsValid() && len(indexes) == 0 && !s.nilSafe {
		return zero, fmt.Errorf("index of untyped nil")
	}
	return item, nil
}

// indexOne mirrors one step of the text/template index builtin; missing map keys still yield the
// element's zero value so map[string]string lookups print "" as before.
func (s *state) indexOne(item, index reflect.Value) (reflect.Value, error) {
	if s.values != nil && index.Kind() == reflect.String {
		if v, found := s.providerField(item, index.String()); found {
			if !v.IsValid() {
				return s.missingField(index.String()), nil
			}
			return v, nil
		}
		if item.Kind() == reflect.Struct {
			return s.missingField(index.String()), nil
		}
	}
	switch item.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		x, err := indexArg(index, item.Len())
		if s.nilSafe && intLike(index.Kind()) && (err != nil || x >= item.Len()) {
			return zero, nil
		}
		if err != nil {
			return zero, err
		}
		return item.Index(x), nil
	case reflect.Map:
		key, err := prepareArg(index, item.Type().Key())
		if err != nil {
			return zero, err
		}
		if x := item.MapIndex(key); x.IsValid() {
			return x, nil
		}
		return reflect.Zero(item.Type().Elem()), nil
	}
	return zero, fmt.Errorf("can't index item of type %s", item.Type())
}
