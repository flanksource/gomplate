package gomplate

import (
	gocontext "context"
	"fmt"
	"reflect"
	"regexp"

	"github.com/flanksource/commons/context"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/flanksource/gomplate/v3/kubernetes"
	"github.com/flanksource/gomplate/v3/strings"
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
	opts = append(opts, getGoTemplateCelFunction())

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
	"true":      {},
	"false":     {},
	"null":      {},
	"in":        {},
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
	"type":      {},
}

var celIdentifierRegexp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// IsCelKeyword returns true if the given key is a reserved word in Cel
func IsCelKeyword(key string) bool {
	_, ok := celKeywords[key]
	return ok
}

func IsValidCELIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	return !IsCelKeyword(s) && celIdentifierRegexp.MatchString(s)
}

// getGoTemplateCelFunction returns a CEL function that calls gotemplate on a format string
func getGoTemplateCelFunction() cel.EnvOption {
	return cel.Function("f",
		cel.Overload("f_string_any",
			[]*cel.Type{
				cel.StringType, cel.DynType,
			},
			cel.StringType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				format := conv.ToString(args[0])
				data := args[1].Value()

				env := map[string]any{}
				switch v := data.(type) {
				case map[string]any:
					env = v
				case map[string]string:
					for k, v := range v {
						env[k] = v
					}
				default:
					// Otherwise, make data available as 'data' variable
					env["data"] = v
				}

				// Use struct templater as it supports ValueFunctions and multiple delims
				st := StructTemplater{
					Context:        context.NewContext(gocontext.Background()),
					Values:         env,
					ValueFunctions: true,
					DelimSets: []Delims{
						{Left: "$(", Right: ")"},
						{Left: "{{", Right: "}}"},
					},
				}
				result, err := st.Template(format)
				if err != nil {
					return types.WrapErr(fmt.Errorf("gotemplate error: %w", err))
				}

				return types.DefaultTypeAdapter.NativeToValue(result)
			}),
		),
	)
}
