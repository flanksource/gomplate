package gomplate

import (
	"reflect"

	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/flanksource/gomplate/v3/kubernetes"
	"github.com/flanksource/gomplate/v3/strings"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

var typeAdapters = []cel.EnvOption{}

func RegisterType(i any) {
	typeAdapters = append(typeAdapters, ext.NativeTypes(reflect.TypeOf(i)))
}

func GetCelEnv(environment map[string]any) []cel.EnvOption {
	// Generated functions
	var opts = funcs.CelEnvOption
	opts = append(opts, kubernetes.Library()...)
	opts = append(opts, ext.Strings(), ext.Encoders(), ext.Lists(), ext.Math(), ext.Sets())
	opts = append(opts, cel.StdLib())
	opts = append(opts, cel.OptionalTypes())
	opts = append(opts, strings.Library...)
	opts = append(opts, typeAdapters...)

	// Load input as variables
	for k := range environment {
		opts = append(opts, cel.Variable(k, cel.AnyType))
	}

	return opts
}

// The following identifiers are reserved to allow easier embedding of CEL into a host language.
//
// Reference: https://github.com/google/cel-spec/blob/master/doc/langdef.md
var celKeywords = map[string]struct{}{
	"as":        {},
	"break":     {},
	"const":     {},
	"continue":  {},
	"else":      {},
	"for":       {},
	"function":  {},
	"if":        {},
	"import":    {},
	"let":       {},
	"loop":      {},
	"namespace": {},
	"package":   {},
	"return":    {},
	"var":       {},
	"void":      {},
	"while":     {},
}

// IsCelKeyword returns true if the given key is a reserved word in Cel
func IsCelKeyword(key string) bool {
	_, ok := celKeywords[key]
	return ok
}
